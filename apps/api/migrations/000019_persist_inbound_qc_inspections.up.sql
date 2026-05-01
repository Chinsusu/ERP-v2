BEGIN;

CREATE TABLE IF NOT EXISTS qc.inbound_qc_inspections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  inspection_ref text NOT NULL,
  org_ref text,
  goods_receipt_id uuid,
  goods_receipt_ref text NOT NULL,
  goods_receipt_no text NOT NULL,
  goods_receipt_line_id uuid,
  goods_receipt_line_ref text NOT NULL,
  purchase_order_id uuid,
  purchase_order_ref text,
  purchase_order_line_id uuid,
  purchase_order_line_ref text,
  item_id uuid,
  item_ref text,
  sku_code text NOT NULL,
  item_name text,
  batch_id uuid,
  batch_ref text,
  batch_no text NOT NULL,
  lot_no text NOT NULL,
  expiry_date date NOT NULL,
  warehouse_id uuid,
  warehouse_ref text,
  location_id uuid,
  location_ref text,
  quantity numeric(18, 6) NOT NULL,
  uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  inspector_id uuid REFERENCES core.users(id),
  inspector_ref text NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  result text,
  passed_qty numeric(18, 6) NOT NULL DEFAULT 0,
  failed_qty numeric(18, 6) NOT NULL DEFAULT 0,
  hold_qty numeric(18, 6) NOT NULL DEFAULT 0,
  reason text,
  note text,
  created_by uuid REFERENCES core.users(id),
  created_by_ref text NOT NULL,
  updated_by uuid REFERENCES core.users(id),
  updated_by_ref text,
  started_at timestamptz,
  started_by uuid REFERENCES core.users(id),
  started_by_ref text,
  decided_at timestamptz,
  decided_by uuid REFERENCES core.users(id),
  decided_by_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_inbound_qc_inspections_org_ref UNIQUE (org_id, inspection_ref),
  CONSTRAINT ck_inbound_qc_inspections_status CHECK (
    status IN ('pending', 'in_progress', 'completed', 'cancelled')
  ),
  CONSTRAINT ck_inbound_qc_inspections_result CHECK (
    result IS NULL OR result IN ('pass', 'fail', 'hold', 'partial')
  ),
  CONSTRAINT ck_inbound_qc_inspections_qty CHECK (
    quantity > 0
    AND passed_qty >= 0
    AND failed_qty >= 0
    AND hold_qty >= 0
  ),
  CONSTRAINT ck_inbound_qc_inspections_refs CHECK (
    nullif(btrim(inspection_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(org_ref, org_id::text)), '') IS NOT NULL
    AND nullif(btrim(goods_receipt_ref), '') IS NOT NULL
    AND nullif(btrim(goods_receipt_line_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code)), '') IS NOT NULL
    AND nullif(btrim(batch_no), '') IS NOT NULL
    AND nullif(btrim(lot_no), '') IS NOT NULL
    AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(location_ref, location_id::text)), '') IS NOT NULL
    AND nullif(btrim(inspector_ref), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_inbound_qc_inspections_lifecycle CHECK (
    status = 'pending'
    OR (
      status = 'in_progress'
      AND started_at IS NOT NULL
      AND (started_by IS NOT NULL OR nullif(btrim(coalesce(started_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'completed'
      AND result IS NOT NULL
      AND decided_at IS NOT NULL
      AND (decided_by IS NOT NULL OR nullif(btrim(coalesce(decided_by_ref, '')), '') IS NOT NULL)
      AND passed_qty + failed_qty + hold_qty = quantity
    )
    OR (
      status = 'cancelled'
      AND nullif(btrim(coalesce(reason, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS qc.inbound_qc_checklist_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  inspection_id uuid NOT NULL REFERENCES qc.inbound_qc_inspections(id) ON DELETE CASCADE,
  checklist_ref text NOT NULL,
  checklist_no integer NOT NULL DEFAULT 1,
  code text NOT NULL,
  label text NOT NULL,
  required boolean NOT NULL DEFAULT false,
  status text NOT NULL DEFAULT 'pending',
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_inbound_qc_checklist_ref UNIQUE (org_id, inspection_id, checklist_ref),
  CONSTRAINT ck_inbound_qc_checklist_status CHECK (
    status IN ('pending', 'pass', 'fail', 'not_applicable')
  ),
  CONSTRAINT ck_inbound_qc_checklist_refs CHECK (
    nullif(btrim(checklist_ref), '') IS NOT NULL
    AND nullif(btrim(code), '') IS NOT NULL
    AND nullif(btrim(label), '') IS NOT NULL
  )
);

CREATE INDEX IF NOT EXISTS ix_inbound_qc_inspections_status
  ON qc.inbound_qc_inspections(org_id, status, updated_at DESC);

CREATE INDEX IF NOT EXISTS ix_inbound_qc_inspections_receipt_line
  ON qc.inbound_qc_inspections(goods_receipt_ref, goods_receipt_line_ref);

CREATE INDEX IF NOT EXISTS ix_inbound_qc_inspections_warehouse_status
  ON qc.inbound_qc_inspections(warehouse_ref, status, updated_at DESC);

CREATE INDEX IF NOT EXISTS ix_inbound_qc_checklist_inspection
  ON qc.inbound_qc_checklist_items(inspection_id, checklist_no);

COMMIT;
