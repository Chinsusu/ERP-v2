# 50_ERP_Sprint6_Changelog_Finance_Lite_COD_AR_AP_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 6 - Finance Lite / COD / AR / AP Core
Date: 2026-04-30
Status: Dev/main merged; production tag hold until GitHub Actions billing blocker is cleared

---

## 1. Summary

Sprint 6 completed the finance-lite core:

```text
Sales / COD source documents
-> customer receivable
-> carrier COD remittance
-> discrepancy trace when remitted amount differs
-> cash receipt allocation
-> receivable close or disputed status

Purchase receiving / subcontract acceptance
-> supplier or factory payable
-> payment request
-> payment approval
-> cash payment allocation
-> payable close

Finance events
-> permission guard
-> audit trail
-> OpenAPI contract
-> Finance Lite dashboard
```

The implementation keeps the Sprint 6 guardrails:

```text
- VND remains the base currency.
- Money values use decimal strings at API/UI boundaries.
- No float/double is used for money, quantity, rate, or percentage.
- COD remittance is traceable before cash is treated as received.
- Customer receivables trace to source documents.
- Supplier and factory payables trace to accepted operational documents.
- Payment requires approval before payment recording.
- Finance mutations require finance permissions and write audit events.
- Sprint 6 does not introduce full general ledger posting.
```

## 2. Merged PRs

Planning and finance foundation:

```text
#298 docs(S6-00-00): add finance lite task board
#299 feat(S6-01-01): add finance action permissions
#300 feat(S6-01-02): add finance money status foundation
#301 feat(S6-01-03): add finance audit event conventions
```

Customer receivables:

```text
#302 feat(S6-02-01): add customer receivable domain model
#303 feat(S6-02-02): add customer receivable API
#304 feat(S6-02-03): add customer receivable UI
```

COD remittance and reconciliation:

```text
#305 feat(S6-03-01): add COD remittance domain model
#306 feat(S6-03-02): add COD reconciliation API
#307 feat(S6-03-03): add COD reconciliation UI
```

Supplier payable:

```text
#308 feat(S6-04-01): add supplier payable domain model
#309 feat(S6-04-02): add supplier payable API
#310 feat(S6-04-03): add supplier payable UI
```

Subcontract payable and payment approval:

```text
#311 feat(S6-05-01): create AP from subcontract final payment
#312 feat(S6-05-02): add supplier payable payment approval service
#313 feat(S6-05-03): add supplier payable payment approval API UI
```

Cash and dashboard:

```text
#314 feat(S6-06-01): add cash transaction domain model
#315 feat(S6-06-02): add cash transaction API UI
#316 feat(S6-07-01): add finance dashboard metrics
#317 feat(S6-07-02): add finance dashboard UI
```

OpenAPI, generated client, and contract gate:

```text
#318 test(S6-08): extend finance OpenAPI contract check
```

Regression and E2E gates:

```text
#319 test(S6-09-01): cover finance permission audit regressions
#320 test(S6-09-02): add COD happy path finance e2e
#321 test(S6-09-03): add COD discrepancy finance e2e
#322 test(S6-09-04): add supplier payable payment e2e
#323 test(S6-09-05): add subcontract payable payment e2e
```

## 3. Verification

Backend verification run on a clean dev-server clone of `main`:

```text
go test ./...
go vet ./...
```

Focused Sprint 6 backend checks run during task PRs:

```text
go test ./cmd/api -run "TestFinance(DeniedMutationsDoNotWriteAuditLogs|MutatingHandlersWriteExpectedAuditLogs)" -count=1
go test ./cmd/api -run "Test(Finance|CustomerReceivable|SupplierPayable|CODRemittance|CashTransaction)" -count=1
go test ./cmd/api -run "TestCODHappyPathFinanceE2ESmoke" -count=1
go test ./cmd/api -run "TestCODDiscrepancyFinanceE2ESmoke" -count=1
go test ./cmd/api -run "TestSupplierPayablePaymentFinanceE2ESmoke" -count=1
go test ./cmd/api -run "TestSubcontractPayablePaymentFinanceE2ESmoke" -count=1
```

OpenAPI verification run on a clean dev-server clone of `main`:

```text
pnpm openapi:validate
pnpm openapi:contract
pnpm openapi:generate
git diff --exit-code apps/web/src/shared/api/generated/schema.ts
```

Latest observed OpenAPI contract result:

```text
Sprint 4/5/6 OpenAPI contract check passed: 64 routes and 35 envelopes.
```

`pnpm openapi:validate` passed with the existing license warning:

```text
License object should contain one of the fields: url, identifier.
```

Frontend verification run on a clean dev-server clone of `main`:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
```

Latest observed frontend test result:

```text
31 test files passed.
194 tests passed.
```

Migration runtime verification run on an isolated PostgreSQL 16 container on the dev server:

```text
All up migrations applied successfully.
All down migrations rolled back successfully.
Result: migration-apply-rollback-pass
```

Local diff verification for this changelog PR:

```text
git diff --check
```

## 4. Dev Deployment Status

Dev server deployment was run after Sprint 6 runtime changes and after the final subcontract payable/payment E2E merge.

Latest deployed `main` commit:

```text
027b36c8ea48acb9dd1ccd24f00efe4186721601
```

Latest deploy result:

```text
Dev source build completed for api, web, and worker.
Database seed/migration step completed.
Internal smoke passed.
Host smoke passed.
ERP dev deployment finished.
```

Latest API health timestamp observed during deploy:

```text
2026-04-30T05:23:11Z
```

The GHCR image pull warnings are still expected in the current dev setup; the deploy script builds the service images from source and smoke checks pass.

## 5. Release Gate Status

Green:

```text
Task branch -> PR -> manual self-review -> manual merge workflow
Dev-server backend checks
Dev-server frontend checks
OpenAPI validate, contract, generated schema check
PostgreSQL 16 migration apply/rollback
Dev deploy and smoke checks
Sprint 6 COD happy path and discrepancy E2E
Sprint 6 supplier payable/payment E2E
Sprint 6 subcontract payable/payment E2E
Finance permission and audit regression
```

Blocked:

```text
GitHub Actions cloud CI is blocked before workflow steps run.
Latest main required-ci run failed with zero job steps:
  Run ID: 25148892362
  Jobs: required-api, required-web, required-openapi, required-migration
Latest main api-ci run failed with zero job steps:
  Run ID: 25148892344
Latest main openapi-ci run failed with zero job steps:
  Run ID: 25148892342
Direct GitHub Actions job log fetch returned BlobNotFound.
Production tag v0.6.0-finance-lite-cod-ar-ap-core was not created.
```

## 6. Known Carry-Forward

```text
Repo owner must clear the GitHub Actions billing/spending-limit blocker.
After billing is fixed: rerun full GitHub CI on main.
After CI is green: create tag v0.6.0-finance-lite-cod-ar-ap-core.
Sprint 5 production tag remains blocked for the same hosted CI reason.
Sprint 6 is dev-verified but not production-tagged until cloud CI is green.
```

Sprint 7 development may start from `main`, but no Sprint 6 production release should be claimed until the hosted CI blocker is cleared and the release tag is created.
