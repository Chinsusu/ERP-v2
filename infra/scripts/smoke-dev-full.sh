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

persisted_finance_runtime_check
persisted_sales_reservation_check
persisted_stock_movement_check
persisted_stock_count_check
persisted_inbound_qc_check
persisted_carrier_manifest_check
persisted_pick_pack_check
persisted_return_rejection_check

echo "Full ERP dev smoke passed"
