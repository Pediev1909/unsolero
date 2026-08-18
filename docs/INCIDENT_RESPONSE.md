# UNSOLERO incident response

Status: provider-neutral runbook; on-call service and external contacts blocked

Phase 11 makes database, readiness, worker backlog/dead jobs, backup age/fail,
media discrepancy, Redis unavailability, sustained 5xx, and latency-budget
alerts expressible through fixed provider-neutral definitions. The disabled
notifier does not claim delivery. Local dependency and replica exercises
validated readiness and recovery signals but did not page a person.

## Priorities

Protect people and credentials, stop ongoing harm, preserve trustworthy
evidence, restore the smallest safe service surface, and communicate verified
facts. Revenue, affiliate relationships, and merchant pressure never override
recommendation integrity or user privacy.

## Severity

| Severity | Examples | Initial response objective |
| --- | --- | --- |
| SEV-1 | credential/MFA compromise, confirmed data exfiltration, recommendation integrity breach, verified financial fact corruption, unrecoverable database outage | acknowledge in 15 minutes; executive/security lead immediately |
| SEV-2 | authentication outage, database unavailable, affiliate open redirect, analytics consent bypass, migration failure, sustained elevated 5xx | acknowledge in 30 minutes |
| SEV-3 | degraded provider, delayed imports, partial analytics coverage, isolated performance regression | business-hours owner within 4 hours |
| SEV-4 | low-risk defect or documentation gap | normal backlog with owner |

Objectives are not guarantees until an on-call provider and rota are funded and
tested.

## Roles

- Incident commander owns severity, timeline, decisions, and handoffs.
- Operations lead controls traffic, deploys, workers, database, and recovery.
- Security/privacy lead owns containment, evidence, notification assessment,
  and credential rotation.
- Product/data integrity lead validates recommendation, evidence, analytics,
  affiliate, and conversion correctness.
- Communications lead publishes approved user/partner updates.

One person may hold multiple roles for a small team, but the incident commander
must remain explicit. Record UTC times, release/image IDs, migration versions,
symptoms, decisions, owners, and validation results. Never paste tokens,
passwords, MFA seeds, raw webhook bodies, or unnecessary personal data into the
incident channel.

## Response sequence

1. Detect and declare: confirm signal, assign severity/commander, start a UTC
   log, and identify the last known-good release and backup.
2. Contain: remove traffic or disable the narrow provider/worker path; preserve
   logs and immutable records; do not make destructive schema changes.
3. Assess scope: affected users, data classes, regions, providers, releases,
   time window, and recommendation/financial integrity.
4. Eradicate: patch the root cause, rotate the affected credential class, and
   invalidate sessions/tokens only to the required scope.
5. Recover: follow deployment or disaster-recovery runbooks, keep readiness
   closed until schema and critical dependencies match, then restore traffic
   gradually.
6. Verify: authentication/MFA, authorization, recommendation reproducibility,
   commerce independence, analytics consent, affiliate redirect safety,
   conversion immutability, backup freshness, and monitoring.
7. Close and learn: user/legal notification decision, evidence retention,
   corrective owners/dates, and a blameless post-incident review.

## Scenario containment

- Authentication compromise: suspend affected accounts, revoke sessions and
  reset/MFA challenges, rotate keys with overlap where possible, and inspect
  append-only security events.
- Recommendation integrity: stop recommendation generation, preserve policy,
  fact, score, and result versions, compare deterministic fingerprints, and
  prove commerce fields were not inputs before reopening.
- Affiliate/conversion breach: disable only the affected provider/link,
  preserve immutable events and delivery fingerprints, reject replay, and do
  not infer revenue from incomplete data.
- Analytics privacy breach: stop client ingestion, preserve consent history,
  bound affected subjects and retention, and involve privacy/legal review.
- Database/schema incident: liveness may stay up but readiness must remain 503;
  use a corrective forward migration or authorized restore, never edit applied
  checksums.

## External requirements

Production still needs an on-call/paging provider, log and metric backend,
security/privacy counsel contacts, merchant escalation contacts, status-page
ownership, evidence retention policy, and at least two trained restore
operators. Until these exist, incident response is PARTIAL.

## Operator playbooks

The automated behavior in this table is a first containment layer. Every
operator action requires a UTC incident record and an explicit owner.

| Scenario | Automated behavior | Operator procedure |
| --- | --- | --- |
| Database outage | readiness returns 503; requests and pool waits are timeout-bounded | stop admission, verify managed-database status, preserve logs, restore service or follow disaster recovery, then validate schema/auth/recommendation/commerce before reopening |
| Media reconciliation discrepancy | dry-run records a bounded class and safe hash; unsafe keys are not deleted | inspect inventory and database ownership, preserve the run, correct registration or enqueue only aged validated orphans after approval |
| Redis limiter outage | protected paths return 503 and readiness degrades | restore the private limiter; do not silently switch to replica-local limiting in a multi-replica incident |
| Compromised application/provider credential | no automatic rotation is claimed | disable the narrow integration, revoke/rotate that credential class through the secret manager, restart affected processes, verify old credentials fail, inspect audit/security events |
| Compromised MFA encryption key | sessions do not reveal the key or seed | remove privileged traffic, rotate to a reviewed key-management design, invalidate/re-enroll affected MFA credentials, revoke privileged sessions, assess notification with security/privacy counsel |
| Affiliate provider outage | provider jobs fail durably and bounded retries apply; existing recommendation rank is unchanged | disable the provider configuration/link if redirects are unsafe, notify commerce owner, reconcile freshness after recovery, never substitute fabricated offers |
| Worker backlog/crash | durable leases, idempotency and expiry preserve work | pause imports if growth threatens the database, inspect oldest lease/failure class, restore worker capacity, recover expired leases once, reconcile provider cursor and duplicates |
| Analytics failure | ingestion returns a generic unavailable response; product operation continues | disable optional collection if repeated, restore database/adapter, verify consent and dedupe invariants, declare coverage gaps in reports |
| Failed deployment/API restart | readiness prevents incompatible traffic; graceful stop drains to its bound | stop rollout, compare image/config/schema identifiers, restore the last schema-compatible image, smoke critical paths, then resume gradually |
| Failed migration | transaction rolls back and the migration is not recorded | keep serving release/readiness policy safe, inspect locks/disk/error, choose a reviewed forward fix or restore; never edit the ledger/checksum manually |
| Restore from backup | corrupt/non-empty targets fail closed | provision an isolated empty target, verify checksum/archive, restore transactionally, validate manifest and invariants, obtain incident approval before traffic cutover |
| Emergency shutdown | no global kill switch is invented | remove ingress, stop workers/provider ingestion, preserve database and logs, revoke only exposed credentials, maintain an evidence/decision timeline |
| Account security incident | tokens/sessions can be selectively revoked; audit events remain append-only | suspend affected identity, revoke sessions/challenges, require credential/MFA recovery, inspect ownership and export scope, make notification decision |
| Suspicious traffic spike | bounded limiter rejects excess traffic; backend failure does not fail open | validate trusted-proxy identity, apply edge/WAF controls, protect auth/recommendation/affiliate paths, preserve aggregate evidence without raw personal data, retune only from measured traffic |
