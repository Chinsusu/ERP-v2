# 65_ERP_Coding_Task_Board_Sprint14_Shipping_Pick_Pack_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 14 - Shipping manifest, pick, and pack persistence v1
Document role: Coding task board for Sprint 14 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 13 release gate

---

## 1. Sprint 14 Context

Sprint 13 persisted end-of-day reconciliation and updated the remaining prototype ledger. The next highest warehouse execution risk is the shipping package:

```text
Carrier manifests -> prototype memory
Pick tasks         -> prototype memory
Pack tasks         -> prototype memory
```

This matters because sales orders, reservations, stock movements, return receipts, and daily close evidence are now durable, while the shipping execution steps between reservation and carrier handover can still reset after API restart.

---

## 2. Sprint 14 Theme

```text
Persist shipping manifest, pick, and pack execution evidence v1
```

Business reason:

```text
Warehouse users must not lose pick progress, pack confirmation, carrier manifest membership, scan exceptions, or handover evidence after API restart. Shipping execution is the operational bridge between reserved stock and completed carrier handover.
```

---

## 3. Sprint 14 Goals

By the end of Sprint 14, the system should support:

```text
1. PostgreSQL-backed carrier manifest store when DB config exists.
2. PostgreSQL-backed pick task store when DB config exists.
3. PostgreSQL-backed pack task store when DB config exists.
4. Manifest list/create/add/remove/ready/exception/handover/scan evidence survives restart or redeploy.
5. Pick task start/confirm-line/complete/exception evidence survives restart or redeploy.
6. Pack task start/confirm/exception evidence survives restart or redeploy.
7. Existing public API response envelopes remain stable.
8. Existing sales order handover and pack adapters keep their current behavior.
9. Prototype fallback remains explicit for no-DB/local mode.
10. Dev smoke proves shipping execution state survives restart.
11. Release evidence records persisted scope, migration status, CI, deploy, and remaining prototype stores.
```

---

## 4. Sprint 14 Non-Goals

Sprint 14 does not include:

```text
- New shipping frontend screens.
- New carrier integration API.
- New route planning, rate shopping, or label purchase workflow.
- Changing sales order status rules unless required by existing shipping adapters.
- Changing reservation, stock movement, available-stock, receiving, returns, finance, or subcontract flows.
- Direct stock balance mutation.
- Changing decimal-string API boundaries.
- Changing public API response envelopes unless OpenAPI and clients are updated in the same task.
- Replacing all remaining prototype runtime stores.
- Broad auth/session, master-data, or frontend fallback persistence rewrites.
- Cosmetic refactors unrelated to shipping persistence correctness.
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
codex/feature-S14-xx-yy-short-task-name
```

Recommended Sprint 14 release tag after completion:

```text
v0.14.0-shipping-pick-pack-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for any new migrations.
4. Sprint 14 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

---

## 6. Sprint 14 Demo Script

### Case 1: Manifest handover persists

```text
1. Create a carrier manifest.
2. Add packed orders to the manifest.
3. Mark it ready to scan.
4. Confirm carrier handover.
5. Restart or redeploy API.
6. Confirm manifest status, orders, scan status, and handover evidence are still present.
```

### Case 2: Pick progress persists

```text
1. Start a pick task.
2. Confirm one or more pick lines.
3. Record an exception when needed.
4. Restart or redeploy API.
5. Confirm task status, line quantities, picker, timestamps, and exception evidence are still present.
```

### Case 3: Pack confirmation persists

```text
1. Start a pack task.
2. Confirm package count, weight, carrier service, and packed-by evidence.
3. Restart or redeploy API.
4. Confirm pack task status and evidence are still present.
```

---

## 7. Sprint 14 Guardrails

These rules are non-negotiable:

```text
1. Do not write stock balances directly.
2. Shipping persistence must not bypass sales order/reservation services.
3. Manifest handover must keep existing sales order handover adapter behavior.
4. Pack confirmation must keep existing sales order pack adapter behavior.
5. Existing shipping API response envelopes must stay stable unless OpenAPI and clients are updated in the same task.
6. Every successful state transition must keep existing audit evidence.
7. Restart or redeploy must not erase manifest, pick, or pack state in DB mode.
8. Prototype fallback may remain for no-DB/local mode, but must be explicit.
9. Frontend fallback state must not be counted as backend persistence evidence.
10. Migration apply/rollback is required if schema changes.
11. Behavioral diffs and cosmetic formatting must stay separate.
```

---

## 8. Dependency Map

```text
S14-00-00 Sprint 14 task board
  -> S14-01-01 shipping persistence design

S14-01-01 shipping persistence design
  -> S14-01-02 runtime contracts and selectors
  -> S14-01-03 carrier manifest PostgreSQL store

S14-01-03 carrier manifest PostgreSQL store
  -> S14-01-04 pick task PostgreSQL store
  -> S14-02-01 manifest handover/scan persistence smoke

S14-01-04 pick task PostgreSQL store
  -> S14-01-05 pack task PostgreSQL store
  -> S14-02-02 pick/pack persistence smoke

S14-01-05 pack task PostgreSQL store
  -> S14-01-06 store and handler tests
  -> S14-03-01 shipping package integration check

S14-03-01 shipping package integration check
  -> S14-04-01 remaining prototype ledger update
  -> S14-05-01 Sprint 14 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S14-00-00 | Sprint 14 task board | File 65 created, reviewed, merged to main | `docs/64_ERP_Sprint13_Changelog_End_of_Day_Reconciliation_Persistence_MyPham_v1.md` |
| S14-01-01 | Shipping persistence design | Map carrier manifest, pick task, and pack task contracts to PostgreSQL tables, audit evidence, selectors, fallback behavior, and restart smoke | `docs/qa/S13-03-01_remaining_prototype_store_ledger.md` |
| S14-01-02 | Runtime contracts and selectors | Shipping handlers/services can use DB-backed stores when DB config exists, with explicit prototype fallback for no-DB/local mode | `apps/api/internal/modules/shipping/application` |
| S14-01-03 | Carrier manifest PostgreSQL store | Manifest list/create/add/remove/ready/exception/handover/scan read and write durable state | `apps/api/internal/modules/shipping/application/carrier_manifest.go` |
| S14-01-04 | Pick task PostgreSQL store | Pick list/get/start/confirm-line/complete/exception read and write durable state | `apps/api/internal/modules/shipping/application/pick_task.go` |
| S14-01-05 | Pack task PostgreSQL store | Pack list/get/start/confirm/exception read and write durable state | `apps/api/internal/modules/shipping/application/pack_task.go` |
| S14-01-06 | Store and handler tests | Tests cover filters, lifecycle transitions, audit, restart-like reload through fresh store instances, and runtime fallback selection | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S14-02-01 | Manifest handover/scan persistence smoke | Dev or focused smoke proves manifest handover/scan state survives restart/redeploy | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S14-02-02 | Pick/pack persistence smoke | Dev or focused smoke proves pick and pack task progress survives restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S14-03-01 | Shipping package integration check | Existing sales order handover and pack adapters still update sales order state while shipping evidence persists | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S14-04-01 | Remaining prototype ledger update | Sprint 13 ledger is updated or superseded after Sprint 14 persistence work | `docs/qa/S13-03-01_remaining_prototype_store_ledger.md` |
| S14-05-01 | Sprint 14 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/64_ERP_Sprint13_Changelog_End_of_Day_Reconciliation_Persistence_MyPham_v1.md` |

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
