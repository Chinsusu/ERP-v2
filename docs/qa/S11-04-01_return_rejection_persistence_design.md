# S11-04-01 Return And Supplier Rejection Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Task: S11-04-01 return and supplier rejection persistence design
Date: 2026-05-01
Status: Design complete; S11-04-02 and S11-04-03 can implement PostgreSQL stores

---

## 1. Goal

Persist the remaining return and supplier rejection owner documents so operational evidence remains traceable after API restart or dev redeploy.

Current risk after S11-03:

```text
stock movements, warehouse receiving, inbound QC, sales orders, and purchase orders persist
return receipt and supplier rejection owner documents still live in prototype stores
```

That can create this mismatch:

```text
return/rejection stock or audit evidence exists
-> owner document disappears from runtime after restart
-> daily board, detail page, QA review, and reconciliation trace weaken
```

S11-04 should close that gap without changing public API envelopes.

---

## 2. Current Runtime Contracts

Keep the current return store boundaries for S11-04-02:

```text
ReturnReceiptStore.List(ctx, filter)
ReturnReceiptStore.Save(ctx, receipt)
ReturnReceiptStore.FindExpectedReturnByCode(ctx, code)

ReturnInspectionStore.FindReceiptByID(ctx, id)
ReturnInspectionStore.SaveInspection(ctx, receipt, inspection)

ReturnDispositionStore.FindReceiptByID(ctx, id)
ReturnDispositionStore.SaveDisposition(ctx, receipt, action)

ReturnAttachmentStore.SaveAttachment(ctx, attachment)
```

Keep the current supplier rejection store boundary for S11-04-03:

```text
SupplierRejectionStore.List(ctx, filter)
SupplierRejectionStore.Get(ctx, id)
SupplierRejectionStore.Save(ctx, rejection)
```

Runtime selection should match the already merged owner document stores:

```text
DATABASE_URL present -> PostgreSQL store
DATABASE_URL empty   -> prototype store
```

---

## 3. Return Receipt Persistence Shape

Use existing `returns.return_orders` and `returns.return_order_lines` as the durable return receipt owner document.

Existing useful columns:

```text
returns.return_orders:
id, org_id, return_no, sales_order_id, shipment_id, customer_id, carrier_id,
warehouse_id, status, source, tracking_no, return_code, return_reason_code,
package_condition, initial_disposition, unknown_case, investigation_note,
received_at, received_by, created_at, created_by, updated_at, updated_by

returns.return_order_lines:
id, org_id, return_order_id, item_id, batch_id, source_order_line_id,
unit_id, returned_qty, return_reason_code, condition_code

returns.return_dispositions:
id, org_id, return_order_line_id, disposition, target_warehouse_id,
target_bin_id, inspected_at, inspected_by, approved_at, approved_by, reason
```

S11-04-02 should add runtime-safe refs and missing receipt fields:

```text
returns.return_orders:
return_ref text not blank
org_ref text not blank
warehouse_ref text
warehouse_code text
received_by_ref text
received_at timestamptz
disposition text
target_location text
original_order_ref text
original_order_no text
customer_name text
stock_movement_ref text
stock_movement_type text
target_stock_status text

returns.return_order_lines:
line_ref text not blank
sku_code text
product_name text
quantity numeric(18,6)
condition_text text
item_ref text
batch_ref text
unit_ref text
uom_code varchar(20)
stock_movement_ref text
```

Column relaxations follow the S11 owner-document pattern because runtime return refs are text:

```text
returns.return_orders warehouse_id drop not null after warehouse_ref backfill
returns.return_order_lines item_id drop not null after item_ref backfill
returns.return_order_lines unit_id drop not null after uom_code/unit_ref backfill
```

Add new tables for the child records currently held inside `PrototypeReturnReceiptStore`:

```text
returns.return_inspections:
  id uuid primary key
  org_id uuid not null
  inspection_ref text not null
  return_order_id uuid not null references returns.return_orders(id)
  return_ref text not null
  condition_code text not null
  disposition text not null
  target_location text not null
  risk_level text not null
  evidence_label text
  note text
  inspector_ref text not null
  inspected_at timestamptz not null
  created_at timestamptz not null

returns.return_disposition_actions:
  id uuid primary key
  org_id uuid not null
  action_ref text not null
  return_order_id uuid not null references returns.return_orders(id)
  return_ref text not null
  disposition text not null
  target_location text not null
  target_stock_status text
  action_code text not null
  note text
  actor_ref text not null
  decided_at timestamptz not null
  stock_movement_ref text
  created_at timestamptz not null

returns.return_attachments:
  id uuid primary key
  org_id uuid not null
  attachment_ref text not null
  return_order_id uuid not null references returns.return_orders(id)
  return_ref text not null
  inspection_ref text
  file_name text not null
  mime_type text
  file_size_bytes bigint not null
  storage_bucket text not null
  storage_key text not null
  uploaded_by_ref text not null
  uploaded_at timestamptz not null
  status text not null
  note text
  source text
```

Keep existing `returns.return_dispositions` for legacy line-level disposition evidence. The new
`returns.return_disposition_actions` table maps the current application-level disposition action and links back to the
return receipt by `return_ref`.

Recommended indexes:

```text
unique returns.return_orders(org_id, return_ref)
unique returns.return_order_lines(return_order_id, line_ref)
unique returns.return_inspections(org_id, inspection_ref)
unique returns.return_disposition_actions(org_id, action_ref)
unique returns.return_attachments(org_id, attachment_ref)
index returns.return_orders(org_id, warehouse_ref, status)
index returns.return_orders(org_id, tracking_no)
index returns.return_orders(org_id, return_code)
index returns.return_inspections(org_id, return_ref)
index returns.return_disposition_actions(org_id, return_ref)
```

---

## 4. Return Store Behavior

`PostgresReturnReceiptStore.List`:

```text
select persisted return headers
load lines and latest movement summary
apply ReturnReceiptFilter by warehouse_id/ref and status
sort with current domain SortReturnReceipts
```

`Save`:

```text
validate receipt has ID and scan identity
resolve org_ref using config default when needed
upsert returns.return_orders by (org_id, return_ref)
replace lines by return_order_id
store stock movement summary refs when receipt.StockMovement is present
```

`FindReceiptByID`:

```text
find by return_ref or id::text
load lines and movement summary
return ErrReturnReceiptNotFound when absent
```

`SaveInspection`:

```text
upsert the updated return receipt
insert/update returns.return_inspections by inspection_ref
keep receipt status/disposition/target_location in the return header
```

`SaveDisposition`:

```text
upsert the updated return receipt
insert/update returns.return_disposition_actions by action_ref
persist stock movement summary refs only; never write stock balances directly
if a stock movement is recorded, preserve the stock movement source_doc_ref as the return receipt ref
```

`SaveAttachment`:

```text
insert/update returns.return_attachments by attachment_ref
require the linked return receipt to exist
preserve inspection_ref when an inspection exists
```

`FindExpectedReturnByCode` remains a lookup, not the owner document. For S11-04-02 it can keep using the current expected-return seed behavior behind the PostgreSQL store until sales/shipping expected-return generation is implemented. The received return document must still persist all resolved expected-return data so runtime restart does not lose the receipt.

---

## 5. Supplier Rejection Persistence Shape

There is no current PostgreSQL table for supplier rejection owner documents. S11-04-03 should add:

```text
inventory.supplier_rejections:
  id uuid primary key
  org_id uuid not null references core.organizations(id)
  rejection_ref text not null
  org_ref text not null
  rejection_no text not null
  supplier_id uuid references mdm.suppliers(id)
  supplier_ref text not null
  supplier_code text
  supplier_name text not null
  purchase_order_id uuid references purchase.purchase_orders(id)
  purchase_order_ref text
  purchase_order_no text
  goods_receipt_id uuid references inventory.warehouse_receivings(id)
  goods_receipt_ref text not null
  goods_receipt_no text
  inbound_qc_inspection_id uuid references qc.inbound_qc_inspections(id)
  inbound_qc_inspection_ref text not null
  warehouse_id uuid references mdm.warehouses(id)
  warehouse_ref text not null
  warehouse_code text
  status text not null
  reason text not null
  created_by uuid references core.users(id)
  created_by_ref text not null
  updated_by uuid references core.users(id)
  updated_by_ref text
  submitted_at timestamptz
  submitted_by uuid references core.users(id)
  submitted_by_ref text
  confirmed_at timestamptz
  confirmed_by uuid references core.users(id)
  confirmed_by_ref text
  cancelled_at timestamptz
  cancelled_by uuid references core.users(id)
  cancelled_by_ref text
  cancel_reason text
  created_at timestamptz not null
  updated_at timestamptz not null

inventory.supplier_rejection_lines:
  id uuid primary key
  org_id uuid not null references core.organizations(id)
  rejection_id uuid not null references inventory.supplier_rejections(id) on delete restrict
  line_ref text not null
  line_no integer not null
  purchase_order_line_id uuid references purchase.purchase_order_lines(id)
  purchase_order_line_ref text
  goods_receipt_line_id uuid references inventory.warehouse_receiving_lines(id)
  goods_receipt_line_ref text not null
  inbound_qc_inspection_id uuid references qc.inbound_qc_inspections(id)
  inbound_qc_inspection_ref text not null
  item_id uuid references mdm.items(id)
  item_ref text not null
  sku_code text not null
  item_name text
  batch_id uuid references inventory.batches(id)
  batch_ref text not null
  batch_no text not null
  lot_no text not null
  expiry_date date not null
  rejected_qty numeric(18,6) not null
  uom_code varchar(20) not null references mdm.uoms(uom_code)
  base_uom_code varchar(20) not null references mdm.uoms(uom_code)
  reason text not null
  created_at timestamptz not null
  updated_at timestamptz not null

inventory.supplier_rejection_attachments:
  id uuid primary key
  org_id uuid not null references core.organizations(id)
  rejection_id uuid not null references inventory.supplier_rejections(id) on delete restrict
  attachment_ref text not null
  line_ref text
  file_name text not null
  object_key text not null
  content_type text
  uploaded_at timestamptz not null
  uploaded_by_ref text not null
  source text
```

Recommended constraints/indexes:

```text
status in ('draft', 'submitted', 'confirmed', 'cancelled')
rejected_qty > 0
unique inventory.supplier_rejections(org_id, rejection_ref)
unique inventory.supplier_rejections(org_id, rejection_no)
unique inventory.supplier_rejection_lines(rejection_id, line_ref)
unique inventory.supplier_rejection_attachments(org_id, attachment_ref)
index inventory.supplier_rejections(org_id, supplier_ref, status)
index inventory.supplier_rejections(org_id, warehouse_ref, status)
index inventory.supplier_rejections(org_id, inbound_qc_inspection_ref)
index inventory.supplier_rejection_lines(org_id, goods_receipt_line_ref)
```

---

## 6. Supplier Rejection Store Behavior

`PostgresSupplierRejectionStore.List`:

```text
select headers
load lines and attachments
apply SupplierRejectionFilter by supplier_id/ref, warehouse_id/ref, and status
sort with current domain SortSupplierRejections
```

`Get`:

```text
find by rejection_ref or id::text
load lines and attachments
return ErrSupplierRejectionNotFound when absent
```

`Save`:

```text
validate domain object
resolve org_ref using config default when needed
upsert header by (org_id, rejection_ref)
replace lines and attachments for that rejection
store UUID columns only when runtime refs are UUID-compatible
preserve lifecycle actors/timestamps and cancel reason
```

The store must not create stock movements or mutate available stock. Supplier rejection remains an owner document linked to failed inbound QC evidence.

---

## 7. Traceability Rules

Return receipt trace must survive restart:

```text
return receipt ref
-> return lines
-> inspection
-> disposition action
-> stock movement source_doc_ref
-> audit entity_ref
```

Supplier rejection trace must survive restart:

```text
supplier rejection ref
-> PO ref / PO line ref when present
-> goods receipt ref / goods receipt line ref
-> inbound QC inspection ref
-> rejected lines and attachments
-> audit entity_ref
```

Do not require runtime IDs to be UUIDs. Store both UUID columns and text refs, following the S10/S11 owner document pattern.

---

## 8. Verification Plan

S11-04-02 return receipt store should verify:

```text
PostgreSQL 16 migration apply/rollback
store selector prototype fallback and PostgreSQL mode
receive return persists header and lines
inspect return persists receipt status and inspection row
apply disposition persists action row and movement summary
return disposition stock movement uses UUID-compatible item/warehouse refs or a deliberate resolver before DB smoke
attachment metadata persists
go test ./internal/modules/returns/application ./cmd/api -count=1
go test ./... -count=1
go vet ./...
```

S11-04-03 supplier rejection store should verify:

```text
PostgreSQL 16 migration apply/rollback
store selector prototype fallback and PostgreSQL mode
create/submit/confirm/cancel persists lifecycle and actor refs
lines and attachments persist
filter/list/get behavior matches prototype semantics
go test ./internal/modules/inventory/application ./cmd/api -count=1
go test ./... -count=1
go vet ./...
```

S11-04-04 smoke should verify:

```text
return receipt create -> inspect -> disposition remains queryable in PostgreSQL
return stock movement/audit links use return receipt refs
supplier rejection create -> submit -> confirm remains queryable in PostgreSQL
supplier rejection links to PO, receiving, inbound QC, rejected lines, attachments, and audit
full dev smoke passes after redeploy
```

---

## 9. Acceptance

S11-04 is complete when:

```text
1. Return receipt owner documents persist in PostgreSQL in DB-backed runtime.
2. Return inspection, disposition, attachment metadata, movement summary, and audit trace remain linked by refs.
3. Supplier rejection owner documents persist in PostgreSQL in DB-backed runtime.
4. Supplier rejection lines and attachments remain linked to inbound QC failure evidence.
5. API envelopes and permissions remain unchanged.
6. Prototype fallback still works when DATABASE_URL is empty.
7. No direct stock balance writes are introduced.
8. Migration apply/rollback, backend tests, vet, CI, and dev smoke are green.
```
