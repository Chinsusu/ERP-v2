#!/usr/bin/env sh
set -eu

usage() {
  echo "Usage: $0 dev|staging"
}

environment="${1:-dev}"
case "$environment" in
  dev|staging) ;;
  *)
    usage
    exit 2
    ;;
esac

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
compose_file="$root_dir/infra/compose/docker-compose.$environment.yml"
env_file="$root_dir/infra/env/$environment.env"
example_env_file="$root_dir/infra/env/$environment.env.example"

if [ ! -f "$env_file" ]; then
  env_file="$example_env_file"
fi

if [ ! -f "$env_file" ]; then
  echo "Missing env file: $env_file" >&2
  exit 1
fi

set -a
. "$env_file"
set +a

base_url="${SMOKE_BASE_URL:-http://localhost:${PUBLIC_HTTP_PORT:-8088}}"
api_url="${SMOKE_API_URL:-$base_url/api/v1/health}"
web_url="${SMOKE_WEB_URL:-$base_url/}"

compose() {
  docker compose --env-file "$env_file" -f "$compose_file" "$@"
}

echo "Running internal smoke for ERP $environment"
compose --profile smoke run --rm smoke

if command -v curl >/dev/null 2>&1; then
  echo "Running host smoke for ERP $environment"
  curl -fsS "$base_url/healthz" >/dev/null
  curl -fsS "$api_url" >/dev/null
  curl -fsS "$web_url" >/dev/null

  if [ "$environment" = "dev" ]; then
    "$root_dir/infra/scripts/smoke-dev-full.sh"
  fi
else
  echo "curl not found on host; internal smoke already passed"
fi

echo "Smoke passed for ERP $environment"
