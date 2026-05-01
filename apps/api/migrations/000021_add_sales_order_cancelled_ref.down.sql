BEGIN;

ALTER TABLE sales.sales_orders
  DROP COLUMN IF EXISTS cancelled_by_ref;

COMMIT;
