# S12-04-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 12 - Batch QC status persistence
Task: S12-04-01 Remaining prototype ledger update
Date: 2026-05-01
Status: Inventory complete; S12-05-01 release evidence can use this ledger

---

## 1. Purpose

This ledger supersedes `docs/qa/S11-05-01_remaining_prototype_store_ledger.md` after the Sprint 12 batch/QC persistence work.

Sprint 12 closed the next cross-store mismatch from the Sprint 11 ledger:

```text
batch catalog / QC status changes
batch QC transition audit history
inbound QC batch-status updater consistency
batch QC status effects in available-stock reads
```

Prototype fallback still exists for no-DB/local mode. That fallback is intentional and must not be counted as production persistence evidence.

---

## 2. Stores Persisted Through Sprint 12

| Area | Runtime path | Persistence status | Evidence |
| --- | --- | --- | --- |
| Stock movement writer | `newRuntimeStockMovementStore` | PostgreSQL when DB config exists; memory fallback for non-DB/local | Full dev smoke checks `inventory.stock_ledger` and `inventory.stock_balances` |
| Audit log | `newRuntimeAuditLogStore` | PostgreSQL when DB config exists; prototype fallback for no-DB/local | Full dev smoke checks login/audit persistence |
| Sales order reservations | `newRuntimeSalesOrderReservationStore` | PostgreSQL-backed reservation rows | Full dev smoke checks reserve/release rows and audit |
| Stock adjustments | `newRuntimeStockAdjustmentStore` | PostgreSQL-backed document lifecycle | Full dev smoke checks posted adjustment document and stock movement |
| Stock counts | `newRuntimeStockCountStore` | PostgreSQL-backed count session lifecycle | Full dev smoke checks variance-review document and audit |
| Warehouse receiving | `newRuntimeWarehouseReceivingStore` | PostgreSQL-backed receiving document lifecycle | Inbound QC full smoke uses persisted receiving evidence |
| Inbound QC | `newRuntimeInboundQCInspectionStore` | PostgreSQL-backed inspection/checklist lifecycle | S12-03-02 smoke checks PASS/FAIL/PARTIAL decisions and persisted audit |
| Available stock read model | `newRuntimeStockAvailabilityStore` | PostgreSQL-backed reads from `inventory.stock_balances` when DB config exists | S12-02-02 smoke checks persisted batch QC status effects |
| Sales orders | `newRuntimeSalesOrderStore` | PostgreSQL-backed owner documents when DB config exists | Sprint 11 release evidence and dev smoke |
| Purchase orders | `newRuntimePurchaseOrderStore` | PostgreSQL-backed owner documents when DB config exists | Sprint 11 release evidence and inbound trace |
| Return receipts | `newRuntimeReturnReceiptStore` | PostgreSQL-backed receipt, line, inspection, disposition, and attachment refs when DB config exists | Sprint 11 release evidence and dev smoke |
| Supplier rejections | `newRuntimeSupplierRejectionStore` | PostgreSQL-backed rejection header, line, attachment, and status lifecycle when DB config exists | Sprint 11 release evidence and dev smoke |
| Batch catalog / QC status | `newRuntimeBatchCatalogStore` | PostgreSQL-backed `inventory.batches` reads/writes when DB config exists; prototype fallback for no-DB/local | S12-01-03, S12-01-04, S12-02-01, S12-02-02, S12-03-01, and S12-03-02 evidence |

---

## 3. Remaining Backend Prototype Stores

### P1 - Operational Evidence Stores

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| End-of-day reconciliation | `inventoryapp.NewPrototypeEndOfDayReconciliationStore()` | Shift close and variance evidence reset | Persist before relying on shift-close history as release evidence |
| Carrier manifests | `shippingapp.NewPrototypeCarrierManifestStore()` | Handover and scan evidence reset | Persist manifest/header/line/scan exception state |
| Pick tasks | `shippingapp.NewPrototypePickTaskStore(...)` | Pick progress resets while sales orders and reservations persist | Persist with shipping task package |
| Pack tasks | `shippingapp.NewPrototypePackTaskStore(...)` | Pack progress resets while sales orders and reservations persist | Persist with pick/manifest work |
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

## 4. Recommended Post-Sprint-12 Persistence Order

```text
1. End-of-day reconciliation.
2. Shipping manifest, pick, and pack task package.
3. Finance AR/AP/COD/cash runtime stores.
4. Subcontract runtime stores.
5. Master data catalogs and auth/session hardening.
```

Rationale:

```text
Sprint 12 closed the batch/QC mismatch between inbound QC, available stock, and inventory.batches. The next restart risks are warehouse execution evidence, finance state, subcontract lifecycle state, and editable catalogs.
```

---

## 5. Verification Notes

Inventory checks performed:

```text
- Inspected current main.go runtime constructors after PR #430.
- Inspected runtime selector files for PostgreSQL/prototype fallback behavior.
- Compared against the Sprint 11 remaining prototype store ledger.
- Confirmed Sprint 12 batch/QC evidence from S12-01-03 through S12-03-02.
- No runtime behavior changed in this task.
```
