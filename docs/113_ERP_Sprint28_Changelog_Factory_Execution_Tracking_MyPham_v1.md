# 113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 28 - Factory Execution Tracking
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Implemented on branch; PR, CI, merge, dev deploy, and browser smoke pending

---

## 1. Summary

Sprint 28 adds a production-facing execution tracker for external factory orders:

```text
Factory confirmed
-> deposit/payment condition
-> material handover
-> sample gate
-> mass production
-> finished goods receipt
-> QC closeout / factory claim
-> final payment readiness
```

No email, Zalo, supplier portal, external factory API, or internal MES behavior is included.

---

## 2. Runtime Changes

Frontend:

```text
- Added a factory execution tracker service.
- Added unit tests for current gate and blocking behavior.
- Added "Theo dõi thực thi nhà máy" to /production/factory-orders/:orderId.
- Added direct action links to existing dispatch, material handover, sample, inbound/QC, claim, and payment sections.
- Added hidden /subcontract anchors for sample and factory claim execution sections.
```

Backend/API:

```text
- No new backend API or database migration is included.
- Existing SubcontractOrder state, latest dispatch status, material line issue quantities, deposit status, and closeout quantities drive the tracker.
```

---

## 3. Verification

Local branch verification:

```text
- apps/web: vitest run --passWithNoTests passed (55 files, 316 tests)
- apps/web: vitest subcontractFactoryExecutionTracker.test.ts passed
- apps/web: vitest subcontractOrderTimeline.test.ts passed
- apps/web: tsc --noEmit passed
- apps/web: next build passed
- apps/api: go test ./... passed via D:\toolcache\go1.24.2\go\bin\go.exe
- apps/api: go vet ./... passed via D:\toolcache\go1.24.2\go\bin\go.exe
- packages/openapi: contract-check passed
- git diff --check passed
```

Pending before closeout:

```text
- GitHub CI
- Dev deploy
- Browser smoke for /production/factory-orders/:orderId execution tracker
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
No v0.28.0-factory-execution-tracking tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
