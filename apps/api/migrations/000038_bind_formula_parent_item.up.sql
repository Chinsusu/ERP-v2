BEGIN;

ALTER TABLE mdm.item_formulas
  ADD COLUMN IF NOT EXISTS finished_item_id uuid;

UPDATE mdm.item_formulas AS formula
SET finished_item_id = item.id
FROM mdm.items AS item
WHERE formula.finished_item_id IS NULL
  AND item.org_id = formula.org_id
  AND (
    item.id::text = formula.finished_item_ref
    OR lower(COALESCE(item.item_ref, item.id::text)) = lower(formula.finished_item_ref)
    OR lower(item.sku) = lower(formula.finished_sku)
    OR lower(COALESCE(item.item_code, '')) = lower(formula.finished_item_ref)
  )
  AND item.item_type IN ('finished_good', 'semi_finished')
  AND item.status = 'active';

ALTER TABLE mdm.item_formulas
  ALTER COLUMN finished_item_id SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'fk_item_formulas_finished_item'
      AND conrelid = 'mdm.item_formulas'::regclass
  ) THEN
    ALTER TABLE mdm.item_formulas
      ADD CONSTRAINT fk_item_formulas_finished_item
      FOREIGN KEY (finished_item_id)
      REFERENCES mdm.items(id);
  END IF;
END $$;

CREATE UNIQUE INDEX IF NOT EXISTS uq_item_formulas_org_finished_item_version
  ON mdm.item_formulas(org_id, finished_item_id, lower(formula_version));

CREATE UNIQUE INDEX IF NOT EXISTS uq_item_formulas_one_active_finished_item
  ON mdm.item_formulas(org_id, finished_item_id)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS ix_item_formulas_finished_item
  ON mdm.item_formulas(org_id, finished_item_id, status, updated_at DESC);

COMMIT;
