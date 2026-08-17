package domain

import (
	"reflect"
	"strings"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	planning "rigmark/internal/modules/planning/domain"
)

func TestCompetingProductsUseStableTieBreakAndIgnoreCandidateOrder(t *testing.T) {
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceIntermediate, 50000)
	alpha := testCandidate("alpha", "adjustable-dumbbells", 20000, balancedScores())
	bravo := testCandidate("bravo", "adjustable-dumbbells", 20000, balancedScores())
	engine := testEngine(t)

	first, err := engine.Recommend(input, []CandidateSnapshot{bravo, alpha})
	if err != nil {
		t.Fatalf("first Recommend(): %v", err)
	}
	second, err := engine.Recommend(input, []CandidateSnapshot{alpha, bravo})
	if err != nil {
		t.Fatalf("second Recommend(): %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("candidate order changed result\nfirst: %#v\nsecond: %#v", first, second)
	}
	if len(first.Selected) != 1 || first.Selected[0].Product.Candidate.ProductID != "alpha" {
		t.Fatalf("selected = %#v, want alpha", selectedIDs(first))
	}
	if first.InputFingerprint == "" || first.InputFingerprint != second.InputFingerprint {
		t.Fatalf("fingerprints = %q and %q", first.InputFingerprint, second.InputFingerprint)
	}
}

func TestCandidateFromProductCopiesOnlyObjectiveCatalogFacts(t *testing.T) {
	product := catalog.Product{
		ID: "product-1", Name: "Demo Product", CategorySlug: "kettlebells",
		Price:      catalog.Money{AmountMinor: 9900, Currency: "USD"},
		Dimensions: catalog.Dimensions{LengthMM: 250, WidthMM: 200, HeightMM: 300},
		Scores:     balancedScores(),
	}
	snapshot := CandidateFromProduct(product)
	if snapshot.ProductID != product.ID || snapshot.Price != product.Price ||
		snapshot.Dimensions != product.Dimensions || snapshot.Scores != product.Scores {
		t.Fatalf("snapshot = %#v", snapshot)
	}
}

func TestConfigurableWeightsChangeObjectiveRankingPredictably(t *testing.T) {
	quality := testCandidate("quality", "adjustable-dumbbells", 20000, balancedScores())
	quality.Scores.Quality = 98
	quality.Scores.Value = 50
	value := testCandidate("value", "adjustable-dumbbells", 20000, balancedScores())
	value.Scores.Quality = 50
	value.Scores.Value = 98
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceIntermediate, 50000)

	qualityConfig := testPolicyConfig("test-policy-v2")
	qualityConfig.Weights = Weights{Quality: 1}
	qualityEngine, err := NewDeterministicRecommendationEngine(qualityConfig)
	if err != nil {
		t.Fatalf("quality engine: %v", err)
	}
	qualityResult, err := qualityEngine.Recommend(input, []CandidateSnapshot{value, quality})
	if err != nil {
		t.Fatalf("quality Recommend(): %v", err)
	}
	if !selectedIDs(qualityResult)["quality"] {
		t.Fatalf("quality-weighted selection = %#v", selectedIDs(qualityResult))
	}

	valueConfig := testPolicyConfig("test-policy-v2")
	valueConfig.Weights = Weights{Value: 1}
	valueEngine, err := NewDeterministicRecommendationEngine(valueConfig)
	if err != nil {
		t.Fatalf("value engine: %v", err)
	}
	valueResult, err := valueEngine.Recommend(input, []CandidateSnapshot{quality, value})
	if err != nil {
		t.Fatalf("value Recommend(): %v", err)
	}
	if !selectedIDs(valueResult)["value"] {
		t.Fatalf("value-weighted selection = %#v", selectedIDs(valueResult))
	}
}

func TestRecommendationInputsCannotCarryCommercialInfluence(t *testing.T) {
	for _, structure := range []reflect.Type{
		reflect.TypeOf(Input{}),
		reflect.TypeOf(CandidateSnapshot{}),
		reflect.TypeOf(Config{}),
	} {
		for index := 0; index < structure.NumField(); index++ {
			name := strings.ToLower(structure.Field(index).Name)
			for _, forbidden := range []string{"affiliate", "commission", "merchant", "sponsor", "payout"} {
				if strings.Contains(name, forbidden) {
					t.Fatalf("%s exposes forbidden commercial field %q", structure.Name(), name)
				}
			}
		}
	}
}

func TestFreeTextIsAcceptedButDoesNotInventScoringSignals(t *testing.T) {
	candidate := testCandidate("dumbbells", "adjustable-dumbbells", 20000, balancedScores())
	input := testInput(planning.GoalBuildMuscle, planning.ExperienceBeginner, 50000)
	engine := testEngine(t)

	withoutText, err := engine.Recommend(input, []CandidateSnapshot{candidate})
	if err != nil {
		t.Fatalf("Recommend without text: %v", err)
	}
	input.FreeText = "I like equipment that feels magical and futuristic."
	withText, err := engine.Recommend(input, []CandidateSnapshot{candidate})
	if err != nil {
		t.Fatalf("Recommend with text: %v", err)
	}
	if !reflect.DeepEqual(withoutText.Selected, withText.Selected) ||
		!reflect.DeepEqual(withoutText.Ranked, withText.Ranked) {
		t.Fatal("unstructured free text changed deterministic scoring")
	}
	if withoutText.InputFingerprint == withText.InputFingerprint {
		t.Fatal("free text must remain part of the auditable input fingerprint")
	}
}
