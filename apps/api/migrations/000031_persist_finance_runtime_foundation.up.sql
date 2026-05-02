BEGIN;

CREATE SCHEMA IF NOT EXISTS finance;

CREATE TABLE IF NOT EXISTS finance.customer_receivables (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  receivable_ref text NOT NULL,
  receivable_no text NOT NULL,
  customer_ref text NOT NULL,
  customer_code text,
  customer_name text NOT NULL,
  status text NOT NULL,
  source_document_type text NOT NULL,
  source_document_ref text,
  source_document_no text,
  total_amount numeric(18,2) NOT NULL,
  paid_amount numeric(18,2) NOT NULL DEFAULT 0,
  outstanding_amount numeric(18,2) NOT NULL DEFAULT 0,
  currency_code text NOT NULL DEFAULT 'VND',
  due_date date NOT NULL,
  dispute_reason text,
  disputed_by_ref text,
  disputed_at timestamptz,
  void_reason text,
  voided_by_ref text,
  voided_at timestamptz,
  last_receipt_by_ref text,
  last_receipt_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_customer_receivables_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(receivable_ref), '') IS NOT NULL
    AND nullif(btrim(receivable_no), '') IS NOT NULL
    AND nullif(btrim(customer_ref), '') IS NOT NULL
    AND nullif(btrim(customer_name), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND (
      nullif(btrim(source_document_ref), '') IS NOT NULL
      OR nullif(btrim(source_document_no), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_customer_receivables_status CHECK (
    status IN ('draft', 'open', 'partially_paid', 'paid', 'disputed', 'void')
  ),
  CONSTRAINT ck_customer_receivables_money CHECK (
    currency_code = 'VND'
    AND total_amount > 0
    AND paid_amount >= 0
    AND outstanding_amount >= 0
    AND paid_amount + outstanding_amount = total_amount
    AND version > 0
  )
);

CREATE TABLE IF NOT EXISTS finance.customer_receivable_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  customer_receivable_id uuid NOT NULL REFERENCES finance.customer_receivables(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  receivable_ref text NOT NULL,
  description text NOT NULL,
  source_document_type text NOT NULL,
  source_document_ref text,
  source_document_no text,
  amount numeric(18,2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_customer_receivable_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(receivable_ref), '') IS NOT NULL
    AND nullif(btrim(description), '') IS NOT NULL
    AND (
      nullif(btrim(source_document_ref), '') IS NOT NULL
      OR nullif(btrim(source_document_no), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_customer_receivable_lines_money CHECK (amount > 0)
);

CREATE TABLE IF NOT EXISTS finance.supplier_payables (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  payable_ref text NOT NULL,
  payable_no text NOT NULL,
  supplier_ref text NOT NULL,
  supplier_code text,
  supplier_name text NOT NULL,
  status text NOT NULL,
  source_document_type text NOT NULL,
  source_document_ref text,
  source_document_no text,
  total_amount numeric(18,2) NOT NULL,
  paid_amount numeric(18,2) NOT NULL DEFAULT 0,
  outstanding_amount numeric(18,2) NOT NULL DEFAULT 0,
  currency_code text NOT NULL DEFAULT 'VND',
  due_date date NOT NULL,
  payment_requested_by_ref text,
  payment_requested_at timestamptz,
  payment_approved_by_ref text,
  payment_approved_at timestamptz,
  payment_rejected_by_ref text,
  payment_rejected_at timestamptz,
  payment_reject_reason text,
  dispute_reason text,
  disputed_by_ref text,
  disputed_at timestamptz,
  void_reason text,
  voided_by_ref text,
  voided_at timestamptz,
  last_payment_by_ref text,
  last_payment_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_supplier_payables_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(payable_ref), '') IS NOT NULL
    AND nullif(btrim(payable_no), '') IS NOT NULL
    AND nullif(btrim(supplier_ref), '') IS NOT NULL
    AND nullif(btrim(supplier_name), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
    AND (
      nullif(btrim(source_document_ref), '') IS NOT NULL
      OR nullif(btrim(source_document_no), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_supplier_payables_status CHECK (
    status IN (
      'draft',
      'open',
      'payment_requested',
      'payment_approved',
      'partially_paid',
      'paid',
      'disputed',
      'void'
    )
  ),
  CONSTRAINT ck_supplier_payables_money CHECK (
    currency_code = 'VND'
    AND total_amount > 0
    AND paid_amount >= 0
    AND outstanding_amount >= 0
    AND version > 0
  )
);

CREATE TABLE IF NOT EXISTS finance.supplier_payable_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  supplier_payable_id uuid NOT NULL REFERENCES finance.supplier_payables(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  payable_ref text NOT NULL,
  description text NOT NULL,
  source_document_type text NOT NULL,
  source_document_ref text,
  source_document_no text,
  amount numeric(18,2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_supplier_payable_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(payable_ref), '') IS NOT NULL
    AND nullif(btrim(description), '') IS NOT NULL
    AND (
      nullif(btrim(source_document_ref), '') IS NOT NULL
      OR nullif(btrim(source_document_no), '') IS NOT NULL
    )
  ),
  CONSTRAINT ck_supplier_payable_lines_money CHECK (amount > 0)
);

CREATE TABLE IF NOT EXISTS finance.cod_remittances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  remittance_ref text NOT NULL,
  remittance_no text NOT NULL,
  carrier_ref text NOT NULL,
  carrier_code text,
  carrier_name text NOT NULL,
  status text NOT NULL,
  business_date date NOT NULL,
  expected_amount numeric(18,2) NOT NULL,
  remitted_amount numeric(18,2) NOT NULL DEFAULT 0,
  discrepancy_amount numeric(18,2) NOT NULL DEFAULT 0,
  currency_code text NOT NULL DEFAULT 'VND',
  submitted_by_ref text,
  submitted_at timestamptz,
  approved_by_ref text,
  approved_at timestamptz,
  closed_by_ref text,
  closed_at timestamptz,
  void_reason text,
  voided_by_ref text,
  voided_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_cod_remittances_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(remittance_ref), '') IS NOT NULL
    AND nullif(btrim(remittance_no), '') IS NOT NULL
    AND nullif(btrim(carrier_ref), '') IS NOT NULL
    AND nullif(btrim(carrier_name), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_cod_remittances_status CHECK (
    status IN ('draft', 'matching', 'submitted', 'approved', 'discrepancy', 'closed', 'void')
  ),
  CONSTRAINT ck_cod_remittances_money CHECK (
    currency_code = 'VND'
    AND expected_amount > 0
    AND remitted_amount >= 0
    AND discrepancy_amount = remitted_amount - expected_amount
    AND version > 0
  )
);

CREATE TABLE IF NOT EXISTS finance.cod_remittance_lines (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  cod_remittance_id uuid NOT NULL REFERENCES finance.cod_remittances(id) ON DELETE CASCADE,
  line_ref text NOT NULL,
  remittance_ref text NOT NULL,
  receivable_ref text NOT NULL,
  receivable_no text NOT NULL,
  shipment_ref text,
  tracking_no text NOT NULL,
  customer_name text,
  expected_amount numeric(18,2) NOT NULL,
  remitted_amount numeric(18,2) NOT NULL DEFAULT 0,
  discrepancy_amount numeric(18,2) NOT NULL DEFAULT 0,
  match_status text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_cod_remittance_lines_refs CHECK (
    nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(remittance_ref), '') IS NOT NULL
    AND nullif(btrim(receivable_ref), '') IS NOT NULL
    AND nullif(btrim(receivable_no), '') IS NOT NULL
    AND nullif(btrim(tracking_no), '') IS NOT NULL
  ),
  CONSTRAINT ck_cod_remittance_lines_status CHECK (
    match_status IN ('matched', 'short_paid', 'over_paid')
  ),
  CONSTRAINT ck_cod_remittance_lines_money CHECK (
    expected_amount > 0
    AND remitted_amount >= 0
    AND discrepancy_amount = remitted_amount - expected_amount
  )
);

CREATE TABLE IF NOT EXISTS finance.cod_discrepancies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  cod_remittance_id uuid NOT NULL REFERENCES finance.cod_remittances(id) ON DELETE CASCADE,
  cod_remittance_line_id uuid NOT NULL REFERENCES finance.cod_remittance_lines(id) ON DELETE CASCADE,
  discrepancy_ref text NOT NULL,
  remittance_ref text NOT NULL,
  line_ref text NOT NULL,
  receivable_ref text NOT NULL,
  discrepancy_type text NOT NULL,
  status text NOT NULL,
  amount numeric(18,2) NOT NULL,
  reason text NOT NULL,
  owner_ref text NOT NULL,
  recorded_by_ref text NOT NULL,
  recorded_at timestamptz NOT NULL,
  resolved_by_ref text,
  resolved_at timestamptz,
  resolution text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_cod_discrepancies_refs CHECK (
    nullif(btrim(discrepancy_ref), '') IS NOT NULL
    AND nullif(btrim(remittance_ref), '') IS NOT NULL
    AND nullif(btrim(line_ref), '') IS NOT NULL
    AND nullif(btrim(receivable_ref), '') IS NOT NULL
    AND nullif(btrim(reason), '') IS NOT NULL
    AND nullif(btrim(owner_ref), '') IS NOT NULL
    AND nullif(btrim(recorded_by_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_cod_discrepancies_type CHECK (
    discrepancy_type IN ('short_paid', 'over_paid', 'carrier_fee', 'return_claim', 'other')
  ),
  CONSTRAINT ck_cod_discrepancies_status CHECK (status IN ('open', 'resolved')),
  CONSTRAINT ck_cod_discrepancies_money CHECK (amount <> 0)
);

CREATE TABLE IF NOT EXISTS finance.cash_transactions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  org_ref text NOT NULL,
  transaction_ref text NOT NULL,
  transaction_no text NOT NULL,
  direction text NOT NULL,
  status text NOT NULL,
  business_date date NOT NULL,
  counterparty_ref text,
  counterparty_name text NOT NULL,
  payment_method text NOT NULL,
  reference_no text,
  total_amount numeric(18,2) NOT NULL,
  currency_code text NOT NULL DEFAULT 'VND',
  memo text,
  posted_by_ref text,
  posted_at timestamptz,
  void_reason text,
  voided_by_ref text,
  voided_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  created_by_ref text NOT NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  updated_by_ref text NOT NULL,
  version integer NOT NULL DEFAULT 1,
  CONSTRAINT ck_cash_transactions_refs CHECK (
    nullif(btrim(org_ref), '') IS NOT NULL
    AND nullif(btrim(transaction_ref), '') IS NOT NULL
    AND nullif(btrim(transaction_no), '') IS NOT NULL
    AND nullif(btrim(counterparty_name), '') IS NOT NULL
    AND nullif(btrim(payment_method), '') IS NOT NULL
    AND nullif(btrim(created_by_ref), '') IS NOT NULL
    AND nullif(btrim(updated_by_ref), '') IS NOT NULL
  ),
  CONSTRAINT ck_cash_transactions_direction CHECK (direction IN ('cash_in', 'cash_out')),
  CONSTRAINT ck_cash_transactions_status CHECK (status IN ('draft', 'posted', 'void')),
  CONSTRAINT ck_cash_transactions_money CHECK (
    currency_code = 'VND'
    AND total_amount > 0
    AND version > 0
  )
);

CREATE TABLE IF NOT EXISTS finance.cash_transaction_allocations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES core.organizations(id),
  cash_transaction_id uuid NOT NULL REFERENCES finance.cash_transactions(id) ON DELETE CASCADE,
  allocation_ref text NOT NULL,
  transaction_ref text NOT NULL,
  target_type text NOT NULL,
  target_ref text NOT NULL,
  target_no text NOT NULL,
  amount numeric(18,2) NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT ck_cash_transaction_allocations_refs CHECK (
    nullif(btrim(allocation_ref), '') IS NOT NULL
    AND nullif(btrim(transaction_ref), '') IS NOT NULL
    AND nullif(btrim(target_ref), '') IS NOT NULL
    AND nullif(btrim(target_no), '') IS NOT NULL
  ),
  CONSTRAINT ck_cash_transaction_allocations_target CHECK (
    target_type IN (
      'customer_receivable',
      'supplier_payable',
      'cod_remittance',
      'payment_request',
      'manual_adjustment'
    )
  ),
  CONSTRAINT ck_cash_transaction_allocations_money CHECK (amount > 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_customer_receivables_org_ref
  ON finance.customer_receivables(org_id, lower(receivable_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_customer_receivables_org_no
  ON finance.customer_receivables(org_id, lower(receivable_no));

CREATE UNIQUE INDEX IF NOT EXISTS uq_customer_receivable_lines_ref
  ON finance.customer_receivable_lines(org_id, customer_receivable_id, lower(line_ref));

CREATE INDEX IF NOT EXISTS ix_customer_receivables_filters
  ON finance.customer_receivables(org_id, customer_ref, status, due_date, updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_payables_org_ref
  ON finance.supplier_payables(org_id, lower(payable_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_payables_org_no
  ON finance.supplier_payables(org_id, lower(payable_no));

CREATE UNIQUE INDEX IF NOT EXISTS uq_supplier_payable_lines_ref
  ON finance.supplier_payable_lines(org_id, supplier_payable_id, lower(line_ref));

CREATE INDEX IF NOT EXISTS ix_supplier_payables_filters
  ON finance.supplier_payables(org_id, supplier_ref, status, due_date, updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_cod_remittances_org_ref
  ON finance.cod_remittances(org_id, lower(remittance_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_cod_remittances_org_no
  ON finance.cod_remittances(org_id, lower(remittance_no));

CREATE UNIQUE INDEX IF NOT EXISTS uq_cod_remittance_lines_ref
  ON finance.cod_remittance_lines(org_id, cod_remittance_id, lower(line_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_cod_discrepancies_ref
  ON finance.cod_discrepancies(org_id, cod_remittance_id, lower(discrepancy_ref));

CREATE INDEX IF NOT EXISTS ix_cod_remittances_filters
  ON finance.cod_remittances(org_id, carrier_ref, status, business_date, updated_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_cash_transactions_org_ref
  ON finance.cash_transactions(org_id, lower(transaction_ref));

CREATE UNIQUE INDEX IF NOT EXISTS uq_cash_transactions_org_no
  ON finance.cash_transactions(org_id, lower(transaction_no));

CREATE UNIQUE INDEX IF NOT EXISTS uq_cash_transaction_allocations_ref
  ON finance.cash_transaction_allocations(org_id, cash_transaction_id, lower(allocation_ref));

CREATE INDEX IF NOT EXISTS ix_cash_transactions_filters
  ON finance.cash_transactions(org_id, direction, status, business_date, counterparty_ref);

COMMIT;
