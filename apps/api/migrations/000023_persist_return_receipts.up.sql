BEGIN;

ALTER TABLE returns.return_orders
  ADD COLUMN IF NOT EXISTS return_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS received_by_ref text,
  ADD COLUMN IF NOT EXISTS disposition text,
  ADD COLUMN IF NOT EXISTS target_location text,
  ADD COLUMN IF NOT EXISTS original_order_ref text,
  ADD COLUMN IF NOT EXISTS original_order_no text,
  ADD COLUMN IF NOT EXISTS customer_name text,
  ADD COLUMN IF NOT EXISTS scan_code text,
  ADD COLUMN IF NOT EXISTS stock_movement_ref text,
  ADD COLUMN IF NOT EXISTS stock_movement_type text,
  ADD COLUMN IF NOT EXISTS target_stock_status text;

UPDATE returns.return_orders AS return_order
SET
  return_ref = COALESCE(NULLIF(btrim(return_order.return_ref), ''), return_order.id::text),
  org_ref = COALESCE(NULLIF(btrim(return_order.org_ref), ''), return_order.org_id::text),
  warehouse_ref = COALESCE(NULLIF(btrim(return_order.warehouse_ref), ''), return_order.warehouse_id::text),
  warehouse_code = COALESCE(NULLIF(btrim(return_order.warehouse_code), ''), warehouse.code),
  received_by_ref = COALESCE(NULLIF(btrim(return_order.received_by_ref), ''), return_order.received_by::text),
  disposition = COALESCE(NULLIF(btrim(return_order.disposition), ''), return_order.initial_disposition, 'needs_inspection'),
  original_order_ref = COALESCE(NULLIF(btrim(return_order.original_order_ref), ''), return_order.sales_order_id::text),
  scan_code = COALESCE(
    NULLIF(btrim(return_order.scan_code), ''),
    NULLIF(btrim(return_order.return_code), ''),
    NULLIF(btrim(return_order.tracking_no), ''),
    return_order.return_no
  )
FROM mdm.warehouses AS warehouse
WHERE warehouse.id = return_order.warehouse_id;

UPDATE returns.return_orders AS return_order
SET
  return_ref = COALESCE(NULLIF(btrim(return_order.return_ref), ''), return_order.id::text),
  org_ref = COALESCE(NULLIF(btrim(return_order.org_ref), ''), return_order.org_id::text),
  warehouse_ref = COALESCE(NULLIF(btrim(return_order.warehouse_ref), ''), return_order.warehouse_id::text),
  received_by_ref = COALESCE(NULLIF(btrim(return_order.received_by_ref), ''), return_order.received_by::text),
  disposition = COALESCE(NULLIF(btrim(return_order.disposition), ''), return_order.initial_disposition, 'needs_inspection'),
  original_order_ref = COALESCE(NULLIF(btrim(return_order.original_order_ref), ''), return_order.sales_order_id::text),
  scan_code = COALESCE(
    NULLIF(btrim(return_order.scan_code), ''),
    NULLIF(btrim(return_order.return_code), ''),
    NULLIF(btrim(return_order.tracking_no), ''),
    return_order.return_no
  )
WHERE return_order.return_ref IS NULL
   OR return_order.org_ref IS NULL
   OR return_order.warehouse_ref IS NULL
   OR return_order.received_by_ref IS NULL
   OR return_order.disposition IS NULL
   OR return_order.scan_code IS NULL;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_status;

ALTER TABLE returns.return_orders
  ADD CONSTRAINT ck_return_orders_status CHECK (
    status IN ('received', 'pending_inspection', 'inspected', 'disposed', 'dispositioned', 'closed', 'cancelled')
  );

ALTER TABLE returns.return_orders
  ALTER COLUMN return_ref SET NOT NULL,
  ALTER COLUMN org_ref SET NOT NULL,
  ALTER COLUMN warehouse_id DROP NOT NULL;

ALTER TABLE returns.return_order_lines
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS sku_code text,
  ADD COLUMN IF NOT EXISTS product_name text,
  ADD COLUMN IF NOT EXISTS quantity numeric(18,6),
  ADD COLUMN IF NOT EXISTS condition_text text,
  ADD COLUMN IF NOT EXISTS batch_ref text,
  ADD COLUMN IF NOT EXISTS unit_ref text,
  ADD COLUMN IF NOT EXISTS uom_code varchar(20),
  ADD COLUMN IF NOT EXISTS stock_movement_ref text;

UPDATE returns.return_order_lines AS return_line
SET
  line_ref = COALESCE(NULLIF(btrim(return_line.line_ref), ''), return_line.id::text),
  item_ref = COALESCE(NULLIF(btrim(return_line.item_ref), ''), return_line.item_id::text),
  sku_code = COALESCE(NULLIF(btrim(return_line.sku_code), ''), item.sku),
  product_name = COALESCE(NULLIF(btrim(return_line.product_name), ''), item.name),
  quantity = COALESCE(return_line.quantity, return_line.returned_qty),
  condition_text = COALESCE(NULLIF(btrim(return_line.condition_text), ''), return_line.condition_note, return_line.condition_code),
  batch_ref = COALESCE(NULLIF(btrim(return_line.batch_ref), ''), return_line.batch_id::text),
  unit_ref = COALESCE(NULLIF(btrim(return_line.unit_ref), ''), return_line.unit_id::text)
FROM mdm.items AS item
WHERE item.id = return_line.item_id;

UPDATE returns.return_order_lines AS return_line
SET
  line_ref = COALESCE(NULLIF(btrim(return_line.line_ref), ''), return_line.id::text),
  item_ref = COALESCE(NULLIF(btrim(return_line.item_ref), ''), return_line.item_id::text),
  quantity = COALESCE(return_line.quantity, return_line.returned_qty),
  condition_text = COALESCE(NULLIF(btrim(return_line.condition_text), ''), return_line.condition_note, return_line.condition_code),
  batch_ref = COALESCE(NULLIF(btrim(return_line.batch_ref), ''), return_line.batch_id::text),
  unit_ref = COALESCE(NULLIF(btrim(return_line.unit_ref), ''), return_line.unit_id::text)
WHERE return_line.line_ref IS NULL
   OR return_line.item_ref IS NULL
   OR return_line.quantity IS NULL;

ALTER TABLE returns.return_order_lines
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN quantity SET NOT NULL,
  ALTER COLUMN item_id DROP NOT NULL,
  ALTER COLUMN unit_id DROP NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_return_orders_runtime_refs'
      AND conrelid = 'returns.return_orders'::regclass
  ) THEN
    ALTER TABLE returns.return_orders
      ADD CONSTRAINT ck_return_orders_runtime_refs CHECK (
        nullif(btrim(return_ref), '') IS NOT NULL
        AND nullif(btrim(org_ref), '') IS NOT NULL
        AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text, '')), '') IS NOT NULL
        AND nullif(btrim(coalesce(scan_code, tracking_no, return_code, original_order_no, return_no, '')), '') IS NOT NULL
      );
  END IF;

  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'ck_return_order_lines_runtime_refs'
      AND conrelid = 'returns.return_order_lines'::regclass
  ) THEN
    ALTER TABLE returns.return_order_lines
      ADD CONSTRAINT ck_return_order_lines_runtime_refs CHECK (
        nullif(btrim(line_ref), '') IS NOT NULL
        AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code, '')), '') IS NOT NULL
        AND quantity > 0
        AND nullif(btrim(coalesce(unit_ref, unit_id::text, uom_code, '')), '') IS NOT NULL
      );
  END IF;
END $$;

CREATE TABLE IF NOT EXISTS returns.return_inspections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  inspection_ref text NOT NULL,
  return_order_id uuid NOT NULL REFERENCES returns.return_orders(id) ON DELETE RESTRICT,
  return_ref text NOT NULL,
  condition_code text NOT NULL,
  disposition text NOT NULL,
  status text NOT NULL,
  target_location text NOT NULL,
  risk_level text NOT NULL,
  evidence_label text,
  note text,
  inspector_ref text NOT NULL,
  inspected_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_return_inspections_disposition CHECK (
    disposition IN ('reusable', 'not_reusable', 'needs_inspection')
  ),
  CONSTRAINT ck_return_inspections_status CHECK (
    status IN ('inspection_recorded', 'return_qa_hold')
  )
);

CREATE TABLE IF NOT EXISTS returns.return_disposition_actions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  action_ref text NOT NULL,
  return_order_id uuid NOT NULL REFERENCES returns.return_orders(id) ON DELETE RESTRICT,
  return_ref text NOT NULL,
  disposition text NOT NULL,
  target_location text NOT NULL,
  target_stock_status text,
  action_code text NOT NULL,
  note text,
  actor_ref text NOT NULL,
  decided_at timestamptz NOT NULL,
  stock_movement_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_return_disposition_actions_disposition CHECK (
    disposition IN ('reusable', 'not_reusable', 'needs_inspection')
  )
);

CREATE TABLE IF NOT EXISTS returns.return_attachments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  attachment_ref text NOT NULL,
  return_order_id uuid NOT NULL REFERENCES returns.return_orders(id) ON DELETE RESTRICT,
  return_ref text NOT NULL,
  inspection_ref text,
  file_name text NOT NULL,
  file_ext text,
  mime_type text,
  file_size_bytes bigint NOT NULL,
  storage_bucket text NOT NULL,
  storage_key text NOT NULL,
  uploaded_by_ref text NOT NULL,
  uploaded_at timestamptz NOT NULL,
  status text NOT NULL,
  note text,
  source text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_return_attachments_size CHECK (file_size_bytes > 0),
  CONSTRAINT ck_return_attachments_status CHECK (status IN ('active'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_return_orders_org_ref
  ON returns.return_orders(org_id, return_ref);

CREATE INDEX IF NOT EXISTS ix_return_orders_warehouse_ref_status
  ON returns.return_orders(org_id, warehouse_ref, status);

CREATE INDEX IF NOT EXISTS ix_return_orders_return_code
  ON returns.return_orders(org_id, return_code)
  WHERE return_code IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_return_order_lines_order_ref
  ON returns.return_order_lines(return_order_id, line_ref);

CREATE UNIQUE INDEX IF NOT EXISTS uq_return_inspections_org_ref
  ON returns.return_inspections(org_id, inspection_ref);

CREATE INDEX IF NOT EXISTS ix_return_inspections_return_ref
  ON returns.return_inspections(org_id, return_ref);

CREATE UNIQUE INDEX IF NOT EXISTS uq_return_disposition_actions_org_ref
  ON returns.return_disposition_actions(org_id, action_ref);

CREATE INDEX IF NOT EXISTS ix_return_disposition_actions_return_ref
  ON returns.return_disposition_actions(org_id, return_ref);

CREATE UNIQUE INDEX IF NOT EXISTS uq_return_attachments_org_ref
  ON returns.return_attachments(org_id, attachment_ref);

CREATE INDEX IF NOT EXISTS ix_return_attachments_return_ref
  ON returns.return_attachments(org_id, return_ref);

COMMIT;
