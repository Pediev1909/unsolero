// Package newsletterpostgres stores the double opt-in list in
// audience.newsletter_subscriptions. Every statement here works on token
// hashes; the raw tokens never reach the database.
package newsletterpostgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/newsletter/domain"
	"rigmark/internal/modules/newsletter/ports"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// UpsertPending inserts the address as pending or, when it already exists in
// a non-confirmed state, returns it to pending with the new tokens. The
// conflict target is the unique lower(email) index, so the whole decision is
// one statement and concurrent requests for the same address cannot race.
// A confirmed row is excluded by the WHERE clause, which makes the statement
// return no row and the method report false.
func (repository *Repository) UpsertPending(ctx context.Context, pending domain.PendingSubscription) (bool, error) {
	var recorded bool
	err := repository.pool.QueryRow(ctx, `
		INSERT INTO audience.newsletter_subscriptions (
			email, status, confirm_token_hash, confirm_expires_at, unsubscribe_token_hash,
			source, consent_text_version, requested_at, created_at, updated_at
		) VALUES ($1, 'pending', $2, $3, $4, $5, $6, $7, $7, $7)
		ON CONFLICT ((lower(email))) DO UPDATE SET
			status = 'pending',
			confirm_token_hash = EXCLUDED.confirm_token_hash,
			confirm_expires_at = EXCLUDED.confirm_expires_at,
			unsubscribe_token_hash = EXCLUDED.unsubscribe_token_hash,
			source = EXCLUDED.source,
			consent_text_version = EXCLUDED.consent_text_version,
			requested_at = EXCLUDED.requested_at,
			confirmed_at = NULL,
			unsubscribed_at = NULL,
			updated_at = EXCLUDED.updated_at
		WHERE audience.newsletter_subscriptions.status <> 'confirmed'
		RETURNING true`,
		pending.Email, pending.ConfirmTokenHash, pending.ConfirmExpiresAt, pending.UnsubscribeTokenHash,
		pending.Source, pending.ConsentTextVersion, pending.RequestedAt,
	).Scan(&recorded)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("upsert newsletter subscription: %w", err)
	}
	return recorded, nil
}

func (repository *Repository) Confirm(ctx context.Context, tokenHash []byte, now time.Time) error {
	tag, err := repository.pool.Exec(ctx, `
		UPDATE audience.newsletter_subscriptions
		SET status = 'confirmed', confirmed_at = $2, confirm_token_hash = NULL,
			confirm_expires_at = NULL, updated_at = $2
		WHERE confirm_token_hash = $1 AND status = 'pending' AND confirm_expires_at > $2`,
		tokenHash, now)
	if err != nil {
		return fmt.Errorf("confirm newsletter subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

// Unsubscribe matches on the unsubscribe hash regardless of status, so a
// second click on the same link finds the row again and succeeds without
// moving any timestamp.
func (repository *Repository) Unsubscribe(ctx context.Context, tokenHash []byte, now time.Time) error {
	tag, err := repository.pool.Exec(ctx, `
		UPDATE audience.newsletter_subscriptions
		SET status = 'unsubscribed',
			unsubscribed_at = COALESCE(unsubscribed_at, $2),
			confirm_token_hash = NULL,
			confirm_expires_at = NULL,
			updated_at = CASE WHEN status = 'unsubscribed' THEN updated_at ELSE $2 END
		WHERE unsubscribe_token_hash = $1`,
		tokenHash, now)
	if err != nil {
		return fmt.Errorf("unsubscribe newsletter address: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ports.ErrNotFound
	}
	return nil
}

func (repository *Repository) PurgeExpiredPending(ctx context.Context, now time.Time) (int64, error) {
	tag, err := repository.pool.Exec(ctx, `
		DELETE FROM audience.newsletter_subscriptions
		WHERE status = 'pending' AND confirm_expires_at < $1`, now)
	if err != nil {
		return 0, fmt.Errorf("purge expired newsletter subscriptions: %w", err)
	}
	return tag.RowsAffected(), nil
}
