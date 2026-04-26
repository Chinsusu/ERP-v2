# 15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Phase:** Phase 1  
**Frontend stack:** React / Next.js + TypeScript  
**Backend liên kết:** Go Modular Monolith + REST/OpenAPI  
**Database backend:** PostgreSQL  
**Tài liệu liên quan:**

- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt kiến trúc frontend cho ERP Phase 1.

Mục tiêu không phải chỉ là chọn React hay Next.js, mà là chốt cách tổ chức frontend sao cho:

- Dễ mở rộng nhiều module ERP.
- Dễ bảo trì khi nghiệp vụ thay đổi.
- UI nhất quán giữa kho, mua hàng, QA/QC, sản xuất, bán hàng, giao hàng, hàng hoàn.
- Bám đúng backend Go qua REST/OpenAPI.
- Tránh mỗi dev frontend tự viết một kiểu.
- Hạn chế sai dữ liệu trong các thao tác nhạy cảm như tồn kho, batch, QC, bàn giao ĐVVC, hàng hoàn, gia công ngoài.

Frontend ERP không phải website marketing. Đây là **hệ điều hành nghiệp vụ trên trình duyệt**.

---

## 2. Workflow thực tế cần frontend bám sát

Frontend phải bám các workflow As-Is đã thu thập từ tài liệu nội bộ:

1. Kho xử lý công việc hằng ngày theo nhịp: tiếp nhận đơn trong ngày, xuất/nhập theo nội quy, soạn và đóng gói, tối ưu vị trí kho, kiểm kê cuối ngày, đối soát số liệu, báo cáo quản lý, kết thúc ca.
2. Nội quy kho tách rõ quy trình nhập kho, xuất kho, đóng hàng và xử lý hàng hoàn.
3. Bàn giao ĐVVC có quy trình chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, lấy hàng và quét mã, ký xác nhận bàn giao; nếu chưa đủ đơn phải kiểm tra lại mã hoặc tìm lại trong khu vực đóng hàng.
4. Sản xuất hiện có nhánh gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao hàng về kho, kiểm tra số lượng/chất lượng, nhận hàng hoặc báo lỗi nhà máy trong 3–7 ngày.

Vì vậy frontend Phase 1 không được thiết kế kiểu CRUD chung chung. Phải ưu tiên các màn hình thao tác nhanh, ít nhầm, có quét mã, có trạng thái rõ, có audit, có phê duyệt, có cảnh báo theo batch/hạn dùng/QC.

---

## 3. Tech stack frontend chốt đề xuất

### 3.1. Core stack

```text
Language: TypeScript
Framework: Next.js
UI Runtime: React
Package manager: pnpm
API contract: OpenAPI generated types/client
Server state: TanStack Query
Client/UI state: Zustand hoặc React Context theo phạm vi nhỏ
Form: React Hook Form + Zod
Table: Ant Design Table / ProTable hoặc TanStack Table tùy độ phức tạp
UI component base: Ant Design + custom theme tokens
Chart: Recharts hoặc ECharts
Date/time: dayjs
Testing: Vitest + React Testing Library + Playwright
Lint/format: ESLint + Prettier
```

### 3.2. Vì sao chọn hướng này

**Next.js + React + TypeScript** giúp frontend có cấu trúc route rõ, type-safety tốt, dễ mở rộng module.  
**Ant Design** phù hợp ERP vì có sẵn table, form, modal, drawer, upload, date picker, pagination, layout, notification.  
**TanStack Query** giúp quản lý dữ liệu server tốt, cache tốt, invalidate đúng sau thao tác nghiệp vụ.  
**React Hook Form + Zod** giúp form lớn nhẹ hơn, validation rõ hơn, dễ test hơn.  
**OpenAPI generated client** giúp frontend và backend Go nói cùng một hợp đồng, giảm lỗi tự đoán DTO.

### 3.3. Nguyên tắc khi dùng Ant Design

Không dùng Ant Design như template mặc định rồi bỏ mặc UI.

Phải customize theo design system của ERP:

- Token màu theo trạng thái nghiệp vụ.
- Table density phù hợp màn hình kho/sales/admin.
- Button/action hierarchy rõ.
- Form layout chuẩn.
- Status chip nhất quán.
- Drawer detail chuẩn.
- Modal xác nhận cho hành động rủi ro.
- Không dùng component tùy tiện ngoài design system nếu chưa được duyệt.

---

## 4. Kiến trúc frontend tổng thể

```text
Browser
  ↓
Next.js Web App
  ↓
Route Layer
  ↓
Page Container
  ↓
Module Feature Components
  ↓
Shared UI Components
  ↓
API Client / Query Layer
  ↓
Go Backend REST API
```

Frontend chia thành 5 lớp:

```text
1. Route Layer
2. Page Container Layer
3. Feature Component Layer
4. Shared Component / Design System Layer
5. API / Query / State Layer
```

Không để page route chứa toàn bộ logic nghiệp vụ.

---

## 5. Cấu trúc thư mục đề xuất

```text
apps/
  web/
    src/
      app/
        (auth)/
          login/
            page.tsx
        (erp)/
          layout.tsx
          dashboard/
            page.tsx
          master-data/
            items/
              page.tsx
          purchase/
            requests/
              page.tsx
            orders/
              page.tsx
          inventory/
            receipts/
              page.tsx
            stock-ledger/
              page.tsx
            batch-trace/
              page.tsx
            cycle-count/
              page.tsx
            shift-closing/
              page.tsx
          qc/
            inspections/
              page.tsx
            batch-release/
              page.tsx
          production/
            subcontract-orders/
              page.tsx
            material-handover/
              page.tsx
            sample-approval/
              page.tsx
          sales/
            orders/
              page.tsx
            returns/
              page.tsx
          shipping/
            pick-pack/
              page.tsx
            manifests/
              page.tsx
            handover/
              page.tsx

      modules/
        master-data/
          components/
          hooks/
          api/
          schemas/
          types/
          pages/
        purchase/
        inventory/
        qc/
        production/
        sales/
        shipping/
        returns/
        finance/

      shared/
        api/
          client.ts
          generated/
          queryKeys.ts
        auth/
          AuthProvider.tsx
          permissions.ts
          routeGuard.tsx
        components/
          AppShell/
          DataTable/
          EntityDrawer/
          StatusChip/
          ApprovalPanel/
          AuditTimeline/
          AttachmentPanel/
          BarcodeScanInput/
          ConfirmActionModal/
          EmptyState/
          ErrorState/
        design-system/
          tokens.ts
          theme.ts
          status.ts
        hooks/
          useDebounce.ts
          useHotkeys.ts
          usePermission.ts
          useScanBuffer.ts
        utils/
          formatDate.ts
          formatMoney.ts
          formatNumber.ts
          validators.ts
        constants/
          routes.ts
          permissions.ts
          statuses.ts
```

### 5.1. Quy tắc tổ chức thư mục

- `app/` chỉ giữ route, layout và page entry.
- `modules/` chứa logic UI theo module nghiệp vụ.
- `shared/` chỉ chứa component/hook/util dùng chung thật sự.
- Không để module này import sâu vào component nội bộ của module khác.
- Nếu cần dùng chung, đưa lên `shared` hoặc tạo public export của module.
- Không để business rule quan trọng chỉ tồn tại ở frontend. Backend vẫn là nguồn quyết định cuối cùng.

---

## 6. Route architecture

### 6.1. Nhóm route chính

```text
/login
/dashboard
/master-data/*
/purchase/*
/inventory/*
/qc/*
/production/*
/sales/*
/shipping/*
/returns/*
/finance/*
/settings/*
```

### 6.2. Route Phase 1 bắt buộc

```text
/dashboard

/master-data/items
/master-data/skus
/master-data/suppliers
/master-data/customers
/master-data/warehouses
/master-data/batches

/purchase/requests
/purchase/orders
/purchase/inbound-schedules

/inventory/receipts
/inventory/issues
/inventory/transfers
/inventory/stock-ledger
/inventory/batch-trace
/inventory/cycle-count
/inventory/shift-closing

/qc/inbound-inspections
/qc/batch-release
/qc/nonconformance

/production/subcontract-orders
/production/material-handover
/production/sample-approval
/production/receipts

/sales/orders
/sales/returns

/shipping/pick-pack
/shipping/manifests
/shipping/handover
/shipping/delivery-status
```

### 6.3. Route naming rule

- URL dùng kebab-case.
- Module route dùng danh từ số nhiều khi là list.
- Action không nên tạo URL riêng nếu chỉ là modal/drawer.
- Action nghiệp vụ phức tạp có thể có route riêng.

Ví dụ:

```text
/inventory/shift-closing
/shipping/handover
/production/material-handover
```

Không dùng:

```text
/inventory/doCloseShift
/shipping/shipNow
```

---

## 7. Page pattern chuẩn

Mỗi màn hình nghiệp vụ chính nên đi theo pattern:

```text
List Page
  → Filter Bar
  → Data Table
  → Row Actions
  → Detail Drawer
  → Create/Edit Drawer hoặc Full Form Page
  → Approval Panel
  → Audit Timeline
  → Attachment Panel
```

### 7.1. List Page

Bắt buộc có:

- Title rõ.
- Module context.
- Filter bar.
- Saved filters nếu màn hình dùng nhiều.
- Table có pagination/sort/filter.
- Bulk action nếu được phép.
- Export nếu role được phép.
- Empty state.
- Error state.
- Loading/skeleton.

### 7.2. Detail Drawer

Dùng cho xem nhanh:

- Thông tin chính.
- Trạng thái.
- Line items.
- Lịch sử duyệt.
- Audit timeline.
- File đính kèm.
- Action hợp lệ theo trạng thái.

### 7.3. Full Form Page

Dùng cho form dài/rủi ro cao:

- PO nhiều dòng.
- Phiếu nhận hàng.
- QC inspection.
- Subcontract production order.
- Material handover.
- Shipment manifest.
- Shift closing.

---

## 8. Component architecture

### 8.1. Component layers

```text
Base UI Component
  ↓
ERP Shared Component
  ↓
Module Feature Component
  ↓
Page Container
```

### 8.2. Base UI Component

Lấy từ Ant Design hoặc component nền.

Ví dụ:

- Button
- Input
- Select
- DatePicker
- Table
- Modal
- Drawer
- Upload
- Tooltip

Không sửa trực tiếp component gốc nếu không cần. Customize qua theme/token/wrapper.

### 8.3. ERP Shared Component

Component dùng chung có nghĩa nghiệp vụ ERP.

Ví dụ:

```text
StatusChip
MoneyText
QuantityText
BatchBadge
ExpiryWarningBadge
PermissionGate
ApprovalPanel
AuditTimeline
AttachmentPanel
EntityDrawer
BarcodeScanInput
ConfirmActionModal
ReasonRequiredModal
StockAvailabilityCard
```

### 8.4. Module Feature Component

Component riêng theo module.

Ví dụ module `shipping`:

```text
ManifestTable
CarrierSelector
HandoverScanPanel
MissingOrderResolver
HandoverSummary
```

Ví dụ module `inventory`:

```text
StockLedgerTable
BatchTraceTree
CycleCountSheet
ShiftClosingBoard
StockMovementReasonForm
```

Ví dụ module `production`:

```text
SubcontractOrderForm
MaterialHandoverLines
SampleApprovalPanel
FactoryIssueReportForm
```

---

## 9. API client architecture

### 9.1. API contract

Backend Go phải xuất OpenAPI spec.

Frontend không tự viết type bằng tay cho DTO quan trọng.

```text
OpenAPI spec
  → generated types/client
  → module api wrapper
  → TanStack Query hooks
  → components/pages
```

### 9.2. API response envelope

Frontend kỳ vọng backend trả format thống nhất:

```json
{
  "success": true,
  "data": {},
  "meta": {}
}
```

Lỗi:

```json
{
  "success": false,
  "code": "INSUFFICIENT_STOCK",
  "message": "Tồn khả dụng không đủ",
  "details": {
    "sku": "SERUM-VITC-30ML",
    "requested_qty": 10,
    "available_qty": 6
  }
}
```

### 9.3. Query hook pattern

Mỗi module có hooks riêng:

```text
usePurchaseOrders(params)
usePurchaseOrderDetail(id)
useCreatePurchaseOrder()
useSubmitPurchaseOrder()
useApprovePurchaseOrder()
```

Không gọi API trực tiếp trong component sâu.

Component gọi hook. Hook gọi module API wrapper. API wrapper gọi generated client.

### 9.4. Query key naming

```ts
['purchase-orders', filters]
['purchase-order', id]
['inventory-stock-ledger', filters]
['shipping-manifest', id]
['qc-inspection', id]
```

Quy tắc:

- Query key phải ổn định.
- Filter object phải normalize trước khi đưa vào key.
- Sau mutation phải invalidate đúng key.
- Không invalidate toàn bộ app nếu không cần.

---

## 10. State management rule

### 10.1. Server state

Dữ liệu từ backend dùng TanStack Query:

- danh sách đơn hàng
- chi tiết PO
- tồn kho
- stock ledger
- batch trace
- QC inspection
- shipment manifest
- return inspection

### 10.2. Client state

Dùng React state / Zustand cho:

- trạng thái sidebar
- filter tạm thời
- scan buffer
- selected rows
- wizard step
- temporary draft UI state

### 10.3. Không dùng global store cho mọi thứ

Không đưa toàn bộ dữ liệu ERP vào Zustand/Redux.

Sai:

```text
Lưu tất cả orders, stock, users, permissions, products vào một global store khổng lồ.
```

Đúng:

```text
Server data dùng query cache.
UI state dùng local state hoặc small store.
```

---

## 11. Auth, RBAC và permission UI

### 11.1. Auth strategy frontend

Khuyến nghị:

- Backend set session/access token bằng HttpOnly secure cookie.
- Frontend không lưu token nhạy cảm trong localStorage.
- `/me` endpoint trả user profile, roles, permissions, warehouse scope, branch scope.

### 11.2. Route guard

Mỗi route phải khai báo permission cần có.

Ví dụ:

```ts
const routePermission = 'inventory.shift_closing.view'
```

Nếu user không có quyền:

- Không hiện menu.
- Không cho vào route.
- Nếu truy cập trực tiếp thì redirect hoặc hiển thị 403.

### 11.3. Component-level permission

Action button phải bọc bằng `PermissionGate`.

Ví dụ:

```tsx
<PermissionGate permission="shipping.handover.confirm">
  <Button>Confirm handover</Button>
</PermissionGate>
```

### 11.4. Field-level permission

Một số field không hiện hoặc read-only theo quyền:

- giá vốn
- giá mua
- công nợ
- margin
- lý do điều chỉnh tồn
- trạng thái QC
- payout/commission

Frontend có thể ẩn/disable, nhưng backend vẫn phải kiểm tra cuối cùng.

---

## 12. Form architecture

### 12.1. Form phải chia 3 lớp

```text
Form UI
Form Schema
Submit Mapping
```

- Form UI: layout và input.
- Form Schema: required, type, format, basic validation.
- Submit Mapping: biến dữ liệu form thành request DTO.

### 12.2. Rule bắt buộc

- Không submit nếu field bắt buộc thiếu.
- Không cho submit line item rỗng.
- Không cho số lượng âm, giá âm.
- Không cho ngày hết hạn nhỏ hơn ngày sản xuất nếu có cả hai.
- Không cho QC pass nếu chưa nhập đủ checklist bắt buộc.
- Không cho handover nếu còn đơn chưa scan đủ.
- Không cho shift closing nếu còn discrepancy chưa có reason.

### 12.3. Form lớn phải có section

Ví dụ `Subcontract Production Order`:

```text
1. Factory / Supplier
2. Product / SKU
3. Quantity / Specification
4. Material & Packaging Handover
5. Sample Approval
6. Timeline
7. Deposit / Payment Terms
8. Attachments
9. Approval
```

### 12.4. Autosave

Phase 1 chỉ nên autosave cho draft form dài nếu cần.

Không autosave các hành động nhạy cảm:

- submit duyệt
- QC pass/fail
- stock adjustment
- shipment handover
- final receiving

---

## 13. Table architecture

### 13.1. Table chuẩn ERP

Mỗi table cần có:

- column config rõ
- sorting
- filtering
- pagination server-side
- row action
- bulk action nếu được phép
- sticky action column nếu bảng dài
- status chip
- empty/loading/error state

### 13.2. Server-side pagination

Danh sách lớn bắt buộc dùng pagination từ backend:

- orders
- stock ledger
- audit log
- batch trace
- shipping manifest
- return records

Không load toàn bộ dữ liệu về frontend rồi filter.

### 13.3. Column visibility

Một số cột phải phụ thuộc role:

- cost
- margin
- supplier price
- internal note
- audit-sensitive field

### 13.4. Export

Export không nên lấy dữ liệu hiện có trong table nếu báo cáo lớn.  
Nên gọi backend export job.

---

## 14. Status, state machine và action UX

### 14.1. Status chip

Status phải có màu semantic nhất quán.

Ví dụ:

```text
Draft: neutral
Submitted: info
Approved: success/info
Processing: warning/info
Hold: warning
Pass: success
Fail: danger
Cancelled: muted/danger
Closed: neutral/success
```

### 14.2. Action hợp lệ theo trạng thái

Frontend không hiển thị action không hợp lệ.

Ví dụ Sales Order:

```text
Draft → Submit
Confirmed → Reserve Stock
Reserved → Pick
Picked → Pack
Packed → Add to Manifest
Manifested → Handover
Delivered → Close
```

Nếu đơn đã `HandedOver`, không cho edit line item.

### 14.3. Reason-required action

Các hành động sau phải bắt nhập lý do:

- hủy chứng từ
- reject approval
- fail QC
- stock adjustment
- reopen shift
- mark missing order
- override expiry warning
- split/merge shipment nếu có

---

## 15. Warehouse UX chuẩn

Kho là vùng thao tác tốc độ cao. UI phải ít chữ, rõ trạng thái, hỗ trợ scan và tránh click thừa.

### 15.1. Warehouse Daily Board

Màn hình chính của kho nên có:

```text
- Đơn cần soạn hôm nay
- Đơn đang soạn
- Đơn đã đóng gói
- Đơn chờ bàn giao
- Đơn thiếu hàng
- Đơn lỗi scan
- Hàng hoàn cần xử lý
- Kiểm kê cuối ngày
- Cảnh báo batch/hạn dùng/QC hold
```

### 15.2. Scan UX

`BarcodeScanInput` phải hỗ trợ:

- focus lock
- enter-submit
- beep hoặc visual feedback khi scan thành công/lỗi
- chống duplicate scan
- hiển thị đơn vừa scan
- danh sách scan gần nhất
- cảnh báo nếu scan sai carrier/manifest
- cảnh báo nếu đơn không thuộc khu vực hiện tại

### 15.3. Pick/Pack UX

Màn hình pick/pack cần:

- đơn hàng
- SKU
- batch đề xuất theo FEFO
- vị trí kho
- số lượng cần lấy
- số lượng đã scan
- trạng thái đủ/chưa đủ
- lỗi thiếu hàng
- nút báo issue

### 15.4. Shift closing UX

Cuối ca phải có màn hình riêng:

```text
1. Tổng đơn nhận trong ngày
2. Tổng đơn soạn
3. Tổng đơn đóng gói
4. Tổng đơn bàn giao
5. Đơn còn treo
6. Hàng hoàn đã xử lý
7. Stock discrepancy
8. Lý do lệch
9. Người xác nhận
10. Trưởng kho duyệt/ký
```

Không nên để đối soát cuối ngày nằm trong Excel ngoài hệ thống.

---

## 16. Shipping handover UX

### 16.1. Manifest screen

Manifest là danh sách bàn giao theo ĐVVC/chuyến.

Màn hình cần:

- carrier
- manifest code
- warehouse/area
- số đơn dự kiến
- số đơn đã scan
- số đơn thiếu
- số đơn sai carrier
- trạng thái manifest
- người bàn giao
- thời gian bàn giao
- file biên bản nếu có

### 16.2. Handover scan flow

```text
Chọn carrier/manifest
→ scan từng đơn/thùng
→ hệ thống check order thuộc manifest không
→ đủ đơn thì cho confirm
→ thiếu đơn thì mở Missing Order Resolver
→ ký xác nhận/bàn giao
→ phát event shipment handed over
```

### 16.3. Missing Order Resolver

Khi chưa đủ đơn:

- hiện danh sách đơn thiếu
- hiển thị trạng thái gần nhất: packed / packing / missing / wrong area
- cho nhập lý do
- cho đánh dấu tìm lại trong khu vực đóng hàng
- không cho confirm đủ nếu scan chưa đủ, trừ khi có quyền override

---

## 17. Returns / hàng hoàn UX

### 17.1. Return intake screen

Màn hình nhận hàng hoàn từ shipper:

- scan mã đơn/mã vận đơn
- hiển thị thông tin đơn
- lý do hoàn nếu có
- ảnh tình trạng hàng
- tình trạng niêm phong
- tình trạng móp/hỏng
- batch/expiry nếu cần

### 17.2. Return inspection decision

Quyết định hàng hoàn:

```text
Reusable → chuyển về kho khả dụng hoặc kho chờ QC
Not reusable → chuyển hàng lỗi/lab/hủy theo rule
Need QA review → quarantine
```

### 17.3. UX bắt buộc

- Không cho hàng hoàn quay lại available stock nếu chưa qua bước kiểm.
- Nếu sản phẩm có dấu hiệu hỏng/mở seal, phải yêu cầu ảnh/lý do.
- Nếu hàng hoàn thuộc batch đang bị complaint nhiều, phải cảnh báo QA.

---

## 18. QC UX

### 18.1. Inbound QC

Khi nhập nguyên liệu/bao bì/thành phẩm từ nhà máy:

- hiển thị PO/subcontract order liên quan
- mã NCC/nhà máy
- batch/lô
- hạn dùng
- số lượng nhận
- checklist QC
- ảnh/file COA/MSDS nếu có
- quyết định hold/pass/fail

### 18.2. Batch release UX

Trạng thái phải rõ:

```text
HOLD → PASS
HOLD → FAIL
FAIL → Re-evaluate nếu có quyền
```

Không cho xuất bán batch HOLD/FAIL.

### 18.3. QC fail modal

Nếu fail, bắt buộc:

- lý do fail
- mức độ nghiêm trọng
- ảnh/file minh chứng
- đề xuất xử lý
- thông báo cho kho/sản xuất/purchasing nếu liên quan

---

## 19. Subcontract manufacturing UX

Vì quy trình sản xuất thực tế có nhánh gia công ngoài, frontend cần module riêng cho `Subcontract Production`.

### 19.1. Subcontract order screen

Thông tin cần có:

- nhà máy
- sản phẩm/SKU
- quy cách/mẫu mã
- số lượng đặt
- ngày dự kiến sản xuất
- ngày dự kiến nhận hàng
- điều khoản cọc/thanh toán
- nguyên vật liệu/bao bì cần bàn giao
- checklist chứng từ: COA, MSDS, tem phụ, hóa đơn VAT nếu cần

### 19.2. Material handover screen

Màn hình chuyển NVL/bao bì sang nhà máy:

- danh sách vật tư
- batch/lô vật tư
- số lượng bàn giao
- người bàn giao
- người nhận
- biên bản/file đính kèm
- trạng thái: draft/submitted/approved/handed over

### 19.3. Sample approval screen

Luồng:

```text
Factory makes sample
→ upload sample info/photo
→ R&D/QA/Brand review
→ approve/reject
→ nếu approve thì mở sản xuất hàng loạt
```

Không cho chuyển sang mass production nếu sample chưa approved.

### 19.4. Factory issue report

Nếu hàng giao về không đạt:

- tạo issue report
- ghi số lượng lỗi
- loại lỗi
- ảnh/video
- liên kết batch/subcontract order
- deadline phản hồi nhà máy 3–7 ngày

---

## 20. Dashboard UX

### 20.1. CEO/COO dashboard Phase 1

Nên có các card:

```text
- Doanh thu hôm nay
- Đơn chờ xử lý
- Đơn chờ bàn giao
- Đơn giao thất bại/hoàn
- Tồn khả dụng
- Batch HOLD/FAIL
- Hàng cận date
- PO trễ
- Subcontract order đang chờ mẫu
- Subcontract order trễ nhận hàng
- Tồn lệch cuối ca
```

### 20.2. Warehouse dashboard

```text
- Đơn cần pick
- Đơn đang pack
- Đơn chờ handover
- Manifest chưa đủ scan
- Hàng hoàn chưa kiểm
- Kiểm kê cuối ca chưa hoàn thành
```

### 20.3. Drill-down rule

Mọi số trên dashboard phải drill-down về danh sách chứng từ gốc.

Không có số “trang trí”.

---

## 21. Approval UI

### 21.1. Approval panel

Mọi chứng từ cần duyệt phải có block:

```text
Current status
Current approver
Approval history
Submitted by
Submitted at
Comments
Approve / Reject / Request changes
```

### 21.2. Inbox phê duyệt

Người quản lý cần màn hình gom toàn bộ task chờ duyệt:

```text
/pending-approvals
```

Filter theo:

- module
- loại chứng từ
- người submit
- ngày submit
- mức độ ưu tiên

### 21.3. Reject phải có lý do

Reject không được để trống lý do.

---

## 22. Audit và attachment UX

### 22.1. Audit Timeline

Các màn hình chi tiết phải có audit timeline:

- ai tạo
- ai sửa
- sửa trường nào
- trước/sau nếu nhạy cảm
- ai duyệt
- ai hủy
- lý do

### 22.2. Attachment Panel

Dùng cho:

- COA/MSDS
- phiếu giao hàng
- biên bản bàn giao
- ảnh hàng hoàn
- ảnh hàng lỗi
- chứng từ nhà máy
- hóa đơn/chứng từ thanh toán nếu scope có

### 22.3. Upload rule

- Hiện file type được phép.
- Hiện max size.
- Cho preview ảnh/PDF.
- Không cho xóa file quan trọng nếu chứng từ đã approved, trừ role có quyền và có audit.

---

## 23. Error handling UX

### 23.1. Lỗi nghiệp vụ

Hiển thị message rõ, không hiển thị lỗi kỹ thuật thô.

Ví dụ:

```text
Tồn khả dụng không đủ. SKU SERUM-VITC-30ML còn 6, yêu cầu 10.
```

Không hiển thị:

```text
500 Internal Server Error
```

### 23.2. Lỗi validation

- Field lỗi hiển thị ngay tại field.
- Form lỗi tổng hợp ở đầu form nếu nhiều lỗi.
- Line item lỗi phải scroll/highlight tới dòng lỗi.

### 23.3. Lỗi mạng

- Hiện retry.
- Không tự submit lại hành động nhạy cảm nếu không có idempotency key.
- Với scan, phải hiển thị scan chưa đồng bộ nếu có cơ chế queue.

### 23.4. Idempotency UX

Các action nhạy cảm cần chống bấm 2 lần:

- disable button sau submit
- loading state rõ
- backend dùng idempotency key nếu cần

---

## 24. Loading, empty, disabled states

### 24.1. Loading

- Table dùng skeleton/table loading.
- Form submit dùng loading button.
- Dashboard card dùng skeleton.

### 24.2. Empty state

Empty state phải hướng dẫn hành động tiếp theo.

Ví dụ:

```text
Chưa có manifest nào hôm nay. Tạo manifest mới để bắt đầu bàn giao ĐVVC.
```

### 24.3. Disabled state

Nếu action disabled, phải có tooltip/lý do.

Ví dụ:

```text
Không thể bàn giao vì còn 3 đơn chưa scan đủ.
```

---

## 25. Performance standard

### 25.1. Bắt buộc

- Server-side pagination cho table lớn.
- Debounce search input.
- Không fetch detail hàng loạt nếu không cần.
- Không render table quá lớn không virtualize.
- Không reload cả page sau mỗi mutation.
- Cache hợp lý bằng TanStack Query.

### 25.2. UX mục tiêu

```text
Mở list thông thường: < 2 giây trong mạng nội bộ tốt
Search/filter: phản hồi rõ trong < 1 giây sau debounce
Scan đơn: feedback gần như tức thì
Submit action: có loading/confirmation ngay lập tức
```

### 25.3. Các màn hình cần đặc biệt tối ưu

- stock ledger
- order list
- pick/pack board
- manifest handover
- audit log
- batch trace

---

## 26. Accessibility và usability

### 26.1. Keyboard-first

Kho và backoffice cần thao tác bàn phím tốt.

- Tab order hợp lý.
- Enter để submit scan.
- Esc đóng modal/drawer nếu an toàn.
- Shortcut cho thao tác lặp lại nếu được duyệt.

### 26.2. Contrast và status

Không chỉ dùng màu để thể hiện trạng thái.  
Status chip phải có text.

Ví dụ:

```text
PASS
HOLD
FAIL
NEAR EXPIRY
MISSING
```

### 26.3. Ngôn ngữ UI

UI tiếng Việt là mặc định.

Text phải dễ hiểu cho nhân viên vận hành, không dùng thuật ngữ kỹ thuật thô.

Ví dụ:

Dùng:

```text
Tồn khả dụng không đủ
```

Không dùng:

```text
Stock reservation failed
```

---

## 27. Internationalization / localization

Phase 1 dùng tiếng Việt.  
Nhưng nên chuẩn bị cấu trúc i18n nếu sau này có đa ngôn ngữ.

### 27.1. Format chuẩn

- tiền: VND, có phân tách nghìn
- ngày: `dd/MM/yyyy`
- giờ: `HH:mm`
- timezone: theo cấu hình công ty
- số lượng: format theo unit

### 27.2. Không hardcode text quan trọng trong nhiều nơi

Status label nên map từ constant.

```ts
const QC_STATUS_LABEL = {
  HOLD: 'Đang giữ',
  PASS: 'Đạt',
  FAIL: 'Không đạt',
}
```

---

## 28. Security frontend baseline

### 28.1. Không lưu token nhạy cảm trong localStorage

Dùng HttpOnly cookie nếu backend hỗ trợ.

### 28.2. Không tin frontend permission

Frontend chỉ là lớp ẩn/hiện UX.  
Backend là lớp kiểm soát thật.

### 28.3. Chống lộ dữ liệu nhạy cảm

Không render dữ liệu mà role không được xem.

Ví dụ:

- cost
- margin
- supplier price
- payroll
- payout

### 28.4. File upload

- validate file type ở frontend
- backend validate lại
- không cho upload file executable
- preview an toàn

---

## 29. Frontend coding standards rút gọn

### 29.1. Naming

```text
Component: PascalCase
Hook: useSomething
File component: PascalCase.tsx
Utility: camelCase.ts
Type/interface: PascalCase
Constant: UPPER_SNAKE_CASE
Route: kebab-case
```

### 29.2. Component rule

- Component không quá lớn.
- Page không chứa business logic dài.
- Hook xử lý data fetching/mutation.
- Schema validation đặt trong `schemas`.
- Mapping DTO đặt trong `api` hoặc `mappers`.

### 29.3. Import rule

- Không import chéo module sâu.
- Shared component import từ `shared/components`.
- Module public API import qua index nếu cần.

### 29.4. Business rule rule

Frontend có thể validate để UX tốt, nhưng rule nghiệp vụ cuối cùng nằm ở backend.

Ví dụ frontend kiểm tra thiếu scan, nhưng backend vẫn phải reject handover nếu chưa đủ.

---

## 30. Testing strategy

### 30.1. Unit test

Test:

- formatters
- validators
- permission helpers
- status mappers
- DTO mappers

### 30.2. Component test

Test:

- DataTable render
- StatusChip mapping
- PermissionGate
- BarcodeScanInput
- ApprovalPanel
- ReturnInspectionForm

### 30.3. E2E test bằng Playwright

Critical flows Phase 1:

```text
1. Login → mở dashboard
2. Tạo PO → submit → approve
3. Nhận hàng → QC hold/pass → nhập kho khả dụng
4. Tạo sales order → reserve stock → pick/pack
5. Tạo manifest → scan đủ đơn → confirm handover
6. Scan thiếu đơn → Missing Order Resolver
7. Nhận hàng hoàn → inspection → reusable/not reusable
8. Tạo subcontract order → material handover → sample approval → receive goods
9. Shift closing cuối ngày → ghi discrepancy reason → submit
10. User không có quyền truy cập route/action nhạy cảm
```

### 30.4. Visual regression

Có thể bổ sung sau Phase 1 nếu UI ổn định.

---

## 31. Frontend deployment

### 31.1. Environment

```text
local
staging
production
```

### 31.2. Env variables

```text
NEXT_PUBLIC_API_BASE_URL
NEXT_PUBLIC_APP_ENV
NEXT_PUBLIC_SENTRY_DSN optional
```

Không đưa secret server vào frontend env public.

### 31.3. Build pipeline

CI cần chạy:

```text
pnpm install
pnpm lint
pnpm typecheck
pnpm test
pnpm build
```

E2E chạy ở staging hoặc môi trường test riêng.

---

## 32. Observability frontend

### 32.1. Cần log gì

Frontend không log dữ liệu nhạy cảm, nhưng cần tracking:

- page load error
- API error code
- failed scan event
- UI crash
- long task/performance issue
- failed file upload

### 32.2. Error boundary

Cần có:

- global error boundary
- route-level error fallback
- module-level error fallback cho màn hình quan trọng

### 32.3. User support trace

Nên hiện `request_id` hoặc `trace_id` từ backend error để support tra log.

Ví dụ:

```text
Mã lỗi: REQ-20260424-ABC123
```

---

## 33. Module frontend design Phase 1

## 33.1. Master Data

Màn hình:

```text
Items
SKUs
Suppliers
Customers
Warehouses
Bins/Locations
Batch templates
Units of measure
```

Chuẩn UX:

- tránh tạo mã trùng
- cảnh báo khi item đang được dùng
- không cho xóa master data đã phát sinh giao dịch
- hỗ trợ active/inactive

## 33.2. Purchase

Màn hình:

```text
Purchase Requests
Purchase Orders
Inbound Schedule
Supplier Detail
```

Chuẩn UX:

- line item rõ SKU/UOM/qty/price
- trạng thái duyệt rõ
- PO có attachment báo giá/chứng từ
- highlight PO trễ

## 33.3. Inventory

Màn hình:

```text
Goods Receipt
Stock Issue
Stock Transfer
Stock Ledger
Batch Trace
Cycle Count
Shift Closing
```

Chuẩn UX:

- batch và expiry phải nổi bật
- stock ledger chỉ xem, không sửa
- adjustment phải có reason
- shift closing có discrepancy summary

## 33.4. QC

Màn hình:

```text
Inbound Inspection
Batch Release
Nonconformance
```

Chuẩn UX:

- hold/pass/fail rõ
- fail bắt buộc reason/attachment
- pass mới mở available stock nếu rule yêu cầu

## 33.5. Production / Subcontract Manufacturing

Màn hình:

```text
Subcontract Orders
Material Handover
Sample Approval
Production Receipt
Factory Issue Report
```

Chuẩn UX:

- sample chưa approve thì không cho mass production
- material handover phải có batch/qty/signature/attachment
- hàng lỗi phải report trong SLA 3–7 ngày

## 33.6. Sales

Màn hình:

```text
Sales Orders
Order Reservation
Returns
Customer Detail
```

Chuẩn UX:

- không cho reserve vượt tồn khả dụng
- discount vượt quyền phải xin duyệt
- order status rõ
- returns nối về order gốc

## 33.7. Shipping

Màn hình:

```text
Pick/Pack Board
Manifest List
Handover Scan
Delivery Status
```

Chuẩn UX:

- scan nhanh
- check carrier/manifest
- chưa đủ đơn không cho confirm
- missing order resolver rõ

---

## 34. Frontend/backend contract rule

### 34.1. Không tự suy luận nghiệp vụ phức tạp ở frontend

Ví dụ frontend không tự tính available stock cuối cùng từ nhiều bảng.  
Backend trả `available_stock` đã chuẩn.

Frontend có thể hiển thị:

```text
physical_stock
reserved_stock
qc_hold_stock
available_stock
```

Nhưng backend là nơi tính đúng.

### 34.2. Action endpoint rõ

Các hành động nghiệp vụ dùng endpoint action:

```text
POST /sales-orders/{id}/submit
POST /sales-orders/{id}/reserve-stock
POST /shipping-manifests/{id}/confirm-handover
POST /qc-inspections/{id}/pass
POST /qc-inspections/{id}/fail
POST /inventory-shifts/{id}/close
```

Frontend không tự PATCH status tùy tiện.

---

## 35. Definition of Done cho frontend task

Một task frontend chỉ được coi là xong khi:

- Đúng route/module.
- Đúng permission display.
- Đúng API contract.
- Có loading state.
- Có empty state nếu là list.
- Có error state.
- Form có validation cơ bản.
- Mutation có confirmation nếu rủi ro.
- Action disabled có lý do.
- Status chip đúng.
- Audit/attachment/approval block nếu màn hình yêu cầu.
- Responsive tối thiểu theo chuẩn ERP.
- Lint/typecheck pass.
- Test cần thiết pass.
- Không hardcode dữ liệu giả.
- Không expose field nhạy cảm sai role.

---

## 36. Checklist handoff UI/UX → Frontend

UI/UX phải bàn giao cho frontend:

```text
1. Screen name
2. Route
3. User role/permission
4. Data source/API dự kiến
5. Status list
6. Main actions
7. Disabled state rule
8. Empty state
9. Error state
10. Loading state
11. Form validation
12. Table columns
13. Filter fields
14. Mobile/tablet behavior nếu có
15. Audit/attachment/approval block
```

Không nhận màn hình chỉ có hình đẹp mà thiếu rule nghiệp vụ.

---

## 37. Rủi ro frontend cần né

### 37.1. Page nào cũng tự gọi API một kiểu

Kết quả: khó maintain, lỗi cache, lỗi permission.

### 37.2. Table không server-side

Kết quả: chậm, treo browser khi dữ liệu lớn.

### 37.3. Scan UX không tối ưu

Kết quả: kho thao tác chậm, scan nhầm, bàn giao sai.

### 37.4. Permission chỉ ẩn menu

Kết quả: user vẫn gọi route/action trực tiếp được nếu backend không chặn.

### 37.5. Không có error code mapping

Kết quả: user không hiểu lỗi, support khó xử lý.

### 37.6. Quá nhiều custom component không cần thiết

Kết quả: chậm build, khó maintain.

Dùng Ant Design làm nền, chỉ custom nơi tạo khác biệt nghiệp vụ.

---

## 38. Kết luận kiến trúc frontend

Frontend Phase 1 nên chốt:

```text
Framework: Next.js + React
Language: TypeScript
UI Base: Ant Design + custom ERP design tokens
API: REST + OpenAPI generated client
Server state: TanStack Query
Form: React Hook Form + Zod
State: Local state/Zustand phạm vi nhỏ
Testing: Vitest + RTL + Playwright
Architecture: module-based frontend, bám Go modular backend
```

Triết lý chính:

```text
Frontend ERP không chỉ hiển thị dữ liệu.
Frontend ERP phải dẫn người dùng làm đúng quy trình, đúng quyền, đúng trạng thái, đúng dữ liệu.
```

Với công ty mỹ phẩm, frontend phải đặc biệt bảo vệ các điểm:

```text
batch
hạn dùng
QC hold/pass/fail
tồn khả dụng
soạn/đóng/bàn giao ĐVVC
hàng hoàn
gia công ngoài
kiểm kê/đối soát cuối ngày
```

Nếu làm đúng, UI không chỉ đẹp mà còn giảm sai sót vận hành, giảm thất thoát kho, giảm nhầm hàng, và giúp ERP sống thật trong công ty.

---

## 39. Next document đề xuất

Sau tài liệu này, tài liệu kỹ thuật tiếp theo nên là:

```text
16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md
```

Mục tiêu: chốt chuẩn REST API, request/response, error code, pagination, filtering, action endpoint, idempotency, và OpenAPI codegen giữa frontend Next.js và backend Go.
