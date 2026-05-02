# S15-07-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 15 - Finance runtime store persistence
Task: S15-07-01 Remaining prototype ledger update
Date: 2026-05-02
Status: Inventory complete; S15-08-01 release evidence can use this ledger

---

## 1. Purpose

This ledger supersedes the Sprint 14 state recorded in `docs/qa/S14-04-01_remaining_prototype_store_ledger.md` after the Sprint 15 finance runtime persistence work.

Sprint 15 closed the highest remaining finance reset risk from the Sprint 14 ledger:

```text
customer receivable header, lines, receipts, dispute, void, and status state
supplier payable header, lines, payment request, approval, payment, void, and status state
COD remittance header, lines, discrepancy, match, submit, approve, close, and status state
cash transaction header, allocations, source refs, post/void status, and audit state
finance dashboard, summary report, and CSV export runtime reads
```

Prototype fallback still exists for no-DB/local mode. That fallback is intentional and must not be counted as production persistence evidence.

---

## 2. Stores Persisted Through Sprint 15

| Area | Runtime path | Persistence status | Evidence |
| --- | --- | --- | --- |
| Stock movement writer | `newRuntimeStockMovementStore` | PostgreSQL when DB config exists; memory fallback for non-DB/local | Full dev smoke checks `inventory.stock_ledger` and `inventory.stock_balances` |
| Audit log | `newRuntimeAuditLogStore` | PostgreSQL when DB config exists; prototype fallback for no-DB/local | Full dev smoke checks login/audit persistence |
| Sales order reservations | `newRuntimeSalesOrderReservationStore` | PostgreSQL-backed reservation rows | Full dev smoke checks reserve/release rows and audit |
| Stock adjustments | `newRuntimeStockAdjustmentStore` | PostgreSQL-backed document lifecycle | Full dev smoke checks posted adjustment document and stock movement |
| Stock counts | `newRuntimeStockCountStore` | PostgreSQL-backed count session lifecycle | Full dev smoke checks variance-review document and audit |
| Warehouse receiving | `newRuntimeWarehouseReceivingStore` | PostgreSQL-backed receiving document lifecycle | Inbound QC full smoke uses persisted receiving evidence |
| Inbound QC | `newRuntimeInboundQCInspectionStore` | PostgreSQL-backed inspection/checklist lifecycle | Sprint 12 and full dev smoke evidence |
| Available stock read model | `newRuntimeStockAvailabilityStore` | PostgreSQL-backed reads from `inventory.stock_balances` when DB config exists | Sprint 12 available-stock consistency smoke |
| Sales orders | `newRuntimeSalesOrderStore` | PostgreSQL-backed owner documents when DB config exists | Sprint 11 release evidence, dev smoke, and S14 pack integration check |
| Purchase orders | `newRuntimePurchaseOrderStore` | PostgreSQL-backed owner documents when DB config exists | Sprint 11 release evidence and inbound trace |
| Return receipts | `newRuntimeReturnReceiptStore` | PostgreSQL-backed receipt, line, inspection, disposition, and attachment refs when DB config exists | Sprint 11 release evidence and dev smoke |
| Supplier rejections | `newRuntimeSupplierRejectionStore` | PostgreSQL-backed rejection header, line, attachment, and status lifecycle when DB config exists | Sprint 11 release evidence and dev smoke |
| Batch catalog / QC status | `newRuntimeBatchCatalogStore` | PostgreSQL-backed `inventory.batches` reads/writes when DB config exists; prototype fallback for no-DB/local | Sprint 12 design, tests, smoke, and release evidence |
| End-of-day reconciliation | `newRuntimeEndOfDayReconciliationStore` | PostgreSQL-backed `inventory.warehouse_daily_closings`, checklist, and line evidence when DB config exists; prototype fallback for no-DB/local | S13-01-02/S13-01-03 runtime store and migration, S13-01-04 focused tests, S13-02-01 dev smoke |
| Carrier manifests | `newRuntimeCarrierManifestStore` | PostgreSQL-backed `shipping.carrier_manifests`, shipments, scans, and exceptions when DB config exists; prototype fallback for no-DB/local | PR #443, PR #447, migration `000027`, full dev smoke checks persisted handover/scan evidence |
| Pick tasks | `newRuntimePickTaskStore` | PostgreSQL-backed `shipping.pick_tasks`, pick lines, and exceptions when DB config exists; prototype fallback for no-DB/local | PR #444, PR #446, PR #448, migrations `000028` and `000030`, full dev smoke checks persisted pick completion |
| Pack tasks | `newRuntimePackTaskStore` | PostgreSQL-backed `shipping.pack_tasks`, pack lines, and exceptions when DB config exists; prototype fallback for no-DB/local | PR #445, PR #446, PR #448, PR #449, migrations `000029` and `000030`, full dev smoke checks persisted pack confirmation and sales order packed state |
| Finance runtime package | `newRuntimeFinanceStores` | PostgreSQL-backed AR/AP/COD/cash stores selected together when DB config exists; prototype fallback only for no-DB/local | PR #464, migration `000031`, full dev smoke checks finance state after API restart |
| Customer receivables | `financeStores.customerReceivables` | PostgreSQL-backed `finance.customer_receivables` and `finance.customer_receivable_lines` when DB config exists | PR #456, PR #457, PR #464, PR #466, migration `000031`, service/store lifecycle tests |
| Supplier payables | `financeStores.supplierPayables` | PostgreSQL-backed `finance.supplier_payables` and `finance.supplier_payable_lines` when DB config exists | PR #458, PR #459, PR #464, PR #466, migration `000031`, service/store lifecycle tests |
| COD remittances | `financeStores.codRemittances` | PostgreSQL-backed `finance.cod_remittances` and `finance.cod_remittance_lines` when DB config exists | PR #460, PR #461, PR #464, PR #466, migration `000031`, service/store lifecycle tests |
| Cash transactions | `financeStores.cashTransactions` | PostgreSQL-backed `finance.cash_transactions` and `finance.cash_transaction_allocations` when DB config exists | PR #462, PR #463, PR #464, PR #466, migration `000031`, service/store lifecycle tests |
| Finance dashboard/report/export | `NewFinanceDashboardService`, `financeSummaryReportHandler`, `financeSummaryCSVExportHandler` | Reads the selected runtime finance store package instead of independent prototype stores | PR #464, PR #465, PR #466, dashboard/report integration test and full dev smoke |

---

## 3. Sprint 15 Resolution Notes

### Finance Runtime Stores

| Store / service | Prior Sprint 14 risk | Sprint 15 result |
| --- | --- | --- |
| Customer receivables | AR status and receipts reset while sales order documents persist | Persisted with PostgreSQL store, lifecycle tests, runtime selector, and dev smoke |
| Supplier payables | AP/payment approval state resets while PO, supplier rejection, and subcontract evidence can persist | Persisted with PostgreSQL store, lifecycle tests, runtime selector, and dev smoke |
| COD remittances | COD match/discrepancy/approval state resets | Persisted with PostgreSQL store, lifecycle tests, runtime selector, and dev smoke |
| Cash transactions | Cash movement evidence resets | Persisted with PostgreSQL store, lifecycle tests, runtime selector, and dev smoke |

### Finance Runtime Package Rule

S15-06-01 made finance runtime selection a package-level decision. When DB config exists, AR, AP, COD, cash, dashboard, summary report, and CSV export share the PostgreSQL-backed finance stores. When DB config is absent, they fall back together to prototype stores for no-DB/local use only.

---

## 4. Remaining Backend Prototype Stores

### P1 - Subcontract Runtime Stores

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Subcontract orders | `productionapp.NewPrototypeSubcontractOrderStore(auditLogStore)` | Subcontract order state resets | Persist before expanding subcontract finance/reporting |
| Subcontract material transfers | `productionapp.NewPrototypeSubcontractMaterialTransferStore()` | Material and packaging transfer evidence resets | Persist with subcontract order lifecycle |
| Subcontract samples / receipts / claims / payment milestones | Prototype subcontract stores | Sample/QC/claim/final payment state resets | Persist as one subcontract runtime package or in dependency order |

### P2 - Master Data, Auth, and Dev Fallbacks

| Store / service | Current constructor or path | Runtime risk | Recommendation |
| --- | --- | --- | --- |
| Item catalog | `masterdataapp.NewPrototypeItemCatalog(auditLogStore)` | MDM edits reset | Persist when MDM editing becomes a primary workflow |
| Warehouse/location catalog | `masterdataapp.NewPrototypeWarehouseLocationCatalog(auditLogStore)` | Location edits reset | Persist before more warehouse layout features |
| Party catalog | `masterdataapp.NewPrototypePartyCatalog(auditLogStore)` | Supplier/customer edits reset | Persist before supplier/customer maintenance workflows |
| UOM catalog | `masterdataapp.NewPrototypeUOMCatalog()` | UOM edits reset | Lower risk while standards are mostly static |
| Access/refresh sessions | `auth.NewSessionManager(...)` | Sessions and lockout state reset | Acceptable for current mock/dev auth; revisit before production auth |
| Frontend fallback services | `apps/web/src/modules/**/services/*` | Can hide backend failures during UI testing | Keep dev-only; never count frontend fallback as persistence evidence |

---

## 5. Recommended Post-Sprint-15 Persistence Order

```text
1. Subcontract runtime stores.
2. Master data catalogs.
3. Auth/session hardening.
```

Rationale:

```text
Sprint 15 closed the finance runtime reset risk across AR, AP, COD, cash, dashboard, and report/export reads. The next restart risks are subcontract lifecycle state, editable catalogs, and production-grade auth/session state.
```

---

## 6. Verification Notes

Inventory checks performed:

```text
- Inspected current main.go runtime constructors after PR #466.
- Confirmed DB mode uses newRuntimeFinanceStores for customer receivables, supplier payables, COD remittances, cash transactions, dashboard, summary report, and CSV export.
- Confirmed finance runtime selection uses PostgreSQL stores when DATABASE_URL exists and prototype fallback when DATABASE_URL is empty.
- Confirmed Sprint 15 migration 000031 covers customer receivables, supplier payables, COD remittances, cash transactions, lines, allocations, and source refs.
- Confirmed full dev smoke creates AR/AP/COD/cash documents, restarts the API service, reads them back, and checks finance DB/audit rows.
- No runtime behavior changed in this task.
```
