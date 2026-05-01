BEGIN;

CREATE TABLE IF NOT EXISTS inventory.supplier_rejections (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  rejection_ref text NOT NULL,
  org_ref text NOT NULL,
  rejection_no text NOT NULL,
  supplier_id uuid REFERENCES mdm.suppliers(id),
  supplier_ref text NOT NULL,
  supplier_code text,
  supplier_name text NOT NULL,
  purchase_order_id uuid REFERENCES purchase.purchase_orders(id),
  purchase_order_ref text,
  purchase_order_no text,
  goods_receipt_id uuid REFERENCES inventory.warehouse_receivings(id),
  goods_receipt_ref text NOT NULL,
  goods_receipt_no text,
  inbound_qc_inspection_id uuid REFERENCES qc.inbound_qc_inspections(id),
  inbound_qc_inspection_ref text NOT NULL,
  warehouse_id uuid REFERENCES mdm.warehouses(id),
  warehouse_ref text NOT NULL,
  warehouse_code text,
  status text NOT NULL,
  reason text NOT NULL,
  created_by uuid REFERENCES core.users(id),
  created_by_ref text NOT NULL,
  updated_by uuid REFERENCES core.users(id),
  updated_by_ref text,
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  submitted_by_ref text,
  confirmed_at timestamptz,
  confirmed_by uuid REFERENCES core.users(id),
  confirmed_by_ref text,
  cancelled_at timestamptz,
  cancelled_by uuid REFERENCES core.users(id),
  cancelled_by_ref text,
  cancel_reason text,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT uq_supplier_rejections_org_ref UNIQUE (org_id, rejection_ref),
  CONSTRAINT uq_supplier_rejections_org_no UNIQUE (org_id, rejection_no),
  CONSTRAINT ck_supplier_rejections_status CHECK (
    status IN ('draft', 'submitted', 'confirmed', 'cancelled')
  ),
  CONSTRAINT ck_supplier_rejections_refs CHECK (
    nullif(btrim(rejection_ref), '') IS NOT NULL
    AND nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(supplier_ref), '') IS NOT NULL
    AND nullif(btrim(goods_receipt_ref), '') IS NOT NULL
    AND nullif(btrim(inbound_qc_inspection_ref), '') IS NOT NULL
    AND nullif(btrim(warehouse_ref), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_supplier_rejections_lifecycle CHECK (
    status = 'draft'
    OR (
      status = 'submitted'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'confirmed'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
      AND confirmed_at IS NOT NULL
      AND (confirmed_by IS NOT NULL OR nullif(btrim(coalesce(confirmed_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'cancelled'
      AND cancelled_at IS NOT NULL
      AND (cancelled_by IS NOT NULL OR nullif(btrim(coalesce(cancelled_by_ref, '')), '') IS NOT NULL)
      AND nullif(btrim(coalesce(cancel_reason, '')), '') IS NOT NULL
    )
  )
);

CREATE TABLE IF NOT EXISTS inventory.supplier_rejection_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  rejection_id uuid NOT NULL REFERENCES inventory.supplier_rejections(id) ON DELETE RESTRICT,
  line_ref text NOT NULL,
  line_no integer NOT NULL,
  purchase_order_line_id uuid REFERENCES purchase.purchase_order_lines(id),
  purchase_order_line_ref text,
  goods_receipt_line_id uuid REFERENCES inventory.warehouse_receiving_lines(id),
  goods_receipt_line_ref text NOT NULL,
  inbound_qc_inspection_id uuid REFERENCES qc.inbound_qc_inspections(id),
  inbound_qc_inspection_ref text NOT NULL,
  item_id uuid REFERENCES mdm.items(id),
  item_ref text NOT NULL,
  sku_code text NOT NULL,
  item_name text,
  batch_id uuid REFERENCES inventory.batches(id),
  batch_ref text NOT NULL,
  batch_no text NOT NULL,
  lot_no text NOT NULL,
  expiry_date date NOT NULL,
  rejected_qty numeric(18,6) NOT NULL,
  uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  reason text NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT uq_supplier_rejection_lines_ref UNIQUE (rejection_id, line_ref),
  CONSTRAINT ck_supplier_rejection_lines_qty CHECK (rejected_qty > 0),
  CONSTRAINT ck_supplier_rejection_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(goods_receipt_line_ref), '') IS NOT NULL
    AND nullif(btrim(inbound_qc_inspection_ref), '') IS NOT NULL
    AND nullif(btrim(item_ref), '') IS NOT NULL
    AND nullif(btrim(sku_code), '') IS NOT NULL
    AND nullif(btrim(batch_ref), '') IS NOT NULL
    AND nullif(btrim(batch_no), '') IS NOT NULL
    AND nullif(btrim(lot_no), '') IS NOT NULL
  )
);

CREATE TABLE IF NOT EXISTS inventory.supplier_rejection_attachments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  rejection_id uuid NOT NULL REFERENCES inventory.supplier_rejections(id) ON DELETE RESTRICT,
  attachment_ref text NOT NULL,
  line_ref text,
  file_name text NOT NULL,
  object_key text NOT NULL,
  content_type text,
  uploaded_at timestamptz NOT NULL,
  uploaded_by_ref text NOT NULL,
  source text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_supplier_rejection_attachments_ref UNIQUE (org_id, attachment_ref),
  CONSTRAINT ck_supplier_rejection_attachments_refs CHECK (
    nullif(btrim(attachment_ref), '') IS NOT NULL
    AND nullif(btrim(file_name), '') IS NOT NULL
    AND nullif(btrim(object_key), '') IS NOT NULL
    AND nullif(btrim(uploaded_by_ref), '') IS NOT NULL
  )
);

CREATE INDEX IF NOT EXISTS ix_supplier_rejections_supplier_status
  ON inventory.supplier_rejections(org_id, supplier_ref, status);

CREATE INDEX IF NOT EXISTS ix_supplier_rejections_warehouse_status
  ON inventory.supplier_rejections(org_id, warehouse_ref, status);

CREATE INDEX IF NOT EXISTS ix_supplier_rejections_inbound_qc
  ON inventory.supplier_rejections(org_id, inbound_qc_inspection_ref);

CREATE INDEX IF NOT EXISTS ix_supplier_rejection_lines_receipt_line
  ON inventory.supplier_rejection_lines(org_id, goods_receipt_line_ref);

CREATE INDEX IF NOT EXISTS ix_supplier_rejection_attachments_rejection
  ON inventory.supplier_rejection_attachments(rejection_id);

COMMIT;
