-- Real merchant offers and affiliate links.
--
-- Unlike the *_demo fixtures, nothing here is fictional: these are live
-- programmes with signed terms. An offer only appears once its product exists
-- in the catalog and its programme has approved this site, so this file grows
-- one vendor at a time rather than being generated.
--
-- Prices are the vendor's own list price, observed on last_checked_at. They are
-- a claim with a date on it, not a live feed.

INSERT INTO commerce.merchants (name, slug, website_url, country_code, trust_score, status)
VALUES ('Zoho', 'zoho', 'https://www.zoho.com', 'US', 85, 'active')
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    website_url = EXCLUDED.website_url,
    country_code = EXCLUDED.country_code,
    trust_score = EXCLUDED.trust_score,
    status = EXCLUDED.status,
    updated_at = now();

-- Zoho Books Standard -------------------------------------------------------
INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url,
    price_minor, shipping_minor, currency, availability, condition, last_checked_at
)
SELECT m.id, p.id, 'zoho-books-standard',
       'https://www.zoho.com/us/books/pricing/pricing-comparison.html?highlight=standard',
       2000, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug = 'zoho' AND p.slug = 'zoho-books-standard'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    product_url = EXCLUDED.product_url,
    price_minor = EXCLUDED.price_minor,
    currency = EXCLUDED.currency,
    availability = EXCLUDED.availability,
    last_checked_at = EXCLUDED.last_checked_at,
    updated_at = now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference,
    disclosure_label, is_active
)
SELECT o.id, 'zoho', 'https://go.zoho.com/K0nf', 'PE2263909',
       'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id = o.merchant_id
WHERE m.slug = 'zoho' AND o.merchant_sku = 'zoho-books-standard'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url = EXCLUDED.destination_url,
    external_reference = EXCLUDED.external_reference,
    is_active = EXCLUDED.is_active,
    updated_at = now();

-- Zoho CRM Standard ---------------------------------------------------------
INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url,
    price_minor, shipping_minor, currency, availability, condition, last_checked_at
)
SELECT m.id, p.id, 'zoho-crm-standard', 'https://www.zoho.com/crm/zohocrm-pricing.html',
       2000, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug = 'zoho' AND p.slug = 'zoho-crm-standard'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    product_url = EXCLUDED.product_url, price_minor = EXCLUDED.price_minor,
    last_checked_at = EXCLUDED.last_checked_at, updated_at = now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference, disclosure_label, is_active
)
SELECT o.id, 'zoho', 'https://go.zoho.com/dNbV', 'PE2263909', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id = o.merchant_id
WHERE m.slug = 'zoho' AND o.merchant_sku = 'zoho-crm-standard'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url = EXCLUDED.destination_url, is_active = EXCLUDED.is_active, updated_at = now();

-- Bigin Express -------------------------------------------------------------
INSERT INTO commerce.merchant_offers (
    merchant_id, product_id, merchant_sku, product_url,
    price_minor, shipping_minor, currency, availability, condition, last_checked_at
)
SELECT m.id, p.id, 'bigin-express', 'https://www.bigin.com/pricing.html',
       900, 0, 'USD', 'in_stock', 'new', now()
FROM commerce.merchants m, catalog.products p
WHERE m.slug = 'zoho' AND p.slug = 'bigin-express'
ON CONFLICT (merchant_id, merchant_sku) DO UPDATE SET
    product_url = EXCLUDED.product_url, price_minor = EXCLUDED.price_minor,
    last_checked_at = EXCLUDED.last_checked_at, updated_at = now();

INSERT INTO commerce.affiliate_links (
    merchant_offer_id, provider, destination_url, external_reference, disclosure_label, is_active
)
SELECT o.id, 'zoho', 'https://go.zoho.com/SsgT', 'PE2263909', 'Affiliate link', true
FROM commerce.merchant_offers o
JOIN commerce.merchants m ON m.id = o.merchant_id
WHERE m.slug = 'zoho' AND o.merchant_sku = 'bigin-express'
ON CONFLICT (merchant_offer_id, provider) DO UPDATE SET
    destination_url = EXCLUDED.destination_url, is_active = EXCLUDED.is_active, updated_at = now();
