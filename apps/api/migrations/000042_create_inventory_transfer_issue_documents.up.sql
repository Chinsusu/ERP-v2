BEGIN;

CREATE TABLE IF NOT EXISTS inventory.stock_transfer_documents (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  transfer_ref text NOT NULL,
  transfer_no text NOT NULL,
  source_warehouse_ref text NOT NULL,
  destination_warehouse_ref text NOT NULL,
  status text NOT NULL,
  transfer_payload jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT uq_stock_transfer_documents_org_ref UNIQUE (org_id, transfer_ref),
  CONSTRAINT ck_stock_transfer_documents_required CHECK (
    nullif(btrim(transfer_ref), '') IS NOT NULL
    AND nullif(btrim(transfer_no), '') IS NOT NULL
    AND nullif(btrim(source_warehouse_ref), '') IS NOT NULL
    AND nullif(btrim(destination_warehouse_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_stock_transfer_documents_status CHECK (
    status IN ('draft', 'submitted', 'approved', 'posted', 'cancelled')
  )
);

CREATE INDEX IF NOT EXISTS ix_stock_transfer_documents_status
  ON inventory.stock_transfer_documents(org_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS inventory.warehouse_issue_documents (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  issue_ref text NOT NULL,
  issue_no text NOT NULL,
  warehouse_ref text NOT NULL,
  destination_type text NOT NULL,
  destination_name text NOT NULL,
  status text NOT NULL,
  issue_payload jsonb NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  CONSTRAINT uq_warehouse_issue_documents_org_ref UNIQUE (org_id, issue_ref),
  CONSTRAINT ck_warehouse_issue_documents_required CHECK (
    nullif(btrim(issue_ref), '') IS NOT NULL
    AND nullif(btrim(issue_no), '') IS NOT NULL
    AND nullif(btrim(warehouse_ref), '') IS NOT NULL
    AND nullif(btrim(destination_type), '') IS NOT NULL
    AND nullif(btrim(destination_name), '') IS NOT NULL
  ),
  CONSTRAINT ck_warehouse_issue_documents_status CHECK (
    status IN ('draft', 'submitted', 'approved', 'posted', 'cancelled')
  )
);

CREATE INDEX IF NOT EXISTS ix_warehouse_issue_documents_status
  ON inventory.warehouse_issue_documents(org_id, status, created_at DESC);

COMMIT;
