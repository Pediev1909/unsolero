# UNSOLERO Phase 9 production validation matrix

Status: PARTIAL — Phase 9 repository work is complete, but public launch remains rejected

`PASS` means the repository/local evidence satisfied the stated scope. It does
not mean the system or vendor dependency is production-approved.

| Area | Status | Evidence / blocker |
| --- | --- | --- |
| core public functionality | PASS | catalog, recommendations, redirects, health and frontend checks pass locally |
| authenticated functionality | PASS | account/admin/commerce and persisted recommendation scenarios pass locally |
| authentication and MFA correctness | PASS | unit/integration/concurrency/race suite; production email provider still blocked |
| authorization and ownership | PASS | role permissions, admin denial, setup/session/recommendation ownership tests |
| recommendation determinism/reproducibility | PASS | scenario, policy, historical snapshot and commercial-invariance tests |
| recommendation/commercial separation | PASS | structural type tests and full commerce mutation integration test |
| database migrations | PASS | transactional rollback, advisory lock, checksums, embedded readiness manifest |
| database scale/query behavior | PARTIAL | 100k-event rollback fixture and plans measured; no production-sized clone/soak |
| database concurrency/pool exhaustion | PASS | single-use/idempotency tests, race suite, caller deadline and slow-query timeout |
| merchant ingestion | PARTIAL | provider-neutral jobs/retries/audit tested; live adapters and credentials absent |
| verified conversions | PARTIAL | signed/mock lifecycle, replay, reversal, reporting tested; no live provider certification |
| affiliate tracking | PASS | fresh/owned redirect, idempotency, bot classification, normal 302 load |
| analytics/privacy governance | PARTIAL | consent, dedupe, claims, retention and filtered reports tested; legal approval absent |
| worker resilience | PASS | leases, bounded retries, disabled adapters, cleanup and queue state tested locally |
| API resilience | PASS | timeouts, body/header limits, DB outage, schema mismatch, 401/403/429/5xx policies |
| frontend resilience | PASS | loading/error/empty paths, cancellation/retry tests, and Chrome desktop/mobile Playwright paths pass locally |
| accessibility | PARTIAL | semantic/component/static tests plus Playwright/Axe serious/critical checks pass; manual AT, zoom/reflow, high-contrast and independent WCAG review remain blocked |
| observability | PARTIAL | structured logs, request IDs, local metrics/health; external backend, alerts and SLOs absent |
| abuse protection | PARTIAL | bounded local limiter and trusted-proxy fix; distributed backend and production tuning absent |
| container hardening | PARTIAL | API/worker non-root/read-only/cap-drop; dev web/Postgres exceptions and no CVE scan |
| load/performance | PARTIAL | bounded local baselines and query plans; no sustained/production-equivalent test |
| backup/restore | PARTIAL | checksum, migration-manifest fingerprint, non-empty rejection, clean restore and schema validation pass locally; no PITR/off-site drill |
| incident response | PARTIAL | runbook exists; rota, paging, contacts and exercise absent |
| dependency security | PARTIAL | npm audit and govulncheck clean for reachable code; CI gates are configured but have not run on a protected repository and independent review remains blocked |
| deployment/rollback | PARTIAL | runbooks and fail-closed configuration exist; no staging/production provider rehearsal |

## Strict readiness scorecard

| Category | Score | Reason |
| --- | ---: | --- |
| architecture | 8.5/10 | strong modular boundaries and embedded schema contract; operational providers remain abstract |
| security | 7.5/10 | strong repository controls; no independent pentest, full SAST/secret/image scan, or production edge validation |
| privacy/data governance | 7.0/10 | consent and retention architecture is credible; legal decisions and live deletion operations unvalidated |
| authentication/MFA | 8.0/10 | robust local lifecycle/concurrency; external delivery and production MFA operations blocked |
| recommendation trust | 9.0/10 | deterministic, versioned, reproducible, provenance-governed, commerce-independent |
| commerce/affiliate | 7.0/10 | provider-neutral and verified-data model is strong; no live partner certification |
| database/migrations | 8.0/10 | checksum/readiness/rollback/scale evidence; no production-size lock/replica rehearsal |
| resilience/performance | 6.5/10 | useful local baseline and fault tests; no soak, chaos platform, limits, or production topology |
| observability/operations | 5.5/10 | local primitives and runbooks; external metrics/logs/alerts/on-call absent |
| frontend/accessibility | 7.0/10 | defensive async UX and component tests; real browser/AT matrix absent |
| backup/disaster recovery | 6.0/10 | local verified restore and corruption rejection; no off-site/PITR/failover proof |
| production readiness overall | 6.5/10 | repository foundation is credible; external operational controls remain launch blockers |

## Failure-injection classification

`Automatic recovery` below means the local process/container behavior was
observed or covered by an executable test. It does not imply an orchestrator or
on-call service exists.

| Failure | Evidence | Behavior | Recovery | Status |
| --- | --- | --- | --- | --- |
| PostgreSQL unavailable/restart | live isolated container stop/start | liveness stays up; readiness and dependent work fail closed and remain bounded | pool reconnects after database return; traffic admission is an operator/orchestrator concern | PASS locally |
| connection pool exhausted | one-slot PostgreSQL integration test | waiting caller exits at its deadline; no unbounded hang | next request succeeds after lease release | PASS |
| slow query | `pg_sleep` beyond configured statement limit | PostgreSQL cancels and classifies timeout | connection remains usable | PASS |
| worker termination/restart | Phase 7 SIGTERM and lease-recovery tests | cycle cancels; durable job/lease state is retained | normal restart/expired-lease recovery; hard-kill orchestration remains external | PARTIAL |
| API termination/restart | Phase 7 in-flight drain and Compose restart | stops admission, bounded drain, forced-close fallback | restart requires orchestrator/operator | PARTIAL |
| migration failure/schema mismatch | rollback test and reversible extra-ledger live drill | transaction rolls back; API readiness closes | corrective migration or authorized restore/operator action | PASS locally |
| backup failure | duplicate-name/integrity/archive checks | command exits nonzero and does not overwrite artifact | operator must fix storage/credentials and rerun | PARTIAL; disk-full/cloud failure not tested |
| restore failure/corrupt input | truncated archive and non-empty target drills | rejected before destructive restore | clean target and verified artifact required | PASS locally |
| alert provider unavailable | disabled-adapter tests/readiness | service is visibly degraded; delivery is never claimed | operator must configure external provider | PASS locally; BLOCKED externally |
| rate-limit backend unavailable | adapter failure tests | protected request returns 503; limiter never fails open | backend recovery/operator action | PASS locally; distributed adapter BLOCKED |
| provider unavailable/malformed response | disabled/malformed adapter and bounded-job tests | no offer/conversion is fabricated; failure is durable and bounded | scheduled/manual retry or operator intervention | PASS for adapters; live providers BLOCKED |
| malformed/duplicate/expired webhook | signature, freshness, replay and idempotency tests | rejected or acknowledged without duplicate fact | valid provider retry where applicable | PASS |
| analytics database failure | HTTP handler service-error regression | generic 503; internal error is not leaked | database recovery then client retry under policy | PASS |
| local image storage failure | adapter outage test | save/open return an error; no false success | storage/operator recovery | PASS locally |
| frontend API unavailable/timeout/401/403/429/500 | API client/query-policy tests plus Playwright 429/500 flows | bounded error/cancellation, no infinite retry, auth state invalidated on 401 | user retry or reauthentication; optional surfaces remain honest | PASS locally |

## Phase result and gate

Phase 9 result: **PARTIAL**, not PASS and not a repository failure. The local
Chrome desktop/mobile suite completed with 21 passing tests and three
intentionally skipped duplicate project scenarios. It includes a 320, 375, 390,
430, 768, 1024, 1280, 1440, and 1920 px overflow matrix. A fresh PostgreSQL 17 database
applied all 17 migrations, accepted the idempotent seed twice, passed serialized
integration and race suites, and completed a clean backup/restore drill. CI and
security workflows are checked in but have not executed in a protected remote
repository.

Production-equivalent infrastructure, provider credentials, legal decisions,
independent penetration/accessibility testing, off-site/PITR recovery,
distributed abuse protection, durable object media, and external observability
are unavailable. It is not safe to call the application production-ready,
enable live financial providers, or send public traffic based on this evidence.
