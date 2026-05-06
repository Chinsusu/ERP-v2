# 113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 28 - Factory Execution Tracking
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Completed and merged; CI, dev deploy, full dev smoke, and browser smoke passed

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

GitHub and dev verification:

```text
- PR #595 required-api passed
- PR #595 required-web passed
- PR #595 required-openapi passed
- PR #595 required-migration passed
- PR #595 web passed
- PR #595 e2e passed
- Dev deploy passed with ./infra/scripts/deploy-dev-staging.sh dev
- Dev deploy reported no new migration for Sprint 28
- Full dev smoke passed
- Browser smoke passed for /production/factory-orders/sco-s16-07-01-1777715855439203730 execution tracker
```

---

## 4. Evidence

```text
PR number: #595
Branch: codex/s28-factory-execution-tracking
Runtime commit: 023a54b9
Merge commit: cd3a5b18
GitHub CI: required-api, required-web, required-openapi, required-migration, web, and e2e passed
Dev deploy: ./infra/scripts/deploy-dev-staging.sh dev passed on 2026-05-06
Full dev smoke: passed
Browser smoke: passed for /production/factory-orders/sco-s16-07-01-1777715855439203730
Screenshot: output/playwright/s28-factory-execution-tracker.png
```

---

## 5. Tag Status

```text
No v0.28.0-factory-execution-tracking tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
