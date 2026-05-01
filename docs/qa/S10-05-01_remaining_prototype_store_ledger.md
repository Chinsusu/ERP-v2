# S10-05-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 10 - Persist operational runtime stores v1
Task: S10-05-01 Remaining prototype store ledger
Date: 2026-05-01
Status: Inventory complete; S10-06-01 release evidence can use this ledger

---

## 1. Purpose

This ledger updates `docs/qa/S9-03-01_prototype_store_inventory.md` after the Sprint 10 persistence work.

Sprint 10 has reduced the highest restart-risk stores, but the application still contains prototype stores that reset on API restart. The next persistence work should be selected by operational correctness risk, not by implementation convenience.

---

## 2. Stores Persisted Before This Ledger

| Area | Runtime path | Persistence status | Evidence |
| --- | --- | --- | --- |
| Stock movement writer | `newRuntimeStockMovementStore` | PostgreSQL when DB config exists; memory fallback for non-DB/local | Dev smoke checks `inventory.stock_ledger` rows |
| Audit log | `newRuntimeAuditLogStore` | PostgreSQL when DB config exists; prototype fallback for no DB | Dev smoke checks login audit count increases |
| Sales order reservations | `newRuntimeSalesOrderReservationStore` | PostgreSQL-backed reservation rows | Dev smoke checks reserve/release rows and audit |
| Stock adjustments | `newRuntimeStockAdjustmentStore` | PostgreSQL-backed document lifecycle | Dev smoke checks posted adjustment document and stock movement |
| Stock counts | `newRuntimeStockCountStore` | PostgreSQL-backed count session lifecycle | Dev smoke checks variance-review document and audit |
| Warehouse receiving | `newRuntimeWarehouseReceivingStore` | PostgreSQL-backed receiving document lifecycle | PostgreSQL integration test and dev deploy migration evidence |
| Inbound QC | `newRuntimeInboundQCInspectionStore` | PostgreSQL-backed inspection/checklist lifecycle | Dev smoke checks QC partial decision, audit, and stock ledger available/qc_hold rows |

Important note:

```text
Stock movement persistence is now broader for UUID core refs, but optional runtime refs such as text batch/bin/source-line refs may still be persisted with nullable UUID columns. This preserves ledger/balance writes without pretending every prototype ref is master-data-backed.
```

---

## 3. Remaining Backend Prototype Stores

### P0 - Next Persistence Targets

These can create correctness drift because related writes are already persisted, but their owning documents/read models can still reset.

| Store / service | Current constructor | Risk after Sprint 10 | Recommended next task |
| --- | --- | --- | --- |
| Available stock read model | `inventoryapp.NewPrototypeStockAvailabilityStore()` | Stock movement writes are persisted, but the available-stock query path can still read prototype snapshots instead of PostgreSQL stock balances | Add PostgreSQL available-stock read store backed by `inventory.stock_balances` |
| Sales orders | `salesapp.NewPrototypeSalesOrderStore(auditLogStore)` | Reservations can survive restart while sales order documents reset, creating order/reservation mismatch | Persist sales order header/line/status lifecycle |
| Purchase orders | `purchaseapp.NewPrototypePurchaseOrderStore(auditLogStore)` | Receiving can persist while PO lifecycle resets, weakening inbound traceability | Persist PO header/line/status lifecycle |
| Returns | `returnsapp.NewPrototypeReturnReceiptStore()` | Return receipt inspection/disposition can reset while stock movement/audit evidence survives | Persist return receipt, line, inspection, disposition, attachment metadata refs |
| Supplier rejections | `inventoryapp.NewPrototypeSupplierRejectionStore()` | Failed inbound QC can persist while return-to-supplier document resets | Persist supplier rejection header/line/status lifecycle |

### P1 - Operational Evidence Stores

These are important operational records, but they do not currently create as much cross-store mismatch as the P0 group.

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| End-of-day reconciliation | `inventoryapp.NewPrototypeEndOfDayReconciliationStore()` | Shift close and variance evidence reset | Persist after P0 inventory/order documents |
| Batch catalog / QC status | `inventoryapp.NewPrototypeBatchCatalog(auditLogStore)` | Batch QC status changes reset even when inbound QC decision persists | Add PostgreSQL batch read/write adapter or tie to existing `inventory.batches` |
| Carrier manifests | `shippingapp.NewPrototypeCarrierManifestStore()` | Handover and scan evidence reset | Persist manifest/header/line/scan exception state |
| Pick tasks | `shippingapp.NewPrototypePickTaskStore(...)` | Pick progress resets | Persist after order and reservation state are stable |
| Pack tasks | `shippingapp.NewPrototypePackTaskStore(...)` | Pack progress resets | Persist with pick/manifest work |
| Subcontract orders | `productionapp.NewPrototypeSubcontractOrderStore(auditLogStore)` | Gia cong order state resets | Persist before expanding subcontract finance/reporting |
| Subcontract material transfers | `productionapp.NewPrototypeSubcontractMaterialTransferStore()` | NVL/bao bi transfer evidence resets | Persist with subcontract order lifecycle |
| Subcontract samples / receipts / claims / payment milestones | Prototype subcontract stores | Sample/QC/claim/final payment state resets | Persist as one subcontract runtime package or in dependency order |

### P1 - Finance Runtime Stores

Finance runtime stores should be promoted together enough to avoid partial financial truth.

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Customer receivables | `financeapp.NewPrototypeCustomerReceivableStore()` | AR status and receipts reset | Persist with sales order store or immediately after |
| Supplier payables | `financeapp.NewPrototypeSupplierPayableStore()` | AP/payment approval state resets | Persist with PO/subcontract payable flows |
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

## 4. Recommended Post-Sprint-10 Persistence Order

```text
1. Available stock read model from PostgreSQL stock_balances.
2. Sales order document store.
3. Purchase order document store.
4. Return receipt + supplier rejection stores.
5. Batch/QC status persistence or adapter to inventory.batches.
6. Finance AR/AP/COD/cash stores.
7. End-of-day reconciliation.
8. Shipping manifest/pick/pack stores.
9. Subcontract runtime stores.
10. Master data catalogs and auth/session hardening.
```

Rationale:

```text
The next risk is cross-store mismatch: persisted evidence without persisted owning documents or read models. Available stock, sales orders, purchase orders, returns, and supplier rejections should move before lower-risk UI/catalog persistence.
```

---

## 5. Verification Notes

Inventory checks performed:

```text
- Inspected current main.go runtime constructors after S10-04-03.
- Compared against S9-03-01 prototype store inventory.
- Re-ranked remaining stores based on persisted stores now present in main.
- No runtime behavior changed in this task.
```
