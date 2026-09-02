package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// pagedCatalogStub serves a fixed product list through the real paging
// contract: it honours Page and PageSize and reports TotalPages the way the
// application does, so a handler that walks the catalog is exercised against
// more products than one page holds.
type pagedCatalogStub struct {
	routeCatalogStub
	products  []catalogdomain.Product
	pageSizes []int
	err       error
}

func (stub *pagedCatalogStub) Search(_ context.Context, query catalog.Query) (catalog.Page, error) {
	if stub.err != nil {
		return catalog.Page{}, stub.err
	}
	if query.PageSize <= 0 || query.PageSize > catalog.MaximumPageSize || query.Page <= 0 {
		return catalog.Page{}, catalog.ErrInvalidQuery
	}
	stub.pageSizes = append(stub.pageSizes, query.PageSize)
	start := (query.Page - 1) * query.PageSize
	end := min(start+query.PageSize, len(stub.products))
	if start > len(stub.products) {
		start, end = len(stub.products), len(stub.products)
	}
	totalPages := (len(stub.products) + query.PageSize - 1) / query.PageSize
	return catalog.Page{
		Products: stub.products[start:end], Page: query.Page, PageSize: query.PageSize,
		Total: len(stub.products), TotalPages: totalPages,
	}, nil
}

func liveOffersFixture(count int) []catalogdomain.Product {
	products := make([]catalogdomain.Product, 0, count)
	for index := range count {
		category := "Email marketing"
		if index%2 == 1 {
			category = "CRM"
		}
		products = append(products, catalogdomain.Product{
			ID:           catalogdomain.ProductID(fmt.Sprintf("product-%d", index)),
			Name:         fmt.Sprintf("Product %d", index),
			Slug:         fmt.Sprintf("product-%d", index),
			CategoryName: category, CategorySlug: strings.ToLower(category),
			BrandName: "Vendor", BrandSlug: "vendor",
			Price: catalogdomain.Money{AmountMinor: 1000 + int64(index), Currency: "USD"},
		})
	}
	return products
}

func TestListLiveOffersWalksTheWholeCatalogAndAsksCommerceOnce(t *testing.T) {
	// More products than one catalog page holds, so a handler that reads only
	// the first page would silently drop the offers on the second.
	products := liveOffersFixture(catalog.MaximumPageSize + 5)
	// An odd index on the second page: odd is CRM in the fixture, and second
	// page is the part a single-page read would miss.
	secondPage := len(products) - 1
	if secondPage%2 == 0 {
		secondPage--
	}
	commerce := &commerceStub{
		purchasable: map[catalogdomain.ProductID]commercedomain.PurchasableOffer{
			"product-0": {
				OfferID: "offer-0", MerchantName: "Mailchimp", DisclosureLabel: "Affiliate link",
				Price:         catalogdomain.Money{AmountMinor: 1300, Currency: "USD"},
				LastCheckedAt: time.Now().Add(-time.Hour),
			},
			products[secondPage].ID: {
				OfferID: "offer-last", MerchantName: "HubSpot", DisclosureLabel: "Affiliate link",
				Price:         catalogdomain.Money{AmountMinor: 2000, Currency: "USD"},
				LastCheckedAt: time.Now().Add(-10 * 24 * time.Hour),
			},
		},
	}
	stub := &pagedCatalogStub{products: products}
	handler := &Handler{catalog: stub, commerce: commerce, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	response := httptest.NewRecorder()

	handler.listLiveOffers(response, httptest.NewRequest(http.MethodGet, "/api/catalog/offers", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if got := response.Header().Get("Cache-Control"); got != "public, max-age=30" {
		t.Fatalf("Cache-Control = %q", got)
	}
	if commerce.purchasableCalls != 1 {
		t.Fatalf("commerce asked %d times, want 1", commerce.purchasableCalls)
	}
	if len(commerce.purchasableIDs) != len(products) {
		t.Fatalf("commerce asked about %d products, want %d", len(commerce.purchasableIDs), len(products))
	}
	if len(stub.pageSizes) != 2 {
		t.Fatalf("catalog read in %d pages, want 2", len(stub.pageSizes))
	}

	var body liveOffersResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, err := time.Parse(time.RFC3339, body.GeneratedAt); err != nil {
		t.Fatalf("generated_at = %q: %v", body.GeneratedAt, err)
	}
	if len(body.Items) != 2 {
		t.Fatalf("items = %d, want only the two products commerce vouched for", len(body.Items))
	}
	// Sorted by category: the second-page product is CRM, which sorts before
	// Email marketing even though the catalog returned it last.
	first, second := body.Items[0], body.Items[1]
	if first.Product.Category.Name != "CRM" || second.Product.Category.Name != "Email marketing" {
		t.Fatalf("category order = (%q, %q)", first.Product.Category.Name, second.Product.Category.Name)
	}
	if first.Offer.FreshnessStatus != "stale" || second.Offer.FreshnessStatus != "fresh" {
		t.Fatalf("freshness = (%q, %q)", first.Offer.FreshnessStatus, second.Offer.FreshnessStatus)
	}
	if second.Product.PurchasePath == nil || *second.Product.PurchasePath != "/api/affiliate/click/offer-0" {
		t.Fatalf("purchase path = %v", second.Product.PurchasePath)
	}
	if second.Product.MerchantName == nil || *second.Product.MerchantName != "Mailchimp" ||
		second.Offer.MerchantName != "Mailchimp" {
		t.Fatalf("merchant = %v / %q", second.Product.MerchantName, second.Offer.MerchantName)
	}
	if second.Offer.Price.AmountMinor != 1300 || second.Offer.Price.Currency != "USD" {
		t.Fatalf("offer price = %+v", second.Offer.Price)
	}
	if _, err := time.Parse("2006-01-02T15:04:05Z", second.Offer.LastCheckedAt); err != nil {
		t.Fatalf("last_checked_at = %q: %v", second.Offer.LastCheckedAt, err)
	}
}

// The catalog grid tolerates a commerce outage and draws without buttons,
// because the facts are still the page. Here the offers are the page, and an
// empty 200 would tell a reader nothing is on offer when the truth is that
// nobody could check.
func TestListLiveOffersFailsClosedWhenCommerceIsUnavailable(t *testing.T) {
	commerce := &commerceStub{purchasableErr: errors.New("commerce down")}
	stub := &pagedCatalogStub{products: liveOffersFixture(3)}
	handler := &Handler{catalog: stub, commerce: commerce, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	response := httptest.NewRecorder()

	handler.listLiveOffers(response, httptest.NewRequest(http.MethodGet, "/api/catalog/offers", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(response.Body.String(), `"code":"offers_unavailable"`) {
		t.Fatalf("body = %s", response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("an error must not be cached as the offers list: %q", response.Header().Get("Cache-Control"))
	}
}

func TestListLiveOffersReportsACatalogFailureAsTheCatalogs(t *testing.T) {
	stub := &pagedCatalogStub{err: errors.New("catalog down")}
	handler := &Handler{catalog: stub, commerce: &commerceStub{}, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	response := httptest.NewRecorder()

	handler.listLiveOffers(response, httptest.NewRequest(http.MethodGet, "/api/catalog/offers", nil))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(response.Body.String(), `"code":"catalog_unavailable"`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestListLiveOffersWithNothingLiveIsAnEmptyListNotAnError(t *testing.T) {
	stub := &pagedCatalogStub{products: liveOffersFixture(3)}
	handler := &Handler{catalog: stub, commerce: &commerceStub{}, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	response := httptest.NewRecorder()

	handler.listLiveOffers(response, httptest.NewRequest(http.MethodGet, "/api/catalog/offers", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	// `items: []`, never `items: null` — a client that maps over the list
	// must not have to guard against a missing array.
	if !strings.Contains(response.Body.String(), `"items":[]`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}
