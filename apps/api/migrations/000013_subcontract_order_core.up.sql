BEGIN;

DROP INDEX IF EXISTS subcontract.ix_subcontract_orders_factory_status;

ALTER TABLE subcontract.subcontract_orders
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_status,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_qty,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_deposit,
  DROP CONSTRAINT IF EXISTS ck_subcontract_orders_payment_status;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN subcontract_no TO order_no;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN factory_supplier_id TO factory_id;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN product_item_id TO finished_item_id;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN order_qty TO planned_qty;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN spec TO spec_summary;

ALTER TABLE subcontract.subcontract_orders
  RENAME COLUMN expected_delivery_date TO expected_receipt_date;

ALTER TABLE subcontract.subcontract_orders
  ADD COLUMN uom_code varchar(20),
  ADD COLUMN received_qty numeric(18, 6) NOT NULL DEFAULT 0,
  ADD COLUMN accepted_qty numeric(18, 6) NOT NULL DEFAULT 0,
  ADD COLUMN rejected_qty numeric(18, 6) NOT NULL DEFAULT 0,
  ADD COLUMN base_planned_qty numeric(18, 6),
  ADD COLUMN base_received_qty numeric(18, 6) NOT NULL DEFAULT 0,
  ADD COLUMN base_accepted_qty numeric(18, 6) NOT NULL DEFAULT 0,
  ADD COLUMN base_rejected_qty numeric(18, 6) NOT NULL DEFAULT 0,
  ADD COLUMN base_uom_code varchar(20),
  ADD COLUMN conversion_factor numeric(18, 6) NOT NULL DEFAULT 1.000000,
  ADD COLUMN currency_code varchar(3) NOT NULL DEFAULT 'VND',
  ADD COLUMN estimated_cost_amount numeric(18, 2) NOT NULL DEFAULT 0,
  ADD COLUMN sample_required boolean NOT NULL DEFAULT true,
  ADD COLUMN claim_window_days integer NOT NULL DEFAULT 7,
  ADD COLUMN target_start_date date,
  ADD COLUMN sample_reject_reason text,
  ADD COLUMN factory_issue_reason text,
  ADD COLUMN closed_at timestamptz,
  ADD COLUMN closed_by uuid REFERENCES core.users(id),
  ADD COLUMN factory_confirmed_at timestamptz,
  ADD COLUMN factory_confirmed_by uuid REFERENCES core.users(id),
  ADD COLUMN deposit_recorded_at timestamptz,
  ADD COLUMN deposit_recorded_by uuid REFERENCES core.users(id),
  ADD COLUMN materials_issued_at timestamptz,
  ADD COLUMN materials_issued_by uuid REFERENCES core.users(id),
  ADD COLUMN sample_submitted_at timestamptz,
  ADD COLUMN sample_submitted_by uuid REFERENCES core.users(id),
  ADD COLUMN sample_approved_at timestamptz,
  ADD COLUMN sample_approved_by uuid REFERENCES core.users(id),
  ADD COLUMN sample_rejected_at timestamptz,
  ADD COLUMN sample_rejected_by uuid REFERENCES core.users(id),
  ADD COLUMN mass_production_started_at timestamptz,
  ADD COLUMN mass_production_started_by uuid REFERENCES core.users(id),
  ADD COLUMN finished_goods_received_at timestamptz,
  ADD COLUMN finished_goods_received_by uuid REFERENCES core.users(id),
  ADD COLUMN qc_started_at timestamptz,
  ADD COLUMN qc_started_by uuid REFERENCES core.users(id),
  ADD COLUMN rejected_factory_issue_at timestamptz,
  ADD COLUMN rejected_factory_issue_by uuid REFERENCES core.users(id),
  ADD COLUMN final_payment_ready_at timestamptz,
  ADD COLUMN final_payment_ready_by uuid REFERENCES core.users(id);

ALTER TABLE subcontract.subcontract_orders
  ALTER COLUMN planned_qty TYPE numeric(18, 6) USING planned_qty::numeric(18, 6),
  ALTER COLUMN deposit_amount TYPE numeric(18, 2);

UPDATE subcontract.subcontract_orders AS subcontract_order
SET
  uom_code = COALESCE((
    SELECT units.code
    FROM mdm.units AS units
    WHERE units.id = subcontract_order.unit_id
  ), 'PCS'),
  base_uom_code = COALESCE((
    SELECT units.code
    FROM mdm.units AS units
    WHERE units.id = subcontract_order.unit_id
  ), 'PCS'),
  base_planned_qty = subcontract_order.planned_qty,
  status = CASE subcontract_order.status
    WHEN 'confirmed' THEN 'factory_confirmed'
    WHEN 'deposit_paid' THEN 'deposit_recorded'
    WHEN 'materials_prepared' THEN 'factory_confirmed'
    WHEN 'materials_sent' THEN 'materials_issued_to_factory'
    WHEN 'sample_pending' THEN 'sample_submitted'
    WHEN 'sample_approved' THEN 'sample_approved'
    WHEN 'sample_rejected' THEN 'sample_rejected'
    WHEN 'mass_production' THEN 'mass_production_started'
    WHEN 'inbound_pending' THEN 'finished_goods_received'
    WHEN 'qc_checking' THEN 'qc_in_progress'
    WHEN 'accepted' THEN 'accepted'
    WHEN 'claimed' THEN 'rejected_with_factory_issue'
    WHEN 'closed' THEN 'closed'
    WHEN 'cancelled' THEN 'cancelled'
    ELSE 'draft'
  END,
  final_payment_status = CASE subcontract_order.final_payment_status
    WHEN 'pending' THEN 'ready'
    ELSE subcontract_order.final_payment_status
  END;

UPDATE subcontract.subcontract_orders
SET
  uom_code = COALESCE(uom_code, 'PCS'),
  base_uom_code = COALESCE(base_uom_code, 'PCS'),
  base_planned_qty = COALESCE(base_planned_qty, planned_qty);

ALTER TABLE subcontract.subcontract_orders
  ALTER COLUMN finished_item_id SET NOT NULL,
  ALTER COLUMN uom_code SET NOT NULL,
  ALTER COLUMN base_planned_qty SET NOT NULL,
  ALTER COLUMN base_uom_code SET NOT NULL,
  ADD CONSTRAINT fk_subcontract_orders_uom_code FOREIGN KEY (uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT fk_subcontract_orders_base_uom_code FOREIGN KEY (base_uom_code) REFERENCES mdm.uoms(uom_code),
  ADD CONSTRAINT ck_subcontract_orders_quantities CHECK (
    planned_qty > 0
    AND received_qty >= 0
    AND accepted_qty >= 0
    AND rejected_qty >= 0
    AND base_planned_qty > 0
    AND base_received_qty >= 0
    AND base_accepted_qty >= 0
    AND base_rejected_qty >= 0
    AND conversion_factor > 0
    AND received_qty <= planned_qty
    AND accepted_qty <= received_qty
    AND rejected_qty <= received_qty
    AND accepted_qty + rejected_qty <= received_qty
    AND base_received_qty <= base_planned_qty
    AND base_accepted_qty <= base_received_qty
    AND base_rejected_qty <= base_received_qty
    AND base_accepted_qty + base_rejected_qty <= base_received_qty
  ),
  ADD CONSTRAINT ck_subcontract_orders_currency_code CHECK (currency_code = 'VND'),
  ADD CONSTRAINT ck_subcontract_orders_amounts CHECK (
    deposit_amount >= 0
    AND estimated_cost_amount >= 0
  ),
  ADD CONSTRAINT ck_subcontract_orders_status CHECK (
    status IN (
      'draft',
      'submitted',
      'approved',
      'factory_confirmed',
      'deposit_recorded',
      'materials_issued_to_factory',
      'sample_submitted',
      'sample_approved',
      'sample_rejected',
      'mass_production_started',
      'finished_goods_received',
      'qc_in_progress',
      'accepted',
      'rejected_with_factory_issue',
      'final_payment_ready',
      'closed',
      'cancelled'
    )
  ),
  ADD CONSTRAINT ck_subcontract_orders_final_payment_status CHECK (
    final_payment_status IN ('not_due', 'ready', 'paid', 'held')
  ),
  ADD CONSTRAINT ck_subcontract_orders_business_keys CHECK (
    btrim(order_no) <> ''
    AND claim_window_days BETWEEN 3 AND 7
    AND (target_start_date IS NULL OR expected_receipt_date IS NULL OR expected_receipt_date >= target_start_date)
  );

CREATE INDEX ix_subcontract_orders_factory_status
  ON subcontract.subcontract_orders(org_id, factory_id, status);

CREATE INDEX ix_subcontract_orders_expected_status
  ON subcontract.subcontract_orders(org_id, status, expected_receipt_date DESC);

CREATE TABLE subcontract.subcontract_order_material_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  subcontract_order_id uuid NOT NULL REFERENCES subcontract.subcontract_orders(id) ON DELETE RESTRICT,
  line_no integer NOT NULL,
  item_id uuid NOT NULL REFERENCES mdm.items(id),
  planned_qty numeric(18, 6) NOT NULL,
  issued_qty numeric(18, 6) NOT NULL DEFAULT 0,
  uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  base_planned_qty numeric(18, 6) NOT NULL,
  base_issued_qty numeric(18, 6) NOT NULL DEFAULT 0,
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  conversion_factor numeric(18, 6) NOT NULL DEFAULT 1.000000,
  unit_cost numeric(18, 6) NOT NULL DEFAULT 0,
  currency_code varchar(3) NOT NULL DEFAULT 'VND',
  line_cost_amount numeric(18, 2) NOT NULL DEFAULT 0,
  lot_trace_required boolean NOT NULL DEFAULT true,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by uuid REFERENCES core.users(id),
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by uuid REFERENCES core.users(id),
  CONSTRAINT uq_subcontract_order_material_lines_order_line UNIQUE (subcontract_order_id, line_no),
  CONSTRAINT ck_subcontract_order_material_lines_qty CHECK (
    planned_qty > 0
    AND issued_qty >= 0
    AND base_planned_qty > 0
    AND base_issued_qty >= 0
    AND conversion_factor > 0
    AND issued_qty <= planned_qty
    AND base_issued_qty <= base_planned_qty
  ),
  CONSTRAINT ck_subcontract_order_material_lines_amounts CHECK (
    unit_cost >= 0
    AND line_cost_amount >= 0
  ),
  CONSTRAINT ck_subcontract_order_material_lines_currency CHECK (currency_code = 'VND')
);

CREATE INDEX ix_subcontract_order_material_lines_order
  ON subcontract.subcontract_order_material_lines(subcontract_order_id);

CREATE INDEX ix_subcontract_order_material_lines_item
  ON subcontract.subcontract_order_material_lines(org_id, item_id);

CREATE TABLE subcontract.subcontract_order_status_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  subcontract_order_id uuid NOT NULL REFERENCES subcontract.subcontract_orders(id) ON DELETE RESTRICT,
  from_status text,
  to_status text NOT NULL,
  actor_id uuid REFERENCES core.users(id),
  reason text,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  changed_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_subcontract_order_status_history_from_status CHECK (
    from_status IS NULL
    OR from_status IN (
      'draft',
      'submitted',
      'approved',
      'factory_confirmed',
      'deposit_recorded',
      'materials_issued_to_factory',
      'sample_submitted',
      'sample_approved',
      'sample_rejected',
      'mass_production_started',
      'finished_goods_received',
      'qc_in_progress',
      'accepted',
      'rejected_with_factory_issue',
      'final_payment_ready',
      'closed',
      'cancelled'
    )
  ),
  CONSTRAINT ck_subcontract_order_status_history_to_status CHECK (
    to_status IN (
      'draft',
      'submitted',
      'approved',
      'factory_confirmed',
      'deposit_recorded',
      'materials_issued_to_factory',
      'sample_submitted',
      'sample_approved',
      'sample_rejected',
      'mass_production_started',
      'finished_goods_received',
      'qc_in_progress',
      'accepted',
      'rejected_with_factory_issue',
      'final_payment_ready',
      'closed',
      'cancelled'
    )
  )
);

CREATE INDEX ix_subcontract_order_status_history_order
  ON subcontract.subcontract_order_status_history(subcontract_order_id, changed_at DESC);

CREATE INDEX ix_subcontract_order_status_history_org_status
  ON subcontract.subcontract_order_status_history(org_id, to_status, changed_at DESC);

COMMIT;
