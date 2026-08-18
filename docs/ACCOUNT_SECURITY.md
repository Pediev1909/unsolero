# Account security

This document describes the Phase 5 repository security model. It is an engineering contract, not a certification.

## Authentication and passwords

UNSOLERO uses opaque server sessions in an `HttpOnly`, `SameSite=Lax` cookie. Production requires `Secure`; state-changing cookie-authenticated requests pass the same-origin/Fetch-Metadata middleware. The browser never receives a reusable bearer token in JavaScript.

Passwords are hashed with Argon2id (`64 MiB`, three iterations, parallelism two, 32-byte key by default) and a random 16-byte salt. Verification is constant-time after deriving the candidate key. Login performs a dummy Argon2id verification for unknown accounts and upgrades old parameters after a successful password check. Passwords are limited to 128 bytes and must contain at least 12 Unicode characters.

Registration, password-reset requests, and verification resend requests use generic responses. Registration does not establish a browser session. There is no account lockout that another person can trigger; bounded per-client throttling applies to authentication and security operations instead.

## Email delivery and one-time links

`identity/ports.EmailDelivery` separates delivery from token lifecycle. Verification and password-reset tokens contain 256 random bits, are stored only as SHA-256 hashes, expire, invalidate earlier active tokens, and are consumed transactionally once.

`EMAIL_PROVIDER=development` records delivery intent in process memory and sets `accepted=false`; it does not claim mail was sent. Developers can inspect `GET /api/dev/email-deliveries?recipient=...` only while `APP_ENV=development` and the development adapter is selected. Verification and reset links should put the returned token in the URL fragment:

- `http://localhost:5173/verify-email#TOKEN`
- `http://localhost:5173/reset-password#TOKEN`

Fragments are not sent in ordinary HTTP request targets. Never expose the development API on a shared or public environment. `disabled` records no delivery. Production rejects both `development` and `disabled`. `external` is the configuration contract for a reviewed provider adapter; no live adapter is linked yet, so the API fails closed until one is implemented.

## Sessions and credential changes

Session tokens contain 256 random bits; PostgreSQL stores only their SHA-256 hashes. Sessions enforce a 30-day default absolute lifetime and a seven-day sliding idle lifetime. Stable UUIDs, creation time, last-used time, expiry, and authentication method are safe for account display. Raw tokens and hashes are never returned.

Password changes require the current password, keep the current session so the response is not lost, and revoke every other session. Password reset revokes every session and does not silently create a new one. Users can revoke one owned session or all other sessions. Repository predicates always derive the owner from the authenticated principal; client user IDs are not accepted.

The worker deletes expired/old-revoked sessions and expired or consumed temporary credentials. Immutable audit events are retained even after temporary credentials and sessions are removed.

## MFA and privileged step-up

TOTP follows RFC 6238 with a 30-second period, six digits, and one-step clock drift. Secrets are encrypted with AES-256-GCM using `MFA_ENCRYPTION_KEY`; the key is injected from the server secret manager and is never stored in PostgreSQL. Production requires a stable 32-byte raw-standard-base64 key.

Recovery codes are random, displayed only after enrollment or regeneration, stored only as SHA-256 hashes, and consumed atomically once. Regeneration deletes every earlier code. MFA login uses a five-minute, five-attempt challenge whose 256-bit token is stored hashed and held in a `SameSite=Strict`, `HttpOnly` cookie scoped only to `/api/auth/mfa/complete`.

When `MFA_ENFORCE_PRIVILEGED=true` (the production default), every privileged permission gate requires a verified email and backend-recorded MFA authentication no older than `MFA_STEP_UP_TTL` (15 minutes by default). A frontend state claim cannot satisfy either. Privileged users can access the ordinary account security page to verify/enroll but cannot access privileged APIs until both controls are satisfied.

## Roles and permissions

PostgreSQL role membership is resolved on every session-authenticated request. `admin` is the administrator role and has every capability. Narrow roles are:

| Role | Read | Create/update/delete | Approve/activate/export |
| --- | --- | --- | --- |
| `catalog_editor` | catalog, evidence provenance | catalog drafts, images, attributes | cannot publish governed evidence or policy |
| `evidence_editor` | catalog and evidence | sources, observations, draft revisions, submit | cannot approve/publish |
| `evidence_reviewer` | catalog and evidence | review decisions | approve/publish; repository separation-of-duties still applies |
| `policy_editor` | policy | author/submit policy | cannot approve/activate |
| `policy_reviewer` | policy | review decisions | approve/activate; author/reviewer/activator separation remains enforced |
| `commerce_operator` | merchants, offers, imports, conversions | commerce configuration and operations | provider activation; never recommendation policy |
| `content_editor` | content | content capability reserved for governed content endpoints | no commerce, analytics, evidence, or policy grants |
| `analyst` | analytics, recommendation diagnostics, catalog, commerce reports | none | aggregate analytics export capability only |
| `admin` | all protected data | all administrative operations | all capabilities, still subject to MFA and domain workflow rules |

The catalog status endpoint remains administrator-only so a catalog editor cannot bypass evidence publication. Existing evidence and recommendation-policy repositories continue to reject self-review and prohibited self-activation even when a caller has the route permission.

## Account export and deletion

Export derives ownership exclusively from the authenticated principal and contains appropriate account/profile data, wishlists, setups, recommendation sessions, consent events, and limited security metadata. It excludes password hashes, tokens, MFA ciphertext, recovery-code hashes, provider credentials, internal audit metadata, and other users' records.

Deletion requires the current password and the literal confirmation `DELETE`; privileged deletion also requires recent MFA. It removes profiles, drafts, comparisons, wishlists, setups, role grants, temporary credentials, and MFA credentials. It detaches the user from analytics, affiliate-click attribution, and retained recommendation sessions, then replaces the email with an irreversible `deleted+UUID@users.invalid` address, removes the password, and revokes sessions. Recommendation history, verified commerce/financial records, administrator audit records, and immutable security events remain for operational integrity without the former email address. Retention periods still require legal review before launch.

## Security audit and logging

`identity.security_events` is append-only through a database trigger. It records bounded event type, outcome, request ID, surface, account/session references where appropriate, time, and allowlisted metadata such as a permission key. It never records passwords, raw tokens, MFA secrets, recovery codes, authorization headers, email addresses, IP addresses, or user-agent fingerprints.

HTTP logs contain request ID, method, query-free path, status, size, and duration. Request bodies, headers, cookies, query values, and fragments are not logged.
