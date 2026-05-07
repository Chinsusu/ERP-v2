# 134_ERP_Sprint35_Changelog_Factory_Final_Payment_Finance_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 35
Change type: Finance UI traceability and closeout guidance
Version: v1
Date: 2026-05-07
Status: Completed; runtime PRs merged, CI green, dev deploy passed, and S35 Finance browser smoke passed

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

## 2. Runtime Surface

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
PR #610 Add S35 factory payment finance closeout
- Branch: codex/s35-finance-ap-invoice-payment-closeout
- Commit: 49361e53917e01e8a221c45f227c72afa8be7ee1
- Merge commit: 68b4d3d5

PR #611 Allow factory AP supplier invoice sources
- Branch: codex/s35-fix-factory-supplier-invoice-source
- Commit: 7bda5a58a9dbddab66e55280f3eab6be481d1c8f
- Merge commit: 64851338
```

Local verification:

```text
PR #610:
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

PR #611:
Red test first reproduced the S35 smoke blocker: factory final-payment supplier invoice source documents were rejected.
go test ./internal/modules/finance/domain ./internal/modules/finance/application -run 'SupplierInvoice.*Factory|FactoryFinalPayment' -count=1: passed
go test ./internal/modules/finance/domain ./internal/modules/finance/application -run 'SupplierInvoice' -count=1: passed
go test ./...: passed
go vet ./...: passed
node packages/openapi/contract-check.mjs: passed
git diff --check: passed
```

GitHub CI:

```text
PR #610: e2e, required-api, required-web, required-openapi, required-migration, and web checks passed.
PR #611: api, e2e, required-api, required-web, required-openapi, and required-migration checks passed.
```

Dev deploy and smoke:

```text
Dev deploy after PR #611 merge:
- main commit: 64851338
- command: ./infra/scripts/deploy-dev-staging.sh dev
- result: passed
- full ERP dev smoke: passed

Target S35 API smoke:
- AP: AP-SPM-S34-AP-SMOKE-0507060226-FINAL
- AP source document: subcontract_payment_milestone
- AP line source document: subcontract_order
- Supplier invoice: INV-S35-13609491
- Supplier invoice status: matched
- Payment flow: request payment -> approve payment -> record payment
- Final AP status: paid
- Final outstanding amount: 0.00

Target S35 Finance browser smoke:
- Route: /finance?ap_q=AP-SPM-S34-AP-SMOKE-0507060226-FINAL#supplier-payables
- Factory closeout card rendered.
- Back link rendered to /production/factory-orders/sco-s34-ap-smoke-0507060226#factory-claim-final-payment-closeout.
- Closeout checklist showed 5 completed steps.
- Screenshot: output/playwright/s35-finance-factory-payment-closeout-paid.png
- Browser automation caveat: local npx/Playwright CLI unavailable, so the browser smoke used Chrome headless CDP.
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
