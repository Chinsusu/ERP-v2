# 32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1

## 1. Mục đích tài liệu

Tài liệu này là **Master Document Index + Traceability + Handoff Pack** cho bộ tài liệu ERP mỹ phẩm Phase 1.

Sau khi đã có bộ tài liệu từ **01 đến 31**, rủi ro lớn nhất không còn là thiếu tài liệu, mà là:

- Team không biết đọc tài liệu nào trước.
- BA, PM, Dev, QA, UI/UX hiểu khác nhau.
- Vendor chỉ đọc một phần rồi build lệch.
- File này nói một kiểu, file kia nói một kiểu.
- Không biết requirement nào đã được map sang màn hình, API, DB, test case và sprint backlog.
- Không biết tài liệu nào là **source of truth** khi có xung đột.

Tài liệu này dùng để làm “bản đồ tổng chỉ huy” cho toàn bộ bộ tài liệu ERP.

Nói ngắn gọn:

```text
Bộ 01–31 = kho tri thức ERP.
File 32 = bản đồ để đội thi công không lạc.
```

---

## 2. Phạm vi

Tài liệu này bao phủ:

- Toàn bộ tài liệu ERP hiện có từ **01 đến 31**.
- Phase 1 ERP cho công ty mỹ phẩm.
- Workflow thực tế đã bóc từ 4 tài liệu nội bộ:
  - Công việc hằng ngày của kho.
  - Nội quy nhập kho, xuất kho, đóng hàng, xử lý hàng hoàn.
  - Quy trình bàn giao hàng cho đơn vị vận chuyển.
  - Quy trình sản xuất/gia công ngoài với nhà máy.
- Các module lõi Phase 1:
  - Master Data.
  - Purchasing.
  - QA/QC.
  - Inventory/Warehouse.
  - Sales Order.
  - Shipping/Handover.
  - Returns.
  - Subcontract Manufacturing.
  - Finance basic.
  - Security/RBAC/Audit.
  - Reporting/KPI.

Không dùng tài liệu này để thay thế PRD, Process Flow, API Spec, Database Schema hoặc Coding Standard. File này là **file điều phối, tra cứu và bàn giao**.

---

## 3. Nguyên tắc dùng bộ tài liệu

### 3.1. Không đọc rời rạc

Không được lấy một tài liệu riêng lẻ làm căn cứ build nếu requirement đó còn liên quan tới file khác.

Ví dụ:

```text
Bàn giao ĐVVC bằng scan
không chỉ nằm ở Process Flow.
Nó còn liên quan tới Screen List, API, Database, Security, QA, UAT, Sprint Backlog, SOP và Incident Playbook.
```

### 3.2. Khi có xung đột, dùng source of truth

Nếu hai tài liệu nói khác nhau, xử lý theo thứ tự:

```text
1. Decision Log / Change Log mới nhất
2. PRD/SRS bản mới nhất
3. Process Flow bản mới nhất
4. Data Dictionary bản mới nhất
5. Technical/API/Database spec liên quan
6. Nếu vẫn mâu thuẫn → tạo decision item, không tự đoán
```

### 3.3. Workflow thực tế phải được ưu tiên kiểm chứng

Vì công ty đã có workflow thực tế về kho, hàng hoàn, bàn giao ĐVVC và gia công ngoài, mọi thiết kế cần đối chiếu với As-Is trước khi build.

Ví dụ bắt buộc kiểm chứng:

- Kho có nhịp công việc theo ngày.
- Có kiểm kê tồn cuối ngày.
- Có đối soát số liệu và báo cáo quản lý cuối ca.
- Đóng hàng có phân loại đơn theo ĐVVC/đơn lẻ/đơn sàn.
- Bàn giao ĐVVC có phân khu, để theo thùng/rổ, đối chiếu số lượng, quét mã.
- Hàng hoàn có nhận từ shipper, đưa vào khu hàng hoàn, quét hàng hoàn, kiểm tra tình trạng, phân loại còn dùng/không dùng.
- Sản xuất hiện có nhánh gia công ngoài: lên đơn với nhà máy, cọc đơn, chuyển NVL/bao bì, làm mẫu, chốt mẫu, sản xuất hàng loạt, giao hàng về kho, kiểm tra và phản hồi lỗi nhà máy trong 3–7 ngày.

### 3.4. Tài liệu không phải để trang trí

Một tài liệu chỉ có giá trị khi nó giúp:

- Khóa scope.
- Khóa quy trình.
- Khóa dữ liệu.
- Khóa quyền hạn.
- Giảm lỗi build.
- Giảm lỗi go-live.
- Giảm tranh cãi giữa business và tech.

---

## 4. Document Index tổng thể

### 4.1. Nhóm chiến lược và định hướng

| No. | Tài liệu | Mục đích | Ai đọc chính | Khi dùng |
|---|---|---|---|---|
| 01 | `ERP_Blueprint_My_Pham_v1.md` | Bức tranh tổng thể ERP mỹ phẩm | CEO, PM, BA, Solution Architect | Đọc đầu tiên để hiểu tầm nhìn |
| 02 | `Tai_lieu_tiep_theo_ERP.md` | Lộ trình tài liệu cần làm sau Blueprint | CEO, PM, BA | Định hướng bộ hồ sơ dự án |
| 31 | `31_ERP_Phase2_Scope_CRM_HRM_KOL_Finance_MyPham_v1.md` | Phạm vi Phase 2 | CEO, PM, Product Owner | Không nhồi Phase 2 vào Phase 1 |

### 4.2. Nhóm tài liệu nghiệp vụ lõi

| No. | Tài liệu | Mục đích | Ai đọc chính | Khi dùng |
|---|---|---|---|---|
| 03 | `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md` | Requirement Phase 1 | BA, PM, Dev Lead, QA | Source chính cho scope nghiệp vụ |
| 04 | `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md` | Quyền hạn và duyệt | BA, PM, Admin, Security, Dev | Khi thiết kế quyền, approval, menu |
| 05 | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` | Từ điển dữ liệu và master data | BA, Dev, DB, QA | Khi định nghĩa field, status, data rules |
| 06 | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` | Quy trình To-Be | BA, PM, Dev, QA, Business Owner | Khi vẽ flow và acceptance logic |
| 07 | `07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1.md` | Báo cáo và KPI | CEO, PM, BI, Dev, QA | Khi build dashboard/report |
| 08 | `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md` | Danh sách màn hình và wireframe | UI/UX, FE, BA, QA | Khi thiết kế giao diện |
| 09 | `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md` | Kịch bản UAT | QA, BA, Business User | Khi nghiệm thu với user |
| 10 | `10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1.md` | Migration và cutover | PM, Data, DevOps, Business | Khi chuyển dữ liệu và go-live |

### 4.3. Nhóm kiến trúc kỹ thuật và engineering

| No. | Tài liệu | Mục đích | Ai đọc chính | Khi dùng |
|---|---|---|---|---|
| 11 | `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md` | Kiến trúc backend Go | Tech Lead, BE, Architect | Khi dựng backend |
| 12 | `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md` | Chuẩn code Go | BE, Tech Lead, Code Reviewer | Khi code và review |
| 13 | `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md` | Chuẩn module/component BE | BE, Architect | Khi chia module, boundary, service |
| 14 | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` | Chuẩn UI/UX | UI/UX, FE, PM | Khi thiết kế giao diện |
| 15 | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` | Kiến trúc frontend | FE, Tech Lead | Khi dựng Next.js app |
| 16 | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` | Chuẩn API/OpenAPI | BE, FE, QA, Architect | Khi định nghĩa API contract |
| 17 | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` | Chuẩn DB PostgreSQL | BE, DB, Architect | Khi thiết kế schema/migration |
| 18 | `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` | DevOps/CI-CD/env | DevOps, Tech Lead | Khi setup môi trường/deploy |
| 19 | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` | Security/RBAC/audit | Security, BE, Admin, PM | Khi khóa bảo mật và audit |

### 4.4. Nhóm As-Is, Gap và revision

| No. | Tài liệu | Mục đích | Ai đọc chính | Khi dùng |
|---|---|---|---|---|
| 20 | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` | Workflow hiện tại từ tài liệu nội bộ | BA, PM, Business Owner | Khi kiểm chứng thiết kế |
| 21 | `21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md` | So sánh As-Is vs To-Be, quyết định chỉnh | BA, PM, CEO, Architect | Khi chốt workflow cuối |
| 22 | `22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham.md` | Log tài liệu lõi cần sửa lên v1.1 | PM, BA, Tech Lead | Khi update core docs |
| 41 | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` | Sprint 2 Order Fulfillment Core task board | PM, Tech Lead, Dev, QA | Khi triển khai order-to-handover |
| 42 | `42_ERP_Sprint2_Changelog_Order_Fulfillment_Core_MyPham_v1.md` | Changelog/release note Sprint 2 | PM, Tech Lead, QA | Khi theo dõi kickoff và release evidence Sprint 2 |
| 43 | `43_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` | Sprint 3 Returns Reconciliation Core task board | PM, Tech Lead, Dev, QA | Khi triển khai returns, stock count, adjustment, shift closing |
| 44 | `44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1.md` | Changelog/release note Sprint 3 | PM, Tech Lead, QA, DevOps | Khi theo dõi release evidence Sprint 3 |
| 45 | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` | Sprint 4 Purchase Order + Inbound QC task board | PM, Tech Lead, Dev, QA, DevOps | Khi triển khai purchase, receiving, inbound QC, supplier rejection |

### 4.5. Nhóm tích hợp, test, delivery và vận hành

| No. | Tài liệu | Mục đích | Ai đọc chính | Khi dùng |
|---|---|---|---|---|
| 23 | `23_ERP_Integration_Spec_Phase1_MyPham_v1.md` | Chuẩn tích hợp ngoài | BE, FE, DevOps, Vendor | Khi tích hợp carrier, POS, marketplace, payment |
| 24 | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` | Chiến lược QA/test automation | QA, Dev, PM | Khi kiểm thử kỹ thuật |
| 25 | `25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md` | Backlog và sprint plan | PM, PO, Dev, QA | Khi triển khai theo sprint |
| 26 | `26_ERP_SOP_Training_Manual_Phase1_MyPham_v1.md` | SOP và training user | Trainer, Business User, PM | Khi đào tạo người dùng |
| 27 | `27_ERP_GoLive_Runbook_Hypercare_Phase1_MyPham_v1.md` | Go-live và hypercare | PM, DevOps, Business, Support | Khi bật hệ thống |
| 28 | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` | Xử lý rủi ro/sự cố | Support, PM, QA, Tech Lead | Khi có incident |
| 29 | `29_ERP_Operations_Support_Model_Phase1_MyPham_v1.md` | Mô hình support sau go-live | Support, Admin, PM | Khi vận hành hệ thống |
| 30 | `30_ERP_Data_Governance_Change_Control_Phase1_MyPham_v1.md` | Quản trị dữ liệu và change control | Data Owner, PM, Admin | Khi kiểm soát thay đổi |

---

## 5. Source of Truth Matrix

| Chủ đề | Source of Truth chính | Source phụ | Ghi chú |
|---|---|---|---|
| Tầm nhìn ERP tổng thể | 01 Blueprint | 03 PRD | Dùng để align CEO/PM/vendor |
| Scope Phase 1 | 03 PRD/SRS | 25 Backlog | Nếu backlog khác PRD, phải tạo change request |
| Workflow To-Be | 06 Process Flow | 21 Gap Analysis, 22 Revision Log | Sau Gap Analysis, cần update 06 lên v1.1 |
| Workflow hiện tại | 20 As-Is | 4 tài liệu nội bộ PDF | Chỉ dùng để đối chiếu, không tự động thành To-Be |
| Quyền hạn | 04 Permission Matrix | 19 Security/RBAC | 04 cho business quyền; 19 cho technical enforcement |
| Duyệt nghiệp vụ | 04 Permission Matrix | 06 Process Flow | Approval phải hiện trên API/UI/audit |
| Dữ liệu, field, status | 05 Data Dictionary | 17 DB Schema, 16 API | 05 là nghĩa nghiệp vụ, 16/17 là implementation |
| Màn hình | 08 Screen/Wireframe | 14 UI/UX, 15 FE Architecture | 08 là screen scope, 14/15 là chuẩn xây |
| UI/UX standard | 14 UI/UX | 15 Frontend | 14 là design system; 15 là implementation FE |
| Backend architecture | 11 Technical Architecture | 12, 13 | 11 là hướng tổng; 12/13 là rule code/module |
| API contract | 16 OpenAPI | 11, 15, 24 | API là contract giữa FE/BE/QA |
| Database | 17 PostgreSQL | 05 Data Dictionary, 11 Architecture | DB không được tự thêm status trái 05 |
| Integration | 23 Integration Spec | 16 API, 18 DevOps | Carrier/POS/payment/marketplace theo 23 |
| QA kỹ thuật | 24 QA Strategy | 09 UAT | 24 cho test kỹ thuật, 09 cho business UAT |
| UAT | 09 UAT Scenarios | 06 Process Flow, 08 Screen List | User nghiệm thu theo 09 |
| Sprint plan | 25 Product Backlog, 37 Sprint 1 Board, 41 Sprint 2 Board | 03 PRD, 21 Gap | Nếu sprint scope đổi, phải ghi CR |
| Training | 26 SOP Manual | 08 Screen, 06 Flow | SOP phải update khi màn hình/workflow đổi |
| Go-live | 27 GoLive Runbook | 10 Migration, 29 Support | Không go-live nếu thiếu checklist |
| Incident | 28 Risk/Incident Playbook | 29 Support, 19 Security | Sự cố phải đi theo playbook |
| Support vận hành | 29 Support Model | 30 Governance | Sau go-live dùng 29/30 |
| Change control | 30 Data Governance | 22 Revision Log | Mọi sửa lớn phải qua CR |
| Phase 2 | 31 Phase2 Scope | 01 Blueprint | Không đưa Phase 2 vào Phase 1 nếu chưa duyệt |

---

## 6. Traceability Matrix theo capability

### 6.1. Master Data

| Requirement | PRD | Process | Screen | API | DB | Test | Sprint | SOP |
|---|---|---|---|---|---|---|---|---|
| Quản lý SKU/thành phẩm | 03 | 06 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Quản lý nguyên liệu/bao bì | 03 | 06 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Quản lý kho/vị trí/khu vực | 03 | 06/20 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Quản lý nhà cung cấp/nhà máy | 03 | 06/20 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Quản lý ĐVVC/carrier | 03/23 | 06/20 | 08 | 16/23 | 17 | 09/24 | 25 | 26 |

### 6.2. Kho hằng ngày / Daily Warehouse Board

| Requirement | PRD | As-Is | Gap | Screen | API | DB | Test | Sprint | SOP/Go-live |
|---|---|---|---|---|---|---|---|---|---|
| Tiếp nhận đơn trong ngày | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 26/27 |
| Theo dõi công việc kho theo ca | 03 | 20 | 21 | 08 | 16 | 17 | 24 | 25 | 26/29 |
| Kiểm kê cuối ngày | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 26/27 |
| Đối soát số liệu cuối ca | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 26/27/29 |
| Báo cáo quản lý cuối ngày | 07 | 20 | 21 | 08 | 16 | 17 | 24 | 25 | 26 |

### 6.3. Nhập kho / Inbound Receiving

| Requirement | PRD | Process | Screen | API | DB | Security | Test | Sprint |
|---|---|---|---|---|---|---|---|---|
| Nhận chứng từ giao hàng | 03 | 06/20 | 08 | 16 | 17 | 19 | 09/24 | 25 |
| Kiểm tra số lượng/bao bì/lô | 03 | 06/20 | 08 | 16 | 17 | 19 | 09/24 | 25 |
| Đạt → sắp xếp vào kho | 03 | 06/20 | 08 | 16 | 17 | 19 | 09/24 | 25 |
| Không đạt → trả NCC/nhà máy | 03 | 06/20 | 08 | 16 | 17 | 19 | 09/24 | 25 |
| Ký xác nhận và lưu phiếu nhập | 03 | 06/20 | 08/14 | 16 | 17 | 19 | 09/24 | 25 |

### 6.4. QA/QC và batch

| Requirement | PRD | Data | Process | Screen | API | DB | Test | Incident |
|---|---|---|---|---|---|---|---|---|
| Batch QC Hold/Pass/Fail | 03 | 05 | 06 | 08 | 16 | 17 | 09/24 | 28 |
| Không cho bán batch chưa pass | 03 | 05 | 06 | 08 | 16 | 17 | 09/24 | 28 |
| Lưu QC result và attachment | 03 | 05 | 06 | 08 | 16 | 17 | 09/24 | 28 |
| Trace batch khi complaint | 03 | 05 | 06 | 08 | 16 | 17 | 09/24 | 28 |
| Báo lỗi nhà máy trong 3–7 ngày | 03/20 | 05 | 06/20 | 08 | 16 | 17 | 09/24 | 28 |

### 6.5. Xuất kho / Outbound

| Requirement | PRD | Process | Screen | API | DB | Test | Sprint | SOP |
|---|---|---|---|---|---|---|---|---|
| Lập phiếu xuất kho | 03 | 06/20 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Xuất hàng theo phiếu | 03 | 06/20 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Đối chiếu số lượng thực tế | 03 | 06/20 | 08 | 16 | 17 | 09/24 | 25 | 26 |
| Ký bàn giao | 03 | 06/20 | 08/14 | 16 | 17 | 09/24 | 25 | 26 |
| Trưởng kho lưu phiếu | 03 | 06/20 | 08 | 16 | 17 | 09/24 | 25 | 26 |

### 6.6. Đóng hàng / Pick-Pack

| Requirement | PRD | As-Is | Gap | Screen | API | DB | Test | Sprint |
|---|---|---|---|---|---|---|---|---|
| Nhận phiếu đơn hàng hợp lệ | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 |
| Lọc/phân loại đơn theo ĐVVC/đơn lẻ/đơn sàn | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 |
| Soạn hàng theo từng đơn | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 |
| Kiểm tra tại khu đóng hàng | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 |
| Đếm lại tổng đơn của mỗi sàn | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 |
| Chuyển đến khu bàn giao ĐVVC | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 |

### 6.7. Bàn giao ĐVVC / Carrier Handover

| Requirement | PRD | As-Is | Gap | Screen | API | DB | Integration | Test | Sprint | SOP/Incident |
|---|---|---|---|---|---|---|---|---|---|---|
| Phân chia khu vực để hàng | 03 | 20 | 21 | 08 | 16 | 17 | 23 | 09/24 | 25 | 26/28 |
| Để theo thùng/rổ, mỗi thùng/rổ số lượng bằng nhau | 03 | 20 | 21 | 08 | 16 | 17 | 23 | 09/24 | 25 | 26/28 |
| Đối chiếu số lượng đơn dựa trên bảng | 03 | 20 | 21 | 08 | 16 | 17 | 23 | 09/24 | 25 | 26/28 |
| Quét mã trực tiếp tại hầm/khu bàn giao | 03 | 20 | 21 | 08 | 16 | 17 | 23 | 09/24 | 25 | 26/28 |
| Đủ đơn → ký xác nhận bàn giao | 03 | 20 | 21 | 08 | 16 | 17 | 23 | 09/24 | 25 | 26/28 |
| Chưa đủ → kiểm tra mã/tìm lại khu đóng hàng | 03 | 20 | 21 | 08 | 16 | 17 | 23 | 09/24 | 25 | 26/28 |

### 6.8. Hàng hoàn / Returns

| Requirement | PRD | As-Is | Gap | Screen | API | DB | Test | SOP/Incident |
|---|---|---|---|---|---|---|---|---|
| Nhận hàng từ shipper | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |
| Đưa vào khu vực hàng hoàn | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |
| Quét hàng hoàn | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |
| Quay/chụp tình trạng sản phẩm nếu cần | 03 | 20 | 21 | 08/14 | 16 | 17 | 09/24 | 26/28 |
| Kiểm tra bên trong: tình trạng, tem niêm phong, hư hỏng | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |
| Còn sử dụng → chuyển vào kho | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |
| Không sử dụng → chuyển lên Lab/kho hỏng | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |
| Lập phiếu nhập/return disposition đầy đủ | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 26/28 |

### 6.9. Gia công ngoài / Subcontract Manufacturing

| Requirement | PRD | As-Is | Gap | Screen | API | DB | Test | Sprint | Incident |
|---|---|---|---|---|---|---|---|---|---|
| Lên đơn hàng với nhà máy | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Xác nhận số lượng/quy cách/mẫu mã | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Cọc đơn và xác nhận thời gian sản xuất/nhận hàng | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Chuyển NVL/bao bì cho nhà máy | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Ký biên bản bàn giao, lưu COA/MSDS/tem phụ/hóa đơn nếu cần | 03 | 20 | 21 | 08/14 | 16 | 17 | 09/24 | 25 | 28 |
| Làm mẫu → chốt mẫu | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Lưu mẫu | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Sản xuất hàng loạt | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Giao hàng về kho | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Kiểm tra số lượng/chất lượng | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Hàng được nhận → nhập kho | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Hàng không nhận → báo lỗi nhà máy trong 3–7 ngày | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |
| Thanh toán lần cuối | 03 | 20 | 21 | 08 | 16 | 17 | 09/24 | 25 | 28 |

---

## 7. Reading Path theo vai trò

### 7.1. CEO / Business Owner

Đọc theo thứ tự:

```text
01 Blueprint
03 PRD/SRS
07 Report/KPI
20 As-Is Workflow
21 Gap Analysis
25 Backlog/Sprint Plan
27 GoLive Runbook
31 Phase2 Scope
32 Master Index
```

CEO không cần đọc sâu code standard, nhưng nên xem:

```text
19 Security/RBAC
28 Incident Playbook
30 Data Governance
```

để hiểu rủi ro và quyền kiểm soát.

### 7.2. PM / Product Owner

Đọc bắt buộc:

```text
01, 03, 04, 05, 06, 07, 08, 09, 10
20, 21, 22
23, 24, 25, 26, 27, 28, 29, 30, 32
```

PM là người giữ consistency giữa toàn bộ tài liệu.

### 7.3. Business Analyst

Đọc bắt buộc:

```text
03 PRD/SRS
04 Permission Matrix
05 Data Dictionary
06 Process Flow
08 Screen List
09 UAT
20 As-Is
21 Gap Analysis
22 Revision Log
26 SOP
32 Master Index
```

BA phải chịu trách nhiệm khi requirement không map được sang screen/API/test.

### 7.4. UI/UX Designer

Đọc bắt buộc:

```text
03 PRD/SRS
04 Permission Matrix
05 Data Dictionary
06 Process Flow
08 Screen List/Wireframe
14 UI/UX Design System
15 Frontend Architecture
20 As-Is
21 Gap Analysis
26 SOP
32 Master Index
```

UI/UX phải đặc biệt chú ý:

- Batch/hạn dùng hiển thị nổi.
- QC Hold/Pass/Fail phải nhìn rõ.
- Warehouse scan UX phải nhanh.
- Shift closing/đối soát cuối ca phải dễ dùng.
- Bàn giao ĐVVC cần trạng thái đủ/chưa đủ.
- Hàng hoàn cần ảnh/video/ghi chú và disposition rõ.

### 7.5. Backend Go Developer

Đọc bắt buộc:

```text
03 PRD/SRS
04 Permission Matrix
05 Data Dictionary
06 Process Flow
11 Go Architecture
12 Go Coding Standards
13 Module/Component Standards
16 API Contract
17 Database Schema
19 Security/RBAC
20 As-Is
21 Gap Analysis
23 Integration Spec
24 QA Strategy
25 Backlog
32 Master Index
```

BE không được tự ý update stock table trực tiếp. Mọi biến động tồn kho phải đi qua stock ledger/movement.

### 7.6. Frontend Developer

Đọc bắt buộc:

```text
03 PRD/SRS
04 Permission Matrix
05 Data Dictionary
08 Screen List
14 UI/UX Design System
15 Frontend Architecture
16 API Contract
19 Security/RBAC
20 As-Is
21 Gap Analysis
24 QA Strategy
25 Backlog
32 Master Index
```

FE phải bám quyền, trạng thái và error code API.

### 7.7. QA / Tester

Đọc bắt buộc:

```text
03 PRD/SRS
04 Permission Matrix
05 Data Dictionary
06 Process Flow
08 Screen List
09 UAT Scenarios
16 API Contract
17 Database Schema
20 As-Is
21 Gap Analysis
24 QA Strategy
25 Backlog
27 GoLive Runbook
28 Incident Playbook
32 Master Index
```

QA phải test cả happy path và exception path.

### 7.8. DevOps

Đọc bắt buộc:

```text
11 Technical Architecture
16 API Contract
17 Database Schema
18 DevOps/CI-CD
19 Security/RBAC
23 Integration Spec
24 QA Strategy
27 GoLive Runbook
28 Incident Playbook
29 Support Model
32 Master Index
```

### 7.9. ERP Admin / Support

Đọc bắt buộc:

```text
04 Permission Matrix
19 Security/RBAC
26 SOP Training
27 GoLive Runbook
28 Incident Playbook
29 Operations Support Model
30 Data Governance
32 Master Index
```

---

## 8. Dependency Map

### 8.1. Tài liệu nền tảng

```text
01 Blueprint
  ↓
03 PRD/SRS
  ↓
04 Permission Matrix
05 Data Dictionary
06 Process Flow
08 Screen List
```

### 8.2. Tài liệu kiểm chứng thực tế

```text
4 PDF nội bộ
  ↓
20 Current Workflow As-Is
  ↓
21 Gap Analysis
  ↓
22 Core Docs Revision Log
  ↓
Core Docs v1.1
```

### 8.3. Tài liệu kỹ thuật

```text
03/05/06/08
  ↓
11 Technical Architecture
  ↓
12 Coding Standards
13 Module Standards
16 API Contract
17 DB Schema
18 DevOps
19 Security
```

### 8.4. Tài liệu triển khai

```text
03 PRD + 06 Flow + 08 Screen + 16 API + 17 DB
  ↓
24 QA Strategy
25 Product Backlog/Sprint Plan
  ↓
26 SOP
27 GoLive Runbook
28 Incident Playbook
29 Support Model
30 Governance
```

---

## 9. Core Docs v1.1 Revision Priority

Sau As-Is/GAP, các tài liệu sau nên được update lên v1.1 trước khi code sâu:

| Ưu tiên | Tài liệu cần revise | Lý do |
|---|---|---|
| P0 | 03 PRD/SRS | Scope phải bổ sung rõ daily warehouse, shift closing, carrier manifest, returns, subcontract manufacturing |
| P0 | 06 Process Flow | Flow To-Be phải bám workflow thực tế từ PDF |
| P0 | 05 Data Dictionary | Cần thêm status/field cho scan, manifest, return disposition, factory claim |
| P0 | 08 Screen List | Cần thêm màn hình daily board, shift closing, carrier manifest scan, return inspection, subcontract order |
| P0 | 16 API Contract | Cần action endpoint cho scan, close shift, return disposition, factory claim |
| P0 | 17 DB Schema | Cần schema/table cho warehouse shift, scan event, manifest, return disposition, subcontract materials transfer |
| P1 | 04 Permission Matrix | Cần quyền riêng cho shift close, return disposition, factory claim, carrier handover |
| P1 | 09 UAT | Cần UAT case cho flow thực tế |
| P1 | 24 QA Strategy | Cần test automation theo flow mới |
| P1 | 25 Backlog | Sprint backlog phải reflect v1.1 scope |
| P1 | 26 SOP | SOP phải khớp màn hình v1.1 |

---

## 10. Handoff Checklist cho Vendor/Dev Team

### 10.1. Checklist bắt buộc trước khi kickoff build

| Item | Owner | Status |
|---|---|---|
| Vendor/Dev đã nhận đủ file 01–32 | PM | Pending |
| Vendor/Dev đã đọc file 32 trước | PM | Pending |
| PM đã giải thích source of truth matrix | PM | Pending |
| Scope Phase 1 đã được chốt theo file 03 + 21 + 22 | PO/PM | Pending |
| Core docs cần revise v1.1 đã được xác nhận | PM/BA | Pending |
| Tech stack Go + Next.js + PostgreSQL đã được chốt | Tech Lead | Pending |
| OpenAPI-first workflow đã được chốt | Tech Lead | Pending |
| Stock ledger immutable đã được chốt | Architect/BE | Pending |
| RBAC/audit/security baseline đã được chốt | Security/BE | Pending |
| Sprint 0 plan đã sẵn sàng | PM/Tech Lead | Pending |

### 10.2. Checklist cho BA

| Item | Status |
|---|---|
| Requirement trong PRD có module owner rõ | Pending |
| Flow có happy path và exception path | Pending |
| Field/status đã map vào Data Dictionary | Pending |
| Permission/approval đã map vào 04/19 | Pending |
| Mỗi flow quan trọng có UAT case | Pending |
| Mỗi flow quan trọng có screen/API/DB mapping | Pending |

### 10.3. Checklist cho Backend

| Item | Status |
|---|---|
| Module boundary đã rõ | Pending |
| API contract trước khi code | Pending |
| Repository không vượt module ownership | Pending |
| Transaction boundary đã định nghĩa | Pending |
| Audit log đã gắn vào sensitive actions | Pending |
| Stock movement không thể sửa/xóa | Pending |
| Idempotency cho action endpoint quan trọng | Pending |
| Error code chuẩn | Pending |
| Unit/integration test có đủ | Pending |

### 10.4. Checklist cho Frontend

| Item | Status |
|---|---|
| Route/module map theo 15 | Pending |
| Component dùng theo design system 14 | Pending |
| Form validation khớp API/Data Dictionary | Pending |
| Button/action ẩn/hiện theo RBAC | Pending |
| Error state theo API error code | Pending |
| Scan UX có offline/slow network fallback nếu cần | Pending |
| Table filter/sort/search chuẩn | Pending |
| Audit/attachment block chuẩn | Pending |

### 10.5. Checklist cho QA

| Item | Status |
|---|---|
| Test case map với requirement | Pending |
| API contract test có trong CI | Pending |
| Stock ledger test bắt buộc | Pending |
| QC/batch test bắt buộc | Pending |
| Carrier handover scan test bắt buộc | Pending |
| Returns disposition test bắt buộc | Pending |
| Subcontract manufacturing test bắt buộc | Pending |
| Permission/security test bắt buộc | Pending |
| Regression suite có trước UAT | Pending |

---

## 11. Handoff Package theo giai đoạn

### 11.1. Giai đoạn Kickoff

Bàn giao:

```text
01 Blueprint
03 PRD/SRS
20 As-Is
21 Gap Analysis
22 Revision Log
32 Master Index
```

Mục tiêu:

```text
Toàn đội hiểu business, workflow thật, gap và source of truth.
```

### 11.2. Giai đoạn Solution Design

Bàn giao:

```text
04 Permission Matrix
05 Data Dictionary
06 Process Flow
08 Screen List
11 Technical Architecture
14 UI/UX Standard
16 API Contract
17 DB Schema
19 Security
23 Integration Spec
```

Mục tiêu:

```text
Chốt thiết kế nghiệp vụ + kỹ thuật trước khi sprint build.
```

### 11.3. Giai đoạn Implementation

Bàn giao:

```text
12 Go Coding Standards
13 Module Standards
15 Frontend Architecture
18 DevOps/CI-CD
24 QA Strategy
25 Product Backlog/Sprint Plan
```

Mục tiêu:

```text
Code theo chuẩn, test theo chuẩn, không build tự phát.
```

### 11.4. Giai đoạn UAT/Training/Go-live

Bàn giao:

```text
09 UAT Scenarios
10 Migration/Cutover
26 SOP Training
27 GoLive Runbook
28 Incident Playbook
29 Support Model
30 Data Governance
```

Mục tiêu:

```text
User nghiệm thu, được đào tạo, go-live có rollback/support rõ.
```

### 11.5. Giai đoạn Phase 2 Planning

Bàn giao:

```text
31 Phase2 Scope
07 KPI Catalog
30 Data Governance
```

Mục tiêu:

```text
Không nhồi Phase 2 vào Phase 1, nhưng giữ đường scale.
```

---

## 12. Requirement Traceability Template

Dùng template này cho mọi requirement mới hoặc thay đổi.

```text
Requirement ID:
Tên requirement:
Module:
Business owner:
Mô tả:
Nguồn:
- PRD:
- Process Flow:
- As-Is/GAP:

Ảnh hưởng:
- Permission:
- Data Dictionary:
- Screen:
- API:
- Database:
- Integration:
- Test/UAT:
- SOP:
- Incident Playbook:

Quyết định:
- Keep ERP standard
- Customize to real workflow
- Redesign

Priority:
Dependency:
Acceptance Criteria:
Open Questions:
Status:
```

---

## 13. Conflict Resolution Rule

Khi team gặp xung đột tài liệu, làm theo quy trình:

```text
1. Ghi lại conflict.
2. Xác định requirement/module liên quan.
3. Tra Source of Truth Matrix trong file 32.
4. Kiểm tra file 21 Gap Analysis và file 22 Revision Log.
5. Nếu vẫn chưa rõ, tạo Decision Item.
6. PO/PM/Business Owner/Tech Lead chốt.
7. Update tài liệu nguồn sự thật.
8. Thông báo change cho các team liên quan.
```

Không được giải quyết bằng cách:

```text
- Dev tự đoán.
- BA tự sửa trong chat không cập nhật tài liệu.
- PM chỉ nói miệng.
- Business user yêu cầu sửa trực tiếp trong sprint mà không có CR.
```

---

## 14. Definition of Ready cho một Epic/User Story

Một story chỉ được đưa vào sprint build khi đủ:

| Điều kiện | Bắt buộc |
|---|---|
| Requirement nằm trong PRD hoặc có Change Request | Yes |
| Flow đã rõ happy path/exception path | Yes |
| Permission/approval đã rõ | Yes |
| Data field/status đã rõ | Yes |
| Screen hoặc wireframe đã rõ | Yes |
| API impact đã rõ | Yes |
| DB impact đã rõ | Yes |
| Acceptance criteria đã rõ | Yes |
| Test case/UAT hoặc QA scenario đã có | Yes |
| Không còn blocking open question | Yes |

---

## 15. Definition of Done ở cấp requirement

Một requirement được xem là hoàn tất khi:

```text
1. UI hoàn thành theo design system.
2. API hoàn thành theo OpenAPI.
3. DB migration đã chạy trên staging/UAT.
4. Permission/RBAC đúng.
5. Audit log đúng nếu là sensitive action.
6. Unit/integration/API/E2E test đạt.
7. UAT case đạt.
8. SOP/training note cập nhật nếu ảnh hưởng user.
9. Incident/playbook cập nhật nếu là nghiệp vụ rủi ro cao.
10. Không còn bug P0/P1.
```

---

## 16. Risk Register cấp tài liệu

| Rủi ro | Tác động | Cách kiểm soát |
|---|---|---|
| Team đọc thiếu tài liệu | Build lệch | Dùng file 32 làm handoff bắt buộc |
| PRD khác Process Flow | Cãi scope | Dùng file 21/22 và revise v1.1 |
| API không khớp UI | FE/BE lệch | OpenAPI-first + contract test |
| DB không khớp Data Dictionary | Dữ liệu sai nghĩa | Review DB against file 05 |
| Workflow thật không khớp To-Be | User không dùng | As-Is + Gap + v1.1 revision |
| SOP lỗi thời | User thao tác sai | SOP update sau mỗi release |
| Permission thiếu | Lộ dữ liệu/sửa sai | 04 + 19 + security test |
| Không có traceability | Không nghiệm thu được | Requirement traceability matrix |
| Change không kiểm soát | Scope creep | File 30 change control |

---

## 17. Next Step đề xuất

Sau khi có file 32, bước hợp lý nhất là làm **Core Docs v1.1 Update Pack**.

Tên đề xuất:

```text
33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md
```

Mục tiêu:

```text
Cập nhật các tài liệu lõi theo quyết định từ As-Is/GAP:
- 03 PRD/SRS
- 04 Permission Matrix
- 05 Data Dictionary
- 06 Process Flow
- 08 Screen List
- 09 UAT
- 16 API Contract
- 17 DB Schema
```

Sau đó mới vào:

```text
34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md
```

---

## 18. Kết luận

Bộ tài liệu ERP hiện tại đã đủ rộng và sâu để chuyển từ thiết kế sang triển khai. Nhưng vì số lượng tài liệu lớn, bắt buộc phải có một file tổng điều phối để tránh đội thi công bị lạc.

File 32 này là tài liệu khóa 3 việc:

```text
1. Ai đọc gì.
2. Requirement nằm ở đâu.
3. Khi xung đột thì tin file nào.
```

Từ thời điểm này, mọi triển khai nên đi theo nguyên tắc:

```text
Không build requirement nào nếu không trace được từ business → flow → data → screen → API → DB → test → sprint.
```

Đó là cách giữ ERP không biến thành một đống chức năng rời rạc.
