# 59_ERP_Coding_Task_Board_Sprint11_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Document role: Coding task board for Sprint 11 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 10 dev release gate

---

## 1. Sprint 11 Context

Sprint 10 made the highest restart-risk runtime stores PostgreSQL-backed when DB config exists:

```text
stock movement writer
audit log
sales order reservations
stock adjustments
stock counts
warehouse receiving
inbound QC
```

The next risk is cross-store mismatch:

```text
persisted evidence and stock balances
-> prototype read model or owner document resets
-> UI/report/API can show stale or seed-only operational state after restart
```

Sprint 11 starts with the highest P0 item from `docs/qa/S10-05-01_remaining_prototype_store_ledger.md`: available stock reads must use persisted `inventory.stock_balances` instead of prototype snapshots when DB config exists.

---

## 2. Sprint 11 Theme

```text
Persist inventory read model and owner documents v1
```

Business reason:

```text
The ERP can already write persisted stock movement evidence. The next correctness bar is that users read the same persisted operational truth after restart, especially inventory availability and the documents that own persisted evidence.
```

---

## 3. Sprint 11 Goals

By the end of Sprint 11, the system should support:

```text
1. PostgreSQL-backed available-stock read model from inventory.stock_balances.
2. Available-stock API/reporting paths that reflect persisted stock movement writes after restart/redeploy.
3. Sales order owner documents persisted or explicitly deferred with evidence and reason.
4. Purchase order owner documents persisted or explicitly deferred with evidence and reason.
5. Return receipt and supplier rejection owner documents persisted or explicitly split.
6. Dev smoke or focused persistence smoke for each promoted read/document store.
7. Release evidence separating persisted stores, remaining prototype stores, CI, deploy, and tag status.
```

---

## 4. Sprint 11 Non-Goals

Sprint 11 does not include:

```text
- New warehouse, sales, purchase, returns, finance, or subcontract workflows.
- Full master-data persistence for every catalog.
- Full auth/session rewrite or SSO.
- Replacing all frontend prototype fallback services.
- Direct stock balance mutation.
- Changing decimal-string API boundaries.
- Changing public API response envelopes unless a task explicitly updates OpenAPI and generated clients.
- Cosmetic refactors unrelated to persistence correctness.
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
codex/feature-S11-xx-yy-short-task-name
```

Recommended Sprint 11 release tag after completion:

```text
v0.11.0-persist-inventory-read-model-owner-documents
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for any new migrations.
4. Sprint 11 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

---

## 6. Sprint 11 Demo Script

### Case 1: Available stock reflects persisted balance

```text
1. Create/post a stock movement through an existing workflow.
2. Confirm inventory.stock_balances changes in PostgreSQL.
3. Query available stock API/reporting.
4. Restart or redeploy API.
5. Confirm available stock API/reporting still reflects PostgreSQL stock_balances.
```

### Case 2: Reservation remains consistent with sales order document

```text
1. Create and confirm a sales order.
2. Confirm reservation rows persist.
3. Restart or redeploy API.
4. Confirm sales order document and reservation state remain queryable and consistent.
```

### Case 3: Inbound owner document remains consistent

```text
1. Create/approve a purchase order.
2. Receive goods and run inbound QC.
3. Restart or redeploy API.
4. Confirm PO, receiving, QC, and stock availability remain traceable.
```

### Case 4: Return/rejection owner document remains consistent

```text
1. Receive a return or supplier rejection.
2. Move it through inspection/disposition.
3. Restart or redeploy API.
4. Confirm owner document, audit, and stock evidence remain traceable.
```

---

## 7. Sprint 11 Guardrails

These rules are non-negotiable:

```text
1. Do not write stock balances directly.
2. Available-stock reads may read inventory.stock_balances but must not mutate it.
3. All stock changes must continue through the stock movement service.
4. Read-model persistence must preserve existing API response envelopes.
5. Money, quantity, rate, and UOM values must keep file 40 decimal-string rules.
6. Do not weaken auth, permission, or audit checks to make persistence easier.
7. Do not commit secrets or real environment files.
8. Prototype fallback may remain for local/no-DB mode, but must be explicit.
9. Frontend fallback state must not be counted as backend persistence evidence.
10. Behavioral diffs and cosmetic formatting must stay separate.
```

---

## 8. Dependency Map

```text
S11-00-00 Sprint 11 task board
  -> S11-01-01 available-stock read model design

S11-01-01 available-stock read model design
  -> S11-01-02 available-stock PostgreSQL store
  -> S11-01-03 available-stock persistence smoke

S11-01-03 available-stock persistence smoke
  -> S11-02-01 sales order document persistence design

S11-02-01 sales order document persistence design
  -> S11-02-02 sales order PostgreSQL store
  -> S11-02-03 sales order document persistence smoke

S11-02-03 sales order document persistence smoke
  -> S11-03-01 purchase order document persistence design

S11-03-01 purchase order document persistence design
  -> S11-03-02 purchase order PostgreSQL store
  -> S11-03-03 purchase order document persistence smoke

S11-03-03 purchase order document persistence smoke
  -> S11-04-01 return and supplier rejection persistence design

S11-04-01 return and supplier rejection persistence design
  -> S11-04-02 return receipt PostgreSQL store
  -> S11-04-03 supplier rejection PostgreSQL store
  -> S11-04-04 return/rejection persistence smoke

S11-04-04 return/rejection persistence smoke
  -> S11-05-01 remaining prototype store ledger update
  -> S11-06-01 Sprint 11 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S11-00-00 | Sprint 11 task board | File 59 created, reviewed, merged to main | `58_ERP_Sprint10_Changelog_Persist_Operational_Runtime_Stores_MyPham_v1.md` |
| S11-01-01 | Available-stock read model design | Map current `StockAvailabilityStore` contract to PostgreSQL `inventory.stock_balances` joins without changing API envelopes | `docs/qa/S10-05-01_remaining_prototype_store_ledger.md` |
| S11-01-02 | Available-stock PostgreSQL store | Runtime available-stock reads use PostgreSQL when DB config exists, with prototype fallback for no-DB mode | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S11-01-03 | Available-stock persistence smoke | Dev smoke proves API/reporting reads reflect persisted `inventory.stock_balances` after stock movement evidence | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S11-02-01 | Sales order document persistence design | Sales order header/line/status lifecycle is mapped against existing reservation persistence | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S11-02-02 | Sales order PostgreSQL store | Sales order owner documents persist through create/confirm/cancel and stay consistent with reservations | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S11-02-03 | Sales order document persistence smoke | Sales order document and reservation remain queryable after restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S11-03-01 | Purchase order document persistence design | PO header/line/status lifecycle is mapped against receiving/inbound QC persistence | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S11-03-02 | Purchase order PostgreSQL store | PO owner documents persist through create/submit/approve/close/cancel paths | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S11-03-03 | Purchase order document persistence smoke | PO, receiving, and inbound QC remain traceable after restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S11-04-01 | Return and supplier rejection persistence design | Return receipt and supplier rejection owner document contracts are mapped before implementation | `43_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S11-04-02 | Return receipt PostgreSQL store | Return inspection/disposition state persists and stays traceable to stock/audit evidence | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S11-04-03 | Supplier rejection PostgreSQL store | Supplier rejection state persists and stays traceable to failed inbound QC evidence | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S11-04-04 | Return/rejection persistence smoke | Return and supplier rejection owner docs remain queryable after restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S11-05-01 | Remaining prototype store ledger update | S10 ledger is updated or superseded after Sprint 11 persistence work | `docs/qa/S10-05-01_remaining_prototype_store_ledger.md` |
| S11-06-01 | Sprint 11 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `58_ERP_Sprint10_Changelog_Persist_Operational_Runtime_Stores_MyPham_v1.md` |

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
Confirm no data-destroying down migration is added casually.
```

OpenAPI checks when API shapes change:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
```

Frontend checks when UI/service behavior changes:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
```

Release gate checks:

```text
./infra/scripts/dev-release-gate.sh dev
GitHub required-ci green on main
```

---

## 11. Definition of Done

Sprint 11 is complete when:

```text
1. All Sprint 11 task PRs are merged to main through manual review/merge.
2. Available-stock runtime reads are PostgreSQL-backed when DB config is present.
3. Available-stock API/reporting reflects persisted stock_balances after restart/redeploy.
4. Sales order, purchase order, return, and supplier rejection owner document stores are persisted or explicitly deferred with evidence and reason.
5. Every new table has migration apply/rollback evidence.
6. Dev smoke or focused persistence smoke proves promoted stores survive runtime restart/redeploy paths.
7. Remaining prototype stores are re-ranked after Sprint 11.
8. Dev release gate passes on main.
9. Sprint 11 changelog records persisted scope, remaining risk, CI, deploy, and tag status.
```
