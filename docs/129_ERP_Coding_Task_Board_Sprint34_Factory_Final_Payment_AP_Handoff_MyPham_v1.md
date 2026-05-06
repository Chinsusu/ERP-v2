# 129_ERP_Coding_Task_Board_Sprint34_Factory_Final_Payment_AP_Handoff_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 34
Document role: Coding task board
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 34 implementation

---

## 1. Sprint Goal

Sprint 34 connects factory final payment readiness to the existing Finance/AP control flow:

```text
Factory order final payment ready
-> Supplier payable for final factory payment
-> Finance AP detail
-> Supplier invoice matching
-> AP payment request / approval / payment recording
```

The user-facing production route remains:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

Finance remains the payment execution surface:

```text
/finance#supplier-payables
```

---

## 2. Scope

Runtime scope:

```text
1. Expose supplier payable handoff identifiers from the final-payment-ready API response.
2. Show AP handoff evidence in the production factory order detail after final payment readiness.
3. Add a production-to-finance deep link that opens Finance with the created AP selected/searchable.
4. Keep AP payment execution inside existing Finance supplier payable and supplier invoice matching flow.
5. Add helper tests for AP handoff status/link behavior.
6. Add backend response/handler tests for supplier payable handoff traceability.
7. Update OpenAPI and Sprint 34 docs/evidence.
```

---

## 3. Out Of Scope

```text
Email/Zalo/factory portal payment submission.
Bank file generation.
Accounting journal posting.
VAT, withholding tax, or multi-currency final payment matching.
Factory debit note / credit note settlement.
Replacement or rework production order generation.
Internal MES or factory work-center production.
v0.34 release tag.
```

---

## 4. Acceptance Criteria

```text
- Mark final payment ready creates or exposes the source-linked supplier payable id/no.
- Production factory order detail shows the AP handoff state after final payment readiness.
- Production factory order detail links to Finance supplier payables using the created AP no/id.
- Reloaded final-payment-ready factory orders can still open Finance by fallback AP search on the factory order no.
- Finance supplier payables auto-filter/select the AP from the production link.
- Finance AP still requires matched supplier invoice before request/approve/record payment.
- API response/OpenAPI include supplier_payable handoff evidence for final payment readiness.
- Backend tests cover the final-payment-ready response exposing AP handoff identifiers.
- Frontend tests cover production AP handoff helper behavior.
```

---

## 5. Verification Plan

```text
Local:
- git diff --check
- targeted Go test for subcontract final-payment AP handoff
- targeted web Vitest for AP handoff helper if local Node/pnpm tooling is available

GitHub CI:
- required-api
- required-web
- required-openapi
- required-migration

Dev smoke after merge:
- deploy-dev-staging.sh dev
- full dev smoke
- browser smoke production factory order detail -> AP handoff link -> Finance AP selection
```

---

## 6. Implementation Notes

The backend already creates a supplier payable from the final-payment milestone through the subcontract payable adapter.
Sprint 34 should not create a second AP mechanism.

Expected contract:

```text
POST /api/v1/subcontract-orders/{id}/mark-final-payment-ready
returns:
  subcontract_order
  milestone
  supplier_payable:
    payable_id
    payable_no
    audit_log_id
```

The production page uses that handoff evidence to link the operator to Finance/AP.
Finance keeps the payment gate from files 96 and 97:

```text
AP cannot request/approve/record payment until a matched supplier invoice exists.
```
