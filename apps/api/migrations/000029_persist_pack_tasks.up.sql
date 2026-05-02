BEGIN;

ALTER TABLE shipping.pack_tasks
  ADD COLUMN IF NOT EXISTS pack_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_ref text,
  ADD COLUMN IF NOT EXISTS order_no text,
  ADD COLUMN IF NOT EXISTS pick_task_ref text,
  ADD COLUMN IF NOT EXISTS pick_task_no text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS assigned_to_ref text,
  ADD COLUMN IF NOT EXISTS started_by_ref text,
  ADD COLUMN IF NOT EXISTS packed_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE shipping.pack_tasks AS pack_task
SET pack_ref = COALESCE(NULLIF(btrim(pack_task.pack_ref), ''), pack_task.pack_task_no, pack_task.id::text),
    org_ref = COALESCE(NULLIF(btrim(pack_task.org_ref), ''), pack_task.org_id::text),
    sales_order_ref = COALESCE(NULLIF(btrim(pack_task.sales_order_ref), ''), sales_order.order_ref, sales_order.order_no, pack_task.sales_order_id::text),
    order_no = COALESCE(NULLIF(btrim(pack_task.order_no), ''), sales_order.order_no),
    pick_task_ref = COALESCE(NULLIF(btrim(pack_task.pick_task_ref), ''), pick_task.pick_ref, pick_task.pick_task_no, pack_task.pick_task_id::text),
    pick_task_no = COALESCE(NULLIF(btrim(pack_task.pick_task_no), ''), pick_task.pick_task_no),
    warehouse_ref = COALESCE(NULLIF(btrim(pack_task.warehouse_ref), ''), pack_task.warehouse_id::text),
    warehouse_code = COALESCE(NULLIF(btrim(pack_task.warehouse_code), ''), warehouse.code),
    assigned_to_ref = COALESCE(NULLIF(btrim(pack_task.assigned_to_ref), ''), pack_task.assigned_to::text),
    started_by_ref = COALESCE(NULLIF(btrim(pack_task.started_by_ref), ''), pack_task.started_by::text),
    packed_by_ref = COALESCE(NULLIF(btrim(pack_task.packed_by_ref), ''), pack_task.packed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(pack_task.created_by_ref), ''), pack_task.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(pack_task.updated_by_ref), ''), pack_task.updated_by::text)
FROM sales.sales_orders AS sales_order,
     shipping.pick_tasks AS pick_task,
     mdm.warehouses AS warehouse
WHERE sales_order.id = pack_task.sales_order_id
  AND pick_task.id = pack_task.pick_task_id
  AND warehouse.id = pack_task.warehouse_id;

UPDATE shipping.pack_tasks
SET pack_ref = COALESCE(NULLIF(btrim(pack_ref), ''), pack_task_no, id::text),
    org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text),
    sales_order_ref = COALESCE(NULLIF(btrim(sales_order_ref), ''), sales_order_id::text),
    pick_task_ref = COALESCE(NULLIF(btrim(pick_task_ref), ''), pick_task_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text),
    assigned_to_ref = COALESCE(NULLIF(btrim(assigned_to_ref), ''), assigned_to::text),
    started_by_ref = COALESCE(NULLIF(btrim(started_by_ref), ''), started_by::text),
    packed_by_ref = COALESCE(NULLIF(btrim(packed_by_ref), ''), packed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE pack_ref IS NULL
   OR org_ref IS NULL
   OR sales_order_ref IS NULL
   OR pick_task_ref IS NULL
   OR warehouse_ref IS NULL
   OR assigned_to_ref IS NULL
   OR started_by_ref IS NULL
   OR packed_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

ALTER TABLE shipping.pack_tasks
  ALTER COLUMN pack_ref SET NOT NULL,
  ALTER COLUMN org_ref SET NOT NULL,
  ALTER COLUMN sales_order_ref SET NOT NULL,
  ALTER COLUMN pick_task_ref SET NOT NULL,
  ALTER COLUMN warehouse_ref SET NOT NULL;

ALTER TABLE shipping.pack_task_lines
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS pack_task_ref text,
  ADD COLUMN IF NOT EXISTS pick_task_line_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_line_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS batch_ref text,
  ADD COLUMN IF NOT EXISTS batch_no text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS packed_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE shipping.pack_task_lines AS pack_line
SET line_ref = COALESCE(NULLIF(btrim(pack_line.line_ref), ''), pack_line.id::text),
    pack_task_ref = COALESCE(NULLIF(btrim(pack_line.pack_task_ref), ''), pack_task.pack_ref, pack_task.id::text),
    pick_task_line_ref = COALESCE(NULLIF(btrim(pack_line.pick_task_line_ref), ''), pick_line.line_ref, pack_line.pick_task_line_id::text),
    sales_order_line_ref = COALESCE(NULLIF(btrim(pack_line.sales_order_line_ref), ''), sales_order_line.line_ref, pack_line.sales_order_line_id::text),
    item_ref = COALESCE(NULLIF(btrim(pack_line.item_ref), ''), item.sku, pack_line.item_id::text),
    batch_ref = COALESCE(NULLIF(btrim(pack_line.batch_ref), ''), batch.batch_ref, pack_line.batch_id::text),
    batch_no = COALESCE(NULLIF(btrim(pack_line.batch_no), ''), batch.batch_no),
    warehouse_ref = COALESCE(NULLIF(btrim(pack_line.warehouse_ref), ''), pack_line.warehouse_id::text),
    packed_by_ref = COALESCE(NULLIF(btrim(pack_line.packed_by_ref), ''), pack_line.packed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(pack_line.created_by_ref), ''), pack_line.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(pack_line.updated_by_ref), ''), pack_line.updated_by::text)
FROM shipping.pack_tasks AS pack_task,
     shipping.pick_task_lines AS pick_line,
     sales.sales_order_lines AS sales_order_line,
     mdm.items AS item,
     inventory.batches AS batch
WHERE pack_task.id = pack_line.pack_task_id
  AND pick_line.id = pack_line.pick_task_line_id
  AND sales_order_line.id = pack_line.sales_order_line_id
  AND item.id = pack_line.item_id
  AND batch.id = pack_line.batch_id;

UPDATE shipping.pack_task_lines
SET line_ref = COALESCE(NULLIF(btrim(line_ref), ''), id::text),
    pack_task_ref = COALESCE(NULLIF(btrim(pack_task_ref), ''), pack_task_id::text),
    pick_task_line_ref = COALESCE(NULLIF(btrim(pick_task_line_ref), ''), pick_task_line_id::text),
    sales_order_line_ref = COALESCE(NULLIF(btrim(sales_order_line_ref), ''), sales_order_line_id::text),
    item_ref = COALESCE(NULLIF(btrim(item_ref), ''), item_id::text),
    batch_ref = COALESCE(NULLIF(btrim(batch_ref), ''), batch_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text),
    packed_by_ref = COALESCE(NULLIF(btrim(packed_by_ref), ''), packed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE line_ref IS NULL
   OR pack_task_ref IS NULL
   OR pick_task_line_ref IS NULL
   OR sales_order_line_ref IS NULL
   OR item_ref IS NULL
   OR batch_ref IS NULL
   OR warehouse_ref IS NULL
   OR packed_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

ALTER TABLE shipping.pack_task_lines
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN pack_task_ref SET NOT NULL,
  ALTER COLUMN pick_task_line_ref SET NOT NULL,
  ALTER COLUMN sales_order_line_ref SET NOT NULL,
  ALTER COLUMN item_ref SET NOT NULL,
  ALTER COLUMN batch_ref SET NOT NULL,
  ALTER COLUMN warehouse_ref SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_pack_tasks_org_ref
  ON shipping.pack_tasks(org_id, pack_ref);

CREATE INDEX IF NOT EXISTS ix_pack_tasks_runtime_filters
  ON shipping.pack_tasks(org_id, warehouse_ref, status, assigned_to_ref, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_pack_task_lines_ref
  ON shipping.pack_task_lines(org_id, pack_task_id, line_ref);

CREATE INDEX IF NOT EXISTS ix_pack_task_lines_runtime_task
  ON shipping.pack_task_lines(org_id, pack_task_ref, status, line_no);

COMMIT;
