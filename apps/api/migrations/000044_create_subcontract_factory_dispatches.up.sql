BEGIN;

CREATE TABLE IF NOT EXISTS subcontract.subcontract_factory_dispatches (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  dispatch_ref text NOT NULL,
  dispatch_no text NOT NULL,
  subcontract_order_id uuid REFERENCES subcontract.subcontract_orders(id) ON DELETE SET NULL,
  subcontract_order_ref text NOT NULL,
  subcontract_order_no text NOT NULL,
  source_production_plan_ref text,
  source_production_plan_no text,
  factory_ref text NOT NULL,
  factory_code text,
  factory_name text NOT NULL,
  finished_item_ref text NOT NULL,
  finished_sku_code text NOT NULL,
  finished_item_name text NOT NULL,
  planned_qty numeric(18, 6) NOT NULL,
  uom_code text NOT NULL,
  spec_summary text,
  sample_required boolean NOT NULL DEFAULT false,
  target_start_date date,
  expected_receipt_date date,
  status text NOT NULL,
  ready_at timestamptz,
  ready_by_ref text,
  sent_at timestamptz,
  sent_by_ref text,
  responded_at timestamptz,
  response_by_ref text,
  factory_response_note text,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT uq_subcontract_factory_dispatches_org_ref UNIQUE (org_id, dispatch_ref),
  CONSTRAINT uq_subcontract_factory_dispatches_org_no UNIQUE (org_id, dispatch_no),
  CONSTRAINT ck_subcontract_factory_dispatches_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(dispatch_ref), '') IS NOT NULL
    AND nullif(btrim(dispatch_no), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_ref), '') IS NOT NULL
    AND nullif(btrim(subcontract_order_no), '') IS NOT NULL
    AND nullif(btrim(factory_ref), '') IS NOT NULL
    AND nullif(btrim(factory_name), '') IS NOT NULL
    AND nullif(btrim(finished_item_ref), '') IS NOT NULL
    AND nullif(btrim(finished_sku_code), '') IS NOT NULL
    AND nullif(btrim(finished_item_name), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND version > 0
  ),
  CONSTRAINT ck_subcontract_factory_dispatches_qty CHECK (
    planned_qty > 0
  ),
  CONSTRAINT ck_subcontract_factory_dispatches_status CHECK (
    status IN ('draft', 'ready', 'sent', 'confirmed', 'revision_requested', 'rejected', 'cancelled')
  ),
  CONSTRAINT ck_subcontract_factory_dispatches_ready CHECK (
    status NOT IN ('ready', 'sent', 'confirmed', 'revision_requested', 'rejected')
    OR (
      ready_at IS NOT NULL
      AND nullif(btrim(coalesce(ready_by_ref, '')), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_subcontract_factory_dispatches_sent CHECK (
    status NOT IN ('sent', 'confirmed', 'revision_requested', 'rejected')
    OR (
      sent_at IS NOT NULL
      AND nullif(btrim(coalesce(sent_by_ref, '')), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_subcontract_factory_dispatches_response CHECK (
    status NOT IN ('confirmed', 'revision_requested', 'rejected')
    OR (
      responded_at IS NOT NULL
      AND nullif(btrim(coalesce(response_by_ref, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_factory_dispatch_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  factory_dispatch_id uuid NOT NULL REFERENCES subcontract.subcontract_factory_dispatches(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL,
  order_material_line_ref text NOT NULL,
  item_ref text NOT NULL,
  sku_code text NOT NULL,
  item_name text NOT NULL,
  planned_qty numeric(18, 6) NOT NULL,
  uom_code text NOT NULL,
  lot_trace_required boolean NOT NULL DEFAULT true,
  note text,
  CONSTRAINT uq_subcontract_factory_dispatch_lines_ref UNIQUE (factory_dispatch_id, line_ref),
  CONSTRAINT uq_subcontract_factory_dispatch_lines_no UNIQUE (factory_dispatch_id, line_no),
  CONSTRAINT ck_subcontract_factory_dispatch_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND line_no > 0
    AND nullif(btrim(order_material_line_ref), '') IS NOT NULL
    AND nullif(btrim(item_ref), '') IS NOT NULL
    AND nullif(btrim(sku_code), '') IS NOT NULL
    AND nullif(btrim(item_name), '') IS NOT NULL
  ),
  CONSTRAINT ck_subcontract_factory_dispatch_lines_qty CHECK (
    planned_qty > 0
  )
);

CREATE TABLE IF NOT EXISTS subcontract.subcontract_factory_dispatch_evidence (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  factory_dispatch_id uuid NOT NULL REFERENCES subcontract.subcontract_factory_dispatches(id) ON DELETE CASCADE,
  evidence_ref text NOT NULL,
  evidence_type text NOT NULL,
  file_name text,
  object_key text,
  external_url text,
  note text,
  CONSTRAINT uq_subcontract_factory_dispatch_evidence_ref UNIQUE (factory_dispatch_id, evidence_ref),
  CONSTRAINT ck_subcontract_factory_dispatch_evidence_refs CHECK (
    nullif(btrim(evidence_ref), '') IS NOT NULL
    AND nullif(btrim(evidence_type), '') IS NOT NULL
    AND (
      nullif(btrim(coalesce(object_key, '')), '') IS NOT NULL
      OR nullif(btrim(coalesce(external_url, '')), '') IS NOT NULL
    )
  )
);

CREATE INDEX IF NOT EXISTS ix_subcontract_factory_dispatches_order
  ON subcontract.subcontract_factory_dispatches(org_id, subcontract_order_ref, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_subcontract_factory_dispatches_status
  ON subcontract.subcontract_factory_dispatches(org_id, status, updated_at DESC);

COMMIT;
