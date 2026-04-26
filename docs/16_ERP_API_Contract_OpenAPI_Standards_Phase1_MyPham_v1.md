# 16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1

**Dự án:** Web ERP công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** API Contract + OpenAPI Standards  
**Scope:** Phase 1  
**Version:** v1.0  
**Date:** 2026-04-24  
**Language:** Vietnamese  
**Backend:** Go Modular Monolith  
**Frontend:** React / Next.js + TypeScript  
**API Style:** REST + OpenAPI 3.1  
**Owner:** ERP Solution Architect / Technical Lead / Backend Lead / Frontend Lead  

**Related Documents:**

- `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
- `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`
- `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`
- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`
- `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`
- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt **chuẩn API contract** cho ERP Phase 1.

Mục tiêu là làm cho backend Go, frontend React/Next.js, QA, DevOps và các bên tích hợp ngoài cùng nói một ngôn ngữ khi làm API.

Tài liệu này trả lời:

1. API đặt tên thế nào.
2. Versioning thế nào.
3. Request/response chuẩn ra sao.
4. Error format chuẩn ra sao.
5. Pagination/filter/sort/search dùng chung thế nào.
6. Auth/RBAC/permission thể hiện trong API thế nào.
7. Các action nghiệp vụ như submit, approve, reserve, pick, pack, handover, return, QC pass/fail viết API ra sao.
8. OpenAPI được tổ chức, generate client và test contract như thế nào.
9. API cho scan kho, bàn giao ĐVVC, hàng hoàn và gia công ngoài cần chuẩn gì.

Tóm gọn:

> PRD nói hệ thống làm gì.  
> Technical Architecture nói xây hệ thống bằng gì.  
> API Contract nói các module nói chuyện với nhau và với frontend như thế nào.

---

## 2. Kết luận API đã chốt

| Hạng mục | Quyết định |
|---|---|
| API Style | REST-first |
| Contract | OpenAPI 3.1 |
| API Version | `/api/v1` |
| Data Format | JSON |
| Date/Time | ISO 8601 UTC ở API; frontend hiển thị theo timezone cấu hình |
| Auth | Bearer token / session hybrid, tùy triển khai bảo mật cuối |
| Permission | RBAC + action permission + record-level guard nếu cần |
| Pagination | Cursor hoặc page-based; Phase 1 ưu tiên page-based cho list quản trị, cursor cho scan/log nếu cần |
| Error Format | Chuẩn hóa theo `code`, `message`, `details`, `trace_id` |
| API Documentation | OpenAPI YAML/JSON sinh tự động + review thủ công |
| Client Generation | Frontend dùng OpenAPI generated types/client |
| Idempotency | Bắt buộc cho các mutation dễ double-submit |
| Audit | Mọi mutation quan trọng phải tạo audit log |
| State Change | Không update status tùy tiện; dùng action endpoint |

---

## 3. Business reality anchor từ workflow thực tế

API Phase 1 không được thiết kế kiểu CRUD chung chung. Nó phải bám 4 thực tế vận hành đã thu từ tài liệu nội bộ.

### 3.1. Kho có nhịp xử lý trong ngày và đóng ca

Kho có chuỗi việc: nhận đơn trong ngày, xuất/nhập theo nội quy, soạn - đóng gói, tối ưu vị trí kho, kiểm kê tồn cuối ngày, đối soát số liệu, báo cáo quản lý và kết thúc ca.

API cần có các nhóm:

- Warehouse daily board.
- Cycle count / stock count.
- Shift closing.
- End-of-day reconciliation.
- Stock discrepancy.
- Adjustment request.

### 3.2. Nội quy kho có 4 nhánh nghiệp vụ

Nội quy kho tách rõ:

1. Nhập kho.
2. Xuất kho.
3. Đóng hàng.
4. Xử lý hàng hoàn.

API cần có các nhóm endpoint riêng cho:

- inbound receiving,
- outbound issue/pick/pack,
- packing verification,
- return receiving/inspection/disposition.

### 3.3. Bàn giao ĐVVC cần scan và manifest

Bàn giao đơn cho đơn vị vận chuyển có các bước: phân chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, lấy hàng và quét mã trực tiếp tại hầm/khu bàn giao, nếu đủ thì ký xác nhận; nếu chưa đủ thì kiểm tra mã hoặc tìm lại trong khu đóng hàng.

API cần có:

- shipment manifest,
- scan-to-verify,
- handover confirmation,
- missing package handling,
- carrier handover record,
- scan event log.

### 3.4. Sản xuất có nhánh gia công ngoài

Luồng sản xuất hiện tại có nhánh gia công/đặt nhà máy: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, nhận hàng về kho, kiểm tra số lượng/chất lượng, nhận hàng hoặc báo lỗi nhà máy trong 3–7 ngày.

API cần có:

- subcontract manufacturing order,
- deposit/final payment milestone,
- material/packaging transfer to factory,
- sample approval,
- finished goods receiving,
- factory defect report,
- acceptance/rejection window.

---

## 4. Nguyên tắc thiết kế API

### 4.1. API phải phản ánh nghiệp vụ, không chỉ phản ánh database

Không tạo API kiểu:

```text
PUT /api/v1/sales-orders/{id}
body: { status: "SHIPPED" }
```

Vì kiểu đó cho phép nhảy trạng thái bừa.

Phải dùng action endpoint:

```text
POST /api/v1/sales-orders/{id}/reserve-stock
POST /api/v1/sales-orders/{id}/confirm
POST /api/v1/shipments/{id}/handover
```

Ý nghĩa:

- Backend kiểm tra trạng thái hiện tại.
- Backend kiểm tra quyền.
- Backend kiểm tra điều kiện nghiệp vụ.
- Backend ghi audit log.
- Backend phát event nếu cần.

### 4.2. List/detail/create/update/action phải tách rõ

Chuẩn endpoint:

```text
GET    /api/v1/resources
POST   /api/v1/resources
GET    /api/v1/resources/{id}
PATCH  /api/v1/resources/{id}
POST   /api/v1/resources/{id}/actions/{action-name}
```

Ví dụ:

```text
GET    /api/v1/purchase-orders
POST   /api/v1/purchase-orders
GET    /api/v1/purchase-orders/{po_id}
PATCH  /api/v1/purchase-orders/{po_id}
POST   /api/v1/purchase-orders/{po_id}/submit
POST   /api/v1/purchase-orders/{po_id}/approve
POST   /api/v1/purchase-orders/{po_id}/cancel
```

### 4.3. Không để frontend tự tính nghiệp vụ quan trọng

Frontend có thể hiển thị tính toán gợi ý, nhưng quyết định cuối phải ở backend.

Backend phải tính/kiểm tra:

- tồn khả dụng,
- reserve stock,
- batch QC status,
- hạn dùng,
- quyền giảm giá,
- trạng thái chứng từ,
- điều kiện handover,
- điều kiện return,
- điều kiện nhận hàng gia công,
- điều kiện final payment.

### 4.4. API phải dễ generate frontend client

Mọi response schema phải rõ trong OpenAPI.

Không dùng response mơ hồ kiểu:

```json
{
  "data": {}
}
```

Nếu endpoint trả sales order thì schema phải là:

```text
SalesOrderDetailResponse
```

Nếu endpoint trả list thì schema phải là:

```text
PagedResponse<SalesOrderListItem>
```

### 4.5. Mutation quan trọng phải idempotent

Các API dễ bị bấm 2 lần hoặc mạng lag phải hỗ trợ idempotency.

Áp dụng cho:

- tạo đơn hàng,
- reserve stock,
- confirm shipment handover,
- receive inbound,
- approve QC,
- create return receipt,
- post stock adjustment,
- create payment request,
- receive subcontract goods.

Header chuẩn:

```text
Idempotency-Key: <uuid>
```

Backend lưu key theo user + endpoint + payload hash trong một khoảng thời gian cấu hình.

### 4.6. Trạng thái phải đi qua state machine

Không cho API patch status tự do.

Ví dụ batch QC:

```text
HOLD -> PASS
HOLD -> FAIL
FAIL -> REWORK_REVIEW nếu Phase sau cần
PASS không quay lại HOLD nếu không có reversal/quality incident riêng
```

Sales order:

```text
DRAFT -> CONFIRMED -> RESERVED -> PICKED -> PACKED -> HANDED_OVER -> DELIVERED -> CLOSED
```

Nếu cần hủy hoặc rollback, dùng action:

```text
POST /api/v1/sales-orders/{id}/cancel
POST /api/v1/shipments/{id}/mark-delivery-failed
POST /api/v1/qc-inspections/{id}/void-release
```

---

## 5. Base URL, versioning và environment

### 5.1. Base URL

```text
Production: https://erp.company.com/api/v1
Staging:    https://staging-erp.company.com/api/v1
Dev:        https://dev-erp.company.com/api/v1
Local:      http://localhost:8080/api/v1
```

Tên domain thực tế sẽ chốt khi triển khai hạ tầng.

### 5.2. API versioning

Phase 1 dùng:

```text
/api/v1
```

Không đưa version vào query string.

Không dùng:

```text
/api/resources?version=1
```

Nếu sau này có breaking change lớn:

```text
/api/v2
```

### 5.3. Breaking change là gì?

Các thay đổi sau là breaking change:

- xóa field đang dùng,
- đổi type field,
- đổi meaning field,
- đổi enum value,
- đổi error code,
- đổi rule bắt buộc request,
- đổi pagination response,
- đổi status workflow.

Các thay đổi không breaking nếu:

- thêm optional field,
- thêm endpoint mới,
- thêm enum value nếu frontend đã xử lý unknown value,
- thêm filter mới,
- thêm metadata mới.

---

## 6. HTTP method convention

| Method | Dùng cho | Ghi chú |
|---|---|---|
| GET | đọc dữ liệu | không thay đổi state |
| POST | tạo mới hoặc action nghiệp vụ | dùng cho submit/approve/reserve/handover |
| PATCH | sửa một phần thông tin khi chưa locked | không dùng để nhảy status |
| PUT | thay thế toàn bộ resource | hạn chế dùng Phase 1 |
| DELETE | xóa mềm resource chưa phát sinh giao dịch | không dùng cho chứng từ quan trọng |

### 6.1. Delete policy

Không xóa cứng:

- phiếu nhập,
- phiếu xuất,
- stock movement,
- QC release,
- sales order đã confirmed,
- shipment,
- return receipt,
- subcontract manufacturing order,
- payment/công nợ.

Chỉ cho:

```text
POST /api/v1/resources/{id}/cancel
POST /api/v1/resources/{id}/void
POST /api/v1/resources/{id}/reverse
```

Tùy nghiệp vụ.

---

## 7. Naming convention

### 7.1. Path naming

Dùng kebab-case, số nhiều cho resource.

Đúng:

```text
/api/v1/purchase-orders
/api/v1/goods-receipts
/api/v1/qc-inspections
/api/v1/stock-movements
/api/v1/sales-orders
/api/v1/shipment-manifests
/api/v1/return-receipts
/api/v1/subcontract-orders
```

Không dùng:

```text
/api/v1/getPurchaseOrders
/api/v1/PurchaseOrder
/api/v1/purchase_order
```

### 7.2. JSON field naming

Dùng `snake_case` cho JSON field.

Ví dụ:

```json
{
  "sales_order_id": "so_01HX...",
  "order_no": "SO-20260424-0001",
  "customer_id": "cus_01HX...",
  "order_date": "2026-04-24",
  "created_at": "2026-04-24T08:30:00Z"
}
```

Lý do:

- nhất quán với nhiều API enterprise,
- dễ đọc trong JSON,
- tách khỏi Go struct naming và frontend variable naming.

Frontend generated client có thể map hoặc dùng trực tiếp tùy codegen.

### 7.3. ID naming

ID field luôn có hậu tố `_id`.

Ví dụ:

```text
sku_id
batch_id
warehouse_id
supplier_id
purchase_order_id
sales_order_id
shipment_id
return_receipt_id
```

Human-readable number dùng hậu tố `_no`.

```text
sku_code
batch_no
purchase_order_no
sales_order_no
shipment_no
return_receipt_no
```

Phân biệt:

- `id`: internal unique ID, không có ý nghĩa nghiệp vụ.
- `*_no`: số chứng từ/mã nghiệp vụ đọc được.
- `*_code`: mã master data do doanh nghiệp quản lý.

---

## 8. Response envelope chuẩn

### 8.1. Single object success response

```json
{
  "success": true,
  "data": {
    "id": "po_01HX...",
    "purchase_order_no": "PO-20260424-0001",
    "status": "APPROVED"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 8.2. List response

```json
{
  "success": true,
  "data": [
    {
      "id": "sku_01HX...",
      "sku_code": "SERUM-VITC-30ML",
      "name": "Serum Vitamin C 30ml",
      "status": "ACTIVE"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 245,
    "total_pages": 13
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 8.3. Action success response

```json
{
  "success": true,
  "data": {
    "id": "shp_01HX...",
    "shipment_no": "SHP-20260424-0012",
    "previous_status": "PACKED",
    "current_status": "HANDED_OVER",
    "handover_at": "2026-04-24T10:15:00Z"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 8.4. Không trả raw array

Không dùng:

```json
[
  { "id": "sku_1" }
]
```

Vì thiếu pagination, trace_id, metadata và khó mở rộng.

---

## 9. Error response chuẩn

### 9.1. Error envelope

```json
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_AVAILABLE_STOCK",
    "message": "Tồn khả dụng không đủ để giữ hàng.",
    "details": {
      "sku_code": "SERUM-VITC-30ML",
      "requested_qty": "10",
      "available_qty": "6"
    }
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 9.2. Error code convention

Dùng UPPER_SNAKE_CASE.

Ví dụ:

```text
VALIDATION_ERROR
UNAUTHORIZED
FORBIDDEN
RESOURCE_NOT_FOUND
RESOURCE_CONFLICT
INVALID_STATE_TRANSITION
INSUFFICIENT_AVAILABLE_STOCK
BATCH_QC_NOT_PASSED
BATCH_EXPIRED
DUPLICATE_DOCUMENT_NO
IDEMPOTENCY_CONFLICT
APPROVAL_REQUIRED
DISCOUNT_LIMIT_EXCEEDED
SHIPMENT_NOT_READY_FOR_HANDOVER
PACKAGE_SCAN_MISMATCH
RETURN_ITEM_NOT_INSPECTED
SUBCONTRACT_SAMPLE_NOT_APPROVED
```

### 9.3. HTTP status mapping

| HTTP Status | Dùng cho | Ví dụ |
|---|---|---|
| 200 | đọc/action thành công | approve, handover |
| 201 | tạo mới thành công | create PO |
| 202 | nhận request async | export report |
| 400 | request sai | invalid payload |
| 401 | chưa đăng nhập | missing token |
| 403 | không đủ quyền | không được duyệt PO |
| 404 | không tìm thấy | PO không tồn tại |
| 409 | xung đột state/dữ liệu | batch đã release, stock không đủ |
| 422 | dữ liệu hợp lệ về format nhưng sai nghiệp vụ | batch QC chưa pass |
| 429 | rate limit | scan quá nhanh bất thường/API abuse |
| 500 | lỗi hệ thống | unexpected error |

### 9.4. Validation error format

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Dữ liệu không hợp lệ.",
    "details": {
      "fields": [
        {
          "field": "items[0].qty",
          "code": "REQUIRED",
          "message": "Số lượng là bắt buộc."
        },
        {
          "field": "supplier_id",
          "code": "INVALID_REFERENCE",
          "message": "Nhà cung cấp không tồn tại hoặc đã bị khóa."
        }
      ]
    }
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 9.5. Error message rule

- `code` để frontend/test xử lý logic.
- `message` để người dùng đọc.
- `details` để debug/nghiệp vụ hiểu rõ.
- Không expose SQL, stack trace, internal table name.
- Không trả thông tin nhạy cảm như cost, margin cho user không có quyền.

---

## 10. Pagination, filtering, sorting, search

### 10.1. Page-based pagination chuẩn

Query:

```text
GET /api/v1/sales-orders?page=1&page_size=20
```

Response:

```json
{
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 123,
    "total_pages": 7
  }
}
```

### 10.2. Page size limit

| Context | Default | Max |
|---|---:|---:|
| Admin list | 20 | 100 |
| Report list | 50 | 500 |
| Scan/event log | 50 | 200 |
| Export | async job | theo config |

Không cho frontend gọi `page_size=100000`.

### 10.3. Filtering convention

Dùng query params rõ nghĩa:

```text
GET /api/v1/sales-orders?status=CONFIRMED&channel=B2B&date_from=2026-04-01&date_to=2026-04-24
```

Đối với multi-value:

```text
GET /api/v1/sales-orders?status=CONFIRMED,RESERVED,PACKED
```

Hoặc:

```text
GET /api/v1/sales-orders?status[]=CONFIRMED&status[]=RESERVED
```

Chọn một kiểu trong implementation. Khuyến nghị Phase 1 dùng comma-separated cho đơn giản.

### 10.4. Sorting convention

```text
GET /api/v1/purchase-orders?sort=-created_at
GET /api/v1/stock-batches?sort=expiry_date
GET /api/v1/sales-orders?sort=-order_date,order_no
```

Dấu `-` là descending.

### 10.5. Search convention

```text
GET /api/v1/products?q=serum
GET /api/v1/sales-orders?q=SO-20260424
GET /api/v1/customers?q=0988
```

Search chỉ dùng cho tìm kiếm nhanh. Không thay thế filter nghiệp vụ.

### 10.6. Filter fields whitelist

Backend phải whitelist filter/sort fields. Không nhận query tùy ý map thẳng vào SQL.

Ví dụ module sales order cho phép:

```text
q
status
channel
customer_id
date_from
date_to
warehouse_id
payment_status
shipment_status
created_by
sort
page
page_size
```

---

## 11. Date/time, money, quantity và decimal

### 11.1. Date/time

API trả datetime theo UTC:

```json
{
  "created_at": "2026-04-24T08:30:00Z"
}
```

Date không có giờ dùng format:

```json
{
  "expiry_date": "2028-04-24",
  "manufacturing_date": "2026-04-24"
}
```

Frontend hiển thị theo timezone công ty.

### 11.2. Money

Không dùng float cho tiền.

Khuyến nghị API trả tiền dạng string decimal hoặc integer minor unit. Phase 1 khuyến nghị string decimal dễ đọc:

```json
{
  "currency": "VND",
  "unit_price": "125000",
  "discount_amount": "10000",
  "total_amount": "115000"
}
```

Nếu dùng minor unit:

```json
{
  "currency": "VND",
  "unit_price_minor": 12500000
}
```

Chỉ chọn một chuẩn. Khuyến nghị VND dùng decimal string để tránh nhầm multiplier.

### 11.3. Quantity

Không dùng float cho quantity.

Dùng string decimal:

```json
{
  "qty": "10.5",
  "uom_code": "KG"
}
```

Vì nguyên liệu có thể tính gram/kg/ml/lít, còn thành phẩm tính chai/hộp/thùng.

### 11.4. UOM conversion

API không để frontend tự quy đổi đơn vị quan trọng.

Backend trả rõ:

```json
{
  "qty": "1000",
  "uom_code": "G",
  "base_qty": "1",
  "base_uom_code": "KG"
}
```

---

## 12. Auth, permission và tenant/company context

### 12.1. Auth header

```text
Authorization: Bearer <access_token>
```

Nếu dùng session cookie nội bộ, vẫn phải có CSRF protection. API contract vẫn mô tả bearer token để rõ machine-to-machine và generated client.

### 12.2. User context trong request

Backend lấy user từ token/session, không nhận `user_id` từ request body cho hành động nghiệp vụ.

Không cho:

```json
{
  "approved_by": "user_123"
}
```

Backend tự set:

```text
approved_by = current_user.id
approved_at = now()
```

### 12.3. Permission code convention

Permission dùng dạng:

```text
module.resource.action
```

Ví dụ:

```text
inventory.stock.view
inventory.stock.adjust
inventory.receiving.create
inventory.receiving.approve
qc.inspection.release
sales.order.create
sales.order.confirm
shipping.manifest.handover
returns.receipt.inspect
subcontract.order.approve
```

### 12.4. Forbidden response

```json
{
  "success": false,
  "error": {
    "code": "FORBIDDEN",
    "message": "Bạn không có quyền thực hiện hành động này.",
    "details": {
      "required_permission": "shipping.manifest.handover"
    }
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 12.5. Field-level masking

API phải mask/hide field nhạy cảm nếu user không có quyền.

Ví dụ user kho không được xem cost:

```json
{
  "sku_code": "SERUM-VITC-30ML",
  "name": "Serum Vitamin C 30ml",
  "available_qty": "120"
}
```

Không trả:

```json
{
  "standard_cost": "52000",
  "gross_margin": "0.41"
}
```

Không chỉ hide ở frontend. Backend phải không trả field.

---

## 13. Audit và mutation metadata

### 13.1. Audit fields chuẩn

Các resource quan trọng có:

```json
{
  "created_at": "2026-04-24T08:30:00Z",
  "created_by": {
    "id": "usr_01HX...",
    "display_name": "Nguyen A"
  },
  "updated_at": "2026-04-24T09:00:00Z",
  "updated_by": {
    "id": "usr_01HY...",
    "display_name": "Tran B"
  }
}
```

### 13.2. Audit log endpoint

Chuẩn chung:

```text
GET /api/v1/audit-logs?resource_type=sales_order&resource_id=so_01HX...
```

Response item:

```json
{
  "id": "aud_01HX...",
  "resource_type": "sales_order",
  "resource_id": "so_01HX...",
  "action": "RESERVE_STOCK",
  "actor": {
    "id": "usr_01HX...",
    "display_name": "Kho A"
  },
  "occurred_at": "2026-04-24T08:40:00Z",
  "before": {
    "status": "CONFIRMED"
  },
  "after": {
    "status": "RESERVED"
  },
  "reason": null,
  "trace_id": "trc_01HX..."
}
```

### 13.3. Các mutation bắt buộc audit

- tạo/sửa/hủy master data quan trọng,
- tạo/sửa/duyệt PR/PO,
- nhận hàng,
- QC pass/fail/hold,
- stock movement,
- stock adjustment,
- reserve/release stock,
- pick/pack/handover,
- nhận hàng hoàn,
- phân loại hàng hoàn,
- tạo/duyệt subcontract order,
- chuyển NVL/bao bì sang nhà máy,
- duyệt mẫu,
- nghiệm thu hàng gia công,
- thay đổi giá/chiết khấu,
- thay đổi quyền user.

---

## 14. Idempotency standard

### 14.1. Header

```text
Idempotency-Key: 9f90a4f7-073e-4c8d-97e2-5ce3c4cf2c2b
```

### 14.2. Áp dụng bắt buộc

| API | Lý do |
|---|---|
| POST /sales-orders | tránh tạo trùng đơn |
| POST /sales-orders/{id}/reserve-stock | tránh giữ tồn 2 lần |
| POST /goods-receipts | tránh nhập hàng 2 lần |
| POST /qc-inspections/{id}/release | tránh release 2 lần |
| POST /shipment-manifests/{id}/handover | tránh bàn giao 2 lần |
| POST /return-receipts | tránh nhận hàng hoàn 2 lần |
| POST /stock-adjustments/{id}/post | tránh điều chỉnh tồn 2 lần |
| POST /subcontract-orders/{id}/receive-finished-goods | tránh nhận thành phẩm 2 lần |

### 14.3. Idempotency conflict

Nếu cùng key nhưng payload khác:

```json
{
  "success": false,
  "error": {
    "code": "IDEMPOTENCY_CONFLICT",
    "message": "Idempotency key đã được dùng với payload khác.",
    "details": {
      "idempotency_key": "9f90a4f7-073e-4c8d-97e2-5ce3c4cf2c2b"
    }
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

---

## 15. Concurrency và optimistic locking

### 15.1. Row version

Resource dễ bị nhiều người sửa cùng lúc phải có `version` hoặc `row_version`.

```json
{
  "id": "po_01HX...",
  "status": "DRAFT",
  "version": 3
}
```

Khi update:

```json
{
  "version": 3,
  "expected_status": "DRAFT",
  "note": "Update delivery date"
}
```

Nếu version lệch:

```json
{
  "success": false,
  "error": {
    "code": "RESOURCE_VERSION_CONFLICT",
    "message": "Dữ liệu đã được người khác cập nhật. Vui lòng tải lại trước khi lưu.",
    "details": {
      "current_version": 4,
      "submitted_version": 3
    }
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 15.2. Không dùng frontend stale data để quyết định

Ví dụ frontend nhìn thấy batch còn 10, nhưng lúc submit chỉ còn 4. Backend phải tính lại.

---

## 16. File upload, attachment và document API

### 16.1. Loại file liên quan Phase 1

- COA/MSDS/CO/CQ,
- chứng từ giao hàng,
- hình ảnh bao bì/lỗi hàng,
- biên bản bàn giao,
- phiếu nhập/xuất,
- ảnh hàng hoàn,
- mẫu duyệt gia công,
- hợp đồng/đơn đặt nhà máy,
- biên bản QC.

### 16.2. Upload flow khuyến nghị

```text
POST /api/v1/files/presign-upload
PUT  <presigned_url>
POST /api/v1/attachments
```

### 16.3. Presign request

```json
{
  "file_name": "coa-vitc-active.pdf",
  "content_type": "application/pdf",
  "file_size": 124587,
  "purpose": "QC_DOCUMENT"
}
```

### 16.4. Attachment create

```json
{
  "file_id": "file_01HX...",
  "resource_type": "qc_inspection",
  "resource_id": "qci_01HX...",
  "label": "COA nguyên liệu"
}
```

### 16.5. Attachment security

- Không public file trực tiếp.
- Download qua signed URL có hạn.
- Check permission trước khi cấp download URL.
- Log download với file nhạy cảm nếu cần.

---

## 17. Approval API standard

### 17.1. Các resource có approval

- Purchase request.
- Purchase order.
- Stock adjustment.
- Discount vượt ngưỡng.
- QC release exception.
- Subcontract order.
- Material transfer to factory nếu giá trị lớn.
- Payment milestone nếu Phase 1 mở finance cơ bản.

### 17.2. Approval action

```text
POST /api/v1/{resource}/{id}/submit
POST /api/v1/{resource}/{id}/approve
POST /api/v1/{resource}/{id}/reject
POST /api/v1/{resource}/{id}/cancel
```

### 17.3. Approve request

```json
{
  "comment": "Đã kiểm tra đủ chứng từ và ngân sách."
}
```

### 17.4. Reject request

```json
{
  "reason_code": "MISSING_DOCUMENT",
  "comment": "Thiếu báo giá NCC thứ 2."
}
```

### 17.5. Approval trail endpoint

```text
GET /api/v1/purchase-orders/{id}/approval-trail
```

Response:

```json
{
  "success": true,
  "data": [
    {
      "step_no": 1,
      "role": "PURCHASING_MANAGER",
      "status": "APPROVED",
      "actor": {
        "id": "usr_01HX...",
        "display_name": "Manager A"
      },
      "acted_at": "2026-04-24T09:00:00Z",
      "comment": "OK"
    }
  ],
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

---

## 18. Stock, batch và inventory API standard

### 18.1. Nguyên tắc tối thượng

Không có API nào update tồn kho trực tiếp kiểu:

```text
PATCH /api/v1/stocks/{id}
```

Mọi thay đổi tồn phải đi qua chứng từ hoặc movement.

### 18.2. Stock view API

```text
GET /api/v1/inventory/stock-balances?warehouse_id=wh_01&sku_id=sku_01
```

Response item:

```json
{
  "sku_id": "sku_01HX...",
  "sku_code": "SERUM-VITC-30ML",
  "warehouse_id": "wh_01HX...",
  "warehouse_code": "MAIN",
  "batch_id": "bat_01HX...",
  "batch_no": "SVC-260424-01",
  "expiry_date": "2028-04-24",
  "qc_status": "PASS",
  "physical_qty": "120",
  "reserved_qty": "30",
  "qc_hold_qty": "0",
  "available_qty": "90",
  "uom_code": "PCS"
}
```

### 18.3. Stock ledger API

```text
GET /api/v1/inventory/stock-movements?sku_id=sku_01&batch_id=bat_01
```

Movement item:

```json
{
  "id": "stm_01HX...",
  "movement_no": "STM-20260424-0001",
  "movement_type": "GOODS_RECEIPT",
  "sku_id": "sku_01HX...",
  "batch_id": "bat_01HX...",
  "warehouse_id": "wh_01HX...",
  "qty_in": "100",
  "qty_out": "0",
  "uom_code": "PCS",
  "source_type": "goods_receipt",
  "source_id": "grn_01HX...",
  "occurred_at": "2026-04-24T08:30:00Z"
}
```

### 18.4. Stock movement types chuẩn Phase 1

```text
GOODS_RECEIPT
QC_RELEASE_TO_AVAILABLE
QC_FAIL_TO_REJECTED
SALES_RESERVE
SALES_RESERVE_RELEASE
SALES_PICK
SALES_ISSUE
RETURN_RECEIPT
RETURN_TO_AVAILABLE
RETURN_TO_REJECTED
SUBCONTRACT_MATERIAL_ISSUE
SUBCONTRACT_FINISHED_GOODS_RECEIPT
INTERNAL_TRANSFER_OUT
INTERNAL_TRANSFER_IN
STOCK_COUNT_ADJUSTMENT
MANUAL_ADJUSTMENT_APPROVED
```

### 18.5. Batch rules

API liên quan mỹ phẩm phải luôn có batch khi giao dịch với hàng tồn.

Áp dụng cho:

- nhập hàng,
- QC,
- xuất kho,
- pick/pack,
- handover,
- hàng hoàn,
- chuyển NVL sang nhà máy,
- nhận thành phẩm gia công.

Nếu thiếu batch:

```json
{
  "success": false,
  "error": {
    "code": "BATCH_REQUIRED",
    "message": "Mã lô là bắt buộc cho giao dịch tồn kho.",
    "details": {
      "sku_code": "SERUM-VITC-30ML"
    }
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

---

## 19. Scan API standard cho kho, đóng hàng, bàn giao và hàng hoàn

### 19.1. Nguyên tắc scan UX/API

Scan API phải:

- phản hồi nhanh,
- trả kết quả rõ đúng/sai,
- không bắt user đọc nhiều chữ,
- có `scan_result_code`,
- ghi scan event log,
- không làm thay đổi nghiệp vụ lớn nếu chưa confirm.

### 19.2. Scan request chuẩn

```json
{
  "barcode": "SO-20260424-0001",
  "context": "SHIPMENT_HANDOVER",
  "location_id": "loc_01HX...",
  "device_id": "scanner_01",
  "client_scan_at": "2026-04-24T10:00:00Z"
}
```

### 19.3. Scan response chuẩn

```json
{
  "success": true,
  "data": {
    "scan_result_code": "MATCHED",
    "message": "Đơn thuộc manifest và đã đóng gói.",
    "resource_type": "shipment",
    "resource_id": "shp_01HX...",
    "resource_no": "SHP-20260424-0012",
    "current_status": "PACKED",
    "next_action": "ADD_TO_HANDOVER_BATCH"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 19.4. Scan result codes

```text
MATCHED
NOT_FOUND
DUPLICATE_SCAN
WRONG_MANIFEST
NOT_PACKED
ALREADY_HANDED_OVER
QC_HOLD
BATCH_EXPIRED
MISSING_REQUIRED_STEP
LOCATION_MISMATCH
```

### 19.5. Bàn giao ĐVVC scan endpoint

```text
POST /api/v1/shipment-manifests/{manifest_id}/scan-package
POST /api/v1/shipment-manifests/{manifest_id}/confirm-handover
```

`scan-package` chỉ ghi nhận scan và validate.  
`confirm-handover` mới đổi trạng thái manifest/shipment.

### 19.6. Hàng hoàn scan endpoint

```text
POST /api/v1/return-receipts/scan-return-package
POST /api/v1/return-receipts
POST /api/v1/return-receipts/{id}/inspect
POST /api/v1/return-receipts/{id}/dispose
```

---

## 20. Module API catalog Phase 1

Đây là catalog API ở mức contract direction. Chi tiết từng schema sẽ nằm trong OpenAPI file.

### 20.1. Auth & user context

```text
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
GET  /api/v1/me
GET  /api/v1/me/permissions
```

### 20.2. Master data

```text
GET    /api/v1/products
POST   /api/v1/products
GET    /api/v1/products/{id}
PATCH  /api/v1/products/{id}
POST   /api/v1/products/{id}/activate
POST   /api/v1/products/{id}/deactivate

GET    /api/v1/materials
POST   /api/v1/materials
GET    /api/v1/materials/{id}
PATCH  /api/v1/materials/{id}

GET    /api/v1/skus
POST   /api/v1/skus
GET    /api/v1/skus/{id}
PATCH  /api/v1/skus/{id}

GET    /api/v1/suppliers
POST   /api/v1/suppliers
GET    /api/v1/suppliers/{id}
PATCH  /api/v1/suppliers/{id}

GET    /api/v1/customers
POST   /api/v1/customers
GET    /api/v1/customers/{id}
PATCH  /api/v1/customers/{id}

GET    /api/v1/warehouses
POST   /api/v1/warehouses
GET    /api/v1/warehouses/{id}
PATCH  /api/v1/warehouses/{id}

GET    /api/v1/warehouse-locations
POST   /api/v1/warehouse-locations
PATCH  /api/v1/warehouse-locations/{id}

GET    /api/v1/uoms
POST   /api/v1/uoms
GET    /api/v1/price-lists
POST   /api/v1/price-lists
```

### 20.3. Purchase

```text
GET    /api/v1/purchase-requests
POST   /api/v1/purchase-requests
GET    /api/v1/purchase-requests/{id}
PATCH  /api/v1/purchase-requests/{id}
POST   /api/v1/purchase-requests/{id}/submit
POST   /api/v1/purchase-requests/{id}/approve
POST   /api/v1/purchase-requests/{id}/reject
POST   /api/v1/purchase-requests/{id}/cancel

GET    /api/v1/purchase-orders
POST   /api/v1/purchase-orders
GET    /api/v1/purchase-orders/{id}
PATCH  /api/v1/purchase-orders/{id}
POST   /api/v1/purchase-orders/{id}/submit
POST   /api/v1/purchase-orders/{id}/approve
POST   /api/v1/purchase-orders/{id}/reject
POST   /api/v1/purchase-orders/{id}/send-to-supplier
POST   /api/v1/purchase-orders/{id}/cancel
POST   /api/v1/purchase-orders/{id}/close
```

### 20.4. Receiving / nhập kho

```text
GET    /api/v1/goods-receipts
POST   /api/v1/goods-receipts
GET    /api/v1/goods-receipts/{id}
PATCH  /api/v1/goods-receipts/{id}
POST   /api/v1/goods-receipts/{id}/submit-qc
POST   /api/v1/goods-receipts/{id}/cancel

POST   /api/v1/goods-receipts/{id}/lines/{line_id}/attach-batch
POST   /api/v1/goods-receipts/{id}/attachments
```

### 20.5. QC

```text
GET    /api/v1/qc-inspections
POST   /api/v1/qc-inspections
GET    /api/v1/qc-inspections/{id}
PATCH  /api/v1/qc-inspections/{id}
POST   /api/v1/qc-inspections/{id}/pass
POST   /api/v1/qc-inspections/{id}/fail
POST   /api/v1/qc-inspections/{id}/hold
POST   /api/v1/qc-inspections/{id}/request-recheck
GET    /api/v1/qc-inspections/{id}/approval-trail

GET    /api/v1/qc-checklists
POST   /api/v1/qc-checklists
```

### 20.6. Inventory/WMS

```text
GET    /api/v1/inventory/stock-balances
GET    /api/v1/inventory/stock-movements
GET    /api/v1/inventory/batches
GET    /api/v1/inventory/batches/{id}
POST   /api/v1/inventory/batches/{id}/hold
POST   /api/v1/inventory/batches/{id}/release-hold

GET    /api/v1/stock-counts
POST   /api/v1/stock-counts
GET    /api/v1/stock-counts/{id}
POST   /api/v1/stock-counts/{id}/submit
POST   /api/v1/stock-counts/{id}/approve-adjustment
POST   /api/v1/stock-counts/{id}/post-adjustment

GET    /api/v1/stock-adjustments
POST   /api/v1/stock-adjustments
POST   /api/v1/stock-adjustments/{id}/submit
POST   /api/v1/stock-adjustments/{id}/approve
POST   /api/v1/stock-adjustments/{id}/post
```

### 20.7. Warehouse daily board / shift closing

```text
GET    /api/v1/warehouse-daily-board
GET    /api/v1/warehouse-shifts
POST   /api/v1/warehouse-shifts/open
GET    /api/v1/warehouse-shifts/{id}
POST   /api/v1/warehouse-shifts/{id}/close
POST   /api/v1/warehouse-shifts/{id}/reconcile
GET    /api/v1/warehouse-shifts/{id}/exceptions
```

### 20.8. Sales order / OMS

```text
GET    /api/v1/sales-orders
POST   /api/v1/sales-orders
GET    /api/v1/sales-orders/{id}
PATCH  /api/v1/sales-orders/{id}
POST   /api/v1/sales-orders/{id}/confirm
POST   /api/v1/sales-orders/{id}/reserve-stock
POST   /api/v1/sales-orders/{id}/release-reservation
POST   /api/v1/sales-orders/{id}/cancel
GET    /api/v1/sales-orders/{id}/stock-allocation
```

### 20.9. Pick/pack

```text
GET    /api/v1/pick-tasks
POST   /api/v1/pick-tasks/generate
GET    /api/v1/pick-tasks/{id}
POST   /api/v1/pick-tasks/{id}/start
POST   /api/v1/pick-tasks/{id}/scan-item
POST   /api/v1/pick-tasks/{id}/complete

GET    /api/v1/pack-tasks
POST   /api/v1/pack-tasks/{id}/scan-item
POST   /api/v1/pack-tasks/{id}/complete
```

### 20.10. Shipping / bàn giao ĐVVC

```text
GET    /api/v1/shipments
GET    /api/v1/shipments/{id}
POST   /api/v1/shipments/{id}/mark-ready-for-handover
POST   /api/v1/shipments/{id}/mark-delivery-failed
POST   /api/v1/shipments/{id}/mark-delivered

GET    /api/v1/shipment-manifests
POST   /api/v1/shipment-manifests
GET    /api/v1/shipment-manifests/{id}
POST   /api/v1/shipment-manifests/{id}/add-shipments
POST   /api/v1/shipment-manifests/{id}/scan-package
POST   /api/v1/shipment-manifests/{id}/confirm-handover
POST   /api/v1/shipment-manifests/{id}/report-missing-package
POST   /api/v1/shipment-manifests/{id}/cancel
```

### 20.11. Returns / hàng hoàn

```text
GET    /api/v1/return-receipts
POST   /api/v1/return-receipts
GET    /api/v1/return-receipts/{id}
POST   /api/v1/return-receipts/scan-return-package
POST   /api/v1/return-receipts/{id}/inspect
POST   /api/v1/return-receipts/{id}/dispose-to-available
POST   /api/v1/return-receipts/{id}/dispose-to-rejected
POST   /api/v1/return-receipts/{id}/send-to-lab
POST   /api/v1/return-receipts/{id}/close
```

### 20.12. Subcontract manufacturing / gia công ngoài

```text
GET    /api/v1/subcontract-orders
POST   /api/v1/subcontract-orders
GET    /api/v1/subcontract-orders/{id}
PATCH  /api/v1/subcontract-orders/{id}
POST   /api/v1/subcontract-orders/{id}/submit
POST   /api/v1/subcontract-orders/{id}/approve
POST   /api/v1/subcontract-orders/{id}/confirm-factory
POST   /api/v1/subcontract-orders/{id}/record-deposit
POST   /api/v1/subcontract-orders/{id}/issue-materials
POST   /api/v1/subcontract-orders/{id}/submit-sample
POST   /api/v1/subcontract-orders/{id}/approve-sample
POST   /api/v1/subcontract-orders/{id}/reject-sample
POST   /api/v1/subcontract-orders/{id}/start-mass-production
POST   /api/v1/subcontract-orders/{id}/receive-finished-goods
POST   /api/v1/subcontract-orders/{id}/report-factory-defect
POST   /api/v1/subcontract-orders/{id}/accept
POST   /api/v1/subcontract-orders/{id}/close
```

### 20.13. Reports / export jobs

```text
GET    /api/v1/reports/inventory-aging
GET    /api/v1/reports/near-expiry-stock
GET    /api/v1/reports/sales-summary
GET    /api/v1/reports/warehouse-shift-summary
POST   /api/v1/export-jobs
GET    /api/v1/export-jobs/{id}
GET    /api/v1/export-jobs/{id}/download
```

---

## 21. Detailed contract examples

### 21.1. Create sales order

```text
POST /api/v1/sales-orders
```

Request:

```json
{
  "channel": "B2B",
  "customer_id": "cus_01HX...",
  "warehouse_id": "wh_01HX...",
  "order_date": "2026-04-24",
  "price_list_id": "pl_01HX...",
  "items": [
    {
      "sku_id": "sku_01HX...",
      "qty": "10",
      "uom_code": "PCS",
      "unit_price": "250000",
      "discount_amount": "0"
    }
  ],
  "note": "Giao trong ngày nếu đủ hàng"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "so_01HX...",
    "sales_order_no": "SO-20260424-0001",
    "status": "DRAFT",
    "total_amount": "2500000",
    "currency": "VND"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 21.2. Reserve stock

```text
POST /api/v1/sales-orders/{id}/reserve-stock
```

Request:

```json
{
  "strategy": "FEFO",
  "allow_partial": false,
  "note": "Reserve cho đơn đã xác nhận"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "sales_order_id": "so_01HX...",
    "previous_status": "CONFIRMED",
    "current_status": "RESERVED",
    "allocations": [
      {
        "sku_id": "sku_01HX...",
        "sku_code": "SERUM-VITC-30ML",
        "batch_id": "bat_01HX...",
        "batch_no": "SVC-260424-01",
        "expiry_date": "2028-04-24",
        "reserved_qty": "10",
        "uom_code": "PCS"
      }
    ]
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 21.3. Scan package into manifest

```text
POST /api/v1/shipment-manifests/{manifest_id}/scan-package
```

Request:

```json
{
  "barcode": "SHP-20260424-0012",
  "scan_location_id": "loc_handover_01",
  "device_id": "scanner-warehouse-01",
  "client_scan_at": "2026-04-24T10:12:00Z"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "scan_result_code": "MATCHED",
    "shipment_id": "shp_01HX...",
    "shipment_no": "SHP-20260424-0012",
    "sales_order_no": "SO-20260424-0001",
    "current_status": "PACKED",
    "manifest_id": "man_01HX...",
    "manifest_no": "MAN-GHN-20260424-001",
    "scanned_count": 33,
    "expected_count": 40,
    "next_action": "CONTINUE_SCAN"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 21.4. Confirm handover

```text
POST /api/v1/shipment-manifests/{manifest_id}/confirm-handover
```

Request:

```json
{
  "carrier_staff_name": "Nhân viên GHN A",
  "carrier_staff_phone": "0900000000",
  "handover_note": "Đã bàn giao đủ số lượng theo manifest.",
  "signature_file_id": "file_01HX..."
}
```

Response:

```json
{
  "success": true,
  "data": {
    "manifest_id": "man_01HX...",
    "manifest_no": "MAN-GHN-20260424-001",
    "previous_status": "READY_FOR_HANDOVER",
    "current_status": "HANDED_OVER",
    "expected_count": 40,
    "scanned_count": 40,
    "handover_at": "2026-04-24T10:30:00Z"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 21.5. Receive return package

```text
POST /api/v1/return-receipts
```

Request:

```json
{
  "source": "CARRIER_RETURN",
  "carrier_id": "car_01HX...",
  "tracking_no": "GHN123456789",
  "original_sales_order_no": "SO-20260420-0009",
  "received_warehouse_id": "wh_01HX...",
  "received_location_id": "loc_return_01",
  "items": [
    {
      "sku_id": "sku_01HX...",
      "batch_no": "SVC-260424-01",
      "qty": "1",
      "uom_code": "PCS"
    }
  ],
  "received_note": "Hàng hoàn từ shipper"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "id": "rr_01HX...",
    "return_receipt_no": "RR-20260424-0001",
    "status": "RECEIVED_PENDING_INSPECTION"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 21.6. Inspect return

```text
POST /api/v1/return-receipts/{id}/inspect
```

Request:

```json
{
  "inspection_result": "USABLE",
  "condition_notes": "Sản phẩm còn nguyên seal, bao bì không móp.",
  "photos": ["file_01HX..."],
  "inspected_lines": [
    {
      "line_id": "rrl_01HX...",
      "result": "USABLE",
      "qty": "1"
    }
  ]
}
```

Response:

```json
{
  "success": true,
  "data": {
    "return_receipt_id": "rr_01HX...",
    "previous_status": "RECEIVED_PENDING_INSPECTION",
    "current_status": "INSPECTED_USABLE",
    "recommended_action": "DISPOSE_TO_AVAILABLE"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 21.7. Issue materials to subcontract factory

```text
POST /api/v1/subcontract-orders/{id}/issue-materials
```

Request:

```json
{
  "factory_id": "sup_01HX...",
  "issue_warehouse_id": "wh_main_01",
  "transfer_note": "Chuyển NVL/bao bì cho đợt gia công serum C.",
  "items": [
    {
      "material_id": "mat_01HX...",
      "batch_id": "bat_01HX...",
      "qty": "50",
      "uom_code": "KG"
    },
    {
      "material_id": "pkg_01HX...",
      "batch_id": "bat_01HY...",
      "qty": "5000",
      "uom_code": "PCS"
    }
  ],
  "attachments": ["file_01HX..."]
}
```

Response:

```json
{
  "success": true,
  "data": {
    "subcontract_order_id": "sco_01HX...",
    "subcontract_order_no": "SCO-20260424-0001",
    "material_transfer_no": "MTF-20260424-0001",
    "current_status": "MATERIALS_ISSUED_TO_FACTORY",
    "stock_movements_created": 2
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

---

## 22. Status enum chuẩn Phase 1

### 22.1. Common document status

```text
DRAFT
SUBMITTED
APPROVED
REJECTED
CANCELLED
CLOSED
```

### 22.2. Purchase order status

```text
DRAFT
SUBMITTED
APPROVED
SENT_TO_SUPPLIER
PARTIALLY_RECEIVED
FULLY_RECEIVED
CLOSED
CANCELLED
```

### 22.3. Goods receipt status

```text
DRAFT
RECEIVED_PENDING_QC
QC_IN_PROGRESS
QC_PASSED
QC_FAILED
POSTED_TO_STOCK
CANCELLED
```

### 22.4. QC inspection status

```text
DRAFT
IN_PROGRESS
HOLD
PASSED
FAILED
RECHECK_REQUIRED
CANCELLED
```

### 22.5. Sales order status

```text
DRAFT
CONFIRMED
RESERVED
PICKING
PICKED
PACKING
PACKED
HANDED_OVER
DELIVERED
PARTIALLY_RETURNED
RETURNED
CLOSED
CANCELLED
```

### 22.6. Shipment manifest status

```text
DRAFT
READY_FOR_HANDOVER
HANDOVER_IN_PROGRESS
HANDED_OVER
PARTIALLY_HANDED_OVER
CANCELLED
```

### 22.7. Return receipt status

```text
RECEIVED_PENDING_INSPECTION
INSPECTION_IN_PROGRESS
INSPECTED_USABLE
INSPECTED_UNUSABLE
DISPOSED_TO_AVAILABLE
DISPOSED_TO_REJECTED
SENT_TO_LAB
CLOSED
CANCELLED
```

### 22.8. Subcontract order status

```text
DRAFT
SUBMITTED
APPROVED
FACTORY_CONFIRMED
DEPOSIT_RECORDED
MATERIALS_ISSUED_TO_FACTORY
SAMPLE_SUBMITTED
SAMPLE_APPROVED
SAMPLE_REJECTED
MASS_PRODUCTION_STARTED
FINISHED_GOODS_RECEIVED
QC_IN_PROGRESS
ACCEPTED
REJECTED_WITH_FACTORY_ISSUE
FINAL_PAYMENT_READY
CLOSED
CANCELLED
```

---

## 23. OpenAPI file organization

### 23.1. Folder structure

```text
api/
  openapi.yaml
  paths/
    auth.yaml
    masterdata-products.yaml
    masterdata-suppliers.yaml
    purchase-orders.yaml
    goods-receipts.yaml
    qc-inspections.yaml
    inventory.yaml
    warehouse-shifts.yaml
    sales-orders.yaml
    pick-pack.yaml
    shipments.yaml
    returns.yaml
    subcontract-orders.yaml
    reports.yaml
  components/
    schemas/
      common.yaml
      error.yaml
      pagination.yaml
      audit.yaml
      approval.yaml
      masterdata.yaml
      purchase.yaml
      inventory.yaml
      sales.yaml
      shipping.yaml
      returns.yaml
      subcontract.yaml
    parameters.yaml
    responses.yaml
    securitySchemes.yaml
```

### 23.2. OpenAPI version

```yaml
openapi: 3.1.0
info:
  title: MyPham ERP API
  version: 1.0.0
servers:
  - url: https://erp.company.com/api/v1
    description: Production
  - url: https://staging-erp.company.com/api/v1
    description: Staging
```

### 23.3. Security scheme

```yaml
components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
security:
  - bearerAuth: []
```

---

## 24. Common schema examples

### 24.1. Error schema

```yaml
ErrorResponse:
  type: object
  required: [success, error, meta]
  properties:
    success:
      type: boolean
      const: false
    error:
      type: object
      required: [code, message]
      properties:
        code:
          type: string
          example: INSUFFICIENT_AVAILABLE_STOCK
        message:
          type: string
          example: Tồn khả dụng không đủ để giữ hàng.
        details:
          type: object
          additionalProperties: true
    meta:
      $ref: '#/components/schemas/ResponseMeta'
```

### 24.2. Response meta

```yaml
ResponseMeta:
  type: object
  required: [trace_id]
  properties:
    trace_id:
      type: string
      example: trc_01HXABCDEF
```

### 24.3. Pagination

```yaml
Pagination:
  type: object
  required: [page, page_size, total_items, total_pages]
  properties:
    page:
      type: integer
      minimum: 1
    page_size:
      type: integer
      minimum: 1
      maximum: 500
    total_items:
      type: integer
      minimum: 0
    total_pages:
      type: integer
      minimum: 0
```

### 24.4. Money

```yaml
MoneyAmount:
  type: string
  pattern: '^[-]?[0-9]+(\\.[0-9]+)?$'
  example: '125000'
```

### 24.5. Quantity

```yaml
Quantity:
  type: string
  pattern: '^[-]?[0-9]+(\\.[0-9]+)?$'
  example: '10.5'
```

### 24.6. User reference

```yaml
UserRef:
  type: object
  required: [id, display_name]
  properties:
    id:
      type: string
      example: usr_01HXABCDEF
    display_name:
      type: string
      example: Nguyen Van A
```

---

## 25. API documentation và generated client

### 25.1. Source of truth

OpenAPI contract là source of truth cho API public/internal frontend.

Backend không được tạo endpoint mới mà không cập nhật OpenAPI.

Frontend không được gọi endpoint không có trong OpenAPI.

### 25.2. Generated client

Frontend dùng generated client từ OpenAPI để có:

- type-safe request,
- type-safe response,
- auto-complete schema,
- giảm lỗi field name,
- giảm tự đoán error payload.

### 25.3. Codegen rules

- Không sửa file generated bằng tay.
- Nếu contract sai, sửa OpenAPI rồi generate lại.
- Generated client để trong thư mục riêng, ví dụ:

```text
frontend/src/generated/api/
```

- Business hook bọc generated client đặt ở:

```text
frontend/src/modules/<module>/api/
```

Ví dụ:

```text
frontend/src/modules/sales/api/useSalesOrders.ts
```

### 25.4. Backend contract validation

CI nên có bước:

- validate OpenAPI syntax,
- lint OpenAPI naming,
- generate client thử,
- chạy contract tests nếu có.

---

## 26. Async job và export API

### 26.1. Khi nào dùng async job?

Dùng async job cho:

- export báo cáo lớn,
- tính KPI nặng,
- sync dữ liệu hãng vận chuyển,
- import file lớn,
- data migration batch,
- reconciliation cuối ngày nếu cần chạy nền.

### 26.2. Create export job

```text
POST /api/v1/export-jobs
```

Request:

```json
{
  "report_type": "INVENTORY_AGING",
  "format": "xlsx",
  "filters": {
    "warehouse_id": "wh_01HX...",
    "as_of_date": "2026-04-24"
  }
}
```

Response:

```json
{
  "success": true,
  "data": {
    "job_id": "job_01HX...",
    "status": "QUEUED"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

### 26.3. Job status

```text
GET /api/v1/export-jobs/{id}
```

Response:

```json
{
  "success": true,
  "data": {
    "job_id": "job_01HX...",
    "status": "COMPLETED",
    "progress_percent": 100,
    "download_url_expires_at": "2026-04-24T11:30:00Z"
  },
  "meta": {
    "trace_id": "trc_01HX..."
  }
}
```

---

## 27. Event/outbox API visibility

### 27.1. Event không nhất thiết expose ra frontend

Domain event/outbox chủ yếu là backend internal.

Ví dụ event:

```text
SalesOrderReserved
ShipmentHandedOver
ReturnReceiptInspected
QCInspectionPassed
SubcontractMaterialsIssued
SubcontractFinishedGoodsReceived
WarehouseShiftClosed
```

### 27.2. Event log read-only cho admin

Có thể có endpoint read-only cho admin/debug:

```text
GET /api/v1/system/outbox-events
GET /api/v1/system/outbox-events/{id}
POST /api/v1/system/outbox-events/{id}/retry
```

Chỉ System Admin/Tech Admin được dùng.

---

## 28. Rate limit và scan performance

### 28.1. Scan API target

Scan API nên tối ưu để phản hồi nhanh.

Target nội bộ:

| API type | Target |
|---|---:|
| Scan validate | < 300ms p95 trong LAN/cloud ổn định |
| List admin | < 800ms p95 |
| Detail resource | < 500ms p95 |
| Action mutation | < 1000ms p95 với transaction thường |
| Export/report lớn | async |

### 28.2. Rate limit

Rate limit theo:

- user,
- IP/device,
- endpoint,
- tenant/company nếu multi-company sau này.

Scan API không nên bị rate-limit quá thấp vì thao tác kho cần nhanh, nhưng phải phát hiện bất thường.

---

## 29. Security baseline API

### 29.1. Input validation

Tất cả input phải validate:

- required field,
- enum,
- decimal format,
- date format,
- max length,
- reference exists,
- status condition,
- permission.

### 29.2. SQL injection prevention

- Không build SQL bằng string concat từ query params.
- Sort/filter phải whitelist.
- Dùng parameterized query qua pgx/sqlc.

### 29.3. Sensitive field protection

Các field nhạy cảm:

- cost,
- margin,
- supplier price,
- salary/payroll nếu Phase sau,
- commission,
- discount approval reason,
- personal phone/email nếu role không cần.

Backend kiểm soát field, không giao cho frontend tự giấu.

### 29.4. PII audit

Đối với dữ liệu khách hàng/cá nhân, log không nên ghi tràn lan phone/email đầy đủ. Có thể mask trong log:

```text
098****789
```

### 29.5. Upload security

- validate content type,
- giới hạn file size,
- scan virus nếu hạ tầng hỗ trợ,
- không execute file upload,
- không public bucket.

---

## 30. Testing standard cho API contract

### 30.1. Contract test

Mỗi endpoint quan trọng phải có test:

- request schema đúng,
- response schema đúng,
- error response đúng,
- permission guard đúng,
- invalid state transition đúng.

### 30.2. API test case minimum

Với mỗi module:

1. Create valid resource.
2. Create invalid resource.
3. Get list with pagination.
4. Get detail.
5. Update allowed field.
6. Try forbidden update.
7. Submit/approve/reject nếu có.
8. Invalid state transition.
9. Permission denied.
10. Audit log created.

### 30.3. ERP critical API test

Bắt buộc test sâu:

- reserve stock không âm tồn,
- batch QC hold không được xuất,
- batch expired không được xuất nếu rule cấm,
- scan package sai manifest báo lỗi,
- confirm handover thiếu package bị chặn hoặc tạo exception theo rule,
- return usable chuyển về available đúng movement,
- return unusable chuyển rejected/lab đúng movement,
- subcontract issue materials tạo stock movement đúng,
- receive subcontract goods tạo GRN/QC đúng,
- idempotency không tạo trùng chứng từ.

---

## 31. OpenAPI lint checklist

Trước khi merge contract:

- [ ] Path dùng kebab-case.
- [ ] Resource dùng plural noun.
- [ ] Action endpoint dùng verb rõ nghiệp vụ.
- [ ] Request schema có required fields.
- [ ] Response dùng envelope chuẩn.
- [ ] Error response được khai báo.
- [ ] Status enum được mô tả.
- [ ] Permission/security được mô tả.
- [ ] Pagination có schema chung.
- [ ] Decimal không dùng float.
- [ ] Date/time format rõ.
- [ ] Example request/response đủ để frontend hiểu.
- [ ] Không có field mơ hồ kiểu `data1`, `type`, `status2`.
- [ ] Không expose internal table/column name vô nghĩa với business.

---

## 32. Frontend integration rules

### 32.1. Frontend không gọi fetch tự do

Không viết rải rác:

```ts
fetch('/api/v1/sales-orders')
```

Phải đi qua generated client hoặc module API wrapper.

### 32.2. Error handling

Frontend phải hiển thị theo error envelope:

- `VALIDATION_ERROR`: show field errors.
- `FORBIDDEN`: show permission message.
- `INVALID_STATE_TRANSITION`: show status conflict + reload suggestion.
- `INSUFFICIENT_AVAILABLE_STOCK`: show stock detail nếu có quyền.
- `PACKAGE_SCAN_MISMATCH`: show scan error mạnh, âm thanh/cảnh báo nếu màn hình scan.

### 32.3. Mutation invalidation

Sau mutation:

- invalidate list/detail liên quan,
- refresh dashboard counter nếu cần,
- không tự sửa local cache cho nghiệp vụ quá phức tạp nếu dễ sai.

Ví dụ after `reserve-stock`:

- invalidate sales order detail,
- invalidate stock balance,
- invalidate pick task list,
- invalidate warehouse board.

---

## 33. API contract governance

### 33.1. Ai được thay API contract?

Chỉ các vai trò sau được đề xuất/chốt thay đổi:

- Tech Lead,
- Backend Lead,
- Frontend Lead,
- BA/PM liên quan module,
- Product Owner/Business Owner nếu thay đổi nghiệp vụ.

### 33.2. Quy trình thay đổi API

```text
1. Tạo API Change Request
2. Nêu lý do thay đổi
3. Xác định breaking/non-breaking
4. Review với FE/BE/QA
5. Update OpenAPI
6. Generate client test
7. Update test case
8. Merge
```

### 33.3. Breaking change policy

Nếu breaking change trong Phase 1:

- phải có lý do nghiệp vụ rõ,
- phải thông báo frontend/QA,
- phải update UAT nếu ảnh hưởng flow,
- không merge sát ngày UAT/go-live nếu không khẩn cấp.

---

## 34. Common anti-patterns cần cấm

### 34.1. Cấm patch status trực tiếp

Không dùng:

```text
PATCH /api/v1/sales-orders/{id}
{ "status": "HANDED_OVER" }
```

### 34.2. Cấm update tồn trực tiếp

Không dùng:

```text
PATCH /api/v1/stock-balances/{id}
{ "available_qty": "100" }
```

### 34.3. Cấm frontend truyền actor nghiệp vụ

Không dùng:

```json
{
  "approved_by": "usr_123"
}
```

### 34.4. Cấm error response mỗi nơi một kiểu

Không dùng:

```json
{ "error": "bad request" }
```

hoặc:

```json
{ "message": "fail" }
```

### 34.5. Cấm trả float cho tiền/số lượng

Không dùng:

```json
{
  "qty": 0.1,
  "price": 99999.99
}
```

### 34.6. Cấm endpoint mơ hồ

Không dùng:

```text
POST /api/v1/process
POST /api/v1/updateStatus
POST /api/v1/doAction
```

---

## 35. Definition of Done cho API endpoint

Một API endpoint Phase 1 chỉ được coi là xong khi:

- [ ] Có trong OpenAPI.
- [ ] Có request schema.
- [ ] Có response schema.
- [ ] Có success example.
- [ ] Có error examples cho lỗi chính.
- [ ] Có permission guard.
- [ ] Có validation.
- [ ] Có transaction nếu mutation liên quan nhiều bảng.
- [ ] Có audit log nếu là mutation quan trọng.
- [ ] Có idempotency nếu là mutation dễ double-submit.
- [ ] Có test happy path.
- [ ] Có test invalid input.
- [ ] Có test forbidden.
- [ ] Có test invalid state nếu có state machine.
- [ ] Có generated client chạy được ở frontend.
- [ ] Không expose field không có quyền.
- [ ] Log có trace_id.

---

## 36. Definition of Done cho OpenAPI contract

OpenAPI contract Phase 1 được coi là đủ khi:

- [ ] Bao phủ toàn bộ endpoint Phase 1 đã chốt trong PRD/SRS.
- [ ] Bao phủ workflow thực tế: kho daily board, nội quy nhập/xuất/đóng/hàng hoàn, bàn giao ĐVVC, gia công ngoài.
- [ ] Có schema common chuẩn.
- [ ] Có status enum chuẩn.
- [ ] Có error code catalog.
- [ ] Có pagination/filter/sort convention.
- [ ] Có auth/security scheme.
- [ ] Có examples đủ cho FE build màn hình.
- [ ] Validate bằng OpenAPI lint.
- [ ] Generate frontend client thành công.
- [ ] QA có thể dùng làm cơ sở viết API/UAT tests.

---

## 37. Lộ trình triển khai API contract

### Step 1: Chốt API skeleton

Tạo OpenAPI với:

- auth,
- common schemas,
- error,
- pagination,
- master data,
- inventory,
- purchase,
- QC,
- sales,
- shipping,
- returns,
- subcontract.

### Step 2: Review với FE/BE/QA

Review các điểm:

- endpoint có đủ cho màn hình 08 không,
- schema có đủ field không,
- action endpoint có bám Process Flow không,
- error code có đủ cho UX không.

### Step 3: Generate client thử

Frontend generate client và thử build:

- list screen,
- detail screen,
- create form,
- action approve/handover/scan.

### Step 4: Backend implement theo module

Ưu tiên:

1. Auth/user context.
2. Master data.
3. Inventory/batch/stock ledger.
4. Purchase/receiving/QC.
5. Sales/reserve/pick/pack.
6. Shipping/manifest/handover.
7. Returns.
8. Subcontract.
9. Reports/export.

### Step 5: Contract test + UAT mapping

Map API endpoint với UAT scenario trong file 09.

---

## 38. Câu chốt

API của ERP không chỉ là đường truyền dữ liệu.

API chính là **hàng rào kỷ luật** giữa:

- người dùng và nghiệp vụ,
- frontend và backend,
- dữ liệu thật và thao tác sai,
- workflow trên giấy và vận hành ngoài kho.

Với ERP mỹ phẩm, API Phase 1 phải khóa được 5 chuyện sống còn:

1. Không sai tồn.
2. Không bán batch chưa được phép.
3. Không bàn giao đơn thiếu kiểm soát.
4. Không xử lý hàng hoàn mơ hồ.
5. Không để gia công ngoài rời khỏi kiểm soát NVL, mẫu, QC và thanh toán.

Nếu API contract làm chắc, backend Go và frontend Next.js sẽ đi cùng một nhịp, đội dev không tự đoán, QA test được, và ERP có nền để scale.

