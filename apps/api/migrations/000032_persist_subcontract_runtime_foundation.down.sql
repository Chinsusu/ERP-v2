BEGIN;

DROP INDEX IF EXISTS subcontract.ix_subcontract_payment_milestones_payable;
DROP INDEX IF EXISTS subcontract.ix_subcontract_payment_milestones_order_status;
DROP INDEX IF EXISTS subcontract.ix_subcontract_factory_claims_order_status;
DROP INDEX IF EXISTS subcontract.ix_subcontract_finished_goods_receipts_order;
DROP INDEX IF EXISTS subcontract.ix_subcontract_sample_approvals_order;
DROP INDEX IF EXISTS subcontract.ix_subcontract_material_transfer_lines_item;
DROP INDEX IF EXISTS subcontract.ix_subcontract_material_transfers_order;
DROP INDEX IF EXISTS subcontract.ix_subcontract_order_material_lines_ref;
DROP INDEX IF EXISTS subcontract.ix_subcontract_orders_org_ref;

DROP TABLE IF EXISTS subcontract.subcontract_payment_milestones;
DROP TABLE IF EXISTS subcontract.subcontract_factory_claim_evidence;
DROP TABLE IF EXISTS subcontract.subcontract_factory_claims;
DROP TABLE IF EXISTS subcontract.subcontract_finished_goods_receipt_evidence;
DROP TABLE IF EXISTS subcontract.subcontract_finished_goods_receipt_lines;
DROP TABLE IF EXISTS subcontract.subcontract_finished_goods_receipts;
DROP TABLE IF EXISTS subcontract.subcontract_sample_approval_evidence;
DROP TABLE IF EXISTS subcontract.subcontract_sample_approvals;
DROP TABLE IF EXISTS subcontract.subcontract_material_transfer_evidence;
DROP TABLE IF EXISTS subcontract.subcontract_material_transfer_lines;
DROP TABLE IF EXISTS subcontract.subcontract_material_transfers;

ALTER TABLE subcontract.subcontract_order_status_history
  DROP COLUMN IF EXISTS actor_ref;

ALTER TABLE subcontract.subcontract_order_material_lines
  DROP CONSTRAINT IF EXISTS ck_subcontract_order_material_lines_runtime_refs,
  DROP CONSTRAINT IF EXISTS uq_subcontract_order_material_lines_order_ref,
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS item_ref,
  DROP COLUMN IF EXISTS line_ref;

ALTER TABLE subcontract.subcontract_order_material_lines
  ALTER COLUMN item_id SET NOT NULL;

ALTER TABLE subcontract.subcontract_orders
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_runtime_refs,
  DROP CONSTRAINT IF EXISTS uq_subcontract_orders_org_ref,
  DROP COLUMN IF EXISTS final_payment_ready_by_ref,
  DROP COLUMN IF EXISTS rejected_factory_issue_by_ref,
  DROP COLUMN IF EXISTS qc_started_by_ref,
  DROP COLUMN IF EXISTS finished_goods_received_by_ref,
  DROP COLUMN IF EXISTS mass_production_started_by_ref,
  DROP COLUMN IF EXISTS sample_rejected_by_ref,
  DROP COLUMN IF EXISTS sample_approved_by_ref,
  DROP COLUMN IF EXISTS sample_submitted_by_ref,
  DROP COLUMN IF EXISTS materials_issued_by_ref,
  DROP COLUMN IF EXISTS deposit_recorded_by_ref,
  DROP COLUMN IF EXISTS factory_confirmed_by_ref,
  DROP COLUMN IF EXISTS closed_by_ref,
  DROP COLUMN IF EXISTS cancelled_by_ref,
  DROP COLUMN IF EXISTS approved_by_ref,
  DROP COLUMN IF EXISTS submitted_by_ref,
  DROP COLUMN IF EXISTS updated_by_ref,
  DROP COLUMN IF EXISTS created_by_ref,
  DROP COLUMN IF EXISTS finished_item_ref,
  DROP COLUMN IF EXISTS factory_ref,
  DROP COLUMN IF EXISTS order_ref,
  DROP COLUMN IF EXISTS org_ref;

ALTER TABLE subcontract.subcontract_orders
  ALTER COLUMN factory_id SET NOT NULL,
  ALTER COLUMN finished_item_id SET NOT NULL;

COMMIT;
