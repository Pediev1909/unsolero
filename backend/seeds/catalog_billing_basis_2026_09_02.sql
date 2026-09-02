-- Billing basis audit, 2026-09-02.
--
-- Every published software product was re-read on its vendor's pricing page
-- on 2026-09-02. This file records, per product, what basis its compared price
-- is on (billing_period, pricing_unit, pricing_unit_note, annual_price_minor
-- — columns added by migration 000029) and, where the site rule requires it,
-- moves the compared price to the monthly-billing list price. The rule and
-- its reasoning are the Zoho Books correction of 2026-08-26: UNSOLERO
-- compares monthly-billing prices, and a per-month figure that only exists on
-- an annual contract is marked `annual` rather than passed off as monthly.
--
-- Each product gets a new published fact revision (the current one is
-- superseded), a score revision carried forward unchanged, an evidence
-- source and observation for today's read, and — when the price moved — its
-- live merchant offers updated to the same figure. Re-running the file is a
-- no-op: each block exits when a revision carrying the audit marker exists.
--
-- Products skipped because the page could not be read reliably (0):
--   none
--
-- Price moves (6):
--   figma-professional: 1600 -> 2000 (monthly billing available)
--   zapier-professional: 1999 -> 2999 (monthly billing available)
--   bigcommerce-core: 2900 -> 3900 (monthly billing available)
--   se-ranking-core: 10320 -> 12900 (monthly billing available)
--   semrush-seo: 11733 -> 13900 (monthly billing available)
--   teachable-starter: 2900 -> 3900 (monthly billing available)
--
-- Applied with psql:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f backend/seeds/catalog_billing_basis_2026_09_02.sql

BEGIN;

-- Stripe (stripe) — usage-based; price unchanged at 0
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'stripe';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'stripe: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'stripe: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('stripe-pricing-2026-09-02', 'manufacturer_documentation', 'Stripe pricing page',
            'Stripe', 'https://stripe.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page geolocates to EEA rates (1.5% + €0.25) and confirms no monthly fee; the US rate is the catalog''s own read from 2026-08-21 (description).')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Stripe read on 2026-09-02 at https://stripe.com/pricing: 0 USD per month billed monthly; unit: 2.9% + 30¢ per successful domestic card charge. Page geolocates to EEA rates (1.5% + €0.25) and confirms no monthly fee; the US rate is the catalog''s own read from 2026-08-21 (description).')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        0, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'usage', 'per_transaction', '2.9% + 30¢ per successful domestic card charge', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: usage-based; period=usage, unit=per_transaction, note=2.9% + 30¢ per successful domestic card charge. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 0,
        billing_period = 'usage',
        pricing_unit = 'per_transaction',
        pricing_unit_note = '2.9% + 30¢ per successful domestic card charge',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Zoho Invoice (zoho-invoice) — free tier; price unchanged at 0
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zoho-invoice';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zoho-invoice: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zoho-invoice: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zoho-invoice-pricing-2026-09-02', 'manufacturer_documentation', 'Zoho Invoice pricing page',
            'Zoho', 'https://www.zoho.com/invoice/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. $0; 2 users, 3 projects, 500 invoices a year.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Zoho Invoice read on 2026-09-02 at https://www.zoho.com/invoice/pricing/: 0 USD per month billed monthly; unit: flat. $0; 2 users, 3 projects, 500 invoices a year.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        0, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'free', 'flat', NULL, NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: free tier; period=free, unit=flat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 0,
        billing_period = 'free',
        pricing_unit = 'flat',
        pricing_unit_note = NULL,
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Ahrefs Starter (ahrefs-starter) — monthly only; price unchanged at 2900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'ahrefs-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'ahrefs-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'ahrefs-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('ahrefs-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Ahrefs Starter pricing page',
            'Ahrefs', 'https://ahrefs.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Starter has no annual option; annual discount applies to Lite and above.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Ahrefs Starter read on 2026-09-02 at https://ahrefs.com/pricing: 29 USD per month billed monthly; unit: flat. Starter has no annual option; annual discount applies to Lite and above.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', NULL, NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly only; period=monthly, unit=flat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2900,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = NULL,
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Framer Basic (framer-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'framer-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'framer-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'framer-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('framer-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Framer Basic pricing page',
            'Framer', 'https://www.framer.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders $10 under yearly billing; monthly figure not shown. Catalog description: per site per month billed yearly. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Framer Basic read on 2026-09-02 at https://www.framer.com/pricing/: no monthly billing; 10 USD per month billed annually; unit: per site; extra editors $20/month. Page renders $10 under yearly billing; monthly figure not shown. Catalog description: per site per month billed yearly. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'flat', 'per site; extra editors $20/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=flat, note=per site; extra editors $20/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1000,
        billing_period = 'annual',
        pricing_unit = 'flat',
        pricing_unit_note = 'per site; extra editors $20/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Make Core (make-core) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1200
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'make-core';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'make-core: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'make-core: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('make-core-pricing-2026-09-02', 'manufacturer_documentation', 'Make Core pricing page',
            'Make', 'https://www.make.com/en/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Catalog description (Aug 2026): $12 per month billed annually, monthly $16. Page today shows only $9 with the toggle state hidden. Re-read in a browser before moving anything.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Make Core read on 2026-09-02 at https://www.make.com/en/pricing: no monthly billing; unit: at 10,000 credits/month. Catalog description (Aug 2026): $12 per month billed annually, monthly $16. Page today shows only $9 with the toggle state hidden. Re-read in a browser before moving anything.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1200, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'usage', 'at 10,000 credits/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=usage, note=at 10,000 credits/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1200,
        billing_period = 'annual',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 10,000 credits/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Wave Starter (wave-starter) — free tier; price unchanged at 0
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'wave-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'wave-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'wave-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('wave-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Wave Starter pricing page',
            'Wave', 'https://www.waveapps.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Starter is free; Pro $19/mo or $190/yr.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Wave Starter read on 2026-09-02 at https://www.waveapps.com/pricing: 0 USD per month billed monthly; unit: flat. Starter is free; Pro $19/mo or $190/yr.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        0, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'free', 'flat', NULL, NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: free tier; period=free, unit=flat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 0,
        billing_period = 'free',
        pricing_unit = 'flat',
        pricing_unit_note = NULL,
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Bigin Express (bigin-express) — monthly billing available; price unchanged at 900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'bigin-express';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'bigin-express: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'bigin-express: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('bigin-express-pricing-2026-09-02', 'manufacturer_documentation', 'Bigin Express pricing page',
            'Zoho', 'https://www.bigin.com/pricing.html', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. USD rendered by default.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Bigin Express read on 2026-09-02 at https://www.bigin.com/pricing.html: 9 USD per month billed monthly; 7 USD per month billed annually; unit: per user. USD rendered by default.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 700,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 900,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 700,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Figma Professional (figma-professional) — monthly billing available; price 1600 -> 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'figma-professional';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'figma-professional: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'figma-professional: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('figma-professional-pricing-2026-09-02', 'manufacturer_documentation', 'Figma Professional pricing page',
            'Figma', 'https://www.figma.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders €16 annual only. Monthly $20 is the catalog''s own read from the vendor page on 2026-08-21 (product description); annual $16 read today. Compared price moves to the monthly-billing figure.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Figma Professional read on 2026-09-02 at https://www.figma.com/pricing/: 20 USD per month billed monthly; 16 USD per month billed annually; unit: per full seat; Dev seat and Collab seat cheaper. Page renders €16 annual only. Monthly $20 is the catalog''s own read from the vendor page on 2026-08-21 (product description); annual $16 read today. Compared price moves to the monthly-billing figure.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per full seat; Dev seat and Collab seat cheaper', 1600,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per full seat; Dev seat and Collab seat cheaper. Compared price moved from 1600 to 2000 minor units (monthly-billing list price rule).')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per full seat; Dev seat and Collab seat cheaper',
        annual_price_minor = 1600,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 2000, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Zoho Projects Premium (zoho-projects-premium) — annual contract only; price unchanged at 400
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zoho-projects-premium';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zoho-projects-premium: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zoho-projects-premium: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zoho-projects-premium-pricing-2026-09-02', 'manufacturer_documentation', 'Zoho Projects Premium pricing page',
            'Zoho', 'https://www.zoho.com/projects/pricing.html', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page (EUR from Bulgaria) shows €4/user/month billed annually and no monthly rate; catalog $4 is the USD annual figure read 2026-08-21.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Zoho Projects Premium read on 2026-09-02 at https://www.zoho.com/projects/pricing.html: no monthly billing; 4 USD per month billed annually; unit: per user. Page (EUR from Bulgaria) shows €4/user/month billed annually and no monthly rate; catalog $4 is the USD annual figure read 2026-08-21.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        400, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per user', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: annual contract only; period=annual, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 400,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Calendly Standard (calendly-standard) — monthly billing available; price unchanged at 1000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'calendly-standard';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'calendly-standard: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'calendly-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('calendly-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Calendly Standard pricing page',
            'Calendly', 'https://calendly.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Calendly Standard read on 2026-09-02 at https://calendly.com/pricing: 10 USD per month billed monthly; 8.3 USD per month billed annually; unit: per seat')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per seat', 830,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per seat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1000,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per seat',
        annual_price_minor = 830,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Fathom Analytics (fathom-analytics) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1500
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'fathom-analytics';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'fathom-analytics: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'fathom-analytics: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('fathom-analytics-pricing-2026-09-02', 'manufacturer_documentation', 'Fathom Analytics pricing page',
            'Fathom Analytics', 'https://usefathom.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Only the 500k tier ($45) renders; the 100k tier price sits behind a slider. Fathom bills monthly with 2 months free on yearly. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Fathom Analytics read on 2026-09-02 at https://usefathom.com/pricing: no monthly billing; unit: at 100,000 pageviews/month. Only the 500k tier ($45) renders; the 100k tier price sits behind a slider. Fathom bills monthly with 2 months free on yearly. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1500, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'usage', 'at 100,000 pageviews/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=usage, note=at 100,000 pageviews/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1500,
        billing_period = 'monthly',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 100,000 pageviews/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- MailerLite Comfort (mailerlite-comfort) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'mailerlite-comfort';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'mailerlite-comfort: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'mailerlite-comfort: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('mailerlite-comfort-pricing-2026-09-02', 'manufacturer_documentation', 'MailerLite Comfort pricing page',
            'MailerLite', 'https://www.mailerlite.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Subscriber-tier table is JavaScript-only and the page timed out headless. Catalog description: monthly billing at 1,000 subscribers. Free plan 250 subscribers / 2,500 emails confirmed by the owner 2026-09-02.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'MailerLite Comfort read on 2026-09-02 at https://www.mailerlite.com/pricing: no monthly billing; unit: at 1,000 subscribers. Subscriber-tier table is JavaScript-only and the page timed out headless. Catalog description: monthly billing at 1,000 subscribers. Free plan 250 subscribers / 2,500 emails confirmed by the owner 2026-09-02.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_contacts', 'at 1,000 subscribers', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=per_contacts, note=at 1,000 subscribers. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1900,
        billing_period = 'monthly',
        pricing_unit = 'per_contacts',
        pricing_unit_note = 'at 1,000 subscribers',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Zoho Books Standard (zoho-books-standard) — monthly billing available; price unchanged at 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zoho-books-standard';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zoho-books-standard: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zoho-books-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zoho-books-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Zoho Books Standard pricing page',
            'Zoho', 'https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Zoho Books Standard read on 2026-09-02 at https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard: 20 USD per month billed monthly; 15 USD per month billed annually; unit: per organization; 3 users included')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per organization; 3 users included', 1500,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat, note=per organization; 3 users included. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per organization; 3 users included',
        annual_price_minor = 1500,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Zoho Campaigns Standard (zoho-campaigns-standard) — annual contract only; price unchanged at 525
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zoho-campaigns-standard';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zoho-campaigns-standard: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zoho-campaigns-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zoho-campaigns-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Zoho Campaigns Standard pricing page',
            'Zoho', 'https://www.zoho.com/campaigns/pricing.html', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page (EUR) shows Standard billed annually with no monthly rate; catalog $5.25 is the USD annual figure at 1,000 contacts read 2026-08-21.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Zoho Campaigns Standard read on 2026-09-02 at https://www.zoho.com/campaigns/pricing.html: no monthly billing; 5.25 USD per month billed annually; unit: at 1,000 contacts. Page (EUR) shows Standard billed annually with no monthly rate; catalog $5.25 is the USD annual figure at 1,000 contacts read 2026-08-21.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        525, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_contacts', 'at 1,000 contacts', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: annual contract only; period=annual, unit=per_contacts, note=at 1,000 contacts. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 525,
        billing_period = 'annual',
        pricing_unit = 'per_contacts',
        pricing_unit_note = 'at 1,000 contacts',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Canva Pro (canva-pro) — annual contract only; price unchanged at 1500
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'canva-pro';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'canva-pro: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'canva-pro: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('canva-pro-pricing-2026-09-02', 'manufacturer_documentation', 'Canva Pro pricing page',
            'Canva', 'https://www.canva.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. canva.com returns 403 to automated reads. Catalog description: 180 USD a year for one person and no monthly rate.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Canva Pro read on 2026-09-02 at https://www.canva.com/pricing/: no monthly billing; 15 USD per month billed annually; unit: for one person. canva.com returns 403 to automated reads. Catalog description: 180 USD a year for one person and no monthly rate.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1500, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'for one person', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: annual contract only; period=annual, unit=per_user, note=for one person. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1500,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'for one person',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- ClickUp Unlimited (clickup-unlimited) — monthly billing available; price unchanged at 1000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'clickup-unlimited';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'clickup-unlimited: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'clickup-unlimited: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('clickup-unlimited-pricing-2026-09-02', 'manufacturer_documentation', 'ClickUp Unlimited pricing page',
            'ClickUp', 'https://clickup.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'ClickUp Unlimited read on 2026-09-02 at https://clickup.com/pricing: 10 USD per month billed monthly; 7 USD per month billed annually; unit: per user')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 700,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1000,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 700,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Google Workspace Business Starter (google-workspace-business-starter) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 700
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'google-workspace-business-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'google-workspace-business-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'google-workspace-business-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('google-workspace-business-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Google Workspace Business Starter pricing page',
            'Google Workspace', 'https://workspace.google.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR only (€6.80 flexible). Catalog $7 is the annual-commitment price; the flexible price was not read. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Google Workspace Business Starter read on 2026-09-02 at https://workspace.google.com/pricing: no monthly billing; unit: per user, annual commitment; flexible monthly plan costs more. Page renders EUR only (€6.80 flexible). Catalog $7 is the annual-commitment price; the flexible price was not read. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        700, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per user, annual commitment; flexible monthly plan costs more', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=per_user, note=per user, annual commitment; flexible monthly plan costs more. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 700,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user, annual commitment; flexible monthly plan costs more',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Microsoft Teams Essentials (microsoft-teams-essentials) — annual contract only; price unchanged at 400
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'microsoft-teams-essentials';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'microsoft-teams-essentials: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'microsoft-teams-essentials: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('microsoft-teams-essentials-pricing-2026-09-02', 'manufacturer_documentation', 'Microsoft Teams Essentials pricing page',
            'Microsoft', 'https://www.microsoft.com/en-us/microsoft-teams/compare-microsoft-teams-business-options', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. $4.00 user/month paid yearly; no monthly-commitment price shown.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Microsoft Teams Essentials read on 2026-09-02 at https://www.microsoft.com/en-us/microsoft-teams/compare-microsoft-teams-business-options: no monthly billing; 4 USD per month billed annually; unit: per user. $4.00 user/month paid yearly; no monthly-commitment price shown.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        400, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per user', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: annual contract only; period=annual, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 400,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Notion Plus (notion-plus) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'notion-plus';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'notion-plus: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'notion-plus: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('notion-plus-pricing-2026-09-02', 'manufacturer_documentation', 'Notion Plus pricing page',
            'Notion', 'https://www.notion.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR (€9.50). Catalog $10 is Notion''s annual per-member figure; monthly USD not read today. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Notion Plus read on 2026-09-02 at https://www.notion.com/pricing: no monthly billing; unit: per member. Page renders EUR (€9.50). Catalog $10 is Notion''s annual per-member figure; monthly USD not read today. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per member', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=per_user, note=per member. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1000,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per member',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Webflow Basic (webflow-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1500
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'webflow-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'webflow-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'webflow-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('webflow-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Webflow Basic pricing page',
            'Webflow', 'https://webflow.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Headless read: Basic $15/mo billed yearly (USD). Monthly figure not rendered. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Webflow Basic read on 2026-09-02 at https://webflow.com/pricing: no monthly billing; 15 USD per month billed annually; unit: per site. Headless read: Basic $15/mo billed yearly (USD). Monthly figure not rendered. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1500, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'flat', 'per site', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=flat, note=per site. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1500,
        billing_period = 'annual',
        pricing_unit = 'flat',
        pricing_unit_note = 'per site',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Brevo Starter (brevo-starter) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'brevo-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'brevo-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'brevo-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('brevo-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Brevo Starter pricing page',
            'Brevo', 'https://www.brevo.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Fully JavaScript-rendered; timed out headless. Catalog description: monthly billing for 5,000 emails a month. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Brevo Starter read on 2026-09-02 at https://www.brevo.com/pricing/: no monthly billing; unit: at 5,000 emails/month. Fully JavaScript-rendered; timed out headless. Catalog description: monthly billing for 5,000 emails a month. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'usage', 'at 5,000 emails/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=usage, note=at 5,000 emails/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 900,
        billing_period = 'monthly',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 5,000 emails/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Help Scout Standard (help-scout-standard) — monthly billing available; price unchanged at 2500
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'help-scout-standard';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'help-scout-standard: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'help-scout-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('help-scout-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Help Scout Standard pricing page',
            'Help Scout', 'https://www.helpscout.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Help Scout Standard read on 2026-09-02 at https://www.helpscout.com/pricing/: 25 USD per month billed monthly; 21 USD per month billed annually; unit: per user')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2500, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 2100,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2500,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 2100,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Shopify Basic (shopify-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 2900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'shopify-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'shopify-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'shopify-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('shopify-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Shopify Basic pricing page',
            'Shopify', 'https://www.shopify.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR (€27 monthly / €19 yearly). Catalog description: read with the billing control set to pay yearly. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Shopify Basic read on 2026-09-02 at https://www.shopify.com/pricing: no monthly billing; unit: per store. Page renders EUR (€27 monthly / €19 yearly). Catalog description: read with the billing control set to pay yearly. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'flat', 'per store', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=flat, note=per store. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2900,
        billing_period = 'annual',
        pricing_unit = 'flat',
        pricing_unit_note = 'per store',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Slack Pro (slack-pro) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 875
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'slack-pro';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'slack-pro: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'slack-pro: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('slack-pro-pricing-2026-09-02', 'manufacturer_documentation', 'Slack Pro pricing page',
            'Slack', 'https://slack.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR. Catalog description: per user per month paying monthly. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Slack Pro read on 2026-09-02 at https://slack.com/pricing: no monthly billing; unit: per user. Page renders EUR. Catalog description: per user per month paying monthly. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        875, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 875,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Umami Cloud Pro (umami-cloud-pro) — monthly only; price unchanged at 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'umami-cloud-pro';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'umami-cloud-pro: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'umami-cloud-pro: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('umami-cloud-pro-pricing-2026-09-02', 'manufacturer_documentation', 'Umami Cloud Pro pricing page',
            'Umami', 'https://umami.is/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Headless read: Pro $20/month, usage-based, no annual option shown.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Umami Cloud Pro read on 2026-09-02 at https://umami.is/pricing: 20 USD per month billed monthly; unit: 1 million events/month included. Headless read: Pro $20/month, usage-based, no annual option shown.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'usage', '1 million events/month included', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly only; period=monthly, unit=usage, note=1 million events/month included. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'monthly',
        pricing_unit = 'usage',
        pricing_unit_note = '1 million events/month included',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Zoho Bookings Basic (zoho-bookings-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 800
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zoho-bookings-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zoho-bookings-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zoho-bookings-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zoho-bookings-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Zoho Bookings Basic pricing page',
            'Zoho', 'https://www.zoho.com/bookings/pricing.html', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR (€6 billed annually). Catalog description: per user per month; $8 is the monthly figure. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Zoho Bookings Basic read on 2026-09-02 at https://www.zoho.com/bookings/pricing.html: no monthly billing; unit: per user. Page renders EUR (€6 billed annually). Catalog description: per user per month; $8 is the monthly figure. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        800, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 800,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Zoho CRM Standard (zoho-crm-standard) — monthly billing available; price unchanged at 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zoho-crm-standard';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zoho-crm-standard: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zoho-crm-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zoho-crm-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Zoho CRM Standard pricing page',
            'Zoho', 'https://www.zoho.com/crm/zohocrm-pricing.html', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR/INR today. Monthly $20 and annual $14 are the catalog''s own read from 2026-08-21 (product description).')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Zoho CRM Standard read on 2026-09-02 at https://www.zoho.com/crm/zohocrm-pricing.html: 20 USD per month billed monthly; 14 USD per month billed annually; unit: per user. Page renders EUR/INR today. Monthly $20 and annual $14 are the catalog''s own read from 2026-08-21 (product description).')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 1400,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 1400,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- n8n Starter (n8n-starter) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'n8n-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'n8n-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'n8n-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('n8n-starter-pricing-2026-09-02', 'manufacturer_documentation', 'n8n Starter pricing page',
            'n8n', 'https://n8n.io/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders €20/mo billed annually; the catalog''s 20 equals the EUR figure and may not be a USD read. Basis recorded, price kept; re-read from a US IP.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'n8n Starter read on 2026-09-02 at https://n8n.io/pricing/: no monthly billing; unit: at 2,500 executions/month. Page renders €20/mo billed annually; the catalog''s 20 equals the EUR figure and may not be a USD read. Basis recorded, price kept; re-read from a US IP.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'usage', 'at 2,500 executions/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=usage, note=at 2,500 executions/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'annual',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 2,500 executions/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Ecwid Starter (ecwid-starter) — monthly billing available; price unchanged at 500
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'ecwid-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'ecwid-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'ecwid-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('ecwid-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Ecwid Starter pricing page',
            'Ecwid', 'https://www.ecwid.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Starter is $5 on both billing cycles.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Ecwid Starter read on 2026-09-02 at https://www.ecwid.com/pricing: 5 USD per month billed monthly; 5 USD per month billed annually; unit: flat. Starter is $5 on both billing cycles.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        500, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', NULL, 500,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 500,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = NULL,
        annual_price_minor = 500,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Freshsales Growth (freshsales-growth) — monthly billing available; price unchanged at 1100
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'freshsales-growth';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'freshsales-growth: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'freshsales-growth: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('freshsales-growth-pricing-2026-09-02', 'manufacturer_documentation', 'Freshsales Growth pricing page',
            'Freshworks', 'https://www.freshworks.com/crm/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders $9 billed annually only. Monthly $11 is the catalog''s read (description: widely reported monthly rate).')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Freshsales Growth read on 2026-09-02 at https://www.freshworks.com/crm/pricing/: 11 USD per month billed monthly; 9 USD per month billed annually; unit: per user. Page renders $9 billed annually only. Monthly $11 is the catalog''s read (description: widely reported monthly rate).')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1100, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1100,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Sketch Standard (sketch-standard) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1200
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'sketch-standard';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'sketch-standard: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'sketch-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('sketch-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Sketch Standard pricing page',
            'Sketch', 'https://www.sketch.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. $12 per editor billed yearly; monthly rate not listed. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Sketch Standard read on 2026-09-02 at https://www.sketch.com/pricing/: no monthly billing; 12 USD per month billed annually; unit: per editor; viewers free. $12 per editor billed yearly; monthly rate not listed. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1200, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per editor; viewers free', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=per_user, note=per editor; viewers free. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1200,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per editor; viewers free',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Zapier Professional (zapier-professional) — monthly billing available; price 1999 -> 2999
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'zapier-professional';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'zapier-professional: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'zapier-professional: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('zapier-professional-pricing-2026-09-02', 'manufacturer_documentation', 'Zapier Professional pricing page',
            'Zapier', 'https://zapier.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Professional 750 tasks: $29.99 monthly, $19.99 billed yearly. Compared price moves to the monthly figure.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Zapier Professional read on 2026-09-02 at https://zapier.com/pricing: 29.99 USD per month billed monthly; 19.99 USD per month billed annually; unit: at 750 tasks/month. Professional 750 tasks: $29.99 monthly, $19.99 billed yearly. Compared price moves to the monthly figure.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2999, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'usage', 'at 750 tasks/month', 1999,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=usage, note=at 750 tasks/month. Compared price moved from 1999 to 2999 minor units (monthly-billing list price rule).')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2999,
        billing_period = 'monthly',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 750 tasks/month',
        annual_price_minor = 1999,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 2999, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Cal.com Teams (cal-com-teams) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1200
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'cal-com-teams';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'cal-com-teams: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'cal-com-teams: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('cal-com-teams-pricing-2026-09-02', 'manufacturer_documentation', 'Cal.com Teams pricing page',
            'Cal.com', 'https://cal.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. $12 per user/month under the YEARLY label (save 25%); monthly figure not displayed. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Cal.com Teams read on 2026-09-02 at https://cal.com/pricing: no monthly billing; 12 USD per month billed annually; unit: per user. $12 per user/month under the YEARLY label (save 25%); monthly figure not displayed. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1200, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per user', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1200,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Squarespace Basic (squarespace-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 1900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'squarespace-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'squarespace-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'squarespace-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('squarespace-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Squarespace Basic pricing page',
            'Squarespace', 'https://www.squarespace.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR (€11 yearly). Catalog description: quoted at monthly billing. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Squarespace Basic read on 2026-09-02 at https://www.squarespace.com/pricing: no monthly billing; unit: per site. Page renders EUR (€11 yearly). Catalog description: quoted at monthly billing. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per site', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=flat, note=per site. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1900,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per site',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Teamwork Basics (teamwork-basics) — monthly billing available; price unchanged at 1299
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'teamwork-basics';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'teamwork-basics: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'teamwork-basics: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('teamwork-basics-pricing-2026-09-02', 'manufacturer_documentation', 'Teamwork Basics pricing page',
            'Teamwork.com', 'https://www.teamwork.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders $9.99 billed yearly; catalog $12.99 is the monthly figure from the 2026-08-21 read.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Teamwork Basics read on 2026-09-02 at https://www.teamwork.com/pricing/: 12.99 USD per month billed monthly; 9.99 USD per month billed annually; unit: per user. Page renders $9.99 billed yearly; catalog $12.99 is the monthly figure from the 2026-08-21 read.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1299, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 999,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1299,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 999,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- BigCommerce Core (bigcommerce-core) — monthly billing available; price 2900 -> 3900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'bigcommerce-core';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'bigcommerce-core: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'bigcommerce-core: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('bigcommerce-core-pricing-2026-09-02', 'manufacturer_documentation', 'BigCommerce Core pricing page',
            'BigCommerce', 'https://www.bigcommerce.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Core $39 monthly / $29 annual. Compared price moves to the monthly figure.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'BigCommerce Core read on 2026-09-02 at https://www.bigcommerce.com/pricing/: 39 USD per month billed monthly; 29 USD per month billed annually; unit: per store. Core $39 monthly / $29 annual. Compared price moves to the monthly figure.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        3900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per store', 2900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat, note=per store. Compared price moved from 2900 to 3900 minor units (monthly-billing list price rule).')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 3900,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per store',
        annual_price_minor = 2900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 3900, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Freshdesk Growth (freshdesk-growth) — monthly billing available; price unchanged at 2300
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'freshdesk-growth';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'freshdesk-growth: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'freshdesk-growth: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('freshdesk-growth-pricing-2026-09-02', 'manufacturer_documentation', 'Freshdesk Growth pricing page',
            'Freshworks', 'https://www.freshworks.com/freshdesk/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders $19 billed annually; catalog description: per agent per month at monthly billing ($23).')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Freshdesk Growth read on 2026-09-02 at https://www.freshworks.com/freshdesk/pricing/: 23 USD per month billed monthly; 19 USD per month billed annually; unit: per agent. Page renders $19 billed annually; catalog description: per agent per month at monthly billing ($23).')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2300, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per agent', 1900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per agent. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2300,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per agent',
        annual_price_minor = 1900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Gumroad (gumroad) — usage-based; price unchanged at 0
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'gumroad';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'gumroad: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'gumroad: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('gumroad-pricing-2026-09-02', 'manufacturer_documentation', 'Gumroad pricing page',
            'Gumroad', 'https://gumroad.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Gumroad read on 2026-09-02 at https://gumroad.com/pricing: 0 USD per month billed monthly; unit: 10% + 50¢ per sale; 30% via Gumroad Discover')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        0, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'usage', 'per_transaction', '10% + 50¢ per sale; 30% via Gumroad Discover', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: usage-based; period=usage, unit=per_transaction, note=10% + 50¢ per sale; 30% via Gumroad Discover. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 0,
        billing_period = 'usage',
        pricing_unit = 'per_transaction',
        pricing_unit_note = '10% + 50¢ per sale; 30% via Gumroad Discover',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Paddle (paddle) — usage-based; price unchanged at 0
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'paddle';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'paddle: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'paddle: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('paddle-pricing-2026-09-02', 'manufacturer_documentation', 'Paddle pricing page',
            'Paddle', 'https://www.paddle.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Paddle read on 2026-09-02 at https://www.paddle.com/pricing: 0 USD per month billed monthly; unit: 5% + 50¢ per checkout transaction')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        0, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'usage', 'per_transaction', '5% + 50¢ per checkout transaction', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: usage-based; period=usage, unit=per_transaction, note=5% + 50¢ per checkout transaction. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 0,
        billing_period = 'usage',
        pricing_unit = 'per_transaction',
        pricing_unit_note = '5% + 50¢ per checkout transaction',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Simple Analytics (simple-analytics) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'simple-analytics';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'simple-analytics: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'simple-analytics: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('simple-analytics-pricing-2026-09-02', 'manufacturer_documentation', 'Simple Analytics pricing page',
            'Simple Analytics', 'https://www.simpleanalytics.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders €20/month; the catalog''s 20 equals the EUR figure. Basis recorded, price kept; re-read in USD.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Simple Analytics read on 2026-09-02 at https://www.simpleanalytics.com/pricing: no monthly billing; unit: at 100,000 datapoints/month. Page renders €20/month; the catalog''s 20 equals the EUR figure. Basis recorded, price kept; re-read in USD.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'usage', 'at 100,000 datapoints/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=monthly, unit=usage, note=at 100,000 datapoints/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'monthly',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 100,000 datapoints/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- FreshBooks Lite (freshbooks-lite) — monthly billing available; price unchanged at 2300
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'freshbooks-lite';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'freshbooks-lite: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'freshbooks-lite: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('freshbooks-lite-pricing-2026-09-02', 'manufacturer_documentation', 'FreshBooks Lite pricing page',
            'FreshBooks', 'https://www.freshbooks.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Yearly billing advertised (10% off) with no figure rendered.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'FreshBooks Lite read on 2026-09-02 at https://www.freshbooks.com/pricing: 23 USD per month billed monthly; unit: per account; extra team members $11/month. Yearly billing advertised (10% off) with no figure rendered.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2300, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per account; extra team members $11/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat, note=per account; extra team members $11/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2300,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per account; extra team members $11/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- HubSpot Starter Customer Platform (hubspot-starter-customer-platform) — monthly billing available; price unchanged at 2000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'hubspot-starter-customer-platform';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'hubspot-starter-customer-platform: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'hubspot-starter-customer-platform: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('hubspot-starter-customer-platform-pricing-2026-09-02', 'manufacturer_documentation', 'HubSpot Starter Customer Platform pricing page',
            'HubSpot', 'https://www.hubspot.com/pricing/crm', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. $20/mo/seat billed monthly; the $7 annual figure shown is a limited-time new-customer promotion, so no regular annual price is recorded.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'HubSpot Starter Customer Platform read on 2026-09-02 at https://www.hubspot.com/pricing/crm: 20 USD per month billed monthly; unit: per seat. $20/mo/seat billed monthly; the $7 annual figure shown is a limited-time new-customer promotion, so no regular annual price is recorded.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per seat', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per seat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2000,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per seat',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Pipedrive Lite (pipedrive-lite) — monthly billing available; price unchanged at 1990
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'pipedrive-lite';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'pipedrive-lite: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'pipedrive-lite: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('pipedrive-lite-pricing-2026-09-02', 'manufacturer_documentation', 'Pipedrive Lite pricing page',
            'Pipedrive', 'https://www.pipedrive.com/en/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. pipedrive.com returns 403 to automated reads. Monthly $19.90 and annual $14 are the catalog''s read from 2026-08-21 (description).')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Pipedrive Lite read on 2026-09-02 at https://www.pipedrive.com/en/pricing: 19.9 USD per month billed monthly; 14 USD per month billed annually; unit: per seat. pipedrive.com returns 403 to automated reads. Monthly $19.90 and annual $14 are the catalog''s read from 2026-08-21 (description).')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1990, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per seat', 1400,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per seat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1990,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per seat',
        annual_price_minor = 1400,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- monday.com Basic (monday-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'monday-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'monday-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'monday-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('monday-basic-pricing-2026-09-02', 'manufacturer_documentation', 'monday.com Basic pricing page',
            'monday.com', 'https://monday.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders €9/seat billed annually; catalog description: per seat per month billed annually. Monthly USD not read. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'monday.com Basic read on 2026-09-02 at https://monday.com/pricing: no monthly billing; unit: per seat. Page renders €9/seat billed annually; catalog description: per seat per month billed annually. Monthly USD not read. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_user', 'per seat', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=per_user, note=per seat. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 900,
        billing_period = 'annual',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per seat',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- ActiveCampaign Starter (activecampaign-starter) — annual contract only; price unchanged at 1500
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'activecampaign-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'activecampaign-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'activecampaign-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('activecampaign-starter-pricing-2026-09-02', 'manufacturer_documentation', 'ActiveCampaign Starter pricing page',
            'ActiveCampaign', 'https://www.activecampaign.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Headless read (USD): Starter $15/mo billed annually; the page publishes no monthly rate.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'ActiveCampaign Starter read on 2026-09-02 at https://www.activecampaign.com/pricing: no monthly billing; 15 USD per month billed annually; unit: at 1,000 contacts. Headless read (USD): Starter $15/mo billed annually; the page publishes no monthly rate.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        1500, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'per_contacts', 'at 1,000 contacts', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: annual contract only; period=annual, unit=per_contacts, note=at 1,000 contacts. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 1500,
        billing_period = 'annual',
        pricing_unit = 'per_contacts',
        pricing_unit_note = 'at 1,000 contacts',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Lemon Squeezy (lemon-squeezy) — usage-based; price unchanged at 0
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'lemon-squeezy';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'lemon-squeezy: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'lemon-squeezy: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('lemon-squeezy-pricing-2026-09-02', 'manufacturer_documentation', 'Lemon Squeezy pricing page',
            'Lemon Squeezy', 'https://www.lemonsqueezy.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Headless read confirms 5% + 50¢ per transaction.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Lemon Squeezy read on 2026-09-02 at https://www.lemonsqueezy.com/pricing: 0 USD per month billed monthly; unit: 5% + 50¢ per transaction. Headless read confirms 5% + 50¢ per transaction.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        0, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'usage', 'per_transaction', '5% + 50¢ per transaction', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: usage-based; period=usage, unit=per_transaction, note=5% + 50¢ per transaction. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 0,
        billing_period = 'usage',
        pricing_unit = 'per_transaction',
        pricing_unit_note = '5% + 50¢ per transaction',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Tidio Starter (tidio-starter) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 2417
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'tidio-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'tidio-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'tidio-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('tidio-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Tidio Starter pricing page',
            'Tidio', 'https://www.tidio.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. $24.17/mo is the annual-billed figure (2 months free); monthly not displayed. Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Tidio Starter read on 2026-09-02 at https://www.tidio.com/pricing/: no monthly billing; 24.17 USD per month billed annually; unit: at 100 billable conversations/month. $24.17/mo is the annual-billed figure (2 months free); monthly not displayed. Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        2417, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'usage', 'at 100 billable conversations/month', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=usage, note=at 100 billable conversations/month. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 2417,
        billing_period = 'annual',
        pricing_unit = 'usage',
        pricing_unit_note = 'at 100 billable conversations/month',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- SE Ranking Core (se-ranking-core) — monthly billing available; price 10320 -> 12900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'se-ranking-core';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'se-ranking-core: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'se-ranking-core: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('se-ranking-core-pricing-2026-09-02', 'manufacturer_documentation', 'SE Ranking Core pricing page',
            'SE Ranking', 'https://seranking.com/subscription.html', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR today (€109 monthly / €87.20 annually, 20% off). Monthly $129 and annual $103.20 are the catalog''s own read from 2026-08-21 (description); the 20% structure matches today''s page. Compared price moves to the monthly figure.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'SE Ranking Core read on 2026-09-02 at https://seranking.com/subscription.html: 129 USD per month billed monthly; 103.2 USD per month billed annually; unit: per account; 10 projects. Page renders EUR today (€109 monthly / €87.20 annually, 20% off). Monthly $129 and annual $103.20 are the catalog''s own read from 2026-08-21 (description); the 20% structure matches today''s page. Compared price moves to the monthly figure.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        12900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per account; 10 projects', 10320,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat, note=per account; 10 projects. Compared price moved from 10320 to 12900 minor units (monthly-billing list price rule).')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 12900,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per account; 10 projects',
        annual_price_minor = 10320,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 12900, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Semrush SEO (semrush-seo) — monthly billing available; price 11733 -> 13900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'semrush-seo';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'semrush-seo: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'semrush-seo: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('semrush-seo-pricing-2026-09-02', 'manufacturer_documentation', 'Semrush SEO pricing page',
            'Semrush', 'https://www.semrush.com/pricing/seo-ai-search/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. SEO $139 monthly / $117.33 annual. Compared price moves to the monthly figure.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Semrush SEO read on 2026-09-02 at https://www.semrush.com/pricing/seo-ai-search/: 139 USD per month billed monthly; 117.33 USD per month billed annually; unit: per account (1 user). SEO $139 monthly / $117.33 annual. Compared price moves to the monthly figure.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        13900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per account (1 user)', 11733,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat, note=per account (1 user). Compared price moved from 11733 to 13900 minor units (monthly-billing list price rule).')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 13900,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per account (1 user)',
        annual_price_minor = 11733,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 13900, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Salesflare Growth (salesflare-growth) — monthly billing available; price unchanged at 3900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'salesflare-growth';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'salesflare-growth: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'salesflare-growth: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('salesflare-growth-pricing-2026-09-02', 'manufacturer_documentation', 'Salesflare Growth pricing page',
            'Salesflare', 'https://www.salesflare.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. ')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Salesflare Growth read on 2026-09-02 at https://www.salesflare.com/pricing: 39 USD per month billed monthly; 29 USD per month billed annually; unit: per user')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        3900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per user', 2900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_user, note=per user. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 3900,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per user',
        annual_price_minor = 2900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Teachable Starter (teachable-starter) — monthly billing available; price 2900 -> 3900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'teachable-starter';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'teachable-starter: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'teachable-starter: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('teachable-starter-pricing-2026-09-02', 'manufacturer_documentation', 'Teachable Starter pricing page',
            'Teachable', 'https://teachable.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Rendered page: $39 monthly, $29 billed annually (JSON-LD contradicts; rendered figures used). Compared price moves to the monthly figure.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Teachable Starter read on 2026-09-02 at https://teachable.com/pricing: 39 USD per month billed monthly; 29 USD per month billed annually; unit: per account; 7.5% transaction fee on Starter. Rendered page: $39 monthly, $29 billed annually (JSON-LD contradicts; rendered figures used). Compared price moves to the monthly figure.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        3900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'flat', 'per account; 7.5% transaction fee on Starter', 2900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=flat, note=per account; 7.5% transaction fee on Starter. Compared price moved from 2900 to 3900 minor units (monthly-billing list price rule).')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 3900,
        billing_period = 'monthly',
        pricing_unit = 'flat',
        pricing_unit_note = 'per account; 7.5% transaction fee on Starter',
        annual_price_minor = 2900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 3900, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Thinkific Basic (thinkific-basic) — basis recorded from the page; USD figure not readable today, price kept; price unchanged at 4000
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'thinkific-basic';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'thinkific-basic: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'thinkific-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('thinkific-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Thinkific Basic pricing page',
            'Thinkific', 'https://www.thinkific.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page renders EUR only (€53 monthly / €39 annual). Catalog description: per month billed annually ($40). Basis recorded, price kept.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 50, 'Thinkific Basic read on 2026-09-02 at https://www.thinkific.com/pricing/: no monthly billing; unit: per account. Page renders EUR only (€53 monthly / €39 annual). Catalog description: per month billed annually ($40). Basis recorded, price kept.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        4000, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'annual', 'flat', 'per account', NULL,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: basis recorded from the page; USD figure not readable today, price kept; period=annual, unit=flat, note=per account. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 4000,
        billing_period = 'annual',
        pricing_unit = 'flat',
        pricing_unit_note = 'per account',
        annual_price_minor = NULL,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Kit Creator (kit-creator) — monthly billing available; price unchanged at 3900
DO $$
DECLARE
    product_row catalog.products%ROWTYPE;
    new_source_id uuid;
    new_observation_id uuid;
    facts_current evidence.product_fact_revisions%ROWTYPE;
    scores_current evidence.score_revisions%ROWTYPE;
    facts_new_id uuid;
    scores_new_id uuid;
    next_fact_version integer;
    next_score_version integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = 'kit-creator';
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION 'kit-creator: product is missing';
    END IF;
    IF EXISTS (
        SELECT 1 FROM evidence.product_fact_revisions
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02%'
    ) THEN
        RAISE NOTICE 'kit-creator: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('kit-creator-pricing-2026-09-02', 'manufacturer_documentation', 'Kit Creator pricing page',
            'Kit', 'https://kit.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. Page shows $33/mo as $390 billed yearly (=$32.50); $39 monthly is the catalog''s read and matches the page''s ''save $78'' arithmetic.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 80, 'Kit Creator read on 2026-09-02 at https://kit.com/pricing: 39 USD per month billed monthly; 32.5 USD per month billed annually; unit: at 1,000 subscribers. Page shows $33/mo as $390 billed yearly (=$32.50); $39 monthly is the catalog''s read and matches the page''s ''save $78'' arithmetic.')
    RETURNING id INTO new_observation_id;

    SELECT * INTO facts_current FROM evidence.product_fact_revisions
    WHERE id = product_row.published_fact_revision_id;
    SELECT * INTO scores_current FROM evidence.score_revisions
    WHERE id = product_row.published_score_revision_id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_fact_version
    FROM evidence.product_fact_revisions WHERE product_id = product_row.id;
    SELECT COALESCE(max(version), 0) + 1 INTO next_score_version
    FROM evidence.score_revisions WHERE product_id = product_row.id;

    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months,
        billing_period, pricing_unit, pricing_unit_note, annual_price_minor,
        workflow_status, submitted_at, reviewed_at, published_at, valid_until, review_note)
    VALUES (
        product_row.id, next_fact_version, product_row.category_id, product_row.brand_id,
        product_row.name, product_row.slug, product_row.description,
        3900, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_contacts', 'at 1,000 subscribers', 3250,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02: monthly billing available; period=monthly, unit=per_contacts, note=at 1,000 subscribers. Compared price unchanged.')
    RETURNING id INTO facts_new_id;

    -- Every fact except the price keeps the provenance it had; the price now
    -- rests on today's observation.
    IF facts_current.id IS NOT NULL THEN
        INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
        SELECT facts_new_id, provenance.fact_key, provenance.observation_id, provenance.public_classification
        FROM evidence.fact_provenance provenance
        WHERE provenance.fact_revision_id = facts_current.id AND provenance.fact_key <> 'price'
        ON CONFLICT DO NOTHING;
    END IF;
    INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
    VALUES (facts_new_id, 'price', new_observation_id, 'manufacturer_claim')
    ON CONFLICT DO NOTHING;

    -- Scores are unchanged: the basis correction says what the price is, not
    -- what the product is worth; affordability is scored separately by
    -- budget fit. The revision exists because a score revision points at one
    -- fact revision.
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note)
    VALUES (
        product_row.id, facts_new_id, next_score_version, product_row.quality_score,
        product_row.value_score, product_row.durability_score, product_row.beginner_score,
        product_row.advanced_score, product_row.apartment_score, product_row.noise_score,
        product_row.portability_score, 'published', now(), now(), now(),
        'Billing basis audit 2026-09-02: scores carried forward unchanged; no affiliate data considered.')
    RETURNING id INTO scores_new_id;
    IF scores_current.id IS NOT NULL THEN
        INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
        SELECT scores_new_id, rationales.score_key, rationales.rationale, rationales.observation_id
        FROM evidence.score_rationales rationales WHERE rationales.score_revision_id = scores_current.id
        ON CONFLICT DO NOTHING;
    END IF;
    IF facts_current.id IS NOT NULL THEN
        UPDATE evidence.product_fact_revisions SET workflow_status = 'superseded'
        WHERE id = facts_current.id AND workflow_status = 'published';
    END IF;
    IF scores_current.id IS NOT NULL THEN
        UPDATE evidence.score_revisions SET workflow_status = 'superseded'
        WHERE id = scores_current.id AND workflow_status = 'published';
    END IF;

    UPDATE catalog.products
    SET price_minor = 3900,
        billing_period = 'monthly',
        pricing_unit = 'per_contacts',
        pricing_unit_note = 'at 1,000 subscribers',
        annual_price_minor = 3250,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;
END $$;

-- Every published software product now states a basis.
DO $$
DECLARE missing integer;
BEGIN
    SELECT count(*) INTO missing FROM catalog.products products
    JOIN catalog.categories categories ON categories.id = products.category_id
    WHERE products.status = 'published'
      AND NOT EXISTS (
          SELECT 1 FROM evidence.product_fact_revisions facts
          WHERE facts.product_id = products.id AND facts.review_note LIKE 'Billing basis audit 2026-09-02%'
      );
    RAISE NOTICE 'published products without an audited basis: %', missing;
END $$;

COMMIT;
