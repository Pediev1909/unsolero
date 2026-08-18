package httpapi

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"rigmark/internal/platform/observability"
)

func TestObservabilityDoesNotLogQueriesOrPanicValues(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	handler := requestObservability(recoverPanics(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("reset_token=secret-value") }), logger), logger, observability.DisabledRecorder{})
	request := httptest.NewRequest(http.MethodGet, "/api/private-email@example.test?token=secret-query", nil)
	handler.ServeHTTP(httptest.NewRecorder(), request)
	logged := output.String()
	for _, forbidden := range []string{"secret-query", "secret-value", "reset_token", "private-email@example.test"} {
		if strings.Contains(logged, forbidden) {
			t.Fatalf("sensitive value %q appeared in log: %s", forbidden, logged)
		}
	}
	if !strings.Contains(logged, "unmatched") || !strings.Contains(logged, "panic_type") {
		t.Fatalf("safe observability context missing: %s", logged)
	}
}

func TestObservabilityCapturesResolvedPatternAcrossRequestCopies(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	recorder := observability.NewMemoryRecorder()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/products/{slug}", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusNoContent)
	})
	handler := requestObservability(
		requestDeadline(captureRoutePattern(mux), time.Second),
		logger,
		recorder,
	)

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/products/fictional-item", nil))

	logged := output.String()
	if !strings.Contains(logged, `"route":"GET /api/products/{slug}"`) {
		t.Fatalf("resolved route pattern missing: %s", logged)
	}
	for _, metric := range recorder.Snapshot().HTTP {
		if metric.Route != "GET /api/products/{slug}" {
			t.Fatalf("metric route=%q, want resolved pattern", metric.Route)
		}
	}
}
