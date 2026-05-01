package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSalesOrderStoreConfig struct {
	DefaultOrgID string
}

type PostgresSalesOrderStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSalesOrderTx struct {
	store PostgresSalesOrderStore
	tx    *sql.Tx
}

type postgresSalesOrderQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSalesOrderHeader struct {
	PersistedID       string
	ID                string
	OrgID             string
	OrderNo           string
	CustomerID        string
	CustomerCode      string
	CustomerName      string
	Channel           string
	WarehouseID       string
	WarehouseCode     string
	OrderDate         string
	Status            string
	CurrencyCode      string
	SubtotalAmount    string
	DiscountAmount    string
	TaxAmount         string
	ShippingFeeAmount string
	NetAmount         string
	TotalAmount       string
	Note              string
	CreatedAt         time.Time
	CreatedBy         string
	UpdatedAt         time.Time
	UpdatedBy         string
	Version           int
	CancelReason      string
	ConfirmedAt       sql.NullTime
	ConfirmedBy       string
	ReservedAt        sql.NullTime
	ReservedBy        string
	PickingStartedAt  sql.NullTime
	PickingStartedBy  string
	PickedAt          sql.NullTime
	PickedBy          string
	PackingStartedAt  sql.NullTime
	PackingStartedBy  string
	PackedAt          sql.NullTime
	PackedBy          string
	WaitingHandoverAt sql.NullTime
	WaitingHandoverBy string
	HandedOverAt      sql.NullTime
	HandedOverBy      string
	ClosedAt          sql.NullTime
	ClosedBy          string
	CancelledAt       sql.NullTime
	CancelledBy       string
	ExceptionAt       sql.NullTime
	ExceptionBy       string
}

func NewPostgresSalesOrderStore(db *sql.DB, cfg PostgresSalesOrderStoreConfig) PostgresSalesOrderStore {
	return PostgresSalesOrderStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSalesOrderOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSalesOrderHeadersBaseSQL = `
SELECT
  id::text,
  COALESCE(order_ref, id::text),
  COALESCE(org_ref, org_id::text),
  order_no,
  COALESCE(customer_ref, customer_id::text, ''),
  COALESCE(customer_code, ''),
  COALESCE(customer_name, ''),
  channel,
  COALESCE(warehouse_ref, warehouse_id::text, ''),
  COALESCE(warehouse_code, ''),
  order_date::text,
  status,
  currency_code,
  subtotal_amount::text,
  discount_amount::text,
  tax_amount::text,
  shipping_fee_amount::text,
  net_amount::text,
  total_amount::text,
  COALESCE(note, ''),
  created_at,
  COALESCE(created_by_ref, created_by::text, ''),
  updated_at,
  COALESCE(updated_by_ref, updated_by::text, ''),
  version,
  COALESCE(cancel_reason, ''),
  confirmed_at,
  COALESCE(confirmed_by_ref, confirmed_by::text, ''),
  reserved_at,
  COALESCE(reserved_by_ref, reserved_by::text, ''),
  picking_started_at,
  COALESCE(picking_started_by_ref, picking_started_by::text, ''),
  picked_at,
  COALESCE(picked_by_ref, picked_by::text, ''),
  packing_started_at,
  COALESCE(packing_started_by_ref, packing_started_by::text, ''),
  packed_at,
  COALESCE(packed_by_ref, packed_by::text, ''),
  waiting_handover_at,
  COALESCE(waiting_handover_by_ref, waiting_handover_by::text, ''),
  handed_over_at,
  COALESCE(handed_over_by_ref, handed_over_by::text, ''),
  closed_at,
  COALESCE(closed_by_ref, closed_by::text, ''),
  cancelled_at,
  COALESCE(cancelled_by_ref, cancelled_by::text, ''),
  exception_at,
  COALESCE(exception_by_ref, exception_by::text, '')
FROM sales.sales_orders`

const selectSalesOrderHeadersSQL = selectSalesOrderHeadersBaseSQL + `
ORDER BY order_date DESC, order_no ASC`

const findSalesOrderHeaderSQL = selectSalesOrderHeadersBaseSQL + `
WHERE order_ref = $1 OR id::text = $1
LIMIT 1`

const findSalesOrderHeaderForUpdateSQL = findSalesOrderHeaderSQL + `
FOR UPDATE`

const selectSalesOrderLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  line.line_no,
  COALESCE(line.item_ref, line.item_id::text, ''),
  COALESCE(line.sku_code, item.sku, ''),
  COALESCE(line.item_name, item.name, ''),
  line.ordered_qty::text,
  line.uom_code,
  line.base_ordered_qty::text,
  line.base_uom_code,
  line.conversion_factor::text,
  line.unit_price::text,
  line.currency_code,
  line.line_discount_amount::text,
  line.line_amount::text,
  line.reserved_qty::text,
  line.shipped_qty::text,
  COALESCE(line.batch_ref, line.batch_id::text, ''),
  COALESCE(line.batch_no, batch.batch_no, '')
FROM sales.sales_order_lines AS line
LEFT JOIN mdm.items AS item ON item.id = line.item_id
LEFT JOIN inventory.batches AS batch ON batch.id = line.batch_id
WHERE line.sales_order_id = $1::uuid
ORDER BY line.line_no, line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertSalesOrderSQL = `
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
  version,
  cancel_reason,
  cancelled_at,
  cancelled_by,
  cancelled_by_ref,
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
  packing_started_by_ref,
  packed_at,
  packed_by,
  packed_by_ref,
  waiting_handover_at,
  waiting_handover_by,
  waiting_handover_by_ref,
  handed_over_at,
  handed_over_by,
  handed_over_by_ref,
  closed_at,
  closed_by,
  closed_by_ref,
  exception_at,
  exception_by,
  exception_by_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6::uuid,
  $7,
  $8,
  $9,
  $10::date,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20::uuid,
  $21,
  $22,
  $23,
  $24,
  $25::uuid,
  $26,
  $27,
  $28::uuid,
  $29,
  $30,
  $31,
  $32,
  $33::uuid,
  $34,
  $35,
  $36::uuid,
  $37,
  $38,
  $39::uuid,
  $40,
  $41,
  $42::uuid,
  $43,
  $44,
  $45::uuid,
  $46,
  $47,
  $48::uuid,
  $49,
  $50,
  $51::uuid,
  $52,
  $53,
  $54::uuid,
  $55,
  $56,
  $57::uuid,
  $58,
  $59,
  $60::uuid,
  $61,
  $62,
  $63::uuid,
  $64
)
ON CONFLICT (org_id, order_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  order_no = EXCLUDED.order_no,
  customer_id = EXCLUDED.customer_id,
  customer_ref = EXCLUDED.customer_ref,
  customer_code = EXCLUDED.customer_code,
  customer_name = EXCLUDED.customer_name,
  order_date = EXCLUDED.order_date,
  channel = EXCLUDED.channel,
  status = EXCLUDED.status,
  currency_code = EXCLUDED.currency_code,
  subtotal_amount = EXCLUDED.subtotal_amount,
  discount_amount = EXCLUDED.discount_amount,
  tax_amount = EXCLUDED.tax_amount,
  shipping_fee_amount = EXCLUDED.shipping_fee_amount,
  net_amount = EXCLUDED.net_amount,
  total_amount = EXCLUDED.total_amount,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  note = EXCLUDED.note,
  created_at = EXCLUDED.created_at,
  created_by = EXCLUDED.created_by,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version,
  cancel_reason = EXCLUDED.cancel_reason,
  cancelled_at = EXCLUDED.cancelled_at,
  cancelled_by = EXCLUDED.cancelled_by,
  cancelled_by_ref = EXCLUDED.cancelled_by_ref,
  confirmed_at = EXCLUDED.confirmed_at,
  confirmed_by = EXCLUDED.confirmed_by,
  confirmed_by_ref = EXCLUDED.confirmed_by_ref,
  reserved_at = EXCLUDED.reserved_at,
  reserved_by = EXCLUDED.reserved_by,
  reserved_by_ref = EXCLUDED.reserved_by_ref,
  picking_started_at = EXCLUDED.picking_started_at,
  picking_started_by = EXCLUDED.picking_started_by,
  picking_started_by_ref = EXCLUDED.picking_started_by_ref,
  picked_at = EXCLUDED.picked_at,
  picked_by = EXCLUDED.picked_by,
  picked_by_ref = EXCLUDED.picked_by_ref,
  packing_started_at = EXCLUDED.packing_started_at,
  packing_started_by = EXCLUDED.packing_started_by,
  packing_started_by_ref = EXCLUDED.packing_started_by_ref,
  packed_at = EXCLUDED.packed_at,
  packed_by = EXCLUDED.packed_by,
  packed_by_ref = EXCLUDED.packed_by_ref,
  waiting_handover_at = EXCLUDED.waiting_handover_at,
  waiting_handover_by = EXCLUDED.waiting_handover_by,
  waiting_handover_by_ref = EXCLUDED.waiting_handover_by_ref,
  handed_over_at = EXCLUDED.handed_over_at,
  handed_over_by = EXCLUDED.handed_over_by,
  handed_over_by_ref = EXCLUDED.handed_over_by_ref,
  closed_at = EXCLUDED.closed_at,
  closed_by = EXCLUDED.closed_by,
  closed_by_ref = EXCLUDED.closed_by_ref,
  exception_at = EXCLUDED.exception_at,
  exception_by = EXCLUDED.exception_by,
  exception_by_ref = EXCLUDED.exception_by_ref
RETURNING id::text`

const deleteSalesOrderLinesSQL = `
DELETE FROM sales.sales_order_lines
WHERE sales_order_id = $1::uuid`

const insertSalesOrderLineSQL = `
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
  batch_id,
  batch_ref,
  batch_no,
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
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6::uuid,
  $7,
  $8,
  $9,
  $10::uuid,
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
  $25::uuid,
  $26,
  $27::uuid
)`

const insertSalesOrderAuditSQL = `
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
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6::uuid,
  $7,
  $8::jsonb,
  $9::jsonb,
  $10::jsonb,
  $11,
  $12,
  $13,
  $14,
  $15
)`

func (s PostgresSalesOrderStore) List(ctx context.Context, filter SalesOrderFilter) ([]salesdomain.SalesOrder, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSalesOrderHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]salesdomain.SalesOrder, 0)
	for rows.Next() {
		order, err := scanPostgresSalesOrder(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if salesOrderMatchesFilter(order, filter) {
			orders = append(orders, order)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortSalesOrders(orders)

	return orders, nil
}

func (s PostgresSalesOrderStore) Get(ctx context.Context, id string) (salesdomain.SalesOrder, error) {
	if s.db == nil {
		return salesdomain.SalesOrder{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSalesOrderHeaderSQL, strings.TrimSpace(id))
	order, err := scanPostgresSalesOrder(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return salesdomain.SalesOrder{}, ErrSalesOrderNotFound
	}
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}

	return order, nil
}

func (s PostgresSalesOrderStore) WithinTx(
	ctx context.Context,
	fn func(context.Context, SalesOrderTx) error,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if fn == nil {
		return errors.New("sales order transaction function is required")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin sales order transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, postgresSalesOrderTx{store: s, tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit sales order transaction: %w", err)
	}
	committed = true

	return nil
}

func (tx postgresSalesOrderTx) GetForUpdate(
	ctx context.Context,
	id string,
) (salesdomain.SalesOrder, error) {
	if tx.tx == nil {
		return salesdomain.SalesOrder{}, errors.New("database transaction is required")
	}
	row := tx.tx.QueryRowContext(ctx, findSalesOrderHeaderForUpdateSQL, strings.TrimSpace(id))
	order, err := scanPostgresSalesOrder(ctx, tx.tx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return salesdomain.SalesOrder{}, ErrSalesOrderNotFound
	}
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}

	return order, nil
}

func (tx postgresSalesOrderTx) Save(ctx context.Context, order salesdomain.SalesOrder) error {
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
	persistedID, err := upsertSalesOrder(ctx, tx.tx, orgID, order)
	if err != nil {
		return err
	}
	if err := replaceSalesOrderLines(ctx, tx.tx, orgID, persistedID, order); err != nil {
		return err
	}

	return nil
}

func (tx postgresSalesOrderTx) RecordAudit(ctx context.Context, log audit.Log) error {
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
	orgID, err := tx.store.resolveOrgID(ctx, tx.tx, firstNonBlank(normalizedLog.OrgID, defaultSalesOrderOrgID))
	if err != nil {
		return err
	}
	beforeJSON, err := postgresSalesOrderJSONMap(normalizedLog.BeforeData)
	if err != nil {
		return fmt.Errorf("encode sales order audit before_data: %w", err)
	}
	afterJSON, err := postgresSalesOrderJSONMap(normalizedLog.AfterData)
	if err != nil {
		return fmt.Errorf("encode sales order audit after_data: %w", err)
	}
	metadataJSON, err := requiredPostgresSalesOrderJSONMap(normalizedLog.Metadata)
	if err != nil {
		return fmt.Errorf("encode sales order audit metadata: %w", err)
	}

	_, err = tx.tx.ExecContext(
		ctx,
		insertSalesOrderAuditSQL,
		nullableSalesOrderUUID(normalizedLog.ID),
		orgID,
		nullableSalesOrderUUID(normalizedLog.ActorID),
		normalizedLog.Action,
		normalizedLog.EntityType,
		nullableSalesOrderUUID(normalizedLog.EntityID),
		nullableSalesOrderText(normalizedLog.RequestID),
		beforeJSON,
		afterJSON,
		metadataJSON,
		normalizedLog.CreatedAt.UTC(),
		nullableSalesOrderText(normalizedLog.ID),
		nullableSalesOrderText(normalizedLog.OrgID),
		nullableSalesOrderText(normalizedLog.ActorID),
		nullableSalesOrderText(normalizedLog.EntityID),
	)
	if err != nil {
		return fmt.Errorf("insert sales order audit: %w", err)
	}

	return nil
}

func (s PostgresSalesOrderStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSalesOrderQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isSalesOrderUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSalesOrderOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isSalesOrderUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve sales order org %q: %w", orgRef, err)
		}
	}
	if isSalesOrderUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("sales order org %q cannot be resolved", orgRef)
}

func scanPostgresSalesOrder(
	ctx context.Context,
	queryer postgresSalesOrderQueryer,
	row interface{ Scan(dest ...any) error },
) (salesdomain.SalesOrder, error) {
	header, err := scanPostgresSalesOrderHeader(row)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	lines, err := listPostgresSalesOrderLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}

	return buildPostgresSalesOrder(header, lines)
}

func scanPostgresSalesOrderHeader(row interface{ Scan(dest ...any) error }) (postgresSalesOrderHeader, error) {
	var header postgresSalesOrderHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.OrderNo,
		&header.CustomerID,
		&header.CustomerCode,
		&header.CustomerName,
		&header.Channel,
		&header.WarehouseID,
		&header.WarehouseCode,
		&header.OrderDate,
		&header.Status,
		&header.CurrencyCode,
		&header.SubtotalAmount,
		&header.DiscountAmount,
		&header.TaxAmount,
		&header.ShippingFeeAmount,
		&header.NetAmount,
		&header.TotalAmount,
		&header.Note,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
		&header.CancelReason,
		&header.ConfirmedAt,
		&header.ConfirmedBy,
		&header.ReservedAt,
		&header.ReservedBy,
		&header.PickingStartedAt,
		&header.PickingStartedBy,
		&header.PickedAt,
		&header.PickedBy,
		&header.PackingStartedAt,
		&header.PackingStartedBy,
		&header.PackedAt,
		&header.PackedBy,
		&header.WaitingHandoverAt,
		&header.WaitingHandoverBy,
		&header.HandedOverAt,
		&header.HandedOverBy,
		&header.ClosedAt,
		&header.ClosedBy,
		&header.CancelledAt,
		&header.CancelledBy,
		&header.ExceptionAt,
		&header.ExceptionBy,
	)
	if err != nil {
		return postgresSalesOrderHeader{}, err
	}

	return header, nil
}

func listPostgresSalesOrderLines(
	ctx context.Context,
	queryer postgresSalesOrderQueryer,
	persistedID string,
) ([]salesdomain.SalesOrderLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSalesOrderLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]salesdomain.SalesOrderLine, 0)
	for rows.Next() {
		line, err := scanPostgresSalesOrderLine(rows)
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

func scanPostgresSalesOrderLine(row interface{ Scan(dest ...any) error }) (salesdomain.SalesOrderLine, error) {
	var (
		id                 string
		lineNo             int
		itemID             string
		skuCode            string
		itemName           string
		orderedQtyText     string
		uomCode            string
		baseOrderedQtyText string
		baseUOMCode        string
		conversionText     string
		unitPriceText      string
		currencyCode       string
		lineDiscountText   string
		lineAmountText     string
		reservedQtyText    string
		shippedQtyText     string
		batchID            string
		batchNo            string
	)
	if err := row.Scan(
		&id,
		&lineNo,
		&itemID,
		&skuCode,
		&itemName,
		&orderedQtyText,
		&uomCode,
		&baseOrderedQtyText,
		&baseUOMCode,
		&conversionText,
		&unitPriceText,
		&currencyCode,
		&lineDiscountText,
		&lineAmountText,
		&reservedQtyText,
		&shippedQtyText,
		&batchID,
		&batchNo,
	); err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	orderedQty, err := decimal.ParseQuantity(orderedQtyText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	baseOrderedQty, err := decimal.ParseQuantity(baseOrderedQtyText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	conversion, err := decimal.ParseQuantity(conversionText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	unitPrice, err := decimal.ParseUnitPrice(unitPriceText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	lineDiscount, err := decimal.ParseMoneyAmount(lineDiscountText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	lineAmount, err := decimal.ParseMoneyAmount(lineAmountText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	reservedQty, err := decimal.ParseQuantity(reservedQtyText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	shippedQty, err := decimal.ParseQuantity(shippedQtyText)
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	line, err := salesdomain.NewSalesOrderLine(salesdomain.NewSalesOrderLineInput{
		ID:                 id,
		LineNo:             lineNo,
		ItemID:             itemID,
		SKUCode:            skuCode,
		ItemName:           itemName,
		OrderedQty:         orderedQty,
		UOMCode:            uomCode,
		BaseOrderedQty:     baseOrderedQty,
		BaseUOMCode:        baseUOMCode,
		ConversionFactor:   conversion,
		UnitPrice:          unitPrice,
		CurrencyCode:       currencyCode,
		LineDiscountAmount: lineDiscount,
		ReservedQty:        reservedQty,
		ShippedQty:         shippedQty,
		BatchID:            batchID,
		BatchNo:            batchNo,
	})
	if err != nil {
		return salesdomain.SalesOrderLine{}, err
	}
	line.LineAmount = lineAmount
	if err := line.Validate(); err != nil {
		return salesdomain.SalesOrderLine{}, err
	}

	return line, nil
}

func buildPostgresSalesOrder(
	header postgresSalesOrderHeader,
	lines []salesdomain.SalesOrderLine,
) (salesdomain.SalesOrder, error) {
	lineInputs := make([]salesdomain.NewSalesOrderLineInput, 0, len(lines))
	for _, line := range lines {
		lineInputs = append(lineInputs, salesdomain.NewSalesOrderLineInput{
			ID:                 line.ID,
			LineNo:             line.LineNo,
			ItemID:             line.ItemID,
			SKUCode:            line.SKUCode,
			ItemName:           line.ItemName,
			OrderedQty:         line.OrderedQty,
			UOMCode:            line.UOMCode.String(),
			BaseOrderedQty:     line.BaseOrderedQty,
			BaseUOMCode:        line.BaseUOMCode.String(),
			ConversionFactor:   line.ConversionFactor,
			UnitPrice:          line.UnitPrice,
			CurrencyCode:       line.CurrencyCode.String(),
			LineDiscountAmount: line.LineDiscountAmount,
			ReservedQty:        line.ReservedQty,
			ShippedQty:         line.ShippedQty,
			BatchID:            line.BatchID,
			BatchNo:            line.BatchNo,
		})
	}
	order, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            header.ID,
		OrgID:         header.OrgID,
		OrderNo:       header.OrderNo,
		CustomerID:    header.CustomerID,
		CustomerCode:  header.CustomerCode,
		CustomerName:  header.CustomerName,
		Channel:       header.Channel,
		WarehouseID:   header.WarehouseID,
		WarehouseCode: header.WarehouseCode,
		OrderDate:     header.OrderDate,
		CurrencyCode:  header.CurrencyCode,
		Note:          header.Note,
		Lines:         lineInputs,
		CreatedAt:     header.CreatedAt,
		CreatedBy:     firstNonBlank(header.CreatedBy, "system"),
		UpdatedAt:     header.UpdatedAt,
	})
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.Lines = lines
	order.Status = salesdomain.SalesOrderStatus(header.Status)
	order.SubtotalAmount, err = decimal.ParseMoneyAmount(header.SubtotalAmount)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.DiscountAmount, err = decimal.ParseMoneyAmount(header.DiscountAmount)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.TaxAmount, err = decimal.ParseMoneyAmount(header.TaxAmount)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.ShippingFeeAmount, err = decimal.ParseMoneyAmount(header.ShippingFeeAmount)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.NetAmount, err = decimal.ParseMoneyAmount(header.NetAmount)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.TotalAmount, err = decimal.ParseMoneyAmount(header.TotalAmount)
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	order.UpdatedBy = firstNonBlank(header.UpdatedBy, header.CreatedBy, "system")
	order.Version = header.Version
	order.CancelReason = strings.TrimSpace(header.CancelReason)
	order.ConfirmedAt = nullableSalesOrderTimeValue(header.ConfirmedAt)
	order.ConfirmedBy = strings.TrimSpace(header.ConfirmedBy)
	order.ReservedAt = nullableSalesOrderTimeValue(header.ReservedAt)
	order.ReservedBy = strings.TrimSpace(header.ReservedBy)
	order.PickingStartedAt = nullableSalesOrderTimeValue(header.PickingStartedAt)
	order.PickingStartedBy = strings.TrimSpace(header.PickingStartedBy)
	order.PickedAt = nullableSalesOrderTimeValue(header.PickedAt)
	order.PickedBy = strings.TrimSpace(header.PickedBy)
	order.PackingStartedAt = nullableSalesOrderTimeValue(header.PackingStartedAt)
	order.PackingStartedBy = strings.TrimSpace(header.PackingStartedBy)
	order.PackedAt = nullableSalesOrderTimeValue(header.PackedAt)
	order.PackedBy = strings.TrimSpace(header.PackedBy)
	order.WaitingHandoverAt = nullableSalesOrderTimeValue(header.WaitingHandoverAt)
	order.WaitingHandoverBy = strings.TrimSpace(header.WaitingHandoverBy)
	order.HandedOverAt = nullableSalesOrderTimeValue(header.HandedOverAt)
	order.HandedOverBy = strings.TrimSpace(header.HandedOverBy)
	order.ClosedAt = nullableSalesOrderTimeValue(header.ClosedAt)
	order.ClosedBy = strings.TrimSpace(header.ClosedBy)
	order.CancelledAt = nullableSalesOrderTimeValue(header.CancelledAt)
	order.CancelledBy = strings.TrimSpace(header.CancelledBy)
	order.ExceptionAt = nullableSalesOrderTimeValue(header.ExceptionAt)
	order.ExceptionBy = strings.TrimSpace(header.ExceptionBy)
	if err := order.Validate(); err != nil {
		return salesdomain.SalesOrder{}, err
	}

	return order, nil
}

func upsertSalesOrder(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	order salesdomain.SalesOrder,
) (string, error) {
	var persistedID string
	err := tx.QueryRowContext(
		ctx,
		upsertSalesOrderSQL,
		nullableSalesOrderUUID(order.ID),
		orgID,
		nullableSalesOrderText(order.ID),
		nullableSalesOrderText(order.OrgID),
		order.OrderNo,
		nullableSalesOrderUUID(order.CustomerID),
		nullableSalesOrderText(order.CustomerID),
		nullableSalesOrderText(order.CustomerCode),
		order.CustomerName,
		order.OrderDate,
		order.Channel,
		string(order.Status),
		order.CurrencyCode.String(),
		order.SubtotalAmount.String(),
		order.DiscountAmount.String(),
		order.TaxAmount.String(),
		order.ShippingFeeAmount.String(),
		order.NetAmount.String(),
		order.TotalAmount.String(),
		nullableSalesOrderUUID(order.WarehouseID),
		nullableSalesOrderText(order.WarehouseID),
		nullableSalesOrderText(order.WarehouseCode),
		nullableSalesOrderText(order.Note),
		order.CreatedAt.UTC(),
		nullableSalesOrderUUID(order.CreatedBy),
		nullableSalesOrderText(order.CreatedBy),
		order.UpdatedAt.UTC(),
		nullableSalesOrderUUID(order.UpdatedBy),
		nullableSalesOrderText(order.UpdatedBy),
		order.Version,
		nullableSalesOrderText(order.CancelReason),
		nullableSalesOrderTime(order.CancelledAt),
		nullableSalesOrderUUID(order.CancelledBy),
		nullableSalesOrderText(order.CancelledBy),
		nullableSalesOrderTime(order.ConfirmedAt),
		nullableSalesOrderUUID(order.ConfirmedBy),
		nullableSalesOrderText(order.ConfirmedBy),
		nullableSalesOrderTime(order.ReservedAt),
		nullableSalesOrderUUID(order.ReservedBy),
		nullableSalesOrderText(order.ReservedBy),
		nullableSalesOrderTime(order.PickingStartedAt),
		nullableSalesOrderUUID(order.PickingStartedBy),
		nullableSalesOrderText(order.PickingStartedBy),
		nullableSalesOrderTime(order.PickedAt),
		nullableSalesOrderUUID(order.PickedBy),
		nullableSalesOrderText(order.PickedBy),
		nullableSalesOrderTime(order.PackingStartedAt),
		nullableSalesOrderUUID(order.PackingStartedBy),
		nullableSalesOrderText(order.PackingStartedBy),
		nullableSalesOrderTime(order.PackedAt),
		nullableSalesOrderUUID(order.PackedBy),
		nullableSalesOrderText(order.PackedBy),
		nullableSalesOrderTime(order.WaitingHandoverAt),
		nullableSalesOrderUUID(order.WaitingHandoverBy),
		nullableSalesOrderText(order.WaitingHandoverBy),
		nullableSalesOrderTime(order.HandedOverAt),
		nullableSalesOrderUUID(order.HandedOverBy),
		nullableSalesOrderText(order.HandedOverBy),
		nullableSalesOrderTime(order.ClosedAt),
		nullableSalesOrderUUID(order.ClosedBy),
		nullableSalesOrderText(order.ClosedBy),
		nullableSalesOrderTime(order.ExceptionAt),
		nullableSalesOrderUUID(order.ExceptionBy),
		nullableSalesOrderText(order.ExceptionBy),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert sales order: %w", err)
	}

	return persistedID, nil
}

func replaceSalesOrderLines(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	order salesdomain.SalesOrder,
) error {
	if _, err := tx.ExecContext(ctx, deleteSalesOrderLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete sales order lines: %w", err)
	}
	for _, line := range order.Lines {
		if _, err := tx.ExecContext(
			ctx,
			insertSalesOrderLineSQL,
			nullableSalesOrderUUID(line.ID),
			orgID,
			persistedID,
			nullableSalesOrderText(line.ID),
			line.LineNo,
			nullableSalesOrderUUID(line.ItemID),
			nullableSalesOrderText(line.ItemID),
			line.SKUCode,
			line.ItemName,
			nullableSalesOrderUUID(line.BatchID),
			nullableSalesOrderText(line.BatchID),
			nullableSalesOrderText(line.BatchNo),
			line.OrderedQty.String(),
			line.ReservedQty.String(),
			line.ShippedQty.String(),
			line.UnitPrice.String(),
			line.UOMCode.String(),
			line.BaseOrderedQty.String(),
			line.BaseUOMCode.String(),
			line.ConversionFactor.String(),
			line.CurrencyCode.String(),
			line.LineDiscountAmount.String(),
			line.LineAmount.String(),
			order.CreatedAt.UTC(),
			nullableSalesOrderUUID(order.CreatedBy),
			order.UpdatedAt.UTC(),
			nullableSalesOrderUUID(order.UpdatedBy),
		); err != nil {
			return fmt.Errorf("insert sales order line: %w", err)
		}
	}

	return nil
}

func postgresSalesOrderJSONMap(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}

	return requiredPostgresSalesOrderJSONMap(value)
}

func requiredPostgresSalesOrderJSONMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func nullableSalesOrderText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullableSalesOrderUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isSalesOrderUUIDText(value) {
		return nil
	}

	return value
}

func nullableSalesOrderTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullableSalesOrderTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isSalesOrderUUIDText(value string) bool {
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
			if !isSalesOrderHexText(char) {
				return false
			}
		}
	}

	return true
}

func isSalesOrderHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SalesOrderStore = PostgresSalesOrderStore{}
