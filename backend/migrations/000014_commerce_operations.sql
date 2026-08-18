-- Provider-neutral merchant ingestion and durable affiliate operations.
-- Commerce remains downstream of recommendation decisions. No table in this
-- migration is referenced by recommendation policy or scoring.

CREATE TABLE commerce.provider_configurations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id uuid NOT NULL REFERENCES commerce.merchants(id) ON DELETE RESTRICT,
    provider_key text NOT NULL CHECK (provider_key ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    adapter_key text NOT NULL CHECK (adapter_key ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    external_merchant_id text NOT NULL
        CHECK (char_length(btrim(external_merchant_id)) BETWEEN 1 AND 200),
    credential_reference text CHECK (
        credential_reference IS NULL OR
        credential_reference ~ '^[A-Za-z0-9][A-Za-z0-9._:/-]{0,199}$'
    ),
    lifecycle_status text NOT NULL DEFAULT 'disabled'
        CHECK (lifecycle_status IN ('disabled', 'configured', 'active', 'degraded', 'suspended')),
    configuration_verified_at timestamptz,
    schedule_interval_minutes integer NOT NULL DEFAULT 360
        CHECK (schedule_interval_minutes BETWEEN 5 AND 10080),
    freshness_ttl_minutes integer NOT NULL DEFAULT 4320
        CHECK (freshness_ttl_minutes BETWEEN 60 AND 43200),
    cursor_value text CHECK (cursor_value IS NULL OR char_length(cursor_value) <= 2000),
    next_import_at timestamptz,
    last_import_started_at timestamptz,
    last_import_succeeded_at timestamptz,
    last_import_failed_at timestamptz,
    consecutive_failures integer NOT NULL DEFAULT 0 CHECK (consecutive_failures >= 0),
    last_error_code text CHECK (
        last_error_code IS NULL OR last_error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_key, merchant_id, external_merchant_id),
    CHECK (
        lifecycle_status NOT IN ('active', 'degraded')
        OR (credential_reference IS NOT NULL AND configuration_verified_at IS NOT NULL)
    )
);

CREATE INDEX provider_configurations_due_idx
    ON commerce.provider_configurations (next_import_at, id)
    WHERE lifecycle_status IN ('active', 'degraded') AND next_import_at IS NOT NULL;

CREATE TABLE commerce.offer_import_runs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL
        REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    trigger_type text NOT NULL CHECK (trigger_type IN ('scheduled', 'manual', 'retry')),
    status text NOT NULL DEFAULT 'queued'
        CHECK (status IN ('queued', 'running', 'retry_wait', 'succeeded', 'partial', 'failed', 'cancelled')),
    idempotency_key text NOT NULL CHECK (char_length(idempotency_key) BETWEEN 8 AND 200),
    requested_by uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    cursor_before text CHECK (cursor_before IS NULL OR char_length(cursor_before) <= 2000),
    cursor_after text CHECK (cursor_after IS NULL OR char_length(cursor_after) <= 2000),
    attempt_count smallint NOT NULL DEFAULT 0 CHECK (attempt_count BETWEEN 0 AND 10),
    max_attempts smallint NOT NULL DEFAULT 3 CHECK (max_attempts BETWEEN 1 AND 10),
    records_received integer NOT NULL DEFAULT 0 CHECK (records_received >= 0),
    records_applied integer NOT NULL DEFAULT 0 CHECK (records_applied >= 0),
    records_rejected integer NOT NULL DEFAULT 0 CHECK (records_rejected >= 0),
    offers_deactivated integer NOT NULL DEFAULT 0 CHECK (offers_deactivated >= 0),
    error_code text CHECK (error_code IS NULL OR error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    error_message text CHECK (error_message IS NULL OR char_length(error_message) <= 500),
    next_retry_at timestamptz,
    started_at timestamptz,
    completed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider_configuration_id, idempotency_key),
    CHECK (records_applied + records_rejected <= records_received),
    CHECK (completed_at IS NULL OR started_at IS NULL OR completed_at >= started_at),
    CHECK (status <> 'retry_wait' OR next_retry_at IS NOT NULL)
);

CREATE INDEX offer_import_runs_queue_idx
    ON commerce.offer_import_runs (COALESCE(next_retry_at, created_at), id)
    WHERE status IN ('queued', 'retry_wait');
CREATE INDEX offer_import_runs_provider_created_idx
    ON commerce.offer_import_runs (provider_configuration_id, created_at DESC);

CREATE TABLE commerce.offer_import_failures (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    import_run_id uuid NOT NULL REFERENCES commerce.offer_import_runs(id) ON DELETE CASCADE,
    external_record_id text CHECK (
        external_record_id IS NULL OR char_length(btrim(external_record_id)) BETWEEN 1 AND 200
    ),
    error_code text NOT NULL CHECK (error_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    error_message text NOT NULL CHECK (char_length(btrim(error_message)) BETWEEN 1 AND 500),
    record_fingerprint character(64) CHECK (
        record_fingerprint IS NULL OR record_fingerprint ~ '^[a-f0-9]{64}$'
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (import_run_id, external_record_id, error_code)
);

CREATE INDEX offer_import_failures_run_idx
    ON commerce.offer_import_failures (import_run_id, created_at, id);

CREATE TABLE commerce.provider_offer_mappings (
    provider_configuration_id uuid NOT NULL
        REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    external_offer_id text NOT NULL
        CHECK (char_length(btrim(external_offer_id)) BETWEEN 1 AND 200),
    merchant_offer_id uuid NOT NULL REFERENCES commerce.merchant_offers(id) ON DELETE RESTRICT,
    last_seen_import_run_id uuid NOT NULL
        REFERENCES commerce.offer_import_runs(id) ON DELETE RESTRICT,
    first_seen_at timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    is_present boolean NOT NULL DEFAULT true,
    PRIMARY KEY (provider_configuration_id, external_offer_id),
    UNIQUE (provider_configuration_id, merchant_offer_id),
    CHECK (last_seen_at >= first_seen_at)
);

ALTER TABLE commerce.merchant_offers
    ADD COLUMN provider_observed_at timestamptz,
    ADD COLUMN imported_at timestamptz,
    ADD COLUMN expires_at timestamptz,
    ADD COLUMN last_import_run_id uuid
        REFERENCES commerce.offer_import_runs(id) ON DELETE SET NULL,
    ADD CONSTRAINT merchant_offers_expiry_check CHECK (
        expires_at IS NULL OR expires_at > COALESCE(provider_observed_at, last_checked_at)
    );

CREATE INDEX merchant_offers_expiry_idx
    ON commerce.merchant_offers (expires_at, product_id)
    WHERE is_active;

CREATE TABLE commerce.price_observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL
        REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    import_run_id uuid NOT NULL REFERENCES commerce.offer_import_runs(id) ON DELETE RESTRICT,
    merchant_offer_id uuid NOT NULL REFERENCES commerce.merchant_offers(id) ON DELETE RESTRICT,
    external_offer_id text NOT NULL
        CHECK (char_length(btrim(external_offer_id)) BETWEEN 1 AND 200),
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    shipping_minor bigint NOT NULL CHECK (shipping_minor >= 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    provider_observed_at timestamptz,
    observed_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    imported_at timestamptz NOT NULL DEFAULT now(),
    observation_fingerprint character(64) NOT NULL
        CHECK (observation_fingerprint ~ '^[a-f0-9]{64}$'),
    UNIQUE (provider_configuration_id, observation_fingerprint),
    CHECK (expires_at > observed_at)
);

CREATE INDEX price_observations_offer_observed_idx
    ON commerce.price_observations (merchant_offer_id, observed_at DESC);

CREATE TABLE commerce.availability_observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_configuration_id uuid NOT NULL
        REFERENCES commerce.provider_configurations(id) ON DELETE RESTRICT,
    import_run_id uuid NOT NULL REFERENCES commerce.offer_import_runs(id) ON DELETE RESTRICT,
    merchant_offer_id uuid NOT NULL REFERENCES commerce.merchant_offers(id) ON DELETE RESTRICT,
    external_offer_id text NOT NULL
        CHECK (char_length(btrim(external_offer_id)) BETWEEN 1 AND 200),
    availability text NOT NULL
        CHECK (availability IN ('in_stock', 'backorder', 'out_of_stock', 'discontinued')),
    provider_observed_at timestamptz,
    observed_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    imported_at timestamptz NOT NULL DEFAULT now(),
    observation_fingerprint character(64) NOT NULL
        CHECK (observation_fingerprint ~ '^[a-f0-9]{64}$'),
    UNIQUE (provider_configuration_id, observation_fingerprint),
    CHECK (expires_at > observed_at)
);

CREATE INDEX availability_observations_offer_observed_idx
    ON commerce.availability_observations (merchant_offer_id, observed_at DESC);

CREATE TABLE commerce.operation_audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    action text NOT NULL CHECK (action ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    entity_type text NOT NULL CHECK (entity_type ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    entity_id uuid NOT NULL,
    request_id text CHECK (request_id IS NULL OR char_length(request_id) BETWEEN 1 AND 128),
    details jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(details) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX operation_audit_entity_idx
    ON commerce.operation_audit_log (entity_type, entity_id, occurred_at DESC);

ALTER TABLE commerce.affiliate_clicks
    ADD COLUMN recommendation_item_id uuid
        REFERENCES recommendation.recommendation_items(id) ON DELETE SET NULL,
    ADD COLUMN classification text NOT NULL DEFAULT 'unknown'
        CHECK (classification IN ('human', 'bot', 'prefetch', 'unknown')),
    ADD COLUMN is_countable boolean NOT NULL DEFAULT false,
    ADD COLUMN user_agent_hash character(64) CHECK (
        user_agent_hash IS NULL OR user_agent_hash ~ '^[a-f0-9]{64}$'
    ),
    ADD COLUMN idempotency_key text CHECK (
        idempotency_key IS NULL OR char_length(idempotency_key) BETWEEN 8 AND 128
    ),
    ADD COLUMN retention_expires_at timestamptz NOT NULL DEFAULT (now() + interval '397 days'),
    ADD COLUMN anonymized_at timestamptz;

ALTER TABLE commerce.affiliate_clicks
    DROP CONSTRAINT affiliate_clicks_actor_check,
    ADD CONSTRAINT affiliate_clicks_actor_check CHECK (
        anonymized_at IS NOT NULL OR user_id IS NOT NULL OR anonymous_id IS NOT NULL
    );

CREATE UNIQUE INDEX affiliate_clicks_idempotency_idx
    ON commerce.affiliate_clicks (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX affiliate_clicks_filtered_occurred_idx
    ON commerce.affiliate_clicks (occurred_at DESC)
    WHERE is_countable;
CREATE INDEX affiliate_clicks_retention_idx
    ON commerce.affiliate_clicks (retention_expires_at, id)
    WHERE anonymized_at IS NULL;

COMMENT ON TABLE commerce.provider_configurations IS
    'Provider lifecycle and cursor state. credential_reference names a secret-manager entry; provider secrets must never be stored here.';
COMMENT ON TABLE commerce.price_observations IS
    'Immutable provider price history. It is downstream commerce data and never a recommendation scoring input.';
COMMENT ON TABLE commerce.availability_observations IS
    'Immutable provider availability history. It is downstream commerce data and never a recommendation scoring input.';
COMMENT ON COLUMN commerce.affiliate_clicks.is_countable IS
    'Filtered analytics eligibility. Raw clicks remain separately auditable.';

CREATE FUNCTION commerce.reject_observation_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $function$
BEGIN
    RAISE EXCEPTION 'commerce observations are immutable'
        USING ERRCODE = '55000';
END;
$function$;

CREATE TRIGGER reject_price_observation_mutation
BEFORE UPDATE OR DELETE ON commerce.price_observations
FOR EACH ROW EXECUTE FUNCTION commerce.reject_observation_mutation();

CREATE TRIGGER reject_availability_observation_mutation
BEFORE UPDATE OR DELETE ON commerce.availability_observations
FOR EACH ROW EXECUTE FUNCTION commerce.reject_observation_mutation();
