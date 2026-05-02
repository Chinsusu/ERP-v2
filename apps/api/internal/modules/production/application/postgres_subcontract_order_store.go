package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSubcontractOrderStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractOrderStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractOrderTx struct {
	store PostgresSubcontractOrderStore
	tx    *sql.Tx
}

type postgresSubcontractOrderQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractOrderHeader struct {
	PersistedID             string
	ID                      string
	OrgID                   string
	OrderNo                 string
	FactoryID               string
	FactoryCode             string
	FactoryName             string
	FinishedItemID          string
	FinishedSKUCode         string
	FinishedItemName        string
	PlannedQty              string
	ReceivedQty             string
	AcceptedQty             string
	RejectedQty             string
	UOMCode                 string
	BasePlannedQty          string
	BaseReceivedQty         string
	BaseAcceptedQty         string
	BaseRejectedQty         string
	BaseUOMCode             string
	ConversionFactor        string
	CurrencyCode            string
	EstimatedCostAmount     string
	DepositAmount           string
	SpecSummary             string
	SampleRequired          bool
	ClaimWindowDays         int
	TargetStartDate         string
	ExpectedReceiptDate     string
	Status                  string
	CreatedAt               time.Time
	CreatedBy               string
	UpdatedAt               time.Time
	UpdatedBy               string
	Version                 int
	CancelReason            string
	SampleRejectReason      string
	FactoryIssueReason      string
	SubmittedAt             sql.NullTime
	SubmittedBy             string
	ApprovedAt              sql.NullTime
	ApprovedBy              string
	FactoryConfirmedAt      sql.NullTime
	FactoryConfirmedBy      string
	DepositRecordedAt       sql.NullTime
	DepositRecordedBy       string
	MaterialsIssuedAt       sql.NullTime
	MaterialsIssuedBy       string
	SampleSubmittedAt       sql.NullTime
	SampleSubmittedBy       string
	SampleApprovedAt        sql.NullTime
	SampleApprovedBy        string
	SampleRejectedAt        sql.NullTime
	SampleRejectedBy        string
	MassProductionStartedAt sql.NullTime
	MassProductionStartedBy string
	FinishedGoodsReceivedAt sql.NullTime
	FinishedGoodsReceivedBy string
	QCStartedAt             sql.NullTime
	QCStartedBy             string
	AcceptedAt              sql.NullTime
	AcceptedBy              string
	RejectedFactoryIssueAt  sql.NullTime
	RejectedFactoryIssueBy  string
	FinalPaymentReadyAt     sql.NullTime
	FinalPaymentReadyBy     string
	ClosedAt                sql.NullTime
	ClosedBy                string
	CancelledAt             sql.NullTime
	CancelledBy             string
}

func NewPostgresSubcontractOrderStore(
	db *sql.DB,
	cfg PostgresSubcontractOrderStoreConfig,
) PostgresSubcontractOrderStore {
	return PostgresSubcontractOrderStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractOrderOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractOrderHeadersBaseSQL = `
SELECT
  subcontract_order.id::text,
  COALESCE(subcontract_order.order_ref, subcontract_order.id::text),
  COALESCE(subcontract_order.org_ref, subcontract_order.org_id::text),
  subcontract_order.order_no,
  COALESCE(subcontract_order.factory_ref, subcontract_order.factory_id::text, ''),
  COALESCE(subcontract_order.factory_code, factory.code, ''),
  COALESCE(subcontract_order.factory_name, factory.name, subcontract_order.factory_ref, subcontract_order.factory_id::text, ''),
  COALESCE(subcontract_order.finished_item_ref, subcontract_order.finished_item_id::text, ''),
  COALESCE(subcontract_order.finished_sku_code, finished_item.sku, subcontract_order.finished_item_ref, subcontract_order.finished_item_id::text, ''),
  COALESCE(subcontract_order.finished_item_name, finished_item.name, subcontract_order.finished_item_ref, subcontract_order.finished_item_id::text, ''),
  subcontract_order.planned_qty::text,
  subcontract_order.received_qty::text,
  subcontract_order.accepted_qty::text,
  subcontract_order.rejected_qty::text,
  subcontract_order.uom_code,
  subcontract_order.base_planned_qty::text,
  subcontract_order.base_received_qty::text,
  subcontract_order.base_accepted_qty::text,
  subcontract_order.base_rejected_qty::text,
  subcontract_order.base_uom_code,
  subcontract_order.conversion_factor::text,
  subcontract_order.currency_code,
  subcontract_order.estimated_cost_amount::text,
  subcontract_order.deposit_amount::text,
  COALESCE(subcontract_order.spec_summary, ''),
  subcontract_order.sample_required,
  subcontract_order.claim_window_days,
  COALESCE(subcontract_order.target_start_date::text, ''),
  COALESCE(subcontract_order.expected_receipt_date::text, ''),
  subcontract_order.status,
  subcontract_order.created_at,
  COALESCE(subcontract_order.created_by_ref, subcontract_order.created_by::text, ''),
  subcontract_order.updated_at,
  COALESCE(subcontract_order.updated_by_ref, subcontract_order.updated_by::text, ''),
  subcontract_order.version,
  COALESCE(subcontract_order.cancel_reason, ''),
  COALESCE(subcontract_order.sample_reject_reason, ''),
  COALESCE(subcontract_order.factory_issue_reason, ''),
  subcontract_order.submitted_at,
  COALESCE(subcontract_order.submitted_by_ref, subcontract_order.submitted_by::text, ''),
  subcontract_order.approved_at,
  COALESCE(subcontract_order.approved_by_ref, subcontract_order.approved_by::text, ''),
  subcontract_order.factory_confirmed_at,
  COALESCE(subcontract_order.factory_confirmed_by_ref, subcontract_order.factory_confirmed_by::text, ''),
  subcontract_order.deposit_recorded_at,
  COALESCE(subcontract_order.deposit_recorded_by_ref, subcontract_order.deposit_recorded_by::text, ''),
  subcontract_order.materials_issued_at,
  COALESCE(subcontract_order.materials_issued_by_ref, subcontract_order.materials_issued_by::text, ''),
  subcontract_order.sample_submitted_at,
  COALESCE(subcontract_order.sample_submitted_by_ref, subcontract_order.sample_submitted_by::text, ''),
  subcontract_order.sample_approved_at,
  COALESCE(subcontract_order.sample_approved_by_ref, subcontract_order.sample_approved_by::text, ''),
  subcontract_order.sample_rejected_at,
  COALESCE(subcontract_order.sample_rejected_by_ref, subcontract_order.sample_rejected_by::text, ''),
  subcontract_order.mass_production_started_at,
  COALESCE(subcontract_order.mass_production_started_by_ref, subcontract_order.mass_production_started_by::text, ''),
  subcontract_order.finished_goods_received_at,
  COALESCE(subcontract_order.finished_goods_received_by_ref, subcontract_order.finished_goods_received_by::text, ''),
  subcontract_order.qc_started_at,
  COALESCE(subcontract_order.qc_started_by_ref, subcontract_order.qc_started_by::text, ''),
  subcontract_order.accepted_at,
  COALESCE(subcontract_order.accepted_by_ref, subcontract_order.accepted_by::text, ''),
  subcontract_order.rejected_factory_issue_at,
  COALESCE(subcontract_order.rejected_factory_issue_by_ref, subcontract_order.rejected_factory_issue_by::text, ''),
  subcontract_order.final_payment_ready_at,
  COALESCE(subcontract_order.final_payment_ready_by_ref, subcontract_order.final_payment_ready_by::text, ''),
  subcontract_order.closed_at,
  COALESCE(subcontract_order.closed_by_ref, subcontract_order.closed_by::text, ''),
  subcontract_order.cancelled_at,
  COALESCE(subcontract_order.cancelled_by_ref, subcontract_order.cancelled_by::text, '')
FROM subcontract.subcontract_orders AS subcontract_order
LEFT JOIN mdm.suppliers AS factory ON factory.id = subcontract_order.factory_id
LEFT JOIN mdm.items AS finished_item ON finished_item.id = subcontract_order.finished_item_id`

const selectSubcontractOrderHeadersSQL = selectSubcontractOrderHeadersBaseSQL + `
ORDER BY subcontract_order.expected_receipt_date DESC, subcontract_order.order_no ASC`

const findSubcontractOrderHeaderSQL = selectSubcontractOrderHeadersBaseSQL + `
WHERE lower(COALESCE(subcontract_order.order_ref, subcontract_order.id::text)) = lower($1)
   OR subcontract_order.id::text = $1
   OR lower(subcontract_order.order_no) = lower($1)
LIMIT 1`

const findSubcontractOrderHeaderForUpdateSQL = findSubcontractOrderHeaderSQL + `
FOR UPDATE OF subcontract_order`

const selectSubcontractOrderLinesSQL = `
SELECT
  COALESCE(material_line.line_ref, material_line.id::text),
  material_line.line_no,
  COALESCE(material_line.item_ref, material_line.item_id::text, ''),
  COALESCE(material_line.sku_code, item.sku, material_line.item_ref, material_line.item_id::text, ''),
  COALESCE(material_line.item_name, item.name, material_line.item_ref, material_line.item_id::text, ''),
  material_line.planned_qty::text,
  material_line.issued_qty::text,
  material_line.uom_code,
  material_line.base_planned_qty::text,
  material_line.base_issued_qty::text,
  material_line.base_uom_code,
  material_line.conversion_factor::text,
  material_line.unit_cost::text,
  material_line.currency_code,
  material_line.line_cost_amount::text,
  material_line.lot_trace_required,
  COALESCE(material_line.note, '')
FROM subcontract.subcontract_order_material_lines AS material_line
LEFT JOIN mdm.items AS item ON item.id = material_line.item_id
WHERE material_line.subcontract_order_id = $1::uuid
ORDER BY material_line.line_no, material_line.created_at, COALESCE(material_line.line_ref, material_line.id::text)`

const upsertSubcontractOrderSQL = `
INSERT INTO subcontract.subcontract_orders (
  id,
  org_id,
  order_ref,
  org_ref,
  order_no,
  factory_id,
  factory_ref,
  factory_code,
  factory_name,
  finished_item_id,
  finished_item_ref,
  finished_sku_code,
  finished_item_name,
  planned_qty,
  received_qty,
  accepted_qty,
  rejected_qty,
  uom_code,
  base_planned_qty,
  base_received_qty,
  base_accepted_qty,
  base_rejected_qty,
  base_uom_code,
  conversion_factor,
  currency_code,
  estimated_cost_amount,
  deposit_amount,
  spec_summary,
  sample_required,
  claim_window_days,
  target_start_date,
  expected_receipt_date,
  status,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref,
  version,
  cancel_reason,
  sample_reject_reason,
  factory_issue_reason,
  submitted_at,
  submitted_by,
  submitted_by_ref,
  approved_at,
  approved_by,
  approved_by_ref,
  factory_confirmed_at,
  factory_confirmed_by,
  factory_confirmed_by_ref,
  deposit_recorded_at,
  deposit_recorded_by,
  deposit_recorded_by_ref,
  materials_issued_at,
  materials_issued_by,
  materials_issued_by_ref,
  sample_submitted_at,
  sample_submitted_by,
  sample_submitted_by_ref,
  sample_approved_at,
  sample_approved_by,
  sample_approved_by_ref,
  sample_rejected_at,
  sample_rejected_by,
  sample_rejected_by_ref,
  mass_production_started_at,
  mass_production_started_by,
  mass_production_started_by_ref,
  finished_goods_received_at,
  finished_goods_received_by,
  finished_goods_received_by_ref,
  qc_started_at,
  qc_started_by,
  qc_started_by_ref,
  accepted_at,
  accepted_by,
  accepted_by_ref,
  rejected_factory_issue_at,
  rejected_factory_issue_by,
  rejected_factory_issue_by_ref,
  final_payment_ready_at,
  final_payment_ready_by,
  final_payment_ready_by_ref,
  closed_at,
  closed_by,
  closed_by_ref,
  cancelled_at,
  cancelled_by,
  cancelled_by_ref
) VALUES (
  COALESCE(CASE WHEN NULLIF($2::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $2::uuid END, gen_random_uuid()),
  $1::uuid,
  $2,
  $3,
  $4,
  CASE WHEN NULLIF($5::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $5::uuid END,
  $5,
  $6,
  $7,
  CASE WHEN NULLIF($8::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $8::uuid END,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27,
  $28::date,
  $29::date,
  $30,
  $31,
  CASE WHEN NULLIF($32::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $32::uuid END,
  $32,
  $33,
  CASE WHEN NULLIF($34::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $34::uuid END,
  $34,
  $35,
  $36,
  $37,
  $38,
  $39,
  CASE WHEN NULLIF($40::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $40::uuid END,
  $40,
  $41,
  CASE WHEN NULLIF($42::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $42::uuid END,
  $42,
  $43,
  CASE WHEN NULLIF($44::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $44::uuid END,
  $44,
  $45,
  CASE WHEN NULLIF($46::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $46::uuid END,
  $46,
  $47,
  CASE WHEN NULLIF($48::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $48::uuid END,
  $48,
  $49,
  CASE WHEN NULLIF($50::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $50::uuid END,
  $50,
  $51,
  CASE WHEN NULLIF($52::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $52::uuid END,
  $52,
  $53,
  CASE WHEN NULLIF($54::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $54::uuid END,
  $54,
  $55,
  CASE WHEN NULLIF($56::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $56::uuid END,
  $56,
  $57,
  CASE WHEN NULLIF($58::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $58::uuid END,
  $58,
  $59,
  CASE WHEN NULLIF($60::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $60::uuid END,
  $60,
  $61,
  CASE WHEN NULLIF($62::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $62::uuid END,
  $62,
  $63,
  CASE WHEN NULLIF($64::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $64::uuid END,
  $64,
  $65,
  CASE WHEN NULLIF($66::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $66::uuid END,
  $66,
  $67,
  CASE WHEN NULLIF($68::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $68::uuid END,
  $68,
  $69,
  CASE WHEN NULLIF($70::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $70::uuid END,
  $70
)
ON CONFLICT (org_id, order_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  order_no = EXCLUDED.order_no,
  factory_id = EXCLUDED.factory_id,
  factory_ref = EXCLUDED.factory_ref,
  factory_code = EXCLUDED.factory_code,
  factory_name = EXCLUDED.factory_name,
  finished_item_id = EXCLUDED.finished_item_id,
  finished_item_ref = EXCLUDED.finished_item_ref,
  finished_sku_code = EXCLUDED.finished_sku_code,
  finished_item_name = EXCLUDED.finished_item_name,
  planned_qty = EXCLUDED.planned_qty,
  received_qty = EXCLUDED.received_qty,
  accepted_qty = EXCLUDED.accepted_qty,
  rejected_qty = EXCLUDED.rejected_qty,
  uom_code = EXCLUDED.uom_code,
  base_planned_qty = EXCLUDED.base_planned_qty,
  base_received_qty = EXCLUDED.base_received_qty,
  base_accepted_qty = EXCLUDED.base_accepted_qty,
  base_rejected_qty = EXCLUDED.base_rejected_qty,
  base_uom_code = EXCLUDED.base_uom_code,
  conversion_factor = EXCLUDED.conversion_factor,
  currency_code = EXCLUDED.currency_code,
  estimated_cost_amount = EXCLUDED.estimated_cost_amount,
  deposit_amount = EXCLUDED.deposit_amount,
  spec_summary = EXCLUDED.spec_summary,
  sample_required = EXCLUDED.sample_required,
  claim_window_days = EXCLUDED.claim_window_days,
  target_start_date = EXCLUDED.target_start_date,
  expected_receipt_date = EXCLUDED.expected_receipt_date,
  status = EXCLUDED.status,
  created_at = EXCLUDED.created_at,
  created_by = EXCLUDED.created_by,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version,
  cancel_reason = EXCLUDED.cancel_reason,
  sample_reject_reason = EXCLUDED.sample_reject_reason,
  factory_issue_reason = EXCLUDED.factory_issue_reason,
  submitted_at = EXCLUDED.submitted_at,
  submitted_by = EXCLUDED.submitted_by,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  approved_at = EXCLUDED.approved_at,
  approved_by = EXCLUDED.approved_by,
  approved_by_ref = EXCLUDED.approved_by_ref,
  factory_confirmed_at = EXCLUDED.factory_confirmed_at,
  factory_confirmed_by = EXCLUDED.factory_confirmed_by,
  factory_confirmed_by_ref = EXCLUDED.factory_confirmed_by_ref,
  deposit_recorded_at = EXCLUDED.deposit_recorded_at,
  deposit_recorded_by = EXCLUDED.deposit_recorded_by,
  deposit_recorded_by_ref = EXCLUDED.deposit_recorded_by_ref,
  materials_issued_at = EXCLUDED.materials_issued_at,
  materials_issued_by = EXCLUDED.materials_issued_by,
  materials_issued_by_ref = EXCLUDED.materials_issued_by_ref,
  sample_submitted_at = EXCLUDED.sample_submitted_at,
  sample_submitted_by = EXCLUDED.sample_submitted_by,
  sample_submitted_by_ref = EXCLUDED.sample_submitted_by_ref,
  sample_approved_at = EXCLUDED.sample_approved_at,
  sample_approved_by = EXCLUDED.sample_approved_by,
  sample_approved_by_ref = EXCLUDED.sample_approved_by_ref,
  sample_rejected_at = EXCLUDED.sample_rejected_at,
  sample_rejected_by = EXCLUDED.sample_rejected_by,
  sample_rejected_by_ref = EXCLUDED.sample_rejected_by_ref,
  mass_production_started_at = EXCLUDED.mass_production_started_at,
  mass_production_started_by = EXCLUDED.mass_production_started_by,
  mass_production_started_by_ref = EXCLUDED.mass_production_started_by_ref,
  finished_goods_received_at = EXCLUDED.finished_goods_received_at,
  finished_goods_received_by = EXCLUDED.finished_goods_received_by,
  finished_goods_received_by_ref = EXCLUDED.finished_goods_received_by_ref,
  qc_started_at = EXCLUDED.qc_started_at,
  qc_started_by = EXCLUDED.qc_started_by,
  qc_started_by_ref = EXCLUDED.qc_started_by_ref,
  accepted_at = EXCLUDED.accepted_at,
  accepted_by = EXCLUDED.accepted_by,
  accepted_by_ref = EXCLUDED.accepted_by_ref,
  rejected_factory_issue_at = EXCLUDED.rejected_factory_issue_at,
  rejected_factory_issue_by = EXCLUDED.rejected_factory_issue_by,
  rejected_factory_issue_by_ref = EXCLUDED.rejected_factory_issue_by_ref,
  final_payment_ready_at = EXCLUDED.final_payment_ready_at,
  final_payment_ready_by = EXCLUDED.final_payment_ready_by,
  final_payment_ready_by_ref = EXCLUDED.final_payment_ready_by_ref,
  closed_at = EXCLUDED.closed_at,
  closed_by = EXCLUDED.closed_by,
  closed_by_ref = EXCLUDED.closed_by_ref,
  cancelled_at = EXCLUDED.cancelled_at,
  cancelled_by = EXCLUDED.cancelled_by,
  cancelled_by_ref = EXCLUDED.cancelled_by_ref
RETURNING id::text`

const deleteSubcontractOrderLinesSQL = `
DELETE FROM subcontract.subcontract_order_material_lines
WHERE subcontract_order_id = $1::uuid`

const insertSubcontractOrderLineSQL = `
INSERT INTO subcontract.subcontract_order_material_lines (
  id,
  org_id,
  subcontract_order_id,
  line_ref,
  line_no,
  item_id,
  item_ref,
  sku_code,
  item_name,
  planned_qty,
  issued_qty,
  uom_code,
  base_planned_qty,
  base_issued_qty,
  base_uom_code,
  conversion_factor,
  unit_cost,
  currency_code,
  line_cost_amount,
  lot_trace_required,
  note,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
  CASE WHEN NULLIF($5::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $5::uuid END,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  CASE WHEN NULLIF($21::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $21::uuid END,
  $21,
  $22,
  CASE WHEN NULLIF($23::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $23::uuid END,
  $23
)`

const insertSubcontractOrderAuditSQL = `
INSERT INTO audit.audit_logs (
  id,
  org_id,
  actor_id,
  action,
  entity_type,
  entity_id,
  request_id,
  before_data,
  after_data,
  metadata,
  created_at,
  log_ref,
  org_ref,
  actor_ref,
  entity_ref
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  CASE WHEN NULLIF($3::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $3::uuid END,
  $4,
  $5,
  CASE WHEN NULLIF($6::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $6::uuid END,
  $7,
  $8::jsonb,
  $9::jsonb,
  $10::jsonb,
  $11,
  $1,
  $12,
  $3,
  $6
)`

func (s PostgresSubcontractOrderStore) List(
	ctx context.Context,
	filter SubcontractOrderFilter,
) ([]productiondomain.SubcontractOrder, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSubcontractOrderHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]productiondomain.SubcontractOrder, 0)
	for rows.Next() {
		order, err := scanPostgresSubcontractOrder(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if subcontractOrderMatchesFilter(order, filter) {
			orders = append(orders, order)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortSubcontractOrders(orders)

	return orders, nil
}

func (s PostgresSubcontractOrderStore) Get(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractOrder, error) {
	if s.db == nil {
		return productiondomain.SubcontractOrder{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSubcontractOrderHeaderSQL, strings.TrimSpace(id))
	order, err := scanPostgresSubcontractOrder(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractOrder{}, ErrSubcontractOrderNotFound
	}
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}

	return order, nil
}

func (s PostgresSubcontractOrderStore) WithinTx(
	ctx context.Context,
	fn func(context.Context, SubcontractOrderTx) error,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if fn == nil {
		return errors.New("subcontract order transaction function is required")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract order transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, postgresSubcontractOrderTx{store: s, tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subcontract order transaction: %w", err)
	}
	committed = true

	return nil
}

func (tx postgresSubcontractOrderTx) GetForUpdate(
	ctx context.Context,
	id string,
) (productiondomain.SubcontractOrder, error) {
	if tx.tx == nil {
		return productiondomain.SubcontractOrder{}, errors.New("database transaction is required")
	}
	row := tx.tx.QueryRowContext(ctx, findSubcontractOrderHeaderForUpdateSQL, strings.TrimSpace(id))
	order, err := scanPostgresSubcontractOrder(ctx, tx.tx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return productiondomain.SubcontractOrder{}, ErrSubcontractOrderNotFound
	}
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}

	return order, nil
}

func (tx postgresSubcontractOrderTx) Save(
	ctx context.Context,
	order productiondomain.SubcontractOrder,
) error {
	if tx.tx == nil {
		return errors.New("database transaction is required")
	}
	if err := order.Validate(); err != nil {
		return err
	}
	orgID, err := tx.store.resolveOrgID(ctx, tx.tx, order.OrgID)
	if err != nil {
		return err
	}
	persistedID, err := upsertPostgresSubcontractOrder(ctx, tx.tx, orgID, order)
	if err != nil {
		return err
	}
	if err := replacePostgresSubcontractOrderLines(ctx, tx.tx, orgID, persistedID, order); err != nil {
		return err
	}

	return nil
}

func (tx postgresSubcontractOrderTx) RecordAudit(ctx context.Context, log audit.Log) error {
	if tx.tx == nil {
		return errors.New("database transaction is required")
	}
	normalizedLog, err := audit.NewLog(audit.NewLogInput{
		ID:         log.ID,
		OrgID:      log.OrgID,
		ActorID:    log.ActorID,
		Action:     log.Action,
		EntityType: log.EntityType,
		EntityID:   log.EntityID,
		RequestID:  log.RequestID,
		BeforeData: log.BeforeData,
		AfterData:  log.AfterData,
		Metadata:   log.Metadata,
		CreatedAt:  log.CreatedAt,
	})
	if err != nil {
		return err
	}
	orgID, err := tx.store.resolveOrgID(ctx, tx.tx, firstNonBlankSubcontractOrder(normalizedLog.OrgID, defaultSubcontractOrderOrgID))
	if err != nil {
		return err
	}
	beforeJSON, err := postgresSubcontractOrderJSONMap(normalizedLog.BeforeData)
	if err != nil {
		return fmt.Errorf("encode subcontract order audit before_data: %w", err)
	}
	afterJSON, err := postgresSubcontractOrderJSONMap(normalizedLog.AfterData)
	if err != nil {
		return fmt.Errorf("encode subcontract order audit after_data: %w", err)
	}
	metadataJSON, err := requiredPostgresSubcontractOrderJSONMap(normalizedLog.Metadata)
	if err != nil {
		return fmt.Errorf("encode subcontract order audit metadata: %w", err)
	}

	_, err = tx.tx.ExecContext(
		ctx,
		insertSubcontractOrderAuditSQL,
		nullablePostgresSubcontractOrderText(normalizedLog.ID),
		orgID,
		nullablePostgresSubcontractOrderText(normalizedLog.ActorID),
		normalizedLog.Action,
		normalizedLog.EntityType,
		nullablePostgresSubcontractOrderText(normalizedLog.EntityID),
		nullablePostgresSubcontractOrderText(normalizedLog.RequestID),
		beforeJSON,
		afterJSON,
		metadataJSON,
		normalizedLog.CreatedAt.UTC(),
		nullablePostgresSubcontractOrderText(normalizedLog.OrgID),
	)
	if err != nil {
		return fmt.Errorf("insert subcontract order audit: %w", err)
	}

	return nil
}

func (s PostgresSubcontractOrderStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractOrderQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractOrderUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractOrderOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractOrderUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract order org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractOrderUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract order org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractOrder(
	ctx context.Context,
	queryer postgresSubcontractOrderQueryer,
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractOrder, error) {
	header, err := scanPostgresSubcontractOrderHeader(row)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	lines, err := listPostgresSubcontractOrderLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}

	return buildPostgresSubcontractOrder(header, lines)
}

func scanPostgresSubcontractOrderHeader(row interface{ Scan(dest ...any) error }) (postgresSubcontractOrderHeader, error) {
	var header postgresSubcontractOrderHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.OrderNo,
		&header.FactoryID,
		&header.FactoryCode,
		&header.FactoryName,
		&header.FinishedItemID,
		&header.FinishedSKUCode,
		&header.FinishedItemName,
		&header.PlannedQty,
		&header.ReceivedQty,
		&header.AcceptedQty,
		&header.RejectedQty,
		&header.UOMCode,
		&header.BasePlannedQty,
		&header.BaseReceivedQty,
		&header.BaseAcceptedQty,
		&header.BaseRejectedQty,
		&header.BaseUOMCode,
		&header.ConversionFactor,
		&header.CurrencyCode,
		&header.EstimatedCostAmount,
		&header.DepositAmount,
		&header.SpecSummary,
		&header.SampleRequired,
		&header.ClaimWindowDays,
		&header.TargetStartDate,
		&header.ExpectedReceiptDate,
		&header.Status,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
		&header.CancelReason,
		&header.SampleRejectReason,
		&header.FactoryIssueReason,
		&header.SubmittedAt,
		&header.SubmittedBy,
		&header.ApprovedAt,
		&header.ApprovedBy,
		&header.FactoryConfirmedAt,
		&header.FactoryConfirmedBy,
		&header.DepositRecordedAt,
		&header.DepositRecordedBy,
		&header.MaterialsIssuedAt,
		&header.MaterialsIssuedBy,
		&header.SampleSubmittedAt,
		&header.SampleSubmittedBy,
		&header.SampleApprovedAt,
		&header.SampleApprovedBy,
		&header.SampleRejectedAt,
		&header.SampleRejectedBy,
		&header.MassProductionStartedAt,
		&header.MassProductionStartedBy,
		&header.FinishedGoodsReceivedAt,
		&header.FinishedGoodsReceivedBy,
		&header.QCStartedAt,
		&header.QCStartedBy,
		&header.AcceptedAt,
		&header.AcceptedBy,
		&header.RejectedFactoryIssueAt,
		&header.RejectedFactoryIssueBy,
		&header.FinalPaymentReadyAt,
		&header.FinalPaymentReadyBy,
		&header.ClosedAt,
		&header.ClosedBy,
		&header.CancelledAt,
		&header.CancelledBy,
	)
	if err != nil {
		return postgresSubcontractOrderHeader{}, err
	}

	return header, nil
}

func listPostgresSubcontractOrderLines(
	ctx context.Context,
	queryer postgresSubcontractOrderQueryer,
	persistedID string,
) ([]productiondomain.SubcontractMaterialLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractOrderLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]productiondomain.SubcontractMaterialLine, 0)
	for rows.Next() {
		line, err := scanPostgresSubcontractOrderLine(rows)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func scanPostgresSubcontractOrderLine(row interface{ Scan(dest ...any) error }) (productiondomain.SubcontractMaterialLine, error) {
	var (
		id               string
		lineNo           int
		itemID           string
		skuCode          string
		itemName         string
		plannedQtyText   string
		issuedQtyText    string
		uomCode          string
		basePlannedText  string
		baseIssuedText   string
		baseUOMCode      string
		conversionText   string
		unitCostText     string
		currencyCode     string
		lineCostText     string
		lotTraceRequired bool
		note             string
	)
	if err := row.Scan(
		&id,
		&lineNo,
		&itemID,
		&skuCode,
		&itemName,
		&plannedQtyText,
		&issuedQtyText,
		&uomCode,
		&basePlannedText,
		&baseIssuedText,
		&baseUOMCode,
		&conversionText,
		&unitCostText,
		&currencyCode,
		&lineCostText,
		&lotTraceRequired,
		&note,
	); err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	plannedQty, err := decimal.ParseQuantity(plannedQtyText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	issuedQty, err := decimal.ParseQuantity(issuedQtyText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	basePlannedQty, err := decimal.ParseQuantity(basePlannedText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	baseIssuedQty, err := decimal.ParseQuantity(baseIssuedText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	conversion, err := decimal.ParseQuantity(conversionText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	unitCost, err := decimal.ParseUnitCost(unitCostText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	lineCost, err := decimal.ParseMoneyAmount(lineCostText)
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	line, err := productiondomain.NewSubcontractMaterialLine(productiondomain.NewSubcontractMaterialLineInput{
		ID:               id,
		LineNo:           lineNo,
		ItemID:           itemID,
		SKUCode:          skuCode,
		ItemName:         itemName,
		PlannedQty:       plannedQty,
		IssuedQty:        issuedQty,
		UOMCode:          uomCode,
		BasePlannedQty:   basePlannedQty,
		BaseIssuedQty:    baseIssuedQty,
		BaseUOMCode:      baseUOMCode,
		ConversionFactor: conversion,
		UnitCost:         unitCost,
		CurrencyCode:     currencyCode,
		LotTraceRequired: lotTraceRequired,
		Note:             note,
	})
	if err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}
	line.LineCostAmount = lineCost
	if err := line.Validate(); err != nil {
		return productiondomain.SubcontractMaterialLine{}, err
	}

	return line, nil
}

func buildPostgresSubcontractOrder(
	header postgresSubcontractOrderHeader,
	lines []productiondomain.SubcontractMaterialLine,
) (productiondomain.SubcontractOrder, error) {
	lineInputs := make([]productiondomain.NewSubcontractMaterialLineInput, 0, len(lines))
	for _, line := range lines {
		lineInputs = append(lineInputs, productiondomain.NewSubcontractMaterialLineInput{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			ItemName:         line.ItemName,
			PlannedQty:       line.PlannedQty,
			IssuedQty:        line.IssuedQty,
			UOMCode:          line.UOMCode.String(),
			BasePlannedQty:   line.BasePlannedQty,
			BaseIssuedQty:    line.BaseIssuedQty,
			BaseUOMCode:      line.BaseUOMCode.String(),
			ConversionFactor: line.ConversionFactor,
			UnitCost:         line.UnitCost,
			CurrencyCode:     line.CurrencyCode.String(),
			LotTraceRequired: line.LotTraceRequired,
			Note:             line.Note,
		})
	}
	plannedQty, err := decimal.ParseQuantity(header.PlannedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	receivedQty, err := decimal.ParseQuantity(header.ReceivedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	acceptedQty, err := decimal.ParseQuantity(header.AcceptedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	rejectedQty, err := decimal.ParseQuantity(header.RejectedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	basePlannedQty, err := decimal.ParseQuantity(header.BasePlannedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	baseReceivedQty, err := decimal.ParseQuantity(header.BaseReceivedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	baseAcceptedQty, err := decimal.ParseQuantity(header.BaseAcceptedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	baseRejectedQty, err := decimal.ParseQuantity(header.BaseRejectedQty)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	conversion, err := decimal.ParseQuantity(header.ConversionFactor)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	depositAmount, err := decimal.ParseMoneyAmount(header.DepositAmount)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	order, err := productiondomain.NewSubcontractOrderDocument(productiondomain.NewSubcontractOrderDocumentInput{
		ID:                  header.ID,
		OrgID:               header.OrgID,
		OrderNo:             header.OrderNo,
		FactoryID:           header.FactoryID,
		FactoryCode:         header.FactoryCode,
		FactoryName:         header.FactoryName,
		FinishedItemID:      header.FinishedItemID,
		FinishedSKUCode:     header.FinishedSKUCode,
		FinishedItemName:    header.FinishedItemName,
		PlannedQty:          plannedQty,
		ReceivedQty:         receivedQty,
		AcceptedQty:         acceptedQty,
		RejectedQty:         rejectedQty,
		UOMCode:             header.UOMCode,
		BasePlannedQty:      basePlannedQty,
		BaseReceivedQty:     baseReceivedQty,
		BaseAcceptedQty:     baseAcceptedQty,
		BaseRejectedQty:     baseRejectedQty,
		BaseUOMCode:         header.BaseUOMCode,
		ConversionFactor:    conversion,
		CurrencyCode:        header.CurrencyCode,
		DepositAmount:       depositAmount,
		SpecSummary:         header.SpecSummary,
		SampleRequired:      header.SampleRequired,
		ClaimWindowDays:     header.ClaimWindowDays,
		TargetStartDate:     header.TargetStartDate,
		ExpectedReceiptDate: header.ExpectedReceiptDate,
		MaterialLines:       lineInputs,
		CreatedAt:           header.CreatedAt,
		CreatedBy:           firstNonBlankSubcontractOrder(header.CreatedBy, "system"),
		UpdatedAt:           header.UpdatedAt,
		UpdatedBy:           firstNonBlankSubcontractOrder(header.UpdatedBy, header.CreatedBy, "system"),
	})
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	order.MaterialLines = lines
	order.Status = productiondomain.SubcontractOrderStatus(header.Status)
	order.EstimatedCostAmount, err = decimal.ParseMoneyAmount(header.EstimatedCostAmount)
	if err != nil {
		return productiondomain.SubcontractOrder{}, err
	}
	order.DepositAmount = depositAmount
	order.Version = header.Version
	order.CancelReason = strings.TrimSpace(header.CancelReason)
	order.SampleRejectReason = strings.TrimSpace(header.SampleRejectReason)
	order.FactoryIssueReason = strings.TrimSpace(header.FactoryIssueReason)
	order.SubmittedAt = nullablePostgresSubcontractOrderTimeValue(header.SubmittedAt)
	order.SubmittedBy = strings.TrimSpace(header.SubmittedBy)
	order.ApprovedAt = nullablePostgresSubcontractOrderTimeValue(header.ApprovedAt)
	order.ApprovedBy = strings.TrimSpace(header.ApprovedBy)
	order.FactoryConfirmedAt = nullablePostgresSubcontractOrderTimeValue(header.FactoryConfirmedAt)
	order.FactoryConfirmedBy = strings.TrimSpace(header.FactoryConfirmedBy)
	order.DepositRecordedAt = nullablePostgresSubcontractOrderTimeValue(header.DepositRecordedAt)
	order.DepositRecordedBy = strings.TrimSpace(header.DepositRecordedBy)
	order.MaterialsIssuedAt = nullablePostgresSubcontractOrderTimeValue(header.MaterialsIssuedAt)
	order.MaterialsIssuedBy = strings.TrimSpace(header.MaterialsIssuedBy)
	order.SampleSubmittedAt = nullablePostgresSubcontractOrderTimeValue(header.SampleSubmittedAt)
	order.SampleSubmittedBy = strings.TrimSpace(header.SampleSubmittedBy)
	order.SampleApprovedAt = nullablePostgresSubcontractOrderTimeValue(header.SampleApprovedAt)
	order.SampleApprovedBy = strings.TrimSpace(header.SampleApprovedBy)
	order.SampleRejectedAt = nullablePostgresSubcontractOrderTimeValue(header.SampleRejectedAt)
	order.SampleRejectedBy = strings.TrimSpace(header.SampleRejectedBy)
	order.MassProductionStartedAt = nullablePostgresSubcontractOrderTimeValue(header.MassProductionStartedAt)
	order.MassProductionStartedBy = strings.TrimSpace(header.MassProductionStartedBy)
	order.FinishedGoodsReceivedAt = nullablePostgresSubcontractOrderTimeValue(header.FinishedGoodsReceivedAt)
	order.FinishedGoodsReceivedBy = strings.TrimSpace(header.FinishedGoodsReceivedBy)
	order.QCStartedAt = nullablePostgresSubcontractOrderTimeValue(header.QCStartedAt)
	order.QCStartedBy = strings.TrimSpace(header.QCStartedBy)
	order.AcceptedAt = nullablePostgresSubcontractOrderTimeValue(header.AcceptedAt)
	order.AcceptedBy = strings.TrimSpace(header.AcceptedBy)
	order.RejectedFactoryIssueAt = nullablePostgresSubcontractOrderTimeValue(header.RejectedFactoryIssueAt)
	order.RejectedFactoryIssueBy = strings.TrimSpace(header.RejectedFactoryIssueBy)
	order.FinalPaymentReadyAt = nullablePostgresSubcontractOrderTimeValue(header.FinalPaymentReadyAt)
	order.FinalPaymentReadyBy = strings.TrimSpace(header.FinalPaymentReadyBy)
	order.ClosedAt = nullablePostgresSubcontractOrderTimeValue(header.ClosedAt)
	order.ClosedBy = strings.TrimSpace(header.ClosedBy)
	order.CancelledAt = nullablePostgresSubcontractOrderTimeValue(header.CancelledAt)
	order.CancelledBy = strings.TrimSpace(header.CancelledBy)
	if err := order.Validate(); err != nil {
		return productiondomain.SubcontractOrder{}, err
	}

	return order, nil
}

func upsertPostgresSubcontractOrder(
	ctx context.Context,
	queryer postgresSubcontractOrderQueryer,
	orgID string,
	order productiondomain.SubcontractOrder,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractOrderSQL,
		orgID,
		order.ID,
		order.OrgID,
		order.OrderNo,
		nullablePostgresSubcontractOrderText(order.FactoryID),
		nullablePostgresSubcontractOrderText(order.FactoryCode),
		order.FactoryName,
		nullablePostgresSubcontractOrderText(order.FinishedItemID),
		order.FinishedSKUCode,
		order.FinishedItemName,
		order.PlannedQty.String(),
		order.ReceivedQty.String(),
		order.AcceptedQty.String(),
		order.RejectedQty.String(),
		order.UOMCode.String(),
		order.BasePlannedQty.String(),
		order.BaseReceivedQty.String(),
		order.BaseAcceptedQty.String(),
		order.BaseRejectedQty.String(),
		order.BaseUOMCode.String(),
		order.ConversionFactor.String(),
		order.CurrencyCode.String(),
		order.EstimatedCostAmount.String(),
		order.DepositAmount.String(),
		nullablePostgresSubcontractOrderText(order.SpecSummary),
		order.SampleRequired,
		order.ClaimWindowDays,
		nullablePostgresSubcontractOrderText(order.TargetStartDate),
		nullablePostgresSubcontractOrderText(order.ExpectedReceiptDate),
		string(order.Status),
		order.CreatedAt.UTC(),
		nullablePostgresSubcontractOrderText(order.CreatedBy),
		order.UpdatedAt.UTC(),
		nullablePostgresSubcontractOrderText(order.UpdatedBy),
		order.Version,
		nullablePostgresSubcontractOrderText(order.CancelReason),
		nullablePostgresSubcontractOrderText(order.SampleRejectReason),
		nullablePostgresSubcontractOrderText(order.FactoryIssueReason),
		nullablePostgresSubcontractOrderTime(order.SubmittedAt),
		nullablePostgresSubcontractOrderText(order.SubmittedBy),
		nullablePostgresSubcontractOrderTime(order.ApprovedAt),
		nullablePostgresSubcontractOrderText(order.ApprovedBy),
		nullablePostgresSubcontractOrderTime(order.FactoryConfirmedAt),
		nullablePostgresSubcontractOrderText(order.FactoryConfirmedBy),
		nullablePostgresSubcontractOrderTime(order.DepositRecordedAt),
		nullablePostgresSubcontractOrderText(order.DepositRecordedBy),
		nullablePostgresSubcontractOrderTime(order.MaterialsIssuedAt),
		nullablePostgresSubcontractOrderText(order.MaterialsIssuedBy),
		nullablePostgresSubcontractOrderTime(order.SampleSubmittedAt),
		nullablePostgresSubcontractOrderText(order.SampleSubmittedBy),
		nullablePostgresSubcontractOrderTime(order.SampleApprovedAt),
		nullablePostgresSubcontractOrderText(order.SampleApprovedBy),
		nullablePostgresSubcontractOrderTime(order.SampleRejectedAt),
		nullablePostgresSubcontractOrderText(order.SampleRejectedBy),
		nullablePostgresSubcontractOrderTime(order.MassProductionStartedAt),
		nullablePostgresSubcontractOrderText(order.MassProductionStartedBy),
		nullablePostgresSubcontractOrderTime(order.FinishedGoodsReceivedAt),
		nullablePostgresSubcontractOrderText(order.FinishedGoodsReceivedBy),
		nullablePostgresSubcontractOrderTime(order.QCStartedAt),
		nullablePostgresSubcontractOrderText(order.QCStartedBy),
		nullablePostgresSubcontractOrderTime(order.AcceptedAt),
		nullablePostgresSubcontractOrderText(order.AcceptedBy),
		nullablePostgresSubcontractOrderTime(order.RejectedFactoryIssueAt),
		nullablePostgresSubcontractOrderText(order.RejectedFactoryIssueBy),
		nullablePostgresSubcontractOrderTime(order.FinalPaymentReadyAt),
		nullablePostgresSubcontractOrderText(order.FinalPaymentReadyBy),
		nullablePostgresSubcontractOrderTime(order.ClosedAt),
		nullablePostgresSubcontractOrderText(order.ClosedBy),
		nullablePostgresSubcontractOrderTime(order.CancelledAt),
		nullablePostgresSubcontractOrderText(order.CancelledBy),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract order %q: %w", order.ID, err)
	}

	return persistedID, nil
}

func replacePostgresSubcontractOrderLines(
	ctx context.Context,
	queryer postgresSubcontractOrderQueryer,
	orgID string,
	persistedID string,
	order productiondomain.SubcontractOrder,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractOrderLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract order material lines: %w", err)
	}
	for _, line := range order.MaterialLines {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractOrderLineSQL,
			line.ID,
			orgID,
			persistedID,
			line.LineNo,
			nullablePostgresSubcontractOrderText(line.ItemID),
			line.SKUCode,
			line.ItemName,
			line.PlannedQty.String(),
			line.IssuedQty.String(),
			line.UOMCode.String(),
			line.BasePlannedQty.String(),
			line.BaseIssuedQty.String(),
			line.BaseUOMCode.String(),
			line.ConversionFactor.String(),
			line.UnitCost.String(),
			line.CurrencyCode.String(),
			line.LineCostAmount.String(),
			line.LotTraceRequired,
			nullablePostgresSubcontractOrderText(line.Note),
			order.CreatedAt.UTC(),
			nullablePostgresSubcontractOrderText(order.CreatedBy),
			order.UpdatedAt.UTC(),
			nullablePostgresSubcontractOrderText(order.UpdatedBy),
		); err != nil {
			return fmt.Errorf("insert subcontract order material line %q: %w", line.ID, err)
		}
	}

	return nil
}

func postgresSubcontractOrderJSONMap(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}

	return requiredPostgresSubcontractOrderJSONMap(value)
}

func requiredPostgresSubcontractOrderJSONMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func nullablePostgresSubcontractOrderText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresSubcontractOrderTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullablePostgresSubcontractOrderTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isPostgresSubcontractOrderUUIDText(value string) bool {
	value = strings.TrimSpace(value)
	if len(value) != 36 {
		return false
	}
	for index, char := range value {
		switch index {
		case 8, 13, 18, 23:
			if char != '-' {
				return false
			}
		default:
			if !isPostgresSubcontractOrderHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractOrderHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractOrderStore = PostgresSubcontractOrderStore{}
