# 122_ERP_Sprint31_Changelog_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 31
Change type: Runtime UI bridge, tracker/timeline link cleanup, receipt helper tests
Version: v1
Date: 2026-05-06
Status: Completed and merged

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

Runtime PR:

```text
PR: #601 Add factory finished goods receipt gate
Merge commit: 7b7952fb
Commit: ea70a461 Add factory finished goods receipt gate
```

Local verification:

```text
Targeted Vitest:
- subcontractFactoryFinishedGoodsReceipt
- subcontractFactoryExecutionTracker
- subcontractOrderTimeline
Status: pass locally on 2026-05-06

Full web test:
- vitest run --passWithNoTests
Status: pass locally on 2026-05-06

Web typecheck:
- tsc --noEmit
Status: pass locally on 2026-05-06

Web build:
- next build
Status: pass locally on 2026-05-06

API:
- go test ./...
- go vet ./...
Status: pass locally on 2026-05-06

OpenAPI:
- node packages/openapi/contract-check.mjs
Status: pass locally on 2026-05-06
```

GitHub CI:

```text
e2e: pass
required-api: pass
required-web: pass
required-openapi: pass
required-migration: pass
web: pass
```

Dev deploy and smoke:

```text
Dev deploy: ./infra/scripts/deploy-dev-staging.sh dev passed on 2026-05-06
Full dev smoke: passed
Browser smoke: /production/factory-orders/sco-s16-08-03-smoke-0064#factory-finished-goods-receipt passed
Browser smoke asserted:
- receipt section exists
- hash route resolves
- receipt controls render
- QC hold action button exists
- tracker/timeline link to #factory-finished-goods-receipt exists
Screenshot evidence: output/playwright/s31-factory-finished-goods-receipt.png
```

---

## 5. Release Tag

No `v0.31` tag is created by this sprint branch.

Sprint 31 is a Phase 1 runtime milestone, but release tagging remains held until target staging/pilot readiness evidence is explicitly approved.
