CREATE SCHEMA IF NOT EXISTS evidence;

INSERT INTO identity.roles (role_key, description) VALUES
    ('evidence_editor', 'Create evidence sources and draft product fact and score revisions'),
    ('evidence_reviewer', 'Independently review, approve, and publish product evidence revisions');

CREATE TABLE evidence.sources (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    external_key text UNIQUE CHECK (external_key IS NULL OR external_key ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    source_type text NOT NULL CHECK (source_type IN (
        'manufacturer_documentation', 'independent_testing',
        'verified_merchant_data', 'editorial_assessment', 'demo_fixture'
    )),
    title text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 1 AND 240),
    publisher text NOT NULL CHECK (char_length(btrim(publisher)) BETWEEN 1 AND 180),
    source_url text,
    is_fictional boolean NOT NULL DEFAULT false,
    review_status text NOT NULL DEFAULT 'pending'
        CHECK (review_status IN ('pending', 'verified', 'rejected')),
    reviewer_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    reviewed_at timestamptz,
    review_note text NOT NULL DEFAULT '' CHECK (char_length(review_note) <= 2000),
    created_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (source_url IS NULL OR source_url ~ '^https://'),
    CHECK (source_type = 'demo_fixture' OR source_url IS NOT NULL),
    CHECK (NOT is_fictional OR source_type = 'demo_fixture'),
    CHECK ((review_status = 'pending') = (reviewed_at IS NULL))
);

CREATE TABLE evidence.observations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id uuid NOT NULL REFERENCES evidence.sources(id) ON DELETE RESTRICT,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    observed_at timestamptz NOT NULL,
    expires_at timestamptz,
    confidence smallint NOT NULL CHECK (confidence BETWEEN 0 AND 100),
    notes text NOT NULL DEFAULT '' CHECK (char_length(notes) <= 4000),
    created_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at IS NULL OR expires_at > observed_at)
);

CREATE INDEX evidence_observations_product_idx
    ON evidence.observations (product_id, observed_at DESC);

CREATE TABLE evidence.product_fact_revisions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    version integer NOT NULL CHECK (version > 0),
    category_id uuid REFERENCES catalog.categories(id) ON DELETE RESTRICT,
    brand_id uuid REFERENCES catalog.brands(id) ON DELETE RESTRICT,
    name text CHECK (name IS NULL OR char_length(btrim(name)) BETWEEN 1 AND 180),
    slug text CHECK (slug IS NULL OR slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text CHECK (description IS NULL OR char_length(btrim(description)) > 0),
    price_minor bigint CHECK (price_minor IS NULL OR price_minor >= 0),
    currency character(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    length_mm integer CHECK (length_mm IS NULL OR length_mm > 0),
    width_mm integer CHECK (width_mm IS NULL OR width_mm > 0),
    height_mm integer CHECK (height_mm IS NULL OR height_mm > 0),
    weight_grams integer CHECK (weight_grams IS NULL OR weight_grams > 0),
    max_capacity_grams integer CHECK (max_capacity_grams IS NULL OR max_capacity_grams > 0),
    material text CHECK (material IS NULL OR char_length(btrim(material)) BETWEEN 1 AND 160),
    warranty_months smallint CHECK (warranty_months IS NULL OR warranty_months >= 0),
    workflow_status text NOT NULL DEFAULT 'draft' CHECK (workflow_status IN (
        'draft', 'in_review', 'approved', 'published', 'rejected', 'superseded'
    )),
    created_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    submitted_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    submitted_at timestamptz,
    reviewed_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    reviewed_at timestamptz,
    published_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    published_at timestamptz,
    valid_until timestamptz,
    review_note text NOT NULL DEFAULT '' CHECK (char_length(review_note) <= 2000),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (product_id, version),
    CHECK ((price_minor IS NULL) = (currency IS NULL)),
    CHECK (submitted_by_user_id IS NULL OR submitted_at IS NOT NULL),
    CHECK (reviewed_by_user_id IS NULL OR reviewed_at IS NOT NULL),
    CHECK (published_by_user_id IS NULL OR published_at IS NOT NULL)
);

CREATE TABLE evidence.fact_provenance (
    fact_revision_id uuid NOT NULL REFERENCES evidence.product_fact_revisions(id) ON DELETE CASCADE,
    fact_key text NOT NULL CHECK (fact_key IN (
        'category', 'brand', 'name', 'slug', 'description', 'price', 'dimensions',
        'weight', 'max_capacity', 'material', 'warranty'
    )),
    observation_id uuid NOT NULL REFERENCES evidence.observations(id) ON DELETE CASCADE,
    public_classification text NOT NULL CHECK (public_classification IN (
        'verified_fact', 'manufacturer_claim', 'merchant_observation', 'editorial_assessment'
    )),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (fact_revision_id, fact_key, observation_id)
);

CREATE TABLE evidence.score_revisions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    fact_revision_id uuid NOT NULL REFERENCES evidence.product_fact_revisions(id) ON DELETE RESTRICT,
    version integer NOT NULL CHECK (version > 0),
    quality_score smallint CHECK (quality_score BETWEEN 0 AND 100),
    value_score smallint CHECK (value_score BETWEEN 0 AND 100),
    durability_score smallint CHECK (durability_score BETWEEN 0 AND 100),
    beginner_score smallint CHECK (beginner_score BETWEEN 0 AND 100),
    advanced_score smallint CHECK (advanced_score BETWEEN 0 AND 100),
    apartment_score smallint CHECK (apartment_score BETWEEN 0 AND 100),
    noise_score smallint CHECK (noise_score BETWEEN 0 AND 100),
    portability_score smallint CHECK (portability_score BETWEEN 0 AND 100),
    workflow_status text NOT NULL DEFAULT 'draft' CHECK (workflow_status IN (
        'draft', 'in_review', 'approved', 'published', 'rejected', 'superseded'
    )),
    created_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    submitted_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    submitted_at timestamptz,
    reviewed_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    reviewed_at timestamptz,
    published_by_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    published_at timestamptz,
    review_note text NOT NULL DEFAULT '' CHECK (char_length(review_note) <= 2000),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (product_id, version),
    CHECK (submitted_by_user_id IS NULL OR submitted_at IS NOT NULL),
    CHECK (reviewed_by_user_id IS NULL OR reviewed_at IS NOT NULL),
    CHECK (published_by_user_id IS NULL OR published_at IS NOT NULL)
);

CREATE TABLE evidence.score_rationales (
    score_revision_id uuid NOT NULL REFERENCES evidence.score_revisions(id) ON DELETE CASCADE,
    score_key text NOT NULL CHECK (score_key IN (
        'quality', 'value', 'durability', 'beginner',
        'advanced', 'apartment', 'noise', 'portability'
    )),
    rationale text NOT NULL CHECK (char_length(btrim(rationale)) BETWEEN 1 AND 2000),
    observation_id uuid NOT NULL REFERENCES evidence.observations(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (score_revision_id, score_key, observation_id)
);

ALTER TABLE catalog.products
    ADD COLUMN published_fact_revision_id uuid REFERENCES evidence.product_fact_revisions(id) ON DELETE SET NULL,
    ADD COLUMN published_score_revision_id uuid REFERENCES evidence.score_revisions(id) ON DELETE SET NULL;

CREATE UNIQUE INDEX products_published_fact_revision_idx
    ON catalog.products (published_fact_revision_id) WHERE published_fact_revision_id IS NOT NULL;
CREATE UNIQUE INDEX products_published_score_revision_idx
    ON catalog.products (published_score_revision_id) WHERE published_score_revision_id IS NOT NULL;

CREATE TABLE recommendation.policy_versions (
    version text PRIMARY KEY CHECK (version ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    goal_match_weight smallint NOT NULL CHECK (goal_match_weight >= 0),
    budget_match_weight smallint NOT NULL CHECK (budget_match_weight >= 0),
    space_match_weight smallint NOT NULL CHECK (space_match_weight >= 0),
    experience_match_weight smallint NOT NULL CHECK (experience_match_weight >= 0),
    preference_match_weight smallint NOT NULL CHECK (preference_match_weight >= 0),
    quality_weight smallint NOT NULL CHECK (quality_weight >= 0),
    value_weight smallint NOT NULL CHECK (value_weight >= 0),
    durability_weight smallint NOT NULL CHECK (durability_weight >= 0),
    compatibility_weight smallint NOT NULL CHECK (compatibility_weight >= 0),
    portability_weight smallint NOT NULL CHECK (portability_weight >= 0),
    noise_weight smallint NOT NULL CHECK (noise_weight >= 0),
    priority_boost_percent smallint NOT NULL CHECK (priority_boost_percent >= 100),
    maximum_setup_items smallint NOT NULL CHECK (maximum_setup_items > 0),
    candidates_per_slot smallint NOT NULL CHECK (candidates_per_slot > 0),
    optional_slot_bonus smallint NOT NULL CHECK (optional_slot_bonus >= 0),
    published_at timestamptz NOT NULL,
    retired_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (retired_at IS NULL OR retired_at > published_at)
);

INSERT INTO recommendation.policy_versions (
    version, goal_match_weight, budget_match_weight, space_match_weight,
    experience_match_weight, preference_match_weight, quality_weight, value_weight,
    durability_weight, compatibility_weight, portability_weight, noise_weight,
    priority_boost_percent, maximum_setup_items, candidates_per_slot,
    optional_slot_bonus, published_at
) VALUES ('home-gym-v1', 20, 12, 12, 10, 8, 8, 9, 7, 10, 2, 2, 150, 4, 12, 8, now());

ALTER TABLE recommendation.recommendations
    ADD CONSTRAINT recommendations_policy_version_fk
    FOREIGN KEY (policy_version) REFERENCES recommendation.policy_versions(version) ON DELETE RESTRICT;

CREATE TABLE recommendation.candidate_snapshots (
    recommendation_id uuid NOT NULL REFERENCES recommendation.recommendations(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE RESTRICT,
    fact_revision_id uuid REFERENCES evidence.product_fact_revisions(id) ON DELETE RESTRICT,
    score_revision_id uuid REFERENCES evidence.score_revisions(id) ON DELETE RESTRICT,
    name text NOT NULL,
    category_slug text NOT NULL,
    price_minor bigint NOT NULL,
    currency character(3) NOT NULL,
    length_mm integer NOT NULL,
    width_mm integer NOT NULL,
    height_mm integer NOT NULL,
    quality_score smallint NOT NULL,
    value_score smallint NOT NULL,
    durability_score smallint NOT NULL,
    beginner_score smallint NOT NULL,
    advanced_score smallint NOT NULL,
    apartment_score smallint NOT NULL,
    noise_score smallint NOT NULL,
    portability_score smallint NOT NULL,
    PRIMARY KEY (recommendation_id, product_id),
    CHECK (price_minor >= 0 AND currency ~ '^[A-Z]{3}$'),
    CHECK (length_mm > 0 AND width_mm > 0 AND height_mm > 0),
    CHECK (quality_score BETWEEN 0 AND 100),
    CHECK (value_score BETWEEN 0 AND 100),
    CHECK (durability_score BETWEEN 0 AND 100),
    CHECK (beginner_score BETWEEN 0 AND 100),
    CHECK (advanced_score BETWEEN 0 AND 100),
    CHECK (apartment_score BETWEEN 0 AND 100),
    CHECK (noise_score BETWEEN 0 AND 100),
    CHECK (portability_score BETWEEN 0 AND 100)
);

-- Existing published rows predate governance. Fail closed for non-demo data;
-- explicitly fictional demo rows are backfilled into versioned evidence.
UPDATE catalog.products SET status = 'draft'
WHERE status = 'published' AND slug NOT LIKE 'demo-%';

WITH created_sources AS (
    INSERT INTO evidence.sources (
        external_key, source_type, title, publisher, is_fictional, review_status,
        reviewed_at, review_note
    )
    SELECT 'unsolero-demo-fixture-v1', 'demo_fixture', 'Fictional UNSOLERO development fixture',
           'UNSOLERO development seed', true, 'verified', now(),
           'Fictional evidence for local development only; not a real-world source.'
    WHERE EXISTS (SELECT 1 FROM catalog.products WHERE slug LIKE 'demo-%')
    RETURNING id
), observations AS (
    INSERT INTO evidence.observations (
        source_id, product_id, observed_at, confidence, notes
    )
    SELECT created_sources.id, products.id, now(), 100,
           'Fictional demo observation; no real product claim is made.'
    FROM created_sources CROSS JOIN catalog.products AS products
    WHERE products.slug LIKE 'demo-%'
    RETURNING id, product_id
), fact_revisions AS (
    INSERT INTO evidence.product_fact_revisions (
        product_id, version, category_id, brand_id, name, slug, description,
        price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
        max_capacity_grams, material, warranty_months, workflow_status,
        submitted_at, reviewed_at, published_at, review_note
    )
    SELECT id, 1, category_id, brand_id, name, slug, description,
           price_minor, currency, length_mm, width_mm, height_mm, weight_grams,
           max_capacity_grams, material, warranty_months, 'published',
           now(), now(), now(), 'Automatically migrated fictional demo fixture.'
    FROM catalog.products WHERE slug LIKE 'demo-%'
    RETURNING id, product_id
), score_revisions AS (
    INSERT INTO evidence.score_revisions (
        product_id, fact_revision_id, version, quality_score, value_score,
        durability_score, beginner_score, advanced_score, apartment_score,
        noise_score, portability_score, workflow_status,
        submitted_at, reviewed_at, published_at, review_note
    )
    SELECT products.id, facts.id, 1, products.quality_score, products.value_score,
           products.durability_score, products.beginner_score, products.advanced_score,
           products.apartment_score, products.noise_score, products.portability_score,
           'published', now(), now(), now(),
           'Automatically migrated fictional demo score fixture.'
    FROM catalog.products AS products
    JOIN fact_revisions AS facts ON facts.product_id = products.id
    RETURNING id, product_id
), fact_links AS (
    INSERT INTO evidence.fact_provenance (
        fact_revision_id, fact_key, observation_id, public_classification
    )
    SELECT facts.id, keys.fact_key, observations.id, 'editorial_assessment'
    FROM fact_revisions AS facts
    JOIN observations ON observations.product_id = facts.product_id
    JOIN catalog.products AS products ON products.id = facts.product_id
    CROSS JOIN (VALUES
        ('category'), ('brand'), ('name'), ('slug'), ('description'), ('price'),
        ('dimensions'), ('weight'), ('max_capacity'), ('material'), ('warranty')
    ) AS keys(fact_key)
    WHERE keys.fact_key <> 'max_capacity' OR products.max_capacity_grams IS NOT NULL
), score_links AS (
    INSERT INTO evidence.score_rationales (
        score_revision_id, score_key, rationale, observation_id
    )
    SELECT scores.id, keys.score_key,
           'Fictional demo score supplied by the development fixture.', observations.id
    FROM score_revisions AS scores
    JOIN observations ON observations.product_id = scores.product_id
    CROSS JOIN (VALUES
        ('quality'), ('value'), ('durability'), ('beginner'),
        ('advanced'), ('apartment'), ('noise'), ('portability')
    ) AS keys(score_key)
)
UPDATE catalog.products AS products
SET published_fact_revision_id = facts.id,
    published_score_revision_id = scores.id
FROM fact_revisions AS facts
JOIN score_revisions AS scores ON scores.product_id = facts.product_id
WHERE products.id = facts.product_id;

CREATE INDEX fact_revisions_product_status_idx
    ON evidence.product_fact_revisions (product_id, workflow_status, version DESC);
CREATE INDEX score_revisions_product_status_idx
    ON evidence.score_revisions (product_id, workflow_status, version DESC);
CREATE INDEX evidence_sources_review_idx
    ON evidence.sources (review_status, source_type, updated_at DESC);

COMMENT ON SCHEMA evidence IS
    'Non-commercial provenance and governed revision history for recommendation-critical product data.';
COMMENT ON TABLE recommendation.candidate_snapshots IS
    'Immutable, commercial-free catalog inputs used by a completed deterministic recommendation.';
