# 64_ERP_Sprint13_Changelog_End_of_Day_Reconciliation_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 13 - End-of-day reconciliation persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Release evidence prepared; production tag pending after this changelog merges and main CI is green

---

## 1. Sprint 13 Scope

Sprint 13 closed the highest remaining warehouse control evidence risk after Sprint 12:

```text
end-of-day reconciliation list/close was prototype-only
shift close status could reset on API restart
exception note, checklist, and variance evidence could reset with memory
```

Promoted scope:

```text
Sprint 13 task board
end-of-day reconciliation persistence design
PostgreSQL-backed runtime selector and store
warehouse_daily_closings runtime refs and child evidence tables
focused PostgreSQL store tests
dev close persistence smoke with API restart
remaining prototype store ledger update
```

No new frontend screen or public API envelope was introduced. Sprint 13 changed persistence behavior behind the existing warehouse end-of-day reconciliation APIs.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S13-00-00 Sprint 13 task board | #433 | Created Sprint 13 task board |
| S13-01-01 Persistence design | #434 | Documented schema, store, selector, audit, tests, and smoke plan |
| S13-01-02/S13-01-03 Runtime selector and PostgreSQL store | #435 | Added DB-mode runtime selector, migration 000026, and PostgreSQL store |
| S13-01-04 Store and handler tests | #436 | Added focused Postgres close persistence, reload, filter, and audit coverage |
| S13-02-01 Close persistence smoke | #437 | Proved close evidence survives API restart on dev |
| S13-03-01 Remaining prototype ledger update | #438 | Superseded Sprint 12 remaining-store ledger |

All PRs used the manual review and merge flow.

---

## 3. Persistence Changes

### Runtime Selector

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeEndOfDayReconciliationStore` | `PostgresEndOfDayReconciliationStore` | `PrototypeEndOfDayReconciliationStore` |

Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000026_persist_end_of_day_reconciliations` | Adds runtime refs and evidence tables for end-of-day reconciliation close state |

Persisted behavior:

```text
GET /api/v1/warehouse/end-of-day-reconciliations
POST /api/v1/warehouse/end-of-day-reconciliations/{reconciliation_id}/close
```

Persisted evidence:

```text
inventory.warehouse_daily_closings.status
inventory.warehouse_daily_closings.closed_at
inventory.warehouse_daily_closings.closed_by_ref
inventory.warehouse_daily_closings.exception_note
inventory.warehouse_daily_closing_checklist
inventory.warehouse_daily_closing_lines
audit.audit_logs action warehouse.shift.closed
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
After PR #435, main was deployed to dev.
Deploy built API, worker, and web images from source.
Migration 000026 applied on dev.
Deploy smoke passed.
Full host smoke passed.
```

Focused Sprint 13 smoke evidence:

```text
S13-02-01: close status, closed_by, exception note, variance evidence, and audit survived API restart.
Evidence file: docs/qa/S13-02-01_end_of_day_reconciliation_close_persistence_smoke.md
```

Release smoke at current Sprint 13 main:

```text
Commit: 273886fa
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

End-of-day reconciliation persistence is covered by the focused S13-02-01 smoke because it requires a dedicated close/restart/read flow.

---

## 5. CI And Migration Evidence

GitHub checks:

```text
PR #435: api, e2e, migration, openapi, required-api, required-migration, required-openapi, required-web passed.
PR #436: api, e2e, required-api, required-migration, required-openapi, required-web passed.
PR #437: e2e, required-api, required-migration, required-openapi, required-web passed.
PR #438: e2e, required-api, required-migration, required-openapi, required-web passed.
```

Local/dev verification:

```text
PR #435 dev server backend test: go test ./... passed.
PR #435 dev server vet: go vet ./... passed.
PR #436 focused DB test: TestPostgresEndOfDayReconciliation passed against dev PostgreSQL.
PR #436 dev server backend test: go test ./... passed.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
Source: /opt/ERP-v2 at PR #435 branch
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
Applied migrations: 26
Rolled back migrations: 26
```

Migration notices observed during rollback were idempotent schema/constraint notices from earlier migrations. They did not fail the gate.

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S13-03-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 13:

```text
1. Shipping manifest, pick, and pack task package.
2. Finance AR/AP/COD/cash runtime stores.
3. Subcontract runtime stores.
4. Master data catalogs and auth/session hardening.
```

---

## 7. Release Status

Sprint 13 release gate status:

```text
Task PRs: merged through S13-03-01
Current changelog PR: pending merge
Main CI: green through PR #438; rerun required for this changelog PR
Dev runtime smoke: green at commit 273886fa
Focused close persistence smoke: green
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: pending
```

Recommended tag after this changelog merges and main CI is green:

```text
v0.13.0-end-of-day-reconciliation-persistence
```

Do not move the tag once pushed. If a post-tag fix is needed, create a new patch tag instead.
