-- Course-platform price re-read, 2026-08-28.
--
-- Prompted by adding a Teachable affiliate link: before putting a commission
-- behind a price, the price was read again rather than trusted from 2026-08-21.
--
-- WHAT WAS CONFIRMED — Teachable Starter, read from the rendered pricing page:
--
--   monthly billing   39.00 USD/mo   (the tab the page opens on)
--   annual billing    29.00 USD/mo   (39.00 struck through, "Save 22%")
--   transaction fee   7.5% on Starter, 0% on Builder and above
--
-- The catalog's recorded 2900 is therefore still the correct ANNUAL figure and
-- is not stale. It is left unchanged. See the note below on why it is not
-- moved to the monthly figure.
--
-- WHAT CONTRADICTS ITSELF — Teachable's own structured data disagrees with its
-- own page. The JSON-LD on the same URL declares "Starter Plan - Monthly" at
-- 59 USD and "Starter Plan - Yearly" at 39 USD, while the rendered price block
-- shows 39 and 29. The rendered figures are what a customer is charged and
-- what the signup links carry (starter-usd-monthly, starter-usd-yearly), so
-- those are recorded. This is worth keeping written down: anything that reads
-- Teachable's markup rather than its page — a search engine, a competitor's
-- scraper, a future importer here — will get numbers no customer ever pays.
--
-- WHAT COULD NOT BE READ — Thinkific Basic in USD. The pricing page geolocates
-- and served EUR only: 53 EUR/mo billed monthly, 39 EUR/mo billed annually.
-- The catalog records 4000, which is consistent with the annual rate having
-- been read in USD on 2026-08-21, but the USD monthly rate is unknown. No USD
-- figure is invented here. Re-read it through a US connection to close this.
--
-- WHY NEITHER PRICE MOVES TO ITS MONTHLY BASIS
--
-- Both products are recorded at annual rates, so within the course-platform
-- category the comparison a visitor sees is already like for like, and the two
-- descriptions compare them to each other explicitly. Moving Teachable alone
-- to 39.00 while Thinkific stays at its annual 40.00 would put them a dollar
-- apart on screen while the real monthly gap is far wider — a correct
-- comparison replaced by a misleading one, in the name of a rule.
--
-- The rule is right and the catalog is inconsistent with it; the fix is not
-- one product at a time. catalog.products has price_minor and currency and no
-- column for the billing basis, so today that basis exists only in prose in
-- the description. Twenty-five published products state an annual or yearly
-- basis in their description text while the recommendation engine compares
-- their price_minor figures as though all were alike. That is the defect, and
-- it needs a schema change plus a policy revision, not a price edit.

INSERT INTO evidence.sources (
    external_key, source_type, title, publisher, source_url,
    is_fictional, review_status, reviewed_at, review_note
) VALUES (
    'teachable-pricing-2026-08-28','manufacturer_documentation',
    'Teachable pricing','Teachable','https://teachable.com/pricing',
    false,'verified',now(),
    'Re-read 2026-08-28 from the rendered pricing page. Starter is 39.00 USD per month at monthly billing and 29.00 USD per month billed annually, the latter shown against a struck-through 39.00 with a 22% saving. The 7.5% Starter transaction fee is stated on the same card and is dropped by Builder and above. Recorded from the rendered price block and corroborated by the signup links starter-usd-monthly and starter-usd-yearly. The page''s own JSON-LD disagrees, declaring the monthly plan at 59 USD and the yearly at 39 USD; those figures match no control on the page and are not recorded.'
),
(
    'thinkific-pricing-2026-08-28','manufacturer_documentation',
    'Thinkific pricing','Thinkific','https://www.thinkific.com/pricing/',
    false,'pending',NULL,
    'Attempted 2026-08-28 and NOT usable for a USD figure. The page geolocates and served EUR only: Basic at 53 EUR per month billed monthly and 39 EUR per month billed annually. The catalog''s 40.00 USD is consistent with the annual rate read on 2026-08-21, but the USD monthly rate remains unread. Left at review_status pending so this gap is visible rather than assumed closed, which is also why it carries no review timestamp. Re-read through a US connection.'
)
ON CONFLICT (external_key) DO UPDATE SET
    review_status=EXCLUDED.review_status,
    -- A pending source must carry no review timestamp, and a verified one
    -- must carry one: evidence.sources enforces
    -- CHECK ((review_status = 'pending') = (reviewed_at IS NULL)).
    reviewed_at=CASE
        WHEN EXCLUDED.review_status = 'pending' THEN NULL
        ELSE COALESCE(evidence.sources.reviewed_at, now())
    END,
    source_url=EXCLUDED.source_url,
    review_note=EXCLUDED.review_note,
    updated_at=now();

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT sources.id, products.id, now(), mapping.confidence, mapping.note
FROM (VALUES
 ('teachable-starter','teachable-pricing-2026-08-28',100,
  'Both billing rates read directly from the rendered pricing page on 2026-08-28. The recorded product price of 29.00 USD is the annual rate; monthly is 39.00 USD. The 7.5% Starter transaction fee still applies and is not included in the figure.'),
 ('thinkific-basic','thinkific-pricing-2026-08-28',40,
  'Read on 2026-08-28 but in EUR only, because the vendor page geolocates. Basic is 53 EUR monthly and 39 EUR billed annually. The recorded 40.00 USD is the annual rate carried forward from 2026-08-21 and is not re-confirmed here; low confidence records that.')
) AS mapping(product_slug, source_key, confidence, note)
JOIN catalog.products products ON products.slug = mapping.product_slug
JOIN evidence.sources sources ON sources.external_key = mapping.source_key;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM evidence.sources
        WHERE external_key = 'teachable-pricing-2026-08-28'
          AND review_status = 'verified'
    ) THEN
        RAISE EXCEPTION 'Teachable 2026-08-28 price re-read was not recorded';
    END IF;
END $$;
