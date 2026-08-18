# UNSOLERO launch governance

Status: human/legal/business approval inventory; nothing here is legal advice  
Last reviewed: 2026-08-17

Engineering cannot mark these items complete. Each needs a named human owner,
qualified counsel where applicable, approved text/policy, jurisdiction and
version, effective date, evidence location, change process, and review date.

| Approval item | Required decision/evidence | Proposed owner | State |
| --- | --- | --- | --- |
| Privacy policy | actual data inventory, purposes, lawful bases, recipients, transfers, retention, rights, contact and versioning | Privacy/legal | LEGAL |
| Terms of service | eligibility, service limits, product-information disclaimers, account rules, liability, disputes and governing jurisdiction | Legal/business | LEGAL |
| Affiliate disclosure | clear placement/language for pages, recommendations, CTAs and social/editorial channels; program-specific terms | Legal/commerce | LEGAL |
| Cookie/analytics consent | jurisdiction-aware categories, optional analytics default, withdrawal, proof and UI copy | Privacy/legal/product | LEGAL |
| Retention policy | approve every operational, security, analytics, commerce and financial retention period and deletion exception | Privacy/security/legal | LEGAL |
| Data-subject requests | verified intake, ownership, identity checks, export/delete scope, deadline, exceptions and audit evidence | Privacy/support | LEGAL |
| Processor agreements | DPA/SCC/transfer assessment and subprocessor register for hosting, email, monitoring, media, AI and analytics | Legal/security | LEGAL |
| Merchant/affiliate agreements | program terms, allowed traffic/content, pricing/freshness, tracking, reversals, tax and termination | Commerce/legal | LEGAL |
| Jurisdictional requirements | launch countries/states, consumer, privacy, accessibility, marketing and tax obligations | Legal/business | LEGAL |
| Minors policy | audience age floor, age assurance/parental requirements and deletion/contact flow | Legal/product | LEGAL |
| Incident notification | regulator/user/partner thresholds, clocks, counsel, insurer and evidence-preservation paths | Security/legal | LEGAL |
| Financial record retention | verified conversion/commission/reversal evidence, reconciliation, tax/accounting retention and access | Finance/legal | LEGAL |
| Editorial/evidence policy | conflicts, sponsorship labels, correction/retraction, reviewer independence and demo-data treatment | Editorial/legal | LEGAL |
| Accessibility position | supported standards, testing evidence, feedback channel and remediation ownership; no unsupported compliance claim | Legal/product/accessibility | LEGAL |

## Launch decision authority

The launch approver must receive the technical scorecard, open risk register,
independent security/accessibility results, DR exercise, provider checklist,
capacity evidence, incident contacts, and approved legal artifacts. Approval is
time-bound and tied to an immutable release. A repository document is not an
approval record.

Public traffic, live providers, production credentials, conversion ingestion,
or financial reporting remain prohibited until all applicable approvals and
technical/external gates are satisfied.

## Phase 10 go/no-go rule

Current decision: **NO-GO**. Engineering is not authorized to reinterpret
`PARTIAL`, `BLOCKED`, `EXTERNAL`, `LEGAL`, or `NOT TESTED` as approval.

A future go decision requires one immutable release and named evidence for:

- production-equivalent staging, migration, rollback, capacity, Redis outage,
  private media/scanner, email, telemetry, alerting, backup, and witnessed restore;
- passing protected CI/security/SBOM/image/browser/accessibility gates;
- independent penetration and accessibility assessments with high-severity
  remediation closed or explicitly accepted by accountable executives;
- provider-specific sandbox certification and two-person activation approval;
- approved policies, terms, disclosures, retention, processors, markets,
  incident notification, support, and financial governance;
- a staffed on-call rota, incident commander, provider owners, database owner,
  privacy contact, and launch/rollback authority.

Any critical readiness failure, stale evidence, reconciliation uncertainty,
scanner/limiter/secret/alert outage, or loss of required approval automatically
returns the launch to no-go. Commercial pressure cannot waive recommendation
independence, verified-data rules, privacy consent, or security controls.
