# 94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 23 follow-up - Purchase Request workflow
Version: v1
Date: 2026-05-05
Status: Implementation design for first-class Purchase Request bridge

---

## 1. Goal

Promote production-plan shortage output from an embedded draft into a traceable Purchase Request workflow:

```text
Production Plan
-> Purchase Request
-> submit / approve
-> convert to PO
-> PO detail
-> Receiving
-> Inbound QC
-> material readiness
-> Subcontract order
```

This keeps the business boundary clean: planning creates demand evidence, purchase owns PR approval and PO creation, warehouse owns receiving/QC, and production only proceeds when materials are ready.

---

## 2. Runtime Boundary

```text
Production page:
- create/select production plan
- show material demand
- open generated Purchase Request
- no direct PO creation

Purchase Request page:
- show source production plan
- show material lines
- submit request
- approve request
- convert approved request to PO
- link to created PO

PO page:
- show PO status/timeline
- link to related receiving records
- keep source reference back to plan/request
```

---

## 3. Purchase Request Statuses

```text
draft
submitted
approved
converted_to_po
cancelled
rejected
```

Allowed transitions:

```text
draft -> submitted
submitted -> approved
submitted -> rejected
approved -> converted_to_po
draft/submitted/approved -> cancelled
```

---

## 4. Acceptance Criteria

```text
1. Production plan detail opens the correct Purchase Request.
2. Purchase Request detail shows status, timeline, material lines, and source plan.
3. Draft request can be submitted.
4. Submitted request can be approved.
5. Approved request can be converted to PO.
6. Converted request stores PO id/no and links to PO detail.
7. Production plan worklist shows PR, approval, PO, receiving/QC, and subcontract steps.
8. Direct PO creation from /production is removed from the UI.
9. OpenAPI covers Purchase Request list/detail/action/convert endpoints.
10. Dev smoke must cover plan -> PR -> approve -> PO -> receiving link.
```

---

## 5. Known Limits

```text
- Production plan material readiness still does not auto-recalculate after receiving/QC.
- PO source relation is still carried by PR conversion fields and PO note, not a generic source-document table.
- Inbound QC detail remains a separate module follow-up.
- Cancel/reject UI can be added later if the business needs it in UAT.
```
