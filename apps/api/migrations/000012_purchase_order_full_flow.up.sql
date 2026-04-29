BEGIN;

ALTER TABLE purchase.purchase_orders
  RENAME COLUMN currency TO currency_code;

ALTER TABLE purchase.purchase_orders
  RENAME COLUMN expected_receipt_date TO expected_date;

ALTER TABLE purchase.purchase_orders
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_status,
  DROP CONSTRAINT IF EXISTS ck_purchase_orders_total,
  ALTER COLUMN currency_code TYPE varchar(3),
  ALTER COLUMN currency_code SET DEFAULT 'VND',
  ALTER COLUMN total_amount TYPE numeric(18, 2),
  ADD COLUMN warehouse_id uuid REFERENCES mdm.warehouses(id),
  ADD COLUMN subtotal_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN closed_at timestamptz,
  ADD COLUMN closed_by uuid REFERENCES core.users(id),
  ADD COLUMN rejected_at timestamptz,
  ADD COLUMN rejected_by uuid REFERENCES core.users(id),
  ADD COLUMN reject_reason text;

UPDATE purchase.purchase_orders
SET
  expected_date = COALESCE(expected_date, order_date),
  subtotal_amount = total_amount;

UPDATE purchase.purchase_orders AS purchase_order
SET warehouse_id = (
  SELECT warehouse.id
  FROM mdm.warehouses AS warehouse
  WHERE warehouse.org_id = purchase_order.org_id
  ORDER BY warehouse.code
  LIMIT 1
)
WHERE purchase_order.warehouse_id IS NULL;

DO $$
BEGIN
  IF EXISTS (
    SELECT 1
    FROM purchase.purchase_orders
    WHERE warehouse_id IS NULL
  ) THEN
    RAISE EXCEPTION 'purchase orders require a warehouse before migration 000012 can complete';
  END IF;
END $$;

ALTER TABLE purchase.purchase_orders
  ALTER COLUMN warehouse_id SET NOT NULL,
  ALTER COLUMN expected_date SET NOT NULL,
  ALTER COLUMN subtotal_amount DROP DEFAULT,
  ADD CONSTRAINT ck_purchase_orders_status CHECK (
    status IN (
      'draft',
      'submitted',
      'approved',
      'partially_received',
      'received',
      'closed',
      'cancelled',
      'rejected'
    )
  ),
  ADD CONSTRAINT ck_purchase_orders_currency_code CHECK (currency_code = 'VND'),
  ADD CONSTRAINT ck_purchase_orders_amounts CHECK (
    subtotal_amount >= 0
    AND total_amount >= 0
  ),
  ADD CONSTRAINT ck_purchase_orders_business_keys CHECK (
    btrim(po_no) <> ''
    AND expected_date >= order_date
  );

CREATE INDEX ix_purchase_orders_supplier_expected
  ON purchase.purchase_orders(org_id, supplier_id, expected_date DESC);

CREATE INDEX ix_purchase_orders_warehouse_status
  ON purchase.purchase_orders(org_id, warehouse_id, status);

CREATE INDEX ix_purchase_orders_status_expected
  ON purchase.purchase_orders(org_id, status, expected_date DESC);

ALTER TABLE purchase.purchase_order_lines
  DROP CONSTRAINT IF EXISTS ck_purchase_order_lines_qty,
  DROP CONSTRAINT IF EXISTS ck_purchase_order_lines_price,
  ALTER COLUMN ordered_qty TYPE numeric(18, 6),
  ALTER COLUMN received_qty TYPE numeric(18, 6),
  ALTER COLUMN unit_price TYPE numeric(18, 4),
  ADD COLUMN uom_code varchar(20),
  ADD COLUMN base_ordered_qty numeric(18, 6),
  ADD COLUMN base_received_qty numeric(18, 6),
  ADD COLUMN base_uom_code varchar(20),
  ADD COLUMN conversion_factor numeric(18, 6),
  ADD COLUMN currency_code varchar(3) NOT NULL DEFAULT 'VND',
  ADD COLUMN line_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN expected_date date;

WITH line_uom AS (
  SELECT
    line.id AS purchase_order_line_id,
    ordered_units.code AS uom_code,
    base_units.code AS base_uom_code,
    COALESCE(conversion.factor, 1.000000) AS conversion_factor
  FROM purchase.purchase_order_lines AS line
  JOIN mdm.units AS ordered_units ON ordered_units.id = line.unit_id
  JOIN mdm.items AS item ON item.id = line.item_id
  JOIN mdm.units AS base_units ON base_units.id = item.base_unit_id
  LEFT JOIN LATERAL (
    SELECT uom_conversion.factor
    FROM mdm.uom_conversions AS uom_conversion
    WHERE uom_conversion.from_uom_code = ordered_units.code
      AND uom_conversion.to_uom_code = base_units.code
      AND uom_conversion.is_active
      AND (uom_conversion.item_id IS NULL OR uom_conversion.item_id = line.item_id)
      AND uom_conversion.effective_from <= CURRENT_DATE
      AND (uom_conversion.effective_to IS NULL OR uom_conversion.effective_to >= CURRENT_DATE)
    ORDER BY (uom_conversion.item_id IS NULL), uom_conversion.effective_from DESC
    LIMIT 1
  ) AS conversion ON ordered_units.code <> base_units.code
  WHERE ordered_units.code = base_units.code
    OR conversion.factor IS NOT NULL
)
UPDATE purchase.purchase_order_lines AS line
SET
  uom_code = line_uom.uom_code,
  base_ordered_qty = round(line.ordered_qty * line_uom.conversion_factor, 6),
  base_received_qty = round(line.received_qty * line_uom.conversion_factor, 6),
  base_uom_code = line_uom.base_uom_code,
  conversion_factor = line_uom.conversion_factor,
  line_amount = round(line.ordered_qty * line.unit_price, 2),
  expected_date = purchase_order.expected_date
FROM line_uom, purchase.purchase_orders AS purchase_order
WHERE line_uom.purchase_order_line_id = line.id
  AND purchase_order.id = line.purchase_order_id;

ALTER TABLE purchase.purchase_order_lines
  ALTER COLUMN uom_code SET NOT NULL,
  ALTER COLUMN base_ordered_qty SET NOT NULL,
  ALTER COLUMN base_received_qty SET NOT NULL,
  ALTER COLUMN base_uom_code SET NOT NULL,
  ALTER COLUMN conversion_factor SET NOT NULL,
  ALTER COLUMN expected_date SET NOT NULL,
  ADD CONSTRAINT fk_purchase_order_lines_uom_code FOREIGN KEY (uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT fk_purchase_order_lines_base_uom_code FOREIGN KEY (base_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT ck_purchase_order_lines_qty CHECK (
    ordered_qty > 0
    AND received_qty >= 0
    AND base_ordered_qty > 0
    AND base_received_qty >= 0
    AND conversion_factor > 0
    AND received_qty <= ordered_qty
    AND base_received_qty <= base_ordered_qty
  ),
  ADD CONSTRAINT ck_purchase_order_lines_amounts CHECK (
    unit_price >= 0
    AND line_amount >= 0
  ),
  ADD CONSTRAINT ck_purchase_order_lines_currency_code CHECK (currency_code = 'VND');

CREATE INDEX ix_purchase_order_lines_item_id
  ON purchase.purchase_order_lines(org_id, item_id);

CREATE INDEX ix_purchase_order_lines_purchase_order_status
  ON purchase.purchase_order_lines(purchase_order_id, received_qty, ordered_qty);

CREATE TABLE purchase.purchase_order_status_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  purchase_order_id uuid NOT NULL REFERENCES purchase.purchase_orders(id) ON DELETE RESTRICT,
  from_status text,
  to_status text NOT NULL,
  actor_id uuid REFERENCES core.users(id),
  reason text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  changed_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_purchase_order_status_history_from_status CHECK (
    from_status IS NULL
    OR from_status IN (
      'draft',
      'submitted',
      'approved',
      'partially_received',
      'received',
      'closed',
      'cancelled',
      'rejected'
    )
  ),
  CONSTRAINT ck_purchase_order_status_history_to_status CHECK (
    to_status IN (
      'draft',
      'submitted',
      'approved',
      'partially_received',
      'received',
      'closed',
      'cancelled',
      'rejected'
    )
  )
);

CREATE INDEX ix_purchase_order_status_history_order
  ON purchase.purchase_order_status_history(purchase_order_id, changed_at DESC);

CREATE INDEX ix_purchase_order_status_history_org_status
  ON purchase.purchase_order_status_history(org_id, to_status, changed_at DESC);

COMMIT;
