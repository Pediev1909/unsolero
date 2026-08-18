# UNSOLERO operations

Status: repository controls implemented; external operations unassigned

## Health model

- `/api/v1/health/live` proves the process can answer HTTP and contains no
  dependency details.
- `/api/v1/health/ready` verifies PostgreSQL and critical abuse protection.
  Critical failure returns 503 and `unavailable`.
- Optional alert delivery failure returns 200 with `degraded`; the service may
  continue safely, but the deployment is not operationally ready for unattended
  traffic.
- Public diagnostics expose only dependency names and `ok`/`unavailable`.

## Worker model

The worker performs one bounded cycle at a time. Each cycle has a configured
deadline, maximum item count, and lease timeout. PostgreSQL `FOR UPDATE SKIP
LOCKED`, idempotency keys, provider event IDs, attempt limits, exponential retry
timestamps, stalled-lease recovery, terminal failure records, and manual retry
paths prevent unbounded or silently duplicated work. This is at-least-once
processing with idempotent boundaries—not exactly once.

SIGINT/SIGTERM cancels the active cycle, waits for adapters to honor context,
closes the pool, and exits. Docker grants 15 seconds before forced termination.
Repeated cycle failures trigger one alert attempt at the configured threshold;
the disabled provider reports that the alert was not delivered.

## Database operating policy

- Pool size/lifetime/idle/health parameters are centralized.
- Connection establishment, statements, locks, idle transactions, migrations,
  HTTP handlers, worker cycles, and shutdown all have bounded timeouts.
- API, worker, migration, and seed sessions use distinct application names.
- Deadlocks and serialization failures are classified but not blindly retried at
  the HTTP layer. Workers retry only through durable bounded job state.
- Migrations are checksum-verified, transactional, advisory-locked, and bounded
  by `DATABASE_MIGRATION_TIMEOUT`.
- Retention deletes are indexed, ordered, bounded, and skip locked rows.
- Offset pagination remains acceptable at current demonstrated volume; it needs
  measured replacement before high-volume claims.

## Failure behavior

- PostgreSQL loss keeps liveness `ok` but makes readiness return `503` with only
  `database: unavailable`. Handlers remain bounded by their context and session
  statement timeouts.
- A rate-limit backend error returns `503` for the protected operation, records a
  backend-failure metric, and attempts an alert. It never opens the gate.
- Disabled merchant/conversion providers record a controlled failure; they do
  not fabricate offers or conversions. Expired leases are moved to durable
  failure state and are eligible only for explicit bounded recovery.
- Disabled alert delivery returns a typed error and appears as optional degraded
  readiness. Operators must not interpret HTTP 200 degraded as delivered alerts.
- API SIGTERM stops accepting traffic and drains in-flight HTTP work up to
  `HTTP_SHUTDOWN_TIMEOUT`; it then forces close. Worker SIGTERM cancels the
  active bounded cycle, closes the database pool, and leaves durable lease state
  available for recovery.
- Failed migrations roll back their transaction and are not inserted into
  `platform.schema_migrations`. Never bypass this by manually inserting a row.
- Backup duplicate names, invalid archives/checksums, non-empty restore targets,
  and migration-count mismatch fail nonzero.

## Local drill evidence

Phase 7 exercised PostgreSQL outage/recovery, API and worker Compose SIGTERM,
in-flight HTTP draining, failed migration rollback, duplicate/idempotent job
execution, expired lease recovery, disabled providers/alerts, unavailable rate
backend, non-root backup, checksum validation, and clean restore. These are
local mechanics, not evidence that a cloud scheduler, on-call channel, or
production recovery objective works.

The protected `/api/v1/metrics` endpoint is process-local. Its aggregates are
useful for a one-replica staging drill; a production collector must replace or
scrape and aggregate every process without increasing label cardinality or
capturing personal data.

## Ownership still required

Name owners and backups for database, deployment, worker/import operations,
security incidents, privacy requests, merchant/provider incidents, backups,
alerts, and access reviews. Define SLOs only after representative staging/load
evidence. Maintain separate runbooks for provider outage, database saturation,
credential compromise, backup failure, and privacy/security incidents.
