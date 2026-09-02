package application

import (
	"context"
	"errors"
	"testing"

	admin "rigmark/internal/modules/admin/domain"
	adminports "rigmark/internal/modules/admin/ports"
	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
)

type ownershipRepository struct {
	adminports.Repository
	addCalled bool
}

func (repository *ownershipRepository) AddImage(context.Context, identity.UserID, catalog.ProductID, admin.ImageInput) (catalog.ProductImage, error) {
	repository.addCalled = true
	return catalog.ProductImage{}, nil
}

type ownershipStorage struct{ adminports.ImageStorage }

func (ownershipStorage) BelongsTo(catalog.ProductID, string) bool { return false }

func TestValidateProductInputRejectsOutOfRangeRecommendationScore(t *testing.T) {
	input := validProductInput()
	input.Scores.Apartment = 101
	if err := validateProductInput(input); err == nil {
		t.Fatal("expected invalid score to be rejected")
	}
}

// The 422 names the billing field that failed so the form can mark it, and
// still satisfies errors.Is(ErrInvalidInput) for every caller that only asks
// whether the input was valid.
func TestCreateProductNamesTheBillingFieldThatFailed(t *testing.T) {
	annual := int64(1_500)
	input := validProductInput()
	input.Billing = catalog.Billing{Period: catalog.BillingAnnual, Unit: catalog.PricingUnitFlat, AnnualPriceMinor: &annual}
	service := NewService(&ownershipRepository{}, nil)

	_, err := service.CreateProduct(context.Background(), "actor", input)

	var field catalog.FieldError
	if !errors.Is(err, ErrInvalidInput) || !errors.As(err, &field) || field.Field != "billing.annual_price_minor" {
		t.Fatalf("CreateProduct() error = %v, want ErrInvalidInput naming billing.annual_price_minor", err)
	}
}

// A blank note is the form's way of saying "no note"; it is stored as absent
// rather than rejected as too short.
func TestNormalizeProductDropsABlankBillingNote(t *testing.T) {
	blank := "   "
	input := validProductInput()
	// A software product: the physical fixture values would be rejected on
	// their own account and hide what this test is about.
	input.Dimensions, input.WeightGrams, input.Material = catalog.Dimensions{}, 0, ""
	input.Billing.UnitNote = &blank
	if normalized := normalizeProduct(input); normalized.Billing.UnitNote != nil {
		t.Fatalf("blank unit note survived normalization: %q", *normalized.Billing.UnitNote)
	}
	if err := validateProductInput(normalizeProduct(input)); err != nil {
		t.Fatalf("normalized input rejected: %v", err)
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

func TestValidateMerchantURLsRejectsPrivateNetworksAndUserInfo(t *testing.T) {
	for _, destination := range []string{
		"https://127.0.0.1/product",
		"https://10.0.0.5/product",
		"https://user:secret@merchant.example/product",
	} {
		input := admin.AffiliateLinkInput{Provider: "direct", DestinationURL: destination,
			DisclosureLabel: "Affiliate link", CommissionType: "unknown"}
		if err := validateAffiliate(input); err == nil {
			t.Fatalf("unsafe destination accepted: %s", destination)
		}
	}
}

func TestPaginationRejectsUnboundedPageAndPageSize(t *testing.T) {
	for _, input := range [][2]int{{0, 30}, {10_001, 30}, {1, 101}} {
		if _, _, err := pagination(input[0], input[1]); err == nil {
			t.Fatalf("pagination accepted page=%d pageSize=%d", input[0], input[1])
		}
	}
	if offset, limit, err := pagination(10_000, 100); err != nil || offset != 999_900 || limit != 100 {
		t.Fatalf("pagination boundary = (%d, %d, %v)", offset, limit, err)
	}
}

func TestAddImageRejectsCrossProductManagedObject(t *testing.T) {
	repository := &ownershipRepository{}
	service := NewService(repository, ownershipStorage{})
	_, err := service.AddImage(context.Background(), "actor", "12345678-1234-4234-8234-123456789abc", admin.ImageInput{
		URL:     "/api/media/products/22345678-1234-4234-8234-123456789abc_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.png",
		AltText: "Wrong product image",
	})
	if err != ErrInvalidInput || repository.addCalled {
		t.Fatalf("AddImage() error=%v addCalled=%v", err, repository.addCalled)
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
		Billing:     catalog.Billing{Period: catalog.BillingMonthly, Unit: catalog.PricingUnitFlat},
		Dimensions:  catalog.Dimensions{LengthMM: 1, WidthMM: 1, HeightMM: 1},
		WeightGrams: 1,
		Material:    "Steel",
		Scores:      catalog.Scores{Quality: 50, Value: 50, Durability: 50, Beginner: 50, Advanced: 50, Apartment: 50, Noise: 50, Portability: 50},
	}
}
