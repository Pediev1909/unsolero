package domain

import (
	"fmt"

	planning "rigmark/internal/modules/planning/domain"
)

func reasonsFor(
	input Input,
	candidate CandidateSnapshot,
	compatibleEquipment *ExistingEquipment,
	breakdown ScoreBreakdown,
) []Reason {
	reasons := []Reason{
		{
			Code: "space.fits", Message: "Fits your available space",
			Dimension: "space_match", Score: breakdown.SpaceMatch,
		},
		{
			Code: "budget.within", Message: "Within your budget",
			Dimension: "budget_match", Score: breakdown.BudgetMatch,
		},
	}
	if breakdown.GoalMatch >= 85 {
		reasons = append(reasons, Reason{
			Code: "goal.strong_match", Message: "Strong match for " + goalLabel(input.Goal),
			Dimension: "goal_match", Score: breakdown.GoalMatch,
		})
	}
	if breakdown.ExperienceMatch >= 85 {
		reasons = append(reasons, Reason{
			Code:      "experience.strong_match",
			Message:   fmt.Sprintf("Well suited to %s training", input.Experience),
			Dimension: "experience_match", Score: breakdown.ExperienceMatch,
		})
	}
	if compatibleEquipment != nil {
		reasons = append(reasons, Reason{
			Code:      "compatibility.existing_equipment",
			Message:   "Compatible with your existing " + compatibleEquipment.Name,
			Dimension: "compatibility", Score: breakdown.Compatibility,
		})
	}
	for _, priority := range input.Priorities {
		reason, include := priorityReason(priority, breakdown)
		if include {
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

func goalLabel(goal planning.Goal) string {
	switch goal {
	case planning.GoalBuildMuscle:
		return "hypertrophy"
	case planning.GoalStrength:
		return "strength training"
	case planning.GoalGeneralFitness:
		return "general fitness"
	case planning.GoalWeightLoss:
		return "conditioning"
	case planning.GoalMobility:
		return "mobility work"
	default:
		return "your goal"
	}
}

func priorityReason(priority Priority, breakdown ScoreBreakdown) (Reason, bool) {
	switch priority {
	case PriorityBudget:
		return Reason{Code: "priority.value", Message: "Strong value for the available budget", Dimension: "value", Score: breakdown.Value}, breakdown.Value >= 85
	case PriorityCompact:
		return Reason{Code: "priority.compact", Message: "Uses your available space efficiently", Dimension: "space_match", Score: breakdown.SpaceMatch}, breakdown.SpaceMatch >= 85
	case PriorityQuality:
		return Reason{Code: "priority.quality", Message: "Strong structured quality score", Dimension: "quality", Score: breakdown.Quality}, breakdown.Quality >= 85
	case PriorityDurability:
		return Reason{Code: "priority.durability", Message: "Strong structured durability score", Dimension: "durability", Score: breakdown.Durability}, breakdown.Durability >= 85
	case PriorityQuiet:
		return Reason{Code: "priority.quiet", Message: "Well suited to quieter training", Dimension: "noise", Score: breakdown.Noise}, breakdown.Noise >= 85
	case PriorityPortability:
		return Reason{Code: "priority.portable", Message: "Easy to move or store", Dimension: "portability", Score: breakdown.Portability}, breakdown.Portability >= 85
	default:
		return Reason{}, false
	}
}
