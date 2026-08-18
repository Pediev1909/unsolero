-- Stable, bounded account collection pages introduced in Phase 10.
CREATE INDEX wishlists_user_created_stable_idx
    ON planning.wishlists (user_id, created_at DESC, product_id DESC);

CREATE INDEX setups_user_updated_stable_idx
    ON planning.setups (user_id, updated_at DESC, id DESC);

CREATE TABLE admin.media_deletion_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id uuid REFERENCES catalog.products(id) ON DELETE SET NULL,
    object_name text NOT NULL UNIQUE
        CHECK (char_length(object_name) BETWEEN 1 AND 240 AND object_name !~ '[\\/]'),
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'processing', 'completed', 'dead')),
    attempt_count smallint NOT NULL DEFAULT 0 CHECK (attempt_count BETWEEN 0 AND 10),
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    last_error_code text CHECK (last_error_code IS NULL OR last_error_code ~ '^[a-z][a-z0-9_.-]{0,63}$'),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz
);

CREATE INDEX media_deletion_jobs_pending_idx
    ON admin.media_deletion_jobs (next_attempt_at, id)
    WHERE status = 'pending';
