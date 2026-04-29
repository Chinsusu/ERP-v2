BEGIN;

DROP INDEX IF EXISTS subcontract.ix_subcontract_order_status_history_org_status;
DROP INDEX IF EXISTS subcontract.ix_subcontract_order_status_history_order;
DROP TABLE IF EXISTS subcontract.subcontract_order_status_history;

DROP INDEX IF EXISTS subcontract.ix_subcontract_order_material_lines_item;
DROP INDEX IF EXISTS subcontract.ix_subcontract_order_material_lines_order;
DROP TABLE IF EXISTS subcontract.subcontract_order_material_lines;

DROP INDEX IF EXISTS subcontract.ix_subcontract_orders_expected_status;
DROP INDEX IF EXISTS subcontract.ix_subcontract_orders_factory_status;

ALTER TABLE subcontract.subcontract_orders
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_business_keys,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_final_payment_status,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_status,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_amounts,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_currency_code,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_quantities,
  DROP CONSTRAINT IF EXISTS fk_subcontract_orders_base_uom_code,
  DROP CONSTRAINT IF EXISTS fk_subcontract_orders_uom_code;

UPDATE subcontract.subcontract_orders
SET status = CASE status
  WHEN 'submitted' THEN 'confirmed'
  WHEN 'approved' THEN 'confirmed'
  WHEN 'factory_confirmed' THEN 'confirmed'
  WHEN 'deposit_recorded' THEN 'deposit_paid'
  WHEN 'materials_issued_to_factory' THEN 'materials_sent'
  WHEN 'sample_submitted' THEN 'sample_pending'
  WHEN 'sample_approved' THEN 'sample_approved'
  WHEN 'sample_rejected' THEN 'sample_rejected'
  WHEN 'mass_production_started' THEN 'mass_production'
  WHEN 'finished_goods_received' THEN 'inbound_pending'
  WHEN 'qc_in_progress' THEN 'qc_checking'
  WHEN 'accepted' THEN 'accepted'
  WHEN 'rejected_with_factory_issue' THEN 'claimed'
  WHEN 'final_payment_ready' THEN 'accepted'
  WHEN 'closed' THEN 'closed'
  WHEN 'cancelled' THEN 'cancelled'
  ELSE 'draft'
END;

UPDATE subcontract.subcontract_orders
SET final_payment_status = CASE final_payment_status
  WHEN 'ready' THEN 'pending'
  ELSE final_payment_status
END;

ALTER TABLE subcontract.subcontract_orders
  DROP COLUMN IF EXISTS final_payment_ready_by,
  DROP COLUMN IF EXISTS final_payment_ready_at,
  DROP COLUMN IF EXISTS rejected_factory_issue_by,
  DROP COLUMN IF EXISTS rejected_factory_issue_at,
  DROP COLUMN IF EXISTS qc_started_by,
  DROP COLUMN IF EXISTS qc_started_at,
  DROP COLUMN IF EXISTS finished_goods_received_by,
  DROP COLUMN IF EXISTS finished_goods_received_at,
  DROP COLUMN IF EXISTS mass_production_started_by,
  DROP COLUMN IF EXISTS mass_production_started_at,
  DROP COLUMN IF EXISTS sample_rejected_by,
  DROP COLUMN IF EXISTS sample_rejected_at,
  DROP COLUMN IF EXISTS sample_approved_by,
  DROP COLUMN IF EXISTS sample_approved_at,
  DROP COLUMN IF EXISTS sample_submitted_by,
  DROP COLUMN IF EXISTS sample_submitted_at,
  DROP COLUMN IF EXISTS materials_issued_by,
  DROP COLUMN IF EXISTS materials_issued_at,
  DROP COLUMN IF EXISTS deposit_recorded_by,
  DROP COLUMN IF EXISTS deposit_recorded_at,
  DROP COLUMN IF EXISTS factory_confirmed_by,
  DROP COLUMN IF EXISTS factory_confirmed_at,
  DROP COLUMN IF EXISTS closed_by,
  DROP COLUMN IF EXISTS closed_at,
  DROP COLUMN IF EXISTS factory_issue_reason,
  DROP COLUMN IF EXISTS sample_reject_reason,
  DROP COLUMN IF EXISTS target_start_date,
  DROP COLUMN IF EXISTS claim_window_days,
  DROP COLUMN IF EXISTS sample_required,
  DROP COLUMN IF EXISTS estimated_cost_amount,
  DROP COLUMN IF EXISTS currency_code,
  DROP COLUMN IF EXISTS conversion_factor,
  DROP COLUMN IF EXISTS base_uom_code,
  DROP COLUMN IF EXISTS base_rejected_qty,
  DROP COLUMN IF EXISTS base_accepted_qty,
  DROP COLUMN IF EXISTS base_received_qty,
  DROP COLUMN IF EXISTS base_planned_qty,
  DROP COLUMN IF EXISTS rejected_qty,
  DROP COLUMN IF EXISTS accepted_qty,
  DROP COLUMN IF EXISTS received_qty,
  DROP COLUMN IF EXISTS uom_code;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN expected_receipt_date TO expected_delivery_date;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN spec_summary TO spec;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN planned_qty TO order_qty;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN finished_item_id TO product_item_id;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN factory_id TO factory_supplier_id;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN order_no TO subcontract_no;

ALTER TABLE subcontract.subcontract_orders
  ALTER COLUMN order_qty TYPE numeric(18, 4) USING round(order_qty, 4),
  ADD CONSTRAINT ck_subcontract_orders_qty CHECK (order_qty > 0),
  ADD CONSTRAINT ck_subcontract_orders_deposit CHECK (deposit_amount >= 0),
  ADD CONSTRAINT ck_subcontract_orders_payment_status CHECK (
    final_payment_status IN ('not_due', 'pending', 'paid', 'held')
  ),
  ADD CONSTRAINT ck_subcontract_orders_status CHECK (
    status IN (
      'draft',
      'confirmed',
      'deposit_paid',
      'materials_prepared',
      'materials_sent',
      'sample_pending',
      'sample_approved',
      'sample_rejected',
      'mass_production',
      'inbound_pending',
      'qc_checking',
      'accepted',
      'claimed',
      'closed',
      'cancelled'
    )
  );

CREATE INDEX ix_subcontract_orders_factory_status
  ON subcontract.subcontract_orders(factory_supplier_id, status);

COMMIT;
