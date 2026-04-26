# 12_ERP_Go_Coding_Standards_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Backend:** Go  
**Kiến trúc tham chiếu:** Modular Monolith + PostgreSQL + REST/OpenAPI  
**Phiên bản:** v1.0  
**Mục tiêu:** Chuẩn hóa cách viết code backend để hệ thống ERP không bị rời rạc, khó bảo trì, sai transaction, sai tồn kho, sai batch, sai audit log.

---

## 1. Mục tiêu của tài liệu

Tài liệu này dùng để khóa chuẩn kỹ thuật cho đội backend Go khi xây ERP Phase 1.

Nó trả lời các câu hỏi:

- Code Go viết theo cấu trúc nào?
- Module tách ranh giới ra sao?
- API đặt tên thế nào?
- Response lỗi trả về ra sao?
- Transaction xử lý thế nào?
- Stock ledger, batch, QC, giao hàng, hàng hoàn phải code theo nguyên tắc nào?
- Dev có được update tồn kho trực tiếp không?
- Khi nào phải audit log?
- Khi nào phải dùng outbox/event/job?
- Test thế nào mới được merge?
- Code review checklist gồm gì?

Một câu chốt:

> ERP không chết vì thiếu framework. ERP chết vì mỗi người code một kiểu, transaction lỏng, dữ liệu sửa tay, tồn kho sai, và audit log không đủ dấu vết.

---

## 2. Phạm vi áp dụng

Áp dụng cho toàn bộ backend Go của Phase 1, gồm các module:

- Authentication / Authorization
- Master Data
- Inventory / Warehouse
- Purchase
- QC / QA
- Production / Subcontract Manufacturing
- Sales Order
- Shipping / Carrier Handover
- Returns
- Finance basic
- Reporting basic
- Workflow / Approval
- Audit Log
- Notification / Background Jobs

Không áp dụng trực tiếp cho frontend UI, nhưng mọi API backend phải hỗ trợ tốt cho UI/UX đã thiết kế ở tài liệu screen/wireframe.

---

## 3. Stack kỹ thuật đã chốt

### 3.1 Backend

```text
Language: Go
Architecture: Modular Monolith
API: REST + OpenAPI
Database: PostgreSQL
Cache: Redis
Queue/Event: NATS hoặc RabbitMQ
Storage: S3-compatible / MinIO
Migration: golang-migrate
DB Access: pgx + sqlc hoặc pgx raw query có chuẩn
Container: Docker
```

### 3.2 Công cụ bắt buộc

```text
gofmt
goimports
golangci-lint
go test
go vet
OpenAPI validation
SQL migration check
```

### 3.3 Công cụ khuyên dùng

```text
sqlc
mockery hoặc gomock
testify
zap hoặc zerolog
```

---

## 4. Nguyên tắc không được phá

Đây là các luật nền. Vi phạm là phải sửa trước khi merge.

### 4.1 Không update tồn kho trực tiếp

Không được viết kiểu:

```sql
UPDATE inventory_balance SET qty = qty - 10 WHERE sku_id = ...
```

Mọi thay đổi tồn kho phải đi qua:

```text
stock movement / stock ledger
```

Ví dụ movement type:

```text
INBOUND_RECEIPT
QC_RELEASE
QC_REJECT
SALES_RESERVE
SALES_PICK
SALES_ISSUE
SHIPMENT_HANDOVER
RETURN_RECEIPT
RETURN_DISPOSITION_USABLE
RETURN_DISPOSITION_DAMAGED
PRODUCTION_ISSUE
PRODUCTION_RECEIPT
SUBCONTRACT_MATERIAL_TRANSFER
SUBCONTRACT_FINISHED_GOODS_RECEIPT
STOCK_ADJUSTMENT
STOCK_COUNT_GAIN
STOCK_COUNT_LOSS
```

Lý do: tồn kho là mạch máu. Sửa trực tiếp là mất dấu vết, không truy được ai làm gì.

---

### 4.2 Không bỏ qua batch / lot / expiry với hàng mỹ phẩm

Với nguyên liệu, bao bì quan trọng, bán thành phẩm, thành phẩm, hệ thống phải ưu tiên batch.

Các nghiệp vụ này bắt buộc có batch nếu item được cấu hình `batch_tracking = true`:

- nhập kho
- QC đầu vào
- cấp nguyên vật liệu cho sản xuất/gia công
- nhập thành phẩm
- xuất bán
- hàng hoàn
- điều chỉnh tồn
- kiểm kê

Không được cho backend tự “cho qua” vì UI chưa gửi batch.

---

### 4.3 Không bán hàng chưa QC pass

Nếu batch đang:

```text
HOLD
FAIL
QUARANTINE
EXPIRED
BLOCKED
```

thì backend phải chặn các hành động:

- reserve stock
- pick
- pack
- shipment handover
- invoice/close order nếu cần

Chỉ batch có trạng thái hợp lệ mới được xuất bán:

```text
QC_PASS / RELEASED
```

---

### 4.4 Không trust dữ liệu từ frontend

Frontend chỉ là lớp nhập liệu. Backend phải tự validate lại:

- quyền user
- trạng thái chứng từ
- trạng thái batch
- số lượng tồn khả dụng
- rule approval
- hạn dùng
- discount limit
- business rule theo module

Không dùng logic kiểu:

```text
UI đã disable nút nên backend không cần check.
```

Đây là tư duy nguy hiểm.

---

### 4.5 Mọi giao dịch quan trọng phải có audit log

Bắt buộc audit log với:

- tạo/sửa/hủy PO
- nhận hàng
- QC pass/fail/hold
- nhập/xuất/chuyển kho
- stock adjustment
- sản xuất/gia công
- sales order
- discount override
- shipment handover
- hàng hoàn
- công nợ/thu chi
- approval/rejection
- thay đổi master data quan trọng
- phân quyền người dùng

Audit log phải trả lời được:

```text
Ai làm?
Làm lúc nào?
Từ IP/device nào nếu có?
Trên object nào?
Trước là gì?
Sau là gì?
Lý do là gì?
```

---

### 4.6 Không hard delete chứng từ nghiệp vụ

Không xóa vật lý các chứng từ sau:

- PO
- GRN
- QC inspection
- stock movement
- sales order
- shipment
- return
- production order
- subcontract order
- payment record
- approval request

Chỉ dùng:

```text
cancelled / voided / reversed / archived
```

Nếu cần sửa sai, tạo chứng từ điều chỉnh hoặc reverse transaction.

---

### 4.7 Mọi state transition phải được kiểm soát

Không được update trạng thái tùy ý kiểu:

```go
order.Status = "DELIVERED"
```

Phải đi qua function hoặc service có rule:

```go
order.MarkDelivered(ctx, actor, deliveredAt)
```

Hoặc application service:

```go
salesService.MarkOrderDelivered(ctx, command)
```

Mỗi transition phải validate:

- trạng thái hiện tại có được chuyển không
- user có quyền không
- dữ liệu liên quan có hợp lệ không
- có cần audit log/event không

---

## 5. Cấu trúc thư mục chuẩn

Codebase backend dùng cấu trúc sau:

```text
cmd/
  api/
    main.go
  worker/
    main.go

internal/
  shared/
    auth/
    authorization/
    approval/
    audit/
    config/
    database/
    errors/
    events/
    files/
    logger/
    middleware/
    pagination/
    response/
    transaction/
    validation/

  modules/
    masterdata/
      handler/
      application/
      domain/
      repository/
      dto/
      events/
      queries/

    inventory/
      handler/
      application/
      domain/
      repository/
      dto/
      events/
      queries/

    purchase/
    qc/
    production/
    sales/
    shipping/
    returns/
    finance/
    reporting/

migrations/
  000001_init.up.sql
  000001_init.down.sql

api/
  openapi.yaml

docs/
  architecture/
  decisions/
```

---

## 6. Layering rule

Mỗi module phải đi theo lớp:

```text
handler
  ↓
application service
  ↓
domain
  ↓
repository
  ↓
database
```

### 6.1 Handler

Handler chỉ làm:

- parse request
- validate format cơ bản
- lấy actor/user context
- gọi application service
- map response

Handler không được:

- viết business rule sâu
- mở transaction trực tiếp nếu không cần
- query database trực tiếp
- gọi repository của module khác
- tự tính tồn kho

Ví dụ tốt:

```go
func (h *SalesOrderHandler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    actor := auth.ActorFromContext(ctx)

    var req ConfirmOrderRequest
    if err := request.DecodeJSON(r, &req); err != nil {
        response.Error(w, errors.BadRequest("INVALID_JSON", "Dữ liệu gửi lên không hợp lệ"))
        return
    }

    result, err := h.service.ConfirmOrder(ctx, actor, req.ToCommand())
    if err != nil {
        response.Error(w, err)
        return
    }

    response.OK(w, result)
}
```

---

### 6.2 Application service

Application service là nơi điều phối use case.

Nó làm:

- kiểm quyền nghiệp vụ
- mở transaction
- gọi domain logic
- gọi repository
- ghi audit log
- phát event/outbox
- gọi service contract của module khác nếu cần

Ví dụ:

```go
func (s *Service) ConfirmOrder(ctx context.Context, actor Actor, cmd ConfirmOrderCommand) (*ConfirmOrderResult, error) {
    if err := s.authz.Require(actor, PermissionSalesOrderConfirm); err != nil {
        return nil, err
    }

    var result *ConfirmOrderResult

    err := s.tx.Run(ctx, func(ctx context.Context) error {
        order, err := s.orders.GetForUpdate(ctx, cmd.OrderID)
        if err != nil {
            return err
        }

        if err := order.CanConfirm(); err != nil {
            return err
        }

        if err := s.inventory.ReserveStock(ctx, ReserveStockCommand{
            RefType: "SALES_ORDER",
            RefID:   order.ID,
            Lines:   order.ToReservationLines(),
            ActorID: actor.ID,
        }); err != nil {
            return err
        }

        order.MarkConfirmed(actor.ID)

        if err := s.orders.Save(ctx, order); err != nil {
            return err
        }

        if err := s.audit.Record(ctx, actor, "SALES_ORDER_CONFIRMED", order.ID, nil, order); err != nil {
            return err
        }

        if err := s.events.Enqueue(ctx, "sales.order.confirmed", order.ID, order.ToEventPayload()); err != nil {
            return err
        }

        result = &ConfirmOrderResult{OrderID: order.ID, Status: order.Status}
        return nil
    })

    if err != nil {
        return nil, err
    }

    return result, nil
}
```

---

### 6.3 Domain

Domain chứa:

- entity
- value object
- state transition
- invariant
- business rule thuần

Domain không được phụ thuộc:

- HTTP
- database
- framework
- logger
- repository implementation

Ví dụ:

```go
func (b Batch) CanBeSold(now time.Time) error {
    if b.QCStatus != QCStatusPassed {
        return domainerr.New("BATCH_NOT_RELEASED", "Batch chưa được QC pass")
    }

    if b.ExpiryDate.Before(now) || b.ExpiryDate.Equal(now) {
        return domainerr.New("BATCH_EXPIRED", "Batch đã hết hạn")
    }

    return nil
}
```

---

### 6.4 Repository

Repository chỉ làm:

- query database
- save entity
- map DB row ↔ domain object

Repository không được:

- quyết định nghiệp vụ
- kiểm quyền
- ghi audit log
- phát event

---

## 7. Module boundary rule

### 7.1 Không chọc thẳng dữ liệu module khác

Module `sales` không được tự update bảng của `inventory`.

Sai:

```go
salesRepo.UpdateInventoryBalance(...)
```

Đúng:

```go
inventoryService.ReserveStock(ctx, cmd)
```

### 7.2 Giao tiếp giữa module

Trong modular monolith, module giao tiếp bằng:

```text
1. Application service interface
2. Domain event / outbox event
3. Read model/query service được cho phép
```

Ví dụ:

```go
type InventoryReservationService interface {
    ReserveStock(ctx context.Context, cmd ReserveStockCommand) error
    ReleaseReservation(ctx context.Context, cmd ReleaseReservationCommand) error
}
```

### 7.3 Query đọc dữ liệu liên module

Đọc cross-module được phép nếu là read-only query và có service/query riêng.

Ví dụ:

```go
inventoryQueries.GetAvailableStockBySKU(ctx, skuID)
```

Không được lạm dụng join lung tung trong code nghiệp vụ.

---

## 8. Naming convention

### 8.1 Package name

Dùng chữ thường, ngắn, không underscore.

Đúng:

```go
package inventory
package purchase
package stockledger
```

Sai:

```go
package inventory_module
package InventoryService
```

### 8.2 File name

Dùng snake_case cho file:

```text
sales_order_handler.go
reserve_stock_command.go
stock_movement_repository.go
```

### 8.3 Struct / type

Dùng PascalCase:

```go
type SalesOrder struct {}
type StockMovement struct {}
type QCInspection struct {}
```

### 8.4 Function / method

Dùng động từ rõ nghĩa:

```go
CreatePurchaseOrder
SubmitForApproval
ApproveInspection
ReserveStock
ReleaseReservation
MarkAsHandedOver
```

Không dùng tên mơ hồ:

```go
Process
Handle
DoAction
UpdateStatus
```

Trừ khi nằm trong interface đặc thù.

### 8.5 ID fields

Dùng nhất quán:

```go
OrderID
SKUId // không khuyên
SKU_ID // không khuyên
```

Chốt chuẩn:

```go
SKUID
BatchID
WarehouseID
CustomerID
SupplierID
```

Trong JSON dùng snake_case:

```json
{
  "sku_id": "...",
  "batch_id": "...",
  "warehouse_id": "..."
}
```

### 8.6 Status enum

Dùng typed constants, không dùng string rải rác.

```go
type SalesOrderStatus string

const (
    SalesOrderStatusDraft      SalesOrderStatus = "DRAFT"
    SalesOrderStatusConfirmed  SalesOrderStatus = "CONFIRMED"
    SalesOrderStatusReserved   SalesOrderStatus = "RESERVED"
    SalesOrderStatusPicked     SalesOrderStatus = "PICKED"
    SalesOrderStatusPacked     SalesOrderStatus = "PACKED"
    SalesOrderStatusHandedOver SalesOrderStatus = "HANDED_OVER"
    SalesOrderStatusDelivered  SalesOrderStatus = "DELIVERED"
    SalesOrderStatusClosed     SalesOrderStatus = "CLOSED"
    SalesOrderStatusCancelled  SalesOrderStatus = "CANCELLED"
)
```

---

## 9. API standard

### 9.1 REST endpoint convention

Dùng plural nouns.

```text
GET    /api/v1/products
POST   /api/v1/products
GET    /api/v1/products/{id}
PATCH  /api/v1/products/{id}

POST   /api/v1/purchase-orders/{id}/submit
POST   /api/v1/purchase-orders/{id}/approve
POST   /api/v1/purchase-orders/{id}/cancel

POST   /api/v1/sales-orders/{id}/confirm
POST   /api/v1/sales-orders/{id}/reserve
POST   /api/v1/sales-orders/{id}/pick
POST   /api/v1/sales-orders/{id}/pack
```

### 9.2 Không dùng endpoint mơ hồ

Sai:

```text
POST /api/v1/process
POST /api/v1/update-status
POST /api/v1/do-action
```

Đúng:

```text
POST /api/v1/shipments/{id}/handover
POST /api/v1/returns/{id}/inspect
POST /api/v1/qc-inspections/{id}/pass
```

### 9.3 Response thành công

Chuẩn response:

```json
{
  "success": true,
  "data": {
    "id": "so_123",
    "status": "CONFIRMED"
  }
}
```

List response:

```json
{
  "success": true,
  "data": [
    {
      "id": "so_123",
      "order_no": "SO-2026-0001"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 250
  }
}
```

### 9.4 Response lỗi

Chuẩn lỗi:

```json
{
  "success": false,
  "code": "INSUFFICIENT_STOCK",
  "message": "Tồn khả dụng không đủ",
  "details": {
    "sku_code": "SERUM-VITC-30ML",
    "requested_qty": 10,
    "available_qty": 6
  },
  "request_id": "req_abc123"
}
```

### 9.5 HTTP status mapping

```text
200 OK: thành công
201 Created: tạo mới thành công
400 Bad Request: request sai format/business input invalid
401 Unauthorized: chưa đăng nhập/token sai
403 Forbidden: không có quyền
404 Not Found: không tìm thấy resource
409 Conflict: conflict trạng thái/dữ liệu/tồn kho/idempotency
422 Unprocessable Entity: validate nghiệp vụ chi tiết
500 Internal Server Error: lỗi hệ thống
```

### 9.6 Idempotency

Các API tạo transaction nên hỗ trợ idempotency key:

- tạo sales order từ external channel
- xác nhận thanh toán
- webhook shipping
- import dữ liệu
- handover shipment nếu scan nhiều lần
- receive return

Header:

```text
Idempotency-Key: <unique-key>
```

Nếu request lặp lại, backend trả lại kết quả cũ, không tạo giao dịch trùng.

---

## 10. Validation standard

### 10.1 Validate nhiều tầng

```text
Handler: format, required field cơ bản
Application: permission, status, use case rule
Domain: invariant cốt lõi
Database: constraint cuối cùng
```

### 10.2 Required fields

Không để request thiếu dữ liệu quan trọng rồi “default bừa”.

Ví dụ nhập kho nguyên liệu bắt buộc:

```text
supplier_id
purchase_order_id nếu nhập theo PO
warehouse_id
sku_id/material_id
quantity
unit
batch_no nếu batch tracking
manufacturing_date nếu có
expiry_date nếu có hạn dùng
received_by
received_at
```

### 10.3 Business validation ví dụ

- Không cho nhận hàng PO đã closed/cancelled.
- Không cho QC pass nếu inspection chưa đủ chỉ tiêu bắt buộc.
- Không cho reserve stock nếu batch chưa release.
- Không cho handover nếu shipment chưa packed.
- Không cho nhận hàng hoàn nếu không có đơn gốc hoặc reason hợp lệ, trừ case manual override có quyền.
- Không cho xuất nguyên vật liệu cho gia công nếu chưa có subcontract order được duyệt.

---

## 11. Error handling standard

### 11.1 Không trả error thô từ database ra frontend

Sai:

```go
return err
```

nếu err là:

```text
duplicate key value violates unique constraint...
```

Đúng:

```go
return errors.Conflict("DUPLICATE_SKU_CODE", "Mã SKU đã tồn tại")
```

### 11.2 Error type chuẩn

Tạo package `internal/shared/errors`.

Các loại error:

```text
BadRequest
Unauthorized
Forbidden
NotFound
Conflict
Validation
Internal
```

Mỗi error có:

```go
type AppError struct {
    Code       string
    Message    string
    Details    map[string]any
    HTTPStatus int
    Cause      error
}
```

### 11.3 Error code naming

Dùng UPPER_SNAKE_CASE:

```text
INSUFFICIENT_STOCK
BATCH_NOT_RELEASED
ORDER_ALREADY_HANDED_OVER
PURCHASE_ORDER_CLOSED
QC_INSPECTION_INCOMPLETE
APPROVAL_REQUIRED
DISCOUNT_LIMIT_EXCEEDED
```

### 11.4 Không dùng panic cho business error

Panic chỉ dùng cho lỗi không thể tiếp tục ở startup/config. Business rule phải return error.

---

## 12. Transaction standard

### 12.1 Mọi use case nhiều bước phải chạy trong transaction

Ví dụ bắt buộc transaction:

- Confirm sales order + reserve stock + audit log + event
- Receive goods + create stock movement + create QC hold + audit log
- QC pass + release stock + update batch + audit log + event
- Pick/pack/handover shipment
- Receive return + inspect + disposition + stock movement
- Subcontract material transfer + stock issue + handover document
- Finished goods receipt from factory + QC hold + stock movement

### 12.2 Transaction helper

Dùng helper chung:

```go
tx.Run(ctx, func(ctx context.Context) error {
    // all repository calls use transaction-bound context
    return nil
})
```

Không tự mở/commit/rollback rải rác.

### 12.3 Locking

Những nghiệp vụ cần lock:

- reserve stock
- stock movement against same SKU/batch/warehouse
- QC status transition
- order status transition
- shipment handover
- return disposition
- end-of-day reconciliation

Dùng:

```sql
SELECT ... FOR UPDATE
```

hoặc advisory lock khi cần.

### 12.4 Không gọi external API trong DB transaction

Sai:

```text
begin tx
update order
call carrier API
commit
```

Đúng:

```text
begin tx
update order
write outbox event
commit
worker reads outbox
call carrier API
update sync result
```

Lý do: external API chậm/lỗi có thể giữ lock, gây nghẽn hệ thống.

---

## 13. Database / SQL standard

### 13.1 PostgreSQL naming

Table/column dùng snake_case:

```sql
sales_orders
sales_order_lines
stock_movements
qc_inspections
shipment_handovers
```

Primary key:

```sql
id uuid primary key
```

Các cột chuẩn:

```sql
created_at timestamptz not null
created_by uuid null
updated_at timestamptz not null
updated_by uuid null
cancelled_at timestamptz null
cancelled_by uuid null
```

### 13.2 Không dùng float cho tiền và số lượng chính xác

Sai:

```sql
price double precision
qty float
```

Đúng:

```sql
price numeric(18,2)
qty numeric(18,6)
```

### 13.3 Money/currency

Cột tiền:

```sql
amount numeric(18,2) not null
currency_code varchar(3) not null default 'VND'
```

### 13.4 Unique constraints

Phải có unique cho mã nghiệp vụ:

```sql
sku_code
batch_no + sku_id
warehouse_code
supplier_code
customer_code
purchase_order_no
sales_order_no
shipment_no
```

### 13.5 Foreign key

Dữ liệu nghiệp vụ quan trọng nên có FK nếu không gây vấn đề vận hành.

Ví dụ:

```sql
sales_order_lines.sales_order_id -> sales_orders.id
stock_movements.warehouse_id -> warehouses.id
stock_movements.sku_id -> skus.id
```

### 13.6 Migration rule

- Mỗi migration có `.up.sql` và `.down.sql` nếu khả thi.
- Không sửa file migration đã chạy ở staging/prod.
- Mọi schema change phải qua PR.
- Data migration lớn phải có script riêng và dry-run.

---

## 14. Stock ledger coding standard

### 14.1 Stock movement là nguồn sự thật

Tồn kho được tính từ ledger hoặc bảng balance được cập nhật từ ledger.

Không có stock movement thì không có thay đổi tồn.

### 14.2 Stock movement fields tối thiểu

```text
id
movement_no
movement_type
warehouse_id
location_id
sku_id
batch_id
qty_delta
unit
ref_type
ref_id
reason_code
created_at
created_by
```

### 14.3 Balance update

Balance có thể dùng để đọc nhanh:

```text
physical_qty
reserved_qty
qc_hold_qty
available_qty
```

Nhưng balance chỉ được update bởi stock movement service.

### 14.4 Available stock formula

```text
available_stock = physical_stock - reserved_stock - qc_hold_stock - blocked_stock
```

Không tự tính mỗi nơi một kiểu.

### 14.5 FEFO/FIFO

Với mỹ phẩm, xuất bán ưu tiên:

```text
FEFO: First Expired, First Out
```

Nếu item không có hạn dùng thì dùng FIFO hoặc rule cấu hình.

### 14.6 Cảnh báo cận date

Backend nên hỗ trợ query:

```text
near_expiry_days = configured threshold by SKU/category
```

Không hardcode 30 ngày nếu chưa cấu hình.

---

## 15. QC coding standard

### 15.1 QC status enum

```text
NOT_REQUIRED
PENDING
HOLD
PASS
FAIL
QUARANTINE
RELEASED
BLOCKED
```

Chốt cụ thể theo PRD/Data Dictionary. Không tự thêm status trong code nếu chưa update tài liệu.

### 15.2 QC transition

Ví dụ:

```text
PENDING -> HOLD
HOLD -> PASS
HOLD -> FAIL
PASS -> RELEASED
FAIL -> QUARANTINE hoặc REJECTED
```

Không cho:

```text
FAIL -> PASS
```

trừ khi có re-inspection record và approval.

### 15.3 QC pass phải tạo event

Khi QC pass/release:

```text
qc.inspection.passed
inventory.batch.released
```

Để module inventory/sales biết batch khả dụng.

### 15.4 Complaint liên quan batch

Nếu CSKH/returns ghi nhận complaint sản phẩm, backend phải hỗ trợ link tới:

```text
sales_order_id
sku_id
batch_id
customer_id
complaint_type
severity
```

Đây là nền cho trace lỗi sau này.

---

## 16. Sales order coding standard

### 16.1 Sales order status

```text
DRAFT
CONFIRMED
RESERVED
PICKING
PICKED
PACKING
PACKED
HANDOVER_READY
HANDED_OVER
DELIVERED
CLOSED
CANCELLED
RETURNED
PARTIALLY_RETURNED
```

### 16.2 Không tạo shipment nếu chưa reserve/pick hợp lệ

Backend phải chặn:

- shipment cho order draft
- handover cho order chưa packed
- delivered cho order chưa handed over

### 16.3 Discount rule

Nếu discount vượt ngưỡng role:

```text
APPROVAL_REQUIRED
```

Không được âm thầm cho qua.

### 16.4 Reserve stock

Confirm order có thể reserve ngay hoặc theo cấu hình. Nhưng nếu có reserve, phải:

- lock stock balance
- kiểm batch release
- kiểm expiry
- tạo reservation record
- update reserved qty qua stock service
- audit log

---

## 17. Shipping / carrier handover coding standard

Workflow thực tế của kho có bước phân khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, lấy hàng và quét mã trước khi ký bàn giao với ĐVVC. Vì vậy backend phải hỗ trợ manifest/handover rõ ràng.

### 17.1 Shipment status

```text
CREATED
PICKED
PACKED
READY_FOR_HANDOVER
HANDOVER_IN_PROGRESS
HANDED_OVER
IN_TRANSIT
DELIVERED
FAILED_DELIVERY
RETURNING
RETURNED
CANCELLED
```

### 17.2 Handover manifest

Một handover manifest gồm:

```text
manifest_id
manifest_no
carrier_id
warehouse_id
handover_date
status
expected_order_count
scanned_order_count
created_by
confirmed_by
```

### 17.3 Scan verify

API scan phải validate:

- mã đơn có tồn tại không
- đơn thuộc manifest/carrier đúng không
- đơn đã packed chưa
- đơn đã handover trước đó chưa
- có bị cancel/hold không

### 17.4 Thiếu đơn khi bàn giao

Nếu scan không đủ:

- không cho confirm manifest full
- cho trạng thái `PARTIAL_PENDING` hoặc `MISMATCH`
- ghi mismatch reason
- audit log

### 17.5 Idempotent scan

Quét lại cùng mã không được tạo double record.

Nếu đã scan:

```json
{
  "success": true,
  "data": {
    "scan_status": "ALREADY_SCANNED"
  }
}
```

---

## 18. Returns coding standard

Workflow thực tế có bước nhận hàng hoàn từ shipper, đưa vào khu hàng hoàn, quét hàng hoàn, kiểm tình trạng, phân loại còn sử dụng/không sử dụng, chuyển kho hoặc chuyển lab. Backend phải bám đúng logic này.

### 18.1 Return status

```text
RECEIVED_FROM_SHIPPER
WAITING_INSPECTION
INSPECTING
USABLE
DAMAGED
SENT_TO_LAB
RETURNED_TO_AVAILABLE_STOCK
SCRAPPED
CLOSED
```

### 18.2 Return inspection bắt buộc

Hàng hoàn không được tự động nhập lại available stock.

Phải qua:

```text
return receipt -> inspection -> disposition
```

### 18.3 Disposition

Kết quả xử lý:

```text
USABLE -> stock movement RETURN_DISPOSITION_USABLE
DAMAGED -> quarantine/damaged stock
NEED_LAB -> transfer to lab
```

### 18.4 Link order/batch

Return line phải cố gắng link tới:

```text
original_order_id
shipment_id
sku_id
batch_id
return_reason
condition
```

Nếu không xác định được batch thì cần quyền override và reason.

---

## 19. Purchase / inbound coding standard

### 19.1 Receiving flow

Nhập hàng cần đi theo:

```text
PO/Delivery note -> Receive -> Initial check -> QC hold/pass/fail -> Putaway/Available
```

### 19.2 Không cho nhập khả dụng trước QC nếu hàng cần QC

Nếu item có `qc_required = true`:

```text
received qty -> qc_hold_qty
```

Chỉ khi QC pass mới chuyển sang available.

### 19.3 Partial receipt

PO phải hỗ trợ nhận một phần:

```text
PARTIALLY_RECEIVED
RECEIVED
CLOSED
```

### 19.4 Over receipt

Nhận vượt PO chỉ được nếu có tolerance cấu hình hoặc approval.

---

## 20. Production / subcontract manufacturing coding standard

Workflow thực tế có nhánh đặt hàng với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển nguyên vật liệu/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, nhận hàng về kho, kiểm tra số lượng/chất lượng, xử lý lỗi trong vòng 3–7 ngày, thanh toán cuối. Backend phải hỗ trợ nhánh gia công ngoài, không chỉ sản xuất nội bộ.

### 20.1 Subcontract order status

```text
DRAFT
SUBMITTED
APPROVED
DEPOSIT_PAID
MATERIAL_PREPARING
MATERIAL_TRANSFERRED
SAMPLE_IN_PROGRESS
SAMPLE_APPROVED
MASS_PRODUCTION
FINISHED_GOODS_RECEIVED
QC_INSPECTION
ACCEPTED
REJECTED_OR_CLAIMED
FINAL_PAYMENT_PENDING
CLOSED
CANCELLED
```

### 20.2 Material transfer to factory

Chuyển nguyên vật liệu/bao bì sang nhà máy phải tạo:

- subcontract material transfer document
- stock movement xuất khỏi kho công ty hoặc chuyển sang kho gia công
- handover record
- file đính kèm nếu có COA/MSDS/tem phụ/hóa đơn VAT

### 20.3 Sample approval

Không cho mass production nếu sample chưa approved, trừ quyền override có lý do.

### 20.4 Finished goods receipt

Hàng từ nhà máy về phải vào QC hold trước.

Không nhập thẳng available.

### 20.5 Claim window

Nếu quy trình thực tế quy định báo lỗi trong 3–7 ngày, backend nên hỗ trợ:

```text
claim_deadline = received_at + configured_days
```

và cảnh báo quá hạn.

---

## 21. Approval coding standard

### 21.1 Approval là service chung

Không viết approval riêng lẻ trong từng module nếu cùng pattern.

Dùng shared approval service:

```text
approval_requests
approval_steps
approval_actions
```

### 21.2 Approval object

```text
object_type
object_id
requested_by
current_step
status
reason
metadata
```

### 21.3 Business object khi pending approval

Khi chờ duyệt:

```text
status = PENDING_APPROVAL
```

hoặc có field:

```text
approval_status = PENDING
```

Chặn hành động tiếp theo cho tới khi approved.

### 21.4 Reject

Reject phải yêu cầu lý do.

### 21.5 Approval audit

Mọi approve/reject/withdraw phải audit log.

---

## 22. Audit log coding standard

### 22.1 Audit log fields

```text
id
actor_id
actor_name
action
object_type
object_id
before_data
after_data
reason
ip_address
user_agent
created_at
request_id
```

### 22.2 Tránh lưu dữ liệu quá nhạy cảm

Không lưu plaintext password/token/secret vào audit.

### 22.3 Before/after

Với object lớn, có thể chỉ lưu changed fields:

```json
{
  "discount_percent": {
    "before": 5,
    "after": 15
  }
}
```

### 22.4 Audit log trong transaction

Audit log của transaction nghiệp vụ phải ghi trong cùng transaction nếu có thể.

---

## 23. Event / outbox / worker standard

### 23.1 Khi nào dùng event

Dùng event cho:

- thông báo sau giao dịch
- sync external system
- cập nhật read model
- gửi email/Zalo/SMS
- report async
- gọi carrier API
- tính KPI cuối ngày

### 23.2 Outbox table

Dùng outbox để đảm bảo event không mất:

```text
outbox_events
id
event_type
aggregate_type
aggregate_id
payload
status
attempts
available_at
created_at
processed_at
```

### 23.3 Event naming

Dùng dạng:

```text
module.entity.action
```

Ví dụ:

```text
sales.order.confirmed
inventory.stock.reserved
qc.inspection.passed
shipping.manifest.handed_over
returns.item.inspected
production.subcontract.material_transferred
```

### 23.4 Worker idempotent

Worker phải xử lý được việc chạy lại cùng event.

Không được gửi trùng payment, không tạo trùng stock movement, không update sai trạng thái.

---

## 24. Logging standard

### 24.1 Log structured JSON

Log dạng structured:

```json
{
  "level": "info",
  "request_id": "req_123",
  "actor_id": "user_456",
  "module": "shipping",
  "action": "manifest_handover",
  "message": "Manifest handed over successfully"
}
```

### 24.2 Không log dữ liệu nhạy cảm

Không log:

- password
- token
- secret
- full personal data không cần thiết
- thông tin thanh toán nhạy cảm

### 24.3 Request ID

Mọi request phải có `request_id`.

Nếu client không gửi, backend tự tạo.

### 24.4 Log level

```text
DEBUG: local/dev only
INFO: sự kiện bình thường
WARN: bất thường nhưng chưa lỗi nghiêm trọng
ERROR: lỗi cần xử lý
```

---

## 25. Security coding standard

### 25.1 Auth

- Không tự parse token rải rác.
- Dùng middleware auth chung.
- Actor context phải có user_id, role, permissions, warehouse_scope nếu có.

### 25.2 RBAC

Mọi endpoint nghiệp vụ phải check permission.

Ví dụ:

```go
s.authz.Require(actor, PermissionInventoryStockAdjust)
```

### 25.3 Warehouse scope

User kho chỉ được thao tác kho được phân quyền.

Backend phải check:

```text
actor.allowed_warehouse_ids contains warehouse_id
```

### 25.4 Cost/finance visibility

Không trả field giá vốn/biên lợi nhuận cho role không có quyền.

Không chỉ hide ở frontend.

### 25.5 Input sanitization

- Validate string length.
- Không cho upload file extension nguy hiểm.
- Escape output nếu cần.
- Dùng parameterized query, không nối SQL bằng string từ user input.

---

## 26. File upload standard

Áp dụng cho:

- COA/MSDS
- phiếu giao hàng
- biên bản bàn giao
- ảnh hàng lỗi
- chứng từ NCC
- hợp đồng
- ảnh kiểm hàng hoàn
- bằng chứng KOL nếu Phase sau

### 26.1 File metadata

```text
file_id
object_type
object_id
file_name
content_type
size
storage_key
uploaded_by
uploaded_at
```

### 26.2 Không lưu file binary trong database

DB chỉ lưu metadata. File lưu S3/MinIO.

### 26.3 Validate file

- giới hạn size
- whitelist content type
- scan malware nếu có điều kiện
- không tin file extension

---

## 27. Timezone / date standard

### 27.1 Storage

Database lưu `timestamptz`.

### 27.2 Business timezone

Business timezone mặc định:

```text
Asia/Ho_Chi_Minh
```

### 27.3 Date-only fields

Hạn dùng, ngày sản xuất có thể là `date`, không phải timestamp.

```text
manufacturing_date date
expiry_date date
```

### 27.4 End-of-day

Đóng ca/kiểm kê cuối ngày phải dùng business date theo timezone công ty, không dùng UTC date thô.

---

## 28. Testing standard

### 28.1 Test tối thiểu trước khi merge

Mỗi use case quan trọng phải có test cho:

- happy path
- permission denied
- invalid state
- invalid data
- insufficient stock
- batch not released
- transaction rollback

### 28.2 Unit test

Dùng cho domain rule.

Ví dụ:

- batch expired không bán được
- order draft không handover được
- QC fail không release được
- return damaged không nhập available được

### 28.3 Integration test

Dùng cho repository/application service có DB test.

Nên có test cho:

- reserve stock concurrent
- stock movement atomic
- receive goods + QC hold
- QC pass + release stock
- shipment scan + manifest handover

### 28.4 Transaction rollback test

Bắt buộc test các use case tiền-hàng-kho.

Ví dụ:

```text
Nếu tạo stock movement thành công nhưng audit log lỗi -> transaction rollback toàn bộ.
```

### 28.5 Test naming

```go
func TestReserveStock_WhenAvailable_ShouldCreateReservation(t *testing.T) {}
func TestReserveStock_WhenInsufficient_ShouldReturnError(t *testing.T) {}
```

---

## 29. Code review standard

Mỗi PR phải được review theo checklist sau.

### 29.1 Business rule

- Đúng PRD chưa?
- Đúng Process Flow chưa?
- Đúng Permission Matrix chưa?
- Đúng Data Dictionary chưa?
- Có xử lý exception không?

### 29.2 Transaction

- Use case nhiều bước có transaction chưa?
- Có lock dữ liệu cần lock chưa?
- Có rollback đúng không?
- Có gọi external API trong transaction không?

### 29.3 Inventory/QC

- Có update tồn trực tiếp không?
- Có tạo stock movement không?
- Có check batch/QC/expiry không?
- Có audit log không?

### 29.4 Security

- Có check permission ở backend không?
- Có check warehouse scope không?
- Có lộ giá vốn/finance data không?
- Có validate input không?

### 29.5 API

- Endpoint đúng convention không?
- Response/error đúng chuẩn không?
- OpenAPI cập nhật chưa?
- Có idempotency nếu cần không?

### 29.6 Testing

- Có test domain rule chưa?
- Có integration test cho transaction quan trọng chưa?
- Test case lỗi có đủ không?

---

## 30. Lint / format / CI rule

### 30.1 Bắt buộc trước khi merge

```bash
gofmt
goimports
golangci-lint run
go test ./...
go vet ./...
```

### 30.2 golangci-lint gợi ý

Bật tối thiểu:

```text
govet
staticcheck
unused
ineffassign
errcheck
gocritic
gosec
revive
misspell
bodyclose
```

### 30.3 Không merge nếu

- test fail
- lint fail
- migration fail
- OpenAPI mismatch
- thiếu review
- thiếu audit/transaction cho nghiệp vụ quan trọng

---

## 31. Branch / commit standard

### 31.1 Branch naming

```text
feature/inventory-stock-ledger
feature/shipping-handover-manifest
fix/qc-release-stock-bug
hotfix/order-reserve-rollback
```

### 31.2 Commit message

Dùng convention:

```text
feat(inventory): add stock movement ledger
fix(shipping): prevent duplicate scan in handover manifest
refactor(qc): move QC transition rule to domain
```

### 31.3 PR title

```text
[Inventory] Add immutable stock ledger for inbound receipt
[Shipping] Add carrier handover manifest scan API
```

---

## 32. Documentation rule

Mỗi module phải có file README ngắn:

```text
modules/inventory/README.md
```

Nội dung:

- module làm gì
- entity chính
- API chính
- event publish/consume
- transaction quan trọng
- permission liên quan
- rule không được phá

Mỗi API phải cập nhật OpenAPI.

Mỗi thay đổi business rule phải cập nhật PRD/Process/Data Dictionary nếu ảnh hưởng.

---

## 33. Workflow-specific guardrails

Đây là phần khóa riêng cho workflow thực tế hiện tại.

### 33.1 Công việc kho hằng ngày

Hệ thống phải hỗ trợ logic:

```text
nhận đơn trong ngày
xuất/nhập theo nội quy
soạn hàng/đóng gói
sắp xếp/tối ưu kho
kiểm kê cuối ngày
đối soát số liệu và báo cáo quản lý
kết thúc ca
```

Coding guardrail:

- mọi giao dịch trong ngày phải có business date/shift nếu áp dụng
- end-of-day reconciliation không được sửa ledger gốc
- chênh lệch kiểm kê phải tạo adjustment có lý do và approval nếu vượt ngưỡng

### 33.2 Nhập kho

- Nhận chứng từ giao hàng.
- Kiểm số lượng/bao bì/lô.
- Đạt thì xếp kho/ký xác nhận.
- Không đạt thì trả NCC hoặc đưa vào trạng thái xử lý.

Coding guardrail:

- inbound receipt không tự tạo available nếu QC required
- reject NCC phải có reason và attachment nếu có
- batch_no bắt buộc khi item batch-tracked

### 33.3 Xuất kho

- Làm phiếu xuất.
- Xuất kho.
- Đối chiếu số lượng thực tế.
- Ký bàn giao.
- Trưởng kho lưu phiếu.

Coding guardrail:

- pick/issue phải dựa trên phiếu xuất hợp lệ
- actual picked qty phải được ghi nhận
- mismatch phải có exception flow

### 33.4 Đóng hàng

- Nhận phiếu đơn hàng hợp lệ.
- Lọc/phân loại đơn.
- Soạn hàng theo từng đơn.
- Đóng gói và kiểm tra lại khu vực đóng hàng.
- Bàn giao ĐVVC.

Coding guardrail:

- không pack order chưa pick đủ
- không handover order chưa pack
- scan verify trước khi manifest confirm

### 33.5 Hàng hoàn

- Nhận từ shipper.
- Đưa vào khu hàng hoàn.
- Quét hàng hoàn.
- Kiểm tra tình trạng.
- Còn dùng được thì chuyển kho.
- Không dùng được thì chuyển lab/khu xử lý.

Coding guardrail:

- return không nhập available trực tiếp
- phải có inspection/disposition
- phải có ảnh/ghi chú nếu damaged

### 33.6 Gia công ngoài

- Lên đơn với nhà máy.
- Xác nhận số lượng/quy cách/mẫu mã.
- Cọc đơn/chốt thời gian.
- Chuyển NVL/bao bì.
- Làm mẫu/chốt mẫu.
- Sản xuất hàng loạt.
- Giao hàng về kho.
- Kiểm tra số lượng/chất lượng.
- Nhận hàng hoặc báo lỗi trong thời hạn.
- Thanh toán cuối.

Coding guardrail:

- material transfer phải tạo stock movement
- sample approval phải là state riêng
- finished goods từ nhà máy phải vào QC hold
- claim window phải có cấu hình/cảnh báo

---

## 34. Ví dụ chuẩn API quan trọng

### 34.1 Reserve stock

Endpoint:

```text
POST /api/v1/sales-orders/{id}/reserve
```

Request:

```json
{
  "reservation_strategy": "FEFO",
  "warehouse_id": "wh_001"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "order_id": "so_001",
    "status": "RESERVED",
    "reserved_lines": [
      {
        "sku_id": "sku_001",
        "batch_id": "batch_001",
        "reserved_qty": "10.000000"
      }
    ]
  }
}
```

Business rules:

- check order status
- check stock available
- check QC pass
- check expiry
- create reservation movement/record
- audit log

---

### 34.2 QC pass inbound batch

Endpoint:

```text
POST /api/v1/qc-inspections/{id}/pass
```

Request:

```json
{
  "result_note": "Đạt tiêu chuẩn đầu vào",
  "attachment_ids": ["file_001"]
}
```

Business rules:

- user must have QC permission
- inspection status must be HOLD/PENDING
- required checklist completed
- release stock from qc_hold to available
- audit log
- event `qc.inspection.passed`

---

### 34.3 Carrier handover scan

Endpoint:

```text
POST /api/v1/shipping-manifests/{id}/scan
```

Request:

```json
{
  "barcode": "SO-2026-0001"
}
```

Response:

```json
{
  "success": true,
  "data": {
    "scan_status": "SCANNED",
    "order_no": "SO-2026-0001",
    "scanned_count": 25,
    "expected_count": 40
  }
}
```

---

## 35. Definition of Done cho backend task

Một backend task chỉ được coi là xong khi:

```text
1. API chạy đúng happy path.
2. Business rule lỗi được xử lý.
3. Permission được check.
4. Transaction đúng.
5. Audit log có nếu là nghiệp vụ quan trọng.
6. Stock movement có nếu ảnh hưởng tồn.
7. Error response đúng chuẩn.
8. OpenAPI cập nhật.
9. Unit/integration test có đủ mức cần thiết.
10. Lint/test pass.
11. Reviewer sign-off.
```

---

## 36. Anti-pattern cấm

### 36.1 God service

Không tạo service khổng lồ:

```go
ERPService
CommonService
DataService
```

### 36.2 Business rule trong handler

Không viết logic trạng thái, tồn kho, QC trong handler.

### 36.3 Query raw lung tung

Không viết SQL raw rải rác trong handler/service mà không qua repository/query object.

### 36.4 Update status tùy tiện

Không dùng API chung kiểu:

```text
PATCH /orders/{id}/status
```

cho mọi trạng thái.

Dùng action endpoint rõ nghiệp vụ:

```text
POST /orders/{id}/confirm
POST /orders/{id}/cancel
POST /orders/{id}/pack
```

### 36.5 Không có reason khi override

Mọi override phải có:

```text
reason
actor
approval nếu cần
audit log
```

### 36.6 Thêm status không cập nhật tài liệu

Nếu thêm trạng thái mới trong code mà không cập nhật Data Dictionary/PRD/UAT, coi như sai quy trình.

---

## 37. Chuẩn bàn giao cho tech lead

Tech lead cần kiểm tra trước khi team bắt đầu sprint:

- Repo đã có folder structure chuẩn.
- Shared packages đã có skeleton.
- Error/response chuẩn đã có.
- Transaction manager đã có.
- Audit service đã có.
- Auth/RBAC middleware đã có.
- Migration pipeline đã có.
- OpenAPI file đã có.
- Module inventory có stock ledger skeleton.
- Module approval có base service.
- CI chạy lint/test.

---

## 38. Roadmap coding standard nâng cấp sau Phase 1

Sau Phase 1 có thể bổ sung:

- event sourcing sâu hơn cho stock nếu cần
- CQRS/read model cho dashboard lớn
- distributed tracing
- advanced BI pipeline
- API rate limiting theo external channel
- data lake/reporting replica
- multi-tenant/multi-brand rule nâng cao
- mobile scanner API optimization

Nhưng Phase 1 không nên phức tạp hóa quá sớm.

---

## 39. Kết luận

Backend Go cho ERP mỹ phẩm phải được viết theo tư duy:

```text
Rõ module
Chặt transaction
Không sửa trực tiếp tồn
Không bỏ qua batch/QC
Có audit log
Có state machine
Có test
Có OpenAPI
```

Nếu giữ được các chuẩn này, hệ thống sẽ không chỉ chạy được lúc demo, mà còn chịu được vận hành thật: nhập kho, xuất kho, đóng hàng, bàn giao ĐVVC, hàng hoàn, gia công ngoài, QC, đối soát cuối ngày và mở rộng về sau.

Câu chốt:

> Code ERP tốt là code biết bảo vệ sự thật của doanh nghiệp: hàng thật, tiền thật, người thật, trách nhiệm thật.

