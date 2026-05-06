# 124_ERP_Factory_Finished_Goods_QC_Closeout_Flow_Sprint32_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 32
Document role: Flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 32 implementation

---

## 1. Flow Position

Sprint 32 sits after factory finished-goods receipt:

```text
Production Plan
-> Factory Order
-> Factory Dispatch
-> Factory Confirmation
-> Material Handover
-> Sample Approval
-> Mass Production Start
-> Finished Goods Receipt to QC Hold
-> Finished Goods QC Closeout
-> Factory Claim or Acceptance
-> Final Payment Readiness
```

Production remains external-factory manufacturing. The user-facing route is `/production`, while the technical runtime continues to reuse subcontract APIs.

---

## 2. QC Gate

QC closeout is allowed only when:

```text
finished goods have been received into QC hold
order.status is finished_goods_received or qc_in_progress
received quantity > accepted quantity + rejected quantity
order is not cancelled
order is not already final_payment_ready or closed
```

QC closeout is blocked when:

```text
factory goods have not been received
accepted plus rejected quantity already covers received quantity
order is already accepted, rejected with factory issue, final payment ready, closed, or cancelled
```

Partial pass and full fail should attach the latest receipt when it is loaded so the factory claim keeps receipt-level traceability. Full pass can close from the order quantity because the backend accept API resolves stored receipt movements.

---

## 3. Decisions

Full pass:

```text
accepted_qty = remaining QC quantity
stock_status moves from qc_hold to available
order can advance to accepted
factory claim is not created
```

Partial pass:

```text
accepted_qty = operator input
rejected_qty = operator input
accepted portion moves from qc_hold to available
rejected portion opens factory claim
final payment remains blocked while claim is open or acknowledged
```

Full fail:

```text
affected_qty = remaining QC quantity
no available stock movement
factory claim is opened
order becomes rejected_with_factory_issue
```

---

## 4. Required UI Data

Display:

```text
factory order no
finished SKU and product name
received quantity
already accepted quantity
already rejected quantity
remaining QC quantity
latest closeout state
latest receipt no, delivery note, QC hold warehouse/location, batch no, lot no, expiry date, and receipt evidence count when receipt detail is loaded
```

Inputs:

```text
accepted_qty
rejected_qty
accepted_by
qc_note
claim_reason_code
claim_reason
claim_severity
claim_owner
claim_evidence_file
claim_evidence_note
```

---

## 5. Traceability

Minimum traceability links:

```text
Production Plan -> Factory Order -> Finished Goods Receipt
Receipt -> QC closeout decision
QC pass -> stock movement to available
QC partial/fail -> factory claim
Factory claim -> final payment hold
```

The UI must not imply final payment is released by QC. Final payment readiness remains a separate action after QC/claim state is acceptable.

---

## 6. Entry Point

Tracker and timeline QC actions point to:

```text
/production/factory-orders/:orderId#factory-finished-goods-qc-closeout
```

The `/subcontract` route can remain technical and addressable, but Sprint 32 makes the production factory order detail the primary user-facing place for finished-goods QC closeout.
