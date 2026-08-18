-- Phase 5: account security primitives. Secret/token material is stored only
-- as hashes or authenticated ciphertext. Security events are append-only.

INSERT INTO identity.roles (role_key, description) VALUES
    ('catalog_editor', 'Creates and updates catalog drafts without bypassing evidence publication.'),
    ('commerce_operator', 'Operates merchant providers, offers, imports, and verified conversion workflows.'),
    ('content_editor', 'Creates and updates editorial content without granting administrative access.'),
    ('analyst', 'Reads operational and analytics reports without mutation privileges.')
ON CONFLICT (role_key) DO NOTHING;

ALTER TABLE identity.sessions
    ADD COLUMN mfa_authenticated_at timestamptz,
    ADD COLUMN authentication_method text NOT NULL DEFAULT 'password'
        CHECK (authentication_method IN ('password', 'password_mfa', 'password_recovery')),
    ADD CONSTRAINT sessions_mfa_time_check CHECK (
        mfa_authenticated_at IS NULL
        OR (mfa_authenticated_at >= created_at AND mfa_authenticated_at <= last_seen_at)
    );

CREATE TABLE identity.email_verification_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    consumed_at timestamptz,
    invalidated_at timestamptz,
    CHECK (expires_at > created_at),
    CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    CHECK (invalidated_at IS NULL OR invalidated_at >= created_at),
    CHECK (consumed_at IS NULL OR invalidated_at IS NULL)
);

CREATE INDEX email_verification_tokens_user_created_idx
    ON identity.email_verification_tokens (user_id, created_at DESC);
CREATE INDEX email_verification_tokens_expiration_idx
    ON identity.email_verification_tokens (expires_at)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE TABLE identity.password_reset_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    consumed_at timestamptz,
    invalidated_at timestamptz,
    CHECK (expires_at > created_at),
    CHECK (consumed_at IS NULL OR consumed_at >= created_at),
    CHECK (invalidated_at IS NULL OR invalidated_at >= created_at),
    CHECK (consumed_at IS NULL OR invalidated_at IS NULL)
);

CREATE INDEX password_reset_tokens_user_created_idx
    ON identity.password_reset_tokens (user_id, created_at DESC);
CREATE INDEX password_reset_tokens_expiration_idx
    ON identity.password_reset_tokens (expires_at)
    WHERE consumed_at IS NULL AND invalidated_at IS NULL;

CREATE TABLE identity.mfa_credentials (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES identity.users(id) ON DELETE CASCADE,
    secret_ciphertext bytea NOT NULL CHECK (octet_length(secret_ciphertext) >= 32),
    secret_nonce bytea NOT NULL CHECK (octet_length(secret_nonce) = 12),
    key_version smallint NOT NULL DEFAULT 1 CHECK (key_version > 0),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'enabled', 'disabled')),
    created_at timestamptz NOT NULL DEFAULT now(),
    verified_at timestamptz,
    disabled_at timestamptz,
    CHECK ((status = 'enabled') = (verified_at IS NOT NULL)),
    CHECK ((status = 'disabled') = (disabled_at IS NOT NULL))
);

CREATE TABLE identity.mfa_recovery_codes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    credential_id uuid NOT NULL REFERENCES identity.mfa_credentials(id) ON DELETE CASCADE,
    code_hash bytea NOT NULL UNIQUE CHECK (octet_length(code_hash) = 32),
    created_at timestamptz NOT NULL DEFAULT now(),
    consumed_at timestamptz
);

CREATE INDEX mfa_recovery_codes_credential_active_idx
    ON identity.mfa_recovery_codes (credential_id)
    WHERE consumed_at IS NULL;

CREATE TABLE identity.mfa_login_challenges (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token_hash bytea NOT NULL UNIQUE CHECK (octet_length(token_hash) = 32),
    expires_at timestamptz NOT NULL,
    attempts smallint NOT NULL DEFAULT 0 CHECK (attempts BETWEEN 0 AND 5),
    created_at timestamptz NOT NULL DEFAULT now(),
    consumed_at timestamptz,
    CHECK (expires_at > created_at),
    CHECK (consumed_at IS NULL OR consumed_at >= created_at)
);

CREATE INDEX mfa_login_challenges_expiration_idx
    ON identity.mfa_login_challenges (expires_at)
    WHERE consumed_at IS NULL;

CREATE TABLE identity.security_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid REFERENCES identity.users(id) ON DELETE RESTRICT,
    -- Deliberately not a foreign key: expired sessions are cleaned while this
    -- append-only audit identifier remains stable.
    session_id uuid,
    event_type text NOT NULL CHECK (event_type ~ '^[a-z][a-z0-9_.-]{1,99}$'),
    outcome text NOT NULL CHECK (outcome IN ('success', 'failure', 'requested', 'denied')),
    request_id text CHECK (request_id IS NULL OR request_id ~ '^[A-Za-z0-9._-]{8,128}$'),
    surface text NOT NULL DEFAULT 'api'
        CHECK (surface ~ '^[a-z][a-z0-9_.-]{1,49}$'),
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(metadata) = 'object'),
    occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX security_events_user_occurred_idx
    ON identity.security_events (user_id, occurred_at DESC)
    WHERE user_id IS NOT NULL;
CREATE INDEX security_events_type_occurred_idx
    ON identity.security_events (event_type, occurred_at DESC);

CREATE FUNCTION identity.reject_security_event_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'identity.security_events is append-only';
END;
$$;

CREATE TRIGGER security_events_immutable
BEFORE UPDATE OR DELETE ON identity.security_events
FOR EACH ROW EXECUTE FUNCTION identity.reject_security_event_mutation();

COMMENT ON TABLE identity.security_events IS
    'Append-only account security audit trail. Metadata must never contain credentials or tokens.';
COMMENT ON COLUMN identity.mfa_credentials.secret_ciphertext IS
    'AES-256-GCM ciphertext; the server-side encryption key is never stored in PostgreSQL.';
