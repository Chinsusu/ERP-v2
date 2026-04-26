# 25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Phiên bản:** v1.0  
**Trạng thái:** Draft dùng để review với CEO / Product Owner / PM / BA / Tech Lead / QA Lead / Vendor  
**Backend:** Go  
**Frontend:** React / Next.js + TypeScript  
**Database:** PostgreSQL  
**API:** REST + OpenAPI 3.1  
**Kiến trúc:** Modular Monolith  

---

## 1. Mục đích tài liệu

Tài liệu này biến bộ PRD, Process Flow, Data Dictionary, Screen List, Technical Architecture và các tài liệu workflow thực tế thành **Product Backlog + Sprint Plan** để đội dự án có thể bắt đầu triển khai.

Nói dễ hiểu:

```text
PRD/SRS nói hệ thống cần làm gì.
Backlog/Sprint Plan nói đội dev phải build cái gì, theo thứ tự nào, nghiệm thu bằng tiêu chí nào.
```

Tài liệu này dùng để:

- chia toàn bộ Phase 1 thành các epic, user story, task lớn;
- xác định ưu tiên build theo rủi ro và giá trị vận hành;
- khóa dependency giữa các module;
- giúp PM lập sprint;
- giúp BA viết chi tiết user story;
- giúp UI/UX thiết kế đúng thứ tự;
- giúp backend/frontend/QA biết phải làm gì trước;
- tránh tình trạng dev build màn hình đẹp nhưng chưa khóa nghiệp vụ sống còn như tồn kho, batch, QC, bàn giao ĐVVC, hàng hoàn, gia công ngoài.

Một câu chốt: **Backlog này không phải danh sách việc cho đẹp. Nó là bản đồ thi công để đưa ERP Phase 1 từ giấy ra sản phẩm chạy thật.**

---

## 2. Nguồn đầu vào

Tài liệu này dựa trên bộ tài liệu ERP đã có:

- `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
- `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`
- `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1.md`
- `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`
- `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`
- `10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1.md`
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
- `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md`

Và 4 tài liệu workflow thực tế:

- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

Các workflow thực tế phải được phản ánh trực tiếp vào backlog:

- kho tiếp nhận đơn hàng trong ngày;
- thực hiện xuất/nhập theo bảng nội quy;
- soạn hàng và đóng gói;
- sắp xếp, tối ưu vị trí kho;
- kiểm kê tồn kho cuối ngày;
- đối soát số liệu và báo cáo quản lý;
- kết thúc ca;
- nhập kho có chứng từ giao hàng, kiểm số lượng, bao bì, lô;
- xuất kho có phiếu xuất, đối chiếu thực tế, ký bàn giao;
- đóng hàng theo đơn/kênh/ĐVVC;
- bàn giao ĐVVC có phân khu, theo thùng/rổ, đối chiếu số lượng, quét mã;
- xử lý thiếu đơn khi bàn giao;
- nhận hàng hoàn từ shipper, quét hàng hoàn, kiểm tra tình trạng, phân loại còn dùng/không dùng;
- sản xuất/gia công ngoài: lên đơn nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, duyệt mẫu, sản xuất hàng loạt, nhận hàng về kho, kiểm số lượng/chất lượng, báo lỗi nhà máy trong 3–7 ngày, thanh toán lần cuối.

---

## 3. Phạm vi Phase 1 theo backlog

Phase 1 tập trung vào lõi vận hành tạo tiền và giữ hàng:

```text
Foundation
→ Master Data
→ Purchase / Inbound
→ QC / Batch
→ Inventory / Stock Ledger
→ Subcontract Manufacturing
→ Sales Order
→ Pick / Pack / Ship
→ Carrier Handover
→ Returns
→ Finance Lite / Reconciliation
→ Dashboard / Report
→ Integration / Migration / UAT
```

### 3.1. Trong Phase 1

Bao gồm:

- đăng nhập, phân quyền, audit log;
- dữ liệu gốc: SKU, nguyên liệu, bao bì, kho, vị trí, NCC, khách hàng, ĐVVC, nhà máy gia công;
- batch/lô, hạn dùng, QC status;
- mua hàng, PO, nhận hàng, QC đầu vào, nhập kho;
- stock ledger bất biến;
- tồn vật lý, tồn khả dụng, tồn giữ, tồn hold QC;
- cycle count, kiểm kê, đối soát cuối ca;
- sản xuất/gia công ngoài cơ bản;
- chuyển NVL/bao bì cho nhà máy;
- duyệt mẫu, nhận hàng gia công, nghiệm thu, claim nhà máy;
- sales order, giữ tồn, xác nhận đơn;
- pick, pack, kiểm tra đóng hàng;
- manifest, scan bàn giao ĐVVC;
- xử lý thiếu đơn khi bàn giao;
- hàng hoàn, kiểm hàng hoàn, disposition;
- công nợ/thu chi/COD cơ bản ở mức Phase 1;
- báo cáo vận hành cơ bản;
- API/OpenAPI, DB schema, DevOps, QA automation cơ bản;
- migration dữ liệu ban đầu.

### 3.2. Không làm sâu trong Phase 1

Để tránh nhồi scope, Phase 1 không làm sâu:

- HRM đầy đủ;
- CRM nâng cao;
- KOL/Affiliate đầy đủ;
- payroll đầy đủ;
- full accounting posting chuẩn kế toán;
- BI nâng cao;
- AI forecast;
- mobile app native;
- supplier portal/KOL portal/dealer portal;
- manufacturing MES nội bộ quá chi tiết nếu thực tế vẫn chủ yếu gia công ngoài.

Các phần này đưa vào Parking Lot / Phase 2.

---

## 4. Nguyên tắc ưu tiên backlog

### 4.1. Ưu tiên theo rủi ro thật

Build trước những thứ nếu sai sẽ làm công ty mất tiền hoặc mất kiểm soát:

```text
Tồn kho sai
Batch/QC sai
Xuất nhầm hàng
Bàn giao thiếu đơn
Hàng hoàn nhập sai
Gia công ngoài thiếu chứng từ
Công nợ/COD lệch
User không có quyền nhưng vẫn thao tác được
```

### 4.2. Ưu tiên theo dependency

Không build sales order sâu khi chưa có:

- SKU;
- kho;
- batch;
- stock ledger;
- available stock;
- reservation;
- QC status.

Không build shipping handover khi chưa có:

- packed order;
- shipment;
- carrier manifest;
- scan event;
- exception handling.

Không build returns khi chưa có:

- order/shipment source;
- return receiving;
- condition check;
- stock disposition;
- audit log.

### 4.3. Ưu tiên theo workflow thực tế

Kho hiện đang có nhịp vận hành theo ngày. Vì vậy backlog phải có:

- Warehouse Daily Board;
- Packing Task Board;
- Carrier Manifest Board;
- Shift Closing / End-of-Day Reconciliation;
- Cycle Count / Daily Count;
- Exception List.

Đây không phải tính năng phụ. Đây là cách kho đang sống mỗi ngày.

---

## 5. Thang ưu tiên

| Priority | Ý nghĩa | Ví dụ |
|---|---|---|
| P0 | Bắt buộc để go-live Phase 1 | Stock ledger, batch/QC, sales order, pick/pack/ship, return receiving, RBAC |
| P1 | Cần có để vận hành mượt | dashboard, scan optimization, report export, alerts |
| P2 | Có thì tốt, có thể lùi | nâng cao UX, bulk action nâng cao, automation sâu |
| P3 | Parking Lot / Phase 2 | HRM sâu, CRM sâu, KOL sâu, AI forecast |

---

## 6. Estimate scale

Backlog dùng thang ước lượng đơn giản:

| Size | Ý nghĩa | Gợi ý effort |
|---|---|---|
| S | Nhỏ, rõ, ít dependency | 0.5–1 ngày dev |
| M | Vừa, có vài rule nghiệp vụ | 2–4 ngày dev |
| L | Lớn, nhiều state/rule/API/UI | 1–2 tuần |
| XL | Rất lớn, nên tách nhỏ | >2 tuần |

Nguyên tắc: **story XL không được đưa vào sprint nguyên cục. Phải tách.**

---

## 7. Giả định đội triển khai

Tài liệu này giả định một squad tối thiểu:

- 01 Product Owner / Business Owner;
- 01 Project Manager / Scrum Master;
- 01 Business Analyst;
- 01 Solution Architect / Tech Lead;
- 02 Backend Go Developer;
- 02 Frontend Developer;
- 01 QA Engineer;
- 01 UI/UX Designer;
- 01 DevOps part-time;
- key users từ kho, sản xuất/gia công, QA/QC, sale, finance.

Nếu đội nhỏ hơn, timeline phải kéo dài. Nếu đội lớn hơn nhưng thiếu BA/PO chốt nghiệp vụ, vẫn không nhanh hơn nhiều.

---

## 8. Epic tổng thể Phase 1

| Epic Code | Epic | Mục tiêu | Priority |
|---|---|---|---|
| E00 | Project Foundation | Setup repo, môi trường, CI/CD, OpenAPI, base app | P0 |
| E01 | Auth / RBAC / Audit | Đăng nhập, quyền, audit log, session | P0 |
| E02 | Master Data | Dữ liệu gốc SKU, kho, NCC, khách, batch config | P0 |
| E03 | Workflow / Approval | Trạng thái chứng từ, phê duyệt, comment, attachment | P0 |
| E04 | Purchase / Inbound | PR/PO, nhận hàng, chứng từ giao hàng | P0 |
| E05 | QC / Batch Release | QC đầu vào, hold/pass/fail, release batch | P0 |
| E06 | Inventory / WMS | Stock ledger, movement, tồn, kiểm kê, đối soát cuối ca | P0 |
| E07 | Subcontract Manufacturing | Gia công ngoài, chuyển NVL/bao bì, duyệt mẫu, nhận hàng | P0 |
| E08 | Sales Order / OMS | Đơn hàng, giữ tồn, giá/chiết khấu cơ bản | P0 |
| E09 | Pick / Pack / Ship / Handover | Soạn hàng, đóng hàng, manifest, scan ĐVVC | P0 |
| E10 | Returns | Hàng hoàn, kiểm hàng, phân loại, nhập lại/hỏng/lab | P0 |
| E11 | Finance Lite / Reconciliation | COD, thu chi cơ bản, công nợ, đối soát | P1 |
| E12 | Report / Dashboard | Dashboard vận hành, tồn, đơn, QC, kho, shipment | P1 |
| E13 | Integration | ĐVVC, website/sàn/POS import, scanner, accounting export | P1 |
| E14 | Data Migration | Import master/opening stock/opening AR/AP | P0 |
| E15 | QA / UAT / Release | Test, UAT, release, bug fixing, training handoff | P0 |
| E99 | Parking Lot Phase 2 | CRM, HRM, KOL, finance nâng cao | P3 |

---

# 9. Chi tiết Product Backlog theo Epic

## E00. Project Foundation

### E00-S01. Khởi tạo repository và cấu trúc dự án

**Vai trò:** Tech Lead  
**Mục tiêu:** Có codebase chuẩn cho backend Go, frontend Next.js, OpenAPI, migration, docs.  
**Priority:** P0  
**Size:** M  
**Dependency:** Không có  

**Acceptance Criteria:**

- Có repo hoặc monorepo được thống nhất.
- Có thư mục backend Go theo chuẩn `cmd/`, `internal/`, `migrations/`, `api/`.
- Có frontend Next.js theo chuẩn module route.
- Có thư mục OpenAPI.
- Có README local setup.
- Có makefile/taskfile cơ bản.
- Có Docker Compose local cho PostgreSQL, Redis, MinIO nếu cần.

### E00-S02. Setup môi trường Local / Dev / Staging / UAT

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Có cấu hình env tách biệt.
- Không hardcode secret.
- Có seed data cơ bản cho role/user/module.
- Staging/UAT có thể deploy độc lập.
- Có hướng dẫn reset local DB.

### E00-S03. CI pipeline cơ bản

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Backend chạy lint/test/build.
- Frontend chạy typecheck/lint/build.
- Migration kiểm tra được syntax.
- OpenAPI validate được.
- Pull request fail nếu quality gate fail.

### E00-S04. Base layout frontend ERP shell

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Có login page.
- Có app shell: sidebar, header, user menu, breadcrumb.
- Có layout list/detail/create/edit chuẩn.
- Có route guard theo auth.
- Có placeholder menu theo module Phase 1.

---

## E01. Auth / RBAC / Audit

### E01-S01. Đăng nhập / đăng xuất / session

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- User đăng nhập bằng email/username + password.
- Token/session hết hạn theo cấu hình.
- Đăng xuất invalidate session.
- Không cho truy cập route/API nếu chưa đăng nhập.
- Audit login success/fail.

### E01-S02. Role và permission

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Có role: Admin, CEO, Warehouse, Warehouse Manager, QA/QC, Purchasing, Production/Subcontract, Sales, Shipping, Finance, CSKH.
- Permission theo action: view/create/update/submit/approve/cancel/export.
- API kiểm permission server-side.
- Frontend ẩn/disable action theo permission.
- Có seed permission Phase 1.

### E01-S03. Field-level permission cơ bản

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Giá vốn/cost chỉ role được phép mới thấy.
- Dữ liệu nhạy cảm finance không hiển thị cho kho/sales nếu không có quyền.
- API không trả field nhạy cảm nếu user không có quyền.
- Export cũng phải tôn trọng field-level permission.

### E01-S04. Audit log chuẩn

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Ghi lại actor, action, module, object, timestamp, before/after khi cần.
- Audit cho các action: tạo/sửa/hủy/duyệt, đổi QC status, stock adjustment, handover, return disposition.
- Audit log read-only.
- Có màn hình audit tab trong detail page.

### E01-S05. Break-glass access

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Admin có thể kích hoạt quyền khẩn cấp có lý do.
- Mỗi action break-glass phải được audit.
- Có report break-glass usage.
- Không dùng break-glass cho thao tác thường ngày.

---

## E02. Master Data

### E02-S01. Master SKU / thành phẩm

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo/sửa/xem SKU.
- SKU có mã, tên, barcode, brand, category, unit, shelf life, status.
- Không cho trùng SKU code/barcode.
- SKU inactive không cho dùng trong đơn mới.
- Có import/export cơ bản.

### E02-S02. Master nguyên vật liệu / bao bì

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Quản lý raw material, packaging, label, box, phụ liệu.
- Có mã, đơn vị tính, quy đổi, supplier default, QC required, storage rule.
- Có batch/expiry required flag.
- Có import/export.

### E02-S03. Master kho / vị trí kho

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Tạo kho: kho nguyên liệu, kho bao bì, kho thành phẩm, kho hàng hoàn, kho hỏng/lab, kho ĐVVC staging.
- Tạo vị trí/bin/shelf.
- Có loại khu: available, QC hold, packing, return, damaged, subcontract staging.
- Không cho xóa kho/vị trí đã có giao dịch.

### E02-S04. Master nhà cung cấp / nhà máy gia công

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Tạo NCC.
- Phân loại supplier/factory/subcontractor/carrier.
- Lưu thông tin liên hệ, payment terms, lead time, status.
- Factory có flag hỗ trợ gia công ngoài.
- Supplier inactive không cho tạo PO mới.

### E02-S05. Master khách hàng / kênh bán / ĐVVC

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Quản lý khách hàng B2B/B2C cơ bản.
- Quản lý kênh bán: website, sàn, POS, B2B, social.
- Quản lý ĐVVC/carrier.
- Carrier có mã, tên, SLA, tracking format nếu có.

### E02-S06. Batch numbering rule

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Cấu hình format batch theo sản phẩm/kho/source.
- Batch có ngày sản xuất/hạn dùng/QC status/source document.
- Không cho batch trùng trong phạm vi rule.
- Batch mặc định HOLD khi chưa QC pass nếu item yêu cầu QC.

---

## E03. Workflow / Approval

### E03-S01. State machine chung cho chứng từ

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Có trạng thái Draft/Submitted/Approved/Rejected/Cancelled/Closed cho chứng từ cần duyệt.
- Không cho action trái trạng thái.
- Có reason khi reject/cancel.
- Có audit state transition.

### E03-S02. Approval workflow cơ bản

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Duyệt PR/PO, stock adjustment, QC release override, discount vượt ngưỡng nếu có.
- Approval theo role.
- User tạo không tự duyệt nếu rule cấm.
- Có inbox phê duyệt.
- Có comment/attachment trong quy trình duyệt.

### E03-S03. Attachment chuẩn

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Cho upload chứng từ giao hàng, phiếu nhập/xuất, COA, biên bản bàn giao, hình ảnh hàng hoàn, mẫu duyệt.
- File lưu vào S3/MinIO.
- File gắn với document source.
- Có quyền xem/download theo permission.

---

## E04. Purchase / Inbound

### E04-S01. Purchase Request

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Tạo PR với item, qty, need date, reason.
- Submit/approve/reject.
- PR approved có thể convert sang PO.
- Không cho sửa line sau khi submitted nếu không có quyền.

### E04-S02. Purchase Order

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo PO từ PR hoặc tạo trực tiếp theo quyền.
- PO có supplier, items, price, qty, expected date, payment terms.
- PO có trạng thái Draft/Submitted/Approved/Sent/Partially Received/Received/Closed/Cancelled.
- Không nhận hàng vượt PO nếu không có override.
- Có attachment báo giá/chứng từ.

### E04-S03. Inbound Receiving / GRN

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Nhận hàng theo PO.
- Ghi nhận số lượng thực nhận, bao bì, batch/lot, expiry.
- Nếu item yêu cầu QC thì nhập vào trạng thái QC HOLD.
- Nếu không đạt cơ bản thì tạo exception/reject line.
- Có phiếu nhận hàng và audit.
- Hỗ trợ nhận một phần.

### E04-S04. Receiving exception

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Ghi nhận thiếu hàng, thừa hàng, hư bao bì, sai batch, sai hạn dùng.
- Exception phải có reason, ảnh, người ghi nhận.
- Exception có trạng thái open/resolved.
- PO/GRN hiển thị exception.

---

## E05. QC / Batch Release

### E05-S01. QC Inspection đầu vào

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo QC task khi GRN có item yêu cầu QC.
- QC nhập kết quả pass/fail/hold.
- Có checklist cơ bản.
- Có attachment COA/hình ảnh nếu cần.
- QC pass chuyển batch/stock từ HOLD sang AVAILABLE.
- QC fail chuyển về rejected/quarantine và không cho bán/sản xuất.

### E05-S02. Batch status engine

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Batch có status HOLD/PASS/FAIL/QUARANTINE/EXPIRED.
- Batch HOLD/FAIL không được reserve/pick/ship.
- Batch status change phải audit.
- Chỉ QA/QC hoặc role được phép mới đổi QC status.

### E05-S03. QC thành phẩm/gia công nhận về

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Khi nhận hàng gia công về kho, tạo QC task nghiệm thu.
- Ghi số lượng đạt/không đạt.
- Hàng đạt nhập available sau QC pass.
- Hàng không đạt tạo claim nhà máy.
- Có deadline phản hồi lỗi nhà máy trong 3–7 ngày theo workflow thực tế.

### E05-S04. QC complaint link cơ bản

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Cho ghi nhận complaint liên quan order/batch.
- Batch có tab complaint.
- Nếu complaint nghiêm trọng, QA có thể set HOLD batch.

---

## E06. Inventory / WMS

### E06-S01. Immutable stock ledger

**Priority:** P0  
**Size:** XL → tách nhỏ  

**Acceptance Criteria:**

- Mọi thay đổi tồn phải tạo stock movement.
- Không sửa trực tiếp stock balance.
- Movement có source document, item, batch, warehouse, location, qty, direction, actor.
- Có reversal/adjustment thay vì xóa.
- Balance được tính/cập nhật từ ledger.

### E06-S02. Stock balance và available stock

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Hiển thị tồn vật lý, tồn hold QC, tồn reserved, tồn available.
- Công thức available được áp dụng nhất quán.
- Tồn theo item/batch/warehouse/location.
- Không cho bán vượt available.

### E06-S03. Manual stock adjustment có duyệt

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo adjustment với reason.
- Adjustment phải được duyệt theo rule.
- Adjustment tạo stock movement.
- Có audit before/after.
- Không cho adjustment batch/QC nhạy cảm nếu không đủ quyền.

### E06-S04. Internal transfer

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo transfer giữa kho/vị trí.
- Có trạng thái Draft/Submitted/Approved/In Transit/Received/Cancelled.
- Transfer tạo movement out/in đúng lúc.
- Hỗ trợ chuyển vào khu packing, return, damaged, subcontract staging.

### E06-S05. Cycle count / kiểm kê

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo phiếu kiểm kê theo kho/vị trí/item.
- Nhập số đếm thực tế.
- Hệ thống tính chênh lệch.
- Chênh lệch cần adjustment approval.
- Có report kiểm kê.

### E06-S06. Warehouse Daily Board

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Hiển thị đơn cần xử lý trong ngày.
- Hiển thị task nhập/xuất/đóng hàng/hàng hoàn/kiểm kê.
- Hiển thị exception chưa xử lý.
- Có filter theo kho, ca, carrier, trạng thái.
- Gắn với workflow công việc hằng ngày của kho.

### E06-S07. Shift Closing / End-of-Day Reconciliation

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Kho có màn hình chốt ca/chốt ngày.
- Kiểm tra các task chưa đóng.
- Kiểm tra đơn packed chưa handover.
- Kiểm tra return chưa disposition.
- Kiểm tra count/stock exception.
- Tạo biên bản/chứng từ chốt ca.
- Sau khi chốt ca, giao dịch trong ca bị khóa theo rule.

### E06-S08. Near-expiry and FEFO alert

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Hiển thị hàng cận date.
- Gợi ý xuất theo FEFO.
- Cảnh báo khi pick batch không theo FEFO nếu rule áp dụng.

---

## E07. Subcontract Manufacturing / Gia công ngoài

### E07-S01. Factory/Subcontract Order

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo đơn gia công với nhà máy.
- Ghi số lượng, quy cách, mẫu mã, ngày dự kiến sản xuất/nhận hàng.
- Có trạng thái Draft/Confirmed/Deposit Paid/Materials Transferred/Sample Approved/In Production/Delivered/QC Accepted/QC Rejected/Closed.
- Có attachment hợp đồng/báo giá/chứng từ.

### E07-S02. Deposit tracking cơ bản

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Ghi nhận khoản cọc đơn gia công.
- Gắn với factory order.
- Hiển thị còn phải thanh toán lần cuối.
- Không cần full accounting posting Phase 1.

### E07-S03. Chuyển NVL/bao bì cho nhà máy

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo material transfer cho factory order.
- Xuất kho NVL/bao bì sang trạng thái at subcontractor hoặc subcontract staging.
- Có biên bản bàn giao.
- Ghi batch/qty của NVL/bao bì chuyển đi.
- Có movement trong stock ledger.

### E07-S04. Sample approval

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo sample task.
- Lưu hình ảnh/mẫu/ghi chú.
- Có trạng thái Pending/Approved/Rejected/Rework.
- Chỉ khi sample approved mới cho chuyển sang production hàng loạt nếu rule bật.
- Có lưu mẫu/reference.

### E07-S05. Receive subcontracted finished goods

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Nhận thành phẩm từ nhà máy về kho.
- Kiểm số lượng/quy cách/chất lượng cơ bản.
- Tạo batch thành phẩm.
- Tạo QC task nghiệm thu.
- Hàng chưa QC pass chưa available.

### E07-S06. Factory claim trong 3–7 ngày

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Nếu QC/kiểm nhận không đạt, tạo claim cho nhà máy.
- Claim có deadline phản hồi 3–7 ngày.
- Có trạng thái Open/Sent/Factory Responded/Resolved/Closed.
- Có attachment hình ảnh/biên bản.

---

## E08. Sales Order / OMS

### E08-S01. Sales Order create/edit

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo đơn hàng với customer/channel/items/price/discount/payment method.
- Validate SKU active.
- Validate available stock khi confirm/reserve.
- Có trạng thái Draft/Confirmed/Reserved/Picking/Packed/HandedOver/Delivered/Closed/Cancelled/Returned.
- Không cho sửa line sau khi đã reserved/picked nếu không có quyền.

### E08-S02. Price/discount basic engine

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Có bảng giá theo kênh cơ bản.
- Có discount line/order.
- Discount vượt ngưỡng phải approval nếu rule bật.
- Hiển thị gross amount/discount/net amount.
- Không expose cost nếu user không có quyền.

### E08-S03. Stock reservation

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Confirm order tạo reservation.
- Reservation theo SKU/batch nếu có rule.
- Không reserve batch HOLD/FAIL/EXPIRED.
- Cancel order release reservation.
- Reservation có audit và ledger/reservation record.

### E08-S04. Order import cơ bản từ kênh ngoài

**Priority:** P1  
**Size:** L  

**Acceptance Criteria:**

- Import order CSV/API từ website/sàn nếu chưa tích hợp trực tiếp.
- Mapping channel/customer/SKU.
- Dedupe theo external order id.
- Import lỗi có report.

---

## E09. Pick / Pack / Ship / Carrier Handover

### E09-S01. Picking task

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo picking task từ order đã reserved.
- Gợi ý vị trí/batch theo FEFO nếu bật.
- Kho xác nhận pick qty.
- Pick sai batch phải có cảnh báo.
- Pick xong chuyển order sang Picked.

### E09-S02. Packing task / packing station

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Đóng hàng theo từng đơn.
- Kiểm SKU/số lượng/xung quanh khu đóng hàng theo nội quy.
- Có trạng thái Packing/Packed.
- Có thể in packing slip/label cơ bản.
- Packed order chuyển sang khu chờ bàn giao ĐVVC.

### E09-S03. Carrier manifest

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Tạo manifest theo carrier/chuyến/ngày/ca.
- Thêm các đơn đã packed.
- Hiển thị số lượng đơn dự kiến.
- Cho chia khu/thùng/rổ theo carrier.
- Manifest có trạng thái Draft/Ready/HandingOver/Completed/Exception.

### E09-S04. Scan handover

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Kho quét mã đơn/mã vận đơn trước khi bàn giao.
- Nếu đơn thuộc manifest: mark scanned.
- Nếu đơn không thuộc manifest: cảnh báo.
- Nếu thiếu đơn: manifest không được completed.
- Nếu đủ đơn: ký/xác nhận bàn giao với ĐVVC.
- Ghi scan event và audit.

### E09-S05. Missing order exception khi bàn giao

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Nếu chưa đủ đơn, hệ thống cho tạo exception.
- Exception ghi reason: chưa đóng, thất lạc khu đóng, sai mã, hủy, khác.
- Cho tìm đơn trong khu đóng hàng.
- Nếu tìm thấy thì scan lại.
- Nếu không tìm thấy, giữ manifest trạng thái exception.

### E09-S06. Shipment status tracking cơ bản

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Update trạng thái shipment: Created/Packed/HandedOver/InTransit/Delivered/Failed/Returned.
- Có thể import tracking status từ carrier/CSV.
- Order status đồng bộ theo shipment status.

---

## E10. Returns / Hàng hoàn

### E10-S01. Return receiving

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Nhận hàng hoàn từ shipper/carrier.
- Quét mã đơn/mã vận đơn.
- Đưa hàng vào khu vực hàng hoàn.
- Tạo return record.
- Nếu không tìm thấy order/shipment source thì tạo unknown return exception.

### E10-S02. Return inspection

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Kiểm tình trạng bên trong: rách, móp, hư hỏng, đã mở, thiếu hàng.
- Ghi condition theo item.
- Upload hình ảnh.
- Ghi nhận phân loại: còn dùng / không dùng / cần kiểm thêm.

### E10-S03. Return disposition

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Còn dùng: nhập lại kho/quarantine tùy rule.
- Không dùng: chuyển kho hỏng/lab.
- Cần kiểm thêm: giữ HOLD.
- Mọi disposition tạo stock movement phù hợp.
- Không cho hàng hoàn tự quay lại available nếu chưa được disposition.

### E10-S04. Return reason and report

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Ghi reason hàng hoàn: khách từ chối, giao thất bại, lỗi hàng, sai địa chỉ, khác.
- Report hàng hoàn theo carrier/channel/SKU/KOL nếu có data.

---

## E11. Finance Lite / Reconciliation

### E11-S01. Basic payment record

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Ghi nhận thanh toán đơn hàng.
- Hỗ trợ COD/chuyển khoản/tiền mặt.
- Payment có trạng thái Pending/Received/Reconciled/Cancelled.
- Gắn với order/customer.

### E11-S02. COD reconciliation cơ bản

**Priority:** P1  
**Size:** L  

**Acceptance Criteria:**

- Import file COD từ carrier.
- Match theo mã vận đơn/order.
- Hiển thị lệch: thiếu tiền, thừa tiền, chưa giao, hoàn.
- Có trạng thái reconciled/unmatched.

### E11-S03. Supplier/factory payment tracking cơ bản

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Theo dõi công nợ PO/factory order ở mức đơn giản.
- Ghi nhận cọc, thanh toán lần cuối.
- Không thay thế phần mềm kế toán nếu chưa scope.

---

## E12. Report / Dashboard

### E12-S01. CEO/Operations dashboard cơ bản

**Priority:** P1  
**Size:** L  

**Acceptance Criteria:**

- Doanh thu/ngày/kênh cơ bản.
- Đơn chờ xử lý, chờ pack, chờ bàn giao.
- Batch hold/fail.
- Tồn cận date.
- Hàng hoàn chưa xử lý.
- Exception kho/ĐVVC.

### E12-S02. Warehouse dashboard

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Task nhập/xuất/pack/handover/return/count trong ngày.
- Task quá hạn.
- Đơn packed chưa handover.
- Manifest exception.
- Shift closing status.

### E12-S03. Inventory health report

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Tồn theo SKU/batch/kho/vị trí.
- Available/Reserved/Hold.
- Near-expiry.
- Stock movement history.

### E12-S04. Subcontract report

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Đơn gia công đang chạy.
- NVL/bao bì đã chuyển cho nhà máy.
- Mẫu chưa duyệt.
- Hàng đã nhận/chưa nhận.
- Claim nhà máy đang mở.

---

## E13. Integration

### E13-S01. Carrier integration abstraction

**Priority:** P1  
**Size:** L  

**Acceptance Criteria:**

- Có interface carrier adapter.
- Hỗ trợ manual/CSV trước nếu chưa API.
- Dễ thêm API carrier sau.
- Shipment/manifest không phụ thuộc trực tiếp vào carrier cụ thể.

### E13-S02. Barcode scanner support

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Scanner keyboard wedge hoạt động ở màn hình pick/pack/handover/return.
- Cursor focus đúng ô scan.
- Scan thành công có âm/thông báo rõ.
- Scan lỗi có cảnh báo rõ, không làm mất context.

### E13-S03. Accounting export

**Priority:** P1  
**Size:** M  

**Acceptance Criteria:**

- Export PO/GRN/payment/order/COD theo định dạng cấu hình.
- Không expose dữ liệu nhạy cảm cho user không có quyền.

---

## E14. Data Migration

### E14-S01. Master data import template

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Template import SKU, material, packaging, warehouse, location, supplier, customer, carrier, factory.
- Có validation trước import.
- Có error report theo dòng.
- Có dry-run mode.

### E14-S02. Opening stock migration

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Import tồn đầu kỳ theo SKU/batch/kho/vị trí/QC status/expiry.
- Tạo opening stock movement.
- Không import stock trực tiếp vào balance.
- Có reconciliation report sau import.

### E14-S03. Opening open orders / documents

**Priority:** P1  
**Size:** L  

**Acceptance Criteria:**

- Import đơn hàng mở, PO mở, factory order mở nếu cần.
- Trạng thái sau import rõ ràng.
- Không tạo dữ liệu mồ côi.

---

## E15. QA / UAT / Release

### E15-S01. Test data pack

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Có bộ dữ liệu test gồm SKU, batch, kho, NCC, factory, carrier, customer, orders.
- Có case đặc thù: QC HOLD, near-expiry, thiếu đơn handover, hàng hoàn không dùng, claim nhà máy.

### E15-S02. UAT support and bug triage

**Priority:** P0  
**Size:** L  

**Acceptance Criteria:**

- Có UAT schedule.
- Có bug severity.
- Có daily bug triage.
- Blocker/Critical phải resolve trước go-live.

### E15-S03. Release readiness checklist

**Priority:** P0  
**Size:** M  

**Acceptance Criteria:**

- Có checklist deploy.
- Có smoke test.
- Có backup trước release.
- Có rollback plan.
- Có sign-off từ business/QA/tech.

---

# 10. Sprint Plan đề xuất

## Tổng quan timeline

Đề xuất sprint 2 tuần. Timeline thực tế phụ thuộc đội và mức độ custom. Bản dưới đây là baseline cho một team cỡ vừa.

```text
Sprint 0: Setup + refine backlog
Sprint 1: Foundation + Auth/RBAC/Audit base
Sprint 2: Master Data + Batch config + Warehouse structure
Sprint 3: Stock Ledger base + Inventory balance + Opening import skeleton
Sprint 4: Purchase/PO + Inbound Receiving
Sprint 5: QC + Batch release + Receiving exception
Sprint 6: Subcontract manufacturing base + Material transfer + Sample approval
Sprint 7: Sales Order + Price/Discount + Reservation
Sprint 8: Picking/Packing + Packing station
Sprint 9: Carrier Manifest + Scan Handover + Missing order exception
Sprint 10: Returns + Return inspection/disposition
Sprint 11: Finance Lite + COD/basic reconciliation + Reports
Sprint 12: Integrations + Migration dry run + Performance hardening
Sprint 13: UAT + Bug fixing + Release readiness
```

Nếu cần go-live nhanh hơn, có thể chạy song song frontend/backend/design nhưng không được bỏ qua test stock ledger, batch/QC, handover, returns.

---

## Sprint 0. Project Setup & Backlog Refinement

**Mục tiêu:** Dự án sẵn sàng thi công, không build mù.  
**Thời lượng:** 1–2 tuần  

### Deliverables

- Setup repo, environment, CI/CD skeleton.
- Backlog refine với PO/BA/key users.
- Confirm scope Phase 1.
- Confirm roles/permissions baseline.
- Confirm data migration source.
- Confirm scanning hardware/barcode format.
- Confirm carrier/ĐVVC integration approach.

### Stories

| Story | Priority | Owner |
|---|---|---|
| E00-S01 Repo structure | P0 | Tech Lead |
| E00-S02 Environment setup | P0 | DevOps/Tech Lead |
| E00-S03 CI pipeline base | P0 | DevOps |
| Backlog review workshop | P0 | PM/BA/PO |
| Confirm workflow As-Is decision | P0 | BA/Business |

### Exit Criteria

- Team chạy được local backend/frontend.
- Có staging/dev environment.
- Có backlog được PO sign-off ở mức epic/story lớn.
- Không còn tranh cãi scope P0 Phase 1.

---

## Sprint 1. Foundation + Auth/RBAC/Audit Base

**Mục tiêu:** Có lớp nền hệ thống: login, role, permission, audit log cơ bản.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E00-S04 ERP shell frontend | P0 | M |
| E01-S01 Login/logout/session | P0 | M |
| E01-S02 Role/permission base | P0 | L |
| E01-S04 Audit log base | P0 | L |
| Basic user management | P0 | M |

### Acceptance Sprint

- User đăng nhập được.
- Sidebar/menu hiển thị theo quyền.
- API check permission server-side.
- Audit log ghi được các action cơ bản.
- Có user/role seed cho UAT.

---

## Sprint 2. Master Data + Batch/Warehouse Structure

**Mục tiêu:** Khóa dữ liệu gốc để các module sau dùng chung.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E02-S01 SKU master | P0 | L |
| E02-S02 Material/Packaging master | P0 | L |
| E02-S03 Warehouse/Location master | P0 | M |
| E02-S04 Supplier/Factory master | P0 | M |
| E02-S05 Customer/Channel/Carrier master | P0 | M |
| E02-S06 Batch numbering rule | P0 | M |

### Acceptance Sprint

- Tạo được SKU/material/packaging/kho/vị trí/NCC/nhà máy/ĐVVC.
- Batch rule hoạt động.
- Không cho duplicate master data trọng yếu.
- Frontend list/detail/create/edit đúng pattern.

---

## Sprint 3. Stock Ledger Base + Inventory Balance

**Mục tiêu:** Dựng lõi tồn kho trước khi build mua/bán.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E06-S01 Stock ledger base | P0 | XL split |
| E06-S02 Stock balance/available | P0 | L |
| E06-S03 Adjustment approval | P0 | L |
| E14-S01 Master data import template | P0 | M |
| E14-S02 Opening stock migration skeleton | P0 | L |

### Acceptance Sprint

- Mọi thay đổi tồn qua stock movement.
- Có available/physical/hold/reserved base.
- Adjustment cần duyệt.
- Opening stock import tạo movement.
- Không có API sửa balance trực tiếp.

---

## Sprint 4. Purchase/PO + Inbound Receiving

**Mục tiêu:** Từ mua hàng đến nhận hàng, chuẩn bị QC/nhập kho.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E04-S01 Purchase Request | P0 | M |
| E04-S02 Purchase Order | P0 | L |
| E04-S03 Inbound Receiving/GRN | P0 | L |
| E04-S04 Receiving exception | P0 | M |
| E03-S03 Attachment | P0 | M |

### Acceptance Sprint

- Tạo PR/PO/GRN được.
- Nhận hàng theo PO, có số lượng thực nhận, batch, hạn dùng.
- Hàng yêu cầu QC vào HOLD.
- Có exception khi thiếu/hư/sai batch.
- Có attachment chứng từ giao hàng.

---

## Sprint 5. QC + Batch Release

**Mục tiêu:** Hàng không pass QC không được dùng/bán.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E05-S01 QC inbound inspection | P0 | L |
| E05-S02 Batch status engine | P0 | L |
| QC release creates stock movement | P0 | M |
| QC fail/quarantine handling | P0 | M |
| E05-S04 Complaint link basic | P1 | M |

### Acceptance Sprint

- QC task tự tạo từ GRN cần QC.
- Pass chuyển stock từ HOLD sang AVAILABLE.
- Fail không cho reserve/pick/ship.
- Batch status update có audit.
- Batch detail hiển thị QC history.

---

## Sprint 6. Subcontract Manufacturing Base

**Mục tiêu:** ERP phản ánh đúng mô hình gia công ngoài thực tế.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E07-S01 Factory/Subcontract Order | P0 | L |
| E07-S02 Deposit tracking basic | P1 | M |
| E07-S03 Material transfer to factory | P0 | L |
| E07-S04 Sample approval | P0 | L |
| Subcontract document attachment | P0 | M |

### Acceptance Sprint

- Tạo đơn gia công với nhà máy.
- Chuyển NVL/bao bì sang nhà máy tạo stock movement.
- Có biên bản bàn giao.
- Sample approval/reject/rework hoạt động.
- Chỉ sample approved mới cho chuyển hàng loạt nếu rule bật.

---

## Sprint 7. Sales Order + Reservation

**Mục tiêu:** Nhận đơn bán và giữ tồn đúng, không bán vượt tồn.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E08-S01 Sales Order | P0 | L |
| E08-S02 Price/discount basic | P0 | L |
| E08-S03 Stock reservation | P0 | L |
| Cancel order releases reservation | P0 | M |
| E08-S04 Order import basic | P1 | L |

### Acceptance Sprint

- Tạo/confirm/cancel order.
- Confirm order reserve stock.
- Không reserve batch HOLD/FAIL/EXPIRED.
- Cancel order release reservation.
- Discount vượt ngưỡng trigger approval nếu rule bật.

---

## Sprint 8. Picking / Packing

**Mục tiêu:** Kho soạn và đóng hàng đúng đơn, đúng SKU, đúng số lượng.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E09-S01 Picking task | P0 | L |
| E09-S02 Packing station | P0 | L |
| Packing QA/check khu đóng hàng | P0 | M |
| Print packing slip/label basic | P1 | M |
| E06-S06 Warehouse Daily Board base | P0 | L |

### Acceptance Sprint

- Order reserved tạo picking task.
- Pick xác nhận SKU/batch/qty.
- Packing kiểm SKU/qty trước khi packed.
- Packed order chuyển khu chờ bàn giao.
- Warehouse Daily Board hiển thị task pick/pack.

---

## Sprint 9. Carrier Manifest + Scan Handover

**Mục tiêu:** Bàn giao ĐVVC bằng scan/manifest, xử lý thiếu đơn rõ.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E09-S03 Carrier manifest | P0 | L |
| E09-S04 Scan handover | P0 | L |
| E09-S05 Missing order exception | P0 | M |
| E09-S06 Shipment status tracking basic | P1 | M |
| E13-S02 Barcode scanner support | P0 | M |

### Acceptance Sprint

- Tạo manifest theo ĐVVC/chuyến/ca.
- Scan đơn thuộc manifest thành công.
- Scan sai cảnh báo.
- Thiếu đơn không cho complete manifest.
- Đủ đơn mới xác nhận bàn giao.
- Scan event/audit đầy đủ.

---

## Sprint 10. Returns / Hàng hoàn

**Mục tiêu:** Hàng hoàn không làm bẩn tồn kho và không tự quay về available.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E10-S01 Return receiving | P0 | L |
| E10-S02 Return inspection | P0 | L |
| E10-S03 Return disposition | P0 | L |
| E10-S04 Return reason/report | P1 | M |
| Unknown return exception | P0 | M |

### Acceptance Sprint

- Quét hàng hoàn từ shipper.
- Tạo return record.
- Kiểm tình trạng, upload hình ảnh.
- Phân loại còn dùng/không dùng/cần kiểm thêm.
- Disposition tạo stock movement đúng.
- Return chưa disposition không available.

---

## Sprint 11. Finance Lite + Reports

**Mục tiêu:** Có đối soát cơ bản và báo cáo vận hành đủ cho sếp/kho.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E11-S01 Payment record | P1 | M |
| E11-S02 COD reconciliation | P1 | L |
| E11-S03 Supplier/factory payment tracking | P1 | M |
| E12-S01 CEO/Ops dashboard | P1 | L |
| E12-S02 Warehouse dashboard | P0 | L |
| E12-S03 Inventory health report | P1 | M |
| E12-S04 Subcontract report | P1 | M |

### Acceptance Sprint

- Có dashboard kho, task trong ngày, exception.
- Có report tồn theo SKU/batch/kho/QC.
- Có COD import/match cơ bản nếu có file.
- Có subcontract report.

---

## Sprint 12. Integration + Migration Dry Run + Hardening

**Mục tiêu:** Chuẩn bị cho UAT/go-live, tránh vỡ ở data và integration.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E13-S01 Carrier abstraction | P1 | L |
| E13-S03 Accounting export | P1 | M |
| E14-S02 Opening stock migration full | P0 | L |
| E14-S03 Opening documents import | P1 | L |
| Performance test for key APIs | P0 | M |
| Security/RBAC regression | P0 | M |

### Acceptance Sprint

- Migration dry run thành công.
- Stock reconciliation sau import khớp.
- Key APIs pass performance baseline.
- Export cơ bản hoạt động.
- RBAC/security test không có lỗi critical.

---

## Sprint 13. UAT + Bug Fixing + Release Readiness

**Mục tiêu:** Business nghiệm thu và chuẩn bị go-live.  

### Stories

| Story | Priority | Size |
|---|---|---|
| E15-S01 Test data pack | P0 | M |
| E15-S02 UAT support/bug triage | P0 | L |
| E15-S03 Release readiness checklist | P0 | M |
| Training handoff package | P0 | M |
| Go-live dry run | P0 | M |

### Acceptance Sprint

- UAT P0 pass.
- Không còn blocker/critical.
- Business sign-off.
- Migration/go-live dry run pass.
- Training materials đủ cho key users.
- Release checklist signed off.

---

# 11. Milestone Release Plan

## Release R0. Foundation Preview

**Bao gồm:** login, RBAC base, app shell, master data skeleton.  
**Mục tiêu:** Team nội bộ nhìn được cấu trúc ERP và bắt đầu review UX.  

## Release R1. Inventory Core

**Bao gồm:** master data, stock ledger, opening stock, stock balance, adjustment.  
**Mục tiêu:** Khóa lõi kho/tồn.

## Release R2. Inbound + QC + Subcontract

**Bao gồm:** PO, nhận hàng, QC, batch release, gia công ngoài, chuyển NVL/bao bì, duyệt mẫu.  
**Mục tiêu:** Kiểm soát hàng vào và sản xuất/gia công.

## Release R3. Order Fulfillment

**Bao gồm:** sales order, reservation, pick, pack, manifest, scan handover.  
**Mục tiêu:** Kiểm soát từ đơn đến bàn giao ĐVVC.

## Release R4. Returns + Reconciliation + Dashboard

**Bao gồm:** hàng hoàn, COD/payment lite, dashboard, reports, migration dry run.  
**Mục tiêu:** Đủ vận hành trước UAT.

## Release R5. UAT / Go-Live Candidate

**Bao gồm:** bug fix, hardening, migration, release readiness.  
**Mục tiêu:** Có bản ứng viên go-live.

---

# 12. Dependency Map

```text
Auth/RBAC/Audit
    ↓
Master Data
    ↓
Batch Rule + Warehouse Structure
    ↓
Stock Ledger + Stock Balance
    ↓
Purchase/Inbound → QC Release → Available Stock
    ↓
Sales Order → Reservation → Pick → Pack → Manifest → Handover
    ↓
Returns → Inspection → Disposition → Stock Movement
```

Luồng gia công ngoài:

```text
Factory Master + Item Master + Warehouse
    ↓
Factory Order
    ↓
Material Transfer to Factory
    ↓
Sample Approval
    ↓
Receive Finished Goods
    ↓
QC Acceptance/Reject
    ↓
Available Stock / Factory Claim
```

---

# 13. Traceability từ workflow thực tế sang backlog

| Workflow thực tế | Backlog phải có | Epic/Story |
|---|---|---|
| Tiếp nhận đơn hàng trong ngày | Warehouse Daily Board, Sales Order queue | E06-S06, E08-S01 |
| Xuất/nhập theo nội quy | PO/GRN, stock movement, outbound/pick | E04, E06, E09 |
| Soạn hàng và đóng gói | Picking task, Packing station | E09-S01, E09-S02 |
| Sắp xếp/tối ưu vị trí kho | Warehouse/location, transfer, bin movement | E02-S03, E06-S04 |
| Kiểm kê cuối ngày | Cycle count, shift closing | E06-S05, E06-S07 |
| Đối soát số liệu báo quản lý | Shift closing, warehouse dashboard | E06-S07, E12-S02 |
| Bàn giao ĐVVC theo thùng/rổ | Carrier manifest | E09-S03 |
| Đối chiếu số lượng đơn theo bảng | Manifest expected vs scanned | E09-S03, E09-S04 |
| Quét mã trước bàn giao | Scan handover | E09-S04, E13-S02 |
| Chưa đủ đơn thì kiểm tra lại mã/tìm khu đóng | Missing order exception | E09-S05 |
| Nhận hàng hoàn từ shipper | Return receiving | E10-S01 |
| Quét hàng hoàn | Return scan | E10-S01, E13-S02 |
| Kiểm tình trạng và phân loại còn dùng/không dùng | Return inspection/disposition | E10-S02, E10-S03 |
| Lên đơn với nhà máy | Factory/Subcontract Order | E07-S01 |
| Cọc đơn và chốt thời gian | Deposit tracking | E07-S02 |
| Chuyển NVL/bao bì sang nhà máy | Material transfer | E07-S03 |
| Làm mẫu/chốt mẫu | Sample approval | E07-S04 |
| Nhận hàng gia công về kho | Receive subcontracted goods | E07-S05 |
| Báo lỗi nhà máy trong 3–7 ngày | Factory claim | E07-S06 |

---

# 14. Definition of Done

## 14.1. Story Done

Một story chỉ được coi là done khi:

- có API/backend hoàn thành;
- có frontend nếu story có UI;
- có validation đúng Data Dictionary;
- có permission/RBAC;
- có audit log nếu là action nghiệp vụ;
- có test tối thiểu theo QA strategy;
- có migration nếu tạo/chỉnh DB schema;
- OpenAPI updated nếu có API;
- bug critical/high được xử lý;
- PO/BA xác nhận acceptance criteria pass.

## 14.2. Module Done

Một module chỉ được coi là done khi:

- các P0 stories của module pass;
- state machine đúng;
- API contract pass;
- UI list/detail/create/edit/approve nếu cần;
- permission đúng;
- audit đúng;
- report cơ bản nếu module cần;
- UAT case liên quan pass.

## 14.3. Sprint Done

Sprint chỉ được coi là done khi:

- demo được end-to-end phần đã build;
- no blocker/critical open;
- backlog status updated;
- tài liệu impacted updated;
- test report available;
- PO sign-off sprint demo hoặc ghi rõ pending.

---

# 15. Change Control cho backlog

Mọi thay đổi backlog phải ghi vào change log:

| Field | Mô tả |
|---|---|
| Change ID | Mã thay đổi |
| Requested By | Người yêu cầu |
| Date | Ngày yêu cầu |
| Current Scope | Scope hiện tại |
| Proposed Change | Thay đổi đề xuất |
| Reason | Lý do |
| Impact | Ảnh hưởng tới time/cost/quality/docs |
| Decision | Approved/Rejected/Deferred |
| Approved By | Người duyệt |
| Target Sprint | Sprint thực hiện |

Nguyên tắc: **đã vào sprint thì hạn chế đổi. Muốn đổi phải qua PO/PM/Tech Lead.**

---

# 16. Risk Log trong quá trình sprint

| Risk | Ảnh hưởng | Mức độ | Mitigation |
|---|---|---|---|
| Stock ledger thiết kế sai | Tồn kho sai toàn hệ thống | Critical | Build/test sớm Sprint 3, review Tech Lead/BA |
| Workflow kho thực tế khác thêm nữa | Màn hình không dùng được | High | Workshop với kho trước Sprint 8/9 |
| Barcode format chưa thống nhất | Scan lỗi khi bàn giao/return | High | Confirm Sprint 0–2, test bằng scanner thật |
| Carrier chưa có API | Handover/tracking delay | Medium | Hỗ trợ CSV/manual adapter trước |
| Dữ liệu master bẩn | Migration fail/go-live lệch | High | Data cleansing trước Sprint 12 |
| QC rule chưa rõ | Batch pass/hold sai | High | QA/QC sign-off Sprint 5 |
| Gia công ngoài có nhiều ngoại lệ | Factory order không đủ | Medium | Build base P0, đưa case phức tạp vào Phase 1.1 |
| User không quen ERP | Data nhập sai | Medium | SOP/training trước UAT |

---

# 17. Backlog Parking Lot / Phase 2

Các item nên để Phase 2 nếu không phải bắt buộc go-live:

- CRM 360° đầy đủ;
- loyalty;
- KOL/Affiliate management đầy đủ;
- KOL claim governance nâng cao;
- HRM, chấm công, payroll;
- full accounting ledger;
- advanced BI/cohort/profitability;
- supplier portal;
- dealer portal;
- KOL portal;
- mobile app native;
- AI demand forecast;
- advanced MRP;
- warehouse handheld app chuyên dụng;
- automated carrier API nhiều hãng nếu Phase 1 dùng CSV/manual là đủ.

---

# 18. Checklist trước khi bắt đầu Sprint 1

- [ ] PO đã sign-off scope Phase 1.
- [ ] Tech Lead sign-off kiến trúc Go/Next/PostgreSQL.
- [ ] BA sign-off Data Dictionary bản dùng build.
- [ ] Kho sign-off workflow pick/pack/handover/return.
- [ ] QA/QC sign-off batch/QC state.
- [ ] Người phụ trách gia công sign-off factory workflow.
- [ ] Finance sign-off mức Finance Lite Phase 1.
- [ ] PM sign-off sprint cadence.
- [ ] DevOps setup repo/environment.
- [ ] Hardware scanner/barcode format đã xác nhận hoặc có kế hoạch test.

---

# 19. Checklist trước UAT

- [ ] P0 backlog completed.
- [ ] Stock ledger test pass.
- [ ] Batch/QC test pass.
- [ ] Purchase/inbound test pass.
- [ ] Subcontract workflow test pass.
- [ ] Sales reservation test pass.
- [ ] Pick/pack test pass.
- [ ] Carrier manifest/scan handover test pass.
- [ ] Missing order exception test pass.
- [ ] Returns inspection/disposition test pass.
- [ ] Shift closing test pass.
- [ ] Migration dry run pass.
- [ ] RBAC/security regression pass.
- [ ] Dashboard/report P0/P1 pass.
- [ ] UAT data prepared.
- [ ] Key users trained.

---

# 20. Kết luận

Backlog Phase 1 nên được triển khai theo nguyên tắc:

```text
Khóa nền tảng
→ Khóa dữ liệu gốc
→ Khóa tồn kho/batch/QC
→ Khóa mua hàng/nhận hàng
→ Khóa gia công ngoài
→ Khóa bán hàng/giữ tồn
→ Khóa pick/pack/bàn giao ĐVVC
→ Khóa hàng hoàn
→ Khóa đối soát/báo cáo
→ UAT/go-live
```

Đây là thứ tự ít rủi ro nhất cho công ty mỹ phẩm có vận hành kho, ĐVVC, hàng hoàn và gia công ngoài.

Một câu để team nhớ:

> **Đừng build ERP theo menu. Hãy build theo dòng chảy tiền-hàng-rủi ro. Cái nào sai làm mất tiền trước thì build và test trước.**

---

## 21. Tài liệu tiếp theo nên làm

Sau tài liệu này, tài liệu tiếp theo nên là:

```text
26_ERP_SOP_Training_Manual_Phase1_MyPham_v1.md
```

Vì backlog đã nói dev build gì, còn SOP/Training Manual sẽ nói nhân viên dùng hệ thống như thế nào sau khi build xong.
