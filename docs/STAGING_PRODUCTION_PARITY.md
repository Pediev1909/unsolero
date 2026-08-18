# Staging and production parity

Status: PARTIAL — production-shaped local topology is exercised; no hosted managed staging exists  
Last reviewed: 2026-08-18

Staging must use the production application build, schema, configuration validation, and provider boundaries. It may use isolated provider sandboxes and fictional fixtures. It must never use production credentials, real customer data, or real financial events.

## Required topology

| Concern | Repository evidence | Staging requirement | Status |
| --- | --- | --- | --- |
| API and worker | separate non-root, read-only image targets | same immutable image digest and release ID as the production candidate | PASS |
| PostgreSQL | shared PostgreSQL 17, 19 checksum-verified migrations, bounded pool/timeouts | managed PostgreSQL 17 over TLS; separate runtime/migration/backup roles | PARTIAL |
| Rate limits | atomic Redis adapter, HMAC-pseudonymous keys, fail-closed outage behavior | TLS/authenticated Redis-compatible service reachable only privately | PARTIAL |
| Media | private S3-compatible adapter and durable deletion jobs | private bucket, TLS, KMS policy, lifecycle, inventory/reconciliation, scanner | PARTIAL |
| Email | provider-neutral SMTP contract and fail-closed configuration | reviewed sandbox provider, dedicated domain, TLS, bounce/complaint path | BLOCKED |
| Telemetry | redacted JSON logs and authenticated bounded OpenMetrics endpoint | centralized collector, retention, dashboards, alert delivery | EXTERNAL |
| Edge | local TLS ingress, secure cookies, trusted proxy, real route resolver | managed certificates, exact proxy CIDR, request/body limits, WAF policy | PARTIAL |
| Secrets | production validation rejects unsafe values | managed secret injection and rotation; no `.env` or image-layer secrets | EXTERNAL |

`compose.staging.yaml` overlays two API replicas, two workers, shared
PostgreSQL/authenticated Redis/private MinIO, S3-only media, one-shot migration
and bucket initialization, local TLS ingress, secure cookies, trusted proxy,
read-only/non-root application containers, health checks, and resource limits.
Live provider adapters remain disabled. This is not a production orchestrator.

## Local parity harness

Use non-production credentials and fictional data only:

```sh
cp .env.staging.example .env.staging
docker compose --env-file .env.staging -f compose.yaml -f compose.staging.yaml \
  --profile staging up --build -d
./scripts/check-routing-semantics.sh https://localhost:8443
./scripts/run-staging-performance-gates.sh https://localhost:8443
```

Image references in Compose are digest-pinned. CI starts isolated pinned Redis and MinIO containers and supplies only test credentials to the gated integration tests.

Phase 11 locally exercised dependency outages, single-replica termination,
backup/clean restore with 19 migrations, media reconciliation, HTTPS routing,
OpenMetrics, browser/accessibility paths, and bounded load. Docker volumes are
not managed HA services; the certificate is not public PKI; MinIO does not
prove a selected S3 provider; local ignored env files are not a secret manager;
there is no collector/alert destination; and backup lacks KMS, off-site copies,
PITR, and a production RPO/RTO.

## Promotion gate

Before promoting an immutable candidate, record:

1. commit and image digests, SBOM, scan results, migration manifest, and configuration review;
2. fresh migration and seed-free staging startup;
3. readiness, authentication, recommendation, media, rate-limit, commerce-disabled, analytics-consent, and 404/API smoke evidence;
4. Redis and object-store outage results, worker restart/lease recovery, alert delivery, backup, restore, and rollback evidence;
5. browser/accessibility matrix and performance-budget result;
6. provider, security, privacy, legal, accessibility, operations, and launch approvals.

No repository document is an approval record. Any unknown, expired, or missing gate blocks promotion.

## Blocker classification

- **A — IMPLEMENT NOW:** repository defect or missing test that can be completed without external systems.
- **B — CONFIGURE IN STAGING:** deployment values, topology, dashboards, limits, or exercises requiring the selected staging environment.
- **C — EXTERNAL INFRASTRUCTURE:** hosted database, Redis, object storage, secret manager, ingress, telemetry, alerting, backup, or KMS.
- **D — EXTERNAL PROVIDER:** email, scanner, merchant, affiliate, conversion, AI, or other provider credentials/contracts/sandboxes.
- **E — LEGAL/BUSINESS:** policies, disclosures, agreements, supported markets, retention, and accountable launch authority.
- **F — INDEPENDENT VALIDATION:** penetration, accessibility, architecture, and observed recovery assessment by independent reviewers.
