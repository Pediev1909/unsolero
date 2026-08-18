# UNSOLERO abuse protection

Status: PARTIAL — distributed Redis-compatible adapter implemented; production service/topology validation blocked

Protected routes use a provider-neutral limiter contract. The local adapter is
bounded to 10,000 active keys and is appropriate only for development,
controlled staging, or an explicitly single API replica. Client addresses are
HMAC-SHA-256 pseudonymized before they reach the limiter; the production key is
provided by `RATE_LIMIT_KEY_SECRET` and must be stable across restarts.

`API_REPLICA_COUNT>1` with `RATE_LIMIT_PROVIDER=local` is rejected by central
configuration validation. `RATE_LIMIT_PROVIDER=redis` selects the implemented
atomic adapter; production requires a `rediss://` URL. Selecting `external`
still fails startup until a reviewed adapter is linked. A backend error fails closed with 503
for the protected operation, increments a failure metric, and attempts an
operational alert. It never silently allows the request.

Before horizontal scaling in a real environment:

1. provision an authenticated/TLS Redis-compatible service and private network;
2. validate persistence, eviction, failover, timeout and connection policies;
3. configure trusted proxy boundaries and prevent direct API access;
4. use a stable secret-managed HMAC key;
5. load-test bucket semantics, backend outage behavior, and failover;
6. add dashboards/alerts for rejected requests, backend failures, and key
   cardinality;
7. document whether edge and application limits are cumulative.

## Distributed adapter acceptance contract

The Redis-compatible adapter implements one Lua-scripted atomic operation
equivalent to `Allow(context, opaqueKey, limit, window)` and a bounded readiness
probe. Keys received by the adapter are already fixed-domain bucket names plus
a 64-character HMAC digest; the adapter must reject empty/oversized keys and
must never persist raw IPs, emails, usernames, tokens, routes, or request data.
It must apply an expiry on first increment, return nonnegative remaining and a
bounded retry duration, and fail closed on timeout, authentication, TLS,
protocol, cluster, or quorum errors.

Acceptance tests must run against the exact production-compatible topology and
cover concurrent first requests, the same key through two logical API replicas,
independent endpoint buckets (login, registration/reset/MFA, recommendation,
analytics, affiliate and generic mutation), expiry, clock behavior, backend
restart/failover, network partition, credential rotation, connection-pool
exhaustion and bounded cardinality. Direct client control over bucket/key is
prohibited.

Phase 10 added `go-redis` and exercised the adapter against isolated Redis with
two logical replicas, 40 concurrent requests, exact limit enforcement, TTL
expiry, route separation, duplicate requests and outage failure. CI provisions
the same isolated dependency. This proves the repository contract, not managed
service TLS/auth, clustering, failover, eviction, capacity, network policy, or
operator readiness; those remain blocker classes B/C.

The current fixed-window policy is an abuse control, not a DDoS service. A WAF,
network-layer protection, credential-stuffing detection, and distributed source
controls remain deployment requirements.
