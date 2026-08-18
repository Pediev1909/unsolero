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
	return Config{PolicyVersion: version, SpatialConstraints: true, Weights: Weights{
		GoalMatch: 20, BudgetMatch: 12, SpaceMatch: 12, ExperienceMatch: 10,
		PreferenceMatch: 8, Quality: 8, Value: 9, Durability: 7,
		Compatibility: 10, Portability: 2, Noise: 2,
	}, PriorityBoostPercent: 150, MaximumSetupItems: 4, CandidatesPerSlot: 12,
		OptionalSlotBonus: 8,
		PreferenceTags: []TrainingPreference{
			"dumbbells", "barbell", "kettlebell", "resistance_bands",
			"bodyweight", "cardio", "low_impact",
		},
		Priorities: testPriorityPolicies(),
		Goals: []GoalPolicy{
			{Goal: planning.GoalBuildMuscle, Label: "hypertrophy", Roles: []SetupRole{
				{Key: "primary", Capabilities: []Capability{Capability("hypertrophy")}, Required: true},
				{Key: "load", Capabilities: []Capability{Capability("weight_plates")}, SortOrder: 1},
				{Key: "safety", Capabilities: []Capability{Capability("safe_barbell_training")}, SortOrder: 2},
				{Key: "support", Capabilities: []Capability{Capability("supported_training")}, SortOrder: 3},
			}},
			{Goal: planning.GoalStrength, Label: "strength work", Roles: []SetupRole{
				{Key: "primary", Capabilities: []Capability{Capability("strength_training")}, Required: true},
				{Key: "load", Capabilities: []Capability{Capability("weight_plates")}, SortOrder: 1},
				{Key: "safety", Capabilities: []Capability{Capability("safe_barbell_training")}, SortOrder: 2},
				{Key: "support", Capabilities: []Capability{Capability("supported_training")}, SortOrder: 3},
			}},
			{Goal: planning.GoalGeneralFitness, Label: "general fitness", Roles: []SetupRole{
				{Key: "resistance", Capabilities: []Capability{Capability("strength_training"), Capability("resistance_training")}, Required: true},
				{Key: "conditioning", Capabilities: []Capability{Capability("conditioning")}, SortOrder: 1},
				{Key: "mobility", Capabilities: []Capability{Capability("mobility")}, SortOrder: 2},
				{Key: "support", Capabilities: []Capability{Capability("supported_training")}, SortOrder: 3},
			}},
			{Goal: planning.GoalWeightLoss, Label: "conditioning", Roles: []SetupRole{
				{Key: "conditioning", Capabilities: []Capability{Capability("conditioning")}, Required: true},
				{Key: "resistance", Capabilities: []Capability{Capability("strength_training"), Capability("resistance_training")}, SortOrder: 1},
				{Key: "mobility", Capabilities: []Capability{Capability("mobility")}, SortOrder: 2},
			}},
			{Goal: planning.GoalMobility, Label: "mobility work", Roles: []SetupRole{
				{Key: "mobility", Capabilities: []Capability{Capability("mobility")}, Required: true},
				{Key: "resistance", Capabilities: []Capability{Capability("resistance_training")}, SortOrder: 1},
				{Key: "conditioning", Capabilities: []Capability{Capability("conditioning")}, SortOrder: 2},
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
		candidate.Capabilities = []Capability{Capability("resistance_training"), Capability("strength_training"), Capability("hypertrophy")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("dumbbells")}
		candidate.RedundancyGroups = []string{"dumbbell_system"}
	case "benches":
		candidate.Capabilities = []Capability{Capability("supported_training")}
		candidate.CompatibleWith = []Capability{Capability("resistance_training"), Capability("barbell_training")}
		candidate.RedundancyGroups = []string{"bench"}
	case "power-racks":
		candidate.Capabilities = []Capability{Capability("safe_barbell_training"), Capability("pull_up"), Capability("anchor_point")}
		candidate.Requires = []Capability{Capability("barbell_training")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("barbell"), TrainingPreference("bodyweight")}
		candidate.RedundancyGroups = []string{"rack"}
	case "barbells":
		candidate.Capabilities = []Capability{Capability("barbell_training"), Capability("strength_training"), Capability("hypertrophy")}
		candidate.Requires = []Capability{Capability("weight_plates")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("barbell")}
		candidate.RedundancyGroups = []string{"barbell"}
	case "weight-plates":
		candidate.Capabilities = []Capability{Capability("weight_plates")}
		candidate.Requires = []Capability{Capability("barbell_training")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("barbell")}
		candidate.RedundancyGroups = []string{"weight_plates"}
	case "kettlebells":
		candidate.Capabilities = []Capability{Capability("resistance_training"), Capability("strength_training"), Capability("hypertrophy"), Capability("conditioning")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("kettlebell")}
		candidate.RedundancyGroups = []string{"kettlebell_system"}
	case "resistance-bands":
		candidate.Capabilities = []Capability{Capability("resistance_training"), Capability("hypertrophy"), Capability("conditioning"), Capability("mobility")}
		candidate.CompatibleWith = []Capability{Capability("anchor_point")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("resistance_bands")}
		candidate.RedundancyGroups = []string{"resistance_band_system"}
	case "cardio-machines":
		candidate.Capabilities = []Capability{Capability("conditioning")}
		candidate.PreferenceTags = []TrainingPreference{TrainingPreference("cardio")}
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

// testPriorityPolicies reproduces the priority behaviour the engine used to
// hard-code, now expressed as the policy data a fitness vertical would ship.
func testPriorityPolicies() []PriorityPolicy {
	return []PriorityPolicy{
		{Key: "budget", BoostDimensions: []Dimension{DimensionBudgetMatch, DimensionValue},
			ReasonCode: "priority.value", ReasonMessage: "Strong value for the available budget",
			ReasonDimension: DimensionValue, ReasonThreshold: 85},
		{Key: "compact", BoostDimensions: []Dimension{DimensionSpaceMatch},
			ReasonCode: "priority.compact", ReasonMessage: "Uses your available space efficiently",
			ReasonDimension: DimensionSpaceMatch, ReasonThreshold: 85},
		{Key: "quality", BoostDimensions: []Dimension{DimensionQuality},
			ReasonCode: "priority.quality", ReasonMessage: "Strong structured quality score",
			ReasonDimension: DimensionQuality, ReasonThreshold: 85},
		{Key: "durability", BoostDimensions: []Dimension{DimensionDurability},
			ReasonCode: "priority.durability", ReasonMessage: "Strong structured durability score",
			ReasonDimension: DimensionDurability, ReasonThreshold: 85},
		{Key: "quiet", BoostDimensions: []Dimension{DimensionNoise},
			ReasonCode: "priority.quiet", ReasonMessage: "Well suited to quieter training",
			ReasonDimension: DimensionNoise, ReasonThreshold: 85},
		{Key: "portability", BoostDimensions: []Dimension{DimensionPortability},
			ReasonCode: "priority.portable", ReasonMessage: "Easy to move or store",
			ReasonDimension: DimensionPortability, ReasonThreshold: 85},
	}
}
