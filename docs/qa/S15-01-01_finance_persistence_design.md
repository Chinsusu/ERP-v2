# S15-01-01 Finance Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 15 - Finance runtime store persistence
Task: S15-01-01 Finance persistence design
Date: 2026-05-02
Status: Design complete; S15-01-02 can introduce the finance migration foundation

---

## 1. Goal

Promote Finance Lite runtime state from prototype memory to PostgreSQL-backed persistence when database configuration exists.

Current restart risk:

```text
customer receivables -> financeapp.NewPrototypeCustomerReceivableStore()
supplier payables    -> financeapp.NewPrototypeSupplierPayableStore()
COD remittances      -> financeapp.NewPrototypeCODRemittanceStore()
cash transactions    -> financeapp.NewPrototypeCashTransactionStore()
finance dashboard    -> reads those runtime stores
finance reports      -> reads those runtime stores
```

Sprint 15 should make the existing Finance Lite APIs durable without changing public response envelopes or introducing full general-ledger behavior.

Important sequencing rule:

```text
Do not wire only one finance store into DB mode while other finance stores remain memory-backed.
Implement and test stores separately, but switch runtime selection to DB mode as one AR/AP/COD/cash package.
```

---

## 2. Existing Contracts

Customer receivable store:

```text
CustomerReceivableStore
  List(ctx, filter) []domain.CustomerReceivable
  Get(ctx, id) domain.CustomerReceivable
  Save(ctx, receivable) error
```

Supplier payable store:

```text
SupplierPayableStore
  List(ctx, filter) []domain.SupplierPayable
  Get(ctx, id) domain.SupplierPayable
  Save(ctx, payable) error
```

COD remittance store:

```text
CODRemittanceStore
  List(ctx, filter) []domain.CODRemittance
  Get(ctx, id) domain.CODRemittance
  Save(ctx, remittance) error
```

Cash transaction store:

```text
CashTransactionStore
  List(ctx, filter) []domain.CashTransaction
  Get(ctx, id) domain.CashTransaction
  Save(ctx, transaction) error
```

Current public routes must keep their response shape:

```text
GET  /api/v1/customer-receivables
POST /api/v1/customer-receivables
GET  /api/v1/customer-receivables/{id}
POST /api/v1/customer-receivables/{id}/record-receipt
POST /api/v1/customer-receivables/{id}/mark-disputed
POST /api/v1/customer-receivables/{id}/void

GET  /api/v1/supplier-payables
POST /api/v1/supplier-payables
GET  /api/v1/supplier-payables/{id}
POST /api/v1/supplier-payables/{id}/request-payment
POST /api/v1/supplier-payables/{id}/approve-payment
POST /api/v1/supplier-payables/{id}/reject-payment
POST /api/v1/supplier-payables/{id}/record-payment
POST /api/v1/supplier-payables/{id}/void

GET  /api/v1/cod-remittances
POST /api/v1/cod-remittances
GET  /api/v1/cod-remittances/{id}
POST /api/v1/cod-remittances/{id}/match
POST /api/v1/cod-remittances/{id}/record-discrepancy
POST /api/v1/cod-remittances/{id}/submit
POST /api/v1/cod-remittances/{id}/approve
POST /api/v1/cod-remittances/{id}/close

GET  /api/v1/cash-transactions
POST /api/v1/cash-transactions
GET  /api/v1/cash-transactions/{id}
GET  /api/v1/finance/dashboard
GET  /api/v1/reports/finance-summary
GET  /api/v1/reports/finance-summary.csv
```

Do not change domain transition rules. Persistence is an adapter concern.

---

## 3. PostgreSQL Source Tables

Add a `finance` schema because current migrations do not create durable finance runtime tables.

Recommended tables:

```text
finance.customer_receivables
finance.customer_receivable_lines
finance.supplier_payables
finance.supplier_payable_lines
finance.cod_remittances
finance.cod_remittance_lines
finance.cod_discrepancies
finance.cash_transactions
finance.cash_transaction_allocations
```

Use UUID primary keys for database identity and text runtime refs for existing public/domain IDs.

```text
id uuid primary key default gen_random_uuid()
<document>_ref text not null
org_id uuid not null references core.organizations(id)
org_ref text not null
created_by_ref text not null
updated_by_ref text not null
version integer not null default 1
```

The stores should lookup by text ref first and UUID id fallback:

```text
WHERE lower(COALESCE(<document>_ref, id::text)) = lower($1)
   OR id::text = $1
```

---

## 4. Table Shape

### Customer Receivables

Header:

```text
receivable_ref text
receivable_no text
customer_ref text
customer_code text
customer_name text
status text
source_document_type text
source_document_ref text
source_document_no text
total_amount numeric(18,2)
paid_amount numeric(18,2)
outstanding_amount numeric(18,2)
currency_code text
due_date date
dispute_reason text
disputed_by_ref text
disputed_at timestamptz
void_reason text
voided_by_ref text
voided_at timestamptz
last_receipt_by_ref text
last_receipt_at timestamptz
created_at timestamptz
created_by_ref text
updated_at timestamptz
updated_by_ref text
version integer
```

Lines:

```text
line_ref text
receivable_ref text
description text
source_document_type text
source_document_ref text
source_document_no text
amount numeric(18,2)
```

### Supplier Payables

Header:

```text
payable_ref text
payable_no text
supplier_ref text
supplier_code text
supplier_name text
status text
source_document_type text
source_document_ref text
source_document_no text
total_amount numeric(18,2)
paid_amount numeric(18,2)
outstanding_amount numeric(18,2)
currency_code text
due_date date
payment_requested_by_ref text
payment_requested_at timestamptz
payment_approved_by_ref text
payment_approved_at timestamptz
payment_rejected_by_ref text
payment_rejected_at timestamptz
payment_reject_reason text
dispute_reason text
disputed_by_ref text
disputed_at timestamptz
void_reason text
voided_by_ref text
voided_at timestamptz
last_payment_by_ref text
last_payment_at timestamptz
created_at timestamptz
created_by_ref text
updated_at timestamptz
updated_by_ref text
version integer
```

Lines:

```text
line_ref text
payable_ref text
description text
source_document_type text
source_document_ref text
source_document_no text
amount numeric(18,2)
```

### COD Remittances

Header:

```text
remittance_ref text
remittance_no text
carrier_ref text
carrier_code text
carrier_name text
status text
business_date date
expected_amount numeric(18,2)
remitted_amount numeric(18,2)
discrepancy_amount numeric(18,2)
currency_code text
submitted_by_ref text
submitted_at timestamptz
approved_by_ref text
approved_at timestamptz
closed_by_ref text
closed_at timestamptz
void_reason text
voided_by_ref text
voided_at timestamptz
created_at timestamptz
created_by_ref text
updated_at timestamptz
updated_by_ref text
version integer
```

Lines:

```text
line_ref text
remittance_ref text
receivable_ref text
receivable_no text
shipment_ref text
tracking_no text
customer_name text
expected_amount numeric(18,2)
remitted_amount numeric(18,2)
discrepancy_amount numeric(18,2)
match_status text
```

Discrepancies:

```text
discrepancy_ref text
remittance_ref text
line_ref text
receivable_ref text
discrepancy_type text
status text
amount numeric(18,2)
reason text
owner_ref text
recorded_by_ref text
recorded_at timestamptz
resolved_by_ref text
resolved_at timestamptz
resolution text
```

### Cash Transactions

Header:

```text
transaction_ref text
transaction_no text
direction text
status text
business_date date
counterparty_ref text
counterparty_name text
payment_method text
reference_no text
total_amount numeric(18,2)
currency_code text
memo text
posted_by_ref text
posted_at timestamptz
void_reason text
voided_by_ref text
voided_at timestamptz
created_at timestamptz
created_by_ref text
updated_at timestamptz
updated_by_ref text
version integer
```

Allocations:

```text
allocation_ref text
transaction_ref text
target_type text
target_ref text
target_no text
amount numeric(18,2)
```

---

## 5. Constraints And Indexes

Core constraints:

```text
currency_code = 'VND'
amount columns use numeric(18,2)
status columns constrained to current domain statuses
direction constrained to cash_in/cash_out
non-empty runtime refs
header totals equal line/allocation totals at store validation layer
```

Recommended unique indexes:

```text
unique finance.customer_receivables(org_id, lower(receivable_ref))
unique finance.customer_receivables(org_id, lower(receivable_no))
unique finance.customer_receivable_lines(org_id, customer_receivable_id, lower(line_ref))

unique finance.supplier_payables(org_id, lower(payable_ref))
unique finance.supplier_payables(org_id, lower(payable_no))
unique finance.supplier_payable_lines(org_id, supplier_payable_id, lower(line_ref))

unique finance.cod_remittances(org_id, lower(remittance_ref))
unique finance.cod_remittances(org_id, lower(remittance_no))
unique finance.cod_remittance_lines(org_id, cod_remittance_id, lower(line_ref))
unique finance.cod_discrepancies(org_id, cod_remittance_id, lower(discrepancy_ref))

unique finance.cash_transactions(org_id, lower(transaction_ref))
unique finance.cash_transactions(org_id, lower(transaction_no))
unique finance.cash_transaction_allocations(org_id, cash_transaction_id, lower(allocation_ref))
```

Recommended filter indexes:

```text
customer receivables: org_id, customer_ref, status, due_date, updated_at
supplier payables: org_id, supplier_ref, status, due_date, updated_at
COD remittances: org_id, carrier_ref, status, business_date, updated_at
cash transactions: org_id, direction, status, business_date, counterparty_ref
```

---

## 6. Domain Mapping

Mapping rule:

```text
domain ID -> COALESCE(<document>_ref, id::text)
line ID -> COALESCE(line_ref, id::text)
external source IDs -> text refs, not required UUID foreign keys
actor IDs -> *_by_ref text columns
money -> numeric(18,2) stored in DB and returned as decimal string
dates -> due/business dates as date, action timestamps as timestamptz
```

Do not require persisted party, carrier, shipment, payable, receivable, or user UUIDs to write finance records. Sprint 15 persistence should preserve the current text-ref behavior because master data and auth/session stores are still prototype or dev-level.

---

## 7. Store Behavior

Add PostgreSQL store types in `apps/api/internal/modules/finance/application`:

```text
PostgresCustomerReceivableStore
PostgresSupplierPayableStore
PostgresCODRemittanceStore
PostgresCashTransactionStore
```

Read behavior:

```text
List/Get load header and children.
Apply existing prototype-compatible filtering for search, status, customer/supplier/carrier/counterparty, direction.
Map sql.ErrNoRows to the existing not-found errors.
Use text refs first and UUID id fallback.
Return cloned domain values where applicable.
```

Write behavior:

```text
Save validates the domain object first.
Upsert header by org_id + lower(ref).
Upsert child rows by parent + lower(line/allocation/discrepancy ref).
Delete stale child rows that are absent from the incoming aggregate.
Use one transaction per aggregate Save.
Do not write audit logs inside stores.
```

Stable child refs matter because COD lines, discrepancies, cash allocations, and finance reports need traceable line evidence after repeated status updates.

---

## 8. Runtime Selection

Add package-level selector:

```text
newRuntimeFinanceStores(cfg)
```

Selection rule:

```text
DATABASE_URL present -> all four PostgreSQL stores opened together
DATABASE_URL empty -> existing prototype stores
```

The selector should return all stores and one cleanup function:

```text
type runtimeFinanceStores struct {
  customerReceivables financeapp.CustomerReceivableStore
  supplierPayables    financeapp.SupplierPayableStore
  codRemittances      financeapp.CODRemittanceStore
  cashTransactions    financeapp.CashTransactionStore
}
```

DB-mode must fail closed:

```text
if any finance PostgreSQL store cannot be initialized, return error and do not fall back to partial memory stores
```

This prevents mixed finance truth.

---

## 9. Audit And Service Behavior

Keep existing application audit actions in services. Stores must not write duplicate audit logs.

Audit should remain after successful store writes:

```text
customer_receivable created/receipt/disputed/void
supplier_payable created/payment_requested/payment_approved/payment_rejected/payment_recorded/void
cod_remittance created/matched/discrepancy_recorded/submitted/approved/closed
cash_transaction created
```

If a store write succeeds and audit write fails, keep the current service behavior unless a task explicitly changes cross-store transaction handling. Do not invent distributed transactions in Sprint 15.

---

## 10. Tests And Smoke

Focused backend tests:

```text
nil DB returns explicit database-required errors
runtime selector chooses prototype without DATABASE_URL and all PostgreSQL stores with DATABASE_URL
customer receivable Save/Get/List persists create, receipt, dispute, void, and line refs
supplier payable Save/Get/List persists payment request, approve/reject, payment, void, and line refs
COD remittance Save/Get/List persists match, discrepancy, submit, approve, close, line refs, and discrepancy refs
cash transaction Save/Get/List persists direction, posted state, allocations, and source refs
fresh store instance can reload records saved by previous store instance
finance dashboard and report read DB-backed stores after runtime selector switch
```

Migration checks:

```text
apply all migrations on PostgreSQL 16
roll back S15 migration on PostgreSQL 16
rerun apply after rollback when practical
```

Dev smoke:

```text
AR:
  create receivable
  record partial or full receipt
  restart/redeploy API
  confirm status, paid/outstanding amounts, line refs, and audit remain

AP:
  create payable
  request payment
  approve payment
  record payment
  restart/redeploy API
  confirm status, paid/outstanding amounts, line refs, and audit remain

COD:
  create remittance
  match or record discrepancy
  submit/approve/close
  restart/redeploy API
  confirm line, discrepancy, status, and audit remain

Cash:
  create cash in/out transaction with allocations
  restart/redeploy API
  confirm allocation and dashboard/report values remain
```

---

## 11. Implementation Split

Recommended task split:

```text
S15-01-02 finance migration foundation
S15-02-01 customer receivable PostgreSQL store
S15-02-02 customer receivable persistence tests
S15-03-01 supplier payable PostgreSQL store
S15-03-02 supplier payable persistence tests
S15-04-01 COD remittance PostgreSQL store
S15-04-02 COD remittance persistence tests
S15-05-01 cash transaction PostgreSQL store
S15-05-02 cash transaction persistence tests
S15-06-01 package runtime selectors
S15-06-02 finance dashboard/report integration check
S15-06-03 finance persistence smoke
S15-07-01 remaining prototype ledger update
S15-08-01 Sprint 15 changelog and release evidence
```

Implementation guardrails:

```text
No public API response shape change.
No frontend fallback counted as persistence.
No direct money balance mutation outside finance domain/service methods.
No partial DB-mode finance selection.
No broad finance refactor.
No cosmetic formatting mixed into behavior changes.
No production tag until CI, dev smoke, and migration apply/rollback evidence are green.
```
