# UNSOLERO CI and security gates

Status: deterministic local gates pass; hosted/external scanner execution not yet observed  
Last reviewed: 2026-08-18

`.github/workflows/ci.yml` defines frontend, backend/database, local-stack, and
Chrome browser gates. `.github/workflows/security.yml` defines dependency
review, secret-history, SAST, filesystem/container vulnerability, base-image
pin, and per-image SPDX SBOM gates. `.github/workflows/release-candidate.yml`
manually builds a reviewed full commit SHA, scans it before publishing, and
records the GHCR digest for each deployable role. Repository workflow
configuration is not a passing hosted CI result.

The manual candidate workflow has an explicit `sign_with_sigstore` input. When
the owner enables it, the workflow uses GitHub OIDC and a SHA-pinned Cosign
installer to sign the exact published digest. A checked-in optional step is not
signature evidence: the workflow run, identity, transparency record, and
verification policy must all be observed before the signature gate can pass.

## Merge-gate policy

| Gate | Failure policy | Local evidence |
| --- | --- | --- |
| formatting, TypeScript, ESLint, Vitest, Vite build | fail | executable locally |
| npm high-severity audit | fail | executable locally |
| frontend gzip/CSS/initial bundle budgets | fail | executable locally |
| npm lockfile consistency and Go checksum verification | fail | executable locally |
| gofmt, unit/integration tests, race, vet, build | fail | executable locally |
| fresh migration and seed twice | fail | executable with PostgreSQL 17 |
| Redis/MinIO adapter integration | fail | CI starts isolated digest-pinned services; exercised locally |
| govulncheck reachable vulnerabilities | fail | executable when tool/module data available |
| Compose render/build/readiness | fail | executable with Docker |
| Playwright version-pinned Chromium and material Axe findings | fail | executable when browser/system dependencies exist |
| gitleaks history scan | fail | configured, not executed in this environment |
| Semgrep default SAST | fail | configured, not executed in this environment |
| Trivy HIGH/CRITICAL filesystem/image findings | fail | configured, not executed in this environment |
| unpinned Docker base image | fail | configured and locally reviewable |
| unpinned Compose image | fail | deterministic repository script |
| secret-pattern and unsafe web-sink review | fail | deterministic repository scripts |
| SPDX JSON SBOM generation | fail | configured, not executed in this environment |
| pull-request dependency review | fail on newly introduced HIGH/CRITICAL findings | configured with immutable action SHA; hosted result not observed |

All workflow tokens have read-only repository permission. Third-party actions
are pinned to the exact commits resolved for their documented version tags on
2026-08-17, and CLI tools use explicit versions. A security owner must still
review those commits and approve future automated updates before accepting them.
Dependency-review availability for the repository visibility/license,
Dependabot/Renovate policy, required checks, protected environments, signed
release provenance, deployment approval, and artifact registry retention are
external repository/hosting configuration.

## Supply-chain and artifact requirements

- Docker base images are digest pinned. Application dependencies are lockfile
  or Go module checksum constrained.
- Build once and deploy the recorded digest; do not rebuild per environment.
  The release-candidate workflow publishes only commit-addressed candidate tags
  and performs no environment deployment.
- Generate and retain SBOM, image scan result, source revision, workflow run,
  migration manifest, and artifact SHA-256 together.
- Sign images and attest provenance in the selected registry. Signature and
  attestation verification must be an admission/deployment gate.
- Scanner suppressions require an owner, evidence, expiry, and review. Unknown
  or unavailable scanners are BLOCKED, never PASS.
- External rule feeds and vulnerability databases are network dependencies;
  inability to refresh them must be visible and must fail release policy after
  the approved cache window.

## Required branch and environment controls

Require review, passing `CI` and `Security gates`, no force push, resolved
threads, protected production environment approval, least-privilege OIDC, and
separate migration/deploy identities. Fork pull requests must never receive
production secrets. Deployment credentials are unavailable to build/test jobs.

## Current limits

No GitHub-hosted workflow has been observed from this workspace. Gitleaks,
Semgrep, Trivy, and Syft/Anchore results therefore remain NOT TESTED here. The
workflow itself and the pinned action commits still need an independent
supply-chain review before they become a trusted production gate.

Phase 10 pins PostgreSQL, Redis, and MinIO Compose/CI images by digest. The
backend CI job supplies isolated test-only Redis and S3 credentials and runs the
gated integration tests as part of the serialized full suite. No provider or
production credential is available to CI. Dependency review is now configured
but remains explicitly NOT TESTED until a pull-request workflow result is
observed; `npm audit` and `govulncheck` are complementary, not substitutes.

Phase 11 adds `go mod verify`, production-only npm audit policy, deterministic
Dockerfile/Compose digest checks, secret-pattern review, unsafe web-sink review,
and bundle budgets. Local deterministic scripts passed. Gitleaks history,
Semgrep, Trivy, SBOM generation, signed provenance, and hosted branch-policy
evidence remain NOT TESTED rather than inferred from local checks.
