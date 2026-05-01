# 57_ERP_Coding_Task_Board_Sprint10_Persist_Operational_Runtime_Stores_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 10 - Persist operational runtime stores v1
Document role: Coding task board for Sprint 10 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 9 dev release gate

---

## 1. Sprint 10 Context

Sprint 9 completed the release-gate foundation:

```text
dev disk preflight
-> full dev smoke
-> deploy evidence
-> dev release gate
-> persisted stock movement runtime path
```

The highest-risk stock movement writer is now PostgreSQL-backed when database configuration is present, and the dev smoke verifies a persisted `inventory.stock_ledger` row.

The remaining risk is that many operational document stores still reset on process restart:

```text
audit logs
sales order reservations
stock adjustments
stock counts
warehouse receiving
inbound QC
returns and supplier rejections
finance runtime documents
shipping and subcontract runtime documents
```

Sprint 10 reduces this risk in priority order, without changing the user-facing workflow scope.

---

## 2. Sprint 10 Theme

```text
Persist operational runtime stores v1
```

Business reason:

```text
The ERP already has order fulfillment, inbound, returns, subcontract, finance, and reporting workflows.
The next correctness risk is not another screen; it is losing operational state after API restart or redeploy.
```

Sprint 10 should make the most important runtime evidence and document state survive restarts.

---

## 3. Sprint 10 Goals

By the end of Sprint 10, the system should support:

```text
1. PostgreSQL-backed audit log reads/writes for runtime action evidence.
2. PostgreSQL-backed sales order reservation state or a documented adapter to existing reservation tables.
3. PostgreSQL-backed stock adjustment lifecycle state.
4. PostgreSQL-backed stock count lifecycle state.
5. PostgreSQL-backed warehouse receiving and inbound QC state, or a documented split if one side must come first.
6. Repeatable restart/redeploy persistence smoke checks for each persisted store.
7. Release evidence that separates persisted stores, remaining prototype stores, and verification status.
```

---

## 4. Sprint 10 Non-Goals

Sprint 10 does not include:

```text
- New business modules.
- Full ERP accounting general ledger.
- Full auth/session rewrite or SSO.
- Replacing all frontend prototype fallback services.
- Migrating every master-data catalog.
- Rewriting completed workflows for cosmetic cleanup.
- Direct stock balance mutation.
- Breaking existing OpenAPI contracts unless a task explicitly updates the contract and generated clients.
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
codex/feature-S10-xx-yy-short-task-name
```

Recommended Sprint 10 release tag after completion:

```text
v0.10.0-persist-operational-runtime-stores
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green.
4. Sprint 10 changelog records persisted and remaining prototype stores.
```

---

## 6. Sprint 10 Demo Script

### Case 1: Audit evidence survives restart

```text
1. Login and perform a permissioned action that writes audit evidence.
2. Confirm the audit log query returns the event.
3. Restart or redeploy API.
4. Confirm the same audit event is still queryable.
```

### Case 2: Reservation state survives restart

```text
1. Create or confirm an order that reserves stock.
2. Confirm available/reserved quantities are correct.
3. Restart or redeploy API.
4. Confirm reservation state remains correct and reports still reflect it.
```

### Case 3: Inventory documents survive restart

```text
1. Create a stock adjustment or stock count.
2. Move it through submit/approve/post or submit/reconcile.
3. Restart or redeploy API.
4. Confirm lifecycle state, audit evidence, and stock movement evidence remain correct.
```

### Case 4: Inbound QC state survives restart

```text
1. Receive goods and create inbound QC inspection.
2. Mark pass/fail/hold/partial.
3. Restart or redeploy API.
4. Confirm receiving, QC decision, quarantine/available behavior, and board/report signals remain correct.
```

---

## 7. Sprint 10 Guardrails

These rules are non-negotiable:

```text
1. Do not write stock balances directly.
2. All stock changes must continue through the stock movement service.
3. Persistence work must preserve existing API response envelopes.
4. Money, quantity, rate, and UOM values must keep file 40 decimal-string rules.
5. Do not weaken auth, permission, or audit checks to make persistence easier.
6. Do not commit secrets or real environment files.
7. Every persistent store must have migration apply/rollback coverage.
8. Every persisted store must have at least one restart/redeploy or DB-backed smoke path before being called complete.
9. Prototype fallback may remain for local/no-DB mode, but must be explicit.
10. Frontend fallback state must not be counted as backend persistence evidence.
11. Behavioral diffs and cosmetic formatting must stay separate.
```

---

## 8. Dependency Map

```text
S10-00-00 Sprint 10 task board
  -> S10-01-01 audit persistence design

S10-01-01 audit persistence design
  -> S10-01-02 audit PostgreSQL migration/store
  -> S10-01-03 audit persistence smoke

S10-01-03 audit persistence smoke
  -> S10-02-01 reservation persistence design

S10-02-01 reservation persistence design
  -> S10-02-02 reservation PostgreSQL store
  -> S10-02-03 reservation persistence smoke

S10-02-03 reservation persistence smoke
  -> S10-03-01 stock adjustment persistence
  -> S10-03-02 stock count persistence

S10-03-01 stock adjustment persistence
  -> S10-03-03 inventory document persistence smoke

S10-03-02 stock count persistence
  -> S10-03-03 inventory document persistence smoke

S10-03-03 inventory document persistence smoke
  -> S10-04-01 receiving persistence
  -> S10-04-02 inbound QC persistence

S10-04-01 receiving persistence
  -> S10-04-03 inbound persistence smoke

S10-04-02 inbound QC persistence
  -> S10-04-03 inbound persistence smoke

S10-05-01 remaining prototype store ledger
  -> S10-06-01 Sprint 10 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S10-00-00 | Sprint 10 task board | File 57 created, reviewed, merged to main | `56_ERP_Sprint9_Changelog_System_Hardening_Production_Readiness_Core_MyPham_v1.md` |
| S10-01-01 | Audit persistence design | Existing audit writers/readers are mapped to a PostgreSQL schema and store interface without changing audit semantics | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S10-01-02 | Audit PostgreSQL migration/store | Audit log writes and reads use PostgreSQL when DB config is present, with prototype fallback for no-DB mode | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S10-01-03 | Audit persistence smoke | Dev smoke or focused script proves audit event survives restart/redeploy or is queryable from PostgreSQL | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S10-02-01 | Reservation persistence design | Sales order reservation state model and existing stock availability contract are mapped before implementation | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S10-02-02 | Reservation PostgreSQL store | Reserved stock state persists through restart without direct balance mutation | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S10-02-03 | Reservation persistence smoke | Order reservation survives restart/redeploy and availability/reporting views still return correct reserved quantity | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S10-03-01 | Stock adjustment persistence | Stock adjustment header/line lifecycle state persists through submit/approve/post | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S10-03-02 | Stock count persistence | Stock count session/line/reconciliation state persists through submit/review/adjustment flow | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S10-03-03 | Inventory document persistence smoke | Stock adjustment/count smoke proves document state plus stock movement evidence survives runtime restart path | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S10-04-01 | Warehouse receiving persistence | Goods receipt/receiving state persists through submit/inspect-ready/post | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S10-04-02 | Inbound QC persistence | Inbound QC inspection decisions persist through pass/fail/hold/partial states | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S10-04-03 | Inbound persistence smoke | Receive -> QC decision -> stock/quarantine signal remains correct after restart/redeploy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S10-05-01 | Remaining prototype store ledger | Remaining runtime stores are re-ranked after Sprint 10 persistence work | `docs/qa/S9-03-01_prototype_store_inventory.md` |
| S10-06-01 | Sprint 10 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `56_ERP_Sprint9_Changelog_System_Hardening_Production_Readiness_Core_MyPham_v1.md` |

---

## 10. Verification Gates

Each implementation PR should run the smallest relevant checks plus broader checks when migrations, shared stores, scripts, or API contracts change.

Backend checks:

```text
go test ./...
go vet ./...
```

Migration checks:

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

Sprint 10 is complete when:

```text
1. All task PRs are merged to main through manual review/merge.
2. Audit log runtime state is PostgreSQL-backed when DB config is present.
3. Sales order reservation state is persisted or explicitly deferred with evidence and reason.
4. Stock adjustment and stock count state are persisted or explicitly split into a follow-up with evidence.
5. Receiving and inbound QC state are persisted or explicitly split into a follow-up with evidence.
6. Every new table has migration apply/rollback evidence.
7. Dev smoke or focused persistence smoke proves the persisted stores survive runtime restart/redeploy paths.
8. Remaining prototype stores are re-ranked.
9. Dev release gate passes on main.
10. Sprint 10 changelog records persisted scope, remaining risk, CI, deploy, and tag status.
```
