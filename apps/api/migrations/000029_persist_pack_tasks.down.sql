BEGIN;

DROP INDEX IF EXISTS shipping.ix_pack_task_lines_runtime_task;
DROP INDEX IF EXISTS shipping.uq_pack_task_lines_ref;
DROP INDEX IF EXISTS shipping.ix_pack_tasks_runtime_filters;
DROP INDEX IF EXISTS shipping.uq_pack_tasks_org_ref;

ALTER TABLE shipping.pack_task_lines
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS packed_by_ref,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS batch_no,
  DROP COLUMN IF EXISTS batch_ref,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS sales_order_line_ref,
  DROP COLUMN IF EXISTS pick_task_line_ref,
  DROP COLUMN IF EXISTS pack_task_ref,
  DROP COLUMN IF EXISTS line_ref;

ALTER TABLE shipping.pack_tasks
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS packed_by_ref,
  DROP COLUMN IF EXISTS started_by_ref,
  DROP COLUMN IF EXISTS assigned_to_ref,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS pick_task_no,
  DROP COLUMN IF EXISTS pick_task_ref,
  DROP COLUMN IF EXISTS order_no,
  DROP COLUMN IF EXISTS sales_order_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS pack_ref;

COMMIT;
