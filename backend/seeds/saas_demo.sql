-- Fictional SaaS development fixture.
--
-- Every vendor, product and price below is invented. Nothing here describes a
-- real company or a real subscription, and no commercial conclusion should be
-- drawn from it. Its only purpose is to exercise the full pipeline for the
-- non-physical vertical: catalog, evidence governance, policy activation and
-- deterministic recommendation.
--
-- Real catalog entries must be added through the admin workflow with sourced
-- evidence. The evidence gate deliberately prevents publishing a product whose
-- facts nobody has attested to, which is what stops guessed pricing reaching a
-- recommendation.
--
-- price_minor holds the entry paid tier's monthly price. Software has tiers
-- rather than one price, so the recommendation compares the cheapest tier that
-- actually delivers the capability the setup role requires.

INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Northwind Software','northwind-software','https://example.invalid/northwind','US'),
 ('Looplane','looplane','https://example.invalid/looplane','US'),
 ('Ledgerly','ledgerly','https://example.invalid/ledgerly','GB'),
 ('Beacon','beacon','https://example.invalid/beacon','US'),
 ('Deskline','deskline','https://example.invalid/deskline','DE'),
 ('Timeslot','timeslot','https://example.invalid/timeslot','NL'),
 ('Chatter','chatter','https://example.invalid/chatter','US'),
 ('Metricly','metricly','https://example.invalid/metricly','SE')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, fixture.name, fixture.slug, fixture.description,
       fixture.price_minor, 'USD', 0, fixture.quality, fixture.value, fixture.durability,
       fixture.beginner, fixture.advanced, 0, 0, fixture.portability
FROM (VALUES
 ('crm','northwind-software','Northwind CRM','saas-northwind-crm',
  'Fictional demo CRM for small client-services teams.',2900,82,74,80,78,72,70),
 ('crm','looplane','Pocket CRM','saas-pocket-crm',
  'Fictional demo entry-level CRM with a narrow feature set.',900,64,88,62,90,44,78),
 ('project-management','looplane','Loop Projects','saas-loop-projects',
  'Fictional demo project and delivery tracker.',1900,84,76,82,74,80,66),
 ('project-management','northwind-software','Taskbin Lite','saas-taskbin-lite',
  'Fictional demo lightweight task board.',700,62,86,60,92,40,74),
 ('accounting-invoicing','ledgerly','Ledgerly Books','saas-ledgerly-books',
  'Fictional demo invoicing and bookkeeping tool.',1500,80,78,84,76,70,58),
 ('email-marketing','beacon','Beacon Mail','saas-beacon-mail',
  'Fictional demo audience email and broadcast tool.',2400,78,72,76,80,68,62),
 ('help-desk','deskline','Deskline Support','saas-deskline-support',
  'Fictional demo shared inbox and ticketing tool.',2100,76,70,78,72,74,60),
 ('scheduling','timeslot','Timeslot Booking','saas-timeslot-booking',
  'Fictional demo meeting scheduling tool.',800,74,84,72,88,52,82),
 ('team-communication','chatter','Chatter Rooms','saas-chatterroom-teams',
  'Fictional demo team chat and meetings tool.',1600,80,74,80,84,66,64),
 ('analytics','metricly','Metricly Insights','saas-metricly-insights',
  'Fictional demo product analytics tool.',2600,82,68,78,64,86,56)
) AS fixture(category_slug, brand_slug, name, slug, description, price_minor,
             quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories
  ON categories.slug = fixture.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = fixture.brand_slug
ON CONFLICT (slug) DO NOTHING;

-- Governance is seeded last so fictional products are never public without
-- explicit, auditable fictional provenance.
WITH source AS (
    INSERT INTO evidence.sources (
        external_key, source_type, title, publisher, is_fictional,
        review_status, reviewed_at, review_note
    ) VALUES (
        'unsolero-saas-demo-fixture-v1', 'demo_fixture',
        'Fictional UNSOLERO SaaS development fixture', 'UNSOLERO development seed',
        true, 'verified', now(),
        'Fictional evidence for local development only; not a real-world source.'
    )
    ON CONFLICT (external_key) DO UPDATE SET
        title = EXCLUDED.title, publisher = EXCLUDED.publisher,
        is_fictional = true, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id
)
INSERT INTO evidence.observations (
    source_id, product_id, observed_at, confidence, notes
)
SELECT source.id, products.id, now(), 100,
       'Fictional demo observation; no real product claim is made.'
FROM source CROSS JOIN catalog.products AS products
WHERE products.slug LIKE 'saas-%'
  AND NOT EXISTS (
      SELECT 1 FROM evidence.observations AS existing
      WHERE existing.source_id = source.id AND existing.product_id = products.id
  );

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, review_note
)
SELECT products.id, 1, products.category_id, products.brand_id, products.name,
       products.slug, products.description, products.price_minor, products.currency,
       products.warranty_months, 'published', now(), now(), now(),
       'Fictional development fixture approved for local demonstration only.'
FROM catalog.products AS products
WHERE products.slug LIKE 'saas-%'
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
       'Fictional development score fixture; not a real-world assessment.'
FROM catalog.products AS products
JOIN evidence.product_fact_revisions AS facts
  ON facts.product_id = products.id AND facts.version = 1
WHERE products.slug LIKE 'saas-%'
ON CONFLICT (product_id, version) DO NOTHING;

-- dimensions, weight, material and max_capacity are absent by design: a
-- subscription has none, and provenance is only required for populated facts.
INSERT INTO evidence.fact_provenance (
    fact_revision_id, fact_key, observation_id, public_classification
)
SELECT facts.id, keys.fact_key, observations.id, 'editorial_assessment'
FROM evidence.product_fact_revisions AS facts
JOIN catalog.products AS products ON products.id = facts.product_id
JOIN evidence.observations AS observations ON observations.product_id = products.id
JOIN evidence.sources AS sources
  ON sources.id = observations.source_id AND sources.external_key = 'unsolero-saas-demo-fixture-v1'
CROSS JOIN (VALUES
    ('category'), ('brand'), ('name'), ('slug'), ('description'), ('price'), ('warranty')
) AS keys(fact_key)
WHERE products.slug LIKE 'saas-%' AND facts.version = 1
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (
    score_revision_id, score_key, rationale, observation_id
)
SELECT scores.id, keys.score_key,
       CASE WHEN keys.score_key IN ('apartment','noise')
            THEN 'Not applicable: a subscription has no physical footprint and makes no noise.'
            ELSE 'Fictional demo score supplied by the development fixture.' END,
       observations.id
FROM evidence.score_revisions AS scores
JOIN catalog.products AS products ON products.id = scores.product_id
JOIN evidence.observations AS observations ON observations.product_id = products.id
JOIN evidence.sources AS sources
  ON sources.id = observations.source_id AND sources.external_key = 'unsolero-saas-demo-fixture-v1'
CROSS JOIN (VALUES
    ('quality'), ('value'), ('durability'), ('beginner'),
    ('advanced'), ('apartment'), ('noise'), ('portability')
) AS keys(score_key)
WHERE products.slug LIKE 'saas-%' AND scores.version = 1
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores
  ON scores.product_id = facts.product_id AND scores.version = 1
WHERE products.id = facts.product_id
  AND products.slug LIKE 'saas-%'
  AND facts.version = 1
  AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published';

INSERT INTO recommendation.product_policies
SELECT 'saas-v1', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug LIKE 'saas-%'
  AND published_fact_revision_id IS NOT NULL AND published_score_revision_id IS NOT NULL
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

-- 'provides' is what the tool does. 'compatible_with' is what it integrates
-- with, which is the software equivalent of equipment that physically fits
-- together, and is what lets the optimizer prefer a stack whose parts connect.
INSERT INTO recommendation.product_policy_capabilities
SELECT 'saas-v1', products.id, mapping.capability, mapping.relation
FROM (VALUES
 ('saas-northwind-crm','crm','provides'),
 ('saas-northwind-crm','email_marketing','compatible_with'),
 ('saas-northwind-crm','invoicing','compatible_with'),
 ('saas-pocket-crm','crm','provides'),
 ('saas-loop-projects','project_management','provides'),
 ('saas-loop-projects','task_tracking','provides'),
 ('saas-loop-projects','crm','compatible_with'),
 ('saas-loop-projects','team_chat','compatible_with'),
 ('saas-taskbin-lite','project_management','provides'),
 ('saas-taskbin-lite','task_tracking','provides'),
 ('saas-ledgerly-books','invoicing','provides'),
 ('saas-ledgerly-books','accounting','provides'),
 ('saas-ledgerly-books','crm','compatible_with'),
 ('saas-beacon-mail','email_marketing','provides'),
 ('saas-beacon-mail','marketing_automation','provides'),
 ('saas-beacon-mail','crm','compatible_with'),
 ('saas-deskline-support','help_desk','provides'),
 ('saas-deskline-support','live_chat','provides'),
 ('saas-deskline-support','crm','compatible_with'),
 ('saas-timeslot-booking','scheduling','provides'),
 ('saas-timeslot-booking','crm','compatible_with'),
 ('saas-chatterroom-teams','team_chat','provides'),
 ('saas-chatterroom-teams','video_meetings','provides'),
 ('saas-chatterroom-teams','project_management','compatible_with'),
 ('saas-metricly-insights','product_analytics','provides')
) AS mapping(slug, capability, relation)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support
SELECT 'saas-v1', products.id, mapping.goal_key, mapping.score
FROM (VALUES
 ('saas-northwind-crm','client_services',92), ('saas-northwind-crm','solo_consulting',78),
 ('saas-pocket-crm','client_services',70), ('saas-pocket-crm','solo_consulting',88),
 ('saas-loop-projects','client_services',90), ('saas-loop-projects','software_product',86),
 ('saas-taskbin-lite','client_services',66), ('saas-taskbin-lite','software_product',62),
 ('saas-ledgerly-books','client_services',88), ('saas-ledgerly-books','solo_consulting',90),
 ('saas-beacon-mail','creator_business',92), ('saas-beacon-mail','sell_products_online',84),
 ('saas-deskline-support','sell_products_online',86), ('saas-deskline-support','software_product',88),
 ('saas-timeslot-booking','solo_consulting',90), ('saas-timeslot-booking','client_services',74),
 ('saas-chatterroom-teams','client_services',80), ('saas-chatterroom-teams','software_product',76),
 ('saas-metricly-insights','software_product',90), ('saas-metricly-insights','sell_products_online',78)
) AS mapping(slug, goal_key, score)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags
SELECT 'saas-v1', products.id, mapping.tag
FROM (VALUES
 ('saas-northwind-crm','all_in_one'), ('saas-northwind-crm','api_first'),
 ('saas-pocket-crm','no_code'), ('saas-pocket-crm','best_of_breed'),
 ('saas-loop-projects','best_of_breed'), ('saas-loop-projects','api_first'),
 ('saas-taskbin-lite','no_code'),
 ('saas-ledgerly-books','eu_hosted'), ('saas-ledgerly-books','best_of_breed'),
 ('saas-beacon-mail','all_in_one'),
 ('saas-deskline-support','eu_hosted'), ('saas-deskline-support','privacy_focused'),
 ('saas-timeslot-booking','no_code'), ('saas-timeslot-booking','eu_hosted'),
 ('saas-chatterroom-teams','all_in_one'),
 ('saas-metricly-insights','privacy_focused'), ('saas-metricly-insights','api_first')
) AS mapping(slug, tag)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups
SELECT 'saas-v1', products.id, mapping.group_key
FROM (VALUES
 ('saas-northwind-crm','crm_suite'), ('saas-pocket-crm','crm_suite'),
 ('saas-loop-projects','project_suite'), ('saas-taskbin-lite','project_suite'),
 ('saas-ledgerly-books','accounting_suite'),
 ('saas-beacon-mail','email_suite'),
 ('saas-deskline-support','support_suite'),
 ('saas-timeslot-booking','scheduling_suite'),
 ('saas-chatterroom-teams','chat_suite'),
 ('saas-metricly-insights','analytics_suite')
) AS mapping(slug, group_key)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

UPDATE recommendation.policy_versions SET workflow_status='active', activated_at=now()
WHERE version='saas-v1' AND workflow_status='draft'
AND EXISTS (SELECT 1 FROM recommendation.category_policies WHERE policy_version='saas-v1' AND support_status='supported')
AND EXISTS (SELECT 1 FROM recommendation.product_policies WHERE policy_version='saas-v1');
