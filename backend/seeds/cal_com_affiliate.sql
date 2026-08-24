-- Cal.com becomes the fifth merchant with a real affiliate relationship, after
-- Zoho, SE Ranking, monday.com and Kit.
--
-- Taken from the Cal.com referral dashboard, not assembled here. Verified with
-- a single request that stopped at the redirect rather than following it:
-- refer.cal.com answers 302 to cal.com carrying via=unsolero-wtpd, tracked
-- through Dub. Following the chain would have registered a click against the
-- account, which is worth avoiding on a link that exists to count real ones.
--
-- 20% of revenue for twelve months. Cal.com runs the programme itself, with no
-- network in front of it, which is why an operator with no audience could join
-- at all.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('Cal.com','cal-com','https://cal.com','US',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'cal-com-teams', 'https://cal.com/pricing',
       1200, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='cal-com' AND p.slug='cal-com-teams'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'cal_com',
       'https://refer.cal.com/unsolero-wtpd',
       'unsolero-wtpd', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='cal-com' AND o.merchant_sku='cal-com-teams'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url,
    external_reference=EXCLUDED.external_reference,
    is_active=EXCLUDED.is_active, updated_at=now();
