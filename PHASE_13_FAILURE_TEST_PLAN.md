# UNSOLERO Phase 13 failure-test plan

Status: **planned; do not execute without G00–G13 authorization**  
Prepared: 2026-08-18

This plan applies only to the isolated hosted staging environment, fictional
data, approved resource IDs, and an announced test window. It does not authorize
production/customer/provider disruption. Each fault is bounded, reversible,
observed by telemetry, and stopped if the rollback path is uncertain.

## 1. Common preconditions

Before every test:

1. Record test ID, owner, UTC window, candidate manifest, source SHA, all image
   digests, exact staging resource IDs, expected alert and rollback owner.
2. Confirm the environment contains no production credentials or real customer,
   financial, affiliate, conversion, merchant, email, AI, or media data.
3. Confirm two healthy web/API/worker instances, schema readiness, zero
   unexplained backlog, successful telemetry heartbeat, latest backup state,
   and a known fictional invariant set.
4. Confirm the real staging alert destination is staffed and maintenance/test
   notifications are labeled as exercises.
5. Take a provider backup/PITR checkpoint where the test changes persistent
   state. Do not assume backup success; observe it.
6. Save baseline request/error/latency, pool, PostgreSQL, Valkey, worker, media,
   backup, container and collector metrics.
7. Execute only one primary fault at a time unless the test explicitly studies
   correlated failure and has separate approval.

Global abort conditions:

- unexpected resource, account, project, hostname, provider, or data is touched;
- error rate or duration exceeds the approved bound;
- telemetry/alert visibility is lost;
- secrets or personal data appear;
- an integrity invariant fails outside the declared test fixture;
- the provider reports a wider incident or cost/rate limit;
- rollback cannot start within five minutes.

After abort, remove load, reverse the fault, verify schema/invariants/readiness,
preserve evidence, and mark the test `FAIL` or `BLOCKED`. Never delete a primary
or backup merely to make a rerun clean.

## 2. Evidence bundle per test

- sanitized exact commands/provider operations and exit statuses;
- start, fault-observed, alert-trigger, alert-delivered, acknowledged,
  recovery-start and recovery-complete timestamps;
- pre/during/post dashboard exports and relevant bounded log query;
- request count, status/error ratio and p50/p95/p99;
- database pool waits/cancellations and provider resource metrics;
- durable backlog/retry/dead/lease-recovery changes;
- before/after fictional data invariants and object digests where relevant;
- measured interruption, recovery point, RPO/RTO when applicable;
- owner conclusion, limitation, remediation and rerun link.

## 3. Process and deployment faults

### F01 — API replica loss

**Purpose:** prove public routing and distributed limiter behavior survive one
API process loss.

**Method:** under steady bounded read/auth/recommendation traffic, use the
provider-supported targeted instance restart if available. If App Platform
cannot target a single instance, reduce API instance count from two to one in a
reviewed temporary App Spec revision, wait for health, then restore two. Do not
deploy an image that intentionally corrupts data or bypasses probes.

**Expected:** remaining API serves successful traffic; no split-brain session
or limiter behavior; a restart/capacity alert arrives; pool/HTTP metrics retain
the replica distinction; second instance rejoins healthy.

**PASS:** no unapproved 5xx/auth bypass/data divergence, recovery and alert
timestamps observed, final count two. Provider inability to target or safely
scale is documented, not hidden.

### F02 — web replica loss

Repeat F01 for `web`. Verify static assets, `/api`, sitemap, robots, canonical
headers, genuine 404s and TLS remain correct. Browser caching must not mask the
test; use a fresh session and direct external monitor.

### F03 — worker replica loss and lease recovery

Queue a small labeled fictional import/reconciliation fixture using the existing
idempotency controls. Stop or scale out one worker after a job acquires a lease,
without killing PostgreSQL. Wait only past the configured test lease or use an
approved shortened staging-only lease. Restore the second worker.

**Expected:** no duplicate canonical offer/conversion/media action; durable job
returns to recoverable state; another worker recovers it after lease expiry;
lease recovery/retry/backlog metrics and alert are visible.

**Abort:** any real provider call, financial fact, unbounded retry, duplicate
verified conversion, or media deletion outside the fixture.

### F04 — unhealthy candidate and application rollback

Use a deliberately non-serving **test-only** candidate whose health check fails
before traffic, or a provider-supported deployment failure fixture. Do not add a
backdoor or change security policy. Observe deployment rejection and alert.
Then deploy a valid new candidate and invoke provider rollback to the prior
known digest.

**PASS:** unhealthy revision receives no traffic; migration state remains
compatible; prior digests return; full smoke/metrics/logs pass; rollback time is
measured. App rollback does not claim database rollback.

## 4. PostgreSQL faults

### F05 — connection-pool saturation

Use a dedicated staging load role and a bounded number of transactions that hold
connections with `pg_sleep` or an equivalent controlled query. Cap clients below
the provider connection maximum and reserve incident/migration capacity. Ramp
slowly while representative API traffic runs.

**Expected:** application pool acquired approaches max; wait count/time rises;
caller deadlines and statement timeout cancel work; no unbounded goroutine,
connection, transaction or lock leak; alerts fire before provider exhaustion;
normal traffic recovers after release.

**Never:** remove timeouts, consume reserved admin connections, or run against
any non-staging cluster.

### F06 — statement timeout and cancellation

Run a read-only slow query through a dedicated test endpoint/tool or database
role with a known bounded duration longer than `DATABASE_STATEMENT_TIMEOUT`.
Cancel one caller early as a separate case.

**Expected:** safe error class, cancellation metric, no query continuing past
the bounded provider observation window, connection returned to pool, next
request succeeds. Do not log SQL parameters or payloads.

### F07 — PostgreSQL network outage

Temporarily remove only the App Platform VPC/trusted-source rule or use a
provider-approved connection block. Do not delete, resize or modify data.

**Expected:** liveness remains process-up; readiness closes; database-dependent
requests fail generically and within timeout; workers do not acknowledge lost
work; alert is delivered; service reconnects and schema/invariants pass when the
rule is restored.

### F08 — managed PostgreSQL failover

Only the database owner may invoke the provider's documented standby
promotion/maintenance/failover procedure. If DigitalOcean does not expose a
safe authorized action, this test is `BLOCKED` and must not be approximated by
deleting a node.

**Expected:** provider promotes standby; private endpoint behavior is observed;
bounded transient failures occur; pools reconnect; no committed fictional fact
is lost beyond the measured provider recovery point; readiness and workers
recover; alert delivery and provider event are recorded.

### F09 — migration failure and compatibility

Clone/fork the staging database into an isolated disposable target. Run an
intentionally invalid throwaway migration fixture using the migration runner;
verify transaction rollback and unchanged manifest. Separately prove the real
candidate's migrations are compatible with both old and new application images
where the release plan claims rolling compatibility.

Never add the invalid fixture to the release migration directory or execute it
on primary staging. A migration checksum mismatch must stop deployment.

## 5. Valkey faults

### F10 — Valkey outage

Temporarily remove only the app trusted source or block its staging credential.
Keep PostgreSQL and object storage healthy.

**Expected:** readiness closes because the distributed limiter is critical;
protected routes fail closed with bounded 503 behavior; no fallback to
replica-local rate limiting; `redis_unavailable`/backend-failure and latency
metrics rise; alert is delivered; service recovers when access is restored.

### F11 — managed Valkey failover

Use only a provider-supported promotion/maintenance operation under the
platform owner's authority. If no safe trigger exists, record `BLOCKED`.

**Expected:** TLS/auth endpoint remains or changes according to documented
provider behavior; clients reconnect; limiter is never bypassed; persistence
and key-loss effects are measured; protected traffic remains closed while state
is unavailable; recovery and alert delivery are observed.

### F12 — Valkey latency/degradation

App Platform does not provide privileged traffic control. Do not install a
privileged sidecar or weaken TLS. A Toxiproxy-like component may be used only
for a separate test Valkey endpoint populated with disposable limiter keys and
explicit approval. Managed-path packet loss/latency remains `BLOCKED` if the
provider offers no safe mechanism.

## 6. Object storage and media faults

### F13 — primary media access denial

Use a test copy of the application media key or temporarily revoke only that
key. Do not change bucket ownership/versioning or recovery credentials.

**Expected:** readiness closes; uploads/reads fail generically and within
timeouts; no database image reference is published without its object; worker
deletions remain retryable; storage alert arrives. Restore/rotate the key and
verify known object digest, upload/read/delete fixture and readiness.

### F14 — missing object and orphan reconciliation

Create two labeled fictional fixtures through approved tools: a database
reference whose test object is withheld, and an aged unreferenced object. Run
bounded reconciliation in dry-run mode and verify both discrepancy classes.
Apply mode may enqueue deletion only for the known orphan after its explicit
grace period and separate approval.

**PASS:** no unrelated object is changed; pagination/cursors complete; audit
record, backlog and discrepancy metrics are correct; final inventory is clean.

### F15 — media recovery copy

Record the primary fixture inventory and SHA-256 digests, run the scheduled copy
to the separate recovery bucket, then deny access to the primary. Restore into
a new isolated primary bucket or a recovery namespace; do not overwrite the
original during the proof. Point an isolated app revision at the restored copy,
run integrity/inventory/product-page checks, and measure recovery.

Versioning alone is not PASS. The copied objects and their usability must be
observed.

## 7. Backup, restore and disaster recovery

### F16 — logical backup success and clean restore

Run the hosted encrypted backup job using the read-only backup role. Observe
archive creation, checksum, migration fingerprint, encryption, upload and
heartbeat. Restore into a new empty hosted PostgreSQL target using a temporary
restore credential. Verify migration checksums, critical row/count and immutable
fact invariants, auth/admin denial, deterministic recommendation, commerce
independence, media references and full smoke.

The backup artifact must remain unavailable to API/worker credentials. Record
backup and restore durations separately from target provisioning and cutover.

### F17 — corrupt and populated-target restore rejection

Use disposable copies only. Corrupt one encrypted/archive copy and attempt
restore to an empty disposable target; then attempt a valid restore to a
populated disposable target.

**Expected:** both fail before destructive modification, with the documented
safe class/exit code and alert. Preserve the original verified backup.

### F18 — backup failure and stale backup alerts

Run a dedicated test backup job with an intentionally invalid **test-only**
destination/credential, leaving the real schedule unchanged. Separately withhold
only the test heartbeat beyond a shortened test threshold.

**Expected:** failure and stale-heartbeat rules deliver and resolve. Never edit
the real checkpoint to claim a stale condition or delete a valid backup.

### F19 — PostgreSQL PITR

Create timestamped fictional facts A, B and C with committed timestamps. Choose
a recovery point after A and before B. Invoke managed PITR into a new cluster.
Record the provider-selected point and actual restored facts.

**PASS:** A is present; B/C absence or presence matches the actual recovered
timeline; migrations/invariants/smoke pass; RPO is calculated from the last
expected committed fact vs recovered state; RTO includes restore provisioning,
validation and application cutover readiness. Never replace primary during the
proof.

### F20 — full isolated recovery rehearsal

Assume app configuration and primary data endpoints are unavailable. From the
separately retained release manifest, App Spec/infrastructure definition,
secrets escrow process, database backup/PITR and media recovery copy, create an
isolated recovery environment. Validate DNS without switching the public
staging hostname until the witness approves.

Measure effective database/media RPO and end-to-end RTO. A partial database-only
restore cannot establish whole-system RPO/RTO.

## 8. Telemetry and alert faults

### F21 — real alert delivery

Create a dedicated rule named as a Phase 13 test. Trigger it with a reversible
synthetic metric/log/uptime condition. Observe provider ingestion, rule state,
email/Slack receipt, human acknowledgement and resolution.

**PASS:** trigger-to-delivery and delivery-to-ack times are recorded from the
actual destination. A dashboard screenshot without delivery is FAIL.

### F22 — collector or log forwarding loss

Stop only the metrics collector, then separately disable one test component's
log forwarding. Keep application traffic healthy.

**Expected:** telemetry-silence/dead-man alerts arrive through an independent
monitoring path. Restore collection and verify gap visibility. If the same path
must detect its own total outage and cannot, document the blind spot and add an
independent external heartbeat.

## 9. Load, soak and network behavior

### F23 — bounded saturation/load

Use the existing load probe plus authenticated admin pagination and
representative catalog/recommendation fixtures. Ramp in defined stages; do not
start at saturation. Capture response percentiles/errors and every dependency
metric. Stop before budget/provider limits or sustained error thresholds.

The goal is to find the knee and safe operating envelope, not to publish an
unbounded requests-per-second claim. Recommendation output must remain
deterministic and commercial-data independent throughout.

### F24 — multi-hour soak

Run the approved representative mix for at least four hours after a shorter
successful ramp. Include periodic login, admin pagination, recommendation,
catalog, media, consented fictional analytics, health and genuine 404 checks.
Observe memory, goroutines/process restarts, pool waits, DB storage/WAL, Valkey,
backlog age, log/metric continuity and cost.

Longer duration may be required after the baseline. Four hours is an initial
staging gate, not proof of production capacity.

### F25 — packet loss/latency

Safe provider-native degradation controls may be used only if they target the
isolated staging resource and are reversible. App Platform containers do not
receive privileged network control; do not add it. A dedicated fault proxy may
exercise disabled provider sandboxes or disposable dependencies, but results do
not prove the managed PostgreSQL/Valkey/Spaces path. Unavailable path-specific
fault injection remains `BLOCKED`.

## 10. Security and independent review tests

Internal security verification includes hosted CI security gates, TLS/header/
cookie checks, public surface discovery, authorization/IDOR tests on fictional
accounts, rate-limit failure, signed webhook replay/invalid-signature tests only
for an approved sandbox, dependency/image/SBOM review, and secret-output review.

It does not prove:

- external penetration testing;
- independent manual security architecture/IAM review;
- independent WCAG/manual assistive-technology review;
- independently witnessed disaster recovery;
- privacy/legal compliance or business approval.

Those remain separate `BLOCKED` gates until named third parties deliver dated
evidence. Internal engineers must not sign both sides of an independent gate.

## 11. Cleanup

After the test window:

1. restore replica counts, trusted sources, credentials, schedules and alerts;
2. verify final schema, data/object invariants, readiness and full smoke;
3. remove only explicitly labeled disposable fixtures/targets through approved
   provider deletion and preserve required evidence;
4. rotate temporary test credentials and revoke assessor/test access;
5. confirm no sandbox/live-provider setting changed;
6. record final cost, resource inventory, unresolved failures and next action;
7. keep valid backups and the prior deployable digest until evidence review is
   complete.

No test is PASS merely because the service eventually returned. PASS requires
bounded behavior, data integrity, observability, alert delivery, successful
recovery, and complete evidence.
