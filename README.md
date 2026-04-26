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

## Local Setup

Required tools:

- Go
- Node.js LTS
- pnpm
- Docker
- Make
- Git

Common commands:

```bash
make local-up
make migrate-up
make seed-local
make api-dev
make worker-dev
make web-dev
```

Local URLs:

- Web: `http://localhost:3000`
- API: `http://localhost:8080/api/v1`
- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- MinIO: `http://localhost:9000`
- Mailhog: `http://localhost:8025`

## Development Flow

1. Create a task branch from `develop`.
2. Follow branch naming from `docs/38`: `feature/<task-id>-short-name`, `fix/<task-id>-short-name`, `hotfix/<incident-id>-short-name`, or `chore/<short-name>`.
3. Keep changes inside the official workspace structure.
4. Run the relevant checks before opening a pull request.
5. Open a pull request with `Primary Ref` and `Task Ref`.
6. Let the automated PR review gate validate title, references, generated-code notes, and credential guardrails.
7. Auto-merge is enabled for non-draft PRs unless the `no-auto-merge` label is set.
8. Merge still happens only after required CI and review requirements are satisfied.

## Verification

```bash
make ci-check
```

`ci-check` validates OpenAPI, backend lint/tests, and frontend lint/tests.
