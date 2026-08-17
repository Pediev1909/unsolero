package domain

import (
	"errors"
	"testing"

	planning "rigmark/internal/modules/planning/domain"
)

func TestExistingEquipmentPreventsRedundantRecommendation(t *testing.T) {
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceIntermediate, 50000)
	input.ExistingEquipment = []ExistingEquipment{{
		Name: "adjustable dumbbells", CategorySlug: "adjustable-dumbbells",
		Capabilities:     []Capability{CapabilityResistanceTraining, CapabilityStrengthTraining, CapabilityHypertrophy},
		RedundancyGroups: []string{"dumbbell_system"},
	}}
	dumbbells := testCandidate("new-dumbbells", "adjustable-dumbbells", 25000, balancedScores())
	bench := testCandidate("bench", "benches", 15000, balancedScores())

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{dumbbells, bench})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["bench"] || selectedIDs(result)["new-dumbbells"] {
		t.Fatalf("selected = %#v", selectedIDs(result))
	}
	rejection, found := rejectionFor(result, "new-dumbbells")
	if !found || rejection.Code != "existing_equipment.redundant" {
		t.Fatalf("new-dumbbells rejection = %#v", rejection)
	}
}

func TestExistingPullUpBarProducesCompatibilityExplanation(t *testing.T) {
	input := testInput(planning.GoalMobility, planning.ExperienceBeginner, 10000)
	input.ExistingEquipment = []ExistingEquipment{{
		Name: "pull-up bar", CategorySlug: "pull-up-bar",
		Capabilities: []Capability{CapabilityPullUp, CapabilityAnchorPoint},
	}}
	bands := testCandidate("bands", "resistance-bands", 4000, balancedScores())

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{bands})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if len(result.Selected) != 1 || !hasReason(result.Selected[0].Product, "compatibility.existing_equipment") {
		t.Fatalf("selected reasons = %#v", result.Selected)
	}
	messageFound := false
	for _, reason := range result.Selected[0].Product.Reasons {
		if reason.Message == "Compatible with your existing pull-up bar" {
			messageFound = true
		}
	}
	if !messageFound {
		t.Fatal("expected deterministic pull-up bar compatibility message")
	}
}

func TestIncompatibleEquipmentIsRemovedBeforeScoring(t *testing.T) {
	input := testInput(planning.GoalWeightLoss, planning.ExperienceIntermediate, 100000)
	input.ExistingEquipment = []ExistingEquipment{{
		Name: "fragile flooring", Capabilities: []Capability{"fragile_floor"},
	}}
	airBike := testCandidate("air-bike", "cardio-machines", 70000, balancedScores())
	airBike.IncompatibleWith = []Capability{"fragile_floor"}
	bands := testCandidate("bands", "resistance-bands", 4000, balancedScores())

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{airBike, bands})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["bands"] {
		t.Fatalf("selected = %#v, want bands", selectedIDs(result))
	}
	rejection, found := rejectionFor(result, "air-bike")
	if !found || rejection.Code != "compatibility.existing_conflict" {
		t.Fatalf("air-bike rejection = %#v", rejection)
	}
	for _, ranked := range result.Ranked {
		if ranked.Candidate.ProductID == "air-bike" {
			t.Fatal("incompatible product reached the scoring stage")
		}
	}
}

func TestMutuallyIncompatibleProductsAreNotCombined(t *testing.T) {
	input := testInput(planning.GoalGeneralFitness, planning.ExperienceIntermediate, 150000)
	dumbbells := testCandidate("dumbbells", "adjustable-dumbbells", 25000, balancedScores())
	cardio := testCandidate("cardio", "cardio-machines", 60000, balancedScores())
	cardio.IncompatibleWith = []Capability{CapabilityResistanceTraining}

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{dumbbells, cardio})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	selected := selectedIDs(result)
	if selected["dumbbells"] && selected["cardio"] {
		t.Fatalf("mutually incompatible products were combined: %#v", selected)
	}
	rejection, found := rejectionFor(result, "cardio")
	if !found || rejection.Code != "setup.incompatible" {
		t.Fatalf("cardio rejection = %#v", rejection)
	}
}

func TestNoSuitableProductReturnsStructuredOutcome(t *testing.T) {
	input := testInput(planning.GoalStrength, planning.ExperienceAdvanced, 10000)
	oversized := testCandidate("oversized-rack", "power-racks", 9000, balancedScores())
	oversized.Dimensions.LengthMM = 5000
	oversized.Space.Footprint.LengthMM = 5000

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{oversized})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if result.Status != ResultNoSuitableProducts || len(result.Selected) != 0 ||
		result.TotalCost.AmountMinor != 0 {
		t.Fatalf("result = %#v", result)
	}
	if rejection, found := rejectionFor(result, "oversized-rack"); !found || rejection.Code != "space.does_not_fit" {
		t.Fatalf("oversized-rack rejection = %#v", rejection)
	}
}

func TestHardConstraintValidationRejectsInvalidBudget(t *testing.T) {
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 0)
	_, err := testEngine(t).Recommend(input, nil)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("error = %v, want ErrInvalidInput", err)
	}
}

func TestMissingCompatibilityRequirementIsRejected(t *testing.T) {
	input := testInput(planning.GoalStrength, planning.ExperienceAdvanced, 50000)
	barbell := testCandidate("barbell", "barbells", 25000, balancedScores())

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{barbell})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	rejection, found := rejectionFor(result, "barbell")
	if !found || rejection.Code != "compatibility.missing_requirement" {
		t.Fatalf("barbell rejection = %#v", rejection)
	}
}

func TestSetupCannotExceedCombinedFloorArea(t *testing.T) {
	input := testInput(planning.GoalGeneralFitness, planning.ExperienceIntermediate, 100000)
	input.AvailableSpace = AvailableSpace{LengthMM: 1000, WidthMM: 1000, HeightMM: 2400}
	dumbbells := testCandidate("dumbbells", "adjustable-dumbbells", 25000, balancedScores())
	dumbbells.Dimensions.LengthMM, dumbbells.Dimensions.WidthMM = 800, 800
	dumbbells.Space.Footprint.LengthMM, dumbbells.Space.Footprint.WidthMM = 800, 800
	cardio := testCandidate("cardio", "cardio-machines", 50000, balancedScores())
	cardio.Dimensions.LengthMM, cardio.Dimensions.WidthMM = 800, 800
	cardio.Space.Footprint.LengthMM, cardio.Space.Footprint.WidthMM = 800, 800

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{dumbbells, cardio})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if len(result.Selected) != 1 {
		t.Fatalf("selected %d products whose combined footprints do not fit", len(result.Selected))
	}
	rejection, found := rejectionFor(result, "cardio")
	if !found || rejection.Code != "setup.space_limit" {
		t.Fatalf("cardio rejection = %#v", rejection)
	}
}
