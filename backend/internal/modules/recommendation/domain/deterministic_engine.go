package domain

import (
	"sort"

	catalog "rigmark/internal/modules/catalog/domain"
)

type DeterministicRecommendationEngine struct {
	config    Config
	optimizer setupOptimizer
}

func NewDeterministicRecommendationEngine(config Config) (*DeterministicRecommendationEngine, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}
	return &DeterministicRecommendationEngine{
		config:    config,
		optimizer: deterministicSetupOptimizer{},
	}, nil
}

func (engine *DeterministicRecommendationEngine) Recommend(
	input Input,
	candidates []CandidateSnapshot,
) (Result, error) {
	if err := validateInput(input); err != nil {
		return Result{}, err
	}
	if err := validateCandidates(candidates); err != nil {
		return Result{}, err
	}
	for _, candidate := range candidates {
		if candidate.PolicyVersion != engine.config.PolicyVersion {
			return Result{}, ErrInvalidCandidate
		}
	}

	normalizedCandidates := normalizeCandidates(candidates)
	existing := normalizedExisting(input)
	fingerprint, err := resultFingerprint(input, normalizedCandidates, existing, engine.config)
	if err != nil {
		return Result{}, err
	}
	eligible, hardRejected := filterEligible(input, normalizedCandidates, existing)
	ranked := scoreCandidates(input, eligible, engine.config)
	selection := engine.optimizer.Optimize(optimizationInput{
		Goal: input.Goal, Budget: input.Budget.AmountMinor,
		Space: input.AvailableSpace, Existing: existing, Candidates: ranked, Config: engine.config,
	})

	result := Result{
		Status:           ResultNoSuitableProducts,
		PolicyVersion:    engine.config.PolicyVersion,
		EngineVersion:    EngineVersion,
		InputFingerprint: fingerprint,
		TotalCost:        catalog.Money{AmountMinor: selection.Total, Currency: input.Budget.Currency},
		Ranked:           ranked,
	}
	if selection.Complete {
		result.Status = ResultComplete
	}
	for index, product := range selection.Products {
		result.Selected = append(result.Selected, RecommendedItem{
			Rank: index + 1, Product: product, Quantity: 1,
			UnitPriceMinor: product.Candidate.Price.AmountMinor,
		})
	}
	result.ObjectiveScore, result.Breakdown = aggregateScores(selection.Products)
	result.Alternatives = selectAlternatives(input, selection, ranked, existing)
	result.Rejected = completeRejections(
		input,
		hardRejected,
		ranked,
		selection,
		result.Alternatives,
		existing,
	)
	return result, nil
}

func aggregateScores(products []RankedProduct) (int, ScoreBreakdown) {
	if len(products) == 0 {
		return 0, ScoreBreakdown{}
	}
	var objective int
	var result ScoreBreakdown
	for _, product := range products {
		objective += product.ObjectiveScore
		result.GoalMatch += product.Breakdown.GoalMatch
		result.BudgetMatch += product.Breakdown.BudgetMatch
		result.SpaceMatch += product.Breakdown.SpaceMatch
		result.ExperienceMatch += product.Breakdown.ExperienceMatch
		result.PreferenceMatch += product.Breakdown.PreferenceMatch
		result.Quality += product.Breakdown.Quality
		result.Value += product.Breakdown.Value
		result.Durability += product.Breakdown.Durability
		result.Compatibility += product.Breakdown.Compatibility
		result.Portability += product.Breakdown.Portability
		result.Noise += product.Breakdown.Noise
	}
	count := len(products)
	rounded := func(value int) int { return (value + count/2) / count }
	result.GoalMatch = rounded(result.GoalMatch)
	result.BudgetMatch = rounded(result.BudgetMatch)
	result.SpaceMatch = rounded(result.SpaceMatch)
	result.ExperienceMatch = rounded(result.ExperienceMatch)
	result.PreferenceMatch = rounded(result.PreferenceMatch)
	result.Quality = rounded(result.Quality)
	result.Value = rounded(result.Value)
	result.Durability = rounded(result.Durability)
	result.Compatibility = rounded(result.Compatibility)
	result.Portability = rounded(result.Portability)
	result.Noise = rounded(result.Noise)
	return rounded(objective), result
}

func completeRejections(
	input Input,
	hardRejected []RejectedProduct,
	ranked []RankedProduct,
	selection setupSelection,
	alternatives []Alternative,
	existing []ExistingEquipment,
) []RejectedProduct {
	result := append([]RejectedProduct(nil), hardRejected...)
	selectedIDs := make(map[catalog.ProductID]bool, len(selection.Products))
	selectedRedundancyGroups := make(map[string]bool, len(selection.Products))
	for _, product := range selection.Products {
		selectedIDs[product.Candidate.ProductID] = true
		for _, group := range product.Candidate.RedundancyGroups {
			selectedRedundancyGroups[group] = true
		}
	}
	alternativeIDs := make(map[catalog.ProductID]bool, len(alternatives))
	for _, alternative := range alternatives {
		alternativeIDs[alternative.Product.Candidate.ProductID] = true
	}
	existingCapabilities := equipmentCapabilities(existing)
	for _, product := range ranked {
		candidate := product.Candidate
		if selectedIDs[candidate.ProductID] || alternativeIDs[candidate.ProductID] {
			continue
		}
		rejection := RejectedProduct{Candidate: candidate}
		switch {
		case intersectsStringSet(candidate.RedundancyGroups, selectedRedundancyGroups):
			rejection.Code = "setup.lower_ranked"
			rejection.Message = "A stronger product filled the same setup role"
		case candidate.Price.AmountMinor > input.Budget.AmountMinor-selection.Total:
			rejection.Code = "setup.budget_limit"
			rejection.Message = "Adding it would exceed your setup budget"
		case !compatibleWithSelection(candidate, selection.Products, existingCapabilities):
			rejection.Code = "setup.incompatible"
			rejection.Message = "It is incompatible with the selected setup"
		case !fitsWithinTotalFloorArea(appendCandidate(selection.Products, product), input.AvailableSpace):
			rejection.Code = "setup.space_limit"
			rejection.Message = "Adding it would exceed your available floor space"
		default:
			rejection.Code = "setup.not_needed"
			rejection.Message = "It does not add enough non-redundant capability to this setup"
		}
		result = append(result, rejection)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].Candidate.ProductID < result[right].Candidate.ProductID
	})
	return result
}

var _ RecommendationEngine = (*DeterministicRecommendationEngine)(nil)
