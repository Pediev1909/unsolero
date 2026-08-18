package domain

// Dimension names a single scored dimension of a recommendation. Policy data
// references dimensions by key so that priority behaviour can be defined per
// vertical instead of being hard-coded in the engine.
type Dimension string

const (
	DimensionGoalMatch       Dimension = "goal_match"
	DimensionBudgetMatch     Dimension = "budget_match"
	DimensionSpaceMatch      Dimension = "space_match"
	DimensionExperienceMatch Dimension = "experience_match"
	DimensionPreferenceMatch Dimension = "preference_match"
	DimensionQuality         Dimension = "quality"
	DimensionValue           Dimension = "value"
	DimensionDurability      Dimension = "durability"
	DimensionCompatibility   Dimension = "compatibility"
	DimensionPortability     Dimension = "portability"
	DimensionNoise           Dimension = "noise"
)

// dimensionOrder is the canonical scoring order. weightValues and
// breakdownValues must emit values in exactly this order because the engine
// relies on index alignment between the two slices. Changing this order
// changes every produced score, so it is covered by a determinism test.
var dimensionOrder = []Dimension{
	DimensionGoalMatch,
	DimensionBudgetMatch,
	DimensionSpaceMatch,
	DimensionExperienceMatch,
	DimensionPreferenceMatch,
	DimensionQuality,
	DimensionValue,
	DimensionDurability,
	DimensionCompatibility,
	DimensionPortability,
	DimensionNoise,
}

func validDimension(dimension Dimension) bool {
	for _, known := range dimensionOrder {
		if known == dimension {
			return true
		}
	}
	return false
}

// breakdownValue reads one dimension from a breakdown by key. It returns zero
// for an unknown dimension; callers validate the key against validDimension
// before scoring, so an unknown key here means the policy passed validation
// incorrectly rather than that zero is a meaningful score.
func breakdownValue(breakdown ScoreBreakdown, dimension Dimension) int {
	values := breakdownValues(breakdown)
	for index, known := range dimensionOrder {
		if known == dimension {
			return values[index]
		}
	}
	return 0
}

// boostDimension scales one weight by percent and returns the updated set.
// Weights is a value type, so the caller receives a copy and the engine's
// configured weights are never mutated between runs.
func boostDimension(weights Weights, dimension Dimension, percent int) Weights {
	switch dimension {
	case DimensionGoalMatch:
		weights.GoalMatch = boostWeight(weights.GoalMatch, percent)
	case DimensionBudgetMatch:
		weights.BudgetMatch = boostWeight(weights.BudgetMatch, percent)
	case DimensionSpaceMatch:
		weights.SpaceMatch = boostWeight(weights.SpaceMatch, percent)
	case DimensionExperienceMatch:
		weights.ExperienceMatch = boostWeight(weights.ExperienceMatch, percent)
	case DimensionPreferenceMatch:
		weights.PreferenceMatch = boostWeight(weights.PreferenceMatch, percent)
	case DimensionQuality:
		weights.Quality = boostWeight(weights.Quality, percent)
	case DimensionValue:
		weights.Value = boostWeight(weights.Value, percent)
	case DimensionDurability:
		weights.Durability = boostWeight(weights.Durability, percent)
	case DimensionCompatibility:
		weights.Compatibility = boostWeight(weights.Compatibility, percent)
	case DimensionPortability:
		weights.Portability = boostWeight(weights.Portability, percent)
	case DimensionNoise:
		weights.Noise = boostWeight(weights.Noise, percent)
	}
	return weights
}
