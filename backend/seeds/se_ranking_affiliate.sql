-- SE Ranking becomes the second merchant, and the first affiliate link on the
-- site that is not Zoho.
--
-- The link points at SE Ranking's pricing page rather than their home page.
-- Somebody who has just read "SE Ranking Core, 103.20 USD a month" and clicks
-- expects to land on that plan. Dropping them on a home page makes them hunt
-- for the number they already had, and most of them will not.
--
-- The affiliate code was copied from SE Ranking's own link builder rather than
-- assembled here from the account id. A link that resolves proves nothing
-- about whose it is, and a stranger's code on this site would pay a stranger.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status) VALUES
 ('SE Ranking','se-ranking','https://seranking.com','US',80,'active')
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url, price_minor,
    shipping_minor, currency, availability, condition, last_checked_at)
SELECT m.id, p.id, 'se-ranking-core', 'https://seranking.com/subscription.html',
       10320, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug='se-ranking' AND p.slug='se-ranking-core'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    price_minor=EXCLUDED.price_minor, product_url=EXCLUDED.product_url,
    last_checked_at=EXCLUDED.last_checked_at, updated_at=now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active)
SELECT o.id, 'se_ranking',
       'https://seranking.com/subscription.html?ga=5233991&source=link',
       '5233991', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id=o.merchant_id
WHERE m.slug='se-ranking' AND o.merchant_sku='se-ranking-core'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url=EXCLUDED.destination_url,
    external_reference=EXCLUDED.external_reference,
    is_active=EXCLUDED.is_active, updated_at=now();
