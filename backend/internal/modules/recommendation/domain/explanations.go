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
	config Config,
) []Reason {
	reasons := []Reason{
		{
			Code: "budget.within", Message: "Within your budget",
			Dimension: "budget_match", Score: breakdown.BudgetMatch,
		},
	}
	// Only a vertical whose products occupy space can claim one fits. In a
	// software catalog the score is a constant, and offering it as a reason
	// told the reader something that was not about the product at all.
	if config.SpatialConstraints {
		reasons = append([]Reason{{
			Code: "space.fits", Message: "Fits your available space",
			Dimension: "space_match", Score: breakdown.SpaceMatch,
		}}, reasons...)
	}
	if breakdown.GoalMatch >= 85 {
		reasons = append(reasons, Reason{
			Code: "goal.strong_match", Message: "Strong match for " + goalLabel(config.Goals, input.Goal),
			Dimension: "goal_match", Score: breakdown.GoalMatch,
		})
	}
	if breakdown.ExperienceMatch >= 85 {
		reasons = append(reasons, Reason{
			Code:      "experience.strong_match",
			Message:   fmt.Sprintf("Well suited to a %s team", input.Experience),
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
		reason, include := priorityReason(priority, breakdown, config)
		if include {
			reasons = append(reasons, reason)
		}
	}
	return reasons
}

func goalLabel(goals []GoalPolicy, goal planning.Goal) string {
	for _, declared := range goals {
		if declared.Goal == goal {
			return declared.Label
		}
	}
	return "your goal"
}

// priorityReason renders the explanation the policy declares for a priority.
// The message and threshold are policy data so that a vertical can explain a
// priority in its own language without an engine change.
func priorityReason(priority Priority, breakdown ScoreBreakdown, config Config) (Reason, bool) {
	for _, policy := range config.Priorities {
		if policy.Key != priority {
			continue
		}
		score := breakdownValue(breakdown, policy.ReasonDimension)
		return Reason{
			Code:      policy.ReasonCode,
			Message:   policy.ReasonMessage,
			Dimension: string(policy.ReasonDimension),
			Score:     score,
		}, score >= policy.ReasonThreshold
	}
	return Reason{}, false
}
