-- Editorial content for the SaaS vertical.
--
-- Unlike saas_demo.sql, this is NOT fictional. Every claim here is about
-- method — how to reason about a software stack — rather than about any
-- named vendor, price or feature. That is deliberate: vendor facts go stale
-- within months and belong in the catalog behind the evidence workflow, while
-- method does not go stale and is what the pages can honestly assert.
--
-- Entries are inserted as drafts. Review each one, supply the hero images, and
-- publish through the admin interface. Nothing here goes live on its own.
--
-- Hero images are simple SVG diagrams committed alongside the frontend. They
-- illustrate the idea rather than decorate: a list of jobs before any tool is
-- chosen, a stack whose parts connect, and facts entering scoring while
-- commercial data is kept out. Replace them with better artwork when there is
-- any; the entries do not depend on these specific files.

INSERT INTO editorial.authors (name, slug, bio)
VALUES (
    'UNSOLERO Editorial',
    'unsolero-editorial',
    'UNSOLERO editors turn structured software facts into practical, constraint-aware stack guidance. We publish our method so it can be argued with.'
)
ON CONFLICT (slug) DO UPDATE SET
    name = EXCLUDED.name, bio = EXCLUDED.bio, updated_at = now()
WHERE (editorial.authors.name, editorial.authors.bio)
    IS DISTINCT FROM (EXCLUDED.name, EXCLUDED.bio);

WITH author AS (
    SELECT id FROM editorial.authors WHERE slug = 'unsolero-editorial'
)
INSERT INTO editorial.entries (
    author_id, content_type, status, title, slug, description,
    hero_image_url, hero_image_alt, content, seo_title, seo_description
)
SELECT author.id, entry.content_type, 'draft', entry.title, entry.slug,
       entry.description, entry.hero_image_url, entry.hero_image_alt,
       entry.content::jsonb, entry.seo_title, entry.seo_description
FROM author CROSS JOIN (VALUES
(
    'buying_guide',
    'How to choose business software without overbuying',
    'how-to-choose-business-software',
    'A method for assembling a software stack around the jobs your business actually has, instead of collecting tools one landing page at a time.',
    '/images/saas-stack-planning.svg',
    'A desk with a notebook showing a simple list of business tasks before any software is chosen',
    $json$[
      {"type":"paragraph","text":"Most software gets bought one tool at a time. Each purchase looks reasonable on its own, and the result is still a mess: three products that overlap, two that will not talk to each other, and a monthly bill nobody has added up. The problem is not that any single choice was wrong. It is that a stack is a portfolio decision disguised as a series of product decisions."},
      {"type":"heading","heading":"Start from jobs, not from tools"},
      {"type":"paragraph","text":"Write down what the business has to do before you look at anything you could buy. A client services business has to find work, deliver it, and get paid. An online store has to show products, take money, and answer customers. These jobs are stable. The tools that serve them are not."},
      {"type":"paragraph","text":"Once the list exists, every tool has to earn its place by covering a job. A tool that covers no job on the list is not a bargain at any price."},
      {"type":"heading","heading":"Count the integration, not just the licence"},
      {"type":"paragraph","text":"The advertised price is rarely the real cost. Two tools that do not connect mean somebody re-types data every week, and that person costs more per month than either subscription. When comparing two options that both cover the job, the one that already connects to what you own is usually cheaper in practice even when it is more expensive on paper."},
      {"type":"callout","heading":"The question worth asking","text":"Not \"what does this cost?\" but \"what will it cost me to keep this in sync with everything else?\""},
      {"type":"heading","heading":"Watch for overlap you are paying for twice"},
      {"type":"paragraph","text":"Suites expand. A tool bought for invoicing grows a CRM; a CRM grows email marketing. Over a couple of years it is normal to end up paying two vendors for the same capability while using neither fully. Before adding anything, check what your existing tools already claim to do."},
      {"type":"paragraph","text":"Overlap is not always waste. Sometimes a specialist tool is genuinely better than the bundled version. But it should be a decision you made, not one you drifted into."},
      {"type":"heading","heading":"Price the exit before the entrance"},
      {"type":"paragraph","text":"The cost of leaving a tool is part of the cost of choosing it. Before committing, find out whether you can export your data in a usable format, and what happens to it if you stop paying. A product that makes leaving hard is charging you a fee you cannot see on the invoice."},
      {"type":"unordered_list","items":[
        "Can you export the data yourself, without asking support?",
        "Is the export a real format, or a screenshot-grade PDF?",
        "What is retained, and for how long, after you cancel?",
        "If the vendor were acquired tomorrow, what would you lose?"
      ]},
      {"type":"heading","heading":"Buy in the order the business needs"},
      {"type":"paragraph","text":"Not everything has to be solved in month one. Cover the jobs that stop work or stop payment first; everything else can wait until there is enough volume to justify it. Deferring a purchase is a legitimate outcome, and usually a better one than buying early and switching later."},
      {"type":"callout","heading":"How we use this","text":"UNSOLERO applies exactly this method: it works out which jobs your goal implies, filters to tools that fit your budget and existing stack, removes ones that duplicate what you already own, and shows the reasoning for each choice."}
    ]$json$,
    'How to choose business software without overbuying',
    'A practical method for building a software stack around your actual jobs, integrations and exit costs rather than one landing page at a time.'
),
(
    'guide',
    'Building a software stack for a small client services agency',
    'software-stack-small-agency',
    'What a two to ten person agency actually needs to run client work, what can wait, and how the pieces should fit together.',
    '/images/saas-agency-stack.svg',
    'A small team workspace with two people reviewing a project board together',
    $json$[
      {"type":"paragraph","text":"A small agency runs the same loop every month: win work, do the work, invoice for the work, and keep everyone informed while it happens. A stack is worth its cost when it makes that loop shorter. This guide walks the loop and says what each stage needs, and what it does not."},
      {"type":"heading","heading":"Client management comes first"},
      {"type":"paragraph","text":"Before delivery tooling, before anything else, you need one place that knows who your clients are and what was agreed. Until that exists it lives in one person's inbox, which is a single point of failure with a holiday schedule."},
      {"type":"paragraph","text":"Early on this can be modest. What matters is that it is shared, searchable, and connected to how you invoice, so the same client record does not get typed twice."},
      {"type":"heading","heading":"Delivery: fewer features than you think"},
      {"type":"paragraph","text":"Project tools sell on features that matter at fifty people and get in the way at five. For a small team the requirements are narrow: who is doing what, what is late, and what is waiting on the client. Sophistication beyond that tends to become a second job."},
      {"type":"callout","heading":"A useful test","text":"If a tool needs someone to maintain it, and nobody is paid to maintain it, it will be abandoned within two quarters. Choose accordingly."},
      {"type":"heading","heading":"Billing is not optional and not the place to save"},
      {"type":"paragraph","text":"Invoicing is where a small agency loses money quietly: late invoices, forgotten expenses, work delivered against a scope nobody recorded. This is the one stage where connecting to the client record pays for itself immediately, because the invoice is generated from what was agreed rather than remembered."},
      {"type":"heading","heading":"Communication and scheduling can wait"},
      {"type":"paragraph","text":"Team chat and booking tools are genuinely useful, but neither stops work or stops payment. Both are reasonable to defer until the first three stages are steady. Adding them early is how a stack ends up costing more than it saves."},
      {"type":"heading","heading":"A staged order that usually works"},
      {"type":"ordered_list","items":[
        "Client records, shared and searchable.",
        "Invoicing, connected to those records.",
        "Delivery tracking, kept deliberately simple.",
        "Team communication, once coordination cost is real.",
        "Scheduling, once client booking is a bottleneck."
      ]},
      {"type":"paragraph","text":"The order matters more than the specific products. A stack assembled in this sequence stays coherent, because each addition has something to connect to."}
    ]$json$,
    'Software stack for a small client services agency',
    'What a two to ten person agency needs to run client work, in what order to buy it, and which tools are safe to defer.'
),
(
    'article',
    'How UNSOLERO ranks software, and why commission never moves it',
    'how-unsolero-ranks-software',
    'Our recommendations are produced by a deterministic engine that cannot read commercial data. This page explains the mechanism and how to check it.',
    '/images/saas-methodology.svg',
    'A diagram showing product facts entering a scoring process separately from commercial data',
    $json$[
      {"type":"paragraph","text":"Every site that recommends software earns money when you click through. That is not a scandal by itself. The question is whether the money changes the advice. On UNSOLERO it cannot, and this page explains the mechanism rather than asking you to take it on faith."},
      {"type":"heading","heading":"Recommendations are computed, not written"},
      {"type":"paragraph","text":"When you describe your goal, budget and existing tools, a deterministic engine scores every eligible product against that specific situation. The same inputs always produce the same output. There is no editorial thumb on the scale, because there is no editorial step between your inputs and the ranking."},
      {"type":"heading","heading":"The engine cannot see commercial data"},
      {"type":"paragraph","text":"Commission rates, merchant relationships and click revenue live in a separate part of the system from the recommendation engine. The engine's inputs are product facts and your constraints; commercial fields are not among them, and an automated test fails the build if anyone adds one."},
      {"type":"callout","heading":"What this means in practice","text":"A tool that pays us nothing can and does outrank a tool that pays us well, because the ranking is computed before commercial data is ever consulted."},
      {"type":"heading","heading":"Facts need a source before they are published"},
      {"type":"paragraph","text":"A product cannot appear in recommendations until its facts carry recorded provenance. This is enforced by the system, not by good intentions: an entry without attested facts stays unpublished. It is why our catalog grows slowly, and why we would rather cover fewer products accurately than more products approximately."},
      {"type":"heading","heading":"Where money does enter"},
      {"type":"paragraph","text":"After a recommendation exists, we may earn a commission if you choose to buy through one of the links. The link is disclosed. What it does not do is change which products were recommended, in what order, or why."},
      {"type":"heading","heading":"How to check us"},
      {"type":"unordered_list","items":[
        "Every recommendation shows the reasons behind it. Read them; they cite the facts used.",
        "Rejected products are shown with the reason they were rejected, not hidden.",
        "Cheaper and more expensive alternatives are always offered alongside the pick.",
        "If a recommendation does not make sense to you, it is a bug or a bad fact. Tell us and we will show our working."
      ]},
      {"type":"paragraph","text":"Trust in this category is mostly asserted and rarely demonstrated. We would rather be checkable than trusted."}
    ]$json$,
    'How UNSOLERO ranks software and why commission never moves it',
    'Our rankings are computed by a deterministic engine that has no access to commission data. Here is the mechanism and how to verify it yourself.'
)
) AS entry(content_type, title, slug, description, hero_image_url, hero_image_alt, content, seo_title, seo_description)
ON CONFLICT (slug) DO NOTHING;

-- Link the agency guide to the categories it discusses so the entry appears on
-- those category pages and the internal linking is real rather than decorative.
-- position is explicit and unique per entry, so the ordering is deterministic
-- rather than dependent on however the join happens to return rows.
INSERT INTO editorial.entry_categories (entry_id, category_id, position)
SELECT entries.id, categories.id, mapping.position
FROM (VALUES
 ('software-stack-small-agency', 'crm', 0),
 ('software-stack-small-agency', 'accounting-invoicing', 1),
 ('software-stack-small-agency', 'project-management', 2),
 ('how-to-choose-business-software', 'crm', 0),
 ('how-to-choose-business-software', 'email-marketing', 1),
 ('how-to-choose-business-software', 'automation', 2)
) AS mapping(entry_slug, category_slug, position)
JOIN editorial.entries entries ON entries.slug = mapping.entry_slug
JOIN catalog.categories categories
  ON categories.slug = mapping.category_slug AND categories.vertical_key = 'saas'
ON CONFLICT DO NOTHING;

INSERT INTO editorial.related_entries (entry_id, related_entry_id, position)
SELECT source.id, target.id, mapping.position
FROM (VALUES
 ('how-to-choose-business-software', 'software-stack-small-agency', 0),
 ('how-to-choose-business-software', 'how-unsolero-ranks-software', 1),
 ('software-stack-small-agency', 'how-to-choose-business-software', 0),
 ('software-stack-small-agency', 'how-unsolero-ranks-software', 1),
 ('how-unsolero-ranks-software', 'how-to-choose-business-software', 0)
) AS mapping(source_slug, target_slug, position)
JOIN editorial.entries source ON source.slug = mapping.source_slug
JOIN editorial.entries target ON target.slug = mapping.target_slug
ON CONFLICT DO NOTHING;
