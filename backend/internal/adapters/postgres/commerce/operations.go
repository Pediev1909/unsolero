package commercepostgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
	identity "rigmark/internal/modules/identity/domain"
)

func (repository *Repository) CreateProviderConfiguration(ctx context.Context, actor identity.UserID, input domain.ProviderConfigurationInput) (domain.ProviderConfiguration, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("begin provider configuration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id domain.ProviderConfigurationID
	err = tx.QueryRow(ctx, `INSERT INTO commerce.provider_configurations
		(merchant_id, provider_key, adapter_key, external_merchant_id, credential_reference,
		 schedule_interval_minutes, freshness_ttl_minutes, lifecycle_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'disabled') RETURNING id`, input.MerchantID,
		input.ProviderKey, input.AdapterKey, input.ExternalMerchantID, input.CredentialReference,
		input.ScheduleIntervalMinutes, input.FreshnessTTLMinutes).Scan(&id)
	if err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("create provider configuration: %w", err)
	}
	if err = insertOperationAudit(ctx, tx, actor, "provider.create", "provider_configuration", string(id), nil); err != nil {
		return domain.ProviderConfiguration{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("commit provider configuration: %w", err)
	}
	return repository.GetProviderConfiguration(ctx, id)
}

func (repository *Repository) ListProviderConfigurations(ctx context.Context) ([]domain.ProviderConfiguration, error) {
	rows, err := repository.pool.Query(ctx, providerConfigurationSelect+` ORDER BY configurations.provider_key, merchants.name`)
	if err != nil {
		return nil, fmt.Errorf("list provider configurations: %w", err)
	}
	defer rows.Close()
	configurations := make([]domain.ProviderConfiguration, 0)
	for rows.Next() {
		configuration, scanErr := scanProviderConfiguration(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		configurations = append(configurations, configuration)
	}
	return configurations, rows.Err()
}

func (repository *Repository) GetProviderConfiguration(ctx context.Context, id domain.ProviderConfigurationID) (domain.ProviderConfiguration, error) {
	configuration, err := scanProviderConfiguration(repository.pool.QueryRow(ctx, providerConfigurationSelect+` WHERE configurations.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ProviderConfiguration{}, ports.ErrImportNotFound
	}
	return configuration, err
}

func (repository *Repository) SetProviderLifecycle(ctx context.Context, actor identity.UserID, id domain.ProviderConfigurationID, status domain.ProviderLifecycle, verified bool) (domain.ProviderConfiguration, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("begin lifecycle update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	command, err := tx.Exec(ctx, `UPDATE commerce.provider_configurations SET lifecycle_status=$2,
		configuration_verified_at=CASE WHEN $3 THEN now() ELSE configuration_verified_at END,
		next_import_at=CASE WHEN $2 IN ('active','degraded') THEN COALESCE(next_import_at, now()) ELSE NULL END,
		updated_at=now() WHERE id=$1`, id, status, verified)
	if err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("update provider lifecycle: %w", err)
	}
	if command.RowsAffected() == 0 {
		return domain.ProviderConfiguration{}, ports.ErrImportNotFound
	}
	if err = insertOperationAudit(ctx, tx, actor, "provider.lifecycle", "provider_configuration", string(id), map[string]any{"status": status}); err != nil {
		return domain.ProviderConfiguration{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("commit provider lifecycle: %w", err)
	}
	return repository.GetProviderConfiguration(ctx, id)
}

func (repository *Repository) QueueImport(ctx context.Context, actor *identity.UserID, id domain.ProviderConfigurationID, trigger domain.ImportTrigger, key string, attempts int16) (domain.ImportRun, error) {
	var actorValue any
	if actor != nil {
		actorValue = *actor
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ImportRun{}, fmt.Errorf("begin queueing commerce import: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var runID domain.ImportRunID
	var created bool
	err = tx.QueryRow(ctx, `INSERT INTO commerce.offer_import_runs
		(provider_configuration_id, trigger_type, status, idempotency_key, requested_by,
		 cursor_before, max_attempts)
		SELECT id, $2, 'queued', $3, $4::uuid, cursor_value, $5
		FROM commerce.provider_configurations
		WHERE id=$1 AND lifecycle_status IN ('active','degraded')
		ON CONFLICT (provider_configuration_id, idempotency_key)
		DO UPDATE SET idempotency_key=EXCLUDED.idempotency_key
		RETURNING id, (xmax=0)`, id, trigger, key, actorValue, attempts).Scan(&runID, &created)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ImportRun{}, ports.ErrProviderDisabled
	}
	if err != nil {
		return domain.ImportRun{}, fmt.Errorf("queue commerce import: %w", err)
	}
	if created && actor != nil {
		if err = insertOperationAudit(ctx, tx, *actor, "import."+string(trigger)+"_queued", "offer_import_run", string(runID),
			map[string]any{"provider_configuration_id": id}); err != nil {
			return domain.ImportRun{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ImportRun{}, fmt.Errorf("commit queued commerce import: %w", err)
	}
	return repository.getImport(ctx, runID)
}

func (repository *Repository) QueueDueImports(ctx context.Context, now time.Time, limit int) (int, error) {
	command, err := repository.pool.Exec(ctx, `WITH due AS (
		SELECT id, next_import_at FROM commerce.provider_configurations
		WHERE lifecycle_status IN ('active','degraded') AND next_import_at <= $1
		ORDER BY next_import_at, id LIMIT $2 FOR UPDATE SKIP LOCKED
	), queued AS (
		INSERT INTO commerce.offer_import_runs
		(provider_configuration_id, trigger_type, status, idempotency_key, cursor_before)
		SELECT configurations.id, 'scheduled', 'queued',
			'scheduled-' || extract(epoch FROM configurations.next_import_at)::bigint,
			configurations.cursor_value
		FROM commerce.provider_configurations configurations JOIN due USING (id)
		ON CONFLICT (provider_configuration_id, idempotency_key) DO NOTHING RETURNING 1
	)
	UPDATE commerce.provider_configurations configurations
	SET next_import_at=$1 + make_interval(mins => configurations.schedule_interval_minutes), updated_at=$1
	FROM due WHERE configurations.id=due.id`, now, limit)
	if err != nil {
		return 0, fmt.Errorf("queue due commerce imports: %w", err)
	}
	return int(command.RowsAffected()), nil
}

func (repository *Repository) RecoverStalledImports(ctx context.Context, staleBefore, recoveredAt time.Time, limit int) (int, error) {
	var recovered int
	err := repository.pool.QueryRow(ctx, `WITH candidates AS (
		SELECT id, provider_configuration_id, attempt_count, max_attempts
		FROM commerce.offer_import_runs
		WHERE status='running' AND updated_at < $1
		ORDER BY updated_at, id LIMIT $3 FOR UPDATE SKIP LOCKED
	), recovered AS (
		UPDATE commerce.offer_import_runs runs SET
			status=CASE WHEN candidates.attempt_count < candidates.max_attempts THEN 'retry_wait' ELSE 'failed' END,
			next_retry_at=CASE WHEN candidates.attempt_count < candidates.max_attempts THEN $2::timestamptz ELSE NULL::timestamptz END,
			completed_at=CASE WHEN candidates.attempt_count < candidates.max_attempts THEN NULL::timestamptz ELSE $2::timestamptz END,
			error_code='worker.lease_expired',
			error_message='The import worker lease expired before completion.', updated_at=$2
		FROM candidates WHERE runs.id=candidates.id
		RETURNING runs.provider_configuration_id
	), failure_counts AS (
		SELECT provider_configuration_id, count(*)::integer AS failures
		FROM recovered GROUP BY provider_configuration_id
	), providers AS (
		UPDATE commerce.provider_configurations configurations SET
			last_import_failed_at=$2,
			consecutive_failures=configurations.consecutive_failures+failure_counts.failures,
			last_error_code='worker.lease_expired',
			lifecycle_status=CASE WHEN lifecycle_status='active' THEN 'degraded' ELSE lifecycle_status END,
			updated_at=$2
		FROM failure_counts WHERE configurations.id=failure_counts.provider_configuration_id
		RETURNING configurations.id
	) SELECT count(*) FROM recovered`, staleBefore, recoveredAt, limit).Scan(&recovered)
	if err != nil {
		return 0, fmt.Errorf("recover stalled commerce imports: %w", err)
	}
	return recovered, nil
}

func (repository *Repository) ClaimNextImport(ctx context.Context, now time.Time) (domain.ImportRun, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ImportRun{}, fmt.Errorf("begin import claim: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var runID domain.ImportRunID
	err = tx.QueryRow(ctx, `WITH candidate AS (
		SELECT runs.id FROM commerce.offer_import_runs runs
		JOIN commerce.provider_configurations configurations ON configurations.id=runs.provider_configuration_id
		WHERE runs.status IN ('queued','retry_wait')
		  AND (runs.next_retry_at IS NULL OR runs.next_retry_at <= $1)
		  AND runs.attempt_count < runs.max_attempts
		  AND configurations.lifecycle_status IN ('active','degraded')
		ORDER BY COALESCE(runs.next_retry_at, runs.created_at), runs.id
		LIMIT 1 FOR UPDATE OF runs SKIP LOCKED
	) UPDATE commerce.offer_import_runs runs SET status='running', attempt_count=attempt_count+1,
		started_at=COALESCE(started_at,$1), next_retry_at=NULL, updated_at=$1
	FROM candidate WHERE runs.id=candidate.id RETURNING runs.id`, now).Scan(&runID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ImportRun{}, ports.ErrImportNotFound
	}
	if err != nil {
		return domain.ImportRun{}, fmt.Errorf("claim commerce import: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE commerce.provider_configurations SET last_import_started_at=$2, updated_at=$2
		WHERE id=(SELECT provider_configuration_id FROM commerce.offer_import_runs WHERE id=$1)`, runID, now); err != nil {
		return domain.ImportRun{}, fmt.Errorf("mark provider import started: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ImportRun{}, fmt.Errorf("commit import claim: %w", err)
	}
	return repository.getImport(ctx, runID)
}

func (repository *Repository) ApplyImport(ctx context.Context, run domain.ImportRun, records []domain.ValidatedOffer, failures []domain.ImportRecordFailure, batch domain.ProviderBatch, importedAt time.Time) (domain.ImportApplyResult, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ImportApplyResult{}, fmt.Errorf("begin applying import: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status domain.ImportStatus
	var configurationID domain.ProviderConfigurationID
	if err = tx.QueryRow(ctx, `SELECT status, provider_configuration_id FROM commerce.offer_import_runs
		WHERE id=$1 FOR UPDATE`, run.ID).Scan(&status, &configurationID); err != nil {
		return domain.ImportApplyResult{}, fmt.Errorf("lock import run: %w", err)
	}
	if status != domain.ImportRunning || configurationID != run.ProviderConfiguration.ID {
		return domain.ImportApplyResult{}, ports.ErrImportConflict
	}
	result := domain.ImportApplyResult{Received: len(records) + len(failures), Rejected: len(failures),
		Failures: failures, NextCursor: batch.NextCursor, Complete: batch.Complete}
	for _, failure := range failures {
		_, err = tx.Exec(ctx, `INSERT INTO commerce.offer_import_failures
			(import_run_id, external_record_id, error_code, error_message, record_fingerprint)
			VALUES ($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, run.ID, failure.ExternalRecordID,
			failure.Code, failure.Message, failure.RecordFingerprint)
		if err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("record rejected provider offer: %w", err)
		}
	}
	for _, record := range records {
		var productExists bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM catalog.products WHERE id=$1)`, record.ProductID).Scan(&productExists); err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("validate imported product: %w", err)
		}
		if !productExists {
			message := "The canonical product ID does not exist."
			_, err = tx.Exec(ctx, `INSERT INTO commerce.offer_import_failures
				(import_run_id, external_record_id, error_code, error_message, record_fingerprint)
				VALUES ($1,$2,'record.product_not_found',$3,$4) ON CONFLICT DO NOTHING`, run.ID,
				record.ExternalOfferID, message, record.PriceFingerprint)
			if err != nil {
				return domain.ImportApplyResult{}, fmt.Errorf("record missing product: %w", err)
			}
			result.Rejected++
			continue
		}
		var offerID domain.OfferID
		err = tx.QueryRow(ctx, `INSERT INTO commerce.merchant_offers
			(merchant_id, product_id, merchant_sku, product_url, price_minor, shipping_minor,
			 currency, availability, condition, last_checked_at, is_active,
			 provider_observed_at, imported_at, expires_at, last_import_run_id)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,true,$11,$12,$13,$14)
			ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
			 product_id=EXCLUDED.product_id, product_url=EXCLUDED.product_url,
			 price_minor=EXCLUDED.price_minor, shipping_minor=EXCLUDED.shipping_minor,
			 currency=EXCLUDED.currency, availability=EXCLUDED.availability,
			 condition=EXCLUDED.condition, last_checked_at=EXCLUDED.last_checked_at,
			 is_active=true, provider_observed_at=EXCLUDED.provider_observed_at,
			 imported_at=EXCLUDED.imported_at, expires_at=EXCLUDED.expires_at,
			 last_import_run_id=EXCLUDED.last_import_run_id, updated_at=now()
			RETURNING id`, run.ProviderConfiguration.MerchantID, record.ProductID, record.MerchantSKU,
			record.ProductURL, record.PriceMinor, record.ShippingMinor, record.Currency,
			record.Availability, record.Condition, record.ObservedAt, record.ProviderObservedAt,
			importedAt, record.ExpiresAt, run.ID).Scan(&offerID)
		if err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("upsert imported offer: %w", err)
		}
		_, err = tx.Exec(ctx, `INSERT INTO commerce.provider_offer_mappings
			(provider_configuration_id, external_offer_id, merchant_offer_id, last_seen_import_run_id,
			 first_seen_at, last_seen_at, is_present)
			VALUES ($1,$2,$3,$4,$5,$5,true)
			ON CONFLICT (provider_configuration_id, external_offer_id) DO UPDATE SET
			 merchant_offer_id=EXCLUDED.merchant_offer_id, last_seen_import_run_id=EXCLUDED.last_seen_import_run_id,
			 last_seen_at=EXCLUDED.last_seen_at, is_present=true`, run.ProviderConfiguration.ID,
			record.ExternalOfferID, offerID, run.ID, importedAt)
		if err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("map imported offer: %w", err)
		}
		if record.AffiliateURL != nil {
			_, err = tx.Exec(ctx, `INSERT INTO commerce.affiliate_links
				(merchant_offer_id, provider, destination_url, external_reference, disclosure_label,
				 is_active, commission_type)
				VALUES ($1,$2,$3,$4,'Affiliate link',true,'unknown')
				ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET destination_url=EXCLUDED.destination_url,
				 external_reference=EXCLUDED.external_reference, is_active=true, updated_at=now()`, offerID,
				run.ProviderConfiguration.ProviderKey, *record.AffiliateURL, record.ExternalOfferID)
			if err != nil {
				return domain.ImportApplyResult{}, fmt.Errorf("upsert imported affiliate link: %w", err)
			}
		}
		if _, err = tx.Exec(ctx, `INSERT INTO commerce.price_observations
			(provider_configuration_id, import_run_id, merchant_offer_id, external_offer_id,
			 price_minor, shipping_minor, currency, provider_observed_at, observed_at, expires_at,
			 imported_at, observation_fingerprint)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12) ON CONFLICT DO NOTHING`,
			run.ProviderConfiguration.ID, run.ID, offerID, record.ExternalOfferID, record.PriceMinor,
			record.ShippingMinor, record.Currency, record.ProviderObservedAt, record.ObservedAt,
			record.ExpiresAt, importedAt, record.PriceFingerprint); err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("record price observation: %w", err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO commerce.availability_observations
			(provider_configuration_id, import_run_id, merchant_offer_id, external_offer_id,
			 availability, provider_observed_at, observed_at, expires_at, imported_at,
			 observation_fingerprint)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT DO NOTHING`,
			run.ProviderConfiguration.ID, run.ID, offerID, record.ExternalOfferID, record.Availability,
			record.ProviderObservedAt, record.ObservedAt, record.ExpiresAt, importedAt,
			record.AvailabilityFingerprint); err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("record availability observation: %w", err)
		}
		result.Applied++
	}
	if _, err = tx.Exec(ctx, `UPDATE commerce.offer_import_runs SET records_received=records_received+$2,
		records_applied=records_applied+$3, records_rejected=records_rejected+$4,
		cursor_after=$5, updated_at=$6 WHERE id=$1 AND status='running'`, run.ID,
		result.Received, result.Applied, result.Rejected, batch.NextCursor, importedAt); err != nil {
		return domain.ImportApplyResult{}, fmt.Errorf("update import progress: %w", err)
	}
	if batch.Complete {
		var totalRejected int
		if err = tx.QueryRow(ctx, `SELECT records_rejected FROM commerce.offer_import_runs WHERE id=$1`, run.ID).Scan(&totalRejected); err != nil {
			return domain.ImportApplyResult{}, fmt.Errorf("read import rejection count: %w", err)
		}
		if totalRejected == 0 {
			command, reconcileErr := tx.Exec(ctx, `WITH missing AS (
				UPDATE commerce.provider_offer_mappings SET is_present=false
				WHERE provider_configuration_id=$1 AND last_seen_import_run_id<>$2 AND is_present=true
				RETURNING merchant_offer_id
			) UPDATE commerce.merchant_offers SET is_active=false, availability='out_of_stock',
				updated_at=$3 WHERE id IN (SELECT merchant_offer_id FROM missing)`, run.ProviderConfiguration.ID,
				run.ID, importedAt)
			if reconcileErr != nil {
				return domain.ImportApplyResult{}, fmt.Errorf("reconcile missing offers: %w", reconcileErr)
			}
			result.OffersDeactivated = int(command.RowsAffected())
			_, _ = tx.Exec(ctx, `UPDATE commerce.offer_import_runs SET offers_deactivated=offers_deactivated+$2 WHERE id=$1`, run.ID, result.OffersDeactivated)
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ImportApplyResult{}, fmt.Errorf("commit provider import: %w", err)
	}
	return result, nil
}

func (repository *Repository) CompleteImport(ctx context.Context, id domain.ImportRunID, result domain.ImportApplyResult, completedAt time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin import completion: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	status := domain.ImportSucceeded
	if result.Rejected > 0 {
		status = domain.ImportPartial
	}
	command, err := tx.Exec(ctx, `UPDATE commerce.offer_import_runs SET status=$2, completed_at=$3,
		cursor_after=$4, next_retry_at=NULL, error_code=NULL, error_message=NULL, updated_at=$3
		WHERE id=$1 AND status='running'`, id, status, completedAt, result.NextCursor)
	if err != nil {
		return fmt.Errorf("complete commerce import: %w", err)
	}
	if command.RowsAffected() == 0 {
		return ports.ErrImportConflict
	}
	succeeded := status == domain.ImportSucceeded
	_, err = tx.Exec(ctx, `UPDATE commerce.provider_configurations configurations SET
		cursor_value=CASE WHEN $3 THEN runs.cursor_after ELSE configurations.cursor_value END,
		last_import_succeeded_at=CASE WHEN $3 THEN $2 ELSE configurations.last_import_succeeded_at END,
		last_import_failed_at=CASE WHEN $3 THEN configurations.last_import_failed_at ELSE $2 END,
		consecutive_failures=CASE WHEN $3 THEN 0 ELSE configurations.consecutive_failures+1 END,
		last_error_code=CASE WHEN $3 THEN NULL ELSE 'import.partial' END,
		lifecycle_status=CASE
			WHEN $3 AND lifecycle_status='degraded' THEN 'active'
			WHEN NOT $3 AND lifecycle_status='active' THEN 'degraded'
			ELSE lifecycle_status END,
		next_import_at=COALESCE(next_import_at,$2::timestamptz + make_interval(mins => schedule_interval_minutes)), updated_at=$2
		FROM commerce.offer_import_runs runs WHERE runs.id=$1 AND configurations.id=runs.provider_configuration_id`, id, completedAt, succeeded)
	if err != nil {
		return fmt.Errorf("complete provider state: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit import completion: %w", err)
	}
	return nil
}

func (repository *Repository) FailImport(ctx context.Context, run domain.ImportRun, code, message string, failedAt time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin import failure: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	nextRetry, retry := domain.NextImportRetry(run.AttemptCount, run.MaxAttempts, failedAt)
	status := domain.ImportFailed
	if retry {
		status = domain.ImportRetryWait
	}
	command, err := tx.Exec(ctx, `UPDATE commerce.offer_import_runs SET status=$2, error_code=$3,
		error_message=$4, next_retry_at=$5::timestamptz,
		completed_at=CASE WHEN $2='failed' THEN $6::timestamptz ELSE NULL::timestamptz END,
		updated_at=$6 WHERE id=$1 AND status='running'`, run.ID, status, code, message, nextRetry, failedAt)
	if err != nil {
		return fmt.Errorf("mark import failed: %w", err)
	}
	if command.RowsAffected() == 0 {
		return ports.ErrImportConflict
	}
	_, err = tx.Exec(ctx, `UPDATE commerce.provider_configurations SET last_import_failed_at=$2,
		consecutive_failures=consecutive_failures+1, last_error_code=$3,
		lifecycle_status=CASE WHEN lifecycle_status='active' THEN 'degraded' ELSE lifecycle_status END,
		updated_at=$2 WHERE id=$1`, run.ProviderConfiguration.ID, failedAt, code)
	if err != nil {
		return fmt.Errorf("mark provider degraded: %w", err)
	}
	return tx.Commit(ctx)
}

func (repository *Repository) RetryImport(ctx context.Context, actor identity.UserID, id domain.ImportRunID, key string) (domain.ImportRun, error) {
	run, err := repository.getImport(ctx, id)
	if err != nil {
		return domain.ImportRun{}, err
	}
	if run.Status != domain.ImportFailed && run.Status != domain.ImportPartial && run.Status != domain.ImportCancelled {
		return domain.ImportRun{}, ports.ErrImportConflict
	}
	return repository.QueueImport(ctx, &actor, run.ProviderConfiguration.ID, domain.ImportRetry, key, run.MaxAttempts)
}

func (repository *Repository) ListImports(ctx context.Context, limit, offset int) ([]domain.ImportRun, int64, error) {
	rows, err := repository.pool.Query(ctx, importSelect+` ORDER BY runs.created_at DESC, runs.id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list commerce imports: %w", err)
	}
	defer rows.Close()
	runs := make([]domain.ImportRun, 0)
	var total int64
	for rows.Next() {
		run, count, scanErr := scanImport(rows, true)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		total = count
		runs = append(runs, run)
	}
	return runs, total, rows.Err()
}

func (repository *Repository) ListImportFailures(ctx context.Context, runID domain.ImportRunID, limit, offset int) ([]domain.ImportFailure, int64, error) {
	rows, err := repository.pool.Query(ctx, `SELECT count(*) OVER(), id, import_run_id,
		external_record_id, error_code, error_message, record_fingerprint, created_at
		FROM commerce.offer_import_failures WHERE import_run_id=$1 ORDER BY created_at,id LIMIT $2 OFFSET $3`, runID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list import failures: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ImportFailure, 0)
	var total int64
	for rows.Next() {
		var item domain.ImportFailure
		if err = rows.Scan(&total, &item.ID, &item.ImportRunID, &item.ExternalRecordID,
			&item.ErrorCode, &item.ErrorMessage, &item.RecordFingerprint, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan import failure: %w", err)
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (repository *Repository) AnonymizeExpiredClicks(ctx context.Context, now time.Time, limit int) (int64, error) {
	var count int64
	err := repository.pool.QueryRow(ctx, `WITH expired AS MATERIALIZED (
		SELECT id FROM commerce.affiliate_clicks WHERE anonymized_at IS NULL
		AND retention_expires_at <= $1 ORDER BY retention_expires_at,id LIMIT $2 FOR UPDATE SKIP LOCKED
	), conversion_attribution AS (
		UPDATE commerce.affiliate_conversions conversions SET recommendation_id=NULL,
		recommendation_item_id=NULL FROM expired WHERE conversions.affiliate_click_id=expired.id
	), anonymized AS (
	UPDATE commerce.affiliate_clicks clicks SET user_id=NULL, anonymous_id=NULL, session_id=NULL,
		campaign=NULL, referrer=NULL, request_id=NULL, traffic_source=NULL, traffic_medium=NULL,
		referrer_host=NULL, recommendation_id=NULL, recommendation_item_id=NULL, user_agent_hash=NULL,
		idempotency_key=NULL, anonymized_at=$1 FROM expired WHERE clicks.id=expired.id RETURNING clicks.id
	) SELECT count(*) FROM anonymized`, now, limit).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("anonymize expired affiliate clicks: %w", err)
	}
	remaining := limit - int(count)
	if remaining <= 0 {
		return count, nil
	}
	var promotionCount int64
	err = repository.pool.QueryRow(ctx, `WITH expired AS MATERIALIZED (
		SELECT id FROM commerce.affiliate_promotion_clicks WHERE anonymized_at IS NULL
		AND retention_expires_at <= $1 ORDER BY retention_expires_at,id LIMIT $2 FOR UPDATE SKIP LOCKED
	), anonymized AS (
	UPDATE commerce.affiliate_promotion_clicks clicks SET user_id=NULL, anonymous_id=NULL,
		session_id=NULL, campaign=NULL, referrer=NULL, request_id=NULL, traffic_source=NULL,
		traffic_medium=NULL, referrer_host=NULL, user_agent_hash=NULL, idempotency_key=NULL,
		anonymized_at=$1 FROM expired WHERE clicks.id=expired.id RETURNING clicks.id
	) SELECT count(*) FROM anonymized`, now, remaining).Scan(&promotionCount)
	if err != nil {
		return count, fmt.Errorf("anonymize expired affiliate promotion clicks: %w", err)
	}
	return count + promotionCount, nil
}

func (repository *Repository) getImport(ctx context.Context, id domain.ImportRunID) (domain.ImportRun, error) {
	run, _, err := scanImport(repository.pool.QueryRow(ctx, importSelect+` WHERE runs.id=$1`, id), false)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ImportRun{}, ports.ErrImportNotFound
	}
	return run, err
}

const providerConfigurationSelect = `SELECT configurations.id, configurations.merchant_id,
	merchants.name, configurations.provider_key, configurations.adapter_key,
	configurations.external_merchant_id, configurations.credential_reference,
	configurations.lifecycle_status, configurations.configuration_verified_at,
	configurations.schedule_interval_minutes, configurations.freshness_ttl_minutes,
	configurations.cursor_value, configurations.next_import_at,
	configurations.last_import_started_at, configurations.last_import_succeeded_at,
	configurations.last_import_failed_at, configurations.consecutive_failures,
	configurations.last_error_code, configurations.conversion_cursor_value,
	configurations.next_conversion_import_at, configurations.last_conversion_import_succeeded_at,
	configurations.last_conversion_import_failed_at, configurations.conversion_consecutive_failures,
	configurations.last_conversion_error_code, configurations.conversion_ingestion_enabled,
	configurations.conversion_configuration_verified_at, configurations.created_at, configurations.updated_at
	FROM commerce.provider_configurations configurations
	JOIN commerce.merchants merchants ON merchants.id=configurations.merchant_id`

const importSelect = `SELECT count(*) OVER(), runs.id, runs.trigger_type, runs.status,
	runs.idempotency_key, runs.requested_by, runs.cursor_before, runs.cursor_after,
	runs.attempt_count, runs.max_attempts, runs.records_received, runs.records_applied,
	runs.records_rejected, runs.offers_deactivated, runs.error_code, runs.error_message,
	runs.next_retry_at, runs.started_at, runs.completed_at, runs.created_at, runs.updated_at,
	configurations.id, configurations.merchant_id, merchants.name, configurations.provider_key,
	configurations.adapter_key, configurations.external_merchant_id, configurations.credential_reference,
	configurations.lifecycle_status, configurations.configuration_verified_at,
	configurations.schedule_interval_minutes, configurations.freshness_ttl_minutes,
	configurations.cursor_value, configurations.next_import_at, configurations.last_import_started_at,
	configurations.last_import_succeeded_at, configurations.last_import_failed_at,
	configurations.consecutive_failures, configurations.last_error_code,
	configurations.conversion_cursor_value, configurations.next_conversion_import_at,
	configurations.last_conversion_import_succeeded_at, configurations.last_conversion_import_failed_at,
	configurations.conversion_consecutive_failures, configurations.last_conversion_error_code,
	configurations.conversion_ingestion_enabled, configurations.conversion_configuration_verified_at,
	configurations.created_at, configurations.updated_at
	FROM commerce.offer_import_runs runs
	JOIN commerce.provider_configurations configurations ON configurations.id=runs.provider_configuration_id
	JOIN commerce.merchants merchants ON merchants.id=configurations.merchant_id`

type rowScanner interface{ Scan(...any) error }

func scanProviderConfiguration(row rowScanner) (domain.ProviderConfiguration, error) {
	var value domain.ProviderConfiguration
	err := row.Scan(&value.ID, &value.MerchantID, &value.MerchantName, &value.ProviderKey,
		&value.AdapterKey, &value.ExternalMerchantID, &value.CredentialReference,
		&value.LifecycleStatus, &value.ConfigurationVerifiedAt, &value.ScheduleIntervalMinutes,
		&value.FreshnessTTLMinutes, &value.Cursor, &value.NextImportAt, &value.LastImportStartedAt,
		&value.LastImportSucceededAt, &value.LastImportFailedAt, &value.ConsecutiveFailures,
		&value.LastErrorCode, &value.ConversionCursor, &value.NextConversionImportAt,
		&value.LastConversionSucceeded, &value.LastConversionFailed, &value.ConversionFailures,
		&value.LastConversionError, &value.ConversionEnabled, &value.ConversionVerifiedAt,
		&value.CreatedAt, &value.UpdatedAt)
	if err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("scan provider configuration: %w", err)
	}
	return value, nil
}

func scanImport(row rowScanner, includesCount bool) (domain.ImportRun, int64, error) {
	var run domain.ImportRun
	var count int64
	targets := []any{&count, &run.ID, &run.Trigger, &run.Status, &run.IdempotencyKey,
		&run.RequestedBy, &run.CursorBefore, &run.CursorAfter, &run.AttemptCount, &run.MaxAttempts,
		&run.RecordsReceived, &run.RecordsApplied, &run.RecordsRejected, &run.OffersDeactivated,
		&run.ErrorCode, &run.ErrorMessage, &run.NextRetryAt, &run.StartedAt, &run.CompletedAt,
		&run.CreatedAt, &run.UpdatedAt, &run.ProviderConfiguration.ID,
		&run.ProviderConfiguration.MerchantID, &run.ProviderConfiguration.MerchantName,
		&run.ProviderConfiguration.ProviderKey, &run.ProviderConfiguration.AdapterKey,
		&run.ProviderConfiguration.ExternalMerchantID, &run.ProviderConfiguration.CredentialReference,
		&run.ProviderConfiguration.LifecycleStatus, &run.ProviderConfiguration.ConfigurationVerifiedAt,
		&run.ProviderConfiguration.ScheduleIntervalMinutes, &run.ProviderConfiguration.FreshnessTTLMinutes,
		&run.ProviderConfiguration.Cursor, &run.ProviderConfiguration.NextImportAt,
		&run.ProviderConfiguration.LastImportStartedAt, &run.ProviderConfiguration.LastImportSucceededAt,
		&run.ProviderConfiguration.LastImportFailedAt, &run.ProviderConfiguration.ConsecutiveFailures,
		&run.ProviderConfiguration.LastErrorCode, &run.ProviderConfiguration.ConversionCursor,
		&run.ProviderConfiguration.NextConversionImportAt,
		&run.ProviderConfiguration.LastConversionSucceeded,
		&run.ProviderConfiguration.LastConversionFailed, &run.ProviderConfiguration.ConversionFailures,
		&run.ProviderConfiguration.LastConversionError, &run.ProviderConfiguration.ConversionEnabled,
		&run.ProviderConfiguration.ConversionVerifiedAt, &run.ProviderConfiguration.CreatedAt,
		&run.ProviderConfiguration.UpdatedAt}
	if !includesCount {
		// importSelect always contains the window count; retain one scan contract.
		_ = includesCount
	}
	if err := row.Scan(targets...); err != nil {
		return domain.ImportRun{}, 0, fmt.Errorf("scan commerce import: %w", err)
	}
	return run, count, nil
}

func insertOperationAudit(ctx context.Context, tx pgx.Tx, actor identity.UserID, action, entityType, entityID string, details map[string]any) error {
	if details == nil {
		details = map[string]any{}
	}
	payload, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("encode operation audit details: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO commerce.operation_audit_log
		(actor_user_id, action, entity_type, entity_id, details) VALUES ($1,$2,$3,$4,$5)`,
		actor, action, entityType, entityID, payload)
	if err != nil {
		return fmt.Errorf("record operation audit: %w", err)
	}
	return nil
}

var _ ports.ImportRepository = (*Repository)(nil)
