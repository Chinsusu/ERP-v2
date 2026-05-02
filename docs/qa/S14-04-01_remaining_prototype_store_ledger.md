# S17-07-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 17 - Master data runtime store persistence
Task: S17-07-01 Remaining prototype ledger update
Date: 2026-05-02
Status: Inventory complete; S17-08-01 release evidence can use this ledger

---

## 1. Purpose

This ledger supersedes the Sprint 16 state recorded in this same file after the Sprint 17 master data runtime persistence work.

Sprint 17 closed the highest remaining master data reset risk from the Sprint 16 ledger:

```text
item catalog lifecycle, UOM fields, decimal cost, status, and audit state
warehouse and location lifecycle, parent relation, status, and audit state
supplier and customer lifecycle, group/type, terms, metrics, credit fields, and audit state
UOM definitions and conversion factors
package-level master data runtime selection and restart smoke
```

Prototype fallback still exists for no-DB/local mode. That fallback is intentional and must not be counted as production persistence evidence.

---

## 2. Stores Persisted Through Sprint 17

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
| Subcontract runtime package | `newRuntimeSubcontractStores` | PostgreSQL-backed order/transfer/sample/receipt/claim/payment stores selected together when DB config exists; prototype fallback only for no-DB/local | PR #479, PR #480, PR #481, migrations `000032` and `000033`, full dev smoke checks subcontract state after API restart |
| Subcontract orders | `subcontractStores.orders` | PostgreSQL-backed `production.subcontract_orders`, material lines, documents, status events, and action state when DB config exists | PR #472, PR #473, PR #479, PR #480, migrations `000013`, `000032`, and `000033`, service/store lifecycle tests and full dev smoke |
| Subcontract material transfers | `subcontractStores.materialTransfers` | PostgreSQL-backed material transfer headers, lines, source refs, and stock movement refs when DB config exists | PR #474, PR #479, migrations `000032` and `000033`, service/store lifecycle tests and full dev smoke |
| Subcontract sample approvals | `subcontractStores.sampleApprovals` | PostgreSQL-backed sample submit/approve/reject lifecycle when DB config exists | PR #475, PR #479, migrations `000032` and `000033`, service/store lifecycle tests and full dev smoke |
| Subcontract finished goods receipts | `subcontractStores.finishedGoodsReceipts` | PostgreSQL-backed receipt, line, accept/reject, and stock movement ref lifecycle when DB config exists | PR #476, PR #479, migrations `000032` and `000033`, service/store lifecycle tests and full dev smoke |
| Subcontract factory claims | `subcontractStores.factoryClaims` | PostgreSQL-backed defect claim evidence, claim window, status, and quantity impact when DB config exists | PR #477, PR #479, migrations `000032` and `000033`, service/store lifecycle tests and full dev smoke |
| Subcontract payment milestones | `subcontractStores.paymentMilestones` | PostgreSQL-backed deposit/final-payment readiness and supplier payable link evidence when DB config exists | PR #478, PR #479, migrations `000032` and `000033`, service/store lifecycle tests and full dev smoke |
| Warehouse daily board subcontract signals | Warehouse daily board handlers using `subcontractStores` | Reads the selected DB-backed subcontract stores when DB config exists | PR #479 and PR #481, full dev smoke checks subcontract board evidence after API restart |
| Master data runtime package | `newRuntimeMasterDataStores` | PostgreSQL-backed item, UOM, warehouse/location, and party catalogs selected together when DB config exists; prototype fallback only for no-DB/local | PR #488, PR #489, PR #490, PR #491, PR #492, migration `000034`, full dev smoke checks master data state after API restart |
| Item catalog | `masterDataStores.items` | PostgreSQL-backed `mdm.items` lifecycle, duplicate-code checks, UOM fields, standard cost, status, and audit when DB config exists | PR #488, PR #492, migration `000034`, store lifecycle tests and full dev smoke |
| Warehouse/location catalog | `masterDataStores.warehouses` | PostgreSQL-backed `mdm.warehouses` and `mdm.warehouse_bins` lifecycle, hierarchy, status, and audit when DB config exists | PR #489, PR #492, migration `000034`, store lifecycle tests and full dev smoke |
| Party catalog | `masterDataStores.parties` | PostgreSQL-backed `mdm.suppliers` and `mdm.customers` lifecycle, group/type, terms, metrics, credit fields, status, and audit when DB config exists | PR #490, PR #492, migration `000034`, store lifecycle tests and full dev smoke |
| UOM catalog | `masterDataStores.uoms` | PostgreSQL-backed `mdm.uoms` and `mdm.uom_conversions` definitions and conversion factors without float/double | PR #491, PR #492, migration `000034`, UOM conversion tests and full dev smoke |

---

## 3. Sprint 17 Resolution Notes

### Master Data Runtime Stores

| Store / service | Prior Sprint 16 risk | Sprint 17 result |
| --- | --- | --- |
| Item catalog | Item code, SKU, UOM, shelf-life, status, standard cost, and audit edits reset on API restart | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, seed compatibility, and dev smoke |
| Warehouse/location catalog | Warehouse and bin/location edits reset on API restart | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, seed compatibility, and dev smoke |
| Party catalog | Supplier/customer edits reset on API restart | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, seed compatibility, and dev smoke |
| UOM catalog | UOM definitions/conversions reset on API restart | Persisted with PostgreSQL store, conversion tests, package runtime selector, and dev smoke |

### Master Data Runtime Package Rule

S17-06-01 made master data runtime selection a package-level decision. When DB config exists, item, UOM, warehouse/location, and party catalogs share PostgreSQL-backed stores. When DB config is absent, they fall back together to prototype stores for no-DB/local use only.

S17-06 also seeds the existing prototype baseline into PostgreSQL at startup so existing purchase, sales, warehouse, return, subcontract, finance, and reporting flows keep their stable master data references after DB-mode selection changes.

---

## 4. Prior Sprint 16 Resolution Notes

### Subcontract Runtime Stores

| Store / service | Prior Sprint 15 risk | Sprint 16 result |
| --- | --- | --- |
| Subcontract orders | Subcontract order state reset while stock, finance, PO, inbound QC, shipping, and returns state persisted | Persisted with PostgreSQL store, lifecycle tests, runtime selector, shared transaction fix, and dev smoke |
| Subcontract material transfers | Material and packaging issue evidence reset | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, and dev smoke |
| Subcontract sample approvals | Sample approval/rejection evidence reset | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, and dev smoke |
| Subcontract finished goods receipts | Finished goods receipt and acceptance evidence reset | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, and dev smoke |
| Subcontract factory claims | Defect claim evidence and quantity impact reset | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, and dev smoke |
| Subcontract payment milestones | Deposit/final payment readiness and payable link evidence reset | Persisted with PostgreSQL store, lifecycle tests, package runtime selector, and dev smoke |

### Subcontract Runtime Package Rule

S16-08-01 made subcontract runtime selection a package-level decision. When DB config exists, subcontract order, material transfer, sample approval, finished goods receipt, factory claim, payment milestone, and warehouse daily board subcontract reads share the PostgreSQL-backed subcontract stores. When DB config is absent, they fall back together to prototype stores for no-DB/local use only.

S16-08 hardening fixed two release risks found during dev smoke:

```text
1. Subcontract order lifecycle now reuses the parent PostgreSQL transaction when child subcontract stores are invoked during order actions, avoiding same-row lock timeouts.
2. Audit log IDs include a random suffix, avoiding duplicate log_ref conflicts after API restarts with fixed smoke-test business timestamps.
```

---

## 5. Remaining Backend Prototype Stores

### P1 - Auth, Session, and Dev Fallbacks

| Store / service | Current constructor or path | Runtime risk | Recommendation |
| --- | --- | --- | --- |
| Access/refresh sessions | `auth.NewSessionManager(...)` | Sessions and lockout state reset | Acceptable for current mock/dev auth; revisit before production auth |
| Frontend fallback services | `apps/web/src/modules/**/services/*` | Can hide backend failures during UI testing | Keep dev-only; never count frontend fallback as persistence evidence |

---

## 6. Recommended Post-Sprint-17 Persistence Order

```text
1. Auth/session hardening.
2. Remove or gate frontend fallback services where backend coverage is now available.
```

Rationale:

```text
Sprint 17 closed the editable master data reset risk across items, UOMs, warehouses/locations, suppliers, and customers. The next restart risk is production-grade auth/session state; frontend fallback cleanup should follow so UI testing cannot hide backend regressions.
```

---

## 7. Verification Notes

Inventory checks performed:

```text
- Inspected `newRuntimeMasterDataStores` after PR #492.
- Confirmed DB mode uses one PostgreSQL `*sql.DB` and selects item, UOM, warehouse/location, and party catalogs together.
- Confirmed prototype master data catalogs are selected only when DATABASE_URL is empty.
- Confirmed Sprint 17 migration `000034` covers master data runtime fields, indexes, constraints, and seed-compatible references.
- Confirmed full dev smoke after PR #492 creates item, warehouse, location, supplier, and customer records, restarts the API service, then reads them back from PostgreSQL.
- Confirmed PostgreSQL 16.13 isolated migration gate applied 34 up migrations and rolled back 34 down migrations during S17-01-02.
- No runtime behavior changed in this task.
```
