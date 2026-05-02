# S13-03-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 13 - End-of-day reconciliation persistence
Task: S13-03-01 Remaining prototype ledger update
Date: 2026-05-02
Status: Inventory complete; S13-04-01 release evidence can use this ledger

---

## 1. Purpose

This ledger supersedes `docs/qa/S12-04-01_remaining_prototype_store_ledger.md` after the Sprint 13 end-of-day reconciliation persistence work.

Sprint 13 closed the highest remaining warehouse control evidence risk from the Sprint 12 ledger:

```text
end-of-day reconciliation list/close
shift close status
closed_at / closed_by evidence
exception note
checklist and variance line evidence
warehouse.shift.closed audit record
```

Prototype fallback still exists for no-DB/local mode. That fallback is intentional and must not be counted as production persistence evidence.

---

## 2. Stores Persisted Through Sprint 13

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
| Sales orders | `newRuntimeSalesOrderStore` | PostgreSQL-backed owner documents when DB config exists | Sprint 11 release evidence and dev smoke |
| Purchase orders | `newRuntimePurchaseOrderStore` | PostgreSQL-backed owner documents when DB config exists | Sprint 11 release evidence and inbound trace |
| Return receipts | `newRuntimeReturnReceiptStore` | PostgreSQL-backed receipt, line, inspection, disposition, and attachment refs when DB config exists | Sprint 11 release evidence and dev smoke |
| Supplier rejections | `newRuntimeSupplierRejectionStore` | PostgreSQL-backed rejection header, line, attachment, and status lifecycle when DB config exists | Sprint 11 release evidence and dev smoke |
| Batch catalog / QC status | `newRuntimeBatchCatalogStore` | PostgreSQL-backed `inventory.batches` reads/writes when DB config exists; prototype fallback for no-DB/local | Sprint 12 design, tests, smoke, and release evidence |
| End-of-day reconciliation | `newRuntimeEndOfDayReconciliationStore` | PostgreSQL-backed `inventory.warehouse_daily_closings`, checklist, and line evidence when DB config exists; prototype fallback for no-DB/local | S13-01-02/S13-01-03 runtime store and migration, S13-01-04 focused tests, S13-02-01 dev smoke |

---

## 3. Remaining Backend Prototype Stores

### P1 - Warehouse Execution Stores

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Carrier manifests | `shippingapp.NewPrototypeCarrierManifestStore()` | Handover and scan evidence reset | Persist manifest/header/line/scan exception state |
| Pick tasks | `shippingapp.NewPrototypePickTaskStore(...)` | Pick progress resets while sales orders and reservations persist | Persist with shipping task package |
| Pack tasks | `shippingapp.NewPrototypePackTaskStore(...)` | Pack progress resets while sales orders and reservations persist | Persist with pick/manifest work |

### P1 - Subcontract Runtime Stores

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Subcontract orders | `productionapp.NewPrototypeSubcontractOrderStore(auditLogStore)` | Gia cong order state resets | Persist before expanding subcontract finance/reporting |
| Subcontract material transfers | `productionapp.NewPrototypeSubcontractMaterialTransferStore()` | NVL/bao bi transfer evidence resets | Persist with subcontract order lifecycle |
| Subcontract samples / receipts / claims / payment milestones | Prototype subcontract stores | Sample/QC/claim/final payment state resets | Persist as one subcontract runtime package or in dependency order |

### P1 - Finance Runtime Stores

Finance runtime stores should be promoted together enough to avoid partial financial truth.

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Customer receivables | `financeapp.NewPrototypeCustomerReceivableStore()` | AR status and receipts reset while sales order documents persist | Persist with sales AR flow |
| Supplier payables | `financeapp.NewPrototypeSupplierPayableStore()` | AP/payment approval state resets while PO, supplier rejection, and subcontract evidence can persist | Persist with PO/subcontract payable flows |
| COD remittances | `financeapp.NewPrototypeCODRemittanceStore()` | COD match/discrepancy/approval state resets | Persist after receivables foundation |
| Cash transactions | `financeapp.NewPrototypeCashTransactionStore()` | Cash movement evidence resets | Persist with finance audit/reporting gate |

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

## 4. Recommended Post-Sprint-13 Persistence Order

```text
1. Shipping manifest, pick, and pack task package.
2. Finance AR/AP/COD/cash runtime stores.
3. Subcontract runtime stores.
4. Master data catalogs and auth/session hardening.
```

Rationale:

```text
Sprint 13 closed daily warehouse control evidence. The next restart risks are shipping execution evidence, finance state, subcontract lifecycle state, and editable catalogs.
```

---

## 5. Verification Notes

Inventory checks performed:

```text
- Inspected current main.go runtime constructors after PR #437.
- Confirmed end-of-day reconciliation now uses newRuntimeEndOfDayReconciliationStore.
- Confirmed DB mode uses PostgresEndOfDayReconciliationStore with prototype fallback when DATABASE_URL is empty.
- Confirmed S13-02-01 smoke evidence proves close evidence survives API restart.
- No runtime behavior changed in this task.
```
