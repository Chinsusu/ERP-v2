# 84_ERP_Coding_Task_Board_Sprint22_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC
Document role: Coding/task board, UAT preparation plan, pilot execution checklist, and evidence pack structure.
Status: Draft for Sprint 22 kickoff
Previous sprint: Sprint 21 - Auth UI Backend Integration + Production Runtime Smoke

---

## 1. Executive Summary

Sprint 22 is not a feature-expansion sprint. It is the first controlled **UAT pilot sprint** after:

- Vietnamese-first UI localization is completed.
- Web auth UI is wired to real backend auth/session APIs.
- Production fallback/mock auth path is blocked.
- Core operational flows for warehouse, sales fulfillment, returns, purchase receiving, inbound QC, stock ledger, and daily board have enough runtime foundation to be tested by real users.

The purpose of Sprint 22 is to move from:

```text
Developer verified
```

to:

```text
Business user verified under controlled UAT scenarios
```

Sprint 22 will prepare UAT environment, users, seed data, test scripts, issue tracking, observation logs, and Go/No-Go evidence for the first operational pilot.

---

## 2. Sprint Goal

Sprint 22 goal:

```text
Run controlled UAT for Warehouse + Sales + QC using real business flows, Vietnamese UI, backend auth/session, and production-like runtime guardrails.
```

The sprint succeeds when business users can complete the following flows without critical blockers:

```text
1. Login / logout / role-based menu access
2. Warehouse Daily Board review
3. Sales Order -> Reserve Stock -> Pick -> Pack -> Carrier Manifest -> Scan Handover
4. Return Receiving -> Return Inspection -> Disposition -> Stock Movement / Quarantine
5. Stock Count -> Adjustment -> End-of-Day Reconciliation -> Shift Closing
6. Purchase Order -> Goods Receiving -> Inbound QC -> QC Pass/Fail/Hold -> Stock Movement / Quarantine
7. Vietnamese UI labels, operational copy, status chips, errors, and validation messages are understandable to users
```

---

## 3. Why Sprint 22 Comes Now

Sprint 21 closed the biggest production-like runtime caveat: web login moved from mock/staging-only auth surface to backend session integration.

After Sprint 21, the next risk is no longer only technical. The next risk is:

```text
Can warehouse, sales, and QC users actually operate the ERP correctly in Vietnamese UI using realistic data?
```

This sprint is designed to expose:

- confusing UI labels;
- missing Vietnamese operational copy;
- permission gaps;
- flow friction;
- incorrect assumptions in warehouse/sales/QC workflows;
- seed data or master data gaps;
- UAT blockers before a wider pilot.

---

## 4. In Scope

Sprint 22 includes:

```text
- Sprint 21 release checkpoint cleanup
- UAT environment readiness
- UAT users and roles
- UAT seed data
- UAT scenario scripts
- UAT session schedule
- Business user observation log
- Issue triage board
- UAT evidence pack
- Go/No-Go report
- Small copy/label fixes if they are non-invasive
- Small seed data or permission fixes needed to unblock UAT
```

---

## 5. Out of Scope

Sprint 22 does not include:

```text
- New large business modules
- CRM / HRM / KOL / advanced finance
- Real marketplace/carrier integrations
- Production go-live
- Major database redesign
- Major API redesign
- Route localization into Vietnamese
- Backend enum localization
- DB enum/value localization
- Permission code localization
- Audit event localization
```

Important rule:

```text
UAT findings are captured and triaged. Do not silently turn every user request into code during UAT.
```

---

## 6. Branch, Tag, and Release Naming

Required branch if Sprint 22 contains docs/scripts/minor fixes:

```bash
git checkout main
git pull origin main
git checkout -b codex/s22-uat-pilot-pack
```

Sprint 21 tag status before Sprint 22 starts:

```text
Sprint 21 is merged and documented.
The v0.21.0-auth-ui-backend-integration-runtime-smoke tag remains on hold until target staging/pilot smoke evidence exists.
```

Recommended Sprint 22 checkpoint tag after UAT evidence is complete:

```text
v0.22.0-uat-pilot-pack-warehouse-sales-qc
```

If Sprint 22 is only UAT documentation and no production-like release is intended, tag may be held, but the changelog must explicitly state:

```text
Sprint 22 is UAT evidence only; no runtime release tag created.
```

---

## 7. Preconditions Before UAT

Before business users touch the system, verify:

```text
- main is clean and synced with origin/main
- Sprint 21 changelog is updated from pending to completed
- Sprint 21 tag is pushed or tag-hold reason is documented
- required-ci is green
- required-migration is green
- dev/staging deployment is healthy
- backend auth/session works
- web auth UI calls backend auth/session
- logout invalidates backend session
- refresh rotation works
- production mock auth path is blocked
- Vietnamese UI is enabled by default
- root language is vi
- Ant Design locale is vi_VN
- basic UAT users exist
- UAT seed data exists
```

---

## 8. UAT Roles

| Role | Purpose | Example access |
|---|---|---|
| UAT Lead | Runs session, collects evidence, controls scope | All UAT scripts and issue board |
| Business Owner | Go/No-Go decision | Dashboard, reports, issue summary |
| Warehouse User | Tests daily board, receiving, pick/pack, handover, returns, stock count, closing | Warehouse, shipping, returns |
| Sales User | Tests sales order and fulfillment visibility | Sales, stock availability, order status |
| QC User | Tests inbound QC, batch/QC status, quarantine | QC, receiving, inventory |
| Finance Observer | Observes COD/finance-lite fields if available | Finance/read-only reports |
| ERP Admin | Creates users, roles, seed data, fixes access | Settings, master data, RBAC |
| QA/Test Lead | Records evidence, triages bugs | UAT scripts, screenshots, logs |
| Dev Support | Fixes confirmed defects | Backend/frontend/devops |

---

## 9. Seed Data Requirements

Minimum seed data for UAT:

```text
Warehouses:
- WH-MAIN: Kho tổng
- WH-RETURN: Khu hàng hoàn
- WH-QA-HOLD: Khu cách ly / QA hold
- WH-DAMAGED: Kho hàng hỏng / Lab

Locations:
- A1, A2: khu lấy hàng
- PK-01: khu đóng hàng
- HO-01: khu bàn giao ĐVVC
- RT-01: khu hàng hoàn

Items/SKUs:
- 5 thành phẩm mỹ phẩm
- 3 nguyên liệu
- 3 bao bì/phụ liệu
- 1 combo/set if supported

Batches:
- 2 batch QC_PASS
- 1 batch QC_HOLD
- 1 batch QC_FAIL
- 1 near-expiry batch if available

Customers:
- 3 retail customers
- 2 B2B/dealer customers if available

Suppliers:
- 2 raw/packaging suppliers

Carriers:
- GHN
- GHTK
- Internal/Manual carrier if available

Orders:
- 5 normal sales orders
- 2 orders requiring reservation
- 1 order with insufficient stock
- 1 order blocked by QC_HOLD batch
- 2 orders ready for manifest/handover

Returns:
- 1 reusable return
- 1 non-reusable return
- 1 QA-hold return

Purchase/Receiving/QC:
- 1 PO full pass
- 1 PO partial receive
- 1 PO QC fail/return-to-supplier
```

All quantities must follow the unit/currency/number standard:

```text
money: decimal string / VND
quantity: decimal string / base UOM
rate: decimal string
locale: vi-VN
currency: VND
timezone: Asia/Ho_Chi_Minh
```

---

## 10. Sprint 22 Backlog

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S22-00-01 | P0 | PM/Release | Sprint 21 release checkpoint | Sprint 21 changelog updated as completed; PR #542 merge commit recorded; v0.21 tag pushed or tag-hold documented | `83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` |
| S22-00-02 | P0 | DevOps | UAT environment readiness | Dev/UAT environment healthy; `/health` returns 200; auth/session APIs work; production fallback blocked | `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` |
| S22-00-03 | P0 | ERP Admin | UAT role/user setup | UAT users created for warehouse, sales, QC, admin, observer; permissions match expected access | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S22-00-04 | P0 | FE/QA | Vietnamese UI baseline smoke | Login, dashboard, menu, common status/error text render in Vietnamese; route remains English | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
| S22-01-01 | P0 | BA/ERP Admin | UAT seed data plan | Seed data list approved: warehouses, locations, SKUs, batches, customers, suppliers, carriers, orders, returns, PO/QC cases | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S22-01-02 | P0 | BE/ERP Admin | Load UAT master data | Seed master data is loaded and visible in UI; no duplicate/conflicting codes | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S22-01-03 | P0 | BE/QA | Load UAT stock/batch data | QC_PASS/QC_HOLD/QC_FAIL batches exist; stock balances and reservations are testable | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S22-01-04 | P1 | QA | UAT evidence folder structure | Evidence folders exist for screenshots, logs, issue export, and session notes | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S22-02-01 | P0 | QA/FE | UAT script: Auth/RBAC/Vietnamese UI | Script covers login, invalid login, logout, menu access by role, Vietnamese error messages | `82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` |
| S22-02-02 | P0 | BA/Warehouse | UAT script: Warehouse Daily Board | Script covers daily board counts, pending tasks, alerts, drill-down, shift context | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S22-02-03 | P0 | BA/Sales | UAT script: Sales Order to Reserve | Script covers create order, confirm, reserve stock, insufficient stock, QC-blocked batch | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S22-02-04 | P0 | BA/Warehouse | UAT script: Pick/Pack/Handover | Script covers pick, pack, carrier manifest, scan handover, missing order exception | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S22-02-05 | P0 | BA/Warehouse/QC | UAT script: Returns/Inspection/Disposition | Script covers return scan, reusable, non-reusable, QA-hold, movement/quarantine | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S22-02-06 | P0 | BA/Warehouse | UAT script: Stock Count/Shift Closing | Script covers count session, variance, adjustment, reconciliation blockers, close shift | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S22-02-07 | P0 | BA/QC | UAT script: Purchase/Receiving/Inbound QC | Script covers PO, goods receiving, QC pass/fail/hold/partial, return-to-supplier | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S22-03-01 | P1 | BA/Trainer | UAT quick training guide | 1-page guide for each role: warehouse, sales, QC, admin; includes Vietnamese terms | `26_ERP_SOP_Training_Manual_Phase1_MyPham_v1.md` |
| S22-03-02 | P1 | PM | UAT session schedule | Session calendar, users, roles, test scripts, data set, timebox, support owner | `29_ERP_Operations_Support_Model_Phase1_MyPham_v1.md` |
| S22-03-03 | P1 | QA/PM | UAT observation log template | Template captures user action, confusion point, screenshot, severity, suggested fix | `29_ERP_Operations_Support_Model_Phase1_MyPham_v1.md` |
| S22-03-04 | P1 | QA/PM | UAT issue triage rules | P0/P1/P2/P3 rules, owner, SLA, bug vs change request distinction | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S22-03-05 | P1 | PM/Business | Change request capture | Non-bug UAT findings are recorded as CR, not silently coded | `30_ERP_Data_Governance_Change_Control_Phase1_MyPham_v1.md` |
| S22-04-01 | P0 | UAT Team | Conduct UAT Session 1 - Auth + Warehouse Daily Board | Users complete login/menu/daily board tasks; issues logged with evidence | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S22-04-02 | P0 | UAT Team | Conduct UAT Session 2 - Sales + Pick/Pack + Handover | End-to-end order fulfillment scenario completed or blockers logged | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S22-04-03 | P0 | UAT Team | Conduct UAT Session 3 - Returns + Closing | Returns, inspection, disposition, stock count, reconciliation, closing tested | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S22-04-04 | P0 | UAT Team | Conduct UAT Session 4 - Purchase + Receiving + QC | PO/receiving/inbound QC tested with pass/fail/hold/partial cases | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S22-05-01 | P0 | QA/PM | UAT issue triage board | All UAT issues triaged as bug/change/training/data; P0/P1 owners assigned | `29_ERP_Operations_Support_Model_Phase1_MyPham_v1.md` |
| S22-05-02 | P1 | QA | UAT evidence pack | Screenshots, logs, test results, environment info, user feedback, issue list collected | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S22-05-03 | P0 | Business Owner/PM | UAT Go/No-Go report | Decision report: Go, Conditional Go, or No-Go; blockers and next actions recorded | `27_ERP_GoLive_Runbook_Hypercare_Phase1_MyPham_v1.md` |
| S22-05-04 | P1 | PM | Sprint 22 changelog draft | Changelog created with scope, evidence, known issues, next sprint recommendations | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |

---

## 11. UAT Scripts - Required Scenario Detail

### 11.1 Auth / RBAC / Vietnamese UI

```text
1. Open login page.
2. Login with warehouse user.
3. Confirm dashboard/menu labels are Vietnamese.
4. Confirm warehouse user does not see restricted admin/finance controls.
5. Logout.
6. Confirm user returns to login page.
7. Try invalid login.
8. Confirm Vietnamese error message is clear.
9. Login as sales user.
10. Confirm role-specific menu access.
```

Expected result:

```text
- Login succeeds for valid users.
- Invalid login shows Vietnamese error.
- Logout invalidates session.
- Menus follow RBAC.
- No mock session behavior appears in production-like mode.
```

---

### 11.2 Warehouse Daily Board

```text
1. Open Bảng công việc kho.
2. Review counts for pending receiving, picking, packing, handover, returns, adjustments, closing.
3. Drill down into one pending task.
4. Return to board.
5. Confirm counts update after action.
```

Expected result:

```text
- Board uses real operational data.
- Vietnamese labels are understandable.
- Drill-down links lead to correct screens.
```

---

### 11.3 Sales Order -> Reserve -> Pick -> Pack -> Handover

```text
1. Create sales order.
2. Confirm order.
3. Reserve stock.
4. Try reserve with insufficient stock case.
5. Try reserve using QC_HOLD/QC_FAIL batch case.
6. Generate pick task.
7. Confirm picking.
8. Pack order.
9. Add packed order to carrier manifest.
10. Scan order/tracking code for handover.
11. Confirm missing-order exception case.
12. Confirm successful handover.
```

Expected result:

```text
- Confirmed order reserves available stock only.
- QC_HOLD/QC_FAIL stock cannot be reserved.
- Pick/pack state transitions are correct.
- Manifest scan rejects invalid/missing items.
- Successful handover is audited.
```

---

### 11.4 Return Receiving -> Inspection -> Disposition

```text
1. Scan returned order/tracking code.
2. Create return receiving record.
3. Inspect item condition.
4. Select reusable disposition.
5. Confirm reusable movement into available/approved stock path.
6. Repeat with non-reusable disposition.
7. Confirm non-reusable item does not enter available stock.
8. Repeat with QA-hold disposition.
9. Confirm item goes to quarantine/HOLD.
```

Expected result:

```text
- Hàng hoàn chưa kiểm không vào tồn khả dụng.
- Còn sử dụng / không sử dụng / cần QA kiểm tra are clearly separated.
- Movement and audit log are created correctly.
```

---

### 11.5 Stock Count -> Adjustment -> Shift Closing

```text
1. Open stock count session.
2. Count SKU/batch/location.
3. Create variance.
4. Submit adjustment request.
5. Approve adjustment.
6. Post adjustment movement.
7. Run end-of-day reconciliation.
8. Attempt close shift with unresolved issue.
9. Resolve issue.
10. Close shift.
```

Expected result:

```text
- No direct stock balance edit.
- Variance requires reason and approval.
- Shift closing blocks unresolved operational issues.
- Closing evidence is recorded.
```

---

### 11.6 Purchase -> Receiving -> Inbound QC

```text
1. Create PO.
2. Receive goods against PO.
3. Enter delivery note, batch, expiry, quantity, package condition.
4. Run inbound QC.
5. QC PASS case: stock becomes available through movement.
6. QC FAIL case: stock does not become available and rejection/return-to-supplier is recorded.
7. QC HOLD/PARTIAL case: stock goes to quarantine/pending.
```

Expected result:

```text
- Tiếp nhận hàng không đồng nghĩa hàng đã khả dụng.
- Chỉ hàng đạt QC mới vào tồn khả dụng.
- QC fail/hold/partial are clearly represented in UI and stock behavior.
```

---

## 12. Issue Severity Rules

| Severity | Definition | Example | Required action |
|---|---|---|---|
| P0 | Blocks UAT or corrupts critical data | Login impossible for all users; stock movement wrong; QC fail becomes available | Stop session, fix immediately |
| P1 | Blocks a core scenario for one role/module | Warehouse cannot close shift; manifest scan cannot confirm | Fix before Go/Conditional Go |
| P2 | Workaround exists but user flow is painful/confusing | Label unclear; table missing useful column; error copy not helpful | Add to Sprint 23 backlog |
| P3 | Cosmetic or minor improvement | Spacing, minor copy polish, optional filter | Track for later |
| CR | Not a bug; new business request | New approval rule, new report, new role | Change request review |

---

## 13. Go / Conditional Go / No-Go Criteria

### Go

```text
- No open P0.
- No open P1 in core UAT flows.
- Users can complete all required scenarios with acceptable guidance.
- Vietnamese UI is understandable enough for pilot.
- RBAC/auth works.
- Stock/QC/returns/receiving behaviors are correct.
- Evidence pack is complete.
```

### Conditional Go

```text
- No open P0.
- Some P1/P2 issues remain but have operational workaround.
- Business Owner accepts controlled pilot limits.
- Fix plan and owner are assigned.
```

### No-Go

```text
- Any P0 remains.
- Any critical stock/QC/auth issue remains.
- Users cannot complete sales/warehouse/QC flows.
- UAT environment is unstable.
- Evidence is incomplete or not trustworthy.
```

---

## 14. Evidence Pack Checklist

UAT evidence pack must include:

```text
- environment URL / commit / tag / branch
- UAT user list and role mapping
- seed data snapshot
- scenario script results
- screenshots for each successful flow
- screenshots/logs for each failure
- issue triage board
- observation notes
- Go/No-Go report
- business owner sign-off or rejection
```

---

## 15. Guardrails

```text
1. Do not change backend enum/status values for UAT copy issues.
2. Do not change routes into Vietnamese.
3. Do not bypass stock ledger for UAT data corrections.
4. Do not manually edit stock balance in database.
5. Do not let QC_FAIL/QC_HOLD stock become available.
6. Do not let return items become sellable without inspection/disposition.
7. Do not close shift with unresolved operational blockers.
8. Do not code new feature requests during UAT without triage.
9. Do not treat training issue as product bug automatically.
10. Do not treat product bug as training issue to hide the problem.
```

---

## 16. Definition of Ready

Sprint 22 is ready to start when:

```text
- Sprint 21 PR #542 is merged into main.
- Sprint 21 changelog reflects completed state.
- v0.21 tag is pushed or tag-hold reason is documented.
- UAT environment is accessible.
- UAT users are created.
- Seed data plan is approved.
- UAT scripts are drafted.
- Business users are scheduled.
```

---

## 17. Definition of Done

Sprint 22 is done when:

```text
- UAT environment readiness is confirmed.
- UAT users and permissions are verified.
- Seed data is loaded and documented.
- All required UAT scripts are executed or explicitly waived.
- All findings are logged and triaged.
- P0/P1 blockers are identified with owners.
- Evidence pack is complete.
- Go/Conditional Go/No-Go decision is recorded.
- Sprint 22 changelog is created.
```

---

## 18. Suggested Next Sprint

Default next sprint after Sprint 22:

```text
Sprint 23 - UAT Fixes + Pilot Hardening
```

Sprint 23 should not add new large features. It should fix UAT-confirmed issues:

```text
- P0/P1 defects
- unclear Vietnamese labels
- confusing warehouse UX
- RBAC gaps
- seed data/import issues
- stock/QC edge cases
- UAT report findings
```

If Sprint 22 returns a clean Go result with only minor P2/P3 items, Sprint 23 can expand into:

```text
Reporting v1 + Operations Dashboard Hardening
```

But only after the UAT issue board is reviewed.

---

## 19. One-Line Rule

```text
Sprint 22 is not about adding more ERP. It is about proving that real users can operate the ERP safely.
```
