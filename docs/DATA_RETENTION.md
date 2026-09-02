# UNSOLERO data retention

Status: configurable engineering defaults requiring legal/privacy approval  
Last reviewed: 2026-08-17

The database registry `analytics.retention_policies` makes every current class
explicit. A `hold` is an honest unresolved policy—not permission to retain data
forever and not a claim of legal necessity.

| Data class | Default | Action | Configuration / owner |
|---|---:|---|---|
| Payload-free analytics ingestion receipt | 30 days | Delete | `ANALYTICS_RECEIPT_RETENTION` |
| Validated anonymous analytics event | 90 days | Delete | `ANALYTICS_ANONYMOUS_RETENTION` |
| Validated authenticated analytics event | 397 days | Delete | `ANALYTICS_AUTHENTICATED_RETENTION` |
| Affiliate click attribution | 397 days | Anonymize identifying/attribution fields | `AFFILIATE_CLICK_RETENTION`, commerce module |
| Verified conversions | Hold | No automatic deletion until partner/finance/legal review | Commerce owner |
| Security events | Hold, append-only | No automatic deletion until security/legal review | Security owner |
| Consent history | Hold | Account deletion immediately unlinks identity; duration unresolved | Privacy owner |
| Administrative audit | Hold | No automatic deletion until security/legal review | Module owner |
| Newsletter subscription, pending | 48 h (confirmation link lifetime) | Delete the row once the link expires, every worker cycle | Audience module, `PurgeExpiredPending` |
| Newsletter subscription, confirmed or unsubscribed | Hold | Kept while subscribed; unsubscribed rows kept as suppression records | Privacy owner |

Allowed configuration bounds are 1–730 days for anonymous events, 1–1,095 days
for authenticated events, and 1–180 days for receipts. The checked-in values are
not legal conclusions.

## Cleanup operation

The existing worker invokes analytics cleanup every
`COMMERCE_WORKER_POLL_INTERVAL`. Each pass deletes at most
`ANALYTICS_CLEANUP_BATCH_SIZE` events and the same number of receipts (1–10,000)
using indexed `retention_expires_at`, ordered selection, and `FOR UPDATE SKIP
LOCKED`. It is safe to retry and reports deleted counts in structured logs only
when work occurred. Affiliate click anonymization remains a separate bounded
commerce operation. The same cycle purges expired pending newsletter sign-ups
(see "Newsletter subscriptions" below).

Cleanup deliberately does not delete security events, conversion records,
consent history, or admin audit records while their policies are unresolved.
This prevents an arbitrary application default from destroying evidence that
may be required, but the holds must be resolved before production acceptance.

## Account deletion interaction

Deletion does not wait for the worker. In the account-deletion transaction:

1. current account consent is removed;
2. consent history becomes unlinked through a random anonymized subject ID;
3. browser identity claims become revoked tombstones with no user link;
4. analytics events and affiliate clicks lose their user link;
5. saved/profile/account identifiers are removed or anonymized under Phase 5;
6. immutable security audit remains separate.

An event already anonymized for aggregate reporting keeps its original expiry.
The deleted account cannot authenticate, has no consent state, and therefore
cannot submit new account-linked optional events. The same browser may continue
as an anonymous consenting subject, but a revoked subject cannot be claimed by
another account.

## Newsletter subscriptions

`audience.newsletter_subscriptions` holds one row per address: the lower-cased
address, status (`pending`, `confirmed`, `unsubscribed`), the surface that
asked (`footer`, `article:<slug>`), the consent text version shown at the time,
and request/confirmation/unsubscribe timestamps. Token material is stored only
as SHA-256 hashes, and the confirmation hash is cleared the moment it is used.
No IP address or user agent is recorded: consent is proven by the confirmation
click itself, and an address that never confirms is not evidence of anything
worth keeping.

- A pending address that never confirms has given no consent, so nothing about
  it is kept once its 48-hour link is dead. `PurgeExpiredPending` on the
  newsletter service deletes those rows outright, and the commerce worker calls
  it on every cycle next to `CleanupExpiredSecurityArtifacts` — see
  `workCycle` in `backend/cmd/worker/main.go`. The pass is unbounded because
  the predicate is narrow (`status = 'pending' AND confirm_expires_at < now`)
  and self-limiting: each cycle removes only what expired since the last one.
  It logs a count only when it deleted something.
- Unsubscribing does not delete the row, and that is deliberate rather than
  neglect. The row is kept with status `unsubscribed` and the timestamp as a
  suppression record: it is what makes a later sign-up, import or re-added list
  fail to bring back an address whose owner opted out. Deleting the row would
  destroy the only evidence that the person said no. The confirmation hash is
  cleared at the same time, so the confirmation link is dead; the unsubscribe
  hash is kept, because the server matches on it whatever the row's status and
  a second click on the same link must succeed rather than accuse the reader of
  a bad link. Whether and when suppression rows are hashed or deleted is a Hold
  pending privacy review.
- A subscriber who wants the address removed entirely, rather than suppressed,
  is handled by deleting the row manually; there is no self-service erasure
  route yet.

## Operating requirements

- Alert on repeated worker failure, growing expired-row backlog, unusual reject
  rates, or a stale reporting checkpoint.
- Record and approve retention changes; test query plans after meaningful data
  growth.
- Backups and replicas need matching expiry/deletion procedures and documented
  restore behavior for deleted data.
- A legal hold process must override scheduled deletion deliberately, be scoped,
  audited, and reversible; no such production process is configured here.
- External logs/exporters need separately approved retention and deletion APIs.

