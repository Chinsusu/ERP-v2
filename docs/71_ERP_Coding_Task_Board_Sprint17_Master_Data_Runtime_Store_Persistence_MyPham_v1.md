# 71_ERP_Coding_Task_Board_Sprint17_Master_Data_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 17 - Master data runtime store persistence v1
Document role: Coding task board for Sprint 17 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL 16, OpenAPI, Docker dev deploy
Status: Ready for implementation after Sprint 16 manual/dev release gate

---

## 1. Sprint 17 Context

Sprint 16 persisted subcontract runtime state and updated the remaining prototype store ledger.

The highest remaining production persistence risk is now master data catalog state:

```text
Item catalog                  -> prototype memory
Warehouse/location catalog    -> prototype memory
Party catalog                 -> prototype memory
UOM catalog                   -> prototype memory
```

This matters because purchasing, inbound receiving, stock movements, shipping, returns, subcontract manufacturing, finance, and reporting all reference master data. Operational documents are now durable, but editable item, location, supplier/customer, and UOM catalog changes can still reset after API restart.

---

## 2. Sprint 17 Theme

```text
Master Data Runtime Store Persistence
```

Business reason:

```text
The ERP cannot be production-traceable if product, warehouse/location, supplier/customer, or unit conversion data can disappear on deploy. Master data must survive restarts before broader admin workflows, pricing setup, supplier maintenance, or warehouse layout control can be trusted.
```

---

## 3. Sprint 17 Goals

By the end of Sprint 17, DB-mode runtime must support:

```text
1. Item catalog state persisted with item code, SKU, type, status, UOM, shelf-life, cost, and audit.
2. Warehouse/location catalog state persisted with warehouse, location, status, type, parent relation, and audit.
3. Party catalog state persisted with supplier/customer identity, group/type, status, credit/lead-time metrics, and audit.
4. UOM catalog state persisted with unit definitions, conversion factors, base UOM mapping, and decimal precision rules.
5. Runtime selectors use PostgreSQL stores when DATABASE_URL exists and prototype stores only for no-DB/local mode.
6. Existing API response shapes and validation behavior remain stable.
7. Dev smoke proves master data state survives API restart.
8. Remaining prototype ledger and Sprint 17 release evidence are updated after master data persistence.
```

---

## 4. Sprint 17 Non-Goals

Sprint 17 does not include:

```text
- New frontend master data screens.
- Product pricing policy redesign.
- Supplier/customer CRM workflows.
- Barcode generation or label printing.
- Warehouse bin optimization.
- Multi-currency master data.
- Production-grade auth/session persistence.
- Broad schema normalization outside master data catalogs.
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
codex/feature-S17-xx-yy-short-task-name
```

Recommended Sprint 17 release tag after completion:

```text
v0.17.0-master-data-runtime-store-persistence
```

Create the production tag only after:

```text
1. Main required-ci is green, or the team explicitly accepts a manual-only release gate while GitHub Actions quota is exhausted.
2. Dev release gate is green.
3. PostgreSQL migration apply/rollback evidence is green for new migrations.
4. Sprint 17 changelog records persisted scope, remaining prototype stores, dev deploy, CI, and tag status.
```

Current CI note:

```text
GitHub Actions cloud CI is blocked because the monthly included minutes are exhausted: 2,000 min used / 2,000 min included.
Until that changes, do not claim cloud CI verification and do not tag production without explicit manual-only release approval.
```

---

## 6. Sprint 17 Demo Script

### Case 1: Product master data survives restart

```text
1. Create item with code, SKU, type, UOM, shelf-life, and cost.
2. Update item fields.
3. Change item status.
4. Restart/redeploy API.
5. Confirm item list/detail and audit evidence remain.
```

### Case 2: Warehouse/location master data survives restart

```text
1. Create warehouse.
2. Create location under that warehouse.
3. Update warehouse/location fields.
4. Change status.
5. Restart/redeploy API.
6. Confirm warehouse/location list/detail and audit evidence remain.
```

### Case 3: Party and UOM master data survive restart

```text
1. Create supplier and customer.
2. Update supplier/customer fields and status.
3. Create or verify UOM conversion data.
4. Restart/redeploy API.
5. Confirm supplier/customer and UOM conversion state remain.
```

---

## 7. Sprint 17 Guardrails

These rules are non-negotiable:

```text
1. DB-mode master data selection must avoid partial catalog truth.
2. Prototype fallback remains no-DB/local only.
3. Do not change public API response shape unless OpenAPI and clients are updated in the same task.
4. Do not loosen master data validation or duplicate-code checks.
5. Preserve existing audit actions and audit payload semantics.
6. Money/quantity/rate fields follow file 40 decimal string and PostgreSQL numeric rules.
7. UOM conversion must not use float/double.
8. Existing seeded/mock catalog behavior must remain available for local/no-DB tests.
9. Runtime changes must have focused store tests and dev smoke evidence.
10. Master data persistence must not directly mutate operational stock, finance, purchase, sales, shipping, returns, or subcontract tables.
```

---

## 8. Dependency Map

```text
S17-00-00 Sprint 17 task board
  -> S17-01-01 master data persistence design

S17-01-01 master data persistence design
  -> S17-01-02 master data migration foundation
  -> S17-02-01 item catalog PostgreSQL store
  -> S17-03-01 warehouse/location catalog PostgreSQL store
  -> S17-04-01 party catalog PostgreSQL store
  -> S17-05-01 UOM catalog PostgreSQL store

S17-02-01 item catalog PostgreSQL store
  -> S17-02-02 item catalog persistence tests
  -> S17-06-01 package runtime selectors

S17-03-01 warehouse/location catalog PostgreSQL store
  -> S17-03-02 warehouse/location persistence tests
  -> S17-06-01 package runtime selectors

S17-04-01 party catalog PostgreSQL store
  -> S17-04-02 party catalog persistence tests
  -> S17-06-01 package runtime selectors

S17-05-01 UOM catalog PostgreSQL store
  -> S17-05-02 UOM catalog persistence tests
  -> S17-06-01 package runtime selectors

S17-06-01 package runtime selectors
  -> S17-06-02 master data restart smoke

S17-06-02 master data restart smoke
  -> S17-07-01 remaining prototype ledger update
  -> S17-08-01 Sprint 17 release evidence
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S17-00-00 | Sprint 17 task board | File 71 created with scope, guardrails, sequencing, verification gates, and task list | `docs/70_ERP_Sprint16_Changelog_Subcontract_Runtime_Store_Persistence_MyPham_v1.md` |
| S17-01-01 | Master data persistence design | Map item, warehouse/location, party, and UOM catalog contracts to PostgreSQL tables, selectors, fallback behavior, tests, smoke, and rollback | `docs/qa/S17-01-01_master_data_persistence_design.md` |
| S17-01-02 | Master data migration foundation | Migration creates/extends master data runtime tables, indexes, constraints, and seed-compatible structures | `apps/api/migrations/000034_persist_master_data_runtime_foundation.up.sql` |
| S17-02-01 | Item catalog PostgreSQL store | Item list/get/create/update/status persists fields, duplicate-code checks, and audit behavior | `apps/api/internal/modules/masterdata/application/postgres_item_catalog.go` |
| S17-02-02 | Item catalog persistence tests | Fresh store reload and item lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/masterdata/application/postgres_item_catalog_test.go` |
| S17-03-01 | Warehouse/location catalog PostgreSQL store | Warehouse and location list/get/create/update/status persist hierarchy, status, and duplicate-code checks | `apps/api/internal/modules/masterdata/application/postgres_warehouse_location_catalog.go` |
| S17-03-02 | Warehouse/location persistence tests | Fresh store reload and warehouse/location lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/masterdata/application/postgres_warehouse_location_catalog_test.go` |
| S17-04-01 | Party catalog PostgreSQL store | Supplier and customer list/get/create/update/status persist identity, group/type, metrics, and duplicate-code checks | `apps/api/internal/modules/masterdata/application/postgres_party_catalog.go` |
| S17-04-02 | Party catalog persistence tests | Fresh store reload and supplier/customer lifecycle tests pass against PostgreSQL store | `apps/api/internal/modules/masterdata/application/postgres_party_catalog_test.go` |
| S17-05-01 | UOM catalog PostgreSQL store | UOM definitions and conversion factors persist without float/double and keep base UOM conversion behavior | `apps/api/internal/modules/masterdata/application/postgres_uom_catalog.go` |
| S17-05-02 | UOM catalog persistence tests | Fresh store reload and conversion tests pass against PostgreSQL store | `apps/api/internal/modules/masterdata/application/postgres_uom_catalog_test.go` |
| S17-06-01 | Package runtime selectors | DB mode wires all master data catalogs together; no partial master data DB selection | `apps/api/cmd/api/masterdata_store_selection.go` |
| S17-06-02 | Master data restart smoke | Full dev smoke proves item, warehouse/location, party, and UOM state survives API restart/redeploy | `infra/scripts/smoke-dev-full.sh` |
| S17-07-01 | Remaining prototype ledger update | Remaining prototype ledger supersedes Sprint 16 and removes master data catalogs from production persistence gaps | `docs/qa/S14-04-01_remaining_prototype_store_ledger.md` |
| S17-08-01 | Sprint 17 release evidence | Changelog records PRs, migrations, verification, persisted stores, remaining prototype stores, dev deploy, CI, and tag status | `docs/72_ERP_Sprint17_Changelog_Master_Data_Runtime_Store_Persistence_MyPham_v1.md` |

---

## 10. Verification Gates

Backend checks:

```text
go test ./...
go vet ./...
```

Focused master data checks:

```text
go test ./internal/modules/masterdata/... ./cmd/api -run "Test(Postgres|MasterData|Product|Warehouse|Supplier|Customer|UOM)" -count=1
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
GitHub required checks green only when Actions minutes are available.
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

Sprint 17 completion requires:

```text
1. All S17 tasks merged to main.
2. DB-mode runtime selection wires item, warehouse/location, party, and UOM catalogs as one package.
3. Dev server full smoke passes after final runtime merge.
4. PostgreSQL migration apply/rollback is verified on PostgreSQL 16.
5. Remaining prototype ledger is updated.
6. Sprint 17 changelog records that GitHub Actions cloud CI is blocked if quota is still exhausted.
7. Tag v0.17.0-master-data-runtime-store-persistence is created only after release gates are green or explicit manual-only release approval is given.
```
