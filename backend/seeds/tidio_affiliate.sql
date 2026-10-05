-- Activate the Tidio partnership using the exact link supplied by the account
-- owner on 2026-10-06. Tidio identifies its affiliate programme as PartnerStack;
-- no account commission or conversion integration has been supplied.
-- The final external landing page has not been independently verified.
--
-- https://www.tidio.com/pricing/ was read on 2026-10-06: Starter is USD 24.17
-- per month billed annually (two months free), at 100 billable conversations.
-- These are already the published catalog facts. This seed changes no catalog
-- facts, billing basis, scores or recommendation policy.
--
-- The read time is fixed in Europe/Sofia's UTC+03 offset for this date.
-- Re-running this file cannot turn the old read into a new one.
-- Existing offer expiry, link validity, priority and commission are preserved.
-- Apply with psql -v ON_ERROR_STOP=1 -f backend/seeds/tidio_affiliate.sql.

BEGIN;

DO $$
DECLARE
    checked_at CONSTANT timestamptz := '2026-10-06 00:15:00+03';
    product_row catalog.products%ROWTYPE;
    brand_row catalog.brands%ROWTYPE;
    read_source_id uuid;
    tidio_merchant_id uuid;
    tidio_offer_id uuid;
    selected_destination text;
    selected_provider text;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'tidio-starter' FOR UPDATE;
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'Tidio affiliate activation: tidio-starter is missing';
    END IF;
    SELECT * INTO brand_row FROM catalog.brands WHERE id = product_row.brand_id;
    IF product_row.status IS DISTINCT FROM 'published'
       OR brand_row.slug IS DISTINCT FROM 'tidio'
       OR NOT brand_row.is_active OR brand_row.country_code IS NULL
       OR product_row.price_minor IS DISTINCT FROM 2417
       OR product_row.currency IS DISTINCT FROM 'USD'
       OR product_row.billing_period IS DISTINCT FROM 'annual'
       OR product_row.pricing_unit IS DISTINCT FROM 'usage'
       OR product_row.pricing_unit_note IS DISTINCT FROM 'at 100 billable conversations/month'
       OR product_row.annual_price_minor IS NOT NULL THEN
        RAISE EXCEPTION 'Tidio affiliate activation: published catalog facts differ from the 2026-10-06 price read';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions facts
        WHERE facts.id = product_row.published_fact_revision_id
          AND facts.product_id = product_row.id AND facts.workflow_status = 'published'
          AND facts.brand_id = brand_row.id AND facts.price_minor = 2417 AND facts.currency = 'USD'
          AND facts.billing_period = 'annual' AND facts.pricing_unit = 'usage'
          AND facts.pricing_unit_note = 'at 100 billable conversations/month'
          AND facts.annual_price_minor IS NULL
    ) THEN
        RAISE EXCEPTION 'Tidio affiliate activation: published fact revision does not match the catalog';
    END IF;

    -- Use the schema's neutral trust default, not an invented merchant rating.
    INSERT INTO commerce.merchants (name, slug, website_url, country_code)
    VALUES ('Tidio', 'tidio', 'https://www.tidio.com', brand_row.country_code)
    ON CONFLICT (slug) DO NOTHING;
    SELECT id INTO tidio_merchant_id FROM commerce.merchants
    WHERE slug = 'tidio' AND status = 'active' AND website_url = 'https://www.tidio.com'
    FOR UPDATE;
    IF tidio_merchant_id IS NULL THEN
        RAISE EXCEPTION 'Tidio affiliate activation: existing merchant is inactive or has an unexpected website';
    END IF;

    IF EXISTS (
        SELECT 1 FROM commerce.merchant_offers
        WHERE merchant_id = tidio_merchant_id AND merchant_sku = 'tidio-starter'
          AND (product_id <> product_row.id OR price_minor <> 2417 OR currency <> 'USD'
               OR shipping_minor <> 0 OR condition <> 'new')
    ) THEN
        RAISE EXCEPTION 'Tidio affiliate activation: existing offer ownership or price differs; review it first';
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('tidio-starter-pricing-2026-10-06', 'manufacturer_documentation',
            'Tidio Starter pricing page', 'Tidio', 'https://www.tidio.com/pricing/',
            false, 'verified', checked_at,
            'Read 2026-10-06: Starter USD 24.17 per month billed annually (two months free), at 100 billable conversations/month. Existing catalog price and billing basis unchanged.')
    ON CONFLICT (external_key) DO NOTHING;
    SELECT id INTO read_source_id FROM evidence.sources
    WHERE external_key = 'tidio-starter-pricing-2026-10-06'
      AND source_type = 'manufacturer_documentation' AND NOT is_fictional
      AND source_url = 'https://www.tidio.com/pricing/' AND review_status = 'verified'
      AND reviewed_at = checked_at;
    IF read_source_id IS NULL THEN
        RAISE EXCEPTION 'Tidio affiliate activation: existing price source conflicts with this read';
    END IF;
    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    SELECT read_source_id, product_row.id, checked_at, checked_at + interval '90 days', 100,
           'Read 2026-10-06 at https://www.tidio.com/pricing/: Starter USD 24.17 per month billed annually (two months free); 100 billable conversations/month.'
    WHERE NOT EXISTS (
        SELECT 1 FROM evidence.observations
        WHERE source_id = read_source_id AND product_id = product_row.id
    );

    INSERT INTO commerce.merchant_offers (
        merchant_id, product_id, merchant_sku, product_url, price_minor,
        shipping_minor, currency, availability, condition, last_checked_at)
    VALUES (tidio_merchant_id, product_row.id, 'tidio-starter', 'https://www.tidio.com/pricing/',
            2417, 0, 'USD', 'in_stock', 'new', checked_at)
    ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
        product_url = EXCLUDED.product_url,
        is_active = true,
        last_checked_at = GREATEST(commerce.merchant_offers.last_checked_at, EXCLUDED.last_checked_at),
        updated_at = now()
    WHERE commerce.merchant_offers.product_url IS DISTINCT FROM EXCLUDED.product_url
       OR NOT commerce.merchant_offers.is_active
       OR commerce.merchant_offers.last_checked_at < EXCLUDED.last_checked_at;
    SELECT id INTO tidio_offer_id FROM commerce.merchant_offers
    WHERE merchant_id = tidio_merchant_id AND merchant_sku = 'tidio-starter';

    INSERT INTO commerce.affiliate_links (
        merchant_offer_id, provider, destination_url, external_reference, disclosure_label, is_active)
    VALUES (tidio_offer_id, 'partnerstack', 'https://affiliate.tidio.com/s92aresrdhui',
            's92aresrdhui', 'Affiliate link', true)
    ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
        destination_url = EXCLUDED.destination_url,
        external_reference = EXCLUDED.external_reference,
        disclosure_label = EXCLUDED.disclosure_label,
        is_active = true,
        updated_at = now()
    WHERE commerce.affiliate_links.destination_url IS DISTINCT FROM EXCLUDED.destination_url
       OR commerce.affiliate_links.external_reference IS DISTINCT FROM EXCLUDED.external_reference
       OR commerce.affiliate_links.disclosure_label IS DISTINCT FROM EXCLUDED.disclosure_label
       OR NOT commerce.affiliate_links.is_active;

    -- Match the redirect's selection order, including validity and expiry.
    -- The 30-day bound is the maximum permitted OFFER_MAXIMUM_AGE; deployment
    -- must also check its configured freshness window through the public API.
    SELECT links.destination_url, links.provider INTO selected_destination, selected_provider
    FROM commerce.affiliate_links links
    JOIN commerce.merchant_offers offers ON offers.id = links.merchant_offer_id
    JOIN commerce.merchants merchants ON merchants.id = offers.merchant_id
    WHERE offers.id = tidio_offer_id AND offers.product_id = product_row.id
      AND offers.price_minor = 2417 AND offers.currency = 'USD'
      AND offers.is_active AND offers.availability IN ('in_stock', 'backorder')
      AND offers.last_checked_at >= now() - interval '30 days'
      AND (offers.expires_at IS NULL OR offers.expires_at > now())
      AND merchants.status = 'active' AND links.is_active
      AND (links.valid_from IS NULL OR links.valid_from <= now())
      AND (links.valid_until IS NULL OR links.valid_until > now())
    ORDER BY links.priority DESC, links.provider
    LIMIT 1;
    IF selected_destination IS DISTINCT FROM 'https://affiliate.tidio.com/s92aresrdhui'
       OR selected_provider IS DISTINCT FROM 'partnerstack' THEN
        RAISE EXCEPTION 'Tidio affiliate activation: exact supplied link is not the usable offer destination';
    END IF;
END $$;

COMMIT;
