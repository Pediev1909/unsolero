-- Pipedrive becomes the seventh merchant with a real affiliate relationship,
-- after Zoho, SE Ranking, monday.com, Kit, MailerLite and Cal.com.
--
-- Programme approved via PartnerStack; the default link was created by the
-- programme on 2026-08-23 and read from the partner dashboard.
--
-- Two things about this link are worse than every other link in this
-- repository, and both are the programme's doing rather than a mistake here:
--
-- 1. The destination is www.pipedrive.com/programLP — an affiliate programme
--    landing page, not the pricing page. Every other seed here deliberately
--    targets pricing, because someone who has just read "Pipedrive Lite,
--    19.90 USD a seat" is looking for that plan and a generic page makes them
--    hunt for the number they already had. Pipedrive's dashboard offers no
--    custom links and the default cannot be regenerated to a different
--    destination, so this is what is available. Ask affiliates@pipedrive.com
--    for a pricing-page destination before treating this as final.
--
-- 2. The partner key is not exposed in the dashboard: the destination template
--    carries the literal placeholder {partner_key}, which PartnerStack fills at
--    redirect time. The external reference below is therefore the link code
--    itself, which is the ownership identifier actually in hand.
--
-- The link code was read from a dashboard screenshot, not copied from the
-- clipboard. docs/affiliate-links-zoho.md records what that costs: a portal
-- font that renders I, l and 1 identically once produced a link that
-- dead-ended. Confirm this string against the dashboard copy button before
-- relying on the row, and note that a wrong code can silently credit another
-- affiliate rather than fail loudly.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('Pipedrive','pipedrive','https://www.pipedrive.com','EE',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'pipedrive-lite', 'https://www.pipedrive.com/en/pricing',
       1990, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='pipedrive' AND p.slug='pipedrive-lite'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'partnerstack',
       'https://aff.trypipedrive.com/8c0fqmk2j8mc',
       '8c0fqmk2j8mc', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='pipedrive' AND o.merchant_sku='pipedrive-lite'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url,
    external_reference=EXCLUDED.external_reference,
    is_active=EXCLUDED.is_active, updated_at=now();

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM commerce.affiliate_links links
        JOIN commerce.merchant_offers offers ON offers.id = links.merchant_offer_id
        JOIN commerce.merchants merchants ON merchants.id = offers.merchant_id
        JOIN catalog.products products ON products.id = offers.product_id
        WHERE merchants.slug = 'pipedrive'
          AND products.slug = 'pipedrive-lite'
          AND links.destination_url = 'https://aff.trypipedrive.com/8c0fqmk2j8mc'
          AND links.is_active
          AND offers.is_active
    ) THEN
        RAISE EXCEPTION 'Pipedrive affiliate activation failed';
    END IF;
END $$;
