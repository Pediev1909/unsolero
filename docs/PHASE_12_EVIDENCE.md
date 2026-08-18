# Phase 12 hosted-staging evidence

Date: 2026-08-18
Verdict: **PARTIAL — repository controls remain verified; hosted and external launch gates are blocked.**
Scope: external-access audit plus locally reproducible repository validation

Phase 12 does not treat the Phase 11 disposable Docker topology as hosted
staging. No cloud control plane, container registry, secret manager, alert
destination, provider sandbox, real customer record, or financial event was
available or activated. Missing external evidence is recorded as `BLOCKED`,
not inferred from interfaces, mocks, workflow files, or local tests.

## Hosted infrastructure and deployment

| Gate | Status | Required owner | Evidence date | Evidence | Remaining action |
| --- | --- | --- | --- | --- | --- |
| Isolated hosted staging | BLOCKED | Platform owner — unassigned | 2026-08-18 | no authorized cloud account/control plane is connected | select provider, owner, region and account; provision through reviewed IaC |
| Managed PostgreSQL | BLOCKED | Database owner — unassigned | 2026-08-18 | Phase 11 used disposable PostgreSQL 17 only | provision private TLS service, separate roles, HA, monitoring, backups and PITR |
| Managed Redis | BLOCKED | Platform owner — unassigned | 2026-08-18 | atomic authenticated local Redis behavior passed in Phase 11 | provision private TLS/auth service; define eviction, persistence and failover |
| Private object storage | BLOCKED | Platform/security owners — unassigned | 2026-08-18 | private MinIO and reconciliation passed locally | provision private bucket, IAM, KMS, versioning, inventory, lifecycle and recovery |
| TLS ingress | BLOCKED | Platform/security owners — unassigned | 2026-08-18 | local self-signed TLS edge passed | provision public-PKI staging name, exact trusted proxies and ingress policy |
| Secret management | BLOCKED | Security owner — unassigned | 2026-08-18 | production validation rejects unsafe values; only disposable local secrets were used | select secret manager; implement identity, injection, rotation and audit procedure |
| Central logs/metrics | BLOCKED | SRE owner — unassigned | 2026-08-18 | authenticated bounded JSON/OpenMetrics sources exist | connect collector, retention, dashboards and per-replica scrape discovery |
| Delivered alerts | BLOCKED | SRE/on-call owner — unassigned | 2026-08-18 | disabled notifier fails honestly; no destination or delivery receipt exists | select destination, assign on-call, deliver and acknowledge every critical test alert |
| Encrypted off-site backup | BLOCKED | Database/security owners — unassigned | 2026-08-18 | checksum local backup/clean restore passed in Phase 11 | schedule encrypted immutable off-site backups and prove access/recovery controls |
| Immutable candidate images | BLOCKED | Release owner — unassigned | 2026-08-18 | no authorized registry or hosted deployment exists | build once in protected CI, sign/attest, record registry digests and promote by digest |
| Migration compatibility | PARTIAL | Release/database owners — unassigned | 2026-08-18 | embedded 19-migration manifest and local clean migration passed | verify the exact candidate image against hosted schema before traffic |

### Infrastructure actually used

No hosted Phase 12 infrastructure was provisioned. The only operational
baseline is the Phase 11 disposable topology: two API replicas, two workers,
PostgreSQL 17, authenticated Redis, private MinIO and a self-signed local TLS
edge, all using fictional seed data. Those containers and volumes were removed
after the exercise. They are not evidence for managed hosting, production
networking, high availability, independent telemetry or capacity.

### Candidate image digests

**None.** Local Docker image identifiers are not registry-backed immutable
candidate digests and were not promoted. Recording them as deployable Phase 12
artifacts would be misleading. Registry digest, signature, SBOM, provenance,
source commit and migration-manifest evidence remain required.

The local BuildKit run emitted local manifest identifiers, but the images were
named with mutable local `:latest` tags, were not pushed to a registry, were not
signed, and were not deployed to hosted staging. They are deliberately excluded
from the candidate-digest register.

## Operational observability

| Signal | Source implemented | Hosted observation | Status | Remaining action |
| --- | --- | --- | --- | --- |
| PostgreSQL pool usage/wait/acquire failure | Yes | none | PARTIAL | scrape every replica and exercise hosted pool saturation |
| Query timeout/cancellation | Partial bounded counters | none | PARTIAL | finish repository-wide instrumentation and trigger a hosted timeout safely |
| Transaction failure | Counter boundary exists | none | PARTIAL | instrument every transaction boundary and prove rollback/error classification |
| Redis failure/latency | Yes | none | PARTIAL | observe managed outage/failover and delivered alert |
| Worker health/backlog/retry/dead jobs/lease recovery | Durable PostgreSQL source | none | PARTIAL | exercise representative hosted queues and backlog recovery |
| Media reconciliation/deletion backlog | Durable PostgreSQL source | none | PARTIAL | run against selected bucket inventory and test discrepancy alert |
| Backup age/failure/restore fingerprint | Durable checkpoint source | none | PARTIAL | connect real backup system and prove age/failure/restore alerts |
| HTTP latency/status/5xx | Fixed bounded histograms | none | PARTIAL | collect fleet rates and validate SLO queries under hosted load |
| Provider failures | Bounded counters and audit records | none | PARTIAL | exercise only approved sandbox adapters |
| Alert delivery | Provider-neutral port | none | BLOCKED | provide destination, delivery receipt, acknowledgement and escalation evidence |

Metrics contain no raw URL, product ID, user ID, click ID, destination, email,
token, credential or arbitrary label. Source availability is not equivalent to
central collection or alert delivery.

## Performance, browser and accessibility

| Gate | Status | Evidence | Limitation / remaining action |
| --- | --- | --- | --- |
| Phase 11 bounded HTTP gates | PASS locally | zero unexpected responses under the recorded short local scenarios | rerun from controlled load generators against hosted ingress |
| Multi-hour soak | BLOCKED | no hosted target or monitoring | run for an approved duration with resource, pool, backlog and error evidence |
| Saturation/capacity | BLOCKED | no hosted target or representative data volume | find safe saturation points and define capacity margins; do not use production data |
| Authenticated admin pagination | BLOCKED | authorization boundary and indexed list tests exist | use a fictional representative admin dataset and measured authenticated session |
| Representative catalog/recommendation | PARTIAL | deterministic local/query-plan suites pass | repeat with approved representative hosted fixtures |
| Lighthouse/Core Web Vitals | NOT TESTED | build-size budgets pass | run mobile/desktop Lighthouse in hosted conditions; add field collection only after privacy review |
| Browser/device matrix | PARTIAL | Phase 11 Playwright: 20 passed, 4 intentional skips | repeat against hosted TLS and complete representative physical-device coverage |
| Automated accessibility | PARTIAL | Axe/keyboard/mobile checks pass locally | repeat hosted and obtain independent WCAG 2.2 AA/manual AT review |

Local measurements in `PERFORMANCE_BUDGETS.md` remain regression evidence only.
No production throughput, concurrency, latency SLO or capacity claim is made.

## Disaster recovery

| Exercise | Status | Observed evidence | RPO | RTO | Remaining action |
| --- | --- | --- | --- | --- | --- |
| Local backup/clean restore | PASS locally | checksum dump restored with 19 migrations in Phase 11 | not a production RPO | not timed as an approved RTO | repeat using encrypted off-site hosted backup |
| Hosted backup | BLOCKED | no hosted database/backup system | NOT MEASURED | NOT MEASURED | configure schedule, retention, encryption and restore access |
| PITR | BLOCKED | no WAL archive/managed PITR target | NOT MEASURED | NOT MEASURED | restore to selected point and reconcile application/media state |
| Database failover | BLOCKED | local stop/start is not managed failover | NOT MEASURED | NOT MEASURED | observe primary loss, promotion, application recovery and data consistency |
| Redis failover | BLOCKED | local outage failed closed but had no managed failover | NOT MEASURED | NOT MEASURED | observe managed failover, limiter correctness and alert delivery |
| Object-storage recovery | BLOCKED | local MinIO outage/recovery only | NOT MEASURED | NOT MEASURED | restore versioned objects/inventory and reconcile database references |
| Application rollback | BLOCKED | forward migration/recreate exercised locally | NOT MEASURED | NOT MEASURED | deploy prior digest through real orchestrator without schema corruption |
| Independently witnessed DR drill | BLOCKED | no independent witness or hosted system | NOT MEASURED | NOT MEASURED | assign witness, approve scenario and sign evidence |

No Phase 12 RPO or RTO is claimed.

## Locally available regression verification

These checks were rerun after the Phase 12 evidence changes. They prove that the
repository baseline did not regress; they do not close any hosted gate.

| Check | Result |
| --- | --- |
| frontend format/typecheck/lint | PASS |
| frontend unit tests | PASS — 23 files, 50 tests |
| frontend production build and gzip budgets | PASS |
| Playwright desktop/mobile/Axe matrix | PASS — 20 passed, 4 intentional skips |
| Go tests | PASS |
| Go race tests | PASS |
| Go vet/build/module verification | PASS |
| npm production audit | PASS — 0 vulnerabilities |
| govulncheck | PASS — 0 reachable/imported-package vulnerabilities; 1 advisory only in an unused required-module package |
| secret-pattern/unsafe-sink/Docker-digest gates | PASS |
| immutable GitHub Action SHA check | PASS |
| Compose config/build/start/health | PASS locally — 2 APIs, 2 workers, PostgreSQL, Redis, MinIO and TLS edge healthy |
| fresh migration | PASS — 19 migrations |
| fictional seed twice | PASS — both runs reported current |
| serialized PostgreSQL/Redis/S3 suite | PASS |
| routing/SEO edge contract | PASS |
| local backup and clean restore | PASS — restored 19 migrations |
| media reconciliation dry run | PASS — 0 objects, references, discrepancies and scheduled jobs |

Bounded local TLS measurements from the Phase 12 rerun:

| Scenario | Requests / concurrency | Error rate | p50 | p95 | p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| readiness | 300 / 16 | 0% | 86.06 ms | 192.96 ms | 291.69 ms |
| catalog | 200 / 12 | 0% | 93.75 ms | 193.13 ms | 202.83 ms |
| recommendation | 12 / 2 | 0% | 7.77 ms | 15.23 ms | 15.23 ms |
| invalid login | 8 / 2 | 0% | 140.70 ms | 188.54 ms | 188.54 ms |
| consented analytics | 40 / 4 | 0% | 6.31 ms | 10.82 ms | 14.42 ms |
| admin authorization boundary | 40 / 4 | 0% | 0.58 ms | 3.31 ms | 3.73 ms |
| genuine public 404 | 100 / 8 | 0% | 1.80 ms | 87.55 ms | 94.33 ms |

No authenticated-admin representative pagination, multi-hour soak,
saturation, Lighthouse, hosted network, managed dependency or real-user metric
was measured.

### Exact verification commands

```sh
cd frontend
npm run format:check
npm run typecheck
npm run lint
npm run test
npm run build
npm run budget:check
npm audit --omit=dev --audit-level=high
E2E_BASE_URL=https://localhost:8443 npm run test:e2e

cd ../backend
go test ./...
go test -race ./...
go vet ./...
go build ./...
go mod verify
govulncheck ./...

cd ..
./scripts/check-secret-patterns.sh
./scripts/check-unsafe-web-sinks.sh
./scripts/check-docker-base-digests.sh backend/Dockerfile frontend/Dockerfile compose.yaml
docker compose --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging --profile validation --profile tools --profile restore config --quiet
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging --profile validation --profile tools --profile restore build
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging up -d --no-build
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging run --rm seed
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging run --rm seed
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging --profile validation run --rm backend-test go test -p 1 -count=1 ./...
./scripts/check-routing-semantics.sh https://localhost:8443
./scripts/run-staging-performance-gates.sh https://localhost:8443
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging --profile tools run --rm media-reconcile
docker compose --project-name unsolero_phase12_local --env-file <temporary-fictional-env> -f compose.yaml -f compose.staging.yaml --profile staging --profile tools run --rm backup
docker compose --project-name unsolero_phase12_restore --env-file <temporary-fictional-env> -f compose.yaml --profile restore run --rm restore
```

`<temporary-fictional-env>` intentionally replaces the exact ignored temporary
file path so documentation does not encourage committing or reusing test
secrets.

### Failures encountered

- The connected GitHub API returned repository `404`, and `gh` is not installed.
  Hosted CI evidence therefore remains blocked; no successful run was inferred.
- Initial npm and Go advisory requests failed because sandbox DNS was disabled.
  Approved network reruns completed: npm found 0 vulnerabilities and
  govulncheck found 0 reachable/imported-package vulnerabilities.
- The first local Compose startup correctly rejected a disposable
  `RATE_LIMIT_KEY_SECRET` encoded with padded rather than raw base64. The
  temporary fictional value was corrected to the documented raw-base64 form;
  application validation was not weakened and the complete affected startup
  path then passed.

## Supply-chain and hosted CI

| Gate | Status | Evidence | Remaining action |
| --- | --- | --- | --- |
| SHA-pinned GitHub Actions | PASS repository | every checked-in `uses:` reference is an immutable commit SHA | retain automated pin review |
| Local dependency/secret/sink/digest gates | PASS locally | Phase 11 npm, govulncheck and deterministic scripts passed | rerun in protected hosted CI |
| Hosted CI execution | BLOCKED | Git remote exists, but no `gh` executable and connected GitHub access returned repository 404 | authorize repository access and capture workflow/run/job evidence |
| Container image scanning | BLOCKED | Trivy workflow definition exists but no hosted run/artifact was available | scan final digest and adjudicate findings |
| SBOM generation | BLOCKED | Anchore workflow definition exists but no artifact was available | retain signed SBOM beside immutable candidate |
| Dependency review | BLOCKED | no protected PR/run evidence | enable and enforce on candidate branch |
| Secret scanning | PARTIAL | deterministic local scan passed; no hosted history/protection evidence | execute hosted full-history scan and enable platform protection |
| Immutable artifact promotion | BLOCKED | no registry, signing identity or deployment platform | define build-once promotion, signatures, attestations and rollback authority |

## Provider sandboxes

No provider sandbox was approved or connected. All runtime provider adapters
remain disabled or fail closed.

| Provider class | Status | Owner | Evidence date | Remaining action |
| --- | --- | --- | --- | --- |
| Transactional email | BLOCKED | Product/security owners — unassigned | 2026-08-18 | approve sandbox/domain, credentials, templates, bounce and complaint process |
| Media scanner | BLOCKED | Security owner — unassigned | 2026-08-18 | select scanner and prove malicious/timeout/unavailable behavior |
| Merchant/offer feeds | BLOCKED | Commerce owner — unassigned | 2026-08-18 | approve sandbox contract and validate idempotency/freshness/reconciliation |
| Affiliate networks | BLOCKED | Commerce/legal owners — unassigned | 2026-08-18 | approve sandbox without live revenue and validate redirect/attribution contract |
| Verified conversions | BLOCKED | Commerce/finance owners — unassigned | 2026-08-18 | approve signed sandbox delivery, replay/reversal/reconciliation evidence |
| AI provider | BLOCKED | Product/privacy/security owners — unassigned | 2026-08-18 | approve sandbox, data handling and structured-output evaluation |
| External analytics | BLOCKED | Privacy/product owners — unassigned | 2026-08-18 | approve purpose, consent, retention and sandbox configuration |
| Alert destination | BLOCKED | SRE/security owners — unassigned | 2026-08-18 | approve destination and test delivery/acknowledgement/escalation |

No production email, customer data, financial event, commission, conversion,
affiliate revenue or provider credential was used.

## Legal and business gates

No approval is inferred. `Unassigned` is itself a launch blocker.

| Gate | Status | Required owner | Evidence | Date | Remaining action |
| --- | --- | --- | --- | --- | --- |
| Privacy notice and lawful basis | BLOCKED | Qualified counsel + privacy owner — unassigned | none | 2026-08-18 | review purposes, regions, consent, processors and data-subject rights |
| Terms of service | BLOCKED | Qualified counsel + business owner — unassigned | none | 2026-08-18 | draft and approve supported service/market terms |
| Affiliate/sponsorship disclosure | BLOCKED | Qualified counsel + commerce owner — unassigned | repository UI disclosure only; no legal approval | 2026-08-18 | review placement, wording, sponsored separation and jurisdictional duties |
| Retention/deletion schedule | BLOCKED | Privacy/legal/security owners — unassigned | engineering defaults documented | 2026-08-18 | approve every class, backup propagation, holds and audit process |
| Provider/processor agreements | BLOCKED | Legal/procurement owners — unassigned | none | 2026-08-18 | execute required agreements and review subprocessors/transfers |
| Supported markets/currencies | BLOCKED | Business/legal owners — unassigned | none | 2026-08-18 | approve launch scope and consumer obligations |
| Incident notification ownership | BLOCKED | Legal/security/executive owners — unassigned | technical runbook only | 2026-08-18 | assign decision authority, contacts and notification procedure |
| Launch authorization | BLOCKED | Accountable executive — unassigned | none | 2026-08-18 | sign only after every applicable evidence gate is current |

## Independent validation

| Gate | Status | Required evidence |
| --- | --- | --- |
| Penetration test | BLOCKED | independent report against the actual hosted candidate and remediation closure |
| Security architecture review | BLOCKED | review of real identities, network, secrets, storage, CI and deployment controls |
| WCAG/manual accessibility | BLOCKED | independent WCAG 2.2 AA, screen-reader, zoom/reflow, contrast and device evidence |
| DR witness | BLOCKED | witnessed hosted backup/PITR/failover/rollback exercise with measured RPO/RTO |
| Legal/privacy review | BLOCKED | written approval from qualified reviewers for selected launch regions |

## Verdict rationale

Phase 12 is `PARTIAL` because the repository/local controls remain valid and
the external gate inventory is now explicit and auditable, but the primary
objective—hosted production-equivalent staging with independently verifiable
launch evidence—could not be executed without authorized infrastructure,
credentials, provider sandboxes and accountable reviewers. Documentation is
not a substitute for those systems or approvals.
