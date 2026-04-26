
# 07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1

**Project:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** Report & KPI Catalog  
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

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt **danh mục báo cáo, dashboard và KPI trọng yếu** cho ERP Phase 1 của công ty mỹ phẩm.  
Mục tiêu là để tất cả các bên cùng nhìn một bộ số thống nhất:

- Business biết hệ thống phải cho ra **những báo cáo nào để điều hành**.
- BA biết **các chỉ số cần đặc tả** trong tài liệu nghiệp vụ.
- UI/UX biết **màn hình nào là dashboard, màn hình nào là báo cáo drill-down**.
- Dev/BI biết **nguồn dữ liệu, công thức và chiều dữ liệu chuẩn**.
- Tester biết **đối chiếu con số thế nào khi UAT**.
- CEO và trưởng bộ phận biết **mỗi buổi sáng phải nhìn vào đâu để ra quyết định**.

Nói ngắn gọn:

**PRD/SRS trả lời “hệ thống phải có gì”.**  
**Process Flow trả lời “quy trình sẽ chạy thế nào”.**  
**Report & KPI Catalog trả lời “chạy xong thì phải đo cái gì để biết công ty đang khỏe hay đang rò máu”.**

---

## 2. Phạm vi Phase 1

### 2.1. Các module trong phạm vi báo cáo

Phase 1 tập trung vào 6 module lõi:

1. Dữ liệu gốc (Master Data)  
2. Mua hàng (Procurement)  
3. QA/QC  
4. Sản xuất (Production)  
5. Kho hàng (Warehouse / WMS cơ bản)  
6. Bán hàng (Sales / OMS cơ bản)  

### 2.2. Các chủ đề báo cáo phải phục vụ

- Điều hành tổng quan cho CEO/COO
- Mua hàng và nhà cung cấp
- Chất lượng nguyên liệu, bán thành phẩm, thành phẩm
- Hiệu suất kế hoạch và thực thi sản xuất
- Tồn kho, batch, cận date, hàng hold
- Đơn hàng, fill rate, doanh thu, hoàn trả
- Một lớp tài chính điều hành cơ bản:
  - giá trị tồn kho
  - công nợ cơ bản nếu có trong Phase 1
  - doanh thu thuần
  - lợi nhuận gộp gần đúng

### 2.3. Ngoài phạm vi chi tiết của tài liệu này

Các mảng sau chỉ ghi nhận ở mức định hướng, chưa đi sâu trong Phase 1:

- CRM nâng cao và vòng đời khách hàng
- KOL / Affiliate attribution
- HRM / payroll
- Accounting posting chi tiết chuẩn mực kế toán
- BI nâng cao như forecasting AI, cohort, anomaly detection

---

## 3. Nguyên tắc thiết kế báo cáo và KPI

### 3.1. Một nguồn sự thật chung

Mọi báo cáo phải lấy từ **dữ liệu giao dịch trong ERP**, không dùng Excel thủ công làm nguồn chính.  
Excel chỉ được phép dùng cho:
- import dữ liệu ban đầu
- đối soát ngoại lệ
- kiểm thử

### 3.2. Từ dashboard phải drill-down được đến chứng từ

Mọi số tổng quan quan trọng phải click xuống được:
- chứng từ mua hàng
- phiếu QC
- lệnh sản xuất
- phiếu nhập/xuất kho
- đơn bán hàng
- batch/lô liên quan

Nếu không drill-down được, dashboard rất dễ đẹp nhưng vô dụng.

### 3.3. Tách bạch rõ các loại tồn kho

Hệ thống và báo cáo phải phân biệt rõ:

- **Physical Stock**: tồn vật lý đang có
- **Available Stock**: tồn có thể bán/có thể dùng
- **Reserved Stock**: tồn đã giữ cho đơn hàng hoặc kế hoạch
- **Quarantine / Hold Stock**: tồn đang bị khóa do QC hoặc điều tra
- **Near-Expiry Stock**: tồn cận date
- **Sample / Tester / Gift Stock**: tồn không dùng để bán thương mại

### 3.4. Batch và hạn dùng là chiều dữ liệu bắt buộc

Với ngành mỹ phẩm, các báo cáo quan trọng phải hỗ trợ phân tích theo:
- SKU
- Batch/Lot
- MFG Date
- Expiry Date
- QC Status
- Warehouse
- Channel
- Supplier

### 3.5. KPI phải có định nghĩa và chủ sở hữu

Mỗi KPI phải có:
- công thức chuẩn
- nguồn dữ liệu
- tần suất refresh
- owner nghiệp vụ
- ý nghĩa quản trị
- ngưỡng cảnh báo gợi ý

### 3.6. Ưu tiên báo cáo phục vụ quyết định

Không xây “nghĩa địa dashboard”.  
Phase 1 chỉ ưu tiên những báo cáo trả lời được các câu hỏi thật sự sống còn:

- Có đủ nguyên liệu để chạy lệnh không?
- Batch nào đang hold hoặc sắp hết hạn?
- Đơn nào đang tắc giao?
- Sản xuất có vượt định mức không?
- NCC nào giao trễ hoặc hàng lỗi?
- Bán được nhiều nhưng có lời không?
- Tồn kho nào đang chôn tiền?

---

## 4. Nhịp quyết định theo vai trò

| Vai trò | Nhịp xem chính | Câu hỏi phải trả lời |
|---|---|---|
| CEO / Chủ doanh nghiệp | Hàng ngày, hàng tuần, cuối tháng | Hàng, tiền, rủi ro, lãi, tắc nghẽn ở đâu |
| COO / Head Operations | Hàng ngày | Sản xuất, kho, QC, mua hàng có đồng bộ không |
| Purchasing Manager | Hàng ngày | Cái gì sắp thiếu, NCC nào trễ, giá mua biến động thế nào |
| QA/QC Manager | Hàng ngày | Batch nào hold/fail, CAPA nào quá hạn |
| Production Manager | Theo ca, hàng ngày | Lệnh nào trễ, chuyền nào hụt năng suất, hao hụt có bất thường không |
| Warehouse Manager | Theo giờ, hàng ngày | Còn bao nhiêu tồn khả dụng, batch nào cận date, vị trí nào lệch tồn |
| Sales Manager | Theo giờ, hàng ngày | Đơn nào chưa fill được, kênh nào đang bán tốt, return ở đâu tăng |
| Finance / Kế toán điều hành | Hàng ngày, cuối tuần | Doanh thu thuần, tồn kho theo giá trị, PO/đơn hàng ảnh hưởng tiền ra tiền vào ra sao |

---

## 5. Danh mục báo cáo tổng quan

### 5.1. Danh sách báo cáo/dashboards đề xuất

| Code | Tên báo cáo / dashboard | Nhóm người dùng chính | Tần suất | Priority |
|---|---|---|---|---|
| EXE-01 | Dashboard Điều Hành CEO 360 | CEO, COO, Finance Lead | Hàng giờ / ngày | P0 |
| EXE-02 | Doanh thu & Lợi nhuận Gần Đúng theo Kênh / SKU | CEO, Sales Lead, Finance Lead | Ngày / tuần | P0 |
| EXE-03 | Sức Khỏe Tồn Kho, Batch và Cận Date | CEO, COO, Warehouse Lead | Ngày | P0 |
| PUR-01 | Purchase Request / PO Pipeline | Purchasing, COO | Theo giờ / ngày | P0 |
| PUR-02 | Supplier OTIF, Chất Lượng và Biến Động Giá | Purchasing, COO, Finance Lead | Tuần | P1 |
| PUR-03 | Rủi Ro Thiếu Nguyên Liệu / Replenishment Risk | Purchasing, Planning, COO | Ngày | P0 |
| QUA-01 | Tổng Quan QC Đầu Vào | QA/QC, Purchasing, Warehouse | Ngày | P0 |
| QUA-02 | Batch Quality & Release Aging | QA/QC, Production, COO | Theo giờ / ngày | P0 |
| QUA-03 | Non-conformance / CAPA / Complaint Linkage | QA/QC, COO | Tuần | P1 |
| PRO-01 | Kế Hoạch vs Thực Tế Sản Xuất | Production, COO | Theo ca / ngày | P0 |
| PRO-02 | Hao Hụt, Yield và Variance NVL | Production, QA, Finance Lead | Theo ca / ngày | P0 |
| PRO-03 | Work Order Exception Board | Production, Planning, COO | Theo giờ | P1 |
| WHS-01 | Stock Position & Availability | Warehouse, Sales Admin, COO | Theo giờ | P0 |
| WHS-02 | Inventory Aging / Near-Expiry / Hold Stock | Warehouse, QA, COO, CEO | Ngày | P0 |
| WHS-03 | Stock Accuracy / Cycle Count Variance | Warehouse, Finance Lead | Tuần | P1 |
| SAL-01 | Sales Order Pipeline & Fill Rate | Sales, Sales Admin, COO | Theo giờ / ngày | P0 |
| SAL-02 | Sales by Channel / SKU / Customer | Sales Lead, CEO, Finance Lead | Ngày / tuần | P0 |
| SAL-03 | Returns / Refund / Complaint Summary | Sales, CSKH, QA, CEO | Ngày / tuần | P1 |
| XFN-01 | Batch Traceability Cockpit | QA, Warehouse, Production, COO | Khi cần / hàng ngày | P1 |
| XFN-02 | Working Capital Snapshot cơ bản | CEO, Finance Lead | Ngày / tuần | P1 |
| XFN-03 | Master Data Completeness & Data Quality | PM, Data Owner, COO | Tuần | P1 |

### 5.2. Quy ước mức ưu tiên

- **P0:** bắt buộc có khi go-live hoặc trong 30 ngày đầu sau go-live  
- **P1:** rất nên có trong 30–60 ngày sau go-live  
- **P2:** để dành cho Phase 2 hoặc khi dữ liệu đã ổn định  

---

## 6. Chiều dữ liệu chuẩn dùng chung cho báo cáo

Để toàn bộ hệ thống nói cùng một ngôn ngữ, các báo cáo phải dùng cùng một bộ dimension cốt lõi.

### 6.1. Thời gian
- Ngày
- Tuần
- Tháng
- Quý
- Ca sản xuất
- Giờ ghi nhận giao dịch

### 6.2. Tổ chức
- Công ty / pháp nhân (nếu có)
- Brand
- Nhà máy
- Kho
- Cửa hàng / điểm bán
- Bộ phận

### 6.3. Sản phẩm
- SKU
- Nhóm sản phẩm
- Dòng sản phẩm
- Quy cách đóng gói
- UOM

### 6.4. Batch / chất lượng
- Batch No
- MFG Date
- Expiry Date
- QC Status
- Release Status
- Reason Code

### 6.5. Giao dịch mua hàng
- Supplier
- PR Status
- PO Status
- Receipt Status
- Lead Time Bucket

### 6.6. Giao dịch sản xuất
- Work Order No
- Work Order Status
- Production Line / Team / Shift
- Standard BOM Version

### 6.7. Giao dịch bán hàng
- Channel
- Customer
- Salesperson / Team
- Sales Order Status
- Return Reason

---

## 7. Đặc tả chi tiết từng báo cáo

## 7.1. Nhóm điều hành cấp CEO / COO

### EXE-01. Dashboard Điều Hành CEO 360

**Mục đích**  
Cung cấp màn hình “phòng lái” cho CEO/COO nhìn ngay tình hình doanh nghiệp mỗi ngày.

**Người dùng**  
CEO, COO, Finance Lead, Head of Operations

**Tần suất refresh**  
- Card tổng quan: mỗi 15–60 phút  
- Snapshot cuối ngày: khóa số lúc 23:59 hoặc theo quy định vận hành  

**Widget/KPI bắt buộc**
- Doanh thu thuần hôm nay / MTD
- Đơn hàng mới, đơn chờ xử lý, đơn chậm giao
- Available stock value
- Near-expiry value
- Batch on hold count
- Open PO overdue count
- Material shortage risk count
- Work orders delayed count
- Fill rate
- Return rate
- Yield bình quân trong ngày
- Top 5 SKU theo doanh thu
- Top 5 ngoại lệ cần xử lý ngay

**Drill-down bắt buộc**
- Từ doanh thu → SAL-02
- Từ tồn kho → WHS-01/WHS-02
- Từ batch hold → QUA-02
- Từ thiếu nguyên liệu → PUR-03
- Từ lệnh sản xuất trễ → PRO-03

**Ngưỡng cảnh báo gợi ý**
- Fill rate < 95%
- Near-expiry ratio > 8% tồn kho theo giá trị
- Batch on hold > 24 giờ chưa xử lý
- Open PO overdue > 15% số PO đang mở
- Material variance > +5% so với định mức ở lệnh bất kỳ

---

### EXE-02. Doanh thu & Lợi nhuận Gần Đúng theo Kênh / SKU

**Mục đích**  
Giúp ban điều hành biết **bán được gì, ở đâu, và gần đúng là có lời hay không** trong Phase 1.

**Người dùng**  
CEO, Sales Lead, Finance Lead, COO

**Tần suất refresh**  
Hàng giờ / cuối ngày

**Phạm vi số liệu**
- Doanh thu gộp
- Chiết khấu trực tiếp
- Hàng trả lại
- Doanh thu thuần
- Giá vốn gần đúng theo standard cost hoặc batch cost nếu đã khóa
- Lợi nhuận gộp gần đúng

**Phân tích theo chiều**
- Channel
- SKU
- Brand
- Customer / distributor
- Warehouse phục vụ
- Date

**Ghi chú quan trọng**
Trong Phase 1, chỉ số lợi nhuận gộp là **gross margin approximate**, chưa bao gồm mọi phân bổ gián tiếp như marketing, payroll, overhead.

**Ngưỡng cảnh báo gợi ý**
- SKU có gross margin gần đúng < ngưỡng tối thiểu công ty
- Channel tăng doanh thu nhưng margin giảm liên tiếp 2 tuần
- Tỷ lệ return của SKU vượt baseline

---

### EXE-03. Sức Khỏe Tồn Kho, Batch và Cận Date

**Mục đích**  
Giúp điều hành biết tiền đang nằm trong kho ở đâu và rủi ro batch/hạn dùng thế nào.

**Người dùng**  
CEO, COO, Warehouse Lead, QA Lead, Finance Lead

**Tần suất refresh**  
Theo giờ / cuối ngày

**Chỉ số chính**
- Tồn vật lý theo giá trị
- Tồn khả dụng theo giá trị
- Tồn hold / quarantine theo giá trị
- Near-expiry value
- Aged stock buckets: 0–30 / 31–60 / 61–90 / >90 ngày hoặc theo chuẩn công ty
- Top SKU chôn vốn nhiều nhất
- Top batch cận date
- Sample/tester/gift stock ratio

**Drill-down**
- Warehouse → Zone/Bin → SKU → Batch
- SKU → Batch → QC status → chứng từ liên quan

---

## 7.2. Nhóm mua hàng / nhà cung cấp

### PUR-01. Purchase Request / PO Pipeline

**Mục đích**  
Quản lý toàn bộ pipeline từ nhu cầu mua đến PO và nhận hàng.

**Người dùng**  
Purchasing, COO, Planning

**Tần suất refresh**  
Theo giờ

**Chỉ số chính**
- Số PR theo trạng thái
- Số PO theo trạng thái
- Giá trị PO mở
- PO quá hạn giao
- PR approval lead time
- PO cycle time
- Partial receipt count
- PO chưa có xác nhận giao từ NCC

**Bộ lọc**
- Supplier
- Nguyên liệu / SKU
- Warehouse
- Buyer
- PR/PO status
- Due date bucket

**Drill-down**
- PR → PO → Receipt → QC → put-away

---

### PUR-02. Supplier OTIF, Chất Lượng và Biến Động Giá

**Mục đích**  
Đánh giá NCC dựa trên đúng thứ quan trọng: giao đúng, giao đủ, hàng đạt chất lượng và giá có biến động không.

**Người dùng**  
Purchasing Manager, COO, Finance Lead

**Tần suất refresh**  
Tuần / tháng

**Chỉ số chính**
- OTIF (On-Time In-Full)
- Supplier defect rate
- Lead time actual vs expected
- Price variance vs last purchase
- Số lô fail QC
- Số đơn trễ > X ngày
- Điểm tổng hợp NCC

**Ngưỡng gợi ý**
- OTIF < 90%
- Defect rate > 3%
- Price variance > 5% mà không có phê duyệt

**Ghi chú**
Điểm NCC nên dùng để hỗ trợ quyết định, không thay thế judgment của purchasing trong các mặt hàng chiến lược.

---

### PUR-03. Rủi Ro Thiếu Nguyên Liệu / Replenishment Risk

**Mục đích**  
Cảnh báo sớm những nguyên liệu có nguy cơ làm tắc kế hoạch sản xuất.

**Người dùng**  
Purchasing, Planning, COO, Production

**Tần suất refresh**  
Ngày / theo giờ khi có lệnh mới

**Nguồn dữ liệu**
- Available stock NVL
- Reserved for production
- Open PO
- Lead time supplier
- Kế hoạch sản xuất đã duyệt

**Chỉ số chính**
- Số nguyên liệu dưới safety level
- Số nguyên liệu projected stockout trong lead time
- Risk score theo nguyên liệu
- PO expedite list
- Impacted work orders count

**Drill-down**
- Nguyên liệu → supplier → open PO → impacted work orders

---

## 7.3. Nhóm QA / QC

### QUA-01. Tổng Quan QC Đầu Vào

**Mục đích**  
Theo dõi chất lượng nguyên liệu đầu vào và ảnh hưởng đến việc nhập kho khả dụng.

**Người dùng**  
QA/QC, Purchasing, Warehouse

**Tần suất refresh**  
Ngày

**Chỉ số chính**
- Tổng lô/phiếu QC đầu vào
- Pass / Hold / Fail count
- Inbound QC pass rate
- Số lô chờ kiểm > SLA
- Top supplier có lỗi cao
- Top nguyên liệu fail nhiều

**Drill-down**
- Supplier → item → batch → inspection record

---

### QUA-02. Batch Quality & Release Aging

**Mục đích**  
Quản lý thời gian từ lúc hoàn tất sản xuất/nhận hàng đến lúc batch được release hoặc fail.

**Người dùng**  
QA/QC, Production, COO, Warehouse

**Tần suất refresh**  
Theo giờ

**Chỉ số chính**
- Batch đang chờ release
- Batch hold count
- Batch fail count
- Release lead time
- Batch aging buckets
- Batch blocked value

**Ngưỡng cảnh báo gợi ý**
- Batch hold > 24h với thành phẩm thương mại
- Batch waiting release > SLA công ty
- Fail rate tăng vượt baseline 2 tuần liên tiếp

**Drill-down**
- Batch → work order / receipt → QC checkpoints → disposition history

---

### QUA-03. Non-conformance / CAPA / Complaint Linkage

**Mục đích**  
Kết nối sự cố chất lượng với hành động khắc phục và các complaint bên ngoài.

**Người dùng**  
QA/QC Manager, COO, R&D (tham khảo), CSKH (tham khảo)

**Tần suất refresh**  
Tuần / khi có sự cố

**Chỉ số chính**
- Số NCR theo mức độ
- CAPA open / overdue
- Complaint liên quan batch
- Root cause category
- Thời gian đóng CAPA

**Ghi chú**
Phase 1 nên hỗ trợ ít nhất mức basic:
- tạo NCR
- gắn batch
- gắn reason code
- theo dõi due date xử lý

---

## 7.4. Nhóm sản xuất

### PRO-01. Kế Hoạch vs Thực Tế Sản Xuất

**Mục đích**  
Theo dõi mức độ bám kế hoạch của sản xuất theo ngày/ca/lệnh.

**Người dùng**  
Production Manager, COO, Planning

**Tần suất refresh**  
Theo ca / theo giờ

**Chỉ số chính**
- Planned qty
- Actual good qty
- Production plan adherence
- Work orders on schedule / delayed
- Output theo line / shift / team
- Thành phẩm nhập kho trong ngày

**Bộ lọc**
- Date
- Shift
- Line
- Work order
- SKU
- Brand

**Ngưỡng cảnh báo gợi ý**
- Plan adherence < 95%
- Work order delay > 1 ca
- Sản lượng line thấp hơn baseline

---

### PRO-02. Hao Hụt, Yield và Variance NVL

**Mục đích**  
Đây là một trong những báo cáo đáng tiền nhất vì nó bóc ra chỗ đốt tiền âm thầm ở nhà máy.

**Người dùng**  
Production, QA, COO, Finance Lead

**Tần suất refresh**  
Theo ca / cuối ngày

**Chỉ số chính**
- Standard usage
- Actual usage
- Material variance %
- Yield rate
- Scrap qty / scrap rate
- Rework qty
- Top work orders vượt định mức

**Drill-down**
- Work order → component → standard vs actual → operator/team/shift nếu có

**Ngưỡng cảnh báo gợi ý**
- Material variance > +5%
- Yield < 95% standard
- Scrap rate vượt baseline của SKU

---

### PRO-03. Work Order Exception Board

**Mục đích**  
Màn hình điều hành ngoại lệ cho lệnh sản xuất.

**Người dùng**  
Production, Planning, COO

**Tần suất refresh**  
Theo giờ

**Các nhóm ngoại lệ**
- Thiếu NVL
- Chờ QC release
- Trễ kế hoạch
- Vượt variance
- Block do batch issue
- Rework required

**Mục tiêu**
Không để Production Manager phải mở 6 màn hình mới biết lệnh nào đang cháy.

---

## 7.5. Nhóm kho

### WHS-01. Stock Position & Availability

**Mục đích**  
Là báo cáo vận hành quan trọng nhất của kho và sale admin.

**Người dùng**  
Warehouse, Sales Admin, COO, Purchasing, Production

**Tần suất refresh**  
Gần real-time

**Chỉ số chính**
- Physical stock
- Available stock
- Reserved stock
- Hold / quarantine stock
- Inbound expected
- Outbound allocated
- Stock by warehouse / bin / batch

**Drill-down**
- Warehouse → Zone/Bin → SKU → Batch → ledger movements

**Yêu cầu bắt buộc**
Phải tách rõ:
- tồn vật lý
- tồn khả dụng
- tồn đã giữ
- tồn hold  
Nếu không tách, mọi báo cáo bán hàng sẽ sai.

---

### WHS-02. Inventory Aging / Near-Expiry / Hold Stock

**Mục đích**  
Giúp quản lý kho và ban điều hành nhìn rõ rủi ro vốn chết và rủi ro hết hạn.

**Người dùng**  
Warehouse, QA, COO, CEO

**Tần suất refresh**  
Ngày

**Chỉ số chính**
- Aging buckets theo số ngày nằm kho
- Near-expiry by batch
- Hold stock by reason
- Slow-moving items
- Batch không có movement > X ngày
- Giá trị tồn theo bucket

**Ngưỡng cảnh báo gợi ý**
- Near-expiry ratio > 8% theo giá trị
- Hold stock > 3% tổng tồn
- Slow-moving stock tăng liên tiếp 4 tuần

---

### WHS-03. Stock Accuracy / Cycle Count Variance

**Mục đích**  
Đo độ tin cậy của dữ liệu kho.

**Người dùng**  
Warehouse Lead, Finance Lead, COO

**Tần suất refresh**  
Tuần / theo đợt kiểm kê

**Chỉ số chính**
- Stock accuracy %
- Số line sai lệch
- Giá trị chênh lệch kiểm kê
- Top khu vực/bin sai lệch cao
- Adjustment trend

**Ngưỡng cảnh báo gợi ý**
- Stock accuracy < 98% (line-based) hoặc < ngưỡng công ty
- Giá trị điều chỉnh vượt ngưỡng cảnh báo tuần

---

## 7.6. Nhóm bán hàng

### SAL-01. Sales Order Pipeline & Fill Rate

**Mục đích**  
Theo dõi pipeline xử lý đơn từ lúc tạo đến lúc giao và đánh giá khả năng “fill” đơn hàng.

**Người dùng**  
Sales, Sales Admin, Warehouse, COO

**Tần suất refresh**  
Theo giờ

**Chỉ số chính**
- Order count theo trạng thái
- Ordered qty vs shipped qty
- Fill rate
- Backorder count
- On-time shipment rate
- Đơn chờ do thiếu tồn
- Đơn bị block do QC/batch issue

**Drill-down**
- Sales order → allocation → shipment → invoice/cash record nếu có

**Ngưỡng cảnh báo gợi ý**
- Fill rate < 95%
- Backorder > baseline
- On-time shipment rate < 95%

---

### SAL-02. Sales by Channel / SKU / Customer

**Mục đích**  
Cho biết bán ở đâu, bán cái gì, bán cho ai.

**Người dùng**  
Sales Lead, CEO, Finance Lead

**Tần suất refresh**  
Ngày / tuần

**Chỉ số chính**
- Gross sales
- Discount
- Returns
- Net sales
- Average selling price
- Gross margin approximate
- Top channel / customer / SKU

**Chiều phân tích**
- Channel
- Customer type
- SKU
- Brand
- Sales team
- Region (nếu có)

---

### SAL-03. Returns / Refund / Complaint Summary

**Mục đích**  
Theo dõi chất lượng bán hàng sau giao và phát hiện vấn đề sản phẩm/kho/giao vận.

**Người dùng**  
Sales, CSKH, QA, COO, CEO

**Tần suất refresh**  
Ngày / tuần

**Chỉ số chính**
- Return count
- Return rate
- Return value
- Complaint count
- Top return reasons
- Batch-linked returns
- Top SKU có tỉ lệ return cao

**Lưu ý**
Với mỹ phẩm, report này cực quan trọng vì có thể lộ ra:
- lỗi batch
- lỗi đóng gói
- lỗi giao sai
- kỳ vọng khách hàng không khớp claim

---

## 7.7. Nhóm xuyên bộ phận

### XFN-01. Batch Traceability Cockpit

**Mục đích**  
Cho phép truy từ batch thành phẩm ngược về:
- work order
- nguyên liệu/batch đầu vào
- supplier
- QC records
- shipment / sales orders đã xuất ra ngoài

**Người dùng**  
QA, Warehouse, Production, COO

**Tần suất refresh**  
Near real-time / on demand

**Use case chính**
- Điều tra complaint
- Điều tra batch fail
- Khoanh vùng thu hồi nội bộ
- Xác định khách hàng nào đã nhận batch liên quan

---

### XFN-02. Working Capital Snapshot cơ bản

**Mục đích**  
Cho lãnh đạo nhìn nhanh góc độ vốn vận hành trong Phase 1.

**Người dùng**  
CEO, Finance Lead, COO

**Tần suất refresh**  
Ngày / tuần

**Chỉ số chính**
- Giá trị tồn kho
- Giá trị PO mở
- Giá trị hàng đang trên đường / chờ QC
- Doanh thu thuần MTD
- Công nợ phải thu/phải trả cơ bản nếu được triển khai trong Phase 1
- Inventory days gần đúng (nếu đủ dữ liệu)

**Ghi chú**
Nếu finance full chưa go-live ở Phase 1 thì report này có thể dùng “operational finance view”, không thay thế sổ sách kế toán pháp lý.

---

### XFN-03. Master Data Completeness & Data Quality

**Mục đích**  
Đo độ sạch dữ liệu gốc, vì dữ liệu bẩn sẽ giết mọi báo cáo còn lại.

**Người dùng**  
PM, Data Owner, COO, System Admin

**Tần suất refresh**  
Tuần

**Chỉ số chính**
- % SKU active có đủ mandatory fields
- % items có UOM mapping đúng
- % batch records có đủ expiry/MFG/QC status
- Duplicate suppliers/customers/items suspected
- Records bị thiếu cost / thiếu category / thiếu brand

**Ngưỡng gợi ý**
- Master data completeness < 98% → phải xử lý
- Duplicate suspected records > baseline → data steward review

---

## 8. Từ điển KPI cốt lõi

### 8.1. KPI điều hành và bán hàng

| KPI Code | KPI | Công thức chuẩn / logic | Chủ sở hữu | Tần suất |
|---|---|---|---|---|
| KPI-01 | Gross Sales | Tổng giá trị bán trước chiết khấu và trả hàng | Sales Lead | Ngày |
| KPI-02 | Net Sales | Gross Sales - Discount - Return/Refund | Sales + Finance | Ngày |
| KPI-03 | Average Selling Price | Net Sales / Shipped Qty | Sales Lead | Ngày |
| KPI-04 | Gross Margin Approx | Net Sales - Estimated COGS | Finance Lead | Ngày/tuần |
| KPI-05 | Order Fill Rate | Shipped Qty / Ordered Qty | Sales Admin | Theo giờ/ngày |
| KPI-06 | On-Time Shipment Rate | Số đơn giao/xuất đúng hạn / tổng số đơn đến hạn | Sales Admin + Warehouse | Ngày |
| KPI-07 | Return Rate | Return Qty hoặc Return Value / Shipped Qty hoặc Net Sales | Sales + QA | Tuần |

### 8.2. KPI mua hàng / NCC

| KPI Code | KPI | Công thức chuẩn / logic | Chủ sở hữu | Tần suất |
|---|---|---|---|---|
| KPI-08 | PR Approval Lead Time | Avg(Submit PR → Approve PR) | Purchasing Manager | Tuần |
| KPI-09 | PO Cycle Time | Avg(PR Approved → PO Sent) hoặc theo rule công ty | Purchasing Manager | Tuần |
| KPI-10 | Supplier OTIF | Số PO dòng giao đúng hạn và đủ số lượng / tổng PO dòng đến hạn | Purchasing Manager | Tuần |
| KPI-11 | Supplier Defect Rate | Số lô/qty fail QC / tổng lô/qty nhận từ NCC | Purchasing + QA | Tuần |
| KPI-12 | Purchase Price Variance | (Current Price - Reference Price) / Reference Price | Purchasing + Finance | Tuần |

### 8.3. KPI chất lượng

| KPI Code | KPI | Công thức chuẩn / logic | Chủ sở hữu | Tần suất |
|---|---|---|---|---|
| KPI-13 | Inbound QC Pass Rate | Pass Lots / Inspected Lots | QA/QC Lead | Ngày |
| KPI-14 | Batch Fail Rate | Fail Batches / Total Reviewed Batches | QA/QC Lead | Ngày/tuần |
| KPI-15 | Batch Release Lead Time | Avg(Ready for QC → Released) | QA/QC Lead | Ngày |
| KPI-16 | CAPA Overdue Rate | CAPA quá hạn / CAPA mở | QA/QC Lead | Tuần |

### 8.4. KPI sản xuất

| KPI Code | KPI | Công thức chuẩn / logic | Chủ sở hữu | Tần suất |
|---|---|---|---|---|
| KPI-17 | Production Plan Adherence | Actual Good Qty / Planned Qty | Production Manager | Theo ca/ngày |
| KPI-18 | Yield Rate | Good Output Qty / Theoretical or Planned Output Qty | Production Manager | Theo ca/ngày |
| KPI-19 | Material Variance % | (Actual Usage - Standard Usage) / Standard Usage | Production + Finance | Theo ca/ngày |
| KPI-20 | Scrap Rate | Scrap Qty / Total Input hoặc Total Output theo rule công ty | Production Manager | Theo ca/ngày |
| KPI-21 | Rework Rate | Rework Qty / Total Produced Qty | Production + QA | Tuần |

### 8.5. KPI kho

| KPI Code | KPI | Công thức chuẩn / logic | Chủ sở hữu | Tần suất |
|---|---|---|---|---|
| KPI-22 | Available Stock Accuracy | So sánh available stock system vs count/validated state | Warehouse Lead | Tuần |
| KPI-23 | Inventory Accuracy | Số line hoặc qty kiểm kê khớp / tổng số line hoặc qty kiểm kê | Warehouse Lead | Tuần |
| KPI-24 | Near-Expiry Ratio | Near-Expiry Stock Value / Total On-Hand Stock Value | Warehouse + QA | Ngày |
| KPI-25 | Hold Stock Ratio | Hold/Quarantine Stock Value / Total On-Hand Stock Value | Warehouse + QA | Ngày |
| KPI-26 | Stockout Risk Count | Số SKU/NVL projected thiếu trong horizon định nghĩa | Warehouse + Purchasing + Planning | Ngày |

### 8.6. KPI dữ liệu và truy xuất

| KPI Code | KPI | Công thức chuẩn / logic | Chủ sở hữu | Tần suất |
|---|---|---|---|---|
| KPI-27 | Master Data Completeness | Records complete mandatory fields / Active records | Data Owner | Tuần |
| KPI-28 | Traceability Resolution Time | Avg(Time to trace batch from finished goods to sources/destinations) | QA + Warehouse | Khi test/sự cố |
| KPI-29 | Duplicate Master Data Rate | Suspected duplicates / active master records | Data Steward | Tuần |

---

## 9. Công thức và lưu ý chuẩn hóa

### 9.1. Doanh thu thuần

`Net Sales = Gross Sales - Discount - Return/Refund`

Ghi chú:
- Discount phải bao gồm discount thương mại trực tiếp trong đơn
- Return/Refund phải bám theo reason code
- Không cộng trừ thủ công ngoài hệ thống

### 9.2. Lợi nhuận gộp gần đúng Phase 1

`Gross Margin Approx = Net Sales - Estimated COGS`

Trong đó `Estimated COGS` ưu tiên theo thứ tự:
1. Batch cost đã khóa (nếu có)  
2. Standard cost hiện hành  
3. Moving/average cost theo rule của Phase 1  

Chỉ số này là **quản trị nội bộ**, chưa thay thế báo cáo tài chính pháp lý.

### 9.3. Tồn khả dụng

`Available Stock = Physical Stock - Reserved Stock - Hold Stock`

Lưu ý:
- Sample/tester/gift stock không đưa vào available stock thương mại
- Hàng fail QC không đưa vào available stock
- Hàng đã allocate cho đơn vẫn phải vào reserved

### 9.4. Near-expiry

Cần thống nhất horizon. Ví dụ:
- NVL cận date: còn dưới 90 ngày
- Thành phẩm cận date: còn dưới 120 hoặc 180 ngày

Con số cụ thể phải do business và QA chốt theo đặc thù từng nhóm hàng.

### 9.5. Material Variance

`Material Variance % = (Actual Usage - Standard Usage) / Standard Usage`

Ý nghĩa:
- Dương: dùng vượt định mức
- Âm: dùng thấp hơn định mức  
Cần cẩn thận vì âm quá lớn cũng có thể là nhập liệu sai hoặc issue chưa ghi nhận đủ.

### 9.6. Fill Rate

`Fill Rate = Shipped Qty / Ordered Qty`

Nếu doanh nghiệp muốn tính theo line hoặc theo đơn, phải chọn **một chuẩn chính** và có thể bổ sung chuẩn phụ.

---

## 10. Danh mục cảnh báo và notification gợi ý

| Alert Code | Điều kiện kích hoạt | Người nhận chính | Mức ưu tiên |
|---|---|---|---|
| ALT-01 | PO quá hạn giao > X ngày | Purchasing, COO | High |
| ALT-02 | Nguyên liệu projected stockout trong lead time | Purchasing, Planning, COO | Critical |
| ALT-03 | Batch hold quá SLA | QA/QC, COO, Warehouse | High |
| ALT-04 | Material variance vượt ngưỡng | Production, COO, Finance Lead | High |
| ALT-05 | Fill rate thấp hơn ngưỡng ngày | Sales Lead, Warehouse, COO | High |
| ALT-06 | Near-expiry stock tăng vượt ngưỡng | Warehouse, QA, CEO | High |
| ALT-07 | Stock accuracy dưới ngưỡng | Warehouse, Finance Lead | Medium |
| ALT-08 | Supplier defect rate vượt ngưỡng | Purchasing, QA/QC | Medium |
| ALT-09 | Return rate của SKU tăng bất thường | Sales, QA/QC, COO | High |
| ALT-10 | Master data completeness thấp hơn ngưỡng | Data Owner, PM, COO | Medium |

**Khuyến nghị:**  
Ngưỡng cảnh báo nên được xem lại sau 8–12 tuần vận hành để tránh:
- báo động giả quá nhiều
- hoặc ngược lại, cảnh báo quá trễ

---

## 11. Gợi ý hình thức hiển thị

### 11.1. Dành cho điều hành
- KPI cards
- xu hướng 7 ngày / 30 ngày
- bảng ngoại lệ top 10
- heatmap theo kho / batch
- waterfall đơn giản cho Net Sales → Gross Margin Approx

### 11.2. Dành cho tác nghiệp
- bảng line-item có filter mạnh
- cột trạng thái màu rõ
- aging buckets
- exception queue
- action links drill xuống chứng từ

### 11.3. Dành cho điều tra sự cố
- traceability tree
- timeline sự kiện theo batch
- linked records: receipt → QC → work order → FG batch → shipment

---

## 12. Ownership dữ liệu và refresh cadence

| Nhóm dữ liệu | Owner nghiệp vụ | Refresh đề xuất | Ghi chú |
|---|---|---|---|
| Master Data | Data Owner từng domain | Theo thay đổi / snapshot ngày | Cần quy trình duyệt |
| Purchase / Supplier | Purchasing | Gần real-time / 15 phút | PO overdue cần near real-time |
| QC | QA/QC | Gần real-time | Batch hold/release là dữ liệu nhạy |
| Production | Production | Theo ca / 15–60 phút | Tùy mức ghi nhận tại chuyền |
| Warehouse | Warehouse | Gần real-time | Stock availability phải cập nhật nhanh |
| Sales Order | Sales Admin / OMS | Gần real-time | Fill rate và backlog cần sát thời gian |
| Executive Snapshot | BI/Reporting Owner | Hàng giờ + snapshot cuối ngày | Tránh chênh số do refresh lệch |

---

## 13. Điều kiện chấp nhận khi UAT cho báo cáo

Một báo cáo được coi là đạt khi thỏa tối thiểu các điều kiện sau:

1. Tổng số trên báo cáo khớp với dữ liệu giao dịch mẫu đã xác nhận.  
2. Công thức KPI khớp với định nghĩa trong tài liệu này.  
3. Bộ lọc theo thời gian, kho, SKU, batch, trạng thái trả kết quả đúng.  
4. Drill-down mở được danh sách chứng từ nguồn tương ứng.  
5. Người dùng không thấy dữ liệu vượt quyền đã được quy định trong Permission Matrix.  
6. Batch hold/fail không được tính vào available stock.  
7. Return/refund làm giảm Net Sales đúng theo logic hệ thống.  
8. Material variance lấy đúng BOM version áp dụng cho work order.  
9. Báo cáo không double-count giữa partial receipt / partial shipment.  
10. Snapshot cuối ngày không thay đổi trừ khi có điều chỉnh được phê duyệt và có audit log.

---

## 14. Ưu tiên triển khai theo đợt

### 14.1. Bắt buộc có ở go-live hoặc ngay sau go-live (P0)
- EXE-01 Dashboard Điều Hành CEO 360
- EXE-02 Doanh thu & Lợi nhuận Gần Đúng
- EXE-03 Sức Khỏe Tồn Kho / Batch / Cận Date
- PUR-01 Purchase Request / PO Pipeline
- PUR-03 Replenishment Risk
- QUA-01 QC Đầu Vào
- QUA-02 Batch Quality & Release Aging
- PRO-01 Kế Hoạch vs Thực Tế Sản Xuất
- PRO-02 Hao Hụt / Yield / Variance
- WHS-01 Stock Position & Availability
- WHS-02 Inventory Aging / Near-Expiry
- SAL-01 Sales Order Pipeline & Fill Rate
- SAL-02 Sales by Channel / SKU / Customer

### 14.2. Nên có trong 30–60 ngày sau go-live (P1)
- PUR-02 Supplier Scorecard
- QUA-03 NCR / CAPA / Complaint Linkage
- PRO-03 Work Order Exception Board
- WHS-03 Stock Accuracy / Cycle Count
- SAL-03 Returns / Complaint Summary
- XFN-01 Batch Traceability Cockpit
- XFN-02 Working Capital Snapshot
- XFN-03 Master Data Completeness

### 14.3. Để dành cho Phase 2
- Margin phân bổ đầy đủ
- Customer repeat behavior
- KOL profitability
- HR productivity dashboard
- Forecast demand nâng cao
- Automated anomaly detection

---

## 15. Rủi ro thường gặp khi làm báo cáo ERP và cách tránh

### 15.1. Tồn kho nhìn “đẹp” nhưng không dùng được
Nguyên nhân:
- không tách available / reserved / hold
- không bám batch/QC status

### 15.2. Doanh thu đúng nhưng margin sai
Nguyên nhân:
- cost chưa khóa
- return ghi nhận trễ
- discount không đi cùng đơn

### 15.3. Báo cáo mua hàng không giúp ngăn thiếu nguyên liệu
Nguyên nhân:
- không nối lead time và kế hoạch sản xuất
- chỉ xem PO mở mà không xem nhu cầu tiêu thụ

### 15.4. QA/QC có màn hình nhưng không chặn rủi ro
Nguyên nhân:
- batch hold vẫn bị tính khả dụng
- complaint không link batch

### 15.5. CEO có dashboard nhưng không ra quyết định được
Nguyên nhân:
- quá nhiều card đẹp nhưng không có exception list
- click không xuống được chứng từ gốc

---

## 16. Khuyến nghị thực chiến

1. Bắt đầu bằng **13 báo cáo P0**, không tham làm hết một lần.  
2. Chốt công thức KPI trước khi vẽ dashboard.  
3. Mọi số quan trọng đều phải drill xuống chứng từ.  
4. Tách snapshot cuối ngày khỏi real-time view để giảm tranh cãi.  
5. Trước go-live, chọn 20–30 case UAT có số liệu thật để test report.  
6. Sau 1 tháng chạy thật, review lại toàn bộ ngưỡng cảnh báo.  
7. Định kỳ hàng tuần phải có cuộc họp “report review” để làm sạch data và điều chỉnh định nghĩa khi cần.

---

## 17. Kết luận

Với doanh nghiệp mỹ phẩm, báo cáo Phase 1 không chỉ để “xem số”.  
Nó phải giúp trả lời 4 câu hỏi điều hành cốt lõi:

- **Hàng có đang khỏe không?**
- **Batch và chất lượng có đang an toàn không?**
- **Tiền có đang bị chôn ở kho hoặc đốt ở xưởng không?**
- **Bán được rồi thì có thật sự tạo ra lợi nhuận không?**

Nếu bộ Report & KPI Catalog này được dùng đúng cách, ERP sẽ không chỉ là hệ thống nhập liệu, mà trở thành **bảng điều khiển sống** của doanh nghiệp.

---

## 18. Phụ lục A – Danh sách câu hỏi quản trị mà báo cáo phải trả lời

### CEO / Chủ doanh nghiệp
- Hôm nay bán bao nhiêu?
- Kênh nào tăng mạnh nhưng margin xấu?
- Batch nào đang hold?
- Tồn kho nào đang chôn vốn?
- Nguyên liệu nào sắp thiếu làm ảnh hưởng sản xuất?

### COO
- Work order nào đang trễ?
- Kho nào đang nghẽn?
- NCC nào làm chậm kế hoạch?
- QC đang là nút thắt ở đâu?

### Purchasing
- Nguyên liệu nào phải đặt ngay?
- NCC nào giao trễ hoặc fail nhiều?
- Giá mua có đang bị đội bất thường không?

### QA/QC
- Lô nào chưa release?
- Batch nào có complaint?
- CAPA nào sắp quá hạn?

### Production
- Lệnh nào đang vượt định mức?
- Line nào đang hụt output?
- Yield đang xấu ở SKU nào?

### Warehouse
- Có đủ hàng để xuất không?
- Hàng nào sắp hết hạn?
- Bin nào hay lệch tồn?

### Sales
- Đơn nào chưa fill?
- SKU nào bán tốt nhất?
- Return đang tăng ở đâu?

---

## 19. Phụ lục B – Gợi ý tên file/dashboard trong hệ thống

- Dashboard CEO 360
- Sales & Margin Daily View
- Inventory Health Monitor
- Purchase Pipeline Board
- Supplier Scorecard
- Inbound QC Daily Monitor
- Batch Release Board
- Production Control Tower
- Material Variance Analyzer
- Warehouse Availability Monitor
- Sales Order Fulfillment Board
- Return & Complaint Pulse
- Traceability Cockpit
- Data Quality Monitor

