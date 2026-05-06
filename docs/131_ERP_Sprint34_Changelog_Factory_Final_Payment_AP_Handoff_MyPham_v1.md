# 131_ERP_Sprint34_Changelog_Factory_Final_Payment_AP_Handoff_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 34
Change type: Runtime API/UI bridge and traceability helper
Version: v1
Date: 2026-05-07
Status: Completed; runtime PR merged, CI green, dev deploy passed, and browser smoke passed

---

## 1. Summary

Sprint 34 connects external-factory final payment readiness to Finance/AP handoff.

Target production route:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

Target finance route:

```text
/finance?ap_q=:payableNo#supplier-payables
```

---

## 2. Planned Runtime Surface

Implementation scope:

```text
1. Expose supplier payable handoff identifiers in final-payment-ready API response.
2. Add production-facing AP handoff helper and UI copy.
3. Link production factory order detail to the matching Finance/AP record.
4. Preserve existing supplier invoice matching and AP payment-readiness gate.
5. Update OpenAPI schemas for the AP handoff response.
6. Add backend and frontend tests for handoff traceability.
```

---

## 3. Guardrails

```text
Production does not approve or record cash payment.
Finance/AP remains the payment execution surface.
Matched supplier invoice remains required before payment request, approval, or recording.
No v0.34 release tag is planned.
```

---

## 4. Verification Evidence

Runtime PR:

```text
PR: #608 Wire factory final payment AP handoff
Merge commit: 602a7354
```

Local verification:

```text
git diff --check: passed
Targeted Go tests: passed
Full API test: go test ./... passed
API vet: go vet ./... passed
Targeted web Vitest: subcontract AP handoff helper and subcontract service tests passed via D:\toolcache\node-v22.22.2-win-x64\node.exe
Web typecheck: passed after Next build generated .next/types
Web build: passed; Next emitted local SWC native-load warnings but completed successfully
OpenAPI contract check: passed
OpenAPI Redocly lint: pending GitHub CI because local pnpm/redocly CLI is unavailable
```

GitHub CI:

```text
api: passed
web: passed
openapi: passed
e2e: passed
required-api: passed
required-web: passed
required-openapi: passed
required-migration: passed
```

Dev deploy and smoke:

```text
deploy-dev-staging.sh dev: passed on 2026-05-07
Full dev smoke: passed
Browser smoke: passed with Chrome headless CDP
Smoke order: SCO-S34-AP-SMOKE-0507060226
Smoke AP: AP-SPM-S34-AP-SMOKE-0507060226-FINAL
Production screenshot: output/playwright/s34-factory-final-payment-ap-handoff.png
Finance screenshot: output/playwright/s34-finance-ap-handoff.png
```

---

## 5. Known Limits

```text
Supplier invoice attachment/OCR remains out of scope.
VAT/withholding/multi-currency final payment treatment remains out of scope.
Factory claim settlement accounting remains out of scope.
External factory payment notification remains out of scope.
```
