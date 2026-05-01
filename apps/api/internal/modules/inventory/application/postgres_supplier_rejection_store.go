package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresSupplierRejectionStoreConfig struct {
	DefaultOrgID string
}

type PostgresSupplierRejectionStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresSupplierRejectionQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresSupplierRejectionHeader struct {
	PersistedID           string
	ID                    string
	OrgID                 string
	RejectionNo           string
	SupplierID            string
	SupplierCode          string
	SupplierName          string
	PurchaseOrderID       string
	PurchaseOrderNo       string
	GoodsReceiptID        string
	GoodsReceiptNo        string
	InboundQCInspectionID string
	WarehouseID           string
	WarehouseCode         string
	Status                string
	Reason                string
	CreatedAt             time.Time
	CreatedBy             string
	UpdatedAt             time.Time
	UpdatedBy             string
	SubmittedAt           sql.NullTime
	SubmittedBy           string
	ConfirmedAt           sql.NullTime
	ConfirmedBy           string
	CancelledAt           sql.NullTime
	CancelledBy           string
	CancelReason          string
}

func NewPostgresSupplierRejectionStore(
	db *sql.DB,
	cfg PostgresSupplierRejectionStoreConfig,
) PostgresSupplierRejectionStore {
	return PostgresSupplierRejectionStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectSupplierRejectionOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectSupplierRejectionHeadersBaseSQL = `
SELECT
  id::text,
  COALESCE(rejection_ref, id::text),
  COALESCE(org_ref, org_id::text),
  rejection_no,
  COALESCE(supplier_ref, supplier_id::text, ''),
  COALESCE(supplier_code, ''),
  supplier_name,
  COALESCE(purchase_order_ref, purchase_order_id::text, ''),
  COALESCE(purchase_order_no, ''),
  COALESCE(goods_receipt_ref, goods_receipt_id::text, ''),
  COALESCE(goods_receipt_no, ''),
  COALESCE(inbound_qc_inspection_ref, inbound_qc_inspection_id::text, ''),
  COALESCE(warehouse_ref, warehouse_id::text, ''),
  COALESCE(warehouse_code, ''),
  status,
  reason,
  created_at,
  COALESCE(created_by_ref, created_by::text, ''),
  updated_at,
  COALESCE(updated_by_ref, updated_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  confirmed_at,
  COALESCE(confirmed_by_ref, confirmed_by::text, ''),
  cancelled_at,
  COALESCE(cancelled_by_ref, cancelled_by::text, ''),
  COALESCE(cancel_reason, '')
FROM inventory.supplier_rejections`

const selectSupplierRejectionHeadersSQL = selectSupplierRejectionHeadersBaseSQL + `
ORDER BY created_at DESC, rejection_no DESC`

const findSupplierRejectionHeaderSQL = selectSupplierRejectionHeadersBaseSQL + `
WHERE rejection_ref = $1 OR id::text = $1
LIMIT 1`

const selectSupplierRejectionLinesSQL = `
SELECT
  line_ref,
  COALESCE(purchase_order_line_ref, purchase_order_line_id::text, ''),
  COALESCE(goods_receipt_line_ref, goods_receipt_line_id::text, ''),
  COALESCE(inbound_qc_inspection_ref, inbound_qc_inspection_id::text, ''),
  COALESCE(item_ref, item_id::text, ''),
  sku_code,
  COALESCE(item_name, ''),
  COALESCE(batch_ref, batch_id::text, ''),
  batch_no,
  lot_no,
  expiry_date::text,
  rejected_qty::text,
  uom_code,
  base_uom_code,
  reason
FROM inventory.supplier_rejection_lines
WHERE rejection_id = $1::uuid
ORDER BY line_no, created_at, line_ref`

const selectSupplierRejectionAttachmentsSQL = `
SELECT
  attachment_ref,
  COALESCE(line_ref, ''),
  file_name,
  object_key,
  COALESCE(content_type, ''),
  uploaded_at,
  uploaded_by_ref,
  COALESCE(source, '')
FROM inventory.supplier_rejection_attachments
WHERE rejection_id = $1::uuid
ORDER BY uploaded_at, attachment_ref`

const findDuplicateSupplierRejectionSQL = `
SELECT rejection_ref
FROM inventory.supplier_rejections
WHERE org_id = $1::uuid
  AND rejection_no = $2
  AND rejection_ref <> $3
LIMIT 1`

const upsertSupplierRejectionSQL = `
INSERT INTO inventory.supplier_rejections (
  id,
  org_id,
  rejection_ref,
  org_ref,
  rejection_no,
  supplier_id,
  supplier_ref,
  supplier_code,
  supplier_name,
  purchase_order_id,
  purchase_order_ref,
  purchase_order_no,
  goods_receipt_id,
  goods_receipt_ref,
  goods_receipt_no,
  inbound_qc_inspection_id,
  inbound_qc_inspection_ref,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  status,
  reason,
  created_by,
  created_by_ref,
  updated_by,
  updated_by_ref,
  submitted_at,
  submitted_by,
  submitted_by_ref,
  confirmed_at,
  confirmed_by,
  confirmed_by_ref,
  cancelled_at,
  cancelled_by,
  cancelled_by_ref,
  cancel_reason,
  created_at,
  updated_at
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
  $13::uuid,
  $14,
  $15,
  $16::uuid,
  $17,
  $18::uuid,
  $19,
  $20,
  $21,
  $22,
  $23::uuid,
  $24,
  $25::uuid,
  $26,
  $27,
  $28::uuid,
  $29,
  $30,
  $31::uuid,
  $32,
  $33,
  $34::uuid,
  $35,
  $36,
  $37,
  $38
)
ON CONFLICT ON CONSTRAINT uq_supplier_rejections_org_ref
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  rejection_no = EXCLUDED.rejection_no,
  supplier_id = EXCLUDED.supplier_id,
  supplier_ref = EXCLUDED.supplier_ref,
  supplier_code = EXCLUDED.supplier_code,
  supplier_name = EXCLUDED.supplier_name,
  purchase_order_id = EXCLUDED.purchase_order_id,
  purchase_order_ref = EXCLUDED.purchase_order_ref,
  purchase_order_no = EXCLUDED.purchase_order_no,
  goods_receipt_id = EXCLUDED.goods_receipt_id,
  goods_receipt_ref = EXCLUDED.goods_receipt_ref,
  goods_receipt_no = EXCLUDED.goods_receipt_no,
  inbound_qc_inspection_id = EXCLUDED.inbound_qc_inspection_id,
  inbound_qc_inspection_ref = EXCLUDED.inbound_qc_inspection_ref,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  status = EXCLUDED.status,
  reason = EXCLUDED.reason,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  submitted_at = EXCLUDED.submitted_at,
  submitted_by = EXCLUDED.submitted_by,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  confirmed_at = EXCLUDED.confirmed_at,
  confirmed_by = EXCLUDED.confirmed_by,
  confirmed_by_ref = EXCLUDED.confirmed_by_ref,
  cancelled_at = EXCLUDED.cancelled_at,
  cancelled_by = EXCLUDED.cancelled_by,
  cancelled_by_ref = EXCLUDED.cancelled_by_ref,
  cancel_reason = EXCLUDED.cancel_reason,
  updated_at = EXCLUDED.updated_at
RETURNING id::text`

const deleteSupplierRejectionLinesSQL = `
DELETE FROM inventory.supplier_rejection_lines
WHERE rejection_id = $1::uuid`

const insertSupplierRejectionLineSQL = `
INSERT INTO inventory.supplier_rejection_lines (
  id,
  org_id,
  rejection_id,
  line_ref,
  line_no,
  purchase_order_line_id,
  purchase_order_line_ref,
  goods_receipt_line_id,
  goods_receipt_line_ref,
  inbound_qc_inspection_id,
  inbound_qc_inspection_ref,
  item_id,
  item_ref,
  sku_code,
  item_name,
  batch_id,
  batch_ref,
  batch_no,
  lot_no,
  expiry_date,
  rejected_qty,
  uom_code,
  base_uom_code,
  reason,
  created_at,
  updated_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6::uuid,
  $7,
  $8::uuid,
  $9,
  $10::uuid,
  $11,
  $12::uuid,
  $13,
  $14,
  $15,
  $16::uuid,
  $17,
  $18,
  $19,
  $20::date,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26
)`

const deleteSupplierRejectionAttachmentsSQL = `
DELETE FROM inventory.supplier_rejection_attachments
WHERE rejection_id = $1::uuid`

const insertSupplierRejectionAttachmentSQL = `
INSERT INTO inventory.supplier_rejection_attachments (
  id,
  org_id,
  rejection_id,
  attachment_ref,
  line_ref,
  file_name,
  object_key,
  content_type,
  uploaded_at,
  uploaded_by_ref,
  source,
  created_at
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

func (s PostgresSupplierRejectionStore) List(
	ctx context.Context,
	filter domain.SupplierRejectionFilter,
) ([]domain.SupplierRejection, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectSupplierRejectionHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rejections := make([]domain.SupplierRejection, 0)
	for rows.Next() {
		rejection, err := scanPostgresSupplierRejection(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if filter.Matches(rejection) {
			rejections = append(rejections, rejection)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortSupplierRejections(rejections)

	return rejections, nil
}

func (s PostgresSupplierRejectionStore) Get(
	ctx context.Context,
	id string,
) (domain.SupplierRejection, error) {
	if s.db == nil {
		return domain.SupplierRejection{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findSupplierRejectionHeaderSQL, strings.TrimSpace(id))
	rejection, err := scanPostgresSupplierRejection(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.SupplierRejection{}, ErrSupplierRejectionNotFound
	}
	if err != nil {
		return domain.SupplierRejection{}, err
	}

	return rejection, nil
}

func (s PostgresSupplierRejectionStore) Save(
	ctx context.Context,
	rejection domain.SupplierRejection,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := rejection.Validate(); err != nil {
		return err
	}
	orgID, err := s.resolveOrgID(ctx, s.db, rejection.OrgID)
	if err != nil {
		return err
	}
	if err := s.ensureNoDuplicate(ctx, orgID, rejection); err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin supplier rejection transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, err := upsertSupplierRejection(ctx, tx, orgID, rejection)
	if err != nil {
		return err
	}
	if err := replaceSupplierRejectionLines(ctx, tx, orgID, persistedID, rejection); err != nil {
		return err
	}
	if err := replaceSupplierRejectionAttachments(ctx, tx, orgID, persistedID, rejection); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit supplier rejection transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresSupplierRejectionStore) resolveOrgID(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectSupplierRejectionOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve supplier rejection org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("supplier rejection org %q cannot be resolved", orgRef)
}

func (s PostgresSupplierRejectionStore) ensureNoDuplicate(
	ctx context.Context,
	orgID string,
	rejection domain.SupplierRejection,
) error {
	var existingID string
	err := s.db.QueryRowContext(
		ctx,
		findDuplicateSupplierRejectionSQL,
		orgID,
		rejection.RejectionNo,
		rejection.ID,
	).Scan(&existingID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find duplicate supplier rejection: %w", err)
	}

	return ErrSupplierRejectionDuplicate
}

func scanPostgresSupplierRejection(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	row interface{ Scan(dest ...any) error },
) (domain.SupplierRejection, error) {
	header, err := scanPostgresSupplierRejectionHeader(row)
	if err != nil {
		return domain.SupplierRejection{}, err
	}
	lines, err := listPostgresSupplierRejectionLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return domain.SupplierRejection{}, err
	}
	attachments, err := listPostgresSupplierRejectionAttachments(ctx, queryer, header.PersistedID)
	if err != nil {
		return domain.SupplierRejection{}, err
	}

	return buildPostgresSupplierRejection(header, lines, attachments)
}

func scanPostgresSupplierRejectionHeader(
	row interface{ Scan(dest ...any) error },
) (postgresSupplierRejectionHeader, error) {
	var header postgresSupplierRejectionHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.RejectionNo,
		&header.SupplierID,
		&header.SupplierCode,
		&header.SupplierName,
		&header.PurchaseOrderID,
		&header.PurchaseOrderNo,
		&header.GoodsReceiptID,
		&header.GoodsReceiptNo,
		&header.InboundQCInspectionID,
		&header.WarehouseID,
		&header.WarehouseCode,
		&header.Status,
		&header.Reason,
		&header.CreatedAt,
		&header.CreatedBy,
		&header.UpdatedAt,
		&header.UpdatedBy,
		&header.SubmittedAt,
		&header.SubmittedBy,
		&header.ConfirmedAt,
		&header.ConfirmedBy,
		&header.CancelledAt,
		&header.CancelledBy,
		&header.CancelReason,
	)
	if err != nil {
		return postgresSupplierRejectionHeader{}, err
	}

	return header, nil
}

func listPostgresSupplierRejectionLines(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	persistedID string,
) ([]domain.NewSupplierRejectionLineInput, error) {
	rows, err := queryer.QueryContext(ctx, selectSupplierRejectionLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.NewSupplierRejectionLineInput, 0)
	for rows.Next() {
		line, err := scanPostgresSupplierRejectionLine(rows)
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

func scanPostgresSupplierRejectionLine(
	row interface{ Scan(dest ...any) error },
) (domain.NewSupplierRejectionLineInput, error) {
	var (
		id                    string
		purchaseOrderLineID   string
		goodsReceiptLineID    string
		inboundQCInspectionID string
		itemID                string
		sku                   string
		itemName              string
		batchID               string
		batchNo               string
		lotNo                 string
		expiryDateText        string
		rejectedQtyText       string
		uomCode               string
		baseUOMCode           string
		reason                string
	)
	if err := row.Scan(
		&id,
		&purchaseOrderLineID,
		&goodsReceiptLineID,
		&inboundQCInspectionID,
		&itemID,
		&sku,
		&itemName,
		&batchID,
		&batchNo,
		&lotNo,
		&expiryDateText,
		&rejectedQtyText,
		&uomCode,
		&baseUOMCode,
		&reason,
	); err != nil {
		return domain.NewSupplierRejectionLineInput{}, err
	}
	rejectedQty, err := decimal.ParseQuantity(rejectedQtyText)
	if err != nil {
		return domain.NewSupplierRejectionLineInput{}, err
	}
	expiryDate, err := time.Parse("2006-01-02", expiryDateText)
	if err != nil {
		return domain.NewSupplierRejectionLineInput{}, err
	}

	return domain.NewSupplierRejectionLineInput{
		ID:                    id,
		PurchaseOrderLineID:   purchaseOrderLineID,
		GoodsReceiptLineID:    goodsReceiptLineID,
		InboundQCInspectionID: inboundQCInspectionID,
		ItemID:                itemID,
		SKU:                   sku,
		ItemName:              itemName,
		BatchID:               batchID,
		BatchNo:               batchNo,
		LotNo:                 lotNo,
		ExpiryDate:            expiryDate,
		RejectedQuantity:      rejectedQty,
		UOMCode:               uomCode,
		BaseUOMCode:           baseUOMCode,
		Reason:                reason,
	}, nil
}

func listPostgresSupplierRejectionAttachments(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	persistedID string,
) ([]domain.NewSupplierRejectionAttachmentInput, error) {
	rows, err := queryer.QueryContext(ctx, selectSupplierRejectionAttachmentsSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := make([]domain.NewSupplierRejectionAttachmentInput, 0)
	for rows.Next() {
		attachment, err := scanPostgresSupplierRejectionAttachment(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

func scanPostgresSupplierRejectionAttachment(
	row interface{ Scan(dest ...any) error },
) (domain.NewSupplierRejectionAttachmentInput, error) {
	var (
		id          string
		lineID      string
		fileName    string
		objectKey   string
		contentType string
		uploadedAt  time.Time
		uploadedBy  string
		source      string
	)
	if err := row.Scan(&id, &lineID, &fileName, &objectKey, &contentType, &uploadedAt, &uploadedBy, &source); err != nil {
		return domain.NewSupplierRejectionAttachmentInput{}, err
	}

	return domain.NewSupplierRejectionAttachmentInput{
		ID:          id,
		LineID:      lineID,
		FileName:    fileName,
		ObjectKey:   objectKey,
		ContentType: contentType,
		UploadedAt:  uploadedAt,
		UploadedBy:  uploadedBy,
		Source:      source,
	}, nil
}

func buildPostgresSupplierRejection(
	header postgresSupplierRejectionHeader,
	lines []domain.NewSupplierRejectionLineInput,
	attachments []domain.NewSupplierRejectionAttachmentInput,
) (domain.SupplierRejection, error) {
	rejection, err := domain.NewSupplierRejection(domain.NewSupplierRejectionInput{
		ID:                    header.ID,
		OrgID:                 header.OrgID,
		RejectionNo:           header.RejectionNo,
		SupplierID:            header.SupplierID,
		SupplierCode:          header.SupplierCode,
		SupplierName:          header.SupplierName,
		PurchaseOrderID:       header.PurchaseOrderID,
		PurchaseOrderNo:       header.PurchaseOrderNo,
		GoodsReceiptID:        header.GoodsReceiptID,
		GoodsReceiptNo:        header.GoodsReceiptNo,
		InboundQCInspectionID: header.InboundQCInspectionID,
		WarehouseID:           header.WarehouseID,
		WarehouseCode:         header.WarehouseCode,
		Reason:                header.Reason,
		Lines:                 lines,
		Attachments:           attachments,
		CreatedAt:             header.CreatedAt,
		CreatedBy:             header.CreatedBy,
		UpdatedAt:             header.UpdatedAt,
		UpdatedBy:             firstNonBlankSupplierRejection(header.UpdatedBy, header.CreatedBy),
	})
	if err != nil {
		return domain.SupplierRejection{}, err
	}
	rejection.Status = domain.SupplierRejectionStatus(header.Status)
	rejection.SubmittedAt = nullableSupplierRejectionTimeValue(header.SubmittedAt)
	rejection.SubmittedBy = strings.TrimSpace(header.SubmittedBy)
	rejection.ConfirmedAt = nullableSupplierRejectionTimeValue(header.ConfirmedAt)
	rejection.ConfirmedBy = strings.TrimSpace(header.ConfirmedBy)
	rejection.CancelledAt = nullableSupplierRejectionTimeValue(header.CancelledAt)
	rejection.CancelledBy = strings.TrimSpace(header.CancelledBy)
	rejection.CancelReason = strings.TrimSpace(header.CancelReason)
	if err := rejection.Validate(); err != nil {
		return domain.SupplierRejection{}, err
	}

	return rejection, nil
}

func upsertSupplierRejection(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	orgID string,
	rejection domain.SupplierRejection,
) (string, error) {
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertSupplierRejectionSQL,
		nullableUUID(rejection.ID),
		orgID,
		nullableText(rejection.ID),
		nullableText(rejection.OrgID),
		rejection.RejectionNo,
		nullableUUID(rejection.SupplierID),
		nullableText(rejection.SupplierID),
		nullableText(rejection.SupplierCode),
		rejection.SupplierName,
		nullableUUID(rejection.PurchaseOrderID),
		nullableText(rejection.PurchaseOrderID),
		nullableText(rejection.PurchaseOrderNo),
		nullableUUID(rejection.GoodsReceiptID),
		nullableText(rejection.GoodsReceiptID),
		nullableText(rejection.GoodsReceiptNo),
		nullableUUID(rejection.InboundQCInspectionID),
		nullableText(rejection.InboundQCInspectionID),
		nullableUUID(rejection.WarehouseID),
		nullableText(rejection.WarehouseID),
		nullableText(rejection.WarehouseCode),
		string(rejection.Status),
		rejection.Reason,
		nullableUUID(rejection.CreatedBy),
		nullableText(rejection.CreatedBy),
		nullableUUID(rejection.UpdatedBy),
		nullableText(rejection.UpdatedBy),
		nullableTime(rejection.SubmittedAt),
		nullableUUID(rejection.SubmittedBy),
		nullableText(rejection.SubmittedBy),
		nullableTime(rejection.ConfirmedAt),
		nullableUUID(rejection.ConfirmedBy),
		nullableText(rejection.ConfirmedBy),
		nullableTime(rejection.CancelledAt),
		nullableUUID(rejection.CancelledBy),
		nullableText(rejection.CancelledBy),
		nullableText(rejection.CancelReason),
		rejection.CreatedAt.UTC(),
		rejection.UpdatedAt.UTC(),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert supplier rejection: %w", err)
	}

	return persistedID, nil
}

func replaceSupplierRejectionLines(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	orgID string,
	persistedID string,
	rejection domain.SupplierRejection,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSupplierRejectionLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete supplier rejection lines: %w", err)
	}
	for index, line := range rejection.Lines {
		if _, err := queryer.ExecContext(
			ctx,
			insertSupplierRejectionLineSQL,
			nullableUUID(line.ID),
			orgID,
			persistedID,
			line.ID,
			index+1,
			nullableUUID(line.PurchaseOrderLineID),
			nullableText(line.PurchaseOrderLineID),
			nullableUUID(line.GoodsReceiptLineID),
			line.GoodsReceiptLineID,
			nullableUUID(line.InboundQCInspectionID),
			line.InboundQCInspectionID,
			nullableUUID(line.ItemID),
			line.ItemID,
			line.SKU,
			nullableText(line.ItemName),
			nullableUUID(line.BatchID),
			line.BatchID,
			line.BatchNo,
			line.LotNo,
			nullableDate(line.ExpiryDate),
			line.RejectedQuantity.String(),
			line.UOMCode.String(),
			line.BaseUOMCode.String(),
			line.Reason,
			rejection.CreatedAt.UTC(),
			rejection.UpdatedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert supplier rejection line: %w", err)
		}
	}

	return nil
}

func replaceSupplierRejectionAttachments(
	ctx context.Context,
	queryer postgresSupplierRejectionQueryer,
	orgID string,
	persistedID string,
	rejection domain.SupplierRejection,
) error {
	if _, err := queryer.ExecContext(ctx, deleteSupplierRejectionAttachmentsSQL, persistedID); err != nil {
		return fmt.Errorf("delete supplier rejection attachments: %w", err)
	}
	for _, attachment := range rejection.Attachments {
		if _, err := queryer.ExecContext(
			ctx,
			insertSupplierRejectionAttachmentSQL,
			nullableUUID(attachment.ID),
			orgID,
			persistedID,
			attachment.ID,
			nullableText(attachment.LineID),
			attachment.FileName,
			attachment.ObjectKey,
			nullableText(attachment.ContentType),
			attachment.UploadedAt.UTC(),
			attachment.UploadedBy,
			nullableText(attachment.Source),
			attachment.UploadedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert supplier rejection attachment: %w", err)
		}
	}

	return nil
}

func nullableSupplierRejectionTimeValue(value sql.NullTime) time.Time {
	if !value.Valid {
		return time.Time{}
	}

	return value.Time.UTC()
}

func firstNonBlankSupplierRejection(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

var _ SupplierRejectionStore = PostgresSupplierRejectionStore{}
