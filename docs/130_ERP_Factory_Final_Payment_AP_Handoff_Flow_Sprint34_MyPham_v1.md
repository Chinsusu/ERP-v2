# 130_ERP_Factory_Final_Payment_AP_Handoff_Flow_Sprint34_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 34
Document role: Flow design
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 34 implementation

---

## 1. Flow Position

Sprint 34 starts after Sprint 33 has marked final payment ready:

```text
Production Plan
-> Factory Order
-> Factory Dispatch
-> Material Handover
-> Sample Approval
-> Mass Production
-> Finished Goods Receipt to QC Hold
-> Finished Goods QC Closeout
-> Factory Claim Closeout
-> Final Payment Readiness
-> Factory Final Payment AP Handoff
```

---

## 2. Business Rule

Production can only say:

```text
This factory order is ready for final payment.
```

Finance must still control:

```text
supplier invoice evidence
AP match status
payment request
payment approval
payment recording
cash-out allocation
```

This avoids mixing production closeout with cash/payment approval.

---

## 3. Handoff Contract

When final payment readiness succeeds, the runtime creates or returns the supplier payable generated from the final payment milestone.

Traceability:

```text
supplier_payable.source_document.type = subcontract_payment_milestone
supplier_payable.source_document.id/no = final payment milestone id/no
supplier_payable.line.source_document.type = subcontract_order
supplier_payable.line.source_document.id/no = factory order id/no
```

Production page displays:

```text
AP no
AP id
factory/order source
milestone no
handoff status
link to Finance AP
```

---

## 4. UI Flow

Factory order detail:

```text
1. User resolves claim or confirms no blocking claim.
2. User marks final payment ready.
3. Section shows AP handoff evidence.
4. User clicks "Mo AP tai Finance".
5. Finance opens /finance?ap_q=:payableNo#supplier-payables with AP search selected.
6. If the factory-order page was reloaded after final payment readiness and the AP no is not in local UI state, the link searches Finance by factory order no because AP lines keep subcontract_order source evidence.
```

Finance supplier payables:

```text
1. Reads ap_q from the URL.
2. Filters AP list by payable no/id.
3. Selects the matching AP.
4. Shows supplier invoice matching card.
5. Blocks payment request until a matched supplier invoice exists.
```

---

## 5. Guardrails

```text
No AP link is shown before final payment readiness.
Open or acknowledged factory claim still blocks final payment readiness.
Full QC fail remains blocked until a later replacement/settlement flow.
Finance AP payment gate remains unchanged and enforced by backend.
```

---

## 6. Later Work

Later sprints can add:

```text
factory invoice attachment
VAT/withholding treatment
debit note / credit note settlement
claim replacement/rework order
cash-out bank transfer evidence UX
factory portal/API or email/Zalo sending
```
