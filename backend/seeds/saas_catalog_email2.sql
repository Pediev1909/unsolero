-- Zoho Campaigns and Brevo, the two that were missing from the email category.
--
-- Both were read on 2026-08-21 through a US connection. From Bulgaria they
-- serve euro and honour neither a browser locale nor an in-page currency
-- control, which is why they were left out on the 20th rather than guessed at.
--
-- BASIS, and it differs by vendor because their models do:
--   Zoho Campaigns  1,000 contacts. The page publishes only the annual rate;
--                   its billing toggle does not yield to automation, so the
--                   figure recorded is per month billed annually and the
--                   description says so.
--   Brevo           priced by EMAILS SENT, not by list size. 5,000 emails a
--                   month at monthly billing. There is no contact-count
--                   equivalent, and pretending otherwise would be the false
--                   comparison this site exists to refuse.

INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Brevo','brevo','https://www.brevo.com','FR')
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
 ('email-marketing','zoho','Zoho Campaigns Standard','zoho-campaigns-standard',
  'The cheapest here, and the free tier is the most generous of any on this page: 2,000 contacts and 6,000 emails a month, permanently. The price shown is per month billed annually at 1,000 contacts; Zoho publishes no monthly rate on this page. Segmentation stays basic until the Professional tier.',
  525,80,94,92,78,74,80),
 ('email-marketing','brevo','Brevo Starter','brevo-starter',
  'Charges for emails sent rather than for the size of your list, which is the whole reason to pick it: a large dormant list costs nothing to keep. The price shown is monthly billing for 5,000 emails a month, so it is not directly comparable with the per-subscriber pricing beside it. Includes SMS and transactional sending.',
  900,82,88,84,84,78,80)
) AS fixture(category_slug, brand_slug, name, slug, description, price_minor,
             quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories
  ON categories.slug = fixture.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = fixture.brand_slug
ON CONFLICT (slug) DO NOTHING;

INSERT INTO evidence.sources (
    external_key, source_type, title, publisher, source_url,
    is_fictional, review_status, reviewed_at, review_note
) VALUES
 ('zoho-campaigns-pricing-2026-08','manufacturer_documentation','Zoho Campaigns pricing','Zoho',
  'https://www.zoho.com/campaigns/pricing.html',false,'verified',now(),
  'Read 2026-08-21 through a US connection, with the contact slider walked until its label read 1,000. Standard tier. The page shows only the annual rate and its billing toggle could not be operated, so the figure is per month billed annually.'),
 ('brevo-pricing-2026-08','manufacturer_documentation','Brevo pricing','Brevo',
  'https://www.brevo.com/pricing/',false,'verified',now(),
  'Read 2026-08-21 through a US connection. Starter tier at monthly billing, 5,000 emails a month. Brevo prices by sends, not by contacts.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status = 'verified', reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
    source_url = EXCLUDED.source_url, review_note = EXCLUDED.review_note, updated_at = now();

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), 100, mapping.note
FROM (VALUES
 ('zoho-campaigns-standard','zoho-campaigns-pricing-2026-08',
  'Price and free-tier limits read from the vendor pricing page on 2026-08-21 at 1,000 contacts. Annual billing; no monthly rate is published.'),
 ('brevo-starter','brevo-pricing-2026-08',
  'Price read from the vendor pricing page on 2026-08-21. Monthly billing, 5,000 emails a month.')
) AS mapping(product_slug, source_key, note)
JOIN catalog.products products ON products.slug = mapping.product_slug
JOIN evidence.sources sources ON sources.external_key = mapping.source_key
WHERE NOT EXISTS (
    SELECT 1 FROM evidence.observations existing
    WHERE existing.source_id = sources.id AND existing.product_id = products.id
);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), 80, 'Suitability scores assigned by UNSOLERO editorial review.'
FROM catalog.products products
CROSS JOIN evidence.sources sources
WHERE sources.external_key = 'unsolero-editorial-saas-2026-08'
  AND products.slug IN ('zoho-campaigns-standard','brevo-starter')
  AND NOT EXISTS (
      SELECT 1 FROM evidence.observations existing
      WHERE existing.source_id = sources.id AND existing.product_id = products.id);

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, review_note)
SELECT products.id, 1, products.category_id, products.brand_id, products.name,
       products.slug, products.description, products.price_minor, products.currency,
       products.warranty_months, 'published', now(), now(), now(),
       'Read from the vendor pricing page on 2026-08-21 through a US connection. Billing basis stated per product.'
FROM catalog.products products
WHERE products.slug IN ('zoho-campaigns-standard','brevo-starter')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score,
    durability_score, beginner_score, advanced_score, apartment_score,
    noise_score, portability_score, workflow_status,
    submitted_at, reviewed_at, published_at, review_note)
SELECT products.id, facts.id, 1, products.quality_score, products.value_score,
       products.durability_score, products.beginner_score, products.advanced_score,
       products.apartment_score, products.noise_score, products.portability_score,
       'published', now(), now(), now(), 'Editorial suitability assessment; not a vendor claim.'
FROM catalog.products products
JOIN evidence.product_fact_revisions facts ON facts.product_id = products.id AND facts.version = 1
WHERE products.slug IN ('zoho-campaigns-standard','brevo-starter')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT facts.id, keys.fact_key, observations.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions facts
JOIN catalog.products products ON products.id = facts.product_id
JOIN evidence.observations observations ON observations.product_id = products.id
JOIN evidence.sources sources ON sources.id = observations.source_id
 AND sources.source_type = 'manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('description'),('price')) AS keys(fact_key)
WHERE facts.version = 1 AND products.slug IN ('zoho-campaigns-standard','brevo-starter')
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
JOIN evidence.sources sources ON sources.id = observations.source_id
 AND sources.external_key = 'unsolero-editorial-saas-2026-08'
CROSS JOIN (VALUES ('quality'),('value'),('durability'),('beginner'),('advanced'),('apartment'),('noise'),('portability')) AS keys(score_key)
WHERE scores.version = 1 AND products.slug IN ('zoho-campaigns-standard','brevo-starter')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id, published_score_revision_id = scores.id, status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores ON scores.product_id = facts.product_id AND scores.version = 1
WHERE products.id = facts.product_id
  AND products.slug IN ('zoho-campaigns-standard','brevo-starter')
  AND facts.version = 1 AND facts.workflow_status = 'published' AND scores.workflow_status = 'published';

-- Zoho Campaigns carries a live affiliate link; Brevo's programme is behind the
-- PartnerStack network profile, which was declined on 2026-08-21.
INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url,
    price_minor, shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'zoho-campaigns-standard', 'https://www.zoho.com/campaigns/pricing.html',
       525, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug = 'zoho' AND p.slug = 'zoho-campaigns-standard'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor = EXCLUDED.price_minor, last_checked_at = EXCLUDED.last_checked_at, updated_at = now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference, disclosure_label, is_active)
SELECT o.id, 'zoho', 'https://go.zoho.com/UCST', 'PE2263909', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id = o.merchant_id
WHERE m.slug = 'zoho' AND o.merchant_sku = 'zoho-campaigns-standard'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url = EXCLUDED.destination_url, is_active = EXCLUDED.is_active, updated_at = now();
