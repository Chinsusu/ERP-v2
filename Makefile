.PHONY: help local-up local-down api-dev worker-dev web-dev api-test web-test api-lint web-lint migrate-up migrate-down seed-local openapi-generate openapi-validate ci-check

help:
	@echo "ERP Platform commands"
	@echo "  local-up           Start local services"
	@echo "  local-down         Stop local services"
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
	@echo "  ci-check           Run required local checks"

local-up:
	docker compose -f infra/compose/docker-compose.local.yml up -d

local-down:
	docker compose -f infra/compose/docker-compose.local.yml down

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
	cd apps/api && ./scripts/migrate.sh up

migrate-down:
	cd apps/api && ./scripts/migrate.sh down

seed-local:
	cd apps/api && ./scripts/seed.sh local

openapi-generate:
	pnpm dlx openapi-typescript packages/openapi/openapi.yaml -o apps/web/src/shared/api/generated/schema.ts

openapi-validate:
	pnpm --package=@redocly/cli dlx redocly lint packages/openapi/openapi.yaml

ci-check: openapi-validate api-lint api-test web-lint web-test
