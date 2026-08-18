#!/bin/sh
set -eu

: "${PGDATABASE:?PGDATABASE is required}"
: "${BACKUP_FILE:?BACKUP_FILE is required}"

record_restore_failure() {
  exit_status=$?
  trap - EXIT HUP INT TERM
  if [ "$exit_status" -ne 0 ] && [ "${RECORD_OPERATIONAL_CHECKPOINTS:-false}" = "true" ]; then
    psql --no-psqlrc "$PGDATABASE" -c "INSERT INTO platform.operational_checkpoints
      (checkpoint_name,status,observed_at,failure_count,detail_code)
      VALUES ('restore_verification','failed',now(),1,'restore.command_failed')
      ON CONFLICT (checkpoint_name) DO UPDATE SET status='failed',observed_at=now(),
      failure_count=platform.operational_checkpoints.failure_count+1,
      detail_code='restore.command_failed',updated_at=now()" >/dev/null || true
  fi
  exit "$exit_status"
}
trap record_restore_failure EXIT HUP INT TERM

case "$BACKUP_FILE" in
  ""|"/") echo "BACKUP_FILE must identify a custom-format backup" >&2; exit 2 ;;
esac

checksum_file="$BACKUP_FILE.sha256"
metadata_file="$BACKUP_FILE.metadata"
if [ ! -f "$BACKUP_FILE" ] || [ ! -f "$checksum_file" ] || [ ! -f "$metadata_file" ]; then
  echo "backup, checksum, and metadata files are required" >&2
  exit 3
fi

(cd "$(dirname "$BACKUP_FILE")" && sha256sum -c "$(basename "$checksum_file")") >/dev/null
pg_restore --list "$BACKUP_FILE" >/dev/null

existing_relations=$(psql --no-psqlrc --tuples-only --no-align "$PGDATABASE" -c \
  "SELECT count(*) FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname NOT IN ('pg_catalog','information_schema') AND n.nspname NOT LIKE 'pg_toast%' AND c.relkind IN ('r','p','v','m','S')")
if [ "$existing_relations" != "0" ]; then
  echo "restore target is not empty" >&2
  exit 4
fi

pg_restore --exit-on-error --single-transaction --no-owner --no-privileges --dbname "$PGDATABASE" "$BACKUP_FILE"
restored_migrations=$(psql --no-psqlrc --tuples-only --no-align "$PGDATABASE" -c \
  "SELECT count(*) FROM platform.schema_migrations")
expected_migrations=$(sed -n 's/^migration_count=//p' "$metadata_file")
if [ -n "$expected_migrations" ] && [ "$expected_migrations" != "unknown" ] && [ "$restored_migrations" != "$expected_migrations" ]; then
  echo "restored migration count does not match backup metadata" >&2
  exit 5
fi

invalid_migrations=$(psql --no-psqlrc --tuples-only --no-align "$PGDATABASE" -c \
  "SELECT count(*) FROM platform.schema_migrations WHERE checksum !~ '^[0-9a-f]{64}$'")
if [ "$invalid_migrations" != "0" ]; then
  echo "restored migration checksums are invalid" >&2
  exit 6
fi

expected_schema_fingerprint=$(sed -n 's/^schema_fingerprint=//p' "$metadata_file")
if [ -n "$expected_schema_fingerprint" ]; then
  restored_migration_manifest=$(psql --no-psqlrc --tuples-only --no-align "$PGDATABASE" -c \
    "SELECT version::text || ':' || name || ':' || checksum FROM platform.schema_migrations ORDER BY version")
  restored_schema_fingerprint=$(printf '%s\n' "$restored_migration_manifest" | sha256sum | awk '{print $1}')
  if [ "$restored_schema_fingerprint" != "$expected_schema_fingerprint" ]; then
	if [ "${RECORD_OPERATIONAL_CHECKPOINTS:-false}" = "true" ]; then
	  psql --no-psqlrc "$PGDATABASE" -c "INSERT INTO platform.operational_checkpoints
	    (checkpoint_name,status,observed_at,failure_count,detail_code)
	    VALUES ('restore_verification','mismatch',now(),1,'restore.migration_fingerprint_mismatch')
	    ON CONFLICT (checkpoint_name) DO UPDATE SET status='mismatch',observed_at=now(),
	    failure_count=platform.operational_checkpoints.failure_count+1,
	    detail_code='restore.migration_fingerprint_mismatch',updated_at=now()" >/dev/null || true
	fi
    echo "restored migration manifest does not match backup metadata" >&2
    exit 7
  fi
fi

if [ "${RECORD_OPERATIONAL_CHECKPOINTS:-false}" = "true" ]; then
  psql --no-psqlrc "$PGDATABASE" -c "INSERT INTO platform.operational_checkpoints
    (checkpoint_name,status,observed_at,last_success_at,failure_count,detail_code)
    VALUES ('restore_verification','ok',now(),now(),0,NULL)
    ON CONFLICT (checkpoint_name) DO UPDATE SET status='ok',observed_at=now(),
    last_success_at=now(),detail_code=NULL,updated_at=now()" >/dev/null
fi

trap - EXIT HUP INT TERM
echo "restore_verified=$BACKUP_FILE migrations=$restored_migrations"
