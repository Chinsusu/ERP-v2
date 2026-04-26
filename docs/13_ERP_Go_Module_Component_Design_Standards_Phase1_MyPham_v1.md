# 13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Backend:** Go  
**Kiến trúc tham chiếu:** Modular Monolith + PostgreSQL + REST/OpenAPI  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-24  
**Mục tiêu:** Khóa chuẩn thiết kế module và component backend Go để ERP Phase 1 không bị biến thành một khối code rối, mỗi module chọc dữ liệu của nhau, sai transaction, sai tồn kho, sai batch, sai trạng thái chứng từ.

---

## 1. Tóm tắt điều hành

Tài liệu này là lớp nằm giữa:

```text
11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md
12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md
```

và code thực tế.

Nếu tài liệu kiến trúc trả lời **hệ thống dùng gì và chạy theo mô hình nào**, coding standard trả lời **dev viết code kiểu gì**, thì tài liệu này trả lời:

```text
Module nào sở hữu dữ liệu nào?
Module nào được gọi module nào?
Component dùng chung được thiết kế ra sao?
Workflow nghiệp vụ đi qua ranh giới module như thế nào?
Chỗ nào bắt buộc dùng event, chỗ nào dùng service call đồng bộ?
Cái gì tuyệt đối không được cross-module write?
```

Một câu chốt:

> ERP không hỏng vì thiếu module. ERP hỏng vì module nào cũng có thể sửa dữ liệu của module khác.

---

## 2. Phạm vi áp dụng

Áp dụng cho toàn bộ backend Go của ERP Phase 1, bao gồm các module nghiệp vụ và shared component sau:

```text
Auth / RBAC
Workflow / Approval
Master Data
Inventory / Warehouse
Purchase
QC / QA
Production / Subcontract Manufacturing
Sales Order
Shipping / Carrier Handover
Returns
Finance Basic
Reporting Basic
Audit Log
Notification / Event / Worker
Attachment / File Storage
```

Không phải tài liệu UI/UX design system. Chuẩn UI/UX sẽ được tách riêng ở tài liệu:

```text
14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md
```

Tuy nhiên, tài liệu này vẫn quy định **API contract, status, action, error code, query/filter** để frontend có thể thiết kế giao diện nhất quán.

---

## 3. Nguồn nghiệp vụ thực tế đã đưa vào chuẩn thiết kế

Chuẩn module/component này không chỉ dựa trên ERP lý thuyết. Nó được điều chỉnh theo workflow thực tế từ các file nội bộ công ty:

```text
Công-việc-hằng-ngày.pdf
Nội-Quy.pdf
Quy-trình-bàn-giao.pdf
Quy-trình-sản-xuất.pdf
```

Các điểm nghiệp vụ đã được phản ánh vào chuẩn thiết kế:

### 3.1 Kho có nhịp vận hành theo ngày

Workflow kho hiện tại có chuỗi:

```text
Tiếp nhận đơn trong ngày
→ thực hiện xuất/nhập theo bảng nội quy
→ soạn hàng và đóng gói
→ sắp xếp/tối ưu vị trí kho
→ kiểm kê tồn kho cuối ngày
→ đối soát số liệu và báo cáo quản lý
→ kết thúc ca
```

Vì vậy backend phải có component hỗ trợ:

```text
Daily warehouse workload
Shift closing
End-of-day reconciliation
Cycle count
Warehouse report snapshot
```

### 3.2 Nội quy kho chia 4 nhánh rõ

Nội quy kho hiện tại có 4 luồng chính:

```text
Nhập kho
Xuất kho
Đóng hàng
Xử lý hàng hoàn
```

Vì vậy backend không được gom tất cả thành một API `updateStock`. Phải thiết kế các module/component riêng:

```text
Receiving component
Outbound issue component
Packing component
Return inspection component
```

### 3.3 Bàn giao đơn vị vận chuyển có manifest và scan verify

Quy trình bàn giao hiện tại có logic:

```text
Phân chia khu vực để hàng
→ để theo thùng/rổ
→ đối chiếu số lượng đơn theo bảng
→ lấy hàng và quét mã trực tiếp tại hàm
→ đủ đơn thì ký xác nhận bàn giao
→ chưa đủ thì kiểm tra lại mã hoặc tìm lại trong khu vực đóng hàng
```

Vì vậy module shipping phải có:

```text
Carrier manifest
Bin/tote grouping
Scan verification
Short/missing package flow
Handover confirmation
```

### 3.4 Sản xuất có nhánh gia công ngoài

Workflow sản xuất thực tế thể hiện mô hình:

```text
Lên đơn hàng với nhà máy
→ xác nhận số lượng/quy cách/mẫu mã
→ cọc đơn và xác nhận thời gian
→ chuyển nguyên vật liệu/bao bì sang nhà máy
→ làm mẫu và chốt mẫu
→ sản xuất hàng loạt
→ giao hàng về kho
→ kiểm tra số lượng/chất lượng
→ nhận hàng hoặc báo lỗi trong 3-7 ngày
→ thanh toán lần cuối
```

Vì vậy module production Phase 1 phải hỗ trợ:

```text
Subcontract manufacturing order
Material/packaging transfer to factory
Sample approval
Mass production tracking
Factory inbound receiving
Factory defect claim window
Deposit/final payment linkage
```

---

## 4. Triết lý thiết kế module

### 4.1 Module là năng lực nghiệp vụ, không phải folder CRUD

Sai:

```text
/modules/products
/modules/orders
/modules/stocks
```

Đúng hơn:

```text
/modules/masterdata
/modules/sales
/modules/inventory
/modules/qc
/modules/shipping
/modules/returns
```

Module phải đại diện cho một **business capability**: một năng lực vận hành có dữ liệu, rule, trạng thái và trách nhiệm rõ.

### 4.2 Mỗi module phải có chủ quyền dữ liệu

Mỗi bảng dữ liệu phải có **module owner**.

Ví dụ:

```text
inventory_stock_movements → Inventory sở hữu
sales_orders → Sales sở hữu
qc_inspections → QC sở hữu
shipping_manifests → Shipping sở hữu
return_cases → Returns sở hữu
```

Module khác không được tự ý ghi trực tiếp vào bảng không thuộc module của mình.

### 4.3 Cross-module write bị cấm

Ví dụ module Sales không được làm:

```sql
UPDATE inventory_balances SET reserved_qty = reserved_qty + 10;
```

Sales muốn giữ hàng thì phải gọi contract của Inventory:

```go
InventoryService.ReserveStock(ctx, command)
```

hoặc phát command/event theo pattern được chốt.

### 4.4 Tất cả nghiệp vụ tiền-hàng-kho phải có state machine

Các object sau bắt buộc có trạng thái chuẩn:

```text
Purchase Request
Purchase Order
Goods Receipt
QC Inspection
Stock Movement
Production/Subcontract Order
Sales Order
Pick Pack Task
Shipping Manifest
Return Case
Payment Request
```

Không được dùng boolean kiểu:

```text
is_done
is_active
is_success
```

cho các nghiệp vụ nhiều trạng thái.

### 4.5 Tồn kho không phải con số, tồn kho là lịch sử movement

Tồn kho hiện tại chỉ là projection từ stock ledger.

```text
Stock Ledger → Balance Projection → Available Stock
```

Không được coi tồn kho là một field để sửa tay.

### 4.6 Workflow thật quan trọng hơn template ERP

Nếu workflow công ty có bước riêng như scan mã khi bàn giao ĐVVC, kiểm kê cuối ngày, hoặc chuyển bao bì/nguyên liệu sang nhà máy gia công, thì module phải phản ánh được. Không được ép workflow thực tế vào template CRUD đơn giản.

---

## 5. Kiến trúc module tổng thể

### 5.1 Sơ đồ module backend Phase 1

```text
                                  ┌─────────────────────┐
                                  │      Auth/RBAC       │
                                  └──────────┬──────────┘
                                             │
┌─────────────────────┐          ┌──────────▼──────────┐          ┌─────────────────────┐
│     Master Data      │◄────────►│ Workflow / Approval │◄────────►│      Audit Log      │
└──────────┬──────────┘          └──────────┬──────────┘          └─────────────────────┘
           │                                │
           ▼                                ▼
┌─────────────────────┐          ┌─────────────────────┐          ┌─────────────────────┐
│      Purchase        │─────────►│        QC/QA         │─────────►│     Inventory       │
└──────────┬──────────┘          └──────────┬──────────┘          └──────────┬──────────┘
           │                                │                                │
           ▼                                ▼                                ▼
┌─────────────────────┐          ┌─────────────────────┐          ┌─────────────────────┐
│ Production/Subcon    │─────────►│      Returns        │─────────►│      Shipping       │
└──────────┬──────────┘          └──────────┬──────────┘          └──────────┬──────────┘
           │                                │                                │
           └─────────────────────┬──────────┴─────────────────────┬──────────┘
                                 ▼                                ▼
                       ┌─────────────────────┐          ┌─────────────────────┐
                       │        Sales         │─────────►│   Finance Basic     │
                       └─────────────────────┘          └─────────────────────┘
                                 │
                                 ▼
                       ┌─────────────────────┐
                       │   Reporting Basic   │
                       └─────────────────────┘
```

### 5.2 Phân loại module

```text
Core Business Modules:
- Master Data
- Purchase
- QC/QA
- Inventory/Warehouse
- Production/Subcontract
- Sales
- Shipping
- Returns
- Finance Basic

Platform Modules:
- Auth/RBAC
- Workflow/Approval
- Audit Log
- Notification
- Attachment/File Storage
- Numbering
- Event/Outbox
- Reporting
```

---

## 6. Cấu trúc chuẩn của một module Go

Mỗi module nghiệp vụ nên có cấu trúc:

```text
internal/modules/<module_name>/
  handler/
    http_handler.go
    request.go
    response.go

  application/
    commands/
    queries/
    services.go
    ports.go

  domain/
    entity.go
    value_objects.go
    state_machine.go
    policies.go
    events.go
    errors.go

  repository/
    repository.go
    postgres_repository.go
    sql/

  readmodel/
    read_model.go
    queries.go

  worker/
    event_handlers.go

  module.go
```

Không phải module nào cũng bắt buộc có đủ `worker` hoặc `readmodel`, nhưng các module lõi như Inventory, Sales, Shipping, Returns, QC nên có.

---

## 7. Chuẩn layer trong module

### 7.1 Handler layer

Handler chỉ làm:

```text
Parse request
Validate input format cơ bản
Lấy user context
Gọi application service
Map response
```

Handler không được:

```text
Viết SQL
Mở transaction nghiệp vụ
Tự xử lý rule business phức tạp
Tự update state
Tự ghi audit log nghiệp vụ
```

### 7.2 Application layer

Application layer là nơi orchestration use case.

Nó được phép:

```text
Mở transaction
Gọi repository của module mình
Gọi contract của module khác nếu được phép
Gọi approval/audit/event component
Điều phối state transition
```

Nó không nên chứa rule domain quá chi tiết nếu rule đó có thể nằm trong domain.

### 7.3 Domain layer

Domain layer chứa:

```text
Entity
Value object
State machine
Domain policy
Domain event
Domain error
```

Ví dụ rule domain:

```text
Batch QC fail không được release stock
Sales order cancelled không được ship
Manifest chưa scan đủ không được handover
Return item inspected as unusable không được đưa về available stock
```

### 7.4 Repository layer

Repository chỉ truy cập dữ liệu thuộc module mình.

Repository không được:

```text
Join tùy tiện sang bảng module khác để ghi dữ liệu
Update bảng của module khác
Nhét business rule vào SQL nếu rule cần audit/validation rõ
```

Nếu cần đọc dữ liệu module khác cho list/report, dùng read model hoặc query service được chốt.

### 7.5 Worker/Event handler layer

Worker xử lý job hoặc event bất đồng bộ:

```text
Gửi notification
Tạo report snapshot
Sync trạng thái ngoài hệ thống
Tính KPI cuối ngày
Rebuild projection
```

Worker phải idempotent.

---

## 8. Chuẩn đặt tên module, command, query, event

### 8.1 Tên module

Dùng danh từ nghiệp vụ số ít hoặc cụm business capability:

```text
inventory
purchase
qc
production
sales
shipping
returns
finance
masterdata
```

Không dùng tên mơ hồ:

```text
manager
common_business
misc
operation
```

### 8.2 Command

Command đặt theo hành động nghiệp vụ:

```go
type ReserveStockCommand struct {}
type ReleaseBatchCommand struct {}
type CreateShippingManifestCommand struct {}
type ConfirmCarrierHandoverCommand struct {}
type InspectReturnedItemCommand struct {}
type ApproveSubcontractSampleCommand struct {}
```

### 8.3 Query

Query đặt theo mục đích đọc:

```go
type ListAvailableStockQuery struct {}
type GetOrderFulfillmentStatusQuery struct {}
type ListManifestPackagesQuery struct {}
type GetBatchTraceQuery struct {}
```

### 8.4 Event

Event đặt ở thì quá khứ:

```go
type StockReservedEvent struct {}
type QCInspectionPassedEvent struct {}
type ShipmentHandedOverEvent struct {}
type ReturnItemInspectedEvent struct {}
type SubcontractSampleApprovedEvent struct {}
```

Không dùng event kiểu mơ hồ:

```text
UpdateEvent
StockChanged
OrderEvent
```

---

## 9. Chuẩn module contract

Module expose ra ngoài qua contract, không expose repository.

Ví dụ Inventory expose:

```go
type InventoryService interface {
    ReserveStock(ctx context.Context, cmd ReserveStockCommand) (*ReservationResult, error)
    ReleaseReservation(ctx context.Context, cmd ReleaseReservationCommand) error
    IssueStock(ctx context.Context, cmd IssueStockCommand) (*StockMovementResult, error)
    ReceiveStock(ctx context.Context, cmd ReceiveStockCommand) (*StockMovementResult, error)
    GetAvailableStock(ctx context.Context, query GetAvailableStockQuery) (*AvailableStock, error)
}
```

Sales không được import:

```go
internal/modules/inventory/repository
```

Sales chỉ được dùng:

```go
internal/modules/inventory/application/ports
```

hoặc một interface được đăng ký tại composition root.

---

## 10. Ma trận sở hữu dữ liệu

| Nhóm dữ liệu | Module owner | Module được đọc | Module được ghi |
|---|---|---|---|
| SKU, item, UOM | Master Data | Tất cả module | Master Data |
| Supplier | Master Data | Purchase, Finance, QC | Master Data |
| Customer | Master Data/Sales | Sales, Shipping, Finance, CRM future | Master Data/Sales |
| Warehouse, bin/location | Master Data/Inventory | Inventory, Shipping, Production | Inventory/Master Data |
| Purchase Request | Purchase | Finance, Approval, Reporting | Purchase |
| Purchase Order | Purchase | Inventory, QC, Finance, Reporting | Purchase |
| Goods Receipt | Inventory | Purchase, QC, Finance | Inventory |
| QC Inspection | QC | Purchase, Inventory, Production, Returns | QC |
| Batch/Lot status | QC + Inventory | Sales, Shipping, Returns | QC/Inventory theo contract |
| Stock Movement | Inventory | Reporting, Finance | Inventory |
| Stock Balance Projection | Inventory | Sales, Shipping, Reporting | Inventory job/projection |
| Production/Subcontract Order | Production | Purchase, Inventory, QC, Finance | Production |
| Material Transfer to Factory | Production/Inventory | QC, Finance | Production + Inventory theo transaction/contract |
| Sales Order | Sales | Inventory, Shipping, Finance | Sales |
| Reservation | Inventory | Sales, Shipping | Inventory |
| Pick/Pack Task | Shipping/Inventory tùy implementation | Sales, Warehouse | Shipping/Inventory theo contract |
| Shipping Manifest | Shipping | Sales, Warehouse, Finance | Shipping |
| Return Case | Returns | Sales, QC, Inventory, Finance | Returns |
| Payment/AR/AP basic | Finance | Sales, Purchase, Reporting | Finance |
| Approval Request | Workflow/Approval | Tất cả module | Workflow/Approval |
| Audit Log | Audit | Tất cả module | Audit component |
| Attachment/File metadata | Attachment | Tất cả module | Attachment component |

Nguyên tắc:

```text
Một bảng chỉ có một module owner.
Một nghiệp vụ có thể chạm nhiều module, nhưng orchestration phải rõ.
Không có chuyện module nào thích update bảng nào cũng được.
```

---

## 11. Ma trận dependency giữa module

| Module | Được gọi | Không được gọi |
|---|---|---|
| Sales | Master Data, Inventory contract, Shipping contract, Workflow, Audit | Inventory repository, Finance repository |
| Inventory | Master Data, QC contract, Workflow, Audit | Sales repository, Finance repository |
| Purchase | Master Data, Inventory receiving contract, QC contract, Workflow, Finance contract | Inventory table direct update |
| QC | Master Data, Inventory status contract, Audit, Workflow | Sales repository |
| Shipping | Sales read contract, Inventory issue contract, Audit, Workflow | Direct stock balance update |
| Returns | Sales read contract, QC inspection contract, Inventory receive/dispose contract, Finance contract | Direct batch/QC table update |
| Production | Master Data, Inventory transfer contract, QC contract, Purchase/Finance contract | Direct supplier AP update |
| Finance | Sales/Purchase/Shipping/Returns read events, Audit | Direct operational state update |
| Reporting | Read models, event snapshots | Write nghiệp vụ |

Nếu module A cần đọc dữ liệu của module B:

```text
Ưu tiên 1: dùng query contract của B
Ưu tiên 2: dùng read model/snapshot đã thiết kế
Ưu tiên 3: dùng SQL join chỉ cho reporting, không cho command transaction
```

---

## 12. Shared components bắt buộc

### 12.1 Auth/RBAC component

Chức năng:

```text
Xác thực user
Load role/permission
Inject user context vào request
Kiểm tra quyền theo action/resource
```

Không nhúng permission check rải rác tùy hứng. Dùng middleware + application guard.

Ví dụ resource/action:

```text
inventory.stock_movement:create
inventory.stock_movement:approve
qc.inspection:release
shipping.manifest:handover
returns.case:inspect
```

### 12.2 Workflow/Approval component

Dùng cho các nghiệp vụ cần duyệt:

```text
PO vượt ngưỡng
Discount vượt ngưỡng
QC release batch
Stock adjustment
Payment request
Subcontract manufacturing order
KOL/sample issue future
```

Component này chỉ quản approval state, không sở hữu nghiệp vụ gốc.

Ví dụ:

```text
Purchase Order status = Submitted
Approval status = Pending Finance Approval
```

Khi approve xong, module owner mới chuyển trạng thái PO.

### 12.3 Audit Log component

Bắt buộc ghi audit cho:

```text
Master data quan trọng
Giá/bảng giá/chiết khấu
Batch/QC status
Stock movement
Stock adjustment
Sales order status
Shipping handover
Return disposition
Subcontract order state
Payment approval
Permission change
```

Audit log phải ghi:

```text
actor_id
actor_role
action
resource_type
resource_id
before_snapshot
after_snapshot
reason
request_id
ip/user_agent nếu có
created_at
```

### 12.4 Numbering component

Sinh mã chứng từ tập trung:

```text
PR-YYYYMMDD-0001
PO-YYYYMMDD-0001
GRN-YYYYMMDD-0001
QC-YYYYMMDD-0001
SO-YYYYMMDD-0001
PK-YYYYMMDD-0001
MN-YYYYMMDD-0001
RET-YYYYMMDD-0001
SUB-YYYYMMDD-0001
```

Không để từng module tự random mã.

### 12.5 Attachment component

Dùng lưu file:

```text
Phiếu giao hàng
Biên bản nhập/xuất
COA/MSDS
Ảnh hàng lỗi
Ảnh hàng hoàn
Biên bản bàn giao nhà máy
Ảnh sample
Manifest bàn giao ĐVVC
```

Module nghiệp vụ chỉ lưu `attachment_id`, không tự quản storage path.

### 12.6 Event/Outbox component

Dùng phát event đáng tin cậy sau transaction.

Ví dụ:

```text
QCInspectionPassed
StockReserved
ShipmentHandedOver
ReturnItemInspected
SubcontractGoodsReceived
```

Không publish event trực tiếp trước khi commit DB.

---

## 13. Module Master Data

### 13.1 Trách nhiệm

Master Data sở hữu dữ liệu nền:

```text
SKU / item
Material
Packaging
UOM
Warehouse
Bin/location
Supplier
Customer basic
Employee basic
Price list basic
Batch coding rule
```

### 13.2 Commands

```text
CreateItem
UpdateItem
DeactivateItem
CreateSupplier
UpdateSupplier
CreateWarehouse
CreateBinLocation
CreateUOMConversion
```

### 13.3 Queries

```text
ListItems
GetItemDetail
SearchSupplier
ListWarehouses
ListActiveSKUs
```

### 13.4 Events

```text
ItemCreated
ItemUpdated
SupplierUpdated
WarehouseCreated
```

### 13.5 Rule đặc biệt

Không cho xóa master data đã phát sinh giao dịch. Chỉ được deactivate.

```text
Item đã có stock movement → không được delete
Supplier đã có PO → không được delete
Warehouse đã có tồn → không được delete
```

---

## 14. Module Purchase

### 14.1 Trách nhiệm

Purchase sở hữu:

```text
Purchase Request
RFQ nếu có
Purchase Order
Supplier delivery schedule
Supplier invoice reference basic
```

Purchase không sở hữu stock. Hàng nhận về phải đi qua Inventory receiving và QC.

### 14.2 Commands

```text
CreatePurchaseRequest
SubmitPurchaseRequest
ApprovePurchaseRequest
CreatePurchaseOrder
SubmitPurchaseOrder
ApprovePurchaseOrder
SendPurchaseOrderToSupplier
ClosePurchaseOrder
CancelPurchaseOrder
```

### 14.3 Queries

```text
ListPurchaseOrders
GetPurchaseOrderDetail
ListPendingReceipts
GetSupplierPurchaseHistory
```

### 14.4 Events

```text
PurchaseOrderApproved
PurchaseOrderSent
PurchaseOrderClosed
PurchaseOrderCancelled
```

### 14.5 Module interactions

```text
Purchase → Master Data: supplier/item info
Purchase → Workflow: approval
Purchase → Inventory: pending receiving reference
Purchase → QC: inspection request after received
Purchase → Finance: AP/payment reference
```

### 14.6 Cấm

Purchase không được:

```text
Tự tăng tồn kho
Tự set QC pass
Tự tạo stock movement
Tự ghi payment final
```

---

## 15. Module Inventory / Warehouse

### 15.1 Trách nhiệm

Inventory là module trọng yếu nhất trong Phase 1.

Sở hữu:

```text
Goods Receipt
Stock Movement / Stock Ledger
Stock Balance Projection
Reservation
Warehouse Transfer
Stock Adjustment
Cycle Count
Warehouse Shift Closing
Bin/Tote/Location assignment
Sample/Tester/Gift stock type nếu Phase 1 cần
```

### 15.2 Commands

```text
CreateGoodsReceipt
ReceiveStockToHold
ReleaseStockAfterQC
RejectStockAfterQC
ReserveStock
ReleaseReservation
IssueStock
TransferStock
AdjustStock
StartCycleCount
SubmitCycleCount
CloseWarehouseShift
```

### 15.3 Queries

```text
GetAvailableStock
GetPhysicalStock
ListStockMovements
GetBatchStock
GetWarehouseDailySummary
GetNearExpiryStock
GetReservedStockByOrder
```

### 15.4 Events

```text
GoodsReceived
StockMoved
StockReserved
ReservationReleased
StockIssued
StockAdjusted
WarehouseShiftClosed
CycleCountSubmitted
```

### 15.5 Stock ownership rule

Inventory là module duy nhất được tạo stock movement.

Các module khác muốn thay đổi tồn phải gọi command contract.

```text
Purchase nhận hàng → Inventory.ReceiveStockToHold
QC pass → Inventory.ReleaseStockAfterQC
Sales xác nhận đơn → Inventory.ReserveStock
Shipping bàn giao → Inventory.IssueStock
Returns hàng còn dùng → Inventory.ReceiveReturnedStock
Production chuyển NVL → Inventory.TransferToFactory
```

### 15.6 Projection rule

Balance table chỉ là projection.

```text
stock_movements = source of truth
stock_balances = current projection
```

Nếu projection sai, rebuild từ ledger.

### 15.7 Daily warehouse component

Do workflow kho có kiểm kê và đối soát cuối ngày, Inventory cần component:

```text
WarehouseShift
WarehouseDailyTask
WarehouseReconciliation
```

Trạng thái gợi ý:

```text
Open
Processing
PendingReconciliation
Closed
ReopenedByAdmin
```

Shift closing nên kiểm:

```text
Đơn pending pick/pack
Manifest chưa bàn giao
Phiếu nhập chưa QC
Hàng hoàn chưa inspect
Stock adjustment chưa duyệt
Cycle count lệch chưa xử lý
```

---

## 16. Module QC / QA

### 16.1 Trách nhiệm

QC sở hữu:

```text
Inbound inspection
In-process inspection nếu Phase 1 cần
Finished goods inspection
Return inspection support
Batch release decision
Non-conformance record
CAPA basic nếu Phase 1 cần
```

### 16.2 Commands

```text
CreateInspection
SubmitInspectionResult
PassInspection
FailInspection
HoldBatch
ReleaseBatch
RejectBatch
RequestReinspection
```

### 16.3 Queries

```text
ListPendingInspections
GetInspectionDetail
ListBatchQCStatus
GetBatchTraceQC
```

### 16.4 Events

```text
QCInspectionCreated
QCInspectionPassed
QCInspectionFailed
BatchHeld
BatchReleased
BatchRejected
```

### 16.5 QC status

Chuẩn trạng thái:

```text
HOLD
PENDING_INSPECTION
PASSED
FAILED
REINSPECTION_REQUIRED
RELEASED
REJECTED
```

### 16.6 Batch rule

Hàng chưa QC pass/release thì không được available for sales.

```text
Goods Receipt → stock type HOLD
QC Passed → Inventory.ReleaseStockAfterQC
QC Failed → Inventory.RejectStockAfterQC hoặc quarantine
```

### 16.7 Cấm

QC không được tự update stock balance. QC chỉ ra quyết định chất lượng, Inventory chuyển trạng thái tồn theo contract.

---

## 17. Module Production / Subcontract Manufacturing

### 17.1 Trách nhiệm

Vì workflow thực tế có sản xuất/gia công ngoài, Production Phase 1 phải ưu tiên nhánh subcontract.

Sở hữu:

```text
Subcontract Manufacturing Order
Factory order confirmation
Material/Packaging transfer request
Factory handover document
Sample request/approval
Mass production tracking
Factory delivery schedule
Factory receiving reference
Factory claim window
```

### 17.2 Commands

```text
CreateSubcontractOrder
ConfirmFactorySpecs
RecordDeposit
RequestMaterialTransferToFactory
ConfirmFactoryHandover
RequestSample
ApproveSample
RejectSample
StartMassProduction
RecordFactoryDelivery
CloseSubcontractOrder
RaiseFactoryDefectClaim
```

### 17.3 Queries

```text
ListSubcontractOrders
GetSubcontractOrderDetail
ListFactoryPendingDeliveries
GetMaterialTransferredToFactory
GetSampleApprovalStatus
```

### 17.4 Events

```text
SubcontractOrderCreated
FactorySpecsConfirmed
MaterialTransferredToFactory
SubcontractSampleApproved
MassProductionStarted
FactoryGoodsDelivered
FactoryDefectClaimRaised
SubcontractOrderClosed
```

### 17.5 Interactions

```text
Production → Master Data: BOM/SKU/material/packaging
Production → Purchase/Finance: deposit/final payment reference
Production → Inventory: transfer NVL/bao bì to factory
Production → QC: inspect goods returned from factory
Production → Inventory: receive finished goods after QC
```

### 17.6 Rule đặc biệt

Không được coi hàng chuyển sang nhà máy là “mất khỏi hệ thống”. Phải có stock location/type:

```text
WAREHOUSE_MAIN
FACTORY_CONSIGNMENT
FACTORY_WIP
QUARANTINE
AVAILABLE
```

Khi chuyển NVL/bao bì sang nhà máy:

```text
Inventory movement: TRANSFER_TO_FACTORY
stock location: FACTORY_CONSIGNMENT
```

Khi nhận hàng thành phẩm từ nhà máy:

```text
Goods Receipt type: SUBCONTRACT_FINISHED_GOODS
QC inspection required
```

### 17.7 Claim window

Workflow thực tế có logic báo lại nhà máy trong khoảng 3-7 ngày nếu hàng không đạt. Module cần lưu:

```text
received_at
inspection_deadline_at
claim_deadline_at
factory_claim_status
factory_response
```

---

## 18. Module Sales Order

### 18.1 Trách nhiệm

Sales sở hữu:

```text
Sales Order
Order line
Price/discount snapshot
Customer snapshot
Order status
Return eligibility reference
Sales channel
```

Sales không sở hữu tồn kho, không sở hữu giao hàng, không sở hữu thu tiền chi tiết.

### 18.2 Commands

```text
CreateSalesOrder
ConfirmSalesOrder
ApplyDiscount
SubmitDiscountApproval
CancelSalesOrder
RequestReservation
ReleaseOrderReservation
MarkOrderReadyForFulfillment
CloseSalesOrder
```

### 18.3 Queries

```text
ListSalesOrders
GetSalesOrderDetail
GetOrderFulfillmentStatus
ListOrdersReadyToPick
GetCustomerOrderHistory
```

### 18.4 Events

```text
SalesOrderCreated
SalesOrderConfirmed
SalesOrderReserved
SalesOrderCancelled
SalesOrderReadyForFulfillment
SalesOrderClosed
```

### 18.5 Rule đặc biệt

Sales Order confirm phải tạo reservation qua Inventory.

```text
Confirm order
→ check price/discount
→ check customer/channel rule
→ Inventory.ReserveStock
→ order status = RESERVED nếu thành công
```

Nếu tồn không đủ:

```text
order status = PENDING_STOCK
```

### 18.6 Cấm

Sales không được:

```text
Trừ tồn trực tiếp
Đổi batch đã pick nếu không qua Inventory/Shipping
Set Delivered nếu Shipping chưa xác nhận
Set Paid nếu Finance chưa xác nhận
```

---

## 19. Module Shipping / Carrier Handover

### 19.1 Trách nhiệm

Shipping sở hữu:

```text
Pick task
Pack task
Package
Tote/Bin grouping for handover
Carrier manifest
Scan verification
Carrier handover confirmation
Missing/short package handling
Delivery status basic
```

### 19.2 Commands

```text
CreatePickTask
ConfirmPicked
CreatePackTask
ConfirmPacked
CreateShippingManifest
AssignPackageToTote
ScanPackageForManifest
ConfirmCarrierHandover
ReportMissingPackage
ResolveMissingPackage
UpdateDeliveryStatus
```

### 19.3 Queries

```text
ListPackagesReadyForHandover
GetManifestDetail
GetManifestScanProgress
ListMissingPackages
GetOrderShippingStatus
```

### 19.4 Events

```text
PackagePicked
PackagePacked
ManifestCreated
PackageScannedForManifest
ShipmentHandedOver
PackageMissingReported
DeliveryStatusUpdated
```

### 19.5 Manifest state

```text
DRAFT
READY_TO_SCAN
SCANNING
SHORT_PACKAGE
READY_TO_HANDOVER
HANDED_OVER
CANCELLED
```

### 19.6 Scan verification rule

Không cho handover nếu:

```text
Có package chưa scan
Có package sai carrier
Có package status không phải PACKED
Có order bị cancel/hold
Có batch bị QC hold sau khi pack
```

### 19.7 Missing package flow

Nếu scan thiếu:

```text
ReportMissingPackage
→ kiểm tra lại mã
→ tìm trong khu vực đóng hàng
→ nếu tìm thấy: scan lại và tiếp tục
→ nếu không tìm thấy: mark SHORT_PACKAGE, báo kho/sales/CSKH
```

### 19.8 Stock issue timing

Chốt một trong hai strategy trước khi build:

```text
Strategy A: Issue stock khi ConfirmPacked
Strategy B: Issue stock khi ConfirmCarrierHandover
```

Khuyến nghị Phase 1:

```text
Reserve khi ConfirmSalesOrder
Issue khi ConfirmCarrierHandover
```

Lý do: bám workflow thực tế bàn giao ĐVVC bằng scan/manifest, giảm rủi ro trừ tồn sớm nhưng đơn chưa ra khỏi kho.

---

## 20. Module Returns

### 20.1 Trách nhiệm

Returns sở hữu:

```text
Return Case
Returned package receiving
Return scan
Return inspection request
Return disposition
Return reason
Refund/credit note reference
```

### 20.2 Commands

```text
CreateReturnCase
ReceiveReturnedPackage
ScanReturnedItem
InspectReturnedItem
MarkReturnUsable
MarkReturnUnusable
SendReturnToLab
CompleteReturnCase
RejectReturnCase
```

### 20.3 Queries

```text
ListReturnCases
GetReturnCaseDetail
ListPendingReturnInspection
GetReturnByOrder
GetReturnReasonAnalytics
```

### 20.4 Events

```text
ReturnCaseCreated
ReturnedPackageReceived
ReturnedItemScanned
ReturnedItemInspected
ReturnMarkedUsable
ReturnMarkedUnusable
ReturnSentToLab
ReturnCaseCompleted
```

### 20.5 Disposition rule

Sau inspection:

```text
USABLE → Inventory.ReceiveReturnedStock vào stock type phù hợp
UNUSABLE → Inventory.MoveToDamaged/Lab/WriteOff pending approval
LAB_REQUIRED → chuyển lab/quarantine
```

Không có inspection thì không được đưa về available.

### 20.6 Link bắt buộc

Return phải link được tới:

```text
Sales Order
Shipment
SKU
Batch nếu có
Customer
Return reason
Inspection result
Disposition
```

---

## 21. Module Finance Basic

### 21.1 Trách nhiệm

Phase 1 chưa cần full accounting engine, nhưng phải có Finance Basic để quản trị tiền-hàng cơ bản.

Sở hữu:

```text
AR basic
AP basic
Payment request
Payment confirmation
COD reconciliation reference
Supplier deposit/final payment reference
Cost snapshot basic
```

### 21.2 Commands

```text
CreatePaymentRequest
ApprovePaymentRequest
RecordPayment
RecordCODReconciliation
RecordSupplierDeposit
RecordSupplierFinalPayment
CreateCustomerReceivable
MarkReceivablePaid
```

### 21.3 Queries

```text
ListReceivables
ListPayables
GetPaymentStatus
GetOrderFinancialSnapshot
GetSupplierPaymentStatus
```

### 21.4 Events

```text
PaymentRequested
PaymentApproved
PaymentRecorded
CODReconciled
SupplierDepositRecorded
SupplierFinalPaymentRecorded
```

### 21.5 Cấm

Finance không được tự sửa trạng thái vận hành:

```text
Không tự set order Delivered
Không tự set PO Received
Không tự set QC Passed
Không tự set stock Issued
```

Finance chỉ ghi nhận tiền và phát event/notification cho module nghiệp vụ nếu cần.

---

## 22. Module Reporting Basic

### 22.1 Trách nhiệm

Reporting không sở hữu giao dịch nghiệp vụ. Reporting tạo:

```text
Read model
Snapshot
KPI aggregation
Export job
```

### 22.2 Queries

```text
GetDailyWarehouseSummary
GetInventoryAgingReport
GetSalesByChannel
GetPendingQCReport
GetManifestHandoverReport
GetReturnsSummary
GetSubcontractOrderReport
```

### 22.3 Rule

Reporting được đọc nhiều bảng/read model, nhưng không được write nghiệp vụ.

Nếu báo cáo cần dữ liệu nặng:

```text
Dùng async job
Dùng snapshot table
Dùng materialized view nếu cần
```

---

## 23. Component State Machine

### 23.1 Mục tiêu

Tất cả object có lifecycle phải dùng state machine rõ.

Ví dụ Sales Order:

```text
DRAFT
CONFIRMED
PENDING_STOCK
RESERVED
PICKING
PACKED
HANDED_OVER
DELIVERED
RETURNED
CLOSED
CANCELLED
```

### 23.2 Chuẩn implement

Trong domain:

```go
func (o *SalesOrder) Confirm() error {
    if o.Status != SalesOrderStatusDraft {
        return ErrInvalidStateTransition
    }
    o.Status = SalesOrderStatusConfirmed
    return nil
}
```

Không set status bừa ở application:

```go
order.Status = "DELIVERED" // cấm
```

### 23.3 Transition phải có guard

Ví dụ Shipping Manifest:

```text
READY_TO_HANDOVER chỉ khi all packages scanned
HANDED_OVER chỉ khi carrier confirmation recorded
SHORT_PACKAGE chỉ khi missing package report tồn tại
```

---

## 24. Component Stock Ledger

### 24.1 Movement types Phase 1

```text
PURCHASE_RECEIPT_HOLD
QC_RELEASE_TO_AVAILABLE
QC_REJECT_TO_QUARANTINE
SALES_RESERVE
SALES_RESERVATION_RELEASE
SALES_ISSUE
RETURN_RECEIPT_HOLD
RETURN_RELEASE_TO_AVAILABLE
RETURN_TO_DAMAGED
TRANSFER_TO_FACTORY
FACTORY_RETURN_UNUSED_MATERIAL
SUBCONTRACT_FINISHED_GOODS_RECEIPT_HOLD
STOCK_ADJUSTMENT
CYCLE_COUNT_ADJUSTMENT
```

### 24.2 Idempotency

Mỗi movement phải có idempotency key:

```text
source_type + source_id + movement_type + line_id
```

Ví dụ:

```text
SALES_ORDER:SO-20260424-0001:SALES_RESERVE:LINE-001
```

### 24.3 Balance projection

Projection cập nhật trong cùng transaction hoặc qua outbox tùy quyết định kỹ thuật.

Khuyến nghị Phase 1:

```text
Stock movement + balance projection update trong cùng transaction cho operation chính.
Async rebuild/checksum dùng job cuối ngày.
```

---

## 25. Component QC/Batch Guard

### 25.1 Batch status guard

Sales/Shipping/Inventory phải check:

```text
Batch status
Expiry date
Stock type
Warehouse/location
Reservation status
```

### 25.2 Rule mỹ phẩm tối thiểu

```text
Batch hết hạn → không được bán
Batch QC HOLD → không được pick/pack/handover
Batch FAILED → không được available
Batch cận date → có cảnh báo hoặc rule bán riêng
```

### 25.3 Traceability

Mọi movement nên giữ:

```text
batch_no
manufacture_date
expiry_date
source_document
source_line_id
```

Đơn hàng phải trace được về batch đã xuất.

---

## 26. Component Carrier Manifest

### 26.1 Aggregate chính

```text
CarrierManifest
ManifestPackage
Tote/BinGroup
ScanRecord
HandoverConfirmation
```

### 26.2 ManifestPackage fields gợi ý

```text
manifest_package_id
manifest_id
sales_order_id
package_code
carrier_code
tote_code
status
scanned_at
scanned_by
missing_reason
resolved_at
```

### 26.3 Rule

Không xác nhận bàn giao nếu:

```text
manifest.status != READY_TO_HANDOVER
all package.status != SCANNED
carrier confirmation missing
```

---

## 27. Component Return Inspection

### 27.1 Aggregate chính

```text
ReturnCase
ReturnPackage
ReturnItem
ReturnInspection
ReturnDisposition
```

### 27.2 Disposition enum

```text
USABLE
UNUSABLE
LAB_REQUIRED
DAMAGED
MISSING
WRONG_ITEM
PENDING_DECISION
```

### 27.3 Rule

```text
Return item chưa inspect → không được nhập available
Return item lab required → chuyển LAB/QUARANTINE
Return item usable → nhập theo batch nếu batch còn hợp lệ
Return item unusable → write-off/huỷ cần approval
```

---

## 28. Component Subcontract Manufacturing

### 28.1 Aggregate chính

```text
SubcontractOrder
FactorySpecConfirmation
FactoryMaterialTransfer
FactorySample
FactoryProductionRun
FactoryDelivery
FactoryInspectionClaim
```

### 28.2 State gợi ý

```text
DRAFT
CONFIRMED
DEPOSIT_RECORDED
MATERIAL_TRANSFER_PENDING
MATERIAL_TRANSFERRED
SAMPLE_PENDING
SAMPLE_APPROVED
MASS_PRODUCTION
DELIVERED_TO_WAREHOUSE
QC_PENDING
ACCEPTED
CLAIM_RAISED
CLOSED
CANCELLED
```

### 28.3 Rule

```text
Không mass production nếu sample chưa approved
Không final payment nếu chưa nhận/QC hàng hoặc chưa có override approval
Không close subcontract order nếu NVL/bao bì consignment chưa reconcile
```

---

## 29. Sync call vs async event

### 29.1 Dùng sync call khi cần quyết định ngay

Ví dụ:

```text
Sales confirm order → Inventory check/reserve stock ngay
Shipping handover → Inventory issue stock ngay
QC pass → Inventory release stock ngay
```

### 29.2 Dùng async event khi không cần block user

Ví dụ:

```text
Gửi notification
Tạo report snapshot
Update dashboard
Export file
Sync external carrier status
```

### 29.3 Dùng outbox cho event quan trọng

Event sau transaction phải ghi vào outbox:

```text
StockReserved
ShipmentHandedOver
QCInspectionPassed
ReturnItemInspected
SubcontractGoodsReceived
```

Worker xử lý retry.

---

## 30. Chuẩn API theo module

### 30.1 URL pattern

```text
/api/v1/<module>/<resource>
```

Ví dụ:

```text
/api/v1/inventory/stock-movements
/api/v1/inventory/reservations
/api/v1/qc/inspections
/api/v1/shipping/manifests
/api/v1/returns/cases
/api/v1/production/subcontract-orders
```

### 30.2 Action endpoint

Dùng action endpoint cho nghiệp vụ chuyển trạng thái:

```text
POST /api/v1/sales/orders/{id}/confirm
POST /api/v1/inventory/reservations/{id}/release
POST /api/v1/qc/inspections/{id}/pass
POST /api/v1/shipping/manifests/{id}/confirm-handover
POST /api/v1/returns/cases/{id}/complete
POST /api/v1/production/subcontract-orders/{id}/approve-sample
```

Không dùng `PATCH status` chung chung cho state quan trọng.

### 30.3 Response phải có action availability

Để frontend hiển thị nút đúng theo state/permission, detail API nên trả:

```json
{
  "data": {},
  "available_actions": [
    "submit",
    "approve",
    "cancel"
  ],
  "permissions": {
    "can_edit": true,
    "can_approve": false
  }
}
```

---

## 31. Error code theo module

### 31.1 Format

```text
<MODULE>_<ERROR_NAME>
```

Ví dụ:

```text
INVENTORY_INSUFFICIENT_STOCK
INVENTORY_BATCH_ON_HOLD
QC_INVALID_INSPECTION_RESULT
SHIPPING_MANIFEST_NOT_FULLY_SCANNED
RETURNS_ITEM_NOT_INSPECTED
PRODUCTION_SAMPLE_NOT_APPROVED
SALES_DISCOUNT_APPROVAL_REQUIRED
```

### 31.2 Error phải phục vụ UI và vận hành

Không trả lỗi mơ hồ:

```text
Bad request
Failed
Error occurred
```

Phải trả lỗi hành động được:

```json
{
  "success": false,
  "code": "SHIPPING_MANIFEST_NOT_FULLY_SCANNED",
  "message": "Chưa thể bàn giao vì manifest còn 3 kiện chưa quét.",
  "details": {
    "manifest_no": "MN-20260424-0001",
    "missing_package_count": 3
  }
}
```

---

## 32. Read model và reporting boundary

### 32.1 Khi nào cần read model

Cần read model nếu màn hình cần dữ liệu tổng hợp từ nhiều module:

```text
Dashboard CEO
Warehouse daily summary
Order fulfillment status
Manifest scan progress
Batch trace
Stock aging
```

### 32.2 Rule

Read model chỉ phục vụ đọc. Không dùng read model làm source of truth.

Ví dụ:

```text
order_fulfillment_read_model
inventory_daily_summary
manifest_scan_summary
batch_trace_read_model
```

### 32.3 Rebuild

Read model phải có khả năng rebuild từ source event/table.

---

## 33. Transaction boundary

### 33.1 Một use case = một transaction chính

Ví dụ ConfirmCarrierHandover:

```text
Validate manifest
Validate all packages scanned
Update manifest status
Create stock issue movement
Update package/order shipping status
Create audit log
Write outbox event
Commit
```

### 33.2 Không gọi external API trong DB transaction

Sai:

```text
Begin transaction
Update order
Call external carrier API
Update shipment
Commit
```

Đúng:

```text
Begin transaction
Update local state
Write outbox job
Commit
Worker calls external carrier API
```

### 33.3 Cross-module transaction

Nếu transaction chạm nhiều module, application service của module điều phối phải gọi contract rõ ràng. Không tự update bảng chéo.

---

## 34. Composition root

Tất cả module được wire ở composition root:

```text
cmd/api/main.go
internal/bootstrap/
```

Ví dụ:

```go
inventoryModule := inventory.NewModule(deps)
salesModule := sales.NewModule(deps, inventoryModule.Service())
shippingModule := shipping.NewModule(deps, inventoryModule.Service(), salesModule.QueryService())
```

Không import vòng tròn giữa module.

Nếu có circular dependency, chứng tỏ boundary đang sai.

---

## 35. Module configuration

Mỗi module có thể có config riêng:

```text
inventory.allow_negative_stock = false
inventory.issue_stock_at = handover
qc.require_release_for_sales = true
shipping.require_scan_before_handover = true
returns.require_inspection_before_available = true
production.subcontract.claim_window_days = 7
```

Không hardcode rule có khả năng thay đổi theo nghiệp vụ.

---

## 36. Testing chuẩn theo module/component

### 36.1 Unit test domain

Bắt buộc test state machine và policy:

```text
Batch hold không cho release stock
Manifest chưa scan đủ không cho handover
Return chưa inspect không cho nhập available
Subcontract sample chưa approve không cho mass production
```

### 36.2 Application test

Test use case có transaction:

```text
ConfirmSalesOrder reserves stock
QC pass releases stock
ConfirmCarrierHandover issues stock
InspectReturn usable creates return receipt
```

### 36.3 Contract test

Test module contract:

```text
Sales gọi Inventory.ReserveStock đúng behavior
Shipping gọi Inventory.IssueStock đúng behavior
Returns gọi QC/Inventory đúng behavior
```

### 36.4 Integration test DB

Test repository với PostgreSQL thật hoặc test container.

Đặc biệt:

```text
Stock movement idempotency
Balance projection
Concurrent reserve stock
Concurrent manifest scan
```

---

## 37. Anti-patterns phải tránh

### 37.1 God service

Sai:

```text
OperationService xử lý purchase, sales, stock, shipping, returns
```

Đúng:

```text
Mỗi module có application service riêng
Cross-module orchestration có contract rõ
```

### 37.2 Shared repository

Sai:

```text
common/repository/order_repository.go
common/repository/stock_repository.go
```

Đúng:

```text
sales/repository
inventory/repository
```

### 37.3 Update status tự do

Sai:

```go
UpdateStatus(id, status string)
```

Đúng:

```go
ConfirmOrder(id)
CancelOrder(id, reason)
ConfirmHandover(id)
```

### 37.4 Boolean nghiệp vụ

Sai:

```text
is_qc_passed
is_shipped
is_returned
```

Đúng:

```text
qc_status
shipping_status
return_status
```

### 37.5 Cross-module SQL write

Sai:

```sql
-- trong sales module
UPDATE inventory_balances SET reserved_qty = reserved_qty + $1;
```

Đúng:

```go
inventoryService.ReserveStock(ctx, cmd)
```

---

## 38. Checklist khi thiết kế một module mới

Trước khi tạo module mới, phải trả lời:

```text
1. Module này đại diện cho business capability nào?
2. Module này sở hữu entity/table nào?
3. Module này expose command/query/event gì?
4. Module này cần gọi module nào?
5. Module nào được gọi module này?
6. Có state machine không?
7. Có approval không?
8. Có audit log không?
9. Có stock/cost/batch/QC impact không?
10. Có cần outbox/event không?
11. Có read model/reporting không?
12. UAT case chính là gì?
```

Nếu không trả lời được 12 câu này, chưa nên code.

---

## 39. Checklist khi thiết kế component dùng chung

Component dùng chung phải có:

```text
1. Mục tiêu rõ
2. Không chứa rule nghiệp vụ riêng của module
3. API contract rõ
4. Error code rõ
5. Logging/audit nếu cần
6. Idempotency nếu có side effect
7. Test riêng
8. Không tạo dependency vòng tròn
```

Ví dụ approval component chỉ biết:

```text
approval request
approval step
approver
decision
```

Nó không nên biết chi tiết PO line, stock movement hay discount formula.

---

## 40. Definition of Done cho module/component

Một module/component được coi là đạt chuẩn khi có:

```text
Module boundary document ngắn
Entity/table ownership rõ
Command/query/event list
State machine nếu có
API endpoint + OpenAPI
Permission mapping
Audit log mapping
Error code mapping
Transaction rule
Unit test domain
Application test
Repository integration test cho flow quan trọng
UAT scenario liên quan pass
Không cross-module SQL write
```

---

## 41. Thứ tự ưu tiên thiết kế chi tiết Phase 1

Ưu tiên theo mức độ rủi ro:

```text
1. Inventory / Stock Ledger / Warehouse Shift
2. QC / Batch Guard
3. Sales Order / Reservation
4. Shipping / Carrier Manifest / Scan Handover
5. Returns / Return Inspection / Disposition
6. Production / Subcontract Manufacturing
7. Purchase / Receiving
8. Finance Basic
9. Reporting Basic
10. Master Data polish
```

Lý do: ERP của công ty mỹ phẩm chết nhiều nhất ở hàng, batch, giao hàng, hàng hoàn, gia công ngoài và đối soát.

---

## 42. Gợi ý thiết kế sprint technical spike

Trước khi code full, nên làm spike 5 flow nguy hiểm:

### Spike 1: Sales reserve stock

```text
Create order → confirm → reserve stock → audit → event
```

### Spike 2: QC pass releases inbound stock

```text
Goods receipt hold → QC pass → release stock → available
```

### Spike 3: Shipping scan manifest and handover

```text
Packed packages → manifest → scan all → handover → issue stock
```

### Spike 4: Return inspection

```text
Return received → scan → inspect → usable/unusable → stock disposition
```

### Spike 5: Subcontract material transfer and finished goods receipt

```text
Subcontract order → transfer material to factory → approve sample → receive finished goods → QC
```

Nếu 5 spike này chạy sạch, backend foundation đủ chắc để mở rộng.

---

## 43. Những quyết định cần business sign-off

Các quyết định sau phải được chủ doanh nghiệp/COO/kho/finance chốt trước khi build sâu:

```text
1. Trừ tồn khi packed hay khi handover ĐVVC?
2. Hàng hoàn usable có nhập available ngay hay qua QC?
3. Hàng chuyển sang nhà máy gia công có tính là tồn của công ty không? Khuyến nghị: có, ở location FACTORY_CONSIGNMENT.
4. Batch cận date có được bán không? Nếu có, theo rule nào?
5. Stock adjustment dưới ngưỡng nào cần duyệt?
6. Đổi batch sau khi đã pick có cho phép không?
7. Close ca kho khi còn manifest thiếu đơn có được không?
8. Supplier/factory claim window mặc định là 3, 5 hay 7 ngày?
```

Không chốt các câu này thì code có thể đúng kỹ thuật nhưng sai vận hành.

---

## 44. Liên kết với tài liệu khác

Tài liệu này cần được đọc cùng:

```text
03_ERP_PRD_SRS_Phase1_My_Pham_v1.md
04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md
05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md
06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md
08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md
11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md
12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md
```

Sau khi phân tích thêm file Excel thực tế hoặc workflow mới, tài liệu này có thể cần bản v1.1 để cập nhật module boundary.

---

## 45. Kết luận

Chuẩn module/component này đặt ra một luật rất đơn giản:

> Module nào sở hữu nghiệp vụ thì module đó sở hữu dữ liệu, state, rule và event của nghiệp vụ đó.

Với ERP mỹ phẩm Phase 1, các điểm phải giữ chắc nhất là:

```text
Stock ledger bất biến
Batch/QC guard
Manifest scan trước bàn giao ĐVVC
Return inspection trước khi nhập lại hàng
Subcontract manufacturing có transfer NVL/bao bì và sample approval
Audit log đầy đủ
Không cross-module write
```

Nếu giữ được các luật này, backend Go sẽ không chỉ chạy được, mà còn đủ sạch để mở rộng sang HRM, KOL/Affiliate, CRM, BI nâng cao và multi-brand/multi-warehouse ở các phase sau.
