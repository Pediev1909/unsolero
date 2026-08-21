package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	catalog "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	"rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

var ErrInvalidDraft = errors.New("invalid recommendation draft")
var ErrInvalidSetupName = errors.New("setup name must contain between 1 and 120 characters")
var ErrInvalidSetupPagination = errors.New("saved setup pagination is invalid")
var ErrCandidateCatalogTooLarge = errors.New("published catalog exceeds recommendation engine capacity")
var ErrStoredProductMissing = errors.New("a product referenced by the saved setup no longer exists")
var ErrUngovernedCandidate = errors.New("a published recommendation candidate has no approved evidence revision")
var ErrActivePolicyUnavailable = errors.New("no approved active recommendation policy is available")

var draftSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// draftCodePattern matches the normalized-code format shared by policy
// vocabularies. It constrains shape only; membership is the engine's job.
var draftCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

type Generated struct {
	RecommendationID *domain.RecommendationID
	SetupID          *planning.SetupID
	SetupName        *string
	Saved            bool
	Input            domain.Input
	Result           domain.Result
	Products         map[catalog.ProductID]catalog.Product
}

type Service struct {
	policies   ports.PolicyRepository
	catalog    ports.CatalogRepository
	repository ports.Repository
	newEngine  func(domain.Config) (domain.RecommendationEngine, error)
}

func NewService(
	policyRepository ports.PolicyRepository,
	catalogRepository ports.CatalogRepository,
	repository ports.Repository,
) *Service {
	return &Service{policies: policyRepository, catalog: catalogRepository, repository: repository,
		newEngine: func(config domain.Config) (domain.RecommendationEngine, error) {
			return domain.NewDeterministicRecommendationEngine(config)
		}}
}

func (service *Service) Generate(
	ctx context.Context,
	userID *identity.UserID,
	input domain.Input,
) (Generated, error) {
	policy, err := service.policies.ActivePolicy(ctx)
	if err != nil {
		return Generated{}, fmt.Errorf("%w: %v", ErrActivePolicyUnavailable, err)
	}
	engine, err := service.newEngine(policy.Config)
	if err != nil {
		return Generated{}, fmt.Errorf("%w: %v", ErrActivePolicyUnavailable, err)
	}
	products, err := service.listRecommendationCandidates(ctx)
	if err != nil {
		return Generated{}, fmt.Errorf("list recommendation candidates: %w", err)
	}
	candidates := make([]domain.CandidateSnapshot, 0, len(products))
	productMap := make(map[catalog.ProductID]catalog.Product, len(products))
	for _, product := range products {
		if product.FactRevisionID == "" || product.ScoreRevisionID == "" {
			return Generated{}, ErrUngovernedCandidate
		}
		candidate, policyErr := policy.Candidate(product)
		if errors.Is(policyErr, domain.ErrUnsupportedCategory) || errors.Is(policyErr, domain.ErrProductPolicyMissing) {
			continue
		}
		if policyErr != nil {
			return Generated{}, policyErr
		}
		candidates = append(candidates, candidate)
		productMap[product.ID] = product
	}
	input = policy.EnrichInput(input)
	result, err := engine.Recommend(input, candidates)
	if err != nil {
		return Generated{}, err
	}
	generated := Generated{Input: input, Result: result, Products: productMap}
	if userID == nil {
		return generated, nil
	}
	saved, err := service.repository.SaveResult(ctx, *userID, input, result, candidates)
	if err != nil {
		return Generated{}, fmt.Errorf("save recommendation result: %w", err)
	}
	generated.RecommendationID = &saved.RecommendationID
	generated.SetupID = &saved.SetupID
	generated.SetupName = &saved.SetupName
	generated.Saved = true
	return generated, nil
}

func (service *Service) GetDraft(ctx context.Context, userID identity.UserID) (ports.Draft, error) {
	return service.repository.GetDraft(ctx, userID)
}

func (service *Service) SaveDraft(
	ctx context.Context,
	userID identity.UserID,
	draft ports.Draft,
) (ports.Draft, error) {
	if err := validateDraft(draft); err != nil {
		return ports.Draft{}, err
	}
	return service.repository.SaveDraft(ctx, userID, draft)
}

func (service *Service) DeleteDraft(ctx context.Context, userID identity.UserID) error {
	return service.repository.DeleteDraft(ctx, userID)
}

func (service *Service) ListSetups(ctx context.Context, userID identity.UserID, page, pageSize int) (ports.SetupPage, error) {
	if page < 1 || page > 10_000 || pageSize < 1 || pageSize > 100 {
		return ports.SetupPage{}, ErrInvalidSetupPagination
	}
	return service.repository.ListSetups(ctx, userID, pageSize, (page-1)*pageSize)
}

func (service *Service) GetSetup(
	ctx context.Context,
	userID identity.UserID,
	setupID planning.SetupID,
) (Generated, error) {
	stored, err := service.repository.GetResultBySetupID(ctx, userID, setupID)
	if err != nil {
		return Generated{}, err
	}
	productIDs := resultProductIDs(stored.Result)
	products, err := service.catalog.ListByIDs(ctx, productIDs)
	if err != nil {
		return Generated{}, fmt.Errorf("load setup products: %w", err)
	}
	productMap := make(map[catalog.ProductID]catalog.Product, len(products))
	for _, product := range products {
		productMap[product.ID] = product
	}
	if len(productMap) != len(productIDs) {
		return Generated{}, ErrStoredProductMissing
	}
	applyStoredCandidates(productMap, stored.Result)
	return Generated{
		RecommendationID: &stored.RecommendationID,
		SetupID:          &stored.SetupID,
		SetupName:        &stored.SetupName,
		Saved:            true,
		Input:            stored.Input,
		Result:           stored.Result,
		Products:         productMap,
	}, nil
}

func (service *Service) listRecommendationCandidates(ctx context.Context) ([]catalog.Product, error) {
	const pageSize = 100
	products := make([]catalog.Product, 0, pageSize)
	for offset := 0; offset < domain.MaximumCandidates; offset += pageSize {
		page, err := service.catalog.ListPublished(ctx, catalogports.ProductFilter{
			Limit:  pageSize,
			Offset: offset,
		})
		if err != nil {
			return nil, err
		}
		products = append(products, page...)
		if len(page) < pageSize {
			return products, nil
		}
	}
	probe, err := service.catalog.ListPublished(ctx, catalogports.ProductFilter{
		Limit:  1,
		Offset: domain.MaximumCandidates,
	})
	if err != nil {
		return nil, err
	}
	if len(probe) > 0 {
		return nil, ErrCandidateCatalogTooLarge
	}
	return products, nil
}

func resultProductIDs(result domain.Result) []catalog.ProductID {
	seen := make(map[catalog.ProductID]bool)
	resultIDs := make([]catalog.ProductID, 0, len(result.Selected)+len(result.Alternatives)+len(result.Rejected))
	add := func(productID catalog.ProductID) {
		if productID == "" || seen[productID] {
			return
		}
		seen[productID] = true
		resultIDs = append(resultIDs, productID)
	}
	for _, item := range result.Selected {
		add(item.Product.Candidate.ProductID)
	}
	for _, item := range result.Alternatives {
		add(item.Product.Candidate.ProductID)
	}
	for _, item := range result.Rejected {
		add(item.Candidate.ProductID)
	}
	return resultIDs
}

func (service *Service) RenameSetup(
	ctx context.Context,
	userID identity.UserID,
	setupID planning.SetupID,
	name string,
) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 120 {
		return ErrInvalidSetupName
	}
	return service.repository.RenameSetup(ctx, userID, setupID, name)
}

func (service *Service) DeleteSetup(
	ctx context.Context,
	userID identity.UserID,
	setupID planning.SetupID,
) error {
	return service.repository.DeleteSetup(ctx, userID, setupID)
}

func applyStoredCandidates(products map[catalog.ProductID]catalog.Product, result domain.Result) {
	apply := func(candidate domain.CandidateSnapshot) {
		product, exists := products[candidate.ProductID]
		if !exists {
			return
		}
		product.Name = candidate.Name
		product.CategorySlug = candidate.CategorySlug
		product.Price = candidate.Price
		product.Dimensions = candidate.Dimensions
		product.Scores = candidate.Scores
		product.FactRevisionID = candidate.FactRevisionID
		product.ScoreRevisionID = candidate.ScoreRevisionID
		products[candidate.ProductID] = product
	}
	for _, item := range result.Selected {
		apply(item.Product.Candidate)
	}
	for _, item := range result.Alternatives {
		apply(item.Product.Candidate)
	}
	for _, item := range result.Rejected {
		apply(item.Candidate)
	}
}

func validateDraft(draft ports.Draft) error {
	if draft.CurrentStep < 1 || draft.CurrentStep > 8 || len(draft.FreeText) > 1000 ||
		len(draft.ExistingEquipment) > 50 || len(draft.TrainingPreferences) > 20 || len(draft.Priorities) > 20 {
		return ErrInvalidDraft
	}
	// Shape only, deliberately. This used to check the goal against a hardcoded
	// list of five fitness goals, which meant that after the vertical changed,
	// picking any real goal made every draft save fail with 422 -- a signed-in
	// visitor's progress silently stopped reaching their account from question
	// one onwards. It is the exact trap the comment further down describes: a
	// second copy of a vocabulary the policy already owns, left to drift.
	//
	// Which goals exist is declared by the active recommendation policy, and
	// the engine enforces that membership when the draft is used to generate.
	// A draft is a half-finished form, not a command.
	if draft.Goal != nil && !draftCodePattern.MatchString(string(*draft.Goal)) {
		return ErrInvalidDraft
	}
	if draft.Experience != nil && !map[planning.ExperienceLevel]bool{
		planning.ExperienceBeginner: true, planning.ExperienceIntermediate: true,
		planning.ExperienceAdvanced: true,
	}[*draft.Experience] {
		return ErrInvalidDraft
	}
	if draft.BudgetMinor != nil && (*draft.BudgetMinor <= 0 || draft.Currency == nil || *draft.Currency != "USD") {
		return ErrInvalidDraft
	}
	if (draft.BudgetMinor == nil) != (draft.Currency == nil) {
		return ErrInvalidDraft
	}
	if draft.AvailableSpace != nil {
		space := draft.AvailableSpace
		if space.LengthMM <= 0 || space.WidthMM <= 0 || space.HeightMM <= 0 ||
			space.LengthMM > 100_000 || space.WidthMM > 100_000 || space.HeightMM > 100_000 {
			return ErrInvalidDraft
		}
		if space.AccessWidthMM != nil && (*space.AccessWidthMM <= 0 || *space.AccessWidthMM > 100_000) {
			return ErrInvalidDraft
		}
	}
	for _, equipment := range draft.ExistingEquipment {
		name := strings.TrimSpace(equipment.Name)
		if name == "" || len(name) > 120 || !draftSlugPattern.MatchString(equipment.CategorySlug) {
			return ErrInvalidDraft
		}
	}
	// A draft is validated for shape only. Which preferences and priorities
	// actually exist is declared by the active recommendation policy, and the
	// engine enforces that membership when the draft is used. Repeating the
	// vocabulary here would mean three copies drifting apart on every policy
	// change, so the domain stays the single authority.
	seenPreferences := make(map[domain.TrainingPreference]bool, len(draft.TrainingPreferences))
	for _, preference := range draft.TrainingPreferences {
		if !draftCodePattern.MatchString(string(preference)) || seenPreferences[preference] {
			return ErrInvalidDraft
		}
		seenPreferences[preference] = true
	}
	seenPriorities := make(map[domain.Priority]bool, len(draft.Priorities))
	for _, priority := range draft.Priorities {
		if !draftCodePattern.MatchString(string(priority)) || seenPriorities[priority] {
			return ErrInvalidDraft
		}
		seenPriorities[priority] = true
	}
	return nil
}
