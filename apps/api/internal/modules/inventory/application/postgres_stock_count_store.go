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

type PostgresStockCountStoreConfig struct {
	DefaultOrgID string
}

type PostgresStockCountStore struct {
	db           *sql.DB
	defaultOrgID string
}

func NewPostgresStockCountStore(db *sql.DB, cfg PostgresStockCountStoreConfig) PostgresStockCountStore {
	return PostgresStockCountStore{
		db:           db,
		defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID),
	}
}

const selectStockCountOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const listStockCountsSQL = `
SELECT
  id::text,
  COALESCE(count_ref, id::text),
  COALESCE(org_ref, org_id::text),
  count_no,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(warehouse_code, ''),
  scope,
  status,
  COALESCE(created_by_ref, created_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  created_at,
  updated_at
FROM inventory.stock_count_sessions
ORDER BY created_at DESC, count_no DESC`

const findStockCountPredicateSQL = `
SELECT
  id::text,
  COALESCE(count_ref, id::text),
  COALESCE(org_ref, org_id::text),
  count_no,
  COALESCE(warehouse_ref, warehouse_id::text),
  COALESCE(warehouse_code, ''),
  scope,
  status,
  COALESCE(created_by_ref, created_by::text, ''),
  submitted_at,
  COALESCE(submitted_by_ref, submitted_by::text, ''),
  created_at,
  updated_at
FROM inventory.stock_count_sessions
WHERE count_ref = $1 OR id::text = $1
LIMIT 1`

const selectStockCountLinesSQL = `
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
  delta_qty::text,
  base_uom_code,
  counted,
  COALESCE(note, '')
FROM inventory.stock_count_session_lines
WHERE session_id = $1::uuid
ORDER BY line_no, created_at, line_ref`

const upsertStockCountSQL = `
INSERT INTO inventory.stock_count_sessions (
  id,
  org_id,
  count_ref,
  count_no,
  org_ref,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  count_date,
  scope,
  status,
  created_by,
  created_by_ref,
  submitted_at,
  submitted_by,
  submitted_by_ref,
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
  $12::uuid,
  $13,
  $14,
  $15::uuid,
  $16,
  $17,
  $18
)
ON CONFLICT ON CONSTRAINT uq_stock_count_sessions_org_ref
DO UPDATE SET
  count_no = EXCLUDED.count_no,
  org_ref = EXCLUDED.org_ref,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  count_date = EXCLUDED.count_date,
  scope = EXCLUDED.scope,
  status = EXCLUDED.status,
  created_by = EXCLUDED.created_by,
  created_by_ref = EXCLUDED.created_by_ref,
  submitted_at = EXCLUDED.submitted_at,
  submitted_by = EXCLUDED.submitted_by,
  submitted_by_ref = EXCLUDED.submitted_by_ref,
  updated_at = EXCLUDED.updated_at
RETURNING id::text`

const deleteStockCountLinesSQL = `
DELETE FROM inventory.stock_count_session_lines
WHERE session_id = $1::uuid`

const insertStockCountLineSQL = `
INSERT INTO inventory.stock_count_session_lines (
  id,
  org_id,
  session_id,
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
  counted,
  note,
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
  $21,
  $22
)`

func (s PostgresStockCountStore) ListStockCounts(ctx context.Context) ([]domain.StockCountSession, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, listStockCountsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions, err := s.scanStockCountSessions(ctx, rows)
	if err != nil {
		return nil, err
	}
	domain.SortStockCountSessions(sessions)

	return sessions, nil
}

func (s PostgresStockCountStore) FindStockCountByID(
	ctx context.Context,
	id string,
) (domain.StockCountSession, error) {
	if s.db == nil {
		return domain.StockCountSession{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findStockCountPredicateSQL, strings.TrimSpace(id))
	session, err := s.scanStockCountSession(ctx, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StockCountSession{}, ErrStockCountNotFound
	}
	if err != nil {
		return domain.StockCountSession{}, err
	}

	return session, nil
}

func (s PostgresStockCountStore) SaveStockCount(
	ctx context.Context,
	session domain.StockCountSession,
) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := session.Validate(); err != nil {
		return err
	}
	orgID, err := s.resolveOrgID(ctx, session.OrgID)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin stock count transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, err := upsertStockCount(ctx, tx, orgID, session)
	if err != nil {
		return err
	}
	if err := replaceStockCountLines(ctx, tx, orgID, persistedID, session); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stock count transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresStockCountStore) resolveOrgID(ctx context.Context, orgRef string) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := s.db.QueryRowContext(ctx, selectStockCountOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve stock count org %q: %w", orgRef, err)
		}
	}
	if isUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("stock count org %q cannot be resolved", orgRef)
}

func (s PostgresStockCountStore) scanStockCountSessions(
	ctx context.Context,
	rows *sql.Rows,
) ([]domain.StockCountSession, error) {
	sessions := make([]domain.StockCountSession, 0)
	for rows.Next() {
		session, err := s.scanStockCountSession(ctx, rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (s PostgresStockCountStore) scanStockCountSession(
	ctx context.Context,
	row interface{ Scan(dest ...any) error },
) (domain.StockCountSession, error) {
	var (
		persistedID   string
		session       domain.StockCountSession
		submittedAt   sql.NullTime
		submittedBy   string
		status        string
		createdBy     string
		warehouseID   string
		warehouseCode string
		orgID         string
	)
	if err := row.Scan(
		&persistedID,
		&session.ID,
		&orgID,
		&session.CountNo,
		&warehouseID,
		&warehouseCode,
		&session.Scope,
		&status,
		&createdBy,
		&submittedAt,
		&submittedBy,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		return domain.StockCountSession{}, err
	}
	session.OrgID = orgID
	session.WarehouseID = warehouseID
	session.WarehouseCode = strings.ToUpper(strings.TrimSpace(warehouseCode))
	session.Status = domain.StockCountStatus(status)
	session.CreatedBy = createdBy
	if submittedAt.Valid {
		session.SubmittedAt = submittedAt.Time.UTC()
		session.SubmittedBy = submittedBy
	}

	lines, err := s.listStockCountLines(ctx, persistedID)
	if err != nil {
		return domain.StockCountSession{}, err
	}
	session.Lines = lines
	if err := session.Validate(); err != nil {
		return domain.StockCountSession{}, err
	}

	return session, nil
}

func (s PostgresStockCountStore) listStockCountLines(
	ctx context.Context,
	sessionID string,
) ([]domain.StockCountLine, error) {
	rows, err := s.db.QueryContext(ctx, selectStockCountLinesSQL, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.StockCountLine, 0)
	for rows.Next() {
		line, err := scanStockCountLine(rows)
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

func scanStockCountLine(row interface{ Scan(dest ...any) error }) (domain.StockCountLine, error) {
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
		deltaQty    string
		baseUOMCode string
		counted     bool
		note        string
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
		&deltaQty,
		&baseUOMCode,
		&counted,
		&note,
	); err != nil {
		return domain.StockCountLine{}, err
	}
	expected, err := decimal.ParseQuantity(expectedQty)
	if err != nil {
		return domain.StockCountLine{}, err
	}
	countedValue, err := decimal.ParseQuantity(countedQty)
	if err != nil {
		return domain.StockCountLine{}, err
	}
	delta, err := decimal.ParseQuantity(deltaQty)
	if err != nil {
		return domain.StockCountLine{}, err
	}
	line, err := domain.NewStockCountLine(domain.NewStockCountLineInput{
		ID:           id,
		ItemID:       itemID,
		SKU:          sku,
		BatchID:      batchID,
		BatchNo:      batchNo,
		LocationID:   locationID,
		LocationCode: location,
		ExpectedQty:  expected,
		BaseUOMCode:  baseUOMCode,
	})
	if err != nil {
		return domain.StockCountLine{}, err
	}
	line.CountedQty = countedValue
	line.DeltaQty = delta
	line.Counted = counted
	line.Note = strings.TrimSpace(note)

	return line, nil
}

func upsertStockCount(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	session domain.StockCountSession,
) (string, error) {
	var persistedID string
	err := tx.QueryRowContext(
		ctx,
		upsertStockCountSQL,
		nullableUUID(session.ID),
		orgID,
		nullableText(session.ID),
		session.CountNo,
		nullableText(session.OrgID),
		nullableUUID(session.WarehouseID),
		nullableText(session.WarehouseID),
		nullableText(session.WarehouseCode),
		session.CreatedAt.UTC().Format("2006-01-02"),
		session.Scope,
		string(session.Status),
		nullableUUID(session.CreatedBy),
		nullableText(session.CreatedBy),
		nullableTime(session.SubmittedAt),
		nullableUUID(session.SubmittedBy),
		nullableText(session.SubmittedBy),
		session.CreatedAt.UTC(),
		session.UpdatedAt.UTC(),
	).Scan(&persistedID)
	if err != nil {
		return "", fmt.Errorf("upsert stock count: %w", err)
	}

	return persistedID, nil
}

func replaceStockCountLines(
	ctx context.Context,
	tx *sql.Tx,
	orgID string,
	persistedID string,
	session domain.StockCountSession,
) error {
	if _, err := tx.ExecContext(ctx, deleteStockCountLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete stock count lines: %w", err)
	}
	for index, line := range session.Lines {
		if _, err := tx.ExecContext(
			ctx,
			insertStockCountLineSQL,
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
			line.Counted,
			nullableText(line.Note),
			session.CreatedAt.UTC(),
			session.UpdatedAt.UTC(),
		); err != nil {
			return fmt.Errorf("insert stock count line: %w", err)
		}
	}

	return nil
}

var _ StockCountStore = PostgresStockCountStore{}
