# UNSOLERO Phase 13 access requirements

Status: **all external access BLOCKED until supplied and authorized**  
Prepared: 2026-08-18

This list follows least privilege. It does not ask anyone to paste passwords,
API tokens, private keys, recovery codes, or connection strings into chat,
issues, commits, logs, screenshots, or evidence documents. Secrets must be
created directly in the owning platform's protected secret field.

## 1. Ownership and authorization

The human operator is the sole project owner and final authority. The labels
below are responsibilities the owner must explicitly accept or later delegate;
they are not imaginary people or evidence of review.

| Role | Minimum responsibility | Current status |
|---|---|---|
| human owner | budget, platform, database, security, telemetry, release and incident authority | AVAILABLE only when explicitly exercised; no approval inferred |
| qualified legal reviewer | privacy, terms, affiliate, cookie and jurisdiction review | BLOCKED — external future reviewer not engaged |
| independent assessor(s) | penetration, accessibility, DR witness and infrastructure review | BLOCKED — not engaged |

Required written authorization:

- fictional/non-production data only;
- maximum monthly staging spend and automatic budget alert;
- DigitalOcean Frankfurt as the selected target, or a documented alternative;
- exact start/end of load, outage, failover, PITR and rollback windows;
- resources that may be restarted, failed over, cloned, or deleted;
- explicit prohibition on live merchant, conversion, affiliate revenue, AI,
  customer, financial, production email, and production alert-provider use;
- evidence retention and who may access it.

## 2. Existing Vercel project and domain discovery

The repository and current environment expose no Vercel project link/token,
project/team identifier, domain value, or deployment evidence. To inspect the
existing project without changing production traffic, the human owner must:

1. confirm the exact existing production domain and Vercel project/team names;
2. invite a project-scoped deployment identity, or run the documented
   inspection commands locally while authenticated;
3. authorize preview-only configuration review; and
4. withhold production DNS/traffic changes until separate explicit approval.

Do not paste a Vercel token into chat or a repository file. Store a deployment
credential only in a protected CI environment. Vercel cannot host the current
Go worker, managed PostgreSQL, managed Redis/Valkey, private media store, or
durable backup scheduler; a backend runtime remains separately required.

## 3. Proposed DigitalOcean access

Create an isolated team/project named for UNSOLERO staging. Do not reuse a
production/customer project. Enable MFA for human members.

### Human bootstrap identity

Needed once to create the project, billing cap, VPC, managed services, registry,
Spaces keys, app, and protected team membership. Use the narrowest DigitalOcean
team role that can perform those tasks. Owner/billing access is not needed for
ordinary deployment after bootstrap.

Supply to the platform owner through the provider console:

- staging project ID and approved Frankfurt region/datacenter;
- permission to create/manage one App Platform app, VPC attachment, PostgreSQL
  cluster, Valkey cluster, DOCR registry, two Spaces subscriptions/buckets,
  trusted sources, alerts and budget controls;
- access to view provider events, backups, metrics and audit information;
- authority to create separate application, migration, worker, backup and
  monitoring database users;
- authority to perform the specifically approved failover/PITR/restore tests.

Do not grant the application or CI token team-owner, billing, user-management,
resource-delete, or unrestricted Spaces access.

### CI deployment identity

Create a staging-only machine token or supported app credential that can:

- push/read only the UNSOLERO DOCR repositories;
- read the current staging App Spec and deployments;
- update only the staging app's image digests and start/inspect deployments;
- read deployment health/log metadata needed by the gate;
- invoke approved staging jobs and provider rollback.

It must not manage billing, team membership, unrelated projects, DNS zones,
managed-database deletion, backup deletion, Spaces recovery data, or production
resources. Store it only in the protected GitHub `staging` environment. If the
provider cannot express these scopes, retain deployment as a human approval
step rather than granting a broad unattended token.

### Managed PostgreSQL

Required resources/credentials, all staging-only:

- private TLS connection/certificate to PostgreSQL 17 primary + standby;
- migration role: schema DDL and migration-ledger ownership only;
- API role: only online API schema/table/sequence privileges;
- worker role: only worker/cleanup/import/conversion privileges;
- backup role: read-only plus the minimum metadata access needed by the backup;
- monitoring role: provider-supported read-only metrics/query insight access;
- temporary restore-test role and isolated restored database/cluster;
- permission for the database owner to list backups and perform PITR/failover
  during the approved test window.

Connection strings belong in App Platform encrypted variables/bindable values.
They are never supplied in documentation. Runtime and backup grants are now
codified in `scripts/postgres-runtime-grants.sql` and exercised locally; the
gate remains BLOCKED until separate hosted identities actually receive and pass
those grants.

### Managed Valkey

Required:

- private TLS/auth endpoint for a 2 GiB primary + standby;
- app credential restricted to that cluster;
- provider settings/read access for persistence, eviction, maintenance,
  trusted sources, metrics and events;
- database/platform-owner authority for a reversible outage/failover exercise.

The Valkey credential must not be shared with CI tests, local development, or
other applications.

### Spaces object storage

Create two private, versioned buckets:

1. primary media in FRA1;
2. recovery in a separate DigitalOcean team/account and region.

Create separate limited keys:

- application media key: read/write/delete only on the primary media bucket;
- media-copy source key: read/list primary only;
- recovery writer: append/write/list recovery only where provider controls
  allow it; no application access;
- recovery reader: normally disabled or held by the database/DR owner;
- infrastructure bootstrap key: temporary full configuration authority,
  revoked after bucket/versioning/key setup.

Spaces keys are created in the provider console. Never put the recovery key in
the App Platform API/worker environment. Spaces has no built-in backup, so the
recovery-copy gate must be exercised rather than inferred from versioning.

### DNS/TLS

Needed only for a staging hostname such as `staging.<approved-domain>`:

- ability to create/update that single host's CNAME/A/verification records;
- ability to inspect certificate issuance and DNS propagation;
- no registrar transfer, nameserver replacement, apex-domain write, unrelated
  record access, or parent-domain HSTS change.

If DigitalOcean hosts the zone, scope the human task to the one record. If the
DNS provider cannot offer record-level automation, use a human DNS change with
recorded before/after values.

## 4. GitHub repository access

The observed remote is `Pediev1909/unsolero`, but Phase 12 could not access the
repository through GitHub tooling. Required access:

- repository read/write for the release owner to add reviewed Phase 13
  infrastructure/workflows;
- Actions enabled with permission to publish artifacts and packages/DOCR;
- repository administrator for one-time branch rules, required checks,
  environment protection, Actions policy, secret scanning/dependency review
  availability, and reviewer assignment;
- a protected `staging` environment with named approvers and no secrets exposed
  to pull requests or forks;
- ability to retain/download candidate manifests, SBOMs, scan results,
  signatures/attestations and test evidence;
- package/registry access only if GHCR is selected instead of DOCR.

The runtime does not need a GitHub token. Build/test jobs do not receive cloud
deployment secrets. Deployment starts only from an approved workflow on a
protected commit after required checks pass.

## 5. Telemetry and alert delivery

Create a Better Stack EU workspace or approve an equivalent provider. Required
least-privilege assets:

- log source token for App Platform forwarding;
- Prometheus remote-write endpoint/credentials for the private collector;
- dashboard/editor access for the telemetry owner;
- read-only evidence access for reviewers;
- uptime monitors for TLS/readiness/404 semantics;
- heartbeats for logical backup and media reconciliation;
- one real alert destination owned by an accountable person or team, such as a
  dedicated staging-alert email and/or Slack channel;
- permission to create a dedicated test alert and acknowledge it.

Tokens are component-specific and stored as App Platform encrypted variables.
Do not use a personal catch-all API token if source-specific ingestion tokens
exist. No log/metric source may receive request bodies, IPs, emails, cookies,
tokens, product free text, affiliate destinations, or credentials.

## 6. Optional approved sandboxes

These are not permission to activate live providers.

| Sandbox | Minimum access | Default if absent |
|---|---|---|
| email | non-delivering SMTP sandbox inbox and staging-only credential | `EMAIL_PROVIDER=disabled`; verification/reset delivery exercise BLOCKED |
| merchant feed | provider's documented sandbox read credential and fictional catalog | provider remains disabled; adapter certification BLOCKED |
| conversion webhook | sandbox signing key and fictional events only | endpoint remains unexercised externally; no revenue metrics |
| media scanner | reviewed sandbox/API that never receives real user data | development validator only; public upload launch gate BLOCKED |
| AI | no sandbox is needed for Phase 13 | `AI_PROVIDER=disabled` |

Amazon Associates, Awin, Impact, CJ, direct merchant, payment, live conversion,
production SMTP, customer analytics, and live AI credentials are explicitly not
requested.

## 7. Independent verification access

Provide named vendors/assessors and narrow test authorization for:

- external penetration test against the staging hostname and fictional test
  accounts, with rate/load limits and emergency contact;
- manual WCAG/accessibility review and approved browser/assistive-technology
  matrix;
- DR witness allowed to observe backup, PITR, failover, rollback and evidence;
- independent infrastructure/IAM/network/configuration review with read-only
  exports, not standing admin access;
- privacy/legal review of retention, disclosures, affiliate behavior, data
  deletion/export and provider terms.

If these people are not actually engaged and reports are not delivered, their
gates remain `BLOCKED`. Internal test output cannot substitute for independence.

## 8. Authorized fictional test identities

Required:

- one ordinary verified fictional user;
- one fictional user with MFA enabled;
- one least-privilege staging admin created through the documented account and
  role process;
- one deliberately unauthorized fictional user;
- provider sandbox identities only where approved.

Never copy a real person's account, email, product purchase, click, conversion,
commission, or analytics history. Test credentials live in the protected
staging test secret scope and are rotated/deleted after the phase.

## 9. What must not be supplied

- cloud owner/root credentials or billing card details;
- production database, Redis, bucket, DNS, email or provider credentials;
- customer exports, database snapshots, production logs, analytics or media;
- live affiliate/conversion secrets or financial data;
- MFA recovery codes or a developer's personal session/token;
- broad GitHub personal tokens when an environment-scoped machine identity is
  available;
- secrets pasted into chat or committed `.env` files.

## 10. Access acceptance checklist

Access is accepted only when:

1. owner, scope, environment, issue date, rotation date and revocation path are
   recorded without secret values;
2. MFA and least privilege are confirmed;
3. a read-only or narrowly scoped test proves the credential works;
4. logs/artifacts show no secret value;
5. break-glass and teardown owners are named;
6. production/live-resource visibility is absent or explicitly denied.

## Single next action

The human owner must first return only the exact existing domain and Vercel
project/team names (no token), then explicitly choose and authorize a maximum
monthly budget for an isolated backend staging platform. If DigitalOcean
Frankfurt is approved, create the isolated project and budget alert in the
provider console and supply only its non-secret project ID/region. Do not send
credentials. Hosted execution cannot begin before those decisions exist.
