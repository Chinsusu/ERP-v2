# 48_ERP_Sprint5_Changelog_Subcontract_Manufacturing_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 5 - Subcontract Manufacturing / Gia cong ngoai
Date: 2026-04-29
Status: Dev/main merged; production tag hold until GitHub Actions billing blocker is cleared

---

## 1. Summary

Sprint 5 completed the subcontract manufacturing core:

```text
Subcontract order
-> factory confirmation
-> material / packaging issue to factory
-> handover evidence
-> sample submission and decision
-> mass production start
-> finished goods receipt into QC hold
-> QC pass / fail / partial accept
-> controlled available-stock release
-> factory claim and SLA tracking
-> guarded final payment readiness
-> subcontract operations UI and daily board signals
```

The implementation keeps the Sprint 5 guardrails:

```text
- Materials issued to factory are controlled by stock movement service.
- Finished goods received from factory enter QC hold first.
- QC PASS is the only path that increases available finished-goods stock.
- QC FAIL creates no available stock.
- Partial accept releases only pass quantity and keeps non-pass quantity traceable through an open factory claim.
- Final payment readiness is blocked while a factory claim remains open.
```

## 2. Merged PRs

Planning and subcontract order foundation:

```text
#267 docs(S5-00-00): add subcontract manufacturing task board
#268 feat(S5-01-01): add subcontract order domain model
#269 feat(S5-01-02): add subcontract order migration
#270 feat(S5-01-03): add subcontract order API foundation
#271 feat(S5-01-04): wire subcontract order UI to API
```

Material issue to factory:

```text
#272 feat(S5-02-01): add subcontract material transfer model
#273 feat(S5-02-02): add subcontract material issue API
#274 feat(S5-02-03): add subcontract material issue UI
```

Sample approval gate:

```text
#275 feat(S5-03-01): add subcontract sample approval model
#276 feat(S5-03-02): add subcontract sample approval API
#277 feat(S5-03-03): add subcontract sample approval UI
```

Finished goods receipt and factory claims:

```text
#278 feat(S5-04-01): add finished goods receipt model
#279 feat(S5-04-02): add finished goods receipt API
#280 feat(S5-04-03): receive finished goods UI
#281 feat(S5-04-04): add factory claim service
#282 feat(S5-04-05): add factory claim UI
```

Payment milestone and daily board:

```text
#283 feat(S5-05-01): add payment milestone model
#284 feat(S5-05-02): add payment milestone API/UI
#285 feat(S5-06-01): add subcontract daily board metrics
#286 feat(S5-06-02): show subcontract daily board metrics
```

OpenAPI, generated client, and contract gate:

```text
#287 docs(S5-07-01): add subcontract OpenAPI contract
#288 chore(S5-07-02): regenerate frontend API schema
#289 test(S5-07-03): cover subcontract OpenAPI contract
```

Regression and E2E gates:

```text
#290 test(S5-08-01): subcontract permission/audit regression
#291 test(S5-08-02): subcontract material issue e2e
#292 test(S5-08-03): subcontract sample rejection e2e
#293 feat(S5-08-04): add subcontract finished goods pass flow
#294 test(S5-08-05): add subcontract finished goods fail e2e
#295 test(S5-08-06): add subcontract partial accept e2e
```

## 3. Verification

Backend verification run on dev server:

```text
go test ./cmd/api -run TestSubcontractFinishedGoodsPassE2ESmoke -count=1
go test ./cmd/api -run TestSubcontractFinishedGoodsFailE2ESmoke -count=1
go test ./cmd/api -run TestSubcontractFinishedGoodsPartialAcceptE2ESmoke -count=1
go test ./cmd/api -run TestSubcontractMaterialIssueE2ESmoke -count=1
go test ./cmd/api -run TestSubcontractSampleRejectionE2ESmoke -count=1
go test ./cmd/api -run TestSubcontractDeniedActionsHaveNoSideEffects -count=1
go test ./...
go vet ./...
```

Frontend and OpenAPI verification run on dev server:

```text
pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml
pnpm openapi:contract
pnpm openapi:generate
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
```

Latest observed OpenAPI contract result:

```text
Sprint 4/5 OpenAPI contract check passed: 42 routes and 23 envelopes.
```

Latest observed frontend test result:

```text
26 test files passed.
171 tests passed.
```

Local diff verification:

```text
git diff --check
```

## 4. Dev Deployment Status

Dev server deployment was run after Sprint 5 runtime changes and after the final partial-accept merge.

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
2026-04-29T19:03:21Z
```

The GHCR image pull warnings are still expected in the current dev setup; the deploy script builds the service images from source and smoke checks pass.

## 5. Release Gate Status

Green:

```text
Task branch -> PR -> manual self-review -> manual merge workflow
Dev-server backend checks
Dev-server frontend checks
OpenAPI lint, contract, and generated schema
Dev deploy and smoke checks
Sprint 5 E2E pass/fail/partial scenarios
Permission and audit regression
```

Blocked:

```text
GitHub Actions cloud CI is blocked by account billing/spending-limit.
Failed jobs exit in a few seconds and gh run view --log-failed returns log not found for completed failures.
Production tag v0.5.0-subcontract-manufacturing-core is on hold until CI can be rerun green.
```

## 6. Known Carry-Forward

```text
S5 release gate: repo owner must clear GitHub Actions billing/spending-limit.
After billing is fixed: rerun full CI on main.
After CI is green: tag v0.5.0-subcontract-manufacturing-core.
```

No Sprint 6 coding should be treated as production-ready until the Sprint 5 release gate is cleared.

## 7. Release Gate Attempt - 2026-04-30

Release gate was rechecked on `main` commit:

```text
d108fafc74c3e8613f99b6d304669d790c1e9d72
```

GitHub Actions rerun:

```text
Workflow run: required-ci
Run ID: 25128437655
Result: failure
Jobs: required-api, required-web, required-openapi, required-migration
Observed behavior: jobs failed in a few seconds with no step logs.
Log fetch result: gh run view --log-failed returned log not found.
Direct job log fetch result: GitHub Actions log blob returned BlobNotFound.
```

Migration runtime gate was rechecked on an isolated PostgreSQL 16 container on the dev server:

```text
All up migrations applied successfully.
All down migrations rolled back successfully.
Result: migration-apply-rollback-pass
```

Gate decision:

```text
Sprint 5 release gate remains blocked.
Production tag v0.5.0-subcontract-manufacturing-core was not created.
Reason: cloud CI is not green because GitHub Actions is still blocked before job steps can run.
```
