ALTER TABLE inventory.batches
  ADD COLUMN IF NOT EXISTS batch_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS supplier_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

UPDATE inventory.batches
SET batch_ref = COALESCE(NULLIF(btrim(batch_ref), ''), id::text),
    org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text),
    item_ref = COALESCE(NULLIF(btrim(item_ref), ''), item_id::text),
    supplier_ref = COALESCE(NULLIF(btrim(supplier_ref), ''), supplier_id::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text);

CREATE UNIQUE INDEX IF NOT EXISTS uq_batches_org_ref
  ON inventory.batches(org_id, lower(batch_ref))
  WHERE nullif(btrim(batch_ref), '') IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_batches_batch_ref
  ON inventory.batches(lower(batch_ref))
  WHERE nullif(btrim(batch_ref), '') IS NOT NULL;
