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

- `docs/32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md`
- `docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md`

## Current Status

Current `main`: Sprint 18 completed.

Latest release candidate:

```text
v0.18.0-auth-session-runtime-store-persistence
```

Latest verified cloud gate:

```text
required-ci #981 on main commit 9112c399: success
required-api, required-web, required-openapi, required-migration: pass
required-migration: PostgreSQL 16 apply plus rollback passed
```

Completed focus through Sprint 18:

- Operational runtime persistence for warehouse, inventory, order, returns, purchase, subcontract, finance, and master data flows
- Auth/session runtime persistence for access sessions, refresh rotation, failed login attempts, and lockout state
- Manual PR review and merge flow, without GitHub auto-review or auto-merge

Current hardening focus:

- Release hygiene for tags, changelogs, and README status
- Migration CI apply, rollback, and reapply proof
- GitHub Actions Node.js 24 compatibility
- API route registration modularization
- Production gating for frontend fallback services

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
