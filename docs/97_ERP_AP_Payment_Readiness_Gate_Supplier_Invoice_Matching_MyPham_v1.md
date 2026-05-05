# 97_ERP_AP_Payment_Readiness_Gate_Supplier_Invoice_Matching_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 23 follow-up - AP payment readiness gate
Version: v1
Date: 2026-05-05
Status: Runtime behavior lock

---

## 1. Goal

Close the remaining finance control gap after supplier invoice matching:

```text
PO -> Receiving / QC PASS -> Supplier Payable -> Supplier Invoice -> Matched -> Payment request
```

The hard rule is:

```text
Supplier payable cannot enter payment request, payment approval, or payment recording unless it has at least one matched supplier invoice.
```

This keeps warehouse acceptance, vendor billing, and cash/payment approval as separate controls.

---

## 2. Runtime Rule

An AP is payment-ready only when all checks are true:

```text
supplier_invoice.payable_id = supplier_payable.id
supplier_invoice.supplier_id = supplier_payable.supplier_id
supplier_invoice.currency_code = supplier_payable.currency_code = VND
supplier_invoice.expected_amount = supplier_payable.total_amount
supplier_invoice.status = matched
supplier_invoice.match_status = matched
supplier_invoice.variance_amount = 0.00
```

If no supplier invoice exists, or the latest visible invoice is draft/mismatch/void:

```text
Payment request: blocked
Payment approval: blocked
Payment recording: blocked
```

Allowed actions remain:

```text
Create supplier invoice
Void AP
Reject an already requested payment
Mark AP disputed where the AP state allows it
```

---

## 3. Backend Contract

Supplier payable payment actions must check supplier invoice readiness before status transition:

```text
POST /api/v1/supplier-payables/{id}/request-payment
POST /api/v1/supplier-payables/{id}/approve-payment
POST /api/v1/supplier-payables/{id}/record-payment
```

When blocked, the API returns a conflict app error:

```text
code: SUPPLIER_PAYABLE_INVOICE_NOT_MATCHED
message: Supplier payable requires a matched supplier invoice before payment
```

The check belongs in the application service because it coordinates two finance aggregates:

```text
Supplier Payable
Supplier Invoice
```

The supplier payable domain model still owns payable status transitions, but readiness checks sit one layer above it.

---

## 4. UI Contract

The Finance AP view should show payment readiness beside payment actions:

```text
Sẵn sàng thanh toán
Cần hóa đơn NCC
Hóa đơn chưa khớp
Đang kiểm tra hóa đơn
```

Button behavior:

```text
Request payment disabled unless payment readiness is true.
Approve payment and record payment remain protected by backend even if older AP state exists.
Mismatch state points finance users to resolve invoice variance or reject/dispute the payment path.
```

The AP view already shows:

```text
Hóa đơn NCC
Số hóa đơn
Số tiền hóa đơn
Số tiền AP
Chênh lệch
Match status
```

This document locks that supplier invoice card as payment-readiness evidence, not only informational display.

---

## 5. Acceptance Criteria

```text
1. AP request-payment without matched supplier invoice returns SUPPLIER_PAYABLE_INVOICE_NOT_MATCHED.
2. AP request-payment with a matched supplier invoice succeeds from open AP.
3. AP approve-payment checks the same readiness gate before approval.
4. AP record-payment checks the same readiness gate before cash/payment posting.
5. Finance AP UI disables Request payment when no matched supplier invoice exists.
6. Finance AP UI shows a clear readiness reason in Vietnamese.
7. Existing create supplier invoice and AP detail flows still work.
8. Backend tests cover blocked and allowed payment request paths.
9. Frontend tests cover readiness helper behavior for no invoice, mismatch, and matched invoice.
```

---

## 6. Out Of Scope

```text
Tolerance policy
VAT and withholding tax matching
Invoice image attachment/OCR
Bank payment file generation
Accounting journal posting
Supplier portal submission
Multi-currency
```

Those follow after the strict VND match gate is stable.
