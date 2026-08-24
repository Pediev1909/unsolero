-- monday.com becomes the third merchant with a real affiliate relationship,
-- after Zoho and SE Ranking.
--
-- The link and the partner key were taken from the PartnerStack dashboard, not
-- assembled here. Verified by following it: try.monday.com issues a 302 to
-- monday.com carrying ps_partner_key=unsolero3455, and sets partner cookies on
-- .monday.com with Max-Age 7776000 — a 90-day attribution window.
--
-- The destination is the pricing page, not the account's default home-page
-- link. Someone who has just read "monday.com Basic, 9.00 USD a month" is
-- looking for that plan; a home page makes them hunt for the number they
-- already had, and most will not.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('monday.com','monday-com','https://monday.com','IL',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'monday-basic', 'https://monday.com/pricing',
       900, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='monday-com' AND p.slug='monday-basic'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'partnerstack',
       'https://try.monday.com/o8rui6cwqodl-p7wfga',
       'unsolero3455', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='monday-com' AND o.merchant_sku='monday-basic'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url,
    external_reference=EXCLUDED.external_reference,
    is_active=EXCLUDED.is_active, updated_at=now();
