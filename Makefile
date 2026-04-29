COMPOSE = docker compose -f infra/compose/docker-compose.local.yml
MIGRATE_DSN = postgres://erp:erp@postgres:5432/erp?sslmode=disable

.PHONY: help local-up local-down local-reset local-logs deploy-dev deploy-staging smoke-dev smoke-staging smoke-test logs-dev logs-staging api-dev worker-dev web-dev api-test web-test api-lint web-lint migrate-up migrate-down seed-local openapi-generate openapi-validate openapi-contract ci-check

help:
	@echo "ERP Platform commands"
	@echo "  local-up           Start local services"
	@echo "  local-down         Stop local services"
	@echo "  local-reset        Reset local data, run migrations, and seed data"
	@echo "  local-logs         Tail local service logs"
	@echo "  deploy-dev         Deploy the shared dev skeleton"
	@echo "  deploy-staging     Deploy the staging skeleton"
	@echo "  smoke-dev          Run shared dev smoke checks"
	@echo "  smoke-staging      Run staging smoke checks"
	@echo "  smoke-test         Run Sprint 0 API and frontend smoke tests"
	@echo "  logs-dev           Tail shared dev deploy logs"
	@echo "  logs-staging       Tail staging deploy logs"
	@echo "  api-dev            Run Go API"
	@echo "  worker-dev         Run Go worker"
	@echo "  web-dev            Run Next.js web app"
	@echo "  api-test           Run backend tests"
	@echo "  web-test           Run frontend tests"
	@echo "  api-lint           Run backend lint checks"
	@echo "  web-lint           Run frontend lint checks"
	@echo "  migrate-up         Run local migrations"
	@echo "  migrate-down       Roll back one migration step"
	@echo "  seed-local         Seed local data"
	@echo "  openapi-generate   Generate API clients"
	@echo "  openapi-validate   Validate OpenAPI contract"
	@echo "  openapi-contract   Check Sprint 4 OpenAPI route coverage"
	@echo "  ci-check           Run required local checks"

local-up:
	$(COMPOSE) up -d postgres redis minio minio-init mailhog api worker web

local-down:
	$(COMPOSE) down --remove-orphans

local-reset:
	$(COMPOSE) down -v --remove-orphans
	$(COMPOSE) up -d postgres redis minio minio-init mailhog
	$(COMPOSE) --profile tools run --rm migrate
	$(COMPOSE) --profile tools run --rm seed
	$(COMPOSE) up -d api worker web

local-logs:
	$(COMPOSE) logs -f --tail=100

deploy-dev:
	./infra/scripts/deploy-dev-staging.sh dev

deploy-staging:
	./infra/scripts/deploy-dev-staging.sh staging

smoke-dev:
	./infra/scripts/smoke-dev-staging.sh dev

smoke-staging:
	./infra/scripts/smoke-dev-staging.sh staging

smoke-test:
	cd apps/api && go test ./cmd/api -run TestSprint0APISmokePack -count=1
	pnpm --filter web test -- src/modules/smoke/sprint0Smoke.test.ts

logs-dev:
	docker compose --env-file infra/env/dev.env.example -f infra/compose/docker-compose.dev.yml logs -f --tail=100 reverse-proxy api worker web

logs-staging:
	docker compose --env-file infra/env/staging.env.example -f infra/compose/docker-compose.staging.yml logs -f --tail=100 reverse-proxy api worker web

api-dev:
	cd apps/api && go run ./cmd/api

worker-dev:
	cd apps/api && go run ./cmd/worker

web-dev:
	pnpm --filter web dev

api-test:
	cd apps/api && go test ./...

web-test:
	pnpm --filter web test

api-lint:
	cd apps/api && go fmt ./... && go vet ./...

web-lint:
	pnpm --filter web lint

migrate-up:
	$(COMPOSE) --profile tools run --rm migrate

migrate-down:
	$(COMPOSE) --profile tools run --rm migrate -path /migrations -database "$(MIGRATE_DSN)" down 1

seed-local:
	$(COMPOSE) --profile tools run --rm seed

openapi-generate:
	pnpm dlx openapi-typescript packages/openapi/openapi.yaml -o apps/web/src/shared/api/generated/schema.ts

openapi-validate:
	pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml

openapi-contract:
	node packages/openapi/sprint4-contract-check.mjs

ci-check: openapi-validate openapi-contract api-lint api-test web-lint web-test
