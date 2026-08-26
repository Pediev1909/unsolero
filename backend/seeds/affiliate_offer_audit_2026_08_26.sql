-- Affiliate offer freshness audit performed against first-party pricing and
-- programme evidence on 2026-08-26.
--
-- This script deliberately does not follow affiliate destinations. Automated
-- requests would create artificial provider-side clicks. Instead, it verifies
-- the exact approved destinations already copied from partner dashboards and
-- refreshes only offers whose current price was independently confirmed.
--
-- Zoho Books now uses its reviewed fact and score revision 2: $20/month at
-- monthly billing. The separate $15/month figure requires annual billing.

BEGIN;

DO $$
DECLARE
    verified_count integer;
BEGIN
    SELECT count(*)
    INTO verified_count
    FROM commerce.affiliate_links links
    JOIN commerce.merchant_offers offers
      ON offers.id = links.merchant_offer_id
    JOIN catalog.products products
      ON products.id = offers.product_id
    JOIN (VALUES
        ('bigin-express', 900::bigint, 'USD', 'zoho',
         'https://go.zoho.com/SsgT'),
        ('cal-com-teams', 1200::bigint, 'USD', 'cal_com',
         'https://refer.cal.com/unsolero-wtpd'),
        ('kit-creator', 3900::bigint, 'USD', 'partnerstack',
         'https://partners.kit.com/ge4ub49bd32p-uw04r'),
        ('mailerlite-comfort', 1900::bigint, 'USD', 'trackdesk',
         'https://www.mailerlite.com/?linkId=lp_170762&sourceId=unsolero&tenantId=mailerlite'),
        ('monday-basic', 900::bigint, 'USD', 'partnerstack',
         'https://try.monday.com/o8rui6cwqodl-p7wfga'),
        ('se-ranking-core', 10320::bigint, 'USD', 'se_ranking',
         'https://seranking.com/subscription.html?ga=5233991&source=link'),
        ('zoho-bookings-basic', 800::bigint, 'USD', 'zoho',
         'https://go.zoho.com/POSi'),
        ('zoho-books-standard', 2000::bigint, 'USD', 'zoho',
         'https://go.zoho.com/K0nf'),
        ('zoho-campaigns-standard', 525::bigint, 'USD', 'zoho',
         'https://go.zoho.com/UCST'),
        ('zoho-crm-standard', 2000::bigint, 'USD', 'zoho',
         'https://go.zoho.com/dNbV'),
        ('zoho-invoice', 0::bigint, 'USD', 'zoho',
         'https://go.zoho.com/dIhI'),
        ('zoho-projects-premium', 400::bigint, 'USD', 'zoho',
         'https://go.zoho.com/PCoS')
    ) AS expected(product_slug, price_minor, currency, provider, destination_url)
      ON expected.product_slug = products.slug
     AND expected.price_minor = offers.price_minor
     AND expected.currency = offers.currency
     AND expected.provider = links.provider
     AND expected.destination_url = links.destination_url
    WHERE offers.is_active
      AND links.is_active;

    IF verified_count <> 12 THEN
        RAISE EXCEPTION
            'Affiliate audit failed: expected 12 exact active offer/link matches, found %',
            verified_count;
    END IF;

    UPDATE commerce.merchant_offers offers
    SET last_checked_at = now(),
        updated_at = now()
    FROM catalog.products products
    WHERE products.id = offers.product_id
      AND (products.slug, offers.price_minor, offers.currency) IN (
          ('bigin-express', 900, 'USD'),
          ('cal-com-teams', 1200, 'USD'),
          ('kit-creator', 3900, 'USD'),
          ('mailerlite-comfort', 1900, 'USD'),
          ('monday-basic', 900, 'USD'),
          ('se-ranking-core', 10320, 'USD'),
          ('zoho-bookings-basic', 800, 'USD'),
          ('zoho-books-standard', 2000, 'USD'),
          ('zoho-campaigns-standard', 525, 'USD'),
          ('zoho-crm-standard', 2000, 'USD'),
          ('zoho-invoice', 0, 'USD'),
          ('zoho-projects-premium', 400, 'USD')
      );

END
$$;

COMMIT;
