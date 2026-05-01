# S9-06-01 Dev Release Gate

Project: Web ERP for cosmetics operations
Sprint: Sprint 9 - System hardening / production readiness core
Task: S9-06-01 End-to-end release gate smoke
Date: 2026-05-01
Status: Release gate command defined

---

## 1. Purpose

This task adds one command for the shared dev-server release gate.

The gate groups backend, OpenAPI, frontend, deploy, smoke, and evidence checks so Sprint release evidence can state exactly what passed.

---

## 2. Command

Run:

```text
make dev-release-gate
```

Equivalent direct script:

```text
./infra/scripts/dev-release-gate.sh dev
```

---

## 3. Checks Included

The release gate runs:

```text
- Disk preflight.
- Backend gofmt check, go vet, go test, and API/worker build.
- OpenAPI lint, route/envelope contract check, and generated client dry run.
- Frontend install, typecheck, tests, and production build.
- Shared dev deploy.
- Shared dev smoke checks.
- Dev deploy evidence capture.
```

Backend, OpenAPI, and frontend checks run on a clean git archive of the current commit. This keeps dependency installs and generated dry-run output out of the live repo checkout.

---

## 4. Operational Notes

The dev host does not need Go or pnpm installed directly. The script runs those checks in Docker containers:

```text
golang:1.23
node:22
pnpm 9.15.4
```

The script refuses dirty workspaces by default so release evidence maps to a specific commit.

---

## 5. Verification Notes

Expected dev-server verification:

```text
sh -n infra/scripts/dev-release-gate.sh
ERP_RELEASE_GATE_SKIP_DEPLOY=1 ./infra/scripts/dev-release-gate.sh dev
```

Final release-gate verification should run without `ERP_RELEASE_GATE_SKIP_DEPLOY=1` after the task is merged to `main`.
