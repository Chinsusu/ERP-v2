package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/domain"
)

type PostgresReturnReceiptStoreConfig struct {
	DefaultOrgID string
}

type PostgresReturnReceiptStore struct {
	db           *sql.DB
	defaultOrgID string
	expected     []domain.ExpectedReturn
}

type postgresReturnReceiptQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type postgresReturnReceiptHeader struct {
	PersistedID       string
	ID                string
	OrgID             string
	ReceiptNo         string
	WarehouseID       string
	WarehouseCode     string
	Source            string
	ReceivedBy        string
	ReceivedAt        time.Time
	PackageCondition  string
	Status            string
	Disposition       string
	TargetLocation    string
	OriginalOrderNo   string
	TrackingNo        string
	ReturnCode        string
	ScanCode          string
	CustomerName      string
	UnknownCase       bool
	StockMovementRef  string
	StockMovementType string
	TargetStockStatus string
	InvestigationNote string
	CreatedAt         time.Time
}

func NewPostgresReturnReceiptStore(
	db *sql.DB,
	cfg PostgresReturnReceiptStoreConfig,
) PostgresReturnReceiptStore {
	return PostgresReturnReceiptStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
		expected:     prototypeExpectedReturns(),
	}
}

const selectReturnReceiptOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectReturnReceiptHeadersBaseSQL = `
SELECT
  id::text,
  COALESCE(return_ref, id::text),
  COALESCE(org_ref, org_id::text),
  return_no,
  COALESCE(warehouse_ref, warehouse_id::text, ''),
  COALESCE(warehouse_code, ''),
  source,
  COALESCE(received_by_ref, received_by::text, ''),
  received_at,
  COALESCE(package_condition, ''),
  status,
  COALESCE(disposition, initial_disposition, ''),
  COALESCE(target_location, ''),
  COALESCE(original_order_no, ''),
  COALESCE(tracking_no, ''),
  COALESCE(return_code, ''),
  COALESCE(scan_code, return_code, tracking_no, return_no),
  COALESCE(customer_name, ''),
  unknown_case,
  COALESCE(stock_movement_ref, ''),
  COALESCE(stock_movement_type, ''),
  COALESCE(target_stock_status, ''),
  COALESCE(investigation_note, ''),
  created_at
FROM returns.return_orders`

const selectReturnReceiptHeadersSQL = selectReturnReceiptHeadersBaseSQL + `
ORDER BY created_at DESC, return_no DESC`

const findReturnReceiptHeaderSQL = selectReturnReceiptHeadersBaseSQL + `
WHERE return_ref = $1 OR id::text = $1
LIMIT 1`

const selectReturnReceiptLinesSQL = `
SELECT
  COALESCE(return_line.line_ref, return_line.id::text),
  COALESCE(return_line.sku_code, item.sku, ''),
  COALESCE(return_line.product_name, item.name, ''),
  COALESCE(return_line.quantity, return_line.returned_qty)::text,
  COALESCE(return_line.condition_text, return_line.condition_note, return_line.condition_code, '')
FROM returns.return_order_lines AS return_line
LEFT JOIN mdm.items AS item ON item.id = return_line.item_id
WHERE return_line.return_order_id = $1::uuid
ORDER BY return_line.line_no, return_line.created_at, COALESCE(return_line.line_ref, return_line.id::text)`

const findDuplicateReturnReceiptSQL = `
SELECT COALESCE(return_ref, id::text)
FROM returns.return_orders
WHERE org_id = $1::uuid
  AND (
    return_ref = $2
    OR return_no = $3
    OR (NULLIF($4, '') IS NOT NULL AND upper(COALESCE(scan_code, '')) = upper($4))
    OR (NULLIF($5, '') IS NOT NULL AND upper(COALESCE(tracking_no, '')) = upper($5))
    OR (NULLIF($6, '') IS NOT NULL AND upper(COALESCE(return_code, '')) = upper($6))
    OR (NULLIF($7, '') IS NOT NULL AND upper(COALESCE(original_order_no, '')) = upper($7))
  )
LIMIT 1`

const upsertReturnReceiptSQL = `
INSERT INTO returns.return_orders (
  id,
  org_id,
  return_ref,
  org_ref,
  return_no,
  sales_order_id,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  status,
  source,
  tracking_no,
  return_code,
  package_condition,
  initial_disposition,
  unknown_case,
  investigation_note,
  received_at,
  received_by,
  received_by_ref,
  created_at,
  created_by,
  updated_at,
  updated_by,
  disposition,
  target_location,
  original_order_ref,
  original_order_no,
  customer_name,
  scan_code,
  stock_movement_ref,
  stock_movement_type,
  target_stock_status
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4,
  $5,
  $6::uuid,
  $7::uuid,
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
  $19::uuid,
  $20,
  $21,
  $22::uuid,
  $23,
  $24::uuid,
  $25,
  $26,
  $27,
  $28,
  $29,
  $30,
  $31,
  $32,
  $33
)
ON CONFLICT (org_id, return_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  return_no = EXCLUDED.return_no,
  sales_order_id = EXCLUDED.sales_order_id,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  status = EXCLUDED.status,
  source = EXCLUDED.source,
  tracking_no = EXCLUDED.tracking_no,
  return_code = EXCLUDED.return_code,
  package_condition = EXCLUDED.package_condition,
  initial_disposition = EXCLUDED.initial_disposition,
  unknown_case = EXCLUDED.unknown_case,
  investigation_note = EXCLUDED.investigation_note,
  received_at = EXCLUDED.received_at,
  received_by = EXCLUDED.received_by,
  received_by_ref = EXCLUDED.received_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  disposition = EXCLUDED.disposition,
  target_location = EXCLUDED.target_location,
  original_order_ref = EXCLUDED.original_order_ref,
  original_order_no = EXCLUDED.original_order_no,
  customer_name = EXCLUDED.customer_name,
  scan_code = EXCLUDED.scan_code,
  stock_movement_ref = EXCLUDED.stock_movement_ref,
  stock_movement_type = EXCLUDED.stock_movement_type,
  target_stock_status = EXCLUDED.target_stock_status
RETURNING id::text`

const deleteReturnReceiptLinesSQL = `
DELETE FROM returns.return_order_lines
WHERE return_order_id = $1::uuid`

const insertReturnReceiptLineSQL = `
INSERT INTO returns.return_order_lines (
  id,
  org_id,
  return_order_id,
  line_ref,
  line_no,
  item_id,
  item_ref,
  sku_code,
  product_name,
  returned_qty,
  quantity,
  condition_note,
  condition_code,
  condition_text,
  batch_id,
  batch_ref,
  unit_id,
  unit_ref,
  uom_code,
  stock_movement_ref,
  created_at,
  created_by
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
  $15::uuid,
  $16,
  $17::uuid,
  $18,
  $19,
  $20,
  $21,
  $22::uuid
)`

const findReturnReceiptPersistedIDSQL = `
SELECT id::text, org_id::text, COALESCE(return_ref, id::text)
FROM returns.return_orders
WHERE return_ref = $1 OR id::text = $1
LIMIT 1`

const upsertReturnInspectionSQL = `
INSERT INTO returns.return_inspections (
  id,
  org_id,
  inspection_ref,
  return_order_id,
  return_ref,
  condition_code,
  disposition,
  status,
  target_location,
  risk_level,
  evidence_label,
  note,
  inspector_ref,
  inspected_at,
  created_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4::uuid,
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
  $15
)
ON CONFLICT (org_id, inspection_ref)
DO UPDATE SET
  return_order_id = EXCLUDED.return_order_id,
  return_ref = EXCLUDED.return_ref,
  condition_code = EXCLUDED.condition_code,
  disposition = EXCLUDED.disposition,
  status = EXCLUDED.status,
  target_location = EXCLUDED.target_location,
  risk_level = EXCLUDED.risk_level,
  evidence_label = EXCLUDED.evidence_label,
  note = EXCLUDED.note,
  inspector_ref = EXCLUDED.inspector_ref,
  inspected_at = EXCLUDED.inspected_at`

const findReturnInspectionSQL = `
SELECT
  COALESCE(inspection.inspection_ref, inspection.id::text),
  COALESCE(return_order.return_ref, return_order.id::text),
  return_order.return_no,
  inspection.condition_code,
  inspection.disposition,
  inspection.status,
  inspection.target_location,
  inspection.risk_level,
  COALESCE(inspection.inspector_ref, ''),
  COALESCE(inspection.note, ''),
  COALESCE(inspection.evidence_label, ''),
  inspection.inspected_at
FROM returns.return_inspections AS inspection
JOIN returns.return_orders AS return_order ON return_order.id = inspection.return_order_id
WHERE inspection.inspection_ref = $1 OR inspection.id::text = $1
LIMIT 1`

const findReturnInspectionForReceiptSQL = `
SELECT COALESCE(inspection_ref, id::text)
FROM returns.return_inspections
WHERE return_order_id = $1::uuid
  AND (inspection_ref = $2 OR id::text = $2)
LIMIT 1`

const upsertReturnDispositionActionSQL = `
INSERT INTO returns.return_disposition_actions (
  id,
  org_id,
  action_ref,
  return_order_id,
  return_ref,
  disposition,
  target_location,
  target_stock_status,
  action_code,
  note,
  actor_ref,
  decided_at,
  stock_movement_ref,
  created_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4::uuid,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14
)
ON CONFLICT (org_id, action_ref)
DO UPDATE SET
  return_order_id = EXCLUDED.return_order_id,
  return_ref = EXCLUDED.return_ref,
  disposition = EXCLUDED.disposition,
  target_location = EXCLUDED.target_location,
  target_stock_status = EXCLUDED.target_stock_status,
  action_code = EXCLUDED.action_code,
  note = EXCLUDED.note,
  actor_ref = EXCLUDED.actor_ref,
  decided_at = EXCLUDED.decided_at,
  stock_movement_ref = EXCLUDED.stock_movement_ref`

const upsertReturnAttachmentSQL = `
INSERT INTO returns.return_attachments (
  id,
  org_id,
  attachment_ref,
  return_order_id,
  return_ref,
  inspection_ref,
  file_name,
  file_ext,
  mime_type,
  file_size_bytes,
  storage_bucket,
  storage_key,
  uploaded_by_ref,
  uploaded_at,
  status,
  note,
  source,
  created_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3,
  $4::uuid,
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
)
ON CONFLICT (org_id, attachment_ref)
DO UPDATE SET
  return_order_id = EXCLUDED.return_order_id,
  return_ref = EXCLUDED.return_ref,
  inspection_ref = EXCLUDED.inspection_ref,
  file_name = EXCLUDED.file_name,
  file_ext = EXCLUDED.file_ext,
  mime_type = EXCLUDED.mime_type,
  file_size_bytes = EXCLUDED.file_size_bytes,
  storage_bucket = EXCLUDED.storage_bucket,
  storage_key = EXCLUDED.storage_key,
  uploaded_by_ref = EXCLUDED.uploaded_by_ref,
  uploaded_at = EXCLUDED.uploaded_at,
  status = EXCLUDED.status,
  note = EXCLUDED.note,
  source = EXCLUDED.source`

func (s PostgresReturnReceiptStore) List(
	ctx context.Context,
	filter domain.ReturnReceiptFilter,
) ([]domain.ReturnReceipt, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectReturnReceiptHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	receipts := make([]domain.ReturnReceipt, 0)
	for rows.Next() {
		receipt, err := scanPostgresReturnReceipt(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if returnReceiptMatchesFilter(receipt, filter) {
			receipts = append(receipts, receipt)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortReturnReceipts(receipts)

	return receipts, nil
}

func (s PostgresReturnReceiptStore) Save(ctx context.Context, receipt domain.ReturnReceipt) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(receipt.ID) == "" {
		return errors.New("return receipt id is required")
	}
	orgID, err := s.resolveOrgID(ctx, s.db, "")
	if err != nil {
		return err
	}
	if err := s.ensureReceiptNotDuplicate(ctx, orgID, receipt); err != nil {
		return err
	}

	return s.saveReceiptTx(ctx, func(tx *sql.Tx) error {
		persistedID, err := upsertReturnReceipt(ctx, tx, orgID, receipt)
		if err != nil {
			return err
		}
		return replaceReturnReceiptLines(ctx, tx, orgID, persistedID, receipt)
	})
}

func (s PostgresReturnReceiptStore) FindReceiptByID(
	ctx context.Context,
	id string,
) (domain.ReturnReceipt, error) {
	if s.db == nil {
		return domain.ReturnReceipt{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findReturnReceiptHeaderSQL, strings.TrimSpace(id))
	receipt, err := scanPostgresReturnReceipt(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ReturnReceipt{}, ErrReturnReceiptNotFound
	}
	if err != nil {
		return domain.ReturnReceipt{}, err
	}

	return receipt, nil
}

func (s PostgresReturnReceiptStore) SaveInspection(
	ctx context.Context,
	receipt domain.ReturnReceipt,
	inspection domain.ReturnInspection,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(receipt.ID) == "" || strings.TrimSpace(inspection.ID) == "" {
		return domain.ErrReturnInspectionRequiredField
	}
	orgID, err := s.resolveOrgID(ctx, s.db, "")
	if err != nil {
		return err
	}

	return s.saveReceiptTx(ctx, func(tx *sql.Tx) error {
		persistedID, err := upsertReturnReceipt(ctx, tx, orgID, receipt)
		if err != nil {
			return err
		}
		if err := replaceReturnReceiptLines(ctx, tx, orgID, persistedID, receipt); err != nil {
			return err
		}

		return upsertReturnInspection(ctx, tx, orgID, persistedID, receipt.ID, inspection)
	})
}

func (s PostgresReturnReceiptStore) FindInspectionByID(
	ctx context.Context,
	id string,
) (domain.ReturnInspection, error) {
	if s.db == nil {
		return domain.ReturnInspection{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findReturnInspectionSQL, strings.TrimSpace(id))
	inspection, err := scanPostgresReturnInspection(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ReturnInspection{}, ErrReturnInspectionNotFound
	}
	if err != nil {
		return domain.ReturnInspection{}, err
	}

	return inspection, nil
}

func (s PostgresReturnReceiptStore) SaveDisposition(
	ctx context.Context,
	receipt domain.ReturnReceipt,
	action domain.ReturnDispositionAction,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(receipt.ID) == "" || strings.TrimSpace(action.ID) == "" {
		return domain.ErrReturnDispositionRequiredField
	}
	orgID, err := s.resolveOrgID(ctx, s.db, "")
	if err != nil {
		return err
	}

	return s.saveReceiptTx(ctx, func(tx *sql.Tx) error {
		persistedID, err := upsertReturnReceipt(ctx, tx, orgID, receipt)
		if err != nil {
			return err
		}
		if err := replaceReturnReceiptLines(ctx, tx, orgID, persistedID, receipt); err != nil {
			return err
		}

		return upsertReturnDispositionAction(ctx, tx, orgID, persistedID, receipt.ID, receipt.StockMovement, action)
	})
}

func (s PostgresReturnReceiptStore) SaveAttachment(
	ctx context.Context,
	attachment domain.ReturnAttachment,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(attachment.ID) == "" {
		return domain.ErrReturnAttachmentRequiredField
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin return attachment transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, returnRef, err := findReturnReceiptPersistedID(ctx, tx, attachment.ReceiptID)
	if err != nil {
		return err
	}
	inspectionRef, err := findReturnInspectionForReceipt(ctx, tx, persistedID, attachment.InspectionID)
	if err != nil {
		return err
	}
	if err := upsertReturnAttachment(ctx, tx, orgID, persistedID, returnRef, inspectionRef, attachment); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit return attachment transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresReturnReceiptStore) FindExpectedReturnByCode(
	_ context.Context,
	code string,
) (domain.ExpectedReturn, error) {
	normalizedCode := domain.NormalizeReturnScanCode(code)
	if normalizedCode == "" {
		return domain.ExpectedReturn{}, ErrExpectedReturnNotFound
	}

	for _, expected := range s.expected {
		if expected.MatchesScanCode(normalizedCode) {
			return expected, nil
		}
	}

	return domain.ExpectedReturn{}, ErrExpectedReturnNotFound
}

func (s PostgresReturnReceiptStore) saveReceiptTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin return receipt transaction: %w", err)
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
		return fmt.Errorf("commit return receipt transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresReturnReceiptStore) ensureReceiptNotDuplicate(
	ctx context.Context,
	orgID string,
	receipt domain.ReturnReceipt,
) error {
	var existingID string
	err := s.db.QueryRowContext(
		ctx,
		findDuplicateReturnReceiptSQL,
		orgID,
		strings.TrimSpace(receipt.ID),
		strings.TrimSpace(receipt.ReceiptNo),
		domain.NormalizeReturnScanCode(receipt.ScanCode),
		domain.NormalizeReturnScanCode(receipt.TrackingNo),
		domain.NormalizeReturnScanCode(receipt.ReturnCode),
		domain.NormalizeReturnScanCode(receipt.OriginalOrderNo),
	).Scan(&existingID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("find duplicate return receipt: %w", err)
	}

	return ErrReturnReceiptDuplicate
}

func (s PostgresReturnReceiptStore) resolveOrgID(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isReturnReceiptUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectReturnReceiptOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isReturnReceiptUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve return receipt org %q: %w", orgRef, err)
		}
	}
	if isReturnReceiptUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("return receipt org %q cannot be resolved", orgRef)
}

func scanPostgresReturnReceipt(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	row interface{ Scan(dest ...any) error },
) (domain.ReturnReceipt, error) {
	header, err := scanPostgresReturnReceiptHeader(row)
	if err != nil {
		return domain.ReturnReceipt{}, err
	}
	lines, err := listPostgresReturnReceiptLines(ctx, queryer, header.PersistedID)
	if err != nil {
		return domain.ReturnReceipt{}, err
	}

	return buildPostgresReturnReceipt(header, lines)
}

func scanPostgresReturnReceiptHeader(row interface{ Scan(dest ...any) error }) (postgresReturnReceiptHeader, error) {
	var header postgresReturnReceiptHeader
	err := row.Scan(
		&header.PersistedID,
		&header.ID,
		&header.OrgID,
		&header.ReceiptNo,
		&header.WarehouseID,
		&header.WarehouseCode,
		&header.Source,
		&header.ReceivedBy,
		&header.ReceivedAt,
		&header.PackageCondition,
		&header.Status,
		&header.Disposition,
		&header.TargetLocation,
		&header.OriginalOrderNo,
		&header.TrackingNo,
		&header.ReturnCode,
		&header.ScanCode,
		&header.CustomerName,
		&header.UnknownCase,
		&header.StockMovementRef,
		&header.StockMovementType,
		&header.TargetStockStatus,
		&header.InvestigationNote,
		&header.CreatedAt,
	)
	if err != nil {
		return postgresReturnReceiptHeader{}, err
	}

	return header, nil
}

func listPostgresReturnReceiptLines(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	persistedID string,
) ([]domain.ReturnReceiptLine, error) {
	rows, err := queryer.QueryContext(ctx, selectReturnReceiptLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.ReturnReceiptLine, 0)
	for rows.Next() {
		line, err := scanPostgresReturnReceiptLine(rows)
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

func scanPostgresReturnReceiptLine(row interface{ Scan(dest ...any) error }) (domain.ReturnReceiptLine, error) {
	var (
		id           string
		sku          string
		productName  string
		quantityText string
		condition    string
	)
	if err := row.Scan(&id, &sku, &productName, &quantityText, &condition); err != nil {
		return domain.ReturnReceiptLine{}, err
	}
	quantity, err := postgresReturnReceiptQuantityInt(quantityText)
	if err != nil {
		return domain.ReturnReceiptLine{}, err
	}

	return domain.ReturnReceiptLine{
		ID:          strings.TrimSpace(id),
		SKU:         strings.ToUpper(strings.TrimSpace(sku)),
		ProductName: strings.TrimSpace(productName),
		Quantity:    quantity,
		Condition:   strings.TrimSpace(condition),
	}, nil
}

func buildPostgresReturnReceipt(
	header postgresReturnReceiptHeader,
	lines []domain.ReturnReceiptLine,
) (domain.ReturnReceipt, error) {
	expected := postgresReturnExpectedReturn(header, lines)
	input := domain.NewReturnReceiptInput{
		ID:                header.ID,
		ReceiptNo:         header.ReceiptNo,
		WarehouseID:       header.WarehouseID,
		WarehouseCode:     header.WarehouseCode,
		Source:            domain.ReturnSource(header.Source),
		ReceivedBy:        header.ReceivedBy,
		ScanCode:          firstNonBlankReturnReceipt(header.ScanCode, header.TrackingNo, header.ReturnCode, header.ReceiptNo),
		PackageCondition:  firstNonBlankReturnReceipt(header.PackageCondition, "pending inspection"),
		Disposition:       domain.ReturnDisposition(firstNonBlankReturnReceipt(header.Disposition, string(domain.ReturnDispositionNeedsInspection))),
		ExpectedReturn:    expected,
		InvestigationNote: header.InvestigationNote,
		CreatedAt:         header.CreatedAt,
	}
	receipt, err := domain.NewReturnReceipt(input)
	if err != nil {
		return domain.ReturnReceipt{}, err
	}
	receipt.ID = header.ID
	receipt.ReceiptNo = header.ReceiptNo
	receipt.WarehouseID = header.WarehouseID
	receipt.WarehouseCode = header.WarehouseCode
	receipt.Source = domain.NormalizeReturnSource(domain.ReturnSource(header.Source))
	receipt.ReceivedBy = header.ReceivedBy
	receipt.ReceivedAt = header.ReceivedAt.UTC()
	receipt.PackageCondition = firstNonBlankReturnReceipt(header.PackageCondition, receipt.PackageCondition)
	receipt.Status = postgresReturnReceiptStatus(header.Status)
	receipt.Disposition = domain.NormalizeReturnDisposition(domain.ReturnDisposition(header.Disposition))
	if receipt.Disposition == "" {
		receipt.Disposition = domain.ReturnDispositionNeedsInspection
	}
	receipt.TargetLocation = firstNonBlankReturnReceipt(header.TargetLocation, receipt.TargetLocation)
	receipt.OriginalOrderNo = header.OriginalOrderNo
	receipt.TrackingNo = header.TrackingNo
	receipt.ReturnCode = header.ReturnCode
	receipt.ScanCode = domain.NormalizeReturnScanCode(firstNonBlankReturnReceipt(header.ScanCode, receipt.ScanCode))
	receipt.CustomerName = firstNonBlankReturnReceipt(header.CustomerName, receipt.CustomerName)
	receipt.UnknownCase = header.UnknownCase
	if len(lines) > 0 {
		receipt.Lines = lines
	}
	receipt.StockMovement = postgresReturnStockMovement(header, receipt)
	receipt.InvestigationNote = header.InvestigationNote
	receipt.CreatedAt = header.CreatedAt.UTC()

	return receipt, nil
}

func postgresReturnExpectedReturn(
	header postgresReturnReceiptHeader,
	lines []domain.ReturnReceiptLine,
) *domain.ExpectedReturn {
	if header.UnknownCase {
		return nil
	}
	line := domain.ReturnReceiptLine{SKU: "UNKNOWN-SKU", ProductName: "Unknown return item", Quantity: 1}
	if len(lines) > 0 {
		line = lines[0]
	}
	if line.Quantity <= 0 {
		line.Quantity = 1
	}

	return &domain.ExpectedReturn{
		OrderNo:       header.OriginalOrderNo,
		TrackingNo:    header.TrackingNo,
		ReturnCode:    header.ReturnCode,
		CustomerName:  header.CustomerName,
		SKU:           line.SKU,
		ProductName:   line.ProductName,
		Quantity:      line.Quantity,
		WarehouseID:   header.WarehouseID,
		WarehouseCode: header.WarehouseCode,
		Source:        domain.NormalizeReturnSource(domain.ReturnSource(header.Source)),
	}
}

func postgresReturnStockMovement(
	header postgresReturnReceiptHeader,
	receipt domain.ReturnReceipt,
) *domain.ReturnStockMovement {
	if strings.TrimSpace(header.StockMovementRef) == "" {
		return nil
	}
	line := domain.ReturnReceiptLine{SKU: "UNKNOWN-SKU", Quantity: 1}
	if len(receipt.Lines) > 0 {
		line = receipt.Lines[0]
	}
	if line.Quantity <= 0 {
		line.Quantity = 1
	}

	return &domain.ReturnStockMovement{
		ID:                strings.TrimSpace(header.StockMovementRef),
		MovementType:      strings.TrimSpace(header.StockMovementType),
		SKU:               line.SKU,
		WarehouseID:       receipt.WarehouseID,
		Quantity:          line.Quantity,
		TargetStockStatus: strings.TrimSpace(header.TargetStockStatus),
		SourceDocID:       receipt.ID,
	}
}

func upsertReturnReceipt(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	orgID string,
	receipt domain.ReturnReceipt,
) (string, error) {
	movementRef, movementType, targetStockStatus := returnReceiptMovementSummary(receipt.StockMovement)
	var persistedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertReturnReceiptSQL,
		nullableReturnReceiptUUID(receipt.ID),
		orgID,
		nullableReturnReceiptText(receipt.ID),
		nullableReturnReceiptText(orgID),
		receipt.ReceiptNo,
		nullableReturnReceiptUUID(receipt.OriginalOrderNo),
		nullableReturnReceiptUUID(receipt.WarehouseID),
		nullableReturnReceiptText(receipt.WarehouseID),
		nullableReturnReceiptText(receipt.WarehouseCode),
		string(postgresReturnReceiptStatus(string(receipt.Status))),
		string(domain.NormalizeReturnSource(receipt.Source)),
		nullableReturnReceiptText(receipt.TrackingNo),
		nullableReturnReceiptText(receipt.ReturnCode),
		nullableReturnReceiptText(receipt.PackageCondition),
		string(receipt.Disposition),
		receipt.UnknownCase,
		nullableReturnReceiptText(receipt.InvestigationNote),
		postgresReturnReceiptTime(receipt.ReceivedAt, receipt.CreatedAt),
		nullableReturnReceiptUUID(receipt.ReceivedBy),
		nullableReturnReceiptText(receipt.ReceivedBy),
		postgresReturnReceiptTime(receipt.CreatedAt, receipt.ReceivedAt),
		nullableReturnReceiptUUID(receipt.ReceivedBy),
		postgresReturnReceiptTime(receipt.CreatedAt, receipt.ReceivedAt),
		nullableReturnReceiptUUID(receipt.ReceivedBy),
		string(receipt.Disposition),
		nullableReturnReceiptText(receipt.TargetLocation),
		nullableReturnReceiptText(receipt.OriginalOrderNo),
		nullableReturnReceiptText(receipt.OriginalOrderNo),
		nullableReturnReceiptText(receipt.CustomerName),
		nullableReturnReceiptText(receipt.ScanCode),
		nullableReturnReceiptText(movementRef),
		nullableReturnReceiptText(movementType),
		nullableReturnReceiptText(targetStockStatus),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert return receipt: %w", err)
	}

	return persistedID, nil
}

func replaceReturnReceiptLines(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	orgID string,
	persistedID string,
	receipt domain.ReturnReceipt,
) error {
	if _, err := queryer.ExecContext(ctx, deleteReturnReceiptLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete return receipt lines: %w", err)
	}
	for index, line := range receipt.Lines {
		quantity := line.Quantity
		if quantity <= 0 {
			quantity = 1
		}
		if _, err := queryer.ExecContext(
			ctx,
			insertReturnReceiptLineSQL,
			nullableReturnReceiptUUID(line.ID),
			orgID,
			persistedID,
			nullableReturnReceiptText(line.ID),
			index+1,
			nullableReturnReceiptUUID(line.SKU),
			nullableReturnReceiptText(line.SKU),
			nullableReturnReceiptText(line.SKU),
			nullableReturnReceiptText(line.ProductName),
			strconv.Itoa(quantity),
			strconv.Itoa(quantity),
			nullableReturnReceiptText(line.Condition),
			postgresReturnReceiptConditionCode(line.Condition),
			nullableReturnReceiptText(line.Condition),
			nil,
			nil,
			nil,
			"EA",
			"EA",
			nil,
			postgresReturnReceiptTime(receipt.CreatedAt, receipt.ReceivedAt),
			nullableReturnReceiptUUID(receipt.ReceivedBy),
		); err != nil {
			return fmt.Errorf("insert return receipt line: %w", err)
		}
	}

	return nil
}

func upsertReturnInspection(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	orgID string,
	persistedID string,
	returnRef string,
	inspection domain.ReturnInspection,
) error {
	_, err := queryer.ExecContext(
		ctx,
		upsertReturnInspectionSQL,
		nullableReturnReceiptUUID(inspection.ID),
		orgID,
		nullableReturnReceiptText(inspection.ID),
		persistedID,
		nullableReturnReceiptText(returnRef),
		string(inspection.Condition),
		string(inspection.Disposition),
		string(inspection.Status),
		inspection.TargetLocation,
		inspection.RiskLevel,
		nullableReturnReceiptText(inspection.EvidenceLabel),
		nullableReturnReceiptText(inspection.Note),
		inspection.InspectorID,
		inspection.InspectedAt.UTC(),
		inspection.InspectedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert return inspection: %w", err)
	}

	return nil
}

func scanPostgresReturnInspection(row interface{ Scan(dest ...any) error }) (domain.ReturnInspection, error) {
	var (
		id             string
		receiptID      string
		receiptNo      string
		condition      string
		disposition    string
		status         string
		targetLocation string
		riskLevel      string
		inspectorID    string
		note           string
		evidenceLabel  string
		inspectedAt    time.Time
	)
	if err := row.Scan(
		&id,
		&receiptID,
		&receiptNo,
		&condition,
		&disposition,
		&status,
		&targetLocation,
		&riskLevel,
		&inspectorID,
		&note,
		&evidenceLabel,
		&inspectedAt,
	); err != nil {
		return domain.ReturnInspection{}, err
	}
	inspection, err := domain.NewReturnInspection(domain.NewReturnInspectionInput{
		ID:            id,
		ReceiptID:     receiptID,
		ReceiptNo:     receiptNo,
		Condition:     domain.ReturnInspectionCondition(condition),
		Disposition:   domain.ReturnDisposition(disposition),
		InspectorID:   inspectorID,
		Note:          note,
		EvidenceLabel: evidenceLabel,
		InspectedAt:   inspectedAt,
	})
	if err != nil {
		return domain.ReturnInspection{}, err
	}
	inspection.Status = domain.ReturnInspectionStatus(status)
	inspection.TargetLocation = targetLocation
	inspection.RiskLevel = riskLevel

	return inspection, nil
}

func upsertReturnDispositionAction(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	orgID string,
	persistedID string,
	returnRef string,
	movement *domain.ReturnStockMovement,
	action domain.ReturnDispositionAction,
) error {
	movementRef, _, _ := returnReceiptMovementSummary(movement)
	_, err := queryer.ExecContext(
		ctx,
		upsertReturnDispositionActionSQL,
		nullableReturnReceiptUUID(action.ID),
		orgID,
		nullableReturnReceiptText(action.ID),
		persistedID,
		nullableReturnReceiptText(returnRef),
		string(action.Disposition),
		action.TargetLocation,
		nullableReturnReceiptText(action.TargetStockStatus),
		action.ActionCode,
		nullableReturnReceiptText(action.Note),
		action.ActorID,
		action.DecidedAt.UTC(),
		nullableReturnReceiptText(movementRef),
		action.DecidedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert return disposition action: %w", err)
	}

	return nil
}

func findReturnReceiptPersistedID(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	id string,
) (string, string, string, error) {
	var persistedID, orgID, returnRef string
	err := queryer.QueryRowContext(ctx, findReturnReceiptPersistedIDSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID, &returnRef)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", "", ErrReturnReceiptNotFound
	}
	if err != nil {
		return "", "", "", fmt.Errorf("find return receipt: %w", err)
	}

	return persistedID, orgID, returnRef, nil
}

func findReturnInspectionForReceipt(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	persistedID string,
	inspectionID string,
) (string, error) {
	var inspectionRef string
	err := queryer.QueryRowContext(ctx, findReturnInspectionForReceiptSQL, persistedID, strings.TrimSpace(inspectionID)).
		Scan(&inspectionRef)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrReturnInspectionNotFound
	}
	if err != nil {
		return "", fmt.Errorf("find return inspection: %w", err)
	}

	return inspectionRef, nil
}

func upsertReturnAttachment(
	ctx context.Context,
	queryer postgresReturnReceiptQueryer,
	orgID string,
	persistedID string,
	returnRef string,
	inspectionRef string,
	attachment domain.ReturnAttachment,
) error {
	_, err := queryer.ExecContext(
		ctx,
		upsertReturnAttachmentSQL,
		nullableReturnReceiptUUID(attachment.ID),
		orgID,
		nullableReturnReceiptText(attachment.ID),
		persistedID,
		nullableReturnReceiptText(returnRef),
		nullableReturnReceiptText(inspectionRef),
		attachment.FileName,
		nullableReturnReceiptText(attachment.FileExt),
		attachment.MIMEType,
		attachment.FileSizeBytes,
		attachment.StorageBucket,
		attachment.StorageKey,
		attachment.UploadedBy,
		attachment.UploadedAt.UTC(),
		attachment.Status,
		nullableReturnReceiptText(attachment.Note),
		"return_attachment",
		attachment.UploadedAt.UTC(),
	)
	if err != nil {
		return fmt.Errorf("upsert return attachment: %w", err)
	}

	return nil
}

func returnReceiptMatchesFilter(receipt domain.ReturnReceipt, filter domain.ReturnReceiptFilter) bool {
	if strings.TrimSpace(filter.WarehouseID) != "" && receipt.WarehouseID != strings.TrimSpace(filter.WarehouseID) {
		return false
	}
	if filter.Status != "" && receipt.Status != filter.Status {
		return false
	}

	return true
}

func returnReceiptMovementSummary(movement *domain.ReturnStockMovement) (string, string, string) {
	if movement == nil {
		return "", "", ""
	}

	return movement.ID, movement.MovementType, movement.TargetStockStatus
}

func postgresReturnReceiptStatus(value string) domain.ReturnReceiptStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "received", string(domain.ReturnStatusPendingInspection):
		return domain.ReturnStatusPendingInspection
	case string(domain.ReturnStatusInspected):
		return domain.ReturnStatusInspected
	case "disposed", string(domain.ReturnStatusDispositioned):
		return domain.ReturnStatusDispositioned
	default:
		return domain.ReturnStatusPendingInspection
	}
}

func postgresReturnReceiptQuantityInt(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("return receipt quantity is required")
	}
	whole := value
	if before, after, ok := strings.Cut(value, "."); ok {
		if strings.TrimRight(after, "0") != "" {
			return 0, fmt.Errorf("return receipt quantity %q is not an integer", value)
		}
		whole = before
	}
	quantity, err := strconv.Atoi(whole)
	if err != nil {
		return 0, err
	}
	if quantity <= 0 {
		return 0, errors.New("return receipt quantity must be positive")
	}

	return quantity, nil
}

func postgresReturnReceiptConditionCode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "sealed", "sealed_good", "intact":
		return "sealed_good"
	case "opened_good", "used", "seal_torn", "dented_box", "missing_accessory":
		return "opened_good"
	case "damaged":
		return "damaged"
	case "expired":
		return "expired"
	case "suspected_quality_issue":
		return "suspected_quality_issue"
	default:
		return "unknown"
	}
}

func postgresReturnReceiptTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary.UTC()
	}
	if !fallback.IsZero() {
		return fallback.UTC()
	}

	return time.Now().UTC()
}

func nullableReturnReceiptText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullableReturnReceiptUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isReturnReceiptUUIDText(value) {
		return nil
	}

	return value
}

func firstNonBlankReturnReceipt(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isReturnReceiptUUIDText(value string) bool {
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
			if !isReturnReceiptHexText(char) {
				return false
			}
		}
	}

	return true
}

func isReturnReceiptHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ ReturnReceiptStore = PostgresReturnReceiptStore{}
var _ ReturnInspectionStore = PostgresReturnReceiptStore{}
var _ ReturnDispositionStore = PostgresReturnReceiptStore{}
var _ ReturnAttachmentStore = PostgresReturnReceiptStore{}
