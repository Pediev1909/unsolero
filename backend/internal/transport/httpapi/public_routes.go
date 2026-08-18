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

	indexable, known, canonicalURL, lookupErr := h.resolvePublicRoute(request, parsed.Path)
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
		indexable = false
	}
	response.Header().Set("Cache-Control", "no-store")
	if !indexable {
		response.Header().Set("X-Robots-Tag", "noindex, nofollow")
	} else if canonicalURL != "" {
		response.Header().Set("Link", "<"+canonicalURL+">; rel=\"canonical\"")
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

func (h *Handler) resolvePublicRoute(request *http.Request, path string) (bool, bool, string, error) {
	if indexable, ok := staticPublicRoutes[path]; ok {
		return indexable, true, h.absolutePublicRoute(path), nil
	}
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(segments) == 3 && segments[0] == "admin" &&
		(segments[1] == "products" || segments[1] == "evidence" || segments[1] == "recommendations") &&
		uuidRoutePattern.MatchString(segments[2]) {
		return false, true, "", nil
	}
	if len(segments) != 2 {
		return false, false, "", nil
	}
	section, value := segments[0], segments[1]
	if !publicRouteSlugPattern.MatchString(value) {
		if (section == "setups" || section == "admin") && uuidRoutePattern.MatchString(value) {
			return false, section == "setups", "", nil
		}
		return false, false, "", nil
	}
	switch section {
	case "products":
		if h.catalog == nil {
			return false, false, "", nil
		}
		_, err := h.catalog.GetProduct(request.Context(), value)
		return true, err == nil, h.absolutePublicRoute(path), err
	case "categories":
		if h.catalog == nil {
			return false, false, "", nil
		}
		_, err := h.catalog.GetCategory(request.Context(), value)
		return true, err == nil, h.absolutePublicRoute(path), err
	case "brands":
		if h.catalog == nil {
			return false, false, "", nil
		}
		_, err := h.catalog.GetBrand(request.Context(), value)
		return true, err == nil, h.absolutePublicRoute(path), err
	case "guides", "articles", "compare":
		if h.content == nil {
			return false, false, "", nil
		}
		entry, err := h.content.Get(request.Context(), value)
		if err != nil {
			return false, false, "", err
		}
		if entry.Path != path {
			return false, false, "", contentports.ErrNotFound
		}
		return true, true, entry.CanonicalURL, nil
	case "setups":
		return false, uuidRoutePattern.MatchString(value), "", nil
	case "admin":
		// Only the UUID-bearing admin detail routes are dynamic. Static admin
		// routes are enumerated above so typos still receive a real 404.
		return false, false, "", nil
	default:
		return false, false, "", nil
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
