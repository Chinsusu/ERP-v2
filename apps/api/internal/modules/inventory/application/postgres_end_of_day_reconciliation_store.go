package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

type PostgresEndOfDayReconciliationStoreConfig struct {
	DefaultOrgID string
}

type PostgresEndOfDayReconciliationStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresEndOfDayReconciliationQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresEndOfDayReconciliationStore(
	db *sql.DB,
	cfg PostgresEndOfDayReconciliationStoreConfig,
) PostgresEndOfDayReconciliationStore {
	return PostgresEndOfDayReconciliationStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectEndOfDayReconciliationOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectEndOfDayReconciliationWarehouseSQL = `
SELECT id::text
FROM mdm.warehouses
WHERE org_id = $1::uuid
  AND (
    id::text = $2
    OR lower(code) = lower($2)
    OR lower(code) = lower($3)
  )
LIMIT 1`

const selectEndOfDayReconciliationHeadersBaseSQL = `
SELECT
  closing.id::text,
  COALESCE(closing.closing_ref, closing.closing_no, closing.id::text),
  COALESCE(closing.warehouse_ref, closing.warehouse_id::text, ''),
  COALESCE(closing.warehouse_code, warehouse.code, ''),
  closing.business_date::text,
  closing.shift_code,
  closing.status,
  COALESCE(closing.owner_ref, closing.created_by_ref, closing.created_by::text, ''),
  closing.orders_processed_count,
  closing.handover_order_count,
  closing.return_order_count,
  closing.stock_movement_count,
  closing.stock_count_session_count,
  closing.pending_task_count,
  closing.closed_at,
  COALESCE(closing.closed_by_ref, closing.closed_by::text, ''),
  COALESCE(closing.exception_note, '')
FROM inventory.warehouse_daily_closings AS closing
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = closing.warehouse_id`

const selectEndOfDayReconciliationHeadersSQL = selectEndOfDayReconciliationHeadersBaseSQL + `
ORDER BY closing.business_date DESC, COALESCE(closing.warehouse_code, warehouse.code, ''), closing.shift_code`

const findEndOfDayReconciliationHeaderSQL = selectEndOfDayReconciliationHeadersBaseSQL + `
WHERE lower(COALESCE(closing.closing_ref, closing.closing_no, closing.id::text)) = lower($1)
   OR closing.id::text = $1
LIMIT 1`

const findPersistedEndOfDayReconciliationSQL = `
SELECT id::text, org_id::text
FROM inventory.warehouse_daily_closings AS closing
WHERE lower(COALESCE(closing.closing_ref, closing.closing_no, closing.id::text)) = lower($1)
   OR closing.id::text = $1
LIMIT 1
FOR UPDATE`

const selectEndOfDayReconciliationChecklistSQL = `
SELECT
  item_ref,
  label,
  complete,
  blocking,
  COALESCE(note, '')
FROM inventory.warehouse_daily_closing_checklist
WHERE closing_id = $1::uuid
ORDER BY created_at, item_ref`

const selectEndOfDayReconciliationLinesSQL = `
SELECT
  line_ref,
  sku_code,
  COALESCE(batch_no, ''),
  COALESCE(bin_code, ''),
  system_qty::text,
  counted_qty::text,
  COALESCE(reason, ''),
  COALESCE(owner_ref, '')
FROM inventory.warehouse_daily_closing_lines
WHERE closing_id = $1::uuid
ORDER BY line_no, created_at, line_ref`

const insertEndOfDayReconciliationSQL = `
INSERT INTO inventory.warehouse_daily_closings (
  id,
  org_id,
  closing_ref,
  closing_no,
  org_ref,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  business_date,
  shift_code,
  status,
  orders_processed_count,
  handover_order_count,
  return_order_count,
  stock_movement_count,
  stock_count_session_count,
  pending_task_count,
  variance_count,
  exception_note,
  closed_at,
  closed_by,
  closed_by_ref,
  created_by,
  created_by_ref,
  updated_by,
  updated_by_ref,
  owner_ref,
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
  $9::date,
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
  $21::uuid,
  $22,
  $23::uuid,
  $24,
  $25::uuid,
  $26,
  $27,
  $28,
  $29
)
RETURNING id::text`

const updateEndOfDayReconciliationSQL = `
UPDATE inventory.warehouse_daily_closings
SET closing_ref = $2,
    closing_no = $3,
    org_ref = $4,
    warehouse_id = $5::uuid,
    warehouse_ref = $6,
    warehouse_code = $7,
    business_date = $8::date,
    shift_code = $9,
    status = $10,
    orders_processed_count = $11,
    handover_order_count = $12,
    return_order_count = $13,
    stock_movement_count = $14,
    stock_count_session_count = $15,
    pending_task_count = $16,
    variance_count = $17,
    exception_note = $18,
    closed_at = $19,
    closed_by = $20::uuid,
    closed_by_ref = $21,
    updated_by = $22::uuid,
    updated_by_ref = $23,
    owner_ref = $24,
    updated_at = $25,
    version = version + 1
WHERE id = $1::uuid
RETURNING id::text`

const deleteEndOfDayReconciliationChecklistSQL = `
DELETE FROM inventory.warehouse_daily_closing_checklist
WHERE closing_id = $1::uuid`

const deleteEndOfDayReconciliationLinesSQL = `
DELETE FROM inventory.warehouse_daily_closing_lines
WHERE closing_id = $1::uuid`

const insertEndOfDayReconciliationChecklistSQL = `
INSERT INTO inventory.warehouse_daily_closing_checklist (
  id,
  org_id,
  closing_id,
  item_ref,
  label,
  complete,
  blocking,
  note,
  created_at,
  updated_at
) VALUES (
  gen_random_uuid(),
  $1::uuid,
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9
)`

const insertEndOfDayReconciliationLineSQL = `
INSERT INTO inventory.warehouse_daily_closing_lines (
  id,
  org_id,
  closing_id,
  line_ref,
  line_no,
  sku_code,
  batch_no,
  bin_code,
  system_qty,
  counted_qty,
  reason,
  owner_ref,
  created_at,
  updated_at
) VALUES (
  gen_random_uuid(),
  $1::uuid,
  $2::uuid,
  $3,
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13
)`

func (s PostgresEndOfDayReconciliationStore) List(
	ctx context.Context,
	filter domain.EndOfDayReconciliationFilter,
) ([]domain.EndOfDayReconciliation, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}

	rows, err := s.db.QueryContext(ctx, selectEndOfDayReconciliationHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reconciliations := make([]domain.EndOfDayReconciliation, 0)
	for rows.Next() {
		reconciliation, err := scanPostgresEndOfDayReconciliation(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		if endOfDayReconciliationMatchesFilter(reconciliation, filter) {
			reconciliations = append(reconciliations, reconciliation)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortEndOfDayReconciliations(reconciliations)

	return reconciliations, nil
}

func (s PostgresEndOfDayReconciliationStore) Get(
	ctx context.Context,
	id string,
) (domain.EndOfDayReconciliation, error) {
	if s.db == nil {
		return domain.EndOfDayReconciliation{}, errors.New("database connection is required")
	}

	row := s.db.QueryRowContext(ctx, findEndOfDayReconciliationHeaderSQL, strings.TrimSpace(id))
	reconciliation, err := scanPostgresEndOfDayReconciliation(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.EndOfDayReconciliation{}, ErrEndOfDayReconciliationNotFound
	}
	if err != nil {
		return domain.EndOfDayReconciliation{}, err
	}

	return reconciliation, nil
}

func (s PostgresEndOfDayReconciliationStore) Save(
	ctx context.Context,
	reconciliation domain.EndOfDayReconciliation,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if strings.TrimSpace(reconciliation.ID) == "" {
		return errors.New("end-of-day reconciliation id is required")
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin end-of-day reconciliation transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPersistedEndOfDayReconciliation(ctx, tx, reconciliation.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, "")
		if err != nil {
			return err
		}
	}
	warehouseID, err := resolveEndOfDayReconciliationWarehouseID(
		ctx,
		tx,
		orgID,
		reconciliation.WarehouseID,
		reconciliation.WarehouseCode,
	)
	if err != nil {
		return err
	}

	if strings.TrimSpace(persistedID) == "" {
		persistedID, err = insertEndOfDayReconciliation(ctx, tx, orgID, warehouseID, reconciliation)
	} else {
		persistedID, err = updateEndOfDayReconciliation(ctx, tx, persistedID, orgID, warehouseID, reconciliation)
	}
	if err != nil {
		return err
	}
	if err := replaceEndOfDayReconciliationChecklist(ctx, tx, orgID, persistedID, reconciliation.Checklist); err != nil {
		return err
	}
	if err := replaceEndOfDayReconciliationLines(ctx, tx, orgID, persistedID, reconciliation.Lines); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit end-of-day reconciliation transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresEndOfDayReconciliationStore) resolveOrgID(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectEndOfDayReconciliationOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve end-of-day reconciliation org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("end-of-day reconciliation org %q cannot be resolved", orgRef)
}

func scanPostgresEndOfDayReconciliation(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	row interface{ Scan(dest ...any) error },
) (domain.EndOfDayReconciliation, error) {
	var (
		persistedID string
		status      string
		closedAt    sql.NullTime
		closedBy    string
		record      domain.EndOfDayReconciliation
	)
	if err := row.Scan(
		&persistedID,
		&record.ID,
		&record.WarehouseID,
		&record.WarehouseCode,
		&record.Date,
		&record.ShiftCode,
		&status,
		&record.Owner,
		&record.Operations.OrderCount,
		&record.Operations.HandoverOrderCount,
		&record.Operations.ReturnOrderCount,
		&record.Operations.StockMovementCount,
		&record.Operations.StockCountSessionCount,
		&record.Operations.PendingIssueCount,
		&closedAt,
		&closedBy,
		&record.ExceptionNote,
	); err != nil {
		return domain.EndOfDayReconciliation{}, err
	}
	record.Status = domain.NormalizeReconciliationStatus(domain.EndOfDayReconciliationStatus(status))
	if record.Status == "" {
		record.Status = domain.ReconciliationStatusOpen
	}
	record.ShiftCode = strings.ToLower(strings.TrimSpace(record.ShiftCode))
	if closedAt.Valid {
		record.ClosedAt = closedAt.Time.UTC()
	}
	record.ClosedBy = strings.TrimSpace(closedBy)

	checklist, err := listPostgresEndOfDayReconciliationChecklist(ctx, queryer, persistedID)
	if err != nil {
		return domain.EndOfDayReconciliation{}, err
	}
	lines, err := listPostgresEndOfDayReconciliationLines(ctx, queryer, persistedID)
	if err != nil {
		return domain.EndOfDayReconciliation{}, err
	}
	record.Checklist = checklist
	record.Lines = lines

	return record, nil
}

func listPostgresEndOfDayReconciliationChecklist(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	closingID string,
) ([]domain.ReconciliationChecklistItem, error) {
	rows, err := queryer.QueryContext(ctx, selectEndOfDayReconciliationChecklistSQL, closingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ReconciliationChecklistItem, 0)
	for rows.Next() {
		var item domain.ReconciliationChecklistItem
		if err := rows.Scan(
			&item.Key,
			&item.Label,
			&item.Complete,
			&item.Blocking,
			&item.Note,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func listPostgresEndOfDayReconciliationLines(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	closingID string,
) ([]domain.ReconciliationLine, error) {
	rows, err := queryer.QueryContext(ctx, selectEndOfDayReconciliationLinesSQL, closingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.ReconciliationLine, 0)
	for rows.Next() {
		var (
			line       domain.ReconciliationLine
			systemQty  string
			countedQty string
		)
		if err := rows.Scan(
			&line.ID,
			&line.SKU,
			&line.BatchNo,
			&line.BinCode,
			&systemQty,
			&countedQty,
			&line.Reason,
			&line.Owner,
		); err != nil {
			return nil, err
		}
		parsedSystemQty, err := parseEndOfDayIntegerQuantity(systemQty)
		if err != nil {
			return nil, err
		}
		parsedCountedQty, err := parseEndOfDayIntegerQuantity(countedQty)
		if err != nil {
			return nil, err
		}
		line.SystemQuantity = parsedSystemQty
		line.CountedQuantity = parsedCountedQty
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func findPersistedEndOfDayReconciliation(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findPersistedEndOfDayReconciliationSQL, strings.TrimSpace(id)).
		Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find end-of-day reconciliation %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func resolveEndOfDayReconciliationWarehouseID(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	orgID string,
	warehouseRef string,
	warehouseCode string,
) (string, error) {
	warehouseRef = strings.TrimSpace(warehouseRef)
	warehouseCode = strings.TrimSpace(warehouseCode)
	if isUUIDText(warehouseRef) {
		return warehouseRef, nil
	}

	var warehouseID string
	err := queryer.QueryRowContext(
		ctx,
		selectEndOfDayReconciliationWarehouseSQL,
		orgID,
		warehouseRef,
		warehouseCode,
	).Scan(&warehouseID)
	if err == nil && isUUIDText(warehouseID) {
		return warehouseID, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("resolve end-of-day reconciliation warehouse %q: %w", warehouseRef, err)
	}

	return "", fmt.Errorf("end-of-day reconciliation warehouse %q cannot be resolved", warehouseRef)
}

func insertEndOfDayReconciliation(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	orgID string,
	warehouseID string,
	reconciliation domain.EndOfDayReconciliation,
) (string, error) {
	args := endOfDayReconciliationHeaderArgs(orgID, warehouseID, reconciliation)
	var persistedID string
	err := queryer.QueryRowContext(ctx, insertEndOfDayReconciliationSQL, args...).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("insert end-of-day reconciliation %q: %w", reconciliation.ID, err)
	}

	return persistedID, nil
}

func updateEndOfDayReconciliation(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	persistedID string,
	orgID string,
	warehouseID string,
	reconciliation domain.EndOfDayReconciliation,
) (string, error) {
	args := append([]any{persistedID}, endOfDayReconciliationUpdateArgs(orgID, warehouseID, reconciliation)...)
	var updatedID string
	err := queryer.QueryRowContext(ctx, updateEndOfDayReconciliationSQL, args...).Scan(&updatedID)
	if err != nil {
		return "", fmt.Errorf("update end-of-day reconciliation %q: %w", reconciliation.ID, err)
	}

	return updatedID, nil
}

func replaceEndOfDayReconciliationChecklist(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	orgID string,
	closingID string,
	items []domain.ReconciliationChecklistItem,
) error {
	if _, err := queryer.ExecContext(ctx, deleteEndOfDayReconciliationChecklistSQL, closingID); err != nil {
		return fmt.Errorf("delete end-of-day reconciliation checklist: %w", err)
	}
	now := time.Now().UTC()
	for _, item := range items {
		if _, err := queryer.ExecContext(
			ctx,
			insertEndOfDayReconciliationChecklistSQL,
			orgID,
			closingID,
			strings.TrimSpace(item.Key),
			strings.TrimSpace(item.Label),
			item.Complete,
			item.Blocking,
			strings.TrimSpace(item.Note),
			now,
			now,
		); err != nil {
			return fmt.Errorf("insert end-of-day reconciliation checklist %q: %w", item.Key, err)
		}
	}

	return nil
}

func replaceEndOfDayReconciliationLines(
	ctx context.Context,
	queryer postgresEndOfDayReconciliationQueryer,
	orgID string,
	closingID string,
	lines []domain.ReconciliationLine,
) error {
	if _, err := queryer.ExecContext(ctx, deleteEndOfDayReconciliationLinesSQL, closingID); err != nil {
		return fmt.Errorf("delete end-of-day reconciliation lines: %w", err)
	}
	now := time.Now().UTC()
	for index, line := range lines {
		if _, err := queryer.ExecContext(
			ctx,
			insertEndOfDayReconciliationLineSQL,
			orgID,
			closingID,
			strings.TrimSpace(line.ID),
			index+1,
			strings.TrimSpace(line.SKU),
			strings.TrimSpace(line.BatchNo),
			strings.TrimSpace(line.BinCode),
			line.SystemQuantity,
			line.CountedQuantity,
			strings.TrimSpace(line.Reason),
			strings.TrimSpace(line.Owner),
			now,
			now,
		); err != nil {
			return fmt.Errorf("insert end-of-day reconciliation line %q: %w", line.ID, err)
		}
	}

	return nil
}

func endOfDayReconciliationHeaderArgs(
	orgID string,
	warehouseID string,
	reconciliation domain.EndOfDayReconciliation,
) []any {
	now := time.Now().UTC()
	closedAt := nullableEndOfDayTime(reconciliation.ClosedAt)
	summary := reconciliation.Summary(reconciliation.ExceptionNote)

	return []any{
		nullableUUID(reconciliation.ID),
		orgID,
		strings.TrimSpace(reconciliation.ID),
		strings.TrimSpace(reconciliation.ID),
		orgID,
		warehouseID,
		strings.TrimSpace(reconciliation.WarehouseID),
		strings.TrimSpace(reconciliation.WarehouseCode),
		strings.TrimSpace(reconciliation.Date),
		strings.ToLower(strings.TrimSpace(reconciliation.ShiftCode)),
		string(reconciliation.Status),
		reconciliation.Operations.OrderCount,
		reconciliation.Operations.HandoverOrderCount,
		reconciliation.Operations.ReturnOrderCount,
		reconciliation.Operations.StockMovementCount,
		reconciliation.Operations.StockCountSessionCount,
		reconciliation.Operations.PendingIssueCount,
		summary.VarianceCount,
		strings.TrimSpace(reconciliation.ExceptionNote),
		closedAt,
		nullableUUID(reconciliation.ClosedBy),
		strings.TrimSpace(reconciliation.ClosedBy),
		nullableUUID(reconciliation.Owner),
		strings.TrimSpace(reconciliation.Owner),
		nullableUUID(reconciliation.ClosedBy),
		strings.TrimSpace(reconciliation.ClosedBy),
		strings.TrimSpace(reconciliation.Owner),
		now,
		now,
	}
}

func endOfDayReconciliationUpdateArgs(
	orgID string,
	warehouseID string,
	reconciliation domain.EndOfDayReconciliation,
) []any {
	now := time.Now().UTC()
	closedAt := nullableEndOfDayTime(reconciliation.ClosedAt)
	summary := reconciliation.Summary(reconciliation.ExceptionNote)

	return []any{
		strings.TrimSpace(reconciliation.ID),
		strings.TrimSpace(reconciliation.ID),
		orgID,
		warehouseID,
		strings.TrimSpace(reconciliation.WarehouseID),
		strings.TrimSpace(reconciliation.WarehouseCode),
		strings.TrimSpace(reconciliation.Date),
		strings.ToLower(strings.TrimSpace(reconciliation.ShiftCode)),
		string(reconciliation.Status),
		reconciliation.Operations.OrderCount,
		reconciliation.Operations.HandoverOrderCount,
		reconciliation.Operations.ReturnOrderCount,
		reconciliation.Operations.StockMovementCount,
		reconciliation.Operations.StockCountSessionCount,
		reconciliation.Operations.PendingIssueCount,
		summary.VarianceCount,
		strings.TrimSpace(reconciliation.ExceptionNote),
		closedAt,
		nullableUUID(reconciliation.ClosedBy),
		strings.TrimSpace(reconciliation.ClosedBy),
		nullableUUID(reconciliation.ClosedBy),
		strings.TrimSpace(reconciliation.ClosedBy),
		strings.TrimSpace(reconciliation.Owner),
		now,
	}
}

func endOfDayReconciliationMatchesFilter(
	reconciliation domain.EndOfDayReconciliation,
	filter domain.EndOfDayReconciliationFilter,
) bool {
	if strings.TrimSpace(filter.WarehouseID) != "" && reconciliation.WarehouseID != strings.TrimSpace(filter.WarehouseID) {
		return false
	}
	if strings.TrimSpace(filter.Date) != "" && reconciliation.Date != strings.TrimSpace(filter.Date) {
		return false
	}
	if strings.TrimSpace(filter.ShiftCode) != "" &&
		reconciliation.ShiftCode != strings.ToLower(strings.TrimSpace(filter.ShiftCode)) {
		return false
	}
	if status := domain.NormalizeReconciliationStatus(filter.Status); status != "" && reconciliation.Status != status {
		return false
	}

	return true
}

func parseEndOfDayIntegerQuantity(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	if strings.Contains(value, ".") {
		parts := strings.SplitN(value, ".", 2)
		if strings.TrimRight(parts[1], "0") != "" {
			return 0, fmt.Errorf("end-of-day reconciliation quantity %q is not an integer", value)
		}
		value = parts[0]
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse end-of-day reconciliation quantity %q: %w", value, err)
	}

	return parsed, nil
}

func nullableEndOfDayTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}
