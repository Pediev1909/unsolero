# UNSOLERO production configuration contract

Status: repository-enforced contract; deployment values not provisioned  
Last reviewed: 2026-08-17

Production is selected only with `APP_ENV=production`. Configuration is loaded
before listeners, workers, storage, email, AI, alerting, or commerce adapters
start. Invalid or incomplete configuration terminates the process. Values in
`.env.example` are development examples and are prohibited as production
secrets.

## Required production values

| Variable | Contract | Startup/readiness behavior |
| --- | --- | --- |
| `APP_ENV` | exactly `production` | selects all production invariants |
| `APP_VERSION` | immutable release identifier; not empty or `development` | startup fails |
| `DATABASE_URL` | PostgreSQL URL with `sslmode=require`, `verify-ca`, or `verify-full`; deployment should use `verify-full` | startup fails; readiness fails on connectivity or schema incompatibility |
| `PUBLIC_SITE_URL` | HTTPS origin only, no path, public DNS hostname; localhost, IPs, single-label and reserved/private suffixes rejected | startup fails |
| `SESSION_COOKIE_SECURE` | `true` | startup fails; session remains HttpOnly, Secure, SameSite=Lax |
| `RATE_LIMIT_KEY_SECRET` | stable raw-standard base64 encoding of exactly 32 random bytes | startup fails |
| `MFA_ENCRYPTION_KEY` | stable raw-standard base64 encoding of exactly 32 random bytes | startup fails |
| `EMAIL_PROVIDER` | `external` | startup later fails until a reviewed adapter is linked |
| `MEDIA_STORAGE_PROVIDER` | `external` | startup fails until a reviewed durable adapter is linked |
| `MEDIA_SCAN_PROVIDER` | `external` | startup fails until a reviewed malware-scanning adapter is linked |
| `ALERT_PROVIDER` | `webhook` or a linked `external` adapter; `disabled` is rejected | startup fails; webhook requires HTTPS and a 32+ character bearer token |
| `METRICS_ENABLED`, `METRICS_TOKEN` | `true`; token at least 32 characters | startup fails when missing; export remains authenticated |

Production cannot currently start completely because hosted email, Redis,
media storage, media scanning, alert destination, metrics collector, and their
credentials have not been selected/provisioned. A provider-neutral HTTPS alert
webhook adapter is linked; actual delivery still requires a controlled hosted
test and accountable recipient.

## Conditional provider and secret values

| Variable | Required when | Rule |
| --- | --- | --- |
| `API_REPLICA_COUNT` | always | `1..1000`; values above one require `RATE_LIMIT_PROVIDER=external` |
| `RATE_LIMIT_PROVIDER` | always | `local` or `external`; local is single-process only |
| `AI_PROVIDER` | AI is approved | provider identifier; `disabled` is the safe default |
| `AI_MODEL`, `AI_API_KEY` | `AI_PROVIDER != disabled` | both required; key remains server-only |
| `ALERT_WEBHOOK_URL`, `ALERT_WEBHOOK_TOKEN` | `ALERT_PROVIDER=webhook` | HTTPS URL without embedded credentials/query; token at least 32 characters |
| `ALERT_TIMEOUT` | webhook alerting | `1s..30s` |
| provider credentials/signing keys | a merchant/conversion provider passes activation review | provider secret manager only; there are no generic checked-in credential variables |

Unknown providers and missing adapters fail closed. Enabling a string in
configuration does not activate an integration. Provider activation also
requires persisted approval state and the checklist in
`PROVIDER_ACTIVATION_CHECKLIST.md`.

## Bounded operational settings

Every duration and numeric setting is parsed strictly and bounded. The current
safe defaults are documented in `.env.example`:

- HTTP: `API_PORT`, read-header/read/write/idle/handler/shutdown timeouts and
  `HTTP_MAX_HEADER_BYTES`. Write timeout must exceed handler timeout.
- Database: minimum/maximum pool size, connection lifetime/idle/health periods,
  connect/statement/lock/idle-transaction/migration timeouts.
- Authentication: `SESSION_COOKIE_NAME`, absolute and idle TTLs; idle cannot
  exceed absolute TTL. Verification, reset, MFA challenge and step-up TTLs are
  bounded. `MFA_ENFORCE_PRIVILEGED` defaults true in production.
- Rate limiting: authentication, recommendation, analytics, affiliate and
  mutation per-minute limits are each `1..100000`.
- Commerce: offer freshness, click retention, worker poll/cycle/lease bounds,
  cycle batch size and failure threshold.
- Analytics: anonymous/authenticated/receipt retention and cleanup batch size.
- AI: response timeout and maximum bytes.

`TRUSTED_PROXY_CIDRS` is optional and empty by default. Empty means forwarding
headers are ignored. If set, use only the exact ingress ranges. Prefixes broader
than IPv4 `/8` or IPv6 `/32` are rejected, but that bound is not permission to
trust a broad network.

The browser and API are deployed as one public origin. The API emits no
permissive CORS policy. State-changing cross-origin requests are rejected using
origin/fetch-metadata checks. Any future cross-origin architecture requires a
separate explicit design review.

## Development-only values prohibited in production

- `PUBLIC_SITE_URL` using HTTP, localhost, an IP, or a reserved/private suffix;
- `SESSION_COOKIE_SECURE=false`;
- `DATABASE_URL` with `sslmode=disable` or absent TLS mode;
- `APP_VERSION=development`;
- `ALLOW_INSECURE_LOCAL_STAGING=true` (accepted only by loopback staging);
- `EMAIL_PROVIDER=development|disabled`;
- `MEDIA_STORAGE_PROVIDER=local`;
- `MEDIA_SCAN_PROVIDER=development|disabled`;
- ephemeral/empty rate-limit or MFA keys;
- more than one replica with local rate limiting;
- disabled alert delivery or metrics export;
- `.env.example` database credentials;
- `VITE_DEV_API_PROXY_TARGET`, `POSTGRES_*`, `WEB_PORT`, `SEEDS_DIR`, local
  backup/restore variables, and Compose development service settings.

## Secrets, logging, and rotation

Secret values belong in a deployment secret manager and must never be passed as
frontend `VITE_*` values, baked into images, committed, or written to logs.
Structured logging redacts known secret field names and does not log request
headers/bodies, email addresses, raw IPs, affiliate destinations, tokens, or
database URLs.

- Rotate session-impacting secrets with an explicit forced-session-revocation
  plan.
- Rotating `MFA_ENCRYPTION_KEY` requires envelope-key/key-version support or a
  controlled re-enrollment plan; blind replacement makes existing seeds
  unreadable.
- Rotate rate-limit HMAC keys with coordinated replica deployment; rotation
  resets bucket identity.
- Provider/webhook keys need overlap, key IDs, replay tests, and documented
  rollback.
- Database credentials and metrics tokens follow the platform secret rotation
  policy and must be verified in staging before production.

## Deployment gate

Before promotion: render the deployment configuration, run the production
configuration test matrix, verify no secret appears in manifests/logs, run
migrations as a separate bounded job, confirm readiness reports compatible
schema and every required provider, and record human approval. A successful
local Compose run is not evidence that this production contract is satisfied.
