ALTER TABLE commerce.merchant_offers
    ADD COLUMN is_active boolean NOT NULL DEFAULT true;

CREATE INDEX merchant_offers_active_product_idx
    ON commerce.merchant_offers (product_id, currency, price_minor)
    WHERE is_active;

ALTER TABLE commerce.affiliate_links
    ADD COLUMN priority smallint NOT NULL DEFAULT 0,
    ADD COLUMN program_id text,
    ADD COLUMN commission_type text NOT NULL DEFAULT 'unknown',
    ADD COLUMN commission_rate_bps integer,
    ADD COLUMN commission_amount_minor bigint,
    ADD COLUMN commission_currency character(3);

ALTER TABLE commerce.affiliate_links
    ADD CONSTRAINT affiliate_links_priority_check
        CHECK (priority BETWEEN -1000 AND 1000),
    ADD CONSTRAINT affiliate_links_program_id_check
        CHECK (program_id IS NULL OR char_length(btrim(program_id)) BETWEEN 1 AND 160),
    ADD CONSTRAINT affiliate_links_commission_type_check
        CHECK (commission_type IN ('unknown', 'percentage', 'fixed')),
    ADD CONSTRAINT affiliate_links_commission_rate_check
        CHECK (commission_rate_bps IS NULL OR commission_rate_bps BETWEEN 0 AND 10000),
    ADD CONSTRAINT affiliate_links_commission_amount_check
        CHECK (commission_amount_minor IS NULL OR commission_amount_minor >= 0),
    ADD CONSTRAINT affiliate_links_commission_currency_check
        CHECK (commission_currency IS NULL OR commission_currency ~ '^[A-Z]{3}$'),
    ADD CONSTRAINT affiliate_links_commission_shape_check
        CHECK (
            (commission_type = 'unknown' AND commission_rate_bps IS NULL
                AND commission_amount_minor IS NULL AND commission_currency IS NULL)
            OR (commission_type = 'percentage' AND commission_rate_bps IS NOT NULL
                AND commission_amount_minor IS NULL AND commission_currency IS NULL)
            OR (commission_type = 'fixed' AND commission_rate_bps IS NULL
                AND commission_amount_minor IS NOT NULL AND commission_currency IS NOT NULL)
        );

COMMENT ON COLUMN commerce.affiliate_links.commission_rate_bps IS
    'Internal monetization metadata. It must never be consumed by recommendation scoring.';
COMMENT ON COLUMN commerce.affiliate_links.commission_amount_minor IS
    'Internal monetization metadata. It must never be consumed by recommendation scoring.';

ALTER TABLE commerce.outbound_clicks RENAME TO affiliate_clicks;
ALTER TABLE commerce.affiliate_clicks RENAME COLUMN surface TO source;
ALTER INDEX commerce.outbound_clicks_affiliate_occurred_idx
    RENAME TO affiliate_clicks_link_occurred_idx;
ALTER INDEX commerce.outbound_clicks_product_occurred_idx
    RENAME TO affiliate_clicks_product_occurred_idx;

ALTER TABLE commerce.affiliate_clicks
    ADD COLUMN user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    ADD COLUMN anonymous_id text,
    ADD COLUMN session_id text,
    ADD COLUMN campaign text,
    ADD COLUMN referrer text,
    ADD COLUMN request_id text;

UPDATE commerce.affiliate_clicks
SET anonymous_id = 'legacy-' || id::text
WHERE user_id IS NULL AND anonymous_id IS NULL;

ALTER TABLE commerce.affiliate_clicks
    ADD CONSTRAINT affiliate_clicks_anonymous_id_check
        CHECK (anonymous_id IS NULL OR char_length(anonymous_id) BETWEEN 1 AND 128),
    ADD CONSTRAINT affiliate_clicks_session_id_check
        CHECK (session_id IS NULL OR char_length(session_id) BETWEEN 1 AND 128),
    ADD CONSTRAINT affiliate_clicks_campaign_check
        CHECK (campaign IS NULL OR campaign ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$'),
    ADD CONSTRAINT affiliate_clicks_referrer_check
        CHECK (referrer IS NULL OR char_length(referrer) <= 500),
    ADD CONSTRAINT affiliate_clicks_request_id_check
        CHECK (request_id IS NULL OR char_length(request_id) BETWEEN 1 AND 128),
    ADD CONSTRAINT affiliate_clicks_actor_check
        CHECK (user_id IS NOT NULL OR anonymous_id IS NOT NULL);

CREATE INDEX affiliate_clicks_offer_occurred_idx
    ON commerce.affiliate_clicks (merchant_offer_id, occurred_at DESC);
CREATE INDEX affiliate_clicks_user_occurred_idx
    ON commerce.affiliate_clicks (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;
CREATE INDEX affiliate_clicks_anonymous_occurred_idx
    ON commerce.affiliate_clicks (anonymous_id, occurred_at DESC)
    WHERE anonymous_id IS NOT NULL;
