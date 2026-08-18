# Performance budgets

Status: PARTIAL — automated local/staging-shaped regression gates pass; field and production capacity evidence is unavailable  
Last reviewed: 2026-08-18

Budgets are release gates, not aspirational averages. Any exception requires an owner, measured user benefit, expiry date, and rollback plan.

## Frontend budgets

| Metric | Initial budget | Current local evidence | Status |
| --- | ---: | ---: | --- |
| main application JavaScript | ≤ 100 KiB gzip | 86,794 bytes (84.76 KiB) gzip | PASS |
| global CSS | ≤ 15 KiB gzip | 11.69 KiB gzip | PASS |
| any lazy route chunk | ≤ 75 KiB gzip | every emitted lazy chunk passed | PASS |
| initial JS + CSS | ≤ 300 KiB gzip | 151,319 bytes | PASS |
| LCP p75 mobile | ≤ 2.5 s | no field/staging RUM | NOT TESTED |
| INP p75 | ≤ 200 ms | no field/staging RUM | NOT TESTED |
| CLS p75 | ≤ 0.1 | no field/staging RUM | NOT TESTED |

Public hero media should use correct dimensions, modern formats, responsive sources, and an editorially approved size budget. Below-fold product/content media remains lazy. A build-size pass does not prove Core Web Vitals.

## API and data budgets

| Surface | Initial staging target | Constraint |
| --- | ---: | --- |
| health/live | p95 ≤ 100 ms | no downstream work |
| health/ready | p95 ≤ 500 ms | bounded database/storage/limiter checks |
| catalog list/detail | p95 ≤ 500 ms | bounded page and query timeout |
| authentication mutations | p95 ≤ 1.5 s | Argon2 cost is security-sensitive; never lower without review |
| deterministic recommendation | p95 ≤ 2 s | bounded candidate set, no provider dependency |
| account/admin lists | p95 ≤ 750 ms | page size ≤ 100, stable indexed ordering |
| affiliate redirect | p95 ≤ 500 ms | durable click decision; stale/invalid offer fails closed |

Targets must be recalibrated with representative staging data and expected concurrency. Required evidence includes p50/p95/p99, error/timeout rate, database pool wait, CPU/memory, query plans, cache behavior, object-store latency, Redis latency, and sustained soak results. No capacity claim is made from local execution.

### Phase 11 TLS staging-shaped snapshot

The disposable local topology used TLS ingress, two API replicas, two workers,
shared PostgreSQL/Redis/MinIO, secure cookies, private S3 media configuration,
and resource limits. Data remained fictional and small. These are regression
measurements, not capacity evidence:

| Scenario | Requests / concurrency | Error rate | p50 | p95 | p99 |
| --- | ---: | ---: | ---: | ---: | ---: |
| readiness | 300 / 16 | 0% | 85.58 ms | 190.27 ms | 306.18 ms |
| catalog | 200 / 12 | 0% | 94.47 ms | 280.93 ms | 298.31 ms |
| recommendation | 12 / 2 | 0% | 12.45 ms | 36.10 ms | 36.10 ms |
| invalid login/dummy Argon2 | 8 / 2 | 0% | 213.78 ms | 265.75 ms | 265.75 ms |
| consented analytics | 40 / 4 | 0% | 6.44 ms | 12.33 ms | 17.20 ms |
| admin authorization boundary | 40 / 4 | 0% | 1.40 ms | 4.55 ms | 6.07 ms |
| actual unknown-route 404 | 100 / 8 | 0% | 4.42 ms | 97.80 ms | 111.70 ms |
| affiliate redirect/no follow | 40 / 4 | 0% | 3.78 ms | 13.87 ms | 16.06 ms |
| 30-second catalog soak | 3,216 / 8 | 0% | 90.07 ms | 183.67 ms | 195.60 ms |

The first soak exposed a harness defect: its own duration deadline could count
canceled in-flight work as transport failure. The probe now excludes only its
own deadline cancellation, with a regression test; real failures remain
failures. Critical slug/session/wishlist/media-deletion lookup plans must
contain an index node in PostgreSQL integration tests.

### Historical Phase 10 local regression snapshot

The disposable Docker stack had no TLS ingress, resource limits, production
network latency, or representative user traffic. A 500-request readiness probe
at concurrency 16 returned 500/500 HTTP 200 with p50 0.822 ms, p95 1.974 ms,
p99 11.152 ms, and zero transport errors. A 300-request catalog probe at
concurrency 12 returned 300/300 HTTP 200 with p50 4.584 ms, p95 8.635 ms,
p99 35.789 ms, and zero transport errors. These are regression observations,
not capacity or SLO evidence.

The rollback-only PostgreSQL fixture inserted 10,000 users, 20,000 sessions,
5,000 governed products, 10,000 offers, 50,000 clicks, 10,000 conversion
projections, 10,000 recommendation snapshots, 100,000 analytics events, 50,000
security events, and 50,000 audit rows. Current representative execution times
included indexed email 0.035 ms, session resolution 0.045 ms, offset-2,400
catalog 3.963 ms, 5,030 recommendation candidates 12.939 ms, analytics ranking
9.651 ms, affiliate aggregation 14.689 ms, conversion aggregation 3.553 ms,
and retention selection 0.043 ms. The transaction rolled back all fixture data.

## Release enforcement

CI fails frontend build/type/lint/test and deterministic gzip artifact-budget
regressions. `scripts/run-staging-performance-gates.sh` enforces local HTTP
ceilings. Lighthouse/Core Web Vitals lab automation, RUM, authenticated admin
pagination with representative data, longer soak, and saturation tests remain
incomplete. Do not collect user-identifying metric labels.
