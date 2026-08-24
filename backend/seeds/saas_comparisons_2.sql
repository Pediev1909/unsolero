-- Five comparisons, one for each category that had no editorial attached.
--
-- The alternative was to link an existing piece that half-fitted, which is
-- padding, and a "related guidance" box holding something unrelated is worse
-- than one that is honestly absent. Each of these is a query people actually
-- type, and every price in them is already in the catalog with a source and a
-- date behind it.
--
-- The SEO one carries the only non-Zoho affiliate link on the site. It still
-- does not conclude in SE Ranking's favour for most readers, because for most
-- readers Ahrefs Starter at a third of the price is the right answer.

WITH author AS (SELECT id FROM editorial.authors WHERE slug = 'andon-pediev')
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description, published_at)
SELECT author.id, 'comparison', 'published', e.title, e.slug, e.description,
       e.hero_image_url, e.hero_image_alt, e.content::jsonb, e.seo_title, e.seo_description, now()
FROM author CROSS JOIN (VALUES

(
    'Canva vs Figma: two different jobs at the same price',
    'canva-vs-figma',
    'Within a dollar of each other and almost never a real choice. One makes a decent asset today; the other designs an interface. Knowing which you need takes one question.',
    '/images/saas-comparison-v2.svg',
    'Two design tools shown side by side at nearly the same monthly price',
    $json$[
      {"type":"paragraph","text":"Canva Pro works out at 15 USD a month, Figma Professional is 16. People compare them constantly and almost nobody should. They do different jobs, and the price being nearly identical is a coincidence rather than a signal."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Sketch Standard: 12 USD per editor per month billed yearly. Mac only, and that single fact decides more choices here than any feature.","Canva Pro: Canva publishes 180 USD a year for one person and no monthly rate at all, so 15 USD is that divided by twelve.","Figma Professional: 16 USD per full seat per month billed annually. Monthly billing is 20."]},
      {"type":"heading","heading":"The question that settles it"},
      {"type":"callout","heading":"One question","text":"Does the thing you are making end up as a picture, or as a screen somebody uses? A picture — a post, a deck, a flyer — is Canva. A screen is Figma. There is very little middle ground and both tools are bad at the other's job."},
      {"type":"heading","heading":"Canva, for people who do not design"},
      {"type":"paragraph","text":"Templates, brand kits, background removal, and a result that looks acceptable in ten minutes without any training. That is the whole product and it is a good one. It will not design an application, and pretending otherwise wastes a week."},
      {"type":"heading","heading":"Figma, for interfaces"},
      {"type":"paragraph","text":"The best collaboration in this category and the standard everyone who does this professionally already knows. The seat economics are the part people miss: developers and reviewers get cheaper seat types, so a small team pays far less than the headline suggests."},
      {"type":"heading","heading":"Where Sketch still wins"},
      {"type":"paragraph","text":"It is cheaper and it is still sharp at interface work. It is also Mac only. If your team is mixed, the conversation ends there, and no feature comparison changes it."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not run a design team on any of these. This compares what each vendor publishes against a specific brief, and the brief is the small business choosing its first paid design tool."}
    ]$json$,
    'Canva vs Figma: which design tool to buy',
    'Within a dollar of each other and rarely a real choice. One question about what you are making settles it in ten seconds.'
),

(
    'Webflow vs Squarespace vs Framer: which website builder',
    'webflow-vs-squarespace-vs-framer',
    'Ten, fifteen and nineteen dollars. The cheapest is the newest, the dearest asks least of you, and the middle one hides its real entry price one tier up.',
    '/images/saas-comparison-v2.svg',
    'Three website builders compared at their entry monthly prices',
    $json$[
      {"type":"paragraph","text":"If nobody at your business is technical, this is a short conversation: Squarespace, and stop reading. The other two are better tools for people who will use them properly, and worse tools for people who will not."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Framer Basic: 10 USD per site per month billed yearly, with a free tier on a Framer domain.","Webflow Basic: 15 USD per month billed yearly — and Basic has no CMS.","Squarespace Basic: 19 USD per month at monthly billing, with hosting, domain and payments in the one bill."]},
      {"type":"heading","heading":"The trap in the middle option"},
      {"type":"paragraph","text":"Webflow Basic excludes the CMS. If you will ever want a blog, a case-study list or anything else that is a collection rather than a page, Basic is not your price — the next tier is. That is not a hidden fee, it is on their page, but it is the single most common surprise in this category."},
      {"type":"callout","heading":"The test that decides it","text":"Who changes the opening hours when they change? If the answer is somebody non-technical, Squarespace, and the four dollars is the cheapest insurance you will buy this year. If it is you and you enjoy it, Framer or Webflow."},
      {"type":"heading","heading":"Framer is the cheapest and the most opinionated"},
      {"type":"paragraph","text":"A design tool that publishes the site directly, so what you arrange is what ships. Choose it when the look matters more than the content model. Its free tier on a Framer domain is a genuine way to find out whether you like it."},
      {"type":"heading","heading":"Squarespace asks the least"},
      {"type":"paragraph","text":"It is the dearest here and the one you can hand to somebody who has never built a site. Selling through it costs 2% on this tier, which the next tier removes — worth knowing before you build a shop on it."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not migrated a site between these. Export quality and how each behaves after two years of edits are not things a pricing page tells you, and we would rather say so than guess."}
    ]$json$,
    'Webflow vs Squarespace vs Framer compared',
    'Ten, fifteen and nineteen dollars a month. One question about who edits the site decides it, and Webflow Basic has no CMS.'
),

(
    'Freshdesk vs Help Scout vs Tidio: which support tool',
    'freshdesk-vs-help-scout-vs-tidio',
    'Within two dollars of each other and built for three different shapes of business. The price is not the decision; where your questions arrive is.',
    '/images/saas-comparison-v2.svg',
    'Three customer support tools compared at their entry prices',
    $json$[
      {"type":"paragraph","text":"These three sit within two dollars of one another, so anyone leading with price is padding. What separates them is when your customers ask questions: before they buy, or after."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Freshdesk Growth: 23 USD per agent per month at monthly billing, with an AI agent included for the first 500 sessions.","Tidio Starter: 24.17 USD per month, capped at 100 billable conversations.","Help Scout Standard: 25 USD per user per month."]},
      {"type":"callout","heading":"The question that decides it","text":"Do most of your questions arrive before the sale or after it? Before — a visitor on your site wondering whether to buy — is Tidio, because that is live chat and the others are queues. After is Help Scout or Freshdesk."},
      {"type":"heading","heading":"Tidio, for questions that arrive before the money"},
      {"type":"paragraph","text":"Live chat first, help desk second, which is the right way round for a shop or a service site. The Starter tier stops at 100 billable conversations a month, and that cap rather than the price is what decides whether it fits."},
      {"type":"heading","heading":"Help Scout, for a small team that hates ticketing"},
      {"type":"paragraph","text":"A shared inbox that looks like email to the customer and like a queue to you: no ticket numbers, no portal, nothing that makes a person feel they are talking to a system. The simplest of the three by a distance, which is the entire reason to choose it."},
      {"type":"heading","heading":"Freshdesk, for when you will outgrow the others"},
      {"type":"paragraph","text":"The heaviest here and the one that scales furthest before you hit a wall: a proper portal, a knowledge base, automation. If you expect more than a handful of agents, it is the one you will not have to leave."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"Zoho Desk belongs in this comparison and is missing from it, because Zoho publishes no price that can be verified. We would rather leave a good product out than print a number nobody read."}
    ]$json$,
    'Freshdesk vs Help Scout vs Tidio compared',
    'Within two dollars of each other. Where your questions arrive — before the sale or after it — decides which of the three you want.'
),

(
    'Ahrefs vs Semrush: and the gap between them',
    'ahrefs-vs-semrush',
    'Twenty-nine dollars against a hundred and seventeen. The interesting part is what sits in between, and whether you need it at all.',
    '/images/saas-comparison-v2.svg',
    'Two SEO platforms compared, with a third option between them',
    $json$[
      {"type":"paragraph","text":"Ahrefs Starter is 29 USD a month. Semrush is 117.33 billed annually. That is not a comparison between two similar things — it is the shape of the whole market, where there is one cheap way in and then a cliff."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Ahrefs Starter: 29 USD per month, one user included. The cheapest genuine way into a serious backlink index.","SE Ranking Core: 103.20 USD per month billed annually; 129 monthly. Ten projects and a manager seat.","Semrush SEO: 117.33 USD per month billed annually; 139 monthly."]},
      {"type":"heading","heading":"What Ahrefs Starter withholds"},
      {"type":"paragraph","text":"Starter is not a smaller version of Ahrefs, it is a slice of it. Keyword and backlink research, yes. The site audit and rank tracking that the higher tiers exist to sell, no. If you only want to see who links to a competitor, that slice is all you need and it is a bargain."},
      {"type":"callout","heading":"The test that decides it","text":"Is SEO your job, or a task on your list? A task — you check a few keywords a month and want to see competitors' links — is Ahrefs Starter, and the other two are money you will not use. A job is a full platform, and then the question is only which."},
      {"type":"heading","heading":"Semrush is more complete and it costs more"},
      {"type":"paragraph","text":"The largest toolkit here by some distance, and hard to justify at any smaller scale. If you run paid and organic together, the fact that it covers both is what earns the difference."},
      {"type":"heading","heading":"SE Ranking sits in the gap"},
      {"type":"paragraph","text":"A full platform at fourteen dollars less, with ten projects included. It is the sensible middle when Ahrefs Starter runs out of room and Semrush is more money than the work justifies. It is also, we should say plainly, the one product in this comparison whose link earns us a commission — and it is still not the one most readers here should buy."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"Index size and data quality are the things that actually separate these platforms, and they cannot be compared from a pricing page. Anyone ranking them on that from public information is guessing."}
    ]$json$,
    'Ahrefs vs Semrush, and what sits between them',
    'Twenty-nine dollars against a hundred and seventeen. One question about whether SEO is your job or a task settles most of it.'
),

(
    'Teachable vs Thinkific vs Gumroad: what selling a course really costs',
    'teachable-vs-thinkific-vs-gumroad',
    'The cheapest of the three takes the biggest cut, and the dearest takes none. Which is cheapest depends entirely on how much you sell.',
    '/images/saas-comparison-v2.svg',
    'Three course platforms compared by subscription and transaction fee',
    $json$[
      {"type":"paragraph","text":"You cannot compare these on the monthly price, and the monthly price is what every other comparison shows you. Gumroad is free and takes 10%. Teachable is 29 and takes 7.5%. Thinkific is 40 and takes nothing. At some volume each of them is the cheapest."},
      {"type":"heading","heading":"The prices, and the cut"},
      {"type":"unordered_list","items":["Gumroad: nothing per month, 10% plus 50 cents of every sale through your own links — 30% if a buyer finds you through Gumroad Discover.","Teachable Starter: 29 USD a month billed annually, 39 billed monthly, plus a 7.5% transaction fee that the higher tiers drop.","Thinkific Basic: 40 USD a month billed annually, with no transaction fee on top."]},
      {"type":"heading","heading":"Where the lines cross"},
      {"type":"paragraph","text":"At nothing sold, Gumroad costs nothing and the others cost their subscription. At 500 USD a month, Gumroad takes about 50 and Teachable about 66 all in — still close. At 1,000, Gumroad takes 100 and beats nothing here. At 2,000, Thinkific's flat 40 is the cheapest of the three by a wide margin and gets cheaper every month after that."},
      {"type":"callout","heading":"The test that decides it","text":"How much do you expect to sell in a month? Under a few hundred, Gumroad and pay nothing until you do. Over about a thousand, Thinkific and stop paying a percentage. In between, work out both — this is arithmetic, not preference."},
      {"type":"heading","heading":"They are also not the same product"},
      {"type":"paragraph","text":"Gumroad sells a file. Teachable and Thinkific host a course: lessons, progress, completion, a school on your own domain. If you are selling one PDF, the course platforms are overkill at any price."},
      {"type":"paragraph","text":"Between the two, Thinkific has the more complete course-building tools and Teachable the better storefront and upsells. That is the real trade once the money question is settled."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"Payment processing fees sit on top of all three and are roughly equal across them, so they do not change the comparison — but they do mean none of these numbers is what lands in your account."}
    ]$json$,
    'Teachable vs Thinkific vs Gumroad: the real cost',
    'The cheapest takes the biggest cut and the dearest takes none. Where the lines cross is arithmetic, and we have done it.'
)

) AS e(title, slug, description, hero_image_url, hero_image_alt, content, seo_title, seo_description)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT e.id, c.id, m.position
FROM (VALUES
 ('canva-vs-figma','design-tools',0),
 ('webflow-vs-squarespace-vs-framer','website-builder',0),
 ('freshdesk-vs-help-scout-vs-tidio','help-desk',0),
 ('ahrefs-vs-semrush','seo-tools',0),
 ('teachable-vs-thinkific-vs-gumroad','course-platform',0)
) AS m(entry_slug, category_slug, position)
JOIN editorial.entries e ON e.slug = m.entry_slug
JOIN catalog.categories c ON c.slug = m.category_slug AND c.vertical_key = 'saas'
ON CONFLICT DO NOTHING;

INSERT INTO editorial.entry_products (entry_id, product_id, position)
SELECT e.id, p.id, m.position
FROM (VALUES
 ('canva-vs-figma','sketch-standard',0),
 ('canva-vs-figma','canva-pro',1),
 ('canva-vs-figma','figma-professional',2),
 ('webflow-vs-squarespace-vs-framer','framer-basic',0),
 ('webflow-vs-squarespace-vs-framer','webflow-basic',1),
 ('webflow-vs-squarespace-vs-framer','squarespace-basic',2),
 ('freshdesk-vs-help-scout-vs-tidio','freshdesk-growth',0),
 ('freshdesk-vs-help-scout-vs-tidio','tidio-starter',1),
 ('freshdesk-vs-help-scout-vs-tidio','help-scout-standard',2),
 ('ahrefs-vs-semrush','ahrefs-starter',0),
 ('ahrefs-vs-semrush','se-ranking-core',1),
 ('ahrefs-vs-semrush','semrush-seo',2),
 ('teachable-vs-thinkific-vs-gumroad','gumroad',0),
 ('teachable-vs-thinkific-vs-gumroad','teachable-starter',1),
 ('teachable-vs-thinkific-vs-gumroad','thinkific-basic',2)
) AS m(entry_slug, product_slug, position)
JOIN editorial.entries e ON e.slug = m.entry_slug
JOIN catalog.products p ON p.slug = m.product_slug
ON CONFLICT DO NOTHING;

INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT s.id, t.id, m.position
FROM (VALUES
 ('canva-vs-figma','how-to-choose-business-software',0),
 ('webflow-vs-squarespace-vs-framer','how-unsolero-ranks-software',0),
 ('freshdesk-vs-help-scout-vs-tidio','how-to-choose-business-software',0),
 ('ahrefs-vs-semrush','how-unsolero-ranks-software',0),
 ('teachable-vs-thinkific-vs-gumroad','mailchimp-alternatives',0),
 ('teachable-vs-thinkific-vs-gumroad','how-to-choose-business-software',1)
) AS m(source_slug, target_slug, position)
JOIN editorial.entries s ON s.slug = m.source_slug
JOIN editorial.entries t ON t.slug = m.target_slug
ON CONFLICT DO NOTHING;
