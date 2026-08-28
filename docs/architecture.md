# Architecture — superseded

The architectural source of truth is [../ARCHITECTURE.md](../ARCHITECTURE.md).

This file held the Phase 1 sketch: planned module boundaries, a three-service
topology, and an intended `/api/v1` contract. It is kept as a stub rather than
expanded, because keeping two architecture documents means one of them is
quietly wrong. Two of its statements had already drifted from the code —
`/api/v1` ended up carrying only health, metrics, and the public-route probe
while catalog and commerce sit under `/api/`, and the domain modules it listed
as "planned, not commitments" are implemented.

Read [../ARCHITECTURE.md](../ARCHITECTURE.md) instead.
