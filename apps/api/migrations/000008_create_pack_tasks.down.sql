BEGIN;

DROP INDEX IF EXISTS shipping.ix_pack_task_lines_pick_line;
DROP INDEX IF EXISTS shipping.ix_pack_task_lines_sales_order_line;
DROP INDEX IF EXISTS shipping.ix_pack_task_lines_task;
DROP TABLE IF EXISTS shipping.pack_task_lines;

DROP INDEX IF EXISTS shipping.ix_pack_tasks_pick_task;
DROP INDEX IF EXISTS shipping.ix_pack_tasks_sales_order;
DROP INDEX IF EXISTS shipping.ix_pack_tasks_status_assignee;
DROP TABLE IF EXISTS shipping.pack_tasks;

COMMIT;
