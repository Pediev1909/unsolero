# Transactional email production boundary

Status: BLOCKED — contract and SMTP adapter are implemented; no real provider is configured or approved  
Last reviewed: 2026-08-17

UNSOLERO sends only account verification, password reset, password-change, MFA-enabled, and recovery-code-regeneration messages. Marketing email is outside this implementation.

## Repository guarantees

- `email.Sender` remains provider-neutral. Development, disabled, SMTP, and future external adapters do not change identity services.
- Verification/reset tokens are random, stored only as hashes, expiring, invalidated on replacement, and one-use.
- Links use the configured public origin and place the one-time token in the URL fragment so it is not sent in the initial HTTP request, proxy logs, or `Referer` header.
- Recipient/from/header values are validated against header injection. Security-event names are allowlisted.
- SMTP requires STARTTLS when configured and uses TLS 1.2 or newer. Production configuration rejects SMTP without TLS and rejects development/disabled delivery.
- Network operations use a bounded timeout. An SMTP message is accepted only after the server completes the transaction.
- Password and MFA state changes are not rolled back because a notification fails. The delivery result is written as a bounded security event without tokens or message content.

## Provider activation requirements

Before `EMAIL_PROVIDER=smtp` or a future provider adapter is enabled outside an isolated test environment:

1. approve a transactional provider, DPA/subprocessor terms, sending domain, markets, and retention;
2. configure SPF, DKIM, DMARC, TLS, least-privilege credentials, rotation, and a secret-manager reference;
3. review templates for enumeration resistance, accessibility, localization, support contacts, expiry copy, and phishing resistance;
4. prove sandbox delivery, expiry, replacement, replay rejection, compromised-account flows, bounce/complaint/suppression handling, and rate limiting;
5. create bounded delivery/failure metrics, dashboards, alerts, incident ownership, provider-status monitoring, and rollback;
6. verify that logs, metrics, traces, provider metadata, and support tooling contain no raw token or password.

## Explicit non-claims

No provider delivery, inbox placement, bounce processing, complaint handling, domain reputation, or production notification has been tested. Mock SMTP unit tests prove the repository protocol contract only. Until the activation requirements are met, production email remains BLOCKED and account flows that require delivery must not be publicly launched.

