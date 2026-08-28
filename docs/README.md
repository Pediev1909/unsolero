# Documentation index

Fifty-six markdown documents live here and at the repository root. This page
exists so none of them is findable only by remembering it. Each line says what the
document decides, so you can tell from here whether you need to open it.

Start with [../README.md](../README.md) for setup and
[../ARCHITECTURE.md](../ARCHITECTURE.md) for how the system is shaped.

## Money and partners

| Document | What it settles |
| --- | --- |
| [AFFILIATE_PROGRAMS.md](./AFFILIATE_PROGRAMS.md) | **Which programmes have approved us and which link is live where.** Start here. |
| [AFFILIATE_LINK_AUDIT_2026_08_26.md](./AFFILIATE_LINK_AUDIT_2026_08_26.md) | The audit method, and why no affiliate URL is ever followed automatically |
| [affiliate-links-zoho.md](./affiliate-links-zoho.md) | The 91-row Zoho link reconciliation and its transcription trap |
| [affiliate-links-mailerlite.md](./affiliate-links-mailerlite.md) | The Trackdesk link, its parameters, and why none may be edited |
| [LAUNCH_CHECKLIST.md](./LAUNCH_CHECKLIST.md) | Applying to programmes, payout thresholds, and the Bulgarian tax position |
| [MERCHANT_INTEGRATION.md](./MERCHANT_INTEGRATION.md) | How merchants, offers, links, and conversions are modelled |
| [email-marketing-pricing-research.md](./email-marketing-pricing-research.md) | Why the email-marketing category was researched and not published |
| [../BUSINESS_MODEL.md](../BUSINESS_MODEL.md) | Why commission is structurally excluded from ranking |

## Growth and content

| Document | What it settles |
| --- | --- |
| [SOCIAL_MEDIA_GROWTH_STRATEGY_BG.md](./SOCIAL_MEDIA_GROWTH_STRATEGY_BG.md) | The distribution strategy (Bulgarian) |
| [FACELESS_VIDEO_PLAYBOOK_BG.md](./FACELESS_VIDEO_PLAYBOOK_BG.md) | Full faceless-video production playbook (Bulgarian) |
| [FACELESS_VIDEO_QUICK_GUIDE_BG.md](./FACELESS_VIDEO_QUICK_GUIDE_BG.md) | The same, compressed to what you do today (Bulgarian) |
| [ROUTING_SEO.md](./ROUTING_SEO.md) · [ROUTING_SEO_AUDIT.md](./ROUTING_SEO_AUDIT.md) | Which routes are indexable, and the evidence they behave |
| [ANALYTICS.md](./ANALYTICS.md) | What is measured, and what deliberately is not |

## Running it in production

| Document | What it settles |
| --- | --- |
| [DEPLOY_SINGLE_BOX.md](./DEPLOY_SINGLE_BOX.md) | **The deployment actually in use:** one ~€5 VPS, no managed services |
| [DEPLOYMENT.md](./DEPLOYMENT.md) | The generic container deployment procedure |
| [PRODUCTION_CONFIGURATION.md](./PRODUCTION_CONFIGURATION.md) | Settings production refuses to start without |
| [server-swap.md](./server-swap.md) | Why the box has swap, after ClamAV was OOM-killed twice |
| [OPERATIONS.md](./OPERATIONS.md) | Day-to-day operational procedures |
| [OBSERVABILITY.md](./OBSERVABILITY.md) | Metrics, logs, and what alerting exists |
| [PERFORMANCE_BUDGETS.md](./PERFORMANCE_BUDGETS.md) | The budgets CI enforces on every build |
| [LOAD_TESTING.md](./LOAD_TESTING.md) | Load-test method and results |
| [EMAIL_PRODUCTION.md](./EMAIL_PRODUCTION.md) | The transactional SMTP boundary |
| [MEDIA_RECONCILIATION.md](./MEDIA_RECONCILIATION.md) | Keeping stored media and database rows in agreement |
| [STAGING_PRODUCTION_PARITY.md](./STAGING_PRODUCTION_PARITY.md) | Where staging and production are allowed to differ |
| [PROVIDER_ACTIVATION_CHECKLIST.md](./PROVIDER_ACTIVATION_CHECKLIST.md) | What must be true before any external provider is enabled |

## Data, safety, and recovery

| Document | What it settles |
| --- | --- |
| [MIGRATION_SAFETY.md](./MIGRATION_SAFETY.md) | Rules for forward-only migrations |
| [BACKUP_RESTORE.md](./BACKUP_RESTORE.md) | Backup schedule and the restore procedure |
| [DISASTER_RECOVERY.md](./DISASTER_RECOVERY.md) · [DR_READINESS.md](./DR_READINESS.md) · [PRODUCTION_DR_EXERCISE.md](./PRODUCTION_DR_EXERCISE.md) | The DR plan, its readiness state, and the exercise that tested it |
| [INCIDENT_RESPONSE.md](./INCIDENT_RESPONSE.md) | What to do when it breaks |
| [DATA_GOVERNANCE.md](./DATA_GOVERNANCE.md) · [DATA_RETENTION.md](./DATA_RETENTION.md) | What is stored about people, and for how long |

## Security

| Document | What it settles |
| --- | --- |
| [ACCOUNT_SECURITY.md](./ACCOUNT_SECURITY.md) | Sessions, MFA, and account-recovery behaviour |
| [ABUSE_PROTECTION.md](./ABUSE_PROTECTION.md) | Rate limiting and abuse control |
| [CI_SECURITY.md](./CI_SECURITY.md) | Supply-chain and CI hardening |
| [SECURITY_VALIDATION.md](./SECURITY_VALIDATION.md) | Evidence the controls above were exercised |

## Launch history and evidence

These record decisions already made. They are history, not instructions.

| Document | What it records |
| --- | --- |
| [LAUNCH_GOVERNANCE.md](./LAUNCH_GOVERNANCE.md) | Who approves a launch, and against what |
| [PRE_LAUNCH_SCORECARD.md](./PRE_LAUNCH_SCORECARD.md) | The pre-launch state assessment |
| [PRODUCTION_VALIDATION.md](./PRODUCTION_VALIDATION.md) | Validation performed against production |
| [PHASE_11_EVIDENCE.md](./PHASE_11_EVIDENCE.md) · [PHASE_12_EVIDENCE.md](./PHASE_12_EVIDENCE.md) | Completion evidence for phases 11 and 12 |
| [architecture.md](./architecture.md) | Superseded stub; see [../ARCHITECTURE.md](../ARCHITECTURE.md) |

## Root-level documents

| Document | What it settles |
| --- | --- |
| [../README.md](../README.md) | Setup, configuration, and repository layout |
| [../ARCHITECTURE.md](../ARCHITECTURE.md) | The architectural source of truth |
| [../API.md](../API.md) | The HTTP contract |
| [../BUSINESS_MODEL.md](../BUSINESS_MODEL.md) | Monetization, guardrails, and funnel metrics |
| [../AGENTS.md](../AGENTS.md) | Working rules for automated contributors |
| [../PRODUCTION_READINESS.md](../PRODUCTION_READINESS.md) | The readiness checklist and known limitations |
| [../PHASE_STATUS.md](../PHASE_STATUS.md) · [../FINAL_AUDIT.md](../FINAL_AUDIT.md) | Phase-by-phase status and the closing audit |
| [../PHASE_13_INFRASTRUCTURE_PLAN.md](../PHASE_13_INFRASTRUCTURE_PLAN.md) and its `PHASE_13_*` siblings | A managed-infrastructure plan written for a budget that does not exist. [DEPLOY_SINGLE_BOX.md](./DEPLOY_SINGLE_BOX.md) is what is actually deployed. |
