# UNSOLERO deployment runbook

Status: provider-neutral runbook; no production environment is approved

## Pre-deployment

1. Identify application version, commit, image digests, migration set, operator,
   rollback version, and maintenance window.
2. Run formatting, typecheck, lint, unit/race/integration tests, builds,
   vulnerability scans, Compose validation, backup/restore drill, and staging
   smoke tests.
3. Confirm production configuration validation succeeds using secret references,
   not printed secret values. Confirm secure cookies, HTTPS origin, PostgreSQL
   TLS, MFA key, rate-limit HMAC key, email adapter, and replica/limiter policy.
4. Confirm alert delivery, metrics scraping, dashboards, on-call ownership,
   backup freshness, database capacity, and provider health.
5. Create and verify a pre-migration backup.

## Database identities

Use separate secret-manager identities for migration, API, worker, backup, and
restore. The migrator/object owner runs migrations and then applies
`scripts/postgres-runtime-grants.sql` with explicit psql role variables. The
script removes public database/schema creation, grants runtime DML without DDL,
and grants the backup role read-only table access. Roles are created and login
credentials are issued by the managed database, not by migrations.

The API and worker currently share several repository modules, so both runtime
roles intentionally receive the same table-level DML set. This is narrower
than owner/superuser access but not the final table-by-table split; further
restriction requires repository-specific integration tests before launch.

## Deployment order

1. Stop or quiesce workers if the migration changes tables they process.
2. Run the migration image once. PostgreSQL advisory locking serializes
   contenders; checksum mismatch or timeout fails the deployment.
3. Inspect migration status and readiness. Never mark a failed migration as
   applied manually.
4. Deploy API instances gradually while preserving at least one compatible
   serving version where the migration allows it.
5. Deploy workers after schema/API compatibility is confirmed.
6. Verify liveness, readiness, degraded dependencies, metrics, logs, and worker
   cycles; execute health, auth, catalog, recommendation, merchant navigation,
   and admin authorization smoke tests without creating fake commercial facts.

## Rollback

Application images/configuration are reversible when the old application is
compatible with the new schema. Database migrations are forward-only and may be
irreversible. Never assume application rollback implies schema rollback.

For a bad application release, stop rollout and restore the prior image. For an
incompatible or corrupt schema change, stop writes, involve the database owner,
and choose a corrective forward migration or verified restore. A restore loses
changes after its recovery point and requires explicit incident authority.

## Secrets and emergencies

- Rotate one credential class at a time with overlap where supported; revoke old
  values only after health and authentication checks.
- Database, MFA, rate-limit, metrics, email, merchant, conversion, AI, alerting,
  and backup credentials require independent ownership and audit.
- Emergency shutdown removes traffic, stops workers, preserves evidence, and
  prevents merchant/webhook ingestion before destructive recovery actions.
- Record timeline, decisions, affected versions, and recovery evidence without
  copying secrets or customer payloads into the incident log.
