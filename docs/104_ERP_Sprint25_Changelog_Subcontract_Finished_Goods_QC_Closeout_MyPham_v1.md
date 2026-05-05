# 104_ERP_Sprint25_Changelog_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 25 - Subcontract finished goods QC closeout traceability
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Implementation branch verified locally; PR, CI, merge, dev deploy, and browser smoke pending

---

## 1. Summary

Sprint 25 connects the production planning workspace to external factory execution closeout.

The subcontract runtime already had finished goods receipt, QC accept, partial accept, factory claim, final payment readiness, and supplier payable handoff behavior. Sprint 25 adds the missing source traceability and closeout visibility:

```text
Production Plan
-> source-linked Subcontract Order
-> finished goods receipt / QC / claim / payment closeout visibility
-> Production Plan detail can show related subcontract status
```

This sprint does not implement internal MES/work-center production, factory dispatch/export, costing, or GL posting.

---

## 2. Runtime Changes

Backend/API:

```text
- Subcontract Order stores source_production_plan_id and source_production_plan_no.
- POST/PATCH /api/v1/subcontract-orders accepts source production plan fields.
- GET /api/v1/subcontract-orders returns source production plan fields.
- GET /api/v1/subcontract-orders?source_production_plan_id=... filters linked subcontract orders.
- Subcontract order search includes the source production plan number.
- PostgreSQL persistence includes source production plan columns and index.
- OpenAPI documents the new request, response, and filter contract.
```

Frontend:

```text
- buildSubcontractOrderFromProductionPlan sends sourceProductionPlanId/sourceProductionPlanNo.
- Production Plan worklist opens /subcontract with source production plan context.
- /subcontract reads source_production_plan_id/search query params from deep links.
- Production Plan detail lists related subcontract orders with receipt/QC/factory claim/final payment closeout state.
```

Database:

```text
- Migration 000043_add_subcontract_order_source_production_plan adds source_production_plan_ref and source_production_plan_no to subcontract.subcontract_orders.
- Migration adds an org/source production plan index for filtered subcontract lookup.
```

---

## 3. Verification

Local verification on the implementation branch:

```text
- gofmt check for changed Go files: pass
- go vet ./...: pass
- go test ./...: pass
- web tsc --noEmit: pass
- web vitest full suite: 53 files / 304 tests pass
- web Next production build: pass
- OpenAPI contract check: pass
- Redocly OpenAPI lint: valid, with pre-existing info/license warning
- git diff --check: pass
```

Local command notes:

```text
- Toolcache Go was used: D:\toolcache\go1.24.2\go\bin.
- Toolcache Node was used: D:\toolcache\node-v22.22.2-win-x64\node.exe.
- pnpm .cmd/vitest shim execution hit Windows Access denied, so vitest, tsc, Next, and Redocly were run directly through node.
```

---

## 4. Evidence Pending

These must be filled before Sprint 25 is marked completed on main:

```text
- PR number: pending
- GitHub CI: pending
- Manual diff review: pending
- Merge commit: pending
- Dev deploy: pending
- Dev full smoke: pending
- Browser smoke: pending
```

Required browser smoke after merge/deploy:

```text
1. Login to dev.
2. Open /production.
3. Open a production plan detail.
4. Confirm related subcontract closeout table is visible.
5. Use "Mo gia cong" link.
6. Confirm /subcontract opens filtered by the selected production plan context.
```

---

## 5. Tag Status

```text
No v0.25.0-subcontract-finished-goods-qc-closeout tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested after CI, merge, dev deploy, and smoke evidence are recorded.
```
