# S14-01-01 Shipping Pick/Pack Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 14 - Shipping manifest, pick, and pack persistence v1
Task: S14-01-01 Shipping persistence design
Date: 2026-05-02
Status: Design complete; S14-01-02 can introduce runtime selectors and S14-01-03 can implement the carrier manifest PostgreSQL store

---

## 1. Goal

Promote shipping execution state from prototype memory to PostgreSQL-backed persistence when database configuration exists.

Current restart risk:

```text
sales orders / reservations / stock movements / returns / daily close -> persisted
carrier manifests -> NewPrototypeCarrierManifestStore()
pick tasks -> NewPrototypePickTaskStore(...)
pack tasks -> NewPrototypePackTaskStore(...)
```

That can erase manifest membership, handover scan evidence, pick line progress, pack confirmation, and exception state after API restart or redeploy. Sprint 14 should make those existing APIs durable without changing response envelopes or sales order adapter behavior.

---

## 2. Existing Contracts

Carrier manifest store:

```text
CarrierManifestStore
  List(ctx, filter) []domain.CarrierManifest
  Get(ctx, id) domain.CarrierManifest
  Save(ctx, manifest) error
  GetPackedShipment(ctx, id) domain.PackedShipment
  FindPackedShipmentByCode(ctx, code) domain.PackedShipment
  FindCarrierManifestLineByCode(ctx, code) domain.CarrierManifest + line
  RecordScanEvent(ctx, event) error
```

Pick task store:

```text
PickTaskStore
  ListPickTasks(ctx) []domain.PickTask
  GetPickTask(ctx, id) domain.PickTask
  GetPickTaskBySalesOrder(ctx, salesOrderID) domain.PickTask
  SavePickTask(ctx, task) error
```

Pack task store:

```text
PackTaskStore
  ListPackTasks(ctx) []domain.PackTask
  GetPackTask(ctx, id) domain.PackTask
  GetPackTaskBySalesOrder(ctx, salesOrderID) domain.PackTask
  GetPackTaskByPickTask(ctx, pickTaskID) domain.PackTask
  SavePackTask(ctx, task) error
```

Current public routes must keep their response shape:

```text
GET  /api/v1/shipping/manifests
POST /api/v1/shipping/manifests
POST /api/v1/shipping/manifests/{id}/shipments
POST /api/v1/shipping/manifests/{id}/ready
POST /api/v1/shipping/manifests/{id}/confirm-handover
POST /api/v1/shipping/manifests/{id}/scan
GET  /api/v1/pick-tasks
POST /api/v1/pick-tasks/{id}/start
POST /api/v1/pick-tasks/{id}/confirm-line
POST /api/v1/pick-tasks/{id}/complete
GET  /api/v1/pack-tasks
POST /api/v1/pack-tasks/{id}/start
POST /api/v1/pack-tasks/{id}/confirm
```

Do not change domain transition rules. Persistence is an adapter concern.

---

## 3. PostgreSQL Source Tables

Reuse the existing shipping tables:

```text
shipping.carrier_manifests
shipping.carrier_manifest_orders
shipping.scan_events
shipping.shipments
shipping.pick_tasks
shipping.pick_task_lines
shipping.pack_tasks
shipping.pack_task_lines
```

`shipping.carrier_manifest_orders` should be the runtime line source for manifests. It already supports text order/tracking/package data and optional shipment/sales order UUIDs. The older `shipping.carrier_manifest_lines` table can remain for legacy/schema compatibility, but it requires a non-null shipment UUID and does not fit current text-ref prototype IDs as cleanly.

Do not create parallel header tables. Sprint 14 should add the smallest runtime ref layer on top of the existing tables.

---

## 4. Runtime Ref Gap

The domain and public APIs use stable text IDs:

```text
manifest-hcm-ghn-morning
ship-hcm-260426-001
pick-so-...
pack-so-...
user-warehouse-lead
wh-hcm
```

The base PostgreSQL schema stores many of the same fields as UUID foreign keys. The persistence adapter must preserve public text IDs by adding and reading `*_ref` columns while keeping UUID columns populated when a matching persisted record exists.

Lookup rule for every persisted aggregate:

```text
WHERE lower(COALESCE(<aggregate>_ref, id::text)) = lower($1)
   OR id::text = $1
```

Reference resolution rule:

```text
input is UUID -> use as UUID
input is text ref -> resolve by matching existing *_ref/code/no where available
cannot resolve required FK -> return explicit persistence error before writing partial state
```

This keeps UUID integrity for persisted sales/order/reservation data and avoids weakening existing constraints just to store prototype text values.

---

## 5. Migration Plan

Use the next migration after `000026_persist_end_of_day_reconciliations`.

Carrier manifest header additions:

```text
shipping.carrier_manifests.manifest_ref text
shipping.carrier_manifests.org_ref text
shipping.carrier_manifests.warehouse_ref text
shipping.carrier_manifests.warehouse_code text
shipping.carrier_manifests.carrier_ref text
shipping.carrier_manifests.carrier_code text
shipping.carrier_manifests.carrier_name text
shipping.carrier_manifests.owner_ref text
shipping.carrier_manifests.completed_by_ref text
shipping.carrier_manifests.handed_over_by_ref text
shipping.carrier_manifests.created_by_ref text
shipping.carrier_manifests.updated_by_ref text
```

Carrier manifest order additions:

```text
shipping.carrier_manifest_orders.line_ref text
shipping.carrier_manifest_orders.manifest_ref text
shipping.carrier_manifest_orders.shipment_ref text
shipping.carrier_manifest_orders.sales_order_ref text
shipping.carrier_manifest_orders.scanned_by_ref text
shipping.carrier_manifest_orders.created_by_ref text
shipping.carrier_manifest_orders.updated_by_ref text
```

Scan event additions:

```text
shipping.scan_events.scan_ref text
shipping.scan_events.manifest_ref text
shipping.scan_events.expected_manifest_ref text
shipping.scan_events.shipment_ref text
shipping.scan_events.actor_ref text
shipping.scan_events.warehouse_ref text
shipping.scan_events.carrier_code text
```

Pick task additions:

```text
shipping.pick_tasks.pick_ref text
shipping.pick_tasks.org_ref text
shipping.pick_tasks.sales_order_ref text
shipping.pick_tasks.order_no text
shipping.pick_tasks.warehouse_ref text
shipping.pick_tasks.warehouse_code text
shipping.pick_tasks.assigned_to_ref text
shipping.pick_tasks.started_by_ref text
shipping.pick_tasks.completed_by_ref text
shipping.pick_tasks.created_by_ref text
shipping.pick_tasks.updated_by_ref text
```

Pick task line additions:

```text
shipping.pick_task_lines.line_ref text
shipping.pick_task_lines.pick_task_ref text
shipping.pick_task_lines.sales_order_line_ref text
shipping.pick_task_lines.stock_reservation_ref text
shipping.pick_task_lines.item_ref text
shipping.pick_task_lines.batch_ref text
shipping.pick_task_lines.batch_no text
shipping.pick_task_lines.warehouse_ref text
shipping.pick_task_lines.bin_ref text
shipping.pick_task_lines.bin_code text
shipping.pick_task_lines.picked_by_ref text
shipping.pick_task_lines.created_by_ref text
shipping.pick_task_lines.updated_by_ref text
```

Pack task additions:

```text
shipping.pack_tasks.pack_ref text
shipping.pack_tasks.org_ref text
shipping.pack_tasks.sales_order_ref text
shipping.pack_tasks.order_no text
shipping.pack_tasks.pick_task_ref text
shipping.pack_tasks.pick_task_no text
shipping.pack_tasks.warehouse_ref text
shipping.pack_tasks.warehouse_code text
shipping.pack_tasks.assigned_to_ref text
shipping.pack_tasks.started_by_ref text
shipping.pack_tasks.packed_by_ref text
shipping.pack_tasks.created_by_ref text
shipping.pack_tasks.updated_by_ref text
```

Pack task line additions:

```text
shipping.pack_task_lines.line_ref text
shipping.pack_task_lines.pack_task_ref text
shipping.pack_task_lines.pick_task_line_ref text
shipping.pack_task_lines.sales_order_line_ref text
shipping.pack_task_lines.item_ref text
shipping.pack_task_lines.batch_ref text
shipping.pack_task_lines.batch_no text
shipping.pack_task_lines.warehouse_ref text
shipping.pack_task_lines.packed_by_ref text
shipping.pack_task_lines.created_by_ref text
shipping.pack_task_lines.updated_by_ref text
```

Backfill pattern:

```text
<aggregate>_ref = COALESCE(NULLIF(btrim(<aggregate>_ref), ''), <business_no>, id::text)
org_ref = COALESCE(NULLIF(btrim(org_ref), ''), org_id::text)
warehouse_ref = COALESCE(NULLIF(btrim(warehouse_ref), ''), warehouse_id::text)
actor refs = COALESCE(existing text ref, uuid_column::text)
```

Required indexes:

```text
UNIQUE (org_id, lower(manifest_ref)) WHERE manifest_ref is not blank
UNIQUE (org_id, lower(pick_ref)) WHERE pick_ref is not blank
UNIQUE (org_id, lower(pack_ref)) WHERE pack_ref is not blank
INDEX carrier manifest filters on org_id, warehouse_ref, handover_date, carrier_code, status
INDEX pick filters on org_id, warehouse_ref, status, assigned_to_ref, created_at
INDEX pack filters on org_id, warehouse_ref, status, assigned_to_ref, created_at
INDEX scan lookup on org_id, lower(scan_ref)
INDEX scan barcode lookup on scan_context, barcode, scanned_at
```

Rollback drops only S14-added indexes and columns.

---

## 6. Domain Mapping

Carrier manifest header:

| Domain field | PostgreSQL source |
| --- | --- |
| `ID` | `COALESCE(manifest_ref, manifest_no, id::text)` |
| `CarrierCode` | `COALESCE(carrier_code, mdm.carriers.code)` |
| `CarrierName` | `COALESCE(carrier_name, mdm.carriers.name, carrier_code)` |
| `WarehouseID` | `COALESCE(warehouse_ref, warehouse_id::text)` |
| `WarehouseCode` | `COALESCE(warehouse_code, mdm.warehouses.code)` |
| `Date` | `handover_date::text` |
| `HandoverBatch` | `handover_batch` |
| `StagingZone` | `handover_zone` |
| `HandoverZoneID` | `handover_zone_id::text` |
| `HandoverZoneCode` | `handover_zone_code` |
| `HandoverBinID` | `handover_bin_id::text` |
| `HandoverBinCode` | `handover_bin_code` |
| `Status` | normalized `status` |
| `Owner` | `COALESCE(owner_ref, created_by_ref, created_by::text, '')` |
| `CreatedAt` | `created_at` |

Carrier manifest line:

| Domain field | PostgreSQL source |
| --- | --- |
| `ID` | `COALESCE(line_ref, id::text)` |
| `ShipmentID` | `COALESCE(shipment_ref, shipment_id::text, '')` |
| `OrderNo` | `order_no` |
| `TrackingNo` | `tracking_no` |
| `PackageCode` | `package_code` |
| `StagingZone` | `staging_zone` |
| `HandoverZoneID` | `handover_zone_id::text` |
| `HandoverZoneCode` | `handover_zone_code` |
| `HandoverBinID` | `handover_bin_id::text` |
| `HandoverBinCode` | `handover_bin_code` |
| `Scanned` | `scan_status = 'scanned'` |

Pick and pack task mapping follows the same rule:

```text
header ID -> COALESCE(pick_ref/pack_ref, task_no, id::text)
line ID -> COALESCE(line_ref, id::text)
foreign refs -> COALESCE(<thing>_ref, <thing>_id::text)
quantities -> numeric(18,6) string parsed through decimal quantity helpers
timestamps/actors -> persisted timestamp columns plus *_by_ref text columns
```

Scan event mapping:

```text
event.ID -> COALESCE(scan_ref, idempotency_key, id::text)
event.ManifestID -> COALESCE(manifest_ref, carrier_manifest_id::text)
event.Code -> barcode
event.ResultCode -> map scan_result to domain result code
event.ActorID -> COALESCE(actor_ref, scanned_by::text, '')
station/device/source/severity/message/order/tracking -> metadata jsonb plus explicit station/barcode fields
```

---

## 7. Store Behavior

Add three PostgreSQL store types in `apps/api/internal/modules/shipping/application`:

```text
PostgresCarrierManifestStore
PostgresPickTaskStore
PostgresPackTaskStore
```

Read behavior:

```text
List/Get load header and lines.
Apply existing domain sort functions before returning.
Map sql.ErrNoRows to the existing not-found errors.
Use text refs first and UUID id fallback.
Keep filters in application behavior compatible with prototype behavior.
```

Write behavior:

```text
Save validates the domain object first.
Resolve org, warehouse, carrier, sales order, reservation, batch, bin, user UUIDs when possible.
Upsert header by org_id + ref.
Replace aggregate-owned lines for the saved header inside the same transaction.
Do not create or mutate stock balances directly.
Do not write audit logs inside stores.
Use serializable transactions for Save operations that replace child rows.
```

Packed shipment lookup:

```text
Prefer shipping.shipments joined to sales.sales_orders/mdm.carriers/mdm.warehouses.
Use shipment_ref/tracking_no/shipment_no/order_no/package code lookup where available.
Only return PackedShipment.Packed = true when shipment status is packed, ready_for_handover, handed_over, or delivered.
```

If the required referenced sales order, reservation, batch, warehouse, bin, or carrier cannot be resolved for a write, fail the write and leave the current aggregate unchanged.

---

## 8. Runtime Selection

Add runtime selectors:

```text
newRuntimeCarrierManifestStore(cfg)
newRuntimePickTaskStore(cfg)
newRuntimePackTaskStore(cfg)
```

Selection rule:

```text
DATABASE_URL present -> PostgreSQL store
DATABASE_URL empty -> existing prototype store
```

The selectors should follow the existing `newRuntimeEndOfDayReconciliationStore` pattern:

```text
sql.Open("pgx", cfg.DatabaseURL)
DefaultOrgID = localAuditOrgID only when static-auth/local access is allowed
return db.Close as the cleanup function in DB mode
```

`main.go` should replace direct constructors:

```text
shippingapp.NewPrototypeCarrierManifestStore()
shippingapp.NewPrototypePickTaskStore(...)
shippingapp.NewPrototypePackTaskStore(...)
```

with selectors. Prototype fixtures remain only for no-DB/local mode and tests that explicitly need in-memory stores.

---

## 9. Audit And Adapter Behavior

Keep existing application audit actions:

```text
shipping.manifest.created
shipping.manifest.shipment_added
shipping.manifest.ready_to_scan
shipping.manifest.scan_recorded
shipping.manifest.handed_over
shipping.pick_task.created
shipping.pick_task.started
shipping.pick_task.line_confirmed
shipping.pick_task.completed
shipping.pick_task.exception_reported
shipping.pack_task.created
shipping.pack_task.started
shipping.pack_task.confirmed
shipping.pack_task.exception_reported
```

Stores must not write duplicate audit logs. The application services already record audit after successful store writes.

Keep adapter behavior unchanged:

```text
ConfirmCarrierManifestHandover -> CarrierManifestSalesOrderHandover
ConfirmPackTask -> PackTaskSalesOrderPacker
```

If adapter updates succeed and the subsequent shipping store save fails, report the persistence error honestly. Do not silently swallow it. Cross-store transaction coordination is a future task and should not be invented inside Sprint 14.

---

## 10. Tests And Smoke

Focused backend tests:

```text
nil DB returns explicit database-required errors
runtime selectors choose prototype without DATABASE_URL and PostgreSQL with DATABASE_URL
carrier manifest Save/Get/List persists header, lines, ready/scanning/handed_over statuses
manifest scan match persists line scan status and scan event
manifest mismatch/not-found scan persists scan event without changing expected line state
pick Save/Get/List persists assigned/start/line picked/complete/exception evidence
pack Save/Get/List persists start/line packed/complete/exception evidence
fresh store instance can reload records saved by previous store instance
audit action still comes from application service, not store
```

Migration checks:

```text
apply all migrations on PostgreSQL 16
roll back S14 migration on PostgreSQL 16
rerun apply after rollback
```

Dev smoke:

```text
Manifest:
  create or load manifest
  add packed shipment
  mark ready
  scan expected package/tracking code
  confirm handover
  restart/redeploy API
  confirm status, scanned line, and scan event remain in PostgreSQL

Pick:
  generate or load pick task from reserved order
  start task
  confirm one line
  complete when all lines are picked
  restart/redeploy API
  confirm status, picked qty, picker, and timestamps remain

Pack:
  generate or load pack task from completed pick task
  start task
  confirm pack
  restart/redeploy API
  confirm packed status, packed qty, packer, and sales order packed adapter behavior remain
```

---

## 11. Implementation Split

Recommended task split:

```text
S14-01-02 runtime selectors and contracts
S14-01-03 carrier manifest migration and PostgreSQL store
S14-01-04 pick task PostgreSQL store
S14-01-05 pack task PostgreSQL store
S14-01-06 store and handler tests
S14-02-01 manifest handover/scan persistence smoke
S14-02-02 pick/pack persistence smoke
S14-03-01 shipping integration check
S14-04-01 remaining prototype ledger update
S14-05-01 Sprint 14 changelog and release evidence
```

Implementation guardrails:

```text
No public API response shape change.
No frontend state counted as persistence.
No direct stock balance writes.
No broad shipping refactor.
No cosmetic formatting mixed into behavior changes.
No production tag until CI, dev smoke, and migration apply/rollback evidence are green.
```
