# 96_ERP_Supplier_Invoice_Three_Way_Matching_Flow_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 23 follow-up - Supplier invoice and AP matching
Version: v1
Date: 2026-05-05
Status: Implementation design and runtime behavior lock

---

## 1. Goal

Close the finance gap after supplier payable creation:

```text
Purchase Order
-> Goods Receipt / QC PASS
-> Supplier Payable
-> Supplier Invoice
-> 3-way match decision
-> Payment request / approval / payment recording
```

The business rule is:

```text
Supplier payable value comes from accepted receiving.
Supplier invoice is the vendor evidence that must be matched before finance treats the AP as payment-ready.
```

This keeps inventory acceptance and vendor billing as separate controls.

---

## 2. Document Boundary

Current implemented boundary from file 95:

```text
PO-linked posted receipt with QC PASS lines creates supplier payable value.
QC HOLD / FAIL lines do not create AP value.
```

This follow-up adds:

```text
Supplier Invoice document
Supplier Invoice -> AP link
Supplier Invoice -> PO / receipt trace through AP source and AP lines
Match decision: matched or mismatch
Finance UI visibility before payment action
```

Out of scope for this increment:

```text
VAT allocation
Withholding tax
Landed cost allocation
Invoice image OCR
Supplier portal/email ingestion
Multi-currency
Line-level price tolerance matrix
Accounting journal posting
Automatic bank payment file generation
```

---

## 3. Runtime Flow

### Step 1: AP Exists

The system already has a supplier payable created from a posted warehouse receipt.

```text
AP source document: warehouse_receipt
AP line source document: purchase_order
AP total amount: accepted QC PASS received value
```

### Step 2: Finance Records Supplier Invoice

Finance opens AP and records supplier invoice information:

```text
Supplier invoice no
Invoice date
Invoice amount
Currency VND
Linked AP id/no
```

The system derives these from AP:

```text
Supplier identity
Expected AP amount
Warehouse receipt source
Purchase order line sources
```

### Step 3: Matching

The first production-safe matching rule is strict:

```text
supplier_id must match AP supplier
invoice total amount must equal AP total amount
currency must be VND
```

If all checks pass:

```text
supplier invoice status = matched
match_status = matched
variance_amount = 0.00
```

If any check fails:

```text
supplier invoice status = mismatch
match_status = mismatch
variance_amount = invoice_total - ap_total
```

### Step 4: Payment Readiness

Matched invoices make the linked AP payment-ready from a finance control perspective.

Mismatch invoices keep the AP under manual review. Finance should not request payment until the invoice mismatch is resolved or explicitly handled by a future approval policy.

---

## 4. Traceability Contract

Supplier invoice header:

```text
payable_ref = supplier payable id
payable_no = supplier payable no
supplier_ref = AP supplier id
invoice_no = vendor invoice no
match_status = matched | mismatch
expected_amount = AP total amount
invoice_amount = invoice total amount
variance_amount = invoice_amount - expected_amount
```

Supplier invoice line:

```text
source_document.type = supplier_payable | warehouse_receipt | purchase_order
source_document.id/no = linked AP, receipt, or PO reference
amount = invoice line amount
```

This gives navigation:

```text
PO detail -> AP -> Supplier Invoice
Supplier Invoice -> AP detail
Supplier Invoice -> Receipt source
Supplier Invoice -> PO line source
AP detail -> invoice match status
Finance list/search -> invoice no, AP no, PO no, receipt no
```

---

## 5. UI Contract

Finance AP view should show:

```text
AP amount
Supplier invoice status
Invoice no
Invoice amount
Variance
Match decision
```

Finance user actions:

```text
Create supplier invoice from selected AP
Open supplier invoice detail
Search supplier invoices by invoice no, AP no, PO no, receipt no, supplier
```

Vietnamese display copy can use:

```text
Hóa đơn NCC
Đã khớp
Lệch đối chiếu
Số tiền hóa đơn
Số tiền AP
Chênh lệch
```

Technical API/entity names remain English.

---

## 6. Acceptance Criteria

```text
1. Finance can create a supplier invoice linked to an existing supplier payable.
2. Supplier invoice inherits supplier/AP traceability from the AP.
3. Invoice amount equal to AP total creates matched status with zero variance.
4. Invoice amount different from AP total creates mismatch status with non-zero variance.
5. Supplier invoice list/search can find invoice by invoice no, AP no, PO no, receipt no, or supplier.
6. Supplier invoice detail exposes AP, receipt, PO source references.
7. Finance AP UI shows invoice match visibility for the selected AP.
8. OpenAPI contract includes supplier invoice endpoints and schemas.
9. PostgreSQL runtime persists supplier invoices and lines.
10. Prototype fallback remains available only where prototype fallback is allowed.
```

---

## 7. Follow-Up Scope

After this increment, the next finance hardening should decide:

```text
AP payment readiness hard gate: implemented and locked in file 97.
Tolerance policy: allow small rounding/price differences with approval.
Tax handling: VAT, discounts, freight, landed cost allocation.
Invoice attachment: scanned invoice files linked to AP/invoice.
```
