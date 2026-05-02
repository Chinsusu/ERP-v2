# 63_ERP_Coding_Task_Board_Sprint13_End_of_Day_Reconciliation_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 13 - End-of-day reconciliation persistence v1
Document role: Coding task board for Sprint 13 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 12 release gate

---

## 1. Sprint 13 Context

Sprint 12 persisted the batch catalog and QC transition path. The remaining prototype ledger identifies the next highest warehouse risk:

```text
End-of-day reconciliation
Current runtime path: prototype in-memory reconciliation store
Runtime risk: shift close and variance evidence can reset on API restart or redeploy.
Recommended handling: Add PostgreSQL-backed end-of-day reconciliation store and prove close evidence survives restart.
```

This matters because end-of-day reconciliation is the warehouse control point for closing a shift, recording variance exceptions, and confirming that operational evidence matches persisted warehouse truth.

---

## 2. Sprint 13 Theme

```text
Persist end-of-day reconciliation and shift close evidence v1
```

Business reason:

```text
Warehouse leads must not lose close status, exception notes, checklist evidence, or variance decisions after the API restarts. End-of-day reconciliation should be durable because it is the daily control record for stock, receiving, returns, and fulfillment discrepancies.
```

---

## 3. Sprint 13 Goals

By the end of Sprint 13, the system should support:

```text
1. PostgreSQL-backed end-of-day reconciliation store when DB config exists.
2. Reconciliation list/detail/close behavior survives API restart or redeploy.
3. Close writes persisted status, closed_at, closed_by, exception note, checklist, and line evidence.
4. Domain close rules remain unchanged and continue to block unsafe closes.
5. Close actions write persisted audit evidence.
6. Existing public API response envelopes remain stable.
7. Prototype fallback remains explicit for no-DB/local mode.
8. Focused dev smoke proves close evidence survives restart.
9. Release evidence records persisted scope, migration status, CI, deploy, and remaining prototype stores.
```

---

## 4. Sprint 13 Non-Goals

Sprint 13 does not include:

```text
- New end-of-day frontend screens.
- New daily board widgets beyond existing API behavior.
- Changing stock movement, available-stock, receiving, returns, shipping, finance, or subcontract flows.
- Direct stock balance mutation.
- Changing decimal-string API boundaries.
- Changing public API response envelopes unless a task explicitly updates OpenAPI and generated clients.
- Replacing all remaining prototype runtime stores.
- Broad auth/session, master-data, or frontend fallback persistence rewrites.
- Cosmetic refactors unrelated to reconciliation persistence correctness.
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
codex/feature-S13-xx-yy-short-task-name
```

Recommended Sprint 13 release tag after completion:

```text
v0.13.0-end-of-day-reconciliation-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for any new migrations.
4. Sprint 13 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

---

## 6. Sprint 13 Demo Script

### Case 1: Clean close persists

```text
1. Query today's end-of-day reconciliation.
2. Close the reconciliation with all required checks satisfied.
3. Confirm the persisted store has CLOSED status, closed_at, closed_by, checklist, and audit evidence.
4. Restart or redeploy API.
5. Confirm the reconciliation still shows CLOSED with the same evidence.
```

### Case 2: Unsafe close remains blocked

```text
1. Query a reconciliation with unresolved blocking issues.
2. Attempt to close it without an allowed exception.
3. Confirm the API rejects the close and does not persist CLOSED status.
4. Confirm audit evidence records the rejected attempt only if current audit behavior already supports it.
```

### Case 3: Variance exception close persists

```text
1. Query a reconciliation with variance lines that allow exception close.
2. Close it with the required exception note.
3. Confirm the variance decision and exception note persist.
4. Restart or redeploy API.
5. Confirm the same close evidence is returned.
```

---

## 7. Sprint 13 Guardrails

These rules are non-negotiable:

```text
1. Do not write stock balances directly.
2. End-of-day close rules must remain owned by the domain/application service.
3. A restart or redeploy must not erase close status or exception evidence in DB mode.
4. Existing reconciliation API response envelopes must stay stable unless OpenAPI and clients are updated in the same task.
5. Every successful close must write audit evidence.
6. Unresolved blocking issues must not become closeable by persistence adapter behavior.
7. Prototype fallback may remain for no-DB/local mode, but must be explicit.
8. Frontend fallback state must not be counted as backend persistence evidence.
9. Migration apply/rollback is required if schema changes.
10. Behavioral diffs and cosmetic formatting must stay separate.
```

---

## 8. Dependency Map

```text
S13-00-00 Sprint 13 task board
  -> S13-01-01 end-of-day reconciliation persistence design

S13-01-01 end-of-day reconciliation persistence design
  -> S13-01-02 runtime selector and store adapter
  -> S13-01-03 PostgreSQL reconciliation schema/store

S13-01-03 PostgreSQL reconciliation schema/store
  -> S13-01-04 store and handler tests
  -> S13-02-01 close persistence smoke

S13-02-01 close persistence smoke
  -> S13-03-01 remaining prototype ledger update
  -> S13-04-01 Sprint 13 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S13-00-00 | Sprint 13 task board | File 63 created, reviewed, merged to main | `docs/62_ERP_Sprint12_Changelog_Batch_QC_Status_Persistence_MyPham_v1.md` |
| S13-01-01 | End-of-day reconciliation persistence design | Map current reconciliation domain/store contract to PostgreSQL tables, audit evidence, selectors, fallback behavior, and restart smoke | `docs/qa/S12-04-01_remaining_prototype_store_ledger.md` |
| S13-01-02 | Runtime selector and store adapter | API runtime uses DB-backed reconciliation store when DB config exists, with explicit prototype fallback for no-DB/local mode | `apps/api/internal/modules/inventory/application/end_of_day_reconciliation.go` |
| S13-01-03 | PostgreSQL reconciliation schema/store | List/Get/Save read and write durable reconciliation, checklist, line, close, and exception evidence | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S13-01-04 | Store and handler tests | Tests cover list/get/save, close persistence, blocked close, exception close, audit, and runtime fallback selection | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S13-02-01 | Close persistence smoke | Dev or focused smoke proves POST close updates PostgreSQL state and survives restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S13-03-01 | Remaining prototype ledger update | Sprint 12 ledger is updated or superseded after Sprint 13 persistence work | `docs/qa/S12-04-01_remaining_prototype_store_ledger.md` |
| S13-04-01 | Sprint 13 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/62_ERP_Sprint12_Changelog_Batch_QC_Status_Persistence_MyPham_v1.md` |

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

