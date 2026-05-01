# 62_ERP_Sprint12_Changelog_Batch_QC_Status_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 12 - Batch/QC status persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-01
Status: Release evidence prepared; production tag pending after this changelog merges and main CI is green

---

## 1. Sprint 12 Scope

Sprint 12 closed the batch/QC persistence mismatch left after Sprint 11:

```text
available stock reads persisted inventory.batches
inbound QC and stock movement evidence persisted
batch catalog / QC transition path could still be prototype-only
```

Promoted scope:

```text
batch catalog runtime interface and selector
PostgreSQL-backed batch catalog reads/writes
persisted batch QC transition audit history
available-stock consistency after batch QC transition
inbound QC batch-status persistence integration
remaining prototype store ledger update
```

No new user workflow or API envelope was introduced. Sprint 12 changed persistence behavior behind existing batch, inbound QC, and inventory reads.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S12-00-00 Sprint 12 task board | #422 | Created Sprint 12 task board |
| S12-01-01 Batch/QC persistence design | #423 | Documented `inventory.batches`, audit, available-stock, reservation, and inbound QC mapping |
| S12-01-02 Batch catalog runtime interface and selector | #424 | Batch handlers/services depend on `BatchCatalogStore`; DB-mode selector added |
| S12-01-03 PostgreSQL batch catalog store | #425 | List/Get/ChangeQCStatus read and write `inventory.batches`; audit history remains persisted |
| S12-01-04 Batch catalog store tests | #426 | Added PostgreSQL integration coverage for QC transition persistence and audit |
| S12-02-01 Batch QC transition persistence smoke | #427 | Dev smoke proves QC transition persists and survives API restart |
| S12-02-02 Available-stock QC consistency smoke | #428 | Dev smoke proves available-stock reads persisted batch QC status after transition |
| S12-03-01 Inbound QC batch-status persistence integration | #429 | Added adapter tests proving inbound QC updater writes through PostgreSQL batch catalog in DB mode |
| S12-03-02 Inbound QC batch-status smoke | #430 | Dev smoke proves PASS/FAIL/PARTIAL consistency across batch detail, available stock, audit, and restart |
| S12-04-01 Remaining prototype ledger update | #431 | Superseded Sprint 11 ledger with Sprint 12 remaining-store ledger |

All PRs used the manual review and merge flow.

---

## 3. Persistence Changes

### Runtime Selector

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeBatchCatalogStore` | `PostgresBatchCatalogStore` | `PrototypeBatchCatalog` |

Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### New PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000025_persist_batch_runtime_refs` | Adds runtime-safe batch refs to `inventory.batches` so existing text refs and UUID refs can both use the persisted batch catalog |

Persisted behavior:

```text
GET /api/v1/inventory/batches
GET /api/v1/inventory/batches/{batch_id}
GET /api/v1/inventory/batches/{batch_id}/qc-transitions
POST /api/v1/inventory/batches/{batch_id}/qc-transitions
Inbound QC batch-status updater
Available-stock reads joined to inventory.batches
```

---

## 4. Dev Release Evidence

Dev server:

```text
Host: 10.1.1.120
Repo: /opt/ERP-v2
Runtime dev URL: http://10.1.1.120:8088
```

Runtime deploy evidence:

```text
After PR #425, main was deployed to dev.
Deploy built API, worker, and web images from source.
Deploy smoke passed.
Full host smoke passed.
```

Focused Sprint 12 smoke evidence:

```text
S12-02-01: batch QC transition persisted to inventory.batches and survived API restart.
S12-02-02: available-stock recalculated from persisted batch QC status.
S12-03-02: inbound QC PASS/FAIL/PARTIAL stayed consistent across batch detail, available stock, audit, and API restart.
```

Release smoke at current Sprint 12 main:

```text
Commit: cccecc1f
Command: ./infra/scripts/smoke-dev-full.sh
Result: Full ERP dev smoke passed
```

Full dev smoke included these persisted checks:

```text
persisted_audit_login
persisted_sales_reservation
persisted_sales_order
persisted_stock_adjustment
persisted_stock_movement
persisted_available_stock
persisted_stock_count
persisted_purchase_order
persisted_inbound_qc
persisted_return_receipt
persisted_supplier_rejection
```

After PR #431, `/opt/ERP-v2` was fast-forwarded to main commit `cccecc1f`.

---

## 5. CI And Migration Evidence

GitHub checks:

```text
PR #429: api, e2e, openapi, required-api, required-migration, required-openapi, required-web passed.
PR #430: e2e, required-api, required-migration, required-openapi, required-web passed.
PR #431: e2e, required-api, required-migration, required-openapi, required-web passed.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
Source: /opt/ERP-v2 at main commit cccecc1f
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
Applied migrations: 25
Rolled back migrations: 25
```

Migration notices observed during rollback were idempotent schema/constraint notices from earlier migrations. They did not fail the gate.

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S12-04-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 12:

```text
1. End-of-day reconciliation.
2. Shipping manifest, pick, and pack task package.
3. Finance AR/AP/COD/cash runtime stores.
4. Subcontract runtime stores.
5. Master data catalogs and auth/session hardening.
```

---

## 7. Release Status

Sprint 12 release gate status:

```text
Task PRs: merged through S12-04-01
Current changelog PR: pending merge
Main CI: green through PR #431; rerun required for this changelog PR
Dev runtime smoke: green at commit cccecc1f
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: pending
```

Recommended tag after this changelog merges and main CI is green:

```text
v0.12.0-batch-qc-status-persistence
```

Do not move the tag once pushed. If a post-tag fix is needed, create a new patch tag instead.
