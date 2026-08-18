#!/bin/sh
set -eu

: "${BACKUP_FILE:?BACKUP_FILE is required}"
: "${ARCHIVE_DIRECTORY:?ARCHIVE_DIRECTORY is required}"
: "${AGE_RECIPIENT:?AGE_RECIPIENT is required}"

command -v age >/dev/null 2>&1 || {
  echo "age is required to encrypt backup archives" >&2
  exit 2
}

case "$ARCHIVE_DIRECTORY" in
  ""|"/") echo "ARCHIVE_DIRECTORY must be a dedicated directory" >&2; exit 2 ;;
esac

backup_directory=$(dirname "$BACKUP_FILE")
backup_filename=$(basename "$BACKUP_FILE")
case "$backup_filename" in
  *.dump) backup_name=${backup_filename%.dump} ;;
  *) echo "BACKUP_FILE must end in .dump" >&2; exit 2 ;;
esac

checksum_file="$BACKUP_FILE.sha256"
metadata_file="$BACKUP_FILE.metadata"
if [ ! -f "$BACKUP_FILE" ] || [ ! -f "$checksum_file" ] || [ ! -f "$metadata_file" ]; then
  echo "backup, checksum, and metadata files are required" >&2
  exit 3
fi
(cd "$backup_directory" && sha256sum -c "$(basename "$checksum_file")") >/dev/null

umask 077
mkdir -p "$ARCHIVE_DIRECTORY"
encrypted_file="$ARCHIVE_DIRECTORY/$backup_name.tar.age"
encrypted_checksum="$encrypted_file.sha256"
temporary_file="$ARCHIVE_DIRECTORY/.$backup_name.tar.age.partial"
if [ -e "$encrypted_file" ] || [ -e "$encrypted_checksum" ]; then
  echo "encrypted backup target already exists" >&2
  exit 4
fi

cleanup() { rm -f "$temporary_file" "$encrypted_checksum.partial"; }
trap cleanup EXIT HUP INT TERM

tar -C "$backup_directory" -cf - \
  "$backup_filename" "$(basename "$metadata_file")" "$(basename "$checksum_file")" |
  age --encrypt --recipient "$AGE_RECIPIENT" --output "$temporary_file"
test -s "$temporary_file"
sha256sum "$temporary_file" | sed "s|$temporary_file|$(basename "$encrypted_file")|" > "$encrypted_checksum.partial"
mv "$temporary_file" "$encrypted_file"
mv "$encrypted_checksum.partial" "$encrypted_checksum"

trap - EXIT HUP INT TERM
echo "encrypted_backup_created=$encrypted_file"
