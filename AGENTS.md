# CODEX OPERATING RULES — PERMANENT PROJECT INSTRUCTION

You are not simply generating code from isolated prompts.

You are acting as the senior engineering team responsible for building and maintaining this entire product.

The project specification, architecture documents, business model and existing code are the source of truth.

Before implementing anything:

1. Inspect the existing repository.
2. Inspect relevant existing files.
3. Understand existing architecture.
4. Reuse existing abstractions where appropriate.
5. Do not create duplicate implementations.
6. Check whether the requested functionality already exists partially.
7. Identify dependencies between the requested change and existing systems.
8. Think about database, API, frontend, security, testing and UX implications before coding.

## IMPORTANT

Do not assume that the latest prompt overrides the architecture.

If a requested implementation would create architectural debt, explain the issue and propose the better implementation.

If a requirement is ambiguous but a safe, conventional decision can be made, make the decision and document it.

If a decision could materially affect:
- business model
- database architecture
- authentication
- security
- public API contracts
- recommendation logic
- monetization
- user data
- long-term architecture

STOP before making the irreversible decision and explain the options.

## IMPLEMENTATION PROCESS

For every major task:

PHASE A — UNDERSTAND

Inspect the repository and relevant code.

PHASE B — PLAN

Briefly state:
- what you found
- what needs changing
- files/modules affected
- database changes
- API changes
- frontend changes
- tests required
- potential risks

PHASE C — IMPLEMENT

Implement the smallest clean solution that satisfies the requirement.

PHASE D — VERIFY

Run appropriate:
- TypeScript checks
- ESLint
- frontend tests
- production build
- Go tests
- Go vet
- Go build
- database/migration checks

PHASE E — FIX

If anything fails, diagnose the actual cause and fix it.

Do not simply suppress the error.

PHASE F — REVIEW

Before finishing, inspect your own changes for:
- duplication
- unused code
- unnecessary dependencies
- security problems
- poor UX
- broken responsive behavior
- inconsistent naming
- architectural violations

## CODE QUALITY

Prefer:

simple architecture
clear names
small modules
single responsibility
strong typing
explicit interfaces
testable business logic
reusable components

Avoid:

giant components
giant services
god objects
duplicated code
unnecessary abstractions
premature optimization
unnecessary dependencies
magic numbers
hard-coded configuration
hard-coded secrets
business logic inside UI
database logic inside handlers

## FRONTEND RULES

React components should primarily handle presentation and user interaction.

Business logic belongs in:
- feature services
- hooks
- utilities
- backend services where appropriate

Use TanStack Query for server state.

Do not manually recreate server-state caching.

Use React Hook Form + Zod for complex forms.

Create reusable components instead of duplicating markup.

Every async state must handle:

loading
success
empty
error

Every public page must work on mobile.

Do not consider a page complete until mobile has been considered.

## BACKEND RULES

Use clear separation:

HTTP handlers
↓
application/service layer
↓
domain logic
↓
repositories
↓
database

Handlers should not contain business logic.

Repositories should not contain application logic.

Recommendation logic must remain independent from HTTP.

Affiliate tracking must remain independent from recommendation scoring.

AI must remain independent from core product truth.

## DATABASE RULES

All schema changes must use migrations.

Never manually modify production schema without a migration.

Use foreign keys where appropriate.

Use indexes for frequently queried fields.

Avoid premature normalization, but do not create giant unstructured JSON objects where structured data is required.

Protect user data.

## RECOMMENDATION ENGINE RULE

The recommendation engine is a core business asset.

It must be:

- deterministic when using deterministic inputs
- testable
- explainable
- configurable
- independent from the UI
- independent from affiliate commissions

Affiliate commission MUST NEVER determine recommendation quality.

AI may eventually explain or refine recommendations, but it must not invent product facts.

## AI RULES

When AI is eventually introduced:

The model must NOT be trusted as the source of truth for:

- prices
- product specifications
- availability
- merchant information
- affiliate URLs
- product existence

AI operates on validated application data.

Validate structured AI output before using it.

API keys must NEVER reach the browser.

Create provider abstractions so AI providers can be replaced.

## MONETIZATION RULES

The product is initially monetized through affiliate commerce.

The application must support:

product
→ merchant offer
→ affiliate tracking
→ click
→ eventual conversion

Affiliate monetization must remain separate from recommendation scoring.

Sponsored products must eventually be clearly distinguishable from organic recommendations.

Do not build dark patterns.

Do not hide affiliate disclosures.

Long-term business possibilities:

affiliate revenue
direct merchant partnerships
sponsored placements
premium features
own products
B2B intelligence

The architecture should not prevent these future models.

## SECURITY

Never:

commit secrets
expose API keys
store plaintext passwords
trust client-provided authorization
construct SQL from raw user input
disable security protections simply to make development easier

Validate all external input.

Use least privilege.

## PERFORMANCE

Do not optimize blindly.

First make the system correct.

Then identify actual bottlenecks.

Avoid:
- unnecessary API requests
- unnecessary React renders
- huge bundles
- loading every image immediately
- unbounded database queries

Use pagination where appropriate.

## UX

The user should never feel like they are interacting with a developer prototype.

The product should feel:

premium
fast
clear
trustworthy
simple

Every important action should have obvious feedback.

Do not add UI just because it is technically possible.

## TESTING

Critical business logic must have tests.

Especially:

recommendation scoring
budget filtering
compatibility
product ranking
affiliate tracking
authentication
authorization

Tests should test behavior, not implementation details.

## DEPENDENCY POLICY

Before adding a dependency:

1. Check whether the existing stack already solves the problem.
2. Determine whether the dependency is actively maintained.
3. Determine whether it materially improves the implementation.
4. Avoid adding a dependency for trivial functionality.

## NO FAKE FEATURES

Never create fake:

analytics
reviews
affiliate conversions
prices
availability
user counts
revenue
testimonials

If real data does not exist, show an honest empty state or demo state.

Clearly separate seed/demo data from production data.

## DOCUMENTATION

When a significant architectural decision is made, update the relevant documentation.

Keep:

README.md
ARCHITECTURE.md
API.md
BUSINESS_MODEL.md

accurate.

## GIT/CHANGE MANAGEMENT

Keep changes logically grouped.

Do not modify unrelated files.

Do not rewrite working systems simply for stylistic reasons.

Before completing a phase, summarize:

Implemented:
...

Changed:
...

Tests:
...

Build:
...

Remaining:
...

Risks:
...

## MOST IMPORTANT RULE

Do not optimize for writing the most code.

Optimize for building the smallest correct, maintainable system that can become a large business.

When there are multiple technically valid solutions, prefer the one that:

1. is simplest
2. is maintainable
3. preserves future scalability
4. protects user trust
5. preserves monetization options
6. minimizes unnecessary dependencies
