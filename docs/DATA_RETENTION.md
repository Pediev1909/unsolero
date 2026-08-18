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
commerce operation.

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

