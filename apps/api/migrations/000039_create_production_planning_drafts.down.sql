BEGIN;

DROP INDEX IF EXISTS subcontract.ix_purchase_request_drafts_source_plan;
DROP TABLE IF EXISTS subcontract.purchase_request_drafts;

DROP INDEX IF EXISTS subcontract.ix_production_plans_output_item;
DROP INDEX IF EXISTS subcontract.ix_production_plans_org_status;
DROP TABLE IF EXISTS subcontract.production_plans;

COMMIT;
