# 53_ERP_Coding_Task_Board_Sprint8_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 8 - Reporting hardening / persisted reporting stores / dashboard drilldowns
Document role: Coding task board for Sprint 8 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL, Next.js frontend, OpenAPI, CSV export
Status: Ready for implementation; production tags remain on hold until GitHub Actions billing blocker is cleared

---

## 1. Sprint 8 Context

Sprint 7 delivered Reporting v1 for inventory snapshot, operations daily KPI, finance summary, and CSV exports.

Sprint 8 makes those reports more useful for daily operations:

```text
Report summary
-> source record references
-> dashboard drilldowns
-> stronger persisted backing for prototype report signals
-> stable filters and exports
-> regression coverage for drilldown and finance access
```

This sprint hardens the reporting layer. It must not become a full BI warehouse, custom report builder, accounting engine, or workflow mutation surface.

---

## 2. Sprint 8 Theme

```text
Reporting hardening + Dashboard drilldowns
```

Business reason:

```text
Sprint 7 answers "what is happening".
Sprint 8 should let managers move from a report signal to the source work item:
- Which stock row is low, held, quarantined, or near expiry.
- Which receiving, QC, return, stock count, pick/pack, or subcontract signal needs action.
- Which COD, AR, AP, or cash item explains a finance number.
- Which dashboard card can open the relevant report with the right filters.
```

---

## 3. Sprint 8 Goals

By the end of Sprint 8, the system must support:

```text
1. Add report source reference DTOs for drilldown rows.
2. Add inventory report drilldown links to source item, batch, warehouse, and stock state rows where available.
3. Add operations report drilldown links to source receiving, QC, outbound, return, stock count, and subcontract records.
4. Add finance report drilldown links to source AR, AP, COD, cash, and payment approval records.
5. Harden prototype-backed report signal stores where Sprint 7 left known gaps.
6. Persist COD discrepancy bucket rows or expose an auditable source for them.
7. Preserve filters in URLs so dashboards, reports, and refreshes keep the same business context.
8. Improve report empty, loading, error, and export feedback states.
9. Add dashboard cards and links that open reports with matching filters.
10. Update OpenAPI, generated frontend client, and contract checks for source references.
11. Add focused regression and smoke coverage for drilldowns, finance permissions, filters, and exports.
12. Keep every report endpoint read-only.
```

---

## 4. Sprint 8 Non-Goals

Sprint 8 does not include:

```text
- Full data warehouse or OLAP cube.
- Custom report builder.
- Scheduled email reports.
- PDF report generation.
- External BI integration.
- Tax/VAT reports.
- Full accounting financial statements.
- Rewriting Sprint 7 reports.
- Real-time websocket dashboards.
- Editing operational records from report screens.
- Row-level security beyond existing role permission gates.
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
codex/feature-S8-xx-yy-short-task-name
```

Recommended Sprint 8 release tag after completion:

```text
v0.8.0-reporting-hardening-dashboard-drilldowns
```

Production tags remain on hold while GitHub Actions is blocked by the billing/spending-limit issue.
Sprint 5, Sprint 6, and Sprint 7 production tags remain carry-forward blockers until CI can run green on GitHub and the release tags are created.

---

## 6. Sprint 8 Demo Script

### Case 1: Inventory drilldown

```text
1. Open Reporting -> Inventory Snapshot.
2. Select business date and warehouse.
3. Open a low stock, quarantine, or expiry warning row.
4. Confirm the drilldown opens the source item, batch, warehouse, or stock state context.
5. Export CSV and confirm source reference columns remain stable and machine-readable.
```

### Case 2: Operations drilldown

```text
1. Open Reporting -> Operations Daily.
2. Select a business date and warehouse.
3. Open inbound, QC, outbound, return, stock count, or subcontract signals.
4. Confirm each drilldown opens the source operational record or source filtered work queue.
5. Refresh the page and confirm filters remain in the URL.
```

### Case 3: Finance drilldown

```text
1. Open Reporting -> Finance Summary with a finance-capable user.
2. Select from/to date filters.
3. Open AR, AP, COD discrepancy, cash in, cash out, and payment approval rows.
4. Confirm non-finance users cannot access finance report drilldowns.
5. Export CSV and confirm decimal string values and source references are stable.
```

---

## 7. Sprint 8 Guardrails

These rules are non-negotiable:

```text
1. Reports and dashboard drilldowns are read-only.
2. No report endpoint may mutate stock, AR/AP, COD, cash, QC, orders, receiving, returns, or subcontract records.
3. Drilldown links must point to existing source records or existing filtered work queues.
4. Missing source records must degrade to a clear unavailable state, not a fabricated link.
5. Money, quantity, and rate values remain decimal strings in APIs.
6. VND, vi-VN, and Asia/Ho_Chi_Minh rules from file 40 still apply.
7. Inventory drilldowns must not mix available, reserved, quarantine, hold, and rejected stock.
8. Finance drilldowns must remain protected by finance/reporting permissions.
9. CSV exports must keep stable headers and machine-readable decimal strings.
10. Date filters must use business dates and inclusive ranges.
11. Dashboard cards may link to reports, but may not bypass report permission gates.
12. OpenAPI must be updated for every response shape change.
13. Sprint 8 is not production-ready until the GitHub CI blocker is cleared and release gates are green.
```

---

## 8. Dependency Map

```text
S8-00-00 Sprint 8 task board
  -> S8-01-01 report source reference model
  -> S8-00-01 Sprint 7 release gate carry-forward

S8-01-01 report source reference model
  -> S8-01-02 inventory drilldown links
  -> S8-01-03 operations drilldown links
  -> S8-01-04 finance drilldown links
  -> S8-05-01 OpenAPI source reference contract

S8-01-02 inventory drilldown links
  -> S8-04-01 dashboard report entry points
  -> S8-06-02 inventory drilldown smoke

S8-01-03 operations drilldown links
  -> S8-04-02 warehouse dashboard drilldowns
  -> S8-06-03 operations drilldown smoke

S8-01-04 finance drilldown links
  -> S8-04-03 finance dashboard drilldowns
  -> S8-06-04 finance drilldown smoke

S8-02-01 persist operations report signal adapters
  -> S8-03-02 shared report states
  -> S8-06-03 operations drilldown smoke

S8-02-02 inventory report source completeness
  -> S8-06-02 inventory drilldown smoke

S8-02-03 finance discrepancy source rows
  -> S8-06-04 finance drilldown smoke

S8-03-01 reporting filter URL state
  -> S8-04-01 dashboard report entry points
  -> S8-06-02 inventory drilldown smoke
  -> S8-06-03 operations drilldown smoke
  -> S8-06-04 finance drilldown smoke

S8-05-01 OpenAPI source reference contract
  -> S8-05-02 generated frontend client
  -> S8-06-01 report access regression

S8-07-01 Sprint 8 release evidence
  -> recommended release tag after CI is unblocked
```

---

## 9. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S8-00-00 | Sprint 8 task board | File 53 created, reviewed, merged to main | `52_ERP_Sprint7_Changelog_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` |
| S8-00-01 | Sprint 7 release gate carry-forward | CI blocker, held tags, and release gate status remain visible in Sprint 8 evidence | `52_ERP_Sprint7_Changelog_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` |
| S8-01-01 | Report source reference model | Shared backend/frontend DTO supports source entity type, id, label, href, and unavailable state | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S8-01-02 | Inventory drilldown links | Inventory snapshot rows expose source references for item, batch, warehouse, stock state, and warning context where available | `51_ERP_Coding_Task_Board_Sprint7_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` |
| S8-01-03 | Operations drilldown links | Operations daily rows expose receiving, QC, outbound, return, stock count, subcontract, or filtered queue links | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S8-01-04 | Finance drilldown links | Finance summary rows expose AR, AP, COD, cash, and approval source links behind finance permissions | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S8-02-01 | Persist operations report signal adapters | Prototype or fixture-backed operations signals are backed by domain stores where current code supports it | `44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1.md` |
| S8-02-02 | Inventory source completeness check | Report query covers available, reserved, quarantine, blocked, low stock, batch, and expiry source contexts consistently | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| S8-02-03 | Finance discrepancy source rows | COD discrepancy bucket rows have stable source records or auditable source references | `50_ERP_Sprint6_Changelog_Finance_Lite_COD_AR_AP_Core_MyPham_v1.md` |
| S8-03-01 | Reporting filter URL state | Inventory, operations, and finance report filters are encoded in URLs and survive refresh/navigation | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S8-03-02 | Shared report states | Report pages show consistent loading, empty, error, forbidden, and unavailable drilldown states | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S8-03-03 | Export UX hardening | CSV export buttons show filename/status and preserve current filters | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S8-04-01 | Dashboard report entry points | Dashboard cards open the matching report with date/warehouse/status filters | `51_ERP_Coding_Task_Board_Sprint7_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` |
| S8-04-02 | Warehouse dashboard drilldowns | Warehouse board links inbound, outbound, returns, stock count, QC, and subcontract signals to reporting views | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S8-04-03 | Finance dashboard drilldowns | Finance dashboard/report entry points link to finance summary without bypassing permissions | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S8-05-01 | OpenAPI source reference contract | OpenAPI documents source reference response shapes and CSV source columns | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S8-05-02 | Generated frontend client | Generated schema/client is refreshed and stable after generation | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S8-06-01 | Report access regression | Finance/reporting permission tests cover drilldown access and dashboard entry points | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S8-06-02 | Inventory drilldown smoke | Inventory report smoke covers source references, URL filters, and CSV source headers | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S8-06-03 | Operations drilldown smoke | Operations report smoke covers source references, URL filters, and read-only behavior | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S8-06-04 | Finance drilldown smoke | Finance report smoke covers permission gates, source references, and COD discrepancy rows | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S8-07-01 | Sprint 8 release evidence | Changelog captures PRs, verification, dev deploy status, CI blocker, and tag hold/release status | `52_ERP_Sprint7_Changelog_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` |

---

## 10. Verification Gates

Each implementation PR should run the smallest relevant checks plus broader checks when contracts or shared UI change.

Backend checks:

```text
go test ./...
go vet ./...
```

Frontend checks:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
```

OpenAPI checks when API shapes change:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --stat -- apps/web/src/shared/api/generated/schema.ts
```

Dev server checks when runtime behavior changes:

```text
deploy dev staging
API health smoke
targeted live report JSON/CSV smoke
```

GitHub Actions cloud CI must be rerun after the billing/spending-limit blocker is cleared. Production tags must remain on hold until CI is green.

---

## 11. Definition of Done

Sprint 8 is complete when:

```text
1. All task PRs are merged to main through manual review/merge.
2. Inventory, operations, and finance report drilldowns work on dev.
3. Dashboard report entry points preserve filters and permission gates.
4. API and CSV report shapes are documented and stable.
5. Generated frontend schema is stable after OpenAPI generation.
6. Backend, frontend, OpenAPI, and targeted smoke checks pass on the dev server.
7. Dev deployment is green after final merge.
8. Sprint 8 changelog is created with release evidence.
9. Production tag remains held if GitHub Actions is still blocked.
```
