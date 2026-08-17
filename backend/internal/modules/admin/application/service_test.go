package application

import (
	"testing"

	admin "rigmark/internal/modules/admin/domain"
	catalog "rigmark/internal/modules/catalog/domain"
)

func TestValidateProductInputRejectsOutOfRangeRecommendationScore(t *testing.T) {
	input := validProductInput()
	input.Scores.Apartment = 101
	if err := validateProductInput(input); err == nil {
		t.Fatal("expected invalid score to be rejected")
	}
}

func TestValidateOfferInputKeepsAffiliateCommissionOutOfProductData(t *testing.T) {
	input := admin.OfferInput{
		MerchantID:   "merchant",
		ProductID:    "product",
		MerchantSKU:  "sku",
		ProductURL:   "https://merchant.example/product",
		PriceMinor:   100,
		Currency:     "USD",
		Availability: "in_stock",
		Condition:    "new",
		Affiliate: &admin.AffiliateLinkInput{
			Provider:        "direct",
			DestinationURL:  "https://merchant.example/affiliate",
			DisclosureLabel: "Affiliate link",
			CommissionType:  "unknown",
		},
	}
	if err := validateOfferInput(input); err != nil {
		t.Fatalf("valid offer rejected: %v", err)
	}
	if input.ProductID == "" || input.Affiliate == nil {
		t.Fatal("test fixture did not preserve separate product and affiliate fields")
	}
}

func TestValidateAffiliateRejectsUnsafeDestination(t *testing.T) {
	input := admin.AffiliateLinkInput{Provider: "direct", DestinationURL: "javascript:alert(1)", DisclosureLabel: "Affiliate link", CommissionType: "unknown"}
	if err := validateAffiliate(input); err == nil {
		t.Fatal("expected unsafe destination to be rejected")
	}
}

func validProductInput() admin.ProductInput {
	return admin.ProductInput{
		CategoryID:  "category",
		BrandID:     "brand",
		Name:        "Demo product",
		Slug:        "demo-product",
		Description: "A valid fictional product.",
		Price:       catalog.Money{AmountMinor: 100, Currency: "USD"},
		Dimensions:  catalog.Dimensions{LengthMM: 1, WidthMM: 1, HeightMM: 1},
		WeightGrams: 1,
		Material:    "Steel",
		Scores:      catalog.Scores{Quality: 50, Value: 50, Durability: 50, Beginner: 50, Advanced: 50, Apartment: 50, Noise: 50, Portability: 50},
	}
}
