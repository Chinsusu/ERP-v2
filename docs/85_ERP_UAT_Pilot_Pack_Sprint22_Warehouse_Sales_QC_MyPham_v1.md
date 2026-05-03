# 85_ERP_UAT_Pilot_Pack_Sprint22_Warehouse_Sales_QC_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC
Document role: UAT execution pack, role guide, scenario index, and evidence rules
Version: v1.0
Date: 2026-05-03
Status: Prepared for controlled UAT execution; business UAT not yet executed

---

## 1. Purpose

This pack turns Sprint 22 from a planning board into a controlled UAT execution kit.

It is intentionally not a feature plan. It prepares business users, data, scripts, evidence capture, issue triage, and Go/No-Go reporting for Warehouse, Sales, and QC pilot validation.

UAT must prove this statement:

```text
Real users can complete core Warehouse + Sales + QC flows in Vietnamese UI with backend auth/session and production-like runtime guardrails.
```

---

## 2. Execution Boundary

Sprint 22 UAT is allowed to capture issues, clarify copy, and propose small non-invasive fixes.

Do not treat every UAT request as code scope.

```text
Bug: system contradicts approved docs, data rules, security rules, or expected flow.
Change request: business asks for new behavior, new report, new approval rule, or wider feature.
Training/data issue: user confusion or missing seed/master data, without confirmed product defect.
```

---

## 3. Entry Gate

UAT can start only after these are marked ready in `docs/uat/sprint22/templates/scenario_results.csv` or the issue board:

```text
main synced and clean
Sprint 21 changelog completed
Sprint 21 tag-hold reason documented
required-ci green
dev/UAT environment health returns 200
backend auth/session smoke passed
web login/logout/refresh smoke passed
Vietnamese UI baseline smoke passed
UAT users and roles created
UAT seed data approved
UAT session schedule agreed
evidence folder and templates ready
```

---

## 4. UAT Roles

Use `docs/uat/sprint22/templates/uat_users_roles.csv` as the working register.

Minimum role coverage:

| Role | UAT responsibility | Required access |
| --- | --- | --- |
| UAT Lead | Runs session and controls scope | All scripts and issue board |
| Business Owner | Go/No-Go decision | Dashboard, issue summary, sign-off |
| Warehouse User | Daily board, pick/pack, handover, returns, stock count | Warehouse, shipping, returns |
| Sales User | Sales order, reservation, order visibility | Sales, stock availability |
| QC User | Receiving QC, batch/QC status, quarantine | QC, receiving, inventory |
| ERP Admin | User/role setup and seed support | Settings, master data, RBAC |
| QA/Test Lead | Evidence, screenshots, triage | UAT templates, logs, screenshots |
| Dev Support | Fixes confirmed defects | Repo, logs, deployment |

---

## 5. Seed Data Checklist

Use `docs/uat/sprint22/templates/seed_data_plan.csv` as the controlled seed register.

Minimum data set:

```text
Warehouses: WH-MAIN, WH-RETURN, WH-QA-HOLD, WH-DAMAGED
Locations: A1, A2, PK-01, HO-01, RT-01
Finished goods: 5 active cosmetics SKUs
Materials: 3 raw materials
Packaging: 3 packaging/accessory SKUs
Batches: QC_PASS, QC_HOLD, QC_FAIL, near-expiry
Customers: retail and B2B/dealer samples
Suppliers: raw/packaging suppliers
Carriers: GHN, GHTK, Internal/Manual
Orders: normal, reservation, insufficient stock, QC blocked, ready for handover
Returns: reusable, non-reusable, QA-hold
Purchase/Receiving/QC: full pass, partial receive, QC fail/return-to-supplier
```

Rules:

```text
money: decimal string / VND
quantity: decimal string / base UOM
rate: decimal string
locale: vi-VN
currency: VND
timezone: Asia/Ho_Chi_Minh
```

---

## 6. Required Scenario Scripts

Record each run in `docs/uat/sprint22/templates/scenario_results.csv`.

### S22-UAT-001 - Auth / RBAC / Vietnamese UI

Objective:

```text
Verify backend-backed login/logout, invalid login copy, role-specific menu access, and Vietnamese UI baseline.
```

Steps:

```text
1. Open login page.
2. Login as warehouse user.
3. Confirm dashboard/menu labels are Vietnamese.
4. Confirm restricted admin/finance controls are hidden or blocked.
5. Logout.
6. Confirm user returns to login page.
7. Try invalid login.
8. Confirm Vietnamese error message is clear.
9. Login as sales user.
10. Confirm role-specific menu access.
```

Pass criteria:

```text
Valid login succeeds.
Invalid login shows Vietnamese error.
Logout invalidates session.
Menus follow RBAC.
No mock session behavior appears in production-like mode.
```

### S22-UAT-002 - Warehouse Daily Board

Objective:

```text
Verify the daily warehouse board reflects real work queues and users understand the Vietnamese operational copy.
```

Steps:

```text
1. Open Bảng công việc kho.
2. Review receiving, picking, packing, handover, returns, adjustments, and closing counts.
3. Drill down into one pending task.
4. Return to board.
5. Complete or update one task if the UAT data supports it.
6. Confirm count or status changes are understandable.
```

Pass criteria:

```text
Board uses operational data.
Drill-down goes to the expected screen.
Vietnamese labels are understandable to warehouse users.
```

### S22-UAT-003 - Sales Order -> Reserve -> Pick -> Pack -> Handover

Objective:

```text
Verify available stock, QC status, pick/pack transitions, carrier manifest, and handover scan behavior.
```

Steps:

```text
1. Create or select sales order.
2. Confirm order.
3. Reserve stock.
4. Attempt insufficient-stock reservation case.
5. Attempt QC_HOLD/QC_FAIL batch reservation case.
6. Generate or open pick task.
7. Confirm picking.
8. Pack order.
9. Add packed order to carrier manifest.
10. Scan order or tracking code for handover.
11. Test missing-order exception.
12. Confirm successful handover only after expected scans pass.
```

Pass criteria:

```text
Only available stock can be reserved.
QC_HOLD/QC_FAIL stock cannot be reserved.
Pick/pack/handover state transitions are correct.
Manifest scan rejects invalid or missing items.
Successful handover leaves evidence and audit trail.
```

### S22-UAT-004 - Return Receiving -> Inspection -> Disposition

Objective:

```text
Verify return items do not become sellable before inspection and disposition.
```

Steps:

```text
1. Scan returned order or tracking code.
2. Create return receiving record.
3. Inspect item condition.
4. Select reusable disposition.
5. Confirm reusable path follows approved stock movement rules.
6. Repeat with non-reusable disposition.
7. Confirm non-reusable item does not enter available stock.
8. Repeat with QA-hold disposition.
9. Confirm item goes to quarantine or hold.
```

Pass criteria:

```text
Hàng hoàn chưa kiểm tra không vào tồn khả dụng.
Còn sử dụng, không sử dụng, and cần QA kiểm tra paths are separated.
Movement and audit evidence are recorded.
```

### S22-UAT-005 - Stock Count -> Adjustment -> Shift Closing

Objective:

```text
Verify stock variance handling, approval, reconciliation, and shift close blockers.
```

Steps:

```text
1. Open stock count session.
2. Count SKU/batch/location.
3. Create variance.
4. Submit adjustment request.
5. Approve adjustment if role supports it.
6. Post adjustment movement.
7. Run end-of-day reconciliation.
8. Attempt close shift with unresolved issue.
9. Resolve or document the blocker.
10. Close shift when criteria pass.
```

Pass criteria:

```text
No direct stock balance edit.
Variance requires reason and approval.
Shift closing blocks unresolved operational issues.
Closing evidence is recorded.
```

### S22-UAT-006 - Purchase -> Receiving -> Inbound QC

Objective:

```text
Verify goods receiving, inbound QC, pass/fail/hold behavior, and quarantine rules.
```

Steps:

```text
1. Create or select PO.
2. Receive goods against PO.
3. Enter delivery note, batch, expiry, quantity, and package condition.
4. Run inbound QC.
5. QC PASS case: stock becomes available through approved movement.
6. QC FAIL case: stock does not become available and rejection/return-to-supplier is recorded.
7. QC HOLD/PARTIAL case: stock goes to quarantine or pending.
```

Pass criteria:

```text
Tiếp nhận hàng không đồng nghĩa hàng đã khả dụng.
Chỉ hàng đạt QC mới vào tồn khả dụng.
QC fail/hold/partial are represented clearly in UI and stock behavior.
```

---

## 7. Evidence Rules

Use `docs/uat/sprint22/evidence/` as the evidence staging structure.

Do not commit raw screenshots, logs, or exports that contain real customer data, passwords, access tokens, private addresses, phone numbers, or commercial secrets.

Allowed in git:

```text
sanitized screenshots
sanitized logs
session notes without PII
template CSV files
redacted issue exports
Go/No-Go report without secrets
```

Evidence naming convention:

```text
S22-UAT-001_login_warehouse_pass_YYYYMMDD-HHMM.png
S22-UAT-003_manifest_missing_order_fail_YYYYMMDD-HHMM.png
S22-UAT-006_qc_fail_quarantine_pass_YYYYMMDD-HHMM.png
```

---

## 8. Issue Triage

Use `docs/uat/sprint22/templates/issue_triage_board.csv`.

Severity rules:

| Severity | Meaning | Required action |
| --- | --- | --- |
| P0 | Blocks UAT or risks critical data integrity | Stop session, assign owner immediately |
| P1 | Blocks core scenario for a role/module | Fix before Go or Conditional Go |
| P2 | Workaround exists but flow is painful or confusing | Sprint 23 backlog candidate |
| P3 | Cosmetic or minor improvement | Track for later |
| CR | New request, not a confirmed defect | Change request review |

---

## 9. Go / Conditional Go / No-Go

Use `docs/uat/sprint22/templates/go_no_go_report.md`.

Go:

```text
No open P0.
No open P1 in core UAT flows.
Users can complete all required scenarios with acceptable guidance.
Vietnamese UI is understandable enough for pilot.
RBAC/auth works.
Stock/QC/returns/receiving behaviors are correct.
Evidence pack is complete.
```

Conditional Go:

```text
No open P0.
Some P1/P2 items remain but have business-accepted workaround.
Fix plan and owner are assigned.
Business Owner accepts controlled pilot limits.
```

No-Go:

```text
Any P0 remains.
Any critical stock/QC/auth issue remains.
Users cannot complete sales/warehouse/QC flows.
UAT environment is unstable.
Evidence is incomplete or not trustworthy.
```

---

## 10. Sprint 22 Completion Rule

Sprint 22 is complete only when:

```text
UAT readiness is confirmed.
Users and permissions are verified.
Seed data is loaded and documented.
Required scripts are executed or explicitly waived.
All findings are logged and triaged.
P0/P1 blockers have owners.
Evidence pack is complete.
Go/Conditional Go/No-Go decision is recorded.
Sprint 22 changelog is updated from draft to completed.
```

Do not mark this pack as completed until business UAT evidence exists.
