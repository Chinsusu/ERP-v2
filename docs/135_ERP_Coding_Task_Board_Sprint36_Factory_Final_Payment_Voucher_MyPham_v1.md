# 135_ERP_Coding_Task_Board_Sprint36_Factory_Final_Payment_Voucher_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 36
Document role: Coding task board
Version: v1
Date: 2026-05-07
Status: Opened for Sprint 36 implementation

---

## 1. Sprint Goal

Sprint 36 closes the cash/bank evidence layer after Sprint 35 proves Finance can close a factory final-payment AP:

```text
Factory final-payment AP handoff
-> supplier invoice matched
-> payment requested
-> payment approved
-> AP payment recorded
-> payment voucher / cash-out evidence linked to AP, invoice, and factory order
```

The user-facing Finance route remains:

```text
/finance?ap_q=:payableNo#supplier-payables
```

The cash/bank transaction surface remains:

```text
/finance#cash-transactions
```

The production source route remains:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

---

## 2. Scope

Runtime scope:

```text
1. Treat a payment voucher as a posted cash_out CashTransaction allocated to supplier_payable.
2. Add Finance-side guidance for creating or finding a payment voucher from a factory final-payment AP.
3. Prefill cash_out voucher fields from the selected AP: supplier, payable id/no, amount, VND currency, reference no, memo, and allocation.
4. Detect existing payment voucher evidence by cash transaction allocation target_type=supplier_payable and target_id/target_no matching the AP.
5. Add a voucher step to the factory final-payment Finance closeout checklist after payment approval/payment recording.
6. Link the AP closeout card to the matching cash transaction detail or filtered cash transaction list.
7. Link the cash transaction detail back to the source AP when allocation target is supplier_payable.
8. Keep the existing matched supplier invoice payment gate intact.
9. Add focused backend/frontend tests around voucher creation, allocation matching, and factory AP closeout state.
10. Update README, master docs index, and Sprint 36 changelog evidence.
```

---

## 3. Out Of Scope

```text
Direct bank API integration.
Bank transfer batch file export.
Automatic email/Zalo payment notification.
Binary payment attachment upload/OCR.
General ledger journal posting.
VAT, withholding tax, multi-currency, or payment tolerance policy.
Factory debit note / credit note settlement.
Supplier statement reconciliation.
Internal MES or work-center production.
v0.36 release tag unless explicitly requested later.
```

---

## 4. Acceptance Criteria

```text
- A factory final-payment AP can show whether a supplier_payable cash_out voucher exists.
- When no voucher exists, Finance can open/prefill a payment voucher action from the AP detail.
- The voucher posts as cash_out with allocation target_type=supplier_payable, target_id=AP id, target_no=AP no, amount=AP paid/outstanding amount as appropriate.
- The voucher records business date, payment method, reference no, counterparty, total amount, currency VND, and memo/source evidence.
- The factory final-payment closeout card shows voucher no/reference once voucher evidence exists.
- The closeout checklist distinguishes AP payment status from cash/bank voucher evidence.
- Cash transaction detail shows AP allocation and lets Finance trace back to the AP by target no.
- Non-factory APs keep the existing generic cash transaction behavior.
- Payment request/approval/recording still require a matched supplier invoice.
- No bank automation or real money movement is implied by posting a voucher.
```

---

## 5. Verification Plan

```text
Local:
- git diff --check
- targeted Go tests for CashTransaction supplier_payable allocation matching and AP/voucher helper behavior
- targeted web Vitest for Finance voucher helper and affected services/components
- full API tests if backend runtime changes are made
- web typecheck/build if frontend code changes are made
- OpenAPI contract check if API schema changes are made

GitHub CI:
- required-api
- required-web
- required-openapi
- required-migration

Dev smoke after merge:
- deploy-dev-staging.sh dev
- full dev smoke
- browser smoke Finance AP closeout -> create/find payment voucher -> cash transaction detail -> AP/factory source trace
```

---

## 6. Implementation Notes

Existing Finance runtime already has:

```text
CashTransaction
- direction: cash_in / cash_out
- status: posted / void
- business_date
- counterparty
- payment_method
- reference_no
- allocations
- total_amount
- currency_code
- memo

CashTransactionAllocation
- target_type
- target_id
- target_no
- amount
```

Sprint 36 should reuse this model first. A separate PaymentVoucher entity is not required unless implementation proves that cash transaction cannot express the workflow.

For factory AP vouchers, the canonical allocation is:

```text
direction = cash_out
allocation.target_type = supplier_payable
allocation.target_id = supplier_payable.id
allocation.target_no = supplier_payable.payable_no
```

The voucher is financial evidence of cash/bank movement. It must not bypass AP payment approval or matched-invoice controls.
