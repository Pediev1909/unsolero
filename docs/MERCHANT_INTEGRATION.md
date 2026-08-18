# Merchant integration and affiliate operations

Status: Phase 4 provider-neutral offer and verified-conversion foundation; no live provider enabled  
Last reviewed: 2026-08-17

UNSOLERO treats merchant data as downstream commerce enrichment. Merchant price,
availability, commission, attribution, campaigns, and performance are not
representable in recommendation policy or candidate inputs.

## Provider contract

Provider adapters implement `commerce/ports.ProviderAdapter`: `Key` identifies
the adapter, `ValidateConfiguration` verifies externally stored credentials, and
`FetchOffers` returns a normalized bounded page with cursor, completion state,
timestamps/expiry, and optional rate-limit state. Amazon Associates, Awin,
Impact, CJ, and direct-feed API details must remain inside their adapters.

No live adapter ships in this phase. An unknown adapter resolves to
`DisabledAdapter`, returns `ErrProviderDisabled`, and never creates offers.
PostgreSQL stores only a bounded `credential_reference`; access tokens, refresh
tokens, and provider secrets belong in the deployment secret manager.

## Provider lifecycle

| State | Imports | Historical observations | Public offers |
| --- | --- | --- | --- |
| `disabled` | stopped | retained | only while independently fresh and otherwise eligible |
| `configured` | stopped | retained | same freshness rule |
| `active` | scheduled/manual | appended | fresh available offers may appear |
| `degraded` | bounded retry/manual recovery | retained | failed runs never extend freshness |
| `suspended` | stopped | retained | same freshness rule |

Configured/active/degraded transitions require a credential reference and
successful adapter validation. Missing adapters or credentials fail closed.
Provider state never changes recommendation behavior.

## Imports, retries, and reconciliation

The worker queues due configurations, claims jobs with
`FOR UPDATE SKIP LOCKED`, and processes at most 100 pages / 10,000 records per
run. Manual triggers require backend `admin` authorization and an
`Idempotency-Key`. Automatic retries are capped at three attempts with bounded
exponential delay. Operators may create a new retry run for terminal
failed/partial runs.

Running imports use a one-hour worker lease. A job abandoned beyond that lease
is atomically returned to `retry_wait` when attempts remain, or marked `failed`
when its bounded attempt budget is exhausted. The provider is marked degraded,
and the lease-expiry reason remains visible in import history.

Database uniqueness protects configuration identity, run idempotency, external
offer mapping, and observation fingerprints. Each page is transactional. A
crash therefore commits the whole page or none of it. Provider cursors advance
only after complete successful runs.

Normalized records require a real canonical product, integer-minor-unit money,
uppercase currency, allowlisted availability/condition, bounded timestamps, and
HTTPS public-network destinations without user info, fragments, control
characters, or local/private hosts. Malformed records are fingerprinted and
recorded without persisting a raw provider payload.

Only a complete successful snapshot may mark unseen mappings absent and
deactivate their current offers. Partial/failed runs remain visible, degrade the
provider, do not advance its cursor, and never reconcile missing records.

## Freshness

Provider observations store `observed_at`, optional provider time,
`imported_at`, and `expires_at`. A bounded provider expiry is used when supplied;
otherwise the deliberately configured `freshness_ttl_minutes` applies. Legacy
manual offers use `OFFER_MAXIMUM_AGE`.

Public offer reads and redirects require active merchant/offer/link state,
`in_stock` or `backorder`, platform maximum-age freshness, and an unexpired
explicit expiry when present. Price and availability observations are
append-only; database triggers reject updates and deletes.

## Affiliate clicks

The API resolves a fresh safe destination and recommendation ownership before
attempting attribution persistence. If click/analytics storage fails after
resolution, normal navigation still redirects. Invalid, stale, expired, or
forged recommendation-linked destinations fail closed.

`X-Request-ID` is a bounded idempotency key. Obvious prefetch/bot requests remain
in raw history while only human-classified rows emit the filtered
`affiliate_clicked` event and enter CTR/rankings. User agents are SHA-256 hashed.
Source/campaign/medium values are bounded and normalized.

The default retention policy is 397 days
(`AFFILIATE_CLICK_RETENTION=9528h`). The worker then clears account, anonymous,
session, campaign, referrer, recommendation, request, idempotency, and user-agent
fields while retaining non-identifying aggregate dimensions.

## Operator API

All endpoints require server-verified `admin` authorization:

- `GET|POST /api/admin/commerce/providers`
- `PUT /api/admin/commerce/providers/{providerID}/lifecycle`
- `GET|POST /api/admin/commerce/imports`
- `POST /api/admin/commerce/imports/{importID}/retry`
- `GET /api/admin/commerce/imports/{importID}/failures`

The `/admin/commerce` UI exposes lifecycle, freshness policy, last
success/failure, failures, counts, retries, and honest empty/error states.
Provider creation always starts disabled.

## Verified conversion contract

Conversion adapters separately implement
`commerce/ports.ConversionProviderAdapter`. `VerifyWebhook` owns the real
provider signature/timestamp/parser contract and returns structured events only
after authentication. `FetchConversions` owns authenticated cursor-based
imports and complete coverage declarations. The disabled adapter implements
both paths and fails closed; UNSOLERO does not claim support for a network until
its actual sandbox contract is reviewed and tested.

Provider-supplied signed file/artifact ingestion is deliberately not exposed.
Its verifier depends on the selected provider's signature format, transport,
replay identifier, size limits, and key-rotation rules. Add that
adapter-specific boundary and endpoint only after a real contract and sandbox
fixtures exist. Authenticated pull imports and webhooks are the only currently
implemented ingestion shapes.

Webhook bodies are limited to 256 KiB and are never stored. The delivery audit
contains only SHA-256 request/body fingerprints, verification outcome,
signature timestamp, event count, and bounded error code. A processed replay is
acknowledged without reapplying facts. A verified delivery interrupted before
processing safely resumes on retry. Provider-event uniqueness is enforced in
PostgreSQL and conflicting reuse of an event ID is rejected.

The normalized lifecycle supports pending, confirmed, cancelled, reversed, and
rejected orders and pending, approved, reversed, rejected, and paid commission.
Pending is never paid; rejected/reversed amounts are never earned. Order value
and commission are bounded independent integer-minor-unit facts supplied by the
provider. Currencies are uppercased and reported separately. No exchange rate
or currency exponent is assumed; the operator UI labels raw minor units.

Click attribution is optional. It requires a countable click for the same
merchant, preceding the conversion, within a maximum thirty-day window. A
provider-verified event with insufficient attribution evidence remains verified
and unattributed. A click alone can never create a conversion.

Complete successful imports may be reconciled. Reconciliation records matched,
missing, conflicting, stale, and unresolved outcomes as immutable audit items;
it never overwrites provider-event history. Monetization metrics require a
successful reconciliation that covers the entire requested window:

- affiliate conversion rate: confirmed attributed conversions / eligible countable clicks;
- earnings per click: approved or paid commission / eligible countable clicks;
- revenue per visitor: approved or paid commission / observed page-view sessions;
- revenue per recommendation: approved or paid commission / recommendations with eligible clicks;
- commission: approved or paid commission by original currency;
- reversal rate: currently reversed commission / conversions previously approved or paid;
- repeat user rate: authenticated users observed on at least two dates / observed authenticated users.

Every response includes its definition and time window. Missing reconciliation
coverage is `no_data`; a known zero denominator is `insufficient_data`; a known
zero numerator with a non-zero denominator is a real zero.

Additional protected endpoints are:

- `PUT /api/admin/commerce/providers/{providerID}/conversions`
- `GET /api/admin/commerce/conversions`
- `GET|POST /api/admin/commerce/conversion-imports`
- `POST /api/admin/commerce/conversion-imports/{importID}/retry`
- `GET /api/admin/commerce/reconciliations`
- `POST /api/admin/commerce/conversion-imports/{importID}/reconcile`
- `GET /api/admin/commerce/metrics`

## External requirements and limitations

Activation requires a reviewed provider adapter, secret-manager credentials,
affiliate program approval, separate offer/conversion contract evidence,
provider sandbox tests, and rate-limit, incident, reconciliation, and rotation runbooks. The worker uses PostgreSQL
polling, which is appropriate at current scale but not load-tested. A total
PostgreSQL read outage prevents safe destination resolution; UNSOLERO will not
expose raw affiliate URLs merely to fail open. With every real conversion
provider disabled, no conversions are created and metrics correctly stay **No data**.

## Phase 3 status

| Requirement | Status | Evidence / risk |
| --- | --- | --- |
| Provider-neutral contract and disabled adapter | Complete | Registry tests; live semantics require real adapters |
| Scheduled/manual imports and cursors | Complete | Worker, protected API, cursor tests |
| Idempotency, observations, audit history | Complete | Migration 14 and PostgreSQL tests |
| Bounded retry and observable failures | Complete | Retry and persisted failure tests; no external alert delivery |
| Freshness and safe reconciliation | Complete | Stale/expired and import tests |
| Click filtering, idempotency, privacy | Complete | Domain/application/PostgreSQL tests; classification remains heuristic |
| Navigation after tracking write failure | Complete | Application/auth fail-open tests; resolver still needs database reads |
| Commercial/recommendation isolation | Complete | PostgreSQL invariance and architecture tests |
| Live provider connection | Blocked | Credentials, agreements, and adapters required |
| Verified conversion model and ingestion boundary | Complete | Migration 15, disabled adapter, webhook/import tests; real adapters blocked |
| Event idempotency, lifecycle, attribution, reconciliation | Complete | Database constraints and PostgreSQL integration tests |
| Verified monetization metrics | Complete | Coverage-gated queries distinguish no data, insufficient data, and zero |
| Live verified conversion provider | Blocked | Provider agreement, credentials, contract, sandbox fixtures, and operational owner required |
