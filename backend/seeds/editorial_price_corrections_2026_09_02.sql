-- Editorial price corrections, 2026-09-02.
--
-- The billing basis audit of the same date moved ten compared prices to their
-- monthly-billing figure. Seven published pieces state those prices in their
-- own prose, where no schema can keep them honest, so each sentence is
-- corrected here to match the catalog it is written about.
--
-- What changed, and why it is not only a number: three of these sentences
-- rested on the old figures being close. "Shopify Basic and BigCommerce Core
-- both cost 29 USD a month" and "the price being nearly identical is a
-- coincidence" were arguments about a coincidence that has stopped being one,
-- so those sentences are rewritten rather than renumbered. Where a piece
-- already gave both figures it now leads with the monthly one, because that
-- is the figure the catalog compares.
--
-- Each block asserts the exact sentence it expects before replacing it. A
-- piece edited since this file was written raises instead of being
-- overwritten, and a second run is a no-op.
--
-- Applied with psql:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f backend/seeds/editorial_price_corrections_2026_09_02.sql

BEGIN;

-- shopify-vs-bigcommerce · {0,text}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{0,text}' INTO current_value
    FROM editorial.entries WHERE slug = 'shopify-vs-bigcommerce';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'shopify-vs-bigcommerce: no value at {0,text}';
    END IF;
    IF current_value = 'Shopify Basic is 29 USD a month and BigCommerce Core is 39, and the gap closes to nothing if you pay either of them annually. They are not the same money in a second sense either. BigCommerce charges no platform transaction fee on top when you use a supported payment provider. Shopify Basic charges 2% on every order taken through a payment provider other than Shopify Payments, on top of the card rate you are already paying. On any real volume that line matters more than the subscription.' THEN
        RAISE NOTICE 'shopify-vs-bigcommerce {0,text}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Shopify Basic and BigCommerce Core both cost 29 USD a month. They are not the same 29 dollars. BigCommerce charges no platform transaction fee on top when you use a supported payment provider. Shopify Basic charges 2% on every order taken through a payment provider other than Shopify Payments, on top of the card rate you are already paying. On any real volume that line matters more than the subscription.' THEN
        RAISE EXCEPTION 'shopify-vs-bigcommerce {0,text} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{0,text}', to_jsonb('Shopify Basic is 29 USD a month and BigCommerce Core is 39, and the gap closes to nothing if you pay either of them annually. They are not the same money in a second sense either. BigCommerce charges no platform transaction fee on top when you use a supported payment provider. Shopify Basic charges 2% on every order taken through a payment provider other than Shopify Payments, on top of the card rate you are already paying. On any real volume that line matters more than the subscription.'::text)),
        updated_at = now()
    WHERE slug = 'shopify-vs-bigcommerce';
END $$;
-- shopify-vs-bigcommerce · {2,items,2}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,2}' INTO current_value
    FROM editorial.entries WHERE slug = 'shopify-vs-bigcommerce';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'shopify-vs-bigcommerce: no value at {2,items,2}';
    END IF;
    IF current_value = 'BigCommerce Core: 39 USD per month at monthly billing, 29 billed annually. Capped at 30,000 USD of trailing annual sales.' THEN
        RAISE NOTICE 'shopify-vs-bigcommerce {2,items,2}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'BigCommerce Core: 29 USD per month, billed annually. Capped at 30,000 USD of trailing annual sales.' THEN
        RAISE EXCEPTION 'shopify-vs-bigcommerce {2,items,2} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,2}', to_jsonb('BigCommerce Core: 39 USD per month at monthly billing, 29 billed annually. Capped at 30,000 USD of trailing annual sales.'::text)),
        updated_at = now()
    WHERE slug = 'shopify-vs-bigcommerce';
END $$;
-- canva-vs-figma · {0,text}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{0,text}' INTO current_value
    FROM editorial.entries WHERE slug = 'canva-vs-figma';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'canva-vs-figma: no value at {0,text}';
    END IF;
    IF current_value = 'Canva Pro works out at 15 USD a month, Figma Professional is 20 on monthly billing and 16 if you pay for the year. People compare them constantly and almost nobody should. They do different jobs, and any resemblance in the price is a coincidence rather than a signal.' THEN
        RAISE NOTICE 'canva-vs-figma {0,text}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Canva Pro works out at 15 USD a month, Figma Professional is 16. People compare them constantly and almost nobody should. They do different jobs, and the price being nearly identical is a coincidence rather than a signal.' THEN
        RAISE EXCEPTION 'canva-vs-figma {0,text} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{0,text}', to_jsonb('Canva Pro works out at 15 USD a month, Figma Professional is 20 on monthly billing and 16 if you pay for the year. People compare them constantly and almost nobody should. They do different jobs, and any resemblance in the price is a coincidence rather than a signal.'::text)),
        updated_at = now()
    WHERE slug = 'canva-vs-figma';
END $$;
-- canva-vs-figma · {2,items,0}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,0}' INTO current_value
    FROM editorial.entries WHERE slug = 'canva-vs-figma';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'canva-vs-figma: no value at {2,items,0}';
    END IF;
    IF current_value = 'Sketch Standard: 14 USD per editor per month at monthly billing, 12 billed yearly. Mac only, and that single fact decides more choices here than any feature.' THEN
        RAISE NOTICE 'canva-vs-figma {2,items,0}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Sketch Standard: 12 USD per editor per month billed yearly. Mac only, and that single fact decides more choices here than any feature.' THEN
        RAISE EXCEPTION 'canva-vs-figma {2,items,0} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,0}', to_jsonb('Sketch Standard: 14 USD per editor per month at monthly billing, 12 billed yearly. Mac only, and that single fact decides more choices here than any feature.'::text)),
        updated_at = now()
    WHERE slug = 'canva-vs-figma';
END $$;
-- canva-vs-figma · {2,items,2}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,2}' INTO current_value
    FROM editorial.entries WHERE slug = 'canva-vs-figma';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'canva-vs-figma: no value at {2,items,2}';
    END IF;
    IF current_value = 'Figma Professional: 20 USD per full seat per month at monthly billing, 16 billed annually.' THEN
        RAISE NOTICE 'canva-vs-figma {2,items,2}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Figma Professional: 16 USD per full seat per month billed annually. Monthly billing is 20.' THEN
        RAISE EXCEPTION 'canva-vs-figma {2,items,2} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,2}', to_jsonb('Figma Professional: 20 USD per full seat per month at monthly billing, 16 billed annually.'::text)),
        updated_at = now()
    WHERE slug = 'canva-vs-figma';
END $$;
-- webflow-vs-squarespace-vs-framer · {2,items,2}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,2}' INTO current_value
    FROM editorial.entries WHERE slug = 'webflow-vs-squarespace-vs-framer';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'webflow-vs-squarespace-vs-framer: no value at {2,items,2}';
    END IF;
    IF current_value = 'Squarespace Basic: 25 USD per month at monthly billing, 19 billed annually, with hosting, domain and payments in the one bill.' THEN
        RAISE NOTICE 'webflow-vs-squarespace-vs-framer {2,items,2}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Squarespace Basic: 19 USD per month at monthly billing, with hosting, domain and payments in the one bill.' THEN
        RAISE EXCEPTION 'webflow-vs-squarespace-vs-framer {2,items,2} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,2}', to_jsonb('Squarespace Basic: 25 USD per month at monthly billing, 19 billed annually, with hosting, domain and payments in the one bill.'::text)),
        updated_at = now()
    WHERE slug = 'webflow-vs-squarespace-vs-framer';
END $$;
-- teachable-vs-thinkific-vs-gumroad · {0,text}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{0,text}' INTO current_value
    FROM editorial.entries WHERE slug = 'teachable-vs-thinkific-vs-gumroad';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'teachable-vs-thinkific-vs-gumroad: no value at {0,text}';
    END IF;
    IF current_value = 'You cannot compare these on the monthly price, and the monthly price is what every other comparison shows you. Gumroad is free and takes 10%. Teachable is 39 and takes 7.5%. Thinkific is 40 and takes nothing. At some volume each of them is the cheapest.' THEN
        RAISE NOTICE 'teachable-vs-thinkific-vs-gumroad {0,text}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'You cannot compare these on the monthly price, and the monthly price is what every other comparison shows you. Gumroad is free and takes 10%. Teachable is 29 and takes 7.5%. Thinkific is 40 and takes nothing. At some volume each of them is the cheapest.' THEN
        RAISE EXCEPTION 'teachable-vs-thinkific-vs-gumroad {0,text} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{0,text}', to_jsonb('You cannot compare these on the monthly price, and the monthly price is what every other comparison shows you. Gumroad is free and takes 10%. Teachable is 39 and takes 7.5%. Thinkific is 40 and takes nothing. At some volume each of them is the cheapest.'::text)),
        updated_at = now()
    WHERE slug = 'teachable-vs-thinkific-vs-gumroad';
END $$;
-- teachable-vs-thinkific-vs-gumroad · {2,items,1}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,1}' INTO current_value
    FROM editorial.entries WHERE slug = 'teachable-vs-thinkific-vs-gumroad';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'teachable-vs-thinkific-vs-gumroad: no value at {2,items,1}';
    END IF;
    IF current_value = 'Teachable Starter: 39 USD a month at monthly billing, 29 billed annually, plus a 7.5% transaction fee that the higher tiers drop.' THEN
        RAISE NOTICE 'teachable-vs-thinkific-vs-gumroad {2,items,1}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Teachable Starter: 29 USD a month billed annually, 39 billed monthly, plus a 7.5% transaction fee that the higher tiers drop.' THEN
        RAISE EXCEPTION 'teachable-vs-thinkific-vs-gumroad {2,items,1} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,1}', to_jsonb('Teachable Starter: 39 USD a month at monthly billing, 29 billed annually, plus a 7.5% transaction fee that the higher tiers drop.'::text)),
        updated_at = now()
    WHERE slug = 'teachable-vs-thinkific-vs-gumroad';
END $$;
-- zapier-vs-make · {2,items,1}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,1}' INTO current_value
    FROM editorial.entries WHERE slug = 'zapier-vs-make';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'zapier-vs-make: no value at {2,items,1}';
    END IF;
    IF current_value = 'Zapier Professional: 29.99 USD per month at monthly billing, 19.99 billed annually. Priced by task volume, so it rises with use.' THEN
        RAISE NOTICE 'zapier-vs-make {2,items,1}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Zapier Professional: 19.99 USD per month, billed annually. Priced by task volume, so it rises with use.' THEN
        RAISE EXCEPTION 'zapier-vs-make {2,items,1} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,1}', to_jsonb('Zapier Professional: 29.99 USD per month at monthly billing, 19.99 billed annually. Priced by task volume, so it rises with use.'::text)),
        updated_at = now()
    WHERE slug = 'zapier-vs-make';
END $$;
-- ahrefs-vs-semrush · {0,text}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{0,text}' INTO current_value
    FROM editorial.entries WHERE slug = 'ahrefs-vs-semrush';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'ahrefs-vs-semrush: no value at {0,text}';
    END IF;
    IF current_value = 'Ahrefs Starter is 29 USD a month. Semrush is 139. That is not a comparison between two similar things — it is the shape of the whole market, where there is one cheap way in and then a cliff.' THEN
        RAISE NOTICE 'ahrefs-vs-semrush {0,text}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Ahrefs Starter is 29 USD a month. Semrush is 117.33 billed annually. That is not a comparison between two similar things — it is the shape of the whole market, where there is one cheap way in and then a cliff.' THEN
        RAISE EXCEPTION 'ahrefs-vs-semrush {0,text} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{0,text}', to_jsonb('Ahrefs Starter is 29 USD a month. Semrush is 139. That is not a comparison between two similar things — it is the shape of the whole market, where there is one cheap way in and then a cliff.'::text)),
        updated_at = now()
    WHERE slug = 'ahrefs-vs-semrush';
END $$;
-- ahrefs-vs-semrush · {2,items,1}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,1}' INTO current_value
    FROM editorial.entries WHERE slug = 'ahrefs-vs-semrush';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'ahrefs-vs-semrush: no value at {2,items,1}';
    END IF;
    IF current_value = 'SE Ranking Core: 129 USD per month at monthly billing, 103.20 billed annually. Ten projects and a manager seat.' THEN
        RAISE NOTICE 'ahrefs-vs-semrush {2,items,1}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'SE Ranking Core: 103.20 USD per month billed annually; 129 monthly. Ten projects and a manager seat.' THEN
        RAISE EXCEPTION 'ahrefs-vs-semrush {2,items,1} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,1}', to_jsonb('SE Ranking Core: 129 USD per month at monthly billing, 103.20 billed annually. Ten projects and a manager seat.'::text)),
        updated_at = now()
    WHERE slug = 'ahrefs-vs-semrush';
END $$;
-- ahrefs-vs-semrush · {2,items,2}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{2,items,2}' INTO current_value
    FROM editorial.entries WHERE slug = 'ahrefs-vs-semrush';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'ahrefs-vs-semrush: no value at {2,items,2}';
    END IF;
    IF current_value = 'Semrush SEO: 139 USD per month at monthly billing, 117.33 billed annually.' THEN
        RAISE NOTICE 'ahrefs-vs-semrush {2,items,2}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'Semrush SEO: 117.33 USD per month billed annually; 139 monthly.' THEN
        RAISE EXCEPTION 'ahrefs-vs-semrush {2,items,2} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{2,items,2}', to_jsonb('Semrush SEO: 139 USD per month at monthly billing, 117.33 billed annually.'::text)),
        updated_at = now()
    WHERE slug = 'ahrefs-vs-semrush';
END $$;
-- agency-3-people-under-150 · {11,items,2}
DO $$
DECLARE current_value text;
BEGIN
    SELECT content #>> '{11,items,2}' INTO current_value
    FROM editorial.entries WHERE slug = 'agency-3-people-under-150';
    IF current_value IS NULL THEN
        RAISE EXCEPTION 'agency-3-people-under-150: no value at {11,items,2}';
    END IF;
    IF current_value = 'monday.com Basic (12 USD per seat per month on monthly billing, 9 billed annually, three-seat minimum) and ClickUp Unlimited (10 USD per user per month). Both cost more than double Zoho Projects for the same three questions, and the catalog notes that monday''s Basic tier omits the timeline and calendar views most teams end up wanting. The guide''s test applies: if a tool needs someone to maintain it and nobody is paid to, it is abandoned within two quarters, so extra surface is a cost, not a feature.' THEN
        RAISE NOTICE 'agency-3-people-under-150 {11,items,2}: already corrected';
        RETURN;
    END IF;
    IF current_value <> 'monday.com Basic (9 USD per seat per month billed annually, three-seat minimum) and ClickUp Unlimited (10 USD per user per month). Both cost more than double Zoho Projects for the same three questions, and the catalog notes that monday''s Basic tier omits the timeline and calendar views most teams end up wanting. The guide''s test applies: if a tool needs someone to maintain it and nobody is paid to, it is abandoned within two quarters, so extra surface is a cost, not a feature.' THEN
        RAISE EXCEPTION 'agency-3-people-under-150 {11,items,2} has been re-edited since this correction was written; refusing to overwrite. Found: %', left(current_value, 160);
    END IF;
    UPDATE editorial.entries
    SET content = jsonb_set(content, '{11,items,2}', to_jsonb('monday.com Basic (12 USD per seat per month on monthly billing, 9 billed annually, three-seat minimum) and ClickUp Unlimited (10 USD per user per month). Both cost more than double Zoho Projects for the same three questions, and the catalog notes that monday''s Basic tier omits the timeline and calendar views most teams end up wanting. The guide''s test applies: if a tool needs someone to maintain it and nobody is paid to, it is abandoned within two quarters, so extra surface is a cost, not a feature.'::text)),
        updated_at = now()
    WHERE slug = 'agency-3-people-under-150';
END $$;
-- No published piece still quotes a superseded figure for the ten products
-- whose price moved today.
DO $$
DECLARE stale integer;
BEGIN
    SELECT count(*) INTO stale
    FROM editorial.entries entries
    WHERE entries.status = 'published'
      AND (entries.content::text LIKE '%BigCommerce Core: 29 USD%'
        OR entries.content::text LIKE '%Figma Professional is 16%'
        OR entries.content::text LIKE '%Sketch Standard: 12 USD%'
        OR entries.content::text LIKE '%Squarespace Basic: 19 USD%'
        OR entries.content::text LIKE '%Teachable is 29%'
        OR entries.content::text LIKE '%Zapier Professional: 19.99 USD per month, billed annually%'
        OR entries.content::text LIKE '%Semrush is 117.33 billed annually%'
        OR entries.content::text LIKE '%monday.com Basic (9 USD%');
    IF stale > 0 THEN
        RAISE EXCEPTION 'a superseded price is still published in % piece(s)', stale;
    END IF;
    RAISE NOTICE 'no superseded prices remain in published editorial';
END $$;

COMMIT;
