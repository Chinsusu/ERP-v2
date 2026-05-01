# S11-05-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Task: S11-05-01 Remaining prototype store ledger update
Date: 2026-05-01
Status: Superseded after Sprint 12 by `docs/qa/S12-04-01_remaining_prototype_store_ledger.md`

---

## 1. Purpose

This ledger supersedes `docs/qa/S10-05-01_remaining_prototype_store_ledger.md` after the Sprint 11 persistence work.

Sprint 11 closed the P0 mismatch group from the Sprint 10 ledger:

```text
available stock read model
sales order owner documents
purchase order owner documents
return receipt owner documents
supplier rejection owner documents
```

The remaining prototype stores are lower priority than the Sprint 10 P0 group, but they can still reset on API restart and must not be treated as persisted evidence.

---

## 2. Stores Persisted Through Sprint 11

| Area | Runtime path | Persistence status | Evidence |
| --- | --- | --- | --- |
| Stock movement writer | `newRuntimeStockMovementStore` | PostgreSQL when DB config exists; memory fallback for non-DB/local | Full dev smoke checks `inventory.stock_ledger` and `inventory.stock_balances` |
| Audit log | `newRuntimeAuditLogStore` | PostgreSQL when DB config exists; prototype fallback for no DB | Full dev smoke checks login audit count increases |
| Sales order reservations | `newRuntimeSalesOrderReservationStore` | PostgreSQL-backed reservation rows | Full dev smoke checks reserve/release rows and audit |
| Stock adjustments | `newRuntimeStockAdjustmentStore` | PostgreSQL-backed document lifecycle | Full dev smoke checks posted adjustment document and stock movement |
| Stock counts | `newRuntimeStockCountStore` | PostgreSQL-backed count session lifecycle | Full dev smoke checks variance-review document and audit |
| Warehouse receiving | `newRuntimeWarehouseReceivingStore` | PostgreSQL-backed receiving document lifecycle | Inbound QC full smoke uses persisted receiving evidence |
| Inbound QC | `newRuntimeInboundQCInspectionStore` | PostgreSQL-backed inspection/checklist lifecycle | Full dev smoke checks QC partial decision, audit, and stock ledger rows |
| Available stock read model | `newRuntimeStockAvailabilityStore` | PostgreSQL-backed reads from `inventory.stock_balances` when DB config exists | PR #407, PR #408, full dev smoke `persisted_available_stock` |
| Sales orders | `newRuntimeSalesOrderStore` | PostgreSQL-backed owner documents when DB config exists | PR #410, PR #411, PR #412, full dev smoke `persisted_sales_order` |
| Purchase orders | `newRuntimePurchaseOrderStore` | PostgreSQL-backed owner documents when DB config exists | PR #414, PR #415, full dev smoke `persisted_purchase_order` and inbound trace |
| Return receipts | `newRuntimeReturnReceiptStore` | PostgreSQL-backed receipt, line, inspection, disposition, and attachment refs when DB config exists | PR #417, PR #419, full dev smoke `persisted_return_receipt` |
| Supplier rejections | `newRuntimeSupplierRejectionStore` | PostgreSQL-backed rejection header, line, attachment, and status lifecycle when DB config exists | PR #418, PR #419, full dev smoke `persisted_supplier_rejection` |

Important note:

```text
Prototype fallback still exists for no-DB/local mode in the runtime selectors above. That fallback is intentional and must not be counted as production persistence evidence.
```

---

## 3. Remaining Backend Prototype Stores

### P1 - Operational Evidence Stores

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Batch catalog / QC status | `inventoryapp.NewPrototypeBatchCatalog(auditLogStore)` | Batch QC status changes reset even when inbound QC and stock balances persist | Add PostgreSQL batch read/write adapter or tie changes to `inventory.batches` |
| End-of-day reconciliation | `inventoryapp.NewPrototypeEndOfDayReconciliationStore()` | Shift close and variance evidence reset | Persist after Sprint 11 owner documents, before relying on shift-close history |
| Carrier manifests | `shippingapp.NewPrototypeCarrierManifestStore()` | Handover and scan evidence reset | Persist manifest/header/line/scan exception state |
| Pick tasks | `shippingapp.NewPrototypePickTaskStore(...)` | Pick progress resets while sales orders and reservations now persist | Persist with shipping task package |
| Pack tasks | `shippingapp.NewPrototypePackTaskStore(...)` | Pack progress resets while sales orders and reservations now persist | Persist with shipping task package |
| Subcontract orders | `productionapp.NewPrototypeSubcontractOrderStore(auditLogStore)` | Gia cong order state resets | Persist before expanding subcontract finance/reporting |
| Subcontract material transfers | `productionapp.NewPrototypeSubcontractMaterialTransferStore()` | NVL/bao bi transfer evidence resets | Persist with subcontract order lifecycle |
| Subcontract samples / receipts / claims / payment milestones | Prototype subcontract stores | Sample/QC/claim/final payment state resets | Persist as one subcontract runtime package or in dependency order |

### P1 - Finance Runtime Stores

Finance runtime stores should be promoted together enough to avoid partial financial truth.

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Customer receivables | `financeapp.NewPrototypeCustomerReceivableStore()` | AR status and receipts reset while sales order documents now persist | Persist with sales AR flow |
| Supplier payables | `financeapp.NewPrototypeSupplierPayableStore()` | AP/payment approval state resets while PO and supplier rejection documents now persist | Persist with PO/subcontract payable flows |
| COD remittances | `financeapp.NewPrototypeCODRemittanceStore()` | COD match/discrepancy/approval state resets | Persist after receivables foundation |
| Cash transactions | `financeapp.NewPrototypeCashTransactionStore()` | Cash movement evidence resets | Persist with finance audit/reporting gate |

### P2 - Master Data, Auth, and Dev Fallbacks

| Store / service | Current constructor or path | Runtime risk | Recommendation |
| --- | --- | --- | --- |
| Item catalog | `masterdataapp.NewPrototypeItemCatalog(auditLogStore)` | MDM edits reset | Persist after operational transaction stores or when MDM editing becomes primary workflow |
| Warehouse/location catalog | `masterdataapp.NewPrototypeWarehouseLocationCatalog(auditLogStore)` | Location edits reset | Persist before more warehouse layout features |
| Party catalog | `masterdataapp.NewPrototypePartyCatalog(auditLogStore)` | Supplier/customer edits reset | Persist before supplier/customer maintenance workflows |
| UOM catalog | `masterdataapp.NewPrototypeUOMCatalog()` | UOM edits reset | Lower risk while standards are mostly static |
| Access/refresh sessions | `auth.NewSessionManager(...)` | Sessions and lockout state reset | Acceptable for current mock/dev auth; revisit before production auth |
| Frontend fallback services | `apps/web/src/modules/**/services/*` | Can hide backend failures during UI testing | Keep dev-only; never count frontend fallback as persistence evidence |

---

## 4. Recommended Post-Sprint-11 Persistence Order

```text
1. Batch/QC status persistence or adapter to inventory.batches.
2. End-of-day reconciliation.
3. Shipping manifest, pick, and pack task package.
4. Finance AR/AP/COD/cash runtime stores.
5. Subcontract runtime stores.
6. Master data catalogs and auth/session hardening.
```

Rationale:

```text
Sprint 11 closed the highest cross-store mismatch between persisted operational evidence and owner documents. The next risk is operational evidence that still resets inside warehouse execution, finance, and subcontract flows.
```

---

## 5. Verification Notes

Inventory checks performed:

```text
- Inspected current main.go runtime constructors after S11-04-04.
- Inspected runtime selector files for PostgreSQL/prototype fallback behavior.
- Compared against the Sprint 10 remaining prototype store ledger.
- Confirmed S11 PR chain from git history: #405 through #419.
- Confirmed dev deploy after PR #419 ran full dev smoke with persisted_available_stock, persisted_sales_order, persisted_purchase_order, persisted_return_receipt, and persisted_supplier_rejection.
- No runtime behavior changed in this task.
```
