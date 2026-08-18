# UNSOLERO analytics architecture

Status: first-party repository implementation  
Policy version: `analytics-v1`

UNSOLERO uses a small typed event pipeline. It does not embed a third-party
analytics SDK. Optional product analytics is separate from merchant attribution,
security auditing, operational logs, and deterministic recommendations.

## Event taxonomy

| Event | Exact properties | Author |
|---|---|---|
| `page_view` | none | Browser |
| `onboarding_started` | `onboarding_id` UUID | Browser |
| `onboarding_completed` | `onboarding_id` UUID; outcome enum | Browser |
| `recommendation_generated` | status enum; persistence enum | Browser |
| `product_viewed` | `product_id` UUID | Browser |
| `product_saved` | `product_id` UUID; persistence enum | Browser |
| `comparison_created` | product count 2–4; persistence enum | Browser |
| `setup_saved` | `setup_id` UUID; persistence enum | Browser |
| `affiliate_clicked` | bounded offer/product/source/campaign subset | Server only, after a countable recorded redirect |

Browser envelopes require a UUID `event_id`, UUID `session_id`, bounded
`surface`, current `consent_version`, exact properties, and optional safe
context. `page_path` excludes query/fragment; attribution values are bounded
tokens; referrer is hostname only. Occurrence time, identity, reportability,
schema version, and retention expiry are server-authored.

## Pipeline

```mermaid
flowchart LR
  UI[Typed browser producer] --> HTTP[Strict JSON transport]
  HTTP --> CLASS[Bot / prefetch classification]
  CLASS --> SCHEMA[Exact schema and privacy allowlist]
  SCHEMA --> CONSENT[Locked server consent check]
  CONSENT --> UNIQUE[PostgreSQL unique public event ID]
  UNIQUE --> EVENT[(Validated event)]
  CLASS --> RECEIPT[(Payload-free receipt)]
  SCHEMA --> RECEIPT
  CONSENT --> RECEIPT
  UNIQUE --> RECEIPT
  EVENT --> REPORT[Filtered aggregate report]
  CLICK[Commerce click facts] --> REPORT
  REPORT --> ADMIN[Analyst/admin dashboard]
```

`received` is a payload-free receipt, not a business fact. Outcomes are
`accepted`, `rejected`, `privacy_filtered`, `bot_filtered`, and `deduplicated`.
Only `analytics.events.is_reportable=true` enters product-event metrics. Click
metrics use `commerce.affiliate_clicks.is_countable=true`. Conversion/revenue
metrics require verified, reconciled Phase 4 data and are never inferred.

The unique `public_event_id` database constraint handles retries and concurrent
duplicates. A duplicate returns the same empty success surface as acceptance;
there is no read-by-ID endpoint. The internal server UUID remains authoritative.

## Consent

- No record means `unknown`; optional events fail closed.
- `PUT /api/analytics/consent` stores `granted` or an initial `denied` decision.
- Declining after a grant records `withdrawn`; it does not rewrite history or
  claim deletion completed.
- Every decision records the policy version, safe source, and server time.
- The browser cache controls presentation only. Ingestion locks and verifies
  current server state and policy version in the same transaction as insertion.
- A new policy version requires a new explicit valid decision. Old-version
  event envelopes are rejected.

## Reporting semantics

`GET /api/admin/analytics` defaults to 30 days; `from`/`to` are RFC3339, the
maximum range is 366 days, and ranking `limit` is 1–50. Reports identify:

- layer: `validated_filtered`;
- coverage: `complete` or `partial` against the v3 coverage checkpoint;
- state: `no_data`, `insufficient_data`, or `available`;
- minimum rate sample: 20 eligible observations;
- receipt outcome counts separately from business metrics.

Rates are `null` below the sample threshold or without a denominator. The UI
renders “No data” or “Insufficient data,” never an invented zero percentage.
Historical v1/v2 event rows are non-reportable because current server consent
cannot be proven. Countable affiliate clicks remain commerce facts.

## Access

| Surface | Allowed role |
|---|---|
| Own current consent | Anonymous browser subject or authenticated account |
| Account exportable events/consent history | Authenticated owner |
| Aggregate validated analytics | Analyst, administrator |
| Event-level validated records | Administrator only |
| Payload-free receipt aggregates | Analyst, administrator |
| Commerce click/conversion reports | Commerce operator, administrator; analyst gets read-only aggregate commerce access where routed |
| Security events | Administrator/security operations only; no analytics route |

Backend permission middleware is authoritative. Frontend navigation hiding is
only a usability aid.

## Provider extension

An external exporter may implement an analytics-owned port later. It must
receive only validated allowlisted records, honor consent/withdrawal/retention,
pass contractual/privacy review, and never become a recommendation input. No
provider or write key is configured now.

