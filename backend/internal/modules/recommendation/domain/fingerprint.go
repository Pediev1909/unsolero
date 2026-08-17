package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

type fingerprintPayload struct {
	Input      Input
	Candidates []CandidateSnapshot
	Existing   []ExistingEquipment
	Config     Config
	Engine     string
}

func resultFingerprint(
	input Input,
	candidates []CandidateSnapshot,
	existing []ExistingEquipment,
	config Config,
) (string, error) {
	canonicalInput := input
	canonicalInput.ExistingEquipment = nil
	canonicalInput.TrainingPreferences = append([]TrainingPreference(nil), input.TrainingPreferences...)
	canonicalInput.Priorities = append([]Priority(nil), input.Priorities...)
	sort.Slice(canonicalInput.TrainingPreferences, func(left, right int) bool {
		return canonicalInput.TrainingPreferences[left] < canonicalInput.TrainingPreferences[right]
	})
	sort.Slice(canonicalInput.Priorities, func(left, right int) bool {
		return canonicalInput.Priorities[left] < canonicalInput.Priorities[right]
	})
	encoded, err := json.Marshal(fingerprintPayload{
		Input: canonicalInput, Candidates: candidates, Existing: existing,
		Config: config, Engine: EngineVersion,
	})
	if err != nil {
		return "", fmt.Errorf("fingerprint recommendation: %w", err)
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}
