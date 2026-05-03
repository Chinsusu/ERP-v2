# 80_ERP_Master_Document_Index_Current_Status_MyPham_v2

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Current master document index and traceability map
Version: v2.0
Date: 2026-05-03
Status: Current source-of-truth index for docs 01-81

---

## 1. Purpose

This v2 index is the current navigation map for the ERP documentation set.

The original file `32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` remains the historical handoff index for the early Phase 1 documents. This file extends the map through Sprint 20 hardening and the post-localization documentation cleanup.

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
Current main: Sprint 20 hardening completed after Sprint 19 Vietnamese UI localization.
Latest release tag: v0.19.0-vietnamese-ui-localization.
Release tag migration gate: PostgreSQL 16 apply + rollback passed.
Current main migration gate after Sprint 20: PostgreSQL 16 apply -> rollback -> reapply passed.
Technical contract: English.
Business display: Vietnamese-first.
Routes: English.
Locale: vi-VN.
Currency: VND.
Timezone: Asia/Ho_Chi_Minh.
```

Production caveat:

```text
Web auth UI is still mock/staging-only until wired to the backend auth/session API.
Backend auth/session persistence exists, but frontend login must not be called production-ready until that integration is explicit.
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
6. 78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md
7. 75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md
8. 77_ERP_Sprint19_Changelog_Vietnamese_UI_Localization_MyPham_v1.md
9. 81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md
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
8. 78 Production runtime checklist
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

---

## 7. Runtime And Localization Addenda

| Document | Role | When to use |
| --- | --- | --- |
| `78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md` | Production-like runtime checklist | Before staging, pilot, production rehearsal, production release, or production-like smoke |
| `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` | Vietnamese operational terminology lock | Before changing UI copy, status labels, validation messages, warehouse copy, returns copy, receiving copy, or QC copy |

---

## 8. Release Tags

| Tag | Evidence scope | Note |
| --- | --- | --- |
| `v0.15.0-finance-runtime-store-persistence` | Sprint 15 finance runtime persistence | Last explicit tag before the runtime persistence consolidation window |
| `v0.18.0-auth-session-runtime-store-persistence` | Sprint 18 auth/session runtime store persistence | Sprint 16-17 runtime persistence work was merged to main and consolidated under this release evidence |
| `v0.19.0-vietnamese-ui-localization` | Sprint 19 Vietnamese UI localization | Latest product release tag; Sprint 20 is hardening after this tag |

---

## 9. Current Backlog Notes

Track these as documentation or product follow-up, not as completed Sprint 20 scope:

```text
English mode still has deeper table/filter labels that need localization cleanup.
Web auth UI remains mock/staging-only until backend auth/session API wiring is explicit.
Any future production-like release must record whether prototype fallback gaps remain.
```
