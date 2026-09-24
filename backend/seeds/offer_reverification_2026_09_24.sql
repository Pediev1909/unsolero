-- Offer re-verification, 2026-09-24.
--
-- Why: every offer query and both redirects filter on
-- last_checked_at >= now() - OFFER_MAXIMUM_AGE (720h in production), and
-- last_checked_at is only ever written by a seed. No seed had touched the
-- Zoho, MailerLite, Kit and Cal.com offers since 2026-08-26, so ten of the
-- fifteen live offers were due to stop resolving at 2026-09-25 10:48 UTC,
-- Pipedrive on 09-27, ActiveCampaign on 09-28, and the rest on 10-02. The
-- three affiliate promotions expire on the same clock.
--
-- What this file does, and only this: for each offer whose price was read
-- again today and matched what the catalog and the offer already say, it
-- records today's read as an evidence source and observation and moves the
-- offer's last_checked_at to now(). It changes no price, no fact revision
-- and no score. It refuses to mark an offer checked if the catalog price or
-- the offer price differs from the figure read today — a mismatch has to be
-- corrected by a price seed first, never papered over with a fresh date.
--
-- Read 2026-09-24 (headless Chrome from Bulgaria; where the visible page
-- rendered EUR, the USD figure was taken from the page's own multi-currency
-- payload, as the 2026-09-02 audit did):
--   zoho-invoice             0      free          zoho.com/invoice/pricing — "$0"
--   zoho-books-standard      2000   monthly       zoho.com/us/books/pricing — Standard $20 monthly, $15 annually
--   bigin-express            900    monthly       bigin.com/pricing.html — data-price USD 9 monthly // 7 annually
--   zoho-crm-standard        2000   monthly       zoho.com/crm/zohocrm-pricing.html — data-price USD 20 // 14
--   zoho-campaigns-standard  525    annual        zoho.com/campaigns/pricing.html — Standard 501–1,000 contacts, USD 7 monthly // 5.25 annually
--   zoho-projects-premium    400    annual        zoho.com/projects/pricing.html — Premium USD 5 monthly // 4 annually; Free up to 5 users
--   zoho-bookings-basic      800    monthly       zoho.com/bookings/pricing.html — Basic USD 8 // 6
--   teachable-starter        3900   monthly       teachable.com/pricing — rendered $39 billed monthly, $29 billed annually, 7.5% fee
--   mailerlite-comfort       1900   monthly       mailerlite.com/pricing — plan data, 1,000 subscribers, USD 19 monthly // 17.10 yearly
--   activecampaign-starter   1500   annual        activecampaign.com/pricing — "Starts at $15 /mo billed annually"
--   kit-creator              3900   monthly       kit.com/pricing — $390 billed yearly, "Save $78 per year" = 12 x $39
--   monday-basic             1200   monthly       monday.com/pricing — page's Offer markup: USD 12 monthly, 9 annually, per seat
--   cal-com-teams            1200   annual        cal.com/pricing — Teams "YEARLY $12 per user/month"
--
-- Not re-verified, deliberately left to expire unless read by hand:
--   pipedrive-lite   — pipedrive.com/pricing returns a Cloudflare block page to
--                      automated reads; no price could be read.
--   se-ranking-core  — seranking.com renders EUR only from Bulgaria (Core
--                      €109 / €87.20) and carries no USD figure in the page.
--
-- Found, not acted on here (a price move, not a re-verification):
--   zoho-campaigns-standard and zoho-projects-premium now publish a monthly
--   rate ($7 and $5). The site rule compares monthly-billing prices, so both
--   should move — which also changes editorial prose that quotes 5.25 and 4
--   (the agency stack total of 59). The annual figures recorded today are
--   still true, and are labelled annual on the site, so re-verifying them is
--   honest; moving them is a separate, deliberate seed.
--
-- The three promotions were checked by following their destination URLs,
-- which returned 200 at the expected landing pages.
--
-- Re-running the file is safe: sources upsert, an observation is written
-- once per product per day, and last_checked_at moves to the time of the run.
--
-- Applied with psql:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f backend/seeds/offer_reverification_2026_09_24.sql

BEGIN;

CREATE FUNCTION pg_temp.reverify_offer(
    product_slug text,
    expected_minor bigint,
    source_url text,
    source_title text,
    source_publisher text,
    read_note text
) RETURNS void
LANGUAGE plpgsql AS $$
DECLARE
    product_row catalog.products%ROWTYPE;
    read_source_id uuid;
    mismatched integer;
    updated integer;
BEGIN
    SELECT * INTO product_row FROM catalog.products WHERE slug = product_slug;
    IF product_row.id IS NULL THEN
        RAISE EXCEPTION '%: product is missing', product_slug;
    END IF;
    IF product_row.price_minor <> expected_minor THEN
        RAISE EXCEPTION '%: catalog price is % but % was read today; correct the price first',
            product_slug, product_row.price_minor, expected_minor;
    END IF;
    SELECT count(*) INTO mismatched FROM commerce.merchant_offers
    WHERE product_id = product_row.id AND is_active AND price_minor <> expected_minor;
    IF mismatched > 0 THEN
        RAISE EXCEPTION '%: % active offer(s) quote a price other than %', product_slug, mismatched, expected_minor;
    END IF;

    INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url,
                                  is_fictional, review_status, reviewed_at, review_note)
    VALUES (product_slug || '-pricing-2026-09-24', 'manufacturer_documentation', source_title,
            source_publisher, source_url, false, 'verified', now(),
            'Offer re-verification 2026-09-24. ' || read_note)
    ON CONFLICT (external_key) DO UPDATE SET
        source_url = EXCLUDED.source_url, review_status = 'verified',
        reviewed_at = COALESCE(evidence.sources.reviewed_at, now()),
        review_note = EXCLUDED.review_note, updated_at = now()
    RETURNING id INTO read_source_id;

    INSERT INTO evidence.observations (source_id, product_id, observed_at, expires_at, confidence, notes)
    SELECT read_source_id, product_row.id, now(), now() + interval '90 days', 100,
           product_row.name || ' read on 2026-09-24 at ' || source_url || ': ' || read_note
    WHERE NOT EXISTS (
        SELECT 1 FROM evidence.observations existing
        WHERE existing.source_id = read_source_id AND existing.product_id = product_row.id
    );

    UPDATE commerce.merchant_offers
    SET last_checked_at = now(), updated_at = now()
    WHERE product_id = product_row.id AND is_active;
    GET DIAGNOSTICS updated = ROW_COUNT;
    IF updated = 0 THEN
        RAISE EXCEPTION '%: no active offer to re-verify', product_slug;
    END IF;
END $$;

SELECT pg_temp.reverify_offer('zoho-invoice', 0,
    'https://www.zoho.com/invoice/pricing/', 'Zoho Invoice pricing page', 'Zoho',
    'Free: the page states $0 and "free for all small businesses".');
SELECT pg_temp.reverify_offer('zoho-books-standard', 2000,
    'https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard', 'Zoho Books pricing comparison', 'Zoho',
    'Standard $20 per organization per month billed monthly; $15 billed annually.');
SELECT pg_temp.reverify_offer('bigin-express', 900,
    'https://www.bigin.com/pricing.html', 'Bigin pricing page', 'Zoho',
    'Express USD 9 per user per month billed monthly, 7 billed annually, read from the page''s data-price payload (currency order USD first per zohobigin-pricing-val.json; visible page renders EUR from Bulgaria). Free edition $0.');
SELECT pg_temp.reverify_offer('zoho-crm-standard', 2000,
    'https://www.zoho.com/crm/zohocrm-pricing.html', 'Zoho CRM pricing page', 'Zoho',
    'Standard USD 20 per user per month billed monthly, 14 billed annually, read from the page''s data-price payload (currency order USD first per crm-pricing-val.json).');
SELECT pg_temp.reverify_offer('zoho-campaigns-standard', 525,
    'https://www.zoho.com/campaigns/pricing.html', 'Zoho Campaigns pricing page', 'Zoho',
    'Standard, 501-1,000 contacts: USD 5.25 per month billed annually (the recorded figure) and USD 7 billed monthly, read from the page''s data-price payload. A monthly rate is now published; the compared price has not been moved by this file.');
SELECT pg_temp.reverify_offer('zoho-projects-premium', 400,
    'https://www.zoho.com/projects/pricing.html', 'Zoho Projects pricing page', 'Zoho',
    'Premium USD 4 per user per month billed annually (the recorded figure) and USD 5 billed monthly, read from the page''s data-price payload. Free plan: up to 5 users. A monthly rate is now published; the compared price has not been moved by this file.');
SELECT pg_temp.reverify_offer('zoho-bookings-basic', 800,
    'https://www.zoho.com/bookings/pricing.html', 'Zoho Bookings pricing page', 'Zoho',
    'Basic USD 8 per user per month billed monthly, 6 billed annually, read from the page''s data-price payload.');
SELECT pg_temp.reverify_offer('teachable-starter', 3900,
    'https://teachable.com/pricing', 'Teachable pricing page', 'Teachable',
    'Rendered page: Starter $39/mo billed monthly, $29/mo billed annually; 7.5% transaction fee on Starter.');
SELECT pg_temp.reverify_offer('mailerlite-comfort', 1900,
    'https://www.mailerlite.com/pricing', 'MailerLite pricing page', 'MailerLite',
    'Comfort at 1,000 subscribers: USD 19 billed monthly, 17.10 billed yearly, read from the page''s plan data. Free plan: up to 250 subscribers, 2,500 monthly emails.');
SELECT pg_temp.reverify_offer('activecampaign-starter', 1500,
    'https://www.activecampaign.com/pricing', 'ActiveCampaign pricing page', 'ActiveCampaign',
    'Starter "Starts at $15 /mo, billed annually"; no monthly rate published.');
SELECT pg_temp.reverify_offer('kit-creator', 3900,
    'https://kit.com/pricing', 'Kit pricing page', 'Kit',
    'Creator at 1,000 subscribers: $390 billed yearly ($33/mo), "Save $78 per year", so the monthly-billing price is $39. Free plan up to 10,000 subscribers.');
SELECT pg_temp.reverify_offer('monday-basic', 1200,
    'https://monday.com/pricing', 'monday.com pricing page', 'monday.com',
    'Basic USD 12 per seat billed monthly, 9 billed annually, read from the page''s Offer markup (visible page renders EUR from Bulgaria: EUR 9 per seat billed annually). Free plan now up to 2 seats.');
SELECT pg_temp.reverify_offer('cal-com-teams', 1200,
    'https://cal.com/pricing', 'Cal.com pricing page', 'Cal.com',
    'Teams "YEARLY $12 per user/month", save 25%; monthly figure not displayed.');

DO $$
DECLARE
    updated integer;
BEGIN
    UPDATE commerce.affiliate_promotions
    SET last_checked_at = now(), updated_at = now()
    WHERE is_active AND slug IN (
        'funnel-hacking-secrets-webinar',
        'funnel-hacking-secrets-order',
        'activecampaign-mailchimp-switch');
    GET DIAGNOSTICS updated = ROW_COUNT;
    IF updated <> 3 THEN
        RAISE EXCEPTION 'expected 3 active promotions to re-verify, found %', updated;
    END IF;
END $$;

COMMIT;
