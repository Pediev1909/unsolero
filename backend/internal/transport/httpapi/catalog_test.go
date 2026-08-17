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
	"time"

	catalog "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
	commercedomain "rigmark/internal/modules/commerce/domain"
)

type commerceStub struct {
	offers       []commercedomain.Offer
	destination  string
	trackedID    commercedomain.AffiliateLinkID
	trackedOffer commercedomain.OfferID
	surface      string
}

func (stub *commerceStub) ListOffers(context.Context, catalogdomain.ProductID, string) ([]commercedomain.Offer, error) {
	return stub.offers, nil
}

func (stub *commerceStub) TrackOfferClick(_ context.Context, click commercedomain.AffiliateClick) (string, error) {
	stub.trackedOffer = click.OfferID
	stub.surface = click.Source
	return stub.destination, nil
}

func (stub *commerceStub) TrackLegacyLinkClick(_ context.Context, click commercedomain.AffiliateClick) (string, error) {
	stub.trackedID = click.LinkID
	stub.surface = click.Source
	return stub.destination, nil
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

func TestCatalogQueryRejectsInvalidPagination(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/catalog/products?page=0", nil)
	if _, err := catalogQuery(request); err != catalog.ErrInvalidQuery {
		t.Fatalf("catalogQuery() error = %v, want %v", err, catalog.ErrInvalidQuery)
	}
}
