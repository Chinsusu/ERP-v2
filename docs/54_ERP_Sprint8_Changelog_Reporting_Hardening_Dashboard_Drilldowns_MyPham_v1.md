# 54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 8 - Reporting hardening / persisted reporting stores / dashboard drilldowns
Date: 2026-04-30
Status: Dev/main merged; production tag hold until GitHub Actions billing blocker is cleared

---

## 1. Summary

Sprint 8 hardened Reporting v1 and dashboard drilldowns:

```text
Report rows
-> shared source reference DTO
-> inventory item / batch / stock-state sources
-> operations receiving / QC / outbound / returns / stock count / subcontract sources
-> finance AR / AP / COD / cash / payment approval sources

Reporting UI
-> filters survive URL refresh/navigation
-> shared loading / empty / error / unavailable states
-> CSV export feedback and stable filenames
-> dashboard cards open matching reports with preserved filters

Contracts and gates
-> OpenAPI documents source references and CSV headers
-> generated web schema is refreshed
-> report access, drilldown, and smoke coverage is strengthened
```

Sprint 8 kept the reporting guardrails:

```text
- Reports and dashboard drilldowns are read-only.
- Report endpoints do not mutate stock, AR/AP, COD, cash, QC, orders, receiving, returns, or subcontract records.
- Drilldown links point to existing source records or filtered work queues.
- Missing source records degrade through unavailable source references.
- Money and quantity values remain decimal strings.
- VND, vi-VN, and Asia/Ho_Chi_Minh rules remain active.
- Finance report drilldowns remain behind reports:view and reports:finance:view.
- CSV exports keep stable headers and machine-readable values.
```

## 2. Merged PRs

Planning:

```text
#349 docs(S8-00-00): add sprint 8 task board
```

Source reference model and report drilldowns:

```text
#350 feat(S8-01-01): add report source reference model
#351 feat(S8-01-02): add inventory source references
#352 feat(S8-01-03): link operations report sources
#353 feat(S8-01-04): add finance source references
```

Reporting store/source hardening:

```text
#354 feat(S8-02-01): back operations reports with domain stores
#355 feat(S8-02-02): complete inventory report source contexts
#356 feat(S8-02-03): add auditable COD discrepancy sources
```

Report URL state, shared states, and export UX:

```text
#357 feat(S8-03-01): preserve reporting filters in URLs
#358 feat(S8-03-02): add shared report states
#359 feat(S8-03-03): harden report export UX
```

Dashboard report entry points:

```text
#360 feat(S8-04-01): add dashboard report entry points
#361 feat(S8-04-02): add warehouse signal report drilldowns
#362 feat(S8-04-03): add finance dashboard report drilldowns
```

OpenAPI and generated client:

```text
#363 docs(S8-05-01): document report source contracts
#364 chore(S8-05-02): refresh generated api schema
```

Regression and smoke gates:

```text
#365 test(S8-06-01): cover report access gates
#366 test(S8-06-02): cover inventory report smoke refs
#367 test(S8-06-03): cover operations report smoke refs
#368 test(S8-06-04): cover finance report smoke refs
```

## 3. Verification

Backend verification run on a clean dev-server clone of `main`:

```text
go test ./...
go vet ./...
```

Focused Sprint 8 backend checks run during task PRs:

```text
go test ./cmd/api -run PermissionRegression -count=1
go test ./cmd/api -run TestInventorySnapshotReportSmoke -count=1
go test ./cmd/api -run TestOperationsDailyReportSmoke -count=1
go test ./cmd/api -run TestFinanceSummaryReportSmoke -count=1
```

OpenAPI verification run on a clean dev-server clone of `main`:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --exit-code apps/web/src/shared/api/generated/schema.ts
```

Latest observed OpenAPI contract result:

```text
Sprint 4/5/6/7 OpenAPI contract check passed: 70 routes and 38 envelopes.
```

`pnpm openapi:validate` passed with the existing license warning:

```text
License object should contain one of the fields: url, identifier.
```

Frontend verification run on a clean dev-server clone of `main`:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
```

Latest observed frontend test result:

```text
35 test files passed.
224 tests passed.
```

The first combined frontend/OpenAPI verification attempt hit dev-server disk pressure while pnpm was downloading Redocly dlx packages:

```text
ENOSPC: no space left on device
```

Temporary verification clones/caches were cleaned, then the checks above were rerun successfully.

Live dev report endpoint smoke after final Sprint 8 runtime deploy:

```text
inventory_json   200 2448
inventory_csv    200 438
operations_json  200 451
operations_csv   200 140
finance_json     200 5909
finance_csv      200 856
```

Local migration diff check:

```text
git diff --name-only 97a533503691d3d52b05ac9bfefad547ead19486..HEAD -- apps/api/migrations infra/migrations db migrations
```

Result:

```text
No migration files changed in Sprint 8.
```

Local diff verification for this changelog PR:

```text
git diff --check
```

## 4. Dev Deployment Status

Dev server deployment was run after the final Sprint 8 runtime/test merge set.

Latest deployed `main` commit:

```text
82c9bf13
```

Latest deploy result:

```text
Dev source build completed for api, web, and worker.
Database seed/migration step completed.
Internal smoke passed.
Host smoke passed.
ERP dev deployment finished.
```

Latest API health timestamp observed during deploy:

```text
2026-04-30T19:26:43Z
```

The GHCR image pull warnings are still expected in the current dev setup; the deploy script builds service images from source and smoke checks pass.

## 5. Release Gate Status

Green:

```text
- All Sprint 8 task PRs are merged to main.
- API tests pass on the dev-server verification clone.
- API vet passes on the dev-server verification clone.
- Web tests, typecheck, and build pass on the dev-server verification clone.
- OpenAPI validate, contract, and generate pass on the dev-server verification clone.
- Generated web OpenAPI schema is stable after generation.
- Dev deployment and smoke pass.
- Live report JSON and CSV endpoints return HTTP 200 on dev.
- Sprint 8 added no database migrations.
```

Hold:

```text
- GitHub Actions cloud CI is still blocked by billing/spending-limit behavior.
- PR checks fail before logs are available; gh run view --log-failed returns log not found.
- Production tag v0.8.0-reporting-hardening-dashboard-drilldowns is not created.
```

Tag command to run only after GitHub Actions is unblocked and full CI is green:

```bash
git checkout main
git pull --ff-only origin main
git tag v0.8.0-reporting-hardening-dashboard-drilldowns
git push origin v0.8.0-reporting-hardening-dashboard-drilldowns
```

## 6. Known Notes

```text
- The OpenAPI contract check script name/output still says Sprint 4/5/6/7 even though Sprint 8 routes continue to pass the current route/envelope gate.
- Finance CSV remains an aggregate summary CSV; detailed finance source references are available in the JSON response.
- Sprint 5, Sprint 6, Sprint 7, and Sprint 8 production tags remain on hold until GitHub Actions can run green.
- Reporting hardening remains operational reporting, not a full BI warehouse or accounting reporting engine.
```

## 7. Next Step

```text
Clear the GitHub Actions billing/spending-limit blocker.
Rerun full GitHub CI on main.
Create held production tags only after cloud CI is green.
Plan the next sprint from the latest main state.
```
