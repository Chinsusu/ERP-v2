BEGIN;

INSERT INTO mdm.uoms (
  uom_code,
  name_vi,
  name_en,
  uom_group,
  decimal_scale,
  allow_decimal,
  is_global_convertible,
  description
) VALUES (
  'EA',
  'Each',
  'Each',
  'COUNT',
  0,
  false,
  false,
  'Runtime compatibility UOM for existing sales order reservation fixtures'
)
ON CONFLICT (uom_code) DO UPDATE SET
  name_vi = EXCLUDED.name_vi,
  name_en = EXCLUDED.name_en,
  uom_group = EXCLUDED.uom_group,
  decimal_scale = EXCLUDED.decimal_scale,
  allow_decimal = EXCLUDED.allow_decimal,
  is_global_convertible = EXCLUDED.is_global_convertible,
  description = EXCLUDED.description,
  is_active = true,
  updated_at = now();

ALTER TABLE inventory.stock_reservations
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_sales_source,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_lifecycle,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_sales_order_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_sales_order_line_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_source_doc_line_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_bin_id;

ALTER TABLE inventory.stock_reservations
  ALTER COLUMN sales_order_id DROP NOT NULL,
  ALTER COLUMN sales_order_line_id DROP NOT NULL,
  ALTER COLUMN source_doc_id DROP NOT NULL,
  ALTER COLUMN source_doc_line_id DROP NOT NULL,
  ALTER COLUMN item_id DROP NOT NULL,
  ALTER COLUMN warehouse_id DROP NOT NULL,
  ALTER COLUMN bin_id DROP NOT NULL,
  ADD COLUMN IF NOT EXISTS reservation_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_ref text,
  ADD COLUMN IF NOT EXISTS sales_order_line_ref text,
  ADD COLUMN IF NOT EXISTS source_doc_ref text,
  ADD COLUMN IF NOT EXISTS source_doc_line_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS sku_code text,
  ADD COLUMN IF NOT EXISTS batch_ref text,
  ADD COLUMN IF NOT EXISTS batch_no text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS bin_ref text,
  ADD COLUMN IF NOT EXISTS bin_code text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS released_by_ref text,
  ADD COLUMN IF NOT EXISTS consumed_by_ref text;

UPDATE inventory.stock_reservations
SET
  reservation_ref = COALESCE(reservation_ref, id::text),
  org_ref = COALESCE(org_ref, org_id::text),
  sales_order_ref = COALESCE(sales_order_ref, sales_order_id::text),
  sales_order_line_ref = COALESCE(sales_order_line_ref, sales_order_line_id::text),
  source_doc_ref = COALESCE(source_doc_ref, source_doc_id::text),
  source_doc_line_ref = COALESCE(source_doc_line_ref, source_doc_line_id::text),
  item_ref = COALESCE(item_ref, item_id::text),
  batch_ref = COALESCE(batch_ref, batch_id::text),
  warehouse_ref = COALESCE(warehouse_ref, warehouse_id::text),
  bin_ref = COALESCE(bin_ref, bin_id::text),
  created_by_ref = COALESCE(created_by_ref, created_by::text),
  released_by_ref = COALESCE(released_by_ref, released_by::text),
  consumed_by_ref = COALESCE(consumed_by_ref, consumed_by::text);

ALTER TABLE inventory.stock_reservations
  ADD CONSTRAINT fk_stock_reservations_sales_order_id
    FOREIGN KEY (sales_order_id) REFERENCES sales.sales_orders(id),
  ADD CONSTRAINT fk_stock_reservations_sales_order_line_id
    FOREIGN KEY (sales_order_line_id) REFERENCES sales.sales_order_lines(id),
  ADD CONSTRAINT fk_stock_reservations_source_doc_line_id
    FOREIGN KEY (source_doc_line_id) REFERENCES sales.sales_order_lines(id),
  ADD CONSTRAINT fk_stock_reservations_bin_id
    FOREIGN KEY (bin_id) REFERENCES mdm.warehouse_bins(id),
  ADD CONSTRAINT ck_stock_reservations_runtime_refs CHECK (
    nullif(btrim(coalesce(reservation_ref, id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(org_ref, org_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(sales_order_ref, sales_order_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(sales_order_line_ref, sales_order_line_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(source_doc_ref, source_doc_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(source_doc_line_ref, source_doc_line_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(item_ref, item_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(batch_ref, batch_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(bin_ref, bin_id::text)), '') IS NOT NULL
  ),
  ADD CONSTRAINT ck_stock_reservations_sales_source CHECK (
    source_doc_type = 'sales_order'
    AND (
      source_doc_id = sales_order_id
      OR nullif(btrim(coalesce(source_doc_ref, '')), '') = nullif(btrim(coalesce(sales_order_ref, '')), '')
    )
    AND (
      source_doc_line_id = sales_order_line_id
      OR nullif(btrim(coalesce(source_doc_line_ref, '')), '') = nullif(btrim(coalesce(sales_order_line_ref, '')), '')
    )
  ),
  ADD CONSTRAINT ck_stock_reservations_lifecycle CHECK (
    (
      status = 'active'
      AND released_at IS NULL
      AND released_by IS NULL
      AND nullif(btrim(coalesce(released_by_ref, '')), '') IS NULL
      AND consumed_at IS NULL
      AND consumed_by IS NULL
      AND nullif(btrim(coalesce(consumed_by_ref, '')), '') IS NULL
    )
    OR (
      status = 'released'
      AND released_at IS NOT NULL
      AND (
        released_by IS NOT NULL
        OR nullif(btrim(coalesce(released_by_ref, '')), '') IS NOT NULL
      )
      AND consumed_at IS NULL
      AND consumed_by IS NULL
      AND nullif(btrim(coalesce(consumed_by_ref, '')), '') IS NULL
    )
    OR (
      status = 'consumed'
      AND consumed_at IS NOT NULL
      AND (
        consumed_by IS NOT NULL
        OR nullif(btrim(coalesce(consumed_by_ref, '')), '') IS NOT NULL
      )
    )
  );

CREATE UNIQUE INDEX IF NOT EXISTS uq_stock_reservations_reservation_ref
  ON inventory.stock_reservations(org_id, reservation_ref)
  WHERE reservation_ref IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_stock_reservations_sales_order_ref_active
  ON inventory.stock_reservations(org_id, sales_order_ref)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS ix_stock_reservations_line_ref_active
  ON inventory.stock_reservations(org_id, sales_order_line_ref)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS ix_stock_reservations_stock_ref_active
  ON inventory.stock_reservations(org_id, warehouse_ref, item_ref, batch_ref, bin_ref, stock_status)
  WHERE status = 'active';

COMMIT;
