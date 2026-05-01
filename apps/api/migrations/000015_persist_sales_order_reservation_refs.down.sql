BEGIN;

DROP INDEX IF EXISTS inventory.ix_stock_reservations_stock_ref_active;
DROP INDEX IF EXISTS inventory.ix_stock_reservations_line_ref_active;
DROP INDEX IF EXISTS inventory.ix_stock_reservations_sales_order_ref_active;
DROP INDEX IF EXISTS inventory.uq_stock_reservations_reservation_ref;

ALTER TABLE inventory.stock_reservations
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_sales_source,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_lifecycle,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_runtime_refs,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_bin_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_source_doc_line_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_sales_order_line_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_sales_order_id;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM inventory.stock_reservations
    WHERE sales_order_id IS NULL
      OR sales_order_line_id IS NULL
      OR source_doc_id IS NULL
      OR source_doc_line_id IS NULL
      OR item_id IS NULL
      OR warehouse_id IS NULL
      OR bin_id IS NULL
      OR (status = 'released' AND released_by IS NULL)
      OR (status = 'consumed' AND consumed_by IS NULL)
  ) THEN
    RAISE EXCEPTION 'cannot roll back 000015 while text-only stock reservation rows exist';
  END IF;
END $$;

ALTER TABLE inventory.stock_reservations
  DROP COLUMN IF EXISTS consumed_by_ref,
  DROP COLUMN IF EXISTS released_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS bin_code,
  DROP COLUMN IF EXISTS bin_ref,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS batch_no,
  DROP COLUMN IF EXISTS batch_ref,
  DROP COLUMN IF EXISTS sku_code,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS source_doc_line_ref,
  DROP COLUMN IF EXISTS source_doc_ref,
  DROP COLUMN IF EXISTS sales_order_line_ref,
  DROP COLUMN IF EXISTS sales_order_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS reservation_ref;

ALTER TABLE inventory.stock_reservations
  ALTER COLUMN sales_order_id SET NOT NULL,
  ALTER COLUMN sales_order_line_id SET NOT NULL,
  ALTER COLUMN source_doc_id SET NOT NULL,
  ALTER COLUMN source_doc_line_id SET NOT NULL,
  ALTER COLUMN item_id SET NOT NULL,
  ALTER COLUMN warehouse_id SET NOT NULL,
  ALTER COLUMN bin_id SET NOT NULL,
  ADD CONSTRAINT fk_stock_reservations_sales_order_id
    FOREIGN KEY (sales_order_id) REFERENCES sales.sales_orders(id),
  ADD CONSTRAINT fk_stock_reservations_sales_order_line_id
    FOREIGN KEY (sales_order_line_id) REFERENCES sales.sales_order_lines(id),
  ADD CONSTRAINT fk_stock_reservations_source_doc_line_id
    FOREIGN KEY (source_doc_line_id) REFERENCES sales.sales_order_lines(id),
  ADD CONSTRAINT fk_stock_reservations_bin_id
    FOREIGN KEY (bin_id) REFERENCES mdm.warehouse_bins(id),
  ADD CONSTRAINT ck_stock_reservations_sales_source CHECK (
    source_doc_type = 'sales_order'
    AND source_doc_id = sales_order_id
    AND source_doc_line_id = sales_order_line_id
  ),
  ADD CONSTRAINT ck_stock_reservations_lifecycle CHECK (
    (status = 'active' AND released_at IS NULL AND released_by IS NULL AND consumed_at IS NULL AND consumed_by IS NULL)
    OR (
      status = 'released'
      AND released_at IS NOT NULL
      AND released_by IS NOT NULL
      AND consumed_at IS NULL
      AND consumed_by IS NULL
    )
    OR (status = 'consumed' AND consumed_at IS NOT NULL AND consumed_by IS NOT NULL)
  );

COMMIT;
