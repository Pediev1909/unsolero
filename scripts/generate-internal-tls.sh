#!/usr/bin/env bash
# Generates the TLS material the single-box production stack needs for its
# in-network dependencies.
#
# Production configuration requires encrypted transport to PostgreSQL and to
# the Redis-compatible rate-limit store. On a managed platform those come with
# provider certificates; on one box they have to be issued here. This is the
# supported way to satisfy that requirement — the alternative would be relaxing
# a security check in the application, which is never correct.
#
# Two different trust models are used deliberately:
#
#   PostgreSQL uses sslmode=require, which encrypts without verifying the
#   server certificate, so a self-signed certificate is sufficient.
#
#   The Redis client performs full verification, so its certificate is signed
#   by a local CA and that CA is appended to a copy of the runtime image's
#   trust bundle. The bundle keeps the public roots, so outbound TLS to object
#   storage, SMTP and the alert webhook continues to work.
#
# Re-running regenerates everything. Certificates are valid for 825 days.

set -euo pipefail

output_dir="${1:-./secrets/tls}"
runtime_image="${RUNTIME_IMAGE:-alpine:3.22}"
days=825

if ! command -v openssl >/dev/null 2>&1; then
    echo "openssl is required" >&2
    exit 1
fi

mkdir -p "$output_dir"
cd "$output_dir"

echo "==> Issuing internal CA"
openssl req -x509 -newkey rsa:4096 -sha256 -days "$days" -nodes \
    -keyout internal-ca.key -out internal-ca.crt \
    -subj "/CN=UNSOLERO internal CA" \
    -addext "basicConstraints=critical,CA:TRUE,pathlen:0" \
    -addext "keyUsage=critical,keyCertSign,cRLSign" 2>/dev/null

issue_certificate() {
    local name="$1"
    echo "==> Issuing certificate for ${name}"
    openssl req -newkey rsa:2048 -sha256 -nodes \
        -keyout "${name}.key" -out "${name}.csr" \
        -subj "/CN=${name}" 2>/dev/null
    openssl x509 -req -in "${name}.csr" -days "$days" -sha256 \
        -CA internal-ca.crt -CAkey internal-ca.key -CAcreateserial \
        -extfile <(printf 'subjectAltName=DNS:%s\nkeyUsage=critical,digitalSignature,keyEncipherment\nextendedKeyUsage=serverAuth\n' "$name") \
        -out "${name}.crt" 2>/dev/null
    rm -f "${name}.csr"
}

# The service names are the Compose DNS names the application connects to, so
# they must match the certificate subject exactly.
issue_certificate postgres
issue_certificate valkey

echo "==> Building the combined trust bundle"
# The runtime image's own bundle is the source of the public roots; appending
# rather than replacing is what keeps external TLS working.
if ! docker run --rm --entrypoint cat "$runtime_image" \
        /etc/ssl/certs/ca-certificates.crt > system-roots.pem 2>/dev/null; then
    echo "Could not read the trust bundle from ${runtime_image}." >&2
    echo "Set RUNTIME_IMAGE to an image that has ca-certificates installed." >&2
    exit 1
fi
cat system-roots.pem internal-ca.crt > trust-bundle.pem
rm -f system-roots.pem

# PostgreSQL refuses to start if its key is group- or world-readable, and the
# container runs as a different uid than the host user that generated it.
chmod 600 ./*.key
chmod 644 ./*.crt trust-bundle.pem

echo
echo "TLS material written to $(pwd)"
echo
echo "  postgres.crt / postgres.key   PostgreSQL server certificate"
echo "  valkey.crt / valkey.key       Valkey server certificate"
echo "  internal-ca.crt               local CA that signed both"
echo "  trust-bundle.pem              public roots plus the local CA"
echo
echo "internal-ca.key is the private key of the CA that your services trust."
echo "Keep it off the server once issuing is done, and never commit it."
