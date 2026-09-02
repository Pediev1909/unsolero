package ports

import (
	"context"
	"errors"
	"time"

	"rigmark/internal/modules/newsletter/domain"
)

var ErrNotFound = errors.New("newsletter subscription not found")

type Repository interface {
	// UpsertPending creates a pending row, or returns an existing pending or
	// unsubscribed address to pending with fresh tokens. It reports false and
	// changes nothing when the address is already confirmed, so the caller can
	// stay silent instead of mailing a subscriber who never asked again.
	UpsertPending(context.Context, domain.PendingSubscription) (bool, error)
	// Confirm consumes a live confirmation token exactly once. An unknown,
	// expired, or already-consumed hash returns ErrNotFound.
	Confirm(context.Context, []byte, time.Time) error
	// Unsubscribe marks the row owning the unsubscribe token as unsubscribed.
	// Repeating it is a no-op; an unknown hash returns ErrNotFound.
	Unsubscribe(context.Context, []byte, time.Time) error
	// PurgeExpiredPending deletes pending rows whose confirmation window has
	// closed and reports how many were removed.
	PurgeExpiredPending(context.Context, time.Time) (int64, error)
}

type ConfirmationMessage struct {
	Recipient string
	Token     string
	ExpiresAt time.Time
}

// Delivery is the newsletter's view of the email boundary. The same adapter
// that carries identity's security mail implements it.
type Delivery interface {
	SendNewsletterConfirmation(context.Context, ConfirmationMessage) error
}
