# PRD / SRS Phase 1
## Hệ thống ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm

**Mã tài liệu:** ERP-PRD-SRS-P1  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-23  
**Người sở hữu tài liệu:** Product Owner / ERP Solution Architect  
**Mục đích:** Chốt phạm vi, quy trình, màn hình, yêu cầu chức năng, quy tắc nghiệp vụ và tiêu chí nghiệm thu cho Phase 1 của hệ thống ERP web.

---

## 1. Mục tiêu của tài liệu

Tài liệu này dùng để:

1. Biến bản Blueprint thành phạm vi build cụ thể cho đội BA, UI/UX, dev, tester và đội nghiệp vụ nội bộ.
2. Chốt **Phase 1** theo nguyên tắc “khóa hàng trước, khóa tiền sau, tối ưu sâu ở phase sau”.
3. Định nghĩa rõ:
   - các module sẽ làm ngay,
   - các luồng nghiệp vụ bắt buộc,
   - các trạng thái chứng từ,
   - rule phê duyệt,
   - dữ liệu bắt buộc,
   - báo cáo tối thiểu,
   - tiêu chí nghiệm thu.

---

## 2. Mục tiêu kinh doanh của Phase 1

Phase 1 không nhằm “làm cho hệ thống đủ mọi thứ”. Mục tiêu là giải quyết 6 bài toán sống còn của công ty mỹ phẩm:

1. Biết chính xác đang có gì trong kho, ở kho nào, batch nào, hạn đến đâu.
2. Kiểm soát từ lúc mua nguyên vật liệu đến lúc nhập kho, QC và sẵn sàng sản xuất.
3. Kiểm soát lệnh sản xuất, cấp phát nguyên liệu, hao hụt và nhập thành phẩm theo batch.
4. Không cho hàng chưa QC pass đi vào bán hàng.
5. Kiểm soát đơn bán, giữ tồn, xuất kho, trả hàng và nhìn được tình trạng giao dịch.
6. Tạo một nền dữ liệu chuẩn để phase sau mở sang CRM, HRM, KOL, POS, tích hợp website/sàn và kế toán sâu.

---

## 3. Phạm vi Phase 1

### 3.1. In scope

Phase 1 bao gồm 6 module lõi:

1. **Dữ liệu gốc (Master Data)**
2. **Mua hàng (Procurement)**
3. **QA/QC**
4. **Sản xuất (Production)**
5. **Kho hàng (Warehouse / WMS cơ bản)**
6. **Bán hàng (Sales / OMS cơ bản)**

Ngoài 6 module trên, Phase 1 cũng bao gồm các thành phần nền bắt buộc để hệ thống chạy được:

- Đăng nhập và phân quyền
- Workflow phê duyệt cơ bản
- Nhật ký thay đổi (audit log)
- Đánh số chứng từ
- Đính kèm file
- Import / export Excel cơ bản
- Dashboard và báo cáo tối thiểu cho CEO / vận hành / kho / mua hàng / QC / sales

### 3.2. Out of scope

Những phần dưới đây **không build sâu trong Phase 1** và sẽ đi vào phase sau:

- HRM đầy đủ: chấm công, payroll, KPI, tuyển dụng
- CRM nâng cao: loyalty, ticket CSKH, automation chăm lại
- KOL / KOC / affiliate management
- POS cửa hàng hoàn chỉnh
- Tích hợp thời gian thực với website, sàn, đơn vị vận chuyển
- Kế toán tài chính đầy đủ theo chuẩn hạch toán / báo cáo thuế
- MRP tự động nâng cao và forecasting AI
- Mobile app riêng
- Supplier portal, dealer portal, KOL portal
- CAPA nâng cao, complaint management chi tiết, compliance workflow sâu
- Đa pháp nhân (multi-company legal entity) phức tạp

### 3.3. Ranh giới triển khai đề xuất

- **1 pháp nhân chính**
- **nhiều kho / chi nhánh / địa điểm**
- **nhiều loại hàng**: nguyên liệu, bao bì, bán thành phẩm, thành phẩm, sample/tester
- **quản lý batch/lô và hạn dùng**
- **bán hàng nhiều kênh theo nhập tay / import**, chưa đồng bộ tự động thời gian thực

---

## 4. Mục tiêu thành công của Phase 1

Phase 1 được xem là thành công khi đạt được các điều kiện sau:

1. Tồn kho theo batch và kho khớp thực tế ở mức tin cậy vận hành.
2. Hàng inbound phải qua QC trước khi trở thành tồn khả dụng.
3. Lệnh sản xuất sử dụng nguyên liệu theo BOM và ghi được actual consumption.
4. Thành phẩm có batch, ngày sản xuất, hạn dùng và trạng thái QC.
5. Đơn bán không được xác nhận vượt tồn khả dụng trừ khi có quyền đặc biệt.
6. Có thể truy vết từ đơn bán → batch thành phẩm → lệnh sản xuất → batch nguyên liệu chính.
7. Có báo cáo đủ để CEO và COO thấy: tồn kho, hàng cận date, batch hold, đơn mua mở, đơn bán mở, sản lượng, hao hụt, doanh thu cơ bản.

---

## 5. Vai trò người dùng trong Phase 1

### 5.1. Nhóm vai trò chính

1. **System Admin**
   - Quản trị user, role, danh mục cấu hình.
2. **Master Data Admin**
   - Quản lý item, BOM, supplier, customer, warehouse, price list.
3. **Purchasing Officer**
   - Tạo đề nghị mua, PO, theo dõi nhận hàng.
4. **Purchasing Manager**
   - Duyệt mua hàng theo ngưỡng.
5. **QC Officer**
   - Tạo phiếu kiểm, nhập kết quả QC.
6. **QA Manager**
   - Duyệt pass / fail / hold với lô kiểm tra.
7. **Production Planner**
   - Lập kế hoạch sản xuất thủ công, tạo lệnh sản xuất.
8. **Production Supervisor**
   - Triển khai lệnh, ghi sản lượng, ghi hao hụt, xác nhận hoàn thành.
9. **Warehouse Staff**
   - Nhập, xuất, chuyển kho, kiểm kê.
10. **Warehouse Manager**
    - Duyệt điều chỉnh kho, duyệt chuyển kho theo rule.
11. **Sales Admin**
    - Tạo báo giá / đơn bán, theo dõi giao hàng.
12. **Sales Manager**
    - Duyệt discount / duyệt đơn vượt ngưỡng / duyệt khách vượt công nợ.
13. **Finance Viewer / Approver**
    - Xem công nợ tham chiếu, duyệt một số giao dịch tài chính liên quan đến mua bán.
14. **CEO / COO / Head**
    - Xem dashboard, báo cáo, phê duyệt giao dịch vượt ngưỡng.

### 5.2. Nguyên tắc phân quyền

- Người tạo không mặc định có quyền duyệt.
- Giao dịch đã được duyệt và phát sinh hệ quả kho không được hard delete.
- Mọi sửa dữ liệu quan trọng phải có audit log.
- Vai trò xem báo cáo không được tự ý sửa dữ liệu vận hành nếu không có quyền.

---

## 6. Nguyên tắc thiết kế nghiệp vụ

1. **Một nguồn sự thật duy nhất** cho item, batch, kho, supplier, customer, BOM.
2. **Không hard delete chứng từ nghiệp vụ**. Chỉ hủy hoặc reverse theo rule.
3. **Batch/lô và hạn dùng là bắt buộc** đối với nguyên liệu và thành phẩm cần quản lý.
4. **QC status quyết định tồn khả dụng**.
5. **Tồn khả dụng khác tồn vật lý**.
6. **BOM chỉ có 1 version active tại 1 thời điểm** cho cùng một thành phẩm trong Phase 1.
7. **Mọi giao dịch kho phải đi vào stock ledger**.
8. **Mọi batch thành phẩm phải truy được về lệnh sản xuất và nguyên liệu cấp phát chính**.
9. **Approval theo ngưỡng giá trị/rủi ro**, không phải mọi giao dịch đều CEO duyệt.
10. **Phase 1 ưu tiên vận hành thực dụng**, chưa ôm logic tài chính quá sâu.

---

## 7. Thuật ngữ chính

- **Item**: Mã hàng trong hệ thống.
- **Raw Material**: Nguyên vật liệu.
- **Packaging**: Bao bì, tem, nhãn, phụ kiện.
- **Semi-finished Goods**: Bán thành phẩm.
- **Finished Goods**: Thành phẩm.
- **Batch/Lot**: Mã lô.
- **Expiry Date**: Hạn sử dụng.
- **BOM**: Bill of Materials, định mức nguyên liệu/cấu phần.
- **QC Hold**: Trạng thái tạm giữ, chưa cho dùng / chưa cho bán.
- **Available Stock**: Tồn khả dụng.
- **Reserved Stock**: Tồn đã giữ cho đơn.
- **Quarantine Stock**: Tồn cách ly / chờ xử lý.
- **GRN**: Goods Receipt Note, phiếu nhận hàng/nhập hàng.
- **PR**: Purchase Requisition, đề nghị mua.
- **PO**: Purchase Order, đơn mua hàng.
- **WO / MO**: Work Order / Manufacturing Order, lệnh sản xuất.
- **SO**: Sales Order, đơn bán hàng.
- **DO**: Delivery Order, lệnh giao / xuất giao.
- **NCR**: Non-conformance Report, biên bản không phù hợp.

---

## 8. Các giả định và phụ thuộc

1. Danh mục item, supplier, customer sẽ được chuẩn hóa trước khi migrate.
2. Công ty có quy ước mã hàng, mã batch, mã chứng từ hoặc đồng ý dùng quy ước hệ thống đề xuất.
3. Bảng giá cơ bản được chốt trước khi go-live.
4. BOM và quy cách đóng gói được cung cấp bởi R&D / sản xuất.
5. Kho vận có thể nhập batch và hạn dùng khi thao tác.
6. Đội nghiệp vụ đồng ý dùng hệ thống làm nguồn dữ liệu chính thay cho Excel sau go-live.
7. Các tích hợp ngoài (sàn, website, vận chuyển) sẽ tạm xử lý bằng import / export trong Phase 1 nếu cần.

---

## 9. Kiến trúc chức năng tổng quan Phase 1

```text
Master Data
   ├─ Item / BOM / Warehouse / Supplier / Customer / Price List
   ├─ User / Role / Approval
   └─ Batch Rules / Numbering Rules

Procurement
   ├─ PR
   ├─ PO
   ├─ Receiving
   └─ Supplier Return

QA/QC
   ├─ Inbound QC
   ├─ Production / FG QC
   ├─ Hold / Pass / Fail
   └─ Release Control

Production
   ├─ Production Order
   ├─ Material Issue
   ├─ Output / Scrap / Variance
   └─ FG Batch Creation

Warehouse
   ├─ Receipt
   ├─ Issue
   ├─ Transfer
   ├─ Stock Count
   └─ Stock Ledger / Batch Trace

Sales
   ├─ Quotation (basic)
   ├─ Sales Order
   ├─ Stock Reservation
   ├─ Delivery Request
   └─ Sales Return
```

---

## 10. Danh sách màn hình cấp cao (Screen Inventory)

### 10.1. Common / hệ thống
- COM-01: Login
- COM-02: Dashboard phê duyệt
- COM-03: Audit log viewer
- COM-04: Attachment center
- COM-05: Notification center
- COM-06: User / Role / Permission
- COM-07: Approval matrix configuration
- COM-08: Document numbering configuration

### 10.2. Master Data
- MD-01: Danh mục item
- MD-02: Chi tiết item
- MD-03: Import item
- MD-04: BOM list
- MD-05: BOM detail / version
- MD-06: Supplier list / detail
- MD-07: Customer list / detail
- MD-08: Warehouse list / detail
- MD-09: UOM & conversion
- MD-10: Price list & discount rule
- MD-11: Item category / brand / line
- MD-12: Batch rule configuration

### 10.3. Procurement
- PUR-01: PR list / create / detail
- PUR-02: PO list / create / detail
- PUR-03: Receiving against PO
- PUR-04: Supplier return
- PUR-05: Purchase dashboard

### 10.4. QA/QC
- QC-01: Inbound inspection list / detail
- QC-02: FG inspection list / detail
- QC-03: QC decision / release
- QC-04: NCR list / detail
- QC-05: QC dashboard

### 10.5. Production
- PROD-01: Production order list
- PROD-02: Production order detail
- PROD-03: Material issue
- PROD-04: Production confirmation
- PROD-05: Output & scrap entry
- PROD-06: Production dashboard

### 10.6. Warehouse
- WH-01: Goods receipt
- WH-02: Goods issue
- WH-03: Transfer order
- WH-04: Transfer receipt
- WH-05: Stock count
- WH-06: Stock adjustment
- WH-07: Stock ledger
- WH-08: Stock by batch / expiry
- WH-09: Batch trace

### 10.7. Sales
- SAL-01: Quotation list / detail
- SAL-02: Sales order list / detail
- SAL-03: SO approval
- SAL-04: Delivery request / delivery order
- SAL-05: Sales return
- SAL-06: Sales dashboard

---

## 11. Functional Requirements chung (Common Requirements)

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-COM-01 | Hệ thống hỗ trợ đăng nhập bằng username/password, role-based access. | Must |
| FR-COM-02 | Mọi chứng từ phải có số chứng từ tự động, duy nhất theo loại chứng từ. | Must |
| FR-COM-03 | Hệ thống phải lưu audit log cho thao tác create/edit/approve/cancel trên chứng từ quan trọng. | Must |
| FR-COM-04 | Hỗ trợ attachment cho item, supplier, customer, PO, QC, production order, SO. | Must |
| FR-COM-05 | Hỗ trợ workflow phê duyệt theo vai trò và ngưỡng cấu hình. | Must |
| FR-COM-06 | Hỗ trợ export Excel ở các màn hình danh sách chính. | Must |
| FR-COM-07 | Hỗ trợ import Excel cho master data và một số giao dịch đầu kỳ theo cấu hình. | Should |
| FR-COM-08 | Hệ thống phải hỗ trợ tiếng Việt giao diện; field code dùng tiếng Anh chuẩn hóa nội bộ. | Must |
| FR-COM-09 | Hệ thống sử dụng múi giờ Asia/Ho_Chi_Minh cho mọi timestamp nghiệp vụ. | Must |
| FR-COM-10 | Không cho hard delete chứng từ đã được approve hoặc đã phát sinh giao dịch kho. | Must |

---

## 12. Module 1 — Dữ liệu gốc (Master Data)

### 12.1. Mục tiêu nghiệp vụ

Module này tạo ra “ngôn ngữ chung” cho toàn bộ ERP. Nếu module này sai, tất cả các module còn lại sẽ sai theo.

### 12.2. Đối tượng dữ liệu chính

1. Item master
2. UOM / conversion
3. Item category / type / brand / line
4. BOM
5. Supplier
6. Customer
7. Warehouse / location
8. Price list
9. Batch rule
10. Role & approval matrix

### 12.3. Phân loại item trong Phase 1

Item phải hỗ trợ ít nhất các loại sau:

- Raw Material
- Packaging
- Semi-finished Goods
- Finished Goods
- Sample / Tester
- Consumable (vật tư phụ)
- Service (nếu cần cho giao dịch mua đơn giản)

### 12.4. Trường dữ liệu bắt buộc cho Item

| Nhóm | Trường |
|---|---|
| Nhận diện | item_code, item_name, item_type, category |
| Đơn vị | base_uom, purchase_uom, sales_uom, conversion_rate |
| Kiểm soát | batch_required, expiry_required, qc_required |
| Kho | default_warehouse (nếu có), storage_condition |
| Kinh doanh | brand, product_line, status |
| Giá trị | standard_cost (tham chiếu), tax_group (tham chiếu) |
| Hạn dùng | shelf_life_days hoặc shelf_life_months |
| Truy xuất | barcode / internal code (nếu có) |

### 12.5. Functional Requirements

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-MD-01 | Tạo/sửa/khóa item master với loại hàng, đơn vị tính và thuộc tính kiểm soát. | Must |
| FR-MD-02 | Hỗ trợ import item từ Excel theo template chuẩn. | Must |
| FR-MD-03 | Mỗi item_code là duy nhất trong hệ thống. | Must |
| FR-MD-04 | Hỗ trợ khai báo batch_required, expiry_required, qc_required theo item. | Must |
| FR-MD-05 | Hỗ trợ quản lý BOM cho thành phẩm và bán thành phẩm. | Must |
| FR-MD-06 | Mỗi BOM có version, effective date và chỉ 1 version active tại 1 thời điểm. | Must |
| FR-MD-07 | Hỗ trợ supplier master với payment term, lead time, approved status. | Must |
| FR-MD-08 | Hỗ trợ customer master với channel, price list, credit limit cơ bản. | Must |
| FR-MD-09 | Hỗ trợ warehouse và location master; phân loại kho nguyên liệu, thành phẩm, quarantine, sample/tester. | Must |
| FR-MD-10 | Hỗ trợ price list cơ bản theo nhóm khách hàng/kênh. | Must |
| FR-MD-11 | Hỗ trợ cấu hình rule sinh batch cho thành phẩm và lô nhập hàng. | Should |
| FR-MD-12 | Hỗ trợ gắn file đính kèm như COA template, hình ảnh sản phẩm, spec nội bộ. | Should |
| FR-MD-13 | Hỗ trợ khóa item/supplier/customer thay vì xóa nếu đã có giao dịch. | Must |
| FR-MD-14 | Hỗ trợ lịch sử thay đổi cho BOM, price list và item status. | Must |

### 12.6. Quy tắc nghiệp vụ chính

1. Không được tạo item trùng mã.
2. Item đã có giao dịch không được xóa.
3. Item có `batch_required = true` thì mọi receipt/issue phải mang batch.
4. Item có `expiry_required = true` thì receipt bắt buộc nhập expiry date.
5. BOM chỉ được active khi đầy đủ component, UOM và yield.
6. BOM active mới được dùng khi tạo production order.
7. Supplier không ở trạng thái approved thì không được chọn trên PO nếu item thuộc nhóm cần NCC đã duyệt.
8. Customer credit limit chỉ dùng tham chiếu trong Phase 1 để cảnh báo/approval, chưa tự động chặn mọi trường hợp nếu chưa cấu hình.

### 12.7. Acceptance Criteria

- Tạo mới item raw material có batch_required và expiry_required thành công.
- Không thể tạo item trùng item_code.
- BOM version mới có thể được lưu, nhưng khi active version mới thì version cũ tự về inactive.
- Supplier bị khóa không thể chọn trên PO mới.
- Customer có price list mặc định thì SO tự gợi ý giá theo price list đó.

---

## 13. Module 2 — Mua hàng (Procurement)

### 13.1. Mục tiêu nghiệp vụ

Kiểm soát vòng đời mua hàng từ nhu cầu mua đến nhận hàng thực tế, đồng thời gắn chặt với QC đầu vào và nhập kho.

### 13.2. Phạm vi Phase 1 của module Procurement

Bao gồm:

1. Purchase Requisition (PR)
2. Approval PR
3. Purchase Order (PO)
4. Approval PO
5. Goods Receiving against PO
6. Partial receiving / full receiving
7. Supplier return cơ bản
8. Theo dõi tình trạng PO

Không bao gồm sâu trong Phase 1:

- RFQ nhiều vòng nâng cao
- So sánh báo giá tự động nhiều nhà cung cấp
- Contract purchasing
- Supplier portal
- AP accounting hoàn chỉnh

### 13.3. Functional Requirements

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-PUR-01 | Hỗ trợ tạo PR với item, qty, required date, reason, department. | Must |
| FR-PUR-02 | PR có workflow draft → pending approval → approved/rejected/cancelled. | Must |
| FR-PUR-03 | Từ PR đã duyệt có thể tạo PO một phần hoặc toàn bộ. | Must |
| FR-PUR-04 | Hỗ trợ tạo PO trực tiếp nếu user có quyền. | Must |
| FR-PUR-05 | PO phải lưu supplier, item, qty, unit price, tax, expected date, payment term. | Must |
| FR-PUR-06 | PO có thể được approve theo ngưỡng giá trị. | Must |
| FR-PUR-07 | Hỗ trợ nhận hàng theo PO với partial receipt và full receipt. | Must |
| FR-PUR-08 | Hệ thống phải ghi nhận batch và expiry date khi receipt nếu item bắt buộc. | Must |
| FR-PUR-09 | Hàng receipt từ PO có trạng thái mặc định là QC Pending hoặc Available tùy item.qc_required. | Must |
| FR-PUR-10 | Hỗ trợ supplier return cho hàng fail QC hoặc trả lại sau nhận. | Should |
| FR-PUR-11 | Hỗ trợ đính kèm invoice / delivery note của supplier. | Should |
| FR-PUR-12 | Hỗ trợ theo dõi open PO, partially received, fully received, closed. | Must |
| FR-PUR-13 | Hỗ trợ cảnh báo nhận hàng vượt PO vượt quá tolerance cấu hình. | Must |
| FR-PUR-14 | Hỗ trợ lock chỉnh sửa PO sau khi đã có receiving, trừ một số field được phép. | Must |

### 13.4. Trạng thái chứng từ

#### PR
- Draft
- Pending Approval
- Approved
- Rejected
- Cancelled
- Closed

#### PO
- Draft
- Pending Approval
- Approved
- Partially Received
- Fully Received
- Closed
- Cancelled

#### Receiving
- Draft
- Confirmed
- Cancelled

### 13.5. Quy tắc nghiệp vụ chính

1. PR chưa duyệt không được tạo PO nếu không có quyền bypass.
2. PO chưa approved không được nhận hàng.
3. Không được nhận vượt PO quá tolerance cấu hình (ví dụ 2% hoặc theo policy).
4. Nếu item yêu cầu QC, hàng nhận phải đi vào trạng thái `QC Pending` hoặc `Quarantine`.
5. Nếu item không yêu cầu QC, hàng nhận có thể vào `Available`.
6. PO đã có receiving không được sửa item, qty chính, supplier; chỉ được chỉnh field mô tả/đính kèm nếu có quyền.
7. Supplier return phải tham chiếu receiving và batch đã nhập.

### 13.6. Màn hình chính và hành vi

#### PUR-01: Purchase Requisition
- Tạo PR từ form tay hoặc import.
- Chọn department yêu cầu.
- Chọn item, qty, required date, note.
- Submit for approval.

#### PUR-02: Purchase Order
- Tạo từ PR hoặc tạo trực tiếp.
- Chọn supplier.
- Kéo item từ PR hoặc nhập tay.
- Tính subtotal, tax, total.
- Submit for approval.

#### PUR-03: Receiving against PO
- Chọn PO approved.
- Hệ thống hiển thị remaining qty.
- Nhập số lượng nhận, batch, expiry, warehouse, location.
- Confirm receiving.
- Tự tạo stock ledger và record QC Pending nếu cần.

### 13.7. Acceptance Criteria

- Tạo PR và submit approval thành công.
- PO chỉ được tạo từ PR approved hoặc bởi user có quyền direct PO.
- Receiving partial lần 1 và lần 2 phải cộng dồn đúng vào PO.
- Hàng item có qc_required sau receiving không làm tăng available stock trước khi QC pass.
- Supplier return làm giảm stock đúng batch và cập nhật lịch sử PO/receipt.

---

## 14. Module 3 — QA/QC

### 14.1. Mục tiêu nghiệp vụ

Đảm bảo nguyên liệu và thành phẩm chỉ được đưa vào sử dụng hoặc đưa ra bán khi đạt chất lượng theo chuẩn nội bộ.

### 14.2. Phạm vi QC trong Phase 1

Bao gồm:

1. Inbound QC cho nguyên liệu / bao bì / vật tư cần kiểm
2. Finished Goods QC cho thành phẩm sau sản xuất
3. Quyết định QC: Pass / Fail / Hold
4. Release control
5. NCR cơ bản

Không bao gồm sâu trong Phase 1:

- CAPA workflow nâng cao
- Complaint management toàn diện
- Stability study / regulatory management
- Laboratory integration

### 14.3. Functional Requirements

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-QC-01 | Hệ thống tự sinh phiếu QC inbound khi receiving của item.qc_required = true. | Must |
| FR-QC-02 | Hỗ trợ QC officer nhập kết quả kiểm và đề xuất pass/fail/hold. | Must |
| FR-QC-03 | Hỗ trợ QA manager duyệt quyết định QC. | Must |
| FR-QC-04 | Khi QC pass, tồn từ trạng thái pending/quarantine chuyển sang available. | Must |
| FR-QC-05 | Khi QC fail, tồn chuyển sang fail/quarantine và không khả dụng cho sản xuất hoặc bán hàng. | Must |
| FR-QC-06 | Hỗ trợ QC cho thành phẩm sau sản xuất trước khi đưa vào available stock. | Must |
| FR-QC-07 | Hỗ trợ tạo NCR cơ bản kèm nguyên nhân, mô tả, hướng xử lý. | Should |
| FR-QC-08 | Hỗ trợ đính kèm COA, ảnh, file kiểm nghiệm nội bộ. | Should |
| FR-QC-09 | Hỗ trợ dashboard pending QC, pass/fail theo thời gian, theo supplier, theo item. | Must |
| FR-QC-10 | Hệ thống phải lưu người kiểm, thời gian kiểm, người duyệt release. | Must |

### 14.4. Trạng thái QC

- Draft
- In Inspection
- Pending Approval
- Passed
- Failed
- On Hold
- Cancelled

### 14.5. Quy tắc nghiệp vụ chính

1. Hàng inbound cần QC không được dùng cho sản xuất hoặc bán trước khi pass.
2. Thành phẩm sau sản xuất cần QC không được vào available stock trước khi pass.
3. Quyết định `Hold` phải chặn usage và chờ xử lý tiếp.
4. Quyết định `Fail` phải chặn usage; cho phép return supplier, destroy hoặc chuyển kho lỗi theo quyền.
5. QA manager hoặc vai trò tương đương mới được release cuối cùng nếu cấu hình yêu cầu.
6. Mọi thay đổi từ Fail/Hold sang Pass phải có quyền đặc biệt và audit log.

### 14.6. Màn hình chính và hành vi

#### QC-01: Inbound Inspection
- Tự tạo từ receiving hoặc tạo tay bởi QC.
- Hiển thị item, supplier, batch, qty, receiving ref.
- Nhập chỉ tiêu kiểm, kết quả, conclusion.

#### QC-02: FG Inspection
- Tạo từ production output completed.
- Gắn production order, FG batch, qty.
- Nhập chỉ tiêu theo thành phẩm.

#### QC-03: QC Decision / Release
- Duyệt pass/fail/hold.
- Chọn action khi fail.
- Tự cập nhật stock status.

### 14.7. Acceptance Criteria

- Receiving của item cần QC phải sinh record QC inbound.
- Sau khi inbound QC pass, available stock tăng đúng batch và qty.
- Sau khi inbound QC fail, available stock không tăng; stock vẫn nằm fail/quarantine.
- Thành phẩm chỉ vào available sau FG QC pass.
- User không có quyền release không thể thay đổi quyết định final.

---

## 15. Module 4 — Sản xuất (Production)

### 15.1. Mục tiêu nghiệp vụ

Biến nguyên vật liệu thành thành phẩm có thể bán được, trong khi ghi nhận đầy đủ cấp phát, tiêu hao thực tế, hao hụt và batch thành phẩm.

### 15.2. Phạm vi Production trong Phase 1

Bao gồm:

1. Tạo production order thủ công
2. Chọn BOM active
3. Reserve / issue nguyên liệu cho lệnh
4. Ghi actual consumption
5. Ghi output thành phẩm
6. Ghi scrap / hao hụt
7. Sinh batch thành phẩm
8. Chuyển thành phẩm sang QC / FG stock pending

Không bao gồm sâu trong Phase 1:

- Scheduling APS phức tạp
- OEE / machine sensor integration
- Multi-stage routing phức tạp
- Rework phức tạp
- Subcontract manufacturing
- Full MRP tự động

### 15.3. Functional Requirements

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-PROD-01 | Hỗ trợ tạo production order cho một thành phẩm dựa trên BOM active. | Must |
| FR-PROD-02 | Hệ thống tự tính planned material requirement theo BOM và qty planned. | Must |
| FR-PROD-03 | Hỗ trợ reserve nguyên liệu cho lệnh sản xuất. | Must |
| FR-PROD-04 | Hỗ trợ issue nguyên liệu theo batch thực tế từ kho nguyên liệu. | Must |
| FR-PROD-05 | Hỗ trợ nhập actual consumption khác planned và ghi variance. | Must |
| FR-PROD-06 | Hỗ trợ ghi output thành phẩm, bán thành phẩm nếu có, và scrap/loss. | Must |
| FR-PROD-07 | Hỗ trợ sinh batch thành phẩm theo rule hoặc nhập tay có kiểm soát. | Must |
| FR-PROD-08 | Thành phẩm output phải vào trạng thái QC Pending nếu FG cần QC. | Must |
| FR-PROD-09 | Hỗ trợ trạng thái lệnh: draft, released, in progress, completed, closed, cancelled. | Must |
| FR-PROD-10 | Hỗ trợ in / xem batch trace từ FG batch về nguyên liệu issue chính. | Must |
| FR-PROD-11 | Hỗ trợ đính kèm hồ sơ lệnh sản xuất / batch record cơ bản. | Should |
| FR-PROD-12 | Hỗ trợ dashboard lệnh đang chạy, completed, variance, scrap. | Must |

### 15.4. Trạng thái Production Order

- Draft
- Released
- In Progress
- Completed
- QC Pending
- Closed
- Cancelled

Lưu ý: `QC Pending` có thể là trạng thái logic của output, còn production order có thể ở `Completed` nếu thao tác output xong. Tùy thiết kế chi tiết, có thể giữ production order `Completed` và QC status nằm ở output batch. Tuy nhiên dashboard phải thể hiện rõ batch thành phẩm nào chưa release.

### 15.5. Quy tắc nghiệp vụ chính

1. Production order chỉ dùng BOM active.
2. Không được release lệnh nếu thiếu nguyên liệu bắt buộc theo policy, trừ khi có quyền override.
3. Issue nguyên liệu phải chọn batch còn available.
4. Issue không được vượt available stock.
5. Actual consumption phải được lưu riêng, không ghi đè planned BOM.
6. Output thành phẩm phải có batch.
7. Thành phẩm output chưa QC pass không được xuất bán.
8. Khi hủy lệnh, các nguyên liệu đã reserve chưa issue phải được giải phóng reserve.
9. Nếu đã issue nguyên liệu và có output, không được hard delete lệnh.

### 15.6. Màn hình chính và hành vi

#### PROD-01: Production Order
- Chọn FG item, qty planned, planned date, warehouse output.
- Hệ thống load BOM active.
- Tự tính requirement.

#### PROD-03: Material Issue
- Hiển thị planned requirement.
- Chọn batch thực issue cho từng component.
- Xác nhận qty issue.
- Tạo stock issue ledger.

#### PROD-04/05: Production Confirmation & Output
- Nhập actual output good qty.
- Nhập scrap / loss.
- Sinh FG batch.
- Đưa output về FG pending QC hoặc available tùy item.

### 15.7. Acceptance Criteria

- Tạo production order từ FG có BOM active thành công.
- Hệ thống tự tính requirement đúng theo BOM × planned qty.
- Material issue không cho chọn batch không available.
- Actual consumption và variance được lưu trên lệnh.
- FG batch được tạo và trace được tới component batch issue.
- FG output chưa QC pass thì không thể allocate cho sales order.

---

## 16. Module 5 — Kho hàng (Warehouse / WMS cơ bản)

### 16.1. Mục tiêu nghiệp vụ

Tạo sổ cái hàng hóa theo batch, kho, location, trạng thái tồn; kiểm soát nhập, xuất, chuyển, kiểm kê và tồn khả dụng.

### 16.2. Phạm vi Warehouse trong Phase 1

Bao gồm:

1. Goods receipt
2. Goods issue
3. Transfer order / transfer receipt
4. Stock count
5. Stock adjustment
6. Stock ledger
7. Batch / expiry tracking
8. Reserve stock
9. Near-expiry reporting
10. Batch trace

### 16.3. Functional Requirements

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-WH-01 | Mọi biến động hàng hóa phải tạo stock ledger. | Must |
| FR-WH-02 | Hệ thống quản lý stock theo warehouse, location, item, batch, expiry, status. | Must |
| FR-WH-03 | Hỗ trợ goods receipt từ PO, từ production output, từ adjustment. | Must |
| FR-WH-04 | Hỗ trợ goods issue cho production, cho sales delivery, cho adjustment. | Must |
| FR-WH-05 | Hỗ trợ transfer giữa kho/location. | Must |
| FR-WH-06 | Hỗ trợ reserve stock cho sales order hoặc production order. | Must |
| FR-WH-07 | Hỗ trợ stock count và tạo adjustment với approval. | Must |
| FR-WH-08 | Hỗ trợ FEFO suggestion khi xuất hàng cho item có expiry. | Should |
| FR-WH-09 | Hỗ trợ xem available, reserved, quarantine, fail stock tách biệt. | Must |
| FR-WH-10 | Hỗ trợ báo cáo near-expiry theo ngưỡng ngày cấu hình. | Must |
| FR-WH-11 | Hỗ trợ batch trace theo hướng: batch thành phẩm → component batch, receiving batch → usage. | Must |
| FR-WH-12 | Hỗ trợ quản lý stock type riêng cho sample/tester/quarantine. | Should |

### 16.4. Trạng thái/tồn kho cần theo dõi

Mỗi stock record cần ít nhất các trạng thái logic sau:

- Available
- Reserved
- QC Pending
- On Hold
- Fail / Rejected
- Quarantine
- Sample / Tester (nếu triển khai ở Phase 1)
- In Transit (nếu dùng cho transfer)

### 16.5. Quy tắc nghiệp vụ chính

1. Tồn vật lý = tổng stock theo mọi trạng thái.
2. Tồn khả dụng = Available - Reserved (nếu reserve được tách logic riêng) hoặc tổng Available chưa reserve.
3. Hàng QC Pending/Hold/Fail không được coi là available.
4. Goods issue cho sales hoặc production chỉ lấy từ available.
5. Transfer giữa kho phải bảo toàn batch, expiry và status theo rule.
6. Stock adjustment phải có lý do và approval nếu vượt ngưỡng.
7. Kiểm kê chốt ra chênh lệch thì phải tạo adjustment riêng, không sửa tay stock ledger.
8. Batch đã dùng hết vẫn phải giữ lịch sử trong stock ledger.

### 16.6. Màn hình chính và hành vi

#### WH-07: Stock Ledger
- Bộ lọc theo kho, item, batch, date range, status.
- Hiển thị opening, in, out, closing.

#### WH-08: Stock by Batch / Expiry
- Xem tồn theo batch, expiry, số ngày còn hạn.
- Flag cận date theo ngưỡng.

#### WH-09: Batch Trace
- Với FG batch: xem production order, material issue batches.
- Với inbound batch: xem đã issue cho lệnh nào / đã bán batch nào.

### 16.7. Acceptance Criteria

- Sau mỗi receiving/issue/transfer/adjustment, stock ledger cập nhật đúng.
- Hệ thống hiển thị đúng available vs quarantine vs reserved.
- Near-expiry report hiển thị đúng số ngày còn lại.
- Batch trace cho một FG batch hiển thị được production order và batch nguyên liệu chính.
- Không thể issue hàng fail QC cho sản xuất hoặc sales.

---

## 17. Module 6 — Bán hàng (Sales / OMS cơ bản)

### 17.1. Mục tiêu nghiệp vụ

Kiểm soát báo giá/đơn bán cơ bản, giữ tồn, xuất hàng và trả hàng; tạo nền cho đa kênh và CRM ở phase sau.

### 17.2. Phạm vi Sales trong Phase 1

Bao gồm:

1. Quotation cơ bản
2. Sales Order
3. Approval SO theo discount / credit / ngưỡng
4. Stock reservation
5. Delivery request / delivery order
6. Sales return cơ bản
7. Theo dõi trạng thái đơn

Không bao gồm sâu trong Phase 1:

- Marketplace integration
- Website integration realtime
- Shipping carrier integration sâu
- Loyalty / CRM automation
- Commission engine phức tạp
- POS đầy đủ

### 17.3. Functional Requirements

| ID | Yêu cầu | Priority |
|---|---|---|
| FR-SAL-01 | Hỗ trợ tạo quotation cơ bản và convert sang sales order. | Should |
| FR-SAL-02 | Hỗ trợ tạo sales order với customer, channel, item, qty, unit price, discount. | Must |
| FR-SAL-03 | Hệ thống tự gợi ý giá theo price list/customer/channel. | Must |
| FR-SAL-04 | Hỗ trợ approval cho SO nếu discount vượt ngưỡng hoặc khách vượt credit limit tham chiếu. | Must |
| FR-SAL-05 | SO approved có thể reserve tồn khả dụng. | Must |
| FR-SAL-06 | Hỗ trợ partial delivery và full delivery. | Must |
| FR-SAL-07 | Hệ thống không cho confirm delivery vượt available stock trừ user có quyền override. | Must |
| FR-SAL-08 | Hỗ trợ sales return tham chiếu SO/DO và batch trả về. | Must |
| FR-SAL-09 | Hỗ trợ trạng thái SO từ draft đến closed/cancelled. | Must |
| FR-SAL-10 | Hỗ trợ báo cáo open SO, delivered, pending delivery, sales return. | Must |
| FR-SAL-11 | Hỗ trợ ghi nhận địa chỉ giao hàng, ghi chú giao hàng, contact person. | Should |
| FR-SAL-12 | Hỗ trợ flag đơn thủ công theo channel: B2B, retail, online manual, affiliate manual. | Should |

### 17.4. Trạng thái chứng từ

#### Quotation
- Draft
- Sent
- Approved (nếu cần)
- Expired
- Converted
- Cancelled

#### Sales Order
- Draft
- Pending Approval
- Confirmed
- Reserved
- Partially Delivered
- Delivered
- Closed
- Cancelled

#### Delivery Order
- Draft
- Confirmed
- Shipped
- Delivered
- Cancelled

#### Sales Return
- Draft
- Pending Approval
- Received
- Closed
- Cancelled

### 17.5. Quy tắc nghiệp vụ chính

1. Giá mặc định lấy từ price list; user có quyền có thể chỉnh discount trong ngưỡng.
2. Discount vượt ngưỡng phải xin duyệt.
3. SO chưa approved không được reserve stock.
4. Reserve chỉ lấy từ available stock.
5. Delivery chỉ issue được từ batch available.
6. Item batch_required phải xuất theo batch.
7. Sales return phải xác định tình trạng hàng trả về:
   - sellable / return to available,
   - quarantine,
   - fail/destroy.
8. Không được hủy SO đã giao hết nếu không qua quy trình return/reversal.
9. Customer bị khóa hoặc vượt credit rule có thể bị cảnh báo hoặc chặn tùy cấu hình.

### 17.6. Màn hình chính và hành vi

#### SAL-02: Sales Order
- Chọn customer, channel, warehouse, delivery date.
- Chọn item, qty.
- Hệ thống gợi ý price list.
- Tính subtotal, discount, tax, total.
- Submit approval nếu cần.

#### SAL-04: Delivery Request / Delivery Order
- Chọn SO confirmed.
- Hệ thống hiển thị reserved/available.
- Chọn batch xuất.
- Confirm delivery.

#### SAL-05: Sales Return
- Chọn DO/SO gốc.
- Chọn item, qty trả.
- Chọn batch trả về hoặc hệ thống gợi ý nếu có.
- Chọn disposition: available / quarantine / fail.
- Confirm return.

### 17.7. Acceptance Criteria

- SO lấy đúng giá mặc định theo price list của customer/channel.
- Discount vượt ngưỡng phải chuyển `Pending Approval`.
- SO approved có thể reserve stock và giảm available-to-promise.
- Delivery không cho xuất quá available stock.
- Sales return cập nhật stock đúng batch và đúng disposition.
- Batch chưa QC pass không xuất được cho SO.

---

## 18. Luồng nghiệp vụ end-to-end bắt buộc trong Phase 1

### 18.1. Luồng 1 — Mua hàng đến nhập kho và QC

```text
PR → phê duyệt PR → PO → phê duyệt PO → receiving → inbound QC → pass/fail → stock available/quarantine
```

**Kết quả mong muốn:** nguyên liệu pass QC mới trở thành tồn khả dụng cho sản xuất.

### 18.2. Luồng 2 — Sản xuất thành phẩm

```text
Production Order → reserve/issue NVL → actual consumption → output FG batch → FG QC → available FG stock
```

**Kết quả mong muốn:** thành phẩm có batch, trace được nguồn nguyên liệu và chỉ bán được khi QC pass.

### 18.3. Luồng 3 — Bán hàng đến xuất kho

```text
Quotation/SO → approval → reserve stock → delivery order → goods issue → delivered
```

**Kết quả mong muốn:** không bán vượt tồn khả dụng, batch xuất được kiểm soát.

### 18.4. Luồng 4 — Trả hàng

```text
Sales Return Request → approval → return receipt → disposition (available/quarantine/fail)
```

**Kết quả mong muốn:** hàng trả về không nhập mù vào kho; phải phân loại lại trạng thái.

---

## 19. Ma trận phê duyệt đề xuất cho Phase 1

| Giao dịch | Điều kiện | Người duyệt đề xuất |
|---|---|---|
| PR | Theo mọi PR hoặc theo ngưỡng | Trưởng bộ phận / Purchasing Manager |
| PO | Tổng giá trị vượt ngưỡng A | Purchasing Manager |
| PO | Tổng giá trị vượt ngưỡng B | Purchasing Manager + Finance/CEO |
| Stock Adjustment | Chênh lệch nhỏ | Warehouse Manager |
| Stock Adjustment | Chênh lệch lớn | Warehouse Manager + COO/Finance |
| QC Release | Pass/Fail/Hold | QA Manager |
| Production Order | Qty / giá trị vượt ngưỡng | Production Manager / COO |
| SO | Discount vượt ngưỡng | Sales Manager |
| SO | Khách vượt credit limit | Sales Manager / Finance |
| Sales Return | Giá trị thấp | Sales Manager / Warehouse Manager |
| Sales Return | Giá trị cao / batch nghi vấn | Sales Manager + QA |

> Ghi chú: ngưỡng A/B phải được cấu hình trong tài liệu policy riêng hoặc trong bảng cấu hình hệ thống.

---

## 20. Quy ước trạng thái và đánh số chứng từ

### 20.1. Đánh số đề xuất

- PR: `PR-YYYYMM-####`
- PO: `PO-YYYYMM-####`
- GRN/Receiving: `GRN-YYYYMM-####`
- IQC: `IQC-YYYYMM-####`
- FGQC: `FGQC-YYYYMM-####`
- MO: `MO-YYYYMM-####`
- GI (Goods Issue): `GI-YYYYMM-####`
- TO (Transfer Order): `TO-YYYYMM-####`
- SC (Stock Count): `SC-YYYYMM-####`
- SO: `SO-YYYYMM-####`
- DO: `DO-YYYYMM-####`
- SR (Sales Return): `SR-YYYYMM-####`

### 20.2. Quy ước batch đề xuất

#### Batch nguyên liệu inbound
- Có thể dùng batch từ supplier + mapping nội bộ, hoặc
- Hệ thống sinh: `RM-[item_code]-[YYMMDD]-[seq]`

#### Batch thành phẩm
- `FG-[sku_code]-[YYMMDD]-[seq]`

> Nếu công ty đã có quy ước batch riêng, hệ thống phải cho phép cấu hình hoặc nhập tay có validate.

---

## 21. Dữ liệu bắt buộc trên các chứng từ chính

### 21.1. PR
- pr_no
- requester
- department
- request_date
- item
- qty
- required_date
- reason
- approval_status

### 21.2. PO
- po_no
- supplier
- po_date
- item lines
- qty
- unit price
- expected receipt date
- warehouse
- payment term
- tax
- total amount
- approval_status

### 21.3. Receiving
- receipt_no
- po_ref
- supplier
- receipt_date
- item
- qty_received
- warehouse
- location
- batch
- expiry_date
- qc_status_default

### 21.4. QC record
- qc_no
- qc_type (inbound / fg)
- reference_doc
- item
- batch
- qty
- sample size (nếu dùng)
- inspection date
- result summary
- decision
- approver
- attachment

### 21.5. Production Order
- mo_no
- fg_item
- bom_version
- planned_qty
- planned_date
- source warehouse
- output warehouse
- status

### 21.6. Material Issue
- issue_no
- mo_ref
- issue_date
- component item
- batch
- qty_issue
- warehouse/location
- operator

### 21.7. SO
- so_no
- customer
- channel
- order_date
- delivery_date
- warehouse
- item
- qty
- unit price
- discount
- total
- approval_status

### 21.8. Delivery Order
- do_no
- so_ref
- warehouse
- item
- qty
- batch
- ship_to
- status

### 21.9. Sales Return
- sr_no
- so_ref / do_ref
- customer
- item
- qty_return
- return_reason
- disposition
- batch_return
- approval_status

---

## 22. Báo cáo và dashboard bắt buộc cho Phase 1

### 22.1. Dashboard CEO / COO
1. Tồn kho tổng theo loại hàng
2. Tồn khả dụng vs quarantine
3. Hàng cận date
4. Open PO / overdue PO
5. QC pending / hold / fail
6. Lệnh sản xuất đang chạy
7. Sản lượng hoàn thành theo ngày
8. Doanh thu cơ bản theo ngày / kênh
9. Open SO / pending delivery

### 22.2. Procurement reports
1. Danh sách PR mở
2. Danh sách PO mở / quá hạn nhận
3. Lịch sử mua theo item / supplier
4. Supplier delivery performance cơ bản

### 22.3. QC reports
1. Inbound QC pending
2. FG QC pending
3. Tỷ lệ pass/fail theo supplier
4. Tỷ lệ pass/fail theo item
5. Danh sách batch hold

### 22.4. Production reports
1. Production order status
2. Planned vs actual consumption
3. Scrap / loss by order
4. FG output by batch/date

### 22.5. Warehouse reports
1. Stock on hand by warehouse
2. Stock by batch / expiry
3. Stock movement ledger
4. Near-expiry report
5. Reserved stock report

### 22.6. Sales reports
1. Open quotation / SO
2. Sales by item / customer / channel
3. Pending delivery
4. Sales return by reason
5. Rough gross sales report (doanh thu chưa trừ sâu)

---

## 23. Non-Functional Requirements (NFR)

| ID | Yêu cầu | Priority |
|---|---|---|
| NFR-01 | Hệ thống là web app chạy trên trình duyệt desktop hiện đại. | Must |
| NFR-02 | Giao diện responsive cơ bản cho tablet ở màn hình kho/chứng từ. | Should |
| NFR-03 | Thời gian phản hồi các màn hình danh sách phổ biến dưới 3 giây với dataset vận hành thông thường. | Must |
| NFR-04 | Hệ thống phải hỗ trợ ít nhất 100 người dùng nội bộ đồng thời ở Phase 1. | Should |
| NFR-05 | Role-based access control bắt buộc cho tất cả module. | Must |
| NFR-06 | Audit log phải lưu user, timestamp, action, before/after summary cho dữ liệu quan trọng. | Must |
| NFR-07 | Dữ liệu ngày giờ lưu theo chuẩn có timezone; hiển thị theo Asia/Ho_Chi_Minh. | Must |
| NFR-08 | Hỗ trợ backup/restore theo chính sách hạ tầng. | Must |
| NFR-09 | Hỗ trợ export Excel/CSV cho các màn hình danh sách chính. | Must |
| NFR-10 | Hệ thống phải có cơ chế validate dữ liệu đầu vào bắt buộc. | Must |
| NFR-11 | Không cho sửa tay trực tiếp stock ledger. | Must |
| NFR-12 | Hỗ trợ phân trang, tìm kiếm, lọc theo trạng thái, ngày, kho, batch ở các màn hình chính. | Must |

---

## 24. Ràng buộc dữ liệu và kiểm soát tính toàn vẹn

1. `item_code`, `supplier_code`, `customer_code`, `warehouse_code` là duy nhất.
2. Một chứng từ chỉ có thể chuyển trạng thái theo flow hợp lệ.
3. Batch và expiry không được bỏ trống nếu item yêu cầu.
4. Stock ledger là bất biến; mọi điều chỉnh tạo transaction mới.
5. Delivery không thể âm tồn available.
6. Material issue không thể âm tồn available.
7. BOM version active phải hợp lệ về UOM và component.
8. Sales return không thể vượt số lượng đã giao trừ quyền đặc biệt.
9. Supplier return không thể vượt số lượng đã nhận trừ quyền đặc biệt.
10. Hủy chứng từ đã phát sinh hậu quả downstream phải đi theo reverse flow, không sửa tay.

---

## 25. Yêu cầu tích hợp tối thiểu trong Phase 1

Phase 1 ưu tiên xây lõi trước. Tích hợp tối thiểu nên ở mức:

1. **Import/Export Excel**
   - Item master
   - Supplier master
   - Customer master
   - Opening stock
   - Opening price list
   - Opening SO/PO nếu cần

2. **Barcode / batch scan (tùy chọn)**
   - Nếu chưa tích hợp thiết bị scanner, vẫn phải cho nhập batch nhanh.

3. **Email / notification nội bộ**
   - Gửi thông báo approval hoặc cảnh báo QC pending / PO overdue.

> Tất cả tích hợp thời gian thực với website, sàn, vận chuyển, kế toán ngoài để phase sau.

---

## 26. Migration requirements

### 26.1. Dữ liệu cần migrate trước go-live

1. Item master
2. Supplier master
3. Customer master
4. Warehouse & location master
5. BOM active
6. Price list
7. Opening stock by warehouse/batch/expiry/status
8. Open PO (nếu quyết định migrate)
9. Open SO (nếu quyết định migrate)

### 26.2. Điều kiện migrate

- Dữ liệu phải clean và mapping một lần cuối.
- Mọi item phải có base_uom.
- Mọi stock opening cho item batch-required phải có batch và expiry nếu cần.
- Chỉ migrate BOM đã được business xác nhận active.

---

## 27. UAT scenarios bắt buộc

### 27.1. Master Data
1. Tạo mới raw material có batch/expiry/QC.
2. Tạo BOM cho FG và active version.
3. Khóa supplier đã có giao dịch, không xóa được.

### 27.2. Procurement
1. Tạo PR → approve → tạo PO → approve → receiving partial.
2. Receiving lần 2 cho cùng PO.
3. Hàng item cần QC sau receiving không tăng available stock.

### 27.3. QC
1. Inbound QC pass làm tăng available stock.
2. Inbound QC fail giữ hàng ở quarantine/fail.
3. FG QC pass release batch thành phẩm.

### 27.4. Production
1. Tạo production order từ BOM active.
2. Issue nguyên liệu theo batch.
3. Xác nhận output FG, scrap, variance.
4. Truy batch FG về component batch.

### 27.5. Warehouse
1. Transfer warehouse vẫn giữ batch/expiry.
2. Stock count tạo adjustment.
3. Near-expiry report ra đúng lô.

### 27.6. Sales
1. SO lấy đúng giá từ price list.
2. SO discount vượt ngưỡng phải approve.
3. Delivery không vượt available stock.
4. Sales return nhập lại với disposition quarantine.

---

## 28. Backlog ưu tiên triển khai đề xuất

### Sprint / Giai đoạn 0 — Nền móng
- User / Role / Permission
- Numbering
- Approval engine cơ bản
- Audit log
- Attachment
- Master Data cơ bản: item, supplier, customer, warehouse, UOM

### Sprint / Giai đoạn 1 — Kho + Mua + QC đầu vào
- PR / PO
- Receiving
- Inbound QC
- Stock ledger
- Stock by batch/status

### Sprint / Giai đoạn 2 — Sản xuất
- BOM
- Production order
- Material issue
- Output / scrap
- FG QC
- Batch trace cơ bản

### Sprint / Giai đoạn 3 — Bán hàng
- Price list
- Quotation/SO
- Approval SO
- Reserve stock
- Delivery
- Sales return

### Sprint / Giai đoạn 4 — Báo cáo + UAT + Go-live
- Dashboards
- Core reports
- UAT fixes
- Migration
- User training
- Go-live checklist

---

## 29. Go-live checklist cấp cao

1. Item master chốt
2. BOM active chốt
3. Supplier/customer/warehouse chốt
4. Price list chốt
5. Opening stock kiểm
6. Phân quyền user chốt
7. Approval matrix chốt
8. UAT pass cho 6 module
9. Training user key
10. Kế hoạch cutover và backup rõ ràng
11. Quy trình hỗ trợ 2 tuần đầu sau go-live

---

## 30. Rủi ro chính của Phase 1 và cách giảm

### Rủi ro 1: Dữ liệu master bẩn
**Giảm bằng:** data cleansing trước migrate, owner từng danh mục.

### Rủi ro 2: Không thống nhất quy trình
**Giảm bằng:** sign-off flow PR → PO → QC → Production → SO.

### Rủi ro 3: Kho không nhập batch/expiry kỷ luật
**Giảm bằng:** bắt buộc validate ở receipt/issue; training và audit.

### Rủi ro 4: Business muốn nhét thêm CRM/KOL/HRM ngay
**Giảm bằng:** giữ ranh giới Phase 1; backlog phase 2.

### Rủi ro 5: User vẫn chạy Excel song song
**Giảm bằng:** quyết định “source of truth” rõ từ ngày go-live.

---

## 31. Những quyết định cần business sign-off trước khi dev build

1. Quy ước mã item, mã chứng từ, mã batch.
2. Danh sách loại item và category.
3. Danh sách warehouse/location và stock type.
4. Chính sách QC inbound/FG cho từng nhóm item.
5. Approval threshold cho PO, stock adjustment, SO discount, sales return.
6. Chính sách giá/discount cơ bản.
7. Quy tắc xử lý sales return: về available hay quarantine theo từng lý do.
8. Chính sách tolerance khi receiving vượt PO.
9. Chính sách cho phép direct PO/direct SO của vai trò nào.
10. Ranh giới multi-channel trong Phase 1: nhập tay hay import.

---

## 32. Kết luận

Phase 1 của ERP cho công ty mỹ phẩm phải làm đúng một việc: **khóa vận hành lõi từ nguyên liệu đến thành phẩm, tồn kho và đơn bán**.

Nếu 6 module trong tài liệu này được build đúng, công ty sẽ có:

- dữ liệu gốc sạch hơn,
- tồn kho tin cậy hơn,
- QC rõ ràng hơn,
- sản xuất ít mù hơn,
- bán hàng bớt bán quá tay,
- nền tốt để mở rộng CRM, HRM, KOL, tài chính sâu và tích hợp đa kênh ở Phase 2.

Tư duy đúng ở đây không phải là “ERP càng nhiều tính năng càng tốt”.  
Tư duy đúng là: **ERP phải khóa đúng điểm rò máu trước, rồi mới tối ưu tăng trưởng sau**.

---

## 33. Phụ lục A — Danh sách trạng thái tóm tắt

| Đối tượng | Trạng thái |
|---|---|
| PR | Draft, Pending Approval, Approved, Rejected, Cancelled, Closed |
| PO | Draft, Pending Approval, Approved, Partially Received, Fully Received, Closed, Cancelled |
| Receiving | Draft, Confirmed, Cancelled |
| QC | Draft, In Inspection, Pending Approval, Passed, Failed, On Hold, Cancelled |
| Production Order | Draft, Released, In Progress, Completed, Closed, Cancelled |
| Stock | Available, Reserved, QC Pending, On Hold, Fail, Quarantine, Sample/Tester, In Transit |
| Quotation | Draft, Sent, Approved, Expired, Converted, Cancelled |
| SO | Draft, Pending Approval, Confirmed, Reserved, Partially Delivered, Delivered, Closed, Cancelled |
| DO | Draft, Confirmed, Shipped, Delivered, Cancelled |
| Sales Return | Draft, Pending Approval, Received, Closed, Cancelled |

---

## 34. Phụ lục B — Công thức hiển thị tồn kho đề xuất

### 34.1. Tồn vật lý
`Physical Stock = Available + Reserved + QC Pending + Hold + Fail + Quarantine + Sample/Tester + In Transit (nếu cùng quyền sở hữu)`

### 34.2. Tồn khả dụng
`Available to Promise = Available - Reserved`

### 34.3. Tồn cận date
`Near Expiry Stock = tổng tồn của các batch có số ngày tới expiry <= ngưỡng cấu hình`

---

## 35. Phụ lục C — Traceability tối thiểu bắt buộc

### 35.1. Chiều xuôi
Receiving batch nguyên liệu  
→ inbound QC  
→ material issue batch  
→ production order  
→ FG batch  
→ FG QC  
→ delivery batch  

### 35.2. Chiều ngược
Sales return / complaint batch  
→ FG batch  
→ production order  
→ material issue batches  
→ inbound receiving batch  
→ supplier

> Ở Phase 1, complaint management chưa làm sâu, nhưng traceability data model phải sẵn sàng cho chiều ngược.

---

## 36. Phụ lục D — Danh sách Must / Should / Could tổng hợp

### Must
- Item master
- Supplier/customer/warehouse master
- BOM active
- PR/PO
- Receiving
- Inbound QC
- Production order
- Material issue
- FG output & FG QC
- Stock ledger
- Reserve stock
- SO
- Delivery
- Sales return
- Audit log
- Approval engine cơ bản
- Core reports

### Should
- Quotation
- FEFO suggestion
- Supplier return
- Batch rule config linh hoạt
- Attachment mở rộng
- Responsive tablet cơ bản
- Import/export mở rộng
- Dashboard chi tiết hơn

### Could
- Barcode scan tích hợp sâu
- Rule engine nâng cao cho pricing
- Notification nâng cao
- QR label printing nâng cao

---

## 37. Đề xuất tên file tài liệu

`03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
