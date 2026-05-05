# 98_ERP_Stock_Transfer_Warehouse_Issue_Runtime_Flow_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Runtime design and implementation evidence for stock transfer and warehouse issue note
Version: v1.0
Date: 2026-05-05
Status: Runtime implementation scope for Sprint 23 inventory document follow-up

---

## 1. Purpose

This document closes the next warehouse document slice after production planning, Purchase Request, PO, receiving, QC, supplier payable, supplier invoice matching, and AP payment readiness gate.

The implemented scope adds two first-class inventory documents:

```text
Stock Transfer = Chuyen kho
Warehouse Issue Note = Phieu xuat kho
```

These documents cover internal warehouse movement and material/factory issue notes before costing and ledger-backed inventory dashboard work.

---

## 2. Business Flow

### 2.1. Stock Transfer

Use Stock Transfer when stock moves between warehouses or warehouse zones while remaining company-owned stock.

```text
Draft
-> Submit
-> Approve
-> Post
-> Inventory ledger:
   - transfer_out from source warehouse/bin
   - transfer_in to destination warehouse/bin
```

Rules:

```text
Source warehouse and destination warehouse must be different.
Quantity must be positive.
Posted transfer writes stock movement ledger.
Technical movement types stay English: transfer_out and transfer_in.
UI label is Vietnamese: Chuyen kho.
```

### 2.2. Warehouse Issue Note

Use Warehouse Issue Note when warehouse issues material, packaging, or other stock out to a destination such as subcontract factory, production preparation, lab, or manual operational issue.

```text
Draft
-> Submit
-> Approve
-> Post
-> Inventory ledger:
   - warehouse_issue from source warehouse/bin
```

Rules:

```text
Destination type and destination name are required.
Quantity must be positive.
Posted issue reduces on-hand and available stock.
Technical movement type stays English: warehouse_issue.
UI label is Vietnamese: Phieu xuat kho.
```

---

## 3. Runtime Surface

### 3.1. API

New API routes:

```text
GET  /api/v1/stock-transfers
POST /api/v1/stock-transfers
POST /api/v1/stock-transfers/{stock_transfer_id}/submit
POST /api/v1/stock-transfers/{stock_transfer_id}/approve
POST /api/v1/stock-transfers/{stock_transfer_id}/post

GET  /api/v1/warehouse-issues
POST /api/v1/warehouse-issues
POST /api/v1/warehouse-issues/{warehouse_issue_id}/submit
POST /api/v1/warehouse-issues/{warehouse_issue_id}/approve
POST /api/v1/warehouse-issues/{warehouse_issue_id}/post
```

Permission behavior:

```text
List requires inventory:view.
Create and lifecycle actions require record:create.
Routes, enum values, movement types, OpenAPI schemas, and audit actions remain English technical contracts.
```

### 3.2. Database

New migration:

```text
000042_create_inventory_transfer_issue_documents
```

Tables:

```text
inventory.stock_transfer_documents
inventory.warehouse_issue_documents
```

Storage approach:

```text
Header columns support listing/filtering/status checks.
JSONB payload preserves complete document structure.
Stock movement ledger remains the source for posted inventory balance effects.
```

### 3.3. UI

Inventory page adds two operational cards:

```text
Chuyen kho
Phieu xuat kho
```

Both cards use the selected available-stock row as the source line and support:

```text
Create draft
Submit
Approve
Post
List existing documents
Vietnamese status labels
vi-VN quantity display
```

---

## 4. Guardrails

Do not treat these documents as costing documents.

```text
Stock Transfer moves stock ownership internally.
Warehouse Issue reduces stock availability for an operational destination.
Costing, landed cost, manufacturing variance, and financial ledger posting remain follow-up scope.
```

Do not bypass QC/receiving/payable rules:

```text
PO-linked receiving and QC PASS payable rules remain governed by file 95.
Supplier invoice matching remains governed by file 96.
AP payment readiness remains governed by file 97.
```

Do not translate technical contracts:

```text
API paths: English
DB tables/columns: English
Movement types: English
Audit actions: English
UI labels: Vietnamese-first
```

---

## 5. Follow-Up Scope

Recommended next work:

```text
1. Link Warehouse Issue Note directly from production-plan material shortage rows.
2. Add document detail pages/timelines for stock transfer and issue note if operators need richer traceability.
3. Add costing policy for issued material and subcontract cost allocation.
4. Build ledger-backed inventory dashboard from posted movements and balances.
5. Add receipt confirmation for stock transfers if two-step transfer is required by operations.
```

Out of scope for this slice:

```text
Costing
MRP
Internal work-center/MES execution
Two-step in-transit warehouse accounting
Finance ledger posting
```
