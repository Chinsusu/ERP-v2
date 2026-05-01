BEGIN;

DROP INDEX IF EXISTS inventory.ix_supplier_rejection_attachments_rejection;
DROP INDEX IF EXISTS inventory.ix_supplier_rejection_lines_receipt_line;
DROP INDEX IF EXISTS inventory.ix_supplier_rejections_inbound_qc;
DROP INDEX IF EXISTS inventory.ix_supplier_rejections_warehouse_status;
DROP INDEX IF EXISTS inventory.ix_supplier_rejections_supplier_status;

DROP TABLE IF EXISTS inventory.supplier_rejection_attachments;
DROP TABLE IF EXISTS inventory.supplier_rejection_lines;
DROP TABLE IF EXISTS inventory.supplier_rejections;

COMMIT;
