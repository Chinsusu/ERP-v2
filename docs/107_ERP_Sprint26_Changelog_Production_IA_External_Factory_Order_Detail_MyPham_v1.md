# 107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 26 - Production IA cleanup and external factory order detail
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Completed on main; dev deploy and Production browser smoke passed

---

## 1. Summary

Sprint 26 changes the user-facing model from separate Production/Subcontract sidebar entrypoints to:

```text
Production = user-facing module
External factory / subcontract = production execution method
```

The technical subcontract runtime remains in place. The UI now starts moving factory-order visibility under Production.

---

## 2. Runtime Changes

Frontend:

```text
- Subcontract is hidden from the sidebar while remaining route-addressable for operational execution.
- /production remains the primary Production entrypoint.
- Production Plan worklist uses Production-facing navigation for the factory-order step.
- Production Plan detail links related factory orders to /production/factory-orders/:orderId.
- /production/factory-orders/:orderId renders a factory-order detail page with source plan link, summary, timeline, closeout state, and material lines.
- Factory-order timeline service maps subcontract statuses into user-facing production steps.
```

Backend/API/DB:

```text
- No backend, OpenAPI, or DB schema changes in Sprint 26 implementation scope.
- Existing subcontract order API remains the source for factory order detail data.
```

---

## 3. Verification

Local implementation verification:

```text
- Targeted web tests for menu, Production Plan worklist, and factory-order timeline: pass.
- Web TypeScript check: pass.
- Full web vitest suite: pass.
- Web production build: pass.
- API go test ./...: pass.
- API go vet ./...: pass.
- OpenAPI contract check: pass.
- git diff --check: pass.
```

GitHub PR verification:

```text
- PR #591: Consolidate production factory order navigation.
- Merge commit: 5e8003a9c2bb7263ef70155229c5960af9d60ff6.
- Required checks passed: e2e, required-api, required-migration, required-openapi, required-web, web.
- Manual self-review completed before merge.
```

Dev verification:

```text
- Initial dev deploy attempt was blocked by /tmp free-space threshold.
- Safe Docker builder cache cleanup reclaimed 1.987GB; runtime volumes were not pruned.
- deploy-dev-staging.sh dev: pass.
- Full ERP dev smoke: pass.
- Browser smoke: login -> /production -> sidebar Production entry -> /production/factory-orders/:orderId detail.
```

---

## 4. Evidence

```text
PR number: #591
Merge commit: 5e8003a9c2bb7263ef70155229c5960af9d60ff6
Dev deploy: passed on 2026-05-06
Full dev smoke: passed
Browser smoke:
- Sidebar exposes /production once.
- Sidebar exposes no /subcontract link.
- Factory order detail opened at /production/factory-orders/sco-s25-ui-04225182.
Screenshots:
- D:\ERP-v2\output\playwright\s26-production-sidebar.png
- D:\ERP-v2\output\playwright\s26-factory-order-detail.png
```

---

## 5. Tag Status

```text
No v0.26.0-production-ia-factory-order-detail tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
