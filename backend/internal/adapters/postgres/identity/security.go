package identitypostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"rigmark/internal/modules/identity/domain"
	"rigmark/internal/modules/identity/ports"
)

func (repository *Repository) CreateEmailVerificationToken(ctx context.Context, email string, tokenHash []byte,
	expiresAt, now time.Time, event domain.SecurityEvent) (string, bool, error) {
	return repository.createEmailToken(ctx, "identity.email_verification_tokens", email, tokenHash, expiresAt, now, event, true)
}

func (repository *Repository) CreatePasswordResetToken(ctx context.Context, email string, tokenHash []byte,
	expiresAt, now time.Time, event domain.SecurityEvent) (string, bool, error) {
	return repository.createEmailToken(ctx, "identity.password_reset_tokens", email, tokenHash, expiresAt, now, event, false)
}

func (repository *Repository) createEmailToken(ctx context.Context, table, email string, tokenHash []byte,
	expiresAt, now time.Time, event domain.SecurityEvent, requireUnverified bool) (string, bool, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return "", false, fmt.Errorf("begin security token request: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	query := `SELECT id,email FROM identity.users WHERE lower(email)=$1 AND status='active' AND deleted_at IS NULL`
	if requireUnverified {
		query += ` AND email_verified_at IS NULL`
	}
	var userID domain.UserID
	var recipient string
	err = tx.QueryRow(ctx, query, normalizeEmail(email)).Scan(&userID, &recipient)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := insertSecurityEvent(ctx, tx, event); err != nil {
			return "", false, err
		}
		if err := tx.Commit(ctx); err != nil {
			return "", false, fmt.Errorf("commit generic token request: %w", err)
		}
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("resolve token recipient: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE `+table+` SET invalidated_at=$2
		WHERE user_id=$1 AND consumed_at IS NULL AND invalidated_at IS NULL`, userID, now); err != nil {
		return "", false, fmt.Errorf("invalidate earlier security tokens: %w", err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO `+table+` (user_id,token_hash,expires_at,created_at) VALUES ($1,$2,$3,$4)`,
		userID, tokenHash, expiresAt, now); err != nil {
		return "", false, fmt.Errorf("create security token: %w", err)
	}
	event.UserID = &userID
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return "", false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", false, fmt.Errorf("commit security token request: %w", err)
	}
	return recipient, true, nil
}

func (repository *Repository) ConsumeEmailVerificationToken(ctx context.Context, tokenHash []byte, now time.Time,
	event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin email verification: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	userID, err := consumeToken(ctx, tx, "identity.email_verification_tokens", tokenHash, now)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.users SET email_verified_at=COALESCE(email_verified_at,$2),updated_at=$2
		WHERE id=$1 AND status='active' AND deleted_at IS NULL`, userID, now); err != nil {
		return fmt.Errorf("verify user email: %w", err)
	}
	event.UserID = &userID
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "email verification")
}

func (repository *Repository) ConsumePasswordResetToken(ctx context.Context, tokenHash []byte, passwordHash string,
	now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password reset: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	userID, err := consumeToken(ctx, tx, "identity.password_reset_tokens", tokenHash, now)
	if err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE identity.users SET password_hash=$2,updated_at=$3
		WHERE id=$1 AND status='active' AND deleted_at IS NULL`, userID, passwordHash, now)
	if err != nil {
		return fmt.Errorf("replace reset password: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ports.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.sessions SET revoked_at=COALESCE(revoked_at,$2) WHERE user_id=$1 AND revoked_at IS NULL`, userID, now); err != nil {
		return fmt.Errorf("revoke reset sessions: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.password_reset_tokens SET invalidated_at=$2
		WHERE user_id=$1 AND consumed_at IS NULL AND invalidated_at IS NULL`, userID, now); err != nil {
		return fmt.Errorf("invalidate reset tokens: %w", err)
	}
	event.UserID = &userID
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "password reset")
}

func consumeToken(ctx context.Context, tx pgx.Tx, table string, tokenHash []byte, now time.Time) (domain.UserID, error) {
	var userID domain.UserID
	var expires time.Time
	var consumed, invalidated sql.NullTime
	err := tx.QueryRow(ctx, `SELECT user_id,expires_at,consumed_at,invalidated_at FROM `+table+` WHERE token_hash=$1 FOR UPDATE`, tokenHash).
		Scan(&userID, &expires, &consumed, &invalidated)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ports.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("load security token: %w", err)
	}
	if consumed.Valid || invalidated.Valid {
		return "", ports.ErrConsumed
	}
	if !expires.After(now) {
		_, _ = tx.Exec(ctx, `UPDATE `+table+` SET invalidated_at=$2 WHERE token_hash=$1`, tokenHash, now)
		return "", ports.ErrExpired
	}
	if _, err := tx.Exec(ctx, `UPDATE `+table+` SET consumed_at=$2 WHERE token_hash=$1`, tokenHash, now); err != nil {
		return "", fmt.Errorf("consume security token: %w", err)
	}
	return userID, nil
}

func (repository *Repository) GetPasswordCredentialByID(ctx context.Context, userID domain.UserID) (domain.PasswordCredential, error) {
	var credential domain.PasswordCredential
	var verified, lastLogin, deleted sql.NullTime
	var hash sql.NullString
	err := repository.pool.QueryRow(ctx, `SELECT id,email,status,email_verified_at,last_login_at,created_at,updated_at,deleted_at,password_hash
		FROM identity.users WHERE id=$1`, userID).Scan(&credential.User.ID, &credential.User.Email, &credential.User.Status,
		&verified, &lastLogin, &credential.User.CreatedAt, &credential.User.UpdatedAt, &deleted, &hash)
	if errors.Is(err, pgx.ErrNoRows) || !hash.Valid {
		return domain.PasswordCredential{}, ports.ErrNotFound
	}
	if err != nil {
		return domain.PasswordCredential{}, fmt.Errorf("get password credential by id: %w", err)
	}
	assignOptionalUserTimes(&credential.User, verified, lastLogin, deleted)
	credential.PasswordHash = hash.String
	return credential, nil
}

func (repository *Repository) ChangePassword(ctx context.Context, userID domain.UserID, currentSessionID, passwordHash string,
	now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password change: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE identity.users SET password_hash=$2,updated_at=$3 WHERE id=$1 AND status='active' AND deleted_at IS NULL`, userID, passwordHash, now)
	if err != nil {
		return fmt.Errorf("change password: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ports.ErrNotFound
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.sessions SET revoked_at=COALESCE(revoked_at,$3)
		WHERE user_id=$1 AND id<>$2::uuid AND revoked_at IS NULL`, userID, currentSessionID, now); err != nil {
		return fmt.Errorf("revoke other password sessions: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE identity.password_reset_tokens SET invalidated_at=$2
		WHERE user_id=$1 AND consumed_at IS NULL AND invalidated_at IS NULL`, userID, now); err != nil {
		return fmt.Errorf("invalidate password reset tokens: %w", err)
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "password change")
}

func (repository *Repository) ListSessions(ctx context.Context, userID domain.UserID, currentSessionID string, now time.Time) ([]domain.ActiveSession, error) {
	rows, err := repository.pool.Query(ctx, `SELECT id,created_at,last_seen_at,expires_at,idle_expires_at,mfa_authenticated_at,authentication_method
		FROM identity.sessions WHERE user_id=$1 AND revoked_at IS NULL AND expires_at>$2 AND idle_expires_at>$2 ORDER BY created_at DESC`, userID, now)
	if err != nil {
		return nil, fmt.Errorf("list active sessions: %w", err)
	}
	defer rows.Close()
	result := []domain.ActiveSession{}
	for rows.Next() {
		var item domain.ActiveSession
		var mfa sql.NullTime
		if err := rows.Scan(&item.ID, &item.CreatedAt, &item.LastSeenAt, &item.ExpiresAt, &item.IdleExpiresAt, &mfa, &item.AuthenticationMethod); err != nil {
			return nil, fmt.Errorf("scan active session: %w", err)
		}
		item.Current = item.ID == currentSessionID
		if mfa.Valid {
			item.MFAAuthenticatedAt = &mfa.Time
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (repository *Repository) RevokeOwnedSession(ctx context.Context, userID domain.UserID, currentSessionID, targetSessionID string,
	now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE identity.sessions SET revoked_at=COALESCE(revoked_at,$3) WHERE id=$2::uuid AND user_id=$1 AND revoked_at IS NULL`, userID, targetSessionID, now)
	if err != nil {
		return fmt.Errorf("revoke owned session: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ports.ErrNotFound
	}
	event.Metadata = map[string]string{"target": "single", "current": fmt.Sprint(targetSessionID == currentSessionID)}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "session revocation")
}

func (repository *Repository) RevokeOtherSessions(ctx context.Context, userID domain.UserID, currentSessionID string,
	now time.Time, event domain.SecurityEvent) (int64, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE identity.sessions SET revoked_at=$3 WHERE user_id=$1 AND id<>$2::uuid AND revoked_at IS NULL`, userID, currentSessionID, now)
	if err != nil {
		return 0, fmt.Errorf("revoke other sessions: %w", err)
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit other session revocation: %w", err)
	}
	return tag.RowsAffected(), nil
}

func (repository *Repository) DeleteAccount(ctx context.Context, userID domain.UserID, sessionID string, now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	statements := []string{
		`DELETE FROM planning.comparison_items WHERE user_id=$1`,
		`DELETE FROM planning.wishlists WHERE user_id=$1`,
		`DELETE FROM planning.setups WHERE user_id=$1`,
		`DELETE FROM recommendation.drafts WHERE user_id=$1`,
		`DELETE FROM planning.profiles WHERE user_id=$1`,
		`DELETE FROM recommendation.session_existing_equipment WHERE session_id IN (SELECT id FROM recommendation.recommendation_sessions WHERE user_id=$1)`,
		`UPDATE recommendation.recommendation_sessions SET user_id=NULL,profile_id=NULL,anonymous_id=NULL,free_text='' WHERE user_id=$1`,
		`UPDATE analytics.events SET user_id=NULL,anonymous_subject_hash=NULL,anonymous_id=NULL WHERE user_id=$1`,
		`UPDATE analytics.identity_claims SET user_id=NULL,status='revoked',revoked_at=$2 WHERE user_id=$1 AND status='claimed'`,
		`UPDATE analytics.consent_history SET user_id=NULL,anonymized_subject_id=gen_random_uuid() WHERE user_id=$1`,
		`DELETE FROM analytics.consent_states WHERE user_id=$1`,
		`UPDATE commerce.affiliate_clicks SET user_id=NULL WHERE user_id=$1`,
		`DELETE FROM identity.user_roles WHERE user_id=$1`,
		`DELETE FROM identity.email_verification_tokens WHERE user_id=$1`,
		`DELETE FROM identity.password_reset_tokens WHERE user_id=$1`,
		`DELETE FROM identity.mfa_credentials WHERE user_id=$1`,
		`UPDATE identity.sessions SET revoked_at=COALESCE(revoked_at,$2) WHERE user_id=$1`,
	}
	for _, statement := range statements {
		arguments := []any{userID}
		if strings.Contains(statement, "$2") {
			arguments = append(arguments, now)
		}
		if _, err := tx.Exec(ctx, statement, arguments...); err != nil {
			return fmt.Errorf("apply account deletion retention: %w", err)
		}
	}
	tag, err := tx.Exec(ctx, `UPDATE identity.users SET email='deleted+'||id::text||'@users.invalid',password_hash=NULL,
		status='deleted',deleted_at=$2,updated_at=$2,email_verified_at=NULL,last_login_at=NULL WHERE id=$1 AND status<>'deleted'`, userID, now)
	if err != nil {
		return fmt.Errorf("anonymize deleted account: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ports.ErrNotFound
	}
	return commit(tx, ctx, "account deletion")
}

func (repository *Repository) ExportAccount(ctx context.Context, userID domain.UserID, now time.Time) (domain.AccountExport, error) {
	result := domain.AccountExport{GeneratedAt: now, Wishlist: []map[string]any{}, Setups: []map[string]any{}, Recommendations: []map[string]any{}, ConsentEvents: []map[string]any{}, AnalyticsEvents: []map[string]any{}}
	if err := scanJSONObject(repository.pool.QueryRow(ctx, `SELECT jsonb_build_object('id',id,'email',email,'status',status,
		'email_verified_at',email_verified_at,'created_at',created_at,'updated_at',updated_at) FROM identity.users WHERE id=$1 AND deleted_at IS NULL`, userID), &result.Account); err != nil {
		return result, err
	}
	_ = scanOptionalJSONObject(repository.pool.QueryRow(ctx, `SELECT to_jsonb(profiles)-'id'-'user_id' FROM planning.profiles WHERE user_id=$1`, userID), &result.Profile)
	if err := scanJSONArray(repository.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object('product_id',product_id,'priority',priority,'note',note,'created_at',created_at) ORDER BY created_at),'[]') FROM planning.wishlists WHERE user_id=$1`, userID), &result.Wishlist); err != nil {
		return result, err
	}
	if err := scanJSONArray(repository.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object(
		'id',setups.id,'name',setups.name,'description',setups.description,'currency',setups.currency,
		'created_at',setups.created_at,'updated_at',setups.updated_at,'items',(SELECT COALESCE(jsonb_agg(
		jsonb_build_object('product_id',items.product_id,'merchant_offer_id',items.merchant_offer_id,
		'quantity',items.quantity,'purchase_status',items.purchase_status,'paid_price_minor',items.paid_price_minor,
		'currency',items.currency,'added_at',items.added_at,'updated_at',items.updated_at) ORDER BY items.added_at),'[]')
		FROM planning.setup_items items WHERE items.setup_id=setups.id)) ORDER BY setups.created_at),'[]')
		FROM planning.setups setups WHERE setups.user_id=$1`, userID), &result.Setups); err != nil {
		return result, err
	}
	if err := scanJSONArray(repository.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object(
		'session_id',sessions.id,'status',sessions.status,'primary_goal',sessions.primary_goal,
		'experience_level',sessions.experience_level,'budget_minor',sessions.budget_minor,'currency',sessions.currency,
		'space_length_mm',sessions.space_length_mm,'space_width_mm',sessions.space_width_mm,
		'space_height_mm',sessions.space_height_mm,'apartment_living',sessions.apartment_living,
		'training_preferences',sessions.training_preferences,'priorities',sessions.priorities,
		'free_text',sessions.free_text,'started_at',sessions.started_at,'completed_at',sessions.completed_at,
		'results',(SELECT COALESCE(jsonb_agg(jsonb_build_object('id',recommendations.id,
		'policy_version',recommendations.policy_version,'engine_version',recommendations.engine_version,
		'objective_score',recommendations.objective_score,'total_price_minor',recommendations.total_price_minor,
		'currency',recommendations.currency,'created_at',recommendations.created_at) ORDER BY recommendations.created_at),'[]')
		FROM recommendation.recommendations recommendations WHERE recommendations.session_id=sessions.id))
		ORDER BY sessions.started_at),'[]') FROM recommendation.recommendation_sessions sessions WHERE sessions.user_id=$1`, userID), &result.Recommendations); err != nil {
		return result, err
	}
	if err := scanJSONArray(repository.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object(
		'state',state,'policy_version',policy_version,'source',source,'decided_at',decided_at) ORDER BY decided_at),'[]')
		FROM analytics.consent_history WHERE user_id=$1`, userID), &result.ConsentEvents); err != nil {
		return result, err
	}
	if err := scanJSONArray(repository.pool.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(jsonb_build_object(
		'event_id',public_event_id,'name',event_name,'surface',surface,'properties',properties,
		'page_path',page_path,'occurred_at',occurred_at,'consent_policy_version',consent_policy_version)
		ORDER BY occurred_at),'[]') FROM analytics.events WHERE user_id=$1 AND origin='client'`, userID), &result.AnalyticsEvents); err != nil {
		return result, err
	}
	result.SecurityMetadata = map[string]any{"active_session_count": 0, "mfa_enabled": false}
	var sessionCount int64
	var mfaEnabled bool
	_ = repository.pool.QueryRow(ctx, `SELECT count(*) FROM identity.sessions WHERE user_id=$1 AND revoked_at IS NULL AND expires_at>$2 AND idle_expires_at>$2`, userID, now).Scan(&sessionCount)
	_ = repository.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM identity.mfa_credentials WHERE user_id=$1 AND status='enabled')`, userID).Scan(&mfaEnabled)
	result.SecurityMetadata["active_session_count"] = sessionCount
	result.SecurityMetadata["mfa_enabled"] = mfaEnabled
	return result, nil
}

func (repository *Repository) RecordSecurityEvent(ctx context.Context, event domain.SecurityEvent) error {
	return insertSecurityEvent(ctx, repository.pool, event)
}

func (repository *Repository) UpsertPendingMFA(ctx context.Context, userID domain.UserID, ciphertext, nonce []byte,
	version int16, now time.Time, event domain.SecurityEvent) (domain.MFACredential, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.MFACredential{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var value domain.MFACredential
	err = tx.QueryRow(ctx, `INSERT INTO identity.mfa_credentials (user_id,secret_ciphertext,secret_nonce,key_version,status,created_at)
		VALUES ($1,$2,$3,$4,'pending',$5) ON CONFLICT (user_id) DO UPDATE SET secret_ciphertext=EXCLUDED.secret_ciphertext,
		secret_nonce=EXCLUDED.secret_nonce,key_version=EXCLUDED.key_version,status='pending',created_at=EXCLUDED.created_at,verified_at=NULL,disabled_at=NULL
		RETURNING id,user_id,secret_ciphertext,secret_nonce,key_version,status,created_at`, userID, ciphertext, nonce, version, now).
		Scan(&value.ID, &value.UserID, &value.SecretCiphertext, &value.SecretNonce, &value.KeyVersion, &value.Status, &value.CreatedAt)
	if err != nil {
		return value, fmt.Errorf("upsert pending MFA: %w", err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM identity.mfa_recovery_codes WHERE credential_id=$1`, value.ID); err != nil {
		return value, err
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return value, err
	}
	if err := tx.Commit(ctx); err != nil {
		return value, err
	}
	return value, nil
}

func (repository *Repository) GetMFA(ctx context.Context, userID domain.UserID) (domain.MFACredential, error) {
	var value domain.MFACredential
	var verified sql.NullTime
	err := repository.pool.QueryRow(ctx, `SELECT id,user_id,secret_ciphertext,secret_nonce,key_version,status,created_at,verified_at FROM identity.mfa_credentials WHERE user_id=$1`, userID).
		Scan(&value.ID, &value.UserID, &value.SecretCiphertext, &value.SecretNonce, &value.KeyVersion, &value.Status, &value.CreatedAt, &verified)
	if errors.Is(err, pgx.ErrNoRows) {
		return value, ports.ErrNotFound
	}
	if err != nil {
		return value, fmt.Errorf("get MFA: %w", err)
	}
	if verified.Valid {
		value.VerifiedAt = &verified.Time
	}
	return value, nil
}

func (repository *Repository) EnableMFA(ctx context.Context, userID domain.UserID, hashes [][]byte, now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id string
	if err := tx.QueryRow(ctx, `UPDATE identity.mfa_credentials SET status='enabled',verified_at=$2,disabled_at=NULL WHERE user_id=$1 AND status='pending' RETURNING id`, userID, now).Scan(&id); errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrConflict
	} else if err != nil {
		return err
	}
	if err := insertRecoveryHashes(ctx, tx, id, hashes, now); err != nil {
		return err
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "enable MFA")
}

func (repository *Repository) ReplaceRecoveryCodes(ctx context.Context, userID domain.UserID, hashes [][]byte, now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id string
	if err := tx.QueryRow(ctx, `SELECT id FROM identity.mfa_credentials WHERE user_id=$1 AND status='enabled' FOR UPDATE`, userID).Scan(&id); errors.Is(err, pgx.ErrNoRows) {
		return ports.ErrNotFound
	} else if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM identity.mfa_recovery_codes WHERE credential_id=$1`, id); err != nil {
		return err
	}
	if err := insertRecoveryHashes(ctx, tx, id, hashes, now); err != nil {
		return err
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "replace recovery codes")
}

func insertRecoveryHashes(ctx context.Context, tx pgx.Tx, credentialID string, hashes [][]byte, now time.Time) error {
	for _, hash := range hashes {
		if _, err := tx.Exec(ctx, `INSERT INTO identity.mfa_recovery_codes (credential_id,code_hash,created_at) VALUES ($1,$2,$3)`, credentialID, hash, now); err != nil {
			return fmt.Errorf("insert recovery code hash: %w", err)
		}
	}
	return nil
}

func (repository *Repository) ConsumeRecoveryCode(ctx context.Context, userID domain.UserID, hash []byte, now time.Time, event domain.SecurityEvent) (bool, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE identity.mfa_recovery_codes codes SET consumed_at=$3 FROM identity.mfa_credentials credentials WHERE codes.credential_id=credentials.id AND credentials.user_id=$1 AND credentials.status='enabled' AND codes.code_hash=$2 AND codes.consumed_at IS NULL`, userID, hash, now)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() == 0 {
		return false, nil
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (repository *Repository) CreateMFAChallenge(ctx context.Context, challenge domain.MFAChallenge, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `INSERT INTO identity.mfa_login_challenges (user_id,token_hash,expires_at,created_at) VALUES ($1,$2,$3,$4)`, challenge.UserID, challenge.TokenHash, challenge.ExpiresAt, challenge.CreatedAt); err != nil {
		return err
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "create MFA challenge")
}

func (repository *Repository) GetMFAChallenge(ctx context.Context, hash []byte, now time.Time) (domain.MFAChallenge, error) {
	var value domain.MFAChallenge
	var consumed sql.NullTime
	var attempts int16
	err := repository.pool.QueryRow(ctx, `SELECT user_id,token_hash,expires_at,created_at,consumed_at,attempts FROM identity.mfa_login_challenges WHERE token_hash=$1`, hash).Scan(&value.UserID, &value.TokenHash, &value.ExpiresAt, &value.CreatedAt, &consumed, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return value, ports.ErrNotFound
	}
	if err != nil {
		return value, err
	}
	if consumed.Valid {
		return value, ports.ErrConsumed
	}
	if !value.ExpiresAt.After(now) {
		return value, ports.ErrExpired
	}
	if attempts >= 5 {
		return value, ports.ErrAttemptsExceeded
	}
	return value, nil
}

func (repository *Repository) FailMFAChallenge(ctx context.Context, hash []byte, now time.Time, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `UPDATE identity.mfa_login_challenges SET attempts=LEAST(5,attempts+1) WHERE token_hash=$1 AND consumed_at IS NULL`, hash); err != nil {
		return err
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "fail MFA challenge")
}

func (repository *Repository) ConsumeMFAChallengeAndCreateSession(ctx context.Context, hash []byte, session domain.Session, now time.Time, event domain.SecurityEvent) (domain.User, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.User{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var user domain.User
	var challengeUser domain.UserID
	err = tx.QueryRow(ctx, `UPDATE identity.mfa_login_challenges SET consumed_at=$2 WHERE token_hash=$1 AND consumed_at IS NULL AND expires_at>$2 AND attempts<5 RETURNING user_id`, hash, now).Scan(&challengeUser)
	if errors.Is(err, pgx.ErrNoRows) {
		return user, ports.ErrNotFound
	}
	if err != nil {
		return user, err
	}
	if challengeUser != session.UserID {
		return user, ports.ErrConflict
	}
	if err := tx.QueryRow(ctx, `INSERT INTO identity.sessions (user_id,token_hash,expires_at,idle_expires_at,last_seen_at,created_at,mfa_authenticated_at,authentication_method) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, challengeUser, session.TokenHash, session.ExpiresAt, session.IdleExpiresAt, session.LastSeenAt, session.CreatedAt, session.MFAAuthenticatedAt, session.AuthenticationMethod).Scan(&session.ID); err != nil {
		return user, fmt.Errorf("create MFA session: %w", err)
	}
	var verified, lastLogin, deleted sql.NullTime
	err = tx.QueryRow(ctx, `UPDATE identity.users SET last_login_at=$2 WHERE id=$1 AND status='active' AND deleted_at IS NULL RETURNING id,email,status,email_verified_at,last_login_at,created_at,updated_at,deleted_at`, challengeUser, now).Scan(&user.ID, &user.Email, &user.Status, &verified, &lastLogin, &user.CreatedAt, &user.UpdatedAt, &deleted)
	if err != nil {
		return user, err
	}
	assignOptionalUserTimes(&user, verified, lastLogin, deleted)
	if err := loadRolesTx(ctx, tx, &user); err != nil {
		return user, err
	}
	event.SessionID = &session.ID
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return user, err
	}
	if err := tx.Commit(ctx); err != nil {
		return user, err
	}
	return user, nil
}

func (repository *Repository) MarkSessionMFA(ctx context.Context, userID domain.UserID, sessionID string, now time.Time, method string, event domain.SecurityEvent) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `UPDATE identity.sessions SET mfa_authenticated_at=$3,authentication_method=$4,last_seen_at=GREATEST(last_seen_at,$3) WHERE id=$2::uuid AND user_id=$1 AND revoked_at IS NULL AND expires_at>$3 AND idle_expires_at>$3`, userID, sessionID, now, method)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return ports.ErrNotFound
	}
	if err := insertSecurityEvent(ctx, tx, event); err != nil {
		return err
	}
	return commit(tx, ctx, "mark MFA step-up")
}

func (repository *Repository) CleanupExpiredSecurityArtifacts(ctx context.Context, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	type cleanupStatement struct {
		query string
		args  []any
	}
	statements := []cleanupStatement{
		{`DELETE FROM identity.email_verification_tokens WHERE expires_at<$1 OR consumed_at<$2 OR invalidated_at<$2`, []any{now, now.Add(-7 * 24 * time.Hour)}},
		{`DELETE FROM identity.password_reset_tokens WHERE expires_at<$1 OR consumed_at<$2 OR invalidated_at<$2`, []any{now, now.Add(-7 * 24 * time.Hour)}},
		{`DELETE FROM identity.mfa_login_challenges WHERE expires_at<$1 OR consumed_at<$2`, []any{now, now.Add(-24 * time.Hour)}},
		{`DELETE FROM identity.sessions WHERE expires_at<$1 OR revoked_at<$1`, []any{now.Add(-30 * 24 * time.Hour)}},
	}
	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement.query, statement.args...); err != nil {
			return fmt.Errorf("clean security artifacts: %w", err)
		}
	}
	return commit(tx, ctx, "security cleanup")
}

type rowScanner interface{ Scan(...any) error }

func scanJSONObject(row rowScanner, target *map[string]any) error {
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
func scanOptionalJSONObject(row rowScanner, target *map[string]any) error {
	var raw []byte
	if err := row.Scan(&raw); errors.Is(err, pgx.ErrNoRows) {
		return nil
	} else if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}
func scanJSONArray(row rowScanner, target *[]map[string]any) error {
	var raw []byte
	if err := row.Scan(&raw); err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

type securityEventExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertSecurityEvent(ctx context.Context, executor securityEventExecutor, event domain.SecurityEvent) error {
	metadata := event.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	surface := event.Surface
	if surface == "" {
		surface = "api"
	}
	_, err = executor.Exec(ctx, `INSERT INTO identity.security_events (user_id,session_id,event_type,outcome,request_id,surface,metadata,occurred_at) VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8)`, event.UserID, event.SessionID, event.Type, event.Outcome, event.RequestID, surface, encoded, event.OccurredAt)
	if err != nil {
		return fmt.Errorf("record security event: %w", err)
	}
	return nil
}
func commit(tx pgx.Tx, ctx context.Context, operation string) error {
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit %s: %w", operation, err)
	}
	return nil
}
func loadRolesTx(ctx context.Context, tx pgx.Tx, user *domain.User) error {
	return tx.QueryRow(ctx, `SELECT ARRAY(SELECT role_key FROM identity.user_roles WHERE user_id=$1 ORDER BY role_key)`, user.ID).Scan(&user.Roles)
}
