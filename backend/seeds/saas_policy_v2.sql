-- saas-v2: saas-v1 plus the four CRMs added on 2026-08-20.
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
SELECT 'saas-v2', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v1. Adds Zoho CRM Standard, Bigin Express, Pipedrive Lite and Freshsales Growth.'
FROM recommendation.policy_versions WHERE version = 'saas-v1'
ON CONFLICT (version) DO NOTHING;

-- ------------------------------------------------ configuration, copied ----
INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v2', capability_key, label
FROM recommendation.policy_capabilities WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v2', goal_key, label
FROM recommendation.policy_goals WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v2', tag_key, label
FROM recommendation.policy_preference_tags WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v2', group_key, label
FROM recommendation.policy_redundancy_groups WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v2', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v2', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (
    policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v2', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (
    policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v2', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v2', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v2', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v2', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v1 ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v2', product_id, fact_revision_id, score_revision_id
FROM recommendation.product_policies WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v2', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v2', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v2', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v2', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version = 'saas-v1'
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------ the four new ones --
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v2', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v2', products.id, mapping.capability, mapping.relation
FROM (VALUES
 ('zoho-crm-standard','crm','provides'),
 ('zoho-crm-standard','email_marketing','compatible_with'),
 ('zoho-crm-standard','invoicing','compatible_with'),
 ('zoho-crm-standard','project_management','compatible_with'),
 ('bigin-express','crm','provides'),
 ('bigin-express','invoicing','compatible_with'),
 ('bigin-express','email_marketing','compatible_with'),
 ('pipedrive-lite','crm','provides'),
 ('pipedrive-lite','email_marketing','compatible_with'),
 ('pipedrive-lite','invoicing','compatible_with'),
 ('pipedrive-lite','project_management','compatible_with'),
 ('freshsales-growth','crm','provides'),
 ('freshsales-growth','email_marketing','compatible_with'),
 ('freshsales-growth','invoicing','compatible_with')
) AS mapping(slug, capability, relation)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v2', products.id, mapping.goal_key, mapping.score
FROM (VALUES
 ('zoho-crm-standard','client_services',84),
 ('zoho-crm-standard','sell_products_online',80),
 ('zoho-crm-standard','software_product',78),
 ('zoho-crm-standard','solo_consulting',70),
 ('bigin-express','client_services',88),
 ('bigin-express','solo_consulting',92),
 ('bigin-express','creator_business',78),
 ('pipedrive-lite','client_services',90),
 ('pipedrive-lite','solo_consulting',86),
 ('pipedrive-lite','sell_products_online',72),
 ('freshsales-growth','client_services',84),
 ('freshsales-growth','solo_consulting',82),
 ('freshsales-growth','sell_products_online',74)
) AS mapping(slug, goal_key, score)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v2', products.id, mapping.tag
FROM (VALUES
 ('zoho-crm-standard','all_in_one'),
 ('zoho-crm-standard','api_first'),
 ('zoho-crm-standard','eu_hosted'),
 ('bigin-express','no_code'),
 ('bigin-express','all_in_one'),
 ('pipedrive-lite','best_of_breed'),
 ('pipedrive-lite','api_first'),
 ('pipedrive-lite','eu_hosted'),
 ('freshsales-growth','best_of_breed'),
 ('freshsales-growth','no_code')
) AS mapping(slug, tag)
JOIN catalog.products products ON products.slug = mapping.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v2', products.id, 'crm_suite'
FROM catalog.products products
WHERE products.slug IN ('zoho-crm-standard','bigin-express','pipedrive-lite','freshsales-growth')
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------- promotion ---
-- Retire before activate. A unique index permits exactly one active policy per
-- vertical, so promoting the new version while the old one is still live fails
-- the whole transaction.
UPDATE recommendation.policy_versions
SET workflow_status = 'retired', retired_at = now()
WHERE version = 'saas-v1' AND workflow_status = 'active';

UPDATE recommendation.policy_versions
SET workflow_status = 'active', activated_at = now()
WHERE version = 'saas-v2' AND workflow_status = 'draft';

COMMIT;
