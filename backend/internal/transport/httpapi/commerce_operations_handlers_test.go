package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"rigmark/internal/modules/commerce/domain"
	identity "rigmark/internal/modules/identity/domain"
)

type commerceOperationsStub struct{ triggerCalls int }

func (stub *commerceOperationsStub) CreateConfiguration(context.Context, identity.UserID, domain.ProviderConfigurationInput) (domain.ProviderConfiguration, error) {
	return domain.ProviderConfiguration{}, nil
}
func (stub *commerceOperationsStub) ListConfigurations(context.Context) ([]domain.ProviderConfiguration, error) {
	return []domain.ProviderConfiguration{}, nil
}
func (stub *commerceOperationsStub) SetLifecycle(context.Context, identity.UserID, domain.ProviderConfigurationID, domain.ProviderLifecycle) (domain.ProviderConfiguration, error) {
	return domain.ProviderConfiguration{}, nil
}
func (stub *commerceOperationsStub) TriggerManual(context.Context, identity.UserID, domain.ProviderConfigurationID, string) (domain.ImportRun, error) {
	stub.triggerCalls++
	return domain.ImportRun{}, nil
}
func (stub *commerceOperationsStub) Retry(context.Context, identity.UserID, domain.ImportRunID, string) (domain.ImportRun, error) {
	return domain.ImportRun{}, nil
}
func (stub *commerceOperationsStub) ListImports(context.Context, int, int) ([]domain.ImportRun, int64, error) {
	return []domain.ImportRun{}, 0, nil
}
func (stub *commerceOperationsStub) ListFailures(context.Context, domain.ImportRunID, int, int) ([]domain.ImportFailure, int64, error) {
	return []domain.ImportFailure{}, 0, nil
}

func TestManualCommerceImportRequiresAdministrator(t *testing.T) {
	operations := &commerceOperationsStub{}
	authService := &authStub{principal: identity.Principal{UserID: "565a3e84-2c44-433a-a316-a898d5a18bdc", Email: "member@example.invalid"}}
	router := NewRouter(healthStub{}, authService, testCookieConfig,
		slog.New(slog.NewTextHandler(io.Discard, nil)), PublicServices{CommerceOperations: operations})
	request := httptest.NewRequest(http.MethodPost, "/api/admin/commerce/imports",
		strings.NewReader(`{"provider_configuration_id":"80cb0af6-1c46-486c-87fd-32a838ad4f71"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Idempotency-Key", "manual-test-key")
	request.AddCookie(&http.Cookie{Name: testCookieConfig.Name, Value: "member-token"})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || operations.triggerCalls != 0 {
		t.Fatalf("status=%d triggerCalls=%d", response.Code, operations.triggerCalls)
	}
}
