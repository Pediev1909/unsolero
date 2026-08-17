CREATE TABLE planning.comparison_items (
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    product_id uuid NOT NULL REFERENCES catalog.products(id) ON DELETE CASCADE,
    position smallint NOT NULL CHECK (position BETWEEN 0 AND 3),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, product_id),
    UNIQUE (user_id, position)
);

CREATE INDEX comparison_items_product_idx
    ON planning.comparison_items (product_id);
