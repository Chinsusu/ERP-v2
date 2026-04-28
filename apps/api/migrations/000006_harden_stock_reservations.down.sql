BEGIN;

DROP INDEX IF EXISTS inventory.ix_stock_reservations_stock_key_active;
DROP INDEX IF EXISTS inventory.ix_stock_reservations_sales_order_line_active;
DROP INDEX IF EXISTS inventory.ix_stock_reservations_active;

ALTER TABLE inventory.stock_reservations
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_lifecycle,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_sales_source,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_stock_status,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_status,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_qty,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_base_uom_code,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_bin_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_source_doc_line_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_sales_order_line_id,
  DROP CONSTRAINT IF EXISTS fk_stock_reservations_sales_order_id,
  ALTER COLUMN source_doc_type DROP DEFAULT,
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS consumed_by,
  DROP COLUMN IF EXISTS consumed_at,
  DROP COLUMN IF EXISTS stock_status,
  DROP COLUMN IF EXISTS base_uom_code,
  DROP COLUMN IF EXISTS bin_id,
  DROP COLUMN IF EXISTS source_doc_line_id,
  DROP COLUMN IF EXISTS sales_order_line_id,
  DROP COLUMN IF EXISTS sales_order_id;

ALTER TABLE inventory.stock_reservations
  ALTER COLUMN reserved_qty TYPE numeric(18, 4);

ALTER TABLE inventory.stock_reservations
  RENAME COLUMN reserved_qty TO quantity;

ALTER TABLE inventory.stock_reservations
  ADD CONSTRAINT ck_stock_reservations_qty CHECK (quantity > 0),
  ADD CONSTRAINT ck_stock_reservations_status CHECK (status IN ('active', 'released', 'consumed', 'expired', 'cancelled'));

CREATE INDEX ix_stock_reservations_active
  ON inventory.stock_reservations(source_doc_type, source_doc_id)
  WHERE status = 'active';

COMMIT;
