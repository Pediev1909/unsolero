package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	adminapp "rigmark/internal/modules/admin/application"
	adminports "rigmark/internal/modules/admin/ports"
	catalogapp "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
)

// The wording lives in one function; these are the sentences the contract
// promised the frontend.
func TestBillingPhrase(t *testing.T) {
	contacts := "at 1,000 contacts"
	transaction := "2.9% + 30¢ per transaction"
	tasks := "per 1,000 automation tasks"
	seats := "per seat, minimum 3 seats"
	cases := []struct {
		name    string
		billing catalogdomain.Billing
		want    string
	}{
		{"per user monthly", catalogdomain.Billing{Period: catalogdomain.BillingMonthly, Unit: catalogdomain.PricingUnitPerUser}, "Per user, monthly billing"},
		{"per user ignores the note", catalogdomain.Billing{Period: catalogdomain.BillingMonthly, Unit: catalogdomain.PricingUnitPerUser, UnitNote: &seats}, "Per user, monthly billing"},
		{"flat annual only", catalogdomain.Billing{Period: catalogdomain.BillingAnnual, Unit: catalogdomain.PricingUnitFlat}, "Flat rate, billed yearly"},
		{"flat free", catalogdomain.Billing{Period: catalogdomain.BillingFree, Unit: catalogdomain.PricingUnitFlat}, "Flat rate, free plan"},
		{"contacts with note", catalogdomain.Billing{Period: catalogdomain.BillingMonthly, Unit: catalogdomain.PricingUnitPerContacts, UnitNote: &contacts}, "At 1,000 contacts, monthly billing"},
		{"contacts without note", catalogdomain.Billing{Period: catalogdomain.BillingMonthly, Unit: catalogdomain.PricingUnitPerContacts}, "Per contact tier, monthly billing"},
		{"transaction with note", catalogdomain.Billing{Period: catalogdomain.BillingUsage, Unit: catalogdomain.PricingUnitPerTransaction, UnitNote: &transaction}, "2.9% + 30¢ per transaction"},
		{"transaction without note", catalogdomain.Billing{Period: catalogdomain.BillingUsage, Unit: catalogdomain.PricingUnitPerTransaction}, "Per transaction"},
		{"usage with note", catalogdomain.Billing{Period: catalogdomain.BillingUsage, Unit: catalogdomain.PricingUnitUsage, UnitNote: &tasks}, "Per 1,000 automation tasks"},
		{"usage without note", catalogdomain.Billing{Period: catalogdomain.BillingUsage, Unit: catalogdomain.PricingUnitUsage}, "Usage-based"},
		{"usage unit on a monthly plan keeps the period", catalogdomain.Billing{Period: catalogdomain.BillingMonthly, Unit: catalogdomain.PricingUnitUsage}, "Usage-based, monthly billing"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := billingPhrase(testCase.billing); got != testCase.want {
				t.Fatalf("billingPhrase() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestLowerFirstLeavesAnAcronymAlone(t *testing.T) {
	for input, want := range map[string]string{
		"Per user, monthly billing":  "per user, monthly billing",
		"API calls, monthly billing": "API calls, monthly billing",
		"2.9% + 30¢ per transaction": "2.9% + 30¢ per transaction",
		"":                           "",
	} {
		if got := lowerFirst(input); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", input, got, want)
		}
	}
}

// The summary is what every card, grid and alternative list is built from, so
// the exact JSON shape of `billing` is pinned here alongside the derived key
// specification.
func TestProductSummaryCarriesBillingAndDerivesTheKeySpecification(t *testing.T) {
	annual := int64(1_500)
	product := catalogdomain.Product{
		ID: "11111111-2222-4333-8444-555555555555", Name: "Zoho Books Standard", Slug: "zoho-books-standard",
		BrandName: "Zoho", BrandSlug: "zoho", CategoryName: "Accounting", CategorySlug: "accounting",
		Price:   catalogdomain.Money{AmountMinor: 2_000, Currency: "USD"},
		Billing: catalogdomain.Billing{Period: catalogdomain.BillingMonthly, Unit: catalogdomain.PricingUnitPerUser, AnnualPriceMinor: &annual},
	}
	encoded, err := json.Marshal(productSummaryDTO(product))
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Billing          json.RawMessage `json:"billing"`
		KeySpecification struct {
			Label string `json:"label"`
			Value string `json:"value"`
		} `json:"key_specification"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	const wantBilling = `{"period":"monthly","unit":"per_user","unit_note":null,"annual_price_minor":1500}`
	if string(decoded.Billing) != wantBilling {
		t.Fatalf("billing = %s, want %s", decoded.Billing, wantBilling)
	}
	if decoded.KeySpecification.Label != "Billing" || decoded.KeySpecification.Value != "Per user, monthly billing" {
		t.Fatalf("key_specification = %+v", decoded.KeySpecification)
	}

	// The detail response embeds the summary, so it carries the same object.
	detail, err := json.Marshal(productDetailDTO(catalogapp.ProductDetail{Product: product}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(detail), `"billing":`+wantBilling) {
		t.Fatalf("detail response lacks billing: %s", detail)
	}
}

// A service with no repository is enough to exercise validation: it rejects
// the input before reaching the repository, and the embedded nil interface
// would panic if it did not.
type adminRepositoryStub struct{ adminports.Repository }

func TestAdminCreateProductNamesTheBillingFieldInThe422(t *testing.T) {
	handler := &Handler{
		admin:  adminapp.NewService(adminRepositoryStub{}, nil),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	body := `{"category_id":"c","brand_id":"b","name":"Kit Creator","slug":"kit-creator",
		"description":"Email for creators.","price_minor":2900,"currency":"USD",
		"billing":{"period":"annual","unit":"flat","unit_note":null,"annual_price_minor":2500},
		"warranty_months":0,"quality_score":80,"value_score":80,"durability_score":80,
		"beginner_score":80,"advanced_score":80,"apartment_score":0,"noise_score":0,"portability_score":80}`
	request := httptest.NewRequest(http.MethodPost, "/api/admin/products", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.adminCreateProduct(response, request.WithContext(context.Background()))

	if response.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422: %s", response.Code, response.Body.String())
	}
	var envelope struct {
		Error struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != "invalid_admin_input" || envelope.Error.Fields["billing.annual_price_minor"] == "" {
		t.Fatalf("error envelope = %s", response.Body.String())
	}
}

// A request that omits billing altogether is told which field is missing
// rather than left to guess from "check the submitted fields".
func TestAdminCreateProductWithoutBillingIsToldSo(t *testing.T) {
	handler := &Handler{
		admin:  adminapp.NewService(adminRepositoryStub{}, nil),
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	body := `{"category_id":"c","brand_id":"b","name":"Kit Creator","slug":"kit-creator",
		"description":"Email for creators.","price_minor":2900,"currency":"USD",
		"warranty_months":0,"quality_score":80,"value_score":80,"durability_score":80,
		"beginner_score":80,"advanced_score":80,"apartment_score":0,"noise_score":0,"portability_score":80}`
	request := httptest.NewRequest(http.MethodPost, "/api/admin/products", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.adminCreateProduct(response, request)

	if response.Code != http.StatusUnprocessableEntity || !strings.Contains(response.Body.String(), `"billing.period"`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}
