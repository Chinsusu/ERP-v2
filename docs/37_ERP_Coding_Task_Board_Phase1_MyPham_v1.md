# 37_ERP_Coding_Task_Board_Phase1_MyPham_v1

**Project:** ERP my pham - Phase 1
**Document type:** Coding Task Board / Implementation Tracker
**Version:** v1.2
**Date:** 2026-04-26
**Status:** Active
**Primary audience:** PM, BA, Tech Lead, BE, FE, DevOps, QA, vendor

---

## 1. Purpose

This file is the official implementation task board for Phase 1 coding work.

It turns the source documents into trackable engineering tasks with stable task IDs, owners, priorities, acceptance criteria, references, and PR traceability rules.

This file replaces the previous placeholder task board.

---

## 2. Source Of Truth

| Topic | Primary source |
| --- | --- |
| Sprint 0 backlog | `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md` |
| Workspace and repo structure | `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md` |
| DevOps and CI/CD | `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` |
| QA and test gates | `docs/24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| UI template and visual direction | `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| Backend architecture | `docs/11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md` |
| Go coding standard | `docs/12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md` |
| Module boundary | `docs/13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md` |
| Frontend architecture | `docs/15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| OpenAPI standard | `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` + `packages/openapi/openapi.yaml` |
| Database standard | `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` + `apps/api/migrations` |

If this file conflicts with file 34 for Sprint 0 task IDs, file 34 wins and this file must be corrected.

---

## 3. Board Columns

```text
Backlog -> Ready -> In Progress -> Review -> QA -> Done
```

Status rules:

- `Backlog`: task exists but is not ready to start.
- `Ready`: scope, source docs, acceptance criteria, and dependency are clear.
- `In Progress`: implementation branch is active.
- `Review`: PR is open.
- `QA`: implementation merged to develop or test environment and needs QA evidence.
- `Done`: merged through the required flow, relevant CI passes, and evidence is linked.

Do not mark a task `Done` only because code exists locally.

---

## 4. Git And PR Rules

Branch naming:

```text
feature/<task-id>-short-name
fix/<task-id>-short-name
hotfix/<incident-id>-short-name
chore/<short-name>
```

Commit style:

```text
feat(<scope>): short description
fix(<scope>): short description
test(<scope>): short description
docs(<scope>): short description
chore(<scope>): short description
```

PR title:

```text
[TASK-ID] Short description
```

Every PR must include:

```text
Primary Ref: docs/<source-document>.md
Task Ref: docs/37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md#<task-id>
```

Every task below has a heading so GitHub anchors work, for example:

```text
Task Ref: docs/37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md#s0-01-01-setup-repository-and-branch-strategy
```

---

## 5. Sprint 0 Summary Board

| Task ID | Title | Epic | Owner | Priority | Status | Primary Ref |
| --- | --- | --- | --- | --- | --- | --- |
| S0-01-01 | Setup repository and branch strategy | Project foundation | Tech Lead + PM | P0 | Done | `docs/34_...` + `docs/38_...` |
| S0-01-02 | Setup issue/story board | Project foundation | PM + BA | P0 | In Progress | `docs/34_...` |
| S0-02-01 | Go backend skeleton | Backend Go foundation | BE Lead | P0 | In Progress | `docs/11_...` + `docs/12_...` |
| S0-02-02 | Module structure base | Backend Go foundation | BE Lead + Architect | P0 | Done | `docs/13_...` + `docs/38_...` |
| S0-02-03 | Error model and API response standard | Backend Go foundation | BE Lead + FE Lead | P0 | Done | `docs/16_...` |
| S0-03-01 | Next.js app shell | Frontend foundation | FE Lead | P0 | Done | `docs/15_...` + `docs/39_...` |
| S0-03-02 | Core UI components | Frontend foundation | FE Lead + UI/UX | P0 | Done | `docs/39_...` |
| S0-03-03 | Industrial Minimal ERP UI tokens | Frontend foundation | FE Lead + UI/UX | P0 | Done | `docs/39_...` |
| S0-03-04 | UI page templates foundation | Frontend foundation | FE Lead + UI/UX | P0 | Done | `docs/39_...` |
| S0-04-01 | PostgreSQL migration setup | Database foundation | BE Lead + DevOps | P0 | In Progress | `docs/17_...` |
| S0-04-02 | Base tables | Database foundation | BE Lead | P0 | Done | `docs/17_...` |
| S0-05-01 | OpenAPI base file | API contract foundation | BE Lead + FE Lead | P0 | Done | `docs/16_...` |
| S0-05-02 | API codegen integration | API contract foundation | FE Lead + BE Lead | P0 | Done | `docs/16_...` + `docs/15_...` |
| S0-06-01 | Auth skeleton | Auth, RBAC, audit | BE Lead + FE Lead | P0 | Done | `docs/19_...` |
| S0-06-02 | RBAC skeleton | Auth, RBAC, audit | BE Lead + FE Lead + BA | P0 | Done | `docs/04_...` + `docs/19_...` |
| S0-06-03 | Audit log base | Auth, RBAC, audit | BE Lead + FE Lead | P0 | Done | `docs/19_...` |
| S0-07-01 | Stock movement write path | Stock ledger prototype | BE Lead | P0 | Done | `docs/17_...` + `docs/33_...` |
| S0-07-02 | Available stock calculation prototype | Stock ledger prototype | BE Lead + FE Lead | P0 | Done | `docs/33_...` |
| S0-08-01 | Warehouse daily board skeleton | Warehouse daily board | FE Lead + BA + Warehouse Super User | P0 | Done | `docs/33_...` + `docs/39_...` |
| S0-08-02 | End-of-day reconciliation skeleton | Warehouse daily board | BE Lead + FE Lead + Warehouse Super User | P0 | Done | `docs/33_...` |
| S0-08-03 | Warehouse Daily Board UI template | Warehouse daily board | FE Lead + UI/UX | P0 | Done | `docs/39_...` |
| S0-09-01 | Carrier manifest skeleton | Shipping handover scan | BE Lead + FE Lead | P0 | Done | `docs/33_...` |
| S0-09-02 | Scan verify endpoint/UI | Shipping handover scan | BE Lead + FE Lead + Warehouse Super User | P0 | Done | `docs/33_...` |
| S0-09-03 | Shipping handover scan UI template | Shipping handover scan | FE Lead + UI/UX | P0 | Done | `docs/39_...` |
| S0-10-01 | Return receiving skeleton | Returns skeleton | BE Lead + FE Lead + Warehouse Super User | P0 | Done | `docs/33_...` |
| S0-10-02 | Return inspection UI template | Returns skeleton | FE Lead + UI/UX | P0 | Done | `docs/39_...` |
| S0-11-01 | External factory order skeleton | Subcontract manufacturing | BE Lead + FE Lead + Production/Ops Super User | P1 | Review | `docs/33_...` |
| S0-11-02 | Material transfer to factory skeleton | Subcontract manufacturing | BE Lead + FE Lead | P1 | Backlog | `docs/33_...` |
| S0-11-03 | Subcontract UI template | Subcontract manufacturing | FE Lead + UI/UX | P1 | Backlog | `docs/39_...` |
| S0-12-01 | Docker compose local | DevOps/CI/CD foundation | DevOps + Tech Leads | P0 | Done | `docs/18_...` + `docs/38_...` |
| S0-12-02 | CI pipeline | DevOps/CI/CD foundation | DevOps | P0 | Done | `docs/18_...` |
| S0-12-04 | Automated PR review gate and auto-merge | DevOps/CI/CD foundation | DevOps + Tech Lead | P0 | Done | `docs/18_...` + `docs/38_...` |
| S0-12-03 | Dev/Staging deployment skeleton | DevOps/CI/CD foundation | DevOps | P0 | Backlog | `docs/18_...` |
| S0-13-01 | Smoke test pack | QA foundation | QA Lead | P0 | Backlog | `docs/24_...` |
| S0-13-02 | Sprint 0 demo script | QA foundation | QA Lead + BA + PO | P0 | Backlog | `docs/34_...` |

---

## 6. Sprint 0 Detailed Tasks

### S0-01-01 Setup Repository And Branch Strategy

**Owner:** Tech Lead + PM
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md`, `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Repo created and connected to `git@github.com:Chinsusu/ERP-v2.git`.
- `main` and `develop` branches exist.
- Branch protection is enabled for `main` and `develop`.
- PR template exists.
- Branch naming and commit convention are documented.
- README local setup exists.

Evidence:

- PR #1: repository foundation into `develop`.
- PR #2: repository foundation promoted to `main`.
- PR #6: required CI gate promoted to `main`.

### S0-01-02 Setup Issue Story Board

**Owner:** PM + BA
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md`

Acceptance criteria:

- Board columns: Backlog, Ready, In Progress, Review, QA, Done.
- Story template includes goal, acceptance criteria, dependency, and test note.
- Bug template includes severity, reproduce steps, expected, actual, and evidence.
- Decision log exists.

Current state:

- Story and bug templates exist under `.github/ISSUE_TEMPLATE`.
- GitHub Project board and decision log still need to be created or linked.

### S0-02-01 Go Backend Skeleton

**Owner:** BE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`, `docs/12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- API runs locally.
- Healthcheck endpoint exists.
- Config reads from environment.
- Base middleware exists: request id, logging, auth placeholder, error handler.
- Standard response envelope exists.
- Folder structure follows file 12, 13, and 38.

Current state:

- `apps/api/cmd/api` and `apps/api/cmd/worker` exist.
- API health endpoint exists.
- Config package exists.
- Middleware and response envelope still need implementation.

### S0-02-02 Module Structure Base

**Owner:** BE Lead + Architect
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`, `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Modules exist for Phase 1: masterdata, purchase, inventory, qc, production, sales, shipping, returns, finance, reporting.
- Each module has handler/application/domain/repository/dto/events/queries/tests once code is added.
- No module calls another module repository directly.
- Shared packages contain only cross-module infrastructure concerns.

Current state:

- Module root folders, module README files, and tracked component folders exist.
- Inventory has an initial domain/application prototype.
- Structure guard test verifies required module/component folders and blocks direct imports of another module repository package.
- Implementation and promotion are merged.

Evidence:

- PR #71: Module structure base merged to `develop`.
- PR #72: Module structure base promoted to `main`.

### S0-02-03 Error Model And API Response Standard

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Success response is consistent.
- Error response has `code`, `message`, `details`, and `request_id`.
- Common error codes exist.
- Frontend can parse all API errors consistently.

Evidence:

- PR #9: API response standard into `develop`.
- PR #12: API response standard promoted to `main`.

### S0-03-01 Next.js App Shell

**Owner:** FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`, `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- App runs locally.
- Layout has sidebar, header, and content area.
- Login page or controlled mock login exists.
- Protected route exists.
- Menu follows permission mock.
- Design token base exists.

Current state:

- Next.js skeleton exists.
- Login page skeleton exists.
- Warehouse board placeholder exists.
- Real app shell, sidebar, header, protected route, permission menu, and module placeholders are merged to `main`.

Evidence:

- PR #21: App shell into `develop`.
- PR #22: App shell promoted to `main`.

### S0-03-02 Core UI Components

**Owner:** FE Lead + UI/UX
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Data table base.
- Form wrapper base.
- Status chip base.
- Confirm modal.
- Drawer detail.
- Toast notification.
- Empty/loading/error states.
- Scan input prototype.

Current state:

- Core reusable UI components are merged to `main` under `apps/web/src/shared/design-system`.

Evidence:

- PR #38: core UI components merged to `develop`.
- PR #39: core UI components promoted to `main`.

### S0-03-03 Industrial Minimal ERP UI Tokens

**Owner:** FE Lead + UI/UX
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Ant Design theme tokens match file 39.
- CSS variables match file 39.
- Radius, border, shadow, color, spacing, and typography rules are implemented.
- Red accent is reserved for primary/critical actions, not used as general decoration.
- UI is dense but readable.

Evidence:

- PR #15: UI token foundation into `develop`.
- PR #17: UI token foundation promoted to `main`.

### S0-03-04 UI Page Templates Foundation

**Owner:** FE Lead + UI/UX
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- App shell template.
- Page header template.
- Table template.
- Form template.
- Detail page template.
- Modal/drawer/popover template.
- Empty/loading/error state template.
- Audit log and attachment panel templates.

Current state:

- Shared page template components cover page header, filter bar, table page, form page, detail page, modal, drawer, popover, audit log panel, and attachment panel.
- Existing AppShell remains the app shell template.
- Empty/loading/error state templates are covered by shared design-system state components.
- Implementation and promotion are merged.

Evidence:

- PR #67: UI page templates foundation merged to `develop`.
- PR #68: UI page templates foundation promoted to `main`.

### S0-04-01 PostgreSQL Migration Setup

**Owner:** BE Lead + DevOps
**Priority:** P0
**Status:** In Progress
**Primary Ref:** `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Migration tool runs local/dev.
- Baseline migration exists.
- Rollback/down script exists for baseline where appropriate.
- Naming convention follows file 17.
- Migration is checked in CI.

Current state:

- Baseline migration exists.
- Migration CI exists and passes.
- Local migration tool setup still needs final verification on a developer machine with Docker/toolchain.

### S0-04-02 Base Tables

**Owner:** BE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Base tables exist for users, roles, permissions, audit, idempotency, outbox, warehouses, SKUs, batches, stock movements, stock balances, orders, shipments, manifests, scan events, returns, subcontract orders.
- Primary keys and foreign keys are defined.
- Audit columns exist where needed.
- Stock changes are not direct table updates.

Current state:

- Phase 1 base table migration is merged to `main`.
- Migration CI now applies all migration files in order and rolls back in reverse order.
- Stock ledger immutability and stock balance write guard are enforced at database level.

Evidence:

- PR #25: Phase 1 base tables into `develop`.
- PR #26: Phase 1 base tables promoted to `main`.

### S0-05-01 OpenAPI Base File

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- `packages/openapi/openapi.yaml` exists.
- `/api/v1` prefix is represented.
- Error, success, pagination, and auth schemas exist.
- Security scheme exists.
- Healthcheck, auth, and master data sample endpoints exist.
- Frontend generated client can be created.

Current state:

- Done via PR #51 and PR #52.
- Health, auth, inventory, pagination, and master data sample endpoints are represented.
- OpenAPI lint and frontend client generation dry run pass in GitHub CI.

### S0-05-02 API Codegen Integration

**Owner:** FE Lead + BE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`, `docs/15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`

Acceptance criteria:

- Generated frontend client is created from OpenAPI.
- Frontend calls API through generated client.
- CI checks OpenAPI lint and codegen dry run.
- No hard-coded API shape in frontend when schema exists.

Current state:

- Done via PR #55 and PR #56.
- Generated schema is created from OpenAPI and committed for frontend use.
- Shared frontend API wrapper uses generated GET paths, query parameters, and response data types.
- Inventory available-stock API service consumes generated schema types instead of hard-coded DTO shape.

### S0-06-01 Auth Skeleton

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Controlled login exists using seed account or mock auth.
- Token/session storage follows environment rules.
- Protected API rejects unauthenticated requests.
- Frontend redirects unauthenticated users.

Current state:

- Mock login, protected API, and frontend mock session are merged to `main`.

Evidence:

- PR #29: auth skeleton merged to `develop`.
- PR #30: auth skeleton promoted to `main`.

### S0-06-02 RBAC Skeleton

**Owner:** BE Lead + FE Lead + BA
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`, `docs/19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Roles include CEO, ERP_ADMIN, WAREHOUSE_STAFF, WAREHOUSE_LEAD, QA, SALES_OPS, PRODUCTION_OPS.
- API permission check exists.
- Frontend hides menu/action by permission.
- Permission denied returns standard `FORBIDDEN`.

Current state:

- RBAC role catalog, backend permission middleware, protected role catalog endpoint, and frontend menu/action permission filtering are merged to `main`.

Evidence:

- PR #33: RBAC skeleton merged to `develop`.
- PR #34: RBAC Go format fix merged to `develop`.
- PR #35: RBAC skeleton promoted to `main`.

### S0-06-03 Audit Log Base

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Sensitive action writes audit log.
- Audit log stores actor, action, entity type, entity id, before/after or metadata.
- Audit log screen prototype exists.
- Normal users cannot delete audit logs.

Current state:

- Audit package, audit table baseline, protected audit log API, and audit log screen prototype are implemented.
- Stock adjustment prototype API writes immutable audit metadata for actor, action, entity type, entity id, request id, and after-data.
- Implementation and promotion are merged.

Evidence:

- PR #59: Audit log base merged to `develop`.
- PR #60: Audit log base promoted to `main`.
- PR #61: Audit log task board close-out merged to `develop`.

### S0-07-01 Stock Movement Write Path

**Owner:** BE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`, `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Stock movement types exist.
- Movement is written in transaction.
- Balance updates follow movement.
- Posted movement cannot be edited or deleted.
- Adjustment writes audit log.

Current state:

- Done via PR #42 and PR #43.
- Transaction store writes stock ledger, stock balance, and adjustment audit rows under unit test coverage.

### S0-07-02 Available Stock Calculation Prototype

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Physical stock can be calculated.
- Reserved stock can be calculated.
- Available stock equals physical minus reserved minus hold where applicable.
- API returns stock by warehouse, SKU, and batch.
- UI displays stock prototype.

Current state:

- Done via PR #46, PR #47, and PR #48.
- Inventory API and UI display available stock by warehouse, SKU, and batch.

### S0-08-01 Warehouse Daily Board Skeleton

**Owner:** FE Lead + BA + Warehouse Super User
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`, `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- UI has daily warehouse work board.
- Counters include waiting, picking, packed, handover, returns, and reconciliation mismatch.
- Filter by warehouse, date, and status.
- Links to order/shipment/return prototypes.

Current state:

- Warehouse daily board skeleton has filters by warehouse, date, and status.
- Counters cover waiting, picking, packed, handover, returns, and reconciliation mismatch.
- Task rows link to order, shipment, return, and inventory prototypes.
- Implementation and promotion are merged.

Evidence:

- PR #63: Warehouse daily board skeleton merged to `develop`.
- PR #64: Warehouse daily board skeleton promoted to `main`.

### S0-08-02 End Of Day Reconciliation Skeleton

**Owner:** BE Lead + FE Lead + Warehouse Super User
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- End-of-day reconciliation prototype exists.
- Shows system quantity versus counted quantity.
- Statuses: Open, In Review, Closed.
- Closing writes audit log.
- Checklist exists before shift close.

Current state:

- Backend prototype exposes end-of-day reconciliation sessions and close action.
- Closing action records an audit log with variance and checklist summary.
- Warehouse Daily Board includes a shift closing panel with checklist and system-versus-counted quantities.
- Implementation and promotion are merged.

Evidence:

- PR #79: End-of-day reconciliation skeleton merged to `develop`.
- PR #80: End-of-day reconciliation skeleton promoted to `main`.

### S0-08-03 Warehouse Daily Board UI Template

**Owner:** FE Lead + UI/UX
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Implements file 39 section 20.
- Uses dense operational layout.
- Supports counters, active work queues, exception area, and scan-first interactions.
- Does not use decorative dashboard cards without operational action.

Current state:

- Warehouse Daily Board exposes shift context, actionable queue counters, scan station, exception lane, visible closing panel, and dense task board.
- P0/stock-variance work is sorted ahead of routine warehouse tasks.
- Implementation and promotion are merged.

Evidence:

- PR #83: Warehouse Daily Board UI template merged to `develop`.
- PR #84: Warehouse Daily Board UI template promoted to `main`.

### S0-09-01 Carrier Manifest Skeleton

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Manifest can be created by carrier, date, and warehouse.
- Shipment can be added to manifest.
- Manifest tracks expected, scanned, and missing counts.
- Statuses: Draft, Ready, Scanning, Completed, Exception.

Current state:

- Shipping module exposes carrier manifest domain/application skeleton and API endpoints for list, create, and add shipment.
- Shipping UI prototype shows manifest batches, expected/scanned/missing counters, selected manifest lines, and add-shipment action.
- Implementation and promotion are merged.

Evidence:

- PR #87: Carrier manifest skeleton merged to `develop`.
- PR #88: Carrier manifest skeleton promoted to `main`.

### S0-09-02 Scan Verify Endpoint UI

**Owner:** BE Lead + FE Lead + Warehouse Super User
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Scan order code or tracking code.
- Valid and invalid states return clear response.
- Unpacked order returns `INVALID_STATE`.
- Unknown code returns `NOT_FOUND`.
- Wrong manifest returns `MANIFEST_MISMATCH`.
- Scan event is recorded.
- UI supports scanner and keyboard speed.

Current state:

- Shipping domain/application can verify scans by order code, tracking code, shipment id, or package code.
- Backend scan result codes cover `MATCHED`, `DUPLICATE_SCAN`, `MANIFEST_MISMATCH`, `INVALID_STATE`, and `NOT_FOUND`.
- Scan events are recorded in the prototype store and audit log for matched and warning outcomes.
- Shipping UI exposes an auto-focused scan input, immediate result feedback, recent scan history, and live manifest count updates.
- Implementation and promotion are merged.

Evidence:

- PR #91: Scan verify endpoint/UI merged to `develop`.
- PR #92: Scan verify endpoint/UI promoted to `main`.

### S0-09-03 Shipping Handover Scan UI Template

**Owner:** FE Lead + UI/UX
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Implements file 39 section 24.
- Scan input is primary and always focused where appropriate.
- Success/error feedback is immediate.
- Missing/exception state is visible.

Current state:

- Shipping page now exposes a dedicated carrier handover panel with manifest status, carrier, zone, owner, expected/scanned/missing counters, and primary scan input.
- Scan feedback uses success/warning/danger border states and keeps failed scan codes selected for fast retry.
- Missing/exception queue shows missing order rows with find/report actions.
- Confirm handover action is disabled until all expected lines are scanned.
- Implementation and promotion are merged.

Evidence:

- PR #95: Shipping handover scan UI template merged to `develop`.
- PR #96: Shipping handover scan UI template promoted to `main`.

### S0-10-01 Return Receiving Skeleton

**Owner:** BE Lead + FE Lead + Warehouse Super User
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Return receiving form exists.
- User can scan order/tracking code.
- User can choose reusable, not reusable, or needs inspection.
- Reusable return can create `RETURN_RECEIPT` movement after confirmation.
- Not reusable return goes to lab/damaged placeholder.
- Audit log is written.

Current state:

- Returns module has a prototype return receiving use case and API contract for listing and creating return receipts.
- Known order, tracking, return code, and shipment scans link to expected return records; unmatched scans create unknown return cases.
- Reusable receipts create a `RETURN_RECEIPT` movement into `return_pending`; not reusable receipts route to the lab/damaged placeholder.
- Returns page exposes a scan-first receiving form with source, package condition, disposition choice, latest receipt, and receipt table.
- Return receipt creation writes audit evidence through the shared audit log store.
- Implementation and promotion are merged.

Evidence:

- PR #99: Return receiving skeleton merged to `develop`.
- PR #100: Return receiving skeleton promoted to `main`.

### S0-10-02 Return Inspection UI Template

**Owner:** FE Lead + UI/UX
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Implements file 39 section 25.
- Return condition and disposition are visible.
- Risk/status chips use text and color.
- Inspector action is clear and traceable.

Current state:

- Returns page includes a return inspection panel with receipt/order/tracking lookup and selected receipt order details.
- Return condition options cover intact, dented box, seal torn, used, damaged, and QA required states with visible tone chips.
- Inspection disposition options cover usable, not usable, and QA hold with target locations and result status chips.
- Inspector actions include confirm inspection and explicit QA escalation with a recorded result preview.
- Implementation and promotion are merged.

Evidence:

- PR #103: Return inspection UI template merged to `develop`.
- PR #104: Return inspection UI template promoted to `main`.

### S0-11-01 External Factory Order Skeleton

**Owner:** BE Lead + FE Lead + Production/Ops Super User
**Priority:** P1
**Status:** Review
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- External factory order can be created.
- Fields include factory, product, quantity, spec, sample required, expected delivery date, and deposit status.
- Status model exists: Draft, Confirmed, MaterialTransferred, SampleApproved, InProduction, Delivered, QCReview, Accepted, Rejected, Closed.
- Status change writes audit log.

Current state:

- Subcontract module exposes an external factory order skeleton at `/subcontract`.
- Order creation captures factory, product, quantity, spec, sample required, expected delivery date, deposit status, and deposit amount.
- Sprint 0 status model is implemented with visible workflow controls from Draft through Closed.
- Status changes produce a `subcontract.order.status_changed` audit log preview with before/after status.

### S0-11-02 Material Transfer To Factory Skeleton

**Owner:** BE Lead + FE Lead
**Priority:** P1
**Status:** Backlog
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Material/packaging transfer document can be created.
- Attachment placeholder exists for COA/MSDS/label/VAT invoice where needed.
- Signed handover flag exists.
- Stock movement or placeholder movement type `SUBCONTRACT_ISSUE` exists.

### S0-11-03 Subcontract UI Template

**Owner:** FE Lead + UI/UX
**Priority:** P1
**Status:** Backlog
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Implements file 39 section 27.
- List/detail pages support external factory status.
- Sample approval and factory claim blocks exist.
- Critical status and SLA are visible.

### S0-12-01 Docker Compose Local

**Owner:** DevOps + Tech Leads
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`, `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Backend, frontend, PostgreSQL, Redis, MinIO, and mail service can run locally.
- Dev seed data exists.
- README local setup is clear.
- New developer can set up within one day.

Current state:

- Compose runs app services with PostgreSQL, Redis, MinIO, and Mailhog using container service networking.
- Migration and seed commands run through Dockerized tooling.
- README documents local startup, reset, service URLs, mock login, and seeded demo data.
- Implementation and promotion are merged.

Evidence:

- PR #75: Docker compose local implementation merged to `develop`.
- PR #76: Docker compose local implementation promoted to `main`.

### S0-12-02 CI Pipeline

**Owner:** DevOps
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Backend runs format, vet, test, build.
- Frontend runs typecheck, test, build.
- OpenAPI validates and client generation dry run passes.
- Migration dry run/check passes.
- Pipeline fails if quality gate fails.

Evidence:

- `required-ci` exists and passes on `main`.
- Branch protection requires `required-api`, `required-web`, `required-openapi`, and `required-migration`.

### S0-12-04 Automated PR Review Gate And Auto Merge

**Owner:** DevOps + Tech Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`, `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Automated PR review gate validates PR title, `Primary Ref`, `Task Ref`, generated-code notes, and credential guardrails.
- Branch protection requires the automated review gate for `main` and `develop`.
- Auto-merge is enabled for non-draft same-repo PRs unless the `no-auto-merge` label is set.
- Auto-merge does not bypass required CI or human review requirements.
- PR template and README document the flow.

Evidence:

- PR #13, PR #14, PR #16, PR #19, PR #20: review gate and auto-merge workflow fixes.
- PR #21 and PR #22: review gate and auto-merge verified on real feature and promote PRs.

### S0-12-03 Dev Staging Deployment Skeleton

**Owner:** DevOps
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Deploy works to Dev or Staging.
- Environment variables are separated by environment.
- Basic healthcheck exists.
- Basic access logs exist.
- Smoke test runs after deploy.

### S0-13-01 Smoke Test Pack

**Owner:** QA Lead
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md`

Acceptance criteria:

- Smoke checklist covers login, healthcheck, master data, stock movement, and scan handover.
- Sample API test exists.
- Sample frontend smoke test exists.
- Test seed data exists.

### S0-13-02 Sprint 0 Demo Script

**Owner:** QA Lead + BA + PO
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md`

Acceptance criteria:

- Demo script has clear steps.
- Demo uses cosmetics sample data.
- Demo includes successful and failed scan cases.
- Demo includes audit log.
- Demo includes stock movement and available stock.

---

## 7. Done Evidence Log

| Task ID | Evidence |
| --- | --- |
| S0-01-01 | PR #1, PR #2, PR #6; `main` and `develop`; branch protection |
| S0-12-02 | PR #3, PR #4, PR #5, PR #6; `required-ci` pass on `main` |
| S0-02-02 | PR #71, PR #72; tracked module component folders and module boundary guard test |
| S0-02-03 | PR #9, PR #12; API response standard, shared error envelope, parsing tests |
| S0-03-03 | PR #15, PR #17; CSS variables, Ant Design theme, token tests |
| S0-12-04 | PR #13, PR #14, PR #16, PR #19, PR #20, PR #21, PR #22; review gate and auto-merge flow |
| S0-03-01 | PR #21, PR #22; protected ERP app shell, permission menu, module placeholders |
| S0-03-04 | PR #67, PR #68; page header, filter, table, form, detail, overlay, audit, and attachment templates |
| S0-04-02 | PR #25, PR #26; Phase 1 base tables, FK/check constraints, migration apply/rollback CI |
| S0-06-01 | PR #29, PR #30; mock login, protected API, frontend mock session, OpenAPI contract |
| S0-06-02 | PR #33, PR #34, PR #35; role catalog, API permission middleware, frontend menu/action permission filtering |
| S0-06-03 | PR #59, PR #60; audit package, protected audit API, audit screen prototype, and adjustment audit metadata |
| S0-03-02 | PR #38, PR #39; reusable DataTable, FormSection, StatusChip, modal, drawer, toast, state, and scan components |
| S0-07-01 | PR #42, PR #43; stock movement transaction store, balance delta updates, immutable ledger path, adjustment audit logging |
| S0-07-02 | PR #46, PR #47, PR #48; available stock calculator, inventory API endpoint, OpenAPI contract, and Inventory UI prototype |
| S0-08-01 | PR #63, PR #64; warehouse daily board skeleton, filters, counters, task links, scan station, and exception panel |
| S0-05-01 | PR #51, PR #52; OpenAPI base contract with pagination schema, auth/health coverage, master data sample endpoints, and client generation dry run |
| S0-05-02 | PR #55, PR #56; generated frontend OpenAPI schema, typed API wrapper, inventory service generated DTO integration, and CI web/OpenAPI gates |

---

## 8. Guardrails

- Do not create folders outside file 38.
- Do not edit generated code manually.
- Do not mark UI tasks done unless they follow file 39.
- Do not mark backend tasks done unless module boundary rules in file 13 hold.
- Do not merge high-risk workflow changes without relevant tests from file 24.
- Do not implement stock change by direct balance update.
- Do not create PR without `Primary Ref` and `Task Ref`.

---

## 9. Next Recommended Ready Tasks

Recommended next tasks:

1. `S0-11-01` - External factory order skeleton.
2. `S0-11-02` - Material transfer to factory skeleton.
3. `S0-11-03` - Subcontract UI template.
4. `S0-12-03` - Dev/Staging deployment skeleton.

These unlock later inventory, shipping, returns, and subcontract workflows.
