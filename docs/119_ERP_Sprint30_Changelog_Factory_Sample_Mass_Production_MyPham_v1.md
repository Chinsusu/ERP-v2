# 119_ERP_Sprint30_Changelog_Factory_Sample_Mass_Production_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 30 - Factory Sample Approval And Mass Production Start
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Completed and merged; CI, dev deploy, full dev smoke, and browser smoke passed

---

## 1. Summary

Sprint 30 adds production-facing controls for the next two external-factory execution gates:

```text
Material handover complete
-> factory sample submitted
-> sample approved or rejected
-> mass production started
```

No email, Zalo, supplier portal, external factory API, new backend API, finished-goods receipt, inbound QC, or internal MES behavior is included.

---

## 2. Runtime Changes

Frontend:

```text
- Add a factory sample and mass-production readiness/payload service.
- Add unit tests for sample readiness, sample decision readiness, no-sample mass-production readiness, and runtime payload generation.
- Add production-facing sample approval section to /production/factory-orders/:orderId.
- Add production-facing mass-production start section to /production/factory-orders/:orderId.
- Update tracker and timeline sample/mass actions to link to #factory-sample-approval and #factory-mass-production.
- Refresh in-page order state after sample submit, sample decision, and mass-production start.
```

Backend/API:

```text
- No new backend API or database migration is included.
- Existing submitSubcontractSample, approveSubcontractSample, rejectSubcontractSample, and startMassProductionSubcontractOrder runtime calls drive the flow.
```

---

## 3. Verification

Local branch verification:

```text
- apps/web: subcontractFactorySampleMassProduction.test.ts passed
- apps/web: subcontractFactoryExecutionTracker.test.ts passed
- apps/web: subcontractOrderTimeline.test.ts passed
- apps/web: vitest run --passWithNoTests passed, 57 files / 328 tests
- apps/web: tsc --noEmit passed
- apps/web: next build passed
- apps/api: go test ./... passed
- apps/api: go vet ./... passed
- packages/openapi: contract-check passed, 90 routes / 48 envelopes
```

Remote/dev verification:

```text
- GitHub PR #599 required-api, required-web, required-openapi, required-migration, web, and e2e checks passed.
- PR #599 was manually merged into main at bd645404.
- Dev deploy passed with ./infra/scripts/deploy-dev-staging.sh dev on 2026-05-06.
- Full dev smoke passed during deploy.
- Browser smoke passed on /production/factory-orders/sco-s16-08-03-smoke-0063#factory-sample-approval.
- Browser smoke verified #factory-sample-approval, #factory-mass-production, timeline/tracker hash links, sample heading/copy, mass-production heading/copy, and screenshot capture.
```

---

## 4. Evidence

```text
Runtime branch: codex/s30-sample-mass-production
Runtime commit: 84a323b0
PR number: #599
Merge commit: bd645404
GitHub CI: passed
Dev deploy: passed
Full dev smoke: passed
Browser smoke: passed
Screenshot: output/playwright/s30-factory-sample-mass-production.png
```

---

## 5. Tag Status

```text
No v0.30.0-factory-sample-mass-production tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
