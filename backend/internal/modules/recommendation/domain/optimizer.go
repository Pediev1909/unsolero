package domain

import (
	"sort"
	"strings"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

type setupOptimizer interface {
	Optimize(optimizationInput) setupSelection
}

type deterministicSetupOptimizer struct{}

type optimizationInput struct {
	Goal       planning.Goal
	Budget     int64
	Space      AvailableSpace
	Existing   []ExistingEquipment
	Candidates []RankedProduct
	Config     Config
}

type setupSelection struct {
	Products []RankedProduct
	Total    int64
	Complete bool
	utility  int
	key      string
}

type setupSlot struct {
	Name         string
	Capabilities []Capability
	Required     bool
}

func (deterministicSetupOptimizer) Optimize(input optimizationInput) setupSelection {
	slots := slotsForGoal(input.Config, input.Goal)
	existingCapabilities := equipmentCapabilities(input.Existing)
	existingRedundancyGroups := equipmentRedundancyGroups(input.Existing)
	best := setupSelection{utility: -1}

	var search func(int, []RankedProduct, map[string]bool, int64, int)
	search = func(
		slotIndex int,
		selected []RankedProduct,
		selectedRedundancyGroups map[string]bool,
		total int64,
		optionalCount int,
	) {
		if slotIndex == len(slots) || len(selected) == input.Config.MaximumSetupItems {
			if requiredSlotsCovered(slots[slotIndex:], existingCapabilities, selected) &&
				validCombination(selected, existingCapabilities) {
				considerSelection(&best, selected, total, optionalCount, true, input.Config)
			}
			return
		}

		slot := slots[slotIndex]
		if intersects(slot.Capabilities, existingCapabilities) ||
			selectedCoverSlot(selected, slot) {
			search(slotIndex+1, selected, selectedRedundancyGroups, total, optionalCount)
			return
		}
		if !slot.Required {
			search(slotIndex+1, selected, selectedRedundancyGroups, total, optionalCount)
		}

		branches := 0
		for _, candidate := range input.Candidates {
			if !intersects(slot.Capabilities, candidate.Candidate.Capabilities) ||
				intersectsStrings(candidate.Candidate.RedundancyGroups, existingRedundancyGroups) ||
				intersectsStringSet(candidate.Candidate.RedundancyGroups, selectedRedundancyGroups) ||
				selectedContains(selected, candidate.Candidate.ProductID) ||
				candidate.Candidate.Price.AmountMinor > input.Budget-total ||
				!compatibleWithSelection(candidate.Candidate, selected, existingCapabilities) ||
				!fitsWithinFloorArea(appendCandidate(selected, candidate), input.Space, input.Config) {
				continue
			}

			next := appendCopy(selected, candidate)
			nextRedundancyGroups := copyStringSet(selectedRedundancyGroups)
			for _, group := range candidate.Candidate.RedundancyGroups {
				nextRedundancyGroups[group] = true
			}
			nextOptionalCount := optionalCount
			if !slot.Required {
				nextOptionalCount++
			}
			search(
				slotIndex+1,
				next,
				nextRedundancyGroups,
				total+candidate.Candidate.Price.AmountMinor,
				nextOptionalCount,
			)
			branches++
			if branches >= input.Config.CandidatesPerSlot {
				break
			}
		}
	}

	search(0, nil, make(map[string]bool), 0, 0)
	if best.utility < 0 {
		return setupSelection{}
	}
	return best
}

func slotsForGoal(config Config, goal planning.Goal) []setupSlot {
	for _, policy := range config.Goals {
		if policy.Goal != goal {
			continue
		}
		roles := append([]SetupRole(nil), policy.Roles...)
		sort.Slice(roles, func(left, right int) bool {
			if roles[left].SortOrder != roles[right].SortOrder {
				return roles[left].SortOrder < roles[right].SortOrder
			}
			return roles[left].Key < roles[right].Key
		})
		result := make([]setupSlot, 0, len(roles))
		for _, role := range roles {
			result = append(result, setupSlot{Name: role.Key, Capabilities: role.Capabilities, Required: role.Required})
		}
		return result
	}
	return nil
}

func considerSelection(
	best *setupSelection,
	selected []RankedProduct,
	total int64,
	optionalCount int,
	complete bool,
	config Config,
) {
	utility := selectionUtility(selected, optionalCount, config)
	key := selectionKey(selected)
	if utility > best.utility ||
		utility == best.utility && len(selected) > len(best.Products) ||
		utility == best.utility && len(selected) == len(best.Products) && total < best.Total ||
		utility == best.utility && len(selected) == len(best.Products) && total == best.Total && key < best.key {
		best.Products = append([]RankedProduct(nil), selected...)
		best.Total = total
		best.Complete = complete
		best.utility = utility
		best.key = key
	}
}

func selectionUtility(selected []RankedProduct, optionalCount int, config Config) int {
	if len(selected) == 0 {
		return optionalCount * config.OptionalSlotBonus * 100
	}
	score := 0
	for _, product := range selected {
		score += product.ObjectiveScore
	}
	average := (score + len(selected)/2) / len(selected)
	return average*100 + optionalCount*config.OptionalSlotBonus*100
}

func validCombination(selected []RankedProduct, existingCapabilities []Capability) bool {
	capabilities := append([]Capability(nil), existingCapabilities...)
	for _, product := range selected {
		capabilities = mergedCapabilities(capabilities, product.Candidate.Capabilities)
	}
	for index, product := range selected {
		for _, required := range product.Candidate.Requires {
			if !containsCapability(capabilities, required) {
				return false
			}
		}
		for other := index + 1; other < len(selected); other++ {
			if intersects(product.Candidate.IncompatibleWith, selected[other].Candidate.Capabilities) ||
				intersects(selected[other].Candidate.IncompatibleWith, product.Candidate.Capabilities) {
				return false
			}
		}
	}
	return true
}

func compatibleWithSelection(
	candidate CandidateSnapshot,
	selected []RankedProduct,
	existingCapabilities []Capability,
) bool {
	if intersects(candidate.IncompatibleWith, existingCapabilities) {
		return false
	}
	for _, product := range selected {
		if intersects(candidate.IncompatibleWith, product.Candidate.Capabilities) ||
			intersects(product.Candidate.IncompatibleWith, candidate.Capabilities) {
			return false
		}
	}
	return true
}

func requiredSlotsCovered(
	slots []setupSlot,
	existingCapabilities []Capability,
	selected []RankedProduct,
) bool {
	for _, slot := range slots {
		if slot.Required && !intersects(slot.Capabilities, existingCapabilities) &&
			!selectedCoverSlot(selected, slot) {
			return false
		}
	}
	return true
}

func selectedCoverSlot(selected []RankedProduct, slot setupSlot) bool {
	for _, product := range selected {
		if intersects(slot.Capabilities, product.Candidate.Capabilities) {
			return true
		}
	}
	return false
}

func selectedContains(selected []RankedProduct, productID catalog.ProductID) bool {
	for _, product := range selected {
		if product.Candidate.ProductID == productID {
			return true
		}
	}
	return false
}

func appendCopy(selected []RankedProduct, product RankedProduct) []RankedProduct {
	result := make([]RankedProduct, len(selected), len(selected)+1)
	copy(result, selected)
	return append(result, product)
}

func appendCandidate(selected []RankedProduct, product RankedProduct) []RankedProduct {
	return appendCopy(selected, product)
}

// fitsWithinTotalFloorArea is a conservative setup-level constraint. Individual
// products may each fit while their combined footprints do not. The catalog does
// not yet model safe overlap zones, so summing footprints is the only defensible
// deterministic rule.
// fitsWithinFloorArea applies the floor-area constraint only where products
// occupy floor. A non-physical vertical has no floor to run out of, so the
// constraint is skipped rather than evaluated against a zero-sized room.
func fitsWithinFloorArea(products []RankedProduct, space AvailableSpace, config Config) bool {
	if !config.SpatialConstraints {
		return true
	}
	return fitsWithinTotalFloorArea(products, space)
}

func fitsWithinTotalFloorArea(products []RankedProduct, space AvailableSpace) bool {
	availableArea := space.LengthMM * space.WidthMM
	var requiredArea int64
	overlapAreas := make(map[string]int64)
	for _, product := range products {
		envelope, known := requiredEnvelope(product.Candidate)
		if !known {
			return false
		}
		area := envelope.LengthMM * envelope.WidthMM
		group := product.Candidate.Space.OverlapGroup
		if group == "" {
			requiredArea += area
		} else if area > overlapAreas[group] {
			overlapAreas[group] = area
		}
	}
	for _, area := range overlapAreas {
		requiredArea += area
	}
	return requiredArea <= availableArea
}

func intersectsStringSet(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}

func copyStringSet(source map[string]bool) map[string]bool {
	result := make(map[string]bool, len(source)+1)
	for key, value := range source {
		result[key] = value
	}
	return result
}

func selectionKey(selected []RankedProduct) string {
	identifiers := make([]string, 0, len(selected))
	for _, product := range selected {
		identifiers = append(identifiers, string(product.Candidate.ProductID))
	}
	sort.Strings(identifiers)
	return strings.Join(identifiers, "|")
}
