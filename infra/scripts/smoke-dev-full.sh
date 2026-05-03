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

json_string_field() {
  field="$1"
  sed -n "s/.*\"$field\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p" "$tmp_body" | head -n 1
}

csv_check() {
  curl_check "$1" GET "$2" 200 "" auth
}

postgres_scalar() {
  sql="$1"

  if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required for persisted runtime smoke" >&2
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

postgres_exec() {
  sql="$1"

  if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required for persisted runtime smoke" >&2
    exit 1
  fi
  if [ ! -f "$compose_file" ]; then
    echo "Missing compose file: $compose_file" >&2
    exit 1
  fi

  printf '%s\n' "$sql" |
    docker compose --env-file "$env_file" -f "$compose_file" exec -T postgres \
      psql -v ON_ERROR_STOP=1 -q -X -U "${POSTGRES_USER:-erp_dev}" -d "${POSTGRES_DB:-erp_dev}" >/dev/null
}

restart_api_service() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "docker is required for api restart smoke" >&2
    exit 1
  fi
  if [ ! -f "$compose_file" ]; then
    echo "Missing compose file: $compose_file" >&2
    exit 1
  fi

  docker compose --env-file "$env_file" -f "$compose_file" restart api >/dev/null

  retries=30
  while [ "$retries" -gt 0 ]; do
    if curl -fsS "$api_base/health" >/dev/null 2>&1; then
      printf '%-28s %s %s\n' "api_restart" "ok" "finance-runtime"
      return
    fi
    retries=$((retries - 1))
    sleep 2
  done

  echo "api_restart failed: API health did not recover" >&2
  exit 1
}

persisted_masterdata_runtime_check() {
  org_id="00000000-0000-4000-8000-000000000001"
  postgres_exec "DELETE FROM mdm.warehouse_bins WHERE org_id = '$org_id'::uuid AND code LIKE 'LOC-S17-06-%'; DELETE FROM mdm.warehouses WHERE org_id = '$org_id'::uuid AND code LIKE 'WH-S17-06-%'; DELETE FROM mdm.suppliers WHERE org_id = '$org_id'::uuid AND code LIKE 'SUP-S17-06-%'; DELETE FROM mdm.customers WHERE org_id = '$org_id'::uuid AND code LIKE 'CUS-S17-06-%'; DELETE FROM mdm.items WHERE org_id = '$org_id'::uuid AND (item_code LIKE 'ITEM-S17-06-01-SMOKE-%' OR sku LIKE 'SKU-S17-06-%');"
  smoke_index="$(postgres_scalar "select greatest((select count(*) from mdm.items where item_code like 'ITEM-S17-06-01-SMOKE-%'), (select count(*) from mdm.warehouses where code like 'WH-S17-06-%'), (select count(*) from mdm.suppliers where code like 'SUP-S17-06-%'), (select count(*) from mdm.customers where code like 'CUS-S17-06-%')) + 1")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_masterdata_runtime failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  item_code="ITEM-S17-06-01-SMOKE-$suffix"
  sku_code="SKU-S17-06-$suffix"
  warehouse_code="WH-S17-06-$suffix"
  location_code="LOC-S17-06-$suffix"
  supplier_code="SUP-S17-06-$suffix"
  customer_code="CUS-S17-06-$suffix"

  product_body="$(cat <<EOF
{"item_code":"$item_code","sku_code":"$sku_code","name":"S17 master data smoke item $suffix","item_type":"finished_good","item_group":"smoke","brand_code":"MYH","uom_base":"EA","uom_purchase":"EA","uom_issue":"EA","lot_controlled":false,"expiry_controlled":false,"shelf_life_days":0,"qc_required":false,"status":"active","standard_cost":"125000.000000","is_sellable":true,"is_purchasable":true,"is_producible":false,"spec_version":"S17-06-01"}
EOF
)"
  warehouse_body="$(cat <<EOF
{"warehouse_code":"$warehouse_code","warehouse_name":"S17 Master Data Smoke Warehouse $suffix","warehouse_type":"finished_good","site_code":"HCM","address":"S17 smoke warehouse","allow_sale_issue":true,"allow_prod_issue":false,"allow_quarantine":true,"status":"active"}
EOF
)"
  location_body="$(cat <<EOF
{"warehouse_id":"$warehouse_code","location_code":"$location_code","location_name":"S17 Smoke Location $suffix","location_type":"storage","zone_code":"SMOKE","allow_receive":true,"allow_pick":true,"allow_store":true,"is_default":false,"status":"active"}
EOF
)"
  supplier_body="$(cat <<EOF
{"supplier_code":"$supplier_code","supplier_name":"S17 Smoke Supplier $suffix","supplier_group":"service","contact_name":"S17 Supplier","phone":"0900000000","email":"s17.supplier.$suffix@example.local","tax_code":"031706$suffix","address":"Ho Chi Minh","payment_terms":"NET15","lead_time_days":5,"moq":"1.000000","quality_score":"90.0000","delivery_score":"91.0000","status":"active"}
EOF
)"
  customer_body="$(cat <<EOF
{"customer_code":"$customer_code","customer_name":"S17 Smoke Customer $suffix","customer_type":"dealer","channel_code":"dealer","price_list_code":"PL-S17","discount_group":"SMOKE","credit_limit":"1000000.00","payment_terms":"NET15","contact_name":"S17 Customer","phone":"0911111111","email":"s17.customer.$suffix@example.local","tax_code":"031716$suffix","address":"Ho Chi Minh","status":"active"}
EOF
)"

  curl_check "masterdata_item_create" POST "$api_base/products" 201 "$product_body" auth
  curl_check "masterdata_wh_create" POST "$api_base/warehouses" 201 "$warehouse_body" auth
  curl_check "masterdata_loc_create" POST "$api_base/warehouse-locations" 201 "$location_body" auth
  curl_check "masterdata_supplier_create" POST "$api_base/suppliers" 201 "$supplier_body" auth
  curl_check "masterdata_customer_create" POST "$api_base/customers" 201 "$customer_body" auth

  restart_api_service

  curl_check "masterdata_item_read" GET "$api_base/products/$sku_code" 200 "" auth
  if ! grep -q "\"sku_code\":\"$sku_code\"" "$tmp_body"; then
    echo "persisted_masterdata_item failed: item not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "masterdata_wh_read" GET "$api_base/warehouses/$warehouse_code" 200 "" auth
  if ! grep -q "\"warehouse_code\":\"$warehouse_code\"" "$tmp_body"; then
    echo "persisted_masterdata_warehouse failed: warehouse not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "masterdata_loc_read" GET "$api_base/warehouse-locations/$location_code" 200 "" auth
  if ! grep -q "\"location_code\":\"$location_code\"" "$tmp_body"; then
    echo "persisted_masterdata_location failed: location not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "masterdata_supplier_read" GET "$api_base/suppliers/$supplier_code" 200 "" auth
  if ! grep -q "\"supplier_code\":\"$supplier_code\"" "$tmp_body"; then
    echo "persisted_masterdata_supplier failed: supplier not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "masterdata_customer_read" GET "$api_base/customers/$customer_code" 200 "" auth
  if ! grep -q "\"customer_code\":\"$customer_code\"" "$tmp_body"; then
    echo "persisted_masterdata_customer failed: customer not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  document_count="$(postgres_scalar "select (select count(*) from mdm.items where org_id = '$org_id'::uuid and item_code = '$item_code' and sku = '$sku_code' and status = 'active') + (select count(*) from mdm.warehouses where org_id = '$org_id'::uuid and code = '$warehouse_code' and status = 'active') + (select count(*) from mdm.warehouse_bins where org_id = '$org_id'::uuid and code = '$location_code' and status = 'active') + (select count(*) from mdm.suppliers where org_id = '$org_id'::uuid and code = '$supplier_code' and status = 'active') + (select count(*) from mdm.customers where org_id = '$org_id'::uuid and code = '$customer_code' and status = 'active')")"
  if [ "$document_count" != "5" ]; then
    echo "persisted_masterdata_runtime failed: document=$document_count" >&2
    exit 1
  fi

  postgres_exec "DELETE FROM mdm.warehouse_bins WHERE org_id = '$org_id'::uuid AND code = '$location_code'; DELETE FROM mdm.warehouses WHERE org_id = '$org_id'::uuid AND code = '$warehouse_code'; DELETE FROM mdm.suppliers WHERE org_id = '$org_id'::uuid AND code = '$supplier_code'; DELETE FROM mdm.customers WHERE org_id = '$org_id'::uuid AND code = '$customer_code'; DELETE FROM mdm.items WHERE org_id = '$org_id'::uuid AND sku = '$sku_code';"
  printf '%-28s %s %s\n' "persisted_masterdata" "ok" "$suffix"
}

persisted_finance_runtime_check() {
  org_id="00000000-0000-4000-8000-000000000001"
  smoke_index="$(postgres_scalar "select greatest((select count(*) from finance.customer_receivables where receivable_ref like 'ar-s15-06-03-smoke-%'), (select count(*) from finance.supplier_payables where payable_ref like 'ap-s15-06-03-smoke-%'), (select count(*) from finance.cod_remittances where remittance_ref like 'cod-s15-06-03-smoke-%'), (select count(*) from finance.cash_transactions where transaction_ref like 'cash-s15-06-03-smoke-%')) + 1")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_finance_runtime failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  ar_id="ar-s15-06-03-smoke-$suffix"
  ar_no="AR-S15-06-03-SMOKE-$suffix"
  ar_line_id="ar-line-s15-06-03-smoke-$suffix"
  shipment_id="shipment-s15-06-03-smoke-$suffix"
  shipment_no="SHP-S15-06-03-SMOKE-$suffix"
  ap_id="ap-s15-06-03-smoke-$suffix"
  ap_no="AP-S15-06-03-SMOKE-$suffix"
  ap_line_id="ap-line-s15-06-03-smoke-$suffix"
  qc_id="qc-s15-06-03-smoke-$suffix"
  qc_no="QC-S15-06-03-SMOKE-$suffix"
  receipt_id="gr-s15-06-03-smoke-$suffix"
  receipt_no="GR-S15-06-03-SMOKE-$suffix"
  cod_id="cod-s15-06-03-smoke-$suffix"
  cod_no="COD-S15-06-03-SMOKE-$suffix"
  cod_line_id="cod-line-s15-06-03-smoke-$suffix"
  tracking_no="GHN-S15-06-03-SMOKE-$suffix"
  cash_id="cash-s15-06-03-smoke-$suffix"
  cash_no="CASH-IN-S15-06-03-SMOKE-$suffix"
  cash_allocation_id="cash-alloc-s15-06-03-smoke-$suffix"

  before_count="$(postgres_scalar "select (select count(*) from finance.customer_receivables where receivable_ref = '$ar_id') + (select count(*) from finance.supplier_payables where payable_ref = '$ap_id') + (select count(*) from finance.cod_remittances where remittance_ref = '$cod_id') + (select count(*) from finance.cash_transactions where transaction_ref = '$cash_id')")"

  ar_body="$(cat <<EOF
{"id":"$ar_id","receivable_no":"$ar_no","customer_id":"customer-s15-06-03","customer_code":"CUS-S15-06-03","customer_name":"Sprint 15 Finance Smoke Customer","source_document":{"type":"shipment","id":"$shipment_id","no":"$shipment_no"},"lines":[{"id":"$ar_line_id","description":"S15-06-03 finance persistence smoke AR","source_document":{"type":"shipment","id":"$shipment_id","no":"$shipment_no"},"amount":"1250000.00"}],"total_amount":"1250000.00","currency_code":"VND","due_date":"2026-05-02"}
EOF
)"
  ap_body="$(cat <<EOF
{"id":"$ap_id","payable_no":"$ap_no","supplier_id":"supplier-s15-06-03","supplier_code":"SUP-S15-06-03","supplier_name":"Sprint 15 Finance Smoke Supplier","source_document":{"type":"qc_inspection","id":"$qc_id","no":"$qc_no"},"lines":[{"id":"$ap_line_id","description":"S15-06-03 finance persistence smoke AP","source_document":{"type":"warehouse_receipt","id":"$receipt_id","no":"$receipt_no"},"amount":"4250000.00"}],"total_amount":"4250000.00","currency_code":"VND","due_date":"2026-05-02"}
EOF
)"
  cod_body="$(cat <<EOF
{"id":"$cod_id","remittance_no":"$cod_no","carrier_id":"carrier-s15-06-03","carrier_code":"GHN","carrier_name":"GHN Express","business_date":"2026-05-02","expected_amount":"1250000.00","remitted_amount":"1200000.00","currency_code":"VND","lines":[{"id":"$cod_line_id","receivable_id":"$ar_id","receivable_no":"$ar_no","shipment_id":"$shipment_id","tracking_no":"$tracking_no","customer_name":"Sprint 15 Finance Smoke Customer","expected_amount":"1250000.00","remitted_amount":"1200000.00"}]}
EOF
)"
  cash_body="$(cat <<EOF
{"id":"$cash_id","transaction_no":"$cash_no","direction":"cash_in","business_date":"2026-05-02","counterparty_id":"carrier-s15-06-03","counterparty_name":"GHN Express","payment_method":"bank_transfer","reference_no":"BANK-S15-06-03-$suffix","allocations":[{"id":"$cash_allocation_id","target_type":"customer_receivable","target_id":"$ar_id","target_no":"$ar_no","amount":"1250000.00"}],"total_amount":"1250000.00","currency_code":"VND","memo":"S15-06-03 finance persistence smoke"}
EOF
)"

  curl_check "finance_ar_create" POST "$api_base/customer-receivables" 201 "$ar_body" auth
  curl_check "finance_ap_create" POST "$api_base/supplier-payables" 201 "$ap_body" auth
  curl_check "finance_cod_create" POST "$api_base/cod-remittances" 201 "$cod_body" auth
  curl_check "finance_cash_create" POST "$api_base/cash-transactions" 201 "$cash_body" auth

  restart_api_service

  curl_check "finance_ar_read" GET "$api_base/customer-receivables/$ar_id" 200 "" auth
  if ! grep -q "\"id\":\"$ar_id\"" "$tmp_body" ||
    ! grep -q '"status":"open"' "$tmp_body" ||
    ! grep -q '"outstanding_amount":"1250000.00"' "$tmp_body"; then
    echo "persisted_finance_ar failed: receivable not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  curl_check "finance_ap_read" GET "$api_base/supplier-payables/$ap_id" 200 "" auth
  if ! grep -q "\"id\":\"$ap_id\"" "$tmp_body" ||
    ! grep -q '"status":"open"' "$tmp_body" ||
    ! grep -q '"outstanding_amount":"4250000.00"' "$tmp_body"; then
    echo "persisted_finance_ap failed: payable not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  curl_check "finance_cod_read" GET "$api_base/cod-remittances/$cod_id" 200 "" auth
  if ! grep -q "\"id\":\"$cod_id\"" "$tmp_body" ||
    ! grep -q '"status":"draft"' "$tmp_body" ||
    ! grep -q '"discrepancy_amount":"-50000.00"' "$tmp_body"; then
    echo "persisted_finance_cod failed: COD remittance not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  curl_check "finance_cash_read" GET "$api_base/cash-transactions/$cash_id" 200 "" auth
  if ! grep -q "\"id\":\"$cash_id\"" "$tmp_body" ||
    ! grep -q '"status":"posted"' "$tmp_body" ||
    ! grep -q "\"id\":\"$cash_allocation_id\"" "$tmp_body"; then
    echo "persisted_finance_cash failed: cash transaction not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  json_check "finance_dashboard_after" "$api_base/finance/dashboard?business_date=2026-05-02"
  json_check "finance_report_after" "$api_base/reports/finance-summary?from_date=2026-05-02&to_date=2026-05-02&business_date=2026-05-02"

  ar_count="$(postgres_scalar "select count(*) from finance.customer_receivables r join finance.customer_receivable_lines l on l.customer_receivable_id = r.id where r.org_id = '$org_id'::uuid and r.receivable_ref = '$ar_id' and r.receivable_no = '$ar_no' and r.status = 'open' and r.outstanding_amount = 1250000.00 and l.line_ref = '$ar_line_id' and l.source_document_ref = '$shipment_id' and l.amount = 1250000.00")"
  ap_count="$(postgres_scalar "select count(*) from finance.supplier_payables p join finance.supplier_payable_lines l on l.supplier_payable_id = p.id where p.org_id = '$org_id'::uuid and p.payable_ref = '$ap_id' and p.payable_no = '$ap_no' and p.status = 'open' and p.outstanding_amount = 4250000.00 and l.line_ref = '$ap_line_id' and l.source_document_ref = '$receipt_id' and l.amount = 4250000.00")"
  cod_count="$(postgres_scalar "select count(*) from finance.cod_remittances r join finance.cod_remittance_lines l on l.cod_remittance_id = r.id where r.org_id = '$org_id'::uuid and r.remittance_ref = '$cod_id' and r.remittance_no = '$cod_no' and r.status = 'draft' and r.discrepancy_amount = -50000.00 and l.line_ref = '$cod_line_id' and l.receivable_ref = '$ar_id' and l.discrepancy_amount = -50000.00")"
  cash_count="$(postgres_scalar "select count(*) from finance.cash_transactions t join finance.cash_transaction_allocations a on a.cash_transaction_id = t.id where t.org_id = '$org_id'::uuid and t.transaction_ref = '$cash_id' and t.transaction_no = '$cash_no' and t.status = 'posted' and t.direction = 'cash_in' and t.total_amount = 1250000.00 and a.allocation_ref = '$cash_allocation_id' and a.target_ref = '$ar_id' and a.amount = 1250000.00")"
  audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref in ('$ar_id', '$ap_id', '$cod_id', '$cash_id') and action in ('finance.customer_receivable.created', 'finance.supplier_payable.created', 'finance.cod_remittance.created', 'finance.cash_transaction.recorded')")"
  if [ "$before_count" != "0" ] || [ "$ar_count" != "1" ] || [ "$ap_count" != "1" ] || [ "$cod_count" != "1" ] || [ "$cash_count" != "1" ] || [ "$audit_count" != "4" ]; then
    echo "persisted_finance_runtime failed: before=$before_count ar=$ar_count ap=$ap_count cod=$cod_count cash=$cash_count audit=$audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_finance_ar" "ok" "$ar_no"
  printf '%-28s %s %s\n' "persisted_finance_ap" "ok" "$ap_no"
  printf '%-28s %s %s\n' "persisted_finance_cod" "ok" "$cod_no"
  printf '%-28s %s %s\n' "persisted_finance_cash" "ok" "$cash_no"
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
  document_count="$(postgres_scalar "select count(*) from inventory.stock_adjustments a join inventory.stock_adjustment_lines l on l.adjustment_id = a.id where a.org_id = '00000000-0000-4000-8000-000000000001'::uuid and a.adjustment_ref = '$adjustment_id' and a.status = 'posted' and a.posted_by_ref = 'user-erp-admin' and l.line_ref = '$line_id' and l.delta_qty = 1.000000")"
  if [ "$before_count" != "0" ] || [ "$after_count" != "1" ] || [ "$document_count" != "1" ]; then
    echo "persisted_stock_movement failed: ledger before=$before_count after=$after_count document=$document_count" >&2
    exit 1
  fi

  expected_on_hand="$(postgres_scalar "select to_char(coalesce(sum(qty_on_hand), 0), 'FM999999999999990.000000') from inventory.stock_balances where org_id = '00000000-0000-4000-8000-000000000001'::uuid and warehouse_id = '00000000-0000-4000-8000-000000000801'::uuid and bin_id = '00000000-0000-4000-8000-000000001001'::uuid and item_id = '00000000-0000-4000-8000-000000001101'::uuid and batch_id is null")"
  expected_available="$(postgres_scalar "select to_char(coalesce(sum(qty_available), 0), 'FM999999999999990.000000') from inventory.stock_balances where org_id = '00000000-0000-4000-8000-000000000001'::uuid and warehouse_id = '00000000-0000-4000-8000-000000000801'::uuid and bin_id = '00000000-0000-4000-8000-000000001001'::uuid and item_id = '00000000-0000-4000-8000-000000001101'::uuid and batch_id is null")"
  curl_check "available_stock_read" GET "$api_base/inventory/available-stock?warehouse_id=00000000-0000-4000-8000-000000000801&location_id=00000000-0000-4000-8000-000000001001&sku=FG-LIP-001" 200 "" auth
  if ! grep -q '"success"[[:space:]]*:[[:space:]]*true' "$tmp_body" ||
    ! grep -q '"sku":"FG-LIP-001"' "$tmp_body" ||
    ! grep -q "\"physical_qty\":\"$expected_on_hand\"" "$tmp_body" ||
    ! grep -q "\"available_qty\":\"$expected_available\"" "$tmp_body"; then
    echo "persisted_available_stock failed: expected physical=$expected_on_hand available=$expected_available" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_stock_adjustment" "ok" "$adjustment_no"
  printf '%-28s %s %s\n' "persisted_stock_movement" "ok" "$adjustment_no"
  printf '%-28s %s %s/%s\n' "persisted_available_stock" "ok" "$expected_on_hand" "$expected_available"
}

persisted_stock_count_check() {
  smoke_index="$(postgres_scalar "select count(*) + 1 from inventory.stock_count_sessions where count_ref like 'count-s10-03-03-smoke-%'")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_stock_count failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  count_id="count-s10-03-03-smoke-$suffix"
  line_id="count-line-s10-03-03-smoke-$suffix"
  count_no="CNT-S10-03-03-SMOKE-$suffix"
  before_count="$(postgres_scalar "select count(*) from inventory.stock_count_sessions where count_ref = '$count_id'")"

  body="$(printf '{"id":"%s","count_no":"%s","org_id":"00000000-0000-4000-8000-000000000001","warehouse_id":"00000000-0000-4000-8000-000000000801","warehouse_code":"warehouse_main","scope":"cycle_count","lines":[{"id":"%s","item_id":"00000000-0000-4000-8000-000000001101","sku":"FG-LIP-001","location_id":"00000000-0000-4000-8000-000000001001","expected_qty":"20","base_uom_code":"PCS"}]}' "$count_id" "$count_no" "$line_id")"

  curl_check "stock_count_create" POST "$api_base/stock-counts" 201 "$body" auth
  curl_check "stock_count_submit" POST "$api_base/stock-counts/$count_id/submit" 200 "$(printf '{"lines":[{"id":"%s","counted_qty":"18","note":"S10-03-03 persisted stock count smoke"}]}' "$line_id")" auth

  document_count="$(postgres_scalar "select count(*) from inventory.stock_count_sessions s join inventory.stock_count_session_lines l on l.session_id = s.id where s.org_id = '00000000-0000-4000-8000-000000000001'::uuid and s.count_ref = '$count_id' and s.status = 'variance_review' and s.submitted_by_ref = 'user-erp-admin' and l.line_ref = '$line_id' and l.counted = true and l.delta_qty = -2.000000")"
  audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '00000000-0000-4000-8000-000000000001'::uuid and entity_ref = '$count_id' and action in ('inventory.stock_count.created', 'inventory.stock_count.submitted')")"
  if [ "$before_count" != "0" ] || [ "$document_count" != "1" ] || [ "$audit_count" != "2" ]; then
    echo "persisted_stock_count failed: before=$before_count document=$document_count audit=$audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_stock_count" "ok" "$count_no"
}

persisted_inbound_qc_check() {
  smoke_index="$(postgres_scalar "select greatest((select count(*) from qc.inbound_qc_inspections where goods_receipt_ref like '00000000-0000-4000-8000-00000018%'), (select count(*) from purchase.purchase_orders where po_ref like 'po-s11-03-03-smoke-%')) + 1")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_inbound_qc failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  org_id="00000000-0000-4000-8000-000000000001"
  warehouse_id="00000000-0000-4000-8000-000000000801"
  location_id="00000000-0000-4000-8000-000000001001"
  item_id="00000000-0000-4000-8000-000000001102"
  purchase_order_id="po-s11-03-03-smoke-$suffix"
  purchase_order_line_id="po-line-s11-03-03-smoke-$suffix"
  purchase_order_no="PO-S11-03-03-SMOKE-$suffix"
  receipt_id="00000000-0000-4000-8000-00000018$suffix"
  line_id="00000000-0000-4000-8000-00000020$suffix"
  inspection_id="00000000-0000-4000-8000-00000019$suffix"
  receipt_no="GRN-S10-04-03-SMOKE-$suffix"

  purchase_body="$(printf '{"id":"%s","po_no":"%s","supplier_id":"sup-rm-bioactive","warehouse_id":"wh-hcm-rm","expected_date":"2026-05-08","currency_code":"VND","lines":[{"id":"%s","line_no":1,"item_id":"item-serum-30ml","ordered_qty":"12","uom_code":"EA","unit_price":"125000"}]}' "$purchase_order_id" "$purchase_order_no" "$purchase_order_line_id")"
  curl_check "purchase_order_create" POST "$api_base/purchase-orders" 201 "$purchase_body" auth
  curl_check "purchase_order_submit" POST "$api_base/purchase-orders/$purchase_order_id/submit" 200 '{"expected_version":1}' auth
  curl_check "purchase_order_approve" POST "$api_base/purchase-orders/$purchase_order_id/approve" 200 '{"expected_version":2}' auth
  curl_check "purchase_order_read" GET "$api_base/purchase-orders/$purchase_order_id" 200 "" auth
  if ! grep -q '"success"[[:space:]]*:[[:space:]]*true' "$tmp_body" ||
    ! grep -q "\"id\":\"$purchase_order_id\"" "$tmp_body" ||
    ! grep -q '"status":"approved"' "$tmp_body"; then
    echo "persisted_purchase_order failed: approved purchase order response mismatch" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  purchase_order_count="$(postgres_scalar "select count(*) from purchase.purchase_orders p join purchase.purchase_order_lines l on l.purchase_order_id = p.id where p.org_id = '$org_id'::uuid and p.po_ref = '$purchase_order_id' and p.po_no = '$purchase_order_no' and p.status = 'approved' and p.submitted_by_ref = 'user-erp-admin' and p.approved_by_ref = 'user-erp-admin' and l.line_ref = '$purchase_order_line_id' and l.item_ref = 'item-serum-30ml' and l.ordered_qty = 12.000000 and l.uom_code = 'EA'")"
  purchase_order_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$purchase_order_id' and action in ('purchase.order.created', 'purchase.order.submitted', 'purchase.order.approved')")"
  if [ "$purchase_order_count" != "1" ] || [ "$purchase_order_audit_count" != "3" ]; then
    echo "persisted_purchase_order failed: document=$purchase_order_count audit=$purchase_order_audit_count" >&2
    exit 1
  fi

  postgres_exec "INSERT INTO inventory.warehouse_receivings (id, org_id, receipt_ref, receipt_no, org_ref, warehouse_id, warehouse_ref, warehouse_code, location_id, location_ref, location_code, reference_doc_type, reference_doc_ref, supplier_ref, delivery_note_no, status, created_by_ref, submitted_at, submitted_by_ref, inspect_ready_at, inspect_ready_by_ref, created_at, updated_at) VALUES ('$receipt_id'::uuid, '$org_id'::uuid, '$receipt_id', '$receipt_no', '$org_id', '$warehouse_id'::uuid, '$warehouse_id', 'warehouse_main', '$location_id'::uuid, '$location_id', 'A-01', 'purchase_order', '$purchase_order_id', 'sup-rm-bioactive', 'DN-S10-04-03-$suffix', 'inspect_ready', 'user-erp-admin', now(), 'user-erp-admin', now(), 'user-qa', now(), now()) ON CONFLICT ON CONSTRAINT uq_warehouse_receivings_org_ref DO NOTHING; INSERT INTO inventory.warehouse_receiving_lines (id, org_id, receipt_id, line_ref, line_no, purchase_order_line_ref, item_id, item_ref, sku_code, item_name, batch_ref, batch_no, lot_no, expiry_date, warehouse_id, warehouse_ref, location_id, location_ref, quantity, uom_code, base_uom_code, packaging_status, qc_status, created_at, updated_at) VALUES ('$line_id'::uuid, '$org_id'::uuid, '$receipt_id'::uuid, '$line_id', 1, '$purchase_order_line_id', '$item_id'::uuid, '$item_id', 'FG-SER-001', 'Vitamin C Serum', 'batch-serum-2604a', 'LOT-2604A', 'LOT-2604A', '2027-04-01', '$warehouse_id'::uuid, '$warehouse_id', '$location_id'::uuid, '$location_id', 12.000000, 'PCS', 'PCS', 'intact', 'hold', now(), now()) ON CONFLICT ON CONSTRAINT uq_warehouse_receiving_lines_ref DO NOTHING;"

  create_body="$(printf '{"id":"%s","goods_receipt_id":"%s","goods_receipt_line_id":"%s","inspector_id":"user-qa","note":"S10-04-03 inbound QC persistence smoke"}' "$inspection_id" "$receipt_id" "$line_id")"
  curl_check "inbound_qc_create" POST "$api_base/inbound-qc-inspections" 201 "$create_body" auth
  curl_check "inbound_qc_start" POST "$api_base/inbound-qc-inspections/$inspection_id/start" 200 "" auth

  decision_body='{"passed_qty":"7","hold_qty":"5","reason":"S10-04-03 split hold smoke","checklist":[{"id":"check-packaging","code":"PACKAGING","label":"Packaging condition","required":true,"status":"pass"},{"id":"check-lot-expiry","code":"LOT_EXPIRY","label":"Lot and expiry match delivery","required":true,"status":"pass"},{"id":"check-sample","code":"SAMPLE","label":"Sample retained","required":false,"status":"not_applicable"}]}'
  curl_check "inbound_qc_partial" POST "$api_base/inbound-qc-inspections/$inspection_id/partial" 200 "$decision_body" auth

  document_count="$(postgres_scalar "select count(*) from qc.inbound_qc_inspections where org_id = '$org_id'::uuid and inspection_ref = '$inspection_id' and goods_receipt_ref = '$receipt_id' and goods_receipt_line_ref = '$line_id' and status = 'completed' and result = 'partial' and passed_qty = 7.000000 and hold_qty = 5.000000")"
  checklist_count="$(postgres_scalar "select count(*) from qc.inbound_qc_inspections i join qc.inbound_qc_checklist_items c on c.inspection_id = i.id where i.org_id = '$org_id'::uuid and i.inspection_ref = '$inspection_id' and c.status in ('pass', 'not_applicable')")"
  ledger_count="$(postgres_scalar "select count(*) from inventory.stock_ledger where org_id = '$org_id'::uuid and source_doc_type = 'inbound_qc_inspection' and source_doc_id = '$inspection_id'::uuid and ((stock_status = 'available' and movement_qty = 7.000000) or (stock_status = 'qc_hold' and movement_qty = 5.000000))")"
  audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$inspection_id' and action in ('qc.inbound_inspection.created', 'qc.inbound_inspection.started', 'qc.inbound_inspection.partial')")"
  trace_count="$(postgres_scalar "select count(*) from purchase.purchase_orders p join purchase.purchase_order_lines pol on pol.purchase_order_id = p.id join inventory.warehouse_receivings r on r.reference_doc_ref = p.po_ref join inventory.warehouse_receiving_lines rl on rl.receipt_id = r.id and rl.purchase_order_line_ref = pol.line_ref join qc.inbound_qc_inspections i on i.goods_receipt_ref = r.receipt_ref and i.goods_receipt_line_ref = rl.line_ref and i.purchase_order_ref = p.po_ref and i.purchase_order_line_ref = pol.line_ref where p.org_id = '$org_id'::uuid and p.po_ref = '$purchase_order_id' and r.receipt_ref = '$receipt_id' and i.inspection_ref = '$inspection_id'")"
  if [ "$document_count" != "1" ] || [ "$checklist_count" != "3" ] || [ "$ledger_count" != "2" ] || [ "$audit_count" != "3" ] || [ "$trace_count" != "1" ]; then
    echo "persisted_inbound_qc failed: document=$document_count checklist=$checklist_count ledger=$ledger_count audit=$audit_count trace=$trace_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_purchase_order" "ok" "$purchase_order_no"
  printf '%-28s %s %s\n' "persisted_inbound_qc" "ok" "$receipt_no"
}

persisted_carrier_manifest_check() {
  smoke_index="$(postgres_scalar "select greatest((select count(*) from shipping.carrier_manifests where manifest_ref like 'manifest-s14-02-01-smoke-%'), (select count(*) from shipping.shipments where shipment_no like 'SHP-S14-02-01-SMOKE-%'), (select count(*) from sales.sales_orders where order_ref like 'so-s14-02-01-smoke-%')) + 1")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_carrier_manifest failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  org_id="00000000-0000-4000-8000-000000000001"
  user_id="00000000-0000-4000-8000-000000000101"
  customer_id="00000000-0000-4000-8000-000000000601"
  carrier_id="00000000-0000-4000-8000-00000044$suffix"
  warehouse_id="00000000-0000-4000-8000-000000000801"
  item_id="00000000-0000-4000-8000-000000001101"
  unit_id="00000000-0000-4000-8000-000000000401"
  sales_order_uuid="00000000-0000-4000-8000-00000041$suffix"
  sales_order_line_uuid="00000000-0000-4000-8000-00000043$suffix"
  shipment_uuid="00000000-0000-4000-8000-00000042$suffix"
  sales_order_id="so-s14-02-01-smoke-$suffix"
  sales_order_line_id="so-line-s14-02-01-smoke-$suffix"
  sales_order_no="SO-S14-02-01-SMOKE-$suffix"
  shipment_no="SHP-S14-02-01-SMOKE-$suffix"
  tracking_no="TRK-S14-02-01-SMOKE-$suffix"
  manifest_id="manifest-s14-02-01-smoke-$suffix"

  before_count="$(postgres_scalar "select count(*) from shipping.carrier_manifests where org_id = '$org_id'::uuid and manifest_ref = '$manifest_id'")"
  postgres_exec "$(cat <<EOF
INSERT INTO mdm.carriers (
  id,
  org_id,
  code,
  name,
  contact_name,
  phone,
  status,
  created_by,
  updated_by
) VALUES (
  '$carrier_id'::uuid,
  '$org_id'::uuid,
  'GHN',
  'GHN Express',
  'GHN Dispatcher',
  '0900000014',
  'active',
  '$user_id'::uuid,
  '$user_id'::uuid
) ON CONFLICT (org_id, code) DO UPDATE
SET name = EXCLUDED.name,
    contact_name = EXCLUDED.contact_name,
    phone = EXCLUDED.phone,
    status = EXCLUDED.status,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;

INSERT INTO sales.sales_orders (
  id,
  org_id,
  order_ref,
  org_ref,
  order_no,
  customer_id,
  customer_ref,
  customer_code,
  customer_name,
  order_date,
  channel,
  status,
  currency_code,
  subtotal_amount,
  discount_amount,
  tax_amount,
  shipping_fee_amount,
  net_amount,
  total_amount,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  note,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref,
  confirmed_at,
  confirmed_by_ref,
  reserved_at,
  reserved_by_ref,
  picking_started_at,
  picking_started_by_ref,
  picked_at,
  picked_by_ref,
  packing_started_at,
  packing_started_by_ref,
  packed_at,
  packed_by_ref,
  waiting_handover_at,
  waiting_handover_by_ref,
  version
) VALUES (
  '$sales_order_uuid'::uuid,
  '$org_id'::uuid,
  '$sales_order_id',
  'org-my-pham',
  '$sales_order_no',
  '$customer_id'::uuid,
  'cus-local-001',
  'CUS-LOCAL-001',
  'Local Retail Customer',
  '2026-05-02',
  'internal',
  'waiting_handover',
  'VND',
  125000.00,
  0.00,
  0.00,
  0.00,
  125000.00,
  125000.00,
  '$warehouse_id'::uuid,
  'warehouse_main',
  'warehouse_main',
  'S14-02-01 manifest persistence smoke fixture',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  'user-erp-admin',
  now(),
  'user-erp-admin',
  now(),
  'user-erp-admin',
  now(),
  'user-erp-admin',
  now(),
  'user-erp-admin',
  now(),
  'user-erp-admin',
  now(),
  'user-erp-admin',
  1
) ON CONFLICT (org_id, order_ref) DO NOTHING;

INSERT INTO sales.sales_order_lines (
  id,
  org_id,
  sales_order_id,
  line_ref,
  line_no,
  item_id,
  item_ref,
  sku_code,
  item_name,
  unit_id,
  ordered_qty,
  reserved_qty,
  shipped_qty,
  unit_price,
  uom_code,
  base_ordered_qty,
  base_uom_code,
  conversion_factor,
  currency_code,
  line_discount_amount,
  line_amount,
  created_at,
  created_by,
  updated_at,
  updated_by
) VALUES (
  '$sales_order_line_uuid'::uuid,
  '$org_id'::uuid,
  '$sales_order_uuid'::uuid,
  '$sales_order_line_id',
  1,
  '$item_id'::uuid,
  'item-lipstick-matte',
  'FG-LIP-001',
  'Matte Lipstick',
  '$unit_id'::uuid,
  1.000000,
  1.000000,
  1.000000,
  125000.0000,
  'PCS',
  1.000000,
  'PCS',
  1.000000,
  'VND',
  0.00,
  125000.00,
  now(),
  '$user_id'::uuid,
  now(),
  '$user_id'::uuid
) ON CONFLICT (sales_order_id, line_ref) DO NOTHING;

INSERT INTO shipping.shipments (
  id,
  org_id,
  shipment_no,
  sales_order_id,
  warehouse_id,
  carrier_id,
  tracking_no,
  status,
  packed_at,
  packed_by,
  created_at,
  created_by,
  updated_at,
  updated_by
) VALUES (
  '$shipment_uuid'::uuid,
  '$org_id'::uuid,
  '$shipment_no',
  '$sales_order_uuid'::uuid,
  '$warehouse_id'::uuid,
  (SELECT id FROM mdm.carriers WHERE org_id = '$org_id'::uuid AND code = 'GHN'),
  '$tracking_no',
  'packed',
  now(),
  '$user_id'::uuid,
  now(),
  '$user_id'::uuid,
  now(),
  '$user_id'::uuid
) ON CONFLICT (org_id, shipment_no) DO UPDATE
SET sales_order_id = EXCLUDED.sales_order_id,
    warehouse_id = EXCLUDED.warehouse_id,
    carrier_id = EXCLUDED.carrier_id,
    tracking_no = EXCLUDED.tracking_no,
    status = EXCLUDED.status,
    packed_at = EXCLUDED.packed_at,
    packed_by = EXCLUDED.packed_by,
    updated_at = now(),
    updated_by = EXCLUDED.updated_by;
EOF
)"

  manifest_body="$(printf '{"id":"%s","carrier_code":"GHN","carrier_name":"GHN Express","warehouse_id":"warehouse_main","warehouse_code":"warehouse_main","date":"2026-05-02","handover_batch":"s14-smoke","staging_zone":"handover","owner":"user-erp-admin"}' "$manifest_id")"
  curl_check "carrier_manifest_create" POST "$api_base/shipping/manifests" 201 "$manifest_body" auth
  curl_check "carrier_manifest_add" POST "$api_base/shipping/manifests/$manifest_id/shipments" 200 "$(printf '{"shipment_id":"%s"}' "$shipment_no")" auth
  curl_check "carrier_manifest_ready" POST "$api_base/shipping/manifests/$manifest_id/ready" 200 '{}' auth
  curl_check "carrier_manifest_scan" POST "$api_base/shipping/manifests/$manifest_id/scan" 200 "$(printf '{"code":"%s","station_id":"dock-s14-02-01","device_id":"scanner-s14-02-01","source":"smoke"}' "$tracking_no")" auth
  if ! grep -q '"result_code":"MATCHED"' "$tmp_body" ||
    ! grep -q "\"id\":\"$manifest_id\"" "$tmp_body" ||
    ! grep -q "\"tracking_no\":\"$tracking_no\"" "$tmp_body"; then
    echo "persisted_carrier_manifest failed: scan response mismatch" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  curl_check "carrier_manifest_handover" POST "$api_base/shipping/manifests/$manifest_id/confirm-handover" 200 '{"reason":"S14-02-01 manifest persistence smoke"}' auth
  curl_check "carrier_manifest_read" GET "$api_base/shipping/manifests?warehouse_id=warehouse_main&date=2026-05-02&carrier_code=GHN&status=handed_over" 200 "" auth
  if ! grep -q "\"id\":\"$manifest_id\"" "$tmp_body" ||
    ! grep -q '"status":"handed_over"' "$tmp_body" ||
    ! grep -q "\"tracking_no\":\"$tracking_no\"" "$tmp_body"; then
    echo "persisted_carrier_manifest failed: handed-over manifest not queryable" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  document_count="$(postgres_scalar "select count(*) from shipping.carrier_manifests m join shipping.carrier_manifest_orders l on l.carrier_manifest_id = m.id where m.org_id = '$org_id'::uuid and m.manifest_ref = '$manifest_id' and m.status = 'handed_over' and m.expected_count = 1 and m.scanned_count = 1 and m.missing_count = 0 and l.manifest_ref = '$manifest_id' and l.shipment_ref = '$shipment_no' and l.order_no = '$sales_order_no' and l.tracking_no = '$tracking_no' and l.scan_status = 'scanned'")"
  scan_count="$(postgres_scalar "select count(*) from shipping.scan_events where org_id = '$org_id'::uuid and manifest_ref = '$manifest_id' and shipment_ref = '$shipment_no' and barcode = '$tracking_no' and scan_context = 'handover' and scan_result = 'matched' and actor_ref = 'user-erp-admin'")"
  sales_order_count="$(postgres_scalar "select count(*) from sales.sales_orders where org_id = '$org_id'::uuid and order_ref = '$sales_order_id' and order_no = '$sales_order_no' and status = 'handed_over' and handed_over_by_ref = 'user-erp-admin'")"
  manifest_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$manifest_id' and action in ('shipping.manifest.created', 'shipping.manifest.shipment_added', 'shipping.manifest.ready_to_scan', 'shipping.manifest.handed_over')")"
  scan_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and action = 'shipping.manifest.scan_recorded' and after_data->>'manifest_id' = '$manifest_id' and actor_ref = 'user-erp-admin'")"
  sales_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$sales_order_id' and action = 'sales.order.handed_over'")"
  if [ "$before_count" != "0" ] || [ "$document_count" != "1" ] || [ "$scan_count" != "1" ] || [ "$sales_order_count" != "1" ] || [ "$manifest_audit_count" != "4" ] || [ "$scan_audit_count" != "1" ] || [ "$sales_audit_count" != "1" ]; then
    echo "persisted_carrier_manifest failed: before=$before_count document=$document_count scan=$scan_count sales=$sales_order_count manifest_audit=$manifest_audit_count scan_audit=$scan_audit_count sales_audit=$sales_audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_carrier_manifest" "ok" "$manifest_id"
}

persisted_pick_pack_check() {
  smoke_index="$(postgres_scalar "select greatest((select count(*) from shipping.pick_tasks where pick_ref like 'pick-s14-02-02-smoke-%'), (select count(*) from shipping.pack_tasks where pack_ref like 'pack-s14-02-02-smoke-%'), (select count(*) from sales.sales_orders where order_ref like 'so-s14-02-02-smoke-%')) + 1")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_pick_pack failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  org_id="00000000-0000-4000-8000-000000000001"
  user_id="00000000-0000-4000-8000-000000000101"
  customer_id="00000000-0000-4000-8000-000000000601"
  unit_id="00000000-0000-4000-8000-000000000401"
  warehouse_id="00000000-0000-4000-8000-000000000801"
  bin_id="00000000-0000-4000-8000-000000001001"
  item_id="00000000-0000-4000-8000-000000001101"
  batch_id="00000000-0000-4000-8000-000000001201"
  sales_order_uuid="00000000-0000-4000-8000-00000051$suffix"
  sales_order_line_uuid="00000000-0000-4000-8000-00000052$suffix"
  reservation_uuid="00000000-0000-4000-8000-00000053$suffix"
  pick_task_uuid="00000000-0000-4000-8000-00000054$suffix"
  pick_line_uuid="00000000-0000-4000-8000-00000055$suffix"
  pack_task_uuid="00000000-0000-4000-8000-00000056$suffix"
  pack_line_uuid="00000000-0000-4000-8000-00000057$suffix"
  sales_order_id="so-s14-02-02-smoke-$suffix"
  sales_order_line_id="so-line-s14-02-02-smoke-$suffix"
  sales_order_no="SO-S14-02-02-SMOKE-$suffix"
  reservation_ref="rsv-s14-02-02-smoke-$suffix"
  pick_id="pick-s14-02-02-smoke-$suffix"
  pick_no="PICK-S14-02-02-SMOKE-$suffix"
  pick_line_id="pick-line-s14-02-02-smoke-$suffix"
  pack_id="pack-s14-02-02-smoke-$suffix"
  pack_no="PACK-S14-02-02-SMOKE-$suffix"
  pack_line_id="pack-line-s14-02-02-smoke-$suffix"

  before_pick_count="$(postgres_scalar "select count(*) from shipping.pick_tasks where org_id = '$org_id'::uuid and pick_ref = '$pick_id'")"
  before_pack_count="$(postgres_scalar "select count(*) from shipping.pack_tasks where org_id = '$org_id'::uuid and pack_ref = '$pack_id'")"
  postgres_exec "$(cat <<EOF
INSERT INTO sales.sales_orders (
  id,
  org_id,
  order_no,
  customer_id,
  order_date,
  channel,
  status,
  currency_code,
  total_amount,
  subtotal_amount,
  net_amount,
  order_ref,
  org_ref,
  customer_ref,
  customer_code,
  customer_name,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref,
  confirmed_at,
  confirmed_by,
  confirmed_by_ref,
  reserved_at,
  reserved_by,
  reserved_by_ref,
  picking_started_at,
  picking_started_by,
  picking_started_by_ref,
  picked_at,
  picked_by,
  picked_by_ref,
  packing_started_at,
  packing_started_by,
  packing_started_by_ref
) VALUES (
  '$sales_order_uuid'::uuid,
  '$org_id'::uuid,
  '$sales_order_no',
  '$customer_id'::uuid,
  '2026-05-02',
  'online',
  'packing',
  'VND',
  375000,
  375000,
  375000,
  '$sales_order_id',
  '$org_id',
  '$customer_id',
  '$customer_id',
  'Sprint 14 Smoke Customer',
  '$warehouse_id'::uuid,
  '$warehouse_id',
  'warehouse_main',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin'
) ON CONFLICT (org_id, order_no) DO NOTHING;

INSERT INTO sales.sales_order_lines (
  id,
  org_id,
  sales_order_id,
  line_ref,
  line_no,
  item_id,
  item_ref,
  sku_code,
  item_name,
  unit_id,
  ordered_qty,
  reserved_qty,
  unit_price,
  uom_code,
  base_ordered_qty,
  base_uom_code,
  conversion_factor,
  currency_code,
  line_amount,
  batch_id,
  batch_ref,
  batch_no,
  created_at,
  created_by,
  updated_at,
  updated_by
) VALUES (
  '$sales_order_line_uuid'::uuid,
  '$org_id'::uuid,
  '$sales_order_uuid'::uuid,
  '$sales_order_line_id',
  1,
  '$item_id'::uuid,
  'FG-LIP-001',
  'FG-LIP-001',
  'Matte Lipstick',
  '$unit_id'::uuid,
  3.000000,
  3.000000,
  125000,
  'PCS',
  3.000000,
  'PCS',
  1.000000,
  'VND',
  375000,
  '$batch_id'::uuid,
  '$batch_id',
  'BATCH-LIP-LOCAL',
  now(),
  '$user_id'::uuid,
  now(),
  '$user_id'::uuid
) ON CONFLICT (sales_order_id, line_no) DO NOTHING;

INSERT INTO inventory.stock_reservations (
  id,
  org_id,
  reservation_no,
  item_id,
  batch_id,
  warehouse_id,
  reserved_qty,
  source_doc_type,
  source_doc_id,
  status,
  created_at,
  created_by,
  sales_order_id,
  sales_order_line_id,
  source_doc_line_id,
  bin_id,
  base_uom_code,
  stock_status,
  updated_at,
  reservation_ref,
  org_ref,
  sales_order_ref,
  sales_order_line_ref,
  source_doc_ref,
  source_doc_line_ref,
  item_ref,
  sku_code,
  batch_ref,
  batch_no,
  warehouse_ref,
  warehouse_code,
  bin_ref,
  bin_code,
  created_by_ref
) VALUES (
  '$reservation_uuid'::uuid,
  '$org_id'::uuid,
  'RSV-S14-02-02-SMOKE-$suffix',
  '$item_id'::uuid,
  '$batch_id'::uuid,
  '$warehouse_id'::uuid,
  3.000000,
  'sales_order',
  '$sales_order_uuid'::uuid,
  'active',
  now(),
  '$user_id'::uuid,
  '$sales_order_uuid'::uuid,
  '$sales_order_line_uuid'::uuid,
  '$sales_order_line_uuid'::uuid,
  '$bin_id'::uuid,
  'PCS',
  'available',
  now(),
  '$reservation_ref',
  '$org_id',
  '$sales_order_id',
  '$sales_order_line_id',
  '$sales_order_id',
  '$sales_order_line_id',
  'FG-LIP-001',
  'FG-LIP-001',
  '$batch_id',
  'BATCH-LIP-LOCAL',
  'warehouse_main',
  'warehouse_main',
  'A-01',
  'A-01',
  'user-erp-admin'
) ON CONFLICT (org_id, reservation_no) DO NOTHING;

INSERT INTO shipping.pick_tasks (
  id,
  org_id,
  pick_ref,
  org_ref,
  pick_task_no,
  source_doc_type,
  sales_order_id,
  sales_order_ref,
  order_no,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  status,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  '$pick_task_uuid'::uuid,
  '$org_id'::uuid,
  '$pick_id',
  '$org_id',
  '$pick_no',
  'sales_order',
  '$sales_order_uuid'::uuid,
  '$sales_order_id',
  '$sales_order_no',
  '$warehouse_id'::uuid,
  'warehouse_main',
  'warehouse_main',
  'created',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin'
) ON CONFLICT (org_id, pick_ref) DO NOTHING;

INSERT INTO shipping.pick_task_lines (
  id,
  org_id,
  pick_task_id,
  line_ref,
  pick_task_ref,
  line_no,
  sales_order_line_id,
  sales_order_line_ref,
  stock_reservation_id,
  stock_reservation_ref,
  item_id,
  item_ref,
  sku_code,
  batch_id,
  batch_ref,
  batch_no,
  warehouse_id,
  warehouse_ref,
  bin_id,
  bin_ref,
  bin_code,
  base_uom_code,
  qty_to_pick,
  qty_picked,
  status,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  '$pick_line_uuid'::uuid,
  '$org_id'::uuid,
  '$pick_task_uuid'::uuid,
  '$pick_line_id',
  '$pick_id',
  1,
  '$sales_order_line_uuid'::uuid,
  '$sales_order_line_id',
  '$reservation_uuid'::uuid,
  '$reservation_ref',
  '$item_id'::uuid,
  'FG-LIP-001',
  'FG-LIP-001',
  '$batch_id'::uuid,
  '$batch_id',
  'BATCH-LIP-LOCAL',
  '$warehouse_id'::uuid,
  'warehouse_main',
  '$bin_id'::uuid,
  'A-01',
  'A-01',
  'PCS',
  3.000000,
  0.000000,
  'pending',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin'
) ON CONFLICT (org_id, pick_task_id, line_ref) DO NOTHING;
EOF
)"

  curl_check "pick_task_start" POST "$api_base/pick-tasks/$pick_id/start" 200 "" auth
  curl_check "pick_task_confirm_line" POST "$api_base/pick-tasks/$pick_id/confirm-line" 200 "$(printf '{"line_id":"%s","picked_qty":"3.000000"}' "$pick_line_id")" auth
  curl_check "pick_task_complete" POST "$api_base/pick-tasks/$pick_id/complete" 200 "" auth
  curl_check "pick_task_read" GET "$api_base/pick-tasks/$pick_id" 200 "" auth
  if ! grep -q "\"id\":\"$pick_id\"" "$tmp_body" ||
    ! grep -q '"status":"completed"' "$tmp_body" ||
    ! grep -q "\"id\":\"$pick_line_id\"" "$tmp_body" ||
    ! grep -q '"qty_picked":"3.000000"' "$tmp_body" ||
    ! grep -q '"picked_by":"user-erp-admin"' "$tmp_body"; then
    echo "persisted_pick_task failed: completed pick task response mismatch" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "pick_task_list" GET "$api_base/pick-tasks?warehouse_id=warehouse_main&status=completed&assigned_to=user-erp-admin" 200 "" auth
  if ! grep -q "\"id\":\"$pick_id\"" "$tmp_body"; then
    echo "persisted_pick_task failed: completed pick task not queryable" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  pick_document_count="$(postgres_scalar "select count(*) from shipping.pick_tasks t join shipping.pick_task_lines l on l.pick_task_id = t.id where t.org_id = '$org_id'::uuid and t.pick_ref = '$pick_id' and t.status = 'completed' and t.assigned_to_ref = 'user-erp-admin' and t.started_by_ref = 'user-erp-admin' and t.completed_by_ref = 'user-erp-admin' and l.line_ref = '$pick_line_id' and l.status = 'picked' and l.qty_picked = 3.000000 and l.picked_by_ref = 'user-erp-admin'")"
  pick_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$pick_id' and action in ('shipping.pick_task.started', 'shipping.pick_task.line_confirmed', 'shipping.pick_task.completed')")"
  if [ "$before_pick_count" != "0" ] || [ "$pick_document_count" != "1" ] || [ "$pick_audit_count" != "3" ]; then
    echo "persisted_pick_task failed: before=$before_pick_count document=$pick_document_count audit=$pick_audit_count" >&2
    exit 1
  fi

  postgres_exec "$(cat <<EOF
INSERT INTO shipping.pack_tasks (
  id,
  org_id,
  pack_ref,
  org_ref,
  pack_task_no,
  source_doc_type,
  sales_order_id,
  sales_order_ref,
  order_no,
  pick_task_id,
  pick_task_ref,
  pick_task_no,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  status,
  assigned_to,
  assigned_to_ref,
  assigned_at,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  '$pack_task_uuid'::uuid,
  '$org_id'::uuid,
  '$pack_id',
  '$org_id',
  '$pack_no',
  'sales_order',
  '$sales_order_uuid'::uuid,
  '$sales_order_id',
  '$sales_order_no',
  '$pick_task_uuid'::uuid,
  '$pick_id',
  '$pick_no',
  '$warehouse_id'::uuid,
  'warehouse_main',
  'warehouse_main',
  'created',
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin'
) ON CONFLICT (org_id, pack_ref) DO NOTHING;

INSERT INTO shipping.pack_task_lines (
  id,
  org_id,
  pack_task_id,
  line_ref,
  pack_task_ref,
  line_no,
  pick_task_line_id,
  pick_task_line_ref,
  sales_order_line_id,
  sales_order_line_ref,
  item_id,
  item_ref,
  sku_code,
  batch_id,
  batch_ref,
  batch_no,
  warehouse_id,
  warehouse_ref,
  base_uom_code,
  qty_to_pack,
  qty_packed,
  status,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  '$pack_line_uuid'::uuid,
  '$org_id'::uuid,
  '$pack_task_uuid'::uuid,
  '$pack_line_id',
  '$pack_id',
  1,
  (select id from shipping.pick_task_lines where org_id = '$org_id'::uuid and line_ref = '$pick_line_id' limit 1),
  '$pick_line_id',
  '$sales_order_line_uuid'::uuid,
  '$sales_order_line_id',
  '$item_id'::uuid,
  'FG-LIP-001',
  'FG-LIP-001',
  '$batch_id'::uuid,
  '$batch_id',
  'BATCH-LIP-LOCAL',
  '$warehouse_id'::uuid,
  'warehouse_main',
  'PCS',
  3.000000,
  0.000000,
  'pending',
  now(),
  '$user_id'::uuid,
  'user-erp-admin',
  now(),
  '$user_id'::uuid,
  'user-erp-admin'
) ON CONFLICT (org_id, pack_task_id, line_ref) DO NOTHING;
EOF
)"

  curl_check "pack_task_start" POST "$api_base/pack-tasks/$pack_id/start" 200 "" auth
  curl_check "pack_task_confirm" POST "$api_base/pack-tasks/$pack_id/confirm" 200 "$(printf '{"lines":[{"line_id":"%s","packed_qty":"3.000000"}]}' "$pack_line_id")" auth
  if ! grep -q "\"id\":\"$pack_id\"" "$tmp_body" ||
    ! grep -q '"status":"packed"' "$tmp_body" ||
    ! grep -q '"sales_order_status":"packed"' "$tmp_body" ||
    ! grep -q "\"id\":\"$pack_line_id\"" "$tmp_body" ||
    ! grep -q '"qty_packed":"3.000000"' "$tmp_body" ||
    ! grep -q '"packed_by":"user-erp-admin"' "$tmp_body"; then
    echo "persisted_pack_task failed: confirmed pack task response mismatch" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "pack_task_read" GET "$api_base/pack-tasks/$pack_id" 200 "" auth
  if ! grep -q "\"id\":\"$pack_id\"" "$tmp_body" ||
    ! grep -q '"status":"packed"' "$tmp_body" ||
    ! grep -q "\"id\":\"$pack_line_id\"" "$tmp_body"; then
    echo "persisted_pack_task failed: packed task not readable" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "pack_task_list" GET "$api_base/pack-tasks?warehouse_id=warehouse_main&status=packed&assigned_to=user-erp-admin" 200 "" auth
  if ! grep -q "\"id\":\"$pack_id\"" "$tmp_body"; then
    echo "persisted_pack_task failed: packed task not queryable" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  pack_document_count="$(postgres_scalar "select count(*) from shipping.pack_tasks t join shipping.pack_task_lines l on l.pack_task_id = t.id where t.org_id = '$org_id'::uuid and t.pack_ref = '$pack_id' and t.status = 'packed' and t.assigned_to_ref = 'user-erp-admin' and t.started_by_ref = 'user-erp-admin' and t.packed_by_ref = 'user-erp-admin' and l.line_ref = '$pack_line_id' and l.status = 'packed' and l.qty_packed = 3.000000 and l.packed_by_ref = 'user-erp-admin'")"
  sales_pack_count="$(postgres_scalar "select count(*) from sales.sales_orders o join sales.sales_order_lines l on l.sales_order_id = o.id where o.org_id = '$org_id'::uuid and o.order_ref = '$sales_order_id' and o.status = 'packed' and o.packed_by_ref = 'user-erp-admin' and l.id = '$sales_order_line_uuid'::uuid and l.line_ref = '$sales_order_line_id'")"
  pack_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$pack_id' and action in ('shipping.pack_task.started', 'shipping.pack_task.confirmed')")"
  sales_pack_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$sales_order_id' and action = 'sales.order.packed'")"
  if [ "$before_pack_count" != "0" ] || [ "$pack_document_count" != "1" ] || [ "$sales_pack_count" != "1" ] || [ "$pack_audit_count" != "2" ] || [ "$sales_pack_audit_count" != "1" ]; then
    echo "persisted_pack_task failed: before=$before_pack_count document=$pack_document_count sales=$sales_pack_count pack_audit=$pack_audit_count sales_audit=$sales_pack_audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_pick_task" "ok" "$pick_no"
  printf '%-28s %s %s\n' "persisted_pack_task" "ok" "$pack_no"
}

persisted_return_rejection_check() {
  org_id="00000000-0000-4000-8000-000000000001"

  return_index="$(postgres_scalar "select count(*) + 1 from returns.return_orders where return_ref like 'rr-unknown-s11-04-04-return-%'")"
  case "$return_index" in
    ''|*[!0-9]*)
      echo "persisted_return_receipt failed: invalid smoke index '$return_index'" >&2
      exit 1
      ;;
  esac

  return_suffix="$(printf '%04d' "$return_index")"
  return_scan="S11-04-04-RETURN-$return_suffix"
  return_id="rr-unknown-s11-04-04-return-$return_suffix"
  return_inspection_id="inspect-$return_id-damaged"
  return_action_id="dispose-$return_id-not_reusable"
  return_body="$(printf '{"warehouse_id":"wh-hcm-return","warehouse_code":"WH-HCM-RETURN","source":"CARRIER","code":"%s","package_condition":"damaged box","disposition":"needs_inspection","investigation_note":"S11-04-04 return persistence smoke"}' "$return_scan")"
  return_inspection_body='{"condition":"damaged","disposition":"not_reusable","note":"S11-04-04 return inspection persistence smoke","evidence_label":"photo-s11-04-04"}'
  return_disposition_body='{"disposition":"not_reusable","note":"S11-04-04 return disposition persistence smoke"}'

  curl_check "return_receipt_scan" POST "$api_base/returns/scan" 201 "$return_body" auth
  curl_check "return_receipt_inspect" POST "$api_base/returns/$return_id/inspect" 200 "$return_inspection_body" auth
  curl_check "return_receipt_dispose" POST "$api_base/returns/$return_id/disposition" 200 "$return_disposition_body" auth
  curl_check "return_receipt_read" GET "$api_base/returns/receipts?warehouse_id=wh-hcm-return&status=dispositioned" 200 "" auth
  if ! grep -q "\"id\":\"$return_id\"" "$tmp_body" ||
    ! grep -q '"status":"dispositioned"' "$tmp_body" ||
    ! grep -q '"disposition":"not_reusable"' "$tmp_body"; then
    echo "persisted_return_receipt failed: dispositioned return not queryable" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  return_document_count="$(postgres_scalar "select count(*) from returns.return_orders r join returns.return_order_lines l on l.return_order_id = r.id where r.org_id = '$org_id'::uuid and r.return_ref = '$return_id' and r.status = 'dispositioned' and r.disposition = 'not_reusable' and r.scan_code = '$return_scan' and r.target_location = 'lab-damaged-placeholder' and l.line_ref = 'line-unknown-return' and l.sku_code = 'UNKNOWN-SKU' and l.quantity = 1.000000")"
  return_inspection_count="$(postgres_scalar "select count(*) from returns.return_inspections where org_id = '$org_id'::uuid and inspection_ref = '$return_inspection_id' and return_ref = '$return_id' and condition_code = 'damaged' and disposition = 'not_reusable' and status = 'inspection_recorded'")"
  return_action_count="$(postgres_scalar "select count(*) from returns.return_disposition_actions where org_id = '$org_id'::uuid and action_ref = '$return_action_id' and return_ref = '$return_id' and disposition = 'not_reusable' and target_location = 'lab-damaged-placeholder' and target_stock_status = 'damaged' and action_code = 'route_to_lab_or_damaged'")"
  return_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$return_id' and action in ('returns.receipt.created', 'returns.receipt.inspected', 'returns.inspection.disposition')")"
  if [ "$return_document_count" != "1" ] || [ "$return_inspection_count" != "1" ] || [ "$return_action_count" != "1" ] || [ "$return_audit_count" != "3" ]; then
    echo "persisted_return_receipt failed: document=$return_document_count inspection=$return_inspection_count action=$return_action_count audit=$return_audit_count" >&2
    exit 1
  fi

  rejection_index="$(postgres_scalar "select count(*) + 1 from inventory.supplier_rejections where rejection_ref like 'srj-s11-04-04-smoke-%'")"
  case "$rejection_index" in
    ''|*[!0-9]*)
      echo "persisted_supplier_rejection failed: invalid smoke index '$rejection_index'" >&2
      exit 1
      ;;
  esac

  rejection_suffix="$(printf '%04d' "$rejection_index")"
  rejection_id="srj-s11-04-04-smoke-$rejection_suffix"
  rejection_no="SRJ-S11-04-04-SMOKE-$rejection_suffix"
  rejection_line_id="srj-line-s11-04-04-smoke-$rejection_suffix"
  rejection_attachment_id="srj-att-s11-04-04-smoke-$rejection_suffix"
  purchase_order_id="po-s11-04-04-smoke-$rejection_suffix"
  purchase_order_line_id="po-line-s11-04-04-smoke-$rejection_suffix"
  goods_receipt_id="grn-s11-04-04-smoke-$rejection_suffix"
  goods_receipt_line_id="grn-line-s11-04-04-smoke-$rejection_suffix"
  inbound_qc_id="iqc-s11-04-04-smoke-$rejection_suffix"
  supplier_body="$(cat <<EOF
{"id":"$rejection_id","org_id":"$org_id","rejection_no":"$rejection_no","supplier_id":"sup-rm-bioactive","supplier_code":"SUP-RM-BIOACTIVE","supplier_name":"Bioactive Supplier","purchase_order_id":"$purchase_order_id","purchase_order_no":"PO-S11-04-04-SMOKE-$rejection_suffix","goods_receipt_id":"$goods_receipt_id","goods_receipt_no":"GRN-S11-04-04-SMOKE-$rejection_suffix","inbound_qc_inspection_id":"$inbound_qc_id","warehouse_id":"wh-hcm-rm","warehouse_code":"WH-HCM-RM","reason":"S11-04-04 supplier rejection persistence smoke","lines":[{"id":"$rejection_line_id","purchase_order_line_id":"$purchase_order_line_id","goods_receipt_line_id":"$goods_receipt_line_id","inbound_qc_inspection_id":"$inbound_qc_id","item_id":"item-serum-30ml","sku":"SERUM-30ML","item_name":"Vitamin C Serum","batch_id":"batch-serum-2604a","batch_no":"LOT-2604A","lot_no":"LOT-2604A","expiry_date":"2027-04-01","rejected_qty":"6.000000","uom_code":"PCS","base_uom_code":"PCS","reason":"damaged packaging"}],"attachments":[{"id":"$rejection_attachment_id","line_id":"$rejection_line_id","file_name":"s11-04-04-qc-photo.jpg","object_key":"smoke/s11-04-04/$rejection_attachment_id.jpg","content_type":"image/jpeg","source":"qc_photo"}]}
EOF
)"

  curl_check "supplier_rejection_create" POST "$api_base/supplier-rejections" 201 "$supplier_body" auth
  curl_check "supplier_rejection_submit" POST "$api_base/supplier-rejections/$rejection_id/submit" 200 "" auth
  curl_check "supplier_rejection_confirm" POST "$api_base/supplier-rejections/$rejection_id/confirm" 200 "" auth
  curl_check "supplier_rejection_read" GET "$api_base/supplier-rejections/$rejection_id" 200 "" auth
  if ! grep -q "\"id\":\"$rejection_id\"" "$tmp_body" ||
    ! grep -q '"status":"confirmed"' "$tmp_body" ||
    ! grep -q "\"id\":\"$rejection_attachment_id\"" "$tmp_body"; then
    echo "persisted_supplier_rejection failed: confirmed supplier rejection not queryable" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  supplier_document_count="$(postgres_scalar "select count(*) from inventory.supplier_rejections r join inventory.supplier_rejection_lines l on l.rejection_id = r.id join inventory.supplier_rejection_attachments a on a.rejection_id = r.id where r.org_id = '$org_id'::uuid and r.rejection_ref = '$rejection_id' and r.rejection_no = '$rejection_no' and r.status = 'confirmed' and r.submitted_by_ref = 'user-erp-admin' and r.confirmed_by_ref = 'user-erp-admin' and l.line_ref = '$rejection_line_id' and l.goods_receipt_line_ref = '$goods_receipt_line_id' and l.inbound_qc_inspection_ref = '$inbound_qc_id' and l.sku_code = 'SERUM-30ML' and l.rejected_qty = 6.000000 and a.attachment_ref = '$rejection_attachment_id' and a.line_ref = '$rejection_line_id'")"
  supplier_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$rejection_id' and action in ('inventory.supplier_rejection.created', 'inventory.supplier_rejection.submitted', 'inventory.supplier_rejection.confirmed')")"
  if [ "$supplier_document_count" != "1" ] || [ "$supplier_audit_count" != "3" ]; then
    echo "persisted_supplier_rejection failed: document=$supplier_document_count audit=$supplier_audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_return_receipt" "ok" "$return_id"
  printf '%-28s %s %s\n' "persisted_supplier_rejection" "ok" "$rejection_no"
}

persisted_subcontract_runtime_check() {
  org_id="00000000-0000-4000-8000-000000000001"
  smoke_index="$(postgres_scalar "select count(*) + 1 from subcontract.subcontract_orders where order_ref like 'sco-s16-08-03-smoke-%'")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_subcontract_runtime failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  order_id="sco-s16-08-03-smoke-$suffix"
  order_no="SCO-S16-08-03-SMOKE-$suffix"
  material_line_id="$order_id-material-01"
  transfer_id="smt-s16-08-03-smoke-$suffix"
  transfer_no="SMT-S16-08-03-SMOKE-$suffix"
  transfer_evidence_id="$transfer_id-handover"
  milestone_id="spm-s16-08-03-smoke-$suffix"
  milestone_no="SPM-S16-08-03-SMOKE-$suffix"
  sample_id="sample-s16-08-03-smoke-$suffix"
  sample_code="SAMPLE-S16-08-03-SMOKE-$suffix"
  receipt_id="sfgr-s16-08-03-smoke-$suffix"
  receipt_no="SFGR-S16-08-03-SMOKE-$suffix"
  claim_id="sfc-s16-08-03-smoke-$suffix"
  claim_no="SFC-S16-08-03-SMOKE-$suffix"
  before_count="$(postgres_scalar "select count(*) from subcontract.subcontract_orders where org_id = '$org_id'::uuid and order_ref = '$order_id'")"

  order_body="$(cat <<EOF
{"id":"$order_id","order_no":"$order_no","factory_id":"sup-out-lotus","finished_item_id":"item-serum-30ml","planned_qty":"100","uom_code":"EA","currency_code":"VND","spec_summary":"S16-08-03 subcontract persistence smoke","sample_required":true,"claim_window_days":7,"target_start_date":"2026-04-29","expected_receipt_date":"2026-04-29","material_lines":[{"id":"$material_line_id","item_id":"item-cream-50g","planned_qty":"20","uom_code":"EA","unit_cost":"58000","lot_trace_required":true}]}
EOF
)"
  deposit_body="$(cat <<EOF
{"expected_version":4,"milestone_id":"$milestone_id","milestone_no":"$milestone_no","amount":"250000","recorded_by":"finance-user","recorded_at":"2026-04-29T08:30:00Z","note":"S16-08-03 deposit persistence smoke"}
EOF
)"
  issue_body="$(cat <<EOF
{"expected_version":5,"transfer_id":"$transfer_id","transfer_no":"$transfer_no","source_warehouse_id":"wh-hcm-rm","source_warehouse_code":"WH-HCM-RM","handover_by":"warehouse-user","handover_at":"2026-04-29T09:30:00Z","received_by":"factory-receiver","receiver_contact":"0988000111","vehicle_no":"51A-12345","lines":[{"order_material_line_id":"$material_line_id","issue_qty":"20","uom_code":"EA","batch_id":"batch-cream-2603b","source_bin_id":"rm-a01"}],"evidence":[{"id":"$transfer_evidence_id","evidence_type":"handover","file_name":"handover.pdf","object_key":"subcontract/$transfer_id/handover.pdf"}]}
EOF
)"
  sample_submit_body="$(cat <<EOF
{"expected_version":6,"sample_approval_id":"$sample_id","sample_code":"$sample_code","formula_version":"FORMULA-2026.04","spec_version":"SPEC-2026.04","submitted_by":"factory-user","submitted_at":"2026-04-29T10:30:00Z","evidence":[{"evidence_type":"photo","file_name":"sample-front.jpg","object_key":"subcontract/$sample_id/sample-front.jpg"}]}
EOF
)"
  sample_approve_body="$(cat <<EOF
{"expected_version":7,"sample_approval_id":"$sample_id","reason":"Approved for S16-08-03 smoke","storage_status":"retained_in_qa_cabinet","approved_by":"qa-lead","approved_at":"2026-04-29T11:30:00Z"}
EOF
)"
  receipt_body="$(cat <<EOF
{"expected_version":9,"receipt_id":"$receipt_id","receipt_no":"$receipt_no","warehouse_id":"wh-hcm-fg","warehouse_code":"WH-HCM-FG","location_id":"loc-hcm-fg-qc","location_code":"FG-QC-01","delivery_note_no":"DN-S16-08-03-$suffix","received_by":"warehouse-user","received_at":"2026-04-29T14:00:00Z","lines":[{"line_no":1,"item_id":"item-serum-30ml","receive_qty":"80","uom_code":"EA","base_receive_qty":"80","base_uom_code":"EA","conversion_factor":"1","batch_id":"batch-fg-s16-08-03-$suffix","batch_no":"LOT-S16-08-03-$suffix","lot_no":"LOT-S16-08-03-$suffix","expiry_date":"2028-04-29","packaging_status":"intact"}],"evidence":[{"id":"$receipt_id-photo","evidence_type":"qc_photo","file_name":"qc-photo.jpg","object_key":"subcontract/$receipt_id/qc-photo.jpg"}]}
EOF
)"
  claim_body="$(cat <<EOF
{"expected_version":10,"claim_id":"$claim_id","claim_no":"$claim_no","receipt_id":"$receipt_id","receipt_no":"$receipt_no","reason_code":"PACKAGING_DAMAGED","reason":"S16-08-03 factory claim persistence smoke","severity":"P1","affected_qty":"80","uom_code":"EA","base_affected_qty":"80","base_uom_code":"EA","owner_id":"qa-lead","opened_by":"qa-lead","opened_at":"2026-04-29T15:00:00Z","evidence":[{"id":"$claim_id-photo","evidence_type":"photo","file_name":"damaged-carton.jpg","object_key":"subcontract/$claim_id/damaged-carton.jpg"}]}
EOF
)"

  curl_check "subcontract_create" POST "$api_base/subcontract-orders" 201 "$order_body" auth
  curl_check "subcontract_submit" POST "$api_base/subcontract-orders/$order_id/submit" 200 '{"expected_version":1}' auth
  curl_check "subcontract_approve" POST "$api_base/subcontract-orders/$order_id/approve" 200 '{"expected_version":2}' auth
  curl_check "subcontract_confirm" POST "$api_base/subcontract-orders/$order_id/confirm-factory" 200 '{"expected_version":3}' auth
  curl_check "subcontract_deposit" POST "$api_base/subcontract-orders/$order_id/record-deposit" 200 "$deposit_body" auth
  curl_check "subcontract_issue" POST "$api_base/subcontract-orders/$order_id/issue-materials" 200 "$issue_body" auth
  curl_check "subcontract_sample_submit" POST "$api_base/subcontract-orders/$order_id/submit-sample" 200 "$sample_submit_body" auth
  curl_check "subcontract_sample_approve" POST "$api_base/subcontract-orders/$order_id/approve-sample" 200 "$sample_approve_body" auth
  curl_check "subcontract_mass_start" POST "$api_base/subcontract-orders/$order_id/start-mass-production" 200 '{"expected_version":8}' auth
  curl_check "subcontract_receive" POST "$api_base/subcontract-orders/$order_id/receive-finished-goods" 200 "$receipt_body" auth
  curl_check "subcontract_claim" POST "$api_base/subcontract-orders/$order_id/report-factory-defect" 200 "$claim_body" auth

  restart_api_service

  curl_check "subcontract_order_after" GET "$api_base/subcontract-orders/$order_id" 200 "" auth
  if ! grep -q "\"id\":\"$order_id\"" "$tmp_body" ||
    ! grep -q '"status":"rejected_with_factory_issue"' "$tmp_body" ||
    ! grep -q '"deposit_amount":"250000.00"' "$tmp_body" ||
    ! grep -q '"received_qty":"80.000000"' "$tmp_body"; then
    echo "persisted_subcontract_order failed: order not readable after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  json_check "warehouse_subcontract_after" "$api_base/warehouse/daily-board/subcontract-metrics?business_date=2026-04-29&warehouse_id=wh-hcm-rm"

  order_count="$(postgres_scalar "select count(*) from subcontract.subcontract_orders where org_id = '$org_id'::uuid and order_ref = '$order_id' and status = 'rejected_with_factory_issue' and deposit_amount = 250000.00 and received_qty = 80.000000")"
  transfer_count="$(postgres_scalar "select count(*) from subcontract.subcontract_material_transfers t join subcontract.subcontract_material_transfer_evidence e on e.material_transfer_id = t.id where t.org_id = '$org_id'::uuid and t.transfer_ref = '$transfer_id' and t.subcontract_order_ref = '$order_id' and t.status = 'sent_to_factory' and e.evidence_ref = '$transfer_evidence_id'")"
  sample_count="$(postgres_scalar "select count(*) from subcontract.subcontract_sample_approvals where org_id = '$org_id'::uuid and sample_ref = '$sample_id' and subcontract_order_ref = '$order_id' and status = 'approved' and storage_status = 'retained_in_qa_cabinet'")"
  receipt_count="$(postgres_scalar "select count(*) from subcontract.subcontract_finished_goods_receipts r join subcontract.subcontract_finished_goods_receipt_lines l on l.finished_goods_receipt_id = r.id where r.org_id = '$org_id'::uuid and r.receipt_ref = '$receipt_id' and r.subcontract_order_ref = '$order_id' and l.receive_qty = 80.000000 and l.packaging_status = 'intact'")"
  claim_count="$(postgres_scalar "select count(*) from subcontract.subcontract_factory_claims c join subcontract.subcontract_factory_claim_evidence e on e.factory_claim_id = c.id where c.org_id = '$org_id'::uuid and c.claim_ref = '$claim_id' and c.subcontract_order_ref = '$order_id' and c.receipt_ref = '$receipt_id' and c.status = 'open' and c.affected_qty = 80.000000 and e.evidence_ref = '$claim_id-photo'")"
  milestone_count="$(postgres_scalar "select count(*) from subcontract.subcontract_payment_milestones where org_id = '$org_id'::uuid and milestone_ref = '$milestone_id' and subcontract_order_ref = '$order_id' and kind = 'deposit' and status = 'recorded' and amount = 250000.00")"
  audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '$org_id'::uuid and entity_ref = '$order_id' and action in ('subcontract.order.created', 'subcontract.order.submitted', 'subcontract.order.approved', 'subcontract.order.factory_confirmed', 'subcontract.deposit_recorded', 'subcontract.materials_issued', 'subcontract.sample_submitted', 'subcontract.sample_approved', 'subcontract.order.mass_production_started', 'subcontract.finished_goods_received', 'subcontract.factory_claim_opened')")"
  if [ "$before_count" != "0" ] || [ "$order_count" != "1" ] || [ "$transfer_count" != "1" ] || [ "$sample_count" != "1" ] || [ "$receipt_count" != "1" ] || [ "$claim_count" != "1" ] || [ "$milestone_count" != "1" ] || [ "$audit_count" -lt "11" ]; then
    echo "persisted_subcontract_runtime failed: before=$before_count order=$order_count transfer=$transfer_count sample=$sample_count receipt=$receipt_count claim=$claim_count milestone=$milestone_count audit=$audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_subcontract_order" "ok" "$order_no"
  printf '%-28s %s %s\n' "persisted_subcontract_flow" "ok" "$transfer_no/$receipt_no/$claim_no"
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

persisted_auth_session_check() {
  previous_access_token="$access_token"

  curl_check "auth_session_login" POST "$api_base/auth/login" 200 "$login_body" noauth
  session_access_token="$(json_string_field "access_token")"
  session_refresh_token="$(json_string_field "refresh_token")"
  if [ "$session_access_token" = "" ] || [ "$session_refresh_token" = "" ]; then
    echo "persisted_auth_session failed: login tokens missing" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  access_token="$session_access_token"
  json_check "auth_me_before_restart" "$api_base/me"
  restart_api_service
  json_check "auth_me_after_restart" "$api_base/me"

  refresh_body="$(printf '{"refresh_token":"%s"}' "$(json_escape "$session_refresh_token")")"
  curl_check "auth_refresh_rotate" POST "$api_base/auth/refresh" 200 "$refresh_body" noauth
  rotated_access_token="$(json_string_field "access_token")"
  rotated_refresh_token="$(json_string_field "refresh_token")"
  if [ "$rotated_access_token" = "" ] || [ "$rotated_access_token" = "$session_access_token" ]; then
    echo "persisted_auth_session failed: refresh token did not rotate access token" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  if [ "$rotated_refresh_token" = "" ] || [ "$rotated_refresh_token" = "$session_refresh_token" ]; then
    echo "persisted_auth_session failed: refresh token did not rotate refresh token" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  curl_check "auth_old_refresh_reject" POST "$api_base/auth/refresh" 401 "$refresh_body" noauth

  access_token="$rotated_access_token"
  json_check "auth_me_after_refresh" "$api_base/me"
  logout_body="$(printf '{"refresh_token":"%s"}' "$(json_escape "$rotated_refresh_token")")"
  curl_check "auth_logout" POST "$api_base/auth/logout" 200 "$logout_body" noauth
  curl_check "auth_me_after_logout" GET "$api_base/me" 401 "" auth
  curl_check "auth_refresh_after_logout" POST "$api_base/auth/refresh" 401 "$logout_body" noauth

  lock_email="s18-smoke-lockout-$(date +%s)@example.local"
  lock_body="$(printf '{"email":"%s","password":"wrong-password!"}' "$(json_escape "$lock_email")")"
  for attempt in 1 2 3 4 5; do
    curl_check "auth_lockout_attempt_$attempt" POST "$api_base/auth/login" 401 "$lock_body" noauth
  done
  restart_api_service
  lock_check_body="$(printf '{"email":"%s","password":"%s"}' "$(json_escape "$lock_email")" "$(json_escape "$login_password")")"
  curl_check "auth_lockout_after_restart" POST "$api_base/auth/login" 401 "$lock_check_body" noauth
  if ! grep -q '"reason"[[:space:]]*:[[:space:]]*"locked"' "$tmp_body" &&
    ! grep -q 'Account temporarily locked' "$tmp_body"; then
    echo "persisted_auth_session failed: lockout reason missing after restart" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi
  postgres_exec "DELETE FROM core.auth_login_failures WHERE email_normalized = '$lock_email';"

  access_token="$previous_access_token"
  printf '%-28s %s %s\n' "persisted_auth_session" "ok" "access/refresh/lockout"
}

role_auth_check() {
  role_name="$1"
  role_email="$2"
  expected_role="$3"
  shift 3

  role_login_body="$(printf '{"email":"%s","password":"%s"}' "$(json_escape "$role_email")" "$(json_escape "$login_password")")"
  curl_check "auth_${role_name}_login" POST "$api_base/auth/login" 200 "$role_login_body" noauth
  role_access_token="$(json_string_field "access_token")"
  role_refresh_token="$(json_string_field "refresh_token")"
  if [ "$role_access_token" = "" ] || [ "$role_refresh_token" = "" ]; then
    echo "auth_${role_name} failed: tokens missing" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  previous_access_token="$access_token"
  access_token="$role_access_token"
  json_check "auth_${role_name}_me" "$api_base/me"
  if ! grep -q "\"email\"[[:space:]]*:[[:space:]]*\"$role_email\"" "$tmp_body" ||
    ! grep -q "\"role\"[[:space:]]*:[[:space:]]*\"$expected_role\"" "$tmp_body"; then
    echo "auth_${role_name} failed: /me did not return expected role" >&2
    sed -n '1,20p' "$tmp_body" >&2
    exit 1
  fi

  endpoint_index=1
  for role_endpoint in "$@"; do
    json_check "auth_${role_name}_route_$endpoint_index" "$role_endpoint"
    endpoint_index=$((endpoint_index + 1))
  done
  access_token="$previous_access_token"

  role_logout_body="$(printf '{"refresh_token":"%s"}' "$(json_escape "$role_refresh_token")")"
  curl_check "auth_${role_name}_logout" POST "$api_base/auth/logout" 200 "$role_logout_body" noauth
}

persisted_sales_reservation_check() {
  smoke_index="$(postgres_scalar "select count(*) + 1 from inventory.stock_reservations where reservation_ref like 'rsv-so-s10-02-03-smoke-%'")"
  case "$smoke_index" in
    ''|*[!0-9]*)
      echo "persisted_sales_reservation failed: invalid smoke index '$smoke_index'" >&2
      exit 1
      ;;
  esac

  suffix="$(printf '%04d' "$smoke_index")"
  order_id="so-s10-02-03-smoke-$suffix"
  line_id="line-s10-02-03-smoke-$suffix"
  order_no="SO-S10-02-03-SMOKE-$suffix"
  reservation_ref="rsv-$order_id-$line_id"
  before_count="$(postgres_scalar "select count(*) from inventory.stock_reservations where org_id = '00000000-0000-4000-8000-000000000001'::uuid and reservation_ref = '$reservation_ref'")"

  body="$(printf '{"id":"%s","order_no":"%s","customer_id":"cus-dl-minh-anh","channel":"B2B","warehouse_id":"wh-hcm-fg","order_date":"2026-05-01","currency_code":"VND","lines":[{"id":"%s","line_no":1,"item_id":"item-serum-30ml","ordered_qty":"1","uom_code":"EA","unit_price":"125000"}]}' "$order_id" "$order_no" "$line_id")"

  curl_check "sales_order_create" POST "$api_base/sales-orders" 201 "$body" auth
  curl_check "sales_order_confirm" POST "$api_base/sales-orders/$order_id/confirm" 200 '{"expected_version":1}' auth

  active_count="$(postgres_scalar "select count(*) from inventory.stock_reservations where org_id = '00000000-0000-4000-8000-000000000001'::uuid and reservation_ref = '$reservation_ref' and sales_order_ref = '$order_id' and sales_order_line_ref = '$line_id' and status = 'active' and created_by_ref = 'user-erp-admin' and base_uom_code = 'EA'")"
  if [ "$before_count" != "0" ] || [ "$active_count" != "1" ]; then
    echo "persisted_sales_reservation failed: reservation count before=$before_count active=$active_count" >&2
    exit 1
  fi

  curl_check "sales_order_cancel" POST "$api_base/sales-orders/$order_id/cancel" 200 '{"expected_version":3,"reason":"S10-02-03 reservation persistence smoke cleanup"}' auth

  released_count="$(postgres_scalar "select count(*) from inventory.stock_reservations where org_id = '00000000-0000-4000-8000-000000000001'::uuid and reservation_ref = '$reservation_ref' and status = 'released' and released_by_ref = 'user-erp-admin'")"
  audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '00000000-0000-4000-8000-000000000001'::uuid and entity_ref = '$reservation_ref' and action in ('inventory.stock_reservation.reserved', 'inventory.stock_reservation.released')")"
  if [ "$released_count" != "1" ] || [ "$audit_count" != "2" ]; then
    echo "persisted_sales_reservation failed: released=$released_count audit=$audit_count" >&2
    exit 1
  fi

  owner_count="$(postgres_scalar "select count(*) from sales.sales_orders where org_id = '00000000-0000-4000-8000-000000000001'::uuid and order_ref = '$order_id' and order_no = '$order_no' and status = 'cancelled' and created_by_ref = 'user-erp-admin' and cancelled_by_ref = 'user-erp-admin' and cancel_reason = 'S10-02-03 reservation persistence smoke cleanup'")"
  line_count="$(postgres_scalar "select count(*) from sales.sales_order_lines as line join sales.sales_orders as order_header on order_header.id = line.sales_order_id where order_header.org_id = '00000000-0000-4000-8000-000000000001'::uuid and order_header.order_ref = '$order_id' and line.line_ref = '$line_id' and line.item_ref = 'item-serum-30ml' and line.sku_code = 'SERUM-30ML' and line.ordered_qty = 1 and line.reserved_qty = 1 and line.uom_code = 'EA'")"
  order_audit_count="$(postgres_scalar "select count(*) from audit.audit_logs where org_id = '00000000-0000-4000-8000-000000000001'::uuid and entity_ref = '$order_id' and action in ('sales.order.created', 'sales.order.reserved', 'sales.order.cancelled')")"
  if [ "$owner_count" != "1" ] || [ "$line_count" != "1" ] || [ "$order_audit_count" != "3" ]; then
    echo "persisted_sales_order failed: owner=$owner_count line=$line_count audit=$order_audit_count" >&2
    exit 1
  fi

  printf '%-28s %s %s\n' "persisted_sales_reservation" "ok" "$order_no"
  printf '%-28s %s %s\n' "persisted_sales_order" "ok" "$order_no"
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
persisted_auth_session_check
role_auth_check "role_warehouse" "warehouse_user@example.local" "WAREHOUSE_STAFF" "$api_base/warehouse/daily-board/fulfillment-metrics?business_date=2026-04-30&warehouse_id=wh-hcm"
role_auth_check "role_sales" "sales_user@example.local" "SALES_OPS" "$api_base/inventory/available-stock" "$api_base/sales-orders"
role_auth_check "role_qc" "qc_user@example.local" "QA" "$api_base/inbound-qc-inspections" "$api_base/warehouse/daily-board/inbound-metrics?business_date=2026-04-30&warehouse_id=wh-hcm"

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

persisted_masterdata_runtime_check
persisted_finance_runtime_check
persisted_sales_reservation_check
persisted_stock_movement_check
persisted_stock_count_check
persisted_inbound_qc_check
persisted_carrier_manifest_check
persisted_pick_pack_check
persisted_return_rejection_check
persisted_subcontract_runtime_check

echo "Full ERP dev smoke passed"
