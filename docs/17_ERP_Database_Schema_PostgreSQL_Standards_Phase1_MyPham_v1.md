# 17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1

**Dự án:** Web ERP công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** Database Schema + PostgreSQL Standards  
**Scope:** Phase 1  
**Version:** v1.0  
**Date:** 2026-04-24  
**Language:** Vietnamese  
**Backend:** Go Modular Monolith  
**Database:** PostgreSQL  
**API:** REST + OpenAPI 3.1  
**Frontend:** React / Next.js + TypeScript  
**Owner:** ERP Solution Architect / Technical Lead / Backend Lead / DBA  

**Related Documents:**

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
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`
- `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`
- `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`
- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt **chuẩn thiết kế database PostgreSQL** cho ERP Phase 1.

Nó trả lời các câu hỏi đội backend Go, DBA, QA, DevOps, data migration và BI cần thống nhất trước khi build:

1. Database chia schema/module thế nào.
2. Table đặt tên thế nào.
3. Primary key, document number, foreign key, index, constraint dùng chuẩn gì.
4. Stock ledger thiết kế thế nào để không lệch tồn.
5. Batch, hạn dùng, QC, hàng hoàn, bàn giao ĐVVC và gia công ngoài được lưu thế nào.
6. Trạng thái chứng từ lưu ra sao.
7. Audit log, outbox event, idempotency key, file attachment được thiết kế thế nào.
8. Migration script, seed data, backup/restore, test data cần chuẩn gì.
9. Database rule nào là bắt buộc, rule nào là gợi ý.

Tài liệu này không thay thế Data Dictionary.  
Data Dictionary nói **mỗi dữ liệu nghĩa là gì**.  
Database Standards nói **dữ liệu đó được lưu thế nào để hệ thống chạy bền, đúng, dễ mở rộng**.

Một câu chốt:

> ERP sống nhờ dữ liệu đúng.  
> Database là nơi giữ lời thề của hệ thống: hàng không tự sinh ra, tiền không tự mất đi, batch không được nhập nhầm, và mọi thay đổi quan trọng phải có dấu vết.

---

## 2. Business reality anchor từ workflow thực tế

Thiết kế database Phase 1 phải bám theo thực tế vận hành đã được cung cấp trong 4 tài liệu nội bộ.

### 2.1. Kho có nhịp vận hành theo ngày

Workflow hằng ngày của kho gồm:

```text
Tiếp nhận đơn hàng trong ngày
→ Thực hiện xuất/nhập theo nội quy
→ Soạn hàng và đóng gói
→ Sắp xếp, tối ưu vị trí kho
→ Kiểm kê hàng tồn cuối ngày
→ Đối soát số liệu và báo cáo quản lý
→ Kết thúc ca làm
```

Hệ quả database:

- Cần có bảng ghi nhận **warehouse daily shift / daily closing**.
- Cần lưu số liệu kiểm kê cuối ngày.
- Cần lưu chênh lệch kiểm kê và lý do điều chỉnh.
- Không được chỉ có bảng `stock_balance` tĩnh.
- Phải có `stock_ledger` bất biến làm nguồn sự thật.

### 2.2. Nội quy kho tách 4 nhánh nghiệp vụ

Nội quy đang có 4 quy trình rõ:

1. Nhập kho.
2. Xuất kho.
3. Đóng hàng.
4. Xử lý hàng hoàn.

Hệ quả database:

- Nhập kho cần `goods_receipts`, `goods_receipt_lines`, batch, hạn dùng, tình trạng QC.
- Xuất kho cần `stock_issue`, `picking`, `packing`, `shipment` hoặc liên kết với sales order.
- Đóng hàng cần lưu trạng thái `picked`, `packed`, `ready_for_handover`.
- Hàng hoàn cần `returns`, `return_inspections`, `return_dispositions`.
- Hàng hoàn còn dùng được và không dùng được phải đi vào trạng thái tồn khác nhau.

### 2.3. Bàn giao ĐVVC cần quét mã và đối chiếu

Quy trình bàn giao cho đơn vị vận chuyển gồm:

```text
Phân chia khu vực để hàng
→ Để theo thùng/rổ, mỗi thùng/rổ có số lượng bằng nhau nếu có thể
→ Đối chiếu số lượng đơn dựa trên bảng
→ Lấy hàng và quét mã trực tiếp tại hàm
→ Nếu đủ đơn thì ký xác nhận bàn giao với ĐVVC
→ Nếu chưa đủ thì kiểm tra lại mã hoặc tìm lại trong khu đóng hàng
```

Hệ quả database:

- Cần `carrier_manifests` và `carrier_manifest_lines`.
- Cần `scan_events` để lưu từng lần quét.
- Cần `shipment_handover_confirmations` hoặc status trên manifest.
- Cần lưu thiếu/thừa đơn lúc bàn giao.
- Cần idempotency để quét trùng không làm sai trạng thái.

### 2.4. Sản xuất thực tế có nhánh gia công ngoài

Quy trình sản xuất cho thấy luồng:

```text
Lên đơn hàng với nhà máy
→ Xác nhận số lượng/quy cách/mẫu mã
→ Cọc đơn hàng và xác nhận thời gian sản xuất/nhận hàng
→ Chuyển nguyên vật liệu/bao bì qua nhà máy
→ Làm mẫu và chốt mẫu
→ Sản xuất hàng loạt
→ Giao hàng về kho
→ Kiểm tra số lượng/chất lượng
→ Nhận hàng hoặc báo lỗi nhà máy trong vòng 3–7 ngày
→ Thanh toán lần cuối
```

Hệ quả database:

- Không chỉ có `production_orders` kiểu xưởng nội bộ.
- Cần module `subcontract` hoặc bảng hỗ trợ gia công ngoài.
- Cần lưu chuyển NVL/bao bì sang nhà máy.
- Cần lưu sample approval.
- Cần lưu deposit/final payment reference.
- Cần lưu inbound QC khi hàng từ nhà máy về.
- Cần trace nguyên vật liệu gửi đi và thành phẩm nhận về.

---

## 3. Kết luận database đã chốt

| Hạng mục | Quyết định |
|---|---|
| DB Engine | PostgreSQL |
| Architecture | Database chung cho Go modular monolith, nhưng tách schema theo module |
| Primary Key | UUID, ưu tiên sinh từ application bằng UUIDv7/ULID-compatible nếu team đã chuẩn hóa; fallback `gen_random_uuid()` |
| Document Number | Mã nghiệp vụ riêng, không dùng primary key làm số chứng từ |
| Quantity Type | `numeric(18,4)` |
| Money Type | `numeric(18,2)` hoặc `numeric(18,4)` nếu cần phân bổ chi tiết |
| Date/Time | `timestamptz` cho thời điểm; `date` cho ngày sản xuất/hạn dùng |
| Status | `text` + `check constraint` hoặc lookup table, không update tùy tiện |
| Source of Truth tồn kho | `inventory.stock_ledger` bất biến |
| Balance Table | Có `inventory.stock_balances` để đọc nhanh, cập nhật trong transaction |
| Audit | Bảng `audit.audit_logs`, bắt buộc với mutation quan trọng |
| Outbox | `integration.outbox_events`, dùng cho event async |
| Idempotency | `core.idempotency_keys`, bắt buộc cho mutation dễ double-submit/scan |
| Attachment | `file.attachments`, lưu metadata, file thật ở S3/MinIO |
| Soft Delete | Không soft delete giao dịch; master data dùng `status`/`is_active` |
| Migration Tool | `golang-migrate` hoặc tool tương đương, migration versioned |
| SQL Access | Go dùng `pgx + sqlc` nếu theo hướng explicit SQL |

---

## 4. Nguyên tắc thiết kế database

### 4.1. Ledger trước, balance sau

Bảng balance chỉ dùng để đọc nhanh.  
Sự thật nằm ở ledger.

Không được cho bất kỳ module nào sửa trực tiếp tồn kho theo kiểu:

```sql
UPDATE stock_balances SET qty_on_hand = qty_on_hand - 1;
```

Mọi thay đổi tồn phải đi qua movement:

```text
create stock ledger row
→ update stock balance trong cùng transaction
→ ghi audit log
→ phát outbox event nếu cần
```

### 4.2. Không xóa giao dịch nghiệp vụ

Không xóa:

- sales order
- purchase order
- goods receipt
- stock movement
- QC inspection
- return
- shipment
- carrier manifest
- payment transaction
- audit log

Chỉ được:

- cancel
- reverse
- adjustment
- void có reason

### 4.3. Master data có thể inactive, không xóa tùy tiện

Ví dụ:

- item ngừng bán → `status = 'inactive'`
- supplier ngừng hợp tác → `status = 'inactive'`
- warehouse không dùng nữa → `status = 'inactive'`

Không xóa vì transaction cũ vẫn cần reference.

### 4.4. Document number không phải primary key

Mỗi bảng nghiệp vụ nên có:

```text
id              UUID technical primary key
code/doc_no     mã chứng từ hiển thị cho người dùng
```

Ví dụ:

```text
id = 0f5f8f7a-...
po_no = PO-20260424-000123
```

Lý do:

- `id` dùng cho hệ thống.
- `po_no` dùng cho người dùng, in ấn, tìm kiếm, đối soát.
- Sau này đổi format số chứng từ không ảnh hưởng khóa chính.

### 4.5. Tất cả mutation quan trọng phải có người tạo và thời điểm

Tất cả bảng nghiệp vụ phải có:

```text
created_at
created_by
updated_at
updated_by
```

Bảng giao dịch quan trọng có thêm:

```text
submitted_at / submitted_by
approved_at / approved_by
cancelled_at / cancelled_by
cancel_reason
```

### 4.6. Status không được là chữ tự do

Không được để status kiểu nhập tay tùy ý.

Sai:

```text
status = 'ok'
status = 'xong'
status = 'done rồi'
```

Đúng:

```text
status IN ('draft', 'submitted', 'approved', 'rejected', 'cancelled')
```

### 4.7. Không dùng float cho số lượng và tiền

Không dùng:

```sql
float
real
double precision
```

Dùng:

```sql
numeric(18,4) -- quantity
numeric(18,2) -- money
```

### 4.8. Batch và hạn dùng là first-class data

Với mỹ phẩm, `batch_no`, `mfg_date`, `expiry_date`, `qc_status` không phải dữ liệu phụ.

Chúng là dữ liệu sống còn.

Bất kỳ luồng nào chạm hàng hóa đều phải nghĩ tới:

- item
- batch
- warehouse
- location/bin
- quantity
- QC status
- expiry date

### 4.9. Integration và scan phải có idempotency

Các action dễ bị lặp:

- scan đơn
- scan barcode
- handover shipment
- create order từ website/sàn
- receive webhook shipping
- submit payment

Phải có idempotency key hoặc unique constraint chống double-processing.

### 4.10. Reporting không được phá transaction core

Bảng core phục vụ giao dịch.  
Báo cáo nặng nên dùng:

- read replica nếu có
- materialized view
- reporting table
- async export job

Không cho dashboard chạy query quá nặng trực tiếp vào bảng core trong giờ cao điểm.

---

## 5. Schema organization

Database chia schema theo module.

```text
core          user, role, permission, idempotency, numbering
workflow      approval flow, approval request, approval step
mdm           master data: item, supplier, customer, warehouse, carrier
inventory     stock ledger, stock balance, reservation, count, transfer
purchase      PR, PO, inbound purchase reference
qc            inspection, QC result, batch release
production    internal production nếu dùng trong Phase 1
subcontract   gia công ngoài, material issue to factory, sample approval
sales         quotation, sales order, order line, discount snapshot
shipping      shipment, package, carrier manifest, scan event
returns       return order, return inspection, disposition
finance       AR/AP basic, payment request, payment allocation
file          attachment metadata
integration   outbox events, external sync log, webhook inbox
audit         audit logs
reporting     materialized/report views nếu cần
```

### 5.1. Vì sao dùng PostgreSQL schema theo module?

Vì ERP là modular monolith.  
Dùng database schema giúp:

- nhìn rõ module ownership
- giảm cảnh table nằm lẫn trong `public`
- dễ quản lý permission nội bộ
- dễ generate sqlc package theo module
- dễ trace khi một module thay đổi dữ liệu

### 5.2. Quy tắc ownership

Mỗi table có một module owner.

Ví dụ:

| Table | Owner |
|---|---|
| `inventory.stock_ledger` | Inventory |
| `inventory.stock_balances` | Inventory |
| `sales.sales_orders` | Sales |
| `shipping.shipments` | Shipping |
| `returns.return_orders` | Returns |
| `qc.inspections` | QC |
| `subcontract.subcontract_orders` | Subcontract/Production |

Module khác không update trực tiếp table owner.

Ví dụ:

- Sales không update `inventory.stock_balances` trực tiếp.
- Shipping không update `sales.sales_orders` trực tiếp trừ khi qua application service/transaction đã được định nghĩa.
- Returns không tự đưa hàng về available stock nếu chưa qua inspection/disposition.

---

## 6. Naming convention

### 6.1. Table name

Dùng plural snake_case.

```text
sales_orders
sales_order_lines
purchase_orders
purchase_order_lines
stock_ledger
stock_balances
carrier_manifests
carrier_manifest_lines
```

Không dùng:

```text
SalesOrder
SALE_ORDER
saleOrder
DonHang
```

### 6.2. Column name

Dùng snake_case.

```text
id
org_id
item_id
warehouse_id
batch_id
created_at
created_by
```

### 6.3. Primary key

Luôn là:

```text
id uuid primary key
```

### 6.4. Foreign key

Dạng:

```text
{entity}_id
```

Ví dụ:

```text
supplier_id
customer_id
warehouse_id
sales_order_id
```

### 6.5. Constraint name

```text
pk_{table}
fk_{table}_{column}_{ref_table}
uq_{table}_{columns}
ck_{table}_{rule}
ix_{table}_{columns}
```

Ví dụ:

```text
pk_sales_orders
fk_sales_orders_customer_id_customers
uq_items_sku
ck_batches_expiry_after_mfg
ix_stock_ledger_item_batch_warehouse
```

### 6.6. Index name

```text
ix_{table}_{column1}_{column2}
```

Unique index:

```text
uq_{table}_{column1}_{column2}
```

Partial index:

```text
ix_{table}_{condition_hint}
```

Ví dụ:

```text
ix_sales_orders_status_order_date
uq_items_sku
ix_stock_balances_available_positive
```

---

## 7. Data types chuẩn

| Loại dữ liệu | PostgreSQL type | Rule |
|---|---|---|
| ID | `uuid` | Sinh từ app hoặc DB fallback |
| Mã chứng từ | `text` | Có unique constraint theo org nếu cần |
| Status | `text` | Có check constraint |
| Số lượng | `numeric(18,4)` | Không dùng float |
| Tiền | `numeric(18,2)` | Có currency nếu multi-currency |
| Tỷ lệ | `numeric(9,4)` | Ví dụ discount %, loss rate |
| Ngày | `date` | MFG date, expiry date |
| Thời điểm | `timestamptz` | created_at, approved_at, scanned_at |
| Boolean | `boolean` | Tránh text yes/no |
| JSON | `jsonb` | Chỉ dùng cho metadata phụ, không thay schema chính |
| Text dài | `text` | remark, reason, description |
| Search text | `text` + index phù hợp | Không lạm dụng |

### 7.1. `jsonb` dùng khi nào?

Được dùng cho:

- external payload snapshot
- audit before/after
- dynamic metadata không phải core
- webhook body
- report filter saved config

Không dùng `jsonb` để nhét dữ liệu core như:

- order lines
- stock qty
- QC result chính
- item master
- batch info

Sai:

```text
sales_orders.lines jsonb
```

Đúng:

```text
sales.sales_order_lines
```

---

## 8. Common columns chuẩn

### 8.1. Bảng master data

```sql
id uuid primary key,
org_id uuid not null,
code text not null,
name text not null,
status text not null default 'active',
is_active boolean not null default true,
created_at timestamptz not null default now(),
created_by uuid not null,
updated_at timestamptz not null default now(),
updated_by uuid,
version int not null default 1
```

### 8.2. Bảng transaction header

```sql
id uuid primary key,
org_id uuid not null,
doc_no text not null,
status text not null,
created_at timestamptz not null default now(),
created_by uuid not null,
updated_at timestamptz not null default now(),
updated_by uuid,
submitted_at timestamptz,
submitted_by uuid,
approved_at timestamptz,
approved_by uuid,
cancelled_at timestamptz,
cancelled_by uuid,
cancel_reason text,
version int not null default 1
```

### 8.3. Bảng transaction line

```sql
id uuid primary key,
org_id uuid not null,
header_id uuid not null,
line_no int not null,
item_id uuid not null,
qty numeric(18,4) not null,
uom_id uuid not null,
created_at timestamptz not null default now(),
created_by uuid not null,
updated_at timestamptz not null default now(),
updated_by uuid
```

### 8.4. Bảng scan event

```sql
id uuid primary key,
org_id uuid not null,
scan_type text not null,
barcode text not null,
source_doc_type text,
source_doc_id uuid,
result text not null,
error_code text,
scanned_at timestamptz not null default now(),
scanned_by uuid not null,
station_id text,
metadata jsonb
```

---

## 9. Core schemas và table catalog Phase 1

### 9.1. `core`

| Table | Mục đích |
|---|---|
| `core.users` | Người dùng hệ thống |
| `core.roles` | Vai trò |
| `core.permissions` | Quyền action/module |
| `core.user_roles` | Gán role cho user |
| `core.role_permissions` | Gán permission cho role |
| `core.idempotency_keys` | Chống double submit/scan/webhook |
| `core.document_sequences` | Sinh số chứng từ |
| `core.user_sessions` | Session nếu dùng |

### 9.2. `workflow`

| Table | Mục đích |
|---|---|
| `workflow.approval_flows` | Cấu hình luồng duyệt |
| `workflow.approval_flow_steps` | Bước duyệt |
| `workflow.approval_requests` | Yêu cầu duyệt thực tế |
| `workflow.approval_actions` | Lịch sử approve/reject |

### 9.3. `mdm`

| Table | Mục đích |
|---|---|
| `mdm.organizations` | Công ty/pháp nhân |
| `mdm.brands` | Brand mỹ phẩm |
| `mdm.units` | Đơn vị tính |
| `mdm.unit_conversions` | Quy đổi đơn vị |
| `mdm.item_categories` | Nhóm hàng |
| `mdm.items` | SKU/nguyên liệu/bao bì/thành phẩm |
| `mdm.item_boms` | BOM/công thức cơ bản Phase 1 |
| `mdm.item_bom_lines` | Thành phần BOM |
| `mdm.suppliers` | Nhà cung cấp/nhà máy |
| `mdm.customers` | Khách hàng/đại lý |
| `mdm.warehouses` | Kho |
| `mdm.warehouse_zones` | Khu vực kho |
| `mdm.warehouse_bins` | Vị trí/bin |
| `mdm.carriers` | Đơn vị vận chuyển |

### 9.4. `purchase`

| Table | Mục đích |
|---|---|
| `purchase.purchase_requisitions` | Đề nghị mua |
| `purchase.purchase_requisition_lines` | Dòng đề nghị mua |
| `purchase.purchase_orders` | PO |
| `purchase.purchase_order_lines` | Dòng PO |
| `purchase.supplier_delivery_notes` | Chứng từ giao hàng NCC nếu cần |

### 9.5. `qc`

| Table | Mục đích |
|---|---|
| `qc.inspections` | Phiếu kiểm QC |
| `qc.inspection_lines` | Dòng kiểm QC |
| `qc.inspection_results` | Kết quả chi tiết nếu cần |
| `qc.batch_quality_statuses` | Lịch sử đổi trạng thái QC batch |
| `qc.non_conformances` | Lỗi không phù hợp |

### 9.6. `inventory`

| Table | Mục đích |
|---|---|
| `inventory.batches` | Batch/lot |
| `inventory.stock_ledger` | Sổ cái tồn kho bất biến |
| `inventory.stock_balances` | Tồn đọc nhanh |
| `inventory.stock_reservations` | Giữ hàng cho đơn/lệnh |
| `inventory.goods_receipts` | Phiếu nhập kho |
| `inventory.goods_receipt_lines` | Dòng phiếu nhập |
| `inventory.stock_issues` | Phiếu xuất kho |
| `inventory.stock_issue_lines` | Dòng phiếu xuất |
| `inventory.stock_transfers` | Phiếu chuyển kho/vị trí |
| `inventory.stock_transfer_lines` | Dòng chuyển kho |
| `inventory.stock_counts` | Phiếu kiểm kê |
| `inventory.stock_count_lines` | Dòng kiểm kê |
| `inventory.warehouse_daily_closings` | Đối soát/kết ca kho cuối ngày |

### 9.7. `sales`

| Table | Mục đích |
|---|---|
| `sales.sales_orders` | Đơn bán |
| `sales.sales_order_lines` | Dòng đơn bán |
| `sales.price_snapshots` | Snapshot giá/discount nếu cần |
| `sales.order_status_history` | Lịch sử trạng thái đơn |

### 9.8. `shipping`

| Table | Mục đích |
|---|---|
| `shipping.shipments` | Phiếu giao/vận đơn nội bộ |
| `shipping.shipment_lines` | Dòng shipment |
| `shipping.packages` | Kiện/thùng/rổ nếu cần |
| `shipping.carrier_manifests` | Bảng kê bàn giao ĐVVC |
| `shipping.carrier_manifest_lines` | Dòng bảng kê |
| `shipping.scan_events` | Log quét mã |
| `shipping.handover_issues` | Sự cố thiếu/thừa đơn khi bàn giao |

### 9.9. `returns`

| Table | Mục đích |
|---|---|
| `returns.return_orders` | Phiếu hàng hoàn/đổi trả |
| `returns.return_order_lines` | Dòng hàng hoàn |
| `returns.return_inspections` | Kiểm tra tình trạng hàng hoàn |
| `returns.return_dispositions` | Quyết định xử lý: restock, lab, scrap, quarantine |

### 9.10. `subcontract`

| Table | Mục đích |
|---|---|
| `subcontract.subcontract_orders` | Đơn gia công với nhà máy |
| `subcontract.subcontract_order_lines` | Thành phẩm đặt gia công |
| `subcontract.material_issues` | Chuyển NVL/bao bì cho nhà máy |
| `subcontract.material_issue_lines` | Dòng NVL/bao bì chuyển đi |
| `subcontract.sample_approvals` | Làm mẫu/chốt mẫu |
| `subcontract.subcontract_receipts` | Nhận hàng gia công về kho |
| `subcontract.subcontract_receipt_lines` | Dòng nhận hàng gia công |
| `subcontract.factory_claims` | Báo lỗi nhà máy trong 3–7 ngày |

### 9.11. `finance`

| Table | Mục đích |
|---|---|
| `finance.payment_requests` | Yêu cầu thanh toán |
| `finance.payments` | Ghi nhận thanh toán |
| `finance.ar_entries` | Phải thu cơ bản |
| `finance.ap_entries` | Phải trả cơ bản |
| `finance.cod_reconciliations` | Đối soát COD nếu Phase 1 làm |

### 9.12. `audit`, `file`, `integration`

| Schema | Table | Mục đích |
|---|---|---|
| `audit` | `audit_logs` | Lưu ai làm gì, trước/sau |
| `file` | `attachments` | File metadata |
| `integration` | `outbox_events` | Event cần publish |
| `integration` | `webhook_inbox` | Webhook nhận từ ngoài |
| `integration` | `external_sync_logs` | Log sync |

---

## 10. ERD logic tổng quát Phase 1

```text
mdm.items
  ├── inventory.batches
  ├── purchase.purchase_order_lines
  ├── inventory.goods_receipt_lines
  ├── inventory.stock_ledger
  ├── sales.sales_order_lines
  ├── returns.return_order_lines
  └── subcontract.material_issue_lines / subcontract_receipt_lines

mdm.warehouses
  ├── mdm.warehouse_zones
  ├── mdm.warehouse_bins
  ├── inventory.stock_balances
  ├── inventory.stock_ledger
  ├── inventory.stock_counts
  └── inventory.warehouse_daily_closings

purchase.purchase_orders
  └── inventory.goods_receipts
        └── qc.inspections
              └── inventory.batches / stock_ledger after QC pass

sales.sales_orders
  ├── inventory.stock_reservations
  ├── shipping.shipments
  │     └── shipping.carrier_manifests
  └── returns.return_orders

subcontract.subcontract_orders
  ├── subcontract.material_issues
  │     └── inventory.stock_ledger: SUBCONTRACT_ISSUE
  ├── subcontract.sample_approvals
  └── subcontract.subcontract_receipts
        └── inventory.goods_receipts / qc.inspections / stock_ledger
```

---

## 11. Master data schema standards

### 11.1. `mdm.items`

`items` là bảng xương sống cho:

- nguyên vật liệu
- bao bì
- bán thành phẩm
- thành phẩm
- sample/tester nếu quản lý như SKU riêng
- quà tặng nếu quản lý tồn riêng

Gợi ý DDL:

```sql
create table mdm.items (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  brand_id uuid,
  sku text not null,
  name text not null,
  item_type text not null,
  category_id uuid,
  base_uom_id uuid not null,
  barcode text,
  track_batch boolean not null default true,
  track_expiry boolean not null default true,
  qc_required boolean not null default true,
  shelf_life_days int,
  status text not null default 'active',
  is_sellable boolean not null default false,
  is_purchasable boolean not null default false,
  is_manufacturable boolean not null default false,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  version int not null default 1,
  constraint uq_items_org_sku unique (org_id, sku),
  constraint ck_items_item_type check (item_type in (
    'raw_material', 'packaging', 'semi_finished', 'finished_good', 'sample', 'tester', 'gift', 'service'
  )),
  constraint ck_items_status check (status in ('draft', 'active', 'inactive', 'blocked'))
);
```

Rule:

- Thành phẩm mỹ phẩm mặc định `track_batch = true`, `track_expiry = true`.
- Nguyên liệu mỹ phẩm mặc định `qc_required = true`.
- Item chưa `active` không được đưa vào PO/SO/lệnh sản xuất mới.
- Không đổi `base_uom_id` nếu item đã phát sinh giao dịch.

### 11.2. `mdm.warehouses`, `warehouse_zones`, `warehouse_bins`

Kho phải hỗ trợ:

- kho nguyên liệu
- kho bao bì
- kho thành phẩm
- kho hàng hoàn
- khu quarantine
- khu đóng hàng
- khu bàn giao ĐVVC
- khu lab/hàng không sử dụng

```sql
create table mdm.warehouses (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  code text not null,
  name text not null,
  warehouse_type text not null,
  status text not null default 'active',
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_warehouses_org_code unique (org_id, code),
  constraint ck_warehouses_type check (warehouse_type in (
    'main', 'factory', 'store', 'return', 'quarantine', 'sample', 'subcontractor'
  ))
);

create table mdm.warehouse_zones (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  warehouse_id uuid not null references mdm.warehouses(id),
  code text not null,
  name text not null,
  zone_type text not null,
  status text not null default 'active',
  constraint uq_warehouse_zones_wh_code unique (warehouse_id, code),
  constraint ck_warehouse_zones_type check (zone_type in (
    'receiving', 'storage', 'picking', 'packing', 'handover', 'return', 'quarantine', 'lab', 'scrap'
  ))
);

create table mdm.warehouse_bins (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  warehouse_id uuid not null references mdm.warehouses(id),
  zone_id uuid references mdm.warehouse_zones(id),
  code text not null,
  name text,
  status text not null default 'active',
  constraint uq_warehouse_bins_wh_code unique (warehouse_id, code)
);
```

Rule:

- Hàng QC hold phải nằm ở quarantine hoặc trạng thái không khả dụng.
- Hàng hoàn chưa kiểm tra phải vào return/quarantine zone.
- Hàng không dùng được không được vào available stock.

### 11.3. `mdm.suppliers`

Supplier gồm cả:

- nhà cung cấp nguyên liệu/bao bì
- nhà máy gia công
- đơn vị dịch vụ nếu cần

```sql
create table mdm.suppliers (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  supplier_code text not null,
  name text not null,
  supplier_type text not null,
  tax_code text,
  phone text,
  email text,
  address text,
  payment_terms text,
  status text not null default 'active',
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_suppliers_org_code unique (org_id, supplier_code),
  constraint ck_suppliers_type check (supplier_type in ('material', 'packaging', 'factory', 'service', 'carrier_related'))
);
```

---

## 12. Inventory schema standards

Inventory là lõi sống còn.  
Mọi sai sót ở đây sẽ lan sang sales, shipping, finance, report và CEO dashboard.

### 12.1. Batch table

```sql
create table inventory.batches (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  item_id uuid not null references mdm.items(id),
  batch_no text not null,
  supplier_batch_no text,
  mfg_date date,
  expiry_date date,
  qc_status text not null default 'hold',
  batch_status text not null default 'active',
  source_doc_type text,
  source_doc_id uuid,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_batches_item_batch unique (org_id, item_id, batch_no),
  constraint ck_batches_qc_status check (qc_status in ('hold', 'pass', 'fail', 'conditional_pass')),
  constraint ck_batches_status check (batch_status in ('active', 'blocked', 'expired', 'recalled', 'closed')),
  constraint ck_batches_expiry_after_mfg check (expiry_date is null or mfg_date is null or expiry_date >= mfg_date)
);
```

Rule:

- Batch mới từ nhập kho hoặc sản xuất/gia công mặc định `qc_status = 'hold'` nếu item yêu cầu QC.
- Batch `fail` không được xuất bán.
- Batch `hold` không được xuất bán, trừ case đặc biệt được QA/CEO duyệt rõ.
- Batch hết hạn không được bán.

### 12.2. Stock ledger

`stock_ledger` là bảng quan trọng nhất của inventory.

```sql
create table inventory.stock_ledger (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  movement_no text not null,
  movement_type text not null,
  movement_at timestamptz not null default now(),
  item_id uuid not null references mdm.items(id),
  batch_id uuid references inventory.batches(id),
  warehouse_id uuid not null references mdm.warehouses(id),
  bin_id uuid references mdm.warehouse_bins(id),
  qty numeric(18,4) not null,
  uom_id uuid not null references mdm.units(id),
  direction text not null,
  source_doc_type text not null,
  source_doc_id uuid not null,
  source_doc_line_id uuid,
  reason_code text,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  constraint uq_stock_ledger_movement_no unique (org_id, movement_no),
  constraint ck_stock_ledger_qty_positive check (qty > 0),
  constraint ck_stock_ledger_direction check (direction in ('in', 'out')),
  constraint ck_stock_ledger_type check (movement_type in (
    'purchase_receipt',
    'qc_release',
    'sales_reserve',
    'sales_unreserve',
    'sales_issue',
    'return_receipt',
    'return_restock',
    'return_to_lab',
    'production_issue',
    'production_receipt',
    'subcontract_issue',
    'subcontract_receipt',
    'transfer_out',
    'transfer_in',
    'adjustment_in',
    'adjustment_out',
    'scrap'
  ))
);
```

Rule bất biến:

- Không update `qty`, `direction`, `movement_type` sau khi tạo.
- Không delete row.
- Nếu sai, tạo movement đảo hoặc adjustment.
- Tất cả movement phải có `source_doc_type` và `source_doc_id`.
- Không tạo stock movement không rõ nguồn.

### 12.3. Stock balances

`stock_balances` là bảng đọc nhanh.

```sql
create table inventory.stock_balances (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  item_id uuid not null references mdm.items(id),
  batch_id uuid references inventory.batches(id),
  warehouse_id uuid not null references mdm.warehouses(id),
  bin_id uuid references mdm.warehouse_bins(id),
  qty_on_hand numeric(18,4) not null default 0,
  qty_reserved numeric(18,4) not null default 0,
  qty_quarantine numeric(18,4) not null default 0,
  qty_damaged numeric(18,4) not null default 0,
  qty_available numeric(18,4) generated always as (
    qty_on_hand - qty_reserved - qty_quarantine - qty_damaged
  ) stored,
  updated_at timestamptz not null default now(),
  version int not null default 1,
  constraint uq_stock_balances_key unique (org_id, item_id, batch_id, warehouse_id, bin_id),
  constraint ck_stock_balances_non_negative check (
    qty_on_hand >= 0 and qty_reserved >= 0 and qty_quarantine >= 0 and qty_damaged >= 0
  )
);
```

Rule:

- `qty_available` không nhập tay.
- Khi reserve hàng, tăng `qty_reserved`.
- Khi issue bán hàng, giảm `qty_on_hand` và giảm `qty_reserved` nếu đã reserve.
- Khi QC pass, chuyển từ quarantine sang available bằng movement/transaction đúng.

### 12.4. Stock reservations

```sql
create table inventory.stock_reservations (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  reservation_no text not null,
  source_doc_type text not null,
  source_doc_id uuid not null,
  source_doc_line_id uuid,
  item_id uuid not null references mdm.items(id),
  batch_id uuid references inventory.batches(id),
  warehouse_id uuid not null references mdm.warehouses(id),
  qty_reserved numeric(18,4) not null,
  status text not null default 'active',
  reserved_at timestamptz not null default now(),
  reserved_by uuid not null,
  released_at timestamptz,
  released_by uuid,
  release_reason text,
  constraint uq_stock_reservations_no unique (org_id, reservation_no),
  constraint ck_stock_reservations_qty check (qty_reserved > 0),
  constraint ck_stock_reservations_status check (status in ('active', 'consumed', 'released', 'expired', 'cancelled'))
);
```

### 12.5. Goods receipt

Dùng cho:

- nhập từ NCC
- nhập thành phẩm từ nhà máy gia công
- nhập hàng hoàn nếu muốn tách return receipt
- nhập điều chỉnh có kiểm soát

```sql
create table inventory.goods_receipts (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  receipt_no text not null,
  receipt_type text not null,
  source_doc_type text,
  source_doc_id uuid,
  warehouse_id uuid not null references mdm.warehouses(id),
  supplier_id uuid references mdm.suppliers(id),
  status text not null default 'draft',
  received_at timestamptz,
  received_by uuid,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_goods_receipts_no unique (org_id, receipt_no),
  constraint ck_goods_receipts_type check (receipt_type in ('purchase', 'subcontract_receipt', 'return', 'adjustment')),
  constraint ck_goods_receipts_status check (status in ('draft', 'submitted', 'qc_pending', 'partially_received', 'received', 'cancelled'))
);

create table inventory.goods_receipt_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  goods_receipt_id uuid not null references inventory.goods_receipts(id),
  line_no int not null,
  item_id uuid not null references mdm.items(id),
  batch_id uuid references inventory.batches(id),
  batch_no text,
  mfg_date date,
  expiry_date date,
  qty_expected numeric(18,4),
  qty_received numeric(18,4) not null,
  uom_id uuid not null references mdm.units(id),
  qc_required boolean not null default true,
  qc_status text not null default 'hold',
  note text,
  constraint uq_goods_receipt_lines_line unique (goods_receipt_id, line_no),
  constraint ck_goods_receipt_lines_qty check (qty_received > 0),
  constraint ck_goods_receipt_lines_qc_status check (qc_status in ('hold', 'pass', 'fail', 'not_required'))
);
```

### 12.6. Warehouse daily closing

Vì kho có kiểm kê và đối soát cuối ngày, cần bảng này.

```sql
create table inventory.warehouse_daily_closings (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  warehouse_id uuid not null references mdm.warehouses(id),
  closing_date date not null,
  shift_code text,
  status text not null default 'draft',
  total_orders_received int not null default 0,
  total_orders_packed int not null default 0,
  total_orders_handed_over int not null default 0,
  total_returns_received int not null default 0,
  total_stock_count_variances int not null default 0,
  note text,
  closed_at timestamptz,
  closed_by uuid,
  approved_at timestamptz,
  approved_by uuid,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_warehouse_daily_closing unique (org_id, warehouse_id, closing_date, shift_code),
  constraint ck_warehouse_daily_closing_status check (status in ('draft', 'submitted', 'approved', 'reopened', 'cancelled'))
);
```

---

## 13. QC schema standards

### 13.1. QC inspection

QC inspection có thể gắn với:

- goods receipt
- production/subcontract receipt
- return inspection
- batch complaint

```sql
create table qc.inspections (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  inspection_no text not null,
  inspection_type text not null,
  source_doc_type text not null,
  source_doc_id uuid not null,
  item_id uuid,
  batch_id uuid,
  status text not null default 'draft',
  result text,
  inspected_at timestamptz,
  inspected_by uuid,
  approved_at timestamptz,
  approved_by uuid,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_inspections_no unique (org_id, inspection_no),
  constraint ck_inspections_type check (inspection_type in ('inbound', 'in_process', 'finished_goods', 'return', 'complaint')),
  constraint ck_inspections_status check (status in ('draft', 'submitted', 'in_review', 'completed', 'cancelled')),
  constraint ck_inspections_result check (result is null or result in ('pass', 'fail', 'conditional_pass'))
);
```

### 13.2. QC result update rule

Khi QC pass:

```text
qc.inspections.result = pass
→ inventory.batches.qc_status = pass
→ inventory.stock_balances chuyển qty_quarantine về available logic
→ audit log
→ outbox event BatchQCPassed
```

Khi QC fail:

```text
qc.inspections.result = fail
→ inventory.batches.qc_status = fail
→ hàng không available
→ nếu cần tạo non_conformance
→ audit log
→ outbox event BatchQCFailed
```

Không cho module khác tự update `batch.qc_status` nếu không qua QC service.

---

## 14. Sales, shipping và handover schema standards

### 14.1. Sales order

```sql
create table sales.sales_orders (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  order_no text not null,
  order_date date not null default current_date,
  customer_id uuid references mdm.customers(id),
  channel text not null,
  status text not null default 'draft',
  payment_status text not null default 'unpaid',
  fulfillment_status text not null default 'unfulfilled',
  total_amount numeric(18,2) not null default 0,
  discount_amount numeric(18,2) not null default 0,
  net_amount numeric(18,2) not null default 0,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  version int not null default 1,
  constraint uq_sales_orders_no unique (org_id, order_no),
  constraint ck_sales_orders_status check (status in (
    'draft', 'confirmed', 'reserved', 'picking', 'packed', 'handed_over', 'delivered', 'closed', 'cancelled', 'returned'
  )),
  constraint ck_sales_orders_payment_status check (payment_status in ('unpaid', 'partially_paid', 'paid', 'refunded')),
  constraint ck_sales_orders_fulfillment_status check (fulfillment_status in ('unfulfilled', 'partial', 'fulfilled', 'returned'))
);
```

### 14.2. Sales order lines

```sql
create table sales.sales_order_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  sales_order_id uuid not null references sales.sales_orders(id),
  line_no int not null,
  item_id uuid not null references mdm.items(id),
  qty_ordered numeric(18,4) not null,
  qty_reserved numeric(18,4) not null default 0,
  qty_picked numeric(18,4) not null default 0,
  qty_shipped numeric(18,4) not null default 0,
  uom_id uuid not null references mdm.units(id),
  unit_price numeric(18,2) not null,
  discount_amount numeric(18,2) not null default 0,
  line_amount numeric(18,2) not null,
  constraint uq_sales_order_lines_line unique (sales_order_id, line_no),
  constraint ck_sales_order_lines_qty check (qty_ordered > 0)
);
```

### 14.3. Shipments

```sql
create table shipping.shipments (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  shipment_no text not null,
  sales_order_id uuid not null references sales.sales_orders(id),
  carrier_id uuid references mdm.carriers(id),
  tracking_no text,
  warehouse_id uuid not null references mdm.warehouses(id),
  status text not null default 'draft',
  picked_at timestamptz,
  picked_by uuid,
  packed_at timestamptz,
  packed_by uuid,
  handed_over_at timestamptz,
  handed_over_by uuid,
  delivered_at timestamptz,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_shipments_no unique (org_id, shipment_no),
  constraint ck_shipments_status check (status in (
    'draft', 'picking', 'picked', 'packing', 'packed', 'ready_for_handover', 'handed_over', 'in_transit', 'delivered', 'failed', 'returned', 'cancelled'
  ))
);
```

### 14.4. Carrier manifest

Manifest là bảng kê bàn giao cho ĐVVC.

```sql
create table shipping.carrier_manifests (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  manifest_no text not null,
  carrier_id uuid not null references mdm.carriers(id),
  warehouse_id uuid not null references mdm.warehouses(id),
  handover_zone_id uuid references mdm.warehouse_zones(id),
  manifest_date date not null default current_date,
  status text not null default 'draft',
  expected_shipment_count int not null default 0,
  scanned_shipment_count int not null default 0,
  handed_over_at timestamptz,
  handed_over_by uuid,
  carrier_receiver_name text,
  carrier_signature_ref text,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_carrier_manifests_no unique (org_id, manifest_no),
  constraint ck_carrier_manifests_status check (status in ('draft', 'scanning', 'ready', 'handed_over', 'cancelled'))
);

create table shipping.carrier_manifest_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  manifest_id uuid not null references shipping.carrier_manifests(id),
  line_no int not null,
  shipment_id uuid not null references shipping.shipments(id),
  expected_tracking_no text,
  scan_status text not null default 'pending',
  scanned_at timestamptz,
  scanned_by uuid,
  issue_code text,
  note text,
  constraint uq_carrier_manifest_lines_shipment unique (manifest_id, shipment_id),
  constraint ck_carrier_manifest_lines_scan_status check (scan_status in ('pending', 'scanned', 'missing', 'extra', 'cancelled'))
);
```

### 14.5. Scan events

```sql
create table shipping.scan_events (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  scan_type text not null,
  barcode text not null,
  shipment_id uuid,
  manifest_id uuid,
  result text not null,
  error_code text,
  scanned_at timestamptz not null default now(),
  scanned_by uuid not null,
  station_id text,
  idempotency_key text,
  metadata jsonb,
  constraint ck_scan_events_type check (scan_type in ('pick', 'pack', 'handover', 'return', 'stock_count')),
  constraint ck_scan_events_result check (result in ('success', 'duplicate', 'not_found', 'wrong_manifest', 'missing', 'error'))
);

create unique index uq_scan_events_idempotency
on shipping.scan_events(org_id, idempotency_key)
where idempotency_key is not null;
```

Rule:

- Quét trùng không tạo trạng thái sai.
- Mỗi scan phải có user và thời điểm.
- Scan sai manifest phải lưu event để audit.
- Không đủ đơn thì phải có issue record hoặc scan result tương ứng.

---

## 15. Returns schema standards

### 15.1. Return order

```sql
create table returns.return_orders (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  return_no text not null,
  source_order_id uuid references sales.sales_orders(id),
  source_shipment_id uuid references shipping.shipments(id),
  carrier_id uuid references mdm.carriers(id),
  return_type text not null,
  status text not null default 'draft',
  received_at timestamptz,
  received_by uuid,
  inspected_at timestamptz,
  inspected_by uuid,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_return_orders_no unique (org_id, return_no),
  constraint ck_return_orders_type check (return_type in ('customer_return', 'delivery_failed', 'exchange', 'internal_return')),
  constraint ck_return_orders_status check (status in ('draft', 'received', 'inspection_pending', 'inspected', 'disposed', 'closed', 'cancelled'))
);
```

### 15.2. Return lines

```sql
create table returns.return_order_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  return_order_id uuid not null references returns.return_orders(id),
  line_no int not null,
  item_id uuid not null references mdm.items(id),
  batch_id uuid references inventory.batches(id),
  qty_returned numeric(18,4) not null,
  uom_id uuid not null references mdm.units(id),
  condition_status text not null default 'unknown',
  reason_code text,
  note text,
  constraint uq_return_order_lines_line unique (return_order_id, line_no),
  constraint ck_return_order_lines_qty check (qty_returned > 0),
  constraint ck_return_order_lines_condition check (condition_status in ('unknown', 'sealed_good', 'opened_good', 'damaged', 'expired', 'suspected_fake', 'lab_required'))
);
```

### 15.3. Return disposition

```sql
create table returns.return_dispositions (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  return_order_line_id uuid not null references returns.return_order_lines(id),
  disposition text not null,
  target_warehouse_id uuid references mdm.warehouses(id),
  target_bin_id uuid references mdm.warehouse_bins(id),
  decided_at timestamptz not null default now(),
  decided_by uuid not null,
  note text,
  constraint ck_return_dispositions check (disposition in ('restock', 'quarantine', 'send_to_lab', 'scrap', 'supplier_claim'))
);
```

Rule:

- Hàng hoàn mới nhận không tự quay lại available stock.
- Phải qua inspection/disposition.
- `restock` mới tạo movement đưa về available.
- `send_to_lab`, `scrap`, `quarantine` phải đi vào kho/trạng thái riêng.

---

## 16. Subcontract manufacturing schema standards

Vì quy trình thực tế có gia công ngoài, cần thiết kế DB rõ.

### 16.1. Subcontract order

```sql
create table subcontract.subcontract_orders (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  subcontract_no text not null,
  factory_supplier_id uuid not null references mdm.suppliers(id),
  order_date date not null default current_date,
  expected_delivery_date date,
  status text not null default 'draft',
  deposit_amount numeric(18,2) not null default 0,
  final_payment_amount numeric(18,2) not null default 0,
  sample_required boolean not null default true,
  sample_status text not null default 'not_started',
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  submitted_at timestamptz,
  submitted_by uuid,
  approved_at timestamptz,
  approved_by uuid,
  constraint uq_subcontract_orders_no unique (org_id, subcontract_no),
  constraint ck_subcontract_orders_status check (status in (
    'draft', 'submitted', 'approved', 'deposit_paid', 'materials_issued', 'sample_pending', 'sample_approved', 'in_production', 'received', 'qc_pending', 'closed', 'cancelled'
  )),
  constraint ck_subcontract_sample_status check (sample_status in ('not_started', 'pending', 'approved', 'rejected', 'not_required'))
);
```

### 16.2. Subcontract order lines

```sql
create table subcontract.subcontract_order_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  subcontract_order_id uuid not null references subcontract.subcontract_orders(id),
  line_no int not null,
  finished_item_id uuid not null references mdm.items(id),
  qty_ordered numeric(18,4) not null,
  qty_received numeric(18,4) not null default 0,
  uom_id uuid not null references mdm.units(id),
  specification_note text,
  constraint uq_subcontract_order_lines_line unique (subcontract_order_id, line_no),
  constraint ck_subcontract_order_lines_qty check (qty_ordered > 0)
);
```

### 16.3. Material issue to factory

```sql
create table subcontract.material_issues (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  issue_no text not null,
  subcontract_order_id uuid not null references subcontract.subcontract_orders(id),
  source_warehouse_id uuid not null references mdm.warehouses(id),
  factory_supplier_id uuid not null references mdm.suppliers(id),
  status text not null default 'draft',
  issued_at timestamptz,
  issued_by uuid,
  receiver_name text,
  handover_doc_ref text,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  constraint uq_material_issues_no unique (org_id, issue_no),
  constraint ck_material_issues_status check (status in ('draft', 'submitted', 'approved', 'issued', 'cancelled'))
);

create table subcontract.material_issue_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  material_issue_id uuid not null references subcontract.material_issues(id),
  line_no int not null,
  item_id uuid not null references mdm.items(id),
  batch_id uuid references inventory.batches(id),
  qty_issued numeric(18,4) not null,
  uom_id uuid not null references mdm.units(id),
  constraint uq_material_issue_lines_line unique (material_issue_id, line_no),
  constraint ck_material_issue_lines_qty check (qty_issued > 0)
);
```

Khi material issue được confirm:

```text
subcontract.material_issues.status = issued
→ inventory.stock_ledger movement_type = subcontract_issue, direction = out
→ stock_balance giảm
→ audit log
```

### 16.4. Sample approval

```sql
create table subcontract.sample_approvals (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  subcontract_order_id uuid not null references subcontract.subcontract_orders(id),
  sample_no text not null,
  status text not null default 'pending',
  submitted_at timestamptz,
  reviewed_at timestamptz,
  reviewed_by uuid,
  decision_note text,
  attachment_group_id uuid,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  constraint uq_sample_approvals_no unique (org_id, sample_no),
  constraint ck_sample_approvals_status check (status in ('pending', 'approved', 'rejected', 'cancelled'))
);
```

Rule:

- Nếu `sample_required = true`, không được chuyển đơn gia công sang `in_production` khi chưa sample approved.
- Rejected sample phải lưu lý do và file/ảnh nếu có.

### 16.5. Factory claim 3–7 ngày

```sql
create table subcontract.factory_claims (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  claim_no text not null,
  subcontract_order_id uuid not null references subcontract.subcontract_orders(id),
  receipt_id uuid,
  claim_type text not null,
  status text not null default 'draft',
  description text not null,
  reported_at timestamptz,
  reported_by uuid,
  factory_response text,
  resolved_at timestamptz,
  resolved_by uuid,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  constraint uq_factory_claims_no unique (org_id, claim_no),
  constraint ck_factory_claims_type check (claim_type in ('quantity_shortage', 'quality_issue', 'wrong_spec', 'damaged_packaging', 'late_delivery', 'other')),
  constraint ck_factory_claims_status check (status in ('draft', 'submitted', 'sent_to_factory', 'in_review', 'resolved', 'rejected', 'cancelled'))
);
```

---

## 17. Purchase schema standards

### 17.1. Purchase order

```sql
create table purchase.purchase_orders (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  po_no text not null,
  supplier_id uuid not null references mdm.suppliers(id),
  order_date date not null default current_date,
  expected_delivery_date date,
  status text not null default 'draft',
  total_amount numeric(18,2) not null default 0,
  payment_status text not null default 'unpaid',
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  updated_at timestamptz not null default now(),
  updated_by uuid,
  submitted_at timestamptz,
  submitted_by uuid,
  approved_at timestamptz,
  approved_by uuid,
  constraint uq_purchase_orders_no unique (org_id, po_no),
  constraint ck_purchase_orders_status check (status in ('draft', 'submitted', 'approved', 'sent', 'partially_received', 'received', 'closed', 'cancelled')),
  constraint ck_purchase_orders_payment_status check (payment_status in ('unpaid', 'deposit_paid', 'partially_paid', 'paid'))
);
```

### 17.2. Purchase order lines

```sql
create table purchase.purchase_order_lines (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  purchase_order_id uuid not null references purchase.purchase_orders(id),
  line_no int not null,
  item_id uuid not null references mdm.items(id),
  qty_ordered numeric(18,4) not null,
  qty_received numeric(18,4) not null default 0,
  uom_id uuid not null references mdm.units(id),
  unit_price numeric(18,4) not null,
  line_amount numeric(18,2) not null,
  note text,
  constraint uq_purchase_order_lines_line unique (purchase_order_id, line_no),
  constraint ck_purchase_order_lines_qty check (qty_ordered > 0)
);
```

---

## 18. Finance basic schema standards

Phase 1 không cần full accounting engine quá sâu, nhưng phải có nền để theo dõi:

- phải thu
- phải trả
- thanh toán cọc/final payment
- đối soát COD nếu trong scope

### 18.1. Payment requests

```sql
create table finance.payment_requests (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  request_no text not null,
  request_type text not null,
  source_doc_type text not null,
  source_doc_id uuid not null,
  payee_type text not null,
  payee_id uuid,
  amount numeric(18,2) not null,
  currency text not null default 'VND',
  status text not null default 'draft',
  due_date date,
  note text,
  created_at timestamptz not null default now(),
  created_by uuid not null,
  submitted_at timestamptz,
  submitted_by uuid,
  approved_at timestamptz,
  approved_by uuid,
  paid_at timestamptz,
  paid_by uuid,
  constraint uq_payment_requests_no unique (org_id, request_no),
  constraint ck_payment_requests_type check (request_type in ('supplier_payment', 'subcontract_deposit', 'subcontract_final', 'refund', 'expense')),
  constraint ck_payment_requests_status check (status in ('draft', 'submitted', 'approved', 'paid', 'rejected', 'cancelled')),
  constraint ck_payment_requests_amount check (amount > 0)
);
```

---

## 19. Audit, outbox, idempotency, file attachment

### 19.1. Audit log

```sql
create table audit.audit_logs (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  actor_user_id uuid,
  action text not null,
  entity_type text not null,
  entity_id uuid not null,
  before_data jsonb,
  after_data jsonb,
  reason text,
  ip_address text,
  user_agent text,
  trace_id text,
  created_at timestamptz not null default now()
);

create index ix_audit_logs_entity on audit.audit_logs(entity_type, entity_id, created_at desc);
create index ix_audit_logs_actor on audit.audit_logs(actor_user_id, created_at desc);
```

Bắt buộc audit với:

- item master
- supplier/customer
- price/discount
- purchase order approval
- goods receipt
- QC pass/fail
- stock adjustment
- sales order cancel
- shipment handover
- return disposition
- subcontract material issue
- payment request/payment
- permission/role changes

### 19.2. Outbox events

```sql
create table integration.outbox_events (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  event_type text not null,
  aggregate_type text not null,
  aggregate_id uuid not null,
  payload jsonb not null,
  status text not null default 'pending',
  attempts int not null default 0,
  available_at timestamptz not null default now(),
  processed_at timestamptz,
  error_message text,
  created_at timestamptz not null default now(),
  constraint ck_outbox_events_status check (status in ('pending', 'processing', 'processed', 'failed', 'dead'))
);

create index ix_outbox_events_pending
on integration.outbox_events(status, available_at)
where status in ('pending', 'failed');
```

Rule:

- Event được insert trong cùng transaction với nghiệp vụ.
- Worker đọc outbox, publish queue hoặc xử lý async.
- Không publish event trước khi transaction commit.

### 19.3. Idempotency keys

```sql
create table core.idempotency_keys (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  idempotency_key text not null,
  request_hash text,
  endpoint text,
  response_status int,
  response_body jsonb,
  status text not null default 'processing',
  expires_at timestamptz not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint uq_idempotency_keys unique (org_id, idempotency_key),
  constraint ck_idempotency_keys_status check (status in ('processing', 'completed', 'failed'))
);
```

### 19.4. File attachments

```sql
create table file.attachments (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  entity_type text not null,
  entity_id uuid not null,
  file_name text not null,
  file_ext text,
  mime_type text,
  file_size_bytes bigint,
  storage_bucket text not null,
  storage_key text not null,
  checksum text,
  uploaded_at timestamptz not null default now(),
  uploaded_by uuid not null,
  status text not null default 'active',
  constraint ck_attachments_status check (status in ('active', 'deleted', 'quarantined'))
);

create index ix_attachments_entity on file.attachments(entity_type, entity_id);
```

---

## 20. Indexing standards

### 20.1. Bắt buộc index cho foreign key hay query thường dùng

PostgreSQL không tự tạo index cho foreign key.  
Mỗi FK quan trọng cần index nếu dùng join/filter thường xuyên.

Ví dụ:

```sql
create index ix_sales_order_lines_sales_order_id
on sales.sales_order_lines(sales_order_id);

create index ix_stock_ledger_item_batch_warehouse
on inventory.stock_ledger(item_id, batch_id, warehouse_id, movement_at desc);
```

### 20.2. Index theo status/date cho list screen

Các màn list thường lọc theo:

- status
- date range
- warehouse
- supplier/customer
- channel
- carrier

Ví dụ:

```sql
create index ix_sales_orders_status_order_date
on sales.sales_orders(org_id, status, order_date desc);

create index ix_shipments_status_carrier
on shipping.shipments(org_id, status, carrier_id, created_at desc);
```

### 20.3. Unique index cho mã nghiệp vụ

```sql
create unique index uq_sales_orders_org_order_no
on sales.sales_orders(org_id, order_no);
```

### 20.4. Partial index cho dữ liệu active/pending

```sql
create index ix_outbox_pending
on integration.outbox_events(available_at)
where status = 'pending';

create index ix_stock_reservations_active
on inventory.stock_reservations(source_doc_type, source_doc_id)
where status = 'active';
```

### 20.5. Search index

Phase 1 ưu tiên search đơn giản bằng:

- exact code
- partial text search có index nếu cần
- không triển khai full-text phức tạp ngay nếu chưa cần

Ví dụ:

```sql
create index ix_items_sku_lower
on mdm.items(lower(sku));
```

Nếu search text lớn, cân nhắc `pg_trgm`, nhưng phải có quyết định kỹ thuật riêng.

---

## 21. Transaction và locking standards

### 21.1. Giao dịch phải nằm trong transaction DB

Các nghiệp vụ sau bắt buộc transaction:

- reserve stock
- release reservation
- issue stock
- receive goods
- QC pass/fail update batch and balance
- handover shipment
- return disposition
- subcontract material issue
- payment confirmation
- stock adjustment

### 21.2. Lock balance row khi thay đổi tồn

Khi update `stock_balances`, dùng transaction và lock row:

```sql
select *
from inventory.stock_balances
where org_id = $1
  and item_id = $2
  and batch_id = $3
  and warehouse_id = $4
  and bin_id = $5
for update;
```

Nếu chưa có row thì insert với unique constraint, xử lý conflict an toàn.

### 21.3. Optimistic locking cho header chứng từ

Các bảng header có `version`.

Khi update:

```sql
update sales.sales_orders
set status = $new_status,
    version = version + 1,
    updated_at = now(),
    updated_by = $user_id
where id = $id
  and version = $expected_version;
```

Nếu affected row = 0, báo lỗi:

```text
CONFLICT_VERSION
```

### 21.4. Không khóa bảng lớn nếu không cần

Tránh:

```sql
LOCK TABLE inventory.stock_ledger;
```

Chỉ lock row liên quan.

---

## 22. State machine database rules

Database chỉ kiểm soát status value hợp lệ.  
State transition hợp lệ phải nằm ở application/domain service.

Ví dụ `sales_orders.status` có check constraint:

```text
draft, confirmed, reserved, picking, packed, handed_over, delivered, closed, cancelled, returned
```

Nhưng việc:

```text
packed → delivered
```

có hợp lệ hay không phải do state machine trong code kiểm tra.

### 22.1. Status history

Với table quan trọng nên có history.

Ví dụ:

```sql
create table sales.order_status_history (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  sales_order_id uuid not null references sales.sales_orders(id),
  from_status text,
  to_status text not null,
  changed_at timestamptz not null default now(),
  changed_by uuid not null,
  reason text
);
```

Tương tự có thể áp dụng cho:

- purchase order
- shipment
- return order
- subcontract order
- QC inspection

---

## 23. Document numbering standards

Số chứng từ không sinh bằng cách `max(no) + 1` không khóa.  
Dùng bảng sequence nghiệp vụ hoặc database sequence.

### 23.1. Document sequence table

```sql
create table core.document_sequences (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  doc_type text not null,
  prefix text not null,
  current_value bigint not null default 0,
  reset_policy text not null default 'yearly',
  updated_at timestamptz not null default now(),
  constraint uq_document_sequences unique (org_id, doc_type, prefix),
  constraint ck_document_sequences_reset check (reset_policy in ('none', 'daily', 'monthly', 'yearly'))
);
```

Example format:

```text
PO-20260424-000001
GRN-20260424-000001
SO-20260424-000001
SHP-20260424-000001
RET-20260424-000001
MAN-20260424-000001
SUB-20260424-000001
```

Rule:

- Sinh số trong transaction.
- Không tái sử dụng số chứng từ đã cancel.
- Mỗi doc type có prefix riêng.

---

## 24. Data migration/staging standards

Khi import dữ liệu cũ, không đổ thẳng vào bảng core.

Cần staging schema:

```text
staging.items_import
staging.suppliers_import
staging.customers_import
staging.opening_stock_import
staging.opening_ar_import
staging.opening_ap_import
```

Quy trình:

```text
Load raw data vào staging
→ validate required fields
→ detect duplicate
→ map code/id
→ preview error
→ business sign-off
→ import vào core tables
→ generate audit/migration log
```

### 24.1. Migration batch log

```sql
create table integration.data_migration_batches (
  id uuid primary key default gen_random_uuid(),
  org_id uuid not null,
  batch_no text not null,
  import_type text not null,
  status text not null default 'draft',
  total_rows int not null default 0,
  success_rows int not null default 0,
  error_rows int not null default 0,
  started_at timestamptz,
  completed_at timestamptz,
  created_by uuid not null,
  created_at timestamptz not null default now(),
  constraint uq_data_migration_batches_no unique (org_id, batch_no),
  constraint ck_data_migration_batches_status check (status in ('draft', 'validating', 'validated', 'importing', 'completed', 'failed', 'cancelled'))
);
```

---

## 25. Security and privacy standards

### 25.1. PII fields

Các trường có thể là PII:

- customer phone
- customer address
- employee personal data
- supplier contact
- user email/phone

Rule:

- Không log raw PII trong application logs nếu không cần.
- Audit log nên mask một số trường nhạy cảm nếu chính sách yêu cầu.
- DB access production phải giới hạn.
- Backup production phải được bảo vệ.

### 25.2. Password/auth secret

Không lưu password plain text.

```text
password_hash
password_updated_at
last_login_at
```

Dùng hashing chuẩn do backend auth quyết định.

### 25.3. Row-level security

Phase 1 có thể chưa cần PostgreSQL RLS nếu hệ thống một công ty, một backend.  
Nhưng tất cả query bắt buộc filter `org_id`.

Nếu sau này multi-tenant thật, cân nhắc RLS hoặc database separation.

---

## 26. Performance standards

### 26.1. Bảng có khả năng lớn

Các bảng sẽ tăng nhanh:

- `inventory.stock_ledger`
- `shipping.scan_events`
- `audit.audit_logs`
- `integration.outbox_events`
- `sales.sales_orders`
- `shipping.shipments`

Phase 1 chưa cần partition ngay nếu volume thấp/trung bình, nhưng phải chừa thiết kế để partition sau.

### 26.2. Partition candidate

Nếu volume cao, cân nhắc partition theo tháng/quý:

- stock_ledger by `movement_at`
- scan_events by `scanned_at`
- audit_logs by `created_at`
- outbox_events by `created_at`

### 26.3. Materialized views cho dashboard

Ví dụ:

```text
reporting.daily_sales_summary
reporting.inventory_health_summary
reporting.warehouse_daily_ops_summary
reporting.shipment_handover_summary
```

Refresh:

- theo lịch
- hoặc event-driven
- không block transaction core

---

## 27. Backup, restore, retention

### 27.1. Backup tối thiểu

- Daily full backup.
- WAL/archive nếu cần point-in-time recovery.
- Backup file storage metadata và object storage policy đi kèm.
- Test restore định kỳ.

### 27.2. Retention gợi ý

| Loại dữ liệu | Retention |
|---|---|
| Transaction ERP | Không xóa trong vòng đời hệ thống, trừ chính sách pháp lý khác |
| Audit log | Tối thiểu 2–5 năm, tùy chính sách |
| Outbox processed | Có thể archive sau 90–180 ngày |
| Scan events | Giữ tối thiểu theo nhu cầu đối soát/khiếu nại |
| File chứng từ | Theo chính sách pháp lý và vận hành |

### 27.3. Restore drill

Ít nhất mỗi quý nên test:

```text
restore DB staging từ backup
→ chạy smoke test
→ kiểm tra stock ledger/balance
→ kiểm tra sales/shipping/return
→ ký xác nhận restore OK
```

---

## 28. Migration script standards

### 28.1. File naming

```text
migrations/
  000001_create_core_schema.up.sql
  000001_create_core_schema.down.sql
  000002_create_mdm_tables.up.sql
  000002_create_mdm_tables.down.sql
  000003_create_inventory_tables.up.sql
  000003_create_inventory_tables.down.sql
```

### 28.2. Migration rule

- Không sửa migration đã chạy production.
- Muốn sửa thì tạo migration mới.
- Mỗi migration phải review bởi backend lead/DB owner.
- Migration có thay đổi lớn phải có rollback strategy.
- Không chạy destructive migration giờ cao điểm.

### 28.3. Down migration

Down migration nên có, nhưng với destructive changes production không được rollback tùy tiện.

Ví dụ:

```sql
-- Down migration chỉ dùng cho dev/staging nếu an toàn.
```

### 28.4. Seed data

Seed chỉ dùng cho:

- status lookup nếu có
- default roles/permissions
- default warehouses/zones nếu business sign-off
- test data dev/staging

Không seed dữ liệu production nghiệp vụ nếu chưa qua migration plan.

---

## 29. Database testing checklist

Mỗi module DB phải test:

### 29.1. Constraint test

- Không cho insert status sai.
- Không cho qty âm.
- Không cho duplicate code.
- Không cho expiry date trước mfg date.
- Không cho thiếu foreign key.

### 29.2. Transaction test

- Reserve stock đủ hàng.
- Reserve stock thiếu hàng → rollback.
- QC pass → batch pass + stock available đúng.
- Handover shipment → shipment/manifest/scan/audit đúng.
- Return restock → return disposition + stock movement đúng.
- Subcontract material issue → stock out đúng.

### 29.3. Concurrency test

- 2 user reserve cùng SKU cùng lúc.
- 2 user scan cùng shipment.
- 2 user approve cùng chứng từ.
- 2 worker xử lý cùng outbox event.

### 29.4. Reconciliation test

- Tổng ledger = balance.
- Sales shipped = stock issue.
- Return restock = stock movement in.
- Goods receipt pass QC = available stock.
- Handover manifest count = scanned shipment count.

---

## 30. Database Definition of Done

Một database task chỉ được xem là xong khi:

1. Table đúng schema owner.
2. Tên table/column/constraint đúng convention.
3. Có primary key UUID.
4. Có `org_id` nếu là dữ liệu nghiệp vụ.
5. Có audit columns phù hợp.
6. Có foreign key cần thiết.
7. Có unique constraint cho code/doc_no.
8. Có check constraint cho status/qty quan trọng.
9. Có index cho filter/join thường dùng.
10. Có migration up/down hoặc rollback note.
11. Có repository/query tương ứng trong Go.
12. Có test constraint/transaction chính.
13. Có update Data Dictionary nếu thêm field mới.
14. Có update OpenAPI nếu field expose ra API.
15. Có review bởi backend lead/DB owner.

---

## 31. Các anti-pattern phải cấm

### 31.1. Cấm sửa tồn trực tiếp

Không:

```sql
update inventory.stock_balances set qty_on_hand = 100 where ...;
```

Trừ khi là migration/correction có approval và có audit.

### 31.2. Cấm dùng table chung kiểu `transactions`

Không tạo một bảng `transactions` chứa mọi nghiệp vụ bằng `jsonb`.

ERP cần rõ:

- PO là PO.
- GRN là GRN.
- SO là SO.
- Shipment là shipment.
- Return là return.

### 31.3. Cấm lưu line items trong JSON

Không:

```text
sales_orders.items jsonb
```

### 31.4. Cấm status tự do

Không để user/dev tự thêm status không được chốt.

### 31.5. Cấm xóa audit/log

Audit log là dấu vết sự thật.

### 31.6. Cấm thiếu source document cho stock movement

Movement không có nguồn là movement mù.

### 31.7. Cấm nhập hàng bán khi QC chưa pass

DB và service phải hỗ trợ chặn logic này.

### 31.8. Cấm return tự động available

Hàng hoàn phải qua inspection/disposition.

### 31.9. Cấm nhận hàng gia công mà không trace subcontract order

Thành phẩm gia công phải gắn về:

- subcontract order
- factory
- receipt
- QC
- batch

---

## 32. Minimum schema pack cho Phase 1 MVP

Nếu cần build MVP gọn nhưng đúng xương sống, tối thiểu phải có:

```text
core.users
core.roles
core.permissions
core.idempotency_keys
core.document_sequences

mdm.items
mdm.units
mdm.suppliers
mdm.customers
mdm.warehouses
mdm.warehouse_zones
mdm.warehouse_bins
mdm.carriers

purchase.purchase_orders
purchase.purchase_order_lines

inventory.batches
inventory.goods_receipts
inventory.goods_receipt_lines
inventory.stock_ledger
inventory.stock_balances
inventory.stock_reservations
inventory.stock_counts
inventory.warehouse_daily_closings

qc.inspections
qc.batch_quality_statuses

sales.sales_orders
sales.sales_order_lines

shipping.shipments
shipping.carrier_manifests
shipping.carrier_manifest_lines
shipping.scan_events

returns.return_orders
returns.return_order_lines
returns.return_dispositions

subcontract.subcontract_orders
subcontract.material_issues
subcontract.material_issue_lines
subcontract.sample_approvals
subcontract.factory_claims

audit.audit_logs
file.attachments
integration.outbox_events
```

---

## 33. Implementation roadmap cho database

### Sprint DB-01: Foundation

- schemas
- users/roles/permissions
- document sequences
- idempotency
- audit logs
- outbox events
- attachments

### Sprint DB-02: Master Data

- items
- units/conversions
- suppliers
- customers
- warehouses/zones/bins
- carriers

### Sprint DB-03: Inventory Core

- batches
- stock ledger
- stock balances
- stock reservations
- goods receipt
- stock count
- warehouse daily closing

### Sprint DB-04: Purchase + QC

- purchase orders
- purchase order lines
- inspections
- batch quality status history

### Sprint DB-05: Sales + Shipping

- sales orders
- sales order lines
- shipments
- carrier manifest
- scan events

### Sprint DB-06: Returns + Subcontract

- return orders
- return lines
- return disposition
- subcontract orders
- material issue
- sample approval
- factory claim

### Sprint DB-07: Finance Basic + Reporting

- payment requests
- payments
- AR/AP basic
- COD reconciliation if in scope
- reporting summary views

---

## 34. Kết luận

Database Phase 1 phải giữ 5 nguyên tắc sống còn:

1. **Stock ledger bất biến**: mọi thay đổi tồn đều có dấu vết.
2. **Batch/QC/hạn dùng là dữ liệu lõi**: không phải ghi chú phụ.
3. **Scan/handover/return có log riêng**: để kho và ĐVVC không cãi bằng miệng.
4. **Gia công ngoài có schema riêng**: vì thực tế vận hành có nhà máy, cọc tiền, chuyển NVL/bao bì, duyệt mẫu, nhận hàng, báo lỗi.
5. **Audit/outbox/idempotency là nền kỹ thuật bắt buộc**: để hệ thống bền, không double-submit, không mất dấu vết.

Nếu Go backend là bộ não xử lý nghiệp vụ, thì PostgreSQL là bộ nhớ dài hạn của doanh nghiệp.  
Bộ nhớ này phải sạch, khóa đúng chỗ, mở đúng chỗ, và không cho ai “tô lại sự thật” sau khi giao dịch đã xảy ra.

---

## 35. Next document đề xuất

Sau tài liệu này, tài liệu kỹ thuật tiếp theo nên là:

```text
18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md
```

Mục tiêu:

- chuẩn môi trường dev/staging/prod
- Docker/Docker Compose
- CI/CD pipeline
- secrets management
- deployment strategy
- backup/restore automation
- monitoring/logging/tracing
- release checklist
- rollback playbook
