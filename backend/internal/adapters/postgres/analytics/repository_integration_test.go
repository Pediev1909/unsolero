package analyticspostgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

func TestConsentDeduplicationClaimAndRetention(t *testing.T) {
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
	now := time.Now().UTC()
	subject := sha256.Sum256([]byte(fmt.Sprintf("analytics-subject-%d", now.UnixNano())))
	otherSubject := sha256.Sum256([]byte(fmt.Sprintf("analytics-other-%d", now.UnixNano())))
	userID := insertAnalyticsTestUser(t, ctx, pool, "owner", now.UnixNano())
	otherUserID := insertAnalyticsTestUser(t, ctx, pool, "other", now.UnixNano())
	eventID := analyticsTestUUID(t, ctx, pool)
	claimedEventID := analyticsTestUUID(t, ctx, pool)
	expiredEventID := analyticsTestUUID(t, ctx, pool)
	postDeletionEventID := analyticsTestUUID(t, ctx, pool)
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM analytics.ingestion_receipts WHERE public_event_id=ANY($1::uuid[])`, []string{eventID, claimedEventID, expiredEventID, postDeletionEventID})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM analytics.events WHERE public_event_id=ANY($1::uuid[])`, []string{eventID, claimedEventID, expiredEventID, postDeletionEventID})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM analytics.identity_claims WHERE anonymous_subject_hash=ANY($1::bytea[])`, [][]byte{subject[:], otherSubject[:]})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM analytics.consent_history WHERE user_id=ANY($1::uuid[]) OR anonymous_subject_hash=ANY($2::bytea[])`, []string{userID, otherUserID}, [][]byte{subject[:], otherSubject[:]})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM analytics.consent_states WHERE user_id=ANY($1::uuid[]) OR anonymous_subject_hash=ANY($2::bytea[])`, []string{userID, otherUserID}, [][]byte{subject[:], otherSubject[:]})
		_, _ = pool.Exec(cleanupCtx, `DELETE FROM identity.users WHERE id=ANY($1::uuid[])`, []string{userID, otherUserID})
	})

	consent, err := repository.SetConsent(ctx, domain.ConsentDecision{Subject: domain.Subject{AnonymousSubjectHash: subject[:]}, RequestedState: "granted", PolicyVersion: domain.CurrentConsentPolicyVersion, Source: "banner", DecidedAt: now})
	if err != nil || consent.State != "granted" {
		t.Fatalf("SetConsent() = %#v, %v", consent, err)
	}
	stored, err := repository.GetConsent(ctx, domain.Subject{AnonymousSubjectHash: subject[:]})
	if err != nil || stored.State != "granted" {
		t.Fatalf("GetConsent() = %#v, %v", stored, err)
	}

	event := analyticsTestEvent(eventID, subject[:], now)
	const submissions = 8
	results := make(chan domain.IngestionOutcome, submissions)
	errorsChannel := make(chan error, submissions)
	var group sync.WaitGroup
	for range submissions {
		group.Add(1)
		go func() {
			defer group.Done()
			result, ingestErr := repository.Ingest(ctx, event, 24*time.Hour)
			if ingestErr != nil {
				errorsChannel <- ingestErr
				return
			}
			results <- result.Outcome
		}()
	}
	group.Wait()
	close(results)
	close(errorsChannel)
	for ingestErr := range errorsChannel {
		t.Fatalf("concurrent Ingest(): %v", ingestErr)
	}
	accepted, deduplicated := 0, 0
	for outcome := range results {
		if outcome == domain.IngestionAccepted {
			accepted++
		}
		if outcome == domain.IngestionDeduplicated {
			deduplicated++
		}
	}
	if accepted != 1 || deduplicated != submissions-1 {
		t.Fatalf("accepted=%d deduplicated=%d", accepted, deduplicated)
	}
	var eventCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM analytics.events WHERE public_event_id=$1`, eventID).Scan(&eventCount); err != nil || eventCount != 1 {
		t.Fatalf("event count=%d err=%v", eventCount, err)
	}

	if _, err := repository.SetConsent(ctx, domain.ConsentDecision{Subject: domain.Subject{AnonymousSubjectHash: subject[:]}, RequestedState: "denied", PolicyVersion: domain.CurrentConsentPolicyVersion, Source: "preferences", DecidedAt: now.Add(time.Minute)}); err != nil {
		t.Fatal(err)
	}
	newID := analyticsTestUUID(t, ctx, pool)
	withdrawnResult, err := repository.Ingest(ctx, analyticsTestEvent(newID, subject[:], now.Add(time.Minute)), 24*time.Hour)
	if err != nil || withdrawnResult.Outcome != domain.IngestionRejected {
		t.Fatalf("withdrawn ingest=%#v err=%v", withdrawnResult, err)
	}
	_, _ = pool.Exec(ctx, `DELETE FROM analytics.ingestion_receipts WHERE public_event_id=$1`, newID)

	if _, err := repository.SetConsent(ctx, domain.ConsentDecision{Subject: domain.Subject{AnonymousSubjectHash: otherSubject[:]}, RequestedState: "granted", PolicyVersion: domain.CurrentConsentPolicyVersion, Source: "banner", DecidedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.SetConsent(ctx, domain.ConsentDecision{Subject: domain.Subject{UserID: &userID}, RequestedState: "granted", PolicyVersion: domain.CurrentConsentPolicyVersion, Source: "account_sync", DecidedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.SetConsent(ctx, domain.ConsentDecision{Subject: domain.Subject{UserID: &otherUserID}, RequestedState: "granted", PolicyVersion: domain.CurrentConsentPolicyVersion, Source: "account_sync", DecidedAt: now}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Ingest(ctx, analyticsTestEvent(claimedEventID, otherSubject[:], now), 24*time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := repository.ClaimIdentity(ctx, otherSubject[:], userID, domain.CurrentConsentPolicyVersion, now); err != nil {
		t.Fatalf("ClaimIdentity(): %v", err)
	}
	if err := repository.ClaimIdentity(ctx, otherSubject[:], userID, domain.CurrentConsentPolicyVersion, now); err != nil {
		t.Fatalf("idempotent claim: %v", err)
	}
	if err := repository.ClaimIdentity(ctx, otherSubject[:], otherUserID, domain.CurrentConsentPolicyVersion, now); !errors.Is(err, ports.ErrIdentityClaimConflict) {
		t.Fatalf("cross-user claim error=%v", err)
	}
	var claimedUser string
	if err := pool.QueryRow(ctx, `SELECT user_id FROM analytics.events WHERE public_event_id=$1`, claimedEventID).Scan(&claimedUser); err != nil || claimedUser != userID {
		t.Fatalf("claimed user=%q err=%v", claimedUser, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE identity.users SET status='deleted',deleted_at=$2 WHERE id=$1`, userID, now); err != nil {
		t.Fatal(err)
	}
	authenticatedEvent := analyticsTestEvent(postDeletionEventID, nil, now)
	authenticatedEvent.UserID = &userID
	postDeletionResult, err := repository.Ingest(ctx, authenticatedEvent, 24*time.Hour)
	if err != nil || postDeletionResult.Outcome != domain.IngestionRejected {
		t.Fatalf("post-deletion ingest=%#v err=%v", postDeletionResult, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE identity.users SET status='active',deleted_at=NULL WHERE id=$1`, userID); err != nil {
		t.Fatal(err)
	}

	old := now.Add(-48 * time.Hour)
	_, err = pool.Exec(ctx, `INSERT INTO analytics.events (public_event_id,event_name,schema_version,session_id,surface,properties,consent_state,origin,classification,is_reportable,occurred_at,received_at,retention_expires_at)
		VALUES ($1,'page_view',3,$2,'test','{}','granted','client','human',true,$3,$3,$4)`, expiredEventID, "1191bb26-a9a2-41df-9346-74d693350ce8", old, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO analytics.ingestion_receipts (public_event_id,event_name,outcome,reason_code,received_at,retention_expires_at)
		VALUES ($1,'page_view','accepted','accepted',$2,$3)`, expiredEventID, old, now.Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	cleanup, err := repository.Cleanup(ctx, now, 100)
	if err != nil || cleanup.EventsDeleted < 1 || cleanup.ReceiptsDeleted < 1 {
		t.Fatalf("Cleanup()=%#v err=%v", cleanup, err)
	}
	secondCleanup, err := repository.Cleanup(ctx, now, 100)
	if err != nil || secondCleanup.EventsDeleted != 0 || secondCleanup.ReceiptsDeleted != 0 {
		t.Fatalf("idempotent Cleanup()=%#v err=%v", secondCleanup, err)
	}
}

func analyticsTestEvent(id string, subject []byte, now time.Time) domain.Event {
	sessionID := "1191bb26-a9a2-41df-9346-74d693350ce8"
	return domain.Event{ID: domain.EventID(id), Name: domain.EventPageView, SchemaVersion: 3, AnonymousSubjectHash: subject,
		SessionID: &sessionID, Surface: "test", Properties: map[string]json.RawMessage{}, ConsentState: "granted",
		ConsentPolicyVersion: domain.CurrentConsentPolicyVersion, Origin: "client", Classification: "human", Reportable: true,
		RetentionExpiresAt: now.Add(24 * time.Hour), OccurredAt: now}
}

func analyticsTestUUID(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(ctx, `SELECT gen_random_uuid()`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertAnalyticsTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, label string, nonce int64) string {
	t.Helper()
	var id string
	if err := pool.QueryRow(ctx, `INSERT INTO identity.users (email,status) VALUES ($1,'active') RETURNING id`, fmt.Sprintf("analytics-%s-%d@example.invalid", label, nonce)).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}
