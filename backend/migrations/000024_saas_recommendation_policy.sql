-- The SaaS vertical: recommending a coherent software stack for a business
-- rather than equipment for a room. The engine is unchanged; this is entirely
-- policy and catalog structure.
--
-- The mapping onto the existing model is direct: a tool's features are
-- capabilities, its integrations are compatibility, tools that overlap share a
-- redundancy group, and monthly spend is the budget constraint. space_match
-- and noise carry zero weight because software has no footprint and makes no
-- noise; spatial_constraints is false so the engine skips those checks
-- entirely rather than scoring them as neutral.

-- A category belongs to exactly one vertical. Without this the public catalog
-- in a SaaS deployment would list squat racks beside CRMs, because migrations
-- create every vertical's categories in the same database. Scoping lives in
-- catalog rather than being derived from recommendation.category_policies so
-- the catalog module keeps no dependency on the recommendation schema.
ALTER TABLE catalog.categories
    ADD COLUMN vertical_key text NOT NULL DEFAULT 'fitness'
        CHECK (vertical_key ~ '^[a-z][a-z0-9_]*$');

COMMENT ON COLUMN catalog.categories.vertical_key IS
    'Vertical that owns this category. The application serves one vertical, selected by RECOMMENDATION_VERTICAL.';

CREATE INDEX categories_vertical_active_idx ON catalog.categories (vertical_key, sort_order) WHERE is_active = true;

INSERT INTO catalog.categories (name, slug, description, sort_order, is_physical, vertical_key) VALUES
 ('CRM','crm','Customer relationship management tools.',0,false,'saas'),
 ('Email marketing','email-marketing','Audience email, broadcasts and automation.',1,false,'saas'),
 ('Project management','project-management','Project, task and delivery tracking.',2,false,'saas'),
 ('Accounting and invoicing','accounting-invoicing','Invoicing, bookkeeping and accounting.',3,false,'saas'),
 ('Ecommerce platform','ecommerce-platform','Online storefronts and product catalogs.',4,false,'saas'),
 ('Website builder','website-builder','Sites, pages and publishing.',5,false,'saas'),
 ('Help desk','help-desk','Customer support ticketing and inboxes.',6,false,'saas'),
 ('Analytics','analytics','Product and marketing analytics.',7,false,'saas'),
 ('Scheduling','scheduling','Meeting scheduling and calendar booking.',8,false,'saas'),
 ('Automation','automation','Workflow automation between tools.',9,false,'saas'),
 ('Team communication','team-communication','Team chat and meetings.',10,false,'saas'),
 ('Design tools','design-tools','Visual design and brand assets.',11,false,'saas'),
 ('Course platform','course-platform','Course hosting and digital product delivery.',12,false,'saas'),
 ('SEO tools','seo-tools','Keyword, ranking and content research.',13,false,'saas'),
 ('Payments','payments','Payment processing and checkout.',14,false,'saas');

INSERT INTO recommendation.policy_versions (
    version, vertical_key, workflow_status, spatial_constraints,
    goal_match_weight, budget_match_weight, space_match_weight,
    experience_match_weight, preference_match_weight, quality_weight,
    value_weight, durability_weight, compatibility_weight,
    portability_weight, noise_weight, priority_boost_percent,
    maximum_setup_items, candidates_per_slot, optional_slot_bonus,
    published_at, activated_at
) VALUES ('saas-v1','saas','draft',false,22,14,0,12,8,10,10,8,14,2,0,150,6,12,8,now(),now());

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label) VALUES
 ('saas-v1','crm','Customer relationship management'),
 ('saas-v1','email_marketing','Email marketing'),
 ('saas-v1','marketing_automation','Marketing automation'),
 ('saas-v1','project_management','Project management'),
 ('saas-v1','task_tracking','Task tracking'),
 ('saas-v1','time_tracking','Time tracking'),
 ('saas-v1','invoicing','Invoicing'),
 ('saas-v1','accounting','Accounting'),
 ('saas-v1','payments','Payment processing'),
 ('saas-v1','online_store','Online store'),
 ('saas-v1','website_builder','Website builder'),
 ('saas-v1','landing_pages','Landing pages'),
 ('saas-v1','form_builder','Form builder'),
 ('saas-v1','scheduling','Scheduling'),
 ('saas-v1','help_desk','Help desk'),
 ('saas-v1','live_chat','Live chat'),
 ('saas-v1','knowledge_base','Knowledge base'),
 ('saas-v1','product_analytics','Product analytics'),
 ('saas-v1','seo_research','SEO research'),
 ('saas-v1','social_scheduling','Social scheduling'),
 ('saas-v1','course_hosting','Course hosting'),
 ('saas-v1','file_storage','File storage'),
 ('saas-v1','esignature','Electronic signature'),
 ('saas-v1','team_chat','Team chat'),
 ('saas-v1','video_meetings','Video meetings'),
 ('saas-v1','workflow_automation','Workflow automation'),
 ('saas-v1','design_tool','Design tool');

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label) VALUES
 ('saas-v1','client_services','running a client services business'),
 ('saas-v1','sell_products_online','selling products online'),
 ('saas-v1','creator_business','running a creator business'),
 ('saas-v1','software_product','running a software product'),
 ('saas-v1','solo_consulting','working solo as a consultant');

INSERT INTO recommendation.policy_setup_roles
 (policy_version, goal_key, role_key, label, is_required, sort_order) VALUES
 ('saas-v1','client_services','client_management','Client management',true,0),
 ('saas-v1','client_services','delivery','Delivery',true,1),
 ('saas-v1','client_services','billing','Billing',true,2),
 ('saas-v1','client_services','communication','Communication',false,3),
 ('saas-v1','client_services','booking','Booking',false,4),
 ('saas-v1','sell_products_online','storefront','Storefront',true,0),
 ('saas-v1','sell_products_online','checkout','Checkout',true,1),
 ('saas-v1','sell_products_online','customer_email','Customer email',true,2),
 ('saas-v1','sell_products_online','support','Support',false,3),
 ('saas-v1','sell_products_online','measurement','Measurement',false,4),
 ('saas-v1','creator_business','audience_email','Audience email',true,0),
 ('saas-v1','creator_business','publishing','Publishing',true,1),
 ('saas-v1','creator_business','monetization','Monetization',true,2),
 ('saas-v1','creator_business','distribution','Distribution',false,3),
 ('saas-v1','creator_business','design','Design',false,4),
 ('saas-v1','software_product','measurement','Measurement',true,0),
 ('saas-v1','software_product','support','Support',true,1),
 ('saas-v1','software_product','delivery','Delivery',true,2),
 ('saas-v1','software_product','checkout','Checkout',false,3),
 ('saas-v1','software_product','automation','Automation',false,4),
 ('saas-v1','solo_consulting','client_management','Client management',true,0),
 ('saas-v1','solo_consulting','billing','Billing',true,1),
 ('saas-v1','solo_consulting','booking','Booking',true,2),
 ('saas-v1','solo_consulting','documents','Documents',false,3),
 ('saas-v1','solo_consulting','effort_tracking','Effort tracking',false,4);

INSERT INTO recommendation.policy_setup_role_capabilities
 (policy_version, goal_key, role_key, capability_key) VALUES
 ('saas-v1','client_services','client_management','crm'),
 ('saas-v1','client_services','delivery','project_management'),
 ('saas-v1','client_services','billing','invoicing'),
 ('saas-v1','client_services','communication','team_chat'),
 ('saas-v1','client_services','booking','scheduling'),
 ('saas-v1','sell_products_online','storefront','online_store'),
 ('saas-v1','sell_products_online','checkout','payments'),
 ('saas-v1','sell_products_online','customer_email','email_marketing'),
 ('saas-v1','sell_products_online','support','help_desk'),
 ('saas-v1','sell_products_online','measurement','product_analytics'),
 ('saas-v1','creator_business','audience_email','email_marketing'),
 ('saas-v1','creator_business','publishing','website_builder'),
 ('saas-v1','creator_business','monetization','course_hosting'),
 ('saas-v1','creator_business','distribution','social_scheduling'),
 ('saas-v1','creator_business','design','design_tool'),
 ('saas-v1','software_product','measurement','product_analytics'),
 ('saas-v1','software_product','support','help_desk'),
 ('saas-v1','software_product','delivery','project_management'),
 ('saas-v1','software_product','checkout','payments'),
 ('saas-v1','software_product','automation','workflow_automation'),
 ('saas-v1','solo_consulting','client_management','crm'),
 ('saas-v1','solo_consulting','billing','invoicing'),
 ('saas-v1','solo_consulting','booking','scheduling'),
 ('saas-v1','solo_consulting','documents','esignature'),
 ('saas-v1','solo_consulting','effort_tracking','time_tracking');

INSERT INTO recommendation.policy_preference_tags (policy_version, tag_key, label) VALUES
 ('saas-v1','all_in_one','Prefer one suite over many tools'),
 ('saas-v1','best_of_breed','Prefer the strongest tool per job'),
 ('saas-v1','open_source','Prefer open source'),
 ('saas-v1','no_code','Prefer no-code tools'),
 ('saas-v1','api_first','Prefer API-first tools'),
 ('saas-v1','privacy_focused','Prefer privacy-focused vendors'),
 ('saas-v1','eu_hosted','Prefer EU-hosted data');

-- durability is reused as vendor stability and portability as data
-- portability. Both are about how safe the commitment is rather than how the
-- product performs today, which is the same question the fitness vertical
-- asked about build quality and moving equipment.
INSERT INTO recommendation.policy_priorities
 (policy_version, priority_key, label, reason_code, reason_message, reason_dimension, reason_threshold, sort_order) VALUES
 ('saas-v1','budget','Budget','priority.value','Strong value for the available budget','value',85,0),
 ('saas-v1','ease_of_use','Ease of use','priority.ease_of_use','Straightforward to adopt without specialist help','experience_match',85,1),
 ('saas-v1','integrations','Integrations','priority.integrations','Connects cleanly with the rest of your stack','compatibility',85,2),
 ('saas-v1','reliability','Reliability','priority.reliability','Strong structured reliability score','quality',85,3),
 ('saas-v1','vendor_stability','Vendor stability','priority.vendor_stability','Backed by an established vendor','durability',85,4),
 ('saas-v1','data_portability','Data portability','priority.data_portability','Your data stays exportable if you leave','portability',85,5);

INSERT INTO recommendation.policy_priority_dimensions (policy_version, priority_key, dimension) VALUES
 ('saas-v1','budget','budget_match'), ('saas-v1','budget','value'),
 ('saas-v1','ease_of_use','experience_match'),
 ('saas-v1','integrations','compatibility'),
 ('saas-v1','reliability','quality'),
 ('saas-v1','vendor_stability','durability'),
 ('saas-v1','data_portability','portability');

-- Redundancy groups stop the optimizer proposing two tools that do the same
-- job, which is the software equivalent of recommending two squat racks.
INSERT INTO recommendation.policy_redundancy_groups (policy_version, group_key, label) VALUES
 ('saas-v1','crm_suite','CRM'),
 ('saas-v1','email_suite','Email marketing'),
 ('saas-v1','project_suite','Project management'),
 ('saas-v1','accounting_suite','Accounting and invoicing'),
 ('saas-v1','store_suite','Online store'),
 ('saas-v1','site_suite','Website and pages'),
 ('saas-v1','support_suite','Customer support'),
 ('saas-v1','analytics_suite','Analytics'),
 ('saas-v1','scheduling_suite','Scheduling'),
 ('saas-v1','automation_suite','Automation'),
 ('saas-v1','chat_suite','Team communication'),
 ('saas-v1','design_suite','Design'),
 ('saas-v1','course_suite','Course hosting'),
 ('saas-v1','seo_suite','SEO'),
 ('saas-v1','payment_suite','Payments');

INSERT INTO recommendation.category_policies (policy_version, category_id, support_status)
SELECT 'saas-v1', id, 'supported' FROM catalog.categories WHERE slug IN (
 'crm','email-marketing','project-management','accounting-invoicing','ecommerce-platform',
 'website-builder','help-desk','analytics','scheduling','automation','team-communication',
 'design-tools','course-platform','seo-tools','payments');

INSERT INTO recommendation.category_policy_capabilities (policy_version, category_id, capability_key)
SELECT 'saas-v1', categories.id, mapping.capability
FROM (VALUES
 ('crm','crm'),
 ('email-marketing','email_marketing'), ('email-marketing','marketing_automation'),
 ('project-management','project_management'), ('project-management','task_tracking'),
 ('accounting-invoicing','invoicing'), ('accounting-invoicing','accounting'),
 ('ecommerce-platform','online_store'),
 ('website-builder','website_builder'), ('website-builder','landing_pages'),
 ('help-desk','help_desk'), ('help-desk','live_chat'), ('help-desk','knowledge_base'),
 ('analytics','product_analytics'),
 ('scheduling','scheduling'),
 ('automation','workflow_automation'),
 ('team-communication','team_chat'), ('team-communication','video_meetings'),
 ('design-tools','design_tool'),
 ('course-platform','course_hosting'),
 ('seo-tools','seo_research'),
 ('payments','payments')
) AS mapping(category_slug, capability)
JOIN catalog.categories categories ON categories.slug = mapping.category_slug;

INSERT INTO recommendation.category_redundancy_groups (policy_version, category_id, group_key)
SELECT 'saas-v1', categories.id, mapping.group_key
FROM (VALUES
 ('crm','crm_suite'),
 ('email-marketing','email_suite'),
 ('project-management','project_suite'),
 ('accounting-invoicing','accounting_suite'),
 ('ecommerce-platform','store_suite'),
 ('website-builder','site_suite'),
 ('help-desk','support_suite'),
 ('analytics','analytics_suite'),
 ('scheduling','scheduling_suite'),
 ('automation','automation_suite'),
 ('team-communication','chat_suite'),
 ('design-tools','design_suite'),
 ('course-platform','course_suite'),
 ('seo-tools','seo_suite'),
 ('payments','payment_suite')
) AS mapping(category_slug, group_key)
JOIN catalog.categories categories ON categories.slug = mapping.category_slug;

-- Deliberately left in draft. An active policy is immutable, so activating it
-- here would permanently lock out the product policies that every catalog
-- entry needs. Activation happens once a catalog exists, either through the
-- admin policy workflow for real products or through the fictional demo seed
-- for local development, matching how the fitness policy is handled.
