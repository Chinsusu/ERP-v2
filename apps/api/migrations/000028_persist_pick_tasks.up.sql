BEGIN;

ALTER TABLE shipping.pick_tasks
  ADD COLUMN IF NOT EXISTS pick_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_ref text,
  ADD COLUMN IF NOT EXISTS order_no text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS assigned_to_ref text,
  ADD COLUMN IF NOT EXISTS started_by_ref text,
  ADD COLUMN IF NOT EXISTS completed_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE shipping.pick_tasks AS pick_task
SET pick_ref = COALESCE(NULLIF(btrim(pick_task.pick_ref), ''), pick_task.pick_task_no, pick_task.id::text),
    org_ref = COALESCE(NULLIF(btrim(pick_task.org_ref), ''), pick_task.org_id::text),
    sales_order_ref = COALESCE(NULLIF(btrim(pick_task.sales_order_ref), ''), sales_order.order_ref, sales_order.order_no, pick_task.sales_order_id::text),
    order_no = COALESCE(NULLIF(btrim(pick_task.order_no), ''), sales_order.order_no),
    warehouse_ref = COALESCE(NULLIF(btrim(pick_task.warehouse_ref), ''), pick_task.warehouse_id::text),
    warehouse_code = COALESCE(NULLIF(btrim(pick_task.warehouse_code), ''), warehouse.code),
    assigned_to_ref = COALESCE(NULLIF(btrim(pick_task.assigned_to_ref), ''), pick_task.assigned_to::text),
    started_by_ref = COALESCE(NULLIF(btrim(pick_task.started_by_ref), ''), pick_task.started_by::text),
    completed_by_ref = COALESCE(NULLIF(btrim(pick_task.completed_by_ref), ''), pick_task.completed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(pick_task.created_by_ref), ''), pick_task.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(pick_task.updated_by_ref), ''), pick_task.updated_by::text)
FROM sales.sales_orders AS sales_order,
     mdm.warehouses AS warehouse
WHERE sales_order.id = pick_task.sales_order_id
  AND warehouse.id = pick_task.warehouse_id;

UPDATE shipping.pick_tasks
SET pick_ref = COALESCE(NULLIF(btrim(pick_ref), ''), pick_task_no, id::text),
    org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text),
    sales_order_ref = COALESCE(NULLIF(btrim(sales_order_ref), ''), sales_order_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text),
    assigned_to_ref = COALESCE(NULLIF(btrim(assigned_to_ref), ''), assigned_to::text),
    started_by_ref = COALESCE(NULLIF(btrim(started_by_ref), ''), started_by::text),
    completed_by_ref = COALESCE(NULLIF(btrim(completed_by_ref), ''), completed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE pick_ref IS NULL
   OR org_ref IS NULL
   OR sales_order_ref IS NULL
   OR warehouse_ref IS NULL
   OR assigned_to_ref IS NULL
   OR started_by_ref IS NULL
   OR completed_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

ALTER TABLE shipping.pick_tasks
  ALTER COLUMN pick_ref SET NOT NULL,
  ALTER COLUMN org_ref SET NOT NULL,
  ALTER COLUMN sales_order_ref SET NOT NULL,
  ALTER COLUMN warehouse_ref SET NOT NULL;

ALTER TABLE shipping.pick_task_lines
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS pick_task_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_line_ref text,
  ADD COLUMN IF NOT EXISTS stock_reservation_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS batch_ref text,
  ADD COLUMN IF NOT EXISTS batch_no text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS bin_ref text,
  ADD COLUMN IF NOT EXISTS bin_code text,
  ADD COLUMN IF NOT EXISTS picked_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE shipping.pick_task_lines AS pick_line
SET line_ref = COALESCE(NULLIF(btrim(pick_line.line_ref), ''), pick_line.id::text),
    pick_task_ref = COALESCE(NULLIF(btrim(pick_line.pick_task_ref), ''), pick_task.pick_ref, pick_task.id::text),
    sales_order_line_ref = COALESCE(NULLIF(btrim(pick_line.sales_order_line_ref), ''), sales_order_line.line_ref, pick_line.sales_order_line_id::text),
    stock_reservation_ref = COALESCE(NULLIF(btrim(pick_line.stock_reservation_ref), ''), stock_reservation.reservation_ref, pick_line.stock_reservation_id::text),
    item_ref = COALESCE(NULLIF(btrim(pick_line.item_ref), ''), item.sku, pick_line.item_id::text),
    batch_ref = COALESCE(NULLIF(btrim(pick_line.batch_ref), ''), batch.batch_ref, pick_line.batch_id::text),
    batch_no = COALESCE(NULLIF(btrim(pick_line.batch_no), ''), batch.batch_no),
    warehouse_ref = COALESCE(NULLIF(btrim(pick_line.warehouse_ref), ''), pick_line.warehouse_id::text),
    bin_ref = COALESCE(NULLIF(btrim(pick_line.bin_ref), ''), warehouse_bin.code, pick_line.bin_id::text),
    bin_code = COALESCE(NULLIF(btrim(pick_line.bin_code), ''), warehouse_bin.code),
    picked_by_ref = COALESCE(NULLIF(btrim(pick_line.picked_by_ref), ''), pick_line.picked_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(pick_line.created_by_ref), ''), pick_line.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(pick_line.updated_by_ref), ''), pick_line.updated_by::text)
FROM shipping.pick_tasks AS pick_task,
     sales.sales_order_lines AS sales_order_line,
     inventory.stock_reservations AS stock_reservation,
     mdm.items AS item,
     inventory.batches AS batch,
     mdm.warehouse_bins AS warehouse_bin
WHERE pick_task.id = pick_line.pick_task_id
  AND sales_order_line.id = pick_line.sales_order_line_id
  AND stock_reservation.id = pick_line.stock_reservation_id
  AND item.id = pick_line.item_id
  AND batch.id = pick_line.batch_id
  AND warehouse_bin.id = pick_line.bin_id;

UPDATE shipping.pick_task_lines
SET line_ref = COALESCE(NULLIF(btrim(line_ref), ''), id::text),
    pick_task_ref = COALESCE(NULLIF(btrim(pick_task_ref), ''), pick_task_id::text),
    sales_order_line_ref = COALESCE(NULLIF(btrim(sales_order_line_ref), ''), sales_order_line_id::text),
    stock_reservation_ref = COALESCE(NULLIF(btrim(stock_reservation_ref), ''), stock_reservation_id::text),
    item_ref = COALESCE(NULLIF(btrim(item_ref), ''), item_id::text),
    batch_ref = COALESCE(NULLIF(btrim(batch_ref), ''), batch_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text),
    bin_ref = COALESCE(NULLIF(btrim(bin_ref), ''), bin_id::text),
    picked_by_ref = COALESCE(NULLIF(btrim(picked_by_ref), ''), picked_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE line_ref IS NULL
   OR pick_task_ref IS NULL
   OR sales_order_line_ref IS NULL
   OR stock_reservation_ref IS NULL
   OR item_ref IS NULL
   OR batch_ref IS NULL
   OR warehouse_ref IS NULL
   OR bin_ref IS NULL
   OR picked_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

ALTER TABLE shipping.pick_task_lines
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN pick_task_ref SET NOT NULL,
  ALTER COLUMN sales_order_line_ref SET NOT NULL,
  ALTER COLUMN stock_reservation_ref SET NOT NULL,
  ALTER COLUMN item_ref SET NOT NULL,
  ALTER COLUMN batch_ref SET NOT NULL,
  ALTER COLUMN warehouse_ref SET NOT NULL,
  ALTER COLUMN bin_ref SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_pick_tasks_org_ref
  ON shipping.pick_tasks(org_id, pick_ref);

CREATE INDEX IF NOT EXISTS ix_pick_tasks_runtime_filters
  ON shipping.pick_tasks(org_id, warehouse_ref, status, assigned_to_ref, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_pick_task_lines_ref
  ON shipping.pick_task_lines(org_id, pick_task_id, line_ref);

CREATE INDEX IF NOT EXISTS ix_pick_task_lines_runtime_task
  ON shipping.pick_task_lines(org_id, pick_task_ref, status, line_no);

COMMIT;
