package domain

import (
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
)

func TestRevisionRequiresProvenanceAndRationales(t *testing.T) {
	input := validRevisionInput()
	if err := input.Validate(); err != nil {
		t.Fatalf("valid revision rejected: %v", err)
	}
	input.FactLinks = input.FactLinks[1:]
	if err := input.Validate(); err == nil {
		t.Fatal("revision without complete fact provenance was accepted")
	}
	input = validRevisionInput()
	input.Rationales = input.Rationales[:7]
	if err := input.Validate(); err == nil {
		t.Fatal("revision without every score rationale was accepted")
	}
}

func TestDemoEvidenceMustBeExplicitlyFictional(t *testing.T) {
	if err := (SourceInput{Type: SourceDemo, Title: "Demo", Publisher: "Seed"}).Validate(); err == nil {
		t.Fatal("unmarked demo evidence was accepted")
	}
	if err := (SourceInput{Type: SourceIndependent, Title: "Test", Publisher: "Lab", IsFictional: true}).Validate(); err == nil {
		t.Fatal("fictional independent evidence was accepted")
	}
}

func validRevisionInput() RevisionInput {
	product := catalog.Product{ID: "product", CategoryID: "category", BrandID: "brand",
		Name: "Product", Slug: "product", Description: "Description",
		Price:       catalog.Money{AmountMinor: 100, Currency: "USD"},
		Dimensions:  catalog.Dimensions{LengthMM: 1, WidthMM: 1, HeightMM: 1},
		WeightGrams: 1, Material: "Steel", WarrantyMonths: 0,
		Scores: catalog.Scores{Quality: 80, Value: 80, Durability: 80, Beginner: 80,
			Advanced: 80, Apartment: 80, Noise: 80, Portability: 80},
		Status: catalog.ProductStatusDraft}
	input := RevisionInput{Product: product, Scores: product.Scores}
	for _, key := range []string{"category", "brand", "name", "slug", "description", "price", "dimensions", "weight", "material", "warranty"} {
		input.FactLinks = append(input.FactLinks, FactLink{FactKey: key,
			ObservationID: "observation", Classification: ClassificationVerified})
	}
	for _, key := range []string{"quality", "value", "durability", "beginner", "advanced", "apartment", "noise", "portability"} {
		input.Rationales = append(input.Rationales, ScoreRationale{ScoreKey: key,
			ObservationID: "observation", Rationale: "Evidence-backed rationale"})
	}
	return input
}
