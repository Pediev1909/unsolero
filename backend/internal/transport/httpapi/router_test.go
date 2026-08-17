package httpapi

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	proxied := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	proxied.RemoteAddr = "172.20.0.4:40000"
	proxied.Header.Set("X-Forwarded-For", "198.51.100.9")
	if value := clientAddress(proxied); value != "198.51.100.9" {
		t.Fatalf("proxied client address = %q", value)
	}

	direct := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	direct.RemoteAddr = "203.0.113.10:40000"
	direct.Header.Set("X-Forwarded-For", "198.51.100.9")
	if value := clientAddress(direct); value != "203.0.113.10" {
		t.Fatalf("direct client address trusted spoofed forwarding header: %q", value)
	}
}
