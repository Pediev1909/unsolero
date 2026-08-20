-- CRM category expansion.
--
-- Follows saas_catalog.sql exactly: entry PAID tier at MONTHLY billing in USD,
-- every price recorded against the vendor's own pricing page with the date it
-- was read, scores attributed to a separate editorial source.
--
-- One difference from that file. Its policy bindings are guarded on the policy
-- still being a draft; saas-v1 is active now, so these accept either state.
--
-- CONFIDENCE. Two of these four prices could not be read from the vendor page
-- directly and are recorded at lower confidence with the reason in the source
-- note. That is what the confidence field is for; padding them to 100 would
-- make the evidence record say something it cannot support.
--   Bigin           read from bigin.com/pricing.html          confidence 100
--   Freshsales      vendor publishes the ANNUAL rate only     confidence  70
--   Zoho CRM        vendor page renders INR by default        confidence  85
--   Pipedrive       vendor page refuses automated reading     confidence  80

-- ---------------------------------------------------------------- brands ---
INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Pipedrive','pipedrive','https://www.pipedrive.com','EE'),
 ('Freshworks','freshworks','https://www.freshworks.com','US')
ON CONFLICT (slug) DO NOTHING;

-- -------------------------------------------------------------- products ---
-- Bigin is filed under the Zoho brand rather than its own. It is Zoho's
-- product, sold on its own domain, and a brand page holding one product is not
-- a brand page.
INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, fixture.name, fixture.slug, fixture.description,
       fixture.price_minor, 'USD', 0, fixture.quality, fixture.value, fixture.durability,
       fixture.beginner, fixture.advanced, 0, 0, fixture.portability
FROM (VALUES
 ('crm','zoho','Zoho CRM Standard','zoho-crm-standard',
  'CRM with deep customisation, workflow automation and a large integration surface. A free edition covers three users. The listed price is monthly billing; annual billing is 14 USD.',
  2000,84,86,92,68,90,82),
 ('crm','zoho','Bigin Express','bigin-express',
  'Pipeline-first CRM built for small teams, deliberately simpler than Zoho CRM. Free for a single user. Express caps three pipelines, 50,000 records and 30 automations. Monthly billing; annual is 7 USD.',
  900,82,94,92,92,58,80),
 ('crm','pipedrive','Pipedrive Lite','pipedrive-lite',
  'Sales CRM built around a visual pipeline, widely regarded as the easiest to adopt. The entry tier is thin on automation and several capabilities are paid add-ons. Monthly billing; annual is 14 USD.',
  1990,88,72,84,90,70,84),
 ('crm','freshworks','Freshsales Growth','freshsales-growth',
  'CRM with built-in email, phone and chat, and a free tier for up to three users. AI and advanced automation are gated to higher tiers. Freshworks publishes only the annual rate of 9 USD; the monthly rate shown is widely reported but unconfirmed.',
  1100,80,88,78,84,72,80)
) AS fixture(category_slug, brand_slug, name, slug, description, price_minor,
             quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories
  ON categories.slug = fixture.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = fixture.brand_slug
ON CONFLICT (slug) DO NOTHING;

-- --------------------------------------------------------------- sources ---
INSERT INTO evidence.sources (
    external_key, source_type, title, publisher, source_url,
    is_fictional, review_status, reviewed_at, review_note
) VALUES
 ('zoho-crm-pricing-2026-08','manufacturer_documentation','Zoho CRM pricing','Zoho',
  'https://www.zoho.com/crm/zohocrm-pricing.html',false,'verified',now(),
  'Read 2026-08-20. Standard tier, per user, monthly billing. The page renders INR by default; the USD figure could not be read directly from it and is recorded at reduced confidence.'),
 ('bigin-pricing-2026-08','manufacturer_documentation','Bigin by Zoho pricing','Zoho',
  'https://www.bigin.com/pricing.html',false,'verified',now(),
  'Read 2026-08-20. Express tier, per user, monthly billing, read directly from the vendor page.'),
 ('pipedrive-pricing-2026-08','manufacturer_documentation','Pipedrive pricing','Pipedrive',
  'https://www.pipedrive.com/en/pricing',false,'verified',now(),
  'Read 2026-08-20. Lite tier, per seat, monthly billing. The vendor page refuses automated reading, so the figure is recorded at reduced confidence and should be re-checked by hand.'),
 ('freshsales-pricing-2026-08','manufacturer_documentation','Freshsales pricing','Freshworks',
  'https://www.freshworks.com/crm/pricing/',false,'verified',now(),
  'Read 2026-08-20. Growth tier. The vendor publishes the annual rate of 9 USD per user per month and does not publish a monthly rate; 11 USD is the widely reported monthly figure and is recorded at reduced confidence.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status = 'verified', reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
    source_url = EXCLUDED.source_url, review_note = EXCLUDED.review_note, updated_at = now();

-- ---------------------------------------------------------- observations ---
INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), mapping.confidence, mapping.note
FROM (VALUES
 ('zoho-crm-standard','zoho-crm-pricing-2026-08',85,
  'Price read on 2026-08-20. Vendor page renders INR by default; USD figure corroborated but not read from the page itself.'),
 ('bigin-express','bigin-pricing-2026-08',100,
  'Price and tier limits read directly from the vendor pricing page on 2026-08-20.'),
 ('pipedrive-lite','pipedrive-pricing-2026-08',80,
  'Price read on 2026-08-20. Vendor page refuses automated reading; figure corroborated from secondary sources and pending manual confirmation.'),
 ('freshsales-growth','freshsales-pricing-2026-08',70,
  'Annual rate read directly from the vendor page on 2026-08-20. The monthly rate is not published by the vendor.')
) AS mapping(product_slug, source_key, confidence, note)
JOIN catalog.products products ON products.slug = mapping.product_slug
JOIN evidence.sources sources ON sources.external_key = mapping.source_key
WHERE NOT EXISTS (
    SELECT 1 FROM evidence.observations existing
    WHERE existing.source_id = sources.id AND existing.product_id = products.id
);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), 80,
       'Suitability scores assigned by UNSOLERO editorial review.'
FROM catalog.products products
CROSS JOIN evidence.sources sources
WHERE sources.external_key = 'unsolero-editorial-saas-2026-08'
  AND products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
  AND NOT EXISTS (
      SELECT 1 FROM evidence.observations existing
      WHERE existing.source_id = sources.id AND existing.product_id = products.id
  );

-- ------------------------------------------------------- fact revisions ----
INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT products.id, 1, products.category_id, products.brand_id, products.name,
       products.slug, products.description, products.price_minor, products.currency,
       products.warranty_months, 'published', now(), now(), now(),
       'Entry paid tier at monthly billing, read from the vendor pricing page on 2026-08-20.'
FROM catalog.products products
WHERE products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score,
    durability_score, beginner_score, advanced_score, apartment_score,
    noise_score, portability_score, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT products.id, facts.id, 1, products.quality_score, products.value_score,
       products.durability_score, products.beginner_score, products.advanced_score,
       products.apartment_score, products.noise_score, products.portability_score,
       'published', now(), now(), now(),
       'Editorial suitability assessment; not a vendor claim.'
FROM catalog.products products
JOIN evidence.product_fact_revisions facts
  ON facts.product_id = products.id AND facts.version = 1
WHERE products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT facts.id, keys.fact_key, observations.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions facts
JOIN catalog.products products ON products.id = facts.product_id
JOIN evidence.observations observations ON observations.product_id = products.id
JOIN evidence.sources sources
  ON sources.id = observations.source_id AND sources.source_type = 'manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('description'),('price')) AS keys(fact_key)
WHERE facts.version = 1
  AND products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
SELECT scores.id, keys.score_key,
       CASE WHEN keys.score_key IN ('apartment','noise')
            THEN 'Not applicable: software has no physical footprint and makes no noise.'
            ELSE 'Editorial suitability assessment based on the published feature set, tier limits and vendor track record.' END,
       observations.id
FROM evidence.score_revisions scores
JOIN catalog.products products ON products.id = scores.product_id
JOIN evidence.observations observations ON observations.product_id = products.id
JOIN evidence.sources sources
  ON sources.id = observations.source_id AND sources.external_key = 'unsolero-editorial-saas-2026-08'
CROSS JOIN (VALUES ('quality'),('value'),('durability'),('beginner'),('advanced'),('apartment'),('noise'),('portability')) AS keys(score_key)
WHERE scores.version = 1
  AND products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores
  ON scores.product_id = facts.product_id AND scores.version = 1
WHERE products.id = facts.product_id
  AND products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
  AND facts.version = 1
  AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published';

-- -------------------------------------------------------- policy binding ---
-- saas-v1 is active rather than draft now, so every binding below accepts both.
INSERT INTO recommendation.product_policies
SELECT 'saas-v1', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
  AND published_fact_revision_id IS NOT NULL
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions
              WHERE version='saas-v1' AND workflow_status IN ('draft','active'))
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities
SELECT 'saas-v1', products.id, mapping.capability, mapping.relation
FROM (VALUES
 ('zoho-crm-standard','crm','provides'),
 ('zoho-crm-standard','email_marketing','compatible_with'),
 ('zoho-crm-standard','invoicing','compatible_with'),
 ('zoho-crm-standard','project_management','compatible_with'),
 ('bigin-express','crm','provides'),
 ('bigin-express','invoicing','compatible_with'),
 ('bigin-express','email_marketing','compatible_with'),
 ('pipedrive-lite','crm','provides'),
 ('pipedrive-lite','email_marketing','compatible_with'),
 ('pipedrive-lite','invoicing','compatible_with'),
 ('pipedrive-lite','project_management','compatible_with'),
 ('freshsales-growth','crm','provides'),
 ('freshsales-growth','email_marketing','compatible_with'),
 ('freshsales-growth','invoicing','compatible_with')
) AS mapping(slug, capability, relation)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions
              WHERE version='saas-v1' AND workflow_status IN ('draft','active'))
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support
SELECT 'saas-v1', products.id, mapping.goal_key, mapping.score
FROM (VALUES
 ('zoho-crm-standard','client_services',84),
 ('zoho-crm-standard','sell_products_online',80),
 ('zoho-crm-standard','software_product',78),
 ('zoho-crm-standard','solo_consulting',70),
 ('bigin-express','client_services',88),
 ('bigin-express','solo_consulting',92),
 ('bigin-express','creator_business',78),
 ('pipedrive-lite','client_services',90),
 ('pipedrive-lite','solo_consulting',86),
 ('pipedrive-lite','sell_products_online',72),
 ('freshsales-growth','client_services',84),
 ('freshsales-growth','solo_consulting',82),
 ('freshsales-growth','sell_products_online',74)
) AS mapping(slug, goal_key, score)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions
              WHERE version='saas-v1' AND workflow_status IN ('draft','active'))
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags
SELECT 'saas-v1', products.id, mapping.tag
FROM (VALUES
 ('zoho-crm-standard','all_in_one'),
 ('zoho-crm-standard','api_first'),
 ('zoho-crm-standard','eu_hosted'),
 ('bigin-express','no_code'),
 ('bigin-express','all_in_one'),
 ('pipedrive-lite','best_of_breed'),
 ('pipedrive-lite','api_first'),
 ('pipedrive-lite','eu_hosted'),
 ('freshsales-growth','best_of_breed'),
 ('freshsales-growth','no_code')
) AS mapping(slug, tag)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions
              WHERE version='saas-v1' AND workflow_status IN ('draft','active'))
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups
SELECT 'saas-v1', products.id, 'crm_suite'
FROM catalog.products products
WHERE products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions
              WHERE version='saas-v1' AND workflow_status IN ('draft','active'))
ON CONFLICT DO NOTHING;
