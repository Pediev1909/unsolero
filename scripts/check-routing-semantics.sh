#!/bin/sh
set -eu

origin=${1:-https://localhost:8443}
temporary_headers=$(mktemp)
trap 'rm -f "$temporary_headers"' EXIT HUP INT TERM

expect_status() {
  expected=$1
  path=$2
  actual=$(curl --insecure --path-as-is --silent --output /dev/null --write-out '%{http_code}' "$origin$path")
  if [ "$actual" != "$expected" ]; then
    echo "routing check failed: $path returned $actual, expected $expected" >&2
    exit 1
  fi
}

expect_status 200 /
expect_status 200 /products
expect_status 404 /definitely-not-an-unsolero-route
expect_status 404 /api/definitely-not-an-api-route
expect_status 400 /products/known%2Fmalformed
expect_status 308 /products/
expect_status 200 /sitemap.xml
expect_status 200 /robots.txt

curl --insecure --silent --dump-header "$temporary_headers" --output /dev/null "$origin/admin"
if ! tr -d '\r' < "$temporary_headers" | grep -qi '^x-robots-tag: noindex, nofollow$'; then
  echo "admin route does not emit an HTTP noindex policy" >&2
  exit 1
fi
curl --insecure --silent --dump-header "$temporary_headers" --output /dev/null "$origin/products"
if ! tr -d '\r' < "$temporary_headers" | grep -qi '^link: <.*/products>; rel="canonical"$'; then
  echo "public catalog route does not emit a canonical link" >&2
  exit 1
fi
echo "routing_semantics=pass"
