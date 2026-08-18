#!/bin/sh
set -eu

matches=$(rg --hidden --line-number --no-heading \
  --glob '!**/.git/**' --glob '!**/node_modules/**' --glob '!**/dist/**' \
  --glob '!backups/**' --glob '!scripts/check-secret-patterns.sh' \
  '(-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----|AKIA[0-9A-Z]{16}|sk-proj-[A-Za-z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9]{30,})' . || true)
if [ -n "$matches" ]; then
  echo "high-confidence secret pattern detected:" >&2
  echo "$matches" >&2
  exit 1
fi
echo "secret_pattern_scan=pass"
