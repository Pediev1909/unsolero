-- saas-v9: saas-v8 plus the twelve products that fill the last four empty
-- categories -- automation, design tools, SEO tools and team communication.
--
-- Same reason as every version before it: an active policy is immutable, so
-- binding new products means publishing a new version rather than amending the
-- live one. Configuration is copied column-by-column from saas-v8 so that a
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
SELECT 'saas-v9', goal_match_weight, budget_match_weight, space_match_weight,
       experience_match_weight, preference_match_weight, quality_weight,
       value_weight, durability_weight, compatibility_weight, portability_weight,
       noise_weight, priority_boost_percent, maximum_setup_items,
       candidates_per_slot, optional_slot_bonus, vertical_key, 'draft',
       spatial_constraints, now(),
       'Identical configuration to saas-v8. Adds automation, design-tools, seo-tools and team-communication, which leaves no empty category in the vertical.'
FROM recommendation.policy_versions WHERE version = 'saas-v8'
ON CONFLICT (version) DO NOTHING;

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label)
SELECT 'saas-v9', capability_key, label FROM recommendation.policy_capabilities WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label)
SELECT 'saas-v9', goal_key, label FROM recommendation.policy_goals WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label)
SELECT 'saas-v9', tag_key, label FROM recommendation.policy_preference_tags WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label)
SELECT 'saas-v9', group_key, label FROM recommendation.policy_redundancy_groups WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priorities (
    policy_version, priority_key, label, reason_code, reason_message,
    reason_dimension, reason_threshold, sort_order)
SELECT 'saas-v9', priority_key, label, reason_code, reason_message,
       reason_dimension, reason_threshold, sort_order
FROM recommendation.policy_priorities WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension)
SELECT 'saas-v9', priority_key, dimension
FROM recommendation.policy_priority_dimensions WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_roles (policy_version, goal_key, role_key, label, is_required, sort_order)
SELECT 'saas-v9', goal_key, role_key, label, is_required, sort_order
FROM recommendation.policy_setup_roles WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.policy_setup_role_capabilities (policy_version, goal_key, role_key, capability_key)
SELECT 'saas-v9', goal_key, role_key, capability_key
FROM recommendation.policy_setup_role_capabilities WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policies (
    policy_version, category_id, support_status, requires_storage_footprint,
    requires_operating_clearance, requires_safety_clearance, requires_access_width)
SELECT 'saas-v9', category_id, support_status, requires_storage_footprint,
       requires_operating_clearance, requires_safety_clearance, requires_access_width
FROM recommendation.category_policies WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v9', category_id, capability_key
FROM recommendation.category_policy_capabilities WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v9', category_id, group_key
FROM recommendation.category_redundancy_groups WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

-- ------------------------------------------ products already in saas-v8 ----
-- Revisions come from the catalog's current published pointers, not from
-- saas-v8's frozen ones. Copying the old pointers would carry a superseded
-- description into the new version and defeat the exercise.
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v9', p.id, p.published_fact_revision_id, p.published_score_revision_id
FROM recommendation.product_policies pp
JOIN catalog.products p ON p.id = pp.product_id
WHERE pp.policy_version='saas-v8'
  AND p.published_fact_revision_id IS NOT NULL
  AND p.published_score_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v9', product_id, capability_key, relation_type
FROM recommendation.product_policy_capabilities WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v9', product_id, goal_key, match_score
FROM recommendation.product_goal_support WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v9', product_id, preference_key
FROM recommendation.product_preference_tags WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v9', product_id, group_key
FROM recommendation.product_redundancy_groups WHERE policy_version='saas-v8'
ON CONFLICT DO NOTHING;


-- ------------------------------------------------------ the twelve new ----
INSERT INTO recommendation.product_policies (policy_version, product_id, fact_revision_id, score_revision_id)
SELECT 'saas-v9', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products
WHERE slug IN ('make-core','zapier-professional','n8n-starter','sketch-standard','canva-pro','figma-professional',
               'ahrefs-starter','se-ranking-core','semrush-seo','microsoft-teams-essentials',
               'google-workspace-business-starter','slack-pro')
  AND published_fact_revision_id IS NOT NULL
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_policy_capabilities (policy_version, product_id, capability_key, relation_type)
SELECT 'saas-v9', p.id, m.capability, m.relation
FROM (VALUES
 ('make-core','workflow_automation','provides'),
 ('make-core','crm','compatible_with'),
 ('make-core','email_marketing','compatible_with'),
 ('make-core','online_store','compatible_with'),
 ('zapier-professional','workflow_automation','provides'),
 ('zapier-professional','crm','compatible_with'),
 ('zapier-professional','email_marketing','compatible_with'),
 ('zapier-professional','help_desk','compatible_with'),
 ('n8n-starter','workflow_automation','provides'),
 ('n8n-starter','crm','compatible_with'),
 ('n8n-starter','product_analytics','compatible_with'),
 ('sketch-standard','design_tool','provides'),
 ('sketch-standard','website_builder','compatible_with'),
 ('canva-pro','design_tool','provides'),
 ('canva-pro','social_scheduling','provides'),
 ('canva-pro','website_builder','compatible_with'),
 ('figma-professional','design_tool','provides'),
 ('figma-professional','website_builder','compatible_with'),
 ('ahrefs-starter','seo_research','provides'),
 ('ahrefs-starter','website_builder','compatible_with'),
 ('se-ranking-core','seo_research','provides'),
 ('se-ranking-core','product_analytics','compatible_with'),
 ('semrush-seo','seo_research','provides'),
 ('semrush-seo','social_scheduling','provides'),
 ('semrush-seo','product_analytics','compatible_with'),
 ('microsoft-teams-essentials','team_chat','provides'),
 ('microsoft-teams-essentials','video_meetings','provides'),
 ('microsoft-teams-essentials','file_storage','compatible_with'),
 ('google-workspace-business-starter','team_chat','provides'),
 ('google-workspace-business-starter','video_meetings','provides'),
 ('google-workspace-business-starter','file_storage','provides'),
 ('slack-pro','team_chat','provides'),
 ('slack-pro','video_meetings','provides'),
 ('slack-pro','workflow_automation','compatible_with'),
 ('slack-pro','project_management','compatible_with')
) AS m(slug, capability, relation)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_goal_support (policy_version, product_id, goal_key, match_score)
SELECT 'saas-v9', p.id, m.goal_key, m.score
FROM (VALUES
 ('make-core','software_product',86),
 ('make-core','sell_products_online',82),
 ('make-core','solo_consulting',76),
 ('zapier-professional','client_services',88),
 ('zapier-professional','sell_products_online',86),
 ('zapier-professional','solo_consulting',84),
 ('n8n-starter','software_product',92),
 ('n8n-starter','client_services',70),
 ('sketch-standard','software_product',82),
 ('sketch-standard','client_services',76),
 ('canva-pro','creator_business',94),
 ('canva-pro','solo_consulting',86),
 ('canva-pro','sell_products_online',84),
 ('figma-professional','software_product',94),
 ('figma-professional','client_services',86),
 ('figma-professional','creator_business',72),
 ('ahrefs-starter','creator_business',90),
 ('ahrefs-starter','sell_products_online',84),
 ('ahrefs-starter','solo_consulting',80),
 ('se-ranking-core','client_services',86),
 ('se-ranking-core','sell_products_online',80),
 ('semrush-seo','client_services',90),
 ('semrush-seo','sell_products_online',88),
 ('semrush-seo','creator_business',78),
 ('microsoft-teams-essentials','client_services',82),
 ('microsoft-teams-essentials','solo_consulting',70),
 ('google-workspace-business-starter','client_services',90),
 ('google-workspace-business-starter','solo_consulting',88),
 ('google-workspace-business-starter','creator_business',76),
 ('slack-pro','software_product',92),
 ('slack-pro','client_services',84),
 ('slack-pro','creator_business',72)
) AS m(slug, goal_key, score)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_preference_tags (policy_version, product_id, preference_key)
SELECT 'saas-v9', p.id, m.tag
FROM (VALUES
 ('make-core','no_code'),
 ('make-core','best_of_breed'),
 ('make-core','eu_hosted'),
 ('zapier-professional','no_code'),
 ('zapier-professional','all_in_one'),
 ('n8n-starter','open_source'),
 ('n8n-starter','api_first'),
 ('n8n-starter','privacy_focused'),
 ('n8n-starter','eu_hosted'),
 ('sketch-standard','best_of_breed'),
 ('canva-pro','no_code'),
 ('canva-pro','all_in_one'),
 ('figma-professional','best_of_breed'),
 ('figma-professional','api_first'),
 ('ahrefs-starter','best_of_breed'),
 ('se-ranking-core','all_in_one'),
 ('se-ranking-core','no_code'),
 ('semrush-seo','all_in_one'),
 ('semrush-seo','best_of_breed'),
 ('microsoft-teams-essentials','all_in_one'),
 ('google-workspace-business-starter','all_in_one'),
 ('google-workspace-business-starter','no_code'),
 ('slack-pro','best_of_breed'),
 ('slack-pro','api_first')
) AS m(slug, tag)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

INSERT INTO recommendation.product_redundancy_groups (policy_version, product_id, group_key)
SELECT 'saas-v9', p.id, m.group_key
FROM (VALUES
 ('make-core','automation_suite'),
 ('zapier-professional','automation_suite'),
 ('n8n-starter','automation_suite'),
 ('sketch-standard','design_suite'),
 ('canva-pro','design_suite'),
 ('figma-professional','design_suite'),
 ('ahrefs-starter','seo_suite'),
 ('se-ranking-core','seo_suite'),
 ('semrush-seo','seo_suite'),
 ('microsoft-teams-essentials','chat_suite'),
 ('google-workspace-business-starter','chat_suite'),
 ('slack-pro','chat_suite')
) AS m(slug, group_key)
JOIN catalog.products p ON p.slug = m.slug
ON CONFLICT DO NOTHING;

-- Retire before activate: one active policy per vertical is enforced by a
-- unique index, so promoting while saas-v8 is still live fails the transaction.
UPDATE recommendation.policy_versions
SET workflow_status='retired', retired_at=now()
WHERE version='saas-v8' AND workflow_status='active';

UPDATE recommendation.policy_versions
SET workflow_status='active', activated_at=now()
WHERE version='saas-v9' AND workflow_status='draft';

COMMIT;
