# S11-01-01 Available Stock Read Model Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Task: S11-01-01 Available-stock read model design
Date: 2026-05-01
Status: Design complete; S11-01-02 can implement the PostgreSQL store

---

## 1. Goal

Promote the runtime available-stock read model from prototype snapshots to PostgreSQL-backed reads when database configuration exists.

Current risk:

```text
stock movement writes -> inventory.stock_ledger and inventory.stock_balances persist
available-stock API/reporting -> still reads PrototypeStockAvailabilityStore seed snapshots
```

That creates a correctness mismatch after restart/redeploy: persisted stock evidence exists, but the user-facing stock availability view can still show seed-only data.

---

## 2. Existing Contract

The existing application boundary is already suitable:

```text
StockAvailabilityStore.ListBalances(ctx, filter) []domain.StockBalanceSnapshot
ListAvailableStock.Execute(ctx, filter) []domain.AvailableStockSnapshot
```

Keep this contract unchanged.

S11-01-02 should add a store implementation only:

```text
PostgresStockAvailabilityStore
```

No OpenAPI response shape change is needed.

---

## 3. PostgreSQL Source Tables

Primary source:

```text
inventory.stock_balances
```

Read joins:

```text
mdm.items             -> sku
mdm.warehouses        -> warehouse code
mdm.warehouse_bins    -> location/bin code
inventory.batches     -> batch no, qc status, expiry, batch status
```

Read columns mapped to `domain.StockBalanceSnapshot`:

| Domain field | PostgreSQL source |
| --- | --- |
| `WarehouseID` | `stock_balances.warehouse_id::text` |
| `WarehouseCode` | `warehouses.code` |
| `LocationID` | `stock_balances.bin_id::text`, empty when null |
| `LocationCode` | `warehouse_bins.code`, empty when null |
| `ItemID` | `stock_balances.item_id::text` |
| `SKU` | `items.sku` |
| `BatchID` | `stock_balances.batch_id::text`, empty when null |
| `BatchNo` | `batches.batch_no`, empty when null |
| `BatchQCStatus` | `batches.qc_status`, empty when no batch |
| `BatchStatus` | `batches.status`, empty when no batch |
| `BatchExpiry` | `batches.expiry_date`, zero time when null |
| `BaseUOMCode` | `stock_balances.base_uom_code` |
| `StockStatus` | `stock_balances.stock_status` |
| `QtyOnHand` | `stock_balances.qty_on_hand` |
| `QtyReserved` | `stock_balances.qty_reserved` |

Do not read `stock_balances.qty_available` into the domain. The existing domain calculation derives available quantity from on-hand, reserved, QC hold, blocked, return-pending, and batch availability status. This preserves current API semantics and avoids trusting a database column that may not express all domain grouping rules.

---

## 4. Filters

Preserve current filter behavior:

```text
warehouse_id exact match
location_id/bin_id exact match
item_id exact match
sku case-insensitive exact match
batch_id exact match
```

The current HTTP handler only builds warehouse/location/SKU/batch filters. The store should still support `ItemID` because the domain filter already includes it and tests can cover it.

Org scoping is not added in this task because the current `StockAvailabilityStore` contract does not carry org context. Adding org scope belongs with auth/session org propagation, not this read-store promotion.

---

## 5. Runtime Selection

Add runtime selector:

```text
newRuntimeStockAvailabilityStore(cfg)
```

Selection rule:

```text
DATABASE_URL present -> PostgresStockAvailabilityStore
DATABASE_URL empty   -> PrototypeStockAvailabilityStore
```

This matches the Sprint 10 persistence pattern and keeps no-DB/local behavior working.

---

## 6. Tests

S11-01-02 should include:

```text
1. SQL/query mapping test for PostgresStockAvailabilityStore.
2. Filter coverage for warehouse, location/bin, item_id, SKU, and batch_id.
3. Decimal parsing coverage for numeric(18,6) values.
4. Runtime selector test proving DB config selects PostgreSQL and empty DB config selects prototype.
```

If a PostgreSQL integration test is practical, use an isolated PostgreSQL 16 database and seed only the minimum rows required. Otherwise, a query-runner unit test is acceptable for S11-01-02, with S11-01-03 covering dev smoke against the real dev database.

---

## 7. S11-01-02 Acceptance

S11-01-02 is complete when:

```text
1. Runtime available-stock service uses PostgreSQL when DATABASE_URL exists.
2. No migration is added unless a missing index or schema gap is proven.
3. Prototype fallback remains explicit for no-DB/local mode.
4. Existing available-stock API/reporting envelopes remain unchanged.
5. Backend tests and vet pass.
6. No direct stock balance writes are introduced.
```
