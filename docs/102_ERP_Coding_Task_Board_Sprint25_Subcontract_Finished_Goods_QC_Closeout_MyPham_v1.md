# 102_ERP_Coding_Task_Board_Sprint25_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 25 - Subcontract finished goods QC closeout traceability
Document role: Coding task board
Version: v1
Date: 2026-05-06
Status: Implementation branch verified locally; PR/CI/merge/dev smoke pending

---

## 1. Goal

Sprint 25 closes the gap after Sprint 24:

```text
Production Plan
-> posted material issue readiness
-> Subcontract Order
-> factory production
-> finished goods receipt to QC hold
-> QC accept / reject
-> final payment readiness / closeout
```

Existing runtime already supports subcontract finished-goods receiving, QC accept, partial accept, factory claim, and final payment readiness. Sprint 25 should not rebuild that runtime. The missing work is first-class traceability from the source Production Plan to the related Subcontract Order and back to the plan detail.

---

## 2. Scope

Implement:

```text
1. Store source production plan id/no on Subcontract Order.
2. Expose source production plan fields in API request/response and OpenAPI.
3. Filter/list subcontract orders by source production plan id.
4. Send source fields when Production Plan creates a Subcontract Order.
5. Production Plan detail shows related subcontract order status, receipt, QC, accepted/rejected qty, and final payment readiness.
6. Production Plan worklist opens /subcontract with source production plan context.
7. Subcontract list honors source production plan query params from deep links.
8. Document the Sprint 25 flow and evidence.
```

Do not implement:

```text
1. Internal MES/work-center production.
2. Factory dispatch/export channel.
3. Costing or GL posting.
4. New QC module logic beyond existing subcontract accept/reject/claim endpoints.
5. Release tag v0.25.
```

---

## 3. Acceptance Criteria

Backend:

```text
- POST /api/v1/subcontract-orders accepts source_production_plan_id and source_production_plan_no.
- GET /api/v1/subcontract-orders returns source production fields.
- GET /api/v1/subcontract-orders?source_production_plan_id=... only returns matching orders.
- Search can find source production plan no.
- PostgreSQL persistence keeps source production plan fields.
```

Frontend:

```text
- buildSubcontractOrderFromProductionPlan sends sourceProductionPlanId and sourceProductionPlanNo.
- Production Plan worklist opens Subcontract with source production plan filter.
- Production Plan detail lists related subcontract orders and closeout state.
- Subcontract page reads source_production_plan_id/search query params and selects the matching order.
```

Verification:

```text
- Targeted backend tests pass.
- Targeted frontend tests pass.
- OpenAPI lint/contract checks pass.
- Web build/lint pass before PR.
- Dev deploy and browser smoke cover production plan detail -> subcontract deep link.
```

---

## 4. Status Log

```text
2026-05-06:
- Sprint opened on branch codex/s25-subcontract-fg-receiving-qc.
- Existing FG receipt/QC runtime confirmed present.
- Implementation focus set to Production Plan -> Subcontract traceability and closeout visibility.
- Source production plan fields, API/OpenAPI contract, PostgreSQL migration, Production Plan detail closeout visibility, and Subcontract deep-link filtering implemented on branch.
- Local backend/web/OpenAPI verification completed; GitHub CI, manual merge, dev deploy, and browser smoke remain pending.
```

---

## 5. Tag Status

```text
No v0.25 tag should be created unless a separate release checkpoint is requested.
Tag status: hold.
```
