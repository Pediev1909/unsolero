CREATE SCHEMA editorial;

CREATE TABLE editorial.authors (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (char_length(btrim(name)) BETWEEN 2 AND 120),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    bio text NOT NULL CHECK (char_length(btrim(bio)) BETWEEN 20 AND 600),
    avatar_url text CHECK (
        avatar_url IS NULL OR avatar_url ~ '^https://' OR avatar_url ~ '^/images/[a-zA-Z0-9._/-]+$'
    ),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE editorial.entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id uuid NOT NULL REFERENCES editorial.authors(id) ON DELETE RESTRICT,
    content_type text NOT NULL CHECK (
        content_type IN ('article', 'guide', 'buying_guide', 'comparison')
    ),
    status text NOT NULL DEFAULT 'draft' CHECK (
        status IN ('draft', 'published', 'archived')
    ),
    title text NOT NULL CHECK (char_length(btrim(title)) BETWEEN 10 AND 180),
    slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    description text NOT NULL CHECK (char_length(btrim(description)) BETWEEN 40 AND 500),
    hero_image_url text NOT NULL CHECK (
        hero_image_url ~ '^https://' OR hero_image_url ~ '^/images/[a-zA-Z0-9._/-]+$'
    ),
    hero_image_alt text NOT NULL CHECK (
        char_length(btrim(hero_image_alt)) BETWEEN 5 AND 240
    ),
    content jsonb NOT NULL CHECK (
        jsonb_typeof(content) = 'array'
        AND jsonb_array_length(content) BETWEEN 1 AND 100
    ),
    seo_title text NOT NULL CHECK (char_length(btrim(seo_title)) BETWEEN 10 AND 70),
    seo_description text NOT NULL CHECK (
        char_length(btrim(seo_description)) BETWEEN 40 AND 180
    ),
    canonical_url text CHECK (canonical_url IS NULL OR canonical_url ~ '^https://'),
    published_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (status <> 'published' OR published_at IS NOT NULL)
);

CREATE INDEX entries_published_type_date_idx
    ON editorial.entries (content_type, published_at DESC, id)
    WHERE status = 'published';
CREATE INDEX entries_published_updated_idx
    ON editorial.entries (updated_at DESC, id)
    WHERE status = 'published';

CREATE TABLE editorial.entry_products (
    entry_id uuid NOT NULL REFERENCES editorial.entries(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE RESTRICT,
    position smallint NOT NULL CHECK (position BETWEEN 0 AND 100),
    PRIMARY KEY (entry_id, product_id),
    UNIQUE (entry_id, position)
);

CREATE INDEX entry_products_product_idx
    ON editorial.entry_products (product_id, entry_id);

CREATE TABLE editorial.entry_categories (
    entry_id uuid NOT NULL REFERENCES editorial.entries(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES catalog.categories(id) ON DELETE RESTRICT,
    position smallint NOT NULL CHECK (position BETWEEN 0 AND 100),
    PRIMARY KEY (entry_id, category_id),
    UNIQUE (entry_id, position)
);

CREATE INDEX entry_categories_category_idx
    ON editorial.entry_categories (category_id, entry_id);

CREATE TABLE editorial.related_entries (
    entry_id uuid NOT NULL REFERENCES editorial.entries(id) ON DELETE CASCADE,
    related_entry_id uuid NOT NULL REFERENCES editorial.entries(id) ON DELETE CASCADE,
    position smallint NOT NULL CHECK (position BETWEEN 0 AND 100),
    PRIMARY KEY (entry_id, related_entry_id),
    UNIQUE (entry_id, position),
    CHECK (entry_id <> related_entry_id)
);

CREATE INDEX related_entries_reverse_idx
    ON editorial.related_entries (related_entry_id, entry_id);

COMMENT ON COLUMN editorial.entries.content IS
    'Validated editorial blocks. Arbitrary HTML is not accepted or rendered.';
COMMENT ON COLUMN editorial.entries.canonical_url IS
    'Optional explicit HTTPS canonical. When null, the public site URL and route are used.';
