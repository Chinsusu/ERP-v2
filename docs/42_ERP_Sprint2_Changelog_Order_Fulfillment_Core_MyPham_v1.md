# 42_ERP_Sprint2_Changelog_Order_Fulfillment_Core_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Sprint:** Sprint 2 — Order Fulfillment Core  
**Vai trò tài liệu:** Changelog / release note nhánh triển khai Sprint 2.

---

## 1. Sprint 2 Kickoff

Sprint 1 foundation đã được merge/promote lên `main` và tag:

```text
v0.1.0-foundation
```

Sprint 2 branch:

```text
sprint/2-order-fulfillment-core
```

Sprint 2 task board:

```text
docs/41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md
```

---

## 2. Sprint Goal

Sprint 2 tập trung build luồng order fulfillment lõi:

```text
Sales Order
→ Reserve Stock
→ Pick
→ Pack
→ Carrier Manifest
→ Scan Handover ĐVVC
→ Warehouse Daily Board cập nhật dữ liệu thật
```

Cuối Sprint 2 phải demo được vòng đời đơn hàng từ tạo đơn đến bàn giao cho đơn vị vận chuyển, gồm cả case thiếu đơn khi quét bàn giao.

---

## 3. Guardrails

- Không reserve/pick/pack batch QC `HOLD` hoặc `FAIL`.
- Không update tồn kho trực tiếp ngoài stock ledger/reservation service.
- Không confirm handover khi manifest chưa đủ đơn hợp lệ.
- Không bỏ audit log cho confirm/reserve/pick/pack/handover.
- Không làm UI lệch style đã chốt trong file 39.

---

## 4. Verification Note

GitHub Actions đang bị chặn bởi billing/spending limit của account, nên CI cloud chưa thể dùng làm evidence. Foundation checkpoint dùng local verification thay thế:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
pnpm openapi:validate
pnpm smoke:test
go test ./...
go vet ./...
```

Khi GitHub billing được xử lý, cần rerun CI cho các branch/PR Sprint 2.

---

## 5. Change Log

| Date | Change | Evidence |
|---|---|---|
| 2026-04-28 | Opened Sprint 2 task board and kickoff docs | File 41, README Sprint 2 section |
| 2026-04-28 | Created Sprint 1 foundation tag | `v0.1.0-foundation` |
| 2026-04-28 | Created Sprint 2 base branch | `sprint/2-order-fulfillment-core` |
| 2026-04-28 | Delivered Sales Order foundation, API contract, application service, UI, and smoke coverage | S2-01-01 through S2-01-06 |
| 2026-04-28 | Delivered stock reservation on confirm, oversell prevention, non-sellable batch guard, cancel/unreserve, audit, and reservation test suite | S2-02-01 through S2-02-07 |
| 2026-04-28 | Delivered pick task model, generation, API actions, scan-first UI, exception rules, and picking tests | S2-03-01 through S2-03-06 |
| 2026-04-28 | Delivered pack task model, generation, API actions, station UI, exception handling, and packing tests | S2-04-01 through S2-04-06 |
| 2026-04-28 | Delivered carrier master hardening, manifest model/actions/UI, handover zone/bin model, and manifest tests | S2-05-01 through S2-05-06 |
| 2026-04-28 | Delivered handover scan verification, scan event log, missing order exception, confirm handover, handover scan UI, and negative tests | S2-06-01 through S2-06-06 |
| 2026-04-28 | Delivered Warehouse Daily Board fulfillment metrics, UI integration, drill-down links, and consistency coverage | S2-07-01 through S2-07-04 |
| 2026-04-28 | Delivered E2E happy path, permission/audit regression, Sprint 2 test data, demo script, and this release note | S2-08-01 through S2-08-05 |

---

## 6. Release Summary

Sprint 2 delivers the first end-to-end Order Fulfillment Core for the cosmetics ERP prototype:

```text
Sales Order
-> Reserve Stock
-> Pick
-> Pack
-> Carrier Manifest
-> Scan Handover DVVC
-> Warehouse Daily Board
```

The sprint is demo-ready for:

- creating a sales order and confirming it into reserved stock;
- preventing oversell and blocking QC/unsellable batch reservation;
- generating and executing pick tasks;
- generating and executing pack tasks;
- creating and managing carrier manifests;
- scanning handover codes and recording scan events;
- blocking handover when manifest lines are missing;
- confirming handover only after valid complete scans;
- showing fulfillment counts and drill-down links on Warehouse Daily Board;
- proving RBAC and audit coverage for sensitive fulfillment actions.

## 7. Delivered Features

| Area | Delivered capability | Evidence |
|---|---|---|
| Sales Order | Sales order state model, API, UI, create/update/confirm/cancel path | S2-01 tasks |
| Reservation | Stock reservation service, decimal quantity handling, batch allocation, oversell guard | S2-02 tasks |
| QC/Batch guardrail | Reservation blocks non-sellable QC/batch state | S2-02-04 |
| Picking | Pick task lifecycle, line scan/confirm, exceptions, UI | S2-03 tasks |
| Packing | Pack task lifecycle, station UI, exceptions, tests | S2-04 tasks |
| Carrier Manifest | Carrier master, manifest model, add/remove/ready/cancel actions, UI | S2-05 tasks |
| Handover Scan | Scan verification, event log, duplicate/wrong/unknown scan handling | S2-06 tasks |
| Missing Exception | Missing order reporting and handover block while manifest is incomplete | S2-06-03, S2-06-04 |
| Daily Board | Fulfillment metrics API/UI and KPI drill-downs | S2-07 tasks |
| QA Regression | Happy path, permission/audit regression, seed data, demo script | S2-08 tasks |

## 8. Bugs / Risks Closed

| Risk | Sprint 2 closure |
|---|---|
| Overselling available stock | Confirm/reserve flow blocks insufficient stock and reports available quantity. |
| Selling QC hold/fail or blocked batches | Reservation checks sellable batch/QC state before allocation. |
| Cancelled reserved order leaving stock locked | Cancel/unreserve releases active reservations with audit. |
| Missing package handed over to carrier | Handover confirmation is blocked while manifest has missing lines. |
| Duplicate or wrong-carrier scan silently accepted | Scan result codes distinguish duplicate, manifest mismatch, unknown, invalid state, and matched. |
| Warehouse/Sales role overreach | Permission regression covers Sales Ops, Warehouse Staff/Lead, and ERP Admin boundaries. |
| Sensitive state changes without evidence | Audit regression covers confirm/reserve/pick/pack/scan/handover paths. |
| Daily Board count drift | Board consistency test compares metrics against sales order and manifest source state. |

## 9. Known Issues / Limitations

| Issue | Current impact | Mitigation / next action |
|---|---|---|
| GitHub Actions jobs fail before running steps with empty `steps: []` logs | Cloud CI cannot be used as merge evidence right now. | Continue local verification for each PR; fix GitHub Actions runner/billing/infrastructure before production release gating. |
| Some Sprint 2 demo stores are prototype/in-memory state | Demo data resets on service restart and is not yet production persistence for every fulfillment object. | Move fulfillment runtime state to PostgreSQL-backed stores in a follow-up hardening sprint. |
| Handover happy path and missing-exception demo use the same seeded manifest | Running both branches in one live state mutates the same manifest. | Follow S2-08-04 script: use fresh/reset state for each branch. |
| Real carrier integration is not wired | GHN/VTP/NJV are prototype carrier records; no real carrier API booking or webhook yet. | Implement carrier adapter and webhook reconciliation in a later integration sprint. |
| Windows Next build emits `@next/swc-win32-x64-msvc` DLL warnings | Build still succeeds; warning is local environment noise. | Track separately if it appears in non-Windows CI or blocks a build. |
| Returns/subcontract flows remain outside Sprint 2 fulfillment close-out | Existing skeletons are not fully tied to order fulfillment release flow. | Prioritize returns receiving and subcontract handoff in the next sprint plan. |

## 10. Verification Evidence

Local verification used because cloud CI is currently blocked before job steps run:

```text
go test ./cmd/api -run TestSprint2PrototypeTestDataSeed -count=1
go test ./cmd/api -run TestOrderToHandoverHappyPathSmoke -count=1
go test ./cmd/api -run "TestOrderFulfillment(Permission|Audit)RegressionSmoke" -count=1
go test ./...
go vet ./...
pnpm --filter web typecheck
pnpm --filter web test
pnpm smoke:test
pnpm --filter web build
git diff --check
```

Notes:

- Web build passed with the known Windows Next SWC DLL warning.
- GitHub Actions failures observed on Sprint 2 PRs had empty job step lists, so they did not execute the repo checks.

## 11. Operations / Support Notes

Follow file 29 support principles during demo/UAT:

- Every bug, data issue, permission issue, and support request must have a ticket.
- Do not correct stock, batch, sales order, manifest, scan, or audit state directly in the database.
- Classify incidents by workflow area: sales order, reservation, picking, packing, handover, daily board, audit/RBAC.
- Missing handover orders are operationally high risk because they affect carrier handoff and customer shipment evidence.
- Any manual data correction must preserve audit evidence and be approved as a data correction request.

## 12. Next Sprint Recommendations

Recommended order:

1. Fix GitHub Actions infrastructure so cloud CI becomes reliable release evidence again.
2. Persist Sprint 2 fulfillment runtime state to PostgreSQL-backed stores where prototype stores remain.
3. Harden order fulfillment concurrency around reservation, pick, pack, and handover actions.
4. Add real carrier adapter boundaries for booking, label, handover confirmation, and webhook reconciliation.
5. Extend returns flow into order/customer/shipment status recovery.
6. Extend subcontract manufacturing flow into finished goods receiving, QC, and inventory availability.
7. Add UAT scripts for warehouse staff and sales ops using the S2-08-04 demo script as the baseline.
8. Add operational dashboards for daily handover exceptions, missing scans, stock reservation failures, and audit exceptions.
