# 35_ERP_Vendor_RFP_SOW_Evaluation_Pack_Phase1_MyPham_v1

**Dự án:** Web ERP cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Giai đoạn:** Phase 1  
**Tài liệu:** Vendor RFP / SOW / Evaluation Pack  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-25  
**Ngôn ngữ:** Vietnamese  
**Mục tiêu:** Dùng để gửi cho vendor/dev team, so sánh báo giá, khóa phạm vi triển khai, đánh giá năng lực nhà thầu và ký SOW rõ ràng.

---

## 1. Vai trò của tài liệu này

Tài liệu này dùng để biến bộ hồ sơ ERP 01–34 thành một gói làm việc có thể giao cho vendor hoặc đội triển khai.

Nó trả lời 6 câu hỏi lớn:

1. Vendor phải hiểu doanh nghiệp đang cần gì?
2. Phase 1 bao gồm những module nào?
3. Vendor phải giao những deliverable nào?
4. Tiêu chí nghiệm thu là gì?
5. Chấm điểm vendor như thế nào?
6. Hợp đồng/SOW cần khóa những gì để tránh đội chi phí và lệch scope?

Một câu rất thực chiến:

> Vendor giỏi không chỉ là vendor code được. Vendor giỏi là vendor hiểu được luồng hàng, tiền, kho, QC, giao hàng, dữ liệu và rủi ro vận hành của doanh nghiệp.

---

## 2. Bối cảnh dự án

Công ty vận hành trong ngành mỹ phẩm, có các hoạt động chính:

- R&D / phát triển sản phẩm
- mua nguyên vật liệu, bao bì, phụ liệu
- làm việc với nhà máy/gia công ngoài
- nhập kho, xuất kho, đóng hàng, bàn giao đơn vị vận chuyển
- quản lý batch/lô, hạn dùng, QC
- bán hàng đa kênh
- xử lý hàng hoàn
- đối soát cuối ca/cuối ngày
- tài chính/công nợ cơ bản
- chuẩn bị mở rộng CRM, HRM, KOL, Finance nâng cao ở Phase 2

Workflow thực tế cần vendor đặc biệt chú ý:

- Kho có nhịp vận hành hằng ngày: tiếp nhận đơn trong ngày, xuất/nhập, soạn và đóng gói, tối ưu vị trí kho, kiểm kê cuối ngày, đối soát số liệu, báo cáo quản lý, kết thúc ca.
- Nội quy kho tách rõ 4 luồng: nhập kho, xuất kho, đóng hàng, xử lý hàng hoàn.
- Bàn giao cho ĐVVC có phân khu, để theo thùng/rổ, đối chiếu số lượng, quét mã, xử lý trường hợp đủ/chưa đủ đơn.
- Sản xuất hiện có nhánh gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, nhận hàng về kho, kiểm tra số lượng/chất lượng, phản hồi lỗi nhà máy trong 3–7 ngày.

---

## 3. Mục tiêu Phase 1

Phase 1 không nhằm làm toàn bộ ERP hoàn hảo ngay từ đầu. Phase 1 nhằm khóa các dòng chảy sống còn:

```text
Hàng hóa
→ Batch/QC
→ Kho
→ Đơn hàng
→ Giao hàng
→ Hàng hoàn
→ Đối soát
→ Báo cáo điều hành cơ bản
```

Mục tiêu chính:

1. Có một hệ thống web ERP nội bộ chạy được.
2. Quản lý dữ liệu gốc chuẩn.
3. Quản lý kho theo batch/lô/hạn dùng.
4. Ghi nhận stock movement bằng stock ledger bất biến.
5. Quản lý mua hàng, nhận hàng, QC đầu vào.
6. Quản lý sản xuất/gia công ngoài.
7. Quản lý bán hàng, đóng hàng, bàn giao ĐVVC.
8. Quản lý hàng hoàn và phân loại xử lý.
9. Có RBAC, approval, audit log.
10. Có dashboard và báo cáo vận hành cơ bản.
11. Có nền kỹ thuật để mở rộng Phase 2.

---

## 4. Phạm vi module Phase 1

### 4.1. In-scope

| Nhóm | Module | Mô tả |
|---|---|---|
| Nền tảng | Auth/RBAC | Đăng nhập, phân quyền, session, role, permission |
| Nền tảng | Approval | Quy trình duyệt cho chứng từ quan trọng |
| Nền tảng | Audit Log | Ghi lịch sử hành động, trước/sau, user/time/source |
| Dữ liệu | Master Data | SKU, nguyên liệu, kho, vị trí, nhà cung cấp, khách hàng, nhân viên, ĐVVC |
| Mua hàng | Purchase | PR/PO, nhận hàng, supplier basic |
| QC | Quality Control | QC đầu vào, QC batch/thành phẩm, pass/hold/fail |
| Kho | Inventory/WMS | Nhập/xuất/chuyển kho, batch, hạn dùng, tồn khả dụng |
| Kho | Stock Ledger | Ghi nhận biến động tồn bất biến |
| Kho | Daily Warehouse Board | Quản lý công việc kho theo ngày |
| Kho | Shift Closing | Kiểm kê và đối soát cuối ca/cuối ngày |
| Bán hàng | Sales Order | Tạo/duyệt đơn, giữ tồn, trạng thái đơn |
| Đóng hàng | Pick/Pack | Soạn hàng, đóng gói, scan/verify |
| Giao hàng | Shipping/Handover | Manifest, bàn giao ĐVVC, scan, xử lý thiếu đơn |
| Hàng hoàn | Returns | Nhận hàng hoàn, scan, kiểm tra, phân loại còn dùng/không dùng |
| Gia công | Subcontract Manufacturing | Đơn gia công, chuyển NVL/bao bì, duyệt mẫu, nhận hàng, claim nhà máy |
| Báo cáo | Basic BI | Dashboard vận hành, tồn kho, đơn hàng, QC, hàng hoàn |
| Tích hợp | Integration baseline | ĐVVC, website/marketplace import/export, barcode scanner, file storage |
| Kỹ thuật | Go Backend | Modular monolith, REST/OpenAPI, PostgreSQL |
| Kỹ thuật | Frontend | React/Next.js + TypeScript |

### 4.2. Out-of-scope Phase 1

Các hạng mục này chuyển Phase 2 hoặc phase sau:

- CRM nâng cao
- loyalty
- HRM/payroll đầy đủ
- KOL/KOC/Affiliate đầy đủ
- finance/accounting posting nâng cao
- BI nâng cao
- AI forecast
- mobile app native
- supplier portal
- dealer portal
- KOL portal
- full marketplace real-time integration nếu chưa đủ nguồn lực

---

## 5. Tech stack đã chốt

Vendor phải tuân thủ stack sau, trừ khi có đề xuất thay đổi được phê duyệt bằng văn bản.

| Layer | Stack |
|---|---|
| Backend | Go |
| Backend Architecture | Modular Monolith |
| API | REST + OpenAPI 3.1 |
| Database | PostgreSQL |
| Cache/Queue | Redis + NATS hoặc RabbitMQ |
| Frontend | React / Next.js + TypeScript |
| UI Library | Ant Design + ERP custom design tokens |
| Form | React Hook Form + Zod |
| API Client | OpenAPI generated client |
| File Storage | S3-compatible / MinIO |
| DevOps | Docker, CI/CD, staging/UAT/prod |
| Observability | Logging, monitoring, audit trail |

---

## 6. Tài liệu tham chiếu bắt buộc

Vendor phải đọc và xác nhận đã hiểu các file sau:

| STT | Tài liệu | Vai trò |
|---:|---|---|
| 01 | ERP Blueprint | Bức tranh tổng thể |
| 03 | PRD/SRS Phase 1 | Requirement nghiệp vụ |
| 04 | Permission/Approval Matrix | Quyền hạn và duyệt |
| 05 | Data Dictionary/Master Data | Chuẩn dữ liệu |
| 06 | Process Flow To-Be | Luồng thiết kế |
| 08 | Screen List/Wireframe | Màn hình và UX nghiệp vụ |
| 09 | UAT Test Scenarios | Nghiệm thu business |
| 11 | Technical Architecture Go Backend | Kiến trúc backend |
| 12 | Go Coding Standards | Chuẩn code Go |
| 13 | Go Module/Component Standards | Chuẩn module boundary |
| 14 | UI/UX Design System Standards | Chuẩn UI/UX |
| 15 | Frontend Architecture | Kiến trúc frontend |
| 16 | API Contract/OpenAPI Standards | Chuẩn API |
| 17 | Database Schema PostgreSQL Standards | Chuẩn DB |
| 18 | DevOps/CI-CD Standards | Chuẩn môi trường triển khai |
| 19 | Security/RBAC/Audit/Compliance | Bảo mật và audit |
| 20 | Current Workflow As-Is | Workflow thực tế |
| 21 | Gap Analysis/Decision Log | Quyết định chỉnh To-Be |
| 22 | Core Docs Revision v1.1 Change Log | Những điểm cần revise |
| 23 | Integration Spec | Chuẩn tích hợp |
| 24 | QA Test Strategy | Chiến lược QA/test automation |
| 25 | Product Backlog/Sprint Plan | Backlog và sprint plan |
| 26 | SOP/Training Manual | Hướng dẫn vận hành |
| 27 | Go-Live Runbook/Hypercare | Go-live |
| 28 | Risk/Incident Playbook | Xử lý sự cố |
| 29 | Operations Support Model | Mô hình support |
| 30 | Data Governance/Change Control | Quản trị dữ liệu và thay đổi |
| 31 | Phase 2 Scope | Tầm nhìn mở rộng |
| 32 | Master Document Index/Traceability | Mục lục và traceability |
| 33 | Core Docs v1.1 Update Pack | Gói cập nhật lõi |
| 34 | Sprint 0 Kickoff Plan | Kế hoạch kickoff kỹ thuật |

---

## 7. RFP: Thông tin vendor phải cung cấp

Vendor gửi proposal phải bao gồm các phần sau.

### 7.1. Company profile

- tên công ty
- năm thành lập
- quy mô team
- địa điểm
- pháp nhân
- kinh nghiệm ERP/WMS/OMS/manufacturing
- kinh nghiệm Go/React/PostgreSQL
- case study liên quan
- khách hàng tham chiếu nếu có

### 7.2. Proposed solution

Vendor cần mô tả:

- cách hiểu bài toán
- architecture đề xuất
- module delivery plan
- phương pháp triển khai
- rủi ro nhận diện được
- những assumption
- những dependency từ phía client
- đề xuất cải tiến nếu có

### 7.3. Team proposal

Vendor phải đưa team cụ thể:

| Role | Số lượng | Yêu cầu |
|---|---:|---|
| Project Manager | 1 | Có kinh nghiệm ERP hoặc hệ thống vận hành |
| Business Analyst | 1–2 | Hiểu process, data, UAT |
| Solution Architect / Tech Lead | 1 | Biết Go architecture, PostgreSQL, API, modular monolith |
| Backend Go Developer | 2+ | Go, PostgreSQL, transaction, API |
| Frontend Developer | 2+ | React/Next.js, TypeScript, form/table UX |
| QA Engineer | 1–2 | API/E2E/regression/UAT support |
| DevOps Engineer | 0.5–1 | CI/CD, Docker, deployment |
| UI/UX Designer | 0.5–1 | ERP form/table/workflow UX |

Vendor phải ghi rõ từng người:

- tên
- vai trò
- seniority
- kinh nghiệm
- allocation %
- thời gian tham gia
- có phải nhân sự full-time hay part-time

### 7.4. Delivery methodology

Vendor phải mô tả:

- Agile/Scrum hay hybrid
- sprint length
- demo cadence
- backlog management tool
- issue tracking tool
- communication channel
- weekly report format
- change request process
- QA/UAT process
- release process

### 7.5. Timeline

Vendor phải đề xuất timeline theo milestone:

```text
Sprint 0
MVP Internal Demo
UAT 1
UAT 2
Pilot Go-Live
Production Go-Live
Hypercare
```

### 7.6. Commercial proposal

Vendor phải tách chi phí theo:

- discovery/analysis
- UI/UX
- backend
- frontend
- QA
- DevOps
- project management
- integration
- data migration
- UAT support
- go-live/hypercare
- post-go-live support
- change request rate
- maintenance/support monthly fee

Không chấp nhận báo giá một dòng quá mơ hồ.

---

## 8. SOW Template

### 8.1. SOW Summary

```text
Tên dự án:
ERP Web Phase 1 cho công ty mỹ phẩm

Mục tiêu:
Xây dựng hệ thống ERP web nội bộ để quản lý dữ liệu gốc, mua hàng, QC, kho, batch/hạn dùng, sản xuất/gia công ngoài, đơn hàng, đóng hàng, bàn giao ĐVVC, hàng hoàn, audit, approval, báo cáo cơ bản.

Thời gian dự kiến:
[Vendor đề xuất]

Phương pháp:
Agile/hybrid, triển khai theo sprint, demo định kỳ, UAT trước go-live.
```

### 8.2. Deliverables bắt buộc

| Nhóm | Deliverable |
|---|---|
| Project | Project plan, sprint plan, weekly status report |
| UX | Wireframe chi tiết, clickable prototype nếu cần |
| Backend | Go backend source code |
| Frontend | Next.js frontend source code |
| API | OpenAPI spec |
| Database | Migration scripts, seed data, schema docs |
| Auth | Login, RBAC, session, audit |
| Module | Master Data, Purchase, QC, Inventory, Sales, Shipping, Returns, Subcontract Manufacturing |
| Reports | Dashboard/report cơ bản |
| QA | Test plan, test cases, bug reports, regression checklist |
| DevOps | CI/CD, environment setup, deployment guide |
| Training | User training support |
| UAT | UAT support và issue fixing |
| Go-Live | Go-live support/hypercare |
| Handover | Technical handover, source code, credentials handover, docs |

### 8.3. Acceptance criteria tổng

Một module chỉ được coi là nghiệm thu khi:

1. Chạy đúng flow được mô tả trong PRD/Process Flow.
2. Có phân quyền đúng.
3. Có audit log cho action quan trọng.
4. Có validation dữ liệu.
5. Có test case pass.
6. Có màn hình list/detail/create/update/action cần thiết.
7. Có API documented trong OpenAPI.
8. Có migration DB.
9. Có lỗi được ghi nhận và xử lý theo severity.
10. Được business user UAT pass.

### 8.4. Scope control

Vendor không được tự ý thay đổi:

- business flow
- trạng thái chứng từ
- schema quan trọng
- permission model
- stock ledger logic
- QC/batch status
- API contract
- tech stack
- module boundary

Mọi thay đổi phải đi qua Change Request.

---

## 9. Milestone và payment gợi ý

| Milestone | Deliverable | Payment % gợi ý |
|---|---|---:|
| M0 | Kickoff + Sprint 0 complete | 10% |
| M1 | Core foundation: auth/RBAC, master data, audit, OpenAPI, DB base | 15% |
| M2 | Purchase + QC + Inventory core | 20% |
| M3 | Sales + pick/pack + shipping handover | 20% |
| M4 | Returns + subcontract manufacturing + reports | 15% |
| M5 | UAT pass + fixes | 10% |
| M6 | Production go-live + hypercare complete | 10% |

Payment nên gắn với nghiệm thu thật, không gắn đơn thuần với “đã code xong”.

---

## 10. Vendor evaluation scorecard

Tổng điểm: 100.

| Nhóm tiêu chí | Trọng số |
|---|---:|
| Hiểu nghiệp vụ ERP/WMS/OMS/manufacturing | 20 |
| Năng lực Go/PostgreSQL/React/OpenAPI | 20 |
| Cách tiếp cận architecture/module/quality | 15 |
| Kinh nghiệm triển khai và go-live | 10 |
| Proposal clarity và risk thinking | 10 |
| Team thực tế tham gia | 10 |
| Giá và điều kiện thương mại | 10 |
| Support/handover dài hạn | 5 |

### 10.1. Thang điểm

| Điểm | Ý nghĩa |
|---:|---|
| 5 | Xuất sắc, có bằng chứng rõ |
| 4 | Tốt, phù hợp |
| 3 | Tạm chấp nhận, còn rủi ro |
| 2 | Yếu, nhiều điểm mơ hồ |
| 1 | Không đạt |
| 0 | Không cung cấp thông tin |

### 10.2. Red flags

Loại hoặc cảnh báo mạnh nếu vendor:

- không hiểu stock ledger
- đề xuất sửa tồn trực tiếp
- không có audit log nghiêm
- không hiểu batch/QC/hạn dùng
- không tách được physical stock và available stock
- bỏ qua hàng hoàn
- coi bàn giao ĐVVC chỉ là đổi status
- không hiểu gia công ngoài/chuyển NVL cho nhà máy
- không có OpenAPI contract rõ
- không có QA/UAT plan
- không có DevOps/release/rollback plan
- báo giá quá mơ hồ
- không commit nhân sự cụ thể
- né bàn giao source code hoặc tài liệu

---

## 11. Technical questionnaire cho vendor

Vendor phải trả lời các câu hỏi sau.

### 11.1. Backend Go

1. Vendor sẽ tổ chức modular monolith Go như thế nào?
2. Cách tách module boundary?
3. Handler/application/domain/repository được tách ra sao?
4. Cách quản lý transaction?
5. Cách xử lý idempotency?
6. Cách generate OpenAPI?
7. Cách quản lý error code?
8. Cách xử lý background jobs?
9. Cách implement outbox/event?
10. Cách test API và domain logic?

### 11.2. Database/PostgreSQL

1. Cách thiết kế stock ledger bất biến?
2. Cách đồng bộ stock balance?
3. Cách xử lý reservation?
4. Cách enforce batch/QC status?
5. Cách migration schema?
6. Cách rollback migration?
7. Cách index cho table lớn?
8. Cách backup/restore?
9. Cách kiểm tra data integrity?
10. Cách audit data change?

### 11.3. Frontend

1. Cách tổ chức module route trong Next.js?
2. Cách dùng OpenAPI generated client?
3. Cách quản lý form phức tạp?
4. Cách xử lý table/filter/sort/export?
5. Cách thiết kế scan UX?
6. Cách xử lý permission ở UI?
7. Cách xử lý loading/error/empty state?
8. Cách test frontend?
9. Cách handoff từ design system?
10. Cách đảm bảo user thao tác nhanh trong kho?

### 11.4. Security

1. Auth/session dùng gì?
2. Có hỗ trợ MFA không?
3. RBAC/field-level permission xử lý thế nào?
4. Audit log cho action nhạy cảm ra sao?
5. Export dữ liệu nhạy cảm kiểm soát thế nào?
6. Break-glass access xử lý thế nào?
7. Log có che dữ liệu nhạy cảm không?
8. Có security test không?
9. Secret management thế nào?
10. Incident response thế nào?

---

## 12. Demo script bắt buộc khi chọn vendor

Vendor shortlist phải demo hoặc mô tả chi tiết cách làm các flow sau.

### 12.1. Flow 1: Nhập kho + QC

```text
Tạo PO
→ nhận hàng
→ nhập batch/hạn dùng
→ QC hold
→ QC pass
→ chuyển vào available stock
→ audit log
→ stock ledger movement
```

### 12.2. Flow 2: Đơn hàng + reserve + pick/pack

```text
Tạo sales order
→ check available stock
→ reserve stock
→ tạo pick task
→ scan hàng
→ đóng hàng
→ packed
```

### 12.3. Flow 3: Bàn giao ĐVVC

```text
Tạo manifest
→ phân khu để hàng
→ đối chiếu số lượng đơn
→ scan từng đơn
→ nếu đủ thì xác nhận bàn giao
→ nếu thiếu thì cảnh báo và xử lý
→ ghi audit log
```

### 12.4. Flow 4: Hàng hoàn

```text
Nhận hàng từ shipper
→ đưa vào khu vực hàng hoàn
→ scan hàng hoàn
→ kiểm tra tình trạng
→ còn dùng thì chuyển kho
→ không dùng thì chuyển lab/kho hỏng
→ lập phiếu nhập/điều chỉnh phù hợp
```

### 12.5. Flow 5: Gia công ngoài

```text
Tạo subcontract order
→ xác nhận số lượng/quy cách/mẫu mã
→ cọc đơn
→ chuyển NVL/bao bì sang nhà máy
→ duyệt mẫu
→ sản xuất hàng loạt
→ nhận hàng về kho
→ kiểm tra số lượng/chất lượng
→ nhận hàng hoặc claim nhà máy trong 3–7 ngày
→ thanh toán lần cuối
```

---

## 13. Contract/SOW guardrails

### 13.1. Source code ownership

Client phải sở hữu:

- source code backend
- source code frontend
- database migration
- OpenAPI spec
- design assets
- deployment scripts
- test cases
- documentation

Vendor không được giữ source code làm con tin.

### 13.2. Credential handover

Trước go-live và khi kết thúc dự án, vendor phải bàn giao:

- repository access
- deployment credentials
- environment variables checklist
- domain/DNS nếu có
- cloud credentials nếu vendor giữ
- database credentials theo cơ chế bảo mật
- admin account handover
- backup/restore instruction

### 13.3. Warranty

Gợi ý:

```text
Warranty: 30–60 ngày sau production go-live
Bao gồm: bug fixing cho lỗi nằm trong scope đã nghiệm thu
Không bao gồm: feature mới/change request
```

### 13.4. Change request

Mọi thay đổi phải có:

```text
CR ID
Mô tả
Lý do
Ảnh hưởng scope/timeline/cost
Ảnh hưởng tài liệu
Người duyệt
Ngày hiệu lực
```

### 13.5. Non-functional requirements

Vendor phải đáp ứng:

- response time hợp lý cho thao tác kho/sales
- audit log cho action nhạy cảm
- backup/restore
- staging/UAT/prod separation
- rollback plan
- security baseline
- data validation
- export control
- log/monitoring

---

## 14. Vendor proposal response template

Vendor nên trả lời theo template này.

```text
1. Executive Summary
2. Understanding of Business
3. Proposed Solution
4. Architecture Approach
5. Scope Confirmation
6. Assumptions
7. Exclusions
8. Team Structure
9. Delivery Plan
10. Timeline
11. QA/UAT Plan
12. DevOps/Deployment Plan
13. Security/Compliance Approach
14. Data Migration Approach
15. Support/Hypercare
16. Commercial Proposal
17. Risks and Mitigations
18. Relevant Case Studies
19. Appendices
```

---

## 15. SOW acceptance checklist

Trước khi ký SOW, phải check:

| Checklist | Trạng thái |
|---|---|
| Scope module Phase 1 rõ |
| Out-of-scope rõ |
| Deliverables rõ |
| Timeline rõ |
| Milestone/payment rõ |
| Team allocation rõ |
| UAT/acceptance rõ |
| Change request rule rõ |
| Source code ownership rõ |
| Warranty rõ |
| Support/hypercare rõ |
| Security/audit/RBAC rõ |
| Data migration rõ |
| DevOps/CI-CD rõ |
| Documentation/handover rõ |
| Không có điều khoản mơ hồ gây lock-in |

---

## 16. Recommended vendor shortlist process

### Bước 1: Longlist

Chọn 5–7 vendor có khả năng:

- Go backend
- React/Next.js frontend
- ERP/WMS/OMS hoặc hệ thống vận hành
- PostgreSQL
- CI/CD
- QA/UAT

### Bước 2: RFP

Gửi bộ tài liệu:

```text
03, 04, 05, 06, 08, 11–18, 20–25, 32, 34, 35
```

Tùy mức bảo mật, có thể chưa gửi toàn bộ file có dữ liệu nhạy cảm.

### Bước 3: Q&A

Cho vendor hỏi trong 3–5 ngày.

Tất cả câu hỏi/trả lời phải ghi thành Q&A log, không trả lời riêng lẻ qua chat rồi thất lạc.

### Bước 4: Proposal

Vendor nộp proposal.

### Bước 5: Technical interview

Phỏng vấn:

- PM
- BA
- Tech Lead
- Backend Go Lead
- Frontend Lead
- QA Lead
- DevOps

### Bước 6: Demo/Case review

Vendor phải walkthrough một trong các flow bắt buộc.

### Bước 7: Scorecard

Chấm theo bảng 100 điểm.

### Bước 8: Reference check

Gọi hỏi khách hàng cũ nếu có.

### Bước 9: SOW negotiation

Đàm phán scope, milestone, payment, warranty, source code, change request.

### Bước 10: Kickoff

Dùng file 34 Sprint 0 để kickoff.

---

## 17. Câu hỏi vendor phải hỏi lại client

Vendor tốt không chỉ trả lời. Vendor tốt phải biết hỏi.

Các câu vendor nên hỏi:

1. Có bao nhiêu kho và khu vực kho?
2. Có bao nhiêu người dùng Phase 1?
3. ĐVVC nào cần tích hợp trước?
4. Đơn hàng đến từ kênh nào trước?
5. Barcode hiện có format gì?
6. Có máy quét mã loại nào?
7. Có phần mềm kế toán đang dùng không?
8. Dữ liệu hiện tại nằm ở Excel, phần mềm cũ hay giấy?
9. Có bao nhiêu SKU/batch/đơn/ngày?
10. Quy trình kiểm kê thực tế chi tiết ra sao?
11. Khi hàng hoàn về, ai quyết định còn dùng/không dùng?
12. Khi nhà máy giao lỗi, ai lập claim?
13. Ai là business owner có quyền chốt flow?
14. Ai là super user từng phòng?
15. Go-live có chạy song song hệ thống cũ không?

Vendor không hỏi những câu này có thể chưa hiểu đủ sâu.

---

## 18. Rủi ro khi chọn vendor sai

| Rủi ro | Hậu quả |
|---|---|
| Vendor không hiểu kho/batch | Tồn kho sai, không truy xuất được lô |
| Vendor không hiểu audit | Không truy trách nhiệm khi sai dữ liệu |
| Vendor không hiểu WMS scan | Bàn giao ĐVVC lệch đơn |
| Vendor không hiểu hàng hoàn | Hàng hoàn quay lại kho sai trạng thái |
| Vendor không hiểu gia công ngoài | Không kiểm soát NVL/bao bì gửi nhà máy |
| Vendor yếu QA | Go-live lỗi nặng |
| Vendor yếu DevOps | Deploy/rollback rối |
| Vendor thiếu PM/BA | Scope trôi |
| Vendor báo giá quá thấp | Dễ đội chi phí bằng CR |
| Vendor không bàn giao source | Bị lock-in |

---

## 19. Mô hình hợp tác khuyến nghị

Khuyến nghị không giao khoán mù.

Nên dùng mô hình:

```text
Client giữ Product Ownership
Vendor chịu trách nhiệm Delivery
Hai bên cùng chốt scope theo sprint
```

Phía client cần có:

- Product Owner
- Business Owner kho
- Business Owner sản xuất/gia công
- Business Owner sales/CSKH
- Finance reviewer
- ERP Admin tương lai
- người có quyền chốt quy trình

Phía vendor cần có:

- PM
- BA
- Tech Lead
- Go backend lead
- Frontend lead
- QA lead
- DevOps

---

## 20. Final recommendation

Trước khi chọn vendor, client nên yêu cầu vendor làm một mini discovery hoặc technical workshop 2–5 ngày.

Workshop này phải ra được:

1. Confirmation of Phase 1 scope.
2. Architecture confirmation.
3. High-level sprint plan.
4. Key risks.
5. Integration assumptions.
6. Team allocation.
7. Fixed or capped budget range.
8. Go/No-Go recommendation.

Một câu chốt:

> Đừng chọn vendor chỉ vì giá thấp. Chọn vendor hiểu rủi ro nhất, hỏi câu sắc nhất, và chứng minh được họ có thể giữ dữ liệu hàng-tiền-kho không sai.

---

## 21. Phụ lục: Email mời thầu mẫu

```text
Subject: RFP - Web ERP Phase 1 cho công ty mỹ phẩm

Chào [Vendor],

Chúng tôi đang tìm đối tác triển khai Web ERP Phase 1 cho doanh nghiệp mỹ phẩm, bao gồm các phân hệ: Master Data, Purchase, QC, Inventory/WMS, Sales Order, Pick/Pack, Shipping Handover, Returns, Subcontract Manufacturing, RBAC, Audit Log, Reporting cơ bản.

Tech stack định hướng:
- Backend: Go
- Frontend: React/Next.js + TypeScript
- Database: PostgreSQL
- API: REST + OpenAPI
- Architecture: Modular Monolith

Vui lòng gửi proposal bao gồm:
1. Company profile
2. Understanding of scope
3. Proposed solution
4. Team allocation
5. Timeline
6. Commercial proposal
7. QA/UAT approach
8. DevOps/deployment approach
9. Risks and assumptions
10. Relevant case studies

Deadline gửi proposal: [ngày]
Q&A session: [ngày]
Demo/interview: [ngày]

Trân trọng,
[Client]
```

---

## 22. Phụ lục: Vendor scoring template

| Vendor | Business Fit /20 | Tech Fit /20 | Architecture /15 | Delivery /10 | Proposal /10 | Team /10 | Cost /10 | Support /5 | Total /100 | Note |
|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|
| Vendor A |  |  |  |  |  |  |  |  |  |  |
| Vendor B |  |  |  |  |  |  |  |  |  |  |
| Vendor C |  |  |  |  |  |  |  |  |  |  |

---

## 23. Phụ lục: SOW redline checklist

Trước khi ký, rà lại:

```text
[ ] Có ghi rõ source code thuộc client
[ ] Có ghi rõ vendor không được tái sử dụng dữ liệu client
[ ] Có ghi rõ payment theo milestone nghiệm thu
[ ] Có ghi rõ CR process
[ ] Có ghi rõ warranty
[ ] Có ghi rõ support/hypercare
[ ] Có ghi rõ handover
[ ] Có ghi rõ môi trường staging/UAT/prod
[ ] Có ghi rõ security/audit/RBAC
[ ] Có ghi rõ dữ liệu nhạy cảm
[ ] Có ghi rõ termination/exit plan
[ ] Có ghi rõ IP ownership
[ ] Có ghi rõ confidentiality
[ ] Có ghi rõ SLA nếu support dài hạn
```

---

## 24. Kết luận

File này là cầu nối giữa bộ tài liệu ERP và quá trình chọn đội triển khai.

Nếu dùng đúng, nó giúp:

- vendor hiểu đúng bài toán
- client so sánh vendor công bằng
- scope không bị trôi
- chi phí không bị mơ hồ
- nghiệm thu không cảm tính
- source code và dữ liệu không bị lock-in
- dự án chuyển từ thiết kế sang thi công có kiểm soát

Tài liệu tiếp theo nên làm sau file này:

```text
36_ERP_Executive_Summary_Board_Presentation_Phase1_MyPham_v1.md
```

Hoặc nếu đã chọn vendor:

```text
36_ERP_SOW_Final_Draft_Phase1_MyPham_v1.md
```
