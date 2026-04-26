#!/usr/bin/env sh
set -eu

direction="${1:-up}"

if ! command -v migrate >/dev/null 2>&1; then
  echo "golang-migrate CLI is required for migrations" >&2
  exit 1
fi

: "${DB_HOST:=localhost}"
: "${DB_PORT:=5432}"
: "${DB_NAME:=erp}"
: "${DB_USER:=erp}"
: "${DB_PASSWORD:=erp}"
: "${DB_SSL_MODE:=disable}"

database_url="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}"

case "$direction" in
  up)
    migrate -path migrations -database "$database_url" up
    ;;
  down)
    migrate -path migrations -database "$database_url" down 1
    ;;
  *)
    echo "usage: migrate.sh up|down" >&2
    exit 1
    ;;
esac
