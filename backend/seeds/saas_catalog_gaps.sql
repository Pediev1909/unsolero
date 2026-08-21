-- Filling the two categories that were below three products.
--
-- Accounting had Zoho Books and FreshBooks; project management had ClickUp and
-- Teamwork. A category with two entries is a choice between two things, not a
-- comparison, and it was my own rule that no category ships that way.
--
-- All five prices read on 2026-08-21 through a US connection, in USD, from the
-- vendor's own pricing page. The billing basis is stated per product because
-- they differ: two of these are free outright, one publishes only an annual rate.
--
-- The two free ones are not filler. For a freelancer invoicing a handful of
-- clients, "free, permanently" is the correct answer, and a comparison that
-- leaves it out to keep the column tidy is worse than no comparison.

INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Wave','wave','https://www.waveapps.com','CA'),
 ('monday.com','monday','https://monday.com','IL'),
 ('Notion','notion','https://www.notion.com','US')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, f.name, f.slug, f.description, f.price_minor, 'USD', 0,
       f.quality, f.value, f.durability, f.beginner, f.advanced, 0, 0, f.portability
FROM (VALUES
 ('accounting-invoicing','zoho','Zoho Invoice','zoho-invoice',
  'Free, with no paid tier at all: Zoho charges nothing for invoicing and expects you to grow into Zoho Books later. Invoicing, estimates, expenses and time tracking. It is not accounting — there is no bank reconciliation and no full ledger.',
  0,82,100,92,88,62,80),
 ('accounting-invoicing','wave','Wave Starter','wave-starter',
  'Free invoicing and accounting for one user, including double-entry bookkeeping, which the other free option here does not have. Payments and payroll are charged separately. The Pro tier is 190 USD a year and adds receipt capture and automation.',
  0,80,98,74,90,58,74),
 ('project-management','zoho','Zoho Projects Premium','zoho-projects-premium',
  'Project tracking with time logging, blueprints and dependencies, free for up to five users. The price shown is per user per month billed annually; Zoho publishes no monthly rate on this page.',
  400,82,94,92,80,78,80),
 ('project-management','monday','monday.com Basic','monday-basic',
  'Work management built around visual boards, priced per seat with a three-seat minimum. Basic is the entry paid tier and deliberately omits the timeline and calendar views most teams end up wanting. Per seat per month billed annually.',
  900,86,74,86,88,72,78),
 ('project-management','notion','Notion Plus','notion-plus',
  'Documents, databases and project tracking on one surface, which is both its strength and its cost: nothing is set up for you. Best where a team wants to design its own process. Per member per month.',
  1000,88,84,82,74,86,86)
) AS f(category_slug, brand_slug, name, slug, description, price_minor,
       quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories ON categories.slug = f.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = f.brand_slug
ON CONFLICT (slug) DO NOTHING;

INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url, is_fictional, review_status, reviewed_at, review_note) VALUES
 ('zoho-invoice-pricing-2026-08','manufacturer_documentation','Zoho Invoice pricing','Zoho','https://www.zoho.com/invoice/pricing/',false,'verified',now(),'Read 2026-08-21 through a US connection. The product is free and has no paid tier.'),
 ('wave-pricing-2026-08','manufacturer_documentation','Wave pricing','Wave','https://www.waveapps.com/pricing',false,'verified',now(),'Read 2026-08-21. Starter tier is free; Pro is 190 USD per year.'),
 ('zoho-projects-pricing-2026-08','manufacturer_documentation','Zoho Projects pricing','Zoho','https://www.zoho.com/projects/pricing.html',false,'verified',now(),'Read 2026-08-21 through a US connection. Premium tier, per user, billed annually; no monthly rate is published.'),
 ('monday-pricing-2026-08','manufacturer_documentation','monday.com pricing','monday.com','https://monday.com/pricing',false,'verified',now(),'Read 2026-08-21 through a US connection. Basic tier, per seat, billed annually. Three-seat minimum.'),
 ('notion-pricing-2026-08','manufacturer_documentation','Notion pricing','Notion','https://www.notion.com/pricing',false,'verified',now(),'Read 2026-08-21. Plus tier, per member per month.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status='verified', reviewed_at=COALESCE(evidence.sources.reviewed_at,now()),
    source_url=EXCLUDED.source_url, review_note=EXCLUDED.review_note, updated_at=now();

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), 100, 'Price read from the vendor pricing page on 2026-08-21 through a US connection.'
FROM (VALUES
 ('zoho-invoice','zoho-invoice-pricing-2026-08'),
 ('wave-starter','wave-pricing-2026-08'),
 ('zoho-projects-premium','zoho-projects-pricing-2026-08'),
 ('monday-basic','monday-pricing-2026-08'),
 ('notion-plus','notion-pricing-2026-08')
) AS m(product_slug, source_key)
JOIN catalog.products p ON p.slug = m.product_slug
JOIN evidence.sources s ON s.external_key = m.source_key
WHERE NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), 80, 'Suitability scores assigned by UNSOLERO editorial review.'
FROM catalog.products p CROSS JOIN evidence.sources s
WHERE s.external_key='unsolero-editorial-saas-2026-08'
  AND p.slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
  AND NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,1,p.category_id,p.brand_id,p.name,p.slug,p.description,p.price_minor,p.currency,p.warranty_months,
       'published',now(),now(),now(),'Read from the vendor pricing page on 2026-08-21 through a US connection.'
FROM catalog.products p
WHERE p.slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score,
    workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,f.id,1,p.quality_score,p.value_score,p.durability_score,p.beginner_score,p.advanced_score,
       p.apartment_score,p.noise_score,p.portability_score,'published',now(),now(),now(),
       'Editorial suitability assessment; not a vendor claim.'
FROM catalog.products p JOIN evidence.product_fact_revisions f ON f.product_id=p.id AND f.version=1
WHERE p.slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT f.id, k.fact_key, o.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions f
JOIN catalog.products p ON p.id=f.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.source_type='manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('description'),('price')) AS k(fact_key)
WHERE f.version=1 AND p.slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
SELECT sc.id, k.score_key,
       CASE WHEN k.score_key IN ('apartment','noise')
            THEN 'Not applicable: software has no physical footprint and makes no noise.'
            ELSE 'Editorial suitability assessment based on the published feature set, tier limits and vendor track record.' END,
       o.id
FROM evidence.score_revisions sc
JOIN catalog.products p ON p.id=sc.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.external_key='unsolero-editorial-saas-2026-08'
CROSS JOIN (VALUES ('quality'),('value'),('durability'),('beginner'),('advanced'),('apartment'),('noise'),('portability')) AS k(score_key)
WHERE sc.version=1 AND p.slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS p
SET published_fact_revision_id=f.id, published_score_revision_id=sc.id, status='published'
FROM evidence.product_fact_revisions AS f
JOIN evidence.score_revisions AS sc ON sc.product_id=f.product_id AND sc.version=1
WHERE p.id=f.product_id AND p.slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
  AND f.version=1 AND f.workflow_status='published' AND sc.workflow_status='published';

-- Zoho Invoice and Zoho Projects both carry live affiliate links.
INSERT INTO commerce.merchant_offers (merchant_id, product_id, merchant_sku, product_url, price_minor, shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, x.sku, x.url, x.price, 0, 'USD', 'in_stock', 'new', now()
FROM (VALUES
 ('zoho-invoice','zoho-invoice','https://www.zoho.com/invoice/pricing/',0),
 ('zoho-projects-premium','zoho-projects-premium','https://www.zoho.com/projects/pricing.html',400)
) AS x(product_slug, sku, url, price)
JOIN catalog.products p ON p.slug = x.product_slug
JOIN commerce.merchants m ON m.slug = 'zoho'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (merchant_offer_id, provider, destination_url, external_reference, disclosure_label, is_active)
SELECT o.id, 'zoho', x.dest, 'PE2263909', 'Affiliate link', true
FROM (VALUES
 ('zoho-invoice','https://go.zoho.com/dIhI'),
 ('zoho-projects-premium','https://go.zoho.com/PCoS')
) AS x(sku, dest)
JOIN commerce.merchant_offers o ON o.merchant_sku = x.sku
JOIN commerce.merchants m ON m.id = o.merchant_id AND m.slug = 'zoho'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url, is_active=EXCLUDED.is_active, updated_at=now();
