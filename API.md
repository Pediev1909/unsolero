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

Returns `201 Created`, establishes the session cookie, and returns:

```json
{
  "user": {
    "id": "opaque-user-id",
    "email": "person@example.com",
    "roles": []
  }
}
```

Registration normalizes the email address. Roles are resolved from server-owned membership and never accepted from registration or login input. Duplicate email addresses return `409 email_already_registered`. Invalid fields return `422 validation_failed`.

### Login

`POST /api/auth/login`

Uses the same request and success response as registration, returning `200 OK`. Invalid credentials and unavailable accounts share the `401 invalid_credentials` response to avoid account-state disclosure.

### Current account

`GET /api/auth/me`

Requires a valid session and returns `200 OK` with the user response above. Missing, expired, idle-expired, revoked, or malformed sessions return `401 authentication_required`.

### Logout

`POST /api/auth/logout`

Revokes the supplied server session, clears the cookie, and returns `204 No Content`. The operation is idempotent when no session exists.

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

## Public catalog

Catalog responses are public JSON. Current fictional demo records include `is_demo: true`; clients must preserve that label. Suitability and quality fields are structured product scores, not customer ratings or reviews.

### Products

- `GET /api/catalog/products`
- `GET /api/catalog/products/{slug}`

The collection accepts:

- `q`: product or brand search, up to 100 characters
- `category`, `brand`: lowercase slugs
- `min_price_minor`, `max_price_minor`: non-negative integer minor units
- `sort`: `featured`, `name_asc`, `price_asc`, `price_desc`, `quality_desc`, or `value_desc`
- `page`: positive integer
- `page_size`: positive integer, maximum 48
- `ids`: comma-separated product UUIDs, maximum 48; results preserve the requested order

The collection response contains `products`, `page`, `page_size`, `total`, and `total_pages`. Product detail adds dimensions, weight, capacity, material, warranty, typed attributes, strengths, weaknesses, use cases, and same-category alternatives.

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

The content collection accepts `section=articles|guides|comparisons|all`, an
optional related category `slug`, and `limit` from 1 through 24. Only published
entries are returned. Guides include both guide and buying-guide content types.

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

- `GET /api/catalog/products/{slug}/offers`
- `GET /api/affiliate/click/{offerID}`

Offers include merchant identity and trust score, item price, shipping, landed price, availability, condition, last-check time, disclosure label, and a same-origin `purchase_path`. Responses never include merchant product URLs, affiliate destinations, or affiliate provider references.

The redirect endpoint resolves an active affiliate link attached to an active, available offer from an active merchant, records the commerce click and `affiliate_clicked` analytics event, then returns a `302` redirect to the validated HTTPS destination. It accepts optional `source`, `session_id`, `campaign`, `traffic_source`, `traffic_medium`, and `recommendation_id` attribution parameters. `source` is restricted to `product_detail`, `wishlist`, `recommendation`, `comparison`, or `setup`; attribution tokens and identifiers are validated, and the server generates an anonymous session identifier when one is absent. The authenticated user is associated through the session cookie, never through a query parameter.

`Referer` is reduced to its scheme, host, and path before storage. Responses use `Cache-Control: no-store` and do not expose provider, destination, program, or commission records. Affiliate commission and click data are not inputs to offer ordering, catalog scores, or recommendation scores. `GET /api/out/{affiliateLinkID}` remains as a compatibility route for previously issued paths but is not returned by current APIs.

## Product analytics

`POST /api/analytics/events`

The first-party endpoint accepts a versioned, allowlisted interaction envelope:

```json
{
  "name": "product_viewed",
  "surface": "product_detail",
  "session_id": "1191bb26-a9a2-41df-9346-74d693350ce8",
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

Accepted browser events are `page_view`, `onboarding_started`, `onboarding_completed`, `recommendation_generated`, `product_viewed`, `product_saved`, `comparison_created`, and `setup_saved`. Each has an exact property schema; unknown names, properties, nested context fields, or envelope fields are rejected. Account identity is derived only from the secure session. `affiliate_clicked` is server-authored by the redirect flow and is deliberately rejected here. Page paths exclude query strings, referrers are reduced to hostnames, and source/medium/campaign values are bounded. Successful ingestion returns `204 No Content`.

## Administration

All endpoints below require a valid session whose current database-backed roles include `admin`. Missing authentication returns `401`; an authenticated account without the role receives `403 permission_denied`. Responses use `Cache-Control: no-store`, inputs reject unknown JSON fields, and mutation identity is derived only from the secure session.

- `GET /api/admin/dashboard`
- `GET /api/admin/analytics`
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
- `GET /api/admin/recommendations`
- `GET /api/admin/recommendations/{recommendationID}`
- `GET /api/admin/users`
- `GET /api/admin/events`

Product image creation accepts either strict JSON for a validated external URL or `multipart/form-data` with `file`, `alt_text`, `sort_order`, and `is_primary`. Uploaded JPEG, PNG, and WebP files are limited to 5 MB and returned from an immutable same-origin `/api/media/products/{file}` path. SVG and executable formats are rejected.

Recommendation inspection exposes persisted engine/policy versions, objective and dimension scores, selected/alternative/rejected products, and deterministic reasons. It never exposes password hashes, session tokens, or affiliate commission as scoring data. Event and dashboard values are direct database counts; absent observations are returned as zero and rendered as “No data.”

The analytics report returns observed users, persisted recommendation sessions, paired onboarding starts/completions, recommendation completion rate, product views, affiliate clicks, affiliate CTR, product/merchant/category rankings, and traffic sources by distinct page-view session. Completion pairs unique onboarding IDs. Affiliate CTR uses observed product/session view pairs as its denominator and only matching product-detail click pairs as its numerator. A missing denominator returns `null`, rendered as “No data”; rates and revenue are never estimated.

## Saved equipment

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
