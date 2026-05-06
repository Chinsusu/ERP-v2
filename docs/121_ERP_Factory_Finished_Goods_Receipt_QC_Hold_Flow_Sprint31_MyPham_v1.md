# 121_ERP_Factory_Finished_Goods_Receipt_QC_Hold_Flow_Sprint31_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 31
Document role: Flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 31 implementation

---

## 1. Flow Position

Sprint 31 sits after sample approval and mass-production start:

```text
Production Plan
-> Factory Order
-> Factory Dispatch
-> Factory Confirmation
-> Material Handover
-> Sample Approval
-> Mass Production Start
-> Finished Goods Receipt to QC Hold
-> Finished Goods QC
-> Factory Claim or Acceptance
-> Final Payment Readiness
```

This keeps the company model clear: production is external-factory manufacturing. The user sees it under `/production`, while the technical runtime still reuses subcontract APIs.

---

## 2. Receipt Gate

Receipt is allowed only when:

```text
order.status >= mass_production_started
remaining finished goods quantity > 0
order is not cancelled
order is not blocked by factory issue
```

Receipt is blocked when:

```text
factory has not started mass production
sample is rejected
order is cancelled
order is already fully received
receipt quantity exceeds remaining quantity
```

Partial receipts are allowed. The order may remain receivable until total received quantity reaches planned quantity.

---

## 3. Required Receipt Data

Header:

```text
warehouse_id
warehouse_code
location_id = qc_hold
location_code = QC-HOLD
delivery_note_no
received_by
received_at
note
```

Line:

```text
sku_code
item_name
receive_qty
uom_code
batch_no
lot_no
expiry_date
packaging_status
note
```

Evidence:

```text
delivery_note
packing_list
coa
photo
```

Sprint 31 UI starts with one finished-good line per factory order because each factory order is for one finished product. Multiple evidence types and multi-line split receipt can be expanded later if real receiving needs it.

---

## 4. Inventory Meaning

Receipt stock movement records the factory receipt into QC hold.

It must not mean:

```text
available stock
QC pass
finance acceptance
final payment readiness
```

The next sprint or later closeout work must explicitly move QC-passed quantity from hold to available stock and handle rejected quantity through factory claim.

---

## 5. UI Design

The factory order detail page should show the receipt section after mass production:

```text
Factory Execution Tracker
Material Handover
Sample Approval
Mass Production
Finished Goods Receipt to QC Hold
Factory Dispatch / Timeline / Material lines
```

The receipt section shows:

```text
planned quantity
received quantity
remaining quantity
QC hold destination
delivery note
receiver
warehouse
QC hold location
receipt quantity
batch / lot
expiry date
packaging condition
evidence filename
latest receipt summary
```

Tracker and timeline actions point to:

```text
/production/factory-orders/:orderId#factory-finished-goods-receipt
```

---

## 6. Traceability

Minimum traceability links:

```text
Production Plan -> Factory Order -> Finished Goods Receipt
Factory Order -> Receipt -> Stock Movement
Receipt -> QC Hold
Receipt -> later QC result
Receipt -> later Factory Claim if QC fails
```

The traceability wording should not say "received stock is usable" until QC pass is posted.
