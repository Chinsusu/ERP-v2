# S11-02-01 Sales Order Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Task: S11-02-01 Sales order document persistence design
Date: 2026-05-01
Status: Design complete; S11-02-02 can implement migration and PostgreSQL store

---

## 1. Goal

Persist sales order owner documents so they no longer reset while related reservation rows survive restart.

Current risk after Sprint 10:

```text
sales order reservation rows persist in inventory.stock_reservations
sales order owner documents still live in PrototypeSalesOrderStore
```

That can create a mismatch after API restart:

```text
reservation exists
-> sales order document disappears or falls back to seed data
-> fulfillment, reporting, and audit traceability weaken
```

---

## 2. Existing Application Contract

Keep the existing store boundary:

```text
SalesOrderStore.List(ctx, filter)
SalesOrderStore.Get(ctx, id)
SalesOrderStore.WithinTx(ctx, fn)

SalesOrderTx.GetForUpdate(ctx, id)
SalesOrderTx.Save(ctx, order)
SalesOrderTx.RecordAudit(ctx, log)
```

S11-02-02 should add:

```text
PostgresSalesOrderStore
newRuntimeSalesOrderStore(cfg, auditLogStore)
```

No OpenAPI envelope change is needed.

---

## 3. Schema Gap

The existing base tables are close but not enough for the current domain object.

Existing useful columns:

```text
sales.sales_orders:
id, org_id, order_no, customer_id, order_date, channel, status,
currency_code, subtotal_amount, discount_amount, tax_amount,
shipping_fee_amount, net_amount, total_amount, created_at, created_by,
updated_at, updated_by, cancelled_at, cancelled_by, cancel_reason, version

sales.sales_order_lines:
id, org_id, sales_order_id, line_no, item_id, ordered_qty, reserved_qty,
shipped_qty, unit_price, uom_code, base_ordered_qty, base_uom_code,
conversion_factor, currency_code, line_discount_amount, line_amount
```

Missing for the runtime domain:

```text
sales order header:
order_ref, org_ref, customer_ref, customer_code, customer_name,
warehouse_id, warehouse_ref, warehouse_code, note,
created_by_ref, updated_by_ref,
confirmed_at/by/ref, reserved_at/by/ref,
picking_started_at/by/ref, picked_at/by/ref,
packing_started_at/by/ref, packed_at/by/ref,
waiting_handover_at/by/ref, handed_over_at/by/ref,
closed_at/by/ref, exception_at/by/ref

sales order line:
line_ref, item_ref, sku_code, item_name,
batch_id, batch_ref, batch_no
```

The existing line table also requires `item_id` and `unit_id`, but the runtime domain can carry stable text
references such as `item-serum-30ml` and `pcs` instead of UUIDs. S11-02-02 should relax those columns and
persist text refs explicitly instead of forcing runtime state into unrelated UUID columns.

---

## 4. Migration Shape

Recommended migration:

```text
000020_persist_sales_orders
```

Header additions:

```text
order_ref text
org_ref text
customer_ref text
customer_code text
customer_name text
warehouse_id uuid references mdm.warehouses(id)
warehouse_ref text
warehouse_code text
note text
created_by_ref text
updated_by_ref text
confirmed_at timestamptz
confirmed_by uuid references core.users(id)
confirmed_by_ref text
reserved_at timestamptz
reserved_by uuid references core.users(id)
reserved_by_ref text
picking_started_at timestamptz
picking_started_by uuid references core.users(id)
picking_started_by_ref text
picked_at timestamptz
picked_by uuid references core.users(id)
picked_by_ref text
packing_started_at timestamptz
packing_started_by uuid references core.users(id)
packing_started_by_ref text
packed_at timestamptz
packed_by uuid references core.users(id)
packed_by_ref text
waiting_handover_at timestamptz
waiting_handover_by uuid references core.users(id)
waiting_handover_by_ref text
handed_over_at timestamptz
handed_over_by uuid references core.users(id)
handed_over_by_ref text
closed_at timestamptz
closed_by uuid references core.users(id)
closed_by_ref text
exception_at timestamptz
exception_by uuid references core.users(id)
exception_by_ref text
```

Line additions:

```text
line_ref text
item_ref text
sku_code text
item_name text
batch_id uuid references inventory.batches(id)
batch_ref text
batch_no text
```

Constraints/indexes:

```text
alter sales.sales_order_lines item_id drop not null
alter sales.sales_order_lines unit_id drop not null
check item_id is not null or item_ref is not null
check unit_id is not null or uom_code is not null
unique nulls not distinct (org_id, order_ref)
index on (org_id, status, order_date desc)
index on (org_id, customer_ref, order_date desc)
index on (org_id, warehouse_ref, status)
unique nulls not distinct (sales_order_id, line_ref)
```

Existing seed rows should backfill refs from UUIDs and joined master data where possible.

---

## 5. Store Behavior

`PostgresSalesOrderStore.List`:

```text
select headers
apply current SalesOrderFilter semantics:
  search in order_no/customer_code/customer_name/channel
  statuses
  customer_id/customer_ref exact
  channel case-insensitive exact
  warehouse_id/warehouse_ref exact
  date_from/date_to
load lines for returned headers
sort by order_date desc, order_no asc
```

`PostgresSalesOrderStore.Get`:

```text
find by order_ref or id::text
load lines
return ErrSalesOrderNotFound when absent
```

`PostgresSalesOrderStore.WithinTx`:

```text
open serializable transaction
provide SalesOrderTx backed by the same transaction
commit only if Save and RecordAudit succeed
rollback on any error
```

`GetForUpdate`:

```text
select header by order_ref or id::text FOR UPDATE
load lines
```

`Save`:

```text
upsert sales.sales_orders by (org_id, order_ref)
delete and reinsert lines for the order
store runtime refs and UUID columns when values are UUID-compatible
preserve domain Version
```

`RecordAudit`:

```text
insert audit.audit_logs in the same transaction
use existing audit action/entity/request/before/after/metadata values
store entity_ref for text order IDs
```

---

## 6. Runtime Selection

Selection rule:

```text
DATABASE_URL present -> PostgresSalesOrderStore
DATABASE_URL empty   -> PrototypeSalesOrderStore
```

Keep prototype fallback explicit for local/no-DB mode.

The runtime service should still compose with the existing PostgreSQL reservation store when DB config exists:

```text
PostgresSalesOrderStore + PostgresSalesOrderReservationStore
```

---

## 7. Tests

S11-02-02 should include:

```text
1. Migration apply/rollback through required migration CI.
2. PostgreSQL store selector tests.
3. Store unit tests for SQL mapping where practical.
4. Env-gated PostgreSQL integration test for create -> confirm/reserve -> cancel/release -> reload.
5. Existing sales service tests remain green.
```

S11-02-03 should add dev smoke:

```text
create sales order
confirm to reserve stock
cancel to release stock
assert sales.sales_orders row remains queryable
assert sales.sales_order_lines row remains queryable
assert inventory.stock_reservations rows remain linked by sales_order_ref and sales_order_line_ref
assert audit rows exist
```

---

## 8. S11-02-02 Acceptance

S11-02-02 is complete when:

```text
1. Sales order runtime store uses PostgreSQL when DATABASE_URL exists.
2. Create/update/confirm/reserve/cancel paths persist owner documents.
3. Confirm/reserve and cancel/release stay transactional at the service level.
4. Existing API envelopes remain unchanged.
5. Migration up/down passes on PostgreSQL 16.
6. Backend tests and vet pass.
7. No reservation, stock balance, auth, or permission behavior is weakened.
```
