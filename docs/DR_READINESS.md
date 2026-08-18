# UNSOLERO disaster-recovery readiness

Status: local backup/restore mechanism implemented; production DR external  
Last reviewed: 2026-08-17

## Evidence boundary

Locally implemented:

- PostgreSQL custom-format backup with timestamped metadata;
- SHA-256 integrity over both dump and metadata;
- ordered migration-manifest schema fingerprint;
- deterministic nonzero exit codes and incomplete-backup cleanup;
- empty-target restore guard, checksum/fingerprint verification, `pg_restore`
  validation, migration compatibility check, and post-restore table/count
  checks;
- non-root, read-only backup/restore tool containers.

Production deployment requirements:

- provider-managed encryption in transit and at rest with reviewed key
  ownership/rotation;
- immutable off-site storage, retention/lifecycle policy, access logging, legal
  hold behavior, alert delivery, multi-region/account isolation, and documented
  break-glass access;
- managed database point-in-time recovery/WAL configuration and provider-tested
  restoration;
- a clean staging recovery environment matching production versions.

Not yet exercised:

- real production dataset size, PITR, regional loss, compromised credentials,
  object-lock recovery, key loss/rotation, or a timed cross-team exercise;
- actual alert delivery and operator escalation;
- measured production RPO/RTO.

Local dumps are explicitly marked `encryption=none-local-artifact`. They are
test artifacts and must not contain production data or be treated as a valid
off-site backup.

## Proposed targets—not measurements

- Proposed RPO: 15 minutes using managed PITR plus daily immutable logical
  backups.
- Proposed RTO: 4 hours for database restoration and application validation.

Business, platform, security, and legal owners must approve these targets after
capacity testing. They are not achieved service levels.

## Operator exercise

1. Open an incident/change record and identify an approved backup by immutable
   ID, age, checksum, schema fingerprint, key version, and source environment.
2. Provision a new isolated recovery database with no public ingress. Never
   overwrite the source environment.
3. Verify object authenticity/integrity and decrypt through the approved KMS
   path. Record every access.
4. Run `postgres-restore.sh` against the empty target. Stop on any checksum,
   fingerprint, target-safety, or compatibility error.
5. Run migration compatibility, database integration tests, critical API smoke
   tests, recommendation reproducibility checks, and record/table reconciliation.
6. Verify authorization, session revocation posture, provider-disabled state,
   analytics retention, click/conversion immutability, and media references.
7. Measure last recoverable transaction and elapsed recovery time. Compare to
   approved RPO/RTO without changing the result.
8. Obtain incident commander, database owner, security, and product approval
   before any traffic decision. Keep commercial providers disabled.
9. Document gaps, owners, deadlines, evidence links, and securely destroy the
   exercise environment according to retention policy.

See `BACKUP_RESTORE.md` and `DISASTER_RECOVERY.md` for command-level procedures.
