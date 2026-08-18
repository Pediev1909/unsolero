#!/bin/sh
set -eu

tls_directory=/tmp/unsolero-tls
mkdir -p "$tls_directory"
umask 077
openssl req -x509 -newkey rsa:2048 -nodes -days 2 \
  -keyout "$tls_directory/tls.key" -out "$tls_directory/tls.crt" \
  -subj '/CN=localhost' -addext 'subjectAltName=DNS:localhost,IP:127.0.0.1' >/dev/null 2>&1
exec nginx -g 'daemon off;'
