package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	catalogapp "rigmark/internal/modules/catalog/application"
	catalogdomain "rigmark/internal/modules/catalog/domain"
	catalogports "rigmark/internal/modules/catalog/ports"
	contentports "rigmark/internal/modules/content/ports"
)

const spaIndexRedirect = "/__unsolero_spa/index.html"

// defaultSocialImage backs every page that has no image of its own.
//
// The metadata block promises a summary_large_image card, so one has to exist.
// Two things kept breaking that promise: products carry no images at all, and
// every editorial hero is an SVG, which LinkedIn, X and Facebook all refuse to
// render. Both fell through to no og:image and produced a bare grey card on
// exactly the platforms this site gets shared on.
const defaultSocialImage = "/images/og-default.png"

var (
	publicRouteSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	uuidRoutePattern       = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
)

var staticPublicRoutes = map[string]bool{
	"/": true, "/products": true, "/guides": true, "/articles": true,
	// The three index pages the navigation depends on. Every category and
	// brand had a page long before there was anywhere listing them, so a
	// visitor could reach one only by already knowing its URL.
	"/categories": true, "/brands": true, "/how-it-works": true,
	// The head-to-heads had no index either. /compare is the comparison tool
	// a visitor drives themselves; /comparisons is the writing.
	"/comparisons": true,
	// Indexable on purpose. Affiliate programme reviewers and search engines
	// both look for these two before approving or ranking a commercial site,
	// so a 404 here is expensive.
	"/about": true, "/privacy": true, "/affiliate-disclosure": true,
	"/terms": true,
	"/login": false, "/register": false, "/check-email": false,
	"/verify-email": false, "/forgot-password": false, "/reset-password": false,
	"/login/mfa": false, "/build": false, "/compare": false,
	"/wishlist": false, "/setups": false, "/account": false,
	"/design-system": false,
	"/admin":         false, "/admin/products": false, "/admin/products/new": false,
	"/admin/evidence": false, "/admin/categories": false, "/admin/brands": false,
	"/admin/merchants": false, "/admin/offers": false, "/admin/commerce": false,
	"/admin/affiliate-links": false, "/admin/recommendations": false,
	"/admin/users": false, "/admin/events": false, "/admin/content": false,
	"/admin/settings": false,
}

// publicRouteStatus is an edge-internal route manifest resolver. Nginx sends
// the original browser URI in X-Original-URI and honors X-Accel-Redirect only
// after this handler has established that the route exists. It returns no
// product or account data.
func (h *Handler) publicRouteStatus(response http.ResponseWriter, request *http.Request) {
	rawURI := request.Header.Get("X-Original-URI")
	if rawURI == "" || len(rawURI) > 2048 || strings.ContainsAny(rawURI, "\r\n\\") {
		h.writeRouteStatusError(response, http.StatusBadRequest)
		return
	}
	parsed, err := url.ParseRequestURI(rawURI)
	if err != nil || parsed.Path == "" || strings.Contains(parsed.Path, "//") ||
		containsInvalidPathRune(parsed.Path) || strings.Contains(strings.ToLower(parsed.EscapedPath()), "%2f") ||
		strings.Contains(strings.ToLower(parsed.EscapedPath()), "%5c") {
		h.writeRouteStatusError(response, http.StatusBadRequest)
		return
	}
	if parsed.Path != "/" && strings.HasSuffix(parsed.Path, "/") {
		canonicalPath := strings.TrimRight(parsed.Path, "/")
		if parsed.RawQuery != "" {
			canonicalPath += "?" + parsed.RawQuery
		}
		response.Header().Set("Location", canonicalPath)
		response.WriteHeader(http.StatusPermanentRedirect)
		return
	}

	meta, known, lookupErr := h.resolvePublicRoute(request, parsed.Path)
	if lookupErr != nil {
		if errors.Is(lookupErr, catalogports.ErrNotFound) || errors.Is(lookupErr, contentports.ErrNotFound) {
			h.writeRouteStatusError(response, http.StatusNotFound)
			return
		}
		h.logger.Error("resolve public browser route", "error", lookupErr)
		h.writeRouteStatusError(response, http.StatusServiceUnavailable)
		return
	}
	if !known {
		h.writeRouteStatusError(response, http.StatusNotFound)
		return
	}
	if parsed.RawQuery != "" {
		meta.Indexable = false
	}
	response.Header().Set("Cache-Control", "no-store")
	if !meta.Indexable {
		response.Header().Set("X-Robots-Tag", "noindex, nofollow")
	} else if meta.CanonicalURL != "" {
		response.Header().Set("Link", "<"+meta.CanonicalURL+">; rel=\"canonical\"")
	}

	// Serving the shell with this route's metadata is what makes search
	// results and social previews specific to the page. If the shell cannot be
	// fetched the edge serves the static file instead: losing metadata is a
	// degradation, losing the page is an outage.
	// Resolved here rather than inside renderShell, which has no handler and so
	// no way to make a site-relative path absolute. Social scrapers do not
	// resolve relative image paths.
	if meta.ImageURL == "" || strings.HasSuffix(strings.ToLower(meta.ImageURL), ".svg") {
		meta.ImageURL = h.absoluteImageURL(defaultSocialImage)
	}

	if h.shell != nil {
		if shell, ok := h.shell.Shell(request.Context()); ok {
			if rendered, rendered_ok := renderShell(shell, meta); rendered_ok {
				// This response is a document, not an API payload, so it needs
				// the document policy. The API-wide header would block every
				// script the page loads.
				response.Header().Set("Content-Security-Policy", documentContentSecurityPolicy)
				response.Header().Set("Content-Type", "text/html; charset=utf-8")
				response.WriteHeader(http.StatusOK)
				if _, err := response.Write([]byte(rendered)); err != nil {
					h.logger.Error("write public route html", "error", err)
				}
				return
			}
		}
	}

	response.Header().Set("X-Accel-Redirect", spaIndexRedirect)
	response.WriteHeader(http.StatusOK)
}

func containsInvalidPathRune(path string) bool {
	for _, value := range strings.Split(path, "/") {
		if value == "." || value == ".." {
			return true
		}
		for _, character := range value {
			if character < 0x20 || character == 0x7f {
				return true
			}
		}
	}
	return false
}

// resolvePublicRoute reports whether a browser route exists and describes it.
// The catalog and content lookups it performs to answer the first question are
// the same ones that answer the second, so metadata costs no extra queries.
func (h *Handler) resolvePublicRoute(request *http.Request, path string) (pageMetadata, bool, error) {
	meta := pageMetadata{Indexable: false, CanonicalURL: h.absolutePublicRoute(path)}

	if indexable, ok := staticPublicRoutes[path]; ok {
		meta.Indexable = indexable
		meta.Title, meta.Description = staticRouteMetadata(path)
		return meta, true, nil
	}
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) == 3 && segments[0] == "admin" &&
		(segments[1] == "products" || segments[1] == "evidence" || segments[1] == "recommendations") &&
		uuidRoutePattern.MatchString(segments[2]) {
		return pageMetadata{}, true, nil
	}
	if len(segments) != 2 {
		return pageMetadata{}, false, nil
	}
	section, value := segments[0], segments[1]
	if !publicRouteSlugPattern.MatchString(value) {
		if (section == "setups" || section == "admin") && uuidRoutePattern.MatchString(value) {
			return pageMetadata{}, section == "setups", nil
		}
		return pageMetadata{}, false, nil
	}
	switch section {
	case "products":
		if h.catalog == nil {
			return pageMetadata{}, false, nil
		}
		detail, err := h.catalog.GetProduct(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		meta.Indexable = true
		meta.Title = detail.Product.Name + " — " + detail.Product.BrandName + " | UNSOLERO"
		meta.Description = truncateDescription(detail.Product.Description, 155)
		if len(detail.Product.Images) > 0 {
			meta.ImageURL = h.absoluteImageURL(detail.Product.Images[0].URL)
		}
		meta.StructuredData = productStructuredData(detail.Product, meta.CanonicalURL)
		meta.PrerenderedBody = renderProductBody(detail.Product)
		return meta, true, nil
	case "categories":
		if h.catalog == nil {
			return pageMetadata{}, false, nil
		}
		category, err := h.catalog.GetCategory(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		// The page still resolves — someone following an old link gets the
		// category and its empty state rather than a 404. It is simply not
		// offered to a search engine until it has something to show.
		meta.Indexable = category.PublishedProducts > 0
		meta.Title = category.Name + " — compared on structured facts | UNSOLERO"
		meta.Description = truncateDescription(defaultText(category.Description,
			"Compare "+category.Name+" on structured facts, with no commission-driven ranking."), 155)
		meta.PrerenderedBody = renderCatalogListingBody(category.Name, category.Description,
			h.listingProducts(request.Context(), catalogapp.Query{CategorySlug: value}))
		return meta, true, nil
	case "brands":
		if h.catalog == nil {
			return pageMetadata{}, false, nil
		}
		brand, err := h.catalog.GetBrand(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		meta.Indexable = brand.PublishedProducts > 0
		meta.Title = brand.Name + " — plans, pricing and fit | UNSOLERO"
		meta.Description = truncateDescription(defaultText(brand.Description,
			"Products from "+brand.Name+", assessed on structured facts."), 155)
		meta.PrerenderedBody = renderCatalogListingBody(brand.Name, brand.Description,
			h.listingProducts(request.Context(), catalogapp.Query{BrandSlug: value}))
		return meta, true, nil
	case "guides", "articles", "compare":
		if h.content == nil {
			return pageMetadata{}, false, nil
		}
		entry, err := h.content.Get(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		if entry.Path != path {
			return pageMetadata{}, false, contentports.ErrNotFound
		}
		meta.Indexable = true
		meta.CanonicalURL = entry.CanonicalURL
		meta.Title = defaultText(entry.SEOTitle, entry.Title+" | UNSOLERO")
		meta.Description = truncateDescription(defaultText(entry.SEODescription, entry.Description), 155)
		meta.ImageURL = h.absoluteImageURL(entry.HeroImageURL)
		meta.StructuredData = h.articleStructuredData(entry, meta.CanonicalURL)
		meta.PrerenderedBody = renderEntryBody(entry)
		return meta, true, nil
	case "author":
		if h.content == nil {
			return pageMetadata{}, false, nil
		}
		author, entries, err := h.content.Author(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		meta.Indexable = true
		meta.Title = author.Name + " | UNSOLERO"
		meta.Description = truncateDescription(author.Bio, 155)
		meta.StructuredData = map[string]any{
			"@context": "https://schema.org", "@type": "Person",
			"name": author.Name, "description": author.Bio, "url": meta.CanonicalURL,
		}
		meta.PrerenderedBody = renderAuthorBody(author, entries)
		return meta, true, nil
	case "setups":
		return pageMetadata{}, uuidRoutePattern.MatchString(value), nil
	case "admin":
		// Only the UUID-bearing admin detail routes are dynamic. Static admin
		// routes are enumerated above so typos still receive a real 404.
		return pageMetadata{}, false, nil
	default:
		return pageMetadata{}, false, nil
	}
}

func defaultText(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

// absoluteImageURL turns a site-relative image path into the absolute URL that
// social preview scrapers require; they do not resolve relative paths.
func (h *Handler) absoluteImageURL(value string) string {
	if value == "" || strings.HasPrefix(value, "https://") {
		return value
	}
	return h.absolutePublicRoute(value)
}

func staticRouteMetadata(path string) (string, string) {
	switch path {
	case "/":
		return "UNSOLERO — Build the right software stack",
			"Tell us what your business does, what you already run, and what you can spend. We work out what you actually need."
	case "/products":
		return "Business software, judged on what matters | UNSOLERO",
			"Structured facts, suitability scores and comparable offers, with commission excluded from the ranking."
	case "/guides":
		return "Software stack guides | UNSOLERO",
			"Practical guides for planning a software stack around real constraints rather than feature lists."
	case "/articles":
		return "Software stack articles | UNSOLERO",
			"Editorial notes on choosing, combining and replacing the tools a business runs on."
	case "/comparisons":
		return "Software comparisons, head to head | UNSOLERO",
			"Direct comparisons of the business software people weigh against each other, with every price read from the vendor and the billing basis stated."
	case "/categories":
		return "Every software category we cover | UNSOLERO",
			"Fifteen categories of business software, from CRM and invoicing to analytics and help desk, each with the tools compared inside it."
	case "/brands":
		return "Every software vendor we cover | UNSOLERO",
			"The vendors in the UNSOLERO catalog, listed alphabetically with the products compared for each."
	case "/how-it-works":
		return "How UNSOLERO works | UNSOLERO",
			"Where the prices come from, how the scores are decided, and why commission cannot move a ranking."
	case "/about":
		return "About UNSOLERO | Who runs this site",
			"UNSOLERO is built and run by Andon Pediev. Who writes it, where the facts come from, and what it does not yet claim to know."
	case "/privacy":
		return "Privacy policy | UNSOLERO",
			"What we store, why we store it, and how to have it removed."
	case "/terms":
		return "Terms of use | UNSOLERO",
			"What UNSOLERO is, what it does not promise, who your purchase contract is with, and the rules for using the site."
	case "/affiliate-disclosure":
		return "Affiliate disclosure | UNSOLERO",
			"How UNSOLERO earns, and why commission is excluded from the ranking."
	// Not indexable, so these titles never reach a search result. They reach the
	// browser tab on the first paint, before the app has booted and set its own.
	case "/login":
		return "Sign in | UNSOLERO",
			"Sign in to UNSOLERO to return to your saved software decisions and comparisons."
	case "/register":
		return "Create an account | UNSOLERO",
			"Create a free UNSOLERO account to save software setups, comparisons and recommendations."
	default:
		return "UNSOLERO", "Independent business software recommendations built around your real constraints."
	}
}

func (h *Handler) absolutePublicRoute(path string) string {
	if h.content == nil {
		return ""
	}
	return h.content.AbsoluteURL(path)
}

func (h *Handler) writeRouteStatusError(response http.ResponseWriter, status int) {
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("X-Robots-Tag", "noindex, nofollow")
	response.WriteHeader(status)
}

// listingProducts fetches the products a category or brand page lists, for the
// server-rendered body only. A failure here must not fail the page: the
// document still carries its heading and metadata, and the application renders
// the full listing once it runs. The cap keeps the shell small — a crawler
// needs a route into each product, not the whole catalogue on one page.
func (h *Handler) listingProducts(ctx context.Context, query catalogapp.Query) []catalogdomain.Product {
	if h.catalog == nil {
		return nil
	}
	query.Page, query.PageSize = 1, 60
	page, err := h.catalog.Search(ctx, query)
	if err != nil {
		h.logger.Warn("listing products for prerendered body", "error", err)
		return nil
	}
	return page.Products
}
