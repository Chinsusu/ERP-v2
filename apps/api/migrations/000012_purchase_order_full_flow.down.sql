BEGIN;

DROP INDEX IF EXISTS purchase.ix_purchase_order_status_history_org_status;
DROP INDEX IF EXISTS purchase.ix_purchase_order_status_history_order;
DROP TABLE IF EXISTS purchase.purchase_order_status_history;

DROP INDEX IF EXISTS purchase.ix_purchase_order_lines_purchase_order_status;
DROP INDEX IF EXISTS purchase.ix_purchase_order_lines_item_id;

ALTER TABLE purchase.purchase_order_lines
  DROP CONSTRAINT IF EXISTS ck_purchase_order_lines_currency_code,
  DROP CONSTRAINT IF EXISTS ck_purchase_order_lines_amounts,
  DROP CONSTRAINT IF EXISTS ck_purchase_order_lines_qty,
  DROP CONSTRAINT IF EXISTS fk_purchase_order_lines_base_uom_code,
  DROP CONSTRAINT IF EXISTS fk_purchase_order_lines_uom_code,
  DROP COLUMN IF EXISTS expected_date,
  DROP COLUMN IF EXISTS line_amount,
  DROP COLUMN IF EXISTS currency_code,
  DROP COLUMN IF EXISTS conversion_factor,
  DROP COLUMN IF EXISTS base_uom_code,
  DROP COLUMN IF EXISTS base_received_qty,
  DROP COLUMN IF EXISTS base_ordered_qty,
  DROP COLUMN IF EXISTS uom_code,
  ALTER COLUMN unit_price TYPE numeric(18, 2) USING round(unit_price, 2),
  ALTER COLUMN received_qty TYPE numeric(18, 4) USING round(received_qty, 4),
  ALTER COLUMN ordered_qty TYPE numeric(18, 4) USING round(ordered_qty, 4),
  ADD CONSTRAINT ck_purchase_order_lines_qty CHECK (ordered_qty > 0 AND received_qty >= 0),
  ADD CONSTRAINT ck_purchase_order_lines_price CHECK (unit_price >= 0);

DROP INDEX IF EXISTS purchase.ix_purchase_orders_status_expected;
DROP INDEX IF EXISTS purchase.ix_purchase_orders_warehouse_status;
DROP INDEX IF EXISTS purchase.ix_purchase_orders_supplier_expected;

ALTER TABLE purchase.purchase_orders
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_business_keys,
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_amounts,
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_currency_code,
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_status,
  DROP COLUMN IF EXISTS reject_reason,
  DROP COLUMN IF EXISTS rejected_by,
  DROP COLUMN IF EXISTS rejected_at,
  DROP COLUMN IF EXISTS closed_by,
  DROP COLUMN IF EXISTS closed_at,
  DROP COLUMN IF EXISTS subtotal_amount,
  DROP COLUMN IF EXISTS warehouse_id,
  ADD CONSTRAINT ck_purchase_orders_status CHECK (
    status IN ('draft', 'submitted', 'approved', 'partially_received', 'received', 'closed', 'cancelled')
  ),
  ADD CONSTRAINT ck_purchase_orders_total CHECK (total_amount >= 0);

ALTER TABLE purchase.purchase_orders
  RENAME COLUMN expected_date TO expected_receipt_date;

ALTER TABLE purchase.purchase_orders
  ALTER COLUMN expected_receipt_date DROP NOT NULL;

ALTER TABLE purchase.purchase_orders
  RENAME COLUMN currency_code TO currency;

ALTER TABLE purchase.purchase_orders
  ALTER COLUMN currency TYPE text,
  ALTER COLUMN currency SET DEFAULT 'VND';

COMMIT;
