BEGIN;

CREATE TABLE IF NOT EXISTS inventory.stock_count_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  count_ref text NOT NULL,
  count_no text NOT NULL,
  org_ref text,
  warehouse_id uuid REFERENCES mdm.warehouses(id),
  warehouse_ref text,
  warehouse_code text,
  count_date date NOT NULL,
  scope text NOT NULL DEFAULT 'warehouse',
  status text NOT NULL DEFAULT 'open',
  created_by uuid REFERENCES core.users(id),
  created_by_ref text,
  submitted_at timestamptz,
  submitted_by uuid REFERENCES core.users(id),
  submitted_by_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_stock_count_sessions_org_ref UNIQUE (org_id, count_ref),
  CONSTRAINT uq_stock_count_sessions_org_no UNIQUE (org_id, count_no),
  CONSTRAINT ck_stock_count_sessions_status CHECK (
    status IN ('open', 'submitted', 'variance_review')
  ),
  CONSTRAINT ck_stock_count_sessions_refs CHECK (
    nullif(btrim(count_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(org_ref, org_id::text)), '') IS NOT NULL
    AND nullif(btrim(coalesce(warehouse_ref, warehouse_id::text)), '') IS NOT NULL
  ),
  CONSTRAINT ck_stock_count_sessions_lifecycle CHECK (
    (
      status = 'open'
      AND submitted_at IS NULL
      AND submitted_by IS NULL
      AND nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NULL
    )
    OR (
      status IN ('submitted', 'variance_review')
      AND submitted_at IS NOT NULL
      AND (submitted_by IS NOT NULL OR nullif(btrim(coalesce(submitted_by_ref, '')), '') IS NOT NULL)
    )
  )
);

CREATE TABLE IF NOT EXISTS inventory.stock_count_session_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  session_id uuid NOT NULL REFERENCES inventory.stock_count_sessions(id) ON DELETE CASCADE,
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
  counted_qty numeric(18, 6) NOT NULL DEFAULT 0,
  delta_qty numeric(18, 6) NOT NULL DEFAULT 0,
  base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code),
  counted boolean NOT NULL DEFAULT false,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_stock_count_session_lines_ref UNIQUE (org_id, session_id, line_ref),
  CONSTRAINT ck_stock_count_session_lines_qty CHECK (expected_qty >= 0 AND counted_qty >= 0),
  CONSTRAINT ck_stock_count_session_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(coalesce(item_ref, item_id::text, sku_code)), '') IS NOT NULL
  )
);

CREATE INDEX IF NOT EXISTS ix_stock_count_sessions_status
  ON inventory.stock_count_sessions(org_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_stock_count_session_lines_session
  ON inventory.stock_count_session_lines(session_id, line_no);

COMMIT;
