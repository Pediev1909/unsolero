-- Newsletter subscriptions are the audience schema's first table. The list is
-- double opt-in: a row is created as `pending` and becomes `confirmed` only when
-- the one-time confirmation token from the email is presented back. Token
-- material is stored as SHA-256 hashes; the raw token exists only inside the
-- email. No IP address, user agent, or account link is recorded: the address
-- and its consent state are all the list needs.
CREATE SCHEMA IF NOT EXISTS audience;

CREATE TABLE audience.newsletter_subscriptions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL CHECK (
        email = lower(btrim(email))
        AND char_length(email) BETWEEN 3 AND 254
        AND position('@' IN email) > 1
    ),
    status text NOT NULL CHECK (status IN ('pending', 'confirmed', 'unsubscribed')),
    confirm_token_hash bytea UNIQUE CHECK (
        confirm_token_hash IS NULL OR octet_length(confirm_token_hash) = 32
    ),
    confirm_expires_at timestamptz,
    unsubscribe_token_hash bytea NOT NULL UNIQUE
        CHECK (octet_length(unsubscribe_token_hash) = 32),
    source text NOT NULL CHECK (source ~ '^[a-z][a-z0-9_.:-]{0,99}$'),
    consent_text_version text NOT NULL
        CHECK (char_length(btrim(consent_text_version)) BETWEEN 1 AND 64),
    requested_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    unsubscribed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    -- A pending row carries exactly one live confirmation token, and the token
    -- is cleared the moment it is consumed or the address leaves that state.
    CHECK ((status = 'pending') = (confirm_token_hash IS NOT NULL)),
    CHECK ((confirm_token_hash IS NULL) = (confirm_expires_at IS NULL)),
    CHECK (confirm_expires_at IS NULL OR confirm_expires_at > requested_at),
    CHECK (status <> 'confirmed' OR confirmed_at IS NOT NULL),
    CHECK ((status = 'unsubscribed') = (unsubscribed_at IS NOT NULL))
);

CREATE UNIQUE INDEX newsletter_subscriptions_email_idx
    ON audience.newsletter_subscriptions (lower(email));
CREATE INDEX newsletter_subscriptions_pending_expiry_idx
    ON audience.newsletter_subscriptions (confirm_expires_at)
    WHERE status = 'pending';

COMMENT ON TABLE audience.newsletter_subscriptions IS
    'Double opt-in newsletter list: address, consent state, consent text version, and hashed one-time tokens only. No IP address or user agent.';
