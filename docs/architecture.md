# Architecture

## Product boundary

UNSOLERO is a product decision engine. The system must keep objective product suitability separate from commercial monetization. Affiliate commission, sponsorship, and merchant relationships are never inputs to the objective recommendation score.

## Initial topology

- `frontend`: React single-page application. Pages compose reusable UI; server state belongs in TanStack Query; forms use React Hook Form and Zod.
- `backend`: Go REST API with explicit transport, application, and infrastructure boundaries.
- `postgres`: the future source of truth for structured products, evidence, compatibility, prices, and user-owned equipment.

The frontend depends on versioned REST contracts under `/api/v1`. Domain services must not depend on HTTP or PostgreSQL implementations.

## Planned domain modules

These are boundaries, not Phase 1 implementation commitments:

- Catalog: normalized products, variants, specifications, categories, and evidence provenance.
- Fit: goals, experience, space, budget, constraints, and owned equipment.
- Compatibility: physical, functional, and ecosystem relationships.
- Recommendations: deterministic eligibility, scoring, explanations, rejected candidates, and upgrade paths.
- Commerce: merchants, price observations, affiliate destinations, and sponsored-placement metadata.
- Identity: accounts, saved spaces, and recommendation histories.

## Recommendation integrity

The recommendation pipeline will be layered:

1. Validate and normalize user constraints.
2. Eliminate ineligible or incompatible products.
3. Score eligible products using versioned, explainable criteria.
4. Assemble a budget-constrained set and detect redundancy.
5. Generate evidence-linked reasons, trade-offs, alternatives, and rejections.
6. Attach commerce destinations only after ranking is complete.

Every result should retain the scoring-policy version and source evidence needed to reproduce it. AI may later help interpret intent or explain results, but it must not bypass hard constraints or invent product facts.

## API conventions

- JSON request and response bodies.
- Versioned paths under `/api/v1`.
- Input validation at the transport boundary and domain invariant enforcement in application services.
- Stable machine-readable error codes with safe user-facing messages.
- Liveness reports process health; readiness verifies required dependencies.

## Data and migrations

Schema changes are forward SQL migrations in `backend/migrations`. Queries belong in repositories, never HTTP handlers. The core identity, planning, catalog, commerce, recommendation, and analytics tables are now defined; product scoring inputs use structured columns and typed attributes. Recommendation drafts are stored separately from immutable completed runs, while completed score dimensions and explanation rows remain queryable rather than being hidden in a result JSON blob.

## Security baseline

- Secrets arrive through environment variables and are never committed.
- Production authentication will use secure, short-lived sessions or JWTs with explicit threat modeling before implementation.
- Logs must not contain credentials, tokens, or sensitive user profile data.
- Sponsored placements must be visibly labeled and remain separate from objective ranking.
