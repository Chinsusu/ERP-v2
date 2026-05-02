BEGIN;

ALTER TABLE subcontract.subcontract_order_material_lines
  DROP COLUMN IF EXISTS item_name,
  DROP COLUMN IF EXISTS sku_code;

ALTER TABLE subcontract.subcontract_orders
  DROP COLUMN IF EXISTS accepted_by_ref,
  DROP COLUMN IF EXISTS accepted_by,
  DROP COLUMN IF EXISTS accepted_at,
  DROP COLUMN IF EXISTS finished_item_name,
  DROP COLUMN IF EXISTS finished_sku_code,
  DROP COLUMN IF EXISTS factory_name,
  DROP COLUMN IF EXISTS factory_code;

COMMIT;
