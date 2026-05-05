# 107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 26 - Production IA cleanup and external factory order detail
Document role: Changelog and verification evidence
Version: v1
Date: 2026-05-06
Status: Implementation branch; merge and dev smoke pending

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

Implementation branch verification:

```text
- Targeted web tests for menu, Production Plan worklist, and factory-order timeline: pass.
```

Pending before PR/merge:

```text
- Full web vitest suite.
- Web TypeScript check.
- Web production build.
- OpenAPI contract check if no backend contract changed can be recorded as not applicable or pass through CI.
- GitHub CI.
- Manual diff review.
- Dev deploy and browser smoke for Production sidebar and factory-order detail route.
```

---

## 4. Evidence Pending

```text
PR number: pending
Merge commit: pending
Dev deploy: pending
Browser smoke: pending
Screenshots: pending
```

---

## 5. Tag Status

```text
No v0.26.0-production-ia-factory-order-detail tag has been created.
Tag status: hold.
Create a tag only if a deliberate release checkpoint is requested.
```
