# Production-like disaster-recovery exercise

Status: NOT TESTED — local logical restore tooling exists; no production-like environment or approved RPO/RTO exists  
Last reviewed: 2026-08-17

This exercise validates recovery evidence; it is not performed against production. Use fictional staging data and an isolated target account/network.

## Required decisions before the clock starts

- Business owner approves RPO, RTO, service priorities, acceptable data loss, and degraded-mode behavior.
- Platform owner identifies immutable database backups/PITR, private media versions/inventory, KMS/key recovery, configuration, secret references, image digests, and DNS/ingress recovery.
- Incident commander, database owner, storage owner, security observer, application verifier, and independent witness are assigned.
- The restore target is proven isolated. Destructive restore commands require two-person confirmation of exact account, cluster, database, bucket, and timestamp.

## Timed scenario

1. Declare a simulated regional loss or destructive database incident. Record start time and candidate recovery point.
2. Freeze deployments/providers and preserve logs/audit evidence.
3. Provision isolated network, PostgreSQL, Redis, object storage, secrets, and immutable application images.
4. Restore the database to the approved point. Verify backup checksum, migration manifest, roles, constraints, row-count invariants, and no seed data.
5. Restore or reattach private media. Compare inventory, ownership prefixes, object versions, checksums, and deletion-job backlog. Do not make a bucket public.
6. Start worker, then API. Prove schema-compatible readiness before ingress.
7. Exercise authentication/session revocation, recommendation reproducibility, saved user data, media read/write/delete, rate-limit sharing/outage, analytics consent, and provider-disabled behavior.
8. Reconcile merchant/import/click/conversion state without inventing missing financial events. Unverified metrics remain `No data`.
9. Enable isolated ingress, run browser/accessibility/API smoke tests, and record RTO.
10. Test rollback/failback, evidence preservation, stakeholder communication, and post-incident access removal.

## Required evidence

- immutable release/config/migration identifiers;
- backup timestamp, checksum, encryption/key identity, source and target;
- measured RPO and RTO with every pause explained;
- database and media integrity results;
- readiness, smoke, security, authorization, and recommendation-fingerprint results;
- lost/unrecoverable records, orphaned media, unresolved jobs, and reconciliation gaps;
- alert and escalation timestamps;
- named approvals, independent observations, remediation owners, and due dates.

## Pass rule

`PASS` requires the approved RPO/RTO and every integrity/control check to succeed in a production-equivalent environment. A local Docker backup/restore is useful repository evidence but cannot satisfy this gate. Missing media recovery, key recovery, alert delivery, or an independent witness makes the exercise `PARTIAL` or `BLOCKED`.

## Phase 10 local evidence

A disposable PostgreSQL 17 database with all 18 migrations and fictional seed
data produced an atomic custom-format archive, metadata, SHA-256 integrity file,
and migration-manifest fingerprint. Restore into a separate empty PostgreSQL 17
volume verified the checksums and reported `migrations=18`. Stopping PostgreSQL
made API readiness return 503; restarting it restored database/schema checks.
This validates repository tooling and recovery behavior only. The archive is
explicitly marked `encryption=none-local-artifact`; off-site encryption, PITR,
media recovery, retention, RPO/RTO, alerting, access control, and an independent
witness remain NOT TESTED/EXTERNAL.
