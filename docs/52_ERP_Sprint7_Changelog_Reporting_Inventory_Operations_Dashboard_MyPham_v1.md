# 52_ERP_Sprint7_Changelog_Reporting_Inventory_Operations_Dashboard_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 7 - Reporting v1 / Inventory / Operations Dashboard
Date: 2026-04-30
Status: Dev/main merged; production tag hold until GitHub Actions billing blocker is cleared

---

## 1. Summary

Sprint 7 completed reporting v1:

```text
Inventory snapshot
-> available / reserved / quarantine / blocked quantities
-> low-stock and expiry warning visibility
-> batch/location rows
-> CSV export

Operations daily
-> inbound, QC, outbound, returns, stock count, subcontract signals
-> daily KPI cards and exception rows
-> CSV export

Finance summary
-> AR / AP / COD / cash summary
-> aging and discrepancy summary visibility
-> finance permission guard
-> CSV export
```

Sprint 7 kept the reporting guardrails:

```text
- Reports are read-only.
- Reports do not mutate stock, orders, QC, COD, AR/AP, or cash.
- Money and quantity values remain decimal strings at API/UI boundaries.
- VND, vi-VN, and Asia/Ho_Chi_Minh rules remain active.
- Finance reports require finance-report permission.
- CSV exports use stable headers.
- OpenAPI and generated web client are updated.
```

## 2. Merged PRs

Planning and foundation:

```text
#325 docs(S7-00-00): add reporting sprint task board
#326 feat(S7-01-01): add reporting permissions menu
#327 feat(S7-01-02): add reporting filter export foundation
```

Inventory snapshot:

```text
#328 feat(S7-02-01): add inventory snapshot query model
#329 feat(S7-02-02): add inventory snapshot report API
#330 feat(S7-02-03): add inventory snapshot reporting UI
```

Operations daily:

```text
#331 feat(S7-03-01): add operations daily query model
#332 feat(S7-03-02): add operations daily report API
#333 feat(S7-03-03): add operations daily report UI
```

Finance summary:

```text
#334 feat(S7-04-01): add finance summary query model
#335 feat(S7-04-02): add finance summary report API
#336 feat(S7-04-03): add finance summary reporting UI
```

CSV export:

```text
#337 feat(S7-05-02): add inventory snapshot CSV export
#338 feat(S7-05-03): add operations daily CSV export
#339 feat(S7-05-04): add finance summary CSV export
```

OpenAPI, generated client, and contract gate:

```text
#340 docs(S7-06-01): add reporting OpenAPI endpoints
#341 feat(S7-06-02): update generated reporting client
#342 test(S7-06-03): cover reporting OpenAPI contract
```

Regression and smoke gates:

```text
#343 test(S7-07-01): add reporting permission regression
#344 test(S7-07-02): add inventory report smoke
#345 test(S7-07-03): add operations report smoke
#346 test(S7-07-04): add finance report smoke
```

## 3. Verification

Backend verification run on a clean dev-server clone of `main`:

```text
go test ./...
go vet ./...
```

Focused Sprint 7 backend checks run during task PRs:

```text
go test ./cmd/api -run TestReportingPermissionRegression -count=1
go test ./cmd/api -run TestInventorySnapshotReportSmoke -count=1
go test ./cmd/api -run TestOperationsDailyReportSmoke -count=1
go test ./cmd/api -run TestFinanceSummaryReportSmoke -count=1
```

OpenAPI verification run on a clean dev-server clone of `main`:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --stat -- apps/web/src/shared/api/generated/schema.ts
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
34 test files passed.
216 tests passed.
```

Live dev report endpoint smoke after final deploy:

```text
200 2026 /api/v1/reports/inventory-snapshot?business_date=2026-04-30
200 783  /api/v1/reports/inventory-snapshot/export.csv?business_date=2026-04-30
200 4093 /api/v1/reports/operations-daily?business_date=2026-04-30
200 1615 /api/v1/reports/operations-daily/export.csv?business_date=2026-04-30
200 1280 /api/v1/reports/finance-summary?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08
200 795  /api/v1/reports/finance-summary/export.csv?from_date=2026-04-30&to_date=2026-05-08&business_date=2026-05-08
```

Local diff verification for this changelog PR:

```text
git diff --check
```

## 4. Dev Deployment Status

Dev server deployment was run after the final Sprint 7 merge set.

Latest deployed `main` commit:

```text
48dcd573
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
2026-04-30T09:38:23Z
```

The GHCR image pull warnings are still expected in the current dev setup; the deploy script builds the service images from source and smoke checks pass.

## 5. Release Gate Status

Green:

```text
- All Sprint 7 PRs are merged to main.
- API tests pass on the dev-server verification clone.
- API vet passes on the dev-server verification clone.
- Web tests, typecheck, and build pass on the dev-server verification clone.
- OpenAPI validate, contract, and generate pass on the dev-server verification clone.
- Generated web OpenAPI schema is stable after generation.
- Dev deployment and smoke pass.
- Live report JSON and CSV endpoints return HTTP 200 on dev.
```

Hold:

```text
- GitHub Actions cloud CI is still blocked by billing/spending-limit behavior.
- Multiple PR checks fail before logs are available; gh run view --log-failed returns log not found.
- Production tag v0.7.0-reporting-inventory-operations-dashboard is not created.
```

Migration gate:

```text
- Sprint 7 added no database migrations.
- PostgreSQL 16 apply/rollback verification is not required for Sprint 7 changes.
```

Tag command to run only after GitHub Actions is unblocked and full CI is green:

```bash
git checkout main
git pull --ff-only origin main
git tag v0.7.0-reporting-inventory-operations-dashboard
git push origin v0.7.0-reporting-inventory-operations-dashboard
```

## 6. Known Notes

```text
- Finance CSV currently includes COD discrepancy summary rows.
- Separate cod_discrepancy bucket rows are not emitted by the prototype finance summary data.
- Reporting v1 is operational reporting, not a full BI warehouse or accounting reporting engine.
- Sprint 5, Sprint 6, and Sprint 7 production tags remain on hold until GitHub Actions can run green.
```

## 7. Next Sprint

Recommended next sprint remains:

```text
Sprint 8 - Reporting hardening / persisted reporting stores / dashboard drilldowns
```

Do not start a production release tag until the CI blocker is resolved.
