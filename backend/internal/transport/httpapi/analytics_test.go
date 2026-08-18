package httpapi

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	analyticsdomain "rigmark/internal/modules/analytics/domain"
	analyticsports "rigmark/internal/modules/analytics/ports"
	"rigmark/internal/modules/identity/domain"
)

func TestRecordAnalyticsEventFailsClosedWhenStorageUnavailable(t *testing.T) {
	analytics := &analyticsStub{err: errors.New("synthetic database outage")}
	handler := &Handler{analytics: analytics, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	body := `{"event_id":"c3448188-244c-4b2a-9f97-53c1ad10a7ee","name":"page_view","surface":"phase8_validation","session_id":"1191bb26-a9a2-41df-9346-74d693350ce8","consent_version":"analytics-v1","properties":{},"context":{"page_path":"/phase8-validation"}}`
	request := httptest.NewRequest(http.MethodPost, "/api/analytics/events", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: defaultAnalyticsSubjectCookie, Value: base64.RawURLEncoding.EncodeToString(make([]byte, 32))})
	response := httptest.NewRecorder()

	handler.recordAnalyticsEvent(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "synthetic database outage") {
		t.Fatal("internal storage failure leaked to client")
	}
}

type analyticsStub struct {
	AnalyticsService
	event    analyticsdomain.Event
	err      error
	consent  analyticsdomain.Consent
	decision analyticsdomain.ConsentDecision
	claimErr error
}

func (stub *analyticsStub) RecordClientEvent(_ context.Context, event analyticsdomain.Event) (analyticsdomain.IngestionResult, error) {
	stub.event = event
	return analyticsdomain.IngestionResult{Outcome: analyticsdomain.IngestionAccepted}, stub.err
}

func (stub *analyticsStub) SetConsent(_ context.Context, decision analyticsdomain.ConsentDecision) (analyticsdomain.Consent, error) {
	stub.decision = decision
	if stub.err != nil {
		return analyticsdomain.Consent{}, stub.err
	}
	state := decision.RequestedState
	stub.consent = analyticsdomain.Consent{State: state, PolicyVersion: decision.PolicyVersion, Source: decision.Source, DecidedAt: time.Now().UTC()}
	return stub.consent, nil
}

func (stub *analyticsStub) GetConsent(context.Context, analyticsdomain.Subject) (analyticsdomain.Consent, error) {
	if stub.consent.State == "" {
		return analyticsdomain.Consent{}, analyticsports.ErrConsentNotFound
	}
	return stub.consent, nil
}

func (stub *analyticsStub) ClaimIdentity(context.Context, []byte, string) error { return stub.claimErr }

func TestRecordAnalyticsEventAssociatesAnonymousSession(t *testing.T) {
	const sessionID = "1191bb26-a9a2-41df-9346-74d693350ce8"
	analytics := &analyticsStub{}
	handler := &Handler{
		analytics: analytics,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	body := `{"event_id":"c3448188-244c-4b2a-9f97-53c1ad10a7ee","name":"product_viewed","surface":"product_detail","session_id":"` + sessionID + `","consent_version":"analytics-v1","properties":{"product_id":"4ba7d524-9fd5-4d18-8c42-778c42d996f3"},"context":{"page_path":"/products/demo","traffic_source":"newsletter"}}`
	request := httptest.NewRequest(http.MethodPost, "/api/analytics/events", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: defaultAnalyticsSubjectCookie, Value: base64.RawURLEncoding.EncodeToString(make([]byte, 32))})
	response := httptest.NewRecorder()

	handler.recordAnalyticsEvent(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if len(analytics.event.AnonymousSubjectHash) != 32 {
		t.Fatalf("anonymous subject hash length = %d", len(analytics.event.AnonymousSubjectHash))
	}
	if analytics.event.UserID != nil {
		t.Fatalf("unexpected user ID %q", *analytics.event.UserID)
	}
	if analytics.event.PagePath == nil || *analytics.event.PagePath != "/products/demo" {
		t.Fatalf("page path = %v", analytics.event.PagePath)
	}
	if analytics.event.ConsentPolicyVersion != analyticsdomain.CurrentConsentPolicyVersion {
		t.Fatalf("consent policy = %q", analytics.event.ConsentPolicyVersion)
	}
}

func TestAnalyticsConsentIsPersistedForOpaqueBrowserSubject(t *testing.T) {
	analytics := &analyticsStub{}
	handler := &Handler{analytics: analytics, cookie: AuthCookieConfig{AnalyticsSubjectName: "subject", Secure: true}, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	request := httptest.NewRequest(http.MethodPut, "/api/analytics/consent", strings.NewReader(`{"state":"granted","policy_version":"analytics-v1","source":"banner"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.setAnalyticsConsent(response, request)
	if response.Code != http.StatusOK || len(analytics.decision.Subject.AnonymousSubjectHash) != 32 {
		t.Fatalf("status=%d decision=%#v body=%s", response.Code, analytics.decision, response.Body)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode || !cookies[0].Secure {
		t.Fatalf("privacy cookie=%#v", cookies)
	}
	if strings.Contains(response.Body.String(), cookies[0].Value) {
		t.Fatal("opaque subject token leaked in response")
	}
}

func TestAnalyticsIdentityClaimUsesGenericConflict(t *testing.T) {
	analytics := &analyticsStub{claimErr: analyticsports.ErrIdentityClaimConflict}
	handler := &Handler{analytics: analytics, logger: slog.New(slog.NewTextHandler(io.Discard, nil))}
	request := httptest.NewRequest(http.MethodPost, "/api/analytics/identity/claim", nil)
	request.AddCookie(&http.Cookie{Name: defaultAnalyticsSubjectCookie, Value: base64.RawURLEncoding.EncodeToString(make([]byte, 32))})
	request = request.WithContext(context.WithValue(request.Context(), principalContextKey{}, domain.Principal{UserID: "4ba7d524-9fd5-4d18-8c42-778c42d996f3"}))
	response := httptest.NewRecorder()
	handler.claimAnalyticsIdentity(response, request)
	if response.Code != http.StatusConflict || strings.Contains(response.Body.String(), "event") || strings.Contains(response.Body.String(), "user") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func TestRecordAnalyticsEventRejectsUnknownEnvelopeField(t *testing.T) {
	analytics := &analyticsStub{}
	handler := &Handler{
		analytics: analytics,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	body := `{"event_id":"c3448188-244c-4b2a-9f97-53c1ad10a7ee","name":"product_viewed","surface":"product_detail","session_id":"1191bb26-a9a2-41df-9346-74d693350ce8","consent_version":"analytics-v1","properties":{"product_id":"4ba7d524-9fd5-4d18-8c42-778c42d996f3"},"user_id":"spoofed"}`
	request := httptest.NewRequest(http.MethodPost, "/api/analytics/events", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.recordAnalyticsEvent(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if analytics.event.Name != "" {
		t.Fatal("invalid event reached the analytics service")
	}
}
