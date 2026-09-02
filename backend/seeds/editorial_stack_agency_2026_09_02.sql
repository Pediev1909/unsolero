-- The first published stack: a whole set of tools for one kind of business and
-- one budget, written 2026-09-02.
--
-- Applied with the other editorial seeds, by psql against the application
-- database, after saas_editorial.sql (the author and the guide this builds
-- on), saas_commercial_pages.sql (best-crm-small-agency) and the catalog
-- seeds that publish the three products:
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 \
--        -f backend/seeds/editorial_stack_agency_2026_09_02.sql
--
-- Re-running it is safe. The entry is upserted by slug and its links are
-- replaced inside one transaction; a second run against an unchanged database
-- writes nothing, so updated_at (which the sitemap publishes) only moves when
-- the text does.
--
-- ---------------------------------------------------------------------------
-- Where every statement comes from.
--
-- The argument is the published guide /guides/software-stack-small-agency,
-- restated in the stack's own words: client records first, invoicing connected
-- to them, delivery tracking kept simple, chat and scheduling deferred.
--
-- Every price, billing basis and free tier is the catalog's own, as served by
-- /api/catalog/products/{slug} on 2026-09-02, and the per-user basis for Bigin
-- Express and Zoho CRM Standard is the one recorded in their price evidence.
-- Zoho Books Standard is counted once, as the catalog lists it and as the
-- published guide best-crm-small-agency already does. Nothing below is a new
-- claim about any product.
--
-- The three listed products each had a live affiliate offer on 2026-09-02, so
-- each gets an `offer` block. They were chosen by the guide's argument and the
-- catalog's prices; Zoho Projects Premium is the cheapest project tool in the
-- catalog, and Bigin is the cheaper of the two Zoho CRMs. Commission did not
-- enter the order and, per the offer block, cannot.

BEGIN;

WITH author AS (
    SELECT id FROM editorial.authors WHERE slug = 'andon-pediev'
), entry AS (
    SELECT
        'stack'::text AS content_type,
        $t$A 3-person agency's software stack under $150 a month$t$::text AS title,
        'agency-3-people-under-150'::text AS slug,
        'Client records, invoicing and delivery tracking for three people at 59 USD a month: what to run, what we deliberately left out and why, and where every price came from.'::text AS description,
        '/images/saas-agency-stack-v2.svg'::text AS hero_image_url,
        'A small team workspace with two people reviewing a project board together'::text AS hero_image_alt,
        $json$[
          {"type":"paragraph","text":"Three people, client work, nobody paid to look after software, and a ceiling of 150 USD a month for everything. This is the whole stack for that brief: what to run, what it costs, and the tools we deliberately left out. It follows our guide to building a software stack for a small agency, which argues the order — client records, then invoicing connected to them, then delivery tracking, and only then chat and scheduling — and it comes in well under the ceiling on purpose."},
          {"type":"heading","heading":"Who this is for"},
          {"type":"paragraph","text":"A client services agency of three: a designer, a developer and whoever sells, or any mix that wins work, does the work and invoices for it. Nobody on the team is paid to configure or maintain tools, so anything that needs an administrator is out before price is considered. You are not yet running a CRM or an invoicing tool that clients are used to. If you are, read the last section first, because what you already own changes the answer."},
          {"type":"heading","heading":"The stack"},
          {"type":"ordered_list","items":[
            "Bigin Express (Zoho) for client records — 9 USD per user per month on monthly billing, 7 USD billed annually. Three seats: 27 USD. Free for a single user.",
            "Zoho Books Standard for invoicing and bookkeeping — 20 USD per month on monthly billing, counted once. The catalog notes a free tier for businesses under 50,000 USD in annual revenue.",
            "Zoho Projects Premium for delivery tracking — 4 USD per user per month, billed annually; Zoho publishes no monthly rate. Three seats: 12 USD. Free for up to five users, so a three-person team can start at nothing."
          ]},
          {"type":"paragraph","text":"27 + 20 + 12 = 59 USD a month for three people, or 47 USD while Projects stays on its free tier. Bigin and Zoho Projects are priced per user and are counted here for three seats; Zoho Books is counted once, as the catalog lists it. Two of the three are on monthly billing and Zoho Projects is billed annually, so the 59 is a blended monthly figure and the annual part is paid up front. Prices as read from vendor pages on the dates shown on each product page. That leaves 91 USD of the 150 unspent, which is the result the guide predicts rather than a gap to fill: the tools people add next are the ones it says to defer."},
          {"type":"paragraph","text":"Why these three. The guide's first requirement is one shared, searchable place that knows who your clients are and what was agreed, connected to how you invoice so a client record is never typed twice; Bigin and Books come from the same vendor, which is the cheapest way to meet that without an integration project. Its second is that invoicing is not the place to save, because late and forgotten invoices are where a small agency loses money quietly, and Books adds the bank feed that tells you what has actually been paid. Its third is delivery tracking kept deliberately simple — who is doing what, what is late, what is waiting on the client — and Zoho Projects is the cheapest tool in the catalog that answers those three questions for three people."},
          {"type":"offer","product":"bigin-express","heading":"Where to get them","text":"Zoho's deliberately simpler pipeline CRM: 9 USD per user per month on monthly billing, 7 USD billed annually, free for a single user. Express caps three pipelines, 50,000 records and 30 automations, which is more room than a three-person agency needs."},
          {"type":"offer","product":"zoho-books-standard","text":"Invoicing and bookkeeping with bank feeds, 20 USD per month on monthly billing. Check the free tier first: the catalog records it as covering businesses under 50,000 USD in annual revenue."},
          {"type":"offer","product":"zoho-projects-premium","text":"Project tracking with time logging, blueprints and dependencies. 4 USD per user per month billed annually — Zoho publishes no monthly rate — and free for up to five users, so start on the free tier and pay when you need what Premium adds."},
          {"type":"heading","heading":"What we left out, and why"},
          {"type":"unordered_list","items":[
            "Zoho CRM Standard — 20 USD per user per month on monthly billing, 60 USD for three seats. The guide says the client record can start modest, and the catalog describes Bigin as Zoho's deliberately simpler CRM. Zoho CRM scores 90 for depth for power users; Bigin's one recorded weakness is exactly that depth. At three people you would be paying 33 USD a month for room you are not using. Move up when depth is what you are missing.",
            "Zoho Invoice — free, with no paid tier. It would do the invoicing, and the catalog is plain that it is not accounting: no bank reconciliation and no full ledger. The guide's warning about billing is that money leaks through invoices nobody chased, and knowing which ones were actually paid is a bank feed, not a template. If your books already live elsewhere, take Zoho Invoice and the stack drops to 39 USD.",
            "monday.com Basic (9 USD per seat per month billed annually, three-seat minimum) and ClickUp Unlimited (10 USD per user per month). Both cost more than double Zoho Projects for the same three questions, and the catalog notes that monday's Basic tier omits the timeline and calendar views most teams end up wanting. The guide's test applies: if a tool needs someone to maintain it and nobody is paid to, it is abandoned within two quarters, so extra surface is a cost, not a feature.",
            "Team chat — Slack Pro (8.75 USD per user per month), Microsoft Teams Essentials (4 USD per user per month, paid yearly) and Google Workspace Business Starter (7 USD per user per month). The guide's rule is that chat does not stop work or stop payment, so it waits until coordination cost is real, and three people can usually feel that moment coming. When it arrives, Workspace is the one that also brings business email on your own domain, Docs and Drive, so check what you already pay for before adding a second bill.",
            "Scheduling — Zoho Bookings Basic (8 USD per user per month, with a free tier), Cal.com Teams (12 USD per user per month billed yearly, free for an individual) and Calendly Standard (10 USD per seat per month). Same rule: add a booking tool once client booking is a bottleneck, not before. When it is, the free tiers of Zoho Bookings and Cal.com are where to start, and the paid seat is a decision for later."
          ]},
          {"type":"heading","heading":"When this stack is wrong"},
          {"type":"paragraph","text":"When you already own part of it. If clients are used to your invoicing tool or the team lives in a CRM, keep it: the guide's rule is that each addition should have something to connect to, and what you already run counts for more than what this page lists. When you sell products rather than time, the loop is different and so is the stack. When you are one person, the arithmetic changes in your favour: Bigin is free for a single user, Zoho Projects is free for up to five, and Zoho Invoice is free with no paid tier, so a solo consultant can run this whole shape at 0 USD until the second seat. And when you are past ten people, the guide this is built on stops applying — the features that get in the way at three start to earn their keep, and a different page should make that case."},
          {"type":"faq","heading":"Questions people ask","questions":[
            {"question":"Why is a CRM the first purchase for a three-person agency?","answer":"Because until there is one shared place that knows who your clients are and what was agreed, that knowledge lives in one person's inbox, which is a single point of failure with a holiday schedule. It does not need to be sophisticated. It needs to be shared, searchable and connected to how you invoice, so the same client record is never typed twice. That is why the pick is the modest Bigin rather than Zoho CRM."},
            {"question":"Do three people need a project management tool at all?","answer":"They need answers to three questions: who is doing what, what is late, and what is waiting on the client. Anything that answers those is enough, and anything much beyond them tends to become a second job: if a tool needs someone to maintain it and nobody is paid to, it is abandoned within two quarters. Zoho Projects is here because it is the cheapest tool in the catalog that answers the three, and it is free for up to five users."},
            {"question":"The budget has 91 USD left. Why not add Slack and a booking tool now?","answer":"Because neither stops work or stops payment. The order the guide argues is client records, invoicing connected to them, delivery tracking, and only then team chat once coordination cost is real and scheduling once client booking is a bottleneck. Adding them early is how a stack ends up costing more than it saves. The unspent budget is the result, not a gap to fill."}
          ]}
        ]$json$::jsonb AS content,
        $t$A 3-person agency's software stack under $150 a month | UNSOLERO$t$::text AS seo_title,
        'Three tools, 59 USD a month for three seats: Bigin Express, Zoho Books Standard and Zoho Projects. What we left out and why, with every price read from the vendor and dated.'::text AS seo_description
)
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description,
    published_at
)
SELECT author.id, entry.content_type, 'published', entry.title, entry.slug,
       entry.description, entry.hero_image_url, entry.hero_image_alt,
       entry.content, entry.seo_title, entry.seo_description,
       '2026-09-02 09:00:00+00'::timestamptz
FROM author CROSS JOIN entry
ON CONFLICT (slug) DO UPDATE SET
    author_id = EXCLUDED.author_id,
    content_type = EXCLUDED.content_type,
    status = 'published',
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    hero_image_url = EXCLUDED.hero_image_url,
    hero_image_alt = EXCLUDED.hero_image_alt,
    content = EXCLUDED.content,
    seo_title = EXCLUDED.seo_title,
    seo_description = EXCLUDED.seo_description,
    updated_at = now()
WHERE (editorial.entries.author_id, editorial.entries.content_type, editorial.entries.status,
       editorial.entries.title, editorial.entries.description, editorial.entries.hero_image_url,
       editorial.entries.hero_image_alt, editorial.entries.content, editorial.entries.seo_title,
       editorial.entries.seo_description)
    IS DISTINCT FROM
      (EXCLUDED.author_id, EXCLUDED.content_type, 'published',
       EXCLUDED.title, EXCLUDED.description, EXCLUDED.hero_image_url,
       EXCLUDED.hero_image_alt, EXCLUDED.content, EXCLUDED.seo_title,
       EXCLUDED.seo_description);

-- The three products in the stack, in the order the piece lists them. The
-- omitted products are named in the text but not linked: the link table is
-- what the "at a glance" strip and the card nameplate draw from, and a strip
-- that showed the rejected tools next to the chosen ones would be the stack
-- contradicting itself. Replaced wholesale, so a product removed from the
-- text is removed from the links on the same run.
DELETE FROM editorial.entry_products
WHERE entry_id = (SELECT id FROM editorial.entries WHERE slug = 'agency-3-people-under-150');

INSERT INTO editorial.entry_products (entry_id, product_id, position)
SELECT entries.id, products.id, mapping.position
FROM (VALUES
 ('bigin-express', 0),
 ('zoho-books-standard', 1),
 ('zoho-projects-premium', 2)
) AS mapping(product_slug, position)
JOIN editorial.entries ON entries.slug = 'agency-3-people-under-150'
JOIN catalog.products ON products.slug = mapping.product_slug;

DELETE FROM editorial.entry_categories
WHERE entry_id = (SELECT id FROM editorial.entries WHERE slug = 'agency-3-people-under-150');

INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT entries.id, categories.id, mapping.position
FROM (VALUES
 ('crm', 0),
 ('accounting-invoicing', 1),
 ('project-management', 2)
) AS mapping(category_slug, position)
JOIN editorial.entries ON entries.slug = 'agency-3-people-under-150'
JOIN catalog.categories
  ON categories.slug = mapping.category_slug AND categories.vertical_key = 'saas';

-- Outgoing: the guide this restates, the CRM guide for the same kind of
-- agency, and the method piece. Replaced wholesale like the products.
DELETE FROM editorial.related_entries
WHERE entry_id = (SELECT id FROM editorial.entries WHERE slug = 'agency-3-people-under-150');

INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT source.id, target.id, mapping.position
FROM (VALUES
 ('software-stack-small-agency', 0),
 ('best-crm-small-agency', 1),
 ('how-to-choose-business-software', 2)
) AS mapping(target_slug, position)
JOIN editorial.entries source ON source.slug = 'agency-3-people-under-150'
JOIN editorial.entries target ON target.slug = mapping.target_slug;

-- Incoming: the guide links to the stack that applies it, appended after
-- whatever the guide already relates to, once. This is the link a reader
-- arriving from a video follows; without it the stack has no inbound editorial
-- link at all.
INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT source.id, target.id,
       COALESCE((SELECT max(position) + 1 FROM editorial.related_entries WHERE entry_id = source.id), 0)
FROM editorial.entries source
JOIN editorial.entries target ON target.slug = 'agency-3-people-under-150'
WHERE source.slug = 'software-stack-small-agency'
  AND NOT EXISTS (
    SELECT 1 FROM editorial.related_entries
    WHERE entry_id = source.id AND related_entry_id = target.id
  );

-- Assertions. The inserts above join through slugs, so a missing author,
-- product or category would silently produce a stack with no links rather
-- than fail. Every one of those is checked here and raises, which rolls the
-- whole transaction back.
DO $$
DECLARE
    stack_id      uuid;
    product_links integer;
    category_links integer;
    outgoing_links integer;
    incoming_links integer;
    offer_blocks  integer;
    unlinked_offers integer;
    faq_blocks    integer;
    unknown_blocks integer;
BEGIN
    SELECT id INTO stack_id
    FROM editorial.entries
    WHERE slug = 'agency-3-people-under-150'
      AND status = 'published' AND content_type = 'stack';
    IF stack_id IS NULL THEN
        RAISE EXCEPTION 'agency-3-people-under-150 is not a published stack; is the andon-pediev author seeded?';
    END IF;

    SELECT count(*) INTO product_links FROM editorial.entry_products WHERE entry_id = stack_id;
    IF product_links <> 3 THEN
        RAISE EXCEPTION 'expected 3 product links on the stack, found %; a product slug is missing from the catalog', product_links;
    END IF;

    SELECT count(*) INTO category_links FROM editorial.entry_categories WHERE entry_id = stack_id;
    IF category_links <> 3 THEN
        RAISE EXCEPTION 'expected 3 category links on the stack, found %', category_links;
    END IF;

    SELECT count(*) INTO outgoing_links FROM editorial.related_entries WHERE entry_id = stack_id;
    IF outgoing_links <> 3 THEN
        RAISE EXCEPTION 'expected 3 related entries on the stack, found %; an editorial seed has not been applied', outgoing_links;
    END IF;

    SELECT count(*) INTO incoming_links
    FROM editorial.related_entries
    JOIN editorial.entries source ON source.id = related_entries.entry_id
    WHERE related_entry_id = stack_id AND source.slug = 'software-stack-small-agency';
    IF incoming_links <> 1 THEN
        RAISE EXCEPTION 'expected the agency guide to relate to the stack exactly once, found %', incoming_links;
    END IF;

    -- One offer block per linked product, and none for a product the stack
    -- does not list: the strip and the blocks must agree on what the stack is.
    SELECT count(*) INTO offer_blocks
    FROM editorial.entries, jsonb_array_elements(content) AS block
    WHERE id = stack_id AND block ->> 'type' = 'offer';
    SELECT count(*) INTO unlinked_offers
    FROM editorial.entries, jsonb_array_elements(content) AS block
    WHERE id = stack_id AND block ->> 'type' = 'offer'
      AND NOT EXISTS (
        SELECT 1 FROM editorial.entry_products
        JOIN catalog.products ON products.id = entry_products.product_id
        WHERE entry_products.entry_id = stack_id AND products.slug = block ->> 'product'
      );
    IF offer_blocks <> 3 OR unlinked_offers <> 0 THEN
        RAISE EXCEPTION 'expected 3 offer blocks, all for linked products; found % blocks, % unlinked', offer_blocks, unlinked_offers;
    END IF;

    SELECT count(*) INTO faq_blocks
    FROM editorial.entries, jsonb_array_elements(content) AS block
    WHERE id = stack_id AND block ->> 'type' = 'faq';
    IF faq_blocks <> 1 THEN
        RAISE EXCEPTION 'expected exactly one FAQ block, found %', faq_blocks;
    END IF;

    SELECT count(*) INTO unknown_blocks
    FROM editorial.entries, jsonb_array_elements(content) AS block
    WHERE id = stack_id
      AND block ->> 'type' NOT IN ('paragraph', 'heading', 'unordered_list', 'ordered_list',
                                   'quote', 'callout', 'cta', 'pros_cons', 'faq', 'offer');
    IF unknown_blocks <> 0 THEN
        RAISE EXCEPTION '% block(s) of a type the content domain does not render', unknown_blocks;
    END IF;
END $$;

COMMIT;
