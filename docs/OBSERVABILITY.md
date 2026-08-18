# UNSOLERO observability

Status: PARTIAL — cross-process source metrics are exposed; centralized collection and alert delivery remain external  
Last reviewed: 2026-08-18

Observability is operational evidence, not product analytics. It must not copy
customer profiles, analytics payloads, affiliate destinations, webhook bodies,
credentials, free text, or security secrets into logs or metric labels.

## Structured logs

Every Go process uses the same JSON logger. Sensitive attribute names are
redacted. Arbitrary error strings are not rendered; PostgreSQL failures are
reduced to safe classes such as `timeout`, `deadlock`, `serialization_failure`,
`constraint_violation`, and `unavailable`. HTTP completion records contain only
request ID, method, registered route pattern, status, response bytes, and
duration. They never contain the query, headers, cookies, body, IP address, raw
user agent, or route parameter value.

The API accepts a bounded safe `X-Request-ID` or generates a random one and
returns it to callers. Worker cycles log bounded counts and consecutive failure
counts. Provider/import/conversion records retain their own database run IDs;
an external log sink may correlate those identifiers but must apply a reviewed
retention policy.

## Metrics

When `METRICS_ENABLED=true`, `GET /api/v1/metrics` returns privacy-safe JSON
aggregates after constant-time validation of `Authorization: Bearer
<METRICS_TOKEN>`. The token must contain at least 32 characters and is never
logged. The endpoint is disabled when no token is configured.

`GET /api/v1/metrics/openmetrics` exposes the same bounded process snapshot in
OpenMetrics text format for a provider-neutral collector. It has the same bearer
authentication. It contains no product, user, session, click, conversion,
provider-event, request-body, query, IP, user-agent, destination, email, token,
or credential labels. The endpoint belongs on a backend-private network and
must not be exposed through public ingress.

Implemented bounded aggregates:

- request count, 5xx count, a fixed-bucket duration histogram, sum, and maximum
  by HTTP method, registered route pattern, and status class;
- rate-limit rejection and rate-limit backend failure;
- analytics accepted/rejected outcomes;
- authentication failures;
- recommendation generation failures;
- webhook failures;
- readiness, media-storage, email-delivery, provider-operation, and
  reconciliation-operation failures;
- pgx pool maximum/total/acquired/idle, wait count/time, and canceled acquire;
- durable worker active/backlog/success/failure/retry/dead/lease-recovery and
  average processing-latency state from PostgreSQL;
- durable worker last-success timestamp, heartbeat age, and consecutive
  checkpoint failure count from PostgreSQL;
- import, media deletion/reconciliation, backup, restore, and migration-
  fingerprint checkpoint state;
- Redis limiter latency and unavailable/failure counters.

Counter names are a compile-time allowlist. Unknown names are discarded rather
than retained, which prevents user/provider input from creating unbounded
series. HTTP labels are limited to method, registered route pattern, and status
class; IDs and raw paths are never labels.

PostgreSQL is the durable cross-process source for job/import/media/backup state;
each API scrape combines it with that replica's pgx pool and request counters.
An external collector must scrape every replica, deduplicate durable gauges,
retain history, and compute fleet rates. Current SQL aggregates have bounded
labels but scan retained operational tables; summary tables are required if
real retention makes scrape cost material.

Histogram buckets are fixed at 5, 10, 25, 50, 100, 250, 500, 1,000, 2,500 and
5,000 milliseconds plus infinity. Local probe percentiles are regression
evidence only; production SLOs, volume, and capacity remain unverified.

## Alerting boundary

The `alerting.Notifier` port defines database, migration, repeated-worker,
backlog, provider, webhook, reconciliation, backup, retention, authentication,
and rate-limit-backend categories. The checked-in `disabled` provider returns a
specific delivery-disabled error; it never reports success. Repeated worker
failures attempt an alert at the configured threshold.

`ALERT_PROVIDER=webhook` sends a bounded structured payload through an
authenticated HTTPS POST with a configured timeout. It rejects credentials,
query parameters, and fragments in the destination URL and never includes user
or commerce payloads. `ALERT_PROVIDER=external` still fails startup because no
provider-specific adapter is linked. Hosted staging and production reject
disabled alerting; local development reports it as degraded. No hosted alert
delivery was observed or fabricated.

## Production alert matrix

Thresholds below are initial engineering proposals. The service owner must tune
them against staging baselines and approved SLOs; until a destination is tested,
delivery is BLOCKED.

| Signal | Severity | Proposed threshold | Expected operator action | Escalation |
| --- | --- | --- | --- | --- |
| API readiness/schema incompatible | critical | any failure for 2 consecutive probes | stop rollout; inspect database and migration manifest | page primary immediately; database owner after 5 min |
| API 5xx ratio | high | >2% for 5 min and at least 50 requests | correlate release/request IDs; rollback or mitigate | page primary; incident commander at 10 min |
| p95 latency | warning/high | >1 s for 10 min / >3 s for 5 min | inspect database/provider saturation and deploy | ticket / page primary |
| Rate-limit backend unavailable | critical | any protected request cannot be evaluated | keep fail-closed; restore distributed limiter | page security/SRE immediately |
| Authentication failures | warning/high | approved anomaly threshold over 10 min | investigate abuse without logging identities | security on-call; incident for coordinated attack |
| Worker cycle failures | high | configured consecutive threshold (default 3) | inspect lease, DB and provider; avoid duplicate manual runs | page commerce/SRE |
| Import/provider age | warning/high | no success by 1x / 2x expected interval | disable stale publication/redirect; inspect provider | commerce / incident commander if impact |
| Failed import/webhook/replay | high | terminal failure or signature anomaly burst | quarantine input; preserve audit evidence; rotate key if indicated | commerce + security |
| Reconciliation backlog | high | oldest item exceeds approved settlement SLA | stop metrics if integrity is uncertain; reconcile | finance/commerce |
| Backup failed/old | critical | failed backup or newest verified backup exceeds approved RPO window | restore backup path; determine data-risk window | database/SRE + incident commander |
| Restore exercise | high | exercise missed or validation fails | block launch exception; remediate procedure | engineering leadership |
| Retention cleanup | warning/high | repeated failure or eligible record >24 h past policy | restore cleanup; assess privacy exposure | privacy/security if breached |
| Media scanner/storage | high | scanner unavailable or storage errors exceed baseline | uploads remain failed closed; inspect provider | platform/security |
| Telemetry export | high | no heartbeat from expected replica for 5 min | verify process/exporter and alternate health evidence | SRE; incident if blind during impact |

## Required external work

- Select metric/log/alert providers and review contracts and retention.
- Supply credentials through a secret manager, not `.env` or image layers.
- Define dashboards, paging severity, ownership, escalation, and maintenance
  windows.
- Alert on database/readiness failures, repeated worker failures, queue age,
  failed imports/webhooks, reconciliation backlog, backup age/failure, expired
  retention backlog, authentication anomalies, and provider degradation.
- Test alert delivery and redaction in staging before public traffic.
