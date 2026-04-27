BEGIN;

ALTER TABLE inventory.stock_ledger
  DROP CONSTRAINT IF EXISTS ck_stock_ledger_quantity;

ALTER TABLE inventory.stock_ledger
  RENAME COLUMN quantity TO movement_qty;

ALTER TABLE inventory.stock_ledger
  ALTER COLUMN movement_qty TYPE numeric(18, 6),
  ALTER COLUMN unit_id DROP NOT NULL,
  ADD COLUMN base_uom_code varchar(20),
  ADD COLUMN source_qty numeric(18, 6),
  ADD COLUMN source_uom_code varchar(20),
  ADD COLUMN conversion_factor numeric(18, 6);

UPDATE inventory.stock_ledger AS ledger
SET
  base_uom_code = units.code,
  source_qty = ledger.movement_qty,
  source_uom_code = units.code,
  conversion_factor = 1.000000
FROM mdm.units AS units
WHERE ledger.unit_id = units.id;

ALTER TABLE inventory.stock_ledger
  ALTER COLUMN base_uom_code SET NOT NULL,
  ALTER COLUMN source_qty SET NOT NULL,
  ALTER COLUMN source_uom_code SET NOT NULL,
  ALTER COLUMN conversion_factor SET NOT NULL,
  ADD CONSTRAINT fk_stock_ledger_base_uom FOREIGN KEY (base_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT fk_stock_ledger_source_uom FOREIGN KEY (source_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT ck_stock_ledger_movement_qty CHECK (movement_qty > 0),
  ADD CONSTRAINT ck_stock_ledger_source_qty CHECK (source_qty > 0),
  ADD CONSTRAINT ck_stock_ledger_conversion_factor CHECK (conversion_factor > 0);

ALTER TABLE inventory.stock_balances
  ADD COLUMN base_uom_code varchar(20),
  ALTER COLUMN qty_on_hand TYPE numeric(18, 6),
  ALTER COLUMN qty_reserved TYPE numeric(18, 6),
  ALTER COLUMN qty_available TYPE numeric(18, 6);

UPDATE inventory.stock_balances AS balance
SET base_uom_code = units.code
FROM mdm.items AS items
JOIN mdm.units AS units ON items.base_unit_id = units.id
WHERE balance.item_id = items.id;

DROP INDEX IF EXISTS inventory.uq_stock_balances_key;

ALTER TABLE inventory.stock_balances
  ALTER COLUMN base_uom_code SET NOT NULL,
  ADD CONSTRAINT fk_stock_balances_base_uom FOREIGN KEY (base_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT uq_stock_balances_key UNIQUE NULLS NOT DISTINCT (
    org_id,
    item_id,
    batch_id,
    warehouse_id,
    bin_id,
    stock_status
  );

COMMIT;
