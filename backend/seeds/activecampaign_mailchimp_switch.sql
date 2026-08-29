-- The ActiveCampaign MailChimp Switch link, and the article that carries it.
--
-- The link was activated in the PartnerStack dashboard and copied from its
-- copy control on 2026-08-29. The account owner confirmed the pairing:
--   am4yesxqhxo9-c8qk4  MailChimp Switch   <- this file
--   yb5i7jsind0c-txqy7s CRM                <- held; see AFFILIATE_PROGRAMS.md
--
-- ---------------------------------------------------------------------------
-- Why this is a promotion and not a merchant offer.
--
-- 000026's comment says promotions are "editorial acquisition offers that are
-- not merchant offers for a catalog product", and this one does sell a catalog
-- product, so the fit deserves stating rather than assuming.
--
-- Two reasons it belongs here anyway. First, mechanically: the affiliate offer
-- path is already taken. commerce.affiliate_links is unique on
-- (merchant_offer_id, provider) and ResolveOfferDestination returns one row per
-- offer, so activecampaign-starter already serves the pricing link and cannot
-- serve a second. Second, and the real reason: this destination is not that
-- plan. It is a vendor-level acquisition page for people leaving Mailchimp,
-- tied to no tier and no price, and attaching it to activecampaign-starter
-- would claim it sells a 15 USD plan that the page never mentions.
--
-- The property the promotions table actually guarantees is the one that
-- matters here: it has no product_id, so nothing about this link can reach
-- recommendation scoring. That is true whether or not the destination happens
-- to sell something we also rank.
--
-- promotion_type is 'lead'. The page ends in a trial signup, not a checkout,
-- and calling it 'purchase' would misreport what a click is worth.
--
-- public_url is https://www.activecampaign.com/compare/mailchimp, read on
-- 2026-08-29 and returning 200 with the title "ActiveCampaign vs. Mailchimp:
-- Which Should You Choose in 2026?". It is the unaffiliated equivalent of the
-- destination, recorded so anyone can see where the link goes without
-- following it and creating a click.

BEGIN;

INSERT INTO commerce.merchants (
    name, slug, website_url, country_code, trust_score, status
) VALUES (
    'ActiveCampaign','activecampaign','https://www.activecampaign.com','US',80,'active'
)
ON CONFLICT (slug) DO UPDATE SET
    website_url=EXCLUDED.website_url, updated_at=now();

INSERT INTO commerce.affiliate_promotions (
    merchant_id, name, slug, promotion_type, public_url, destination_url,
    provider, external_reference, disclosure_label, is_active, last_checked_at
)
SELECT merchants.id,
       'ActiveCampaign for people leaving Mailchimp',
       'activecampaign-mailchimp-switch',
       'lead',
       'https://www.activecampaign.com/compare/mailchimp',
       'https://try.activecampaign.com/am4yesxqhxo9-c8qk4',
       'partnerstack', 'am4yesxqhxo9-c8qk4', 'Affiliate link', true, now()
FROM commerce.merchants merchants
WHERE merchants.slug = 'activecampaign'
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

-- ------------------------------------------------------- the placement ---
--
-- The block goes immediately after "Which one, in one line each", which is
-- where the article has just told an automation-driven reader that
-- ActiveCampaign is their answer. Anywhere later is after the section on
-- staying with Mailchimp, where a switch button would be arguing against the
-- paragraph above it.
--
-- The insert is positional, so it asserts the shape it is inserting into
-- first. If the article has been re-edited, this fails loudly rather than
-- dropping a sponsored link into the middle of a sentence it was not written
-- for.
DO $$
DECLARE
    block jsonb := jsonb_build_object(
        'type', 'cta',
        'heading', 'If automation is why you are leaving',
        'text', 'ActiveCampaign is the one tool here that is genuinely deeper than Mailchimp on automation, and the only one with no free tier. Their own Mailchimp comparison is the honest place to start, because it is where they have to state the trade rather than the pitch.',
        'label', 'See ActiveCampaign against Mailchimp',
        'promotion', 'activecampaign-mailchimp-switch'
    );
    current_content jsonb;
BEGIN
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = 'mailchimp-alternatives';

    IF current_content IS NULL THEN
        RAISE EXCEPTION 'mailchimp-alternatives entry is missing';
    END IF;

    IF current_content @> '[{"type":"cta","promotion":"activecampaign-mailchimp-switch"}]'::jsonb THEN
        RAISE NOTICE 'CTA already present; leaving the article unchanged';
        RETURN;
    END IF;

    IF current_content -> 4 ->> 'heading' IS DISTINCT FROM 'Which one, in one line each'
       OR current_content -> 5 ->> 'type' IS DISTINCT FROM 'unordered_list' THEN
        RAISE EXCEPTION
            'mailchimp-alternatives has been re-edited: expected the "Which one" list at index 5, found %',
            current_content -> 5 ->> 'type';
    END IF;

    UPDATE editorial.entries
    SET content = jsonb_insert(current_content, '{6}', block, true),
        updated_at = now()
    WHERE slug = 'mailchimp-alternatives';
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM commerce.affiliate_promotions
        WHERE slug = 'activecampaign-mailchimp-switch'
          AND is_active
          AND destination_url = 'https://try.activecampaign.com/am4yesxqhxo9-c8qk4'
    ) THEN
        RAISE EXCEPTION 'ActiveCampaign MailChimp Switch promotion activation failed';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM editorial.entries
        WHERE slug = 'mailchimp-alternatives'
          AND content @> '[{"type":"cta","promotion":"activecampaign-mailchimp-switch"}]'::jsonb
    ) THEN
        RAISE EXCEPTION 'mailchimp-alternatives did not receive the CTA block';
    END IF;
END $$;

COMMIT;
