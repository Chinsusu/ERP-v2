# 116_ERP_Sprint29_Changelog_Factory_Material_Handover_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 29 - Factory Material Handover
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Implemented on branch; PR, CI, merge, dev deploy, and browser smoke pending

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
- Added "Bàn giao vật tư cho nhà máy" to /production/factory-orders/:orderId.
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
- apps/web: tsc --noEmit passed
```

Pending before closeout:

```text
- Full web test/build
- API regression checks
- OpenAPI contract check
- GitHub CI
- Dev deploy
- Browser smoke for /production/factory-orders/:orderId#factory-material-handover
```

---

## 4. Evidence

```text
PR number: pending
Merge commit: pending
GitHub CI: pending
Dev deploy: pending
Full dev smoke: pending
Browser smoke: pending
Screenshot: pending
```

---

## 5. Tag Status

```text
No v0.29.0-factory-material-handover tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
