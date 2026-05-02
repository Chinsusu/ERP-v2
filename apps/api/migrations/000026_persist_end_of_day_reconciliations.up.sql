BEGIN;

ALTER TABLE inventory.warehouse_daily_closings
  ADD COLUMN IF NOT EXISTS closing_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS owner_ref text,
  ADD COLUMN IF NOT EXISTS closed_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text,
  ADD COLUMN IF NOT EXISTS handover_order_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS return_order_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS stock_movement_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS stock_count_session_count integer NOT NULL DEFAULT 0;

UPDATE inventory.warehouse_daily_closings AS closing
SET closing_ref = COALESCE(NULLIF(btrim(closing.closing_ref), ''), closing.closing_no, closing.id::text),
    org_ref = COALESCE(NULLIF(btrim(closing.org_ref), ''), closing.org_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(closing.warehouse_ref), ''), closing.warehouse_id::text),
    warehouse_code = COALESCE(NULLIF(btrim(closing.warehouse_code), ''), warehouse.code),
    owner_ref = COALESCE(NULLIF(btrim(closing.owner_ref), ''), NULLIF(btrim(closing.created_by_ref), ''), closing.created_by::text),
    closed_by_ref = COALESCE(NULLIF(btrim(closing.closed_by_ref), ''), closing.closed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(closing.created_by_ref), ''), closing.created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(closing.updated_by_ref), ''), closing.updated_by::text)
FROM mdm.warehouses AS warehouse
WHERE warehouse.id = closing.warehouse_id;

UPDATE inventory.warehouse_daily_closings AS closing
SET closing_ref = COALESCE(NULLIF(btrim(closing_ref), ''), closing_no, id::text),
    org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text),
    warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text),
    owner_ref = COALESCE(NULLIF(btrim(owner_ref), ''), NULLIF(btrim(created_by_ref), ''), created_by::text),
    closed_by_ref = COALESCE(NULLIF(btrim(closed_by_ref), ''), closed_by::text),
    created_by_ref = COALESCE(NULLIF(btrim(created_by_ref), ''), created_by::text),
    updated_by_ref = COALESCE(NULLIF(btrim(updated_by_ref), ''), updated_by::text)
WHERE closing_ref IS NULL
   OR org_ref IS NULL
   OR warehouse_ref IS NULL
   OR owner_ref IS NULL
   OR closed_by_ref IS NULL
   OR created_by_ref IS NULL
   OR updated_by_ref IS NULL;

CREATE TABLE IF NOT EXISTS inventory.warehouse_daily_closing_checklist (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  closing_id uuid NOT NULL REFERENCES inventory.warehouse_daily_closings(id) ON DELETE CASCADE,
  item_ref text NOT NULL,
  label text NOT NULL,
  complete boolean NOT NULL DEFAULT false,
  blocking boolean NOT NULL DEFAULT true,
  note text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_warehouse_daily_closing_checklist_ref UNIQUE (org_id, closing_id, item_ref),
  CONSTRAINT ck_warehouse_daily_closing_checklist_refs CHECK (
    nullif(btrim(item_ref), '') IS NOT NULL
    AND nullif(btrim(label), '') IS NOT NULL
  )
);

CREATE TABLE IF NOT EXISTS inventory.warehouse_daily_closing_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  closing_id uuid NOT NULL REFERENCES inventory.warehouse_daily_closings(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL DEFAULT 1,
  sku_code text NOT NULL,
  batch_no text,
  bin_code text,
  system_qty numeric(18,6) NOT NULL DEFAULT 0,
  counted_qty numeric(18,6) NOT NULL DEFAULT 0,
  reason text,
  owner_ref text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT uq_warehouse_daily_closing_lines_ref UNIQUE (org_id, closing_id, line_ref),
  CONSTRAINT ck_warehouse_daily_closing_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(sku_code), '') IS NOT NULL
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_warehouse_daily_closings_org_ref
  ON inventory.warehouse_daily_closings(org_id, lower(closing_ref))
  WHERE nullif(btrim(closing_ref), '') IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_warehouse_daily_closings_filters
  ON inventory.warehouse_daily_closings(org_id, warehouse_ref, business_date, shift_code, status);

CREATE INDEX IF NOT EXISTS ix_warehouse_daily_closing_checklist_closing
  ON inventory.warehouse_daily_closing_checklist(closing_id, item_ref);

CREATE INDEX IF NOT EXISTS ix_warehouse_daily_closing_lines_closing
  ON inventory.warehouse_daily_closing_lines(closing_id, line_no);

COMMIT;
