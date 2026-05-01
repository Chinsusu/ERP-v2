# S10-02-01 Reservation Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 10 - Persist operational runtime stores v1
Task: S10-02-01 Reservation persistence design
Date: 2026-05-01
Status: Design complete; ready for S10-02-02 implementation

---

## 1. Purpose

This document maps the current sales order stock reservation runtime to a PostgreSQL-backed store.

The goal is to persist reservation lifecycle state without changing sales order action semantics, stock allocation rules, audit behavior, or API response envelopes.

---

## 2. Current Runtime Behavior

Current wiring:

```text
apps/api/cmd/api/main.go
salesOrderReservationStore := inventoryapp.NewPrototypeSalesOrderReservationStore(auditLogStore)
salesOrderService := salesapp.NewSalesOrderService(...).WithStockReserver(salesOrderReservationStore)
```

Current interface owned by the sales module:

```text
apps/api/internal/modules/sales/application/sales_order_service.go

type SalesOrderStockReserver interface {
  ReserveSalesOrder(ctx context.Context, input SalesOrderStockReservationInput) (SalesOrderStockReservationResult, error)
  ReleaseSalesOrder(ctx context.Context, input SalesOrderStockReleaseInput) (SalesOrderStockReleaseResult, error)
}
```

Current implementation:

```text
apps/api/internal/modules/inventory/application/reserve_sales_order_stock.go
PrototypeSalesOrderReservationStore
```

The prototype store owns two in-memory datasets:

```text
rows         []domain.StockBalanceSnapshot
reservations []domain.StockReservation
```

Reservation flow:

```text
1. Clone in-memory stock rows.
2. Calculate available stock from physical, reserved, QC hold, blocked, and batch status.
3. Allocate the requested base quantity from one sellable stock row.
4. Increase the cloned row's QtyReserved.
5. Create domain.StockReservation records.
6. Record audit logs.
7. Swap the store's in-memory rows and append reservations.
```

Release flow:

```text
1. Find active reservations by sales order id.
2. Mark each active reservation released.
3. Subtract reserved quantity from the in-memory row.
4. Record audit logs.
5. Swap the in-memory rows and reservations.
```

Runtime risk:

```text
Reservations and reserved quantity effects disappear on API restart or redeploy.
```

---

## 3. Existing Domain Contract

Current reservation domain:

```text
apps/api/internal/modules/inventory/domain/stock_reservation.go

StockReservation:
  ID
  OrgID
  ReservationNo
  SalesOrderID
  SalesOrderLineID
  ItemID
  SKUCode
  BatchID
  BatchNo
  WarehouseID
  WarehouseCode
  BinID
  BinCode
  StockStatus
  ReservedQty
  BaseUOMCode
  Status
  ReservedAt / ReservedBy
  ReleasedAt / ReleasedBy
  ConsumedAt / ConsumedBy
  CreatedAt / UpdatedAt
```

Supported statuses:

```text
active
released
consumed
```

Guardrails already enforced:

```text
1. Reservation requires actor id.
2. Reservation quantity must be positive decimal quantity.
3. Batch and bin are required.
4. Only active reservation can be released or consumed.
5. Released/consumed records must carry actor and timestamp metadata.
```

---

## 4. Existing PostgreSQL Schema

Current canonical table:

```text
inventory.stock_reservations
```

After migration `000006_harden_stock_reservations`, the table carries:

```text
id uuid PRIMARY KEY DEFAULT gen_random_uuid()
org_id uuid NOT NULL REFERENCES core.organizations(id)
reservation_no text NOT NULL
item_id uuid NOT NULL REFERENCES mdm.items(id)
batch_id uuid REFERENCES inventory.batches(id)
warehouse_id uuid NOT NULL REFERENCES mdm.warehouses(id)
reserved_qty numeric(18,6) NOT NULL
source_doc_type text NOT NULL DEFAULT 'sales_order'
source_doc_id uuid NOT NULL
sales_order_id uuid NOT NULL REFERENCES sales.sales_orders(id)
sales_order_line_id uuid NOT NULL REFERENCES sales.sales_order_lines(id)
source_doc_line_id uuid NOT NULL REFERENCES sales.sales_order_lines(id)
bin_id uuid NOT NULL REFERENCES mdm.warehouse_bins(id)
base_uom_code varchar(20) NOT NULL REFERENCES mdm.uoms(uom_code)
stock_status text NOT NULL DEFAULT 'available'
status text NOT NULL
created_at timestamptz NOT NULL DEFAULT now()
created_by uuid REFERENCES core.users(id)
released_at timestamptz
released_by uuid REFERENCES core.users(id)
consumed_at timestamptz
consumed_by uuid REFERENCES core.users(id)
updated_at timestamptz NOT NULL DEFAULT now()
```

Indexes already available:

```text
ix_stock_reservations_active
ix_stock_reservations_sales_order_line_active
ix_stock_reservations_stock_key_active
```

---

## 5. Compatibility Gap

The current API runtime still uses many text references:

```text
org-my-pham
so-reserve
line-serum
item-serum-30ml
batch-serum-2604a
wh-hcm-fg
bin-hcm-pick-a01
user-sales
```

The existing PostgreSQL table requires UUID foreign keys for sales order, line, item, warehouse, and bin.

Do not solve this by forcing text runtime ids into UUID columns. That would fail writes and hide the current prototype boundary.

Do not bypass reservation persistence by routing all text ids to memory when `DATABASE_URL` exists. That would keep the current restart data loss and fail the Sprint 10 goal.

---

## 6. Design Decision

Use `inventory.stock_reservations` as the persisted reservation table and add explicit runtime compatibility reference columns.

The PostgreSQL store should preserve current `SalesOrderStockReserver` behavior:

```text
ReserveSalesOrder
ReleaseSalesOrder
```

The store should keep the current domain model unchanged.

For S10-02-02, keep the prototype stock-balance baseline as the physical stock source for text-runtime sales flows, but load active persisted reservations from PostgreSQL and apply them to the baseline before allocation. This gives restart-safe reservation effects without pretending the full sales/master-data stack is already UUID-backed.

When the input is fully UUID-compatible and matching master/sales rows exist, the same table can carry UUID FK values. When runtime ids are text-only, the UUID FK columns may be null and the `*_ref` columns carry the durable business reference.

---

## 7. Migration Plan For S10-02-02

Add compatibility columns:

```text
reservation_ref text
org_ref text
sales_order_ref text
sales_order_line_ref text
source_doc_ref text
source_doc_line_ref text
item_ref text
sku_code text
batch_ref text
batch_no text
warehouse_ref text
warehouse_code text
bin_ref text
bin_code text
created_by_ref text
released_by_ref text
consumed_by_ref text
```

Backfill refs for existing rows from UUID columns:

```text
reservation_ref = COALESCE(reservation_ref, id::text)
org_ref = COALESCE(org_ref, org_id::text)
sales_order_ref = COALESCE(sales_order_ref, sales_order_id::text)
sales_order_line_ref = COALESCE(sales_order_line_ref, sales_order_line_id::text)
source_doc_ref = COALESCE(source_doc_ref, source_doc_id::text)
source_doc_line_ref = COALESCE(source_doc_line_ref, source_doc_line_id::text)
item_ref = COALESCE(item_ref, item_id::text)
batch_ref = COALESCE(batch_ref, batch_id::text)
warehouse_ref = COALESCE(warehouse_ref, warehouse_id::text)
bin_ref = COALESCE(bin_ref, bin_id::text)
created_by_ref = COALESCE(created_by_ref, created_by::text)
released_by_ref = COALESCE(released_by_ref, released_by::text)
consumed_by_ref = COALESCE(consumed_by_ref, consumed_by::text)
```

Relax only the non-tenant UUID columns needed for text runtime compatibility:

```text
sales_order_id
sales_order_line_id
source_doc_id
source_doc_line_id
item_id
batch_id
warehouse_id
bin_id
created_by
released_by
consumed_by
```

Keep `org_id` required. For local/dev, resolve unknown org refs to the seeded dev org id while preserving `org_ref`, matching the audit persistence pattern.

Replace strict source constraint with a compatibility-safe constraint:

```text
source_doc_type = 'sales_order'
AND (
  source_doc_id = sales_order_id
  OR source_doc_ref = sales_order_ref
)
AND (
  source_doc_line_id = sales_order_line_id
  OR source_doc_line_ref = sales_order_line_ref
)
```

Add indexes:

```text
CREATE UNIQUE INDEX uq_stock_reservations_reservation_ref
  ON inventory.stock_reservations(org_id, reservation_ref)
  WHERE reservation_ref IS NOT NULL;

CREATE INDEX ix_stock_reservations_sales_order_ref_active
  ON inventory.stock_reservations(org_id, sales_order_ref)
  WHERE status = 'active';

CREATE INDEX ix_stock_reservations_line_ref_active
  ON inventory.stock_reservations(org_id, sales_order_line_ref)
  WHERE status = 'active';

CREATE INDEX ix_stock_reservations_stock_ref_active
  ON inventory.stock_reservations(org_id, warehouse_ref, item_ref, batch_ref, bin_ref, stock_status)
  WHERE status = 'active';
```

Down migration should drop only the new indexes and compatibility columns, then restore the previous strict UUID constraints. If existing text-only rows are present, the down migration should fail rather than silently deleting reservation evidence.

---

## 8. Store Plan For S10-02-02

Add a PostgreSQL implementation in inventory application:

```text
PostgresSalesOrderReservationStore
```

Keep the existing interface:

```text
ReserveSalesOrder(ctx, input)
ReleaseSalesOrder(ctx, input)
```

Add runtime selection:

```text
if DATABASE_URL is empty:
  NewPrototypeSalesOrderReservationStore(auditLogStore)
else:
  NewPostgresSalesOrderReservationStore(db, auditLogStore, prototype baseline rows, config)
```

Reservation algorithm:

```text
1. Start serializable transaction.
2. Load active persisted reservations for the relevant stock refs.
3. Apply active persisted reservation qty to cloned baseline stock rows.
4. Reuse the current allocation rules against the adjusted rows.
5. Insert one inventory.stock_reservations row per allocated reservation.
6. Record audit log for each inserted reservation.
7. Return the same SalesOrderReservedLine payload shape.
```

Release algorithm:

```text
1. Start serializable transaction.
2. Select active reservations by sales_order_ref or sales_order_id.
3. Rehydrate domain.StockReservation records.
4. Transition each record to released with existing domain logic.
5. Update status, released_at, released_by/ref, and updated_at.
6. Record audit log for each released reservation.
7. Return ReleasedReservationCount.
```

The PostgreSQL store must not update `inventory.stock_balances` directly. Reservation availability is represented by persisted reservation rows and the reservation store's allocation view until a broader canonical stock-balance reservation service is introduced.

---

## 9. Required Tests For S10-02-02

Backend unit tests:

```text
PostgresSalesOrderReservationStore reserves against baseline minus active persisted reservations.
PostgresSalesOrderReservationStore inserts reservation_ref and runtime refs for text ids.
PostgresSalesOrderReservationStore preserves UUID columns when ids are UUID-compatible.
PostgresSalesOrderReservationStore does not partially reserve on insufficient stock.
PostgresSalesOrderReservationStore releases active reservations by sales_order_ref.
PostgresSalesOrderReservationStore records reserve and release audit logs.
Runtime selector uses prototype store without DATABASE_URL.
Runtime selector uses PostgreSQL store with DATABASE_URL.
```

Migration checks:

```text
Apply all migrations on PostgreSQL 16.
Roll back migration 000015 on PostgreSQL 16.
Re-apply migration 000015 on PostgreSQL 16.
```

Regression checks:

```text
go test ./internal/modules/inventory/application ./internal/modules/inventory/domain ./internal/modules/sales/application -count=1
go test ./cmd/api -run 'Test.*SalesOrder|Test.*Reservation' -count=1
```

---

## 10. Smoke Plan For S10-02-03

Add a focused smoke path after S10-02-02:

```text
1. Create or use a deterministic sales order reservation request.
2. Record active reservation count for the sales order ref in PostgreSQL.
3. Confirm/reserve through API.
4. Assert inventory.stock_reservations has active rows for that sales order ref.
5. Restart or redeploy API.
6. Cancel/release the same sales order if the sales order runtime still exists, or query PostgreSQL directly for persisted active rows if sales order state is still prototype.
7. Confirm the persisted reservation rows remain queryable after restart/redeploy.
```

If sales order document state is still prototype when S10-02-03 is implemented, the smoke should state that boundary explicitly and only claim reservation-store persistence, not full sales-order persistence.

---

## 11. Out Of Scope For S10-02-01

This design task does not:

```text
- Implement the PostgreSQL reservation store.
- Add migrations.
- Change API response shapes.
- Persist the full sales order document store.
- Persist the general available stock endpoint.
- Update stock balances directly for reservation side effects.
- Create pick tasks from reservations.
```

---

## 12. Acceptance For S10-02-01

S10-02-01 is complete when:

```text
1. Current reservation runtime behavior is documented.
2. Current DB schema and UUID/text compatibility gap are explicit.
3. S10-02-02 has a concrete migration and store implementation plan.
4. S10-02-03 has a concrete persistence smoke plan.
5. No runtime behavior changes in this task.
```
