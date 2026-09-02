package newsletterpostgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/newsletter/domain"
	"rigmark/internal/modules/newsletter/ports"
)

func TestSubscriptionLifecycleAgainstPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)
	email := fmt.Sprintf("lifecycle-%d@example.invalid", now.UnixNano())
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM audience.newsletter_subscriptions WHERE email = $1`, email)
	})
	first := pendingFor(email, "footer", "first", now)

	created, err := repository.UpsertPending(ctx, first)
	if err != nil || !created {
		t.Fatalf("first UpsertPending() = %v, %v; want created", created, err)
	}
	// A second request before confirmation refreshes the tokens: the old link
	// must stop working and the new one must work.
	second := pendingFor(email, "article:mailchimp-alternatives", "second", now.Add(time.Minute))
	created, err = repository.UpsertPending(ctx, second)
	if err != nil || !created {
		t.Fatalf("refresh UpsertPending() = %v, %v; want created", created, err)
	}
	if err := repository.Confirm(ctx, first.ConfirmTokenHash, now.Add(2*time.Minute)); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("superseded token Confirm() error = %v, want ErrNotFound", err)
	}
	if err := repository.Confirm(ctx, second.ConfirmTokenHash, second.ConfirmExpiresAt.Add(time.Second)); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("expired token Confirm() error = %v, want ErrNotFound", err)
	}
	if err := repository.Confirm(ctx, second.ConfirmTokenHash, now.Add(2*time.Minute)); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if err := repository.Confirm(ctx, second.ConfirmTokenHash, now.Add(3*time.Minute)); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("second Confirm() error = %v, want ErrNotFound (consume once)", err)
	}
	assertRow(t, pool, email, domain.StatusConfirmed, "article:mailchimp-alternatives")

	// A confirmed subscriber who submits the form again is left untouched.
	created, err = repository.UpsertPending(ctx, pendingFor(email, "footer", "third", now.Add(4*time.Minute)))
	if err != nil || created {
		t.Fatalf("confirmed UpsertPending() = %v, %v; want not created", created, err)
	}
	assertRow(t, pool, email, domain.StatusConfirmed, "article:mailchimp-alternatives")

	for range 2 {
		if err := repository.Unsubscribe(ctx, second.UnsubscribeTokenHash, now.Add(5*time.Minute)); err != nil {
			t.Fatalf("Unsubscribe() error = %v", err)
		}
	}
	assertRow(t, pool, email, domain.StatusUnsubscribed, "article:mailchimp-alternatives")
	if err := repository.Unsubscribe(ctx, digest("unknown"), now); !errors.Is(err, ports.ErrNotFound) {
		t.Errorf("unknown Unsubscribe() error = %v, want ErrNotFound", err)
	}

	// Re-subscribing after unsubscribing starts the opt-in over.
	fourth := pendingFor(email, "footer", "fourth", now.Add(6*time.Minute))
	created, err = repository.UpsertPending(ctx, fourth)
	if err != nil || !created {
		t.Fatalf("re-subscribe UpsertPending() = %v, %v; want created", created, err)
	}
	assertRow(t, pool, email, domain.StatusPending, "footer")

	purged, err := repository.PurgeExpiredPending(ctx, fourth.ConfirmExpiresAt.Add(time.Second))
	if err != nil || purged < 1 {
		t.Fatalf("PurgeExpiredPending() = %d, %v; want at least the expired row", purged, err)
	}
	var remaining int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM audience.newsletter_subscriptions WHERE email = $1`, email).Scan(&remaining); err != nil || remaining != 0 {
		t.Errorf("rows after purge = %d, %v; want 0", remaining, err)
	}
}

func pendingFor(email, source, seed string, now time.Time) domain.PendingSubscription {
	return domain.PendingSubscription{
		Email: email, Source: source, ConsentTextVersion: domain.ConsentTextVersion,
		ConfirmTokenHash: digest("confirm-" + seed + email), ConfirmExpiresAt: now.Add(48 * time.Hour),
		UnsubscribeTokenHash: digest("unsubscribe-" + seed + email), RequestedAt: now,
	}
}

func assertRow(t *testing.T, pool *pgxpool.Pool, email string, want domain.Status, wantSource string) {
	t.Helper()
	var status, source string
	if err := pool.QueryRow(context.Background(), `SELECT status, source FROM audience.newsletter_subscriptions WHERE email = $1`, email).Scan(&status, &source); err != nil {
		t.Fatalf("load row: %v", err)
	}
	if status != string(want) || source != wantSource {
		t.Errorf("row = %s/%s, want %s/%s", status, source, want, wantSource)
	}
}

func digest(value string) []byte {
	sum := sha256.Sum256([]byte(value))
	return sum[:]
}
