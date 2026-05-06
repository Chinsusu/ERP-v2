# 117_ERP_Coding_Task_Board_Sprint30_Factory_Sample_Mass_Production_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 30 - Factory Sample Approval And Mass Production Start
Document role: Coding task board
Version: v1
Date: 2026-05-06
Status: Implementation in progress; PR, CI, merge, dev deploy, and browser smoke pending

---

## 1. Decision

Sprint 30 continues the external-factory production flow after Sprint 29 material handover:

```text
Factory confirmed
-> deposit/payment condition satisfied
-> material handover to factory
-> sample submitted / approved or rejected
-> mass production started
-> finished goods receipt to QC hold
```

The company production model remains external factory / subcontract manufacturing. Production users work from `/production`, while the existing subcontract runtime APIs remain the execution backend.

---

## 2. Scope

In scope:

```text
- Add a production-facing sample approval section on /production/factory-orders/:orderId.
- Add a production-facing mass-production start section on /production/factory-orders/:orderId.
- Use existing submit-sample, approve-sample, reject-sample, and start-mass-production runtime APIs.
- Keep tracker and timeline sample/mass actions linked to the factory order detail page.
- Show sample code, formula/spec version, evidence reference, decision reason, storage status, and current gate state.
- Update order state in-page after sample submit, sample decision, and mass-production start.
```

Out of scope:

```text
- Email, Zalo, factory portal/API delivery.
- Digital signature or automatic file upload.
- New backend API or database tables.
- Finished goods receipt, inbound QC, factory claim, or final payment closeout changes.
- Internal MES/work-center production.
```

---

## 3. Frontend Tasks

```text
1. Add a sample/mass-production gate service for readiness, blocking reason, and runtime payload generation.
2. Add tests for material-issued sample readiness, submitted sample decision, approved sample mass readiness, no-sample mass readiness, and payload generation.
3. Move sample and mass-production tracker/timeline actions to /production/factory-orders/:orderId hash sections.
4. Render sample submission and sample decision controls on the factory order detail page.
5. Render mass-production start controls on the factory order detail page.
6. Refresh local order state after each runtime action so tracker/timeline move forward without leaving the page.
```

---

## 4. Backend/API Tasks

No new backend API is required for Sprint 30.

The production-facing UI uses existing runtime endpoints:

```text
POST /api/v1/subcontract-orders/{id}/submit-sample
POST /api/v1/subcontract-orders/{id}/approve-sample
POST /api/v1/subcontract-orders/{id}/reject-sample
POST /api/v1/subcontract-orders/{id}/start-mass-production
```

Existing runtime output remains:

```text
- updated SubcontractOrder
- SubcontractSampleApproval for sample actions
- audit log id
```

---

## 5. Acceptance Criteria

```text
- After material handover, a sample-required order can submit a factory sample from /production/factory-orders/:orderId#factory-sample-approval.
- A submitted sample can be approved or rejected from the same page.
- A rejected sample blocks mass production and allows resubmission.
- An approved sample opens the mass-production start gate.
- A no-sample order can start mass production after material handover.
- Mass-production start updates the order to mass_production_started through existing runtime.
- Tracker and timeline sample/mass actions point to #factory-sample-approval and #factory-mass-production.
- No email/Zalo/API delivery, finished goods receipt, inbound QC, or internal MES behavior is added or implied.
- Local tests/build, PR CI, self-review, merge, dev deploy, and browser smoke are recorded before closeout.
```
