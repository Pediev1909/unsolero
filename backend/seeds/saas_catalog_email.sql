-- Email marketing category.
--
-- Three products, not five. Zoho Campaigns and Brevo are deliberately absent:
-- both serve their pricing in euro from this location and ignore a browser
-- locale and their own currency control, so their USD figure could not be read
-- at all. A euro number published as a dollar one is a wrong number with a
-- source beside it. See docs/email-marketing-pricing-research.md.
--
-- BASIS. Email tools price by list size, so a bare price means nothing. All
-- three below are quoted at **1,000 subscribers or contacts**, which is where a
-- small business actually sits, and the billing period is stated per product
-- because one vendor publishes no monthly rate at all.
--
-- CONFIDENCE.
--   MailerLite      read from the vendor page, slider set to 1,000    100
--   ActiveCampaign  read from the vendor page; annual rate only        90
--   Kit             secondary sources; vendor page did not yield       75

-- ---------------------------------------------------------------- brands ---
INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('MailerLite','mailerlite','https://www.mailerlite.com','LT'),
 ('ActiveCampaign','activecampaign','https://www.activecampaign.com','US'),
 ('Kit','kit','https://kit.com','US')
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
 ('email-marketing','mailerlite','MailerLite Comfort','mailerlite-comfort',
  'Email marketing with a clean editor and a free tier covering 250 subscribers and 2,500 emails a month. The price shown is monthly billing at 1,000 subscribers and rises with list size.',
  1900,86,88,82,92,70,82),
 ('email-marketing','activecampaign','ActiveCampaign Starter','activecampaign-starter',
  'Automation-led email marketing, the deepest here by a distance. No free tier, and the entry plan caps automations at five actions. The price shown is 1,000 contacts billed annually; ActiveCampaign publishes no monthly rate. Above 2,500 contacts the entry plan more than doubles.',
  1500,88,70,84,66,94,78),
 ('email-marketing','kit','Kit Creator','kit-creator',
  'Email marketing built for people who publish, formerly ConvertKit. A free tier reaches 10,000 subscribers with features withheld. The price shown is monthly billing at 1,000 subscribers.',
  3900,86,62,80,84,76,84)
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
 ('mailerlite-pricing-2026-08','manufacturer_documentation','MailerLite pricing','MailerLite',
  'https://www.mailerlite.com/pricing',false,'verified',now(),
  'Read 2026-08-20 directly from the vendor page, with the subscriber slider walked until its label read 1,000 and the billing toggle set to monthly. Comfort tier.'),
 ('activecampaign-pricing-2026-08','manufacturer_documentation','ActiveCampaign pricing','ActiveCampaign',
  'https://www.activecampaign.com/pricing',false,'verified',now(),
  'Read 2026-08-20 from the vendor page with the contact selector set to 500-1,000. Starter tier. The page publishes only the annual rate; no monthly rate is shown anywhere, so the figure recorded is the annual one and is labelled as such.'),
 ('kit-pricing-2026-08','manufacturer_documentation','Kit pricing','Kit',
  'https://kit.com/pricing',false,'verified',now(),
  'Read 2026-08-20. The vendor page did not expose its plan prices to automated reading, so the figure is corroborated from secondary sources and recorded at reduced confidence pending manual confirmation.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status = 'verified', reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
    source_url = EXCLUDED.source_url, review_note = EXCLUDED.review_note, updated_at = now();

-- ---------------------------------------------------------- observations ---
INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), mapping.confidence, mapping.note
FROM (VALUES
 ('mailerlite-comfort','mailerlite-pricing-2026-08',100,
  'Price read from the vendor pricing page on 2026-08-20 at 1,000 subscribers, monthly billing.'),
 ('activecampaign-starter','activecampaign-pricing-2026-08',90,
  'Price read from the vendor pricing page on 2026-08-20 at 1,000 contacts. Annual billing; the vendor publishes no monthly rate.'),
 ('kit-creator','kit-pricing-2026-08',75,
  'Price corroborated from secondary sources on 2026-08-20. The vendor page did not yield its plan prices to reading; pending manual confirmation.')
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
  AND products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
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
       'Entry paid tier at 1,000 subscribers, read from the vendor pricing page on 2026-08-20. Billing period stated per product.'
FROM catalog.products products
WHERE products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
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
WHERE products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
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
  AND products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
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
  AND products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id,
    status = 'published'
FROM evidence.product_fact_revisions AS facts
JOIN evidence.score_revisions AS scores
  ON scores.product_id = facts.product_id AND scores.version = 1
WHERE products.id = facts.product_id
  AND products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
  AND facts.version = 1
  AND facts.workflow_status = 'published'
  AND scores.workflow_status = 'published';
