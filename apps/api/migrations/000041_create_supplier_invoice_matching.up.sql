BEGIN;

CREATE TABLE IF NOT EXISTS finance.supplier_invoices (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  invoice_ref text NOT NULL,
  invoice_no text NOT NULL,
  supplier_ref text NOT NULL,
  supplier_code text,
  supplier_name text NOT NULL,
  payable_ref text NOT NULL,
  payable_no text NOT NULL,
  status text NOT NULL,
  match_status text NOT NULL,
  source_document_type text NOT NULL,
  source_document_ref text,
  source_document_no text,
  invoice_amount numeric(18,2) NOT NULL,
  expected_amount numeric(18,2) NOT NULL,
  variance_amount numeric(18,2) NOT NULL DEFAULT 0,
  currency_code text NOT NULL DEFAULT 'VND',
  invoice_date date NOT NULL,
  void_reason text,
  voided_by_ref text,
  voided_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_supplier_invoices_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(invoice_ref), '') IS NOT NULL
    AND nullif(btrim(invoice_no), '') IS NOT NULL
    AND nullif(btrim(supplier_ref), '') IS NOT NULL
    AND nullif(btrim(supplier_name), '') IS NOT NULL
    AND nullif(btrim(payable_ref), '') IS NOT NULL
    AND nullif(btrim(payable_no), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND (
      nullif(btrim(source_document_ref), '') IS NOT NULL
      OR nullif(btrim(source_document_no), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_supplier_invoices_status CHECK (
    status IN ('draft', 'matched', 'mismatch', 'void')
    AND match_status IN ('pending', 'matched', 'mismatch')
  ),
  CONSTRAINT ck_supplier_invoices_money CHECK (
    currency_code = 'VND'
    AND invoice_amount > 0
    AND expected_amount > 0
    AND variance_amount = invoice_amount - expected_amount
    AND version > 0
  )
);

CREATE TABLE IF NOT EXISTS finance.supplier_invoice_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  supplier_invoice_id uuid NOT NULL REFERENCES finance.supplier_invoices(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  invoice_ref text NOT NULL,
  description text NOT NULL,
  source_document_type text NOT NULL,
  source_document_ref text,
  source_document_no text,
  amount numeric(18,2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_supplier_invoice_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(invoice_ref), '') IS NOT NULL
    AND nullif(btrim(description), '') IS NOT NULL
    AND (
      nullif(btrim(source_document_ref), '') IS NOT NULL
      OR nullif(btrim(source_document_no), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_supplier_invoice_lines_money CHECK (amount <> 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_invoices_org_ref
  ON finance.supplier_invoices(org_id, lower(invoice_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_invoices_org_no
  ON finance.supplier_invoices(org_id, lower(invoice_no));

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_invoice_lines_ref
  ON finance.supplier_invoice_lines(org_id, supplier_invoice_id, lower(line_ref));

CREATE INDEX IF NOT EXISTS ix_supplier_invoices_filters
  ON finance.supplier_invoices(org_id, supplier_ref, payable_ref, status, match_status, invoice_date DESC);

COMMIT;
