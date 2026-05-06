# 127_ERP_Factory_Claim_Final_Payment_Closeout_Flow_Sprint33_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 33
Document role: Flow design
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 33 implementation

---

## 1. Flow Position

Sprint 33 sits after Sprint 32 finished goods QC closeout:

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
-> Factory Claim Closeout
-> Final Payment Readiness
```

---

## 2. Factory Claim States

```text
open:
  Claim has been created from QC partial fail or full fail.
  Final payment is blocked.

acknowledged:
  Factory has acknowledged the issue.
  Final payment is still blocked.

resolved:
  Internal owner has recorded the accepted resolution/workaround.
  Final payment can proceed only if accepted QC quantity exists and the order is otherwise eligible.

closed/cancelled:
  Reserved for later lifecycle cleanup.
```

---

## 3. Payment Gate

Final payment readiness is allowed only when:

```text
accepted quantity exists
no open or acknowledged factory claim blocks final payment
order is not full rejected with factory issue
order is not cancelled or closed
```

Final payment readiness is blocked when:

```text
there is an open factory claim
there is an acknowledged but unresolved factory claim
all received quantity failed QC
the order has no accepted finished goods quantity
```

---

## 4. User Experience

The production-facing factory order detail page must show one closeout surface:

```text
Claim nha may & thanh toan cuoi
```

The section shows:

```text
latest claim number
latest claim status
blocking claim count
affected quantity
factory acknowledgement actor/time
resolution actor/time
final payment gate status
```

Operators can:

```text
1. Acknowledge factory issue after receiving factory response.
2. Resolve the claim after agreeing compensation/rework/workaround outside this sprint's finance scope.
3. Mark final payment ready once the gate is clean.
```

---

## 5. Timeline And Tracker

Timeline and tracker must not jump from QC directly to final payment when a claim exists.

Required production-facing sequence:

```text
Finished goods QC closeout
-> Factory claim resolution
-> Final payment readiness
-> Order closeout
```

All actions should point to:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

---

## 6. Data Contract

Backend/API/DB contracts remain English:

```text
SubcontractFactoryClaim.status
acknowledged_by
acknowledged_at
resolved_by
resolved_at
resolution_note
blocks_final_payment
```

Business display remains Vietnamese-first in the UI.

---

## 7. Later Work

Sprint 33 intentionally leaves these for later:

```text
replacement production order
factory credit/debit note
claim settlement payable adjustment
factory portal/API acknowledgement
email/Zalo dispatch
internal production/MES
```
