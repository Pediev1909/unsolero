# UNSOLERO phase status

## Full production-readiness execution — repository closure

Audit date: 2026-08-18
Verdict: **PARTIAL — repository and isolated-local controls pass; production launch remains blocked.**

| Area | Status | Evidence | Remaining gate |
| --- | --- | --- | --- |
| Frontend application | PASS | format, typecheck, lint, 50 unit tests, production build, transfer/entry budgets, 21 Playwright scenarios at 320–1920 px | hosted Lighthouse/RUM and independent manual accessibility review |
| Go API and workers | PASS | unit/integration/race/vet/build/module verification; durable worker checkpoint and protected operational metrics | hosted capacity, failover and central telemetry |
| PostgreSQL | PARTIAL | 20 fresh migrations, idempotent fictional seed, clean logical restore, tested least-privilege grant template | managed TLS/PITR/HA, production roles, off-site encrypted restore evidence |
| Redis and object storage | PARTIAL | isolated Redis and private S3-compatible integration suites pass | managed private TLS/auth, KMS/IAM, versioning/inventory and recovery evidence |
| Authentication/authorization | PASS repository | server-side sessions, Argon2id, MFA and permission tests pass | deployment topology and independent penetration testing |
| Recommendation trust boundary | PASS | deterministic tests and architecture controls exclude commerce inputs | governed real product evidence before public recommendations |
| CI and supply chain | PARTIAL | immutable-SHA workflows, dependency review, all-role scan/SBOM matrix and digest-based release-candidate workflow defined | protected hosted run, registry digest, signature/provenance verification |
| Monitoring | PARTIAL | authenticated bounded metrics include pool, worker, jobs, media and backup state | collector, dashboards, retention and hosted query/load evidence |
| Alerting | BLOCKED | fail-closed authenticated webhook adapter and rule tests exist | approved destination, actual delivered/acknowledged alert and on-call owner |
| Backups and DR | PARTIAL | checksum logical backup/clean restore succeeds; age encryption handoff is checked in | install/authorize age identity, encrypted off-site schedule, PITR/failover/rollback and measured RPO/RTO |
| Vercel/frontend edge | UNKNOWN | no Vercel project, team, domain or environment evidence is available | identify account/project/domain and preserve same-origin API plus genuine route-level 404 semantics |
| Legal/privacy/business | BLOCKED | engineering policy/runbooks exist; no approval is claimed | accountable human and qualified legal/privacy review |
| Production deployment | BLOCKED | no authorized hosted production-equivalent control plane or immutable deployed candidate | provision, deploy, observe, recover, validate and approve before public traffic |

No live financial, affiliate, conversion, email, AI, scanner or alert provider
was activated. The sole human operator remains the owner of every external
decision until explicitly delegated. The current strict evidence score is
**8.0/10 for repository readiness**, not production approval. The final verdict
is **NOT PRODUCTION READY**.

## Phase 12 — hosted staging and external launch gates

Audit date: 2026-08-18
Verdict: **PARTIAL — no hosted environment or external approval was available; production launch remains rejected.**

| Requirement | Current status | Evidence | Owner | Remaining action | Final status |
| --- | --- | --- | --- | --- | --- |
| Managed production-equivalent staging | BLOCKED | no authorized cloud control plane or credentials | Platform owner — unassigned | provision reviewed isolated staging through IaC | BLOCKED external |
| Immutable image deployment/digests | BLOCKED | no registry, signing identity or deployment platform | Release owner — unassigned | build once, sign, attest and promote by registry digest | BLOCKED external |
| Central telemetry and delivered alerts | BLOCKED | bounded sources exist; no collector or delivery receipt | SRE/on-call owner — unassigned | deploy collector/dashboards and exercise acknowledged alerts | BLOCKED external |
| Hosted load/soak/browser/accessibility | BLOCKED | Phase 11 local regression evidence only | Performance/quality owners — unassigned | execute against hosted candidate with fictional representative fixtures | BLOCKED external |
| Backup/PITR/failover/rollback/RPO/RTO | BLOCKED | Phase 11 checksum local restore only | Database/SRE owners — unassigned | prove managed recovery and measure approved objectives | BLOCKED external |
| Hosted supply-chain execution | BLOCKED | workflows are SHA-pinned; repository access/run artifacts unavailable | Release/security owners — unassigned | authorize protected CI, scan digest, retain SBOM/provenance | BLOCKED external |
| Provider sandboxes | BLOCKED | every live provider remains disabled/fail-closed | Product/commerce/security owners — unassigned | approve and exercise sandbox-only adapters | BLOCKED external |
| Legal/business approval | BLOCKED | no qualified approval or accountable launch owner | Counsel/business owners — unassigned | complete the gate register in Phase 12 evidence | BLOCKED external |
| Independent validation | BLOCKED | no assessor, report or witnessed exercise | Security/accessibility/DR owners — unassigned | commission and close independent findings | BLOCKED external |
| Production readiness claim | NOT APPLICABLE | Phase 12 has no evidence supporting approval | Accountable executive — unassigned | keep public traffic disabled | NOT APPROVED |

The authoritative owner/evidence/date/action register is
[`docs/PHASE_12_EVIDENCE.md`](./docs/PHASE_12_EVIDENCE.md). No provider,
infrastructure, legal, capacity, security or recovery evidence was fabricated.

## Phase 11 — remaining engineering closure and local staging execution

Audit date: 2026-08-18
Verdict: **PARTIAL — local engineering blockers were reduced; public production launch remains rejected.**

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Media crash reconciliation | COMPLETE locally | bounded dry/apply service, strict keys, audit rows and CLI | admin/storage/CLI/docs | `000019` | managed inventory semantics | unit, PostgreSQL, MinIO, dry-run | High | PASS repository |
| Real public 404/index policy | COMPLETE locally | edge resolver, canonical/noindex and TLS smoke | HTTP/Nginx/scripts/docs | No | crawler validation | HTTP + real edge | High | PASS repository |
| Cross-process telemetry | PARTIAL | durable PostgreSQL gauges plus replica pool/HTTP/OpenMetrics | observability/backup/API | `000019` checkpoint | collector/history/dashboards | unit + authenticated scrape | High | PASS source; external blocked |
| Alert delivery | BLOCKED | provider-neutral rules; disabled adapter fails honestly | alerting/docs | No | alert destination/on-call | unit only | Critical | BLOCKED external |
| Performance gates | PARTIAL | bundle budgets, bounded API gates, query plans, 30 s soak | frontend/loadtest/scripts/CI | No | Web Vitals/RUM/representative scale | local staging | High | PASS gates; capacity unknown |
| Dependency/CI security | PARTIAL | deterministic scans and SHA/digest checks | workflows/scripts | No | hosted CI/image/SBOM evidence | local constituents | High | remote execution pending |
| Production-shaped local staging | COMPLETE locally | 2 API, 2 worker, TLS, shared DB/Redis/MinIO, limits | Compose/Docker/config | No | managed services/secret manager | build/up/fault/browser | Critical | PASS local only |
| Failure/recovery/DR | PARTIAL | dependency and replica faults; backup/clean restore; short soak | scripts/docs | No | HA/PITR/long soak/real alerts | controlled exercises | Critical | local PASS; external blocked |
| Live providers/legal/independent review | BLOCKED | all providers disabled; no claims fabricated | configuration/docs | No | contracts, counsel, assessors | not available | Critical | BLOCKED |

The exact evidence and limitations are recorded in
[`docs/PHASE_11_EVIDENCE.md`](./docs/PHASE_11_EVIDENCE.md). No live provider or
real customer/financial record was used.

## Phase 10 — staging parity and blocker ownership

Audit date: 2026-08-17
Verdict: **PARTIAL — repository controls improved; public production launch remains rejected.**

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Distributed limiter | PARTIAL | Atomic Redis Lua decision, TTL, namespace, HMAC keys, separate route policies, readiness and fail-closed outage | abuse/config/API/Compose/CI | No | TLS/auth Redis and real proxy topology | unit, concurrency and isolated Redis integration | Critical | PASS repository; B/C external |
| Private media | PARTIAL | S3-compatible conditional create, product ownership, signature/digest validation, private reads and outage handling | media adapters/composition/admin | No | managed private bucket, KMS/IAM and scanner | unit/service/isolated MinIO integration | Critical | PASS repository; B/D external |
| Media deletion recovery | PARTIAL | durable lease/retry/dead jobs process known failed/deleted objects | admin repository/service/worker | `000018` | provider inventory for crash-window orphan sweep | unit and PostgreSQL integration | High | PARTIAL class A |
| Transactional email | PARTIAL | provider-neutral SMTP, STARTTLS policy, safe token links and security notices | email/config/identity/API | No | approved provider/domain/credentials/complaint handling | protocol/config/service tests | Critical | PASS contract; BLOCKED class D |
| Observability | PARTIAL | redacted logs and authenticated bounded JSON/OpenMetrics process snapshot | observability/API/docs | No | collector, storage, dashboards, alerts/on-call | unit and HTTP authorization/format tests | High | PARTIAL classes B/C |
| User collection scale | COMPLETE | wishlist/setup endpoints page 1–10,000, size 1–100, stable indexed order and frontend aggregation | planning/recommendation/API/frontend | `000018` indexes | representative query/load evidence | unit, PostgreSQL and frontend schemas/build | Medium | PASS repository; B scale proof |
| Staging/CI parity | PARTIAL | digest-pinned PostgreSQL/Redis/MinIO; CI starts isolated adapter dependencies | Compose/workflows/docs | No | protected GitHub and hosted staging | Compose config/build and local constituent checks | High | NOT TESTED remotely |
| Public HTTP routing/SSR | PARTIAL | API 404 is genuine; unknown SPA paths still return shell 200 | frontend/Nginx audit/docs | No | rendering/deployment decision | API/browser route checks | Medium | PARTIAL class A |
| DR/capacity/accessibility/security validation | BLOCKED | runbooks/budgets exist without production-equivalent or independent evidence | documentation | No | hosted infrastructure and assessors | local evidence only | Critical | classes B/C/F |
| Legal/provider launch | BLOCKED | no live provider or launch approval | provider/launch governance docs | No | providers, agreements, counsel, owners | disabled/fail-closed tests | Critical | classes D/E |
| Production readiness claim | NOT APPLICABLE | Phase 10 explicitly forbids a launch claim | readiness/audit/scorecard | No | every applicable gate | evidence reconciliation | Critical | NOT APPROVED |

Phase 10 added migration `000018_phase10_collection_indexes.sql`, direct
dependencies `github.com/redis/go-redis/v9` and `github.com/minio/minio-go/v7`,
and no live provider connection. Detailed current blockers are in
`docs/PRE_LAUNCH_SCORECARD.md`.

## Historical Phase 9 — controlled pre-launch blocker closure

Audit date: 2026-08-17
Verdict: **PARTIAL — repository/local blockers were reduced; public production launch remains rejected.**

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Production configuration | PARTIAL | Production rejects non-public origins, insecure database/cookies, local media, unsafe scanners and unsupported adapters | config, composition root, environment docs | No | Production secret manager and topology | config/unit and startup validation | Critical | COMPLETE in repository; BLOCKED externally |
| Distributed rate limiting | PARTIAL | Provider-neutral boundary and fail-closed external selection exist; local limiter is limited to one replica | abuse platform/config/docs | No | Reviewed distributed/edge provider | middleware/config/failure tests | High | BLOCKED class B |
| Durable media | PARTIAL | Product-scoped deterministic atomic storage and scanner ports exist; production refuses local/development adapters | admin media ports, local storage/scanner, config | No | Object store/CDN/scanner | unit/service/config tests | High | BLOCKED class B/C |
| Observability | PARTIAL | Structured logs, bounded metric names, readiness degradation and alert matrix exist | observability and operations docs | No | Central telemetry/alert provider/on-call | unit/live fault tests | High | BLOCKED class B |
| Disaster recovery | PARTIAL | Local manifest-fingerprinted backup restored into a clean PostgreSQL 17 target; populated target rejected | backup/restore scripts and runbooks | No | Encrypted off-site backup/PITR/staging | shell validation and live drill | Critical | PASS locally; BLOCKED class B |
| CI/security automation | PARTIAL | CI and security workflows cover frontend, Go, database, Compose, browser, audit, SAST, secrets, image and SBOM gates | `.github/workflows` and CI docs | No | Protected GitHub environment/registry | local constituent commands only | High | NOT TESTED remotely |
| Browser/accessibility automation | PARTIAL | Chrome desktop/mobile paths, full account lifecycle, tracked redirect, and 320–1920 px overflow matrix pass with Axe serious/critical checks | Playwright config and `frontend/e2e` | No | CI browser runner; independent AT reviewer | 21 passed, 3 deliberately skipped duplicate project scenarios | High | PASS locally; class E remains |
| Pagination/scale | PARTIAL | Catalog/admin list requests are bounded and stable; user wishlist/setup collections retain their legacy complete-response contract | catalog/admin services and PostgreSQL adapters | No | API contract migration for user collections | unit/integration suites | Medium | PARTIAL |
| Provider activation | BLOCKED | Provider-neutral adapters fail closed and an activation checklist records evidence required | provider checklist and composition roots | No | Credentials, contracts, sandbox certification | disabled/missing-provider tests | Critical | BLOCKED class C |
| Security review | PARTIAL | Route ownership, redirect, SQL, upload, secret and dependency boundaries were locally reviewed; independent testing absent | security validation and code fixes | No | Independent assessor/staging | unit/integration/race/audit/browser | Critical | PARTIAL; class E remains |
| Legal/business launch package | BLOCKED | Decision inventory exists without legal conclusions | launch governance | No | Counsel/business owners/jurisdiction | Document review only | Critical | BLOCKED class D |
| Production readiness claim | NOT APPLICABLE | Phase 9 explicitly does not authorize public traffic or live financial providers | readiness/audit/scorecard docs | No | All external gates above | Evidence reconciliation | Critical | NOT APPROVED |

Problems discovered during validation were fixed: recommendation DTO list fields
now serialize as arrays; one-time verification is safe under React strict
effects; homepage bronze labels meet automated contrast thresholds; shared
PostgreSQL suites run serially against their shared fixture; and the backup
fingerprint query plus shell error propagation were corrected after the live
drill exposed them. No migration or live-provider dependency was added.

## Historical Phase 0 status

Audit date: 2026-08-17

This table records repository evidence, not a production-readiness declaration.
`BLOCKED` means an external system, credential, deployment decision, or operating
process is required; blocked features fail closed or remain undisplayed.

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
|---|---|---|---|---|---|---|---|---|
| Read required instructions and project documents | COMPLETE | `AGENTS.md`, `ARCHITECTURE.md`, `PRODUCTION_READINESS.md`, `FINAL_AUDIT.md`, `BUSINESS_MODEL.md`, and `README.md` inspected | Documentation only | No | None | Repository inventory | Low | COMPLETE |
| Public UNSOLERO branding | PARTIAL | The former product name remained in public UI, SEO, seed editorial copy, and documentation | Frontend copy/metadata, demo seed, public docs | No; seed is development-only and idempotent | None | Frontend tests; exact-case repository search | Low | COMPLETE |
| Lowercase `rigmark` compatibility identifiers | NOT APPLICABLE | Go imports/module, npm package, Compose/DB defaults, cookies/storage, asset paths, slugs, Nginx names, and health service identifier are internal compatibility keys | `README.md` documents classification | No | None | Case-insensitive repository inventory | High if renamed blindly | NOT APPLICABLE |
| Recommendation catalog pagination | COMPLETE | Service reads all pages up to a fail-closed 1,000-candidate bound | `backend/internal/modules/recommendation/application/service.go` | No | None | `TestGenerateLoadsTheEntireBoundedCatalog` | High | COMPLETE |
| Discontinued products in saved setups | COMPLETE | Setup reopening loads exact referenced IDs and overlays immutable stored prices | Recommendation service and catalog repository | No | None | `TestGetSetupLoadsReferencedArchivedProductsByID`; PostgreSQL suite | High | COMPLETE |
| Combined setup floor-area constraints | COMPLETE | Setup assembly tracks cumulative equipment footprint and rejects overflow | Recommendation domain engine | No | None | `TestSetupCannotExceedCombinedFloorArea` | High | COMPLETE |
| Affiliate recommendation ownership | COMPLETE | Redirect SQL requires authenticated ownership and clicked-product membership | Commerce PostgreSQL adapter and integration test | No | None | `TestAffiliateClickRequiresOwnedRecommendationAndFreshOffer` | Critical | COMPLETE |
| Stale-offer rejection | COMPLETE | Offer listing and redirects enforce configured `last_checked_at` maximum age | Commerce adapter/config and integration tests | No | None | Catalog stale-list assertion; fresh/stale redirect regression | High | COMPLETE |
| Analytics consent | COMPLETE | Optional browser events default off and require explicit `granted`; affiliate clicks remain essential server events | Frontend analytics and backend analytics validation | No | None | Consent banner, dispatcher, and API tests | High | COMPLETE |
| Accessible contrast | COMPLETE | Token overrides maintain a stronger muted-text/placeholder contrast floor | `frontend/src/styles.css` | No | None | Production CSS build; component suite | Medium | COMPLETE |
| Catalog indexing and metadata | COMPLETE | Canonical landing pages index; filtered/query variants noindex; product metadata omits invented ratings | Catalog SEO and product detail pages | No | None | `CatalogListing.test.ts`; metadata tests | High | COMPLETE |
| Trust separation for recommendations | COMPLETE | Recommendation inputs/candidates/config expose no commerce fields; package-import architecture test blocks commerce, analytics, and AI dependencies | Recommendation domain/application | No | None | Determinism, architecture, and scenario suites | Critical | COMPLETE |
| Authentication | COMPLETE | Argon2id; opaque hashed sessions; expiry/idle/revocation/cleanup; hashed one-time verification/reset; password change; export/deletion; encrypted TOTP; hashed recovery codes; generic request responses | Identity module, auth/email adapters, account HTTP/UI | `000016` | Reviewed external email adapter and production secrets remain deployment requirements | Unit, HTTP, PostgreSQL identity, cookie/origin/rate tests | High | COMPLETE at repository level |
| Authorization | COMPLETE | Explicit backend permissions for catalog, evidence, policy, commerce, content, analytics, users, and administration; production privileged step-up; ownership-derived export/deletion/session operations | Identity domain, HTTP middleware/routes, role-filtered admin navigation | `000016` | Staff provisioning and IAM operating process | Permission matrix, denial, forged-role, IDOR/session ownership, separation tests | High | COMPLETE at repository level |
| Database migrations | COMPLETE | Seventeen checksum-verified, transactional, advisory-locked migrations apply to a fresh PostgreSQL 17 volume | `backend/migrations`, migration runner | `000017` in Phase 6 | PostgreSQL 17 | Fresh-volume migration and full DB integration suite | High | COMPLETE |
| PostgreSQL integration-test isolation | PARTIAL | Cleanup callbacks previously ran after the pool was closed and silently leaked transient rows | All PostgreSQL integration tests | No | None | Full integration suite; post-suite counts remain 30 products, 90 offers, 0 users/events/clicks | Medium | COMPLETE |
| Recommendation evidence governance | MISSING | Scores and facts have no source provenance, reviewer approval, or freshness/version chain | No code changed | Future schema likely | Editorial/data operations | Not available | Critical to trust | MISSING |
| Merchant and affiliate operations | PARTIAL | Offers, safe redirects, ownership, freshness, disclosure, and click recording exist; no production merchant refresh/conversion import, retry idempotency, or bot classification | Commerce module | Future schema may be needed | Merchant/affiliate providers | Unit and PostgreSQL integration tests | High | PARTIAL |
| Analytics and reporting | PARTIAL | Typed consent-aware events and observed admin metrics exist; verified conversion, revenue, EPC, CAC, LTV, and repeat-user reporting do not | Analytics module/admin UI | Future schema/reporting possible | Verified providers and consent policy | Unit and PostgreSQL reporting tests | High | PARTIAL |
| Search | PARTIAL | Parameterized PostgreSQL catalog search/filter/sort/pagination works; no typo tolerance, ranking model, synonym governance, or dedicated search service | Catalog repository/UI | No | None for current scale | Catalog unit/integration tests | Medium | PARTIAL |
| SEO | PARTIAL | Metadata, canonical links, robots, sitemap, editorial structured data, and noindex rules exist; acquisition pages remain client-rendered | Frontend SEO and content HTTP handlers | No | Prerender/SSR deployment decision | Unit tests and live sitemap/robots smoke tests | High | PARTIAL |
| Media handling | PARTIAL | Uploads are bounded, MIME-detected, random-named, path-safe, and immutable-served; local storage is not multi-replica, scanned, transformed, or CDN-backed | Admin media handler/local image adapter | No | Object storage/CDN/scanner for production | Store and handler tests; container volume smoke | High | PARTIAL |
| Rate limiting | PARTIAL | A provider-neutral boundary protects auth, recommendation, analytics, affiliate, and mutation traffic. The bounded local adapter is single-replica only; external selection fails closed until implemented | Abuse platform, HTTP middleware/config | No | Distributed edge/rate-limit service before scaling | Middleware, config and backend-outage tests | High | PARTIAL externally |
| API versioning | PARTIAL | Versioned liveness/readiness exist, but feature endpoints remain under unversioned `/api` and legacy redirect remains active | HTTP router/API docs | No | Client migration plan | Router and smoke tests | Medium | PARTIAL |
| CI/CD | MISSING | No checked-in CI workflow or deployment pipeline exists | None | No | CI runner, registry, deployment target | Local release gate only | High | MISSING |
| Security headers and same-origin protection | COMPLETE | API emits CSP, framing, MIME, referrer, permissions, opener/resource, cache, request-ID, and conditional HSTS headers; unsafe cross-origin mutations are rejected | HTTP middleware; production Nginx config | No | TLS/reverse proxy in deployment | Middleware tests and live header smoke | High | COMPLETE |
| Logging and error handling | COMPLETE | Structured request logs, request IDs, panic recovery, bounded generic API errors, and no secret/destination logging found | HTTP/platform code | No | External log sink for deployment | HTTP tests and live smoke | Medium | COMPLETE |
| Docker development architecture | COMPLETE | Compose config/build/up, healthy PostgreSQL/API, migrations, seed, frontend, and isolated-network smoke checks pass | Existing Docker/Compose files | No | Docker | Compose, migration, seed, readiness, API/UI smoke | Medium | COMPLETE |
| Dependency vulnerability scanning | COMPLETE | npm audit reports zero vulnerabilities; govulncheck reports zero reachable/imported-package vulnerabilities | No dependency files changed | No | npm registry and Go vulnerability DB | `npm audit`; `govulncheck` | Medium | COMPLETE |
| Mobile and accessibility verification | PARTIAL | Mobile-first responsive components and accessible interaction tests exist; no real-browser matrix at all requested widths, screen-reader audit, or automated axe gate exists | No code changed | No | Browser test/assistive-tech environment | Component tests only | High | PARTIAL |
| Performance and caching | PARTIAL | Route splitting, lazy images, catalog/content cache policy, and an 87.54 KiB gzip main entry chunk are present; no load test, RUM, CDN, or production Core Web Vitals evidence exists | Frontend routing also lazy-loads the admin shell | No | Deployment/observability services | Production build only | Medium | PARTIAL |
| AI provider integration | BLOCKED | Provider-neutral validated boundary and disabled provider exist; no live provider is registered and the model cannot access repositories | AI module/config | No | Chosen provider, model, credentials, privacy review, evals | Mock/validation tests | High | BLOCKED |
| Production backup, monitoring, TLS, domains, and secret delivery | BLOCKED | Local executable backup/restore and privacy-safe telemetry/alert boundaries now exist, but no external deployment environment, durable schedule/storage, telemetry backend, or operating evidence was supplied | Production and operations documentation | No | Hosting, DNS, TLS, secret manager, backup and telemetry systems | Local recovery/health/failure drills only | Critical | BLOCKED externally |
| No fabricated product/customer/commercial data | COMPLETE | Seed is explicitly fictional and contains no users, reviews, conversion, commission, revenue, or customer activity claims | Demo seed/docs | No | None | Seed counts and PostgreSQL integration suite | Critical | COMPLETE |

## Compatibility-reference classification

- **Public branding corrected:** visible UI, page titles, social/SEO metadata,
  editorial author/copy, setup descriptions, and public project documentation.
- **Intentionally retained internals:** `module rigmark` and Go imports,
  `@rigmark/web`, Compose/local database defaults, `rigmark_session`, existing
  `rigmark:*` browser keys, `rigmark-api`, Nginx identifiers, test-only domains,
  demo tracking/slugs, and the existing hero-image filename.
- **Migration decision:** no schema/data migration was justified. New seed runs use
  UNSOLERO copy; changing previously seeded development rows requires rerunning
  the idempotent development seed. Production customer data is not touched.

## Phase 1 — Product evidence and recommendation governance

This section supersedes the Phase 0 `MISSING` evidence-governance row. It does
not change the overall controlled-staging-only decision.

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
|---|---|---|---|---|---|---|---|---|
| Evidence sources and classifications | MISSING | No runtime source model existed | Evidence domain/application/adapter/HTTP | `000011` | None | Domain validation; PostgreSQL lifecycle | Critical | COMPLETE |
| Dated observations, freshness and confidence | MISSING | Catalog facts had no observed/expiry/confidence chain | Evidence schema and repository | `000011` | Operational source policy for real data | Publication and expiry validation | Critical | COMPLETE |
| Product fact revisions | MISSING | Recommendation-critical values were mutable projection columns only | Evidence schema/module; catalog projection | `000011` | None | Provenance completeness; integration publication | Critical | COMPLETE |
| Score revisions and rationale | MISSING | Scores were unversioned and unsupported | Evidence schema/module | `000011` | Human scoring rubric for real data | Eight-dimension rationale coverage | Critical | COMPLETE |
| Approval/publication workflow | MISSING | Generic admin status mutation could activate products | Evidence service/repository/routes; admin editor | `000011` | At least three trained staff accounts for real operation | State transition, separation-of-duties, authorization | Critical | COMPLETE |
| Unapproved data excluded from public recommendations | PARTIAL | Catalog status filtered rows but had no revision gate | Catalog PostgreSQL adapter | `000011` | None | Fresh migration; publication lifecycle; source-revocation regression; catalog integration | Critical | COMPLETE |
| Recommendation replay snapshots | PARTIAL | Constraints, prices, scores and reasons persisted, but the complete candidate fact/score revision universe did not | Recommendation domain/port/adapter/service | `000011` | Retain engine release identified by `engine_version` | Repository lifecycle; saved setup regression | Critical | COMPLETE for current engine inputs |
| Versioned recommendation policy | PARTIAL | String policy version existed without a relational policy record | Recommendation policy schema | `000011` | Policy change/evaluation process | FK integration and deterministic suite | High | COMPLETE for `home-gym-v1` |
| Commercial incentive isolation | COMPLETE | Existing package boundary blocked commerce imports | Recommendation/evidence architecture tests and PostgreSQL invariance test | No additional | None | Type reflection, import architecture, commission/priority mutation invariance | Critical | COMPLETE |
| AI cannot publish canonical data | PARTIAL | AI had no repository, but no evidence publication module existed | Evidence architecture and role-gated routes | No additional | Human review policy | Evidence import architecture test; authorization tests | Critical | COMPLETE |
| Explicit fictional demo evidence | PARTIAL | Products were labeled fictional but had no evidence record | Demo seed and evidence schema | `000011` | None | Seed counts and provenance counts | Critical | COMPLETE |
| Admin provenance inspection | MISSING | No source/revision/provenance view existed | Admin evidence routes, schemas, queries and pages | `000011` | `admin` role | Frontend type/lint/build; API repository integration | High | COMPLETE |
| Public evidence distinctions | MISSING | Public product pages could not label fact origin | Catalog DTO/schema and product detail UI | `000011` | Real reviewed sources for non-demo use | Frontend schema/build; catalog integration | High | COMPLETE |
| Real manufacturer/merchant/testing evidence population | BLOCKED | Repository contains no supplied real-world evidence and fabrication is prohibited | No real records added | No additional schema | Manufacturer documents, lab results, verified merchant feeds, reviewers | Cannot be truthfully tested without source inputs | Critical | BLOCKED |
| Evidence operations and alerts | BLOCKED | No staff roster, freshness SLA, withdrawal runbook or notification service was supplied | Documentation only | Future only if job state is needed | Operations ownership and alert destination | External operational exercise | High | BLOCKED |

### Phase 1 dependency and migration summary

- Migration added: `backend/migrations/000011_product_evidence_governance.sql`.
- Runtime dependencies added: none.
- External providers/services added: none.
- Demo seed remains opt-in and now creates one verified, explicitly fictional
  source with product observations, complete fact links, and eight score
  rationales per product.

### Phase 1 verification record

- Fresh PostgreSQL 17 volume: all 11 migrations applied; 30 fictional products
  seeded with governed fact/score pointers, 326 fact-provenance links, 240 score
  rationales, one verified fictional source, and one policy version.
- Go: formatting verification, `go test ./...`, `go test -race ./...`,
  `go vet ./...`, and `go build ./...` passed in Go 1.25 containers.
- PostgreSQL: all nine adapter packages passed serial integration tests against
  the fresh volume.
- Frontend: formatting, TypeScript, ESLint, 20 Vitest files / 40 tests, and the
  production Vite build passed.
- Runtime: readiness and governed catalog/product-detail smoke tests passed; the
  detail response exposed 18 classified evidence records and both revision IDs.

## Phase 5 — Complete account security

Phase 5 is complete at repository level. This does not change the overall
controlled-staging-only decision.

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
|---|---|---|---|---|---|---|---|---|
| Email verification and resend | PARTIAL | Registration existed without verification | Identity service/repository, email port, auth HTTP/UI | `000016` | Live provider blocked | One-use, expiry, anti-enumeration, live smoke | High | COMPLETE at repository level |
| Development email behavior | MISSING | No honest local delivery mechanism | Development intent sink and gated inspection endpoint | No additional | None | Adapter and live verification/reset smoke | High | COMPLETE |
| Password reset and change | MISSING | No reset/change lifecycle | Identity security service, PostgreSQL transactions, account UI | `000016` | Live provider blocked for reset delivery | One-use, validation, global/other-session revocation | Critical | COMPLETE at repository level |
| Session inventory and revocation | PARTIAL | Opaque sessions existed without account controls | Identity adapter/service, account API/UI, worker cleanup | `000016` | None | Ownership, individual/other revocation, cleanup | High | COMPLETE |
| Account export and deletion | MISSING | No self-service data controls | Identity adapter/service, account API/UI | `000016` | Legal retention approval | Cross-account integration tests and live deletion/export smoke | High | COMPLETE at repository level |
| TOTP MFA and recovery | MISSING | No second factor existed | AES-GCM/TOTP adapter, identity service/repository, login/account UI | `000016` | Production key manager | Unit, recovery replay, scoped challenge, live MFA smoke | Critical | COMPLETE at repository level |
| Least-privilege authorization | PARTIAL | Administration relied on a coarse role | Permission matrix, route gates, role-filtered navigation | `000016` | Staff provisioning process | Matrix, forged-role, denial, recent-MFA tests | Critical | COMPLETE at repository level |
| Immutable security audit history | MISSING | Security-sensitive events were incomplete | Append-only PostgreSQL event log and service instrumentation | `000016` | Retention policy and external monitoring | Mutation rejection, cleanup retention, log/API secret scan | High | COMPLETE at repository level |
| Production email transport | BLOCKED | No provider, credentials, domain, templates, bounce handling, or delivery evidence supplied | Fail-closed `external` configuration boundary only | No additional yet | Reviewed transactional email provider | Cannot be truthfully completed locally | Critical | BLOCKED |
| External key rotation and global abuse control | BLOCKED | Local AES key version is fixed and rate limits are replica-local | Documented boundary | Future adapter/infrastructure work | Secret manager/KMS and trusted edge | Deployment exercises | High | BLOCKED |

### Phase 5 dependency and migration summary

- Migration added: `backend/migrations/000016_complete_account_security.sql`.
- Runtime dependencies added: none.
- External providers/services connected: none.
- Production email delivery remains deliberately unavailable until a reviewed
  adapter is linked; development records delivery intent and never reports mail
  as sent.

### Phase 5 verification record

- Frontend formatting, TypeScript, ESLint, 21 Vitest files / 42 tests, production
  build, and the production dependency audit passed; npm reported zero
  vulnerabilities.
- Go formatting, `go test ./...`, `go test -race ./...`, `go vet ./...`, and
  `go build ./...` passed in official Go 1.25 containers.
- `govulncheck` v1.7.0 reported zero reachable or imported-package
  vulnerabilities; one required-module advisory has no called symbols.
- A fresh PostgreSQL 17 volume applied all 16 migrations. The development seed
  succeeded twice, and all nine PostgreSQL adapter packages passed sequentially.
- Compose configuration, every image build, startup, liveness/readiness, the
  Vite service, and the complete live verification/reset/session/MFA/admin/export/deletion
  smoke flow passed. API and worker logs contained no error-level records or
  searched secret fields.

## Phase 6 — Analytics privacy and data governance

Phase 6 passes every available repository-level check. This is a technical
privacy-control result, not a legal-compliance or whole-product production-readiness
declaration.

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
|---|---|---|---|---|---|---|---|---|
| Analytics data inventory | PARTIAL | Collection existed without a complete classification, purpose, access, or retention inventory | Governance and analytics documentation | No | Qualified legal/privacy review remains external | Repository-wide source/schema audit | High | COMPLETE at repository level |
| Client event identity and deduplication | MISSING | Browser retries had no durable event identity | Analytics frontend, service, PostgreSQL adapter | `000017` | None | UUID validation; concurrent eight-request integration test; unique constraint | High | COMPLETE |
| Server-authoritative consent | PARTIAL | Browser storage was authoritative and the request asserted its own consent state | Analytics HTTP/service/repository and consent UI | `000017` | Approved production notice and policy version | Pre-consent, grant, withdrawal, stale version, persistence, and live smoke tests | Critical | COMPLETE at repository level |
| Privacy-safe event schemas | PARTIAL | Event names were typed, but arbitrary sensitive property attempts needed a fail-closed boundary | Analytics service and transport | No additional | None | Exact schema and forbidden-field regression matrix | Critical | COMPLETE |
| Anonymous identity claiming | MISSING | No auditable authenticated association lifecycle existed | Analytics domain/service/repository/HTTP/frontend | `000017` | Product/legal decision whether claiming remains enabled | Idempotent claim, concurrent/cross-user/revoked/deleted-user tests | High | COMPLETE |
| Configurable retention | MISSING | Analytics events and ingestion metadata had no explicit expiry worker | Analytics schema/repository, worker, configuration | `000017` | Legal approval and operational alerting | Bounded cleanup, idempotence, index plans, live worker smoke | Critical | COMPLETE at repository level |
| Account export and deletion integration | PARTIAL | Phase 5 account controls did not export consent/events or fully sever analytics claims | Identity repository/domain and account UI copy | `000017` | Legal decision for retained consent/security/audit evidence | PostgreSQL export/deletion/repeat-deletion and live lifecycle smoke | Critical | COMPLETE at repository level |
| Raw versus filtered reporting | PARTIAL | Raw accepted rows could enter reports without explicit reportability/coverage semantics | Analytics reporting repository/service, admin API/UI | `000017` | None | Filtered-report integration, no-data/partial-window HTTP tests, live report smoke | High | COMPLETE |
| Affiliate/privacy boundary | PARTIAL | Referrer handling and server analytics needed defense beyond the HTTP caller | Commerce application/HTTP/repository tests | No additional | None | Origin-only referrer, attribution preservation, secret/URL/UA/commission exclusion | Critical | COMPLETE |
| Security-event separation | COMPLETE | Security events already used a separate immutable schema | Identity permissions and HTTP authorization | No additional | Security retention/operations decision | Analyst/raw/consent/security permission matrix; immutable-event regression | Critical | COMPLETE |
| Least-privilege analytics administration | PARTIAL | Aggregate and event-level access shared a permission | Identity permissions, admin routes/navigation | No additional | Staff provisioning and access-review process | Endpoint authorization unit tests and analyst/admin live smoke | High | COMPLETE at repository level |
| Privacy-safe observability | PARTIAL | Panic recovery could log a raw panic value and referrer paths could carry unnecessary data | HTTP observability and affiliate normalization | No additional | External log-sink policy | Query/panic secret regression and live log inspection | Critical | COMPLETE at repository level |
| Privacy documentation | MISSING | No consolidated governance, analytics, or retention specification existed | `docs/DATA_GOVERNANCE.md`, `docs/ANALYTICS.md`, `docs/DATA_RETENTION.md` and core docs | No | Legal/product approvals remain external | Documentation cross-check against schema/config/code | High | COMPLETE at repository level |
| Production legal decisions and external processors | BLOCKED | No approved purposes, notices, regional policy, processor contracts, or legally reviewed retention schedule were supplied | Documentation only | Unknown until review | Qualified privacy counsel, product owner, security/operations owners | Cannot be proven by repository tests | Critical | BLOCKED |

### Phase 6 dependency and migration summary

- Migration added: `backend/migrations/000017_analytics_privacy_governance.sql`.
- Runtime dependencies added: none.
- External analytics providers connected: none.
- The browser analytics envelope moved to schema version 3. Older stored rows
  remain non-reportable because current consent cannot be proven retroactively.

### Phase 6 exact verification record

- Frontend: Prettier passed; TypeScript passed; ESLint passed with zero warnings;
  Vitest passed 21 files / 42 tests using `npm run test -- --run`; production
  build passed. The main entry is 286.85 kB raw / 87.54 kB gzip. Production npm
  audit reported zero vulnerabilities.
- Backend: `gofmt -l` returned no files; `go test ./...`, `go test -race ./...`,
  `go vet ./...`, and `go build ./...` passed. `govulncheck` found zero
  vulnerabilities in called or imported packages; one required-module advisory
  has no called symbols.
- PostgreSQL: a fresh PostgreSQL 17 volume applied all 17 migrations. The
  fictional seed succeeded twice and remained exactly 8 categories, 10 brands,
  30 products, and 90 offers, with zero users, analytics events, or conversions.
  The complete Go suite passed serially with every PostgreSQL integration test
  enabled.
- Query plans used `events_retention_idx`,
  `ingestion_receipts_retention_idx`, and
  `events_reportable_name_occurred_idx` for the cleanup/reporting predicates.
- Docker: Compose configuration and all image builds passed. PostgreSQL and API
  were healthy, migrations exited zero, the worker and frontend remained up,
  and liveness, readiness, and frontend requests succeeded.
- Live privacy smoke: normal-browser events returned `403` before consent and
  after withdrawal; grant returned `200`; accepted and duplicate submissions
  returned `204` but persisted one event plus accepted/deduplicated receipts.
  Bot requests produced payload-free `bot_filtered` receipts only.
- Live identity/account smoke: account consent sync, first claim, repeated claim,
  export, and deletion passed. Export included one consent decision and one
  claimed event. Deletion left a non-authenticating account tombstone, removed
  current consent and user links, revoked the claim, anonymized consent history,
  and allowed subsequent separately consented browser activity only as anonymous.
- Live authorization smoke: an ordinary account received `403` for aggregates
  and event-level data; an analyst received `200` for filtered aggregates and
  `403` for events; an administrator received `200` for event-level data.
- Live retention smoke: one expired event and one expired payload-free receipt
  were removed in a bounded worker cycle; logs reported counts only and contained
  no request bodies, credentials, user email, analytics payload, or destination.

## Phase 7 — Production infrastructure and operations hardening

Phase 7 passes at repository/local-system level. It does not approve a public
production launch: every row marked `BLOCKED` needs real infrastructure,
credentials, ownership, and operating evidence.

| Requirement | Current status | Evidence | Files affected | Migration required | Dependencies | Tests | Risk | Final status |
|---|---|---|---|---|---|---|---|---|
| Central production configuration | PARTIAL | Production values existed but HTTP/DB/worker/operations policy and stable limiter identity were incomplete | Config, composition roots, `.env.example`, readiness docs | No | Deployment secret manager | Production rejection and bounded-value tests | Critical | COMPLETE at repository level |
| PostgreSQL reliability | PARTIAL | Pool sizing existed; session statement/lock/idle timeouts and safe error classification did not | Database platform, API/worker/migrate/seed roots | No | Production capacity and load evidence | Unit and PostgreSQL runtime-setting integration tests | High | COMPLETE at repository level |
| Migration failure recovery | PARTIAL | Transaction and advisory lock existed without a direct rollback regression | Migration integration test and bounded runner context | No | PostgreSQL | Deliberately failing migration rolls back and remains unrecorded | Critical | COMPLETE |
| Logical backup and clean restore | MISSING | No executable repository tooling | Scripts, Compose tool profiles, Makefile, backup runbook | No | Production scheduler, durable encrypted storage and ownership | Non-root backup, checksum, clean restore, migration/count verification, closed-failure drills | Critical | COMPLETE locally; BLOCKED externally |
| Privacy-safe structured observability | PARTIAL | Request logs existed without central error redaction, route-value suppression, worker logging policy or metrics | Observability platform and HTTP middleware | No | Central log/metrics provider and retention policy | Redaction, route, metrics auth/privacy, live logs | Critical | COMPLETE at repository level |
| Operational metrics | MISSING | Product analytics could not serve as infrastructure metrics | Recorder abstraction and protected metrics endpoint | No | External multi-process collector | Aggregate/concurrency/auth endpoint tests and live smoke | High | PARTIAL; process-local only |
| External alerting boundary | MISSING | No notifier port or honest disabled behavior | Alerting platform, readiness, worker/rate-limit wiring | No | Reviewed notifier and on-call destination | Disabled/external selection and failure-threshold tests | Critical | COMPLETE boundary; BLOCKED delivery |
| Distributed abuse protection | PARTIAL | Fixed-window buckets were process-local and directly embedded in middleware | Abuse port/adapters, central config, middleware | No | Reviewed distributed/edge adapter | Backend-outage, raw-address, replica/config tests | Critical | COMPLETE boundary; BLOCKED multi-replica adapter |
| Liveness/readiness/degraded model | PARTIAL | Database readiness existed without dependency criticality | Health service, router, API composition | No | Deployment probe routing | Critical/optional dependency and live database-outage tests | High | COMPLETE |
| Worker reliability and shutdown | PARTIAL | Durable leases/retries existed; cycle deadlines, bounded batch configuration, repeated-failure alerting and shutdown proof did not | Worker command/config/Compose | No | Real provider operational tests | Cancellation, alert threshold, lease recovery, duplicate/idempotency tests and Compose SIGTERM | Critical | COMPLETE at repository level |
| API graceful shutdown | PARTIAL | Shutdown existed without forced-close fallback or in-flight regression | API command/tests/Compose | No | Deployment termination policy | Real-socket in-flight drain and Compose SIGTERM logs | Critical | COMPLETE |
| HTTP and frontend production hardening | PARTIAL | Server bounds existed; auth expiry and query retry policy were incomplete | Router/middleware, query client/providers, Vite | No | Browser E2E and load environment | Unit/type/lint/build and live smoke | High | COMPLETE at repository level |
| Container hardening | PARTIAL | API non-root existed; binaries shared one runtime and production Nginx was unpinned/root | Dockerfiles and Compose | No | Registry/image scanner | Complete builds; UID/read-only/capability and proxy smoke | High | COMPLETE at repository level |
| Deployment/operations runbooks | MISSING | Requirements were dispersed in readiness notes | `docs/DEPLOYMENT.md`, `docs/OPERATIONS.md` and core docs | No | Named operators and chosen platform | Documentation-to-code/drill cross-check | High | COMPLETE repository runbooks; BLOCKED operations |
| Supply-chain audit | PARTIAL | Lockfiles existed without this phase's scans/base-image pinning | Dockerfiles and documentation | No | CI scanner and registry policy | npm audit, govulncheck, Go checks and image builds | High | COMPLETE for this revision; BLOCKED automated gate |

### Phase 7 dependency and migration summary

- Database migrations added: none.
- Runtime/package dependencies added: none.
- External providers connected: none.
- Docker base images are immutable-digest pinned. Backend runtimes explicitly
  include CA roots and separate API, worker, migration, and seed artifacts.

### Phase 7 exact verification record

- Frontend: Prettier, TypeScript, zero-warning ESLint, 22 Vitest files / 45
  tests, and Vite production build passed. Largest entry: 287.11 KiB raw /
  87.61 KiB gzip. `npm audit --omit=dev --audit-level=high`: zero findings.
- Backend: `gofmt -l` empty; `go test ./...`, race suite, vet, and build passed.
  `govulncheck`: zero reachable/imported-package vulnerabilities; one
  required-module advisory has no reachable call.
- PostgreSQL: fresh 17 database, all 17 migrations, fictional seed, and complete
  PostgreSQL-enabled Go suite passed. Failed-migration rollback passed.
- Containers: Compose config, all default images, and the production Nginx
  target built. Fresh stack startup, web/API smoke, health, readiness, protected
  metrics, database outage/recovery, and API/worker graceful SIGTERM passed.
- Recovery: non-root custom-format backup and checksum passed; clean restore
  recovered 17 migrations and 30 fictional products. Duplicate backup and
  non-empty restore attempts failed closed.
- Not run/available: `shellcheck` is not installed. Syntax validation with
  `bash -n` passed. No load/browser/penetration/image scan, external alert,
  cloud backup, PITR, or production deployment was fabricated.

## Phase 8 — Production validation and resilience

Phase 8 result: **PARTIAL**. All reasonable repository and isolated-local-stack
validation passed after the defects below were corrected. External operational
and independent-assessment requirements remain blocked, so this phase does not
approve production readiness.

| Requirement | Prior status | Evidence / files affected | Migration | Dependencies | Tests | Risk | Final status |
|---|---|---|---|---|---|---|---|
| Formal validation matrix | MISSING | `docs/PRODUCTION_VALIDATION.md` records PASS/PARTIAL/BLOCKED scope and strict scores | No | None | Evidence cross-check | High | COMPLETE |
| Reproducible HTTP load baseline | MISSING | `backend/cmd/loadtest`, `scripts/load`, `docs/LOAD_TESTING.md` | No | None | Mock transport unit tests plus 11 live scenarios | High | COMPLETE locally; PARTIAL operationally |
| Database scale/query validation | MISSING | rollback-only `scripts/phase8-scale-validation.sql`; 10k users, 5k products, 100k events | No | Disposable PostgreSQL | Counts, plans, rollback | High | COMPLETE locally; PARTIAL operationally |
| Concurrency invariants | PARTIAL | identity and affiliate PostgreSQL integration tests | No | PostgreSQL | concurrent registration/reset/MFA/click tests; full race suite | Critical | COMPLETE at repository level |
| Release/schema compatibility | PARTIAL | embedded migration manifest and readiness checker | No | Explicit migration job | current/changed/missing/extra manifest tests and live 503 drill | Critical | COMPLETE |
| Trusted proxy/rate identity | PARTIAL | explicit `TRUSTED_PROXY_CIDRS`; forwarded headers ignored by default | No | Exact production ingress ranges | config and router spoof/trust tests | Critical | COMPLETE at repository level; BLOCKED topology |
| Query abuse resistance | PARTIAL | identity queries now match `users_email_unique` expression index | No | Production-volume rehearsal | scale `EXPLAIN ANALYZE` and auth tests | High | COMPLETE locally |
| Failure injection | PARTIAL | DB/schema/pool/statement/storage/analytics/provider/restore failures | No | Production chaos environment | unit, integration and live local drills | High | COMPLETE locally; PARTIAL operationally |
| Frontend resilience | PARTIAL | cancellation preservation, safe error and retry classification, safe new-window relation | No | Browser automation | 22 Vitest files / 49 tests; type/lint/build | High | COMPLETE at repository level; PARTIAL browser evidence |
| Backup/disaster recovery | PARTIAL | local checksum/clean restore plus corrupt/non-empty rejection and schema verification | No | Off-site storage, PITR, managed DB | timed local backup/restore drills | Critical | PARTIAL |
| Security/dependency validation | PARTIAL | `docs/SECURITY_VALIDATION.md`; npm audit, govulncheck, secret/sink review | No | SAST, image/SBOM/secret scanners, pentest | available scans passed | Critical | PARTIAL |
| Incident response | MISSING | `docs/INCIDENT_RESPONSE.md` | No | Pager, rota, contacts, exercise | Runbook review only | Critical | PARTIAL |

### Phase 8 change summary

- Database migrations added: none.
- Runtime/package dependencies added: none.
- External providers connected: none.
- Material defects fixed: untrusted forwarded-address rate-limit bypass,
  unindexed case-insensitive identity queries, release/schema mismatch not
  represented in readiness, misleading frontend cancellation errors, and one
  unsafe new-window relation.
- Exact commands and measured latency/query/restore evidence are recorded in
  the Phase 8 documents. No live revenue, commission, conversion, user, product,
  review, provider, browser, or operational evidence was fabricated.
