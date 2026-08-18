package ports

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/analytics/domain"
)

var ErrConsentNotFound = errors.New("analytics consent not found")
var ErrIdentityClaimConflict = errors.New("analytics identity already claimed")
var ErrIdentityClaimNotAllowed = errors.New("analytics identity claim not allowed")

type EventRecorder interface {
	Ingest(context.Context, domain.Event, time.Duration) (domain.IngestionResult, error)
	RecordRejected(context.Context, domain.EventID, string, domain.IngestionOutcome, string, time.Duration) error
	SetConsent(context.Context, domain.ConsentDecision) (domain.Consent, error)
	GetConsent(context.Context, domain.Subject) (domain.Consent, error)
	ClaimIdentity(context.Context, []byte, string, string, time.Time) error
	Cleanup(context.Context, time.Time, int) (domain.CleanupResult, error)
}
