package domain

import (
	"fmt"

	planning "rigmark/internal/modules/planning/domain"
)

type eligibleCandidate struct {
	Candidate CandidateSnapshot
	Existing  *ExistingEquipment
}

func filterEligible(
	input Input,
	candidates []CandidateSnapshot,
	existing []ExistingEquipment,
	config Config,
) ([]eligibleCandidate, []RejectedProduct) {
	availableCapabilities := equipmentCapabilities(existing)
	availableRedundancyGroups := equipmentRedundancyGroups(existing)
	providerCapabilities := append([]Capability(nil), availableCapabilities...)
	pending := make([]CandidateSnapshot, 0, len(candidates))
	rejected := make([]RejectedProduct, 0)
	for _, candidate := range candidates {
		code, message := basicHardRejection(
			input,
			candidate,
			availableRedundancyGroups,
			availableCapabilities,
			config,
		)
		if code != "" {
			rejected = append(rejected, RejectedProduct{
				Candidate: candidate,
				Code:      code,
				Message:   message,
			})
			continue
		}
		pending = append(pending, candidate)
		providerCapabilities = mergedCapabilities(providerCapabilities, candidate.Capabilities)
	}

	eligible := make([]eligibleCandidate, 0, len(pending))
	for _, candidate := range pending {
		missing := missingRequirement(candidate, providerCapabilities)
		if missing != "" {
			rejected = append(rejected, RejectedProduct{
				Candidate: candidate,
				Code:      "compatibility.missing_requirement",
				Message:   fmt.Sprintf("Requires unavailable capability %s", missing),
			})
			continue
		}
		eligible = append(eligible, eligibleCandidate{
			Candidate: candidate,
			Existing:  compatibleExisting(candidate, existing),
		})
	}
	return eligible, rejected
}

func basicHardRejection(
	input Input,
	candidate CandidateSnapshot,
	existingRedundancyGroups []string,
	existingCapabilities []Capability,
	config Config,
) (string, string) {
	if !supportsGoal(candidate, input.Goal) {
		return "goal.unsupported", "The active policy does not support this product for your goal"
	}
	if candidate.Price.Currency != input.Budget.Currency {
		return "currency.mismatch", "Uses a different currency from your budget"
	}
	if candidate.Price.AmountMinor > input.Budget.AmountMinor {
		return "budget.exceeded", "Costs more than your total budget"
	}
	// Space is only a constraint where products occupy space. Without this a
	// non-physical vertical rejects its whole catalog for having no footprint.
	if config.SpatialConstraints {
		if code, message := spaceRejection(candidate, input.AvailableSpace); code != "" {
			return code, message
		}
	}
	if intersectsStrings(candidate.RedundancyGroups, existingRedundancyGroups) {
		return "existing_equipment.redundant", "Duplicates equipment you already own"
	}
	if intersects(candidate.IncompatibleWith, existingCapabilities) {
		return "compatibility.existing_conflict", "Conflicts with your existing equipment"
	}
	return "", ""
}

func supportsGoal(candidate CandidateSnapshot, goal planning.Goal) bool {
	for _, support := range candidate.GoalSupport {
		if support.Goal == goal {
			return true
		}
	}
	return false
}

func missingRequirement(candidate CandidateSnapshot, capabilities []Capability) Capability {
	for _, required := range candidate.Requires {
		if !containsCapability(capabilities, required) {
			return required
		}
	}
	return ""
}

func compatibleExisting(candidate CandidateSnapshot, existing []ExistingEquipment) *ExistingEquipment {
	for index := range existing {
		if intersects(candidate.CompatibleWith, existing[index].Capabilities) {
			return &existing[index]
		}
	}
	return nil
}

func equipmentCapabilities(existing []ExistingEquipment) []Capability {
	result := make([]Capability, 0)
	for _, equipment := range existing {
		result = mergedCapabilities(result, equipment.Capabilities)
	}
	return result
}

func equipmentRedundancyGroups(existing []ExistingEquipment) []string {
	result := make([]string, 0)
	for _, equipment := range existing {
		result = append(result, equipment.RedundancyGroups...)
	}
	return mergedStrings(result)
}

func containsCapability(capabilities []Capability, wanted Capability) bool {
	for _, capability := range capabilities {
		if capability == wanted {
			return true
		}
	}
	return false
}

func intersects(left, right []Capability) bool {
	for _, leftCapability := range left {
		if containsCapability(right, leftCapability) {
			return true
		}
	}
	return false
}

func intersectsStrings(left, right []string) bool {
	for _, value := range left {
		for _, other := range right {
			if value == other {
				return true
			}
		}
	}
	return false
}
