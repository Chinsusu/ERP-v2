# 101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 24 - Production Material Issue Readiness
Document role: Changelog and evidence record
Version: v1
Date: 2026-05-05
Status: Completed on dev; release tag hold

---

## 1. Summary

Sprint 24 connects production-plan material demand to first-class Warehouse Issue Note evidence.

Implemented behavior:

```text
Production Plan
-> material demand line
-> source-linked Warehouse Issue Note draft for issue-ready stock
-> Warehouse Issue lifecycle submit / approve / post
-> posted issue quantity rolls back into production-plan readiness
-> subcontract creation remains blocked until required material is issued
```

Shortage is still handled by Purchase Request, PO, Receiving, and QC before warehouse issue.

---

## 2. PR Evidence

| PR | Merge commit | Scope | Status |
| --- | --- | --- | --- |
| #585 Open Sprint 24 production material issue readiness docs | `7415e125` | Task board and flow design | Merged |
| #586 Add production material issue readiness runtime | `9e28c05e` | Backend, frontend, OpenAPI, tests, docs status | Merged |
| #587 Use same-origin API base for dev web | `114105b2` | Dev web API-base fix found during browser smoke | Merged |

---

## 3. Runtime Changes

Backend:

```text
- Production plan lines expose issued_qty, remaining_issue_qty, issue_status, and warehouse_issues.
- Posted Warehouse Issue lines linked by source_document_line_id roll up to production plan demand lines.
- POST /api/v1/production-plans/{production_plan_id}/warehouse-issues creates source-linked Warehouse Issue drafts.
- Shortage and not-ready material lines are blocked from issue creation.
- Subcontract creation from production plan is gated by posted material issue readiness.
- Warehouse Issue line source_document_line_id is carried through domain, app service, handler, API, OpenAPI, and web client.
```

Frontend:

```text
- Production plan detail shows issue readiness per material line.
- Material table shows required, available, issued, remaining, shortage, issue status, and action.
- Ready/partial lines can create Warehouse Issue Note drafts from the production plan.
- Production plan worklist adds the Warehouse Issue step before subcontract order creation.
- Subcontract action remains blocked until stock-managed material lines are issued.
```

Dev runtime:

```text
- Dev web API base now supports browser access through the dev server host instead of browser-local localhost.
```

---

## 4. Verification

Local/remote targeted backend verification:

```text
go test ./internal/modules/production/application ./internal/modules/inventory/application ./cmd/api -run 'TestProductionPlanServiceCreatesWarehouseIssueFromReadyDemandLineAndRollsUpPostedIssue|TestProductionPlanServiceRejectsWarehouseIssueFromShortageLine|TestProductionPlanHandlersCreateWarehouseIssueFromPlanLine|TestWarehouseIssueServicePostsWarehouseIssueMovement|TestWarehouseIssueHandlersCreateAndPost'
```

Full backend verification:

```text
go test ./...
go vet ./...
```

Frontend verification:

```text
pnpm --filter web test -- src/modules/production-planning/services/productionPlanService.test.ts src/modules/production-planning/services/productionPlanWorklist.test.ts src/modules/production-planning/services/productionPlanWorkflowContext.test.ts src/modules/production-planning/services/productionPlanNextActions.test.ts src/modules/inventory/services/warehouseDocumentService.test.ts
pnpm --filter web test
pnpm --filter web lint
pnpm --filter web build
```

OpenAPI verification:

```text
pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml
node packages/openapi/contract-check.mjs
```

Notes:

```text
OpenAPI lint exited 0 with the existing proprietary-license warning.
OpenAPI contract check passed with 86 routes and 46 envelopes.
```

GitHub CI:

```text
PR #586 passed: api, web, openapi, e2e, required-api, required-web, required-openapi, required-migration.
PR #587 passed: required-api, required-web, required-openapi, required-migration.
```

Dev deploy and smoke:

```text
./infra/scripts/deploy-dev-staging.sh dev passed.
Full ERP dev smoke passed.
Browser smoke passed for login, /production, and /production/plans/{plan_id}.
```

Screenshot evidence:

```text
output/playwright/s24-production-list.png
output/playwright/s24-production-detail.png
```

---

## 5. Known Limits

```text
1. Sprint 24 does not calculate costing.
2. Sprint 24 does not create finished goods receipt.
3. Sprint 24 does not dispatch factory orders externally.
4. Sprint 24 does not implement two-step in-transit material issue.
5. Sprint 24 does not implement advanced MRP allocation.
6. Waiver remains documented follow-up scope; current runtime requires posted issue readiness.
7. Warehouse Issue detail deep-linking from every production-plan task row remains a follow-up UX hardening item if users need direct document drilldown from the worklist.
```

---

## 6. Tag Status

```text
No v0.24.0-production-material-issue-readiness tag has been created.
Tag status: hold.
Reason: Sprint 24 is completed on dev, but no release checkpoint tag was requested.
```

---

## 7. Next Candidate Work

Recommended next sprint direction:

```text
1. Improve direct traceability links from production plan tasks to Warehouse Issue detail.
2. Add finished goods receipt and inbound QC from subcontract output.
3. Add factory dispatch/export for subcontract execution.
4. Add costing input from posted material issue evidence.
```
