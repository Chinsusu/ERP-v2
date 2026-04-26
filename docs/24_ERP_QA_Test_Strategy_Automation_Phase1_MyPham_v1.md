# 24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Phiên bản:** v1.0  
**Trạng thái:** Draft dùng để review với PM / QA Lead / Tech Lead / Business Owner  
**Backend:** Go  
**Frontend:** React / Next.js + TypeScript  
**Database:** PostgreSQL  
**API:** REST + OpenAPI 3.1  
**Kiến trúc:** Modular Monolith  

---

## 1. Mục đích tài liệu

Tài liệu này định nghĩa **chiến lược kiểm thử kỹ thuật và tự động hóa test** cho ERP Phase 1.

Tài liệu này không thay thế UAT của business. UAT dùng để người dùng nghiệp vụ nghiệm thu. Tài liệu này dành cho đội kỹ thuật để đảm bảo hệ thống:

- chạy đúng nghiệp vụ lõi;
- không làm sai tồn kho;
- không làm sai batch / QC / hạn dùng;
- không làm sai trạng thái đơn hàng;
- không làm sai bàn giao đơn vị vận chuyển;
- không làm sai hàng hoàn;
- không làm sai luồng gia công ngoài;
- không để lỗi phân quyền, audit, dữ liệu nhạy cảm;
- không vỡ khi release mới.

Nói đơn giản: **UAT hỏi “người dùng có dùng được không”, còn QA/Test Strategy hỏi “hệ thống có đủ chắc để vận hành thật không”.**

---

## 2. Nguồn đầu vào

Tài liệu này dựa trên bộ tài liệu ERP đã có:

- `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
- `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`
- `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`
- `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`
- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`
- `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`
- `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`
- `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`
- `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`
- `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`
- `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md`
- `21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md`
- `22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham.md`
- `23_ERP_Integration_Spec_Phase1_MyPham_v1.md`

Và 4 tài liệu workflow thực tế:

- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

Các điểm workflow thực tế phải được ưu tiên test kỹ:

- kho tiếp nhận đơn trong ngày;
- thực hiện xuất/nhập theo nội quy;
- soạn hàng và đóng gói;
- tối ưu vị trí kho;
- kiểm kê tồn kho cuối ngày;
- đối soát số liệu và báo cáo quản lý;
- kết thúc ca;
- nhập kho có kiểm số lượng, bao bì, lô;
- xuất kho có phiếu xuất, đối chiếu số lượng thực tế, ký bàn giao;
- đóng hàng theo đơn/kênh/ĐVVC;
- bàn giao ĐVVC bằng phân khu, đối chiếu số lượng và quét mã;
- xử lý thiếu đơn khi bàn giao;
- nhận hàng hoàn, quét hàng hoàn, kiểm tra tình trạng, phân loại còn dùng/không dùng;
- gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, duyệt mẫu, sản xuất hàng loạt, nhận hàng, QC, báo lỗi nhà máy trong 3–7 ngày, thanh toán lần cuối.

---

## 3. Nguyên tắc QA cấp cao

### 3.1. Test những nơi có thể làm mất tiền trước

Ưu tiên test những luồng gây thiệt hại trực tiếp:

- sai tồn kho;
- bán vượt tồn khả dụng;
- xuất nhầm batch;
- QC fail nhưng vẫn bán;
- hàng hoàn nhập sai trạng thái;
- bàn giao thiếu đơn cho ĐVVC;
- công nợ/COD lệch;
- trả commission sai;
- user không có quyền nhưng vẫn làm được thao tác nhạy cảm.

### 3.2. Không chỉ test màn hình, phải test trạng thái nghiệp vụ

ERP không chết vì giao diện xấu. ERP chết vì trạng thái sai.

Ví dụ:

```text
Order đã Packed nhưng chưa Handed Over
→ không được coi là đã giao cho ĐVVC.

Batch QC = HOLD
→ không được reserve / pick / ship.

Return Item = unusable
→ không được quay lại available stock.
```

### 3.3. Stock ledger phải là vùng cấm sửa tay

Tồn kho phải được test bằng nguyên tắc:

```text
Không có stock movement hợp lệ
= không được thay đổi tồn.
```

Không test kiểu “xem con số tồn đúng chưa” là chưa đủ. Phải test:

- movement nào tạo ra số tồn;
- movement có document source không;
- movement có audit log không;
- movement có reversal không;
- balance có khớp ledger không.

### 3.4. Batch/QC/hạn dùng là risk layer đặc thù mỹ phẩm

Bất kỳ test nào liên quan nhập, xuất, sản xuất, bán, hoàn đều phải hỏi:

```text
Batch nào?
Hạn dùng nào?
QC status nào?
Kho nào?
Tồn khả dụng hay tồn vật lý?
```

### 3.5. Automation để khóa regression, không phải để khoe tỷ lệ coverage

Test tự động phải ưu tiên:

- stock ledger;
- state machine;
- API contract;
- RBAC;
- approval;
- warehouse scan;
- carrier manifest;
- returns;
- subcontract manufacturing;
- migration;
- integration.

Coverage 80% nhưng bỏ sót stock ledger thì vẫn nguy hiểm. Coverage 50% nhưng khóa được nghiệp vụ tiền-hàng-kho thì có giá trị hơn.

---

## 4. Phạm vi kiểm thử Phase 1

### 4.1. In scope

Phase 1 QA phải bao phủ:

1. Authentication / Authorization
2. RBAC / field-level permission
3. Master Data
4. Purchase / PO / inbound
5. QC receiving / batch release
6. Inventory / stock ledger / stock balance
7. Warehouse daily operations
8. Picking / packing
9. Carrier manifest / shipping handover
10. Sales order / reservation / fulfillment
11. Returns / hàng hoàn
12. Subcontract manufacturing / gia công ngoài
13. Finance basic: AP/AR/COD/expense hooks nếu có trong Phase 1
14. Audit log
15. Approval workflow
16. Notification
17. Integration inbound/outbound
18. Reports/KPI Phase 1
19. Data migration
20. Backup/restore smoke
21. Performance smoke
22. Security smoke

### 4.2. Out of scope hoặc test nhẹ ở Phase 1

Các phần này chỉ test smoke hoặc để Phase 2 nếu chưa build sâu:

- CRM nâng cao;
- HRM nâng cao;
- payroll chi tiết;
- KOL profitability nâng cao;
- BI nâng cao;
- AI forecast;
- mobile app riêng;
- supplier portal;
- dealer portal;
- KOL portal.

---

## 5. Test levels

## 5.1. Unit Test

### Mục tiêu

Test logic nhỏ trong domain/application layer.

### Nên áp dụng cho

- tính available stock;
- validate batch expiry;
- validate QC status;
- order state transition;
- shipment state transition;
- return disposition rule;
- approval rule;
- discount/price rule cơ bản;
- warehouse close rule;
- subcontract manufacturing rule;
- permission policy;
- error code mapping.

### Tool đề xuất

```text
Go: go test, testify
Frontend: Vitest / Jest
```

### Ví dụ unit test phải có

```text
Nếu batch QC = HOLD
Khi sales order reserve stock
Thì hệ thống phải reject với code BATCH_NOT_RELEASED
```

```text
Nếu return disposition = unusable
Khi hoàn tất return inspection
Thì không được cộng vào available stock
```

---

## 5.2. Integration Test

### Mục tiêu

Test module kết nối với database, transaction, repository, queue/outbox.

### Nên áp dụng cho

- tạo PO và nhận hàng;
- QC pass tạo stock movement;
- sales reserve tạo reservation;
- pick/pack/update order status;
- manifest scan/update shipment;
- return inspection/update stock;
- subcontract material transfer;
- factory receipt/QC;
- audit log trong transaction;
- outbox event sau nghiệp vụ.

### Tool đề xuất

```text
Go: testcontainers-go + PostgreSQL container
DB: migration test bằng golang-migrate
Queue: test NATS/RabbitMQ nếu dùng
Redis: test container nếu cần
```

### Nguyên tắc

Integration test không mock database cho các luồng stock, QC, shipping, returns. Những luồng này phải test với PostgreSQL thật trong container.

---

## 5.3. API Contract Test

### Mục tiêu

Đảm bảo backend Go và frontend Next.js hiểu cùng một hợp đồng API.

### Nên test

- OpenAPI schema valid;
- endpoint path/method đúng convention;
- request body validation;
- response envelope;
- error envelope;
- pagination/filter/sort;
- auth header;
- idempotency key;
- optimistic locking;
- permission error;
- action endpoint.

### Tool đề xuất

```text
OpenAPI linter: Spectral
API test: Postman/Newman hoặc Bruno CLI
Contract fuzz: Schemathesis nếu phù hợp
Generated client: openapi-typescript hoặc Orval
```

### Rule bắt buộc

Mọi API public cho frontend phải có trong OpenAPI.

Không chấp nhận endpoint “ẩn” không có contract.

---

## 5.4. End-to-End Test

### Mục tiêu

Test luồng người dùng từ UI tới backend tới database.

### Tool đề xuất

```text
Playwright
```

### Luồng E2E tối thiểu

- login theo role;
- tạo master data cơ bản;
- tạo PO;
- nhận hàng;
- QC pass;
- tồn khả dụng tăng;
- tạo sales order;
- reserve stock;
- pick/pack;
- tạo manifest;
- scan handover;
- delivered;
- return receiving;
- return inspection;
- shift closing.

### Chú ý

E2E không cần bao phủ mọi case nhỏ. E2E phải bao phủ **đường sống còn end-to-end**.

---

## 5.5. Regression Test

### Mục tiêu

Mỗi release không phá nghiệp vụ cũ.

### Regression pack bắt buộc

- auth/RBAC;
- master data active/inactive;
- PO/inbound;
- QC hold/pass/fail;
- stock ledger/balance;
- sales order/reserve;
- pick/pack;
- shipping manifest/handover;
- returns;
- subcontract manufacturing;
- audit log;
- report basic;
- migration smoke.

---

## 5.6. UAT Support Test

UAT do business chạy theo tài liệu `09_ERP_UAT_Test_Scenarios...`. QA kỹ thuật phải chuẩn bị:

- môi trường UAT;
- dữ liệu mẫu;
- user/role;
- checklist;
- bug triage;
- support logging;
- daily UAT summary;
- sign-off record.

---

## 5.7. Performance Smoke Test

### Mục tiêu

Không cần load test quá lớn ở Phase 1, nhưng phải đảm bảo thao tác kho không chậm.

### Nên test

- API list đơn hàng;
- scan đơn;
- reserve stock;
- pick/pack;
- manifest scan;
- stock balance query;
- report tồn kho;
- dashboard kho cuối ngày;
- export CSV/Excel.

### Ngưỡng gợi ý

```text
API đọc danh sách: p95 < 800ms
API action nghiệp vụ: p95 < 1200ms
Scan đơn/handover: p95 < 500ms nếu không có external call
Export lớn: chạy async, không block UI
```

---

## 5.8. Security Smoke Test

### Mục tiêu

Không để lỗi bảo mật cơ bản.

### Test tối thiểu

- user chưa login không gọi API được;
- user khác role không gọi action nhạy cảm được;
- field giá vốn không hiện với role không được phép;
- user không thể sửa stock ledger trực tiếp;
- user không thể pass QC nếu không có quyền QA;
- user không thể handover shipment nếu chưa packed;
- audit log được ghi cho action nhạy cảm;
- export dữ liệu nhạy cảm bị kiểm soát;
- session/token hết hạn đúng.

---

## 6. Test environment strategy

## 6.1. Môi trường

```text
Local      : dev chạy máy cá nhân
Dev        : tích hợp nội bộ dev
Staging    : kiểm thử gần giống production
UAT        : business nghiệm thu
Production : hệ thống thật
```

## 6.2. Nguyên tắc môi trường

- UAT không dùng chung dữ liệu với Staging nếu business đang nghiệm thu.
- Production không chạy test phá dữ liệu.
- Test data phải reset được.
- Database migration phải chạy trên Dev/Staging trước Production.
- Feature flag dùng cho flow chưa muốn bật.
- Không dùng tài khoản admin để test mọi thứ; phải test bằng role thật.

## 6.3. Seed data chuẩn

Bộ seed data phải có:

- 3 kho: kho tổng, kho hàng hoàn, kho quarantine/lab;
- 2 nhà cung cấp;
- 1 nhà máy gia công;
- 5 nguyên liệu;
- 3 bao bì;
- 3 SKU thành phẩm;
- 4 batch với trạng thái khác nhau: HOLD, PASS, FAIL, EXPIRED/CLOSE_TO_EXPIRY;
- 5 khách hàng;
- 2 ĐVVC;
- 5 user theo role: warehouse, QA, sales, purchasing, manager;
- 1 order đủ tồn;
- 1 order thiếu tồn;
- 1 return còn dùng;
- 1 return không dùng;
- 1 subcontract order.

---

## 7. Test data strategy

### 7.1. Nguyên tắc

Test data phải mô phỏng đủ case xấu, không chỉ case đẹp.

Ví dụ với batch:

```text
Batch A: QC HOLD, còn hạn
Batch B: QC PASS, còn hạn
Batch C: QC FAIL, còn hạn
Batch D: QC PASS, cận hạn
Batch E: QC PASS, hết hạn
```

Ví dụ với tồn:

```text
SKU có tồn vật lý nhưng không khả dụng vì QC HOLD
SKU có tồn vật lý nhưng đã reserved hết
SKU có tồn ở kho hàng hoàn nhưng chưa inspect
SKU có tồn ở quarantine nhưng không được bán
```

### 7.2. Không dùng dữ liệu production thật nếu chưa ẩn danh

Nếu dùng dữ liệu thật để test:

- ẩn số điện thoại khách;
- ẩn địa chỉ;
- ẩn thông tin thanh toán;
- ẩn lương/thưởng;
- ẩn giá vốn nếu không cần;
- giới hạn quyền truy cập.

---

## 8. Automation strategy

## 8.1. Automation pyramid đề xuất

```text
Nhiều nhất: Unit tests
Vừa phải : Integration/API tests
Ít hơn   : E2E UI tests
Có chọn lọc: Performance/security smoke
```

## 8.2. Những gì bắt buộc tự động hóa

### Backend

- domain rules;
- state machine;
- permission policy;
- stock ledger transaction;
- QC/batch rule;
- returns disposition;
- shipping manifest scan;
- subcontract manufacturing milestones;
- API response/error contract;
- migration validation.

### Frontend

- form validation;
- role-based UI visibility;
- critical pages render;
- scan UX basic;
- table filter/sort/pagination;
- status chip;
- action button disabled/enabled theo trạng thái;
- Playwright E2E cho luồng kho và order.

### Database

- migration up/down nếu policy cho phép;
- unique constraints;
- foreign keys;
- stock ledger immutability guard;
- index cho query critical;
- idempotency key;
- outbox insert.

## 8.3. Những gì chưa nên tự động hóa quá sớm

- mọi biến thể UI nhỏ;
- mọi filter ít dùng;
- dashboard phụ;
- export format chi tiết chưa ổn định;
- các flow chưa chốt workflow.

---

## 9. Tooling đề xuất

## 9.1. Backend Go

```text
go test
Testify
Testcontainers-Go
pgx test helper
golang-migrate
GolangCI-Lint
```

## 9.2. Frontend

```text
Vitest
React Testing Library
Playwright
TypeScript strict mode
ESLint
```

## 9.3. API

```text
OpenAPI 3.1
Spectral
Postman/Newman hoặc Bruno
Schemathesis nếu cần fuzz API
```

## 9.4. Performance

```text
k6
```

## 9.5. Security

```text
OWASP ZAP baseline scan
npm audit / pnpm audit
govulncheck
trivy image scan
```

## 9.6. CI/CD

```text
GitHub Actions / GitLab CI
Docker
Staging deploy preview nếu có
```

---

## 10. Test ownership

## 10.1. Developer

Dev chịu trách nhiệm:

- unit test;
- integration test module mình làm;
- API contract đúng OpenAPI;
- migration test;
- không merge nếu thiếu test cho rule nghiệp vụ quan trọng.

## 10.2. QA Engineer

QA chịu trách nhiệm:

- test plan;
- test case;
- manual exploratory test;
- API regression pack;
- E2E regression pack;
- defect triage;
- UAT support;
- test report.

## 10.3. Business Owner / Key User

Business chịu trách nhiệm:

- xác nhận expected result;
- chạy UAT;
- xác nhận flow thực tế;
- sign-off;
- phân loại bug hay change request.

## 10.4. Tech Lead

Tech Lead chịu trách nhiệm:

- test architecture;
- CI gate;
- code review test quality;
- quyết định automation scope;
- đảm bảo high-risk flow có test.

---

## 11. Critical risk-based test matrix

| Risk Area | Rủi ro | Test bắt buộc |
|---|---|---|
| Stock Ledger | Tồn sai, thất thoát | Movement, balance, reservation, reversal |
| Batch/QC | Batch fail vẫn bán | QC hold/pass/fail, expiry, release |
| Purchase/Inbound | Nhập sai lô/số lượng | PO receiving, inspection, reject |
| Warehouse Daily | Cuối ca lệch số | shift close, count, reconciliation |
| Picking/Packing | Soạn nhầm đơn | pick task, pack verify, SKU/location |
| Carrier Handover | Giao thiếu đơn | manifest, scan, missing order handling |
| Returns | Hàng hoàn nhập sai | return receive, inspect, disposition |
| Subcontract | Gia công thiếu/mất NVL | material transfer, sample approval, factory receipt |
| RBAC | User làm vượt quyền | permission API/UI, field-level |
| Audit | Không truy vết được | audit event for sensitive actions |
| Integration | Sync sai đơn/trạng thái | idempotency, retry, error handling |
| Migration | Dữ liệu cũ sai | import validation, reconciliation |
| Reporting | Dashboard sai | source mapping, KPI formula test |

---

# 12. Test scenarios chi tiết theo module

## 12.1. Authentication / RBAC / Security

### SEC-001 — Login hợp lệ

**Given** user active  
**When** nhập đúng credential  
**Then** login thành công, nhận session/token hợp lệ.

### SEC-002 — Login sai nhiều lần

**Given** user tồn tại  
**When** nhập sai password nhiều lần theo cấu hình  
**Then** hệ thống lock/rate limit/cảnh báo theo policy.

### SEC-003 — User không có quyền gọi API nhạy cảm

**Given** user Sales  
**When** gọi API `POST /qc/batches/{id}/pass`  
**Then** response `403 FORBIDDEN` và không thay đổi QC status.

### SEC-004 — Field-level permission

**Given** user Warehouse  
**When** xem stock item  
**Then** không thấy giá vốn/cost field.

### SEC-005 — Audit log cho action nhạy cảm

**Given** user QA  
**When** đổi batch từ HOLD sang PASS  
**Then** audit log ghi actor, timestamp, entity, before/after, reason.

---

## 12.2. Master Data

### MD-001 — Tạo SKU thành phẩm hợp lệ

Expected:

- SKU code unique;
- có đơn vị tính;
- có category;
- trạng thái default Draft/Inactive nếu thiếu hồ sơ bắt buộc;
- audit log được ghi.

### MD-002 — Không cho trùng mã SKU

Expected:

- reject với error code `DUPLICATE_SKU_CODE`.

### MD-003 — Inactive item không được dùng tạo PO/SO mới

Expected:

- item inactive không xuất hiện trong dropdown active;
- nếu gọi API trực tiếp thì reject.

### MD-004 — Batch number format

Expected:

- batch_no đúng rule;
- không trùng trong phạm vi SKU + lot source nếu rule yêu cầu.

---

## 12.3. Purchase / Inbound

### PUR-001 — Tạo Purchase Request

Expected:

- trạng thái Draft;
- submit chuyển Submitted;
- nếu vượt ngưỡng phải vào approval.

### PUR-002 — Tạo PO từ PR đã duyệt

Expected:

- PO reference PR;
- giá/số lượng đúng;
- supplier active;
- audit log.

### PUR-003 — Nhận hàng từ PO đầy đủ

Expected:

- tạo inbound receipt;
- số lượng nhận không vượt PO nếu không có quyền override;
- batch/expiry bắt buộc với nguyên liệu/thành phẩm cần quản lô;
- QC status default HOLD hoặc PENDING_QC;
- chưa cộng vào available stock nếu QC chưa pass.

### PUR-004 — Nhận thiếu hàng

Expected:

- PO status Partially Received;
- inbound record ghi số lượng thực nhận;
- không tự close PO.

### PUR-005 — Nhận hàng không đạt

Expected:

- có trạng thái rejected/failed inspection;
- không vào available stock;
- có reason;
- có thể tạo return to supplier nếu workflow bật.

---

## 12.4. QC / Batch

### QC-001 — QC Pass nguyên liệu đầu vào

Expected:

- batch status PASS;
- stock movement `QC_RELEASE` hoặc movement tương đương;
- available stock tăng;
- audit log;
- outbox event `BatchReleased`.

### QC-002 — QC Fail nguyên liệu đầu vào

Expected:

- batch status FAIL;
- available stock không tăng;
- stock ở quarantine/rejected nếu có;
- không cho cấp phát sản xuất;
- không cho bán.

### QC-003 — QC Hold batch

Expected:

- batch HOLD không cho reserve/pick/ship;
- API trả error `BATCH_NOT_RELEASED`.

### QC-004 — Hạn dùng hết hạn

Expected:

- batch expired không được bán;
- nếu cố reserve phải reject;
- near-expiry alert hiển thị nếu cận date.

### QC-005 — Đổi QC status bắt buộc reason

Expected:

- không nhập reason thì reject;
- audit log lưu reason.

---

## 12.5. Inventory / Stock Ledger

### INV-001 — Stock ledger bất biến

Expected:

- không có API update/delete stock movement thông thường;
- chỉ có reversal/adjustment có quyền và reason;
- database không cho xóa movement nếu rule enforce.

### INV-002 — Balance khớp ledger

Expected:

- tổng movement = balance theo SKU + warehouse + batch;
- test job reconciliation phát hiện lệch.

### INV-003 — Reserve stock

Expected:

- available stock giảm theo reservation;
- physical stock không đổi;
- reserved stock tăng;
- không reserve vượt available.

### INV-004 — Release reservation

Expected:

- reserved giảm;
- available tăng lại;
- audit log.

### INV-005 — Xuất kho bán hàng

Expected:

- physical stock giảm;
- reserved stock giảm nếu xuất từ reservation;
- stock movement có source sales order/shipment.

### INV-006 — Chuyển kho

Expected:

- tạo movement OUT ở kho nguồn;
- tạo movement IN ở kho đích;
- transaction atomic;
- batch/expiry giữ nguyên.

### INV-007 — Điều chỉnh kiểm kê

Expected:

- cần quyền Warehouse Manager/Finance tùy rule;
- cần reason;
- tạo adjustment movement;
- audit log.

---

## 12.6. Warehouse Daily Board / Shift Closing

Workflow thực tế có bước tiếp nhận đơn trong ngày, làm xuất/nhập, soạn/đóng hàng, sắp xếp tối ưu kho, kiểm kê cuối ngày, đối soát và kết thúc ca. Vì vậy Phase 1 phải test riêng daily board và shift closing.

### WHD-001 — Mở ca kho

Expected:

- tạo shift/session;
- ghi người phụ trách;
- chỉ user warehouse được thao tác;
- không mở trùng ca cùng kho nếu rule không cho.

### WHD-002 — Daily board hiển thị đúng task trong ngày

Expected:

- đơn cần pick/pack;
- inbound cần nhận;
- return cần inspect;
- manifest cần bàn giao;
- stock count cần làm;
- exception cần xử lý.

### WHD-003 — Kiểm kê cuối ngày

Expected:

- nhập counted qty;
- so với system qty;
- nếu lệch phải có variance;
- variance vượt ngưỡng cần approval;
- chưa close shift nếu còn variance chưa xử lý theo rule.

### WHD-004 — Đối soát cuối ngày

Expected:

- số đơn packed/handed over/delivered/returned khớp;
- tồn cuối ngày khớp movement;
- có report cho quản lý;
- close shift ghi timestamp/user.

### WHD-005 — Không cho close shift nếu còn task critical chưa xử lý

Expected:

- nếu manifest chưa handover, return chưa inspect hoặc variance chưa xác nhận thì cảnh báo/block theo config.

---

## 12.7. Sales Order / Reservation

### SO-001 — Tạo sales order hợp lệ

Expected:

- order status Draft/Confirmed theo flow;
- kiểm SKU active;
- kiểm customer active;
- giá/discount theo quyền;
- audit log.

### SO-002 — Confirm order và reserve stock

Expected:

- reserve theo available stock;
- batch được chọn theo FEFO/FIFO nếu áp dụng;
- batch QC PASS;
- không reserve hàng hoàn chưa inspect.

### SO-003 — Thiếu tồn

Expected:

- reject hoặc partial reserve tùy rule;
- hiển thị SKU thiếu;
- không tạo pick task nếu chưa reserve.

### SO-004 — Discount vượt quyền

Expected:

- order vào approval;
- không cho ship trước khi duyệt nếu rule yêu cầu.

### SO-005 — Cancel order đã reserved

Expected:

- release reservation;
- order cancelled;
- audit log;
- không giải phóng nếu hàng đã shipped nếu không qua return/reversal.

---

## 12.8. Picking / Packing

### PK-001 — Tạo pick task từ order đã reserved

Expected:

- pick task có SKU/batch/location/qty;
- user warehouse nhận task;
- không pick batch khác nếu không override có quyền.

### PK-002 — Pick thiếu

Expected:

- ghi thiếu hàng;
- task exception;
- order không chuyển Packed;
- báo manager.

### PK-003 — Pack verify theo đơn

Expected:

- quét SKU/order;
- kiểm đúng số lượng;
- trạng thái chuyển Packed khi đủ;
- nếu sai SKU/batch cảnh báo.

### PK-004 — Khu vực đóng hàng

Expected:

- đơn Packed được chuyển vào packing/handover staging area;
- ghi vị trí rổ/thùng nếu áp dụng;
- phục vụ quy trình bàn giao ĐVVC.

---

## 12.9. Carrier Manifest / Shipping Handover

Workflow thực tế yêu cầu phân chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn trên bảng, lấy hàng và quét mã trước bàn giao. Nếu đủ thì ký xác nhận, nếu chưa đủ thì kiểm tra mã và tìm lại trong khu vực đóng hàng.

### SHIP-001 — Tạo manifest theo ĐVVC

Expected:

- chỉ include order Packed;
- gom theo carrier/channel/cutoff;
- manifest có số lượng đơn expected;
- có trạng thái Draft/Open/Ready/HandedOver.

### SHIP-002 — Scan đơn vào manifest

Expected:

- scan đúng mã đơn/vận đơn;
- nếu đơn thuộc manifest thì mark scanned;
- nếu không thuộc manifest thì cảnh báo;
- không scan trùng.

### SHIP-003 — Đối chiếu số lượng manifest

Expected:

- expected count vs scanned count;
- nếu bằng nhau mới cho handover;
- nếu thiếu thì trạng thái exception.

### SHIP-004 — Handover đủ đơn

Expected:

- manifest status HandedOver;
- shipment status HandedOver;
- order status HandedOver/In Transit tùy rule;
- audit log;
- có biên bản bàn giao hoặc export manifest.

### SHIP-005 — Handover thiếu đơn

Expected:

- hệ thống list đơn thiếu;
- yêu cầu kiểm tra mã;
- nếu không tìm thấy thì tạo exception;
- không cho xác nhận đủ nếu chưa scan đủ.

### SHIP-006 — Scan sau khi handover

Expected:

- không cho scan thêm nếu manifest đã HandedOver trừ quyền reopen;
- reopen cần manager approval/audit.

---

## 12.10. Returns / Hàng hoàn

Workflow thực tế: nhận hàng từ shipper, đưa vào khu vực hàng hoàn, quét hàng hoàn, kiểm tra tình trạng, phân loại còn sử dụng hoặc không sử dụng, chuyển vào kho hoặc lab, lập phiếu nhập kho ký tên.

### RET-001 — Nhận hàng hoàn từ shipper

Expected:

- tạo return receiving;
- scan order/tracking;
- đưa vào return area;
- trạng thái Pending Inspection;
- chưa cộng available stock.

### RET-002 — Quét hàng hoàn

Expected:

- scan đúng mã đơn/SKU;
- nếu không tìm thấy order thì tạo unknown return case;
- audit log.

### RET-003 — Inspect hàng hoàn còn sử dụng

Expected:

- disposition = reusable;
- chuyển về kho khả dụng hoặc kho chờ QC theo rule;
- stock movement phù hợp;
- batch/expiry được ghi nhận;
- có phiếu nhập.

### RET-004 — Inspect hàng hoàn không sử dụng

Expected:

- disposition = unusable;
- chuyển lab/kho hỏng/quarantine;
- không vào available stock;
- có reason/hình ảnh nếu cần.

### RET-005 — Hàng hoàn thiếu/thừa so với đơn

Expected:

- tạo exception;
- không auto close;
- báo CSKH/warehouse manager.

---

## 12.11. Subcontract Manufacturing / Gia công ngoài

Workflow thực tế thể hiện: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn/chốt thời gian, chuyển NVL/bao bì qua nhà máy, ký biên bản bàn giao, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao về kho, kiểm tra số lượng/chất lượng, nếu lỗi báo lại nhà máy trong 3–7 ngày, thanh toán lần cuối.

### SUB-001 — Tạo đơn gia công

Expected:

- factory active;
- SKU/spec/BOM/version rõ;
- số lượng/quy cách/mẫu mã;
- expected delivery date;
- trạng thái Draft/Confirmed;
- deposit schedule nếu có.

### SUB-002 — Chuyển NVL/bao bì cho nhà máy

Expected:

- tạo transfer order/source document;
- xuất khỏi kho công ty hoặc chuyển sang kho “Factory Consignment”;
- ghi batch/qty;
- có biên bản bàn giao;
- stock ledger không mất dấu.

### SUB-003 — Duyệt mẫu

Expected:

- sample status Pending/Approved/Rejected;
- không cho sản xuất hàng loạt nếu sample chưa approved;
- lưu file/hình ảnh/ghi chú.

### SUB-004 — Nhận hàng gia công về kho

Expected:

- inbound receipt reference subcontract order;
- batch finished goods;
- QC status HOLD/PENDING;
- kiểm số lượng;
- chưa available nếu chưa QC pass.

### SUB-005 — Hàng gia công đạt

Expected:

- QC pass;
- nhập kho thành phẩm khả dụng;
- đóng subcontract order nếu đủ;
- chuẩn bị final payment nếu finance module có.

### SUB-006 — Hàng gia công không đạt

Expected:

- reject/claim case;
- không available;
- ghi lỗi/chứng cứ;
- deadline claim 3–7 ngày được theo dõi;
- notification cho purchasing/QA/manager.

### SUB-007 — Nhà máy giao thiếu

Expected:

- partial receipt;
- remaining qty còn open;
- tạo claim hoặc backorder;
- không thanh toán final nếu chưa xử lý theo rule.

---

## 12.12. Finance Basic / COD / AP / AR

### FIN-001 — PO receipt tạo AP pending nếu scope có

Expected:

- công nợ NCC chỉ ghi khi hóa đơn/receipt hợp lệ theo rule;
- PO rejected không tạo payable.

### FIN-002 — COD reconciliation

Expected:

- shipment delivered/COD collected;
- đối soát với file ĐVVC;
- chênh lệch tạo exception.

### FIN-003 — Return ảnh hưởng công nợ/doanh thu

Expected:

- return approved tạo refund/credit note nếu scope có;
- commission/COD không tính sai.

---

## 12.13. Reports / KPI

### RPT-001 — Tồn kho theo batch

Expected:

- physical/reserved/available đúng;
- filter kho/SKU/batch;
- batch HOLD/FAIL không tính available.

### RPT-002 — Near-expiry report

Expected:

- hiển thị batch cận date theo threshold;
- không lẫn batch hết hạn với cận hạn nếu report tách.

### RPT-003 — Warehouse daily close report

Expected:

- số inbound/outbound/packed/handed over/return/count variance;
- khớp với shift data.

### RPT-004 — Carrier handover report

Expected:

- manifest count;
- scanned count;
- missing count;
- handover time/user/carrier.

### RPT-005 — Subcontract claim report

Expected:

- hàng lỗi;
- deadline claim 3–7 ngày;
- status claim;
- nhà máy liên quan.

---

# 13. API test standards

## 13.1. Mọi API phải test các nhóm sau

```text
1. Success case
2. Validation error
3. Permission denied
4. Not found
5. Conflict/state invalid
6. Idempotency nếu là action tạo giao dịch
7. Audit log nếu là action nhạy cảm
```

## 13.2. Error response phải thống nhất

Ví dụ:

```json
{
  "success": false,
  "code": "INSUFFICIENT_STOCK",
  "message": "Tồn khả dụng không đủ",
  "details": {
    "sku_code": "SERUM-VITC-30ML",
    "requested_qty": 10,
    "available_qty": 6
  }
}
```

## 13.3. Action endpoint cần test idempotency

Các action cần idempotency:

- confirm order;
- reserve stock;
- receive goods;
- QC pass/fail;
- create stock adjustment;
- pack order;
- scan manifest;
- confirm handover;
- receive return;
- complete return inspection;
- transfer material to factory;
- receive subcontract goods.

Test:

```text
Gửi cùng request với same idempotency key 2 lần
→ kết quả không nhân đôi movement/chứng từ.
```

---

# 14. Database testing standards

## 14.1. Migration test

Mỗi migration phải test:

- chạy được trên database sạch;
- chạy được trên database có dữ liệu mẫu;
- không mất dữ liệu quan trọng;
- constraint đúng;
- index cho query critical;
- rollback strategy rõ, dù không phải migration nào cũng down được an toàn.

## 14.2. Constraint test

Phải có test hoặc verification cho:

- unique SKU code;
- unique batch rule;
- FK source document;
- stock movement source_type/source_id;
- không negative qty nếu không cho phép;
- idempotency key unique;
- audit log entity reference;
- outbox event status.

## 14.3. Reconciliation SQL

Cần có SQL/test job để đối soát:

```text
stock_balance = SUM(stock_movements)
reserved_qty = SUM(active_reservations)
available_qty = physical_qty - reserved_qty - hold_qty/quarantine_qty tùy rule
```

---

# 15. UI/E2E test standards

## 15.1. UI critical interactions

Phải test:

- login;
- route guard;
- role-based menu;
- table filter/sort/pagination;
- create/edit/detail page;
- action button enabled/disabled theo status;
- modal confirm sensitive action;
- scan input auto-focus;
- scan success/fail feedback;
- file attachment;
- audit timeline display.

## 15.2. Playwright smoke pack

### E2E-001 — Warehouse daily happy path

```text
Login Warehouse
→ xem daily board
→ pick order
→ pack order
→ add to manifest
→ scan handover
→ close shift
```

### E2E-002 — QC blocks sale

```text
Login QA/Warehouse/Sales
→ receive goods pending QC
→ attempt sales reserve
→ expect blocked
→ QA pass
→ attempt sales reserve again
→ expect success
```

### E2E-003 — Return reusable

```text
Receive return
→ scan item
→ inspect reusable
→ stock moves to available/quarantine according to rule
```

### E2E-004 — Subcontract manufacturing happy path

```text
Create subcontract order
→ transfer materials
→ approve sample
→ receive finished goods
→ QC pass
→ stock available
```

---

# 16. Performance test scope

## 16.1. k6 smoke scenarios

### PERF-001 — Scan manifest

- 20 concurrent users;
- 1000 scan actions;
- p95 < 500ms nếu không gọi external carrier;
- error rate < 1%.

### PERF-002 — Order list

- filter theo ngày/kênh/status;
- p95 < 800ms;
- pagination không load toàn bộ dữ liệu.

### PERF-003 — Stock balance query

- filter SKU/kho/batch;
- p95 < 800ms;
- dùng index đúng.

### PERF-004 — Export report

- export lớn phải async;
- không block API request;
- có trạng thái job.

---

# 17. Security test scope

## 17.1. OWASP baseline

Test tối thiểu:

- broken access control;
- injection basic;
- sensitive data exposure;
- weak auth/session;
- CSRF nếu dùng cookie session;
- file upload validation;
- rate limiting sensitive endpoints;
- insecure direct object reference.

## 17.2. ERP-specific security tests

- user không được xem dữ liệu kho/giá của chi nhánh không có quyền nếu multi-branch;
- warehouse không thấy giá vốn;
- sales không pass QC;
- QA không sửa giá;
- finance không sửa stock movement;
- admin kỹ thuật không tự động có quyền nghiệp vụ nếu policy tách;
- break-glass access có audit riêng;
- export stock/cost/payable phải log.

---

# 18. Defect management

## 18.1. Severity

### Sev 1 — Blocker

- không login được;
- không tạo/xử lý đơn;
- sai tồn kho;
- QC fail vẫn bán;
- giao hàng không cập nhật;
- mất dữ liệu;
- security breach nghiêm trọng.

### Sev 2 — High

- sai trạng thái nghiệp vụ;
- thiếu audit log action nhạy cảm;
- hàng hoàn cập nhật sai;
- manifest scan sai;
- report quan trọng sai;
- permission sai nhưng chưa gây mất tiền ngay.

### Sev 3 — Medium

- lỗi UI ảnh hưởng thao tác nhưng có workaround;
- filter/sort sai ở màn phụ;
- message chưa rõ;
- export thiếu cột không critical.

### Sev 4 — Low

- lỗi chính tả;
- spacing UI;
- icon/format nhỏ.

## 18.2. Bug report template

```text
Title:
Environment:
Module:
Severity:
Role/User:
Steps to reproduce:
Actual result:
Expected result:
Evidence:
Data used:
Related document:
Suspected impact:
```

## 18.3. Bug vs Change Request

Nếu hệ thống không đúng tài liệu đã chốt → bug.

Nếu business muốn đổi quy trình/tính năng ngoài tài liệu → change request.

Nếu tài liệu mâu thuẫn nhau → document issue, cần decision log.

---

# 19. Entry / Exit criteria

## 19.1. Entry criteria cho QA cycle

- PRD/story đã rõ acceptance criteria;
- API contract draft có;
- test environment sẵn;
- seed data có;
- build deploy thành công;
- migration chạy thành công;
- critical logs hoạt động.

## 19.2. Exit criteria cho module release

- không còn Sev 1/Sev 2 open;
- Sev 3 có workaround và được business chấp nhận;
- unit/integration/API test pass;
- E2E smoke pass;
- migration smoke pass;
- audit/security smoke pass cho action nhạy cảm;
- UAT key flow pass hoặc scheduled;
- release notes có.

## 19.3. Exit criteria cho Go-Live readiness

- stock reconciliation pass;
- opening balance verified;
- user/role verified;
- critical workflows UAT signed off;
- integration smoke pass;
- backup/restore rehearsal pass;
- rollback plan có;
- support team trực hypercare;
- SOP/training đã bàn giao.

---

# 20. CI quality gates

## 20.1. Pull request gate

PR không được merge nếu:

- Go lint fail;
- frontend lint/typecheck fail;
- unit test fail;
- API contract fail;
- migration compile/check fail;
- thiếu test cho action high-risk;
- thay đổi OpenAPI nhưng chưa regenerate client;
- thay đổi DB nhưng chưa có migration;
- thay đổi permission nhưng chưa update test.

## 20.2. Staging deploy gate

Không deploy staging nếu:

- integration test critical fail;
- Playwright smoke fail;
- migration fail;
- Docker image scan có critical vulnerability chưa xử lý;
- environment config thiếu.

## 20.3. Production deploy gate

Không deploy production nếu:

- chưa backup;
- release notes chưa có;
- rollback plan chưa rõ;
- smoke test staging fail;
- database migration chưa review;
- on-call/support chưa sẵn sàng;
- business chưa đồng ý release window.

---

# 21. Test metrics

QA Lead nên theo dõi:

```text
- số test case theo module
- pass/fail/block rate
- defect count theo severity
- defect reopen rate
- automation pass rate
- flaky test count
- API contract violation count
- regression execution time
- UAT issue count
- Sev 1/2 aging
- production incident sau release
```

Metric không phải để làm đẹp báo cáo. Metric dùng để biết module nào đang rủi ro.

---

# 22. Minimum automation coverage by module

| Module | Unit | Integration | API | E2E | Ghi chú |
|---|---:|---:|---:|---:|---|
| Auth/RBAC | High | Medium | High | Medium | Permission phải chắc |
| Master Data | Medium | Medium | High | Low | CRUD + validation |
| Purchase/Inbound | Medium | High | High | Medium | Có transaction |
| QC/Batch | High | High | High | Medium | Risk cao |
| Inventory | High | Very High | High | Medium | Stock ledger bắt buộc |
| Sales Order | High | High | High | Medium | Reserve/discount/state |
| Picking/Packing | Medium | High | High | High | Kho dùng thật |
| Shipping/Handover | High | High | High | High | Scan/manifest critical |
| Returns | High | High | High | High | Hàng hoàn dễ thất thoát |
| Subcontract | High | High | High | Medium | Gia công ngoài đặc thù |
| Reports | Low | Medium | Medium | Low | Kiểm formula/source |
| Integration | Medium | High | High | Medium | Retry/idempotency |

---

# 23. Test case priority

## P0 — Bắt buộc trước go-live

- login/RBAC;
- master SKU/batch basic;
- PO/inbound;
- QC hold/pass/fail;
- stock ledger/balance;
- order reserve;
- pick/pack;
- manifest scan/handover;
- return receive/inspect;
- subcontract material transfer/receipt/QC;
- shift close/reconciliation;
- audit log;
- migration opening stock;
- backup/restore smoke.

## P1 — Nên có trước go-live

- report tồn kho;
- near-expiry;
- partial receipt;
- partial shipment;
- return exception;
- COD reconciliation;
- notification;
- export control;
- permission field-level.

## P2 — Có thể sau go-live trong hypercare

- performance nâng cao;
- automation mở rộng;
- dashboard phụ;
- complex finance edge cases;
- integration marketplace nâng sâu nếu Phase 1 chưa live toàn bộ kênh.

---

# 24. QA checklist theo nghiệp vụ đặc thù

## 24.1. Checklist Stock

- [ ] Tồn vật lý không tự thay đổi nếu không có movement.
- [ ] Available stock đúng formula.
- [ ] Reserved stock đúng order reservation.
- [ ] QC HOLD/FAIL không available.
- [ ] Hàng hoàn pending inspect không available.
- [ ] Hàng gia công ở nhà máy không biến mất khỏi tracking.
- [ ] Adjustment có reason/approval/audit.
- [ ] Reconciliation phát hiện lệch.

## 24.2. Checklist Batch/QC

- [ ] Batch bắt buộc ở item quản lý lô.
- [ ] Expiry bắt buộc nếu item yêu cầu hạn dùng.
- [ ] QC pass mới bán được.
- [ ] QC fail không cấp phát/sell.
- [ ] QC change có reason.
- [ ] QC status hiển thị rõ trên UI.

## 24.3. Checklist Warehouse Daily

- [ ] Daily board đủ task.
- [ ] Pick/pack không bỏ qua scan nếu rule yêu cầu.
- [ ] Packing area/staging area có tracking.
- [ ] End-of-day count có variance.
- [ ] Shift close có report.
- [ ] Không close ca khi còn exception critical nếu config block.

## 24.4. Checklist Carrier Handover

- [ ] Manifest theo ĐVVC.
- [ ] Expected count đúng.
- [ ] Scan không cho trùng.
- [ ] Missing order hiển thị rõ.
- [ ] Không handover nếu chưa đủ đơn.
- [ ] Handover có audit và export/biên bản.

## 24.5. Checklist Returns

- [ ] Return nhận vào khu riêng.
- [ ] Pending inspection không available.
- [ ] Reusable và unusable đi hai luồng khác nhau.
- [ ] Unusable không bán lại.
- [ ] Unknown return có exception.
- [ ] Có phiếu nhập/ký xác nhận nếu workflow yêu cầu.

## 24.6. Checklist Subcontract

- [ ] Đơn gia công có factory/spec/qty/timeline.
- [ ] Chuyển NVL/bao bì có ledger/source doc.
- [ ] Duyệt mẫu trước sản xuất hàng loạt.
- [ ] Nhận hàng về kho có QC HOLD/PENDING.
- [ ] Hàng lỗi có claim và deadline 3–7 ngày.
- [ ] Final payment chỉ mở khi điều kiện nghiệm thu đạt nếu finance scope có.

---

# 25. Release regression pack đề xuất

## Daily smoke pack

Chạy mỗi ngày trên Dev/Staging:

```text
- login
- create/read SKU
- create PO
- receive goods
- QC pass
- check stock
- create order
- reserve
- pack
- manifest scan
- handover
```

## Weekly regression pack

Chạy trước release:

```text
- full P0 pack
- returns
- subcontract
- stock adjustment
- shift closing
- RBAC matrix
- API contract
- Playwright critical flows
```

## Pre-production pack

Trước deploy production:

```text
- migration dry run
- backup/restore rehearsal
- production-like smoke
- security smoke
- performance smoke
- rollback rehearsal nếu migration rủi ro
```

---

# 26. Definition of Done cho QA

Một feature được coi là done khi:

- acceptance criteria được test;
- unit/integration/API test phù hợp đã có;
- E2E smoke cập nhật nếu feature thuộc luồng critical;
- RBAC/permission test nếu có action nhạy cảm;
- audit log test nếu có data sensitive/action sensitive;
- OpenAPI cập nhật;
- test data/seed cập nhật nếu cần;
- không còn Sev 1/2;
- QA sign-off;
- business sign-off nếu thuộc UAT scope.

---

# 27. Những bẫy QA cần tránh

### 27.1. Chỉ test case đẹp

ERP thực tế chết vì case xấu:

- thiếu hàng;
- dư hàng;
- quét trùng;
- scan sai manifest;
- hàng hoàn không có order;
- QC fail;
- nhà máy giao thiếu;
- user không đủ quyền;
- thao tác lặp do mạng lag.

### 27.2. Không test idempotency

Người dùng bấm 2 lần, mạng retry, webhook gửi lại. Nếu không test idempotency, ERP sẽ nhân đôi phiếu, nhân đôi movement, nhân đôi thanh toán.

### 27.3. Không test quyền ở API

Ẩn nút trên UI chưa đủ. Phải test API từ backend.

### 27.4. Không test reconciliation

Số tồn nhìn đúng lúc đầu chưa chắc đúng sau 1000 giao dịch. Phải có reconciliation.

### 27.5. Không test workflow thực tế của kho

Nếu chỉ test theo flow “ERP mẫu” mà không test daily board, packing area, manifest scan, hàng hoàn, shift closing, hệ thống sẽ lệch thực tế ngay khi go-live.

---

# 28. Roadmap QA automation

## Sprint 1–2

- setup test framework;
- seed data;
- unit test domain core;
- API contract lint;
- migration test baseline.

## Sprint 3–4

- integration tests cho inventory/QC/sales;
- Playwright login/master data;
- regression pack P0 draft.

## Sprint 5–6

- shipping handover automation;
- returns automation;
- subcontract automation;
- performance smoke scripts.

## Pre-Go-Live

- full regression pack;
- migration dry run;
- UAT support;
- security baseline;
- backup/restore smoke;
- hypercare test checklist.

---

# 29. Handoff checklist cho QA Lead

Trước khi bắt đầu execution, QA Lead cần có:

- [ ] môi trường test ổn định;
- [ ] user/role test;
- [ ] seed data;
- [ ] API docs/OpenAPI;
- [ ] PRD/story acceptance criteria;
- [ ] bug tracking tool;
- [ ] UAT schedule;
- [ ] release calendar;
- [ ] module owner;
- [ ] escalation contact;
- [ ] test report template;
- [ ] regression pack;
- [ ] automation repo configured.

---

# 30. Kết luận

Phase 1 của ERP mỹ phẩm không được test như một hệ CRUD thông thường.

Phải test như một hệ vận hành thật, nơi mỗi action đều có thể làm lệch:

- hàng;
- tiền;
- batch;
- hạn dùng;
- đơn hàng;
- bàn giao vận chuyển;
- hàng hoàn;
- gia công ngoài;
- audit trách nhiệm.

Ưu tiên số một của QA không phải là bắt lỗi giao diện. Ưu tiên số một là **khóa rủi ro tiền-hàng-kho-batch-QC**.

Một hệ ERP tốt không phải là hệ không bao giờ có lỗi. Một hệ ERP tốt là hệ mà khi lỗi xảy ra, lỗi bị chặn sớm, có dấu vết, có đối soát, có người chịu trách nhiệm, và không lan thành thất thoát lớn.

---

## 31. Tài liệu tiếp theo đề xuất

Sau tài liệu này, nên làm tiếp:

```text
25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md
```

Tài liệu đó sẽ biến toàn bộ PRD, API, DB, UI, QA strategy thành:

- epic;
- user story;
- acceptance criteria;
- priority;
- dependency;
- sprint plan;
- release milestone.

Đó là cầu nối từ “thiết kế hệ thống” sang “đội dev làm việc từng tuần”.
