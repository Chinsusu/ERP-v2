DROP INDEX IF EXISTS inventory.ix_batches_batch_ref;
DROP INDEX IF EXISTS inventory.uq_batches_org_ref;

ALTER TABLE inventory.batches
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS supplier_ref,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS batch_ref;
