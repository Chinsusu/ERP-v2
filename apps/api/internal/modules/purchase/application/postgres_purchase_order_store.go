package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	purchasedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresPurchaseOrderStoreConfig struct {
	DefaultOrgID string
}

type PostgresPurchaseOrderStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresPurchaseOrderTx struct {
	store PostgresPurchaseOrderStore
	tx    *sql.Tx
}

type postgresPurchaseOrderQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresPurchaseOrderHeader struct {
	PersistedID         string
	ID                  string
	OrgID               string
	PONo                string
	SupplierID          string
	SupplierCode        string
	SupplierName        string
	WarehouseID         string
	WarehouseCode       string
	ExpectedDate        string
	Status              string
	CurrencyCode        string
	SubtotalAmount      string
	TotalAmount         string
	Note                string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
	CancelReason        string
	RejectReason        string
	SubmittedAt         sql.NullTime
	SubmittedBy         string
	ApprovedAt          sql.NullTime
	ApprovedBy          string
	PartiallyReceivedAt sql.NullTime
	PartiallyReceivedBy string
	ReceivedAt          sql.NullTime
	ReceivedBy          string
	ClosedAt            sql.NullTime
	ClosedBy            string
	CancelledAt         sql.NullTime
	CancelledBy         string
	RejectedAt          sql.NullTime
	RejectedBy          string
}

func NewPostgresPurchaseOrderStore(db *sql.DB, cfg PostgresPurchaseOrderStoreConfig) PostgresPurchaseOrderStore {
	return PostgresPurchaseOrderStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectPurchaseOrderOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectPurchaseOrderHeadersBaseSQL = `
SELECT
  id::text,
  COALESCE(po_ref, id::text),
  COALESCE(org_ref, org_id::text),
  po_no,
  COALESCE(supplier_ref, supplier_id::text, ''),
  COALESCE(supplier_code, ''),
  COALESCE(supplier_name, ''),
  COALESCE(warehouse_ref, warehouse_id::text, ''),
  COALESCE(warehouse_code, ''),
  expected_date::text,
  status,
  currency_code,
  subtotal_amount::text,
  total_amount::text,
  COALESCE(note, ''),
  created_at,
  COALESCE(created_by_ref, created_by::text, ''),
  updated_at,
  COALESCE(updated_by_ref, updated_by::text, ''),
  version,
  COALESCE(cancel_reason, ''),
  COALESCE(reject_reason, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  approved_at,
  COALESCE(approved_by_ref, approved_by::text, ''),
  partially_received_at,
  COALESCE(partially_received_by_ref, partially_received_by::text, ''),
  received_at,
  COALESCE(received_by_ref, received_by::text, ''),
  closed_at,
  COALESCE(closed_by_ref, closed_by::text, ''),
  cancelled_at,
  COALESCE(cancelled_by_ref, cancelled_by::text, ''),
  rejected_at,
  COALESCE(rejected_by_ref, rejected_by::text, '')
FROM purchase.purchase_orders`

const selectPurchaseOrderHeadersSQL = selectPurchaseOrderHeadersBaseSQL + `
ORDER BY expected_date DESC, po_no ASC`

const findPurchaseOrderHeaderSQL = selectPurchaseOrderHeadersBaseSQL + `
WHERE po_ref = $1 OR id::text = $1
LIMIT 1`

const findPurchaseOrderHeaderForUpdateSQL = findPurchaseOrderHeaderSQL + `
FOR UPDATE`

const selectPurchaseOrderLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  line.line_no,
  COALESCE(line.item_ref, line.item_id::text, ''),
  COALESCE(line.sku_code, item.sku, ''),
  COALESCE(line.item_name, item.name, ''),
  line.ordered_qty::text,
  line.received_qty::text,
  line.uom_code,
  line.base_ordered_qty::text,
  line.base_received_qty::text,
  line.base_uom_code,
  line.conversion_factor::text,
  line.unit_price::text,
  line.currency_code,
  line.line_amount::text,
  line.expected_date::text,
  COALESCE(line.note, '')
FROM purchase.purchase_order_lines AS line
LEFT JOIN mdm.items AS item ON item.id = line.item_id
WHERE line.purchase_order_id = $1::uuid
ORDER BY line.line_no, line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertPurchaseOrderSQL = `
INSERT INTO purchase.purchase_orders (
  id,
  org_id,
  po_ref,
  org_ref,
  po_no,
  supplier_id,
  supplier_ref,
  supplier_code,
  supplier_name,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  order_date,
  expected_date,
  status,
  currency_code,
  subtotal_amount,
  total_amount,
  note,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref,
  version,
  cancel_reason,
  reject_reason,
  submitted_at,
  submitted_by,
  submitted_by_ref,
  approved_at,
  approved_by,
  approved_by_ref,
  partially_received_at,
  partially_received_by,
  partially_received_by_ref,
  received_at,
  received_by,
  received_by_ref,
  closed_at,
  closed_by,
  closed_by_ref,
  cancelled_at,
  cancelled_by,
  cancelled_by_ref,
  rejected_at,
  rejected_by,
  rejected_by_ref
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
  $10::uuid,
  $11,
  $12,
  $13::date,
  $14::date,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21::uuid,
  $22,
  $23,
  $24::uuid,
  $25,
  $26,
  $27,
  $28,
  $29,
  $30::uuid,
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
  $49
)
ON CONFLICT (org_id, po_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  po_no = EXCLUDED.po_no,
  supplier_id = EXCLUDED.supplier_id,
  supplier_ref = EXCLUDED.supplier_ref,
  supplier_code = EXCLUDED.supplier_code,
  supplier_name = EXCLUDED.supplier_name,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  order_date = EXCLUDED.order_date,
  expected_date = EXCLUDED.expected_date,
  status = EXCLUDED.status,
  currency_code = EXCLUDED.currency_code,
  subtotal_amount = EXCLUDED.subtotal_amount,
  total_amount = EXCLUDED.total_amount,
  note = EXCLUDED.note,
  created_at = EXCLUDED.created_at,
  created_by = EXCLUDED.created_by,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version,
  cancel_reason = EXCLUDED.cancel_reason,
  reject_reason = EXCLUDED.reject_reason,
  submitted_at = EXCLUDED.submitted_at,
  submitted_by = EXCLUDED.submitted_by,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  approved_at = EXCLUDED.approved_at,
  approved_by = EXCLUDED.approved_by,
  approved_by_ref = EXCLUDED.approved_by_ref,
  partially_received_at = EXCLUDED.partially_received_at,
  partially_received_by = EXCLUDED.partially_received_by,
  partially_received_by_ref = EXCLUDED.partially_received_by_ref,
  received_at = EXCLUDED.received_at,
  received_by = EXCLUDED.received_by,
  received_by_ref = EXCLUDED.received_by_ref,
  closed_at = EXCLUDED.closed_at,
  closed_by = EXCLUDED.closed_by,
  closed_by_ref = EXCLUDED.closed_by_ref,
  cancelled_at = EXCLUDED.cancelled_at,
  cancelled_by = EXCLUDED.cancelled_by,
  cancelled_by_ref = EXCLUDED.cancelled_by_ref,
  rejected_at = EXCLUDED.rejected_at,
  rejected_by = EXCLUDED.rejected_by,
  rejected_by_ref = EXCLUDED.rejected_by_ref
RETURNING id::text`

const deletePurchaseOrderLinesSQL = `
DELETE FROM purchase.purchase_order_lines
WHERE purchase_order_id = $1::uuid`

const insertPurchaseOrderLineSQL = `
INSERT INTO purchase.purchase_order_lines (
  id,
  org_id,
  purchase_order_id,
  line_ref,
  line_no,
  item_id,
  item_ref,
  sku_code,
  item_name,
  ordered_qty,
  received_qty,
  unit_price,
  uom_code,
  base_ordered_qty,
  base_received_qty,
  base_uom_code,
  conversion_factor,
  currency_code,
  line_amount,
  expected_date,
  note,
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
  $20::date,
  $21,
  $22,
  $23::uuid,
  $24,
  $25::uuid
)`

const insertPurchaseOrderAuditSQL = `
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

func (s PostgresPurchaseOrderStore) List(
	ctx context.Context,
	filter PurchaseOrderFilter,
) ([]purchasedomain.PurchaseOrder, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectPurchaseOrderHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]purchasedomain.PurchaseOrder, 0)
	for rows.Next() {
		order, err := scanPostgresPurchaseOrder(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if purchaseOrderMatchesFilter(order, filter) {
			orders = append(orders, order)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortPurchaseOrders(orders)

	return orders, nil
}

func (s PostgresPurchaseOrderStore) Get(
	ctx context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	if s.db == nil {
		return purchasedomain.PurchaseOrder{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findPurchaseOrderHeaderSQL, strings.TrimSpace(id))
	order, err := scanPostgresPurchaseOrder(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return purchasedomain.PurchaseOrder{}, ErrPurchaseOrderNotFound
	}
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}

	return order, nil
}

func (s PostgresPurchaseOrderStore) WithinTx(
	ctx context.Context,
	fn func(context.Context, PurchaseOrderTx) error,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if fn == nil {
		return errors.New("purchase order transaction function is required")
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin purchase order transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, postgresPurchaseOrderTx{store: s, tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit purchase order transaction: %w", err)
	}
	committed = true

	return nil
}

func (tx postgresPurchaseOrderTx) GetForUpdate(
	ctx context.Context,
	id string,
) (purchasedomain.PurchaseOrder, error) {
	if tx.tx == nil {
		return purchasedomain.PurchaseOrder{}, errors.New("database transaction is required")
	}
	row := tx.tx.QueryRowContext(ctx, findPurchaseOrderHeaderForUpdateSQL, strings.TrimSpace(id))
	order, err := scanPostgresPurchaseOrder(ctx, tx.tx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return purchasedomain.PurchaseOrder{}, ErrPurchaseOrderNotFound
	}
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}

	return order, nil
}

func (tx postgresPurchaseOrderTx) Save(ctx context.Context, order purchasedomain.PurchaseOrder) error {
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
	persistedID, err := upsertPurchaseOrder(ctx, tx.tx, orgID, order)
	if err != nil {
		return err
	}
	if err := replacePurchaseOrderLines(ctx, tx.tx, orgID, persistedID, order); err != nil {
		return err
	}

	return nil
}

func (tx postgresPurchaseOrderTx) RecordAudit(ctx context.Context, log audit.Log) error {
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
	orgID, err := tx.store.resolveOrgID(ctx, tx.tx, firstNonBlankPurchaseOrder(normalizedLog.OrgID, defaultPurchaseOrderOrgID))
	if err != nil {
		return err
	}
	beforeJSON, err := postgresPurchaseOrderJSONMap(normalizedLog.BeforeData)
	if err != nil {
		return fmt.Errorf("encode purchase order audit before_data: %w", err)
	}
	afterJSON, err := postgresPurchaseOrderJSONMap(normalizedLog.AfterData)
	if err != nil {
		return fmt.Errorf("encode purchase order audit after_data: %w", err)
	}
	metadataJSON, err := requiredPostgresPurchaseOrderJSONMap(normalizedLog.Metadata)
	if err != nil {
		return fmt.Errorf("encode purchase order audit metadata: %w", err)
	}

	_, err = tx.tx.ExecContext(
		ctx,
		insertPurchaseOrderAuditSQL,
		nullablePurchaseOrderUUID(normalizedLog.ID),
		orgID,
		nullablePurchaseOrderUUID(normalizedLog.ActorID),
		normalizedLog.Action,
		normalizedLog.EntityType,
		nullablePurchaseOrderUUID(normalizedLog.EntityID),
		nullablePurchaseOrderText(normalizedLog.RequestID),
		beforeJSON,
		afterJSON,
		metadataJSON,
		normalizedLog.CreatedAt.UTC(),
		nullablePurchaseOrderText(normalizedLog.ID),
		nullablePurchaseOrderText(normalizedLog.OrgID),
		nullablePurchaseOrderText(normalizedLog.ActorID),
		nullablePurchaseOrderText(normalizedLog.EntityID),
	)
	if err != nil {
		return fmt.Errorf("insert purchase order audit: %w", err)
	}

	return nil
}

func (s PostgresPurchaseOrderStore) resolveOrgID(
	ctx context.Context,
	queryer postgresPurchaseOrderQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPurchaseOrderUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectPurchaseOrderOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPurchaseOrderUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve purchase order org %q: %w", orgRef, err)
		}
	}
	if isPurchaseOrderUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("purchase order org %q cannot be resolved", orgRef)
}

func scanPostgresPurchaseOrder(
	ctx context.Context,
	queryer postgresPurchaseOrderQueryer,
	row interface{ Scan(dest ...any) error },
) (purchasedomain.PurchaseOrder, error) {
	header, err := scanPostgresPurchaseOrderHeader(row)
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}
	lines, err := listPostgresPurchaseOrderLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}

	return buildPostgresPurchaseOrder(header, lines)
}

func scanPostgresPurchaseOrderHeader(row interface{ Scan(dest ...any) error }) (postgresPurchaseOrderHeader, error) {
	var header postgresPurchaseOrderHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.PONo,
		&header.SupplierID,
		&header.SupplierCode,
		&header.SupplierName,
		&header.WarehouseID,
		&header.WarehouseCode,
		&header.ExpectedDate,
		&header.Status,
		&header.CurrencyCode,
		&header.SubtotalAmount,
		&header.TotalAmount,
		&header.Note,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
		&header.CancelReason,
		&header.RejectReason,
		&header.SubmittedAt,
		&header.SubmittedBy,
		&header.ApprovedAt,
		&header.ApprovedBy,
		&header.PartiallyReceivedAt,
		&header.PartiallyReceivedBy,
		&header.ReceivedAt,
		&header.ReceivedBy,
		&header.ClosedAt,
		&header.ClosedBy,
		&header.CancelledAt,
		&header.CancelledBy,
		&header.RejectedAt,
		&header.RejectedBy,
	)
	if err != nil {
		return postgresPurchaseOrderHeader{}, err
	}

	return header, nil
}

func listPostgresPurchaseOrderLines(
	ctx context.Context,
	queryer postgresPurchaseOrderQueryer,
	persistedID string,
) ([]purchasedomain.PurchaseOrderLine, error) {
	rows, err := queryer.QueryContext(ctx, selectPurchaseOrderLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]purchasedomain.PurchaseOrderLine, 0)
	for rows.Next() {
		line, err := scanPostgresPurchaseOrderLine(rows)
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

func scanPostgresPurchaseOrderLine(row interface{ Scan(dest ...any) error }) (purchasedomain.PurchaseOrderLine, error) {
	var (
		id                  string
		lineNo              int
		itemID              string
		skuCode             string
		itemName            string
		orderedQtyText      string
		receivedQtyText     string
		uomCode             string
		baseOrderedQtyText  string
		baseReceivedQtyText string
		baseUOMCode         string
		conversionText      string
		unitPriceText       string
		currencyCode        string
		lineAmountText      string
		expectedDate        string
		note                string
	)
	if err := row.Scan(
		&id,
		&lineNo,
		&itemID,
		&skuCode,
		&itemName,
		&orderedQtyText,
		&receivedQtyText,
		&uomCode,
		&baseOrderedQtyText,
		&baseReceivedQtyText,
		&baseUOMCode,
		&conversionText,
		&unitPriceText,
		&currencyCode,
		&lineAmountText,
		&expectedDate,
		&note,
	); err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	orderedQty, err := decimal.ParseQuantity(orderedQtyText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	receivedQty, err := decimal.ParseQuantity(receivedQtyText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	baseOrderedQty, err := decimal.ParseQuantity(baseOrderedQtyText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	baseReceivedQty, err := decimal.ParseQuantity(baseReceivedQtyText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	conversion, err := decimal.ParseQuantity(conversionText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	unitPrice, err := decimal.ParseUnitPrice(unitPriceText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	lineAmount, err := decimal.ParseMoneyAmount(lineAmountText)
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	line, err := purchasedomain.NewPurchaseOrderLine(purchasedomain.NewPurchaseOrderLineInput{
		ID:               id,
		LineNo:           lineNo,
		ItemID:           itemID,
		SKUCode:          skuCode,
		ItemName:         itemName,
		OrderedQty:       orderedQty,
		ReceivedQty:      receivedQty,
		UOMCode:          uomCode,
		BaseOrderedQty:   baseOrderedQty,
		BaseReceivedQty:  baseReceivedQty,
		BaseUOMCode:      baseUOMCode,
		ConversionFactor: conversion,
		UnitPrice:        unitPrice,
		CurrencyCode:     currencyCode,
		ExpectedDate:     expectedDate,
		Note:             note,
	})
	if err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}
	line.LineAmount = lineAmount
	if err := line.Validate(); err != nil {
		return purchasedomain.PurchaseOrderLine{}, err
	}

	return line, nil
}

func buildPostgresPurchaseOrder(
	header postgresPurchaseOrderHeader,
	lines []purchasedomain.PurchaseOrderLine,
) (purchasedomain.PurchaseOrder, error) {
	lineInputs := make([]purchasedomain.NewPurchaseOrderLineInput, 0, len(lines))
	for _, line := range lines {
		lineInputs = append(lineInputs, purchasedomain.NewPurchaseOrderLineInput{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			ItemName:         line.ItemName,
			OrderedQty:       line.OrderedQty,
			ReceivedQty:      line.ReceivedQty,
			UOMCode:          line.UOMCode.String(),
			BaseOrderedQty:   line.BaseOrderedQty,
			BaseReceivedQty:  line.BaseReceivedQty,
			BaseUOMCode:      line.BaseUOMCode.String(),
			ConversionFactor: line.ConversionFactor,
			UnitPrice:        line.UnitPrice,
			CurrencyCode:     line.CurrencyCode.String(),
			ExpectedDate:     line.ExpectedDate,
			Note:             line.Note,
		})
	}
	order, err := purchasedomain.NewPurchaseOrderDocument(purchasedomain.NewPurchaseOrderDocumentInput{
		ID:            header.ID,
		OrgID:         header.OrgID,
		PONo:          header.PONo,
		SupplierID:    header.SupplierID,
		SupplierCode:  header.SupplierCode,
		SupplierName:  header.SupplierName,
		WarehouseID:   header.WarehouseID,
		WarehouseCode: header.WarehouseCode,
		ExpectedDate:  header.ExpectedDate,
		CurrencyCode:  header.CurrencyCode,
		Note:          header.Note,
		Lines:         lineInputs,
		CreatedAt:     header.CreatedAt,
		CreatedBy:     firstNonBlankPurchaseOrder(header.CreatedBy, "system"),
		UpdatedAt:     header.UpdatedAt,
		UpdatedBy:     firstNonBlankPurchaseOrder(header.UpdatedBy, header.CreatedBy, "system"),
	})
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}
	order.Lines = lines
	order.Status = purchasedomain.PurchaseOrderStatus(header.Status)
	order.SubtotalAmount, err = decimal.ParseMoneyAmount(header.SubtotalAmount)
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}
	order.TotalAmount, err = decimal.ParseMoneyAmount(header.TotalAmount)
	if err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}
	order.Version = header.Version
	order.CancelReason = strings.TrimSpace(header.CancelReason)
	order.RejectReason = strings.TrimSpace(header.RejectReason)
	order.SubmittedAt = nullablePurchaseOrderTimeValue(header.SubmittedAt)
	order.SubmittedBy = strings.TrimSpace(header.SubmittedBy)
	order.ApprovedAt = nullablePurchaseOrderTimeValue(header.ApprovedAt)
	order.ApprovedBy = strings.TrimSpace(header.ApprovedBy)
	order.PartiallyReceivedAt = nullablePurchaseOrderTimeValue(header.PartiallyReceivedAt)
	order.PartiallyReceivedBy = strings.TrimSpace(header.PartiallyReceivedBy)
	order.ReceivedAt = nullablePurchaseOrderTimeValue(header.ReceivedAt)
	order.ReceivedBy = strings.TrimSpace(header.ReceivedBy)
	order.ClosedAt = nullablePurchaseOrderTimeValue(header.ClosedAt)
	order.ClosedBy = strings.TrimSpace(header.ClosedBy)
	order.CancelledAt = nullablePurchaseOrderTimeValue(header.CancelledAt)
	order.CancelledBy = strings.TrimSpace(header.CancelledBy)
	order.RejectedAt = nullablePurchaseOrderTimeValue(header.RejectedAt)
	order.RejectedBy = strings.TrimSpace(header.RejectedBy)
	if err := order.Validate(); err != nil {
		return purchasedomain.PurchaseOrder{}, err
	}

	return order, nil
}

func upsertPurchaseOrder(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	order purchasedomain.PurchaseOrder,
) (string, error) {
	var persistedID string
	err := tx.QueryRowContext(
		ctx,
		upsertPurchaseOrderSQL,
		nullablePurchaseOrderUUID(order.ID),
		orgID,
		nullablePurchaseOrderText(order.ID),
		nullablePurchaseOrderText(order.OrgID),
		order.PONo,
		nullablePurchaseOrderUUID(order.SupplierID),
		nullablePurchaseOrderText(order.SupplierID),
		nullablePurchaseOrderText(order.SupplierCode),
		order.SupplierName,
		nullablePurchaseOrderUUID(order.WarehouseID),
		nullablePurchaseOrderText(order.WarehouseID),
		nullablePurchaseOrderText(order.WarehouseCode),
		order.ExpectedDate,
		order.ExpectedDate,
		string(order.Status),
		order.CurrencyCode.String(),
		order.SubtotalAmount.String(),
		order.TotalAmount.String(),
		nullablePurchaseOrderText(order.Note),
		order.CreatedAt.UTC(),
		nullablePurchaseOrderUUID(order.CreatedBy),
		nullablePurchaseOrderText(order.CreatedBy),
		order.UpdatedAt.UTC(),
		nullablePurchaseOrderUUID(order.UpdatedBy),
		nullablePurchaseOrderText(order.UpdatedBy),
		order.Version,
		nullablePurchaseOrderText(order.CancelReason),
		nullablePurchaseOrderText(order.RejectReason),
		nullablePurchaseOrderTime(order.SubmittedAt),
		nullablePurchaseOrderUUID(order.SubmittedBy),
		nullablePurchaseOrderText(order.SubmittedBy),
		nullablePurchaseOrderTime(order.ApprovedAt),
		nullablePurchaseOrderUUID(order.ApprovedBy),
		nullablePurchaseOrderText(order.ApprovedBy),
		nullablePurchaseOrderTime(order.PartiallyReceivedAt),
		nullablePurchaseOrderUUID(order.PartiallyReceivedBy),
		nullablePurchaseOrderText(order.PartiallyReceivedBy),
		nullablePurchaseOrderTime(order.ReceivedAt),
		nullablePurchaseOrderUUID(order.ReceivedBy),
		nullablePurchaseOrderText(order.ReceivedBy),
		nullablePurchaseOrderTime(order.ClosedAt),
		nullablePurchaseOrderUUID(order.ClosedBy),
		nullablePurchaseOrderText(order.ClosedBy),
		nullablePurchaseOrderTime(order.CancelledAt),
		nullablePurchaseOrderUUID(order.CancelledBy),
		nullablePurchaseOrderText(order.CancelledBy),
		nullablePurchaseOrderTime(order.RejectedAt),
		nullablePurchaseOrderUUID(order.RejectedBy),
		nullablePurchaseOrderText(order.RejectedBy),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert purchase order: %w", err)
	}

	return persistedID, nil
}

func replacePurchaseOrderLines(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	order purchasedomain.PurchaseOrder,
) error {
	if _, err := tx.ExecContext(ctx, deletePurchaseOrderLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete purchase order lines: %w", err)
	}
	for _, line := range order.Lines {
		if _, err := tx.ExecContext(
			ctx,
			insertPurchaseOrderLineSQL,
			nullablePurchaseOrderUUID(line.ID),
			orgID,
			persistedID,
			nullablePurchaseOrderText(line.ID),
			line.LineNo,
			nullablePurchaseOrderUUID(line.ItemID),
			nullablePurchaseOrderText(line.ItemID),
			line.SKUCode,
			line.ItemName,
			line.OrderedQty.String(),
			line.ReceivedQty.String(),
			line.UnitPrice.String(),
			line.UOMCode.String(),
			line.BaseOrderedQty.String(),
			line.BaseReceivedQty.String(),
			line.BaseUOMCode.String(),
			line.ConversionFactor.String(),
			line.CurrencyCode.String(),
			line.LineAmount.String(),
			line.ExpectedDate,
			nullablePurchaseOrderText(line.Note),
			order.CreatedAt.UTC(),
			nullablePurchaseOrderUUID(order.CreatedBy),
			order.UpdatedAt.UTC(),
			nullablePurchaseOrderUUID(order.UpdatedBy),
		); err != nil {
			return fmt.Errorf("insert purchase order line: %w", err)
		}
	}

	return nil
}

func postgresPurchaseOrderJSONMap(value map[string]any) (any, error) {
	if value == nil {
		return nil, nil
	}

	return requiredPostgresPurchaseOrderJSONMap(value)
}

func requiredPostgresPurchaseOrderJSONMap(value map[string]any) (string, error) {
	if value == nil {
		value = map[string]any{}
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func nullablePurchaseOrderText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePurchaseOrderUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPurchaseOrderUUIDText(value) {
		return nil
	}

	return value
}

func nullablePurchaseOrderTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func nullablePurchaseOrderTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func isPurchaseOrderUUIDText(value string) bool {
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
			if !isPurchaseOrderHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPurchaseOrderHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ PurchaseOrderStore = PostgresPurchaseOrderStore{}
