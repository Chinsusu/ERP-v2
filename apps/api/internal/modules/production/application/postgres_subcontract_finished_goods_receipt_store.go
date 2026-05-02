package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	productiondomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/production/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSubcontractFinishedGoodsReceiptStoreConfig struct {
	DefaultOrgID string
}

type PostgresSubcontractFinishedGoodsReceiptStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSubcontractFinishedGoodsReceiptQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSubcontractFinishedGoodsReceiptHeader struct {
	PersistedID        string
	ID                 string
	OrgID              string
	ReceiptNo          string
	SubcontractOrderID string
	SubcontractOrderNo string
	FactoryID          string
	FactoryCode        string
	FactoryName        string
	WarehouseID        string
	WarehouseCode      string
	LocationID         string
	LocationCode       string
	DeliveryNoteNo     string
	ReceivedBy         string
	ReceivedAt         time.Time
	Note               string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
	Version            int
}

func NewPostgresSubcontractFinishedGoodsReceiptStore(
	db *sql.DB,
	cfg PostgresSubcontractFinishedGoodsReceiptStoreConfig,
) PostgresSubcontractFinishedGoodsReceiptStore {
	return PostgresSubcontractFinishedGoodsReceiptStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSubcontractFinishedGoodsReceiptOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSubcontractFinishedGoodsReceiptHeadersBaseSQL = `
SELECT
  receipt.id::text,
  receipt.receipt_ref,
  receipt.org_ref,
  receipt.receipt_no,
  receipt.subcontract_order_ref,
  receipt.subcontract_order_no,
  COALESCE(subcontract_order.factory_ref, ''),
  COALESCE(subcontract_order.factory_code, ''),
  COALESCE(subcontract_order.factory_name, ''),
  receipt.warehouse_ref,
  COALESCE(receipt.warehouse_code, ''),
  receipt.location_ref,
  COALESCE(receipt.location_code, ''),
  COALESCE(receipt.delivery_note_no, ''),
  receipt.received_by_ref,
  receipt.received_at,
  COALESCE(receipt.note, ''),
  receipt.created_at,
  receipt.created_by_ref,
  receipt.updated_at,
  receipt.updated_by_ref,
  receipt.version
FROM subcontract.subcontract_finished_goods_receipts AS receipt
LEFT JOIN subcontract.subcontract_orders AS subcontract_order
  ON subcontract_order.id = receipt.subcontract_order_id`

const selectSubcontractFinishedGoodsReceiptsByOrderSQL = selectSubcontractFinishedGoodsReceiptHeadersBaseSQL + `
WHERE lower(receipt.subcontract_order_ref) = lower($1)
   OR receipt.subcontract_order_id::text = $1
   OR lower(receipt.subcontract_order_no) = lower($1)
ORDER BY receipt.received_at DESC, receipt.receipt_no DESC`

const selectSubcontractFinishedGoodsReceiptLinesSQL = `
SELECT
  line.line_ref,
  line.line_no,
  line.item_ref,
  line.sku_code,
  line.item_name,
  COALESCE(line.batch_ref, ''),
  COALESCE(line.batch_no, ''),
  COALESCE(line.lot_no, ''),
  line.expiry_date,
  line.receive_qty::text,
  line.uom_code,
  line.base_receive_qty::text,
  line.base_uom_code,
  line.conversion_factor::text,
  COALESCE(line.packaging_status, ''),
  COALESCE(line.note, '')
FROM subcontract.subcontract_finished_goods_receipt_lines AS line
WHERE line.finished_goods_receipt_id = $1::uuid
ORDER BY line.line_no, line.line_ref`

const selectSubcontractFinishedGoodsReceiptEvidenceSQL = `
SELECT
  evidence.evidence_ref,
  evidence.evidence_type,
  COALESCE(evidence.file_name, ''),
  COALESCE(evidence.object_key, ''),
  COALESCE(evidence.external_url, ''),
  COALESCE(evidence.note, '')
FROM subcontract.subcontract_finished_goods_receipt_evidence AS evidence
WHERE evidence.finished_goods_receipt_id = $1::uuid
ORDER BY evidence.evidence_type, evidence.evidence_ref`

const upsertSubcontractFinishedGoodsReceiptSQL = `
INSERT INTO subcontract.subcontract_finished_goods_receipts (
  id,
  org_id,
  org_ref,
  receipt_ref,
  receipt_no,
  subcontract_order_id,
  subcontract_order_ref,
  subcontract_order_no,
  warehouse_ref,
  warehouse_code,
  location_ref,
  location_code,
  delivery_note_no,
  received_by_ref,
  received_at,
  note,
  created_at,
  created_by_ref,
  updated_at,
  updated_by_ref,
  version
) VALUES (
  COALESCE(CASE WHEN NULLIF($2::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $2::uuid END, gen_random_uuid()),
  $1::uuid,
  $3,
  $2,
  $4,
  (
    SELECT subcontract_order.id
    FROM subcontract.subcontract_orders AS subcontract_order
    WHERE subcontract_order.org_id = $1::uuid
      AND (
        subcontract_order.id::text = $5
        OR lower(subcontract_order.order_ref) = lower($5)
      )
    LIMIT 1
  ),
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
  $19
)
ON CONFLICT (org_id, receipt_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  receipt_no = EXCLUDED.receipt_no,
  subcontract_order_id = EXCLUDED.subcontract_order_id,
  subcontract_order_ref = EXCLUDED.subcontract_order_ref,
  subcontract_order_no = EXCLUDED.subcontract_order_no,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  location_ref = EXCLUDED.location_ref,
  location_code = EXCLUDED.location_code,
  delivery_note_no = EXCLUDED.delivery_note_no,
  received_by_ref = EXCLUDED.received_by_ref,
  received_at = EXCLUDED.received_at,
  note = EXCLUDED.note,
  created_at = EXCLUDED.created_at,
  created_by_ref = EXCLUDED.created_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = EXCLUDED.version
RETURNING id::text`

const deleteSubcontractFinishedGoodsReceiptLinesSQL = `
DELETE FROM subcontract.subcontract_finished_goods_receipt_lines
WHERE finished_goods_receipt_id = $1::uuid`

const insertSubcontractFinishedGoodsReceiptLineSQL = `
INSERT INTO subcontract.subcontract_finished_goods_receipt_lines (
  id,
  org_id,
  finished_goods_receipt_id,
  line_ref,
  line_no,
  item_ref,
  sku_code,
  item_name,
  batch_ref,
  batch_no,
  lot_no,
  expiry_date,
  receive_qty,
  uom_code,
  base_receive_qty,
  base_uom_code,
  conversion_factor,
  packaging_status,
  note
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
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
  $18
)`

const deleteSubcontractFinishedGoodsReceiptEvidenceSQL = `
DELETE FROM subcontract.subcontract_finished_goods_receipt_evidence
WHERE finished_goods_receipt_id = $1::uuid`

const insertSubcontractFinishedGoodsReceiptEvidenceSQL = `
INSERT INTO subcontract.subcontract_finished_goods_receipt_evidence (
  id,
  org_id,
  finished_goods_receipt_id,
  evidence_ref,
  evidence_type,
  file_name,
  object_key,
  external_url,
  note
) VALUES (
  COALESCE(CASE WHEN NULLIF($1::text, '') ~* '^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$' THEN $1::uuid END, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $1,
  $4,
  $5,
  $6,
  $7,
  $8
)`

func (s PostgresSubcontractFinishedGoodsReceiptStore) Save(
	ctx context.Context,
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	return s.withTx(ctx, func(tx *sql.Tx) error {
		orgID, err := s.resolveOrgID(ctx, tx, receipt.OrgID)
		if err != nil {
			return err
		}
		persistedID, err := upsertPostgresSubcontractFinishedGoodsReceipt(ctx, tx, orgID, receipt)
		if err != nil {
			return err
		}
		if err := replacePostgresSubcontractFinishedGoodsReceiptLines(ctx, tx, orgID, persistedID, receipt); err != nil {
			return err
		}
		if err := replacePostgresSubcontractFinishedGoodsReceiptEvidence(ctx, tx, orgID, persistedID, receipt); err != nil {
			return err
		}

		return nil
	})
}

func (s PostgresSubcontractFinishedGoodsReceiptStore) ListBySubcontractOrder(
	ctx context.Context,
	subcontractOrderID string,
) ([]productiondomain.SubcontractFinishedGoodsReceipt, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSubcontractFinishedGoodsReceiptsByOrderSQL, strings.TrimSpace(subcontractOrderID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	receipts := make([]productiondomain.SubcontractFinishedGoodsReceipt, 0)
	for rows.Next() {
		receipt, err := scanPostgresSubcontractFinishedGoodsReceipt(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return receipts, nil
}

func (s PostgresSubcontractFinishedGoodsReceiptStore) withTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin subcontract finished goods receipt transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit subcontract finished goods receipt transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSubcontractFinishedGoodsReceiptStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresSubcontractFinishedGoodsReceiptUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSubcontractFinishedGoodsReceiptOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresSubcontractFinishedGoodsReceiptUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve subcontract finished goods receipt org %q: %w", orgRef, err)
		}
	}
	if isPostgresSubcontractFinishedGoodsReceiptUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("subcontract finished goods receipt org %q cannot be resolved", orgRef)
}

func scanPostgresSubcontractFinishedGoodsReceipt(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFinishedGoodsReceipt, error) {
	header, err := scanPostgresSubcontractFinishedGoodsReceiptHeader(row)
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceipt{}, err
	}
	lines, err := listPostgresSubcontractFinishedGoodsReceiptLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceipt{}, err
	}
	evidence, err := listPostgresSubcontractFinishedGoodsReceiptEvidence(ctx, queryer, header.PersistedID)
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceipt{}, err
	}

	return buildPostgresSubcontractFinishedGoodsReceipt(header, lines, evidence)
}

func scanPostgresSubcontractFinishedGoodsReceiptHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSubcontractFinishedGoodsReceiptHeader, error) {
	var header postgresSubcontractFinishedGoodsReceiptHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.ReceiptNo,
		&header.SubcontractOrderID,
		&header.SubcontractOrderNo,
		&header.FactoryID,
		&header.FactoryCode,
		&header.FactoryName,
		&header.WarehouseID,
		&header.WarehouseCode,
		&header.LocationID,
		&header.LocationCode,
		&header.DeliveryNoteNo,
		&header.ReceivedBy,
		&header.ReceivedAt,
		&header.Note,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.Version,
	)
	if err != nil {
		return postgresSubcontractFinishedGoodsReceiptHeader{}, err
	}

	return header, nil
}

func listPostgresSubcontractFinishedGoodsReceiptLines(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	persistedID string,
) ([]productiondomain.SubcontractFinishedGoodsReceiptLine, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractFinishedGoodsReceiptLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]productiondomain.SubcontractFinishedGoodsReceiptLine, 0)
	for rows.Next() {
		line, err := scanPostgresSubcontractFinishedGoodsReceiptLine(rows)
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

func scanPostgresSubcontractFinishedGoodsReceiptLine(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFinishedGoodsReceiptLine, error) {
	var (
		id             string
		lineNo         int
		itemID         string
		skuCode        string
		itemName       string
		batchID        string
		batchNo        string
		lotNo          string
		expiryDate     time.Time
		receiveQtyText string
		uomCode        string
		baseQtyText    string
		baseUOMCode    string
		conversionText string
		packaging      string
		note           string
	)
	if err := row.Scan(
		&id,
		&lineNo,
		&itemID,
		&skuCode,
		&itemName,
		&batchID,
		&batchNo,
		&lotNo,
		&expiryDate,
		&receiveQtyText,
		&uomCode,
		&baseQtyText,
		&baseUOMCode,
		&conversionText,
		&packaging,
		&note,
	); err != nil {
		return productiondomain.SubcontractFinishedGoodsReceiptLine{}, err
	}
	receiveQty, err := decimal.ParseQuantity(receiveQtyText)
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceiptLine{}, err
	}
	baseReceiveQty, err := decimal.ParseQuantity(baseQtyText)
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceiptLine{}, err
	}
	conversion, err := decimal.ParseQuantity(conversionText)
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceiptLine{}, err
	}

	return productiondomain.NewSubcontractFinishedGoodsReceiptLine(productiondomain.NewSubcontractFinishedGoodsReceiptLineInput{
		ID:               id,
		LineNo:           lineNo,
		ItemID:           itemID,
		SKUCode:          skuCode,
		ItemName:         itemName,
		BatchID:          batchID,
		BatchNo:          batchNo,
		LotNo:            lotNo,
		ExpiryDate:       expiryDate,
		ReceiveQty:       receiveQty,
		UOMCode:          uomCode,
		BaseReceiveQty:   baseReceiveQty,
		BaseUOMCode:      baseUOMCode,
		ConversionFactor: conversion,
		PackagingStatus:  packaging,
		Note:             note,
	})
}

func listPostgresSubcontractFinishedGoodsReceiptEvidence(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	persistedID string,
) ([]productiondomain.SubcontractFinishedGoodsReceiptEvidence, error) {
	rows, err := queryer.QueryContext(ctx, selectSubcontractFinishedGoodsReceiptEvidenceSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evidence := make([]productiondomain.SubcontractFinishedGoodsReceiptEvidence, 0)
	for rows.Next() {
		item, err := scanPostgresSubcontractFinishedGoodsReceiptEvidence(rows)
		if err != nil {
			return nil, err
		}
		evidence = append(evidence, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return evidence, nil
}

func scanPostgresSubcontractFinishedGoodsReceiptEvidence(
	row interface{ Scan(dest ...any) error },
) (productiondomain.SubcontractFinishedGoodsReceiptEvidence, error) {
	var (
		id           string
		evidenceType string
		fileName     string
		objectKey    string
		externalURL  string
		note         string
	)
	if err := row.Scan(&id, &evidenceType, &fileName, &objectKey, &externalURL, &note); err != nil {
		return productiondomain.SubcontractFinishedGoodsReceiptEvidence{}, err
	}

	return productiondomain.NewSubcontractFinishedGoodsReceiptEvidence(productiondomain.NewSubcontractFinishedGoodsReceiptEvidenceInput{
		ID:           id,
		EvidenceType: evidenceType,
		FileName:     fileName,
		ObjectKey:    objectKey,
		ExternalURL:  externalURL,
		Note:         note,
	})
}

func buildPostgresSubcontractFinishedGoodsReceipt(
	header postgresSubcontractFinishedGoodsReceiptHeader,
	lines []productiondomain.SubcontractFinishedGoodsReceiptLine,
	evidence []productiondomain.SubcontractFinishedGoodsReceiptEvidence,
) (productiondomain.SubcontractFinishedGoodsReceipt, error) {
	lineInputs := make([]productiondomain.NewSubcontractFinishedGoodsReceiptLineInput, 0, len(lines))
	for _, line := range lines {
		lineInputs = append(lineInputs, productiondomain.NewSubcontractFinishedGoodsReceiptLineInput{
			ID:               line.ID,
			LineNo:           line.LineNo,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			ItemName:         line.ItemName,
			BatchID:          line.BatchID,
			BatchNo:          line.BatchNo,
			LotNo:            line.LotNo,
			ExpiryDate:       line.ExpiryDate,
			ReceiveQty:       line.ReceiveQty,
			UOMCode:          line.UOMCode.String(),
			BaseReceiveQty:   line.BaseReceiveQty,
			BaseUOMCode:      line.BaseUOMCode.String(),
			ConversionFactor: line.ConversionFactor,
			PackagingStatus:  line.PackagingStatus,
			Note:             line.Note,
		})
	}
	evidenceInputs := make([]productiondomain.NewSubcontractFinishedGoodsReceiptEvidenceInput, 0, len(evidence))
	for _, item := range evidence {
		evidenceInputs = append(evidenceInputs, productiondomain.NewSubcontractFinishedGoodsReceiptEvidenceInput{
			ID:           item.ID,
			EvidenceType: item.EvidenceType,
			FileName:     item.FileName,
			ObjectKey:    item.ObjectKey,
			ExternalURL:  item.ExternalURL,
			Note:         item.Note,
		})
	}

	receipt, err := productiondomain.NewSubcontractFinishedGoodsReceipt(productiondomain.NewSubcontractFinishedGoodsReceiptInput{
		ID:                 header.ID,
		OrgID:              header.OrgID,
		ReceiptNo:          header.ReceiptNo,
		SubcontractOrderID: header.SubcontractOrderID,
		SubcontractOrderNo: header.SubcontractOrderNo,
		FactoryID:          header.FactoryID,
		FactoryCode:        header.FactoryCode,
		FactoryName:        header.FactoryName,
		WarehouseID:        header.WarehouseID,
		WarehouseCode:      header.WarehouseCode,
		LocationID:         header.LocationID,
		LocationCode:       header.LocationCode,
		DeliveryNoteNo:     header.DeliveryNoteNo,
		Status:             productiondomain.SubcontractFinishedGoodsReceiptStatusQCHold,
		Lines:              lineInputs,
		Evidence:           evidenceInputs,
		ReceivedBy:         header.ReceivedBy,
		ReceivedAt:         header.ReceivedAt,
		Note:               header.Note,
		CreatedAt:          header.CreatedAt,
		CreatedBy:          header.CreatedBy,
		UpdatedAt:          header.UpdatedAt,
		UpdatedBy:          header.UpdatedBy,
	})
	if err != nil {
		return productiondomain.SubcontractFinishedGoodsReceipt{}, err
	}
	receipt.Version = header.Version
	if err := receipt.Validate(); err != nil {
		return productiondomain.SubcontractFinishedGoodsReceipt{}, err
	}

	return receipt, nil
}

func upsertPostgresSubcontractFinishedGoodsReceipt(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	orgID string,
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSubcontractFinishedGoodsReceiptSQL,
		orgID,
		receipt.ID,
		receipt.OrgID,
		receipt.ReceiptNo,
		receipt.SubcontractOrderID,
		receipt.SubcontractOrderNo,
		receipt.WarehouseID,
		nullablePostgresSubcontractFinishedGoodsReceiptText(receipt.WarehouseCode),
		receipt.LocationID,
		nullablePostgresSubcontractFinishedGoodsReceiptText(receipt.LocationCode),
		receipt.DeliveryNoteNo,
		receipt.ReceivedBy,
		receipt.ReceivedAt.UTC(),
		nullablePostgresSubcontractFinishedGoodsReceiptText(receipt.Note),
		receipt.CreatedAt.UTC(),
		receipt.CreatedBy,
		receipt.UpdatedAt.UTC(),
		receipt.UpdatedBy,
		receipt.Version,
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert subcontract finished goods receipt %q: %w", receipt.ID, err)
	}

	return persistedID, nil
}

func replacePostgresSubcontractFinishedGoodsReceiptLines(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	orgID string,
	persistedID string,
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractFinishedGoodsReceiptLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract finished goods receipt lines: %w", err)
	}
	for _, line := range receipt.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractFinishedGoodsReceiptLineSQL,
			line.ID,
			orgID,
			persistedID,
			line.LineNo,
			line.ItemID,
			line.SKUCode,
			line.ItemName,
			line.BatchID,
			line.BatchNo,
			line.LotNo,
			line.ExpiryDate.UTC(),
			line.ReceiveQty.String(),
			line.UOMCode.String(),
			line.BaseReceiveQty.String(),
			line.BaseUOMCode.String(),
			line.ConversionFactor.String(),
			nullablePostgresSubcontractFinishedGoodsReceiptText(line.PackagingStatus),
			nullablePostgresSubcontractFinishedGoodsReceiptText(line.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract finished goods receipt line %q: %w", line.ID, err)
		}
	}

	return nil
}

func replacePostgresSubcontractFinishedGoodsReceiptEvidence(
	ctx context.Context,
	queryer postgresSubcontractFinishedGoodsReceiptQueryer,
	orgID string,
	persistedID string,
	receipt productiondomain.SubcontractFinishedGoodsReceipt,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSubcontractFinishedGoodsReceiptEvidenceSQL, persistedID); err != nil {
		return fmt.Errorf("delete subcontract finished goods receipt evidence: %w", err)
	}
	for _, evidence := range receipt.Evidence {
		if _, err := queryer.ExecContext(
			ctx,
			insertSubcontractFinishedGoodsReceiptEvidenceSQL,
			evidence.ID,
			orgID,
			persistedID,
			evidence.EvidenceType,
			nullablePostgresSubcontractFinishedGoodsReceiptText(evidence.FileName),
			nullablePostgresSubcontractFinishedGoodsReceiptText(evidence.ObjectKey),
			nullablePostgresSubcontractFinishedGoodsReceiptText(evidence.ExternalURL),
			nullablePostgresSubcontractFinishedGoodsReceiptText(evidence.Note),
		); err != nil {
			return fmt.Errorf("insert subcontract finished goods receipt evidence %q: %w", evidence.ID, err)
		}
	}

	return nil
}

func nullablePostgresSubcontractFinishedGoodsReceiptText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func isPostgresSubcontractFinishedGoodsReceiptUUIDText(value string) bool {
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
			if !isPostgresSubcontractFinishedGoodsReceiptHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresSubcontractFinishedGoodsReceiptHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ SubcontractFinishedGoodsReceiptStore = PostgresSubcontractFinishedGoodsReceiptStore{}
