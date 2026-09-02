# Database migrations

Migrations are immutable, forward-only SQL files named `NNNNNN_description.sql`. The migration command verifies the checksum of every previously applied file and refuses to continue if history was modified.

Run from `backend/` with `go run ./cmd/migrate`, or use the one-shot `migrate` service in Docker Compose.

The API binary embeds the same immutable SQL files and compares their ordered
names and checksums with `platform.schema_migrations` during readiness checks.
This is a compatibility check only: the API does not apply migrations. A
database that is older, newer, renamed, or checksum-divergent from the release
stays unready until an operator runs the explicit migration/recovery process.

`000003_create_authentication_sessions.sql` adds the identity-owned opaque session store. Token hashes are fixed at 32 bytes, and absolute, idle, and revocation timestamps are constrained in the database.

`000005_recommendation_experience.sql` adds structured authenticated drafts, existing-equipment child rows, immutable recommendation score dimensions and explanation rows, and alternative-to-selected-product links used by the primary builder and saved-setup experience.

`000006_persisted_comparisons.sql` adds the authenticated user's ordered zero-to-four product comparison selection. Foreign keys enforce ownership and product integrity; application and repository layers restrict entries to published products.

`000007_affiliate_commerce_analytics.sql` adds offer activation, provider-neutral program and commission metadata, and attributed affiliate clicks. It evolves the earlier outbound-click table without losing existing rows, adds authenticated and anonymous actor fields, and indexes offer/user/session reporting paths. Commission columns are internal commerce data and are prohibited from recommendation scoring. Product analytics continues to use the allowlisted `analytics.events` envelope created by the core schema.

`000008_admin_dashboard.sql` adds normalized role membership, seeds the non-self-assigning `admin` role definition, and creates the actor-attributed `admin.audit_log`. It also permits controlled same-origin product media URLs while preserving the existing HTTPS and demo-asset constraints. The migration does not grant administrative access to any user.

`000009_analytics_foundation.sql` canonicalizes first-party event names, adds structured and privacy-bounded traffic and recommendation attribution, and indexes product/onboarding reporting paths. It also creates empty provider-reported affiliate conversion and acquisition-cost models for future verified commission, revenue, EPC, and CAC reporting; it does not create conversion, spend, or revenue data.

`000010_editorial_content.sql` adds reviewed editorial authors and entries plus normalized product, category, and related-entry links. Public routes expose published entries only; structured content blocks are validated by the content domain before rendering.

`000014_commerce_operations.sql` adds provider configurations, scheduled and manual offer imports, immutable price/availability observations, import failures, reconciliation-safe external mappings, durable click classification/idempotency fields, and operator audit history. Live providers remain disabled without a reviewed adapter and credential reference.

`000015_verified_conversions.sql` adds authenticated conversion delivery/import history, immutable provider-event facts, normalized order and commission lifecycle projections, optional evidence-based click attribution, reconciliation history, provider conversion health, and reporting indexes. It does not seed conversions, commissions, orders, or revenue.

`000016_complete_account_security.sql` adds hashed email-verification and password-reset credentials, encrypted TOTP credential storage, hashed one-use recovery codes, bounded MFA login challenges, session MFA metadata, least-privilege staff roles, and append-only security events. Security-event session identifiers deliberately outlive cleaned sessions and contain no token material.

`000017_analytics_privacy_governance.sql` adds server-held versioned consent,
opaque browser subjects and guarded identity claims, unique public event IDs,
payload-free ingestion outcomes, reportability/traffic classification,
reporting coverage, indexed expiry, and an explicit data-class policy registry.
Historical event rows are non-reportable because current server consent cannot
be proven. Checked-in retention values are engineering defaults requiring
privacy/legal review.

`000018_phase10_collection_indexes.sql` adds stable user-collection indexes and
durable media-deletion queue indexes for bounded page and worker paths.

`000019_phase11_media_reconciliation.sql` adds leased reconciliation runs,
privacy-bounded discrepancy results, and provider-neutral operational
checkpoints for backup/restore/fingerprint telemetry. Unexpected or unvalidated
provider keys are retained only as SHA-256 identifiers. The migration creates
no provider, customer, conversion, revenue, or commission data.

`000020_worker_operational_checkpoint.sql` extends the bounded checkpoint key
set with a durable worker heartbeat. It carries only success time, failure
count, and a fixed detail code so cross-process monitoring can distinguish a
live container from a worker that is not completing cycles.

`000025_optional_recommendation_dimensions.sql` aligns immutable recommendation
history with non-physical catalogs. Candidate dimensions and session space are
stored as `NULL` when they do not apply, while complete positive dimensions
remain mandatory whenever physical measurements are present.

`000026_affiliate_promotions.sql` adds privacy-bounded affiliate promotions and
click records for editorial offers such as free training. Promotions have no
catalog product or recommendation relationship, so a webinar, book, or bundle
cannot enter software scoring merely because it has an affiliate destination.

`000027_newsletter_subscriptions.sql` creates the `audience` schema and the
double opt-in `newsletter_subscriptions` table. A row stays `pending` until the
hashed one-time confirmation token from the email is consumed, and
state-consistency checks keep status, tokens, and timestamps in step. Only the
lower-cased address, its consent state, the consent text version, the
requesting surface, and SHA-256 token hashes are stored: no IP address, user
agent, or account link. Expired pending rows are eligible for deletion;
unsubscribed rows remain as suppression records.

`000028_editorial_stack_type.sql` widens the editorial content-type check to
admit `stack`: a whole set of tools for one kind of business and budget,
published at `/stacks/{slug}`. It changes no rows and adds no columns; the
content domain, the public route resolver and the sitemap learn the type in the
same release.
