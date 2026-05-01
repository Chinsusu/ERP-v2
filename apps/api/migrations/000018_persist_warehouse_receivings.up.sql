BEGIN;

CREATE TABLE IF NOT EXISTS inventory.warehouse_receivings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  receipt_ref text NOT NULL,
  receipt_no text NOT NULL,
  org_ref text,
  warehouse_id uuid REFERENCES mdm.warehouses(id),
  warehouse_ref text,
  warehouse_code text,
  location_id uuid REFERENCES mdm.warehouse_bins(id),
  location_ref text,
  location_code text,
  reference_doc_type text NOT NULL,
  reference_doc_id uuid,
  reference_doc_ref text,
  supplier_id uuid REFERENCES mdm.suppliers(id),
  supplier_ref text,
  delivery_note_no text,
  status text NOT NULL DEFAULT 'draft',
  created_by uuid REFERENCES core.users(id),
  created_by_ref text,
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  submitted_by_ref text,
  inspect_ready_at timestamptz,
  inspect_ready_by uuid REFERENCES core.users(id),
  inspect_ready_by_ref text,
  posted_at timestamptz,
  posted_by uuid REFERENCES core.users(id),
  posted_by_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_warehouse_receivings_org_ref UNIQUE (org_id, receipt_ref),
  CONSTRAINT uq_warehouse_receivings_org_no UNIQUE (org_id, receipt_no),
  CONSTRAINT ck_warehouse_receivings_status CHECK (
    status IN ('draft', 'submitted', 'inspect_ready', 'posted')
  ),
  CONSTRAINT ck_warehouse_receivings_refs CHECK (
    nullif(btrim(receipt_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(org_ref, org_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(location_ref, location_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(reference_doc_ref, reference_doc_id::text)), '') IS NOT NULL
  ),
  CONSTRAINT ck_warehouse_receivings_lifecycle CHECK (
    status = 'draft'
    OR (
      status = 'submitted'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'inspect_ready'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
      AND inspect_ready_at IS NOT NULL
      AND (inspect_ready_by IS NOT NULL OR nullif(btrim(coalesce(inspect_ready_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'posted'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
      AND inspect_ready_at IS NOT NULL
      AND (inspect_ready_by IS NOT NULL OR nullif(btrim(coalesce(inspect_ready_by_ref, '')), '') IS NOT NULL)
      AND posted_at IS NOT NULL
      AND (posted_by IS NOT NULL OR nullif(btrim(coalesce(posted_by_ref, '')), '') IS NOT NULL)
    )
  )
);

CREATE TABLE IF NOT EXISTS inventory.warehouse_receiving_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  receipt_id uuid NOT NULL REFERENCES inventory.warehouse_receivings(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL DEFAULT 1,
  purchase_order_line_id uuid,
  purchase_order_line_ref text,
  item_id uuid REFERENCES mdm.items(id),
  item_ref text,
  sku_code text NOT NULL,
  item_name text,
  batch_id uuid REFERENCES inventory.batches(id),
  batch_ref text,
  batch_no text,
  lot_no text,
  expiry_date date,
  warehouse_id uuid REFERENCES mdm.warehouses(id),
  warehouse_ref text,
  location_id uuid REFERENCES mdm.warehouse_bins(id),
  location_ref text,
  quantity numeric(18, 6) NOT NULL,
  uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  packaging_status text NOT NULL,
  qc_status text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_warehouse_receiving_lines_ref UNIQUE (org_id, receipt_id, line_ref),
  CONSTRAINT ck_warehouse_receiving_lines_qty CHECK (quantity > 0),
  CONSTRAINT ck_warehouse_receiving_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code)), '') IS NOT NULL
    AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(location_ref, location_id::text)), '') IS NOT NULL
  )
);

CREATE TABLE IF NOT EXISTS inventory.warehouse_receiving_stock_movements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  receipt_id uuid NOT NULL REFERENCES inventory.warehouse_receivings(id) ON DELETE CASCADE,
  movement_ref text NOT NULL,
  movement_no text NOT NULL,
  movement_type text NOT NULL,
  movement_at timestamptz NOT NULL,
  item_id uuid REFERENCES mdm.items(id),
  item_ref text,
  batch_id uuid REFERENCES inventory.batches(id),
  batch_ref text,
  warehouse_id uuid REFERENCES mdm.warehouses(id),
  warehouse_ref text,
  location_id uuid REFERENCES mdm.warehouse_bins(id),
  location_ref text,
  quantity numeric(18, 6) NOT NULL,
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  source_quantity numeric(18, 6) NOT NULL,
  source_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  conversion_factor numeric(18, 6) NOT NULL,
  stock_status text NOT NULL,
  source_doc_ref text NOT NULL,
  source_doc_line_ref text,
  reason text NOT NULL,
  created_by uuid REFERENCES core.users(id),
  created_by_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_warehouse_receiving_movements_ref UNIQUE (org_id, receipt_id, movement_ref),
  CONSTRAINT ck_warehouse_receiving_movements_qty CHECK (
    quantity > 0
    AND source_quantity > 0
    AND conversion_factor > 0
  )
);

CREATE INDEX IF NOT EXISTS ix_warehouse_receivings_status
  ON inventory.warehouse_receivings(org_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_warehouse_receivings_warehouse_status
  ON inventory.warehouse_receivings(warehouse_ref, status, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_warehouse_receiving_lines_receipt
  ON inventory.warehouse_receiving_lines(receipt_id, line_no);

CREATE INDEX IF NOT EXISTS ix_warehouse_receiving_movements_receipt
  ON inventory.warehouse_receiving_stock_movements(receipt_id, movement_no);

COMMIT;
