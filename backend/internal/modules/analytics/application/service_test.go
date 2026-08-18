package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

const testEventID = "c3448188-244c-4b2a-9f97-53c1ad10a7ee"
const testSessionID = "bbfb4afe-e85b-4537-aef7-f74375be68e3"

type recorderStub struct {
	ports.EventRecorder
	event    domain.Event
	result   domain.IngestionResult
	rejected domain.IngestionOutcome
}

func (recorder *recorderStub) Ingest(_ context.Context, event domain.Event, _ time.Duration) (domain.IngestionResult, error) {
	recorder.event = event
	if recorder.result.Outcome == "" {
		return domain.IngestionResult{Outcome: domain.IngestionAccepted}, nil
	}
	return recorder.result, nil
}

func (recorder *recorderStub) RecordRejected(_ context.Context, _ domain.EventID, _ string, outcome domain.IngestionOutcome, _ string, _ time.Duration) error {
	recorder.rejected = outcome
	return nil
}

func validAnonymousEvent(name string, properties map[string]json.RawMessage) domain.Event {
	return domain.Event{ID: testEventID, Name: name, Surface: "test", SessionID: stringPointer(testSessionID),
		AnonymousSubjectHash: make([]byte, 32), ConsentPolicyVersion: domain.CurrentConsentPolicyVersion,
		Classification: "human", Properties: properties}
}

func TestRecordClientEventValidatesAndNormalizes(t *testing.T) {
	recorder := &recorderStub{}
	userID := "0fdb3972-854d-48f1-bf29-14c80a633b90"
	event := validAnonymousEvent(domain.EventProductSaved, map[string]json.RawMessage{
		"product_id":  json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`),
		"persistence": json.RawMessage(`"account"`),
	})
	event.AnonymousSubjectHash = nil
	event.UserID = &userID
	result, err := NewService(recorder).RecordClientEvent(context.Background(), event)
	if err != nil || result.Outcome != domain.IngestionAccepted {
		t.Fatalf("RecordClientEvent() = %#v, %v", result, err)
	}
	if recorder.event.SchemaVersion != 3 || recorder.event.ConsentState != "granted" || !recorder.event.Reportable || recorder.event.Origin != "client" || recorder.event.OccurredAt.IsZero() {
		t.Fatalf("event was not normalized: %#v", recorder.event)
	}
}

func TestRecordClientEventRejectsAffiliateClickFromClient(t *testing.T) {
	recorder := &recorderStub{}
	_, err := NewService(recorder).RecordClientEvent(context.Background(), validAnonymousEvent(domain.EventAffiliateClicked, map[string]json.RawMessage{}))
	if !errors.Is(err, ErrInvalidEvent) || recorder.rejected != domain.IngestionRejected {
		t.Fatalf("error/outcome = %v/%s", err, recorder.rejected)
	}
}

func TestRecordClientEventRequiresPersistedConsent(t *testing.T) {
	recorder := &recorderStub{result: domain.IngestionResult{Outcome: domain.IngestionRejected}}
	_, err := NewService(recorder).RecordClientEvent(context.Background(), validAnonymousEvent(domain.EventPageView, map[string]json.RawMessage{}))
	if !errors.Is(err, ErrConsentRequired) {
		t.Fatalf("error = %v, want ErrConsentRequired", err)
	}
}

func TestRecordClientEventPrivacyFiltersForbiddenProperties(t *testing.T) {
	for _, forbidden := range []string{"email", "password", "authorization", "free_text", "affiliate_url", "order_details", "user_agent", "ip_address"} {
		t.Run(forbidden, func(t *testing.T) {
			recorder := &recorderStub{}
			event := validAnonymousEvent(domain.EventProductViewed, map[string]json.RawMessage{
				"product_id": json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`),
				forbidden:    json.RawMessage(`"must-not-persist"`),
			})
			_, err := NewService(recorder).RecordClientEvent(context.Background(), event)
			if !errors.Is(err, ErrInvalidEvent) || recorder.rejected != domain.IngestionPrivacyFiltered {
				t.Fatalf("error/outcome = %v/%s", err, recorder.rejected)
			}
		})
	}
}

func TestRecordClientEventRejectsStaleConsentVersion(t *testing.T) {
	recorder := &recorderStub{}
	event := validAnonymousEvent(domain.EventPageView, map[string]json.RawMessage{})
	event.ConsentPolicyVersion = "analytics-v0"
	_, err := NewService(recorder).RecordClientEvent(context.Background(), event)
	if !errors.Is(err, ErrInvalidEvent) || recorder.rejected != domain.IngestionRejected {
		t.Fatalf("error/outcome = %v/%s", err, recorder.rejected)
	}
}

func TestRecordClientEventFiltersBotsWithoutStoringPayload(t *testing.T) {
	recorder := &recorderStub{}
	event := validAnonymousEvent(domain.EventPageView, map[string]json.RawMessage{})
	event.Classification = "prefetch"
	result, err := NewService(recorder).RecordClientEvent(context.Background(), event)
	if err != nil || result.Outcome != domain.IngestionBotFiltered || recorder.event.Name != "" {
		t.Fatalf("result=%#v err=%v stored=%#v", result, err, recorder.event)
	}
}

func TestRecordClientEventAcceptsOnboardingAndNormalizesAttribution(t *testing.T) {
	recorder := &recorderStub{}
	source, referrer := " newsletter ", "EXAMPLE.COM"
	event := validAnonymousEvent(domain.EventOnboardingStarted, map[string]json.RawMessage{
		"onboarding_id": json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`),
	})
	event.TrafficSource, event.ReferrerHost = &source, &referrer
	if _, err := NewService(recorder).RecordClientEvent(context.Background(), event); err != nil {
		t.Fatal(err)
	}
	if recorder.event.TrafficSource == nil || *recorder.event.TrafficSource != "newsletter" || recorder.event.ReferrerHost == nil || *recorder.event.ReferrerHost != "example.com" {
		t.Fatalf("not normalized: %#v", recorder.event)
	}
}

func TestCanonicalClientEventsAreAccepted(t *testing.T) {
	uuid := json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`)
	tests := []domain.Event{
		validAnonymousEvent(domain.EventPageView, map[string]json.RawMessage{}),
		validAnonymousEvent(domain.EventOnboardingStarted, map[string]json.RawMessage{"onboarding_id": uuid}),
		validAnonymousEvent(domain.EventOnboardingCompleted, map[string]json.RawMessage{"onboarding_id": uuid, "outcome": json.RawMessage(`"complete"`)}),
		validAnonymousEvent(domain.EventRecommendationGenerated, map[string]json.RawMessage{"status": json.RawMessage(`"complete"`), "persistence": json.RawMessage(`"browser"`)}),
		validAnonymousEvent(domain.EventProductViewed, map[string]json.RawMessage{"product_id": uuid}),
		validAnonymousEvent(domain.EventProductSaved, map[string]json.RawMessage{"product_id": uuid, "persistence": json.RawMessage(`"account"`)}),
		validAnonymousEvent(domain.EventComparisonCreated, map[string]json.RawMessage{"product_count": json.RawMessage(`2`), "persistence": json.RawMessage(`"browser"`)}),
		validAnonymousEvent(domain.EventSetupSaved, map[string]json.RawMessage{"setup_id": uuid, "persistence": json.RawMessage(`"account"`)}),
	}
	for _, event := range tests {
		t.Run(event.Name, func(t *testing.T) {
			if _, err := NewService(&recorderStub{}).RecordClientEvent(context.Background(), event); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
