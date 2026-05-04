# 91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Module roadmap and sequencing decision from the `note` Google Sheet
Version: v1
Date: 2026-05-04
Status: Proposed roadmap and implementation-order decision; no runtime delivery claimed

---

## 1. Purpose

This document converts the current `note` spreadsheet into ERP module boundaries and an implementation sequence.

Source reviewed:

```text
Google Sheet: note
URL: https://docs.google.com/spreadsheets/d/1Y2_hJy1OqsPfK5jJal3EswE2jTsGdhgJf-daoFKc8LA/edit?gid=742547214#gid=742547214
```

The sheet is useful as a business roadmap. It should not be implemented as one wide editable spreadsheet screen.

Correct ERP interpretation:

```text
Master data defines identities.
Formula/BOM defines product material structure.
Production planning calculates material demand.
Purchase Request records approved demand.
Purchase Order, receiving, payment, and invoice remain separate documents.
Stock ledger remains the quantity source of truth.
Dashboards and reports are derived from operational records.
```

---

## 2. Executive Decision

The current ERP direction is still correct, but the next implementation priority should be adjusted.

Previous candidate Sprint 23 file `90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md` focuses on:

```text
Stock Transfer
Warehouse Issue Note
Purchase Request
Inventory Dashboard
```

The `note` sheet shows a more important dependency chain:

```text
Finished product
-> Formula/BOM
-> Production plan / subcontract plan
-> Material requirement calculation
-> Purchase Request draft for shortage
-> Purchase Order / receiving / QC
-> Warehouse documents and reporting
```

Decision:

```text
Do the production-planning/material-demand slice before the broader warehouse document slice.
```

Chosen first implementation document:

```text
92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md
```

Supporting design documents:

```text
88_ERP_BOM_Formula_Module_Design_MyPham_v1.md
89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md
90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md
```

File 90 remains valid, but it should become the follow-up implementation track after the formula-to-material-demand-to-purchase-request boundary is stable.

---

## 3. Sheet-To-ERP Module Map

### 3.1. Dashboard

Sheet intent:

```text
Quick view for stock availability, slow-moving goods, low-stock warnings, near-expiry stock, top movers, inventory report, sales-channel mix, turnover speed, and return ratio.
```

ERP placement:

```text
Reporting -> Inventory dashboard
Reporting -> Operations dashboard
Inventory -> Stock availability drilldown
Sales/Returns -> Sales and return analytics
```

Rule:

```text
Dashboard numbers must be derived from stock ledger, sales/outbound records, return records, batch/expiry data, and master data.
No dashboard-only stock number should be manually edited.
```

---

### 3.2. Setup Data

Sheet intent:

```text
Supplier list
Customer list
Sales channel list
Warehouse list
Item category list
Item/SKU list
Accounting account list
Opening balance
Product formula setup
```

ERP placement:

```text
Dữ liệu gốc -> Thành phẩm
Dữ liệu gốc -> Nguyên liệu / Bao bì
Dữ liệu gốc -> Kho / Vị trí
Dữ liệu gốc -> Nhà cung cấp
Dữ liệu gốc -> Khách hàng
Future: Kênh bán hàng, tài khoản kế toán, số dư đầu kỳ
```

Rule:

```text
Formula belongs to the finished or semi-finished product, not to a free-floating material table.
Opening balance should create controlled opening stock movements, not direct manual edits to current stock.
```

---

### 3.3. Sales

Sheet intent:

```text
Sales order shaped around TikTok-style orders
Warehouse outbound order
Carrier reconciliation
Returns
Sales reports, AR by carrier/channel/customer, shipping cost, order progress, return reasons
```

ERP placement:

```text
Sales Orders
Shipping / Pick Pack Handover
Carrier manifest / reconciliation
Returns
Finance Lite AR/COD
Reporting
```

Rule:

```text
Sales outbound stock should come from pick/pack/handover or approved warehouse issue documents.
Return stock should not become available until return inspection and QC/disposition rules pass.
```

---

### 3.4. Production Planning

Sheet intent:

```text
Rename material requirement into production plan/order.
Create production orders with planned date, production days, timeline, tasks, assignee, progress, output item, output quantity, and BOM.
```

ERP placement:

```text
Production planning / subcontract planning
Formula/BOM snapshot
Material requirement calculation
Purchase Request draft generation
```

Phase 1 production scope:

```text
The operational manufacturing flow remains subcontract manufacturing.
Internal work-center/MES production remains out of scope.
```

Correct Phase 1 naming:

```text
Lệnh sản xuất / Kế hoạch gia công
```

Rule:

```text
Production planning starts from a finished product and an active formula version.
The production plan snapshots the formula so later formula edits do not rewrite historical demand.
```

---

### 3.5. Purchasing

Sheet intent:

```text
Purchase request generated from production order demand.
Tool calculates needed quantity from production quantity, product formula, and current stock.
Approved demand moves toward purchase order.
Buyer selects supplier and updates ordering/payment/invoice timeline.
Receiving completion closes the PO.
```

ERP placement:

```text
Production plan
-> Material requirement
-> Purchase Request draft
-> Purchase Request approval
-> Purchase Order
-> Goods Receiving
-> Inbound QC when required
-> AP/payment/invoice evidence
```

Critical correction:

```text
The system may create a Purchase Request draft from shortage.
It must not silently create a final Purchase Order, receiving record, payment record, or invoice.
```

Rule:

```text
Purchase Request records demand and approval.
Purchase Order records supplier commitment.
Receiving records physical stock.
Payment and invoice records finance evidence.
```

---

### 3.6. Warehouse

Sheet intent:

```text
Receiving
Outbound
Transfer
Warehouse reports
```

ERP placement:

```text
Receiving / Nhập kho
Warehouse Issue / Xuất kho
Stock Transfer / Chuyển kho
Inventory movement ledger
Inventory reporting
```

Corrections:

```text
Outbound should use "số lượng thực tế xuất", not "số lượng thực tế nhận được".
Stock transfer must move the same SKU between locations.
If SKU changes, it is conversion/repack/gift split, not transfer.
```

Rule:

```text
Any stock quantity change must come from a posted movement document.
```

---

### 3.7. Costing

Sheet intent:

```text
Giá thành
```

ERP interpretation:

```text
Costing should wait until formula, purchase actual cost, receiving, inventory valuation, and stock movement rules are stable.
```

Rule:

```text
Do not implement production costing before formula snapshots and actual purchase/receiving cost evidence exist.
```

---

## 4. Required Implementation Sequence

Recommended order:

```text
1. Complete formula/BOM foundation for finished and semi-finished products.
2. Add production/subcontract planning document that selects output item, quantity, formula version, and planned dates.
3. Snapshot formula lines into the production plan.
4. Calculate material requirements with decimal-safe UOM conversion.
5. Compare requirement against available stock.
6. Generate Purchase Request draft lines for shortages.
7. Approve Purchase Request and link downstream PO/subcontract/service follow-up.
8. Execute receiving and inbound QC for purchased stock.
9. Implement warehouse transfer/issue documents and issue-note export.
10. Harden dashboard/reporting from ledger and source documents.
11. Add costing after actual cost and stock valuation are reliable.
```

Why this order:

```text
Purchase Request from production demand needs formula and requirement calculation first.
Warehouse dashboard accuracy needs stock ledger/source documents.
Costing needs formula snapshots, actual purchase cost, receiving evidence, and inventory valuation.
```

---

## 5. Priority Decision

### First

```text
92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md
```

Reason:

```text
It creates the missing bridge from product formula to production plan to purchase demand.
Without this bridge, purchase request and warehouse documents still depend on manual interpretation of spreadsheet rows.
```

### Second

```text
90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md
```

Reason:

```text
Stock Transfer, Warehouse Issue Note, Purchase Request approval, and Inventory Dashboard remain important, but they should follow the production demand foundation.
```

### Later

```text
Costing / Giá thành
Sales-channel analytics
Advanced production scheduling
Conversion/repack/gift split
```

Reason:

```text
These depend on stable formula snapshots, source documents, actual cost evidence, and ledger-backed reporting.
```

---

## 6. Scope Boundaries

In scope for the first implementation track:

```text
Formula readiness check
Production/subcontract plan header
Output finished product and planned quantity
Formula version selection and snapshot
Material requirement calculation
Available stock comparison
Purchase Request draft generation for shortage
Vietnamese UI copy
API/OpenAPI updates if runtime work is implemented
Smoke evidence on dev
```

Out of scope for the first implementation track:

```text
Internal MES routing, work centers, labor costing, machine scheduling
Automatic final PO creation
Automatic receiving/payment/invoice creation
Full production costing
Stock conversion/repack
Advanced amount-based approval matrix
Dashboard redesign beyond requirement drilldowns
```

---

## 7. Acceptance For This Document

This document is accepted when:

```text
1. The sheet is treated as roadmap evidence, not a UI spec to copy.
2. Formula/BOM is confirmed as the prerequisite for production demand.
3. Purchase Request draft is confirmed as the output of shortage calculation.
4. PO, receiving, payment, and invoice remain separate downstream documents.
5. File 92 is selected as the first implementation task board.
6. File 90 remains the follow-up task board for stock transfer, warehouse issue note, and dashboard hardening.
```
