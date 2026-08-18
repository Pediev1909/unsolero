# UNSOLERO security validation

Status: PARTIAL — Phase 10 repository/integration controls pass; deployment review and independent assessment remain blocked

## Phase 10 findings fixed

1. The single-process limiter could not safely support replicas. An atomic
   Redis-compatible fixed-window adapter now uses namespaced, HMAC-pseudonymous
   keys, server TTLs, strict input bounds, separate route policies, and
   fail-closed outage behavior. Production requires authenticated TLS Redis and
   exact trusted-proxy ranges.
2. Product media had no production-capable storage adapter. The new private
   S3-compatible adapter binds deterministic object keys to product ownership,
   conditionally creates content-addressed objects, validates size/type/digest
   on write and read, and fails closed. Deletion failures enter a durable,
   bounded-retry worker queue. A reviewed malware scanner and managed bucket
   configuration remain BLOCKED.
3. Account collection endpoints were unbounded. Wishlist and saved-setup lists
   now enforce page/page-size bounds, stable indexed ordering, ownership
   predicates, and explicit page metadata. The frontend consumes pages without
   silently truncating user data.
4. The first database integration pass exposed an invalid timestamp/interval
   expression in stale media-deletion lease recovery. The query now applies an
   explicit `timestamptz` cast; the targeted and complete serialized PostgreSQL
   integration suites pass.
5. Compose dependencies were mutable tags. PostgreSQL, Redis, and MinIO service
   references are now tag-plus-digest pinned, and CI uses pinned isolated Redis
   and MinIO dependencies for adapter integration tests.
6. The first real Redis outage/restart drill exposed a container-permission
   defect: the official entrypoint could not traverse Redis's persisted 0700 AOF
   directory after capabilities were dropped. Compose now grants only the
   entrypoint handoff capabilities (`CHOWN`, `DAC_OVERRIDE`, `SETGID`, `SETUID`);
   it still drops every other capability and uses `no-new-privileges`. A rerun
   proved readiness/auth requests changed to 503 during outage and recovered
   after the persisted service restarted.

## Phase 10 adversarial scope

Repository tests exercise concurrent distributed-limit decisions, TTL expiry,
route separation, backend outage, object duplicate races, cross-product access,
path traversal, executable/invalid/oversized uploads, object-store outage,
managed-image ownership, deletion retry, SMTP header injection and TLS policy,
token replay/expiry, forwarded-address trust, analytics consent, affiliate
ownership/staleness, provider signature/replay, and recommendation independence
from commercial data.

The review found no user-controlled SQL identifier concatenation: the remaining
dynamic identity-token table names are internal constants chosen by dedicated
repository methods. Affiliate redirects use validated, active, fresh,
product-owned destinations loaded server-side; the client cannot supply the
destination. API unknown routes return 404. The client-rendered public SPA still
returns an HTML shell with HTTP 200 for unknown browser paths; this is a known
SEO/routing limitation, not a valid server-side 404 implementation.

## Phase 9 findings fixed

1. Production origin validation accepted reserved/private hostnames. Production
   now requires a public-DNS HTTPS origin and rejects IP, localhost, single-label,
   and reserved/private suffixes.
2. Local media names did not bind content to a product and scanning was not an
   explicit trust boundary. Uploads now pass through application-owned storage
   and scanner ports, fail closed without a scanner, use deterministic
   product-scoped digest keys, validate magic bytes, and write atomically.
   Production refuses local storage or development/disabled scanning.
3. Metrics accepted arbitrary counter names. The in-process recorder now has a
   compile-time allowlist, preventing attacker-controlled metric cardinality.
4. Recommendation input responses serialized absent list fields as `null`, which
   violated the frontend schema. The transport now emits stable empty arrays and
   has a regression test.
5. Backup schema fingerprinting initially referenced a nonexistent migration
   column and a shell pipeline masked that database failure. The drill exposed
   both defects; the scripts now hash the ordered version/name/checksum manifest
   and capture database output before hashing so failures remain nonzero.

## Phase 8 findings fixed

1. Trusted-proxy bypass: rate limiting accepted `X-Forwarded-For` from any
   private/loopback peer. A directly exposed container could spoof client IPs.
   Forwarding headers are now ignored by default and trusted only when the
   socket peer belongs to an explicit `TRUSTED_PROXY_CIDRS` prefix. Overbroad
   IPv4 prefixes below /8 and IPv6 below /32 are rejected. Production must use
   only its exact edge proxy ranges.
2. Schema compatibility: readiness previously proved only database reachability.
   The API now embeds the immutable migration manifest and fails readiness for
   missing, extra, renamed, or checksum-mismatched migrations.
3. Indexed identity lookup: authentication and recovery queries now use the
   indexed `lower(email)` expression. This is both a load-resilience and abuse
   resistance improvement.
4. Browser cancellation: caller-aborted frontend requests now remain aborts
   instead of becoming misleading network errors; 403/429/500 and retry bounds
   have explicit tests.
5. New-window links consistently set `noopener`/`noreferrer` or the stronger
   sponsored affiliate relation.

## Validated controls

- Opaque HttpOnly SameSite sessions; Secure required outside development;
  absolute/idle expiration and revocation; password/reset changes revoke
  sessions.
- Argon2id password hashing; enumeration-resistant registration/reset; hashed
  verification, reset, recovery, session, and challenge tokens.
- AES-256-GCM MFA seed storage; one-use recovery/challenge semantics; bounded
  attempts; staff/admin permission checks and optional privileged MFA policy.
- Strict JSON decoding, unknown-field rejection, body/header limits, validated
  IDs/enums/URLs, parameterized SQL, generic error responses, same-origin
  mutation defense, and rate limits that fail closed when their backend fails.
- Affiliate HTTPS destination validation, no raw destination exposure, fresh
  offer checks, ownership checks, idempotent clicks, bot/prefetch filtering,
  signed/replay-resistant conversion ingestion, and immutable verified facts.
- Server-authoritative analytics consent, opaque subjects, allowlisted schemas,
  identity-claim conflict handling, reportability classification, and indexed
  retention.
- Product-scoped deterministic digest image keys, magic-byte/type validation,
  atomic owner-only creation, strict path validation, read-only containers, and
  scanner/storage outage failure tests.
- Deterministic recommendation types have no affiliate/commission/revenue/click
  fields; database tests mutate offer price/availability, merchant status,
  affiliate priority/commission, provider data, clicks, attribution, and
  conversions without changing output.

Targeted concurrent integration tests cover duplicate registration, one-use
password reset and MFA recovery, session ownership, analytics deduplication and
identity claims, affiliate click idempotency, webhook/conversion replay,
conversion reversal, worker leases, policy activation, and failed migration
rollback. Go race testing is part of the final gate.

## Dependency and source scans

| Check | Result |
| --- | --- |
| `govulncheck ./...` | 0 reachable symbol or imported-package vulnerabilities; GO-2026-5932 exists only in unimported `x/crypto/openpgp` and has no fixed version |
| `npm audit --audit-level=high` | 0 vulnerabilities |
| direct frontend license inventory | every direct dependency declared MIT, ISC, or Apache-2.0 |
| repository secret-pattern scan | no private-key, common cloud key, GitHub token, Google key, or OpenAI-key signature found outside ignored/generated paths |
| parameterized SQL/static sink review | dynamic SQL is limited to internal constant table/select fragments; no user-controlled identifier concatenation found |

Checked-in security workflows configure Gitleaks, Semgrep, Trivy filesystem and
image scans, a Docker base-image digest gate, `govulncheck`, npm audit policy,
and an SPDX SBOM. Those remote jobs have **not executed**. Third-party actions
are pinned to immutable tag commits, but those commits and workflows still
require independent review. Local availability remains insufficient for a full SAST/image
scan. These controls are NOT TESTED/BLOCKED—not passed.

## Container validation

The isolated API and worker ran as `app`, with read-only roots,
`no-new-privileges`, all Linux capabilities dropped, and writable `/tmp` only;
API uploads use a dedicated volume. They were not privileged. The Compose web
service is intentionally a root Vite development container and the official
PostgreSQL development service retains its image defaults. The production web
Dockerfile runs as `nginx` on an unprivileged port, but it was not assessed by
an image scanner. No Compose service currently has CPU or memory limits.

Production requires private database/API networks, an explicit ingress proxy
CIDR, TLS termination, managed secrets, egress policy, resource requests and
limits, signed immutable image promotion, image/SBOM scanning, and runtime
monitoring. Development Compose is not a production orchestrator.

## Remaining assessment

Commission an independent penetration test covering authentication/MFA,
authorization, CSRF/origin policy, request smuggling at the real ingress,
SSRF/redirect paths, webhook signatures/replay, object ownership, file upload,
abuse economics, and privacy deletion/export. Run DAST only against an approved
staging environment with seeded fictional data and notification to operators.
