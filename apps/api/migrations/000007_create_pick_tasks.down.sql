BEGIN;

DROP INDEX IF EXISTS shipping.ix_pick_task_lines_reservation;
DROP INDEX IF EXISTS shipping.ix_pick_task_lines_sales_order_line;
DROP INDEX IF EXISTS shipping.ix_pick_task_lines_task;
DROP TABLE IF EXISTS shipping.pick_task_lines;

DROP INDEX IF EXISTS shipping.ix_pick_tasks_sales_order;
DROP INDEX IF EXISTS shipping.ix_pick_tasks_status_assignee;
DROP TABLE IF EXISTS shipping.pick_tasks;

COMMIT;
