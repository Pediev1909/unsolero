package analyticspostgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"rigmark/internal/modules/analytics/domain"
	"rigmark/internal/modules/analytics/ports"
)

type Repository struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (repository *Repository) Ingest(ctx context.Context, event domain.Event, receiptRetention time.Duration) (domain.IngestionResult, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.IngestionResult{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if ok, consentErr := consentGranted(ctx, tx, event); consentErr != nil {
		return domain.IngestionResult{}, consentErr
	} else if !ok {
		if err := insertReceipt(ctx, tx, event.ID, event.Name, domain.IngestionRejected, "consent_required", receiptRetention); err != nil {
			return domain.IngestionResult{}, err
		}
		if err := advanceCoverage(ctx, tx); err != nil {
			return domain.IngestionResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return domain.IngestionResult{}, err
		}
		return domain.IngestionResult{Outcome: domain.IngestionRejected}, nil
	}
	properties, err := json.Marshal(event.Properties)
	if err != nil {
		return domain.IngestionResult{}, fmt.Errorf("encode analytics properties: %w", err)
	}
	tag, err := tx.Exec(ctx, `INSERT INTO analytics.events (
		public_event_id,event_name,schema_version,user_id,recommendation_session_id,
		anonymous_subject_hash,session_id,request_id,surface,properties,page_path,
		traffic_source,traffic_medium,campaign,referrer_host,consent_state,
		consent_policy_version,origin,classification,is_reportable,retention_expires_at,occurred_at)
		VALUES ($1,$2,$3,$4::uuid,$5::uuid,$6,$7,$8,$9,$10::jsonb,$11,$12,$13,$14,$15,
			$16,$17,$18,$19,$20,$21,$22)
		ON CONFLICT (public_event_id) DO NOTHING`, event.ID, event.Name, event.SchemaVersion,
		event.UserID, event.RecommendationSessionID, nullableBytes(event.AnonymousSubjectHash),
		event.SessionID, event.RequestID, event.Surface, properties, event.PagePath,
		event.TrafficSource, event.TrafficMedium, event.Campaign, event.ReferrerHost,
		event.ConsentState, event.ConsentPolicyVersion, event.Origin, event.Classification,
		event.Reportable, event.RetentionExpiresAt, event.OccurredAt)
	if err != nil {
		return domain.IngestionResult{}, fmt.Errorf("record analytics event: %w", err)
	}
	outcome, reason := domain.IngestionAccepted, "accepted"
	if tag.RowsAffected() == 0 {
		outcome, reason = domain.IngestionDeduplicated, "duplicate_event"
	}
	if err := insertReceipt(ctx, tx, event.ID, event.Name, outcome, reason, receiptRetention); err != nil {
		return domain.IngestionResult{}, err
	}
	if err := advanceCoverage(ctx, tx); err != nil {
		return domain.IngestionResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.IngestionResult{}, fmt.Errorf("commit analytics event: %w", err)
	}
	return domain.IngestionResult{Outcome: outcome}, nil
}

func consentGranted(ctx context.Context, tx pgx.Tx, event domain.Event) (bool, error) {
	var granted bool
	var err error
	if event.UserID != nil {
		err = tx.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM analytics.consent_states consent
			JOIN identity.users users ON users.id=consent.user_id
			WHERE consent.user_id=$1::uuid AND consent.state='granted'
				AND consent.policy_version=$2 AND users.deleted_at IS NULL
			FOR SHARE OF consent)`, *event.UserID, event.ConsentPolicyVersion).Scan(&granted)
	} else {
		err = tx.QueryRow(ctx, `SELECT EXISTS(
			SELECT 1 FROM analytics.consent_states consent
			WHERE consent.anonymous_subject_hash=$1 AND consent.state='granted'
				AND consent.policy_version=$2 FOR SHARE OF consent)`, event.AnonymousSubjectHash, event.ConsentPolicyVersion).Scan(&granted)
	}
	return granted, err
}

func (repository *Repository) RecordRejected(ctx context.Context, eventID domain.EventID, eventName string, outcome domain.IngestionOutcome, reason string, retention time.Duration) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := insertReceipt(ctx, tx, eventID, eventName, outcome, reason, retention); err != nil {
		return err
	}
	if err := advanceCoverage(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func insertReceipt(ctx context.Context, tx pgx.Tx, eventID domain.EventID, eventName string, outcome domain.IngestionOutcome, reason string, retention time.Duration) error {
	_, err := tx.Exec(ctx, `INSERT INTO analytics.ingestion_receipts
		(public_event_id,event_name,outcome,reason_code,retention_expires_at)
		VALUES ($1::uuid,NULLIF($2,''),$3,$4,now()+($5 * interval '1 second'))`, eventID, eventName, outcome, reason, int64(retention.Seconds()))
	if err != nil {
		return fmt.Errorf("record analytics receipt: %w", err)
	}
	return nil
}

func advanceCoverage(ctx context.Context, tx pgx.Tx) error {
	_, err := tx.Exec(ctx, `UPDATE analytics.reporting_coverage SET complete_through=GREATEST(complete_through,now()),updated_at=now()
		WHERE pipeline_key='first_party_events_v3'`)
	return err
}

func (repository *Repository) SetConsent(ctx context.Context, decision domain.ConsentDecision) (domain.Consent, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.Consent{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	current, found, err := getConsentForUpdate(ctx, tx, decision.Subject)
	if err != nil {
		return domain.Consent{}, err
	}
	state := decision.RequestedState
	if state == "denied" && found && (current.State == "granted" || current.State == "withdrawn") {
		state = "withdrawn"
	}
	result := domain.Consent{State: state, PolicyVersion: decision.PolicyVersion, Source: decision.Source, DecidedAt: decision.DecidedAt}
	if found && current.State == result.State && current.PolicyVersion == result.PolicyVersion && current.Source == result.Source {
		return current, tx.Commit(ctx)
	}
	if decision.Subject.UserID != nil {
		_, err = tx.Exec(ctx, `INSERT INTO analytics.consent_states (user_id,state,policy_version,source,decided_at,updated_at)
			VALUES ($1::uuid,$2,$3,$4,$5,$5) ON CONFLICT (user_id) WHERE user_id IS NOT NULL
			DO UPDATE SET state=EXCLUDED.state,policy_version=EXCLUDED.policy_version,source=EXCLUDED.source,
				decided_at=EXCLUDED.decided_at,updated_at=EXCLUDED.updated_at`, *decision.Subject.UserID, state, decision.PolicyVersion, decision.Source, decision.DecidedAt)
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO analytics.consent_history (user_id,state,policy_version,source,decided_at)
			VALUES ($1::uuid,$2,$3,$4,$5)`, *decision.Subject.UserID, state, decision.PolicyVersion, decision.Source, decision.DecidedAt)
		}
	} else {
		_, err = tx.Exec(ctx, `INSERT INTO analytics.consent_states (anonymous_subject_hash,state,policy_version,source,decided_at,updated_at)
			VALUES ($1,$2,$3,$4,$5,$5) ON CONFLICT (anonymous_subject_hash) WHERE anonymous_subject_hash IS NOT NULL
			DO UPDATE SET state=EXCLUDED.state,policy_version=EXCLUDED.policy_version,source=EXCLUDED.source,
				decided_at=EXCLUDED.decided_at,updated_at=EXCLUDED.updated_at`, decision.Subject.AnonymousSubjectHash, state, decision.PolicyVersion, decision.Source, decision.DecidedAt)
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO analytics.consent_history (anonymous_subject_hash,state,policy_version,source,decided_at)
			VALUES ($1,$2,$3,$4,$5)`, decision.Subject.AnonymousSubjectHash, state, decision.PolicyVersion, decision.Source, decision.DecidedAt)
		}
	}
	if err != nil {
		return domain.Consent{}, fmt.Errorf("persist analytics consent: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Consent{}, err
	}
	return result, nil
}

func getConsentForUpdate(ctx context.Context, tx pgx.Tx, subject domain.Subject) (domain.Consent, bool, error) {
	var value domain.Consent
	var err error
	if subject.UserID != nil {
		err = tx.QueryRow(ctx, `SELECT state,policy_version,source,decided_at FROM analytics.consent_states WHERE user_id=$1::uuid FOR UPDATE`, *subject.UserID).
			Scan(&value.State, &value.PolicyVersion, &value.Source, &value.DecidedAt)
	} else {
		err = tx.QueryRow(ctx, `SELECT state,policy_version,source,decided_at FROM analytics.consent_states WHERE anonymous_subject_hash=$1 FOR UPDATE`, subject.AnonymousSubjectHash).
			Scan(&value.State, &value.PolicyVersion, &value.Source, &value.DecidedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Consent{}, false, nil
	}
	return value, err == nil, err
}

func (repository *Repository) GetConsent(ctx context.Context, subject domain.Subject) (domain.Consent, error) {
	var value domain.Consent
	var err error
	if subject.UserID != nil {
		err = repository.pool.QueryRow(ctx, `SELECT state,policy_version,source,decided_at FROM analytics.consent_states WHERE user_id=$1::uuid`, *subject.UserID).
			Scan(&value.State, &value.PolicyVersion, &value.Source, &value.DecidedAt)
	} else {
		err = repository.pool.QueryRow(ctx, `SELECT state,policy_version,source,decided_at FROM analytics.consent_states WHERE anonymous_subject_hash=$1`, subject.AnonymousSubjectHash).
			Scan(&value.State, &value.PolicyVersion, &value.Source, &value.DecidedAt)
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Consent{}, ports.ErrConsentNotFound
	}
	return value, err
}

func (repository *Repository) ClaimIdentity(ctx context.Context, subjectHash []byte, userID, policyVersion string, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var existingUser sql.NullString
	var status string
	err = tx.QueryRow(ctx, `SELECT user_id::text,status FROM analytics.identity_claims WHERE anonymous_subject_hash=$1 FOR UPDATE`, subjectHash).Scan(&existingUser, &status)
	if err == nil {
		if status == "claimed" && existingUser.Valid && existingUser.String == userID {
			return tx.Commit(ctx)
		}
		return ports.ErrIdentityClaimConflict
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	var allowed bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM analytics.consent_states anonymous_consent
		JOIN analytics.consent_states account_consent ON account_consent.user_id=$2::uuid
		JOIN identity.users users ON users.id=$2::uuid AND users.deleted_at IS NULL
		WHERE anonymous_consent.anonymous_subject_hash=$1
			AND anonymous_consent.state='granted' AND account_consent.state='granted'
			AND anonymous_consent.policy_version=$3 AND account_consent.policy_version=$3
		FOR SHARE OF anonymous_consent,account_consent)`, subjectHash, userID, policyVersion).Scan(&allowed); err != nil {
		return err
	}
	if !allowed {
		return ports.ErrIdentityClaimNotAllowed
	}
	if _, err := tx.Exec(ctx, `INSERT INTO analytics.identity_claims
		(anonymous_subject_hash,user_id,status,consent_policy_version,claimed_at)
		VALUES ($1,$2::uuid,'claimed',$3,$4)`, subjectHash, userID, policyVersion, now); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE analytics.events SET user_id=$2::uuid,anonymous_subject_hash=NULL,anonymous_id=NULL
		WHERE anonymous_subject_hash=$1 AND user_id IS NULL`, subjectHash, userID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (repository *Repository) Cleanup(ctx context.Context, now time.Time, batch int) (domain.CleanupResult, error) {
	var result domain.CleanupResult
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := tx.QueryRow(ctx, `WITH expired AS (
		SELECT id FROM analytics.events WHERE retention_expires_at <= $1 ORDER BY retention_expires_at,id LIMIT $2 FOR UPDATE SKIP LOCKED
	), deleted AS (DELETE FROM analytics.events WHERE id IN (SELECT id FROM expired) RETURNING 1)
	SELECT count(*) FROM deleted`, now, batch).Scan(&result.EventsDeleted); err != nil {
		return result, err
	}
	if err := tx.QueryRow(ctx, `WITH expired AS (
		SELECT id FROM analytics.ingestion_receipts WHERE retention_expires_at <= $1 ORDER BY retention_expires_at,id LIMIT $2 FOR UPDATE SKIP LOCKED
	), deleted AS (DELETE FROM analytics.ingestion_receipts WHERE id IN (SELECT id FROM expired) RETURNING 1)
	SELECT count(*) FROM deleted`, now, batch).Scan(&result.ReceiptsDeleted); err != nil {
		return result, err
	}
	return result, tx.Commit(ctx)
}

func nullableBytes(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

var _ ports.EventRecorder = (*Repository)(nil)
var _ ports.ReportingRepository = (*Repository)(nil)
