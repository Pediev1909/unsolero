package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	newsletter "rigmark/internal/modules/newsletter/application"
)

type newsletterStub struct {
	subscribed     []string
	sources        []string
	receipt        newsletter.Receipt
	subscribeErr   error
	confirmed      []string
	confirmErr     error
	unsubscribed   []string
	unsubscribeErr error
}

func (stub *newsletterStub) Subscribe(_ context.Context, email, source string) (newsletter.Receipt, error) {
	stub.subscribed = append(stub.subscribed, email)
	stub.sources = append(stub.sources, source)
	return stub.receipt, stub.subscribeErr
}

func (stub *newsletterStub) Confirm(_ context.Context, token string) error {
	stub.confirmed = append(stub.confirmed, token)
	return stub.confirmErr
}

func (stub *newsletterStub) Unsubscribe(_ context.Context, token string) error {
	stub.unsubscribed = append(stub.unsubscribed, token)
	return stub.unsubscribeErr
}

// newsletterRoutes builds the same three registrations the router carries so
// the handler is exercised through the mux patterns it will be mounted on.
func newsletterRoutes(stub *newsletterStub) http.Handler {
	handler := newNewsletterHandler(stub, slog.New(slog.NewTextHandler(io.Discard, nil)), nil)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/newsletter/subscriptions", handler.subscribe)
	mux.HandleFunc("POST /api/newsletter/confirmations", handler.confirm)
	mux.HandleFunc("POST /api/newsletter/unsubscriptions", handler.unsubscribe)
	return mux
}

func postNewsletterJSON(t *testing.T, routes http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	routes.ServeHTTP(response, request)
	return response
}

func decodeNewsletterError(t *testing.T, response *httptest.ResponseRecorder) apiError {
	t.Helper()
	var envelope errorEnvelope
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode error envelope: %v (%s)", err, response.Body.String())
	}
	return envelope.Error
}

func TestNewsletterSubscriptionRecordsAndReturnsNeutralReceipt(t *testing.T) {
	stub := &newsletterStub{receipt: newsletter.Receipt{Recorded: true}}
	response := postNewsletterJSON(t, newsletterRoutes(stub), "/api/newsletter/subscriptions",
		`{"email":"reader@example.com","source":"article:mailchimp-alternatives"}`)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want 202: %s", response.Code, response.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["recorded"] != true || len(body) != 1 {
		t.Errorf("body = %v, want only {recorded: true}", body)
	}
	if len(stub.subscribed) != 1 || stub.subscribed[0] != "reader@example.com" || stub.sources[0] != "article:mailchimp-alternatives" {
		t.Errorf("service received %v / %v", stub.subscribed, stub.sources)
	}
}

func TestNewsletterSubscriptionStillAcceptsWhenDeliveryFails(t *testing.T) {
	stub := &newsletterStub{receipt: newsletter.Receipt{Recorded: true}, subscribeErr: errors.New("smtp refused")}
	response := postNewsletterJSON(t, newsletterRoutes(stub), "/api/newsletter/subscriptions",
		`{"email":"reader@example.com","source":"footer"}`)

	if response.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202 when the row was recorded", response.Code)
	}
}

func TestNewsletterSubscriptionFailsClosedWhenNothingWasRecorded(t *testing.T) {
	stub := &newsletterStub{subscribeErr: errors.New("database unavailable")}
	response := postNewsletterJSON(t, newsletterRoutes(stub), "/api/newsletter/subscriptions",
		`{"email":"reader@example.com","source":"footer"}`)

	if response.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500 when the row was not written", response.Code)
	}
	if apiErr := decodeNewsletterError(t, response); apiErr.Code != "newsletter_unavailable" {
		t.Errorf("error code = %q", apiErr.Code)
	}
}

func TestNewsletterSubscriptionRejectsMalformedInput(t *testing.T) {
	stub := &newsletterStub{subscribeErr: newsletter.ValidationError{Fields: map[string]string{"email": "Enter a valid email address."}}}
	routes := newsletterRoutes(stub)

	response := postNewsletterJSON(t, routes, "/api/newsletter/subscriptions", `{"email":"not-an-address","source":"footer"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.Code)
	}
	apiErr := decodeNewsletterError(t, response)
	if apiErr.Code != "validation_failed" || apiErr.Fields["email"] == "" {
		t.Errorf("error = %+v, want validation_failed with an email field", apiErr)
	}

	response = postNewsletterJSON(t, routes, "/api/newsletter/subscriptions", `{"email":"reader@example.com","source":"footer","extra":true}`)
	if response.Code != http.StatusBadRequest || decodeNewsletterError(t, response).Code != "invalid_json" {
		t.Errorf("unknown field status/code = %d/%s, want 400 invalid_json", response.Code, response.Body.String())
	}

	request := httptest.NewRequest(http.MethodPost, "/api/newsletter/subscriptions", strings.NewReader("email=reader"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recorder := httptest.NewRecorder()
	routes.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusUnsupportedMediaType {
		t.Errorf("form body status = %d, want 415", recorder.Code)
	}
	if len(stub.subscribed) != 1 {
		t.Errorf("service calls = %d, want only the well-formed request", len(stub.subscribed))
	}
}

func TestNewsletterConfirmationConsumesTokenOrExplains(t *testing.T) {
	stub := &newsletterStub{}
	routes := newsletterRoutes(stub)

	response := postNewsletterJSON(t, routes, "/api/newsletter/confirmations", `{"token":"raw-token"}`)
	if response.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", response.Code)
	}
	if len(stub.confirmed) != 1 || stub.confirmed[0] != "raw-token" {
		t.Errorf("service received %v", stub.confirmed)
	}

	stub.confirmErr = newsletter.ErrInvalidToken
	response = postNewsletterJSON(t, routes, "/api/newsletter/confirmations", `{"token":"stale"}`)
	if response.Code != http.StatusBadRequest || decodeNewsletterError(t, response).Code != "invalid_token" {
		t.Errorf("invalid token status/body = %d/%s, want 400 invalid_token", response.Code, response.Body.String())
	}

	stub.confirmErr = errors.New("connection reset")
	response = postNewsletterJSON(t, routes, "/api/newsletter/confirmations", `{"token":"raw-token"}`)
	if response.Code != http.StatusInternalServerError || decodeNewsletterError(t, response).Code != "newsletter_unavailable" {
		t.Errorf("infrastructure failure status/body = %d/%s, want 500 newsletter_unavailable", response.Code, response.Body.String())
	}
}

func TestNewsletterUnsubscriptionIsIdempotentAndRejectsUnknownTokens(t *testing.T) {
	stub := &newsletterStub{}
	routes := newsletterRoutes(stub)

	for range 2 {
		response := postNewsletterJSON(t, routes, "/api/newsletter/unsubscriptions", `{"token":"raw-token"}`)
		if response.Code != http.StatusNoContent {
			t.Errorf("status = %d, want 204", response.Code)
		}
	}
	if len(stub.unsubscribed) != 2 {
		t.Errorf("service calls = %d, want 2", len(stub.unsubscribed))
	}

	stub.unsubscribeErr = newsletter.ErrInvalidToken
	response := postNewsletterJSON(t, routes, "/api/newsletter/unsubscriptions", `{"token":"unknown"}`)
	if response.Code != http.StatusBadRequest || decodeNewsletterError(t, response).Code != "invalid_token" {
		t.Errorf("unknown token status/body = %d/%s, want 400 invalid_token", response.Code, response.Body.String())
	}
}
