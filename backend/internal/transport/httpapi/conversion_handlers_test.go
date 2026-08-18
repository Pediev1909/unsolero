package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	commerce "rigmark/internal/modules/commerce/application"
	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
	identity "rigmark/internal/modules/identity/domain"
)

type conversionOperationsStub struct {
	listCalls    int
	webhookCalls int
	webhookErr   error
}

func (stub *conversionOperationsStub) IngestWebhook(context.Context, domain.ProviderConfigurationID, domain.WebhookRequest) (commerce.WebhookResult, error) {
	stub.webhookCalls++
	return commerce.WebhookResult{}, stub.webhookErr
}
func (*conversionOperationsStub) SetProviderEnabled(context.Context, identity.UserID, domain.ProviderConfigurationID, bool) (domain.ProviderConfiguration, error) {
	return domain.ProviderConfiguration{}, nil
}
func (*conversionOperationsStub) TriggerManualImport(context.Context, identity.UserID, domain.ProviderConfigurationID, string) (domain.ConversionImportRun, error) {
	return domain.ConversionImportRun{}, nil
}
func (*conversionOperationsStub) RetryImport(context.Context, identity.UserID, domain.ConversionImportRunID, string) (domain.ConversionImportRun, error) {
	return domain.ConversionImportRun{}, nil
}
func (stub *conversionOperationsStub) ListConversions(context.Context, domain.ConversionFilter) ([]domain.Conversion, int64, error) {
	stub.listCalls++
	return []domain.Conversion{}, 0, nil
}
func (*conversionOperationsStub) ListImports(context.Context, int, int) ([]domain.ConversionImportRun, int64, error) {
	return []domain.ConversionImportRun{}, 0, nil
}
func (*conversionOperationsStub) ListReconciliations(context.Context, int, int) ([]domain.ReconciliationRun, int64, error) {
	return []domain.ReconciliationRun{}, 0, nil
}
func (*conversionOperationsStub) Reconcile(context.Context, identity.UserID, domain.ConversionImportRunID, string) (domain.ReconciliationRun, error) {
	return domain.ReconciliationRun{}, nil
}
func (*conversionOperationsStub) Metrics(context.Context, time.Time, time.Time) (domain.MonetizationReport, error) {
	return domain.MonetizationReport{}, nil
}

func TestConversionDiagnosticsRequireAdministrator(t *testing.T) {
	operations := &conversionOperationsStub{}
	authService := &authStub{principal: identity.Principal{UserID: "565a3e84-2c44-433a-a316-a898d5a18bdc", Email: "member@example.invalid"}}
	router := NewRouter(healthStub{}, authService, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{ConversionOperations: operations})
	request := httptest.NewRequest(http.MethodGet, "/api/admin/commerce/conversions", nil)
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "member-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || operations.listCalls != 0 {
		t.Fatalf("status=%d listCalls=%d", response.Code, operations.listCalls)
	}
}

func TestConversionWebhookEnforcesBodyLimitBeforeService(t *testing.T) {
	operations := &conversionOperationsStub{}
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{ConversionOperations: operations})
	body := strings.NewReader(strings.Repeat("x", maximumConversionWebhookBytes+1))
	request := httptest.NewRequest(http.MethodPost,
		"/api/webhooks/commerce/80cb0af6-1c46-486c-87fd-32a838ad4f71", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge || operations.webhookCalls != 0 {
		t.Fatalf("status=%d webhookCalls=%d", response.Code, operations.webhookCalls)
	}
}

func TestRejectedConversionWebhookDoesNotLeakProviderError(t *testing.T) {
	operations := &conversionOperationsStub{webhookErr: ports.ErrWebhookRejected}
	router := NewRouter(healthStub{}, &authStub{}, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{ConversionOperations: operations})
	request := httptest.NewRequest(http.MethodPost,
		"/api/webhooks/commerce/80cb0af6-1c46-486c-87fd-32a838ad4f71", strings.NewReader(`{"event":"forged"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "secret") ||
		strings.Contains(response.Body.String(), "signature") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}
