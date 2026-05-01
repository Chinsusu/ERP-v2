#!/usr/bin/env sh
set -eu

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
env_file="$root_dir/infra/env/dev.env"
example_env_file="$root_dir/infra/env/dev.env.example"
compose_file="$root_dir/infra/compose/docker-compose.dev.yml"
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

postgres_scalar() {
  sql="$1"

  if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required for persisted stock movement smoke" >&2
    exit 1
  fi
  if [ ! -f "$compose_file" ]; then
    echo "Missing compose file: $compose_file" >&2
    exit 1
  fi

  docker compose --env-file "$env_file" -f "$compose_file" exec -T postgres \
    psql -U "${POSTGRES_USER:-erp_dev}" -d "${POSTGRES_DB:-erp_dev}" -tAc "$sql" </dev/null |
    tr -d '[:space:]'
}

persisted_stock_movement_check() {
  smoke_index="$(postgres_scalar "select count(*) + 1 from inventory.stock_ledger where source_doc_type = 'stock_adjustment' and movement_no like 'ADJ-S9-03-03-SMOKE-%'")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_stock_movement failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  adjustment_id="00000000-0000-4000-8000-00000009$suffix"
  line_id="00000000-0000-4000-8000-00000008$suffix"
  adjustment_no="ADJ-S9-03-03-SMOKE-$suffix"
  before_count="$(postgres_scalar "select count(*) from inventory.stock_ledger where source_doc_id = '$adjustment_id'::uuid")"

  body="$(printf '{"id":"%s","adjustment_no":"%s","org_id":"00000000-0000-4000-8000-000000000001","warehouse_id":"00000000-0000-4000-8000-000000000801","warehouse_code":"warehouse_main","source_type":"smoke","source_id":"%s","reason":"S9-03-03 persisted stock movement smoke","lines":[{"id":"%s","item_id":"00000000-0000-4000-8000-000000001101","sku":"FG-LIP-001","location_id":"00000000-0000-4000-8000-000000001001","expected_qty":"20","counted_qty":"21","base_uom_code":"PCS","reason":"persisted stock movement smoke"}]}' "$adjustment_id" "$adjustment_no" "$adjustment_id" "$line_id")"

  curl_check "stock_adjustment_create" POST "$api_base/stock-adjustments" 201 "$body" auth
  curl_check "stock_adjustment_submit" POST "$api_base/stock-adjustments/$adjustment_id/submit" 200 "" auth
  curl_check "stock_adjustment_approve" POST "$api_base/stock-adjustments/$adjustment_id/approve" 200 "" auth
  curl_check "stock_adjustment_post" POST "$api_base/stock-adjustments/$adjustment_id/post" 200 "" auth

  after_count="$(postgres_scalar "select count(*) from inventory.stock_ledger where source_doc_id = '$adjustment_id'::uuid")"
  if [ "$before_count" != "0" ] || [ "$after_count" != "1" ]; then
    echo "persisted_stock_movement failed: ledger count before=$before_count after=$after_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_stock_movement" "ok" "$adjustment_no"
}

persisted_audit_login_count() {
  postgres_scalar "select count(*) from audit.audit_logs where actor_ref = 'user-erp-admin' and entity_type = 'auth.session' and action = 'auth.login_succeeded'"
}

persisted_audit_login_check() {
  before_count="$1"
  after_count="$(persisted_audit_login_count)"
  case "$before_count" in
    ''|*[!0-9]*)
      echo "persisted_audit_login failed: invalid before count '$before_count'" >&2
      exit 1
      ;;
  esac
  case "$after_count" in
    ''|*[!0-9]*)
      echo "persisted_audit_login failed: invalid after count '$after_count'" >&2
      exit 1
      ;;
  esac
  if [ "$after_count" -le "$before_count" ]; then
    echo "persisted_audit_login failed: count did not increase before=$before_count after=$after_count" >&2
    exit 1
  fi

  printf '%-28s %s %s->%s\n' "persisted_audit_login" "ok" "$before_count" "$after_count"
}

login_body="$(printf '{"email":"%s","password":"%s"}' "$(json_escape "$login_email")" "$(json_escape "$login_password")")"
audit_login_before_count="$(persisted_audit_login_count)"

echo "Running full ERP dev smoke against $base_url"
curl_check "healthz" GET "$base_url/healthz" 200 "" noauth
json_check "api_health" "$api_base/health"
curl_check "login" POST "$api_base/auth/login" 200 "$login_body" noauth
if ! grep -q '"access_token"[[:space:]]*:' "$tmp_body"; then
  echo "login failed: access_token missing" >&2
  sed -n '1,20p' "$tmp_body" >&2
  exit 1
fi
persisted_audit_login_check "$audit_login_before_count"

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

persisted_stock_movement_check

echo "Full ERP dev smoke passed"
