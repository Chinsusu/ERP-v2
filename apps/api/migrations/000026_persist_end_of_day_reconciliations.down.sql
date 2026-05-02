BEGIN;

DROP INDEX IF EXISTS inventory.ix_warehouse_daily_closing_lines_closing;
DROP INDEX IF EXISTS inventory.ix_warehouse_daily_closing_checklist_closing;
DROP INDEX IF EXISTS inventory.ix_warehouse_daily_closings_filters;
DROP INDEX IF EXISTS inventory.uq_warehouse_daily_closings_org_ref;

DROP TABLE IF EXISTS inventory.warehouse_daily_closing_lines;
DROP TABLE IF EXISTS inventory.warehouse_daily_closing_checklist;

ALTER TABLE inventory.warehouse_daily_closings
  DROP COLUMN IF EXISTS stock_count_session_count,
  DROP COLUMN IF EXISTS stock_movement_count,
  DROP COLUMN IF EXISTS return_order_count,
  DROP COLUMN IF EXISTS handover_order_count,
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS closed_by_ref,
  DROP COLUMN IF EXISTS owner_ref,
  DROP COLUMN IF EXISTS warehouse_code,
  DROP COLUMN IF EXISTS warehouse_ref,
  DROP COLUMN IF EXISTS org_ref,
  DROP COLUMN IF EXISTS closing_ref;

COMMIT;
