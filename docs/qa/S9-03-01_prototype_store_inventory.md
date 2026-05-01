# S9-03-01 Prototype Store Inventory

Project: Web ERP for cosmetics operations
Sprint: Sprint 9 - System hardening / production readiness core
Task: S9-03-01 Prototype store inventory
Date: 2026-05-01
Status: Inventory complete; S9-03-02 persistence target selected

---

## 1. Purpose

This inventory records runtime stores that still reset on process restart or fall back to frontend-only prototype state.

The goal is to choose the next persistence target based on operational risk, not implementation convenience.

---

## 2. Backend Runtime Stores Observed In `main.go`

High-risk operational stores:

| Store / service | Current constructor | Runtime risk | Priority |
| --- | --- | --- | --- |
| Stock movement ledger/balance writer | `inventoryapp.NewInMemoryStockMovementStore()` | Stock ledger and balance changes disappear on restart; highest correctness risk | P0 |
| Audit log | `audit.NewPrototypeLogStore()` | Evidence and auth/action audit records reset on restart | P0 |
| Sales order reservations | `inventoryapp.NewPrototypeSalesOrderReservationStore(...)` | Reserved stock can drift after restart | P0 |
| Stock adjustments | `inventoryapp.NewPrototypeStockAdjustmentStore()` | Adjustment approval/posting state resets | P1 |
| Stock counts | `inventoryapp.NewPrototypeStockCountStore()` | Cycle count and reconciliation evidence resets | P1 |
| End-of-day reconciliation | `inventoryapp.NewPrototypeEndOfDayReconciliationStore()` | Shift close evidence resets | P1 |
| Warehouse receiving | `inventoryapp.NewPrototypeWarehouseReceivingStore()` | Receiving status resets before/after QC | P1 |
| Inbound QC | `qcapp.NewPrototypeInboundQCInspectionStore()` | QC pass/fail/hold decisions reset | P1 |
| Return receipts | `returnsapp.NewPrototypeReturnReceiptStore()` | Return inspection/disposition state resets | P1 |
| Supplier rejections | `inventoryapp.NewPrototypeSupplierRejectionStore()` | Return-to-supplier evidence resets | P1 |

Finance and subcontract stores:

| Store / service | Current constructor | Runtime risk | Priority |
| --- | --- | --- | --- |
| Customer receivables | `financeapp.NewPrototypeCustomerReceivableStore()` | AR status and receipts reset | P1 |
| Supplier payables | `financeapp.NewPrototypeSupplierPayableStore()` | AP/payment approval state resets | P1 |
| COD remittances | `financeapp.NewPrototypeCODRemittanceStore()` | COD match/discrepancy/approval state resets | P1 |
| Cash transactions | `financeapp.NewPrototypeCashTransactionStore()` | Cash movement evidence resets | P1 |
| Subcontract orders | `productionapp.NewPrototypeSubcontractOrderStore(...)` | Gia cong order state resets | P1 |
| Subcontract material transfers | `productionapp.NewPrototypeSubcontractMaterialTransferStore()` | NVL/bao bi transfer state resets | P1 |
| Subcontract samples / receipts / claims / payment milestones | Prototype subcontract stores | QC/sample/claim/final payment state resets | P1 |

Fulfillment and master-data stores:

| Store / service | Current constructor | Runtime risk | Priority |
| --- | --- | --- | --- |
| Sales orders | `salesapp.NewPrototypeSalesOrderStore(...)` | Order lifecycle resets | P1 |
| Carrier manifests | `shippingapp.NewPrototypeCarrierManifestStore()` | Handover/scan evidence resets | P1 |
| Pick tasks | `shippingapp.NewPrototypePickTaskStore(...)` | Pick progress resets | P1 |
| Pack tasks | `shippingapp.NewPrototypePackTaskStore(...)` | Pack progress resets | P1 |
| Product/item catalog | `masterdataapp.NewPrototypeItemCatalog(...)` | MDM edits reset | P2 |
| Warehouse/location catalog | `masterdataapp.NewPrototypeWarehouseLocationCatalog(...)` | Location edits reset | P2 |
| Party catalog | `masterdataapp.NewPrototypePartyCatalog(...)` | Supplier/customer edits reset | P2 |
| Batch catalog | `inventoryapp.NewPrototypeBatchCatalog(...)` | Batch/QC status edits reset | P1 |

Auth/session state:

| Store / service | Current constructor | Runtime risk | Priority |
| --- | --- | --- | --- |
| Access/refresh sessions | `auth.NewSessionManager(...)` | Sessions and lockout state reset on restart | P2 |

Attachment storage:

| Store / service | Current constructor | Runtime risk | Priority |
| --- | --- | --- | --- |
| Return attachment object store | `storage.NewS3CompatibleObjectStore(...)` | Object storage is already externalized when S3/MinIO config is present | P2 |

---

## 3. Frontend Prototype Fallback Stores

Frontend services still contain prototype fallback state for offline/dev behavior. These are lower persistence priority than backend stores because the API is the system of record, but they can hide backend failures during UI work.

Examples:

```text
apps/web/src/modules/finance/services/*Service.ts
apps/web/src/modules/inventory/services/*Service.ts
apps/web/src/modules/shipping/services/*Service.ts
apps/web/src/modules/subcontract/services/subcontractOrderService.ts
apps/web/src/modules/warehouse/services/warehouseDailyBoardService.ts
apps/web/src/modules/reporting/services/*ReportService.ts
```

Guardrail:

```text
Frontend fallback should remain dev-only behavior and must not be counted as persistence evidence.
```

---

## 4. Existing Persistence Hook Worth Using

The codebase already has a PostgreSQL stock movement store:

```text
apps/api/internal/modules/inventory/application/postgres_stock_movement_store.go
```

It writes:

```text
- inventory.stock_ledger
- inventory.stock_balances
- audit.audit_logs for inventory.stock_movement
```

It also uses the stock balance write guard:

```text
SET LOCAL erp.allow_stock_balance_write = 'on'
```

This makes stock movement the best S9-03-02 persistence target because it is high-risk and already has a domain-specific persistent implementation.

---

## 5. Selected S9-03-02 Target

Selected target:

```text
Persist stock movement recording by wiring runtime stockMovementStore to PostgreSQL where a DB connection is configured.
```

Acceptance for S9-03-02:

```text
1. Dev/local can still run without PostgreSQL if the current deploy path requires it.
2. When PostgreSQL runtime config is present, stock movements use PostgresStockMovementStore.
3. Stock movement tests cover local fallback and PostgreSQL store selection.
4. Existing stock movement guard tests remain green.
5. No direct stock balance writes are introduced.
```

Out of scope for S9-03-02:

```text
- Persisting every inventory document store.
- Replacing all prototype master-data catalogs.
- Rewriting frontend fallback stores.
- Adding new stock movement schema.
```

---

## 6. Follow-Up Persistence Order

Recommended order after S9-03-02:

```text
1. Audit log store.
2. Sales order reservation store.
3. Stock count and stock adjustment stores.
4. Receiving + inbound QC stores.
5. Returns + supplier rejection stores.
6. Finance AR/AP/COD/cash stores.
7. Shipping pick/pack/manifest stores.
8. Subcontract order/material/sample/receipt/claim/payment stores.
9. Master data catalogs.
```

---

## 7. Verification Notes

Inventory checks performed:

```text
- Inspected backend main.go runtime store wiring.
- Searched backend application packages for NewPrototype*Store and NewInMemory*Store.
- Confirmed PostgresStockMovementStore already exists.
- Confirmed migration files contain inventory.stock_ledger and inventory.stock_balances.
- Checked frontend services for prototype fallback state.
```

No runtime behavior changed in this task.
