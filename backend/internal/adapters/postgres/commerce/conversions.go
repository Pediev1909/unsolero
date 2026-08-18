package commercepostgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"rigmark/internal/modules/commerce/domain"
	"rigmark/internal/modules/commerce/ports"
	identity "rigmark/internal/modules/identity/domain"
)

func (repository *Repository) SetConversionProviderEnabled(ctx context.Context, actor identity.UserID, id domain.ProviderConfigurationID, enabled bool, verifiedAt time.Time) (domain.ProviderConfiguration, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ProviderConfiguration{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	command, err := tx.Exec(ctx, `UPDATE commerce.provider_configurations SET
		conversion_ingestion_enabled=$2,
		conversion_configuration_verified_at=CASE WHEN $2 THEN $3 ELSE conversion_configuration_verified_at END,
		next_conversion_import_at=CASE WHEN $2 THEN COALESCE(next_conversion_import_at,$3) ELSE NULL END,
		updated_at=$3 WHERE id=$1`, id, enabled, verifiedAt)
	if err != nil {
		return domain.ProviderConfiguration{}, fmt.Errorf("set conversion provider state: %w", err)
	}
	if command.RowsAffected() != 1 {
		return domain.ProviderConfiguration{}, ports.ErrImportNotFound
	}
	if err = insertOperationAudit(ctx, tx, actor, "provider.conversions", "provider_configuration", string(id), map[string]any{"enabled": enabled}); err != nil {
		return domain.ProviderConfiguration{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ProviderConfiguration{}, err
	}
	return repository.GetProviderConfiguration(ctx, id)
}

func (repository *Repository) RecordWebhookDelivery(ctx context.Context, configurationID domain.ProviderConfigurationID, requestFingerprint, bodyFingerprint, state string, signatureTimestamp *time.Time, errorCode *string, receivedAt time.Time) (domain.WebhookDelivery, error) {
	var delivery domain.WebhookDelivery
	err := repository.pool.QueryRow(ctx, `INSERT INTO commerce.conversion_webhook_deliveries
		(provider_configuration_id, request_fingerprint, body_fingerprint, verification_state,
		 signature_timestamp, error_code, received_at, processed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7::timestamptz,CASE WHEN $4='rejected' THEN $7::timestamptz ELSE NULL END)
		ON CONFLICT (provider_configuration_id, request_fingerprint) DO UPDATE
		SET request_fingerprint=EXCLUDED.request_fingerprint
		WHERE conversion_webhook_deliveries.body_fingerprint=EXCLUDED.body_fingerprint
		RETURNING id,verification_state,processed_at IS NOT NULL`,
		configurationID, requestFingerprint, bodyFingerprint, state, signatureTimestamp, errorCode, receivedAt).Scan(
		&delivery.ID, &delivery.VerificationState, &delivery.Processed)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.WebhookDelivery{}, ports.ErrWebhookReplay
	}
	if err != nil {
		return domain.WebhookDelivery{}, fmt.Errorf("record conversion webhook delivery: %w", err)
	}
	return delivery, nil
}

func (repository *Repository) ResolveConversionAttribution(ctx context.Context, configuration domain.ProviderConfiguration, clickID *string, eventAt time.Time, window time.Duration) (domain.ConversionAttribution, error) {
	result := domain.ConversionAttribution{Status: "unattributed"}
	if clickID == nil || strings.TrimSpace(*clickID) == "" {
		return result, nil
	}
	var storedClickID string
	var recommendationID, recommendationItemID, source, campaign *string
	err := repository.pool.QueryRow(ctx, `SELECT clicks.id, clicks.recommendation_id,
		clicks.recommendation_item_id, clicks.source, clicks.campaign
		FROM commerce.affiliate_clicks clicks
		JOIN commerce.merchant_offers offers ON offers.id=clicks.merchant_offer_id
		WHERE clicks.id=$1 AND offers.merchant_id=$2 AND clicks.is_countable
		  AND clicks.occurred_at <= $3 AND clicks.occurred_at >= $3-make_interval(secs=>$4)`,
		strings.TrimSpace(*clickID), configuration.MerchantID, eventAt, int64(window/time.Second)).Scan(
		&storedClickID, &recommendationID, &recommendationItemID, &source, &campaign)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return domain.ConversionAttribution{}, fmt.Errorf("resolve conversion attribution: %w", err)
	}
	result.Status = "attributed"
	result.ClickID = &storedClickID
	result.RecommendationID = recommendationID
	result.RecommendationItemID = recommendationItemID
	result.Source = source
	result.Campaign = campaign
	return result, nil
}

func (repository *Repository) ApplyWebhookEvents(ctx context.Context, deliveryID domain.WebhookDeliveryID, configuration domain.ProviderConfiguration, events []domain.VerifiedConversionEvent, processedAt time.Time) (int, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin conversion webhook: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	applied, err := applyVerifiedConversionEvents(ctx, tx, configuration, events)
	if err != nil {
		return 0, err
	}
	command, err := tx.Exec(ctx, `UPDATE commerce.conversion_webhook_deliveries
		SET provider_event_count=$2, processed_at=$3 WHERE id=$1 AND verification_state='verified'`,
		deliveryID, len(events), processedAt)
	if err != nil || command.RowsAffected() != 1 {
		return 0, fmt.Errorf("complete conversion webhook delivery: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit conversion webhook: %w", err)
	}
	return applied, nil
}

func applyVerifiedConversionEvents(ctx context.Context, tx pgx.Tx, configuration domain.ProviderConfiguration, events []domain.VerifiedConversionEvent) (int, error) {
	applied := 0
	for _, event := range events {
		var conversionID domain.ConversionID
		var previousOrder domain.OrderStatus
		var previousCommission *domain.CommissionStatus
		var latestAt *time.Time
		err := tx.QueryRow(ctx, `INSERT INTO commerce.affiliate_conversions
			(affiliate_click_id, provider, external_conversion_id, order_reference, order_status,
			 order_value_minor, order_currency, commission_amount_minor, commission_currency,
			 converted_at, provider_configuration_id, merchant_id, recommendation_id,
			 recommendation_item_id, event_type, event_received_at, commission_status,
			 attribution_status, source, campaign, raw_provider_reference, verification_state,
			 latest_event_at, cancelled_at, reversed_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::timestamptz,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,
				 'verified',$10::timestamptz,CASE WHEN $5='cancelled' THEN $10::timestamptz ELSE NULL END,
				 CASE WHEN $5='reversed' OR $17='reversed' THEN $10::timestamptz ELSE NULL END)
			ON CONFLICT (provider_configuration_id, external_conversion_id)
			WHERE provider_configuration_id IS NOT NULL DO UPDATE SET updated_at=affiliate_conversions.updated_at
			RETURNING id, order_status, commission_status, latest_event_at`,
			event.Attribution.ClickID, event.Provider, event.ExternalConversionID, event.OrderReference,
			event.OrderStatus, event.OrderValueMinor, event.OrderCurrency, event.CommissionMinor,
			event.CommissionCurrency, event.EventTimestamp, event.ProviderConfigurationID,
			event.MerchantID, event.Attribution.RecommendationID, event.Attribution.RecommendationItemID,
			event.EventType, event.ReceivedAt, event.CommissionStatus, event.Attribution.Status,
			event.Attribution.Source, event.Attribution.Campaign, event.RawProviderReference).Scan(
			&conversionID, &previousOrder, &previousCommission, &latestAt)
		if err != nil {
			return 0, fmt.Errorf("load conversion projection: %w", err)
		}
		var eventID domain.ConversionEventID
		err = tx.QueryRow(ctx, `INSERT INTO commerce.conversion_events
			(provider_configuration_id, webhook_delivery_id, import_run_id, affiliate_conversion_id,
			 provider, provider_event_id, event_type, external_conversion_id, order_reference,
			 order_status, order_value_minor, order_currency, commission_amount_minor,
			 commission_currency, commission_status, affiliate_click_id, attribution_status,
			 source, campaign, raw_provider_reference, event_timestamp, received_at,
			 verification_state, payload_fingerprint)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,
			 $21,$22,'verified',$23)
			ON CONFLICT (provider_configuration_id, provider_event_id) DO NOTHING RETURNING id`,
			event.ProviderConfigurationID, event.WebhookDeliveryID, event.ImportRunID, conversionID,
			event.Provider, event.ProviderEventID, event.EventType, event.ExternalConversionID,
			event.OrderReference, event.OrderStatus, event.OrderValueMinor, event.OrderCurrency,
			event.CommissionMinor, event.CommissionCurrency, event.CommissionStatus,
			event.Attribution.ClickID, event.Attribution.Status, event.Attribution.Source,
			event.Attribution.Campaign, event.RawProviderReference, event.EventTimestamp,
			event.ReceivedAt, event.PayloadFingerprint).Scan(&eventID)
		if errors.Is(err, pgx.ErrNoRows) {
			var storedFingerprint string
			if queryErr := tx.QueryRow(ctx, `SELECT payload_fingerprint FROM commerce.conversion_events
				WHERE provider_configuration_id=$1 AND provider_event_id=$2`,
				event.ProviderConfigurationID, event.ProviderEventID).Scan(&storedFingerprint); queryErr != nil {
				return 0, fmt.Errorf("read duplicate conversion event: %w", queryErr)
			}
			if storedFingerprint != event.PayloadFingerprint {
				return 0, ports.ErrConversionConflict
			}
			continue
		}
		if err != nil {
			return 0, fmt.Errorf("record conversion event: %w", err)
		}
		if latestAt != nil && event.EventTimestamp.Before(*latestAt) {
			continue
		}
		correction := event.EventType == domain.EventCorrection
		if latestAt != nil && !domain.ValidateOrderTransition(previousOrder, event.OrderStatus, correction) {
			return 0, ports.ErrConversionConflict
		}
		if latestAt != nil && previousCommission != nil && event.CommissionStatus != nil &&
			!domain.ValidateCommissionTransition(*previousCommission, *event.CommissionStatus, correction) {
			return 0, ports.ErrConversionConflict
		}
		_, err = tx.Exec(ctx, `UPDATE commerce.affiliate_conversions SET
			affiliate_click_id=$2, order_reference=$3, order_status=$4, order_value_minor=$5,
			order_currency=$6, commission_amount_minor=$7, commission_currency=$8,
			recommendation_id=$9, recommendation_item_id=$10, event_type=$11,
			event_received_at=$12, commission_status=$13, attribution_status=$14, source=$15,
			campaign=$16, raw_provider_reference=$17, latest_event_at=$18, latest_event_id=$19,
			cancelled_at=CASE WHEN $4='cancelled' THEN $18 ELSE cancelled_at END,
			reversed_at=CASE WHEN $4='reversed' OR $13='reversed' THEN $18 ELSE reversed_at END,
			confirmed_at=CASE WHEN $4='confirmed' THEN COALESCE(confirmed_at,$18) ELSE confirmed_at END,
			updated_at=$12 WHERE id=$1`, conversionID, event.Attribution.ClickID,
			event.OrderReference, event.OrderStatus, event.OrderValueMinor, event.OrderCurrency,
			event.CommissionMinor, event.CommissionCurrency, event.Attribution.RecommendationID,
			event.Attribution.RecommendationItemID, event.EventType, event.ReceivedAt,
			event.CommissionStatus, event.Attribution.Status, event.Attribution.Source,
			event.Attribution.Campaign, event.RawProviderReference, event.EventTimestamp, eventID)
		if err != nil {
			return 0, fmt.Errorf("update conversion projection: %w", err)
		}
		applied++
	}
	return applied, nil
}

func (repository *Repository) QueueConversionImport(ctx context.Context, actor *identity.UserID, configurationID domain.ProviderConfigurationID, trigger domain.ImportTrigger, key string, attempts int16) (domain.ConversionImportRun, error) {
	var actorValue any
	if actor != nil {
		actorValue = *actor
	}
	var id domain.ConversionImportRunID
	err := repository.pool.QueryRow(ctx, `INSERT INTO commerce.conversion_import_runs
		(provider_configuration_id, trigger_type, idempotency_key, requested_by, cursor_before, max_attempts)
		SELECT id,$2,$3,$4::uuid,conversion_cursor_value,$5 FROM commerce.provider_configurations
		WHERE id=$1 AND lifecycle_status IN ('active','degraded') AND conversion_ingestion_enabled
		ON CONFLICT (provider_configuration_id,idempotency_key)
		DO UPDATE SET idempotency_key=EXCLUDED.idempotency_key RETURNING id`,
		configurationID, trigger, key, actorValue, attempts).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConversionImportRun{}, ports.ErrProviderDisabled
	}
	if err != nil {
		return domain.ConversionImportRun{}, fmt.Errorf("queue conversion import: %w", err)
	}
	return repository.getConversionImport(ctx, id)
}

func (repository *Repository) QueueDueConversionImports(ctx context.Context, now time.Time, limit int) (int, error) {
	command, err := repository.pool.Exec(ctx, `WITH due AS (
		SELECT id,next_conversion_import_at FROM commerce.provider_configurations
		WHERE lifecycle_status IN ('active','degraded') AND conversion_ingestion_enabled
		AND next_conversion_import_at <= $1
		ORDER BY next_conversion_import_at,id LIMIT $2 FOR UPDATE SKIP LOCKED
	), queued AS (
		INSERT INTO commerce.conversion_import_runs
		(provider_configuration_id,trigger_type,status,idempotency_key,cursor_before)
		SELECT configurations.id,'scheduled','queued',
		'conversion-'||extract(epoch FROM configurations.next_conversion_import_at)::bigint,
		configurations.conversion_cursor_value
		FROM commerce.provider_configurations configurations JOIN due USING(id)
		ON CONFLICT(provider_configuration_id,idempotency_key) DO NOTHING
	)
	UPDATE commerce.provider_configurations configurations SET
		next_conversion_import_at=$1+make_interval(mins=>schedule_interval_minutes),updated_at=$1
	FROM due WHERE configurations.id=due.id`, now, limit)
	if err != nil {
		return 0, fmt.Errorf("queue due conversion imports: %w", err)
	}
	return int(command.RowsAffected()), nil
}

func (repository *Repository) RecoverStalledConversionImports(ctx context.Context, staleBefore, recoveredAt time.Time, limit int) (int, error) {
	var count int
	err := repository.pool.QueryRow(ctx, `WITH candidates AS (
		SELECT id,attempt_count,max_attempts FROM commerce.conversion_import_runs
		WHERE status='running' AND updated_at<$1 ORDER BY updated_at,id LIMIT $3 FOR UPDATE SKIP LOCKED
	), recovered AS (
		UPDATE commerce.conversion_import_runs runs SET
		status=CASE WHEN candidates.attempt_count<candidates.max_attempts THEN 'retry_wait' ELSE 'failed' END,
		next_retry_at=CASE WHEN candidates.attempt_count<candidates.max_attempts THEN $2::timestamptz ELSE NULL END,
		completed_at=CASE WHEN candidates.attempt_count<candidates.max_attempts THEN NULL ELSE $2::timestamptz END,
		error_code='worker.lease_expired',error_message='The conversion worker lease expired.',updated_at=$2
		FROM candidates WHERE runs.id=candidates.id RETURNING runs.id,runs.provider_configuration_id
	), provider_health AS (
		UPDATE commerce.provider_configurations configurations SET
		last_conversion_import_failed_at=$2,conversion_consecutive_failures=conversion_consecutive_failures+1,
		last_conversion_error_code='worker.lease_expired',
		lifecycle_status=CASE WHEN lifecycle_status='active' THEN 'degraded' ELSE lifecycle_status END,
		updated_at=$2 FROM recovered WHERE configurations.id=recovered.provider_configuration_id
		RETURNING configurations.id
	) SELECT count(*) FROM recovered`, staleBefore, recoveredAt, limit).Scan(&count)
	return count, err
}

func (repository *Repository) ClaimNextConversionImport(ctx context.Context, now time.Time) (domain.ConversionImportRun, error) {
	var id domain.ConversionImportRunID
	err := repository.pool.QueryRow(ctx, `WITH candidate AS (
		SELECT runs.id FROM commerce.conversion_import_runs runs
		JOIN commerce.provider_configurations configurations ON configurations.id=runs.provider_configuration_id
		WHERE runs.status IN ('queued','retry_wait') AND (runs.next_retry_at IS NULL OR runs.next_retry_at<=$1)
		AND runs.attempt_count<runs.max_attempts AND configurations.lifecycle_status IN ('active','degraded')
		AND configurations.conversion_ingestion_enabled
		ORDER BY COALESCE(runs.next_retry_at,runs.created_at),runs.id LIMIT 1 FOR UPDATE OF runs SKIP LOCKED
	), claimed AS (
	UPDATE commerce.conversion_import_runs runs SET status='running',attempt_count=attempt_count+1,
		started_at=COALESCE(started_at,$1),next_retry_at=NULL,updated_at=$1 FROM candidate
		WHERE runs.id=candidate.id RETURNING runs.id,runs.provider_configuration_id
	), provider_health AS (
		UPDATE commerce.provider_configurations configurations SET
		last_conversion_import_started_at=$1,updated_at=$1 FROM claimed
		WHERE configurations.id=claimed.provider_configuration_id RETURNING configurations.id
	) SELECT claimed.id FROM claimed JOIN provider_health
	ON provider_health.id=claimed.provider_configuration_id`, now).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConversionImportRun{}, ports.ErrImportNotFound
	}
	if err != nil {
		return domain.ConversionImportRun{}, fmt.Errorf("claim conversion import: %w", err)
	}
	return repository.getConversionImport(ctx, id)
}

func (repository *Repository) ApplyConversionImport(ctx context.Context, run domain.ConversionImportRun, events []domain.VerifiedConversionEvent, failures []domain.ImportRecordFailure, _ domain.ConversionBatch, now time.Time) (int, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var status domain.ImportStatus
	if err = tx.QueryRow(ctx, `SELECT status FROM commerce.conversion_import_runs WHERE id=$1 FOR UPDATE`, run.ID).Scan(&status); errors.Is(err, pgx.ErrNoRows) {
		return 0, ports.ErrImportNotFound
	} else if err != nil {
		return 0, fmt.Errorf("lock conversion import: %w", err)
	} else if status != domain.ImportRunning {
		return 0, ports.ErrImportConflict
	}
	for _, failure := range failures {
		_, err = tx.Exec(ctx, `INSERT INTO commerce.conversion_import_failures
			(import_run_id,provider_event_id,error_code,error_message,record_fingerprint)
			VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, run.ID, failure.ExternalRecordID,
			failure.Code, failure.Message, failure.RecordFingerprint)
		if err != nil {
			return 0, err
		}
	}
	applied, err := applyVerifiedConversionEvents(ctx, tx, run.ProviderConfiguration, events)
	if err != nil {
		return 0, err
	}
	command, err := tx.Exec(ctx, `UPDATE commerce.conversion_import_runs SET
		records_received=records_received+$2,records_applied=records_applied+$3,
		records_rejected=records_rejected+$4,updated_at=$5 WHERE id=$1 AND status='running'`,
		run.ID, len(events)+len(failures), applied, len(failures), now)
	if err != nil || command.RowsAffected() != 1 {
		return 0, ports.ErrImportConflict
	}
	return applied, tx.Commit(ctx)
}

func (repository *Repository) CompleteConversionImport(ctx context.Context, run domain.ConversionImportRun, batch domain.ConversionBatch, applied, rejected int, completedAt time.Time) error {
	status := domain.ImportSucceeded
	if rejected > 0 {
		status = domain.ImportPartial
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	command, err := tx.Exec(ctx, `UPDATE commerce.conversion_import_runs SET status=$2,
		cursor_after=$3,coverage_start=$4,coverage_end=$5,completed_at=$6,updated_at=$6
		WHERE id=$1 AND status='running'`, run.ID, status, batch.NextCursor, batch.CoverageStart,
		batch.CoverageEnd, completedAt)
	if err != nil || command.RowsAffected() != 1 {
		return ports.ErrImportConflict
	}
	succeeded := status == domain.ImportSucceeded
	_, err = tx.Exec(ctx, `UPDATE commerce.provider_configurations SET
		conversion_cursor_value=CASE WHEN $3 THEN $2 ELSE conversion_cursor_value END,
		last_conversion_import_succeeded_at=CASE WHEN $3 THEN $4 ELSE last_conversion_import_succeeded_at END,
		last_conversion_import_failed_at=CASE WHEN $3 THEN last_conversion_import_failed_at ELSE $4 END,
		conversion_consecutive_failures=CASE WHEN $3 THEN 0 ELSE conversion_consecutive_failures+1 END,
		last_conversion_error_code=CASE WHEN $3 THEN NULL ELSE 'import.partial' END,
		next_conversion_import_at=COALESCE(next_conversion_import_at,$4+make_interval(mins=>schedule_interval_minutes)),
		updated_at=$4 WHERE id=$1`, run.ProviderConfiguration.ID, batch.NextCursor, succeeded, completedAt)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (repository *Repository) FailConversionImport(ctx context.Context, run domain.ConversionImportRun, code, message string, failedAt time.Time) error {
	next, retry := domain.NextImportRetry(run.AttemptCount, run.MaxAttempts, failedAt)
	status := domain.ImportFailed
	if retry {
		status = domain.ImportRetryWait
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	command, err := tx.Exec(ctx, `UPDATE commerce.conversion_import_runs SET
		status=$2,error_code=$3,error_message=$4,next_retry_at=$5,
		completed_at=CASE WHEN $2='failed' THEN $6::timestamptz ELSE NULL END,updated_at=$6
		WHERE id=$1 AND status='running'`, run.ID, status, code, message, next, failedAt)
	if err != nil || command.RowsAffected() != 1 {
		return ports.ErrImportConflict
	}
	_, err = tx.Exec(ctx, `UPDATE commerce.provider_configurations SET
		last_conversion_import_failed_at=$2,conversion_consecutive_failures=conversion_consecutive_failures+1,
		last_conversion_error_code=$3,lifecycle_status=CASE WHEN lifecycle_status='active' THEN 'degraded' ELSE lifecycle_status END,
		updated_at=$2 WHERE id=$1`, run.ProviderConfiguration.ID, failedAt, code)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (repository *Repository) RetryConversionImport(ctx context.Context, actor identity.UserID, id domain.ConversionImportRunID, key string) (domain.ConversionImportRun, error) {
	run, err := repository.getConversionImport(ctx, id)
	if err != nil {
		return domain.ConversionImportRun{}, err
	}
	if run.Status != domain.ImportFailed && run.Status != domain.ImportPartial && run.Status != domain.ImportCancelled {
		return domain.ConversionImportRun{}, ports.ErrImportConflict
	}
	return repository.QueueConversionImport(ctx, &actor, run.ProviderConfiguration.ID, domain.ImportRetry, key, run.MaxAttempts)
}

func (repository *Repository) ListConversions(ctx context.Context, filter domain.ConversionFilter) ([]domain.Conversion, int64, error) {
	rows, err := repository.pool.Query(ctx, `SELECT count(*) OVER(), conversions.id,
		conversions.provider_configuration_id,conversions.provider,conversions.merchant_id,
		merchants.name,conversions.external_conversion_id,conversions.order_reference,
		conversions.order_status,conversions.order_value_minor,conversions.order_currency,
		conversions.commission_amount_minor,conversions.commission_currency,
		conversions.commission_status,conversions.attribution_status,conversions.affiliate_click_id,
		conversions.recommendation_id,conversions.source,conversions.campaign,
		conversions.verification_state,conversions.latest_event_at,conversions.event_received_at,
		conversions.updated_at,reconciliation.result
		FROM commerce.affiliate_conversions conversions
		JOIN commerce.merchants merchants ON merchants.id=conversions.merchant_id
		LEFT JOIN LATERAL (SELECT items.result FROM commerce.conversion_reconciliation_items items
			WHERE items.affiliate_conversion_id=conversions.id ORDER BY items.created_at DESC LIMIT 1) reconciliation ON true
		WHERE conversions.verification_state='verified'
		AND ($1='' OR conversions.provider=$1) AND ($2='' OR conversions.order_status=$2)
		AND ($3='' OR conversions.commission_status=$3) AND ($4='' OR conversions.attribution_status=$4)
		AND ($5='' OR COALESCE(reconciliation.result,'unresolved')=$5)
		AND ($6='' OR conversions.commission_currency=$6 OR conversions.order_currency=$6)
		ORDER BY conversions.latest_event_at DESC,conversions.id LIMIT $7 OFFSET $8`,
		filter.Provider, filter.OrderStatus, filter.CommissionStatus, filter.AttributionStatus,
		filter.ReconciliationStatus, strings.ToUpper(filter.Currency), filter.Limit, filter.Offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.Conversion, 0)
	var total int64
	for rows.Next() {
		var item domain.Conversion
		if err = rows.Scan(&total, &item.ID, &item.ProviderConfigurationID, &item.Provider,
			&item.MerchantID, &item.MerchantName, &item.ExternalConversionID, &item.OrderReference,
			&item.OrderStatus, &item.OrderValueMinor, &item.OrderCurrency, &item.CommissionMinor,
			&item.CommissionCurrency, &item.CommissionStatus, &item.AttributionStatus,
			&item.ClickID, &item.RecommendationID, &item.Source, &item.Campaign,
			&item.VerificationState, &item.EventTimestamp, &item.ReceivedAt, &item.UpdatedAt,
			&item.ReconciliationStatus); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (repository *Repository) ReconcileConversionImport(ctx context.Context, actor *identity.UserID, importID domain.ConversionImportRunID, key string, now time.Time) (domain.ReconciliationRun, error) {
	run, err := repository.getConversionImport(ctx, importID)
	if err != nil {
		return domain.ReconciliationRun{}, err
	}
	if run.Status != domain.ImportSucceeded || run.CoverageStart == nil || run.CoverageEnd == nil {
		return domain.ReconciliationRun{}, ports.ErrImportConflict
	}
	var actorValue any
	if actor != nil {
		actorValue = *actor
	}
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.ReconciliationRun{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var id domain.ReconciliationRunID
	var created bool
	err = tx.QueryRow(ctx, `INSERT INTO commerce.conversion_reconciliation_runs
		(provider_configuration_id,conversion_import_run_id,requested_by,idempotency_key,status,
		 coverage_start,coverage_end,started_at)
		VALUES($1,$2,$3::uuid,$4,'running',$5,$6,$7)
		ON CONFLICT(provider_configuration_id,idempotency_key)
		DO UPDATE SET idempotency_key=EXCLUDED.idempotency_key RETURNING id,(xmax=0)`,
		run.ProviderConfiguration.ID, run.ID, actorValue, key, run.CoverageStart, run.CoverageEnd, now).Scan(&id, &created)
	if err != nil {
		return domain.ReconciliationRun{}, err
	}
	if !created {
		if err = tx.Commit(ctx); err != nil {
			return domain.ReconciliationRun{}, err
		}
		return repository.getReconciliation(ctx, id)
	}
	_, err = tx.Exec(ctx, `INSERT INTO commerce.conversion_reconciliation_items
		(reconciliation_run_id,affiliate_conversion_id,provider_event_id,result,reason_code,comparison_fingerprint)
		SELECT $1,events.affiliate_conversion_id,events.provider_event_id,
		CASE WHEN conversions.latest_event_id=events.id THEN 'matched'
		     WHEN events.event_timestamp<conversions.latest_event_at THEN 'stale'
		     ELSE 'conflicting' END,
		CASE WHEN conversions.latest_event_id=events.id THEN 'provider.current_matches'
		     WHEN events.event_timestamp<conversions.latest_event_at THEN 'provider.historical_event'
		     ELSE 'provider.current_conflict' END,
		events.payload_fingerprint
		FROM commerce.conversion_events events
		JOIN commerce.affiliate_conversions conversions ON conversions.id=events.affiliate_conversion_id
		WHERE events.import_run_id=$2`, id, run.ID)
	if err != nil {
		return domain.ReconciliationRun{}, fmt.Errorf("record reconciliation event results: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO commerce.conversion_reconciliation_items
		(reconciliation_run_id,affiliate_conversion_id,result,reason_code)
		SELECT $1,conversions.id,'missing','provider.snapshot_missing'
		FROM commerce.affiliate_conversions conversions
		WHERE conversions.provider_configuration_id=$2
		AND conversions.converted_at >= $3 AND conversions.converted_at <= $4
		AND NOT EXISTS (SELECT 1 FROM commerce.conversion_events events
			WHERE events.import_run_id=$5 AND events.external_conversion_id=conversions.external_conversion_id)`,
		id, run.ProviderConfiguration.ID, run.CoverageStart, run.CoverageEnd, run.ID)
	if err != nil {
		return domain.ReconciliationRun{}, fmt.Errorf("record missing reconciliation results: %w", err)
	}
	var matched, missing, conflicting, stale, unresolved int
	if err = tx.QueryRow(ctx, `SELECT
		count(*) FILTER(WHERE result='matched'),count(*) FILTER(WHERE result='missing'),
		count(*) FILTER(WHERE result='conflicting'),count(*) FILTER(WHERE result='stale'),
		count(*) FILTER(WHERE result='unresolved')
		FROM commerce.conversion_reconciliation_items WHERE reconciliation_run_id=$1`, id).Scan(
		&matched, &missing, &conflicting, &stale, &unresolved); err != nil {
		return domain.ReconciliationRun{}, err
	}
	status := "succeeded"
	if missing+conflicting+unresolved > 0 {
		status = "partial"
	}
	_, err = tx.Exec(ctx, `UPDATE commerce.conversion_reconciliation_runs SET status=$2,
		matched_count=$3,missing_count=$4,conflicting_count=$5,stale_count=$6,
		unresolved_count=$7,completed_at=$8 WHERE id=$1`, id, status, matched, missing,
		conflicting, stale, unresolved, now)
	if err != nil {
		return domain.ReconciliationRun{}, err
	}
	if actor != nil {
		if err = insertOperationAudit(ctx, tx, *actor, "conversion.reconcile", "conversion_import", string(importID), map[string]any{"reconciliation_id": id}); err != nil {
			return domain.ReconciliationRun{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.ReconciliationRun{}, err
	}
	return repository.getReconciliation(ctx, id)
}

func (repository *Repository) ListConversionImports(ctx context.Context, limit, offset int) ([]domain.ConversionImportRun, int64, error) {
	rows, err := repository.pool.Query(ctx, conversionImportSelect+` ORDER BY runs.created_at DESC,runs.id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.ConversionImportRun, 0)
	var total int64
	for rows.Next() {
		item, count, scanErr := scanConversionImport(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		total = count
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (repository *Repository) ListReconciliations(ctx context.Context, limit, offset int) ([]domain.ReconciliationRun, int64, error) {
	rows, err := repository.pool.Query(ctx, reconciliationSelect+` ORDER BY reconciliations.created_at DESC,reconciliations.id LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]domain.ReconciliationRun, 0)
	var total int64
	for rows.Next() {
		item, count, scanErr := scanReconciliation(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		total = count
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (repository *Repository) MonetizationReport(ctx context.Context, start, end time.Time) (domain.MonetizationReport, error) {
	report := emptyMonetizationReport(start, end)
	var coverageCount int64
	if err := repository.pool.QueryRow(ctx, `SELECT count(*),max(coverage_end) FROM commerce.conversion_reconciliation_runs
		WHERE status='succeeded' AND coverage_start<=$1 AND coverage_end>=$2`, start, end).Scan(&coverageCount, &report.FreshThrough); err != nil {
		return report, err
	}
	if coverageCount == 0 {
		return report, nil
	}
	var confirmed, clicks, reversed, previouslyApproved, visitors, recommendations, repeatUsers, observedUsers int64
	err := repository.pool.QueryRow(ctx, `WITH covered AS (
		SELECT DISTINCT provider_configuration_id FROM commerce.conversion_reconciliation_runs
		WHERE status='succeeded' AND coverage_start<=$1 AND coverage_end>=$2
	), eligible_clicks AS (
		SELECT DISTINCT clicks.id,clicks.recommendation_id FROM commerce.affiliate_clicks clicks
		JOIN commerce.merchant_offers offers ON offers.id=clicks.merchant_offer_id
		JOIN commerce.provider_configurations configurations ON configurations.merchant_id=offers.merchant_id
		JOIN covered ON covered.provider_configuration_id=configurations.id
		WHERE clicks.is_countable AND clicks.occurred_at>=$1 AND clicks.occurred_at<$2
	), confirmed AS (
		SELECT DISTINCT conversions.id FROM commerce.affiliate_conversions conversions JOIN covered USING(provider_configuration_id)
		WHERE conversions.verification_state='verified' AND conversions.attribution_status='attributed'
		AND conversions.order_status='confirmed' AND conversions.converted_at>=$1 AND conversions.converted_at<$2
	), approved_history AS (
		SELECT DISTINCT events.affiliate_conversion_id FROM commerce.conversion_events events JOIN covered USING(provider_configuration_id)
		WHERE events.commission_status IN ('approved','paid') AND events.event_timestamp>=$1 AND events.event_timestamp<$2
	), activity AS (
		SELECT user_id,count(DISTINCT occurred_at::date) days FROM analytics.events
		WHERE is_reportable AND user_id IS NOT NULL AND event_name='page_view' AND occurred_at>=$1 AND occurred_at<$2 GROUP BY user_id
	)
	SELECT (SELECT count(*) FROM confirmed),(SELECT count(*) FROM eligible_clicks),
	(SELECT count(*) FROM commerce.affiliate_conversions conversions JOIN covered USING(provider_configuration_id)
	 WHERE conversions.verification_state='verified' AND conversions.commission_status='reversed'
	 AND conversions.latest_event_at>=$1 AND conversions.latest_event_at<$2),
	(SELECT count(*) FROM approved_history),
	(SELECT count(DISTINCT session_id) FROM analytics.events WHERE is_reportable AND event_name='page_view' AND session_id IS NOT NULL AND occurred_at>=$1 AND occurred_at<$2),
	(SELECT count(DISTINCT recommendation_id) FROM eligible_clicks WHERE recommendation_id IS NOT NULL),
	(SELECT count(*) FROM activity WHERE days>=2),(SELECT count(*) FROM activity)`, start, end).Scan(
		&confirmed, &clicks, &reversed, &previouslyApproved, &visitors, &recommendations, &repeatUsers, &observedUsers)
	if err != nil {
		return report, err
	}
	report.ConversionRate = ratioMetric(confirmed, clicks, "Confirmed attributed conversions / eligible human clicks")
	report.ReversalRate = ratioMetric(reversed, previouslyApproved, "Currently reversed commission / conversions previously approved or paid")
	report.RepeatUserRate = ratioMetric(repeatUsers, observedUsers, "Authenticated users active on at least two dates / observed authenticated users")
	report.Commission, err = repository.currencyMetrics(ctx, start, end, 0, "Verified approved or paid commission by original currency")
	if err != nil {
		return report, err
	}
	report.EarningsPerClick, err = repository.currencyMetrics(ctx, start, end, clicks, "Verified approved or paid commission / eligible human clicks")
	if err != nil {
		return report, err
	}
	report.RevenuePerVisitor, err = repository.currencyMetrics(ctx, start, end, visitors, "Verified approved or paid commission / observed visitors")
	if err != nil {
		return report, err
	}
	report.RevenuePerRecommendation, err = repository.currencyMetrics(ctx, start, end, recommendations, "Verified approved or paid commission / recommendations with eligible clicks")
	return report, err
}

func (repository *Repository) currencyMetrics(ctx context.Context, start, end time.Time, denominator int64, definition string) (domain.CurrencyMetricGroup, error) {
	group := domain.CurrencyMetricGroup{Status: domain.MetricAvailable, Values: []domain.CurrencyMetric{}, Definition: definition}
	rows, err := repository.pool.Query(ctx, `WITH covered AS (
		SELECT DISTINCT provider_configuration_id FROM commerce.conversion_reconciliation_runs
		WHERE status='succeeded' AND coverage_start<=$1 AND coverage_end>=$2)
		SELECT commission_currency,sum(commission_amount_minor) FROM commerce.affiliate_conversions
		JOIN covered USING(provider_configuration_id)
		WHERE verification_state='verified' AND commission_status IN ('approved','paid')
		AND order_status='confirmed' AND latest_event_at>=$1 AND latest_event_at<$2
		AND commission_amount_minor IS NOT NULL GROUP BY commission_currency ORDER BY commission_currency`, start, end)
	if err != nil {
		return group, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric domain.CurrencyMetric
		if err = rows.Scan(&metric.Currency, &metric.AmountMinor); err != nil {
			return group, err
		}
		metric.Denominator = denominator
		if denominator > 0 {
			value := float64(metric.AmountMinor) / float64(denominator)
			metric.ValueMinor = &value
		}
		group.Values = append(group.Values, metric)
	}
	if denominator == 0 && definition != "Verified approved or paid commission by original currency" {
		group.Status = domain.MetricInsufficient
	} else if len(group.Values) == 0 {
		group.Status = domain.MetricAvailable
	}
	return group, rows.Err()
}

func emptyMonetizationReport(start, end time.Time) domain.MonetizationReport {
	return domain.MonetizationReport{WindowStart: start, WindowEnd: end,
		ConversionRate:           ratioNoData("Confirmed attributed conversions / eligible human clicks"),
		EarningsPerClick:         domain.CurrencyMetricGroup{Status: domain.MetricNoData, Values: []domain.CurrencyMetric{}, Definition: "Verified approved or paid commission / eligible human clicks"},
		RevenuePerVisitor:        domain.CurrencyMetricGroup{Status: domain.MetricNoData, Values: []domain.CurrencyMetric{}, Definition: "Verified approved or paid commission / observed visitors"},
		RevenuePerRecommendation: domain.CurrencyMetricGroup{Status: domain.MetricNoData, Values: []domain.CurrencyMetric{}, Definition: "Verified approved or paid commission / recommendations with eligible clicks"},
		Commission:               domain.CurrencyMetricGroup{Status: domain.MetricNoData, Values: []domain.CurrencyMetric{}, Definition: "Verified approved or paid commission by original currency"},
		ReversalRate:             ratioNoData("Currently reversed commission / conversions previously approved or paid"),
		RepeatUserRate:           ratioNoData("Authenticated users active on at least two dates / observed authenticated users"),
		CurrencyPolicy:           "Original provider currencies remain separate; no FX conversion is performed."}
}

func ratioMetric(numerator, denominator int64, definition string) domain.RatioMetric {
	metric := domain.RatioMetric{Status: domain.MetricInsufficient, Numerator: numerator, Denominator: denominator, Definition: definition}
	if denominator > 0 {
		value := float64(numerator) / float64(denominator)
		metric.Status = domain.MetricAvailable
		metric.Value = &value
	}
	return metric
}

func ratioNoData(definition string) domain.RatioMetric {
	return domain.RatioMetric{Status: domain.MetricNoData, Definition: definition}
}

func (repository *Repository) getConversionImport(ctx context.Context, id domain.ConversionImportRunID) (domain.ConversionImportRun, error) {
	run, _, err := scanConversionImport(repository.pool.QueryRow(ctx, conversionImportSelect+` WHERE runs.id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConversionImportRun{}, ports.ErrImportNotFound
	}
	return run, err
}

const conversionImportSelect = `SELECT count(*) OVER(),runs.id,runs.trigger_type,runs.status,
	runs.attempt_count,runs.max_attempts,runs.records_received,runs.records_applied,runs.records_rejected,
	runs.cursor_before,runs.cursor_after,runs.coverage_start,runs.coverage_end,runs.error_code,
	runs.error_message,runs.created_at,runs.started_at,runs.completed_at,
	configurations.id,configurations.merchant_id,merchants.name,configurations.provider_key,
	configurations.adapter_key,configurations.external_merchant_id,configurations.credential_reference,
	configurations.lifecycle_status,configurations.configuration_verified_at,
	configurations.schedule_interval_minutes,configurations.freshness_ttl_minutes,
	configurations.cursor_value,configurations.next_import_at,configurations.last_import_started_at,
	configurations.last_import_succeeded_at,configurations.last_import_failed_at,
	configurations.consecutive_failures,configurations.last_error_code,
	configurations.conversion_cursor_value,configurations.next_conversion_import_at,
	configurations.last_conversion_import_succeeded_at,configurations.last_conversion_import_failed_at,
	configurations.conversion_consecutive_failures,configurations.last_conversion_error_code,
	configurations.conversion_ingestion_enabled,configurations.conversion_configuration_verified_at,
	configurations.created_at,configurations.updated_at
	FROM commerce.conversion_import_runs runs
	JOIN commerce.provider_configurations configurations ON configurations.id=runs.provider_configuration_id
	JOIN commerce.merchants merchants ON merchants.id=configurations.merchant_id`

func scanConversionImport(row rowScanner) (domain.ConversionImportRun, int64, error) {
	var run domain.ConversionImportRun
	var count int64
	err := row.Scan(&count, &run.ID, &run.Trigger, &run.Status, &run.AttemptCount, &run.MaxAttempts,
		&run.RecordsReceived, &run.RecordsApplied, &run.RecordsRejected, &run.CursorBefore, &run.CursorAfter,
		&run.CoverageStart, &run.CoverageEnd, &run.ErrorCode, &run.ErrorMessage, &run.CreatedAt,
		&run.StartedAt, &run.CompletedAt, &run.ProviderConfiguration.ID, &run.ProviderConfiguration.MerchantID,
		&run.ProviderConfiguration.MerchantName, &run.ProviderConfiguration.ProviderKey,
		&run.ProviderConfiguration.AdapterKey, &run.ProviderConfiguration.ExternalMerchantID,
		&run.ProviderConfiguration.CredentialReference, &run.ProviderConfiguration.LifecycleStatus,
		&run.ProviderConfiguration.ConfigurationVerifiedAt, &run.ProviderConfiguration.ScheduleIntervalMinutes,
		&run.ProviderConfiguration.FreshnessTTLMinutes, &run.ProviderConfiguration.Cursor,
		&run.ProviderConfiguration.NextImportAt, &run.ProviderConfiguration.LastImportStartedAt,
		&run.ProviderConfiguration.LastImportSucceededAt, &run.ProviderConfiguration.LastImportFailedAt,
		&run.ProviderConfiguration.ConsecutiveFailures, &run.ProviderConfiguration.LastErrorCode,
		&run.ProviderConfiguration.ConversionCursor, &run.ProviderConfiguration.NextConversionImportAt,
		&run.ProviderConfiguration.LastConversionSucceeded, &run.ProviderConfiguration.LastConversionFailed,
		&run.ProviderConfiguration.ConversionFailures, &run.ProviderConfiguration.LastConversionError,
		&run.ProviderConfiguration.ConversionEnabled, &run.ProviderConfiguration.ConversionVerifiedAt,
		&run.ProviderConfiguration.CreatedAt, &run.ProviderConfiguration.UpdatedAt)
	return run, count, err
}

const reconciliationSelect = `SELECT count(*) OVER(),reconciliations.id,reconciliations.status,
	reconciliations.coverage_start,reconciliations.coverage_end,reconciliations.matched_count,
	reconciliations.missing_count,reconciliations.conflicting_count,reconciliations.stale_count,
	reconciliations.unresolved_count,reconciliations.error_code,reconciliations.started_at,
	reconciliations.completed_at,configurations.id,configurations.merchant_id,merchants.name,
	configurations.provider_key,configurations.adapter_key
	FROM commerce.conversion_reconciliation_runs reconciliations
	JOIN commerce.provider_configurations configurations ON configurations.id=reconciliations.provider_configuration_id
	JOIN commerce.merchants merchants ON merchants.id=configurations.merchant_id`

func scanReconciliation(row rowScanner) (domain.ReconciliationRun, int64, error) {
	var run domain.ReconciliationRun
	var count int64
	err := row.Scan(&count, &run.ID, &run.Status, &run.CoverageStart, &run.CoverageEnd, &run.Matched,
		&run.Missing, &run.Conflicting, &run.Stale, &run.Unresolved, &run.ErrorCode, &run.StartedAt,
		&run.CompletedAt, &run.ProviderConfiguration.ID, &run.ProviderConfiguration.MerchantID,
		&run.ProviderConfiguration.MerchantName, &run.ProviderConfiguration.ProviderKey,
		&run.ProviderConfiguration.AdapterKey)
	return run, count, err
}

func (repository *Repository) getReconciliation(ctx context.Context, id domain.ReconciliationRunID) (domain.ReconciliationRun, error) {
	run, _, err := scanReconciliation(repository.pool.QueryRow(ctx, reconciliationSelect+` WHERE reconciliations.id=$1`, id))
	return run, err
}

var _ ports.ConversionRepository = (*Repository)(nil)
