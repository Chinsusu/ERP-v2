BEGIN;

DROP INDEX IF EXISTS subcontract.ix_subcontract_factory_dispatches_status;
DROP INDEX IF EXISTS subcontract.ix_subcontract_factory_dispatches_order;

DROP TABLE IF EXISTS subcontract.subcontract_factory_dispatch_evidence;
DROP TABLE IF EXISTS subcontract.subcontract_factory_dispatch_lines;
DROP TABLE IF EXISTS subcontract.subcontract_factory_dispatches;

COMMIT;
