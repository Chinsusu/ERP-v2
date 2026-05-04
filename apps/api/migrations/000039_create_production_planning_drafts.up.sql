BEGIN;

CREATE TABLE IF NOT EXISTS subcontract.production_plans (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  plan_ref text NOT NULL,
  plan_no text NOT NULL,
  output_item_ref text NOT NULL,
  output_sku text NOT NULL,
  output_item_name text NOT NULL,
  formula_ref text NOT NULL,
  formula_code text NOT NULL,
  formula_version text NOT NULL,
  status text NOT NULL,
  plan_payload jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_production_plans_required CHECK (
    nullif(btrim(plan_ref), '') IS NOT NULL
    AND nullif(btrim(plan_no), '') IS NOT NULL
    AND nullif(btrim(output_item_ref), '') IS NOT NULL
    AND nullif(btrim(output_sku), '') IS NOT NULL
    AND nullif(btrim(output_item_name), '') IS NOT NULL
    AND nullif(btrim(formula_ref), '') IS NOT NULL
    AND nullif(btrim(formula_code), '') IS NOT NULL
    AND nullif(btrim(formula_version), '') IS NOT NULL
    AND version > 0
  ),
  CONSTRAINT ck_production_plans_status CHECK (
    status IN ('draft', 'purchase_request_draft_created', 'cancelled')
  ),
  CONSTRAINT uq_production_plans_org_ref UNIQUE (org_id, plan_ref)
);

CREATE INDEX IF NOT EXISTS ix_production_plans_org_status
  ON subcontract.production_plans(org_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS ix_production_plans_output_item
  ON subcontract.production_plans(org_id, lower(output_item_ref), created_at DESC);

CREATE TABLE IF NOT EXISTS subcontract.purchase_request_drafts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  draft_ref text NOT NULL,
  request_no text NOT NULL,
  source_plan_ref text NOT NULL,
  status text NOT NULL,
  draft_payload jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  CONSTRAINT ck_purchase_request_drafts_required CHECK (
    nullif(btrim(draft_ref), '') IS NOT NULL
    AND nullif(btrim(request_no), '') IS NOT NULL
    AND nullif(btrim(source_plan_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_purchase_request_drafts_status CHECK (
    status IN ('draft')
  ),
  CONSTRAINT uq_purchase_request_drafts_org_ref UNIQUE (org_id, draft_ref)
);

CREATE INDEX IF NOT EXISTS ix_purchase_request_drafts_source_plan
  ON subcontract.purchase_request_drafts(org_id, lower(source_plan_ref), created_at DESC);

COMMIT;
