# 69_ERP_Coding_Task_Board_Sprint16_Subcontract_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 16 - Subcontract runtime store persistence v1
Document role: Coding task board for Sprint 16 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 15 release gate

---

## 1. Sprint 16 Context

Sprint 15 persisted Finance Lite runtime state and updated the remaining prototype store ledger.

The highest remaining production persistence risk is now subcontract runtime state:

```text
Subcontract orders                  -> prototype memory
Subcontract material transfers      -> prototype memory
Subcontract sample approvals        -> prototype memory
Subcontract finished goods receipts -> prototype memory
Subcontract factory claims          -> prototype memory
Subcontract payment milestones      -> prototype memory
Warehouse daily board subcontract signals read those prototype-backed stores
```

This matters because stock movements, finance payables, purchase orders, inbound QC, shipping, returns, and daily close evidence are now durable while subcontract lifecycle evidence can still reset after API restart.

---

## 2. Sprint 16 Theme

```text
Subcontract Runtime Store Persistence
```

Business reason:

```text
The real production flow depends on outside factories: order confirmation, deposit, issue materials/packaging, sample approval, mass production, receive finished goods, inspect quality, claim factory defects, and release final payment.
Those steps must survive deploys and restarts because they affect stock, finance, factory accountability, and warehouse daily operations.
```

---

## 3. Sprint 16 Goals

By the end of Sprint 16, DB-mode runtime must support:

```text
1. Subcontract order state persisted with material lines, status transitions, actor refs, quantity fields, amount fields, and audit.
2. Material transfer state persisted with issued material/packaging lines, source refs, lot/batch evidence, and stock movement links.
3. Sample approval state persisted with submit/approve/reject evidence and reason.
4. Finished goods receipt state persisted with received/accepted/rejected quantities and stock movement links.
5. Factory claim state persisted with defect reason, evidence refs, claim window, and status.
6. Payment milestone state persisted with deposit/final payment readiness and supplier payable link evidence.
7. Warehouse daily board subcontract signals read the same durable subcontract stores.
8. Prototype fallback remains explicit for no-DB/local mode.
9. Dev smoke proves subcontract runtime state survives restart/redeploy.
10. Remaining prototype ledger and release evidence are updated after subcontract persistence.
```

Important runtime rule:

```text
Do not wire only one subcontract store into DB mode while the other subcontract stores remain memory-backed.
Implement and test stores separately, but switch runtime selection to DB mode as one subcontract package after order, transfer, sample, receipt, claim, and milestone stores are all present and tested.
```

---

## 4. Sprint 16 Non-Goals

Sprint 16 does not include:

```text
- Factory portal or supplier-facing login.
- Factory capacity planning.
- Production scheduling optimization.
- Full manufacturing cost accounting or COGS allocation.
- Bank payment integration.
- New subcontract frontend screens beyond smoke/test adjustments if needed.
- Master data catalog persistence.
- Auth/session hardening.
- Broad warehouse daily board redesign.
```

---

## 5. Branch / PR / Release Rules

Current repo workflow remains:

```text
task branch
-> build/test on dev server when runtime changes require it
-> PR
-> manual self-review comment
-> manual merge into main
-> sync/deploy dev server when runtime changes require it
```

Do not use GitHub auto review or auto merge.
Do not create a long-lived sprint branch unless the team explicitly changes that policy.

Default task branch pattern:

```text
codex/feature-S16-xx-yy-short-task-name
```

Recommended Sprint 16 release tag after completion:

```text
v0.16.0-subcontract-runtime-store-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for new migrations.
4. Sprint 16 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

---

## 6. Sprint 16 Demo Script

### Case 1: Outside factory happy path survives restart

```text
1. Create subcontract order with finished item, factory, material/packaging lines, quantity, UOM, estimated cost, and deposit.
2. Submit and approve subcontract order.
3. Confirm factory and record deposit.
4. Issue materials/packaging to factory.
5. Submit and approve sample.
6. Start mass production.
7. Receive finished goods.
8. Accept finished goods.
9. Mark final payment ready and create supplier payable evidence.
10. Restart/redeploy API.
11. Confirm order, transfer, sample, receipt, payment milestone, stock movement, payable, audit, and daily board state remain.
```

### Case 2: Factory defect survives restart

```text
1. Create and approve subcontract order.
2. Issue materials and receive finished goods.
3. Report factory defect within claim window.
4. Keep rejected quantity out of available stock.
5. Restart/redeploy API.
6. Confirm claim, rejected quantity, stock movement evidence, audit, and daily board state remain.
```

### Case 3: Partial acceptance survives restart

```text
1. Create subcontract order for 100 units.
2. Issue required materials.
3. Receive 80 units.
4. Accept 70 and reject/hold 10.
5. Restart/redeploy API.
6. Confirm available finished goods increase only by accepted quantity and the open/pending quantities remain traceable.
```

---

## 7. Sprint 16 Guardrails

These rules are non-negotiable:

```text
1. DB-mode subcontract selection must avoid partial subcontract truth.
2. Prototype fallback remains no-DB/local only.
3. Do not directly mutate stock balance.
4. Material issue and finished goods receipt stock changes must go through stock movement service.
5. Materials/packaging issued to factory must be traceable by item, UOM, base UOM, quantity, actor, and source document.
6. Lot/batch evidence must persist when lot trace is required.
7. Finished goods are not available until accepted/QC-pass behavior says they are available.
8. Factory defects must not silently increase available stock.
9. Payment milestone/final-payment readiness must preserve supplier payable link evidence.
10. Money/quantity/rate fields follow file 40 string-decimal and PostgreSQL numeric rules.
11. Every lifecycle action writes audit.
12. No public API response shape change unless OpenAPI and clients are updated in the same task.
```

---

## 8. Dependency Map

```text
S16-00-00 Sprint 16 task board
  -> S16-01-01 subcontract persistence design

S16-01-01 subcontract persistence design
  -> S16-01-02 subcontract migration foundation
  -> S16-02-01 subcontract order PostgreSQL store
  -> S16-03-01 material transfer PostgreSQL store
  -> S16-04-01 sample approval PostgreSQL store
  -> S16-05-01 finished goods receipt PostgreSQL store
  -> S16-06-01 factory claim PostgreSQL store
  -> S16-07-01 payment milestone PostgreSQL store

S16-02-01 subcontract order PostgreSQL store
  -> S16-02-02 subcontract order persistence tests
  -> S16-08-01 package runtime selectors

S16-03-01 material transfer PostgreSQL store
  -> S16-03-02 material transfer persistence tests
  -> S16-08-01 package runtime selectors

S16-04-01 sample approval PostgreSQL store
  -> S16-04-02 sample approval persistence tests
  -> S16-08-01 package runtime selectors

S16-05-01 finished goods receipt PostgreSQL store
  -> S16-05-02 finished goods receipt persistence tests
  -> S16-08-01 package runtime selectors

S16-06-01 factory claim PostgreSQL store
  -> S16-06-02 factory claim persistence tests
  -> S16-08-01 package runtime selectors

S16-07-01 payment milestone PostgreSQL store
  -> S16-07-02 payment milestone persistence tests
  -> S16-08-01 package runtime selectors

S16-08-01 package runtime selectors
  -> S16-08-02 warehouse daily board integration check
  -> S16-08-03 subcontract persistence smoke

S16-08-03 subcontract persistence smoke
  -> S16-09-01 remaining prototype ledger update
  -> S16-10-01 Sprint 16 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S16-00-00 | Sprint 16 task board | File 69 created with scope, guardrails, sequencing, verification gates, and task list | `docs/68_ERP_Sprint15_Changelog_Finance_Runtime_Store_Persistence_MyPham_v1.md` |
| S16-01-01 | Subcontract persistence design | Map existing subcontract domain contracts to PostgreSQL tables, selectors, fallback behavior, tests, smoke, and rollback | `docs/qa/S16-01-01_subcontract_persistence_design.md` |
| S16-01-02 | Subcontract migration foundation | Migration extends/creates subcontract runtime tables and indexes for order, transfer, sample, receipt, claim, and milestone persistence | `apps/api/migrations/000013_subcontract_order_core.up.sql` |
| S16-02-01 | Subcontract order PostgreSQL store | Order list/get/save persists status transitions, material lines, quantities, amounts, actor refs, and stable IDs | `apps/api/internal/modules/production/application/postgres_subcontract_order_store.go` |
| S16-02-02 | Subcontract order persistence tests | Fresh store reload and lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/production/application/postgres_subcontract_order_store_test.go` |
| S16-03-01 | Material transfer PostgreSQL store | Material/packaging transfer list/save persists lines, source refs, issue evidence, and stock movement refs | `apps/api/internal/modules/production/application/postgres_subcontract_material_transfer_store.go` |
| S16-03-02 | Material transfer persistence tests | Fresh store reload and issue-materials lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/production/application/postgres_subcontract_material_transfer_store_test.go` |
| S16-04-01 | Sample approval PostgreSQL store | Sample submit/approve/reject evidence persists with reason and actor refs | `apps/api/internal/modules/production/application/postgres_subcontract_sample_approval_store.go` |
| S16-04-02 | Sample approval persistence tests | Fresh store reload and sample lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/production/application/postgres_subcontract_sample_approval_store_test.go` |
| S16-05-01 | Finished goods receipt PostgreSQL store | Receipt/accept/partial accept persists quantities, stock movement refs, and actor refs | `apps/api/internal/modules/production/application/postgres_subcontract_finished_goods_receipt_store.go` |
| S16-05-02 | Finished goods receipt persistence tests | Fresh store reload and receipt/accept lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/production/application/postgres_subcontract_finished_goods_receipt_store_test.go` |
| S16-06-01 | Factory claim PostgreSQL store | Defect/claim evidence persists with reason, claim window, quantity impact, and actor refs | `apps/api/internal/modules/production/application/subcontract_factory_claim_service.go` |
| S16-06-02 | Factory claim persistence tests | Fresh store reload and factory defect lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/production/application/subcontract_factory_claim_service_test.go` |
| S16-07-01 | Payment milestone PostgreSQL store | Deposit/final payment milestone state persists with supplier payable link evidence | `apps/api/internal/modules/production/application/subcontract_payment_milestone_service.go` |
| S16-07-02 | Payment milestone persistence tests | Fresh store reload and payable-link lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/production/application/subcontract_payment_milestone_service_test.go` |
| S16-08-01 | Package runtime selectors | DB mode wires all subcontract stores together; no partial subcontract DB selection | `apps/api/cmd/api/main.go` |
| S16-08-02 | Warehouse daily board integration check | Daily board reads the same DB-backed subcontract stores and keeps response shape | `apps/api/cmd/api/main.go` |
| S16-08-03 | Subcontract persistence smoke | Full dev smoke proves order/transfer/sample/receipt/claim/milestone state survives restart/redeploy | `infra/scripts/smoke-dev-full.sh` |
| S16-09-01 | Remaining prototype ledger update | Remaining prototype ledger supersedes Sprint 15 and removes subcontract stores from production persistence gaps | `docs/qa/S14-04-01_remaining_prototype_store_ledger.md` |
| S16-10-01 | Sprint 16 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/70_ERP_Sprint16_Changelog_Subcontract_Runtime_Store_Persistence_MyPham_v1.md` |

---

## 10. Verification Gates

Backend checks:

```text
go test ./...
go vet ./...
```

Focused subcontract checks:

```text
go test ./internal/modules/production/... ./cmd/api -run "Test(Subcontract|WarehouseDailyBoard)" -count=1
```

Migration checks when migrations change:

```text
Apply all migrations on PostgreSQL 16.
Roll back all migrations on PostgreSQL 16.
Apply again after rollback when practical.
```

OpenAPI checks when API contracts change:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --exit-code apps/web/src/shared/api/generated/schema.ts
```

Dev release gate:

```text
Dev deploy or repo sync evidence for merged main.
Dev release gate smoke.
GitHub required checks green.
```

---

## 11. Definition Of Done

For each code task:

```text
1. Code is scoped to the task.
2. Money, quantity, and rate fields follow file 40 decimal string rules.
3. Backend tests pass for touched services/stores/handlers.
4. Web test/typecheck/build pass for touched UI areas.
5. OpenAPI validate/contract/generate pass when API contracts change.
6. PR includes manual self-review comment.
7. Runtime changes are deployed to dev server after merge.
8. Any remaining unverified release gate is called out explicitly.
```

Sprint 16 completion requires:

```text
1. All S16 tasks merged to main.
2. DB-mode runtime selection wires subcontract order/transfer/sample/receipt/claim/milestone as one package.
3. Dev server full smoke passes after final runtime merge.
4. GitHub required checks are green.
5. Migration apply/rollback is verified on PostgreSQL 16.
6. Remaining prototype ledger is updated.
7. Tag v0.16.0-subcontract-runtime-store-persistence is created only after release gates are green.
```
