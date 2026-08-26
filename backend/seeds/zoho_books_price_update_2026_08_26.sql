-- Zoho Books Standard price correction and affiliate reactivation.
--
-- Zoho's official US comparison page was checked on 2026-08-26. Standard is
-- $20 per organization per month with monthly billing, or $15/month billed
-- annually. The SaaS catalog compares monthly-billing list prices, so $20 is
-- the correct comparable value.
--
-- This creates immutable fact and score revision 2. The value score is
-- explicitly reviewed and retained: the engine already scores the verified
-- price separately through budget fit, while the value assessment represents
-- the plan's accounting capability relative to the category. No affiliate or
-- commission data is used by either revision.

BEGIN;

INSERT INTO evidence.sources (
    external_key, source_type, title, publisher, source_url,
    is_fictional, review_status, reviewed_at, review_note
) VALUES (
    'zoho-books-us-pricing-2026-08-26',
    'manufacturer_documentation',
    'Zoho Books US plan comparison',
    'Zoho',
    'https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard',
    false,
    'verified',
    now(),
    'Checked 2026-08-26. Standard is $20 per organization per month on monthly billing and $15/month billed annually.'
), (
    'unsolero-zoho-books-value-review-2026-08-26',
    'editorial_assessment',
    'UNSOLERO Zoho Books value review',
    'Andon Pediev',
    'https://unsolero.com/articles/how-unsolero-ranks-software',
    false,
    'verified',
    now(),
    'Reviewed after the verified monthly price changed from $12 to $20. Affiliate availability and commission were excluded.'
)
ON CONFLICT (external_key) DO UPDATE SET
    source_url = EXCLUDED.source_url,
    review_status = 'verified',
    reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
    review_note = EXCLUDED.review_note,
    updated_at = now();

INSERT INTO evidence.observations (
    source_id, product_id, observed_at, expires_at, confidence, notes
)
SELECT sources.id, products.id, now(), now() + interval '90 days', 100,
       'Zoho Books Standard observed at $20 per organization per month with monthly billing; $15/month is available only when billed annually.'
FROM evidence.sources sources
JOIN catalog.products products ON products.slug = 'zoho-books-standard'
WHERE sources.external_key = 'zoho-books-us-pricing-2026-08-26'
  AND NOT EXISTS (
      SELECT 1
      FROM evidence.observations existing
      WHERE existing.source_id = sources.id
        AND existing.product_id = products.id
  );

INSERT INTO evidence.observations (
    source_id, product_id, observed_at, confidence, notes
)
SELECT sources.id, products.id, now(), 85,
       'Value score reviewed independently from affiliate data after the verified price correction. Direct cost remains represented separately by price and budget fit.'
FROM evidence.sources sources
JOIN catalog.products products ON products.slug = 'zoho-books-standard'
WHERE sources.external_key = 'unsolero-zoho-books-value-review-2026-08-26'
  AND NOT EXISTS (
      SELECT 1
      FROM evidence.observations existing
      WHERE existing.source_id = sources.id
        AND existing.product_id = products.id
  );

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, valid_until, review_note
)
SELECT products.id, 2, products.category_id, products.brand_id, products.name,
       products.slug, products.description, 2000, 'USD',
       products.warranty_months, 'published', now(), now(), now(),
       now() + interval '90 days',
       'Monthly-billing list price corrected to $20 from Zoho official US pricing observed 2026-08-26. The $15 figure requires annual billing.'
FROM catalog.products products
WHERE products.slug = 'zoho-books-standard'
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score,
    durability_score, beginner_score, advanced_score, apartment_score,
    noise_score, portability_score, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT products.id, facts.id, 2, products.quality_score, products.value_score,
       products.durability_score, products.beginner_score,
       products.advanced_score, products.apartment_score,
       products.noise_score, products.portability_score,
       'published', now(), now(), now(),
       'Scores reviewed after the price correction. Value remains 92 because price is scored separately by budget fit and the plan remains the strongest paid accounting value in this catalog; no affiliate data was considered.'
FROM catalog.products products
JOIN evidence.product_fact_revisions facts
  ON facts.product_id = products.id AND facts.version = 2
WHERE products.slug = 'zoho-books-standard'
ON CONFLICT (product_id, version) DO NOTHING;

-- Carry unchanged fact provenance forward, but never attach the obsolete
-- price observation to revision 2.
INSERT INTO evidence.fact_provenance (
    fact_revision_id, fact_key, observation_id, public_classification
)
SELECT facts_v2.id, provenance.fact_key, provenance.observation_id,
       provenance.public_classification
FROM catalog.products products
JOIN evidence.product_fact_revisions facts_v1
  ON facts_v1.product_id = products.id AND facts_v1.version = 1
JOIN evidence.product_fact_revisions facts_v2
  ON facts_v2.product_id = products.id AND facts_v2.version = 2
JOIN evidence.fact_provenance provenance
  ON provenance.fact_revision_id = facts_v1.id
WHERE products.slug = 'zoho-books-standard'
  AND provenance.fact_key <> 'price'
ON CONFLICT DO NOTHING;

INSERT INTO evidence.fact_provenance (
    fact_revision_id, fact_key, observation_id, public_classification
)
SELECT facts.id, 'price', observations.id, 'manufacturer_claim'
FROM catalog.products products
JOIN evidence.product_fact_revisions facts
  ON facts.product_id = products.id AND facts.version = 2
JOIN evidence.observations observations
  ON observations.product_id = products.id
JOIN evidence.sources sources ON sources.id = observations.source_id
WHERE products.slug = 'zoho-books-standard'
  AND sources.external_key = 'zoho-books-us-pricing-2026-08-26'
ON CONFLICT DO NOTHING;

-- Carry non-price score rationales forward. Value receives its own dated
-- editorial review so a manufacturer price claim never becomes a score claim.
INSERT INTO evidence.score_rationales (
    score_revision_id, score_key, rationale, observation_id
)
SELECT scores_v2.id, rationales.score_key, rationales.rationale,
       rationales.observation_id
FROM catalog.products products
JOIN evidence.score_revisions scores_v1
  ON scores_v1.product_id = products.id AND scores_v1.version = 1
JOIN evidence.score_revisions scores_v2
  ON scores_v2.product_id = products.id AND scores_v2.version = 2
JOIN evidence.score_rationales rationales
  ON rationales.score_revision_id = scores_v1.id
WHERE products.slug = 'zoho-books-standard'
  AND rationales.score_key <> 'value'
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (
    score_revision_id, score_key, rationale, observation_id
)
SELECT scores.id, 'value',
       'Reviewed at the corrected $20 monthly-billing price. The plan retains broad accounting, banking and invoicing capability below FreshBooks Lite in monthly cost; direct affordability is independently scored through budget fit.',
       observations.id
FROM catalog.products products
JOIN evidence.score_revisions scores
  ON scores.product_id = products.id AND scores.version = 2
JOIN evidence.observations observations
  ON observations.product_id = products.id
JOIN evidence.sources sources ON sources.id = observations.source_id
WHERE products.slug = 'zoho-books-standard'
  AND sources.external_key = 'unsolero-zoho-books-value-review-2026-08-26'
ON CONFLICT DO NOTHING;

UPDATE catalog.products products
SET price_minor = 2000,
    currency = 'USD',
    published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    updated_at = now()
FROM evidence.product_fact_revisions facts
JOIN evidence.score_revisions scores
  ON scores.product_id = facts.product_id AND scores.version = 2
WHERE products.id = facts.product_id
  AND products.slug = 'zoho-books-standard'
  AND facts.version = 2
  AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published';

UPDATE commerce.merchant_offers offers
SET product_url = 'https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard',
    price_minor = 2000,
    shipping_minor = 0,
    currency = 'USD',
    availability = 'in_stock',
    condition = 'new',
    last_checked_at = now(),
    is_active = true,
    updated_at = now()
FROM catalog.products products, commerce.merchants merchants
WHERE products.id = offers.product_id
  AND merchants.id = offers.merchant_id
  AND products.slug = 'zoho-books-standard'
  AND merchants.slug = 'zoho'
  AND offers.merchant_sku = 'zoho-books-standard';

UPDATE commerce.affiliate_links links
SET is_active = true,
    updated_at = now()
FROM commerce.merchant_offers offers
JOIN catalog.products products ON products.id = offers.product_id
WHERE offers.id = links.merchant_offer_id
  AND products.slug = 'zoho-books-standard'
  AND links.provider = 'zoho'
  AND links.destination_url = 'https://go.zoho.com/K0nf';

-- Correct the two already-published editorial sentences that quote the old
-- price. The content model has no revision table, so these are narrow exact
-- replacements rather than a broad rewrite.
UPDATE editorial.entries
SET content = replace(
        content::text,
        'Zoho Books Standard at 12 USD for invoicing',
        'Zoho Books Standard at 20 USD for invoicing'
    )::jsonb,
    updated_at = now()
WHERE slug = 'best-crm-small-agency'
  AND content::text LIKE '%Zoho Books Standard at 12 USD for invoicing%';

UPDATE editorial.entries
SET content = replace(
        content::text,
        'Zoho Books picks up where it leaves off at 12 USD a month',
        'Zoho Books picks up where it leaves off at 20 USD a month'
    )::jsonb,
    updated_at = now()
WHERE slug = 'zoho-invoice-vs-wave'
  AND content::text LIKE '%Zoho Books picks up where it leaves off at 12 USD a month%';

DO $$
DECLARE
    exact_count integer;
BEGIN
    SELECT count(*) INTO exact_count
    FROM catalog.products products
    JOIN evidence.product_fact_revisions facts
      ON facts.id = products.published_fact_revision_id
    JOIN evidence.score_revisions scores
      ON scores.id = products.published_score_revision_id
    JOIN commerce.merchant_offers offers ON offers.product_id = products.id
    JOIN commerce.affiliate_links links
      ON links.merchant_offer_id = offers.id
    WHERE products.slug = 'zoho-books-standard'
      AND products.price_minor = 2000
      AND products.currency = 'USD'
      AND facts.version = 2
      AND facts.price_minor = 2000
      AND facts.workflow_status = 'published'
      AND scores.version = 2
      AND scores.value_score = 92
      AND scores.workflow_status = 'published'
      AND offers.price_minor = 2000
      AND offers.is_active
      AND links.provider = 'zoho'
      AND links.destination_url = 'https://go.zoho.com/K0nf'
      AND links.is_active;

    IF exact_count <> 1 THEN
        RAISE EXCEPTION
            'Zoho Books correction failed: expected one exact active product/fact/score/offer/link chain, found %',
            exact_count;
    END IF;
END
$$;

COMMIT;
