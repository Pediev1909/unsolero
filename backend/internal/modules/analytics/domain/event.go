package domain

import (
	"encoding/json"
	"time"
)

type EventID string

const CurrentConsentPolicyVersion = "analytics-v1"

const (
	EventPageView                = "page_view"
	EventOnboardingStarted       = "onboarding_started"
	EventOnboardingCompleted     = "onboarding_completed"
	EventRecommendationGenerated = "recommendation_generated"
	EventProductViewed           = "product_viewed"
	EventProductSaved            = "product_saved"
	EventComparisonCreated       = "comparison_created"
	EventSetupSaved              = "setup_saved"
	EventAffiliateClicked        = "affiliate_clicked"
)

type Event struct {
	ID                      EventID
	Name                    string
	SchemaVersion           int16
	UserID                  *string
	RecommendationSessionID *string
	AnonymousID             *string
	AnonymousSubjectHash    []byte
	SessionID               *string
	RequestID               *string
	Surface                 string
	Properties              map[string]json.RawMessage
	PagePath                *string
	TrafficSource           *string
	TrafficMedium           *string
	Campaign                *string
	ReferrerHost            *string
	ConsentState            string
	ConsentPolicyVersion    string
	Origin                  string
	Classification          string
	Reportable              bool
	RetentionExpiresAt      time.Time
	OccurredAt              time.Time
	ReceivedAt              time.Time
}

type Subject struct {
	UserID               *string
	AnonymousSubjectHash []byte
}

type ConsentDecision struct {
	Subject        Subject
	RequestedState string
	PolicyVersion  string
	Source         string
	DecidedAt      time.Time
}

type Consent struct {
	State         string
	PolicyVersion string
	Source        string
	DecidedAt     time.Time
}

type IngestionOutcome string

const (
	IngestionAccepted        IngestionOutcome = "accepted"
	IngestionRejected        IngestionOutcome = "rejected"
	IngestionPrivacyFiltered IngestionOutcome = "privacy_filtered"
	IngestionBotFiltered     IngestionOutcome = "bot_filtered"
	IngestionDeduplicated    IngestionOutcome = "deduplicated"
)

type IngestionResult struct{ Outcome IngestionOutcome }

type CleanupResult struct {
	EventsDeleted   int64
	ReceiptsDeleted int64
}
