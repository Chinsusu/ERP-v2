# 137_ERP_Sprint36_Changelog_Factory_Final_Payment_Voucher_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 36
Change type: Finance cash/bank evidence and payment voucher traceability
Version: v1
Date: 2026-05-07
Status: Opened; implementation, PR, CI, merge, deploy, and browser smoke pending

---

## 1. Summary

Sprint 36 is opened to add payment voucher / cash-out evidence after Sprint 35 closes the factory final-payment AP in Finance.

Target Finance AP route:

```text
/finance?ap_q=:payableNo#supplier-payables
```

Target cash transaction route:

```text
/finance?cash_q=:cashTransactionNo#cash-transactions
```

Target production source route:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

---

## 2. Planned Runtime Surface

Implementation scope:

```text
1. Reuse CashTransaction as the payment voucher surface for posted cash_out supplier_payable allocations.
2. Detect vouchers linked to factory final-payment AP records.
3. Prefill voucher creation from selected AP source evidence.
4. Add voucher evidence to the factory final-payment Finance closeout checklist.
5. Link AP closeout to voucher/cash transaction detail and voucher detail back to AP.
6. Keep matched supplier invoice and AP approval/payment gates unchanged.
7. Add focused tests and update OpenAPI/schema only if the API surface changes.
```

---

## 3. Guardrails

```text
Finance remains the payment and cash/bank evidence surface.
Production remains the factory-order source surface.
CashTransaction is the first implementation model; do not create a separate PaymentVoucher entity unless necessary.
Matched supplier invoice remains required before payment request, approval, or recording.
Payment voucher posting is manual evidence, not bank automation.
No email/Zalo/bank API behavior is included.
No v0.36 release tag is planned unless explicitly requested.
```

---

## 4. Verification Evidence

Runtime PR:

```text
Pending.
```

Local verification:

```text
Pending.
```

GitHub CI:

```text
Pending.
```

Dev deploy and smoke:

```text
Pending.
```

---

## 5. Known Limits

```text
Binary payment attachment upload/OCR remains out of scope.
Bank transfer export/API remains out of scope.
VAT/withholding/multi-currency factory payment treatment remains out of scope.
Supplier statement reconciliation remains out of scope.
Factory debit note / credit note settlement remains out of scope.
General ledger posting remains out of scope.
```
