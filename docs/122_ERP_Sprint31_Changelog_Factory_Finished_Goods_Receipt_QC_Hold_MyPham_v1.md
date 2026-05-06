# 122_ERP_Sprint31_Changelog_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 31
Change type: Runtime UI bridge, tracker/timeline link cleanup, receipt helper tests
Version: v1
Date: 2026-05-06
Status: Draft; PR, CI, merge, dev deploy, and browser smoke pending

---

## 1. Summary

Sprint 31 adds the production-facing finished goods receipt step for external-factory production.

The user-facing route is:

```text
/production/factory-orders/:orderId#factory-finished-goods-receipt
```

The section receives finished goods from the factory into QC hold using the existing subcontract finished-goods receipt runtime. It records delivery note, receiver, target QC hold, quantity, batch/lot, expiry date, packaging condition, evidence filename, receipt result, and updated order state.

---

## 2. Changed Runtime Surface

- Added finished-goods receipt helper/gate for factory order detail.
- Added receipt section on `/production/factory-orders/:orderId`.
- Reused existing `receiveSubcontractFinishedGoods` service/API behavior.
- Updated factory execution tracker finished-goods action to point to the production factory order detail section.
- Updated factory order timeline finished-goods action to point to the production factory order detail section.

---

## 3. Guardrails

- Receipt is blocked before mass production starts.
- Receipt goes to QC hold only.
- Receipt does not mark QC pass.
- Receipt does not create available stock.
- Receipt does not release final payment.
- Receipt does not add email/Zalo/factory portal/API delivery.
- Receipt does not introduce internal MES/work-center production.

---

## 4. Verification Evidence

Current branch evidence:

```text
Targeted Vitest:
- subcontractFactoryFinishedGoodsReceipt
- subcontractFactoryExecutionTracker
- subcontractOrderTimeline
Status: pass locally on 2026-05-06

Web typecheck:
- tsc --noEmit
Status: pass locally on 2026-05-06
```

Pending before completion:

```text
Full web test/build
API test/vet if code-change gate requires it
OpenAPI contract validation
GitHub CI
Manual diff review
Manual merge
Dev deploy
Full dev smoke
Browser smoke for #factory-finished-goods-receipt
Screenshot evidence
```

---

## 5. Release Tag

No `v0.31` tag is created by this sprint branch.

Sprint 31 is a Phase 1 runtime milestone, but release tagging remains held until target staging/pilot readiness evidence is explicitly approved.
