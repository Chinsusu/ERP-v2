package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresInboundQCInspectionStoreConfig struct {
	DefaultOrgID string
}

type PostgresInboundQCInspectionStore struct {
	db           *sql.DB
	defaultOrgID string
}

func NewPostgresInboundQCInspectionStore(
	db *sql.DB,
	cfg PostgresInboundQCInspectionStoreConfig,
) PostgresInboundQCInspectionStore {
	return PostgresInboundQCInspectionStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectInboundQCOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const listInboundQCInspectionsSQL = `
SELECT
  id::text,
  COALESCE(inspection_ref, id::text),
  COALESCE(org_ref, org_id::text),
  goods_receipt_ref,
  goods_receipt_no,
  goods_receipt_line_ref,
  COALESCE(purchase_order_ref, purchase_order_id::text, ''),
  COALESCE(purchase_order_line_ref, purchase_order_line_id::text, ''),
  COALESCE(item_ref, item_id::text, ''),
  sku_code,
  COALESCE(item_name, ''),
  COALESCE(batch_ref, batch_id::text, ''),
  batch_no,
  lot_no,
  expiry_date::text,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(location_ref, location_id::text),
  quantity::text,
  uom_code,
  COALESCE(inspector_ref, inspector_id::text),
  status,
  COALESCE(result, ''),
  passed_qty::text,
  failed_qty::text,
  hold_qty::text,
  COALESCE(reason, ''),
  COALESCE(note, ''),
  COALESCE(created_by_ref, created_by::text),
  COALESCE(updated_by_ref, updated_by::text, ''),
  started_at,
  COALESCE(started_by_ref, started_by::text, ''),
  decided_at,
  COALESCE(decided_by_ref, decided_by::text, ''),
  created_at,
  updated_at
FROM qc.inbound_qc_inspections
ORDER BY updated_at DESC, inspection_ref ASC`

const findInboundQCInspectionPredicateSQL = `
SELECT
  id::text,
  COALESCE(inspection_ref, id::text),
  COALESCE(org_ref, org_id::text),
  goods_receipt_ref,
  goods_receipt_no,
  goods_receipt_line_ref,
  COALESCE(purchase_order_ref, purchase_order_id::text, ''),
  COALESCE(purchase_order_line_ref, purchase_order_line_id::text, ''),
  COALESCE(item_ref, item_id::text, ''),
  sku_code,
  COALESCE(item_name, ''),
  COALESCE(batch_ref, batch_id::text, ''),
  batch_no,
  lot_no,
  expiry_date::text,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(location_ref, location_id::text),
  quantity::text,
  uom_code,
  COALESCE(inspector_ref, inspector_id::text),
  status,
  COALESCE(result, ''),
  passed_qty::text,
  failed_qty::text,
  hold_qty::text,
  COALESCE(reason, ''),
  COALESCE(note, ''),
  COALESCE(created_by_ref, created_by::text),
  COALESCE(updated_by_ref, updated_by::text, ''),
  started_at,
  COALESCE(started_by_ref, started_by::text, ''),
  decided_at,
  COALESCE(decided_by_ref, decided_by::text, ''),
  created_at,
  updated_at
FROM qc.inbound_qc_inspections
WHERE inspection_ref = $1 OR id::text = $1
LIMIT 1`

const selectInboundQCChecklistItemsSQL = `
SELECT
  COALESCE(checklist_ref, id::text),
  code,
  label,
  required,
  status,
  COALESCE(note, '')
FROM qc.inbound_qc_checklist_items
WHERE inspection_id = $1::uuid
ORDER BY checklist_no, created_at, checklist_ref`

const upsertInboundQCInspectionSQL = `
INSERT INTO qc.inbound_qc_inspections (
  id,
  org_id,
  inspection_ref,
  org_ref,
  goods_receipt_id,
  goods_receipt_ref,
  goods_receipt_no,
  goods_receipt_line_id,
  goods_receipt_line_ref,
  purchase_order_id,
  purchase_order_ref,
  purchase_order_line_id,
  purchase_order_line_ref,
  item_id,
  item_ref,
  sku_code,
  item_name,
  batch_id,
  batch_ref,
  batch_no,
  lot_no,
  expiry_date,
  warehouse_id,
  warehouse_ref,
  location_id,
  location_ref,
  quantity,
  uom_code,
  inspector_id,
  inspector_ref,
  status,
  result,
  passed_qty,
  failed_qty,
  hold_qty,
  reason,
  note,
  created_by,
  created_by_ref,
  updated_by,
  updated_by_ref,
  started_at,
  started_by,
  started_by_ref,
  decided_at,
  decided_by,
  decided_by_ref,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5::uuid,
  $6,
  $7,
  $8::uuid,
  $9,
  $10::uuid,
  $11,
  $12::uuid,
  $13,
  $14::uuid,
  $15,
  $16,
  $17,
  $18::uuid,
  $19,
  $20,
  $21,
  $22::date,
  $23::uuid,
  $24,
  $25::uuid,
  $26,
  $27,
  $28,
  $29::uuid,
  $30,
  $31,
  $32,
  $33,
  $34,
  $35,
  $36,
  $37,
  $38::uuid,
  $39,
  $40::uuid,
  $41,
  $42,
  $43::uuid,
  $44,
  $45,
  $46::uuid,
  $47,
  $48,
  $49
)
ON CONFLICT ON CONSTRAINT uq_inbound_qc_inspections_org_ref
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  goods_receipt_id = EXCLUDED.goods_receipt_id,
  goods_receipt_ref = EXCLUDED.goods_receipt_ref,
  goods_receipt_no = EXCLUDED.goods_receipt_no,
  goods_receipt_line_id = EXCLUDED.goods_receipt_line_id,
  goods_receipt_line_ref = EXCLUDED.goods_receipt_line_ref,
  purchase_order_id = EXCLUDED.purchase_order_id,
  purchase_order_ref = EXCLUDED.purchase_order_ref,
  purchase_order_line_id = EXCLUDED.purchase_order_line_id,
  purchase_order_line_ref = EXCLUDED.purchase_order_line_ref,
  item_id = EXCLUDED.item_id,
  item_ref = EXCLUDED.item_ref,
  sku_code = EXCLUDED.sku_code,
  item_name = EXCLUDED.item_name,
  batch_id = EXCLUDED.batch_id,
  batch_ref = EXCLUDED.batch_ref,
  batch_no = EXCLUDED.batch_no,
  lot_no = EXCLUDED.lot_no,
  expiry_date = EXCLUDED.expiry_date,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  location_id = EXCLUDED.location_id,
  location_ref = EXCLUDED.location_ref,
  quantity = EXCLUDED.quantity,
  uom_code = EXCLUDED.uom_code,
  inspector_id = EXCLUDED.inspector_id,
  inspector_ref = EXCLUDED.inspector_ref,
  status = EXCLUDED.status,
  result = EXCLUDED.result,
  passed_qty = EXCLUDED.passed_qty,
  failed_qty = EXCLUDED.failed_qty,
  hold_qty = EXCLUDED.hold_qty,
  reason = EXCLUDED.reason,
  note = EXCLUDED.note,
  created_by = EXCLUDED.created_by,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  started_at = EXCLUDED.started_at,
  started_by = EXCLUDED.started_by,
  started_by_ref = EXCLUDED.started_by_ref,
  decided_at = EXCLUDED.decided_at,
  decided_by = EXCLUDED.decided_by,
  decided_by_ref = EXCLUDED.decided_by_ref,
  updated_at = EXCLUDED.updated_at
RETURNING id::text`

const deleteInboundQCChecklistItemsSQL = `
DELETE FROM qc.inbound_qc_checklist_items
WHERE inspection_id = $1::uuid`

const insertInboundQCChecklistItemSQL = `
INSERT INTO qc.inbound_qc_checklist_items (
  id,
  org_id,
  inspection_id,
  checklist_ref,
  checklist_no,
  code,
  label,
  required,
  status,
  note,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12
)`

func (s PostgresInboundQCInspectionStore) List(
	ctx context.Context,
	filter InboundQCInspectionFilter,
) ([]qcdomain.InboundQCInspection, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, listInboundQCInspectionsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	filter = NewInboundQCInspectionFilter(filter.Status, filter.GoodsReceiptID, filter.GoodsReceiptLineID, filter.WarehouseID)
	inspections := make([]qcdomain.InboundQCInspection, 0)
	for rows.Next() {
		inspection, err := s.scanInboundQCInspection(ctx, rows)
		if err != nil {
			return nil, err
		}
		if filter.matches(inspection) {
			inspections = append(inspections, inspection)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortInboundQCInspections(inspections)

	return inspections, nil
}

func (s PostgresInboundQCInspectionStore) Get(
	ctx context.Context,
	id string,
) (qcdomain.InboundQCInspection, error) {
	if s.db == nil {
		return qcdomain.InboundQCInspection{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findInboundQCInspectionPredicateSQL, strings.TrimSpace(id))
	inspection, err := s.scanInboundQCInspection(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return qcdomain.InboundQCInspection{}, ErrInboundQCInspectionNotFound
	}
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}

	return inspection, nil
}

func (s PostgresInboundQCInspectionStore) Save(
	ctx context.Context,
	inspection qcdomain.InboundQCInspection,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := inspection.Validate(); err != nil {
		return err
	}
	orgID, err := s.resolveOrgID(ctx, inspection.OrgID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin inbound qc transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, err := upsertInboundQCInspection(ctx, tx, orgID, inspection)
	if err != nil {
		return err
	}
	if err := replaceInboundQCChecklistItems(ctx, tx, orgID, persistedID, inspection); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit inbound qc transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresInboundQCInspectionStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if inboundQCIsUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := s.db.QueryRowContext(ctx, selectInboundQCOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && inboundQCIsUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve inbound qc org %q: %w", orgRef, err)
		}
	}
	if inboundQCIsUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("inbound qc org %q cannot be resolved", orgRef)
}

func (s PostgresInboundQCInspectionStore) scanInboundQCInspection(
	ctx context.Context,
	row interface{ Scan(dest ...any) error },
) (qcdomain.InboundQCInspection, error) {
	var (
		persistedID         string
		id                  string
		orgID               string
		goodsReceiptID      string
		goodsReceiptNo      string
		goodsReceiptLineID  string
		purchaseOrderID     string
		purchaseOrderLineID string
		itemID              string
		sku                 string
		itemName            string
		batchID             string
		batchNo             string
		lotNo               string
		expiryDateText      string
		warehouseID         string
		locationID          string
		quantityText        string
		uomCode             string
		inspectorID         string
		status              string
		result              string
		passedQtyText       string
		failedQtyText       string
		holdQtyText         string
		reason              string
		note                string
		createdBy           string
		updatedBy           string
		startedAt           sql.NullTime
		startedBy           string
		decidedAt           sql.NullTime
		decidedBy           string
		createdAt           time.Time
		updatedAt           time.Time
	)
	if err := row.Scan(
		&persistedID,
		&id,
		&orgID,
		&goodsReceiptID,
		&goodsReceiptNo,
		&goodsReceiptLineID,
		&purchaseOrderID,
		&purchaseOrderLineID,
		&itemID,
		&sku,
		&itemName,
		&batchID,
		&batchNo,
		&lotNo,
		&expiryDateText,
		&warehouseID,
		&locationID,
		&quantityText,
		&uomCode,
		&inspectorID,
		&status,
		&result,
		&passedQtyText,
		&failedQtyText,
		&holdQtyText,
		&reason,
		&note,
		&createdBy,
		&updatedBy,
		&startedAt,
		&startedBy,
		&decidedAt,
		&decidedBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return qcdomain.InboundQCInspection{}, err
	}

	checklist, err := s.listInboundQCChecklistItems(ctx, persistedID)
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}
	quantity, err := decimal.ParseQuantity(quantityText)
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}
	expiryDate, err := time.Parse("2006-01-02", expiryDateText)
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}
	inspection, err := qcdomain.NewInboundQCInspection(qcdomain.NewInboundQCInspectionInput{
		ID:                  id,
		OrgID:               orgID,
		GoodsReceiptID:      goodsReceiptID,
		GoodsReceiptNo:      goodsReceiptNo,
		GoodsReceiptLineID:  goodsReceiptLineID,
		PurchaseOrderID:     purchaseOrderID,
		PurchaseOrderLineID: purchaseOrderLineID,
		ItemID:              itemID,
		SKU:                 sku,
		ItemName:            itemName,
		BatchID:             batchID,
		BatchNo:             batchNo,
		LotNo:               lotNo,
		ExpiryDate:          expiryDate,
		WarehouseID:         warehouseID,
		LocationID:          locationID,
		Quantity:            quantity,
		UOMCode:             uomCode,
		InspectorID:         inspectorID,
		Checklist:           checklist,
		Note:                note,
		CreatedAt:           createdAt,
		CreatedBy:           createdBy,
		UpdatedAt:           updatedAt,
		UpdatedBy:           updatedBy,
	})
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}

	inspection.Status = qcdomain.InboundQCInspectionStatus(status)
	inspection.Result = qcdomain.InboundQCResult(result)
	inspection.PassedQuantity, err = decimal.ParseQuantity(passedQtyText)
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}
	inspection.FailedQuantity, err = decimal.ParseQuantity(failedQtyText)
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}
	inspection.HoldQuantity, err = decimal.ParseQuantity(holdQtyText)
	if err != nil {
		return qcdomain.InboundQCInspection{}, err
	}
	inspection.Reason = strings.TrimSpace(reason)
	inspection.Note = strings.TrimSpace(note)
	if startedAt.Valid {
		inspection.StartedAt = startedAt.Time.UTC()
		inspection.StartedBy = strings.TrimSpace(startedBy)
	}
	if decidedAt.Valid {
		inspection.DecidedAt = decidedAt.Time.UTC()
		inspection.DecidedBy = strings.TrimSpace(decidedBy)
	}
	if err := inspection.Validate(); err != nil {
		return qcdomain.InboundQCInspection{}, err
	}

	return inspection, nil
}

func (s PostgresInboundQCInspectionStore) listInboundQCChecklistItems(
	ctx context.Context,
	inspectionID string,
) ([]qcdomain.NewInboundQCChecklistItemInput, error) {
	rows, err := s.db.QueryContext(ctx, selectInboundQCChecklistItemsSQL, inspectionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]qcdomain.NewInboundQCChecklistItemInput, 0)
	for rows.Next() {
		item, err := scanInboundQCChecklistItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func scanInboundQCChecklistItem(
	row interface{ Scan(dest ...any) error },
) (qcdomain.NewInboundQCChecklistItemInput, error) {
	var (
		id       string
		code     string
		label    string
		required bool
		status   string
		note     string
	)
	if err := row.Scan(&id, &code, &label, &required, &status, &note); err != nil {
		return qcdomain.NewInboundQCChecklistItemInput{}, err
	}

	return qcdomain.NewInboundQCChecklistItemInput{
		ID:       id,
		Code:     code,
		Label:    label,
		Required: required,
		Status:   qcdomain.InboundQCChecklistStatus(status),
		Note:     note,
	}, nil
}

func upsertInboundQCInspection(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	inspection qcdomain.InboundQCInspection,
) (string, error) {
	var persistedID string
	err := tx.QueryRowContext(
		ctx,
		upsertInboundQCInspectionSQL,
		inboundQCNullableUUID(inspection.ID),
		orgID,
		inboundQCNullableText(inspection.ID),
		inboundQCNullableText(inspection.OrgID),
		inboundQCNullableUUID(inspection.GoodsReceiptID),
		inspection.GoodsReceiptID,
		inspection.GoodsReceiptNo,
		inboundQCNullableUUID(inspection.GoodsReceiptLineID),
		inspection.GoodsReceiptLineID,
		inboundQCNullableUUID(inspection.PurchaseOrderID),
		inboundQCNullableText(inspection.PurchaseOrderID),
		inboundQCNullableUUID(inspection.PurchaseOrderLineID),
		inboundQCNullableText(inspection.PurchaseOrderLineID),
		inboundQCNullableUUID(inspection.ItemID),
		inboundQCNullableText(inspection.ItemID),
		inspection.SKU,
		inboundQCNullableText(inspection.ItemName),
		inboundQCNullableUUID(inspection.BatchID),
		inboundQCNullableText(inspection.BatchID),
		inspection.BatchNo,
		inspection.LotNo,
		inboundQCNullableDate(inspection.ExpiryDate),
		inboundQCNullableUUID(inspection.WarehouseID),
		inspection.WarehouseID,
		inboundQCNullableUUID(inspection.LocationID),
		inspection.LocationID,
		inspection.Quantity.String(),
		inspection.UOMCode.String(),
		inboundQCNullableUUID(inspection.InspectorID),
		inspection.InspectorID,
		string(inspection.Status),
		inboundQCNullableText(string(inspection.Result)),
		inboundQCQuantityString(inspection.PassedQuantity),
		inboundQCQuantityString(inspection.FailedQuantity),
		inboundQCQuantityString(inspection.HoldQuantity),
		inboundQCNullableText(inspection.Reason),
		inboundQCNullableText(inspection.Note),
		inboundQCNullableUUID(inspection.CreatedBy),
		inspection.CreatedBy,
		inboundQCNullableUUID(inspection.UpdatedBy),
		inboundQCNullableText(inspection.UpdatedBy),
		inboundQCNullableTime(inspection.StartedAt),
		inboundQCNullableUUID(inspection.StartedBy),
		inboundQCNullableText(inspection.StartedBy),
		inboundQCNullableTime(inspection.DecidedAt),
		inboundQCNullableUUID(inspection.DecidedBy),
		inboundQCNullableText(inspection.DecidedBy),
		inspection.CreatedAt.UTC(),
		inspection.UpdatedAt.UTC(),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert inbound qc inspection: %w", err)
	}

	return persistedID, nil
}

func replaceInboundQCChecklistItems(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	inspection qcdomain.InboundQCInspection,
) error {
	if _, err := tx.ExecContext(ctx, deleteInboundQCChecklistItemsSQL, persistedID); err != nil {
		return fmt.Errorf("delete inbound qc checklist items: %w", err)
	}
	for index, item := range inspection.Checklist {
		if _, err := tx.ExecContext(
			ctx,
			insertInboundQCChecklistItemSQL,
			inboundQCNullableUUID(item.ID),
			orgID,
			persistedID,
			item.ID,
			index+1,
			item.Code,
			item.Label,
			item.Required,
			string(item.Status),
			inboundQCNullableText(item.Note),
			inspection.CreatedAt.UTC(),
			inspection.UpdatedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert inbound qc checklist item: %w", err)
		}
	}

	return nil
}

func inboundQCNullableUUID(value string) any {
	value = strings.TrimSpace(value)
	if !inboundQCIsUUIDText(value) {
		return nil
	}

	return value
}

func inboundQCNullableText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func inboundQCNullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func inboundQCNullableDate(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC().Format("2006-01-02")
}

func inboundQCIsUUIDText(value string) bool {
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
			if !inboundQCIsHexText(char) {
				return false
			}
		}
	}

	return true
}

func inboundQCIsHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ InboundQCInspectionStore = PostgresInboundQCInspectionStore{}
