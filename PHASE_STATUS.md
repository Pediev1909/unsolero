# UNSOLERO Phase 0 Status

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
| Authentication | PARTIAL | Argon2id, opaque hashed sessions, expiry/idle expiry, secure cookie policy, generic login failures, and protected routes exist; recovery, verification, and MFA do not | Identity module, auth adapter, HTTP middleware | No for current scope | Email/identity provider for expansion | Unit, HTTP, and PostgreSQL identity tests | High | PARTIAL |
| Authorization | PARTIAL | Server-enforced admin role and per-user repository predicates exist; role provisioning is manual and only coarse `admin`/`user` roles exist | Identity/admin/planning/recommendation modules | No | Operator IAM process | Authorization HTTP and PostgreSQL tests | High | PARTIAL |
| Database migrations | COMPLETE | Ten checksum-verified, transactional, advisory-locked migrations apply to a fresh PostgreSQL 17 volume | `backend/migrations`, migration runner | No new migration | PostgreSQL 17 | Fresh-volume migration and full DB integration suite | High | COMPLETE |
| PostgreSQL integration-test isolation | PARTIAL | Cleanup callbacks previously ran after the pool was closed and silently leaked transient rows | All PostgreSQL integration tests | No | None | Full integration suite; post-suite counts remain 30 products, 90 offers, 0 users/events/clicks | Medium | COMPLETE |
| Recommendation evidence governance | MISSING | Scores and facts have no source provenance, reviewer approval, or freshness/version chain | No code changed | Future schema likely | Editorial/data operations | Not available | Critical to trust | MISSING |
| Merchant and affiliate operations | PARTIAL | Offers, safe redirects, ownership, freshness, disclosure, and click recording exist; no production merchant refresh/conversion import, retry idempotency, or bot classification | Commerce module | Future schema may be needed | Merchant/affiliate providers | Unit and PostgreSQL integration tests | High | PARTIAL |
| Analytics and reporting | PARTIAL | Typed consent-aware events and observed admin metrics exist; verified conversion, revenue, EPC, CAC, LTV, and repeat-user reporting do not | Analytics module/admin UI | Future schema/reporting possible | Verified providers and consent policy | Unit and PostgreSQL reporting tests | High | PARTIAL |
| Search | PARTIAL | Parameterized PostgreSQL catalog search/filter/sort/pagination works; no typo tolerance, ranking model, synonym governance, or dedicated search service | Catalog repository/UI | No | None for current scale | Catalog unit/integration tests | Medium | PARTIAL |
| SEO | PARTIAL | Metadata, canonical links, robots, sitemap, editorial structured data, and noindex rules exist; acquisition pages remain client-rendered | Frontend SEO and content HTTP handlers | No | Prerender/SSR deployment decision | Unit tests and live sitemap/robots smoke tests | High | PARTIAL |
| Media handling | PARTIAL | Uploads are bounded, MIME-detected, random-named, path-safe, and immutable-served; local storage is not multi-replica, scanned, transformed, or CDN-backed | Admin media handler/local image adapter | No | Object storage/CDN/scanner for production | Store and handler tests; container volume smoke | High | PARTIAL |
| Rate limiting | PARTIAL | Bounded in-memory route buckets protect auth, recommendation, analytics, affiliate, and mutation traffic; counters are not shared across replicas | HTTP middleware/config | No | Distributed edge/rate-limit service if scaled | Middleware tests | High | PARTIAL |
| API versioning | PARTIAL | Versioned liveness/readiness exist, but feature endpoints remain under unversioned `/api` and legacy redirect remains active | HTTP router/API docs | No | Client migration plan | Router and smoke tests | Medium | PARTIAL |
| CI/CD | MISSING | No checked-in CI workflow or deployment pipeline exists | None | No | CI runner, registry, deployment target | Local release gate only | High | MISSING |
| Security headers and same-origin protection | COMPLETE | API emits CSP, framing, MIME, referrer, permissions, opener/resource, cache, request-ID, and conditional HSTS headers; unsafe cross-origin mutations are rejected | HTTP middleware; production Nginx config | No | TLS/reverse proxy in deployment | Middleware tests and live header smoke | High | COMPLETE |
| Logging and error handling | COMPLETE | Structured request logs, request IDs, panic recovery, bounded generic API errors, and no secret/destination logging found | HTTP/platform code | No | External log sink for deployment | HTTP tests and live smoke | Medium | COMPLETE |
| Docker development architecture | COMPLETE | Compose config/build/up, healthy PostgreSQL/API, migrations, seed, frontend, and isolated-network smoke checks pass | Existing Docker/Compose files | No | Docker | Compose, migration, seed, readiness, API/UI smoke | Medium | COMPLETE |
| Dependency vulnerability scanning | COMPLETE | npm audit reports zero vulnerabilities; govulncheck reports zero reachable/imported-package vulnerabilities | No dependency files changed | No | npm registry and Go vulnerability DB | `npm audit`; `govulncheck` | Medium | COMPLETE |
| Mobile and accessibility verification | PARTIAL | Mobile-first responsive components and accessible interaction tests exist; no real-browser matrix at all requested widths, screen-reader audit, or automated axe gate exists | No code changed | No | Browser test/assistive-tech environment | Component tests only | High | PARTIAL |
| Performance and caching | PARTIAL | Route splitting, lazy images, catalog/content cache policy, and an approximately 80.6 KiB gzip entry chunk are present; no load test, RUM, CDN, or production Core Web Vitals evidence exists | No code changed | No | Deployment/observability services | Production build only | Medium | PARTIAL |
| AI provider integration | BLOCKED | Provider-neutral validated boundary and disabled provider exist; no live provider is registered and the model cannot access repositories | AI module/config | No | Chosen provider, model, credentials, privacy review, evals | Mock/validation tests | High | BLOCKED |
| Production backup, monitoring, TLS, domains, and secret delivery | BLOCKED | Requirements are documented but no external deployment environment or operating evidence was supplied | `PRODUCTION_READINESS.md` | No | Hosting, DNS, TLS, secret manager, backup and telemetry systems | Local checks cannot prove these | Critical | BLOCKED |
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
