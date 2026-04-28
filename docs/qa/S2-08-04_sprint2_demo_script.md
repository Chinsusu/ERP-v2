# S2-08-04 Sprint 2 Demo Script

Primary Ref: docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md
Task Ref: docs/41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md#79-test--regression--demo

## Purpose

This script is the Sprint 2 review runbook. It proves the Order Fulfillment Core can move a cosmetics sales order through:

- sales order creation and confirmation,
- stock reservation with batch/QC guardrails,
- picking,
- packing,
- carrier manifest handover scan,
- missing order exception handling,
- Warehouse Daily Board fulfillment metrics,
- audit and regression evidence.

Expected duration: 35 to 45 minutes.

## Audience

- Sponsor / PO
- BA
- QA Lead
- Sales Ops super user
- Warehouse Lead / Warehouse Staff
- Tech Lead / DevOps

## Prerequisites

- Dev or Staging stack is deployed and points to the Sprint 2 build.
- Sprint 2 prototype seed data is loaded.
- Local or pipeline evidence is ready for:
  - `go test ./cmd/api -run TestOrderToHandoverHappyPathSmoke -count=1`
  - `go test ./cmd/api -run "TestOrderFulfillment(Permission|Audit)RegressionSmoke|TestSprint2PrototypeTestDataSeed" -count=1`
  - `pnpm smoke:test`
- Demo browser starts at `/login`.
- Use a fresh/reset demo state for each pass below. The happy path and missing-exception branch both use `manifest-hcm-ghn-morning`, so do not run both against the same in-memory state without reset.

## Demo Data

| Area | Value |
|---|---|
| ERP Admin login | `admin@example.local` / `local-only-mock-password` |
| Sales user | `sales_user@example.local` |
| Warehouse user | `warehouse_user@example.local` |
| Warehouse | `wh-hcm-fg` / `WH-HCM-FG` |
| Handover warehouse | `wh-hcm` / `HCM` |
| Customer | `CUS-DL-MINHANH` |
| Marketplace customers | `CUS-MP-SHOPEE`, `CUS-MP-TIKTOK` |
| SKUs | `SERUM-30ML`, `CREAM-50G`, `TONER-100ML` |
| Sellable batches | `LOT-2604A`, `LOT-2603B` |
| QC/blocked batch evidence | `LOT-2604C` |
| Seeded order count | 20 prototype sales orders |
| Active carriers | `GHN`, `NJV`, `VTP` |
| Demo manifest | `manifest-hcm-ghn-morning` |
| Missing order | `SO-260426-003` |
| Successful scan code | `GHN260426003` |
| Duplicate scan code | `GHN260426001` |
| Wrong manifest code | `VTP260426011` |
| Unknown scan code | `UNKNOWN-CODE` |

## Run Order

### 1. Open With Sprint 2 Gate

Narrative:

Sprint 2 is not a generic UI demo. It proves that order fulfillment state changes are controlled and auditable from Sales Order through carrier handover.

Show:

- `docs/41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md`
- `docs/42_ERP_Sprint2_Changelog_Order_Fulfillment_Core_MyPham_v1.md`
- Latest PR list for Sprint 2 tasks.

Expected result:

- Stakeholders understand the Sprint 2 goal: Sales Order -> Reserve Stock -> Pick -> Pack -> Carrier Manifest -> Scan Handover -> Warehouse Daily Board.

Evidence to capture:

- Screenshot of the Sprint 2 task board and changelog.

### 2. Login and Role Boundary

Steps:

1. Open `/login`.
2. Sign in as `admin@example.local`.
3. Open `/dashboard`.
4. Point out available modules: Sales, Shipping, Warehouse, Inventory, Master Data, Audit Log.
5. Mention regression evidence for Sales Ops, Warehouse Staff, Warehouse Lead, and ERP Admin permissions.

Expected result:

- ERP Admin can access the full Sprint 2 demo path.
- Sales and warehouse sensitive actions are covered by permission regression tests.

Evidence to capture:

- Dashboard screenshot.
- Test evidence from `TestOrderFulfillmentPermissionRegressionSmoke`.

### 3. Confirm Seed Readiness

Steps:

1. Open `/master-data`.
2. Show customers including `CUS-MP-TIKTOK`.
3. Show products/SKUs: `SERUM-30ML`, `CREAM-50G`, `TONER-100ML`.
4. Open `/inventory` and show stock for `SERUM-30ML`, `CREAM-50G`, and `TONER-100ML`.
5. Open `/sales` and show seeded prototype orders.

Expected result:

- Demo data covers customers, SKUs, batches, stock, 20 sales orders, 3 active carriers, and 2 HCM manifests for the handover date.

Evidence to capture:

- Screenshot of master data or inventory.
- Test evidence from `TestSprint2PrototypeTestDataSeed`.

### 4. Sales Order to Reservation

Steps:

1. Open `/sales`.
2. Create a new draft order:
   - customer `CUS-DL-MINHANH`,
   - warehouse `wh-hcm-fg`,
   - date `2026-04-28`,
   - line `SERUM-30ML`, quantity `2`, unit price `125000`.
3. Open the created order.
4. Confirm the order.
5. Show that the order moves to reserved and the line has a reserved batch.

Expected result:

- Draft creation succeeds.
- Confirmation reserves stock.
- Reservation uses string decimal quantity and batch data, not floating point math.

Evidence to capture:

- Sales order detail screenshot after confirm.
- Mention automated coverage: `TestSalesOrderAPISmokePack`.

### 5. Picking

Steps:

1. Open `/shipping`.
2. Select the Picking tab.
3. Select a pick task for an HCM order.
4. Start picking if the task is draft.
5. Scan or confirm the pick line.
6. Complete the pick task.

Expected result:

- Pick task moves through draft/in-progress/completed states.
- Pick line records SKU, batch, bin, picked quantity, and actor.
- The related sales order can move into picked/packing state.

Evidence to capture:

- Picking tab screenshot.
- Mention regression evidence from the order-to-handover smoke path.

### 6. Packing

Steps:

1. Stay in `/shipping`.
2. Select the Packing tab.
3. Select a pack task for the picked order.
4. Start packing if needed.
5. Scan or confirm the pack line.
6. Confirm packing.

Expected result:

- Pack task reaches packed/completed state.
- Packed order is eligible for carrier manifest handover.
- Audit evidence exists for pack actions.

Evidence to capture:

- Packing tab screenshot.
- Test evidence from `TestOrderFulfillmentAuditRegressionSmoke`.

### 7. Handover Happy Path

Use a fresh/reset demo state for this pass.

Steps:

1. Open `/shipping`.
2. Select the Carrier handover tab.
3. Select manifest `manifest-hcm-ghn-morning`.
4. Show starting counts:
   - expected: 3,
   - scanned: 2,
   - missing: 1.
5. Scan `GHN260426003`.
6. Confirm handover.

Expected result:

- Scan result is `MATCHED`.
- Missing count becomes zero.
- Confirm handover succeeds.
- Manifest moves to handed over.
- Sales orders in the manifest move to handed over.

Evidence to capture:

- Handover screenshot after scan.
- Handover confirmation screenshot.
- Test evidence from `TestOrderToHandoverHappyPathSmoke`.

### 8. Missing Order Exception Branch

Use a fresh/reset demo state for this pass. Do not scan `GHN260426003` before this branch.

Steps:

1. Open `/shipping`.
2. Select Carrier handover.
3. Select manifest `manifest-hcm-ghn-morning`.
4. Show missing line `SO-260426-003`.
5. Explain that confirm handover is blocked while missing count is non-zero.
6. Use the missing line action to report the missing order exception.

Expected result:

- Manifest status moves to exception.
- Missing line remains visible as business evidence.
- Audit log records the missing exception action.
- Warehouse users have an explicit exception flow instead of silently handing over a short manifest.

Evidence to capture:

- Missing line screenshot.
- Exception status screenshot.
- Audit evidence from `TestReportCarrierManifestMissingOrdersHandlerMarksException`.

### 9. Handover Negative Scan Cases

Steps:

Run these as quick negative cases on a fresh/reset manifest state:

| Case | Input | Expected result |
|---|---|---|
| Duplicate scan | `GHN260426001` | Duplicate scan warning |
| Wrong manifest | `VTP260426011` | Manifest mismatch warning |
| Unknown code | `UNKNOWN-CODE` | Not found warning |

Expected result:

- Each failed scan produces a clear non-success result.
- Scan event evidence includes actor, device/source, manifest, order/tracking number when available.

Evidence to capture:

- One screenshot of recent scan issues.
- Test evidence from handover negative tests and scan event coverage.

### 10. Warehouse Daily Board

Steps:

1. Open `/warehouse`.
2. Set date `2026-04-26`, warehouse `wh-hcm`, carrier `GHN`.
3. Show fulfillment cards for waiting handover and missing.
4. Click the Missing KPI drill-down.
5. Return to the board and show that fulfillment metrics match sales/manifest state.

Expected result:

- Daily Board shows real fulfillment counts.
- KPI drill-down opens the relevant list.
- Board metrics are backed by sales order and manifest source data.

Evidence to capture:

- Daily Board screenshot.
- Test evidence from `TestWarehouseDailyBoardFulfillmentMetricsMatchSalesAndManifestState`.

### 11. Audit and Verification Evidence

Steps:

1. Open `/audit-log`.
2. Show audit rows for:
   - `sales.order.reserved`,
   - `inventory.stock_reservation.reserved`,
   - `shipping.pick_task.*`,
   - `shipping.pack_task.*`,
   - `shipping.manifest.scan_recorded`,
   - `shipping.manifest.handed_over` or missing exception.
3. Show local verification commands/results.

Expected result:

- Sensitive state changes have audit evidence.
- Demo does not rely only on screenshots; automated checks prove the core path.

Evidence to capture:

- Audit Log screenshot.
- Terminal or PR evidence for local checks.

## Demo Close

Close with these decisions:

- Sprint 2 Order Fulfillment Core is demo-ready if the happy path, missing exception branch, Daily Board, and audit evidence all pass.
- Any failed step becomes a P0/P1 defect before release note sign-off.
- Known infrastructure caveat: GitHub Actions can fail before running steps in the current repo setup; local verification remains the merge evidence until CI infrastructure is fixed.

## Presenter Checklist

- [ ] Dev/Staging stack is up.
- [ ] Seed data reset is available.
- [ ] Login works.
- [ ] Sales order create/confirm path is ready.
- [ ] Picking tab opens.
- [ ] Packing tab opens.
- [ ] Carrier handover manifest `manifest-hcm-ghn-morning` is in seed state.
- [ ] Missing exception branch is run on fresh/reset state.
- [ ] Daily Board filters are ready.
- [ ] Audit Log page opens.
- [ ] Automated test evidence is ready.

## Traceability

| Acceptance item | Script coverage |
|---|---|
| Clear step-by-step demo | Run Order sections 1-11 |
| Sales order to handover lifecycle | Sections 4-7 |
| Missing order exception | Section 8 |
| Negative scan cases | Section 9 |
| Daily Board integration | Section 10 |
| Audit evidence | Section 11 |
| Seeded Sprint 2 test data | Section 3 |
