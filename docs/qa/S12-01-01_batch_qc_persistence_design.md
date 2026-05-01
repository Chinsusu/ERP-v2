# S12-01-01 Batch/QC Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 12 - Batch/QC status persistence v1
Task: S12-01-01 Batch/QC persistence design
Date: 2026-05-01
Status: Design complete; S12-01-02 can introduce the runtime interface/selector and S12-01-03 can implement the PostgreSQL store

---

## 1. Goal

Promote the runtime batch catalog and QC status changes from prototype memory to PostgreSQL-backed persistence when database configuration exists.

Current risk:

```text
inbound QC and stock movement evidence -> persisted
available stock read model -> reads inventory.stock_balances joined to inventory.batches
batch catalog / QC transition API -> still uses NewPrototypeBatchCatalog(auditLogStore)
```

That can create a correctness mismatch after restart/redeploy: the user changes or derives a batch QC status, but only the prototype catalog remembers it. Persisted stock availability continues to read `inventory.batches`.

---

## 2. Existing Contract

The current prototype catalog exposes the behavior needed by the API and services:

```text
ListBatches(ctx, filter) []domain.Batch
GetBatch(ctx, id) domain.Batch
ChangeQCStatus(ctx, input) ChangeBatchQCStatusResult
ListQCTransitions(ctx, batchID) []domain.BatchQCTransition
```

Consumers today:

```text
GET /api/v1/inventory/batches
GET /api/v1/inventory/batches/{batch_id}
GET /api/v1/inventory/batches/{batch_id}/qc-transitions
POST /api/v1/inventory/batches/{batch_id}/qc-transitions
WarehouseReceivingService batch reader
InboundQCInspectionService batch QC status updater
```

Keep the public API response shape unchanged.

S12-01-02 should split the concrete prototype type from the runtime contract with a small interface:

```text
BatchCatalogStore
  ListBatches
  GetBatch
  ChangeQCStatus
  ListQCTransitions
```

The existing `WarehouseReceivingBatchReader` can remain as the narrower read contract because it already only requires `GetBatch`.

---

## 3. PostgreSQL Source Tables

Primary source:

```text
inventory.batches
```

Read joins:

```text
mdm.items      -> SKU and item name
mdm.suppliers  -> supplier identity when needed
audit.audit_logs via audit.LogStore -> QC transition history
```

Read columns mapped to `domain.Batch`:

| Domain field | PostgreSQL source |
| --- | --- |
| `ID` | `COALESCE(batch_ref, batches.id::text)` after ref migration, otherwise `batches.id::text` |
| `OrgID` | `COALESCE(org_ref, batches.org_id::text)` after ref migration, otherwise `batches.org_id::text` |
| `ItemID` | `COALESCE(item_ref, batches.item_id::text)` after ref migration, otherwise `batches.item_id::text` |
| `SKU` | `items.sku` |
| `ItemName` | `items.name` |
| `BatchNo` | `batches.batch_no` |
| `SupplierID` | `COALESCE(supplier_ref, batches.supplier_id::text, '')` after ref migration, otherwise `COALESCE(batches.supplier_id::text, '')` |
| `MfgDate` | `batches.mfg_date` |
| `ExpiryDate` | `batches.expiry_date` |
| `QCStatus` | `batches.qc_status` |
| `Status` | `batches.status` |
| `CreatedAt` | `batches.created_at` |
| `UpdatedAt` | `batches.updated_at` |

Do not read from `inventory.stock_balances` for the batch catalog. Stock balances are the availability read model; the batch catalog is the batch master/QC state source.

---

## 4. Runtime Ref Gap

There is a schema gap to close before the DB-backed catalog can safely replace the prototype:

```text
inventory.batches.id is uuid
existing API/tests/frontend prototype IDs often use stable text refs like batch-serum-2604a
persisted owner documents already carry batch_ref text in warehouse receiving, inbound QC, sales orders, returns, and supplier rejections
inventory.batches currently has no batch_ref text column
```

S12 implementation should add the smallest migration needed to align batch master data with runtime refs:

```text
inventory.batches.batch_ref text
inventory.batches.org_ref text
inventory.batches.item_ref text
inventory.batches.supplier_ref text
inventory.batches.created_by_ref text
inventory.batches.updated_by_ref text
```

Backfill rule:

```text
batch_ref = COALESCE(batch_ref, id::text)
org_ref = COALESCE(org_ref, org_id::text)
item_ref = COALESCE(item_ref, item_id::text)
supplier_ref = COALESCE(supplier_ref, supplier_id::text)
created_by_ref = COALESCE(created_by_ref, created_by::text)
updated_by_ref = COALESCE(updated_by_ref, updated_by::text)
```

Lookup rule:

```text
WHERE lower(COALESCE(batch_ref, id::text)) = lower($1)
   OR id::text = $1
```

This preserves UUID compatibility and supports existing runtime text refs without changing API envelopes.

---

## 5. Filters

Preserve current batch filter behavior:

```text
sku case-insensitive exact match
qc_status normalized exact match
status normalized exact match
```

Org scoping is not added in this task because the current batch API does not carry org context. The DB store should return all matching rows under current behavior. Org scoping belongs with auth/session org propagation, not this persistence promotion.

Sort order should match `domain.SortBatches`:

```text
sku ascending
expiry_date with non-null dates first
expiry_date ascending
batch_no ascending
```

---

## 6. ChangeQCStatus Behavior

The PostgreSQL-backed implementation must preserve the domain transition rules:

```text
Read current batch.
Normalize and validate next QC status through domain.Batch.ChangeQCStatus.
Reject missing actor or reason before any write.
Persist inventory.batches.qc_status and updated_at.
Write audit action inventory.batch.qc_status_changed.
Return updated batch and transition response.
```

Audit shape should remain compatible with the prototype:

```text
EntityType: inventory.batch
Action: inventory.batch.qc_status_changed
BeforeData: qc_status, updated_at
AfterData: qc_status, updated_at
Metadata: batch_no, business_ref, reason, sku
```

Do not change `inventory.batches.status` automatically during S12 unless the domain rules are deliberately extended in a separate reviewed task. Existing availability and reservation checks already block non-PASS QC status even when batch status is `active`.

---

## 7. Runtime Selection

Add runtime selector:

```text
newRuntimeBatchCatalogStore(cfg, auditLogStore)
```

Selection rule:

```text
DATABASE_URL present -> PostgresBatchCatalogStore
DATABASE_URL empty   -> PrototypeBatchCatalog
```

The selector should follow the existing Sprint 10/11 pattern used by available stock, warehouse receiving, inbound QC, purchase orders, sales orders, returns, and supplier rejections.

Close function behavior:

```text
DB mode returns db.Close.
No-DB mode returns nil.
```

---

## 8. Inbound QC Integration

Current path:

```text
InboundQCInspectionService.updateBatchQCStatus
-> inboundQCBatchQCStatusAdapter
-> BatchCatalog.ChangeQCStatus
```

S12 should keep this path but make the adapter depend on the runtime interface instead of the concrete prototype pointer.

Expected DB-mode behavior:

```text
Inbound QC PASS    -> inventory.batches.qc_status = pass
Inbound QC FAIL    -> inventory.batches.qc_status = fail
Inbound QC HOLD    -> inventory.batches.qc_status = quarantine
Inbound QC PARTIAL -> pass when passed_qty > 0, otherwise quarantine
```

Partial quantity stock movements remain owned by the inbound QC service. Batch catalog only owns the batch QC status.

---

## 9. Available Stock And Reservation Consistency

`PostgresStockAvailabilityStore` already reads:

```text
inventory.stock_balances
LEFT JOIN inventory.batches AS batch ON batch.id = balance.batch_id
```

Sales reservation sellability already rejects rows when:

```text
stock_status != available
batch.status != active
batch.expiry_date is expired
batch.qc_status is not pass
```

S12 should not add a new availability algorithm. The consistency requirement is simpler:

```text
Batch QC transition writes inventory.batches.qc_status.
Available stock/reservation reads see that same persisted qc_status through the existing join.
```

---

## 10. Tests

S12-01-02 and S12-01-03 should include:

```text
1. Interface/selector test proving DB config selects PostgreSQL and empty DB config selects prototype.
2. Postgres batch catalog query mapping tests for ListBatches and GetBatch.
3. Filter tests for SKU, qc_status, and batch status.
4. ChangeQCStatus tests for valid transition, missing actor, missing reason, missing batch, and invalid transition.
5. Audit compatibility test proving ListQCTransitions can read the persisted audit event by batch ref.
6. Inbound QC adapter test proving the adapter works against the interface, not the prototype concrete type.
```

If a PostgreSQL integration test is practical, seed the minimum rows required:

```text
core.organizations
mdm.items
inventory.batches
audit.audit_logs
```

Otherwise, a SQL query-runner unit test is acceptable for the implementation PR, with S12-02-01 covering dev smoke against the real dev database.

---

## 11. S12-01-02 And S12-01-03 Acceptance

S12-01-02 is complete when:

```text
1. Handlers and inbound QC adapter depend on a runtime batch catalog interface.
2. Prototype fallback still passes existing handler/service tests.
3. Runtime selector exists and is tested.
4. No public API shape changes are introduced.
```

S12-01-03 is complete when:

```text
1. DB mode reads batches from inventory.batches.
2. DB mode writes QC status changes to inventory.batches.
3. QC transition audit remains queryable through ListQCTransitions.
4. Runtime text refs are supported for batch lookup.
5. Migration apply/rollback passes if ref columns are added.
6. Backend tests and vet pass.
7. No direct stock balance writes are introduced.
```
