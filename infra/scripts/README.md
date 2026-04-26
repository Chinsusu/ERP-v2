# Infra Scripts

Deployment and environment automation scripts belong here.

## Dev/Staging Deploy

- `deploy-dev-staging.sh dev` starts the shared dev stack, runs migrations, seeds dev data, starts API/worker/web/proxy, and runs smoke checks.
- `deploy-dev-staging.sh staging` starts the staging stack, runs migrations, starts API/worker/web/proxy, and runs smoke checks without resetting or seeding staging data.
- `smoke-dev-staging.sh dev|staging` verifies the reverse proxy health endpoint, API health endpoint, and web shell.

Copy `infra/env/dev.env.example` or `infra/env/staging.env.example` to a non-committed `.env` file before real deployment.
