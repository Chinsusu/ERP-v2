BEGIN;

DROP INDEX IF EXISTS returns.ix_return_order_lines_item_batch;
DROP INDEX IF EXISTS returns.ix_return_order_lines_source_order_line;
DROP INDEX IF EXISTS returns.ix_return_orders_customer_created;
DROP INDEX IF EXISTS returns.ix_return_orders_tracking_no;

ALTER TABLE returns.return_order_lines
  DROP CONSTRAINT IF EXISTS ck_return_order_lines_condition_code;

ALTER TABLE returns.return_order_lines
  ALTER COLUMN returned_qty TYPE numeric(18,4) USING returned_qty::numeric(18,4);

ALTER TABLE returns.return_order_lines
  DROP COLUMN IF EXISTS condition_code,
  DROP COLUMN IF EXISTS return_reason_code,
  DROP COLUMN IF EXISTS source_order_line_id;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_receiving_identity;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_initial_disposition;

ALTER TABLE returns.return_orders
  DROP CONSTRAINT IF EXISTS ck_return_orders_source;

ALTER TABLE returns.return_orders
  DROP COLUMN IF EXISTS investigation_note,
  DROP COLUMN IF EXISTS unknown_case,
  DROP COLUMN IF EXISTS initial_disposition,
  DROP COLUMN IF EXISTS package_condition,
  DROP COLUMN IF EXISTS return_reason_code,
  DROP COLUMN IF EXISTS return_code,
  DROP COLUMN IF EXISTS tracking_no,
  DROP COLUMN IF EXISTS source,
  DROP COLUMN IF EXISTS carrier_id,
  DROP COLUMN IF EXISTS customer_id;

COMMIT;
