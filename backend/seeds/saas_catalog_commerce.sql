-- The four categories the recommendation wizard needed to stop saying "no".
--
-- Two of the five visitor profiles -- "sell products online" and "creator
-- business" -- returned no_suitable_products, and not because the engine was
-- wrong. Each of those goals declares required roles, and nothing in the
-- catalog provided website_builder, course_hosting, online_store or payments.
-- A required role no product can fill makes the whole goal unanswerable.
--
-- All twelve prices read on 2026-08-21 through a US connection, in USD, from
-- the vendor's own pricing page. The billing basis is stated per product,
-- because these four categories charge in three different shapes: a monthly
-- subscription, a subscription billed annually, and nothing at all up front
-- with a percentage taken from each sale.
--
-- Four products cost nothing per month and take a cut instead. Their price is
-- recorded as zero, which is the truth, and the percentage lives in the
-- description where it decides the choice. For anyone selling seriously, 10%
-- of revenue dwarfs a $40 subscription -- so "free" here is the beginning of
-- the comparison, not the end of it.
--
-- Zoho is absent from all four despite holding a live referral link for each
-- (Sites D7ye, Commerce ooMj, Checkout mDpl, Learn yEbW). Its pricing pages
-- render every price into an empty span marked nosnippet and fill it from a
-- script that does not run for automated reading. Same wall as Zoho Campaigns
-- and Zoho Desk. A referral link is not a reason to publish a price nobody
-- read.

INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Framer','framer','https://www.framer.com','NL'),
 ('Webflow','webflow','https://webflow.com','US'),
 ('Squarespace','squarespace','https://www.squarespace.com','US'),
 ('Ecwid','ecwid','https://www.ecwid.com','US'),
 ('Shopify','shopify','https://www.shopify.com','CA'),
 ('BigCommerce','bigcommerce','https://www.bigcommerce.com','US'),
 ('Stripe','stripe','https://stripe.com','US'),
 ('Paddle','paddle','https://www.paddle.com','GB'),
 ('Lemon Squeezy','lemon-squeezy','https://www.lemonsqueezy.com','US'),
 ('Gumroad','gumroad','https://gumroad.com','US'),
 ('Teachable','teachable','https://teachable.com','US'),
 ('Thinkific','thinkific','https://www.thinkific.com','CA')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, f.name, f.slug, f.description, f.price_minor, 'USD', 0,
       f.quality, f.value, f.durability, f.beginner, f.advanced, 0, 0, f.portability
FROM (VALUES
 -- ---- website builder -------------------------------------------------
 ('website-builder','framer','Framer Basic','framer-basic',
  'A design tool that publishes the site directly, so what you arrange is what ships. Per site per month billed yearly, with a free tier on a Framer domain. Choose it when the look matters more than the content model.',
  1000,88,90,74,78,84,66),
 ('website-builder','webflow','Webflow Basic','webflow-basic',
  'Visual development with real CSS underneath and a CMS above it. The figure is the Basic site plan per month billed yearly, and Basic deliberately excludes the CMS -- if you need collections, the next tier is the real entry price.',
  1500,90,82,84,66,94,78),
 ('website-builder','squarespace','Squarespace Basic','squarespace-basic',
  'The one that asks least of you: templates, hosting, domain and payments in a single monthly bill, quoted here at monthly billing. Selling through it costs 2% on top on this tier, which the next tier removes.',
  1900,86,78,90,94,62,58),
 -- ---- ecommerce -------------------------------------------------------
 ('ecommerce-platform','ecwid','Ecwid Starter','ecwid-starter',
  'A store you drop into a site you already have, rather than a site built around a store. Billed annually. The Starter tier stops at 10 products, and that cap, not the price, is what decides whether it fits.',
  500,76,92,82,88,62,72),
 ('ecommerce-platform','shopify','Shopify Basic','shopify-basic',
  'The default answer, and defensibly so: the largest app ecosystem and the fewest unpleasant surprises at scale. Card rates start at 2.9% + 30c. The figure was read with the page''s billing control set to pay yearly.',
  2900,92,78,94,88,88,74),
 ('ecommerce-platform','bigcommerce','BigCommerce Core','bigcommerce-core',
  'Per month billed annually, with no platform transaction fee on top when you use a supported payment provider. Core is capped at 30,000 USD of trailing annual sales and upgrades itself past that, which is a cost worth planning for.',
  2900,86,76,84,78,86,78),
 -- ---- payments --------------------------------------------------------
 ('payments','stripe','Stripe','stripe',
  'No monthly fee; 2.9% + 30c per successful domestic card charge, with more for international cards and currency conversion. The deepest API here by a wide margin, and the one that expects a developer. You remain the seller of record, so the tax is yours to handle.',
  0,94,92,96,66,96,84),
 ('payments','paddle','Paddle','paddle',
  'No monthly fee; 5% + 50c per checkout transaction. Paddle is the merchant of record, meaning it sells to your customer and takes on the sales tax and VAT filing worldwide. That is what the extra 2% over a plain processor buys.',
  0,86,76,86,88,78,66),
 ('payments','lemon-squeezy','Lemon Squeezy','lemon-squeezy',
  'No monthly fee; 5% + 50c per transaction, as merchant of record like Paddle, with the friendliest setup of the three. It is now part of Stripe, which is worth weighing: the product is well supported today, and its independent roadmap no longer exists.',
  0,82,76,64,92,70,64),
 -- ---- course platform -------------------------------------------------
 ('course-platform','gumroad','Gumroad','gumroad',
  'Nothing per month and 10% + 50c of every sale through your own links, rising to 30% when a buyer finds you through Gumroad Discover. Free until you sell, expensive once you do: at 1,000 USD a month it costs more than any subscription here.',
  0,78,84,80,94,58,70),
 ('course-platform','teachable','Teachable Starter','teachable-starter',
  'Course hosting with a storefront and upsells, at 29 USD a month billed annually -- 39 USD if billed monthly. Starter also takes a 7.5% transaction fee, which the higher tiers drop; count that in before comparing it with Thinkific.',
  2900,84,70,84,88,78,72),
 ('course-platform','thinkific','Thinkific Basic','thinkific-basic',
  'Per month billed annually, with no transaction fee on top -- which is why it looks dearer than Teachable Starter and often is not. Built for a single course first, with the course-building tools the most complete of the three.',
  4000,86,68,86,82,84,74)
) AS f(category_slug, brand_slug, name, slug, description, price_minor,
       quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories ON categories.slug = f.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = f.brand_slug
ON CONFLICT (slug) DO NOTHING;

INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url, is_fictional, review_status, reviewed_at, review_note) VALUES
 ('framer-pricing-2026-08','manufacturer_documentation','Framer pricing','Framer','https://www.framer.com/pricing/',false,'verified',now(),'Read 2026-08-21 through a US connection. Basic tier, per month, with the page''s billing control reading yearly.'),
 ('webflow-pricing-2026-08','manufacturer_documentation','Webflow pricing','Webflow','https://webflow.com/pricing',false,'verified',now(),'Read 2026-08-21. Basic site plan, per month billed yearly.'),
 ('squarespace-pricing-2026-08','manufacturer_documentation','Squarespace pricing','Squarespace','https://www.squarespace.com/pricing',false,'verified',now(),'Read 2026-08-21 with the billing control set to pay monthly. Basic tier.'),
 ('ecwid-pricing-2026-08','manufacturer_documentation','Ecwid pricing','Ecwid','https://www.ecwid.com/pricing',false,'verified',now(),'Read 2026-08-21. Starter tier, annual billing, capped at 10 products.'),
 ('shopify-pricing-2026-08','manufacturer_documentation','Shopify pricing','Shopify','https://www.shopify.com/pricing',false,'verified',now(),'Read 2026-08-21 through a US connection. Basic tier. The page''s billing control read pay yearly when the figure was taken; Shopify charges more at monthly billing.'),
 ('bigcommerce-pricing-2026-08','manufacturer_documentation','BigCommerce pricing','BigCommerce','https://www.bigcommerce.com/pricing/',false,'verified',now(),'Read 2026-08-21. Core tier, stated on the page as per month billed annually.'),
 ('stripe-pricing-2026-08','manufacturer_documentation','Stripe pricing','Stripe','https://stripe.com/pricing',false,'verified',now(),'Read 2026-08-21. No monthly fee; 2.9% + 30c per successful domestic card transaction, read from the vendor page.'),
 ('paddle-pricing-2026-08','manufacturer_documentation','Paddle pricing','Paddle','https://www.paddle.com/pricing',false,'verified',now(),'Read 2026-08-21. No monthly fees; 5% + 50c per checkout transaction, read from the vendor page.'),
 ('lemonsqueezy-pricing-2026-08','manufacturer_documentation','Lemon Squeezy pricing','Lemon Squeezy','https://www.lemonsqueezy.com/pricing',false,'verified',now(),'Read 2026-08-21. No monthly charges; 5% + 50c per transaction. The same page carries a 2026 notice that Lemon Squeezy is now part of Stripe.'),
 ('gumroad-pricing-2026-08','manufacturer_documentation','Gumroad pricing','Gumroad','https://gumroad.com/pricing',false,'verified',now(),'Read 2026-08-21. No monthly charges; 10% + 50c per transaction on direct sales, 30% through Gumroad Discover.'),
 ('teachable-pricing-2026-08','manufacturer_documentation','Teachable pricing','Teachable','https://teachable.com/pricing',false,'verified',now(),'Read 2026-08-21. Starter tier at 29 USD billed annually, 39 USD billed monthly, plus a 7.5% transaction fee stated on the same card.'),
 ('thinkific-pricing-2026-08','manufacturer_documentation','Thinkific pricing','Thinkific','https://www.thinkific.com/pricing/',false,'verified',now(),'Read 2026-08-21. Basic tier, stated on the page as per month billed annually.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status='verified', reviewed_at=COALESCE(evidence.sources.reviewed_at,now()),
    source_url=EXCLUDED.source_url, review_note=EXCLUDED.review_note, updated_at=now();

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), m.confidence, m.note
FROM (VALUES
 ('framer-basic','framer-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21.'),
 ('webflow-basic','webflow-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21.'),
 ('squarespace-basic','squarespace-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21 at monthly billing.'),
 ('ecwid-starter','ecwid-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21.'),
 ('shopify-basic','shopify-pricing-2026-08',90,'Price read from the vendor pricing page on 2026-08-21. The billing control read pay yearly; the monthly rate is higher.'),
 ('bigcommerce-core','bigcommerce-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21, stated as per month billed annually.'),
 ('stripe','stripe-pricing-2026-08',100,'No monthly fee; the per-transaction rate was read from the vendor pricing page on 2026-08-21.'),
 ('paddle','paddle-pricing-2026-08',100,'No monthly fee; the per-transaction rate was read from the vendor pricing page on 2026-08-21.'),
 ('lemon-squeezy','lemonsqueezy-pricing-2026-08',100,'No monthly fee; the per-transaction rate was read from the vendor pricing page on 2026-08-21.'),
 ('gumroad','gumroad-pricing-2026-08',100,'No monthly fee; the per-transaction rate was read from the vendor pricing page on 2026-08-21.'),
 ('teachable-starter','teachable-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21 at annual billing, with the transaction fee stated on the same card.'),
 ('thinkific-basic','thinkific-pricing-2026-08',100,'Price read from the vendor pricing page on 2026-08-21, stated as per month billed annually.')
) AS m(product_slug, source_key, confidence, note)
JOIN catalog.products p ON p.slug = m.product_slug
JOIN evidence.sources s ON s.external_key = m.source_key
WHERE NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), 80, 'Suitability scores assigned by UNSOLERO editorial review.'
FROM catalog.products p CROSS JOIN evidence.sources s
WHERE s.external_key='unsolero-editorial-saas-2026-08'
  AND p.slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
                 'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
  AND NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,1,p.category_id,p.brand_id,p.name,p.slug,p.description,p.price_minor,p.currency,p.warranty_months,
       'published',now(),now(),now(),'Read from the vendor pricing page on 2026-08-21 through a US connection. Billing basis stated per product; a zero price means no monthly fee, with the per-sale percentage in the description.'
FROM catalog.products p
WHERE p.slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
                 'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score,
    workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,f.id,1,p.quality_score,p.value_score,p.durability_score,p.beginner_score,p.advanced_score,
       p.apartment_score,p.noise_score,p.portability_score,'published',now(),now(),now(),
       'Editorial suitability assessment; not a vendor claim.'
FROM catalog.products p JOIN evidence.product_fact_revisions f ON f.product_id=p.id AND f.version=1
WHERE p.slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
                 'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT f.id, k.fact_key, o.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions f
JOIN catalog.products p ON p.id=f.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.source_type='manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('description'),('price')) AS k(fact_key)
WHERE f.version=1 AND p.slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
                 'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
SELECT sc.id, k.score_key,
       CASE WHEN k.score_key IN ('apartment','noise')
            THEN 'Not applicable: software has no physical footprint and makes no noise.'
            ELSE 'Editorial suitability assessment based on the published feature set, tier limits, per-sale fees and vendor track record.' END,
       o.id
FROM evidence.score_revisions sc
JOIN catalog.products p ON p.id=sc.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.external_key='unsolero-editorial-saas-2026-08'
CROSS JOIN (VALUES ('quality'),('value'),('durability'),('beginner'),('advanced'),('apartment'),('noise'),('portability')) AS k(score_key)
WHERE sc.version=1 AND p.slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
                 'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS p
SET published_fact_revision_id=f.id, published_score_revision_id=sc.id, status='published'
FROM evidence.product_fact_revisions AS f
JOIN evidence.score_revisions AS sc ON sc.product_id=f.product_id AND sc.version=1
WHERE p.id=f.product_id
  AND p.slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
                 'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
  AND f.version=1 AND f.workflow_status='published' AND sc.workflow_status='published';
