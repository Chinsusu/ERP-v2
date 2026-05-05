# ERP Platform

Phase 1 monorepo for the cosmetics ERP implementation.

## Stack

- Backend: Go modular monolith
- Worker: Go worker entrypoint in the same backend module
- Frontend: Next.js, TypeScript, Ant Design-ready structure
- Database: PostgreSQL migrations under `apps/api/migrations`
- API contract: OpenAPI under `packages/openapi`
- Infra: Docker and compose files under `infra`
- Documentation: project source-of-truth documents under `docs`

## Workspace

```text
apps/api         Go API, worker, migrations, SQL queries
apps/web         Next.js web app
packages/openapi OpenAPI source of truth
infra            Docker, compose, deployment scripts
tools            Seed, mock, import, export data
docs             ERP project documents
```

Start with:

- `docs/80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md`
- `docs/84_ERP_Coding_Task_Board_Sprint22_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md`
- `docs/85_ERP_UAT_Pilot_Pack_Sprint22_Warehouse_Sales_QC_MyPham_v1.md`
- `docs/89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md`
- `docs/90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md`
- `docs/91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md`
- `docs/92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md`
- `docs/94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md`
- `docs/88_ERP_BOM_Formula_Module_Design_MyPham_v1.md`
- `docs/82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md`
- `docs/32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` for the historical Phase 1 handoff index
- `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

## Current Status

Current line: Sprint 23 production planning and material-demand bridge after Sprint 22 Session 0 readiness.

Latest release tag:

```text
v0.19.0-vietnamese-ui-localization
```

Sprint 21 release tag status:

```text
Tag hold.
Sprint 21 is merged to main at c07409cc with CI, dev deploy, full dev smoke, and auth UI browser smoke evidence.
No v0.21.0-auth-ui-backend-integration-runtime-smoke tag has been created because target staging/pilot environment smoke evidence is still required before release tagging.
```

Sprint 22 UAT status:

```text
UAT pilot pack prepared for Warehouse + Sales + QC.
S22-ISSUE-001 role-based UAT authentication blocker resolved by PR #546 at db894ddb.
Session 0 readiness rerun passed on dev, including Warehouse/Sales/QC role login, /me role payloads, RBAC menu/route guard, logout, and invalid-login Vietnamese copy.
Business UAT execution, business issue triage, Go/No-Go decision, and v0.22 tag are pending.
Do not treat Sprint 22 preparation docs as UAT pass evidence.
```

Sprint 23 implementation status:

```text
Inventory/purchase/warehouse document-flow design is documented in file 89.
Inventory/purchase/warehouse task-board candidate is documented in file 90.
The note-sheet module roadmap and sequencing decision are documented in file 91.
Selected first Sprint 23 implementation track is documented in file 92: production planning, material demand, and Purchase Request draft.
Purchase Request workflow follow-up is documented in file 94: production plan -> Purchase Request -> submit/approve -> convert to PO -> receiving/QC traceability.
The Sprint 23 runtime bridge adds /production planning UI, backend production-plan API, active-formula snapshot, material demand/shortage calculation, internal Purchase Request lines, PostgreSQL persistence, and OpenAPI contract coverage.
Stock transfer, warehouse issue note, costing, and ledger-backed inventory dashboard implementation remain pending follow-up scope.
No v0.23 tag has been created.
```

Phase 1 production scope:

```text
The user-facing Production entrypoint at /production is for planning, active-formula snapshot, material demand, and generated Purchase Request review.
PO creation belongs to the approved Purchase Request conversion flow, not a direct /production shortcut.
The Subcontract entrypoint at /subcontract remains the external factory execution flow.
Internal work-center/MES production remains out of Phase 1 scope.
```

Latest verified release tag gate:

```text
release tag v0.19.0-vietnamese-ui-localization on commit df9b9567
required-ci on release commit df9b9567: success
required-api, required-web, required-openapi, required-migration: pass
required-migration at release tag: PostgreSQL 16 apply + rollback passed
```

Sprint 21 baseline before auth UI backend integration:

```text
main baseline 020d6a13: Sprint 20 traceability cleanup merged after required-ci success
required-migration after Sprint 20: PostgreSQL 16 apply -> rollback -> reapply passed
```

Completed focus through Sprint 21:

- Operational runtime persistence for warehouse, inventory, order, returns, purchase, subcontract, finance, and master data flows
- Auth/session runtime persistence for access sessions, refresh rotation, failed login attempts, and lockout state
- Vietnamese-first ERP UI foundation across navigation, dashboard, warehouse, sales, shipping, returns, purchase, QC, master data, inventory, auth, audit, and attachment surfaces
- Release hygiene: migration apply -> rollback -> reapply gate, GitHub Actions Node 24 compatibility, modular API route registration, and production-mode prototype fallback blocking
- Backend-backed web auth UI integration: login, `/me` session bootstrap, refresh rotation, logout, RBAC menu/route guard, Vietnamese auth errors, and production-like mock auth blocking
- Sprint 23 production planning bridge: `/production` planning surface, backend production-plan API, formula snapshot demand calculation, shortage comparison, first-class Purchase Request submit/approve/convert-to-PO flow, and controlled PO/receiving traceability
- Backend/API/DB codes, routes, enum values, permission keys, and audit event codes remain English technical contracts
- Manual PR review and merge flow, without GitHub auto-review or auto-merge

Production runtime reference:

- `docs/84_ERP_Coding_Task_Board_Sprint22_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md`
- `docs/85_ERP_UAT_Pilot_Pack_Sprint22_Warehouse_Sales_QC_MyPham_v1.md`
- `docs/86_ERP_Sprint22_Changelog_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md`
- `docs/89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md`
- `docs/90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md`
- `docs/91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md`
- `docs/92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md`
- `docs/94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md`
- `docs/88_ERP_BOM_Formula_Module_Design_MyPham_v1.md`
- `docs/82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md`
- `docs/83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md`
- `docs/78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md`
- `docs/79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md`
- `docs/81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md`

Production auth status:

```text
Web auth UI is backend-wired for the existing email/password auth surface.
Production-like deployments still require target-environment auth smoke evidence before release.
SSO, MFA, password reset email, and device/session management remain out of scope.
```

UAT preparation reference:

```text
docs/uat/sprint22/README.md
docs/uat/sprint22/templates/
```

Release tag traceability note:

```text
Sprint 16-17 runtime persistence work was merged to main and consolidated under the v0.18.0 auth/session runtime store release evidence.
```

## Local Setup

Required tools:

- Docker
- Make
- Git

First-time Docker setup:

```bash
make local-reset
```

Normal Docker restart:

```bash
make local-up
make migrate-up
make seed-local
```

This starts PostgreSQL, Redis, MinIO, Mailhog, API, worker, and web through `infra/compose/docker-compose.local.yml`.
Use `make local-reset` when you want to recreate local volumes, run migrations, seed demo data, and restart app services.

Host-based app development:

```bash
make api-dev
make worker-dev
make web-dev
```

Host-based development also requires Go, Node.js LTS, and pnpm. The default `.env.example` values work for services exposed by Docker on localhost.

Local URLs:

- Web: `http://localhost:3000`
- API: `http://localhost:8080/api/v1`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- MinIO: `http://localhost:9000`
- Mailhog: `http://localhost:8025`

Local test data:

- Mock login: `admin@example.local` / `local-only-mock-password`
- Local auth session: access token expires after 8 hours; refresh token and policy endpoints are available at `/api/v1/auth/refresh` and `/api/v1/auth/policy`.
- Seeded users: `admin@example.local`, `warehouse_user@example.local`, `sales_user@example.local`
- Seeded warehouses: `warehouse_main`, `warehouse_return`
- Seeded SKUs: `FG-LIP-001`, `FG-SER-001`, `FG-CRM-001`, `FG-SUN-001`, `PKG-BOX-001`

## Development Flow

1. Create a task branch from `main`.
2. Follow the local Codex branch prefix plus task naming, for example `codex/feature-S19-01-short-name`, `codex/fix-S19-01-short-name`, or `codex/docs-S19-01-short-name`.
3. Keep changes inside the official workspace structure.
4. Run the relevant checks before opening a pull request.
5. Open a pull request with `Primary Ref` and `Task Ref`.
6. Self-review the full diff for title/reference quality, generated-code notes, credential guardrails, tests, and docs.
7. Merge manually only after validation evidence is recorded.
8. Do not rely on GitHub auto-review or auto-merge.

## Verification

```bash
make ci-check
make smoke-test
```

`ci-check` validates OpenAPI, backend lint/tests, and frontend lint/tests.
`smoke-test` runs the Sprint 0 API and frontend smoke pack from `docs/qa/S0-13-01_smoke_test_pack.md`.

## Dev/Staging Deployment Skeleton

Shared dev and staging use Docker Compose stacks under `infra/compose`.

Prepare environment variables:

```bash
cp infra/env/dev.env.example infra/env/dev.env
cp infra/env/staging.env.example infra/env/staging.env
```

Deploy or smoke-check:

```bash
make deploy-dev
make smoke-dev
make logs-dev

make deploy-staging
make smoke-staging
make logs-staging
```

The deploy script uses environment-specific env files, runs migrations, starts API/worker/web behind an Nginx reverse proxy, writes proxy access logs, and runs post-deploy smoke checks.
