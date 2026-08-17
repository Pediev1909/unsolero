package domain

import (
	"sort"

	planning "rigmark/internal/modules/planning/domain"
)

func scoreCandidates(input Input, candidates []eligibleCandidate, config Config) []RankedProduct {
	ranked := make([]RankedProduct, 0, len(candidates))
	for _, eligible := range candidates {
		breakdown := scoreBreakdown(input, eligible.Candidate, eligible.Existing)
		ranked = append(ranked, RankedProduct{
			Candidate:      eligible.Candidate,
			ObjectiveScore: objectiveScore(breakdown, config, input.Priorities),
			Breakdown:      breakdown,
			Reasons:        reasonsFor(input, eligible.Candidate, eligible.Existing, breakdown),
		})
	}
	sortRanked(ranked)
	return ranked
}

func scoreBreakdown(
	input Input,
	candidate CandidateSnapshot,
	compatibleEquipment *ExistingEquipment,
) ScoreBreakdown {
	return ScoreBreakdown{
		GoalMatch:       goalScore(input.Goal, candidate),
		BudgetMatch:     budgetScore(input.Budget.AmountMinor, candidate.Price.AmountMinor),
		SpaceMatch:      spaceScore(input.AvailableSpace, candidate),
		ExperienceMatch: experienceScore(input.Experience, candidate),
		PreferenceMatch: preferenceScore(input.TrainingPreferences, candidate),
		Quality:         int(candidate.Scores.Quality),
		Value:           int(candidate.Scores.Value),
		Durability:      int(candidate.Scores.Durability),
		Compatibility:   compatibilityScore(candidate, compatibleEquipment, input.ExistingEquipment),
		Portability:     int(candidate.Scores.Portability),
		Noise:           int(candidate.Scores.Noise),
	}
}

func objectiveScore(breakdown ScoreBreakdown, config Config, priorities []Priority) int {
	weights := config.Weights
	for _, priority := range priorities {
		switch priority {
		case PriorityBudget:
			weights.BudgetMatch = boostWeight(weights.BudgetMatch, config.PriorityBoostPercent)
			weights.Value = boostWeight(weights.Value, config.PriorityBoostPercent)
		case PriorityCompact:
			weights.SpaceMatch = boostWeight(weights.SpaceMatch, config.PriorityBoostPercent)
		case PriorityQuality:
			weights.Quality = boostWeight(weights.Quality, config.PriorityBoostPercent)
		case PriorityDurability:
			weights.Durability = boostWeight(weights.Durability, config.PriorityBoostPercent)
		case PriorityQuiet:
			weights.Noise = boostWeight(weights.Noise, config.PriorityBoostPercent)
		case PriorityPortability:
			weights.Portability = boostWeight(weights.Portability, config.PriorityBoostPercent)
		}
	}
	values := breakdownValues(breakdown)
	weightList := weightValues(weights)
	totalWeight := 0
	weightedScore := 0
	for index, value := range values {
		weightedScore += value * weightList[index]
		totalWeight += weightList[index]
	}
	return clampScore((weightedScore + totalWeight/2) / totalWeight)
}

func goalScore(goal planning.Goal, candidate CandidateSnapshot) int {
	for _, support := range candidate.GoalSupport {
		if support.Goal == goal {
			return support.Score
		}
	}
	return 0
}

func budgetScore(budget, price int64) int {
	if price > budget {
		return 0
	}
	usagePercent := int(price * 100 / budget)
	return clampScore(100 - usagePercent/4)
}

func spaceScore(space AvailableSpace, candidate CandidateSnapshot) int {
	dimensions, known := requiredEnvelope(candidate)
	if !known {
		return 0
	}
	direct := maxInt(
		percentage(dimensions.LengthMM, space.LengthMM),
		percentage(dimensions.WidthMM, space.WidthMM),
	)
	rotated := maxInt(
		percentage(dimensions.WidthMM, space.LengthMM),
		percentage(dimensions.LengthMM, space.WidthMM),
	)
	utilization := minInt(direct, rotated)
	utilization = maxInt(utilization, percentage(dimensions.HeightMM, space.HeightMM))
	fit := clampScore(100 - utilization*30/100)
	if space.ApartmentLiving {
		fit = (fit + int(candidate.Scores.Apartment) + 1) / 2
	}
	return clampScore(fit)
}

func experienceScore(experience planning.ExperienceLevel, candidate CandidateSnapshot) int {
	switch experience {
	case planning.ExperienceBeginner:
		return int(candidate.Scores.Beginner)
	case planning.ExperienceAdvanced:
		return int(candidate.Scores.Advanced)
	default:
		return (int(candidate.Scores.Beginner) + int(candidate.Scores.Advanced) + 1) / 2
	}
}

func preferenceScore(preferences []TrainingPreference, candidate CandidateSnapshot) int {
	if len(preferences) == 0 {
		return 75
	}
	for _, preference := range preferences {
		if preferenceMatches(preference, candidate) {
			return 100
		}
	}
	return 50
}

func preferenceMatches(preference TrainingPreference, candidate CandidateSnapshot) bool {
	for _, tag := range candidate.PreferenceTags {
		if tag == preference {
			return true
		}
	}
	return false
}

func compatibilityScore(
	candidate CandidateSnapshot,
	compatibleEquipment *ExistingEquipment,
	existing []ExistingEquipment,
) int {
	if compatibleEquipment != nil {
		return 100
	}
	existingCapabilities := equipmentCapabilities(existing)
	if len(candidate.Requires) == 0 {
		return 85
	}
	for _, required := range candidate.Requires {
		if !containsCapability(existingCapabilities, required) {
			return 65
		}
	}
	return 95
}

func sortRanked(ranked []RankedProduct) {
	sort.SliceStable(ranked, func(left, right int) bool {
		if ranked[left].ObjectiveScore != ranked[right].ObjectiveScore {
			return ranked[left].ObjectiveScore > ranked[right].ObjectiveScore
		}
		if ranked[left].Breakdown.GoalMatch != ranked[right].Breakdown.GoalMatch {
			return ranked[left].Breakdown.GoalMatch > ranked[right].Breakdown.GoalMatch
		}
		if ranked[left].Breakdown.Value != ranked[right].Breakdown.Value {
			return ranked[left].Breakdown.Value > ranked[right].Breakdown.Value
		}
		if ranked[left].Breakdown.Quality != ranked[right].Breakdown.Quality {
			return ranked[left].Breakdown.Quality > ranked[right].Breakdown.Quality
		}
		if ranked[left].Candidate.Price.AmountMinor != ranked[right].Candidate.Price.AmountMinor {
			return ranked[left].Candidate.Price.AmountMinor < ranked[right].Candidate.Price.AmountMinor
		}
		return ranked[left].Candidate.ProductID < ranked[right].Candidate.ProductID
	})
}

func weightValues(weights Weights) []int {
	return []int{
		weights.GoalMatch, weights.BudgetMatch, weights.SpaceMatch,
		weights.ExperienceMatch, weights.PreferenceMatch, weights.Quality,
		weights.Value, weights.Durability, weights.Compatibility,
		weights.Portability, weights.Noise,
	}
}

func breakdownValues(breakdown ScoreBreakdown) []int {
	return []int{
		breakdown.GoalMatch, breakdown.BudgetMatch, breakdown.SpaceMatch,
		breakdown.ExperienceMatch, breakdown.PreferenceMatch, breakdown.Quality,
		breakdown.Value, breakdown.Durability, breakdown.Compatibility,
		breakdown.Portability, breakdown.Noise,
	}
}

func boostWeight(weight, percent int) int {
	return weight * percent / 100
}

func percentage(value, total int64) int {
	return int(value * 100 / total)
}

func clampScore(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
