# S14-04-01 Remaining Prototype Store Ledger

Project: Web ERP for cosmetics operations
Sprint: Sprint 14 - Shipping pick/pack persistence
Task: S14-04-01 Remaining prototype ledger update
Date: 2026-05-02
Status: Inventory complete; S14-05-01 release evidence can use this ledger

---

## 1. Purpose

This ledger supersedes `docs/qa/S13-03-01_remaining_prototype_store_ledger.md` after the Sprint 14 shipping execution persistence work.

Sprint 14 closed the highest remaining warehouse execution reset risk from the Sprint 13 ledger:

```text
carrier manifest header, shipments, scan evidence, exception evidence, and handover state
pick task header, line progress, exception evidence, assignment, and completion state
pack task header, line progress, exception evidence, assignment, and packed state
shipping audit evidence for manifest, pick, and pack lifecycle events
sales order handover/pack integration while shipping evidence persists
```

Prototype fallback still exists for no-DB/local mode. That fallback is intentional and must not be counted as production persistence evidence.

---

## 2. Stores Persisted Through Sprint 14

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

---

## 3. Sprint 14 Resolution Notes

### Warehouse Execution Stores

| Store / service | Prior Sprint 13 risk | Sprint 14 result |
| --- | --- | --- |
| Carrier manifests | Handover and scan evidence reset | Persisted with PostgreSQL runtime selector, migration, store tests, and dev smoke |
| Pick tasks | Pick progress reset while sales orders and reservations persisted | Persisted with PostgreSQL runtime selector, migration, store tests, and dev smoke |
| Pack tasks | Pack progress reset while sales orders and reservations persisted | Persisted with PostgreSQL runtime selector, migration, store tests, dev smoke, and sales order pack integration hardening |

### Sales Order Integration

S14-03-01 changed the sales order PostgreSQL store to preserve existing sales order line row IDs when saving order state from the pack adapter. This keeps shipping pick/pack evidence references valid while still deleting intentionally removed stale lines.

---

## 4. Remaining Backend Prototype Stores

### P1 - Finance Runtime Stores

Finance runtime stores should be promoted together enough to avoid partial financial truth.

| Store / service | Current constructor | Runtime risk | Recommended handling |
| --- | --- | --- | --- |
| Customer receivables | `financeapp.NewPrototypeCustomerReceivableStore()` | AR status and receipts reset while sales order documents persist | Persist with sales AR flow |
| Supplier payables | `financeapp.NewPrototypeSupplierPayableStore()` | AP/payment approval state resets while PO, supplier rejection, and subcontract evidence can persist | Persist with PO/subcontract payable flows |
| COD remittances | `financeapp.NewPrototypeCODRemittanceStore()` | COD match/discrepancy/approval state resets | Persist after receivables foundation |
| Cash transactions | `financeapp.NewPrototypeCashTransactionStore()` | Cash movement evidence resets | Persist with finance audit/reporting gate |

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

## 5. Recommended Post-Sprint-14 Persistence Order

```text
1. Finance AR/AP/COD/cash runtime stores.
2. Subcontract runtime stores.
3. Master data catalogs.
4. Auth/session hardening.
```

Rationale:

```text
Sprint 14 closed the shipping execution reset risk between sales reservation and carrier handover. The next restart risks are finance state, subcontract lifecycle state, editable catalogs, and production-grade auth/session state.
```

---

## 6. Verification Notes

Inventory checks performed:

```text
- Inspected current main.go runtime constructors after PR #449.
- Confirmed shipping DB mode uses newRuntimeCarrierManifestStore, newRuntimePickTaskStore, and newRuntimePackTaskStore.
- Confirmed those runtime selectors use PostgreSQL stores when DATABASE_URL exists and prototype fallback when DATABASE_URL is empty.
- Confirmed Sprint 14 migrations 000027, 000028, 000029, and 000030 cover manifest, pick, pack, and actor-reference hardening.
- Confirmed full dev smoke records persisted carrier manifest, pick task, pack task, and sales order pack integration evidence.
- No runtime behavior changed in this task.
```
