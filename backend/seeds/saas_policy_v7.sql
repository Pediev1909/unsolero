-- saas-v7: saas-v6 plus the twelve products that open website builders, online
-- stores, payments and course platforms on 2026-08-21.
--
-- Same reason as every version before it: an active policy is immutable, so
-- binding new products means publishing a new version rather than amending the
-- live one. Configuration is copied column-by-column from saas-v6 so that a
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
SELECT 'saas-v7', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v6. Opens website-builder, ecommerce-platform, payments and course-platform, which makes the sell_products_online and creator_business goals answerable for the first time.'
FROM recommendation.policy_versions WHERE version = 'saas-v6'
ON CONFLICT (version) DO NOTHING;

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v7', capability_key, label FROM recommendation.policy_capabilities WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v7', goal_key, label FROM recommendation.policy_goals WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v7', tag_key, label FROM recommendation.policy_preference_tags WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v7', group_key, label FROM recommendation.policy_redundancy_groups WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v7', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v7', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v7', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v7', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v7', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v7', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v7', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v6 ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v7', product_id, fact_revision_id, score_revision_id
FROM recommendation.product_policies WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v7', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v7', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v7', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v7', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version='saas-v6'
ON CONFLICT DO NOTHING;

-- ------------------------------------------------------ the twelve new ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v7', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('framer-basic','webflow-basic','squarespace-basic','ecwid-starter','shopify-basic',
               'bigcommerce-core','stripe','paddle','lemon-squeezy','gumroad','teachable-starter','thinkific-basic')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- A processor "provides" payments; a store is only "compatible_with" it, even
-- though every store here can take a card. The distinction is what lets the
-- optimizer fill the checkout role with a processor rather than counting the
-- storefront twice and calling the setup complete.
INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v7', p.id, m.capability, m.relation
FROM (VALUES
 ('framer-basic','website_builder','provides'),
 ('framer-basic','product_analytics','compatible_with'),
 ('framer-basic','form_builder','compatible_with'),
 ('webflow-basic','website_builder','provides'),
 ('webflow-basic','online_store','compatible_with'),
 ('webflow-basic','product_analytics','compatible_with'),
 ('webflow-basic','form_builder','compatible_with'),
 ('squarespace-basic','website_builder','provides'),
 ('squarespace-basic','online_store','compatible_with'),
 ('squarespace-basic','payments','compatible_with'),
 ('squarespace-basic','scheduling','compatible_with'),
 ('ecwid-starter','online_store','provides'),
 ('ecwid-starter','payments','compatible_with'),
 ('ecwid-starter','website_builder','compatible_with'),
 ('shopify-basic','online_store','provides'),
 ('shopify-basic','payments','compatible_with'),
 ('shopify-basic','product_analytics','compatible_with'),
 ('shopify-basic','help_desk','compatible_with'),
 ('bigcommerce-core','online_store','provides'),
 ('bigcommerce-core','payments','compatible_with'),
 ('bigcommerce-core','product_analytics','compatible_with'),
 ('stripe','payments','provides'),
 ('stripe','online_store','compatible_with'),
 ('stripe','invoicing','compatible_with'),
 ('stripe','accounting','compatible_with'),
 ('paddle','payments','provides'),
 ('paddle','online_store','compatible_with'),
 ('paddle','invoicing','compatible_with'),
 ('lemon-squeezy','payments','provides'),
 ('lemon-squeezy','online_store','compatible_with'),
 ('lemon-squeezy','course_hosting','compatible_with'),
 ('gumroad','course_hosting','provides'),
 ('gumroad','payments','compatible_with'),
 ('gumroad','email_marketing','compatible_with'),
 ('teachable-starter','course_hosting','provides'),
 ('teachable-starter','payments','compatible_with'),
 ('teachable-starter','email_marketing','compatible_with'),
 ('thinkific-basic','course_hosting','provides'),
 ('thinkific-basic','payments','compatible_with'),
 ('thinkific-basic','email_marketing','compatible_with')
) AS m(slug, capability, relation)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v7', p.id, m.goal_key, m.score
FROM (VALUES
 ('framer-basic','creator_business',90),
 ('framer-basic','solo_consulting',84),
 ('framer-basic','software_product',70),
 ('webflow-basic','software_product',88),
 ('webflow-basic','client_services',86),
 ('webflow-basic','creator_business',80),
 ('squarespace-basic','solo_consulting',92),
 ('squarespace-basic','client_services',88),
 ('squarespace-basic','creator_business',82),
 ('ecwid-starter','sell_products_online',86),
 ('ecwid-starter','creator_business',74),
 ('ecwid-starter','solo_consulting',68),
 ('shopify-basic','sell_products_online',96),
 ('shopify-basic','creator_business',78),
 ('bigcommerce-core','sell_products_online',88),
 ('bigcommerce-core','software_product',70),
 ('stripe','software_product',96),
 ('stripe','sell_products_online',88),
 ('stripe','creator_business',72),
 ('paddle','software_product',90),
 ('paddle','sell_products_online',84),
 ('paddle','creator_business',80),
 ('lemon-squeezy','creator_business',88),
 ('lemon-squeezy','sell_products_online',82),
 ('lemon-squeezy','software_product',76),
 ('gumroad','creator_business',92),
 ('gumroad','solo_consulting',70),
 ('gumroad','sell_products_online',68),
 ('teachable-starter','creator_business',90),
 ('teachable-starter','solo_consulting',78),
 ('thinkific-basic','creator_business',88),
 ('thinkific-basic','client_services',72)
) AS m(slug, goal_key, score)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v7', p.id, m.tag
FROM (VALUES
 ('framer-basic','no_code'),
 ('framer-basic','best_of_breed'),
 ('webflow-basic','api_first'),
 ('webflow-basic','best_of_breed'),
 ('squarespace-basic','all_in_one'),
 ('squarespace-basic','no_code'),
 ('ecwid-starter','no_code'),
 ('ecwid-starter','all_in_one'),
 ('shopify-basic','all_in_one'),
 ('shopify-basic','best_of_breed'),
 ('bigcommerce-core','api_first'),
 ('bigcommerce-core','best_of_breed'),
 ('stripe','api_first'),
 ('stripe','best_of_breed'),
 ('paddle','all_in_one'),
 ('paddle','no_code'),
 ('lemon-squeezy','no_code'),
 ('lemon-squeezy','all_in_one'),
 ('gumroad','no_code'),
 ('teachable-starter','all_in_one'),
 ('teachable-starter','no_code'),
 ('thinkific-basic','best_of_breed'),
 ('thinkific-basic','no_code')
) AS m(slug, tag)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v7', p.id, m.group_key
FROM (VALUES
 ('framer-basic','site_suite'),
 ('webflow-basic','site_suite'),
 ('squarespace-basic','site_suite'),
 ('ecwid-starter','store_suite'),
 ('shopify-basic','store_suite'),
 ('bigcommerce-core','store_suite'),
 ('stripe','payment_suite'),
 ('paddle','payment_suite'),
 ('lemon-squeezy','payment_suite'),
 ('gumroad','course_suite'),
 ('teachable-starter','course_suite'),
 ('thinkific-basic','course_suite')
) AS m(slug, group_key)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

-- Retire before activate: one active policy per vertical is enforced by a
-- unique index, so promoting while saas-v6 is still live fails the transaction.
UPDATE recommendation.policy_versions
SET workflow_status='retired', retired_at=now()
WHERE version='saas-v6' AND workflow_status='active';

UPDATE recommendation.policy_versions
SET workflow_status='active', activated_at=now()
WHERE version='saas-v7' AND workflow_status='draft';

COMMIT;
