package application

import (
	"context"
	"encoding/json"
	"testing"

	"rigmark/internal/modules/analytics/domain"
)

type recorderStub struct{ event domain.Event }

func (recorder *recorderStub) Record(_ context.Context, event domain.Event) error {
	recorder.event = event
	return nil
}

func TestRecordClientEventValidatesAndNormalizes(t *testing.T) {
	recorder := &recorderStub{}
	service := NewService(recorder)
	userID := "0fdb3972-854d-48f1-bf29-14c80a633b90"
	sessionID := "bbfb4afe-e85b-4537-aef7-f74375be68e3"
	event := domain.Event{
		Name: domain.EventProductSaved, Surface: "catalog", UserID: &userID, SessionID: &sessionID,
		ConsentState: "granted",
		Properties: map[string]json.RawMessage{
			"product_id":  json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`),
			"persistence": json.RawMessage(`"account"`),
		},
	}
	if err := service.RecordClientEvent(context.Background(), event); err != nil {
		t.Fatalf("RecordClientEvent(): %v", err)
	}
	if recorder.event.SchemaVersion != 2 || recorder.event.ConsentState != "granted" || recorder.event.OccurredAt.IsZero() {
		t.Fatalf("event was not normalized: %#v", recorder.event)
	}
}

func TestRecordClientEventRejectsAffiliateClickFromClient(t *testing.T) {
	recorder := &recorderStub{}
	service := NewService(recorder)
	sessionID := "bbfb4afe-e85b-4537-aef7-f74375be68e3"
	event := domain.Event{Name: domain.EventAffiliateClicked, Surface: "wishlist", AnonymousID: &sessionID,
		SessionID: &sessionID, ConsentState: "granted", Properties: map[string]json.RawMessage{}}
	if err := service.RecordClientEvent(context.Background(), event); err != ErrInvalidEvent {
		t.Fatalf("error = %v, want ErrInvalidEvent", err)
	}
}

func TestRecordClientEventRejectsMissingConsent(t *testing.T) {
	recorder := &recorderStub{}
	sessionID := "bbfb4afe-e85b-4537-aef7-f74375be68e3"
	event := domain.Event{
		Name: domain.EventPageView, Surface: "route", AnonymousID: &sessionID,
		SessionID: &sessionID, Properties: map[string]json.RawMessage{},
	}
	if err := NewService(recorder).RecordClientEvent(context.Background(), event); err != ErrInvalidEvent {
		t.Fatalf("error = %v, want ErrInvalidEvent", err)
	}
}

func TestRecordClientEventRejectsExtraProperties(t *testing.T) {
	recorder := &recorderStub{}
	service := NewService(recorder)
	sessionID := "bbfb4afe-e85b-4537-aef7-f74375be68e3"
	event := domain.Event{Name: domain.EventProductViewed, Surface: "product_detail", AnonymousID: &sessionID,
		SessionID: &sessionID, ConsentState: "granted", Properties: map[string]json.RawMessage{
			"product_id": json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`),
			"email":      json.RawMessage(`"person@example.com"`),
		}}
	if err := service.RecordClientEvent(context.Background(), event); err != ErrInvalidEvent {
		t.Fatalf("error = %v, want ErrInvalidEvent", err)
	}
}

func TestRecordClientEventAcceptsOnboardingAndNormalizesAttribution(t *testing.T) {
	recorder := &recorderStub{}
	service := NewService(recorder)
	sessionID := "bbfb4afe-e85b-4537-aef7-f74375be68e3"
	onboardingID := json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`)
	source, referrer := " newsletter ", "EXAMPLE.COM"
	event := domain.Event{
		Name: domain.EventOnboardingStarted, Surface: "recommendation", AnonymousID: &sessionID,
		SessionID: &sessionID, ConsentState: "granted", TrafficSource: &source, ReferrerHost: &referrer,
		Properties: map[string]json.RawMessage{"onboarding_id": onboardingID},
	}
	if err := service.RecordClientEvent(context.Background(), event); err != nil {
		t.Fatalf("RecordClientEvent(): %v", err)
	}
	if recorder.event.TrafficSource == nil || *recorder.event.TrafficSource != "newsletter" ||
		recorder.event.ReferrerHost == nil || *recorder.event.ReferrerHost != "example.com" {
		t.Fatalf("attribution was not normalized: %#v", recorder.event)
	}
}

func TestCanonicalClientEventsAreAccepted(t *testing.T) {
	uuid := json.RawMessage(`"97bfb760-6d09-4b96-8a39-d2bb16445537"`)
	tests := []domain.Event{
		{Name: domain.EventPageView, Properties: map[string]json.RawMessage{}},
		{Name: domain.EventOnboardingStarted, Properties: map[string]json.RawMessage{"onboarding_id": uuid}},
		{Name: domain.EventOnboardingCompleted, Properties: map[string]json.RawMessage{
			"onboarding_id": uuid, "outcome": json.RawMessage(`"complete"`)}},
		{Name: domain.EventRecommendationGenerated, Properties: map[string]json.RawMessage{
			"status": json.RawMessage(`"complete"`), "persistence": json.RawMessage(`"browser"`)}},
		{Name: domain.EventProductViewed, Properties: map[string]json.RawMessage{"product_id": uuid}},
		{Name: domain.EventProductSaved, Properties: map[string]json.RawMessage{
			"product_id": uuid, "persistence": json.RawMessage(`"account"`)}},
		{Name: domain.EventComparisonCreated, Properties: map[string]json.RawMessage{
			"product_count": json.RawMessage(`2`), "persistence": json.RawMessage(`"browser"`)}},
		{Name: domain.EventSetupSaved, Properties: map[string]json.RawMessage{
			"setup_id": uuid, "persistence": json.RawMessage(`"account"`)}},
	}
	for _, event := range tests {
		t.Run(event.Name, func(t *testing.T) {
			recorder := &recorderStub{}
			sessionID := "bbfb4afe-e85b-4537-aef7-f74375be68e3"
			event.Surface = "test"
			event.SessionID = &sessionID
			event.AnonymousID = &sessionID
			event.ConsentState = "granted"
			if err := NewService(recorder).RecordClientEvent(context.Background(), event); err != nil {
				t.Fatalf("RecordClientEvent(): %v", err)
			}
		})
	}
}
