package observability

import (
	"sort"
	"strconv"
	"sync"
	"time"
)

const (
	MetricRateLimited             = "rate_limit_rejected"
	MetricRateLimitBackendFailure = "rate_limit_backend_failure"
	MetricAnalyticsAccepted       = "analytics_accepted"
	MetricAnalyticsRejected       = "analytics_rejected"
	MetricRecommendationFailure   = "recommendation_failure"
	MetricAuthenticationFailure   = "authentication_failure"
	MetricWebhookFailure          = "webhook_failure"
	MetricReadinessFailure        = "readiness_failure"
	MetricStorageFailure          = "storage_failure"
	MetricProviderFailure         = "provider_failure"
	MetricReconciliationFailure   = "reconciliation_failure"
	MetricEmailDeliveryFailure    = "email_delivery_failure"
	MetricWorkerFailure           = "worker_failure"
	MetricJobRetry                = "job_retry"
	MetricDatabaseCancellation    = "database_query_cancellation"
	MetricBackupFailure           = "backup_failure"
	MetricDatabaseAcquireFailure  = "database_pool_acquisition_failure"
	MetricDatabaseTransactionFail = "database_transaction_failure"
	MetricMigrationFailure        = "migration_readiness_failure"
	MetricMediaDiscrepancy        = "media_reconciliation_discrepancy"
	MetricRedisUnavailable        = "redis_unavailable"
)

var allowedCounters = map[string]struct{}{
	MetricRateLimited:             {},
	MetricRateLimitBackendFailure: {},
	MetricAnalyticsAccepted:       {},
	MetricAnalyticsRejected:       {},
	MetricRecommendationFailure:   {},
	MetricAuthenticationFailure:   {},
	MetricWebhookFailure:          {},
	MetricReadinessFailure:        {},
	MetricStorageFailure:          {},
	MetricProviderFailure:         {},
	MetricReconciliationFailure:   {},
	MetricEmailDeliveryFailure:    {},
	MetricWorkerFailure:           {},
	MetricJobRetry:                {},
	MetricDatabaseCancellation:    {},
	MetricBackupFailure:           {},
	MetricDatabaseAcquireFailure:  {},
	MetricDatabaseTransactionFail: {},
	MetricMigrationFailure:        {},
	MetricMediaDiscrepancy:        {},
	MetricRedisUnavailable:        {},
}

var allowedGauges = map[string]struct{}{
	"database_pool_acquired": {}, "database_pool_idle": {}, "database_pool_total": {},
	"database_pool_max": {}, "database_pool_wait_count": {}, "database_pool_wait_seconds_total": {},
	"database_pool_canceled_acquire_count": {}, "worker_active_jobs": {}, "worker_backlog_depth": {},
	"worker_successful_jobs": {}, "worker_failed_jobs": {}, "worker_retry_count": {},
	"worker_dead_jobs": {}, "worker_lease_recovery_count": {}, "worker_processing_latency_seconds": {},
	"worker_last_success_timestamp": {}, "worker_heartbeat_age_seconds": {}, "worker_heartbeat_failure_count": {},
	"backup_last_success_timestamp": {}, "backup_age_seconds": {}, "backup_failure_count": {},
	"backup_restore_verified": {}, "backup_migration_fingerprint_mismatch": {},
	"media_pending_deletions": {}, "media_deletion_retry_count": {}, "media_dead_deletion_jobs": {},
	"media_reconciliation_discrepancies": {}, "media_storage_failure_count": {},
	"redis_limiter_latency_milliseconds": {},
}

var httpDurationBucketsMS = [...]uint64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000}

func HTTPDurationBucketsMS() []uint64 {
	result := make([]uint64, len(httpDurationBucketsMS))
	copy(result, httpDurationBucketsMS[:])
	return result
}

type Recorder interface {
	ObserveHTTP(method, route string, status int, duration time.Duration)
	Increment(name string)
	SetGauge(name string, value float64)
	Snapshot() Snapshot
}

type HTTPMetric struct {
	Method          string                                 `json:"method"`
	Route           string                                 `json:"route"`
	StatusClass     string                                 `json:"status_class"`
	Requests        uint64                                 `json:"requests"`
	Errors          uint64                                 `json:"errors"`
	DurationSumMS   uint64                                 `json:"duration_sum_ms"`
	DurationMaxMS   uint64                                 `json:"duration_max_ms"`
	DurationBuckets [len(httpDurationBucketsMS) + 1]uint64 `json:"duration_buckets"`
}

type Snapshot struct {
	GeneratedAt time.Time          `json:"generated_at"`
	HTTP        []HTTPMetric       `json:"http"`
	Counters    map[string]uint64  `json:"counters"`
	Gauges      map[string]float64 `json:"gauges"`
}

type MemoryRecorder struct {
	mu       sync.RWMutex
	http     map[string]HTTPMetric
	counters map[string]uint64
	gauges   map[string]float64
	now      func() time.Time
}

func NewMemoryRecorder() *MemoryRecorder {
	return &MemoryRecorder{http: map[string]HTTPMetric{}, counters: map[string]uint64{}, gauges: map[string]float64{}, now: time.Now}
}

func (recorder *MemoryRecorder) ObserveHTTP(method, route string, status int, duration time.Duration) {
	if route == "" {
		route = "unmatched"
	}
	statusClass := strconv.Itoa(status/100) + "xx"
	key := method + "\x00" + route + "\x00" + statusClass
	durationMS := uint64(max(0, duration.Milliseconds()))
	recorder.mu.Lock()
	metric := recorder.http[key]
	metric.Method, metric.Route, metric.StatusClass = method, route, statusClass
	metric.Requests++
	if status >= 500 {
		metric.Errors++
	}
	metric.DurationSumMS += durationMS
	if durationMS > metric.DurationMaxMS {
		metric.DurationMaxMS = durationMS
	}
	for index, bound := range httpDurationBucketsMS {
		if durationMS <= bound {
			metric.DurationBuckets[index]++
		}
	}
	metric.DurationBuckets[len(httpDurationBucketsMS)]++
	recorder.http[key] = metric
	recorder.mu.Unlock()
}

func (recorder *MemoryRecorder) SetGauge(name string, value float64) {
	if _, allowed := allowedGauges[name]; !allowed {
		return
	}
	recorder.mu.Lock()
	recorder.gauges[name] = value
	recorder.mu.Unlock()
}

func (recorder *MemoryRecorder) Increment(name string) {
	if _, allowed := allowedCounters[name]; !allowed {
		return
	}
	recorder.mu.Lock()
	recorder.counters[name]++
	recorder.mu.Unlock()
}

func (recorder *MemoryRecorder) Snapshot() Snapshot {
	recorder.mu.RLock()
	result := Snapshot{GeneratedAt: recorder.now().UTC(), HTTP: make([]HTTPMetric, 0, len(recorder.http)),
		Counters: make(map[string]uint64, len(recorder.counters)), Gauges: make(map[string]float64, len(recorder.gauges))}
	for _, metric := range recorder.http {
		result.HTTP = append(result.HTTP, metric)
	}
	for name, value := range recorder.counters {
		result.Counters[name] = value
	}
	for name, value := range recorder.gauges {
		result.Gauges[name] = value
	}
	recorder.mu.RUnlock()
	sort.Slice(result.HTTP, func(i, j int) bool {
		if result.HTTP[i].Route != result.HTTP[j].Route {
			return result.HTTP[i].Route < result.HTTP[j].Route
		}
		if result.HTTP[i].Method != result.HTTP[j].Method {
			return result.HTTP[i].Method < result.HTTP[j].Method
		}
		return result.HTTP[i].StatusClass < result.HTTP[j].StatusClass
	})
	return result
}

type DisabledRecorder struct{}

func (DisabledRecorder) ObserveHTTP(string, string, int, time.Duration) {}
func (DisabledRecorder) Increment(string)                               {}
func (DisabledRecorder) SetGauge(string, float64)                       {}
func (DisabledRecorder) Snapshot() Snapshot                             { return Snapshot{} }
