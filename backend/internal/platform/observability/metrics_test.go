package observability

import (
	"testing"
	"time"
)

func TestMemoryRecorderUsesRoutePatternsAndAggregatesErrors(t *testing.T) {
	recorder := NewMemoryRecorder()
	recorder.ObserveHTTP("GET", "GET /api/catalog/products/{slug}", 200, 12*time.Millisecond)
	recorder.ObserveHTTP("GET", "GET /api/catalog/products/{slug}", 503, 8*time.Millisecond)
	recorder.Increment(MetricRateLimited)
	recorder.Increment("user-supplied-label")
	recorder.SetGauge("worker_backlog_depth", 7)
	recorder.SetGauge("user-supplied-gauge", 99)
	snapshot := recorder.Snapshot()
	if len(snapshot.HTTP) != 2 || snapshot.Counters[MetricRateLimited] != 1 {
		t.Fatalf("snapshot=%#v", snapshot)
	}
	if _, exists := snapshot.Counters["user-supplied-label"]; exists {
		t.Fatal("unexpected high-cardinality counter was retained")
	}
	if snapshot.Gauges["worker_backlog_depth"] != 7 {
		t.Fatalf("bounded gauge missing: %#v", snapshot.Gauges)
	}
	if _, exists := snapshot.Gauges["user-supplied-gauge"]; exists {
		t.Fatal("unexpected high-cardinality gauge was retained")
	}
	for _, metric := range snapshot.HTTP {
		if metric.Route != "GET /api/catalog/products/{slug}" || metric.Requests != 1 {
			t.Fatalf("metric=%#v", metric)
		}
		if metric.DurationBuckets[len(metric.DurationBuckets)-1] != 1 {
			t.Fatalf("histogram count=%#v", metric.DurationBuckets)
		}
	}
}
