BEGIN;

ALTER TABLE sales.sales_orders
  ADD COLUMN IF NOT EXISTS cancelled_by_ref text;

UPDATE sales.sales_orders
SET cancelled_by_ref = COALESCE(NULLIF(btrim(cancelled_by_ref), ''), cancelled_by::text)
WHERE cancelled_by_ref IS NULL
  AND cancelled_by IS NOT NULL;

COMMIT;
