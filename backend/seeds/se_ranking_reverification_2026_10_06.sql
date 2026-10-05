-- Re-verify the existing SE Ranking Core offer after its freshness window
-- elapsed. On 2026-10-06 the official https://seranking.com/subscription.html
-- page's USD web rendering showed Core at $129 monthly and $103.20 per month
-- billed annually (20% saving), for 10 projects and one manager seat.
-- A separate local HTTP read rendered EUR; no currency conversion is used.
--
-- Only the existing matching offer's last_checked_at is refreshed. Prices,
-- billing facts, recommendation scores, affiliate URLs and all IDs stay put.
-- The fixed read time prevents reruns from claiming a new price observation.
-- Apply with psql -v ON_ERROR_STOP=1 -f backend/seeds/se_ranking_reverification_2026_10_06.sql.

BEGIN;

DO $$
DECLARE
    checked_at CONSTANT timestamptz := '2026-10-06 00:25:00+03';
    product_row catalog.products%ROWTYPE;
    verified_offer_id uuid;
    read_source_id uuid;
    selected_destination text;
    selected_provider text;
    selected_reference text;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'se-ranking-core' FOR UPDATE;
    IF product_row.id IS NULL OR product_row.status IS DISTINCT FROM 'published'
       OR product_row.price_minor IS DISTINCT FROM 12900
       OR product_row.currency IS DISTINCT FROM 'USD'
       OR product_row.billing_period IS DISTINCT FROM 'monthly'
       OR product_row.pricing_unit IS DISTINCT FROM 'flat'
       OR product_row.pricing_unit_note IS DISTINCT FROM 'per account; 10 projects'
       OR product_row.annual_price_minor IS DISTINCT FROM 10320
       OR NOT EXISTS (
           SELECT 1 FROM catalog.brands WHERE id = product_row.brand_id AND slug = 'se-ranking' AND is_active
       ) THEN
        RAISE EXCEPTION 'SE Ranking re-verification: published catalog differs from the 2026-10-06 price read';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions facts
        WHERE facts.id = product_row.published_fact_revision_id
          AND facts.product_id = product_row.id AND facts.workflow_status = 'published'
          AND facts.brand_id = product_row.brand_id AND facts.price_minor = 12900 AND facts.currency = 'USD'
          AND facts.billing_period = 'monthly' AND facts.pricing_unit = 'flat'
          AND facts.pricing_unit_note = 'per account; 10 projects' AND facts.annual_price_minor = 10320
    ) THEN
        RAISE EXCEPTION 'SE Ranking re-verification: published fact revision does not match the catalog';
    END IF;

    SELECT offers.id INTO verified_offer_id
    FROM commerce.merchant_offers offers
    JOIN commerce.merchants merchants ON merchants.id = offers.merchant_id
    WHERE merchants.slug = 'se-ranking' AND merchants.status = 'active'
      AND offers.product_id = product_row.id AND offers.merchant_sku = 'se-ranking-core'
      AND offers.product_url = 'https://seranking.com/subscription.html'
      AND offers.price_minor = 12900 AND offers.currency = 'USD' AND offers.shipping_minor = 0
      AND offers.condition = 'new' AND offers.is_active AND offers.availability IN ('in_stock', 'backorder')
      AND (offers.expires_at IS NULL OR offers.expires_at > now())
    FOR UPDATE OF offers, merchants;
    IF verified_offer_id IS NULL THEN
        RAISE EXCEPTION 'SE Ranking re-verification: matching active unexpired offer is missing';
    END IF;
    SELECT destination_url, provider, external_reference
    INTO selected_destination, selected_provider, selected_reference
    FROM commerce.affiliate_links
    WHERE merchant_offer_id = verified_offer_id AND is_active
      AND (valid_from IS NULL OR valid_from <= now())
      AND (valid_until IS NULL OR valid_until > now())
    ORDER BY priority DESC, provider LIMIT 1 FOR UPDATE;
    IF selected_destination IS DISTINCT FROM 'https://seranking.com/subscription.html?ga=5233991&source=link'
       OR selected_provider IS DISTINCT FROM 'se_ranking'
       OR selected_reference IS DISTINCT FROM '5233991' THEN
        RAISE EXCEPTION 'SE Ranking re-verification: existing selected affiliate destination differs or is unusable';
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('se-ranking-core-pricing-2026-10-06', 'manufacturer_documentation',
            'SE Ranking Core pricing page', 'SE Ranking', 'https://seranking.com/subscription.html',
            false, 'verified', checked_at,
            'Read 2026-10-06 from the official page USD web rendering: Core USD 129 monthly, USD 103.20 per month billed annually (20% saving), 10 projects and one manager seat. Local HTTP rendering was EUR; no FX conversion. Existing catalog and offer prices match.')
    ON CONFLICT (external_key) DO NOTHING;
    SELECT id INTO read_source_id FROM evidence.sources
    WHERE external_key = 'se-ranking-core-pricing-2026-10-06'
      AND source_type = 'manufacturer_documentation' AND NOT is_fictional
      AND source_url = 'https://seranking.com/subscription.html' AND review_status = 'verified'
      AND reviewed_at = checked_at;
    IF read_source_id IS NULL THEN
        RAISE EXCEPTION 'SE Ranking re-verification: existing evidence source conflicts with this read';
    END IF;
    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    SELECT read_source_id, product_row.id, checked_at, checked_at + interval '90 days', 100,
           'Read 2026-10-06 at https://seranking.com/subscription.html: Core USD 129 monthly, USD 103.20 per month billed annually (20% saving), 10 projects and one manager seat; USD official-page web rendering.'
    WHERE NOT EXISTS (
        SELECT 1 FROM evidence.observations
        WHERE source_id = read_source_id AND product_id = product_row.id
    );

    UPDATE commerce.merchant_offers
    SET last_checked_at = checked_at, updated_at = now()
    WHERE id = verified_offer_id AND last_checked_at < checked_at;
    IF NOT EXISTS (
        SELECT 1 FROM commerce.merchant_offers
        WHERE id = verified_offer_id AND last_checked_at >= checked_at
          AND last_checked_at >= now() - interval '30 days'
    ) THEN
        RAISE EXCEPTION 'SE Ranking re-verification: offer is stale; a new price read is required';
    END IF;
END $$;

COMMIT;
