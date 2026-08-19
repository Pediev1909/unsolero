-- The three page types that carry commercial search intent.
--
-- Buyers who are close to a decision search with a small set of modifiers:
-- "X vs Y", "alternatives to X", and "best X for [use case]". The site had
-- product pages, category pages and method essays, and none of these — it was
-- writing for people who did not know what they wanted yet, and silent for
-- people about to choose.
--
-- Every price below comes from the catalog, which records the vendor pricing
-- page it was read from and the date. Nothing here asserts a fact the catalog
-- does not hold.
--
-- Each piece leads with its answer. That is how a reader in a hurry uses it,
-- and it is also what makes a page quotable by an assistant rather than merely
-- crawlable.

WITH author AS (
    SELECT id FROM editorial.authors WHERE slug = 'unsolero-editorial'
)
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description,
    published_at
)
SELECT author.id, entry.content_type, 'published', entry.title, entry.slug,
       entry.description, entry.hero_image_url, entry.hero_image_alt,
       entry.content::jsonb, entry.seo_title, entry.seo_description, now()
FROM author CROSS JOIN (VALUES
(
    'comparison',
    'ClickUp vs Teamwork for client work',
    'clickup-vs-teamwork',
    'Both track projects. One is built around client work and bills for it; the other is broader and cheaper. Which is right depends on whether you invoice by the hour.',
    '/images/saas-comparison.svg',
    'Two product cards side by side, differing on one row',
    $json$[
      {"type":"paragraph","text":"If you bill clients for time, Teamwork is the better fit despite costing more. If you do not, ClickUp does the same job for less and gives you more room to grow into. That is the whole decision; the rest of this page is why."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":[
        "ClickUp Unlimited: 10 USD per user per month, billed monthly. A free tier exists with a storage cap.",
        "Teamwork Basics: 12.99 USD per user per month, billed monthly. Free for up to five users."
      ]},
      {"type":"paragraph","text":"Both are cheaper on annual billing. We compare monthly rates because that is what you pay if you are not yet certain, and being uncertain is the normal state when you are choosing."},
      {"type":"heading","heading":"What actually separates them"},
      {"type":"paragraph","text":"Teamwork is built for agencies: billable time, client access and the accounting connection are the product rather than additions to it. If your week ends with an invoice, that shape saves you a step every single time."},
      {"type":"paragraph","text":"ClickUp is broader. It will track client projects perfectly well, and it will also track everything else you do. That breadth is a real advantage if your work is not all client work, and a real cost if it is: more surface to configure, more decisions nobody has time to make."},
      {"type":"callout","heading":"The test that decides it","text":"Does your project tool need to know what is billable? If yes, Teamwork. If the answer is \"we handle that separately\", ClickUp and keep the difference."},
      {"type":"heading","heading":"Where the free tiers land"},
      {"type":"paragraph","text":"Teamwork's free tier covers five users, which is a real team. ClickUp's is unlimited on users but capped on storage, which usually bites later and less predictably. For a small agency, Teamwork's free tier is the more honest starting point."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not used either daily for a year, and this page does not pretend otherwise. It compares what each vendor publishes and how those facts land against a specific brief. Anyone claiming a definitive verdict on which is nicer to live in is describing their preference, not your business."}
    ]$json$,
    'ClickUp vs Teamwork for client work',
    'Teamwork if you bill clients for time, ClickUp if you do not. Prices, free tiers and the one question that decides it.'
),
(
    'buying_guide',
    'HubSpot alternatives for a small team',
    'hubspot-alternatives',
    'HubSpot is the default CRM answer, and for a small team the default is often oversized. What the alternatives trade away, and when that trade is right.',
    '/images/saas-alternatives.svg',
    'One large product beside several smaller candidates',
    $json$[
      {"type":"paragraph","text":"Most small teams looking past HubSpot want one of two things: a lower bill, or less product to configure. Those point in different directions, so decide which one you are actually chasing before you look at anything."},
      {"type":"heading","heading":"Why people leave, and why they stay"},
      {"type":"paragraph","text":"HubSpot Starter is 20 USD per seat per month and bundles entry-level marketing, sales and service tools. That bundle is the reason to stay: it removes several purchases at once. It is also the reason to leave, because you pay per seat for surface you may never open."},
      {"type":"paragraph","text":"There is a free tier covering two users, which is often enough to postpone the decision entirely. Postponing is underrated."},
      {"type":"heading","heading":"If you want less product to configure"},
      {"type":"paragraph","text":"Salesflare Growth is 39 USD per user per month — more per seat, not less — and earns it by building contact records automatically from email and calendar activity. For a team with nobody assigned to keeping a CRM tidy, a CRM that maintains itself is worth more than a cheaper one that does not."},
      {"type":"callout","heading":"The cost nobody quotes","text":"A tool that needs someone to maintain it, with nobody paid to maintain it, is abandoned within two quarters. The subscription is not the expensive part."},
      {"type":"heading","heading":"If you want a lower bill"},
      {"type":"paragraph","text":"Then the honest answer is that the alternative is not another CRM. It is HubSpot's free tier until two users stops being enough, and putting the saved money toward the invoicing and delivery tools that actually stop work when they are missing."},
      {"type":"heading","heading":"What to check before switching"},
      {"type":"ordered_list","items":[
        "Can you export your contacts and their history yourself, without asking support?",
        "Does the replacement connect to whatever you invoice with?",
        "Who is going to run the migration, and in whose week does that fit?",
        "What breaks on the day you stop paying?"
      ]},
      {"type":"paragraph","text":"Switching CRM is the most disruptive change in a small stack because everything else references it. It is worth doing for a reason, and rarely worth doing for a discount."}
    ]$json$,
    'HubSpot alternatives for a small team',
    'What you trade for a cheaper or simpler CRM than HubSpot, and the four questions to answer before switching.'
),
(
    'buying_guide',
    'Best CRM for a small client services agency',
    'best-crm-small-agency',
    'A recommendation for a four-person agency with no dedicated admin, and the reasoning behind it, including what was rejected.',
    '/images/saas-use-case.svg',
    'A brief on one side producing a shortlist on the other',
    $json$[
      {"type":"paragraph","text":"For a four-person agency with nobody assigned to maintaining tools, start on HubSpot's free tier and move to Starter at 20 USD per seat when two users stops being enough. Choose Salesflare Growth instead if your team will not keep records up to date by hand, which is most teams."},
      {"type":"heading","heading":"The brief this answers"},
      {"type":"unordered_list","items":[
        "Four people, billing clients for project work",
        "Nobody paid to configure or maintain software",
        "Around 120 USD a month for the whole stack, not just the CRM",
        "Already running team chat"
      ]},
      {"type":"paragraph","text":"If that is not your brief, the answer changes, which is the point. A CRM recommendation with no stated constraints is a preference wearing a suit."},
      {"type":"heading","heading":"Why the CRM is not the first purchase"},
      {"type":"paragraph","text":"It usually is bought first and usually should not be. At four people the jobs that stop work or stop payment are delivery tracking and invoicing. A shared record of clients matters, but it can start free while the other two cannot start at all."},
      {"type":"paragraph","text":"Cover it within the whole stack: ClickUp Unlimited at 10 USD per user for delivery, Zoho Books Standard at 12 USD for invoicing, and the CRM on a free tier until it hurts. That leaves most of the budget unspent, which is a result rather than a failure."},
      {"type":"heading","heading":"What we rejected, and why"},
      {"type":"paragraph","text":"A help desk and an analytics tool. At four people the shared inbox already covers support, and there is no traffic to analyse yet. Both are good products bought two years early, which is the most common way a small stack becomes expensive."},
      {"type":"callout","heading":"Where the real cost hides","text":"Not the subscriptions. It is the afternoon a week somebody spends re-typing between two tools that do not connect. Pick for the connection before the feature list."},
      {"type":"heading","heading":"When to revisit"},
      {"type":"paragraph","text":"When the free CRM tier runs out of seats, when a client asks for something your invoicing cannot express, or when somebody is spending real hours moving data by hand. Not on a schedule, and not because a comparison article told you the category moved."}
    ]$json$,
    'Best CRM for a small client services agency',
    'A recommendation for a four-person agency with no dedicated admin, the whole-stack budget behind it, and what we deliberately rejected.'
)
) AS entry(content_type, title, slug, description, hero_image_url, hero_image_alt, content, seo_title, seo_description)
ON CONFLICT (slug) DO NOTHING;

-- Linking each page to the products it discusses puts it on those product
-- pages and vice versa. Internal links between related pages are how a cluster
-- reads as one authority rather than three loose essays.
INSERT INTO editorial.entry_products (entry_id, product_id, position)
SELECT entries.id, products.id, mapping.position
FROM (VALUES
 ('clickup-vs-teamwork','clickup-unlimited',0),
 ('clickup-vs-teamwork','teamwork-basics',1),
 ('hubspot-alternatives','hubspot-starter-customer-platform',0),
 ('hubspot-alternatives','salesflare-growth',1),
 ('best-crm-small-agency','hubspot-starter-customer-platform',0),
 ('best-crm-small-agency','salesflare-growth',1),
 ('best-crm-small-agency','clickup-unlimited',2),
 ('best-crm-small-agency','zoho-books-standard',3)
) AS mapping(entry_slug, product_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.products products ON products.slug = mapping.product_slug
ON CONFLICT DO NOTHING;

INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT entries.id, categories.id, mapping.position
FROM (VALUES
 ('clickup-vs-teamwork','project-management',0),
 ('hubspot-alternatives','crm',0),
 ('best-crm-small-agency','crm',0),
 ('best-crm-small-agency','project-management',1),
 ('best-crm-small-agency','accounting-invoicing',2)
) AS mapping(entry_slug, category_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.categories categories
  ON categories.slug = mapping.category_slug AND categories.vertical_key = 'saas'
ON CONFLICT DO NOTHING;

INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT source.id, target.id, mapping.position
FROM (VALUES
 ('best-crm-small-agency','hubspot-alternatives',0),
 ('best-crm-small-agency','clickup-vs-teamwork',1),
 ('best-crm-small-agency','software-stack-small-agency',2),
 ('hubspot-alternatives','best-crm-small-agency',0),
 ('hubspot-alternatives','how-unsolero-ranks-software',1),
 ('clickup-vs-teamwork','best-crm-small-agency',0),
 ('clickup-vs-teamwork','software-stack-small-agency',1),
 ('software-stack-small-agency','best-crm-small-agency',2),
 ('how-to-choose-business-software','clickup-vs-teamwork',2)
) AS mapping(source_slug, target_slug, position)
JOIN editorial.entries source ON source.slug = mapping.source_slug
JOIN editorial.entries target ON target.slug = mapping.target_slug
ON CONFLICT DO NOTHING;
