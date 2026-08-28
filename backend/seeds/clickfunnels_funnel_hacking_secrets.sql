-- ClickFunnels Funnel Hacking Secrets affiliate promotion.
--
-- The two destinations were copied from the approved affiliate dashboard by
-- the account owner on 2026-08-26. They are intentionally promotions, not
-- catalog merchant offers: the free training and its order form must not be
-- represented as the standard ClickFunnels software subscription.

INSERT INTO commerce.merchants (
    name, slug, website_url, country_code, trust_score, status
) VALUES (
    'ClickFunnels', 'clickfunnels', 'https://www.clickfunnels.com', 'US', 80, 'active'
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name,
    website_url = EXCLUDED.website_url,
    country_code = EXCLUDED.country_code,
    status = EXCLUDED.status,
    updated_at = now();

INSERT INTO commerce.affiliate_promotions (
    merchant_id, name, slug, promotion_type, public_url, destination_url,
    provider, external_reference, disclosure_label, is_active, last_checked_at
)
SELECT merchants.id, promotion.name, promotion.slug, promotion.promotion_type,
       promotion.public_url, promotion.destination_url, 'clickfunnels',
       '4330879', 'Affiliate link', true, now()
FROM commerce.merchants merchants
CROSS JOIN (VALUES
    (
        'Funnel Hacking Secrets free training',
        'funnel-hacking-secrets-webinar',
        'lead',
        'https://www.funnelhackingsecrets.com',
        'https://www.funnelhackingsecrets.com?cf_affiliate_id=4330879&affiliate_id=4330879'
    ),
    (
        'Funnel Hacking Secrets current package',
        'funnel-hacking-secrets-order',
        'purchase',
        'https://www.funnelhackingsecrets.com/go',
        'https://www.funnelhackingsecrets.com/go?cf_affiliate_id=4330879&affiliate_id=4330879'
    )
) AS promotion(name, slug, promotion_type, public_url, destination_url)
WHERE merchants.slug = 'clickfunnels'
ON CONFLICT (slug) DO UPDATE SET
    merchant_id = EXCLUDED.merchant_id,
    name = EXCLUDED.name,
    promotion_type = EXCLUDED.promotion_type,
    public_url = EXCLUDED.public_url,
    destination_url = EXCLUDED.destination_url,
    provider = EXCLUDED.provider,
    external_reference = EXCLUDED.external_reference,
    disclosure_label = EXCLUDED.disclosure_label,
    is_active = true,
    last_checked_at = EXCLUDED.last_checked_at,
    updated_at = now();

DO $$
BEGIN
    IF (
        SELECT count(*)
        FROM commerce.affiliate_promotions
        WHERE slug IN ('funnel-hacking-secrets-webinar', 'funnel-hacking-secrets-order')
          AND is_active
          AND external_reference = '4330879'
    ) <> 2 THEN
        RAISE EXCEPTION 'ClickFunnels promotion activation failed';
    END IF;
END $$;
