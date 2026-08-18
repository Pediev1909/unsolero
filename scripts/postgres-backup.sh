#!/bin/sh
set -eu

: "${PGDATABASE:?PGDATABASE is required}"
: "${BACKUP_DIRECTORY:?BACKUP_DIRECTORY is required}"

case "$BACKUP_DIRECTORY" in
  ""|"/") echo "BACKUP_DIRECTORY must be a dedicated directory" >&2; exit 2 ;;
esac

umask 077
mkdir -p "$BACKUP_DIRECTORY"
backup_name=${BACKUP_NAME:-"unsolero-$(date -u +%Y%m%dT%H%M%SZ)"}
case "$backup_name" in
  *[!A-Za-z0-9._-]*) echo "BACKUP_NAME contains unsupported characters" >&2; exit 2 ;;
esac

final_path="$BACKUP_DIRECTORY/$backup_name.dump"
temporary_path="$BACKUP_DIRECTORY/.$backup_name.partial"
checksum_path="$final_path.sha256"
metadata_path="$final_path.metadata"

if [ -e "$final_path" ] || [ -e "$checksum_path" ] || [ -e "$metadata_path" ]; then
  echo "backup target already exists" >&2
  exit 3
fi

cleanup() { rm -f "$temporary_path" "$checksum_path.partial" "$metadata_path.partial"; }
on_exit() {
  exit_status=$?
  trap - EXIT HUP INT TERM
  cleanup
  if [ "$exit_status" -ne 0 ] && [ "${RECORD_OPERATIONAL_CHECKPOINTS:-false}" = "true" ]; then
    psql --no-psqlrc "$PGDATABASE" -c "INSERT INTO platform.operational_checkpoints
      (checkpoint_name,status,observed_at,failure_count,detail_code)
      VALUES ('backup','failed',now(),1,'backup.command_failed')
      ON CONFLICT (checkpoint_name) DO UPDATE SET status='failed',observed_at=now(),
      failure_count=platform.operational_checkpoints.failure_count+1,
      detail_code='backup.command_failed',updated_at=now()" >/dev/null || true
  fi
  exit "$exit_status"
}
trap on_exit EXIT HUP INT TERM

pg_dump --format=custom --compress=9 --no-owner --no-privileges --file "$temporary_path" "$PGDATABASE"
pg_restore --list "$temporary_path" >/dev/null
migration_count=$(psql --no-psqlrc --tuples-only --no-align "$PGDATABASE" -c \
  "SELECT count(*) FROM platform.schema_migrations")
migration_manifest=$(psql --no-psqlrc --tuples-only --no-align "$PGDATABASE" -c \
  "SELECT version::text || ':' || name || ':' || checksum FROM platform.schema_migrations ORDER BY version")
schema_fingerprint=$(printf '%s\n' "$migration_manifest" | sha256sum | awk '{print $1}')
created_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)

{
  echo "format=postgres-custom"
  echo "created_at=$created_at"
  echo "database=$PGDATABASE"
  echo "migration_count=$migration_count"
  echo "schema_fingerprint=$schema_fingerprint"
  echo "encryption=none-local-artifact"
} > "$metadata_path.partial"
{
  sha256sum "$temporary_path" | sed "s|$temporary_path|$(basename "$final_path")|"
  sha256sum "$metadata_path.partial" | sed "s|$metadata_path.partial|$(basename "$metadata_path")|"
} > "$checksum_path.partial"

mv "$temporary_path" "$final_path"
mv "$metadata_path.partial" "$metadata_path"
mv "$checksum_path.partial" "$checksum_path"
if [ "${RECORD_OPERATIONAL_CHECKPOINTS:-false}" = "true" ]; then
  psql --no-psqlrc "$PGDATABASE" -c "INSERT INTO platform.operational_checkpoints
    (checkpoint_name,status,observed_at,last_success_at,failure_count,detail_code)
    VALUES ('backup','ok',now(),now(),0,NULL)
    ON CONFLICT (checkpoint_name) DO UPDATE SET status='ok',observed_at=now(),
    last_success_at=now(),detail_code=NULL,updated_at=now()" >/dev/null
fi
trap - EXIT HUP INT TERM
echo "backup_created=$final_path"
