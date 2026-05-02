BEGIN;

ALTER TABLE subcontract.subcontract_orders
  ADD COLUMN IF NOT EXISTS factory_code text,
  ADD COLUMN IF NOT EXISTS factory_name text,
  ADD COLUMN IF NOT EXISTS finished_sku_code text,
  ADD COLUMN IF NOT EXISTS finished_item_name text,
  ADD COLUMN IF NOT EXISTS accepted_at timestamptz,
  ADD COLUMN IF NOT EXISTS accepted_by uuid REFERENCES core.users(id),
  ADD COLUMN IF NOT EXISTS accepted_by_ref text;

UPDATE subcontract.subcontract_orders AS subcontract_order
SET
  factory_code = COALESCE(NULLIF(btrim(subcontract_order.factory_code), ''), supplier.code),
  factory_name = COALESCE(NULLIF(btrim(subcontract_order.factory_name), ''), supplier.name)
FROM mdm.suppliers AS supplier
WHERE supplier.id = subcontract_order.factory_id;

UPDATE subcontract.subcontract_orders AS subcontract_order
SET
  finished_sku_code = COALESCE(NULLIF(btrim(subcontract_order.finished_sku_code), ''), item.sku),
  finished_item_name = COALESCE(NULLIF(btrim(subcontract_order.finished_item_name), ''), item.name)
FROM mdm.items AS item
WHERE item.id = subcontract_order.finished_item_id;

UPDATE subcontract.subcontract_orders AS subcontract_order
SET accepted_by_ref = COALESCE(NULLIF(btrim(subcontract_order.accepted_by_ref), ''), subcontract_order.accepted_by::text);

ALTER TABLE subcontract.subcontract_order_material_lines
  ADD COLUMN IF NOT EXISTS sku_code text,
  ADD COLUMN IF NOT EXISTS item_name text;

UPDATE subcontract.subcontract_order_material_lines AS material_line
SET
  sku_code = COALESCE(NULLIF(btrim(material_line.sku_code), ''), item.sku),
  item_name = COALESCE(NULLIF(btrim(material_line.item_name), ''), item.name)
FROM mdm.items AS item
WHERE item.id = material_line.item_id;

COMMIT;
