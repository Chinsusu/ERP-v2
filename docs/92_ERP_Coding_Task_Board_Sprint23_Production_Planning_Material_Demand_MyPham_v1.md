# 92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 23 - Production Planning, Material Demand, and Purchase Request Draft
Document role: Selected first implementation task board after the `note` sheet roadmap review
Version: v1
Date: 2026-05-04
Status: Proposed; selected as first implementation track; no runtime delivery claimed
Previous sprint: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC

---

## 1. Executive Summary

Sprint 23 should start with the production-planning/material-demand bridge.

The `note` sheet shows that the business workflow expects this chain:

```text
Lệnh sản xuất / kế hoạch gia công
-> Chọn thành phẩm
-> Chọn công thức/BOM
-> Nhập số lượng cần sản xuất
-> Tính nhu cầu nguyên liệu / bao bì
-> So với tồn khả dụng
-> Tạo đề nghị mua draft nếu thiếu
```

This is the missing bridge between the finished product master data work, the formula design, and the purchase/warehouse document work.

File `90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md` remains valid, but should follow this sprint because Purchase Request should be driven by controlled production demand, not by manual spreadsheet interpretation.

---

## 2. Sprint Goal

Sprint 23 goal:

```text
Let users create a production/subcontract plan for a finished product, snapshot its active formula, calculate material demand, compare shortage against available stock, and create a Purchase Request draft for missing materials.
```

The sprint succeeds when:

```text
1. A user can select a finished or semi-finished output item.
2. The system selects or validates an active formula version for that output.
3. The user enters planned output quantity and planned dates.
4. The system snapshots formula component lines into the plan.
5. The system calculates required material/packaging quantities with decimal-safe UOM handling.
6. The system compares required quantities with available stock.
7. The user can create Purchase Request draft lines for shortages.
8. No Purchase Order, receiving, payment, invoice, or stock movement is created automatically by this step.
```

---

## 3. Primary References

| Ref | Document / Source | Use |
| --- | --- | --- |
| `91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md` | Roadmap and priority decision | Primary sequencing source |
| `88_ERP_BOM_Formula_Module_Design_MyPham_v1.md` | BOM/formula design | Formula version, decimal quantity, UOM, requirement calculation |
| `89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md` | Inventory/purchase/warehouse flow design | Purchase Request and warehouse source-document boundaries |
| `90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md` | Follow-up Sprint 23 candidate | Stock transfer, warehouse issue note, inventory dashboard follow-up |
| Google Sheet gid `742547214` | `note` sheet | Business roadmap source |
| `47_ERP_Coding_Task_Board_Sprint5_Subcontract_Manufacturing_Core_MyPham_v1.md` | Subcontract manufacturing core | Phase 1 manufacturing scope |
| `70_ERP_Sprint16_Changelog_Subcontract_Runtime_Store_Persistence_MyPham_v1.md` | Subcontract persistence evidence | Runtime persistence context |
| `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` | Vietnamese glossary | UI labels and operational copy |

---

## 4. In Scope

```text
- Production/subcontract plan document
- Output item selection: finished good or semi-finished good
- Formula version selection and validation
- Formula snapshot on the plan
- Planned output quantity and planned start/end dates
- Material requirement calculation from formula batch size
- Available-stock comparison
- Shortage result table
- Purchase Request draft generation from shortage lines
- Vietnamese UI labels and validation copy
- Permissions for admin, production/subcontract, purchase, warehouse observer roles
- OpenAPI/API/client updates if runtime implementation is done
- Dev smoke for create plan, calculate demand, create PR draft, and verify no stock/PO side effects
```

---

## 5. Out Of Scope

```text
- Internal MES routing and work centers
- Labor costing and machine scheduling
- Advanced production calendar optimization
- Automatic final Purchase Order creation
- Automatic receiving, payment, or invoice creation
- Automatic stock reservation unless explicitly planned later
- Full costing / giá thành
- Stock transfer, warehouse issue note, and dashboard hardening from file 90
- SKU conversion/repack/gift split
- Multi-level approval matrix beyond the existing practical approval pattern
```

Important boundary:

```text
This sprint creates demand evidence and Purchase Request draft lines.
It must not post stock movement or create finance evidence.
```

---

## 6. Guardrails

```text
1. A production plan must have one output item.
2. The output item must be a finished good or semi-finished good.
3. A production plan must snapshot formula lines before demand is calculated.
4. Later edits to the formula master must not rewrite existing production-plan demand.
5. Requirement calculation must not use floating point.
6. User-facing quantities use Vietnamese number format.
7. API/DB decimal values keep the English technical contract with dot decimal separator.
8. Shortage can create Purchase Request draft only.
9. Purchase Request draft must not create PO, receiving, payment, invoice, or ledger movement.
10. QC-required materials received later must still follow inbound QC before available stock.
11. Technical routes, enum values, permission keys, and audit codes remain English.
```

---

## 7. Dependency Map

```text
S23P-00 Confirm production-demand boundary
  -> S23P-01 Formula readiness
  -> S23P-02 Production/subcontract plan model
  -> S23P-03 Formula snapshot
  -> S23P-04 Requirement calculation
  -> S23P-05 Shortage comparison
  -> S23P-06 Purchase Request draft
  -> S23P-07 UI, smoke, docs, changelog

S23P-04 Requirement calculation
  -> future inventory dashboard shortage drilldowns
  -> future costing

S23P-06 Purchase Request draft
  -> file 90 Purchase Request approval/PO conversion follow-up
```

---

## 8. Task Board

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
| --- | ---: | --- | --- | --- | --- |
| S23P-00-01 | P0 | BA/PM | Confirm Phase 1 production naming | UI uses `Lệnh sản xuất / Kế hoạch gia công`; internal MES remains out of scope | `91`, section 3.4 |
| S23P-00-02 | P0 | BA/PM | Confirm Purchase Request draft boundary | Business accepts that shortage creates PR draft, not PO/receiving/payment/invoice | `91`, section 3.5 |
| S23P-00-03 | P0 | BA/PM | Confirm formula source of truth | Production plan must select finished/semi item and active formula version | `88`, section 3 |
| S23P-01-01 | P0 | BE/FE | Formula active-version readiness | Finished/semi item has one usable active formula for demand calculation | `88`, section 7 |
| S23P-01-02 | P0 | BE/FE | Decimal/UOM readiness | Formula quantities support small dosage display and decimal-safe calculation | `88`, section 5 |
| S23P-02-01 | P0 | BE | Production plan domain model | Header supports output item, planned quantity, planned dates, status, formula version | `91`, section 3.4 |
| S23P-02-02 | P0 | BE | Production plan persistence/API | PostgreSQL-backed plan documents and API endpoints exist | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S23P-02-03 | P1 | BE | Production plan status flow | Draft -> calculated -> submitted -> approved/rejected/cancelled status flow with audit | `11_ERP_Backend_Architecture_Phase1_MyPham_v1.md` |
| S23P-03-01 | P0 | BE | Formula snapshot on plan | Snapshot stores component code, name, formula quantity, UOM, requirement basis | `88`, section 9 |
| S23P-03-02 | P0 | BE | Snapshot immutability test | Later formula edits do not mutate existing plan snapshot | `88`, section 9 |
| S23P-04-01 | P0 | BE | Requirement calculator | Required quantity = formula component quantity / formula batch output * planned output quantity | `88`, section 8 |
| S23P-04-02 | P0 | BE | Decimal calculation tests | Tests cover 0.000001 KG, mg/g/kg display, and no float rounding | `88`, section 5 |
| S23P-05-01 | P0 | BE | Available-stock comparison | Requirement result shows required, available, shortage, and UOM per component | `89`, section 2.2 |
| S23P-05-02 | P1 | BE | QC-held stock exclusion | Available stock excludes QC_HOLD/QC_FAIL where item requires QC | `88`, section 11 |
| S23P-06-01 | P0 | BE | Purchase Request draft generation | Shortage lines can create PR draft lines linked to production plan and component lines | `91`, section 3.5 |
| S23P-06-02 | P0 | BE | No downstream side-effect tests | Creating PR draft does not create PO, receiving, payment, invoice, or ledger movement | `89`, section 2.3 |
| S23P-06-03 | P1 | BE | PR conversion link placeholder | PR stores source production plan reference for later PO/subcontract follow-up | `89`, section 5 |
| S23P-07-01 | P1 | FE | Production plan UI | Vietnamese UI supports list/detail/create, formula selection, demand preview, shortage table, PR draft action | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
| S23P-07-02 | P1 | FE | Quantity formatting UI | UI displays small quantities as mg/g/kg and accepts Vietnamese decimal input | `88`, section 5 |
| S23P-07-03 | P1 | QA | Production demand smoke | Create plan, calculate demand, create PR draft, confirm no stock/PO side effect | This task board |
| S23P-07-04 | P0 | QA/DevOps | Required checks and dev deploy | Required CI green, dev deploy pass, production-demand smoke pass if UI/API changes are made | `README.md` |
| S23P-07-05 | P0 | PM | Sprint 23 changelog | Changelog records PRs, checks, smoke, known limits, tag hold or release status | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |

---

## 9. Demo Script

```text
1. Login as admin or production/subcontract user.
2. Open Lệnh sản xuất / Kế hoạch gia công.
3. Create a new plan.
4. Select a finished product.
5. Confirm active formula version is selected.
6. Enter planned output quantity and planned dates.
7. Save draft.
8. Calculate material demand.
9. Review formula snapshot lines.
10. Review required quantity, available quantity, shortage quantity.
11. Create Purchase Request draft for shortage lines.
12. Open Purchase Request draft and confirm source production plan link.
13. Confirm no PO exists yet.
14. Confirm no receiving record exists yet.
15. Confirm no payment/invoice record exists yet.
16. Confirm no stock ledger movement exists from the plan or PR draft.
```

Expected:

```text
Production plan produces demand evidence.
Shortage becomes Purchase Request draft.
Downstream purchase, receiving, payment, invoice, and stock movement remain controlled separate flows.
```

---

## 10. Release And Tag Rule

Recommended release checkpoint after completed implementation and verification:

```text
v0.23.0-production-planning-material-demand
```

Tag hold rule:

```text
Do not create v0.23 until code is merged, CI is green, dev deploy and smoke pass, changelog evidence is complete, and known limits are recorded.
```

---

## 11. Handoff To File 90

After this sprint is complete, file 90 becomes the next natural implementation track.

Recommended follow-up sequence:

```text
1. Purchase Request approval and conversion hardening
2. Stock Transfer
3. Warehouse Issue Note and print/export
4. Inventory dashboard ledger hardening
5. Costing design after actual cost evidence stabilizes
```
