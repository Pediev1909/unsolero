#!/usr/bin/env bash
# Takes a database backup and copies it off the server.
#
# A backup that only exists on the machine it protects is not a backup. On a
# single-box deployment the host is the single point of failure, so the dump is
# encrypted and pushed to object storage in one run. Everything here uses
# containers, so the server needs no tooling beyond Docker.
#
# Intended to run from cron on the server:
#   0 3 * * * cd /opt/unsolero && ./scripts/backup-to-r2.sh >> /var/log/unsolero-backup.log 2>&1
#
# Exits non-zero on any failure so cron's mail, or a wrapper, surfaces it. A
# silent backup failure is the expensive kind.

set -euo pipefail

cd "$(dirname "$0")/.."

if [ ! -f .env ]; then
    echo "No .env in $(pwd); run this from the deployment directory." >&2
    exit 1
fi
# shellcheck disable=SC1091
set -a && . ./.env && set +a

: "${BACKUP_S3_BUCKET:?Set BACKUP_S3_BUCKET in .env}"
: "${MEDIA_S3_ENDPOINT:?Set MEDIA_S3_ENDPOINT in .env}"
: "${MEDIA_S3_ACCESS_KEY:?Set MEDIA_S3_ACCESS_KEY in .env}"
: "${MEDIA_S3_SECRET_KEY:?Set MEDIA_S3_SECRET_KEY in .env}"

retention_days="${BACKUP_RETENTION_DAYS:-30}"
backup_dir="./backups"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
backup_name="unsolero-${stamp}"

echo "==> Dumping the database"
BACKUP_NAME="$backup_name" docker compose \
    -f compose.yaml -f compose.production.yaml \
    --profile tools run --rm backup

dump_path="${backup_dir}/${backup_name}.dump"
if [ ! -f "$dump_path" ]; then
    echo "Expected dump at ${dump_path} but it was not produced." >&2
    exit 1
fi

upload_path="$dump_path"
# Encryption is applied when a recipient key is configured. The dump contains
# every user record, so an unencrypted copy in third-party storage is a real
# exposure, not a theoretical one.
if [ -n "${BACKUP_AGE_RECIPIENT:-}" ]; then
    echo "==> Encrypting"
    # encrypt-backup-age.sh takes its input through the environment and writes
    # a .tar.age archive that also carries the checksum and metadata files.
    BACKUP_FILE="$dump_path" \
    ARCHIVE_DIRECTORY="$backup_dir" \
    AGE_RECIPIENT="$BACKUP_AGE_RECIPIENT" \
        ./scripts/encrypt-backup-age.sh
    upload_path="${backup_dir}/${backup_name}.tar.age"
    if [ ! -f "$upload_path" ]; then
        echo "Encryption did not produce ${upload_path}." >&2
        exit 1
    fi
    # The plaintext dump must not linger next to the encrypted archive; the
    # archive already contains it along with its checksum and metadata.
    shred -u "$dump_path" "${dump_path}.sha256" "${dump_path}.metadata" 2>/dev/null \
        || rm -f "$dump_path" "${dump_path}.sha256" "${dump_path}.metadata"
else
    echo "WARNING: BACKUP_AGE_RECIPIENT is not set, so the dump is uploaded unencrypted." >&2
    echo "         It contains all user data. Set a recipient key before relying on this." >&2
fi

echo "==> Uploading to r2://${BACKUP_S3_BUCKET}"
run_aws() {
    docker run --rm \
        -e AWS_ACCESS_KEY_ID="$MEDIA_S3_ACCESS_KEY" \
        -e AWS_SECRET_ACCESS_KEY="$MEDIA_S3_SECRET_KEY" \
        -e AWS_DEFAULT_REGION="${MEDIA_S3_REGION:-auto}" \
        -v "$(pwd)/${backup_dir}:/backups:ro" \
        amazon/aws-cli:latest \
        --endpoint-url "https://${MEDIA_S3_ENDPOINT}" \
        "$@"
}

run_aws s3 cp "/backups/$(basename "$upload_path")" "s3://${BACKUP_S3_BUCKET}/$(basename "$upload_path")"
# The checksum travels with the object so a restore can prove it arrived intact.
# The encrypted archive carries its own checksum file alongside it.
for companion in "${upload_path}.sha256" "${dump_path}.sha256"; do
    if [ -f "$companion" ]; then
        run_aws s3 cp "/backups/$(basename "$companion")" \
            "s3://${BACKUP_S3_BUCKET}/$(basename "$companion")"
        break
    fi
done

echo "==> Verifying the upload is listed"
if ! run_aws s3 ls "s3://${BACKUP_S3_BUCKET}/$(basename "$upload_path")" | grep -q .; then
    echo "Upload did not appear in the bucket listing." >&2
    exit 1
fi

echo "==> Pruning local copies older than ${retention_days} days"
find "$backup_dir" -type f -mtime "+${retention_days}" -name 'unsolero-*' -delete

echo
echo "Backup complete: $(basename "$upload_path")"
echo
echo "A backup is only proven by a restore. Exercise scripts/postgres-restore.sh"
echo "against a scratch database on a schedule; an untested backup is a guess."
