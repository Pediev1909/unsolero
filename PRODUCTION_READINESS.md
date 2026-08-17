# Production readiness

Last audited: 2026-08-17

## Readiness decision

UNSOLERO is suitable for a controlled staging deployment and production-like load, security, backup, and recovery exercises. It is **not yet approved for an unattended public production launch** until the deployment requirements and incomplete items in this document have owners and evidence.

This audit covers the application repository. It does not certify a cloud account, DNS, TLS termination, managed PostgreSQL service, object store, CDN, WAF, secret manager, backup schedule, alert destination, or incident-response process because none is defined in this repository.

## Completed controls

| Area                | Implemented state                                                                                                                                                                                                                                                                |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Security            | API and production-web responses set content-type, frame, referrer, permissions, opener, transport, and content-security policies. State-changing browser requests receive exact-origin and Fetch Metadata checks. Panic responses are generic.                                  |
| Authentication      | Passwords use bounded Argon2id parameters. Sessions use random opaque tokens, store only SHA-256 token hashes, enforce absolute and idle expiry, and use `HttpOnly`, `SameSite=Lax` cookies. Production refuses insecure session cookies.                                        |
| Authorization       | Server-side database-backed roles protect every admin route. User-owned wishlist, comparison, draft, and setup queries include the authenticated user identifier. Browser route guards are convenience only.                                                                     |
| Product evidence    | Recommendation-critical facts and scores use versioned revisions, per-field provenance, dated/confidence-rated observations, score rationale, independent approval/publication roles, freshness checks, and append-only audit history. Public catalog queries fail closed for ungoverned, unpublished, expired, or subsequently withdrawn provenance. |
| Recommendation replay | Authenticated completed recommendations persist constraints, policy/engine versions, result fingerprints, reasons, prices, and the complete commercial-free candidate fact/score/revision snapshot used by the deterministic run. |
| Input validation    | JSON body sizes are bounded, unknown fields are rejected, content types are checked, domain validation is repeated below the transport layer, UUIDs and query bounds are validated, analytics fields are allowlisted, and image uploads are size/MIME constrained.               |
| SQL injection       | PostgreSQL adapters use positional parameters. The one formatted commerce query selects between two constant predicates; it never interpolates request data. Sorting uses validated values through SQL `CASE` expressions.                                                       |
| CORS and CSRF       | The supported deployment is same-origin. No permissive CORS headers are emitted. Mutations reject mismatched origins and cross-site Fetch Metadata while `SameSite=Lax` provides another browser boundary.                                                                       |
| Rate limiting       | Authentication, recommendation generation, analytics ingestion, affiliate redirects, and other mutations have configurable per-client, per-process fixed-window limits and return `429` with `Retry-After`. Private reverse-proxy addresses may supply one validated client IP.  |
| Secrets             | `.env` and local override files are ignored. AI keys remain server-only. Production validates HTTPS public URLs, secure cookies, and a TLS-enabled PostgreSQL URL. No affiliate destination is compiled into the frontend.                                                       |
| Error handling      | API errors have stable codes, safe messages, `no-store`, and correlation IDs. The frontend normalizes network, timeout, malformed success, and API failures. A route-level error screen provides recovery from lazy-route/render failures.                                       |
| Logging             | The API emits structured JSON logs, a validated/generated `X-Request-ID`, method, query-free path, status, response size, and duration. Request bodies, cookies, authorization material, and query strings are not logged. Successful probes are debug-level to avoid log noise. |
| Database migrations | Migrations are ordered, immutable, SHA-256 checked, advisory-locked, and transactional. A fresh database is created entirely from migrations. Demo seeding is separate, explicit, and idempotent.                                                                                |
| Frontend resilience | TanStack Query centralizes server state; important async views expose loading, empty, error, and success states. API requests have a 15-second default timeout and safe retry behavior.                                                                                          |
| Accessibility       | Semantic landmarks, skip links, visible focus styles, labeled controls, keyboard-operable dialogs/drawers/tabs, reduced-motion handling, alt text, and responsive comparison behavior are present.                                                                               |
| SEO                 | Public pages set canonical, robots, Open Graph, Twitter, and relevant JSON-LD metadata. Backend-generated sitemap and robots responses include published resources only. Production Nginx proxies both discovery files instead of returning the SPA shell.                       |
| Performance         | Route modules are lazy-loaded, TanStack Query avoids duplicate server-state fetching, product/content images below the fold are lazy-loaded, primary hero/gallery images are prioritized, and Vite produces hashed split assets.                                                 |
| Caching             | Account/admin/error/health API responses default to `no-store`. Public catalog and editorial APIs opt into bounded caching. Hashed frontend assets receive long-lived caching; SPA documents revalidate. Uploaded immutable media receives a one-year policy.                    |
| Containers          | The API runs as a non-root user with a read-only root filesystem, dropped capabilities, `no-new-privileges`, a writable media volume, graceful shutdown time, and a database-aware readiness probe. Production Nginx has proxy timeouts and a bounded upload size.               |

## Incomplete before public launch

The following work requires deployment infrastructure, policy decisions, or external validation and cannot be completed solely in this repository:

1. Provision production TLS, DNS, a private application network, managed PostgreSQL, a secret manager, centralized logs, metrics, traces, alerts, and an on-call destination.
2. Run an independent penetration test and authenticated dependency/container scan in CI. Add automated secret scanning, SAST, `govulncheck`, npm audit policy, and image scanning to the release gate.
3. Define service-level objectives, traffic assumptions, per-route latency targets, and load-test the recommendation, login, catalog search, and affiliate-click paths at expected peak concurrency.
4. Replace process-local uploaded product media with durable object storage plus malware/content scanning and image transformation before running multiple API replicas.
5. Add email verification, account recovery, credential-change session revocation, and administrative MFA before granting real staff access.
6. Establish a session-retention/cleanup job for expired and revoked sessions and documented retention/deletion schedules for analytics, click attribution, audit logs, and user accounts.
7. Choose and validate recovery point and recovery time objectives, then prove database and media restoration in a staging environment.
8. Add SSR or deterministic prerendering for acquisition pages. Current client-rendered metadata works in modern crawlers but is less reliable for all crawlers and social unfurlers.
9. Replace the CSP allowance for inline scripts/styles with nonce- or hash-based delivery when the rendering strategy supports it. Inline JSON-LD and two bounded React style values currently require the documented allowance.
10. Add automated browser accessibility testing and a manual WCAG 2.2 AA review with screen readers, zoom/reflow, high contrast, and keyboard-only navigation.
11. Add release performance budgets and automated Lighthouse/Web Vitals checks. Current route splitting and bundle output are reasonable, but a single build-size snapshot is not a user-performance guarantee.
12. Implement and monitor a trusted merchant-feed refresh job. Offers now fail closed after `OFFER_MAXIMUM_AGE`; without a scheduled importer, merchant actions will correctly disappear instead of serving stale prices.
13. Define the real evidence-review organization: name trained editors, independent reviewers and publishers; document acceptable source classes, confidence calibration, freshness periods, conflicts, withdrawal, and emergency unpublication; and require privileged-action MFA before real product revisions are published.
14. Build a reviewed evidence import/normalization workflow for real products. The implemented API is deliberately provider-neutral and human-governed; it does not scrape, infer, or invent missing product facts.

## Known limitations

- Rate limits are in-memory and replica-local. They reduce accidental abuse and direct brute force, but a production edge or distributed limiter must enforce global policy across replicas. The API must not be publicly reachable around that edge.
- A private/loopback reverse proxy may provide a single `X-Forwarded-For` IP. The supplied Nginx configuration overwrites, rather than appends, the untrusted client header. Different proxy topologies must define and test their trust boundary.
- The checked-in Compose file is a local development topology. It exposes PostgreSQL indirectly only to the Compose network, exposes the API for developer access, and runs the Vite development server. It is not the production orchestrator definition.
- PostgreSQL application and migration traffic currently use the same configured database credential in local Compose. Production must use separate least-privilege runtime, migration, backup, and reporting roles.
- API paths contain both foundational `/api/v1/health/*` probes and unversioned product endpoints. Existing contracts are documented, but a breaking public API will need an explicit versioning plan.
- Database migrations are forward-only. Rollback uses application rollback plus a new corrective migration, or database restoration for a destructive incident.
- Uploaded images are stored on a local persistent volume and served through the API. This is acceptable for one staging replica, not resilient multi-region production.
- The CSP permits inline scripts and styles for current JSON-LD and bounded dynamic presentation. External scripts, frames, plugins, and arbitrary connection targets remain blocked.
- No revenue, conversion, customer, or review data is fabricated. Seed content and products are explicitly fictional and must not be loaded in production.
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
| `API_PORT`                             | No                     | Internal listener, default `8080`.                                                                                      |
| `SESSION_COOKIE_NAME`                  | Yes                    | Stable cookie name containing only letters, numbers, `_`, or `-`.                                                       |
| `SESSION_COOKIE_SECURE`                | Yes                    | Must be `true` in production.                                                                                           |
| `SESSION_TTL`                          | Yes                    | Absolute session lifetime; default `720h`.                                                                              |
| `SESSION_IDLE_TTL`                     | Yes                    | Idle lifetime, positive and no longer than `SESSION_TTL`; default `168h`.                                               |
| `RATE_LIMIT_AUTH_PER_MINUTE`           | Yes                    | Default `10` per client. Confirm through security/load testing.                                                         |
| `RATE_LIMIT_RECOMMENDATION_PER_MINUTE` | Yes                    | Default `20` per client.                                                                                                |
| `RATE_LIMIT_ANALYTICS_PER_MINUTE`      | Yes                    | Default `120` per client.                                                                                               |
| `RATE_LIMIT_AFFILIATE_PER_MINUTE`      | Yes                    | Default `120` per client.                                                                                               |
| `RATE_LIMIT_MUTATION_PER_MINUTE`       | Yes                    | Default `240` per client for other mutations.                                                                           |
| `OFFER_MAXIMUM_AGE`                    | Yes                    | Default `72h`; older offers are hidden and cannot redirect until refreshed.                                               |
| `PRODUCT_IMAGE_UPLOAD_DIR`             | Only for local adapter | Persistent writable directory. Production should replace the adapter with object storage.                               |
| `MIGRATIONS_DIR`                       | Migration job          | Directory containing immutable migration SQL, default `./migrations`.                                                   |
| `SEEDS_DIR`                            | Development only       | Never invoke the seed command in production.                                                                            |
| `AI_PROVIDER`                          | Yes                    | Keep `disabled` until a reviewed provider adapter is enabled.                                                           |
| `AI_MODEL`                             | If AI enabled          | Server-side configured model identifier.                                                                                |
| `AI_API_KEY`                           | If AI enabled, secret  | Secret-manager value; never use a `VITE_*` name.                                                                        |
| `AI_TIMEOUT`                           | If AI enabled          | Default `15s`.                                                                                                          |
| `AI_MAX_RESPONSE_BYTES`                | If AI enabled          | Default `65536`, valid range 1 KiB–1 MiB.                                                                               |

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

- Select business-approved RPO and RTO values before launch; they are intentionally not invented here.
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

The release gate for this audit is:

```bash
npm --prefix frontend run format:check
npm --prefix frontend run lint
npm --prefix frontend run typecheck
npm --prefix frontend run test
npm --prefix frontend run build

cd backend
go test ./...
go vet ./...
go build ./...

docker compose --env-file .env.example config --quiet
docker compose --env-file .env.example build
docker compose --env-file .env.example up -d
```

The final audit report must record actual results rather than treating this checklist as evidence by itself.

### Audit evidence for this revision

- Frontend formatting, ESLint, TypeScript, 20 Vitest files / 40 tests, and the Vite production build passed.
- The largest emitted JavaScript chunk was approximately 80.62 KiB gzip; route pages remain split into lazy chunks.
- `npm audit --omit=dev --audit-level=high` reported zero vulnerabilities.
- `gofmt` verification, `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...` passed in the official Go 1.25 container.
- `govulncheck` v1.7.0 reported zero affected vulnerabilities and zero vulnerabilities in imported packages. It reported one advisory in a required module with no reachable symbols.
- Docker Compose configuration, builds, and startup passed. PostgreSQL, migrations, API readiness, and the Vite development container became healthy/available.
- A separate empty PostgreSQL 17 volume applied all 13 migrations. The development seed produced exactly 8 categories, 10 brands, 30 governed products, 90 offers, 90 affiliate links, one explicitly fictional verified evidence source, 326 fact-provenance links, 240 score rationales, and one active data-driven fitness policy. It creates no users.
- The full PostgreSQL integration suite passed. A post-suite isolation check still reported exactly 30 products, 90 offers, and zero users, analytics events, or affiliate clicks.
- Phase 1-specific tests verified review/publication separation of duties,
  completeness and freshness enforcement, immediate fail-closed source
  withdrawal, immutable candidate snapshots, and unchanged recommendation
  output after commercial commission/priority changes.
- Phase 2-specific tests verify policy approval/activation/retirement separation,
  active-policy immutability, unsupported-category exclusion, revision binding,
  compatibility/incompatibility, redundancy, room/access/clearance constraints,
  explicit overlap, missing-measurement rejection, deterministic output,
  historical policy-input persistence, and commercial-data independence.
- Live smoke tests verified `200` health/catalog/sitemap/robots responses, `401` account/admin guards, `403` hostile-origin mutation rejection, UNSOLERO frontend/editorial branding, and a complete register/session/logout flow.
- Production deployment, backup/restore, load, browser-matrix, and external monitoring evidence remain outstanding; this audit does not approve production readiness.
