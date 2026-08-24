-- Kit becomes the fourth merchant with a real affiliate relationship, after
-- Zoho, SE Ranking and monday.com.
--
-- The link and the partner key were taken from the PartnerStack dashboard, not
-- assembled here. Verified by following it: partners.kit.com issues a 302 to
-- kit.com carrying ps_partner_key, and sets partner cookies on .kit.com with
-- Max-Age 7776000 — a 90-day attribution window.
--
-- Kit's dashboard also offers an "Upgrade" link. It is not used here: following
-- it lands on app.kit.com/upgrade, which bounces an unauthenticated visitor to
-- a login form. Someone arriving from a product page has no Kit account yet, so
-- that link would send every one of them to a dead end.
--
-- The destination is the pricing page, not the home page. Someone who has just
-- read "Kit Creator, 39.00 USD a month" is looking for that plan; a home page
-- makes them hunt for the number they already had, and most will not.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('Kit','kit','https://kit.com','US',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'kit-creator', 'https://kit.com/pricing',
       3900, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='kit' AND p.slug='kit-creator'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'partnerstack',
       'https://partners.kit.com/ge4ub49bd32p-uw04r',
       '93ff92f9d392', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='kit' AND o.merchant_sku='kit-creator'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url,
    external_reference=EXCLUDED.external_reference,
    is_active=EXCLUDED.is_active, updated_at=now();
