package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	catalogports "rigmark/internal/modules/catalog/ports"
	contentports "rigmark/internal/modules/content/ports"
)

const spaIndexRedirect = "/__unsolero_spa/index.html"

var (
	publicRouteSlugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	uuidRoutePattern       = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[1-5][a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
)

var staticPublicRoutes = map[string]bool{
	"/": true, "/products": true, "/guides": true, "/articles": true,
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
		meta.Description = truncateDescription(detail.Product.Description, 160)
		if len(detail.Product.Images) > 0 {
			meta.ImageURL = h.absoluteImageURL(detail.Product.Images[0].URL)
		}
		meta.StructuredData = productStructuredData(detail.Product, meta.CanonicalURL)
		return meta, true, nil
	case "categories":
		if h.catalog == nil {
			return pageMetadata{}, false, nil
		}
		category, err := h.catalog.GetCategory(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		meta.Indexable = true
		meta.Title = category.Name + " | UNSOLERO"
		meta.Description = truncateDescription(defaultText(category.Description,
			"Compare "+category.Name+" on structured facts, with no commission-driven ranking."), 160)
		return meta, true, nil
	case "brands":
		if h.catalog == nil {
			return pageMetadata{}, false, nil
		}
		brand, err := h.catalog.GetBrand(request.Context(), value)
		if err != nil {
			return pageMetadata{}, false, err
		}
		meta.Indexable = true
		meta.Title = brand.Name + " | UNSOLERO"
		meta.Description = truncateDescription(defaultText(brand.Description,
			"Products from "+brand.Name+", assessed on structured facts."), 160)
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
		meta.Description = truncateDescription(defaultText(entry.SEODescription, entry.Description), 160)
		meta.ImageURL = h.absoluteImageURL(entry.HeroImageURL)
		meta.StructuredData = articleStructuredData(entry, meta.CanonicalURL)
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
