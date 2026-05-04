# 89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Design source document for inventory, purchase request, stock transfer, warehouse issue note, and inventory dashboard alignment
Version: v1
Date: 2026-05-04
Status: Proposed design for implementation planning

---

## 1. Purpose

This document converts the current spreadsheet-based warehouse and purchasing workflow into ERP design decisions.

Source spreadsheets reviewed:

```text
Tổng quan: Google Sheet gid 471150862
Quản Lý Kho: Google Sheet gid 2123530252
Đề Nghị Mua Hàng: Google Sheet gid 1060593906
Nhập Kho: Google Sheet gid 2139731924
Xuất Kho: Google Sheet gid 1438559352
Chuyển Kho: Google Sheet gid 259974100
Phiếu Xuất Kho: Google Sheet gid 1174387056
```

The design rule is:

```text
Do not rebuild the spreadsheet as one wide editable table.
Use the spreadsheet as workflow evidence, then split it into ERP master data, business documents, stock movement ledger, and reports.
```

This keeps the current ERP direction intact:

```text
Master data is the identity source of truth.
Stock movement ledger is the quantity source of truth.
Dashboard/reporting is derived, not manually edited.
Purchase request and subcontract/purchase documents are separate controlled documents.
Warehouse issue notes are printable/exportable evidence generated from outbound flows.
```

---

## 2. Source Sheet Observations

### 2.1. Tổng quan

The overview tab mixes:

```text
low-stock warning products
date range filters
sales channel selection
top moving products
inventory in/out summary
product category summary
daily movement figures
```

ERP interpretation:

```text
This is an operations dashboard/report.
It should not be the source of stock quantity.
```

Correct ERP placement:

```text
Reporting -> Inventory dashboard
Reporting -> Operations daily dashboard
Inventory -> Stock availability / movement drilldown
```

---

### 2.2. Quản Lý Kho

The warehouse management tab contains:

```text
SKU
Tên mặt hàng
Loại
ĐVT
Tình trạng bán
Định mức / warning threshold
Tồn hiện tại
Trạng thái tồn: Còn hàng / Hết hàng / Số lượng ít
Đơn giá vốn
Tổng vốn tồn
```

ERP interpretation:

```text
SKU, name, type, UOM, sale status belong to master data.
Current stock and stock status are derived from inventory ledger/read model.
Cost and inventory value belong to inventory/finance reporting.
```

Correct ERP placement:

```text
Dữ liệu gốc -> Thành phẩm / Nguyên liệu - Bao bì
Kho hàng -> Tồn khả dụng
Báo cáo -> Inventory snapshot / Inventory valuation
```

Important rule:

```text
Do not let users manually edit "Tồn hiện tại" in this report.
Any stock change must come from a movement document.
```

---

### 2.3. Đề Nghị Mua Hàng

The purchase request tab currently combines many concepts:

```text
Mã đề nghị
Ngày đề nghị
Nội dung đề nghị
SKU/material/service code
Supplier
Requested quantity
Received quantity
Unit price
VAT
Total
Approval status
Planned payment date
Ordering status
Receiving status
Payment status
Payment amount
Invoice status
Invoice date / number
```

It also contains subcontract manufacturing/service rows, for example:

```text
Chi phí khuấy trộn
Chi phí đóng gói
BTP_XB
BTP_SN
BTP_BT
Gia công xịt bưởi
Gia công sữa ủ
Gia công bột tắm
```

ERP interpretation:

```text
This sheet is not one ERP document.
It is a cross-functional tracker spanning request, approval, order, receiving, payable, payment, invoice, and subcontract service tracking.
```

Correct ERP split:

```text
Purchase Request / Đề nghị mua
-> Purchase Order / Đơn mua
-> Goods Receiving / Nhập kho
-> Inbound QC if stock item requires QC
-> Supplier Payable / Công nợ phải trả
-> Payment / Thanh toán
-> Supplier Invoice / Hóa đơn NCC
```

Subcontract rows should flow through:

```text
Subcontract order
-> factory confirmation
-> deposit / payment milestone
-> material issue to factory if applicable
-> finished goods receipt
-> inbound QC
-> final payable/payment readiness
```

Important rule:

```text
Purchase Request may create a draft PO or draft subcontract demand after approval.
It must not silently become a received stock record or paid invoice.
```

---

### 2.4. Nhập Kho

The receiving tab contains:

```text
date
supplier or source warehouse
destination warehouse
SKU
product name
category
quantity
unit cost
total cost
receiving reason / opening balance
```

ERP interpretation:

```text
This is an inbound stock movement document and ledger source.
```

Correct ERP placement:

```text
Nhập kho / Receiving
Inventory movement ledger
Inbound QC when QC-required item is received
```

Important rule:

```text
Receiving does not always mean available stock.
Only QC-passed stock can become available when the item requires QC.
```

---

### 2.5. Xuất Kho

The outbound tab contains:

```text
date
warehouse
sales channel
SKU
product name
category
quantity
```

ERP interpretation:

```text
This is an outbound stock movement ledger source.
For sales orders it should be generated from pick/pack/handover flow.
For non-sales reasons it should come from a warehouse issue document.
```

Correct ERP placement:

```text
Shipping / Pick Pack Handover for sales outbound
Inventory -> Warehouse issue for non-sales outbound
Inventory movement ledger
```

Important rule:

```text
Stock out must reference a reason/source document:
sales_order, carrier_manifest, sample_issue, damage_writeoff, internal_use, subcontract_issue, adjustment, or manual_warehouse_issue.
```

---

### 2.6. Chuyển Kho

The stock transfer tab contains:

```text
date
source warehouse
destination warehouse
SKU
product name
category
quantity
note/specification
```

Some rows represent paired transformation-like movements between product and gift/sample SKUs, for example a product SKU moving out and a related gift SKU moving in.

ERP interpretation:

```text
Normal warehouse transfer is a two-sided movement between source and destination locations for the same SKU.
Product conversion/repack/gift split is not a simple transfer and should require a separate conversion/repack document or be treated as a production/subcontract/formula output later.
```

Correct ERP placement:

```text
Inventory -> Stock transfer
Inventory -> Stock conversion/repack backlog item if SKU changes
```

Important rule:

```text
A stock transfer must not change SKU identity.
If SKU changes, use a conversion/repack document with consume/output lines and audit.
```

---

### 2.7. Phiếu Xuất Kho

The issue note tab is a printable document containing:

```text
company identity
date
warehouse issue note number
destination/channel, for example SHOPEE
product name
SKU
category
quantity
specification
```

ERP interpretation:

```text
This is not a primary data-entry table.
It is printable/exportable evidence generated from outbound issue lines.
```

Correct ERP placement:

```text
Inventory -> Warehouse issue note
Shipping -> Manifest/fulfillment issue note export
PDF/print/export from source document
```

Important rule:

```text
The printed issue note must be generated from posted or approved outbound lines.
Users should not re-enter the same SKU quantities in a separate print-only sheet.
```

---

## 3. Design Decision Summary

```text
Tổng quan = report/dashboard.
Quản Lý Kho = inventory snapshot/read model plus valuation report.
Nhập Kho = inbound stock movement document.
Xuất Kho = outbound stock movement document or generated from fulfillment.
Chuyển Kho = stock transfer document; same SKU only.
Phiếu Xuất Kho = printable/exportable warehouse issue note.
Đề Nghị Mua Hàng = purchase request workflow, split from PO/receiving/payment/invoice/subcontract.
```

Primary source-of-truth hierarchy:

```text
SKU/item identity: master data
Warehouse/location identity: master data
Supplier/customer identity: master data
Inventory quantity: stock movement ledger
Inventory current balance: inventory read model derived from ledger
Purchase request state: purchase request document
PO/receiving/QC/payment/invoice state: their own documents
Dashboard values: reports derived from documents and ledger
```

---

## 4. Target Information Architecture

Recommended Phase 1 navigation:

```text
Dữ liệu gốc
  - Thành phẩm
  - Nguyên liệu / Bao bì
  - Kho / Vị trí
  - Nhà cung cấp
  - Khách hàng

Kho hàng
  - Tồn khả dụng
  - Nhập kho
  - Xuất kho
  - Chuyển kho
  - Điều chỉnh kho
  - Kiểm kê

Mua hàng
  - Đề nghị mua
  - Đơn mua
  - Theo dõi nhận hàng

Nhập kho
  - Phiếu nhập
  - Nhập theo PO
  - Nhập nội bộ / chuyển kho đến

Kiểm chất lượng
  - QC hàng nhập
  - QC hold / quarantine

Giao hàng
  - Picking
  - Packing
  - Bàn giao ĐVVC
  - Phiếu xuất kho / export

Báo cáo
  - Tổng quan tồn kho
  - Inventory snapshot
  - Inventory movement
  - Purchase request tracker
```

Do not add duplicate sidebar entries if one module already owns the flow. For example, "Phiếu xuất kho" can be an action/export inside Shipping or Inventory Outbound before becoming a standalone module.

---

## 5. Purchase Request Design

### 5.1. Purpose

Purchase Request records internal demand before supplier commitment.

It answers:

```text
What is requested?
Why is it requested?
Who requested it?
Which department/function needs it?
Who approved it?
Should it become a PO, subcontract demand, or non-stock service expense?
```

### 5.2. Header Fields

Recommended fields:

| Field | Required | Notes |
| --- | --- | --- |
| `request_no` | Yes | Business code, for example `MDN230725-05` |
| `request_date` | Yes | Date in Asia/Ho_Chi_Minh |
| `request_type` | Yes | `purchase`, `subcontract_service`, `non_stock_service`, `inventory_replenishment` |
| `requester_ref` | Yes | User/department reference |
| `department_code` | Optional | Warehouse, Production, QC, Sales, Admin |
| `purpose` | Yes | Operational reason |
| `needed_by_date` | Optional | Requested delivery date |
| `supplier_id` | Optional | Suggested supplier |
| `status` | Yes | Draft/submitted/approved/rejected/converted/cancelled |
| `approval_status` | Yes | Draft/pending/approved/rejected |
| `currency` | Yes | VND for Phase 1 |
| `note` | Optional | Request note |

### 5.3. Line Fields

Recommended fields:

| Field | Required | Notes |
| --- | --- | --- |
| `line_no` | Yes | Display order |
| `item_id` | Required for stock lines | Material, packaging, finished, semi-finished, service |
| `sku_code` | Yes | Snapshot |
| `item_name` | Yes | Snapshot |
| `line_type` | Yes | stock_item, service, subcontract_processing, packaging_service |
| `requested_qty` | Yes | Decimal string |
| `requested_uom_code` | Yes | Existing UOM catalog |
| `estimated_unit_price` | Optional | VND decimal string |
| `vat_rate` | Optional | Decimal string |
| `estimated_line_total` | Derived | Display only |
| `received_qty` | Derived | From receiving/subcontract receipts |
| `conversion_status` | Derived | Not converted/partial/converted |
| `note` | Optional | Line note |

### 5.4. Status Model

Recommended status flow:

```text
draft
-> submitted
-> approved
-> converted
```

Exception paths:

```text
submitted -> rejected
draft/submitted/approved -> cancelled
approved -> partially_converted
```

Do not combine request status with:

```text
ordering status
receiving status
payment status
invoice status
```

Those belong to linked downstream documents.

### 5.5. Conversion Rules

After approval:

```text
stock_item lines -> purchase order draft
subcontract_processing / packaging_service lines -> subcontract order or subcontract payment milestone draft
non_stock_service lines -> supplier payable draft or service purchase order draft
```

The conversion must create traceability:

```text
purchase_request_id
purchase_request_line_id
target_document_type
target_document_id
converted_qty
converted_at
converted_by
```

---

## 6. Stock Transfer Design

### 6.1. Purpose

Stock Transfer moves the same SKU/batch from one warehouse/location to another.

It answers:

```text
What moved?
From where?
To where?
How much?
Who requested, picked, shipped, received, and posted it?
```

### 6.2. Guardrail

```text
Stock transfer must not change SKU.
If SKU changes, use conversion/repack, not transfer.
```

Examples:

```text
Valid transfer:
SKU XBN moves from WH-MAIN/A1 to WH-SHOPEE/STAGE.

Invalid transfer:
SKU XBN moves out and SKU XB100 moves in.
This is SKU conversion/repack/gift split and needs a conversion document.
```

### 6.3. Header Fields

| Field | Required | Notes |
| --- | --- | --- |
| `transfer_no` | Yes | Business code |
| `transfer_date` | Yes | Date |
| `source_warehouse_id` | Yes | Source warehouse |
| `destination_warehouse_id` | Yes | Destination warehouse |
| `source_location_id` | Optional | Source bin/location |
| `destination_location_id` | Optional | Destination bin/location |
| `reason_code` | Yes | replenishment, staging, return_to_main, quarantine_move, branch_transfer |
| `status` | Yes | Draft/submitted/approved/picked/in_transit/received/posted/cancelled |
| `requested_by` | Yes | Actor |
| `approved_by` | Optional | Actor |
| `note` | Optional | Note |

### 6.4. Line Fields

| Field | Required | Notes |
| --- | --- | --- |
| `line_no` | Yes | Display order |
| `item_id` | Yes | Same item for source and destination |
| `sku_code` | Yes | Snapshot |
| `batch_id` | Optional | Required for batch-controlled item |
| `requested_qty` | Yes | Quantity requested |
| `picked_qty` | Optional | Quantity picked |
| `received_qty` | Optional | Quantity received |
| `uom_code` | Yes | Movement UOM |
| `stock_status` | Yes | Available/QC hold/damaged/etc. |
| `note` | Optional | Line note |

### 6.5. Posting Rules

Recommended posting:

```text
When transfer is picked:
  create source reservation or pending transfer movement.

When transfer is posted/received:
  create OUT movement from source.
  create IN movement to destination.
  keep same SKU, batch, stock status unless explicitly moved into quarantine/hold reason.
```

If partial receive is allowed:

```text
remaining_qty = picked_qty - received_qty
exception status = short_received
```

---

## 7. Warehouse Issue / Stock Out Design

### 7.1. Purpose

Warehouse Issue is the controlled document for non-sales stock out and printable issue evidence.

It covers:

```text
sample issue
internal use
damage write-off
marketing gift
manual channel issue
subcontract material issue
other approved outbound
```

Sales fulfillment should usually flow through:

```text
Sales order -> pick -> pack -> carrier manifest -> handover
```

Warehouse Issue should not duplicate that flow unless it is only exporting/printing the issue note from it.

### 7.2. Header Fields

| Field | Required | Notes |
| --- | --- | --- |
| `issue_no` | Yes | Printed document number |
| `issue_date` | Yes | Date |
| `warehouse_id` | Yes | Source warehouse |
| `destination_type` | Yes | channel, customer, internal_department, factory, disposal, other |
| `destination_name` | Yes | Example: SHOPEE |
| `reason_code` | Yes | sales_handover, sample, internal_use, damage, subcontract_issue, manual_adjusted_issue |
| `status` | Yes | Draft/submitted/approved/picked/posted/cancelled |
| `created_by` | Yes | Actor |
| `approved_by` | Optional | Actor |
| `posted_by` | Optional | Actor |
| `note` | Optional | Note |

### 7.3. Line Fields

| Field | Required | Notes |
| --- | --- | --- |
| `line_no` | Yes | Display order |
| `item_id` | Yes | Item reference |
| `sku_code` | Yes | Snapshot |
| `item_name` | Yes | Snapshot |
| `category` | Optional | Product category snapshot |
| `batch_id` | Optional | Required for batch-controlled item |
| `qty` | Yes | Decimal string |
| `uom_code` | Yes | UOM |
| `specification` | Optional | Printed specification |
| `source_document_type` | Optional | sales_order, carrier_manifest, subcontract_order |
| `source_document_id` | Optional | Linked source |

### 7.4. Print / Export Requirements

The printed issue note should show:

```text
Company name
Tax code
Address
Issue note number
Issue date
Destination/channel
Line number
Product name
SKU
Category
Quantity
Specification
Prepared by
Warehouse issuer
Receiver
Approval/signature area
```

Important rule:

```text
Print/export must read from approved/posted document lines.
It must not create a second editable quantity source.
```

---

## 8. Inventory Dashboard / Report Design

### 8.1. Purpose

The dashboard replaces the spreadsheet overview and warehouse status summary.

It answers:

```text
Which products are low?
Which products are out?
Which products moved fastest?
How much stock value is held?
Which channels consumed stock?
What changed in the selected period?
```

### 8.2. Inputs

Dashboard must derive from:

```text
master data items
stock movement ledger
available stock read model
sales/channel outbound movements
purchase/receiving inbound movements
transfer movements
adjustment movements
standard cost / valuation source
```

### 8.3. KPIs

Recommended KPIs:

```text
Total active SKUs
In stock SKUs
Out-of-stock SKUs
Low-stock SKUs
Inventory value
Inbound quantity in period
Outbound quantity in period
Net movement in period
Top moving SKUs
Slow/no movement SKUs
Low stock by category/channel
```

### 8.4. Drilldowns

Each dashboard number must drill down to source records:

```text
Low-stock SKU -> stock availability row + movement history
Inbound quantity -> receiving/transfer-in/adjustment-in movements
Outbound quantity -> issue/shipping/transfer-out/adjustment-out movements
Inventory value -> SKU stock and unit cost breakdown
Top moving SKU -> outbound movement lines
```

Important rule:

```text
Every dashboard number should be explainable from ledger rows.
No manual dashboard-only quantity.
```

---

## 9. Ledger And Movement Taxonomy

Recommended movement types:

```text
opening_balance
purchase_receipt
inbound_qc_release
sales_issue
carrier_handover
warehouse_issue
stock_transfer_out
stock_transfer_in
subcontract_material_issue
subcontract_finished_receipt
return_receipt
return_disposition_available
return_disposition_scrap
stock_adjustment_in
stock_adjustment_out
conversion_consume
conversion_output
```

Recommended source document fields:

```text
source_type
source_id
source_no
source_line_id
source_line_no
```

Recommended invariant:

```text
Every stock balance change must have a ledger movement.
Every ledger movement must have a source document or approved adjustment reason.
```

---

## 10. Integration With Existing Modules

### 10.1. Master Data

Depends on:

```text
items/SKU
item type
UOM
warehouse/location
supplier/customer
sellable/purchasable/producible flags
QC required flag
lot/expiry control flags
```

### 10.2. Purchase / Receiving / QC

Flow:

```text
Purchase Request
-> Purchase Order
-> Receiving
-> Inbound QC if required
-> Available stock movement only after QC pass
-> Supplier payable/invoice/payment tracking
```

### 10.3. Shipping / Sales

Flow:

```text
Sales Order
-> Reserve
-> Pick
-> Pack
-> Carrier Manifest
-> Handover
-> Warehouse issue note export if business needs printable evidence
```

### 10.4. Subcontract

Flow:

```text
Purchase Request service/subcontract demand if needed
-> Subcontract Order
-> Payment milestone/deposit
-> Material issue to factory
-> Finished goods receipt
-> Inbound QC
-> Final payment readiness
```

### 10.5. BOM / Formula

After formula master exists:

```text
Formula requirement
-> shortage calculation
-> purchase request draft
-> approved purchase/subcontract flow
```

Do not generate purchase orders directly from formula calculation.

---

## 11. Gap List

### Gap 1: Purchase Request is not a first-class workflow

Current risk:

```text
The spreadsheet combines approval, order, receiving, payment, invoice, and subcontract status.
```

Needed:

```text
First-class Purchase Request with approval and conversion links.
```

### Gap 2: Stock Transfer needs explicit same-SKU document

Current risk:

```text
Sheet transfer rows may hide conversion/repack movements.
```

Needed:

```text
Stock Transfer document for same-SKU moves.
Separate Stock Conversion/Repack backlog item for SKU-changing movements.
```

### Gap 3: Warehouse Issue Note needs generated print/export

Current risk:

```text
Manual issue note can diverge from stock-out ledger.
```

Needed:

```text
Printable/exportable issue note generated from outbound source lines.
```

### Gap 4: Dashboard must be ledger-backed

Current risk:

```text
Overview values may become manually curated and hard to reconcile.
```

Needed:

```text
Inventory dashboard backed by movement ledger and stock read model with drilldowns.
```

---

## 12. Suggested Implementation Order

Recommended sequence:

```text
1. Stock movement taxonomy and source document invariants.
2. Stock Transfer document and posting flow.
3. Warehouse Issue document and issue note print/export.
4. Purchase Request workflow with approval and conversion links.
5. Inventory dashboard/report drilldowns from ledger.
6. Stock Conversion/Repack design if SKU-changing transfer rows are confirmed.
```

Why this order:

```text
Stock transfer and issue note tighten warehouse correctness first.
Purchase request then separates request/approval from downstream PO/receiving/payment.
Dashboard comes after source documents are clean enough to explain its numbers.
Conversion/repack is deliberately separate because it changes SKU identity and affects costing/audit.
```

---

## 13. Acceptance Criteria For This Design

The design is ready for implementation planning when these decisions are accepted:

```text
1. Tổng quan is treated as derived dashboard/report.
2. Quản Lý Kho current stock is not manually edited.
3. Purchase Request is split from PO, receiving, payment, and invoice.
4. Subcontract/service request lines are routed to subcontract or service payable flows.
5. Stock Transfer is same-SKU only.
6. SKU-changing movement is treated as conversion/repack, not transfer.
7. Warehouse Issue Note is generated from outbound document lines.
8. Every stock quantity change goes through ledger movement.
9. Dashboard values must drill down to source documents or ledger rows.
```

---

## 14. Open Decisions

Before implementation, confirm:

```text
1. Whether stock transfer requires approval before pick/post.
2. Whether partial transfer receive is allowed in Phase 1.
3. Whether warehouse issue note print is needed from sales manifest only, manual warehouse issue only, or both.
4. Whether Purchase Request approval is one-level or amount-based.
5. Whether supplier invoice tracking stays in Finance Lite or appears as a status summary on Purchase Request.
6. Whether SKU-changing rows in the transfer sheet represent real repack/conversion or spreadsheet workaround.
```

Recommended defaults:

```text
1. Require approval for stock transfer that leaves main warehouse or affects QC/defect stock.
2. Allow partial transfer receive but record exception.
3. Support print/export from both manual warehouse issue and carrier manifest.
4. Start with one-level approval, keep amount-based approval as backlog.
5. Keep invoice/payment source in Finance/Purchase documents, show linked status only.
6. Treat SKU-changing transfer as conversion/repack backlog until business confirms exact rule.
```

---

## 15. One-Line Rule

```text
The spreadsheet can show the business view, but ERP stock truth must come from controlled documents and ledger movements.
```
