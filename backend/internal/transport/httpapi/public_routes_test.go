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
	catalogports "rigmark/internal/modules/catalog/ports"
	commercedomain "rigmark/internal/modules/commerce/domain"
	contentapp "rigmark/internal/modules/content/application"
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

func (routeContentStub) List(context.Context, contentapp.ListQuery) ([]contentdomain.Summary, error) {
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
		// The link in a newsletter email has to open a page. Both token
		// landings resolve, and neither is offered to a search engine.
		{"newsletter confirm", "/newsletter/confirm", http.StatusOK, true, true},
		{"newsletter unsubscribe", "/newsletter/unsubscribe", http.StatusOK, true, true},
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

// Fixtures for the prerendered static bodies. The catalog stub answers every
// listing the bodies draw from; the content stub answers a hub and one entry.

var (
	prerenderCRMProduct = catalogdomain.Product{
		ID: "11111111-1111-4111-8111-111111111111", Name: "Zoho CRM Standard", Slug: "zoho-crm-standard",
		BrandName: "Zoho", BrandSlug: "zoho", CategoryName: "CRM", CategorySlug: "crm",
		Price: catalogdomain.Money{AmountMinor: 1400, Currency: "USD"},
	}
	prerenderHelpDeskProduct = catalogdomain.Product{
		ID: "22222222-2222-4222-8222-222222222222", Name: "Freshdesk Growth", Slug: "freshdesk-growth",
		BrandName: "Freshworks", BrandSlug: "freshworks", CategoryName: "Help desk", CategorySlug: "help-desk",
		Price: catalogdomain.Money{AmountMinor: 1500, Currency: "USD"},
	}
)

type prerenderCatalog struct {
	routeCatalogStub
	products   []catalogdomain.Product
	categories []catalogdomain.Category
	brands     []catalogdomain.Brand
}

func (stub prerenderCatalog) Search(_ context.Context, query catalog.Query) (catalog.Page, error) {
	products := stub.products
	if query.PageSize > 0 && len(products) > query.PageSize {
		products = products[:query.PageSize]
	}
	return catalog.Page{Products: products, Page: 1, PageSize: len(products), Total: len(stub.products), TotalPages: 1}, nil
}

func (stub prerenderCatalog) GetProduct(_ context.Context, slug string) (catalog.ProductDetail, error) {
	for _, product := range stub.products {
		if product.Slug == slug {
			return catalog.ProductDetail{Product: product}, nil
		}
	}
	return catalog.ProductDetail{}, catalogports.ErrNotFound
}

func (stub prerenderCatalog) ListCategories(context.Context) ([]catalogdomain.Category, error) {
	return stub.categories, nil
}

func (stub prerenderCatalog) ListBrands(context.Context) ([]catalogdomain.Brand, error) {
	return stub.brands, nil
}

type prerenderContent struct {
	routeContentStub
	entries []contentdomain.Summary
	entry   contentdomain.Entry
}

func (stub prerenderContent) List(context.Context, contentapp.ListQuery) ([]contentdomain.Summary, error) {
	return stub.entries, nil
}

func (stub prerenderContent) Get(_ context.Context, slug string) (contentdomain.Entry, error) {
	if stub.entry.Slug != slug {
		return contentdomain.Entry{}, contentports.ErrNotFound
	}
	return stub.entry, nil
}

func newPrerenderHandler(catalog CatalogService, content ContentService, commerce CommerceService) *Handler {
	return &Handler{
		catalog: catalog, content: content, commerce: commerce,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func resolveForTest(t *testing.T, handler *Handler, path string) pageMetadata {
	t.Helper()
	meta, known, err := handler.resolvePublicRoute(httptest.NewRequest(http.MethodGet, "/", nil), path)
	if err != nil || !known {
		t.Fatalf("%s should resolve: known=%v err=%v", path, known, err)
	}
	return meta
}

func structuredJSON(t *testing.T, meta pageMetadata) string {
	t.Helper()
	encoded, err := json.Marshal(meta.StructuredData)
	if err != nil {
		t.Fatalf("structured data does not marshal: %v", err)
	}
	return string(encoded)
}

func assertContainsAll(t *testing.T, label, got string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Errorf("%s is missing %q\ngot: %s", label, want, got)
		}
	}
}

// The home page answered with 3.5 KB whose only text was the <title>. A
// crawler that does not run the application saw a site with no heading and no
// way into the catalog.
func TestHomeRouteCarriesHeadingAndCategoryLinks(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(prerenderCatalog{
		products: []catalogdomain.Product{prerenderCRMProduct},
		categories: []catalogdomain.Category{
			{Name: "CRM", Slug: "crm", PublishedProducts: 1},
			{Name: "Analytics", Slug: "analytics"},
		},
	}, routeContentStub{}, nil)

	meta := resolveForTest(t, handler, "/")
	if !meta.Indexable {
		t.Fatal("the home page must stay indexable")
	}
	assertContainsAll(t, "home body", meta.PrerenderedBody,
		"<h1", "Build the right software stack.",
		`href="/build"`, `href="/categories"`, `href="/comparisons"`, `href="/guides"`,
		"Define your constraints", "Compare complete setups", "Buy in the right order",
		`href="/categories/crm"`, `href="/products/zoho-crm-standard"`, "USD 14.00", "entry paid tier",
	)
	// An empty category is a non-indexable page; the home page does not hand
	// its strongest links to one.
	if strings.Contains(meta.PrerenderedBody, `href="/categories/analytics"`) {
		t.Errorf("home body links a category with nothing published\ngot: %s", meta.PrerenderedBody)
	}
}

func TestProductsRouteListsEveryProductUnderItsCategory(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(prerenderCatalog{
		products: []catalogdomain.Product{prerenderHelpDeskProduct, prerenderCRMProduct},
	}, routeContentStub{}, nil)

	meta := resolveForTest(t, handler, "/products")
	assertContainsAll(t, "/products body", meta.PrerenderedBody,
		"<h1", "Software, judged on what matters.",
		`href="/products/zoho-crm-standard"`, `href="/products/freshdesk-growth"`,
		`href="/categories/crm"`, `href="/categories/help-desk"`,
	)
	// Groups come in category order, whatever order the catalog returned.
	if strings.Index(meta.PrerenderedBody, "CRM") > strings.Index(meta.PrerenderedBody, "Help desk") {
		t.Errorf("categories are not grouped alphabetically\ngot: %s", meta.PrerenderedBody)
	}
	assertContainsAll(t, "/products structured data", structuredJSON(t, meta),
		`"@type":"ItemList"`, `"numberOfItems":2`, `"@type":"BreadcrumbList"`,
		`"url":"https://unsolero.example/products/zoho-crm-standard"`,
	)
}

func TestCategoryAndBrandIndexesListCountsAndLinks(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(prerenderCatalog{
		categories: []catalogdomain.Category{{Name: "CRM", Slug: "crm", PublishedProducts: 2}},
		brands: []catalogdomain.Brand{
			{Name: "Zoho", Slug: "zoho", PublishedProducts: 1},
			{Name: "Nobody", Slug: "nobody"},
		},
	}, routeContentStub{}, nil)

	categories := resolveForTest(t, handler, "/categories")
	assertContainsAll(t, "/categories body", categories.PrerenderedBody,
		"<h1", `href="/categories/crm"`, "CRM", "2 products")

	brands := resolveForTest(t, handler, "/brands")
	assertContainsAll(t, "/brands body", brands.PrerenderedBody, "<h1", `href="/brands/zoho"`, "1 product")
	if strings.Contains(brands.PrerenderedBody, `href="/brands/nobody"`) {
		t.Errorf("a vendor with nothing published was listed\ngot: %s", brands.PrerenderedBody)
	}
}

func TestHubRouteListsItsEntries(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(routeCatalogStub{}, prerenderContent{
		entries: []contentdomain.Summary{{
			Title: "Choosing a CRM for a five-person agency", Path: "/guides/crm-for-agencies",
			Description: "What the entry tiers actually include.",
			PublishedAt: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC),
		}},
	}, nil)

	meta := resolveForTest(t, handler, "/guides")
	assertContainsAll(t, "/guides body", meta.PrerenderedBody,
		"<h1", `href="/guides/crm-for-agencies"`, "Choosing a CRM for a five-person agency",
		"What the entry tiers actually include.", `<time datetime="2026-08-19T00:00:00Z">`)
	assertContainsAll(t, "/guides structured data", structuredJSON(t, meta),
		`"@type":"ItemList"`, `"name":"Choosing a CRM for a five-person agency"`,
		`"@type":"BreadcrumbList"`, `"name":"Guides"`)
}

// The text pages are prose in a React component, so their server body is a
// copy. These pin that the copy is present and links where the page links.
func TestTextRoutesCarryTheirHeadlines(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(routeCatalogStub{}, routeContentStub{}, nil)
	for path, wants := range map[string][]string{
		"/how-it-works":         {"Start wherever you actually are.", "Every price was read from the vendor", `href="/categories"`},
		"/about":                {"Who is behind this site.", "Andon Pediev", "What this site does not claim"},
		"/affiliate-disclosure": {"How we make money.", "What commission does not do", "We do not accept payment for placement or ranking."},
	} {
		meta := resolveForTest(t, handler, path)
		assertContainsAll(t, path+" body", meta.PrerenderedBody, append([]string{"<h1"}, wants...)...)
	}
}

// /offers lists what a reader can act on now, and only that: a product whose
// offer commerce will not vouch for is not on offer. The affiliate redirect
// stays out of the static body; the link is to the product page.
func TestOffersRouteRendersOnlyPurchasableProducts(t *testing.T) {
	t.Parallel()

	commerce := &commerceStub{purchasable: map[catalogdomain.ProductID]commercedomain.PurchasableOffer{
		prerenderCRMProduct.ID: {OfferID: "offer-1", MerchantName: "Zoho Corporation"},
	}}
	handler := newPrerenderHandler(prerenderCatalog{
		products: []catalogdomain.Product{prerenderCRMProduct, prerenderHelpDeskProduct},
	}, routeContentStub{}, commerce)

	meta := resolveForTest(t, handler, "/offers")
	assertContainsAll(t, "/offers body", meta.PrerenderedBody,
		"<h1", `href="/products/zoho-crm-standard"`, "Zoho Corporation", "USD 14.00")
	if strings.Contains(meta.PrerenderedBody, "freshdesk-growth") {
		t.Errorf("a product without a live offer was listed\ngot: %s", meta.PrerenderedBody)
	}
	if strings.Contains(meta.PrerenderedBody, "/api/affiliate/") || strings.Contains(meta.PrerenderedBody, "offer-1") {
		t.Errorf("the static body must not carry the affiliate redirect\ngot: %s", meta.PrerenderedBody)
	}
	if commerce.purchasableCalls != 1 || len(commerce.purchasableIDs) != 2 {
		t.Errorf("commerce asked %d time(s) about %d product(s), want once about 2",
			commerce.purchasableCalls, len(commerce.purchasableIDs))
	}
	if meta.FallbackImageURL != "/images/og-offers.png" {
		t.Errorf("offers social card = %q", meta.FallbackImageURL)
	}
}

// A commerce outage must not turn /offers into a page claiming nothing is on
// offer with confidence; the body keeps its heading and drops the list.
func TestOffersRouteSurvivesACommerceFailure(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(prerenderCatalog{
		products: []catalogdomain.Product{prerenderCRMProduct},
	}, routeContentStub{}, &commerceStub{purchasableErr: context.DeadlineExceeded})

	meta := resolveForTest(t, handler, "/offers")
	if !strings.Contains(meta.PrerenderedBody, "<h1") || strings.Contains(meta.PrerenderedBody, "/products/") {
		t.Errorf("unexpected body after a commerce failure:\n%s", meta.PrerenderedBody)
	}
}

func TestProductRouteEmitsProductAndBreadcrumbTogether(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(prerenderCatalog{
		products: []catalogdomain.Product{prerenderCRMProduct},
	}, routeContentStub{}, nil)

	meta := resolveForTest(t, handler, "/products/zoho-crm-standard")
	encoded := structuredJSON(t, meta)
	assertContainsAll(t, "product structured data", encoded,
		`"@graph"`, `"@type":"Product"`, `"@type":"BreadcrumbList"`,
		`"name":"Home"`, `"item":"https://unsolero.example/"`,
		`"name":"CRM"`, `"item":"https://unsolero.example/categories/crm"`,
		`"name":"Zoho CRM Standard"`, `"position":3`,
	)
	// One @context for the whole block; a nested one is what validators flag.
	if strings.Count(encoded, `"@context"`) != 1 {
		t.Errorf("expected exactly one @context\ngot: %s", encoded)
	}
	if meta.FallbackImageURL != "/images/og-product.png" {
		t.Errorf("product social card = %q", meta.FallbackImageURL)
	}
}

func TestCategoryRouteEmitsBreadcrumbAndItemList(t *testing.T) {
	t.Parallel()

	handler := newPrerenderHandler(emptyListingCatalog{}, routeContentStub{}, nil)
	meta := resolveForTest(t, handler, "/categories/crm")
	assertContainsAll(t, "category structured data", structuredJSON(t, meta),
		`"@type":"BreadcrumbList"`, `"name":"Categories"`, `"name":"CRM"`, `"@type":"ItemList"`)
}

func guideEntry(blocks ...contentdomain.Block) contentdomain.Entry {
	entry := contentdomain.Entry{Content: blocks, CanonicalURL: "https://unsolero.example/guides/crm-for-agencies"}
	entry.Type = contentdomain.ContentTypeGuide
	entry.Slug = "crm-for-agencies"
	entry.Path = "/guides/crm-for-agencies"
	entry.Title = "Choosing a CRM for a five-person agency"
	entry.HeroImageURL = "/images/editorial/crm.svg"
	return entry
}

// FAQPage markup is emitted only when the page shows a FAQ. Declaring one the
// reader cannot see is exactly the structured-data overstatement that earns a
// manual action.
func TestEditorialRouteEmitsFAQPageOnlyWithAFAQBlock(t *testing.T) {
	t.Parallel()

	withFAQ := newPrerenderHandler(routeCatalogStub{}, prerenderContent{entry: guideEntry(
		contentdomain.Block{Type: contentdomain.BlockParagraph, Text: "Start with the pipeline."},
		contentdomain.Block{Type: contentdomain.BlockFAQ, Questions: []contentdomain.QuestionAnswer{
			{Question: "Does the free tier include email sync?", Answer: "No. It starts on the entry paid tier."},
		}},
	)}, nil)
	meta := resolveForTest(t, withFAQ, "/guides/crm-for-agencies")
	assertContainsAll(t, "editorial structured data", structuredJSON(t, meta),
		`"@type":"Article"`, `"@type":"BreadcrumbList"`, `"name":"Guides"`,
		`"@type":"FAQPage"`, `"@type":"Question"`, `"name":"Does the free tier include email sync?"`,
		`"@type":"Answer"`, `"text":"No. It starts on the entry paid tier."`)
	// An SVG hero is refused by every social platform; the typed raster card
	// stands in for it.
	if meta.FallbackImageURL != "/images/og-guide.png" {
		t.Errorf("guide social card = %q", meta.FallbackImageURL)
	}

	withoutFAQ := newPrerenderHandler(routeCatalogStub{}, prerenderContent{entry: guideEntry(
		contentdomain.Block{Type: contentdomain.BlockParagraph, Text: "Start with the pipeline."},
	)}, nil)
	meta = resolveForTest(t, withoutFAQ, "/guides/crm-for-agencies")
	if encoded := structuredJSON(t, meta); strings.Contains(encoded, "FAQPage") {
		t.Errorf("FAQPage emitted for an entry with no FAQ block\ngot: %s", encoded)
	}
}

func stackEntry(blocks ...contentdomain.Block) contentdomain.Entry {
	entry := contentdomain.Entry{Content: blocks, CanonicalURL: "https://unsolero.example/stacks/agency-3-people-under-150"}
	entry.Type = contentdomain.ContentTypeStack
	entry.Slug = "agency-3-people-under-150"
	entry.Path = "/stacks/agency-3-people-under-150"
	entry.Title = "A 3-person agency's software stack under $150 a month"
	entry.HeroImageURL = "/images/saas-agency-stack-v2.svg"
	return entry
}

// A stack is the fourth editorial section and goes through every switch the
// other three do: the hub is a static indexable route with a listing body, the
// entry resolves only under its own segment, and both carry the Stacks crumb.
func TestStackRoutesMirrorTheOtherEditorialSections(t *testing.T) {
	t.Parallel()

	entry := stackEntry(
		contentdomain.Block{Type: contentdomain.BlockParagraph, Text: "Three tools, 59 USD a month."},
		contentdomain.Block{Type: contentdomain.BlockFAQ, Questions: []contentdomain.QuestionAnswer{
			{Question: "Do three people need a project tool?", Answer: "Only the simplest one."},
		}},
	)
	handler := newPrerenderHandler(routeCatalogStub{}, prerenderContent{
		entries: []contentdomain.Summary{entry.Summary},
		entry:   entry,
	}, nil)

	hub := resolveForTest(t, handler, "/stacks")
	if !hub.Indexable {
		t.Fatal("the stacks hub must be indexable")
	}
	if hub.Title != "Software stacks, priced for one kind of business | UNSOLERO" {
		t.Errorf("stacks hub title = %q", hub.Title)
	}
	// The title carries an apostrophe, which the body escapes; the fragment
	// after it is what a reader and a crawler both see.
	assertContainsAll(t, "/stacks body", hub.PrerenderedBody,
		"<h1", "Software stacks", `href="/stacks/agency-3-people-under-150"`, "software stack under $150 a month")
	assertContainsAll(t, "/stacks structured data", structuredJSON(t, hub),
		`"@type":"ItemList"`, `"@type":"BreadcrumbList"`, `"name":"Stacks"`)

	page := resolveForTest(t, handler, "/stacks/agency-3-people-under-150")
	if !page.Indexable || page.CanonicalURL != entry.CanonicalURL {
		t.Errorf("stack page indexable=%v canonical=%q", page.Indexable, page.CanonicalURL)
	}
	if page.FallbackImageURL != "/images/og-stack.png" {
		t.Errorf("stack social card = %q", page.FallbackImageURL)
	}
	assertContainsAll(t, "stack structured data", structuredJSON(t, page),
		`"@type":"Article"`, `"@type":"BreadcrumbList"`, `"name":"Stacks"`, `/stacks"`,
		`"@type":"FAQPage"`, `"name":"Do three people need a project tool?"`)
	assertContainsAll(t, "stack body", page.PrerenderedBody, "<h1", "software stack under $150 a month", "Three tools, 59 USD a month.")

	// The same slug under another section is a different URL, and the entry
	// must not answer for it: two indexable addresses for one page is the
	// duplicate-content case the path check exists to prevent.
	_, known, err := handler.resolvePublicRoute(httptest.NewRequest(http.MethodGet, "/", nil), "/guides/agency-3-people-under-150")
	if known || !errors.Is(err, contentports.ErrNotFound) {
		t.Errorf("stack under /guides/: known=%v err=%v", known, err)
	}
}

// One card for every page made every shared link look the same. The card now
// says what kind of page the link is, and the default remains for anything the
// mapping does not know.
func TestSocialImageFollowsThePageType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path        string
		contentType contentdomain.ContentType
		want        string
	}{
		{"/compare/mailchimp-vs-kit", contentdomain.ContentTypeComparison, "/images/og-comparison.png"},
		{"/guides/crm-for-agencies", contentdomain.ContentTypeGuide, "/images/og-guide.png"},
		{"/guides/crm-for-agencies", contentdomain.ContentTypeBuyingGuide, "/images/og-guide.png"},
		{"/articles/how-unsolero-ranks-software", contentdomain.ContentTypeArticle, "/images/og-article.png"},
		{"/stacks/agency-3-people-under-150", contentdomain.ContentTypeStack, "/images/og-stack.png"},
		{"/stacks", "", defaultSocialImage},
		{"/products/zoho-crm-standard", "", "/images/og-product.png"},
		{"/products", "", "/images/og-product.png"},
		{"/categories/crm", "", "/images/og-product.png"},
		{"/brands/zoho", "", "/images/og-product.png"},
		{"/offers", "", "/images/og-offers.png"},
		// The builder's output is a stack, so a shared builder link previews
		// as one. og-stack.png existed for the editorial stacks and was
		// reachable from nowhere else.
		{"/build", "", "/images/og-stack.png"},
		{"/offers/funnel-hacking-secrets", "", defaultSocialImage},
		{"/", "", defaultSocialImage},
		{"/about", "", defaultSocialImage},
		{"/guides", "", defaultSocialImage},
	}
	for _, test := range tests {
		if got := socialImagePath(test.path, test.contentType); got != test.want {
			t.Errorf("socialImagePath(%q, %q) = %q, want %q", test.path, test.contentType, got, test.want)
		}
	}
}

// End to end through the edge handler: the shell comes back with this route's
// body inside the mount point and the typed card in og:image, resolved to an
// absolute URL the way scrapers need it.
func TestPublicRouteServesPrerenderedBodyAndTypedSocialCard(t *testing.T) {
	shellServer := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, testShell)
	}))
	defer shellServer.Close()

	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{
			Catalog: prerenderCatalog{
				products:   []catalogdomain.Product{prerenderCRMProduct},
				categories: []catalogdomain.Category{{Name: "CRM", Slug: "crm", PublishedProducts: 1}},
			},
			Content:     prerenderContent{entry: guideEntry(contentdomain.Block{Type: contentdomain.BlockParagraph, Text: "Start with the pipeline."})},
			SPAShellURL: shellServer.URL,
		})

	serve := func(uri string) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/api/v1/public-route", nil)
		request.Header.Set("X-Original-URI", uri)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		return response
	}

	home := serve("/")
	if home.Code != http.StatusOK || !strings.HasPrefix(home.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("home status=%d content-type=%q", home.Code, home.Header().Get("Content-Type"))
	}
	assertContainsAll(t, "home document", home.Body.String(),
		`<div id="root"><main`, "Build the right software stack.", `href="/categories/crm"`,
		`<meta property="og:image" content="https://unsolero.example/images/og-default.png" />`)

	guide := serve("/guides/crm-for-agencies")
	assertContainsAll(t, "guide document", guide.Body.String(),
		"Start with the pipeline.",
		`<meta property="og:image" content="https://unsolero.example/images/og-guide.png" />`)
	if strings.Contains(guide.Body.String(), "crm.svg") {
		t.Errorf("an SVG hero reached og:image:\n%s", guide.Body.String())
	}

	product := serve("/products/zoho-crm-standard")
	assertContainsAll(t, "product document", product.Body.String(),
		`<meta property="og:image" content="https://unsolero.example/images/og-product.png" />`,
		`"@type":"BreadcrumbList"`)
}
