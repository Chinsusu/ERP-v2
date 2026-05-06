# 120_ERP_Coding_Task_Board_Sprint31_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 31
Scope: Factory finished goods receipt to QC hold
Version: v1
Date: 2026-05-06
Status: In progress on PR branch

---

## 1. Goal

Sprint 31 moves the external-factory production flow one step after Sprint 30.

After a factory order has started mass production, production/warehouse users must be able to receive finished goods from the factory into QC hold directly from:

```text
/production/factory-orders/:orderId#factory-finished-goods-receipt
```

The receipt records delivery note, warehouse, QC-hold location, received quantity, batch/lot, expiry date, packaging condition, evidence, stock movement, and updated order received quantity.

---

## 2. Business Rule

Finished goods from an external factory do not become available stock at receipt time.

```text
Factory receipt -> QC hold
QC PASS -> available stock
QC FAIL -> factory claim / blocked closeout
```

Sprint 31 only covers factory receipt into QC hold. It does not implement QC pass/fail, claim closeout, final payment release, email/Zalo factory delivery, portal/API integration, or internal MES/work-center production.

---

## 3. Runtime Surface

Primary user-facing route:

```text
/production/factory-orders/:orderId
```

New section:

```text
#factory-finished-goods-receipt
```

Existing runtime API/service reused:

```text
receiveSubcontractFinishedGoods
POST /api/v1/subcontract-orders/{id}/receive-finished-goods
```

Existing prototype fallback remains only for allowed non-production-like local fallback mode.

---

## 4. Acceptance Criteria

- Factory finished goods receipt is blocked before `mass_production_started`.
- Receipt is enabled at `mass_production_started`.
- Receipt supports partial receipt until planned quantity is fully received.
- Receipt blocks over-receipt at the UI helper/runtime boundary.
- Receipt requires delivery note, receiver, batch/lot, expiry date, and positive quantity.
- Receipt target is QC hold, not available stock.
- Receipt result updates the in-page order state and shows latest receipt evidence.
- Factory execution tracker links finished goods receipt to `/production/factory-orders/:orderId#factory-finished-goods-receipt`.
- Factory order timeline links finished goods receipt to `/production/factory-orders/:orderId#factory-finished-goods-receipt`.
- Existing `/subcontract` technical execution route remains available but is not the primary user-facing entrypoint for this step.

---

## 5. Verification Plan

Local checks before PR:

```text
Targeted Vitest:
- subcontractFactoryFinishedGoodsReceipt
- subcontractFactoryExecutionTracker
- subcontractOrderTimeline

Typecheck:
- apps/web tsc --noEmit

Full code-change checks:
- web tests
- web build
- API tests/vet
- OpenAPI contract
```

Post-merge dev verification:

```text
1. Deploy dev with ./infra/scripts/deploy-dev-staging.sh dev.
2. Run full dev smoke.
3. Browser smoke /production/factory-orders/:orderId#factory-finished-goods-receipt.
4. Confirm receipt section is visible and linked from tracker/timeline.
5. Confirm receiving into QC hold updates the order state or use an existing smoke-created order where safe.
```

---

## 6. Out Of Scope

- QC pass/fail execution UI.
- Available stock posting after QC pass.
- Factory claim UI changes.
- Final payment release.
- Email/Zalo/factory portal/API dispatch.
- Internal MES, routing, work centers, labor costing, and internal production shop-floor tracking.
- Schema or OpenAPI redesign beyond existing receipt contract reuse.
