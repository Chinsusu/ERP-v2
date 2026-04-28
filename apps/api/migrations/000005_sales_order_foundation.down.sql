BEGIN;

DROP INDEX IF EXISTS sales.ix_sales_order_lines_item_id;
DROP INDEX IF EXISTS sales.ix_sales_orders_channel_status;
DROP INDEX IF EXISTS sales.ix_sales_orders_customer_order_date;

ALTER TABLE sales.sales_order_lines
  DROP CONSTRAINT IF EXISTS ck_sales_order_lines_currency_code,
  DROP CONSTRAINT IF EXISTS ck_sales_order_lines_amounts,
  DROP CONSTRAINT IF EXISTS ck_sales_order_lines_qty,
  DROP CONSTRAINT IF EXISTS fk_sales_order_lines_base_uom_code,
  DROP CONSTRAINT IF EXISTS fk_sales_order_lines_uom_code,
  DROP COLUMN IF EXISTS line_amount,
  DROP COLUMN IF EXISTS line_discount_amount,
  DROP COLUMN IF EXISTS currency_code,
  DROP COLUMN IF EXISTS conversion_factor,
  DROP COLUMN IF EXISTS base_uom_code,
  DROP COLUMN IF EXISTS base_ordered_qty,
  DROP COLUMN IF EXISTS uom_code,
  ALTER COLUMN unit_price TYPE numeric(18, 2),
  ALTER COLUMN shipped_qty TYPE numeric(18, 4),
  ALTER COLUMN reserved_qty TYPE numeric(18, 4),
  ALTER COLUMN ordered_qty TYPE numeric(18, 4),
  ADD CONSTRAINT ck_sales_order_lines_qty CHECK (
    ordered_qty > 0 AND reserved_qty >= 0 AND shipped_qty >= 0
  );

ALTER TABLE sales.sales_orders
  DROP CONSTRAINT IF EXISTS ck_sales_orders_business_keys,
  DROP CONSTRAINT IF EXISTS ck_sales_orders_amounts,
  DROP CONSTRAINT IF EXISTS ck_sales_orders_currency_code,
  DROP CONSTRAINT IF EXISTS ck_sales_orders_status,
  DROP COLUMN IF EXISTS net_amount,
  DROP COLUMN IF EXISTS shipping_fee_amount,
  DROP COLUMN IF EXISTS tax_amount,
  DROP COLUMN IF EXISTS discount_amount,
  DROP COLUMN IF EXISTS subtotal_amount,
  ADD CONSTRAINT ck_sales_orders_status CHECK (
    status IN ('draft', 'confirmed', 'reserved', 'picking', 'packed', 'handed_over', 'delivered', 'closed', 'cancelled', 'returned')
  );

ALTER TABLE sales.sales_orders
  RENAME COLUMN currency_code TO currency;

ALTER TABLE sales.sales_orders
  ALTER COLUMN currency TYPE text,
  ALTER COLUMN currency SET DEFAULT 'VND';

COMMIT;
