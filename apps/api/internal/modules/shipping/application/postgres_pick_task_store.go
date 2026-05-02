package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type PostgresPickTaskStoreConfig struct {
	DefaultOrgID string
}

type PostgresPickTaskStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresPickTaskQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresPickTaskStore(db *sql.DB, cfg PostgresPickTaskStoreConfig) PostgresPickTaskStore {
	return PostgresPickTaskStore{db: db, defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID)}
}

const selectPickTaskOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectPickTaskWarehouseSQL = `
SELECT id::text, code
FROM mdm.warehouses
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(code) = lower($2) OR lower(code) = lower($3))
LIMIT 1`

const selectPickTaskSalesOrderSQL = `
SELECT id::text, order_no
FROM sales.sales_orders
WHERE org_id = $1::uuid
  AND (
    id::text = $2
    OR lower(order_ref) = lower($2)
    OR lower(order_no) = lower($2)
    OR lower(order_no) = lower($3)
  )
LIMIT 1`

const selectPickTaskSalesOrderLineSQL = `
SELECT id::text
FROM sales.sales_order_lines
WHERE id::text = $1 OR lower(line_ref) = lower($1)
LIMIT 1`

const selectPickTaskStockReservationSQL = `
SELECT id::text
FROM inventory.stock_reservations
WHERE id::text = $1 OR lower(reservation_ref) = lower($1)
LIMIT 1`

const selectPickTaskItemSQL = `
SELECT id::text, sku
FROM mdm.items
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(sku) = lower($2) OR lower(sku) = lower($3))
LIMIT 1`

const selectPickTaskBatchSQL = `
SELECT id::text, batch_no
FROM inventory.batches
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(batch_ref) = lower($2) OR lower(batch_no) = lower($2) OR lower(batch_no) = lower($3))
LIMIT 1`

const selectPickTaskBinSQL = `
SELECT id::text, code
FROM mdm.warehouse_bins
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(code) = lower($2) OR lower(code) = lower($3))
LIMIT 1`

const selectPickTaskHeadersBaseSQL = `
SELECT
  pick_task.id::text,
  COALESCE(pick_task.pick_ref, pick_task.pick_task_no, pick_task.id::text),
  COALESCE(pick_task.org_ref, pick_task.org_id::text),
  pick_task.pick_task_no,
  COALESCE(pick_task.sales_order_ref, pick_task.sales_order_id::text),
  COALESCE(pick_task.order_no, sales_order.order_no, ''),
  COALESCE(pick_task.warehouse_ref, pick_task.warehouse_id::text),
  COALESCE(pick_task.warehouse_code, warehouse.code, ''),
  pick_task.status,
  COALESCE(pick_task.assigned_to_ref, pick_task.assigned_to::text, ''),
  pick_task.assigned_at,
  pick_task.started_at,
  COALESCE(pick_task.started_by_ref, pick_task.started_by::text, ''),
  pick_task.completed_at,
  COALESCE(pick_task.completed_by_ref, pick_task.completed_by::text, ''),
  pick_task.created_at,
  pick_task.updated_at
FROM shipping.pick_tasks AS pick_task
LEFT JOIN sales.sales_orders AS sales_order ON sales_order.id = pick_task.sales_order_id
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = pick_task.warehouse_id`

const selectPickTaskHeadersSQL = selectPickTaskHeadersBaseSQL + `
ORDER BY pick_task.created_at, pick_task.pick_task_no`

const findPickTaskHeaderSQL = selectPickTaskHeadersBaseSQL + `
WHERE lower(COALESCE(pick_task.pick_ref, pick_task.pick_task_no, pick_task.id::text)) = lower($1)
   OR pick_task.id::text = $1
LIMIT 1`

const findPickTaskBySalesOrderSQL = selectPickTaskHeadersBaseSQL + `
WHERE lower(COALESCE(pick_task.sales_order_ref, pick_task.sales_order_id::text)) = lower($1)
   OR pick_task.sales_order_id::text = $1
LIMIT 1`

const findPickTaskPersistedSQL = `
SELECT id::text, org_id::text
FROM shipping.pick_tasks
WHERE lower(COALESCE(pick_ref, pick_task_no, id::text)) = lower($1)
   OR id::text = $1
LIMIT 1
FOR UPDATE`

const selectPickTaskLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  COALESCE(line.pick_task_ref, line.pick_task_id::text),
  line.line_no,
  COALESCE(line.sales_order_line_ref, line.sales_order_line_id::text),
  COALESCE(line.stock_reservation_ref, line.stock_reservation_id::text),
  COALESCE(line.item_ref, line.item_id::text),
  COALESCE(line.sku_code, item.sku, ''),
  COALESCE(line.batch_ref, line.batch_id::text),
  COALESCE(line.batch_no, batch.batch_no, ''),
  COALESCE(line.warehouse_ref, line.warehouse_id::text),
  COALESCE(line.bin_ref, line.bin_id::text),
  COALESCE(line.bin_code, bin.code, ''),
  line.qty_to_pick::text,
  line.qty_picked::text,
  line.base_uom_code,
  line.status,
  line.picked_at,
  COALESCE(line.picked_by_ref, line.picked_by::text, ''),
  line.created_at,
  line.updated_at
FROM shipping.pick_task_lines AS line
LEFT JOIN mdm.items AS item ON item.id = line.item_id
LEFT JOIN inventory.batches AS batch ON batch.id = line.batch_id
LEFT JOIN mdm.warehouse_bins AS bin ON bin.id = line.bin_id
WHERE line.pick_task_id = $1::uuid
ORDER BY line.line_no, line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertPickTaskSQL = `
INSERT INTO shipping.pick_tasks (
  id,
  org_id,
  pick_ref,
  org_ref,
  pick_task_no,
  sales_order_id,
  sales_order_ref,
  order_no,
  warehouse_id,
  warehouse_ref,
  warehouse_code,
  status,
  assigned_to,
  assigned_to_ref,
  assigned_at,
  started_at,
  started_by,
  started_by_ref,
  completed_at,
  completed_by,
  completed_by_ref,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
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
  $27
)
ON CONFLICT (org_id, pick_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  pick_task_no = EXCLUDED.pick_task_no,
  sales_order_id = EXCLUDED.sales_order_id,
  sales_order_ref = EXCLUDED.sales_order_ref,
  order_no = EXCLUDED.order_no,
  warehouse_id = EXCLUDED.warehouse_id,
  warehouse_ref = EXCLUDED.warehouse_ref,
  warehouse_code = EXCLUDED.warehouse_code,
  status = EXCLUDED.status,
  assigned_to = EXCLUDED.assigned_to,
  assigned_to_ref = EXCLUDED.assigned_to_ref,
  assigned_at = EXCLUDED.assigned_at,
  started_at = EXCLUDED.started_at,
  started_by = EXCLUDED.started_by,
  started_by_ref = EXCLUDED.started_by_ref,
  completed_at = EXCLUDED.completed_at,
  completed_by = EXCLUDED.completed_by,
  completed_by_ref = EXCLUDED.completed_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = shipping.pick_tasks.version + 1
RETURNING id::text`

const deletePickTaskLinesSQL = `
DELETE FROM shipping.pick_task_lines
WHERE pick_task_id = $1::uuid`

const insertPickTaskLineSQL = `
INSERT INTO shipping.pick_task_lines (
  id,
  org_id,
  pick_task_id,
  line_ref,
  pick_task_ref,
  line_no,
  sales_order_line_id,
  sales_order_line_ref,
  stock_reservation_id,
  stock_reservation_ref,
  item_id,
  item_ref,
  sku_code,
  batch_id,
  batch_ref,
  batch_no,
  warehouse_id,
  warehouse_ref,
  bin_id,
  bin_ref,
  bin_code,
  base_uom_code,
  qty_to_pick,
  qty_picked,
  status,
  picked_at,
  picked_by,
  picked_by_ref,
  created_at,
  created_by,
  created_by_ref,
  updated_at,
  updated_by,
  updated_by_ref
) VALUES (
  COALESCE($1::uuid, gen_random_uuid()),
  $2::uuid,
  $3::uuid,
  $4,
  $5,
  $6,
  $7::uuid,
  $8,
  $9::uuid,
  $10,
  $11::uuid,
  $12,
  $13,
  $14::uuid,
  $15,
  $16,
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
  $27::uuid,
  $28,
  $29,
  $30::uuid,
  $31,
  $32,
  $33::uuid,
  $34
)`

func (s PostgresPickTaskStore) ListPickTasks(ctx context.Context) ([]domain.PickTask, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectPickTaskHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.PickTask, 0)
	for rows.Next() {
		task, err := scanPostgresPickTask(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortPickTasks(tasks)

	return tasks, nil
}

func (s PostgresPickTaskStore) GetPickTask(ctx context.Context, id string) (domain.PickTask, error) {
	if s.db == nil {
		return domain.PickTask{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findPickTaskHeaderSQL, strings.TrimSpace(id))
	task, err := scanPostgresPickTask(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PickTask{}, ErrPickTaskNotFound
	}
	if err != nil {
		return domain.PickTask{}, err
	}

	return task, nil
}

func (s PostgresPickTaskStore) GetPickTaskBySalesOrder(
	ctx context.Context,
	salesOrderID string,
) (domain.PickTask, error) {
	if s.db == nil {
		return domain.PickTask{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findPickTaskBySalesOrderSQL, strings.TrimSpace(salesOrderID))
	task, err := scanPostgresPickTask(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PickTask{}, ErrPickTaskNotFound
	}
	if err != nil {
		return domain.PickTask{}, err
	}

	return task, nil
}

func (s PostgresPickTaskStore) SavePickTask(ctx context.Context, task domain.PickTask) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin pick task transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresPickTask(ctx, tx, task.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, task.OrgID)
		if err != nil {
			return err
		}
	}
	salesOrderID, orderNo, err := resolvePostgresPickTaskSalesOrder(ctx, tx, orgID, task.SalesOrderID, task.OrderNo)
	if err != nil {
		return err
	}
	warehouseID, warehouseCode, err := resolvePostgresPickTaskWarehouse(ctx, tx, orgID, task.WarehouseID, task.WarehouseCode)
	if err != nil {
		return err
	}
	persistedID, err = upsertPostgresPickTask(ctx, tx, persistedID, orgID, salesOrderID, orderNo, warehouseID, warehouseCode, task)
	if err != nil {
		return err
	}
	if err := replacePostgresPickTaskLines(ctx, tx, orgID, persistedID, task); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pick task transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresPickTaskStore) resolveOrgID(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresPickTaskUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectPickTaskOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresPickTaskUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve pick task org %q: %w", orgRef, err)
		}
	}
	if isPostgresPickTaskUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("pick task org %q cannot be resolved", orgRef)
}

func scanPostgresPickTask(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	row interface{ Scan(dest ...any) error },
) (domain.PickTask, error) {
	var (
		persistedID string
		status      string
		assignedAt  sql.NullTime
		startedAt   sql.NullTime
		completedAt sql.NullTime
		task        domain.PickTask
	)
	if err := row.Scan(
		&persistedID,
		&task.ID,
		&task.OrgID,
		&task.PickTaskNo,
		&task.SalesOrderID,
		&task.OrderNo,
		&task.WarehouseID,
		&task.WarehouseCode,
		&status,
		&task.AssignedTo,
		&assignedAt,
		&startedAt,
		&task.StartedBy,
		&completedAt,
		&task.CompletedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return domain.PickTask{}, err
	}
	task.Status = domain.NormalizePickTaskStatus(domain.PickTaskStatus(status))
	if assignedAt.Valid {
		task.AssignedAt = assignedAt.Time.UTC()
	}
	if startedAt.Valid {
		task.StartedAt = startedAt.Time.UTC()
	}
	if completedAt.Valid {
		task.CompletedAt = completedAt.Time.UTC()
	}
	lines, err := listPostgresPickTaskLines(ctx, queryer, persistedID)
	if err != nil {
		return domain.PickTask{}, err
	}
	task.Lines = lines
	if err := task.Validate(); err != nil {
		return domain.PickTask{}, err
	}

	return task, nil
}

func listPostgresPickTaskLines(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	persistedID string,
) ([]domain.PickTaskLine, error) {
	rows, err := queryer.QueryContext(ctx, selectPickTaskLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.PickTaskLine, 0)
	for rows.Next() {
		line, err := scanPostgresPickTaskLine(rows)
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

func scanPostgresPickTaskLine(row interface{ Scan(dest ...any) error }) (domain.PickTaskLine, error) {
	var (
		line          domain.PickTaskLine
		qtyToPick     string
		qtyPicked     string
		baseUOMCode   string
		status        string
		pickedAt      sql.NullTime
		parsedUOMCode decimal.UOMCode
		err           error
	)
	if err := row.Scan(
		&line.ID,
		&line.PickTaskID,
		&line.LineNo,
		&line.SalesOrderLineID,
		&line.StockReservationID,
		&line.ItemID,
		&line.SKUCode,
		&line.BatchID,
		&line.BatchNo,
		&line.WarehouseID,
		&line.BinID,
		&line.BinCode,
		&qtyToPick,
		&qtyPicked,
		&baseUOMCode,
		&status,
		&pickedAt,
		&line.PickedBy,
		&line.CreatedAt,
		&line.UpdatedAt,
	); err != nil {
		return domain.PickTaskLine{}, err
	}
	line.QtyToPick, err = decimal.ParseQuantity(qtyToPick)
	if err != nil {
		return domain.PickTaskLine{}, err
	}
	line.QtyPicked, err = decimal.ParseQuantity(qtyPicked)
	if err != nil {
		return domain.PickTaskLine{}, err
	}
	parsedUOMCode, err = decimal.NormalizeUOMCode(baseUOMCode)
	if err != nil {
		return domain.PickTaskLine{}, err
	}
	line.BaseUOMCode = parsedUOMCode
	line.Status = domain.NormalizePickTaskLineStatus(domain.PickTaskLineStatus(status))
	if pickedAt.Valid {
		line.PickedAt = pickedAt.Time.UTC()
	}
	if err := line.Validate(); err != nil {
		return domain.PickTaskLine{}, err
	}

	return line, nil
}

func findPostgresPickTask(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findPickTaskPersistedSQL, strings.TrimSpace(id)).Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find pick task %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func resolvePostgresPickTaskWarehouse(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	orgID string,
	warehouseRef string,
	warehouseCode string,
) (string, string, error) {
	warehouseRef = strings.TrimSpace(warehouseRef)
	warehouseCode = strings.TrimSpace(warehouseCode)
	if isPostgresPickTaskUUIDText(warehouseRef) && warehouseCode != "" {
		return warehouseRef, warehouseCode, nil
	}
	var id string
	var code string
	err := queryer.QueryRowContext(ctx, selectPickTaskWarehouseSQL, orgID, warehouseRef, warehouseCode).Scan(&id, &code)
	if err != nil {
		return "", "", fmt.Errorf("resolve pick task warehouse %q: %w", warehouseRef, err)
	}

	return id, code, nil
}

func resolvePostgresPickTaskSalesOrder(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	orgID string,
	salesOrderRef string,
	orderNo string,
) (string, string, error) {
	salesOrderRef = strings.TrimSpace(salesOrderRef)
	orderNo = strings.TrimSpace(orderNo)
	if isPostgresPickTaskUUIDText(salesOrderRef) && orderNo != "" {
		return salesOrderRef, orderNo, nil
	}
	var id string
	var resolvedOrderNo string
	err := queryer.QueryRowContext(ctx, selectPickTaskSalesOrderSQL, orgID, salesOrderRef, orderNo).Scan(&id, &resolvedOrderNo)
	if err != nil {
		return "", "", fmt.Errorf("resolve pick task sales order %q: %w", salesOrderRef, err)
	}

	return id, resolvedOrderNo, nil
}

func upsertPostgresPickTask(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	persistedID string,
	orgID string,
	salesOrderID string,
	orderNo string,
	warehouseID string,
	warehouseCode string,
	task domain.PickTask,
) (string, error) {
	var savedID string
	err := queryer.QueryRowContext(
		ctx,
		upsertPickTaskSQL,
		nullablePostgresPickTaskUUID(firstNonBlankPostgresPickTask(persistedID, task.ID)),
		orgID,
		nullablePostgresPickTaskText(task.ID),
		nullablePostgresPickTaskText(task.OrgID),
		task.PickTaskNo,
		salesOrderID,
		nullablePostgresPickTaskText(task.SalesOrderID),
		nullablePostgresPickTaskText(firstNonBlankPostgresPickTask(task.OrderNo, orderNo)),
		warehouseID,
		nullablePostgresPickTaskText(task.WarehouseID),
		nullablePostgresPickTaskText(firstNonBlankPostgresPickTask(task.WarehouseCode, warehouseCode)),
		string(task.Status),
		nullablePostgresPickTaskUUID(task.AssignedTo),
		nullablePostgresPickTaskText(task.AssignedTo),
		nullablePostgresPickTaskTime(task.AssignedAt),
		nullablePostgresPickTaskTime(task.StartedAt),
		nullablePostgresPickTaskUUID(task.StartedBy),
		nullablePostgresPickTaskText(task.StartedBy),
		nullablePostgresPickTaskTime(task.CompletedAt),
		nullablePostgresPickTaskUUID(task.CompletedBy),
		nullablePostgresPickTaskText(task.CompletedBy),
		postgresPickTaskTime(task.CreatedAt),
		nullablePostgresPickTaskUUID(task.AssignedTo),
		nullablePostgresPickTaskText(task.AssignedTo),
		postgresPickTaskTime(task.UpdatedAt),
		nullablePostgresPickTaskUUID(firstNonBlankPostgresPickTask(task.CompletedBy, task.StartedBy, task.AssignedTo)),
		nullablePostgresPickTaskText(firstNonBlankPostgresPickTask(task.CompletedBy, task.StartedBy, task.AssignedTo)),
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert pick task %q: %w", task.ID, err)
	}

	return savedID, nil
}

func replacePostgresPickTaskLines(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	orgID string,
	persistedID string,
	task domain.PickTask,
) error {
	if _, err := queryer.ExecContext(ctx, deletePickTaskLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete pick task lines: %w", err)
	}
	for _, line := range task.Lines {
		salesOrderLineID, err := resolvePostgresPickTaskID(ctx, queryer, selectPickTaskSalesOrderLineSQL, line.SalesOrderLineID, "sales order line")
		if err != nil {
			return err
		}
		reservationID, err := resolvePostgresPickTaskID(ctx, queryer, selectPickTaskStockReservationSQL, line.StockReservationID, "stock reservation")
		if err != nil {
			return err
		}
		itemID, skuCode, err := resolvePostgresPickTaskItem(ctx, queryer, orgID, line.ItemID, line.SKUCode)
		if err != nil {
			return err
		}
		batchID, batchNo, err := resolvePostgresPickTaskBatch(ctx, queryer, orgID, line.BatchID, line.BatchNo)
		if err != nil {
			return err
		}
		warehouseID, _, err := resolvePostgresPickTaskWarehouse(ctx, queryer, orgID, line.WarehouseID, task.WarehouseCode)
		if err != nil {
			return err
		}
		binID, binCode, err := resolvePostgresPickTaskBin(ctx, queryer, orgID, line.BinID, line.BinCode)
		if err != nil {
			return err
		}
		if _, err := queryer.ExecContext(
			ctx,
			insertPickTaskLineSQL,
			nullablePostgresPickTaskUUID(line.ID),
			orgID,
			persistedID,
			nullablePostgresPickTaskText(line.ID),
			nullablePostgresPickTaskText(task.ID),
			line.LineNo,
			salesOrderLineID,
			nullablePostgresPickTaskText(line.SalesOrderLineID),
			reservationID,
			nullablePostgresPickTaskText(line.StockReservationID),
			itemID,
			nullablePostgresPickTaskText(line.ItemID),
			firstNonBlankPostgresPickTask(line.SKUCode, skuCode),
			batchID,
			nullablePostgresPickTaskText(line.BatchID),
			nullablePostgresPickTaskText(firstNonBlankPostgresPickTask(line.BatchNo, batchNo)),
			warehouseID,
			nullablePostgresPickTaskText(line.WarehouseID),
			binID,
			nullablePostgresPickTaskText(line.BinID),
			nullablePostgresPickTaskText(firstNonBlankPostgresPickTask(line.BinCode, binCode)),
			line.BaseUOMCode.String(),
			line.QtyToPick.String(),
			line.QtyPicked.String(),
			string(line.Status),
			nullablePostgresPickTaskTime(line.PickedAt),
			nullablePostgresPickTaskUUID(line.PickedBy),
			nullablePostgresPickTaskText(line.PickedBy),
			postgresPickTaskTime(line.CreatedAt),
			nullablePostgresPickTaskUUID(line.PickedBy),
			nullablePostgresPickTaskText(line.PickedBy),
			postgresPickTaskTime(line.UpdatedAt),
			nullablePostgresPickTaskUUID(line.PickedBy),
			nullablePostgresPickTaskText(line.PickedBy),
		); err != nil {
			return fmt.Errorf("insert pick task line %q: %w", line.ID, err)
		}
	}

	return nil
}

func resolvePostgresPickTaskID(
	ctx context.Context,
	queryer postgresPickTaskQueryer,
	query string,
	ref string,
	label string,
) (string, error) {
	var id string
	err := queryer.QueryRowContext(ctx, query, strings.TrimSpace(ref)).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("resolve pick task %s %q: %w", label, ref, err)
	}

	return id, nil
}

func resolvePostgresPickTaskItem(ctx context.Context, queryer postgresPickTaskQueryer, orgID string, itemRef string, sku string) (string, string, error) {
	var id string
	var resolvedSKU string
	err := queryer.QueryRowContext(ctx, selectPickTaskItemSQL, orgID, strings.TrimSpace(itemRef), strings.TrimSpace(sku)).
		Scan(&id, &resolvedSKU)
	if err != nil {
		return "", "", fmt.Errorf("resolve pick task item %q: %w", itemRef, err)
	}

	return id, resolvedSKU, nil
}

func resolvePostgresPickTaskBatch(ctx context.Context, queryer postgresPickTaskQueryer, orgID string, batchRef string, batchNo string) (string, string, error) {
	var id string
	var resolvedBatchNo string
	err := queryer.QueryRowContext(ctx, selectPickTaskBatchSQL, orgID, strings.TrimSpace(batchRef), strings.TrimSpace(batchNo)).
		Scan(&id, &resolvedBatchNo)
	if err != nil {
		return "", "", fmt.Errorf("resolve pick task batch %q: %w", batchRef, err)
	}

	return id, resolvedBatchNo, nil
}

func resolvePostgresPickTaskBin(ctx context.Context, queryer postgresPickTaskQueryer, orgID string, binRef string, binCode string) (string, string, error) {
	var id string
	var resolvedBinCode string
	err := queryer.QueryRowContext(ctx, selectPickTaskBinSQL, orgID, strings.TrimSpace(binRef), strings.TrimSpace(binCode)).
		Scan(&id, &resolvedBinCode)
	if err != nil {
		return "", "", fmt.Errorf("resolve pick task bin %q: %w", binRef, err)
	}

	return id, resolvedBinCode, nil
}

func nullablePostgresPickTaskText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresPickTaskUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresPickTaskUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresPickTaskTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresPickTaskTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func firstNonBlankPostgresPickTask(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresPickTaskUUIDText(value string) bool {
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
			if !isPostgresPickTaskHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresPickTaskHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ PickTaskStore = PostgresPickTaskStore{}
