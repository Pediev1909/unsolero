# UNSOLERO API

Status: implemented authentication, public catalog, editorial content, recommendation, setup, affiliate, and analytics contracts  
Last reviewed: 2026-08-16

UNSOLERO exposes same-origin JSON endpoints under `/api`. Authentication responses are private and send `Cache-Control: no-store`. Unknown JSON fields are rejected.

## Error format

```json
{
  "error": {
    "code": "validation_failed",
    "message": "Check the highlighted fields.",
    "fields": {
      "email": "Enter a valid email address."
    }
  }
}
```

`fields` is omitted when an error does not apply to a specific input.

## Authentication

Authentication uses an opaque cookie. The cookie is `HttpOnly`, `SameSite=Lax`, scoped to `/`, and `Secure` outside local HTTP development. Clients must send requests with credentials enabled. Secret tokens and password hashes never appear in JSON.

### Register

`POST /api/auth/register`

```json
{
  "email": "person@example.com",
  "password": "a password of at least 12 characters"
}
```

Returns the same `202 Accepted` response whether the address is new or already registered. Registration does not establish a browser session. If the address is eligible, a verification delivery intent is recorded:

```json
{
  "recorded": true,
  "message": "If the address is eligible, an account verification intent has been recorded."
}
```

Registration normalizes the email address. Roles are resolved from server-owned membership and never accepted from registration or login input. Invalid fields return `422 validation_failed`.

### Login

`POST /api/auth/login`

Uses the credential request above and returns `200 OK` with `user.id`, `user.email`, server-resolved `user.roles`, `user.email_verified`, and `user.mfa_enabled`. Invalid credentials and unavailable accounts share `401 invalid_credentials`. An MFA-enabled account returns `202` with `mfa_required: true`; the challenge secret is held only in a scoped `HttpOnly` cookie. Complete it with `POST /api/auth/mfa/complete` and `{ "code": "123456" }`.

### Current account

`GET /api/auth/me`

Requires a valid session and returns `200 OK` with the user response above. Missing, expired, idle-expired, revoked, or malformed sessions return `401 authentication_required`.

### Logout

`POST /api/auth/logout`

Revokes the supplied server session, clears the cookie, and returns `204 No Content`. The operation is idempotent when no session exists.

### Verification and recovery

- `POST /api/auth/email-verification/request` — `{ "email": "..." }`; always generic `202`.
- `POST /api/auth/email-verification/complete` — `{ "token": "..." }`; returns `200`, `400 token_invalid`, or `410 token_expired|token_used`.
- `POST /api/auth/password-reset/request` — `{ "email": "..." }`; always generic `202`.
- `POST /api/auth/password-reset/complete` — `{ "token": "...", "password": "..." }`; consumes the token once, replaces the Argon2id hash, revokes every session, and returns `204`.

### Account security

All endpoints derive ownership from the authenticated session and never accept a user ID:

- `POST /api/account/security/password` — current/new password; keeps the current session and revokes all others.
- `GET /api/account/security/sessions` — stable session metadata only.
- `DELETE /api/account/security/sessions/{sessionID}` — revokes one owned session; cross-account IDs return `404`.
- `DELETE /api/account/security/sessions` — revokes every session except the current one.
- `GET /api/account/export` — structured owned-data JSON; no credentials or internal secrets.
- `DELETE /api/account` — current password plus `confirmation: "DELETE"`; anonymizes retained history and clears the session.
- `POST /api/account/security/mfa/enroll` — current password; returns the TOTP secret once.
- `POST /api/account/security/mfa/verify` — TOTP code; enables MFA and returns recovery codes once.
- `POST /api/account/security/mfa/recovery-codes` — current TOTP/recovery code; replaces all recovery codes.
- `POST /api/account/security/mfa/step-up` — records bounded recent MFA on the current server session.

Development-only `GET /api/dev/email-deliveries` exposes local delivery intents only when the application is in development with the development adapter. It is absent in production.

## Request protections

- Authentication request bodies are limited to 16 KiB and must use `application/json`.
- State-changing requests with an `Origin` header must match the request host.
- Session tokens contain 256 bits of cryptographic randomness; PostgreSQL stores only their SHA-256 hashes.
- Sessions enforce both absolute and sliding idle expirations.
- Backend authorization is authoritative. Frontend protected routes are a user-experience boundary only.

## Health

- `GET /api/health`
- `GET /api/v1/health/live`
- `GET /api/v1/health/ready`

See the README for the health response contract.

When `METRICS_ENABLED=true`, `GET /api/v1/metrics` requires an exact
`Authorization: Bearer <METRICS_TOKEN>` credential and returns only
process-local, bounded-cardinality operational aggregates. The route is absent
when metrics are disabled. It is not product analytics and contains no raw
request path, client address, user agent, body, query, or account identifier.

## Public catalog

Catalog responses are public JSON. Current fictional demo records include `is_demo: true`; clients must preserve that label. Suitability and quality fields are structured product scores, not customer ratings or reviews.

### Products

- `GET /api/catalog/products`
- `GET /api/catalog/products/{slug}`

The collection accepts:

- `q`: product or brand search, up to 100 characters
- `category`, `brand`: lowercase slugs
- `min_price_minor`, `max_price_minor`: non-negative integer minor units
- `has_offer`: `true` keeps only products with a live vendor offer, under the same active, available, fresh and unexpired offer-plus-link conditions that set `purchase_path` and that the redirect re-checks; `total` and `total_pages` count the filtered set. Leave it out to switch the filter off — any value other than `true` is rejected as `invalid_catalog_query`
- `sort`: `featured`, `name_asc`, `price_asc`, `price_desc`, `quality_desc`, or `value_desc`
- `page`: positive integer
- `page_size`: positive integer, maximum 48
- `ids`: comma-separated product UUIDs, maximum 48; results preserve the requested order

The collection response contains `products`, `page`, `page_size`, `total`, and `total_pages`. Product detail adds dimensions, weight, capacity, material, warranty, typed attributes, strengths, weaknesses, use cases, and same-category alternatives.

#### Price and billing basis

Every product summary and detail carries `price` and, beside it, `billing`:

```json
"price": {"amount_minor": 2000, "currency": "USD"},
"billing": {"period": "monthly", "unit": "per_user", "unit_note": null, "annual_price_minor": 1500}
```

`price` is the compared figure — the one budget filters, sorting, value and
the cheaper/premium alternatives use — and `billing` says what it is a price
for. `period` is one of:

- `monthly`: the vendor sells monthly and `price` is the monthly-billing list
  price. `annual_price_minor`, when not null, is the per-month equivalent of
  the annual contract; it is shown, never compared, and is present only when
  the vendor offers both.
- `annual`: the vendor offers no monthly billing; `price` is the per-month
  equivalent of an annual contract and `annual_price_minor` is null.
- `free`: the product line's entry tier is free and `price.amount_minor` is 0.
- `usage`: the price is usage-based (payments, automation tasks); `price` is
  the entry figure or 0 and `unit_note` explains.

`unit` is one of `flat`, `per_user`, `per_contacts`, `per_transaction` or
`usage`; `unit_note` is an optional qualifier of up to 120 characters ("at
1,000 contacts", "per seat, minimum 3 seats", "2.9% + 30¢ per transaction").

For a non-physical product `key_specification` is derived from `billing` as
`{"label": "Billing", "value": <unit phrase>, <period phrase>}`. Unit phrases:
`flat` → "Flat rate"; `per_user` → "Per user"; `per_contacts` → the note if
present else "Per contact tier"; `per_transaction` → the note if present else
"Per transaction"; `usage` → the note if present else "Usage-based". Period
phrases: `monthly` → "monthly billing"; `annual` → "billed yearly"; `free` →
"free plan"; `usage` → omitted. Examples: "Per user, monthly billing", "Flat
rate, billed yearly", "At 1,000 contacts, monthly billing", "2.9% + 30¢ per
transaction". The same phrase follows the price in the server-rendered product
body.

#### Price record

Product **detail** adds `price_record`: what this product has cost, newest
first, built from its own immutable fact revisions.

```json
"price_record": [
  {"observed_at": "2026-09-02T09:28:00Z", "price_minor": 3900, "currency": "USD",
   "billing": {"period": "monthly", "unit": "flat", "unit_note": "per account", "annual_price_minor": 2900},
   "is_current": true},
  {"observed_at": "2026-08-21T11:05:00Z", "price_minor": 2900, "currency": "USD",
   "billing": null, "note": "Price read from the vendor pricing page on 2026-08-21.",
   "is_current": false}
]
```

- `observed_at` is the revision's publication date (its creation date if it was
  never stamped), and is the **earliest** date the entry's figure was recorded:
  consecutive revisions repeating the same claim are collapsed into one entry,
  so a row says what the price has been and since when. A revision is written
  whenever any fact changes — the billing-basis audit of 2026-09-02 wrote one
  for all fifty-three products — and a row repeating the number above it would
  be noise rather than history.
- `billing` has the same shape as the product's own `billing`, or is `null`
  where that revision stated no basis. Every revision published before
  2026-09-02 predates the billing columns and states none; the basis is not
  guessed at from the current one. Where a run of collapsed revisions contains
  one that states a basis, the entry carries the most recently stated.
- `note` is absent unless a reviewer wrote a sentence worth publishing. It is
  their first sentence, unedited; notes that only restate the basis the entry
  already carries are left out.
- `is_current` is true on the first entry only, and its figure always equals
  the product's `price`.
- Only revisions that were published (`published` or `superseded`) appear.
  Drafts and rejected revisions are prices nobody ever saw. Entries are capped
  at the ten most recent distinct figures.
- The list is empty or one entry long for most products; a single figure is not
  a history and clients should render nothing rather than a one-row table.

### Categories and brands

- `GET /api/catalog/categories`
- `GET /api/catalog/categories/{slug}`
- `GET /api/catalog/brands`
- `GET /api/catalog/brands/{slug}`

Only active categories and brands are public.

## Editorial content and discovery

- `GET /api/content`
- `GET /api/content/{slug}`
- `GET /sitemap.xml`
- `GET /robots.txt`

The content collection accepts `section=articles|guides|comparisons|stacks|all`,
an optional related `category` slug, an optional `product` slug, and `limit`
from 1 through 24. Only published entries are returned. Guides include both
guide and buying-guide content types; `stacks` lists the `stack` type, a whole
set of tools for one kind of business and budget, served at `/stacks/{slug}`. `product={slug}` returns the entries whose
product references include that product — the "Compared in" list on a product
page — in the same summary shape; a value that is not a slug is rejected with
`400 invalid_content_query`.

Content detail returns validated structured blocks, author information,
publication and update dates, related published products, related active
categories, curated related editorial entries, and explicit SEO metadata.
Arbitrary HTML is neither accepted nor returned. Related product facts are
resolved from the current published catalog so editorial records cannot become
an alternative source of product prices or specifications.

`/sitemap.xml` includes indexable static pages, active category pages,
non-demo published products and brands, and every published editorial route.
`/robots.txt` identifies the sitemap and excludes private application and API
surfaces. Both use the server-controlled `PUBLIC_SITE_URL` origin.

### Merchant offers and tracked redirects

- `GET /api/catalog/offers`
- `GET /api/catalog/products/{slug}/offers`
- `GET /api/affiliate/click/{offerID}`

Offers include merchant identity and trust score, item price, shipping, landed price, availability, condition, observation/check time, optional expiry, freshness status, disclosure label, and a same-origin `purchase_path`. Responses never include merchant product URLs, affiliate destinations, credentials, or provider references.

`GET /api/catalog/offers` is the catalog-wide listing behind the `/offers` page: every published product that has a servable affiliate offer right now, under the same active/available/fresh/unexpired conditions the redirect re-checks. It returns `{ items: [{ product, offer }], generated_at }`, where `product` is the catalog summary used by the product listing (with `purchase_path`, `merchant_name` and `disclosure_label` set) and `offer` is `{ price: { amount_minor, currency }, merchant_name, last_checked_at, freshness_status }`. Items are sorted by category name, then in catalog order. The list is assembled from the published catalog and one batched commerce lookup; a commerce failure returns `500 offers_unavailable` rather than an empty list, and a catalog with no live offers returns `items: []`. Cached with `Cache-Control: public, max-age=30`, the same as the per-product endpoint.

The redirect resolves an active, available, fresh/unexpired offer, validates recommendation ownership when supplied, and validates the HTTPS destination before click persistence. A click/analytics write or optional-session lookup failure does not block an already resolved `302`. Obvious bots/prefetches remain raw but do not emit filtered `affiliate_clicked` events. `X-Request-ID` is the bounded idempotency key. Source/session/campaign/traffic/recommendation values are normalized and validated; account identity comes only from the session cookie.

`Referer` is reduced to scheme and host before storage; paths, queries, fragments, and user information are discarded. Responses use `Cache-Control: no-store` and do not expose provider, destination, program, or commission records. Affiliate commission and click data are not inputs to offer ordering, catalog scores, or recommendation scores. `GET /api/out/{affiliateLinkID}` remains as a compatibility route for previously issued paths but is not returned by current APIs.

## Newsletter

A double opt-in list for dated software prices. Every route is public, JSON
only, limited to 16 KiB bodies, and covered by the generic mutation rate
limit; nothing here accepts or returns an account identifier.

- `POST /api/newsletter/subscriptions` — `{ "email": "...", "source": "footer" }`; always `202 { "recorded": true }` for well-formed input, whether the address is new, pending, or already confirmed. A confirmed address is not mailed again. `400 validation_failed` names the field (`email` or `source`); `400 invalid_json` and `415` follow the authentication body rules; `500 newsletter_unavailable` means nothing was recorded. `source` names the surface that asked (`^[a-z][a-z0-9_.:-]{0,99}$`, e.g. `footer`, `article:<slug>`).
- `POST /api/newsletter/confirmations` — `{ "token": "..." }`; consumes the one-time token from the email and returns `204`. Unknown, expired, malformed, and already-used tokens all return `400 invalid_token`.
- `POST /api/newsletter/unsubscriptions` — `{ "token": "..." }`; marks the address unsubscribed and returns `204`; repeating it returns `204` again. An unknown token returns `400 invalid_token`.

The confirmation email carries the token in the URL fragment
(`/newsletter/confirm#<token>`), as the account security links do, so it never
appears in an edge access log or a `Referer` header; the confirm page also
accepts `?token=` for clients that drop fragments. Tokens are 32 random bytes;
PostgreSQL stores only their SHA-256 hashes, and the confirmation hash is
cleared once used. The list stores the lower-cased address, status, source,
consent text version, and timestamps only: no IP address, user agent, or
account link. In development the confirmation intent is visible through
`GET /api/dev/email-deliveries` as kind `newsletter_confirmation`. Unsubscribe
tokens are issued at sign-up and spent at `/newsletter/unsubscribe#<token>`,
the same fragment form the confirmation link uses. The page exists and the
route works; no newsletter is sent yet, so nothing mails that link, and the
first send has to carry it before the form's "unsubscribe in one click" is a
promise the site can keep.

## Product analytics

`POST /api/analytics/events`

- `GET /api/analytics/consent`
- `PUT /api/analytics/consent`
- `POST /api/analytics/identity/claim` (authenticated)

The first-party endpoint accepts a versioned, allowlisted interaction envelope:

```json
{
  "event_id": "c3448188-244c-4b2a-9f97-53c1ad10a7ee",
  "name": "product_viewed",
  "surface": "product_detail",
  "session_id": "1191bb26-a9a2-41df-9346-74d693350ce8",
  "consent_version": "analytics-v1",
  "properties": {
    "product_id": "4ba7d524-9fd5-4d18-8c42-778c42d996f3"
  },
  "context": {
    "page_path": "/products/demo-bench",
    "traffic_source": "newsletter",
    "traffic_medium": "email",
    "campaign": "strength_launch"
  }
}
```

Accepted browser events are `page_view`, `onboarding_started`, `onboarding_completed`, `recommendation_generated`, `product_viewed`, `product_saved`, `comparison_created`, and `setup_saved`. Each has an exact property schema; unknown names, properties, nested context fields, or envelope fields are rejected. Account identity is derived only from the secure session. Anonymous identity comes from a server-issued random HttpOnly subject cookie whose SHA-256 digest is stored; the body cannot select identity. `affiliate_clicked` is server-authored by the redirect flow and rejected here. Page paths exclude query strings, referrers are hostnames, and attribution values are bounded. Successful and duplicate ingestion return the same `204 No Content` surface.

`PUT /api/analytics/consent` accepts `state` (`granted` or `denied`), `policy_version` (`analytics-v1`), and `source` (`banner`, `preferences`, or `account_sync`). A decline after a grant becomes `withdrawn`. The frontend cache is not authoritative: insertion locks and verifies current persisted consent/version. The claim endpoint accepts no body identifiers and requires matching current grants for the authenticated account and browser subject.

## Administration

All endpoints below require a valid session and their explicit current database-backed permission; administrators receive all permissions while scoped staff roles receive only their documented routes. Missing authentication returns `401`; insufficient permission returns `403 permission_denied`. Responses use `Cache-Control: no-store`, inputs reject unknown JSON fields, and mutation identity is derived only from the secure session.

- `GET /api/admin/dashboard`
- `GET /api/admin/analytics?from={RFC3339}&to={RFC3339}&limit={1..50}`
- `GET /api/admin/references`
- `GET|POST /api/admin/products`
- `GET|PATCH /api/admin/products/{productID}`
- `PUT /api/admin/products/{productID}/status`
- `POST /api/admin/products/{productID}/images`
- `DELETE /api/admin/products/{productID}/images/{imageID}`
- `PUT|DELETE /api/admin/products/{productID}/attributes/{key}`
- `GET /api/admin/categories`
- `GET /api/admin/brands`
- `GET /api/admin/merchants`
- `GET|POST /api/admin/offers`
- `PATCH /api/admin/offers/{offerID}`
- `GET /api/admin/affiliate-links`
- `PATCH /api/admin/affiliate-links/{linkID}`
- `GET|POST /api/admin/commerce/providers`
- `PUT /api/admin/commerce/providers/{providerID}/lifecycle`
- `GET|POST /api/admin/commerce/imports`
- `POST /api/admin/commerce/imports/{importID}/retry`
- `GET /api/admin/commerce/imports/{importID}/failures`
- `PUT /api/admin/commerce/providers/{providerID}/conversions`
- `GET /api/admin/commerce/conversions`
- `GET|POST /api/admin/commerce/conversion-imports`
- `POST /api/admin/commerce/conversion-imports/{importID}/retry`
- `GET /api/admin/commerce/reconciliations`
- `POST /api/admin/commerce/conversion-imports/{importID}/reconcile`
- `GET /api/admin/commerce/metrics`
- `GET /api/admin/recommendations`
- `GET /api/admin/recommendations/{recommendationID}`
- `GET /api/admin/users`
- `GET /api/admin/events`

Product bodies (`POST /api/admin/products`, `PATCH
/api/admin/products/{productID}`, and the `product` object of an evidence
revision) carry `price_minor`, `currency` and a required `billing` object with
the shape and semantics documented under the public catalog: `period`, `unit`,
`unit_note` (nullable) and `annual_price_minor` (nullable, only with
`period: "monthly"`; a `free` period requires `price_minor: 0`). Admin product
responses return the same `billing` object. A basis that fails validation is a
`422 invalid_admin_input` (or `invalid_evidence_input`) whose `fields` map
names the field, for example `{"billing.annual_price_minor": "..."}`. A
published fact revision replaces the product's whole billing basis; a revision
row with no `billing_period` — as a data seed may write — leaves the product's
current basis in place.

Product image creation accepts either strict JSON for a validated external URL or `multipart/form-data` with `file`, `alt_text`, `sort_order`, and `is_primary`. Uploaded JPEG, PNG, and WebP files are limited to 5 MB and returned from an immutable same-origin `/api/media/products/{file}` path. SVG and executable formats are rejected.

Recommendation inspection exposes persisted engine/policy versions, objective and dimension scores, selected/alternative/rejected products, and deterministic reasons. It never exposes password hashes, session tokens, or affiliate commission as scoring data. Aggregate analytics requires `analytics.read`; event-level `/api/admin/events` requires administrator-only `analytics.raw.read`.

The analytics report returns observed users, recommendation sessions, paired onboarding, product views, countable/raw click counts, rankings, traffic sources, campaign attribution, and payload-free ingestion outcomes. Campaign attribution is three arrays, each limited by `limit` and the window like the rankings: `campaigns` rows of `{campaign, traffic_source, traffic_medium, sessions, page_views, affiliate_clicks}` grouped by the UTM triple the visit arrived with; `landing_pages` rows of `{campaign, page_path, sessions}` counting the first campaign-bearing `page_view` of each session; and `sources_by_medium` rows of `{traffic_source, traffic_medium, sessions}`. `traffic_source` and `traffic_medium` are `null` when the link carried no such parameter. Campaign `sessions` and `page_views` come from consented `page_view` events; campaign `affiliate_clicks` count countable `affiliate_clicked` events by their stored `campaign`, `traffic_source`, and `traffic_medium`, which need no analytics consent, so a campaign may report clicks without sessions. Only `is_reportable` events and `is_countable` clicks enter metrics. It labels `no_data`, `insufficient_data`, or `available`, plus partial/complete coverage. Rates remain `null` below 20 eligible denominator observations. Windows are at most 366 days and limits 1–50. Revenue/conversion data remains in verified commerce reporting and is never inferred here.

`POST /api/webhooks/commerce/{providerConfigurationID}` is the provider-neutral
conversion webhook terminus. It accepts JSON up to 256 KiB and delegates
authentication, signature timestamp extraction, and structured parsing to the
configured adapter. Disabled or unknown adapters fail closed. Replays of a
processed verified delivery return `200`; accepted processing returns `202`.
Responses never expose credentials, raw payloads, or provider error details.

Conversion and metric operator routes are admin-only. Conversion lists expose
normalized verified facts and reconciliation status, never raw provider
payloads or credentials. Metrics return `available`, `no_data`, or
`insufficient_data` plus numerator, denominator, definition, reporting window,
freshness, and original-currency values. No FX conversion is performed.

## Saved products

These endpoints require an authenticated session:

- `GET /api/account/wishlist`
- `PUT /api/account/wishlist/{productID}`
- `DELETE /api/account/wishlist/{productID}`

Only published products can be saved. Responses contain product identifiers, never private catalog or authentication fields.

Guests use the same frontend selection model backed by browser local storage. Local state is never sent to these authenticated endpoints implicitly.

## Product comparison

These endpoints require an authenticated session:

- `GET /api/account/comparison`
- `PUT /api/account/comparison`

The resource is one ordered comparison selection per user. `PUT` accepts `{ "product_ids": [...] }` with zero to four distinct published product UUIDs and replaces the selection atomically. Guest comparison state uses the same client-side rules and remains in browser local storage.

## Personalized recommendations

Recommendation requests are limited to 64 KiB, require JSON, and reject unknown fields. Inputs contain only user constraints and preferences; merchant, affiliate, commission, and sponsorship fields are not accepted.

### Save or resume an authenticated draft

- `GET /api/recommendations/draft`
- `PUT /api/recommendations/draft`
- `DELETE /api/recommendations/draft`

All draft endpoints require authentication. `GET` returns `204 No Content` when no draft exists. One structured draft is stored per user; existing equipment is stored in child rows rather than an opaque JSON document. The current step is an integer from 1 through 8.

### Generate a setup

`POST /api/recommendations/generate`

The endpoint permits guests and optionally attaches a valid authenticated session. Example input:

```json
{
  "goal": "build_muscle",
  "experience": "beginner",
  "budget_minor": 70000,
  "currency": "USD",
  "available_space": {
    "length_mm": 2400,
    "width_mm": 1800,
    "height_mm": 2400,
    "apartment_living": true
  },
  "existing_equipment": [
    { "name": "Pull-up bar", "category_slug": "pull-up-bars" }
  ],
  "training_preferences": ["dumbbells", "bodyweight"],
  "priorities": ["compact", "budget"],
  "free_text": "I train early and share a wall with a neighbor."
}
```

The response contains the normalized `input`, `total_cost`, `recommendation_score`, the complete eleven-dimension `fit` breakdown, `recommended_products`, `alternatives`, `rejected_alternatives`, deterministic reason codes/messages, and policy/engine versions. Product objects use the same public summary contract as the catalog. Guest responses have `saved: false` and null recommendation/setup IDs.

For an authenticated request, the completed session, score breakdowns, item reasons, alternatives, explicit rejections, and setup items are committed atomically. The response has `saved: true` plus opaque `recommendation_id` and `setup_id` values, and the completed draft is deleted.

Free text is validated, saved, and included in the deterministic input fingerprint. It does not alter scores until a separately validated interpretation layer exists.

### Revisit saved setups

- `GET /api/account/setups`
- `GET /api/account/setups/{setupID}`
- `PATCH /api/account/setups/{setupID}` with `{ "name": "..." }`
- `DELETE /api/account/setups/{setupID}`

All endpoints require authentication. Collection items include name, item count, stored total, recommendation score, and timestamps. Detail reconstructs the saved recommendation while hydrating current public product presentation data and includes the original normalized input so it can be reopened for editing. Rename and delete queries always scope the setup by authenticated owner; another user's opaque ID returns `404 setup_not_found`.

Guest setup saves, renames, deletes, and reopens use the same frontend collection interface backed by browser local storage. Editing always creates a new deterministic recommendation revision; it does not mutate an immutable prior recommendation result.
