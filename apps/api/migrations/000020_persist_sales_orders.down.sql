BEGIN;

DROP INDEX IF EXISTS sales.uq_sales_order_lines_order_ref;
DROP INDEX IF EXISTS sales.ix_sales_orders_warehouse_ref_status;
DROP INDEX IF EXISTS sales.ix_sales_orders_customer_ref_order_date;
DROP INDEX IF EXISTS sales.uq_sales_orders_org_ref;

ALTER TABLE sales.sales_order_lines
  DROP CONSTRAINT IF EXISTS ck_sales_order_lines_runtime_refs;

ALTER TABLE sales.sales_orders
  DROP CONSTRAINT IF EXISTS ck_sales_orders_runtime_refs;

ALTER TABLE sales.sales_order_lines
  ALTER COLUMN item_id SET NOT NULL,
  ALTER COLUMN unit_id SET NOT NULL,
  DROP COLUMN IF EXISTS batch_no,
  DROP COLUMN IF EXISTS batch_ref,
  DROP COLUMN IF EXISTS batch_id,
  DROP COLUMN IF EXISTS item_name,
  DROP COLUMN IF EXISTS sku_code,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS line_ref;

ALTER TABLE sales.sales_orders
  DROP COLUMN IF EXISTS exception_by_ref,
  DROP COLUMN IF EXISTS exception_by,
  DROP COLUMN IF EXISTS exception_at,
  DROP COLUMN IF EXISTS closed_by_ref,
  DROP COLUMN IF EXISTS closed_by,
  DROP COLUMN IF EXISTS closed_at,
  DROP COLUMN IF EXISTS handed_over_by_ref,
  DROP COLUMN IF EXISTS handed_over_by,
  DROP COLUMN IF EXISTS handed_over_at,
  DROP COLUMN IF EXISTS waiting_handover_by_ref,
  DROP COLUMN IF EXISTS waiting_handover_by,
  DROP COLUMN IF EXISTS waiting_handover_at,
  DROP COLUMN IF EXISTS packed_by_ref,
  DROP COLUMN IF EXISTS packed_by,
  DROP COLUMN IF EXISTS packed_at,
  DROP COLUMN IF EXISTS packing_started_by_ref,
  DROP COLUMN IF EXISTS packing_started_by,
  DROP COLUMN IF EXISTS packing_started_at,
  DROP COLUMN IF EXISTS picked_by_ref,
  DROP COLUMN IF EXISTS picked_by,
  DROP COLUMN IF EXISTS picked_at,
  DROP COLUMN IF EXISTS picking_started_by_ref,
  DROP COLUMN IF EXISTS picking_started_by,
  DROP COLUMN IF EXISTS picking_started_at,
  DROP COLUMN IF EXISTS reserved_by_ref,
  DROP COLUMN IF EXISTS reserved_by,
  DROP COLUMN IF EXISTS reserved_at,
  DROP COLUMN IF EXISTS confirmed_by_ref,
  DROP COLUMN IF EXISTS confirmed_by,
  DROP COLUMN IF EXISTS confirmed_at,
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS note,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS warehouse_id,
  DROP COLUMN IF EXISTS customer_name,
  DROP COLUMN IF EXISTS customer_code,
  DROP COLUMN IF EXISTS customer_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS order_ref;

COMMIT;
