# 133_ERP_Factory_Final_Payment_Finance_Closeout_Flow_Sprint35_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 35
Document role: Flow design
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 35 implementation

---

## 1. Flow Position

Sprint 35 starts after Sprint 34:

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
-> Finance Final Payment Closeout
```

---

## 2. Business Rule

Production can identify that the factory order is ready for final payment and can create/expose the AP handoff.

Finance must own:

```text
supplier invoice evidence
invoice/AP matching
payment request
payment approval
payment recording
AP closeout evidence
```

This keeps production execution separate from cash/payment authority.

---

## 3. Finance Closeout Checklist

A factory final-payment AP is treated as a closeout checklist:

```text
1. AP created
2. Supplier invoice matched
3. Payment requested
4. Payment approved
5. Payment recorded
```

Step state rules:

```text
No matched supplier invoice:
  invoice matching is current
  payment request is blocked

Matched supplier invoice + open AP:
  invoice matching is complete
  payment request is current

Payment requested:
  payment approval is current

Payment approved or partially paid:
  payment recording is current

Paid:
  payment recording is complete
```

---

## 4. Source Traceability

Finance AP source evidence should remain visible:

```text
AP source:
  subcontract_payment_milestone id/no

AP line source:
  subcontract_order id/no
```

The Finance panel should link back to:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

This gives Finance enough context to confirm:

```text
factory order
accepted finished goods
claim status
final payment readiness
AP source milestone
```

---

## 5. Guardrails

```text
Do not let Finance request/approve/record AP payment without matched supplier invoice.
Do not turn Production into a payment surface.
Do not create a second AP or invoice model for factory final payment.
Do not claim bank/cash closeout unless payment recording is actually posted.
Do not tag v0.35 from dev-only evidence.
```

---

## 6. Later Work

Later sprints can add:

```text
factory invoice attachments
payment voucher print/export
bank transfer batch export
cash transaction allocation detail
VAT/withholding rules
supplier statement reconciliation
factory settlement debit/credit notes
```
