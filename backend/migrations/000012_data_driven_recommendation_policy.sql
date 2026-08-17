-- Versioned, data-driven recommendation policy. Commercial schemas are
-- intentionally absent from every policy foreign key and projection.

INSERT INTO identity.roles (role_key, description) VALUES
    ('policy_editor', 'Creates and submits recommendation policy revisions.'),
    ('policy_reviewer', 'Approves and activates recommendation policy revisions.')
ON CONFLICT (role_key) DO NOTHING;

ALTER TABLE recommendation.policy_versions
    ADD COLUMN vertical_key text NOT NULL DEFAULT 'fitness' CHECK (vertical_key ~ '^[a-z][a-z0-9_]*$'),
    ADD COLUMN workflow_status text NOT NULL DEFAULT 'retired'
        CHECK (workflow_status IN ('draft','in_review','approved','active','retired','rejected')),
    ADD COLUMN created_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    ADD COLUMN submitted_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    ADD COLUMN reviewed_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    ADD COLUMN activated_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    ADD COLUMN submitted_at timestamptz,
    ADD COLUMN reviewed_at timestamptz,
    ADD COLUMN activated_at timestamptz,
    ADD COLUMN review_note text NOT NULL DEFAULT '' CHECK (char_length(review_note) <= 2000);

UPDATE recommendation.policy_versions
SET workflow_status='retired', retired_at=COALESCE(retired_at, now())
WHERE version='home-gym-v1';

CREATE UNIQUE INDEX policy_versions_one_active_vertical_idx
    ON recommendation.policy_versions(vertical_key) WHERE workflow_status='active';

CREATE TABLE recommendation.policy_capabilities (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    capability_key text NOT NULL CHECK (capability_key ~ '^[a-z][a-z0-9_]*$'),
    label text NOT NULL CHECK (char_length(btrim(label)) BETWEEN 1 AND 120),
    PRIMARY KEY (policy_version, capability_key)
);

CREATE TABLE recommendation.policy_goals (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    goal_key text NOT NULL CHECK (goal_key ~ '^[a-z][a-z0-9_]*$'),
    label text NOT NULL CHECK (char_length(btrim(label)) BETWEEN 1 AND 120),
    PRIMARY KEY (policy_version, goal_key)
);

CREATE TABLE recommendation.policy_setup_roles (
    policy_version text NOT NULL,
    goal_key text NOT NULL,
    role_key text NOT NULL CHECK (role_key ~ '^[a-z][a-z0-9_]*$'),
    label text NOT NULL CHECK (char_length(btrim(label)) BETWEEN 1 AND 120),
    is_required boolean NOT NULL,
    sort_order smallint NOT NULL CHECK (sort_order >= 0),
    PRIMARY KEY (policy_version, goal_key, role_key),
    UNIQUE (policy_version, goal_key, sort_order),
    FOREIGN KEY (policy_version, goal_key)
        REFERENCES recommendation.policy_goals(policy_version, goal_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.policy_setup_role_capabilities (
    policy_version text NOT NULL,
    goal_key text NOT NULL,
    role_key text NOT NULL,
    capability_key text NOT NULL,
    PRIMARY KEY (policy_version, goal_key, role_key, capability_key),
    FOREIGN KEY (policy_version, goal_key, role_key)
        REFERENCES recommendation.policy_setup_roles(policy_version, goal_key, role_key) ON DELETE RESTRICT,
    FOREIGN KEY (policy_version, capability_key)
        REFERENCES recommendation.policy_capabilities(policy_version, capability_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.category_policies (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    category_id uuid NOT NULL REFERENCES catalog.categories(id) ON DELETE RESTRICT,
    support_status text NOT NULL CHECK (support_status IN ('supported','unsupported')),
    requires_storage_footprint boolean NOT NULL DEFAULT false,
    requires_operating_clearance boolean NOT NULL DEFAULT false,
    requires_safety_clearance boolean NOT NULL DEFAULT false,
    requires_access_width boolean NOT NULL DEFAULT false,
    PRIMARY KEY (policy_version, category_id)
);

CREATE TABLE recommendation.category_policy_capabilities (
    policy_version text NOT NULL,
    category_id uuid NOT NULL,
    capability_key text NOT NULL,
    PRIMARY KEY (policy_version, category_id, capability_key),
    FOREIGN KEY (policy_version, category_id)
        REFERENCES recommendation.category_policies(policy_version, category_id) ON DELETE RESTRICT,
    FOREIGN KEY (policy_version, capability_key)
        REFERENCES recommendation.policy_capabilities(policy_version, capability_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.policy_redundancy_groups (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    group_key text NOT NULL CHECK (group_key ~ '^[a-z][a-z0-9_]*$'),
    label text NOT NULL CHECK (char_length(btrim(label)) BETWEEN 1 AND 120),
    PRIMARY KEY (policy_version, group_key)
);

CREATE TABLE recommendation.category_redundancy_groups (
    policy_version text NOT NULL,
    category_id uuid NOT NULL,
    group_key text NOT NULL,
    PRIMARY KEY (policy_version, category_id, group_key),
    FOREIGN KEY (policy_version, category_id)
        REFERENCES recommendation.category_policies(policy_version, category_id) ON DELETE RESTRICT,
    FOREIGN KEY (policy_version, group_key)
        REFERENCES recommendation.policy_redundancy_groups(policy_version, group_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.product_policies (
    policy_version text NOT NULL REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE RESTRICT,
    fact_revision_id uuid NOT NULL REFERENCES evidence.product_fact_revisions(id) ON DELETE RESTRICT,
    score_revision_id uuid NOT NULL REFERENCES evidence.score_revisions(id) ON DELETE RESTRICT,
    PRIMARY KEY (policy_version, product_id)
);

CREATE TABLE recommendation.product_policy_capabilities (
    policy_version text NOT NULL,
    product_id uuid NOT NULL,
    capability_key text NOT NULL,
    relation_type text NOT NULL CHECK (relation_type IN ('provides','requires','compatible_with','incompatible_with')),
    PRIMARY KEY (policy_version, product_id, capability_key, relation_type),
    FOREIGN KEY (policy_version, product_id)
        REFERENCES recommendation.product_policies(policy_version, product_id) ON DELETE RESTRICT,
    FOREIGN KEY (policy_version, capability_key)
        REFERENCES recommendation.policy_capabilities(policy_version, capability_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.product_goal_support (
    policy_version text NOT NULL,
    product_id uuid NOT NULL,
    goal_key text NOT NULL,
    match_score smallint NOT NULL CHECK (match_score BETWEEN 0 AND 100),
    PRIMARY KEY (policy_version, product_id, goal_key),
    FOREIGN KEY (policy_version, product_id)
        REFERENCES recommendation.product_policies(policy_version, product_id) ON DELETE RESTRICT,
    FOREIGN KEY (policy_version, goal_key)
        REFERENCES recommendation.policy_goals(policy_version, goal_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.product_preference_tags (
    policy_version text NOT NULL,
    product_id uuid NOT NULL,
    preference_key text NOT NULL CHECK (preference_key ~ '^[a-z][a-z0-9_]*$'),
    PRIMARY KEY (policy_version, product_id, preference_key),
    FOREIGN KEY (policy_version, product_id)
        REFERENCES recommendation.product_policies(policy_version, product_id) ON DELETE RESTRICT
);

CREATE TABLE recommendation.product_redundancy_groups (
    policy_version text NOT NULL,
    product_id uuid NOT NULL,
    group_key text NOT NULL,
    PRIMARY KEY (policy_version, product_id, group_key),
    FOREIGN KEY (policy_version, product_id)
        REFERENCES recommendation.product_policies(policy_version, product_id) ON DELETE RESTRICT,
    FOREIGN KEY (policy_version, group_key)
        REFERENCES recommendation.policy_redundancy_groups(policy_version, group_key) ON DELETE RESTRICT
);

CREATE TABLE recommendation.product_space_profiles (
    policy_version text NOT NULL,
    product_id uuid NOT NULL,
    footprint_length_mm integer NOT NULL CHECK (footprint_length_mm > 0),
    footprint_width_mm integer NOT NULL CHECK (footprint_width_mm > 0),
    footprint_height_mm integer NOT NULL CHECK (footprint_height_mm > 0),
    storage_length_mm integer CHECK (storage_length_mm > 0),
    storage_width_mm integer CHECK (storage_width_mm > 0),
    storage_height_mm integer CHECK (storage_height_mm > 0),
    operating_front_mm integer CHECK (operating_front_mm >= 0),
    operating_back_mm integer CHECK (operating_back_mm >= 0),
    operating_left_mm integer CHECK (operating_left_mm >= 0),
    operating_right_mm integer CHECK (operating_right_mm >= 0),
    operating_top_mm integer CHECK (operating_top_mm >= 0),
    safety_front_mm integer CHECK (safety_front_mm >= 0),
    safety_back_mm integer CHECK (safety_back_mm >= 0),
    safety_left_mm integer CHECK (safety_left_mm >= 0),
    safety_right_mm integer CHECK (safety_right_mm >= 0),
    safety_top_mm integer CHECK (safety_top_mm >= 0),
    minimum_room_height_mm integer CHECK (minimum_room_height_mm > 0),
    minimum_access_width_mm integer CHECK (minimum_access_width_mm > 0),
    overlap_group text CHECK (overlap_group IS NULL OR overlap_group ~ '^[a-z][a-z0-9_]*$'),
    PRIMARY KEY (policy_version, product_id),
    FOREIGN KEY (policy_version, product_id)
        REFERENCES recommendation.product_policies(policy_version, product_id) ON DELETE RESTRICT,
    CHECK ((storage_length_mm IS NULL AND storage_width_mm IS NULL AND storage_height_mm IS NULL)
        OR (storage_length_mm IS NOT NULL AND storage_width_mm IS NOT NULL AND storage_height_mm IS NOT NULL)),
    CHECK ((operating_front_mm IS NULL AND operating_back_mm IS NULL AND operating_left_mm IS NULL AND operating_right_mm IS NULL AND operating_top_mm IS NULL)
        OR (operating_front_mm IS NOT NULL AND operating_back_mm IS NOT NULL AND operating_left_mm IS NOT NULL AND operating_right_mm IS NOT NULL AND operating_top_mm IS NOT NULL)),
    CHECK ((safety_front_mm IS NULL AND safety_back_mm IS NULL AND safety_left_mm IS NULL AND safety_right_mm IS NULL AND safety_top_mm IS NULL)
        OR (safety_front_mm IS NOT NULL AND safety_back_mm IS NOT NULL AND safety_left_mm IS NOT NULL AND safety_right_mm IS NOT NULL AND safety_top_mm IS NOT NULL))
);

-- Immutable historical engine inputs not already stored in candidate_snapshots.
ALTER TABLE recommendation.candidate_snapshots
    ADD COLUMN policy_version text NOT NULL DEFAULT 'home-gym-v1'
        REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT,
    ADD COLUMN capabilities text[] NOT NULL DEFAULT '{}',
    ADD COLUMN requirements text[] NOT NULL DEFAULT '{}',
    ADD COLUMN compatible_with text[] NOT NULL DEFAULT '{}',
    ADD COLUMN incompatible_with text[] NOT NULL DEFAULT '{}',
    ADD COLUMN preference_tags text[] NOT NULL DEFAULT '{}',
    ADD COLUMN redundancy_groups text[] NOT NULL DEFAULT '{}',
    ADD COLUMN storage_length_mm integer,
    ADD COLUMN storage_width_mm integer,
    ADD COLUMN storage_height_mm integer,
    ADD COLUMN operating_clearance_mm integer[],
    ADD COLUMN safety_clearance_mm integer[],
    ADD COLUMN minimum_room_height_mm integer,
    ADD COLUMN minimum_access_width_mm integer,
    ADD COLUMN overlap_group text,
    ADD COLUMN requires_storage_footprint boolean NOT NULL DEFAULT false,
    ADD COLUMN requires_operating_clearance boolean NOT NULL DEFAULT false,
    ADD COLUMN requires_safety_clearance boolean NOT NULL DEFAULT false,
    ADD COLUMN requires_access_width boolean NOT NULL DEFAULT false,
    ADD CONSTRAINT candidate_snapshots_storage_check CHECK (
        (storage_length_mm IS NULL AND storage_width_mm IS NULL AND storage_height_mm IS NULL)
        OR (storage_length_mm > 0 AND storage_width_mm > 0 AND storage_height_mm > 0)),
    ADD CONSTRAINT candidate_snapshots_operating_check CHECK (
        operating_clearance_mm IS NULL OR cardinality(operating_clearance_mm)=5),
    ADD CONSTRAINT candidate_snapshots_safety_check CHECK (
        safety_clearance_mm IS NULL OR cardinality(safety_clearance_mm)=5);

ALTER TABLE recommendation.recommendation_sessions
    ADD COLUMN access_width_mm integer CHECK (access_width_mm IS NULL OR access_width_mm > 0);
ALTER TABLE recommendation.drafts
    ADD COLUMN access_width_mm integer CHECK (access_width_mm IS NULL OR access_width_mm > 0);

CREATE TABLE recommendation.candidate_snapshot_goal_support (
    recommendation_id uuid NOT NULL,
    product_id uuid NOT NULL,
    goal_key text NOT NULL,
    match_score smallint NOT NULL CHECK (match_score BETWEEN 0 AND 100),
    PRIMARY KEY (recommendation_id, product_id, goal_key),
    FOREIGN KEY (recommendation_id, product_id)
        REFERENCES recommendation.candidate_snapshots(recommendation_id, product_id) ON DELETE CASCADE
);

-- The legacy row remains referenced by historical recommendations. New runs
-- use this version, whose complete behavior is represented in tables below.
INSERT INTO recommendation.policy_versions (
    version, vertical_key, workflow_status, goal_match_weight, budget_match_weight,
    space_match_weight, experience_match_weight, preference_match_weight,
    quality_weight, value_weight, durability_weight, compatibility_weight,
    portability_weight, noise_weight, priority_boost_percent, maximum_setup_items,
    candidates_per_slot, optional_slot_bonus, published_at, activated_at
) VALUES ('fitness-v2','fitness','draft',20,12,12,10,8,8,9,7,10,2,2,150,4,12,8,now(),now());

INSERT INTO recommendation.policy_capabilities (policy_version, capability_key, label) VALUES
 ('fitness-v2','resistance_training','Resistance training'), ('fitness-v2','strength_training','Strength training'),
 ('fitness-v2','hypertrophy','Hypertrophy'), ('fitness-v2','supported_training','Supported training'),
 ('fitness-v2','barbell_training','Barbell training'), ('fitness-v2','weight_plates','Weight plates'),
 ('fitness-v2','safe_barbell_training','Safe barbell training'), ('fitness-v2','conditioning','Conditioning'),
 ('fitness-v2','mobility','Mobility'), ('fitness-v2','low_impact','Low impact'),
 ('fitness-v2','pull_up','Pull-up training'), ('fitness-v2','anchor_point','Anchor point');

INSERT INTO recommendation.policy_goals (policy_version, goal_key, label) VALUES
 ('fitness-v2','build_muscle','Build muscle'), ('fitness-v2','strength','Strength'),
 ('fitness-v2','general_fitness','General fitness'), ('fitness-v2','weight_loss','Weight loss'),
 ('fitness-v2','mobility','Mobility');

INSERT INTO recommendation.policy_setup_roles (policy_version,goal_key,role_key,label,is_required,sort_order) VALUES
 ('fitness-v2','build_muscle','primary','Primary resistance',true,0), ('fitness-v2','build_muscle','load','Load',false,1), ('fitness-v2','build_muscle','safety','Safety',false,2), ('fitness-v2','build_muscle','support','Support',false,3),
 ('fitness-v2','strength','primary','Primary resistance',true,0), ('fitness-v2','strength','load','Load',false,1), ('fitness-v2','strength','safety','Safety',false,2), ('fitness-v2','strength','support','Support',false,3),
 ('fitness-v2','general_fitness','resistance','Resistance',true,0), ('fitness-v2','general_fitness','conditioning','Conditioning',false,1), ('fitness-v2','general_fitness','mobility','Mobility',false,2), ('fitness-v2','general_fitness','support','Support',false,3),
 ('fitness-v2','weight_loss','conditioning','Conditioning',true,0), ('fitness-v2','weight_loss','resistance','Resistance',false,1), ('fitness-v2','weight_loss','mobility','Mobility',false,2),
 ('fitness-v2','mobility','mobility','Mobility',true,0), ('fitness-v2','mobility','resistance','Resistance',false,1), ('fitness-v2','mobility','conditioning','Conditioning',false,2);

INSERT INTO recommendation.policy_setup_role_capabilities VALUES
 ('fitness-v2','build_muscle','primary','hypertrophy'), ('fitness-v2','build_muscle','load','weight_plates'), ('fitness-v2','build_muscle','safety','safe_barbell_training'), ('fitness-v2','build_muscle','support','supported_training'),
 ('fitness-v2','strength','primary','strength_training'), ('fitness-v2','strength','load','weight_plates'), ('fitness-v2','strength','safety','safe_barbell_training'), ('fitness-v2','strength','support','supported_training'),
 ('fitness-v2','general_fitness','resistance','strength_training'), ('fitness-v2','general_fitness','resistance','resistance_training'), ('fitness-v2','general_fitness','conditioning','conditioning'), ('fitness-v2','general_fitness','mobility','mobility'), ('fitness-v2','general_fitness','support','supported_training'),
 ('fitness-v2','weight_loss','conditioning','conditioning'), ('fitness-v2','weight_loss','resistance','strength_training'), ('fitness-v2','weight_loss','resistance','resistance_training'), ('fitness-v2','weight_loss','mobility','mobility'),
 ('fitness-v2','mobility','mobility','mobility'), ('fitness-v2','mobility','resistance','resistance_training'), ('fitness-v2','mobility','conditioning','conditioning');

INSERT INTO recommendation.category_policies (policy_version,category_id,support_status)
SELECT 'fitness-v2', id, 'supported' FROM catalog.categories
WHERE slug IN ('adjustable-dumbbells','benches','power-racks','barbells','weight-plates','kettlebells','resistance-bands','cardio-machines');

INSERT INTO recommendation.policy_redundancy_groups VALUES
 ('fitness-v2','dumbbell_system','Dumbbell system'), ('fitness-v2','bench','Bench'),
 ('fitness-v2','rack','Rack'), ('fitness-v2','barbell','Barbell'),
 ('fitness-v2','weight_plates','Weight plates'), ('fitness-v2','kettlebell_system','Kettlebell system'),
 ('fitness-v2','resistance_band_system','Resistance band system'), ('fitness-v2','cardio_machine','Cardio machine');

INSERT INTO recommendation.category_redundancy_groups
SELECT 'fitness-v2', id, CASE slug
 WHEN 'adjustable-dumbbells' THEN 'dumbbell_system' WHEN 'benches' THEN 'bench'
 WHEN 'power-racks' THEN 'rack' WHEN 'barbells' THEN 'barbell'
 WHEN 'weight-plates' THEN 'weight_plates' WHEN 'kettlebells' THEN 'kettlebell_system'
 WHEN 'resistance-bands' THEN 'resistance_band_system' ELSE 'cardio_machine' END
FROM catalog.categories WHERE slug IN ('adjustable-dumbbells','benches','power-racks','barbells','weight-plates','kettlebells','resistance-bands','cardio-machines');

WITH mappings(slug, capability) AS (VALUES
 ('adjustable-dumbbells','resistance_training'),('adjustable-dumbbells','strength_training'),('adjustable-dumbbells','hypertrophy'),
 ('benches','supported_training'),('power-racks','safe_barbell_training'),('power-racks','pull_up'),('power-racks','anchor_point'),
 ('barbells','barbell_training'),('barbells','strength_training'),('barbells','hypertrophy'),('weight-plates','weight_plates'),
 ('kettlebells','resistance_training'),('kettlebells','strength_training'),('kettlebells','hypertrophy'),('kettlebells','conditioning'),
 ('resistance-bands','resistance_training'),('resistance-bands','hypertrophy'),('resistance-bands','conditioning'),('resistance-bands','mobility'),
 ('cardio-machines','conditioning'))
INSERT INTO recommendation.category_policy_capabilities
SELECT 'fitness-v2', categories.id, mappings.capability FROM mappings
JOIN catalog.categories categories ON categories.slug=mappings.slug;

INSERT INTO recommendation.product_policies
SELECT 'fitness-v2', id, published_fact_revision_id, published_score_revision_id
FROM catalog.products WHERE slug LIKE 'demo-%' AND published_fact_revision_id IS NOT NULL AND published_score_revision_id IS NOT NULL;

INSERT INTO recommendation.product_space_profiles (policy_version,product_id,footprint_length_mm,footprint_width_mm,footprint_height_mm)
SELECT 'fitness-v2', id, length_mm, width_mm, height_mm FROM catalog.products
WHERE slug LIKE 'demo-%' AND published_fact_revision_id IS NOT NULL AND published_score_revision_id IS NOT NULL;

WITH mappings(slug, capability, relation_type) AS (VALUES
 ('barbells','weight_plates','requires'),('weight-plates','barbell_training','requires'),('power-racks','barbell_training','requires'),
 ('benches','resistance_training','compatible_with'),('benches','barbell_training','compatible_with'),
 ('resistance-bands','anchor_point','compatible_with'))
INSERT INTO recommendation.product_policy_capabilities
SELECT 'fitness-v2', products.id, mappings.capability, mappings.relation_type FROM mappings
JOIN catalog.categories categories ON categories.slug=mappings.slug
JOIN catalog.products products ON products.category_id=categories.id AND products.slug LIKE 'demo-%';

WITH goal_scores(category_slug,build_muscle,strength,general_fitness,weight_loss,mobility) AS (VALUES
 ('adjustable-dumbbells',100,88,92,76,50),('benches',92,84,76,52,54),('power-racks',86,100,62,42,35),
 ('barbells',94,100,70,50,35),('weight-plates',92,100,65,45,30),('kettlebells',84,86,96,92,65),
 ('resistance-bands',76,62,94,86,100),('cardio-machines',40,42,90,100,58)), expanded AS (
 SELECT category_slug, unnest(ARRAY['build_muscle','strength','general_fitness','weight_loss','mobility']) goal_key,
        unnest(ARRAY[build_muscle,strength,general_fitness,weight_loss,mobility]) score FROM goal_scores)
INSERT INTO recommendation.product_goal_support
SELECT 'fitness-v2', products.id, expanded.goal_key, expanded.score FROM expanded
JOIN catalog.categories categories ON categories.slug=expanded.category_slug
JOIN catalog.products products ON products.category_id=categories.id AND products.slug LIKE 'demo-%';

WITH tags(slug, preference) AS (VALUES
 ('adjustable-dumbbells','dumbbells'),('benches','barbell'),('power-racks','barbell'),('power-racks','bodyweight'),
 ('barbells','barbell'),('weight-plates','barbell'),('kettlebells','kettlebell'),
 ('resistance-bands','resistance_bands'),('resistance-bands','bodyweight'),('cardio-machines','cardio'))
INSERT INTO recommendation.product_preference_tags
SELECT 'fitness-v2', products.id, tags.preference FROM tags
JOIN catalog.categories categories ON categories.slug=tags.slug
JOIN catalog.products products ON products.category_id=categories.id AND products.slug LIKE 'demo-%';

INSERT INTO recommendation.product_preference_tags
SELECT 'fitness-v2', id, 'low_impact' FROM catalog.products
WHERE slug LIKE 'demo-%' AND noise_score >= 85;

UPDATE recommendation.policy_versions SET workflow_status='active', activated_at=now()
WHERE version='fitness-v2'
  AND EXISTS (SELECT 1 FROM recommendation.category_policies WHERE policy_version='fitness-v2' AND support_status='supported')
  AND EXISTS (SELECT 1 FROM recommendation.product_policies WHERE policy_version='fitness-v2');

COMMENT ON TABLE recommendation.category_policies IS
 'Explicit support gate: catalog categories are never recommendation eligible merely because they exist.';
COMMENT ON TABLE recommendation.product_space_profiles IS
 'Evidence-revision-bound spatial policy; NULL optional measurements mean unknown, never zero.';
