package httpapi

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	health "rigmark/internal/modules/health/application"
	"rigmark/internal/platform/observability"
)

func TestOperationalMetricsRequireConfiguredBearerAndExposeOnlyAggregates(t *testing.T) {
	recorder := observability.NewMemoryRecorder()
	recorder.Increment(observability.MetricRateLimited)
	recorder.ObserveHTTP(http.MethodGet, "GET /api/catalog/products/{slug}", http.StatusOK, 12*time.Millisecond)
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{
			Metrics: recorder, MetricsToken: "a-monitoring-token-that-is-long-enough",
		})

	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil))
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d", unauthorized.Code)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil)
	request.Header.Set("Authorization", "Bearer a-monitoring-token-that-is-long-enough")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"rate_limit_rejected":1`) {
		t.Fatalf("response=%d %s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "a-monitoring-token") {
		t.Fatal("metrics response exposed credentials")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/v1/metrics/openmetrics", nil)
	request.Header.Set("Authorization", "Bearer a-monitoring-token-that-is-long-enough")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK ||
		!strings.Contains(response.Header().Get("Content-Type"), "application/openmetrics-text") ||
		!strings.Contains(response.Body.String(), "unsolero_http_requests_total") ||
		!strings.Contains(response.Body.String(), "unsolero_http_duration_milliseconds_bucket") ||
		!strings.Contains(response.Body.String(), "unsolero_rate_limit_rejected_total 1") ||
		!strings.HasSuffix(response.Body.String(), "# EOF\n") {
		t.Fatalf("OpenMetrics response=%d %s", response.Code, response.Body.String())
	}
}

func TestReadinessFailureClassifiesBoundedDependencies(t *testing.T) {
	recorder := observability.NewMemoryRecorder()
	recordReadinessFailure(recorder, health.Report{Checks: map[string]string{
		"database": "unavailable", "schema": "unavailable", "rate_limit": "unavailable", "media_storage": "unavailable",
	}})
	counters := recorder.Snapshot().Counters
	for _, name := range []string{
		observability.MetricReadinessFailure, observability.MetricDatabaseAcquireFailure,
		observability.MetricMigrationFailure, observability.MetricRedisUnavailable, observability.MetricStorageFailure,
	} {
		if counters[name] != 1 {
			t.Fatalf("counter %s = %d, want 1", name, counters[name])
		}
	}
}
