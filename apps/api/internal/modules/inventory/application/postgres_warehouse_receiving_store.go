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

type PostgresWarehouseReceivingStoreConfig struct {
	DefaultOrgID string
}

type PostgresWarehouseReceivingStore struct {
	db           *sql.DB
	defaultOrgID string
}

func NewPostgresWarehouseReceivingStore(
	db *sql.DB,
	cfg PostgresWarehouseReceivingStoreConfig,
) PostgresWarehouseReceivingStore {
	return PostgresWarehouseReceivingStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectWarehouseReceivingOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const listWarehouseReceivingsSQL = `
SELECT
  id::text,
  COALESCE(receipt_ref, id::text),
  COALESCE(org_ref, org_id::text),
  receipt_no,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(warehouse_code, ''),
  COALESCE(location_ref, location_id::text),
  COALESCE(location_code, ''),
  reference_doc_type,
  COALESCE(reference_doc_ref, reference_doc_id::text),
  COALESCE(supplier_ref, supplier_id::text, ''),
  COALESCE(delivery_note_no, ''),
  status,
  COALESCE(created_by_ref, created_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  inspect_ready_at,
  COALESCE(inspect_ready_by_ref, inspect_ready_by::text, ''),
  posted_at,
  COALESCE(posted_by_ref, posted_by::text, ''),
  created_at,
  updated_at
FROM inventory.warehouse_receivings
ORDER BY updated_at DESC, receipt_no ASC`

const findWarehouseReceivingPredicateSQL = `
SELECT
  id::text,
  COALESCE(receipt_ref, id::text),
  COALESCE(org_ref, org_id::text),
  receipt_no,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(warehouse_code, ''),
  COALESCE(location_ref, location_id::text),
  COALESCE(location_code, ''),
  reference_doc_type,
  COALESCE(reference_doc_ref, reference_doc_id::text),
  COALESCE(supplier_ref, supplier_id::text, ''),
  COALESCE(delivery_note_no, ''),
  status,
  COALESCE(created_by_ref, created_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  inspect_ready_at,
  COALESCE(inspect_ready_by_ref, inspect_ready_by::text, ''),
  posted_at,
  COALESCE(posted_by_ref, posted_by::text, ''),
  created_at,
  updated_at
FROM inventory.warehouse_receivings
WHERE receipt_ref = $1 OR id::text = $1
LIMIT 1`

const selectWarehouseReceivingLinesSQL = `
SELECT
  COALESCE(line_ref, id::text),
  COALESCE(purchase_order_line_ref, purchase_order_line_id::text, ''),
  COALESCE(item_ref, item_id::text, ''),
  sku_code,
  COALESCE(item_name, ''),
  COALESCE(batch_ref, batch_id::text, ''),
  COALESCE(batch_no, ''),
  COALESCE(lot_no, ''),
  COALESCE(expiry_date::text, ''),
  COALESCE(warehouse_ref, warehouse_id::text, ''),
  COALESCE(location_ref, location_id::text, ''),
  quantity::text,
  uom_code,
  base_uom_code,
  packaging_status,
  COALESCE(qc_status, '')
FROM inventory.warehouse_receiving_lines
WHERE receipt_id = $1::uuid
ORDER BY line_no, created_at, line_ref`

const selectWarehouseReceivingMovementsSQL = `
SELECT
  org_id::text,
  COALESCE(movement_ref, id::text),
  movement_no,
  movement_type,
  movement_at,
  COALESCE(item_ref, item_id::text, ''),
  COALESCE(batch_ref, batch_id::text, ''),
  COALESCE(warehouse_ref, warehouse_id::text, ''),
  COALESCE(location_ref, location_id::text, ''),
  quantity::text,
  base_uom_code,
  source_quantity::text,
  source_uom_code,
  conversion_factor::text,
  stock_status,
  source_doc_ref,
  COALESCE(source_doc_line_ref, ''),
  reason,
  COALESCE(created_by_ref, created_by::text, '')
FROM inventory.warehouse_receiving_stock_movements
WHERE receipt_id = $1::uuid
ORDER BY movement_no, created_at`

const upsertWarehouseReceivingSQL = `
INSERT INTO inventory.warehouse_receivings (
  id,
  org_id,
  receipt_ref,
  receipt_no,
  org_ref,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  location_id,
  location_ref,
  location_code,
  reference_doc_type,
  reference_doc_id,
  reference_doc_ref,
  supplier_id,
  supplier_ref,
  delivery_note_no,
  status,
  created_by,
  created_by_ref,
  submitted_at,
  submitted_by,
  submitted_by_ref,
  inspect_ready_at,
  inspect_ready_by,
  inspect_ready_by_ref,
  posted_at,
  posted_by,
  posted_by_ref,
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
  $9::uuid,
  $10,
  $11,
  $12,
  $13::uuid,
  $14,
  $15::uuid,
  $16,
  $17,
  $18,
  $19::uuid,
  $20,
  $21,
  $22::uuid,
  $23,
  $24,
  $25::uuid,
  $26,
  $27,
  $28::uuid,
  $29,
  $30,
  $31
)
ON CONFLICT ON CONSTRAINT uq_warehouse_receivings_org_ref
DO UPDATE SET
  receipt_no = EXCLUDED.receipt_no,
  org_ref = EXCLUDED.org_ref,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  location_id = EXCLUDED.location_id,
  location_ref = EXCLUDED.location_ref,
  location_code = EXCLUDED.location_code,
  reference_doc_type = EXCLUDED.reference_doc_type,
  reference_doc_id = EXCLUDED.reference_doc_id,
  reference_doc_ref = EXCLUDED.reference_doc_ref,
  supplier_id = EXCLUDED.supplier_id,
  supplier_ref = EXCLUDED.supplier_ref,
  delivery_note_no = EXCLUDED.delivery_note_no,
  status = EXCLUDED.status,
  created_by = EXCLUDED.created_by,
  created_by_ref = EXCLUDED.created_by_ref,
  submitted_at = EXCLUDED.submitted_at,
  submitted_by = EXCLUDED.submitted_by,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  inspect_ready_at = EXCLUDED.inspect_ready_at,
  inspect_ready_by = EXCLUDED.inspect_ready_by,
  inspect_ready_by_ref = EXCLUDED.inspect_ready_by_ref,
  posted_at = EXCLUDED.posted_at,
  posted_by = EXCLUDED.posted_by,
  posted_by_ref = EXCLUDED.posted_by_ref,
  updated_at = EXCLUDED.updated_at
RETURNING id::text`

const deleteWarehouseReceivingLinesSQL = `
DELETE FROM inventory.warehouse_receiving_lines
WHERE receipt_id = $1::uuid`

const deleteWarehouseReceivingMovementsSQL = `
DELETE FROM inventory.warehouse_receiving_stock_movements
WHERE receipt_id = $1::uuid`

const insertWarehouseReceivingLineSQL = `
INSERT INTO inventory.warehouse_receiving_lines (
  id,
  org_id,
  receipt_id,
  line_ref,
  line_no,
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
  base_uom_code,
  packaging_status,
  qc_status,
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
  $10,
  $11,
  $12::uuid,
  $13,
  $14,
  $15,
  $16::date,
  $17::uuid,
  $18,
  $19::uuid,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27
)`

const insertWarehouseReceivingMovementSQL = `
INSERT INTO inventory.warehouse_receiving_stock_movements (
  id,
  org_id,
  receipt_id,
  movement_ref,
  movement_no,
  movement_type,
  movement_at,
  item_id,
  item_ref,
  batch_id,
  batch_ref,
  warehouse_id,
  warehouse_ref,
  location_id,
  location_ref,
  quantity,
  base_uom_code,
  source_quantity,
  source_uom_code,
  conversion_factor,
  stock_status,
  source_doc_ref,
  source_doc_line_ref,
  reason,
  created_by,
  created_by_ref,
  created_at
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
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
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25::uuid,
  $26,
  $27
)`

func (s PostgresWarehouseReceivingStore) List(
	ctx context.Context,
	filter domain.WarehouseReceivingFilter,
) ([]domain.WarehouseReceiving, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, listWarehouseReceivingsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	receipts := make([]domain.WarehouseReceiving, 0)
	for rows.Next() {
		receipt, err := s.scanWarehouseReceiving(ctx, rows)
		if err != nil {
			return nil, err
		}
		if filter.Matches(receipt) {
			receipts = append(receipts, receipt)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortWarehouseReceivings(receipts)

	return receipts, nil
}

func (s PostgresWarehouseReceivingStore) Get(
	ctx context.Context,
	id string,
) (domain.WarehouseReceiving, error) {
	if s.db == nil {
		return domain.WarehouseReceiving{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findWarehouseReceivingPredicateSQL, strings.TrimSpace(id))
	receipt, err := s.scanWarehouseReceiving(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.WarehouseReceiving{}, ErrWarehouseReceivingNotFound
	}
	if err != nil {
		return domain.WarehouseReceiving{}, err
	}

	return receipt, nil
}

func (s PostgresWarehouseReceivingStore) Save(
	ctx context.Context,
	receipt domain.WarehouseReceiving,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := receipt.Validate(); err != nil {
		return err
	}
	orgID, err := s.resolveOrgID(ctx, receipt.OrgID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin warehouse receiving transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, err := upsertWarehouseReceiving(ctx, tx, orgID, receipt)
	if err != nil {
		return err
	}
	if err := replaceWarehouseReceivingLines(ctx, tx, orgID, persistedID, receipt); err != nil {
		return err
	}
	if err := replaceWarehouseReceivingMovements(ctx, tx, orgID, persistedID, receipt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit warehouse receiving transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresWarehouseReceivingStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := s.db.QueryRowContext(ctx, selectWarehouseReceivingOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve warehouse receiving org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("warehouse receiving org %q cannot be resolved", orgRef)
}

func (s PostgresWarehouseReceivingStore) scanWarehouseReceiving(
	ctx context.Context,
	row interface{ Scan(dest ...any) error },
) (domain.WarehouseReceiving, error) {
	var (
		persistedID    string
		id             string
		orgID          string
		receiptNo      string
		warehouseID    string
		warehouseCode  string
		locationID     string
		locationCode   string
		referenceType  string
		referenceID    string
		supplierID     string
		deliveryNoteNo string
		status         string
		createdBy      string
		submittedAt    sql.NullTime
		submittedBy    string
		inspectReadyAt sql.NullTime
		inspectReadyBy string
		postedAt       sql.NullTime
		postedBy       string
		createdAt      time.Time
		updatedAt      time.Time
	)
	if err := row.Scan(
		&persistedID,
		&id,
		&orgID,
		&receiptNo,
		&warehouseID,
		&warehouseCode,
		&locationID,
		&locationCode,
		&referenceType,
		&referenceID,
		&supplierID,
		&deliveryNoteNo,
		&status,
		&createdBy,
		&submittedAt,
		&submittedBy,
		&inspectReadyAt,
		&inspectReadyBy,
		&postedAt,
		&postedBy,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.WarehouseReceiving{}, err
	}

	lines, err := s.listWarehouseReceivingLines(ctx, persistedID)
	if err != nil {
		return domain.WarehouseReceiving{}, err
	}
	receipt, err := domain.NewWarehouseReceiving(domain.NewWarehouseReceivingInput{
		ID:               id,
		OrgID:            orgID,
		ReceiptNo:        receiptNo,
		WarehouseID:      warehouseID,
		WarehouseCode:    warehouseCode,
		LocationID:       locationID,
		LocationCode:     locationCode,
		ReferenceDocType: referenceType,
		ReferenceDocID:   referenceID,
		SupplierID:       supplierID,
		DeliveryNoteNo:   deliveryNoteNo,
		Status:           domain.WarehouseReceivingStatus(status),
		Lines:            lines,
		CreatedBy:        createdBy,
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
	})
	if err != nil {
		return domain.WarehouseReceiving{}, err
	}
	if submittedAt.Valid {
		receipt.SubmittedAt = submittedAt.Time.UTC()
		receipt.SubmittedBy = submittedBy
	}
	if inspectReadyAt.Valid {
		receipt.InspectReadyAt = inspectReadyAt.Time.UTC()
		receipt.InspectReadyBy = inspectReadyBy
	}
	if postedAt.Valid {
		receipt.PostedAt = postedAt.Time.UTC()
		receipt.PostedBy = postedBy
	}
	movements, err := s.listWarehouseReceivingMovements(ctx, persistedID)
	if err != nil {
		return domain.WarehouseReceiving{}, err
	}
	receipt.StockMovements = movements

	return receipt, nil
}

func (s PostgresWarehouseReceivingStore) listWarehouseReceivingLines(
	ctx context.Context,
	receiptID string,
) ([]domain.NewWarehouseReceivingLineInput, error) {
	rows, err := s.db.QueryContext(ctx, selectWarehouseReceivingLinesSQL, receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.NewWarehouseReceivingLineInput, 0)
	for rows.Next() {
		line, err := scanWarehouseReceivingLineInput(rows)
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

func scanWarehouseReceivingLineInput(
	row interface{ Scan(dest ...any) error },
) (domain.NewWarehouseReceivingLineInput, error) {
	var (
		id                  string
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
		baseUOMCode         string
		packagingStatus     string
		qcStatus            string
	)
	if err := row.Scan(
		&id,
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
		&baseUOMCode,
		&packagingStatus,
		&qcStatus,
	); err != nil {
		return domain.NewWarehouseReceivingLineInput{}, err
	}
	quantity, err := decimal.ParseQuantity(quantityText)
	if err != nil {
		return domain.NewWarehouseReceivingLineInput{}, err
	}
	expiryDate := time.Time{}
	if strings.TrimSpace(expiryDateText) != "" {
		expiryDate, err = time.Parse("2006-01-02", expiryDateText)
		if err != nil {
			return domain.NewWarehouseReceivingLineInput{}, err
		}
	}

	return domain.NewWarehouseReceivingLineInput{
		ID:                  id,
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
		BaseUOMCode:         baseUOMCode,
		PackagingStatus:     domain.ReceivingPackagingStatus(packagingStatus),
		QCStatus:            domain.QCStatus(qcStatus),
	}, nil
}

func (s PostgresWarehouseReceivingStore) listWarehouseReceivingMovements(
	ctx context.Context,
	receiptID string,
) ([]domain.StockMovement, error) {
	rows, err := s.db.QueryContext(ctx, selectWarehouseReceivingMovementsSQL, receiptID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movements := make([]domain.StockMovement, 0)
	for rows.Next() {
		movement, err := scanWarehouseReceivingMovement(rows)
		if err != nil {
			return nil, err
		}
		movements = append(movements, movement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return movements, nil
}

func scanWarehouseReceivingMovement(row interface{ Scan(dest ...any) error }) (domain.StockMovement, error) {
	var (
		orgID           string
		id              string
		movementNo      string
		movementType    string
		movementAt      time.Time
		itemID          string
		batchID         string
		warehouseID     string
		locationID      string
		quantityText    string
		baseUOMCode     string
		sourceQtyText   string
		sourceUOMCode   string
		conversionText  string
		stockStatus     string
		sourceDocID     string
		sourceDocLineID string
		reason          string
		createdBy       string
	)
	if err := row.Scan(
		&orgID,
		&id,
		&movementNo,
		&movementType,
		&movementAt,
		&itemID,
		&batchID,
		&warehouseID,
		&locationID,
		&quantityText,
		&baseUOMCode,
		&sourceQtyText,
		&sourceUOMCode,
		&conversionText,
		&stockStatus,
		&sourceDocID,
		&sourceDocLineID,
		&reason,
		&createdBy,
	); err != nil {
		return domain.StockMovement{}, err
	}
	quantity, err := decimal.ParseQuantity(quantityText)
	if err != nil {
		return domain.StockMovement{}, err
	}
	sourceQuantity, err := decimal.ParseQuantity(sourceQtyText)
	if err != nil {
		return domain.StockMovement{}, err
	}
	conversionFactor, err := decimal.ParseQuantity(conversionText)
	if err != nil {
		return domain.StockMovement{}, err
	}

	return domain.NewStockMovement(domain.NewStockMovementInput{
		MovementNo:       movementNo,
		MovementType:     domain.MovementType(movementType),
		OrgID:            orgID,
		ItemID:           itemID,
		BatchID:          batchID,
		WarehouseID:      warehouseID,
		BinID:            locationID,
		Quantity:         quantity,
		BaseUOMCode:      baseUOMCode,
		SourceQuantity:   sourceQuantity,
		SourceUOMCode:    sourceUOMCode,
		ConversionFactor: conversionFactor,
		StockStatus:      domain.StockStatus(stockStatus),
		SourceDocType:    receivingSourceDocType,
		SourceDocID:      sourceDocID,
		SourceDocLineID:  sourceDocLineID,
		Reason:           reason,
		CreatedBy:        createdBy,
		MovementAt:       movementAt,
	})
}

func upsertWarehouseReceiving(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	receipt domain.WarehouseReceiving,
) (string, error) {
	var persistedID string
	err := tx.QueryRowContext(
		ctx,
		upsertWarehouseReceivingSQL,
		nullableUUID(receipt.ID),
		orgID,
		nullableText(receipt.ID),
		receipt.ReceiptNo,
		nullableText(receipt.OrgID),
		nullableUUID(receipt.WarehouseID),
		nullableText(receipt.WarehouseID),
		nullableText(receipt.WarehouseCode),
		nullableUUID(receipt.LocationID),
		nullableText(receipt.LocationID),
		nullableText(receipt.LocationCode),
		receipt.ReferenceDocType,
		nullableUUID(receipt.ReferenceDocID),
		nullableText(receipt.ReferenceDocID),
		nullableUUID(receipt.SupplierID),
		nullableText(receipt.SupplierID),
		nullableText(receipt.DeliveryNoteNo),
		string(receipt.Status),
		nullableUUID(receipt.CreatedBy),
		nullableText(receipt.CreatedBy),
		nullableTime(receipt.SubmittedAt),
		nullableUUID(receipt.SubmittedBy),
		nullableText(receipt.SubmittedBy),
		nullableTime(receipt.InspectReadyAt),
		nullableUUID(receipt.InspectReadyBy),
		nullableText(receipt.InspectReadyBy),
		nullableTime(receipt.PostedAt),
		nullableUUID(receipt.PostedBy),
		nullableText(receipt.PostedBy),
		receipt.CreatedAt.UTC(),
		receipt.UpdatedAt.UTC(),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert warehouse receiving: %w", err)
	}

	return persistedID, nil
}

func replaceWarehouseReceivingLines(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	receipt domain.WarehouseReceiving,
) error {
	if _, err := tx.ExecContext(ctx, deleteWarehouseReceivingLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete warehouse receiving lines: %w", err)
	}
	for index, line := range receipt.Lines {
		if _, err := tx.ExecContext(
			ctx,
			insertWarehouseReceivingLineSQL,
			nullableUUID(line.ID),
			orgID,
			persistedID,
			nullableText(line.ID),
			index+1,
			nullableUUID(line.PurchaseOrderLineID),
			nullableText(line.PurchaseOrderLineID),
			nullableUUID(line.ItemID),
			nullableText(line.ItemID),
			line.SKU,
			nullableText(line.ItemName),
			nullableUUID(line.BatchID),
			nullableText(line.BatchID),
			nullableText(line.BatchNo),
			nullableText(line.LotNo),
			nullableDate(line.ExpiryDate),
			nullableUUID(line.WarehouseID),
			nullableText(line.WarehouseID),
			nullableUUID(line.LocationID),
			nullableText(line.LocationID),
			line.Quantity.String(),
			line.UOMCode.String(),
			line.BaseUOMCode.String(),
			string(line.PackagingStatus),
			nullableText(string(line.QCStatus)),
			receipt.CreatedAt.UTC(),
			receipt.UpdatedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert warehouse receiving line: %w", err)
		}
	}

	return nil
}

func replaceWarehouseReceivingMovements(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	receipt domain.WarehouseReceiving,
) error {
	if _, err := tx.ExecContext(ctx, deleteWarehouseReceivingMovementsSQL, persistedID); err != nil {
		return fmt.Errorf("delete warehouse receiving movements: %w", err)
	}
	for _, movement := range receipt.StockMovements {
		movementRef := movement.MovementNo
		if strings.TrimSpace(movement.SourceDocLineID) != "" {
			movementRef = movement.SourceDocLineID
		}
		if _, err := tx.ExecContext(
			ctx,
			insertWarehouseReceivingMovementSQL,
			nullableUUID(movementRef),
			orgID,
			persistedID,
			nullableText(movementRef),
			movement.MovementNo,
			string(movement.MovementType),
			movement.MovementAt.UTC(),
			nullableUUID(movement.ItemID),
			nullableText(movement.ItemID),
			nullableUUID(movement.BatchID),
			nullableText(movement.BatchID),
			nullableUUID(movement.WarehouseID),
			nullableText(movement.WarehouseID),
			nullableUUID(movement.BinID),
			nullableText(movement.BinID),
			movement.Quantity.String(),
			movement.BaseUOMCode.String(),
			movement.SourceQuantity.String(),
			movement.SourceUOMCode.String(),
			movement.ConversionFactor.String(),
			string(movement.StockStatus),
			movement.SourceDocID,
			nullableText(movement.SourceDocLineID),
			movement.Reason,
			nullableUUID(movement.CreatedBy),
			nullableText(movement.CreatedBy),
			movement.CreatedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert warehouse receiving movement: %w", err)
		}
	}

	return nil
}

func nullableDate(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC().Format("2006-01-02")
}

var _ WarehouseReceivingStore = PostgresWarehouseReceivingStore{}
