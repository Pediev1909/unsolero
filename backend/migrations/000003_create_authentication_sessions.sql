CREATE TABLE identity.sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    expires_at timestamptz NOT NULL,
    idle_expires_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CHECK (expires_at > created_at),
    CHECK (idle_expires_at > created_at AND idle_expires_at <= expires_at),
    CHECK (last_seen_at >= created_at),
    CHECK (revoked_at IS NULL OR revoked_at >= created_at)
);

CREATE INDEX sessions_user_active_idx
    ON identity.sessions (user_id, expires_at DESC)
    WHERE revoked_at IS NULL;

CREATE INDEX sessions_expiration_idx
    ON identity.sessions (expires_at)
    WHERE revoked_at IS NULL;
