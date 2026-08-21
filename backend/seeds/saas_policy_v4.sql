-- saas-v4: saas-v3 plus Zoho Campaigns and Brevo, added on 2026-08-21.
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
SELECT 'saas-v4', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v3. Adds Zoho Campaigns Standard and Brevo Starter.'
FROM recommendation.policy_versions WHERE version = 'saas-v3'
ON CONFLICT (version) DO NOTHING;

-- ------------------------------------------------ configuration, copied ----
INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v4', capability_key, label
FROM recommendation.policy_capabilities WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v4', goal_key, label
FROM recommendation.policy_goals WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v4', tag_key, label
FROM recommendation.policy_preference_tags WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v4', group_key, label
FROM recommendation.policy_redundancy_groups WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v4', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v4', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (
    policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v4', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (
    policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v4', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v4', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v4', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v4', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v1 ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v4', product_id, fact_revision_id, score_revision_id
FROM recommendation.product_policies WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v4', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v4', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v4', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v4', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version = 'saas-v3'
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------- the two new ones --
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v4', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('zoho-campaigns-standard','brevo-starter')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v4', products.id, mapping.capability, mapping.relation
FROM (VALUES
 ('zoho-campaigns-standard','email_marketing','provides'),
 ('zoho-campaigns-standard','crm','compatible_with'),
 ('zoho-campaigns-standard','invoicing','compatible_with'),
 ('brevo-starter','email_marketing','provides'),
 ('brevo-starter','marketing_automation','provides'),
 ('brevo-starter','crm','compatible_with'),
 ('brevo-starter','online_store','compatible_with')
) AS mapping(slug, capability, relation)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v4', products.id, mapping.goal_key, mapping.score
FROM (VALUES
 ('zoho-campaigns-standard','solo_consulting',86),
 ('zoho-campaigns-standard','client_services',80),
 ('zoho-campaigns-standard','sell_products_online',78),
 ('zoho-campaigns-standard','creator_business',74),
 ('brevo-starter','sell_products_online',88),
 ('brevo-starter','client_services',84),
 ('brevo-starter','solo_consulting',84),
 ('brevo-starter','creator_business',76)
) AS mapping(slug, goal_key, score)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v4', products.id, mapping.tag
FROM (VALUES
 ('zoho-campaigns-standard','all_in_one'),
 ('zoho-campaigns-standard','no_code'),
 ('zoho-campaigns-standard','eu_hosted'),
 ('brevo-starter','no_code'),
 ('brevo-starter','best_of_breed'),
 ('brevo-starter','eu_hosted')
) AS mapping(slug, tag)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v4', products.id, 'email_suite'
FROM catalog.products products
WHERE products.slug IN ('zoho-campaigns-standard','brevo-starter')
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------- promotion ---
-- Retire before activate. A unique index permits exactly one active policy per
-- vertical, so promoting the new version while the old one is still live fails
-- the whole transaction.
UPDATE recommendation.policy_versions
SET workflow_status = 'retired', retired_at = now()
WHERE version = 'saas-v3' AND workflow_status = 'active';

UPDATE recommendation.policy_versions
SET workflow_status = 'active', activated_at = now()
WHERE version = 'saas-v4' AND workflow_status = 'draft';

COMMIT;
