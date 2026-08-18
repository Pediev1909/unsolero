\set ON_ERROR_STOP on
\timing on

-- Rollback-only synthetic scale fixture. This script deliberately refuses to
-- run against a database whose name does not identify it as disposable.
DO $phase8$
BEGIN
    IF current_database() !~ '(phase8|scale|validation)' THEN
        RAISE EXCEPTION 'phase8 scale validation requires a disposable database name';
    END IF;
END
$phase8$;

BEGIN;
SET LOCAL synchronous_commit = off;
SET LOCAL statement_timeout = '10min';

\echo 'Creating 10,000 explicitly synthetic accounts and 20,000 sessions'
INSERT INTO identity.users (email, password_hash, status, email_verified_at, created_at, updated_at)
SELECT format('phase8-scale-user-%s@example.invalid', value),
       '$argon2id$v=19$m=65536,t=3,p=2$cGhhc2U4LXNjYWxlLXNhbHQ$cGhhc2U4LXNjYWxlLWhhc2g',
       'active', now(), now() - (value % 90) * interval '1 day', now()
FROM generate_series(1, 10000) AS value;

INSERT INTO identity.sessions (
    user_id, token_hash, expires_at, idle_expires_at, last_seen_at, created_at,
    authentication_method
)
SELECT users.id,
       decode(md5(users.email || ':' || session_number::text) ||
              md5('phase8:' || users.email || ':' || session_number::text), 'hex'),
       now() + interval '30 days', now() + interval '7 days', now(), now(), 'password'
FROM identity.users AS users
CROSS JOIN generate_series(1, 2) AS session_number
WHERE users.email LIKE 'phase8-scale-user-%@example.invalid';

\echo 'Creating 5,000 fictional governed products with provenance'
WITH fixture_refs AS (
    SELECT (SELECT id FROM catalog.categories WHERE is_active ORDER BY sort_order, id LIMIT 1) AS category_id,
           (SELECT id FROM catalog.brands WHERE is_active ORDER BY id LIMIT 1) AS brand_id
)
INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    length_mm, width_mm, height_mm, weight_grams, max_capacity_grams, material,
    warranty_months, quality_score, value_score, durability_score, beginner_score,
    advanced_score, apartment_score, noise_score, portability_score, status
)
SELECT fixture_refs.category_id, fixture_refs.brand_id,
       format('Phase 8 Synthetic Product %s', value),
       format('phase8-synthetic-product-%s', value),
       'Explicitly fictional rollback-only performance fixture.',
       1000 + (value % 1000) * 25, 'USD', 300 + value % 900, 250 + value % 700,
       100 + value % 1200, 1000 + value % 30000, 10000 + value % 200000,
       'Fictional composite', 12, 50 + value % 45, 50 + value % 45,
       50 + value % 45, 50 + value % 45, 50 + value % 45, 50 + value % 45,
       50 + value % 45, 50 + value % 45, 'draft'
FROM generate_series(1, 5000) AS value
CROSS JOIN fixture_refs;

INSERT INTO evidence.sources (
    external_key, source_type, title, publisher, is_fictional, review_status,
    reviewed_at, review_note
) VALUES (
    'phase8-scale-fixture', 'demo_fixture', 'Phase 8 rollback-only scale fixture',
    'UNSOLERO validation tooling', true, 'verified', now(),
    'Explicitly fictional and rolled back after query-plan validation.'
);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), 100,
       'Explicitly fictional Phase 8 rollback-only observation.'
FROM evidence.sources AS sources
CROSS JOIN catalog.products AS products
WHERE sources.external_key = 'phase8-scale-fixture'
  AND products.slug LIKE 'phase8-synthetic-product-%';

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
    max_capacity_grams, material, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT id, 1, category_id, brand_id, name, slug, description, price_minor,
       currency, length_mm, width_mm, height_mm, weight_grams,
       max_capacity_grams, material, warranty_months, 'published', now(), now(),
       now(), 'Explicitly fictional Phase 8 rollback-only fact revision.'
FROM catalog.products WHERE slug LIKE 'phase8-synthetic-product-%';

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score,
    durability_score, beginner_score, advanced_score, apartment_score,
    noise_score, portability_score, workflow_status, submitted_at, reviewed_at,
    published_at, review_note
)
SELECT products.id, facts.id, 1, products.quality_score, products.value_score,
       products.durability_score, products.beginner_score, products.advanced_score,
       products.apartment_score, products.noise_score, products.portability_score,
       'published', now(), now(), now(),
       'Explicitly fictional Phase 8 rollback-only score revision.'
FROM catalog.products AS products
JOIN evidence.product_fact_revisions AS facts ON facts.product_id = products.id AND facts.version = 1
WHERE products.slug LIKE 'phase8-synthetic-product-%';

INSERT INTO evidence.fact_provenance (
    fact_revision_id, fact_key, observation_id, public_classification
)
SELECT facts.id, keys.fact_key, observations.id, 'editorial_assessment'
FROM evidence.product_fact_revisions AS facts
JOIN catalog.products AS products ON products.id = facts.product_id
JOIN evidence.observations AS observations ON observations.product_id = products.id
CROSS JOIN (VALUES ('category'), ('brand'), ('name'), ('slug'), ('description'),
    ('price'), ('dimensions'), ('weight'), ('max_capacity'), ('material'), ('warranty')) AS keys(fact_key)
WHERE products.slug LIKE 'phase8-synthetic-product-%';

INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
SELECT scores.id, keys.score_key,
       'Explicitly fictional Phase 8 rollback-only scoring rationale.', observations.id
FROM evidence.score_revisions AS scores
JOIN catalog.products AS products ON products.id = scores.product_id
JOIN evidence.observations AS observations ON observations.product_id = products.id
CROSS JOIN (VALUES ('quality'), ('value'), ('durability'), ('beginner'),
    ('advanced'), ('apartment'), ('noise'), ('portability')) AS keys(score_key)
WHERE products.slug LIKE 'phase8-synthetic-product-%';

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores ON scores.product_id = facts.product_id
WHERE products.id = facts.product_id
  AND products.slug LIKE 'phase8-synthetic-product-%';

\echo 'Creating 10,000 offers, 20,000 observations, and 50,000 clicks'
WITH merchants AS (
    SELECT id, row_number() OVER (ORDER BY id) AS merchant_number
    FROM commerce.merchants WHERE status = 'active' ORDER BY id LIMIT 2
)
INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at,
    provider_observed_at, imported_at, expires_at, is_active
)
SELECT merchants.id, products.id,
       format('phase8-%s-%s', merchants.merchant_number, products.slug),
       format('https://example.invalid/phase8/%s/%s', merchants.merchant_number, products.slug),
       products.price_minor + merchants.merchant_number * 100, 0, products.currency,
       'in_stock', 'new', now(), now(), now(), now() + interval '3 days', true
FROM catalog.products AS products
CROSS JOIN merchants
WHERE products.slug LIKE 'phase8-synthetic-product-%';

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, disclosure_label, is_active,
    priority, commission_type
)
SELECT offers.id, 'phase8-disabled', offers.product_url, 'Fictional validation link',
       true, 0, 'unknown'
FROM commerce.merchant_offers AS offers
WHERE offers.merchant_sku LIKE 'phase8-%';

INSERT INTO commerce.provider_configurations (
    merchant_id, provider_key, adapter_key, external_merchant_id,
    credential_reference, lifecycle_status, configuration_verified_at,
    conversion_ingestion_enabled, conversion_configuration_verified_at
)
SELECT merchants.id, format('phase8scale%s', row_number() OVER (ORDER BY merchants.id)),
       'disabled', format('phase8-merchant-%s', row_number() OVER (ORDER BY merchants.id)),
       'phase8/validation', 'configured', now(), true, now()
FROM commerce.merchants AS merchants
WHERE merchants.id IN (
    SELECT DISTINCT merchant_id FROM commerce.merchant_offers WHERE merchant_sku LIKE 'phase8-%'
);

INSERT INTO commerce.offer_import_runs (
    provider_configuration_id, trigger_type, status, idempotency_key,
    attempt_count, records_received, records_applied, started_at, completed_at
)
SELECT id, 'manual', 'succeeded', 'phase8-scale-offers', 1, 10000, 10000, now(), now()
FROM commerce.provider_configurations WHERE provider_key LIKE 'phase8scale%';

INSERT INTO commerce.price_observations (
    provider_configuration_id, import_run_id, merchant_offer_id,
    external_offer_id, price_minor, shipping_minor, currency,
    provider_observed_at, observed_at, expires_at, observation_fingerprint
)
SELECT configurations.id, runs.id, offers.id, offers.merchant_sku, offers.price_minor,
       offers.shipping_minor, offers.currency, now(), now(), now() + interval '3 days',
       md5(configurations.id::text || offers.id::text || ':price:1') ||
       md5(configurations.id::text || offers.id::text || ':price:2')
FROM commerce.merchant_offers AS offers
JOIN commerce.provider_configurations AS configurations ON configurations.merchant_id = offers.merchant_id
    AND configurations.provider_key LIKE 'phase8scale%'
JOIN commerce.offer_import_runs AS runs ON runs.provider_configuration_id = configurations.id
    AND runs.idempotency_key = 'phase8-scale-offers'
WHERE offers.merchant_sku LIKE 'phase8-%';

INSERT INTO commerce.availability_observations (
    provider_configuration_id, import_run_id, merchant_offer_id,
    external_offer_id, availability, provider_observed_at, observed_at,
    expires_at, observation_fingerprint
)
SELECT configurations.id, runs.id, offers.id, offers.merchant_sku, 'in_stock',
       now(), now(), now() + interval '3 days',
       md5(configurations.id::text || offers.id::text || ':availability:1') ||
       md5(configurations.id::text || offers.id::text || ':availability:2')
FROM commerce.merchant_offers AS offers
JOIN commerce.provider_configurations AS configurations ON configurations.merchant_id = offers.merchant_id
    AND configurations.provider_key LIKE 'phase8scale%'
JOIN commerce.offer_import_runs AS runs ON runs.provider_configuration_id = configurations.id
    AND runs.idempotency_key = 'phase8-scale-offers'
WHERE offers.merchant_sku LIKE 'phase8-%';

INSERT INTO commerce.affiliate_clicks (
    affiliate_link_id, merchant_offer_id, product_id, anonymous_id, session_id,
    source, request_id, traffic_source, traffic_medium, classification,
    is_countable, idempotency_key, retention_expires_at, occurred_at
)
SELECT links.id, offers.id, offers.product_id,
       format('phase8-anonymous-%s-%s', click_number, offers.id),
       format('phase8-session-%s-%s', click_number, offers.id), 'product_detail',
       format('phase8-request-%s-%s', click_number, offers.id), 'phase8', 'validation',
       'human', true, format('phase8-click-%s-%s', click_number, offers.id),
       now() + interval '397 days', now() - (click_number % 7) * interval '1 day'
FROM commerce.merchant_offers AS offers
JOIN commerce.affiliate_links AS links ON links.merchant_offer_id = offers.id
CROSS JOIN generate_series(1, 5) AS click_number
WHERE offers.merchant_sku LIKE 'phase8-%';

\echo 'Creating 10,000 recommendation snapshots and 100,000 analytics events'
INSERT INTO recommendation.recommendation_sessions (
    user_id, status, primary_goal, experience_level, budget_minor, currency,
    space_length_mm, space_width_mm, space_height_mm, apartment_living,
    started_at, completed_at
)
SELECT id, 'completed', 'build_muscle', 'beginner', 150000, 'USD',
       3000, 3000, 2400, false, now() - interval '1 minute', now()
FROM identity.users WHERE email LIKE 'phase8-scale-user-%@example.invalid';

INSERT INTO recommendation.recommendations (
    session_id, policy_version, engine_version, objective_score,
    total_price_minor, currency, result_fingerprint
)
SELECT id, (SELECT version FROM recommendation.policy_versions WHERE workflow_status = 'active' ORDER BY published_at DESC LIMIT 1),
       'phase8-validation', 80, 10000, 'USD', 'phase8-' || id::text
FROM recommendation.recommendation_sessions
WHERE user_id IN (SELECT id FROM identity.users WHERE email LIKE 'phase8-scale-user-%@example.invalid');

WITH fixture_product AS (
    SELECT products.*, categories.slug AS category_slug
    FROM catalog.products AS products
    JOIN catalog.categories AS categories ON categories.id = products.category_id
    WHERE products.slug LIKE 'phase8-synthetic-product-%' ORDER BY products.slug LIMIT 1
)
INSERT INTO recommendation.candidate_snapshots (
    recommendation_id, product_id, fact_revision_id, score_revision_id, name,
    category_slug, price_minor, currency, length_mm, width_mm, height_mm,
    quality_score, value_score, durability_score, beginner_score, advanced_score,
    apartment_score, noise_score, portability_score, policy_version
)
SELECT recommendations.id, products.id, products.published_fact_revision_id,
       products.published_score_revision_id, products.name, products.category_slug,
       products.price_minor, products.currency, products.length_mm, products.width_mm,
       products.height_mm, products.quality_score, products.value_score,
       products.durability_score, products.beginner_score, products.advanced_score,
       products.apartment_score, products.noise_score, products.portability_score,
       recommendations.policy_version
FROM recommendation.recommendations AS recommendations
CROSS JOIN fixture_product AS products
WHERE recommendations.engine_version = 'phase8-validation';

WITH fixture_product AS (
    SELECT id, price_minor, currency FROM catalog.products
    WHERE slug LIKE 'phase8-synthetic-product-%' ORDER BY slug LIMIT 1
)
INSERT INTO recommendation.recommendation_items (
    recommendation_id, product_id, item_type, rank, quantity, unit_price_minor,
    currency, objective_score, reason_code, reason_summary
)
SELECT recommendations.id, products.id, 'selected', 1, 1, products.price_minor,
       products.currency, 80, 'phase8.synthetic',
       'Explicitly fictional Phase 8 rollback-only recommendation.'
FROM recommendation.recommendations AS recommendations
CROSS JOIN fixture_product AS products
WHERE recommendations.engine_version = 'phase8-validation';

WITH users AS (
    SELECT id, row_number() OVER (ORDER BY id) AS user_number
    FROM identity.users WHERE email LIKE 'phase8-scale-user-%@example.invalid'
), fixture_product AS (
    SELECT id FROM catalog.products WHERE slug LIKE 'phase8-synthetic-product-%' ORDER BY slug LIMIT 1
)
INSERT INTO analytics.events (
    public_event_id, event_name, schema_version, user_id, session_id, surface,
    properties, page_path, traffic_source, traffic_medium, consent_state,
    consent_policy_version, origin, classification, is_reportable, occurred_at,
    received_at, retention_expires_at
)
SELECT gen_random_uuid(),
       CASE WHEN event_number % 4 = 0 THEN 'product_viewed'
            WHEN event_number % 4 = 1 THEN 'page_view'
            WHEN event_number % 4 = 2 THEN 'onboarding_started'
            ELSE 'onboarding_completed' END,
       3, users.id, format('phase8-session-%s', users.user_number), 'phase8_scale',
       jsonb_build_object('product_id', fixture_product.id::text,
                          'onboarding_id', format('phase8-onboarding-%s', users.user_number)),
       '/phase8-validation', 'phase8', 'validation', 'granted', 'analytics-v1',
       'server', 'human', true, now() - (event_number % 30) * interval '1 day',
       now(), now() + interval '397 days'
FROM users CROSS JOIN generate_series(1, 10) AS event_number CROSS JOIN fixture_product;

\echo 'Creating 50,000 security events, 50,000 audit rows, and 10,000 verified conversion projections'
INSERT INTO identity.security_events (user_id, event_type, outcome, request_id, surface, metadata, occurred_at)
SELECT users.id, 'phase8.validation', 'success',
       format('phase8-security-%s-%s', event_number, users.id), 'validation',
       jsonb_build_object('fixture', true), now() - (event_number % 30) * interval '1 day'
FROM identity.users AS users CROSS JOIN generate_series(1, 5) AS event_number
WHERE users.email LIKE 'phase8-scale-user-%@example.invalid';

INSERT INTO admin.audit_log (actor_user_id, action, entity_type, entity_id, changes, occurred_at)
SELECT users.id, 'phase8.validation', 'synthetic_fixture',
       format('phase8-%s-%s', event_number, users.id), jsonb_build_object('fixture', true),
       now() - (event_number % 30) * interval '1 day'
FROM identity.users AS users CROSS JOIN generate_series(1, 5) AS event_number
WHERE users.email LIKE 'phase8-scale-user-%@example.invalid';

WITH ranked_clicks AS (
    SELECT clicks.*, row_number() OVER (ORDER BY clicks.id) AS item_number,
           offers.merchant_id, configurations.id AS configuration_id,
           configurations.provider_key
    FROM commerce.affiliate_clicks AS clicks
    JOIN commerce.merchant_offers AS offers ON offers.id = clicks.merchant_offer_id
    JOIN commerce.provider_configurations AS configurations ON configurations.merchant_id = offers.merchant_id
        AND configurations.provider_key LIKE 'phase8scale%'
    WHERE clicks.idempotency_key LIKE 'phase8-click-%'
)
INSERT INTO commerce.affiliate_conversions (
    affiliate_click_id, provider, external_conversion_id, order_status,
    order_value_minor, order_currency, commission_amount_minor, commission_currency,
    converted_at, confirmed_at, provider_configuration_id, merchant_id,
    event_type, event_received_at, commission_status, attribution_status,
    source, verification_state, latest_event_at
)
SELECT id, provider_key, format('phase8-conversion-%s', item_number), 'confirmed',
       10000, 'USD', 500, 'USD', now(), now(), configuration_id, merchant_id,
       'conversion_created', now(), 'approved', 'attributed', source, 'verified', now()
FROM ranked_clicks WHERE item_number <= 10000;

INSERT INTO commerce.conversion_reconciliation_runs (
    provider_configuration_id, idempotency_key, status, coverage_start,
    coverage_end, matched_count, started_at, completed_at
)
SELECT id, 'phase8-scale-reconciliation', 'succeeded', now() - interval '90 days',
       now() + interval '1 day', 5000, now(), now()
FROM commerce.provider_configurations WHERE provider_key LIKE 'phase8scale%';

ANALYZE identity.users;
ANALYZE identity.sessions;
ANALYZE catalog.products;
ANALYZE recommendation.recommendation_sessions;
ANALYZE recommendation.recommendations;
ANALYZE recommendation.candidate_snapshots;
ANALYZE commerce.merchant_offers;
ANALYZE commerce.affiliate_clicks;
ANALYZE commerce.affiliate_conversions;
ANALYZE analytics.events;
ANALYZE identity.security_events;
ANALYZE admin.audit_log;

\echo 'Synthetic row counts'
SELECT 'users' AS relation, count(*) FROM identity.users WHERE email LIKE 'phase8-scale-user-%@example.invalid'
UNION ALL SELECT 'sessions', count(*) FROM identity.sessions WHERE user_id IN (SELECT id FROM identity.users WHERE email LIKE 'phase8-scale-user-%@example.invalid')
UNION ALL SELECT 'products', count(*) FROM catalog.products WHERE slug LIKE 'phase8-synthetic-product-%'
UNION ALL SELECT 'offers', count(*) FROM commerce.merchant_offers WHERE merchant_sku LIKE 'phase8-%'
UNION ALL SELECT 'clicks', count(*) FROM commerce.affiliate_clicks WHERE idempotency_key LIKE 'phase8-click-%'
UNION ALL SELECT 'conversions', count(*) FROM commerce.affiliate_conversions WHERE external_conversion_id LIKE 'phase8-conversion-%'
UNION ALL SELECT 'analytics_events', count(*) FROM analytics.events WHERE surface = 'phase8_scale'
UNION ALL SELECT 'security_events', count(*) FROM identity.security_events WHERE event_type = 'phase8.validation'
UNION ALL SELECT 'recommendation_snapshots', count(*) FROM recommendation.candidate_snapshots JOIN recommendation.recommendations ON recommendations.id = recommendation_id WHERE recommendations.engine_version = 'phase8-validation'
UNION ALL SELECT 'audit_rows', count(*) FROM admin.audit_log WHERE action = 'phase8.validation'
ORDER BY relation;

\echo 'Authentication lookup plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT id, email, status FROM identity.users
WHERE lower(email) = 'phase8-scale-user-9999@example.invalid';

\echo 'Session resolution plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT sessions.id, users.id FROM identity.sessions AS sessions
JOIN identity.users AS users ON users.id = sessions.user_id
WHERE sessions.token_hash = decode(md5('phase8-scale-user-9999@example.invalid:1') ||
    md5('phase8:phase8-scale-user-9999@example.invalid:1'), 'hex')
  AND sessions.revoked_at IS NULL AND sessions.expires_at > now();

\echo 'Public catalog pagination plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT products.id, products.name, products.price_minor, count(*) OVER()
FROM catalog.products AS products
WHERE products.status = 'published'
  AND products.published_fact_revision_id IS NOT NULL
  AND products.published_score_revision_id IS NOT NULL
ORDER BY products.price_minor, products.name LIMIT 24 OFFSET 2400;

\echo 'Recommendation candidate catalog plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT products.id, products.price_minor, products.quality_score,
       products.value_score, products.apartment_score
FROM catalog.products AS products
JOIN evidence.product_fact_revisions AS facts ON facts.id = products.published_fact_revision_id
JOIN evidence.score_revisions AS scores ON scores.id = products.published_score_revision_id
WHERE products.status = 'published' AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published'
ORDER BY products.id;

\echo 'Analytics product ranking plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT properties->>'product_id', count(*)
FROM analytics.events
WHERE is_reportable AND event_name = 'product_viewed'
  AND occurred_at >= now() - interval '90 days'
GROUP BY properties->>'product_id' ORDER BY count(*) DESC LIMIT 20;

\echo 'Affiliate attribution plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT merchant_offer_id, count(*)
FROM commerce.affiliate_clicks
WHERE is_countable AND occurred_at >= now() - interval '30 days'
GROUP BY merchant_offer_id ORDER BY count(*) DESC LIMIT 20;

\echo 'Verified conversion reporting plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT commission_currency, sum(commission_amount_minor)
FROM commerce.affiliate_conversions
WHERE verification_state = 'verified' AND commission_status IN ('approved', 'paid')
  AND order_status = 'confirmed' AND latest_event_at >= now() - interval '90 days'
GROUP BY commission_currency;

\echo 'Retention cleanup selection plan'
EXPLAIN (ANALYZE, BUFFERS)
SELECT id FROM analytics.events
WHERE retention_expires_at < now() ORDER BY retention_expires_at, id LIMIT 1000;

\echo 'High-volume relation sizes before rollback'
SELECT n.nspname || '.' || c.relname AS relation,
       pg_size_pretty(pg_total_relation_size(c.oid)) AS total_size
FROM pg_class AS c JOIN pg_namespace AS n ON n.oid = c.relnamespace
WHERE n.nspname IN ('identity', 'catalog', 'recommendation', 'commerce', 'analytics', 'admin')
  AND c.relkind = 'r'
ORDER BY pg_total_relation_size(c.oid) DESC LIMIT 20;

ROLLBACK;
\echo 'Phase 8 scale fixture rolled back; no synthetic facts persisted.'
