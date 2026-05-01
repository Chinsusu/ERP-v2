package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresStockAdjustmentStoreConfig struct {
	DefaultOrgID string
}

type PostgresStockAdjustmentStore struct {
	db           *sql.DB
	defaultOrgID string
}

func NewPostgresStockAdjustmentStore(db *sql.DB, cfg PostgresStockAdjustmentStoreConfig) PostgresStockAdjustmentStore {
	return PostgresStockAdjustmentStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectStockAdjustmentOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const listStockAdjustmentsSQL = `
SELECT
  id::text,
  COALESCE(adjustment_ref, id::text),
  COALESCE(org_ref, org_id::text),
  adjustment_no,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(warehouse_code, ''),
  source_type,
  COALESCE(source_ref, source_id::text, ''),
  reason,
  status,
  COALESCE(requested_by_ref, requested_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  approved_at,
  COALESCE(approved_by_ref, approved_by::text, ''),
  rejected_at,
  COALESCE(rejected_by_ref, rejected_by::text, ''),
  posted_at,
  COALESCE(posted_by_ref, posted_by::text, ''),
  created_at,
  updated_at
FROM inventory.stock_adjustments
ORDER BY created_at DESC, adjustment_no DESC`

const findStockAdjustmentPredicateSQL = `
SELECT
  id::text,
  COALESCE(adjustment_ref, id::text),
  COALESCE(org_ref, org_id::text),
  adjustment_no,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(warehouse_code, ''),
  source_type,
  COALESCE(source_ref, source_id::text, ''),
  reason,
  status,
  COALESCE(requested_by_ref, requested_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  approved_at,
  COALESCE(approved_by_ref, approved_by::text, ''),
  rejected_at,
  COALESCE(rejected_by_ref, rejected_by::text, ''),
  posted_at,
  COALESCE(posted_by_ref, posted_by::text, ''),
  created_at,
  updated_at
FROM inventory.stock_adjustments
WHERE adjustment_ref = $1 OR id::text = $1
LIMIT 1`

const selectStockAdjustmentLinesSQL = `
SELECT
  COALESCE(line_ref, id::text),
  COALESCE(item_ref, item_id::text, ''),
  sku_code,
  COALESCE(batch_ref, batch_id::text, ''),
  COALESCE(batch_no, ''),
  COALESCE(location_ref, location_id::text, ''),
  COALESCE(location_code, ''),
  expected_qty::text,
  counted_qty::text,
  base_uom_code,
  COALESCE(reason, '')
FROM inventory.stock_adjustment_lines
WHERE adjustment_id = $1::uuid
ORDER BY line_no, created_at, line_ref`

const upsertStockAdjustmentSQL = `
INSERT INTO inventory.stock_adjustments (
  id,
  org_id,
  adjustment_ref,
  adjustment_no,
  org_ref,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  source_type,
  source_id,
  source_ref,
  reason,
  status,
  requested_by,
  requested_by_ref,
  submitted_at,
  submitted_by,
  submitted_by_ref,
  approved_at,
  approved_by,
  approved_by_ref,
  rejected_at,
  rejected_by,
  rejected_by_ref,
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
  $9,
  $10::uuid,
  $11,
  $12,
  $13,
  $14::uuid,
  $15,
  $16,
  $17::uuid,
  $18,
  $19,
  $20::uuid,
  $21,
  $22,
  $23::uuid,
  $24,
  $25,
  $26::uuid,
  $27,
  $28,
  $29
)
ON CONFLICT ON CONSTRAINT uq_stock_adjustments_org_ref
DO UPDATE SET
  adjustment_no = EXCLUDED.adjustment_no,
  org_ref = EXCLUDED.org_ref,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  source_type = EXCLUDED.source_type,
  source_id = EXCLUDED.source_id,
  source_ref = EXCLUDED.source_ref,
  reason = EXCLUDED.reason,
  status = EXCLUDED.status,
  requested_by = EXCLUDED.requested_by,
  requested_by_ref = EXCLUDED.requested_by_ref,
  submitted_at = EXCLUDED.submitted_at,
  submitted_by = EXCLUDED.submitted_by,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  approved_at = EXCLUDED.approved_at,
  approved_by = EXCLUDED.approved_by,
  approved_by_ref = EXCLUDED.approved_by_ref,
  rejected_at = EXCLUDED.rejected_at,
  rejected_by = EXCLUDED.rejected_by,
  rejected_by_ref = EXCLUDED.rejected_by_ref,
  posted_at = EXCLUDED.posted_at,
  posted_by = EXCLUDED.posted_by,
  posted_by_ref = EXCLUDED.posted_by_ref,
  updated_at = EXCLUDED.updated_at
RETURNING id::text`

const deleteStockAdjustmentLinesSQL = `
DELETE FROM inventory.stock_adjustment_lines
WHERE adjustment_id = $1::uuid`

const insertStockAdjustmentLineSQL = `
INSERT INTO inventory.stock_adjustment_lines (
  id,
  org_id,
  adjustment_id,
  line_ref,
  line_no,
  item_id,
  item_ref,
  sku_code,
  batch_id,
  batch_ref,
  batch_no,
  location_id,
  location_ref,
  location_code,
  expected_qty,
  counted_qty,
  delta_qty,
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
  $8,
  $9::uuid,
  $10,
  $11,
  $12::uuid,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21
)`

func (s PostgresStockAdjustmentStore) ListStockAdjustments(ctx context.Context) ([]domain.StockAdjustment, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, listStockAdjustmentsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	adjustments, err := s.scanStockAdjustments(ctx, rows)
	if err != nil {
		return nil, err
	}
	domain.SortStockAdjustments(adjustments)

	return adjustments, nil
}

func (s PostgresStockAdjustmentStore) FindStockAdjustmentByID(
	ctx context.Context,
	id string,
) (domain.StockAdjustment, error) {
	if s.db == nil {
		return domain.StockAdjustment{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findStockAdjustmentPredicateSQL, strings.TrimSpace(id))
	adjustment, err := s.scanStockAdjustment(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StockAdjustment{}, ErrStockAdjustmentNotFound
	}
	if err != nil {
		return domain.StockAdjustment{}, err
	}

	return adjustment, nil
}

func (s PostgresStockAdjustmentStore) SaveStockAdjustment(
	ctx context.Context,
	adjustment domain.StockAdjustment,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := adjustment.Validate(); err != nil {
		return err
	}
	orgID, err := s.resolveOrgID(ctx, adjustment.OrgID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin stock adjustment transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, err := upsertStockAdjustment(ctx, tx, orgID, adjustment)
	if err != nil {
		return err
	}
	if err := replaceStockAdjustmentLines(ctx, tx, orgID, persistedID, adjustment); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stock adjustment transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresStockAdjustmentStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := s.db.QueryRowContext(ctx, selectStockAdjustmentOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve stock adjustment org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("stock adjustment org %q cannot be resolved", orgRef)
}

func (s PostgresStockAdjustmentStore) scanStockAdjustments(
	ctx context.Context,
	rows *sql.Rows,
) ([]domain.StockAdjustment, error) {
	adjustments := make([]domain.StockAdjustment, 0)
	for rows.Next() {
		adjustment, err := s.scanStockAdjustment(ctx, rows)
		if err != nil {
			return nil, err
		}
		adjustments = append(adjustments, adjustment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return adjustments, nil
}

func (s PostgresStockAdjustmentStore) scanStockAdjustment(
	ctx context.Context,
	row interface{ Scan(dest ...any) error },
) (domain.StockAdjustment, error) {
	var (
		persistedID   string
		adjustment    domain.StockAdjustment
		submittedAt   sql.NullTime
		submittedBy   string
		approvedAt    sql.NullTime
		approvedBy    string
		rejectedAt    sql.NullTime
		rejectedBy    string
		postedAt      sql.NullTime
		postedBy      string
		status        string
		requestedBy   string
		sourceID      string
		warehouseID   string
		warehouseCode string
		orgID         string
	)
	if err := row.Scan(
		&persistedID,
		&adjustment.ID,
		&orgID,
		&adjustment.AdjustmentNo,
		&warehouseID,
		&warehouseCode,
		&adjustment.SourceType,
		&sourceID,
		&adjustment.Reason,
		&status,
		&requestedBy,
		&submittedAt,
		&submittedBy,
		&approvedAt,
		&approvedBy,
		&rejectedAt,
		&rejectedBy,
		&postedAt,
		&postedBy,
		&adjustment.CreatedAt,
		&adjustment.UpdatedAt,
	); err != nil {
		return domain.StockAdjustment{}, err
	}
	adjustment.OrgID = orgID
	adjustment.WarehouseID = warehouseID
	adjustment.WarehouseCode = strings.ToUpper(strings.TrimSpace(warehouseCode))
	adjustment.SourceID = sourceID
	adjustment.Status = domain.StockAdjustmentStatus(status)
	adjustment.RequestedBy = requestedBy
	if submittedAt.Valid {
		adjustment.SubmittedAt = submittedAt.Time.UTC()
		adjustment.SubmittedBy = submittedBy
	}
	if approvedAt.Valid {
		adjustment.ApprovedAt = approvedAt.Time.UTC()
		adjustment.ApprovedBy = approvedBy
	}
	if rejectedAt.Valid {
		adjustment.RejectedAt = rejectedAt.Time.UTC()
		adjustment.RejectedBy = rejectedBy
	}
	if postedAt.Valid {
		adjustment.PostedAt = postedAt.Time.UTC()
		adjustment.PostedBy = postedBy
	}

	lines, err := s.listStockAdjustmentLines(ctx, persistedID)
	if err != nil {
		return domain.StockAdjustment{}, err
	}
	adjustment.Lines = lines
	if err := adjustment.Validate(); err != nil {
		return domain.StockAdjustment{}, err
	}

	return adjustment, nil
}

func (s PostgresStockAdjustmentStore) listStockAdjustmentLines(
	ctx context.Context,
	adjustmentID string,
) ([]domain.StockAdjustmentLine, error) {
	rows, err := s.db.QueryContext(ctx, selectStockAdjustmentLinesSQL, adjustmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.StockAdjustmentLine, 0)
	for rows.Next() {
		line, err := scanStockAdjustmentLine(rows)
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

func scanStockAdjustmentLine(row interface{ Scan(dest ...any) error }) (domain.StockAdjustmentLine, error) {
	var (
		id          string
		itemID      string
		sku         string
		batchID     string
		batchNo     string
		locationID  string
		location    string
		expectedQty string
		countedQty  string
		baseUOMCode string
		reason      string
	)
	if err := row.Scan(
		&id,
		&itemID,
		&sku,
		&batchID,
		&batchNo,
		&locationID,
		&location,
		&expectedQty,
		&countedQty,
		&baseUOMCode,
		&reason,
	); err != nil {
		return domain.StockAdjustmentLine{}, err
	}
	expected, err := decimal.ParseQuantity(expectedQty)
	if err != nil {
		return domain.StockAdjustmentLine{}, err
	}
	counted, err := decimal.ParseQuantity(countedQty)
	if err != nil {
		return domain.StockAdjustmentLine{}, err
	}

	return domain.NewStockAdjustmentLine(domain.NewStockAdjustmentLineInput{
		ID:           id,
		ItemID:       itemID,
		SKU:          sku,
		BatchID:      batchID,
		BatchNo:      batchNo,
		LocationID:   locationID,
		LocationCode: location,
		ExpectedQty:  expected,
		CountedQty:   counted,
		BaseUOMCode:  baseUOMCode,
		Reason:       reason,
	})
}

func upsertStockAdjustment(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	adjustment domain.StockAdjustment,
) (string, error) {
	var persistedID string
	err := tx.QueryRowContext(
		ctx,
		upsertStockAdjustmentSQL,
		nullableUUID(adjustment.ID),
		orgID,
		nullableText(adjustment.ID),
		adjustment.AdjustmentNo,
		nullableText(adjustment.OrgID),
		nullableUUID(adjustment.WarehouseID),
		nullableText(adjustment.WarehouseID),
		nullableText(adjustment.WarehouseCode),
		adjustment.SourceType,
		nullableUUID(adjustment.SourceID),
		nullableText(adjustment.SourceID),
		adjustment.Reason,
		string(adjustment.Status),
		nullableUUID(adjustment.RequestedBy),
		nullableText(adjustment.RequestedBy),
		nullableTime(adjustment.SubmittedAt),
		nullableUUID(adjustment.SubmittedBy),
		nullableText(adjustment.SubmittedBy),
		nullableTime(adjustment.ApprovedAt),
		nullableUUID(adjustment.ApprovedBy),
		nullableText(adjustment.ApprovedBy),
		nullableTime(adjustment.RejectedAt),
		nullableUUID(adjustment.RejectedBy),
		nullableText(adjustment.RejectedBy),
		nullableTime(adjustment.PostedAt),
		nullableUUID(adjustment.PostedBy),
		nullableText(adjustment.PostedBy),
		adjustment.CreatedAt.UTC(),
		adjustment.UpdatedAt.UTC(),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert stock adjustment: %w", err)
	}

	return persistedID, nil
}

func replaceStockAdjustmentLines(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	adjustment domain.StockAdjustment,
) error {
	if _, err := tx.ExecContext(ctx, deleteStockAdjustmentLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete stock adjustment lines: %w", err)
	}
	for index, line := range adjustment.Lines {
		if _, err := tx.ExecContext(
			ctx,
			insertStockAdjustmentLineSQL,
			nullableUUID(line.ID),
			orgID,
			persistedID,
			nullableText(line.ID),
			index+1,
			nullableUUID(line.ItemID),
			nullableText(line.ItemID),
			line.SKU,
			nullableUUID(line.BatchID),
			nullableText(line.BatchID),
			nullableText(line.BatchNo),
			nullableUUID(line.LocationID),
			nullableText(line.LocationID),
			nullableText(line.LocationCode),
			line.ExpectedQty.String(),
			line.CountedQty.String(),
			line.DeltaQty.String(),
			line.BaseUOMCode.String(),
			nullableText(line.Reason),
			adjustment.CreatedAt.UTC(),
			adjustment.UpdatedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert stock adjustment line: %w", err)
		}
	}

	return nil
}

var _ StockAdjustmentStore = PostgresStockAdjustmentStore{}
