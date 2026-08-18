-- Phase 6: privacy-governed analytics. Browser consent is server-authoritative,
-- accepted events are idempotent, and raw ingestion outcomes are separated
-- from reportable facts. Defaults below are engineering defaults and require
-- legal/privacy review before production deployment.

CREATE TABLE analytics.consent_states (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES identity.users(id) ON DELETE RESTRICT,
    anonymous_subject_hash bytea,
    state text NOT NULL CHECK (state IN ('granted', 'denied', 'withdrawn')),
    policy_version text NOT NULL CHECK (policy_version ~ '^[a-z0-9][a-z0-9._-]{0,49}$'),
    source text NOT NULL CHECK (source IN ('banner', 'preferences', 'account_sync')),
    decided_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (octet_length(anonymous_subject_hash) = 32 OR anonymous_subject_hash IS NULL),
    CHECK ((user_id IS NULL) <> (anonymous_subject_hash IS NULL))
);

CREATE UNIQUE INDEX consent_states_user_idx
    ON analytics.consent_states (user_id) WHERE user_id IS NOT NULL;
CREATE UNIQUE INDEX consent_states_anonymous_idx
    ON analytics.consent_states (anonymous_subject_hash)
    WHERE anonymous_subject_hash IS NOT NULL;

CREATE TABLE analytics.consent_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES identity.users(id) ON DELETE RESTRICT,
    anonymous_subject_hash bytea,
    anonymized_subject_id uuid,
    state text NOT NULL CHECK (state IN ('granted', 'denied', 'withdrawn')),
    policy_version text NOT NULL CHECK (policy_version ~ '^[a-z0-9][a-z0-9._-]{0,49}$'),
    source text NOT NULL CHECK (source IN ('banner', 'preferences', 'account_sync')),
    decided_at timestamptz NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    retention_expires_at timestamptz,
    CHECK (octet_length(anonymous_subject_hash) = 32 OR anonymous_subject_hash IS NULL),
    CHECK (num_nonnulls(user_id, anonymous_subject_hash, anonymized_subject_id) = 1),
    CHECK (retention_expires_at IS NULL OR retention_expires_at > recorded_at)
);

CREATE INDEX consent_history_user_decided_idx
    ON analytics.consent_history (user_id, decided_at DESC) WHERE user_id IS NOT NULL;
CREATE INDEX consent_history_anonymous_decided_idx
    ON analytics.consent_history (anonymous_subject_hash, decided_at DESC)
    WHERE anonymous_subject_hash IS NOT NULL;
CREATE INDEX consent_history_retention_idx
    ON analytics.consent_history (retention_expires_at, id)
    WHERE retention_expires_at IS NOT NULL;

CREATE TABLE analytics.identity_claims (
    anonymous_subject_hash bytea PRIMARY KEY CHECK (octet_length(anonymous_subject_hash) = 32),
    user_id uuid REFERENCES identity.users(id) ON DELETE RESTRICT,
    status text NOT NULL DEFAULT 'claimed' CHECK (status IN ('claimed', 'revoked')),
    consent_policy_version text NOT NULL
        CHECK (consent_policy_version ~ '^[a-z0-9][a-z0-9._-]{0,49}$'),
    claimed_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CHECK ((status = 'claimed') = (user_id IS NOT NULL)),
    CHECK ((status = 'revoked') = (revoked_at IS NOT NULL))
);

ALTER TABLE analytics.events
    ADD COLUMN public_event_id uuid,
    ADD COLUMN origin text NOT NULL DEFAULT 'server'
        CHECK (origin IN ('client', 'server')),
    ADD COLUMN consent_policy_version text,
    ADD COLUMN anonymous_subject_hash bytea,
    ADD COLUMN classification text NOT NULL DEFAULT 'unknown'
        CHECK (classification IN ('human', 'bot', 'prefetch', 'unknown')),
    ADD COLUMN is_reportable boolean NOT NULL DEFAULT false,
    ADD COLUMN retention_expires_at timestamptz,
    ADD COLUMN anonymized_at timestamptz;

UPDATE analytics.events
SET public_event_id = id,
    retention_expires_at = received_at + interval '397 days';

ALTER TABLE analytics.events
    ALTER COLUMN public_event_id SET NOT NULL,
    ALTER COLUMN public_event_id SET DEFAULT gen_random_uuid(),
    ADD CONSTRAINT events_public_event_id_unique UNIQUE (public_event_id),
    ADD CONSTRAINT events_consent_policy_check CHECK (
        consent_policy_version IS NULL OR
        consent_policy_version ~ '^[a-z0-9][a-z0-9._-]{0,49}$'
    ),
    ADD CONSTRAINT events_anonymous_subject_hash_check CHECK (
        anonymous_subject_hash IS NULL OR octet_length(anonymous_subject_hash) = 32
    ),
    ADD CONSTRAINT events_retention_check CHECK (
        retention_expires_at IS NULL OR retention_expires_at > received_at
    );

CREATE INDEX events_reportable_name_occurred_idx
    ON analytics.events (event_name, occurred_at DESC)
    WHERE is_reportable;
CREATE INDEX events_reportable_product_idx
    ON analytics.events (event_name, (properties->>'product_id'), occurred_at DESC)
    WHERE is_reportable AND properties ? 'product_id';
CREATE INDEX events_retention_idx
    ON analytics.events (retention_expires_at, id)
    WHERE retention_expires_at IS NOT NULL;
CREATE INDEX events_anonymous_subject_idx
    ON analytics.events (anonymous_subject_hash, occurred_at DESC)
    WHERE anonymous_subject_hash IS NOT NULL;

CREATE TABLE analytics.ingestion_receipts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    public_event_id uuid,
    event_name text,
    outcome text NOT NULL CHECK (outcome IN (
        'accepted', 'rejected', 'privacy_filtered', 'bot_filtered', 'deduplicated'
    )),
    reason_code text NOT NULL CHECK (reason_code ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    received_at timestamptz NOT NULL DEFAULT now(),
    retention_expires_at timestamptz NOT NULL,
    CHECK (event_name IS NULL OR event_name ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    CHECK (retention_expires_at > received_at)
);

CREATE INDEX ingestion_receipts_outcome_received_idx
    ON analytics.ingestion_receipts (outcome, received_at DESC);
CREATE INDEX ingestion_receipts_event_idx
    ON analytics.ingestion_receipts (public_event_id, received_at DESC)
    WHERE public_event_id IS NOT NULL;
CREATE INDEX ingestion_receipts_retention_idx
    ON analytics.ingestion_receipts (retention_expires_at, id);

CREATE TABLE analytics.reporting_coverage (
    pipeline_key text PRIMARY KEY CHECK (pipeline_key ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    reportable_from timestamptz NOT NULL,
    complete_through timestamptz NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (complete_through >= reportable_from)
);

INSERT INTO analytics.reporting_coverage (pipeline_key, reportable_from, complete_through)
VALUES ('first_party_events_v3', now(), now());

CREATE TABLE analytics.retention_policies (
    data_class text PRIMARY KEY CHECK (data_class ~ '^[a-z][a-z0-9_.-]{0,99}$'),
    disposition text NOT NULL CHECK (disposition IN ('delete', 'anonymize', 'hold')),
    default_retention interval,
    configurable_key text,
    legal_review_required boolean NOT NULL DEFAULT true,
    description text NOT NULL CHECK (char_length(description) BETWEEN 1 AND 500),
    CHECK ((disposition = 'hold') = (default_retention IS NULL))
);

INSERT INTO analytics.retention_policies
    (data_class, disposition, default_retention, configurable_key, description)
VALUES
    ('raw_ingestion_receipt', 'delete', interval '30 days', 'ANALYTICS_RECEIPT_RETENTION',
        'Payload-free ingestion outcome metadata.'),
    ('anonymous_analytics', 'delete', interval '90 days', 'ANALYTICS_ANONYMOUS_RETENTION',
        'Validated reportable events associated only with an opaque browser subject.'),
    ('authenticated_analytics', 'delete', interval '397 days', 'ANALYTICS_AUTHENTICATED_RETENTION',
        'Validated reportable events associated with an account.'),
    ('affiliate_click', 'anonymize', interval '397 days', 'AFFILIATE_CLICK_RETENTION',
        'Attribution click record governed by the commerce module.'),
    ('conversion', 'hold', NULL, NULL,
        'Verified conversion retention awaits legal, finance, and partner-contract review.'),
    ('security_event', 'hold', NULL, NULL,
        'Immutable security audit retention awaits security and legal review.'),
    ('consent_record', 'hold', NULL, NULL,
        'Consent evidence retention awaits privacy and legal review.'),
    ('administrative_audit', 'hold', NULL, NULL,
        'Administrative audit retention awaits security and legal review.');

COMMENT ON TABLE analytics.ingestion_receipts IS
    'Payload-free receipt ledger. A received request is not a reportable analytics fact.';
COMMENT ON COLUMN analytics.events.is_reportable IS
    'True only after schema, consent, identity, and traffic classification checks pass.';
COMMENT ON TABLE analytics.retention_policies IS
    'Engineering defaults and explicit holds; none are assertions of legal sufficiency.';
