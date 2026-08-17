package application

import (
	"context"
	"fmt"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	"rigmark/internal/modules/recommendation/domain"
	"rigmark/internal/modules/recommendation/ports"
)

type engineStub struct {
	result domain.Result
	count  int
}

type policyStub struct{ policy domain.Policy }

func (repository *policyStub) ActivePolicy(context.Context) (domain.Policy, error) {
	return repository.policy, nil
}
func (*policyStub) ListPolicies(context.Context) ([]domain.PolicySummary, error) { return nil, nil }
func (*policyStub) TransitionPolicy(context.Context, identity.UserID, string, domain.PolicyWorkflowStatus, string) error {
	return nil
}

func policyForProducts(products []catalog.Product) domain.Policy {
	policy := domain.Policy{Config: domain.Config{PolicyVersion: "test-v2"},
		Categories: make(map[string]domain.CategoryPolicy), Products: make(map[catalog.ProductID]domain.ProductPolicy)}
	for _, product := range products {
		policy.Categories[product.CategorySlug] = domain.CategoryPolicy{CategorySlug: product.CategorySlug, Supported: true}
		policy.Products[product.ID] = domain.ProductPolicy{ProductID: product.ID,
			FactRevisionID: product.FactRevisionID, ScoreRevisionID: product.ScoreRevisionID,
			GoalSupport: []domain.GoalSupport{{Goal: planning.GoalBuildMuscle, Score: 80}},
			Space:       domain.SpaceProfile{Footprint: domain.SpatialEnvelope{LengthMM: 1, WidthMM: 1, HeightMM: 1}}}
	}
	return policy
}

func serviceWithEngine(engine domain.RecommendationEngine, products []catalog.Product, catalogRepository *catalogStub, repository *repositoryStub) *Service {
	service := NewService(&policyStub{policy: policyForProducts(products)}, catalogRepository, repository)
	service.newEngine = func(domain.Config) (domain.RecommendationEngine, error) { return engine, nil }
	return service
}

func (engine *engineStub) Recommend(_ domain.Input, candidates []domain.CandidateSnapshot) (domain.Result, error) {
	engine.count = len(candidates)
	return engine.result, nil
}

type catalogStub struct {
	products       []catalog.Product
	productsByID   []catalog.Product
	listCalls      int
	listByIDsCalls int
}

func (repository *catalogStub) ListPublished(_ context.Context, filter catalogports.ProductFilter) ([]catalog.Product, error) {
	repository.listCalls++
	if filter.Offset >= len(repository.products) {
		return nil, nil
	}
	end := filter.Offset + filter.Limit
	if end > len(repository.products) {
		end = len(repository.products)
	}
	return repository.products[filter.Offset:end], nil
}

func (repository *catalogStub) ListByIDs(context.Context, []catalog.ProductID) ([]catalog.Product, error) {
	repository.listByIDsCalls++
	return repository.productsByID, nil
}

type repositoryStub struct {
	saveCalls int
	persisted ports.PersistedResult
}

func (*repositoryStub) GetDraft(context.Context, identity.UserID) (ports.Draft, error) {
	return ports.Draft{}, ports.ErrNotFound
}
func (*repositoryStub) SaveDraft(_ context.Context, _ identity.UserID, draft ports.Draft) (ports.Draft, error) {
	return draft, nil
}
func (*repositoryStub) DeleteDraft(context.Context, identity.UserID) error { return nil }
func (repository *repositoryStub) SaveResult(context.Context, identity.UserID, domain.Input, domain.Result, []domain.CandidateSnapshot) (ports.SavedResult, error) {
	repository.saveCalls++
	return ports.SavedResult{RecommendationID: "recommendation-1", SetupID: "setup-1"}, nil
}
func (*repositoryStub) ListSetups(context.Context, identity.UserID) ([]ports.SetupSummary, error) {
	return nil, nil
}
func (repository *repositoryStub) GetResultBySetupID(context.Context, identity.UserID, planning.SetupID) (ports.PersistedResult, error) {
	if repository.persisted.SetupID == "" {
		return ports.PersistedResult{}, ports.ErrNotFound
	}
	return repository.persisted, nil
}
func (*repositoryStub) RenameSetup(context.Context, identity.UserID, planning.SetupID, string) error {
	return nil
}
func (*repositoryStub) DeleteSetup(context.Context, identity.UserID, planning.SetupID) error {
	return nil
}

func TestGeneratePersistsOnlyForAuthenticatedUser(t *testing.T) {
	product := catalog.Product{ID: "product-1", Name: "Demo product", CategorySlug: "dumbbells",
		FactRevisionID: "fact-1", ScoreRevisionID: "score-1"}
	engine := &engineStub{result: domain.Result{Status: domain.ResultComplete}}
	repository := &repositoryStub{}
	catalogRepository := &catalogStub{products: []catalog.Product{product}}
	service := serviceWithEngine(engine, []catalog.Product{product}, catalogRepository, repository)

	guest, err := service.Generate(context.Background(), nil, domain.Input{})
	if err != nil {
		t.Fatalf("generate guest result: %v", err)
	}
	if guest.Saved || repository.saveCalls != 0 || engine.count != 1 {
		t.Fatalf("guest generation unexpectedly persisted: %#v", guest)
	}

	userID := identity.UserID("user-1")
	authenticated, err := service.Generate(context.Background(), &userID, domain.Input{})
	if err != nil {
		t.Fatalf("generate authenticated result: %v", err)
	}
	if !authenticated.Saved || authenticated.SetupID == nil || repository.saveCalls != 1 {
		t.Fatalf("authenticated generation was not persisted: %#v", authenticated)
	}
}

func TestSaveDraftRejectsInvalidPartialState(t *testing.T) {
	service := NewService(&policyStub{}, &catalogStub{}, &repositoryStub{})
	_, err := service.SaveDraft(context.Background(), "user-1", ports.Draft{CurrentStep: 9})
	if err != ErrInvalidDraft {
		t.Fatalf("expected ErrInvalidDraft, got %v", err)
	}
}

func TestGenerateLoadsTheEntireBoundedCatalog(t *testing.T) {
	products := make([]catalog.Product, 101)
	for index := range products {
		products[index] = catalog.Product{
			ID:              catalog.ProductID(fmt.Sprintf("product-%03d", index)),
			Name:            "Demo product",
			CategorySlug:    "dumbbells",
			FactRevisionID:  fmt.Sprintf("fact-%03d", index),
			ScoreRevisionID: fmt.Sprintf("score-%03d", index),
		}
	}
	engine := &engineStub{result: domain.Result{Status: domain.ResultComplete}}
	catalogRepository := &catalogStub{products: products}
	service := serviceWithEngine(engine, products, catalogRepository, &repositoryStub{})

	if _, err := service.Generate(context.Background(), nil, domain.Input{}); err != nil {
		t.Fatalf("generate recommendation: %v", err)
	}
	if engine.count != 101 || catalogRepository.listCalls != 2 {
		t.Fatalf("expected all 101 candidates in two pages, got candidates=%d calls=%d", engine.count, catalogRepository.listCalls)
	}
}

func TestGetSetupLoadsReferencedArchivedProductsByID(t *testing.T) {
	archived := catalog.Product{ID: "product-archived", Name: "Archived product", Status: catalog.ProductStatusDiscontinued}
	storedResult := domain.Result{Selected: []domain.RecommendedItem{{
		Product: domain.RankedProduct{Candidate: domain.CandidateSnapshot{
			ProductID: archived.ID,
			Price:     catalog.Money{AmountMinor: 12900, Currency: "USD"},
		}},
	}}}
	catalogRepository := &catalogStub{productsByID: []catalog.Product{archived}}
	service := NewService(&policyStub{}, catalogRepository, &repositoryStub{persisted: ports.PersistedResult{
		RecommendationID: "recommendation-1",
		SetupID:          "setup-1",
		SetupName:        "Saved setup",
		Result:           storedResult,
	}})

	generated, err := service.GetSetup(context.Background(), "user-1", "setup-1")
	if err != nil {
		t.Fatalf("GetSetup(): %v", err)
	}
	product, exists := generated.Products[archived.ID]
	if !exists || product.Status != catalog.ProductStatusDiscontinued || product.Price.AmountMinor != 12900 {
		t.Fatalf("archived setup product = %#v", product)
	}
	if catalogRepository.listByIDsCalls != 1 || catalogRepository.listCalls != 0 {
		t.Fatalf("catalog calls = byIDs %d, published %d", catalogRepository.listByIDsCalls, catalogRepository.listCalls)
	}
}
