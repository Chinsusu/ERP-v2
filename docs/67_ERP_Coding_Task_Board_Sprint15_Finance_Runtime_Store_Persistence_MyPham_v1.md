# 67_ERP_Coding_Task_Board_Sprint15_Finance_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 15 - Finance runtime store persistence v1
Document role: Coding task board for Sprint 15 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, Next.js frontend, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 14 release gate

---

## 1. Sprint 15 Context

Sprint 6 built the Finance Lite COD, AR, AP, cash, dashboard, OpenAPI, UI, permission, audit, and E2E flows.
Sprint 14 closed shipping execution persistence and updated the remaining prototype ledger.

The highest remaining production persistence risk is now Finance Lite runtime state:

```text
Customer receivables -> prototype memory
Supplier payables    -> prototype memory
COD remittances      -> prototype memory
Cash transactions    -> prototype memory
Finance dashboard    -> reads prototype-backed runtime stores
```

This matters because sales orders, reservations, purchase orders, inbound QC, return/rejection evidence, stock movements, daily close, and shipping execution state are now durable while the money state can still reset after API restart.

---

## 2. Sprint 15 Theme

```text
Finance Runtime Store Persistence
```

Business reason:

```text
Operations can now keep durable evidence for outbound, inbound, returns, daily close, and shipping execution.
Finance must keep durable AR/AP/COD/cash state so open receivables, payable approvals, COD discrepancies, and cash allocations survive deploys and restarts.
```

---

## 3. Sprint 15 Goals

By the end of Sprint 15, DB-mode runtime must support:

```text
1. Customer receivables persisted with header, lines, source-document refs, receipt/dispute/void status, and audit.
2. Supplier payables persisted with header, lines, payment request/approval/payment status, and audit.
3. COD remittances persisted with header, lines, discrepancies, submit/approve/close status, and audit.
4. Cash transactions persisted with allocations and source refs.
5. Finance dashboard and finance report read the same durable finance stores.
6. Prototype fallback remains explicit for no-DB/local mode.
7. Dev smoke proves finance runtime state survives restart/redeploy.
8. Remaining prototype ledger is updated after finance persistence.
9. Release evidence records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status.
```

Important runtime rule:

```text
Do not wire only one finance store into DB mode while the other finance stores remain memory-backed.
Finance DB-mode selection should switch as a package after AR, AP, COD, and cash stores are all present and tested.
```

---

## 4. Sprint 15 Non-Goals

Sprint 15 does not include:

```text
- Full double-entry general ledger.
- Bank statement import or bank API integration.
- Payment gateway integration.
- Tax filing, VAT declaration, or e-invoice integration.
- New finance frontend screens beyond smoke/test adjustments if needed.
- Full COGS, landed cost, or P&L allocation.
- Multi-currency.
- Broad finance UI redesign.
- Subcontract runtime store persistence.
- Master data catalog persistence.
- Auth/session hardening.
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
codex/feature-S15-xx-yy-short-task-name
```

Recommended Sprint 15 release tag after completion:

```text
v0.15.0-finance-runtime-store-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for new migrations.
4. Sprint 15 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

---

## 6. Sprint 15 Demo Script

### Case 1: COD happy path survives restart

```text
1. Create customer receivable for COD source.
2. Create COD remittance with matching line.
3. Match, submit, approve, and close remittance.
4. Record cash receipt allocation.
5. Restart/redeploy API.
6. Confirm receivable, remittance, cash transaction, dashboard, and audit state remain.
```

### Case 2: COD discrepancy survives restart

```text
1. Create COD remittance with short remitted amount.
2. Record discrepancy reason and owner.
3. Keep remittance/receivable open or disputed according to current rules.
4. Restart/redeploy API.
5. Confirm discrepancy, status, and audit state remain.
```

### Case 3: Supplier payable payment survives restart

```text
1. Create supplier payable from accepted source.
2. Request payment.
3. Approve payment.
4. Record cash payment allocation.
5. Restart/redeploy API.
6. Confirm payable status, payment evidence, cash transaction, dashboard, and audit state remain.
```

---

## 7. Sprint 15 Guardrails

These rules are non-negotiable:

```text
1. VND remains the base currency.
2. Money values remain string decimals at API boundaries.
3. Never use float/double for money, quantity, rate, or percentage.
4. Do not directly mutate AR/AP/COD/cash balances outside finance document services.
5. Do not treat COD as cash received until remittance/receipt evidence exists.
6. Payment requires approval before cash payment is recorded.
7. Every create/receipt/dispute/void/match/submit/approve/close/payment action writes audit.
8. Finance stores must preserve source-document refs and line IDs across saves.
9. Runtime DB-mode selection must avoid partial finance truth.
10. Prototype fallback remains explicit for no-DB/local mode.
11. No public API response shape change unless OpenAPI and clients are updated in the same task.
12. No broad finance UI or reporting refactor.
```

---

## 8. Dependency Map

```text
S15-00-00 Sprint 15 task board
  -> S15-01-01 finance persistence design

S15-01-01 finance persistence design
  -> S15-01-02 finance migration foundation
  -> S15-02-01 customer receivable PostgreSQL store
  -> S15-03-01 supplier payable PostgreSQL store
  -> S15-04-01 COD remittance PostgreSQL store
  -> S15-05-01 cash transaction PostgreSQL store

S15-01-02 finance migration foundation
  -> S15-02-01 customer receivable PostgreSQL store
  -> S15-03-01 supplier payable PostgreSQL store
  -> S15-04-01 COD remittance PostgreSQL store
  -> S15-05-01 cash transaction PostgreSQL store

S15-02-01 customer receivable PostgreSQL store
  -> S15-02-02 customer receivable persistence tests
  -> S15-06-01 package runtime selectors

S15-03-01 supplier payable PostgreSQL store
  -> S15-03-02 supplier payable persistence tests
  -> S15-06-01 package runtime selectors

S15-04-01 COD remittance PostgreSQL store
  -> S15-04-02 COD remittance persistence tests
  -> S15-06-01 package runtime selectors

S15-05-01 cash transaction PostgreSQL store
  -> S15-05-02 cash transaction persistence tests
  -> S15-06-01 package runtime selectors

S15-06-01 package runtime selectors
  -> S15-06-02 finance dashboard/report integration check
  -> S15-06-03 finance persistence smoke

S15-06-03 finance persistence smoke
  -> S15-07-01 remaining prototype ledger update
  -> S15-08-01 Sprint 15 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S15-00-00 | Sprint 15 task board | File 67 created with scope, guardrails, sequencing, verification gates, and task list | `docs/66_ERP_Sprint14_Changelog_Shipping_Pick_Pack_Persistence_MyPham_v1.md` |
| S15-01-01 | Finance persistence design | Map AR/AP/COD/cash domain contracts to PostgreSQL tables, selectors, fallback behavior, tests, smoke, and migration rollback | `docs/qa/S14-04-01_remaining_prototype_store_ledger.md` |
| S15-01-02 | Finance migration foundation | Migration adds finance runtime refs/indexes/child tables needed by AR/AP/COD/cash persistence | `apps/api/internal/modules/finance/domain` |
| S15-02-01 | Customer receivable PostgreSQL store | AR list/get/save persist header, lines, source refs, receipt/dispute/void state, and stable line IDs | `apps/api/internal/modules/finance/application/customer_receivable_service.go` |
| S15-02-02 | Customer receivable persistence tests | Fresh store reload and action lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/finance/application/customer_receivable_service_test.go` |
| S15-03-01 | Supplier payable PostgreSQL store | AP list/get/save persist header, lines, approval/payment status, source refs, and stable line IDs | `apps/api/internal/modules/finance/application/supplier_payable_service.go` |
| S15-03-02 | Supplier payable persistence tests | Fresh store reload and payment lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/finance/application/supplier_payable_service_test.go` |
| S15-04-01 | COD remittance PostgreSQL store | COD list/get/save persist header, lines, discrepancy evidence, submit/approve/close state, and stable line IDs | `apps/api/internal/modules/finance/application/cod_remittance_service.go` |
| S15-04-02 | COD remittance persistence tests | Fresh store reload and discrepancy lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/finance/application/cod_remittance_service_test.go` |
| S15-05-01 | Cash transaction PostgreSQL store | Cash list/get/save persist direction, amount, allocations, source refs, and status | `apps/api/internal/modules/finance/application/cash_transaction_service.go` |
| S15-05-02 | Cash transaction persistence tests | Fresh store reload and allocation tests pass against PostgreSQL store | `apps/api/internal/modules/finance/application/cash_transaction_service_test.go` |
| S15-06-01 | Package runtime selectors | DB mode wires AR/AP/COD/cash stores together; no partial finance DB selection | `apps/api/cmd/api/main.go` |
| S15-06-02 | Finance dashboard/report integration check | Dashboard/report read the same DB-backed finance stores and keep existing response shape | `apps/api/internal/modules/finance/application/finance_dashboard_service.go` |
| S15-06-03 | Finance persistence smoke | Full dev smoke proves AR/AP/COD/cash state survives restart/redeploy | `infra/scripts/smoke-dev-full.sh` |
| S15-07-01 | Remaining prototype ledger update | Remaining prototype ledger supersedes S14 and removes finance stores from production persistence gaps | `docs/qa/S14-04-01_remaining_prototype_store_ledger.md` |
| S15-08-01 | Sprint 15 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/68_ERP_Sprint15_Changelog_Finance_Runtime_Store_Persistence_MyPham_v1.md` |

---

## 10. Verification Gates

Backend checks:

```text
go test ./...
go vet ./...
```

Focused finance checks:

```text
go test ./internal/modules/finance/... ./cmd/api -run "Test(Finance|CustomerReceivable|SupplierPayable|CODRemittance|CashTransaction)" -count=1
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

Sprint 15 completion requires:

```text
1. All S15 tasks merged to main.
2. DB-mode runtime selection wires AR/AP/COD/cash as one finance package.
3. Dev server full smoke passes after final runtime merge.
4. GitHub required checks are green.
5. Migration apply/rollback is verified on PostgreSQL 16.
6. Remaining prototype ledger is updated.
7. Tag v0.15.0-finance-runtime-store-persistence is created only after release gates are green.
```
