# UNSOLERO pre-launch scorecard

Status: PARTIAL — Phase 10 repository validation; public production launch is forbidden
Last reviewed: 2026-08-17

## Phase 10 launch matrix

Allowed statuses are `PASS`, `PARTIAL`, `BLOCKED`, `EXTERNAL`, `LEGAL`, and `NOT TESTED`. Classes are A (implement now), B (configure in staging), C (external infrastructure), D (external provider), E (legal/business), and F (independent validation).

| Requirement | Evidence | Status | Owner | Next action | Class |
| --- | --- | --- | --- | --- | --- |
| Distributed rate limiting | atomic shared Redis, TTL, concurrency and outage tests | PASS | Backend | deploy TLS/auth Redis and exercise failover | B |
| Trusted proxy topology | exact-CIDR parsing and spoof regression tests | PARTIAL | Platform/security | validate real ingress hops | B |
| Private S3-compatible media | ownership, conditional-create, type/digest and outage tests | PASS | Backend | deploy private bucket/KMS/IAM/versioning | B |
| Malware/content scanner | fail-closed port only | BLOCKED | Security/platform | select and validate scanner | D |
| Media deletion/orphans | durable retry covers known failed/deleted objects | PARTIAL | Backend/platform | add provider inventory reconciliation for crash-window orphans | A |
| Transactional email | SMTP contract/TLS/header tests; no provider | BLOCKED | Platform/security | approve provider/domain and sandbox delivery | D |
| Central telemetry | redacted logs and bounded authenticated OpenMetrics export | PARTIAL | SRE | collector, retention, dashboards and alerts | C |
| Delivered alerts/on-call | notifier fails closed; no destination/rota | BLOCKED | SRE | provision and exercise paging | C |
| Fresh database | 18 migrations from empty PostgreSQL; seed passed twice | PASS | Database | retain protected CI gate | A |
| Recommendation/commercial isolation | deterministic unit/integration invariance tests | PASS | Recommendation | retain mandatory protected gate | A |
| Bounded collections | wishlist/setup/catalog/admin pages stable and bounded | PASS | Backend | representative EXPLAIN/load evidence | B |
| API 404 semantics | unknown API routes return 404 | PASS | Backend | retain smoke test | A |
| Public-page HTTP semantics | SPA shell still returns 200 for unknown browser paths | PARTIAL | Web/platform | choose SSR/prerender/edge routing | A |
| Performance | bundle budgets defined; no field/staging latency | PARTIAL | Performance | automate Lighthouse and staged load/soak | B |
| Browser/accessibility automation | current Chrome desktop/mobile, 320–1920 px and Axe suite: 21 passed, 3 intentional duplicate-project skips | PASS | QA | retain protected CI gate | A |
| CI/security gates | pinned workflows configured but not remotely executed | NOT TESTED | Engineering/security | run in protected repository | B |
| Disaster recovery | local logical tooling only; timed production-like exercise absent | NOT TESTED | Platform/database | approve RPO/RTO and run witnessed exercise | C |
| Live providers | all remain disabled | BLOCKED | Provider owners | complete activation checklist | D |
| Legal/business approval | inventory exists; no approval claimed | LEGAL | Legal/business | approve applicable artifacts/markets | E |
| Independent penetration/accessibility validation | not performed | BLOCKED | Security/accessibility | commission against approved staging | F |

No row may be promoted from this document alone. Evidence must identify an immutable release and environment.

## Phase 9 historical matrix

Statuses are exactly `PASS`, `PARTIAL`, `BLOCKED`, `EXTERNAL`, `LEGAL`, or
`NOT TESTED`. Blocker class is exactly A (repository), B (local validation plus
production infrastructure), C (external provider), D (legal/business), or E
(independent validation).

| Requirement | Current state | Evidence | Status | Owner | Next action | Dependency | Class |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Production config fail-closed | Critical URL/TLS/cookie/key/provider/media invariants are code-tested | config package; `PRODUCTION_CONFIGURATION.md` | PASS | Engineering | retain regression gate | none | A |
| Production deployment | No reviewed orchestrator, ingress, secret manager, network or promotion environment | `DEPLOYMENT.md` | EXTERNAL | Platform | provision and validate staging/production | hosting and secret manager | B |
| Distributed rate limiting | Clean fail-closed port; local adapter single-process; multiple replicas rejected | abuse package; `ABUSE_PROTECTION.md` | BLOCKED | Platform/backend | select Redis-compatible service, implement/test adapter against deployed TLS/auth/failover | distributed store | B |
| Durable media | Local atomic/scoped/signature-validated store plus object-storage/scanner ports; production deliberately refuses local/dev providers | admin media ports; localimages/imagescan tests | PARTIAL | Platform/backend | select object store and scanner, add adapters and integration/failure tests | storage/scanner accounts | C |
| Central logs/metrics | Redacted structured logs, bounded route/counter metrics, request IDs and readiness exist; no central sink | observability package/doc | PARTIAL | SRE | select exporters, retention and dashboards | monitoring platform | C |
| Delivered alerts | Alert port and failure visibility exist; external provider unavailable | alerting package; alert matrix | BLOCKED | SRE | integrate destination and exercise paging | alert provider/on-call | C |
| Backup/restore repository tooling | Dump+metadata integrity, migration fingerprint and safe restore implemented | scripts; `DR_READINESS.md` | PASS | Database engineering | keep local restore regression | PostgreSQL 17/Docker | A |
| Production DR | No off-site immutable backup, PITR or timed production-like exercise | `DR_READINESS.md` | EXTERNAL | Platform/database | provision and run recorded exercise | managed database/object/KMS | B |
| Core CI workflow | Frontend, backend, race, DB, Compose and browser gates configured | `.github/workflows/ci.yml` | NOT TESTED | Engineering | run in protected repository and fix failures | GitHub Actions/config | B |
| Security CI workflow | gitleaks, Semgrep, Trivy, image pin and SPDX SBOM configured with immutable action SHAs | `.github/workflows/security.yml` | NOT TESTED | Security | review pinned commits and execute | CI network/tool registries | B |
| Browser automation | Chrome desktop/mobile Playwright paths, recommendation/full-account/consent/error/affiliate/keyboard flows, 320–1920 px overflow matrix, and Axe serious/critical checks passed against a fresh seeded stack | `frontend/e2e`; 21 passed, 3 deliberately skipped duplicate project scenarios | PASS | Frontend/QA | retain as a protected CI gate and archive failures | CI browser runner | A |
| Independent accessibility assessment | Automated checks do not prove WCAG compliance | `PRODUCTION_VALIDATION.md` | BLOCKED | Accessibility/product | manual AT/browser audit by qualified reviewer | independent reviewer | E |
| Pagination bounds | Catalog/admin/evidence/import pages bounded and SQL ties stabilized; legacy wishlist/setup lists remain unpaginated | application/repository tests and review | PARTIAL | Backend | paginate user collections before material scale | API compatibility plan | A |
| Analytics/report scale | Windows/batches bounded; process metrics cardinality allowlisted; production query plans/capacity unmeasured | analytics/observability code | PARTIAL | Backend/data | EXPLAIN at representative volume and load test | production-like dataset | B |
| Commerce provider lifecycle | Disabled/unknown/missing adapter failures, audit history and approval model exist; no real adapter active | commerce tests; provider checklist | PASS | Commerce engineering | retain disabled state | none | A |
| First live commerce provider | No credentials, agreement, sandbox certification or operational approval | provider checklist | EXTERNAL | Commerce/legal/security | complete checklist for chosen provider | provider/account/agreement | C |
| Recommendation/commercial isolation | Type and integration tests keep commission, offers, clicks, conversions and revenue outside scoring | recommendation invariance tests | PASS | Recommendation owner | mandatory regression gate | PostgreSQL for integration | A |
| Authentication/authorization | Opaque sessions, Argon2id, MFA, recovery, route/ownership checks and rate limits implemented | account/security tests and review | PARTIAL | Security/backend | independent pentest and real email delivery validation | provider + assessor | C |
| Repository security review | Request flows reviewed and local tests/scans run where available; external scanners and ingress unavailable | `SECURITY_VALIDATION.md` | PARTIAL | Security | close scan findings and commission pentest | CI/staging | E |
| Capacity/performance | Bounded load harness exists; no representative production infrastructure or traffic | `LOAD_TESTING.md` | NOT TESTED | Performance/platform | approve and run staged capacity test | production-like infra/data | B |
| Legal/business launch package | Approval inventory exists; no legal conclusion or approval claimed | `LAUNCH_GOVERNANCE.md` | LEGAL | Legal/business | approve applicable artifacts and jurisdictions | counsel/business decisions | D |
| Incident/on-call operation | Runbook exists; no real rota, alert destination or exercise | `INCIDENT_RESPONSE.md` | EXTERNAL | Security/SRE | assign rota and exercise staging incident | people + alert platform | B |
| Independent penetration test | Not performed | `SECURITY_VALIDATION.md` | BLOCKED | Security | commission scoped test and remediate findings | independent assessor/staging | E |

The browser suite and bounded catalog/admin pagination close their Phase 9
class-A scope. Wishlist and saved-setup collection pagination remains PARTIAL;
silently truncating those user-owned collections would be worse than preserving
the current complete response contract. No score increase is earned merely by
the existence of this document or an unexecuted workflow.
