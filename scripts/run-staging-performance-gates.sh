#!/bin/sh
set -eu

origin=${1:-https://localhost:8443}
repository_root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$repository_root/backend"

go run ./cmd/loadtest -scenario readiness -url "$origin/api/v1/health/ready" \
  -allow-self-signed-localhost -requests 300 -concurrency 16 -success 200 \
  -max-error-rate 0 -max-p95 500ms
go run ./cmd/loadtest -scenario catalog -url "$origin/api/catalog/products?page_size=24" \
  -allow-self-signed-localhost -requests 200 -concurrency 12 -success 200 \
  -max-error-rate 0 -max-p95 500ms
go run ./cmd/loadtest -scenario recommendation -url "$origin/api/recommendations/generate" \
  -allow-self-signed-localhost -method POST \
  -header 'Content-Type: application/json' \
  -body-file ../scripts/load/recommendation.json \
  -requests 12 -concurrency 2 -success 200 \
  -max-error-rate 0 -max-p95 2s
go run ./cmd/loadtest -scenario invalid-login -url "$origin/api/auth/login" \
  -allow-self-signed-localhost -method POST \
  -header 'Content-Type: application/json' \
  -body-file ../scripts/load/login-invalid.json \
  -requests 8 -concurrency 2 -success 401 \
  -max-error-rate 0 -max-p95 1500ms
go run ./cmd/loadtest -scenario consented-analytics -url "$origin/api/analytics/events" \
  -allow-self-signed-localhost -method POST \
  -header 'Content-Type: application/json' \
  -body-file ../scripts/load/analytics-page-view.json \
  -setup-url "$origin/api/analytics/consent" -setup-method PUT \
  -setup-body-file ../scripts/load/analytics-consent.json -setup-success 200 \
  -requests 40 -concurrency 4 -success 204 \
  -max-error-rate 0 -max-p95 750ms
go run ./cmd/loadtest -scenario admin-authorization-boundary \
  -url "$origin/api/admin/products?page_size=25" \
  -allow-self-signed-localhost -requests 40 -concurrency 4 -success 401 \
  -max-error-rate 0 -max-p95 500ms
go run ./cmd/loadtest -scenario public-route-404 -url "$origin/not-a-real-route" \
  -allow-self-signed-localhost -requests 100 -concurrency 8 -success 404 \
  -max-error-rate 0 -max-p95 500ms
