BEGIN;

ALTER TABLE sales.sales_orders
  ADD COLUMN IF NOT EXISTS order_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS customer_ref text,
  ADD COLUMN IF NOT EXISTS customer_code text,
  ADD COLUMN IF NOT EXISTS customer_name text,
  ADD COLUMN IF NOT EXISTS warehouse_id uuid REFERENCES mdm.warehouses(id),
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS note text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text,
  ADD COLUMN IF NOT EXISTS confirmed_at timestamptz,
  ADD COLUMN IF NOT EXISTS confirmed_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS confirmed_by_ref text,
  ADD COLUMN IF NOT EXISTS reserved_at timestamptz,
  ADD COLUMN IF NOT EXISTS reserved_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS reserved_by_ref text,
  ADD COLUMN IF NOT EXISTS picking_started_at timestamptz,
  ADD COLUMN IF NOT EXISTS picking_started_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS picking_started_by_ref text,
  ADD COLUMN IF NOT EXISTS picked_at timestamptz,
  ADD COLUMN IF NOT EXISTS picked_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS picked_by_ref text,
  ADD COLUMN IF NOT EXISTS packing_started_at timestamptz,
  ADD COLUMN IF NOT EXISTS packing_started_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS packing_started_by_ref text,
  ADD COLUMN IF NOT EXISTS packed_at timestamptz,
  ADD COLUMN IF NOT EXISTS packed_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS packed_by_ref text,
  ADD COLUMN IF NOT EXISTS waiting_handover_at timestamptz,
  ADD COLUMN IF NOT EXISTS waiting_handover_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS waiting_handover_by_ref text,
  ADD COLUMN IF NOT EXISTS handed_over_at timestamptz,
  ADD COLUMN IF NOT EXISTS handed_over_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS handed_over_by_ref text,
  ADD COLUMN IF NOT EXISTS closed_at timestamptz,
  ADD COLUMN IF NOT EXISTS closed_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS closed_by_ref text,
  ADD COLUMN IF NOT EXISTS exception_at timestamptz,
  ADD COLUMN IF NOT EXISTS exception_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS exception_by_ref text;

UPDATE sales.sales_orders AS order_header
SET
  order_ref = COALESCE(NULLIF(btrim(order_header.order_ref), ''), order_header.id::text),
  org_ref = COALESCE(NULLIF(btrim(order_header.org_ref), ''), order_header.org_id::text),
  customer_ref = COALESCE(NULLIF(btrim(order_header.customer_ref), ''), order_header.customer_id::text),
  customer_code = COALESCE(NULLIF(btrim(order_header.customer_code), ''), customer.code),
  customer_name = COALESCE(NULLIF(btrim(order_header.customer_name), ''), customer.name),
  created_by_ref = COALESCE(NULLIF(btrim(order_header.created_by_ref), ''), order_header.created_by::text),
  updated_by_ref = COALESCE(NULLIF(btrim(order_header.updated_by_ref), ''), order_header.updated_by::text)
FROM mdm.customers AS customer
WHERE customer.id = order_header.customer_id;

UPDATE sales.sales_orders AS order_header
SET
  order_ref = COALESCE(NULLIF(btrim(order_header.order_ref), ''), order_header.id::text),
  org_ref = COALESCE(NULLIF(btrim(order_header.org_ref), ''), order_header.org_id::text),
  customer_ref = COALESCE(NULLIF(btrim(order_header.customer_ref), ''), order_header.customer_id::text),
  created_by_ref = COALESCE(NULLIF(btrim(order_header.created_by_ref), ''), order_header.created_by::text),
  updated_by_ref = COALESCE(NULLIF(btrim(order_header.updated_by_ref), ''), order_header.updated_by::text)
WHERE order_header.order_ref IS NULL
   OR order_header.org_ref IS NULL
   OR order_header.customer_ref IS NULL
   OR order_header.created_by_ref IS NULL
   OR order_header.updated_by_ref IS NULL;

ALTER TABLE sales.sales_orders
  ALTER COLUMN order_ref SET NOT NULL,
  ALTER COLUMN org_ref SET NOT NULL;

ALTER TABLE sales.sales_order_lines
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS sku_code text,
  ADD COLUMN IF NOT EXISTS item_name text,
  ADD COLUMN IF NOT EXISTS batch_id uuid REFERENCES inventory.batches(id),
  ADD COLUMN IF NOT EXISTS batch_ref text,
  ADD COLUMN IF NOT EXISTS batch_no text;

UPDATE sales.sales_order_lines AS order_line
SET
  line_ref = COALESCE(NULLIF(btrim(order_line.line_ref), ''), order_line.id::text),
  item_ref = COALESCE(NULLIF(btrim(order_line.item_ref), ''), order_line.item_id::text),
  sku_code = COALESCE(NULLIF(btrim(order_line.sku_code), ''), item.sku),
  item_name = COALESCE(NULLIF(btrim(order_line.item_name), ''), item.name)
FROM mdm.items AS item
WHERE item.id = order_line.item_id;

UPDATE sales.sales_order_lines AS order_line
SET
  line_ref = COALESCE(NULLIF(btrim(order_line.line_ref), ''), order_line.id::text),
  item_ref = COALESCE(NULLIF(btrim(order_line.item_ref), ''), order_line.item_id::text)
WHERE order_line.line_ref IS NULL
   OR order_line.item_ref IS NULL;

ALTER TABLE sales.sales_order_lines
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN item_id DROP NOT NULL,
  ALTER COLUMN unit_id DROP NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_sales_orders_runtime_refs'
      AND conrelid = 'sales.sales_orders'::regclass
  ) THEN
    ALTER TABLE sales.sales_orders
      ADD CONSTRAINT ck_sales_orders_runtime_refs CHECK (
        nullif(btrim(order_ref), '') IS NOT NULL
        AND nullif(btrim(org_ref), '') IS NOT NULL
        AND nullif(btrim(coalesce(customer_ref, customer_id::text, '')), '') IS NOT NULL
      );
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_sales_order_lines_runtime_refs'
      AND conrelid = 'sales.sales_order_lines'::regclass
  ) THEN
    ALTER TABLE sales.sales_order_lines
      ADD CONSTRAINT ck_sales_order_lines_runtime_refs CHECK (
        nullif(btrim(line_ref), '') IS NOT NULL
        AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code, '')), '') IS NOT NULL
        AND (unit_id IS NOT NULL OR nullif(btrim(coalesce(uom_code, '')), '') IS NOT NULL)
      );
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_sales_orders_org_ref
  ON sales.sales_orders(org_id, order_ref);

CREATE INDEX IF NOT EXISTS ix_sales_orders_customer_ref_order_date
  ON sales.sales_orders(org_id, customer_ref, order_date DESC);

CREATE INDEX IF NOT EXISTS ix_sales_orders_warehouse_ref_status
  ON sales.sales_orders(org_id, warehouse_ref, status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_sales_order_lines_order_ref
  ON sales.sales_order_lines(sales_order_id, line_ref);

COMMIT;
