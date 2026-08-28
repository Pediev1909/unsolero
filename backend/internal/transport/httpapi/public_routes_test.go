package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	catalog "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	contentdomain "rigmark/internal/modules/content/domain"
	contentports "rigmark/internal/modules/content/ports"
)

type routeCatalogStub struct{}

func (routeCatalogStub) Search(context.Context, catalog.Query) (catalog.Page, error) {
	return catalog.Page{}, nil
}
func (routeCatalogStub) GetProduct(_ context.Context, slug string) (catalog.ProductDetail, error) {
	if slug != "known-product" {
		return catalog.ProductDetail{}, catalogports.ErrNotFound
	}
	return catalog.ProductDetail{}, nil
}
func (routeCatalogStub) ListCategories(context.Context) ([]catalogdomain.Category, error) {
	return nil, nil
}
func (routeCatalogStub) GetCategory(_ context.Context, slug string) (catalogdomain.Category, error) {
	if slug != "known-category" {
		return catalogdomain.Category{}, catalogports.ErrNotFound
	}
	return catalogdomain.Category{}, nil
}
func (routeCatalogStub) ListBrands(context.Context) ([]catalogdomain.Brand, error) { return nil, nil }
func (routeCatalogStub) ListBrandsInCategory(context.Context, string) ([]catalogdomain.Brand, error) {
	return nil, nil
}
func (routeCatalogStub) GetBrand(_ context.Context, slug string) (catalogdomain.Brand, error) {
	if slug != "known-brand" {
		return catalogdomain.Brand{}, catalogports.ErrNotFound
	}
	return catalogdomain.Brand{}, nil
}

// Same contract as routeCatalogStub, but the listings report how much is
// actually published under them.
type emptyListingCatalog struct{ routeCatalogStub }

func (emptyListingCatalog) GetCategory(_ context.Context, slug string) (catalogdomain.Category, error) {
	switch slug {
	case "crm":
		return catalogdomain.Category{Slug: slug, Name: "CRM", PublishedProducts: 2}, nil
	case "analytics":
		return catalogdomain.Category{Slug: slug, Name: "Analytics"}, nil
	}
	return catalogdomain.Category{}, catalogports.ErrNotFound
}

func (emptyListingCatalog) GetBrand(_ context.Context, slug string) (catalogdomain.Brand, error) {
	if slug == "nobody" {
		return catalogdomain.Brand{Slug: slug, Name: "Nobody"}, nil
	}
	return catalogdomain.Brand{}, catalogports.ErrNotFound
}

type routeContentStub struct{}

func (routeContentStub) List(context.Context, string, string, int) ([]contentdomain.Summary, error) {
	return nil, nil
}
func (routeContentStub) Get(_ context.Context, slug string) (contentdomain.Entry, error) {
	if slug != "known-guide" {
		return contentdomain.Entry{}, contentports.ErrNotFound
	}
	return contentdomain.Entry{Summary: contentdomain.Summary{Path: "/guides/known-guide"}, CanonicalURL: "https://unsolero.example/guides/known-guide"}, nil
}
func (routeContentStub) Author(context.Context, string) (contentdomain.Author, []contentdomain.Summary, error) {
	return contentdomain.Author{}, nil, nil
}

func (routeContentStub) Sitemap(context.Context) ([]contentdomain.SitemapEntry, error) {
	return nil, nil
}
func (routeContentStub) AbsoluteURL(path string) string { return "https://unsolero.example" + path }

func TestPublicRouteStatusManifestAndCanonicalBehavior(t *testing.T) {
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{Catalog: routeCatalogStub{}, Content: routeContentStub{}})
	tests := []struct {
		name, uri string
		status    int
		accel     bool
		noindex   bool
	}{
		{"known static public", "/products", http.StatusOK, true, false},
		// These two shipped in the router and the footer but not here, so they
		// answered 404 and noindex. They are what an affiliate programme
		// review opens first, which made the omission expensive and silent.
		{"about", "/about", http.StatusOK, true, false},
		{"privacy policy", "/privacy", http.StatusOK, true, false},
		{"affiliate disclosure", "/affiliate-disclosure", http.StatusOK, true, false},
		{"funnel training offer", "/offers/funnel-hacking-secrets", http.StatusOK, true, false},
		{"known dynamic public", "/products/known-product", http.StatusOK, true, false},
		{"known editorial", "/guides/known-guide", http.StatusOK, true, false},
		{"unknown public", "/products/not-present", http.StatusNotFound, false, true},
		{"unknown path", "/definitely-not-a-route", http.StatusNotFound, false, true},
		{"admin protected", "/admin/products", http.StatusOK, true, true},
		{"admin detail protected", "/admin/products/12345678-1234-4234-8234-123456789abc", http.StatusOK, true, true},
		{"account protected", "/account", http.StatusOK, true, true},
		{"malformed slash", "/products//known-product", http.StatusBadRequest, false, true},
		{"encoded slash", "/products%2Fknown-product", http.StatusBadRequest, false, true},
		{"encoded control", "/products/known%00product", http.StatusBadRequest, false, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/v1/public-route", nil)
			request.Header.Set("X-Original-URI", test.uri)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status=%d, want %d", response.Code, test.status)
			}
			if got := response.Header().Get("X-Accel-Redirect") != ""; got != test.accel {
				t.Fatalf("accel=%v, want %v", got, test.accel)
			}
			if got := strings.Contains(response.Header().Get("X-Robots-Tag"), "noindex"); got != test.noindex {
				t.Fatalf("noindex=%v, want %v", got, test.noindex)
			}
		})
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/public-route", nil)
	request.Header.Set("X-Original-URI", "/products/?page=2")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusPermanentRedirect || response.Header().Get("Location") != "/products?page=2" {
		t.Fatalf("trailing slash response=%d location=%q", response.Code, response.Header().Get("Location"))
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/public-route", nil)
	request.Header.Set("X-Original-URI", "/products?page=2")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Header().Get("X-Robots-Tag"), "noindex") {
		t.Fatalf("filtered catalog indexability status=%d robots=%q", response.Code, response.Header().Get("X-Robots-Tag"))
	}
	request = httptest.NewRequest(http.MethodGet, "/api/v1/public-route", nil)
	request.Header.Set("X-Original-URI", "/products/known-product")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Header().Get("Link") != `<https://unsolero.example/products/known-product>; rel="canonical"` {
		t.Fatalf("canonical link=%q", response.Header().Get("Link"))
	}
}

func TestPublicRouteStatusRequiresEdgeHeaderAndDoesNotCaptureAPIs(t *testing.T) {
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/public-route", nil))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("missing edge header status=%d", response.Code)
	}
	apiResponse := httptest.NewRecorder()
	router.ServeHTTP(apiResponse, httptest.NewRequest(http.MethodGet, "/api/not-a-route", nil))
	if apiResponse.Code != http.StatusNotFound {
		t.Fatalf("unknown API status=%d", apiResponse.Code)
	}
}

// Twelve of fifteen categories had no products at all, yet every one of them
// returned 200 with no robots directive and sat in the sitemap. That is an
// invitation to index nothing, and it is the shape of thin content that the
// 2026 core updates penalise hardest on exactly this kind of site.
func TestEmptyListingsAreNotOfferedToSearchEngines(t *testing.T) {
	t.Parallel()

	handler := &Handler{catalog: emptyListingCatalog{}, content: nil}

	meta, known, err := handler.resolvePublicRoute(
		httptest.NewRequest(http.MethodGet, "/", nil), "/categories/analytics")
	if err != nil || !known {
		t.Fatalf("category route should resolve: known=%v err=%v", known, err)
	}
	if meta.Indexable {
		t.Error("a category with no published products must not be indexable")
	}

	meta, known, err = handler.resolvePublicRoute(
		httptest.NewRequest(http.MethodGet, "/", nil), "/categories/crm")
	if err != nil || !known {
		t.Fatalf("category route should resolve: known=%v err=%v", known, err)
	}
	if !meta.Indexable {
		t.Error("a category that has products must stay indexable")
	}

	meta, known, err = handler.resolvePublicRoute(
		httptest.NewRequest(http.MethodGet, "/", nil), "/brands/nobody")
	if err != nil || !known {
		t.Fatalf("brand route should resolve: known=%v err=%v", known, err)
	}
	if meta.Indexable {
		t.Error("a brand with no published products must not be indexable")
	}
}
