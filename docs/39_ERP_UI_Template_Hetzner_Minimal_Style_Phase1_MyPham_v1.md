# 39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1

**Project:** Web ERP cho công ty mỹ phẩm  
**Phase:** Phase 1  
**Document Type:** UI Template / Visual Direction / Screen Template Pack  
**Version:** v1.0  
**Date:** 2026-04-26  
**Style Direction:** Hetzner-inspired Industrial Minimal ERP UI  
**Primary Audience:** CEO, Product Owner, UI/UX Designer, Frontend Developer, Backend Developer, QA, Vendor  

---

## 1. Mục tiêu tài liệu

Tài liệu này khóa **ngôn ngữ giao diện** cho ERP Phase 1 theo hướng tối giản, công nghiệp, rõ ràng, giống tinh thần giao diện của Hetzner: nhiều khoảng trắng, typography chắc, layout chức năng, card/tile rõ ràng, ít màu, dùng đỏ làm điểm nhấn.

Đây không phải tài liệu để copy y nguyên website Hetzner. Đây là bản chuyển hóa tinh thần đó thành **ERP productivity UI** dùng cho kho, sản xuất/gia công ngoài, bán hàng, QC, giao hàng, hàng hoàn, master data và dashboard vận hành.

Mục tiêu:

- Giao diện nhìn sạch, nghiêm túc, dễ dùng.
- Người dùng kho/sales/QC/sản xuất thao tác nhanh, ít suy nghĩ.
- Designer có style rõ để vẽ Figma.
- Frontend có token, component, layout, template để code đồng bộ.
- Không tạo “SaaS màu mè”, không dashboard trình diễn vô dụng.
- Ưu tiên tốc độ đọc, tốc độ nhập, tốc độ quét mã, tốc độ xử lý nghiệp vụ.

---

## 2. Style name

```text
Industrial Minimal ERP UI
```

Tinh thần thiết kế:

```text
Clean. Sharp. Functional. Dense but readable.
Enterprise without being heavy.
Minimal without being empty.
Serious without being ugly.
```

Dịch sang ngôn ngữ thực chiến:

> Giao diện này không cố làm người dùng “wow”. Nó làm người dùng xử lý đúng việc, nhanh hơn, ít sai hơn.

---

## 3. Nguồn cảm hứng thiết kế

Hetzner có một số đặc điểm đáng lấy làm cảm hứng:

- Cấu trúc điều hướng rõ, nhóm sản phẩm theo cụm lớn.
- Các mục như Dedicated, Cloud, Web & Managed, Storage, Services được phân nhóm trực diện.
- Có nhiều card/tile product overview ngắn, ít trang trí.
- Dùng đỏ làm màu nhận diện mạnh.
- Dùng xám/trắng làm nền rất rõ.
- Tổng thể tạo cảm giác công nghiệp, kỹ thuật, đáng tin.

Chuyển sang ERP:

```text
Hetzner Website Pattern
→ ERP Admin/Productivity Pattern
```

Ví dụ:

```text
Product overview card
→ KPI card / module card / task card

Mega navigation nhóm product
→ Sidebar nhóm module ERP

Red brand accent
→ Primary action / active state / critical risk

White + grey surface
→ Dense table / form / warehouse screen
```

---

## 4. Không được làm gì

Đây là phần phải khóa sớm. ERP hay chết giao diện vì “mỗi người thêm một chút cho đẹp”.

Không dùng:

- Gradient lòe loẹt.
- Shadow nặng.
- Card bo tròn quá nhiều.
- Icon minh họa kiểu startup/cute.
- Chart màu cầu vồng.
- Primary color tràn lan khắp màn hình.
- Table quá rộng nhưng thiếu sticky header/action.
- Form quá dài không chia section.
- Modal chồng modal.
- Dashboard chỉ đẹp mà không drill-down được.
- Status chỉ hiển thị màu mà không có chữ.
- Nút đỏ cho mọi hành động.

Quy tắc nhớ:

```text
Đỏ là dao mổ, không phải sơn tường.
```

---

## 5. Design principles

### 5.1. Function first

Mỗi màn hình phải trả lời:

- Người dùng vào đây để làm gì?
- Hành động chính là gì?
- Lỗi/rủi ro quan trọng nhất là gì?
- Bước tiếp theo là gì?

Nếu một thành phần không giúp người dùng quyết định hoặc thao tác, bỏ.

### 5.2. Dense but readable

ERP cần nhiều dữ liệu. Không làm quá thoáng kiểu landing page.

Chuẩn:

- Table row 44–48px.
- Font 13–14px cho data.
- Header rõ.
- Action nằm đúng vị trí.
- Có filter nhanh.
- Có sticky header và sticky action column.

### 5.3. Status must be visible

Các trạng thái như QC HOLD, QC FAIL, thiếu tồn, cận date, thiếu đơn, chưa bàn giao phải nhìn thấy ngay.

Không để user phải mở detail mới biết rủi ro.

### 5.4. Scan-first for warehouse

Các màn kho, bàn giao ĐVVC, hàng hoàn phải ưu tiên quét mã:

- Ô scan lớn.
- Focus tự động.
- Feedback tức thì.
- Âm thanh/visual feedback nếu có.
- Không yêu cầu chuột quá nhiều.

### 5.5. Every critical action has trace

Các hành động nghiệp vụ quan trọng phải hiện rõ:

- ai tạo
- ai sửa
- ai duyệt
- lúc nào
- trạng thái trước/sau
- chứng từ đính kèm

### 5.6. ERP is not a website

Không dùng layout marketing quá nhiều hero, banner, ảnh. ERP là công cụ vận hành.

---

## 6. Color system

### 6.1. Brand-inspired palette

```text
Primary Red:              #D50C2D
Dark Grey:                #3C3C3B
Secondary Background:     #F5F5F5
Surface White:            #FFFFFF
```

### 6.2. ERP extended palette

```text
Text Primary:             #1F1F1F
Text Secondary:           #666666
Text Muted:               #8C8C8C
Border Default:           #D9D9D9
Border Soft:              #E8E8E8
Background App:           #F5F5F5
Background Section:       #FAFAFA
Surface Card:             #FFFFFF
Surface Hover:            #F7F7F7
Surface Selected:         #FFF1F3
```

### 6.3. Semantic colors

```text
Success:                  #2E7D32
Success Background:       #EAF5EC
Warning:                  #B26A00
Warning Background:       #FFF4E0
Danger:                   #D50C2D
Danger Background:        #FFE8EC
Info:                     #246BFE
Info Background:          #EAF1FF
Neutral:                  #666666
Neutral Background:       #F0F0F0
```

### 6.4. Usage rules

| Màu | Dùng cho | Không dùng cho |
|---|---|---|
| `#D50C2D` | Primary action, active nav, critical risk, danger | Background lớn, chart tràn lan |
| `#3C3C3B` | Header text, dark footer nếu cần | Sidebar quá nặng |
| `#F5F5F5` | App background | Text |
| `#FFFFFF` | Card, table, form surface | Không có |
| Success | Pass, completed, delivered | Decoration |
| Warning | QC hold, near expiry, pending | Success action |
| Danger | QC fail, missing stock, overdue | Mọi nút |

---

## 7. Typography

### 7.1. Font stack

```css
font-family: Inter, "Roboto", "Helvetica Neue", Arial, sans-serif;
```

Nếu không cài Inter thì dùng Roboto/Arial.

### 7.2. Font scale

```text
Page Title:      24px / 32px / 600
Section Title:   18px / 26px / 600
Card Title:      14px / 20px / 600
Body:            14px / 22px / 400
Table Text:      13px / 20px / 400
Helper Text:     12px / 18px / 400
Label:           13px / 20px / 500
Button:          14px / 20px / 500
```

### 7.3. Typography rules

- Không dùng quá 3 weight trong app: 400, 500, 600.
- Không dùng heading quá lớn trong ERP.
- Dữ liệu số nên căn phải trong table.
- Mã phiếu, SKU, batch có thể dùng monospace nhẹ.

```css
.code-text {
  font-family: "Roboto Mono", "SFMono-Regular", Consolas, monospace;
  font-size: 13px;
}
```

---

## 8. Spacing system

Dùng scale 4px:

```text
2px, 4px, 8px, 12px, 16px, 20px, 24px, 32px, 40px, 48px
```

Chuẩn chính:

```text
Page padding desktop:       24px
Page padding tablet:        16px
Card padding:               16px
Section gap:                16px / 24px
Form field vertical gap:    12px
Table toolbar gap:          12px
Button gap:                 8px
```

---

## 9. Shape, border, shadow

### 9.1. Radius

```text
Default radius:       4px
Small radius:         2px
Large radius:         6px
No pill unless status chip/tag needs it.
```

### 9.2. Border

```css
border: 1px solid #E5E5E5;
```

### 9.3. Shadow

Hạn chế dùng shadow. Ưu tiên border.

```css
box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
```

Dùng shadow nhẹ cho:

- dropdown
- drawer
- sticky footer
- sticky action bar

Không dùng shadow nặng cho card thường.

---

## 10. App shell template

### 10.1. Desktop layout

```text
┌─────────────────────────────────────────────────────────────┐
│ Top Bar                                                     │
│ Logo | Global Search | Quick Create | Alerts | User          │
├───────────────┬─────────────────────────────────────────────┤
│ Sidebar       │ Page Header                                  │
│               │ Title / Breadcrumb / Primary Action          │
│ Dashboard     ├─────────────────────────────────────────────┤
│ Master Data   │ Filter / Action Bar                          │
│ Purchase      ├─────────────────────────────────────────────┤
│ Inventory     │ Content Area                                 │
│ QC            │ Table / Form / Board / Detail                │
│ Production    │                                             │
│ Sales         │                                             │
│ Shipping      │                                             │
│ Returns       │                                             │
│ Reports       │                                             │
│ Settings      │                                             │
└───────────────┴─────────────────────────────────────────────┘
```

### 10.2. Sidebar

Width:

```text
Expanded: 240px
Collapsed: 72px
```

Sidebar style:

```text
Background: #FFFFFF
Border right: #E5E5E5
Active item: left border 3px #D50C2D + background #FFF1F3
Icon: line icon, 18–20px
Text: 14px, medium
```

Menu groups:

```text
Overview
- Dashboard
- Alert Center

Operations
- Warehouse Daily Board
- Inventory
- Purchase
- QC
- Production / Subcontract
- Sales Orders
- Shipping
- Returns

Data
- Master Data
- SKU / Batch
- Supplier / Factory
- Customer

Control
- Approvals
- Audit Log
- Reports
- Settings
```

### 10.3. Top bar

Height:

```text
56px
```

Content:

```text
Logo / App name
Global search
Quick create button
Alert icon
Help / Docs
User menu
```

Global search placeholder:

```text
Tìm đơn hàng, SKU, batch, phiếu nhập, vận đơn...
```

---

## 11. Page header template

```text
┌─────────────────────────────────────────────────────────────┐
│ Breadcrumb: Kho / Bàn giao ĐVVC                              │
│ Title: Bàn giao ĐVVC                                         │
│ Description: Quét đơn, đối chiếu manifest và xác nhận bàn giao│
│                                                             │
│ [Primary Action] [Secondary] [More]                          │
└─────────────────────────────────────────────────────────────┘
```

Rules:

- Title luôn rõ hành động hoặc khu vực nghiệp vụ.
- Description ngắn, không quá 1 dòng nếu có thể.
- Primary action luôn nằm bên phải.
- Không đặt quá 3 nút cấp cao ở header.

---

## 12. Button system

### 12.1. Primary button

```css
background: #D50C2D;
color: #FFFFFF;
border: 1px solid #D50C2D;
border-radius: 4px;
height: 36px;
font-weight: 500;
```

Dùng cho:

- Tạo phiếu
- Lưu
- Submit
- Xác nhận bàn giao
- Xác nhận kiểm hàng

Không dùng nhiều hơn 1 primary button trong một vùng hành động.

### 12.2. Secondary button

```css
background: #FFFFFF;
color: #3C3C3B;
border: 1px solid #D9D9D9;
```

Dùng cho:

- Xuất Excel
- Lọc nâng cao
- In phiếu
- Đính kèm

### 12.3. Danger button

```css
background: #FFFFFF;
color: #D50C2D;
border: 1px solid #D50C2D;
```

Dùng cho:

- Hủy phiếu
- Reject
- Block batch
- Báo thiếu đơn

### 12.4. Button labels

Nên dùng động từ rõ:

```text
Tạo phiếu nhập
Xác nhận kiểm hàng
Xác nhận bàn giao
Gửi duyệt
Khóa batch
In manifest
```

Không dùng label mơ hồ:

```text
OK
Submit
Action
Process
```

---

## 13. Table template

### 13.1. Table layout

```text
┌─────────────────────────────────────────────────────────────┐
│ Filter Bar                                                   │
│ [Search] [Status] [Date Range] [Warehouse] [Advanced]        │
├─────────────────────────────────────────────────────────────┤
│ Bulk Action Bar, only when rows selected                     │
├─────────────────────────────────────────────────────────────┤
│ Table                                                        │
│ Header sticky                                                │
│ Row hover                                                    │
│ Status chips                                                 │
│ Sticky action column                                         │
└─────────────────────────────────────────────────────────────┘
```

### 13.2. Table visual

```css
.table {
  background: #FFFFFF;
  border: 1px solid #E5E5E5;
}

.table th {
  background: #F5F5F5;
  color: #3C3C3B;
  font-weight: 600;
  font-size: 13px;
}

.table td {
  font-size: 13px;
  border-bottom: 1px solid #EFEFEF;
}
```

### 13.3. Standard columns

For list pages:

```text
Checkbox
Code / ID
Main Entity Name
Status
Owner / Assignee
Date
Risk / Warning
Updated At
Actions
```

For inventory:

```text
SKU | Tên hàng | Batch | HSD | Kho | Tồn vật lý | Đã giữ | Hold QC | Tồn khả dụng | Trạng thái | Action
```

For shipment:

```text
Manifest | ĐVVC | Khu vực | Tổng đơn | Đã quét | Thiếu | Trạng thái | Người phụ trách | Action
```

### 13.4. Table rules

- Mã phiếu/mã đơn là link vào detail.
- Status luôn là chip, không chỉ text.
- Số lượng căn phải.
- Tiền căn phải.
- Ngày giờ format thống nhất.
- Action không quá 3 item; phần còn lại cho vào dropdown.
- Row lỗi/rủi ro có marker nhỏ, không tô đỏ cả dòng trừ lỗi nghiêm trọng.

---

## 14. Status chip system

### 14.1. Chip style

```css
.status-chip {
  display: inline-flex;
  align-items: center;
  height: 22px;
  padding: 0 8px;
  border-radius: 3px;
  font-size: 12px;
  font-weight: 500;
  border: 1px solid transparent;
}
```

### 14.2. Common statuses

| Status | Label | Style |
|---|---|---|
| `DRAFT` | Nháp | Neutral |
| `SUBMITTED` | Chờ duyệt | Info |
| `APPROVED` | Đã duyệt | Success |
| `REJECTED` | Từ chối | Danger |
| `CANCELLED` | Đã hủy | Neutral dark |
| `PROCESSING` | Đang xử lý | Info |
| `COMPLETED` | Hoàn tất | Success |

### 14.3. Inventory/QC statuses

| Status | Label | Style |
|---|---|---|
| `QC_HOLD` | QC HOLD | Warning |
| `QC_PASS` | QC PASS | Success |
| `QC_FAIL` | QC FAIL | Danger |
| `NEAR_EXPIRY` | Cận date | Warning |
| `EXPIRED` | Hết hạn | Danger |
| `QUARANTINE` | Cách ly | Warning |
| `AVAILABLE` | Khả dụng | Success |

### 14.4. Shipping statuses

| Status | Label | Style |
|---|---|---|
| `PACKED` | Đã đóng | Info |
| `READY_TO_HANDOVER` | Chờ bàn giao | Warning |
| `HANDED_OVER` | Đã bàn giao | Success |
| `MISSING_ORDER` | Thiếu đơn | Danger |
| `DELIVERED` | Đã giao | Success |
| `RETURNED` | Hoàn | Warning |

---

## 15. Form template

### 15.1. Standard form layout

```text
┌─────────────────────────────────────────────────────────────┐
│ Page Header                                                  │
├─────────────────────────────────────────────────────────────┤
│ Section: Thông tin chung                                     │
│ [Field] [Field] [Field]                                      │
│ [Field] [Field] [Field]                                      │
├─────────────────────────────────────────────────────────────┤
│ Section: Chi tiết hàng                                       │
│ Editable line-item table                                     │
├─────────────────────────────────────────────────────────────┤
│ Section: Tệp đính kèm                                        │
├─────────────────────────────────────────────────────────────┤
│ Section: Ghi chú / Audit summary                             │
├─────────────────────────────────────────────────────────────┤
│ Sticky Footer: [Cancel] [Save Draft] [Submit]                │
└─────────────────────────────────────────────────────────────┘
```

### 15.2. Field rules

- Label trên input.
- Required field có `*` đỏ.
- Helper text ngắn dưới field.
- Error message rõ nghiệp vụ.
- Read-only field nền xám nhẹ.
- Auto-generated field có icon/tooltip.

### 15.3. Form section style

```css
.form-section {
  background: #FFFFFF;
  border: 1px solid #E5E5E5;
  border-radius: 4px;
  padding: 16px;
  margin-bottom: 16px;
}
```

### 15.4. Sticky footer

```css
.sticky-footer {
  position: sticky;
  bottom: 0;
  background: #FFFFFF;
  border-top: 1px solid #E5E5E5;
  padding: 12px 24px;
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
```

---

## 16. Detail page template

```text
┌─────────────────────────────────────────────────────────────┐
│ Header: PO-000123                         [Status Chip]      │
│ Supplier ABC | Created by User | Updated 10 phút trước        │
│ [Submit] [Approve] [Print] [More]                            │
├─────────────────────────────────────────────────────────────┤
│ Left/Main Content                                             │
│ - General Information                                         │
│ - Line Items                                                  │
│ - Related Documents                                           │
│ - Attachments                                                 │
├───────────────────────────────┬─────────────────────────────┤
│ Main Tabs                      │ Right Panel                  │
│ Details / Movements / QC       │ Approval Timeline            │
│ Audit / Notes                  │ Audit Summary                │
└───────────────────────────────┴─────────────────────────────┘
```

Rules:

- Trạng thái phải nằm ngay header.
- Action theo trạng thái.
- Timeline duyệt/audit hiển thị bên phải hoặc drawer.
- Không giấu line items sau nhiều click.

---

## 17. Modal, drawer, popover

### 17.1. Modal

Dùng cho:

- Confirm action.
- Form nhỏ dưới 5 field.
- Cảnh báo nghiệp vụ.

Không dùng modal cho form dài.

### 17.2. Drawer

Dùng cho:

- Detail nhanh.
- Audit log.
- Approval history.
- Scan result detail.
- Attachment preview.

### 17.3. Confirmation copy

Ví dụ:

```text
Xác nhận bàn giao manifest GHN-260426-01?
Hệ thống sẽ ghi nhận 128 đơn đã bàn giao cho GHN và khóa thao tác đóng gói với các đơn này.
```

Không dùng:

```text
Are you sure?
```

---

## 18. Empty, loading, error state

### 18.1. Empty state

```text
Chưa có phiếu nhập nào
Tạo phiếu nhập đầu tiên để bắt đầu nhận hàng vào kho.
[ Tạo phiếu nhập ]
```

### 18.2. Loading state

- Table: skeleton row.
- Detail: skeleton block.
- Scan: không dùng skeleton lâu; hiển thị “Đang kiểm tra mã...” ngắn.

### 18.3. Error state

ERP error phải nói nghiệp vụ, không nói kỹ thuật.

Ví dụ:

```text
Không thể xuất kho
Batch này đang ở trạng thái QC HOLD, chưa được phép xuất bán.
```

Không dùng:

```text
Error 400
```

---

## 19. Dashboard template

### 19.1. Executive dashboard

```text
Tổng quan vận hành
──────────────────────────────────────────────────────────────
Ngày: 26/04/2026      Kho: Tất cả      Brand: Tất cả

[KPI Card] Đơn cần xử lý hôm nay        236
[KPI Card] Đơn chờ bàn giao             48
[KPI Card] Hàng hoàn chờ kiểm           12
[KPI Card] Batch đang HOLD              5
[KPI Card] Tồn cận date                 128 SKU
[KPI Card] Đơn thiếu tồn                9

Cảnh báo vận hành
──────────────────────────────────────────────────────────────
Mức độ | Cảnh báo | Module | Owner | Deadline | Action

Luồng hôm nay
──────────────────────────────────────────────────────────────
Đơn mới → Đang soạn → Đã đóng → Chờ bàn giao → Đã bàn giao
```

### 19.2. KPI card style

```text
┌─────────────────────────────┐
│ Đơn chờ bàn giao             │
│ 48                           │
│ +12 so với hôm qua           │
└─────────────────────────────┘
```

Visual:

- Border 1px.
- Không shadow lớn.
- Số chính 28px/600.
- Risk indicator nhỏ ở góc phải.
- Card có thể click để drill-down.

---

## 20. Warehouse Daily Board template

Workflow kho thực tế có chuỗi: tiếp nhận đơn hàng trong ngày, thực hiện xuất/nhập theo nội quy, soạn và đóng gói, sắp xếp/tối ưu vị trí kho, kiểm kê cuối ngày, đối soát số liệu/báo cáo và kết thúc ca. Vì vậy màn này phải là trung tâm vận hành kho mỗi ngày.

### 20.1. Screen layout

```text
Warehouse Daily Board
──────────────────────────────────────────────────────────────
Ngày: 26/04/2026    Ca: Sáng    Kho: Tổng    Người phụ trách: A

[Đơn mới: 236] [Đang soạn: 120] [Đã đóng: 88] [Chờ bàn giao: 48]
[Hàng hoàn: 12] [Lệch tồn: 3] [Task quá hạn: 5]

Task Board
──────────────────────────────────────────────────────────────
Loại task | Mã liên quan | Trạng thái | Người phụ trách | SLA | Action

End-of-day Closing
──────────────────────────────────────────────────────────────
[Kiểm kê cuối ngày] [Đối soát số liệu] [Gửi báo cáo] [Kết thúc ca]
```

### 20.2. UX rules

- User kho vào app là thấy Daily Board trước.
- Các task thiếu đơn/lệch tồn/quá SLA phải nổi lên trên.
- Closing cuối ngày không được ẩn trong menu sâu.
- Không cho kết thúc ca nếu còn exception P0/P1 chưa xử lý hoặc chưa có note.

---

## 21. Receiving / nhập kho template

### 21.1. List page

```text
Phiếu nhập kho
──────────────────────────────────────────────────────────────
[Search] [NCC] [Kho] [Trạng thái QC] [Ngày nhận] [Tạo phiếu nhập]

Mã phiếu | NCC | PO | Kho | Tổng dòng | QC | Trạng thái | Ngày nhận | Action
```

### 21.2. Create page

```text
Tạo phiếu nhập kho
──────────────────────────────────────────────────────────────
Thông tin chung
- NCC
- PO tham chiếu
- Kho nhận
- Ngày nhận
- Người nhận

Chi tiết hàng
SKU | Tên hàng | Batch | HSD | Số lượng | Tình trạng bao bì | QC yêu cầu | Ghi chú

Tệp đính kèm
- chứng từ giao hàng
- ảnh kiện hàng
- COA/MSDS nếu có

[Save Draft] [Submit QC]
```

### 21.3. Rules

- Batch/HSD bắt buộc với hàng có quản lý lô.
- Nếu QC chưa pass, hàng không vào tồn khả dụng.
- Nếu số lượng/bao bì/lô không đạt, phải có reason và evidence.

---

## 22. Outbound / xuất kho template

### 22.1. List page

```text
Phiếu xuất kho
──────────────────────────────────────────────────────────────
[Search] [Kho] [Loại xuất] [Trạng thái] [Ngày] [Tạo phiếu xuất]

Mã phiếu | Loại | Kho | Đơn liên quan | Tổng SKU | Trạng thái | Người tạo | Action
```

### 22.2. Detail page

```text
Phiếu xuất PX-000123                    [Chờ duyệt]
──────────────────────────────────────────────────────────────
Thông tin chung
Loại xuất: Sales / Production / Transfer / Sample / Adjustment
Kho xuất
Người tạo
Lý do

Line Items
SKU | Batch | HSD | Số lượng yêu cầu | Số lượng thực xuất | Tồn khả dụng | Status

Audit / Approval
```

### 22.3. Rules

- Không cho xuất batch QC HOLD/FAIL.
- Không cho xuất vượt tồn khả dụng.
- Thực xuất khác yêu cầu phải có reason.
- Phiếu đã hoàn tất không được sửa trực tiếp.

---

## 23. Packing template

Workflow đóng hàng thực tế gồm nhận phiếu đơn hợp lệ, lọc/phân loại theo ĐVVC/đơn lẻ/đơn sàn, soạn từng đơn, kiểm tra tại khu vực đóng hàng, đếm tổng số đơn mỗi sàn, chuyển đến khu vực bàn giao ĐVVC và ký xác nhận.

### 23.1. Packing list

```text
Đóng hàng
──────────────────────────────────────────────────────────────
[Search đơn] [ĐVVC] [Sàn/Kênh] [Khu vực] [Trạng thái]

Mã đơn | Kênh | ĐVVC | SKU | Batch | Trạng thái soạn | Trạng thái đóng | Action
```

### 23.2. Packing detail

```text
Đóng đơn SO-000123
──────────────────────────────────────────────────────────────
Thông tin đơn
Khách hàng | Kênh | ĐVVC | COD | Ghi chú

Checklist
[ ] Đúng SKU
[ ] Đúng batch nếu có
[ ] Đủ quà tặng/combo
[ ] Đóng gói đúng quy cách
[ ] Dán mã vận đơn

Line Items
SKU | Batch | SL | Đã pick | Đã pack

[ Xác nhận đóng hàng ]
```

### 23.3. Visual rules

- Màn đóng hàng ưu tiên thao tác nhanh.
- Nút xác nhận rõ.
- Nếu thiếu SKU/batch/quà tặng phải báo ngay.

---

## 24. Shipping handover scan template

Quy trình bàn giao thực tế có phân chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, lấy hàng và quét mã trực tiếp. Nếu đủ đơn thì ký xác nhận bàn giao; nếu chưa đủ thì kiểm tra lại mã hoặc tìm trong khu vực đóng hàng.

### 24.1. Screen layout

```text
Bàn giao ĐVVC
──────────────────────────────────────────────────────────────
Manifest: GHN-260426-01        Trạng thái: Đang quét
ĐVVC: GHN                      Khu vực: A3
Tổng đơn: 128                  Đã quét: 126
Còn thiếu: 2                   Người phụ trách: Nguyễn A

┌────────────────────────────────────────────────────────────┐
│ Quét mã vận đơn / mã đơn                                   │
│ [____________________________________________________]      │
└────────────────────────────────────────────────────────────┘

Danh sách thiếu
──────────────────────────────────────────────────────────────
Mã đơn | Mã vận đơn | Khu vực đóng | Trạng thái | Action
SO-000123 | GHN123 | A3/Rổ 02 | Chưa quét | [Tìm đơn]
SO-000147 | GHN147 | A3/Rổ 04 | Chưa quét | [Báo thiếu]

[In manifest] [Báo thiếu đơn] [Xác nhận bàn giao]
```

### 24.2. Scan feedback

| Result | UI feedback |
|---|---|
| Mã hợp lệ | Viền xanh, âm beep ngắn, tăng count |
| Mã đã quét | Warning, không tăng count |
| Mã không thuộc manifest | Danger, yêu cầu kiểm tra |
| Mã thuộc manifest khác | Danger + link manifest liên quan |
| Đủ tất cả đơn | Enable nút xác nhận bàn giao |

### 24.3. Rules

- Không cho xác nhận bàn giao nếu còn thiếu đơn, trừ khi có quyền override và reason.
- Override phải ghi audit log.
- Manifest đã bàn giao không được sửa danh sách đơn trực tiếp.

---

## 25. Return inspection / hàng hoàn template

Nội quy hàng hoàn thực tế có nhánh: nhận hàng từ shipper, đưa vào khu vực hàng hoàn, quét hàng hoàn, kiểm tra tình trạng, ghi tình trạng thực tế, sau đó phân loại còn sử dụng hoặc không sử dụng; hàng còn dùng chuyển vào kho, hàng không dùng chuyển lên lab.

### 25.1. Screen layout

```text
Kiểm hàng hoàn
──────────────────────────────────────────────────────────────
Mã đơn / mã vận đơn
[____________________________________________________] [Quét]

Thông tin đơn
──────────────────────────────────────────────────────────────
Mã đơn: SO-000123
Khách hàng: Nguyễn B
Kênh: Website
SKU: Serum X 30ml
Batch: SVC-260426-01
Ngày giao: 24/04/2026
Lý do hoàn: Khách từ chối nhận

Tình trạng hàng
──────────────────────────────────────────────────────────────
( ) Nguyên vẹn
( ) Móp hộp
( ) Rách seal
( ) Đã sử dụng
( ) Hỏng / đổ vỡ
( ) Không xác định, cần QA kiểm

Kết luận
[ Còn sử dụng ] [ Không sử dụng ] [ Cần QA kiểm tra ]

Ảnh / Ghi chú
[Upload ảnh] [Ghi chú]

[ Xác nhận kiểm hàng hoàn ]
```

### 25.2. Rules

- Hàng hoàn chưa kiểm không được nhập lại tồn khả dụng.
- Nếu còn sử dụng, phải tạo movement về kho phù hợp.
- Nếu không sử dụng, chuyển lab/kho hỏng/scrap theo rule.
- Nếu cần QA, trạng thái là `RETURN_QA_HOLD`.

---

## 26. QC batch template

### 26.1. QC list

```text
QC Batch
──────────────────────────────────────────────────────────────
[Search batch] [SKU] [Kho] [Trạng thái QC] [Ngày nhận]

Batch | SKU | HSD | Nguồn | SL | QC Status | Deadline | Action
```

### 26.2. QC detail

```text
QC Batch SVC-260426-01               [QC HOLD]
──────────────────────────────────────────────────────────────
Thông tin batch
SKU | Batch | NSX | HSD | Nguồn nhập | Số lượng

Checklist QC
[ ] Ngoại quan
[ ] Bao bì
[ ] Mùi/màu/trạng thái
[ ] Chứng từ COA
[ ] Test nội bộ nếu có

Kết quả
( ) PASS
( ) FAIL
( ) HOLD thêm

Evidence
Ảnh | File COA | Ghi chú

[ Lưu kết quả ] [ Release batch ] [ Fail batch ]
```

### 26.3. Rules

- `Release batch` là critical action.
- QC FAIL phải có reason.
- QC PASS tạo event và cập nhật tồn khả dụng nếu rule cho phép.

---

## 27. Subcontract manufacturing template

Quy trình sản xuất thực tế của công ty có nhánh gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn/chốt thời gian, chuyển nguyên vật liệu/bao bì, ký bàn giao kèm chứng từ, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao hàng về kho, kiểm tra số lượng/chất lượng, nhận hàng hoặc báo lỗi nhà máy trong 3–7 ngày.

### 27.1. List page

```text
Lệnh gia công
──────────────────────────────────────────────────────────────
[Search] [Nhà máy] [Sản phẩm] [Trạng thái] [Ngày giao dự kiến]

Mã lệnh | Nhà máy | Sản phẩm | SL đặt | Trạng thái | ETA | QC | Thanh toán | Action
```

### 27.2. Detail page

```text
Lệnh gia công SC-000123                    [Sample Approved]
──────────────────────────────────────────────────────────────
Nhà máy: ABC Factory
Sản phẩm: Serum X 30ml
Số lượng: 5,000
Ngày đặt: 10/04/2026
Ngày giao dự kiến: 30/04/2026

Timeline
[Đặt đơn] → [Cọc] → [Gửi NVL/BB] → [Duyệt mẫu] → [SX hàng loạt] → [Nhận hàng] → [QC] → [Đóng]

Tabs
1. Thông tin đơn
2. NVL/bao bì gửi đi
3. Biên bản bàn giao
4. Mẫu duyệt
5. Lịch giao
6. Nhận hàng & QC
7. Claim nhà máy
8. Thanh toán
```

### 27.3. Sample approval block

```text
Duyệt mẫu
──────────────────────────────────────────────────────────────
Mẫu số | Ngày nhận | Người kiểm | Kết quả | Ảnh/File | Ghi chú
M01    | 14/04     | QA A       | Fail    | ...      | Mùi chưa đạt
M02    | 16/04     | QA A       | Pass    | ...      | Chốt mẫu

[ Tạo mẫu mới ] [ Chốt mẫu ]
```

### 27.4. Factory claim block

```text
Claim nhà máy
──────────────────────────────────────────────────────────────
Loại lỗi | Số lượng ảnh hưởng | Ngày phát hiện | Deadline phản hồi | Status

[ Tạo claim ]
```

Rules:

- Claim sau khi nhận hàng phải bám SLA 3–7 ngày.
- Không đóng lệnh nếu claim P0/P1 chưa xử lý hoặc chưa có quyết định override.

---

## 28. Approval page template

```text
Phê duyệt
──────────────────────────────────────────────────────────────
[Search] [Module] [Trạng thái] [Mức độ rủi ro] [Người yêu cầu]

Yêu cầu | Module | Người yêu cầu | Giá trị/Rủi ro | Deadline | Action

Detail drawer:
- Nội dung yêu cầu
- Dữ liệu trước/sau
- Attachment
- Comment
- Approve / Reject
```

Rules:

- Approver phải thấy đủ ngữ cảnh, không duyệt mù.
- Reject bắt buộc reason.
- Approval action ghi audit log.

---

## 29. Audit log panel template

```text
Audit Log
──────────────────────────────────────────────────────────────
Thời gian | Người dùng | Hành động | Trước | Sau | IP/Device

Filters:
- Action type
- User
- Date
- Field changed
```

Visual:

- Drawer bên phải.
- Diff trước/sau dùng bảng 2 cột.
- Dữ liệu nhạy cảm có masking theo quyền.

---

## 30. Attachment panel template

```text
Tệp đính kèm
──────────────────────────────────────────────────────────────
Tên file | Loại | Upload bởi | Ngày | Action

[Upload file]
```

File types:

- PDF
- Image
- Excel
- COA/MSDS
- Biên bản bàn giao
- Ảnh hàng lỗi
- Ảnh hàng hoàn
- Hợp đồng/PO nếu cần

Rules:

- Critical evidence không được xóa trực tiếp, chỉ archive theo quyền.
- File phải gắn module/entity.

---

## 31. Login page template

```text
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                 [Company Logo / ERP]                        │
│                                                             │
│                 Đăng nhập ERP                               │
│                 Email                                      │
│                 Password                                   │
│                 [ Đăng nhập ]                               │
│                                                             │
│                 Quên mật khẩu?                              │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

Visual:

- Nền `#F5F5F5`.
- Card trắng, border mảnh.
- Primary button đỏ.
- Không ảnh minh họa phức tạp.

---

## 32. Master Data page template

```text
Master Data / SKU
──────────────────────────────────────────────────────────────
[Search SKU] [Loại hàng] [Trạng thái] [Brand] [Tạo SKU]

SKU | Tên hàng | Loại | Đơn vị | Quản lý batch | Quản lý HSD | Trạng thái | Action
```

Detail:

```text
SKU: SERUM-X-30ML
Tabs:
- Thông tin cơ bản
- Quy đổi đơn vị
- Batch rules
- Pricing rules
- Attachment
- Audit log
```

Rules:

- Batch/HSD rule phải rõ ngay trong SKU detail.
- Không cho active SKU thiếu dữ liệu bắt buộc.

---

## 33. Report page template

```text
Báo cáo tồn kho
──────────────────────────────────────────────────────────────
[Kho] [SKU] [Batch] [HSD] [QC Status] [Export]

Summary cards
- Tổng tồn
- Tồn khả dụng
- Tồn QC HOLD
- Tồn cận date

Table
SKU | Batch | HSD | Kho | Tồn vật lý | Đã giữ | Hold QC | Khả dụng | Giá trị ước tính
```

Rules:

- Report phải drill-down về chứng từ gốc nếu có thể.
- Export theo quyền.
- Các số nhạy cảm như giá vốn ẩn theo role.

---

## 34. Responsive behavior

### 34.1. Breakpoints

```text
Desktop:       >= 1280px
Laptop:        1024–1279px
Tablet:        768–1023px
Mobile:        < 768px
```

### 34.2. Desktop

- Sidebar expanded.
- Table full.
- Detail with right panel.

### 34.3. Tablet warehouse mode

- Sidebar collapsed.
- Scan input lớn.
- Action buttons lớn hơn.
- Table có ít cột hơn.

### 34.4. Mobile

Phase 1 không ưu tiên mobile full ERP.

Mobile chỉ nên hỗ trợ:

- scan nhanh
- xem task
- xác nhận đơn giản
- xem cảnh báo

Không làm form nghiệp vụ dài trên mobile nếu chưa cần.

---

## 35. Keyboard and scan behavior

### 35.1. Keyboard shortcuts

```text
Ctrl/Cmd + K: Global search
Ctrl/Cmd + S: Save draft nếu form dirty
Esc: Close drawer/modal nếu không có unsaved change
Enter: Submit scan input
Tab: Move field
```

### 35.2. Scan input rules

- Auto focus khi vào màn scan.
- Sau mỗi scan thành công, clear input.
- Scan lỗi giữ input và select text.
- Có visual feedback trong dưới 200ms.
- Không bắt người dùng bấm nút sau mỗi scan.

---

## 36. Ant Design theme tokens

Nếu dùng Ant Design, cấu hình token gợi ý:

```ts
export const erpTheme = {
  token: {
    colorPrimary: '#D50C2D',
    colorText: '#1F1F1F',
    colorTextSecondary: '#666666',
    colorBgLayout: '#F5F5F5',
    colorBgContainer: '#FFFFFF',
    colorBorder: '#D9D9D9',
    borderRadius: 4,
    fontSize: 14,
    controlHeight: 36,
  },
  components: {
    Button: {
      borderRadius: 4,
      controlHeight: 36,
    },
    Table: {
      headerBg: '#F5F5F5',
      headerColor: '#3C3C3B',
      rowHoverBg: '#F7F7F7',
      fontSize: 13,
    },
    Card: {
      borderRadiusLG: 4,
      paddingLG: 16,
    },
    Tag: {
      borderRadiusSM: 3,
    },
    Input: {
      controlHeight: 36,
    },
    Select: {
      controlHeight: 36,
    },
  },
};
```

---

## 37. CSS variables

```css
:root {
  --erp-red: #D50C2D;
  --erp-dark-grey: #3C3C3B;
  --erp-bg: #F5F5F5;
  --erp-surface: #FFFFFF;
  --erp-border: #D9D9D9;
  --erp-border-soft: #E8E8E8;
  --erp-text: #1F1F1F;
  --erp-text-secondary: #666666;
  --erp-text-muted: #8C8C8C;

  --erp-success: #2E7D32;
  --erp-success-bg: #EAF5EC;
  --erp-warning: #B26A00;
  --erp-warning-bg: #FFF4E0;
  --erp-danger: #D50C2D;
  --erp-danger-bg: #FFE8EC;
  --erp-info: #246BFE;
  --erp-info-bg: #EAF1FF;

  --erp-radius: 4px;
  --erp-page-padding: 24px;
  --erp-card-padding: 16px;
  --erp-topbar-height: 56px;
  --erp-sidebar-width: 240px;
}
```

---

## 38. Page templates to build first

Thứ tự làm template trong Figma/code:

```text
1. Login Page
2. App Shell
3. Dashboard / Operations Overview
4. List/Table Page
5. Detail Page
6. Create/Edit Form Page
7. Warehouse Daily Board
8. Shipping Handover Scan
9. Return Inspection
10. QC Batch Detail
11. Subcontract Manufacturing Detail
12. Approval Inbox
13. Audit Log Drawer
14. Empty/Error/Loading State
```

Đây là bộ minimum visual system để frontend không phải bịa giao diện từng màn.

---

## 39. Figma frame checklist

Nếu làm Figma, cần tạo các frame:

```text
00 Cover / Design Direction
01 Color Tokens
02 Typography
03 App Shell
04 Buttons
05 Inputs / Select / Date Picker
06 Table
07 Status Chips
08 Card / KPI Card
09 Modal / Drawer
10 Empty / Loading / Error
11 Login
12 Dashboard
13 Warehouse Daily Board
14 Receiving Form
15 Outbound Form
16 Packing Page
17 Shipping Handover Scan
18 Return Inspection
19 QC Batch
20 Subcontract Manufacturing
21 Approval Inbox
22 Audit Log Drawer
```

---

## 40. Frontend component naming

Suggested structure:

```text
shared/design-system/
  Button/
  StatusChip/
  DataTable/
  PageHeader/
  FilterBar/
  FormSection/
  StickyFooter/
  AuditDrawer/
  AttachmentPanel/
  ScanInput/
  KpiCard/
  EmptyState/
  ErrorState/

modules/warehouse/components/
  WarehouseDailyBoard/
  PackingTaskTable/
  ShiftClosingPanel/

modules/shipping/components/
  ManifestSummary/
  HandoverScanPanel/
  MissingOrderTable/

modules/returns/components/
  ReturnInspectionForm/
  ReturnConditionSelector/

modules/qc/components/
  QCBatchChecklist/
  QCDecisionPanel/

modules/production/components/
  SubcontractTimeline/
  MaterialTransferTable/
  SampleApprovalTable/
  FactoryClaimPanel/
```

Rules:

- Shared components không chứa nghiệp vụ riêng.
- Business components nằm trong module.
- Không copy table/form component mỗi module.
- StatusChip phải dùng chung token.

---

## 41. UI microcopy direction

Tone:

```text
Rõ, ngắn, nghiệp vụ, không vòng vo.
```

Ví dụ tốt:

```text
Batch này đang QC HOLD, chưa được phép xuất bán.
```

Ví dụ không tốt:

```text
Thao tác thất bại.
```

Thông báo scan:

```text
Đã quét đơn SO-000123.
Đơn này đã được quét trước đó.
Mã này không thuộc manifest hiện tại.
Manifest còn thiếu 2 đơn, chưa thể bàn giao.
```

Thông báo hàng hoàn:

```text
Hàng hoàn đã được phân loại: Còn sử dụng.
Hàng hoàn cần QA kiểm tra trước khi nhập lại kho.
```

---

## 42. Accessibility baseline

- Text contrast đạt mức dễ đọc.
- Không dùng màu là tín hiệu duy nhất; status phải có chữ.
- Focus state rõ cho input/button/scan.
- Form error liên kết với field.
- Keyboard usable cho form/table/action cơ bản.
- Button icon-only phải có tooltip/aria-label.
- Modal/drawer trap focus đúng.

---

## 43. Handoff rules for designer → frontend

Designer phải bàn giao:

- Figma frame theo template.
- Token màu/font/spacing.
- Component variants.
- Table columns.
- Form states.
- Empty/loading/error state.
- Responsive tablet state cho scan/warehouse.
- Interaction notes cho approval, scan, handover, return.
- Không chỉ bàn giao màn hình “happy path”.

Frontend phải xác nhận:

- Token đã map vào Ant Design/custom CSS.
- Component dùng lại được.
- API state loading/error được xử lý.
- Permission/hidden action đã kiểm tra.
- Audit/attachment panel có pattern chung.

---

## 44. Definition of Done for UI template implementation

Một template được coi là xong khi:

```text
[ ] Có layout desktop.
[ ] Có layout tablet nếu liên quan kho/scan.
[ ] Có loading state.
[ ] Có empty state.
[ ] Có error state.
[ ] Có permission/disabled state.
[ ] Có status chip đúng màu/chữ.
[ ] Có primary/secondary/danger action đúng rule.
[ ] Có audit/attachment vị trí chuẩn nếu nghiệp vụ cần.
[ ] Có validation message rõ nghiệp vụ.
[ ] Có keyboard/focus behavior cơ bản.
[ ] Có link tới route/module liên quan.
[ ] Được PO/BA xác nhận không lệch nghiệp vụ.
```

---

## 45. Priority screen build order

Nếu bắt đầu code UI thật, làm theo thứ tự:

```text
P0. App Shell + Login
P0. DataTable / PageHeader / FilterBar / StatusChip / FormSection / StickyFooter
P0. Master Data List + Detail Template
P0. Inventory Stock Ledger List
P0. Warehouse Daily Board
P0. Shipping Handover Scan
P0. Return Inspection
P1. Receiving Form
P1. QC Batch Detail
P1. Sales Order List/Detail
P1. Packing Screen
P1. Subcontract Manufacturing Detail
P2. Reports
P2. Advanced Dashboard
```

Lý do: các template P0 là nền dùng lại cho toàn bộ ERP và bám sát workflow kho/giao hàng/hàng hoàn thực tế.

---

## 46. Design QA checklist

Trước khi release UI:

```text
[ ] Màu primary red không bị dùng quá mức.
[ ] Table đủ cột chính và có sticky header/action.
[ ] Batch/QC/HSD/rủi ro hiển thị rõ.
[ ] Scan input auto focus đúng.
[ ] Empty/error/loading states đầy đủ.
[ ] Hàng hoàn có phân loại còn dùng/không dùng/cần QA.
[ ] Bàn giao ĐVVC có count tổng/đã quét/còn thiếu.
[ ] Warehouse Daily Board có task và shift closing.
[ ] Gia công ngoài có timeline và tabs đúng nghiệp vụ.
[ ] Action nguy hiểm có confirm và reason nếu cần.
[ ] Role không có quyền không thấy hoặc không bấm được action.
[ ] Export/cost sensitive data bị ẩn theo quyền.
```

---

## 47. Final design direction summary

Giao diện ERP này nên nhìn như:

```text
Một hệ thống vận hành công nghiệp sạch, nhanh, chắc.
Không màu mè.
Không rối.
Không trang trí thừa.
Mỗi màn hình giúp người dùng làm đúng việc tiếp theo.
```

Chốt style:

```text
Hetzner-inspired minimalism
+ ERP productivity density
+ warehouse scan-first UX
+ cosmetics batch/QC discipline
+ subcontract manufacturing visibility
```

Một câu để team nhớ:

> UI tốt trong ERP không phải là UI đẹp nhất. UI tốt là UI khiến người dùng ít sai hơn, xử lý nhanh hơn, và để lại dấu vết rõ hơn.

---

## 48. Related documents

Primary related docs:

```text
08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md
14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md
15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md
20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md
21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md
37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md
38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md
```

Input workflow references:

```text
Công-việc-hằng-ngày.pdf
Nội-Quy.pdf
Quy-trình-bàn-giao.pdf
Quy-trình-sản-xuất.pdf
```
