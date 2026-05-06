# 115_ERP_Factory_Material_Handover_Flow_Sprint29_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 29 - Factory Material Handover
Document role: Flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 29 implementation

---

## 1. Flow Position

Sprint 29 sits after factory confirmation and payment-condition readiness:

```text
Production plan
-> Factory order
-> Dispatch pack
-> Factory confirmed
-> Deposit / payment condition
-> Material handover to factory
-> Sample gate, if required
-> Mass production at factory
-> Finished goods receipt to QC hold
-> QC pass / factory claim
-> Final payment readiness
-> Close
```

The ERP must continue to present this as external-factory production, not internal MES production.

---

## 2. Material Handover Contract

The factory order detail page should expose one material handover section:

```text
/production/factory-orders/:orderId#factory-material-handover
```

The section should show:

```text
- Source warehouse
- Factory receiver
- Receiver contact
- Vehicle / carrier reference
- Handover evidence file/reference
- Handover note
- Material lines with planned, issued, remaining quantity, UOM, lot-control status
- Issue quantity per pending line
- Batch/lot and source bin per pending line
- Transfer result and movement evidence after submit
```

---

## 3. Gate Rules

Material handover can submit only when:

```text
1. Factory order is at least factory_confirmed.
2. Required deposit is paid, not required, or order status has reached deposit_recorded.
3. At least one material line still has remaining quantity.
4. Receiver is filled.
5. Handover evidence checkbox is checked.
6. Every pending lot-controlled line has batch/lot entered.
```

Material handover is blocked when:

```text
- Factory has not confirmed the order.
- Required deposit remains pending.
- The order has already moved past the material handover step.
- The order has no material lines.
```

---

## 4. Runtime Behavior

Submit uses the existing runtime:

```text
POST /api/v1/subcontract-orders/{id}/issue-materials
```

The payload preserves:

```text
- source_warehouse_id
- source_warehouse_code
- handover_by
- received_by
- receiver_contact
- vehicle_no
- note
- order_material_line_id
- issue_qty
- uom_code
- batch_no
- source_bin_id
- handover evidence metadata
```

The runtime returns:

```text
- updated subcontract order
- material transfer document
- SUBCONTRACT_ISSUE stock movements
- audit log reference
```

The UI should update the local page state from the returned order so the S28 tracker and timeline immediately move forward when all lines are issued.

---

## 5. Quantity Rules

```text
- Quantity values keep the existing six-decimal runtime scale.
- Display uses Vietnamese number formatting through existing shared formatters.
- Default issue quantity is the remaining quantity for each pending line.
- Complete lines are read-only.
- Over-issue remains blocked by existing runtime validation.
- QC hold/fail stock availability enforcement remains a backend/inventory responsibility.
```

---

## 6. Guardrails

```text
- Do not imply automatic sending to factory.
- Do not expose /subcontract as a primary navigation entry.
- Do not create a new production/MES module.
- Do not bypass deposit/payment condition.
- Do not bypass lot/batch evidence for controlled materials.
- Keep backend/API/DB values English.
- Keep user-facing production copy Vietnamese.
```
