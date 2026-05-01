BEGIN;

DROP INDEX IF EXISTS purchase.uq_purchase_order_lines_order_ref;
DROP INDEX IF EXISTS purchase.ix_purchase_orders_warehouse_ref_status;
DROP INDEX IF EXISTS purchase.ix_purchase_orders_supplier_ref_expected;
DROP INDEX IF EXISTS purchase.uq_purchase_orders_org_ref;

ALTER TABLE purchase.purchase_order_lines
  DROP CONSTRAINT IF EXISTS ck_purchase_order_lines_runtime_refs;

ALTER TABLE purchase.purchase_orders
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_runtime_refs;

ALTER TABLE purchase.purchase_order_lines
  ALTER COLUMN item_id SET NOT NULL,
  ALTER COLUMN unit_id SET NOT NULL,
  DROP COLUMN IF EXISTS note,
  DROP COLUMN IF EXISTS item_name,
  DROP COLUMN IF EXISTS sku_code,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS line_ref;

ALTER TABLE purchase.purchase_orders
  ALTER COLUMN supplier_id SET NOT NULL,
  ALTER COLUMN warehouse_id SET NOT NULL,
  DROP COLUMN IF EXISTS rejected_by_ref,
  DROP COLUMN IF EXISTS cancelled_by_ref,
  DROP COLUMN IF EXISTS closed_by_ref,
  DROP COLUMN IF EXISTS received_by_ref,
  DROP COLUMN IF EXISTS received_by,
  DROP COLUMN IF EXISTS received_at,
  DROP COLUMN IF EXISTS partially_received_by_ref,
  DROP COLUMN IF EXISTS partially_received_by,
  DROP COLUMN IF EXISTS partially_received_at,
  DROP COLUMN IF EXISTS approved_by_ref,
  DROP COLUMN IF EXISTS submitted_by_ref,
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS note,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS supplier_name,
  DROP COLUMN IF EXISTS supplier_code,
  DROP COLUMN IF EXISTS supplier_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS po_ref;

COMMIT;
