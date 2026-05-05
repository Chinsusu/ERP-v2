# 108_ERP_Coding_Task_Board_Sprint27_Factory_Dispatch_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 27 - Factory Dispatch
Document role: Coding task board
Version: v1
Date: 2026-05-06
Status: Implementation complete on branch; PR, merge, dev deploy, and smoke pending

---

## 1. Decision

Sprint 27 adds a production-facing factory dispatch MVP:

```text
Production Plan -> Factory Order -> Factory Dispatch Pack -> Factory response
```

No email, Zalo, supplier portal, or external factory API integration is included in this sprint.

Dispatch is a persisted communication/evidence document linked to the factory order. It does not replace the existing subcontract/factory order runtime.

---

## 2. Scope

In scope:

```text
- Create a dispatch pack from an approved factory order.
- Persist dispatch header, material lines, evidence, send metadata, and factory response.
- Mark dispatch as ready.
- Mark dispatch as manually sent to factory.
- Record factory response: confirmed, revision requested, rejected.
- When factory confirms, transition the factory order to factory_confirmed.
- Show dispatch state and actions on /production/factory-orders/:orderId.
- Add dispatch step before factory confirmation in the production-facing timeline.
```

Out of scope:

```text
- Email sending.
- Zalo sending.
- Factory portal/API integration.
- Digital signatures.
- Full PDF/print template polish.
- Costing variance.
- Internal MES/work-center production.
```

---

## 3. Backend Tasks

```text
1. Add factory dispatch domain model.
2. Add PostgreSQL migration and store.
3. Add prototype store for no-DB/local fallback.
4. Add application service to build packs from SubcontractOrder snapshots.
5. Add API endpoints under /api/v1/subcontract-orders/:id/factory-dispatches.
6. Add OpenAPI contract.
7. Add audit actions for create/ready/sent/response.
```

---

## 4. Frontend Tasks

```text
1. Add factory dispatch service/types.
2. Load dispatches on factory order detail.
3. Render "Gửi nhà máy" section.
4. Add buttons/forms for create, ready, sent, and response.
5. Add dispatch status to timeline.
6. Add tests for service mapping and timeline behavior.
```

---

## 5. Acceptance Criteria

```text
- Approved factory order can create a dispatch pack.
- Dispatch pack includes finished product, planned quantity, spec summary, expected receipt date, material lines, and notes.
- Dispatch can be marked ready.
- Dispatch can be marked sent with manual evidence.
- Factory confirmed response transitions order to factory_confirmed.
- Revision requested or rejected response does not advance the order to factory_confirmed.
- Factory order detail page shows dispatch status, evidence, and response.
- No email/Zalo/API integration is implemented or implied.
- CI, local build/test, dev deploy, and browser smoke are recorded before closeout.
```
