# 136_ERP_Factory_Final_Payment_Voucher_Flow_Sprint36_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 36
Document role: Flow design
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 36 implementation

---

## 1. Flow Position

Sprint 36 starts after Sprint 35:

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
-> Payment Voucher / Cash-Out Evidence
```

---

## 2. Business Rule

Sprint 35 proves Finance can close the AP. Sprint 36 proves Finance can produce the cash/bank evidence for that closeout.

The separation stays:

```text
Production owns factory order execution context.
Finance owns supplier invoice, AP approval, payment recording, and cash/bank evidence.
CashTransaction owns the posted cash-out voucher evidence.
```

A payment voucher is not a bank transfer integration. It is an internal Finance document that records:

```text
who was paid
which AP was paid
how much was paid
payment date
payment method
bank/reference no
memo/source evidence
```

---

## 3. Voucher Data Contract

Use the existing CashTransaction model first.

Factory final-payment voucher:

```text
CashTransaction.direction = cash_out
CashTransaction.status = posted
CashTransaction.counterparty_id = supplier_payable.supplier_id
CashTransaction.counterparty_name = supplier_payable.supplier_name
CashTransaction.payment_method = bank_transfer / cash / other manual method
CashTransaction.reference_no = bank transaction no, payment slip no, or internal reference
CashTransaction.total_amount = payment amount
CashTransaction.currency_code = VND
CashTransaction.memo includes factory order no, AP no, invoice no, and final-payment milestone no when available
```

Voucher allocation:

```text
CashTransactionAllocation.target_type = supplier_payable
CashTransactionAllocation.target_id = supplier_payable.id
CashTransactionAllocation.target_no = supplier_payable.payable_no
CashTransactionAllocation.amount = allocated payment amount
```

---

## 4. Closeout Checklist

The factory final-payment Finance closeout becomes:

```text
1. AP created
2. Supplier invoice matched
3. Payment requested
4. Payment approved
5. AP payment recorded
6. Payment voucher posted
```

Step state rules:

```text
No matched supplier invoice:
  invoice matching is current
  payment request and voucher are blocked

Matched supplier invoice + open AP:
  invoice matching is complete
  payment request is current

Payment requested:
  payment approval is current

Payment approved or partially paid:
  AP payment recording is current
  voucher can be drafted/prefilled but should not claim closeout until posted

Paid AP + no matching cash_out voucher:
  AP payment recording is complete
  payment voucher is current

Paid AP + matching posted cash_out voucher:
  payment voucher is complete
  factory final-payment Finance closeout is complete
```

---

## 5. Navigation

Finance AP detail should provide:

```text
Open source factory order:
  /production/factory-orders/:orderId#factory-claim-final-payment-closeout

Open voucher:
  /finance?cash_q=:cashTransactionNo#cash-transactions

Create voucher:
  prefilled cash_out form or action using selected AP evidence
```

Cash transaction detail should provide:

```text
Open AP:
  /finance?ap_q=:payableNo#supplier-payables
```

This prevents the user from hunting across Finance tabs after a factory AP is paid.

---

## 6. Error Handling And Guardrails

```text
Do not create a voucher without a supplier_payable allocation.
Do not allow cash_in for supplier_payable allocation.
Do not allow allocation total to differ from voucher total.
Do not mark voucher closeout complete from memo/reference text alone.
Do not claim bank transfer automation from manual reference/evidence.
Do not bypass matched supplier invoice or AP approval controls.
Do not create duplicate vouchers silently; surface existing voucher evidence first.
```

If multiple vouchers allocate to one AP, Sprint 36 may display them as multiple voucher rows and sum allocated amount. It should not invent settlement tolerance rules.

---

## 7. Later Work

Later sprints can add:

```text
binary payment evidence attachments
payment voucher print/export template
bank transfer batch export
bank statement import/reconciliation
VAT/withholding handling
multi-currency settlement
supplier statement reconciliation
factory debit note / credit note settlement
general ledger posting
```
