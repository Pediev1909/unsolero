# Phase 11 evidence matrix

Date: 2026-08-18  
Verdict: **PARTIAL**  
Scope: repository and disposable local production-shaped staging only

| Area | Requirement | Test | Result | Evidence | Limitations | Production implication |
| --- | --- | --- | --- | --- | --- | --- |
| Media | bounded crash-window reconciliation | unit, PostgreSQL, MinIO adapter, dry-run CLI | PASS | deterministic key parser, leases, cursors, audit results, safety grace | no managed-bucket inventory | validate selected object provider before launch |
| Routing | real unknown-route status | HTTP tests and `check-routing-semantics.sh` through TLS edge | PASS | known 200, unknown/API 404, malformed 400, slash 308 | no crawler lab | false-200 blocker closed locally |
| SEO | canonical/noindex/sitemap/robots | handler and edge header assertions | PASS | canonical `Link`; private/admin/query noindex | React content remains client-rendered | search rendering still requires validation |
| Telemetry | cross-process operational state | authenticated JSON/OpenMetrics scrape | PARTIAL | DB pool plus durable worker/import/media/backup state and bounded HTTP histograms; staging scrape preserves registered patterns across context copies | no collector/history; repository-wide query-timeout/transaction-failure instrumentation is incomplete | finish instrumentation, then deploy collector/dashboards |
| Alerts | definitions and provider boundary | unit tests | PARTIAL | bounded provider-neutral rule catalog | delivery adapter intentionally disabled | real paging exercise is blocked |
| Frontend performance | build artifact budgets | `npm run build && npm run budget:check` | PASS | entry JS 86,794 B gzip; CSS 11,577 B; initial 151,319 B | no field Web Vitals | keep as regression gate, not UX SLA |
| API performance | bounded staging gates | `run-staging-performance-gates.sh` | PASS | all configured scenarios 0% errors and under local p95 ceilings | fictional small dataset/local machine | no capacity claim |
| Soak | sustained bounded catalog traffic | 30-second, concurrency-8 load probe | PASS | 3,216/3,216; p50 90.1 ms, p95 183.7 ms, p99 195.6 ms | short local soak | longer hosted soak required |
| Query plans | critical bounded lookups | PostgreSQL `EXPLAIN (FORMAT JSON)` integration test | PASS | index nodes required with seqscan disabled | small fictional dataset | repeat at representative scale |
| Dependency security | deterministic repository gates | npm audit, govulncheck, module/lock verification, secret/sink/digest scripts | PARTIAL | npm production audit found 0; govulncheck found 0 reachable/imported vulnerabilities; deterministic scripts passed; workflows are SHA-pinned | hosted CI/SBOM/image scanners not executed here | run protected CI and artifact scan |
| Staging topology | two API/two worker/shared dependencies/TLS | Compose config/build/up and health | PASS locally | secure cookie, trusted proxy, S3 media, Redis, resource limits | Docker is not managed infrastructure | hosted staging still required |
| Redis recovery | fail closed and recover | controlled stop/start | PASS | readiness/public/analytics/affiliate 503; recovery 200 | no managed failover | managed Redis drill required |
| PostgreSQL recovery | fail closed and recover | controlled stop/start | PASS | readiness 503; catalog generic 500; recovery 200 | no HA/PITR | database outage removes serving capacity as expected |
| Object storage recovery | readiness transition | controlled MinIO stop/start | PASS | readiness 503 then 200 | no managed provider | provider-specific drill required |
| Replica loss | API/worker termination | one exact container stopped/restarted | PASS locally | API remained 200; second worker remained healthy | Compose DNS is not a production load balancer | validate real ingress/orchestrator |
| Jobs | duplicate workers and lease recovery | two-worker staging plus integration tests | PASS | SKIP LOCKED/idempotency/expired lease cases | no provider traffic | repeat with real sandbox feed |
| Backup/restore | clean restore and fingerprint | backup and separate restore project | PASS locally | checksum-verified dump restored with 19 migrations | local bind mount, no PITR/KMS/off-site copy | production DR blocked |
| Browser | desktop/mobile/accessibility/error flows | Playwright/Axe matrix | PASS with intentional skips | 20 passed, 4 intentional skips; public/account/admin/affiliate/consent/error and 320–1920 px checks | development email lifecycle is not runnable with staging email disabled | sandbox email and independent AT review required |
| Live providers | financial/email/AI/scanner/alerts | configuration inspection | BLOCKED | every live adapter remains disabled/fail-closed | credentials/contracts absent by design | cannot launch affected capability |
| Legal/business | policies and launch authorization | document inventory only | NOT TESTED | no legal conclusions asserted | counsel and accountable owners absent | public launch remains rejected |

`PASS locally` is not a claim about managed production infrastructure,
production traffic, legal compliance, provider certification, or capacity.
