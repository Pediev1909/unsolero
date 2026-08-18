package alerting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var ErrDeliveryDisabled = errors.New("operational alert delivery is disabled")

type Category string

const (
	DatabaseUnavailable     Category = "database_unavailable"
	MigrationFailure        Category = "migration_failure"
	WorkerRepeatedFailure   Category = "worker_repeated_failure"
	QueueBacklog            Category = "queue_backlog"
	ProviderDegradation     Category = "provider_degradation"
	WebhookFailureSpike     Category = "webhook_failure_spike"
	ReconciliationBacklog   Category = "reconciliation_backlog"
	BackupFailure           Category = "backup_failure"
	RetentionCleanupFailure Category = "retention_cleanup_failure"
	AuthenticationAnomaly   Category = "authentication_security_anomaly"
	RateLimitBackendFailure Category = "rate_limit_backend_failure"
	ReadinessDegraded       Category = "readiness_degraded"
	DeadJobsDetected        Category = "dead_jobs_detected"
	BackupTooOld            Category = "backup_too_old"
	MediaDiscrepancy        Category = "media_reconciliation_discrepancy"
	RedisUnavailable        Category = "redis_unavailable"
	SustainedServerErrors   Category = "sustained_server_errors"
	LatencyBudgetViolation  Category = "latency_budget_violation"
	WorkerUnavailable       Category = "worker_unavailable"
	CertificateExpiring     Category = "certificate_expiring"
	DeploymentFailure       Category = "deployment_failure"
)

type Rule struct {
	Category       Category
	Metric         string
	Comparison     string
	Threshold      float64
	For            time.Duration
	Severity       string
	RunbookSection string
}

// DefaultRules is provider-neutral policy. A staging or production telemetry
// system must translate and install these rules; defining them does not imply
// that an external alert destination is active.
func DefaultRules() []Rule {
	return []Rule{
		{DatabaseUnavailable, "unsolero_readiness_failure_total", ">", 0, time.Minute, "critical", "database-unavailable"},
		{ReadinessDegraded, "unsolero_readiness_failure_total", ">", 0, 5 * time.Minute, "warning", "readiness-degraded"},
		{QueueBacklog, "unsolero_worker_backlog_depth", ">", 100, 10 * time.Minute, "warning", "worker-backlog"},
		{DeadJobsDetected, "unsolero_worker_dead_jobs", ">", 0, time.Minute, "critical", "dead-jobs"},
		{BackupTooOld, "unsolero_backup_age_seconds", ">", 90000, 5 * time.Minute, "critical", "backup-too-old"},
		{BackupFailure, "unsolero_backup_failure_count", ">", 0, time.Minute, "critical", "backup-failure"},
		{MediaDiscrepancy, "unsolero_media_reconciliation_discrepancies", ">", 0, 5 * time.Minute, "warning", "media-reconciliation"},
		{RedisUnavailable, "unsolero_redis_unavailable_total", ">", 0, time.Minute, "critical", "redis-unavailable"},
		{SustainedServerErrors, "unsolero_http_errors_total", ">", 5, 5 * time.Minute, "critical", "sustained-5xx"},
		{LatencyBudgetViolation, "unsolero_http_duration_milliseconds", ">", 1000, 10 * time.Minute, "warning", "latency-budget"},
		{WorkerUnavailable, "unsolero_worker_heartbeat_age_seconds", ">", 120, 2 * time.Minute, "critical", "worker-unavailable"},
		{CertificateExpiring, "unsolero_tls_certificate_expiry_seconds", "<", 1209600, time.Minute, "warning", "certificate-expiring"},
		{DeploymentFailure, "unsolero_deployment_failure", ">", 0, time.Minute, "critical", "deployment-failure"},
	}
}

type Alert struct {
	Category   Category
	Summary    string
	OccurredAt time.Time
	Severity   string
	RunID      string
}

type Notifier interface {
	Notify(context.Context, Alert) error
	Ready(context.Context) error
}

type Disabled struct{}

func (Disabled) Notify(context.Context, Alert) error { return ErrDeliveryDisabled }
func (Disabled) Ready(context.Context) error         { return ErrDeliveryDisabled }

type Config struct {
	Provider string
	Endpoint string
	Token    string
	Timeout  time.Duration
}

type Webhook struct {
	endpoint string
	token    string
	client   *http.Client
}

func NewWebhook(endpoint, token string, timeout time.Duration) (*Webhook, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("alert webhook endpoint must be an absolute HTTP(S) URL")
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, errors.New("alert webhook endpoint must not contain credentials, query parameters, or fragments")
	}
	if len(token) < 32 {
		return nil, errors.New("alert webhook token must contain at least 32 characters")
	}
	if timeout < time.Second || timeout > 30*time.Second {
		return nil, errors.New("alert webhook timeout must be between 1s and 30s")
	}
	return &Webhook{endpoint: parsed.String(), token: token, client: &http.Client{Timeout: timeout}}, nil
}

func (webhook *Webhook) Notify(ctx context.Context, alert Alert) error {
	if alert.Category == "" || strings.TrimSpace(alert.Summary) == "" || alert.OccurredAt.IsZero() {
		return errors.New("operational alert is incomplete")
	}
	body, err := json.Marshal(struct {
		Category   Category  `json:"category"`
		Summary    string    `json:"summary"`
		OccurredAt time.Time `json:"occurred_at"`
		Severity   string    `json:"severity,omitempty"`
		RunID      string    `json:"run_id,omitempty"`
	}{alert.Category, alert.Summary, alert.OccurredAt.UTC(), alert.Severity, alert.RunID})
	if err != nil {
		return fmt.Errorf("encode operational alert: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.endpoint, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("create operational alert request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+webhook.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "unsolero-alerting/1")
	response, err := webhook.client.Do(request)
	if err != nil {
		return fmt.Errorf("deliver operational alert: %w", err)
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("deliver operational alert: endpoint returned HTTP %d", response.StatusCode)
	}
	return nil
}

// Ready confirms that the adapter was constructed with valid configuration.
// Actual alert-delivery evidence must still be obtained with a controlled test
// alert in the deployed environment.
func (webhook *Webhook) Ready(context.Context) error { return nil }

func Select(config Config) (Notifier, error) {
	switch config.Provider {
	case "disabled":
		return Disabled{}, nil
	case "webhook":
		return NewWebhook(config.Endpoint, config.Token, config.Timeout)
	case "external":
		return nil, errors.New("ALERT_PROVIDER=external requires a reviewed adapter that is not linked in this repository")
	default:
		return nil, fmt.Errorf("unsupported alert provider %q", config.Provider)
	}
}
