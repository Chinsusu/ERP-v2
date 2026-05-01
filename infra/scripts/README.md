# Infra Scripts

Deployment and environment automation scripts belong here.

## Dev/Staging Deploy

- `deploy-dev-staging.sh dev` starts the shared dev stack, runs migrations, seeds dev data, starts API/worker/web/proxy, and runs smoke checks.
- `deploy-dev-staging.sh staging` starts the staging stack, runs migrations, starts API/worker/web/proxy, and runs smoke checks without resetting or seeding staging data.
- `smoke-dev-staging.sh dev|staging` verifies the reverse proxy health endpoint, API health endpoint, and web shell.
- `smoke-dev-full.sh` verifies dev health, login, warehouse dashboards, finance dashboard, report JSON, and report CSV endpoints.
- `dev-verification-preflight.sh report|cleanup|preflight` reports disk state and safely cleans only task-local verification temp paths before expensive dev verification runs.

Copy `infra/env/dev.env.example` or `infra/env/staging.env.example` to a non-committed `.env` file before real deployment.

## Dev Verification Disk Preflight

Run this before source builds, frontend installs, or full dev deploy checks when the dev server has been used for many task branches:

```sh
./infra/scripts/dev-verification-preflight.sh preflight
```

The cleanup is intentionally narrow. It only removes temp paths under `/tmp` that match:

```text
/tmp/erp-v2-verify-*
/tmp/erp-v2-s9-*
/tmp/pnpm-store-erp-v2-*
/tmp/pnpm-store-s9*
```

It does not remove Docker images, Docker volumes, application data, repository working trees outside `/tmp`, or environment files.

Optional environment variables:

```text
ERP_VERIFY_TMP_ROOT=/tmp
ERP_VERIFY_MIN_FREE_MB=2048
ERP_VERIFY_DRY_RUN=1
```

## Full Dev Smoke

Run this after a dev deploy when endpoint-level release evidence is needed:

```sh
./infra/scripts/smoke-dev-full.sh
```

`smoke-dev-staging.sh dev` runs the full dev smoke automatically after the basic host smoke.

Optional environment variables:

```text
SMOKE_BASE_URL=http://10.1.1.120:8088
SMOKE_API_BASE_URL=http://10.1.1.120:8088/api/v1
SMOKE_ACCESS_TOKEN=local-dev-access-token
SMOKE_LOGIN_EMAIL=admin@example.local
SMOKE_LOGIN_PASSWORD=local-only-mock-password
```
