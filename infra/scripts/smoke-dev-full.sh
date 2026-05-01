#!/usr/bin/env sh
set -eu

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
env_file="$root_dir/infra/env/dev.env"
example_env_file="$root_dir/infra/env/dev.env.example"
override_base_url="${SMOKE_BASE_URL:-}"
override_api_base="${SMOKE_API_BASE_URL:-}"
override_access_token="${SMOKE_ACCESS_TOKEN:-}"
override_login_email="${SMOKE_LOGIN_EMAIL:-}"
override_login_password="${SMOKE_LOGIN_PASSWORD:-}"

if [ ! -f "$env_file" ]; then
  env_file="$example_env_file"
fi

if [ -f "$env_file" ]; then
  set -a
  . "$env_file"
  set +a
fi

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required for full dev smoke" >&2
  exit 1
fi

base_url="${override_base_url:-${SMOKE_BASE_URL:-http://localhost:${PUBLIC_HTTP_PORT:-8088}}}"
api_base="${override_api_base:-${SMOKE_API_BASE_URL:-$base_url/api/v1}}"
access_token="${override_access_token:-${SMOKE_ACCESS_TOKEN:-local-dev-access-token}}"
login_email="${override_login_email:-${SMOKE_LOGIN_EMAIL:-admin@example.local}}"
login_password="${override_login_password:-${SMOKE_LOGIN_PASSWORD:-local-only-mock-password}}"

tmp_body="$(mktemp)"
trap 'rm -f "$tmp_body"' EXIT

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

curl_check() {
  name="$1"
  method="$2"
  url="$3"
  expected_status="${4:-200}"
  body="${5:-}"
  auth="${6:-auth}"

  set -- -sS -o "$tmp_body" -w '%{http_code} %{size_download}' -X "$method" -H "Accept: application/json"
  if [ "$auth" = "auth" ]; then
    set -- "$@" -H "Authorization: Bearer $access_token"
  fi
  if [ "$body" != "" ]; then
    set -- "$@" -H "Content-Type: application/json" --data "$body"
  fi

  status_size="$(curl "$@" "$url")"
  status="${status_size%% *}"
  size="${status_size##* }"

  if [ "$status" != "$expected_status" ]; then
    echo "$name failed: HTTP $status, expected $expected_status" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  if [ "$size" = "0" ]; then
    echo "$name failed: empty response body" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "$name" "$status" "$size"
}

json_check() {
  curl_check "$1" GET "$2" 200 "" auth
  if ! grep -q '"success"[[:space:]]*:[[:space:]]*true' "$tmp_body"; then
    echo "$1 failed: response is not a success envelope" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
}

csv_check() {
  curl_check "$1" GET "$2" 200 "" auth
}

login_body="$(printf '{"email":"%s","password":"%s"}' "$(json_escape "$login_email")" "$(json_escape "$login_password")")"

echo "Running full ERP dev smoke against $base_url"
curl_check "healthz" GET "$base_url/healthz" 200 "" noauth
json_check "api_health" "$api_base/health"
curl_check "login" POST "$api_base/auth/login" 200 "$login_body" noauth
if ! grep -q '"access_token"[[:space:]]*:' "$tmp_body"; then
  echo "login failed: access_token missing" >&2
  sed -n '1,20p' "$tmp_body" >&2
  exit 1
fi

json_check "warehouse_fulfillment" "$api_base/warehouse/daily-board/fulfillment-metrics?business_date=2026-04-30&warehouse_id=wh-hcm"
json_check "warehouse_inbound" "$api_base/warehouse/daily-board/inbound-metrics?business_date=2026-04-30&warehouse_id=wh-hcm"
json_check "warehouse_subcontract" "$api_base/warehouse/daily-board/subcontract-metrics?business_date=2026-04-30&warehouse_id=wh-hcm"
json_check "finance_dashboard" "$api_base/finance/dashboard?business_date=2026-05-08"

json_check "inventory_report_json" "$api_base/reports/inventory-snapshot?business_date=2026-04-30&warehouse_id=wh-hcm&item_id=item-serum-30ml&status=quarantine&expiry_warning_days=45"
csv_check "inventory_report_csv" "$api_base/reports/inventory-snapshot/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&item_id=item-serum-30ml&status=quarantine&expiry_warning_days=45"

json_check "operations_report_json" "$api_base/reports/operations-daily?business_date=2026-04-30&warehouse_id=wh-hcm&status=blocked"
csv_check "operations_report_csv" "$api_base/reports/operations-daily/export.csv?business_date=2026-04-30&warehouse_id=wh-hcm&status=pending"

json_check "finance_report_json" "$api_base/reports/finance-summary?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08"
csv_check "finance_report_csv" "$api_base/reports/finance-summary/export.csv?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08"

echo "Full ERP dev smoke passed"
