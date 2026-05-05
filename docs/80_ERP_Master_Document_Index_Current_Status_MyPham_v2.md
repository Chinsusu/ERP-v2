# 80_ERP_Master_Document_Index_Current_Status_MyPham_v2

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Current master document index and traceability map
Version: v2.3
Date: 2026-05-04
Status: Current source-of-truth index for current Phase 1 docs, design addenda, and Sprint 23 production-demand bridge

---

## 1. Purpose

This v2 index is the current navigation map for the ERP documentation set.

The original file `32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` remains the historical handoff index for the early Phase 1 documents. This file extends the map through Sprint 22 UAT pilot pack preparation after Sprint 21 auth UI backend integration and records current design addenda.

Use this file when a reader needs to answer:

```text
Which document is current?
Which sprint board or changelog proves a delivery?
Which document explains production runtime risk?
Which document locks Vietnamese operational terminology?
```

---

## 2. Current Status Snapshot

```text
Current line: Sprint 23 production planning and material-demand bridge after Sprint 22 Session 0 readiness.
Latest release tag: v0.19.0-vietnamese-ui-localization.
Sprint 21 tag status: hold; no v0.21.0-auth-ui-backend-integration-runtime-smoke tag has been created pending target staging/pilot smoke evidence.
Sprint 21 merge evidence: PR #542 merged to main at c07409cc; CI, dev deploy, full dev smoke, and auth UI browser smoke passed.
Sprint 22 status: UAT pilot pack prepared; S22-ISSUE-001 resolved by PR #546 at db894ddb; Session 0 readiness rerun passed; business UAT execution, business issue triage, Go/No-Go decision, and v0.22 tag are pending.
Sprint 23 implementation status: first runtime bridge selected in file 92 adds /production planning UI, backend production-plan API, active-formula snapshot, material demand/shortage calculation, internal Purchase Request draft lines, PostgreSQL persistence, and OpenAPI contract coverage; follow-up file 94 promotes Purchase Request submit/approve/convert-to-PO workflow; follow-up file 95 locks PO -> receiving -> QC PASS -> supplier payable traceability; follow-up file 96 locks supplier invoice and 3-way matching behavior; stock transfer, warehouse issue note, costing, payment hard-gating, and ledger-backed inventory dashboard remain pending; no v0.23 tag exists.
Release tag migration gate: PostgreSQL 16 apply + rollback passed.
Current main migration gate after Sprint 20: PostgreSQL 16 apply -> rollback -> reapply passed.
Technical contract: English.
Business display: Vietnamese-first.
Routes: English.
Locale: vi-VN.
Currency: VND.
Timezone: Asia/Ho_Chi_Minh.
Phase 1 production entrypoints: /production is planning/material-demand/PR-draft review; /subcontract remains external factory execution.
Purchase flow boundary: /production opens generated Purchase Request; PO creation belongs to approved Purchase Request conversion, not direct production-page shortcut.
Post-PO finance boundary: posted PO-linked goods receipts create supplier payable value only for QC PASS lines; supplier invoice and three-way match are locked in file 96 as separate vendor-bill evidence before payment readiness.
Internal work-center/MES production remains out of Phase 1 scope.
```

Production auth status:

```text
Web auth UI is backend-wired for the existing email/password auth surface.
Production-like deployments still require target-environment auth smoke evidence before release.
SSO, MFA, password reset email, and device/session management remain out of scope.
```

---

## 3. Recommended Read Order

For a new engineer or reviewer:

```text
1. README.md
2. 80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md
3. 38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md
4. 76_ERP_Coding_Task_Board_Sprint20_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md
5. 79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md
6. 82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md
7. 83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md
8. 84_ERP_Coding_Task_Board_Sprint22_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md
9. 85_ERP_UAT_Pilot_Pack_Sprint22_Warehouse_Sales_QC_MyPham_v1.md
10. 86_ERP_Sprint22_Changelog_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md
11. 89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md
12. 90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md
13. 91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md
14. 92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md
15. 94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md
16. 95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md
17. 96_ERP_Supplier_Invoice_Three_Way_Matching_Flow_MyPham_v1.md
18. 88_ERP_BOM_Formula_Module_Design_MyPham_v1.md
19. 78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md
20. 75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md
21. 77_ERP_Sprint19_Changelog_Vietnamese_UI_Localization_MyPham_v1.md
22. 81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md
```

For product or operations review:

```text
1. 01 Blueprint
2. 03 PRD/SRS
3. 06 To-Be process flow
4. 20 As-Is workflow
5. 21 Gap analysis and decision log
6. 75 Sprint 19 Vietnamese UI localization board
7. 81 Vietnamese operational glossary
8. 84 Sprint 22 UAT pilot task board
9. 85 Sprint 22 UAT pilot pack
10. 89 Inventory/Purchase/Warehouse document-flow design
11. 90 Sprint 23 candidate task board
12. 91 Note-sheet module roadmap and sequencing decision
13. 92 Selected Sprint 23 production planning/material demand task board
14. 94 Purchase Request workflow and PO traceability bridge
15. 95 PO, receiving, QC, supplier payable flow
16. 96 Supplier invoice and 3-way matching flow
17. 88 BOM / formula module design
18. 78 Production runtime checklist
```

---

## 4. Source Of Truth Rules

When documents conflict, use this order:

```text
1. Latest changelog or release evidence document
2. Latest sprint task board
3. Latest production/runtime checklist
4. PRD/SRS and process flow
5. API, database, security, and architecture standards
6. Historical index or planning documents
```

Technical contract rule:

```text
Do not translate API routes, DB enum values, OpenAPI schema names, permission keys, audit event codes, or backend error codes.
Translate user-facing display labels, validation copy, status labels, empty states, and operational microcopy.
```

---

## 5. Core Phase 1 Document Map

| Range | Documents | Role |
| --- | --- | --- |
| 01-02 | Blueprint and next-document roadmap | Executive direction and documentation roadmap |
| 03-10 | PRD/SRS, permissions, data dictionary, process flow, reporting, screen list, UAT, cutover | Business and delivery source documents |
| 11-19 | Backend, coding, module, UI/UX, frontend, OpenAPI, DB, DevOps, security standards | Engineering source documents |
| 20-22 | As-Is workflow, gap analysis, core docs revision log | Business reality check and decision history |
| 23-31 | Integration, QA, backlog, SOP, go-live, incident, support, governance, Phase 2 scope | Delivery, operations, and future scope |
| 32-40 | Historical master index, update pack, kickoff, vendor pack, executive summary, coding board, workspace, UI template, formatting standards | Phase 1 handoff and implementation standards |

---

## 6. Sprint Delivery Index

| Sprint | Board | Changelog / Evidence | Delivery focus |
| --- | --- | --- | --- |
| Sprint 2 | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` | `42_ERP_Sprint2_Changelog_Order_Fulfillment_Core_MyPham_v1.md` | Order fulfillment core |
| Sprint 3 | `43_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` | `44_ERP_Sprint3_Changelog_Returns_Reconciliation_Core_MyPham_v1.md` | Returns and reconciliation core |
| Sprint 4 | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` | `46_ERP_Sprint4_Changelog_Purchase_Inbound_QC_Core_MyPham_v1.md` | Purchase, receiving, inbound QC |
| Sprint 5 | `47_ERP_Coding_Task_Board_Sprint5_Subcontract_Manufacturing_Core_MyPham_v1.md` | `48_ERP_Sprint5_Changelog_Subcontract_Manufacturing_Core_MyPham_v1.md` | Subcontract manufacturing core |
| Sprint 6 | `49_ERP_Coding_Task_Board_Sprint6_Finance_Lite_COD_AR_AP_Core_MyPham_v1.md` | `50_ERP_Sprint6_Changelog_Finance_Lite_COD_AR_AP_Core_MyPham_v1.md` | Finance Lite, COD, AR, AP |
| Sprint 7 | `51_ERP_Coding_Task_Board_Sprint7_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` | `52_ERP_Sprint7_Changelog_Reporting_Inventory_Operations_Dashboard_MyPham_v1.md` | Reporting and inventory operations dashboard |
| Sprint 8 | `53_ERP_Coding_Task_Board_Sprint8_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` | `54_ERP_Sprint8_Changelog_Reporting_Hardening_Dashboard_Drilldowns_MyPham_v1.md` | Reporting hardening and drilldowns |
| Sprint 9 | `55_ERP_Coding_Task_Board_Sprint9_System_Hardening_Production_Readiness_Core_MyPham_v1.md` | `56_ERP_Sprint9_Changelog_System_Hardening_Production_Readiness_Core_MyPham_v1.md` | System hardening and production readiness core |
| Sprint 10 | `57_ERP_Coding_Task_Board_Sprint10_Persist_Operational_Runtime_Stores_MyPham_v1.md` | `58_ERP_Sprint10_Changelog_Persist_Operational_Runtime_Stores_MyPham_v1.md` | Operational runtime store persistence |
| Sprint 11 | `59_ERP_Coding_Task_Board_Sprint11_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1.md` | `60_ERP_Sprint11_Changelog_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1.md` | Inventory read model and owner documents persistence |
| Sprint 12 | `61_ERP_Coding_Task_Board_Sprint12_Batch_QC_Status_Persistence_MyPham_v1.md` | `62_ERP_Sprint12_Changelog_Batch_QC_Status_Persistence_MyPham_v1.md` | Batch QC status persistence |
| Sprint 13 | `63_ERP_Coding_Task_Board_Sprint13_End_of_Day_Reconciliation_Persistence_MyPham_v1.md` | `64_ERP_Sprint13_Changelog_End_of_Day_Reconciliation_Persistence_MyPham_v1.md` | End-of-day reconciliation persistence |
| Sprint 14 | `65_ERP_Coding_Task_Board_Sprint14_Shipping_Pick_Pack_Persistence_MyPham_v1.md` | `66_ERP_Sprint14_Changelog_Shipping_Pick_Pack_Persistence_MyPham_v1.md` | Shipping pick/pack persistence |
| Sprint 15 | `67_ERP_Coding_Task_Board_Sprint15_Finance_Runtime_Store_Persistence_MyPham_v1.md` | `68_ERP_Sprint15_Changelog_Finance_Runtime_Store_Persistence_MyPham_v1.md` | Finance runtime store persistence |
| Sprint 16 | `69_ERP_Coding_Task_Board_Sprint16_Subcontract_Runtime_Store_Persistence_MyPham_v1.md` | `70_ERP_Sprint16_Changelog_Subcontract_Runtime_Store_Persistence_MyPham_v1.md` | Subcontract runtime store persistence |
| Sprint 17 | `71_ERP_Coding_Task_Board_Sprint17_Master_Data_Runtime_Store_Persistence_MyPham_v1.md` | `72_ERP_Sprint17_Changelog_Master_Data_Runtime_Store_Persistence_MyPham_v1.md` | Master data runtime store persistence |
| Sprint 18 | `73_ERP_Coding_Task_Board_Sprint18_Auth_Session_Runtime_Store_Persistence_MyPham_v1.md` | `74_ERP_Sprint18_Changelog_Auth_Session_Runtime_Store_Persistence_MyPham_v1.md` | Auth/session runtime store persistence |
| Sprint 19 | `75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md` | `77_ERP_Sprint19_Changelog_Vietnamese_UI_Localization_MyPham_v1.md` | Vietnamese-first UI localization |
| Sprint 20 | `76_ERP_Coding_Task_Board_Sprint20_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` | `79_ERP_Sprint20_Changelog_Release_Hygiene_API_Modularization_Fallback_Cleanup_MyPham_v1.md` | Release hygiene, API modularization, production fallback hardening |
| Sprint 21 | `82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` | `83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` | Web auth UI backend integration and production runtime smoke |
| Sprint 22 | `84_ERP_Coding_Task_Board_Sprint22_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md` | `86_ERP_Sprint22_Changelog_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md` | UAT pilot pack for Warehouse, Sales, and QC |

---

## 7. Runtime And Localization Addenda

| Document | Role | When to use |
| --- | --- | --- |
| `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` | Production-like runtime checklist | Before staging, pilot, production rehearsal, production release, or production-like smoke |
| `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` | Vietnamese operational terminology lock | Before changing UI copy, status labels, validation messages, warehouse copy, returns copy, receiving copy, or QC copy |
| `82_ERP_Coding_Task_Board_Sprint21_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` | Sprint 21 auth UI integration task board | Before changing auth UI/backend session behavior |
| `83_ERP_Sprint21_Changelog_Auth_UI_Backend_Integration_Production_Runtime_Smoke_MyPham_v1.md` | Sprint 21 auth integration changelog | Before closing auth UI backend integration or recording release smoke evidence |
| `84_ERP_Coding_Task_Board_Sprint22_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md` | Sprint 22 UAT pilot task board | Before preparing or changing Warehouse/Sales/QC UAT scope |
| `85_ERP_UAT_Pilot_Pack_Sprint22_Warehouse_Sales_QC_MyPham_v1.md` | Sprint 22 UAT execution pack | During UAT setup, scenario execution, evidence capture, and Go/No-Go reporting |
| `86_ERP_Sprint22_Changelog_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md` | Sprint 22 changelog and UAT evidence register | Before closing Sprint 22 or claiming UAT execution evidence |
| `docs/uat/sprint22/` | Sprint 22 UAT templates and evidence structure | During UAT user setup, seed setup, execution logging, issue triage, and sign-off |
| `88_ERP_BOM_Formula_Module_Design_MyPham_v1.md` | BOM / formula module design | Before implementing formula master, formula import, material requirement calculation, or production/subcontract formula snapshots |
| `89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md` | Inventory, purchase request, stock transfer, warehouse issue note, and inventory dashboard design | Before implementing spreadsheet-to-ERP warehouse/purchase document flows |
| `90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md` | Follow-up Sprint 23 candidate task board | Before starting Stock Transfer, Warehouse Issue Note, and ledger-backed Inventory Dashboard implementation after the production-demand bridge is stable |
| `91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md` | Note-sheet module roadmap and sequencing decision | Before choosing the next implementation order from dashboard, setup, sales, production planning, purchasing, warehouse, and costing requests |
| `92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md` | Selected first Sprint 23 task board | Before changing production planning, formula snapshot, material demand calculation, and Purchase Request draft generation |
| `94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md` | Purchase Request workflow bridge | Before changing production-plan to Purchase Request to PO traceability, approval, or conversion behavior |
| `95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md` | PO, receiving, QC, and supplier payable flow | Before changing post-PO receiving, AP creation, PO timeline AP links, or payable traceability behavior |
| `96_ERP_Supplier_Invoice_Three_Way_Matching_Flow_MyPham_v1.md` | Supplier invoice and 3-way matching flow | Before changing supplier invoice capture, AP invoice matching, or payment-readiness traceability behavior |

---

## 8. Release Tags

| Tag | Evidence scope | Note |
| --- | --- | --- |
| `v0.15.0-finance-runtime-store-persistence` | Sprint 15 finance runtime persistence | Last explicit tag before the runtime persistence consolidation window |
| `v0.18.0-auth-session-runtime-store-persistence` | Sprint 18 auth/session runtime store persistence | Sprint 16-17 runtime persistence work was merged to main and consolidated under this release evidence |
| `v0.19.0-vietnamese-ui-localization` | Sprint 19 Vietnamese UI localization | Latest product release tag; Sprint 20 is hardening after this tag |

Sprint 21 tag hold:

```text
No v0.21.0-auth-ui-backend-integration-runtime-smoke tag has been created.
Sprint 21 has main merge, CI, dev deploy, full dev smoke, and auth UI browser smoke evidence.
Create the Sprint 21 release checkpoint tag only after target staging/pilot environment smoke evidence is recorded.
```

Sprint 22 tag hold:

```text
No v0.22.0-uat-pilot-pack-warehouse-sales-qc tag has been created.
Sprint 22 has passed Session 0 readiness after resolving S22-ISSUE-001, but business UAT has not started.
Create the Sprint 22 checkpoint tag only after business UAT evidence and Go/No-Go decision are recorded.
```

---

## 9. Current Backlog Notes

Track these as documentation or product follow-up, not as completed Sprint 20 scope:

```text
English mode still has deeper table/filter labels that need localization cleanup.
Production-like auth smoke evidence must be recorded per target environment before release.
Any future production-like release must record whether prototype fallback gaps remain.
Sprint 22 business UAT execution remains pending until business users run the prepared scripts.
Sprint 23 selected implementation track is file 92: production planning, material demand, and Purchase Request draft.
Sprint 23 follow-up Purchase Request workflow design is file 94; PR/CI/merge/dev-smoke evidence must be recorded before claiming mainline completion.
File 90 remains the follow-up track for Stock Transfer, Warehouse Issue Note, and ledger-backed Inventory Dashboard hardening.
Stock Transfer must remain same-SKU only; SKU-changing movements require separate conversion/repack design.
Purchase Request must remain separate from PO, receiving, payment, and invoice source documents.
```
