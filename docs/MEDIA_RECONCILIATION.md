# Media reconciliation

Status: repository and isolated-S3 validation complete; managed-provider inventory is not validated  
Last reviewed: 2026-08-18

Product media uses private, product-scoped object keys:

```text
products/<product UUID>/<lowercase SHA-256>.<jpg|png|webp>
```

Anything outside that grammar is unexpected and is never automatically deleted.
The database stores the public application URL; object storage remains private.

## Crash windows

| Window | Possible state | Reconciliation classification |
| --- | --- | --- |
| object created before database registration | unreferenced object | `orphan_object` |
| database registration completed but object unavailable | reference without object | `missing_object` |
| replacement created before old deletion is queued | young unreferenced object | quarantined by safety grace |
| deletion queued or processing but interrupted | object plus stale job | `stale_deletion` |
| deletion completed in storage but database reference remains | reference without object | `missing_object` |
| duplicate/cross-product database references | ambiguous ownership | `duplicate_reference` / `ownership_mismatch` |
| provider object outside the product namespace | unsafe to classify | `unexpected_namespace` |

## Safety invariants

- Dry-run is the default. Apply mode requires `--apply-safe-orphans`.
- Apply mode only enqueues deletion; the existing bounded worker performs it.
- Only grammar-valid, product-scoped, unreferenced objects older than the
  configured grace period can be enqueued.
- Unknown timestamps, young objects, unexpected keys, ownership conflicts, and
  unclassifiable records are reported but never deleted.
- Existing fresh pending/processing jobs are left alone. Stale work is
  idempotently requeued through the unique object-name constraint.
- Object and database scans are independently cursor-paginated, with a batch
  range of 1–500.
- A database lease permits one active reconciliation; abandoned runs older
  than 30 minutes are failed before a new run starts.
- Unsafe provider keys are retained only as SHA-256 identifiers in audit rows.
  A plain object name is stored only after strict validation.

## Operation

Run against the same database and private object store used by the workers:

```sh
# Non-destructive inventory pass.
docker compose --profile staging --profile tools run --rm media-reconcile \
  /usr/local/bin/media-reconcile --batch-size 100 --orphan-grace 24h

# Explicitly enqueue only aged, validated safe orphans.
docker compose --profile staging --profile tools run --rm media-reconcile \
  /usr/local/bin/media-reconcile --apply-safe-orphans \
  --batch-size 100 --orphan-grace 24h
```

Continue with `NextObjectCursor` and `NextReferenceCursor` until both are empty.
Operators must review discrepancy classes before apply mode. Do not shorten the
grace period to conceal an upload incident.

Phase 11's isolated MinIO dry run inspected an empty canonical media namespace,
found zero discrepancies, and scheduled zero deletions. Unit and PostgreSQL/S3
integration tests cover interrupted registration, replacement safety, stale and
completed deletion, duplicate references, ownership mismatch, unexpected keys,
missing objects, repeated apply, and concurrent-run exclusion. This does not
prove compatibility with a future managed provider's inventory semantics.
