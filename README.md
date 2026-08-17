# UNSOLERO

UNSOLERO is a trust-first fitness-equipment decision engine for people building a home gym. It is designed to turn goals, available space, budget, experience, and owned equipment into explainable purchasing decisions.

The current implementation includes the production-quality project foundation, the core relational data model, typed domain models, PostgreSQL repository adapters, password authentication backed by secure opaque sessions, a responsive public homepage, a public equipment catalog, a structured editorial-content foundation, a pure deterministic recommendation engine, versioned product evidence and score governance, reproducible recommendation candidate snapshots, the primary recommendation-builder experience, product comparison, wishlists, managed saved setups, offer-based affiliate redirects, a minimal first-party analytics pipeline, a role-protected operations dashboard, and a provider-neutral AI integration boundary. No live AI provider or public AI endpoint is enabled yet.

## Technology

- Frontend: React, Vite, TypeScript, Tailwind CSS, React Router, TanStack Query, React Hook Form, Zod, and Lucide React
- Backend: Go REST API and PostgreSQL through `pgx`
- Infrastructure: Docker, Docker Compose, forward-only SQL migrations, Nginx production image
- Quality: ESLint, Prettier, Vitest, Testing Library, `go test`, `go vet`, and `gofmt`

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the approved target architecture and dependency rules, and [PRODUCTION_READINESS.md](./PRODUCTION_READINESS.md) for the deployment, backup, monitoring, and known-limitation checklist.

## Repository layout

```text
backend/
├── cmd/api/                  API process
├── cmd/migrate/              explicit migration process
├── cmd/seed/                 opt-in fictional demo-data loader
├── internal/adapters/        PostgreSQL repository implementations
├── internal/modules/         domain models and repository ports
├── internal/platform/        configuration and database infrastructure
├── internal/transport/       HTTP handlers and routing
├── migrations/               immutable forward SQL migrations
└── seeds/                    idempotent, explicitly invoked demo data

frontend/
├── public/                   static assets
└── src/
    ├── app/                  router, providers, and query-client composition
    ├── components/           reusable presentation components
    ├── features/             API schemas, queries, and feature logic
    ├── pages/                route-level page composition
    └── test/                 shared test setup
```

Folders for future domain modules are created only when their first real use case is implemented.

## Branding and compatibility identifiers

The public product name is **UNSOLERO**. Existing lowercase `rigmark` values are
retained only where they are compatibility-sensitive internal identifiers: the
Go module/import path, npm package name, Compose project and local database
defaults, session-cookie and browser-storage keys, Nginx include/variable names,
demo reference codes and slugs, existing asset paths, and the health service
identifier. Renaming those values would invalidate imports, sessions, persisted
browser state, operational dashboards, or existing development data without
improving the customer-facing brand. New public copy must use UNSOLERO; any
future internal rename needs an explicit compatibility and migration plan.

## Prerequisites

For the container workflow, install Docker with Docker Compose. For direct host development, also install Node.js 22+ and Go 1.25+.

## Environment configuration

Create a local environment file:

```bash
cp .env.example .env
```

The checked-in example contains local-only placeholder credentials. Replace every secret for shared, staging, or production environments. Never commit `.env`.

Authentication uses the `SESSION_*` settings documented in `.env.example`. Local HTTP development disables the cookie's `Secure` flag; production configuration fails startup unless secure cookies are enabled.

`PUBLIC_SITE_URL` is the canonical public origin used for editorial canonical
URLs, the sitemap, and the robots sitemap directive. Production startup rejects
an insecure or path-bearing value.

AI integration is disabled by default. `AI_PROVIDER`, `AI_MODEL`, `AI_API_KEY`, `AI_TIMEOUT`, and `AI_MAX_RESPONSE_BYTES` are backend-only settings. Enabling a provider requires its adapter to be explicitly registered at the composition root; unsupported configuration fails startup. Never place model credentials in a `VITE_*` variable because those values are compiled into browser assets.

`PRODUCT_IMAGE_UPLOAD_DIR` configures the server-side product media adapter. Docker Compose mounts that path from the persistent `product_uploads` volume. A multi-replica production deployment should replace the local adapter with shared object storage through the existing application port.

`DATABASE_URL` points to `localhost` for host-run Go commands. Docker Compose constructs the container-only database URL from `POSTGRES_DB`, `POSTGRES_USER`, and `POSTGRES_PASSWORD`, so application code is not coupled to Compose DNS. `API_PORT` controls the host-side port; the API container consistently listens on port 8080.

## Run with Docker Compose

```bash
cp .env.example .env
docker compose up --build
```

Compose starts services in dependency order:

1. PostgreSQL becomes healthy.
2. The one-shot migration container applies pending migrations.
3. The Go API becomes healthy.
4. The Vite development server starts with an internal proxy to the API.

Open:

- Web application: `http://localhost:5173`
- API health: `http://localhost:8080/api/health`

Stop containers without deleting the PostgreSQL volume:

```bash
docker compose down
```

## Run directly on the host

Install locked dependencies:

```bash
make install
```

Start PostgreSQL, apply migrations, and run the API:

```bash
docker compose up postgres -d
set -a && source .env && set +a
make migrate
make dev-api
```

In another terminal, load the same environment and run the frontend:

```bash
set -a && source .env && set +a
make dev-web
```

## Fictional demo data

Demo data is never loaded during normal application startup. Load it explicitly after migrations:

```bash
docker compose --env-file .env --profile tools run --rm seed
```

For host-run development with the environment already loaded:

```bash
make seed
```

The idempotent seed contains exactly 8 categories, 10 brands, 30 products, 30 illustrative product images, 3 merchants, 90 merchant offers, and 90 affiliate-link records. All product and company names are prefixed `Demo`, descriptions identify the records as fictional, merchant URLs use the reserved `.invalid` domain, and every published demo fact and score links to a verified `demo_fixture` evidence source explicitly marked fictional. The dataset contains no users, customer information, reviews, ratings, or claims about real products.

## Database migrations

Migration files live in `backend/migrations` and use the form `NNNNNN_description.sql`. Applied filenames and SHA-256 checksums are recorded in `platform.schema_migrations`. The runner uses a PostgreSQL advisory lock, runs each migration transactionally, and refuses modified migration history.

Create a new migration by adding the next immutable SQL file. Do not edit an applied migration; add a new corrective migration.

```bash
make migrate
```

The core schema uses integer minor units for money, millimeters for dimensions, grams for weight/capacity, and explicit 0–100 score columns. Recommendation-critical product facts are columns on `catalog.products`; heterogeneous category facts use the typed and constrained `catalog.product_attributes` table. Analytics event properties are the only intentionally flexible JSON payload in this phase.

## API health contract

`GET /api/health` checks the API and its required PostgreSQL dependency.

Healthy response (`200 OK`):

```json
{
  "status": "ok",
  "service": "rigmark-api",
  "version": "development",
  "checks": {
    "database": "ok"
  }
}
```

A dependency failure returns `503 Service Unavailable`, changes `status` to `degraded`, and identifies the unavailable check without exposing internal connection details.

## Authentication

The browser authenticates through an opaque `HttpOnly`, `SameSite=Lax` session cookie. Raw session tokens are never stored in PostgreSQL or returned in JSON; only SHA-256 token hashes are persisted. Passwords use Argon2id with bounded parameters and are automatically rehashed after a successful login when configured costs change.

Available endpoints:

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `GET /api/auth/me`

See [API.md](./API.md) for request, response, cookie, and error contracts.

## Public catalog

The public storefront is available at `/products`, with detail, category, and brand routes. Catalog state is represented in the URL so searches, filters, sorting, and pagination remain linkable and browser-navigation friendly.

The API exposes structured product facts and suitability scores, but no review rating because the demo dataset contains no review records. Offer responses expose only a same-origin `/api/affiliate/click/{offerID}` purchase path. The endpoint resolves an active link on the server, records click attribution and an `affiliate_clicked` event in one database statement, then redirects. Raw destinations, provider references, and commission metadata never appear in catalog JSON. Affiliate data is owned by commerce and cannot participate in catalog or recommendation ranking.

See [API.md](./API.md) for catalog query parameters and endpoint contracts.

## Editorial acquisition foundation

Articles, guides, buying guides, and editorial product comparisons live in the
separate `editorial` database schema and `content` backend module. Content is a
validated sequence of safe presentation blocks; arbitrary stored HTML is not
rendered. Authors, related catalog products, related categories, and curated
related reading use normalized relationships. Catalog products remain the only
source for product price and specification facts.

Public routes include `/articles`, `/articles/:slug`, `/guides`,
`/guides/:slug`, and `/compare/:slug`. Published editorial pages provide
canonical URLs, Open Graph and Twitter metadata, Article JSON-LD, internal
catalog links, and related reading. Category pages are indexable and expose
curated editorial content. The backend generates `/sitemap.xml` and
`/robots.txt` from published state using `PUBLIC_SITE_URL`.

The demo seed contains four reviewed example pieces only. The foundation is
intentionally designed for a small, high-quality editorial library rather than
bulk AI publication. The protected Content admin page is a read-only published
inventory in this phase; authoring and revision approval remain a future
workflow rather than an unsafe raw-HTML editor.

## Deterministic recommendation engine

The pure engine lives under `backend/internal/modules/recommendation/domain`. It accepts normalized user constraints plus bounded catalog candidate snapshots, applies hard eligibility and compatibility rules, calculates an explicit eleven-dimension score breakdown, assembles a budget-constrained non-redundant setup, and returns structured reasons, alternatives, and rejections.

The default policy is versioned as `home-gym-v1`; every weight can be configured through `recommendation.Config`. Stable product-ID tie-breaking, canonical input fingerprints, and candidate-order normalization make results replayable. Free text is validated and fingerprinted but deliberately does not affect scoring until a separately validated interpretation layer exists.

Recommendation inputs contain no merchant, commission, sponsorship, payout, or affiliate fields. Commerce enrichment remains downstream of the finalized result. The consumer builder is available at `/build`; it validates eight focused steps, preserves guest progress for the browser session, saves authenticated drafts, and renders complete recommendations, alternatives, explicit rejections, and tracked merchant actions.

Authenticated generation atomically stores the immutable recommendation input/result and creates a planning setup. Saved setups can be listed at `/setups`, reopened, renamed, deleted, compared, or used as the input for a new edited revision. Guest generation remains stateless on the server by design; guests can deliberately save setup results in browser local storage.

The comparison workspace at `/compare` supports two to four products across structured specifications and suitability scores. `/wishlist` shows saved equipment with current catalog prices and tracked merchant actions. Authenticated comparison and wishlist IDs persist in PostgreSQL; guests use the same selection rules through a local-storage adapter, so presentation logic is not duplicated.

## AI integration boundary

The backend AI module exposes a provider-neutral `AIProvider` port for understanding user text, extracting requirements, asking clarifying questions, planning recommendation explanations, refining requirements, and planning product comparisons. Provider selection is configuration-driven through a registry so OpenAI, Anthropic, Gemini, or another provider can be added as an adapter without changing application contracts.

The provider has no repository or database capability. Requirement extraction returns only validated planning constraints. Clarifying output selects an allowlisted prompt key rendered from trusted localized copy, rather than model-authored product prose. Refinement returns a revised requirement draft that must be confirmed and sent back through the deterministic engine. Explanation and comparison outputs are reference plans containing only allowlisted product IDs, deterministic reason codes, and structured fact keys; canonical names, prices, specifications, scores, and order are rendered from the application-owned snapshot after validation. Unknown JSON fields, trailing JSON, oversized responses, unavailable facts, invented IDs, invented reason codes, and product reordering are rejected.

The disabled provider preserves the existing structured-form and deterministic-template fallback. No SDK dependency, external model call, AI route, or frontend secret was introduced in this phase. A live adapter must add provider-specific transport, bounded timeouts, response-schema prompting, operational metrics, and recorded-response tests before it is registered.

## Affiliate commerce and product analytics

Every active product offer may have one or more provider-neutral affiliate links. The commerce adapter supports a provider key, external link reference, program identifier, internal commission metadata, and operator priority without exposing those fields to clients. This permits later Amazon Associates, Awin, Impact, CJ, or direct-program adapters without changing the redirect or recommendation contracts.

Merchant actions are labeled **View at Merchant**, because UNSOLERO does not control checkout. Product, wishlist, recommendation, comparison, and setup surfaces attach bounded first-party attribution to the same-origin offer path and display an affiliate disclosure. The server accepts only known sources, validates an active merchant/offer/link and an HTTPS destination, associates the current account when authenticated or a per-tab anonymous session otherwise, records source/campaign/referrer, and returns a temporary redirect. The legacy `/api/out/{affiliateLinkID}` route remains temporarily available for old clients.

Offers older than `OFFER_MAXIMUM_AGE` (72 hours by default) are not displayed and cannot redirect until a trusted import refreshes them. Recommendation attribution is accepted only for the authenticated owner and only when the clicked product belongs to that recommendation.

The typed analytics boundary accepts `page_view`, `onboarding_started`, `onboarding_completed`, `recommendation_generated`, `product_viewed`, `product_saved`, `comparison_created`, and `setup_saved` only after explicit optional-analytics consent. The browser sends `consent_state: granted`; the backend rejects missing or different states. Users can reopen analytics preferences from the footer. `affiliate_clicked` is an essential, server-authored record inside the redirect transaction and cannot be spoofed through the browser endpoint. Each event has an exact property allowlist; email, free text, affiliate URLs, commission data, query strings, and client-supplied user IDs are rejected or structurally unavailable. First-touch source/medium/campaign and the external referrer hostname are stored in bounded structured columns.

The admin dashboard reports only observed data: users, recommendation sessions, paired onboarding completion, a product-view-based affiliate CTR, product/merchant/category rankings, and traffic sources. Missing denominators and empty rankings render as **No data**. `commerce.affiliate_conversions` and `analytics.acquisition_costs` are intentionally empty targets for future verified provider imports; no conversion, commission, spend, EPC, CAC, or revenue metric is displayed until that data exists.

Infrastructure also uses separate liveness and readiness probes:

- `GET /api/v1/health/live`
- `GET /api/v1/health/ready`

## Administration

The protected dashboard is available at `/admin`. Authentication alone is insufficient: every `/api/admin/*` request resolves current role membership from PostgreSQL and requires the `admin` role. The frontend route guard improves navigation, while the backend remains authoritative.

No account becomes an administrator automatically. After registering the initial operator, grant the role through a controlled database session:

```bash
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
  -c "INSERT INTO identity.user_roles (user_id, role_key) SELECT id, 'admin' FROM identity.users WHERE email = 'operator@example.com' ON CONFLICT DO NOTHING;"
```

Sign out and back in after changing role membership. The dashboard provides real inventory and event totals, draft product creation/editing, governed evidence inspection, JPEG/PNG/WebP upload or external image management, structured attributes, offer and affiliate-link management, recommendation score/reason inspection, a published editorial inventory, and read-only category, brand, merchant, user, and event views. Mutations are actor-attributed in `admin.audit_log`. Settings deliberately shows **No data** until a durable runtime-settings model exists.

Recommendation-critical publication uses two additional least-privilege roles.
`evidence_editor` creates sources, observations, and draft revisions;
`evidence_reviewer` verifies sources and approves/publishes revisions. The
repository rejects self-review and requires publication by someone other than
the approving reviewer, so a real workflow needs at least three staff accounts.
The ordinary `admin` role can inspect provenance at `/admin/evidence` but cannot
create or publish evidence unless the relevant additional role is granted.

```sql
INSERT INTO identity.user_roles (user_id, role_key)
SELECT id, 'evidence_editor' FROM identity.users
WHERE email = 'editor@example.com' ON CONFLICT DO NOTHING;

INSERT INTO identity.user_roles (user_id, role_key)
SELECT id, 'evidence_reviewer' FROM identity.users
WHERE email IN ('reviewer@example.com', 'publisher@example.com')
ON CONFLICT DO NOTHING;
```

Published product facts are read-only in the ordinary product editor. Changes
must be submitted as a new fact/score revision with complete per-field
provenance and per-score rationale. Manufacturer documents publish as claims,
independent tests as verified facts, merchant sources as observations, and
editorial/demo sources as assessments; mismatched labels fail publication.
Direct activation is intentionally absent.

## Quality commands

```bash
make typecheck       # TypeScript compiler
make lint            # ESLint and go vet
make format          # Prettier and gofmt
make format-check    # formatting verification
make test            # frontend and Go unit tests
make build           # production frontend and Go binaries
make check           # all checks above except mutation formatting
make compose-config  # validate Compose configuration
```

Build the production frontend image independently with:

```bash
docker build --target production -t unsolero-web ./frontend
```

## Current boundaries

- HTTP handlers depend on an application service interface and never query PostgreSQL.
- Health dependency checks live in a module application service.
- PostgreSQL migration and opt-in seed mechanics live under platform infrastructure.
- Catalog, commerce, identity, planning, recommendation, and analytics domain types do not import PostgreSQL.
- Repository interfaces are owned by consuming modules; PostgreSQL implementations live under adapters.
- Affiliate links belong to commerce and are absent from catalog product scores.
- Recommendation production packages are architecture-tested against commerce and analytics imports; commissions cannot enter candidate snapshots or configured score weights.
- Evidence production packages are architecture-tested against AI, commerce, and analytics imports; only reviewed human workflows can publish canonical product data.
- Public catalog and recommendation candidate reads require published fact and score revision pointers and reject expired observations or subsequently withdrawn sources.
- Completed authenticated recommendations persist the entire commercial-free candidate universe, revision IDs, constraints, policy version, engine version, and result fingerprint needed for historical reproduction.
- Affiliate provider selection happens only after an offer is chosen by consumer price and merchant eligibility. Commission metadata is stored for reporting but is absent from every ordering expression.
- Product analytics uses a typed frontend dispatcher, validated backend service, and replaceable recorder/reporting ports; direct browser `affiliate_clicked` events are rejected.
- Administrative authorization uses database-backed roles resolved on each authenticated request; no role or password hash is trusted from client input.
- Product and commerce mutations pass through the admin application/repository boundary and write an actor-attributed audit entry in the same transaction.
- Configuration is validated once at startup and injected into services.
- Authentication handlers delegate to the identity application service; reusable middleware resolves authenticated principals before protected handlers run.
- Password hashing and opaque-token generation are adapters, so a future OpenID Connect login can reuse the same server-session and authorization boundary.
- Future recommendation code cannot depend on commerce, affiliate, analytics, or AI adapters.
- The implemented deterministic recommendation engine depends only on planning and catalog domain facts; it has no HTTP, PostgreSQL, commerce, analytics, affiliate, or AI imports.
- The AI application layer is downstream of recommendation, has no database/repository dependency, and can only reference canonical recommendation facts through validated identifiers.
- AI explanations and comparisons are structured reference plans rather than model-authored product facts; strict decoding rejects fields outside their response contracts.
- Recommendation HTTP handlers delegate to an orchestration service; the PostgreSQL adapter owns the atomic completed-run/setup transaction and never performs ranking.
- Public catalog search is application-owned; handlers parse HTTP, services validate use cases, and the PostgreSQL adapter owns query construction.
- Wishlist commands are authenticated and live behind the planning repository boundary.
- Product comparison is a separate ordered planning resource; recommendation quality and commerce data are not inputs to it.

## Frontend design system

Reusable tokens, controls, feedback states, commerce presentation, and responsive site chrome live under `frontend/src/components` and `frontend/src/design-system`. A temporary visual inventory is available at `/design-system`; see [the design-system notes](./frontend/src/design-system/README.md).
