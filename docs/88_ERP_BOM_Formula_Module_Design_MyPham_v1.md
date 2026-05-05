# 88_ERP_BOM_Formula_Module_Design_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Design source document for BOM / production formula master
Version: v1
Date: 2026-05-03
Status: Proposed design for implementation planning

---

## 1. Purpose

This document defines the Phase 1 design for the BOM / production formula module.

The current runtime system has item, SKU, UOM, supplier, customer, warehouse, subcontract order material lines, and sample `formula_version` text. It does not yet have a reusable BOM / formula master with versioned component lines and material requirement calculation.

The next implementation should add that missing source-of-truth layer:

```text
Finished item
-> active formula version
-> component requirements
-> production/subcontract material demand
-> purchase request draft
-> receiving and QC
```

The design follows the existing project contract:

```text
Technical contract: English
Business display: Vietnamese-first
Locale: vi-VN
Currency: VND
Timezone: Asia/Ho_Chi_Minh
Decimal API/DB contract: string/numeric with dot decimal separator
User-facing number display/input: Vietnamese number format
```

---

## 2. Business Context

The real formula pattern used by operations is per finished product. Each formula line stores the quantity needed to make one output unit of the finished or semi-finished item.

Example shape:

```text
Finished product: Tinh chat buoi Fast & Furious 150ml
Formula basis: 1 PCS finished product
Formula lines: 11 materials

Line:
- Material code
- Material name
- Quantity
- Unit
- Note
```

The formula may include very small material quantities:

```text
0,000001 kg = 1 mg
0,001 g = 1 mg
1 mg = 0,000001 kg
```

The system must make these quantities easy to read and safe to calculate. Users should not have to read every small dosage as `0,000001 kg`.

---

## 3. Goals

The module must support:

```text
1. Create formula versions for finished goods and semi-finished goods.
2. Define a standard batch quantity, for example 81 PCS.
3. Add component lines from existing master data items.
4. Support raw materials, fragrance, packaging, semi-finished goods, and service components.
5. Preserve the quantity and UOM entered by the user.
6. Store a canonical calculation quantity to prevent precision loss.
7. Display quantities using Vietnamese number format and readable units.
8. Allow draft formulas to contain incomplete lines during data cleanup.
9. Block activation when required lines have invalid quantity, UOM, or item references.
10. Ensure only one active formula version exists for a finished item at a time in Phase 1.
11. Snapshot the active formula into production/subcontract documents when used.
12. Calculate material requirements from per-finished-product formula quantity and planned production quantity.
```

The module must not:

```text
1. Automatically create purchase orders directly from a production order.
2. Replace purchase request, quotation, approval, or supplier selection workflow.
3. Implement internal MES routing, work centers, labor costing, or machine scheduling.
4. Implement complex R&D formulation lab features such as pH, viscosity, stability testing, or regulatory claim management.
5. Change existing backend API routes, enum values, permission keys, or audit event codes to Vietnamese.
```

---

## 4. Terminology

| Vietnamese UI label | Technical term | Meaning |
| --- | --- | --- |
| Công thức | Formula / BOM | Versioned component list used to make a finished or semi-finished item |
| Phiên bản công thức | Formula version | A controlled version of a formula |
| Thành phẩm | Finished good | Sellable or finished output item |
| Bán thành phẩm | Semi-finished good | Intermediate output that can be used in another formula |
| Nguyên liệu | Raw material | Inventory material consumed by formula |
| Hương liệu | Fragrance | Formula component, usually mass-based and small dosage |
| Bao bì | Packaging | Formula component, usually count-based |
| Dịch vụ | Service | Non-stock component, for example outsourced processing |
| Batch chuẩn | Standard batch | Formula output quantity used as the calculation base |
| Định lượng | Component dosage | Quantity of a component for the standard batch |
| Nhu cầu vật tư | Material requirement | Calculated required component quantity for a production plan/order |

---

## 5. User-Facing Number And Unit Standard

### 5.1. Vietnamese Number Format

User-facing input and display must follow Vietnamese number convention:

```text
Decimal separator: comma
Thousands separator: dot
Example display: 1.234,567891
Example input: 0,000001 kg
```

API and DB must keep the technical decimal contract:

```text
API decimal string: "0.000001"
DB numeric value: 0.000001
No thousands separator in API/DB values
```

The UI may normalize tolerant input, but save preview must show the normalized Vietnamese value before commit when the input is ambiguous.

### 5.2. Unit Display

Technical UOM codes stay uppercase:

```text
MG, G, KG, ML, L, PCS, BOTTLE, TUBE, JAR, PACK
```

Vietnamese UI should display familiar unit labels:

```text
MG -> mg
G -> g
KG -> kg
ML -> ml
L -> l
PCS -> cái / pcs depending on item context
BOTTLE -> chai
TUBE -> tuýp
JAR -> hũ
PACK -> gói
```

For formula dosage, mass units should be shown as `mg`, `g`, or `kg` according to readability.

### 5.3. Smart Mass Display

When a mass value can be displayed more clearly in another mass unit, the UI should convert it for display only.

Examples:

```text
0.000001 KG -> 1 mg
0.000010 KG -> 10 mg
0.000500 KG -> 500 mg
0.001000 KG -> 1 g
0.003000 KG -> 3 g
0.500000 KG -> 500 g
1.250000 KG -> 1,25 kg
```

The display value must not change the stored technical quantity.

### 5.4. Formula Precision Rule

Phase 1 formula quantities must support at least:

```text
Smallest business-required mass dosage: 1 mg
Equivalent in kg: 0,000001 kg
Quantity scale: 6 decimal places
```

No formula calculation should use floating point.

Implementation must use decimal string, scaled integer, or backend decimal helpers. JavaScript `Number` must not be used for formula requirement math.

### 5.5. Strict Formula Quantity Handling

For formula entry, the system should not silently round material quantities.

Recommended behavior:

```text
1. Accept Vietnamese decimal input, for example 0,000001.
2. Normalize to API decimal string, for example "0.000001".
3. If input exceeds supported precision, show a validation error.
4. Suggest changing to a smaller UOM before saving.
```

Example:

```text
Input: 0,0000005 kg
Problem: smaller than 1 mg if KG has 6-decimal precision
UI guidance: Nhap bang don vi mg hoac dieu chinh dinh luong toi thieu 1 mg.
```

If sub-milligram dosage becomes a real business requirement, add `UG` and increase formula precision in a separate design.

---

## 6. UOM Calculation Model

### 6.1. Preserve Entered Quantity

Each formula line should store both the user-entered quantity and the canonical calculation quantity.

```text
entered_qty
entered_uom_code
calc_qty
calc_uom_code
stock_base_qty
stock_base_uom_code
conversion_factor
```

Example:

```text
User enters: 1 mg

entered_qty = 1.000000
entered_uom_code = MG
calc_qty = 1.000000
calc_uom_code = MG
stock_base_qty = 0.000001
stock_base_uom_code = KG
```

### 6.2. Canonical Calculation UOM

Formula calculation should use a canonical UOM per UOM group:

| UOM group | Formula calculation UOM | Reason |
| --- | --- | --- |
| Mass | MG | Avoids tiny KG display and preserves 1 mg exactly |
| Volume | ML | Practical for liquid dosage and packaging volumes |
| Count / each | PCS | Practical for packaging and finished unit counts |
| Pack | PACK or PCS by item conversion | Depends on item definition |
| Service | SERVICE or EA | Not stock-managed |

The stock ledger can still use each item's configured base UOM. The formula module converts from formula calculation UOM to item stock base UOM when creating material requirement or issue documents.

### 6.3. Conversion Rules

Mass conversions:

```text
1 KG = 1.000 G
1 G = 1.000 MG
1 KG = 1.000.000 MG
```

Volume conversions:

```text
1 L = 1.000 ML
```

Count conversions must come from item/UOM conversion master data, not from generic assumptions.

Example:

```text
1 CARTON of SKU A may equal 48 PCS
1 CARTON of SKU B may equal 24 PCS
```

---

## 7. Logical Data Model

### 7.1. Formula Header

Recommended table concept:

```text
mdm.item_formulas
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `id` | Yes | UUID |
| `org_id` | Yes | Organization |
| `formula_code` | Yes | Stable technical code, for example `FORMULA-XFF150-V1` |
| `finished_item_id` | Yes | Parent item |
| `finished_sku` | Yes | Snapshot for readability |
| `finished_item_name` | Yes | Snapshot for readability |
| `formula_version` | Yes | Business version, for example `V1` |
| `batch_qty` | Yes | Standard output batch quantity |
| `batch_uom_code` | Yes | Usually `PCS`, `KG`, `L`, or item base UOM |
| `base_batch_qty` | Yes | Converted batch quantity |
| `base_batch_uom_code` | Yes | Finished item base UOM |
| `status` | Yes | `draft`, `active`, `inactive`, `archived` |
| `effective_from` | Optional | Activation date |
| `effective_to` | Optional | Expiry date |
| `approval_status` | Yes | `draft`, `pending_approval`, `approved`, `rejected` |
| `created_by_ref` | Yes | Actor reference |
| `approved_by_ref` | Optional | Actor reference |
| `approved_at` | Optional | Approval timestamp |
| `note` | Optional | Operational note |
| `version` | Yes | Optimistic concurrency |
| `created_at` | Yes | Audit timestamp |
| `updated_at` | Yes | Audit timestamp |

### 7.2. Formula Lines

Recommended table concept:

```text
mdm.item_formula_lines
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `id` | Yes | UUID |
| `formula_id` | Yes | Parent formula |
| `line_no` | Yes | Display order |
| `component_item_id` | Yes for active required lines | Component item |
| `component_sku` | Yes | Snapshot |
| `component_name` | Yes | Snapshot |
| `component_type` | Yes | raw material, fragrance, packaging, semi-finished, service |
| `entered_qty` | Yes | Quantity as entered by user |
| `entered_uom_code` | Yes | UOM entered by user |
| `calc_qty` | Yes | Canonical calculation quantity |
| `calc_uom_code` | Yes | `MG`, `ML`, `PCS`, etc. |
| `stock_base_qty` | Yes for stock items | Converted to item stock base UOM |
| `stock_base_uom_code` | Yes for stock items | Item stock base UOM |
| `waste_percent` | Optional | Planned loss/scrap |
| `is_required` | Yes | Required for active formula |
| `is_stock_managed` | Yes | False for service components |
| `line_status` | Yes | `active`, `excluded`, `needs_review` |
| `note` | Optional | Operational note |

### 7.3. Formula Audit Events

Recommended event concepts:

```text
formula.created
formula.updated
formula.submitted
formula.approved
formula.activated
formula.deactivated
formula.archived
formula.imported
formula.validation_failed
```

Audit metadata should include:

```text
formula_id
formula_code
finished_sku
formula_version
changed_fields
actor_ref
```

Do not put sensitive R&D details into generic logs beyond what is needed for traceability.

---

## 8. Version And Activation Rules

### 8.1. Single Active Formula

Phase 1 rule:

```text
Only one active formula version is allowed for one finished item at a time.
```

When activating a new formula version:

```text
1. Validate the new version.
2. Deactivate the previous active version.
3. Activate the new version.
4. Write audit events for both changes.
```

### 8.2. No Direct Mutation After Use

When a formula version has been used by a production/subcontract order:

```text
Do not edit active component quantities in place.
Create a new formula version instead.
```

Allowed after use:

```text
- Add operational note.
- Archive if no longer used for new orders.
- Correct display typo only if it does not change meaning or calculation.
```

Calculation-changing edits require a new version.

### 8.3. Draft Tolerance

Draft formulas may contain incomplete or questionable data during import cleanup:

```text
- Missing component item mapping
- Quantity = 0
- Unknown UOM
- Duplicate component SKU
- Inactive component
```

Active formulas may not contain these issues unless the line is explicitly marked `excluded`.

---

## 9. Validation Rules

A formula can be activated only when:

```text
1. Finished item exists and is active.
2. Finished item type is finished good or semi-finished good.
3. Batch quantity is greater than 0.
4. Batch UOM is valid and convertible to finished item base UOM.
5. Every required line has a valid component item.
6. Every required line has quantity greater than 0.
7. Every required line has a valid UOM.
8. Every stock-managed line can convert to item stock base UOM.
9. No active line references an inactive component unless explicitly approved.
10. Formula version is unique for the finished item.
11. Activation does not create overlapping active formulas.
```

Recommended warnings:

```text
- Quantity is very small and should be displayed in mg.
- Component appears more than once.
- Component has no supplier mapping.
- Component has no current stock.
- Component has no purchase UOM conversion.
- Component is service type and will not affect stock.
```

Duplicate components should be warnings, not automatic blockers, because the same material can appear in different process phases.

---

## 10. Requirement Calculation

### 10.1. Formula Scaling

The required component quantity is calculated from the per-finished-product formula quantity:

```text
required_calc_qty = formula_line_calc_qty * planned_output_qty
```

Example:

```text
Formula basis: 1 PCS
MOI_PG dosage: 3 g = 3000 mg
Planned output: 162 PCS

Required MOI_PG = 3000 mg * 162
Required MOI_PG = 486000 mg = 486 g
```

If waste percentage exists:

```text
required_with_waste = required_calc_qty * (1 + waste_percent / 100)
```

### 10.2. Material Requirement

The material requirement document should show:

```text
Component SKU
Component name
Required quantity
Available stock
Reserved stock
Incoming quantity
Shortage quantity
Suggested purchase request quantity
Suggested UOM
```

Basic shortage calculation:

```text
shortage_qty = required_qty - available_stock - incoming_qty
```

If `shortage_qty <= 0`, no purchase request line is needed.

### 10.3. Purchase Request Boundary

The system may create a purchase request draft from material requirement.

The system must not directly create purchase orders from formula calculation in Phase 1.

Correct flow:

```text
Production/subcontract demand
-> Material requirement
-> Purchase request draft
-> Approval
-> Quotation / supplier selection
-> Purchase order
```

---

## 11. UI Design

### 11.1. Navigation

Recommended first placement:

```text
Dữ liệu gốc
-> Mã hàng / SKU
-> Chi tiết thành phẩm
-> Tab: Thông tin | Công thức | Tồn kho | Lịch sử
```

Do not add a new top-level sidebar item in the first cut unless business users need a cross-product formula workbench.

### 11.2. Formula Tab

The `Công thức` tab should show:

```text
Công thức sản xuất
[+ Thêm công thức]

Product name - Batch: 81 pcs - 11 nguyên liệu - Version V1 - Đang dùng

# | Mã NVL | Tên nguyên vật liệu | Định lượng | ĐVT | Quy đổi chuẩn | Ghi chú
```

The quantity display should use readable units:

```text
1 mg
3 g
500 mg
1 chai
1 pcs
```

Technical tooltip or detail drawer may show:

```text
entered_qty / entered_uom
calc_qty / calc_uom
stock_base_qty / stock_base_uom
conversion_factor
```

### 11.3. Formula Editor

Editor fields:

```text
Finished item
Formula version
Batch quantity
Batch UOM
Effective from
Note
Lines
```

Line editor:

```text
Component SKU
Component name
Quantity
UOM
Waste %
Required?
Stock managed?
Note
```

For quantity input:

```text
Display placeholder: 0,000001
Accept typed unit: mg, g, kg where useful
Preview converted quantity before save
```

Example preview:

```text
Nhập: 0,000001 kg
Hiển thị: 1 mg
Quy đổi tồn kho: 0,000001 kg
```

### 11.4. Import Experience

Initial import should be a dry-run first:

```text
Upload/import sheet
-> Map columns
-> Normalize codes and UOM
-> Validate rows
-> Show warnings/errors
-> Save as draft formulas
-> Business review
-> Activate approved formulas
```

Expected sheet columns:

```text
Finished SKU
Finished name
Batch quantity
Batch UOM
Formula version
Line no
Component SKU
Component name
Quantity
UOM
Note
```

If existing sheets are product-specific and one formula occupies one visual block, the importer should convert that block into one formula header plus multiple lines.

---

## 12. Handling Source Data Issues

### 12.1. Zero Quantity Lines

Rows such as `0KG` should not become active required formula lines.

Recommended import behavior:

```text
Import as draft line with line_status = needs_review
or import as excluded line if business confirms it is intentionally not used.
```

Activation should block required lines with quantity = 0.

### 12.2. Unknown Components

If a component SKU is not found in master data:

```text
1. Keep the draft line.
2. Mark `needs_review`.
3. Show error: Chưa tìm thấy mã nguyên vật liệu trong dữ liệu gốc.
4. Block formula activation until mapped or excluded.
```

### 12.3. Duplicate Component Lines

Duplicates should show a warning:

```text
Mã nguyên vật liệu xuất hiện nhiều lần trong công thức.
```

Do not auto-merge duplicates during import. Business may intentionally repeat a component for separate phases.

Future enhancement can add a `process_phase` field.

---

## 13. Integration With Existing Modules

### 13.1. Master Data

Formula depends on:

```text
Products / items
UOM catalog
UOM conversions
Supplier mappings
Item status
Item type
Stock management flags
```

Formula should not duplicate item names as source of truth. Name snapshots are for traceability only.

### 13.2. Subcontract Manufacturing

Current subcontract orders already support material lines and sample `formula_version` text.

After formula master exists:

```text
1. Create subcontract/production demand for a finished item.
2. Select active formula version.
3. Generate material lines from the formula snapshot.
4. Issue materials to factory from generated lines.
5. Keep the snapshot even if the master formula changes later.
```

The existing `formula_version` sample field should eventually reference or mirror the selected formula version, but it should not be treated as the formula master itself.

### 13.3. Purchase

Formula should feed purchase through material requirement and purchase request.

```text
Formula -> requirement -> purchase request draft
```

Do not bypass:

```text
Approval
Quotation
Supplier selection
Payment terms
Invoice tracking
```

### 13.4. Inventory And QC

Formula requirement uses available stock, not physical stock.

Inbound goods still follow:

```text
Receiving
-> Inbound QC
-> Available stock only after QC pass
```

QC hold, damaged, blocked, or return-pending stock must not be counted as available for formula requirement.

### 13.5. Reporting

Formula unlocks later reports:

```text
Planned material requirement by production demand
Shortage by ingredient
Material variance actual vs formula
Formula version usage history
Cost estimate by formula
```

Cost rollup should be a later step after the formula master is stable.

---

## 14. API Contract Direction

Recommended endpoints:

```text
GET    /api/v1/formulas
POST   /api/v1/formulas
GET    /api/v1/formulas/{formula_id}
PATCH  /api/v1/formulas/{formula_id}
POST   /api/v1/formulas/{formula_id}/submit
POST   /api/v1/formulas/{formula_id}/approve
POST   /api/v1/formulas/{formula_id}/activate
POST   /api/v1/formulas/{formula_id}/archive
POST   /api/v1/formulas/{formula_id}/calculate-requirement
POST   /api/v1/formula-imports/dry-run
POST   /api/v1/formula-imports
```

API payloads must use technical English field names and decimal strings:

```json
{
  "finished_item_id": "item-xff-150",
  "formula_version": "V1",
  "batch_qty": "81.000000",
  "batch_uom_code": "PCS",
  "lines": [
    {
      "component_item_id": "item-moi-pg",
      "entered_qty": "3.000000",
      "entered_uom_code": "G",
      "note": ""
    }
  ]
}
```

Vietnamese copy belongs in frontend labels, validation messages, and operational microcopy, not in API field names.

---

## 15. Permissions

Recommended permissions:

```text
formula:view
formula:create
formula:update
formula:submit
formula:approve
formula:activate
formula:archive
formula:import
formula:calculate_requirement
```

Role direction:

| Role | Permission direction |
| --- | --- |
| ERP Admin | Full |
| Master Data Admin | Create/update/import, submit |
| R&D / Product Owner | Create/update/submit, view |
| Production Planner | View, calculate requirement |
| QC Manager | View, approve where quality-sensitive |
| Finance | View for costing |
| Warehouse | View formula-derived requirement, not formula edit |

Activation should require approval separation where practical:

```text
Creator should not be the only final approver for an active formula.
```

---

## 16. Suggested Implementation Slices

Recommended order:

```text
1. Formula domain, validation, migrations, stores.
2. API CRUD and activation workflow.
3. Shared decimal/UOM helpers for formula quantities and smart mass display.
4. Master Data UI formula tab and formula editor.
5. Requirement calculation preview.
6. Formula import dry-run and draft import.
7. Subcontract material line generation from formula snapshot.
8. Purchase request draft generation from shortage calculation.
```

Do not combine all slices into one large PR.

First implementation target should be:

```text
Formula master CRUD + activation + Vietnamese quantity/UOM UI + requirement preview.
```

Generation of purchase request drafts can follow after formula validation and calculation are proven.

---

## 17. Acceptance Criteria

The module design is implemented correctly when:

```text
1. A finished item can have draft formula versions.
2. A formula can define batch quantity such as 81 PCS.
3. Lines can contain mass components in mg, g, or kg.
4. UI displays Vietnamese number format with comma decimal separator.
5. API/DB store decimal values using dot decimal string/numeric contract.
6. 0,000001 kg can be represented and displayed as 1 mg.
7. Active formula validation blocks required zero-quantity lines.
8. Only one active formula exists per finished item in Phase 1.
9. Requirement calculation scales per-finished-product quantity to planned quantity.
10. Calculation does not use floating point.
11. Formula used by production/subcontract documents is snapshotted.
12. Purchase order is not auto-created directly from formula calculation.
```

Suggested verification examples:

```text
Input: 0,000001 kg
API normalized: "0.000001"
Display: 1 mg

Formula basis: 1 PCS
Component: 3 g
Planned output: 162 PCS
Required: 486 g

Formula basis: 1 PCS
Component: 1 mg
Planned output: 162 PCS
Required: 162 mg
```

---

## 18. Open Decisions

These decisions should be confirmed before implementation:

```text
1. Whether Phase 1 formula supports only per-finished-product quantities or also percentage-based formulas.
2. Whether services are included in formula lines now or deferred to costing.
3. Whether formula activation needs one approver or two approvers.
4. Whether component duplicate lines need a process phase field in the first cut.
5. Whether formula import source is Google Sheets export only or also Excel upload.
```

Recommended defaults:

```text
1. Per-finished-product formulas only for first implementation.
2. Allow service lines, but mark them non-stock-managed.
3. One approval role plus audit log for first implementation.
4. Warn on duplicates; defer process phase.
5. Start with CSV/Google Sheets export import profile.
```

---

## 19. Design Decision Summary

```text
Build a versioned BOM / formula master.
Use per-finished-product formulas as the first production-ready scope.
Preserve entered quantity/UOM and store canonical calculation quantity.
Use MG as formula calculation UOM for mass to avoid tiny KG precision/display problems.
Display numbers and units in Vietnamese format.
Keep API/DB decimal contract technical and English.
Block activation for invalid or zero required lines.
Snapshot formula version into production/subcontract documents.
Generate material requirement first, then purchase request draft, not purchase order.
```

This design keeps the current ERP direction intact while adding the missing formula layer needed for production planning, subcontract material issue, purchase planning, and later costing.
