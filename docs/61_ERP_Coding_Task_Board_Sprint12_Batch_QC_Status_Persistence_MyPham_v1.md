# 61_ERP_Coding_Task_Board_Sprint12_Batch_QC_Status_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 12 - Batch/QC status persistence v1
Document role: Coding task board for Sprint 12 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 11 release gate

---

## 1. Sprint 12 Context

Sprint 11 persisted the inventory read model and the main owner documents that hold operational evidence:

```text
available stock read model
sales order owner documents
purchase order owner documents
return receipt owner documents
supplier rejection owner documents
```

The remaining Sprint 11 ledger identifies the next highest warehouse risk:

```text
Batch catalog / QC status
Current constructor: inventoryapp.NewPrototypeBatchCatalog(auditLogStore)
Runtime risk: Batch QC status changes reset even when inbound QC and stock balances persist.
Recommended handling: Add PostgreSQL batch read/write adapter or tie changes to inventory.batches.
```

This matters because available stock now reads persisted warehouse truth, including batch and QC status, while the batch catalog/QC transition path can still update prototype memory only. Sprint 12 closes that mismatch.

---

## 2. Sprint 12 Theme

```text
Persist batch catalog and QC status changes v1
```

Business reason:

```text
Warehouse users must not see a batch as QC PASS, HOLD, FAIL, or QUARANTINE only until the API restarts. Batch QC status controls whether stock can be sold, reserved, picked, or quarantined, so it must be durable and aligned with persisted stock availability.
```

---

## 3. Sprint 12 Goals

By the end of Sprint 12, the system should support:

```text
1. PostgreSQL-backed batch catalog reads from inventory.batches when DB config exists.
2. Batch list/detail/QC transition APIs remain queryable after restart/redeploy.
3. Batch QC status changes update persisted inventory.batches, not only audit or memory.
4. Inbound QC batch-status updates persist through the same durable path.
5. Available-stock and reservation behavior reflects persisted batch QC status.
6. QC transition audit history remains persisted.
7. Dev smoke or focused persistence smoke proves batch QC transition -> inventory.batches -> available-stock read consistency.
8. Release evidence records persisted scope, migration status, CI, deploy, and remaining prototype stores.
```

---

## 4. Sprint 12 Non-Goals

Sprint 12 does not include:

```text
- New master-data batch creation UI.
- Full lot genealogy or manufacturing traceability tree.
- New receiving, returns, shipping, finance, or subcontract workflows.
- Direct stock balance mutation.
- Changing decimal-string API boundaries.
- Changing public API response envelopes unless a task explicitly updates OpenAPI and generated clients.
- Replacing all remaining prototype runtime stores.
- Broad auth/session, master-data, or frontend fallback persistence rewrites.
- Cosmetic refactors unrelated to batch/QC persistence correctness.
```

---

## 5. Branch / PR / Release Rules

Current repo workflow remains:

```text
task branch
-> build/test on dev server
-> PR
-> manual self-review comment
-> manual merge into main
-> sync/deploy dev server when runtime changes require it
```

Do not use GitHub auto review or auto merge.
Do not create a long-lived sprint branch unless the team explicitly changes that policy.

Default task branch pattern:

```text
codex/feature-S12-xx-yy-short-task-name
```

Recommended Sprint 12 release tag after completion:

```text
v0.12.0-batch-qc-status-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for any new migrations.
4. Sprint 12 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

---

## 6. Sprint 12 Demo Script

### Case 1: Manual batch QC transition persists

```text
1. Query a HOLD batch through the batch API.
2. Change the batch QC status to PASS through the QC transition API.
3. Confirm inventory.batches has the updated QC status.
4. Restart or redeploy API.
5. Confirm batch detail and QC transition history still show the persisted status and audit evidence.
```

### Case 2: Failed or quarantined batch blocks availability

```text
1. Change or inspect a batch with QC FAIL, HOLD, QUARANTINE, or RETEST_REQUIRED.
2. Query available stock and reservation-facing reads.
3. Confirm the batch is not treated as sellable, reservable, or pickable.
```

### Case 3: Inbound QC updates persisted batch status

```text
1. Run an inbound QC decision that updates batch QC state.
2. Confirm the batch status changes in inventory.batches.
3. Confirm available stock reads the same persisted batch status.
4. Restart or redeploy API.
5. Confirm inbound QC, batch detail, audit, and available stock remain consistent.
```

---

## 7. Sprint 12 Guardrails

These rules are non-negotiable:

```text
1. Do not write stock balances directly.
2. Batch QC transitions must update inventory.batches in DB mode.
3. Domain QC transition rules must remain the authority for valid state changes.
4. Existing batch API response envelopes must stay stable unless OpenAPI and clients are updated in the same task.
5. Every batch QC status change must write audit evidence.
6. FAIL, HOLD, QUARANTINE, and RETEST_REQUIRED batches must not become sellable, reservable, or pickable.
7. Prototype fallback may remain for no-DB/local mode, but must be explicit.
8. Frontend fallback state must not be counted as backend persistence evidence.
9. Migration apply/rollback is required if schema changes.
10. Behavioral diffs and cosmetic formatting must stay separate.
```

---

## 8. Dependency Map

```text
S12-00-00 Sprint 12 task board
  -> S12-01-01 batch/QC persistence design

S12-01-01 batch/QC persistence design
  -> S12-01-02 batch catalog runtime interface and selector
  -> S12-01-03 PostgreSQL batch catalog store

S12-01-03 PostgreSQL batch catalog store
  -> S12-01-04 batch catalog store tests
  -> S12-02-01 batch QC transition persistence smoke

S12-02-01 batch QC transition persistence smoke
  -> S12-02-02 available-stock QC consistency smoke
  -> S12-03-01 inbound QC batch-status persistence integration

S12-03-01 inbound QC batch-status persistence integration
  -> S12-03-02 inbound QC batch-status smoke

S12-03-02 inbound QC batch-status smoke
  -> S12-04-01 remaining prototype ledger update
  -> S12-05-01 Sprint 12 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S12-00-00 | Sprint 12 task board | File 61 created, reviewed, merged to main | `docs/60_ERP_Sprint11_Changelog_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1.md` |
| S12-01-01 | Batch/QC persistence design | Map current `BatchCatalog` contract to `inventory.batches`, audit, available-stock, reservation, and inbound QC consumers | `docs/qa/S11-05-01_remaining_prototype_store_ledger.md` |
| S12-01-02 | Batch catalog runtime interface and selector | Batch handlers/services can use DB-backed catalog when DB config exists, with prototype fallback for no-DB/local mode | `apps/api/internal/modules/inventory/application/batch_catalog.go` |
| S12-01-03 | PostgreSQL batch catalog store | List/Get/ChangeQCStatus read and write `inventory.batches`; QC transition audit remains persisted | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S12-01-04 | Batch catalog store tests | Tests cover filters, detail lookup, valid/invalid QC transitions, audit write, and runtime fallback selection | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S12-02-01 | Batch QC transition persistence smoke | Dev or focused smoke proves POST QC transition updates `inventory.batches` and survives restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S12-02-02 | Available-stock QC consistency smoke | Available stock and reservation-facing reads respect persisted batch QC status after transition | `docs/qa/S11-05-01_remaining_prototype_store_ledger.md` |
| S12-03-01 | Inbound QC batch-status persistence integration | Inbound QC status updater writes persisted batch status in DB mode, not prototype-only memory | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S12-03-02 | Inbound QC batch-status smoke | Inbound QC PASS/FAIL/PARTIAL status effects remain consistent across batch detail, available stock, and audit after restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S12-04-01 | Remaining prototype ledger update | Sprint 11 ledger is updated or superseded after Sprint 12 persistence work | `docs/qa/S11-05-01_remaining_prototype_store_ledger.md` |
| S12-05-01 | Sprint 12 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/60_ERP_Sprint11_Changelog_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1.md` |

---

## 10. Verification Gates

Backend checks:

```text
go test ./...
go vet ./...
```

Migration checks when migrations change:

```text
Apply all migrations on PostgreSQL 16.
Roll back all migrations on PostgreSQL 16.
```

API checks when public API shape changes:

```text
Validate OpenAPI.
Regenerate clients if the OpenAPI contract changes.
```

Frontend checks when frontend code changes:

```text
web lint
web test
```

Release gate:

```text
Dev deploy or repo sync evidence for merged main.
Dev release gate smoke.
GitHub required checks green.
```

---

## 11. Definition Of Done

Sprint 12 is done only when:

```text
1. All Sprint 12 task PRs are merged through manual review flow.
2. Runtime batch catalog is PostgreSQL-backed when DB config exists.
3. Batch list/detail/QC transition APIs query persisted batch state.
4. Batch QC status changes update inventory.batches and persisted audit evidence.
5. Inbound QC batch-status updater uses the persisted path in DB mode.
6. Available-stock and reservation-facing reads reflect persisted batch QC status.
7. PostgreSQL migration apply/rollback evidence is recorded for any schema change.
8. Dev smoke or focused persistence smoke records restart/redeploy evidence.
9. Sprint 12 changelog records persisted scope, remaining prototype stores, CI, deploy, and tag status.
```
