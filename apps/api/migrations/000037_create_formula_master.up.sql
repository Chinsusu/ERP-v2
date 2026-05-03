BEGIN;

CREATE TABLE IF NOT EXISTS mdm.item_formulas (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  formula_ref text NOT NULL,
  formula_code text NOT NULL,
  finished_item_ref text NOT NULL,
  finished_sku text NOT NULL,
  finished_item_name text NOT NULL,
  finished_item_type text NOT NULL,
  formula_version text NOT NULL,
  batch_qty numeric(18,6) NOT NULL,
  batch_uom_code text NOT NULL,
  base_batch_qty numeric(18,6) NOT NULL,
  base_batch_uom_code text NOT NULL,
  status text NOT NULL,
  approval_status text NOT NULL,
  effective_from date,
  effective_to date,
  note text,
  approved_by_ref text,
  approved_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_item_formulas_required CHECK (
    nullif(btrim(formula_ref), '') IS NOT NULL
    AND nullif(btrim(formula_code), '') IS NOT NULL
    AND nullif(btrim(finished_item_ref), '') IS NOT NULL
    AND nullif(btrim(finished_sku), '') IS NOT NULL
    AND nullif(btrim(finished_item_name), '') IS NOT NULL
    AND nullif(btrim(formula_version), '') IS NOT NULL
    AND batch_qty > 0
    AND base_batch_qty > 0
    AND version > 0
  ),
  CONSTRAINT ck_item_formulas_finished_type CHECK (
    finished_item_type IN ('finished_good', 'semi_finished')
  ),
  CONSTRAINT ck_item_formulas_status CHECK (
    status IN ('draft', 'active', 'inactive', 'archived')
  ),
  CONSTRAINT ck_item_formulas_approval_status CHECK (
    approval_status IN ('draft', 'pending_approval', 'approved', 'rejected')
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_item_formulas_org_ref
  ON mdm.item_formulas(org_id, lower(formula_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_item_formulas_org_finished_version
  ON mdm.item_formulas(org_id, lower(finished_item_ref), lower(formula_version));

CREATE UNIQUE INDEX IF NOT EXISTS uq_item_formulas_one_active
  ON mdm.item_formulas(org_id, lower(finished_item_ref))
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS ix_item_formulas_finished_status
  ON mdm.item_formulas(org_id, lower(finished_item_ref), status, updated_at DESC);

CREATE TABLE IF NOT EXISTS mdm.item_formula_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  formula_id uuid NOT NULL REFERENCES mdm.item_formulas(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  line_no integer NOT NULL,
  component_item_ref text,
  component_sku text,
  component_name text,
  component_type text NOT NULL,
  entered_qty numeric(18,6) NOT NULL,
  entered_uom_code text NOT NULL,
  calc_qty numeric(18,6) NOT NULL,
  calc_uom_code text NOT NULL,
  stock_base_qty numeric(18,6) NOT NULL,
  stock_base_uom_code text NOT NULL,
  waste_percent numeric(9,4) NOT NULL DEFAULT 0,
  is_required boolean NOT NULL DEFAULT true,
  is_stock_managed boolean NOT NULL DEFAULT true,
  line_status text NOT NULL DEFAULT 'active',
  note text,
  CONSTRAINT uq_item_formula_lines_no UNIQUE (formula_id, line_no),
  CONSTRAINT ck_item_formula_lines_required CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND line_no > 0
    AND entered_qty >= 0
    AND calc_qty >= 0
    AND stock_base_qty >= 0
    AND waste_percent >= 0
  ),
  CONSTRAINT ck_item_formula_lines_component_type CHECK (
    component_type IN ('raw_material', 'fragrance', 'packaging', 'semi_finished', 'service')
  ),
  CONSTRAINT ck_item_formula_lines_status CHECK (
    line_status IN ('active', 'excluded', 'needs_review')
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_item_formula_lines_ref
  ON mdm.item_formula_lines(formula_id, lower(line_ref));

CREATE INDEX IF NOT EXISTS ix_item_formula_lines_formula
  ON mdm.item_formula_lines(formula_id, line_no);

CREATE INDEX IF NOT EXISTS ix_item_formula_lines_component
  ON mdm.item_formula_lines(org_id, lower(component_item_ref))
  WHERE nullif(btrim(component_item_ref), '') IS NOT NULL;

COMMIT;
