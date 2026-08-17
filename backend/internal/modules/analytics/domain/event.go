package domain

import (
	"encoding/json"
	"time"
)

type EventID string

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
	OccurredAt              time.Time
	ReceivedAt              time.Time
}
