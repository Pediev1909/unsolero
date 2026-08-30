package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	catalog "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
	commercedomain "rigmark/internal/modules/commerce/domain"
)

type commerceStub struct {
	offers           []commercedomain.Offer
	destination      string
	trackedID        commercedomain.AffiliateLinkID
	trackedOffer     commercedomain.OfferID
	trackedPromotion commercedomain.PromotionSlug
	surface          string
	purchasable      map[catalogdomain.ProductID]commercedomain.PurchasableOffer
	purchasableIDs   []catalogdomain.ProductID
	purchasableErr   error
	purchasableCalls int
}

func (stub *commerceStub) ListOffers(context.Context, catalogdomain.ProductID, string) ([]commercedomain.Offer, error) {
	return stub.offers, nil
}

func (stub *commerceStub) ListPurchasable(
	_ context.Context,
	ids []catalogdomain.ProductID,
) (map[catalogdomain.ProductID]commercedomain.PurchasableOffer, error) {
	stub.purchasableCalls++
	stub.purchasableIDs = ids
	return stub.purchasable, stub.purchasableErr
}

func (stub *commerceStub) TrackOfferClick(_ context.Context, click commercedomain.AffiliateClick) (commercedomain.AffiliateRedirectResult, error) {
	stub.trackedOffer = click.OfferID
	stub.surface = click.Source
	return commercedomain.AffiliateRedirectResult{DestinationURL: stub.destination}, nil
}

func (stub *commerceStub) TrackLegacyLinkClick(_ context.Context, click commercedomain.AffiliateClick) (commercedomain.AffiliateRedirectResult, error) {
	stub.trackedID = click.LinkID
	stub.surface = click.Source
	return commercedomain.AffiliateRedirectResult{DestinationURL: stub.destination}, nil
}

func (stub *commerceStub) TrackPromotionClick(_ context.Context, click commercedomain.AffiliateClick) (commercedomain.AffiliateRedirectResult, error) {
	stub.trackedPromotion = click.PromotionSlug
	stub.surface = click.Source
	return commercedomain.AffiliateRedirectResult{DestinationURL: stub.destination}, nil
}

func TestOfferResponseExposesTrackedPathNotMerchantURLs(t *testing.T) {
	offer := commercedomain.Offer{
		ID: "offer-1", ProductID: "product-1", ProductURL: "https://merchant.invalid/raw-product",
		Merchant: commercedomain.Merchant{
			ID: "merchant-1", Name: "Demo Merchant", Slug: "demo-merchant",
			WebsiteURL: "https://merchant.invalid", CountryCode: "US", TrustScore: 88,
		},
		Price: catalogdomain.Money{AmountMinor: 12900, Currency: "USD"}, Availability: "in_stock",
		Condition: "new", LastCheckedAt: time.Date(2026, 8, 16, 12, 0, 0, 0, time.UTC),
		AffiliateLinks: []commercedomain.AffiliateLink{{
			ID: "affiliate-1", DestinationURL: "https://merchant.invalid/secret-affiliate",
			DisclosureLabel: "Commissionable", IsActive: true,
		}},
	}

	encoded, err := json.Marshal(offerDTO(offer))
	if err != nil {
		t.Fatalf("marshal offer response: %v", err)
	}
	body := string(encoded)
	if !strings.Contains(body, `"purchase_path":"/api/affiliate/click/offer-1"`) {
		t.Fatalf("response does not contain tracked path: %s", body)
	}
	if strings.Contains(body, "merchant.invalid") || strings.Contains(body, "secret-affiliate") {
		t.Fatalf("response exposed raw merchant URL: %s", body)
	}
}

func TestAffiliateRedirectTracksOfferAndAttributionBeforeRedirecting(t *testing.T) {
	const offerID = "4ba7d524-9fd5-4d18-8c42-778c42d996f3"
	const sessionID = "1191bb26-a9a2-41df-9346-74d693350ce8"
	commerce := &commerceStub{destination: "https://merchant.invalid/destination"}
	handler := &Handler{
		commerce: commerce,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	request := httptest.NewRequest(http.MethodGet, "/api/affiliate/click/"+offerID+"?source=wishlist&session_id="+sessionID, nil)
	request.SetPathValue("offerID", offerID)
	response := httptest.NewRecorder()

	handler.affiliateClickRedirect(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusFound)
	}
	if commerce.trackedOffer != commercedomain.OfferID(offerID) || commerce.surface != "wishlist" {
		t.Fatalf("tracked redirect = (%q, %q)", commerce.trackedOffer, commerce.surface)
	}
	if response.Header().Get("Location") != commerce.destination {
		t.Fatalf("Location = %q", response.Header().Get("Location"))
	}
}

func TestOutboundRedirectRejectsMalformedIdentifier(t *testing.T) {
	commerce := &commerceStub{destination: "https://merchant.invalid/destination"}
	handler := &Handler{
		commerce: commerce,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	request := httptest.NewRequest(http.MethodGet, "/api/affiliate/click/not-a-uuid", nil)
	request.SetPathValue("offerID", "not-a-uuid")
	response := httptest.NewRecorder()

	handler.affiliateClickRedirect(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
	if commerce.trackedOffer != "" {
		t.Fatal("malformed identifier reached the commerce service")
	}
}

func TestAffiliatePromotionRedirectUsesStableSlugAndAttribution(t *testing.T) {
	const sessionID = "1191bb26-a9a2-41df-9346-74d693350ce8"
	commerce := &commerceStub{destination: "https://merchant.invalid/training?affiliate_id=123"}
	handler := &Handler{
		commerce: commerce,
		logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	request := httptest.NewRequest(http.MethodGet,
		"/api/affiliate/promotion/funnel-training?source=promotion&session_id="+sessionID, nil)
	request.SetPathValue("slug", "funnel-training")
	response := httptest.NewRecorder()

	handler.affiliatePromotionRedirect(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusFound)
	}
	if commerce.trackedPromotion != "funnel-training" || commerce.surface != "promotion" {
		t.Fatalf("tracked promotion = (%q, %q)", commerce.trackedPromotion, commerce.surface)
	}
	if response.Header().Get("Location") != commerce.destination {
		t.Fatalf("Location = %q", response.Header().Get("Location"))
	}
}

func TestCatalogQueryRejectsInvalidPagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/catalog/products?page=0", nil)
	if _, err := catalogQuery(request); err != catalog.ErrInvalidQuery {
		t.Fatalf("catalogQuery() error = %v, want %v", err, catalog.ErrInvalidQuery)
	}
}

// A grid of cards must cost one commerce query, not one per card. The catalog
// listing draws twenty-four, and the per-card alternative is what kept the
// vendor button off these surfaces in the first place.
func TestAttachPurchasePathsAsksOnceForTheWholePage(t *testing.T) {
	commerce := &commerceStub{
		purchasable: map[catalogdomain.ProductID]commercedomain.PurchasableOffer{
			"product-1": {
				OfferID: "offer-1", MerchantName: "ActiveCampaign",
				DisclosureLabel: "Affiliate link",
			},
		},
	}
	handler := &Handler{commerce: commerce, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	summaries := []productSummaryResponse{
		{ID: "product-1"}, {ID: "product-2"}, {ID: "product-3"},
	}

	handler.attachPurchasePaths(context.Background(), summaries)

	if commerce.purchasableCalls != 1 {
		t.Fatalf("commerce queried %d times for one page", commerce.purchasableCalls)
	}
	if len(commerce.purchasableIDs) != 3 {
		t.Fatalf("asked about %d products, want 3", len(commerce.purchasableIDs))
	}
	if summaries[0].PurchasePath == nil ||
		*summaries[0].PurchasePath != "/api/affiliate/click/offer-1" {
		t.Fatalf("purchase path = %v", summaries[0].PurchasePath)
	}
	if summaries[0].MerchantName == nil || *summaries[0].MerchantName != "ActiveCampaign" {
		t.Fatalf("merchant name = %v", summaries[0].MerchantName)
	}
	// A product with no servable offer gets no button, not a broken one.
	if summaries[1].PurchasePath != nil || summaries[1].MerchantName != nil {
		t.Fatal("product without an offer was given a purchase path")
	}
}

// The catalog is the product; the button is the business model. If commerce is
// down, a reader should still get the facts.
func TestAttachPurchasePathsSurvivesACommerceFailure(t *testing.T) {
	commerce := &commerceStub{purchasableErr: errors.New("commerce unavailable")}
	handler := &Handler{commerce: commerce, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	summaries := []productSummaryResponse{{ID: "product-1", Name: "ActiveCampaign Starter"}}

	handler.attachPurchasePaths(context.Background(), summaries)

	if summaries[0].PurchasePath != nil {
		t.Fatal("a failed lookup produced a purchase path")
	}
	if summaries[0].Name != "ActiveCampaign Starter" {
		t.Fatal("a failed lookup damaged the product facts")
	}
}

// A nil commerce service is how several tests and the admin router build a
// handler. It must not panic a public listing.
func TestAttachPurchasePathsToleratesNoCommerceService(t *testing.T) {
	handler := &Handler{logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	summaries := []productSummaryResponse{{ID: "product-1"}}
	handler.attachPurchasePaths(context.Background(), summaries)
	if summaries[0].PurchasePath != nil {
		t.Fatal("purchase path set without a commerce service")
	}
}

// freshness_status was the literal "fresh" on every offer this API ever
// served. That was survivable only because OFFER_MAXIMUM_AGE was 72 hours and
// nothing older could be returned — and that same 72-hour cut is what took
// every affiliate link offline three days after its seed last ran. Widening
// the window is the fix; this is what stops the widening from turning the
// field into a claim nobody checked.
func TestOfferFreshnessReportsTheAgeOfThePriceRatherThanAConstant(t *testing.T) {
	cases := map[string]struct {
		age  time.Duration
		want string
	}{
		"read minutes ago": {time.Minute, "fresh"},
		"read yesterday":   {24 * time.Hour, "fresh"},
		"read a week ago":  {6 * 24 * time.Hour, "fresh"},
		"read eight days":  {8 * 24 * time.Hour, "stale"},
		"read last month":  {30 * 24 * time.Hour, "stale"},
	}
	for name, testCase := range cases {
		offer := commercedomain.Offer{
			ID: "offer-1", LastCheckedAt: time.Now().Add(-testCase.age),
			Price: catalogdomain.Money{AmountMinor: 1500, Currency: "USD"},
		}
		if got := offerDTO(offer).FreshnessStatus; got != testCase.want {
			t.Fatalf("%s: freshness = %q, want %q", name, got, testCase.want)
		}
	}
}

// The date a price was read is the site's whole claim to being trustworthy
// about prices. It must survive into the response whatever the status says.
func TestOfferResponseAlwaysCarriesTheDateThePriceWasRead(t *testing.T) {
	// Fixed and long past, so the assertion below does not quietly change
	// meaning as the calendar moves past the seven-day window.
	checked := time.Date(2026, 6, 26, 10, 48, 13, 0, time.UTC)
	offer := commercedomain.Offer{
		ID: "offer-1", LastCheckedAt: checked,
		Price: catalogdomain.Money{AmountMinor: 1500, Currency: "USD"},
	}
	encoded, err := json.Marshal(offerDTO(offer))
	if err != nil {
		t.Fatalf("marshal offer: %v", err)
	}
	if !strings.Contains(string(encoded), `"last_checked_at":"2026-06-26T10:48:13Z"`) {
		t.Fatalf("response lost the read date: %s", encoded)
	}
	if !strings.Contains(string(encoded), `"freshness_status":"stale"`) {
		t.Fatalf("a price read two months ago is not fresh: %s", encoded)
	}
}
