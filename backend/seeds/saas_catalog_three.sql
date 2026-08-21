-- Three categories opened at once: help desk, scheduling, analytics.
--
-- All nine prices read on 2026-08-21 through a US connection, in USD, from the
-- vendor's own pricing page. Where a page would not give up its numbers to
-- reading, the prices were pulled from the page structure instead — the element
-- whose own text is a price, and the nearest heading above it. That is how
-- Freshdesk and Calendly were finally read after the plain text failed twice.
--
-- BASIS. Stated per product because the three categories measure differently:
--   help desk    per agent or seat, per month, monthly billing
--   scheduling   per user or seat, per month
--   analytics    per month at 100,000 pageviews — the same volume for all
--                three, which is what makes that column an actual comparison
--
-- Zoho Desk is absent. Its pricing page publishes no number that can be read,
-- by text or by structure, and a Zoho product was not going to be added on a
-- guess just because a referral link for it was already in hand.

INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Help Scout','help-scout','https://www.helpscout.com','US'),
 ('Tidio','tidio','https://www.tidio.com','PL'),
 ('Calendly','calendly','https://calendly.com','US'),
 ('Cal.com','cal-com','https://cal.com','US'),
 ('Fathom Analytics','fathom','https://usefathom.com','CA'),
 ('Simple Analytics','simple-analytics','https://www.simpleanalytics.com','NL'),
 ('Umami','umami','https://umami.is','US')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, f.name, f.slug, f.description, f.price_minor, 'USD', 0,
       f.quality, f.value, f.durability, f.beginner, f.advanced, 0, 0, f.portability
FROM (VALUES
 -- ---- help desk -------------------------------------------------------
 ('help-desk','freshworks','Freshdesk Growth','freshdesk-growth',
  'Shared inbox, ticketing and a customer portal, with an AI agent included for the first 500 sessions. Per agent per month at monthly billing. The heaviest of the three here, and the one that scales furthest before you outgrow it.',
  2300,84,78,78,80,82,76),
 ('help-desk','tidio','Tidio Starter','tidio-starter',
  'Live chat first, help desk second, which is the right way round if most of your support arrives before the sale rather than after it. The Starter tier is capped at 100 billable conversations a month — that cap, not the price, is what decides whether it fits.',
  2417,82,76,76,88,70,74),
 ('help-desk','help-scout','Help Scout Standard','help-scout-standard',
  'A shared inbox that looks like email to the customer and like a queue to you, with no ticket numbers and no portal. Per user per month. The simplest of the three by some distance, which is the whole reason to choose it.',
  2500,90,80,84,90,76,82),
 -- ---- scheduling ------------------------------------------------------
 ('scheduling','zoho','Zoho Bookings Basic','zoho-bookings-basic',
  'Appointment booking with staff calendars and payment collection, per user per month. A free tier exists. Its advantage is not the price but that it already knows about your Zoho CRM records.',
  800,78,92,92,82,72,78),
 ('scheduling','calendly','Calendly Standard','calendly-standard',
  'The one everyone recognises, which matters more than it should: a booking link nobody has to be taught to use. Per seat per month. Seats are needed for anyone who connects a calendar and hosts meetings.',
  1000,90,84,88,94,78,80),
 ('scheduling','cal-com','Cal.com Teams','cal-com-teams',
  'Open source, and self-hostable if you would rather your booking data never left your own server. Per user per month billed yearly; free forever for an individual. The trade is setup effort against control.',
  1200,84,80,74,76,88,92),
 -- ---- analytics -------------------------------------------------------
 ('analytics','fathom','Fathom Analytics','fathom-analytics',
  'Privacy-first web analytics with no cookie banner required, priced by pageviews rather than by seat. The figure shown is for 100,000 pageviews a month, and covers up to 50 sites at that tier.',
  1500,88,86,80,92,72,84),
 ('analytics','simple-analytics','Simple Analytics','simple-analytics',
  'A single-page dashboard with nothing to configure, which is the entire product. Priced by pageviews; the figure shown is for 100,000 a month. Choose it when you want one number a day, not a reporting tool.',
  2000,84,78,76,92,68,82),
 ('analytics','umami','Umami Cloud Pro','umami-cloud-pro',
  'Open source analytics, hosted for you or on your own server. Free up to 100,000 events a month, which is where the other two here start charging. Counts events rather than pageviews, so a busy single-page app registers more than it looks.',
  2000,82,88,74,84,80,94)
) AS f(category_slug, brand_slug, name, slug, description, price_minor,
       quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories ON categories.slug = f.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = f.brand_slug
ON CONFLICT (slug) DO NOTHING;

INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url, is_fictional, review_status, reviewed_at, review_note) VALUES
 ('freshdesk-pricing-2026-08','manufacturer_documentation','Freshdesk pricing','Freshworks','https://www.freshworks.com/freshdesk/pricing/',false,'verified',now(),'Read 2026-08-21 through a US connection. Growth tier, per agent per month, monthly billing. The figure was read from the page structure after two attempts at reading its text failed.'),
 ('tidio-pricing-2026-08','manufacturer_documentation','Tidio pricing','Tidio','https://www.tidio.com/pricing/',false,'verified',now(),'Read 2026-08-21 through a US connection. Starter tier, monthly, 100 billable conversations.'),
 ('helpscout-pricing-2026-08','manufacturer_documentation','Help Scout pricing','Help Scout','https://www.helpscout.com/pricing/',false,'verified',now(),'Read 2026-08-21. Standard tier, per user per month.'),
 ('zoho-bookings-pricing-2026-08','manufacturer_documentation','Zoho Bookings pricing','Zoho','https://www.zoho.com/bookings/pricing.html',false,'verified',now(),'Read 2026-08-21 through a US connection. Basic tier, per user per month.'),
 ('calendly-pricing-2026-08','manufacturer_documentation','Calendly pricing','Calendly','https://calendly.com/pricing',false,'verified',now(),'Read 2026-08-21. Standard tier, per seat per month. Read from the page structure; the plain text gave plan names without prices.'),
 ('calcom-pricing-2026-08','manufacturer_documentation','Cal.com pricing','Cal.com','https://cal.com/pricing',false,'verified',now(),'Read 2026-08-21. Teams tier, per user per month billed yearly. Individuals are free.'),
 ('fathom-pricing-2026-08','manufacturer_documentation','Fathom Analytics pricing','Fathom Analytics','https://usefathom.com/pricing',false,'verified',now(),'Read 2026-08-21 with the pageview selector set to 100,000 a month.'),
 ('simpleanalytics-pricing-2026-08','manufacturer_documentation','Simple Analytics pricing','Simple Analytics','https://www.simpleanalytics.com/pricing',false,'verified',now(),'Read 2026-08-21 at 100,000 pageviews a month.'),
 ('umami-pricing-2026-08','manufacturer_documentation','Umami pricing','Umami','https://umami.is/pricing',false,'verified',now(),'Read 2026-08-21. Pro tier; the Hobby tier is free up to 100,000 events a month.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status='verified', reviewed_at=COALESCE(evidence.sources.reviewed_at,now()),
    source_url=EXCLUDED.source_url, review_note=EXCLUDED.review_note, updated_at=now();

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), 100, 'Price read from the vendor pricing page on 2026-08-21 through a US connection.'
FROM (VALUES
 ('freshdesk-growth','freshdesk-pricing-2026-08'),
 ('tidio-starter','tidio-pricing-2026-08'),
 ('help-scout-standard','helpscout-pricing-2026-08'),
 ('zoho-bookings-basic','zoho-bookings-pricing-2026-08'),
 ('calendly-standard','calendly-pricing-2026-08'),
 ('cal-com-teams','calcom-pricing-2026-08'),
 ('fathom-analytics','fathom-pricing-2026-08'),
 ('simple-analytics','simpleanalytics-pricing-2026-08'),
 ('umami-cloud-pro','umami-pricing-2026-08')
) AS m(product_slug, source_key)
JOIN catalog.products p ON p.slug = m.product_slug
JOIN evidence.sources s ON s.external_key = m.source_key
WHERE NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), 80, 'Suitability scores assigned by UNSOLERO editorial review.'
FROM catalog.products p CROSS JOIN evidence.sources s
WHERE s.external_key='unsolero-editorial-saas-2026-08'
  AND p.slug IN ('freshdesk-growth','tidio-starter','help-scout-standard','zoho-bookings-basic',
                 'calendly-standard','cal-com-teams','fathom-analytics','simple-analytics','umami-cloud-pro')
  AND NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,1,p.category_id,p.brand_id,p.name,p.slug,p.description,p.price_minor,p.currency,p.warranty_months,
       'published',now(),now(),now(),'Read from the vendor pricing page on 2026-08-21 through a US connection. Billing basis stated per product.'
FROM catalog.products p
WHERE p.slug IN ('freshdesk-growth','tidio-starter','help-scout-standard','zoho-bookings-basic',
                 'calendly-standard','cal-com-teams','fathom-analytics','simple-analytics','umami-cloud-pro')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score,
    workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,f.id,1,p.quality_score,p.value_score,p.durability_score,p.beginner_score,p.advanced_score,
       p.apartment_score,p.noise_score,p.portability_score,'published',now(),now(),now(),
       'Editorial suitability assessment; not a vendor claim.'
FROM catalog.products p JOIN evidence.product_fact_revisions f ON f.product_id=p.id AND f.version=1
WHERE p.slug IN ('freshdesk-growth','tidio-starter','help-scout-standard','zoho-bookings-basic',
                 'calendly-standard','cal-com-teams','fathom-analytics','simple-analytics','umami-cloud-pro')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT f.id, k.fact_key, o.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions f
JOIN catalog.products p ON p.id=f.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.source_type='manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('description'),('price')) AS k(fact_key)
WHERE f.version=1 AND p.slug IN ('freshdesk-growth','tidio-starter','help-scout-standard','zoho-bookings-basic',
                 'calendly-standard','cal-com-teams','fathom-analytics','simple-analytics','umami-cloud-pro')
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
WHERE sc.version=1 AND p.slug IN ('freshdesk-growth','tidio-starter','help-scout-standard','zoho-bookings-basic',
                 'calendly-standard','cal-com-teams','fathom-analytics','simple-analytics','umami-cloud-pro')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS p
SET published_fact_revision_id=f.id, published_score_revision_id=sc.id, status='published'
FROM evidence.product_fact_revisions AS f
JOIN evidence.score_revisions AS sc ON sc.product_id=f.product_id AND sc.version=1
WHERE p.id=f.product_id
  AND p.slug IN ('freshdesk-growth','tidio-starter','help-scout-standard','zoho-bookings-basic',
                 'calendly-standard','cal-com-teams','fathom-analytics','simple-analytics','umami-cloud-pro')
  AND f.version=1 AND f.workflow_status='published' AND sc.workflow_status='published';

-- Zoho Bookings is the seventh of the 91 referral links to find a home.
INSERT INTO commerce.merchant_offers (merchant_id, product_id, merchant_sku, product_url, price_minor, shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'zoho-bookings-basic', 'https://www.zoho.com/bookings/pricing.html', 800, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='zoho' AND p.slug='zoho-bookings-basic'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (merchant_offer_id, provider, destination_url, external_reference, disclosure_label, is_active)
SELECT o.id, 'zoho', 'https://go.zoho.com/POSi', 'PE2263909', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='zoho' AND o.merchant_sku='zoho-bookings-basic'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url, is_active=EXCLUDED.is_active, updated_at=now();
