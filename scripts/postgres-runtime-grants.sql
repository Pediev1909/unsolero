\set ON_ERROR_STOP on

-- Run as the migration/object-owner role after migrations. Roles are created
-- and authenticated by the managed database/secret manager, never by this
-- repository. Example:
-- psql "$MIGRATOR_DATABASE_URL" \
--   --set=database_name=unsolero --set=api_role=unsolero_api \
--   --set=worker_role=unsolero_worker --set=backup_role=unsolero_backup \
--   --file=scripts/postgres-runtime-grants.sql

REVOKE ALL ON DATABASE :"database_name" FROM PUBLIC;
REVOKE CREATE ON SCHEMA public FROM PUBLIC;

GRANT CONNECT ON DATABASE :"database_name" TO :"api_role", :"worker_role", :"backup_role";

GRANT USAGE ON SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
TO :"api_role", :"worker_role", :"backup_role";

-- Runtime roles can manipulate domain records but cannot create/alter/drop
-- schema objects, manage roles, or bypass row-level security. The API and
-- worker currently share repository modules, so a narrower table split would
-- be unsafe until repository ownership is separated and integration-tested.
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
TO :"api_role", :"worker_role";
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
TO :"api_role", :"worker_role";

GRANT SELECT ON ALL TABLES IN SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
TO :"backup_role";

ALTER DEFAULT PRIVILEGES IN SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO :"api_role", :"worker_role";
ALTER DEFAULT PRIVILEGES IN SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
GRANT USAGE, SELECT ON SEQUENCES TO :"api_role", :"worker_role";
ALTER DEFAULT PRIVILEGES IN SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
GRANT SELECT ON TABLES TO :"backup_role";

REVOKE CREATE ON SCHEMA
  identity, planning, catalog, evidence, compatibility, recommendation,
  commerce, analytics, administration, admin, editorial, platform
FROM :"api_role", :"worker_role", :"backup_role";
