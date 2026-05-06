# 134_ERP_Sprint35_Changelog_Factory_Final_Payment_Finance_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 35
Change type: Finance UI traceability and closeout guidance
Version: v1
Date: 2026-05-07
Status: Implemented locally; PR, CI, merge, deploy, and browser smoke pending

---

## 1. Summary

Sprint 35 adds Finance-side closeout guidance for factory final-payment AP records created by Sprint 34.

Target finance route:

```text
/finance?ap_q=:payableNo#supplier-payables
```

Target production source route:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

---

## 2. Planned Runtime Surface

Implementation scope:

```text
1. Detect factory final-payment AP records by subcontract_payment_milestone and subcontract_order source documents.
2. Show a Finance closeout card for factory final payment AP records.
3. List AP created, supplier invoice matched, payment requested, payment approved, and payment recorded steps.
4. Link Finance AP back to the source production factory order when source order id is present.
5. Preserve existing supplier invoice matching and AP payment-readiness backend gate.
6. Add focused frontend tests for helper behavior.
```

---

## 3. Guardrails

```text
Finance remains the payment execution surface.
Production remains the factory-order source surface.
Matched supplier invoice remains required before payment request, approval, or recording.
No email/Zalo/bank automation is included.
No v0.35 release tag is planned.
```

---

## 4. Verification Evidence

Runtime PR:

```text
Pending.
```

Local verification:

```text
git diff --check: passed
Targeted Finance web Vitest: supplierPayableFactoryCloseout, supplierPayableService, supplierInvoiceService passed
Full web Vitest: 62 files / 360 tests passed via D:\toolcache\node-v22.22.2-win-x64\node.exe
Web typecheck: tsc --noEmit passed
Web build: next build passed; Next emitted local SWC native-load warnings but completed successfully
Targeted API Finance tests: go test ./internal/modules/finance/application -run 'SupplierPayable|SupplierInvoice' -count=1 passed via D:\toolcache\go1.24.2\go\bin\go.exe
Full API tests: go test ./... passed
API vet: go vet ./... passed
OpenAPI contract check: passed
OpenAPI Redocly lint: pending GitHub CI because local pnpm/redocly CLI is unavailable
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
Supplier invoice attachment/OCR remains out of scope.
VAT/withholding/multi-currency factory payment treatment remains out of scope.
Payment voucher/export remains out of scope.
Bank transfer automation remains out of scope.
Factory settlement debit/credit notes remain out of scope.
```
