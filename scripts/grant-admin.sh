#!/usr/bin/env bash
# Grants a role to an existing account.
#
# The account itself is created through the normal registration form, so the
# password is chosen by its owner, hashed by the application, and never passes
# through this script or anyone else's hands. Promotion is a separate step on
# purpose: it means an administrator is an ordinary member who was granted a
# role, not a special account created out of band with unknown provenance.
#
# Run from the deployment directory on the server:
#   ./scripts/grant-admin.sh you@example.com
#   ./scripts/grant-admin.sh you@example.com catalog_editor
#
# Defaults to 'admin', which carries every permission. Pass a narrower role
# when the person only needs part of the surface.
set -euo pipefail

EMAIL="${1:?usage: grant-admin.sh <email> [role_key]}"
ROLE="${2:-admin}"

cd "$(dirname "$0")/.."

compose() {
    docker compose -f compose.yaml -f compose.production.yaml "$@"
}

# ON CONFLICT DO NOTHING keeps a repeat run harmless. The RETURNING clause is
# what confirms the grant landed: a silent success would otherwise be
# indistinguishable from an email that matched no account.
result=$(compose exec -T postgres psql -U rigmark -d rigmark -qtAX -v ON_ERROR_STOP=1 \
    -v email="$EMAIL" -v role="$ROLE" <<'SQL'
WITH target AS (
    SELECT id FROM identity.users WHERE lower(email) = lower(:'email')
), granted AS (
    INSERT INTO identity.user_roles (user_id, role_key)
    SELECT id, :'role' FROM target
    ON CONFLICT (user_id, role_key) DO NOTHING
    RETURNING user_id
)
SELECT
    (SELECT count(*) FROM target),
    (SELECT count(*) FROM granted);
SQL
)

found="${result%%|*}"
inserted="${result##*|}"

if [ "$found" -eq 0 ]; then
    echo "No account found for ${EMAIL}. Register it on the site first, then re-run." >&2
    exit 1
fi

if [ "$inserted" -eq 0 ]; then
    echo "${EMAIL} already held ${ROLE}; nothing changed."
else
    echo "Granted ${ROLE} to ${EMAIL}."
fi

echo "Sign out and back in so the new session carries the role."
