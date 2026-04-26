# 18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1

**Dự án:** Web ERP cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Phiên bản:** v1.0  
**Phạm vi:** Phase 1  
**Backend:** Go  
**Frontend:** React / Next.js + TypeScript  
**Database:** PostgreSQL  
**Architecture:** Modular Monolith  
**Mục tiêu tài liệu:** Chuẩn hóa môi trường, CI/CD, deployment, release, backup, monitoring, bảo mật và vận hành kỹ thuật cho ERP Phase 1.

---

## 1. Tư duy nền tảng

ERP không giống website marketing. ERP là hệ thống vận hành tiền, hàng, kho, batch, sản xuất, giao hàng, công nợ và dữ liệu quản trị. Vì vậy DevOps của ERP không được thiết kế theo kiểu “deploy cho chạy là xong”.

DevOps cho ERP phải bảo vệ 5 thứ:

1. **Dữ liệu đúng**: đặc biệt là tồn kho, batch, QC, đơn hàng, phiếu nhập/xuất, bàn giao ĐVVC, hàng hoàn và công nợ.
2. **Deploy an toàn**: không để phiên bản mới làm vỡ luồng kho/sales/sản xuất.
3. **Rollback có kiểm soát**: lỗi code có thể rollback, nhưng lỗi dữ liệu phải xử lý bằng quy trình rõ ràng.
4. **Quan sát được hệ thống**: biết lỗi xảy ra ở API nào, module nào, user nào, chứng từ nào.
5. **Sẵn sàng vận hành thật**: hỗ trợ giờ cao điểm, cuối ngày đối soát, kiểm kê, bàn giao vận chuyển, xử lý hàng hoàn và nhập hàng gia công.

Một câu chốt:

> CI/CD của ERP không chỉ để ship code nhanh. Nó để ship code **mà không làm lệch kho, lệch tiền, lệch batch, lệch quy trình**.

---

## 2. Phạm vi Phase 1

Tài liệu này áp dụng cho các module Phase 1:

- Master Data
- Purchase / Mua hàng
- QC / QA cơ bản
- Inventory / Kho
- Production / Subcontract Manufacturing / Gia công ngoài
- Sales Order
- Shipping / Bàn giao ĐVVC
- Returns / Hàng hoàn
- Finance cơ bản liên quan thu/chi/công nợ sơ bộ
- Dashboard vận hành cơ bản
- Auth / RBAC / Approval / Audit Log

Các workflow thực tế cần bảo vệ trong Phase 1:

- Kho tiếp nhận đơn trong ngày, thực hiện xuất/nhập, soạn hàng, đóng gói, tối ưu vị trí kho, kiểm kê cuối ngày, đối soát số liệu và kết thúc ca.
- Nội quy kho gồm nhập kho, xuất kho, đóng hàng và xử lý hàng hoàn.
- Bàn giao ĐVVC gồm phân chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, quét mã, xử lý đơn thiếu và ký xác nhận bàn giao.
- Sản xuất/gia công gồm đặt đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, duyệt mẫu, sản xuất hàng loạt, giao hàng về kho, QC nhận hàng, báo lỗi nhà máy trong 3–7 ngày và thanh toán lần cuối.

---

## 3. Tech stack DevOps đề xuất

### 3.1. Runtime / Application

```text
Frontend: Next.js + TypeScript
Backend: Go
Database: PostgreSQL
Cache: Redis
Queue/Event: NATS hoặc RabbitMQ
Object Storage: S3-compatible / MinIO
Reverse Proxy: Nginx / Caddy / Cloud Load Balancer
Container: Docker
CI/CD: GitHub Actions hoặc GitLab CI
Observability: Prometheus + Grafana hoặc cloud equivalent
Log: Loki / ELK / cloud logging
Error Tracking: Sentry hoặc equivalent
```

### 3.2. Nguyên tắc chọn công nghệ

- Ưu tiên công nghệ ổn định, dễ vận hành, dễ tuyển người.
- Không over-engineer Phase 1 bằng microservices hoặc Kubernetes phức tạp nếu chưa cần.
- Backend Go chạy dạng modular monolith.
- Database PostgreSQL là nguồn sự thật chính.
- Stock ledger và audit log là dữ liệu nhạy cảm, không được xử lý tùy tiện.
- Queue dùng cho tác vụ nền, không dùng để che giấu lỗi transaction chính.

---

## 4. Mô hình môi trường

ERP nên có tối thiểu 5 môi trường.

```text
Local → Dev → Staging → UAT → Production
```

Có thể thêm DR nếu quy mô vận hành yêu cầu.

---

## 5. Local Environment

### 5.1. Mục tiêu

Local là môi trường để developer chạy đầy đủ module cơ bản trên máy cá nhân.

Local phải chạy được:

- Backend API Go
- Frontend Next.js
- PostgreSQL
- Redis
- Queue
- MinIO
- Migration
- Seed data cơ bản

### 5.2. Yêu cầu

Sử dụng Docker Compose cho local.

```text
docker-compose.local.yml
```

Local không được phụ thuộc vào database production/staging.

### 5.3. Seed data local

Seed data cần có:

- 3 users: admin, warehouse_user, sales_user
- 2 kho: warehouse_main, warehouse_return
- 5 SKU mẫu
- 5 nguyên vật liệu mẫu
- 3 batch mẫu
- 2 nhà cung cấp
- 2 ĐVVC
- 2 đơn bán mẫu
- 1 PO mẫu
- 1 subcontract order mẫu
- 1 return case mẫu

### 5.4. Local reset

Phải có lệnh reset local:

```bash
make local-reset
```

Lệnh này:

1. Drop database local.
2. Recreate database.
3. Run migrations.
4. Seed data.
5. In ra account test.

---

## 6. Development Environment

### 6.1. Mục tiêu

Dev environment dùng để team dev test tích hợp nhanh giữa frontend, backend, database, queue và storage.

### 6.2. Quy tắc

- Có thể deploy nhiều lần/ngày.
- Không dùng dữ liệu thật của khách hàng nếu chưa được masking.
- Không gửi notification thật cho khách hàng/KOL/NCC/ĐVVC.
- Cho phép reset data định kỳ.
- Dùng domain riêng, ví dụ:

```text
dev-erp.company.internal
```

### 6.3. Deploy lên dev

Tự động deploy khi merge vào branch:

```text
develop
```

Hoặc khi tag dev release:

```text
dev-YYYYMMDD-N
```

---

## 7. Staging Environment

### 7.1. Mục tiêu

Staging là môi trường gần production nhất về cấu hình và quy trình.

Dùng để:

- Test release candidate.
- Test regression.
- Test migration.
- Test performance cơ bản.
- Test tích hợp vận chuyển/sàn/website nếu có.

### 7.2. Quy tắc

- Không deploy trực tiếp từ branch cá nhân.
- Chỉ deploy từ release branch hoặc release candidate tag.
- Dữ liệu staging có thể lấy từ production đã masking hoặc data UAT giả lập.
- Không reset staging tùy tiện khi đang UAT.

### 7.3. Domain gợi ý

```text
staging-erp.company.vn
```

---

## 8. UAT Environment

### 8.1. Mục tiêu

UAT là môi trường cho business user nghiệm thu.

UAT phải mô phỏng đúng các luồng:

- nhập kho
- QC đầu vào
- xuất kho
- soạn/đóng hàng
- bàn giao ĐVVC
- hàng hoàn
- nhập hàng gia công
- QC thành phẩm
- đơn bán
- đối soát cuối ngày

### 8.2. Quy tắc

- UAT không phải playground cho dev.
- Chỉ deploy bản đã qua staging smoke test.
- User UAT phải có quyền giống thực tế.
- Không thay đổi dữ liệu UAT giữa chừng nếu đang chạy kịch bản nghiệm thu, trừ khi có yêu cầu reset rõ.

### 8.3. UAT release gate

Một bản chỉ được đẩy lên UAT nếu:

- Backend test pass.
- Frontend build pass.
- Migration chạy thành công ở staging.
- Smoke test pass.
- OpenAPI contract không broken.
- Không có critical security issue.

---

## 9. Production Environment

### 9.1. Mục tiêu

Production là môi trường thật, chứa dữ liệu thật.

Production phải ưu tiên:

- ổn định
- bảo mật
- backup
- monitoring
- rollback
- kiểm soát quyền truy cập

### 9.2. Quy tắc production

- Không SSH tùy tiện.
- Không sửa database trực tiếp nếu không có change ticket.
- Không chạy migration thủ công ngoài pipeline, trừ emergency đã được duyệt.
- Không dùng account admin chung.
- Không deploy trong giờ cao điểm nếu không cần thiết.
- Không deploy gần thời điểm kho đang đóng ca/đối soát cuối ngày.

### 9.3. Production release window gợi ý

Tùy vận hành thực tế, nhưng gợi ý:

```text
Khung deploy thường: 22:00 - 01:00
Không deploy: trước/sau giờ kho bàn giao ĐVVC, lúc kiểm kê cuối ca, lúc sale peak.
```

Nếu công ty vận hành nhiều ca, cần thống nhất maintenance window riêng.

---

## 10. DR / Disaster Recovery Environment

### 10.1. Có cần DR ngay Phase 1 không?

Nếu công ty còn quy mô nhỏ/vừa, Phase 1 có thể chưa cần DR realtime. Nhưng phải có backup/restore nghiêm túc.

Tối thiểu:

- daily database backup
- point-in-time recovery nếu có điều kiện
- file storage backup
- restore test định kỳ

### 10.2. RPO / RTO gợi ý

```text
RPO mục tiêu Phase 1: <= 24h
RTO mục tiêu Phase 1: <= 8h
RPO nâng cao: <= 1h
RTO nâng cao: <= 2h
```

RPO là lượng dữ liệu tối đa có thể mất khi sự cố. RTO là thời gian khôi phục dịch vụ.

---

## 11. Repository Strategy

### 11.1. Mô hình repo khuyến nghị

Khuyến nghị dùng monorepo hoặc multi-repo có cấu trúc rõ.

#### Option A: Monorepo

```text
erp-platform/
  apps/
    web/
    api/
    worker/
  packages/
    openapi/
    shared-types/
  infra/
  docs/
```

Ưu điểm:

- dễ đồng bộ version API/frontend/backend
- dễ chạy CI toàn hệ thống
- phù hợp team nhỏ/vừa

#### Option B: Multi-repo

```text
erp-api-go
erp-web-nextjs
erp-infra
erp-docs
```

Ưu điểm:

- boundary rõ hơn
- phù hợp team lớn hơn

### 11.2. Chốt Phase 1

Nếu team nhỏ/vừa, nên chọn:

```text
Monorepo
```

Vì ERP Phase 1 cần tốc độ phối hợp cao giữa backend, frontend và tài liệu API.

---

## 12. Branching Strategy

### 12.1. Branch chính

```text
main        → production-ready
develop     → integration/dev
release/*   → release candidate
feature/*   → task/feature
hotfix/*    → sửa lỗi production khẩn cấp
```

### 12.2. Quy tắc

- Không commit trực tiếp vào `main`.
- Không commit trực tiếp vào `release/*`, trừ fix đã review.
- Mọi thay đổi phải qua pull request / merge request.
- PR phải có ít nhất 1 reviewer.
- PR ảnh hưởng module critical cần tech lead review.

### 12.3. Mapping deploy

```text
feature/*  → không auto deploy hoặc deploy preview
 develop   → dev environment
release/*  → staging/UAT
main       → production
hotfix/*   → staging smoke test → production
```

---

## 13. Commit Convention

Dùng Conventional Commits.

```text
feat: add shipment handover scan endpoint
fix: prevent shipping handover for unpacked orders
chore: update docker base image
refactor: extract inventory reservation policy
test: add return inspection test cases
docs: update API contract for QC release
```

### 13.1. Scope gợi ý

```text
feat(inventory): add immutable stock movement
fix(shipping): prevent duplicate scan in manifest
feat(returns): add return disposition flow
feat(subcontract): add sample approval status
```

### 13.2. Commit không chấp nhận

```text
update
fix bug
abc
final
new code
```

---

## 14. Versioning Strategy

### 14.1. App version

Dùng semantic versioning theo release:

```text
v1.0.0
v1.0.1
v1.1.0
```

### 14.2. Release candidate

```text
v1.0.0-rc.1
v1.0.0-rc.2
```

### 14.3. Docker image tag

Mỗi image phải có:

```text
app:v1.0.0
app:git-<short-sha>
app:release-YYYYMMDD
```

Không deploy production bằng tag `latest`.

---

## 15. CI Pipeline Tổng Quan

CI phải chạy khi:

- push vào feature branch
- mở PR/MR
- merge vào develop
- tạo release branch
- tạo tag release

### 15.1. Pipeline stages

```text
1. Checkout
2. Install dependencies
3. Lint
4. Format check
5. Unit test
6. Integration test
7. Build
8. OpenAPI validation
9. Security scan
10. Docker image build
11. Push artifact/image
12. Deploy theo environment
13. Smoke test
```

---

## 16. Backend CI Standards - Go

### 16.1. Backend checks

Backend CI phải chạy:

```bash
go fmt ./...
go vet ./...
go test ./...
golangci-lint run
```

### 16.2. Test coverage

Tối thiểu Phase 1:

```text
Domain/policy critical: >= 80%
Application service critical: >= 70%
Repository/integration: theo module critical
Overall: không ép số ảo, ưu tiên test đúng chỗ nguy hiểm
```

Critical business rules cần test:

- không xuất batch QC hold/fail
- không bàn giao ĐVVC nếu đơn chưa packed
- không scan trùng đơn trong manifest
- không reserve vượt available stock
- hàng hoàn phải có disposition
- stock ledger không sửa trực tiếp
- subcontract order không nhận hàng nếu chưa có nghiệm thu/QC phù hợp

### 16.3. Integration test

Dùng PostgreSQL container trong CI.

Các test integration nên chạy:

- migration up
- seed minimal
- tạo PO → receiving → QC → stock movement
- tạo sales order → reserve → pick → pack → handover
- return receipt → inspect → disposition
- subcontract order → material transfer → receive finished goods

---

## 17. Frontend CI Standards - Next.js

### 17.1. Frontend checks

```bash
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

### 17.2. UI critical tests

Các màn hình cần test smoke:

- login
- dashboard warehouse daily board
- purchase receiving
- QC release
- stock movement list
- sales order detail
- pick/pack screen
- shipping handover scan
- return inspection
- subcontract order detail

### 17.3. API client generation

Frontend phải dùng API client generate từ OpenAPI.

Pipeline phải validate:

- OpenAPI file hợp lệ
- generated client không lỗi type
- frontend build pass sau khi generate

---

## 18. OpenAPI CI Standards

### 18.1. Validate OpenAPI

Mỗi PR thay đổi API phải validate:

```text
- schema hợp lệ
- không thiếu response envelope
- error code có trong danh mục
- không phá contract cũ nếu chưa versioning
```

### 18.2. Breaking change

Breaking change gồm:

- đổi tên field
- xóa field
- đổi type
- đổi enum
- đổi path
- đổi response envelope
- đổi behavior của status/action endpoint

Breaking change phải:

1. Có ghi chú trong PR.
2. Có approval từ tech lead.
3. Có update frontend/client.
4. Có update UAT case nếu ảnh hưởng nghiệp vụ.

---

## 19. Database Migration Standards

### 19.1. Tool

Dùng:

```text
golang-migrate
```

hoặc tool migration tương đương đã được tech lead chốt.

### 19.2. Migration naming

```text
YYYYMMDDHHMMSS_create_inventory_stock_movements.up.sql
YYYYMMDDHHMMSS_create_inventory_stock_movements.down.sql
```

### 19.3. Quy tắc migration

- Không sửa file migration đã chạy ở shared environment.
- Migration phải chạy được từ database rỗng.
- Migration phải idempotent ở mức pipeline nếu cần.
- Không drop column/table production nếu chưa có deprecation plan.
- Không đổi enum/state critical mà không có migration dữ liệu rõ.
- Migration ảnh hưởng stock ledger, audit log, order, QC phải review kỹ.

### 19.4. Migration gate

Trước khi deploy staging/prod:

```text
1. Backup database
2. Run migration dry-run nếu tool hỗ trợ
3. Run migration trên staging
4. Run smoke test
5. Chỉ sau đó mới cho production
```

### 19.5. Rollback database

Không tin vào `down.sql` cho mọi tình huống production.

Production rollback phải phân loại:

```text
Code rollback: rollback image
Schema rollback: chỉ dùng nếu an toàn
Data rollback: phải có plan riêng, backup/restore hoặc corrective migration
```

Với ERP, nhiều khi không rollback dữ liệu được vì đã có giao dịch thật. Khi đó phải dùng **corrective migration** hoặc **compensating transaction**.

---

## 20. Docker Standards

### 20.1. Backend Dockerfile

Nguyên tắc:

- multi-stage build
- build binary Go tối ưu
- image runtime nhỏ
- không chạy bằng root user nếu có thể
- không chứa secret
- healthcheck rõ

Ví dụ structure:

```text
Dockerfile.api
Dockerfile.worker
Dockerfile.web
```

### 20.2. Image build rule

Mỗi image tag theo:

```text
<service>:<version>
<service>:<git-sha>
```

Không dùng `latest` cho production.

### 20.3. Container healthcheck

Backend API phải có endpoint:

```text
GET /healthz
GET /readyz
```

`/healthz`: process còn sống.  
`/readyz`: kết nối DB/Redis/queue cần thiết sẵn sàng.

---

## 21. Deployment Architecture Phase 1

### 21.1. Option khuyến nghị Phase 1

Nếu chưa cần scale lớn:

```text
Docker Compose / Docker Swarm / lightweight VM deployment
```

Nếu công ty có team infra tốt:

```text
Kubernetes managed service
```

### 21.2. Đề xuất thực dụng

Phase 1 nên bắt đầu bằng:

```text
Docker Compose trên VM/VPS/Cloud Instance + managed PostgreSQL nếu có
```

Lý do:

- ít phức tạp
- dễ vận hành
- đủ cho ERP nội bộ giai đoạn đầu
- sau này có thể nâng lên Kubernetes

### 21.3. Production services

```text
reverse-proxy
frontend-web
backend-api
worker
postgresql
redis
queue
minio/s3
monitoring
logging
```

Nếu dùng managed DB/S3, không chạy DB/storage trong cùng VM production.

---

## 22. CD Pipeline Tổng Quan

### 22.1. Deploy Dev

Trigger:

```text
merge vào develop
```

Steps:

1. Build images.
2. Push image registry.
3. Deploy dev.
4. Run migrations.
5. Run smoke test.
6. Notify team.

### 22.2. Deploy Staging

Trigger:

```text
release branch hoặc rc tag
```

Steps:

1. Build release images.
2. Backup staging database.
3. Run migrations.
4. Deploy staging.
5. Run smoke test.
6. Run regression critical flow.
7. Publish release notes.

### 22.3. Deploy UAT

Trigger:

```text
manual approval từ PM/Tech Lead
```

Steps:

1. Deploy release candidate.
2. Freeze UAT data nếu đang nghiệm thu.
3. Run smoke test.
4. Notify business users.
5. Ghi version UAT.

### 22.4. Deploy Production

Trigger:

```text
manual approval từ Tech Lead + Product Owner/CEO delegate
```

Steps:

1. Confirm release note.
2. Confirm downtime/maintenance window nếu có.
3. Backup database.
4. Backup file storage metadata nếu cần.
5. Run migration.
6. Deploy backend/worker/frontend.
7. Run smoke test.
8. Monitor logs/metrics 30–60 phút.
9. Confirm go-live success.

---

## 23. Production Deployment Checklist

Trước deploy production:

```text
[ ] Release notes đã có
[ ] PR đã merged đúng branch
[ ] CI pass
[ ] Security scan pass hoặc waiver đã duyệt
[ ] Migration đã test ở staging
[ ] Smoke test staging pass
[ ] UAT sign-off nếu là release lớn
[ ] Backup database hoàn tất
[ ] Rollback plan rõ
[ ] Người trực deploy đã phân công
[ ] Không trùng giờ cao điểm kho/sale/đối soát
```

Sau deploy:

```text
[ ] /healthz pass
[ ] /readyz pass
[ ] Login pass
[ ] Dashboard load pass
[ ] Create test sales order pass nếu được phép
[ ] Stock movement list load pass
[ ] Shipping scan page load pass
[ ] QC status page load pass
[ ] Worker running
[ ] Outbox queue không backlog bất thường
[ ] Error rate không tăng
[ ] Business owner xác nhận luồng chính ổn
```

---

## 24. Smoke Test Standards

Smoke test production/staging phải ngắn nhưng đánh đúng điểm nguy hiểm.

### 24.1. Smoke test kỹ thuật

```text
- API health
- DB connection
- Redis connection
- Queue connection
- File storage connection
- Frontend load
- Auth login
- RBAC permission check
```

### 24.2. Smoke test nghiệp vụ Phase 1

```text
1. Xem danh sách SKU
2. Xem tồn kho theo batch
3. Tạo phiếu nhập nháp
4. Xem màn QC hold/pass/fail
5. Xem sales order list
6. Xem pick/pack screen
7. Mở màn shipping handover
8. Scan thử bằng test code ở staging/UAT
9. Mở return inspection
10. Mở subcontract order
```

Production không nên tạo giao dịch test thật nếu không có sandbox flag. Nếu cần test thật, phải có chứng từ test được hủy/đánh dấu rõ.

---

## 25. Rollback Strategy

### 25.1. Code rollback

Rollback code bằng cách deploy image version trước.

```text
app:v1.0.1 → rollback app:v1.0.0
```

### 25.2. Database rollback

Database rollback phải cực kỳ cẩn trọng.

Không được tự động rollback nếu:

- đã tạo stock movement thật
- đã thay đổi QC status
- đã tạo shipment handover
- đã tạo return disposition
- đã ghi nhận công nợ/thanh toán

### 25.3. Safe rollback pattern

Ưu tiên:

```text
1. Rollback code
2. Disable feature flag
3. Stop worker liên quan
4. Apply corrective migration nếu cần
5. Không sửa tay data critical
```

### 25.4. Feature flag

Các tính năng có rủi ro nên có feature flag:

- shipping scan flow mới
- return inspection flow mới
- stock reservation logic mới
- subcontract manufacturing enhancement
- pricing/discount rule mới

---

## 26. Feature Flag Standards

### 26.1. Khi nào dùng feature flag

Dùng khi:

- release tính năng lớn nhưng muốn bật dần
- cần rollback nhanh mà không rollback toàn app
- có logic mới ảnh hưởng kho/sales/shipping
- muốn test với một nhóm user nhỏ

### 26.2. Quy tắc

- Feature flag phải có owner.
- Phải có ngày review/cleanup.
- Không để flag chết tồn tại mãi.
- Không dùng flag thay cho permission.

---

## 27. Secret Management

### 27.1. Không được làm

Không commit:

- database password
- JWT secret
- S3 key
- SMTP credential
- payment/shipping API key
- production `.env`

### 27.2. Nên làm

Secrets lưu trong:

```text
GitHub Actions Secrets / GitLab Variables
Cloud Secret Manager
Vault
Doppler/1Password Secrets Automation
```

### 27.3. Environment variables

Naming convention:

```text
APP_ENV=production
DATABASE_URL=...
REDIS_URL=...
JWT_SECRET=...
S3_BUCKET=...
S3_ENDPOINT=...
QUEUE_URL=...
```

### 27.4. Rotation

Secrets cần rotate khi:

- nhân sự kỹ thuật nghỉ việc
- nghi ngờ lộ key
- vendor đổi người
- định kỳ theo policy

---

## 28. Access Control DevOps

### 28.1. Role kỹ thuật

```text
Developer
Senior Developer
Tech Lead
DevOps Engineer
QA Engineer
Product Owner
Business UAT User
Read-only Auditor
```

### 28.2. Quy tắc

- Developer không có quyền production DB write.
- QA không có quyền deploy production.
- Product Owner không có quyền sửa infra.
- DevOps có quyền infra nhưng không sửa business data nếu không có ticket.
- Production deploy cần ít nhất 2 người biết.

### 28.3. Production DB access

Chỉ cho phép:

- read-only cho kiểm tra
- write access tạm thời qua break-glass
- mọi query write phải log/ticket

---

## 29. Logging Standards

### 29.1. Log phải có

Mỗi request log:

```text
request_id
user_id
role
module
method
path
status_code
latency_ms
client_ip
user_agent
```

Business action log thêm:

```text
document_type
document_id
action
old_status
new_status
warehouse_id
sku_id/batch_id nếu có
```

### 29.2. Không log

Không log:

- password
- token
- secret
- full personal data không cần thiết
- payment sensitive data

### 29.3. Log levels

```text
DEBUG: local/dev only
INFO: action bình thường
WARN: bất thường nhưng chưa lỗi hệ thống
ERROR: lỗi cần xử lý
CRITICAL: sự cố ảnh hưởng vận hành
```

### 29.4. Log retention

Gợi ý:

```text
Application log: 30–90 ngày
Audit log nghiệp vụ: >= 3 năm hoặc theo quy định công ty
Security log: >= 1 năm
```

---

## 30. Audit Log Standards

Audit log là bắt buộc với ERP.

### 30.1. Action phải audit

- tạo/sửa/hủy PO
- nhận hàng
- QC pass/fail/hold
- tạo/sửa/hủy stock adjustment
- reserve stock
- issue stock
- shipment handover
- return disposition
- update batch/expiry
- update price/discount
- approve/reject document
- update user permission
- production/subcontract order status change

### 30.2. Audit data

```text
actor_id
action
module
entity_type
entity_id
before_snapshot
after_snapshot
reason
request_id
created_at
```

### 30.3. Audit không được sửa

Audit log production không cho sửa/xóa bởi user thường.

Nếu cần xóa do pháp lý/privacy, phải có quy trình riêng.

---

## 31. Monitoring & Metrics

### 31.1. System metrics

Theo dõi:

- CPU
- memory
- disk
- network
- DB connections
- Redis memory
- queue backlog
- API latency
- API error rate

### 31.2. Business critical metrics

Theo dõi riêng:

```text
stock_reservation_failures
qc_hold_batches_count
shipping_scan_failures
manifest_mismatch_count
return_pending_inspection_count
outbox_pending_count
worker_failed_jobs
end_of_day_reconciliation_mismatch_count
subcontract_receiving_pending_qc_count
```

### 31.3. API SLO gợi ý

```text
P95 API latency: < 500ms cho list/detail cơ bản
P95 scan API latency: < 300ms trong LAN/ổn định
Error rate: < 1% ngoài lỗi user input
```

### 31.4. Dashboard kỹ thuật

Cần có dashboard:

- API health
- DB health
- worker/queue health
- error tracking
- release/version status
- business operational exceptions

---

## 32. Alerting Standards

### 32.1. Alert critical

Alert ngay khi:

- API down
- DB down
- disk sắp đầy
- migration failed
- queue backlog tăng bất thường
- worker stopped
- error rate tăng mạnh
- shipping scan fail liên tục
- stock reservation fail bất thường
- end-of-day reconciliation mismatch lớn

### 32.2. Alert routing

```text
Critical: DevOps + Tech Lead + PM
Business critical: Tech Lead + Module Owner + Business Owner
Security: Tech Lead + Admin/CEO delegate
```

### 32.3. Alert không được spam

Alert phải có threshold và grouping. Nếu alert quá nhiều, người trực sẽ bỏ qua.

---

## 33. Background Jobs / Worker Standards

### 33.1. Các job Phase 1

```text
- send notification
- process outbox events
- generate reports
- sync shipping status nếu có integration
- calculate inventory aging
- near-expiry alert
- daily warehouse closing summary
- cleanup temporary files
- export data async
```

### 33.2. Job rule

- Job phải idempotent.
- Job thất bại phải retry có giới hạn.
- Job critical phải có dead-letter queue hoặc failed_jobs table.
- Job không được âm thầm fail.
- Job không được update dữ liệu critical mà không audit.

### 33.3. Retry policy gợi ý

```text
Retry 1: sau 1 phút
Retry 2: sau 5 phút
Retry 3: sau 15 phút
Retry 4: sau 1 giờ
Sau đó đưa vào failed_jobs
```

---

## 34. Outbox Pattern Standards

### 34.1. Vì sao dùng outbox

ERP cần đảm bảo khi transaction chính thành công thì event cũng được ghi nhận.

Ví dụ:

```text
Sales order packed
→ update order status
→ create audit log
→ create outbox event OrderPacked
```

Sau đó worker mới publish event/send notification.

### 34.2. Quy tắc

- Outbox insert nằm cùng transaction với business action.
- Worker xử lý outbox async.
- Event phải có idempotency key.
- Event publish failed phải retry.

### 34.3. Event critical

```text
StockReserved
StockMovementCreated
QCStatusChanged
OrderPacked
ShipmentHandedOver
ReturnReceived
ReturnDispositionCompleted
SubcontractGoodsReceived
WarehouseShiftClosed
```

---

## 35. Backup Standards

### 35.1. Database backup

Tối thiểu:

```text
Daily full backup
Retention: 30 ngày
Monthly backup: 12 tháng nếu cần
```

Nếu có điều kiện:

```text
Point-in-time recovery
WAL archiving
```

### 35.2. File storage backup

Backup:

- chứng từ
- biên bản bàn giao
- hồ sơ QC
- hình ảnh hàng hoàn
- file đính kèm sản xuất/gia công
- hợp đồng/vendor documents

### 35.3. Backup encryption

Backup production phải mã hóa hoặc lưu trong vùng bảo mật.

### 35.4. Restore test

Không test restore thì backup chỉ là niềm tin.

Gợi ý:

```text
Mỗi tháng test restore 1 lần ở môi trường isolated.
```

---

## 36. Restore / Recovery Runbook

### 36.1. Khi cần restore

- database corruption
- deploy làm hỏng dữ liệu nghiêm trọng
- mất server
- ransomware/sự cố bảo mật
- xóa nhầm dữ liệu critical không thể corrective

### 36.2. Quy trình restore

```text
1. Dừng ghi dữ liệu nếu cần
2. Xác định thời điểm restore
3. Thông báo stakeholders
4. Restore DB vào môi trường tạm
5. Kiểm tra dữ liệu critical
6. Quyết định promote restored DB
7. Run smoke test
8. Mở lại hệ thống
9. Postmortem
```

### 36.3. Kiểm tra sau restore

Cần kiểm tra:

- stock ledger
- stock balance
- sales orders pending
- shipments handover
- returns pending
- QC hold/pass/fail
- subcontract orders
- user permissions

---

## 37. Security Baseline

### 37.1. Network

- Production dùng HTTPS.
- DB không public internet.
- Redis/queue không public internet.
- Admin endpoints có IP allowlist hoặc VPN nếu cần.

### 37.2. Application security

- Password hashing mạnh.
- Session/JWT có expiry.
- RBAC enforced ở backend, không chỉ frontend.
- Rate limit login/API nhạy cảm.
- CSRF protection nếu dùng cookie session.
- Input validation.
- File upload validation.

### 37.3. Dependency scanning

CI phải scan:

- Go dependencies
- npm dependencies
- Docker base image

### 37.4. Security headers

Frontend/reverse proxy cần:

```text
Content-Security-Policy
X-Frame-Options
X-Content-Type-Options
Strict-Transport-Security
Referrer-Policy
```

---

## 38. File Upload / Storage Ops

### 38.1. File types cần hỗ trợ

- PDF
- JPG/PNG/WebP
- XLSX/CSV có kiểm soát
- document attachments

### 38.2. File upload rule

- giới hạn size
- validate content type
- scan malware nếu có điều kiện
- lưu metadata vào DB
- không lưu file trực tiếp trong database
- file liên quan chứng từ phải có owner/entity

### 38.3. File critical

- hình ảnh hàng hoàn
- biên bản bàn giao ĐVVC
- hồ sơ QC
- tài liệu COA/MSDS
- biên bản giao NVL/bao bì cho nhà máy
- mẫu duyệt sản xuất

---

## 39. Data Privacy & Masking

### 39.1. Dữ liệu cần bảo vệ

- thông tin khách hàng
- số điện thoại
- địa chỉ giao hàng
- thông tin nhân viên
- thông tin tài chính
- cost/margin
- thông tin nhà cung cấp nhạy cảm

### 39.2. Masking staging/UAT

Nếu lấy data production sang staging/UAT:

- mask số điện thoại
- mask địa chỉ
- mask email nếu cần
- không sync secret/token
- không gửi notification thật

---

## 40. Performance Baseline

### 40.1. API critical

Các API phải nhanh:

- scan order for handover
- get stock availability
- reserve stock
- pick/pack confirmation
- QC release
- return scan
- sales order detail

### 40.2. Query optimization

- list API có pagination
- filter dùng index phù hợp
- không query N+1
- report nặng chạy async hoặc materialized view
- không để dashboard quét toàn database mỗi lần load

### 40.3. Load test gợi ý

Trước production:

```text
- 50 concurrent users
- 1000 order scan simulation
- 5000 stock movement records
- dashboard load under normal data volume
```

Tùy quy mô thực tế sẽ tăng.

---

## 41. Release Notes Standards

Mỗi release phải có release notes.

Format:

```text
Version: v1.0.0
Date:
Environment:

New Features:
- ...

Improvements:
- ...

Bug Fixes:
- ...

Database Changes:
- ...

Known Issues:
- ...

UAT Impact:
- ...

Rollback Plan:
- ...
```

Release notes phải dễ hiểu cho business owner, không chỉ cho dev.

---

## 42. Change Management

### 42.1. Change categories

```text
Minor change: UI text, small bug, no data impact
Normal change: feature/update có ảnh hưởng workflow
Major change: ảnh hưởng kho/tồn/QC/sales/finance
Emergency change: production incident/hotfix
```

### 42.2. Approval

```text
Minor: Tech Lead
Normal: Tech Lead + Product Owner
Major: Tech Lead + Product Owner + Business Owner
Emergency: Tech Lead + CEO/Delegate, post-approval nếu quá gấp
```

### 42.3. Change ticket phải có

- lý do thay đổi
- module ảnh hưởng
- dữ liệu ảnh hưởng
- test đã chạy
- rollback plan
- thời gian deploy

---

## 43. Incident Management

### 43.1. Severity

```text
SEV1: hệ thống down / không xuất được hàng / dữ liệu critical sai diện rộng
SEV2: module critical lỗi nhưng có workaround
SEV3: lỗi chức năng nhỏ, không chặn vận hành
SEV4: cosmetic/minor issue
```

### 43.2. Ví dụ SEV1

- Không login được toàn hệ thống.
- Không thể bàn giao ĐVVC trong giờ vận hành.
- Stock balance sai hàng loạt.
- QC hold batch nhưng vẫn bán được.
- Database down.

### 43.3. Incident flow

```text
1. Detect
2. Triage
3. Assign owner
4. Communicate impact
5. Mitigate
6. Resolve
7. Verify
8. Postmortem
```

### 43.4. Postmortem

Mỗi SEV1/SEV2 phải có postmortem:

- chuyện gì xảy ra
- ảnh hưởng gì
- nguyên nhân gốc
- vì sao không phát hiện sớm
- đã xử lý gì
- action item phòng ngừa

---

## 44. ERP-Specific Operational Safeguards

### 44.1. Không deploy lúc kho đóng ca

Vì workflow kho có kiểm kê và đối soát cuối ngày, không deploy vào khung:

```text
trước/sau thời điểm kiểm kê cuối ngày và đối soát số liệu
```

Trừ khi emergency.

### 44.2. Không deploy shipping change khi đang bàn giao ĐVVC

Bàn giao ĐVVC có quét mã, đối chiếu số lượng, xử lý đủ/chưa đủ đơn. Change ở màn/API này phải deploy ngoài giờ bàn giao chính.

### 44.3. Không deploy stock/QC change nếu chưa có UAT

Stock và QC ảnh hưởng trực tiếp tới việc hàng có được bán/xuất hay không. Mọi thay đổi cần UAT flow.

### 44.4. Không sửa production stock bằng SQL tay

Nếu tồn kho lệch:

- tạo stock adjustment có phê duyệt
- ghi audit log
- có lý do
- có attachment nếu cần

Không update thẳng bảng balance.

---

## 45. Infrastructure as Code

### 45.1. Phase 1

Nếu hạ tầng đơn giản, tối thiểu lưu cấu hình trong repo:

```text
infra/
  docker-compose/
  nginx/
  scripts/
  env.example
```

### 45.2. Nâng cao

Có thể dùng:

```text
Terraform
Ansible
Helm
```

### 45.3. Quy tắc

- Infra config phải version control.
- Không cấu hình production bằng tay mà không ghi lại.
- Thay đổi infra cần review.

---

## 46. Nginx / Reverse Proxy Standards

### 46.1. Nhiệm vụ

- terminate HTTPS
- route frontend/backend
- gzip/brotli nếu cần
- rate limit endpoint nhạy cảm
- access log
- security headers

### 46.2. Routes gợi ý

```text
/              → frontend
/api/v1/*      → backend api
/healthz       → backend health or proxy health
/storage/*     → signed file route nếu cần
```

### 46.3. Timeout

API thường:

```text
30s
```

Export/report async không nên giữ request quá lâu. Dùng job + download link.

---

## 47. Environment Variable Standard

### 47.1. Backend env

```text
APP_ENV
APP_NAME
APP_VERSION
HTTP_PORT
DATABASE_URL
REDIS_URL
QUEUE_URL
JWT_SECRET
ACCESS_TOKEN_TTL
REFRESH_TOKEN_TTL
S3_ENDPOINT
S3_BUCKET
S3_ACCESS_KEY
S3_SECRET_KEY
LOG_LEVEL
SENTRY_DSN
```

### 47.2. Frontend env

```text
NEXT_PUBLIC_APP_ENV
NEXT_PUBLIC_API_BASE_URL
NEXT_PUBLIC_APP_VERSION
NEXT_PUBLIC_SENTRY_DSN
```

### 47.3. Không public secret frontend

Biến bắt đầu bằng `NEXT_PUBLIC_` sẽ lộ ra client. Không đặt secret ở đó.

---

## 48. Makefile / Developer Commands

Nên có Makefile chuẩn.

```bash
make dev
make test
make lint
make fmt
make migrate-up
make migrate-down
make seed
make openapi-generate
make docker-build
make local-reset
```

Mục tiêu: dev mới vào dự án chạy được trong 1 ngày.

---

## 49. CI Example Flow

```text
Pull Request Opened
        ↓
Lint Backend
        ↓
Test Backend
        ↓
Lint Frontend
        ↓
Typecheck Frontend
        ↓
Build Frontend
        ↓
Validate OpenAPI
        ↓
Run DB migration test
        ↓
Build Docker Images
        ↓
Security Scan
        ↓
Ready for Review
```

---

## 50. CD Example Flow Production

```text
Create Release Tag
        ↓
Build Immutable Images
        ↓
Push Registry
        ↓
Manual Approval
        ↓
Backup DB
        ↓
Run Migration
        ↓
Deploy API/Worker
        ↓
Deploy Frontend
        ↓
Run Smoke Test
        ↓
Monitor 60 Minutes
        ↓
Close Release
```

---

## 51. Deployment Order

Khi deploy production:

```text
1. Backup database
2. Apply backward-compatible migrations
3. Deploy backend API
4. Deploy worker
5. Deploy frontend
6. Run smoke test
7. Enable feature flags if needed
```

Nếu có breaking API change, phải dùng chiến lược:

```text
expand → migrate → contract
```

Tức là thêm field/table trước, deploy code mới, migrate data, sau đó mới bỏ cái cũ ở release sau.

---

## 52. Testing Gates Theo Module Critical

### 52.1. Inventory

Trước release:

```text
[ ] stock movement insert pass
[ ] balance update pass
[ ] reserve stock pass
[ ] insufficient stock rejected
[ ] QC hold stock not available
[ ] expiry/batch filter pass
```

### 52.2. Shipping

```text
[ ] packed order can be scanned
[ ] un-packed order cannot be handed over
[ ] duplicate scan rejected
[ ] manifest count matched
[ ] missing order scenario handled
```

### 52.3. Returns

```text
[ ] return receipt created
[ ] return scanned
[ ] inspection required
[ ] reusable item moved to stock only after pass
[ ] unusable item moved to lab/damaged area
```

### 52.4. Subcontract Manufacturing

```text
[ ] subcontract order created
[ ] material/packaging transfer created
[ ] sample approval required
[ ] finished goods receiving created
[ ] QC receiving pass/fail
[ ] defect report within 3–7 days supported
```

---

## 53. Data Migration Deployment Tie-In

Khi có migration dữ liệu cũ sang ERP:

- migration data phải chạy ở staging/UAT trước
- kết quả phải đối chiếu với checklist
- không import production data bằng script chưa test
- script import phải log record lỗi
- import phải có dry-run mode nếu được

Các dữ liệu cần cẩn trọng:

- SKU
- batch
- stock balance
- khách hàng
- nhà cung cấp
- đơn hàng đang mở
- hàng hoàn đang xử lý
- công nợ mở

---

## 54. Production Support Model

### 54.1. Tuần đầu go-live

Nên có hypercare:

```text
Ngày 1–3: trực sát sao giờ hành chính + sau giờ đóng ca
Ngày 4–7: trực theo khung peak
Tuần 2–4: daily review lỗi và cải tiến
```

### 54.2. Người trực

- 1 tech lead/backend
- 1 frontend/dev
- 1 QA/BA
- 1 business key user kho
- 1 business key user sales/shipping
- 1 business key user finance/ops

### 54.3. Kênh support

- ticket system
- group khẩn cấp riêng cho go-live
- không xử lý yêu cầu thay đổi lớn qua chat không ticket

---

## 55. DoD - Definition of Done cho DevOps/Release

Một release chỉ được coi là xong khi:

```text
[ ] Code merged đúng branch
[ ] CI pass
[ ] Docker images built and tagged
[ ] Migration reviewed
[ ] Staging deploy pass
[ ] Smoke test pass
[ ] Release notes published
[ ] Rollback plan documented
[ ] Production backup completed
[ ] Production deploy pass
[ ] Post-deploy smoke test pass
[ ] Monitoring không có alert bất thường
[ ] Business owner xác nhận luồng chính ổn
```

---

## 56. Checklist triển khai ngay cho team

### 56.1. Việc cần làm trước khi code mạnh

```text
[ ] Chốt repo model
[ ] Tạo CI pipeline backend
[ ] Tạo CI pipeline frontend
[ ] Tạo Dockerfile backend/frontend/worker
[ ] Tạo docker-compose local
[ ] Tạo migration workflow
[ ] Tạo OpenAPI validation
[ ] Tạo dev environment
[ ] Tạo staging environment
[ ] Tạo backup script
[ ] Tạo logging baseline
[ ] Tạo monitoring dashboard cơ bản
```

### 56.2. Việc cần làm trước UAT

```text
[ ] UAT environment ready
[ ] Seed UAT data
[ ] User roles UAT ready
[ ] Release candidate deployed
[ ] Smoke test pass
[ ] Test data for warehouse/shipping/returns/subcontract ready
```

### 56.3. Việc cần làm trước go-live

```text
[ ] Production environment ready
[ ] Domain/HTTPS ready
[ ] Backup/restore tested
[ ] Migration rehearsal done
[ ] Cutover plan done
[ ] Support team assigned
[ ] Monitoring/alerting ready
[ ] Rollback plan ready
```

---

## 57. Anti-patterns cần tránh

### 57.1. Dùng production như dev

Không test trực tiếp trên production bằng dữ liệu thật nếu chưa có quy trình.

### 57.2. Deploy không migration plan

Không deploy code cần schema mới khi migration chưa chạy hoặc migration chưa test.

### 57.3. Dùng `latest` trong production

Không biết đang chạy version nào là tự sát vận hành.

### 57.4. Sửa DB trực tiếp để “chữa cháy nhanh”

ERP cần audit và chứng từ điều chỉnh. SQL tay làm mất dấu vết.

### 57.5. Không backup trước deploy lớn

Backup trước deploy lớn là bắt buộc.

### 57.6. Alert quá nhiều

Alert spam khiến người trực bỏ qua alert thật.

### 57.7. Không test restore

Backup chưa từng restore không đáng tin.

---

## 58. Roadmap DevOps theo giai đoạn

### Giai đoạn A - Foundation

```text
- Docker local
- CI backend/frontend
- migration pipeline
- dev environment
- staging environment
- basic logs
```

### Giai đoạn B - Release Ready

```text
- UAT environment
- image registry
- CD pipeline
- release notes
- smoke test automation
- backup scripts
```

### Giai đoạn C - Go-Live Ready

```text
- production environment
- monitoring dashboard
- alerting
- restore test
- security baseline
- rollback plan
```

### Giai đoạn D - Scale

```text
- infra as code nâng cao
- blue/green hoặc canary deploy
- Kubernetes nếu cần
- advanced observability
- PITR database
- DR environment
```

---

## 59. Kết luận

DevOps/CICD cho ERP Phase 1 phải đặt trọng tâm vào **an toàn dữ liệu, ổn định vận hành, kiểm soát release và quan sát hệ thống**.

Với công ty mỹ phẩm có workflow kho, bàn giao ĐVVC, hàng hoàn, QC và gia công ngoài, pipeline kỹ thuật phải bảo vệ các điểm có rủi ro cao nhất:

- stock ledger
- batch/QC status
- pick/pack/handover
- return inspection/disposition
- subcontract receiving/QC
- end-of-day reconciliation
- audit log

Chốt chuẩn Phase 1:

```text
Architecture: Modular Monolith
Backend: Go
Frontend: Next.js + TypeScript
DB: PostgreSQL
Deploy: Docker-based
CI/CD: GitHub Actions/GitLab CI
API Contract: OpenAPI
Storage: S3/MinIO
Monitoring: Prometheus/Grafana hoặc cloud equivalent
Logging: structured logs + centralized logging
Release: versioned, backed up, smoke-tested, rollback-ready
```

Một câu cuối:

> DevOps tốt là thứ khiến team dám deploy, business dám dùng, và ông chủ ngủ được vì biết hệ thống có lỗi cũng không làm mất dấu vết tiền-hàng.

---

## 60. Tài liệu liên quan

- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`
- `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`
- `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`
- `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`

