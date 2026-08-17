package httpapi

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("X-Content-Type-Options", "nosniff")
		response.Header().Set("X-Frame-Options", "DENY")
		response.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		response.Header().Set("Permissions-Policy", "camera=(), geolocation=(), microphone=(), payment=(), usb=()")
		response.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		response.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		response.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		if request.TLS != nil || strings.EqualFold(request.Header.Get("X-Forwarded-Proto"), "https") {
			response.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		next.ServeHTTP(response, request)
	})
}

func apiCacheDefaults(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, "/api/") {
			response.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(response, request)
	})
}

func sameOriginProtection(next http.Handler, logger *slog.Logger, allowedOrigin string) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet || request.Method == http.MethodHead || request.Method == http.MethodOptions {
			next.ServeHTTP(response, request)
			return
		}
		origin := request.Header.Get("Origin")
		if origin == "" {
			if strings.EqualFold(request.Header.Get("Sec-Fetch-Site"), "cross-site") {
				writeOriginError(response, logger)
				return
			}
			next.ServeHTTP(response, request)
			return
		}
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" ||
			(parsed.Scheme != "http" && parsed.Scheme != "https") || !originAllowed(parsed, request, allowedOrigin) {
			writeOriginError(response, logger)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func originAllowed(origin *url.URL, request *http.Request, allowedOrigin string) bool {
	if allowedOrigin != "" {
		return strings.EqualFold(origin.Scheme+"://"+origin.Host, allowedOrigin)
	}
	return strings.EqualFold(origin.Host, request.Host)
}

func writeOriginError(response http.ResponseWriter, logger *slog.Logger) {
	writeAPIError(
		response,
		http.StatusForbidden,
		"origin_not_allowed",
		"This request did not originate from this application.",
		nil,
		logger,
	)
}
