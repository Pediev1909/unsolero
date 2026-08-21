package domain

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

var normalizedCode = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
var normalizedSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func validateConfig(config Config) error {
	if strings.TrimSpace(config.PolicyVersion) == "" ||
		config.PriorityBoostPercent < 100 || config.PriorityBoostPercent > 300 ||
		config.MaximumSetupItems < 1 || config.MaximumSetupItems > 10 ||
		config.CandidatesPerSlot < 1 || config.CandidatesPerSlot > 50 ||
		config.OptionalSlotBonus < 0 || config.OptionalSlotBonus > 100 {
		return ErrInvalidConfig
	}
	weights := weightValues(config.Weights)
	total := 0
	for _, weight := range weights {
		if weight < 0 || weight > 1000 {
			return ErrInvalidConfig
		}
		total += weight
	}
	if total == 0 {
		return ErrInvalidConfig
	}
	if len(config.Goals) == 0 {
		return ErrInvalidConfig
	}
	seenGoals := make(map[planning.Goal]bool, len(config.Goals))
	for _, goal := range config.Goals {
		if !validCode(string(goal.Goal)) || strings.TrimSpace(goal.Label) == "" ||
			seenGoals[goal.Goal] || len(goal.Roles) == 0 {
			return ErrInvalidConfig
		}
		seenGoals[goal.Goal] = true
		seenRoles := make(map[string]bool, len(goal.Roles))
		for _, role := range goal.Roles {
			if !validCode(role.Key) || seenRoles[role.Key] || len(role.Capabilities) == 0 || role.SortOrder < 0 ||
				validateCapabilities(role.Capabilities) != nil {
				return ErrInvalidConfig
			}
			seenRoles[role.Key] = true
		}
	}
	if err := validatePriorityPolicies(config.Priorities); err != nil {
		return err
	}
	return validatePreferenceVocabulary(config.PreferenceTags)
}

// validatePriorityPolicies rejects a policy whose priorities are malformed or
// point at dimensions the engine does not score. A priority that boosts an
// unknown dimension would silently do nothing, which would look like a
// scoring bug rather than a policy error.
func validatePriorityPolicies(priorities []PriorityPolicy) error {
	seen := make(map[Priority]bool, len(priorities))
	for _, priority := range priorities {
		if !validCode(string(priority.Key)) || seen[priority.Key] {
			return ErrInvalidConfig
		}
		seen[priority.Key] = true
		if len(priority.BoostDimensions) == 0 {
			return ErrInvalidConfig
		}
		seenDimensions := make(map[Dimension]bool, len(priority.BoostDimensions))
		for _, dimension := range priority.BoostDimensions {
			if !validDimension(dimension) || seenDimensions[dimension] {
				return ErrInvalidConfig
			}
			seenDimensions[dimension] = true
		}
		if !validCode(strings.ReplaceAll(priority.ReasonCode, ".", "_")) ||
			strings.TrimSpace(priority.ReasonMessage) == "" ||
			!validDimension(priority.ReasonDimension) ||
			priority.ReasonThreshold < 0 || priority.ReasonThreshold > 100 {
			return ErrInvalidConfig
		}
	}
	return nil
}

func validatePreferenceVocabulary(tags []TrainingPreference) error {
	seen := make(map[TrainingPreference]bool, len(tags))
	for _, tag := range tags {
		if !validCode(string(tag)) || seen[tag] {
			return ErrInvalidConfig
		}
		seen[tag] = true
	}
	return nil
}

func validateInput(input Input, config Config) error {
	if !declaresGoal(config.Goals, input.Goal) || !validExperience(input.Experience) {
		return fmt.Errorf("%w: unsupported goal or experience", ErrInvalidInput)
	}
	if !validMoney(input.Budget) || input.Budget.AmountMinor <= 0 ||
		input.Budget.AmountMinor > maxMoneyMinor {
		return fmt.Errorf("%w: budget must be positive with an uppercase currency", ErrInvalidInput)
	}
	// Space is only meaningful for a vertical with physical products. A
	// non-spatial vertical never reads these fields, so requiring them would
	// force callers to invent a fake room.
	if config.SpatialConstraints {
		space := input.AvailableSpace
		if space.LengthMM <= 0 || space.WidthMM <= 0 || space.HeightMM <= 0 ||
			space.LengthMM > maxDimensionMM || space.WidthMM > maxDimensionMM ||
			space.HeightMM > maxDimensionMM {
			return fmt.Errorf("%w: available space dimensions are invalid", ErrInvalidInput)
		}
		if space.AccessWidthMM != nil && (*space.AccessWidthMM <= 0 || *space.AccessWidthMM > maxDimensionMM) {
			return fmt.Errorf("%w: available access width is invalid", ErrInvalidInput)
		}
	}
	if len(input.FreeText) > maxFreeText || len(input.ExistingEquipment) > maxInputItems ||
		len(input.TrainingPreferences) > 20 || len(input.Priorities) > 20 {
		return fmt.Errorf("%w: input exceeds supported limits", ErrInvalidInput)
	}
	if err := validatePreferences(input.TrainingPreferences, config.PreferenceTags); err != nil {
		return err
	}
	if err := validatePriorities(input.Priorities, config.Priorities); err != nil {
		return err
	}
	for _, equipment := range input.ExistingEquipment {
		if strings.TrimSpace(equipment.Name) == "" ||
			(equipment.CategorySlug != "" && !validSlug(equipment.CategorySlug)) {
			return fmt.Errorf("%w: existing equipment is invalid", ErrInvalidInput)
		}
		if err := validateCapabilities(equipment.Capabilities); err != nil {
			return fmt.Errorf("%w: existing equipment capability: %v", ErrInvalidInput, err)
		}
	}
	return nil
}

func validateCandidates(candidates []CandidateSnapshot, config Config) error {
	if len(candidates) > maxInputItems {
		return fmt.Errorf("%w: too many candidates", ErrInvalidCandidate)
	}
	seen := make(map[catalog.ProductID]struct{}, len(candidates))
	for _, candidate := range candidates {
		if candidate.ProductID == "" || strings.TrimSpace(candidate.Name) == "" ||
			!validSlug(candidate.CategorySlug) || strings.TrimSpace(candidate.PolicyVersion) == "" {
			return fmt.Errorf("%w: identity fields are required", ErrInvalidCandidate)
		}
		if _, exists := seen[candidate.ProductID]; exists {
			return fmt.Errorf("%w: duplicate product %q", ErrInvalidCandidate, candidate.ProductID)
		}
		seen[candidate.ProductID] = struct{}{}
		// Zero is a real price. A free tier is a genuine product in a software
		// vertical -- Zoho Invoice and Wave charge nothing at all -- and
		// rejecting it here rejected the whole request, because one bad
		// candidate fails the batch. The rule came from a physical vertical,
		// where a free treadmill meant a data error. validMoney still refuses
		// a negative amount, and the upper bound still stands.
		if !validMoney(candidate.Price) || candidate.Price.AmountMinor > maxMoneyMinor {
			return fmt.Errorf("%w: product %q price", ErrInvalidCandidate, candidate.ProductID)
		}
		if config.SpatialConstraints && !validDimensions(candidate.Dimensions) {
			return fmt.Errorf("%w: product %q dimensions", ErrInvalidCandidate, candidate.ProductID)
		}
		if !validScores(candidate.Scores) {
			return fmt.Errorf("%w: product %q scores", ErrInvalidCandidate, candidate.ProductID)
		}
		if err := validateSpaceProfile(candidate.Space, config); err != nil {
			return fmt.Errorf("%w: product %q space profile: %v", ErrInvalidCandidate, candidate.ProductID, err)
		}
		if len(candidate.GoalSupport) == 0 {
			return fmt.Errorf("%w: product %q has no supported goal", ErrInvalidCandidate, candidate.ProductID)
		}
		seenGoals := make(map[planning.Goal]bool, len(candidate.GoalSupport))
		for _, support := range candidate.GoalSupport {
			if !declaresGoal(config.Goals, support.Goal) || support.Score < 0 || support.Score > 100 || seenGoals[support.Goal] {
				return fmt.Errorf("%w: product %q goal support", ErrInvalidCandidate, candidate.ProductID)
			}
			seenGoals[support.Goal] = true
		}
		if err := validatePreferences(candidate.PreferenceTags, config.PreferenceTags); err != nil {
			return fmt.Errorf("%w: product %q preference tags", ErrInvalidCandidate, candidate.ProductID)
		}
		if err := validateCodes(candidate.RedundancyGroups); err != nil {
			return fmt.Errorf("%w: product %q redundancy group", ErrInvalidCandidate, candidate.ProductID)
		}
		for _, capabilities := range [][]Capability{
			candidate.Capabilities, candidate.Requires,
			candidate.CompatibleWith, candidate.IncompatibleWith,
		} {
			if err := validateCapabilities(capabilities); err != nil {
				return fmt.Errorf("%w: product %q capability: %v", ErrInvalidCandidate, candidate.ProductID, err)
			}
		}
	}
	return nil
}

// validatePreferences accepts only tags the active policy declares. The
// vocabulary is per-vertical data, but membership is still mandatory: an
// undeclared tag is rejected rather than silently scoring as no match.
func validatePreferences(preferences []TrainingPreference, declared []TrainingPreference) error {
	seen := make(map[TrainingPreference]bool, len(preferences))
	for _, preference := range preferences {
		known := false
		for _, candidate := range declared {
			if candidate == preference {
				known = true
				break
			}
		}
		if !known || seen[preference] {
			return fmt.Errorf("%w: unsupported or duplicate training preference", ErrInvalidInput)
		}
		seen[preference] = true
	}
	return nil
}

func validatePriorities(priorities []Priority, declared []PriorityPolicy) error {
	seen := make(map[Priority]bool, len(priorities))
	for _, priority := range priorities {
		known := false
		for _, candidate := range declared {
			if candidate.Key == priority {
				known = true
				break
			}
		}
		if !known || seen[priority] {
			return fmt.Errorf("%w: unsupported or duplicate priority", ErrInvalidInput)
		}
		seen[priority] = true
	}
	return nil
}

func validateCapabilities(capabilities []Capability) error {
	seen := make(map[Capability]bool, len(capabilities))
	for _, capability := range capabilities {
		if !validCode(string(capability)) || seen[capability] {
			return fmt.Errorf("invalid or duplicate capability %q", capability)
		}
		seen[capability] = true
	}
	return nil
}

func validateCodes(codes []string) error {
	seen := make(map[string]bool, len(codes))
	for _, code := range codes {
		if !validCode(code) || seen[code] {
			return fmt.Errorf("invalid or duplicate code %q", code)
		}
		seen[code] = true
	}
	return nil
}

func validateSpaceProfile(profile SpaceProfile, config Config) error {
	// A non-spatial vertical carries no footprint at all; only the overlap
	// group stays meaningful because the optimizer still uses it to stop two
	// mutually exclusive items being selected together.
	if !config.SpatialConstraints {
		if profile.OverlapGroup != "" && !validCode(profile.OverlapGroup) {
			return ErrInvalidCandidate
		}
		return nil
	}
	if !validEnvelope(profile.Footprint) ||
		(profile.StorageFootprint != nil && !validEnvelope(*profile.StorageFootprint)) ||
		(profile.OperatingClearance != nil && !validClearance(*profile.OperatingClearance)) ||
		(profile.SafetyClearance != nil && !validClearance(*profile.SafetyClearance)) ||
		(profile.MinimumRoomHeightMM != nil && (*profile.MinimumRoomHeightMM <= 0 || *profile.MinimumRoomHeightMM > maxDimensionMM)) ||
		(profile.MinimumAccessWidthMM != nil && (*profile.MinimumAccessWidthMM <= 0 || *profile.MinimumAccessWidthMM > maxDimensionMM)) ||
		(profile.OverlapGroup != "" && !validCode(profile.OverlapGroup)) {
		return ErrInvalidCandidate
	}
	return nil
}

func validEnvelope(value SpatialEnvelope) bool {
	return value.LengthMM > 0 && value.WidthMM > 0 && value.HeightMM > 0 &&
		value.LengthMM <= maxDimensionMM && value.WidthMM <= maxDimensionMM && value.HeightMM <= maxDimensionMM
}

func validClearance(value Clearance) bool {
	for _, measurement := range []int64{value.FrontMM, value.BackMM, value.LeftMM, value.RightMM, value.TopMM} {
		if measurement < 0 || measurement > maxDimensionMM {
			return false
		}
	}
	return true
}

func validCode(value string) bool {
	return normalizedCode.MatchString(value)
}

func validSlug(value string) bool {
	return normalizedSlug.MatchString(value)
}

// declaresGoal reports whether the active policy defines this goal. Goals are
// per-vertical policy data, so membership replaces what used to be a fixed
// fitness enumeration.
func declaresGoal(goals []GoalPolicy, goal planning.Goal) bool {
	for _, declared := range goals {
		if declared.Goal == goal {
			return true
		}
	}
	return false
}

func validExperience(experience planning.ExperienceLevel) bool {
	switch experience {
	case planning.ExperienceBeginner, planning.ExperienceIntermediate, planning.ExperienceAdvanced:
		return true
	default:
		return false
	}
}

func normalizeCandidates(candidates []CandidateSnapshot) []CandidateSnapshot {
	result := append([]CandidateSnapshot(nil), candidates...)
	for index := range result {
		result[index].Capabilities = mergedCapabilities(result[index].Capabilities)
		result[index].Requires = mergedCapabilities(result[index].Requires)
		result[index].CompatibleWith = mergedCapabilities(result[index].CompatibleWith)
		result[index].IncompatibleWith = mergedCapabilities(result[index].IncompatibleWith)
		sort.Slice(result[index].GoalSupport, func(left, right int) bool {
			return result[index].GoalSupport[left].Goal < result[index].GoalSupport[right].Goal
		})
		sort.Slice(result[index].PreferenceTags, func(left, right int) bool {
			return result[index].PreferenceTags[left] < result[index].PreferenceTags[right]
		})
		sort.Strings(result[index].RedundancyGroups)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].ProductID < result[right].ProductID
	})
	return result
}

func validMoney(money catalog.Money) bool {
	return money.AmountMinor >= 0 && len(money.Currency) == 3 &&
		money.Currency == strings.ToUpper(money.Currency)
}

func validDimensions(dimensions catalog.Dimensions) bool {
	return dimensions.LengthMM > 0 && dimensions.WidthMM > 0 && dimensions.HeightMM > 0 &&
		dimensions.LengthMM <= maxDimensionMM && dimensions.WidthMM <= maxDimensionMM &&
		dimensions.HeightMM <= maxDimensionMM
}

func validScores(scores catalog.Scores) bool {
	values := []int16{
		scores.Quality, scores.Value, scores.Durability, scores.Beginner,
		scores.Advanced, scores.Apartment, scores.Noise, scores.Portability,
	}
	for _, score := range values {
		if score < 0 || score > 100 {
			return false
		}
	}
	return true
}

func normalizedExisting(input Input) []ExistingEquipment {
	result := append([]ExistingEquipment(nil), input.ExistingEquipment...)
	for index := range result {
		result[index].Capabilities = mergedCapabilities(result[index].Capabilities)
		sort.Strings(result[index].RedundancyGroups)
	}
	sort.Slice(result, func(left, right int) bool {
		if result[left].Name != result[right].Name {
			return result[left].Name < result[right].Name
		}
		return result[left].CategorySlug < result[right].CategorySlug
	})
	return result
}

func mergedCapabilities(groups ...[]Capability) []Capability {
	unique := make(map[Capability]struct{})
	for _, group := range groups {
		for _, capability := range group {
			unique[capability] = struct{}{}
		}
	}
	result := make([]Capability, 0, len(unique))
	for capability := range unique {
		result = append(result, capability)
	}
	sort.Slice(result, func(left, right int) bool { return result[left] < result[right] })
	return result
}

func mergedStrings(groups ...[]string) []string {
	unique := make(map[string]struct{})
	for _, group := range groups {
		for _, value := range group {
			unique[value] = struct{}{}
		}
	}
	result := make([]string, 0, len(unique))
	for value := range unique {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
