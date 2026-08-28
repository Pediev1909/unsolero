-- Teachable becomes the eighth merchant with a real affiliate relationship,
-- after Zoho, SE Ranking, monday.com, Kit, MailerLite, Cal.com and Pipedrive.
--
-- Programme approved via PartnerStack on 2026-08-21; 30% for one year. The
-- link was copied from the partner dashboard's copy control, not transcribed
-- from the screen — the dashboard truncates it for display, and
-- docs/affiliate-links-zoho.md records what transcribing these costs.
--
-- The destination is already the pricing page, set by the programme itself:
-- teachable.com/pricing?utm_term=self-serve&utm_medium=affiliate-link
-- &utm_source=partnerstack&utm_content={partner_key}&utm_campaign={xid}
-- The two braced values are literal placeholders that PartnerStack fills at
-- redirect time, so the partner key is never exposed to the partner. The
-- external reference below is therefore the link code itself, which is the
-- ownership identifier actually in hand — the same situation as Pipedrive.
--
-- ---------------------------------------------------------------------------
-- On the price, because it is the one number here that is not simply read off
-- a page.
--
-- The offer is recorded at 29.00 USD, which is Teachable Starter's ANNUAL
-- billing rate. Monthly billing is 39.00 USD, and the Zoho Books correction of
-- 2026-08-26 established that UNSOLERO compares monthly-billing prices.
--
-- It is left at the annual rate deliberately, and this is not the Zoho Books
-- situation. Thinkific Basic — Teachable Starter's direct rival in the same
-- category — is also recorded at its annual rate of 40.00 USD, and the two
-- products' descriptions compare them to each other explicitly. Within the
-- course-platform category the comparison is therefore already like for like.
-- Moving Teachable alone to 39.00 USD would break a correct comparison in
-- order to fix a catalog-wide inconsistency, making the visible result worse.
--
-- The catalog-wide question — that several products outside this category
-- carry annual rates while the stated policy compares monthly ones — is real
-- and still open. It needs both vendor pages re-read and a policy revision, in
-- the same shape as the Zoho Books correction. It is not fixed here, and this
-- comment exists so that it is not mistaken for having been overlooked.
--
-- The product description already states both figures and names the billing
-- basis, so nothing shown to a visitor is ambiguous today.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('Teachable','teachable','https://teachable.com','US',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'teachable-starter', 'https://teachable.com/pricing',
       2900, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='teachable' AND p.slug='teachable-starter'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'partnerstack',
       'https://partnerstack.teachable.com/y6u7cxavunjg',
       'y6u7cxavunjg', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='teachable' AND o.merchant_sku='teachable-starter'
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
        WHERE merchants.slug = 'teachable'
          AND products.slug = 'teachable-starter'
          AND links.destination_url = 'https://partnerstack.teachable.com/y6u7cxavunjg'
          AND links.is_active
          AND offers.is_active
    ) THEN
        RAISE EXCEPTION 'Teachable affiliate activation failed';
    END IF;
END $$;
