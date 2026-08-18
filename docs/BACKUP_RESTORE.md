# UNSOLERO PostgreSQL backup and restore

Status: repository tooling verified locally; scheduled durable storage blocked

The repository provides custom-format logical backups using
`scripts/postgres-backup.sh` and clean-target restoration using
`scripts/postgres-restore.sh`. A successful backup is written atomically with a
SHA-256 checksum and metadata containing creation time, migration count, and a
fingerprint of the ordered migration version/name/checksum manifest. The
completion marker is written last and covers both dump and metadata. The restore
refuses a non-empty target, validates the marker/checksums/archive, restores in
one transaction, and verifies the exact migration-manifest fingerprint.

`scripts/encrypt-backup-age.sh` wraps the complete dump/metadata/checksum set in
an authenticated `age` recipient-encrypted archive, and
`scripts/decrypt-backup-age.sh` verifies the ciphertext checksum, rejects
unexpected archive paths, and restores the original verified set for the
normal restore command. The private identity belongs in a separate secret
management boundary and is never stored in this repository.

The local Phase 9 drill created `unsolero-phase9-verified`, restored it into a
fresh PostgreSQL 17 target, and verified all 17 migrations. A second restore to
the populated target was rejected with the documented exit code `4`. This is
local tooling evidence only; artifacts were unencrypted local test data and do
not prove off-site durability, PITR, RPO, or RTO.

## Local Docker drill

Choose a unique `BACKUP_NAME`; an existing backup is never overwritten.
Create the bind directory as the operator account and set `BACKUP_UID` and
`BACKUP_GID` to that account's numeric IDs. Backup and restore tooling runs as
that non-root identity and writes owner-only artifacts.

```bash
mkdir -p backups
BACKUP_UID=$(id -u) BACKUP_GID=$(id -g) \
BACKUP_NAME=phase9-drill docker compose --env-file .env --profile tools run --rm backup
BACKUP_UID=$(id -u) BACKUP_GID=$(id -g) \
BACKUP_NAME=phase9-drill docker compose --env-file .env --profile restore up -d restore-postgres
BACKUP_UID=$(id -u) BACKUP_GID=$(id -g) \
BACKUP_NAME=phase9-drill docker compose --env-file .env --profile restore run --rm restore
```

After restoration, run the current migration image against the restored
database, then application-specific smoke checks and row/count invariants before
considering it usable. A backup from a newer schema must not be restored into an
older application release. Migration checksums are immutable; forward-only
schema changes may make application rollback impossible.

## Production procedure

1. The database owner starts a provider snapshot and a logical backup before a
   schema deployment.
2. Store encrypted artifacts and checksums in a separate account/region with
   write-once or deletion-protected retention.
3. Record backup ID, schema migration, application version, start/end time, size,
   integrity result, and operator—never credentials.
4. Restore into an isolated clean database at least quarterly.
5. Apply current compatible migrations, verify checksums, run critical API and
   recommendation/commerce/account invariants, and record measured recovery time.
6. Restrict backup credentials to read-only backup duties where the provider
   supports it; restore credentials are separate and temporary.

## Objectives, not achieved claims

- Proposed initial RPO: 24 hours, requiring daily durable backups plus provider
  point-in-time recovery if the business approves a lower RPO.
- Proposed initial RTO: 4 hours, requiring a measured restore drill, trained
  owner, available infrastructure, and current runbook.
- Suggested retention: 14 daily, 8 weekly, and 12 monthly backups, subject to
  legal/privacy/security approval and deletion propagation requirements.

No cloud schedule, encryption key, off-site artifact, PITR configuration,
measured RPO/RTO, or successful production restore exists in this repository.
Backup failure must page an operator once an alert adapter is linked.

## Encrypted archive handoff

After a logical backup, an operator holding only the public recipient can
create an archive for off-site upload:

```bash
BACKUP_FILE=backups/phase-drill.dump \
ARCHIVE_DIRECTORY=encrypted-backups \
AGE_RECIPIENT='age1...' \
scripts/encrypt-backup-age.sh
```

Recovery requires a temporary restricted directory and an identity file
supplied by the secret manager:

```bash
ENCRYPTED_BACKUP_FILE=encrypted-backups/phase-drill.tar.age \
AGE_IDENTITY_FILE=/run/secrets/backup-age-identity \
RESTORE_DIRECTORY=restore-input \
scripts/decrypt-backup-age.sh
```

This tooling does not upload, schedule, retain, or prove durability. A hosted
object store, lifecycle/immutability policy, separated identity custody, and a
measured restore remain launch blockers.
