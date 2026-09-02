// Package application implements the double opt-in newsletter flow. Subscribe
// records a pending address and mails a one-time link; Confirm consumes that
// link; Unsubscribe honours the token that will accompany every newsletter.
package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"rigmark/internal/modules/newsletter/domain"
	"rigmark/internal/modules/newsletter/ports"
)

// ErrInvalidToken covers unknown, expired, malformed, and already-used tokens
// alike. Distinguishing them would tell a guesser which hashes exist.
var ErrInvalidToken = errors.New("newsletter token is invalid or expired")

// DefaultConfirmationTTL is how long a confirmation link stays usable. Two days
// covers a weekend inbox without leaving unconfirmed addresses around for long.
const DefaultConfirmationTTL = 48 * time.Hour

type ValidationError struct {
	Fields map[string]string
}

func (err ValidationError) Error() string {
	return "newsletter request failed validation"
}

// Tokens is satisfied by the identity session token manager: 32 random bytes
// encoded raw for the email, SHA-256 hashed for storage.
type Tokens interface {
	Generate() (string, []byte, error)
	Hash(string) ([]byte, error)
}

type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now().UTC()
}

type Config struct {
	ConfirmationTTL time.Duration
}

// Receipt is deliberately neutral: it never says whether the address was new,
// already confirmed, or unknown.
type Receipt struct {
	Recorded bool `json:"recorded"`
}

type Service struct {
	repository ports.Repository
	tokens     Tokens
	delivery   ports.Delivery
	clock      Clock
	config     Config
}

func NewService(repository ports.Repository, tokens Tokens, delivery ports.Delivery, config Config) (*Service, error) {
	return newService(repository, tokens, delivery, systemClock{}, config)
}

func newService(repository ports.Repository, tokens Tokens, delivery ports.Delivery, clock Clock, config Config) (*Service, error) {
	if repository == nil || tokens == nil || delivery == nil || clock == nil {
		return nil, errors.New("newsletter dependencies are required")
	}
	if config.ConfirmationTTL < time.Hour || config.ConfirmationTTL > 7*24*time.Hour {
		return nil, errors.New("newsletter confirmation lifetime must be between one hour and seven days")
	}
	return &Service{repository: repository, tokens: tokens, delivery: delivery, clock: clock, config: config}, nil
}

// Subscribe validates the request, records or refreshes a pending row, and
// mails the confirmation link. A confirmed address is left alone and receives
// nothing; the receipt looks the same either way. A delivery failure is
// returned alongside a recorded receipt so the caller can log it without
// telling the requester anything about the address.
func (service *Service) Subscribe(ctx context.Context, email, source string) (Receipt, error) {
	fields := make(map[string]string)
	normalized, err := domain.NormalizeEmail(email)
	if err != nil {
		fields["email"] = "Enter a valid email address."
	}
	if err := domain.ValidateSource(source); err != nil {
		fields["source"] = "The subscription source is invalid."
	}
	if len(fields) > 0 {
		return Receipt{}, ValidationError{Fields: fields}
	}
	rawConfirm, confirmHash, err := service.tokens.Generate()
	if err != nil {
		return Receipt{}, fmt.Errorf("generate newsletter confirmation token: %w", err)
	}
	_, unsubscribeHash, err := service.tokens.Generate()
	if err != nil {
		return Receipt{}, fmt.Errorf("generate newsletter unsubscribe token: %w", err)
	}
	now := service.clock.Now()
	expires := now.Add(service.config.ConfirmationTTL)
	created, err := service.repository.UpsertPending(ctx, domain.PendingSubscription{
		Email: normalized, Source: source, ConsentTextVersion: domain.ConsentTextVersion,
		ConfirmTokenHash: confirmHash, ConfirmExpiresAt: expires,
		UnsubscribeTokenHash: unsubscribeHash, RequestedAt: now,
	})
	if err != nil {
		return Receipt{}, fmt.Errorf("record newsletter subscription: %w", err)
	}
	if !created {
		return Receipt{Recorded: true}, nil
	}
	if err := service.delivery.SendNewsletterConfirmation(ctx, ports.ConfirmationMessage{
		Recipient: normalized, Token: rawConfirm, ExpiresAt: expires,
	}); err != nil {
		return Receipt{Recorded: true}, fmt.Errorf("deliver newsletter confirmation: %w", err)
	}
	return Receipt{Recorded: true}, nil
}

// Confirm consumes a confirmation token once and moves the address to
// confirmed.
func (service *Service) Confirm(ctx context.Context, rawToken string) error {
	hash, err := service.tokens.Hash(strings.TrimSpace(rawToken))
	if err != nil {
		return ErrInvalidToken
	}
	return translateTokenError(service.repository.Confirm(ctx, hash, service.clock.Now()), "confirm newsletter subscription")
}

// Unsubscribe marks the address unsubscribed. It is idempotent: repeating a
// valid token succeeds again without changing anything.
func (service *Service) Unsubscribe(ctx context.Context, rawToken string) error {
	hash, err := service.tokens.Hash(strings.TrimSpace(rawToken))
	if err != nil {
		return ErrInvalidToken
	}
	return translateTokenError(service.repository.Unsubscribe(ctx, hash, service.clock.Now()), "unsubscribe newsletter address")
}

// PurgeExpiredPending removes addresses that never confirmed. An unconfirmed
// address has given no consent, so nothing about it is kept once its link is
// dead.
func (service *Service) PurgeExpiredPending(ctx context.Context) (int64, error) {
	return service.repository.PurgeExpiredPending(ctx, service.clock.Now())
}

func translateTokenError(err error, operation string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ports.ErrNotFound):
		return ErrInvalidToken
	default:
		return fmt.Errorf("%s: %w", operation, err)
	}
}
