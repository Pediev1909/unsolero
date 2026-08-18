package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"rigmark/internal/platform/abuse"
	"rigmark/internal/platform/alerting"
	"rigmark/internal/platform/observability"

	health "rigmark/internal/modules/health/application"
)

type healthStub struct {
	err error
}

func (stub healthStub) Live() health.Report {
	return health.Report{Status: "ok", Service: "rigmark-api", Version: "test"}
}

func (stub healthStub) Check(context.Context) (health.Report, error) {
	if stub.err != nil {
		return health.Report{
			Status:  "degraded",
			Service: "rigmark-api",
			Version: "test",
			Checks:  map[string]string{"database": "unavailable"},
		}, stub.err
	}
	return health.Report{
		Status:  "ok",
		Service: "rigmark-api",
		Version: "test",
		Checks:  map[string]string{"database": "ok"},
	}, nil
}

func TestHealthRoutes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		serviceErr error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "public health",
			path:       "/api/health",
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok","service":"rigmark-api","version":"test","checks":{"database":"ok"}}`,
		},
		{
			name:       "liveness",
			path:       "/api/v1/health/live",
			wantStatus: http.StatusOK,
			wantBody:   `{"status":"ok","service":"rigmark-api","version":"test"}`,
		},
		{
			name:       "dependency unavailable",
			path:       "/api/health",
			serviceErr: errors.New("down"),
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"status":"degraded","service":"rigmark-api","version":"test","checks":{"database":"unavailable"}}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			NewRouter(
				healthStub{err: test.serviceErr},
				&authStub{},
				testCookieConfig,
				logger,
			).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if strings.TrimSpace(response.Body.String()) != test.wantBody {
				t.Errorf("body = %q, want %q", response.Body.String(), test.wantBody)
			}
			if response.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Error("expected security headers")
			}
			if response.Header().Get("Content-Security-Policy") == "" || response.Header().Get("Permissions-Policy") == "" {
				t.Error("expected browser security policy headers")
			}
			if !requestIDPattern.MatchString(response.Header().Get("X-Request-ID")) {
				t.Error("expected a generated request identifier")
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Error("expected no-store cache policy")
			}
		})
	}
}

func TestSensitiveEndpointRateLimit(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := rateLimitRequests(
		http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusNoContent)
		}),
		RateLimitConfig{AuthenticationPerMinute: 1},
		logger,
	)

	first := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	first.RemoteAddr = "203.0.113.8:12000"
	firstResponse := httptest.NewRecorder()
	handler.ServeHTTP(firstResponse, first)
	if firstResponse.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", firstResponse.Code, http.StatusNoContent)
	}

	second := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	second.RemoteAddr = "203.0.113.8:12001"
	secondResponse := httptest.NewRecorder()
	handler.ServeHTTP(secondResponse, second)
	if secondResponse.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", secondResponse.Code, http.StatusTooManyRequests)
	}
	if secondResponse.Header().Get("Retry-After") == "" {
		t.Error("rate-limited response did not include Retry-After")
	}
}

func TestRateLimitBackendFailureFailsClosed(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	metrics := observability.NewMemoryRecorder()
	handler := rateLimitRequestsWithBackend(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	}), RateLimitConfig{AuthenticationPerMinute: 1}, abuse.UnavailableLimiter{}, []byte("opaque-key-secret"), metrics, alerting.Disabled{}, logger)
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d", response.Code)
	}
	if metrics.Snapshot().Counters[observability.MetricRateLimitBackendFailure] != 1 {
		t.Fatal("backend failure metric was not recorded")
	}
}

func TestRateLimitKeyNeverContainsRawAddress(t *testing.T) {
	key := opaqueRateLimitKey([]byte("stable-secret"), "authentication", "203.0.113.22")
	if strings.Contains(key, "203.0.113.22") || !strings.HasPrefix(key, "authentication:") {
		t.Fatalf("unsafe key=%q", key)
	}
}

func TestAccountSecurityEndpointsUseAuthenticationRateBucket(t *testing.T) {
	config := DefaultRateLimitConfig()
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/api/auth/email-verification/complete", nil),
		httptest.NewRequest(http.MethodPost, "/api/auth/mfa/complete", nil),
		httptest.NewRequest(http.MethodPost, "/api/account/security/password", nil),
		httptest.NewRequest(http.MethodDelete, "/api/account", nil),
	} {
		bucket, limit, protected := rateLimitRule(request, config)
		if !protected || limit != config.AuthenticationPerMinute || (bucket != "authentication" && bucket != "account-security") {
			t.Fatalf("%s %s bucket=%q limit=%d protected=%v", request.Method, request.URL.Path, bucket, limit, protected)
		}
	}
}

func TestSensitiveRoutesUseSeparateRateLimitPolicies(t *testing.T) {
	config := DefaultRateLimitConfig()
	tests := []struct {
		method string
		path   string
		bucket string
		limit  int
	}{
		{http.MethodPost, "/api/auth/register", "registration", config.RegistrationPerMinute},
		{http.MethodPost, "/api/auth/password-reset/request", "password-reset", config.PasswordResetPerMinute},
		{http.MethodPost, "/api/analytics/events", "analytics", config.AnalyticsPerMinute},
		{http.MethodGet, "/api/affiliate/click/offer", "affiliate", config.AffiliatePerMinute},
		{http.MethodGet, "/api/admin/users", "admin", config.AdminPerMinute},
		{http.MethodPost, "/api/admin/products", "admin", config.AdminPerMinute},
	}
	for _, test := range tests {
		request := httptest.NewRequest(test.method, test.path, nil)
		bucket, limit, protected := rateLimitRule(request, config)
		if !protected || bucket != test.bucket || limit != test.limit {
			t.Errorf("%s %s = (%q, %d, %v)", test.method, test.path, bucket, limit, protected)
		}
	}
}

func TestEveryAdminRouteUsesPermissionMiddleware(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve router test path")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(testFile), "router.go"))
	if err != nil {
		t.Fatalf("read router source: %v", err)
	}

	adminRoutes := 0
	for lineNumber, line := range strings.Split(string(source), "\n") {
		if !strings.Contains(line, "mux.Handle(") || !strings.Contains(line, "/api/admin/") {
			continue
		}
		adminRoutes++
		protected := strings.Contains(line, "allowed(") ||
			strings.Contains(line, "evidenceEditor(") ||
			strings.Contains(line, "evidenceReviewer(") ||
			strings.Contains(line, "policyEditor(") ||
			strings.Contains(line, "policyReviewer(")
		if !protected {
			t.Errorf("router.go:%d registers an admin route without an explicit permission wrapper", lineNumber+1)
		}
	}
	if adminRoutes < 40 {
		t.Fatalf("inspected %d admin routes, want at least 40", adminRoutes)
	}
}

func TestCrossSiteFetchMetadataRejectsMutationWithoutOrigin(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := sameOriginProtection(
		http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
			response.WriteHeader(http.StatusNoContent)
		}),
		logger,
		"https://rigmark.example",
	)
	request := httptest.NewRequest(http.MethodPost, "/api/account/comparison", nil)
	request.Header.Set("Sec-Fetch-Site", "cross-site")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestRateLimitClientAddressTrustsOnlyPrivateProxy(t *testing.T) {
	trustedProxies := []netip.Prefix{netip.MustParsePrefix("172.20.0.0/24")}
	proxied := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	proxied.RemoteAddr = "172.20.0.4:40000"
	proxied.Header.Set("X-Forwarded-For", "198.51.100.9")
	if value := clientAddress(proxied, trustedProxies); value != "198.51.100.9" {
		t.Fatalf("proxied client address = %q", value)
	}
	if value := clientAddress(proxied, nil); value != "172.20.0.4" {
		t.Fatalf("unconfigured private proxy was trusted: %q", value)
	}

	direct := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	direct.RemoteAddr = "203.0.113.10:40000"
	direct.Header.Set("X-Forwarded-For", "198.51.100.9")
	if value := clientAddress(direct, trustedProxies); value != "203.0.113.10" {
		t.Fatalf("direct client address trusted spoofed forwarding header: %q", value)
	}
}
