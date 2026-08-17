CREATE TABLE recommendation.drafts (
    user_id uuid PRIMARY KEY REFERENCES identity.users(id) ON DELETE CASCADE,
    current_step smallint NOT NULL DEFAULT 1 CHECK (current_step BETWEEN 1 AND 8),
    primary_goal text CHECK (
        primary_goal IS NULL OR primary_goal IN (
            'build_muscle', 'strength', 'general_fitness', 'weight_loss', 'mobility'
        )
    ),
    experience_level text CHECK (
        experience_level IS NULL OR experience_level IN ('beginner', 'intermediate', 'advanced')
    ),
    budget_minor bigint CHECK (budget_minor IS NULL OR budget_minor > 0),
    currency character(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    space_length_mm integer CHECK (space_length_mm IS NULL OR space_length_mm > 0),
    space_width_mm integer CHECK (space_width_mm IS NULL OR space_width_mm > 0),
    space_height_mm integer CHECK (space_height_mm IS NULL OR space_height_mm > 0),
    apartment_living boolean,
    training_preferences text[] NOT NULL DEFAULT '{}',
    priorities text[] NOT NULL DEFAULT '{}',
    free_text text NOT NULL DEFAULT '' CHECK (char_length(free_text) <= 1000),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((budget_minor IS NULL) = (currency IS NULL)),
    CHECK (
        (space_length_mm IS NULL AND space_width_mm IS NULL AND space_height_mm IS NULL AND apartment_living IS NULL)
        OR
        (space_length_mm IS NOT NULL AND space_width_mm IS NOT NULL AND space_height_mm IS NOT NULL AND apartment_living IS NOT NULL)
    )
);

CREATE TABLE recommendation.draft_existing_equipment (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES recommendation.drafts(user_id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    category_slug text NOT NULL CHECK (category_slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    sort_order smallint NOT NULL CHECK (sort_order >= 0),
    UNIQUE (user_id, sort_order)
);

ALTER TABLE recommendation.recommendation_sessions
    ADD COLUMN training_preferences text[] NOT NULL DEFAULT '{}',
    ADD COLUMN priorities text[] NOT NULL DEFAULT '{}',
    ADD COLUMN free_text text NOT NULL DEFAULT '' CHECK (char_length(free_text) <= 1000);

CREATE TABLE recommendation.session_existing_equipment (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL
        REFERENCES recommendation.recommendation_sessions(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    category_slug text NOT NULL CHECK (category_slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    sort_order smallint NOT NULL CHECK (sort_order >= 0),
    UNIQUE (session_id, sort_order)
);

ALTER TABLE recommendation.recommendations
    ADD COLUMN goal_match_score smallint NOT NULL DEFAULT 0 CHECK (goal_match_score BETWEEN 0 AND 100),
    ADD COLUMN budget_match_score smallint NOT NULL DEFAULT 0 CHECK (budget_match_score BETWEEN 0 AND 100),
    ADD COLUMN space_match_score smallint NOT NULL DEFAULT 0 CHECK (space_match_score BETWEEN 0 AND 100),
    ADD COLUMN experience_match_score smallint NOT NULL DEFAULT 0 CHECK (experience_match_score BETWEEN 0 AND 100),
    ADD COLUMN preference_match_score smallint NOT NULL DEFAULT 0 CHECK (preference_match_score BETWEEN 0 AND 100),
    ADD COLUMN quality_score smallint NOT NULL DEFAULT 0 CHECK (quality_score BETWEEN 0 AND 100),
    ADD COLUMN value_score smallint NOT NULL DEFAULT 0 CHECK (value_score BETWEEN 0 AND 100),
    ADD COLUMN durability_score smallint NOT NULL DEFAULT 0 CHECK (durability_score BETWEEN 0 AND 100),
    ADD COLUMN compatibility_score smallint NOT NULL DEFAULT 0 CHECK (compatibility_score BETWEEN 0 AND 100),
    ADD COLUMN portability_score smallint NOT NULL DEFAULT 0 CHECK (portability_score BETWEEN 0 AND 100),
    ADD COLUMN noise_score smallint NOT NULL DEFAULT 0 CHECK (noise_score BETWEEN 0 AND 100);

ALTER TABLE recommendation.recommendation_items
    ADD COLUMN alternative_for_product_id uuid REFERENCES catalog.products(id) ON DELETE RESTRICT,
    ADD COLUMN goal_match_score smallint NOT NULL DEFAULT 0 CHECK (goal_match_score BETWEEN 0 AND 100),
    ADD COLUMN budget_match_score smallint NOT NULL DEFAULT 0 CHECK (budget_match_score BETWEEN 0 AND 100),
    ADD COLUMN space_match_score smallint NOT NULL DEFAULT 0 CHECK (space_match_score BETWEEN 0 AND 100),
    ADD COLUMN experience_match_score smallint NOT NULL DEFAULT 0 CHECK (experience_match_score BETWEEN 0 AND 100),
    ADD COLUMN preference_match_score smallint NOT NULL DEFAULT 0 CHECK (preference_match_score BETWEEN 0 AND 100),
    ADD COLUMN quality_score smallint NOT NULL DEFAULT 0 CHECK (quality_score BETWEEN 0 AND 100),
    ADD COLUMN value_score smallint NOT NULL DEFAULT 0 CHECK (value_score BETWEEN 0 AND 100),
    ADD COLUMN durability_score smallint NOT NULL DEFAULT 0 CHECK (durability_score BETWEEN 0 AND 100),
    ADD COLUMN compatibility_score smallint NOT NULL DEFAULT 0 CHECK (compatibility_score BETWEEN 0 AND 100),
    ADD COLUMN portability_score smallint NOT NULL DEFAULT 0 CHECK (portability_score BETWEEN 0 AND 100),
    ADD COLUMN noise_score smallint NOT NULL DEFAULT 0 CHECK (noise_score BETWEEN 0 AND 100),
    ADD CONSTRAINT recommendation_items_alternative_target_check CHECK (
        (item_type IN ('cheaper_alternative', 'premium_alternative') AND alternative_for_product_id IS NOT NULL)
        OR
        (item_type NOT IN ('cheaper_alternative', 'premium_alternative') AND alternative_for_product_id IS NULL)
    );

CREATE TABLE recommendation.item_reasons (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recommendation_item_id uuid NOT NULL
        REFERENCES recommendation.recommendation_items(id) ON DELETE CASCADE,
    sort_order smallint NOT NULL CHECK (sort_order >= 0),
    code text NOT NULL CHECK (code ~ '^[a-z][a-z0-9_.-]*$'),
    message text NOT NULL CHECK (char_length(btrim(message)) > 0),
    dimension text NOT NULL CHECK (dimension ~ '^[a-z][a-z0-9_]*$'),
    score smallint NOT NULL CHECK (score BETWEEN 0 AND 100),
    UNIQUE (recommendation_item_id, sort_order)
);

CREATE INDEX item_reasons_item_idx
    ON recommendation.item_reasons (recommendation_item_id, sort_order);
