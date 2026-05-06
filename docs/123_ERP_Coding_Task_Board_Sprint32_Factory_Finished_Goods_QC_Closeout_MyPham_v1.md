# 123_ERP_Coding_Task_Board_Sprint32_Factory_Finished_Goods_QC_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 32
Scope: Factory finished goods QC closeout
Version: v1
Date: 2026-05-06
Status: In progress

---

## 1. Goal

Sprint 32 moves the external-factory production flow one step after Sprint 31.

After finished goods have been received from the factory into QC hold, production/QC users must be able to close the QC decision directly from:

```text
/production/factory-orders/:orderId#factory-finished-goods-qc-closeout
```

The QC decision must keep the core inventory rule explicit:

```text
Factory receipt -> QC hold
QC PASS -> available stock
QC PARTIAL -> accepted stock plus factory claim for rejected quantity
QC FAIL -> factory claim / blocked closeout
```

---

## 2. Business Rule

Finished goods from an external factory are not usable at receipt time.

```text
Only QC accepted quantity can become available stock.
Rejected quantity must remain blocked by factory claim evidence.
Final payment readiness stays separate and must not be auto-released by the QC screen.
```

---

## 3. Runtime Surface

Primary user-facing route:

```text
/production/factory-orders/:orderId
```

New section:

```text
#factory-finished-goods-qc-closeout
```

Existing runtime API/service to reuse:

```text
POST /api/v1/subcontract-orders/{id}/accept
POST /api/v1/subcontract-orders/{id}/partial-accept
POST /api/v1/subcontract-orders/{id}/report-factory-defect
```

---

## 4. Acceptance Criteria

- QC closeout is blocked before finished goods are received into QC hold.
- QC closeout is enabled for `finished_goods_received` and `qc_in_progress`.
- Full pass posts accepted quantity for the remaining received quantity; backend resolves stored receipt movements.
- Partial pass posts accepted quantity and opens a factory claim for rejected quantity.
- Full fail opens a factory claim and does not release available stock.
- The section shows received quantity, accepted quantity, rejected quantity, remaining QC hold, and current closeout state.
- When the latest receipt is loaded in the page session, the section also shows receipt traceability data such as receipt number, delivery note, QC warehouse/location, batch/lot, expiry, and evidence.
- Tracker and timeline QC actions point to `#factory-finished-goods-qc-closeout`.
- The existing `/subcontract` technical execution route remains available but is not the primary user-facing entrypoint for this step.

---

## 5. Verification Plan

Local checks before PR where toolchain is available:

```text
Targeted Vitest:
- subcontractFactoryFinishedGoodsQcCloseout
- subcontractOrderService
- subcontractFactoryExecutionTracker
- subcontractOrderTimeline

Typecheck/build:
- apps/web tsc --noEmit
- apps/web next build

Full code-change checks:
- web tests
- API tests/vet
- OpenAPI contract
```

Post-merge dev verification:

```text
1. Deploy dev with ./infra/scripts/deploy-dev-staging.sh dev.
2. Run full dev smoke.
3. Browser smoke /production/factory-orders/:orderId#factory-finished-goods-qc-closeout.
4. Confirm QC closeout section is visible and linked from tracker/timeline.
5. Confirm at least the section renders for a received factory order; run mutating QC decision smoke only on safe smoke-created data.
```

---

## 6. Out Of Scope

- Email, Zalo, factory portal, or API dispatch.
- Internal MES, routing, work centers, labor costing, and internal shop-floor tracking.
- Automatic final payment release.
- Supplier invoice or AP payment changes.
- Broad QC module redesign.
- Release tag `v0.32` unless separately requested.

---

## 7. Completion Evidence

```text
Runtime PR: pending
Merge commit: pending
GitHub CI: pending
Dev deploy: pending
Full dev smoke: pending
Browser smoke: pending
Release tag: no v0.32 tag planned
```
