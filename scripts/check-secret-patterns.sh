#!/bin/sh
set -eu

matches=$(git grep -n -E \
  '(-----BEGIN (RSA |EC |OPENSSH |DSA )?PRIVATE KEY-----|AKIA[0-9A-Z]{16}|sk-proj-[A-Za-z0-9_-]{20,}|gh[pousr]_[A-Za-z0-9]{30,})' \
  -- . ':(exclude)backups/**' ':(exclude)scripts/check-secret-patterns.sh' || true)
if [ -n "$matches" ]; then
  echo "high-confidence secret pattern detected:" >&2
  echo "$matches" >&2
  exit 1
fi
echo "secret_pattern_scan=pass"
