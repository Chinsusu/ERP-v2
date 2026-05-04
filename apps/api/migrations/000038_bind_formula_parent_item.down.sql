BEGIN;

DROP INDEX IF EXISTS mdm.ix_item_formulas_finished_item;
DROP INDEX IF EXISTS mdm.uq_item_formulas_one_active_finished_item;
DROP INDEX IF EXISTS mdm.uq_item_formulas_org_finished_item_version;

ALTER TABLE mdm.item_formulas
  DROP CONSTRAINT IF EXISTS fk_item_formulas_finished_item;

ALTER TABLE mdm.item_formulas
  DROP COLUMN IF EXISTS finished_item_id;

COMMIT;
