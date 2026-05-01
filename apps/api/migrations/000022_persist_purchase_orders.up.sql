BEGIN;

ALTER TABLE purchase.purchase_orders
  ADD COLUMN IF NOT EXISTS po_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS supplier_ref text,
  ADD COLUMN IF NOT EXISTS supplier_code text,
  ADD COLUMN IF NOT EXISTS supplier_name text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS note text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text,
  ADD COLUMN IF NOT EXISTS submitted_by_ref text,
  ADD COLUMN IF NOT EXISTS approved_by_ref text,
  ADD COLUMN IF NOT EXISTS partially_received_at timestamptz,
  ADD COLUMN IF NOT EXISTS partially_received_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS partially_received_by_ref text,
  ADD COLUMN IF NOT EXISTS received_at timestamptz,
  ADD COLUMN IF NOT EXISTS received_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS received_by_ref text,
  ADD COLUMN IF NOT EXISTS closed_by_ref text,
  ADD COLUMN IF NOT EXISTS cancelled_by_ref text,
  ADD COLUMN IF NOT EXISTS rejected_by_ref text;

UPDATE purchase.purchase_orders AS purchase_order
SET
  po_ref = COALESCE(NULLIF(btrim(purchase_order.po_ref), ''), purchase_order.id::text),
  org_ref = COALESCE(NULLIF(btrim(purchase_order.org_ref), ''), purchase_order.org_id::text),
  supplier_ref = COALESCE(NULLIF(btrim(purchase_order.supplier_ref), ''), purchase_order.supplier_id::text),
  supplier_code = COALESCE(NULLIF(btrim(purchase_order.supplier_code), ''), supplier.code),
  supplier_name = COALESCE(NULLIF(btrim(purchase_order.supplier_name), ''), supplier.name),
  warehouse_ref = COALESCE(NULLIF(btrim(purchase_order.warehouse_ref), ''), purchase_order.warehouse_id::text),
  warehouse_code = COALESCE(NULLIF(btrim(purchase_order.warehouse_code), ''), warehouse.code),
  created_by_ref = COALESCE(NULLIF(btrim(purchase_order.created_by_ref), ''), purchase_order.created_by::text),
  updated_by_ref = COALESCE(NULLIF(btrim(purchase_order.updated_by_ref), ''), purchase_order.updated_by::text),
  submitted_by_ref = COALESCE(NULLIF(btrim(purchase_order.submitted_by_ref), ''), purchase_order.submitted_by::text),
  approved_by_ref = COALESCE(NULLIF(btrim(purchase_order.approved_by_ref), ''), purchase_order.approved_by::text),
  closed_by_ref = COALESCE(NULLIF(btrim(purchase_order.closed_by_ref), ''), purchase_order.closed_by::text),
  cancelled_by_ref = COALESCE(NULLIF(btrim(purchase_order.cancelled_by_ref), ''), purchase_order.cancelled_by::text),
  rejected_by_ref = COALESCE(NULLIF(btrim(purchase_order.rejected_by_ref), ''), purchase_order.rejected_by::text)
FROM mdm.suppliers AS supplier, mdm.warehouses AS warehouse
WHERE supplier.id = purchase_order.supplier_id
  AND warehouse.id = purchase_order.warehouse_id;

UPDATE purchase.purchase_orders AS purchase_order
SET
  po_ref = COALESCE(NULLIF(btrim(purchase_order.po_ref), ''), purchase_order.id::text),
  org_ref = COALESCE(NULLIF(btrim(purchase_order.org_ref), ''), purchase_order.org_id::text),
  supplier_ref = COALESCE(NULLIF(btrim(purchase_order.supplier_ref), ''), purchase_order.supplier_id::text),
  warehouse_ref = COALESCE(NULLIF(btrim(purchase_order.warehouse_ref), ''), purchase_order.warehouse_id::text),
  created_by_ref = COALESCE(NULLIF(btrim(purchase_order.created_by_ref), ''), purchase_order.created_by::text),
  updated_by_ref = COALESCE(NULLIF(btrim(purchase_order.updated_by_ref), ''), purchase_order.updated_by::text),
  submitted_by_ref = COALESCE(NULLIF(btrim(purchase_order.submitted_by_ref), ''), purchase_order.submitted_by::text),
  approved_by_ref = COALESCE(NULLIF(btrim(purchase_order.approved_by_ref), ''), purchase_order.approved_by::text),
  closed_by_ref = COALESCE(NULLIF(btrim(purchase_order.closed_by_ref), ''), purchase_order.closed_by::text),
  cancelled_by_ref = COALESCE(NULLIF(btrim(purchase_order.cancelled_by_ref), ''), purchase_order.cancelled_by::text),
  rejected_by_ref = COALESCE(NULLIF(btrim(purchase_order.rejected_by_ref), ''), purchase_order.rejected_by::text)
WHERE purchase_order.po_ref IS NULL
   OR purchase_order.org_ref IS NULL
   OR purchase_order.supplier_ref IS NULL
   OR purchase_order.warehouse_ref IS NULL
   OR purchase_order.created_by_ref IS NULL
   OR purchase_order.updated_by_ref IS NULL;

ALTER TABLE purchase.purchase_orders
  ALTER COLUMN po_ref SET NOT NULL,
  ALTER COLUMN org_ref SET NOT NULL,
  ALTER COLUMN supplier_id DROP NOT NULL,
  ALTER COLUMN warehouse_id DROP NOT NULL;

ALTER TABLE purchase.purchase_order_lines
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS sku_code text,
  ADD COLUMN IF NOT EXISTS item_name text,
  ADD COLUMN IF NOT EXISTS note text;

UPDATE purchase.purchase_order_lines AS order_line
SET
  line_ref = COALESCE(NULLIF(btrim(order_line.line_ref), ''), order_line.id::text),
  item_ref = COALESCE(NULLIF(btrim(order_line.item_ref), ''), order_line.item_id::text),
  sku_code = COALESCE(NULLIF(btrim(order_line.sku_code), ''), item.sku),
  item_name = COALESCE(NULLIF(btrim(order_line.item_name), ''), item.name)
FROM mdm.items AS item
WHERE item.id = order_line.item_id;

UPDATE purchase.purchase_order_lines AS order_line
SET
  line_ref = COALESCE(NULLIF(btrim(order_line.line_ref), ''), order_line.id::text),
  item_ref = COALESCE(NULLIF(btrim(order_line.item_ref), ''), order_line.item_id::text)
WHERE order_line.line_ref IS NULL
   OR order_line.item_ref IS NULL;

ALTER TABLE purchase.purchase_order_lines
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN item_id DROP NOT NULL,
  ALTER COLUMN unit_id DROP NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_purchase_orders_runtime_refs'
      AND conrelid = 'purchase.purchase_orders'::regclass
  ) THEN
    ALTER TABLE purchase.purchase_orders
      ADD CONSTRAINT ck_purchase_orders_runtime_refs CHECK (
        nullif(btrim(po_ref), '') IS NOT NULL
        AND nullif(btrim(org_ref), '') IS NOT NULL
        AND nullif(btrim(coalesce(supplier_ref, supplier_id::text, '')), '') IS NOT NULL
        AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text, '')), '') IS NOT NULL
      );
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_purchase_order_lines_runtime_refs'
      AND conrelid = 'purchase.purchase_order_lines'::regclass
  ) THEN
    ALTER TABLE purchase.purchase_order_lines
      ADD CONSTRAINT ck_purchase_order_lines_runtime_refs CHECK (
        nullif(btrim(line_ref), '') IS NOT NULL
        AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code, '')), '') IS NOT NULL
        AND (unit_id IS NOT NULL OR nullif(btrim(coalesce(uom_code, '')), '') IS NOT NULL)
      );
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_purchase_orders_org_ref
  ON purchase.purchase_orders(org_id, po_ref);

CREATE INDEX IF NOT EXISTS ix_purchase_orders_supplier_ref_expected
  ON purchase.purchase_orders(org_id, supplier_ref, expected_date DESC);

CREATE INDEX IF NOT EXISTS ix_purchase_orders_warehouse_ref_status
  ON purchase.purchase_orders(org_id, warehouse_ref, status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_purchase_order_lines_order_ref
  ON purchase.purchase_order_lines(purchase_order_id, line_ref);

COMMIT;
