package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"rigmark/internal/platform/observability"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{8,128}$`)

type routePatternState struct {
	pattern string
}

type routePatternStateKey struct{}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (recorder *responseRecorder) WriteHeader(status int) {
	if recorder.status != 0 {
		return
	}
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func (recorder *responseRecorder) Write(data []byte) (int, error) {
	if recorder.status == 0 {
		recorder.WriteHeader(http.StatusOK)
	}
	written, err := recorder.ResponseWriter.Write(data)
	recorder.bytes += written
	return written, err
}

func requestObservability(next http.Handler, logger *slog.Logger, metrics observability.Recorder) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		startedAt := time.Now()
		routeState := &routePatternState{}
		request = request.WithContext(context.WithValue(request.Context(), routePatternStateKey{}, routeState))
		requestID := request.Header.Get("X-Request-ID")
		if !requestIDPattern.MatchString(requestID) {
			requestID = newRequestID()
		}
		request.Header.Set("X-Request-ID", requestID)
		response.Header().Set("X-Request-ID", requestID)
		recorder := &responseRecorder{ResponseWriter: response}
		next.ServeHTTP(recorder, request)
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		level := slog.LevelInfo
		route := routeState.pattern
		if route == "" {
			route = "unmatched"
		}
		if status < http.StatusBadRequest &&
			(route == "GET /api/v1/health/live" || route == "GET /api/v1/health/ready") {
			level = slog.LevelDebug
		}
		logger.Log(
			request.Context(),
			level,
			"HTTP request completed",
			"request_id", requestID,
			"method", request.Method,
			"route", route,
			"status", status,
			"response_bytes", recorder.bytes,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
		metrics.ObserveHTTP(request.Method, route, status, time.Since(startedAt))
		recordOperationalOutcome(metrics, route, status)
	})
}

// captureRoutePattern runs immediately around the standard library mux. The
// outer observability middleware cannot read request.Pattern directly because
// bounded request-context middleware creates request copies before dispatch.
func captureRoutePattern(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if state, ok := request.Context().Value(routePatternStateKey{}).(*routePatternState); ok {
				state.pattern = request.Pattern
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func recordOperationalOutcome(metrics observability.Recorder, pattern string, status int) {
	switch {
	case pattern == "POST /api/analytics/events" && status < http.StatusBadRequest:
		metrics.Increment(observability.MetricAnalyticsAccepted)
	case pattern == "POST /api/analytics/events" && status >= http.StatusBadRequest:
		metrics.Increment(observability.MetricAnalyticsRejected)
	case pattern == "POST /api/recommendations/generate" && status >= http.StatusInternalServerError:
		metrics.Increment(observability.MetricRecommendationFailure)
	case strings.HasPrefix(pattern, "POST /api/auth/") && (status == http.StatusUnauthorized || status == http.StatusForbidden):
		metrics.Increment(observability.MetricAuthenticationFailure)
	case pattern == "POST /api/webhooks/commerce/{providerConfigurationID}" && status >= http.StatusBadRequest:
		metrics.Increment(observability.MetricWebhookFailure)
	case strings.Contains(pattern, "/api/admin/commerce/imports") && status >= http.StatusInternalServerError:
		metrics.Increment(observability.MetricProviderFailure)
	case strings.Contains(pattern, "/api/admin/commerce/reconciliations") && status >= http.StatusInternalServerError:
		metrics.Increment(observability.MetricReconciliationFailure)
	}
}

func requestDeadline(next http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		ctx, cancel := context.WithTimeout(request.Context(), timeout)
		defer cancel()
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func recoverPanics(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error(
					"panic while serving HTTP request",
					"request_id", request.Header.Get("X-Request-ID"),
					"panic_type", fmt.Sprintf("%T", recovered),
					"stack", string(debug.Stack()),
				)
				writeAPIError(
					response,
					http.StatusInternalServerError,
					"internal_error",
					"The request could not be completed.",
					nil,
					logger,
				)
			}
		}()
		next.ServeHTTP(response, request)
	})
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "request-unknown"
	}
	return hex.EncodeToString(value)
}
