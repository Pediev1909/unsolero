# Final CTO audit

Date: 2026-08-17  
Verdict: **controlled staging only; public production launch rejected**

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

The weakest part of UNSOLERO is now operational execution around its trust model. The provenance and approval boundary exists, but it contains no real product evidence and has no staffed freshness/withdrawal operation. Merchant data has no refresh pipeline. Account recovery and staff MFA do not exist. Acquisition pages are a client-rendered SPA. Mobile quality is asserted through responsive CSS, not proven through browser tests. Those are launch blockers, not cleanup tickets.

The original audit covered 356 repository files, about 17,600 lines of Go, about 15,200 lines of frontend TypeScript/TSX, 43 Go test files, 20 frontend test files, and 10 SQL migrations. Phase 1 added the evidence module, tests, UI, and migration 11; generated dependencies and build output remain excluded from source review.

## Scorecard

| Area | Score | Brutal assessment |
| --- | ---: | --- |
| Architecture | **9/10** | Sound modular monolith with evidence and immutable data-driven policy boundaries plus reproducible run snapshots. Oversized adapters/handlers and API contract/versioning mismatch remain. |
| UX | **7/10** | Strong decision flow and honest states. Account lifecycle is incomplete, free text is non-functional, and several production transitions still expose demo assumptions. |
| UI | **8/10** | Restrained, coherent, and much better than generic SaaS. The design system is real. Legacy muted-color utilities required a global accessibility correction. |
| Mobile | **7/10** | Layouts are mobile-first and comparison is usable, but there is no automated viewport, reflow, real-device, or touch regression evidence. |
| Security | **6/10** | Session design, SQL parameterization, origin checks, upload restrictions, and role enforcement are good. No verification/recovery/MFA, distributed abuse control, malware scanning, or independent security gate. |
| Recommendation engine | **9/10** | Deterministic, revision-reproducible, evidence-gated, structurally commission-independent, and governed by explicit category/product policy data. Clearance/overlap modeling is conservative and only as complete as reviewed measurements; real evidence has not been populated. |
| Monetization readiness | **4/10** | Multiple offers and tracked redirects exist. There is no feed refresh worker, conversion import, click deduplication/bot filtering, reconciliation, or verified revenue reporting. |
| SEO readiness | **5/10** | Sitemap, canonicals, metadata, internal links, and editorial models exist. SPA rendering, soft 404s, unreliable unfurls, and no prerender/SSR prevent serious acquisition readiness. |
| Performance | **7/10** | Route splitting and image discipline are good; the production entry is about 81 kB gzip. Search and analytics storage have no demonstrated scale plan or performance budgets. |
| Production readiness | **4/10** | A credible staging system, not a public service. Local Go, PostgreSQL, frontend, build, and container checks pass, but critical deployment, recovery, security, merchant-data, and observability work remains. |

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

**Fix:** optional events now default off, a concise accessible preference surface records grant/decline, users can reopen preferences from the footer, event envelopes carry `consent_state: granted`, and the backend rejects missing/non-granted states. Merchant-click recording remains clearly disclosed essential attribution tied to an explicit outbound action.

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
| No merchant-feed/import worker exists | The new freshness control will correctly remove every offer after 72 hours if nobody refreshes it. Monetization therefore stops safely rather than remaining wrong. | Build provider adapters, scheduled imports, reconciliation, freshness alerts, and operator runbooks. |
| Clicks lack durable idempotency and bot/crawler classification | GET links can be prefetched and retries can duplicate raw click counts. Affiliate CTR is directional, not financially reliable. | Add a bounded click intent/idempotency design, bot classification, and reporting filters without breaking normal navigation. |
| No verified conversion or commission ingestion | Conversion rate, EPC, revenue per visitor/recommendation, and payouts cannot be computed. | Implement authenticated provider postbacks/imports and reconciliation before showing revenue. Keep current “No data” behavior. |
| Email verification, password recovery/change, session revocation on credential change, and admin MFA are absent | Public accounts and privileged administration are incomplete. | Implement the complete account lifecycle and require MFA for staff before production access. |
| Rate limits are process-local | Multiple replicas multiply limits and allow distributed abuse. | Enforce policy at the trusted edge or a distributed limiter; keep the API private behind it. |
| Local image storage and signature-only MIME checks | It is not multi-replica safe and is not a malware/content pipeline. | Move to object storage, decode/re-encode images, scan uploads, and serve transformed variants through a CDN. |
| SPA-only acquisition pages | Metadata depends on client execution; unknown routes return the SPA shell with HTTP 200, creating soft 404s and unreliable social unfurls. | Add SSR or deterministic prerendering with real 404 status behavior. |
| Public API versioning is inconsistent | Health probes use `/api/v1`; application resources use unversioned `/api`; the target architecture still proposes different resource paths. | Freeze an actual contract, publish it, and introduce aliases/versioning before external consumers exist. |
| Catalog search is `%ILIKE%` without a search index | It will scan as catalog size grows. | Add measured PostgreSQL full-text/trigram indexing or a dedicated search adapter only when volume justifies it. |
| Analytics has no client event ID/deduplication or retention/deletion job | Funnels can overcount and privacy operations are manual. | Add event IDs, idempotent ingestion, retention policy, user deletion/export workflows, and warehouse/export boundaries. |
| **PARTIALLY RESOLVED IN PHASE 1:** Admin authorization was one coarse role | Evidence and score publication now needs separate editor/reviewer roles and repository-enforced separation of duties. Other admin areas remain coarse, and privileged MFA is absent. | Extend least privilege to remaining admin capabilities and require privileged-action MFA before production staff access. |
| `is_demo` is inferred from a `demo-` slug prefix | SEO and disclosures depend on naming convention instead of a data invariant. | Add an explicit structured flag/provenance field through a migration and admin contract. |
| Homepage featured/example product content is hard-coded in the frontend | Replacing seed data can leave stale names and dead product links. | Drive production homepage product modules from reviewed content/catalog records with honest empty states. |
| No real browser/mobile/accessibility regression suite | Responsive classes are not proof at 320–1920 px, 200–400% zoom, keyboard, or screen readers. | Add Playwright viewport/reflow tests, axe checks, real-device smoke tests, and a manual WCAG 2.2 AA audit. |
| No production performance or load budget | A successful local build says nothing about p95 recommendation/search/redirect latency. | Define budgets, run representative load tests, collect Web Vitals, and enforce release thresholds. |
| No dependency/container/security release gate | `npm audit` could not reach the registry in this audit; no independent scan result is available. | Add locked CI scans, secret scanning, SAST, `govulncheck`, SBOMs, image scanning, and a reviewed exception process. |

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

## Phase report

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

Real evidence ingestion and staffed governance, merchant ingestion, account recovery/verification/MFA, distributed abuse controls, conversion reconciliation, SSR/prerendering, durable media, data retention/deletion, production infrastructure, observability, backup/restore proof, security scanning, load testing, and browser/mobile/accessibility validation.
