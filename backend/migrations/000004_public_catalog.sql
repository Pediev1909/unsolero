ALTER TABLE catalog.product_images
    DROP CONSTRAINT product_images_url_check;

ALTER TABLE catalog.product_images
    ADD CONSTRAINT product_images_url_check
    CHECK (url ~ '^https://' OR url ~ '^/images/[a-zA-Z0-9._/-]+$');

CREATE INDEX products_public_price_idx
    ON catalog.products (price_minor, name)
    WHERE status = 'published';

CREATE INDEX products_public_name_idx
    ON catalog.products (lower(name))
    WHERE status = 'published';

CREATE TABLE commerce.outbound_clicks (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    affiliate_link_id uuid NOT NULL
        REFERENCES commerce.affiliate_links(id) ON DELETE RESTRICT,
    merchant_offer_id uuid NOT NULL
        REFERENCES commerce.merchant_offers(id) ON DELETE RESTRICT,
    product_id uuid NOT NULL
        REFERENCES catalog.products(id) ON DELETE RESTRICT,
    surface text NOT NULL CHECK (surface ~ '^[a-z][a-z0-9_.-]*$'),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX outbound_clicks_affiliate_occurred_idx
    ON commerce.outbound_clicks (affiliate_link_id, occurred_at DESC);

CREATE INDEX outbound_clicks_product_occurred_idx
    ON commerce.outbound_clicks (product_id, occurred_at DESC);
