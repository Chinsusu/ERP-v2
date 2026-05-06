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
- `docs/95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md`
- `docs/96_ERP_Supplier_Invoice_Three_Way_Matching_Flow_MyPham_v1.md`
- `docs/97_ERP_AP_Payment_Readiness_Gate_Supplier_Invoice_Matching_MyPham_v1.md`
- `docs/98_ERP_Stock_Transfer_Warehouse_Issue_Runtime_Flow_MyPham_v1.md`
- `docs/99_ERP_Coding_Task_Board_Sprint24_Production_Material_Issue_Readiness_MyPham_v1.md`
- `docs/100_ERP_Production_Material_Issue_Subcontract_Readiness_Flow_MyPham_v1.md`
- `docs/101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1.md`
- `docs/102_ERP_Coding_Task_Board_Sprint25_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md`
- `docs/103_ERP_Subcontract_Finished_Goods_QC_Closeout_Flow_MyPham_v1.md`
- `docs/104_ERP_Sprint25_Changelog_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md`
- `docs/105_ERP_Coding_Task_Board_Sprint26_Production_IA_External_Factory_Order_Detail_MyPham_v1.md`
- `docs/106_ERP_Production_IA_External_Factory_Order_Detail_Flow_MyPham_v1.md`
- `docs/107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1.md`
- `docs/108_ERP_Coding_Task_Board_Sprint27_Factory_Dispatch_MyPham_v1.md`
- `docs/109_ERP_Factory_Dispatch_Flow_Sprint27_MyPham_v1.md`
- `docs/110_ERP_Sprint27_Changelog_Factory_Dispatch_MyPham_v1.md`
- `docs/111_ERP_Coding_Task_Board_Sprint28_Factory_Execution_Tracking_MyPham_v1.md`
- `docs/112_ERP_Factory_Execution_Tracking_Flow_Sprint28_MyPham_v1.md`
- `docs/113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1.md`
- `docs/88_ERP_BOM_Formula_Module_Design_MyPham_v1.md`
- `docs/82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md`
- `docs/32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` for the historical Phase 1 handoff index
- `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

## Current Status

Current line: Sprint 28 factory execution tracking for external-factory production.

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
Post-PO payable traceability is documented in file 95: posted PO-linked receiving with QC PASS lines creates supplier payable, while QC hold/fail lines do not create AP value.
Supplier invoice and 3-way matching follow-up is documented in file 96: supplier invoice remains a separate finance document linked to AP/receipt/PO traceability before payment readiness.
AP payment readiness gate is documented in file 97: request/approval/payment recording are blocked unless a matched supplier invoice exists for the AP.
Stock transfer and warehouse issue runtime flow is documented in file 98: inventory now has first-class Stock Transfer and Warehouse Issue Note documents with submit/approve/post lifecycle, PostgreSQL persistence, OpenAPI coverage, and posted stock movements.
The Sprint 23 runtime bridge adds /production planning UI, backend production-plan API, active-formula snapshot, material demand/shortage calculation, internal Purchase Request lines, PostgreSQL persistence, and OpenAPI contract coverage.
Costing, payment tolerance policy, finished goods receipt/QC automation, factory dispatch, and ledger-backed inventory dashboard implementation remain pending follow-up scope.
No v0.23 tag has been created.
```

Sprint 24 implementation status:

```text
Sprint 24 is documented in file 99, flow design is locked in file 100, and changelog/evidence is recorded in file 101.
Production-plan material demand can create source-linked Warehouse Issue Notes only for available issue-ready stock.
Shortage remains a Purchase Request / PO / receiving / QC problem before warehouse issue.
Subcontract readiness is gated on posted material issue evidence; waiver remains follow-up scope.
PR #586 merged runtime at 9e28c05e with CI green.
PR #587 fixed dev web API base at 114105b2 so browser clients can use the dev server URL.
Dev deploy passed with full smoke, and browser smoke passed for /production and production-plan detail material issue readiness.
No v0.24 tag has been created.
```

Sprint 25 implementation status:

```text
Sprint 25 is documented in file 102, flow design is locked in file 103, and changelog/evidence is tracked in file 104.
Scope is Production Plan -> source-linked Subcontract Order traceability and Production Plan detail visibility for subcontract receipt/QC/claim/final-payment closeout.
Existing subcontract finished goods receipt/QC/factory claim/final payment runtime remains the execution surface; Sprint 25 does not add internal MES or factory dispatch.
PR #589 merged at a4b96c84 with GitHub CI green.
Dev deploy passed on 2026-05-06 with migration 43 applied and full dev smoke passed.
Browser smoke passed for Production Plan detail -> source-linked Subcontract Order visibility -> /subcontract source filter.
No v0.25 tag has been created.
```

Sprint 26 implementation status:

```text
Sprint 26 is documented in file 105, and the Production IA / external factory order detail flow is locked in file 106.
User-facing production navigation is being consolidated under /production because the current company model is external-factory production, not internal MES/work-center production.
Subcontract remains the technical/legacy execution surface for external factory operations, but it should not appear as a primary sidebar sibling of Production.
Factory order detail is production-facing at /production/factory-orders/:orderId.
PR #591 merged at 5e8003a9 with GitHub CI green.
Dev deploy passed on 2026-05-06 after Docker builder cache cleanup restored /tmp free-space headroom; full dev smoke passed.
Browser smoke passed for /production sidebar consolidation and /production/factory-orders/:orderId detail.
No v0.26 tag has been created.
```

Sprint 27 implementation status:

```text
Sprint 27 is documented in file 108, and the factory dispatch flow is locked in file 109.
Scope is manual factory dispatch pack creation, ready/sent evidence, and factory response on /production/factory-orders/:orderId.
Confirmed factory response advances the external factory order to factory_confirmed.
Email, Zalo, factory portal/API delivery, digital signatures, and internal MES production remain out of scope.
PR #593 merged at 3cc5852d with GitHub CI green.
Dev deploy passed on 2026-05-06 with migration 44 applied; full dev smoke passed.
Browser smoke passed for /production/factory-orders/:orderId factory dispatch create -> ready -> sent -> confirmed.
No v0.27 tag has been created.
```

Sprint 28 implementation status:

```text
Sprint 28 is documented in file 111, and the factory execution tracking flow is locked in file 112.
Scope is a production-facing execution tracker on /production/factory-orders/:orderId after dispatch/factory confirmation.
The tracker shows current gate, next action, status metrics, and links for deposit, material handover, sample gate, mass production, finished goods receipt, QC/claim, and final payment readiness.
No new backend API, email, Zalo, supplier portal/API, or internal MES behavior is included.
PR #595 merged at cd3a5b18 with GitHub CI green.
Dev deploy passed on 2026-05-06 with no new migration; full dev smoke passed.
Browser smoke passed for /production/factory-orders/sco-s16-07-01-1777715855439203730 execution tracker.
Screenshot evidence: output/playwright/s28-factory-execution-tracker.png.
No v0.28 tag has been created.
```

Phase 1 production scope:

```text
The user-facing Production entrypoint at /production is for planning, active-formula snapshot, material demand, generated Purchase Request review, and external-factory production navigation.
PO creation belongs to the approved Purchase Request conversion flow, not a direct /production shortcut.
External factory / subcontract is the current production execution method.
/subcontract remains route-addressable for existing operational execution but is not the primary sidebar entrypoint.
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

Completed focus through Sprint 28:

- Operational runtime persistence for warehouse, inventory, order, returns, purchase, subcontract, finance, and master data flows
- Auth/session runtime persistence for access sessions, refresh rotation, failed login attempts, and lockout state
- Vietnamese-first ERP UI foundation across navigation, dashboard, warehouse, sales, shipping, returns, purchase, QC, master data, inventory, auth, audit, and attachment surfaces
- Release hygiene: migration apply -> rollback -> reapply gate, GitHub Actions Node 24 compatibility, modular API route registration, and production-mode prototype fallback blocking
- Backend-backed web auth UI integration: login, `/me` session bootstrap, refresh rotation, logout, RBAC menu/route guard, Vietnamese auth errors, and production-like mock auth blocking
- Sprint 23 production planning bridge: `/production` planning surface, backend production-plan API, formula snapshot demand calculation, shortage comparison, first-class Purchase Request submit/approve/convert-to-PO flow, and controlled PO/receiving traceability
- Post-PO payable traceability: PO-linked posted receipts create supplier payables only for QC PASS accepted lines, with AP search traceable back to PO and receipt
- Inventory warehouse document runtime: Stock Transfer and Warehouse Issue Note lifecycle with PostgreSQL document persistence and posted ledger movement effects
- External factory dispatch MVP: manual dispatch pack, ready/sent evidence, factory response, and production-facing timeline step before factory confirmation
- External factory execution tracking: production-facing current gate and worklist after dispatch confirmation, linking to deposit, material handover, sample, receiving/QC, claim, and payment execution
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
- `docs/95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md`
- `docs/96_ERP_Supplier_Invoice_Three_Way_Matching_Flow_MyPham_v1.md`
- `docs/97_ERP_AP_Payment_Readiness_Gate_Supplier_Invoice_Matching_MyPham_v1.md`
- `docs/98_ERP_Stock_Transfer_Warehouse_Issue_Runtime_Flow_MyPham_v1.md`
- `docs/99_ERP_Coding_Task_Board_Sprint24_Production_Material_Issue_Readiness_MyPham_v1.md`
- `docs/100_ERP_Production_Material_Issue_Subcontract_Readiness_Flow_MyPham_v1.md`
- `docs/101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1.md`
- `docs/102_ERP_Coding_Task_Board_Sprint25_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md`
- `docs/103_ERP_Subcontract_Finished_Goods_QC_Closeout_Flow_MyPham_v1.md`
- `docs/104_ERP_Sprint25_Changelog_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md`
- `docs/105_ERP_Coding_Task_Board_Sprint26_Production_IA_External_Factory_Order_Detail_MyPham_v1.md`
- `docs/106_ERP_Production_IA_External_Factory_Order_Detail_Flow_MyPham_v1.md`
- `docs/107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1.md`
- `docs/108_ERP_Coding_Task_Board_Sprint27_Factory_Dispatch_MyPham_v1.md`
- `docs/109_ERP_Factory_Dispatch_Flow_Sprint27_MyPham_v1.md`
- `docs/110_ERP_Sprint27_Changelog_Factory_Dispatch_MyPham_v1.md`
- `docs/111_ERP_Coding_Task_Board_Sprint28_Factory_Execution_Tracking_MyPham_v1.md`
- `docs/112_ERP_Factory_Execution_Tracking_Flow_Sprint28_MyPham_v1.md`
- `docs/113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1.md`
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
