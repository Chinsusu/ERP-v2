BEGIN;

CREATE TABLE IF NOT EXISTS inventory.stock_adjustments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  adjustment_ref text NOT NULL,
  adjustment_no text NOT NULL,
  org_ref text,
  warehouse_id uuid REFERENCES mdm.warehouses(id),
  warehouse_ref text,
  warehouse_code text,
  source_type text NOT NULL,
  source_id uuid,
  source_ref text,
  reason text NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  requested_by uuid REFERENCES core.users(id),
  requested_by_ref text,
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  submitted_by_ref text,
  approved_at timestamptz,
  approved_by uuid REFERENCES core.users(id),
  approved_by_ref text,
  rejected_at timestamptz,
  rejected_by uuid REFERENCES core.users(id),
  rejected_by_ref text,
  posted_at timestamptz,
  posted_by uuid REFERENCES core.users(id),
  posted_by_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_stock_adjustments_org_ref UNIQUE (org_id, adjustment_ref),
  CONSTRAINT uq_stock_adjustments_org_no UNIQUE (org_id, adjustment_no),
  CONSTRAINT ck_stock_adjustments_status CHECK (
    status IN ('draft', 'submitted', 'approved', 'rejected', 'posted', 'cancelled')
  ),
  CONSTRAINT ck_stock_adjustments_refs CHECK (
    nullif(btrim(adjustment_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(org_ref, org_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text)), '') IS NOT NULL
  ),
  CONSTRAINT ck_stock_adjustments_lifecycle CHECK (
    status = 'draft'
    OR (
      status = 'submitted'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'approved'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
      AND approved_at IS NOT NULL
      AND (approved_by IS NOT NULL OR nullif(btrim(coalesce(approved_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'rejected'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
      AND rejected_at IS NOT NULL
      AND (rejected_by IS NOT NULL OR nullif(btrim(coalesce(rejected_by_ref, '')), '') IS NOT NULL)
    )
    OR (
      status = 'posted'
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
      AND approved_at IS NOT NULL
      AND (approved_by IS NOT NULL OR nullif(btrim(coalesce(approved_by_ref, '')), '') IS NOT NULL)
      AND posted_at IS NOT NULL
      AND (posted_by IS NOT NULL OR nullif(btrim(coalesce(posted_by_ref, '')), '') IS NOT NULL)
    )
    OR status = 'cancelled'
  )
);

CREATE TABLE IF NOT EXISTS inventory.stock_adjustment_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  adjustment_id uuid NOT NULL REFERENCES inventory.stock_adjustments(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL DEFAULT 1,
  item_id uuid REFERENCES mdm.items(id),
  item_ref text,
  sku_code text NOT NULL,
  batch_id uuid REFERENCES inventory.batches(id),
  batch_ref text,
  batch_no text,
  location_id uuid REFERENCES mdm.warehouse_bins(id),
  location_ref text,
  location_code text,
  expected_qty numeric(18, 6) NOT NULL,
  counted_qty numeric(18, 6) NOT NULL,
  delta_qty numeric(18, 6) NOT NULL,
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_stock_adjustment_lines_ref UNIQUE (org_id, adjustment_id, line_ref),
  CONSTRAINT ck_stock_adjustment_lines_qty CHECK (expected_qty >= 0 AND counted_qty >= 0),
  CONSTRAINT ck_stock_adjustment_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code)), '') IS NOT NULL
  )
);

CREATE INDEX IF NOT EXISTS ix_stock_adjustments_status
  ON inventory.stock_adjustments(org_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_stock_adjustment_lines_adjustment
  ON inventory.stock_adjustment_lines(adjustment_id, line_no);

COMMIT;
