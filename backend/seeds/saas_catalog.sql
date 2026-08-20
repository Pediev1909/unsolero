-- Real SaaS catalog for the client_services goal.
--
-- Unlike saas_demo.sql, nothing here is invented. Every price was read from the
-- vendor's own pricing page on 2026-08-18 and is recorded with that page as its
-- source, which is what the evidence workflow requires before a product can be
-- published.
--
-- Prices are the entry PAID tier at MONTHLY billing in USD. Annual billing is
-- cheaper almost everywhere and promotional rates are excluded deliberately:
-- comparing a promotional first-year rate against a standard rate would make
-- the cheaper product look better than it is once the promotion lapses.
--
-- MAINTENANCE: software pricing changes. Each source carries the date it was
-- read. Re-check before relying on a recommendation, and update the fact
-- revision rather than editing the product in place, so the change is auditable.
--
-- Scores are editorial assessments, attributed to a separate editorial source
-- rather than to the vendor. They are judgements and are labelled as such.

-- ---------------------------------------------------------------- brands ---
INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('HubSpot','hubspot','https://www.hubspot.com','US'),
 ('Salesflare','salesflare','https://www.salesflare.com','BE'),
 ('ClickUp','clickup','https://clickup.com','US'),
 ('Teamwork.com','teamwork','https://www.teamwork.com','IE'),
 ('Zoho','zoho','https://www.zoho.com','IN'),
 ('FreshBooks','freshbooks','https://www.freshbooks.com','CA')
ON CONFLICT (slug) DO NOTHING;

-- -------------------------------------------------------------- products ---
INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, fixture.name, fixture.slug, fixture.description,
       fixture.price_minor, 'USD', 0, fixture.quality, fixture.value, fixture.durability,
       fixture.beginner, fixture.advanced, 0, 0, fixture.portability
FROM (VALUES
 ('crm','hubspot','HubSpot Starter Customer Platform','hubspot-starter-customer-platform',
  'CRM bundled with entry-level marketing, sales and service tools. A free tier covers two users, and the paid Starter tier is priced per seat.',
  2000,88,72,92,82,86,70),
 ('crm','salesflare','Salesflare Growth','salesflare-growth',
  'CRM that builds contact records automatically from email and calendar activity, aimed at small teams selling to other businesses.',
  3900,84,70,74,88,76,76),
 ('project-management','clickup','ClickUp Unlimited','clickup-unlimited',
  'Project and task tracking with unlimited storage and integrations on the entry paid tier. A free tier exists with a storage cap.',
  1000,82,90,84,74,88,78),
 ('project-management','teamwork','Teamwork Basics','teamwork-basics',
  'Project management built around client work, with billable time and client access. Free for up to five users.',
  1299,84,80,82,84,80,76),
 ('accounting-invoicing','zoho','Zoho Books Standard','zoho-books-standard',
  'Invoicing and bookkeeping with bank feeds. A free tier covers businesses under 50,000 USD annual revenue.',
  1200,82,92,88,80,78,80),
 ('accounting-invoicing','freshbooks','FreshBooks Lite','freshbooks-lite',
  'Invoicing and expense tracking aimed at service businesses. The entry tier is limited to five billable clients.',
  2300,86,74,86,90,72,74)
) AS fixture(category_slug, brand_slug, name, slug, description, price_minor,
             quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories
  ON categories.slug = fixture.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = fixture.brand_slug
ON CONFLICT (slug) DO NOTHING;

-- --------------------------------------------------------------- sources ---
-- One source per vendor pricing page. source_url is mandatory for every type
-- except demo_fixture, which is what keeps a real catalog honest: a fact
-- cannot be published without somewhere to check it.
INSERT INTO evidence.sources (
    external_key, source_type, title, publisher, source_url,
    is_fictional, review_status, reviewed_at, review_note
) VALUES
 ('hubspot-pricing-2026-08','manufacturer_documentation','HubSpot CRM pricing','HubSpot',
  'https://www.hubspot.com/pricing/crm',false,'verified',now(),'Read 2026-08-18. Starter seat price at monthly billing; promotional new-customer rate excluded.'),
 ('salesflare-pricing-2026-08','manufacturer_documentation','Salesflare pricing','Salesflare',
  'https://www.salesflare.com/pricing',false,'verified',now(),'Read 2026-08-18. Growth tier, per user, monthly billing.'),
 ('clickup-pricing-2026-08','manufacturer_documentation','ClickUp pricing','ClickUp',
  'https://clickup.com/pricing',false,'verified',now(),'Read 2026-08-18. Unlimited tier, per user, monthly billing.'),
 ('teamwork-pricing-2026-08','manufacturer_documentation','Teamwork.com pricing','Teamwork.com',
  'https://www.teamwork.com/pricing/',false,'verified',now(),'Read 2026-08-18. Basics tier, per user, monthly billing.'),
 ('zoho-books-pricing-2026-08','manufacturer_documentation','Zoho Books pricing','Zoho',
  'https://www.zoho.com/books/pricing/',false,'verified',now(),'Read 2026-08-18. Standard tier, monthly billing.'),
 ('freshbooks-pricing-2026-08','manufacturer_documentation','FreshBooks pricing','FreshBooks',
  'https://www.freshbooks.com/pricing',false,'verified',now(),'Read 2026-08-18. Lite tier, monthly billing; introductory discount excluded.'),
 ('unsolero-editorial-saas-2026-08','editorial_assessment','UNSOLERO suitability assessment','Andon Pediev',
  'https://unsolero.com/articles/how-unsolero-ranks-software',false,'verified',now(),
  'Suitability scores are editorial judgements, not vendor claims, and are attributed separately from pricing facts.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status = 'verified', reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
    source_url = EXCLUDED.source_url, review_note = EXCLUDED.review_note, updated_at = now();

-- ---------------------------------------------------------- observations ---
INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), 100,
       'Pricing and product facts read from the vendor pricing page on 2026-08-18.'
FROM (VALUES
 ('hubspot-starter-customer-platform','hubspot-pricing-2026-08'),
 ('salesflare-growth','salesflare-pricing-2026-08'),
 ('clickup-unlimited','clickup-pricing-2026-08'),
 ('teamwork-basics','teamwork-pricing-2026-08'),
 ('zoho-books-standard','zoho-books-pricing-2026-08'),
 ('freshbooks-lite','freshbooks-pricing-2026-08')
) AS mapping(product_slug, source_key)
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
  AND products.slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
                        'teamwork-basics','zoho-books-standard','freshbooks-lite')
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
       'Entry paid tier at monthly billing, read from the vendor pricing page on 2026-08-18.'
FROM catalog.products products
WHERE products.slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
                        'teamwork-basics','zoho-books-standard','freshbooks-lite')
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
WHERE products.slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
                        'teamwork-basics','zoho-books-standard','freshbooks-lite')
ON CONFLICT (product_id, version) DO NOTHING;

-- Facts are attributed to the vendor page; scores to the editorial source.
INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT facts.id, keys.fact_key, observations.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions facts
JOIN catalog.products products ON products.id = facts.product_id
JOIN evidence.observations observations ON observations.product_id = products.id
JOIN evidence.sources sources
  ON sources.id = observations.source_id AND sources.source_type = 'manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('slug'),('description'),('price'),('warranty')) AS keys(fact_key)
WHERE facts.version = 1
  AND products.slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
                        'teamwork-basics','zoho-books-standard','freshbooks-lite')
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
  AND products.slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
                        'teamwork-basics','zoho-books-standard','freshbooks-lite')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores
  ON scores.product_id = facts.product_id AND scores.version = 1
WHERE products.id = facts.product_id
  AND products.slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
                        'teamwork-basics','zoho-books-standard','freshbooks-lite')
  AND facts.version = 1
  AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published';

-- -------------------------------------------------------- policy binding ---
INSERT INTO recommendation.product_policies
SELECT 'saas-v1', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('hubspot-starter-customer-platform','salesflare-growth','clickup-unlimited',
               'teamwork-basics','zoho-books-standard','freshbooks-lite')
  AND published_fact_revision_id IS NOT NULL
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

-- 'compatible_with' records the integrations each tool advertises with the
-- other categories in this catalog. It is what lets the optimizer prefer a
-- stack whose parts connect over one that merely fits the budget.
INSERT INTO recommendation.product_policy_capabilities
SELECT 'saas-v1', products.id, mapping.capability, mapping.relation
FROM (VALUES
 ('hubspot-starter-customer-platform','crm','provides'),
 ('hubspot-starter-customer-platform','email_marketing','provides'),
 ('hubspot-starter-customer-platform','project_management','compatible_with'),
 ('hubspot-starter-customer-platform','invoicing','compatible_with'),
 ('salesflare-growth','crm','provides'),
 ('salesflare-growth','email_marketing','compatible_with'),
 ('salesflare-growth','invoicing','compatible_with'),
 ('clickup-unlimited','project_management','provides'),
 ('clickup-unlimited','task_tracking','provides'),
 ('clickup-unlimited','crm','compatible_with'),
 ('clickup-unlimited','team_chat','compatible_with'),
 ('teamwork-basics','project_management','provides'),
 ('teamwork-basics','task_tracking','provides'),
 ('teamwork-basics','time_tracking','provides'),
 ('teamwork-basics','crm','compatible_with'),
 ('teamwork-basics','invoicing','compatible_with'),
 ('zoho-books-standard','invoicing','provides'),
 ('zoho-books-standard','accounting','provides'),
 ('zoho-books-standard','crm','compatible_with'),
 ('freshbooks-lite','invoicing','provides'),
 ('freshbooks-lite','accounting','provides'),
 ('freshbooks-lite','time_tracking','provides'),
 ('freshbooks-lite','crm','compatible_with')
) AS mapping(slug, capability, relation)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support
SELECT 'saas-v1', products.id, mapping.goal_key, mapping.score
FROM (VALUES
 ('hubspot-starter-customer-platform','client_services',88),
 ('hubspot-starter-customer-platform','sell_products_online',78),
 ('hubspot-starter-customer-platform','creator_business',74),
 ('hubspot-starter-customer-platform','solo_consulting',72),
 ('salesflare-growth','client_services',90),
 ('salesflare-growth','solo_consulting',84),
 ('clickup-unlimited','client_services',88),
 ('clickup-unlimited','software_product',90),
 ('clickup-unlimited','solo_consulting',76),
 ('teamwork-basics','client_services',92),
 ('teamwork-basics','software_product',80),
 ('zoho-books-standard','client_services',86),
 ('zoho-books-standard','solo_consulting',92),
 ('zoho-books-standard','sell_products_online',74),
 ('freshbooks-lite','client_services',84),
 ('freshbooks-lite','solo_consulting',94)
) AS mapping(slug, goal_key, score)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags
SELECT 'saas-v1', products.id, mapping.tag
FROM (VALUES
 ('hubspot-starter-customer-platform','all_in_one'),
 ('hubspot-starter-customer-platform','api_first'),
 ('salesflare-growth','best_of_breed'),
 ('salesflare-growth','no_code'),
 ('clickup-unlimited','best_of_breed'),
 ('clickup-unlimited','api_first'),
 ('teamwork-basics','best_of_breed'),
 ('zoho-books-standard','all_in_one'),
 ('zoho-books-standard','no_code'),
 ('freshbooks-lite','no_code')
) AS mapping(slug, tag)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups
SELECT 'saas-v1', products.id, mapping.group_key
FROM (VALUES
 ('hubspot-starter-customer-platform','crm_suite'),
 ('salesflare-growth','crm_suite'),
 ('clickup-unlimited','project_suite'),
 ('teamwork-basics','project_suite'),
 ('zoho-books-standard','accounting_suite'),
 ('freshbooks-lite','accounting_suite')
) AS mapping(slug, group_key)
JOIN catalog.products products ON products.slug = mapping.slug
WHERE EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version='saas-v1' AND workflow_status='draft')
ON CONFLICT DO NOTHING;

UPDATE recommendation.policy_versions SET workflow_status='active', activated_at=now()
WHERE version='saas-v1' AND workflow_status='draft'
AND EXISTS (SELECT 1 FROM recommendation.category_policies WHERE policy_version='saas-v1' AND support_status='supported')
AND EXISTS (SELECT 1 FROM recommendation.product_policies WHERE policy_version='saas-v1');
