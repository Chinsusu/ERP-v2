BEGIN;

DROP INDEX IF EXISTS inventory.ix_stock_reservations_active;

ALTER TABLE inventory.stock_reservations
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_qty,
  DROP CONSTRAINT IF EXISTS ck_stock_reservations_status;

ALTER TABLE inventory.stock_reservations
  RENAME COLUMN quantity TO reserved_qty;

ALTER TABLE inventory.stock_reservations
  ALTER COLUMN reserved_qty TYPE numeric(18, 6),
  ADD COLUMN sales_order_id uuid,
  ADD COLUMN sales_order_line_id uuid,
  ADD COLUMN source_doc_line_id uuid,
  ADD COLUMN bin_id uuid,
  ADD COLUMN base_uom_code varchar(20),
  ADD COLUMN stock_status text NOT NULL DEFAULT 'available',
  ADD COLUMN consumed_at timestamptz,
  ADD COLUMN consumed_by uuid REFERENCES core.users(id),
  ADD COLUMN updated_at timestamptz NOT NULL DEFAULT now();

UPDATE inventory.stock_reservations AS reservation
SET
  sales_order_id = CASE
    WHEN reservation.source_doc_type = 'sales_order' THEN reservation.source_doc_id
    ELSE reservation.sales_order_id
  END,
  base_uom_code = base_uom.uom_code
FROM mdm.items AS item
JOIN mdm.units AS legacy_unit ON legacy_unit.id = item.base_unit_id
JOIN mdm.uoms AS base_uom ON base_uom.uom_code = legacy_unit.code
WHERE reservation.item_id = item.id;

ALTER TABLE inventory.stock_reservations
  ALTER COLUMN sales_order_id SET NOT NULL,
  ALTER COLUMN sales_order_line_id SET NOT NULL,
  ALTER COLUMN source_doc_line_id SET NOT NULL,
  ALTER COLUMN bin_id SET NOT NULL,
  ALTER COLUMN base_uom_code SET NOT NULL,
  ALTER COLUMN source_doc_type SET DEFAULT 'sales_order',
  ADD CONSTRAINT fk_stock_reservations_sales_order_id
    FOREIGN KEY (sales_order_id) REFERENCES sales.sales_orders(id),
  ADD CONSTRAINT fk_stock_reservations_sales_order_line_id
    FOREIGN KEY (sales_order_line_id) REFERENCES sales.sales_order_lines(id),
  ADD CONSTRAINT fk_stock_reservations_source_doc_line_id
    FOREIGN KEY (source_doc_line_id) REFERENCES sales.sales_order_lines(id),
  ADD CONSTRAINT fk_stock_reservations_bin_id
    FOREIGN KEY (bin_id) REFERENCES mdm.warehouse_bins(id),
  ADD CONSTRAINT fk_stock_reservations_base_uom_code
    FOREIGN KEY (base_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT ck_stock_reservations_qty CHECK (reserved_qty > 0),
  ADD CONSTRAINT ck_stock_reservations_status CHECK (status IN ('active', 'released', 'consumed')),
  ADD CONSTRAINT ck_stock_reservations_stock_status CHECK (
    stock_status IN ('available', 'reserved', 'qc_hold', 'return_pending', 'damaged', 'subcontract_issued')
  ),
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

CREATE INDEX ix_stock_reservations_active
  ON inventory.stock_reservations(source_doc_type, source_doc_id, source_doc_line_id)
  WHERE status = 'active';

CREATE INDEX ix_stock_reservations_sales_order_line_active
  ON inventory.stock_reservations(sales_order_line_id)
  WHERE status = 'active';

CREATE INDEX ix_stock_reservations_stock_key_active
  ON inventory.stock_reservations(org_id, item_id, batch_id, warehouse_id, bin_id, stock_status)
  WHERE status = 'active';

COMMIT;
