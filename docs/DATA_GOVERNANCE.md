# UNSOLERO data governance

Status: repository-enforced engineering policy  
Last reviewed: 2026-08-17

This document inventories data handled by the current repository. It describes
technical behavior, not legal compliance, certification, or regulatory
approval. Production purposes, lawful bases, notices, regional requirements,
and retention periods require qualified privacy/legal review.

## Data inventory

| Data class | Collected fields | Purpose | Store | Classification | Required | Access | Retention behavior |
|---|---|---|---|---|---|---|---|
| Validated analytics event | Opaque event/session UUIDs, allowlisted event properties, route path without query, bounded campaign/source/medium, external referrer hostname, optional internal user ID or hashed anonymous subject | Product journey reporting | `analytics.events` | Pseudonymous or identifiable when account-linked | Optional except server-authored merchant-click observation | Account owner in export; aggregate analyst; event-level administrator | Anonymous 90 days; authenticated 397 days; delete |
| Ingestion receipt | Event UUID, bounded event name, outcome/reason, timestamps | Deduplication audit, filter health, pipeline operations | `analytics.ingestion_receipts` | Pseudonymous; no payload or actor | Operational | Aggregate analyst/admin; no public/event lookup API | 30 days; delete |
| Consent state/history | State, policy version, source, decision time, internal user ID or hashed anonymous subject | Enforce and evidence preference changes | `analytics.consent_states`, `analytics.consent_history` | Pseudonymous or identifiable | Required before optional analytics | Subject; account export; administrator by database policy (no general API listing) | Explicit hold pending legal/privacy decision; account deletion removes current state and anonymizes history |
| Analytics identity claim | Hashed browser subject, internal user ID, policy version, claimed/revoked times | One-time authenticated association | `analytics.identity_claims` | Pseudonymous/identifiable | Optional | Analytics service only | Revoked tombstone after account deletion; hold pending review |
| Affiliate click | Click UUID, offer/product/link IDs, session/anonymous/internal user ID, bounded attribution, scheme+host referrer origin, SHA-256 user-agent hash, bot class | Merchant navigation attribution and fraud/filter operations | `commerce.affiliate_clicks` | Pseudonymous or identifiable | Disclosed consequence of merchant navigation | Commerce operator/admin; verified aggregate analysts | 397 days default; identifying fields anonymized |
| Verified conversion | Provider event/order references, click association, status, currency/value/commission lifecycle | Verified commerce reconciliation and metrics | `commerce.affiliate_conversions`, conversion event/audit tables | Commercial and potentially pseudonymous | Only when verified provider data exists | Commerce operator/admin | Explicit hold pending partner, finance, legal, and privacy review |
| Authentication/security event | Internal user/session IDs, event category/outcome, safe bounded metadata, request ID | Account security investigation | `identity.security_events` | Security-sensitive | Essential | Administrator/security operations only; not analytics | Immutable hold pending security/legal review |
| Administrative audit | Internal actor/entity IDs, bounded action/details/timestamps | Governance and operator accountability | Evidence, recommendation-policy, commerce operation/conversion audit tables | Internal confidential | Essential for privileged mutation | Relevant operator/admin roles | Explicit hold pending security/legal review |
| Account data | Email, password hash, profile, wishlist, setups, recommendations | Product account and saved experiences | Identity/planning/recommendation schemas | Identifiable | Account feature | Account owner and authorized services | Phase 5 deletion anonymizes/removes by class |
| Recommendation free text | Optional user-authored description and structured room/preferences | Generate deterministic setup | Recommendation schema | Potentially sensitive free text | Optional | Account owner and recommendation service; excluded from analytics | Cleared on account deletion; never copied to analytics |
| Request observability | Request ID, HTTP method, route path without query, status, bytes, duration | Reliability/security operations | Structured process logs/external sink when configured | Operational | Essential | Deployment operators | External sink policy required; application never logs request body, query, cookies, email, token, destination URL, or free text |

Seed data is fictional catalog/evidence data. It creates no users, customer
activity, analytics consent, clicks, conversions, revenue, or engagement.

## Prohibited general analytics data

The event service uses an exact event/property allowlist. Unknown keys reject
the event as `privacy_filtered`; there is no arbitrary metadata escape hatch.
General analytics cannot accept passwords/hashes, reset or verification tokens,
session/access/API tokens, authorization headers, MFA material, recovery codes,
email, exact IP, room description, free text, payment data, order details, raw
user-agent strings, affiliate destinations, tracking secrets, or commission.

## Ownership and lifecycle

The browser receives a random 256-bit HttpOnly `SameSite=Lax` subject token only
when it saves a preference. PostgreSQL stores SHA-256 of that random token, not
the token. The token is not an authentication credential.

An authenticated account may explicitly claim the current browser subject only
when both subject and account have a current grant for the same policy version.
The claim endpoint accepts no subject/user/event ID from the body. It is
idempotent for the same account, rejects cross-account/revoked claims, exposes a
generic conflict, and moves only eligible event links. No event enumeration API
exists.

Account deletion atomically removes account consent state, changes consent
history to an unlinked anonymized subject, revokes the claim tombstone, and
clears analytics/click user links. Validated anonymous aggregate facts may
remain until their existing expiry. Security/audit records remain under their
separate holds. The product must not claim that every historical aggregate was
deleted.

## Boundaries

- Recommendation production packages are architecture-tested against analytics,
  commerce, affiliate, conversion, revenue, and AI dependencies.
- Analytics never receives raw affiliate destinations, credentials, provider
  secrets, order payloads, or commission. Verified monetization reporting stays
  in commerce.
- Security events are not copied into analytics. Aggregate analytics permission
  does not grant event-level or security-event access.
- AI has no product or analytics repository and cannot publish canonical facts.

## External decisions still required

- Qualified review of purposes, consent wording/versioning, jurisdictions,
  child/minor handling, data-subject procedure, controller/processor roles, and
  production retention.
- Contracts and data-processing review before any external analytics/logging
  exporter is enabled.
- Named owners for privacy requests, retention jobs, access reviews, incidents,
  and legal holds.
