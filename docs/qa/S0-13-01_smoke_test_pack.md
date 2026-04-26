# S0-13-01 Smoke Test Pack

Primary Ref: docs/24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md
Task Ref: docs/37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md#s0-13-01-smoke-test-pack

## Scope

This pack is the Sprint 0 daily smoke baseline for Dev and Staging.

It covers:

- Login and RBAC session readiness.
- API health and readiness checks.
- Master data route and seed data presence.
- Stock movement audit path.
- Shipping manifest scan handover path.

## Preconditions

- Dev or Staging stack is deployed from `infra/scripts/deploy-dev-staging.sh`.
- Smoke seed data from `tools/seed/smoke/sprint0_smoke_seed.json` is available to QA.
- API base URL points to the target environment.
- Frontend is reachable through the reverse proxy.

## Checklist

| ID | Area | Check | Expected result | Automation |
|---|---|---|---|---|
| S0SMK-001 | Healthcheck | Call `/healthz`, `/readyz`, and `/api/v1/health`. | API returns success envelopes and ready status. | `TestSprint0APISmokePack` |
| S0SMK-002 | Login | Sign in with the smoke admin user. | Access token is issued and `/api/v1/me` returns ERP Admin. | `TestSprint0APISmokePack` |
| S0SMK-003 | Master data | Open Master Data, SKU / Batch, Supplier / Factory, and Customer routes. | ERP Admin can access all master data shells. | `sprint0Smoke.test.ts` |
| S0SMK-004 | Stock movement | Submit the smoke adjustment movement. | Movement is recorded and an audit log row is written. | `TestSprint0APISmokePack` |
| S0SMK-005 | Scan handover | Scan the packed shipment code in the smoke manifest. | Scan result is `MATCHED`, missing count reaches zero, and audit log exists. | `TestSprint0APISmokePack` |

## Negative Smoke

| ID | Area | Check | Expected result |
|---|---|---|---|
| S0SMK-N01 | Scan handover | Scan a duplicate code. | API returns duplicate scan warning. |
| S0SMK-N02 | Scan handover | Scan a code from another manifest. | API returns manifest mismatch warning. |
| S0SMK-N03 | Scan handover | Scan an unknown code. | API returns not found warning. |
| S0SMK-N04 | RBAC | Warehouse staff opens master data route. | Route is not visible to that role. |

## Commands

```bash
make smoke-test
```

CI also runs the same sample smoke tests from `.github/workflows/e2e-ci.yml`.

## Exit Criteria

- API smoke test passes.
- Frontend smoke test passes.
- Checklist has no P0 failure.
- Any warning result is either expected negative smoke or logged as a defect.
