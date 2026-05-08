# 137_ERP_Sprint36_Changelog_Factory_Final_Payment_Voucher_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 36
Change type: Finance cash/bank evidence and payment voucher traceability
Version: v1
Date: 2026-05-07
Status: Completed and merged; dev deploy and browser smoke passed

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
PR #616: Wire factory final payment voucher evidence.
Implementation branch: codex/s36-factory-final-payment-voucher-runtime.
Merge commit: f97bc0d9.
```

Local verification:

```text
git diff --check: pass.
Targeted web test command attempted:
pnpm --filter web test -- apps/web/src/modules/finance/services/supplierPayableFactoryCloseout.test.ts
Result: blocked locally because pnpm is not installed on this workstation.
node --version: blocked locally by WindowsApps Access is denied.
```

GitHub CI:

```text
PR #616 CI passed after follow-up test expectation fix:
- required-api: pass
- required-web: pass
- required-openapi: pass
- required-migration: pass
- e2e: pass
- web: pass
```

Dev deploy and smoke:

```text
Dev deploy: deploy-dev-staging.sh dev passed on 2026-05-08.
Full dev smoke: passed.
Browser smoke: passed for /finance cash voucher prefill -> Record cash -> AP closeout voucher evidence.
Browser smoke AP: AP-SPM-S34-AP-SMOKE-0507060226-FINAL.
API evidence: cash-transactions query by AP number returned one posted cash_out supplier_payable allocation.
Screenshots:
- output/playwright/s36-finance-cash-voucher-created.png
- output/playwright/s36-finance-ap-voucher-closeout.png
```

---

## 5. Runtime Implementation Notes

Implemented branch scope:

```text
1. Finance factory AP closeout detects posted cash_out CashTransaction records allocated to supplier_payable by AP ID or AP number.
2. Paid factory final-payment AP records without a posted voucher show the current step as payment voucher evidence required.
3. The AP closeout surface exposes a prefilled cash transaction link for creating the voucher from AP, supplier, invoice, factory order, amount, and memo evidence.
4. Existing voucher evidence links from AP closeout to /finance?cash_q=:cashTransactionNo#cash-transactions.
5. Cash transaction allocation detail links supplier_payable allocations back to /finance?ap_q=:payableNo#supplier-payables.
6. Cash transaction form accepts cash_* URL query parameters for direction, counterparty, method, reference, amount, allocation target, and memo prefill.
7. Matched supplier invoice and AP request/approval/payment controls remain unchanged.
```

---

## 6. Known Limits

```text
Binary payment attachment upload/OCR remains out of scope.
Bank transfer export/API remains out of scope.
VAT/withholding/multi-currency factory payment treatment remains out of scope.
Supplier statement reconciliation remains out of scope.
Factory debit note / credit note settlement remains out of scope.
General ledger posting remains out of scope.
```
