# 126_ERP_Coding_Task_Board_Sprint33_Factory_Claim_Final_Payment_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 33
Document role: Coding task board
Version: v1
Date: 2026-05-07
Status: Locked for Sprint 33 implementation

---

## 1. Sprint Goal

Sprint 33 closes the external-factory production loop after finished goods QC:

```text
Factory finished-goods QC closeout
-> Factory claim acknowledgement/resolution
-> Final payment readiness
-> Production timeline/tracker closeout
```

The user-facing route remains:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

The technical runtime continues to reuse the existing subcontract order domain because Phase 1 production is external-factory manufacturing, not internal MES/work-center production.

---

## 2. Scope

Runtime scope:

```text
1. Add API support to list factory claims for a subcontract/factory order.
2. Add API support to acknowledge an open factory claim.
3. Add API support to resolve an acknowledged/open factory claim.
4. Preserve final payment blocking while factory claims are open or acknowledged.
5. Allow final payment readiness after accepted QC and after partial-QC claim resolution.
6. Keep full QC fail blocked from final payment readiness until replacement/compensation flow is handled later.
7. Add production-facing closeout UI for claim status, claim action, and final payment readiness.
8. Add production timeline/tracker claim-resolution step.
9. Update OpenAPI and tests for claim/final payment behavior.
```

---

## 3. Out Of Scope

```text
Email/Zalo/factory portal delivery.
Factory digital signatures.
Supplier debit note accounting.
Replacement production order generation.
Claim settlement finance posting.
Internal factory/MES routing, work centers, labor, or machine costing.
v0.33 release tag.
```

---

## 4. Acceptance Criteria

```text
- Open factory claim blocks final payment readiness.
- Acknowledged factory claim still blocks final payment readiness.
- Resolved factory claim no longer blocks final payment readiness when accepted quantity exists.
- Full factory rejection remains blocked from final payment readiness.
- /production/factory-orders/:orderId shows a claim/final-payment closeout section.
- Timeline and execution tracker point claim/payment work to #factory-claim-final-payment-closeout.
- OpenAPI documents list, acknowledge, and resolve factory claim endpoints.
- Backend audit records claim acknowledgement and resolution.
- Tests cover open claim block, resolved claim allow, and full rejection block.
```

---

## 5. Verification Plan

```text
Local:
- git diff --check
- targeted Go test if Go toolchain exists
- targeted web Vitest if pnpm/node tooling exists

GitHub CI:
- required-api
- required-web
- required-openapi
- required-migration

Dev smoke after merge:
- deploy-dev-staging.sh dev
- make smoke-dev or equivalent deploy smoke
- browser smoke production factory order detail claim/final-payment closeout
```

---

## 6. Implementation Notes

The domain already models factory claim lifecycle:

```text
open -> acknowledged -> resolved
```

Sprint 33 connects that lifecycle to runtime API/UI and final payment readiness. It does not invent a new claim module.

Final payment guardrail:

```text
QC full pass: final payment can be marked ready.
QC partial pass: final payment remains blocked until factory claim is resolved.
QC full fail: final payment remains blocked; later sprint must decide replacement, credit, or cancellation flow.
```
