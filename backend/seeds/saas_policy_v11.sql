-- saas-v11: saas-v10 with current published evidence pointers.
--
-- Zoho Books Standard received a verified $20 monthly-billing fact revision.
-- Active policies are immutable, so saas-v10 remains a reproducible record of
-- earlier recommendations and this version binds the newly published revision.

BEGIN;

INSERT INTO recommendation.policy_versions (
    version, goal_match_weight, budget_match_weight, space_match_weight,
    experience_match_weight, preference_match_weight, quality_weight,
    value_weight, durability_weight, compatibility_weight, portability_weight,
    noise_weight, priority_boost_percent, maximum_setup_items,
    candidates_per_slot, optional_slot_bonus, vertical_key, workflow_status,
    spatial_constraints, published_at, review_note
)
SELECT 'saas-v11', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Configuration is identical to saas-v10. Product bindings use current published evidence, including Zoho Books Standard fact and score revision 2 at $20 monthly billing.'
FROM recommendation.policy_versions
WHERE version = 'saas-v10'
  AND NOT EXISTS (
      SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11'
  )
ON CONFLICT (version) DO NOTHING;

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v11', capability_key, label
FROM recommendation.policy_capabilities WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v11', goal_key, label
FROM recommendation.policy_goals WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v11', tag_key, label
FROM recommendation.policy_preference_tags WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v11', group_key, label
FROM recommendation.policy_redundancy_groups WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order
)
SELECT 'saas-v11', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (
    policy_version, priority_key, dimension
)
SELECT 'saas-v11', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (
    policy_version, goal_key, role_key, label, is_required, sort_order
)
SELECT 'saas-v11', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (
    policy_version, goal_key, role_key, capability_key
)
SELECT 'saas-v11', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width
)
SELECT 'saas-v11', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance,
       requires_access_width
FROM recommendation.category_policies WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (
    policy_version, category_id, capability_key
)
SELECT 'saas-v11', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (
    policy_version, category_id, group_key
)
SELECT 'saas-v11', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policies (
    policy_version, product_id, fact_revision_id, score_revision_id
)
SELECT 'saas-v11', products.id, products.published_fact_revision_id,
       products.published_score_revision_id
FROM recommendation.product_policies previous
JOIN catalog.products products ON products.id = previous.product_id
WHERE previous.policy_version = 'saas-v10'
  AND products.published_fact_revision_id IS NOT NULL
  AND products.published_score_revision_id IS NOT NULL
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (
    policy_version, product_id, capability_key, relation_type
)
SELECT 'saas-v11', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities
WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (
    policy_version, product_id, goal_key, match_score
)
SELECT 'saas-v11', product_id, goal_key, match_score
FROM recommendation.product_goal_support
WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (
    policy_version, product_id, preference_key
)
SELECT 'saas-v11', product_id, preference_key
FROM recommendation.product_preference_tags
WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (
    policy_version, product_id, group_key
)
SELECT 'saas-v11', product_id, group_key
FROM recommendation.product_redundancy_groups
WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_space_profiles (
    policy_version, product_id,
    footprint_length_mm, footprint_width_mm, footprint_height_mm,
    storage_length_mm, storage_width_mm, storage_height_mm,
    operating_front_mm, operating_back_mm, operating_left_mm,
    operating_right_mm, operating_top_mm,
    safety_front_mm, safety_back_mm, safety_left_mm, safety_right_mm,
    safety_top_mm, minimum_room_height_mm, minimum_access_width_mm,
    overlap_group
)
SELECT 'saas-v11', product_id,
       footprint_length_mm, footprint_width_mm, footprint_height_mm,
       storage_length_mm, storage_width_mm, storage_height_mm,
       operating_front_mm, operating_back_mm, operating_left_mm,
       operating_right_mm, operating_top_mm,
       safety_front_mm, safety_back_mm, safety_left_mm, safety_right_mm,
       safety_top_mm, minimum_room_height_mm, minimum_access_width_mm,
       overlap_group
FROM recommendation.product_space_profiles
WHERE policy_version = 'saas-v10'
  AND EXISTS (SELECT 1 FROM recommendation.policy_versions WHERE version = 'saas-v11' AND workflow_status = 'draft')
ON CONFLICT DO NOTHING;

DO $$
DECLARE
    previous_products integer;
    new_products integer;
    zoho_binding integer;
BEGIN
    SELECT count(*) INTO previous_products
    FROM recommendation.product_policies
    WHERE policy_version = 'saas-v10';

    SELECT count(*) INTO new_products
    FROM recommendation.product_policies
    WHERE policy_version = 'saas-v11';

    SELECT count(*) INTO zoho_binding
    FROM recommendation.product_policies policies
    JOIN catalog.products products ON products.id = policies.product_id
    JOIN evidence.product_fact_revisions facts
      ON facts.id = policies.fact_revision_id
    JOIN evidence.score_revisions scores
      ON scores.id = policies.score_revision_id
    WHERE policies.policy_version = 'saas-v11'
      AND products.slug = 'zoho-books-standard'
      AND facts.version = 2
      AND facts.price_minor = 2000
      AND scores.version = 2;

    IF previous_products = 0 OR new_products <> previous_products THEN
        RAISE EXCEPTION
            'saas-v11 product copy failed: saas-v10 has %, saas-v11 has %',
            previous_products, new_products;
    END IF;

    IF zoho_binding <> 1 THEN
        RAISE EXCEPTION
            'saas-v11 activation blocked: expected one Zoho Books revision-2 binding, found %',
            zoho_binding;
    END IF;
END
$$;

UPDATE recommendation.policy_versions
SET workflow_status = 'retired', retired_at = now()
WHERE version = 'saas-v10' AND workflow_status = 'active';

UPDATE recommendation.policy_versions
SET workflow_status = 'active', activated_at = now()
WHERE version = 'saas-v11' AND workflow_status = 'draft';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM recommendation.policy_versions
        WHERE version = 'saas-v11' AND workflow_status = 'active'
    ) THEN
        RAISE EXCEPTION 'saas-v11 activation failed';
    END IF;
END
$$;

COMMIT;
