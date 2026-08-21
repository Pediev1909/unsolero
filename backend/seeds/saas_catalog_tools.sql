-- The last four empty categories: automation, design tools, SEO tools and
-- team communication.
--
-- All twelve prices read on 2026-08-21 through a US connection, in USD, from
-- the vendor's own pricing page. Billing basis stated per product.
--
-- Three things here needed care rather than a regex.
--
-- Two vendors were running a half-price promotion when the page was read, and
-- showed the standing rate struck through beside it. Slack marks the struck
-- price with a class of its own, so which number was the real one could be
-- checked rather than guessed. Both products record the standing rate: a
-- promotion expires, and a comparison site that quotes one silently goes from
-- accurate to wrong without anybody touching it.
--
-- Make and Figma publish one figure and let a toggle change it without saying
-- which state it is in. Both were confirmed by flipping the toggle and watching
-- the number move -- Make Core reads 12 annually and 16 monthly, Figma's full
-- seat 16 annually and 20 monthly -- so the annual reading is established, not
-- assumed.
--
-- Canva is the one derived figure in the whole catalog. Its page publishes only
-- "180 US$ per year for one person" and offers no monthly view, so the monthly
-- price here is that divided by twelve. It is recorded at reduced confidence
-- and the division is stated on the source, because a number nobody read is
-- worth less than one that was, even when the arithmetic is trivial.

INSERT INTO catalog.brands (name, slug, website_url, country_code) VALUES
 ('Zapier','zapier','https://zapier.com','US'),
 ('Make','make','https://www.make.com','CZ'),
 ('n8n','n8n','https://n8n.io','DE'),
 ('Canva','canva','https://www.canva.com','AU'),
 ('Figma','figma','https://www.figma.com','US'),
 ('Sketch','sketch','https://www.sketch.com','NL'),
 ('Ahrefs','ahrefs','https://ahrefs.com','SG'),
 ('Semrush','semrush','https://www.semrush.com','US'),
 ('SE Ranking','se-ranking','https://seranking.com','US'),
 ('Slack','slack','https://slack.com','US'),
 ('Microsoft','microsoft','https://www.microsoft.com','US'),
 ('Google Workspace','google-workspace','https://workspace.google.com','US')
ON CONFLICT (slug) DO NOTHING;

INSERT INTO catalog.products (
    category_id, brand_id, name, slug, description, price_minor, currency,
    warranty_months, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score
)
SELECT categories.id, brands.id, f.name, f.slug, f.description, f.price_minor, 'USD', 0,
       f.quality, f.value, f.durability, f.beginner, f.advanced, 0, 0, f.portability
FROM (VALUES
 -- ---- automation ------------------------------------------------------
 ('automation','make','Make Core','make-core',
  'Automation drawn as a diagram rather than listed as steps, which makes a branching workflow far easier to follow than a linear builder does. Per month billed annually at 10,000 credits; the monthly rate is 16 USD. Cheaper than Zapier at the same volume and harder to learn.',
  1200,86,92,80,72,92,74),
 ('automation','zapier','Zapier Professional','zapier-professional',
  'The largest app library in this category by a wide margin, which is usually the only question that matters: if a tool you use is connected anywhere, it is connected here. Per month billed annually, and priced by task volume, so the figure rises with use.',
  1999,90,78,92,94,80,70),
 ('automation','n8n','n8n Starter','n8n-starter',
  'Open source and self-hostable, so the workflows and the data in them can stay on your own server. Per month billed annually for 2,500 executions on their cloud, with unlimited steps per run. Expects someone comfortable with an API; rewards them.',
  2000,86,84,76,62,96,94),
 -- ---- design tools ----------------------------------------------------
 ('design-tools','sketch','Sketch Standard','sketch-standard',
  'The original of this category, still sharp at interface work, and priced per editor per month billed yearly. It is Mac only, and that single fact decides the choice more often than any feature does.',
  1200,84,84,76,74,86,64),
 ('design-tools','canva','Canva Pro','canva-pro',
  'Design for people who do not design: templates, brand kits and background removal, aimed squarely at getting a decent-looking asset out today. Canva publishes 180 USD a year for one person and no monthly rate; the figure shown is that divided by twelve.',
  1500,84,88,88,96,60,66),
 ('design-tools','figma','Figma Professional','figma-professional',
  'The default for interface design and the best collaboration in this category, priced per full seat per month billed annually; the monthly rate is 20 USD. Cheaper seat types exist for developers and reviewers, which is what makes it affordable for a small team.',
  1600,94,82,88,74,94,76),
 -- ---- SEO tools -------------------------------------------------------
 ('seo-tools','ahrefs','Ahrefs Starter','ahrefs-starter',
  'The cheapest genuine way into a serious backlink and keyword index, at a fraction of what the full platforms charge. One user included, and the Starter tier withholds the site audit and rank tracking that the higher tiers exist to sell.',
  2900,92,88,90,78,88,70),
 ('seo-tools','se-ranking','SE Ranking Core','se-ranking-core',
  'A full platform at a lower price than Semrush, with ten projects and a manager seat included. Per month billed annually; the monthly rate is 129 USD. The sensible middle choice when Ahrefs Starter runs out of room and Semrush is too much money.',
  10320,82,74,78,80,82,72),
 ('seo-tools','semrush','Semrush SEO','semrush-seo',
  'The most complete toolkit here and the most expensive by some distance, at 117.33 USD a month billed annually against 139 USD monthly. Worth it when SEO is the job rather than a task, and hard to justify at any smaller scale.',
  11733,90,66,92,76,94,70),
 -- ---- team communication ----------------------------------------------
 ('team-communication','microsoft','Microsoft Teams Essentials','microsoft-teams-essentials',
  'Chat, calls and video for a small team at the lowest price in this category, per user per month paid yearly. It is the standalone plan: no Word, no Excel, no business email. Those come with Microsoft 365 Business Basic at 7 USD instead.',
  400,80,92,94,78,74,62),
 ('team-communication','google-workspace','Google Workspace Business Starter','google-workspace-business-starter',
  'Chat is the smallest part of what you get: business email on your own domain, 30 GB per person, Docs, Meet and Drive. Per user per month. A half-price introductory offer for the first three months was running when this was read.',
  700,86,86,94,88,76,72),
 ('team-communication','slack','Slack Pro','slack-pro',
  'The best chat product of the three, and the one whose search and integrations people actually miss when they leave. Per user per month paying monthly; a half-price promotion was running when this was read, at 4.38 USD. The free tier now hides messages after 90 days.',
  875,92,78,88,92,86,70)
) AS f(category_slug, brand_slug, name, slug, description, price_minor,
       quality, value, durability, beginner, advanced, portability)
JOIN catalog.categories categories ON categories.slug = f.category_slug AND categories.vertical_key = 'saas'
JOIN catalog.brands brands ON brands.slug = f.brand_slug
ON CONFLICT (slug) DO NOTHING;

INSERT INTO evidence.sources (external_key, source_type, title, publisher, source_url, is_fictional, review_status, reviewed_at, review_note) VALUES
 ('make-pricing-2026-08','manufacturer_documentation','Make pricing','Make','https://www.make.com/en/pricing',false,'verified',now(),'Read 2026-08-21 through a US connection. Core tier at 10,000 credits a month. The page does not label its billing toggle state, so it was flipped to monthly and back: the default view is annual at 12 USD, monthly is 16 USD.'),
 ('zapier-pricing-2026-08','manufacturer_documentation','Zapier pricing','Zapier','https://zapier.com/pricing',false,'verified',now(),'Read 2026-08-21. Professional tier. The billing control was checked directly in the DOM and read year, so the figure is the annually billed rate. Currency selector read USD.'),
 ('n8n-pricing-2026-08','manufacturer_documentation','n8n pricing','n8n','https://n8n.io/pricing/',false,'verified',now(),'Read 2026-08-21. Starter tier, stated on the page as per month billed annually, for 2,500 workflow executions.'),
 ('sketch-pricing-2026-08','manufacturer_documentation','Sketch pricing','Sketch','https://www.sketch.com/pricing/',false,'verified',now(),'Read 2026-08-21. Standard tier, stated on the page as per editor per month billed yearly.'),
 ('canva-pricing-2026-08','manufacturer_documentation','Canva pricing','Canva','https://www.canva.com/en_us/pricing/',false,'verified',now(),'Read 2026-08-21 through a US connection, after forcing the en_us path because the default redirect priced the page in Bulgarian lev. Canva publishes only 180 US$ per year for one person on the Pro tier and offers no monthly view; the monthly figure recorded is that divided by twelve, and is a derivation rather than a reading.'),
 ('figma-pricing-2026-08','manufacturer_documentation','Figma pricing','Figma','https://www.figma.com/pricing/',false,'verified',now(),'Read 2026-08-21. Professional tier, full seat. The billing toggle was flipped to confirm the state: the default view is annual at 16 USD per seat, monthly is 20 USD.'),
 ('ahrefs-pricing-2026-08','manufacturer_documentation','Ahrefs pricing','Ahrefs','https://ahrefs.com/pricing',false,'verified',now(),'Read 2026-08-21. Starter tier, one included user.'),
 ('seranking-pricing-2026-08','manufacturer_documentation','SE Ranking pricing','SE Ranking','https://seranking.com/subscription.html',false,'verified',now(),'Read 2026-08-21. Core tier. The page shows 103.20 USD against a struck-through 129.00 USD with an annual toggle marked Save 20%, so the figure recorded is the annually billed rate.'),
 ('semrush-pricing-2026-08','manufacturer_documentation','Semrush pricing','Semrush','https://www.semrush.com/pricing/seo-ai-search/',false,'verified',now(),'Read 2026-08-21. The entry tier is named SEO. The page states 117.33 USD per month billed annually instead of 139 USD monthly.'),
 ('msteams-pricing-2026-08','manufacturer_documentation','Microsoft Teams pricing','Microsoft','https://www.microsoft.com/en-us/microsoft-teams/compare-microsoft-teams-business-options',false,'verified',now(),'Read 2026-08-21 on the en-us page. Microsoft Teams Essentials, stated as per user per month paid yearly on an annual auto-renewing subscription.'),
 ('googleworkspace-pricing-2026-08','manufacturer_documentation','Google Workspace pricing','Google','https://workspace.google.com/pricing',false,'verified',now(),'Read 2026-08-21. Business Starter, per user per month. A 50%-off-for-three-months offer was running and displayed 3.50 USD beside the standing 7.00 USD; the standing rate is what is recorded, because a promotion expires and the catalog would silently go stale.'),
 ('slack-pricing-2026-08','manufacturer_documentation','Slack pricing','Slack','https://slack.com/pricing',false,'verified',now(),'Read 2026-08-21. Pro tier, per user per month paying monthly. A half-price promotion was running; the standing rate of 8.75 USD was identified by its strikeprice class in the markup rather than inferred from position, and is what is recorded.')
ON CONFLICT (external_key) DO UPDATE SET
    review_status='verified', reviewed_at=COALESCE(evidence.sources.reviewed_at,now()),
    source_url=EXCLUDED.source_url, review_note=EXCLUDED.review_note, updated_at=now();

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), m.confidence, m.note
FROM (VALUES
 ('make-core','make-pricing-2026-08',100,'Annual rate confirmed by flipping the billing toggle and watching the figure change.'),
 ('zapier-professional','zapier-pricing-2026-08',100,'Annual billing confirmed from the checked state of the billing control.'),
 ('n8n-starter','n8n-pricing-2026-08',100,'Billing period stated explicitly on the vendor page.'),
 ('sketch-standard','sketch-pricing-2026-08',100,'Billing period stated explicitly on the vendor page.'),
 ('canva-pro','canva-pricing-2026-08',85,'Derived: Canva publishes only an annual figure of 180 US$ for one person. The monthly price is that divided by twelve and was not read from the page.'),
 ('figma-professional','figma-pricing-2026-08',100,'Annual rate confirmed by flipping the billing toggle and watching the figure change.'),
 ('ahrefs-starter','ahrefs-pricing-2026-08',100,'Price read from the vendor pricing page.'),
 ('se-ranking-core','seranking-pricing-2026-08',95,'Annual rate; the page shows it against a struck-through monthly figure.'),
 ('semrush-seo','semrush-pricing-2026-08',100,'Both the annual and the monthly rate are stated on the page.'),
 ('microsoft-teams-essentials','msteams-pricing-2026-08',100,'Billing period stated explicitly on the vendor page.'),
 ('google-workspace-business-starter','googleworkspace-pricing-2026-08',100,'Standing rate recorded; a temporary promotion was displayed alongside it.'),
 ('slack-pro','slack-pricing-2026-08',100,'Standing rate identified by its strikeprice markup; a temporary promotion was displayed alongside it.')
) AS m(product_slug, source_key, confidence, note)
JOIN catalog.products p ON p.slug = m.product_slug
JOIN evidence.sources s ON s.external_key = m.source_key
WHERE NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.observations (source_id, product_id, observed_at, confidence, notes)
SELECT s.id, p.id, now(), 80, 'Suitability scores assigned by UNSOLERO editorial review.'
FROM catalog.products p CROSS JOIN evidence.sources s
WHERE s.external_key='unsolero-editorial-saas-2026-08'
  AND p.slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
                 'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
                 'google-workspace-business-starter','slack-pro')
  AND NOT EXISTS (SELECT 1 FROM evidence.observations e WHERE e.source_id=s.id AND e.product_id=p.id);

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,1,p.category_id,p.brand_id,p.name,p.slug,p.description,p.price_minor,p.currency,p.warranty_months,
       'published',now(),now(),now(),'Read from the vendor pricing page on 2026-08-21 through a US connection. Billing basis stated per product. Where a promotion was running, the standing rate is recorded.'
FROM catalog.products p
WHERE p.slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
                 'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
                 'google-workspace-business-starter','slack-pro')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score,
    workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id,f.id,1,p.quality_score,p.value_score,p.durability_score,p.beginner_score,p.advanced_score,
       p.apartment_score,p.noise_score,p.portability_score,'published',now(),now(),now(),
       'Editorial suitability assessment; not a vendor claim.'
FROM catalog.products p JOIN evidence.product_fact_revisions f ON f.product_id=p.id AND f.version=1
WHERE p.slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
                 'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
                 'google-workspace-business-starter','slack-pro')
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT f.id, k.fact_key, o.id, 'manufacturer_claim'
FROM evidence.product_fact_revisions f
JOIN catalog.products p ON p.id=f.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.source_type='manufacturer_documentation'
CROSS JOIN (VALUES ('category'),('brand'),('name'),('description'),('price')) AS k(fact_key)
WHERE f.version=1 AND p.slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
                 'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
                 'google-workspace-business-starter','slack-pro')
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
SELECT sc.id, k.score_key,
       CASE WHEN k.score_key IN ('apartment','noise')
            THEN 'Not applicable: software has no physical footprint and makes no noise.'
            ELSE 'Editorial suitability assessment based on the published feature set, tier limits and vendor track record.' END,
       o.id
FROM evidence.score_revisions sc
JOIN catalog.products p ON p.id=sc.product_id
JOIN evidence.observations o ON o.product_id=p.id
JOIN evidence.sources s ON s.id=o.source_id AND s.external_key='unsolero-editorial-saas-2026-08'
CROSS JOIN (VALUES ('quality'),('value'),('durability'),('beginner'),('advanced'),('apartment'),('noise'),('portability')) AS k(score_key)
WHERE sc.version=1 AND p.slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
                 'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
                 'google-workspace-business-starter','slack-pro')
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS p
SET published_fact_revision_id=f.id, published_score_revision_id=sc.id, status='published'
FROM evidence.product_fact_revisions AS f
JOIN evidence.score_revisions AS sc ON sc.product_id=f.product_id AND sc.version=1
WHERE p.id=f.product_id
  AND p.slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
                 'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
                 'google-workspace-business-starter','slack-pro')
  AND f.version=1 AND f.workflow_status='published' AND sc.workflow_status='published';
