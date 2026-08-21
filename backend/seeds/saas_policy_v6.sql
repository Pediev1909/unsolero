-- saas-v6: saas-v5 plus the nine products that open help desk, scheduling and
-- analytics on 2026-08-21.
--
-- Same reason as every version before it: an active policy is immutable, so
-- binding new products means publishing a new version rather than amending the
-- live one. Configuration is copied column-by-column from saas-v5 so that a
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
SELECT 'saas-v6', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v5. Opens help desk, scheduling and analytics with three products each.'
FROM recommendation.policy_versions WHERE version = 'saas-v5'
ON CONFLICT (version) DO NOTHING;

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v6', capability_key, label FROM recommendation.policy_capabilities WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v6', goal_key, label FROM recommendation.policy_goals WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v6', tag_key, label FROM recommendation.policy_preference_tags WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v6', group_key, label FROM recommendation.policy_redundancy_groups WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v6', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v6', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v6', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v6', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v6', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v6', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v6', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v5 ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v6', product_id, fact_revision_id, score_revision_id
FROM recommendation.product_policies WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v6', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v6', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v6', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v6', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version='saas-v5'
ON CONFLICT DO NOTHING;

-- --------------------------------------------------------- the nine new ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v6', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('freshdesk-growth','tidio-starter','help-scout-standard',
               'zoho-bookings-basic','calendly-standard','cal-com-teams',
               'fathom-analytics','simple-analytics','umami-cloud-pro')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v6', p.id, m.capability, m.relation
FROM (VALUES
 ('freshdesk-growth','help_desk','provides'),
 ('freshdesk-growth','knowledge_base','provides'),
 ('freshdesk-growth','live_chat','provides'),
 ('freshdesk-growth','crm','compatible_with'),
 ('tidio-starter','help_desk','provides'),
 ('tidio-starter','live_chat','provides'),
 ('tidio-starter','online_store','compatible_with'),
 ('tidio-starter','crm','compatible_with'),
 ('help-scout-standard','help_desk','provides'),
 ('help-scout-standard','knowledge_base','provides'),
 ('help-scout-standard','crm','compatible_with'),
 ('zoho-bookings-basic','scheduling','provides'),
 ('zoho-bookings-basic','payments','compatible_with'),
 ('zoho-bookings-basic','crm','compatible_with'),
 ('calendly-standard','scheduling','provides'),
 ('calendly-standard','video_meetings','compatible_with'),
 ('calendly-standard','crm','compatible_with'),
 ('calendly-standard','payments','compatible_with'),
 ('cal-com-teams','scheduling','provides'),
 ('cal-com-teams','video_meetings','compatible_with'),
 ('cal-com-teams','payments','compatible_with'),
 ('fathom-analytics','product_analytics','provides'),
 ('fathom-analytics','website_builder','compatible_with'),
 ('simple-analytics','product_analytics','provides'),
 ('simple-analytics','website_builder','compatible_with'),
 ('umami-cloud-pro','product_analytics','provides'),
 ('umami-cloud-pro','website_builder','compatible_with')
) AS m(slug, capability, relation)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v6', p.id, m.goal_key, m.score
FROM (VALUES
 ('freshdesk-growth','sell_products_online',88),
 ('freshdesk-growth','software_product',86),
 ('freshdesk-growth','client_services',78),
 ('tidio-starter','sell_products_online',92),
 ('tidio-starter','creator_business',76),
 ('tidio-starter','client_services',70),
 ('help-scout-standard','software_product',90),
 ('help-scout-standard','client_services',84),
 ('help-scout-standard','solo_consulting',80),
 ('zoho-bookings-basic','solo_consulting',88),
 ('zoho-bookings-basic','client_services',84),
 ('zoho-bookings-basic','creator_business',72),
 ('calendly-standard','solo_consulting',94),
 ('calendly-standard','client_services',90),
 ('calendly-standard','creator_business',80),
 ('cal-com-teams','software_product',88),
 ('cal-com-teams','solo_consulting',78),
 ('cal-com-teams','client_services',76),
 ('fathom-analytics','creator_business',90),
 ('fathom-analytics','software_product',82),
 ('fathom-analytics','sell_products_online',78),
 ('simple-analytics','creator_business',88),
 ('simple-analytics','solo_consulting',80),
 ('simple-analytics','software_product',72),
 ('umami-cloud-pro','software_product',88),
 ('umami-cloud-pro','creator_business',82),
 ('umami-cloud-pro','sell_products_online',74)
) AS m(slug, goal_key, score)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v6', p.id, m.tag
FROM (VALUES
 ('freshdesk-growth','all_in_one'),
 ('freshdesk-growth','no_code'),
 ('tidio-starter','no_code'),
 ('tidio-starter','eu_hosted'),
 ('help-scout-standard','best_of_breed'),
 ('help-scout-standard','no_code'),
 ('zoho-bookings-basic','all_in_one'),
 ('zoho-bookings-basic','eu_hosted'),
 ('calendly-standard','best_of_breed'),
 ('calendly-standard','no_code'),
 ('cal-com-teams','open_source'),
 ('cal-com-teams','api_first'),
 ('cal-com-teams','privacy_focused'),
 ('fathom-analytics','privacy_focused'),
 ('fathom-analytics','best_of_breed'),
 ('simple-analytics','privacy_focused'),
 ('simple-analytics','eu_hosted'),
 ('simple-analytics','no_code'),
 ('umami-cloud-pro','open_source'),
 ('umami-cloud-pro','privacy_focused'),
 ('umami-cloud-pro','api_first')
) AS m(slug, tag)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v6', p.id, m.group_key
FROM (VALUES
 ('freshdesk-growth','support_suite'),
 ('tidio-starter','support_suite'),
 ('help-scout-standard','support_suite'),
 ('zoho-bookings-basic','scheduling_suite'),
 ('calendly-standard','scheduling_suite'),
 ('cal-com-teams','scheduling_suite'),
 ('fathom-analytics','analytics_suite'),
 ('simple-analytics','analytics_suite'),
 ('umami-cloud-pro','analytics_suite')
) AS m(slug, group_key)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

-- Retire before activate: one active policy per vertical is enforced by a
-- unique index, so promoting while saas-v5 is still live fails the transaction.
UPDATE recommendation.policy_versions
SET workflow_status='retired', retired_at=now()
WHERE version='saas-v5' AND workflow_status='active';

UPDATE recommendation.policy_versions
SET workflow_status='active', activated_at=now()
WHERE version='saas-v6' AND workflow_status='draft';

COMMIT;
