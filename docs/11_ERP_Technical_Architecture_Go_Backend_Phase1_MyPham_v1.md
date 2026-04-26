# 11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1

**Project:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** Technical Architecture + Backend Stack Decision  
**Scope:** Phase 1  
**Version:** v1.0  
**Date:** 2026-04-24  
**Language:** Vietnamese  
**Primary Decision:** Backend sử dụng **Go**  
**Owner:** ERP Solution Architect / Technical Lead  

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
- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt **kiến trúc kỹ thuật Phase 1** cho hệ thống ERP Web, với backend sử dụng **Go**.

Tài liệu này không thay thế PRD/SRS hay Process Flow. Nó dùng để trả lời các câu hỏi kỹ thuật mà đội dev, tech lead, QA automation, DevOps và vendor cần biết trước khi build:

1. Hệ thống dùng tech stack gì.
2. Backend Go tổ chức module ra sao.
3. API, database, transaction, audit log, queue, file storage thiết kế thế nào.
4. Module nào được gọi module nào, module nào không được chọc dữ liệu của module khác.
5. Tồn kho, batch, QC, đơn hàng, bàn giao vận chuyển, hàng hoàn, gia công sản xuất được xử lý ở tầng kỹ thuật ra sao.
6. Môi trường dev/staging/prod, CI/CD, observability, backup/restore, security baseline cần có gì.

Nói ngắn gọn:

- **PRD/SRS** nói hệ thống phải làm gì.
- **Process Flow** nói công ty vận hành thế nào.
- **Technical Architecture** nói đội kỹ thuật phải xây nó ra sao để không vỡ.

---

## 2. Kết luận kiến trúc đã chốt

### 2.1. Quyết định chính

| Hạng mục | Quyết định |
|---|---|
| Frontend | TypeScript + React / Next.js |
| Backend | Go |
| Architecture | Modular Monolith |
| API | REST + OpenAPI |
| Database | PostgreSQL |
| Cache | Redis |
| Queue / Background Job | RabbitMQ hoặc NATS JetStream, ưu tiên RabbitMQ nếu team vận hành phổ thông hơn |
| File Storage | S3-compatible storage / MinIO |
| Auth | JWT/session hybrid + RBAC + approval rules |
| Deployment | Docker, containerized services |
| Observability | Structured logging + metrics + tracing cơ bản |
| Reporting | Query/report layer + async export job |

### 2.2. Tinh thần thiết kế

Không làm microservices ngay từ đầu.

Phase 1 nên đi theo hướng:

```text
Modular Monolith
= một backend Go chính
+ chia module rõ
+ database chung nhưng boundary nghiêm
+ API contract rõ
+ event/outbox cho tác vụ bất đồng bộ
```

Lý do:

1. ERP Phase 1 còn đang cần khóa nghiệp vụ thật.
2. Dữ liệu hàng - tiền - batch - QC - đơn hàng liên quan chặt với nhau.
3. Microservices quá sớm sẽ làm tăng chi phí tích hợp, DevOps, monitoring, deployment và debug.
4. Modular Monolith cho phép build nhanh, kiểm soát transaction tốt, nhưng vẫn chuẩn bị được đường tách service sau này.

Câu chốt:

> Phase 1 cần backend chắc, rõ, ít lỗi, transaction đúng. Không cần khoe kiến trúc phức tạp.

---

## 3. Business reality anchor từ workflow thực tế

Sau khi xem các tài liệu vận hành thực tế, architecture phải phản ánh những điểm sau:

### 3.1. Kho có nhịp đóng ca và đối soát cuối ngày

Tài liệu `Công-việc-hằng-ngày.pdf` thể hiện chuỗi công việc kho: tiếp nhận đơn trong ngày, thực hiện xuất/nhập dựa trên bảng nội quy, soạn và đóng gói, sắp xếp/tối ưu kho, kiểm kê tồn kho cuối ngày, đối soát số liệu và báo cáo quản lý, sau đó kết thúc ca.

Hệ thống cần hỗ trợ:

- Shift / ca làm.
- End-of-day stock reconciliation.
- Cycle count / kiểm kê cuối ngày.
- Báo cáo lệch giữa tồn hệ thống và tồn thực tế.
- Audit log cho các điều chỉnh cuối ca.

### 3.2. Nội quy kho có 4 nhánh nghiệp vụ lớn

Tài liệu `Nội-Quy.pdf` thể hiện các nhánh:

1. Nhập kho.
2. Xuất kho.
3. Đóng hàng.
4. Xử lý hàng hoàn.

Hệ thống cần hỗ trợ:

- Receiving / nhập kho có kiểm số lượng, bao bì, lô.
- Outbound / xuất kho có đối chiếu thực tế.
- Packing / đóng hàng theo đơn, theo ĐVVC, theo sàn/kênh.
- Return receiving / nhận hàng hoàn.
- Return inspection / kiểm hàng hoàn.
- Return disposition / phân loại còn dùng được, không dùng được, chuyển kho, chuyển lab.

### 3.3. Bàn giao ĐVVC cần scan verify và manifest

Tài liệu `Quy-trình-bàn-giao.pdf` thể hiện quy trình chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn theo bảng, lấy hàng và quét mã trực tiếp tại hàm, sau đó xác nhận bàn giao nếu đủ; nếu chưa đủ thì kiểm tra lại mã hoặc tìm lại trong khu vực đóng hàng.

Hệ thống cần hỗ trợ:

- Carrier manifest / bảng kê bàn giao theo ĐVVC.
- Scan verify trước bàn giao.
- Đối chiếu số lượng đơn theo rổ/thùng/chuyến.
- Trạng thái thiếu đơn / chưa đủ.
- Xử lý tìm lại đơn trong khu vực đóng hàng.
- Ký xác nhận hoặc xác nhận điện tử bàn giao.

### 3.4. Sản xuất có nhánh gia công ngoài

Tài liệu `Quy-trình-sản-xuất.pdf` thể hiện mô hình lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển nguyên vật liệu/bao bì qua nhà máy, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao hàng về kho, kiểm tra số lượng/chất lượng, nhận hàng hoặc báo lỗi trong 3-7 ngày, rồi thanh toán lần cuối.

Hệ thống cần hỗ trợ:

- Subcontract manufacturing / gia công ngoài.
- Chuyển NVL/bao bì sang nhà máy.
- Biên bản bàn giao NVL/bao bì.
- Sample approval / duyệt mẫu.
- Deposit payment / cọc đơn.
- Final payment / thanh toán cuối.
- Receiving finished goods from factory.
- QC inspection khi nhận hàng gia công.
- Reject/claim lại nhà máy trong thời hạn quy định.

---

## 4. Kiến trúc tổng thể

### 4.1. High-level system diagram

```text
[User Browser / Tablet / Scanner]
          |
          v
[Frontend: React / Next.js]
          |
          | REST API + OpenAPI Contract
          v
[Backend API: Go Modular Monolith]
          |
          |-- Auth / RBAC / Approval
          |-- Master Data
          |-- Procurement
          |-- Inventory / WMS
          |-- QA/QC
          |-- Production / Subcontract
          |-- Sales / OMS
          |-- Shipping / Handover
          |-- Returns
          |-- Finance Lite
          |-- Reporting
          |
          | Transactional Read/Write
          v
[PostgreSQL]
          |
          | Async events / jobs
          v
[Queue: RabbitMQ/NATS] ---> [Worker Service: Go]
          |
          | Cache / Session / Lock
          v
[Redis]
          |
          | Files / Attachments / Proof
          v
[S3 / MinIO]
```

### 4.2. Runtime components

Phase 1 nên có các runtime component sau:

```text
1. web-frontend
   - React / Next.js app
   - dùng cho browser, tablet, màn hình kho

2. api-backend
   - Go backend chính
   - xử lý REST API, transaction, business rules

3. worker
   - Go worker
   - xử lý async jobs: report export, notification, sync, scheduled tasks

4. postgres
   - database chính

5. redis
   - cache, short-lived locks, session/token blacklist nếu cần

6. queue
   - RabbitMQ hoặc NATS
   - xử lý event/job bất đồng bộ

7. object-storage
   - MinIO/S3
   - lưu file đính kèm, ảnh QC, chứng từ, biên bản, proof bàn giao
```

### 4.3. Không đưa business logic vào frontend

Frontend chỉ nên xử lý:

- hiển thị;
- validate nhẹ để tiện người dùng;
- gọi API;
- scan barcode/QR và gửi mã lên backend;
- hiển thị trạng thái, lỗi, cảnh báo.

Business logic quan trọng phải nằm ở backend:

- kiểm tồn khả dụng;
- giữ tồn;
- xuất kho;
- QC pass/hold/fail;
- đổi trạng thái đơn hàng;
- tính trạng thái hàng hoàn;
- phê duyệt;
- audit log;
- phân quyền;
- tính công nợ cơ bản;
- batch trace.

Không được để frontend quyết định các việc này, vì frontend có thể bị bypass.

---

## 5. Tech stack chi tiết

### 5.1. Backend Go

| Layer | Đề xuất |
|---|---|
| HTTP router | `chi` hoặc `gin`, ưu tiên `chi` nếu muốn nhẹ và rõ |
| DB driver | `pgx` |
| SQL access | `sqlc` |
| Migration | `golang-migrate` |
| Validation | `go-playground/validator` hoặc validation thủ công theo domain |
| Logging | `zap` hoặc `zerolog` |
| Testing | `go test`, `testify`, `testcontainers-go` |
| OpenAPI | `oapi-codegen` hoặc maintain `openapi.yaml` rồi generate types |
| Queue client | RabbitMQ client hoặc NATS client |
| Config | env-based config + config struct |
| Auth | JWT/session hybrid, password hashing an toàn, RBAC middleware |

Khuyến nghị mặc định:

```text
Router: chi
Database: pgx + sqlc
Migration: golang-migrate
Logging: zap
Testing: testify + testcontainers-go
API contract: OpenAPI
```

### 5.2. Frontend

| Layer | Đề xuất |
|---|---|
| Framework | React / Next.js |
| Language | TypeScript |
| Form | React Hook Form hoặc tương đương |
| Data fetching | TanStack Query hoặc tương đương |
| Table | Data table component có sorting/filter/pagination server-side |
| State | Ưu tiên server state, hạn chế global state phức tạp |
| API typing | Generate từ OpenAPI nếu được |

### 5.3. Database

```text
PostgreSQL
```

Lý do:

- dữ liệu ERP có quan hệ chặt;
- cần transaction mạnh;
- cần query báo cáo;
- cần constraint;
- cần locking/row version cho stock và document;
- dễ vận hành, backup, restore.

### 5.4. Redis

Redis dùng cho:

- cache dữ liệu ít thay đổi;
- distributed lock ngắn hạn nếu cần;
- rate limit;
- session/token blacklist nếu áp dụng;
- cache dashboard tạm thời;
- job status ngắn hạn.

Không dùng Redis làm nguồn sự thật cho tồn kho, đơn hàng, QC, công nợ.

### 5.5. Queue / background jobs

Queue dùng cho:

- export report;
- gửi notification;
- đồng bộ trạng thái vận chuyển;
- xử lý webhook từ bên ngoài;
- cảnh báo cận date;
- cảnh báo tồn thấp;
- tính snapshot KPI cuối ngày;
- chạy reconciliation job.

Không dùng queue để thay thế transaction chính. Ví dụ: xuất kho, giữ tồn, QC release phải commit trong PostgreSQL trước.

### 5.6. Object storage

S3/MinIO dùng lưu:

- COA/MSDS;
- chứng từ giao nhận;
- ảnh QC;
- ảnh hàng hoàn;
- biên bản bàn giao;
- file import/export;
- bằng chứng vận chuyển;
- hợp đồng nhà cung cấp/gia công.

DB chỉ lưu metadata:

```text
file_id
module
entity_type
entity_id
file_name
mime_type
size
storage_key
uploaded_by
uploaded_at
checksum
```

---

## 6. Kiến trúc backend Go

### 6.1. Cấu trúc repository đề xuất

```text
erp-backend/
  cmd/
    api/
      main.go
    worker/
      main.go

  internal/
    shared/
      auth/
      config/
      database/
      errors/
      logger/
      middleware/
      pagination/
      response/
      transaction/
      validation/
      files/
      events/
      audit/
      workflow/
      rbac/

    modules/
      masterdata/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      procurement/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      inventory/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      qc/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      production/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      sales/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      shipping/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      returns/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      finance/
        handler/
        application/
        domain/
        repository/
        dto/
        queries/
        events/

      reporting/
        handler/
        application/
        repository/
        dto/

  migrations/
  api/
    openapi.yaml
  sql/
    queries/
      inventory.sql
      sales.sql
      procurement.sql

  deployments/
    docker-compose.yml
    k8s-or-server-config/

  docs/
```

### 6.2. Layer rule

Mỗi module nên có 4 lớp chính:

```text
handler
  nhận HTTP request, parse input, gọi application service

application
  orchestration use case, transaction boundary, gọi domain/repository/module contract

domain
  business rules, state machine, validation sâu

repository
  đọc/ghi database, không chứa business rule lớn
```

Luồng chuẩn:

```text
HTTP Request
→ Handler
→ Application Service
→ Domain Logic
→ Repository
→ PostgreSQL
→ Audit/Event
→ HTTP Response
```

### 6.3. Rule phụ thuộc

```text
handler → application → domain
application → repository
application → shared services
repository → database
```

Không cho:

```text
repository gọi handler
repository gọi module khác
handler thao tác database trực tiếp
domain phụ thuộc HTTP request
frontend quyết định rule nghiệp vụ quan trọng
```

### 6.4. Module contract

Module khác không được tự ý update bảng của module khác.

Ví dụ:

- `sales` không được tự update tồn kho.
- `shipping` không được tự đổi stock ledger.
- `qc` không được tự ghi doanh thu.
- `production` không được tự sửa giá bán.

Module phải giao tiếp qua application service contract hoặc event.

Ví dụ:

```text
sales → inventory.ReserveStock(orderID, lines)
sales → inventory.ReleaseReservation(orderID)
shipping → inventory.ConfirmSalesIssue(shipmentID)
qc → inventory.ReleaseBatch(batchID)
returns → inventory.ReceiveReturnedGoods(returnID)
production → inventory.IssueMaterialsToFactory(transferID)
production → inventory.ReceiveFinishedGoods(workOrderID)
```

Câu chốt:

> Database có thể chung, nhưng quyền viết dữ liệu phải thuộc đúng module sở hữu.

---

## 7. Module boundary Phase 1

### 7.1. System / Common

Sở hữu:

- user;
- role;
- permission;
- session/token;
- approval workflow;
- audit log;
- notification;
- sequence/document number;
- file attachment;
- configuration.

Không sở hữu:

- stock;
- sales order;
- purchase order;
- QC result;
- production order.

### 7.2. Master Data

Sở hữu dữ liệu gốc:

- SKU/thành phẩm;
- nguyên vật liệu;
- bao bì;
- đơn vị tính;
- warehouse;
- warehouse zone/bin;
- supplier;
- customer;
- carrier;
- product category;
- QC checklist template;
- reason code;
- channel;
- price list cơ bản.

Rule:

- Master data tạo/sửa phải audit.
- Một số master data cần approval trước khi active.
- Không cho xóa cứng nếu đã phát sinh giao dịch.

### 7.3. Procurement

Sở hữu:

- purchase request;
- RFQ nếu Phase 1 làm nhẹ;
- purchase order;
- supplier delivery schedule;
- supplier invoice placeholder;
- supplier score basic.

Giao tiếp:

- gọi Inventory khi nhận hàng;
- gọi QC khi cần kiểm hàng đầu vào;
- gọi Finance Lite khi phát sinh payable/deposit.

### 7.4. Inventory / WMS

Sở hữu:

- stock ledger;
- stock balance;
- stock reservation;
- batch/lot;
- warehouse movement;
- transfer order;
- stock count;
- adjustment;
- location/bin;
- inventory status.

Đây là module cực quan trọng. Không module nào được tự ý update tồn.

### 7.5. QA/QC

Sở hữu:

- inspection request;
- inspection result;
- QC checklist;
- QC status;
- hold/pass/fail;
- release decision;
- reject reason;
- evidence attachment.

Giao tiếp:

- nhận inspection request từ Procurement/Inventory/Production/Returns;
- phát quyết định QC cho Inventory để thay đổi stock status.

### 7.6. Production / Subcontract Manufacturing

Sở hữu:

- production order;
- subcontract order;
- factory order;
- BOM reference;
- material issue request;
- material transfer to factory;
- sample approval;
- production receipt;
- production loss/scrap;
- factory claim/reject note.

Với workflow hiện tại, Phase 1 phải hỗ trợ gia công ngoài ở mức thực dụng, không chỉ xưởng nội bộ.

### 7.7. Sales / OMS

Sở hữu:

- quotation nếu cần;
- sales order;
- order line;
- discount rule basic;
- order status;
- reservation request;
- customer AR placeholder;
- return request trigger.

Giao tiếp:

- gọi Inventory để reserve/release/issue stock;
- gọi Shipping để tạo shipment;
- gọi Finance Lite để ghi nhận phải thu/thanh toán cơ bản.

### 7.8. Shipping / Handover

Sở hữu:

- packing task;
- shipment;
- carrier manifest;
- handover record;
- tracking code;
- delivery status;
- missing package case;
- handover proof.

Đây là module phát sinh từ workflow bàn giao ĐVVC thực tế.

### 7.9. Returns

Sở hữu:

- return receipt;
- return inspection;
- return disposition;
- usable/unusable status;
- transfer to warehouse/lab;
- linked sales order/shipment.

Return không chỉ là “hoàn đơn”. Nó là một nghiệp vụ kho riêng.

### 7.10. Finance Lite

Phase 1 không làm kế toán đầy đủ nhưng cần:

- payment record;
- AR basic;
- AP basic;
- COD reconciliation placeholder;
- deposit/final payment cho gia công;
- expense/payment approval reference;
- export data cho kế toán.

---

## 8. Database architecture

### 8.1. Nguyên tắc chung

1. PostgreSQL là nguồn sự thật chính.
2. Không ghi tồn kho bằng cách update số lượng tùy tiện.
3. Tất cả thay đổi tồn phải đi qua stock movement/stock ledger.
4. Critical transaction phải dùng DB transaction.
5. Critical document phải có trạng thái và audit log.
6. Không xóa cứng giao dịch phát sinh.
7. Dữ liệu master có thể soft delete/inactive.
8. Dữ liệu nghiệp vụ phải giữ lịch sử.

### 8.2. Schema strategy

Khuyến nghị dùng schema theo module để giữ boundary rõ:

```text
system.*
masterdata.*
procurement.*
inventory.*
qc.*
production.*
sales.*
shipping.*
returns.*
finance.*
reporting.*
```

Ví dụ:

```text
inventory.stock_movements
inventory.stock_balances
inventory.batches
sales.sales_orders
sales.sales_order_lines
shipping.shipments
shipping.carrier_manifests
qc.inspection_requests
qc.inspection_results
```

### 8.3. ID strategy

Mỗi record nên có 2 loại mã:

```text
id: UUID/ULID nội bộ, dùng cho database/API
code/doc_no: mã nghiệp vụ dễ đọc, dùng cho người dùng
```

Ví dụ:

```text
id = 01HV... hoặc UUID
po_no = PO-20260424-0001
so_no = SO-20260424-0001
batch_no = BATCH-SVC-20260424-01
manifest_no = MNF-GHN-20260424-01
```

Không dùng `doc_no` làm primary key.

### 8.4. Common columns

Các bảng nghiệp vụ nên có:

```text
id
created_at
created_by
updated_at
updated_by
deleted_at nullable nếu soft delete
status
version
```

`version` dùng cho optimistic locking ở các document dễ bị nhiều người sửa.

### 8.5. Document status table hay enum?

Khuyến nghị:

- Status nghiệp vụ có thể dùng enum hoặc varchar có constraint.
- Trạng thái cần cấu hình nhiều thì dùng table.
- Phase 1 có thể dùng enum/varchar constraint để nhanh và rõ.

Ví dụ:

```text
sales_order_status:
DRAFT, CONFIRMED, RESERVED, PICKED, PACKED, HANDED_OVER, DELIVERED, CLOSED, CANCELLED, RETURNED
```

### 8.6. Không dùng báo cáo làm nguồn sự thật

Các dashboard/report có thể dùng materialized view/snapshot/cache, nhưng nguồn sự thật vẫn là transaction table:

- stock ledger;
- sales order;
- purchase order;
- QC result;
- shipment;
- return receipt;
- payment record.

---

## 9. Stock ledger architecture

Đây là phần sống còn của ERP.

### 9.1. Nguyên tắc bất biến

Không sửa trực tiếp tồn kho bằng tay.

Mọi thay đổi tồn kho phải tạo stock movement:

```text
INBOUND_RECEIPT
QC_HOLD
QC_RELEASE
QC_REJECT
PURCHASE_RETURN
PRODUCTION_ISSUE
SUBCONTRACT_TRANSFER_OUT
SUBCONTRACT_RETURN_IN
PRODUCTION_RECEIPT
SALES_RESERVE
SALES_RESERVATION_RELEASE
SALES_ISSUE
RETURN_RECEIPT
RETURN_TO_AVAILABLE
RETURN_TO_QUARANTINE
TRANSFER_OUT
TRANSFER_IN
ADJUSTMENT_IN
ADJUSTMENT_OUT
SCRAP
SAMPLE_ISSUE
```

### 9.2. Stock balance là bảng tổng hợp có kiểm soát

Có thể duy trì bảng `stock_balances` để query nhanh:

```text
warehouse_id
location_id
sku_id / material_id
batch_id
stock_status
physical_qty
reserved_qty
available_qty
quarantine_qty
updated_at
```

Nhưng `stock_balances` phải được cập nhật **cùng transaction** với `stock_movements`.

Không cho user sửa trực tiếp `stock_balances`.

### 9.3. Công thức tồn khả dụng

```text
available_stock
= physical_stock
- reserved_stock
- qc_hold_stock
- quarantine_stock
- damaged_stock
- allocated_to_production
```

Tùy thiết kế bảng, `available_stock` có thể lưu dạng calculated snapshot, nhưng logic phải thống nhất.

### 9.4. Batch-level stock

Với mỹ phẩm, tồn kho phải đi theo batch/lô và hạn dùng.

Bắt buộc lưu:

```text
batch_no
manufacturing_date
expiry_date
qc_status
supplier_batch_no nếu có
internal_batch_no
source_type: purchase / production / return / subcontract
```

Không cho hàng chưa QC pass vào available stock.

### 9.5. Reservation strategy

Khi sales order xác nhận:

```text
Sales Order Confirmed
→ Inventory Reserve Stock
→ tạo reservation
→ tăng reserved_qty
→ order status = RESERVED nếu đủ
```

Nếu không đủ hàng:

```text
order status = CONFIRMED / PARTIALLY_RESERVED
show shortage
không cho pack/ship phần chưa reserve
```

### 9.6. FEFO/FIFO

Default cho mỹ phẩm:

```text
FEFO: First Expired, First Out
```

Nếu cùng hạn dùng:

```text
FIFO: batch nhập trước xuất trước
```

Backend phải có allocation service:

```text
AllocateStock(skuID, warehouseID, qty, rule=FEFO)
```

Không để frontend tự chọn batch trừ khi role được phép override.

### 9.7. Stock adjustment

Điều chỉnh tồn chỉ được qua chứng từ:

```text
Stock Adjustment Request
→ approval nếu vượt ngưỡng
→ create stock movement
→ update balance
→ audit log
```

Không có nút “sửa tồn trực tiếp”.

---

## 10. Transaction strategy

### 10.1. Unit of Work

Backend cần shared transaction wrapper:

```go
func (u *UseCase) Execute(ctx context.Context, input Input) error {
    return u.txManager.WithTx(ctx, func(ctx context.Context, tx DBTX) error {
        // repository calls using tx
        // business rules
        // audit log
        // outbox event
        return nil
    })
}
```

Repository nhận interface `DBTX` để dùng được cả connection thường và transaction.

### 10.2. Transaction boundary nằm ở application service

Không để handler mở transaction.
Không để repository tự mở transaction tùy tiện.

Application service là nơi orchestration:

```text
Validate input
→ Check permission/approval
→ Load data
→ Domain decision
→ Write records
→ Audit log
→ Outbox event
→ Commit
```

### 10.3. Những use case bắt buộc transaction

1. Nhập kho.
2. QC release/reject.
3. Giữ tồn đơn hàng.
4. Xuất kho bán hàng.
5. Bàn giao ĐVVC.
6. Nhận hàng hoàn.
7. Điều chỉnh tồn.
8. Cấp NVL cho sản xuất/gia công.
9. Nhận thành phẩm từ nhà máy.
10. Thanh toán/cọc/final payment.

### 10.4. Concurrency control

Các nghiệp vụ tồn kho cần chống double-issue.

Dùng kết hợp:

- row-level lock khi update stock balance;
- optimistic lock bằng `version` trên document;
- idempotency key cho callback/webhook/import;
- unique constraint cho event/external ref.

Ví dụ:

```sql
SELECT * FROM inventory.stock_balances
WHERE sku_id = $1 AND batch_id = $2 AND warehouse_id = $3
FOR UPDATE;
```

---

## 11. Outbox/event architecture

### 11.1. Vì sao cần outbox

Khi một giao dịch chính commit xong, hệ thống thường cần làm việc phụ:

- gửi notification;
- tạo job export;
- sync vận chuyển;
- tính KPI;
- ghi event reporting;
- đẩy webhook.

Không nên làm tất cả trong request chính vì sẽ chậm và dễ fail.

Nhưng cũng không nên publish queue trước khi database commit.

Dùng outbox pattern:

```text
Trong cùng DB transaction:
1. update nghiệp vụ
2. ghi audit log
3. ghi outbox event
4. commit

Worker:
5. đọc outbox event
6. publish queue hoặc xử lý job
7. mark processed
```

### 11.2. Event naming

Event nên đặt theo quá khứ:

```text
InventoryStockReserved
InventoryStockIssued
QCInspectionPassed
QCInspectionFailed
SalesOrderConfirmed
ShipmentPacked
ShipmentHandedOver
ReturnReceived
ReturnInspected
SubcontractMaterialTransferred
SubcontractFinishedGoodsReceived
```

### 11.3. Không dùng event để né transaction

Ví dụ sai:

```text
Sales confirmed → publish event → inventory tự reserve sau
```

Nếu reserve là điều kiện sống còn của order, phải reserve trong transaction chính hoặc có trạng thái pending rõ ràng.

Event chỉ nên dùng cho việc phụ, hoặc việc không cần commit cùng lúc.

---

## 12. API architecture

### 12.1. REST + OpenAPI

Phase 1 dùng REST là đủ.

OpenAPI dùng để:

- định nghĩa contract;
- generate client type cho frontend nếu có;
- generate server stub/type nếu phù hợp;
- làm tài liệu API;
- hỗ trợ testing.

### 12.2. API URL convention

```text
/api/v1/master-data/skus
/api/v1/master-data/materials
/api/v1/procurement/purchase-orders
/api/v1/inventory/stock-movements
/api/v1/inventory/stock-balances
/api/v1/qc/inspection-requests
/api/v1/production/work-orders
/api/v1/production/subcontract-orders
/api/v1/sales/orders
/api/v1/shipping/shipments
/api/v1/shipping/manifests
/api/v1/returns/return-receipts
/api/v1/finance/payments
```

Action endpoint dùng khi hành động không phải CRUD đơn giản:

```text
POST /api/v1/sales/orders/{id}/confirm
POST /api/v1/sales/orders/{id}/reserve-stock
POST /api/v1/shipping/shipments/{id}/pack
POST /api/v1/shipping/manifests/{id}/scan-verify
POST /api/v1/shipping/manifests/{id}/handover
POST /api/v1/qc/inspection-requests/{id}/pass
POST /api/v1/qc/inspection-requests/{id}/fail
POST /api/v1/returns/return-receipts/{id}/inspect
POST /api/v1/inventory/stock-counts/{id}/close
```

### 12.3. Response format

Success:

```json
{
  "success": true,
  "data": {},
  "meta": {
    "request_id": "req_..."
  }
}
```

List response:

```json
{
  "success": true,
  "data": [],
  "pagination": {
    "page": 1,
    "page_size": 50,
    "total": 120
  },
  "meta": {
    "request_id": "req_..."
  }
}
```

Error:

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
  "meta": {
    "request_id": "req_..."
  }
}
```

### 12.4. Error code convention

Error code nên là uppercase snake case:

```text
VALIDATION_ERROR
PERMISSION_DENIED
APPROVAL_REQUIRED
DOCUMENT_NOT_FOUND
DOCUMENT_STATUS_INVALID
INSUFFICIENT_STOCK
BATCH_NOT_RELEASED
QC_STATUS_INVALID
DUPLICATE_DOCUMENT_NO
IDEMPOTENCY_CONFLICT
VERSION_CONFLICT
```

### 12.5. Pagination / filtering

List API cần hỗ trợ:

```text
page
page_size
sort
filter[field]
search
status
from_date
to_date
```

Với dữ liệu lớn, nên dùng cursor pagination cho log/ledger nếu cần.

### 12.6. Idempotency

Các API có thể bị gọi lặp cần idempotency key:

- tạo đơn từ external system;
- callback vận chuyển;
- import batch;
- payment callback;
- handover scan batch nếu scan nhiều lần.

Header:

```text
Idempotency-Key: <unique-key>
```

Backend lưu:

```text
key
request_hash
response_hash/status
created_at
expired_at
```

---

## 13. Auth, RBAC, approval

### 13.1. Auth flow

Khuyến nghị:

```text
Login
→ backend validate credentials
→ issue access token ngắn hạn
→ issue refresh token/session an toàn
→ frontend gọi API với token/cookie
```

Nếu triển khai nội bộ, có thể dùng session-based auth với secure cookie. Nếu cần API/mobile về sau, JWT hybrid linh hoạt hơn.

### 13.2. RBAC

RBAC phải bám theo `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`.

Kiến trúc:

```text
user
role
permission
role_permission
user_role
```

Permission format:

```text
module.resource.action
```

Ví dụ:

```text
inventory.stock_balance.view
inventory.stock_adjustment.create
inventory.stock_adjustment.approve
qc.inspection.pass
sales.order.discount_override
shipping.manifest.handover
returns.receipt.inspect
```

### 13.3. Field-level restriction

Không phải ai xem được document cũng xem được mọi field.

Ví dụ:

- kho không xem giá vốn nếu không được phép;
- sale không sửa QC status;
- QA không sửa giá bán;
- user thường không xem margin;
- kế toán xem payment/cost nhưng không đổi batch.

Backend phải enforce field-level restriction khi trả response hoặc khi update.

### 13.4. Approval engine

Phase 1 nên có workflow engine đơn giản, không cần BPMN phức tạp.

Entity:

```text
approval_request
approval_step
approval_action
approval_rule
```

Rule dựa trên:

```text
module
document_type
action
amount_threshold
warehouse
role
department
risk_level
```

Ví dụ:

```text
Stock adjustment > X cần trưởng kho + finance approve.
Discount > ngưỡng cần sales manager approve.
QC fail cần QA lead confirm.
Subcontract final payment cần finance + CEO/COO approve theo ngưỡng.
```

### 13.5. Approval không thay thế permission

Permission quyết định user có quyền đề xuất/hành động hay không.
Approval quyết định hành động đó có cần duyệt trước khi hiệu lực hay không.

---

## 14. Audit log architecture

### 14.1. Bắt buộc audit log cho critical action

Các hành động bắt buộc audit:

- login/logout quan trọng;
- tạo/sửa/xóa/inactive master data;
- thay đổi giá;
- tạo/duyệt/hủy PO;
- nhập kho;
- QC pass/fail/hold;
- reserve/release stock;
- xuất kho;
- stock adjustment;
- đóng hàng;
- bàn giao ĐVVC;
- nhận hàng hoàn;
- return disposition;
- chuyển NVL sang nhà máy;
- nhận hàng gia công;
- thanh toán/cọc;
- đổi role/permission.

### 14.2. Audit log structure

```text
audit_log
- id
- actor_user_id
- actor_role
- action
- module
- entity_type
- entity_id
- before_json
- after_json
- diff_json
- reason
- ip_address
- user_agent
- request_id
- created_at
```

### 14.3. Không cho user sửa audit log

Audit log là dữ liệu hệ thống. Không cho sửa/xóa qua UI.

Nếu cần archive, làm bằng job quản trị có kiểm soát.

---

## 15. Workflow kỹ thuật cho nghiệp vụ chính

### 15.1. Nhập kho + QC đầu vào

```text
User tạo receiving từ PO
→ Backend validate PO status
→ tạo inbound receipt
→ tạo batch/lot nếu cần
→ stock status = QC_HOLD hoặc PENDING_INSPECTION
→ tạo stock movement INBOUND_RECEIPT
→ tạo QC inspection request
→ audit log
→ outbox event InboundReceiptCreated
```

Khi QC pass:

```text
QA pass inspection
→ validate permission
→ update inspection result
→ update batch.qc_status = PASS
→ create movement QC_RELEASE
→ move qty from hold/quarantine to available
→ audit log
→ event QCInspectionPassed
```

Khi QC fail:

```text
QA fail inspection
→ update inspection result
→ batch.qc_status = FAIL
→ stock status = REJECTED/QUARANTINE
→ create movement QC_REJECT
→ block available
→ audit log
→ event QCInspectionFailed
```

### 15.2. Xuất kho bán hàng

```text
Sales order confirmed
→ inventory reserve stock theo FEFO
→ tạo stock reservation
→ order status = RESERVED hoặc PARTIALLY_RESERVED
```

Pick/Pack:

```text
Warehouse pick
→ scan SKU/batch/order
→ verify reserved batch
→ mark picked
→ pack by order/carrier/channel
→ shipment status = PACKED
```

Handover:

```text
Create carrier manifest
→ scan verify từng shipment/order
→ nếu đủ: manifest ready
→ handover confirmed
→ create SALES_ISSUE stock movement
→ order/shipment status = HANDED_OVER
→ audit log
→ event ShipmentHandedOver
```

### 15.3. Hàng hoàn

```text
Receive return from shipper
→ create return receipt
→ scan order/tracking/SKU if available
→ stock status = RETURN_PENDING_INSPECTION
→ create RETURN_RECEIPT movement
→ put into return area
```

Inspection:

```text
Inspect condition
→ usable: RETURN_TO_AVAILABLE hoặc cần QC_HOLD tùy rule
→ unusable: RETURN_TO_QUARANTINE / SCRAP / LAB
→ update return disposition
→ audit log
→ event ReturnInspected
```

### 15.4. Gia công ngoài

```text
Create subcontract order
→ confirm quantity/spec/sample requirement
→ deposit payment request nếu có
→ prepare material/packaging transfer
→ create inventory transfer to factory
→ issue materials out of internal warehouse
→ create SUBCONTRACT_TRANSFER_OUT movement
→ attach handover documents
```

Duyệt mẫu:

```text
Factory makes sample
→ upload sample evidence
→ R&D/QA/Brand approve sample
→ subcontract order status = SAMPLE_APPROVED
```

Nhận thành phẩm:

```text
Factory delivers finished goods
→ create receiving
→ verify quantity/spec
→ create batch
→ stock status = QC_HOLD
→ QC inspect
→ pass: available stock
→ fail: reject/claim factory within defined window
→ final payment request
```

### 15.5. Đóng ca kho / đối soát cuối ngày

```text
Warehouse shift close started
→ system shows open tasks:
   - pending inbound
   - pending outbound
   - pending packing
   - pending manifests
   - pending returns
   - stock count differences
→ warehouse performs cycle count if required
→ submit reconciliation
→ manager review difference
→ adjustment request if needed
→ close shift
→ daily warehouse report generated
```

---

## 16. Reporting architecture

### 16.1. Phase 1 report type

Có 3 loại report:

```text
1. Operational report
   dùng ngay trong ngày: tồn kho, đơn chờ pack, batch hold

2. Management dashboard
   CEO/manager xem: doanh thu, tồn, hàng hoàn, QC, giao hàng

3. Export report
   xuất Excel/CSV: stock ledger, sales order, PO, payment, QC
```

### 16.2. Query strategy

- Report nhỏ có thể query trực tiếp từ transactional tables.
- Report nặng nên chạy async export.
- Dashboard lặp nhiều nên dùng snapshot/materialized view/cache.
- Không làm dashboard nặng bằng cách join trực tiếp quá nhiều bảng mỗi lần user mở.

### 16.3. Daily snapshot

Có thể tạo snapshot cuối ngày:

```text
inventory_daily_snapshot
sales_daily_snapshot
shipping_daily_snapshot
returns_daily_snapshot
qc_daily_snapshot
```

Dùng cho dashboard nhanh và đối soát.

---

## 17. File/attachment architecture

### 17.1. File categories

```text
QC_EVIDENCE
SUPPLIER_DOCUMENT
PURCHASE_INVOICE
DELIVERY_NOTE
HANDOVER_PROOF
RETURN_PHOTO
SUBCONTRACT_CONTRACT
SAMPLE_APPROVAL_EVIDENCE
PAYMENT_PROOF
IMPORT_EXPORT_FILE
```

### 17.2. Upload flow

```text
Frontend request upload URL
→ Backend validates permission + entity
→ Backend returns pre-signed URL or upload endpoint
→ File stored in S3/MinIO
→ Backend records file metadata
→ Audit log
```

### 17.3. File access

File không nên public thẳng.

Download cần:

- check permission;
- generate signed URL tạm thời;
- audit nếu file nhạy cảm.

---

## 18. Integration architecture

### 18.1. Phase 1 integration level

Phase 1 chưa bắt buộc tích hợp realtime với mọi hệ thống ngoài, nhưng kiến trúc phải chuẩn bị sẵn.

Các nhóm tích hợp tương lai:

- website bán hàng;
- marketplace;
- POS;
- ĐVVC;
- accounting software;
- SMS/email/Zalo notification;
- barcode/QR scanner;
- BI tool.

### 18.2. Integration layer

Không để module nghiệp vụ gọi trực tiếp external API lung tung.

Tạo integration layer:

```text
internal/shared/integration/carrier
internal/shared/integration/marketplace
internal/shared/integration/payment
internal/shared/integration/notification
```

Hoặc tách theo module nếu rõ ownership.

### 18.3. Webhook handling

Webhook từ ngoài vào cần:

- validate signature nếu có;
- idempotency key;
- store raw payload;
- process async nếu nặng;
- audit/log;
- mapping external status sang internal status.

---

## 19. Security baseline

### 19.1. Authentication security

- Password hash an toàn.
- Không lưu password plaintext.
- Token/session có expiry.
- Logout/revoke token nếu cần.
- Lock/rate limit login sai nhiều lần.
- 2FA có thể đưa vào phase sau cho admin/finance.

### 19.2. Authorization security

- Backend enforce RBAC, không tin frontend.
- Field-level permission cho cost/margin/payment.
- Approval rule enforce ở backend.
- API critical cần kiểm tra trạng thái document.

### 19.3. Data security

- TLS cho môi trường production.
- Backup encrypted nếu được.
- File nhạy cảm không public.
- Audit critical data access nếu cần.
- Không log thông tin nhạy cảm như token/password.

### 19.4. Input security

- Validate input.
- Giới hạn file type/size.
- Sanitize filename.
- Chống SQL injection bằng parameterized query/sqlc.
- Chống mass assignment: không bind toàn bộ request vào entity rồi save bừa.

---

## 20. Observability

### 20.1. Logging

Log dạng structured JSON.

Log field tối thiểu:

```text
request_id
user_id
module
action
entity_id
status_code
latency_ms
error_code
```

Không log password/token/secret.

### 20.2. Metrics

Theo dõi:

- API latency;
- error rate;
- DB query latency;
- queue depth;
- job failure;
- login failure;
- stock reservation failure;
- handover scan mismatch;
- import/export duration.

### 20.3. Tracing

Nếu có điều kiện, dùng OpenTelemetry cho request flow:

```text
frontend request
→ backend handler
→ application service
→ database
→ queue/job
```

Phase 1 có thể triển khai tracing cơ bản, không cần quá phức tạp.

---

## 21. Performance baseline

### 21.1. Target thực dụng

| Use case | Target gợi ý |
|---|---|
| Mở list cơ bản | < 1 giây với pagination/filter tốt |
| Scan verify đơn | phản hồi gần realtime, ưu tiên < 300-500ms trong mạng nội bộ tốt |
| Reserve stock | transaction rõ, ưu tiên đúng hơn nhanh |
| Export report lớn | async job, không block UI |
| Dashboard CEO | dùng snapshot/cache nếu query nặng |

### 21.2. Indexing

Các bảng cần index tốt:

- `doc_no`;
- `status`;
- `created_at`;
- `warehouse_id`;
- `sku_id`;
- `batch_id`;
- `expiry_date`;
- `tracking_code`;
- `customer_id`;
- `supplier_id`;
- `external_ref`.

### 21.3. Query rule

- List API phải pagination.
- Không trả toàn bộ dữ liệu lớn.
- Không join vô tội vạ trong API list.
- Search nâng cao có thể tối ưu sau bằng index/full-text nếu cần.

---

## 22. Deployment architecture

### 22.1. Môi trường

Cần có tối thiểu:

```text
local
dev
staging
production
```

`staging` phải giống production nhất có thể về config quan trọng.

### 22.2. Docker services

```text
frontend
api
worker
postgres
redis
queue
minio
```

Local dev có thể dùng docker-compose.

Production có thể deploy bằng:

- Docker Compose trên server mạnh nếu quy mô nhỏ/vừa;
- Kubernetes nếu team DevOps đủ năng lực;
- Managed services nếu muốn giảm vận hành.

Phase 1 không bắt buộc Kubernetes.

### 22.3. Configuration

Dùng environment variables:

```text
APP_ENV
DATABASE_URL
REDIS_URL
QUEUE_URL
JWT_SECRET
S3_ENDPOINT
S3_ACCESS_KEY
S3_SECRET_KEY
LOG_LEVEL
```

Không commit secret vào git.

---

## 23. CI/CD baseline

Pipeline tối thiểu:

```text
1. checkout code
2. run gofmt/go vet/golangci-lint
3. run unit tests
4. run integration tests quan trọng
5. build Docker image
6. run migration dry-check nếu có
7. deploy to dev/staging
8. smoke test
9. manual approval to production
```

### 23.1. Migration rule

- Migration phải versioned.
- Không sửa migration cũ đã chạy production.
- Migration destructive phải có plan riêng.
- Có backup trước migration production.

### 23.2. Rollback

Không phải migration nào cũng rollback dễ. Vì vậy cần:

- backup database;
- blue/green hoặc release theo batch nếu có;
- feature flag cho chức năng mới;
- không deploy thay đổi lớn sát giờ vận hành kho cao điểm.

---

## 24. Testing architecture

### 24.1. Test pyramid

```text
Unit test
→ domain rule, validation, state machine

Integration test
→ repository, transaction, stock movement, DB constraints

API test
→ handler + auth + permission + response format

UAT
→ theo 09_ERP_UAT_Test_Scenarios
```

### 24.2. Critical test cases bắt buộc tự động hóa dần

1. Không cho xuất hàng khi batch QC hold/fail.
2. Không cho reserve vượt tồn khả dụng.
3. Một đơn scan bàn giao 2 lần không được trừ tồn 2 lần.
4. Return usable vào đúng trạng thái.
5. Return unusable không quay về available.
6. Subcontract transfer out tạo movement đúng.
7. Nhận thành phẩm gia công tạo batch QC hold.
8. Stock adjustment cần approval nếu vượt ngưỡng.
9. User không có quyền không gọi được API critical.
10. Optimistic lock phát hiện document bị sửa đồng thời.

---

## 25. State machine chuẩn Phase 1

### 25.1. Sales Order

```text
DRAFT
→ CONFIRMED
→ RESERVED / PARTIALLY_RESERVED
→ PICKED
→ PACKED
→ HANDED_OVER
→ DELIVERED
→ CLOSED
```

Ngoại lệ:

```text
CANCELLED
RETURNED
PARTIALLY_RETURNED
```

### 25.2. Shipment

```text
CREATED
→ PICKING
→ PICKED
→ PACKED
→ READY_FOR_HANDOVER
→ HANDED_OVER
→ IN_TRANSIT
→ DELIVERED
→ FAILED
→ RETURNED
```

### 25.3. Carrier Manifest

```text
DRAFT
→ SCANNING
→ VERIFIED
→ HANDED_OVER
→ CLOSED
```

Ngoại lệ:

```text
MISMATCH
CANCELLED
```

### 25.4. QC Inspection

```text
PENDING
→ IN_PROGRESS
→ PASSED / FAILED
```

Có thể có:

```text
ON_HOLD
RETEST_REQUIRED
```

### 25.5. Batch QC

```text
HOLD
→ PASS / FAIL
```

### 25.6. Return Receipt

```text
RECEIVED
→ PENDING_INSPECTION
→ INSPECTED
→ DISPOSED
→ CLOSED
```

Disposition:

```text
RETURN_TO_AVAILABLE
QUARANTINE
SCRAP
SEND_TO_LAB
SUPPLIER_FACTORY_CLAIM
```

### 25.7. Subcontract Order

```text
DRAFT
→ CONFIRMED
→ DEPOSIT_PAID
→ MATERIAL_TRANSFERRED
→ SAMPLE_PENDING
→ SAMPLE_APPROVED
→ MASS_PRODUCTION
→ DELIVERED_TO_WAREHOUSE
→ QC_CHECKING
→ ACCEPTED / REJECTED
→ FINAL_PAID
→ CLOSED
```

---

## 26. Data migration technical notes

Bám theo `10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1.md`.

### 26.1. Import tool

Backend nên có import framework:

```text
upload file
→ validate schema
→ dry-run
→ show errors
→ approve import
→ execute import
→ audit log
```

### 26.2. Import không được bỏ qua business rule

Ví dụ import tồn đầu kỳ vẫn phải:

- có SKU/material hợp lệ;
- có warehouse hợp lệ;
- có batch/hạn dùng nếu là hàng cần batch;
- có QC status ban đầu;
- tạo opening stock movement;
- ghi audit log.

### 26.3. Opening balance

Tồn đầu kỳ nên tạo movement:

```text
OPENING_BALANCE_IN
```

Không insert thẳng vào `stock_balances` rồi bỏ qua ledger.

---

## 27. Frontend architecture guideline ở mức technical architecture

Chi tiết UI/UX standard sẽ nằm ở tài liệu riêng, nhưng Phase 1 cần chốt baseline.

### 27.1. Page structure

```text
/pages or /app
  /inventory
  /sales
  /shipping
  /returns
/components
  /common
  /forms
  /tables
  /status-chip
  /scanner
/api
  generated-client or typed-fetch
```

### 27.2. UI state rule

- Server state dùng query library.
- Form state dùng form library.
- Global state chỉ dùng cho auth/user/settings.
- Không lưu business-critical state lâu ở frontend.

### 27.3. Scanner UX

Các màn hình scan phải:

- focus input mặc định;
- phản hồi nhanh;
- báo âm/thị giác khi scan sai;
- hiển thị số đã scan / tổng cần scan;
- cho supervisor xử lý exception có audit.

---

## 28. Coding standards cần tách tài liệu riêng

Tài liệu này chốt architecture. Coding standard chi tiết nên nằm ở file kế tiếp:

```text
12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md
```

Những thứ file 12 cần chốt:

- naming convention;
- folder/package rule;
- error handling;
- transaction pattern;
- repository pattern;
- DTO/entity mapping;
- logging;
- validation;
- testing;
- linting;
- code review checklist;
- commit/branch/release rule;
- SQL style;
- migration rule;
- API handler style.

---

## 29. Những quyết định chưa nên over-engineer

### 29.1. Không tách microservices ngay

Chỉ tách service khi có một trong các tín hiệu:

- module quá lớn và có team riêng;
- workload khác biệt lớn;
- cần scale độc lập;
- integration đặc thù;
- transaction boundary đã ổn định.

### 29.2. Không làm event sourcing full

Stock ledger là immutable ledger, nhưng toàn hệ thống không cần event sourcing full.

### 29.3. Không build BPMN engine phức tạp ngay

Approval/workflow Phase 1 dùng rule engine đơn giản là đủ.

### 29.4. Không nhồi BI nặng vào backend API chính

Report nặng nên async hoặc tách report layer.

---

## 30. Roadmap kỹ thuật Phase 1

### Milestone 1: Foundation

- Project structure.
- Auth/RBAC.
- Migration setup.
- API response/error convention.
- Audit log base.
- File storage base.
- OpenAPI setup.

### Milestone 2: Master Data

- SKU/material/warehouse/supplier/customer/carrier.
- Unit conversion.
- Batch config.
- Reason code.
- Active/inactive + audit.

### Milestone 3: Inventory Core

- Batch.
- Stock movement.
- Stock balance.
- Reservation.
- Stock count.
- Adjustment.

### Milestone 4: Procurement + QC

- PO.
- Receiving.
- QC inspection.
- QC release/reject.
- Supplier receipt.

### Milestone 5: Sales + Shipping

- Sales order.
- Reserve stock.
- Pick/pack.
- Shipment.
- Carrier manifest.
- Handover scan verify.

### Milestone 6: Returns + Subcontract

- Return receiving.
- Return inspection/disposition.
- Subcontract order.
- Material transfer to factory.
- Finished goods receipt.

### Milestone 7: Reporting + UAT hardening

- Operational reports.
- Dashboard baseline.
- Export jobs.
- UAT fixes.
- Performance hardening.

---

## 31. Risk list kỹ thuật

| Rủi ro | Tác động | Cách kiểm soát |
|---|---|---|
| Team Go chưa đồng đều | Code rời rạc | Coding standard + tech lead review |
| Stock logic sai | Sai tồn, sai bán hàng | Stock ledger + transaction test |
| Không enforce permission backend | Lộ/sửa dữ liệu sai | RBAC middleware + API test |
| Import dữ liệu bẩn | Go-live sai số | Dry-run import + validation + audit |
| Không có outbox | Event/job mất hoặc chạy sai | Outbox pattern |
| Report query quá nặng | Hệ thống chậm | Snapshot/cache/async export |
| Workflow thực tế thay đổi | Build lệch | As-Is/GAP update trước sprint |
| Tích hợp vận chuyển callback trùng | Sai trạng thái | Idempotency key + raw payload log |
| Scan bàn giao trừ tồn lặp | Lệch kho | idempotency + transaction + status guard |
| Không khóa trạng thái document | User làm tắt | state machine enforced backend |

---

## 32. Definition of Done cho architecture Phase 1

Một module backend được coi là đạt chuẩn architecture khi:

1. Có API contract OpenAPI.
2. Có handler/application/domain/repository rõ.
3. Không gọi chéo database module khác sai boundary.
4. Critical action có transaction.
5. Critical action có audit log.
6. Có permission check.
7. Có state validation.
8. Có test cho rule quan trọng.
9. Có migration versioned.
10. Có error response chuẩn.
11. Có logging request_id.
12. Có UAT scenario liên quan pass.

---

## 33. Kết luận

Backend Go là lựa chọn phù hợp cho ERP Phase 1 nếu dự án chốt architecture và coding standard ngay từ đầu.

Hướng chính:

```text
Frontend: TypeScript + React / Next.js
Backend: Go
Database: PostgreSQL
Architecture: Modular Monolith
API: REST + OpenAPI
Stock: immutable stock ledger + stock balance snapshot
Workflow: RBAC + approval + state machine
Async: outbox + queue + worker
Storage: S3/MinIO
Deployment: Docker
```

Phần đặc thù từ workflow thực tế cần được đưa thẳng vào architecture:

- kho có kiểm kê/đối soát cuối ngày;
- nhập/xuất/đóng/hàng hoàn tách nhánh rõ;
- bàn giao ĐVVC cần scan verify và manifest;
- sản xuất có nhánh gia công ngoài, chuyển NVL/bao bì, duyệt mẫu, nhận hàng và QC.

Câu chốt:

> Đừng để Go chỉ là ngôn ngữ. Hãy biến Go thành một backend có kỷ luật: module rõ, transaction chắc, stock ledger bất biến, audit đầy đủ, API thống nhất. Đó mới là nền đủ bền để ERP sống lâu.

---

## 34. Tài liệu tiếp theo nên làm

Sau tài liệu này, cần làm tiếp:

```text
12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md
13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md
14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md
```

Trong đó, file 12 là ưu tiên cao nhất vì nó khóa cách đội dev viết code hằng ngày.
