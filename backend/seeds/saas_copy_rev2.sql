-- Second fact revision for the four products that charge nothing per month.
--
-- The price slot on a product page now reads "No monthly fee", and all four
-- descriptions opened with the same words directly underneath it. The fee also
-- appeared as "30c" and "50c" because the seed avoided a non-ASCII character it
-- had no reason to avoid.
--
-- The description is a fact with provenance attached, so this is a new revision
-- rather than an UPDATE on the published one. Version 1 stays exactly as it was
-- published, which is the point of keeping revisions at all.

BEGIN;

CREATE TEMP TABLE copy_rev2(slug text PRIMARY KEY, description text) ON COMMIT DROP;
INSERT INTO copy_rev2 VALUES
 ('stripe',
  '2.9% + 30¢ per successful domestic card charge, with more for international cards and currency conversion. The deepest API here by a wide margin, and the one that expects a developer. You remain the seller of record, so the sales tax and VAT are yours to handle.'),
 ('paddle',
  '5% + 50¢ per checkout transaction. Paddle is the merchant of record, meaning it sells to your customer and takes on the sales tax and VAT filing worldwide. That is what the extra two points over a plain processor buys, and for a one-person business it is usually worth it.'),
 ('lemon-squeezy',
  '5% + 50¢ per transaction, as merchant of record like Paddle, with the friendliest setup of the three. It is now part of Stripe, which is worth weighing: the product is well supported today, and its independent roadmap no longer exists.'),
 ('gumroad',
  '10% + 50¢ of every sale through your own links, rising to 30% when a buyer finds you through Gumroad Discover. Free until you sell and expensive once you do: at 1,000 USD of monthly sales it costs more than any subscription in this category.');

INSERT INTO evidence.product_fact_revisions (
    product_id, version, category_id, brand_id, name, slug, description,
    price_minor, currency, warranty_months, workflow_status,
    submitted_at, reviewed_at, published_at, review_note)
SELECT p.id, 2, p.category_id, p.brand_id, p.name, p.slug, c.description,
       p.price_minor, p.currency, p.warranty_months, 'published', now(), now(), now(),
       'Copy revision only. The price and its source are unchanged from version 1; the description no longer repeats the price label and spells the cent sign properly.'
FROM catalog.products p JOIN copy_rev2 c ON c.slug = p.slug
ON CONFLICT (product_id, version) DO NOTHING;

INSERT INTO evidence.score_revisions (
    product_id, fact_revision_id, version, quality_score, value_score, durability_score,
    beginner_score, advanced_score, apartment_score, noise_score, portability_score,
    workflow_status, submitted_at, reviewed_at, published_at, review_note)
SELECT p.id, f.id, 2, p.quality_score, p.value_score, p.durability_score,
       p.beginner_score, p.advanced_score, p.apartment_score, p.noise_score, p.portability_score,
       'published', now(), now(), now(),
       'Unchanged from version 1; a score revision is paired with each fact revision.'
FROM catalog.products p
JOIN copy_rev2 c ON c.slug = p.slug
JOIN evidence.product_fact_revisions f ON f.product_id = p.id AND f.version = 2
ON CONFLICT (product_id, version) DO NOTHING;

-- Provenance and rationale carry over: the same observations back the new
-- revision, because nothing that was observed has changed.
INSERT INTO evidence.fact_provenance (fact_revision_id, fact_key, observation_id, public_classification)
SELECT f2.id, fp.fact_key, fp.observation_id, fp.public_classification
FROM evidence.product_fact_revisions f2
JOIN catalog.products p ON p.id = f2.product_id
JOIN copy_rev2 c ON c.slug = p.slug
JOIN evidence.product_fact_revisions f1 ON f1.product_id = p.id AND f1.version = 1
JOIN evidence.fact_provenance fp ON fp.fact_revision_id = f1.id
WHERE f2.version = 2
ON CONFLICT DO NOTHING;

INSERT INTO evidence.score_rationales (score_revision_id, score_key, rationale, observation_id)
SELECT s2.id, sr.score_key, sr.rationale, sr.observation_id
FROM evidence.score_revisions s2
JOIN catalog.products p ON p.id = s2.product_id
JOIN copy_rev2 c ON c.slug = p.slug
JOIN evidence.score_revisions s1 ON s1.product_id = p.id AND s1.version = 1
JOIN evidence.score_rationales sr ON sr.score_revision_id = s1.id
WHERE s2.version = 2
ON CONFLICT DO NOTHING;

UPDATE catalog.products AS p
SET description = c.description,
    published_fact_revision_id = f.id,
    published_score_revision_id = s.id
FROM copy_rev2 c
JOIN evidence.product_fact_revisions f ON f.version = 2
JOIN evidence.score_revisions s ON s.product_id = f.product_id AND s.version = 2
WHERE p.slug = c.slug AND f.product_id = p.id;

-- The active policy binds a specific revision pair and cannot be edited, which
-- is the whole point: a past recommendation names the policy that produced it,
-- and a policy that could be amended afterwards would make that run
-- unreproducible. Picking up these revisions therefore needs saas-v8, in
-- saas_policy_v8.sql. Until that runs, the live engine keeps serving version 1
-- of the copy while the site shows version 2 -- correct in both places, since
-- only the wording differs.

COMMIT;
