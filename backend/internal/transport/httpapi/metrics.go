package httpapi

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"rigmark/internal/platform/observability"
)

func (h *Handler) operationalMetrics(response http.ResponseWriter, request *http.Request) {
	if !h.monitoringAuthorized(response, request) {
		return
	}
	h.refreshOperationalMetrics(request)
	response.Header().Set("Cache-Control", "no-store")
	writeJSON(response, http.StatusOK, h.metrics.Snapshot(), h.logger)
}

func (h *Handler) openMetrics(response http.ResponseWriter, request *http.Request) {
	if !h.monitoringAuthorized(response, request) {
		return
	}
	h.refreshOperationalMetrics(request)
	snapshot := h.metrics.Snapshot()
	response.Header().Set("Cache-Control", "no-store")
	response.Header().Set("Content-Type", "application/openmetrics-text; version=1.0.0; charset=utf-8")
	response.WriteHeader(http.StatusOK)
	// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- authenticated OpenMetrics text with escaped label values, never HTML.
	_, _ = response.Write([]byte(formatOpenMetrics(snapshot)))
}

func (h *Handler) monitoringAuthorized(response http.ResponseWriter, request *http.Request) bool {
	provided := strings.TrimSpace(strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer "))
	expectedDigest := sha256.Sum256([]byte(h.metricsToken))
	providedDigest := sha256.Sum256([]byte(provided))
	if provided == "" || subtle.ConstantTimeCompare(providedDigest[:], expectedDigest[:]) != 1 {
		writeAPIError(response, http.StatusUnauthorized, "metrics_authentication_required", "Valid monitoring credentials are required.", nil, h.logger)
		return false
	}
	return true
}

func formatOpenMetrics(snapshot observability.Snapshot) string {
	var output strings.Builder
	output.WriteString("# TYPE unsolero_http_requests_total counter\n")
	output.WriteString("# TYPE unsolero_http_errors_total counter\n")
	output.WriteString("# TYPE unsolero_http_duration_milliseconds_sum counter\n")
	output.WriteString("# TYPE unsolero_http_duration_milliseconds_max gauge\n")
	output.WriteString("# TYPE unsolero_http_duration_milliseconds histogram\n")
	for _, metric := range snapshot.HTTP {
		labels := `method="` + escapeMetricLabel(metric.Method) + `",route="` + escapeMetricLabel(metric.Route) + `",status_class="` + escapeMetricLabel(metric.StatusClass) + `"`
		output.WriteString("unsolero_http_requests_total{" + labels + "} " + strconv.FormatUint(metric.Requests, 10) + "\n")
		output.WriteString("unsolero_http_errors_total{" + labels + "} " + strconv.FormatUint(metric.Errors, 10) + "\n")
		output.WriteString("unsolero_http_duration_milliseconds_sum{" + labels + "} " + strconv.FormatUint(metric.DurationSumMS, 10) + "\n")
		output.WriteString("unsolero_http_duration_milliseconds_max{" + labels + "} " + strconv.FormatUint(metric.DurationMaxMS, 10) + "\n")
		for index, bound := range observability.HTTPDurationBucketsMS() {
			output.WriteString("unsolero_http_duration_milliseconds_bucket{" + labels + `,le="` + strconv.FormatUint(bound, 10) + `"} ` + strconv.FormatUint(metric.DurationBuckets[index], 10) + "\n")
		}
		output.WriteString("unsolero_http_duration_milliseconds_bucket{" + labels + `,le="+Inf"} ` + strconv.FormatUint(metric.DurationBuckets[len(metric.DurationBuckets)-1], 10) + "\n")
		output.WriteString("unsolero_http_duration_milliseconds_count{" + labels + "} " + strconv.FormatUint(metric.Requests, 10) + "\n")
	}
	names := make([]string, 0, len(snapshot.Counters))
	for name := range snapshot.Counters {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		output.WriteString("# TYPE unsolero_" + name + "_total counter\n")
		output.WriteString("unsolero_" + name + "_total " + strconv.FormatUint(snapshot.Counters[name], 10) + "\n")
	}
	gaugeNames := make([]string, 0, len(snapshot.Gauges))
	for name := range snapshot.Gauges {
		gaugeNames = append(gaugeNames, name)
	}
	sort.Strings(gaugeNames)
	for _, name := range gaugeNames {
		output.WriteString("# TYPE unsolero_" + name + " gauge\n")
		output.WriteString("unsolero_" + name + " " + strconv.FormatFloat(snapshot.Gauges[name], 'f', -1, 64) + "\n")
	}
	output.WriteString("# EOF\n")
	return output.String()
}

func (h *Handler) refreshOperationalMetrics(request *http.Request) {
	if h.operationalSource == nil {
		return
	}
	gauges, err := h.operationalSource.Collect(request.Context())
	for name, value := range gauges {
		h.metrics.SetGauge(name, value)
	}
	if err != nil {
		h.metrics.Increment(observability.MetricDatabaseAcquireFailure)
		return
	}
}

func escapeMetricLabel(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	return strings.ReplaceAll(value, `"`, `\"`)
}
