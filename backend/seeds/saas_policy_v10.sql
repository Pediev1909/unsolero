-- saas-v10: saas-v9 plus a sixth goal, for a business with a front door.
--
-- The five goals all described a business that sells remotely: client work,
-- online products, an audience, software, consulting. A hairdresser, a garage
-- or a dentist has a different shape -- be found locally, take a booking, take
-- the money -- and every tool that shape needs was already in the catalog with
-- nowhere to point them.
--
-- Same reason as every version before it for the new version: an active policy
-- is immutable, so changing the goal set means publishing a new version rather
-- than amending the live one. Configuration is copied column-by-column from saas-v9 so that a
-- column added to any of these tables fails loudly here instead of being
-- silently dropped from the new version.

BEGIN;

INSERT INTO recommendation.policy_versions (
    version, goal_match_weight, budget_match_weight, space_match_weight,
    experience_match_weight, preference_match_weight, quality_weight,
    value_weight, durability_weight, compatibility_weight, portability_weight,
    noise_weight, priority_boost_percent, maximum_setup_items,
    candidates_per_slot, optional_slot_bonus, vertical_key, workflow_status,
    spatial_constraints, published_at, review_note
)
SELECT 'saas-v10', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Adds the local_business goal. A salon, a restaurant, a clinic or a trade needs to be found, booked and paid -- a shape none of the five existing goals described, while the catalog could already serve it.'
FROM recommendation.policy_versions WHERE version = 'saas-v9'
ON CONFLICT (version) DO NOTHING;

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v10', capability_key, label FROM recommendation.policy_capabilities WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v10', goal_key, label FROM recommendation.policy_goals WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v10', tag_key, label FROM recommendation.policy_preference_tags WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v10', group_key, label FROM recommendation.policy_redundancy_groups WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v10', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v10', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v10', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v10', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v10', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v10', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v10', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v9 ----
-- Revisions come from the catalog's current published pointers, not from
-- saas-v9's frozen ones. Copying the old pointers would carry a superseded
-- description into the new version and defeat the exercise.
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v10', p.id, p.published_fact_revision_id, p.published_score_revision_id
FROM recommendation.product_policies pp
JOIN catalog.products p ON p.id = pp.product_id
WHERE pp.policy_version='saas-v9'
  AND p.published_fact_revision_id IS NOT NULL
  AND p.published_score_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v10', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v10', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v10', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v10', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version='saas-v9'
ON CONFLICT DO NOTHING;


-- ------------------------------------------------- the local business goal --
INSERT INTO recommendation.policy_goals (policy_version, goal_key, label) VALUES
 ('saas-v10','local_business','running a business people visit')
ON CONFLICT DO NOTHING;

-- Three required roles, and they are the three things a local business cannot
-- trade without: be findable, be bookable, be payable. Everything else is a
-- refinement, and a salon that cannot take a booking has a bigger problem than
-- its email marketing.
INSERT INTO recommendation.policy_setup_roles (policy_version, goal_key, role_key, label, is_required, sort_order) VALUES
 ('saas-v10','local_business','presence','Somewhere to be found',true,0),
 ('saas-v10','local_business','booking','Taking bookings',true,1),
 ('saas-v10','local_business','checkout','Taking payment',true,2),
 ('saas-v10','local_business','billing','Invoices and books',false,3),
 ('saas-v10','local_business','regulars','Keeping regulars',false,4),
 ('saas-v10','local_business','questions','Answering questions',false,5)
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (policy_version, goal_key, role_key, capability_key) VALUES
 ('saas-v10','local_business','presence','website_builder'),
 ('saas-v10','local_business','booking','scheduling'),
 ('saas-v10','local_business','checkout','payments'),
 ('saas-v10','local_business','billing','invoicing'),
 ('saas-v10','local_business','billing','accounting'),
 ('saas-v10','local_business','regulars','email_marketing'),
 ('saas-v10','local_business','questions','help_desk')
ON CONFLICT DO NOTHING;

-- Goal support. A product with no score for this goal can never be chosen for
-- it, so anything that genuinely serves a business with a front door needs a
-- number here. The numbers are judgements and are scored as such: a booking
-- tool that already knows the customer record beats one that does not, and a
-- website builder that a non-technical owner can actually maintain beats one
-- that needs a developer every time the opening hours change.
INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v10', p.id, 'local_business', m.score
FROM (VALUES
 -- be found
 ('squarespace-basic',94),   -- least asked of a non-technical owner
 ('framer-basic',74),
 ('webflow-basic',62),       -- needs someone technical for every change
 -- be booked
 ('calendly-standard',92),
 ('zoho-bookings-basic',90), -- cheaper, and knows the customer already
 ('cal-com-teams',70),
 -- be paid
 ('stripe',88),
 ('paddle',56),              -- merchant of record earns its fee on cross-border digital sales, not haircuts
 ('lemon-squeezy',52),
 -- invoices and books
 ('zoho-books-standard',88),
 ('wave-starter',86),
 ('zoho-invoice',84),
 ('freshbooks-lite',80),
 -- keeping regulars
 ('mailerlite-comfort',86),
 ('brevo-starter',84),
 ('zoho-campaigns-standard',80),
 -- answering questions
 ('tidio-starter',90),       -- live chat is what a visitor site actually needs
 ('freshdesk-growth',72),
 ('help-scout-standard',68),
 -- the rest, plausible but secondary
 ('bigin-express',78),
 ('zoho-crm-standard',66),
 ('canva-pro',82),
 ('ecwid-starter',74),
 ('shopify-basic',68),
 ('google-workspace-business-starter',80),
 ('microsoft-teams-essentials',66),
 ('fathom-analytics',62),
 ('zapier-professional',58)
) AS m(slug, score)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

-- Retire before activate: one active policy per vertical is enforced by a
-- unique index, so promoting while saas-v9 is still live fails the transaction.
UPDATE recommendation.policy_versions
SET workflow_status='retired', retired_at=now()
WHERE version='saas-v9' AND workflow_status='active';

UPDATE recommendation.policy_versions
SET workflow_status='active', activated_at=now()
WHERE version='saas-v10' AND workflow_status='draft';

COMMIT;
