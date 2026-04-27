# 37_ERP_Coding_Task_Board_Phase1_MyPham_v1

**Project:** ERP my pham - Phase 1
**Document type:** Coding Task Board / Implementation Tracker
**Version:** v1.3
**Date:** 2026-04-27
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
| Sprint 1 proposed backlog | `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md#20-sprint-1-de-xuat-sau-sprint-0` |
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
| Unit, currency, number, UOM standard | `docs/40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |

If this file conflicts with file 34 for Sprint 0 or Sprint 1 task direction, file 34 wins and this file must be corrected.

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
| S0-11-01 | External factory order skeleton | Subcontract manufacturing | BE Lead + FE Lead + Production/Ops Super User | P1 | Done | `docs/33_...` |
| S0-11-02 | Material transfer to factory skeleton | Subcontract manufacturing | BE Lead + FE Lead | P1 | Done | `docs/33_...` |
| S0-11-03 | Subcontract UI template | Subcontract manufacturing | FE Lead + UI/UX | P1 | Done | `docs/39_...` |
| S0-12-01 | Docker compose local | DevOps/CI/CD foundation | DevOps + Tech Leads | P0 | Done | `docs/18_...` + `docs/38_...` |
| S0-12-02 | CI pipeline | DevOps/CI/CD foundation | DevOps | P0 | Done | `docs/18_...` |
| S0-12-04 | Manual self-review and merge policy | DevOps/CI/CD foundation | DevOps + Tech Lead | P0 | Done | `docs/18_...` + `docs/38_...` |
| S0-12-03 | Dev/Staging deployment skeleton | DevOps/CI/CD foundation | DevOps | P0 | Done | `docs/18_...` |
| S0-13-01 | Smoke test pack | QA foundation | QA Lead | P0 | Done | `docs/24_...` |
| S0-13-02 | Sprint 0 demo script | QA foundation | QA Lead + BA + PO | P0 | Done | `docs/34_...` |

---

## 5A. Sprint 1 Summary Board

Sprint 1 starts only after Sprint 0 gate evidence exists. File 34 section 20 keeps the Sprint 1 priority order: Auth/RBAC, Master Data, Inventory Stock Ledger, Batch/QC, Warehouse Receiving, then Warehouse Daily Board v1.

| Task ID | Title | Epic | Owner | Priority | Status | Primary Ref |
| --- | --- | --- | --- | --- | --- | --- |
| S1-00-01 | Unit currency number format baseline | Cross-cutting standards | Tech Lead + BE Lead + FE Lead + QA | P0 | Done | `docs/40_...` |
| S1-00-02 | Decimal money quantity rate foundation v1 | Cross-cutting standards | BE Lead + FE Lead | P0 | Done | `docs/40_...` + `docs/16_...` |
| S1-00-03 | UOM master and conversion foundation v1 | Cross-cutting standards | BE Lead + Master Data Admin | P0 | Backlog | `docs/40_...` + `docs/05_...` + `docs/17_...` |
| S1-01-01 | Auth session and password policy v1 | Auth/RBAC hardening | BE Lead + FE Lead | P0 | Done | `docs/19_...` + `docs/34_...` |
| S1-01-02 | RBAC enforcement and permission matrix v1 | Auth/RBAC hardening | BE Lead + FE Lead + BA | P0 | Done | `docs/04_...` + `docs/19_...` |
| S1-02-01 | Item and SKU master data CRUD v1 | Master Data | BE Lead + FE Lead + MDA | P0 | Done | `docs/05_...` + `docs/16_...` |
| S1-02-02 | Warehouse and location master data CRUD v1 | Master Data | BE Lead + FE Lead + Warehouse | P0 | Done | `docs/05_...` + `docs/17_...` |
| S1-02-03 | Supplier and customer master data CRUD v1 | Master Data | BE Lead + FE Lead + Purchasing + Sales | P0 | Done | `docs/05_...` |
| S1-03-01 | Inventory stock ledger persistence v1 | Inventory Stock Ledger | BE Lead | P0 | Backlog | `docs/17_...` + `docs/33_...` + `docs/40_...` |
| S1-03-02 | Available and reserved stock service v1 | Inventory Stock Ledger | BE Lead + FE Lead | P0 | Backlog | `docs/33_...` + `docs/16_...` |
| S1-04-01 | Batch and QC status base model | Batch/QC | BE Lead + QA | P0 | Backlog | `docs/05_...` + `docs/17_...` |
| S1-04-02 | QC status transition and audit path | Batch/QC | BE Lead + QA + Internal Control | P1 | Backlog | `docs/19_...` + `docs/33_...` |
| S1-05-01 | Warehouse receiving backend v1 | Warehouse Receiving | BE Lead + Warehouse | P0 | Backlog | `docs/33_...` + `docs/16_...` |
| S1-05-02 | Warehouse receiving UI v1 | Warehouse Receiving | FE Lead + Warehouse + UI/UX | P0 | Backlog | `docs/39_...` |
| S1-06-01 | Warehouse daily board data integration v1 | Warehouse Daily Board | FE Lead + BE Lead + Warehouse | P1 | Backlog | `docs/33_...` + `docs/39_...` |

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
- `main` and `develop` are long-lived branches; enable branch protection when the GitHub plan supports it, otherwise use manual merge discipline.
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
**Status:** Done
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
- Implementation and promotion are merged.

Evidence:

- PR #107: External factory order skeleton merged to `develop`.
- PR #108: External factory order skeleton promoted to `main`.

### S0-11-02 Material Transfer To Factory Skeleton

**Owner:** BE Lead + FE Lead
**Priority:** P1
**Status:** Done
**Primary Ref:** `docs/33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md`

Acceptance criteria:

- Material/packaging transfer document can be created.
- Attachment placeholder exists for COA/MSDS/label/VAT invoice where needed.
- Signed handover flag exists.
- Stock movement or placeholder movement type `SUBCONTRACT_ISSUE` exists.

Current state:

- Subcontract module can create a material and packaging transfer for the selected external factory order.
- Transfer creation validates source warehouse, factory, QC-passed lines, and batch/lot for lot-controlled materials.
- Attachment placeholders cover COA, MSDS, label, and VAT invoice requirements.
- Signed handover flag and `SUBCONTRACT_ISSUE` stock movement placeholders are visible in the transfer result.
- Implementation and promotion are merged.

Evidence:

- PR #111: Material transfer to factory skeleton merged to `develop`.
- PR #112: Material transfer to factory skeleton promoted to `main`.

### S0-11-03 Subcontract UI Template

**Owner:** FE Lead + UI/UX
**Priority:** P1
**Status:** Done
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Implements file 39 section 27.
- List/detail pages support external factory status.
- Sample approval and factory claim blocks exist.
- Critical status and SLA are visible.

Current state:

- Subcontract page includes list filters for search, factory, product, status, and ETA.
- Detail view shows external factory status, timeline, and business tabs for transfer, sample, claim, payment, and audit.
- Sample approval block shows pass/fail samples with reviewer, file placeholder, and notes.
- Factory claim block shows severity, open status, response deadline, and SLA chip.
- Implementation and promotion are merged.

Evidence:

- PR #115: Subcontract UI template merged to `develop`.
- PR #116: Subcontract UI template promoted to `main`.

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
- Technical gates remain `required-api`, `required-web`, `required-openapi`, and `required-migration` when CI and plan support are available.

### S0-12-04 Manual Self Review And Merge Policy

**Owner:** DevOps + Tech Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`, `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- GitHub auto-review and auto-merge workflows are disabled.
- Implementer self-review validates title, `Primary Ref`, `Task Ref`, generated-code notes, credential guardrails, tests, and docs.
- Manual merge is performed only after relevant local or CI validation evidence is recorded.
- Promotion from `develop` to `main` is manual and must preserve long-lived branches.
- PR template and README document the manual review and merge flow.

Evidence:

- PR #13, PR #14, PR #16, PR #19, PR #20: historical automated workflow fixes.
- PR #21 and PR #22: historical automated workflow verification on real feature and promote PRs.
- Automated GitHub review/merge was superseded by the manual self-review and merge policy.

### S0-12-03 Dev Staging Deployment Skeleton

**Owner:** DevOps
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Deploy works to Dev or Staging.
- Environment variables are separated by environment.
- Basic healthcheck exists.
- Basic access logs exist.
- Smoke test runs after deploy.

Current state:

- Dev and staging Compose stacks define API, worker, web, PostgreSQL, Redis, MinIO, Mailhog, reverse proxy, migration, and smoke services.
- Environment examples are split into `infra/env/dev.env.example` and `infra/env/staging.env.example`.
- Nginx reverse proxy exposes `/healthz`, routes API/web traffic, and writes basic access logs.
- API exposes `/healthz` and `/readyz`, with access log middleware for request summaries.
- Deploy scripts run migrations, start services, and run internal plus host smoke checks.
- Implementation and promotion are merged.

Evidence:

- PR #119: Dev/Staging deployment skeleton merged to `develop`.
- PR #120: Dev/Staging deployment skeleton promoted to `main`.

### S0-13-01 Smoke Test Pack

**Owner:** QA Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md`

Acceptance criteria:

- Smoke checklist covers login, healthcheck, master data, stock movement, and scan handover.
- Sample API test exists.
- Sample frontend smoke test exists.
- Test seed data exists.

Current state:

- Smoke checklist lives at `docs/qa/S0-13-01_smoke_test_pack.md`.
- Sample API smoke test covers health/readiness, login, master data permission, stock movement audit, and scan handover.
- Sample frontend smoke test covers login, master data, stock movement, and scan handover route access.
- Smoke seed data lives at `tools/seed/smoke/sprint0_smoke_seed.json`.
- E2E CI runs the Sprint 0 API and frontend smoke tests.
- Implementation and promotion are merged.

Evidence:

- PR #123: Smoke test pack merged to `develop`.
- PR #124: Smoke test pack promoted to `main`.

### S0-13-02 Sprint 0 Demo Script

**Owner:** QA Lead + BA + PO
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md`

Acceptance criteria:

- Demo script has clear steps.
- Demo uses cosmetics sample data.
- Demo includes successful and failed scan cases.
- Demo includes audit log.
- Demo includes stock movement and available stock.

Current state:

- Sprint 0 demo script lives at `docs/qa/S0-13-02_sprint0_demo_script.md`.
- Script uses cosmetics sample data from `tools/seed/smoke/sprint0_smoke_seed.json`.
- Script includes successful scan, duplicate scan, wrong manifest, and unknown code cases.
- Script includes audit log evidence for stock movement and shipping scan actions.
- Script includes available stock and stock movement demo steps.
- Implementation and promotion are merged.

Evidence:

- PR #127: Sprint 0 demo script merged to `develop`.
- PR #128: Sprint 0 demo script promoted to `main`; promotion CI was blocked by a GitHub Actions billing/spending-limit infrastructure issue and the admin merge reason is recorded on the PR.

---

## 6A. Sprint 1 Detailed Tasks

### S1-01-01 Auth Session And Password Policy V1

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`, `docs/34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md`

Acceptance criteria:

- Login flow returns a real session/token contract instead of a pure mock-only skeleton.
- Session expiry and refresh behavior are documented in API and frontend code.
- Password policy covers minimum length, lockout or retry limit, and local development defaults.
- Protected API and protected frontend routes use the same session state.
- Authentication events are audit-ready and do not expose secrets in logs.

Current state:

- Backend local session manager issues access and refresh tokens with expiry.
- `/api/v1/auth/login`, `/api/v1/auth/mock-login`, `/api/v1/auth/refresh`, and `/api/v1/auth/policy` are documented in OpenAPI.
- Password policy requires minimum length, a letter, a number or symbol, and blocks common passwords.
- Login lockout applies after repeated failed attempts.
- Protected backend routes validate issued session tokens while keeping the local static token seeded for existing smoke flows.
- Frontend login validates the same local credential policy and stores an expiring HTTP-only session cookie.

Evidence:

- Self-review completed on implementation diff.
- Validation passed: API tests, API vet, frontend typecheck, frontend tests, frontend build, OpenAPI validate, generated OpenAPI schema, and Sprint 0 smoke pack.

### S1-01-02 RBAC Enforcement And Permission Matrix V1

**Owner:** BE Lead + FE Lead + BA
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`, `docs/19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Permission catalog maps Phase 1 roles to API actions and menu actions.
- Backend rejects protected actions without required permission.
- Frontend hides or disables actions using the same permission key names.
- Permission denial response uses the standard API error envelope.
- Tests cover at least one allowed and one denied path per protected module group.

Current state:

- Backend permission catalog now includes all Phase 1 module/menu permission keys, including `subcontract:view`.
- `/api/v1/rbac/permissions` exposes the permission catalog behind `settings:view`.
- Backend role permissions are tested against the known catalog to block unknown keys.
- Frontend permission catalog uses the same key names as backend/OpenAPI.
- Frontend tests verify role permissions stay inside the shared catalog and subcontract remains an operations permission.

Evidence:

- Self-review completed on implementation diff.
- Validation passed: API tests, API vet, frontend typecheck, frontend tests, OpenAPI validate, generated OpenAPI schema, and Sprint 0 smoke pack.

### S1-00-01 Unit Currency Number Format Baseline

**Owner:** Tech Lead + BE Lead + FE Lead + QA
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- File 40 is tracked in the repo as the approved baseline for unit, currency, number, UOM, rounding, import/export, and testing rules.
- Source-of-truth mapping references file 40.
- Future tasks touching money, quantity, rate, stock, UOM, import/export, or related UI formatting inherit the file 40 Definition of Done.
- Sprint 1 board includes explicit foundation tasks for decimal formatting and UOM conversion before stock ledger persistence.

Evidence:

- PR #131: unit, currency, number, and UOM baseline into `develop`.
- Local check: `git diff --check`.

### S1-00-02 Decimal Money Quantity Rate Foundation V1

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md`, `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`, `docs/12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Backend has shared decimal-backed value objects/helpers for money, quantity, rate, currency code, and UOM code.
- Backend domain code for official money/quantity/rate paths does not use `float32`, `float64`, `real`, `double precision`, or JavaScript `number` as authoritative calculation storage.
- OpenAPI defines reusable `MoneyAmount`, `Quantity`, `Rate`, `CurrencyCode`, and `UOMCode` schemas using string decimal where required.
- Frontend has shared `vi-VN` parse/format helpers and display/input components for money, quantity, rate, currency, UOM, date, and datetime.
- Tests cover decimal parsing, VND formatting, vi-VN input normalization, and rounding scale rules from file 40.

Evidence:

- PR #132: decimal money, quantity, rate, currency, and UOM foundation into `develop`.
- Local checks: API tests, API vet, frontend typecheck, frontend tests, frontend build, OpenAPI generate, OpenAPI validate, and `git diff --check`.

### S1-00-03 UOM Master And Conversion Foundation V1

**Owner:** BE Lead + Master Data Admin
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md`, `docs/05_ERP_Data_Dictionary_Master_Data_Phase1_MyPham_v1.md`, `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- UOM master and UOM conversion persistence model exists with Phase 1 seed values from file 40.
- Global conversions support `KG/G/MG` and `L/ML`.
- Item-specific conversions support pack/commercial UOM such as `BOX`, `CARTON`, and `SET`.
- Backend conversion service calculates `base_qty`, `base_uom_code`, and `conversion_factor`; frontend does not calculate authoritative stock conversion.
- Missing or inactive conversion returns a standard API error with SKU, source UOM, and base UOM details.
- Tests cover global conversion, item-specific conversion, invalid UOM, missing conversion, and base UOM passthrough.

### S1-02-01 Item And SKU Master Data CRUD V1

**Owner:** BE Lead + FE Lead + Master Data Admin
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/05_ERP_Data_Dictionary_Master_Data_Phase1_MyPham_v1.md`, `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Item/SKU create, list, detail, update, and status change paths exist.
- Duplicate item/SKU code is blocked.
- Required cosmetic master data fields follow file 5.
- OpenAPI and generated frontend types are updated through the normal generation path.
- UI includes loading, empty, error, validation, and audit hint states.

Current state:

- Backend `mdm.item` prototype catalog supports item/SKU list, create, detail, update, and status change routes at `/api/v1/products`.
- Duplicate `item_code` and `sku_code` are blocked case-insensitively before writes.
- Cosmetic item fields from file 5 are represented: item code, SKU, name, type, group, brand, base/purchase/issue UOM, lot/expiry/QC controls, shelf life, lifecycle status, cost, sellable/purchasable/producible flags, spec version, and audit timestamps.
- Create, update, and status changes write audit logs with `masterdata.item.*` actions.
- OpenAPI contract and generated frontend schema include product CRUD/status request and response types.
- Frontend `/master-data` renders item/SKU master data with filters, list/detail/edit/create/status actions, loading/empty/error states, validation feedback, and audit hint state.

Evidence:

- Self-review completed on implementation diff.
- Validation passed: API tests, API vet, frontend typecheck, frontend tests, frontend build, OpenAPI validate, generated OpenAPI schema, and Sprint 0 smoke pack.

### S1-02-02 Warehouse And Location Master Data CRUD V1

**Owner:** BE Lead + FE Lead + Warehouse
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/05_ERP_Data_Dictionary_Master_Data_Phase1_MyPham_v1.md`, `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Warehouse and location create, list, detail, update, and active/inactive paths exist.
- Location belongs to one warehouse and cannot be assigned to an invalid warehouse.
- Warehouse type and location type validations support Phase 1 inventory flows.
- API and UI expose enough data for receiving and stock ledger tasks.
- Tests cover duplicate code, invalid warehouse, and inactive location behavior.

Evidence:

- Manual merge completed after self-review.
- API warehouse/location CRUD and status routes implemented with duplicate code, invalid warehouse, invalid type/status, audit log, and inactive location guards.
- OpenAPI contract and generated frontend schema updated.
- `/master-data` UI includes warehouse/location tab with list, detail, create, update, active/inactive, filters, and operational flow flags.
- Validation passed: API tests, API vet, frontend typecheck, frontend tests, frontend build, OpenAPI validate, generated OpenAPI schema, and Sprint 0 smoke pack.

### S1-02-03 Supplier And Customer Master Data CRUD V1

**Owner:** BE Lead + FE Lead + Purchasing + Sales
**Priority:** P0
**Status:** Done
**Primary Ref:** `docs/05_ERP_Data_Dictionary_Master_Data_Phase1_MyPham_v1.md`

Acceptance criteria:

- Supplier and customer create, list, detail, update, and status change paths exist.
- Required tax/contact/address fields follow file 5.
- Duplicate code and invalid status transitions are blocked.
- UI supports search/filter and compact operational list views.
- Audit metadata is captured for create/update/status changes.

Evidence:

- Manual merge completed after self-review.
- API supplier/customer CRUD and status routes implemented with duplicate code, invalid type/group/status, invalid transition, non-negative metric/credit, and audit log guards.
- OpenAPI contract and generated frontend schema updated.
- `/master-data` UI includes supplier/customer tab with list, detail, create, update, status change, filters, contact, tax, address, payment, score, channel, price list, and credit fields.
- Validation passed: API tests, API vet, frontend typecheck, frontend tests, frontend build, OpenAPI validate, generated OpenAPI schema, and Sprint 0 smoke pack.

### S1-03-01 Inventory Stock Ledger Persistence V1

**Owner:** BE Lead
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`, `docs/33_ERP_Sprint0_Technical_Prototype_Scope_Phase1_MyPham_v1.md`, `docs/40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Stock movement write path persists through PostgreSQL transaction boundaries.
- Balance deltas are derived from immutable stock movement rows.
- Direct balance updates outside movement posting remain blocked.
- Movement quantities use decimal string API contracts, PostgreSQL `numeric(18,6)`, and base UOM fields from file 40.
- Stock ledger stores `movement_qty`, `base_uom_code`, and source quantity/UOM/conversion fields where the transaction UOM differs from base UOM.
- Unit and integration tests cover inbound, outbound, adjustment, and rollback behavior.
- Audit log captures sensitive stock movement metadata.

### S1-03-02 Available And Reserved Stock Service V1

**Owner:** BE Lead + FE Lead
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/33_ERP_Sprint0_Technical_Prototype_Scope_Phase1_MyPham_v1.md`, `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Available stock calculation includes on-hand, reserved, QC hold, and blocked quantities.
- Service supports SKU, warehouse, location, and batch filters where data exists.
- API response is OpenAPI-documented and generated frontend types are updated.
- UI shows available, reserved, blocked, and QC hold quantities without visual ambiguity.
- Tests cover zero stock, negative prevention, QC hold, and reservation behavior.

### S1-04-01 Batch And QC Status Base Model

**Owner:** BE Lead + QA
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/05_ERP_Data_Dictionary_Master_Data_Phase1_MyPham_v1.md`, `docs/17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Batch identity, expiry, and QC status fields exist in domain/database/API contracts.
- Valid QC statuses support Phase 1 receiving and inventory availability.
- Invalid batch/QC transitions are rejected.
- Batch/QC status can be used by stock availability calculations.
- Tests cover expiry, hold/release, and rejected transition cases.

### S1-04-02 QC Status Transition And Audit Path

**Owner:** BE Lead + QA + Internal Control
**Priority:** P1
**Status:** Backlog
**Primary Ref:** `docs/19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`, `docs/33_ERP_Sprint0_Technical_Prototype_Scope_Phase1_MyPham_v1.md`

Acceptance criteria:

- QC hold, release, reject, and quarantine transitions require permission.
- Transition reason and actor are mandatory.
- Audit log records before/after status and business reference.
- UI exposes transition history without allowing silent overwrite.
- Tests cover allowed, denied, and invalid transition paths.

### S1-05-01 Warehouse Receiving Backend V1

**Owner:** BE Lead + Warehouse
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/33_ERP_Sprint0_Technical_Prototype_Scope_Phase1_MyPham_v1.md`, `docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`

Acceptance criteria:

- Receiving draft, submit, inspect-ready, and posted states exist.
- Posted receiving creates stock movement rows and never updates balances directly.
- Receiving supports item/SKU, warehouse/location, batch, quantity, and reference document.
- Permission and audit checks protect posting.
- Tests cover happy path, duplicate posting, invalid location, and missing batch/QC data.

### S1-05-02 Warehouse Receiving UI V1

**Owner:** FE Lead + Warehouse + UI/UX
**Priority:** P0
**Status:** Backlog
**Primary Ref:** `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Receiving list and detail/action screens follow file 39 industrial minimal style.
- UI supports draft, submit, posted, exception, loading, empty, and error states.
- Quantity and location validation errors are visible and actionable.
- Posting action requires confirmation and shows audit/status result.
- UI tests cover route access, state rendering, and validation behavior.

### S1-06-01 Warehouse Daily Board Data Integration V1

**Owner:** FE Lead + BE Lead + Warehouse
**Priority:** P1
**Status:** Backlog
**Primary Ref:** `docs/33_ERP_Sprint0_Technical_Prototype_Scope_Phase1_MyPham_v1.md`, `docs/39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

Acceptance criteria:

- Daily board uses real receiving, stock movement, shipping, return, and exception data where available.
- Board counters have documented source fields.
- UI remains dense, scan-friendly, and operational rather than marketing-style.
- Filters support shift/date/warehouse/status.
- Tests cover empty board, active shift, exception, and mixed workload states.

---

## 7. Done Evidence Log

| Task ID | Evidence |
| --- | --- |
| S0-01-01 | PR #1, PR #2, PR #6; `main` and `develop`; manual merge discipline when branch protection is unavailable |
| S0-12-02 | PR #3, PR #4, PR #5, PR #6; `required-ci` pass on `main` |
| S0-02-02 | PR #71, PR #72; tracked module component folders and module boundary guard test |
| S0-02-03 | PR #9, PR #12; API response standard, shared error envelope, parsing tests |
| S0-03-03 | PR #15, PR #17; CSS variables, Ant Design theme, token tests |
| S0-12-04 | PR #13, PR #14, PR #16, PR #19, PR #20, PR #21, PR #22; historical automated workflow later superseded by manual self-review and merge policy |
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
| S0-13-02 | PR #127, PR #128; Sprint 0 demo script, cosmetics sample data path, success/failure scan cases, stock movement, available stock, and audit evidence |
| S1-01-01 | Manual merge; auth session manager, refresh/policy endpoints, password policy, lockout, frontend expiring session cookie, OpenAPI update, and auth tests |
| S1-01-02 | Manual merge; RBAC permission catalog endpoint, subcontract permission alignment, backend/frontend catalog drift tests, OpenAPI update |
| S1-02-01 | Manual merge; item/SKU master data CRUD/status API, duplicate code guards, audit logs, OpenAPI/generated types, and `/master-data` UI states |
| S1-02-02 | Manual merge; warehouse/location CRUD/status API, location warehouse validation, inactive location guard, OpenAPI/generated types, and `/master-data` warehouse/location UI states |
| S1-02-03 | Manual merge; supplier/customer CRUD/status API, duplicate code and invalid transition guards, audit logs, OpenAPI/generated types, and `/master-data` supplier/customer UI states |
| S1-00-01 | PR #131; file 40 tracked as approved baseline; Sprint 1 decimal/UOM foundation tasks added before stock ledger persistence |
| S1-00-02 | PR #132; backend decimal helpers, string decimal OpenAPI contracts, vi-VN frontend format/input helpers, generated schema, and decimal tests |

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

1. `S1-00-02` - Decimal money quantity rate foundation v1.
2. `S1-00-03` - UOM master and conversion foundation v1.
3. `S1-03-01` - Inventory stock ledger persistence v1.

These foundation tasks prevent stock ledger v1 from encoding prototype integer/float assumptions before persistence work starts.
