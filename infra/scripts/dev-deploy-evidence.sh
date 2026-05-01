#!/usr/bin/env sh
set -eu

usage() {
  echo "Usage: $0 [dev]" >&2
}

environment="${1:-dev}"
case "$environment" in
  dev) ;;
  *)
    usage
    exit 2
    ;;
esac

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
compose_file="$root_dir/infra/compose/docker-compose.$environment.yml"
env_file="$root_dir/infra/env/$environment.env"
example_env_file="$root_dir/infra/env/$environment.env.example"
override_base_url="${SMOKE_BASE_URL:-}"
override_api_base_url="${SMOKE_API_BASE_URL:-}"
override_access_token="${SMOKE_ACCESS_TOKEN:-}"
override_login_email="${SMOKE_LOGIN_EMAIL:-}"
override_login_password="${SMOKE_LOGIN_PASSWORD:-}"

if [ ! -f "$compose_file" ]; then
  echo "Missing compose file: $compose_file" >&2
  exit 1
fi

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

if [ "$override_base_url" != "" ]; then
  SMOKE_BASE_URL="$override_base_url"
  export SMOKE_BASE_URL
fi
if [ "$override_api_base_url" != "" ]; then
  SMOKE_API_BASE_URL="$override_api_base_url"
  export SMOKE_API_BASE_URL
fi
if [ "$override_access_token" != "" ]; then
  SMOKE_ACCESS_TOKEN="$override_access_token"
  export SMOKE_ACCESS_TOKEN
fi
if [ "$override_login_email" != "" ]; then
  SMOKE_LOGIN_EMAIL="$override_login_email"
  export SMOKE_LOGIN_EMAIL
fi
if [ "$override_login_password" != "" ]; then
  SMOKE_LOGIN_PASSWORD="$override_login_password"
  export SMOKE_LOGIN_PASSWORD
fi

base_url="${override_base_url:-${SMOKE_BASE_URL:-http://localhost:${PUBLIC_HTTP_PORT:-8088}}}"
api_health_url="$base_url/api/v1/health"

compose() {
  docker compose --env-file "$env_file" -f "$compose_file" "$@"
}

http_status() {
  name="$1"
  url="$2"
  expected_statuses="${3:-200}"

  status="$(curl -sS -o /dev/null -w '%{http_code}' "$url" 2>/dev/null || true)"
  if [ "$status" = "" ]; then
    status="curl_failed"
  fi

  printf '%-16s %s %s\n' "$name" "$status" "$url"
  for expected in $expected_statuses; do
    if [ "$status" = "$expected" ]; then
      return 0
    fi
  done

  echo "$name failed: HTTP $status, expected one of: $expected_statuses" >&2
  exit 1
}

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required for deploy evidence" >&2
  exit 1
fi

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required for deploy evidence" >&2
  exit 1
fi

generated_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
commit_sha="$(git -C "$root_dir" rev-parse HEAD)"
commit_short="$(git -C "$root_dir" rev-parse --short HEAD)"
commit_subject="$(git -C "$root_dir" log -1 --pretty=%s)"
branch_name="$(git -C "$root_dir" rev-parse --abbrev-ref HEAD)"

echo "# ERP $environment Deploy Evidence"
echo
echo "Generated: $generated_at"
echo "Branch: $branch_name"
echo "Commit: $commit_short $commit_subject"
echo "Commit SHA: $commit_sha"
echo "Base URL: $base_url"
echo
echo "## Health"
echo
echo '```text'
http_status "healthz" "$base_url/healthz"
http_status "api_health" "$api_health_url"
http_status "web_root" "$base_url/" "200 307"
echo '```'
echo
echo "## Containers"
echo
echo '```text'
compose ps postgres redis minio mailhog api worker web reverse-proxy
echo '```'
echo
echo "## Full Dev Smoke"
echo
echo '```text'
if "$root_dir/infra/scripts/smoke-dev-full.sh"; then
  echo '```'
  echo
  echo "Smoke result: PASS"
else
  status="$?"
  echo '```'
  echo
  echo "Smoke result: FAIL ($status)"
  exit "$status"
fi
