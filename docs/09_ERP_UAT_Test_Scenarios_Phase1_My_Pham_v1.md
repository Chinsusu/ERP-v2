# 09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1

**Project:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** UAT Test Scenarios + UAT Execution Guide  
**Scope:** Phase 1  
**Version:** v1.0  
**Date:** 2026-04-23  
**Language:** Vietnamese  
**Related Documents:**  
- ERP_Blueprint_My_Pham_v1  
- 03_ERP_PRD_SRS_Phase1_My_Pham_v1  
- 04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1  
- 05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1  
- 06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1  
- 07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1  
- 08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1  

---

## 1. Mục tiêu tài liệu

Tài liệu này định nghĩa bộ **UAT (User Acceptance Testing)** cho Phase 1 của ERP.  
Mục tiêu không chỉ là kiểm tra “chức năng có chạy hay không”, mà là xác nhận 5 điều quan trọng hơn:

1. Hệ thống có chạy đúng **nghiệp vụ thực tế** của công ty mỹ phẩm không.
2. Dữ liệu có phản ánh đúng **batch, hạn dùng, QC status, tồn khả dụng, công nợ, giá vốn cơ bản** không.
3. Quyền hạn và phê duyệt có ngăn được các lỗi vận hành nguy hiểm không.
4. Báo cáo/KPI có lên đúng và **drill-down về chứng từ gốc** được không.
5. Người dùng nghiệp vụ có đủ tự tin để **go-live** mà không phải quay về Excel như nguồn sự thật chính.

Nói ngắn gọn:

- **SIT/QA nội bộ** xác minh hệ thống hoạt động về mặt kỹ thuật.
- **UAT** xác minh hệ thống dùng được trong đời thật.

---

## 2. Phạm vi UAT của Phase 1

### 2.1. In scope

UAT Phase 1 bao gồm các nhóm chức năng sau:

1. Common / System foundation
2. Master Data
3. Procurement / Purchasing
4. QA/QC
5. Production
6. Warehouse
7. Sales / OMS cơ bản
8. Dashboard, báo cáo và log kiểm soát liên quan đến Phase 1

### 2.2. Out of scope

Các nội dung sau chưa là trọng tâm của tài liệu này:

- HRM đầy đủ;
- CRM nâng cao;
- KOL/Affiliate module hoàn chỉnh;
- POS hoàn chỉnh cho retail;
- kế toán tài chính sâu với bút toán pháp lý chi tiết;
- mobile app riêng;
- automation nâng cao ngoài approval căn bản.

### 2.3. Trọng tâm nghiệp vụ của UAT Phase 1

Do đặc thù ngành mỹ phẩm, UAT phải kiểm tra rất kỹ các trục sau:

- quản lý theo **batch/lô**;
- **expiry date / hạn dùng**;
- **QC hold / pass / fail**;
- **quarantine stock**;
- **reserved stock**;
- **actual consumption vs BOM**;
- **FEFO/FIFO**;
- **partial receipt / partial issue / partial delivery**;
- **return / reverse / cancel có log**;
- **approval theo role và ngưỡng**.

---

## 3. Nguyên tắc UAT

### 3.1. UAT phải dùng dữ liệu thực tế mô phỏng gần nhất

Không test bằng dữ liệu quá sạch hoặc quá lý tưởng.  
Phải có đủ các case:

- batch pass;
- batch hold;
- batch fail;
- hàng cận date;
- nhập hàng nhiều đợt;
- xuất kho một phần;
- công thức có hao hụt thực tế;
- đơn vượt discount threshold;
- hàng trả về;
- vật tư thiếu so với kế hoạch.

### 3.2. UAT phải test cả đường đẹp lẫn đường xấu

Một ERP thường demo rất mượt ở happy path.  
Nhưng doanh nghiệp chết ở exception path.  
Vì vậy UAT bắt buộc bao phủ:

- từ chối duyệt;
- sửa chứng từ sau khi duyệt;
- phân quyền sai vai trò;
- QC fail;
- lô bị hold nhưng người dùng cố xuất;
- near-expiry không được chọn theo rule;
- hàng giao thiếu/hoàn hàng;
- reverse/reopen/reject;
- dữ liệu trùng / sai format / thiếu field bắt buộc.

### 3.3. UAT phải nhìn xuyên chuỗi giá trị

Không test module đơn lẻ là đủ.  
Phải test cả luồng liên thông như:

- PO → Receipt → QC → Available stock;
- Forecast/need → Production order → Material issue → Finished goods receipt;
- SO → Reserve → Shipment → Return;
- Report/KPI → drill down về document.

### 3.4. Không dùng Excel làm chân chống trong UAT

Excel có thể dùng để đối chiếu, nhưng **ERP phải là nguồn sự thật đang được test**.  
Nếu quy trình UAT phụ thuộc vào Excel để hoàn thành giao dịch lõi, đó là dấu hiệu hệ thống chưa đủ chín để go-live.

---

## 4. Vai trò trong UAT

| Vai trò | Trách nhiệm chính |
|---|---|
| UAT Sponsor | Phê duyệt phạm vi UAT, ra quyết định go/no-go |
| UAT Lead | Lập kế hoạch UAT, theo dõi tiến độ, tổng hợp defect và sign-off |
| Module Owner | Chịu trách nhiệm nghiệp vụ của từng module, xác nhận expected result |
| Super User | Thực thi test case, chụp evidence, ghi nhận lỗi |
| BA/PM | Hỗ trợ giải thích flow, cập nhật test case, triage lỗi |
| QA System | Hỗ trợ tái hiện lỗi, xác nhận fix, regression |
| Dev Support | Fix bug, hỗ trợ phân tích root cause |
| Data Owner | Chuẩn bị master data và opening data cho UAT |
| Approver Business | Thực hiện các bước approval để test workflow |

### 4.1. Gợi ý phân công Super User

- Master Data: Sales Admin / Purchasing Admin / System Admin
- Procurement: Purchasing + Finance reviewer
- QA/QC: QA lead + QC operator
- Production: Production planner + supervisor
- Warehouse: Warehouse lead + inventory controller
- Sales: Sales admin + customer service + approver sales
- Reporting: COO / Finance / CEO delegate

---

## 5. Tiêu chí vào UAT (Entry Criteria)

UAT chỉ nên bắt đầu khi đạt tối thiểu các điều kiện sau:

1. Các màn hình và flow Phase 1 đã qua SIT/QA nội bộ ở mức cơ bản.
2. Permission và approval matrix bản Phase 1 đã được cấu hình ở môi trường UAT.
3. Master data seed đã được tải lên đủ để test ít nhất 80% case trọng yếu.
4. Có sẵn dữ liệu batch, expiry, supplier, customer, BOM, warehouse location.
5. Các tích hợp tối thiểu cần cho Phase 1 đã sẵn sàng hoặc có stub/mock ổn định.
6. Đã có defect triage rule và kênh ghi lỗi thống nhất.
7. Người dùng nghiệp vụ tham gia UAT đã được walkthrough cơ bản.

---

## 6. Tiêu chí hoàn thành UAT (Exit Criteria)

Có thể xem là đạt để sign-off UAT khi đồng thời thỏa các điều kiện dưới đây:

1. 100% test case **Critical** đã chạy xong.
2. Ít nhất 95% test case **High** đã chạy xong.
3. Không còn defect mức **Critical** hoặc **High** chưa có hướng xử lý được business chấp nhận.
4. Các luồng end-to-end trọng yếu đã pass:
   - Procure to Available Stock
   - Plan to Produce to Finished Goods
   - Order to Reserve to Ship
   - Return / reverse cơ bản
5. Dashboard/KPI trọng yếu khớp số với chứng từ gốc theo sample kiểm tra.
6. Module owner và UAT sponsor ký sign-off hoặc chấp nhận có điều kiện rõ ràng.

---

## 7. Môi trường UAT và dữ liệu test

### 7.1. Yêu cầu môi trường UAT

- URL môi trường UAT ổn định;
- cấu hình role/permission gần giống production dự kiến;
- email/notification nội bộ hoạt động hoặc có giả lập rõ ràng;
- dữ liệu test có thể reset theo chu kỳ;
- log/audit phải bật;
- timestamp và timezone đồng nhất.

### 7.2. Bộ dữ liệu test tối thiểu

#### 7.2.1. Danh mục hàng hóa

- 5 nguyên vật liệu
- 3 bao bì/phụ liệu
- 2 thành phẩm
- 1 bán thành phẩm nếu flow có dùng
- 2 BOM sản xuất

#### 7.2.2. Kho và vị trí

- Kho nguyên liệu
- Kho thành phẩm
- Kho quarantine/hold
- 3–5 vị trí/bin đại diện

#### 7.2.3. Nhà cung cấp

- 1 NCC active tốt
- 1 NCC hay giao trễ
- 1 NCC bị chặn/suspended để test validation

#### 7.2.4. Khách hàng

- 1 khách B2B có credit term
- 1 khách cash/carry
- 1 khách bị hold công nợ (nếu rule có áp dụng)

#### 7.2.5. Batch và trạng thái QC

- 1 batch NVL pass
- 1 batch NVL hold
- 1 batch NVL fail
- 1 batch thành phẩm pass
- 1 batch thành phẩm hold
- 1 batch gần hết hạn

#### 7.2.6. Người dùng và role

- Admin hệ thống
- Purchasing user
- Purchasing manager/approver
- QA user
- QA approver
- Production planner
- Production supervisor
- Warehouse user
- Warehouse approver (nếu có)
- Sales user
- Sales manager approver
- Finance reviewer
- COO/CEO approver test

### 7.3. Dữ liệu mẫu nên có sẵn

- một PO đủ điều kiện pass;
- một PO dùng để test partial receipt;
- một production order có đủ NVL;
- một production order bị thiếu NVL;
- một sales order còn đủ hàng;
- một sales order gặp batch hold;
- một sales order dùng để test return.

---

## 8. Phân loại mức độ lỗi và ưu tiên xử lý

| Severity | Ý nghĩa |
|---|---|
| Critical | Không thể hoàn thành giao dịch trọng yếu hoặc số liệu sai nghiêm trọng, không có workaround chấp nhận được |
| High | Giao dịch hoàn thành nhưng sai nghiệp vụ lớn, sai quyền/phê duyệt, sai tồn, sai trạng thái |
| Medium | Ảnh hưởng năng suất thao tác, UI/logic phụ, có workaround tạm thời |
| Low | Lỗi nhỏ về trình bày, text, format, hoặc tối ưu chưa hoàn hảo |

| Priority | Ý nghĩa |
|---|---|
| P1 | Phải xử lý trước khi tiếp tục UAT/go-live |
| P2 | Xử lý trong chu kỳ UAT hiện tại |
| P3 | Có thể defer có kiểm soát nếu business chấp nhận |
| P4 | Backlog sau go-live nếu không ảnh hưởng nghiệp vụ cốt lõi |

---

## 9. Chiến lược chạy UAT

### 9.1. Vòng UAT đề xuất

#### Cycle 1 – Module validation
Mục tiêu: kiểm tra theo từng module, rule chính, permission và validation.

#### Cycle 2 – End-to-end validation
Mục tiêu: chạy xuyên phòng ban với dữ liệu gần thực tế.

#### Cycle 3 – Regression + sign-off
Mục tiêu: chạy lại các case đã lỗi và các case critical/high trước khi sign-off.

### 9.2. Cách chấm trạng thái test case

| Trạng thái | Ý nghĩa |
|---|---|
| Not Run | Chưa chạy |
| In Progress | Đang chạy |
| Passed | Đạt đúng expected result |
| Failed | Không đạt |
| Blocked | Không chạy được vì phụ thuộc/lỗi khác |
| Passed with Note | Đạt có điều kiện hoặc workaround được business chấp nhận tạm thời |

### 9.3. Bằng chứng cần lưu

Mỗi test case pass/fail nên có tối thiểu:

- screenshot màn hình trước/sau bước quan trọng;
- số chứng từ tạo ra;
- user/role đã thao tác;
- timestamp;
- nếu fail: mô tả actual result và mức độ ảnh hưởng.

---

## 10. Template test case chuẩn

Sử dụng template sau cho từng test case chi tiết khi đưa vào file thực thi hoặc test management tool:

| Field | Nội dung |
|---|---|
| Test Case ID | Mã test case duy nhất |
| Module | Module nghiệp vụ |
| Scenario Title | Tên tình huống test |
| Business Objective | Kiểm tra mục tiêu nghiệp vụ gì |
| Preconditions | Dữ liệu và trạng thái trước khi test |
| User Role | Vai trò thực thi |
| Steps | Các bước thao tác |
| Expected Result | Kết quả mong đợi |
| Evidence | Screenshot / document no / export |
| Priority | Critical / High / Medium / Low |
| Status | Not Run / Passed / Failed / Blocked |
| Notes | Ghi chú |

---

## 11. Scenario catalog tổng quan

| Nhóm | Prefix | Số case gợi ý |
|---|---|---:|
| System / Permission / Approval | SYS-UAT | 8 |
| Master Data | MDM-UAT | 10 |
| Procurement | PUR-UAT | 12 |
| QA/QC | QAC-UAT | 10 |
| Production | PRD-UAT | 12 |
| Warehouse | WH-UAT | 12 |
| Sales | SAL-UAT | 12 |
| End-to-End / Cross Module | E2E-UAT | 10 |
| Reports / KPI / Audit | RPT-UAT | 8 |

Tổng số case gợi ý trọng tâm: **94** case.  
Không bắt buộc chạy tất cả trong cùng một ngày; nên chia theo cycle và owner.

---

# 12. Test scenarios chi tiết

## 12.1. System / Permission / Approval

### Mục tiêu
Xác minh user chỉ thấy đúng menu, chỉ thao tác đúng quyền, và luồng approval hoạt động theo matrix đã chốt.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| SYS-UAT-001 | Đăng nhập đúng role và thấy đúng menu | Tài khoản đã cấu hình role | Purchasing user / QA user / Warehouse user / Sales user | Mỗi role chỉ thấy portal, menu, quick actions đúng phạm vi | Critical |
| SYS-UAT-002 | User không được truy cập menu trái quyền bằng URL trực tiếp | Có URL màn hình nhạy cảm | User không có quyền | Bị chặn truy cập, có thông báo phù hợp, không lộ dữ liệu | Critical |
| SYS-UAT-003 | Workflow approval theo ngưỡng giá trị hoạt động đúng | Có giao dịch cần duyệt vượt ngưỡng | Maker + Approver | Giao dịch tự động route đúng người duyệt theo rule | Critical |
| SYS-UAT-004 | Maker không được tự duyệt giao dịch do chính mình tạo | Rule SoD đã bật | Maker đồng thời có quyền xem approval list | Hệ thống chặn tự duyệt nếu rule cấm | Critical |
| SYS-UAT-005 | Hủy giao dịch sau khi duyệt phải để lại audit log | Có chứng từ đã approved | Role được phép cancel | Log ghi user, thời gian, hành động, giá trị trước/sau theo mức cho phép | High |
| SYS-UAT-006 | Trường read-only không cho sửa trái quyền | Chứng từ ở trạng thái lock | User thường | Field bị khóa, không lưu được thay đổi | High |
| SYS-UAT-007 | Thông báo/inbox duyệt hiển thị đúng trạng thái | Có giao dịch chờ duyệt | Approver | Approver nhìn thấy đúng document, đúng action pending | Medium |
| SYS-UAT-008 | Export dữ liệu tôn trọng phân quyền | Có dữ liệu tồn kho/PO/SO | User hạn chế | Chỉ export được dữ liệu user được phép xem | High |

---

## 12.2. Master Data

### Mục tiêu
Đảm bảo dữ liệu gốc đủ sạch, đủ rule để các module sau vận hành ổn định.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| MDM-UAT-001 | Tạo mới nguyên vật liệu với field bắt buộc đầy đủ | Có quyền tạo item | Master Data Admin | Tạo thành công khi nhập đủ code, name, UOM, type, shelf-life rule, QC required | Critical |
| MDM-UAT-002 | Chặn tạo item trùng mã | Đã có item code tồn tại | Master Data Admin | Hệ thống chặn duplicate code | Critical |
| MDM-UAT-003 | Chặn tạo item thiếu field bắt buộc | Form tạo item mở | Master Data Admin | Không cho lưu, highlight field thiếu | High |
| MDM-UAT-004 | Tạo thành phẩm có BOM/version liên kết đúng | Có BOM hợp lệ | R&D/Admin | SKU thành phẩm được liên kết BOM/version đúng | Critical |
| MDM-UAT-005 | Inactive item không dùng được cho giao dịch mới | Item đã set inactive | Purchasing / Warehouse / Production / Sales | Item không xuất hiện ở chứng từ mới hoặc bị chặn chọn | High |
| MDM-UAT-006 | Supplier suspended không được dùng tạo PO mới | NCC đã bị suspended | Purchasing user | Hệ thống chặn hoặc cảnh báo theo rule | High |
| MDM-UAT-007 | Customer credit policy hiển thị đúng tại sales order | Có customer với credit term | Sales user | SO hiển thị đúng payment term / status / alert | Medium |
| MDM-UAT-008 | Warehouse/location mapping đúng loại hàng | Có kho NVL, kho FG, kho quarantine | Admin / Warehouse | Chỉ chọn được kho hợp lệ theo loại giao dịch | High |
| MDM-UAT-009 | Batch code auto-format đúng quy tắc | Rule mã batch đã cấu hình | QA / Warehouse / Production | Batch sinh đúng format, không trùng | Critical |
| MDM-UAT-010 | Lịch sử thay đổi master data có log | Có chỉnh sửa item/supplier/customer | Admin | Audit log lưu trước/sau và user thay đổi | High |

---

## 12.3. Procurement / Purchasing

### Mục tiêu
Kiểm tra trục từ đề nghị mua đến PO, nhận hàng, partial receipt và liên kết QC.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| PUR-UAT-001 | Tạo Purchase Request hợp lệ | Có item active, supplier, cost center | Purchasing user / requester | PR tạo thành công, sinh số chứng từ, status Draft/Submitted đúng | Critical |
| PUR-UAT-002 | PR thiếu justification hoặc thiếu item bị chặn submit | Form PR | Requester | Submit không thành công nếu thiếu field bắt buộc | High |
| PUR-UAT-003 | PR vượt ngưỡng được route đúng approver | Threshold đã cấu hình | Requester + Approver | Approval route đúng line manager/finance/COO theo rule | Critical |
| PUR-UAT-004 | Maker không sửa PR sau khi approved ngoài field được phép | PR đã approved | Requester | Hệ thống khóa theo rule, chỉ cho revise qua luồng phù hợp | High |
| PUR-UAT-005 | Tạo PO từ PR đã duyệt | Có PR approved | Purchasing user | PO copy đúng item, qty, supplier, expected date | Critical |
| PUR-UAT-006 | PO cho NCC suspended bị chặn | NCC suspended | Purchasing user | Không cho save/approve PO | High |
| PUR-UAT-007 | Nhận hàng đủ số lượng từ PO | PO approved | Warehouse receiver | GRN tạo thành công, cập nhật received qty, chưa vào available stock trước QC nếu rule yêu cầu | Critical |
| PUR-UAT-008 | Partial receipt từ PO nhiều đợt | PO qty > 0 | Warehouse receiver | Hệ thống ghi nhận partial received, còn open qty đúng | Critical |
| PUR-UAT-009 | Nhận hàng vượt số lượng PO bị chặn hoặc cần override được kiểm soát | PO đã có qty | Warehouse receiver | Validation đúng theo tolerance rule | High |
| PUR-UAT-010 | Sau receipt, item vào trạng thái chờ QC/quarantine đúng | QC required = Yes | Warehouse/QA | Stock chưa vào available, nằm ở trạng thái hold/quarantine | Critical |
| PUR-UAT-011 | Cancel phần PO còn mở sau partial receipt | PO partially received | Purchasing manager | Close/cancel phần dư đúng, không ảnh hưởng số đã nhận | Medium |
| PUR-UAT-012 | Drill-down từ PO sang receipts hoạt động đúng | Có PO và GRN liên quan | Purchasing / Finance | Xem được danh sách receipt liên quan và số lượng đã nhận | Medium |

---

## 12.4. QA/QC

### Mục tiêu
Bảo đảm không có lô chưa đạt chất lượng lọt vào tồn khả dụng hoặc sang công đoạn kế tiếp.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| QAC-UAT-001 | Tạo phiếu kiểm nghiệm đầu vào cho batch NVL vừa nhận | Có GRN chờ QC | QC user | Tạo phiếu QC gắn đúng batch, supplier, receipt ref | Critical |
| QAC-UAT-002 | Batch QC Pass chuyển đúng sang available stock | Có phiếu QC chờ kết quả | QC approver | Batch chuyển sang Pass, stock khả dụng tăng đúng | Critical |
| QAC-UAT-003 | Batch QC Hold không được dùng cho sản xuất/xuất bán | Có batch status Hold | Warehouse / Production / Sales | Hệ thống chặn chọn batch hold trong các giao dịch downstream | Critical |
| QAC-UAT-004 | Batch QC Fail chuyển về quarantine/disposal theo rule | Có batch fail | QC approver / Warehouse | Batch fail không vào available, trạng thái/quarantine đúng | Critical |
| QAC-UAT-005 | Chỉ user được phân quyền mới đổi QC status | Có phiếu QC | User thường và QA approver | User thường không đổi được; approver đổi được và có log | High |
| QAC-UAT-006 | Không cho chỉnh sửa QC result sau release nếu không qua exception flow | Batch đã release | QC user | Hệ thống khóa hoặc bắt dùng deviation/reopen flow | High |
| QAC-UAT-007 | Attachment hồ sơ QC lưu và truy xuất được | Có file COA/checksheet | QC user | File đính kèm đúng chứng từ, mở được | Medium |
| QAC-UAT-008 | Batch thành phẩm hold không được reserve cho sales order | Có FG batch hold | Sales / Warehouse | Batch không xuất hiện trong allocation hợp lệ | Critical |
| QAC-UAT-009 | Drill-down từ batch QC sang receipt/production batch đúng | Có dữ liệu liên kết | QA / COO | Trace được nguồn batch về upstream document | High |
| QAC-UAT-010 | Nhật ký thay đổi QC status đầy đủ | Có batch đã đổi trạng thái | QA / Audit | Log đầy đủ user, thời gian, old/new status, note | High |

---

## 12.5. Production

### Mục tiêu
Xác minh lệnh sản xuất, cấp phát nguyên liệu, tiêu hao thực tế, nhập thành phẩm và logic BOM chạy đúng.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| PRD-UAT-001 | Tạo Production Order từ nhu cầu hợp lệ | Có BOM active, đủ quyền | Planner | PO sản xuất tạo thành công, đúng item, qty, BOM version | Critical |
| PRD-UAT-002 | Không cho tạo production order khi BOM inactive/thiếu | BOM lỗi/inactive | Planner | Hệ thống chặn tạo hoặc cảnh báo theo rule | High |
| PRD-UAT-003 | Issue nguyên liệu đủ theo BOM từ batch pass | Có NVL available/pass | Warehouse / Production | Material issue thành công, trừ đúng tồn khả dụng | Critical |
| PRD-UAT-004 | Chặn issue nguyên liệu từ batch Hold/Fail | Có batch hold/fail | Warehouse / Production | Không thể issue batch không hợp lệ | Critical |
| PRD-UAT-005 | Chặn issue vượt tồn khả dụng | Available stock thấp hơn yêu cầu | Warehouse | Validation chặn, không cho âm tồn khả dụng | Critical |
| PRD-UAT-006 | Ghi nhận actual consumption khác BOM và tính variance | Lệnh sản xuất đang chạy | Production supervisor | Hệ thống lưu actual, thể hiện variance rõ ràng | High |
| PRD-UAT-007 | Nhập bán thành phẩm/thành phẩm từ lệnh sản xuất | Lệnh có output | Production / Warehouse | Goods receipt thành công, sinh batch FG đúng format | Critical |
| PRD-UAT-008 | Thành phẩm mới nhập vào trạng thái chờ QC theo rule | FG receipt vừa tạo | Production / QA | FG chưa available nếu cần QC release | Critical |
| PRD-UAT-009 | Đóng lệnh sản xuất khi còn issue/output lệch ngưỡng bị cảnh báo | Lệnh có chênh lệch lớn | Planner / Supervisor | Hệ thống cảnh báo hoặc chặn đóng theo tolerance | High |
| PRD-UAT-010 | Hủy lệnh sản xuất chưa chạy để lại log và reverse reservation/issue đúng | Lệnh chưa hoàn tất | Planner / Approver | Cancel/reverse đúng, tồn kho được hoàn trả theo rule | High |
| PRD-UAT-011 | Trace từ FG batch về NVL batch đã issue | Có dữ liệu sản xuất hoàn tất | QA / Production / Audit | Truy xuất được batch genealogy | Critical |
| PRD-UAT-012 | Dashboard sản xuất hiển thị đúng status lệnh | Có nhiều lệnh các trạng thái | Planner / COO | Status Open/In Progress/Completed/Closed hiển thị đúng | Medium |

---

## 12.6. Warehouse

### Mục tiêu
Kiểm tra nhập/xuất/chuyển/kiểm kê theo tồn vật lý và tồn khả dụng, bám chặt batch, expiry và QC status.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| WH-UAT-001 | Xem stock ledger theo item-batch-location chính xác | Có giao dịch nhập/xuất | Warehouse lead | Ledger hiển thị đủ opening, in, out, balance | Critical |
| WH-UAT-002 | Available stock khác physical stock khi có hold/reserved | Có batch hold và stock reserved | Warehouse / Sales | Số liệu physical và available hiển thị đúng công thức | Critical |
| WH-UAT-003 | Chuyển kho nội bộ hợp lệ giữa location cùng loại | Có stock pass | Warehouse user | Transfer thành công, ledger hai bên cập nhật đúng | High |
| WH-UAT-004 | Chặn chuyển kho batch fail/hold sang kho khả dụng | Có batch fail/hold | Warehouse user | Hệ thống chặn hoặc route theo quarantine flow | Critical |
| WH-UAT-005 | FEFO đề xuất batch gần hết hạn trước khi xuất | Có nhiều batch pass với expiry khác nhau | Warehouse / Sales | Hệ thống gợi ý batch FEFO đúng theo rule | Critical |
| WH-UAT-006 | Chặn xuất batch đã hết hạn | Có batch expired | Warehouse | Không cho allocate/pick/issue batch hết hạn | Critical |
| WH-UAT-007 | Reserve stock cho sales order làm giảm available stock | Có SO confirmed | Sales / Warehouse | Available stock giảm, physical stock chưa giảm cho tới khi shipment | Critical |
| WH-UAT-008 | Hủy reserve khôi phục available stock đúng | Có SO cancelled hoặc unreserve | Sales / Warehouse | Available stock tăng lại đúng | High |
| WH-UAT-009 | Stock count và adjustment có approval/audit | Có cycle count | Warehouse + approver | Chênh lệch kiểm kê được ghi nhận, approval đúng, audit log đầy đủ | High |
| WH-UAT-010 | Quarantine stock không lẫn với available stock trên dashboard | Có stock ở quarantine | Warehouse / COO | Dashboard và report tách đúng | High |
| WH-UAT-011 | Theo dõi stock near-expiry đúng ngưỡng ngày | Có batch gần hết hạn | Warehouse / Sales / COO | Danh sách near-expiry lên đúng theo threshold cấu hình | High |
| WH-UAT-012 | Drill-down từ stock summary tới document nguồn | Có data ledger | Warehouse / Finance | Từ balance mở về được GRN/issue/transfer/shipment liên quan | Medium |

---

## 12.7. Sales / OMS cơ bản

### Mục tiêu
Đảm bảo đơn hàng bán ra chỉ dùng tồn hợp lệ, đi qua đúng giá/discount/approval/reservation/shipment và hỗ trợ return cơ bản.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| SAL-UAT-001 | Tạo sales order với customer hợp lệ và item active | Có customer, price list, stock | Sales user | SO tạo thành công, giá mặc định đúng | Critical |
| SAL-UAT-002 | Chặn tạo SO với item inactive hoặc customer blocked | Item/customer blocked | Sales user | Hệ thống chặn hoặc cảnh báo theo rule | High |
| SAL-UAT-003 | Giá và discount mặc định lên đúng theo rule | Có price list / discount matrix | Sales user | Giá/chiết khấu tự động đúng | Critical |
| SAL-UAT-004 | Discount vượt ngưỡng route approval đúng | Threshold đã cấu hình | Sales user + manager | SO không xác nhận được nếu chưa duyệt; route đúng approver | Critical |
| SAL-UAT-005 | Confirm SO tạo reservation đúng | Có stock available | Sales user | Reserved stock tăng, available stock giảm đúng | Critical |
| SAL-UAT-006 | Chặn confirm SO khi không đủ available stock | Stock không đủ | Sales user | Không confirm hoặc chỉ confirm phần cho phép theo rule | Critical |
| SAL-UAT-007 | Allocation chỉ dùng batch Pass và chưa hết hạn | Có batch pass/hold/expired | Sales / Warehouse | Chỉ batch hợp lệ được allocate | Critical |
| SAL-UAT-008 | Shipment từ SO cập nhật shipped qty và giảm physical stock | Có SO reserved | Warehouse shipper | Shipment thành công, ledger giảm đúng | Critical |
| SAL-UAT-009 | Partial delivery giữ trạng thái open balance đúng | SO qty lớn | Warehouse / Sales | Giao một phần, remaining qty còn mở đúng | High |
| SAL-UAT-010 | Cancel SO trước shipment giải phóng reservation | SO confirmed, chưa ship | Sales / Approver | Reserved stock hoàn lại, status cancelled đúng | High |
| SAL-UAT-011 | Return hàng bán cập nhật tồn và trạng thái QC/inspection phù hợp | Có shipment trước đó | Sales / Warehouse / QA | Returned stock đi đúng flow (quarantine hoặc return inspection), không tự động vào available nếu rule không cho | High |
| SAL-UAT-012 | Trace từ sales order tới shipment và batch đã giao | Có SO hoàn tất | Sales / CS / QA | Drill-down đầy đủ từ order sang shipment và batch | Medium |

---

## 12.8. End-to-End / Cross Module

### Mục tiêu
Đây là nhóm test quan trọng nhất. Nó chứng minh hệ thống không chỉ đúng từng module mà còn đúng cả chuỗi giá trị.

| ID | Scenario | Preconditions | Vai trò tham gia | Expected Result | Priority |
|---|---|---|---|---|---|
| E2E-UAT-001 | Procure to Available Stock – nhập NVL pass QC | Có PR/PO mẫu | Purchasing + Warehouse + QC | Từ PR/PO đến GRN, QC Pass, stock vào available đúng | Critical |
| E2E-UAT-002 | Procure to Quarantine – nhập NVL fail QC | Có PO mẫu | Purchasing + Warehouse + QC | Batch fail không vào available, ở quarantine đúng | Critical |
| E2E-UAT-003 | Plan to Produce – issue NVL pass, nhận FG chờ QC | Có BOM, NVL pass | Planner + Warehouse + Production + QA | Flow sản xuất chạy xuyên suốt, FG batch sinh đúng, chờ QC đúng | Critical |
| E2E-UAT-004 | FG QC Pass rồi mới bán được | Có FG batch chờ QC | QA + Sales + Warehouse | Trước Pass: không bán được; sau Pass: reserve/ship được | Critical |
| E2E-UAT-005 | Batch Hold chặn sales allocation dù vẫn còn physical stock | Có FG batch hold | Sales + Warehouse + QA | Available không tính batch hold, SO không allocate được batch đó | Critical |
| E2E-UAT-006 | Partial receipt NVL rồi sản xuất vẫn dùng đúng phần đã pass | PO nhận nhiều đợt | Purchasing + Warehouse + QA + Production | Chỉ phần đã pass mới issue vào sản xuất | High |
| E2E-UAT-007 | Partial delivery sales order và return một phần | Có SO qty lớn | Sales + Warehouse + QA | Shipment partial, return partial, tồn và trạng thái order đúng | High |
| E2E-UAT-008 | Near-expiry batch được FEFO đề xuất khi giao hàng | Có nhiều FG batch với expiry khác nhau | Sales + Warehouse | Batch gần hết hạn hợp lệ được ưu tiên chọn | High |
| E2E-UAT-009 | Drill-down từ KPI tồn kho về batch và chứng từ nguồn | Có dashboard/report | COO / Warehouse / Finance | Số tổng khớp với ledger và document source | High |
| E2E-UAT-010 | Reverse/cancel có audit log xuyên chuỗi | Có PR/PO/SO/transfer đã tạo | Business owner + Audit | Mọi hành động cancel/reverse có log và không làm lệch ledger | High |

---

## 12.9. Reports / KPI / Audit

### Mục tiêu
Đảm bảo số liệu báo cáo không chỉ đẹp mà còn đúng bản chất nghiệp vụ.

| ID | Scenario | Preconditions | User Role | Expected Result | Priority |
|---|---|---|---|---|---|
| RPT-UAT-001 | Stock summary khớp stock ledger theo item-batch | Có giao dịch đầy đủ | Warehouse / Finance | Tổng tồn summary = ledger balance | Critical |
| RPT-UAT-002 | Available stock report khớp công thức physical - hold - reserved - issue pending | Có hold/reserved | Warehouse / Sales | Công thức và số hiển thị đúng | Critical |
| RPT-UAT-003 | Near-expiry report lên đúng batch theo threshold ngày | Có batch expiry khác nhau | Warehouse / COO | Báo cáo lọc đúng | High |
| RPT-UAT-004 | Open PO / received qty / pending qty report khớp document | Có partial receipt | Purchasing | Số lượng mở còn lại đúng | High |
| RPT-UAT-005 | Production variance report thể hiện actual vs standard | Có lệnh có variance | Production / COO | Số variance đúng theo data thực tế | High |
| RPT-UAT-006 | Sales order status report khớp confirmed/reserved/shipped/returned | Có SO nhiều trạng thái | Sales / COO | Report đúng trạng thái và số lượng | High |
| RPT-UAT-007 | Audit log tra cứu được thay đổi quan trọng | Có activity log | Audit / Admin | Tìm được log create/approve/cancel/status change | High |
| RPT-UAT-008 | Dashboard drill-down từ chỉ số xuống document source | Có dashboard | COO / CEO delegate | Bấm từ KPI mở được danh sách chứng từ và document detail | High |

---

# 13. Kịch bản UAT ưu tiên cao nhất để chạy trong 1–2 ngày đầu

Nếu thời gian hạn chế, bắt buộc phải chạy trước các case dưới đây:

1. SYS-UAT-001, 002, 003, 004  
2. MDM-UAT-001, 002, 004, 009  
3. PUR-UAT-005, 007, 008, 010  
4. QAC-UAT-002, 003, 004, 008  
5. PRD-UAT-003, 004, 007, 011  
6. WH-UAT-002, 005, 006, 007  
7. SAL-UAT-003, 004, 005, 006, 007, 008  
8. E2E-UAT-001, 002, 003, 004, 005  
9. RPT-UAT-001, 002, 008

Đây là “xương sống go-live”.  
Nếu nhóm này không pass, chưa nên nói đến triển khai thực tế.

---

# 14. Checklist dữ liệu và cấu hình trước khi chạy UAT

## 14.1. Checklist cấu hình

- [ ] Role và permission đã đúng theo matrix Phase 1  
- [ ] Approval threshold đã cấu hình ở UAT  
- [ ] Rule batch/expiry/QC status đã bật  
- [ ] Tolerance nhập hàng / xuất hàng / variance đã cấu hình  
- [ ] FEFO/FIFO rule đã xác định rõ  
- [ ] Notification/inbox duyệt có hoạt động hoặc có mô phỏng rõ  

## 14.2. Checklist dữ liệu

- [ ] Có ít nhất 2 thành phẩm active  
- [ ] Có BOM active gắn đúng thành phẩm  
- [ ] Có 5 nguyên vật liệu pass QC  
- [ ] Có ít nhất 1 batch hold và 1 batch fail  
- [ ] Có 1 batch gần hết hạn  
- [ ] Có ít nhất 2 supplier active  
- [ ] Có ít nhất 2 customer với điều kiện bán hàng khác nhau  
- [ ] Có tồn kho opening đủ để test sales/warehouse  

---

# 15. Quy trình xử lý defect trong UAT

## 15.1. Thông tin defect tối thiểu

Mỗi defect cần ghi:

- Defect ID
- Module
- Test case liên quan
- Môi trường
- Người phát hiện
- Ngày giờ phát hiện
- Bước tái hiện
- Kết quả thực tế
- Kết quả mong đợi
- Severity
- Evidence (screenshot/video/document no)
- Trạng thái fix

## 15.2. Trạng thái defect đề xuất

| Trạng thái | Ý nghĩa |
|---|---|
| New | Mới ghi nhận |
| Triaged | Đã phân loại và xác định owner |
| In Fix | Dev đang xử lý |
| Ready for Retest | Đã fix xong, chờ kiểm tra lại |
| Retest Passed | Fix đạt |
| Retest Failed | Fix chưa đạt |
| Deferred | Chưa xử lý trong phase/cycle này, có phê duyệt business |
| Won’t Fix | Không sửa có lý do được chấp nhận |
| Closed | Hoàn tất |

## 15.3. Rule ra quyết định nhanh

- **Critical/P1:** xử lý ngay, không để tích lũy.  
- **High/P2:** vào sprint fix gần nhất của UAT.  
- **Medium/P3:** đánh giá theo ảnh hưởng go-live.  
- **Low/P4:** gom tối ưu sau nếu không ảnh hưởng nghiệp vụ.

---

# 16. Mẫu kế hoạch chạy UAT gợi ý

## Ngày 1
- Walkthrough test environment và data set
- Chạy SYS + MDM + PUR trọng yếu
- Ghi defect nóng

## Ngày 2
- Chạy QAC + PRD trọng yếu
- Chạy WH luồng tồn khả dụng, FEFO, quarantine

## Ngày 3
- Chạy SAL + E2E flow từ sản xuất đến giao hàng
- Chạy report/kpi đối chiếu

## Ngày 4
- Retest defect P1/P2
- Regression case critical

## Ngày 5
- Chạy sign-off round
- Chốt open issues, workaround, go/no-go recommendation

---

# 17. Mẫu sign-off UAT theo module

## 17.1. Thông tin sign-off

- Module:  
- UAT Owner:  
- Date:  
- Số test case đã chạy:  
- Passed:  
- Failed:  
- Blocked:  
- Passed with Note:  
- Số defect Critical/High còn mở:  
- Điều kiện chấp nhận có điều kiện (nếu có):  

## 17.2. Kết luận module

Chọn một trong ba trạng thái:

- [ ] **Accepted** – đủ điều kiện sang regression/sign-off tổng  
- [ ] **Accepted with Conditions** – chấp nhận có điều kiện ghi rõ  
- [ ] **Rejected** – chưa đủ điều kiện  

## 17.3. Chữ ký

- Module Owner:  
- UAT Lead:  
- PM/BA:  
- Sponsor/Approver:  

---

# 18. Go / No-Go recommendation framework

Sau khi hoàn tất UAT, quyết định go-live nên dựa trên bảng sau:

| Tiêu chí | Go | Conditional Go | No-Go |
|---|---|---|---|
| Critical defect | 0 | 0 | >0 |
| High defect | 0–2 có workaround chấp nhận được | 3–5 có kế hoạch fix rõ | >5 hoặc ảnh hưởng luồng lõi |
| E2E core flows | Pass hết | Pass có note nhỏ | Có flow không pass |
| Data integrity | Khớp | Lệch nhỏ có điều chỉnh được | Lệch lớn, không tin được số liệu |
| User confidence | Cao | Trung bình, cần hỗ trợ sát | Thấp, user từ chối dùng |
| Approval/permission | Ổn | Có ngoại lệ nhỏ | Hỏng SoD/quyền nghiêm trọng |

---

# 19. Kết luận

Tài liệu UAT này được thiết kế để trả lời câu hỏi quan trọng nhất trước khi go-live:

**Hệ thống ERP Phase 1 có thực sự giúp công ty mỹ phẩm vận hành an toàn, kiểm soát được hàng, batch, QC, tồn kho, sản xuất và đơn hàng hay chưa?**

Nếu chỉ test nút bấm, UAT sẽ cho cảm giác “gần xong”.  
Nếu test đúng theo tài liệu này, mày sẽ nhìn ra thứ thật sự quan trọng:

- quy trình nào đã khóa được thất thoát;
- điểm nào còn có thể làm sai tồn, sai batch, sai QC;
- chỗ nào user còn mơ hồ;
- chỗ nào chưa đủ an toàn để đưa lên production.

Một ERP mạnh không phải là ERP ít lỗi vặt.  
Mà là ERP mà **lỗi lớn, lỗi chết người, lỗi làm rò máu doanh nghiệp đã bị lôi ra ánh sáng trước khi go-live**.

---

## 20. Phụ lục A – Danh sách 15 case “đinh” phải demo cho lãnh đạo trước sign-off

1. Tạo PO → receipt → QC Pass → stock available tăng đúng  
2. Tạo PO → receipt → QC Fail → stock không vào available  
3. Tạo production order → issue NVL Pass → FG receipt → FG chờ QC  
4. FG Pass rồi mới allocate bán được  
5. Batch Hold không được bán dù còn physical stock  
6. FEFO chọn đúng batch gần hết hạn hợp lệ  
7. Không xuất được batch hết hạn  
8. Discount vượt ngưỡng phải duyệt  
9. SO confirm làm giảm available stock  
10. Shipment làm giảm physical stock  
11. Return hàng không tự động vào available nếu chưa inspect  
12. Stock summary khớp stock ledger  
13. Trace từ FG batch về NVL batch đã issue  
14. Audit log ghi đầy đủ create/approve/cancel/status change  
15. Dashboard drill-down được về chứng từ gốc

---

## 21. Phụ lục B – Danh sách test data mẫu gợi ý

### Finished Goods
- FG-001: Serum Vitamin C 30ml
- FG-002: Gel rửa mặt dịu nhẹ 100ml

### Raw Materials
- RM-001: Vitamin C active
- RM-002: Base gel
- RM-003: Hương liệu A
- RM-004: Preservative B
- RM-005: Purified water

### Packaging
- PK-001: Chai thủy tinh 30ml
- PK-002: Nắp dropper
- PK-003: Hộp giấy

### Warehouses
- WH-RM: Kho nguyên liệu
- WH-FG: Kho thành phẩm
- WH-QA: Kho quarantine/hold

### Sample Batches
- RM-001-B01: Pass
- RM-002-B02: Hold
- RM-003-B03: Fail
- FG-001-B10: Pass, expiry xa
- FG-001-B11: Pass, near-expiry
- FG-002-B20: Hold

### Sample Suppliers
- SUP-001: NCC active chuẩn
- SUP-002: NCC active hay trễ
- SUP-003: NCC suspended

### Sample Customers
- CUS-001: Đại lý miền Nam – credit 30 ngày
- CUS-002: Khách thanh toán ngay
- CUS-003: Customer on hold

---

## 22. Phụ lục C – Handoff sau khi hoàn tất tài liệu này

Sau tài liệu UAT, bước tiếp theo nên làm là:

1. Chuẩn hóa file thực thi test case chi tiết (Excel/TestRail/Jira Xray tùy cách vận hành).  
2. Tạo checklist data migration rehearsal.  
3. Chuẩn bị **10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1**.  
4. Chuẩn bị tài liệu training theo role cho super user trước UAT thật.  

