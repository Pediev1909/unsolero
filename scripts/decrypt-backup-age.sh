#!/bin/sh
set -eu

: "${ENCRYPTED_BACKUP_FILE:?ENCRYPTED_BACKUP_FILE is required}"
: "${AGE_IDENTITY_FILE:?AGE_IDENTITY_FILE is required}"
: "${RESTORE_DIRECTORY:?RESTORE_DIRECTORY is required}"

command -v age >/dev/null 2>&1 || {
  echo "age is required to decrypt backup archives" >&2
  exit 2
}
case "$RESTORE_DIRECTORY" in
  ""|"/") echo "RESTORE_DIRECTORY must be a dedicated directory" >&2; exit 2 ;;
esac
case "$(basename "$ENCRYPTED_BACKUP_FILE")" in
  *.tar.age) backup_name=$(basename "$ENCRYPTED_BACKUP_FILE" .tar.age) ;;
  *) echo "ENCRYPTED_BACKUP_FILE must end in .tar.age" >&2; exit 2 ;;
esac

encrypted_checksum="$ENCRYPTED_BACKUP_FILE.sha256"
if [ ! -f "$ENCRYPTED_BACKUP_FILE" ] || [ ! -f "$encrypted_checksum" ] || [ ! -f "$AGE_IDENTITY_FILE" ]; then
  echo "encrypted backup, checksum, and age identity are required" >&2
  exit 3
fi
(cd "$(dirname "$ENCRYPTED_BACKUP_FILE")" && sha256sum -c "$(basename "$encrypted_checksum")") >/dev/null

umask 077
mkdir -p "$RESTORE_DIRECTORY"
for target in "$backup_name.dump" "$backup_name.dump.metadata" "$backup_name.dump.sha256"; do
  if [ -e "$RESTORE_DIRECTORY/$target" ]; then
    echo "restore target already contains $target" >&2
    exit 4
  fi
done

temporary_tar="$RESTORE_DIRECTORY/.$backup_name.decrypted.tar"
cleanup() { rm -f "$temporary_tar"; }
trap cleanup EXIT HUP INT TERM
age --decrypt --identity "$AGE_IDENTITY_FILE" --output "$temporary_tar" "$ENCRYPTED_BACKUP_FILE"

expected_entries=$(printf '%s\n%s\n%s\n' \
  "$backup_name.dump" "$backup_name.dump.metadata" "$backup_name.dump.sha256")
actual_entries=$(tar -tf "$temporary_tar")
if [ "$actual_entries" != "$expected_entries" ]; then
  echo "encrypted backup archive contains unexpected paths" >&2
  exit 5
fi
tar -xf "$temporary_tar" -C "$RESTORE_DIRECTORY"
(cd "$RESTORE_DIRECTORY" && sha256sum -c "$backup_name.dump.sha256") >/dev/null

rm -f "$temporary_tar"
trap - EXIT HUP INT TERM
echo "decrypted_backup_ready=$RESTORE_DIRECTORY/$backup_name.dump"
