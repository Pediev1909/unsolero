-- Five "alternatives to X" pages.
--
-- A comparison catches somebody who has narrowed the field to two. This shape
-- catches somebody earlier and angrier: they already pay for something, it has
-- annoyed them, and they are looking for the exit. It is the strongest buying
-- intent there is, and the catalog can answer all five of these honestly.
--
-- Two rules were applied throughout, and both cost the page something.
--
-- First, each one says plainly when staying put is the right answer. A page
-- called "alternatives to X" that never concludes "keep X" is an advert with a
-- question mark on the end.
--
-- Second, the incumbent is described accurately rather than as a straw man.
-- Google Analytics and Mailchimp are good products that many readers should
-- keep, and a page that pretends otherwise gets found out by the first reader
-- who knows the category.
--
-- Every price below comes from catalog.products.

WITH author AS (
    SELECT id FROM editorial.authors WHERE slug = 'andon-pediev'
)
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description,
    published_at
)
SELECT author.id, 'buying_guide', 'published', entry.title, entry.slug,
       entry.description, entry.hero_image_url, entry.hero_image_alt,
       entry.content::jsonb, entry.seo_title, entry.seo_description, now()
FROM author CROSS JOIN (VALUES

-- ============================================ 1. GOOGLE ANALYTICS =====
(
    'Google Analytics alternatives that need no cookie banner',
    'google-analytics-alternatives',
    'Three analytics tools that do not use cookies, do not need consent, and fit on one screen. Priced at the same traffic volume so the comparison means something.',
    '/images/saas-alternatives.svg',
    'A simple analytics dashboard shown without a cookie consent banner over it',
    $json$[
      {"type":"paragraph","text":"People leave Google Analytics for one of two reasons: the consent banner, or GA4 itself. If it is the banner, any of the three below removes it, because none of them sets a cookie or collects anything that needs permission under GDPR. If it is GA4, you are looking for something simpler, and all three are that too."},
      {"type":"heading","heading":"The three, priced at 100,000 pageviews a month"},
      {"type":"unordered_list","items":["Umami Cloud Pro: 20 USD per month. The Hobby tier is free up to 100,000 events a month, and the whole thing is open source and self-hostable.","Fathom Analytics: 15 USD per month, covering up to 50 sites on one plan.","Simple Analytics: 20 USD per month."]},
      {"type":"paragraph","text":"Quoting all three at one volume is the only honest way to do this. Every tool here charges by traffic, so a bare monthly price compares nothing unless you say how much traffic is behind it."},
      {"type":"heading","heading":"Which one, in one line each"},
      {"type":"unordered_list","items":["Under 100,000 a month: Umami, free, and revisit when you outgrow it.","Several small sites: Fathom, because 50 sites on one 15 USD plan is unmatched here.","One site and you never want to think about it again: Simple Analytics.","Your own server, non-negotiable: Umami self-hosted, which is free and is work."]},
      {"type":"callout","heading":"The thing that actually saves you money","text":"Removing the consent banner does more than tidy the page. A banner suppresses a meaningful share of measured traffic because people dismiss it without consenting. Cookieless analytics counts everyone, so the numbers go up on the day you switch — not because you got more visitors, but because you were always undercounting."},
      {"type":"heading","heading":"When you should stay on Google Analytics"},
      {"type":"paragraph","text":"If you run paid acquisition, keep it. Conversion tracking, audience export to Google Ads, attribution across campaigns — none of the three above replaces any of that, and no amount of dislike for GA4 changes it. The tools here are for people measuring whether their site works, not people buying traffic."},
      {"type":"paragraph","text":"If you need multi-step funnels, cohort retention or user-level session replay, these are also the wrong shelf. They are deliberately shallow. That is the product, not a gap in it."},
      {"type":"heading","heading":"What migrating actually costs"},
      {"type":"paragraph","text":"None of them imports your Google Analytics history. You start from zero on the day you install the script, which means a year-on-year comparison is impossible for a year. Most people run both side by side for a month, confirm the numbers make sense, and then drop the banner. That month is the real cost of moving."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not audited any of these privacy claims against actual behaviour. Each publishes how it works and we have read those claims rather than verified them, which is a different thing and we would rather say so."}
    ]$json$,
    'Google Analytics alternatives without a cookie banner',
    'Three cookieless analytics tools priced at the same traffic volume, plus the cases where you should keep Google Analytics instead.'
),

-- =================================================== 2. MAILCHIMP =====
(
    'Mailchimp alternatives for a list you actually own',
    'mailchimp-alternatives',
    'Five email tools at every price point from 5.25 to 39 USD, all quoted at 1,000 subscribers, plus the one question that tells you whether to move at all.',
    '/images/saas-alternatives.svg',
    'An email list being moved from one platform to another',
    $json$[
      {"type":"paragraph","text":"Almost nobody leaves Mailchimp because of the features. They leave because the bill grew faster than the list did, and because contacts they no longer email still count toward it. If that is your reason, the move is worth it and the options below are all cheaper. If your reason is that you want better automation, read the last section first."},
      {"type":"heading","heading":"The options, all at 1,000 subscribers"},
      {"type":"unordered_list","items":["Zoho Campaigns Standard: 5.25 USD per month. The cheapest here, and it connects to Zoho CRM without any work.","Brevo Starter: 9 USD per month. Prices by emails sent rather than by contacts held, which is the whole point if your list is large and quiet.","ActiveCampaign Starter: 15 USD per month, billed annually. The deepest automation of the five and the only one with no free tier.","MailerLite Comfort: 19 USD per month, monthly billing. The best editor here and a free tier up to 250 subscribers.","Kit Creator: 39 USD per month. Built for people who publish, with a free tier reaching 10,000 subscribers with features withheld."]},
      {"type":"callout","heading":"The question that decides it","text":"Do you pay for contacts you never email? Mailchimp counts them; Brevo does not, because it charges for sends. For a list that is big and mostly dormant, that single difference is usually the entire saving, and no feature comparison will match it."},
      {"type":"heading","heading":"Which one, in one line each"},
      {"type":"unordered_list","items":["Large list, low send volume: Brevo, priced on sends.","Already running Zoho anything: Zoho Campaigns, cheapest here and already connected.","Automation is the reason you are moving: ActiveCampaign, and expect to spend a weekend on it.","You want a clean editor and no fuss: MailerLite.","You publish — newsletter, courses, a body of writing: Kit, which costs the most and is built for exactly that."]},
      {"type":"heading","heading":"When you should stay on Mailchimp"},
      {"type":"paragraph","text":"If your list is under a few thousand and you send often, the price gap is small enough that moving is not worth a weekend. Mailchimp is also still the easiest to hand to somebody non-technical, and its templates are better than most of what is above."},
      {"type":"paragraph","text":"And if you are moving purely because automation frustrated you, be careful: ActiveCampaign is genuinely deeper, but the other four are not. Swapping for a cheaper tool with less automation solves a bill, not the problem you had."},
      {"type":"heading","heading":"What moving actually costs"},
      {"type":"paragraph","text":"The subscriber export is easy and every tool here imports a CSV. The parts that take time are rebuilding automations, re-authenticating your sending domain, and the deliverability dip while the new provider warms up your reputation. Budget a fortnight of watching open rates before you decide the move failed."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not tested deliverability, and anybody who claims a ranking of it from public information is guessing. What we can compare is what each vendor publishes and what each meter counts, which is where the money actually goes."}
    ]$json$,
    'Mailchimp alternatives compared at 1,000 subscribers',
    'Five email tools from 5.25 to 39 USD, all quoted at the same list size, plus the cases where staying on Mailchimp is the right call.'
),

-- ===================================================== 3. ZAPIER =====
(
    'Zapier alternatives when the task bill stops making sense',
    'zapier-alternatives',
    'Two serious alternatives, one of them free if you host it yourself, plus an honest account of what you give up by leaving the largest connector library in the category.',
    '/images/saas-alternatives.svg',
    'An automation workflow being rebuilt on a different platform',
    $json$[
      {"type":"paragraph","text":"Zapier gets replaced for one reason: the meter. It counts every step of every run, so a workflow that fires often costs more than the plan you priced it on. Before you move, though, check the connector list, because that is the thing Zapier is actually selling and the thing you will miss."},
      {"type":"heading","heading":"The two worth moving to"},
      {"type":"unordered_list","items":["Make Core: 12 USD per month, billed annually, for 10,000 credits. Monthly billing is 16 USD.","n8n Starter: 20 USD per month, billed annually, for 2,500 executions on their cloud — or nothing at all if you run it on your own server.","For reference, Zapier Professional is 19.99 USD per month billed annually, priced by task volume."]},
      {"type":"heading","heading":"Make, if you want cheaper and clearer"},
      {"type":"paragraph","text":"Make draws a workflow as a diagram instead of listing it as steps. That sounds cosmetic until something branches, at which point a five-branch flow in a linear builder is a wall of indented text and the same flow on a canvas can be understood at a glance. It is cheaper than Zapier at comparable volume and harder to learn in the first hour."},
      {"type":"heading","heading":"n8n, if you want to own it"},
      {"type":"paragraph","text":"n8n is open source. Run it on your own server and it costs nothing but your time, and the workflows and every record passing through them stay on hardware you control. For anything touching customer data that is a real argument, and it is the only option here that offers it."},
      {"type":"paragraph","text":"Self-hosting is not free in the way people mean when they say free. Somebody has to keep it running, patched and backed up. If that somebody is you and you enjoy it, this is the cheapest automation on the market. If it is not, pay for the cloud tier or pick Make."},
      {"type":"callout","heading":"The check that takes two minutes","text":"Open Make's app directory and n8n's, and search for the two least famous tools you use. If neither has both, stay on Zapier and stop reading. The connector library is the product, and no price difference compensates for a missing integration."},
      {"type":"heading","heading":"When you should stay on Zapier"},
      {"type":"paragraph","text":"If your automations are few, simple and infrequent, the task meter is not hurting you and moving buys nothing. If your integrations are obscure, Zapier will have them and the others may not. And if somebody non-technical maintains your workflows, Zapier is the easiest of the three by a clear margin — that is worth more than the price gap."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not benchmarked reliability. All three drop a run occasionally and all three tell you about it. What we can compare is published pricing and what each meter counts, which is the part vendors make hardest to line up."}
    ]$json$,
    'Zapier alternatives: Make, n8n and when to stay',
    'Two serious alternatives to Zapier, one free if self-hosted, plus the two-minute check that tells you whether moving is a mistake.'
),

-- =================================================== 4. CALENDLY =====
(
    'Calendly alternatives worth the switch',
    'calendly-alternatives',
    'Two alternatives, one cheaper and one you can host yourself, and a straight answer about why most people should keep paying Calendly anyway.',
    '/images/saas-alternatives.svg',
    'A booking link being replaced with a different scheduling tool',
    $json$[
      {"type":"paragraph","text":"Calendly Standard is 10 USD per seat per month. The two alternatives below are 8 and 12 dollars, so nobody is switching to save money. They are worth it for exactly two reasons: you already run Zoho, or you need the booking data on your own server. If neither applies, this page ends with a recommendation to stay."},
      {"type":"heading","heading":"The two"},
      {"type":"unordered_list","items":["Zoho Bookings Basic: 8 USD per user per month, with a free tier. Arrives already connected to Zoho CRM, invoicing and email.","Cal.com Teams: 12 USD per user per month billed yearly, and free forever for one person. Open source and self-hostable."]},
      {"type":"heading","heading":"Zoho Bookings, if you are already inside Zoho"},
      {"type":"paragraph","text":"The saving is two dollars, which is not the point. The point is that a booking lands against a CRM record that already exists, and the invoice afterwards comes from the same place. If you run Zoho, that connection removes a step every single time. If you do not, you are learning an unfamiliar tool to save the price of a coffee, which is a bad trade."},
      {"type":"heading","heading":"Cal.com, if you need to own the data"},
      {"type":"paragraph","text":"Open source, and it can run on your own server, so the record of who met whom and when never leaves your infrastructure. For most businesses that is worth nothing. For a few — anyone handling medical, legal or otherwise sensitive appointments — it is the whole decision, and Cal.com is the only serious option here that offers it."},
      {"type":"paragraph","text":"It is also free forever for a single person, which makes it the strongest free tier in this category by some distance. A solo consultant who does not need team features can stop paying for scheduling entirely."},
      {"type":"callout","heading":"Why Calendly is still winning","text":"A booking link nobody has to be taught to use is worth more than a feature list suggests. Calendly is the name people recognise in an email, and recognition removes hesitation at the exact moment you are asking somebody to commit their time. That does not appear on any comparison table and it is the strongest thing Calendly has."},
      {"type":"heading","heading":"When you should stay on Calendly"},
      {"type":"paragraph","text":"If your bookings come from strangers — prospects, clients, people who found you online — stay. The two dollars is not worth even the deliberation, let alone the switch. Move only if you are inside Zoho already, or if the data has to live on your own machine."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We are recommending the most expensive of the three for the common case, on a judgement about what recognition is worth to a business asking strangers for meetings. That is a judgement, not a measurement, and we have labelled it as one rather than dressing it up."}
    ]$json$,
    'Calendly alternatives: Cal.com, Zoho Bookings',
    'Two alternatives to Calendly and an honest answer about why most people should keep paying for it. Two reasons justify the switch.'
),

-- ====================================================== 5. SLACK =====
(
    'Slack alternatives when the per-seat bill adds up',
    'slack-alternatives',
    'Two cheaper options at 4 and 7 USD per user, what each one actually includes beyond chat, and why the free Slack tier hiding your history changes the maths.',
    '/images/saas-alternatives.svg',
    'A team chat conversation being moved to a different platform',
    $json$[
      {"type":"paragraph","text":"Slack Pro is 8.75 USD per user per month. On a team of ten that is a bill worth examining, and both alternatives below cost less. But neither is only a chat tool, and that changes what you are actually comparing."},
      {"type":"heading","heading":"The two cheaper options"},
      {"type":"unordered_list","items":["Microsoft Teams Essentials: 4 USD per user per month, paid yearly. The standalone plan — chat, calls and video, with no Word, no Excel and no business email.","Google Workspace Business Starter: 7 USD per user per month. Chat is the smallest part: it includes business email on your own domain, 30 GB per person, Docs, Meet and Drive."]},
      {"type":"heading","heading":"Google Workspace is not really a Slack alternative"},
      {"type":"paragraph","text":"At 7 USD you are buying an email and document suite that happens to include chat. If you already pay separately for business email, this is not a 7 dollar chat tool — it is a consolidation that removes another bill. Compare it against Slack plus whatever you currently pay for email, not against Slack alone."},
      {"type":"heading","heading":"Microsoft Teams Essentials is the genuine price cut"},
      {"type":"paragraph","text":"At 4 USD it is less than half of Slack Pro and it does the job: chat, calls, video, file sharing. The catch is the name. Essentials is the standalone plan; the Teams most people mean comes inside Microsoft 365 Business Basic at 7 USD, which adds the Office web apps and business email. Check which one you are actually being sold."},
      {"type":"callout","heading":"The thing that decides it for most teams","text":"Slack's free tier now hides messages after 90 days. If you were treating free Slack as your archive, that archive is already gone from view, and the choice is not between free and paid — it is between paying Slack and paying somebody else. That reframing is what usually triggers the switch."},
      {"type":"heading","heading":"When you should stay on Slack"},
      {"type":"paragraph","text":"Slack is the best chat product of the three and it is not close. Search is better, the integrations are deeper, and people notice the difference within a week of leaving. If your team lives in it all day, saving four dollars a head per month against a tool everybody quietly resents is a poor trade."},
      {"type":"paragraph","text":"It is also the one that connects to everything else. If your automations, your help desk and your project tool all post into Slack today, count the cost of rebuilding those connections before you count the saving."},
      {"type":"heading","heading":"What migrating actually costs"},
      {"type":"paragraph","text":"History is the hard part. Exporting Slack and importing it somewhere legible is awkward at best, and most teams end up leaving the old workspace read-only and starting fresh. If the archive matters to you, price that in as real work rather than assuming it comes along."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"This compares published pricing and published inclusions. How each one feels after six months, and whether your team will accept the change, is not something we can measure from a pricing page — and it is usually what actually decides whether a migration sticks."}
    ]$json$,
    'Slack alternatives: Teams and Google Workspace',
    'Two cheaper team chat options at 4 and 7 USD per user, what each includes beyond chat, and when staying on Slack is the better call.'
)

) AS entry(title, slug, description, hero_image_url, hero_image_alt, content, seo_title, seo_description)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO editorial.entry_products (entry_id, product_id, position)
SELECT entries.id, products.id, mapping.position
FROM (VALUES
 ('google-analytics-alternatives','fathom-analytics',0),
 ('google-analytics-alternatives','umami-cloud-pro',1),
 ('google-analytics-alternatives','simple-analytics',2),
 ('mailchimp-alternatives','zoho-campaigns-standard',0),
 ('mailchimp-alternatives','brevo-starter',1),
 ('mailchimp-alternatives','activecampaign-starter',2),
 ('mailchimp-alternatives','mailerlite-comfort',3),
 ('mailchimp-alternatives','kit-creator',4),
 ('zapier-alternatives','make-core',0),
 ('zapier-alternatives','n8n-starter',1),
 ('zapier-alternatives','zapier-professional',2),
 ('calendly-alternatives','zoho-bookings-basic',0),
 ('calendly-alternatives','cal-com-teams',1),
 ('calendly-alternatives','calendly-standard',2),
 ('slack-alternatives','microsoft-teams-essentials',0),
 ('slack-alternatives','google-workspace-business-starter',1),
 ('slack-alternatives','slack-pro',2)
) AS mapping(entry_slug, product_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.products products ON products.slug = mapping.product_slug
ON CONFLICT DO NOTHING;

INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT entries.id, categories.id, mapping.position
FROM (VALUES
 ('google-analytics-alternatives','analytics',0),
 ('mailchimp-alternatives','email-marketing',0),
 ('zapier-alternatives','automation',0),
 ('calendly-alternatives','scheduling',0),
 ('slack-alternatives','team-communication',0)
) AS mapping(entry_slug, category_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.categories categories
  ON categories.slug = mapping.category_slug AND categories.vertical_key = 'saas'
ON CONFLICT DO NOTHING;

-- Someone reading "alternatives to X" is one step from a head-to-head between
-- the two finalists, so each alternatives page points at the comparison that
-- follows it rather than at whatever was published most recently.
INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT source.id, target.id, mapping.position
FROM (VALUES
 ('google-analytics-alternatives','fathom-vs-simple-analytics-vs-umami',0),
 ('google-analytics-alternatives','how-unsolero-ranks-software',1),
 ('mailchimp-alternatives','how-to-choose-business-software',0),
 ('mailchimp-alternatives','how-unsolero-ranks-software',1),
 ('zapier-alternatives','zapier-vs-make',0),
 ('zapier-alternatives','how-to-choose-business-software',1),
 ('calendly-alternatives','calendly-vs-cal-com-vs-zoho-bookings',0),
 ('calendly-alternatives','software-stack-small-agency',1),
 ('slack-alternatives','software-stack-small-agency',0),
 ('slack-alternatives','how-to-choose-business-software',1),
 ('fathom-vs-simple-analytics-vs-umami','google-analytics-alternatives',1),
 ('zapier-vs-make','zapier-alternatives',2),
 ('calendly-vs-cal-com-vs-zoho-bookings','calendly-alternatives',2)
) AS mapping(source_slug, target_slug, position)
JOIN editorial.entries source ON source.slug = mapping.source_slug
JOIN editorial.entries target ON target.slug = mapping.target_slug
ON CONFLICT DO NOTHING;
