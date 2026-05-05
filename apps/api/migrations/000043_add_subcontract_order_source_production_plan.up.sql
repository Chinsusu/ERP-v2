ALTER TABLE subcontract.subcontract_orders
  ADD COLUMN IF NOT EXISTS source_production_plan_ref text,
  ADD COLUMN IF NOT EXISTS source_production_plan_no text;

CREATE INDEX IF NOT EXISTS idx_subcontract_orders_source_production_plan_ref
  ON subcontract.subcontract_orders (org_id, source_production_plan_ref);
