# 08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1

**Project:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** Screen List + Low-Fidelity Wireframe  
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

---

## 1. Mục tiêu tài liệu

Tài liệu này là cầu nối giữa **PRD/SRS** và **thiết kế UI/UX + build thực tế**.  
Nó trả lời 5 câu hỏi rất thực dụng:

1. Phase 1 có **những màn hình nào**.
2. Mỗi màn hình phục vụ **vai trò nào**.
3. Mỗi màn hình có **bố cục, block thông tin, hành động chính** ra sao.
4. Màn hình nào cần wireframe chi tiết ngay, màn hình nào có thể kế thừa pattern.
5. Đội BA, UI/UX, dev, tester sẽ **handoff theo cấu trúc nào** để tránh mơ hồ.

Nói ngắn gọn:

- **PRD/SRS** nói hệ thống phải làm gì.
- **Process Flow** nói doanh nghiệp sẽ chạy thế nào.
- **Screen List + Wireframe** nói người dùng sẽ bấm ở đâu, nhìn gì, thao tác ra sao.

---

## 2. Phạm vi của tài liệu

### 2.1. In scope
Phase 1 bao gồm các nhóm màn hình cho 6 module lõi và phần nền hệ thống:

1. Common / System
2. Master Data
3. Procurement
4. QA/QC
5. Production
6. Warehouse
7. Sales

### 2.2. Out of scope
Tài liệu này chưa đi sâu vào:
- visual design / brand guideline pixel-perfect;
- design system final;
- prototype tương tác cao;
- mobile app riêng;
- CRM nâng cao, HRM, KOL/Affiliate, POS hoàn chỉnh, kế toán sâu.

### 2.3. Mức độ wireframe trong tài liệu
Tài liệu chia màn hình thành 3 cấp:

- **L1 – Full wireframe:** cần wireframe chi tiết ngay vì là màn hình lõi của nghiệp vụ.
- **L2 – Derived pattern:** kế thừa từ pattern đã có, chỉ cần chỉnh field/logic.
- **L3 – Config / utility:** ưu tiên sau, có thể theo template admin chuẩn.

---

## 3. Nguyên tắc thiết kế giao diện ERP cho Phase 1

### 3.1. Thiết kế ưu tiên tốc độ thao tác hơn hình thức
ERP là công cụ vận hành. Người dùng mở ra để làm việc, không phải để ngắm.  
Do đó cần ưu tiên:
- ít click dư thừa;
- lọc và tìm nhanh;
- thao tác hàng loạt có kiểm soát;
- nhìn thấy trạng thái ngay;
- bấm vào số tổng để drill down xuống chứng từ gốc.

### 3.2. List-first, detail-second
Hầu hết tác nghiệp ERP bắt đầu từ:
- tìm chứng từ,
- lọc đúng batch / đúng trạng thái,
- rồi mới mở detail.

Vì vậy các màn hình list phải đủ mạnh:
- search,
- filter,
- save view,
- export,
- quick action,
- badge trạng thái.

### 3.3. Batch / expiry / QC status phải luôn nổi
Với ngành mỹ phẩm, 3 thông tin này là mạch sống:
- batch/lô,
- hạn dùng,
- trạng thái QC.

Bất kỳ màn hình nào liên quan đến hàng hóa, tồn kho, xuất nhập, sản xuất, trả hàng đều phải ưu tiên hiển thị 3 thông tin trên.

### 3.4. Không nhồi mọi thứ vào 1 form
Form dài và nhiều field là nguyên nhân gây lỗi nhập liệu.  
Mỗi form nên chia rõ theo:
- header,
- line items,
- attachments,
- approval,
- audit,
- tabs nghiệp vụ.

### 3.5. Mỗi vai trò nhìn giao diện khác nhau
Không để mọi người vào cùng một “siêu admin panel”.  
Cần có portal/landing phù hợp cho:
- Purchasing
- QC/QA
- Production
- Warehouse
- Sales
- COO/CEO

### 3.6. Approval và audit phải ở ngay nơi thao tác
Người duyệt không nên phải đi vòng nhiều màn hình.  
Mọi chứng từ quan trọng đều cần thấy nhanh:
- trạng thái,
- người tạo,
- SLA,
- lý do phê duyệt,
- lịch sử thay đổi,
- file đính kèm.

### 3.7. Thiết kế theo pattern tái sử dụng
Các pattern chính của Phase 1:
- Dashboard pattern
- List pattern
- Detail pattern
- Form with line items pattern
- Approval drawer pattern
- Batch trace pattern
- Stock grid pattern

Làm tốt 6 pattern này sẽ giúp giảm rất mạnh effort UI và dev.

---

## 4. Vai trò người dùng và portal đề xuất

### 4.1. CEO / COO portal
Mục tiêu:
- nhìn KPI,
- mở inbox duyệt,
- drill down từ số tổng sang giao dịch.

### 4.2. Master Data portal
Mục tiêu:
- tạo/sửa item, BOM, supplier, customer, warehouse, price list.

### 4.3. Purchasing portal
Mục tiêu:
- tạo PR/PO,
- theo dõi mở đơn mua,
- theo dõi giao hàng và hàng chờ QC.

### 4.4. QC/QA portal
Mục tiêu:
- xử lý inbound inspection,
- xử lý FG inspection,
- quyết định Pass/Hold/Fail,
- release batch.

### 4.5. Production portal
Mục tiêu:
- tạo và theo dõi Production Order,
- issue nguyên liệu,
- ghi nhận output và hao hụt.

### 4.6. Warehouse portal
Mục tiêu:
- nhập/xuất/chuyển/kiểm kê,
- nhìn tồn theo batch/hạn dùng,
- trace batch nhanh.

### 4.7. Sales portal
Mục tiêu:
- tạo quotation/SO,
- check tồn khả dụng,
- tạo delivery,
- xử lý return.

---

## 5. Kiến trúc điều hướng tổng thể

### 5.1. Menu trái cấp 1

```text
Tổng quan
Phê duyệt
Dữ liệu gốc
Mua hàng
QA / QC
Sản xuất
Kho hàng
Bán hàng
Báo cáo
Cài đặt
```

### 5.2. Menu con cấp 2 theo module

#### Dữ liệu gốc
```text
Item
BOM
Supplier
Customer
Warehouse
UOM
Price List
Batch Rules
```

#### Mua hàng
```text
Purchase Requisition
Purchase Order
Receiving
Supplier Return
Purchase Dashboard
```

#### QA / QC
```text
Inbound Inspection
FG Inspection
QC Decision / Release
NCR
QC Dashboard
```

#### Sản xuất
```text
Production Order
Material Issue
Production Confirmation
Output & Scrap
Production Dashboard
```

#### Kho hàng
```text
Goods Receipt
Goods Issue
Transfer Order
Transfer Receipt
Stock Count
Stock Adjustment
Stock Ledger
Stock by Batch / Expiry
Batch Trace
```

#### Bán hàng
```text
Quotation
Sales Order
SO Approval
Delivery Order
Sales Return
Sales Dashboard
```

### 5.3. Điều hướng chéo giữa module
Hệ thống phải hỗ trợ bấm chéo nhanh:
- từ PO sang Receiving;
- từ Receiving sang Inbound Inspection;
- từ QC Pass sang Stock by Batch;
- từ Production Order sang Material Issue và FG Inspection;
- từ Sales Order sang Delivery Order;
- từ Sales Return sang Batch Trace;
- từ Batch Trace truy ngược về Receiving / QC / Production / SO.

---

## 6. Layout chuẩn toàn hệ thống

### 6.1. App shell chuẩn

```text
+--------------------------------------------------------------------------------------------------+
| Top bar: Global search | Company/branch | Notifications | Quick create | User menu              |
+----------------------+--------------------------------------------------------------------------+
| Left nav             | Breadcrumbs                                                               |
| Module menu          | Page title                                    Primary actions              |
|                      | Secondary tabs / saved views / quick filters                               |
|                      |--------------------------------------------------------------------------|
|                      | Main content area                                                        |
|                      |                                                                          |
|                      |                                                                          |
+----------------------+--------------------------------------------------------------------------+
| Optional right drawer: Summary / Approval / Activity / Attachments / Audit                       |
+--------------------------------------------------------------------------------------------------+
```

### 6.2. Pattern cho màn hình list
Mọi màn hình list nên có:

- page header + actions chính;
- search box;
- filter chips;
- advanced filter drawer;
- saved views;
- table/grid;
- bulk select + bulk action;
- pagination;
- export;
- column chooser;
- row click mở detail drawer hoặc detail page.

### 6.3. Pattern cho màn hình detail
Mọi màn hình detail nên có:

- header chứa số chứng từ, trạng thái, người tạo, ngày tạo;
- action bar theo quyền;
- summary strip;
- tabs: General / Lines / QC / Attachments / Approval / Audit;
- right rail: activity, notes, SLA, related docs.

### 6.4. Pattern cho form có line items
Các form như PO, SO, Material Issue nên chia:

1. Header section  
2. Business info section  
3. Line items grid  
4. Totals / summary  
5. Attachments / notes  
6. Validation + approval status

### 6.5. Breakpoints đề xuất
- **Desktop chuẩn:** 1440px  
- **Laptop vận hành:** 1280px  
- **Tablet kho/xưởng:** 1024px  
- **Mobile:** chỉ hỗ trợ approvals, notifications, tra cứu nhẹ ở Phase 1

---

## 7. Component và quy ước hiển thị chung

### 7.1. Status chip
Mọi trạng thái phải hiển thị bằng chip/badge, không chỉ bằng text thuần.

Nhóm trạng thái chính:
- Draft
- Submitted
- Pending Approval
- Approved
- Partially Received / Partially Issued / Partially Delivered
- Closed
- Cancelled
- Hold
- Pass
- Fail
- Blocked / Overridden

### 7.2. Mandatory field
- hiển thị dấu `*`;
- nếu chưa nhập phải báo lỗi tại chỗ, không chỉ báo cuối form.

### 7.3. Read-only field
Các field hệ thống sinh hoặc bị khóa theo rule phải hiển thị khác biệt rõ, ví dụ:
- document number,
- created by,
- created time,
- QC decision đã khóa,
- available stock tính toán.

### 7.4. Drill-down numbers
Các số KPI trên dashboard/list phải có khả năng click xuống:
- danh sách chứng từ,
- batch,
- line item,
- transaction history.

### 7.5. File đính kèm và audit
Mọi chứng từ chính đều có block:
- Attachments
- Comments / Notes
- Approval history
- Audit trail

---

## 8. Danh sách màn hình tổng thể (Screen Inventory)

| ID | Tên màn hình | Module | Pattern | Priority | Wireframe Level | Người dùng chính |
|---|---|---|---|---|---|---|
| COM-01 | Login | Common | Auth | P1 | L2 | All |
| COM-02 | Approval Inbox | Common | List + Drawer | P1 | L1 | Manager, QA, Finance, COO, CEO |
| COM-03 | Audit Log Viewer | Common | List | P2 | L2 | Admin, Internal Control |
| COM-04 | Attachment Center | Common | Utility | P3 | L3 | All |
| COM-05 | Notification Center | Common | Feed | P2 | L2 | All |
| COM-06 | User / Role / Permission | Common | Admin Config | P2 | L3 | Admin |
| COM-07 | Approval Matrix Config | Common | Admin Config | P2 | L3 | Admin |
| COM-08 | Document Numbering Config | Common | Admin Config | P3 | L3 | Admin |
| MD-01 | Item List | Master Data | List | P1 | L1 | Master Data Admin |
| MD-02 | Item Detail / Create / Edit | Master Data | Detail + Form | P1 | L1 | Master Data Admin |
| MD-03 | Import Item | Master Data | Wizard | P2 | L2 | Master Data Admin |
| MD-04 | BOM List | Master Data | List | P1 | L2 | Master Data Admin, Planner |
| MD-05 | BOM Detail / Version | Master Data | Detail + Form | P1 | L1 | Master Data Admin, R&D, Planner |
| MD-06 | Supplier List / Detail | Master Data | List + Detail | P1 | L2 | Master Data Admin, Purchasing |
| MD-07 | Customer List / Detail | Master Data | List + Detail | P1 | L2 | Master Data Admin, Sales Admin |
| MD-08 | Warehouse List / Detail | Master Data | List + Detail | P1 | L2 | Master Data Admin, Warehouse Manager |
| MD-09 | UOM & Conversion | Master Data | Config | P2 | L3 | Master Data Admin |
| MD-10 | Price List & Discount Rule | Master Data | List + Form | P1 | L2 | Sales Admin, Master Data Admin |
| MD-11 | Item Category / Brand / Line | Master Data | Config | P2 | L3 | Master Data Admin |
| MD-12 | Batch Rule Configuration | Master Data | Config | P1 | L2 | Master Data Admin, QA |
| PUR-01 | Purchase Requisition | Procurement | List + Detail + Form | P1 | L1 | Requester, Purchasing |
| PUR-02 | Purchase Order | Procurement | List + Detail + Form | P1 | L1 | Purchasing |
| PUR-03 | Receiving against PO | Procurement | Action Form | P1 | L1 | Warehouse, Purchasing |
| PUR-04 | Supplier Return | Procurement | Action Form | P2 | L2 | Warehouse, Purchasing, QA |
| PUR-05 | Purchase Dashboard | Procurement | Dashboard | P2 | L2 | Purchasing Manager |
| QC-01 | Inbound Inspection | QA/QC | List + Detail + Decision | P1 | L1 | QC Officer |
| QC-02 | FG Inspection | QA/QC | List + Detail + Decision | P1 | L2 | QC Officer, QA Manager |
| QC-03 | QC Decision / Release | QA/QC | Decision Console | P1 | L1 | QA Manager |
| QC-04 | NCR List / Detail | QA/QC | List + Detail | P2 | L2 | QA, QC |
| QC-05 | QC Dashboard | QA/QC | Dashboard | P2 | L2 | QA Manager |
| PROD-01 | Production Order List | Production | List | P1 | L2 | Planner, Supervisor |
| PROD-02 | Production Order Detail | Production | Detail | P1 | L1 | Planner, Supervisor |
| PROD-03 | Material Issue | Production | Action Form | P1 | L1 | Production, Warehouse |
| PROD-04 | Production Confirmation | Production | Form | P1 | L1 | Supervisor |
| PROD-05 | Output & Scrap Entry | Production | Form | P1 | L1 | Supervisor, QA |
| PROD-06 | Production Dashboard | Production | Dashboard | P2 | L2 | COO, Planner |
| WH-01 | Goods Receipt | Warehouse | Action Form | P1 | L2 | Warehouse |
| WH-02 | Goods Issue | Warehouse | Action Form | P1 | L2 | Warehouse |
| WH-03 | Transfer Order | Warehouse | Form | P1 | L2 | Warehouse |
| WH-04 | Transfer Receipt | Warehouse | Action Form | P2 | L2 | Warehouse |
| WH-05 | Stock Count | Warehouse | Wizard + Count Sheet | P1 | L2 | Warehouse |
| WH-06 | Stock Adjustment | Warehouse | Form + Approval | P1 | L2 | Warehouse Manager |
| WH-07 | Stock Ledger | Warehouse | List / Ledger | P1 | L2 | Warehouse, Finance |
| WH-08 | Stock by Batch / Expiry | Warehouse | Grid + Drilldown | P1 | L1 | Warehouse, QC, Sales |
| WH-09 | Batch Trace | Warehouse | Trace Console | P1 | L1 | QA, Warehouse, Sales, COO |
| SAL-01 | Quotation | Sales | List + Detail + Form | P2 | L2 | Sales Admin |
| SAL-02 | Sales Order | Sales | List + Detail + Form | P1 | L1 | Sales Admin |
| SAL-03 | SO Approval | Sales | Approval List | P1 | L2 | Sales Manager, Finance |
| SAL-04 | Delivery Request / Delivery Order | Sales | Action Form | P1 | L1 | Sales Admin, Warehouse |
| SAL-05 | Sales Return | Sales | Return Form | P1 | L1 | Sales Admin, Warehouse, QA |
| SAL-06 | Sales Dashboard | Sales | Dashboard | P2 | L2 | Sales Manager, CEO |

---

## 9. Pattern tái sử dụng để giảm effort thiết kế

### 9.1. Master data list/detail pattern
Áp dụng cho:
- MD-06 Supplier
- MD-07 Customer
- MD-08 Warehouse
- MD-10 Price List (phần list)
- MD-11 cấu hình danh mục nhỏ

### 9.2. Document form with lines pattern
Áp dụng cho:
- PUR-01 PR
- PUR-02 PO
- SAL-01 Quotation
- SAL-02 SO
- WH-03 Transfer Order

### 9.3. Transaction action pattern
Áp dụng cho:
- PUR-03 Receiving
- PUR-04 Supplier Return
- PROD-03 Material Issue
- PROD-04 Production Confirmation
- PROD-05 Output & Scrap Entry
- WH-01 Goods Receipt
- WH-02 Goods Issue
- WH-04 Transfer Receipt
- SAL-04 Delivery Order
- SAL-05 Sales Return

### 9.4. Dashboard pattern
Áp dụng cho:
- Purchase Dashboard
- QC Dashboard
- Production Dashboard
- Sales Dashboard

### 9.5. Trace / ledger pattern
Áp dụng cho:
- WH-07 Stock Ledger
- WH-08 Stock by Batch / Expiry
- WH-09 Batch Trace

---

## 10. Đặc tả chi tiết các màn hình L1

---

## 10.1. COM-02 — Approval Inbox

**Mục tiêu nghiệp vụ:**  
Cho người duyệt nhìn thấy toàn bộ giao dịch đang chờ mình xử lý theo SLA.

**Người dùng chính:**  
Purchasing Manager, QA Manager, Warehouse Manager, Sales Manager, Finance Approver, COO, CEO.

**Entry points:**  
- Menu `Phê duyệt`
- Notification center
- Link sâu từ dashboard module

**Khối thông tin chính:**
1. KPI strip: Pending / Overdue / Due Today / Approved Today
2. Search + filter
3. Bảng danh sách duyệt
4. Drawer chi tiết bên phải
5. Action bar duyệt

**Bộ lọc chính:**
- Module
- Loại chứng từ
- Trạng thái
- Người tạo
- Đơn vị/kho
- Khoảng ngày
- SLA overdue
- Giá trị vượt ngưỡng

**Cột chính của bảng:**
- Document No
- Module
- Document Type
- Requester
- Related party
- Amount / Qty impact
- Created time
- SLA due
- Current step
- Status

**Hành động chính:**
- Approve
- Reject
- Request change
- Open full detail
- View approval history

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Approval Inbox                                         [Approve] [Reject] [Request change]      |
| KPI: Pending 24 | Overdue 5 | Due today 9 | Approved today 13                                   |
| Search [..............]  Filters [Module][Doc Type][Requester][Status][Date][SLA] [Save View]  |
|--------------------------------------------------------------------------------------------------|
| Table                                                                                            |
| Doc No | Module | Type | Requester | Party | Amount/Qty | Created | SLA | Step | Status        |
|--------------------------------------------------------------------------------------------------|
| PO-... | Proc   | PO   | Lan       | NCC A | 125,000,000| 09:10   | +4h | Mgr  | Pending       |
| SO-... | Sales  | SO   | Nam       | DL C  | 2,400 pcs  | 09:35   | +1h | Fin  | Overdue       |
| ...                                                                                              |
|--------------------------------------------------------------------------------------------------|
| Right drawer on selected row                                                                     |
| Summary | Key fields | Line preview | Attachments | Approval history | Audit                    |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Có thể duyệt nhanh từ drawer mà không phải mở full page.
- Có cảnh báo rõ nếu hành động duyệt sẽ mở khoá bước tiếp theo.
- Chứng từ overdue phải nổi bật hơn trong thứ tự và badge.

---

## 10.2. MD-01 — Item List

**Mục tiêu nghiệp vụ:**  
Tìm nhanh và quản lý danh mục nguyên liệu, bao bì, bán thành phẩm, thành phẩm.

**Người dùng chính:**  
Master Data Admin, Purchasing, Planner, Warehouse, Sales Admin.

**Khối thông tin chính:**
1. KPI mini: total items / active / inactive / missing setup / blocked
2. Search + advanced filters
3. Item table
4. Bulk actions
5. Import / Export

**Bộ lọc chính:**
- Item type
- Category
- Brand / line
- Active status
- QC required
- Batch controlled
- Has BOM
- Supplier assigned
- Missing required fields

**Cột chính:**
- Item code
- Item name
- Type
- UOM
- Brand
- Batch control
- Expiry control
- QC required
- Active
- Last updated

**Hành động chính:**
- Create item
- Edit
- Clone
- Activate / Deactivate
- Export
- Open item detail

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Item Master                                               [Create Item] [Import] [Export]       |
| KPI: Total 6,420 | Active 5,981 | Missing setup 142 | Inactive 439                               |
| Search [item code/name/barcode]  Filters [Type][Brand][QC][Batch][Active][Missing setup]       |
|--------------------------------------------------------------------------------------------------|
| Table                                                                                            |
| Code | Name | Type | UOM | Brand | Batch Ctrl | Expiry Ctrl | QC Req | Active | Updated       |
|--------------------------------------------------------------------------------------------------|
| RM-...                                                                                           |
| FG-...                                                                                           |
| PK-...                                                                                           |
|--------------------------------------------------------------------------------------------------|
| Bottom: pagination | selected rows | bulk activate/deactivate/export                             |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Search phải cho phép tìm theo code, tên, alias, barcode.
- Có cột cảnh báo cấu hình thiếu.
- Chọn 1 row mở sang MD-02.

---

## 10.3. MD-02 — Item Detail / Create / Edit

**Mục tiêu nghiệp vụ:**  
Chuẩn hoá một item để đủ điều kiện dùng trong mua hàng, sản xuất, kho và bán hàng.

**Người dùng chính:**  
Master Data Admin.

**Tabs đề xuất:**
1. General
2. Classification
3. Inventory & Batch
4. Procurement / Sales
5. Cost & Planning
6. Attachments
7. Audit

**Field group chính:**
- Header: item code, item name, status, item type
- General: alias, description, barcode, default UOM
- Classification: category, brand, line, tax group, storage condition
- Inventory: batch control, expiry control, shelf life days, FEFO/FIFO
- QC: QC required, inspection template
- Procurement: preferred supplier, lead time, MOQ
- Sales: sellable flag, price list link
- Cost: standard cost, costing method
- Attachments: spec, COA mẫu, label, artwork

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Item Detail: FG-SERUM-30ML-VC                             [Save Draft] [Submit] [Clone]         |
| Status: Active   Type: Finished Good   Last update: 2026-04-23 14:21                            |
| Tabs: General | Classification | Inventory & Batch | Procurement/Sales | Cost | Attachments... |
|--------------------------------------------------------------------------------------------------|
| Section A - Basic Info                                                                            |
| Code [........]  Name [........................................] Type [FG] UOM [PCS]             |
| Alias [...................] Barcode [................] Active [Yes/No]                            |
|--------------------------------------------------------------------------------------------------|
| Section B - Inventory & QC                                                                        |
| Batch control [Yes] Expiry control [Yes] Shelf life days [1095]                                  |
| FEFO/FIFO [FEFO] QC required [Yes] Inspection template [FG-COSMETIC]                             |
|--------------------------------------------------------------------------------------------------|
| Section C - Business                                                                              |
| Preferred supplier [....] Lead time [..] MOQ [..] Sellable [Yes] Price list [Retail-2026]      |
|--------------------------------------------------------------------------------------------------|
| Right rail: Missing setup checklist | Attachments quick view | Audit summary                     |
+--------------------------------------------------------------------------------------------------+
```

**Validation bắt buộc:**
- Item code unique.
- Nếu `batch_control = Yes` thì phải có batch rule.
- Nếu `expiry_control = Yes` thì phải có shelf life days > 0.
- Nếu item type là Finished Good thì phải có `sellable flag`.
- Không deactivate item đang có giao dịch mở nếu chưa có rule đặc biệt.

---

## 10.4. MD-05 — BOM Detail / Version

**Mục tiêu nghiệp vụ:**  
Quản lý BOM/công thức, version và thành phần tiêu hao chuẩn cho sản xuất.

**Người dùng chính:**  
Master Data Admin, Planner, R&D (phase sau có thể mở thêm).

**Khối thông tin chính:**
1. Header BOM
2. Version list
3. Component lines
4. Yield / loss assumptions
5. Cost estimate
6. Approval / effective date

**Cột line item:**
- Component code
- Component name
- Qty per batch / per unit
- UOM
- Scrap factor
- Alternate allowed
- Mandatory flag
- Notes

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| BOM Detail: BOM-FG-SERUM-VC-30ML                         [New Version] [Activate] [Expire]       |
| Product: FG-SERUM-VC-30ML  Current version: V3  Effective: 2026-05-01                            |
|--------------------------------------------------------------------------------------------------|
| Version panel: V1 (Expired) | V2 (Expired) | V3 (Active)                                         |
|--------------------------------------------------------------------------------------------------|
| Header: Base batch size [5000 pcs]   Expected yield [%] [97.5]   Standard loss [%] [2.5]        |
|--------------------------------------------------------------------------------------------------|
| Components                                                                                       |
| Code | Name | Qty | UOM | Scrap Factor | Alt Item Allowed | Mandatory | Notes                    |
|--------------------------------------------------------------------------------------------------|
| RM-...                                                                                           |
| PK-...                                                                                           |
|--------------------------------------------------------------------------------------------------|
| Summary: estimated material cost | packaging cost | total standard cost                           |
| Right rail: effectivity rules | linked production orders | approval history                       |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- So sánh nhanh 2 version.
- Không cho edit version đã active nếu đã có Production Order sử dụng.
- Phải nhìn được cost estimate ngay trên màn hình.

---

## 10.5. PUR-01 — Purchase Requisition

**Mục tiêu nghiệp vụ:**  
Tạo và theo dõi nhu cầu mua từ bộ phận yêu cầu hoặc purchasing.

**Người dùng chính:**  
Requester, Purchasing Officer.

**Chế độ màn hình:**
- List
- Create/Edit
- Detail

**Field nhóm header:**
- PR number
- Requesting department
- Requester
- Need-by date
- Priority
- Reason / purpose

**Line item chính:**
- Item code
- Item name
- Requested qty
- UOM
- Suggested supplier
- Need-by date
- Reference stock / shortage
- Notes

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Purchase Requisition                                   [Save Draft] [Submit for Approval]        |
| PR No [auto] Department [Production] Requester [Lan] Need-by [2026-04-29] Priority [High]      |
| Reason [......................................................................................]  |
|--------------------------------------------------------------------------------------------------|
| Line items                                                                                       |
| Add item [search code/name]                                                                      |
| Code | Name | Req Qty | UOM | Suggested Supplier | Need-by | Shortage Ref | Notes               |
|--------------------------------------------------------------------------------------------------|
| RM-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Summary: total line count | estimated value (optional)                                           |
| Right rail: approval route preview | attachments | comments                                      |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Khi nhập item phải gợi ý tồn hiện tại, open PO và shortage nếu có.
- Có thể import line item từ Excel chuẩn.
- Submit phải chạy validate các trường bắt buộc.

---

## 10.6. PUR-02 — Purchase Order

**Mục tiêu nghiệp vụ:**  
Tạo PO chuẩn cho NCC, theo dõi nhận hàng từng phần và trạng thái mở/đóng.

**Người dùng chính:**  
Purchasing Officer.

**Khối thông tin chính:**
1. Supplier header
2. Commercial terms
3. Line item grid
4. Receiving summary
5. Attachments & terms

**Field chính:**
- PO number
- Supplier
- Supplier contact
- Currency
- Payment terms
- Expected receipt date
- Ship to warehouse
- Incoterm/notes (nếu dùng)
- Approval status

**Line item cột chính:**
- Item
- Ordered qty
- Received qty
- Open qty
- Unit price
- Discount
- Tax
- Line total

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Purchase Order                                       [Save] [Submit] [Send to Supplier]          |
| PO No [auto] Supplier [NCC A] Warehouse [RM-WH] Expected receipt [2026-05-02]                   |
| Currency [VND] Payment terms [30 days] Buyer [Minh]                                              |
|--------------------------------------------------------------------------------------------------|
| Lines                                                                                            |
| Code | Name | Ordered | Received | Open | UOM | Unit Price | Disc% | Tax | Line Total           |
|--------------------------------------------------------------------------------------------------|
| RM-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Totals: Subtotal | Discount | Tax | Grand Total                                                  |
| Tabs: Attachments | Terms | Approval | Receiving history | Audit                                 |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Nhìn thấy ngay `Received qty` và `Open qty`.
- PO được tạo từ PR phải giữ liên kết source.
- Nếu vượt ngưỡng duyệt thì action `Submit` phải hiển thị route phê duyệt dự kiến.

---

## 10.7. PUR-03 — Receiving against PO

**Mục tiêu nghiệp vụ:**  
Nhập hàng thực nhận theo PO và sinh batch inbound ban đầu.

**Người dùng chính:**  
Warehouse Staff, Purchasing Officer.

**Khối thông tin chính:**
1. PO summary
2. Receipt lines
3. Batch / expiry capture
4. QC flag / hold zone
5. Received document summary

**Cột line receipt:**
- Item
- Ordered qty
- Previously received
- This receipt qty
- UOM
- Warehouse / bin
- Supplier batch
- Internal batch
- MFG date
- EXP date
- QC required
- Putaway status

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Receiving against PO                                [Post Receipt] [Save Draft]                  |
| Source PO: PO-202604-0021   Supplier: NCC A   Warehouse: RM-WH                                  |
|--------------------------------------------------------------------------------------------------|
| Lines to receive                                                                                 |
| Item | Ordered | Prev Rec | This Rec | UOM | Bin | Supplier Batch | Internal Batch | MFG | EXP |
|--------------------------------------------------------------------------------------------------|
| RM-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Options: [x] QC required items go to Hold stock   [ ] Print receiving label                      |
| Summary: receipt lines | receipt qty | exceptions                                                |
| Right rail: PO terms | attachments | QC preview                                                  |
+--------------------------------------------------------------------------------------------------+
```

**Validation bắt buộc:**
- Không cho nhận > open qty nếu không có override.
- Với item bắt buộc batch control, internal batch không được trống.
- Với expiry control, phải nhập MFG/EXP hoặc shelf life rule sinh tự động.
- QC required thì stock sau nhận phải vào trạng thái Hold/Inspection.

---

## 10.8. QC-01 — Inbound Inspection

**Mục tiêu nghiệp vụ:**  
Thực hiện kiểm tra chất lượng đầu vào cho lô nhận hàng.

**Người dùng chính:**  
QC Officer.

**Khối thông tin chính:**
1. Inspection queue
2. Sample info
3. Spec/result sheet
4. Attachments
5. Recommendation

**Bộ lọc chính:**
- Warehouse
- Supplier
- Item
- Batch
- Overdue inspection
- Status

**Cấu trúc detail:**
- Header: inspection no, item, batch, receipt ref
- Sampling info
- Checklist / parameters
- Result entry
- Preliminary recommendation: Pass / Hold / Fail

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Inbound Inspection Queue                               [Start Inspection] [Save Result]          |
| Filters [Warehouse][Supplier][Item][Batch][Status][Overdue]                                     |
|--------------------------------------------------------------------------------------------------|
| Left list: pending inspections                                                                     |
| IQC No | Receipt | Item | Batch | Qty | Received Date | SLA | Status                             |
|--------------------------------------------------------------------------------------------------|
| Right detail                                                                                    |
| Header: IQC-... Item RM-... Batch RM202604... Receipt GRN-...                                    |
| Sampling: sampled qty [..] inspector [..] date [..]                                              |
| Checklist / results                                                                               |
| Param | Spec | Method | Result | Pass/Fail | Note                                                |
|--------------------------------------------------------------------------------------------------|
| Footer: Recommendation [Pass/Hold/Fail] Reason [...........]                                      |
| Tabs: Attachments | Photos | History                                                              |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Có cảnh báo inspection overdue.
- Checklist phải có hỗ trợ nhập nhanh và copy spec chuẩn.
- Không cho QCO tự release final nếu policy yêu cầu QA Manager duyệt.

---

## 10.9. QC-03 — QC Decision / Release

**Mục tiêu nghiệp vụ:**  
Cho QA Manager chốt quyết định cuối cùng và đổi trạng thái lô.

**Người dùng chính:**  
QA Manager.

**Khối thông tin chính:**
1. Pending decisions list
2. Batch summary
3. Inspection results summary
4. Related docs
5. Decision panel

**Quyết định có thể chọn:**
- Pass
- Hold
- Fail
- Partial release (nếu policy cho phép)

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| QC Decision / Release                                 [Pass] [Hold] [Fail] [Request Recheck]    |
| Filters [Inbound/FG][Item][Batch][Status][Overdue]                                               |
|--------------------------------------------------------------------------------------------------|
| Table: IQC/FGQC No | Item | Batch | Qty | Current Status | Inspector | Created | SLA           |
|--------------------------------------------------------------------------------------------------|
| Selected record summary                                                                            |
| Item / Batch / Qty / Receipt or MO / Supplier or Product                                          |
| Result summary: total checks | failed checks | attachments                                        |
| Linked docs: GRN / MO / Stock status / previous inspections                                       |
| Decision panel: reason, disposition, notes, quarantine action                                     |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Pass phải hiển thị rõ stock sẽ chuyển từ Hold sang Available.
- Fail phải gắn luôn disposition gợi ý: return / scrap / quarantine.
- Mọi quyết định phải bắt buộc lưu reason và sinh audit log.

---

## 10.10. PROD-02 — Production Order Detail

**Mục tiêu nghiệp vụ:**  
Là màn hình trung tâm cho một lệnh sản xuất.

**Người dùng chính:**  
Production Planner, Production Supervisor, COO.

**Tabs đề xuất:**
1. Overview
2. BOM snapshot
3. Material issue
4. Progress / confirmation
5. Output & scrap
6. QC / FG inspection
7. Related docs
8. Audit

**Field tổng quan:**
- MO number
- Product
- BOM version
- Planned qty
- Produced qty
- Scrap qty
- Status
- Planned start/end
- Warehouse in/out
- Supervisor

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Production Order: MO-202604-0018                         [Release] [Issue Material] [Confirm]    |
| Product: FG-SERUM-VC-30ML  BOM V3  Planned 5,000  Produced 2,400  Scrap 55  Status: Released   |
|--------------------------------------------------------------------------------------------------|
| Tabs: Overview | BOM Snapshot | Material Issue | Confirmation | Output & Scrap | QC | Audit     |
|--------------------------------------------------------------------------------------------------|
| Overview                                                                                         |
| Planned start [..] Planned end [..] RM warehouse [..] FG warehouse [..] Supervisor [..]        |
| KPI strip: issue progress | output progress | variance vs BOM | pending QC                        |
|--------------------------------------------------------------------------------------------------|
| Right rail: linked issues | linked FGQC | notes | alerts                                          |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- BOM snapshot phải là bản đã khóa theo thời điểm MO.
- Nhìn được variance cơ bản giữa plan và actual.
- Link sang PROD-03 / PROD-04 / PROD-05 phải ngay trên header.

---

## 10.11. PROD-03 — Material Issue

**Mục tiêu nghiệp vụ:**  
Xuất nguyên liệu từ kho cho Production Order theo BOM và batch cụ thể.

**Người dùng chính:**  
Production Supervisor, Warehouse Staff.

**Khối thông tin chính:**
1. MO summary
2. Suggested issue lines
3. Batch selection
4. Issue variance check
5. Post issue

**Cột line item:**
- Component
- Planned qty
- Issued qty to date
- This issue qty
- Available stock
- Batch selected
- Expiry
- Variance vs BOM

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Material Issue for MO-202604-0018                         [Auto Suggest] [Post Issue]            |
| Product FG-SERUM-VC-30ML | Planned 5,000 pcs | RM Warehouse RM-WH                                |
|--------------------------------------------------------------------------------------------------|
| Components                                                                                       |
| Code | Name | Planned | Issued To Date | This Issue | UOM | Available | Batch | EXP | Variance |
|--------------------------------------------------------------------------------------------------|
| RM-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Footer: total planned | total this issue | exception count                                        |
| Right rail: FEFO suggestion | blocked batches | approval note if over-issue                       |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Batch selection nên gợi ý FEFO.
- Over-issue phải cảnh báo và yêu cầu override nếu vượt tolerance.
- Không cho chọn batch đang Hold/Fail.

---

## 10.12. PROD-04 — Production Confirmation

**Mục tiêu nghiệp vụ:**  
Xác nhận tiến độ / sản lượng theo đợt hoặc theo ca.

**Người dùng chính:**  
Production Supervisor.

**Khối thông tin chính:**
- Confirmation time
- Shift / line
- In-progress qty
- Completed qty
- Downtime / incident
- Notes

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Production Confirmation                               [Save Confirmation]                         |
| MO: MO-202604-0018  Shift [A]  Line [Filling-02]  Confirm time [2026-04-23 16:30]              |
|--------------------------------------------------------------------------------------------------|
| In-progress qty [....] Completed qty this step [....] Downtime minutes [....]                    |
| Incident code [....] Notes [..................................................................] |
|--------------------------------------------------------------------------------------------------|
| Previous confirmations list                                                                       |
| Time | Shift | Completed | Downtime | By                                                         |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Có thể ghi nhiều confirmation cho cùng 1 MO.
- Dữ liệu confirmation không thay thế output chính, mà là dữ liệu tiến độ.

---

## 10.13. PROD-05 — Output & Scrap Entry

**Mục tiêu nghiệp vụ:**  
Nhập thành phẩm / bán thành phẩm đầu ra và ghi scrap, chờ QC thành phẩm.

**Người dùng chính:**  
Production Supervisor, QA.

**Khối thông tin chính:**
1. Output header
2. Batch generation
3. Output lines
4. Scrap lines
5. QC flag

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Output & Scrap Entry                                  [Generate Batch] [Post Output]             |
| MO: MO-202604-0018  Product: FG-SERUM-VC-30ML                                                      |
|--------------------------------------------------------------------------------------------------|
| Output section                                                                                   |
| Output qty [4,850] UOM [PCS] FG Warehouse [FG-WH] Internal Batch [FG20260423-01] MFG [..] EXP[..]|
| QC required [Yes] -> stock after post: Hold                                                       |
|--------------------------------------------------------------------------------------------------|
| Scrap / Loss                                                                                      |
| Scrap qty [55] Reason [Process Loss] Notes [....]                                                 |
|--------------------------------------------------------------------------------------------------|
| Summary: planned vs produced vs scrap vs remaining                                                |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Batch thành phẩm có thể sinh tự động theo rule nhưng vẫn cho review.
- Output đăng sổ xong phải liên kết ngay tới FG Inspection nếu QC required.
- Scrap reason phải chọn từ danh mục chuẩn.

---

## 10.14. WH-08 — Stock by Batch / Expiry

**Mục tiêu nghiệp vụ:**  
Là màn hình tác chiến của kho, sales, QC để nhìn tồn theo batch, hạn dùng, trạng thái.

**Người dùng chính:**  
Warehouse, QC, Sales Admin, COO.

**Bộ lọc chính:**
- Warehouse
- Item / group
- Batch
- Expiry range
- Stock status
- Near-expiry
- Hold/Fail only

**Cột chính:**
- Warehouse
- Item
- Batch
- MFG
- EXP
- Days to expiry
- Physical qty
- Reserved qty
- Hold qty
- Available qty
- QC status
- Source doc

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Stock by Batch / Expiry                                [Export] [Batch Trace]                    |
| Filters [Warehouse][Item][Batch][Expiry Range][Status][Near Expiry]                             |
| KPI: Total batches 812 | Near expiry 34 | Hold 19 | Available qty xxx                            |
|--------------------------------------------------------------------------------------------------|
| Warehouse | Item | Batch | MFG | EXP | Days Left | Physical | Reserved | Hold | Available | QC |
|--------------------------------------------------------------------------------------------------|
| FG-WH     | FG-...| ...   | ... | ... | 120       | 500      | 120      | 0    | 380       | Pass|
| RM-WH     | RM-...| ...   | ... | ... | 18        | 300      | 0        | 300  | 0         | Hold|
|--------------------------------------------------------------------------------------------------|
| Row action: open ledger | trace batch | reserve impact | related QC                               |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Cột `Available` phải tính đúng theo rule Phase 1.
- Near-expiry nên có quick filter và sort theo days left.
- Cần export được để phục vụ điều độ và sale.

---

## 10.15. WH-09 — Batch Trace

**Mục tiêu nghiệp vụ:**  
Truy xuất xuôi/ngược một batch để xử lý lỗi, complaint, recall, trả hàng hoặc kiểm tra tồn.

**Người dùng chính:**  
QA, Warehouse, Sales, COO.

**Chế độ truy xuất:**
- Trace forward
- Trace backward

**Nguồn vào có thể là:**
- Batch RM
- Batch FG
- SO / DO
- MO
- GRN / Receipt
- Return

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Batch Trace                                            [Trace] [Export Chain]                    |
| Search input: [Batch No / SO / MO / GRN / Item]      Mode: (Backward / Forward)                 |
|--------------------------------------------------------------------------------------------------|
| Result summary: Item | Batch | Status | Qty | Warehouse | MFG | EXP                              |
|--------------------------------------------------------------------------------------------------|
| Trace chain                                                                                      |
| Batch FG -> MO -> Material Issue -> RM Batch -> Receiving -> Supplier                            |
| Batch FG -> Delivery Orders -> Customers -> Sales Returns                                        |
|--------------------------------------------------------------------------------------------------|
| Side panel: linked QC decisions | attachments | notes | alerts                                   |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Phải hỗ trợ nhìn chain dạng cây hoặc step list.
- Mọi node trong trace chain phải clickable sang chứng từ gốc.
- Nếu dữ liệu trace thiếu, hệ thống phải báo rõ missing link ở đoạn nào.

---

## 10.16. SAL-02 — Sales Order

**Mục tiêu nghiệp vụ:**  
Tạo SO, kiểm tra tồn khả dụng, giữ hàng và mở đường cho giao hàng.

**Người dùng chính:**  
Sales Admin.

**Tabs đề xuất:**
1. Header
2. Lines
3. Allocation / Reservation
4. Delivery
5. Return history
6. Approval
7. Audit

**Field chính:**
- SO number
- Customer
- Ship-to
- Order date
- Requested delivery date
- Price list
- Payment term
- Credit note / warning
- Salesperson

**Cột line chính:**
- Item
- Ordered qty
- UOM
- Available stock
- Reserved qty
- Unit price
- Discount
- Tax
- Line total
- Batch preference (optional policy)

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Sales Order                                           [Save] [Submit] [Reserve Stock]            |
| SO No [auto] Customer [DL-CM-001] Order date [..] Req delivery [..] Price list [DL-2026]       |
| Credit status [Warning / OK]  Salesperson [Nam]                                                  |
|--------------------------------------------------------------------------------------------------|
| Lines                                                                                            |
| Item | Ordered | UOM | Available | Reserved | Unit Price | Disc% | Tax | Line Total             |
|--------------------------------------------------------------------------------------------------|
| FG-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Totals: subtotal | discount | tax | grand total                                                  |
| Right rail: credit warning | approval route | related quotations                                 |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Khi nhập item phải thấy `available stock`, không dùng physical stock.
- Discount vượt ngưỡng phải cảnh báo ngay tại dòng.
- Có action `Reserve Stock` với feedback rõ đã giữ bao nhiêu.

---

## 10.17. SAL-04 — Delivery Request / Delivery Order

**Mục tiêu nghiệp vụ:**  
Chuyển SO sang bước giao hàng và gắn với kho xuất thực tế.

**Người dùng chính:**  
Sales Admin, Warehouse Staff.

**Khối thông tin chính:**
1. SO summary
2. Deliverable lines
3. Batch allocation
4. Shipping info
5. Posting result

**Cột line chính:**
- Item
- Ordered qty
- Reserved qty
- Deliver now qty
- Warehouse
- Batch picked
- QC status
- Remaining after delivery

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Delivery Order                                      [Allocate Batch] [Post Delivery]             |
| Source SO: SO-202604-0095   Customer: DL-CM-001   Ship date [..]   Warehouse [FG-WH]            |
|--------------------------------------------------------------------------------------------------|
| Deliverable lines                                                                                 |
| Item | Ordered | Reserved | Deliver Now | Warehouse | Batch | QC Status | Remaining             |
|--------------------------------------------------------------------------------------------------|
| FG-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Shipping: ship-to [....] carrier [manual] note [....]                                            |
| Right rail: reservation status | return risk note | attachments                                  |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Chỉ cho chọn batch Pass và Available.
- Nếu Deliver Now < Ordered, hệ thống phải tự hiểu partial delivery.
- Sau post, phải cập nhật reservation và stock ledger ngay.

---

## 10.18. SAL-05 — Sales Return

**Mục tiêu nghiệp vụ:**  
Nhận hàng trả lại, phân loại tình trạng và đưa hàng về đúng trạng thái kho/chất lượng.

**Người dùng chính:**  
Sales Admin, Warehouse, QA.

**Khối thông tin chính:**
1. Source document reference
2. Returned item lines
3. Return reason
4. Condition assessment
5. Disposition

**Cột line chính:**
- Item
- Original delivered qty
- Return qty
- Batch
- Reason
- Condition
- Proposed disposition
- QC required after return

**Wireframe chữ:**

```text
+--------------------------------------------------------------------------------------------------+
| Sales Return                                         [Save] [Submit] [Post Return Receipt]       |
| Source SO/DO [search...] Customer [auto] Return date [..]                                        |
|--------------------------------------------------------------------------------------------------|
| Returned lines                                                                                   |
| Item | Original Qty | Return Qty | Batch | Reason | Condition | Disposition | QC Req            |
|--------------------------------------------------------------------------------------------------|
| FG-...                                                                                            |
|--------------------------------------------------------------------------------------------------|
| Summary: return qty | pending QC | scrap candidate                                                |
| Right rail: linked original order | customer note | approval if required                         |
+--------------------------------------------------------------------------------------------------+
```

**Điểm UX bắt buộc:**
- Return phải bám được source SO/DO nếu có.
- Hàng trả lại không tự động vào Available nếu policy yêu cầu QC.
- Condition/disposition phải chuẩn hóa theo danh mục.

---

## 11. Màn hình L2 và cách kế thừa pattern

### 11.1. COM-01 Login
Kế thừa pattern auth chuẩn:
- company logo / tên hệ thống;
- username / password;
- SSO option nếu có;
- forgot password (nếu policy cho phép).

### 11.2. COM-03 Audit Log Viewer
Kế thừa pattern list với filter:
- module,
- doc type,
- record no,
- field changed,
- changed by,
- date range.

### 11.3. MD-03 Import Item
Wizard 3 bước:
1. Download template
2. Upload file và map column
3. Preview lỗi / commit import

### 11.4. MD-06 / MD-07 / MD-08
Reuse `MD-01 + MD-02` pattern với field set khác nhau:
- Supplier: payment terms, lead time, approved status
- Customer: customer type, price list, credit policy
- Warehouse: warehouse type, bin rule, default QC zone

### 11.5. MD-10 Price List & Discount Rule
Reuse list/detail with tabs:
- header,
- item pricing,
- customer group,
- effective dates,
- override rules.

### 11.6. QC-02 FG Inspection
Reuse `QC-01` nhưng source là Production Output / FG batch.

### 11.7. QC-04 NCR
Reuse list + detail + attachment + disposition workflow.

### 11.8. PROD-01 Production Order List
Reuse advanced list pattern với column:
- MO,
- product,
- planned qty,
- status,
- start/end,
- issue progress,
- output progress.

### 11.9. WH-01 / WH-02 / WH-03 / WH-04 / WH-05 / WH-06
Reuse transaction action pattern:
- source reference,
- warehouse/bin,
- item lines,
- batch/expiry,
- posting summary,
- approval/audit.

### 11.10. SAL-01 Quotation
Reuse `SAL-02` pattern nhưng không có reservation/delivery.

### 11.11. SAL-03 SO Approval
Có thể dùng derived version của `COM-02 Approval Inbox` với default filter = Sales Order.

### 11.12. Dashboard screens
Reuse dashboard pattern gồm:
- KPI cards,
- trend chart,
- exception list,
- top issue widgets,
- quick links.

---

## 12. Wireframe pattern cho dashboard

### 12.1. Cấu trúc chuẩn

```text
+--------------------------------------------------------------------------------------------------+
| Page title                                  Date range [Today/Week/Month] [Export]               |
| KPI 1 | KPI 2 | KPI 3 | KPI 4                                                                   |
|--------------------------------------------------------------------------------------------------|
| Trend chart / bar chart                                                                        |
|--------------------------------------------------------------------------------------------------|
| Left widget: exception list         | Right widget: top items / top suppliers / top customers   |
|--------------------------------------------------------------------------------------------------|
| Bottom table: open documents / overdue / alerts                                                 |
+--------------------------------------------------------------------------------------------------+
```

### 12.2. Dashboard Phase 1 nên có các widget sau

#### Purchase Dashboard
- Open PR count
- Open PO count
- Overdue PO
- Supplier on-time receipt
- Items with shortage

#### QC Dashboard
- Pending inspections
- Overdue inspections
- Hold batches
- Fail ratio by supplier / item
- Awaiting QA decision

#### Production Dashboard
- MO by status
- Plan vs output
- Material issue variance
- Pending FG inspection
- Scrap by day / product

#### Sales Dashboard
- Open SO
- Pending delivery
- Partial deliveries
- Sales return count
- Orders blocked by stock / credit

---

## 13. Ma trận màn hình theo vai trò

| Screen ID | MDA | PUR | QCO | QAM | PPL/PSV | WH | SAD | MGR/FIN/EXE |
|---|---|---|---|---|---|---|---|---|
| COM-02 | View | View | View | Approve | View | View | View | Approve |
| MD-01 | CRUD | View | View | View | View | View | View | View |
| MD-02 | CRUD | View | View | View | View | View | View | View |
| MD-05 | CRUD | View | View | View | View | View | - | View |
| PUR-01 | View | CRUD | - | - | Request | View | - | Approve |
| PUR-02 | View | CRUD | - | - | View | View | - | Approve |
| PUR-03 | View | View | View | View | - | Post | - | View |
| QC-01 | View | View | CRUD | View | - | View | - | View |
| QC-03 | View | View | Propose | Approve | - | View | - | View |
| PROD-02 | View | View | View | View | CRUD | View | - | View |
| PROD-03 | View | - | - | - | CRUD | Post | - | Approve if override |
| PROD-04 | - | - | - | - | CRUD | - | - | View |
| PROD-05 | - | - | View | View | CRUD | View | - | View |
| WH-08 | View | View | View | View | View | View/Post context | View | View |
| WH-09 | View | View | View | View | View | View | View | View |
| SAL-02 | View | - | - | - | - | View | CRUD | Approve |
| SAL-04 | - | - | - | - | - | Post | CRUD | View |
| SAL-05 | - | - | View | View | - | Post | CRUD | Approve if needed |

---

## 14. Quy ước UI quan trọng theo nghiệp vụ mỹ phẩm

### 14.1. Hàng Hold phải nổi hơn hàng thường
Ở các màn hình kho, QC, batch trace:
- Hold stock phải dễ nhận biết hơn;
- không để người dùng vô tình tưởng là hàng available.

### 14.2. Cận date phải được cảnh báo sớm
Trên `WH-08`, `SAL-04`, `Batch Trace`, hệ thống nên hiển thị:
- days to expiry,
- quick filter near-expiry,
- cảnh báo khi chọn batch có ngày còn lại thấp.

### 14.3. Batch không được giấu trong tab quá sâu
Batch là field sống còn, phải nhìn thấy ngay tại:
- receiving,
- inspection,
- material issue,
- output,
- delivery,
- sales return.

### 14.4. Line item phải hỗ trợ keyboard-first
Với các form nhiều dòng như PO, SO, Material Issue:
- hỗ trợ tab/enter,
- paste nhiều dòng,
- copy xuống,
- chọn item bằng typeahead search.

### 14.5. Không hiển thị cost nhạy cảm cho mọi vai trò
Một số field như standard cost, estimated value, margin, override log cần hide theo role.

---

## 15. Trạng thái rỗng, lỗi và ngoại lệ

### 15.1. Empty state
Không để màn hình trống vô hồn.  
Ví dụ:
- `Chưa có PR nào trong view này`
- `Chưa có batch gần hết hạn theo bộ lọc hiện tại`
- `Không có inspection đang overdue`

### 15.2. Error state
Thông báo lỗi phải cụ thể:
- `Không thể post receipt vì item RM-001 thiếu expiry date`
- `Không thể issue batch FG202604-01 vì batch đang Hold`
- `Không thể reserve stock vì available qty < ordered qty`

### 15.3. Warning state
Phân biệt warning với blocking:
- warning: still allow but ask confirmation / approval;
- blocking: không cho thao tác tiếp.

### 15.4. Unsaved change guard
Các form dài như Item Detail, PO, SO, MO cần cảnh báo khi đóng tab mà chưa save.

---

## 16. Responsive rules cho kho và xưởng

### 16.1. Tablet-first cho kho
Các màn hình `Receiving`, `Goods Issue`, `Stock Count`, `Transfer Receipt` cần đảm bảo dùng được trên tablet:
- dòng to hơn,
- nút rõ ràng,
- scan barcode dễ,
- tối ưu một tay ở hiện trường.

### 16.2. Mobile-lite cho phê duyệt
Phase 1 không cần full ERP trên mobile, nhưng nên support:
- Approval Inbox
- Notification Center
- Detail read-only cơ bản
- KPI quick glance

### 16.3. Desktop-first cho backoffice
Các màn hình nặng line item như:
- PO,
- SO,
- BOM,
- Material Issue,
- Stock by Batch
phải tối ưu desktop/laptop.

---

## 17. Sequence đề xuất cho UI/UX team triển khai wireframe

### Wave 1 — lõi dùng nhiều nhất
1. COM-02 Approval Inbox
2. MD-01 / MD-02
3. PUR-01 / PUR-02 / PUR-03
4. QC-01 / QC-03

### Wave 2 — lõi sản xuất và kho
5. PROD-02 / PROD-03 / PROD-04 / PROD-05
6. WH-08 / WH-09
7. WH transaction pattern

### Wave 3 — lõi bán hàng
8. SAL-02 / SAL-04 / SAL-05
9. Dashboard pattern
10. Màn hình derived khác

---

## 18. Checklist handoff cho UI/UX -> Dev

Trước khi handoff màn hình cho dev, mỗi screen phải chốt:

1. Screen ID
2. User role
3. Route / entry point
4. List các state của màn hình
5. Danh sách field + data type + required/read-only
6. Validation rule
7. Primary/secondary actions
8. Empty state / error state
9. Permission notes
10. Linked APIs / data source dự kiến
11. Audit / attachment / approval requirements
12. UAT cases liên quan

---

## 19. Checklist sign-off cho business

Business chỉ nên sign-off wireframe khi đã thấy rõ:

- Màn hình này giải đúng bước nghiệp vụ nào.
- Trường nào bắt buộc, trường nào không.
- Người nào dùng màn hình này hằng ngày.
- Có thiếu thông tin batch / QC / expiry / stock status không.
- Có đủ nhanh để đội kho, purchasing, sales thao tác không.
- Có lộ dữ liệu nhạy cảm cho sai vai trò không.
- Khi lỗi phát sinh, người dùng có biết phải làm tiếp gì không.

---

## 20. Các quyết định còn cần chốt trước khi chuyển sang hi-fi design

1. Có hỗ trợ multi-language trong Phase 1 hay không.
2. Có scan barcode/QR trực tiếp trên web tablet hay không.
3. Có cho partial release / partial return trong policy hay không.
4. Có hiển thị giá vốn chuẩn cho Purchasing / Sales hay không.
5. Có dùng drawer detail hay mở full page cho mọi loại chứng từ hay không.
6. Có in label batch / receiving label ngay sau receipt hay không.
7. Có cần dark mode cho môi trường xưởng/kho hay không.
8. Có cần số lượng lớn line item bằng spreadsheet-like grid hay mức vừa đủ.

---

## 21. Kết luận

Tài liệu này chốt 3 thứ sống còn cho đội làm sản phẩm:

1. **Toàn bộ danh sách màn hình của Phase 1**
2. **Pattern UI nào cần dựng trước và tái sử dụng**
3. **Wireframe mức thấp cho các màn hình xương sống của nghiệp vụ**

Nếu làm đúng theo tài liệu này, đội dự án sẽ tránh được 3 sai lầm rất hay gặp:
- design đẹp nhưng không chạy nổi nghiệp vụ,
- cùng một logic mà mỗi màn hình làm một kiểu,
- build xong mới phát hiện thiếu batch, expiry, QC status, approval context.

---

## 22. Đề xuất tên file tài liệu

`08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`