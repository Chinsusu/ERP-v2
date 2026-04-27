BEGIN;

ALTER TABLE inventory.stock_balances
  DROP CONSTRAINT IF EXISTS uq_stock_balances_key,
  DROP CONSTRAINT IF EXISTS fk_stock_balances_base_uom,
  DROP COLUMN IF EXISTS base_uom_code,
  ALTER COLUMN qty_on_hand TYPE numeric(18, 4),
  ALTER COLUMN qty_reserved TYPE numeric(18, 4),
  ALTER COLUMN qty_available TYPE numeric(18, 4);

CREATE UNIQUE INDEX uq_stock_balances_key
  ON inventory.stock_balances(org_id, item_id, batch_id, warehouse_id, bin_id, stock_status)
  NULLS NOT DISTINCT;

ALTER TABLE inventory.stock_ledger
  DROP CONSTRAINT IF EXISTS ck_stock_ledger_conversion_factor,
  DROP CONSTRAINT IF EXISTS ck_stock_ledger_source_qty,
  DROP CONSTRAINT IF EXISTS ck_stock_ledger_movement_qty,
  DROP CONSTRAINT IF EXISTS fk_stock_ledger_source_uom,
  DROP CONSTRAINT IF EXISTS fk_stock_ledger_base_uom,
  DROP COLUMN IF EXISTS conversion_factor,
  DROP COLUMN IF EXISTS source_uom_code,
  DROP COLUMN IF EXISTS source_qty,
  DROP COLUMN IF EXISTS base_uom_code,
  ALTER COLUMN unit_id SET NOT NULL;

ALTER TABLE inventory.stock_ledger
  RENAME COLUMN movement_qty TO quantity;

ALTER TABLE inventory.stock_ledger
  ALTER COLUMN quantity TYPE numeric(18, 4),
  ADD CONSTRAINT ck_stock_ledger_quantity CHECK (quantity > 0);

COMMIT;
