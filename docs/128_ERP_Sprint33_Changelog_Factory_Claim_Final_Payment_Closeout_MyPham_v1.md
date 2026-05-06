# 128_ERP_Sprint33_Changelog_Factory_Claim_Final_Payment_Closeout_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: 33
Change type: Runtime API/UI bridge and closeout helper
Version: v1
Date: 2026-05-07
Status: Completed and merged; dev deploy and browser smoke passed

---

## 1. Summary

Sprint 33 adds the production-facing closeout step for factory claim resolution and final payment readiness after finished-goods QC.

The target user-facing route is:

```text
/production/factory-orders/:orderId#factory-claim-final-payment-closeout
```

---

## 2. Planned Runtime Surface

Implementation scope:

```text
1. Backend API to list factory claims for a factory/subcontract order.
2. Backend API to acknowledge a factory claim.
3. Backend API to resolve a factory claim.
4. OpenAPI contract coverage for claim list and claim decision endpoints.
5. Web service bridge and prototype fallback for claim acknowledge/resolve.
6. Production factory order detail closeout section for claim/payment gate.
7. Timeline and execution tracker claim-resolution step before final payment readiness.
8. Tests for final payment blocking while claim is open and release after resolution.
```

---

## 3. Guardrails

```text
Open or acknowledged factory claim blocks final payment.
Resolved factory claim can allow final payment only when accepted finished goods exist.
Full QC fail remains blocked from final payment readiness.
Email, Zalo, factory portal/API delivery, debit note, replacement order, and internal MES are out of scope.
No v0.33 release tag is planned.
```

---

## 4. Verification Evidence

Runtime PR:

```text
PR: #606 Add factory claim final payment closeout
Runtime branch commit: 75b36bd8
Manual merge commit: 5ac8a1e
```

Local verification:

```text
git diff --check: passed
apps/api gofmt -l .: clean
apps/api go test ./...: passed
Local web/OpenAPI commands: not run on workstation because local Node/pnpm/npx tooling is unavailable; verified by GitHub CI.
```

GitHub CI:

```text
PR #606 checks passed:
- api
- web
- openapi
- e2e
- required-api
- required-web
- required-openapi
- required-migration
```

Dev deploy and smoke:

```text
deploy-dev-staging.sh dev: passed on dev server
Full ERP dev smoke: passed
Subcontract claim smoke endpoint: 200
Browser smoke: passed for /production/factory-orders/sco-s16-08-03-smoke-0068#factory-claim-final-payment-closeout
Screenshot evidence: output/playwright/s33-factory-claim-final-payment-closeout.png
```

---

## 5. Known Limits

```text
Factory claim settlement accounting remains out of scope.
Replacement/rework production order generation remains out of scope.
Manual factory acknowledgement/resolution is recorded inside ERP; external message delivery is not automated.
```
