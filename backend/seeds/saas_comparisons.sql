-- Seven head-to-head comparisons.
--
-- The catalog holds fifty-three products and, until now, exactly one page that
-- put two of them side by side. "X vs Y" is what people actually type when they
-- have narrowed the field and cannot break the tie, and it is the one query
-- where this site's discipline is worth something: the prices here were read
-- from each vendor on a stated date, with the billing basis attached, so a
-- comparison does not quietly turn a monthly rate against an annual one and
-- call the difference a saving.
--
-- Every figure below comes from catalog.products. None is typed from memory.
--
-- Three of the seven feature a Zoho product, because Zoho is the one vendor
-- whose links currently earn anything. Two of those three do not conclude in
-- Zoho's favour. That is the point: a comparison site whose partner always
-- wins is not a comparison site, and the disclosure at the foot of every page
-- is worth nothing if the verdict is decided before the reading starts.

WITH author AS (
    SELECT id FROM editorial.authors WHERE slug = 'andon-pediev'
)
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description,
    published_at
)
SELECT author.id, 'comparison', 'published', entry.title, entry.slug,
       entry.description, entry.hero_image_url, entry.hero_image_alt,
       entry.content::jsonb, entry.seo_title, entry.seo_description, now()
FROM author CROSS JOIN (VALUES

-- ============================================================ 1. CRM =====
(
    'Zoho CRM vs HubSpot: the same price, two different bets',
    'zoho-crm-vs-hubspot',
    'Both cost 20 USD per user per month. That is where the similarity stops. One is a cheap seat in a large suite, the other is a cheap door into an expensive one.',
    '/images/saas-comparison-v2.svg',
    'Two software plans placed side by side at the same monthly price',
    $json$[
      {"type":"paragraph","text":"Zoho CRM Standard and HubSpot Starter Customer Platform both cost 20 USD per user per month. The interesting question is not which is cheaper today, because neither is. It is which one costs more in two years, and the answer is almost always HubSpot."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Zoho CRM Standard: 20 USD per user per month.","HubSpot Starter Customer Platform: 20 USD per user per month."]},
      {"type":"paragraph","text":"A tie on the entry tier is unusual and it is worth understanding why it happens. Zoho prices low across its whole range because it wants you inside the suite. HubSpot prices its entry tier low because it wants you inside the funnel. Those are not the same intention, and they diverge sharply the moment you outgrow the first plan."},
      {"type":"heading","heading":"What actually separates them"},
      {"type":"paragraph","text":"HubSpot is the better product to use. The interface is calmer, the onboarding is genuinely good, and the reporting makes sense without a manual. If you put two people who had never seen either in front of both, most would prefer HubSpot by the end of the first hour."},
      {"type":"paragraph","text":"Zoho CRM is denser and less pleasant, and it is attached to an ecosystem that covers invoicing, email marketing, help desk, projects, bookings and a dozen other jobs at prices nobody else matches. If you expect to need three of those things, the CRM stops being the decision and starts being one line on a bill."},
      {"type":"paragraph","text":"The step up is where the two part company. Zoho's tiers rise gently. HubSpot's next tier is a different order of money, and the features most teams discover they need — real automation, proper reporting, more than a handful of contacts on marketing plans — live above the Starter line rather than in it."},
      {"type":"callout","heading":"The test that decides it","text":"Will you buy more software from this vendor in the next two years? If yes, Zoho, because the suite is the product and the CRM is the entry point. If no, and this CRM is the only thing you will ever run, HubSpot is nicer to live in and you can leave whenever you like."},
      {"type":"heading","heading":"Where each one is genuinely better"},
      {"type":"unordered_list","items":["Choose HubSpot if you are one to five people, the CRM is your only tool, and you want it working this afternoon without configuration.","Choose Zoho CRM if you already run any Zoho product, if invoicing and email will follow, or if you would rather spend an evening configuring something than pay double next year.","Choose neither if you are a one-person business selling services. Bigin, from the same company, costs 9 USD per user per month and is closer to the size of the problem."]},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"This compares what each vendor publishes against a specific brief. It is not a verdict from a year of daily use, and anyone who claims one for two products this large is describing a preference. The prices are ours to keep current; the fit is yours to judge."}
    ]$json$,
    'Zoho CRM vs HubSpot: same price, different bets',
    'Both cost 20 USD per user per month. What separates them is what happens when you outgrow the entry tier, and the answer is not close.'
),

-- ===================================================== 2. AUTOMATION =====
(
    'Zapier vs Make: pay more for the library or less for the canvas',
    'zapier-vs-make',
    'Zapier connects to more things. Make is cheaper and easier to reason about once a workflow branches. Which matters depends on one question about the tools you already own.',
    '/images/saas-comparison-v2.svg',
    'Two automation workflows drawn side by side, one as a list and one as a diagram',
    $json$[
      {"type":"paragraph","text":"If the app you need is only on Zapier, the comparison is over and you pay the difference. If both connect to everything you own, Make does the same work for less money and shows it to you more clearly. Check the connector list first; everything else is secondary."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Make Core: 12 USD per month, billed annually, for 10,000 credits a month. Billed monthly it is 16 USD.","Zapier Professional: 19.99 USD per month, billed annually. Priced by task volume, so it rises with use.","n8n Starter: 20 USD per month, billed annually, for 2,500 workflow executions. Open source and self-hostable."]},
      {"type":"paragraph","text":"Both meters are easy to misread. Zapier counts tasks, meaning each successful step of each run. Make counts credits on a similar basis. A workflow with six steps consumes six of whatever the vendor is counting, every time it fires, and the plan you priced on the entry tier is not the plan you end up on."},
      {"type":"heading","heading":"What actually separates them"},
      {"type":"paragraph","text":"Zapier's advantage is the connector library, and it is a real one. It is the largest in the category by a distance, and for an obscure tool the honest answer is that Zapier will have it and Make may not."},
      {"type":"paragraph","text":"Make's advantage is the canvas. A workflow is drawn as a diagram rather than listed as steps, which sounds cosmetic until something branches. A five-branch flow in a linear builder is a wall of indented text; the same flow drawn out can be understood at a glance, including by whoever inherits it."},
      {"type":"paragraph","text":"Learning cost runs the other way. Zapier is the easier first hour by some margin. Make asks more of you up front and gives more back later."},
      {"type":"callout","heading":"The test that decides it","text":"Open Zapier's app directory and Make's, and search for the two least famous tools you use. If Make has both, buy Make. If it has neither, buy Zapier and stop reading comparisons."},
      {"type":"heading","heading":"The third option most people skip"},
      {"type":"paragraph","text":"n8n costs 20 USD a month on its cloud and nothing at all if you run it yourself. Self-hosting is real work and needs someone comfortable with a server, but the workflows and the data inside them never leave your own machine. If that matters to you — and for anything touching customer records it might — it is the only one of the three that offers it."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not benchmarked reliability. All three will drop a run occasionally and all three will tell you about it. What we can tell you is what each publishes and what each meter counts, which is the part vendors make hardest to compare."}
    ]$json$,
    'Zapier vs Make: which automation tool to buy',
    'Zapier has the bigger app library. Make is cheaper and clearer once a workflow branches. One check on your own tools settles it.'
),

-- ======================================================= 3. PAYMENTS =====
(
    'Stripe vs Paddle: who files your sales tax',
    'stripe-vs-paddle',
    'Stripe charges 2.9% plus 30 cents. Paddle charges 5% plus 50 cents. The gap is not a markup — it is the price of somebody else being legally responsible for your VAT.',
    '/images/saas-comparison-v2.svg',
    'Two payment routes drawn side by side, one passing through a tax authority',
    $json$[
      {"type":"paragraph","text":"Neither charges a monthly fee. Stripe takes 2.9% plus 30 cents per successful domestic card charge. Paddle takes 5% plus 50 cents per checkout transaction. Two points looks like a lot until you find out what it buys, and then for most one-person software businesses it looks cheap."},
      {"type":"heading","heading":"Merchant of record is the whole comparison"},
      {"type":"paragraph","text":"With Stripe you are the seller. You are the one who owes sales tax and VAT in every jurisdiction where you have customers, you are the one who has to register where thresholds are crossed, and you are the one who files. Stripe moves money; the obligations stay with you."},
      {"type":"paragraph","text":"With Paddle, Paddle is the seller. It sells your product to your customer, collects the tax, registers where it must, and files worldwide. You invoice Paddle; Paddle deals with the tax authorities. That is what the extra two points are for."},
      {"type":"paragraph","text":"Lemon Squeezy charges the same as Paddle, 5% plus 50 cents, and works the same way. It is the friendliest of the three to set up. It is also now part of Stripe, which is worth weighing: well supported today, but its independent direction no longer exists."},
      {"type":"callout","heading":"The test that decides it","text":"Do you know, right now, which countries you owe VAT in? If the honest answer is no, and you sell software to people abroad, you are not choosing between 2.9% and 5%. You are choosing between 2.9% plus an accountant and 5% with the problem removed."},
      {"type":"heading","heading":"When Stripe is clearly right"},
      {"type":"unordered_list","items":["You sell to one country and stay under its registration threshold.","You sell physical goods or services rather than digital products, where the rules differ and merchant-of-record cover matters less.","You already have an accountant handling filings, in which case the two points buys you something you have already paid for.","You need the API. Stripe's is the deepest in this category by a wide margin and the difference is not close."]},
      {"type":"heading","heading":"When Paddle is clearly right"},
      {"type":"unordered_list","items":["You sell digital products or subscriptions across borders.","You are one person and have no wish to become an expert on EU VAT thresholds.","Your revenue is small enough that an accountant's fee would eat more than two points."]},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"This is not tax advice and we are not licensed to give any. It sets out what each vendor publishes about its own role, which is a fact, and leaves the consequences for your situation to somebody qualified to judge them."}
    ]$json$,
    'Stripe vs Paddle: the merchant of record question',
    'Stripe is 2.9% + 30c and the tax is yours. Paddle is 5% + 50c and the tax is theirs. Which is cheaper depends on where you sell.'
),

-- ===================================================== 4. SCHEDULING =====
(
    'Calendly vs Cal.com vs Zoho Bookings: the booking link decision',
    'calendly-vs-cal-com-vs-zoho-bookings',
    'Three booking tools within four dollars of each other. The price is not the decision — recognition, ownership and what else you already run are.',
    '/images/saas-comparison-v2.svg',
    'Three booking calendars shown side by side with their monthly prices',
    $json$[
      {"type":"paragraph","text":"These three sit within four dollars of one another, so anyone leading with price is padding. Calendly is the one your customer already recognises. Cal.com is the one you can host yourself. Zoho Bookings is the one that already knows your CRM. Pick the sentence that describes your problem."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Zoho Bookings Basic: 8 USD per user per month. A free tier exists.","Calendly Standard: 10 USD per seat per month. Seats are needed for anyone who connects a calendar and hosts meetings.","Cal.com Teams: 12 USD per user per month, billed yearly. Free forever for an individual."]},
      {"type":"heading","heading":"Calendly is winning on a thing that is hard to argue with"},
      {"type":"paragraph","text":"A booking link nobody has to be taught to use is worth more than it sounds. Calendly is the name people recognise in an email, and recognition removes hesitation at the exact moment you are asking somebody to commit time. For client-facing work that is a real advantage and it does not show up on any feature table."},
      {"type":"heading","heading":"Cal.com is the one you can own"},
      {"type":"paragraph","text":"It is open source and can run on your own server, which means the record of who met whom and when never leaves your infrastructure. That is worth nothing to most businesses and a great deal to a few. It is also the most flexible of the three if you are willing to work for it, and it is free forever for a single person, which makes it the strongest free tier here by some distance."},
      {"type":"heading","heading":"Zoho Bookings wins on one condition"},
      {"type":"paragraph","text":"It is the cheapest, and it is the only one that arrives already connected to a CRM, invoicing and email from the same vendor. If you run Zoho, that connection saves a step every time. If you do not, none of that applies and you are choosing an unfamiliar tool to save two dollars, which is not a good trade."},
      {"type":"callout","heading":"The test that decides it","text":"Are you already inside Zoho? Then Zoho Bookings. Do you need the booking data on your own server? Then Cal.com. Neither? Calendly, and stop thinking about it — the two dollars is not worth the deliberation."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"Calendly is the most expensive of the three at the entry tier and we are still recommending it for the common case. That is a judgement about what recognition is worth to a business asking strangers for meetings, not a measurement. We have marked it as a judgement rather than dressing it up as a finding."}
    ]$json$,
    'Calendly vs Cal.com vs Zoho Bookings compared',
    'Three booking tools within four dollars. Recognition, ownership or an existing suite decides it, not the price. Read the one-question test.'
),

-- ====================================================== 5. ANALYTICS =====
(
    'Fathom vs Simple Analytics vs Umami: privacy analytics compared',
    'fathom-vs-simple-analytics-vs-umami',
    'Three analytics tools that need no cookie banner, at 100,000 pageviews a month. One of them is free at that volume, which changes the shape of the comparison.',
    '/images/saas-comparison-v2.svg',
    'Three analytics dashboards shown side by side at the same traffic volume',
    $json$[
      {"type":"paragraph","text":"All three do the same core job: web analytics without cookies, without a consent banner, and without handing your visitors to an advertising company. Priced at the same volume, one of them is free, which is the fact that should decide most of these comparisons and rarely gets stated plainly."},
      {"type":"heading","heading":"The prices, all at 100,000 a month"},
      {"type":"unordered_list","items":["Umami Cloud Pro: 20 USD per month. The Hobby tier is free up to 100,000 events a month.","Fathom Analytics: 15 USD per month at 100,000 pageviews, covering up to 50 sites.","Simple Analytics: 20 USD per month at 100,000 pageviews."]},
      {"type":"paragraph","text":"Quoting all three at one volume is the only way this comparison means anything. Every tool here charges by traffic, so a bare monthly price compares nothing at all unless the traffic behind it is stated."},
      {"type":"paragraph","text":"One caveat on Umami: it counts events rather than pageviews. A busy single-page application registers more events than pageviews, so its free ceiling arrives sooner than the number suggests. For an ordinary site the two are close enough."},
      {"type":"heading","heading":"What actually separates them"},
      {"type":"paragraph","text":"Umami is open source and can be self-hosted, which makes it the only one where the data can live entirely on your own server. Its free cloud tier at 100,000 events is where the other two start charging. If cost is the constraint, the comparison ends here."},
      {"type":"paragraph","text":"Fathom is the most polished, covers up to 50 sites on one plan, and is the cheapest of the paid options. If you run several small sites rather than one large one, its per-site economics are the best here by a wide margin."},
      {"type":"paragraph","text":"Simple Analytics is a single dashboard with almost nothing to configure, and that is the entire product rather than a limitation of it. Choose it when you want one number a day and no invitation to build reports."},
      {"type":"callout","heading":"The test that decides it","text":"Under 100,000 a month? Umami, free, and revisit when you outgrow it. Several sites? Fathom. One site and you never want to think about analytics again? Simple Analytics."},
      {"type":"heading","heading":"Why none of these is Google Analytics"},
      {"type":"paragraph","text":"All three give you less than Google Analytics does, and that is the trade. No cookie banner, no consent gate, a dashboard that fits on one screen, and no advertising company in the middle. If you need conversion funnels and audience segments for paid acquisition, none of these replaces what you already have."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not audited any of their privacy claims against their actual behaviour. Each publishes how it works; we have read those claims and not verified them independently, and we would rather say so than imply an audit we did not do."}
    ]$json$,
    'Fathom vs Simple Analytics vs Umami compared',
    'Three cookie-free analytics tools priced at 100,000 pageviews a month. One is free at that volume, which settles most of the comparison.'
),

-- ====================================================== 6. ECOMMERCE =====
(
    'Shopify vs BigCommerce: identical price, different bill',
    'shopify-vs-bigcommerce',
    'Both entry plans cost 29 USD a month. What separates them is the transaction fee, the sales ceiling, and how much of your setup you can take with you.',
    '/images/saas-comparison-v2.svg',
    'Two online store dashboards shown side by side at the same monthly price',
    $json$[
      {"type":"paragraph","text":"Shopify Basic and BigCommerce Core both cost 29 USD a month. They are not the same 29 dollars. BigCommerce charges no platform transaction fee on top when you use a supported payment provider; Shopify does unless you use Shopify Payments. On any real volume, that line matters more than the subscription."},
      {"type":"heading","heading":"The prices"},
      {"type":"unordered_list","items":["Ecwid Starter: 5 USD per month, billed annually. Capped at 10 products.","Shopify Basic: 29 USD per month. Card rates start at 2.9% plus 30 cents. The figure was read with the billing control set to pay yearly; monthly is higher.","BigCommerce Core: 29 USD per month, billed annually. Capped at 30,000 USD of trailing annual sales."]},
      {"type":"heading","heading":"The two ceilings nobody mentions"},
      {"type":"paragraph","text":"BigCommerce Core stops at 30,000 USD of trailing twelve-month sales and upgrades itself past that. It is not a penalty, but it is a cost you should plan for rather than discover: crossing 2,500 USD a month in revenue changes your bill."},
      {"type":"paragraph","text":"Shopify's equivalent is the payment processor. Use Shopify Payments and there is no platform fee. Use anything else and there is, on every order, forever. If you are committed to a processor Shopify does not own, price that in before comparing subscriptions."},
      {"type":"heading","heading":"What actually separates them"},
      {"type":"paragraph","text":"Shopify has the largest app ecosystem in the category and the fewest unpleasant surprises at scale. It is the default answer and defensibly so: whatever you need to bolt on, somebody has already built it, and whoever you hire will already know it."},
      {"type":"paragraph","text":"BigCommerce is more open. It puts more in the box before you start paying for apps, its API is stronger, and it does not push you toward its own payment processor. That is worth real money to a store that wants control, and worth nothing to a store that just wants to sell things by Friday."},
      {"type":"callout","heading":"The test that decides it","text":"Will you use Shopify Payments? If yes, Shopify, and the transaction fee argument disappears. If you are tied to another processor, BigCommerce, and the two platforms stop looking like the same price."},
      {"type":"heading","heading":"The cheaper option worth a look"},
      {"type":"paragraph","text":"If you have a website already and simply want to sell from it, Ecwid Starter is 5 USD a month and drops a store into a site you own. It stops at 10 products, and that cap rather than the price is what decides whether it fits. Below ten products it makes both of the others look like overkill."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"We have not run a store on either for a year. This compares published pricing, published fees and published limits against a small-business brief. Migration pain, support quality and the state of any particular app are outside what we can verify."}
    ]$json$,
    'Shopify vs BigCommerce: which store platform',
    'Both cost 29 USD a month. The transaction fee, the sales ceiling and the payment processor decide it, not the subscription.'
),

-- ================================================ 7. FREE INVOICING =====
(
    'Zoho Invoice vs Wave: which free invoicing tool',
    'zoho-invoice-vs-wave',
    'Both cost nothing, permanently, with no trial and no card. One of them does bookkeeping and the other does not, and that single difference decides it.',
    '/images/saas-comparison-v2.svg',
    'Two invoices side by side, both marked as costing nothing',
    $json$[
      {"type":"paragraph","text":"Zoho Invoice and Wave Starter are both free. Not free-for-a-month, not free-until-you-have-customers: free, with no paid tier to graduate into in Zoho Invoice's case at all. If you are a freelancer sending a handful of invoices, one of these is the correct answer and paying for anything else is a mistake."},
      {"type":"heading","heading":"What each one is"},
      {"type":"unordered_list","items":["Zoho Invoice: free. Invoices, estimates, expenses and time tracking. It is not accounting — no bank reconciliation, no full ledger.","Wave Starter: free for one user, and it does include double-entry bookkeeping, which the other does not. Payments and payroll are charged separately."]},
      {"type":"heading","heading":"The difference that decides it"},
      {"type":"paragraph","text":"Wave keeps books. Zoho Invoice sends invoices. If your accountant wants a ledger at the end of the year, Wave hands them one and Zoho Invoice hands them a list of invoices and a shrug."},
      {"type":"paragraph","text":"Against that, Zoho Invoice has time tracking built in, which matters if you bill by the hour, and it sits inside a suite. When invoicing alone stops being enough, Zoho Books picks up where it leaves off at 12 USD a month with your data already in place. Wave's own step up is 190 USD a year for receipt capture and automation."},
      {"type":"callout","heading":"The test that decides it","text":"Does anyone need to see your books, or only your invoices? Books, and it is Wave. Invoices, and Zoho Invoice is lighter and tracks your hours for you."},
      {"type":"heading","heading":"Why free is not the catch you expect"},
      {"type":"paragraph","text":"Both companies are open about the trade. Wave earns on payment processing and payroll; Zoho earns when invoicing turns into accounting, CRM or something else in the suite. Neither is selling your data to make the sums work, and both have been doing this long enough that the model is proven rather than promised."},
      {"type":"heading","heading":"What we are not telling you"},
      {"type":"paragraph","text":"Wave's coverage of banks and tax rules is strongest in North America, and we have not tested how well it handles a European business's requirements. If you are outside the US and Canada, check that before committing a year of records to it."}
    ]$json$,
    'Zoho Invoice vs Wave: free invoicing compared',
    'Both are free permanently. Wave keeps books, Zoho Invoice tracks time. One question about who reads your accounts decides it.'
)

) AS entry(title, slug, description, hero_image_url, hero_image_alt, content, seo_title, seo_description)
ON CONFLICT (slug) DO NOTHING;

-- Link each comparison to the products it weighs. This is what puts the entry
-- on those product pages, and it is the difference between internal linking
-- that means something and a related-articles box filled by date.
INSERT INTO editorial.entry_products (entry_id, product_id, position)
SELECT entries.id, products.id, mapping.position
FROM (VALUES
 ('zoho-crm-vs-hubspot','zoho-crm-standard',0),
 ('zoho-crm-vs-hubspot','hubspot-starter-customer-platform',1),
 ('zoho-crm-vs-hubspot','bigin-express',2),
 ('zapier-vs-make','make-core',0),
 ('zapier-vs-make','zapier-professional',1),
 ('zapier-vs-make','n8n-starter',2),
 ('stripe-vs-paddle','stripe',0),
 ('stripe-vs-paddle','paddle',1),
 ('stripe-vs-paddle','lemon-squeezy',2),
 ('calendly-vs-cal-com-vs-zoho-bookings','zoho-bookings-basic',0),
 ('calendly-vs-cal-com-vs-zoho-bookings','calendly-standard',1),
 ('calendly-vs-cal-com-vs-zoho-bookings','cal-com-teams',2),
 ('fathom-vs-simple-analytics-vs-umami','fathom-analytics',0),
 ('fathom-vs-simple-analytics-vs-umami','simple-analytics',1),
 ('fathom-vs-simple-analytics-vs-umami','umami-cloud-pro',2),
 ('shopify-vs-bigcommerce','shopify-basic',0),
 ('shopify-vs-bigcommerce','bigcommerce-core',1),
 ('shopify-vs-bigcommerce','ecwid-starter',2),
 ('zoho-invoice-vs-wave','zoho-invoice',0),
 ('zoho-invoice-vs-wave','wave-starter',1),
 ('zoho-invoice-vs-wave','zoho-books-standard',2)
) AS mapping(entry_slug, product_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.products products ON products.slug = mapping.product_slug
ON CONFLICT DO NOTHING;

INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT entries.id, categories.id, mapping.position
FROM (VALUES
 ('zoho-crm-vs-hubspot','crm',0),
 ('zapier-vs-make','automation',0),
 ('stripe-vs-paddle','payments',0),
 ('stripe-vs-paddle','ecommerce-platform',1),
 ('calendly-vs-cal-com-vs-zoho-bookings','scheduling',0),
 ('fathom-vs-simple-analytics-vs-umami','analytics',0),
 ('shopify-vs-bigcommerce','ecommerce-platform',0),
 ('shopify-vs-bigcommerce','payments',1),
 ('zoho-invoice-vs-wave','accounting-invoicing',0)
) AS mapping(entry_slug, category_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.categories categories
  ON categories.slug = mapping.category_slug AND categories.vertical_key = 'saas'
ON CONFLICT DO NOTHING;

-- Related entries, chosen by what a reader of one would plausibly want next
-- rather than by category alone. Someone weighing Stripe against Paddle is
-- often about to open a store, so the store comparison follows.
INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT source.id, target.id, mapping.position
FROM (VALUES
 ('zoho-crm-vs-hubspot','hubspot-alternatives',0),
 ('zoho-crm-vs-hubspot','best-crm-small-agency',1),
 ('zoho-crm-vs-hubspot','how-unsolero-ranks-software',2),
 ('zapier-vs-make','how-to-choose-business-software',0),
 ('zapier-vs-make','how-unsolero-ranks-software',1),
 ('stripe-vs-paddle','shopify-vs-bigcommerce',0),
 ('stripe-vs-paddle','how-unsolero-ranks-software',1),
 ('calendly-vs-cal-com-vs-zoho-bookings','software-stack-small-agency',0),
 ('calendly-vs-cal-com-vs-zoho-bookings','how-to-choose-business-software',1),
 ('fathom-vs-simple-analytics-vs-umami','how-unsolero-ranks-software',0),
 ('fathom-vs-simple-analytics-vs-umami','how-to-choose-business-software',1),
 ('shopify-vs-bigcommerce','stripe-vs-paddle',0),
 ('shopify-vs-bigcommerce','how-to-choose-business-software',1),
 ('zoho-invoice-vs-wave','software-stack-small-agency',0),
 ('zoho-invoice-vs-wave','how-to-choose-business-software',1),
 ('clickup-vs-teamwork','zapier-vs-make',1),
 ('hubspot-alternatives','zoho-crm-vs-hubspot',2),
 ('best-crm-small-agency','zoho-crm-vs-hubspot',2),
 ('how-to-choose-business-software','zoho-crm-vs-hubspot',2)
) AS mapping(source_slug, target_slug, position)
JOIN editorial.entries source ON source.slug = mapping.source_slug
JOIN editorial.entries target ON target.slug = mapping.target_slug
ON CONFLICT DO NOTHING;
