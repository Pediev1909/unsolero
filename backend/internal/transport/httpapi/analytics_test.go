package httpapi

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	analyticsdomain "rigmark/internal/modules/analytics/domain"
)

type analyticsStub struct {
	event analyticsdomain.Event
	err   error
}

func (stub *analyticsStub) RecordClientEvent(_ context.Context, event analyticsdomain.Event) error {
	stub.event = event
	return stub.err
}

func TestRecordAnalyticsEventAssociatesAnonymousSession(t *testing.T) {
	const sessionID = "1191bb26-a9a2-41df-9346-74d693350ce8"
	analytics := &analyticsStub{}
	handler := &Handler{
		analytics: analytics,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	body := `{"name":"product_viewed","surface":"product_detail","session_id":"` + sessionID + `","consent_state":"granted","properties":{"product_id":"4ba7d524-9fd5-4d18-8c42-778c42d996f3"},"context":{"page_path":"/products/demo","traffic_source":"newsletter"}}`
	request := httptest.NewRequest(http.MethodPost, "/api/analytics/events", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handler.recordAnalyticsEvent(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if analytics.event.AnonymousID == nil || *analytics.event.AnonymousID != sessionID {
		t.Fatalf("anonymous ID = %v, want %q", analytics.event.AnonymousID, sessionID)
	}
	if analytics.event.UserID != nil {
		t.Fatalf("unexpected user ID %q", *analytics.event.UserID)
	}
	if analytics.event.PagePath == nil || *analytics.event.PagePath != "/products/demo" {
		t.Fatalf("page path = %v", analytics.event.PagePath)
	}
	if analytics.event.ConsentState != "granted" {
		t.Fatalf("consent state = %q, want granted", analytics.event.ConsentState)
	}
}

func TestRecordAnalyticsEventRejectsUnknownEnvelopeField(t *testing.T) {
	analytics := &analyticsStub{}
	handler := &Handler{
		analytics: analytics,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	body := `{"name":"product_viewed","surface":"product_detail","session_id":"1191bb26-a9a2-41df-9346-74d693350ce8","consent_state":"granted","properties":{"product_id":"4ba7d524-9fd5-4d18-8c42-778c42d996f3"},"user_id":"spoofed"}`
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
