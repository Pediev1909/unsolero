CREATE TABLE identity.users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text,
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'active', 'suspended', 'deleted')),
    email_verified_at timestamptz,
    last_login_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CHECK (email = lower(btrim(email))),
    CHECK (position('@' IN email) > 1),
    CHECK ((status = 'deleted') = (deleted_at IS NOT NULL))
);

CREATE UNIQUE INDEX users_email_unique ON identity.users (lower(email));

CREATE TABLE planning.profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES identity.users(id) ON DELETE CASCADE,
    display_name text NOT NULL CHECK (char_length(btrim(display_name)) BETWEEN 1 AND 100),
    experience_level text NOT NULL
        CHECK (experience_level IN ('beginner', 'intermediate', 'advanced')),
    primary_goal text NOT NULL
        CHECK (primary_goal IN ('build_muscle', 'strength', 'general_fitness', 'weight_loss', 'mobility')),
    budget_minor bigint CHECK (budget_minor IS NULL OR budget_minor >= 0),
    currency character(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    space_length_mm integer CHECK (space_length_mm IS NULL OR space_length_mm > 0),
    space_width_mm integer CHECK (space_width_mm IS NULL OR space_width_mm > 0),
    space_height_mm integer CHECK (space_height_mm IS NULL OR space_height_mm > 0),
    apartment_living boolean NOT NULL DEFAULT false,
    noise_tolerance_score smallint CHECK (noise_tolerance_score BETWEEN 0 AND 100),
    country_code character(2) CHECK (country_code IS NULL OR country_code ~ '^[A-Z]{2}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((budget_minor IS NULL) = (currency IS NULL))
);

CREATE TABLE catalog.categories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id uuid REFERENCES catalog.categories(id) ON DELETE SET NULL,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text NOT NULL DEFAULT '',
    sort_order integer NOT NULL DEFAULT 0,
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (parent_id IS NULL OR parent_id <> id)
);

CREATE INDEX categories_parent_id_idx ON catalog.categories (parent_id);

CREATE TABLE catalog.brands (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text NOT NULL DEFAULT '',
    website_url text CHECK (website_url IS NULL OR website_url ~ '^https://'),
    country_code character(2) CHECK (country_code IS NULL OR country_code ~ '^[A-Z]{2}$'),
    is_active boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE catalog.products (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id uuid NOT NULL REFERENCES catalog.categories(id) ON DELETE RESTRICT,
    brand_id uuid NOT NULL REFERENCES catalog.brands(id) ON DELETE RESTRICT,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 180),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text NOT NULL CHECK (char_length(btrim(description)) > 0),
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    length_mm integer NOT NULL CHECK (length_mm > 0),
    width_mm integer NOT NULL CHECK (width_mm > 0),
    height_mm integer NOT NULL CHECK (height_mm > 0),
    weight_grams integer NOT NULL CHECK (weight_grams > 0),
    max_capacity_grams integer CHECK (max_capacity_grams IS NULL OR max_capacity_grams > 0),
    material text NOT NULL CHECK (char_length(btrim(material)) BETWEEN 1 AND 160),
    warranty_months smallint NOT NULL DEFAULT 0 CHECK (warranty_months >= 0),
    quality_score smallint NOT NULL CHECK (quality_score BETWEEN 0 AND 100),
    value_score smallint NOT NULL CHECK (value_score BETWEEN 0 AND 100),
    durability_score smallint NOT NULL CHECK (durability_score BETWEEN 0 AND 100),
    beginner_score smallint NOT NULL CHECK (beginner_score BETWEEN 0 AND 100),
    advanced_score smallint NOT NULL CHECK (advanced_score BETWEEN 0 AND 100),
    apartment_score smallint NOT NULL CHECK (apartment_score BETWEEN 0 AND 100),
    noise_score smallint NOT NULL CHECK (noise_score BETWEEN 0 AND 100),
    portability_score smallint NOT NULL CHECK (portability_score BETWEEN 0 AND 100),
    status text NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'published', 'discontinued')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON COLUMN catalog.products.price_minor IS 'Reference product price in the smallest currency unit.';
COMMENT ON COLUMN catalog.products.noise_score IS 'Quietness score: 100 is most apartment-friendly and quiet.';
COMMENT ON COLUMN catalog.products.max_capacity_grams IS 'Maximum supported training load where applicable.';

CREATE INDEX products_category_status_idx ON catalog.products (category_id, status);
CREATE INDEX products_brand_status_idx ON catalog.products (brand_id, status);
CREATE INDEX products_recommendation_scores_idx ON catalog.products (
    apartment_score DESC,
    beginner_score DESC,
    value_score DESC
) WHERE status = 'published';

CREATE TABLE catalog.product_images (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    url text NOT NULL CHECK (url ~ '^https://'),
    alt_text text NOT NULL CHECK (char_length(btrim(alt_text)) BETWEEN 1 AND 240),
    sort_order integer NOT NULL DEFAULT 0 CHECK (sort_order >= 0),
    is_primary boolean NOT NULL DEFAULT false,
    width_px integer CHECK (width_px IS NULL OR width_px > 0),
    height_px integer CHECK (height_px IS NULL OR height_px > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (product_id, url)
);

CREATE UNIQUE INDEX product_images_one_primary_idx
    ON catalog.product_images (product_id)
    WHERE is_primary;

CREATE TABLE catalog.product_attributes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    attribute_key text NOT NULL CHECK (attribute_key ~ '^[a-z][a-z0-9_]*$'),
    attribute_type text NOT NULL CHECK (attribute_type IN ('number', 'text', 'boolean')),
    numeric_value numeric(14, 4),
    text_value text,
    boolean_value boolean,
    unit text CHECK (unit IS NULL OR unit ~ '^[a-zA-Z0-9/%._-]{1,24}$'),
    is_filterable boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (product_id, attribute_key),
    CHECK (
        (attribute_type = 'number' AND numeric_value IS NOT NULL AND text_value IS NULL AND boolean_value IS NULL)
        OR (attribute_type = 'text' AND numeric_value IS NULL AND text_value IS NOT NULL AND boolean_value IS NULL AND unit IS NULL)
        OR (attribute_type = 'boolean' AND numeric_value IS NULL AND text_value IS NULL AND boolean_value IS NOT NULL AND unit IS NULL)
    )
);

CREATE INDEX product_attributes_key_number_idx
    ON catalog.product_attributes (attribute_key, numeric_value)
    WHERE attribute_type = 'number' AND is_filterable;
CREATE INDEX product_attributes_key_text_idx
    ON catalog.product_attributes (attribute_key, text_value)
    WHERE attribute_type = 'text' AND is_filterable;

CREATE TABLE commerce.merchants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(?:-[a-z0-9]+)*$'),
    website_url text NOT NULL CHECK (website_url ~ '^https://'),
    country_code character(2) NOT NULL CHECK (country_code ~ '^[A-Z]{2}$'),
    trust_score smallint NOT NULL DEFAULT 50 CHECK (trust_score BETWEEN 0 AND 100),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'disabled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE commerce.merchant_offers (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id uuid NOT NULL REFERENCES commerce.merchants(id) ON DELETE RESTRICT,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE RESTRICT,
    merchant_sku text NOT NULL CHECK (char_length(btrim(merchant_sku)) BETWEEN 1 AND 120),
    product_url text NOT NULL CHECK (product_url ~ '^https://'),
    price_minor bigint NOT NULL CHECK (price_minor >= 0),
    shipping_minor bigint NOT NULL DEFAULT 0 CHECK (shipping_minor >= 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    availability text NOT NULL
        CHECK (availability IN ('in_stock', 'backorder', 'out_of_stock', 'discontinued')),
    condition text NOT NULL DEFAULT 'new' CHECK (condition IN ('new', 'refurbished', 'used')),
    last_checked_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (merchant_id, merchant_sku)
);

CREATE INDEX merchant_offers_product_availability_idx
    ON commerce.merchant_offers (product_id, availability, currency, price_minor);

CREATE TABLE commerce.affiliate_links (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_offer_id uuid NOT NULL REFERENCES commerce.merchant_offers(id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (char_length(btrim(provider)) BETWEEN 1 AND 100),
    destination_url text NOT NULL CHECK (destination_url ~ '^https://'),
    external_reference text,
    disclosure_label text NOT NULL DEFAULT 'Affiliate link',
    is_active boolean NOT NULL DEFAULT true,
    valid_from timestamptz,
    valid_until timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (merchant_offer_id, provider),
    CHECK (valid_until IS NULL OR valid_from IS NULL OR valid_until > valid_from)
);

CREATE INDEX affiliate_links_active_offer_idx
    ON commerce.affiliate_links (merchant_offer_id)
    WHERE is_active;

CREATE TABLE recommendation.recommendation_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    profile_id uuid REFERENCES planning.profiles(id) ON DELETE SET NULL,
    anonymous_id text,
    status text NOT NULL DEFAULT 'started'
        CHECK (status IN ('started', 'processing', 'completed', 'failed', 'expired')),
    primary_goal text NOT NULL
        CHECK (primary_goal IN ('build_muscle', 'strength', 'general_fitness', 'weight_loss', 'mobility')),
    experience_level text NOT NULL
        CHECK (experience_level IN ('beginner', 'intermediate', 'advanced')),
    budget_minor bigint NOT NULL CHECK (budget_minor > 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    space_length_mm integer CHECK (space_length_mm IS NULL OR space_length_mm > 0),
    space_width_mm integer CHECK (space_width_mm IS NULL OR space_width_mm > 0),
    space_height_mm integer CHECK (space_height_mm IS NULL OR space_height_mm > 0),
    apartment_living boolean NOT NULL DEFAULT false,
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    expires_at timestamptz,
    CHECK (completed_at IS NULL OR completed_at >= started_at),
    CHECK (expires_at IS NULL OR expires_at > started_at)
);

CREATE INDEX recommendation_sessions_user_started_idx
    ON recommendation.recommendation_sessions (user_id, started_at DESC)
    WHERE user_id IS NOT NULL;
CREATE INDEX recommendation_sessions_anonymous_idx
    ON recommendation.recommendation_sessions (anonymous_id)
    WHERE anonymous_id IS NOT NULL;

CREATE TABLE recommendation.recommendations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES recommendation.recommendation_sessions(id) ON DELETE CASCADE,
    policy_version text NOT NULL,
    engine_version text NOT NULL,
    status text NOT NULL DEFAULT 'complete' CHECK (status IN ('complete', 'superseded')),
    objective_score smallint NOT NULL CHECK (objective_score BETWEEN 0 AND 100),
    total_price_minor bigint NOT NULL CHECK (total_price_minor >= 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    result_fingerprint text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (session_id, result_fingerprint)
);

CREATE INDEX recommendations_session_created_idx
    ON recommendation.recommendations (session_id, created_at DESC);

CREATE TABLE recommendation.recommendation_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    recommendation_id uuid NOT NULL REFERENCES recommendation.recommendations(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE RESTRICT,
    item_type text NOT NULL
        CHECK (item_type IN ('selected', 'cheaper_alternative', 'premium_alternative', 'future_upgrade', 'rejected')),
    rank smallint NOT NULL CHECK (rank > 0),
    quantity smallint NOT NULL DEFAULT 1 CHECK (quantity > 0),
    unit_price_minor bigint NOT NULL CHECK (unit_price_minor >= 0),
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    objective_score smallint NOT NULL CHECK (objective_score BETWEEN 0 AND 100),
    reason_code text NOT NULL CHECK (reason_code ~ '^[a-z][a-z0-9_.-]*$'),
    reason_summary text NOT NULL CHECK (char_length(btrim(reason_summary)) > 0),
    rejection_code text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (recommendation_id, item_type, product_id),
    UNIQUE (recommendation_id, item_type, rank),
    CHECK (
        (item_type = 'rejected' AND rejection_code IS NOT NULL)
        OR (item_type <> 'rejected' AND rejection_code IS NULL)
    )
);

CREATE INDEX recommendation_items_product_idx
    ON recommendation.recommendation_items (product_id);

CREATE TABLE planning.setups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    source_recommendation_id uuid REFERENCES recommendation.recommendations(id) ON DELETE SET NULL,
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 1 AND 120),
    description text NOT NULL DEFAULT '',
    currency character(3) NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE INDEX setups_user_updated_idx ON planning.setups (user_id, updated_at DESC);

CREATE TABLE planning.setup_items (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    setup_id uuid NOT NULL REFERENCES planning.setups(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE RESTRICT,
    merchant_offer_id uuid REFERENCES commerce.merchant_offers(id) ON DELETE SET NULL,
    quantity smallint NOT NULL DEFAULT 1 CHECK (quantity > 0),
    purchase_status text NOT NULL DEFAULT 'planned'
        CHECK (purchase_status IN ('planned', 'owned', 'replaced', 'removed')),
    paid_price_minor bigint CHECK (paid_price_minor IS NULL OR paid_price_minor >= 0),
    currency character(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    added_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (setup_id, product_id),
    CHECK ((paid_price_minor IS NULL) = (currency IS NULL))
);

CREATE TABLE planning.wishlists (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    priority smallint NOT NULL DEFAULT 0 CHECK (priority BETWEEN 0 AND 5),
    note text NOT NULL DEFAULT '' CHECK (char_length(note) <= 1000),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, product_id)
);

CREATE INDEX wishlists_user_priority_idx
    ON planning.wishlists (user_id, priority DESC, created_at DESC);

CREATE TABLE analytics.events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_name text NOT NULL CHECK (event_name ~ '^[a-z][a-z0-9_.-]*$'),
    schema_version smallint NOT NULL DEFAULT 1 CHECK (schema_version > 0),
    user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    recommendation_session_id uuid
        REFERENCES recommendation.recommendation_sessions(id) ON DELETE SET NULL,
    anonymous_id text,
    session_id text,
    request_id text,
    surface text NOT NULL CHECK (char_length(btrim(surface)) BETWEEN 1 AND 100),
    properties jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(properties) = 'object'),
    consent_state text NOT NULL DEFAULT 'essential'
        CHECK (consent_state IN ('essential', 'granted', 'denied', 'unknown')),
    occurred_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX events_name_occurred_idx ON analytics.events (event_name, occurred_at DESC);
CREATE INDEX events_user_occurred_idx
    ON analytics.events (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;
CREATE INDEX events_anonymous_occurred_idx
    ON analytics.events (anonymous_id, occurred_at DESC)
    WHERE anonymous_id IS NOT NULL;
