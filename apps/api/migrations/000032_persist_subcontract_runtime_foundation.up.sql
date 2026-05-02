BEGIN;

ALTER TABLE subcontract.subcontract_orders
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS order_ref text,
  ADD COLUMN IF NOT EXISTS factory_ref text,
  ADD COLUMN IF NOT EXISTS finished_item_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text,
  ADD COLUMN IF NOT EXISTS submitted_by_ref text,
  ADD COLUMN IF NOT EXISTS approved_by_ref text,
  ADD COLUMN IF NOT EXISTS cancelled_by_ref text,
  ADD COLUMN IF NOT EXISTS closed_by_ref text,
  ADD COLUMN IF NOT EXISTS factory_confirmed_by_ref text,
  ADD COLUMN IF NOT EXISTS deposit_recorded_by_ref text,
  ADD COLUMN IF NOT EXISTS materials_issued_by_ref text,
  ADD COLUMN IF NOT EXISTS sample_submitted_by_ref text,
  ADD COLUMN IF NOT EXISTS sample_approved_by_ref text,
  ADD COLUMN IF NOT EXISTS sample_rejected_by_ref text,
  ADD COLUMN IF NOT EXISTS mass_production_started_by_ref text,
  ADD COLUMN IF NOT EXISTS finished_goods_received_by_ref text,
  ADD COLUMN IF NOT EXISTS qc_started_by_ref text,
  ADD COLUMN IF NOT EXISTS rejected_factory_issue_by_ref text,
  ADD COLUMN IF NOT EXISTS final_payment_ready_by_ref text;

ALTER TABLE subcontract.subcontract_orders
  ALTER COLUMN factory_id DROP NOT NULL,
  ALTER COLUMN finished_item_id DROP NOT NULL;

UPDATE subcontract.subcontract_orders
SET
  org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text),
  order_ref = COALESCE(NULLIF(btrim(order_ref), ''), id::text),
  factory_ref = COALESCE(NULLIF(btrim(factory_ref), ''), factory_id::text),
  finished_item_ref = COALESCE(NULLIF(btrim(finished_item_ref), ''), finished_item_id::text),
  created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
  updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text),
  submitted_by_ref = COALESCE(NULLIF(btrim(submitted_by_ref), ''), submitted_by::text),
  approved_by_ref = COALESCE(NULLIF(btrim(approved_by_ref), ''), approved_by::text),
  cancelled_by_ref = COALESCE(NULLIF(btrim(cancelled_by_ref), ''), cancelled_by::text),
  closed_by_ref = COALESCE(NULLIF(btrim(closed_by_ref), ''), closed_by::text),
  factory_confirmed_by_ref = COALESCE(NULLIF(btrim(factory_confirmed_by_ref), ''), factory_confirmed_by::text),
  deposit_recorded_by_ref = COALESCE(NULLIF(btrim(deposit_recorded_by_ref), ''), deposit_recorded_by::text),
  materials_issued_by_ref = COALESCE(NULLIF(btrim(materials_issued_by_ref), ''), materials_issued_by::text),
  sample_submitted_by_ref = COALESCE(NULLIF(btrim(sample_submitted_by_ref), ''), sample_submitted_by::text),
  sample_approved_by_ref = COALESCE(NULLIF(btrim(sample_approved_by_ref), ''), sample_approved_by::text),
  sample_rejected_by_ref = COALESCE(NULLIF(btrim(sample_rejected_by_ref), ''), sample_rejected_by::text),
  mass_production_started_by_ref = COALESCE(NULLIF(btrim(mass_production_started_by_ref), ''), mass_production_started_by::text),
  finished_goods_received_by_ref = COALESCE(NULLIF(btrim(finished_goods_received_by_ref), ''), finished_goods_received_by::text),
  qc_started_by_ref = COALESCE(NULLIF(btrim(qc_started_by_ref), ''), qc_started_by::text),
  rejected_factory_issue_by_ref = COALESCE(NULLIF(btrim(rejected_factory_issue_by_ref), ''), rejected_factory_issue_by::text),
  final_payment_ready_by_ref = COALESCE(NULLIF(btrim(final_payment_ready_by_ref), ''), final_payment_ready_by::text);

ALTER TABLE subcontract.subcontract_orders
  ALTER COLUMN org_ref SET NOT NULL,
  ALTER COLUMN order_ref SET NOT NULL,
  ALTER COLUMN factory_ref SET NOT NULL,
  ALTER COLUMN finished_item_ref SET NOT NULL,
  DROP CONSTRAINT IF EXISTS uq_subcontract_orders_org_ref,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_runtime_refs,
  ADD CONSTRAINT uq_subcontract_orders_org_ref UNIQUE (org_id, order_ref),
  ADD CONSTRAINT ck_subcontract_orders_runtime_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(order_ref), '') IS NOT NULL
    AND nullif(btrim(factory_ref), '') IS NOT NULL
    AND nullif(btrim(finished_item_ref), '') IS NOT NULL
  );

ALTER TABLE subcontract.subcontract_order_material_lines
  ADD COLUMN IF NOT EXISTS line_ref text,
  ADD COLUMN IF NOT EXISTS item_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text;

ALTER TABLE subcontract.subcontract_order_material_lines
  ALTER COLUMN item_id DROP NOT NULL;

UPDATE subcontract.subcontract_order_material_lines
SET
  line_ref = COALESCE(NULLIF(btrim(line_ref), ''), id::text),
  item_ref = COALESCE(NULLIF(btrim(item_ref), ''), item_id::text),
  created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
  updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text);

ALTER TABLE subcontract.subcontract_order_material_lines
  ALTER COLUMN line_ref SET NOT NULL,
  ALTER COLUMN item_ref SET NOT NULL,
  DROP CONSTRAINT IF EXISTS uq_subcontract_order_material_lines_order_ref,
  DROP CONSTRAINT IF EXISTS ck_subcontract_order_material_lines_runtime_refs,
  ADD CONSTRAINT uq_subcontract_order_material_lines_order_ref UNIQUE (subcontract_order_id, line_ref),
  ADD CONSTRAINT ck_subcontract_order_material_lines_runtime_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(item_ref), '') IS NOT NULL
  );

ALTER TABLE subcontract.subcontract_order_status_history
  ADD COLUMN IF NOT EXISTS actor_ref text;

UPDATE subcontract.subcontract_order_status_history
SET actor_ref = COALESCE(NULLIF(btrim(actor_ref), ''), actor_id::text);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_material_transfers (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  transfer_ref text NOT NULL,
  transfer_no text NOT NULL,
  subcontract_order_id uuid REFERENCES subcontract.subcontract_orders(id) ON DELETE SET NULL,
  subcontract_order_ref text NOT NULL,
  subcontract_order_no text NOT NULL,
  factory_ref text NOT NULL,
  factory_code text,
  factory_name text NOT NULL,
  source_warehouse_ref text NOT NULL,
  source_warehouse_code text,
  status text NOT NULL,
  handover_by_ref text NOT NULL,
  handover_at timestamptz NOT NULL,
  received_by_ref text NOT NULL,
  receiver_contact text,
  vehicle_no text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_material_transfers_org_ref UNIQUE (org_id, transfer_ref),
  CONSTRAINT uq_subcontract_material_transfers_org_no UNIQUE (org_id, transfer_no),
  CONSTRAINT ck_subcontract_material_transfers_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(transfer_ref), '') IS NOT NULL
    AND nullif(btrim(transfer_no), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_no), '') IS NOT NULL
    AND nullif(btrim(factory_ref), '') IS NOT NULL
    AND nullif(btrim(factory_name), '') IS NOT NULL
    AND nullif(btrim(source_warehouse_ref), '') IS NOT NULL
    AND nullif(btrim(handover_by_ref), '') IS NOT NULL
    AND nullif(btrim(received_by_ref), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND version > 0
  ),
  CONSTRAINT ck_subcontract_material_transfers_status CHECK (
    status IN ('sent_to_factory', 'partially_sent')
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_material_transfer_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  material_transfer_id uuid NOT NULL REFERENCES subcontract.subcontract_material_transfers(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL,
  order_material_line_ref text NOT NULL,
  item_ref text NOT NULL,
  sku_code text NOT NULL,
  item_name text NOT NULL,
  issue_qty numeric(18,6) NOT NULL,
  uom_code text NOT NULL,
  base_issue_qty numeric(18,6) NOT NULL,
  base_uom_code text NOT NULL,
  conversion_factor numeric(18,6) NOT NULL DEFAULT 1.000000,
  batch_ref text,
  batch_no text,
  lot_no text,
  source_bin_ref text,
  lot_trace_required boolean NOT NULL DEFAULT true,
  note text,
  CONSTRAINT uq_subcontract_material_transfer_lines_ref UNIQUE (material_transfer_id, line_ref),
  CONSTRAINT uq_subcontract_material_transfer_lines_no UNIQUE (material_transfer_id, line_no),
  CONSTRAINT ck_subcontract_material_transfer_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND line_no > 0
    AND nullif(btrim(order_material_line_ref), '') IS NOT NULL
    AND nullif(btrim(item_ref), '') IS NOT NULL
    AND nullif(btrim(sku_code), '') IS NOT NULL
    AND nullif(btrim(item_name), '') IS NOT NULL
  ),
  CONSTRAINT ck_subcontract_material_transfer_lines_qty CHECK (
    issue_qty > 0
    AND base_issue_qty > 0
    AND conversion_factor > 0
  ),
  CONSTRAINT ck_subcontract_material_transfer_lines_lot CHECK (
    lot_trace_required = false
    OR nullif(btrim(coalesce(batch_ref, '')), '') IS NOT NULL
    OR nullif(btrim(coalesce(batch_no, '')), '') IS NOT NULL
    OR nullif(btrim(coalesce(lot_no, '')), '') IS NOT NULL
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_material_transfer_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  material_transfer_id uuid NOT NULL REFERENCES subcontract.subcontract_material_transfers(id) ON DELETE CASCADE,
  evidence_ref text NOT NULL,
  evidence_type text NOT NULL,
  file_name text,
  object_key text,
  external_url text,
  note text,
  CONSTRAINT uq_subcontract_material_transfer_evidence_ref UNIQUE (material_transfer_id, evidence_ref),
  CONSTRAINT ck_subcontract_material_transfer_evidence_refs CHECK (
    nullif(btrim(evidence_ref), '') IS NOT NULL
    AND nullif(btrim(evidence_type), '') IS NOT NULL
    AND (
      nullif(btrim(coalesce(object_key, '')), '') IS NOT NULL
      OR nullif(btrim(coalesce(external_url, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_sample_approvals (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  sample_ref text NOT NULL,
  subcontract_order_id uuid REFERENCES subcontract.subcontract_orders(id) ON DELETE SET NULL,
  subcontract_order_ref text NOT NULL,
  subcontract_order_no text NOT NULL,
  sample_code text NOT NULL,
  formula_version text,
  spec_version text,
  status text NOT NULL,
  submitted_by_ref text NOT NULL,
  submitted_at timestamptz NOT NULL,
  decision_by_ref text,
  decision_at timestamptz,
  decision_reason text,
  storage_status text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_sample_approvals_org_ref UNIQUE (org_id, sample_ref),
  CONSTRAINT ck_subcontract_sample_approvals_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(sample_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_no), '') IS NOT NULL
    AND nullif(btrim(sample_code), '') IS NOT NULL
    AND nullif(btrim(submitted_by_ref), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND version > 0
  ),
  CONSTRAINT ck_subcontract_sample_approvals_status CHECK (
    status IN ('submitted', 'approved', 'rejected')
  ),
  CONSTRAINT ck_subcontract_sample_approvals_decision CHECK (
    status = 'submitted'
    OR (
      decision_at IS NOT NULL
      AND nullif(btrim(coalesce(decision_by_ref, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_sample_approval_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  sample_approval_id uuid NOT NULL REFERENCES subcontract.subcontract_sample_approvals(id) ON DELETE CASCADE,
  evidence_ref text NOT NULL,
  evidence_type text NOT NULL,
  file_name text,
  object_key text,
  external_url text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text,
  CONSTRAINT uq_subcontract_sample_approval_evidence_ref UNIQUE (sample_approval_id, evidence_ref),
  CONSTRAINT ck_subcontract_sample_approval_evidence_refs CHECK (
    nullif(btrim(evidence_ref), '') IS NOT NULL
    AND nullif(btrim(evidence_type), '') IS NOT NULL
    AND (
      nullif(btrim(coalesce(object_key, '')), '') IS NOT NULL
      OR nullif(btrim(coalesce(external_url, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_finished_goods_receipts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  receipt_ref text NOT NULL,
  receipt_no text NOT NULL,
  subcontract_order_id uuid REFERENCES subcontract.subcontract_orders(id) ON DELETE SET NULL,
  subcontract_order_ref text NOT NULL,
  subcontract_order_no text NOT NULL,
  warehouse_ref text NOT NULL,
  warehouse_code text,
  location_ref text NOT NULL,
  location_code text,
  delivery_note_no text,
  received_by_ref text NOT NULL,
  received_at timestamptz NOT NULL,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_finished_goods_receipts_org_ref UNIQUE (org_id, receipt_ref),
  CONSTRAINT uq_subcontract_finished_goods_receipts_org_no UNIQUE (org_id, receipt_no),
  CONSTRAINT ck_subcontract_finished_goods_receipts_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(receipt_ref), '') IS NOT NULL
    AND nullif(btrim(receipt_no), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_no), '') IS NOT NULL
    AND nullif(btrim(warehouse_ref), '') IS NOT NULL
    AND nullif(btrim(location_ref), '') IS NOT NULL
    AND nullif(btrim(received_by_ref), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND version > 0
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_finished_goods_receipt_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  finished_goods_receipt_id uuid NOT NULL REFERENCES subcontract.subcontract_finished_goods_receipts(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL,
  item_ref text NOT NULL,
  sku_code text NOT NULL,
  item_name text NOT NULL,
  batch_ref text,
  batch_no text,
  lot_no text,
  expiry_date date,
  receive_qty numeric(18,6) NOT NULL,
  uom_code text NOT NULL,
  base_receive_qty numeric(18,6) NOT NULL,
  base_uom_code text NOT NULL,
  conversion_factor numeric(18,6) NOT NULL DEFAULT 1.000000,
  packaging_status text,
  note text,
  CONSTRAINT uq_subcontract_finished_goods_receipt_lines_ref UNIQUE (finished_goods_receipt_id, line_ref),
  CONSTRAINT uq_subcontract_finished_goods_receipt_lines_no UNIQUE (finished_goods_receipt_id, line_no),
  CONSTRAINT ck_subcontract_finished_goods_receipt_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND line_no > 0
    AND nullif(btrim(item_ref), '') IS NOT NULL
    AND nullif(btrim(sku_code), '') IS NOT NULL
    AND nullif(btrim(item_name), '') IS NOT NULL
  ),
  CONSTRAINT ck_subcontract_finished_goods_receipt_lines_qty CHECK (
    receive_qty > 0
    AND base_receive_qty > 0
    AND conversion_factor > 0
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_finished_goods_receipt_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  finished_goods_receipt_id uuid NOT NULL REFERENCES subcontract.subcontract_finished_goods_receipts(id) ON DELETE CASCADE,
  evidence_ref text NOT NULL,
  evidence_type text NOT NULL,
  file_name text,
  object_key text,
  external_url text,
  note text,
  CONSTRAINT uq_subcontract_finished_goods_receipt_evidence_ref UNIQUE (finished_goods_receipt_id, evidence_ref),
  CONSTRAINT ck_subcontract_finished_goods_receipt_evidence_refs CHECK (
    nullif(btrim(evidence_ref), '') IS NOT NULL
    AND nullif(btrim(evidence_type), '') IS NOT NULL
    AND (
      nullif(btrim(coalesce(object_key, '')), '') IS NOT NULL
      OR nullif(btrim(coalesce(external_url, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_factory_claims (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  claim_ref text NOT NULL,
  claim_no text NOT NULL,
  subcontract_order_id uuid REFERENCES subcontract.subcontract_orders(id) ON DELETE SET NULL,
  subcontract_order_ref text NOT NULL,
  subcontract_order_no text NOT NULL,
  factory_ref text NOT NULL,
  factory_code text,
  factory_name text NOT NULL,
  receipt_ref text,
  receipt_no text,
  reason_code text,
  reason text NOT NULL,
  severity text NOT NULL,
  status text NOT NULL,
  affected_qty numeric(18,6) NOT NULL,
  uom_code text NOT NULL,
  base_affected_qty numeric(18,6) NOT NULL,
  base_uom_code text NOT NULL,
  owner_ref text NOT NULL,
  opened_by_ref text NOT NULL,
  opened_at timestamptz NOT NULL,
  due_at timestamptz NOT NULL,
  acknowledged_by_ref text,
  acknowledged_at timestamptz,
  resolved_by_ref text,
  resolved_at timestamptz,
  resolution_note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_factory_claims_org_ref UNIQUE (org_id, claim_ref),
  CONSTRAINT uq_subcontract_factory_claims_org_no UNIQUE (org_id, claim_no),
  CONSTRAINT ck_subcontract_factory_claims_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(claim_ref), '') IS NOT NULL
    AND nullif(btrim(claim_no), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_no), '') IS NOT NULL
    AND nullif(btrim(factory_ref), '') IS NOT NULL
    AND nullif(btrim(factory_name), '') IS NOT NULL
    AND nullif(btrim(reason), '') IS NOT NULL
    AND nullif(btrim(severity), '') IS NOT NULL
    AND nullif(btrim(owner_ref), '') IS NOT NULL
    AND nullif(btrim(opened_by_ref), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND version > 0
  ),
  CONSTRAINT ck_subcontract_factory_claims_status CHECK (
    status IN ('open', 'acknowledged', 'resolved', 'closed', 'cancelled')
  ),
  CONSTRAINT ck_subcontract_factory_claims_qty CHECK (
    affected_qty > 0
    AND base_affected_qty > 0
  ),
  CONSTRAINT ck_subcontract_factory_claims_sla CHECK (
    due_at >= opened_at + interval '3 days'
    AND due_at <= opened_at + interval '7 days'
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_factory_claim_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  factory_claim_id uuid NOT NULL REFERENCES subcontract.subcontract_factory_claims(id) ON DELETE CASCADE,
  evidence_ref text NOT NULL,
  evidence_type text NOT NULL,
  file_name text,
  object_key text,
  external_url text,
  note text,
  CONSTRAINT uq_subcontract_factory_claim_evidence_ref UNIQUE (factory_claim_id, evidence_ref),
  CONSTRAINT ck_subcontract_factory_claim_evidence_refs CHECK (
    nullif(btrim(evidence_ref), '') IS NOT NULL
    AND nullif(btrim(evidence_type), '') IS NOT NULL
    AND (
      nullif(btrim(coalesce(file_name, '')), '') IS NOT NULL
      OR nullif(btrim(coalesce(object_key, '')), '') IS NOT NULL
      OR nullif(btrim(coalesce(external_url, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_payment_milestones (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  milestone_ref text NOT NULL,
  milestone_no text NOT NULL,
  subcontract_order_id uuid REFERENCES subcontract.subcontract_orders(id) ON DELETE SET NULL,
  subcontract_order_ref text NOT NULL,
  subcontract_order_no text NOT NULL,
  factory_ref text NOT NULL,
  factory_code text,
  factory_name text NOT NULL,
  kind text NOT NULL,
  status text NOT NULL,
  amount numeric(18,2) NOT NULL,
  currency_code text NOT NULL DEFAULT 'VND',
  note text,
  block_reason text,
  approved_exception_ref text,
  supplier_payable_ref text,
  supplier_payable_no text,
  recorded_by_ref text,
  recorded_at timestamptz,
  ready_by_ref text,
  ready_at timestamptz,
  blocked_by_ref text,
  blocked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_payment_milestones_org_ref UNIQUE (org_id, milestone_ref),
  CONSTRAINT uq_subcontract_payment_milestones_org_no UNIQUE (org_id, milestone_no),
  CONSTRAINT ck_subcontract_payment_milestones_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(milestone_ref), '') IS NOT NULL
    AND nullif(btrim(milestone_no), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_no), '') IS NOT NULL
    AND nullif(btrim(factory_ref), '') IS NOT NULL
    AND nullif(btrim(factory_name), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND version > 0
  ),
  CONSTRAINT ck_subcontract_payment_milestones_kind CHECK (
    kind IN ('deposit', 'final_payment')
  ),
  CONSTRAINT ck_subcontract_payment_milestones_status CHECK (
    status IN ('pending', 'recorded', 'ready', 'blocked', 'cancelled')
  ),
  CONSTRAINT ck_subcontract_payment_milestones_money CHECK (
    currency_code = 'VND'
    AND amount > 0
  ),
  CONSTRAINT ck_subcontract_payment_milestones_status_refs CHECK (
    (
      status = 'recorded'
      AND nullif(btrim(coalesce(recorded_by_ref, '')), '') IS NOT NULL
      AND recorded_at IS NOT NULL
    )
    OR (
      status = 'ready'
      AND nullif(btrim(coalesce(ready_by_ref, '')), '') IS NOT NULL
      AND ready_at IS NOT NULL
    )
    OR (
      status = 'blocked'
      AND nullif(btrim(coalesce(blocked_by_ref, '')), '') IS NOT NULL
      AND blocked_at IS NOT NULL
      AND nullif(btrim(coalesce(block_reason, '')), '') IS NOT NULL
    )
    OR status IN ('pending', 'cancelled')
  )
);

CREATE INDEX IF NOT EXISTS ix_subcontract_orders_org_ref
  ON subcontract.subcontract_orders(org_id, order_ref);
CREATE INDEX IF NOT EXISTS ix_subcontract_order_material_lines_ref
  ON subcontract.subcontract_order_material_lines(subcontract_order_id, line_ref);
CREATE INDEX IF NOT EXISTS ix_subcontract_material_transfers_order
  ON subcontract.subcontract_material_transfers(org_id, subcontract_order_ref, handover_at DESC);
CREATE INDEX IF NOT EXISTS ix_subcontract_material_transfer_lines_item
  ON subcontract.subcontract_material_transfer_lines(org_id, item_ref);
CREATE INDEX IF NOT EXISTS ix_subcontract_sample_approvals_order
  ON subcontract.subcontract_sample_approvals(org_id, subcontract_order_ref, submitted_at DESC);
CREATE INDEX IF NOT EXISTS ix_subcontract_finished_goods_receipts_order
  ON subcontract.subcontract_finished_goods_receipts(org_id, subcontract_order_ref, received_at DESC);
CREATE INDEX IF NOT EXISTS ix_subcontract_factory_claims_order_status
  ON subcontract.subcontract_factory_claims(org_id, subcontract_order_ref, status, opened_at DESC);
CREATE INDEX IF NOT EXISTS ix_subcontract_payment_milestones_order_status
  ON subcontract.subcontract_payment_milestones(org_id, subcontract_order_ref, status, created_at DESC);
CREATE INDEX IF NOT EXISTS ix_subcontract_payment_milestones_payable
  ON subcontract.subcontract_payment_milestones(org_id, supplier_payable_ref)
  WHERE supplier_payable_ref IS NOT NULL;

COMMIT;
