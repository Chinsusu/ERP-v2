BEGIN;

DROP INDEX IF EXISTS returns.ix_return_attachments_return_ref;
DROP INDEX IF EXISTS returns.uq_return_attachments_org_ref;
DROP INDEX IF EXISTS returns.ix_return_disposition_actions_return_ref;
DROP INDEX IF EXISTS returns.uq_return_disposition_actions_org_ref;
DROP INDEX IF EXISTS returns.ix_return_inspections_return_ref;
DROP INDEX IF EXISTS returns.uq_return_inspections_org_ref;
DROP INDEX IF EXISTS returns.uq_return_order_lines_order_ref;
DROP INDEX IF EXISTS returns.ix_return_orders_return_code;
DROP INDEX IF EXISTS returns.ix_return_orders_warehouse_ref_status;
DROP INDEX IF EXISTS returns.uq_return_orders_org_ref;

DROP TABLE IF EXISTS returns.return_attachments;
DROP TABLE IF EXISTS returns.return_disposition_actions;
DROP TABLE IF EXISTS returns.return_inspections;

ALTER TABLE returns.return_order_lines
  DROP CONSTRAINT IF EXISTS ck_return_order_lines_runtime_refs;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_runtime_refs;

ALTER TABLE returns.return_order_lines
  ALTER COLUMN item_id SET NOT NULL,
  ALTER COLUMN unit_id SET NOT NULL,
  DROP COLUMN IF EXISTS stock_movement_ref,
  DROP COLUMN IF EXISTS uom_code,
  DROP COLUMN IF EXISTS unit_ref,
  DROP COLUMN IF EXISTS batch_ref,
  DROP COLUMN IF EXISTS condition_text,
  DROP COLUMN IF EXISTS quantity,
  DROP COLUMN IF EXISTS product_name,
  DROP COLUMN IF EXISTS sku_code,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS line_ref;

ALTER TABLE returns.return_orders
  ALTER COLUMN warehouse_id SET NOT NULL,
  DROP COLUMN IF EXISTS target_stock_status,
  DROP COLUMN IF EXISTS stock_movement_type,
  DROP COLUMN IF EXISTS stock_movement_ref,
  DROP COLUMN IF EXISTS scan_code,
  DROP COLUMN IF EXISTS customer_name,
  DROP COLUMN IF EXISTS original_order_no,
  DROP COLUMN IF EXISTS original_order_ref,
  DROP COLUMN IF EXISTS target_location,
  DROP COLUMN IF EXISTS disposition,
  DROP COLUMN IF EXISTS received_by_ref,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS return_ref;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_status;

ALTER TABLE returns.return_orders
  ADD CONSTRAINT ck_return_orders_status CHECK (
    status IN ('received', 'pending_inspection', 'inspected', 'disposed', 'closed', 'cancelled')
  );

COMMIT;
