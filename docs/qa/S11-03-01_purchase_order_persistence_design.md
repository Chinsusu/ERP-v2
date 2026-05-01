# S11-03-01 Purchase Order Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Task: S11-03-01 Purchase order document persistence design
Date: 2026-05-01
Status: Design complete; S11-03-02 can implement migration and PostgreSQL store

---

## 1. Goal

Persist purchase order owner documents so receiving and inbound QC evidence remains traceable after API restart.

Current risk after Sprint 10 and S11-02:

```text
warehouse receiving and inbound QC rows persist in PostgreSQL
purchase order owner documents still live in PrototypePurchaseOrderStore
```

That can create a mismatch after restart:

```text
receiving/QC evidence references a PO
-> purchase order document disappears from runtime
-> inbound board, PO detail, receiving trace, and audit review weaken
```

---

## 2. Existing Application Contract

Keep the current store boundary:

```text
PurchaseOrderStore.List(ctx, filter)
PurchaseOrderStore.Get(ctx, id)
PurchaseOrderStore.WithinTx(ctx, fn)

PurchaseOrderTx.GetForUpdate(ctx, id)
PurchaseOrderTx.Save(ctx, order)
PurchaseOrderTx.RecordAudit(ctx, log)
```

S11-03-02 should add:

```text
PostgresPurchaseOrderStore
newRuntimePurchaseOrderStore(cfg, auditLogStore)
```

No OpenAPI envelope change is needed.

---

## 3. Existing Persistence Links

Sprint 4 and Sprint 10 already persist downstream inbound evidence:

```text
inventory.warehouse_receivings.reference_doc_ref
inventory.warehouse_receiving_lines.purchase_order_line_ref
qc.inbound_qc_inspections.purchase_order_ref
qc.inbound_qc_inspections.purchase_order_line_ref
```

The purchase order store must preserve the same runtime PO IDs and line IDs as:

```text
purchase_order.ID
purchase_order.Lines[].ID
```

Those IDs may be stable text refs, not UUIDs.

---

## 4. Schema Gap

The current tables are close but still assume UUID master data in places where runtime uses text refs.

Existing useful columns:

```text
purchase.purchase_orders:
id, org_id, po_no, supplier_id, order_date, expected_date, status,
currency_code, subtotal_amount, total_amount, warehouse_id,
submitted_at/by, approved_at/by, closed_at/by,
cancelled_at/by, cancel_reason, rejected_at/by, reject_reason,
created_at/by, updated_at/by, version

purchase.purchase_order_lines:
id, org_id, purchase_order_id, line_no, item_id, unit_id,
ordered_qty, received_qty, unit_price,
uom_code, base_ordered_qty, base_received_qty, base_uom_code,
conversion_factor, currency_code, line_amount, expected_date
```

Missing for the runtime domain:

```text
purchase order header:
po_ref, org_ref, supplier_ref, supplier_code, supplier_name,
warehouse_ref, warehouse_code, note,
created_by_ref, updated_by_ref,
submitted_by_ref, approved_by_ref,
partially_received_at/by/ref, received_at/by/ref,
closed_by_ref, cancelled_by_ref, rejected_by_ref

purchase order line:
line_ref, item_ref, sku_code, item_name, note
```

Existing `supplier_id`, `warehouse_id`, `item_id`, and `unit_id` columns are UUID-oriented. S11-03-02 should
relax the required UUID columns and store runtime text refs explicitly instead of coercing text IDs into UUID columns.

---

## 5. Migration Shape

Recommended migration:

```text
000022_persist_purchase_orders
```

Header additions:

```text
po_ref text
org_ref text
supplier_ref text
supplier_code text
supplier_name text
warehouse_ref text
warehouse_code text
note text
created_by_ref text
updated_by_ref text
submitted_by_ref text
approved_by_ref text
partially_received_at timestamptz
partially_received_by uuid references core.users(id)
partially_received_by_ref text
received_at timestamptz
received_by uuid references core.users(id)
received_by_ref text
closed_by_ref text
cancelled_by_ref text
rejected_by_ref text
```

Line additions:

```text
line_ref text
item_ref text
sku_code text
item_name text
note text
```

Column relaxations:

```text
alter purchase.purchase_orders supplier_id drop not null
alter purchase.purchase_orders warehouse_id drop not null
alter purchase.purchase_order_lines item_id drop not null
alter purchase.purchase_order_lines unit_id drop not null
```

Constraints/indexes:

```text
check po_ref is not blank
check org_ref is not blank
check supplier_id is not null or supplier_ref is not blank
check warehouse_id is not null or warehouse_ref is not blank
check line_ref is not blank
check item_id is not null or item_ref is not blank
check unit_id is not null or uom_code is not blank
unique (org_id, po_ref)
unique (purchase_order_id, line_ref)
index on (org_id, supplier_ref, expected_date desc)
index on (org_id, warehouse_ref, status)
```

Existing rows should backfill refs from UUID columns and joined master data where possible.

---

## 6. Store Behavior

`PostgresPurchaseOrderStore.List`:

```text
select headers
apply current PurchaseOrderFilter semantics:
  search in po_no/supplier_code/supplier_name/warehouse_code
  statuses
  supplier_id/supplier_ref exact
  warehouse_id/warehouse_ref exact
  expected_from/expected_to
load lines for returned headers
sort by expected_date desc, po_no asc
```

`PostgresPurchaseOrderStore.Get`:

```text
find by po_ref or id::text
load lines
return ErrPurchaseOrderNotFound when absent
```

`PostgresPurchaseOrderStore.WithinTx`:

```text
open serializable transaction
provide PurchaseOrderTx backed by the same transaction
commit only if Save and RecordAudit succeed
rollback on any error
```

`GetForUpdate`:

```text
select header by po_ref or id::text FOR UPDATE
load lines
```

`Save`:

```text
upsert purchase.purchase_orders by (org_id, po_ref)
delete and reinsert lines for the order
store runtime refs and UUID columns when values are UUID-compatible
preserve domain Version and lifecycle fields
```

`RecordAudit`:

```text
insert audit.audit_logs in the same transaction
use existing audit action/entity/request/before/after/metadata values
store entity_ref for text PO IDs
```

Do not write stock balances, receiving rows, or QC rows from this store.

---

## 7. Runtime Selection

Selection rule:

```text
DATABASE_URL present -> PostgresPurchaseOrderStore
DATABASE_URL empty   -> PrototypePurchaseOrderStore
```

Keep prototype fallback explicit for local/no-DB mode.

The API runtime should compose inbound services as:

```text
PostgresPurchaseOrderStore
Prototype/other receiving service dependencies unchanged
PostgresWarehouseReceivingStore
PostgresInboundQCInspectionStore
```

---

## 8. Tests

S11-03-02 should include:

```text
1. Migration apply/rollback through required migration CI.
2. Runtime store selector tests.
3. Store mapping/unit tests for header, line, query shape, and nil DB guard.
4. Existing purchase service tests remain green.
5. Backend go test ./... and go vet ./... pass.
```

S11-03-03 should add dev smoke:

```text
create purchase order
submit and approve
create receiving linked to that PO if existing smoke path allows it
assert purchase.purchase_orders row remains queryable by po_ref
assert purchase.purchase_order_lines row remains queryable by line_ref
assert receiving reference_doc_ref and purchase_order_line_ref remain linked
assert audit rows exist for PO lifecycle
```

---

## 9. S11-03-02 Acceptance

S11-03-02 is complete when:

```text
1. Purchase order runtime store uses PostgreSQL when DATABASE_URL exists.
2. Create/update/submit/approve/close/cancel paths persist owner documents.
3. Existing API envelopes remain unchanged.
4. Runtime text IDs are stored as refs without weakening UUID FKs for real UUID rows.
5. Migration up/down passes on PostgreSQL 16.
6. Backend tests and vet pass.
7. No receiving, QC, stock movement, audit, auth, or permission behavior is weakened.
```
