# UNSOLERO load and scale validation

Status: reproducible local baseline; not a production capacity claim

## Test environment

The 2026-08-17 run used an Intel Core i7-12650H (10 physical cores, 16 logical
CPUs), 15 GiB RAM, Linux 7.1.7, Docker Compose, PostgreSQL 17 Alpine, and the
development API image over loopback. The host had only 4.1 GiB available RAM
and its 8 GiB swap was effectively full. Containers had no CPU or memory limits.
TLS, a production ingress, multiple API replicas, network latency, and a
distributed rate limiter were not present. Results are useful for regression
comparison only; they are not an SLA or safe traffic limit.

Rate limits were raised to 100,000/minute for load isolation. Real configured
limits must be tested separately. The post-run snapshot—not a peak sampler—was:

| Container | CPU | Memory | PIDs |
| --- | ---: | ---: | ---: |
| API | 0.00% | 204.8 MiB | 22 |
| PostgreSQL | 0.14% | 197.3 MiB | 16 |
| commerce worker | 0.00% | 13.09 MiB | 12 |
| Vite development server | 0.00% | 219.3 MiB | 53 |

## HTTP baseline

Every scenario completed with zero unexpected HTTP or transport errors. A
redirect was treated as success without following its merchant destination;
401 was the expected result for invalid-login and unauthorized-admin probes.

| Scenario | Requests / concurrency | RPS | p50 | p95 | p99 | Maximum |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| readiness | 2,000 / 32 | 13,687.48 | 1.99 ms | 4.63 ms | 7.68 ms | 20.84 ms |
| public catalog, 24 products | 1,000 / 16 | 1,257.71 | 12.08 ms | 16.35 ms | 29.18 ms | 39.32 ms |
| anonymous recommendation | 200 / 4 | 367.32 | 10.37 ms | 13.92 ms | 21.05 ms | 22.99 ms |
| registration with Argon2id | 8 / 2 | 13.76 | 120.87 ms | 207.62 ms | 207.62 ms | 207.62 ms |
| invalid login with dummy hash | 20 / 4 | 19.90 | 165.89 ms | 282.04 ms | 284.50 ms | 284.50 ms |
| consented analytics ingest | 200 / 8 | 277.07 | 21.65 ms | 75.82 ms | 98.49 ms | 107.75 ms |
| unauthorized admin denial | 500 / 16 | 3,741.87 | 3.39 ms | 9.74 ms | 12.16 ms | 16.72 ms |
| authenticated admin dashboard | 200 / 4 | 547.22 | 7.71 ms | 10.67 ms | 11.94 ms | 16.75 ms |
| authenticated commerce providers | 200 / 4 | 554.90 | 7.48 ms | 10.00 ms | 10.65 ms | 11.42 ms |
| authenticated persisted recommendation | 20 / 2 | 12.01 | 83.29 ms | 334.43 ms | 337.75 ms | 337.75 ms |
| affiliate click and redirect | 200 / 8 | 576.98 | 11.61 ms | 20.82 ms | 56.47 ms | 73.95 ms |

The registration and login samples are deliberately small because each request
performs the production-strength memory-hard password path. They validate
bounded behavior, not saturation capacity.

After the run PostgreSQL reported 11 application/test connections of 100
server slots, 14,689 committed and 104 rolled-back transactions, 32,282 block
reads, 22,513,385 block hits, one 4.12 MiB temporary file, and zero deadlocks.
The queue snapshot contained two successful conversion imports, two successful
offer imports, and expected disabled-adapter/expired-lease test records. Live
providers remain blocked without credentials and adapters.

## Synthetic database scale

[`phase8-scale-validation.sql`](../scripts/phase8-scale-validation.sql) refuses
non-disposable database names, labels all data fictional, wraps the entire run
in one transaction, and rolls it back. It generated:

| Relation scenario | Rows |
| --- | ---: |
| users | 10,000 |
| sessions | 20,000 |
| governed products with evidence | 5,000 |
| offers | 10,000 |
| clicks | 50,000 |
| verified conversion projections | 10,000 |
| analytics events | 100,000 |
| security events | 50,000 |
| recommendation snapshots | 10,000 |
| admin audit rows | 50,000 |

Representative `EXPLAIN (ANALYZE, BUFFERS)` execution times after `ANALYZE`
were: indexed email lookup 0.273 ms; token-hash session lookup 0.253 ms;
offset-2,400 public catalog page 6.859 ms; 5,030 governed recommendation
candidates 23.272 ms; 20,000-event product ranking 17.244 ms; 50,202-click
affiliate aggregation 28.762 ms; 10,006-row verified conversion aggregation
7.235 ms; and indexed retention selection 0.047 ms.

The first run exposed an actual mismatch: email queries did not match the
existing `lower(email)` unique index and scanned all users. Repository queries
now use the indexed expression; the rerun proved an `users_email_unique` index
scan. Catalog count-plus-offset pagination and broad aggregate scans remain
acceptable at this local fixture size but need production-volume thresholds,
keyset pagination, and/or rollups before materially larger traffic.

## Reproduction

Build and use the dependency-free Go probe:

```bash
cd backend
go run ./cmd/loadtest -scenario public-catalog \
  -url 'http://127.0.0.1:8080/api/catalog/products?limit=24' \
  -requests 1000 -concurrency 16
```

Request bodies live in `scripts/load/`. `{{SEQ}}` and `{{UUID}}` are replaced
per request. A setup request can establish a cookie jar per worker:

```bash
go run ./cmd/loadtest -scenario consented-analytics \
  -url http://127.0.0.1:8080/api/analytics/events -method POST \
  -body-file ../scripts/load/analytics-page-view.json \
  -setup-url http://127.0.0.1:8080/api/analytics/consent \
  -setup-method PUT -setup-body-file ../scripts/load/analytics-consent.json \
  -setup-success 200 -success 204 -requests 200 -concurrency 8
```

Run the SQL fixture only against a disposable database:

```bash
psql "$TEST_DATABASE_URL" -f scripts/phase8-scale-validation.sql
```

## Unvalidated capacity

Sustained soak, peak CPU/memory, network and disk throttling, multi-replica
contention, production TLS/ingress, distributed rate limiting, provider webhook
bursts, millions of governed recommendation candidates, and production-sized
financial/audit retention remain not tested. Establish service-level objectives
and rerun on production-equivalent infrastructure before launch.
