package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	catalog "rigmark/internal/modules/catalog/domain"
	identity "rigmark/internal/modules/identity/domain"
	planning "rigmark/internal/modules/planning/domain"
	recommendation "rigmark/internal/modules/recommendation/application"
	recommendationdomain "rigmark/internal/modules/recommendation/domain"
	recommendationports "rigmark/internal/modules/recommendation/ports"
)

type recommendationHTTPStub struct {
	generated recommendation.Generated
	listCalls int
}

func (stub *recommendationHTTPStub) Generate(context.Context, *identity.UserID, recommendationdomain.Input) (recommendation.Generated, error) {
	return stub.generated, nil
}
func (*recommendationHTTPStub) GetDraft(context.Context, identity.UserID) (recommendationports.Draft, error) {
	return recommendationports.Draft{}, recommendationports.ErrNotFound
}
func (*recommendationHTTPStub) SaveDraft(_ context.Context, _ identity.UserID, draft recommendationports.Draft) (recommendationports.Draft, error) {
	return draft, nil
}
func (*recommendationHTTPStub) DeleteDraft(context.Context, identity.UserID) error { return nil }
func (stub *recommendationHTTPStub) ListSetups(context.Context, identity.UserID) ([]recommendationports.SetupSummary, error) {
	stub.listCalls++
	return nil, nil
}
func (*recommendationHTTPStub) GetSetup(context.Context, identity.UserID, planning.SetupID) (recommendation.Generated, error) {
	return recommendation.Generated{}, recommendationports.ErrNotFound
}
func (*recommendationHTTPStub) RenameSetup(context.Context, identity.UserID, planning.SetupID, string) error {
	return nil
}
func (*recommendationHTTPStub) DeleteSetup(context.Context, identity.UserID, planning.SetupID) error {
	return nil
}

func newRecommendationTestRouter(authService AuthenticationService, service recommendationService) http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(healthStub{}, authService, testCookieConfig, logger, PublicServices{Recommendations: service})
}

func TestGenerateRecommendationAllowsGuestWithoutPersistedIdentifiers(t *testing.T) {
	service := &recommendationHTTPStub{generated: recommendation.Generated{Result: recommendationdomain.Result{
		Status:    recommendationdomain.ResultNoSuitableProducts,
		TotalCost: catalog.Money{AmountMinor: 0, Currency: "USD"},
	}}}
	request := httptest.NewRequest(http.MethodPost, "/api/recommendations/generate", strings.NewReader(`{
		"goal":"build_muscle","experience":"beginner","budget_minor":70000,"currency":"USD",
		"available_space":{"length_mm":2400,"width_mm":1800,"height_mm":2400,"apartment_living":true},
		"existing_equipment":[],"training_preferences":["dumbbells"],"priorities":["compact"],"free_text":""
	}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	newRecommendationTestRouter(&authStub{}, service).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusOK, response.Body)
	}
	if !strings.Contains(response.Body.String(), `"saved":false`) ||
		!strings.Contains(response.Body.String(), `"recommendation_id":null`) {
		t.Fatalf("unexpected guest response: %s", response.Body)
	}
}

func TestSavedSetupsRequireAuthenticationBeforeApplicationService(t *testing.T) {
	service := &recommendationHTTPStub{}
	request := httptest.NewRequest(http.MethodGet, "/api/account/setups", nil)
	response := httptest.NewRecorder()

	newRecommendationTestRouter(&authStub{}, service).ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized || service.listCalls != 0 {
		t.Fatalf("status = %d, list calls = %d", response.Code, service.listCalls)
	}
}
