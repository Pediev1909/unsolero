package identitypostgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

func TestConcurrentIdentitySingleUseAndRegistrationInvariants(t *testing.T) {
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

	email := fmt.Sprintf("concurrent-registration-%d@example.invalid", now.UnixNano())
	const contenders = 8
	registrationErrors := concurrently(contenders, func(index int) error {
		_, registerErr := repository.RegisterWithSession(context.Background(), email, "argon2id-test-hash", domain.Session{
			TokenHash: digest(fmt.Sprintf("registration-session-%d-%d", now.UnixNano(), index)),
			ExpiresAt: now.Add(24 * time.Hour), IdleExpiresAt: now.Add(time.Hour), LastSeenAt: now, CreatedAt: now,
		})
		return registerErr
	})
	succeeded, conflicted := 0, 0
	for _, registerErr := range registrationErrors {
		switch {
		case registerErr == nil:
			succeeded++
		case errors.Is(registerErr, ports.ErrConflict):
			conflicted++
		default:
			t.Fatalf("concurrent registration error = %v", registerErr)
		}
	}
	if succeeded != 1 || conflicted != contenders-1 {
		t.Fatalf("registration successes=%d conflicts=%d", succeeded, conflicted)
	}
	var userID domain.UserID
	if err := pool.QueryRow(ctx, `SELECT id FROM identity.users WHERE lower(email)=$1`, email).Scan(&userID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, userID) })

	resetHash := digest("concurrent-reset:" + string(userID))
	if _, created, createErr := repository.CreatePasswordResetToken(ctx, email, resetHash, now.Add(time.Hour), now,
		securityTestEvent("password_reset.request", now)); createErr != nil || !created {
		t.Fatalf("create reset token created=%v error=%v", created, createErr)
	}
	resetErrors := concurrently(contenders, func(index int) error {
		return repository.ConsumePasswordResetToken(context.Background(), resetHash,
			fmt.Sprintf("replacement-hash-%d", index), now.Add(time.Minute), securityTestEvent("password_reset.complete", now))
	})
	succeeded, consumed := 0, 0
	for _, resetErr := range resetErrors {
		switch {
		case resetErr == nil:
			succeeded++
		case errors.Is(resetErr, ports.ErrConsumed):
			consumed++
		default:
			t.Fatalf("concurrent reset error = %v", resetErr)
		}
	}
	if succeeded != 1 || consumed != contenders-1 {
		t.Fatalf("reset successes=%d consumed=%d", succeeded, consumed)
	}

	if _, err := repository.UpsertPendingMFA(ctx, userID, make([]byte, 32), make([]byte, 12), 1, now,
		securityTestEvent("mfa.enrollment_started", now)); err != nil {
		t.Fatal(err)
	}
	recoveryHash := digest("concurrent-recovery:" + string(userID))
	if err := repository.EnableMFA(ctx, userID, [][]byte{recoveryHash}, now, securityTestEvent("mfa.enrollment_complete", now)); err != nil {
		t.Fatal(err)
	}
	recoveryResults := make(chan bool, contenders)
	recoveryErrors := concurrently(contenders, func(_ int) error {
		used, consumeErr := repository.ConsumeRecoveryCode(context.Background(), userID, recoveryHash,
			now.Add(time.Minute), securityTestEvent("mfa.recovery_code", now))
		recoveryResults <- used
		return consumeErr
	})
	close(recoveryResults)
	usedCount := 0
	for used := range recoveryResults {
		if used {
			usedCount++
		}
	}
	for _, recoveryErr := range recoveryErrors {
		if recoveryErr != nil {
			t.Fatalf("concurrent recovery code error = %v", recoveryErr)
		}
	}
	if usedCount != 1 {
		t.Fatalf("recovery code consumed %d times, want 1", usedCount)
	}
}

func concurrently(count int, operation func(int) error) []error {
	results := make([]error, count)
	start := make(chan struct{})
	var wait sync.WaitGroup
	for index := range count {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			<-start
			results[index] = operation(index)
		}(index)
	}
	close(start)
	wait.Wait()
	return results
}

func TestSecurityTokenSessionAndOwnershipLifecycle(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("create test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	repository := New(pool)
	now := time.Now().UTC().Truncate(time.Microsecond)

	user, originalSessionHash := createSecurityTestUser(t, ctx, repository, pool, now, "owner")
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, user.ID) })

	verificationHash := digest("verification-token:" + string(user.ID))
	recipient, created, err := repository.CreateEmailVerificationToken(ctx, user.Email, verificationHash,
		now.Add(time.Hour), now, securityTestEvent("email_verification.request", now))
	if err != nil || !created || recipient != user.Email {
		t.Fatalf("CreateEmailVerificationToken() = %q,%v,%v", recipient, created, err)
	}
	if err := repository.ConsumeEmailVerificationToken(ctx, verificationHash, now.Add(time.Minute), securityTestEvent("email_verification.complete", now)); err != nil {
		t.Fatalf("ConsumeEmailVerificationToken() error = %v", err)
	}
	if err := repository.ConsumeEmailVerificationToken(ctx, verificationHash, now.Add(2*time.Minute), securityTestEvent("email_verification.complete", now)); !errors.Is(err, ports.ErrConsumed) {
		t.Fatalf("verification token reuse error = %v, want ErrConsumed", err)
	}
	var verified bool
	if err := pool.QueryRow(ctx, `SELECT email_verified_at IS NOT NULL FROM identity.users WHERE id=$1`, user.ID).Scan(&verified); err != nil || !verified {
		t.Fatalf("verified state = %v, error = %v", verified, err)
	}

	expiredHash := digest("expired-reset-token:" + string(user.ID))
	_, created, err = repository.CreatePasswordResetToken(ctx, user.Email, expiredHash, now.Add(time.Minute), now,
		securityTestEvent("password_reset.request", now))
	if err != nil || !created {
		t.Fatalf("CreatePasswordResetToken() created=%v error=%v", created, err)
	}
	if err := repository.ConsumePasswordResetToken(ctx, expiredHash, "new-hash", now.Add(2*time.Minute), securityTestEvent("password_reset.complete", now)); !errors.Is(err, ports.ErrExpired) {
		t.Fatalf("expired reset error = %v, want ErrExpired", err)
	}

	resetHash := digest("valid-reset-token:" + string(user.ID))
	_, _, err = repository.CreatePasswordResetToken(ctx, user.Email, resetHash, now.Add(time.Hour), now,
		securityTestEvent("password_reset.request", now))
	if err != nil {
		t.Fatalf("create valid reset: %v", err)
	}
	if err := repository.ConsumePasswordResetToken(ctx, resetHash, "replacement-password-hash", now.Add(time.Minute), securityTestEvent("password_reset.complete", now)); err != nil {
		t.Fatalf("consume valid reset: %v", err)
	}
	if _, err := repository.ResolveSession(ctx, originalSessionHash, now.Add(2*time.Minute), now.Add(time.Hour)); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("session survived reset: %v", err)
	}

	activeHash := digest("active-after-reset:" + string(user.ID))
	activeSession := domain.Session{UserID: user.ID, TokenHash: activeHash, ExpiresAt: now.Add(24 * time.Hour),
		IdleExpiresAt: now.Add(time.Hour), LastSeenAt: now.Add(3 * time.Minute), CreatedAt: now.Add(3 * time.Minute)}
	if err := repository.CreateLoginSession(ctx, activeSession, now.Add(3*time.Minute), nil); err != nil {
		t.Fatalf("create post-reset session: %v", err)
	}
	var activeSessionID string
	if err := pool.QueryRow(ctx, `SELECT id FROM identity.sessions WHERE token_hash=$1`, activeHash).Scan(&activeSessionID); err != nil {
		t.Fatal(err)
	}
	if err := repository.MarkSessionMFA(ctx, user.ID, activeSessionID, now.Add(4*time.Minute), "password_mfa", securityTestEvent("mfa.step_up", now)); err != nil {
		t.Fatalf("MarkSessionMFA() error = %v", err)
	}
	principal, err := repository.ResolveSession(ctx, activeHash, now.Add(5*time.Minute), now.Add(time.Hour))
	if err != nil || principal.MFAAuthenticatedAt == nil {
		t.Fatalf("MFA principal=%#v error=%v", principal, err)
	}

	credential, err := repository.UpsertPendingMFA(ctx, user.ID, make([]byte, 32), make([]byte, 12), 1, now,
		securityTestEvent("mfa.enrollment_started", now))
	if err != nil {
		t.Fatalf("UpsertPendingMFA() error = %v", err)
	}
	firstRecovery := digest("first-recovery:" + string(user.ID))
	if err := repository.EnableMFA(ctx, user.ID, [][]byte{firstRecovery}, now.Add(time.Minute), securityTestEvent("mfa.enrollment_complete", now)); err != nil {
		t.Fatal(err)
	}
	used, err := repository.ConsumeRecoveryCode(ctx, user.ID, firstRecovery, now.Add(2*time.Minute), securityTestEvent("mfa.recovery_code", now))
	if err != nil || !used {
		t.Fatalf("first recovery use=%v error=%v credential=%s", used, err, credential.ID)
	}
	used, err = repository.ConsumeRecoveryCode(ctx, user.ID, firstRecovery, now.Add(3*time.Minute), securityTestEvent("mfa.recovery_code", now))
	if err != nil || used {
		t.Fatalf("reused recovery use=%v error=%v", used, err)
	}
	replacementRecovery := digest("replacement-recovery:" + string(user.ID))
	if err := repository.ReplaceRecoveryCodes(ctx, user.ID, [][]byte{replacementRecovery}, now.Add(4*time.Minute), securityTestEvent("mfa.recovery_codes_regenerated", now)); err != nil {
		t.Fatal(err)
	}
	used, err = repository.ConsumeRecoveryCode(ctx, user.ID, replacementRecovery, now.Add(5*time.Minute), securityTestEvent("mfa.recovery_code", now))
	if err != nil || !used {
		t.Fatalf("replacement recovery use=%v error=%v", used, err)
	}

	other, _ := createSecurityTestUser(t, ctx, repository, pool, now.Add(time.Second), "other")
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM identity.users WHERE id=$1`, other.ID) })
	var otherSessionID string
	if err := pool.QueryRow(ctx, `SELECT id FROM identity.sessions WHERE user_id=$1 ORDER BY created_at DESC LIMIT 1`, other.ID).Scan(&otherSessionID); err != nil {
		t.Fatal(err)
	}
	if err := repository.RevokeOwnedSession(ctx, user.ID, otherSessionID, otherSessionID, now.Add(3*time.Minute), securityTestEvent("session.revoke", now)); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("cross-account session revoke error = %v, want ErrNotFound", err)
	}

	event := securityTestEvent("authorization", now)
	event.UserID = &user.ID
	if err := repository.RecordSecurityEvent(ctx, event); err != nil {
		t.Fatalf("RecordSecurityEvent() error = %v", err)
	}
	var eventID string
	if err := pool.QueryRow(ctx, `SELECT id FROM identity.security_events WHERE user_id=$1 ORDER BY occurred_at DESC LIMIT 1`, user.ID).Scan(&eventID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE identity.security_events SET outcome='failure' WHERE id=$1`, eventID); err == nil {
		t.Fatal("immutable security event accepted an update")
	}
	var analyticsEventID string
	if err := pool.QueryRow(ctx, `INSERT INTO analytics.consent_states (user_id,state,policy_version,source,decided_at)
		VALUES ($1,'granted','analytics-v1','account_sync',$2) RETURNING id`, user.ID, now).Scan(new(string)); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO analytics.consent_history (user_id,state,policy_version,source,decided_at)
		VALUES ($1,'granted','analytics-v1','account_sync',$2)`, user.ID, now); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `INSERT INTO analytics.events (public_event_id,event_name,schema_version,user_id,session_id,surface,properties,
		consent_state,consent_policy_version,origin,classification,is_reportable,occurred_at,retention_expires_at)
		VALUES (gen_random_uuid(),'page_view',3,$1,$2,'account','{}','granted','analytics-v1','client','human',true,$3,$4)
		RETURNING public_event_id`, user.ID, activeSessionID, now, now.Add(24*time.Hour)).Scan(&analyticsEventID); err != nil {
		t.Fatal(err)
	}

	export, err := repository.ExportAccount(ctx, user.ID, now)
	if err != nil {
		t.Fatalf("ExportAccount() error = %v", err)
	}
	if export.Account["email"] != user.Email {
		t.Fatalf("export account = %#v", export.Account)
	}
	if len(export.ConsentEvents) != 1 || len(export.AnalyticsEvents) != 1 {
		t.Fatalf("privacy export consent=%#v analytics=%#v", export.ConsentEvents, export.AnalyticsEvents)
	}

	if err := repository.DeleteAccount(ctx, user.ID, activeSessionID, now.Add(10*time.Minute), securityTestEvent("account.delete", now)); err != nil {
		t.Fatalf("DeleteAccount() error = %v", err)
	}
	var status, anonymizedEmail string
	if err := pool.QueryRow(ctx, `SELECT status,email FROM identity.users WHERE id=$1`, user.ID).Scan(&status, &anonymizedEmail); err != nil {
		t.Fatal(err)
	}
	if status != "deleted" || !strings.HasSuffix(anonymizedEmail, "@users.invalid") || strings.Contains(anonymizedEmail, "owner") {
		t.Fatalf("deleted account status=%q email=%q", status, anonymizedEmail)
	}
	var analyticsLinked, currentConsent, anonymizedConsent bool
	if err := pool.QueryRow(ctx, `SELECT user_id IS NOT NULL FROM analytics.events WHERE public_event_id=$1`, analyticsEventID).Scan(&analyticsLinked); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM analytics.consent_states WHERE user_id=$1)`, user.ID).Scan(&currentConsent); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM analytics.consent_history WHERE anonymized_subject_id IS NOT NULL AND user_id IS NULL AND decided_at=$1)`, now).Scan(&anonymizedConsent); err != nil {
		t.Fatal(err)
	}
	if analyticsLinked || currentConsent || !anonymizedConsent {
		t.Fatalf("deletion analytics_linked=%v current_consent=%v anonymized_consent=%v", analyticsLinked, currentConsent, anonymizedConsent)
	}
	if err := repository.DeleteAccount(ctx, user.ID, activeSessionID, now.Add(11*time.Minute), securityTestEvent("account.delete", now)); !errors.Is(err, ports.ErrNotFound) {
		t.Fatalf("repeated deletion error=%v", err)
	}
	retainedEvent := securityTestEvent("session.revoke", now.Add(10*time.Minute))
	retainedEvent.UserID = &user.ID
	retainedEvent.SessionID = &activeSessionID
	if err := repository.RecordSecurityEvent(ctx, retainedEvent); err != nil {
		t.Fatal(err)
	}
	if err := repository.CleanupExpiredSecurityArtifacts(ctx, now.Add(50*24*time.Hour)); err != nil {
		t.Fatalf("CleanupExpiredSecurityArtifacts() error = %v", err)
	}
	var sessionExists, eventExists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identity.sessions WHERE id=$1)`, activeSessionID).Scan(&sessionExists); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identity.security_events WHERE session_id=$1)`, activeSessionID).Scan(&eventExists); err != nil {
		t.Fatal(err)
	}
	if sessionExists || !eventExists {
		t.Fatalf("cleanup session_exists=%v event_exists=%v", sessionExists, eventExists)
	}
}

func createSecurityTestUser(t *testing.T, ctx context.Context, repository *Repository, pool *pgxpool.Pool,
	now time.Time, prefix string) (domain.User, []byte) {
	t.Helper()
	email := fmt.Sprintf("security-%s-%d@example.invalid", prefix, now.UnixNano())
	hash := digest(email)
	user, err := repository.RegisterWithSession(ctx, email, "argon2id-test-hash", domain.Session{
		TokenHash: hash, ExpiresAt: now.Add(24 * time.Hour), IdleExpiresAt: now.Add(time.Hour),
		LastSeenAt: now, CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("RegisterWithSession() error = %v", err)
	}
	if !strings.EqualFold(user.Email, email) {
		t.Fatalf("created email = %q", user.Email)
	}
	return user, hash
}

func digest(value string) []byte { sum := sha256.Sum256([]byte(value)); return sum[:] }

func securityTestEvent(eventType string, now time.Time) domain.SecurityEvent {
	return domain.SecurityEvent{Type: eventType, Outcome: "success", Surface: "test", OccurredAt: now}
}
