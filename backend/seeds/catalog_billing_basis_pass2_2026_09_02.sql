-- Billing basis audit, 2026-09-02 — pass 2.
--
-- A follow-up to the first file of the same date. Its products were the ones
-- whose vendor page rendered EUR or INR from Bulgaria, so pass one recorded
-- their basis honestly but could not move their price. Their USD figures were
-- then found inside the pages themselves — structured offer markup, data
-- attributes, or the page's own multi-currency payload — which is what each
-- observation below cites.
-- This file records, per product, what basis its compared price is on (billing_period, pricing_unit, pricing_unit_note, annual_price_minor
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
-- Price moves (4):
--   notion-plus: 1000 -> 1200 (monthly billing available)
--   sketch-standard: 1200 -> 1400 (monthly billing available)
--   squarespace-basic: 1900 -> 2500 (monthly billing available)
--   monday-basic: 900 -> 1200 (monthly billing available)
--
-- Applied with psql:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f backend/seeds/catalog_billing_basis_pass2_2026_09_02.sql

BEGIN;

-- Notion Plus (notion-plus) — monthly billing available; price 1000 -> 1200
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
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02 pass 2%'
    ) THEN
        RAISE NOTICE 'notion-plus: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('notion-plus-pricing-2026-09-02', 'manufacturer_documentation', 'Notion Plus pricing page',
            'Notion', 'https://www.notion.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. USD read from the page''s own multi-currency plan payload (the visible page renders EUR from Bulgaria): plus USD month 1200, year 12000 minor units, i.e. $12 monthly and $10/month billed annually.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Notion Plus read on 2026-09-02 at https://www.notion.com/pricing: 12 USD per month billed monthly; 10 USD per month billed annually; unit: per member. USD read from the page''s own multi-currency plan payload (the visible page renders EUR from Bulgaria): plus USD month 1200, year 12000 minor units, i.e. $12 monthly and $10/month billed annually.')
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
        'monthly', 'per_user', 'per member', 1000,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02 pass 2: monthly billing available; period=monthly, unit=per_user, note=per member. Compared price moved from 1000 to 1200 minor units (monthly-billing list price rule).')
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
        'Billing basis audit 2026-09-02 pass 2: scores carried forward unchanged; no affiliate data considered.')
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
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per member',
        annual_price_minor = 1000,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 1200, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Sketch Standard (sketch-standard) — monthly billing available; price 1200 -> 1400
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
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02 pass 2%'
    ) THEN
        RAISE NOTICE 'sketch-standard: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('sketch-standard-pricing-2026-09-02', 'manufacturer_documentation', 'Sketch Standard pricing page',
            'Sketch', 'https://www.sketch.com/pricing/', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. USD read from the page''s own data attributes (data-price-monthly-usd=14, data-price-yearly-usd=12) while the visible page rendered EUR.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Sketch Standard read on 2026-09-02 at https://www.sketch.com/pricing/: 14 USD per month billed monthly; 12 USD per month billed annually; unit: per editor; viewers free. USD read from the page''s own data attributes (data-price-monthly-usd=14, data-price-yearly-usd=12) while the visible page rendered EUR.')
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
        1400, product_row.currency, product_row.length_mm, product_row.width_mm,
        product_row.height_mm, product_row.weight_grams, product_row.max_capacity_grams,
        product_row.material, product_row.warranty_months,
        'monthly', 'per_user', 'per editor; viewers free', 1200,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02 pass 2: monthly billing available; period=monthly, unit=per_user, note=per editor; viewers free. Compared price moved from 1200 to 1400 minor units (monthly-billing list price rule).')
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
        'Billing basis audit 2026-09-02 pass 2: scores carried forward unchanged; no affiliate data considered.')
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
    SET price_minor = 1400,
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per editor; viewers free',
        annual_price_minor = 1200,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 1400, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- Squarespace Basic (squarespace-basic) — monthly billing available; price 1900 -> 2500
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
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02 pass 2%'
    ) THEN
        RAISE NOTICE 'squarespace-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('squarespace-basic-pricing-2026-09-02', 'manufacturer_documentation', 'Squarespace Basic pricing page',
            'Squarespace', 'https://www.squarespace.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. USD read from the page''s Offer markup: Basic monthly $25.00, Basic annually $228.00 a year ($19/month). The visible page rendered EUR.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'Squarespace Basic read on 2026-09-02 at https://www.squarespace.com/pricing: 25 USD per month billed monthly; 19 USD per month billed annually; unit: per site. USD read from the page''s Offer markup: Basic monthly $25.00, Basic annually $228.00 a year ($19/month). The visible page rendered EUR.')
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
        'monthly', 'flat', 'per site', 1900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02 pass 2: monthly billing available; period=monthly, unit=flat, note=per site. Compared price moved from 1900 to 2500 minor units (monthly-billing list price rule).')
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
        'Billing basis audit 2026-09-02 pass 2: scores carried forward unchanged; no affiliate data considered.')
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
        pricing_unit = 'flat',
        pricing_unit_note = 'per site',
        annual_price_minor = 1900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 2500, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
END $$;

-- monday.com Basic (monday-basic) — monthly billing available; price 900 -> 1200
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
        WHERE product_id = product_row.id AND review_note LIKE 'Billing basis audit 2026-09-02 pass 2%'
    ) THEN
        RAISE NOTICE 'monday-basic: already audited, skipping';
        RETURN;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES ('monday-basic-pricing-2026-09-02', 'manufacturer_documentation', 'monday.com Basic pricing page',
            'monday.com', 'https://monday.com/pricing', false, 'verified', now(),
            'Read 2026-09-02 for the billing basis audit. USD read from the page''s Work Management offer markup: Basic $12 per seat billed monthly, $9 billed annually. The visible page rendered EUR and a promotional discount; the markup carries the list price.')
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO new_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    VALUES (new_source_id, product_row.id, now(), now() + interval '90 days', 100, 'monday.com Basic read on 2026-09-02 at https://monday.com/pricing: 12 USD per month billed monthly; 9 USD per month billed annually; unit: per seat. USD read from the page''s Work Management offer markup: Basic $12 per seat billed monthly, $9 billed annually. The visible page rendered EUR and a promotional discount; the markup carries the list price.')
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
        'monthly', 'per_user', 'per seat', 900,
        'published', now(), now(), now(), now() + interval '90 days', 'Billing basis audit 2026-09-02 pass 2: monthly billing available; period=monthly, unit=per_user, note=per seat. Compared price moved from 900 to 1200 minor units (monthly-billing list price rule).')
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
        'Billing basis audit 2026-09-02 pass 2: scores carried forward unchanged; no affiliate data considered.')
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
        billing_period = 'monthly',
        pricing_unit = 'per_user',
        pricing_unit_note = 'per seat',
        annual_price_minor = 900,
        published_fact_revision_id = facts_new_id,
        published_score_revision_id = scores_new_id,
        updated_at = now()
    WHERE id = product_row.id;

    -- The vendor button must quote the same figure the page compares.
    UPDATE commerce.merchant_offers
    SET price_minor = 1200, last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
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
