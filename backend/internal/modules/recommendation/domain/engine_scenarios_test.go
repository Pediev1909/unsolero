package domain

import (
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

func TestBeginnerRecommendationPrefersBeginnerAppropriateProduct(t *testing.T) {
	beginnerScores := balancedScores()
	beginnerScores.Beginner = 98
	beginnerScores.Advanced = 45
	advancedScores := balancedScores()
	advancedScores.Beginner = 48
	advancedScores.Advanced = 98
	candidates := []CandidateSnapshot{
		testCandidate("advanced-pair", "adjustable-dumbbells", 20000, advancedScores),
		testCandidate("beginner-pair", "adjustable-dumbbells", 20000, beginnerScores),
	}

	result, err := testEngine(t).Recommend(
		testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 30000),
		candidates,
	)
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if len(result.Selected) != 1 || result.Selected[0].Product.Candidate.ProductID != "beginner-pair" {
		t.Fatalf("selected = %#v, want beginner-pair", selectedIDs(result))
	}
	if !hasReason(result.Selected[0].Product, "experience.strong_match") {
		t.Fatal("expected an experience explanation")
	}
}

func TestAdvancedRecommendationPrefersAdvancedProduct(t *testing.T) {
	beginnerScores := balancedScores()
	beginnerScores.Beginner = 98
	beginnerScores.Advanced = 45
	advancedScores := balancedScores()
	advancedScores.Beginner = 48
	advancedScores.Advanced = 98
	candidates := []CandidateSnapshot{
		testCandidate("beginner-pair", "adjustable-dumbbells", 20000, beginnerScores),
		testCandidate("advanced-pair", "adjustable-dumbbells", 20000, advancedScores),
	}

	result, err := testEngine(t).Recommend(
		testInput(planning.GoalBuildMuscle, planning.ExperienceAdvanced, 30000),
		candidates,
	)
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if len(result.Selected) != 1 || result.Selected[0].Product.Candidate.ProductID != "advanced-pair" {
		t.Fatalf("selected = %#v, want advanced-pair", selectedIDs(result))
	}
}

func TestSmallApartmentRemovesOversizedEquipment(t *testing.T) {
	input := testInput(planning.GoalStrength, planning.ExperienceBeginner, 120000)
	input.AvailableSpace = AvailableSpace{
		LengthMM: 1000, WidthMM: 1000, HeightMM: 2300, ApartmentLiving: true,
	}
	compact := testCandidate("compact-dumbbells", "adjustable-dumbbells", 25000, balancedScores())
	compact.Scores.Apartment = 96
	largeRack := testCandidate("large-rack", "power-racks", 70000, balancedScores())
	largeRack.Dimensions = catalog.Dimensions{LengthMM: 1600, WidthMM: 1500, HeightMM: 2200}
	largeRack.Space.Footprint = SpatialEnvelope{LengthMM: 1600, WidthMM: 1500, HeightMM: 2200}

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{largeRack, compact})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["compact-dumbbells"] {
		t.Fatalf("selected = %#v, want compact-dumbbells", selectedIDs(result))
	}
	rejection, found := rejectionFor(result, "large-rack")
	if !found || rejection.Code != "space.does_not_fit" {
		t.Fatalf("large-rack rejection = %#v", rejection)
	}
}

func TestLargeGarageBuildsCompatibleBarbellSetup(t *testing.T) {
	input := testInput(planning.GoalStrength, planning.ExperienceAdvanced, 200000)
	input.AvailableSpace = AvailableSpace{LengthMM: 6000, WidthMM: 6000, HeightMM: 3200}
	high := catalog.Scores{
		Quality: 94, Value: 86, Durability: 94, Beginner: 70,
		Advanced: 98, Apartment: 45, Noise: 70, Portability: 30,
	}
	candidates := []CandidateSnapshot{
		testCandidate("barbell", "barbells", 30000, high),
		testCandidate("plates", "weight-plates", 25000, high),
		testCandidate("rack", "power-racks", 70000, high),
		testCandidate("bench", "benches", 20000, high),
		testCandidate("dumbbells", "adjustable-dumbbells", 40000, balancedScores()),
	}

	result, err := testEngine(t).Recommend(input, candidates)
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	selected := selectedIDs(result)
	for _, productID := range []catalog.ProductID{"barbell", "plates", "rack", "bench"} {
		if !selected[productID] {
			t.Fatalf("selected = %#v, missing %s", selected, productID)
		}
	}
	if result.TotalCost.AmountMinor != 145000 {
		t.Fatalf("total = %d, want 145000", result.TotalCost.AmountMinor)
	}
}

func TestLowBudgetSelectsAffordableCompleteOption(t *testing.T) {
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 10000)
	input.Priorities = []Priority{PriorityBudget}
	bandScores := balancedScores()
	bandScores.Value = 95
	bandScores.Beginner = 96
	band := testCandidate("bands", "resistance-bands", 4000, bandScores)
	dumbbells := testCandidate("dumbbells", "adjustable-dumbbells", 22000, balancedScores())

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{dumbbells, band})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["bands"] || result.TotalCost.AmountMinor != 4000 {
		t.Fatalf("unexpected low-budget setup: %#v, total %d", selectedIDs(result), result.TotalCost.AmountMinor)
	}
	rejection, found := rejectionFor(result, "dumbbells")
	if !found || rejection.Code != "budget.exceeded" {
		t.Fatalf("dumbbells rejection = %#v", rejection)
	}
}

func TestPremiumBudgetPrefersHigherQualityProductAndKeepsCheaperAlternative(t *testing.T) {
	cheapScores := balancedScores()
	cheapScores.Quality = 72
	cheapScores.Value = 92
	premiumScores := balancedScores()
	premiumScores.Quality = 97
	premiumScores.Durability = 96
	premiumScores.Value = 80
	cheap := testCandidate("cheap-dumbbells", "adjustable-dumbbells", 15000, cheapScores)
	premium := testCandidate("premium-dumbbells", "adjustable-dumbbells", 55000, premiumScores)
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceIntermediate, 100000)
	input.Priorities = []Priority{PriorityQuality, PriorityDurability}

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{cheap, premium})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["premium-dumbbells"] {
		t.Fatalf("selected = %#v, want premium-dumbbells", selectedIDs(result))
	}
	if len(result.Alternatives) != 1 || result.Alternatives[0].Type != AlternativeCheaper ||
		result.Alternatives[0].Product.Candidate.ProductID != "cheap-dumbbells" {
		t.Fatalf("alternatives = %#v", result.Alternatives)
	}
}

func TestSelectedProductCanHaveCheaperAndPremiumAlternatives(t *testing.T) {
	cheapScores := balancedScores()
	cheapScores.Quality = 70
	cheapScores.Value = 88
	middleScores := balancedScores()
	middleScores.Quality = 86
	middleScores.Beginner = 95
	middleScores.Value = 88
	premiumScores := balancedScores()
	premiumScores.Quality = 96
	premiumScores.Beginner = 30
	premiumScores.Value = 45
	cheap := testCandidate("cheap", "adjustable-dumbbells", 12000, cheapScores)
	middle := testCandidate("middle", "adjustable-dumbbells", 25000, middleScores)
	premium := testCandidate("premium", "adjustable-dumbbells", 50000, premiumScores)
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 80000)

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{premium, cheap, middle})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["middle"] {
		t.Fatalf("selected = %#v, want middle", selectedIDs(result))
	}
	types := make(map[AlternativeType]catalog.ProductID)
	for _, alternative := range result.Alternatives {
		types[alternative.Type] = alternative.Product.Candidate.ProductID
	}
	if types[AlternativeCheaper] != "cheap" || types[AlternativePremium] != "premium" {
		t.Fatalf("alternatives = %#v", types)
	}
}

func TestTrainingPreferenceBreaksOtherwiseComparableProducts(t *testing.T) {
	dumbbells := testCandidate("dumbbells", "adjustable-dumbbells", 20000, balancedScores())
	kettlebell := testCandidate("kettlebell", "kettlebells", 20000, balancedScores())
	input := testInput(planning.GoalGeneralFitness, planning.ExperienceIntermediate, 30000)
	input.TrainingPreferences = []TrainingPreference{PreferenceKettlebell}

	result, err := testEngine(t).Recommend(input, []CandidateSnapshot{dumbbells, kettlebell})
	if err != nil {
		t.Fatalf("Recommend(): %v", err)
	}
	if !selectedIDs(result)["kettlebell"] {
		t.Fatalf("selected = %#v, want kettlebell", selectedIDs(result))
	}
}
