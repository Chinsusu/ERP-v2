# 44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 3 - Returns Reconciliation Core
Document role: Sprint 3 changelog, release note, and release-readiness evidence.

---

## 1. Sprint 3 Kickoff

Sprint 2 Order Fulfillment Core was promoted to `main` and tagged:

```text
v0.2.0-order-fulfillment-core
```

Sprint 3 task board:

```text
docs/43_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md
```

---

## 2. Sprint Goal

Sprint 3 adds the operations loop that happens after delivery handover:

```text
Return receiving
-> Return inspection
-> Return disposition
-> Return stock movement / quarantine
-> Stock count and adjustment
-> End-of-day reconciliation
-> Shift closing
-> Warehouse Daily Board
```

The sprint is demo-ready for warehouse return intake, QA inspection, reusable/non-reusable/QA-hold disposition, stock count variance handling, adjustment approval/posting, and shift closing with operational blockers.

---

## 3. Guardrails

- Return stock movement only goes through the movement service, not direct stock balance mutation.
- Only reusable returns become available stock.
- QA-hold returns stay quarantined and cannot be reserved/picked/sold.
- Stock count variance creates adjustment workflow; it does not edit balance directly.
- Shift closing is blocked while returns, adjustments, manifest issues, or reconciliation checklist items are unresolved.
- Sensitive actions have audit logs.
- Money, quantity, rate, and UOM values use decimal strings and base UOM rules from file 40.
- GitHub auto-review/auto-merge is not used; PRs are self-reviewed and manually merged.

---

## 4. Verification Note

GitHub Actions is currently blocked before job execution by account billing/spending-limit state. PR checks consistently fail with the annotation:

```text
The job was not started because recent account payments have failed or your spending limit needs to be increased.
```

Local verification is used as release evidence until the GitHub account blocker is fixed. Detailed S3-08-01 evidence is recorded in:

```text
docs/qa/S3-08-01_sprint3_release_pipeline_check.md
```

Production tagging must wait for cloud CI rerun and migration apply/rollback verification on an isolated PostgreSQL 16 instance.

---

## 5. Change Log

| Date | Change | Evidence |
|---|---|---|
| 2026-04-28 | Opened and corrected Sprint 3 task board, branch/merge rules, endpoint style, and dependency map | S3-00-02, PR #192 |
| 2026-04-28 | Delivered return reason and disposition master data | S3-01-01, PR #193 |
| 2026-04-28 | Delivered return receiving database model | S3-01-02, PR #194 |
| 2026-04-28 | Delivered return receiving scan API | S3-01-03, PR #195 |
| 2026-04-28 | Delivered return receiving scan UI | S3-01-04, PR #196 |
| 2026-04-28 | Added return receiving API tests | S3-01-05, PR #197 |
| 2026-04-28 | Delivered return inspection workflow | S3-02-01, PR #198 |
| 2026-04-28 | Delivered return disposition action | S3-02-02, PR #199 |
| 2026-04-28 | Delivered return inspection UI | S3-02-03, PR #200 |
| 2026-04-28 | Delivered return photo/attachment audit path | S3-02-04, PR #201 |
| 2026-04-28 | Added return inspection tests | S3-02-05, PR #202 |
| 2026-04-28 | Added no-direct-stock-mutation guardrail | S3-03-03, PR #203 |
| 2026-04-28 | Delivered return stock movement | S3-03-01, PR #204 |
| 2026-04-28 | Delivered quarantine return stock behavior | S3-03-02, PR #205 |
| 2026-04-28 | Added return movement audit coverage | S3-03-04, PR #206 |
| 2026-04-28 | Added return stock movement regression tests | S3-03-05, PR #207 |
| 2026-04-28 | Delivered inventory adjustment request workflow | S3-04-01, PR #208 |
| 2026-04-28 | Delivered stock count session workflow | S3-04-02, PR #209 |
| 2026-04-28 | Delivered variance approval and adjustment posting | S3-04-03, PR #210 |
| 2026-04-28 | Delivered stock count UI | S3-04-04, PR #211 |
| 2026-04-28 | Delivered adjustment approval UI | S3-04-05, PR #212 |
| 2026-04-28 | Added stock count/adjustment regression tests | S3-04-06, PR #213 |
| 2026-04-28 | Delivered shift closing model | S3-05-01, PR #214 |
| 2026-04-28 | Delivered end-of-day reconciliation service | S3-05-02, PR #215 |
| 2026-04-28 | Blocked shift closing with unresolved operational issues | S3-05-03, PR #216 |
| 2026-04-28 | Delivered shift closing UI | S3-05-04, PR #217 |
| 2026-04-28 | Added shift closing tests | S3-05-05, PR #218 |
| 2026-04-28 | Updated Warehouse Daily Board with returns/reconciliation data | S3-06-01, PR #219 |
| 2026-04-28 | Hardened Daily Board alerts and drill-down links | S3-06-02, PR #220 |
| 2026-04-28 | Added Daily Board source-data regression tests | S3-06-03, PR #221 |
| 2026-04-28 | Added return E2E smoke test | S3-07-01, PR #222 |
| 2026-04-28 | Added shift closing E2E smoke test | S3-07-02, PR #223 |
| 2026-04-29 | Added permission and audit regression for returns/closing | S3-07-03, PR #224 |
| 2026-04-29 | Added decimal/UOM regression for backend and frontend | S3-07-04, PR #225 |
| 2026-04-29 | Recorded Sprint 3 release pipeline evidence and blockers | S3-08-01, PR #226 |
| 2026-04-29 | Created this Sprint 3 release note | S3-08-02 |

---

## 6. Release Summary

Sprint 3 completes the returns and reconciliation core for the Phase 1 ERP prototype:

- warehouse staff can scan returned orders or tracking codes;
- returns are linked to order, tracking, customer, SKU, batch, and warehouse context where available;
- QA/warehouse users can inspect returned items and record condition evidence;
- disposition routes reusable, non-reusable, and QA-hold returns to different operational paths;
- reusable returns create controlled inbound stock movement;
- non-reusable returns do not affect available stock;
- QA-hold returns move to quarantine/HOLD;
- stock counts capture expected vs counted quantity by warehouse/location/SKU/batch;
- variances create adjustment requests with approval/reject/post flow;
- end-of-day reconciliation and shift closing block unresolved operational issues;
- Warehouse Daily Board shows return, stock count, adjustment, reconciliation, and closing signals.

---

## 7. Delivered Features

| Area | Delivered capability | Evidence |
|---|---|---|
| Return receiving | Master data, DB model, scan API, scan UI, duplicate/invalid/not-handed-over checks | S3-01 tasks |
| Return inspection | Condition workflow, disposition recommendation, inspection UI, attachment audit | S3-02 tasks |
| Return stock movement | Reusable restock, QA quarantine, non-reusable no available stock, audit | S3-03 tasks |
| Stock count | Count session by warehouse/location/SKU/batch with decimal/base UOM quantity | S3-04-02, S3-04-04 |
| Adjustment | Draft, submit, approve, reject, post, movement and audit path | S3-04-01, S3-04-03, S3-04-05 |
| Shift closing | End-of-day reconciliation, blockers, exception note, close audit | S3-05 tasks |
| Daily Board | Return pending, QA hold, adjustment pending, stock count variance, shift closing status | S3-06 tasks |
| QA/E2E | Return happy path, shift closing happy path, permission/audit, decimal/UOM regression | S3-07 tasks |
| Release evidence | Local pipeline evidence, cloud CI blocker, migration risk captured | S3-08-01 |

---

## 8. Bugs / Risks Closed

| Risk | Sprint 3 closure |
|---|---|
| Returned goods entering sellable stock without inspection | Return receipt defaults to inspection flow and disposition controls movement. |
| Damaged/non-reusable returns increasing available stock | Non-reusable disposition routes away from available stock. |
| QA-hold returns being sellable | QA-hold disposition records quarantine/HOLD status. |
| Stock variance corrected by direct balance edit | Adjustment request and movement workflow replaces direct stock mutation. |
| Shift closed with unresolved returns or adjustments | Shift closing checks reconciliation blockers. |
| Sensitive operations missing audit evidence | Regression tests cover return receipt, inspection, disposition, movement, stock count, adjustment, and shift close audit. |
| Decimal/UOM drift in stock and money values | Regression tests cover backend decimal arithmetic, UOM conversion to base UOM, and frontend VND/quantity/rate formatting. |
| Daily Board displaying stale prototype-only values | Board service tests verify source-backed return, movement, adjustment, and closing data. |

---

## 9. Known Issues / Limitations

| Issue | Current impact | Mitigation / next action |
|---|---|---|
| GitHub Actions blocked by billing/spending-limit | Cloud CI cannot be release evidence today. | Fix billing and rerun all required checks before production tag. |
| Migration apply/rollback not executed in S3-08-01 local evidence | Runtime PostgreSQL migration safety still needs isolated DB verification. | Run migration-ci after GitHub billing is fixed or on a machine with Docker/psql. |
| Several Sprint 3 stores remain prototype/in-memory | Demo state resets on restart and is not production persistence for every object. | Persist remaining runtime stores in hardening sprint. |
| Real carrier/marketplace return integrations are not wired | Returns are based on prototype order/tracking data, not carrier webhooks. | Add carrier/marketplace return adapters later. |
| File upload storage is prototype-level | Attachments prove metadata/audit path but not production object lifecycle. | Wire S3/MinIO object storage and retention rules. |
| Role model has no separate warehouse manager role yet | Warehouse Lead currently acts as operational manager in regression coverage. | Add manager role only when approval matrix requires it. |
| Dev/prod release tag is intentionally withheld | `v0.3.0-returns-reconciliation-core` is not created yet. | Tag only after cloud CI and migration runtime checks are green. |

---

## 10. Verification Evidence

Local verification used because cloud CI is currently blocked:

```text
go test ./cmd/api -run TestReturnReceiptReusableHappyPathSmoke -count=1
go test ./cmd/api -run TestShiftClosingCleanStockCountHappyPathSmoke -count=1
go test ./cmd/api -run TestReturnsAndShiftClosingPermissionAuditRegressionSmoke -count=1
go test ./internal/shared/decimal ./internal/modules/masterdata/application -count=1
go test ./...
go vet ./...
go build ./cmd/api ./cmd/worker
pnpm --filter web test
pnpm --filter web lint
pnpm --filter web typecheck
pnpm --filter web build
pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml
pnpm dlx openapi-typescript packages/openapi/openapi.yaml -o %TEMP%/erp-openapi-schema-s3-08-01.ts
git diff --check
```

Notes:

- Web build passed with the known Windows Next SWC DLL warning.
- OpenAPI validation passed with the existing proprietary license warning and local Node version warning.
- Migration pair static check found 11 up migrations and 11 down migrations.
- Runtime migration apply/rollback was not executed locally because Docker/psql is unavailable on this machine.

---

## 11. Operations / Support Notes

- Do not correct return, stock movement, adjustment, stock count, or reconciliation state directly in the database.
- Every operational correction needs an approval/audit trail.
- Treat end-of-day reconciliation failures as warehouse operations incidents, not only technical bugs.
- Return receiving and QA evidence are sensitive operational records; do not log file contents or customer-sensitive payloads.
- Daily Board alert drill-down links should be used during shift handoff to resolve pending returns, adjustments, stock counts, and closing blockers.

---

## 12. Release Readiness

Recommended release tag after blockers are cleared:

```text
v0.3.0-returns-reconciliation-core
```

Current readiness:

```text
Dev/main merge: done
Local verification: done
Release note: done
Cloud CI: blocked by GitHub billing/spending-limit
Migration runtime apply/rollback: not verified locally
Production tag: hold
```

---

## 13. Next Sprint Recommendations

Recommended next sprint:

```text
Sprint 4 - Purchase Order + Inbound QC Full Flow
```

Suggested order:

1. Fix GitHub Actions billing/spending-limit blocker and rerun Sprint 3 release gates.
2. Run PostgreSQL migration apply/rollback before tagging `v0.3.0-returns-reconciliation-core`.
3. Persist remaining returns/reconciliation runtime state to PostgreSQL-backed stores.
4. Build Purchase Order, goods receiving, and inbound QC full flow.
5. Connect inbound QC pass/fail to stock availability and batch status.
6. Add real object storage for return/QC attachments.
7. Add carrier/marketplace return webhook adapter boundaries.
8. Add UAT scripts for warehouse receiving, QA, stock count, adjustment approval, and shift close.
