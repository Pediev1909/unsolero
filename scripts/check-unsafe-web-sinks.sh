#!/bin/sh
set -eu

matches=$(rg --line-number --no-heading \
  '(dangerouslySetInnerHTML|\.innerHTML[[:space:]]*=|document\.write[[:space:]]*\(|\beval[[:space:]]*\(|new[[:space:]]+Function[[:space:]]*\()' \
  frontend/src backend --glob '*.{ts,tsx,js,jsx,go,html}' || true)
if [ -n "$matches" ]; then
  echo "unsafe HTML or execution sink requires explicit security review:" >&2
  echo "$matches" >&2
  exit 1
fi
echo "unsafe_web_sink_scan=pass"
