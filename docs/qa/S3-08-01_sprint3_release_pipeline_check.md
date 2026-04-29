# S3-08-01 Sprint 3 Release Pipeline Check

Date: 2026-04-29

Scope: Sprint 3 Returns Reconciliation Core after merge to `main` at `b6ec843`.

## Result

Status: conditional pass for local release gates; cloud CI is blocked outside the repository.

GitHub Actions did not execute job steps because the account is blocked by billing/spending-limit state. Latest observed check annotation:

```text
The job was not started because recent account payments have failed or your spending limit needs to be increased.
```

Do not use this as production release evidence until GitHub billing is fixed and the Actions checks are rerun.

## Local Gate Evidence

| Gate | Command | Result |
|---|---|---|
| API format | `gofmt -l .` from `apps/api` | Pass, no files listed |
| API vet | `go vet ./...` from `apps/api` | Pass |
| API tests | `go test ./... -count=1` from `apps/api` | Pass |
| API build | `go build ./cmd/api ./cmd/worker` from `apps/api` | Pass |
| Web typecheck | `pnpm --filter web typecheck` | Pass |
| Web lint | `pnpm --filter web lint` | Pass |
| Web tests | `pnpm --filter web test` | Pass, 23 files / 142 tests |
| Web build | `pnpm --filter web build` | Pass |
| OpenAPI validate | `pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml` | Pass with existing license warning |
| OpenAPI generate dry-run | `pnpm dlx openapi-typescript packages/openapi/openapi.yaml -o %TEMP%/erp-openapi-schema-s3-08-01.ts` | Pass |
| API smoke | `go test ./cmd/api -run TestSprint0APISmokePack -count=1` | Pass |
| Web smoke | `pnpm --filter web test -- src/modules/smoke/sprint0Smoke.test.ts` | Pass, 3 tests |
| Migration pair static check | Count `apps/api/migrations/*.up.sql` and `*.down.sql` | Pass, 11 up / 11 down |

## Warnings And Blockers

- GitHub Actions blocker: jobs fail before any steps run due account billing/spending-limit state.
- Migration runtime apply/rollback was not executed locally because neither Docker nor `psql` is installed on this machine.
- Web build emitted the known Windows `@next/swc-win32-x64-msvc` DLL warning but completed with exit code 0.
- OpenAPI validation emitted the existing `info-license-strict` warning for proprietary license metadata and a local Node version warning; validation still completed successfully.

## Required Follow-Up Before Production Release

1. Fix GitHub Actions billing/spending-limit blocker.
2. Rerun `required-ci`, `api-ci`, `web-ci`, `openapi-ci`, `migration-ci`, and `e2e-ci` on the final release commit.
3. Run migration apply/rollback against an isolated PostgreSQL 16 instance.
4. Confirm all required checks are green before tagging `v0.3.0-returns-reconciliation-core`.
