#!/bin/sh
set -eu

matches=$(git grep -n -E \
  '(dangerouslySetInnerHTML|\.innerHTML[[:space:]]*=|document\.write[[:space:]]*\(|\beval[[:space:]]*\(|new[[:space:]]+Function[[:space:]]*\()' \
  -- 'frontend/src/**/*.ts' 'frontend/src/**/*.tsx' 'frontend/src/**/*.js' \
  'frontend/src/**/*.jsx' 'frontend/src/**/*.html' 'backend/**/*.go' || true)
if [ -n "$matches" ]; then
  echo "unsafe HTML or execution sink requires explicit security review:" >&2
  echo "$matches" >&2
  exit 1
fi
echo "unsafe_web_sink_scan=pass"
