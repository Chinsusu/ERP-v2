# 106_ERP_Production_IA_External_Factory_Order_Detail_Flow_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Production IA and external factory order flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 26 implementation

---

## 1. Decision

For Phase 1, the correct product model is:

```text
Sản xuất = module tổng cho người dùng
Gia công ngoài / nhà máy ngoài = cách thực hiện sản xuất hiện tại
```

The UI should not present **Production** and **Subcontract** as two equal sidebar modules. That makes users think the system has two different production worlds.

The technical contract can continue using `subcontract` routes, API names, DB schemas, enum values, and audit event codes where already implemented. User-facing navigation should use Production as the main entrypoint.

---

## 2. User-Facing Navigation

Primary entrypoints:

```text
/production
/production/plans/:planId
/production/factory-orders/:orderId
```

Hidden/backward-compatible execution surface:

```text
/subcontract
```

`/subcontract` remains available because it contains existing operational forms for status changes, material transfer, sample approval, finished goods receipt, claim, and payment readiness. It should not be exposed as a primary sidebar tab.

---

## 3. Production Flow

```text
Production Plan
-> Material demand
-> Purchase Request / PO for shortages
-> Receiving / inbound QC for purchased material
-> Warehouse Issue Note to factory
-> Factory Order
-> Factory confirmation
-> Deposit record
-> Material handover evidence
-> Sample submission / approval
-> Mass production
-> Finished goods receipt to QC hold
-> QC pass / factory claim
-> Final payment readiness
-> Close factory order and production plan
```

Guardrail:

```text
Finished goods from the external factory do not become available stock until QC passes.
QC fail opens or links a factory claim and must block normal final payment unless an approved exception exists.
```

---

## 4. Detail Page Contract

Factory order detail at `/production/factory-orders/:orderId` should show:

```text
- Order number
- Source Production Plan link
- Factory
- Finished product
- Planned quantity
- Received quantity
- Accepted quantity
- Rejected quantity
- Expected receipt date
- Deposit status
- Final payment status
- Timeline
- Material lines
- Links to hidden operational execution sections where needed
```

Timeline states:

```text
1. Created
2. Submitted
3. Approved
4. Factory confirmed
5. Deposit recorded
6. Materials issued to factory
7. Sample submitted / approved / rejected
8. Mass production started
9. Finished goods received
10. QC in progress / accepted / rejected with factory issue
11. Final payment ready
12. Closed
```

---

## 5. Follow-Up Boundary

Sprint 26 intentionally does not rewrite the whole subcontract workspace. The next production sprint can move more execution forms from `/subcontract` into production-facing detail pages after the detail/timeline navigation is validated with real users.

Likely follow-up:

```text
Production factory order execution page:
- Confirm factory
- Record deposit
- Issue materials
- Submit/approve sample
- Receive finished goods
- Record QC result
- Open/close claim
- Mark final payment ready
```
