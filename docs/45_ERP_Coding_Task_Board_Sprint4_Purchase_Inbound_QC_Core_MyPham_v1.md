# 45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 4 - Purchase Order + Inbound QC Full Flow
Document role: Coding task board for Sprint 4 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL, Next.js frontend, OpenAPI, Redis/worker where needed
Status: Ready for Sprint 4 planning

---

## 1. Sprint 4 Context

Sprint 1 established the operational foundation:

```text
unit/currency/number standards
decimal foundation
UOM conversion
Auth/RBAC
master data
stock ledger
batch/QC model
warehouse receiving foundation
Warehouse Daily Board foundation
```

Sprint 2 delivered outbound order fulfillment:

```text
Sales Order
-> Reserve Stock
-> Pick
-> Pack
-> Carrier Manifest
-> Scan Handover DVVC
-> Warehouse Daily Board update
```

Sprint 3 delivered the reverse and closing loop:

```text
Return receiving
-> Return inspection
-> Return disposition
-> Return stock movement / quarantine
-> Stock count and adjustment
-> End-of-day reconciliation
-> Shift closing
-> Warehouse Daily Board update
```

Sprint 3 changelog says the current release state is:

```text
Dev/main merge: done
Local verification: done
Release note: done
Cloud CI: blocked by GitHub billing/spending-limit
Migration runtime apply/rollback: not verified locally
Production tag: hold
```

Therefore Sprint 4 starts with a release-gate hardening block before building the Purchase + Inbound QC flow.

---

## 2. Sprint 4 Theme

```text
Purchase Order + Inbound QC Full Flow
```

Sprint 4 completes the core inbound flow:

```text
Purchase Order
-> Supplier delivery
-> Goods receiving
-> Check quantity / packaging / lot / expiry
-> Inbound QC
-> QC Pass / Fail / Partial / Hold
-> Stock movement
-> Available stock / Quarantine / Return to supplier
-> Warehouse Daily Board update
```

Business reason:

```text
Sprint 2 controlled outbound goods.
Sprint 3 controlled returned goods.
Sprint 4 must control inbound goods.
```

For the current warehouse workflow, inbound receiving is not just "add stock". It includes checking delivery documents, quantity, packaging, lot, and final confirmation. Goods that do not pass inbound checks must be returned/rejected, quarantined, or held for QC, not silently added to sellable stock.

---

## 3. Sprint 4 Goal

By the end of Sprint 4, the system must support:

```text
1. Create and approve Purchase Orders.
2. Receive supplier deliveries against PO.
3. Capture batch/lot/expiry and packaging condition.
4. Run inbound QC inspection.
5. Convert only QC PASS goods into available stock.
6. Keep QC HOLD/PARTIAL goods in quarantine/pending status.
7. Reject QC FAIL goods and create return-to-supplier record.
8. Show inbound status on Warehouse Daily Board.
9. Persist Sprint 3 critical runtime stores where needed.
10. Clear Sprint 3 release blockers: CI and runtime migration verification.
```

---

## 4. Sprint 4 Non-Goals

Sprint 4 does not include:

```text
- Full AP accounting posting.
- Supplier portal.
- Automated supplier EDI/API integration.
- Advanced landed cost allocation.
- Full subcontract manufacturing flow.
- Production planning/MRP.
- Multi-currency purchasing.
- Full tax accounting.
```

These can be handled later in Finance Lite, Subcontract Manufacturing, or Phase 2.

---

## 5. Sprint 4 Branch / Release Commands

Before starting Sprint 4, close Sprint 3 release gate as far as possible.

```bash
git checkout main
git status
git pull
```

If GitHub billing/spending-limit and migration checks are fixed:

```bash
git tag v0.3.0-returns-reconciliation-core
git push origin v0.3.0-returns-reconciliation-core
```

Current repo workflow remains task branch -> PR -> manual self-review -> manual merge into `main`. Do not create a long-lived sprint branch unless the team explicitly changes that policy.

Default task branch pattern:

```bash
git checkout -b feature/S4-xx-yy-short-task-name
```

Recommended Sprint 4 release tag after completion:

```bash
v0.4.0-purchase-inbound-qc-core
```

---

## 6. Sprint 4 Demo Script

Final sprint demo must run these scenarios.

### Case 1: Normal inbound pass

```text
1. Create supplier.
2. Create item with base UOM and purchase UOM.
3. Create Purchase Order for 100 units.
4. Approve PO.
5. Supplier delivers 100 units.
6. Warehouse receives against PO.
7. User enters lot/batch, expiry, packaging condition, delivery note.
8. QC PASS.
9. System creates inbound stock movement.
10. Available stock increases.
11. Warehouse Daily Board shows receiving/QC pass signals.
12. Audit log captures PO approval, receiving, and QC pass.
```

### Case 2: Inbound QC fail

```text
1. Create approved PO.
2. Receive goods with wrong packaging, missing lot, damaged goods, or failed inspection.
3. QC FAIL.
4. Goods do not enter available stock.
5. System creates reject/return-to-supplier record.
6. Warehouse Daily Board shows rejected/fail inbound item.
7. Audit log captures fail reason and user.
```

### Case 3: Partial receive and partial QC

```text
1. PO quantity = 100 units.
2. Supplier delivers 80 units.
3. QC PASS 70 units.
4. QC HOLD 10 units.
5. Available stock increases only by 70.
6. 10 units stay quarantine/HOLD.
7. PO remains partially received with 20 units pending.
8. Daily Board shows partial receiving and pending QC.
```

---

## 7. Sprint 4 Guardrails

These rules are not optional.

```text
1. Receiving goods is not the same as available stock.
2. Incoming goods must pass QC before becoming available.
3. QC FAIL goods must not be reservable, pickable, packable, or sellable.
4. QC HOLD/PARTIAL goods must stay in quarantine or pending inspection.
5. Batch/lot/expiry is mandatory where item requires traceability.
6. No direct stock balance mutation.
7. All stock changes must go through stock movement service.
8. Partial receive and partial QC must be traceable.
9. Supplier rejection must create an operational record.
10. PO approval, receiving, QC transition, stock movement, and rejection must be audited.
11. Money, quantity, rate, UOM values must obey file 40 decimal/UOM standards.
12. Attachments must not expose sensitive file content in logs.
```

---

## 8. Sprint 4 Task Board

Task fields:

```text
Task ID: Unique sprint task ID.
Priority: P0 / P1 / P2.
Owner: Suggested primary owner.
Primary Ref: exactly one source-of-truth document for this task.
Output / Acceptance: what must be true before merge.
```

### 8.1 Release Gate / Sprint 3 Hardening

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-00-01 | P0 | Repo owner/DevOps | Sprint 3 release gate: fix GitHub Actions blocker | GitHub billing/spending-limit blocker cleared by account owner; full CI rerun green; evidence recorded; no production tag until green | `44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1.md` |
| S4-00-02 | P0 | BE/DevOps | Runtime migration verification on PostgreSQL 16 | Apply all up migrations and rollback all down migrations on isolated PostgreSQL 16; output evidence stored under docs/qa or docs/releases | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S4-00-03 | P0 | BE | Persist critical Sprint 3 runtime stores | Return receipts, return inspections/dispositions/attachments metadata, end-of-day reconciliations, stock counts, and stock adjustments are PostgreSQL-backed or explicitly documented as prototype-only with risk owner | `44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1.md` |
| S4-00-04 | P1 | PM/QA | Sprint 4 kickoff checklist | Task-branch workflow confirmed, Sprint 4 scope confirmed, known blockers logged, demo script accepted by business owner | `34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md` |
| S4-00-05 | P1 | BE/Tech Lead | Sprint 4 RBAC role mapping | Decide whether to add Purchasing/QC/Finance roles or temporarily map actions to current roles; approval matrix and tests reflect the decision | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |

### 8.2 Purchase Order Core

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-01-01 | P0 | BE | Purchase order model and state machine | PO header/line model exists with supplier, item, qty, UOM, unit price, expected date, warehouse, status; states: Draft, Submitted, Approved, PartiallyReceived, Received, Closed, Cancelled | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S4-01-02 | P0 | BE | Purchase order DB migration | PostgreSQL tables, indexes, constraints, optimistic locking fields, created/updated audit fields; migration up/down tested | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S4-01-03 | P0 | BE | Purchase order API | CRUD plus submit, approve, cancel, close action endpoints; response/error envelope follows API standard | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S4-01-04 | P0 | FE | Purchase order UI | PO list/detail/create/edit UI with supplier selector, line items, UOM, unit price, status chip, approval actions | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S4-01-05 | P1 | BE/QA | PO approval permission and audit tests | Only authorized roles approve/cancel/close; audit captures before/after status, actor, timestamp | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |

### 8.3 Goods Receiving

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-02-01 | P0 | BE | Harden existing goods receiving model | Existing receiving foundation is extended, not duplicated; receiving header/line links to PO, supplier, delivery note, item, batch/lot, expiry, qty, UOM, warehouse/location, packaging status | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S4-02-02 | P0 | BE | Harden existing goods receiving API | Existing `/api/v1/goods-receipts` semantics are extended for PO-linked receiving; validate supplier, item, qty, UOM, batch/lot, expiry, warehouse/location; support partial receiving | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S4-02-03 | P0 | FE | Goods receiving UI | Screen captures delivery document, quantity, packaging, lot, expiry, warehouse/location, receiving notes, attachments | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S4-02-04 | P1 | BE | Receiving status update on PO | PO moves to PartiallyReceived or Received based on received qty; never exceeds approved PO qty unless exception path exists | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S4-02-05 | P1 | QA | Receiving validation tests | Tests for over-receive, wrong supplier, missing batch, missing expiry, invalid UOM, duplicate delivery note where applicable | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

### 8.4 Inbound QC

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-03-01 | P0 | BE | Inbound QC inspection model | QC record linked to receiving line, item, batch, inspector, checklist, result: HOLD, PASS, FAIL, PARTIAL; reason and notes supported | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S4-03-02 | P0 | BE | Inbound QC API | QC action endpoints: start inspection, pass, fail, partial, hold; transitions validated; audit recorded | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S4-03-03 | P0 | FE | Inbound QC UI | QC screen shows receiving context, batch, expiry, packaging condition, checklist, attachment block, decision actions | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S4-03-04 | P1 | BE | QC transition audit | Every QC transition records actor, timestamp, old status, new status, reason, linked receiving/PO | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S4-03-05 | P1 | QA | QC transition tests | Tests reject invalid transitions and unauthorized pass/fail; PASS/FAIL/PARTIAL/HOLD must behave as specified | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

### 8.5 Stock Movement / Availability

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-04-01 | P0 | BE | QC PASS inbound stock movement | QC PASS creates controlled inbound stock movement and updates available stock through movement service only | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S4-04-02 | P0 | BE | QC FAIL/HOLD/PARTIAL stock behavior | FAIL does not increase available; HOLD/PARTIAL routes to quarantine/pending status; no reservation allowed | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S4-04-03 | P1 | BE | Batch status integration | Batch status and QC status integrate with existing reservation/pick guardrails; HOLD/FAIL batches blocked | `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md` |
| S4-04-04 | P1 | QA | Stock movement regression | Tests ensure no direct stock balance update, only movement-service updates; available qty matches QC result | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

### 8.6 Return to Supplier / Rejection Flow

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-05-01 | P0 | BE | Supplier rejection / return-to-supplier model | Reject record linked to PO/receiving/QC/supplier with reason, qty, item, batch, attachment, status | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S4-05-02 | P0 | BE | Return-to-supplier API | Create/submit/confirm return-to-supplier; rejected goods do not enter available stock | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S4-05-03 | P1 | FE | Return-to-supplier UI | UI for rejected inbound goods, reason, qty, supplier, attachments, status, audit trail | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S4-05-04 | P1 | QA | Supplier rejection E2E | Receiving -> QC FAIL -> return-to-supplier -> no available stock -> audit exists | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

### 8.7 Attachments / Object Storage

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-06-01 | P0 | BE/DevOps | S3/MinIO attachment storage hardening | Delivery notes, COA/MSDS, QC images, return evidence upload to object storage; metadata persisted; no file content in logs | `23_ERP_Integration_Spec_Phase1_MyPham_v1.md` |
| S4-06-02 | P1 | FE | Attachment UI component reuse | PO/receiving/QC/rejection screens use consistent attachment component with upload/download/delete permissions | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` |
| S4-06-03 | P1 | QA | Attachment security tests | Unauthorized access blocked; audit records upload/delete; storage metadata validation passes | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |

### 8.8 Warehouse Daily Board / Operations Signals

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-07-01 | P0 | BE | Daily Board inbound data source | Board service exposes PO incoming, receiving pending, QC hold, QC fail, QC pass, supplier rejection counts | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S4-07-02 | P0 | FE | Daily Board inbound UI update | Daily Board shows inbound cards and drill-down links using Hetzner-minimal style | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S4-07-03 | P1 | QA | Daily Board inbound regression | Board counts match source data from PO/receiving/QC/rejection tables | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

### 8.9 OpenAPI / Generated Client / Documentation

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-08-01 | P0 | BE/FE | OpenAPI update for Sprint 4 endpoints | OpenAPI includes PO, receiving, inbound QC, supplier rejection, attachments, daily board inbound endpoints | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S4-08-02 | P0 | FE | Regenerate frontend API client | Generated client compiles; FE uses typed client, not hand-written fetch wrappers for new endpoints | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| S4-08-03 | P1 | QA | OpenAPI contract validation | Redocly lint passes; BE handlers match OpenAPI; API response/error envelopes consistent | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

### 8.10 End-to-End / Release Evidence

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S4-09-01 | P0 | QA | Inbound PASS E2E | PO -> receive -> QC PASS -> stock available -> Daily Board update; test evidence recorded | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S4-09-02 | P0 | QA | Inbound FAIL E2E | PO -> receive -> QC FAIL -> no available stock -> return-to-supplier -> audit | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S4-09-03 | P0 | QA | Partial receive/QC E2E | PO 100 -> receive 80 -> pass 70/hold 10 -> available 70/quarantine 10/PO pending 20 | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S4-09-04 | P1 | QA | Permission/audit regression | PO approve, receiving, QC pass/fail/hold, supplier rejection, stock movement all audited and role-gated | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S4-09-05 | P0 | DevOps/QA | Sprint 4 release pipeline evidence | CI green, migration apply/rollback green, local verification complete, release note drafted | `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` |

---

## 9. Recommended Implementation Order

Do not start with UI-first. Sprint 4 should be built in this order:

```text
1. Fix Sprint 3 release blockers: CI + migration verification.
2. Persist required runtime stores from Sprint 3.
3. PO DB model + state machine.
4. PO API.
5. Goods receiving DB/API.
6. Inbound QC DB/API.
7. Stock movement behavior for QC PASS/HOLD/FAIL/PARTIAL.
8. Return-to-supplier flow.
9. Attachment storage hardening.
10. OpenAPI update and generated client.
11. UI screens.
12. Warehouse Daily Board update.
13. E2E/regression tests.
14. Release note and tag.
```

---

## 10. State Machines

### 10.1 Purchase Order State

```text
Draft
-> Submitted
-> Approved
-> PartiallyReceived
-> Received
-> Closed
```

Exception states:

```text
Rejected
Cancelled
```

Rules:

```text
- Draft can be edited.
- Submitted waits for approval.
- Approved can receive goods.
- PartiallyReceived means at least one receiving exists but PO qty not complete.
- Received means received qty reaches approved PO qty.
- Closed means no more receiving expected.
- Cancelled cannot receive.
```

### 10.2 Goods Receiving State

```text
Draft
-> ReceivedPendingQC
-> QCInProgress
-> QCPassed / QCFailed / QCPartial / QCHold
-> Posted / Rejected / Closed
```

Rules:

```text
- ReceivedPendingQC is not available stock.
- QCPassed creates available movement.
- QCFailed creates reject/return-to-supplier flow.
- QCPartial splits quantities into pass/hold/fail lines.
- QCHold stays quarantine/pending.
```

### 10.3 Return-to-Supplier State

```text
Draft
-> Submitted
-> Confirmed
-> ShippedToSupplier
-> Closed
```

Exception states:

```text
Cancelled
Disputed
```

---

## 11. Minimum Database Objects Expected

The exact schema should follow file 17, but Sprint 4 should introduce or harden at least:

```text
purchase_orders
purchase_order_lines
purchase_order_status_history
goods_receipts
goods_receipt_lines
inbound_qc_inspections
inbound_qc_inspection_lines
supplier_rejections
supplier_rejection_lines
attachment_objects / file_metadata if not already persistent
stock_movements extension for inbound QC source references
daily_board_inbound_view/service source query
```

Recommended source references for stock movements:

```text
source_type = PURCHASE_RECEIPT_QC_PASS
source_type = PURCHASE_RECEIPT_QC_PARTIAL_PASS
source_type = SUPPLIER_REJECT
```

---

## 12. Minimum API Endpoints Expected

Endpoint naming must follow file 16.

```text
GET    /api/v1/purchase-orders
POST   /api/v1/purchase-orders
GET    /api/v1/purchase-orders/{id}
PATCH  /api/v1/purchase-orders/{id}
POST   /api/v1/purchase-orders/{id}/submit
POST   /api/v1/purchase-orders/{id}/approve
POST   /api/v1/purchase-orders/{id}/cancel
POST   /api/v1/purchase-orders/{id}/close

POST   /api/v1/goods-receipts
GET    /api/v1/goods-receipts/{id}
POST   /api/v1/goods-receipts/{id}/submit-for-qc

POST   /api/v1/inbound-qc-inspections
GET    /api/v1/inbound-qc-inspections/{id}
POST   /api/v1/inbound-qc-inspections/{id}/pass
POST   /api/v1/inbound-qc-inspections/{id}/fail
POST   /api/v1/inbound-qc-inspections/{id}/partial
POST   /api/v1/inbound-qc-inspections/{id}/hold

POST   /api/v1/supplier-rejections
GET    /api/v1/supplier-rejections/{id}
POST   /api/v1/supplier-rejections/{id}/confirm

POST   /api/v1/attachments/presign-upload
POST   /api/v1/attachments/complete

GET    /api/v1/warehouse-daily-board/inbound-summary
```

`/api/v1/goods-receipts` already exists in the codebase. Sprint 4 must harden and extend that namespace instead of creating a parallel receiving concept with conflicting semantics.

---

## 13. UI Screens Expected

```text
1. Purchase Order List
2. Purchase Order Detail
3. Purchase Order Create/Edit
4. Goods Receiving List
5. Goods Receiving Create / Receive Against PO
6. Inbound QC Queue
7. Inbound QC Inspection Detail
8. Supplier Rejection List
9. Supplier Rejection Detail
10. Warehouse Daily Board Inbound Panel
11. Attachment Panel reused across PO/receiving/QC/rejection
```

Style:

```text
Hetzner-inspired Industrial Minimal ERP
white / grey / red accent
clear tables
dense but readable forms
status chips
no gradient
no decorative dashboard noise
```

---

## 14. Permission Notes

Minimum role rules:

```text
Current RBAC roles in code:
- ERP_ADMIN
- WAREHOUSE_STAFF
- WAREHOUSE_LEAD
- QA
- SALES_OPS
- PRODUCTION_OPS
- CEO

Sprint 4 must either add explicit Purchasing/QC/Finance roles or map actions to current roles before implementation.

Temporary mapping if no new role is added:
- PO draft/create/submit: ERP_ADMIN or authorized operations role selected in S4-00-05.
- PO approve/cancel/close: ERP_ADMIN or CEO/approval-capable role selected in S4-00-05.
- Receiving against approved PO: WAREHOUSE_STAFF, WAREHOUSE_LEAD, ERP_ADMIN.
- Receiving exception approval: WAREHOUSE_LEAD, ERP_ADMIN.
- Inbound QC inspection/pass/fail/hold/partial: QA, ERP_ADMIN.
- Master data configuration: ERP_ADMIN.
- Finance/cost visibility: blocked until Finance role or permission is added.
```

Sensitive actions requiring audit:

```text
PO approve/cancel/close
receiving submit
QC pass/fail/partial/hold
supplier rejection
attachment upload/delete
stock movement creation
manual exception approval
```

---

## 15. Test Matrix

### 15.1 Backend Unit/Integration

```text
- PO state transition tests
- PO approval permission tests
- receiving validation tests
- UOM conversion tests on receiving qty
- inbound QC transition tests
- QC PASS stock movement tests
- QC FAIL no-available-stock tests
- QC HOLD quarantine tests
- partial receiving tests
- partial QC tests
- return-to-supplier tests
- audit tests
```

### 15.2 Frontend

```text
- PO create form validation
- PO status action visibility by role
- receiving form required fields: PO, item, qty, batch, expiry, warehouse/location
- inbound QC decision form
- attachment panel permissions
- Daily Board inbound cards and drilldown
```

### 15.3 E2E

```text
- Inbound PASS flow
- Inbound FAIL flow
- Partial receive/partial QC flow
- Unauthorized QC action blocked
- Daily Board source data update
```

---

## 16. Sprint 4 Acceptance Criteria

Sprint 4 is done only when:

```text
- Sprint 3 release blockers are closed or explicitly carried with owner and risk acceptance.
- Runtime migrations apply/rollback on PostgreSQL 16.
- Purchase Order can be created, submitted, approved, cancelled, closed.
- Goods can be received against approved PO.
- Receiving captures quantity, UOM, batch/lot, expiry, packaging, delivery note.
- Inbound QC can PASS, FAIL, HOLD, PARTIAL.
- Only QC PASS quantity increases available stock.
- QC FAIL does not increase available stock and can create return-to-supplier record.
- QC HOLD/PARTIAL quantity is not sellable.
- Supplier rejection is traceable and audited.
- Attachments are stored through S3/MinIO or approved object storage path.
- Warehouse Daily Board shows inbound receiving and QC signals.
- OpenAPI is updated and generated frontend client compiles.
- E2E tests pass for PASS, FAIL, and PARTIAL scenarios.
- Audit and permission regression pass.
- Release note is written.
```

---

## 17. Definition of Done for Each Task

A Sprint 4 task is done when:

```text
- Code merged from task branch to `main` by PR and manual self-review, unless workflow is explicitly changed.
- Primary Ref has been checked and task output matches it.
- API and UI naming are consistent with existing conventions.
- Decimal/UOM/currency rules from file 40 are followed.
- No direct stock mutation is introduced.
- Audit log is added for sensitive action.
- OpenAPI updated if endpoint changes.
- DB migration up/down exists if schema changes.
- Tests added or explicitly justified as not applicable.
- CI/local verification evidence recorded.
```

---

## 18. Sprint 4 Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Sprint 3 CI still blocked | Cannot create trusted release tag | Fix billing first; do not claim production readiness without CI |
| Migration runtime not verified | DB deploy risk | Run isolated PostgreSQL 16 migration apply/rollback before tag |
| Receiving treated as available stock | Inventory contamination | Enforce QC gate before movement to available |
| Batch/expiry missing | Traceability failure | Validate required fields based on item traceability settings |
| Partial QC mishandled | Wrong available stock | Split pass/hold/fail quantities explicitly |
| Attachment storage remains prototype | Evidence risk | S4 includes object storage hardening |
| Too much accounting scope pulled in | Sprint drag | Keep AP accounting out of Sprint 4 except basic PO price fields |

---

## 19. Recommended Next Sprint After Sprint 4

After Sprint 4, the recommended next sprint is:

```text
Sprint 5 - Subcontract Manufacturing / Gia cong ngoai
```

Reason:

The current production workflow uses factory/subcontract manufacturing logic:

```text
Factory order
-> confirm quantity/spec/sample
-> deposit
-> transfer raw materials/packaging
-> sample approval
-> mass production
-> receive finished goods
-> inspect quantity/quality
-> claim factory within 3-7 days if not accepted
-> final payment
```

Sprint 5 should build that flow on top of Sprint 4 inbound receiving and QC.

---

## 20. Sprint 4 Quick Start Checklist

```text
[ ] Fix GitHub Actions billing/spending-limit.
[ ] Rerun Sprint 3 full CI.
[ ] Run PostgreSQL 16 migration apply/rollback.
[ ] Tag v0.3.0-returns-reconciliation-core if gates pass.
[ ] Keep task-branch PR workflow unless the team explicitly changes release branching.
[ ] Implement PO model/API first.
[ ] Implement receiving model/API second.
[ ] Implement inbound QC + stock movement third.
[ ] Implement UI after backend state is stable.
[ ] Run E2E PASS/FAIL/PARTIAL cases.
[ ] Draft Sprint 4 changelog.
[ ] Tag v0.4.0-purchase-inbound-qc-core after release gates pass.
```

---

End of file.
