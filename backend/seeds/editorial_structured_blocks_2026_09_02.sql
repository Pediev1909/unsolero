-- Structured blocks for the three pieces whose products carry live affiliate
-- offers, applied 2026-09-02: an `offer` row for each product with a servable
-- offer, and a short FAQ that repeats, in question form, what each piece
-- already says.
--
-- Applied with the other seeds, by psql against the application database,
-- after the editorial seeds that create these entries:
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 \
--        -f backend/seeds/editorial_structured_blocks_2026_09_02.sql
--
-- ---------------------------------------------------------------------------
-- What the blocks are.
--
-- `offer` names a catalog product by slug and nothing else that could become
-- a URL. The price, the date it was read and the destination are resolved
-- when the page renders, from the product's live offer, through the same
-- checks the product page applies. So a block can only point at what the
-- catalog already serves, and when the offer goes dark the block degrades to
-- a plain catalog link rather than a dead button.
--
-- Only products with a live offer on 2026-09-02 get a row, verified against
-- the production API that day: zoho-campaigns-standard, mailerlite-comfort
-- and kit-creator (mailchimp-alternatives; ActiveCampaign already has a
-- promotion CTA and keeps it), zoho-bookings-basic and cal-com-teams
-- (calendly-alternatives), zoho-crm-standard and bigin-express
-- (zoho-crm-vs-hubspot). Brevo, Calendly and HubSpot have none and get none;
-- where that could read as a slant the first offer block says so.
--
-- `faq` answers are paraphrases of claims already present in that entry's
-- blocks, or of its catalog prices as read from the vendor pages on
-- 2026-08-26 (2026-08-29 for ActiveCampaign). Nothing below is a new claim
-- about any product.
--
-- ---------------------------------------------------------------------------
-- How the blocks are placed.
--
-- By heading TEXT, never by fixed index. Each block finds its anchor heading
-- with jsonb_array_elements WITH ORDINALITY and raises if the heading is
-- gone, so a re-edited article fails loudly instead of receiving a vendor row
-- in the middle of an argument it was not written for. Every block this file
-- places is stripped before it is placed again, so running the file twice
-- leaves each entry identical to running it once.
--
-- activecampaign_mailchimp_switch.sql asserts its CTA at fixed indexes 4–6
-- of mailchimp-alternatives. It runs before this file on a fresh database
-- and passes. Re-running it AFTER this file raises, by its own design: the
-- article has been re-edited above index 7, which is exactly the case it
-- refuses. That is expected and is not a fault in either file.

BEGIN;

-- ================================================ mailchimp-alternatives ===
--
-- The three rows go directly after the "Which one, in one line each" list,
-- which is the moment the piece has told each kind of reader their answer,
-- and before the ActiveCampaign CTA that closes the section. Anywhere later
-- is under "When you should stay on Mailchimp", arguing against the heading.
DO $$
DECLARE
    entry_slug     CONSTANT text := 'mailchimp-alternatives';
    anchor_heading CONSTANT text := 'Which one, in one line each';
    faq_heading    CONSTANT text := 'Questions people ask';
    offer_products CONSTANT text[] := ARRAY[
        'zoho-campaigns-standard', 'mailerlite-comfort', 'kit-creator'
    ];
    offers jsonb := jsonb_build_array(
        jsonb_build_object(
            'type', 'offer',
            'product', 'zoho-campaigns-standard',
            'heading', 'Where to get them',
            'text', 'The cheapest of the five at 5.25 USD per month, and it connects to Zoho CRM without any work. Brevo has no vendor link on this page, which changes nothing about the advice above.'
        ),
        jsonb_build_object(
            'type', 'offer',
            'product', 'mailerlite-comfort',
            'text', 'The best editor of the five, on monthly billing at 19 USD per month, with a free tier up to 250 subscribers.'
        ),
        jsonb_build_object(
            'type', 'offer',
            'product', 'kit-creator',
            'text', 'The most expensive here at 39 USD per month, built for people who publish, with a free tier that reaches 10,000 subscribers with features withheld.'
        )
    );
    faq jsonb := jsonb_build_object(
        'type', 'faq',
        'heading', faq_heading,
        'questions', jsonb_build_array(
            jsonb_build_object(
                'question', 'Which Mailchimp alternative is the cheapest?',
                'answer', 'Zoho Campaigns Standard, at 5.25 USD per month for 1,000 subscribers. Brevo Starter is 9 USD but prices by emails sent rather than contacts held, so for a list that is large and mostly dormant it is usually the bigger saving.'
            ),
            jsonb_build_object(
                'question', 'Should I leave Mailchimp to get better automation?',
                'answer', 'Only if you are going to ActiveCampaign. It is the one tool here that is genuinely deeper than Mailchimp on automation, and the only one with no free tier. The other four are not deeper, so swapping to a cheaper tool with less automation solves a bill, not the problem you had.'
            ),
            jsonb_build_object(
                'question', 'When is it not worth switching from Mailchimp?',
                'answer', 'When your list is under a few thousand and you send often. The price gap is then small enough that moving is not worth a weekend, and Mailchimp is still the easiest tool to hand to somebody non-technical.'
            ),
            jsonb_build_object(
                'question', 'What does moving off Mailchimp actually cost?',
                'answer', 'The subscriber export is easy and every tool here imports a CSV. The time goes into rebuilding automations, re-authenticating your sending domain, and the deliverability dip while the new provider warms up your reputation. Budget a fortnight of watching open rates before you decide the move failed.'
            )
        )
    );
    current_content jsonb;
    anchor_position integer;
    offer           jsonb;
    placed          integer := 0;
BEGIN
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = entry_slug;

    IF current_content IS NULL THEN
        RAISE EXCEPTION '% entry is missing', entry_slug;
    END IF;

    -- Strip every block this file places, then place each once.
    SELECT COALESCE(jsonb_agg(element ORDER BY position), '[]'::jsonb)
    INTO current_content
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE NOT (element ->> 'type' = 'offer'
               AND element ->> 'product' = ANY (offer_products))
      AND NOT (element ->> 'type' = 'faq'
               AND element ->> 'heading' = faq_heading);

    -- The anchor, by text. `position` is 1-based, so it is also the 0-based
    -- index of the block directly under the heading — the list.
    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'heading'
      AND element ->> 'heading' = anchor_heading;

    IF anchor_position IS NULL THEN
        RAISE EXCEPTION '% has been re-edited: heading "%" is missing',
            entry_slug, anchor_heading;
    END IF;
    IF current_content -> anchor_position ->> 'type' IS DISTINCT FROM 'unordered_list' THEN
        RAISE EXCEPTION '% has been re-edited: expected an unordered_list under "%", found %',
            entry_slug, anchor_heading, current_content -> anchor_position ->> 'type';
    END IF;

    -- insert_after = true at the list's index, then at each row just placed,
    -- so the rows keep the order they are written in above.
    FOR offer IN SELECT value FROM jsonb_array_elements(offers) LOOP
        current_content := jsonb_insert(
            current_content, ARRAY[(anchor_position + placed)::text], offer, true
        );
        placed := placed + 1;
    END LOOP;

    current_content := current_content || jsonb_build_array(faq);

    UPDATE editorial.entries
    SET content = current_content,
        updated_at = now()
    WHERE slug = entry_slug;
END $$;

-- ================================================= calendly-alternatives ===
--
-- Each row closes the section that argues for its product: Zoho Bookings
-- before the "Cal.com" heading, Cal.com before the "Why Calendly is still
-- winning" callout. The piece has no one-liner list, so the end of each
-- argument is the point where a reader has decided.
DO $$
DECLARE
    entry_slug      CONSTANT text := 'calendly-alternatives';
    calcom_heading  CONSTANT text := 'Cal.com, if you need to own the data';
    calendly_callout CONSTANT text := 'Why Calendly is still winning';
    faq_heading     CONSTANT text := 'Questions people ask';
    offer_products  CONSTANT text[] := ARRAY['zoho-bookings-basic', 'cal-com-teams'];
    zoho_offer jsonb := jsonb_build_object(
        'type', 'offer',
        'product', 'zoho-bookings-basic',
        'text', 'Zoho Bookings Basic is 8 USD per user per month with a free tier, and a booking lands against a CRM record that already exists. Worth it if you run Zoho; a bad trade if you do not.'
    );
    calcom_offer jsonb := jsonb_build_object(
        'type', 'offer',
        'product', 'cal-com-teams',
        'text', 'Cal.com Teams is 12 USD per user per month billed yearly, free forever for one person, and it can run on your own server so the booking record never leaves your infrastructure.'
    );
    faq jsonb := jsonb_build_object(
        'type', 'faq',
        'heading', faq_heading,
        'questions', jsonb_build_array(
            jsonb_build_object(
                'question', 'Is Zoho Bookings cheaper than Calendly?',
                'answer', 'By two dollars: Zoho Bookings Basic is 8 USD per user per month against 10 for Calendly Standard. The saving is not the point. The reason to move is that a booking lands against a Zoho CRM record that already exists and the invoice comes from the same place. If you do not run Zoho, you are learning an unfamiliar tool to save the price of a coffee.'
            ),
            jsonb_build_object(
                'question', 'Which Calendly alternative has the best free tier?',
                'answer', 'Cal.com, which is free forever for a single person, the strongest free tier in this category by some distance. A solo consultant who does not need team features can stop paying for scheduling entirely. Zoho Bookings also has a free tier.'
            ),
            jsonb_build_object(
                'question', 'When should I stay on Calendly?',
                'answer', 'When your bookings come from strangers: prospects, clients, people who found you online. Calendly is the name people recognise in an email, and recognition removes hesitation at the exact moment you are asking somebody to commit their time. Move only if you are inside Zoho already, or if the booking data has to live on your own machine.'
            )
        )
    );
    current_content jsonb;
    anchor_position integer;
BEGIN
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = entry_slug;

    IF current_content IS NULL THEN
        RAISE EXCEPTION '% entry is missing', entry_slug;
    END IF;

    SELECT COALESCE(jsonb_agg(element ORDER BY position), '[]'::jsonb)
    INTO current_content
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE NOT (element ->> 'type' = 'offer'
               AND element ->> 'product' = ANY (offer_products))
      AND NOT (element ->> 'type' = 'faq'
               AND element ->> 'heading' = faq_heading);

    -- Zoho Bookings: directly before the Cal.com heading. The block before
    -- that heading must be the section's own paragraph, so the row closes an
    -- argument rather than following a heading.
    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'heading'
      AND element ->> 'heading' = calcom_heading;

    IF anchor_position IS NULL THEN
        RAISE EXCEPTION '% has been re-edited: heading "%" is missing',
            entry_slug, calcom_heading;
    END IF;
    IF current_content -> (anchor_position - 2) ->> 'type' IS DISTINCT FROM 'paragraph' THEN
        RAISE EXCEPTION '% has been re-edited: expected a paragraph before "%", found %',
            entry_slug, calcom_heading, current_content -> (anchor_position - 2) ->> 'type';
    END IF;
    -- insert_after = false at the heading's 0-based index puts the row
    -- before it.
    current_content := jsonb_insert(
        current_content, ARRAY[(anchor_position - 1)::text], zoho_offer, false
    );

    -- Cal.com: directly before the callout, located again because the row
    -- above moved everything after it down by one.
    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'callout'
      AND element ->> 'heading' = calendly_callout;

    IF anchor_position IS NULL THEN
        RAISE EXCEPTION '% has been re-edited: callout "%" is missing',
            entry_slug, calendly_callout;
    END IF;
    IF current_content -> (anchor_position - 2) ->> 'type' IS DISTINCT FROM 'paragraph' THEN
        RAISE EXCEPTION '% has been re-edited: expected a paragraph before the "%" callout, found %',
            entry_slug, calendly_callout, current_content -> (anchor_position - 2) ->> 'type';
    END IF;
    current_content := jsonb_insert(
        current_content, ARRAY[(anchor_position - 1)::text], calcom_offer, false
    );

    current_content := current_content || jsonb_build_array(faq);

    UPDATE editorial.entries
    SET content = current_content,
        updated_at = now()
    WHERE slug = entry_slug;
END $$;

-- =================================================== zoho-crm-vs-hubspot ===
--
-- Both rows go directly after the "Where each one is genuinely better" list,
-- whose last line is the one that introduces Bigin. HubSpot has no live
-- offer and gets no row; the first row says so.
DO $$
DECLARE
    entry_slug     CONSTANT text := 'zoho-crm-vs-hubspot';
    anchor_heading CONSTANT text := 'Where each one is genuinely better';
    faq_heading    CONSTANT text := 'Questions people ask';
    offer_products CONSTANT text[] := ARRAY['zoho-crm-standard', 'bigin-express'];
    offers jsonb := jsonb_build_array(
        jsonb_build_object(
            'type', 'offer',
            'product', 'zoho-crm-standard',
            'heading', 'Where to get them',
            'text', 'Zoho CRM Standard is 20 USD per user per month, the same as HubSpot Starter today. Its tiers rise gently from there, and it is attached to a suite that covers invoicing, email marketing, help desk, projects and bookings at prices nobody else matches. HubSpot has no vendor link on this page, which changes nothing about the comparison above.'
        ),
        jsonb_build_object(
            'type', 'offer',
            'product', 'bigin-express',
            'text', 'Bigin Express, from the same company, is 9 USD per user per month. For a one-person business selling services it is closer to the size of the problem than either CRM above.'
        )
    );
    faq jsonb := jsonb_build_object(
        'type', 'faq',
        'heading', faq_heading,
        'questions', jsonb_build_array(
            jsonb_build_object(
                'question', 'Is Zoho CRM cheaper than HubSpot?',
                'answer', 'Not on the entry tier: Zoho CRM Standard and HubSpot Starter Customer Platform both cost 20 USD per user per month. The difference shows up at the step up, where the Zoho tiers rise gently and the next HubSpot tier is a different order of money, with real automation and proper reporting living above the Starter line rather than in it.'
            ),
            jsonb_build_object(
                'question', 'Which is easier to use, Zoho CRM or HubSpot?',
                'answer', 'HubSpot. The interface is calmer, the onboarding is genuinely good, and the reporting makes sense without a manual. Zoho CRM is denser and less pleasant, and its case rests on the suite around it rather than on the interface.'
            ),
            jsonb_build_object(
                'question', 'Is there a cheaper CRM than either for a one-person business?',
                'answer', 'Bigin, from the same company as Zoho CRM, at 9 USD per user per month. For a one-person business selling services it is closer to the size of the problem than either of the two compared here.'
            )
        )
    );
    current_content jsonb;
    anchor_position integer;
    offer           jsonb;
    placed          integer := 0;
BEGIN
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = entry_slug;

    IF current_content IS NULL THEN
        RAISE EXCEPTION '% entry is missing', entry_slug;
    END IF;

    SELECT COALESCE(jsonb_agg(element ORDER BY position), '[]'::jsonb)
    INTO current_content
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE NOT (element ->> 'type' = 'offer'
               AND element ->> 'product' = ANY (offer_products))
      AND NOT (element ->> 'type' = 'faq'
               AND element ->> 'heading' = faq_heading);

    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'heading'
      AND element ->> 'heading' = anchor_heading;

    IF anchor_position IS NULL THEN
        RAISE EXCEPTION '% has been re-edited: heading "%" is missing',
            entry_slug, anchor_heading;
    END IF;
    IF current_content -> anchor_position ->> 'type' IS DISTINCT FROM 'unordered_list' THEN
        RAISE EXCEPTION '% has been re-edited: expected an unordered_list under "%", found %',
            entry_slug, anchor_heading, current_content -> anchor_position ->> 'type';
    END IF;

    FOR offer IN SELECT value FROM jsonb_array_elements(offers) LOOP
        current_content := jsonb_insert(
            current_content, ARRAY[(anchor_position + placed)::text], offer, true
        );
        placed := placed + 1;
    END LOOP;

    current_content := current_content || jsonb_build_array(faq);

    UPDATE editorial.entries
    SET content = current_content,
        updated_at = now()
    WHERE slug = entry_slug;
END $$;

-- ============================================================ assertions ===
--
-- Presence, position and count, for every block placed above. Position is
-- asserted relative to the same headings the blocks were placed against, so
-- the check holds however the rest of the article moves.
DO $$
DECLARE
    current_content jsonb;
    anchor_position integer;
    last_block      jsonb;
BEGIN
    -- ------------------------------------------------ mailchimp-alternatives
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = 'mailchimp-alternatives';

    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'heading'
      AND element ->> 'heading' = 'Which one, in one line each';

    IF anchor_position IS NULL
       OR current_content -> anchor_position ->> 'type' IS DISTINCT FROM 'unordered_list'
       OR current_content -> (anchor_position + 1) ->> 'type' IS DISTINCT FROM 'offer'
       OR current_content -> (anchor_position + 1) ->> 'product' IS DISTINCT FROM 'zoho-campaigns-standard'
       OR current_content -> (anchor_position + 2) ->> 'product' IS DISTINCT FROM 'mailerlite-comfort'
       OR current_content -> (anchor_position + 3) ->> 'product' IS DISTINCT FROM 'kit-creator'
       -- The ActiveCampaign CTA still closes the section, directly after the rows.
       OR current_content -> (anchor_position + 4) ->> 'type' IS DISTINCT FROM 'cta'
       OR current_content -> (anchor_position + 4) ->> 'promotion' IS DISTINCT FROM 'activecampaign-mailchimp-switch'
       OR current_content -> (anchor_position + 5) ->> 'heading' IS DISTINCT FROM 'When you should stay on Mailchimp' THEN
        RAISE EXCEPTION 'mailchimp-alternatives offer rows are missing or misplaced around "Which one, in one line each"';
    END IF;

    IF (SELECT count(*) FROM jsonb_array_elements(current_content) AS element
        WHERE element ->> 'type' = 'offer'
          AND element ->> 'product' IN ('zoho-campaigns-standard', 'mailerlite-comfort', 'kit-creator')) <> 3
       OR (SELECT count(DISTINCT element ->> 'product') FROM jsonb_array_elements(current_content) AS element
           WHERE element ->> 'type' = 'offer') <> 3 THEN
        RAISE EXCEPTION 'mailchimp-alternatives carries a duplicated or missing offer block';
    END IF;

    last_block := current_content -> (jsonb_array_length(current_content) - 1);
    IF last_block ->> 'type' IS DISTINCT FROM 'faq'
       OR last_block ->> 'heading' IS DISTINCT FROM 'Questions people ask'
       OR jsonb_array_length(last_block -> 'questions') <> 4
       OR (SELECT count(*) FROM jsonb_array_elements(current_content) AS element
           WHERE element ->> 'type' = 'faq') <> 1 THEN
        RAISE EXCEPTION 'mailchimp-alternatives FAQ is missing, duplicated or not the last block';
    END IF;

    -- ------------------------------------------------- calendly-alternatives
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = 'calendly-alternatives';

    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'heading'
      AND element ->> 'heading' = 'Cal.com, if you need to own the data';

    -- 0-based index of the heading is position - 1; the row sits at position - 2.
    IF anchor_position IS NULL
       OR current_content -> (anchor_position - 2) ->> 'type' IS DISTINCT FROM 'offer'
       OR current_content -> (anchor_position - 2) ->> 'product' IS DISTINCT FROM 'zoho-bookings-basic'
       OR current_content -> (anchor_position - 3) ->> 'type' IS DISTINCT FROM 'paragraph' THEN
        RAISE EXCEPTION 'calendly-alternatives Zoho Bookings row is missing or misplaced before the Cal.com heading';
    END IF;

    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'callout'
      AND element ->> 'heading' = 'Why Calendly is still winning';

    IF anchor_position IS NULL
       OR current_content -> (anchor_position - 2) ->> 'type' IS DISTINCT FROM 'offer'
       OR current_content -> (anchor_position - 2) ->> 'product' IS DISTINCT FROM 'cal-com-teams'
       OR current_content -> (anchor_position - 3) ->> 'type' IS DISTINCT FROM 'paragraph' THEN
        RAISE EXCEPTION 'calendly-alternatives Cal.com row is missing or misplaced before the Calendly callout';
    END IF;

    IF (SELECT count(*) FROM jsonb_array_elements(current_content) AS element
        WHERE element ->> 'type' = 'offer'
          AND element ->> 'product' IN ('zoho-bookings-basic', 'cal-com-teams')) <> 2
       OR (SELECT count(DISTINCT element ->> 'product') FROM jsonb_array_elements(current_content) AS element
           WHERE element ->> 'type' = 'offer') <> 2 THEN
        RAISE EXCEPTION 'calendly-alternatives carries a duplicated or missing offer block';
    END IF;

    last_block := current_content -> (jsonb_array_length(current_content) - 1);
    IF last_block ->> 'type' IS DISTINCT FROM 'faq'
       OR last_block ->> 'heading' IS DISTINCT FROM 'Questions people ask'
       OR jsonb_array_length(last_block -> 'questions') <> 3
       OR (SELECT count(*) FROM jsonb_array_elements(current_content) AS element
           WHERE element ->> 'type' = 'faq') <> 1 THEN
        RAISE EXCEPTION 'calendly-alternatives FAQ is missing, duplicated or not the last block';
    END IF;

    -- --------------------------------------------------- zoho-crm-vs-hubspot
    SELECT content INTO current_content
    FROM editorial.entries WHERE slug = 'zoho-crm-vs-hubspot';

    SELECT position INTO anchor_position
    FROM jsonb_array_elements(current_content)
         WITH ORDINALITY AS blocks(element, position)
    WHERE element ->> 'type' = 'heading'
      AND element ->> 'heading' = 'Where each one is genuinely better';

    IF anchor_position IS NULL
       OR current_content -> anchor_position ->> 'type' IS DISTINCT FROM 'unordered_list'
       OR current_content -> (anchor_position + 1) ->> 'type' IS DISTINCT FROM 'offer'
       OR current_content -> (anchor_position + 1) ->> 'product' IS DISTINCT FROM 'zoho-crm-standard'
       OR current_content -> (anchor_position + 2) ->> 'product' IS DISTINCT FROM 'bigin-express'
       OR current_content -> (anchor_position + 3) ->> 'heading' IS DISTINCT FROM 'What we are not telling you' THEN
        RAISE EXCEPTION 'zoho-crm-vs-hubspot offer rows are missing or misplaced around "Where each one is genuinely better"';
    END IF;

    IF (SELECT count(*) FROM jsonb_array_elements(current_content) AS element
        WHERE element ->> 'type' = 'offer'
          AND element ->> 'product' IN ('zoho-crm-standard', 'bigin-express')) <> 2
       OR (SELECT count(DISTINCT element ->> 'product') FROM jsonb_array_elements(current_content) AS element
           WHERE element ->> 'type' = 'offer') <> 2 THEN
        RAISE EXCEPTION 'zoho-crm-vs-hubspot carries a duplicated or missing offer block';
    END IF;

    last_block := current_content -> (jsonb_array_length(current_content) - 1);
    IF last_block ->> 'type' IS DISTINCT FROM 'faq'
       OR last_block ->> 'heading' IS DISTINCT FROM 'Questions people ask'
       OR jsonb_array_length(last_block -> 'questions') <> 3
       OR (SELECT count(*) FROM jsonb_array_elements(current_content) AS element
           WHERE element ->> 'type' = 'faq') <> 1 THEN
        RAISE EXCEPTION 'zoho-crm-vs-hubspot FAQ is missing, duplicated or not the last block';
    END IF;
END $$;

COMMIT;
