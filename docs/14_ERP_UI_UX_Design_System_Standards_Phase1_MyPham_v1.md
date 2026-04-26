# 14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1

**Project:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** UI/UX Design System Standards  
**Scope:** Phase 1  
**Version:** v1.0  
**Date:** 2026-04-24  
**Language:** Vietnamese  
**Primary Audience:** CEO, Product Owner, BA, UI/UX Designer, Frontend Lead, Backend Lead, QA, Vendor triển khai  

**Related Documents:**
- `ERP_Blueprint_My_Pham_v1.md`
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

**Workflow thực tế tham chiếu:**
- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt chuẩn **UI/UX Design System** cho ERP Phase 1.

Nếu `08_ERP_Screen_List_Wireframe` trả lời câu hỏi **có những màn hình nào**, thì tài liệu này trả lời câu hỏi:

- Mỗi màn hình phải nhìn và vận hành theo chuẩn nào?
- Component nào được tái sử dụng?
- Table, form, status, action, modal, drawer, scan, warning, approval, audit log phải thiết kế ra sao?
- Giao diện kho, QC, sản xuất, sales, shipping, hàng hoàn, gia công ngoài cần ưu tiên điều gì?
- Designer, frontend, backend, QA nói chung một ngôn ngữ UI/UX như thế nào?

Nói thẳng: ERP không cần màu mè. ERP cần **nhanh, rõ, khó thao tác sai, dễ kiểm soát và dễ truy vết**.

Một màn hình ERP tốt không phải là màn hình “đẹp để trình diễn”. Màn hình tốt là màn hình giúp nhân viên làm đúng việc trong áp lực thật: đơn đang chờ, hàng đang nằm trên kệ, shipper đang đứng chờ, kho cần đối soát cuối ca, QC cần khóa batch, quản lý cần biết lệch ở đâu.

---

## 2. Phạm vi tài liệu

### 2.1. In scope

Áp dụng cho các màn hình Phase 1:

1. Common/System UI
2. Master Data
3. Procurement / Mua hàng
4. QA/QC
5. Production / Gia công ngoài / Sản xuất
6. Warehouse / Kho
7. Sales Order / Bán hàng
8. Shipping / Bàn giao ĐVVC
9. Returns / Hàng hoàn
10. Basic Finance visibility / Công nợ, thu chi cơ bản
11. Dashboard vận hành Phase 1

### 2.2. Out of scope

Tài liệu này chưa thiết kế chi tiết cho:

- CRM nâng cao;
- HRM đầy đủ;
- KOL/Affiliate đầy đủ;
- POS retail hoàn chỉnh;
- accounting posting chuyên sâu;
- mobile app native riêng;
- full brand identity/public website;
- design prototype pixel-perfect cho toàn bộ màn hình.

### 2.3. Mức áp dụng

Tài liệu này là **chuẩn bắt buộc** cho UI/UX và frontend trong Phase 1. Nếu màn hình nào muốn đi khác chuẩn, phải có lý do nghiệp vụ rõ và được Product Owner duyệt.

---

## 3. Triết lý thiết kế UI/UX cho ERP mỹ phẩm

### 3.1. Ưu tiên vận hành thật hơn trình diễn

ERP là công cụ làm việc. Người dùng dùng nó nhiều giờ mỗi ngày. Do đó:

- ít click;
- ít nhập tay;
- nhiều kiểm tra tự động;
- status rõ;
- dữ liệu quan trọng không bị che;
- thao tác nguy hiểm phải có xác nhận;
- lỗi phải nói rõ nguyên nhân, không nói chung chung.

### 3.2. “Thấy trạng thái trước, thấy chi tiết sau”

Người dùng không nên phải mở từng chứng từ mới biết tình trạng.

Ở mọi list/table quan trọng phải hiển thị ngay:

- mã chứng từ;
- đối tượng liên quan;
- trạng thái;
- người phụ trách;
- hạn xử lý;
- cảnh báo;
- bước tiếp theo.

Ví dụ đơn hàng nên thấy ngay: `Confirmed`, `Reserved`, `Picked`, `Packed`, `HandedOver`, `Delivered`, `Returned`, `Closed`.

### 3.3. Batch, hạn dùng, QC phải luôn nổi

Với mỹ phẩm, các thông tin sau không được giấu sâu:

- batch/lot number;
- ngày sản xuất;
- hạn dùng;
- QC status;
- near-expiry warning;
- quarantine/hold/fail status;
- nguồn gốc nguyên vật liệu hoặc nhà máy gia công.

Nếu UI làm người dùng quên batch và QC, đó là UI nguy hiểm.

### 3.4. Warehouse-first cho các màn hình kho

Các workflow kho thực tế có nhịp: nhận đơn trong ngày, thực hiện xuất/nhập, soạn và đóng gói, tối ưu vị trí kho, kiểm kê cuối ngày, đối soát báo cáo và kết thúc ca. Vì vậy UI kho phải ưu tiên:

- thao tác bằng scanner;
- thao tác nhanh trên tablet/laptop;
- line item rõ, lớn, ít chữ thừa;
- cảnh báo thiếu hàng/nhầm SKU/ngược batch;
- màn hình đóng ca và đối soát cuối ngày.

### 3.5. Scan-first, type-second

Ở kho, bàn giao, đóng hàng, hàng hoàn:

- ưu tiên quét mã;
- nhập tay chỉ là fallback có lý do;
- mọi nhập tay thay cho scan phải ghi audit log;
- UI phải cho biết scan đúng, sai, trùng, thiếu, lệch số lượng.

### 3.6. Không để người dùng đoán bước tiếp theo

Mỗi màn hình chứng từ phải có `Next Action` rõ:

- Submit for approval;
- Approve;
- Send to QC;
- Release to Stock;
- Pick;
- Pack;
- Handover;
- Receive Return;
- Close Shift.

ERP tốt là ERP làm người dùng ít phải hỏi “giờ làm gì tiếp?”.

### 3.7. Exception-first

Quản lý không cần nhìn mọi thứ bình thường. Quản lý cần nhìn thứ lệch:

- đơn chưa đủ hàng;
- batch hold;
- QC fail;
- PO trễ;
- shipment thiếu đơn;
- hàng hoàn chưa xử lý;
- tồn lệch cuối ngày;
- gia công ngoài quá hạn;
- payment pending quá SLA.

UI phải có pattern cho exception, không chỉ list dữ liệu thô.

---

## 4. User context theo vai trò

### 4.1. CEO / Owner

**Nhu cầu UI:** nhìn nhanh tình trạng vận hành, tài chính, hàng, rủi ro.  
**Ưu tiên:** dashboard, drill-down, cảnh báo, phê duyệt lớn.  
**Không cần:** nhập liệu chi tiết.

### 4.2. COO / Operations Manager

**Nhu cầu UI:** theo dõi đơn hàng, kho, QC, gia công, bàn giao, bottleneck.  
**Ưu tiên:** board vận hành, exception center, SLA, trạng thái theo ngày.

### 4.3. Warehouse Staff

**Nhu cầu UI:** nhập/xuất/chuyển/đóng/bàn giao/hàng hoàn/kiểm kê.  
**Ưu tiên:** scan, line item rõ, ít chữ, thao tác liên tục, chống nhầm.

### 4.4. Warehouse Manager

**Nhu cầu UI:** kiểm soát phiếu, đối soát, duyệt điều chỉnh, đóng ca, tồn lệch.  
**Ưu tiên:** shift closing, stock count, discrepancy report, manifest status.

### 4.5. QA/QC

**Nhu cầu UI:** kiểm hàng đầu vào, kiểm batch thành phẩm, hold/pass/fail, ghi lỗi, release.  
**Ưu tiên:** checklist, evidence, attachment, reason code, batch trace.

### 4.6. Production / Outsourcing Coordinator

**Nhu cầu UI:** theo dõi đơn gia công, bàn giao NVL/bao bì, duyệt mẫu, tiến độ nhà máy, nhận hàng, báo lỗi trong 3–7 ngày.  
**Ưu tiên:** timeline, milestone, sample approval, material transfer, factory issue log.

### 4.7. Sales / Sales Admin

**Nhu cầu UI:** tạo đơn, check tồn khả dụng, giá, chiết khấu, trạng thái giao hàng, công nợ.  
**Ưu tiên:** order quick-create, stock availability, customer debt warning, discount approval.

### 4.8. Finance

**Nhu cầu UI:** công nợ, payment approval, đối soát COD, hóa đơn NCC, chi phí.  
**Ưu tiên:** trace từ giao dịch về chứng từ, không cho sửa dữ liệu gốc tùy tiện.

---

## 5. Nguyên tắc App Shell

### 5.1. Layout tổng thể

ERP dùng layout desktop-first, tablet-compatible.

```text
┌────────────────────────────────────────────────────────────┐
│ Top Bar: Search | Alerts | Approvals | User                │
├───────────────┬────────────────────────────────────────────┤
│ Sidebar       │ Page Header                                │
│ Navigation    │ Filters / Actions                          │
│               │ Main Content                               │
│               │ Drawer / Detail / Side Panel               │
└───────────────┴────────────────────────────────────────────┘
```

### 5.2. Sidebar

Sidebar phải chia module theo nghiệp vụ:

```text
Dashboard
Master Data
Purchase
QA/QC
Production
Warehouse
Sales
Shipping
Returns
Finance Basic
Reports
Settings
```

Sidebar phải hỗ trợ:

- collapse/expand;
- icon + label;
- group theo module;
- chỉ hiện menu theo quyền;
- badge cảnh báo nếu có việc cần xử lý.

### 5.3. Top Bar

Top bar gồm:

- global search;
- notification;
- pending approvals;
- current warehouse/branch/brand context nếu có;
- user profile;
- quick action nếu được phân quyền.

### 5.4. Breadcrumb

Mỗi trang detail phải có breadcrumb:

```text
Warehouse / Stock Transfer / TRF-20260424-001
```

Không để người dùng lạc trong hệ thống.

---

## 6. Design Tokens

Design tokens là “ngôn ngữ thị giác chung”. Frontend không tự đặt màu/font/spacing từng nơi.

### 6.1. Color tokens

#### 6.1.1. Semantic color

| Token | Ý nghĩa | Dùng cho |
|---|---|---|
| `color.primary` | hành động chính | nút chính, link chính |
| `color.success` | đạt/pass/hoàn tất | QC pass, completed |
| `color.warning` | cần chú ý | near-expiry, pending, partial |
| `color.danger` | lỗi/rủi ro/cấm | QC fail, stock shortage, destructive action |
| `color.info` | thông tin/trung lập | note, hint, processing |
| `color.neutral` | nền, text, border | layout, table, form |
| `color.locked` | dữ liệu khóa/không sửa | read-only, posted, closed |
| `color.quarantine` | hàng cách ly/hold | QC hold, quarantined stock |

#### 6.1.2. Màu theo trạng thái nghiệp vụ

| Business State | Token | Gợi ý hiển thị |
|---|---|---|
| Draft | `state.draft` | xám nhẹ |
| Submitted | `state.submitted` | xanh thông tin |
| Pending Approval | `state.pending` | vàng |
| Approved | `state.approved` | xanh lá |
| Rejected | `state.rejected` | đỏ |
| Cancelled | `state.cancelled` | xám đậm |
| Hold | `state.hold` | vàng cam |
| Pass | `state.pass` | xanh lá |
| Fail | `state.fail` | đỏ |
| Partial | `state.partial` | tím/xanh cảnh báo |
| Closed | `state.closed` | xám khóa |

Không dùng màu đơn thuần để truyền thông tin. Status phải có cả text.

### 6.2. Typography tokens

| Token | Size | Weight | Dùng cho |
|---|---:|---:|---|
| `text.display` | 28–32 | 600–700 | dashboard headline |
| `text.h1` | 24 | 600 | page title |
| `text.h2` | 20 | 600 | section title |
| `text.h3` | 16–18 | 600 | card title |
| `text.body` | 14 | 400 | nội dung chính |
| `text.small` | 12 | 400 | metadata, helper |
| `text.mono` | 13 | 400 | mã chứng từ, SKU, batch |

Mã chứng từ, SKU, batch nên dùng font mono hoặc style dễ đọc.

### 6.3. Spacing tokens

| Token | Value |
|---|---:|
| `space.1` | 4px |
| `space.2` | 8px |
| `space.3` | 12px |
| `space.4` | 16px |
| `space.5` | 20px |
| `space.6` | 24px |
| `space.8` | 32px |
| `space.10` | 40px |

### 6.4. Radius tokens

| Token | Value | Dùng cho |
|---|---:|---|
| `radius.sm` | 4px | input nhỏ |
| `radius.md` | 8px | button, card |
| `radius.lg` | 12px | modal, drawer |

### 6.5. Elevation tokens

| Token | Dùng cho |
|---|---|
| `elevation.none` | table/form thường |
| `elevation.sm` | card nhẹ |
| `elevation.md` | popover/dropdown |
| `elevation.lg` | modal/drawer |

Không lạm dụng shadow. ERP cần rõ, không cần lung linh.

---

## 7. Responsive & Device Strategy

### 7.1. Breakpoints

| Device | Width | Usage |
|---|---:|---|
| Desktop | >= 1280px | tác nghiệp chính, quản lý, finance, sales admin |
| Laptop | 1024–1279px | tác nghiệp phổ biến |
| Tablet | 768–1023px | kho, scan, QC đơn giản |
| Mobile | < 768px | xem nhanh, duyệt đơn giản, không ưu tiên thao tác phức tạp Phase 1 |

### 7.2. Desktop-first

Các màn hình nhiều dữ liệu như stock ledger, purchase order, sales order, QC checklist phải thiết kế desktop-first.

### 7.3. Tablet-compatible cho kho

Các màn hình sau phải dùng tốt trên tablet:

- nhận hàng;
- pick/pack;
- scan đơn;
- bàn giao ĐVVC;
- hàng hoàn;
- kiểm kê;
- đóng ca kho.

### 7.4. Mobile Phase 1

Mobile chưa phải kênh tác nghiệp chính. Tuy nhiên phải xem được:

- approval inbox;
- dashboard cơ bản;
- trạng thái đơn;
- trạng thái batch;
- alert.

---

## 8. Navigation Pattern

### 8.1. Global Search

Global search tìm theo:

- mã đơn hàng;
- mã PO;
- mã phiếu nhập/xuất;
- SKU;
- batch;
- khách hàng;
- nhà cung cấp;
- shipment/manifest;
- work order/subcontract order.

Search result phải nhóm theo entity:

```text
Orders
- SO-20260424-001
- SO-20260424-002

Batch
- BATCH-SERUM-260424-01

Shipment
- MANIFEST-GHN-260424-01
```

### 8.2. Recently Viewed

Nên có gần đây đã xem để user quay lại nhanh.

### 8.3. Saved Views

Các list quan trọng phải có saved views:

- `Đơn chờ pick`
- `Đơn chờ bàn giao`
- `Batch đang hold`
- `Hàng hoàn chưa xử lý`
- `PO chưa nhận đủ`
- `Tồn cận date`

---

## 9. Page Template Standard

### 9.1. List Page Template

```text
[Page Title]                         [Primary Action]
Subtitle / count / last updated

[Quick filters] [Advanced filter] [Search] [Saved view]

[Data table]
- checkbox
- key code
- status
- main attributes
- warning
- owner
- updated at
- row actions

[Pagination / Load more]
```

### 9.2. Detail Page Template

```text
Breadcrumb
[Entity Code] [Status Chip] [Warning Badges]
Main actions: Submit / Approve / Cancel / Print / Export

Summary Card
Tabs:
- Overview
- Lines
- QC / Batch / Stock
- Documents
- Approval
- Audit Log
- Comments
```

### 9.3. Create/Edit Page Template

```text
[Create Entity]
Section 1: Header information
Section 2: Lines
Section 3: Attachments
Section 4: Notes

Footer sticky:
[Save Draft] [Submit] [Cancel]
```

### 9.4. Operational Action Page Template

Dùng cho pick, pack, scan, handover, return inspection.

```text
[Task Title] [Progress]
Scan/Input zone
Expected items
Scanned/confirmed items
Mismatch alerts
Action footer
```

---

## 10. Table Standard

### 10.1. Table density

ERP phải hỗ trợ 3 density:

- Comfortable: mặc định cho quản lý;
- Compact: user nhiều dữ liệu;
- Scan/Operation: dòng lớn hơn, button lớn hơn, dùng tablet.

### 10.2. Column rule

Mỗi table phải có:

- key code ở cột đầu;
- status gần key code;
- warning gần status;
- updated timestamp;
- owner/assignee nếu là tác vụ;
- row action cuối dòng.

### 10.3. Sticky columns

Với bảng nhiều cột:

- cột chọn dòng sticky left;
- key code sticky left;
- action sticky right.

### 10.4. Sorting/filtering

Mọi list nghiệp vụ phải lọc được theo:

- status;
- date range;
- owner;
- warehouse;
- channel;
- supplier/customer;
- SKU;
- batch nếu liên quan.

### 10.5. Bulk action

Bulk action chỉ hiện khi user chọn dòng và có quyền.

Ví dụ:

- bulk approve;
- bulk print pick list;
- bulk assign carrier;
- bulk export;
- bulk mark counted.

Bulk action nguy hiểm phải có confirmation + reason.

### 10.6. Empty state

Không để bảng trống vô cảm.

Ví dụ:

```text
Chưa có đơn chờ bàn giao.
Các đơn đã đóng gói và sẵn sàng giao cho ĐVVC sẽ hiển thị tại đây.
```

### 10.7. Error state

Nếu load lỗi:

```text
Không tải được danh sách đơn hàng.
Vui lòng thử lại hoặc liên hệ admin nếu lỗi lặp lại.
[Mã lỗi: ORDER_LIST_FETCH_FAILED]
```

---

## 11. Form Standard

### 11.1. Required field

Field bắt buộc có dấu `*`, nhưng quan trọng hơn là helper text rõ.

Ví dụ:

```text
Batch No *
Nhập mã lô theo tem/COA. Không tự tạo nếu chưa được duyệt.
```

### 11.2. Validation timing

- validate format ngay khi blur;
- validate nghiệp vụ khi submit;
- validate trùng/đủ tồn bằng API trước khi xác nhận.

### 11.3. Inline error

Không chỉ hiện toast.

Sai ở field nào, báo ở field đó.

### 11.4. Sectioning

Form dài phải chia section:

- thông tin chung;
- đối tượng liên quan;
- line items;
- batch/QC;
- attachment;
- note;
- approval.

### 11.5. Dirty state

Nếu user sửa form chưa lưu mà rời trang, phải cảnh báo.

### 11.6. Read-only state

Chứng từ `Approved`, `Closed`, `Posted` không được sửa trực tiếp. UI hiển thị:

```text
Chứng từ đã khóa. Muốn thay đổi, tạo phiếu điều chỉnh hoặc yêu cầu mở khóa theo quy trình.
```

### 11.7. Reason modal

Mọi action sau phải yêu cầu reason:

- reject;
- cancel;
- adjust stock;
- override scan;
- manually close;
- QC fail;
- release from hold;
- return item marked unusable;
- modify approved document.

---

## 12. Status Chip Standard

### 12.1. Quy tắc chung

Status chip phải có:

- text rõ;
- màu semantic;
- tooltip giải thích nếu cần;
- không chỉ dùng icon/màu.

### 12.2. Purchase Request Status

```text
Draft → Submitted → Approved / Rejected → Converted to PO → Cancelled
```

### 12.3. Purchase Order Status

```text
Draft → Approved → Sent → Partially Received → Fully Received → Closed
                     ↘ Cancelled
```

### 12.4. GRN / Receiving Status

```text
Draft → Received Pending QC → QC Passed → Stocked
                           ↘ QC Failed → Returned to Supplier
```

### 12.5. QC Status

```text
Hold → Pass / Fail
```

### 12.6. Stock Item Status

```text
Available
Reserved
Allocated
In QC Hold
Quarantine
Damaged
Expired
Returned Pending Inspection
```

### 12.7. Sales Order Status

```text
Draft → Confirmed → Reserved → Picked → Packed → HandedOver → Delivered → Closed
                                               ↘ Return Pending → Returned
```

### 12.8. Shipment / Manifest Status

```text
Draft → Ready for Handover → Scanning → Complete → HandedOver
                                   ↘ Mismatch → Recheck Required
```

### 12.9. Return Status

```text
Received from Shipper → Pending Inspection → Usable / Unusable → Stocked / Sent to Lab / Written Off
```

### 12.10. Subcontract Production Status

```text
Draft → Confirmed with Factory → Deposit Paid → Materials Transferred → Sample Pending
→ Sample Approved → Mass Production → Delivered to Warehouse → QC Inspection
→ Accepted / Issue Reported → Final Payment → Closed
```

---

## 13. Component Library Standard

### 13.1. Component phân lớp

Frontend component chia 4 lớp:

```text
1. Foundation components
2. Common ERP components
3. Business components
4. Page components
```

### 13.2. Foundation components

| Component | Mục đích |
|---|---|
| Button | action |
| Input | nhập text |
| Select | chọn dữ liệu |
| DatePicker | chọn ngày |
| Checkbox | chọn dòng/option |
| Radio | chọn 1 trong nhiều |
| Textarea | note/reason |
| Tooltip | giải thích ngắn |
| Popover | lựa chọn phụ |
| Modal | xác nhận/tác vụ tập trung |
| Drawer | xem/sửa nhanh bên cạnh |
| Tabs | chia detail page |
| Badge | đếm/cảnh báo nhỏ |
| Chip | status/tag |
| Card | block thông tin |
| Table | data grid |

### 13.3. Common ERP components

| Component | Mục đích |
|---|---|
| `PageHeader` | title, status, action |
| `FilterBar` | search, quick filter, advanced filter |
| `SavedViewSelector` | view đã lưu |
| `DataTable` | bảng dữ liệu chuẩn |
| `StatusChip` | trạng thái nghiệp vụ |
| `WarningBadge` | cảnh báo |
| `EntityLookup` | tìm SKU/customer/supplier |
| `LineItemGrid` | bảng dòng hàng |
| `ApprovalTimeline` | lịch sử duyệt |
| `AuditLogPanel` | log thay đổi |
| `AttachmentPanel` | file đính kèm |
| `CommentThread` | trao đổi nội bộ |
| `ReasonModal` | nhập lý do |
| `ConfirmActionModal` | xác nhận action |
| `PrintExportButton` | in/xuất chứng từ |

### 13.4. Business components Phase 1

| Component | Module | Mục đích |
|---|---|---|
| `BatchSelector` | Inventory/QC/Sales | chọn batch theo FEFO/QC |
| `ExpiryWarningBadge` | Inventory | cảnh báo cận date |
| `QCStatusBadge` | QC/Inventory | hold/pass/fail |
| `StockAvailabilityPanel` | Sales/Warehouse | tồn vật lý, tồn khả dụng, reserved |
| `StockLedgerTimeline` | Warehouse | lịch sử biến động tồn |
| `ScanInputPanel` | Warehouse/Shipping/Returns | scan mã |
| `PickPackProgress` | Warehouse | tiến độ soạn/đóng |
| `ManifestScanBoard` | Shipping | bàn giao ĐVVC |
| `ReturnInspectionCard` | Returns/QC | kiểm hàng hoàn |
| `ShiftClosingChecklist` | Warehouse | đóng ca cuối ngày |
| `SubcontractTimeline` | Production | gia công ngoài |
| `MaterialTransferPanel` | Production/Warehouse | bàn giao NVL/bao bì cho nhà máy |
| `SampleApprovalPanel` | Production/QC/R&D | duyệt mẫu |
| `DiscrepancyPanel` | Warehouse/Shipping | lệch số lượng/thiếu hàng |

### 13.5. Page components

Page component không chứa business logic phức tạp. Nó lắp ghép:

- data fetching;
- layout;
- component business;
- action handler.

Business rule phải nằm ở backend hoặc application layer frontend rõ ràng, không rải lung tung trong JSX.

---

## 14. Icon Standard

### 14.1. Icon phải hỗ trợ text

Icon không đứng một mình cho action quan trọng.

Sai:

```text
[icon thùng rác]
```

Đúng:

```text
[Delete] hoặc [Hủy chứng từ]
```

### 14.2. Icon semantic

| Ý nghĩa | Icon gợi ý |
|---|---|
| cảnh báo | triangle alert |
| lỗi | circle x |
| thành công | check circle |
| khóa | lock |
| batch/lot | barcode/tag |
| kho | warehouse/package |
| QC | shield/checklist |
| shipping | truck |
| return | rotate arrow |
| attachment | paperclip |
| audit | clock/history |

---

## 15. Button & Action Standard

### 15.1. Button hierarchy

| Loại | Dùng cho |
|---|---|
| Primary | hành động chính duy nhất trên trang |
| Secondary | hành động phụ |
| Tertiary/Text | hành động nhẹ |
| Danger | hủy/xóa/fail/reject |
| Ghost | action trong table/card |

Mỗi màn hình nên có **một primary action chính**.

### 15.2. Action naming

Dùng động từ rõ:

- `Tạo phiếu nhập`
- `Gửi duyệt`
- `Duyệt`
- `Từ chối`
- `Khóa batch`
- `Release stock`
- `Bắt đầu scan`
- `Xác nhận bàn giao`
- `Đóng ca`

Không dùng chữ mơ hồ như `OK`, `Submit` nếu context không rõ.

### 15.3. Dangerous action

Các action nguy hiểm:

- cancel;
- reject;
- QC fail;
- stock adjustment;
- manual override;
- close shift with discrepancy;
- mark returned item unusable.

Phải có:

- confirmation modal;
- reason required;
- permission check;
- audit log.

---

## 16. Notification & Alert Standard

### 16.1. Notification type

| Type | Mục đích |
|---|---|
| Success | action hoàn tất |
| Info | thông tin trung lập |
| Warning | cần chú ý |
| Error | lỗi cần xử lý |
| Approval | yêu cầu duyệt |
| SLA Alert | quá hạn |

### 16.2. Toast

Toast dùng cho feedback ngắn:

```text
Đã lưu phiếu nhập GRN-20260424-001.
```

Không dùng toast để chứa thông tin sống còn. Thông tin sống còn phải nằm trên màn hình.

### 16.3. Persistent alert

Cảnh báo cần xử lý phải nằm persistent:

- batch hold;
- tồn không đủ;
- gần hết hạn;
- chưa QC;
- discrepancy;
- payment overdue;
- shipment mismatch.

### 16.4. Alert Center

Alert center phải gom:

- approval pending;
- stock shortage;
- QC pending/fail;
- return pending inspection;
- shipment mismatch;
- PO overdue;
- production factory issue.

---

## 17. Scan UX Standard

Scan UX là chuẩn sống còn cho kho, đóng hàng, bàn giao, hàng hoàn.

### 17.1. Scan input

Scan input phải:

- auto-focus;
- accept barcode/QR;
- clear sau khi scan thành công;
- hiển thị kết quả ngay;
- có âm thanh/visual feedback nếu dùng thiết bị hỗ trợ;
- không cần user bấm enter quá nhiều nếu scanner đã gửi newline.

### 17.2. Scan result states

| State | UI |
|---|---|
| Valid | xanh + tick + cập nhật dòng |
| Duplicate | vàng + thông báo đã scan |
| Unknown | đỏ + không tìm thấy mã |
| Wrong batch | đỏ + batch không khớp |
| Wrong order | đỏ + không thuộc manifest/đơn |
| Over quantity | đỏ + vượt số lượng |
| Partial | vàng + còn thiếu |

### 17.3. Manual override

Nếu cho nhập tay thay scan:

- yêu cầu quyền;
- bắt buộc lý do;
- ghi audit log;
- có badge `Manual Override`.

### 17.4. Scan handover board

Màn bàn giao ĐVVC cần hiển thị:

```text
Manifest: MAN-GHN-20260424-001
Carrier: GHN
Expected orders: 120
Scanned: 118
Missing: 2
Extra/Wrong: 0

[Scan Input]

Tabs:
- All
- Scanned
- Missing
- Mismatch
```

### 17.5. Bàn giao theo thùng/rổ

Theo workflow thực tế, hàng được để theo thùng/rổ có số lượng bằng nhau hoặc theo khu. UI phải hỗ trợ:

- mã thùng/rổ;
- số lượng expected/scanned;
- khu vực để hàng;
- đơn trong thùng/rổ;
- trạng thái đủ/chưa đủ;
- in label thùng/rổ nếu cần.

---

## 18. Warehouse UX Patterns

### 18.1. Daily Warehouse Workboard

Màn hình tổng của kho trong ngày:

```text
Warehouse Daily Board

Today Orders:
- Waiting Pick
- Picking
- Packed
- Ready Handover
- HandedOver

Inbound:
- Waiting Receive
- Pending QC
- Stocked

Returns:
- Waiting Receive
- Pending Inspection
- Usable
- Unusable

End of Day:
- Stock Count Pending
- Discrepancy
- Shift Close Status
```

### 18.2. Receiving UX

Nhập kho phải đi theo flow:

```text
Select PO / Delivery Document
→ Enter/scan items
→ Batch + expiry
→ Check quantity/packaging/lot
→ Send to QC
→ QC pass/fail
→ Stock available or return supplier
```

UI phải có:

- PO reference;
- supplier;
- expected vs received;
- batch/expiry;
- packaging condition;
- QC status;
- attachment COA/document.

### 18.3. Outbound UX

Xuất kho phải có:

- phiếu xuất;
- warehouse;
- reason/source document;
- line items;
- expected vs actual issued;
- picker;
- receiver;
- signature/confirmation;
- audit log.

### 18.4. Pick/Pack UX

Màn hình pick/pack:

```text
Order/SKU list
Location/Bin
Batch suggested by FEFO
Qty required
Qty picked
Scan result
Packing status
```

Nếu batch QC chưa pass hoặc cận date vượt policy, UI không cho pick hoặc yêu cầu approval.

### 18.5. End-of-Day / Shift Closing UX

Do workflow kho có kiểm kê và đối soát cuối ngày, phải có màn hình đóng ca.

Checklist đề xuất:

```text
[ ] Tất cả đơn trong ngày đã có trạng thái rõ
[ ] Không còn đơn packed nhưng chưa bàn giao quá SLA
[ ] Hàng hoàn đã đưa vào khu vực pending inspection
[ ] Phiếu nhập/xuất đã đối chiếu
[ ] Kiểm kê nhanh/cycle count hoàn tất nếu có
[ ] Discrepancy đã có lý do hoặc ticket xử lý
[ ] Báo cáo cuối ca đã gửi quản lý
```

Nếu còn lệch, user vẫn có thể đóng ca chỉ khi có quyền và lý do.

---

## 19. QA/QC UX Patterns

### 19.1. QC Inspection Page

```text
QC Inspection: QC-20260424-001 [Hold]
Related Document: GRN / Batch / Return / Production Receipt
Item/SKU/Material
Batch / Expiry
Checklist
Measured result
Decision: Pass / Fail / Hold
Attachments
Reason / Notes
```

### 19.2. Checklist UX

Checklist nên có:

- từng tiêu chí;
- expected standard;
- actual result;
- pass/fail;
- note;
- attachment/photo;
- required rule.

### 19.3. QC decision

QC quyết định phải nổi bật:

- `Pass` → release stock;
- `Fail` → quarantine/return/issue report;
- `Hold` → chưa cho dùng/bán.

### 19.4. Complaint/Return liên quan batch

Nếu hàng hoàn hoặc complaint liên quan batch, UI phải cho trace:

```text
Return Item → Order → SKU → Batch → QC Record → Supplier/Factory
```

---

## 20. Sales Order UX Patterns

### 20.1. Sales order create

Màn tạo đơn phải có:

- customer;
- channel;
- price list;
- discount;
- line items;
- stock availability;
- customer debt warning;
- delivery info;
- payment method;
- approval status nếu vượt quyền.

### 20.2. Stock availability panel

Hiển thị rõ:

```text
Physical Stock: 1,000
Reserved: 250
QC Hold: 100
Available: 650
Near Expiry: 80
```

Không dùng một chữ “Tồn kho” mơ hồ.

### 20.3. Discount warning

Nếu discount vượt ngưỡng:

```text
Chiết khấu vượt quyền của bạn. Đơn sẽ được gửi duyệt cho Sales Manager.
```

### 20.4. Order timeline

Mỗi đơn có timeline:

```text
Created → Confirmed → Reserved → Picked → Packed → HandedOver → Delivered → Closed
```

---

## 21. Shipping / Handover UX Patterns

### 21.1. Manifest list

Table manifest:

- manifest code;
- carrier;
- warehouse;
- expected order count;
- scanned count;
- mismatch count;
- status;
- handover time;
- responsible person.

### 21.2. Manifest detail

```text
Manifest Summary
Carrier
Order Count
Bins/Boxes
Scan Progress
Mismatch Panel
Signature/Confirmation
Audit Log
```

### 21.3. Missing order flow

Nếu chưa đủ đơn:

```text
Missing detected
→ Recheck barcode/order code
→ Search in packing area
→ If found: scan and continue
→ If not found: mark missing with reason
→ Notify warehouse manager
```

UI phải hỗ trợ đúng flow này, không được chỉ báo “thiếu” rồi bỏ mặc user.

### 21.4. Signature / Confirmation

Bàn giao cần ghi:

- người bàn giao;
- người/đơn vị nhận;
- thời gian;
- số lượng đơn;
- manifest;
- mismatch nếu có;
- attachment/biên nhận nếu cần.

---

## 22. Returns / Hàng hoàn UX Patterns

### 22.1. Return receiving

Hàng hoàn từ shipper phải có màn hình nhận riêng:

```text
Receive Return
Carrier/Shipper
Order Code / Tracking Code
Scan item/order
Move to Return Area
Create Return Inspection Task
```

### 22.2. Return inspection

Kiểm hàng hoàn:

- scan hàng hoàn;
- kiểm tình trạng: rách, móp, vỡ, hỏng, thiếu;
- ảnh bằng chứng;
- quyết định: còn sử dụng / không sử dụng;
- nếu còn dùng: chuyển lại kho;
- nếu không dùng: chuyển lab/quarantine/write-off theo rule.

### 22.3. Return disposition card

```text
Item: SERUM-VITC-30ML
Batch: SVC-260424-01
Condition: Box damaged, product intact
Decision: Usable
Next Action: Move to Available Stock after QC confirmation
```

### 22.4. Không nhập hàng hoàn thẳng vào available

UI không được có shortcut nhập hàng hoàn thẳng vào tồn khả dụng nếu chưa qua inspection/QC rule.

---

## 23. Production / Subcontract Manufacturing UX Patterns

Workflow thực tế có nhánh gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao hàng về kho, kiểm tra số lượng/chất lượng, báo lỗi trong 3–7 ngày nếu có, rồi thanh toán lần cuối.

### 23.1. Subcontract order timeline

```text
Draft
→ Confirmed with Factory
→ Deposit Paid
→ Materials/Packaging Transferred
→ Sample Pending
→ Sample Approved
→ Mass Production
→ Factory Delivered
→ Warehouse Received
→ QC Inspection
→ Accepted / Issue Reported
→ Final Payment
→ Closed
```

### 23.2. Material transfer UX

Khi chuyển NVL/bao bì qua nhà máy:

- material/SKU;
- batch;
- quantity;
- packaging/spec;
- documents: COA, MSDS, tem phụ, hóa đơn VAT nếu cần;
- handover confirmation;
- expected return/production output.

### 23.3. Sample approval UX

Sample approval phải có:

- sample code;
- formula/spec version;
- photos/attachments;
- QA/R&D/Brand decision;
- approved sample storage status;
- note thay đổi;
- không cho mass production nếu sample chưa approved.

### 23.4. Factory issue report

Nếu hàng không nhận hoặc có lỗi:

```text
Issue Type
Affected Qty
Photos/Evidence
Reported Date
Factory Response Due Date
Resolution
CAPA/Note
```

UI phải có deadline 3–7 ngày theo policy nội bộ.

---

## 24. Approval UX Standard

### 24.1. Approval inbox

Approval inbox phải gom tất cả yêu cầu duyệt:

- PR/PO;
- stock adjustment;
- discount over limit;
- QC release from hold;
- sample approval;
- payment request;
- manual override;
- close shift with discrepancy.

### 24.2. Approval detail

Approver phải thấy:

- ai tạo;
- tạo lúc nào;
- lý do;
- số tiền/số lượng/rủi ro;
- dữ liệu trước/sau nếu là thay đổi;
- attachment;
- lịch sử duyệt;
- action approve/reject/request info.

### 24.3. Reject UX

Reject bắt buộc lý do.

```text
Từ chối yêu cầu này?
Lý do từ chối *
[Textarea]
```

### 24.4. Approval status on document

Chứng từ detail phải có block approval timeline, không để user đi tìm ở nơi khác.

---

## 25. Audit Log UX Standard

### 25.1. Audit log panel

Mỗi chứng từ quan trọng phải có tab/block audit:

```text
Time | User | Action | Field | Before | After | Reason
```

### 25.2. Field display

Không hiển thị JSON thô cho business user. Phải hiển thị ngôn ngữ dễ hiểu.

Ví dụ:

Sai:

```text
qc_status: HOLD → PASS
```

Đúng:

```text
QC Status: Hold → Pass
```

### 25.3. Sensitive data

Audit log có dữ liệu nhạy cảm như giá vốn, công nợ, payroll về sau phải ẩn theo quyền.

---

## 26. Attachment UX Standard

### 26.1. Attachment categories

File đính kèm phải có loại:

- COA;
- MSDS;
- delivery note;
- invoice;
- signed handover;
- product photo;
- QC evidence;
- factory agreement;
- sample photo;
- claim/issue evidence.

### 26.2. Required attachments

Một số workflow yêu cầu attachment bắt buộc:

- QC fail: ảnh/bằng chứng;
- factory issue: ảnh/bằng chứng;
- manual stock adjustment: lý do + file nếu có;
- supplier return: biên bản;
- material transfer to factory: biên bản bàn giao.

### 26.3. Preview

UI nên hỗ trợ preview ảnh/PDF trong drawer/modal.

---

## 27. Microcopy Standard

### 27.1. Ngôn ngữ

Dùng tiếng Việt rõ, ngắn, không thuật ngữ kế toán/kỹ thuật quá nặng nếu không cần.

### 27.2. Error message

Error tốt gồm:

- chuyện gì sai;
- vì sao;
- user làm gì tiếp;
- mã lỗi nếu cần.

Ví dụ:

```text
Không thể xác nhận bàn giao.
Manifest còn thiếu 2 đơn chưa scan. Vui lòng kiểm tra lại khu vực đóng hàng hoặc đánh dấu thiếu đơn có lý do.
```

### 27.3. Confirmation copy

Action nguy hiểm phải nói hậu quả.

```text
Xác nhận set batch này thành QC Fail?
Batch sẽ không thể dùng cho sản xuất hoặc bán hàng. Hành động này sẽ được ghi audit log.
```

### 27.4. Empty state copy

Nên hướng dẫn bước tiếp.

```text
Chưa có hàng hoàn cần kiểm tra.
Khi kho nhận hàng hoàn từ shipper, các phiếu pending inspection sẽ xuất hiện tại đây.
```

---

## 28. Loading / Empty / Error / Offline State

### 28.1. Loading

- Table: skeleton rows;
- Detail: skeleton cards;
- Action: button loading + disable double click;
- Scan: không khóa toàn màn hình nếu chỉ validate từng mã.

### 28.2. Empty

Empty state phải có context và action.

### 28.3. Error

Error state phải phân biệt:

- network error;
- permission error;
- validation error;
- business rule error;
- server error.

### 28.4. Offline/unstable connection

Nếu kho dùng mạng không ổn, UI phải:

- cảnh báo mất kết nối;
- không cho confirm action quan trọng khi chưa sync;
- tránh tạo ảo tưởng đã lưu;
- không silently fail.

---

## 29. Accessibility Standard

### 29.1. Keyboard

Các màn hình nhập liệu nhiều phải keyboard-friendly:

- tab order đúng;
- enter/scan không làm mất dữ liệu;
- shortcut có hướng dẫn;
- focus state rõ.

### 29.2. Contrast

Status và text phải đủ tương phản. Không dùng màu nhạt khó đọc.

### 29.3. Screen reader basics

Không cần full advanced ngay Phase 1, nhưng input/button/icon phải có label rõ.

### 29.4. Touch target

Tablet/kho:

- button tối thiểu 40px;
- scan action lớn;
- checkbox/dòng dễ chạm.

---

## 30. RBAC & Security UX

### 30.1. Hidden vs Disabled

Nếu user hoàn toàn không có quyền với chức năng: ẩn.  
Nếu user thấy được nhưng không đủ điều kiện thao tác: disable + tooltip lý do.

Ví dụ:

```text
Bạn không thể release batch này vì QC checklist chưa hoàn tất.
```

### 30.2. Cost/finance visibility

Các field như giá vốn, cost, margin, payment status phải ẩn theo role.

### 30.3. Sensitive action

Action nhạy cảm phải có:

- permission check;
- confirmation;
- reason;
- audit.

### 30.4. Session timeout

Nếu session hết hạn trong lúc nhập form, cần cảnh báo và cố gắng giữ draft local nếu được.

---

## 31. Performance UX

### 31.1. Table lớn

Với table lớn:

- server-side pagination;
- server-side filter/sort;
- không load toàn bộ dữ liệu;
- virtualization nếu cần.

### 31.2. Search debounce

Search input debounce 300–500ms.

### 31.3. Export async

Report/export lớn phải chạy async:

```text
Đang tạo file export. Bạn có thể tiếp tục làm việc, hệ thống sẽ thông báo khi file sẵn sàng.
```

### 31.4. Optimistic UI

Không dùng optimistic update cho giao dịch quan trọng như stock movement, QC release, handover, payment. Phải chờ backend xác nhận.

---

## 32. Data Display Standard

### 32.1. Date/time

Hiển thị theo timezone công ty. Format đề xuất:

```text
24/04/2026 15:30
```

### 32.2. Number

Số lượng:

```text
1,250
```

Tiền:

```text
1,250,000 ₫
```

### 32.3. Unit of measure

Luôn hiển thị UOM:

```text
500 chai
20 thùng
3 kg
```

### 32.4. Code display

Mã chứng từ/SKU/batch nên copy được bằng một click.

### 32.5. Relative + absolute time

Với tác vụ cần SLA:

```text
Quá hạn 2 giờ
Due: 24/04/2026 17:00
```

---

## 33. Dashboard UI Standard

### 33.1. Dashboard không chỉ là biểu đồ

Dashboard phải có action/drill-down.

Ví dụ card `Batch Hold: 8` phải click ra danh sách batch hold.

### 33.2. Card standard

Card KPI gồm:

- title;
- value;
- trend nếu có;
- warning nếu bất thường;
- click-through;
- last updated.

### 33.3. Operational dashboard

Phase 1 nên có:

- warehouse daily status;
- QC pending/hold/fail;
- order fulfillment;
- shipment handover;
- returns pending;
- subcontract production status;
- stock near-expiry;
- stock discrepancy.

### 33.4. Không lạm dụng biểu đồ

ERP vận hành cần nhiều bảng/cảnh báo hơn chart trang trí.

---

## 34. Design Pattern theo module Phase 1

### 34.1. Master Data

Pattern:

- list + detail drawer;
- create/edit form;
- active/inactive;
- duplicate warning;
- audit log;
- approval nếu là master data nhạy cảm.

Màn hình master data cần có:

- SKU/material;
- supplier;
- customer;
- warehouse/location;
- UOM;
- price list nếu Phase 1 có;
- QC template.

### 34.2. Procurement

Pattern:

- PR/PO list;
- PO detail;
- receiving action;
- supplier score mini-card;
- expected vs received;
- pending QC.

### 34.3. QA/QC

Pattern:

- inspection queue;
- checklist;
- evidence;
- decision modal;
- status badge;
- release/hold control.

### 34.4. Production / Subcontract

Pattern:

- subcontract order timeline;
- material transfer;
- sample approval;
- factory issue;
- receiving and QC.

### 34.5. Warehouse

Pattern:

- daily board;
- stock list;
- stock ledger;
- receiving;
- issue/transfer;
- pick/pack;
- cycle count;
- shift close.

### 34.6. Sales

Pattern:

- order list;
- order create/detail;
- stock availability;
- price/discount validation;
- order timeline.

### 34.7. Shipping

Pattern:

- manifest list;
- scan board;
- mismatch handling;
- handover confirmation.

### 34.8. Returns

Pattern:

- return queue;
- return receive;
- inspection card;
- disposition;
- trace to order/batch.

---

## 35. Frontend Component Naming Standard

### 35.1. Naming

Component PascalCase:

```text
BatchSelector
StockAvailabilityPanel
ManifestScanBoard
```

Hook camelCase:

```text
useManifestScan
useStockAvailability
useApprovalActions
```

Page route naming:

```text
warehouse/stock-ledger
shipping/manifests/[id]
returns/inspection/[id]
```

### 35.2. Component folder

Gợi ý:

```text
src/
  components/
    foundation/
    erp/
    business/
  modules/
    warehouse/
      pages/
      components/
      hooks/
      types/
    shipping/
    returns/
    production/
```

### 35.3. Không trộn component và business rule

Component UI không tự quyết business rule như “batch này có được bán không”. Nó nhận prop từ backend/application layer:

```text
canRelease
canPick
canHandover
warnings
```

---

## 36. UI State Machine Awareness

UI phải phản ánh state machine backend.

Ví dụ đơn hàng:

| State | Allowed Actions |
|---|---|
| Draft | edit, submit, cancel |
| Confirmed | reserve, cancel |
| Reserved | pick, release reservation |
| Picked | pack |
| Packed | add to manifest, handover |
| HandedOver | track delivery |
| Delivered | close/return |
| Returned | inspect return |
| Closed | view only |

Nếu backend không cho action, UI không nên hiện action đó như bình thường.

---

## 37. Validation UX by Business Rule

### 37.1. Batch missing

```text
Vui lòng chọn batch trước khi xuất kho.
```

### 37.2. Expiry risk

```text
Batch này còn 25 ngày hết hạn, không đạt policy xuất bán. Vui lòng chọn batch khác hoặc xin duyệt.
```

### 37.3. QC hold

```text
Không thể xuất batch đang QC Hold.
```

### 37.4. Insufficient stock

```text
Tồn khả dụng không đủ.
Yêu cầu: 100 chai
Khả dụng: 72 chai
Đang giữ: 20 chai
QC Hold: 8 chai
```

### 37.5. Manifest mismatch

```text
Đơn SO-20260424-001 không thuộc manifest này.
Vui lòng kiểm tra lại khu vực đóng hàng hoặc chọn đúng manifest.
```

### 37.6. Return not inspected

```text
Hàng hoàn chưa được kiểm tra tình trạng. Không thể nhập lại kho khả dụng.
```

---

## 38. Print / Export UX Standard

### 38.1. Print documents

Các chứng từ cần in:

- PO;
- GRN;
- stock issue;
- stock transfer;
- pick list;
- packing list;
- handover manifest;
- return receipt;
- material transfer to factory;
- QC report.

### 38.2. Print preview

Trước khi in phải có preview.

### 38.3. Export permission

Export dữ liệu nhạy cảm phải theo quyền.

### 38.4. Export audit

Export lớn hoặc chứa dữ liệu nhạy cảm nên ghi audit log.

---

## 39. UX cho đối soát cuối ngày

### 39.1. Shift close dashboard

```text
Shift: 24/04/2026 - Day Shift
Warehouse: Main Warehouse

Orders received: 500
Picked: 480
Packed: 470
HandedOver: 460
Pending: 40

Inbound completed: 3
QC pending: 2
Returns received: 20
Returns inspected: 12
Stock discrepancies: 1

[Close Shift]
```

### 39.2. Discrepancy handling

Nếu có discrepancy:

```text
Có 1 lệch tồn chưa xử lý.
Bạn có thể:
- mở ticket kiểm tra;
- tạo yêu cầu điều chỉnh;
- đóng ca với lý do nếu có quyền.
```

### 39.3. Manager sign-off

Đóng ca có lệch phải cần manager sign-off hoặc ít nhất manager notification.

---

## 40. UX cho vị trí kho / tối ưu vị trí

Do workflow có bước sắp xếp/tối ưu vị trí kho, UI kho nên có:

- location/bin;
- item current location;
- suggested location;
- move action;
- capacity warning nếu có;
- near-expiry grouping;
- fast-moving SKU zone.

Phase 1 có thể làm đơn giản:

```text
SKU | Batch | Current Location | Suggested Location | Qty | Action
```

---

## 41. Design Handoff Standard

Mỗi màn hình handoff cho dev phải có:

1. Screen name/code;
2. User role;
3. Purpose;
4. Entry point;
5. Main layout;
6. Components used;
7. Data fields;
8. Actions;
9. Status rules;
10. Validation rules;
11. Empty/loading/error states;
12. Permission behavior;
13. API dependency;
14. Audit/log behavior nếu có;
15. Open questions.

Không handoff kiểu chỉ có hình Figma mà thiếu rule.

---

## 42. Design QA Checklist

Trước khi approve UI cho một màn hình, kiểm:

- màn hình có đúng role không;
- primary action rõ không;
- status có dễ nhìn không;
- batch/QC/expiry có hiện đúng nơi cần không;
- form có required/helper/error không;
- table có filter/sort/search không;
- permission state có rõ không;
- dangerous action có confirm/reason không;
- empty/loading/error state có chưa;
- audit/attachment/approval block có chưa nếu cần;
- tablet/desktop có dùng được không;
- copy tiếng Việt dễ hiểu không;
- có bấm drill-down được không;
- có chống thao tác sai không.

---

## 43. Frontend QA Checklist

Trước khi merge frontend task:

- đúng design token;
- không hard-code màu/spacing tùy tiện;
- component tái sử dụng đúng;
- không duplicate table/form pattern;
- validation hiển thị inline;
- loading/empty/error đầy đủ;
- permission state đúng;
- responsive cơ bản;
- keyboard/tab order ổn;
- API error render đúng format;
- không lộ dữ liệu cost/finance nếu không có quyền;
- action nguy hiểm không thiếu confirmation;
- scan input auto-focus nếu là màn scan.

---

## 44. Definition of Done cho UI/UX task

Một UI/UX task chỉ được coi là xong khi:

1. Có screen/wireframe hoặc design rõ.
2. Có component mapping.
3. Có field list.
4. Có trạng thái và action rule.
5. Có validation/error/empty/loading state.
6. Có permission behavior.
7. Có responsive note.
8. Có copy tiếng Việt chính.
9. Có handoff note cho dev.
10. Được BA/Product Owner xác nhận.

---

## 45. Definition of Done cho Frontend implementation

Một màn hình frontend chỉ được coi là xong khi:

1. Implement đúng design system.
2. API integration đúng.
3. Loading/empty/error đầy đủ.
4. Validation đầy đủ.
5. Permission state đúng.
6. Action dangerous có confirm/reason.
7. Status chip đúng mapping.
8. Table/form đúng chuẩn.
9. Audit/attachment/approval block có nếu scope yêu cầu.
10. Responsive tối thiểu desktop/laptop/tablet nếu màn kho.
11. Pass UAT scenario liên quan.
12. Không có console error.
13. Không hard-code dữ liệu giả.

---

## 46. Những lỗi UI/UX cần cấm

### 46.1. Cấm dùng “Tồn kho” một nghĩa mơ hồ

Phải phân biệt:

- physical stock;
- reserved stock;
- QC hold;
- available stock;
- damaged/returned/quarantine.

### 46.2. Cấm ẩn batch/hạn dùng trong màn xuất bán/xuất kho

Batch/hạn dùng phải hiện ở line item hoặc selector.

### 46.3. Cấm action nguy hiểm không có lý do

Reject, cancel, fail, adjust phải có reason.

### 46.4. Cấm chỉ dùng toast cho lỗi lớn

Lỗi lớn phải persistent.

### 46.5. Cấm nút “Xóa” với chứng từ nghiệp vụ đã phát sinh

Dùng cancel/reverse/adjust, không xóa cứng.

### 46.6. Cấm tạo workflow scan mà không có mismatch handling

Scan phải xử lý sai, trùng, thiếu, thừa.

### 46.7. Cấm đưa hàng hoàn vào available nếu chưa inspection

Hàng hoàn phải qua inspection/disposition.

### 46.8. Cấm mass production nếu sample chưa approved

UI phải khóa bước này trong subcontract production.

---

## 47. MVP UI/UX Phase 1 bắt buộc

Nếu Phase 1 cần tối giản, những chuẩn sau vẫn bắt buộc:

1. App shell chuẩn.
2. Sidebar theo module.
3. List/detail template.
4. Table/filter/search standard.
5. Form/validation standard.
6. Status chip standard.
7. Batch/QC/expiry display.
8. Scan input cho shipping/warehouse/returns.
9. Approval timeline.
10. Audit log panel.
11. Attachment panel.
12. Dangerous action modal.
13. Shift closing UI.
14. Return inspection UI.
15. Subcontract production timeline.

---

## 48. Suggested UI Roadmap

### 48.1. Sprint UI Foundation

- design tokens;
- app shell;
- table;
- form;
- status chip;
- modal/drawer;
- attachment/audit/approval block.

### 48.2. Sprint Warehouse Core

- warehouse board;
- receiving;
- issue/transfer;
- stock ledger;
- shift closing.

### 48.3. Sprint Sales + Fulfillment

- sales order;
- stock availability;
- pick/pack;
- manifest handover.

### 48.4. Sprint QC + Returns

- QC queue/checklist;
- return receiving;
- return inspection;
- disposition.

### 48.5. Sprint Production Subcontract

- subcontract order;
- material transfer;
- sample approval;
- factory issue report.

---

## 49. Open Questions cần chốt

1. ERP có cần đa ngôn ngữ ngay Phase 1 không, hay chỉ tiếng Việt?
2. Có dùng thiết bị scan barcode/QR loại nào?
3. Kho dùng tablet, laptop hay máy tính để bàn là chính?
4. Có in tem/thùng/rổ/manifest ngay trong hệ thống không?
5. Có cần chữ ký điện tử khi bàn giao ĐVVC không?
6. Return inspection có cần chụp ảnh bằng camera thiết bị không?
7. Có cần offline mode cho kho không?
8. Mức near-expiry policy theo từng SKU là bao nhiêu ngày?
9. Có bao nhiêu kho/khu vực/bin cần quản lý trong Phase 1?
10. Sample approval thuộc R&D, QA hay Brand duyệt cuối?

---

## 50. Kết luận

Chuẩn UI/UX này không nhằm làm hệ thống đẹp cho vui. Nó nhằm làm ERP:

- dễ dùng;
- khó nhập sai;
- nhìn ra trạng thái ngay;
- bám workflow kho và sản xuất thực tế;
- bảo vệ batch/QC/hạn dùng;
- xử lý tốt bàn giao ĐVVC, hàng hoàn, gia công ngoài;
- hỗ trợ audit, approval, attachment;
- giúp frontend build đồng bộ, không mỗi màn một kiểu.

Câu chốt:

**ERP mỹ phẩm không được thiết kế như phần mềm quản lý hàng hóa chung chung. Nó phải làm nổi bật batch, hạn dùng, QC, kho, hàng hoàn, bàn giao, gia công ngoài và đối soát cuối ngày. Đó là nơi tiền và rủi ro thật nằm.**

---

## 51. Phụ lục A — Component Inventory Phase 1

| Component | Priority | Owner | Notes |
|---|---|---|---|
| AppShell | P0 | Frontend | sidebar/topbar/layout |
| PageHeader | P0 | Frontend | title/action/status |
| DataTable | P0 | Frontend | filter/sort/bulk/pagination |
| FilterBar | P0 | Frontend | quick/advanced filter |
| StatusChip | P0 | Frontend | semantic mapping |
| WarningBadge | P0 | Frontend | exception visibility |
| FormSection | P0 | Frontend | form structure |
| LineItemGrid | P0 | Frontend | order/PO/stock lines |
| EntityLookup | P0 | Frontend | SKU/customer/supplier |
| AttachmentPanel | P0 | Frontend | file management |
| AuditLogPanel | P0 | Frontend | traceability |
| ApprovalTimeline | P0 | Frontend | approval visibility |
| ConfirmActionModal | P0 | Frontend | dangerous action |
| ReasonModal | P0 | Frontend | required reason |
| BatchSelector | P0 | Frontend + Backend | FEFO/QC rules |
| StockAvailabilityPanel | P0 | Frontend + Backend | physical/reserved/available |
| QCStatusBadge | P0 | Frontend | hold/pass/fail |
| ScanInputPanel | P0 | Frontend | warehouse/shipping/returns |
| ManifestScanBoard | P0 | Frontend | handover carrier |
| ReturnInspectionCard | P0 | Frontend | hàng hoàn |
| ShiftClosingChecklist | P0 | Frontend | đóng ca kho |
| SubcontractTimeline | P1 | Frontend | gia công ngoài |
| MaterialTransferPanel | P1 | Frontend | chuyển NVL/bao bì |
| SampleApprovalPanel | P1 | Frontend | duyệt mẫu |
| DiscrepancyPanel | P0 | Frontend | lệch kho/thiếu đơn |

---

## 52. Phụ lục B — UI Copy mẫu

### 52.1. Stock shortage

```text
Tồn khả dụng không đủ để giữ hàng.
Yêu cầu: {requested_qty} {uom}
Khả dụng: {available_qty} {uom}
Đang giữ: {reserved_qty} {uom}
QC Hold: {qc_hold_qty} {uom}
```

### 52.2. QC Hold

```text
Batch này đang ở trạng thái QC Hold.
Không thể xuất bán hoặc dùng cho sản xuất cho đến khi QA/QC release.
```

### 52.3. Manifest missing

```text
Manifest chưa đủ đơn.
Còn thiếu {missing_count} đơn. Vui lòng kiểm tra lại mã đơn và khu vực đóng hàng.
```

### 52.4. Return pending inspection

```text
Hàng hoàn đã nhận nhưng chưa kiểm tra tình trạng.
Vui lòng hoàn tất inspection trước khi nhập lại kho.
```

### 52.5. Sample not approved

```text
Mẫu chưa được duyệt.
Không thể chuyển sang sản xuất hàng loạt cho đến khi mẫu được phê duyệt.
```

---

## 53. Phụ lục C — Screen-specific UX Guardrails

### 53.1. Receiving Screen

- Không cho submit nếu thiếu batch/expiry với item bắt buộc.
- Không cho stock available nếu chưa QC pass.
- Nếu received qty khác expected qty, yêu cầu reason.

### 53.2. QC Screen

- Fail/Hold yêu cầu reason.
- Fail yêu cầu evidence nếu policy bắt buộc.
- Pass phải ghi user/time.

### 53.3. Sales Order Screen

- Không reserve nếu available stock không đủ.
- Discount over limit gửi approval.
- Customer overdue debt hiển thị warning.

### 53.4. Pick/Pack Screen

- Scan sai SKU/batch báo đỏ ngay.
- Over-pick không cho tiếp tục.
- Packed nhưng chưa handover quá SLA phải alert.

### 53.5. Handover Screen

- Không handover nếu manifest missing mà chưa có reason/approval.
- Lưu danh sách scanned/missing/mismatch.
- In/export biên bản bàn giao.

### 53.6. Return Screen

- Return phải qua pending inspection.
- Usable/unusable phải ghi condition.
- Unusable cần lý do và location xử lý.

### 53.7. Subcontract Screen

- Không mass production nếu sample chưa approved.
- Materials transfer phải track batch/qty.
- Issue report phải có due date cho nhà máy phản hồi.

---

## 54. Phụ lục D — Sign-off

| Role | Name | Sign-off Date | Notes |
|---|---|---|---|
| CEO / Product Owner |  |  |  |
| COO / Operations |  |  |  |
| Warehouse Manager |  |  |  |
| QA/QC Lead |  |  |  |
| Sales Lead |  |  |  |
| Production/Outsource Lead |  |  |  |
| Finance Lead |  |  |  |
| UI/UX Lead |  |  |  |
| Frontend Lead |  |  |  |
| Backend Lead |  |  |  |
| QA/Test Lead |  |  |  |

