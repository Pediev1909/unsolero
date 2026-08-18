# Production readiness

## Full readiness execution status — 2026-08-18

Verdict: **NOT PRODUCTION READY.** Repository-level hardening and isolated
fictional-data validation are strong, but required hosted and independently
verifiable launch evidence does not exist.

This execution added a fail-closed authenticated alert webhook boundary,
durable worker-heartbeat metrics, production topology validation, a tested
least-privilege PostgreSQL grant template, encrypted-backup handoff scripts,
dependency-review and all-image supply-chain gates, and a digest-oriented
release-candidate workflow. It also corrected Compose validation so that the
container test service really executes the complete integration and race suites.
Twenty migrations apply cleanly, the fictional seed remains idempotent, and a
checksum-verified logical backup restored into a clean database.

The strict repository-readiness score is **8.0/10**. That score is capped and
does not authorize launch. Vercel project/team/domain details, a hosted runtime,
managed PostgreSQL/Redis/private storage, a secret manager, registry artifacts,
central telemetry, actual alert delivery, encrypted off-site backup/PITR,
measured RPO/RTO, production-equivalent soak, independent penetration and
accessibility reports, and legal/privacy/business approval remain absent.

The architecture decision is deliberately split: a Vercel frontend may be
viable after its real project and routing constraints are inspected, but the Go
API, workers, migrations, managed data services and backup operations require a
separate long-running container platform. A generic SPA fallback must not
replace UNSOLERO's genuine HTTP 404, canonical and noindex behavior, and a
frontend-only split must not silently break the current same-origin session and
CSRF boundary.

## Phase 12 status — 2026-08-18

Verdict: **PARTIAL. Hosted staging and external launch gates remain blocked; public production traffic is not approved.**

Phase 12 audited external execution capability using Phase 11 as the local
baseline. No authorized cloud control plane, managed services, registry,
central collector, alert destination, provider sandbox, independent assessor or
legal approver was available. Therefore no hosted environment was provisioned,
no immutable candidate digest was promoted, no alert was delivered, no PITR or
managed failover occurred, and no RPO/RTO or production-capacity claim exists.

The strict score remains **7.8/10**. Repository quality did not regress, but
documentation of a missing external gate earns no production-readiness credit.
Every owner, evidence date and remaining action is recorded in
[`docs/PHASE_12_EVIDENCE.md`](./docs/PHASE_12_EVIDENCE.md).

## Phase 11 status — 2026-08-18

Verdict: **PARTIAL. Public production traffic is not approved.**

Repository/local staging now verifies 19 immutable migrations, bounded media
reconciliation, genuine public 404/canonical/noindex semantics, authenticated
OpenMetrics with durable cross-process operational state, deterministic bundle
and HTTP budgets, two API/two worker TLS topology, shared Redis/PostgreSQL/
private object storage, dependency outage recovery, replica loss, short soak,
and checksum backup/clean restore.

Launch remains blocked on hosted managed infrastructure, secrets/KMS, central
telemetry and delivered alerts, production backup/PITR evidence, real provider
sandboxes and contracts, representative load/Web Vitals, security and
accessibility independent review, legal/privacy approvals, and accountable
operations. Local Docker measurements are not production capacity evidence.
See [`docs/PHASE_11_EVIDENCE.md`](./docs/PHASE_11_EVIDENCE.md).

Last audited: 2026-08-17

## Readiness decision

UNSOLERO has a stronger repository foundation and is suitable for a controlled, fictional-data staging exercise. It is **not production-ready and is not approved for public production traffic**. No live provider has been activated.

This audit covers the application repository. It does not certify a cloud account, DNS, TLS termination, managed PostgreSQL service, object store, CDN, WAF, secret manager, backup schedule, alert destination, or incident-response process because none is defined in this repository.

## Strict Phase 10 scorecard

| Category | Score |
| --- | ---: |
| Architecture | 8.5/10 |
| Security | 8.0/10 |
| Privacy/data governance | 7.0/10 |
| Authentication/MFA | 8.0/10 |
| Recommendation trust | 9.0/10 |
| Commerce/affiliate | 7.0/10 |
| Database/migrations | 8.5/10 |
| Resilience/performance | 7.0/10 |
| Observability/operations | 6.0/10 |
| Frontend/accessibility | 7.0/10 |
| Backup/disaster recovery | 6.0/10 |
| Overall production readiness | **7.0/10** |

The score increased only for exercised repository controls: distributed Redis
rate limiting, private S3-compatible media, durable deletion retries,
transactional SMTP boundaries, bounded user collections, OpenMetrics export,
digest-pinned integration dependencies, and fresh migration/integration tests.
It is capped by absent hosted infrastructure, scanner, live email/alerting,
production DR/capacity evidence, remote security gates, legal approval, and
independent penetration/accessibility validation.

See [`docs/PRE_LAUNCH_SCORECARD.md`](./docs/PRE_LAUNCH_SCORECARD.md) and
[`docs/STAGING_PRODUCTION_PARITY.md`](./docs/STAGING_PRODUCTION_PARITY.md).

## Historical Phase 9 scorecard

These scores summarize evidence, not effort. A high repository score cannot
cancel an external launch blocker.

| Category | Score |
| --- | ---: |
| Architecture | 8.5/10 |
| Security | 7.5/10 |
| Privacy/data governance | 7.0/10 |
| Authentication/MFA | 8.0/10 |
| Recommendation trust | 9.0/10 |
| Commerce/affiliate | 7.0/10 |
| Database/migrations | 8.0/10 |
| Resilience/performance | 6.5/10 |
| Observability/operations | 5.5/10 |
| Frontend/accessibility | 7.0/10 |
| Backup/disaster recovery | 6.0/10 |
| Overall production readiness | **6.5/10** |

The detailed evidence and rationale are in
[`docs/PRODUCTION_VALIDATION.md`](./docs/PRODUCTION_VALIDATION.md).

## Completed controls

| Area                | Implemented state                                                                                                                                                                                                                                                                |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Security            | API and production-web responses set content-type, frame, referrer, permissions, opener, transport, and content-security policies. State-changing browser requests receive exact-origin and Fetch Metadata checks. Panic responses are generic.                                  |
| Authentication      | Passwords use bounded Argon2id. Verification/reset/session/challenge tokens are random, stored hashed, expiring, and one-use where applicable. Sessions enforce absolute/idle expiry and revocation. TOTP secrets use AES-256-GCM; recovery codes are hashed and one-use. Production refuses insecure cookies, development email delivery, missing MFA encryption, and privileged access without recent MFA. |
| Authorization       | Explicit server-side permissions map database roles to each admin route; production requires recent MFA for privileged permissions. Owned session/export/deletion/wishlist/comparison/draft/setup operations derive user identity from the session. Browser guards are convenience only. |
| Product evidence    | Recommendation-critical facts and scores use versioned revisions, per-field provenance, dated/confidence-rated observations, score rationale, independent approval/publication roles, freshness checks, and append-only audit history. Public catalog queries fail closed for ungoverned, unpublished, expired, or subsequently withdrawn provenance. |
| Recommendation replay | Authenticated completed recommendations persist constraints, policy/engine versions, result fingerprints, reasons, prices, and the complete commercial-free candidate fact/score/revision snapshot used by the deterministic run. |
| Input validation    | JSON body sizes are bounded, unknown fields are rejected, content types are checked, domain validation is repeated below the transport layer, UUIDs and query bounds are validated, analytics fields are allowlisted, and image uploads are size/MIME constrained.               |
| SQL injection       | PostgreSQL adapters use positional parameters. The one formatted commerce query selects between two constant predicates; it never interpolates request data. Sorting uses validated values through SQL `CASE` expressions.                                                       |
| CORS and CSRF       | The supported deployment is same-origin. No permissive CORS headers are emitted. Mutations reject mismatched origins and cross-site Fetch Metadata while `SameSite=Lax` provides another browser boundary.                                                                       |
| Rate limiting       | Authentication, registration, password reset, account security, recommendation, analytics, affiliate, admin, and other mutations use separate policies. The Redis-compatible adapter is atomic across replicas, TTL-bound, namespaced, uses HMAC-pseudonymous identifiers, and fails closed. Local mode remains single-replica development only. |
| Secrets             | `.env` and local override files are ignored. AI keys remain server-only. Production validates HTTPS public URLs, secure cookies, and a TLS-enabled PostgreSQL URL. No affiliate destination is compiled into the frontend.                                                       |
| Error handling      | API errors have stable codes, safe messages, `no-store`, and correlation IDs. The frontend normalizes network, timeout, malformed success, and API failures. A route-level error screen provides recovery from lazy-route/render failures.                                       |
| Logging             | Every Go process emits privacy-filtered structured JSON. API records use a validated/generated `X-Request-ID`, method, registered route pattern, status, response size, and duration. Bodies, route values, queries, cookies, authorization material, raw addresses, user agents, destinations, and arbitrary database details are not logged. |
| Database reliability | Central pool construction enforces connection, statement, lock, and idle-transaction timeouts plus bounded pool lifetimes. Database failures are reduced to safe operational classes. Migrations are timeout-bounded and rollback on failure. |
| Backup tooling      | Non-root local logical-backup/clean-restore tooling writes atomic PostgreSQL custom archives, SHA-256 checksums, and migration metadata. Restore verifies integrity, requires an empty target, uses one transaction, and validates migration state. Durable scheduling/off-site storage remains external. |
| Database migrations | Migrations are ordered, immutable, SHA-256 checked, advisory-locked, and transactional. A fresh database is created entirely from migrations. Demo seeding is separate, explicit, and idempotent.                                                                                |
| Frontend resilience | TanStack Query centralizes server state; important async views expose loading, empty, error, and success states. API requests have a 15-second default timeout and safe retry behavior.                                                                                          |
| Accessibility       | Semantic landmarks, skip links, visible focus styles, labeled controls, keyboard-operable dialogs/drawers/tabs, reduced-motion handling, alt text, and responsive comparison behavior are present. Chrome desktop/mobile Playwright paths and Axe serious/critical checks pass locally; this is not a WCAG certification. |
| SEO                 | Public pages set canonical, robots, Open Graph, Twitter, and relevant JSON-LD metadata. Backend-generated sitemap and robots responses include published resources only. Production Nginx proxies both discovery files instead of returning the SPA shell.                       |
| Performance         | Route modules are lazy-loaded, TanStack Query avoids duplicate server-state fetching, product/content images below the fold are lazy-loaded, primary hero/gallery images are prioritized, and Vite produces hashed split assets.                                                 |
| Caching             | Account/admin/error/health API responses default to `no-store`. Public catalog and editorial APIs opt into bounded caching. Hashed frontend assets receive long-lived caching; SPA documents revalidate. Uploaded immutable media receives a one-year policy.                    |
| Containers          | Separate API, worker, migration, and seed targets contain only their required executables/data and run non-root. API/worker Compose services are read-only with dropped capabilities and bounded stop time. The digest-pinned production Nginx target runs non-root and was verified read-only with writable tmpfs paths. |
| Media boundary      | Application-owned scanner/storage ports validate content before persistence. Local and private S3-compatible stores use product-scoped deterministic digest keys and conditional/atomic creation. Known deletion failures enter durable bounded retries. Production refuses local storage and development/disabled scanning; a reviewed scanner and managed private bucket remain external. |
| CI definition       | Checked-in workflows define frontend, backend/race/database, Compose/browser, dependency, secret, SAST, image, digest, and SBOM gates. They have not executed in a protected remote repository and do not yet constitute release evidence. |
| Verified conversions | Provider-neutral authenticated webhook/import ports fail closed without a reviewed adapter. PostgreSQL enforces delivery and provider-event idempotency, immutable facts, normalized order/commission lifecycles, optional evidence-based click attribution, reconciliation audit, and original-currency reporting. Missing reconciliation coverage returns `no_data`. |
| Analytics privacy | Optional events require current server-held versioned consent; client UUID uniqueness handles concurrent retry; exact schemas reject sensitive/free-form fields; anonymous identities are opaque and claim-protected; raw receipts and reportable facts are separate; account export/deletion and bounded indexed retention cleanup are integrated. |

## Incomplete before public launch

The following work requires deployment infrastructure, policy decisions, or external validation and cannot be completed solely in this repository:

1. Provision production TLS, DNS, a private application network, managed PostgreSQL, a secret manager, centralized logs, metrics, traces, alerts, and an on-call destination.
2. Independently review the pinned third-party action commits, then run the checked-in dependency, secret, SAST, image, digest, and SBOM workflows in a protected repository; remediate findings; then commission an independent penetration test.
3. Define service-level objectives, traffic assumptions, and per-route latency targets, then repeat the Phase 8 local baseline at expected peak concurrency on production-equivalent infrastructure with a sustained soak and resource limits.
4. Provision and validate the implemented private S3-compatible adapter with least-privilege IAM, KMS, versioning, lifecycle/inventory reconciliation and disaster recovery; integrate a reviewed malware/content scanner before accepting production uploads.
5. Configure and review the implemented transactional SMTP boundary with a real provider, credentials, domain controls, templates, bounce/complaint handling, delivery monitoring and sandbox evidence.
6. Obtain legal/operational approval for the documented account anonymization and retention schedule; validate security-event, recommendation, click, conversion, and administrator-audit retention periods in the production jurisdiction.
7. Choose and validate recovery point and recovery time objectives, then prove database and media restoration in a staging environment.
8. Add SSR or deterministic prerendering for acquisition pages. Current client-rendered metadata works in modern crawlers but is less reliable for all crawlers and social unfurlers.
9. Replace the CSP allowance for inline scripts/styles with nonce- or hash-based delivery when the rendering strategy supports it. Inline JSON-LD and two bounded React style values currently require the documented allowance.
10. Retain the passing Playwright/Axe gate and complete a manual WCAG 2.2 AA review with screen readers, zoom/reflow, high contrast, keyboard-only navigation, and representative real devices.
11. Add release performance budgets and automated Lighthouse/Web Vitals checks. Current route splitting and bundle output are reasonable, but a single build-size snapshot is not a user-performance guarantee.
12. Register and validate real offer and conversion adapters with production credentials, exact signature/import contracts, provider sandbox evidence, external alert delivery, and reconciliation runbooks. The provider-neutral systems exist, but every live provider remains correctly disabled.
13. Define the real evidence-review organization: name trained editors, independent reviewers and publishers; document acceptable source classes, confidence calibration, freshness periods, conflicts, withdrawal, and emergency unpublication; and require privileged-action MFA before real product revisions are published.
14. Build a reviewed evidence import/normalization workflow for real products. The implemented API is deliberately provider-neutral and human-governed; it does not scrape, infer, or invent missing product facts.
15. Obtain qualified privacy/legal approval for purposes, consent language/version, regional behavior, data-subject procedures, and every engineering retention default/hold in `docs/DATA_RETENTION.md`. Configure backup/exporter deletion and prove the runbook.

### Requires production infrastructure

Hosting, TLS/DNS, private networking, managed PostgreSQL, least-privilege roles,
a secret manager, distributed rate limiting/edge protection, central
logs/metrics/traces, delivered alerts, CI/CD promotion controls, resource
limits, durable off-site backups, PITR/failover, object storage, and
production-equivalent load/soak testing.

### Requires external providers

Reviewed transactional email, merchant/offer feeds, affiliate networks,
verified conversion delivery, and their credentials, sandbox certification,
signature/key-rotation contracts, monitoring, reconciliation, and escalation
contacts. Provider-neutral interfaces do not satisfy this requirement.

### Requires legal/business approval

Privacy policy and regional basis, consent language, retention/deletion policy,
terms, affiliate and sponsorship disclosures, processor agreements, evidence
governance staffing, supported currencies/regions, and incident-notification
decision ownership.

### Requires independent validation

Penetration testing, security architecture review of the actual deployment,
automated SAST/secret/image/SBOM gates, WCAG 2.2 AA accessibility audit,
screen-reader and keyboard review, browser/device matrix, and an observed
incident/restore exercise with independent witnesses.

## Known limitations

- Redis-compatible distributed rate limiting is implemented, but no managed TLS/authenticated service, failover policy, eviction policy, capacity evidence, or real ingress topology has been validated. Local limiter mode remains one-replica only.
- Forwarded client addresses are ignored by default. A reverse proxy may provide a single `X-Forwarded-For` IP only when its immediate peer address is inside an exact `TRUSTED_PROXY_CIDRS` range. The supplied Nginx configuration overwrites, rather than appends, the untrusted client header. Production topology must define and test these ranges; broad private-network trust is unsafe.
- The checked-in Compose file is a local development topology. It exposes PostgreSQL indirectly only to the Compose network, exposes the API for developer access, and runs the Vite development server. It is not the production orchestrator definition.
- PostgreSQL application and migration traffic currently use the same configured database credential in local Compose. Production must use separate least-privilege runtime, migration, backup, and reporting roles.
- API paths contain both foundational `/api/v1/health/*` probes and unversioned product endpoints. Existing contracts are documented, but a breaking public API will need an explicit versioning plan.
- Database migrations are forward-only. Rollback uses application rollback plus a new corrective migration, or database restoration for a destructive incident.
- Private S3-compatible media storage is implemented and integration-tested with isolated MinIO. Production still requires a managed private bucket, TLS/KMS/IAM, inventory/versioning/restore evidence and a reviewed scanner. The durable retry queue covers known failed/deleted objects; a process crash between object creation and database registration can require provider inventory reconciliation.
- The CSP permits inline scripts and styles for current JSON-LD and bounded dynamic presentation. External scripts, frames, plugins, and arbitrary connection targets remain blocked.
- No revenue, conversion, customer, or review data is fabricated. Seed content and products are explicitly fictional and must not be loaded in production.
- Verified-conversion infrastructure does not create business data by itself. Until a reviewed provider adapter, credential, contract, and successful reconciliation exist, operator monetization metrics intentionally display **No data**.
- There is no generic signed-file conversion upload route. Such a route cannot be authenticated safely without a real provider's signature, transport, replay, key-rotation, and payload contract; authenticated pull imports and the provider-specific webhook verification port are the implemented ingestion shapes.
- The only populated evidence is the explicitly fictional development fixture. No real product is publishable until real sources are reviewed and linked to every populated recommendation-critical fact and score.
- Recommendation capability, requirement, compatibility, redundancy, goal, role, and spatial behavior is now a separately persisted and immutable policy artifact. Historical replay still requires retaining the source/binary for the recorded `engine_version`; policy data cannot reproduce an unavailable algorithm implementation by itself.
- Policy creation currently has a governed PostgreSQL model and protected list/transition API, but no complete visual policy authoring editor. Operators must create new draft policy data through reviewed migrations or a controlled database workflow until a schema-validating authoring UI/importer exists. Never edit an active policy.
- Evidence source ingestion is manual/provider-neutral. There is no manufacturer feed, independent-lab connector, scheduled freshness review, or notification job yet; expired evidence correctly removes affected products from public candidate reads.
- `govulncheck -show verbose` reports one module-level advisory for the unmaintained `golang.org/x/crypto/openpgp` package. UNSOLERO imports `x/crypto/argon2`, not `openpgp`; the scanner reports zero vulnerable imported packages and zero reachable symbols. The module has no release that removes that package, so CI must continue checking reachability and future dependency changes.

## Environment variables

All production secrets must be injected by the deployment secret manager, not committed files or image layers.

| Variable                               | Required               | Production requirement                                                                                                  |
| -------------------------------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `APP_ENV`                              | Yes                    | Set to `production`; enables production validation.                                                                     |
| `APP_VERSION`                          | Yes                    | Immutable release identifier such as a commit SHA.                                                                      |
| `PUBLIC_SITE_URL`                      | Yes                    | Exact public HTTPS origin with no path or trailing slash. Used for allowed origin, canonical URLs, robots, and sitemap. |
| `DATABASE_URL`                         | Yes, secret            | PostgreSQL URL using `sslmode=require`, `verify-ca`, or preferably `verify-full`; use the least-privilege runtime role. |
| `DATABASE_MAX_CONNECTIONS`             | Yes                    | Per-replica pool maximum; default `20`. Size against database capacity and replica count.                               |
| `DATABASE_MIN_CONNECTIONS`             | Yes                    | Per-replica warm minimum; default `2`, never above the maximum.                                                         |
| `DATABASE_MAX_CONNECTION_LIFETIME`     | Yes                    | Default `30m`; keep below relevant proxy/server lifetime.                                                               |
| `DATABASE_MAX_CONNECTION_IDLE_TIME`    | Yes                    | Default `5m`.                                                                                                           |
| `DATABASE_HEALTH_CHECK_PERIOD`         | Yes                    | Default `1m`.                                                                                                           |
| `DATABASE_CONNECT_TIMEOUT`             | Yes                    | Startup database connection deadline; default `10s`.                                                                    |
| `DATABASE_STATEMENT_TIMEOUT`           | Yes                    | Per-connection statement ceiling, default `15s`; valid range `1s`–`5m`.                                                  |
| `DATABASE_LOCK_TIMEOUT`                | Yes                    | Lock wait ceiling, default `5s`; valid range `100ms`–`1m`.                                                               |
| `DATABASE_IDLE_TRANSACTION_TIMEOUT`    | Yes                    | Idle-in-transaction ceiling, default `30s`; valid range `1s`–`10m`.                                                      |
| `DATABASE_MIGRATION_TIMEOUT`           | Migration job          | Whole migration-run and migration-session ceiling, default `5m`; valid range `1m`–`1h`.                                 |
| `API_PORT`                             | No                     | Internal listener, default `8080`.                                                                                      |
| `HTTP_READ_HEADER_TIMEOUT`             | Yes                    | Slow-header protection, default `5s`.                                                                                   |
| `HTTP_READ_TIMEOUT`                    | Yes                    | Request read ceiling, default `10s`.                                                                                    |
| `HTTP_WRITE_TIMEOUT`                   | Yes                    | Response write ceiling, default `30s`, and must exceed `HTTP_HANDLER_TIMEOUT`.                                           |
| `HTTP_IDLE_TIMEOUT`                    | Yes                    | Keep-alive idle ceiling, default `60s`.                                                                                  |
| `HTTP_HANDLER_TIMEOUT`                 | Yes                    | Context deadline applied before handlers, default `20s`.                                                                |
| `HTTP_SHUTDOWN_TIMEOUT`                | Yes                    | In-flight drain period before forced close, default `10s`.                                                              |
| `HTTP_MAX_HEADER_BYTES`                | Yes                    | Header bound, default `32768`, valid range 8 KiB–1 MiB.                                                                 |
| `SESSION_COOKIE_NAME`                  | Yes                    | Stable cookie name containing only letters, numbers, `_`, or `-`.                                                       |
| `SESSION_COOKIE_SECURE`                | Yes                    | Must be `true` in production.                                                                                           |
| `SESSION_TTL`                          | Yes                    | Absolute session lifetime; default `720h`.                                                                              |
| `SESSION_IDLE_TTL`                     | Yes                    | Idle lifetime, positive and no longer than `SESSION_TTL`; default `168h`.                                               |
| `EMAIL_PROVIDER`                       | Yes                    | `development`/`disabled` locally; `smtp` requires sender/address and STARTTLS in production; `external` remains fail-closed until linked. |
| `EMAIL_SENDER_NAME` / `EMAIL_SENDER_ADDRESS` | SMTP | Validated display name and dedicated transactional sender; no header controls.                                       |
| `EMAIL_SMTP_ADDRESS`                   | SMTP                   | Host and port only; production requires STARTTLS and a reviewed private/allowlisted egress path.                         |
| `EMAIL_SMTP_USERNAME` / `EMAIL_SMTP_PASSWORD` | SMTP secret pair | Both or neither; inject through the secret manager and rotate.                                                         |
| `EMAIL_SMTP_REQUIRE_TLS` / `EMAIL_SMTP_TIMEOUT` | SMTP | TLS must be `true` in production; timeout is bounded.                                                                  |
| `EMAIL_VERIFICATION_TTL`               | Yes                    | Verification token lifetime, default `24h`; valid range 15 minutes–7 days.                                              |
| `PASSWORD_RESET_TTL`                   | Yes                    | Reset token lifetime, default `1h`; valid range 10 minutes–24 hours.                                                    |
| `MFA_ENCRYPTION_KEY`                   | Production secret      | Raw-standard-base64 for exactly 32 random bytes, supplied by the secret manager. Key rotation requires a reviewed multi-key adapter. |
| `MFA_CHALLENGE_TTL`                    | Yes                    | MFA login challenge lifetime, default `5m`; valid range 1–15 minutes.                                                   |
| `MFA_STEP_UP_TTL`                      | Yes                    | Maximum age for privileged verification, default `15m`; valid range 1–60 minutes.                                      |
| `MFA_ENFORCE_PRIVILEGED`               | Yes                    | Defaults `true` in production and must remain enabled for staff access.                                                |
| `API_REPLICA_COUNT`                    | Yes                    | Declared API replica count. Values above one require `redis` or a linked `external` limiter.                            |
| `RATE_LIMIT_PROVIDER`                  | Yes                    | `local` only for one replica; `redis` implemented and requires `rediss://` in production; `external` must be linked.    |
| `RATE_LIMIT_REDIS_URL`                 | Redis secret/config    | Authenticated `rediss://` endpoint in production; never expose through frontend configuration.                          |
| `RATE_LIMIT_NAMESPACE`                 | Yes                    | Stable bounded namespace separating environment/application keys.                                                       |
| `RATE_LIMIT_KEY_SECRET`                | Production secret      | Stable raw-standard-base64 value for exactly 32 random bytes; used only to HMAC client keys.                             |
| `RATE_LIMIT_AUTH_PER_MINUTE`           | Yes                    | Default `10` per client. Confirm through security/load testing.                                                         |
| `RATE_LIMIT_REGISTRATION_PER_MINUTE`   | Yes                    | Default `5` per pseudonymous client.                                                                                    |
| `RATE_LIMIT_PASSWORD_RESET_PER_MINUTE` | Yes                    | Default `5`; generic responses still prevent enumeration.                                                               |
| `RATE_LIMIT_RECOMMENDATION_PER_MINUTE` | Yes                    | Default `20` per client.                                                                                                |
| `RATE_LIMIT_ANALYTICS_PER_MINUTE`      | Yes                    | Default `120` per client.                                                                                               |
| `RATE_LIMIT_AFFILIATE_PER_MINUTE`      | Yes                    | Default `120` per client.                                                                                               |
| `RATE_LIMIT_ADMIN_PER_MINUTE`          | Yes                    | Default `240`; authorization and MFA remain mandatory independent controls.                                             |
| `RATE_LIMIT_MUTATION_PER_MINUTE`       | Yes                    | Default `240` per client for other mutations.                                                                           |
| `OFFER_MAXIMUM_AGE`                    | Yes                    | Default `72h`; older offers are hidden and cannot redirect until refreshed.                                               |
| `AFFILIATE_CLICK_RETENTION`            | Yes                    | Default `9528h` (397 days); the worker anonymizes identifying attribution after expiry.                                  |
| `COMMERCE_WORKER_POLL_INTERVAL`        | Worker                 | Default `15s`, valid range `1s`–`5m`; tune and load-test per deployment.                                                  |
| `WORKER_CYCLE_TIMEOUT`                 | Worker                 | Per-cycle deadline, default `2m`; valid range `10s`–`30m`.                                                              |
| `WORKER_LEASE_TIMEOUT`                 | Worker                 | Stalled-job recovery age, default `1h`; must exceed expected provider processing duration.                              |
| `WORKER_MAX_ITEMS_PER_CYCLE`           | Worker                 | Bounded work per cycle, default `25`; valid range `1`–`1000`.                                                           |
| `WORKER_FAILURE_ALERT_THRESHOLD`       | Worker                 | Consecutive failed cycles before one alert attempt, default `3`.                                                        |
| `ANALYTICS_SUBJECT_COOKIE_NAME`        | Yes                    | Opaque HttpOnly browser-subject cookie; default `unsolero_analytics_subject`.                                            |
| `ANALYTICS_ANONYMOUS_RETENTION`        | Yes                    | Engineering default `2160h` (90 days), valid 1–730 days; legal/privacy approval required.                                |
| `ANALYTICS_AUTHENTICATED_RETENTION`    | Yes                    | Engineering default `9528h` (397 days), valid 1–1,095 days; legal/privacy approval required.                             |
| `ANALYTICS_RECEIPT_RETENTION`          | Yes                    | Payload-free receipt default `720h` (30 days), valid 1–180 days; legal/privacy approval required.                        |
| `ANALYTICS_CLEANUP_BATCH_SIZE`         | Worker                 | Rows per class/pass, default `1000`, valid `1`–`10000`.                                                                 |
| `ALERT_PROVIDER`                       | Yes                    | `disabled` is explicit/degraded; `external` fails startup until a reviewed notifier is linked.                          |
| `METRICS_ENABLED`                      | Yes                    | Enables the process-local aggregate endpoint; default `false`.                                                          |
| `METRICS_TOKEN`                        | If metrics enabled, secret | Bearer credential of at least 32 characters. Replace the in-process recorder for multi-replica aggregation.          |
| `PRODUCT_IMAGE_UPLOAD_DIR`             | Only for local adapter | Persistent writable directory. Production should replace the adapter with object storage.                               |
| `MEDIA_STORAGE_PROVIDER`               | Yes                    | `local` development only; `s3` is implemented and production requires secure transport; `external` must be linked.     |
| `MEDIA_S3_ENDPOINT` / `MEDIA_S3_BUCKET` / `MEDIA_S3_REGION` | S3 | Private S3-compatible target; endpoint excludes scheme and bucket must already exist.                                  |
| `MEDIA_S3_ACCESS_KEY` / `MEDIA_S3_SECRET_KEY` | S3 secrets | Least-privilege object credentials injected by the secret manager.                                                     |
| `MEDIA_S3_SECURE`                      | S3                     | Must be `true` in production. Public bucket access is unsupported.                                                      |
| `MEDIA_SCAN_PROVIDER`                  | Yes                    | `development` is a fixture detector only. Production requires a linked `external` scanner and rejects disabled/development modes. |
| `MIGRATIONS_DIR`                       | Migration job          | Directory containing immutable migration SQL, default `./migrations`.                                                   |
| `SEEDS_DIR`                            | Development only       | Never invoke the seed command in production.                                                                            |
| `AI_PROVIDER`                          | Yes                    | Keep `disabled` until a reviewed provider adapter is enabled.                                                           |
| `AI_MODEL`                             | If AI enabled          | Server-side configured model identifier.                                                                                |
| `AI_API_KEY`                           | If AI enabled, secret  | Secret-manager value; never use a `VITE_*` name.                                                                        |
| `AI_TIMEOUT`                           | If AI enabled          | Default `15s`.                                                                                                          |
| `AI_MAX_RESPONSE_BYTES`                | If AI enabled          | Default `65536`, valid range 1 KiB–1 MiB.                                                                               |
| `BACKUP_NAME`                          | Local backup tool      | Unique artifact stem; existing targets are never overwritten.                                                           |
| `BACKUP_UID`, `BACKUP_GID`             | Local backup tool      | Numeric non-root owner for bind-mounted backup artifacts.                                                               |
| `RESTORE_POSTGRES_DB`                  | Local restore drill    | Isolated clean restore target name.                                                                                     |

`POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `WEB_PORT`, `VITE_DEV_API_PROXY_TARGET`, and the local insecure `DATABASE_URL` in `.env.example` are development-Compose settings, not a production secret-management model.

## Deployment requirements

1. Build immutable API and production frontend images from the reviewed revision. Record image digests and `APP_VERSION`.
2. Scan source dependencies and final images; fail the release for policy-defined critical/high findings.
3. Provision a private PostgreSQL endpoint with TLS verification, encryption at rest, automated backups, point-in-time recovery, and separate roles.
4. Run the migration image as a one-shot job using the migration role. Do not start new application replicas until it succeeds.
5. Deploy API replicas on a private network. Expose only the TLS edge/CDN or production web proxy. Preserve the exact `Host` and scheme and overwrite forwarding headers at the trusted edge.
6. Mount no mutable application filesystem other than the temporary directory and approved media adapter. Prefer object storage rather than the local media volume.
7. Route liveness to `/api/v1/health/live` and readiness to `/api/v1/health/ready`. Remove an instance from service when readiness fails; do not restart solely on a transient readiness failure.
8. Apply edge-level request/body limits, distributed rate limits, bot controls, TLS policy, and the final nonce/hash CSP. Do not enable cross-origin credentialed API access without a separate threat-model review.
9. Perform smoke tests for registration/login/logout, catalog, recommendation generation, account ownership, admin denial/allowance, affiliate redirect, sitemap, robots, and uploaded media.
10. Use a canary or rolling release with automatic rollback based on error rate, latency, readiness, and business-path smoke tests.

## Database migration process

1. Back up the database and confirm the most recent restore test meets the release recovery objectives.
2. Review the new numbered SQL migration for locks, table rewrites, data volume, index-build strategy, and compatibility with both old and new application versions.
3. For destructive changes, use expand/migrate/contract across releases. Never rename/drop a live column in the same release that stops reading it.
4. Run `/usr/local/bin/migrate` once with the migration role. The runner obtains an advisory lock, validates all applied checksums, and applies each pending file in its own transaction.
5. Verify `platform.schema_migrations`, database health, critical queries, and smoke tests before shifting traffic.
6. Do not edit an applied file. Ship a new corrective migration. If restoration is required, stop writes first and follow the tested incident runbook.

## Backup and recovery requirements

- Approve RPO and RTO before launch. The repository runbook proposes 24 hours and 4 hours as initial targets, not achieved measurements.
- Enable encrypted automated PostgreSQL snapshots and continuous WAL archiving/point-in-time recovery in a separate failure domain.
- Back up uploaded media until object storage with versioning and lifecycle policy replaces the local adapter.
- Protect backup credentials separately from runtime credentials and restrict delete/restore privileges.
- Retain backups according to legal, privacy, and business requirements; document deletion behavior for user-data requests.
- Run scheduled restore drills into an isolated environment. A backup is not considered valid until schema checks, row counts, authentication, catalog reads, recommendation reads, and media access pass after restoration.
- Alert on failed backups, WAL/archive lag, expired recovery windows, and failed restore drills.

## Monitoring and alerting requirements

Collect and dashboard at minimum:

- request rate, status classes, route-group p50/p95/p99 latency, timeouts, panics, and rate-limit rejections;
- liveness/readiness state, restart count, CPU, memory, file descriptors, goroutines, and container saturation;
- PostgreSQL availability, connection utilization, wait time, slow queries, locks, deadlocks, replication/WAL lag, storage, and backup status;
- login failures and rate limits without logging passwords, cookies, raw session tokens, or full email addresses;
- admin mutations from `admin.audit_log`, privileged role changes, and repeated authorization failures;
- recommendation generation count/error/latency and deterministic-engine failures;
- analytics ingestion failures and affiliate redirect errors without destination URLs or commission metadata in logs;
- frontend JavaScript errors, failed lazy chunks, API network failures, and Core Web Vitals using a consent-appropriate provider;
- sitemap/robots availability and scheduled synthetic checks for the primary public and authenticated journeys.

Alerts need an owner, severity, runbook link, deduplication policy, and tested delivery channel. Correlate application and edge telemetry with `X-Request-ID`. Define retention and redaction before sending logs to an external provider.

## Release verification

The repository release gate for this audit is:

```bash
npm --prefix frontend run format:check
npm --prefix frontend run lint
npm --prefix frontend run typecheck
npm --prefix frontend run test
npm --prefix frontend run build
npm --prefix frontend audit --omit=dev --audit-level=high

cd backend
go test ./...
go test -race ./...
go vet ./...
go build ./...
govulncheck ./...

docker compose --env-file .env.example config --quiet
docker compose --env-file .env.example build
docker compose --env-file .env.example up -d
```

The backup/restore and failure drills in `docs/BACKUP_RESTORE.md` and
`docs/OPERATIONS.md` are release evidence, not implied by a successful build.

The final audit report must record actual results rather than treating this checklist as evidence by itself.

### Audit evidence for this revision

- Frontend formatting, ESLint, TypeScript, 21 Vitest files / 42 tests, and the Vite production build passed.
- The main emitted JavaScript chunk was 286.85 KiB raw / 87.54 KiB gzip; route pages, including the admin shell, remain split into lazy chunks.
- `npm audit --omit=dev --audit-level=high` reported zero vulnerabilities.
- `gofmt` verification, `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...` passed in the official Go 1.25 container.
- `govulncheck` v1.7.0 reported zero affected vulnerabilities and zero vulnerabilities in imported packages. It reported one advisory in a required module with no reachable symbols.
- Docker Compose configuration, builds, and startup passed. PostgreSQL, migrations, API readiness, and the Vite development container became healthy/available.
- A separate empty PostgreSQL 17 volume applied all 17 migrations. Before integration fixtures, the idempotent development seed succeeded twice and produced exactly 8 categories, 10 brands, 30 governed products, 90 offers, 90 affiliate links, one explicitly fictional verified evidence source, 326 fact-provenance links, 240 score rationales, and one active data-driven fitness policy. It creates no users, analytics events, conversions, orders, commissions, or revenue.
- The full PostgreSQL integration suite passed, including import history/observation fixtures that intentionally remain auditable in the isolated test database. Runtime/customer analytics are not fabricated by the seed.
- Phase 1-specific tests verified review/publication separation of duties,
  completeness and freshness enforcement, immediate fail-closed source
  withdrawal, immutable candidate snapshots, and unchanged recommendation
  output after commercial commission/priority changes.
- Phase 2-specific tests verify policy approval/activation/retirement separation,
  active-policy immutability, unsupported-category exclusion, revision binding,
  compatibility/incompatibility, redundancy, room/access/clearance constraints,
  explicit overlap, missing-measurement rejection, deterministic output,
  historical policy-input persistence, and commercial-data independence.
- Phase 4-specific tests verify forged/disabled webhook rejection, replay and
  timeout retry behavior, concurrent provider-event idempotency, conflicting
  duplicate rejection, normalized order/commission transitions, original
  currency preservation, bounded money, evidence-based attribution, immutable
  facts, reconciliation, pending/reversed commission exclusion, no-data versus
  zero semantics, admin authorization, and recommendation independence from
  conversion/commission/provider/attribution/click mutations.
- Phase 5-specific unit, HTTP, and PostgreSQL tests verify generic registration
  and reset responses, one-use/expired tokens, session ownership and revocation,
  password-reset global revocation, password-change other-session revocation,
  encrypted TOTP, single-use recovery codes, scoped login challenges, immutable
  audit events, safe export/anonymizing deletion, permission boundaries, recent
  backend MFA, and security cleanup that retains audit history.
- Phase 6-specific unit, HTTP, and PostgreSQL tests verify server-authoritative
  versioned consent, pre-consent/withdrawal rejection, strict property schemas,
  concurrent event deduplication, bot/prefetch filtering, authenticated
  identity claims, cross-user/revoked-claim rejection, bounded retention,
  export/deletion unlinking, raw-versus-reportable separation, incomplete/no-data
  reporting, affiliate privacy, security-event isolation, and least-privilege
  analytics access.
- Live Phase 5 smoke tests verified registration without auto-login, email and
  reset token replay rejection, active-session inspection, credential-change
  revocation, MFA enrollment/recovery/TOTP login, verified and recent-MFA admin
  access, secret-free export, authenticated deletion, and retained deletion
  audit history. A post-smoke API/worker log scan found no error-level entries or
  searched secret fields.
- Live Phase 6 smoke tests verified consent grant/withdrawal, one-row persistence
  across duplicate submissions, payload-free bot filtering, anonymous-to-account
  claim and export, deletion/anonymization, post-deletion anonymous handling,
  analyst-versus-admin access, and bounded worker cleanup. Retention and
  reportable-event query plans used their intended partial indexes.
- Live smoke tests verified `200` health/catalog/sitemap/robots responses, `401` account/admin guards, `403` hostile-origin mutation rejection, UNSOLERO frontend/editorial branding, and a complete register/session/logout flow.

### Phase 7 repository and local-system evidence

- Frontend formatting, TypeScript, zero-warning ESLint, 22 Vitest files / 45
  tests, and production build passed. The largest application entry was 287.11
  KiB raw / 87.61 KiB gzip. Production npm audit reported zero vulnerabilities.
- Go formatting, unit tests, race tests, vet, and build passed. `govulncheck`
  found zero reachable/imported-package vulnerabilities; one required-module
  advisory has no reachable symbol.
- A fresh PostgreSQL 17 database applied all 17 migrations and the explicit
  fictional seed. The complete Go suite then passed with PostgreSQL integration
  tests enabled. A deliberately failing migration rolled back its schema change
  and was not recorded.
- The isolated Compose stack built and started. Liveness was `200`; readiness was
  serving/degraded only because alert delivery is honestly disabled. Stopping
  PostgreSQL changed readiness to `503 unavailable`; restarting it restored
  readiness to `200`.
- API and worker SIGTERM via Compose completed within the grace period and logged
  privacy-safe shutdown completion. An in-flight API drain test passed inside a
  network-enabled container.
- Non-root logical backup, SHA-256 verification, clean transactional restore,
  migration verification, and seeded row-count verification passed locally.
  Duplicate backup names and non-empty restore targets failed closed.
- The digest-pinned production frontend image built and served the SPA plus API
  proxy as UID 101 with a read-only root filesystem, dropped capabilities, and
  only bounded tmpfs write paths.

### Phase 8 repository and local-system evidence

- Frontend formatting, TypeScript, zero-warning ESLint, 22 Vitest files / 49
  tests, and the production build passed. The current main entry is 287.11 KiB
  raw / 87.60 KiB gzip. Production npm audit reported zero vulnerabilities.
- Go formatting, unit tests, the complete PostgreSQL integration suite, the
  complete race suite, vet, and build passed. `govulncheck` found zero
  reachable/imported-package vulnerabilities; its single module-only advisory
  is in an unimported package and has no fixed release.
- A dependency-free HTTP load probe measured catalog, anonymous/authenticated
  recommendations, authentication, admin, analytics, commerce, affiliate
  redirect, and health behavior with zero unexpected responses. These are
  short local baselines, not capacity or SLO claims.
- A rollback-only fictional scale fixture exercised 10,000 users, 5,000
  governed products, 20,000 sessions, 50,000 clicks, 100,000 analytics events,
  10,000 recommendation snapshots, 10,000 verified-conversion projections,
  and representative query plans. It exposed and led to correction of an
  unindexed authentication lookup shape.
- Concurrent integration tests passed for duplicate registration, one-use
  reset tokens, one-use MFA recovery codes, and affiliate click/event
  idempotency. Existing replay, claim, conversion, worker lease, and policy
  concurrency tests also passed under the race detector.
- API readiness now verifies the exact embedded 17-migration release manifest.
  PostgreSQL outage and incompatible-schema drills returned `503`, then
  recovered to `200` after the fault was removed. Pool exhaustion honored the
  caller deadline and a slow query was canceled by the statement timeout.
- A checksum-verified backup restored into a clean PostgreSQL 17 target;
  corrupt input and a non-empty destination failed closed. Restore mechanics
  passed, but off-site durability, PITR, replica failover, and real RPO/RTO did
  not.
- Compose configuration, all current images, startup, API/catalog/frontend
  smoke checks, and the separate production Nginx image build passed. The API
  and worker run non-root/read-only/capability-dropped; development web and
  PostgreSQL retain documented exceptions. No image CVE scanner was available.

### Phase 9 repository and local-system evidence

- Frontend formatting, TypeScript, zero-warning ESLint, 23 Vitest files / 50
  tests, the production build, and npm audit passed. The main entry is 287.12
  KiB raw / 87.62 KiB gzip.
- Playwright exercised Chrome desktop and Pixel 5 projects against the fresh
  seeded stack. Twenty-one tests passed; three duplicate project scenarios were
  intentionally skipped. The paths cover homepage/catalog, the explicit
  320–1920 px width matrix,
  deterministic recommendation completion, auth/account/admin boundaries,
  analytics consent, 429/500 errors, affiliate disclosure/navigation, keyboard
  focus, and Axe serious/critical findings on key pages.
- The first expanded run exposed insufficient contrast on homepage bronze
  labels and duplicate one-time verification under React strict effects. Both
  were fixed and covered by the final browser/unit gates.
- Go formatting, serialized full PostgreSQL tests, the serialized race suite,
  vet, build, and `govulncheck` passed. The vulnerability database reported no
  reachable/imported-package findings and one unreachable module advisory.
- A fresh PostgreSQL 17 database applied all 17 migrations, and the explicitly
  fictional development seed succeeded twice. The local stack passed readiness
  outage/recovery and graceful API/worker stop checks.
- Backup `unsolero-phase9-verified` restored into a clean PostgreSQL 17 target
  with the exact 17-migration manifest fingerprint. A repeated restore into the
  populated target failed closed with exit code `4`.
- Checked-in CI/security workflows now define repository, database, Compose,
  browser, dependency, secret, SAST, image, base-digest, and SBOM gates. They
  have not run in a protected remote repository and are not counted as passed.

See `docs/PRODUCTION_VALIDATION.md`, `docs/LOAD_TESTING.md`,
`docs/SECURITY_VALIDATION.md`, `docs/MIGRATION_SAFETY.md`,
`docs/DISASTER_RECOVERY.md`, and `docs/INCIDENT_RESPONSE.md` for exact evidence
and blockers.

Public production deployment, durable/off-site backup scheduling, point-in-time
recovery, production-equivalent soak testing, comprehensive real-device/manual
accessibility validation, centralized telemetry, tested external alert delivery,
distributed abuse protection, durable media scanning/storage, executed remote
security gates, and production incident operations remain outstanding. This
audit does not approve production readiness.
