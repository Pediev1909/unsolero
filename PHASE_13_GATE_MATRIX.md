# UNSOLERO Phase 13 gate matrix

Status: **repository closure executed; hosted/external gates BLOCKED**  
Prepared: 2026-08-18

## Repository execution evidence — 2026-08-18

This execution completed every safe repository/local action available without
inventing a cloud account or external approval. It added production dependency
validation, an authenticated alert-webhook adapter, durable worker heartbeat
metrics, PostgreSQL runtime grants, encrypted-backup handoff, comprehensive
image/SBOM/dependency-review workflows, and immutable candidate promotion by
digest. Validation also fixed the Compose test service: it now runs the test
and race suites and uses a CGO-capable validation stage.

Final observed results:

- frontend format/typecheck/lint: PASS; Vitest: 23 files and 50 tests PASS;
  production build and both bundle budgets PASS;
- Playwright: 21 PASS, 3 intentional project-matrix skips, including 320, 375,
  390, 430, 768, 1024, 1280, 1440, and 1920 px coverage and Axe checks;
- Go unit/integration and race suites, vet, build and module verification: PASS;
- govulncheck: zero reachable/imported vulnerabilities; one required-module
  advisory is not called; npm production audit: zero vulnerabilities;
- fresh PostgreSQL: 20 of 20 migrations; fictional seed succeeds idempotently;
  isolated PostgreSQL/Redis/S3 integration and race suites PASS;
- PostgreSQL grant exercise: API DML allowed, schema creation denied, backup
  SELECT allowed, backup INSERT denied;
- checksum logical backup and clean restore: PASS, with 20 migrations present;
- workflow YAML, immutable action SHA, Docker base digest, secret-pattern,
  unsafe-web-sink, Compose and shell syntax gates: PASS;
- local API health/catalog/unknown-route smoke: HTTP 200/200/404 as expected.

The final local BuildKit image IDs were: API
`sha256:9a659eac24b0c3ebfd9d0284ec8f86db75aa195405a4c9f8f319c4c6690b0344`,
worker `sha256:c21c1b2cdd79bdccd54c61b897d242f7c6a5c695c04d7279ce97b2bf9641d90c`,
migration `sha256:e73a72ba9a38514da714623f5e7054cec42cf2297ae5b84d956d703ce93031a1`,
media reconciliation
`sha256:97a9b9dd2fe2f934156e8439c44c0ceefca7871366fc9bebd0cdc48110f0c6c7`,
media initialization
`sha256:0fd74a155242131cd49de6020f22983c29a111fa7251d431df0e39e6066339f2`,
development web
`sha256:0e06e510e927f56143ecea36fb451d17e4996bf9a1a8a97708cbe29e2924b7a0`,
and production web
`sha256:bb16d56192d3809ba54d89a4c1b364c2d7721c83a16d9aeeb96f1e05cee2d1fc`.
These are local image identities only: they were not pushed, signed, deployed,
or verified from an immutable registry and therefore do not satisfy G04/G05.

The npm audit and localhost socket probes initially failed because the sandbox
blocked network/socket access; they passed after narrowly scoped approval. Go
checks initially hit a read-only host cache and passed with isolated
`GOCACHE`/`GOTMPDIR`. The staging Compose config initially failed because its
intentional secret placeholders were empty; it passed with validation-only,
non-secret values and the staging profile. No control was weakened.

No hosted resource, live provider, immutable registry artifact, signature
verification, alert receipt, off-site ciphertext, PITR, managed failover,
multi-hour hosted soak, independent assessment, or legal approval was observed.

### Material commands executed

Environment inspection used `rg --files`, `rg`, `sed`, `git status --short`,
`git diff --check`, `docker compose config --services`, and `docker compose ps`.
The final verification commands were:

```text
cd frontend && npm run format:check
cd frontend && npm run typecheck
cd frontend && npm run lint
cd frontend && npm run test
cd frontend && npm run build
cd frontend && npm run budget:check
cd frontend && npm audit --omit=dev --audit-level=high
cd frontend && npm run test:e2e

cd backend && test -z "$(gofmt -l .)"
cd backend && GOCACHE=/tmp/unsolero-go-cache GOTMPDIR=/tmp/unsolero-go-tmp go test -p 1 -count=1 ./...
cd backend && GOCACHE=/tmp/unsolero-go-cache GOTMPDIR=/tmp/unsolero-go-tmp go test -race -p 1 -count=1 ./...
cd backend && GOCACHE=/tmp/unsolero-go-cache GOTMPDIR=/tmp/unsolero-go-tmp go vet ./...
cd backend && GOCACHE=/tmp/unsolero-go-cache GOTMPDIR=/tmp/unsolero-go-tmp go build ./...
cd backend && go mod verify
cd backend && GOCACHE=/tmp/unsolero-go-cache GOTMPDIR=/tmp/unsolero-go-tmp govulncheck ./...

docker compose --env-file .env.example config --quiet
docker compose --profile staging --env-file .env.staging.example -f compose.yaml -f compose.staging.yaml config --quiet
docker compose -p unsolero-prodready --profile validation run --rm backend-test
docker compose -p unsolero-prodready --profile staging --env-file .env.example build api commerce-worker migrate media-reconcile media-init web
docker build --target production -t unsolero-prodready-web-production:local frontend
docker image inspect unsolero-prodready-api:latest unsolero-prodready-commerce-worker:latest unsolero-prodready-migrate:latest unsolero-prodready-media-reconcile:latest unsolero-prodready-media-init:latest unsolero-prodready-web:latest unsolero-prodready-web-production:local
docker compose -p unsolero-prodready --env-file .env.example down --remove-orphans

scripts/check-docker-base-digests.sh backend/Dockerfile frontend/Dockerfile compose.yaml
scripts/check-secret-patterns.sh
scripts/check-unsafe-web-sinks.sh
sh -n scripts/encrypt-backup-age.sh scripts/decrypt-backup-age.sh scripts/backup-postgres.sh scripts/restore-postgres.sh
```

The staging Compose configuration command supplied validation-only values for
its intentionally blank required variables; values are omitted from evidence.
Database migration/seed, grant, backup/restore, routing, load, browser and fault
command details from the inherited local staging exercise remain in
`docs/PHASE_11_EVIDENCE.md`; this addendum does not fabricate or rewrite that
record.

`PASS` requires the listed evidence from the authorized hosted environment.
Repository definitions, provider marketing, local results, or an unobserved
capability never count as hosted evidence. Gates are ordered; a failed required
gate stops downstream execution unless a named owner records a time-bounded,
non-security exception. No exception can authorize production traffic.

## Evidence rules

Every evidence record must contain:

- gate ID, status (`PASS`, `FAIL`, `BLOCKED`, `NOT APPLICABLE`), owner and UTC
  date/time;
- source SHA, candidate manifest ID, image digests and environment resource IDs;
- exact command or provider action, expected result, actual result and exit
  status without secret values;
- immutable/raw artifact location and a human-readable summary;
- failure, remediation, rerun and remaining limitation;
- classification as `internally verified` or `independently verified`.

## Ordered execution gates

| ID | Gate | Entry condition | Required PASS evidence | Owner | Current status |
|---|---|---|---|---|---|
| G00 | Scope and safety authorization | none | approved provider/region, fictional-data rule, $250 budget alert, destructive-test window, teardown scope, named owners | executive/platform/security | BLOCKED |
| G01 | Isolated access | G00 | DigitalOcean project/team, MFA membership, least-privilege access inventory, GitHub staging environment, DNS scope, telemetry workspace and revocation paths | platform/security | BLOCKED |
| G02 | Repository baseline | G01 | clean review of inherited changes; frontend/backend/database/security/local staging suites still pass on candidate branch | release | BLOCKED — hosted branch unavailable |
| G03 | Infrastructure definition review | G02 | reviewed App Spec/IaC plan, no secrets in state/repo, private network/trusted sources, resource limits, probe policy, teardown plan | platform/security | BLOCKED — definition not yet implemented |
| G04 | Supply chain | G02 | hosted CI required checks; all role images scanned; source/per-image SBOMs; secret/SAST/dependency review; immutable action SHAs; signature and provenance verification | release/security | BLOCKED |
| G05 | Immutable candidate | G04 | candidate manifest ties source SHA, all image digests, SBOM/scan/signature, migration fingerprint and approver; no mutable reference in App Spec | release | BLOCKED |
| G06 | Managed data plane | G03 | private TLS PostgreSQL 17 + standby, separate tested roles, bounded connection budget, private TLS Valkey + standby/persistence/eviction, provider metrics | database/platform | BLOCKED |
| G07 | Object storage and recovery boundary | G03 | private versioned media bucket, limited app key, separate-team/region recovery bucket/key, inventory baseline, access denial from public/unauthorized identity | platform/security | BLOCKED |
| G08 | Secrets/configuration | G03/G06/G07 | encrypted runtime values, bindable service endpoints, rotation/rollback drill for one non-destructive secret class, no secret in image/state/log/artifact | security/platform | BLOCKED |
| G09 | DNS/TLS/ingress | G03 | staging hostname, valid TLS chain/minimum version, web-only public ingress, API/metrics/private dependencies unreachable publicly, proxy trust documented | platform/security | BLOCKED |
| G10 | Migration and deployment | G05–G09 | PRE_DEPLOY migration digest succeeds once; exact schema manifest ready; two healthy web/API/worker instances; deployment record uses candidate digests | release/database | BLOCKED |
| G11 | Functional hosted smoke | G10 | health, auth/MFA, admin denial/pagination, catalog, deterministic recommendation, media, analytics consent, affiliate stale/ownership protection, 404/SEO/routing all pass with fictional data | application/release | BLOCKED |
| G12 | Central telemetry | G10 | logs from every component, per-replica OpenMetrics scrape, durable gauges, provider metrics, dashboards, uptime/heartbeat monitors; no sensitive labels/payloads | telemetry/security | BLOCKED |
| G13 | Delivered alert | G12 | a dedicated test rule fires and a real destination receives/acknowledges it; trigger/delivery/ack timestamps recorded; resolution also observed | telemetry/on-call | BLOCKED |
| G14 | Performance and soak | G11–G13 | approved realistic load, multi-hour soak, p50/p95/p99/error/timeouts, CPU/RAM, pool/DB/Valkey/storage, backlog and telemetry continuity; no real-user traffic | performance/platform | BLOCKED |
| G15 | Application/process resilience | G11–G13 | API and worker replica-loss exercises pass; health routing, durable leases, graceful recovery and alerts observed | platform/application | BLOCKED |
| G16 | Managed dependency failure/failover | G11–G13 | PostgreSQL and Valkey outage/failover behavior, object-storage denial/recovery, bounded errors, readiness and delivered alerts observed; unsupported provider action remains BLOCKED | database/platform | BLOCKED |
| G17 | Backup and restore | G07/G10/G13 | managed backup visible, encrypted logical copy in recovery target, clean hosted restore, fingerprint/invariant/smoke validation, measured times, backup-failure/stale alert | database/platform | BLOCKED |
| G18 | PITR and database failover | G16/G17 | restore to selected timestamp in isolated cluster, committed facts bounded around recovery point, managed standby promotion/application recovery, measured RPO/RTO | database/DR witness | BLOCKED |
| G19 | Media recovery | G07/G17 | known fictional object set copied, primary loss/denial simulated, recovery objects verified by digest/inventory, DB/media reconciliation clean, measured time | platform/DR witness | BLOCKED |
| G20 | Application rollback | G10/G13 | prior candidate digest restored; probes/smoke/metrics pass; schema compatibility decision recorded; rollback time measured | release/application | BLOCKED |
| G21 | Browser/performance/accessibility | G11 | automated viewport/browser matrix, Lighthouse budgets, Axe; manual keyboard/screen-reader/zoom/contrast review with defects tracked | UX/accessibility | BLOCKED — manual independence unavailable |
| G22 | Provider sandbox certification | G11/G13 | only separately approved non-live sandboxes exercised; signatures/replay/failure behavior evidenced; all absent adapters recorded BLOCKED | commerce/security | BLOCKED |
| G23 | Independent penetration/manual security | G10–G14 | signed scope and delivered external report; critical/high remediated and retested, or explicitly unresolved | independent security assessor | BLOCKED |
| G24 | Independent accessibility | G21 | delivered manual WCAG report from an accountable reviewer and remediation/retest record | independent accessibility assessor | BLOCKED |
| G25 | Independent DR/infrastructure review | G17–G20 | witnessed DR record plus read-only IAM/network/backup/config review and findings disposition | independent DR/infrastructure assessor | BLOCKED |
| G26 | Privacy/legal/business gates | G11/G22 | owner/date/evidence/action for privacy, retention, cookies, affiliate disclosure, terms, provider contracts and launch authority; no compliance claim | privacy/legal/business | BLOCKED |
| G27 | Final full verification | G14–G26 | all relevant frontend/backend/database/integration/race/vulnerability/routing/performance/security suites rerun against final source/digests; exact results | release/security | BLOCKED |
| G28 | Phase verdict | G27 | Phase 13 evidence document distinguishes PASS/FAIL/BLOCKED, scores evidence only, and does not declare production readiness | CTO/release | BLOCKED |

## Required security and review separation

| Review | Internal gate | Independent gate | Current status |
|---|---|---|---|
| automated source/dependency/container security | G04/G27 | not a penetration test | BLOCKED hosted |
| manual application security review | G11/G16/G22 | G23 | BLOCKED |
| automated accessibility | G21 | not independent/manual coverage | BLOCKED hosted |
| manual accessibility | internal UX review in G21 | G24 | BLOCKED |
| backup/restore/failover | engineering exercise G17–G20 | witnessed G25 | BLOCKED |
| IAM/network/configuration | internal G01/G03/G08/G09 | infrastructure review G25 | BLOCKED |
| privacy/legal | engineering data-flow evidence | accountable counsel/reviewer G26 | BLOCKED |

Internal PASS must never be relabeled independent PASS. A missing assessor is
`BLOCKED`, not `NOT APPLICABLE`.

## Stop conditions

Stop deployment or testing immediately if any of these occurs:

- production/customer/financial data or a live provider credential is found;
- an image digest/signature/manifest does not match;
- migration checksum or schema readiness fails;
- a secret appears in output, state, logs, artifacts, browser assets or chat;
- public access reaches API-private metrics, managed data stores, or media
  listing;
- readiness fails unexpectedly, errors are unbounded, or data integrity is
  uncertain;
- backup/recovery target cannot be protected from application credentials;
- the alert destination is silent during a test that depends on paging;
- cost, rate, test-window or provider acceptable-use limits are approached;
- an operator cannot reverse the proposed fault safely.

Preserve evidence, revoke affected credentials, return traffic/test load to
zero, and follow the incident runbook. Never weaken fail-closed controls to
continue a gate.

## Launch interpretation

- Phase 13 `PASS` means the authorized hosted staging plan and its evidence
  passed. It does not itself authorize production traffic.
- Any required hosted gate `FAIL` yields Phase 13 `FAIL` until fixed and rerun.
- Any required external/independent gate `BLOCKED` yields at most `PARTIAL`.
- Legal approval, penetration testing, accessibility review, and witnessed DR
  cannot be inferred from internal work.

Current Phase 13 verdict: **PARTIAL — repository closure PASS; hosted and
external launch gates BLOCKED. Final launch verdict: NOT PRODUCTION READY.**
