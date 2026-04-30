# 51_ERP_Coding_Task_Board_Sprint7_Reporting_Inventory_Operations_Dashboard_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 7 - Reporting v1 / Inventory / Operations Dashboard
Document role: Coding task board for Sprint 7 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL, Next.js frontend, OpenAPI, CSV export
Status: Ready for implementation; production tag remains on hold until GitHub Actions billing blocker is cleared

---

## 1. Sprint 7 Context

Sprint 2 controlled outbound order fulfillment.
Sprint 3 controlled returns, reconciliation, and warehouse shift close.
Sprint 4 controlled purchase inbound and QC.
Sprint 5 controlled subcontract manufacturing and factory material/finished goods flow.
Sprint 6 added Finance Lite for COD, AR, AP, cash in/out, and payment approval.

Sprint 7 turns those operational records into practical reporting:

```text
Inventory and stock status
-> snapshot metrics
-> low stock / quarantine / reserved / available visibility
-> CSV export

Warehouse and operations
-> daily KPI by date range
-> inbound, outbound, returns, stock count, QC, subcontract signals
-> dashboard cards and drilldown tables

Finance Lite
-> AR / AP / COD / cash summary report
-> aging buckets and discrepancy visibility
-> CSV export
```

This sprint is reporting v1. It should make existing operational data visible and exportable. It must not become a full BI warehouse or accounting reporting engine.

---

## 2. Sprint 7 Theme

```text
Reporting v1 + Inventory / Operations Dashboard
```

Business reason:

```text
The system can now create and control orders, warehouse movements, returns, inbound QC, subcontract flow, and finance-lite documents.
Managers need a reliable daily view:
- What stock is available, reserved, quarantined, or low.
- What warehouse work is pending or blocked today.
- What COD, receivables, payables, and cash movements need attention.
- What can be exported for reconciliation without manual spreadsheet rebuilds.
```

---

## 3. Sprint 7 Goals

By the end of Sprint 7, the system must support:

```text
1. Define reporting permissions and navigation.
2. Provide a shared reporting date range / warehouse / status filter model.
3. Provide inventory snapshot report API and UI.
4. Show available, reserved, quarantine, low-stock, and batch/expiry warning signals.
5. Provide operations daily KPI report API and UI.
6. Summarize inbound, outbound, returns, QC, stock adjustment, and subcontract work.
7. Provide finance summary report API and UI using Sprint 6 data.
8. Summarize AR, AP, COD discrepancy, and cash in/out for a selected date range.
9. Provide CSV export for the first reporting screens.
10. Update OpenAPI, generated FE client, and contract checks.
11. Add focused regression and smoke coverage for reports, permissions, filters, and exports.
```

---

## 4. Sprint 7 Non-Goals

Sprint 7 does not include:

```text
- Full data warehouse or OLAP cube.
- Custom report builder.
- Scheduled email reports.
- BI chart designer.
- Full accounting financial statements.
- Tax/VAT reports.
- External BI integration.
- Row-level security beyond existing role permission gates.
- Real-time websocket dashboards.
- PDF report generation.
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
codex/feature-S7-xx-yy-short-task-name
```

Recommended Sprint 7 release tag after completion:

```text
v0.7.0-reporting-inventory-operations-dashboard
```

Production tags remain on hold while GitHub Actions is blocked by the billing/spending-limit issue.
Sprint 5 and Sprint 6 production tags remain carry-forward blockers until CI can run green on GitHub and the release tags are created.

---

## 6. Sprint 7 Demo Script

### Case 1: Inventory snapshot

```text
1. Open Reporting -> Inventory Snapshot.
2. Select warehouse and business date.
3. View available, reserved, quarantine, and low-stock totals.
4. Drill into item/batch rows.
5. Export CSV and confirm decimal string values and vi-VN display formatting.
```

### Case 2: Operations daily KPI

```text
1. Open Reporting -> Operations Daily.
2. Select date range and warehouse.
3. View inbound receiving, QC hold/pass/fail, pick/pack/handover, returns, stock count, and subcontract signals.
4. Confirm blocked work and exceptions are visible.
```

### Case 3: Finance summary

```text
1. Open Reporting -> Finance Summary.
2. Select date range.
3. View AR open/overdue, AP due/open, COD pending/discrepancy, and cash in/out.
4. Export CSV for finance reconciliation.
```

---

## 7. Sprint 7 Guardrails

These rules are non-negotiable:

```text
1. Reports must read from controlled domain records and services, not mutate operational data.
2. No report endpoint may update stock, AR/AP, COD, cash, QC, or order status.
3. Money, quantity, and rate values remain decimal strings in APIs.
4. VND, vi-VN, and Asia/Ho_Chi_Minh rules from file 40 still apply.
5. Inventory figures must distinguish available, reserved, quarantine, hold, and rejected states.
6. Do not mix available stock with QC hold or quarantine stock.
7. COD remitted is not cash received unless Sprint 6 cash transaction data says so.
8. CSV export must use stable headers and machine-readable decimal strings.
9. Date range filters must be inclusive and use business dates, not server local timestamps.
10. Reporting permissions must protect finance reports from non-finance users.
11. OpenAPI must be updated for every new report endpoint.
12. Sprint 7 is not production-ready until the GitHub CI blocker is cleared and release gates are green.
```

---

## 8. Dependency Map

```text
S7-00-00 Sprint 7 task board
  -> S7-01-01 reporting permissions and menu
  -> S7-01-02 reporting filter/export foundation

S7-01-01 reporting permissions and menu
  -> S7-02-03 inventory snapshot UI
  -> S7-03-03 operations daily UI
  -> S7-04-03 finance summary UI
  -> S7-07-01 permission regression

S7-01-02 reporting filter/export foundation
  -> S7-02-01 inventory snapshot query model
  -> S7-03-01 operations daily query model
  -> S7-04-01 finance summary query model
  -> S7-05-01 CSV export foundation

S7-02-01 inventory snapshot query model
  -> S7-02-02 inventory snapshot API
  -> S7-02-03 inventory snapshot UI
  -> S7-07-02 inventory report smoke

S7-03-01 operations daily query model
  -> S7-03-02 operations daily API
  -> S7-03-03 operations daily UI
  -> S7-07-03 operations report smoke

S7-04-01 finance summary query model
  -> S7-04-02 finance summary API
  -> S7-04-03 finance summary UI
  -> S7-07-04 finance report smoke

S7-05-01 CSV export foundation
  -> S7-05-02 inventory CSV export
  -> S7-05-03 operations CSV export
  -> S7-05-04 finance CSV export

S7-06-01 OpenAPI endpoints
  -> S7-06-02 generated FE client
  -> S7-06-03 contract check

S7-07-01 permission regression
  -> S7-08-01 Sprint 7 release evidence
```

---

## 9. API Shape

Use slash action endpoints, not colon action endpoints.

Recommended API surface:

```text
GET /api/v1/reports/inventory-snapshot
GET /api/v1/reports/inventory-snapshot/export.csv

GET /api/v1/reports/operations-daily
GET /api/v1/reports/operations-daily/export.csv

GET /api/v1/reports/finance-summary
GET /api/v1/reports/finance-summary/export.csv
```

Common query parameters:

```text
from_date=YYYY-MM-DD
to_date=YYYY-MM-DD
business_date=YYYY-MM-DD
warehouse_id=...
status=...
item_id=...
category=...
```

Rules:

```text
- Use inclusive date ranges.
- Return decimal strings for money and quantities.
- Return report metadata: generated_at, timezone, filters, source_version.
- Export endpoints must use text/csv and stable column names.
```

---

## 10. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S7-00-00 | Sprint 7 task board | File 51 created with goals, guardrails, dependencies, API shape, and task list | 01, 03, 16, 24, 39, 40, 50 |
| S7-00-01 | Sprint 6 release gate carry-forward | Release evidence states GitHub CI blocker and tag hold status | 50 |
| S7-01-01 | Reporting permissions and menu | Reporting view/export permissions and web navigation entry are available | 04, 19, 39 |
| S7-01-02 | Reporting filter/export foundation | Shared date range, warehouse/status filters, and CSV response helpers are defined | 16, 24, 40 |
| S7-02-01 | Inventory snapshot query model | Report model covers available, reserved, quarantine, low-stock, batch/expiry warning | 17, 40 |
| S7-02-02 | Inventory snapshot API | Inventory snapshot endpoint with permission, filters, and decimal string outputs | 16, 19 |
| S7-02-03 | Inventory snapshot UI | Reporting screen with cards/table/filter/export action | 39 |
| S7-03-01 | Operations daily query model | KPI model covers receiving, QC, shipping, returns, stock count, subcontract | 20, 41, 43, 45, 47 |
| S7-03-02 | Operations daily API | Operations daily endpoint with date/warehouse filters and exception counts | 16, 19 |
| S7-03-03 | Operations daily UI | Daily operations dashboard with actionable KPI cards and tables | 39 |
| S7-04-01 | Finance summary query model | AR/AP/COD/cash summary and aging/discrepancy buckets | 49, 50 |
| S7-04-02 | Finance summary API | Finance summary endpoint with finance permission gates | 16, 19 |
| S7-04-03 | Finance summary UI | Finance reporting screen with summary cards, buckets, and export action | 39 |
| S7-05-01 | CSV export foundation | Shared CSV writer, stable headers, content type, and tests | 16, 24 |
| S7-05-02 | Inventory CSV export | Inventory snapshot export endpoint and UI action | 16, 39 |
| S7-05-03 | Operations CSV export | Operations daily export endpoint and UI action | 16, 39 |
| S7-05-04 | Finance CSV export | Finance summary export endpoint and UI action | 16, 39 |
| S7-06-01 | OpenAPI endpoints | OpenAPI paths/schemas for reporting endpoints and exports | 16 |
| S7-06-02 | Generated FE client | Web generated client updated and consumed by reporting screens | 16 |
| S7-06-03 | Contract check | OpenAPI validate/contract/generate pass on dev server | 16 |
| S7-07-01 | Permission regression | Reporting view/export and finance report permission paths tested | 19, 24 |
| S7-07-02 | Inventory report smoke | Inventory snapshot report and CSV smoke coverage | 24 |
| S7-07-03 | Operations report smoke | Operations daily report and CSV smoke coverage | 24 |
| S7-07-04 | Finance report smoke | Finance summary report and CSV smoke coverage | 24 |
| S7-08-01 | Sprint 7 release evidence | Changelog with merged PRs, dev verification, CI status, migration status, release/tag status | 50 |

---

## 11. Definition of Done

For each code task:

```text
1. Code is scoped to the task.
2. Reports are read-only and do not mutate operational records.
3. Money, quantity, and rate fields follow file 40 decimal string rules.
4. Backend tests pass for touched services/handlers.
5. Web test/typecheck/build pass for touched UI areas.
6. OpenAPI validate/contract/generate pass when API contracts change.
7. Dev server build/test is used as the practical gate while GitHub Actions is blocked.
8. PR includes manual self-review comment.
9. Runtime changes are deployed to dev server after merge.
10. Any remaining unverified release gate is called out explicitly.
```

Sprint 7 completion requires:

```text
1. All S7 tasks merged to main.
2. Dev server smoke passes.
3. OpenAPI contract/generation passes.
4. Backend and web checks pass on dev server.
5. GitHub CI rerun is green after billing/spending-limit blocker is cleared.
6. Migration apply/rollback is verified on PostgreSQL 16 when migrations are added.
7. Tag v0.7.0-reporting-inventory-operations-dashboard is created only after release gates are green.
```
