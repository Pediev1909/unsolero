# UNSOLERO migration safety

Status: repository migration mechanics validated; production rollout rehearsal blocked

Migrations are immutable, checksummed, forward-only, serialized by a PostgreSQL
advisory lock, and executed one file per transaction. A failed statement rolls
back its migration and is not recorded. API readiness now embeds the release
manifest and fails when the database is older, newer, renamed, or has a changed
checksum. The API never applies migrations.

All 17 current files are **transactional** under the migration runner and are
reversible only while their own transaction is still uncommitted. None is
non-transactional. After commit, the repository has no down migrations, so each
is **logically irreversible** operationally: recovery is an explicitly reviewed
corrective forward migration or an authorized database restore. The table below
classifies likely lock/expense risk; it does not claim measured production lock
duration.

## Classification

`Lock risk` describes likely catalog/metadata locks, not measured production
duration. `Rollback` means logical application rollback after commit; every
file is transactionally rolled back on execution failure.

| Migration | Change class | Data/backfill | Lock risk | Application compatibility | Rollback |
| --- | --- | --- | --- | --- | --- |
| 000001 schemas | additive | none | low | broad | corrective forward/drop only |
| 000002 core domain | additive tables/indexes | none | low on empty DB | foundation | restore/corrective forward |
| 000003 sessions | additive | none | low | old app unaffected | corrective forward |
| 000004 public catalog | constraint replacement, indexes, click table | none | medium on image table/index builds | old app mostly compatible | corrective forward |
| 000005 recommendation experience | additive plus recommendation columns/constraints | existing recommendations receive defaults/nulls | medium | old readers likely compatible; new writer required for new fields | corrective forward |
| 000006 comparisons | additive | none | low | compatible | corrective forward |
| 000007 affiliate commerce | rename plus columns, constraints, indexes | legacy clicks receive synthetic `legacy-{id}` anonymous IDs | high on click table | requires coordinated commerce release | restore or corrective forward |
| 000008 admin | additive plus image URL constraint replacement | role definition only; no grants | medium | compatible | corrective forward |
| 000009 analytics foundation | event-name rewrite, columns, indexes, conversion/cost tables | canonicalizes existing event names | high on large event table | deploy after migration; reports depend on new fields | restore or corrective forward |
| 000010 editorial | additive | none | low | compatible | corrective forward |
| 000011 evidence governance | new schema/tables, product FKs, status backfill, demo evidence | non-demo published rows fail closed to draft; demo rows receive fictional provenance | high and business-sensitive | coordinated release mandatory | restore or reviewed corrective forward |
| 000012 data-driven policy | new policy tables, snapshot/session columns, policy activation backfill | maps existing demo policy data | high and recommendation-sensitive | coordinated release mandatory | restore or new policy/migration |
| 000013 policy immutability | triggers/functions | none | medium | new writes must honor workflow | remove only via corrective migration |
| 000014 commerce operations | new operational tables, offer/click columns, immutable observation triggers | existing clicks receive retention/classification defaults | high on offers/clicks | coordinated worker/API release | restore or corrective forward |
| 000015 verified conversions | new delivery/import/event/reconciliation tables plus conversion projection evolution | legacy conversions remain explicitly unverified | high on conversion table | coordinated worker/API release | restore or corrective forward |
| 000016 account security | session columns, security/MFA/token tables, roles, immutable security trigger | existing sessions default to password auth | medium/high on sessions | coordinated auth release | restore or corrective forward |
| 000017 analytics privacy | consent/receipt tables plus large event-table backfill/indexes | historical events become non-reportable and receive IDs/expiry | high on analytics events | coordinated analytics/API release | restore or corrective forward |

No migration uses `CREATE INDEX CONCURRENTLY`; the runner wraps each file in a
transaction, where concurrent index creation is not allowed. For large
production relations, rehearse on a size-equivalent clone. If measured locks
exceed the maintenance budget, split the change into an expand/backfill/validate
/contract sequence and use a separately reviewed non-transactional operational
procedure where necessary.

## Deployment policy

1. Classify every proposed migration for lock duration, rewrite/backfill size,
   old/new application compatibility, worker impact, and recovery path.
2. Test from an exact production schema clone with production-like row counts,
   query plans, lock observation, disk/WAL growth, and replica lag.
3. Take and verify a backup/PITR point before change. Stop affected workers.
4. Apply once through the migration image. Never run DDL from API startup.
5. Require migration command success, checksum match, API readiness, and smoke
   invariants before traffic.
6. Prefer additive expand/contract releases. Do not drop/rename a field while
   an old serving release or queued worker still needs it.
7. Roll application code back only when it remains compatible with the new
   schema. Database rollback is a forward correction or authorized restore.

## Evidence from Phase 8

- A deliberate failing migration created neither its table nor history row.
- Current schema and embedded release manifest matched all 17 checksums.
- Changed-checksum and missing-migration manifests failed closed.
- A reversible extra migration record made readiness return 503; removing it
  restored readiness to 200.
- Clean restore verified 17 migrations; non-empty and corrupt restores were
  rejected.

Production lock duration, replica behavior, WAL/disk headroom, zero-downtime
compatibility between real release pairs, and managed-database permissions
remain not tested.
