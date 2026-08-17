# Database migrations

Migrations are immutable, forward-only SQL files named `NNNNNN_description.sql`. The migration command verifies the checksum of every previously applied file and refuses to continue if history was modified.

Run from `backend/` with `go run ./cmd/migrate`, or use the one-shot `migrate` service in Docker Compose.

`000003_create_authentication_sessions.sql` adds the identity-owned opaque session store. Token hashes are fixed at 32 bytes, and absolute, idle, and revocation timestamps are constrained in the database.

`000005_recommendation_experience.sql` adds structured authenticated drafts, existing-equipment child rows, immutable recommendation score dimensions and explanation rows, and alternative-to-selected-product links used by the primary builder and saved-setup experience.

`000006_persisted_comparisons.sql` adds the authenticated user's ordered zero-to-four product comparison selection. Foreign keys enforce ownership and product integrity; application and repository layers restrict entries to published products.

`000007_affiliate_commerce_analytics.sql` adds offer activation, provider-neutral program and commission metadata, and attributed affiliate clicks. It evolves the earlier outbound-click table without losing existing rows, adds authenticated and anonymous actor fields, and indexes offer/user/session reporting paths. Commission columns are internal commerce data and are prohibited from recommendation scoring. Product analytics continues to use the allowlisted `analytics.events` envelope created by the core schema.

`000008_admin_dashboard.sql` adds normalized role membership, seeds the non-self-assigning `admin` role definition, and creates the actor-attributed `admin.audit_log`. It also permits controlled same-origin product media URLs while preserving the existing HTTPS and demo-asset constraints. The migration does not grant administrative access to any user.

`000009_analytics_foundation.sql` canonicalizes first-party event names, adds structured and privacy-bounded traffic and recommendation attribution, and indexes product/onboarding reporting paths. It also creates empty provider-reported affiliate conversion and acquisition-cost models for future verified commission, revenue, EPC, and CAC reporting; it does not create conversion, spend, or revenue data.

`000010_editorial_content.sql` adds reviewed editorial authors and entries plus normalized product, category, and related-entry links. Public routes expose published entries only; structured content blocks are validated by the content domain before rendering.
