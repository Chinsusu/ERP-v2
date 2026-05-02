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

type PostgresPackTaskStoreConfig struct {
	DefaultOrgID string
}

type PostgresPackTaskStore struct {
	db           *sql.DB
	defaultOrgID string
}

type postgresPackTaskQueryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func NewPostgresPackTaskStore(db *sql.DB, cfg PostgresPackTaskStoreConfig) PostgresPackTaskStore {
	return PostgresPackTaskStore{db: db, defaultOrgID: strings.TrimSpace(cfg.DefaultOrgID)}
}

const selectPackTaskOrgIDSQL = `
SELECT id::text
FROM core.organizations
WHERE code = $1
LIMIT 1`

const selectPackTaskWarehouseSQL = `
SELECT id::text, code
FROM mdm.warehouses
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(code) = lower($2) OR lower(code) = lower($3))
LIMIT 1`

const selectPackTaskSalesOrderSQL = `
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

const selectPackTaskPickTaskSQL = `
SELECT id::text, pick_task_no
FROM shipping.pick_tasks
WHERE org_id = $1::uuid
  AND (
    id::text = $2
    OR lower(pick_ref) = lower($2)
    OR lower(pick_task_no) = lower($2)
    OR lower(pick_task_no) = lower($3)
  )
LIMIT 1`

const selectPackTaskPickTaskLineSQL = `
SELECT id::text
FROM shipping.pick_task_lines
WHERE id::text = $1 OR lower(line_ref) = lower($1)
LIMIT 1`

const selectPackTaskSalesOrderLineSQL = `
SELECT id::text
FROM sales.sales_order_lines
WHERE id::text = $1 OR lower(line_ref) = lower($1)
LIMIT 1`

const selectPackTaskItemSQL = `
SELECT id::text, sku
FROM mdm.items
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(sku) = lower($2) OR lower(sku) = lower($3))
LIMIT 1`

const selectPackTaskBatchSQL = `
SELECT id::text, batch_no
FROM inventory.batches
WHERE org_id = $1::uuid
  AND (id::text = $2 OR lower(batch_ref) = lower($2) OR lower(batch_no) = lower($2) OR lower(batch_no) = lower($3))
LIMIT 1`

const selectPackTaskHeadersBaseSQL = `
SELECT
  pack_task.id::text,
  COALESCE(pack_task.pack_ref, pack_task.pack_task_no, pack_task.id::text),
  COALESCE(pack_task.org_ref, pack_task.org_id::text),
  pack_task.pack_task_no,
  COALESCE(pack_task.sales_order_ref, pack_task.sales_order_id::text),
  COALESCE(pack_task.order_no, sales_order.order_no, ''),
  COALESCE(pack_task.pick_task_ref, pack_task.pick_task_id::text),
  COALESCE(pack_task.pick_task_no, pick_task.pick_task_no, ''),
  COALESCE(pack_task.warehouse_ref, pack_task.warehouse_id::text),
  COALESCE(pack_task.warehouse_code, warehouse.code, ''),
  pack_task.status,
  COALESCE(pack_task.assigned_to_ref, pack_task.assigned_to::text, ''),
  pack_task.assigned_at,
  pack_task.started_at,
  COALESCE(pack_task.started_by_ref, pack_task.started_by::text, ''),
  pack_task.packed_at,
  COALESCE(pack_task.packed_by_ref, pack_task.packed_by::text, ''),
  pack_task.created_at,
  pack_task.updated_at
FROM shipping.pack_tasks AS pack_task
LEFT JOIN sales.sales_orders AS sales_order ON sales_order.id = pack_task.sales_order_id
LEFT JOIN shipping.pick_tasks AS pick_task ON pick_task.id = pack_task.pick_task_id
LEFT JOIN mdm.warehouses AS warehouse ON warehouse.id = pack_task.warehouse_id`

const selectPackTaskHeadersSQL = selectPackTaskHeadersBaseSQL + `
ORDER BY pack_task.created_at, pack_task.pack_task_no`

const findPackTaskHeaderSQL = selectPackTaskHeadersBaseSQL + `
WHERE lower(COALESCE(pack_task.pack_ref, pack_task.pack_task_no, pack_task.id::text)) = lower($1)
   OR pack_task.id::text = $1
LIMIT 1`

const findPackTaskBySalesOrderSQL = selectPackTaskHeadersBaseSQL + `
WHERE lower(COALESCE(pack_task.sales_order_ref, pack_task.sales_order_id::text)) = lower($1)
   OR pack_task.sales_order_id::text = $1
LIMIT 1`

const findPackTaskByPickTaskSQL = selectPackTaskHeadersBaseSQL + `
WHERE lower(COALESCE(pack_task.pick_task_ref, pack_task.pick_task_id::text)) = lower($1)
   OR pack_task.pick_task_id::text = $1
LIMIT 1`

const findPackTaskPersistedSQL = `
SELECT id::text, org_id::text
FROM shipping.pack_tasks
WHERE lower(COALESCE(pack_ref, pack_task_no, id::text)) = lower($1)
   OR id::text = $1
LIMIT 1
FOR UPDATE`

const selectPackTaskLinesSQL = `
SELECT
  COALESCE(line.line_ref, line.id::text),
  COALESCE(line.pack_task_ref, line.pack_task_id::text),
  line.line_no,
  COALESCE(line.pick_task_line_ref, line.pick_task_line_id::text),
  COALESCE(line.sales_order_line_ref, line.sales_order_line_id::text),
  COALESCE(line.item_ref, line.item_id::text),
  COALESCE(line.sku_code, item.sku, ''),
  COALESCE(line.batch_ref, line.batch_id::text),
  COALESCE(line.batch_no, batch.batch_no, ''),
  COALESCE(line.warehouse_ref, line.warehouse_id::text),
  line.qty_to_pack::text,
  line.qty_packed::text,
  line.base_uom_code,
  line.status,
  line.packed_at,
  COALESCE(line.packed_by_ref, line.packed_by::text, ''),
  line.created_at,
  line.updated_at
FROM shipping.pack_task_lines AS line
LEFT JOIN mdm.items AS item ON item.id = line.item_id
LEFT JOIN inventory.batches AS batch ON batch.id = line.batch_id
WHERE line.pack_task_id = $1::uuid
ORDER BY line.line_no, line.created_at, COALESCE(line.line_ref, line.id::text)`

const upsertPackTaskSQL = `
INSERT INTO shipping.pack_tasks (
  id,
  org_id,
  pack_ref,
  org_ref,
  pack_task_no,
  source_doc_type,
  sales_order_id,
  sales_order_ref,
  order_no,
  pick_task_id,
  pick_task_ref,
  pick_task_no,
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
  packed_at,
  packed_by,
  packed_by_ref,
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
  'sales_order',
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
  $16::uuid,
  $17,
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
  $29::uuid,
  $30
)
ON CONFLICT (org_id, pack_ref)
DO UPDATE SET
  org_ref = EXCLUDED.org_ref,
  pack_task_no = EXCLUDED.pack_task_no,
  sales_order_id = EXCLUDED.sales_order_id,
  sales_order_ref = EXCLUDED.sales_order_ref,
  order_no = EXCLUDED.order_no,
  pick_task_id = EXCLUDED.pick_task_id,
  pick_task_ref = EXCLUDED.pick_task_ref,
  pick_task_no = EXCLUDED.pick_task_no,
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
  packed_at = EXCLUDED.packed_at,
  packed_by = EXCLUDED.packed_by,
  packed_by_ref = EXCLUDED.packed_by_ref,
  updated_at = EXCLUDED.updated_at,
  updated_by = EXCLUDED.updated_by,
  updated_by_ref = EXCLUDED.updated_by_ref,
  version = shipping.pack_tasks.version + 1
RETURNING id::text`

const deletePackTaskLinesSQL = `
DELETE FROM shipping.pack_task_lines
WHERE pack_task_id = $1::uuid`

const insertPackTaskLineSQL = `
INSERT INTO shipping.pack_task_lines (
  id,
  org_id,
  pack_task_id,
  line_ref,
  pack_task_ref,
  line_no,
  pick_task_line_id,
  pick_task_line_ref,
  sales_order_line_id,
  sales_order_line_ref,
  item_id,
  item_ref,
  sku_code,
  batch_id,
  batch_ref,
  batch_no,
  warehouse_id,
  warehouse_ref,
  base_uom_code,
  qty_to_pack,
  qty_packed,
  status,
  packed_at,
  packed_by,
  packed_by_ref,
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
  $19,
  $20,
  $21,
  $22,
  $23,
  $24::uuid,
  $25,
  $26,
  $27::uuid,
  $28,
  $29,
  $30::uuid,
  $31
)`

func (s PostgresPackTaskStore) ListPackTasks(ctx context.Context) ([]domain.PackTask, error) {
	if s.db == nil {
		return nil, errors.New("database connection is required")
	}
	rows, err := s.db.QueryContext(ctx, selectPackTaskHeadersSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]domain.PackTask, 0)
	for rows.Next() {
		task, err := scanPostgresPackTask(ctx, s.db, rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	domain.SortPackTasks(tasks)

	return tasks, nil
}

func (s PostgresPackTaskStore) GetPackTask(ctx context.Context, id string) (domain.PackTask, error) {
	if s.db == nil {
		return domain.PackTask{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findPackTaskHeaderSQL, strings.TrimSpace(id))
	task, err := scanPostgresPackTask(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PackTask{}, ErrPackTaskNotFound
	}
	if err != nil {
		return domain.PackTask{}, err
	}

	return task, nil
}

func (s PostgresPackTaskStore) GetPackTaskBySalesOrder(
	ctx context.Context,
	salesOrderID string,
) (domain.PackTask, error) {
	if s.db == nil {
		return domain.PackTask{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findPackTaskBySalesOrderSQL, strings.TrimSpace(salesOrderID))
	task, err := scanPostgresPackTask(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PackTask{}, ErrPackTaskNotFound
	}
	if err != nil {
		return domain.PackTask{}, err
	}

	return task, nil
}

func (s PostgresPackTaskStore) GetPackTaskByPickTask(
	ctx context.Context,
	pickTaskID string,
) (domain.PackTask, error) {
	if s.db == nil {
		return domain.PackTask{}, errors.New("database connection is required")
	}
	row := s.db.QueryRowContext(ctx, findPackTaskByPickTaskSQL, strings.TrimSpace(pickTaskID))
	task, err := scanPostgresPackTask(ctx, s.db, row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PackTask{}, ErrPackTaskNotFound
	}
	if err != nil {
		return domain.PackTask{}, err
	}

	return task, nil
}

func (s PostgresPackTaskStore) SavePackTask(ctx context.Context, task domain.PackTask) error {
	if s.db == nil {
		return errors.New("database connection is required")
	}
	if err := task.Validate(); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin pack task transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	persistedID, orgID, err := findPostgresPackTask(ctx, tx, task.ID)
	if err != nil {
		return err
	}
	if orgID == "" {
		orgID, err = s.resolveOrgID(ctx, tx, task.OrgID)
		if err != nil {
			return err
		}
	}
	salesOrderID, orderNo, err := resolvePostgresPackTaskSalesOrder(ctx, tx, orgID, task.SalesOrderID, task.OrderNo)
	if err != nil {
		return err
	}
	pickTaskID, pickTaskNo, err := resolvePostgresPackTaskPickTask(ctx, tx, orgID, task.PickTaskID, task.PickTaskNo)
	if err != nil {
		return err
	}
	warehouseID, warehouseCode, err := resolvePostgresPackTaskWarehouse(ctx, tx, orgID, task.WarehouseID, task.WarehouseCode)
	if err != nil {
		return err
	}
	persistedID, err = upsertPostgresPackTask(ctx, tx, persistedID, orgID, salesOrderID, orderNo, pickTaskID, pickTaskNo, warehouseID, warehouseCode, task)
	if err != nil {
		return err
	}
	if err := replacePostgresPackTaskLines(ctx, tx, orgID, persistedID, task); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit pack task transaction: %w", err)
	}
	committed = true

	return nil
}

func (s PostgresPackTaskStore) resolveOrgID(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	orgRef string,
) (string, error) {
	orgRef = strings.TrimSpace(orgRef)
	if isPostgresPackTaskUUIDText(orgRef) {
		return orgRef, nil
	}
	if orgRef != "" {
		var orgID string
		err := queryer.QueryRowContext(ctx, selectPackTaskOrgIDSQL, orgRef).Scan(&orgID)
		if err == nil && isPostgresPackTaskUUIDText(orgID) {
			return orgID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("resolve pack task org %q: %w", orgRef, err)
		}
	}
	if isPostgresPackTaskUUIDText(s.defaultOrgID) {
		return s.defaultOrgID, nil
	}

	return "", fmt.Errorf("pack task org %q cannot be resolved", orgRef)
}

func scanPostgresPackTask(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	row interface{ Scan(dest ...any) error },
) (domain.PackTask, error) {
	var (
		persistedID string
		status      string
		assignedAt  sql.NullTime
		startedAt   sql.NullTime
		packedAt    sql.NullTime
		task        domain.PackTask
	)
	if err := row.Scan(
		&persistedID,
		&task.ID,
		&task.OrgID,
		&task.PackTaskNo,
		&task.SalesOrderID,
		&task.OrderNo,
		&task.PickTaskID,
		&task.PickTaskNo,
		&task.WarehouseID,
		&task.WarehouseCode,
		&status,
		&task.AssignedTo,
		&assignedAt,
		&startedAt,
		&task.StartedBy,
		&packedAt,
		&task.PackedBy,
		&task.CreatedAt,
		&task.UpdatedAt,
	); err != nil {
		return domain.PackTask{}, err
	}
	task.Status = domain.NormalizePackTaskStatus(domain.PackTaskStatus(status))
	if assignedAt.Valid {
		task.AssignedAt = assignedAt.Time.UTC()
	}
	if startedAt.Valid {
		task.StartedAt = startedAt.Time.UTC()
	}
	if packedAt.Valid {
		task.PackedAt = packedAt.Time.UTC()
	}
	lines, err := listPostgresPackTaskLines(ctx, queryer, persistedID)
	if err != nil {
		return domain.PackTask{}, err
	}
	task.Lines = lines
	if err := task.Validate(); err != nil {
		return domain.PackTask{}, err
	}

	return task, nil
}

func listPostgresPackTaskLines(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	persistedID string,
) ([]domain.PackTaskLine, error) {
	rows, err := queryer.QueryContext(ctx, selectPackTaskLinesSQL, persistedID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lines := make([]domain.PackTaskLine, 0)
	for rows.Next() {
		line, err := scanPostgresPackTaskLine(rows)
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

func scanPostgresPackTaskLine(row interface{ Scan(dest ...any) error }) (domain.PackTaskLine, error) {
	var (
		line          domain.PackTaskLine
		qtyToPack     string
		qtyPacked     string
		baseUOMCode   string
		status        string
		packedAt      sql.NullTime
		parsedUOMCode decimal.UOMCode
		err           error
	)
	if err := row.Scan(
		&line.ID,
		&line.PackTaskID,
		&line.LineNo,
		&line.PickTaskLineID,
		&line.SalesOrderLineID,
		&line.ItemID,
		&line.SKUCode,
		&line.BatchID,
		&line.BatchNo,
		&line.WarehouseID,
		&qtyToPack,
		&qtyPacked,
		&baseUOMCode,
		&status,
		&packedAt,
		&line.PackedBy,
		&line.CreatedAt,
		&line.UpdatedAt,
	); err != nil {
		return domain.PackTaskLine{}, err
	}
	line.QtyToPack, err = decimal.ParseQuantity(qtyToPack)
	if err != nil {
		return domain.PackTaskLine{}, err
	}
	line.QtyPacked, err = decimal.ParseQuantity(qtyPacked)
	if err != nil {
		return domain.PackTaskLine{}, err
	}
	parsedUOMCode, err = decimal.NormalizeUOMCode(baseUOMCode)
	if err != nil {
		return domain.PackTaskLine{}, err
	}
	line.BaseUOMCode = parsedUOMCode
	line.Status = domain.NormalizePackTaskLineStatus(domain.PackTaskLineStatus(status))
	if packedAt.Valid {
		line.PackedAt = packedAt.Time.UTC()
	}
	if err := line.Validate(); err != nil {
		return domain.PackTaskLine{}, err
	}

	return line, nil
}

func findPostgresPackTask(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	id string,
) (string, string, error) {
	var persistedID string
	var orgID string
	err := queryer.QueryRowContext(ctx, findPackTaskPersistedSQL, strings.TrimSpace(id)).Scan(&persistedID, &orgID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", nil
	}
	if err != nil {
		return "", "", fmt.Errorf("find pack task %q: %w", id, err)
	}

	return persistedID, orgID, nil
}

func resolvePostgresPackTaskWarehouse(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	orgID string,
	warehouseRef string,
	warehouseCode string,
) (string, string, error) {
	warehouseRef = strings.TrimSpace(warehouseRef)
	warehouseCode = strings.TrimSpace(warehouseCode)
	if isPostgresPackTaskUUIDText(warehouseRef) && warehouseCode != "" {
		return warehouseRef, warehouseCode, nil
	}
	var id string
	var code string
	err := queryer.QueryRowContext(ctx, selectPackTaskWarehouseSQL, orgID, warehouseRef, warehouseCode).Scan(&id, &code)
	if err != nil {
		return "", "", fmt.Errorf("resolve pack task warehouse %q: %w", warehouseRef, err)
	}

	return id, code, nil
}

func resolvePostgresPackTaskSalesOrder(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	orgID string,
	salesOrderRef string,
	orderNo string,
) (string, string, error) {
	salesOrderRef = strings.TrimSpace(salesOrderRef)
	orderNo = strings.TrimSpace(orderNo)
	if isPostgresPackTaskUUIDText(salesOrderRef) && orderNo != "" {
		return salesOrderRef, orderNo, nil
	}
	var id string
	var resolvedOrderNo string
	err := queryer.QueryRowContext(ctx, selectPackTaskSalesOrderSQL, orgID, salesOrderRef, orderNo).Scan(&id, &resolvedOrderNo)
	if err != nil {
		return "", "", fmt.Errorf("resolve pack task sales order %q: %w", salesOrderRef, err)
	}

	return id, resolvedOrderNo, nil
}

func resolvePostgresPackTaskPickTask(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	orgID string,
	pickTaskRef string,
	pickTaskNo string,
) (string, string, error) {
	pickTaskRef = strings.TrimSpace(pickTaskRef)
	pickTaskNo = strings.TrimSpace(pickTaskNo)
	if isPostgresPackTaskUUIDText(pickTaskRef) && pickTaskNo != "" {
		return pickTaskRef, pickTaskNo, nil
	}
	var id string
	var resolvedPickTaskNo string
	err := queryer.QueryRowContext(ctx, selectPackTaskPickTaskSQL, orgID, pickTaskRef, pickTaskNo).Scan(&id, &resolvedPickTaskNo)
	if err != nil {
		return "", "", fmt.Errorf("resolve pack task pick task %q: %w", pickTaskRef, err)
	}

	return id, resolvedPickTaskNo, nil
}

func upsertPostgresPackTask(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	persistedID string,
	orgID string,
	salesOrderID string,
	orderNo string,
	pickTaskID string,
	pickTaskNo string,
	warehouseID string,
	warehouseCode string,
	task domain.PackTask,
) (string, error) {
	var savedID string
	createdBy := firstNonBlankPostgresPackTask(task.AssignedTo, task.StartedBy, task.PackedBy)
	updatedBy := firstNonBlankPostgresPackTask(task.PackedBy, task.StartedBy, task.AssignedTo)
	err := queryer.QueryRowContext(
		ctx,
		upsertPackTaskSQL,
		nullablePostgresPackTaskUUID(firstNonBlankPostgresPackTask(persistedID, task.ID)),
		orgID,
		nullablePostgresPackTaskText(task.ID),
		nullablePostgresPackTaskText(task.OrgID),
		task.PackTaskNo,
		salesOrderID,
		nullablePostgresPackTaskText(task.SalesOrderID),
		nullablePostgresPackTaskText(firstNonBlankPostgresPackTask(task.OrderNo, orderNo)),
		pickTaskID,
		nullablePostgresPackTaskText(task.PickTaskID),
		nullablePostgresPackTaskText(firstNonBlankPostgresPackTask(task.PickTaskNo, pickTaskNo)),
		warehouseID,
		nullablePostgresPackTaskText(task.WarehouseID),
		nullablePostgresPackTaskText(firstNonBlankPostgresPackTask(task.WarehouseCode, warehouseCode)),
		string(task.Status),
		nullablePostgresPackTaskUUID(task.AssignedTo),
		nullablePostgresPackTaskText(task.AssignedTo),
		nullablePostgresPackTaskTime(task.AssignedAt),
		nullablePostgresPackTaskTime(task.StartedAt),
		nullablePostgresPackTaskUUID(task.StartedBy),
		nullablePostgresPackTaskText(task.StartedBy),
		nullablePostgresPackTaskTime(task.PackedAt),
		nullablePostgresPackTaskUUID(task.PackedBy),
		nullablePostgresPackTaskText(task.PackedBy),
		postgresPackTaskTime(task.CreatedAt),
		nullablePostgresPackTaskUUID(createdBy),
		nullablePostgresPackTaskText(createdBy),
		postgresPackTaskTime(task.UpdatedAt),
		nullablePostgresPackTaskUUID(updatedBy),
		nullablePostgresPackTaskText(updatedBy),
	).Scan(&savedID)
	if err != nil {
		return "", fmt.Errorf("upsert pack task %q: %w", task.ID, err)
	}

	return savedID, nil
}

func replacePostgresPackTaskLines(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	orgID string,
	persistedID string,
	task domain.PackTask,
) error {
	if _, err := queryer.ExecContext(ctx, deletePackTaskLinesSQL, persistedID); err != nil {
		return fmt.Errorf("delete pack task lines: %w", err)
	}
	for _, line := range task.Lines {
		pickTaskLineID, err := resolvePostgresPackTaskID(ctx, queryer, selectPackTaskPickTaskLineSQL, line.PickTaskLineID, "pick task line")
		if err != nil {
			return err
		}
		salesOrderLineID, err := resolvePostgresPackTaskID(ctx, queryer, selectPackTaskSalesOrderLineSQL, line.SalesOrderLineID, "sales order line")
		if err != nil {
			return err
		}
		itemID, skuCode, err := resolvePostgresPackTaskItem(ctx, queryer, orgID, line.ItemID, line.SKUCode)
		if err != nil {
			return err
		}
		batchID, batchNo, err := resolvePostgresPackTaskBatch(ctx, queryer, orgID, line.BatchID, line.BatchNo)
		if err != nil {
			return err
		}
		warehouseID, _, err := resolvePostgresPackTaskWarehouse(ctx, queryer, orgID, line.WarehouseID, task.WarehouseCode)
		if err != nil {
			return err
		}
		if _, err := queryer.ExecContext(
			ctx,
			insertPackTaskLineSQL,
			nullablePostgresPackTaskUUID(line.ID),
			orgID,
			persistedID,
			nullablePostgresPackTaskText(line.ID),
			nullablePostgresPackTaskText(task.ID),
			line.LineNo,
			pickTaskLineID,
			nullablePostgresPackTaskText(line.PickTaskLineID),
			salesOrderLineID,
			nullablePostgresPackTaskText(line.SalesOrderLineID),
			itemID,
			nullablePostgresPackTaskText(line.ItemID),
			firstNonBlankPostgresPackTask(line.SKUCode, skuCode),
			batchID,
			nullablePostgresPackTaskText(line.BatchID),
			nullablePostgresPackTaskText(firstNonBlankPostgresPackTask(line.BatchNo, batchNo)),
			warehouseID,
			nullablePostgresPackTaskText(line.WarehouseID),
			line.BaseUOMCode.String(),
			line.QtyToPack.String(),
			line.QtyPacked.String(),
			string(line.Status),
			nullablePostgresPackTaskTime(line.PackedAt),
			nullablePostgresPackTaskUUID(line.PackedBy),
			nullablePostgresPackTaskText(line.PackedBy),
			postgresPackTaskTime(line.CreatedAt),
			nullablePostgresPackTaskUUID(line.PackedBy),
			nullablePostgresPackTaskText(line.PackedBy),
			postgresPackTaskTime(line.UpdatedAt),
			nullablePostgresPackTaskUUID(line.PackedBy),
			nullablePostgresPackTaskText(line.PackedBy),
		); err != nil {
			return fmt.Errorf("insert pack task line %q: %w", line.ID, err)
		}
	}

	return nil
}

func resolvePostgresPackTaskID(
	ctx context.Context,
	queryer postgresPackTaskQueryer,
	query string,
	ref string,
	label string,
) (string, error) {
	var id string
	err := queryer.QueryRowContext(ctx, query, strings.TrimSpace(ref)).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("resolve pack task %s %q: %w", label, ref, err)
	}

	return id, nil
}

func resolvePostgresPackTaskItem(ctx context.Context, queryer postgresPackTaskQueryer, orgID string, itemRef string, sku string) (string, string, error) {
	var id string
	var resolvedSKU string
	err := queryer.QueryRowContext(ctx, selectPackTaskItemSQL, orgID, strings.TrimSpace(itemRef), strings.TrimSpace(sku)).
		Scan(&id, &resolvedSKU)
	if err != nil {
		return "", "", fmt.Errorf("resolve pack task item %q: %w", itemRef, err)
	}

	return id, resolvedSKU, nil
}

func resolvePostgresPackTaskBatch(ctx context.Context, queryer postgresPackTaskQueryer, orgID string, batchRef string, batchNo string) (string, string, error) {
	var id string
	var resolvedBatchNo string
	err := queryer.QueryRowContext(ctx, selectPackTaskBatchSQL, orgID, strings.TrimSpace(batchRef), strings.TrimSpace(batchNo)).
		Scan(&id, &resolvedBatchNo)
	if err != nil {
		return "", "", fmt.Errorf("resolve pack task batch %q: %w", batchRef, err)
	}

	return id, resolvedBatchNo, nil
}

func nullablePostgresPackTaskText(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}

func nullablePostgresPackTaskUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isPostgresPackTaskUUIDText(value) {
		return nil
	}

	return value
}

func nullablePostgresPackTaskTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}

	return value.UTC()
}

func postgresPackTaskTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func firstNonBlankPostgresPackTask(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}

func isPostgresPackTaskUUIDText(value string) bool {
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
			if !isPostgresPackTaskHexText(char) {
				return false
			}
		}
	}

	return true
}

func isPostgresPackTaskHexText(char rune) bool {
	return (char >= '0' && char <= '9') ||
		(char >= 'a' && char <= 'f') ||
		(char >= 'A' && char <= 'F')
}

var _ PackTaskStore = PostgresPackTaskStore{}
