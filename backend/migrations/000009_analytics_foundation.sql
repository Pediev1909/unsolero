-- Canonicalize the first-party event vocabulary before adding reporting indexes.
UPDATE analytics.events
SET event_name = CASE event_name
    WHEN 'product_view' THEN 'product_viewed'
    WHEN 'product_save' THEN CASE
        WHEN properties->>'action' = 'removed' THEN 'product_removed'
        ELSE 'product_saved'
    END
    WHEN 'affiliate_click' THEN 'affiliate_clicked'
    ELSE event_name
END
WHERE event_name IN ('product_view', 'product_save', 'affiliate_click');

ALTER TABLE analytics.events
    ADD COLUMN page_path text,
    ADD COLUMN traffic_source text,
    ADD COLUMN traffic_medium text,
    ADD COLUMN campaign text,
    ADD COLUMN referrer_host text;

ALTER TABLE analytics.events
    ADD CONSTRAINT events_page_path_check CHECK (
        page_path IS NULL OR (
            char_length(page_path) BETWEEN 1 AND 500
            AND left(page_path, 1) = '/'
            AND page_path !~ '[?#]'
        )
    ),
    ADD CONSTRAINT events_traffic_source_check CHECK (
        traffic_source IS NULL OR traffic_source ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    ADD CONSTRAINT events_traffic_medium_check CHECK (
        traffic_medium IS NULL OR traffic_medium ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    ADD CONSTRAINT events_campaign_check CHECK (
        campaign IS NULL OR campaign ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    ADD CONSTRAINT events_referrer_host_check CHECK (
        referrer_host IS NULL OR (
            char_length(referrer_host) BETWEEN 1 AND 253
            AND referrer_host = lower(referrer_host)
        )
    );

CREATE INDEX events_product_activity_idx
    ON analytics.events (event_name, (properties->>'product_id'), occurred_at DESC)
    WHERE properties ? 'product_id';
CREATE INDEX events_traffic_source_occurred_idx
    ON analytics.events (traffic_source, occurred_at DESC)
    WHERE traffic_source IS NOT NULL;
CREATE INDEX events_onboarding_attempt_idx
    ON analytics.events (event_name, (properties->>'onboarding_id'), occurred_at DESC)
    WHERE event_name IN ('onboarding_started', 'onboarding_completed');

ALTER TABLE commerce.affiliate_clicks
    ADD COLUMN traffic_source text,
    ADD COLUMN traffic_medium text,
    ADD COLUMN referrer_host text,
    ADD COLUMN recommendation_id uuid
        REFERENCES recommendation.recommendations(id) ON DELETE SET NULL;

ALTER TABLE commerce.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_traffic_source_check CHECK (
        traffic_source IS NULL OR traffic_source ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    ADD CONSTRAINT affiliate_clicks_traffic_medium_check CHECK (
        traffic_medium IS NULL OR traffic_medium ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    ADD CONSTRAINT affiliate_clicks_referrer_host_check CHECK (
        referrer_host IS NULL OR (
            char_length(referrer_host) BETWEEN 1 AND 253
            AND referrer_host = lower(referrer_host)
        )
    );

-- This table intentionally has no ingestion endpoint yet. It gives future
-- provider webhooks a normalized target without representing unverified sales.
CREATE TABLE commerce.affiliate_conversions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    affiliate_click_id uuid NOT NULL
        REFERENCES commerce.affiliate_clicks(id) ON DELETE RESTRICT,
    provider text NOT NULL CHECK (provider ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    external_conversion_id text NOT NULL
        CHECK (char_length(btrim(external_conversion_id)) BETWEEN 1 AND 200),
    order_reference text CHECK (
        order_reference IS NULL OR char_length(btrim(order_reference)) BETWEEN 1 AND 200
    ),
    status text NOT NULL CHECK (status IN ('pending', 'confirmed', 'reversed')),
    order_value_minor bigint CHECK (order_value_minor IS NULL OR order_value_minor >= 0),
    order_currency character(3) CHECK (
        order_currency IS NULL OR order_currency ~ '^[A-Z]{3}$'
    ),
    commission_amount_minor bigint CHECK (
        commission_amount_minor IS NULL OR commission_amount_minor >= 0
    ),
    commission_currency character(3) CHECK (
        commission_currency IS NULL OR commission_currency ~ '^[A-Z]{3}$'
    ),
    converted_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, external_conversion_id),
    CHECK ((order_value_minor IS NULL) = (order_currency IS NULL)),
    CHECK ((commission_amount_minor IS NULL) = (commission_currency IS NULL)),
    CHECK (confirmed_at IS NULL OR confirmed_at >= converted_at)
);

CREATE INDEX affiliate_conversions_click_status_idx
    ON commerce.affiliate_conversions (affiliate_click_id, status, converted_at DESC);
CREATE INDEX affiliate_conversions_status_converted_idx
    ON commerce.affiliate_conversions (status, converted_at DESC);

CREATE TABLE analytics.acquisition_costs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    provider text NOT NULL CHECK (provider ~ '^[a-z0-9][a-z0-9._-]{0,99}$'),
    external_cost_id text NOT NULL
        CHECK (char_length(btrim(external_cost_id)) BETWEEN 1 AND 200),
    traffic_source text NOT NULL
        CHECK (traffic_source ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'),
    traffic_medium text CHECK (
        traffic_medium IS NULL OR traffic_medium ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    campaign text CHECK (
        campaign IS NULL OR campaign ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'
    ),
    spend_amount_minor bigint NOT NULL CHECK (spend_amount_minor >= 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    period_start date NOT NULL,
    period_end date NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (provider, external_cost_id),
    CHECK (period_end >= period_start)
);

CREATE INDEX acquisition_costs_period_source_idx
    ON analytics.acquisition_costs (period_start, period_end, traffic_source);

COMMENT ON TABLE commerce.affiliate_conversions IS
    'Provider-reported conversions only. Empty until a verified provider webhook/import is implemented.';
COMMENT ON COLUMN commerce.affiliate_conversions.commission_amount_minor IS
    'Verified affiliate revenue when supplied by a provider; never a recommendation scoring input.';
COMMENT ON TABLE analytics.acquisition_costs IS
    'Verified provider spend only. Empty until an authenticated cost import is implemented.';
