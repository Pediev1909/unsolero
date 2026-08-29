-- ActiveCampaign becomes the ninth merchant with a real affiliate
-- relationship, after Zoho, SE Ranking, monday.com, Kit, MailerLite, Cal.com,
-- Pipedrive and Teachable.
--
-- Programme approved via PartnerStack. The dashboard publishes three default
-- links, all read from it on 2026-08-29:
--
--   Free Trial Page      https://try.activecampaign.com/egv7yratfy4m-rvs4jt
--   Pricing Page         https://try.activecampaign.com/ydwpszsmins9
--   Affiliate Homepage   https://try.activecampaign.com/4iq2pjt98jsg-b9q17i
--
-- Only the pricing link is used, and only one of the three could have been.
-- ResolveOfferDestination picks a single row per offer — ORDER BY priority
-- DESC, provider LIMIT 1 — and commerce.affiliate_links is unique on
-- (merchant_offer_id, provider), so a second PartnerStack link on this offer
-- could not be stored, and a third stored under an invented provider name
-- would never be served. The other two links are recorded above as evidence of
-- what the programme offers, not as dead rows in the database.
--
-- The pricing page is the right one of the three on the same reasoning as the
-- Kit seed: someone who has just read "ActiveCampaign Starter, 15.00 USD"
-- is looking for that plan, and a home page makes them hunt for the number
-- they already had. The free-trial link is the only real alternative — the
-- product has no free tier, so a trial is the actual entry path — but
-- ActiveCampaign's pricing page carries its own trial call to action, so the
-- pricing destination reaches the trial in one more click while the trial page
-- does not reach the price at all. Hold the trial link in reserve for
-- editorial placements where the price has not already been shown.
--
-- The partner key is not exposed in the dashboard, as with Pipedrive and
-- Teachable: PartnerStack fills it at redirect time. The external reference
-- below is therefore the link code itself, which is the ownership identifier
-- actually in hand.
--
-- All three codes came from the dashboard's copy control, confirmed by the
-- account owner on 2026-08-29. Nothing here was read off the screen, which is
-- the failure docs/affiliate-links-zoho.md exists to warn about — the Zoho
-- portal font renders I, l and 1 identically and once produced a link that
-- dead-ended. Their prefixes also match the dashboard screenshot, which
-- truncates every URL for display.
--
-- Commission terms are deliberately absent. The programme's rate and cookie
-- window were not read from the dashboard, and the commission columns exist to
-- record what a programme states, not what a seed assumes. Fill them in from
-- the PartnerStack programme terms in the shape of mailerlite_affiliate.sql.
--
-- ---------------------------------------------------------------------------
-- On the price.
--
-- The offer is recorded at 15.00 USD, matching the catalog product: 1,000
-- contacts, billed annually, read from the vendor pricing page on 2026-08-20
-- and asserted by source activecampaign-pricing-2026-08.
--
-- ActiveCampaign is one of the twenty-five products carrying an annual rate
-- while the stated policy compares monthly ones. Unlike Teachable, there is no
-- monthly figure to move to: ActiveCampaign publishes no monthly rate for this
-- plan at all, which is why the product description says so. Attaching a link
-- does not change that, and this seed does not attempt to. The catalog-wide
-- billing-basis defect is recorded in docs/AFFILIATE_PROGRAMS.md and needs a
-- migration, not a per-product edit.

BEGIN;

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('ActiveCampaign','activecampaign','https://www.activecampaign.com','US',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'activecampaign-starter', 'https://www.activecampaign.com/pricing',
       1500, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='activecampaign' AND p.slug='activecampaign-starter'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'partnerstack',
       'https://try.activecampaign.com/ydwpszsmins9',
       'ydwpszsmins9', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='activecampaign' AND o.merchant_sku='activecampaign-starter'
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
        WHERE merchants.slug = 'activecampaign'
          AND products.slug = 'activecampaign-starter'
          AND offers.price_minor = 1500
          AND offers.currency = 'USD'
          AND links.provider = 'partnerstack'
          AND links.destination_url = 'https://try.activecampaign.com/ydwpszsmins9'
          AND links.is_active
          AND offers.is_active
    ) THEN
        RAISE EXCEPTION 'ActiveCampaign affiliate activation failed';
    END IF;
END $$;

COMMIT;
