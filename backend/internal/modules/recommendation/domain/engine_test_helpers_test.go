package domain

import (
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

func testEngine(t *testing.T) *DeterministicRecommendationEngine {
	t.Helper()
	engine, err := NewDeterministicRecommendationEngine(testPolicyConfig("test-policy-v2"))
	if err != nil {
		t.Fatalf("NewDeterministicRecommendationEngine(): %v", err)
	}
	return engine
}

func testPolicyConfig(version string) Config {
	return Config{PolicyVersion: version, Weights: Weights{
		GoalMatch: 20, BudgetMatch: 12, SpaceMatch: 12, ExperienceMatch: 10,
		PreferenceMatch: 8, Quality: 8, Value: 9, Durability: 7,
		Compatibility: 10, Portability: 2, Noise: 2,
	}, PriorityBoostPercent: 150, MaximumSetupItems: 4, CandidatesPerSlot: 12,
		OptionalSlotBonus: 8, Goals: []GoalPolicy{
			{Goal: planning.GoalBuildMuscle, Roles: []SetupRole{
				{Key: "primary", Capabilities: []Capability{CapabilityHypertrophy}, Required: true},
				{Key: "load", Capabilities: []Capability{CapabilityWeightPlates}, SortOrder: 1},
				{Key: "safety", Capabilities: []Capability{CapabilitySafeBarbell}, SortOrder: 2},
				{Key: "support", Capabilities: []Capability{CapabilitySupportedTraining}, SortOrder: 3},
			}},
			{Goal: planning.GoalStrength, Roles: []SetupRole{
				{Key: "primary", Capabilities: []Capability{CapabilityStrengthTraining}, Required: true},
				{Key: "load", Capabilities: []Capability{CapabilityWeightPlates}, SortOrder: 1},
				{Key: "safety", Capabilities: []Capability{CapabilitySafeBarbell}, SortOrder: 2},
				{Key: "support", Capabilities: []Capability{CapabilitySupportedTraining}, SortOrder: 3},
			}},
			{Goal: planning.GoalGeneralFitness, Roles: []SetupRole{
				{Key: "resistance", Capabilities: []Capability{CapabilityStrengthTraining, CapabilityResistanceTraining}, Required: true},
				{Key: "conditioning", Capabilities: []Capability{CapabilityConditioning}, SortOrder: 1},
				{Key: "mobility", Capabilities: []Capability{CapabilityMobility}, SortOrder: 2},
				{Key: "support", Capabilities: []Capability{CapabilitySupportedTraining}, SortOrder: 3},
			}},
			{Goal: planning.GoalWeightLoss, Roles: []SetupRole{
				{Key: "conditioning", Capabilities: []Capability{CapabilityConditioning}, Required: true},
				{Key: "resistance", Capabilities: []Capability{CapabilityStrengthTraining, CapabilityResistanceTraining}, SortOrder: 1},
				{Key: "mobility", Capabilities: []Capability{CapabilityMobility}, SortOrder: 2},
			}},
			{Goal: planning.GoalMobility, Roles: []SetupRole{
				{Key: "mobility", Capabilities: []Capability{CapabilityMobility}, Required: true},
				{Key: "resistance", Capabilities: []Capability{CapabilityResistanceTraining}, SortOrder: 1},
				{Key: "conditioning", Capabilities: []Capability{CapabilityConditioning}, SortOrder: 2},
			}},
		}}
}

func testInput(goal planning.Goal, experience planning.ExperienceLevel, budget int64) Input {
	return Input{
		Goal: goal, Experience: experience,
		Budget: catalog.Money{AmountMinor: budget, Currency: "USD"},
		AvailableSpace: AvailableSpace{
			LengthMM: 3000, WidthMM: 3000, HeightMM: 2600,
		},
	}
}

func testCandidate(
	id string,
	category string,
	price int64,
	scores catalog.Scores,
) CandidateSnapshot {
	candidate := CandidateSnapshot{
		ProductID: catalog.ProductID(id), Name: "Demo " + id,
		CategorySlug: category, PolicyVersion: "test-policy-v2",
		Price:      catalog.Money{AmountMinor: price, Currency: "USD"},
		Dimensions: catalog.Dimensions{LengthMM: 500, WidthMM: 400, HeightMM: 400},
		Scores:     scores,
		Space:      SpaceProfile{Footprint: SpatialEnvelope{LengthMM: 500, WidthMM: 400, HeightMM: 400}},
		GoalSupport: []GoalSupport{
			{Goal: planning.GoalBuildMuscle, Score: 80}, {Goal: planning.GoalStrength, Score: 80},
			{Goal: planning.GoalGeneralFitness, Score: 80}, {Goal: planning.GoalWeightLoss, Score: 80},
			{Goal: planning.GoalMobility, Score: 80},
		},
	}
	switch category {
	case "adjustable-dumbbells", "dumbbells":
		candidate.Capabilities = []Capability{CapabilityResistanceTraining, CapabilityStrengthTraining, CapabilityHypertrophy}
		candidate.PreferenceTags = []TrainingPreference{PreferenceDumbbells}
		candidate.RedundancyGroups = []string{"dumbbell_system"}
	case "benches":
		candidate.Capabilities = []Capability{CapabilitySupportedTraining}
		candidate.CompatibleWith = []Capability{CapabilityResistanceTraining, CapabilityBarbellTraining}
		candidate.RedundancyGroups = []string{"bench"}
	case "power-racks":
		candidate.Capabilities = []Capability{CapabilitySafeBarbell, CapabilityPullUp, CapabilityAnchorPoint}
		candidate.Requires = []Capability{CapabilityBarbellTraining}
		candidate.PreferenceTags = []TrainingPreference{PreferenceBarbell, PreferenceBodyweight}
		candidate.RedundancyGroups = []string{"rack"}
	case "barbells":
		candidate.Capabilities = []Capability{CapabilityBarbellTraining, CapabilityStrengthTraining, CapabilityHypertrophy}
		candidate.Requires = []Capability{CapabilityWeightPlates}
		candidate.PreferenceTags = []TrainingPreference{PreferenceBarbell}
		candidate.RedundancyGroups = []string{"barbell"}
	case "weight-plates":
		candidate.Capabilities = []Capability{CapabilityWeightPlates}
		candidate.Requires = []Capability{CapabilityBarbellTraining}
		candidate.PreferenceTags = []TrainingPreference{PreferenceBarbell}
		candidate.RedundancyGroups = []string{"weight_plates"}
	case "kettlebells":
		candidate.Capabilities = []Capability{CapabilityResistanceTraining, CapabilityStrengthTraining, CapabilityHypertrophy, CapabilityConditioning}
		candidate.PreferenceTags = []TrainingPreference{PreferenceKettlebell}
		candidate.RedundancyGroups = []string{"kettlebell_system"}
	case "resistance-bands":
		candidate.Capabilities = []Capability{CapabilityResistanceTraining, CapabilityHypertrophy, CapabilityConditioning, CapabilityMobility}
		candidate.CompatibleWith = []Capability{CapabilityAnchorPoint}
		candidate.PreferenceTags = []TrainingPreference{PreferenceResistanceBand}
		candidate.RedundancyGroups = []string{"resistance_band_system"}
	case "cardio-machines":
		candidate.Capabilities = []Capability{CapabilityConditioning}
		candidate.PreferenceTags = []TrainingPreference{PreferenceCardio}
		candidate.RedundancyGroups = []string{"cardio_machine"}
	}
	return candidate
}

func balancedScores() catalog.Scores {
	return catalog.Scores{
		Quality: 80, Value: 80, Durability: 80, Beginner: 80,
		Advanced: 80, Apartment: 80, Noise: 80, Portability: 80,
	}
}

func selectedIDs(result Result) map[catalog.ProductID]bool {
	identifiers := make(map[catalog.ProductID]bool, len(result.Selected))
	for _, item := range result.Selected {
		identifiers[item.Product.Candidate.ProductID] = true
	}
	return identifiers
}

func rejectionFor(result Result, productID catalog.ProductID) (RejectedProduct, bool) {
	for _, rejection := range result.Rejected {
		if rejection.Candidate.ProductID == productID {
			return rejection, true
		}
	}
	return RejectedProduct{}, false
}

func hasReason(product RankedProduct, code string) bool {
	for _, reason := range product.Reasons {
		if reason.Code == code {
			return true
		}
	}
	return false
}
