# S13-01-01 End-of-Day Reconciliation Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 13 - End-of-day reconciliation persistence v1
Task: S13-01-01 End-of-day reconciliation persistence design
Date: 2026-05-02
Status: Design complete; S13-01-02 can introduce the runtime selector and S13-01-03 can implement the PostgreSQL store

---

## 1. Goal

Promote end-of-day reconciliation from prototype memory to PostgreSQL-backed persistence when database configuration exists.

Current risk:

```text
warehouse execution evidence -> increasingly persisted
audit log -> persisted in DB mode
end-of-day reconciliation list/close -> NewPrototypeEndOfDayReconciliationStore()
```

That can erase shift close status, close actor, close timestamp, checklist state, and variance line evidence after API restart or redeploy. Sprint 13 should make the existing reconciliation API durable without changing the public response envelope.

---

## 2. Existing Contract

The current application contract is already small enough to preserve:

```text
EndOfDayReconciliationStore
  List(ctx, filter) []domain.EndOfDayReconciliation
  Get(ctx, id) domain.EndOfDayReconciliation
  Save(ctx, reconciliation) error
```

Consumers today:

```text
GET  /api/v1/warehouse/end-of-day-reconciliations
POST /api/v1/warehouse/end-of-day-reconciliations/{reconciliation_id}/close
CloseEndOfDayReconciliation application service
```

Keep these response fields stable:

```text
id
warehouse_id
warehouse_code
date
shift_code
status
owner
closed_at
closed_by
summary
operations
checklist
lines
audit_log_id
```

---

## 3. PostgreSQL Source Tables

Reuse the existing header table:

```text
inventory.warehouse_daily_closings
```

The base schema already has:

```text
id
org_id
closing_no
warehouse_id
business_date
shift_code
status
orders_processed_count
pending_task_count
variance_count
exception_note
closed_at
closed_by
created_at
created_by
updated_at
updated_by
version
```

Sprint 13 should add the smallest durable runtime layer:

```text
inventory.warehouse_daily_closings.closing_ref text
inventory.warehouse_daily_closings.org_ref text
inventory.warehouse_daily_closings.warehouse_ref text
inventory.warehouse_daily_closings.warehouse_code text
inventory.warehouse_daily_closings.owner_ref text
inventory.warehouse_daily_closings.closed_by_ref text
inventory.warehouse_daily_closings.created_by_ref text
inventory.warehouse_daily_closings.updated_by_ref text
inventory.warehouse_daily_closings.handover_order_count integer
inventory.warehouse_daily_closings.return_order_count integer
inventory.warehouse_daily_closings.stock_movement_count integer
inventory.warehouse_daily_closings.stock_count_session_count integer
```

Add child tables:

```text
inventory.warehouse_daily_closing_checklist
inventory.warehouse_daily_closing_lines
```

Do not create a parallel reconciliation header table. The standards file already names `inventory.warehouse_daily_closings` as the end-of-day close record.

---

## 4. Domain Mapping

Header mapping:

| Domain field | PostgreSQL source |
| --- | --- |
| `ID` | `COALESCE(closing_ref, id::text)` |
| `WarehouseID` | `COALESCE(warehouse_ref, warehouse_id::text)` |
| `WarehouseCode` | `COALESCE(warehouse_code, '')` |
| `Date` | `business_date::text` |
| `ShiftCode` | `shift_code` |
| `Status` | `status` normalized to `open`, `in_review`, or `closed` |
| `Owner` | `COALESCE(owner_ref, created_by_ref, created_by::text, '')` |
| `Operations.OrderCount` | `orders_processed_count` |
| `Operations.HandoverOrderCount` | `handover_order_count` |
| `Operations.ReturnOrderCount` | `return_order_count` |
| `Operations.StockMovementCount` | `stock_movement_count` |
| `Operations.StockCountSessionCount` | `stock_count_session_count` |
| `Operations.PendingIssueCount` | `pending_task_count` |
| `ClosedAt` | `closed_at` |
| `ClosedBy` | `COALESCE(closed_by_ref, closed_by::text, '')` |

Checklist mapping:

| Domain field | PostgreSQL source |
| --- | --- |
| `Key` | `item_ref` |
| `Label` | `label` |
| `Complete` | `complete` |
| `Blocking` | `blocking` |
| `Note` | `note` |

Line mapping:

| Domain field | PostgreSQL source |
| --- | --- |
| `ID` | `line_ref` |
| `SKU` | `sku_code` |
| `BatchNo` | `batch_no` |
| `BinCode` | `bin_code` |
| `SystemQuantity` | `system_qty` |
| `CountedQuantity` | `counted_qty` |
| `Reason` | `reason` |
| `Owner` | `owner_ref` |

`system_qty` and `counted_qty` should be stored as `numeric(18,6)` to match the quantity standard. The current domain and response expose integer quantities, so S13 should store integer-valued decimals and scan them only when they are exactly representable as `int64`. A future decimal-domain task can widen the API deliberately.

---

## 5. Migration Plan

Use migration `000026_persist_end_of_day_reconciliations`.

Header changes:

```text
ALTER TABLE inventory.warehouse_daily_closings
  ADD COLUMN IF NOT EXISTS closing_ref text,
  ADD COLUMN IF NOT EXISTS org_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_ref text,
  ADD COLUMN IF NOT EXISTS warehouse_code text,
  ADD COLUMN IF NOT EXISTS owner_ref text,
  ADD COLUMN IF NOT EXISTS closed_by_ref text,
  ADD COLUMN IF NOT EXISTS created_by_ref text,
  ADD COLUMN IF NOT EXISTS updated_by_ref text,
  ADD COLUMN IF NOT EXISTS handover_order_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS return_order_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS stock_movement_count integer NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS stock_count_session_count integer NOT NULL DEFAULT 0;
```

Backfill rule:

```text
closing_ref = COALESCE(closing_ref, closing_no, id::text)
org_ref = COALESCE(org_ref, org_id::text)
warehouse_ref = COALESCE(warehouse_ref, warehouse_id::text)
created_by_ref = COALESCE(created_by_ref, created_by::text)
updated_by_ref = COALESCE(updated_by_ref, updated_by::text)
closed_by_ref = COALESCE(closed_by_ref, closed_by::text)
warehouse_code = existing mdm.warehouses.code when resolvable, otherwise existing value
owner_ref = COALESCE(owner_ref, created_by_ref, created_by::text)
```

Indexes and constraints:

```text
CREATE UNIQUE INDEX IF NOT EXISTS uq_warehouse_daily_closings_org_ref
  ON inventory.warehouse_daily_closings(org_id, lower(closing_ref))
  WHERE nullif(btrim(closing_ref), '') IS NOT NULL;

CREATE INDEX IF NOT EXISTS ix_warehouse_daily_closings_filters
  ON inventory.warehouse_daily_closings(org_id, warehouse_ref, business_date, shift_code, status);
```

Checklist table:

```text
inventory.warehouse_daily_closing_checklist
  id uuid primary key default gen_random_uuid()
  org_id uuid not null references core.organizations(id)
  closing_id uuid not null references inventory.warehouse_daily_closings(id) on delete cascade
  item_ref text not null
  label text not null
  complete boolean not null default false
  blocking boolean not null default true
  note text
  created_at timestamptz not null default now()
  updated_at timestamptz not null default now()
  unique(org_id, closing_id, item_ref)
```

Line table:

```text
inventory.warehouse_daily_closing_lines
  id uuid primary key default gen_random_uuid()
  org_id uuid not null references core.organizations(id)
  closing_id uuid not null references inventory.warehouse_daily_closings(id) on delete cascade
  line_ref text not null
  line_no integer not null default 1
  sku_code text not null
  batch_no text
  bin_code text
  system_qty numeric(18,6) not null default 0
  counted_qty numeric(18,6) not null default 0
  reason text
  owner_ref text
  created_at timestamptz not null default now()
  updated_at timestamptz not null default now()
  unique(org_id, closing_id, line_ref)
```

Rollback drops the child tables, drops added indexes, and removes only the S13-added header columns.

---

## 6. Store Behavior

Add `PostgresEndOfDayReconciliationStore` in the inventory application package.

Read behavior:

```text
List:
  filter by warehouse_ref or warehouse_id::text when warehouse_id is provided
  filter by business_date when date is provided
  filter by shift_code when shift_code is provided
  filter by normalized status when status is provided
  load checklist and lines per returned closing
  apply domain.SortEndOfDayReconciliations before returning

Get:
  lookup by closing_ref or id::text
  return ErrEndOfDayReconciliationNotFound for missing rows
  load checklist and lines
```

Write behavior:

```text
Save:
  validate non-empty reconciliation ID
  resolve org id from default org in dev/static-auth mode
  resolve warehouse uuid when WarehouseID is a UUID or matching warehouse_ref can be found
  upsert header by (org_id, closing_ref)
  replace checklist rows for that closing
  replace line rows for that closing
  run in a serializable transaction
```

The first implementation may seed the PostgreSQL table from the existing prototype fixtures when DB mode starts and no matching persisted rows exist. That keeps the dev/demo data available while making every close durable after the first write.

Do not weaken `EndOfDayReconciliation.Close`. Persistence is an adapter concern only.

---

## 7. Runtime Selection

Add runtime selector:

```text
newRuntimeEndOfDayReconciliationStore(cfg)
```

Selection rule:

```text
DATABASE_URL present -> PostgresEndOfDayReconciliationStore
DATABASE_URL empty   -> PrototypeEndOfDayReconciliationStore
```

Close function behavior:

```text
DB mode returns db.Close.
No-DB mode returns nil.
```

`main.go` should replace:

```text
inventoryapp.NewPrototypeEndOfDayReconciliationStore()
```

with the runtime selector. The selector should be closed in the same cleanup chain as neighboring inventory stores.

---

## 8. Audit Behavior

Keep the existing application audit action:

```text
Action: warehouse.shift.closed
EntityType: inventory.warehouse_daily_closing
EntityID: reconciliation.ID
```

The PostgreSQL reconciliation store should not write a second audit log by itself. The `CloseEndOfDayReconciliation` application service already writes audit after `Save`.

Successful close sequence:

```text
Get persisted reconciliation
Domain Close validates status/checklist/issues
Save persisted CLOSED status and close evidence
Record persisted audit log through runtime audit store
Return reconciliation response with audit_log_id
```

If audit recording fails after `Save`, the current service can still leave the close persisted. S13 should not change that transaction boundary unless a separate task deliberately introduces cross-store transaction coordination.

---

## 9. Tests And Smoke

S13-01-04 should add PostgreSQL integration coverage:

```text
nil DB returns explicit error
List filters by warehouse/date/shift/status
Get supports closing_ref and id::text
Save persists CLOSED status, closed_at, closed_by, checklist, and lines
Close blocks unresolved operational issues
Close allows variance exception note when domain permits it
Close writes audit through PostgresLogStore
Runtime selector uses PostgreSQL when DATABASE_URL exists and prototype fallback when empty
```

S13-02-01 dev smoke should prove:

```text
Query open reconciliation
Close it through POST /api/v1/warehouse/end-of-day-reconciliations/{id}/close
Confirm inventory.warehouse_daily_closings status/closed_at/closed_by_ref changed
Restart or redeploy API
Query reconciliation again
Confirm CLOSED status and line/checklist evidence survived restart
Confirm audit row exists for warehouse.shift.closed
```

---

## 10. Implementation Notes

Keep S13 implementation minimal:

```text
Do not change response structs.
Do not add new endpoints.
Do not add frontend state.
Do not alter stock movement or stock balance code.
Do not change domain close rules.
Do not replace unrelated prototype stores.
```

Expected task split:

```text
S13-01-02 runtime selector and store adapter wiring
S13-01-03 migration and PostgreSQL store implementation
S13-01-04 focused tests
S13-02-01 dev close persistence smoke
S13-03-01 remaining prototype ledger update
S13-04-01 changelog and release evidence
```

