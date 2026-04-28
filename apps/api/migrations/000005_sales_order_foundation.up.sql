BEGIN;

ALTER TABLE sales.sales_orders
  RENAME COLUMN currency TO currency_code;

ALTER TABLE sales.sales_orders
  DROP CONSTRAINT ck_sales_orders_status,
  ALTER COLUMN currency_code TYPE varchar(3),
  ALTER COLUMN currency_code SET DEFAULT 'VND',
  ADD COLUMN subtotal_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN discount_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN tax_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN shipping_fee_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN net_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD CONSTRAINT ck_sales_orders_status CHECK (
    status IN (
      'draft',
      'confirmed',
      'reserved',
      'picking',
      'picked',
      'packing',
      'packed',
      'waiting_handover',
      'handed_over',
      'delivered',
      'closed',
      'cancelled',
      'returned',
      'reservation_failed',
      'pick_exception',
      'pack_exception',
      'handover_exception'
    )
  ),
  ADD CONSTRAINT ck_sales_orders_currency_code CHECK (currency_code = 'VND'),
  ADD CONSTRAINT ck_sales_orders_amounts CHECK (
    subtotal_amount >= 0
    AND discount_amount >= 0
    AND tax_amount >= 0
    AND shipping_fee_amount >= 0
    AND net_amount >= 0
  ),
  ADD CONSTRAINT ck_sales_orders_business_keys CHECK (
    btrim(order_no) <> ''
    AND btrim(channel) <> ''
  );

UPDATE sales.sales_orders
SET
  subtotal_amount = total_amount,
  net_amount = total_amount
WHERE total_amount <> 0;

CREATE INDEX ix_sales_orders_customer_order_date
  ON sales.sales_orders(org_id, customer_id, order_date DESC);

CREATE INDEX ix_sales_orders_channel_status
  ON sales.sales_orders(org_id, channel, status);

ALTER TABLE sales.sales_order_lines
  DROP CONSTRAINT ck_sales_order_lines_qty,
  ALTER COLUMN ordered_qty TYPE numeric(18, 6),
  ALTER COLUMN reserved_qty TYPE numeric(18, 6),
  ALTER COLUMN shipped_qty TYPE numeric(18, 6),
  ALTER COLUMN unit_price TYPE numeric(18, 4),
  ADD COLUMN uom_code varchar(20),
  ADD COLUMN base_ordered_qty numeric(18, 6),
  ADD COLUMN base_uom_code varchar(20),
  ADD COLUMN conversion_factor numeric(18, 6),
  ADD COLUMN currency_code varchar(3) NOT NULL DEFAULT 'VND',
  ADD COLUMN line_discount_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN line_amount numeric(18, 2) NOT NULL DEFAULT 0;

UPDATE sales.sales_order_lines AS line
SET
  uom_code = ordered_units.code,
  base_ordered_qty = line.ordered_qty,
  base_uom_code = base_units.code,
  conversion_factor = 1.000000,
  line_amount = round(line.ordered_qty * line.unit_price, 2)
FROM mdm.units AS ordered_units,
  mdm.items AS item,
  mdm.units AS base_units
WHERE ordered_units.id = line.unit_id
  AND item.id = line.item_id
  AND base_units.id = item.base_unit_id;

ALTER TABLE sales.sales_order_lines
  ALTER COLUMN uom_code SET NOT NULL,
  ALTER COLUMN base_ordered_qty SET NOT NULL,
  ALTER COLUMN base_uom_code SET NOT NULL,
  ALTER COLUMN conversion_factor SET NOT NULL,
  ADD CONSTRAINT fk_sales_order_lines_uom_code FOREIGN KEY (uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT fk_sales_order_lines_base_uom_code FOREIGN KEY (base_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT ck_sales_order_lines_qty CHECK (
    ordered_qty > 0
    AND reserved_qty >= 0
    AND shipped_qty >= 0
    AND base_ordered_qty > 0
    AND conversion_factor > 0
    AND reserved_qty <= ordered_qty
    AND shipped_qty <= ordered_qty
  ),
  ADD CONSTRAINT ck_sales_order_lines_amounts CHECK (
    unit_price >= 0
    AND line_discount_amount >= 0
    AND line_amount >= 0
  ),
  ADD CONSTRAINT ck_sales_order_lines_currency_code CHECK (currency_code = 'VND');

CREATE INDEX ix_sales_order_lines_item_id
  ON sales.sales_order_lines(org_id, item_id);

COMMIT;
