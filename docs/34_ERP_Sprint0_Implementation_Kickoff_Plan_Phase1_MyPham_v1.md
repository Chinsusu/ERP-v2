# 34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1

**Dự án:** Web ERP cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Giai đoạn:** Phase 1 – Sprint 0 / Implementation Kickoff  
**Phiên bản:** v1.0  
**Mục tiêu tài liệu:** Biến bộ tài liệu 01–33 thành kế hoạch khởi động triển khai thật, dựng nền kỹ thuật, nền sản phẩm, nền quy trình và nền kiểm soát trước khi bước vào sprint build nghiệp vụ.

---

## 1. Tinh thần của Sprint 0

Sprint 0 không phải sprint để làm thật nhiều màn hình.

Sprint 0 là sprint để trả lời 5 câu hỏi sống còn:

1. Kiến trúc đã chạy thật được chưa?
2. Backend Go, frontend Next.js, PostgreSQL, OpenAPI, CI/CD đã nối được với nhau chưa?
3. Module boundary, RBAC, audit log, stock ledger đã có khung chuẩn chưa?
4. Workflow kho, bàn giao ĐVVC, hàng hoàn, gia công ngoài đã được prototype đủ để chứng minh không lệch thực tế chưa?
5. Team đã có cách làm chung, cách review chung, cách test chung, cách release chung chưa?

Một câu chốt:

> **Sprint 0 không tạo ERP hoàn chỉnh. Sprint 0 tạo cái khung sống để từ Sprint 1 trở đi dev không code bừa.**

---

## 2. Nguồn đầu vào của Sprint 0

Sprint 0 phải bám theo các tài liệu sau:

| Nhóm | Tài liệu nguồn | Mục đích dùng trong Sprint 0 |
|---|---|---|
| Business core | `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md` | Scope Phase 1 và module lõi |
| Permission | `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md` | RBAC, approval, field-level restriction |
| Data | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` | master data, trạng thái, field bắt buộc |
| Process | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` | luồng To-Be |
| Workflow thật | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` | As-Is kho, bàn giao, hàng hoàn, gia công ngoài |
| Gap | `21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md` | quyết định chỉnh To-Be theo thực tế |
| Revision | `33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md` | update pack để core docs không lệch nhau |
| Architecture | `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md` | kiến trúc Go backend |
| Coding | `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md` | chuẩn code Go |
| Module | `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md` | chuẩn module boundary |
| UI/UX | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` | design system nghiệp vụ ERP |
| Frontend | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` | kiến trúc Next.js |
| API | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` | API contract |
| Database | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` | PostgreSQL schema standard |
| DevOps | `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` | CI/CD, môi trường, deploy |
| Security | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` | bảo mật, audit, compliance |
| QA | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` | test strategy |
| Backlog | `25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md` | epic/story/sprint plan |
| Handoff | `32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` | source of truth và traceability |

---

## 3. Workflow thực tế phải được đưa vào Sprint 0

Sprint 0 phải chứng minh hệ thống hiểu đúng 4 workflow thực tế đã được cung cấp.

### 3.1. Công việc hằng ngày của kho

Workflow thực tế:

```text
Tiếp nhận đơn hàng trong ngày
→ thực hiện xuất/nhập theo bảng nội quy
→ soạn hàng và đóng gói
→ sắp xếp, tối ưu vị trí kho
→ kiểm kê hàng tồn cuối ngày
→ đối soát số liệu và báo cáo quản lý
→ kết thúc ca làm
```

Hàm ý cho Sprint 0:

- cần có skeleton cho `Warehouse Daily Board`
- cần có khái niệm `shift` hoặc `working day`
- cần có prototype `End-of-Day Reconciliation`
- cần có audit log cho thao tác kho
- cần phân biệt tồn hệ thống, tồn vật lý, tồn khả dụng, tồn lệch

### 3.2. Nội quy nhập kho, xuất kho, đóng hàng, hàng hoàn

Các nhánh nghiệp vụ cần được phản ánh trong prototype:

```text
Nhập kho:
Nhận chứng từ giao hàng
→ kiểm tra số lượng / bao bì / lô
→ đạt thì sắp xếp vào kho
→ ký xác nhận, trưởng kho lưu phiếu nhập
→ không đạt thì trả nhà cung cấp

Xuất kho:
Làm phiếu xuất kho
→ xuất hàng
→ kiểm tra số lượng, đối chiếu vị trí với thực tế
→ ký tên bàn giao
→ trưởng kho lưu phiếu xuất

Đóng hàng:
Nhận phiếu đơn hàng hợp lệ
→ lọc và phân loại đơn theo ĐVVC / đơn lẻ / đơn sàn
→ soạn hàng theo từng đơn
→ đóng gói và kiểm tra tại khu đóng hàng
→ đếm tổng số lượng đơn của mỗi sàn
→ chuyển đến khu vực bàn giao ĐVVC
→ lập/ký xác nhận với ĐVVC

Hàng hoàn:
Nhận hàng từ shipper
→ đưa vào khu vực hàng hoàn
→ quét hàng hoàn
→ kiểm tra tình trạng
→ còn dùng thì chuyển vào kho
→ không dùng thì chuyển lab/kho xử lý
→ lập phiếu nhập kho đầy đủ
```

Hàm ý cho Sprint 0:

- phải có stock movement type cho inbound/outbound/return
- phải có return disposition: `USABLE`, `UNUSABLE`, `NEED_REVIEW`
- phải có scan event skeleton
- phải có warehouse zone/location skeleton
- phải có attachment/audit cho phiếu nhập/xuất/trả hàng

### 3.3. Bàn giao hàng cho ĐVVC

Workflow thực tế:

```text
Phân chia khu vực để hàng
→ để theo thùng/rổ, mỗi thùng/rổ có số lượng bằng nhau
→ đối chiếu số lượng đơn dựa trên bảng
→ lấy hàng và quét mã trực tiếp tại hàm
→ nếu đủ đơn thì ký xác nhận bàn giao với ĐVVC
→ nếu chưa đủ thì kiểm tra lại mã
→ nếu mã chưa có trên hệ thống thì đóng lại
→ nếu đã có trên hệ thống thì tìm lại trong khu vực đóng hàng
```

Hàm ý cho Sprint 0:

- phải có carrier manifest skeleton
- phải có scan verify endpoint
- phải có trạng thái đơn: `Packed`, `ReadyToHandover`, `HandedOver`
- phải có exception flow: thiếu đơn, mã không tồn tại, đã có hệ thống nhưng chưa tìm thấy hàng
- phải có audit log và scan event cho từng lần quét

### 3.4. Sản xuất/gia công ngoài

Workflow thực tế:

```text
Lên đơn hàng với nhà máy
→ xác nhận số lượng, quy cách, mẫu mã
→ cọc đơn hàng, xác nhận thời gian sản xuất/nhận hàng
→ chuyển NVL/bao bì qua nhà máy
→ ký biên bản bàn giao, kèm COA/MSDS/tem phụ/hóa đơn VAT nếu cần
→ làm mẫu và chốt mẫu
→ lưu mẫu
→ sản xuất hàng loạt
→ giao hàng về kho
→ kiểm tra số lượng/chất lượng/hóa đơn
→ hàng đạt thì nhập kho
→ hàng không đạt thì báo lại nhà máy trong 3–7 ngày
→ thanh toán lần cuối
```

Hàm ý cho Sprint 0:

- module production phải có nhánh `Subcontract Manufacturing`, không chỉ sản xuất nội bộ
- phải có external factory order skeleton
- phải có material transfer to factory
- phải có sample approval state
- phải có factory receipt + QC result
- phải có claim window 3–7 ngày
- phải có deposit/final payment field placeholder

---

## 4. Mục tiêu Sprint 0

### 4.1. Mục tiêu kỹ thuật

Sprint 0 phải hoàn tất nền kỹ thuật tối thiểu:

```text
1. Monorepo hoặc multi-repo được chốt
2. Go backend skeleton chạy được
3. Next.js frontend skeleton chạy được
4. PostgreSQL migration chạy được
5. OpenAPI base sinh client được
6. CI pipeline chạy test/lint/build được
7. Docker compose local chạy được
8. Staging hoặc Dev environment deploy được
9. Auth/RBAC skeleton hoạt động
10. Audit log base hoạt động
11. Stock ledger prototype ghi movement được
12. Warehouse scan prototype nhận mã và trả kết quả được
```

### 4.2. Mục tiêu sản phẩm

Sprint 0 phải có demo mỏng nhưng thật:

```text
Login
→ xem menu theo quyền
→ tạo SKU/kho/vị trí cơ bản
→ tạo batch cơ bản
→ tạo stock movement nhập thử
→ xem tồn khả dụng thử
→ tạo đơn bán thử
→ pack thử
→ scan bàn giao thử
→ ghi audit log
```

Không cần đầy đủ nghiệp vụ, nhưng phải chứng minh xương sống đúng.

### 4.3. Mục tiêu vận hành dự án

Sprint 0 phải khóa cách làm việc:

```text
- cách tạo issue/story
- cách viết acceptance criteria
- cách review code
- cách review API
- cách review DB migration
- cách review UI
- cách test trước khi merge
- cách release lên Dev/UAT
- cách ghi decision log
- cách xử lý scope change
```

---

## 5. Phạm vi Sprint 0

### 5.1. In scope

| Mảng | In scope |
|---|---|
| Repository | setup repo, branch, folder, convention |
| Backend | Go app skeleton, module structure, healthcheck, config, middleware |
| Frontend | Next.js app shell, routing, layout, auth shell, design token base |
| Database | PostgreSQL, migration tool, base tables cho user/role/audit/master data/stock ledger |
| API | OpenAPI base, response envelope, error model, generated client |
| Auth | login mock hoặc real basic auth, session/JWT, RBAC skeleton |
| Audit | audit log base cho sensitive action |
| Stock | immutable stock movement prototype |
| Warehouse | daily board skeleton, scan endpoint skeleton, handover prototype |
| QC | QC status enum/skeleton cho batch |
| Subcontract | external factory order skeleton |
| DevOps | Docker compose, CI pipeline, environment variables, staging deploy skeleton |
| QA | smoke test, API test sample, frontend basic test |
| Docs | kickoff checklist, decision log, sprint review output |

### 5.2. Out of scope

Sprint 0 không làm các phần sau:

```text
- full purchase order flow
- full sales order flow
- full production/gia công ngoài
- full accounting posting
- full HRM/CRM/KOL
- full dashboard BI
- tích hợp thật với ĐVVC/sàn/website
- mobile app riêng
- AI/forecasting
- payroll
- loyalty
```

Những phần đó đi vào Sprint 1 trở đi.

---

## 6. Thời lượng đề xuất

Sprint 0 nên chạy **2 tuần** nếu team mạnh, hoặc **3 tuần** nếu team mới bắt đầu làm cùng nhau.

Tao khuyên chọn 2 tuần + 2 ngày buffer.

```text
Tuần 1: dựng nền kỹ thuật + repo + CI/CD + DB + auth + API
Tuần 2: dựng prototype nghiệp vụ xương sống + smoke test + demo + handoff sang Sprint 1
Buffer: fix setup/env/test/documentation
```

---

## 7. Team tham gia Sprint 0

| Vai trò | Trách nhiệm |
|---|---|
| Project Sponsor / CEO | quyết định scope, ưu tiên, ngân sách, điểm không được thỏa hiệp |
| Product Owner | chốt requirement, ưu tiên backlog, nhận demo |
| PM/Scrum Master | điều phối sprint, meeting, timeline, risk |
| BA | map tài liệu vào story, làm rõ workflow |
| Solution Architect | giữ kiến trúc tổng thể, module boundary, integration direction |
| Tech Lead Backend | setup Go backend, coding standard, DB/API, review code |
| Tech Lead Frontend | setup Next.js, UI shell, component standard, review frontend |
| DevOps | CI/CD, environment, Docker, deployment, monitoring base |
| QA Lead | test strategy, smoke test, automation skeleton |
| UI/UX | app shell, design token, core component prototype |
| Warehouse Super User | validate workflow kho, scan, bàn giao, đối soát |
| QA/QC Super User | validate batch/QC status |
| Sales/Ops Super User | validate order/pack/ship skeleton |
| Finance/Ops | validate dữ liệu tiền-hàng cơ bản, audit, approval risk |

---

## 8. RACI Sprint 0

| Hạng mục | CEO | PO | PM | BA | Architect | BE Lead | FE Lead | DevOps | QA | Super User |
|---|---|---|---|---|---|---|---|---|---|---|
| Kickoff scope | A | R | R | C | C | C | C | C | C | C |
| Repo setup | I | I | C | I | C | R | R | C | C | I |
| Architecture baseline | I | C | C | C | A/R | R | R | C | C | I |
| Backend skeleton | I | I | C | I | C | A/R | C | C | C | I |
| Frontend skeleton | I | I | C | I | C | C | A/R | C | C | I |
| Database migration | I | I | C | C | C | A/R | I | C | C | I |
| API/OpenAPI base | I | C | C | C | C | A/R | R | I | C | I |
| RBAC skeleton | C | R | C | C | C | A/R | C | I | C | C |
| Audit log base | C | C | C | C | C | A/R | C | I | C | C |
| Stock ledger POC | C | R | C | C | C | A/R | C | I | C | C |
| Warehouse scan POC | C | R | C | R | C | R | R | I | C | A/C |
| CI/CD | I | I | C | I | C | C | C | A/R | C | I |
| Smoke test | I | C | C | C | C | C | C | C | A/R | C |
| Sprint demo | A | R | R | C | C | C | C | C | C | C |

A = Accountable, R = Responsible, C = Consulted, I = Informed.

---

## 9. Sprint 0 backlog chi tiết

### EPIC S0-01: Project foundation

#### S0-01-01: Setup repository và branch strategy

**Mục tiêu:** tạo nền codebase chuẩn để team bắt đầu làm việc.

**Acceptance Criteria:**

```text
- Repo được tạo và quyền truy cập được cấp đúng người
- Branch strategy được chốt: main / develop / feature / release / hotfix
- Pull request template có sẵn
- Code review checklist có sẵn
- Commit convention được chốt
- README local setup có sẵn
```

**Owner:** Tech Lead + PM  
**Priority:** P0

---

#### S0-01-02: Setup issue/story board

**Acceptance Criteria:**

```text
- Board có các cột: Backlog, Ready, In Progress, Review, QA, Done
- Story template có: mục tiêu, acceptance criteria, dependency, test note
- Bug template có: severity, step reproduce, expected, actual, evidence
- Decision log được tạo
```

**Owner:** PM + BA  
**Priority:** P0

---

### EPIC S0-02: Backend Go foundation

#### S0-02-01: Go backend skeleton

**Acceptance Criteria:**

```text
- Go app chạy được local
- Có healthcheck endpoint: GET /health
- Config đọc từ environment
- Middleware base: request id, logging, auth placeholder, error handler
- Response envelope chuẩn được implement
- Folder structure theo file 12 và 13
```

**Owner:** BE Lead  
**Priority:** P0

---

#### S0-02-02: Module structure base

**Acceptance Criteria:**

```text
- Tạo skeleton modules: auth, masterdata, inventory, qc, sales, shipping, returns, subcontract, audit
- Mỗi module có handler/application/domain/repository/dto/events
- Không module nào gọi thẳng repository của module khác
- Shared package chỉ chứa thứ thật sự dùng chung
```

**Owner:** BE Lead + Architect  
**Priority:** P0

---

#### S0-02-03: Error model và API response chuẩn

**Acceptance Criteria:**

```text
- Success response thống nhất
- Error response có code/message/details/request_id
- Common error codes có sẵn: VALIDATION_ERROR, UNAUTHORIZED, FORBIDDEN, NOT_FOUND, CONFLICT, INSUFFICIENT_STOCK, INVALID_STATE
- Frontend có thể parse lỗi thống nhất
```

**Owner:** BE Lead + FE Lead  
**Priority:** P0

---

### EPIC S0-03: Frontend Next.js foundation

#### S0-03-01: Next.js app shell

**Acceptance Criteria:**

```text
- App chạy local
- Có layout: sidebar, header, content area
- Có login page hoặc mock login
- Có protected route
- Có menu theo permission mock
- Có design token base
```

**Owner:** FE Lead  
**Priority:** P0

---

#### S0-03-02: Core UI components

**Acceptance Criteria:**

```text
- Data table base
- Form wrapper base
- Status chip base
- Confirm modal
- Drawer detail
- Toast notification
- Empty/loading/error state
- Scan input component prototype
```

**Owner:** FE Lead + UI/UX  
**Priority:** P0

---

### EPIC S0-04: Database foundation

#### S0-04-01: PostgreSQL migration setup

**Acceptance Criteria:**

```text
- Migration tool chạy local/dev
- Có baseline migration
- Có rollback hoặc down script cho baseline khi phù hợp
- Naming convention table/index/constraint đúng chuẩn file 17
- Migration được chạy trong CI
```

**Owner:** BE Lead + DevOps  
**Priority:** P0

---

#### S0-04-02: Base tables

**Tables tối thiểu:**

```text
users
roles
permissions
user_roles
audit_logs
idempotency_keys
outbox_events
warehouses
warehouse_locations
items
skus
batches
stock_movements
stock_balances
orders
shipments
carrier_manifests
scan_events
returns
subcontract_orders
```

**Acceptance Criteria:**

```text
- Tables tạo được bằng migration
- Có primary key, foreign key cơ bản
- Có created_at, updated_at, created_by khi phù hợp
- Không có direct stock mutation ngoài stock_movements/stock_balances strategy
```

**Owner:** BE Lead  
**Priority:** P0

---

### EPIC S0-05: OpenAPI/API contract foundation

#### S0-05-01: OpenAPI base file

**Acceptance Criteria:**

```text
- Có openapi.yaml hoặc split structure
- Có /api/v1 prefix
- Có schema ErrorResponse, SuccessResponse, PaginationMeta
- Có security scheme
- Có endpoint healthcheck/auth/master data mẫu
- FE generated client chạy được
```

**Owner:** BE Lead + FE Lead  
**Priority:** P0

---

#### S0-05-02: API codegen integration

**Acceptance Criteria:**

```text
- Generated client được tạo từ OpenAPI
- Frontend gọi được API qua generated client
- CI check OpenAPI lint/codegen
- Không hard-code API shape trong frontend khi đã có schema
```

**Owner:** FE Lead + BE Lead  
**Priority:** P0

---

### EPIC S0-06: Auth, RBAC, audit

#### S0-06-01: Auth skeleton

**Acceptance Criteria:**

```text
- User login được bằng seed account hoặc mock auth có kiểm soát
- Token/session lưu an toàn theo môi trường dev
- Protected API reject request không có auth
- Frontend redirect nếu chưa login
```

**Owner:** BE Lead + FE Lead  
**Priority:** P0

---

#### S0-06-02: RBAC skeleton

**Acceptance Criteria:**

```text
- Có role: CEO, ERP_ADMIN, WAREHOUSE_STAFF, WAREHOUSE_LEAD, QA, SALES_OPS, PRODUCTION_OPS
- API kiểm tra permission mẫu
- Frontend ẩn menu/action theo permission
- Permission denied trả lỗi FORBIDDEN chuẩn
```

**Owner:** BE Lead + FE Lead + BA  
**Priority:** P0

---

#### S0-06-03: Audit log base

**Acceptance Criteria:**

```text
- Sensitive action có ghi audit log
- Audit log ghi actor, action, entity_type, entity_id, before/after hoặc metadata
- Có màn hình xem audit log prototype
- Không cho user thường xóa audit log
```

**Owner:** BE Lead + FE Lead  
**Priority:** P0

---

### EPIC S0-07: Stock ledger prototype

#### S0-07-01: Stock movement write path

**Acceptance Criteria:**

```text
- Tạo stock movement type: OPENING, INBOUND_RECEIPT, SALES_RESERVE, SALES_ISSUE, RETURN_RECEIPT, ADJUSTMENT
- Ghi movement trong transaction
- Cập nhật stock balance theo movement
- Không cho sửa/xóa movement sau khi posted
- Có audit log cho adjustment
```

**Owner:** BE Lead  
**Priority:** P0

---

#### S0-07-02: Available stock calculation prototype

**Acceptance Criteria:**

```text
- Tính được physical_stock
- Tính được reserved_stock
- Tính được available_stock = physical - reserved - hold nếu có
- API trả tồn theo warehouse/SKU/batch
- UI hiển thị tồn thử
```

**Owner:** BE Lead + FE Lead  
**Priority:** P0

---

### EPIC S0-08: Warehouse daily board prototype

#### S0-08-01: Warehouse daily board skeleton

**Acceptance Criteria:**

```text
- UI có board công việc trong ngày
- Có counters: đơn chờ xử lý, đơn đang soạn, đơn đã đóng, đơn chờ bàn giao, hàng hoàn, lệch kiểm kê
- Có filter theo kho/ngày/trạng thái
- Có link sang order/shipment/return prototype
```

**Owner:** FE Lead + BA + Warehouse Super User  
**Priority:** P0

---

#### S0-08-02: End-of-day reconciliation skeleton

**Acceptance Criteria:**

```text
- Có màn hình đối soát cuối ca prototype
- Hiển thị số hệ thống vs số đếm thực tế mock
- Có trạng thái: Open, In Review, Closed
- Khi closed phải ghi audit log
- Có checklist trước khi kết thúc ca
```

**Owner:** BE Lead + FE Lead + Warehouse Super User  
**Priority:** P0

---

### EPIC S0-09: Shipping handover scan prototype

#### S0-09-01: Carrier manifest skeleton

**Acceptance Criteria:**

```text
- Tạo manifest theo carrier/ngày/kho
- Add shipment vào manifest
- Manifest có expected_count, scanned_count, missing_count
- Status: Draft, Ready, Scanning, Completed, Exception
```

**Owner:** BE Lead + FE Lead  
**Priority:** P0

---

#### S0-09-02: Scan verify endpoint/UI

**Acceptance Criteria:**

```text
- Quét mã đơn/mã vận đơn trả về trạng thái hợp lệ/không hợp lệ
- Nếu đơn chưa packed → báo lỗi INVALID_STATE
- Nếu mã không tồn tại → báo NOT_FOUND
- Nếu mã thuộc manifest khác → báo MANIFEST_MISMATCH
- Scan event được ghi lại
- UI cho nhân viên kho thao tác nhanh bằng scanner/keyboard
```

**Owner:** BE Lead + FE Lead + Warehouse Super User  
**Priority:** P0

---

### EPIC S0-10: Returns skeleton

#### S0-10-01: Return receiving skeleton

**Acceptance Criteria:**

```text
- Có form nhận hàng hoàn
- Quét mã đơn/mã vận đơn
- Chọn tình trạng: còn dùng / không dùng / cần kiểm tra
- Tạo stock movement RETURN_RECEIPT nếu còn dùng và được xác nhận
- Nếu không dùng, chuyển vào khu vực lab/kho hỏng placeholder
- Ghi audit log
```

**Owner:** BE Lead + FE Lead + Warehouse Super User  
**Priority:** P0

---

### EPIC S0-11: Subcontract manufacturing skeleton

#### S0-11-01: External factory order skeleton

**Acceptance Criteria:**

```text
- Tạo đơn gia công ngoài với nhà máy
- Có field: factory, product, quantity, spec, sample_required, expected_delivery_date, deposit_status
- Status: Draft, Confirmed, MaterialTransferred, SampleApproved, InProduction, Delivered, QCReview, Accepted, Rejected, Closed
- Có audit log cho status change
```

**Owner:** BE Lead + FE Lead + Production/Ops Super User  
**Priority:** P1

---

#### S0-11-02: Material transfer to factory skeleton

**Acceptance Criteria:**

```text
- Tạo phiếu chuyển NVL/bao bì cho nhà máy
- Có attachment placeholder cho COA/MSDS/tem phụ/hóa đơn VAT nếu cần
- Có signed handover flag
- Có stock movement hoặc placeholder movement type SUBCONTRACT_ISSUE
```

**Owner:** BE Lead + FE Lead  
**Priority:** P1

---

### EPIC S0-12: DevOps/CI/CD foundation

#### S0-12-01: Docker compose local

**Acceptance Criteria:**

```text
- Chạy được backend, frontend, PostgreSQL, Redis nếu cần
- Có seed data dev
- README local setup rõ ràng
- Dev mới có thể setup trong 1 ngày
```

**Owner:** DevOps + Tech Leads  
**Priority:** P0

---

#### S0-12-02: CI pipeline

**Acceptance Criteria:**

```text
- Backend: go test, go vet/lint, build
- Frontend: typecheck, lint, build
- OpenAPI: lint/schema check
- DB migration: dry run/check
- Fail pipeline nếu quality gate không đạt
```

**Owner:** DevOps  
**Priority:** P0

---

#### S0-12-03: Dev/Staging deployment skeleton

**Acceptance Criteria:**

```text
- Deploy được lên Dev hoặc Staging
- Có environment variables riêng
- Có basic healthcheck
- Có log truy cập cơ bản
- Có smoke test sau deploy
```

**Owner:** DevOps  
**Priority:** P0

---

### EPIC S0-13: QA foundation

#### S0-13-01: Smoke test pack

**Acceptance Criteria:**

```text
- Có smoke test checklist cho login, healthcheck, master data, stock movement, scan handover
- Có API test mẫu
- Có frontend smoke test mẫu
- Có test data seed
```

**Owner:** QA Lead  
**Priority:** P0

---

#### S0-13-02: Sprint 0 demo script

**Acceptance Criteria:**

```text
- Demo script viết rõ từng bước
- Demo có dữ liệu mẫu mỹ phẩm
- Demo có tình huống scan thành công và scan lỗi
- Demo có audit log
- Demo có stock movement và tồn khả dụng
```

**Owner:** QA Lead + BA + PO  
**Priority:** P0

---

## 10. Lịch triển khai Sprint 0 đề xuất

### Day 0 – Pre-kickoff

```text
- Xác nhận team và quyền truy cập
- Xác nhận scope Sprint 0
- Xác nhận tech stack: Go, Next.js, PostgreSQL, OpenAPI
- Xác nhận repo model
- Xác nhận môi trường làm việc
- Chốt lịch daily/weekly/demo
```

### Day 1 – Kickoff chính thức

```text
- Walkthrough mục tiêu Sprint 0
- Walkthrough source of truth file 32
- Walkthrough update pack file 33
- Chốt Definition of Ready/Done
- Tạo backlog Sprint 0
- Phân công owner
```

### Day 2 – Foundation setup

```text
- Setup repo
- Setup branch strategy
- Setup Go backend skeleton
- Setup Next.js frontend skeleton
- Setup Docker compose local
```

### Day 3 – Database/API base

```text
- Setup PostgreSQL migration
- Tạo base tables
- Setup OpenAPI base
- FE generated client thử nghiệm
```

### Day 4 – Auth/RBAC/audit base

```text
- Auth skeleton
- RBAC skeleton
- Audit log base
- Menu/action permission ở frontend
```

### Day 5 – Stock ledger POC

```text
- Stock movement write path
- Stock balance calculation
- API xem tồn
- UI xem tồn cơ bản
```

### Day 6 – Warehouse daily board POC

```text
- Warehouse daily board UI
- Counters mock/real basic
- End-of-day reconciliation skeleton
- Audit close shift
```

### Day 7 – Shipping handover scan POC

```text
- Carrier manifest skeleton
- Scan verify endpoint
- Scan UI
- Exception cases: missing/not found/wrong state
```

### Day 8 – Returns + subcontract skeleton

```text
- Return receiving skeleton
- Return disposition
- External factory order skeleton
- Material transfer placeholder
```

### Day 9 – CI/CD + QA hardening

```text
- CI quality gate
- Dev/Staging deployment
- Smoke test pack
- Fix integration issues
```

### Day 10 – Sprint 0 demo & review

```text
- Demo full flow mỏng
- Review architecture
- Review technical debt
- Review backlog Sprint 1
- Go/No-Go cho Sprint 1
```

### Buffer 1–2 ngày

```text
- Fix setup/environment
- Clean documentation
- Update decision log
- Finalize Sprint 1 planning
```

---

## 11. Demo script Sprint 0

Demo phải ngắn, thật, và chứng minh được xương sống.

```text
1. Login bằng user Warehouse Lead
2. Xem menu theo quyền
3. Tạo SKU mẫu: SERUM-VITC-30ML
4. Tạo kho và vị trí: WH-HCM / A01
5. Tạo batch: SVC-TEST-001
6. Tạo stock movement OPENING hoặc INBOUND_RECEIPT
7. Xem tồn vật lý/tồn khả dụng
8. Tạo sales order mock
9. Chuyển order sang packed mock
10. Tạo carrier manifest
11. Quét mã đơn hợp lệ → scanned_count tăng
12. Quét mã sai → hệ thống báo lỗi
13. Xem scan event
14. Xem audit log
15. Tạo return receiving mock
16. Chọn còn dùng/không dùng
17. Tạo external factory order mock
18. Chuyển trạng thái material transferred/sample approved mock
19. Chạy smoke test
20. Xem CI pipeline pass
```

---

## 12. Definition of Ready cho Sprint 0 task

Một task Sprint 0 chỉ được đưa vào `Ready` khi:

```text
- Có owner rõ
- Có acceptance criteria rõ
- Có dependency rõ
- Có tài liệu source of truth liên quan
- Có phạm vi không làm rõ
- Có test note tối thiểu
```

---

## 13. Definition of Done cho Sprint 0 task

Một task chỉ được xem là Done khi:

```text
- Code đã merge qua PR
- Pass lint/typecheck/test liên quan
- Có migration nếu đụng DB
- Có OpenAPI update nếu đụng API
- Frontend dùng generated client nếu có API
- Có audit/security consideration nếu là sensitive action
- Có test evidence
- Có update README/docs nếu cần
- Không phá smoke test
```

---

## 14. Quality gate Sprint 0

Sprint 0 chỉ được xem là thành công nếu các điều kiện sau đạt:

### Technical gate

```text
- Backend build pass
- Frontend build pass
- DB migration pass
- OpenAPI lint/codegen pass
- Docker local run pass
- Dev/Staging deploy pass
```

### Product gate

```text
- Login/RBAC demo được
- Audit log demo được
- Stock ledger POC demo được
- Warehouse daily board skeleton demo được
- Shipping handover scan POC demo được
- Returns skeleton demo được
- Subcontract manufacturing skeleton demo được hoặc ít nhất có status model/API prototype
```

### Process gate

```text
- Story board hoạt động
- Code review flow hoạt động
- QA smoke test có evidence
- Decision log được cập nhật
- Sprint 1 backlog đã ready ít nhất 60–70%
```

---

## 15. Các quyết định cần chốt trong Sprint 0

| Mã | Quyết định | Owner | Deadline |
|---|---|---|---|
| DEC-S0-001 | Monorepo hay multi-repo | Architect + Tech Leads | Day 1 |
| DEC-S0-002 | Auth dùng JWT, session hay hybrid | BE Lead + Security | Day 3 |
| DEC-S0-003 | OpenAPI split structure | BE Lead + FE Lead | Day 3 |
| DEC-S0-004 | DB migration tool chính thức | BE Lead + DevOps | Day 3 |
| DEC-S0-005 | Queue dùng NATS/RabbitMQ hay tạm deferred | Architect + BE Lead | Day 5 |
| DEC-S0-006 | Stock balance update sync hay async | Architect + BE Lead | Day 5 |
| DEC-S0-007 | Scan code format tạm thời | BA + Warehouse Lead | Day 6 |
| DEC-S0-008 | Carrier manifest data model | BA + BE Lead + Warehouse Lead | Day 7 |
| DEC-S0-009 | Return disposition chuẩn | BA + Warehouse + QA/QC | Day 8 |
| DEC-S0-010 | Subcontract status model | BA + Production/Ops | Day 8 |
| DEC-S0-011 | Sprint 1 module priority | PO + PM + Architect | Day 10 |

---

## 16. Rủi ro Sprint 0 và cách xử lý

| Rủi ro | Tác động | Cách xử lý |
|---|---|---|
| Team tranh luận tech stack quá lâu | trễ kickoff | Go/Next/PostgreSQL đã chốt; chỉ bàn implementation detail |
| Scope creep | Sprint 0 thành sprint build quá lớn | giữ nguyên nguyên tắc: skeleton + prototype, không full flow |
| Workflow kho chưa được super user xác nhận | prototype lệch thực tế | mời Warehouse Lead review Day 6–7 |
| Stock ledger thiết kế sai | ảnh hưởng toàn ERP | Architect + BE Lead review kỹ trước khi build tiếp |
| API/FE không khớp | mất thời gian sửa | dùng OpenAPI/codegen từ đầu |
| DB migration bừa | nợ kỹ thuật sớm | migration review bắt buộc |
| RBAC làm hời hợt | rủi ro dữ liệu | implement permission check tối thiểu ngay từ Sprint 0 |
| CI/CD yếu | deploy khó | DevOps phải có quality gate P0 |
| Không có test evidence | demo đẹp nhưng không chắc | QA smoke test bắt buộc |

---

## 17. Sprint 0 Kickoff Meeting Agenda

Thời lượng đề xuất: 2.5–3 giờ.

```text
1. Mục tiêu dự án và Phase 1 – 15 phút
2. Review bộ tài liệu source of truth – 20 phút
3. Review workflow thật cần bám – 20 phút
4. Review kiến trúc Go/Next/PostgreSQL/OpenAPI – 30 phút
5. Review Sprint 0 scope và out of scope – 20 phút
6. Review backlog Sprint 0 – 30 phút
7. Chốt team, owner, RACI – 15 phút
8. Chốt meeting cadence – 10 phút
9. Chốt Definition of Ready/Done – 15 phút
10. Chốt rủi ro và decision log – 20 phút
11. Q&A và next actions – 20 phút
```

---

## 18. Meeting cadence

| Meeting | Tần suất | Người tham gia | Mục tiêu |
|---|---|---|---|
| Daily standup | hằng ngày, 15 phút | PM, dev, QA, BA, tech leads | blockers, progress, plan |
| Architecture checkpoint | 2 lần/tuần | Architect, tech leads, DevOps | giữ kiến trúc không lệch |
| Product checkpoint | 2 lần/tuần | PO, BA, super users, PM | xác nhận workflow/prototype |
| QA checkpoint | 2 lần/tuần | QA, dev, BA | test scope và evidence |
| Sprint review | cuối sprint | toàn team + sponsor | demo và quyết định đi tiếp |
| Retro | cuối sprint | delivery team | cải thiện cách làm |

---

## 19. Sprint 0 outputs bắt buộc

Cuối Sprint 0 phải có:

```text
1. Repo running
2. Local setup README
3. Dev/Staging environment
4. CI pipeline
5. Backend Go skeleton
6. Frontend Next.js skeleton
7. PostgreSQL baseline migration
8. OpenAPI base + generated client
9. Auth/RBAC skeleton
10. Audit log base
11. Stock ledger POC
12. Warehouse daily board skeleton
13. End-of-day reconciliation skeleton
14. Shipping manifest + scan POC
15. Returns skeleton
16. Subcontract manufacturing skeleton
17. Smoke test pack
18. Demo script
19. Decision log
20. Sprint 1 ready backlog
```

---

## 20. Sprint 1 đề xuất sau Sprint 0

Nếu Sprint 0 đạt gate, Sprint 1 nên bắt đầu với nhóm nền lõi:

```text
Sprint 1 Epic đề xuất:
1. Auth/RBAC hoàn thiện hơn
2. Master Data: SKU, item, warehouse, location, supplier, customer basic
3. Inventory Stock Ledger v1
4. Batch/QC status base
5. Warehouse receiving basic
6. Warehouse daily board v1
```

Không nên nhảy ngay vào full sales/shipping nếu master data và stock ledger chưa chắc.

---

## 21. Nguyên tắc không được phá

Trong Sprint 0 và các sprint sau, không được phá các nguyên tắc sau:

```text
1. Không sửa trực tiếp tồn kho ngoài stock movement.
2. Không bỏ qua audit log cho hành động nhạy cảm.
3. Không để frontend tự đoán API shape nếu OpenAPI đã có.
4. Không để module này chọc repository/database riêng của module khác.
5. Không cho scan/handover bỏ qua trạng thái đơn.
6. Không cho batch/QC bị sửa không dấu vết.
7. Không cho migration chạy prod nếu chưa test staging.
8. Không build UI nghiệp vụ nếu chưa có state/status rõ.
9. Không nhận thêm scope Sprint 0 nếu không ảnh hưởng tới khung sống.
10. Không xem demo là Done nếu test evidence không có.
```

---

## 22. Checklist giao việc cho vendor/dev team

Trước khi bắt đầu Sprint 0, vendor/dev team phải xác nhận đã đọc và hiểu:

```text
[ ] 11 Technical Architecture Go Backend
[ ] 12 Go Coding Standards
[ ] 13 Module Component Design Standards
[ ] 14 UI/UX Design System Standards
[ ] 15 Frontend Architecture
[ ] 16 API Contract/OpenAPI Standards
[ ] 17 Database Schema/PostgreSQL Standards
[ ] 18 DevOps/CI-CD Standards
[ ] 19 Security/RBAC/Audit Standards
[ ] 20 Current Workflow As-Is
[ ] 21 Gap Analysis Decision Log
[ ] 24 QA Test Strategy
[ ] 25 Product Backlog Sprint Plan
[ ] 32 Master Document Index/Traceability/Handoff
[ ] 33 Core Docs v1.1 Update Pack
```

Vendor/dev team cũng phải xác nhận:

```text
[ ] Không build trái source of truth
[ ] Có issue/story trước khi code
[ ] Có PR trước khi merge
[ ] Có test evidence trước khi Done
[ ] Có decision log nếu phát sinh quyết định mới
[ ] Có change request nếu thay đổi scope
```

---

## 23. Mẫu Sprint 0 status report

```text
Sprint 0 Status Report
Ngày:

1. Tổng quan:
- Green / Yellow / Red:
- Lý do:

2. Hoàn thành hôm nay:
- 

3. Đang làm:
- 

4. Blocker:
- 

5. Quyết định cần chốt:
- 

6. Rủi ro mới:
- 

7. Test/CI status:
- Backend:
- Frontend:
- DB migration:
- OpenAPI:
- Deploy:

8. Cần CEO/PO hỗ trợ:
- 
```

---

## 24. Mẫu decision log

```text
Decision ID:
Ngày:
Người đề xuất:
Vấn đề:
Options:
Quyết định:
Lý do:
Ảnh hưởng tới tài liệu/module:
Owner update:
Deadline:
Status:
```

---

## 25. Kết luận

Sprint 0 là lớp móng. Làm Sprint 0 tốt thì Sprint 1 trở đi chạy rất nhanh. Làm Sprint 0 hời hợt thì càng build càng lệch.

Với dự án ERP mỹ phẩm này, Sprint 0 phải chứng minh 4 năng lực lõi:

```text
1. Kiến trúc kỹ thuật chạy được.
2. Quyền/audit/stock ledger không bị bỏ qua.
3. Workflow kho – scan – bàn giao – hàng hoàn bám thực tế.
4. Nhánh gia công ngoài được đặt đúng trong kiến trúc từ đầu.
```

Một câu chốt:

> **Sprint 0 không phải để khoe tính năng. Sprint 0 để khóa nền móng, bắt hệ thống biết đi đúng đường ngay từ bước đầu.**
