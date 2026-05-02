# 68_ERP_Sprint15_Changelog_Finance_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 15 - Finance runtime store persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Release evidence complete; production tag pushed after green CI, dev smoke, and migration gate

---

## 1. Sprint 15 Scope

Sprint 15 closed the highest remaining Finance Lite reset risk after Sprint 14:

```text
customer receivables were prototype-only
supplier payables were prototype-only
COD remittances were prototype-only
cash transactions were prototype-only
finance dashboard/report reads could reset with the prototype runtime stores
```

Promoted scope:

```text
Sprint 15 task board
finance persistence design
finance runtime persistence migration foundation
PostgreSQL-backed customer receivable, supplier payable, COD remittance, and cash transaction stores
service/store lifecycle tests for AR, AP, COD, and cash
package-level runtime finance store selector
dashboard/report integration check against selected runtime stores
full dev finance persistence smoke through API restart
remaining prototype store ledger update
```

No double-entry general ledger, bank integration, tax filing, e-invoice integration, multi-currency, or broad finance UI redesign was introduced. Sprint 15 changed persistence behavior behind the existing Finance Lite APIs.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S15-00-00 Sprint 15 task board | #453 | Created Sprint 15 task board |
| S15-01-01 Finance persistence design | #454 | Documented finance persistence schema, selectors, tests, smoke, and guardrails |
| S15-01-02 Finance migration foundation | #455 | Added migration `000031_persist_finance_runtime_foundation` |
| S15-02-01 Customer receivable PostgreSQL store | #456 | Added PostgreSQL-backed AR header, lines, source refs, receipt, dispute, void, and status persistence |
| S15-02-02 Customer receivable persistence tests | #457 | Added customer receivable PostgreSQL lifecycle coverage |
| S15-03-01 Supplier payable PostgreSQL store | #458 | Added PostgreSQL-backed AP header, lines, source refs, approval/payment, void, and status persistence |
| S15-03-02 Supplier payable persistence tests | #459 | Added supplier payable PostgreSQL lifecycle coverage |
| S15-04-01 COD remittance PostgreSQL store | #460 | Added PostgreSQL-backed COD header, lines, discrepancy, match, submit, approve, close, and status persistence |
| S15-04-02 COD remittance persistence tests | #461 | Added COD remittance PostgreSQL lifecycle coverage |
| S15-05-01 Cash transaction PostgreSQL store | #462 | Added PostgreSQL-backed cash transaction, allocation, source ref, post, void, and status persistence |
| S15-05-02 Cash transaction persistence tests | #463 | Added cash transaction PostgreSQL lifecycle coverage |
| S15-06-01 Package runtime selectors | #464 | Wired finance DB-mode selection as one AR/AP/COD/cash package |
| S15-06-02 Finance dashboard/report integration check | #465 | Verified dashboard and finance report read the same selected runtime finance stores |
| S15-06-03 Finance persistence smoke | #466 | Added full dev smoke for finance docs, API restart, read-back, DB rows, and audit rows |
| S15-07-01 Remaining prototype ledger update | #467 | Superseded Sprint 14 remaining-store ledger and removed finance from production persistence gaps |

All PRs used the manual review and merge flow.

---

## 3. Persistence Changes

### Runtime Selector

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeFinanceStores.customerReceivables` | `PostgresCustomerReceivableStore` | `PrototypeCustomerReceivableStore` |
| `newRuntimeFinanceStores.supplierPayables` | `PostgresSupplierPayableStore` | `PrototypeSupplierPayableStore` |
| `newRuntimeFinanceStores.codRemittances` | `PostgresCODRemittanceStore` | `PrototypeCODRemittanceStore` |
| `newRuntimeFinanceStores.cashTransactions` | `PostgresCashTransactionStore` | `PrototypeCashTransactionStore` |

DB mode selects the four finance stores as one package. Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000031_persist_finance_runtime_foundation` | Adds finance runtime tables, line tables, allocation/source-ref tables, indexes, and constraints for AR, AP, COD, and cash transactions |

Persisted behavior:

```text
GET  /api/v1/customer-receivables
POST /api/v1/customer-receivables
GET  /api/v1/customer-receivables/{customer_receivable_id}
POST /api/v1/customer-receivables/{customer_receivable_id}/record-receipt
POST /api/v1/customer-receivables/{customer_receivable_id}/mark-disputed
POST /api/v1/customer-receivables/{customer_receivable_id}/void
GET  /api/v1/supplier-payables
POST /api/v1/supplier-payables
GET  /api/v1/supplier-payables/{supplier_payable_id}
POST /api/v1/supplier-payables/{supplier_payable_id}/request-payment
POST /api/v1/supplier-payables/{supplier_payable_id}/approve-payment
POST /api/v1/supplier-payables/{supplier_payable_id}/reject-payment
POST /api/v1/supplier-payables/{supplier_payable_id}/record-payment
POST /api/v1/supplier-payables/{supplier_payable_id}/void
GET  /api/v1/cod-remittances
POST /api/v1/cod-remittances
GET  /api/v1/cod-remittances/{cod_remittance_id}
POST /api/v1/cod-remittances/{cod_remittance_id}/match
POST /api/v1/cod-remittances/{cod_remittance_id}/record-discrepancy
POST /api/v1/cod-remittances/{cod_remittance_id}/submit
POST /api/v1/cod-remittances/{cod_remittance_id}/approve
POST /api/v1/cod-remittances/{cod_remittance_id}/close
GET  /api/v1/cash-transactions
POST /api/v1/cash-transactions
GET  /api/v1/cash-transactions/{cash_transaction_id}
GET  /api/v1/finance/dashboard
GET  /api/v1/reports/finance-summary
GET  /api/v1/reports/finance-summary/export.csv
```

Persisted evidence:

```text
finance.customer_receivables
finance.customer_receivable_lines
finance.customer_receivable_source_refs
finance.supplier_payables
finance.supplier_payable_lines
finance.supplier_payable_source_refs
finance.cod_remittances
finance.cod_remittance_lines
finance.cash_transactions
finance.cash_transaction_allocations
audit.audit_logs finance lifecycle actions
```

---

## 4. Dev Release Evidence

Dev server:

```text
Host: 10.1.1.120
Repo: /opt/ERP-v2
Runtime dev URL: http://10.1.1.120:8088
```

Runtime deploy evidence:

```text
After PR #467, main was deployed to dev at commit d87ddcec.
Deploy built API, worker, and web images from source after GHCR dev-image pull warnings.
Migrations through 000031 were present on dev.
Deploy smoke passed.
Full host smoke passed.
```

Release gate at current Sprint 15 main:

```text
Commit: d87ddcec
Command: infra/scripts/dev-release-gate.sh dev
Result: Dev release gate passed
```

Release gate included:

```text
Backend: gofmt, go vet, go test ./..., go build ./cmd/api, go build ./cmd/worker passed.
OpenAPI: redocly lint validated with the existing proprietary-license warning; contract check passed 70 routes and 38 envelopes; generated schema check passed.
Frontend: typecheck passed; 35 Vitest files and 226 tests passed; Next.js production build passed.
Deploy: dev deploy passed after source image build.
Smoke: full dev smoke passed.
```

Latest Sprint 15 smoke evidence on dev:

```text
api_restart                  ok finance-runtime
persisted_finance_ar         ok AR-S15-06-03-SMOKE-0004
persisted_finance_ap         ok AP-S15-06-03-SMOKE-0004
persisted_finance_cod        ok COD-S15-06-03-SMOKE-0004
persisted_finance_cash       ok CASH-IN-S15-06-03-SMOKE-0004
```

The same full smoke also passed the previously persisted audit, sales reservation/order, stock adjustment/movement/count, purchase order, inbound QC, carrier manifest, pick task, pack task, return receipt, and supplier rejection checks.

---

## 5. CI And Migration Evidence

GitHub checks:

```text
PR #453 through PR #467: required checks passed.
```

Local/dev verification highlights:

```text
S15-06-01: git diff --check passed.
S15-06-01: dev server go test ./... passed.
S15-06-01: focused cmd/api runtime selector tests passed.
S15-06-02: git diff --check passed.
S15-06-02: dev server go test ./... passed.
S15-06-02: focused dashboard/report runtime store integration test passed.
S15-06-03: git diff --check passed.
S15-06-03: sh -n infra/scripts/smoke-dev-full.sh passed.
S15-06-03: full dev smoke passed with finance restart persistence evidence.
S15-07-01: git diff --check passed.
S15-07-01: documentation-only change; no runtime test required.
S15-08-01: dev release gate passed at commit d87ddcec.
S15 tag gate: full dev smoke passed at tag commit 87df2da4.
S15 tag gate: PostgreSQL 16 migration apply/rollback passed at tag commit 87df2da4.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
PostgreSQL version: 16.13
Source: /opt/ERP-v2 at main commit 87df2da4
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
Applied migrations: 31
Rolled back migrations: 31
```

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S14-04-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 15:

```text
1. Subcontract runtime stores.
2. Master data catalogs.
3. Auth/session hardening.
```

Finance AR, AP, COD, cash, dashboard, report, and CSV export runtime paths are no longer listed as production persistence gaps when DB config exists.

---

## 7. Release Status

Sprint 15 release gate status:

```text
Task PRs: merged through S15-08-01
Current changelog PR: merged as PR #468
Main CI: green through PR #468
Dev runtime smoke: green at tag commit 87df2da4
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: pushed
```

Production tag:

```text
v0.15.0-finance-runtime-store-persistence
```

Tag commit:

```text
87df2da44f58bc2f9534205ab47e4a5605b51d4f
```

Do not move the tag once pushed. If a post-tag fix is needed, create a new patch tag instead.
