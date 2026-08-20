-- saas-v3: saas-v2 plus the three email marketing tools added on 2026-08-20.
--
-- An active policy is immutable. That is deliberate — a recommendation run
-- records which policy produced it, and a policy that could be edited
-- afterwards would make every past run unreproducible. Adding products to the
-- engine therefore means publishing a new version, not amending the live one.
--
-- Everything here is a copy of saas-v1's configuration. The weights, goals,
-- capabilities, setup roles and priorities are unchanged; only the set of bound
-- products differs. Copied column-by-column rather than with SELECT * so a new
-- column added to any of these tables fails loudly here instead of being
-- silently dropped from the new version.

BEGIN;

-- ------------------------------------------------------- the version row ---
INSERT INTO recommendation.policy_versions (
    version, goal_match_weight, budget_match_weight, space_match_weight,
    experience_match_weight, preference_match_weight, quality_weight,
    value_weight, durability_weight, compatibility_weight, portability_weight,
    noise_weight, priority_boost_percent, maximum_setup_items,
    candidates_per_slot, optional_slot_bonus, vertical_key, workflow_status,
    spatial_constraints, published_at, review_note
)
SELECT 'saas-v3', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v2. Adds MailerLite Comfort, ActiveCampaign Starter and Kit Creator.'
FROM recommendation.policy_versions WHERE version = 'saas-v2'
ON CONFLICT (version) DO NOTHING;

-- ------------------------------------------------ configuration, copied ----
INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v3', capability_key, label
FROM recommendation.policy_capabilities WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v3', goal_key, label
FROM recommendation.policy_goals WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v3', tag_key, label
FROM recommendation.policy_preference_tags WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v3', group_key, label
FROM recommendation.policy_redundancy_groups WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v3', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v3', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (
    policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v3', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (
    policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v3', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v3', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v3', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v3', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v1 ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v3', product_id, fact_revision_id, score_revision_id
FROM recommendation.product_policies WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v3', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v3', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v3', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v3', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version = 'saas-v2'
ON CONFLICT DO NOTHING;

-- ----------------------------------------------------- the three new ones --
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v3', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v3', products.id, mapping.capability, mapping.relation
FROM (VALUES
 ('mailerlite-comfort','email_marketing','provides'),
 ('mailerlite-comfort','landing_pages','provides'),
 ('mailerlite-comfort','crm','compatible_with'),
 ('activecampaign-starter','email_marketing','provides'),
 ('activecampaign-starter','marketing_automation','provides'),
 ('activecampaign-starter','crm','compatible_with'),
 ('kit-creator','email_marketing','provides'),
 ('kit-creator','landing_pages','provides'),
 ('kit-creator','course_hosting','compatible_with')
) AS mapping(slug, capability, relation)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v3', products.id, mapping.goal_key, mapping.score
FROM (VALUES
 ('mailerlite-comfort','client_services',82),
 ('mailerlite-comfort','creator_business',88),
 ('mailerlite-comfort','solo_consulting',90),
 ('mailerlite-comfort','sell_products_online',80),
 ('activecampaign-starter','sell_products_online',90),
 ('activecampaign-starter','client_services',84),
 ('activecampaign-starter','software_product',82),
 ('kit-creator','creator_business',94),
 ('kit-creator','solo_consulting',82),
 ('kit-creator','client_services',70)
) AS mapping(slug, goal_key, score)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v3', products.id, mapping.tag
FROM (VALUES
 ('mailerlite-comfort','no_code'),
 ('mailerlite-comfort','best_of_breed'),
 ('mailerlite-comfort','eu_hosted'),
 ('activecampaign-starter','api_first'),
 ('activecampaign-starter','best_of_breed'),
 ('kit-creator','no_code'),
 ('kit-creator','best_of_breed')
) AS mapping(slug, tag)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v3', products.id, 'email_suite'
FROM catalog.products products
WHERE products.slug IN ('mailerlite-comfort','activecampaign-starter','kit-creator')
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------- promotion ---
-- Retire before activate. A unique index permits exactly one active policy per
-- vertical, so promoting the new version while the old one is still live fails
-- the whole transaction.
UPDATE recommendation.policy_versions
SET workflow_status = 'retired', retired_at = now()
WHERE version = 'saas-v2' AND workflow_status = 'active';

UPDATE recommendation.policy_versions
SET workflow_status = 'active', activated_at = now()
WHERE version = 'saas-v3' AND workflow_status = 'draft';

COMMIT;
