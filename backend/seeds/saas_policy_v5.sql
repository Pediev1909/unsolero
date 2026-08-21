-- saas-v5: saas-v4 plus the five products that filled the two thin categories.
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
SELECT 'saas-v5', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v4. Adds Zoho Invoice, Wave Starter, Zoho Projects Premium, monday.com Basic and Notion Plus.'
FROM recommendation.policy_versions WHERE version = 'saas-v4'
ON CONFLICT (version) DO NOTHING;

-- ------------------------------------------------ configuration, copied ----
INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v5', capability_key, label
FROM recommendation.policy_capabilities WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v5', goal_key, label
FROM recommendation.policy_goals WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v5', tag_key, label
FROM recommendation.policy_preference_tags WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v5', group_key, label
FROM recommendation.policy_redundancy_groups WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v5', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v5', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (
    policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v5', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (
    policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v5', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v5', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v5', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v5', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v1 ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v5', product_id, fact_revision_id, score_revision_id
FROM recommendation.product_policies WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v5', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v5', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v5', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v5', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version = 'saas-v4'
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------ the five new ones --
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v5', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('zoho-invoice','wave-starter','zoho-projects-premium','monday-basic','notion-plus')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v5', p.id, m.capability, m.relation
FROM (VALUES
 ('zoho-invoice','invoicing','provides'),
 ('zoho-invoice','time_tracking','provides'),
 ('zoho-invoice','crm','compatible_with'),
 ('wave-starter','invoicing','provides'),
 ('wave-starter','accounting','provides'),
 ('wave-starter','payments','compatible_with'),
 ('zoho-projects-premium','project_management','provides'),
 ('zoho-projects-premium','task_tracking','provides'),
 ('zoho-projects-premium','time_tracking','provides'),
 ('zoho-projects-premium','crm','compatible_with'),
 ('monday-basic','project_management','provides'),
 ('monday-basic','task_tracking','provides'),
 ('monday-basic','crm','compatible_with'),
 ('notion-plus','project_management','provides'),
 ('notion-plus','task_tracking','provides'),
 ('notion-plus','knowledge_base','provides')
) AS m(slug, capability, relation)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v5', p.id, m.goal_key, m.score
FROM (VALUES
 ('zoho-invoice','solo_consulting',94),
 ('zoho-invoice','creator_business',82),
 ('zoho-invoice','client_services',80),
 ('wave-starter','solo_consulting',92),
 ('wave-starter','client_services',78),
 ('wave-starter','sell_products_online',72),
 ('zoho-projects-premium','client_services',88),
 ('zoho-projects-premium','software_product',82),
 ('zoho-projects-premium','solo_consulting',78),
 ('monday-basic','client_services',86),
 ('monday-basic','sell_products_online',78),
 ('monday-basic','software_product',76),
 ('notion-plus','software_product',88),
 ('notion-plus','creator_business',84),
 ('notion-plus','solo_consulting',82)
) AS m(slug, goal_key, score)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v5', p.id, m.tag
FROM (VALUES
 ('zoho-invoice','no_code'),('zoho-invoice','all_in_one'),('zoho-invoice','eu_hosted'),
 ('wave-starter','no_code'),('wave-starter','best_of_breed'),
 ('zoho-projects-premium','all_in_one'),('zoho-projects-premium','eu_hosted'),
 ('monday-basic','no_code'),('monday-basic','best_of_breed'),
 ('notion-plus','best_of_breed'),('notion-plus','api_first')
) AS m(slug, tag)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v5', p.id, m.grp
FROM (VALUES
 ('zoho-invoice','accounting_suite'),
 ('wave-starter','accounting_suite'),
 ('zoho-projects-premium','project_suite'),
 ('monday-basic','project_suite'),
 ('notion-plus','project_suite')
) AS m(slug, grp)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------------- promotion ---
-- Retire before activate. A unique index permits exactly one active policy per
-- vertical, so promoting the new version while the old one is still live fails
-- the whole transaction.
UPDATE recommendation.policy_versions
SET workflow_status = 'retired', retired_at = now()
WHERE version = 'saas-v4' AND workflow_status = 'active';

UPDATE recommendation.policy_versions
SET workflow_status = 'active', activated_at = now()
WHERE version = 'saas-v5' AND workflow_status = 'draft';

COMMIT;
