-- Affiliate promotions are editorial acquisition offers that are not merchant
-- offers for a catalog product. Keeping them separate prevents a webinar,
-- book, or training bundle from entering software recommendation inputs merely
-- because it has an affiliate destination.
CREATE TABLE commerce.affiliate_promotions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id uuid NOT NULL REFERENCES commerce.merchants(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 2 AND 180),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    promotion_type text NOT NULL CHECK (promotion_type IN ('lead', 'purchase')),
    public_url text NOT NULL CHECK (public_url ~ '^https://'),
    destination_url text NOT NULL CHECK (destination_url ~ '^https://'),
    provider text NOT NULL CHECK (provider ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    external_reference text CHECK (
        external_reference IS NULL OR char_length(btrim(external_reference)) BETWEEN 1 AND 160
    ),
    disclosure_label text NOT NULL DEFAULT 'Affiliate link'
        CHECK (char_length(btrim(disclosure_label)) BETWEEN 1 AND 120),
    is_active boolean NOT NULL DEFAULT true,
    last_checked_at timestamptz NOT NULL,
    valid_from timestamptz,
    valid_until timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (valid_until IS NULL OR valid_from IS NULL OR valid_until > valid_from)
);

CREATE INDEX affiliate_promotions_active_slug_idx
    ON commerce.affiliate_promotions (slug)
    WHERE is_active;

CREATE TABLE commerce.affiliate_promotion_clicks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    promotion_id uuid NOT NULL
        REFERENCES commerce.affiliate_promotions(id) ON DELETE RESTRICT,
    user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    anonymous_id text,
    session_id text,
    source text NOT NULL CHECK (source ~ '^[a-z][a-z0-9_.-]*$'),
    campaign text CHECK (
        campaign IS NULL OR campaign ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    referrer text CHECK (referrer IS NULL OR char_length(referrer) <= 500),
    request_id text CHECK (
        request_id IS NULL OR char_length(request_id) BETWEEN 1 AND 128
    ),
    traffic_source text CHECK (
        traffic_source IS NULL OR traffic_source ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    traffic_medium text CHECK (
        traffic_medium IS NULL OR traffic_medium ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    referrer_host text CHECK (
        referrer_host IS NULL OR (
            char_length(referrer_host) BETWEEN 1 AND 253
            AND referrer_host = lower(referrer_host)
        )
    ),
    classification text NOT NULL DEFAULT 'unknown'
        CHECK (classification IN ('human', 'bot', 'prefetch', 'unknown')),
    is_countable boolean NOT NULL DEFAULT false,
    user_agent_hash character(64) CHECK (
        user_agent_hash IS NULL OR user_agent_hash ~ '^[a-f0-9]{64}$'
    ),
    idempotency_key text CHECK (
        idempotency_key IS NULL OR char_length(idempotency_key) BETWEEN 8 AND 128
    ),
    retention_expires_at timestamptz NOT NULL DEFAULT (now() + interval '397 days'),
    anonymized_at timestamptz,
    occurred_at timestamptz NOT NULL DEFAULT now(),
    CHECK (
        anonymized_at IS NOT NULL OR user_id IS NOT NULL OR anonymous_id IS NOT NULL
    ),
    CHECK (anonymous_id IS NULL OR char_length(anonymous_id) BETWEEN 1 AND 128),
    CHECK (session_id IS NULL OR char_length(session_id) BETWEEN 1 AND 128)
);

CREATE UNIQUE INDEX affiliate_promotion_clicks_idempotency_idx
    ON commerce.affiliate_promotion_clicks (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX affiliate_promotion_clicks_promotion_occurred_idx
    ON commerce.affiliate_promotion_clicks (promotion_id, occurred_at DESC);
CREATE INDEX affiliate_promotion_clicks_filtered_occurred_idx
    ON commerce.affiliate_promotion_clicks (occurred_at DESC)
    WHERE is_countable;
CREATE INDEX affiliate_promotion_clicks_retention_idx
    ON commerce.affiliate_promotion_clicks (retention_expires_at, id)
    WHERE anonymized_at IS NULL;

COMMENT ON TABLE commerce.affiliate_promotions IS
    'Editorial affiliate destinations that do not represent catalog product offers and cannot enter recommendation scoring.';
COMMENT ON TABLE commerce.affiliate_promotion_clicks IS
    'Privacy-bounded first-party click records for standalone affiliate promotions.';
