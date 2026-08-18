# UNSOLERO provider activation checklist

Status: no live provider approved or active  
Last reviewed: 2026-08-17

This applies independently to merchant feeds, affiliate destinations,
conversion ingestion, email, media/object storage, malware scanning, AI,
metrics, and alert delivery. Checking in an adapter or setting a provider name
does not authorize financial or production activity.

## Phase 10 provider state

| Provider class | Repository adapter | Live state | Required external evidence |
| --- | --- | --- | --- |
| Redis-compatible limiter | implemented and integration-tested with isolated Redis | disabled by default | managed TLS/auth endpoint, private network, failover/eviction policy, capacity and outage exercise |
| S3-compatible media | implemented and integration-tested with isolated MinIO | local development remains default | private managed bucket, TLS/KMS/IAM, versioning/lifecycle/inventory, scanner, restoration exercise |
| Transactional email | SMTP contract and mock-server tests implemented | development/disabled by default | approved provider/domain/credentials, SPF/DKIM/DMARC, bounce/complaint handling, sandbox and alert evidence |
| Merchant/affiliate/conversion | provider-neutral lifecycle remains implemented | every real adapter disabled | agreement, credentials, fixtures, signatures, reconciliation and two-person activation |
| Alerts/telemetry | bounded OpenMetrics and notifier boundary implemented | no collector or alert destination | private collector, retention, dashboards, tested paging and on-call ownership |
| AI | provider boundary remains disabled | disabled | separate future security/product approval; never a product-fact authority |

Repository support is not activation. Configuration must continue to fail closed
when any required credential, TLS control, scanner, adapter, or approval is
absent.

## Evidence required before the first real commerce provider

- [ ] Business owner has an executed merchant/affiliate agreement and records
      allowed markets, traffic sources, disclosure terms, deep-link rules,
      attribution window, commission/reversal semantics, and termination path.
- [ ] Legal/privacy approves disclosure, tracking, processor/subprocessor,
      retention, cross-border, data-subject, and financial-record obligations.
- [ ] Security reviews provider authentication, secret storage/rotation,
      webhook signatures, IP assumptions, replay window, TLS, payload limits,
      redirect destinations, incident notification, and least privilege.
- [ ] Provider-specific adapter maps external identifiers and enums into the
      neutral domain without guessing or fabricating values.
- [ ] Missing/invalid credentials, unknown provider, malformed/oversized
      responses, timeouts, partial pages, stale cursors and throttling fail
      closed and create an auditable failed run.
- [ ] Import idempotency, bounded retries, reconciliation, cancellation,
      reversal, currency and duplicate-event behavior pass against provider
      sandbox fixtures whose origin is recorded.
- [ ] Destination allowlisting and HTTPS validation prevent SSRF/open redirect;
      stale, inactive, missing, unowned and unpublished offers cannot redirect.
- [ ] Bot/prefetch and duplicate click handling is validated under the real edge
      proxy configuration without breaking normal navigation.
- [ ] Conversion signatures and replay protection are validated with provider
      test events. Only verified immutable evidence can produce financial
      metrics. No event means “No data.”
- [ ] A regression test mutates commission, sponsorship, payout, merchant state,
      clicks, conversions and revenue data and proves identical recommendation
      eligibility, score, rank, alternatives and rejection output.
- [ ] Monitoring dashboards and tested alerts exist for import age/failure,
      stale offers, signature/replay failures, reconciliation backlog, provider
      degradation, redirect failures and conversion reversals.
- [ ] Backfill, disable, credential compromise, data correction, rollback and
      provider termination runbooks have named owners and were rehearsed in
      staging.
- [ ] Capacity/rate limits and provider terms were validated with representative
      volume; no unapproved production traffic or financial data was used.
- [ ] Two-person technical and business approval records the adapter version,
      credentials/key IDs, enabled programs/markets, rollout percentage,
      monitoring links, rollback trigger and expiration/review date.

## Activation sequence

1. Keep provider persisted as `disabled`; validate configuration and adapter in
   an isolated sandbox.
2. Import non-financial sandbox fixtures and reconcile without publishing.
3. Approve a bounded staging activation with synthetic/test provider data only.
4. Confirm stale-offer, invalid-signature, replay, outage, malformed-response,
   and rollback tests.
5. Obtain legal/business/security/operations approvals.
6. Activate the minimum production scope through an audited operator action;
   monitor and reconcile before expansion.

Any absent evidence, expired approval, secret failure, provider degradation, or
reconciliation uncertainty stops activation or disables the provider. There is
no emergency exception allowing commercial data into recommendation scoring.
