-- saas-v8: saas-v7 rebound to whatever revision each product currently
-- publishes.
--
-- Same reason as every version before it: an active policy is immutable, so
-- binding new products means publishing a new version rather than amending the
-- live one. Configuration is copied column-by-column from saas-v7 so that a
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
SELECT 'saas-v8', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration and product set to saas-v7. Rebinds every product to its currently published fact and score revision, which is how a copy edit reaches the engine without amending a live policy.'
FROM recommendation.policy_versions WHERE version = 'saas-v7'
ON CONFLICT (version) DO NOTHING;

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v8', capability_key, label FROM recommendation.policy_capabilities WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v8', goal_key, label FROM recommendation.policy_goals WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v8', tag_key, label FROM recommendation.policy_preference_tags WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v8', group_key, label FROM recommendation.policy_redundancy_groups WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v8', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v8', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v8', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v8', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v8', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v8', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v8', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

-- ------------------------------------------------ products, rebound ----
-- Revisions come from the catalog's current published pointers, not from
-- saas-v7's frozen ones. Copying the old pointers would carry the superseded
-- description into the new version and defeat the exercise.
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v8', p.id, p.published_fact_revision_id, p.published_score_revision_id
FROM recommendation.product_policies pp
JOIN catalog.products p ON p.id = pp.product_id
WHERE pp.policy_version='saas-v7'
  AND p.published_fact_revision_id IS NOT NULL
  AND p.published_score_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v8', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v8', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v8', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v8', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version='saas-v7'
ON CONFLICT DO NOTHING;


-- Retire before activate: one active policy per vertical is enforced by a
-- unique index, so promoting while saas-v7 is still live fails the transaction.
UPDATE recommendation.policy_versions
SET workflow_status='retired', retired_at=now()
WHERE version='saas-v7' AND workflow_status='active';

UPDATE recommendation.policy_versions
SET workflow_status='active', activated_at=now()
WHERE version='saas-v8' AND workflow_status='draft';

COMMIT;
