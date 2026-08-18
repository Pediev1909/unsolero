package domain

import "testing"

func TestProductValidate(t *testing.T) {
	capacity := int64(100_000)
	product := Product{
		ID:               "product-id",
		CategoryID:       "category-id",
		BrandID:          "brand-id",
		Name:             "Demo Product",
		Slug:             "demo-product",
		Description:      "A fictional demo product.",
		Price:            Money{AmountMinor: 10_000, Currency: "USD"},
		IsPhysical:       true,
		Dimensions:       Dimensions{LengthMM: 100, WidthMM: 100, HeightMM: 100},
		WeightGrams:      10_000,
		MaxCapacityGrams: &capacity,
		Material:         "Steel",
		WarrantyMonths:   12,
		Scores: Scores{
			Quality: 80, Value: 80, Durability: 80, Beginner: 80,
			Advanced: 80, Apartment: 80, Noise: 80, Portability: 80,
		},
		Status: ProductStatusPublished,
	}

	if err := product.Validate(); err != nil {
		t.Fatalf("Validate() returned an unexpected error: %v", err)
	}
}

func TestProductValidateRejectsOutOfRangeScore(t *testing.T) {
	product := Product{
		ID:          "product-id",
		CategoryID:  "category-id",
		BrandID:     "brand-id",
		Name:        "Demo Product",
		Slug:        "demo-product",
		Description: "A fictional demo product.",
		Price:       Money{AmountMinor: 10_000, Currency: "USD"},
		Dimensions:  Dimensions{LengthMM: 100, WidthMM: 100, HeightMM: 100},
		WeightGrams: 10_000,
		Material:    "Steel",
		Scores:      Scores{Quality: 101},
		Status:      ProductStatusPublished,
	}

	if err := product.Validate(); err == nil {
		t.Fatal("Validate() expected an error for an out-of-range score")
	}
}

func TestAttributeValidateEnforcesTypedValue(t *testing.T) {
	number := 5.0
	text := "unexpected"
	attribute := Attribute{
		Key:          "adjustment_steps",
		Type:         AttributeTypeNumber,
		NumericValue: &number,
		TextValue:    &text,
	}

	if err := attribute.Validate(); err == nil {
		t.Fatal("Validate() expected an error when multiple typed values are set")
	}
}

func TestProductInsightsAreDerivedFromScores(t *testing.T) {
	product := Product{Scores: Scores{
		Quality: 90, Value: 84, Durability: 88, Beginner: 92,
		Advanced: 55, Apartment: 95, Noise: 86, Portability: 40,
	}}

	if len(product.Strengths()) != 5 {
		t.Fatalf("expected five strengths, got %#v", product.Strengths())
	}
	if len(product.Considerations()) != 2 {
		t.Fatalf("expected two considerations, got %#v", product.Considerations())
	}
	if len(product.UseCases()) != 3 {
		t.Fatalf("expected three use cases, got %#v", product.UseCases())
	}
}

func TestNonPhysicalProductValidatesWithoutPhysicalAttributes(t *testing.T) {
	product := Product{
		ID: "product-id", CategoryID: "category-id", BrandID: "brand-id",
		Name: "Demo Subscription", Slug: "demo-subscription",
		Description: "A fictional demo subscription.",
		Price:       Money{AmountMinor: 2_900, Currency: "USD"},
		Scores: Scores{
			Quality: 80, Value: 80, Durability: 80, Beginner: 80,
			Advanced: 80, Apartment: 0, Noise: 0, Portability: 80,
		},
		Status: ProductStatusPublished,
	}

	if err := product.Validate(); err != nil {
		t.Fatalf("Validate() rejected a valid non-physical product: %v", err)
	}
}

func TestNonPhysicalProductRejectsPhysicalAttributes(t *testing.T) {
	product := Product{
		ID: "product-id", CategoryID: "category-id", BrandID: "brand-id",
		Name: "Demo Subscription", Slug: "demo-subscription",
		Description: "A fictional demo subscription.",
		Price:       Money{AmountMinor: 2_900, Currency: "USD"},
		Dimensions:  Dimensions{LengthMM: 100, WidthMM: 100, HeightMM: 100},
		Scores: Scores{
			Quality: 80, Value: 80, Durability: 80, Beginner: 80,
			Advanced: 80, Apartment: 0, Noise: 0, Portability: 80,
		},
		Status: ProductStatusPublished,
	}

	if err := product.Validate(); err == nil {
		t.Fatal("Validate() accepted a non-physical product carrying a footprint")
	}
}
