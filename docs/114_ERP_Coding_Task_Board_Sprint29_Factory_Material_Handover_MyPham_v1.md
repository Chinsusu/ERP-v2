# 114_ERP_Coding_Task_Board_Sprint29_Factory_Material_Handover_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 29 - Factory Material Handover
Document role: Coding task board
Version: v1
Date: 2026-05-06
Status: Implemented on branch; PR/CI/dev deploy evidence pending

---

## 1. Decision

Sprint 29 turns the factory execution tracker material gate into a production-facing handover surface.

Current operating model remains:

```text
Company production = external factory / subcontract manufacturing.
Production users should work from /production.
Existing subcontract runtime APIs and operational stores remain the execution backend.
```

Sprint 28 ended at:

```text
Factory confirmed -> current gate/worklist shows Material handover as the next action.
```

Sprint 29 starts there and answers:

```text
Which materials still need to be handed to the factory?
Can the user issue them now?
Which warehouse, batch/lot, bin, receiver, and evidence were used?
Did the order advance when all material lines were issued?
```

---

## 2. Scope

In scope:

```text
- Add a production-facing material handover section on /production/factory-orders/:orderId.
- Keep the S28 material-handover gate linked to the production detail page instead of /subcontract.
- Show material line planned, issued, remaining, lot-control requirement, and current status.
- Let the user select source warehouse, receiver, contact, vehicle, handover evidence, issue quantity, batch/lot, and bin.
- Call the existing subcontract material issue runtime API.
- Show transfer/stock movement evidence returned by the runtime.
- Update the order in-place so tracker and timeline reflect the new material status.
- Keep /subcontract route hidden from primary navigation.
```

Out of scope:

```text
- Email, Zalo, factory portal/API delivery.
- Internal MES or work-center execution.
- New backend API or database tables.
- Warehouse issue note redesign.
- Cost variance, GL posting, and material consumption costing.
- Automatic document attachment upload.
```

---

## 3. Frontend Tasks

```text
1. Add a material handover service for readiness, remaining quantity, default issue payload, and blocking reason.
2. Add tests for paid deposit readiness, pending deposit blocking, complete handover, and issue payload generation.
3. Move material-handover tracker/timeline actions to /production/factory-orders/:orderId#factory-material-handover.
4. Render a production-facing material handover form on the factory order detail page.
5. Require lot/batch for lot-controlled pending lines before submit.
6. Call issueSubcontractMaterials and refresh the in-page order state after success.
7. Show the transfer number, warehouse, factory, movement count, and evidence placeholder count after handover.
```

---

## 4. Backend/API Tasks

```text
No new backend API is required for Sprint 29.

The production-facing UI uses the existing runtime endpoint:
- POST /api/v1/subcontract-orders/{id}/issue-materials

Existing runtime output remains:
- updated SubcontractOrder
- SubcontractMaterialTransfer
- SUBCONTRACT_ISSUE stock movements
- audit log id
```

Follow-up backend work may be opened later if material handover must create a first-class Warehouse Issue Note document instead of the current subcontract material transfer.

---

## 5. Acceptance Criteria

```text
- Factory-confirmed order with paid/not-required deposit can open material handover on /production/factory-orders/:orderId.
- Factory-confirmed order with pending required deposit is blocked with a visible reason.
- Each material line shows planned, issued, remaining quantity, UOM, and lot-control status.
- Lot-controlled lines require batch/lot before handover submit.
- Submit records material issue through the existing runtime and updates the order state in the page.
- Fully issued orders advance to materials_issued_to_factory through existing runtime behavior.
- Tracker and timeline material actions point to #factory-material-handover.
- No email/Zalo/API delivery or internal MES behavior is added or implied.
- Local tests/build, PR CI, self-review, merge, dev deploy, and browser smoke are recorded before closeout.
```
