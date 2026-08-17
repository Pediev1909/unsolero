CREATE SCHEMA admin;

CREATE TABLE identity.roles (
    role_key text PRIMARY KEY CHECK (role_key ~ '^[a-z][a-z0-9_]{1,49}$'),
    description text NOT NULL CHECK (char_length(btrim(description)) > 0),
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO identity.roles (role_key, description) VALUES
    ('admin', 'Full access to protected administration workflows');

CREATE TABLE identity.user_roles (
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    role_key text NOT NULL REFERENCES identity.roles(role_key) ON DELETE RESTRICT,
    granted_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_key)
);

CREATE INDEX user_roles_role_user_idx ON identity.user_roles (role_key, user_id);

CREATE TABLE admin.audit_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id uuid REFERENCES identity.users(id) ON DELETE SET NULL,
    action text NOT NULL CHECK (action ~ '^[a-z][a-z0-9_.-]{1,99}$'),
    entity_type text NOT NULL CHECK (entity_type ~ '^[a-z][a-z0-9_.-]{1,99}$'),
    entity_id text NOT NULL CHECK (char_length(btrim(entity_id)) BETWEEN 1 AND 160),
    changes jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(changes) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_actor_occurred_idx
    ON admin.audit_log (actor_user_id, occurred_at DESC)
    WHERE actor_user_id IS NOT NULL;
CREATE INDEX audit_log_entity_occurred_idx
    ON admin.audit_log (entity_type, entity_id, occurred_at DESC);

ALTER TABLE catalog.product_images DROP CONSTRAINT product_images_url_check;
ALTER TABLE catalog.product_images
    ADD CONSTRAINT product_images_url_check
    CHECK (url ~ '^https://' OR url ~ '^/images/[a-zA-Z0-9._/-]+$'
        OR url ~ '^/api/media/products/[a-zA-Z0-9._-]+$');
