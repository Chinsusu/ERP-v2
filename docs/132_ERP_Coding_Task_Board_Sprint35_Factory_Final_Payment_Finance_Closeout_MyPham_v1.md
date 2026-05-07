# 132_ERP_Coding_Task_Board_Sprint35_Factory_Final_Payment_Finance_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 35
Document role: Coding task board
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 35 implementation

---

## 1. Sprint Goal

Sprint 35 closes the Finance side of the factory final-payment flow after Sprint 34 creates the AP handoff:

```text
Factory final-payment AP handoff
-> Finance AP opens with factory source evidence
-> Supplier invoice is created/matched to the AP
-> Payment request
-> Payment approval
-> Payment recording
-> Factory final-payment AP closeout evidence
```

Finance remains the payment execution surface:

```text
/finance?ap_q=:payableNo#supplier-payables
```

Production remains the factory-order source surface:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

---

## 2. Scope

Runtime scope:

```text
1. Add Finance-side factory final-payment closeout guidance for AP records sourced from external-factory production.
2. Detect AP source evidence from subcontract_payment_milestone and subcontract_order source documents.
3. Show a Finance closeout checklist: AP created, supplier invoice matched, payment requested, payment approved, payment recorded.
4. Link Finance AP detail back to the source production factory order when AP line source evidence contains subcontract_order.
5. Preserve the backend hard gate that payment request/approval/recording require a matched supplier invoice.
6. Add focused frontend tests for the Finance closeout helper.
7. Update README and master docs index with Sprint 35 scope.
```

---

## 3. Out Of Scope

```text
Email/Zalo/factory portal payment notification.
Bank transfer files or direct bank API.
Accounting journal posting.
VAT, withholding tax, multi-currency settlement, or payment tolerance policy.
Supplier invoice OCR or attachment extraction.
Factory debit note / credit note settlement.
Internal MES or work-center production.
v0.35 release tag.
```

---

## 4. Acceptance Criteria

```text
- Finance AP panel recognizes factory final-payment AP records by subcontract source documents.
- Factory final-payment AP detail shows factory order no, milestone no, payable no, invoice match state, and payment closeout steps.
- Finance AP detail links back to /production/factory-orders/:orderId#factory-claim-final-payment-closeout when source order id exists.
- A matched supplier invoice moves the Finance closeout current step from invoice matching to payment request.
- AP status payment_requested moves the current step to payment approval.
- AP status payment_approved or partially_paid moves the current step to payment recording.
- AP status paid marks payment recording complete.
- Non-factory AP documents do not show factory closeout guidance.
- Existing AP payment readiness gate remains unchanged.
```

---

## 5. Verification Plan

```text
Local:
- git diff --check
- targeted web Vitest for supplier payable closeout helper, supplier payable service, and supplier invoice service
- web typecheck/build if local tooling is available
- targeted Go tests for Finance/AP invoice/payment gate
- full API Go tests if runtime code changes or before PR

GitHub CI:
- required-api
- required-web
- required-openapi
- required-migration

Dev smoke after merge:
- deploy-dev-staging.sh dev
- full dev smoke
- browser smoke Finance AP closeout from the Sprint 34 smoke AP or a newly created factory final-payment AP
```

---

## 6. Implementation Notes

The backend already enforces:

```text
AP payment request/approval/recording require a matched supplier invoice.
```

Sprint 35 should not create another payment workflow. It should make the existing Finance workflow readable for the factory final-payment case.

Factory AP source evidence:

```text
supplier_payable.source_document.type = subcontract_payment_milestone
supplier_payable.line.source_document.type = subcontract_order
```

The Finance closeout UI is guidance and traceability. It must not bypass the Supplier Invoice/AP payment backend gate.
