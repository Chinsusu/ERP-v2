# 116_ERP_Sprint29_Changelog_Factory_Material_Handover_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 29 - Factory Material Handover
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Completed and merged; CI, dev deploy, full dev smoke, and browser smoke passed

---

## 1. Summary

Sprint 29 adds a production-facing material handover surface for external factory orders:

```text
Factory confirmed
-> deposit/payment condition satisfied
-> material handover on /production/factory-orders/:orderId
-> existing issue-materials runtime
-> order tracker/timeline updates from returned state
```

No email, Zalo, supplier portal, external factory API, new backend API, or internal MES behavior is included.

---

## 2. Runtime Changes

Frontend:

```text
- Added a factory material handover readiness/payload service.
- Added unit tests for handover readiness, deposit blocking, complete handover, and payload generation.
- Added the production-facing material handover section to /production/factory-orders/:orderId.
- Added source warehouse, receiver, contact, vehicle, evidence, note, issue quantity, batch/lot, and bin inputs.
- Added lot-required gating for pending lot-controlled material lines.
- Updated tracker and timeline material actions to link to #factory-material-handover instead of hidden /subcontract transfer.
- Kept the existing read-only material list for quick audit of planned/issued quantities.
```

Backend/API:

```text
- No new backend API or database migration is included.
- Existing issueSubcontractMaterials runtime drives transfer creation, stock movements, audit, and order status advancement.
```

---

## 3. Verification

Local branch verification:

```text
- apps/web: subcontractFactoryMaterialHandover.test.ts passed
- apps/web: subcontractFactoryExecutionTracker.test.ts passed
- apps/web: subcontractOrderTimeline.test.ts passed
- apps/web: vitest run --passWithNoTests passed, 56 files / 321 tests
- apps/web: tsc --noEmit passed
- apps/web: next build passed
- apps/api: go test ./... passed
- apps/api: go vet ./... passed
- packages/openapi: contract-check passed, 90 routes / 48 envelopes
- git diff --check passed
- git diff --cached --check passed before runtime commit
```

Remote/dev verification:

```text
- GitHub PR #597 required-api, required-web, required-openapi, required-migration, web, and e2e checks passed.
- PR #597 was manually merged into main at 7fd3b2d5.
- Dev deploy passed with ./infra/scripts/deploy-dev-staging.sh dev on 2026-05-06.
- Full dev smoke passed during deploy.
- Browser smoke passed on /production/factory-orders/sco-s16-02-01-1777715855392710950#factory-material-handover.
- Browser smoke verified the handover section DOM, tracker/timeline hash link, order visibility, and expected input/select/button controls.
```

---

## 4. Evidence

```text
Runtime branch: codex/s29-factory-material-handover
Runtime commit: c05813da
PR number: #597
Merge commit: 7fd3b2d5
GitHub CI: passed
Dev deploy: passed
Full dev smoke: passed
Browser smoke: passed
Screenshot: output/playwright/s29-factory-material-handover.png
```

---

## 5. Tag Status

```text
No v0.29.0-factory-material-handover tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
