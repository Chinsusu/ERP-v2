DROP INDEX IF EXISTS subcontract.idx_subcontract_orders_source_production_plan_ref;

ALTER TABLE subcontract.subcontract_orders
  DROP COLUMN IF EXISTS source_production_plan_no,
  DROP COLUMN IF EXISTS source_production_plan_ref;
