#!/usr/bin/env sh
set -eu

environment="${1:-local}"
script_dir="$(CDPATH= cd -- "$(dirname "$0")" && pwd)"
api_dir="$(CDPATH= cd -- "${script_dir}/.." && pwd)"
seed_file="${api_dir}/sql/dev_seed.sql"

if [ "$environment" != "local" ]; then
  echo "usage: seed.sh local" >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql CLI is required for direct seed execution; use make seed-local for Docker-based seeding" >&2
  exit 1
fi

: "${DB_HOST:=localhost}"
: "${DB_PORT:=5432}"
: "${DB_NAME:=erp}"
: "${DB_USER:=erp}"
: "${DB_PASSWORD:=erp}"
: "${DB_SSL_MODE:=disable}"

database_url="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}"

psql "$database_url" -v ON_ERROR_STOP=1 -f "$seed_file"

echo "local seed applied"
echo "mock login: admin@example.local / local-only-mock-password"
echo "seeded users: admin@example.local, warehouse_user@example.local, sales_user@example.local, qc_user@example.local"
