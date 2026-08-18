# Final CTO audit

## Full production-readiness execution addendum — 2026-08-18

Final verdict: **NOT PRODUCTION READY**.

The repository itself is now at **8.0/10 readiness**: frontend and backend
quality gates pass, the deterministic recommendation/commercial separation is
preserved, all 20 migrations apply on a fresh database, containerized database/
Redis/private-object-storage integration and race tests pass, and the backup can
be restored cleanly. New controls cover authenticated external alert delivery,
durable worker health, production dependency validation, PostgreSQL runtime
grants, encrypted-backup handoff, dependency review, SBOM/image scanning for
every deployable role, and digest-based release-candidate promotion.

That is not a production result. There is no identified Vercel project or
domain, authorized cloud control plane, deployed immutable candidate, managed
TLS PostgreSQL/Redis/object storage, secret manager, central telemetry, delivered
alert, off-site encrypted backup, PITR/failover/rollback measurement, long hosted
soak, provider sandbox certification, independent penetration report, manual
WCAG review, or legal/privacy/business approval. These are evidence gaps, not
paperwork gaps, and none is upgraded to PASS.

Current scores: architecture 9.1, UX 8.0, UI 8.0, mobile 8.2, security 8.7,
recommendation engine 9.3, monetization readiness 7.5, SEO 8.0, performance 8.0,
and repository production readiness 8.0. The public-launch gate remains failed.

Date: 2026-08-17  
Verdict: **controlled staging only; public production launch rejected**

## Phase 12 addendum — external execution audit

Phase 12 is **PARTIAL**. It did not provision hosted staging because no
authorized cloud account, registry, secret manager, managed data service,
telemetry collector, alert destination or provider sandbox was connected. The
Git remote identifies `Pediev1909/unsolero`, but the environment has no `gh`
executable and connected GitHub access returned repository `404`; hosted
workflow execution, artifacts, image scans and SBOMs therefore remain
unverified. Treating workflow definitions or local images as hosted evidence
would be false.

No immutable candidate digest, delivered alert, encrypted off-site backup,
PITR, managed failover, application rollback, measured RPO/RTO, multi-hour
hosted soak, Lighthouse result, provider certification, legal approval,
penetration report, independent accessibility report or witnessed DR record
exists. All are explicitly blocked with required owners and actions in
[`docs/PHASE_12_EVIDENCE.md`](./docs/PHASE_12_EVIDENCE.md).

The strict readiness score remains **7.8/10**. Phase 11 repository and local
staging controls remain useful, but Phase 12 produced no external evidence that
justifies a score increase or public launch.

## Phase 11 addendum — production-shaped local execution

Phase 11 is **PARTIAL**. The repository now closes the known media crash-window
and false-200 routing defects, exposes bounded cross-process operational source
metrics, enforces artifact and HTTP regression budgets, and provides a two-API,
two-worker TLS staging overlay with shared PostgreSQL, Redis, and private S3-
compatible storage. Controlled local faults, short soak, browser/accessibility,
backup/clean restore, routing, and reconciliation paths were exercised.

Validation found and fixed real defects: staging Nginx could not write cache
state as non-root; the Compose port merge removed TLS publication; the root
proxy appended browser paths to the resolver URI; internal SPA redirects lost
SEO headers; the validation tmpfs could not execute Go test binaries; a media
SQL CTE used a reserved keyword; the query-plan assertion missed bitmap index
nodes; the load probe counted its own deadline cancellations; and the viewport
matrix caused Chromium network churn by repeatedly creating pages. Final
container verification also found that the web health probe resolved
`localhost` incompatibly with the IPv4-only listener; it now targets the exact
loopback address and the rebuilt TLS edge reports healthy. An authenticated
metrics scrape then exposed route series collapsing to `unmatched` across a
context-copying timeout boundary; an inner mux capture now preserves only the
bounded registered pattern, with regression and staging scrape evidence.

The strict readiness score is **7.8/10**. Architecture 9.0, UX 7.8, UI 7.8,
mobile 8.0, security 8.4, recommendation engine 9.2, monetization readiness
7.2, SEO 7.8, performance 7.8, production readiness 7.8. This remains capped:
there is no hosted managed staging, central telemetry, delivered alert, secret
manager, managed HA/PITR/KMS/off-site backup, real email/scanner/merchant/
conversion/AI provider, Lighthouse/RUM or long capacity test, legal approval,
penetration test, or independent accessibility review.

## Phase 10 addendum — staging parity and blocker ownership

Phase 10 is **PARTIAL**. The repository now has exercised atomic
Redis-compatible rate limiting, private S3-compatible product media,
product-bound content-addressed object keys, durable bounded deletion retries,
a transactional SMTP boundary, bounded wishlist/setup pagination, authenticated
OpenMetrics output, pinned Compose dependency digests, and CI integration hooks
for Redis and MinIO. A fresh PostgreSQL 17 database applied all 18 migrations,
the fictional seed passed twice, and the complete serialized database/Redis/S3
integration suite passed after it exposed and drove a real timestamp-query fix.

The strict evidence-based production-readiness score is now **7.0/10**. This is
not approval. There is still no hosted staging/production topology, reviewed
scanner, real email or alert destination, centralized telemetry, managed
secrets/database/Redis/object storage, production-like capacity or DR exercise,
SSR/prerender routing, executed remote security workflow, legal approval,
independent penetration test, or independent accessibility assessment. All
merchant, affiliate-conversion, email, scanner, alert, AI, and other real
providers remain disabled.

Current category scores: architecture 8.5, UX 7.5, UI 7.5, mobile 7.5,
security 8.0, recommendation engine 9.0, monetization readiness 7.0, SEO 7.0,
performance 7.0, and production readiness 7.0. These scores cover repository
evidence only and are capped by the external blockers in
[`docs/PRE_LAUNCH_SCORECARD.md`](./docs/PRE_LAUNCH_SCORECARD.md).

## Historical Phase 9 addendum — controlled pre-launch blocker closure

Phase 9 is **PARTIAL**. Repository/local controls now include stricter
production-origin and media-provider validation, an application-owned media
storage/scanner boundary, bounded metrics, migration-manifest-fingerprinted
backup/restore, immutable-SHA CI/security workflow definitions, bounded/stable
catalog and operational pagination, and an executed Playwright/Axe gate. The
strict overall production-readiness score remains **6.5/10**; documentation and
unexecuted workflows earn no readiness credit.

Local validation found and fixed real defects in homepage contrast, one-time
email verification under React strict effects, recommendation list
serialization, PostgreSQL integration-suite isolation, and backup fingerprint
error propagation. No live provider, real financial record, public traffic,
legal approval, independent penetration test, WCAG certification, durable
object storage, distributed limiter, central telemetry, delivered alert, PITR,
or production RPO/RTO is claimed. The authoritative current blocker matrix is
[`docs/PRE_LAUNCH_SCORECARD.md`](./docs/PRE_LAUNCH_SCORECARD.md).

## Phase 8 addendum — production validation and resilience

Phase 8 is **PARTIAL**. The repository and isolated local stack passed all
available code, PostgreSQL, race, migration, load, failure, backup, restore,
dependency, and container-build checks. That evidence is substantial, but it
is not production-equivalent operational proof.

The audit fixed three material defects: rate-limit identity trusted forwarded
addresses without an explicit proxy boundary; authentication/recovery email
lookups missed the existing `lower(email)` index; and readiness did not prove
that the live database schema matched the exact application release. Forwarded
addresses are now ignored by default, identity queries use the expression
index, and the API embeds and verifies the full migration manifest without
applying DDL.

A dependency-free load probe, rollback-only scale fixture, targeted concurrent
identity/affiliate tests, database pool exhaustion/statement timeout tests,
storage and analytics failure tests, and frontend cancellation/retry tests were
added. Local baselines covered public and authenticated recommendations,
catalog, registration/login, admin, analytics, commerce, affiliate redirect,
and readiness. A 100,000-event/50,000-click/10,000-user synthetic database run
measured representative plans and rolled back all fictional rows. Database
outage, incompatible schema, corrupt backup, non-empty restore, storage failure,
disabled providers, and pool exhaustion all failed closed or recovered as
designed.

Public production remains rejected. There is no production-equivalent soak,
distributed limiter, external telemetry or alert delivery, on-call exercise,
independent penetration test, browser/assistive-technology matrix, image/SBOM
security gate, managed backup/PITR rehearsal, live provider certification, or
legal/privacy approval. The strict evidence matrix and scores are maintained in
`docs/PRODUCTION_VALIDATION.md`.

## Phase 7 addendum — infrastructure and operational hardening

The repository now has centralized fail-closed production configuration,
bounded PostgreSQL pool/session/migration timeouts, safe database error classes,
privacy-filtered JSON logs, registered-route request metrics, protected
process-local operational aggregates, explicit liveness/readiness/degraded
states, provider-neutral alerting and abuse-control boundaries, bounded worker
cycles, lease recovery, API/worker graceful shutdown, and executable non-root
PostgreSQL backup/clean-restore tooling.

The container layout is materially safer: API, worker, migration, and seed have
separate minimal targets; Go and Alpine bases plus production Nginx are digest
pinned; runtime processes are non-root; API/worker roots are read-only with
dropped capabilities; and the production web image was exercised as UID 101
with a read-only root. Frontend queries no longer retry ordinary 4xx failures,
and a 401 invalidates authenticated account/admin state.

The failure evidence is real and local: PostgreSQL outage changed readiness to
503 and recovered to 200; API and worker drained on Compose SIGTERM; an in-flight
HTTP request completed during shutdown; a failed migration rolled back; stalled
lease/idempotency/provider-outage/disabled-alert/rate-backend paths passed; and a
checksum-verified backup restored 17 migrations and 30 fictional products into
a clean database. The first backup drill exposed root-only operator artifacts;
that defect was fixed by running tools as an explicit non-root UID/GID.

Public production is still rejected. No deployment platform, managed database,
secret manager, distributed limiter, central logs/metrics/traces, functioning
alert destination, on-call owner, durable encrypted backup schedule, PITR,
container security gate, load evidence, or incident exercise exists. Repository
mechanics are not production operations.

## Phase 5 addendum — complete account security

Phase 5 is complete at repository level. Identity now provides anti-enumerated
registration, verification and password-reset requests; hashed, expiring,
single-use credentials; Argon2id password change/reset; absolute and idle
session expiry; session inventory/revocation; structured export; anonymizing
deletion; encrypted TOTP; hashed one-use recovery codes; scoped MFA login
challenges; immutable security events; and explicit backend permissions.

The backend derives identity, ownership, roles, verification, and MFA state from
the current hashed session and PostgreSQL on every protected request. Production
permission gates require a verified email and recent backend-recorded MFA for
every privileged role. Browser guards and role-filtered navigation are usability
controls only and cannot grant access.

A fresh 16-migration database, all Go/race/vet/build checks, all PostgreSQL
adapter packages, 42 frontend tests, production build, dependency scanners,
Compose build/start, and live verification/reset/session/MFA/admin/export/delete
flows passed. The live development adapter records delivery intent and never
claims email delivery.

This does not approve a public launch. No reviewed transactional email adapter,
sender-domain setup, bounce/complaint process, production secret manager, MFA
key-rotation adapter, distributed abuse control, staff provisioning process,
legal retention decision, independent penetration test, or production recovery
evidence exists.

## Phase 4 addendum — verified conversions and monetization metrics

The repository now has a provider-neutral, fail-closed conversion boundary for
authenticated webhooks and scheduled/manual imports. PostgreSQL owns request
replay protection, provider-event uniqueness, immutable verified facts,
normalized order/commission transitions, optional click attribution,
reconciliation history, and reporting coverage. Raw webhook bodies, credentials,
IP addresses, user agents, emails, and tokens are not stored in conversion
facts. Interrupted verified deliveries can resume; processed replays are safely
acknowledged; conflicting provider-event reuse is rejected.

Monetization reporting excludes pending, rejected, and reversed commission,
keeps provider currencies separate, and distinguishes available zero,
insufficient data, and no verified coverage. The protected commerce UI exposes
only real conversion/import/reconciliation state and displays **No data** while
all real providers remain disabled. Integration tests mutate commission, order,
provider, attribution, click, offer, and merchant data without changing the
deterministic recommendation output.

This is infrastructure, not revenue readiness. No Amazon Associates, Awin,
Impact, CJ, or direct adapter has an approved contract, credentials, sandbox
evidence, external alerting, or an operational reconciliation owner. No real
conversion, commission, or revenue data exists.

## Phase 3 addendum — merchant data and affiliate operations

The merchant-refresh and click-integrity findings are resolved at repository
foundation level. Commerce now has provider lifecycle/cursors, scheduled and
protected manual imports, database idempotency, immutable price/availability
observations, explicit expiry, partial-failure records, bounded retry, worker-lease recovery,
successful-snapshot-only reconciliation, operator status, and a disabled adapter
that fails closed without credentials.

Affiliate navigation resolves a fresh safe destination before best-effort
attribution writes. Duplicate request IDs are suppressed; obvious bots and
prefetches remain raw but do not enter filtered CTR/rankings; user agents are
hashed; recommendation-item ownership is persisted; and expired attribution is
anonymized. Tests prove tracking/session write failure does not block an already
resolved redirect and that merchant status, offer price/availability, priority,
or commission cannot change deterministic recommendation output.

Production is still rejected. No reviewed Amazon/Awin/Impact/CJ/direct adapter,
real credentials, external alert delivery, real conversion data, or staffed
reconciliation operation exists. A total database read outage still prevents
destination resolution.

## Phase 2 addendum — data-driven recommendation policy

The prior medium finding for category-slug-driven recommendation behavior is
resolved. Production engine code no longer assigns capabilities, requirements,
goal scores, roles, or redundancy from category switches. `fitness-v2` is an
explicit reviewed policy graph tied to published fact and score revisions.
Unknown categories and unconfigured products remain excluded, active behavior
is immutable, completed runs retain their policy-derived inputs, and activation
requires separate submitter, reviewer, and activator decisions.

Space handling now fails closed for required missing measurements and supports
known storage footprint, operating/safety clearances, room height, access width,
and explicit overlap zones. The commercial boundary remains structural and is
covered by a PostgreSQL test that changes affiliate priority/commission fields
without changing recommendation output.

This raises the architecture and recommendation-engine assessments to **9/10**.
It does not improve the unchanged launch blockers: real product evidence,
staffed governance, merchant refresh, staff MFA, SSR/prerendering, distributed
abuse controls, observability, recovery proof, browser accessibility testing,
and load/security validation remain outstanding.

## Phase 1 addendum — product evidence and recommendation governance

Phase 1 closes the audit's largest trust gap. Recommendation-critical product
facts and scores now have versioned fact/score revisions, per-field provenance,
dated observations, confidence, freshness, reviewer identity, per-score
rationale, independent approval/publication, and audit history. Catalog reads
fail closed unless both referenced revisions are published and fresh. Public
product pages distinguish verified facts, manufacturer claims, merchant
observations, editorial assessments, and explicitly fictional demo evidence.

Completed authenticated recommendations now store the commercial-free candidate
universe with fact and score revision IDs alongside the original constraints,
policy version, engine version, reasons, and fingerprint. Historical setup reads
restore the original recommendation inputs rather than substituting current
scores or dimensions. `home-gym-v1` remains registered for historical records;
new runs use the reviewed `fitness-v2` policy graph.

The recommendation trust boundary is structural: its domain inputs cannot
represent commission, sponsorship, payout, revenue, affiliate performance,
conversion, or click data; architecture tests block commerce/analytics/AI
imports; and an integration test changes affiliate commission and priority data
without changing deterministic output.

This does **not** approve production. The new governance system contains only
fictional demo evidence until operators supply and review real sources. Staff
MFA, evidence policy/runbooks, real-source ingestion, policy authoring tooling,
merchant refresh, recovery, deployment, and operational controls
remain release blockers.

## Executive verdict

This is not a pile of generated CRUD. The modular-monolith boundary is mostly real, authentication is materially better than average, the deterministic engine is isolated from commerce, and the application is unusually honest about demo data and missing revenue. That does not make it production-ready.

The weakest part of UNSOLERO is now operational execution around its trust model. The provenance and approval boundary exists, but it contains no real product evidence and has no staffed freshness/withdrawal operation. Live merchant, conversion, transactional-email, distributed-rate-limit, durable-media, alerting, and telemetry adapters remain unavailable without reviewed providers and credentials. Acquisition pages are a client-rendered SPA. Local Chrome desktop/mobile regression paths now pass, but manual assistive-technology and independent accessibility review remain absent. Those are launch blockers, not cleanup tickets.

The original audit covered 356 repository files, about 17,600 lines of Go, about 15,200 lines of frontend TypeScript/TSX, 43 Go test files, 20 frontend test files, and 10 SQL migrations. Phase 1 added the evidence module, tests, UI, and migration 11; generated dependencies and build output remain excluded from source review.

## Scorecard

| Area | Score | Brutal assessment |
| --- | ---: | --- |
| Architecture | **9/10** | Sound modular monolith with evidence and immutable data-driven policy boundaries plus reproducible run snapshots. Oversized adapters/handlers and API contract/versioning mismatch remain. |
| UX | **7/10** | Strong decision flow and honest states. Free text is non-functional, external email/merchant transitions are unavailable, and several production surfaces still expose demo assumptions. |
| UI | **8/10** | Restrained, coherent, and much better than generic SaaS. The design system is real. Legacy muted-color utilities required a global accessibility correction. |
| Mobile | **7/10** | Layouts are mobile-first and Chrome mobile regression paths pass locally, but comprehensive width/reflow/zoom, real-device, screen-reader, and touch evidence remains absent. |
| Security | **8/10** | Hashed one-time verification/reset tokens, Argon2id, revocable opaque sessions, encrypted TOTP, hashed recovery codes, immutable security events, permission gates, privileged step-up, origin checks, and ownership tests are implemented. Distributed abuse control, live email delivery, malware scanning, external key rotation, and an independent security gate remain. |
| Recommendation engine | **9/10** | Deterministic, revision-reproducible, evidence-gated, structurally commission-independent, and governed by explicit category/product policy data. Clearance/overlap modeling is conservative and only as complete as reviewed measurements; real evidence has not been populated. |
| Monetization readiness | **6/10** | Provider-neutral offers, verified conversion ingestion, reconciliation, coverage-gated metrics, filtered clicks, and operator controls exist. No live provider contract/adapter, credentials, external alerting, or real conversion data exists. |
| SEO readiness | **5/10** | Sitemap, canonicals, metadata, internal links, and editorial models exist. SPA rendering, soft 404s, unreliable unfurls, and no prerender/SSR prevent serious acquisition readiness. |
| Performance | **7/10** | Route splitting and image discipline are good; the production entry is 87.61 KiB gzip. Search and analytics storage have no representative load evidence or enforced performance budgets. |
| Production readiness | **6.5/10** | A credible hardened staging system, not a public service. Repository/local recovery, failure, browser, database, frontend, and container checks pass; external deployment, distributed controls, durable media, telemetry delivery, independent validation, legal approval, and staffed operations remain absent. This reconciles the audit with the unchanged strict Phase 8/9 readiness score; documentation alone earned no increase. |

## Severity model

- **CRITICAL**: exploitable compromise, destructive data loss, or a systemic violation of the recommendation/commission trust rule.
- **HIGH**: silently wrong core decisions, materially corrupt attribution, privacy/accessibility failure, or a major production trust defect.
- **MEDIUM**: public-launch blocker or scaling/design debt that is not currently an immediate exploit or silent core-data failure.
- **LOW**: maintainability, naming, polish, or localized quality debt.

## CRITICAL

No critical defect was proven. In particular, the audit found no plaintext password storage, raw session-token persistence, raw affiliate URL exposure in public catalog DTOs, commission input to the recommendation engine, unparameterized request data in SQL, or frontend access to AI credentials.

That is not a security certification. Dependency audit, penetration testing, production infrastructure, and live attack-surface validation were unavailable.

## HIGH — fixed in this audit

### H1. Recommendation generation silently considered only 100 products

**Problem:** `recommendation/application.Service.Generate` requested one catalog page with `Limit: 100`. The repository also clamps pages to 100. Every published product after that arbitrary prefix was invisible to the engine.

**Why it matters:** a product decision engine that silently ignores inventory can return a confidently wrong answer. Ordering happened before scoring, so this was not merely a capacity limit; it was hidden selection bias.

**Fix:** recommendation generation now pages deterministically through the complete published catalog, up to the engine's explicit 1,000-candidate bound. It probes beyond the bound and fails closed instead of truncating silently.

**Test:** added `TestGenerateLoadsTheEntireBoundedCatalog`, which verifies 101 products are loaded across two calls and all reach the engine.

### H2. Saved setups could silently lose products

**Problem:** reopening a setup loaded only the first 100 currently published products, then omitted result items absent from that map. A discontinued product or any referenced product outside the first page disappeared from the response.

**Why it matters:** saved recommendations are decision records. Mutating their visible contents after a catalog status change destroys user trust and auditability.

**Fix:** the recommendation catalog port now supports exact ID retrieval. The PostgreSQL adapter loads referenced products in bounded chunks and intentionally includes discontinued products. Missing referenced rows fail explicitly.

**Test:** added `TestGetSetupLoadsReferencedArchivedProductsByID`, including restoration of the stored recommendation price.

### H3. A multi-product setup could exceed the room footprint

**Problem:** eligibility checked whether each product fit the room independently. The optimizer could combine several individually fitting products whose total floor area exceeded the available room.

**Why it matters:** “fits your space” is a primary promise. Returning a physically impossible setup is a core recommendation failure.

**Fix:** the optimizer, alternative validator, and rejection explainer now enforce a conservative combined floor-area constraint. Rejected additions receive `setup.space_limit`.

**Test:** added `TestSetupCannotExceedCombinedFloorArea` with two 800 mm × 800 mm products in a 1 m² space.

### H4. Affiliate recommendation attribution could be forged

**Problem:** any syntactically valid recommendation UUID was accepted. A client could attach a click to another user's recommendation or to a recommendation that never contained the clicked product.

**Why it matters:** this corrupts future revenue-per-recommendation reporting and creates cross-account attribution contamination.

**Fix:** recommendation attribution now requires an authenticated user and an appropriate recommendation/setup surface. The redirect query verifies recommendation ownership and product membership before recording the click.

**Test:** added application tests for anonymous and unrelated-surface attribution. The SQL predicate is covered by the repository integration path when a test database is available.

### H5. Stale merchant offers remained clickable indefinitely

**Problem:** active/in-stock flags never expired. An offer checked months ago could still display and redirect as available.

**Why it matters:** stale price and availability data damages users and merchants, and turns an affiliate CTA into misinformation.

**Fix:** `OFFER_MAXIMUM_AGE` is now a validated server setting (72 hours by default). Both offer listing and redirect resolution fail closed for stale records. The configuration is wired through Compose and documented.

**Test:** added configuration-boundary coverage and an integration assertion that an effectively expired offer is omitted.

### H6. Optional product analytics ran without a consent decision

**Problem:** the frontend emitted analytics immediately and the backend stored `consent_state = unknown`. The architecture called the pipeline consent-aware, but the implementation was not.

**Why it matters:** this is a privacy and trust failure, especially for an EU-facing deployment. Documentation cannot substitute for a user decision.

**Fix:** optional events default off; an accessible preference surface records grant/decline/withdrawal against a versioned server-held decision. Client UUIDs, transactional consent checks, opaque subjects, and bounded retention supersede the former client-asserted consent field. Merchant-click attribution remains a disclosed, separate commerce consequence of an explicit outbound action.

**Test:** added consent-banner behavior, declined-event suppression, granted-envelope, backend validation, and transport mapping tests.

### H7. Muted text systematically failed WCAG contrast

**Problem:** small text used ink opacities from 25% to 60%. Against the canvas, measured contrast ranged from roughly 2.2:1 to 4.49:1, below the 4.5:1 threshold for normal text. Placeholder text also used 42% ink.

**Why it matters:** this affected disclaimers, helper text, empty states, specifications, admin data, and commerce disclosures—the exact copy users need to make safe decisions.

**Fix:** the design system now enforces a 65% ink floor for legacy muted-text utilities, raises the inverse 45% canvas utility, and raises placeholder contrast. The production CSS confirms the override is emitted.

**Test:** production build validated the CSS pipeline; automated browser contrast testing remains a medium-severity gap.

### H8. Core catalog pages were accidentally excluded from search and the default share image was broken

**Problem:** `/products` and `/brands/:slug` inherited `noindex`; filtered variants were not separated from canonical landing pages; and the metadata hook referenced a nonexistent `/images/hero-home-gym.webp`.

**Why it matters:** the primary acquisition inventory was telling crawlers not to index it, while social shares could request a 404 image.

**Fix:** clean product and brand landing pages are indexable, query/filter variants are `noindex, follow`, and the verified UNSOLERO hero asset is the default metadata image. Real, non-demo product pages now also emit conservative Product JSON-LD without invented ratings or offers.

**Test:** added crawl-policy and metadata-asset tests; verified the referenced WebP exists.

## MEDIUM — not fixed because they require product, policy, or architectural work

| Finding | Why it matters | Required action |
| --- | --- | --- |
| **RESOLVED IN PHASE 1:** Product suitability/quality scores had no evidence provenance | The engine could explain arithmetic but could not prove that an input score was deserved. | Versioned sources, observations, fact/score revisions, rationales, confidence, freshness, roles, workflow, public disclosure, and audit history are now implemented and tested. Real evidence population remains operational work. |
| **RESOLVED IN PHASE 2:** Recommendation capabilities and requirements were hard-coded by category slug | Adding a category previously required an engine code release and unknown categories could receive defaults. | Versioned category/product policies now explicitly gate support and provide capabilities, requirements, goal support, setup roles, redundancy, and spatial rules. Unknown/unconfigured records are excluded. |
| Free text is stored/fingerprinted but ignored | The conversational promise is ahead of the engine. | Keep the new disclosure; later add validated interpretation through the AI boundary or remove the field. |
| **RESOLVED IN PHASE 3:** No merchant-feed/import worker existed | Provider-neutral scheduled/manual imports, cursoring, immutable observations, freshness, reconciliation, bounded retries, lease recovery, and operator history are implemented. | Real reviewed adapters, credentials, external alerts, and provider runbooks remain operational requirements. |
| **RESOLVED IN PHASE 3:** Clicks lacked durable idempotency and bot/prefetch classification | Request-key idempotency, raw/filtered separation, ownership, bounded attribution, retention, and best-effort persistence are implemented. | Classification remains deliberately conservative and must be monitored with real traffic. |
| **RESOLVED IN PHASE 4:** No verified conversion or commission ingestion | Provider-neutral signed webhook/pull-import ingestion, replay protection, fact history, attribution, reconciliation, and verified-only metrics now exist. | Real provider contracts, credentials, sandbox evidence, and reconciliation ownership remain blocked; keep “No data” until verified records exist. |
| **RESOLVED IN PHASE 5:** Email verification, password recovery/change, session revocation, account export/deletion, and admin MFA were absent | Public accounts and privileged administration were incomplete. | Repository workflows, frontend surfaces, immutable events, TOTP/recovery codes, and production step-up now exist. A reviewed external email adapter remains blocked on provider selection and credentials. |
| **PARTIALLY RESOLVED IN PHASE 7:** Rate limits are process-local | The provider-neutral boundary and replica validation now prevent pretending the local adapter is horizontally safe. | Link and load-test a real distributed/edge adapter; external selection intentionally fails startup today. |
| **PARTIALLY RESOLVED IN PHASE 9:** Local image storage and signature-only MIME checks | Application-owned scanner/storage ports, magic-byte validation, deterministic product-scoped atomic writes, and production fail-closed configuration now exist. | Link reviewed durable object storage, malware/content scanning, transformation, and CDN delivery; no live adapter is claimed. |
| SPA-only acquisition pages | Metadata depends on client execution; unknown routes return the SPA shell with HTTP 200, creating soft 404s and unreliable social unfurls. | Add SSR or deterministic prerendering with real 404 status behavior. |
| Public API versioning is inconsistent | Health probes use `/api/v1`; application resources use unversioned `/api`; the target architecture still proposes different resource paths. | Freeze an actual contract, publish it, and introduce aliases/versioning before external consumers exist. |
| Catalog search is `%ILIKE%` without a search index | It will scan as catalog size grows. | Add measured PostgreSQL full-text/trigram indexing or a dedicated search adapter only when volume justifies it. |
| **RESOLVED IN PHASE 6:** Analytics lacked client IDs, server-authoritative consent, governed retention, and deletion/export integration | Concurrent retries could overcount and a browser consent string was not a security boundary. | Unique client IDs, transactional consent enforcement, payload-free receipts, filtered reporting, bounded cleanup, opaque identity claims, least privilege, and account export/deletion integration are implemented. Production purposes/retention still need qualified legal/privacy approval. |
| **RESOLVED IN PHASE 5:** Admin authorization was one coarse role | Coarse access increased blast radius. | Explicit catalog, evidence, policy, commerce, content, analyst, and administrator permissions now gate routes; production gates also require recent MFA. Domain separation-of-duties remains independent of route permission. |
| `is_demo` is inferred from a `demo-` slug prefix | SEO and disclosures depend on naming convention instead of a data invariant. | Add an explicit structured flag/provenance field through a migration and admin contract. |
| Homepage featured/example product content is hard-coded in the frontend | Replacing seed data can leave stale names and dead product links. | Drive production homepage product modules from reviewed content/catalog records with honest empty states. |
| **PARTIALLY RESOLVED IN PHASE 9:** No real browser/mobile/accessibility regression suite | Chrome desktop/mobile user paths, the requested 320–1920 px width matrix, and Axe serious/critical checks now pass locally. This does not prove zoom/reflow, high contrast, touch, or screen-reader behavior. | Retain Playwright in CI, add real-device/zoom/high-contrast smoke tests, and commission a manual WCAG 2.2 AA audit. |
| No production performance or load budget | A successful local build says nothing about p95 recommendation/search/redirect latency. | Define budgets, run representative load tests, collect Web Vitals, and enforce release thresholds. |
| **PARTIALLY RESOLVED IN PHASE 9:** No automated dependency/container/security release gate | Checked-in workflows now define dependency, secret, SAST, SBOM, image and digest gates with immutable action SHAs, but they have not run in a protected repository and the pinned commits have not received independent supply-chain review. | Protect the remote gate, review the pinned commits, execute it, retain evidence, remediate findings, and define reviewed exceptions. |

## LOW

- Several files are too large for comfortable ownership: the recommendation repository is 560 lines, catalog handler 522, AI validator 513, catalog repository 500, and recommendation handler 461. Split by use case when those areas next change; do not churn them solely for line count.
- Catalog/recommendation HTTP presenters repeat DTO mapping. A small presenter package or contract-generation strategy would reduce drift, but premature OpenAPI generation would add machinery without solving the current launch gaps.
- `draft`, `published`, and `discontinued` are the domain states while the admin language says “archive.” Pick one ubiquitous term.
- The accessibility compatibility override uses `!important` to repair legacy opacity utilities. Migrate call sites toward semantic muted tokens over time.
- Navigation still says “Sign in” for an already authenticated user on generic public headers. It should become Account without causing auth-loading layout shift.
- The legacy affiliate-link redirect remains live beside the offer redirect, increasing contract and reporting surface. Remove it after a measured deprecation window.
- The public design-system showcase ships as a route. Keep it `noindex` and exclude it from production navigation, or compile it only for non-production builds.
- A tiny root `package-lock.json` without a root package manifest is repository debris and should be removed after ownership is confirmed.

## Architecture assessment

The strongest decision is the modular monolith. HTTP handlers call services, services call ports, PostgreSQL stays in adapters, deterministic recommendation code imports no commerce data, affiliate resolution happens after ranking, and AI has no repository capability. Do not split this into services. There is no scale evidence that justifies distributed transactions and operational sprawl.

Phase 1 implemented the versioned chain from source facts → reviewed scoring inputs → policy version → result fingerprint. The next architectural investment is a separately versioned capability/compatibility policy and real evidence operations, not more AI. A deterministic engine can still be wrong when a category capability rule or a human-approved source is wrong.

## Verification record

### Passed after fixes

- Frontend TypeScript check.
- Frontend ESLint with zero warnings.
- Frontend Vitest: 20 files, 40 tests.
- Frontend production build; largest entry chunk about 264.5 kB raw / 80.6 kB gzip, CSS about 64.2 kB raw / 11.5 kB gzip.
- Docker Compose configuration validation using `.env.example`.
- Metadata asset existence and production CSS emission inspected directly.

### Added backend coverage

- Full-catalog recommendation pagination.
- Exact-ID retrieval for discontinued saved-setup products.
- Combined floor-area enforcement.
- Affiliate attribution validation.
- Stale-offer configuration and repository behavior.
- Analytics consent validation and transport mapping.

### Phase 1 verification

- `gofmt -l .`, `go test ./...`, `go test -race ./...`, `go vet ./...`, and
  `go build ./...` passed in official Go 1.25 containers.
- A fresh PostgreSQL 17 volume applied all 11 migrations; the governed fictional
  seed and all nine PostgreSQL adapter integration packages passed.
- Named Phase 1 tests passed for publication/withdrawal provenance, commercial
  data invariance, and the recommendation input trust boundary.
- API readiness plus governed catalog and product-detail responses passed live
  container smoke tests.
- Frontend formatting, type checking, lint, 20 test files / 40 tests, and the
  production build passed. The entry chunk is 265.38 kB raw / 80.78 kB gzip.
- No browser automation, real-device matrix, Lighthouse, load test, penetration
  test, or container scan was performed in Phase 1.

### Phase 2 verification

- Frontend formatting, TypeScript, ESLint, 20 Vitest files / 40 tests, and the
  Vite production build passed.
- `gofmt` verification, `go test ./...`, `go test -race ./...`, `go vet ./...`,
  and `go build ./...` passed in official Go 1.25 containers.
- All PostgreSQL adapter integration packages passed against a fresh PostgreSQL
  17 database with all 13 migrations and the idempotent fictional seed.
- The focused policy lifecycle, unsupported-category, revision-binding,
  compatibility, redundancy, spatial, historical-input, determinism,
  authorization, and commercial-invariance tests passed.
- Docker Compose configuration and all image builds passed. PostgreSQL and API
  readiness were healthy; live anonymous recommendation generation returned a
  complete result carrying `fitness-v2` and `deterministic-v2`.
- No browser automation, load test, penetration test, external monitoring,
  backup/restore drill, or production deployment was performed in Phase 2.

### Phase 5 verification

- Frontend formatting, TypeScript, ESLint, 21 Vitest files / 42 tests,
  production build, and `npm audit --omit=dev --audit-level=high` passed.
- The main entry chunk is 356.12 kB raw / 105.50 kB gzip; the admin shell is a
  lazy route chunk.
- Go formatting, `go test ./...`, `go test -race ./...`, `go vet ./...`, and
  `go build ./...` passed in official Go 1.25 containers.
- `govulncheck` v1.7.0 found zero reachable or imported-package
  vulnerabilities; its one required-module advisory has no reachable symbol.
- A fresh PostgreSQL 17 volume applied all 16 migrations, the fictional seed
  succeeded twice, and all nine PostgreSQL adapter packages passed serially.
- All Compose images built and the isolated stack started healthy. Liveness,
  readiness, web availability, registration/verification/reset/session/MFA,
  enforced admin MFA, export, deletion, retained audit, and log-secret smoke
  assertions passed.
- No live email delivery, browser automation, penetration test, load test,
  backup/restore drill, or public production deployment was performed.

### Phase 6 verification

- Phase 6 passes every available repository-level privacy/data-governance check;
  this does not establish legal compliance or whole-product production readiness.
- Frontend Prettier, TypeScript, ESLint, 21 Vitest files / 42 tests, production
  build, and production npm audit passed. The current main entry is 286.85 kB
  raw / 87.54 kB gzip; npm reported zero vulnerabilities.
- Go formatting, unit tests, race tests, vet, and build passed. The complete
  suite also passed serially with PostgreSQL integration tests enabled.
  `govulncheck` found zero called/imported-package vulnerabilities; one required
  module advisory has no called symbol.
- A fresh PostgreSQL 17 volume applied all 17 migrations. The fictional seed
  succeeded twice and created no users, analytics events, or conversions.
  Cleanup and reporting plans used the new retention/reportable indexes.
- The isolated Compose stack built and ran successfully. Live checks covered
  pre-consent rejection, grant, concurrent-safe deduplication, withdrawal,
  payload-free bot filtering, authenticated identity claiming, export,
  deletion/anonymization, post-deletion anonymous handling, aggregate/raw
  authorization separation, bounded cleanup, health/readiness, and frontend
  availability.
- Production consent wording/purposes, retention approval, processor contracts,
  regional requirements, external monitoring, backup deletion propagation, and
  privacy-request operations remain external blockers. No certification or
  compliance claim is made.

## Original Phase 1 report

### Implemented

- Added versioned evidence sources, observations, fact revisions, score
  revisions, rationale, review state, freshness, and audit history.
- Added three-person review/publication separation, immutable catalog projection,
  fail-closed withdrawal/expiry behavior, and source-label validation.
- Added commercial-free recommendation candidate snapshots, policy registration,
  historical result restoration, and commercial-invariance tests.
- Added admin provenance inspection, public evidence disclosure, and explicitly
  fictional governed development fixtures.

### Dependencies added

None.

### Database migrations added

`000011_product_evidence_governance.sql`.

### Problems found and fixed

The missing recommendation evidence chain, mutable unversioned scoring inputs,
uncontrolled product activation, incomplete historical candidate records, and
potential source withdrawal/misclassification paths were closed.

### Remaining release blockers

Real evidence ingestion and staffed governance, live reviewed merchant/conversion/email adapters and credentials, a distributed limiter, external secret/key rotation, conversion reconciliation, SSR/prerendering, durable media/scanning, legally approved retention, production infrastructure, centralized observability and alert delivery, durable/PITR recovery proof, executed security gates and independent penetration testing, production-equivalent load testing, and manual/real-device accessibility validation.

### Phase 9 verification

- Frontend formatting, TypeScript, zero-warning ESLint, 23 Vitest files / 50
  tests, production build, and npm audit passed. The main entry is 287.12 KiB
  raw / 87.62 KiB gzip.
- Playwright ran Chrome desktop and Pixel 5 projects against a fresh seeded
  stack: 21 tests passed and three duplicate project scenarios were
  intentionally skipped. The explicit 320–1920 px matrix had no horizontal
  overflow, and key pages passed Axe serious/critical checks.
- The browser gate exposed two real frontend defects: homepage bronze labels
  missed the automated contrast threshold, and React strict effects consumed a
  one-time verification token twice. Both were fixed; the latter also has a
  focused StrictMode regression test.
- Go formatting, serialized full PostgreSQL tests, the serialized race suite,
  vet, build, and `govulncheck` passed. The vulnerability result remained zero
  reachable/imported-package findings with one unreachable module advisory.
- A fresh PostgreSQL 17 database applied all 17 migrations and accepted the
  development seed twice. A shared-fixture race between integration packages
  was exposed and the CI database suites were made explicitly serial.
- Compose config/build/startup, readiness outage/recovery, and graceful API and
  worker shutdown passed locally. A host Vite port collision was environmental;
  the browser suite used an isolated Vite server instead.
- A migration-manifest-fingerprinted backup restored successfully into a clean
  PostgreSQL 17 target; a populated target was rejected with exit code `4`.
  The live drill exposed and led to fixes for a wrong migration-ledger column
  and masked pipeline failure.
- CI/security workflows are configured but **not executed** remotely. No live
  provider, public traffic, real financial data, legal approval, independent
  penetration test, or WCAG certification is claimed.
