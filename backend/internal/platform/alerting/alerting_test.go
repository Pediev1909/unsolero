package alerting

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDisabledNotifierFailsClearly(t *testing.T) {
	notifier, err := Select(Config{Provider: "disabled"})
	if err != nil {
		t.Fatal(err)
	}
	if err := notifier.Notify(context.Background(), Alert{Category: BackupFailure}); !errors.Is(err, ErrDeliveryDisabled) {
		t.Fatalf("Notify()=%v", err)
	}
	if err := notifier.Ready(context.Background()); !errors.Is(err, ErrDeliveryDisabled) {
		t.Fatalf("Ready()=%v", err)
	}
}

func TestExternalNotifierFailsClosedWithoutAdapter(t *testing.T) {
	if _, err := Select(Config{Provider: "external"}); err == nil {
		t.Fatal("Select() accepted an unavailable external provider")
	}
}

func TestWebhookDeliversBoundedStructuredAlertWithAuthentication(t *testing.T) {
	token := "12345678901234567890123456789012"
	var received struct {
		Category   Category  `json:"category"`
		Summary    string    `json:"summary"`
		OccurredAt time.Time `json:"occurred_at"`
		Severity   string    `json:"severity"`
		RunID      string    `json:"run_id"`
	}
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodPost || request.Header.Get("Authorization") != "Bearer "+token ||
			request.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("unexpected request: method=%s headers=%v", request.Method, request.Header)
		}
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		return &http.Response{StatusCode: http.StatusAccepted, Body: io.NopCloser(strings.NewReader("accepted")), Header: make(http.Header)}, nil
	})

	notifier, err := Select(Config{Provider: "webhook", Endpoint: "https://alerts.example.test/unsolero", Token: token, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	notifier.(*Webhook).client.Transport = transport
	want := Alert{Category: WorkerRepeatedFailure, Summary: "worker failed repeatedly", OccurredAt: time.Now().UTC(), Severity: "critical", RunID: "test-run"}
	if err := notifier.Notify(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	if received.Category != want.Category || received.Summary != want.Summary || received.Severity != want.Severity || received.RunID != want.RunID {
		t.Fatalf("received=%+v want=%+v", received, want)
	}
}

func TestWebhookFailsClosedOnUnsafeConfigurationAndDeliveryFailure(t *testing.T) {
	for _, test := range []Config{
		{Provider: "webhook", Endpoint: "not-a-url", Token: "12345678901234567890123456789012", Timeout: time.Second},
		{Provider: "webhook", Endpoint: "https://user:password@example.com/alert", Token: "12345678901234567890123456789012", Timeout: time.Second},
		{Provider: "webhook", Endpoint: "https://example.com/alert?secret=value", Token: "12345678901234567890123456789012", Timeout: time.Second},
		{Provider: "webhook", Endpoint: "https://example.com/alert", Token: "short", Timeout: time.Second},
	} {
		if _, err := Select(test); err == nil {
			t.Fatalf("Select() accepted unsafe config: %+v", test)
		}
	}
	notifier, err := Select(Config{Provider: "webhook", Endpoint: "https://alerts.example.test/unsolero", Token: "12345678901234567890123456789012", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	notifier.(*Webhook).client.Transport = roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("do not expose this body")), Header: make(http.Header)}, nil
	})
	err = notifier.Notify(context.Background(), Alert{Category: BackupFailure, Summary: "backup failed", OccurredAt: time.Now()})
	if err == nil || strings.Contains(err.Error(), "do not expose") {
		t.Fatalf("Notify() error=%v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func TestDefaultRulesCoverPhase11OperationalFailuresWithoutDynamicLabels(t *testing.T) {
	rules := DefaultRules()
	if len(rules) < 10 {
		t.Fatalf("rules=%d", len(rules))
	}
	seen := map[Category]bool{}
	for _, rule := range rules {
		if seen[rule.Category] || rule.Metric == "" || rule.For <= 0 || rule.RunbookSection == "" {
			t.Fatalf("invalid or duplicate rule: %+v", rule)
		}
		seen[rule.Category] = true
	}
	for _, category := range []Category{DatabaseUnavailable, ReadinessDegraded, QueueBacklog,
		DeadJobsDetected, BackupTooOld, BackupFailure, MediaDiscrepancy, RedisUnavailable,
		SustainedServerErrors, LatencyBudgetViolation, WorkerUnavailable, CertificateExpiring, DeploymentFailure} {
		if !seen[category] {
			t.Fatalf("missing alert definition %q", category)
		}
	}
}
