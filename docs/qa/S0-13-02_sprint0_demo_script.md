# S0-13-02 Sprint 0 Demo Script

Primary Ref: docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md
Task Ref: docs/37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md#s0-13-02-sprint-0-demo-script

## Purpose

This script is the Sprint 0 review runbook. It proves the ERP foundation is ready for Sprint 1 work by showing:

- the app shell and RBAC entry point,
- cosmetics sample data,
- available stock and stock movement evidence,
- shipping scan handover success and failure cases,
- audit log evidence for sensitive actions,
- smoke test evidence from the automated pack.

Expected duration: 25 to 35 minutes.

## Audience

- Sponsor / PO
- BA
- QA Lead
- Warehouse super user
- Tech Lead / DevOps

## Prerequisites

- Dev or Staging stack is deployed.
- `make smoke-test` or `pnpm smoke:test` has passed.
- Seed data from `tools/seed/smoke/sprint0_smoke_seed.json` is available.
- Demo browser starts at the ERP login page.
- API base URL is known for optional API evidence.

## Demo Data

| Area | Value |
|---|---|
| Login user | `admin@example.local` |
| Product | Hydrating Serum 30ml |
| SKU | `SERUM-30ML` |
| Batch | `SMOKE-SERUM-260426` |
| Warehouse | HCM Main Warehouse |
| Stock movement | `mov-smoke-adjust` |
| Manifest | `manifest-hcm-ghn-morning` |
| Success scan code | `GHN260426003` |
| Duplicate scan code | `GHN260426001` |
| Wrong manifest code | `VTP260426011` |
| Unknown scan code | `UNKNOWN-CODE` |

## Run Order

### 1. Open with Sprint 0 Gate

Narrative:

Sprint 0 is not a full ERP demo. It proves the foundation: repo, CI, deploy skeleton, RBAC, audit, stock ledger prototype, warehouse board, scan handover, returns, subcontract, smoke tests, and demo evidence.

Show:

- `docs/37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md`
- S0 tasks marked Done except this demo task during review.
- Latest smoke pack evidence in `docs/qa/S0-13-01_smoke_test_pack.md`.

Expected result:

- Stakeholders understand this is a foundation acceptance demo, not a full production workflow.

Evidence to capture:

- Screenshot of task board status.

### 2. Login and RBAC Shell

Steps:

1. Open `/login`.
2. Sign in as `admin@example.local`.
3. Land on `/dashboard`.
4. Point out left navigation groups: Overview, Operations, Data, Control.
5. Open Master Data and Shipping from the menu.

Expected result:

- ERP Admin can see Master Data, Inventory, Shipping, Audit Log, and Settings.
- Warehouse staff cannot access Master Data according to the smoke test negative check.

Evidence to capture:

- Screenshot of dashboard navigation.
- Mention automated coverage: `apps/web/src/modules/smoke/sprint0Smoke.test.ts`.

### 3. Healthcheck and Deploy Readiness

Steps:

1. Open or call `/healthz`.
2. Open or call `/readyz`.
3. Open or call `/api/v1/health`.
4. If using a deployed stack, show reverse proxy health from `/healthz`.

Expected result:

- API health returns `ok`.
- API readiness returns `ready`.
- Reverse proxy health returns an OK JSON payload.

Evidence to capture:

- Screenshot or terminal output of health response.

### 4. Cosmetics Master Data

Steps:

1. Open `/master-data`.
2. Open `/sku-batch`.
3. Open `/suppliers`.
4. Open `/customers`.
5. Use the smoke seed story:
   - SKU `SERUM-30ML`
   - Batch `SMOKE-SERUM-260426`
   - Supplier `Smoke Packaging Supplier`
   - Factory `Smoke GMP Factory`
   - Customer `Smoke Retail Customer`

Expected result:

- Master data shells are reachable.
- Demo data is cosmetics-specific and traceable to the smoke seed file.

Evidence to capture:

- Screenshot of Master Data and SKU / Batch pages.

### 5. Available Stock

Steps:

1. Open `/inventory`.
2. Filter or point to HCM Main Warehouse and serum stock rows.
3. Explain the distinction between physical, reserved, hold, and available stock.

Expected result:

- Available stock page shows stock availability structure.
- The team sees that available stock is a derived operational view, not a direct editable balance.

Evidence to capture:

- Screenshot of Available Stock page.

### 6. Stock Movement Audit Path

Steps:

1. Use the sample API smoke data for movement `mov-smoke-adjust`.
2. Execute or show the API smoke test step that posts an `ADJUST` movement for `SERUM-30ML`.
3. Open `/audit-log`.
4. Show audit action `inventory.stock_movement.adjusted`.

Expected result:

- Stock movement returns recorded status.
- Audit log has one row for `mov-smoke-adjust`.
- The demo reinforces the rule: no direct stock balance edit outside stock movement.

Evidence to capture:

- API response or automated test output.
- Audit Log screenshot filtered to `inventory.stock_movement.adjusted`.

### 7. Shipping Handover - Successful Scan

Steps:

1. Open `/shipping`.
2. Select or point to manifest `manifest-hcm-ghn-morning`.
3. Scan or enter `GHN260426003`.
4. Show the result code and manifest counts.

Expected result:

- Result is `MATCHED`.
- Missing count reaches zero for the prototype manifest.
- Audit log is written for `shipping.manifest.scan_recorded`.

Evidence to capture:

- Screenshot of scan result.
- Audit Log screenshot for `shipping.manifest.scan_recorded`.

### 8. Shipping Handover - Failure Cases

Run these as quick negative cases:

| Case | Input | Expected result |
|---|---|---|
| Duplicate scan | `GHN260426001` | Duplicate scan warning |
| Wrong manifest | `VTP260426011` | Manifest mismatch warning |
| Unknown code | `UNKNOWN-CODE` | Not found warning |

Expected result:

- Each failed scan produces a clear, non-success result.
- The presenter explains that warehouse users need visible exceptions, not silent failures.

Evidence to capture:

- One screenshot or test output showing the negative scan cases.

### 9. Returns and Subcontract Context

Steps:

1. Open `/returns`.
2. Point out return receiving and inspection flow skeleton.
3. Open `/subcontract`.
4. Point out external factory order timeline, sample approval, and factory claim SLA.

Expected result:

- Stakeholders see that Sprint 0 also placed returns and subcontract foundations into the app shell.
- No claim is made that full business processing is complete.

Evidence to capture:

- One screenshot for Returns.
- One screenshot for Subcontract.

### 10. Smoke Test Evidence

Steps:

1. Show the command:

   ```bash
   pnpm smoke:test
   ```

2. Show CI e2e check from the latest PR.
3. Show smoke checklist in `docs/qa/S0-13-01_smoke_test_pack.md`.

Expected result:

- Automated API and frontend smoke tests pass.
- QA has a repeatable daily smoke pack for Dev/Staging.

Evidence to capture:

- CI check screenshot or terminal output.

## Demo Close

Close with these decisions:

- Sprint 0 foundation is ready if all smoke checks pass.
- Sprint 1 should start with Auth/RBAC hardening, Master Data, Inventory Stock Ledger v1, Batch/QC status, and Warehouse receiving basics.
- Any failed smoke or demo step becomes a P0/P1 defect before Sprint 1 feature work proceeds.

## Presenter Checklist

- [ ] Demo environment is up.
- [ ] Smoke test passed.
- [ ] Login works.
- [ ] Available stock page opens.
- [ ] Stock movement audit evidence is ready.
- [ ] Successful scan case is ready.
- [ ] Failed scan cases are ready.
- [ ] Audit Log page opens.
- [ ] Returns and subcontract pages open.
- [ ] Screenshots or CI links are captured.

## Traceability

| Acceptance item | Script coverage |
|---|---|
| Clear steps | Run Order sections 1-10 |
| Cosmetics sample data | Demo Data and Master Data sections |
| Successful and failed scan cases | Sections 7 and 8 |
| Audit log | Sections 6 and 7 |
| Stock movement and available stock | Sections 5 and 6 |
