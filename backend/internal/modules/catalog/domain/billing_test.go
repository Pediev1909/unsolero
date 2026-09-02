package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestBillingValidate(t *testing.T) {
	note := "at 1,000 contacts"
	longNote := strings.Repeat("x", 121)
	blank := "   "
	annual := int64(1_500)
	negative := int64(-1)
	cases := []struct {
		name      string
		billing   Billing
		price     int64
		wantField string
	}{
		{"monthly flat", Billing{Period: BillingMonthly, Unit: PricingUnitFlat}, 2_000, ""},
		{"monthly with annual option", Billing{Period: BillingMonthly, Unit: PricingUnitPerUser, AnnualPriceMinor: &annual}, 2_000, ""},
		{"annual only", Billing{Period: BillingAnnual, Unit: PricingUnitFlat}, 2_000, ""},
		{"free at zero", Billing{Period: BillingFree, Unit: PricingUnitFlat}, 0, ""},
		{"usage with note", Billing{Period: BillingUsage, Unit: PricingUnitPerTransaction, UnitNote: &note}, 0, ""},
		{"contact tier note", Billing{Period: BillingMonthly, Unit: PricingUnitPerContacts, UnitNote: &note}, 2_000, ""},
		{"unknown period", Billing{Period: "quarterly", Unit: PricingUnitFlat}, 2_000, "billing.period"},
		{"empty period", Billing{Unit: PricingUnitFlat}, 2_000, "billing.period"},
		{"unknown unit", Billing{Period: BillingMonthly, Unit: "per_seat"}, 2_000, "billing.unit"},
		{"note too long", Billing{Period: BillingMonthly, Unit: PricingUnitFlat, UnitNote: &longNote}, 2_000, "billing.unit_note"},
		{"blank note", Billing{Period: BillingMonthly, Unit: PricingUnitFlat, UnitNote: &blank}, 2_000, "billing.unit_note"},
		{"annual figure beside annual-only billing", Billing{Period: BillingAnnual, Unit: PricingUnitFlat, AnnualPriceMinor: &annual}, 2_000, "billing.annual_price_minor"},
		{"annual figure beside free plan", Billing{Period: BillingFree, Unit: PricingUnitFlat, AnnualPriceMinor: &annual}, 0, "billing.annual_price_minor"},
		{"negative annual figure", Billing{Period: BillingMonthly, Unit: PricingUnitFlat, AnnualPriceMinor: &negative}, 2_000, "billing.annual_price_minor"},
		{"free plan with a price", Billing{Period: BillingFree, Unit: PricingUnitFlat}, 900, "billing.period"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.billing.Validate(testCase.price)
			if testCase.wantField == "" {
				if err != nil {
					t.Fatalf("Validate() rejected a valid basis: %v", err)
				}
				return
			}
			var field FieldError
			if !errors.As(err, &field) || field.Field != testCase.wantField {
				t.Fatalf("Validate() error = %v, want a FieldError on %s", err, testCase.wantField)
			}
		})
	}
}

// A product with no recorded basis is not a valid product: that is exactly the
// state that let annual and monthly figures be compared as alike.
func TestProductValidateRequiresABillingBasis(t *testing.T) {
	product := Product{
		ID: "product-id", CategoryID: "category-id", BrandID: "brand-id",
		Name: "Demo Subscription", Slug: "demo-subscription",
		Description: "A fictional demo subscription.",
		Price:       Money{AmountMinor: 2_900, Currency: "USD"},
		Scores:      Scores{Quality: 80, Value: 80, Durability: 80, Beginner: 80, Advanced: 80, Portability: 80},
		Status:      ProductStatusPublished,
	}
	var field FieldError
	if err := product.Validate(); !errors.As(err, &field) || field.Field != "billing.period" {
		t.Fatalf("Validate() error = %v, want a FieldError on billing.period", err)
	}
	product.Billing = Billing{Period: BillingMonthly, Unit: PricingUnitPerUser}
	if err := product.Validate(); err != nil {
		t.Fatalf("Validate() rejected a product with a basis: %v", err)
	}
}

func TestBillingNormalizedDropsABlankNoteAndTrimsARealOne(t *testing.T) {
	blank := " \t"
	if got := (Billing{UnitNote: &blank}).Normalized(); got.UnitNote != nil {
		t.Fatalf("blank note kept as %q", *got.UnitNote)
	}
	padded := "  per seat, minimum 3 seats "
	got := (Billing{UnitNote: &padded}).Normalized()
	if got.UnitNote == nil || *got.UnitNote != "per seat, minimum 3 seats" {
		t.Fatalf("note not trimmed: %#v", got.UnitNote)
	}
	if padded != "  per seat, minimum 3 seats " {
		t.Fatal("Normalized() mutated the caller's string")
	}
	if got := (Billing{}).Normalized(); got.UnitNote != nil {
		t.Fatal("absent note became present")
	}
}
