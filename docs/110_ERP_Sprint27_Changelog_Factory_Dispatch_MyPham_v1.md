# 110_ERP_Sprint27_Changelog_Factory_Dispatch_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 27 - Factory Dispatch
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Completed on main; dev deploy and browser smoke passed

---

## 1. Summary

Sprint 27 adds a manual factory dispatch MVP for external-factory production:

```text
Approved factory order -> dispatch pack -> manual sent evidence -> factory response -> factory confirmation
```

No email, Zalo, supplier portal, or external factory API integration is included.

---

## 2. Runtime Changes

Backend/API:

```text
- Added subcontract factory dispatch domain model and lifecycle.
- Added PostgreSQL migration/store plus prototype fallback store.
- Added dispatch service to build packs from approved external factory orders.
- Added API endpoints for list/create/mark-ready/mark-sent/record-response.
- Added OpenAPI paths, schemas, and contract-check coverage.
- Added audit events for dispatch create/ready/sent/response/confirmed.
```

Frontend:

```text
- Added factory dispatch service/types and prototype fallback behavior.
- Added factory dispatch step to factory-order timeline.
- Added "Gửi nhà máy" section on /production/factory-orders/:orderId.
- Added manual actions: create pack, mark ready, mark sent, record factory response.
- Confirmed factory response advances the factory order to factory_confirmed.
```

---

## 3. Verification

Local branch verification executed:

```text
- apps/api: go test ./internal/modules/production/domain ./internal/modules/production/application -run "TestSubcontractFactoryDispatch|TestSubcontractOrderServiceFactoryDispatch" passed
- apps/api: go test ./cmd/api -run TestSubcontractOrderSmoke passed
- apps/api: go test ./... passed
- apps/api: go vet ./... passed
- apps/web: vitest subcontractOrderTimeline.test.ts passed
- apps/web: vitest subcontractOrderService.test.ts passed
- apps/web: vitest run --passWithNoTests passed
- apps/web: tsc --noEmit passed
- apps/web: next build passed
- packages/openapi: node packages/openapi/contract-check.mjs passed
```

Local verification caveats:

```text
- make is not available on this workstation.
- Redocly CLI/openapi-validate is not installed locally, so OpenAPI lint remains for GitHub CI.
- next build completed with existing SWC DLL fallback warnings on Windows.
```

---

## 4. Evidence

```text
PR number: #593
Merge commit: 3cc5852d
GitHub CI: api, web, migration, openapi, e2e, required-api, required-web, required-migration, required-openapi passed
Dev deploy: ./infra/scripts/deploy-dev-staging.sh dev passed
Dev migration: 44/u create_subcontract_factory_dispatches applied
Full dev smoke: passed
Browser smoke: login -> /production/factory-orders/:orderId -> create dispatch -> ready -> sent -> confirmed passed
Smoke order: SCO-S27-UI-11653066
Smoke dispatch: FDP-260505-505747 confirmed
Screenshots: output/playwright/s27-factory-dispatch-confirmed.png
```

---

## 5. Tag Status

```text
No v0.27.0-factory-dispatch tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
