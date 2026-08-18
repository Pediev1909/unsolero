-- Verified conversion facts and monetization reporting. Provider events are
-- immutable commercial facts and remain downstream of recommendations.

ALTER TABLE commerce.provider_configurations
    ADD COLUMN conversion_ingestion_enabled boolean NOT NULL DEFAULT false,
    ADD COLUMN conversion_configuration_verified_at timestamptz,
    ADD COLUMN conversion_cursor_value text CHECK (conversion_cursor_value IS NULL OR char_length(conversion_cursor_value) <= 2000),
    ADD COLUMN next_conversion_import_at timestamptz,
    ADD COLUMN last_conversion_import_started_at timestamptz,
    ADD COLUMN last_conversion_import_succeeded_at timestamptz,
    ADD COLUMN last_conversion_import_failed_at timestamptz,
    ADD COLUMN conversion_consecutive_failures integer NOT NULL DEFAULT 0 CHECK (conversion_consecutive_failures >= 0),
    ADD COLUMN last_conversion_error_code text CHECK (last_conversion_error_code IS NULL OR last_conversion_error_code ~ '^[a-z][a-z0-9_.-]{0,99}$');

CREATE INDEX provider_configurations_conversion_due_idx
    ON commerce.provider_configurations (next_conversion_import_at, id)
    WHERE lifecycle_status IN ('active', 'degraded') AND conversion_ingestion_enabled
      AND next_conversion_import_at IS NOT NULL;

ALTER TABLE commerce.provider_configurations
    ADD CONSTRAINT provider_configurations_conversion_enabled_check CHECK (
        NOT conversion_ingestion_enabled OR
        (credential_reference IS NOT NULL AND conversion_configuration_verified_at IS NOT NULL)
    );

CREATE TABLE commerce.conversion_webhook_deliveries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    request_fingerprint character(64) NOT NULL CHECK (request_fingerprint ~ '^[a-f0-9]{64}$'),
    body_fingerprint character(64) NOT NULL CHECK (body_fingerprint ~ '^[a-f0-9]{64}$'),
    verification_state text NOT NULL CHECK (verification_state IN ('verified', 'rejected', 'replayed')),
    signature_timestamp timestamptz,
    provider_event_count integer NOT NULL DEFAULT 0 CHECK (provider_event_count >= 0),
    error_code text CHECK (error_code IS NULL OR error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    received_at timestamptz NOT NULL,
    processed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_configuration_id, request_fingerprint),
    CHECK (processed_at IS NULL OR processed_at >= received_at),
    CHECK (verification_state <> 'verified' OR error_code IS NULL)
);
CREATE INDEX conversion_webhook_deliveries_provider_received_idx
    ON commerce.conversion_webhook_deliveries (provider_configuration_id, received_at DESC, id);

CREATE TABLE commerce.conversion_import_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    trigger_type text NOT NULL CHECK (trigger_type IN ('scheduled', 'manual', 'retry')),
    status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'running', 'retry_wait', 'succeeded', 'partial', 'failed', 'cancelled')),
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 200),
    requested_by uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    cursor_before text CHECK (cursor_before IS NULL OR char_length(cursor_before) <= 2000),
    cursor_after text CHECK (cursor_after IS NULL OR char_length(cursor_after) <= 2000),
    attempt_count smallint NOT NULL DEFAULT 0 CHECK (attempt_count BETWEEN 0 AND 10),
    max_attempts smallint NOT NULL DEFAULT 3 CHECK (max_attempts BETWEEN 1 AND 10),
    records_received integer NOT NULL DEFAULT 0 CHECK (records_received >= 0),
    records_applied integer NOT NULL DEFAULT 0 CHECK (records_applied >= 0),
    records_rejected integer NOT NULL DEFAULT 0 CHECK (records_rejected >= 0),
    coverage_start timestamptz,
    coverage_end timestamptz,
    error_code text CHECK (error_code IS NULL OR error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    error_message text CHECK (error_message IS NULL OR char_length(error_message) <= 500),
    next_retry_at timestamptz,
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_configuration_id, idempotency_key),
    CHECK (records_applied + records_rejected <= records_received),
    CHECK ((coverage_start IS NULL) = (coverage_end IS NULL)),
    CHECK (coverage_end IS NULL OR coverage_end >= coverage_start),
    CHECK (completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at),
    CHECK (status <> 'retry_wait' OR next_retry_at IS NOT NULL)
);
CREATE INDEX conversion_import_runs_queue_idx
    ON commerce.conversion_import_runs (COALESCE(next_retry_at, created_at), id)
    WHERE status IN ('queued', 'retry_wait');
CREATE INDEX conversion_import_runs_provider_created_idx
    ON commerce.conversion_import_runs (provider_configuration_id, created_at DESC);

CREATE TABLE commerce.conversion_import_failures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    import_run_id uuid NOT NULL REFERENCES commerce.conversion_import_runs(id) ON DELETE CASCADE,
    provider_event_id text CHECK (provider_event_id IS NULL OR char_length(btrim(provider_event_id)) BETWEEN 1 AND 200),
    error_code text NOT NULL CHECK (error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    error_message text NOT NULL CHECK (char_length(btrim(error_message)) BETWEEN 1 AND 500),
    record_fingerprint character(64) CHECK (record_fingerprint IS NULL OR record_fingerprint ~ '^[a-f0-9]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (import_run_id, provider_event_id, error_code)
);

ALTER TABLE commerce.affiliate_conversions RENAME COLUMN status TO order_status;
ALTER TABLE commerce.affiliate_conversions
    DROP CONSTRAINT affiliate_conversions_status_check,
    DROP CONSTRAINT affiliate_conversions_provider_external_conversion_id_key,
    ALTER COLUMN affiliate_click_id DROP NOT NULL,
    ADD COLUMN provider_configuration_id uuid REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    ADD COLUMN merchant_id uuid REFERENCES commerce.merchants(id) ON DELETE RESTRICT,
    ADD COLUMN recommendation_id uuid REFERENCES recommendation.recommendations(id) ON DELETE SET NULL,
    ADD COLUMN recommendation_item_id uuid REFERENCES recommendation.recommendation_items(id) ON DELETE SET NULL,
    ADD COLUMN event_type text NOT NULL DEFAULT 'legacy',
    ADD COLUMN event_received_at timestamptz NOT NULL DEFAULT now(),
    ADD COLUMN commission_status text,
    ADD COLUMN attribution_status text NOT NULL DEFAULT 'unattributed',
    ADD COLUMN source text,
    ADD COLUMN campaign text,
    ADD COLUMN raw_provider_reference text,
    ADD COLUMN verification_state text NOT NULL DEFAULT 'unverified',
    ADD COLUMN latest_event_at timestamptz,
    ADD COLUMN cancelled_at timestamptz,
    ADD COLUMN reversed_at timestamptz,
    ADD CONSTRAINT affiliate_conversions_order_status_check CHECK (order_status IN ('pending', 'confirmed', 'cancelled', 'reversed', 'rejected')),
    ADD CONSTRAINT affiliate_conversions_commission_status_check CHECK (commission_status IS NULL OR commission_status IN ('pending', 'approved', 'reversed', 'rejected', 'paid')),
    ADD CONSTRAINT affiliate_conversions_attribution_status_check CHECK (attribution_status IN ('attributed', 'unattributed')),
    ADD CONSTRAINT affiliate_conversions_verification_state_check CHECK (verification_state IN ('unverified', 'verified')),
    ADD CONSTRAINT affiliate_conversions_source_check CHECK (source IS NULL OR source ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    ADD CONSTRAINT affiliate_conversions_campaign_check CHECK (campaign IS NULL OR campaign ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    ADD CONSTRAINT affiliate_conversions_raw_reference_check CHECK (raw_provider_reference IS NULL OR char_length(btrim(raw_provider_reference)) BETWEEN 1 AND 200),
    ADD CONSTRAINT affiliate_conversions_verified_shape_check CHECK (verification_state <> 'verified' OR (provider_configuration_id IS NOT NULL AND merchant_id IS NOT NULL)),
    ADD CONSTRAINT affiliate_conversions_attribution_shape_check CHECK ((attribution_status = 'attributed' AND affiliate_click_id IS NOT NULL) OR (attribution_status = 'unattributed' AND affiliate_click_id IS NULL AND recommendation_id IS NULL AND recommendation_item_id IS NULL)),
    ADD CONSTRAINT affiliate_conversions_order_value_bound_check CHECK (order_value_minor IS NULL OR order_value_minor <= 1000000000000000),
    ADD CONSTRAINT affiliate_conversions_commission_bound_check CHECK (commission_amount_minor IS NULL OR commission_amount_minor <= 1000000000000000),
    ADD CONSTRAINT affiliate_conversions_cancelled_at_check CHECK (cancelled_at IS NULL OR cancelled_at >= converted_at),
    ADD CONSTRAINT affiliate_conversions_reversed_at_check CHECK (reversed_at IS NULL OR reversed_at >= converted_at);

CREATE UNIQUE INDEX affiliate_conversions_configuration_external_idx
    ON commerce.affiliate_conversions (provider_configuration_id, external_conversion_id)
    WHERE provider_configuration_id IS NOT NULL;
CREATE INDEX affiliate_conversions_verified_event_idx
    ON commerce.affiliate_conversions (latest_event_at DESC, id) WHERE verification_state = 'verified';
CREATE INDEX affiliate_conversions_merchant_status_idx
    ON commerce.affiliate_conversions (merchant_id, order_status, converted_at DESC) WHERE verification_state = 'verified';

CREATE TABLE commerce.conversion_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    webhook_delivery_id uuid REFERENCES commerce.conversion_webhook_deliveries(id) ON DELETE RESTRICT,
    import_run_id uuid REFERENCES commerce.conversion_import_runs(id) ON DELETE RESTRICT,
    affiliate_conversion_id uuid REFERENCES commerce.affiliate_conversions(id) ON DELETE RESTRICT,
    provider text NOT NULL CHECK (provider ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    provider_event_id text NOT NULL CHECK (char_length(btrim(provider_event_id)) BETWEEN 1 AND 200),
    event_type text NOT NULL CHECK (event_type IN ('conversion_created', 'order_status_changed', 'commission_status_changed', 'cancelled', 'reversed', 'correction')),
    external_conversion_id text NOT NULL CHECK (char_length(btrim(external_conversion_id)) BETWEEN 1 AND 200),
    order_reference text CHECK (order_reference IS NULL OR char_length(btrim(order_reference)) BETWEEN 1 AND 200),
    order_status text NOT NULL CHECK (order_status IN ('pending', 'confirmed', 'cancelled', 'reversed', 'rejected')),
    order_value_minor bigint CHECK (order_value_minor IS NULL OR order_value_minor BETWEEN 0 AND 1000000000000000),
    order_currency character(3) CHECK (order_currency IS NULL OR order_currency ~ '^[A-Z]{3}$'),
    commission_amount_minor bigint CHECK (commission_amount_minor IS NULL OR commission_amount_minor BETWEEN 0 AND 1000000000000000),
    commission_currency character(3) CHECK (commission_currency IS NULL OR commission_currency ~ '^[A-Z]{3}$'),
    commission_status text CHECK (commission_status IS NULL OR commission_status IN ('pending', 'approved', 'reversed', 'rejected', 'paid')),
    affiliate_click_id uuid REFERENCES commerce.affiliate_clicks(id) ON DELETE RESTRICT,
    attribution_status text NOT NULL CHECK (attribution_status IN ('attributed', 'unattributed')),
    source text CHECK (source IS NULL OR source ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    campaign text CHECK (campaign IS NULL OR campaign ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    raw_provider_reference text CHECK (raw_provider_reference IS NULL OR char_length(btrim(raw_provider_reference)) BETWEEN 1 AND 200),
    event_timestamp timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    verification_state text NOT NULL CHECK (verification_state = 'verified'),
    payload_fingerprint character(64) NOT NULL CHECK (payload_fingerprint ~ '^[a-f0-9]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_configuration_id, provider_event_id),
    CHECK ((order_value_minor IS NULL) = (order_currency IS NULL)),
    CHECK ((commission_amount_minor IS NULL) = (commission_currency IS NULL)),
    CHECK ((webhook_delivery_id IS NOT NULL)::integer + (import_run_id IS NOT NULL)::integer = 1),
    CHECK ((attribution_status = 'attributed') = (affiliate_click_id IS NOT NULL)),
    CHECK (event_timestamp <= received_at + interval '5 minutes')
);
CREATE INDEX conversion_events_conversion_timestamp_idx
    ON commerce.conversion_events (affiliate_conversion_id, event_timestamp DESC, id);
CREATE INDEX conversion_events_provider_timestamp_idx
    ON commerce.conversion_events (provider_configuration_id, event_timestamp DESC, id);
CREATE INDEX conversion_events_commission_status_idx
    ON commerce.conversion_events (commission_status, event_timestamp DESC) WHERE commission_status IS NOT NULL;

ALTER TABLE commerce.affiliate_conversions
    ADD COLUMN latest_event_id uuid REFERENCES commerce.conversion_events(id) ON DELETE RESTRICT;

CREATE TABLE commerce.conversion_reconciliation_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    conversion_import_run_id uuid REFERENCES commerce.conversion_import_runs(id) ON DELETE RESTRICT,
    requested_by uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 200),
    status text NOT NULL CHECK (status IN ('running', 'succeeded', 'partial', 'failed')),
    coverage_start timestamptz NOT NULL,
    coverage_end timestamptz NOT NULL,
    matched_count integer NOT NULL DEFAULT 0 CHECK (matched_count >= 0),
    missing_count integer NOT NULL DEFAULT 0 CHECK (missing_count >= 0),
    conflicting_count integer NOT NULL DEFAULT 0 CHECK (conflicting_count >= 0),
    stale_count integer NOT NULL DEFAULT 0 CHECK (stale_count >= 0),
    unresolved_count integer NOT NULL DEFAULT 0 CHECK (unresolved_count >= 0),
    error_code text CHECK (error_code IS NULL OR error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_configuration_id, idempotency_key),
    CHECK (coverage_end >= coverage_start),
    CHECK (completed_at IS NULL OR completed_at >= started_at),
    CHECK (status = 'running' OR completed_at IS NOT NULL)
);
CREATE INDEX conversion_reconciliation_coverage_idx
    ON commerce.conversion_reconciliation_runs (provider_configuration_id, coverage_start, coverage_end, completed_at DESC)
    WHERE status = 'succeeded';

CREATE TABLE commerce.conversion_reconciliation_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reconciliation_run_id uuid NOT NULL REFERENCES commerce.conversion_reconciliation_runs(id) ON DELETE RESTRICT,
    affiliate_conversion_id uuid REFERENCES commerce.affiliate_conversions(id) ON DELETE RESTRICT,
    provider_event_id text CHECK (provider_event_id IS NULL OR char_length(btrim(provider_event_id)) BETWEEN 1 AND 200),
    result text NOT NULL CHECK (result IN ('matched', 'missing', 'conflicting', 'stale', 'unresolved')),
    reason_code text NOT NULL CHECK (reason_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    comparison_fingerprint character(64) CHECK (comparison_fingerprint IS NULL OR comparison_fingerprint ~ '^[a-f0-9]{64}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (affiliate_conversion_id IS NOT NULL OR provider_event_id IS NOT NULL)
);
CREATE INDEX conversion_reconciliation_items_run_result_idx
    ON commerce.conversion_reconciliation_items (reconciliation_run_id, result, id);

CREATE FUNCTION commerce.reject_conversion_fact_mutation()
RETURNS trigger LANGUAGE plpgsql AS $function$
BEGIN
    RAISE EXCEPTION 'verified conversion facts are immutable' USING ERRCODE = '55000';
END;
$function$;
CREATE TRIGGER reject_conversion_event_mutation
BEFORE UPDATE OR DELETE ON commerce.conversion_events
FOR EACH ROW EXECUTE FUNCTION commerce.reject_conversion_fact_mutation();
CREATE TRIGGER reject_reconciliation_item_mutation
BEFORE UPDATE OR DELETE ON commerce.conversion_reconciliation_items
FOR EACH ROW EXECUTE FUNCTION commerce.reject_conversion_fact_mutation();

COMMENT ON TABLE commerce.conversion_events IS
    'Immutable provider-verified financial facts. They must never be used by recommendation policy or scoring.';
COMMENT ON TABLE commerce.conversion_webhook_deliveries IS
    'Authentication/replay audit metadata only. Raw webhook payloads and secrets are never stored.';
COMMENT ON COLUMN commerce.affiliate_conversions.verification_state IS
    'Only verified projections are eligible for monetization reporting.';
