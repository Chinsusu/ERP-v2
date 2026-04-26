# 38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1

**Project:** ERP mỹ phẩm – Phase 1  
**Document type:** Workspace / Repository Structure Standards  
**Version:** v1.1
**Language:** Vietnamese  
**Primary audience:** Tech Lead, Backend Developer, Frontend Developer, DevOps, QA Automation, PM/BA  
**Primary references:**
- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`
- `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`
- `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`
- `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`
- `25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md`
- `34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md`
- `37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md`
- `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md`

---

## 1. Mục tiêu tài liệu

Tài liệu này khóa cấu trúc workspace/repository cho ERP Phase 1 để đội dev bắt đầu code không bị mỗi người tạo folder một kiểu.

Nó trả lời rõ:

```text
- Repo tên gì?
- Dùng monorepo hay multi-repo?
- Backend Go đặt ở đâu?
- Worker đặt ở đâu?
- Frontend Next.js đặt ở đâu?
- OpenAPI source of truth nằm ở đâu?
- Generated client nằm ở đâu?
- DB migration nằm ở đâu?
- Docker compose nằm ở đâu?
- Env file đặt thế nào?
- Makefile có những command nào?
- CI/CD chạy theo path nào?
- Tài liệu 01–38 lưu ở đâu?
- Quy tắc tạo module/folder mới là gì?
```

Nguyên tắc thiết kế workspace:

```text
Một repo, nhiều app, một chuẩn.
Ít folder nhưng rõ trách nhiệm.
Không tạo thư mục vì cảm giác tiện.
Chỉ tạo thư mục khi có boundary rõ.
```

---

## 2. Quyết định chính thức

### 2.1. Repository strategy

Phase 1 dùng **monorepo**.

Tên repo đề xuất:

```text
erp-platform
```

Lý do chọn monorepo:

```text
- Backend, frontend, OpenAPI, infra, docs đang phụ thuộc chặt.
- Phase 1 cần tốc độ triển khai và đồng bộ contract.
- OpenAPI thay đổi phải kéo theo frontend generated client và backend implementation.
- DevOps/CI/CD dễ kiểm soát hơn trong giai đoạn đầu.
- Tài liệu 01–38 cần nằm cạnh code để dev đối chiếu nhanh.
```

Không dùng multi-repo ở Phase 1, trừ khi sau này hệ thống scale lớn và có team riêng cho từng domain.

---

### 2.2. Backend strategy

Backend dùng:

```text
Go
Modular Monolith
PostgreSQL
Redis
OpenAPI
```

Backend API và worker nằm chung một Go module:

```text
apps/api
```

Không tách `apps/worker` thành app riêng ở Phase 1.

Lý do:

```text
- API và worker dùng chung domain logic.
- Worker cần dùng chung repository, transaction, audit, event, config.
- Tách worker quá sớm dễ duplicate business logic.
- Modular monolith nên giữ domain chung, chỉ tách entrypoint.
```

Entry point:

```text
apps/api/cmd/api/main.go
apps/api/cmd/worker/main.go
```

---

### 2.3. Frontend strategy

Frontend dùng:

```text
React / Next.js
TypeScript
Ant Design + ERP design tokens
TanStack Query
React Hook Form + Zod
OpenAPI generated client
```

Frontend app đặt tại:

```text
apps/web
```

---

### 2.4. OpenAPI strategy

OpenAPI là **API contract source of truth**.

Nguồn chính:

```text
packages/openapi/openapi.yaml
```

Generated client cho frontend:

```text
apps/web/src/shared/api/generated
```

Generated type/contract cho backend nếu dùng:

```text
apps/api/internal/shared/openapi/generated
```

Không sửa generated code bằng tay.

---

### 2.5. Database migration strategy

Migration source of truth nằm trong backend:

```text
apps/api/migrations
```

Lý do:

```text
- DB schema phục vụ backend domain.
- Migration chạy trong CI/CD backend.
- Repository/query của Go phải đi cùng schema.
```

---

### 2.6. Documentation strategy

Tất cả tài liệu dự án lưu tại:

```text
docs/
```

Không để tài liệu rải rác ở root.

File docs phải giữ naming convention đang dùng:

```text
NN_ERP_<Document_Name>_Phase1_MyPham_vX.md
```

Ví dụ:

```text
docs/38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md
```

---

## 3. Cấu trúc workspace chính thức

```text
erp-platform/
  apps/
    api/
      cmd/
        api/
          main.go
        worker/
          main.go

      internal/
        shared/
          auth/
          authorization/
          approval/
          audit/
          config/
          database/
          errors/
          events/
          files/
          idempotency/
          logger/
          middleware/
          notifications/
          openapi/
            generated/
          pagination/
          response/
          security/
          transaction/
          validation/

        modules/
          masterdata/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          purchase/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          inventory/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          qc/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          production/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          sales/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          shipping/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          returns/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          finance/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

          reporting/
            handler/
            application/
            domain/
            repository/
            dto/
            events/
            queries/
            tests/

      migrations/
        000001_init.up.sql
        000001_init.down.sql

      sql/
        queries/
          masterdata.sql
          purchase.sql
          inventory.sql
          qc.sql
          production.sql
          sales.sql
          shipping.sql
          returns.sql
          finance.sql
          reporting.sql

      scripts/
        dev.sh
        test.sh
        migrate.sh
        seed.sh

      go.mod
      go.sum
      README.md

    web/
      src/
        app/
          (auth)/
            login/
              page.tsx
          (erp)/
            dashboard/
              page.tsx
            master-data/
              page.tsx
            purchase/
              page.tsx
            inventory/
              page.tsx
            qc/
              page.tsx
            production/
              page.tsx
            sales/
              page.tsx
            shipping/
              page.tsx
            returns/
              page.tsx
            finance/
              page.tsx
            settings/
              page.tsx

        modules/
          master-data/
            components/
            hooks/
            pages/
            schemas/
            services/
            types/
          purchase/
          inventory/
          qc/
          production/
          sales/
          shipping/
          returns/
          finance/
          reporting/
          settings/

        shared/
          api/
            generated/
            client.ts
            queryKeys.ts
          auth/
          components/
            data-table/
            form/
            status-chip/
            scanner-input/
            audit-timeline/
            attachment-panel/
            approval-panel/
          constants/
          design-system/
            tokens.ts
            theme.ts
          hooks/
          layouts/
          permissions/
          utils/

      public/
      package.json
      next.config.ts
      tsconfig.json
      README.md

  packages/
    openapi/
      openapi.yaml
      generated/
      README.md

    shared-types/
      README.md

  infra/
    docker/
      Dockerfile.api
      Dockerfile.web
      Dockerfile.worker

    compose/
      docker-compose.local.yml
      docker-compose.dev.yml
      docker-compose.uat.yml

    nginx/
      nginx.local.conf
      nginx.uat.conf
      nginx.prod.conf

    k8s/
      README.md

    scripts/
      backup-db.sh
      restore-db.sh
      wait-for-db.sh
      healthcheck.sh

  tools/
    seed/
      masterdata/
      warehouse/
      sample-orders/
      sample-returns/
      subcontract/

    import/
      excel/
      csv/

    export/
      reports/

    mock-data/
      warehouse-daily-board.json
      carrier-manifest.json
      return-receiving.json
      subcontract-order.json

  docs/
    01_ERP_Blueprint_My_Pham_v1.md
    02_ERP_Tai_Lieu_Tiep_Theo_My_Pham_v1.md
    03_ERP_PRD_SRS_Phase1_My_Pham_v1.md
    ...
    38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md

  .github/
    workflows/
      api-ci.yml
      web-ci.yml
      openapi-ci.yml
      migration-ci.yml
      e2e-ci.yml
      release.yml

  .vscode/
    settings.json
    extensions.json

  .env.example
  .gitignore
  Makefile
  README.md
  package.json
  pnpm-workspace.yaml
```

---

## 4. Root workspace rules

### 4.1. Root files

Root chỉ được chứa các file điều phối toàn repo:

```text
README.md
Makefile
.env.example
.gitignore
package.json
pnpm-workspace.yaml
```

Không đặt source code nghiệp vụ ở root.

---

### 4.2. Root README.md

Root README phải có:

```text
- Project overview
- Tech stack
- Workspace structure
- Local setup
- Makefile commands
- How to run API
- How to run Web
- How to run DB migration
- How to generate OpenAPI client
- How to run tests
- Link tới docs/32 và docs/37
```

---

### 4.3. Root package.json

Root `package.json` chỉ dùng để quản lý workspace JS/TS command.

Ví dụ:

```json
{
  "name": "erp-platform",
  "private": true,
  "scripts": {
    "web:dev": "pnpm --filter web dev",
    "web:build": "pnpm --filter web build",
    "web:test": "pnpm --filter web test",
    "openapi:generate": "pnpm --filter openapi generate"
  },
  "devDependencies": {}
}
```

Không nhét backend Go dependency vào `package.json`.

---

### 4.4. pnpm-workspace.yaml

```yaml
packages:
  - "apps/web"
  - "packages/*"
```

---

## 5. Backend Go workspace standard

### 5.1. Backend app path

```text
apps/api
```

### 5.2. Go module name

Đề xuất:

```text
github.com/<company>/erp-platform/apps/api
```

Nếu chưa có GitHub org, tạm dùng:

```text
erp-platform/apps/api
```

Sau khi repo chính thức tạo xong, đổi module path cho chuẩn.

---

### 5.3. Backend entrypoints

```text
cmd/api/main.go
cmd/worker/main.go
```

`cmd/api` dùng cho HTTP API.

`cmd/worker` dùng cho:

```text
- async job
- outbox event publishing
- report export
- notification dispatch
- integration sync
- COD reconciliation job
- stock aging job
- near-expiry alert job
```

---

### 5.4. Backend shared folder

```text
internal/shared
```

Chỉ chứa logic dùng chung thật sự.

Không được biến `shared` thành bãi rác.

Cho phép:

```text
auth/
authorization/
approval/
audit/
config/
database/
errors/
events/
files/
idempotency/
logger/
middleware/
notifications/
pagination/
response/
security/
transaction/
validation/
```

Không cho phép:

```text
shared/helpers/random.go
shared/utils/common.go
shared/services/global_service.go
```

Tên quá chung là dấu hiệu thiết kế lỏng.

---

### 5.5. Backend module folder

Mỗi module nghiệp vụ phải theo layout:

```text
module-name/
  handler/
  application/
  domain/
  repository/
  dto/
  events/
  queries/
  tests/
```

Ý nghĩa:

```text
handler/       HTTP handlers, request parsing, response mapping
application/  use case orchestration, transaction boundary
domain/       entity, value object, domain service, policy, state machine
repository/   DB access implementation
dto/          request/response/input/output structs
events/       domain event definitions
queries/      read model queries, report queries
tests/        module-specific tests
```

---

### 5.6. Backend module list Phase 1

```text
masterdata
purchase
inventory
qc
production
sales
shipping
returns
finance
reporting
```

Trong đó:

```text
inventory owns stock ledger, stock balance, stock reservation.
qc owns QC status and release decision.
shipping owns carrier manifest and handover scan.
returns owns return receiving and return disposition.
production owns subcontract manufacturing workflow in Phase 1.
```

---

### 5.7. Backend module boundary rule

Module khác **không được update trực tiếp table của module khác**.

Ví dụ:

```text
sales không được tự update stock_balance.
sales phải gọi inventory application service hoặc command contract.
```

Đúng:

```text
sales.application.ConfirmSalesOrder()
→ inventory.application.ReserveStock()
```

Sai:

```text
sales.repository.UpdateStockBalanceDirectly()
```

---

### 5.8. Transaction boundary rule

Transaction đặt tại `application/`, không đặt ở `handler/`.

Ví dụ:

```text
handler nhận request
→ gọi application use case
→ application mở transaction
→ gọi domain/repository
→ ghi audit/outbox trong cùng transaction
→ commit
```

Các use case bắt buộc transaction:

```text
- nhập kho
- QC pass/fail/hold
- reserve stock
- issue stock
- transfer stock
- sales order confirm/cancel
- pick/pack/handover
- return receiving/disposition
- subcontract material issue
- subcontract finished goods receipt
- finance settlement
```

---

### 5.9. Stock ledger placement

Stock ledger nằm trong:

```text
apps/api/internal/modules/inventory
```

Các folder gợi ý:

```text
inventory/
  domain/
    stock_movement.go
    stock_balance.go
    stock_reservation.go
    stock_policy.go
    batch.go
  application/
    record_stock_movement.go
    reserve_stock.go
    release_reservation.go
    issue_stock.go
    receive_stock.go
    transfer_stock.go
  repository/
    stock_movement_repository.go
    stock_balance_repository.go
```

Nguyên tắc:

```text
Không sửa trực tiếp tồn kho.
Mọi thay đổi tồn phải đi qua stock movement.
Stock ledger là immutable.
```

---

## 6. Frontend Next.js workspace standard

### 6.1. Frontend app path

```text
apps/web
```

---

### 6.2. Frontend route groups

```text
src/app/(auth)
src/app/(erp)
```

`(auth)` chứa login, forgot password, reset password nếu có.

`(erp)` chứa layout chính sau khi đăng nhập.

---

### 6.3. Frontend module folder

Mỗi module frontend đặt tại:

```text
src/modules/<module-name>
```

Layout chuẩn:

```text
module-name/
  components/
  hooks/
  pages/
  schemas/
  services/
  types/
```

Ý nghĩa:

```text
components/  component riêng của module
hooks/       React hooks của module
pages/       page container nếu không đặt trực tiếp ở app route
schemas/     Zod schemas
services/    API service wrapper nếu cần
types/       frontend-only types
```

---

### 6.4. Frontend shared folder

```text
src/shared
```

Cho phép:

```text
api/
auth/
components/
constants/
design-system/
hooks/
layouts/
permissions/
utils/
```

Không cho phép nhét business module vào `shared`.

Ví dụ sai:

```text
src/shared/components/SalesOrderForm.tsx
```

Phải đặt ở:

```text
src/modules/sales/components/SalesOrderForm.tsx
```

---

### 6.5. Generated API client

Generated API client đặt tại:

```text
apps/web/src/shared/api/generated
```

Không sửa bằng tay.

Nếu cần wrapper riêng cho UI:

```text
apps/web/src/modules/<module>/services/<module>Service.ts
```

---

### 6.6. ERP shared UI components

Các component dùng chung đặt tại:

```text
apps/web/src/shared/components
```

Danh sách bắt buộc Phase 1:

```text
data-table/
form/
status-chip/
scanner-input/
audit-timeline/
attachment-panel/
approval-panel/
batch-selector/
warehouse-location-picker/
quantity-input/
confirm-action-dialog/
empty-state/
error-state/
loading-state/
```

---

## 7. OpenAPI workspace standard

### 7.1. OpenAPI path

```text
packages/openapi/openapi.yaml
```

Nếu file lớn, được tách theo folder:

```text
packages/openapi/
  openapi.yaml
  paths/
    masterdata.yaml
    inventory.yaml
    qc.yaml
    sales.yaml
    shipping.yaml
    returns.yaml
    production.yaml
  components/
    schemas.yaml
    responses.yaml
    parameters.yaml
    security.yaml
```

Nhưng file build cuối vẫn phải xuất ra:

```text
packages/openapi/openapi.yaml
```

---

### 7.2. OpenAPI generated outputs

Frontend:

```text
apps/web/src/shared/api/generated
```

Backend optional:

```text
apps/api/internal/shared/openapi/generated
```

Docs/generated:

```text
packages/openapi/generated
```

---

### 7.3. OpenAPI rules

```text
- API phải dùng /api/v1
- Response envelope thống nhất
- Error envelope thống nhất
- Action endpoint phải rõ động từ nghiệp vụ
- Không dùng endpoint mơ hồ
- Không đổi breaking contract nếu chưa update version/change log
```

Ví dụ đúng:

```text
POST /api/v1/shipping/manifests/{id}/scan
POST /api/v1/shipping/manifests/{id}/handover
POST /api/v1/returns/{id}/inspect
POST /api/v1/qc/inspections/{id}/pass
```

Ví dụ sai:

```text
POST /api/v1/do-action
POST /api/v1/update-status
```

---

## 8. Database workspace standard

### 8.1. Migration folder

```text
apps/api/migrations
```

Naming:

```text
000001_init.up.sql
000001_init.down.sql
000002_create_masterdata_tables.up.sql
000002_create_masterdata_tables.down.sql
```

---

### 8.2. SQL query folder

```text
apps/api/sql/queries
```

Nếu dùng `sqlc`, query file đặt theo module:

```text
masterdata.sql
inventory.sql
qc.sql
sales.sql
shipping.sql
returns.sql
production.sql
```

Generated code nếu dùng `sqlc` đặt tại:

```text
apps/api/internal/shared/database/generated
```

hoặc theo module nếu muốn tách:

```text
apps/api/internal/modules/inventory/repository/generated
```

Quyết định Phase 1 khuyến nghị:

```text
Module repository tự sở hữu query generated của module đó.
```

---

## 9. Infra workspace standard

### 9.1. Docker folder

```text
infra/docker
```

Files:

```text
Dockerfile.api
Dockerfile.web
Dockerfile.worker
```

---

### 9.2. Docker Compose folder

```text
infra/compose
```

Files:

```text
docker-compose.local.yml
docker-compose.dev.yml
docker-compose.uat.yml
```

Local compose phải có tối thiểu:

```text
postgres
redis
minio
mailhog hoặc mailpit
api
worker
web
```

---

### 9.3. Infra scripts

```text
infra/scripts
```

Scripts:

```text
backup-db.sh
restore-db.sh
wait-for-db.sh
healthcheck.sh
```

Scripts phải idempotent nếu có thể.

---

## 10. Tools workspace standard

### 10.1. Seed data

```text
tools/seed
```

Seed data phải phục vụ workflow Phase 1:

```text
masterdata/
warehouse/
sample-orders/
sample-returns/
subcontract/
```

---

### 10.2. Mock data

```text
tools/mock-data
```

File mock bắt buộc:

```text
warehouse-daily-board.json
carrier-manifest.json
return-receiving.json
subcontract-order.json
```

Các mock này phục vụ 4 flow thực tế:

```text
- công việc hằng ngày kho
- nhập/xuất/đóng hàng/hàng hoàn
- bàn giao ĐVVC bằng quét mã
- gia công ngoài với nhà máy
```

---

### 10.3. Import/export tools

```text
tools/import
tools/export
```

Dùng cho:

```text
- migration dữ liệu Excel/CSV
- export report
- test migration
- data correction simulation
```

---

## 11. Environment file standard

### 11.1. Root env example

Root có:

```text
.env.example
```

Không commit file thật:

```text
.env
.env.local
.env.production
```

---

### 11.2. Backend env

Prefix:

```text
APP_
DB_
REDIS_
JWT_
S3_
SMTP_
OPENAPI_
```

Ví dụ:

```env
APP_ENV=local
APP_PORT=8080
APP_NAME=erp-api

DB_HOST=localhost
DB_PORT=5432
DB_NAME=erp
DB_USER=erp
DB_PASSWORD=erp
DB_SSL_MODE=disable

REDIS_URL=redis://localhost:6379/0

JWT_SECRET=change_me
JWT_ACCESS_TOKEN_TTL_MINUTES=30
JWT_REFRESH_TOKEN_TTL_DAYS=7

S3_ENDPOINT=http://localhost:9000
S3_BUCKET=erp-files
S3_ACCESS_KEY=minio
S3_SECRET_KEY=minio123
S3_USE_PATH_STYLE=true
```

---

### 11.3. Frontend env

Next.js public env phải dùng prefix:

```text
NEXT_PUBLIC_
```

Ví dụ:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_APP_NAME=ERP Mỹ Phẩm
```

Không đưa secret vào frontend env.

---

## 12. Makefile standard

Root `Makefile` là command center.

Bắt buộc có các command sau:

```makefile
help:
	@echo "Available commands"

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
	cd apps/api && golangci-lint run

web-lint:
	pnpm --filter web lint

migrate-up:
	cd apps/api && ./scripts/migrate.sh up

migrate-down:
	cd apps/api && ./scripts/migrate.sh down

seed-local:
	cd apps/api && ./scripts/seed.sh local

openapi-generate:
	pnpm --filter openapi generate

openapi-validate:
	pnpm --filter openapi validate

ci-check:
	make openapi-validate && make api-lint && make api-test && make web-lint && make web-test
```

Không bắt dev nhớ command dài. Tất cả command chính đi qua Makefile.

---

## 13. CI/CD workspace path rules

### 13.1. API CI trigger

File:

```text
.github/workflows/api-ci.yml
```

Trigger khi thay đổi:

```text
apps/api/**
packages/openapi/**
infra/docker/Dockerfile.api
.github/workflows/api-ci.yml
```

---

### 13.2. Web CI trigger

File:

```text
.github/workflows/web-ci.yml
```

Trigger khi thay đổi:

```text
apps/web/**
packages/openapi/**
packages/shared-types/**
.github/workflows/web-ci.yml
```

---

### 13.3. OpenAPI CI trigger

File:

```text
.github/workflows/openapi-ci.yml
```

Trigger khi thay đổi:

```text
packages/openapi/**
```

Phải chạy:

```text
openapi validate
frontend client generate dry-run
backend contract check nếu có
```

---

### 13.4. Migration CI trigger

File:

```text
.github/workflows/migration-ci.yml
```

Trigger khi thay đổi:

```text
apps/api/migrations/**
```

Phải chạy:

```text
migration up
migration down
migration up again
basic seed smoke test
```

### 13.5. Automated PR review gate

File:

```text
.github/workflows/pr-review-gate.yml
```

Required status check:

```text
required-review-gate
```

Gate rules:

```text
- PR title must follow [TASK-ID] Short description.
- PR body must include Primary Ref and Task Ref.
- PR body must declare Generated Code: regenerated when generated paths are touched.
- Credential.txt and real env files must not be committed.
```

### 13.6. Auto-merge workflow

File:

```text
.github/workflows/auto-merge.yml
```

Auto-merge rule:

```text
- Non-draft same-repo PRs targeting main or develop get auto-merge enabled.
- Add label no-auto-merge when manual merge is required.
- Auto-merge must not bypass required CI or required human review.
- Auto-merge may delete short-lived feature/fix/chore branches after merge, but must never delete long-lived main or develop branches.
```

---

## 14. Docs workspace standard

### 14.1. Docs folder

```text
docs/
```

Docs phải được commit cùng repo.

Lý do:

```text
- Dev cần tra requirement ngay trong repo.
- PR cần ref tới tài liệu.
- Vendor không được code theo trí nhớ.
- Change log dễ trace.
```

---

### 14.2. Docs index

File index chính:

```text
docs/32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md
```

Mọi developer phải đọc file này trước khi đọc các file khác.

---

### 14.3. PR reference rule

Mỗi PR phải ghi:

```text
Primary Ref: docs/<file-name>.md
Task Ref: docs/37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md#<task-id>
```

Ví dụ:

```text
Primary Ref: docs/16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md
Task Ref: T-SHIPPING-003 Carrier manifest scan endpoint
```

---

## 15. Workspace naming convention

### 15.1. Folder names

Dùng lowercase kebab-case cho frontend/module folder nếu cần:

```text
master-data
warehouse-daily-board
```

Dùng lowercase cho Go module folder:

```text
masterdata
inventory
shipping
returns
```

Không dùng:

```text
MasterData
masterData
master_data
```

Trừ database column/table thì dùng snake_case theo file 17.

---

### 15.2. File names Go

Go file dùng snake_case:

```text
stock_movement.go
reserve_stock.go
carrier_manifest.go
```

Test file:

```text
reserve_stock_test.go
carrier_manifest_test.go
```

---

### 15.3. File names TypeScript/React

Component dùng PascalCase:

```text
WarehouseDailyBoard.tsx
CarrierManifestScanPanel.tsx
ReturnInspectionForm.tsx
```

Hooks dùng camelCase:

```text
useWarehouseDailyBoard.ts
useCarrierManifestScan.ts
```

Schemas:

```text
salesOrderSchema.ts
returnInspectionSchema.ts
```

---

## 16. Workspace dependency rules

### 16.1. Backend dependency direction

```text
handler → application → domain
application → repository interface
repository implementation → database
```

Không được:

```text
domain import handler
domain import repository implementation
domain import database package
repository gọi handler
```

---

### 16.2. Frontend dependency direction

```text
app route → module page → module components/hooks/services → shared components/api/utils
```

Không được:

```text
shared import module
module A import component nội bộ module B
```

Nếu cần dùng chung, đưa lên `shared` sau khi review.

---

### 16.3. Cross-module communication backend

Có 3 cách hợp lệ:

```text
1. Application service contract
2. Domain event/outbox
3. Read model/reporting query được kiểm soát
```

Không hợp lệ:

```text
Module A update trực tiếp table module B.
Module A import repository implementation của module B.
```

---

## 17. Required workspace for Phase 1 workflows

Do workflow thực tế có kho, bàn giao ĐVVC, hàng hoàn và gia công ngoài, workspace phải có đủ module/folder tương ứng ngay từ đầu.

### 17.1. Warehouse daily board

Backend:

```text
apps/api/internal/modules/inventory/application/get_warehouse_daily_board.go
apps/api/internal/modules/inventory/queries/warehouse_daily_board.sql
```

Frontend:

```text
apps/web/src/modules/inventory/components/WarehouseDailyBoard.tsx
apps/web/src/modules/inventory/hooks/useWarehouseDailyBoard.ts
```

Mock:

```text
tools/mock-data/warehouse-daily-board.json
```

---

### 17.2. Shift closing / end-of-day reconciliation

Backend:

```text
apps/api/internal/modules/inventory/application/close_warehouse_shift.go
apps/api/internal/modules/inventory/domain/warehouse_shift.go
```

Frontend:

```text
apps/web/src/modules/inventory/components/ShiftClosingForm.tsx
```

---

### 17.3. Carrier manifest / scan handover

Backend:

```text
apps/api/internal/modules/shipping/domain/carrier_manifest.go
apps/api/internal/modules/shipping/application/scan_manifest_order.go
apps/api/internal/modules/shipping/application/confirm_handover.go
```

Frontend:

```text
apps/web/src/modules/shipping/components/CarrierManifestScanPanel.tsx
apps/web/src/modules/shipping/components/HandoverSummary.tsx
```

Mock:

```text
tools/mock-data/carrier-manifest.json
```

---

### 17.4. Return receiving / return inspection

Backend:

```text
apps/api/internal/modules/returns/domain/return_receipt.go
apps/api/internal/modules/returns/domain/return_disposition.go
apps/api/internal/modules/returns/application/inspect_return_item.go
```

Frontend:

```text
apps/web/src/modules/returns/components/ReturnReceivingScanPanel.tsx
apps/web/src/modules/returns/components/ReturnInspectionForm.tsx
```

Mock:

```text
tools/mock-data/return-receiving.json
```

---

### 17.5. Subcontract manufacturing

Backend:

```text
apps/api/internal/modules/production/domain/subcontract_order.go
apps/api/internal/modules/production/domain/sample_approval.go
apps/api/internal/modules/production/application/issue_material_to_factory.go
apps/api/internal/modules/production/application/approve_factory_sample.go
apps/api/internal/modules/production/application/receive_subcontract_goods.go
apps/api/internal/modules/production/application/create_factory_claim.go
```

Frontend:

```text
apps/web/src/modules/production/components/SubcontractOrderForm.tsx
apps/web/src/modules/production/components/SampleApprovalPanel.tsx
apps/web/src/modules/production/components/FactoryClaimPanel.tsx
```

Mock:

```text
tools/mock-data/subcontract-order.json
```

---

## 18. Generated code rule

### 18.1. Generated folders

Generated code chỉ được nằm trong:

```text
apps/web/src/shared/api/generated
apps/api/internal/shared/openapi/generated
packages/openapi/generated
```

### 18.2. Generated code rule

```text
- Không sửa generated code bằng tay.
- Muốn sửa thì sửa OpenAPI source.
- Generated code phải được tạo lại bằng Makefile command.
- PR có sửa OpenAPI phải include generated diff nếu policy yêu cầu.
```

---

## 19. Git branch and PR workspace rule

### 19.1. Branch naming

```text
feature/<task-id>-short-name
fix/<task-id>-short-name
hotfix/<incident-id>-short-name
chore/<short-name>
```

Ví dụ:

```text
feature/T-SHIPPING-003-carrier-manifest-scan
feature/T-INVENTORY-002-stock-ledger
```

---

### 19.2. PR title

```text
[TASK-ID] Short description
```

Ví dụ:

```text
[T-SHIPPING-003] Implement carrier manifest scan endpoint
```

---

### 19.3. PR checklist bắt buộc

```text
- [ ] Ref đúng tài liệu chính
- [ ] Ref task trong file 37
- [ ] Automated PR review gate passes
- [ ] Không sửa generated code bằng tay
- [ ] Không vi phạm module boundary
- [ ] Có unit/integration test phù hợp
- [ ] Có audit log nếu là sensitive action
- [ ] Có permission check nếu là protected action
- [ ] Có migration nếu thay đổi DB
- [ ] Có OpenAPI update nếu thay đổi API
- [ ] Có UI state loading/error/empty nếu thay đổi frontend
```

### 19.4. Auto-review and merge rule

Every PR into `develop` or `main` must pass:

```text
- required-api
- required-web
- required-openapi
- required-migration
- required-review-gate
```

Auto-merge is enabled by default for non-draft same-repo PRs. It merges only after all required status checks pass and branch protection review requirements are satisfied. Use the `no-auto-merge` label when a PR needs manual merge control.

When promoting `develop` to `main`, use a merge commit and do not delete the `develop` branch after merge. `main` and `develop` are long-lived branches and must always remain available on remote.

---

## 20. Local development setup standard

### 20.1. Minimum tools

Developer cần cài:

```text
Go
Node.js LTS
pnpm
Docker Desktop hoặc Docker Engine
Make
Git
PostgreSQL client optional
```

---

### 20.2. Local startup command

```bash
make local-up
make migrate-up
make seed-local
make api-dev
make worker-dev
make web-dev
```

Developer không được tự chạy command khác nếu Makefile đã có command chuẩn.

---

### 20.3. Local service URLs

```text
Web:        http://localhost:3000
API:        http://localhost:8080/api/v1
PostgreSQL: localhost:5432
Redis:      localhost:6379
MinIO:      http://localhost:9000
Mailhog:    http://localhost:8025
```

---

## 21. Workspace guardrails

### 21.1. Không tạo folder tùy hứng

Trước khi tạo folder mới, phải trả lời:

```text
- Folder này thuộc app nào?
- Nó là module hay shared?
- Ai owner?
- Nó có boundary không?
- Có folder tương tự đang tồn tại không?
```

Nếu không trả lời được, không tạo.

---

### 21.2. Không tạo package tên chung

Cấm hoặc hạn chế mạnh:

```text
common
helpers
utils
misc
global
base
```

Nếu cần dùng `utils`, chỉ đặt trong phạm vi nhỏ và có chức năng rõ.

Ví dụ đúng:

```text
shared/utils/date.ts
shared/utils/money.ts
```

Ví dụ sai:

```text
shared/utils/doEverything.ts
```

---

### 21.3. Không duplicate workflow logic

Các workflow như:

```text
- reserve stock
- QC pass/fail
- carrier handover
- return disposition
- subcontract receipt
```

chỉ được implement trong application service owner.

Frontend không tự quyết định nghiệp vụ lõi.

Frontend chỉ:

```text
- hiển thị state
- validate cơ bản
- gọi API action
- hiển thị kết quả/lỗi
```

Backend mới là nơi quyết định nghiệp vụ.

---

## 22. Workspace source of truth matrix

| Chủ đề | Source of Truth trong workspace |
|---|---|
| Business scope | `docs/03_ERP_PRD_SRS_Phase1_My_Pham_v1.md` + `docs/33_...v1_1_Update_Pack...md` |
| Process flow | `docs/06_ERP_Process_Flow_ToBe...md` + `docs/20_ERP_Current_Workflow_AsIs...md` + `docs/21_ERP_Gap_Analysis...md` |
| Permission/RBAC | `docs/04_ERP_Permission_Approval...md` + `docs/19_ERP_Security...md` |
| Backend architecture | `docs/11_ERP_Technical_Architecture_Go_Backend...md` |
| Go coding standard | `docs/12_ERP_Go_Coding_Standards...md` |
| Module boundary | `docs/13_ERP_Go_Module_Component_Design_Standards...md` |
| UI/UX design | `docs/14_ERP_UI_UX_Design_System_Standards...md` + `docs/39_ERP_UI_Template_Hetzner_Minimal_Style...md` |
| Frontend architecture | `docs/15_ERP_Frontend_Architecture...md` |
| API contract | `packages/openapi/openapi.yaml` + `docs/16_ERP_API_Contract...md` |
| DB schema | `apps/api/migrations` + `docs/17_ERP_Database_Schema...md` |
| DevOps | `infra/` + `docs/18_ERP_DevOps...md` |
| Test strategy | `docs/24_ERP_QA_Test_Strategy...md` |
| Sprint task | `docs/37_ERP_Coding_Task_Board...md` |
| Workspace structure | `docs/38_ERP_Workspace_Repository_Structure_Standards...md` |

---

## 23. Sprint 0 workspace acceptance criteria

Sprint 0 chỉ được coi là đạt khi repo có đủ:

```text
- Root workspace đúng cấu trúc
- apps/api chạy được Go API skeleton
- apps/api/cmd/worker chạy được worker skeleton
- apps/web chạy được Next.js skeleton
- packages/openapi có openapi.yaml validate được
- frontend generate được API client
- PostgreSQL local chạy được qua docker compose
- migration up/down chạy được
- auth/RBAC skeleton có API tối thiểu
- audit log skeleton có API/demo tối thiểu
- stock ledger prototype ghi được movement
- warehouse scan prototype gọi được API
- CI chạy tối thiểu api/web/openapi/migration
- docs/ được đưa vào repo
```

---

## 24. Definition of Done cho workspace setup

Một workspace setup task chỉ được Done khi:

```text
- Folder đúng cấu trúc trong file này
- README hướng dẫn chạy được
- Makefile command hoạt động
- CI liên quan chạy pass
- Không có generated code bị sửa tay
- Không tạo folder ngoài chuẩn
- Có .env.example nếu cần env mới
- Có Docker/local support nếu là service mới
- Có ref tới task trong file 37
```

---

## 25. Kết luận

Workspace Phase 1 chính thức dùng:

```text
Monorepo: erp-platform
Backend: apps/api – Go modular monolith
Worker: apps/api/cmd/worker – chung Go codebase
Frontend: apps/web – Next.js + TypeScript
OpenAPI: packages/openapi – source of truth
Database migration: apps/api/migrations
Infra: infra/
Tools/data: tools/
Docs: docs/
CI/CD: .github/workflows/
```

Câu chốt để đội dev nhớ:

```text
Đừng tạo folder để chứa code.
Hãy tạo boundary để chứa trách nhiệm.
```
