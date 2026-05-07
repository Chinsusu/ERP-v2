# 80_ERP_Master_Document_Index_Current_Status_MyPham_v2

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Current master document index and traceability map
Version: v2.23
Date: 2026-05-07
Status: Current source-of-truth index for current Phase 1 docs, design addenda, Sprint 23 runtime bridge, Sprint 24 production material issue readiness runtime, Sprint 25 subcontract closeout traceability evidence, Sprint 26 production IA cleanup evidence, Sprint 27 factory dispatch MVP, Sprint 28 factory execution tracking closeout evidence, Sprint 29 factory material handover closeout evidence, Sprint 30 factory sample/mass-production closeout evidence, Sprint 31 factory finished-goods receipt to QC hold closeout evidence, Sprint 32 factory finished-goods QC closeout scope, Sprint 33 factory claim/final-payment closeout scope, Sprint 34 factory final-payment AP handoff evidence, and Sprint 35 factory final-payment Finance closeout evidence

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
Current line: Sprint 35 factory final-payment Finance closeout for external-factory production.
Latest release tag: v0.19.0-vietnamese-ui-localization.
Sprint 21 tag status: hold; no v0.21.0-auth-ui-backend-integration-runtime-smoke tag has been created pending target staging/pilot smoke evidence.
Sprint 21 merge evidence: PR #542 merged to main at c07409cc; CI, dev deploy, full dev smoke, and auth UI browser smoke passed.
Sprint 22 status: UAT pilot pack prepared; S22-ISSUE-001 resolved by PR #546 at db894ddb; Session 0 readiness rerun passed; business UAT execution, business issue triage, Go/No-Go decision, and v0.22 tag are pending.
Sprint 23 implementation status: first runtime bridge selected in file 92 adds /production planning UI, backend production-plan API, active-formula snapshot, material demand/shortage calculation, internal Purchase Request draft lines, PostgreSQL persistence, and OpenAPI contract coverage; follow-up file 94 promotes Purchase Request submit/approve/convert-to-PO workflow; follow-up file 95 locks PO -> receiving -> QC PASS -> supplier payable traceability; follow-up file 96 locks supplier invoice and 3-way matching behavior; follow-up file 97 locks AP payment readiness so payment request/approval/recording require a matched supplier invoice; follow-up file 98 adds Stock Transfer and Warehouse Issue Note runtime documents with PostgreSQL persistence, submit/approve/post lifecycle, OpenAPI coverage, and posted inventory movements; no v0.23 tag exists.
Sprint 24 implementation status: file 99 tracks the task board, file 100 locks the flow, and file 101 records changelog/evidence. Runtime PR #586 merged at 9e28c05e; dev web API-base fix PR #587 merged at 114105b2; CI, dev deploy, full dev smoke, and /production browser smoke passed. No v0.24 tag exists.
Sprint 25 implementation status: file 102 tracks the task board, file 103 locks the flow, and file 104 records changelog/evidence. PR #589 merged at a4b96c84 with GitHub CI green. Dev deploy passed on 2026-05-06 with migration 43 applied and full dev smoke passed. Browser smoke passed for Production Plan detail -> source-linked Subcontract Order visibility -> /subcontract source filter. No v0.25 tag exists.
Sprint 26 implementation status: file 105 tracks the task board, file 106 locks the Production IA/external factory order detail flow, and file 107 records implementation evidence. PR #591 merged at 5e8003a9 with GitHub CI green; dev deploy, full dev smoke, and Production browser smoke passed on 2026-05-06. Production is the user-facing module; external factory/subcontract is the current production execution method. /subcontract remains a hidden technical/legacy execution route rather than a primary sidebar sibling. No v0.26 tag exists.
Sprint 27 implementation status: file 108 tracks the task board, file 109 locks the factory dispatch flow, and file 110 records changelog/evidence. PR #593 merged at 3cc5852d with GitHub CI green. Dev deploy passed on 2026-05-06 with migration 44 applied and full dev smoke passed. Browser smoke passed for /production/factory-orders/:orderId factory dispatch create -> ready -> sent -> confirmed. Scope is manual factory dispatch pack creation, ready/sent evidence, and factory response on /production/factory-orders/:orderId. Email, Zalo, factory portal/API delivery, digital signatures, and internal MES production remain out of scope. No v0.27 tag exists.
Sprint 28 implementation status: file 111 tracks the task board, file 112 locks the factory execution tracking flow, and file 113 records changelog/evidence. Scope is a production-facing current gate/worklist on /production/factory-orders/:orderId after factory dispatch confirmation. It links to deposit, material handover, sample, mass production, finished goods receipt, QC/claim, and final payment readiness through existing execution surfaces. Email, Zalo, factory portal/API delivery, and internal MES production remain out of scope. PR #595 merged at cd3a5b18 with GitHub CI green. Dev deploy passed on 2026-05-06 with no new migration; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s16-07-01-1777715855439203730. No v0.28 tag exists.
Sprint 29 implementation status: file 114 tracks the task board, file 115 locks the factory material handover flow, and file 116 records changelog/evidence. Scope is a production-facing material handover section on /production/factory-orders/:orderId using the existing issue-materials runtime. It records source warehouse, receiver, contact, vehicle, handover evidence, issue quantity, batch/lot, bin, transfer result, stock movement evidence, and in-page order state update. Tracker and timeline material actions point to #factory-material-handover instead of hidden /subcontract transfer. Email, Zalo, factory portal/API delivery, warehouse issue redesign, and internal MES production remain out of scope. PR #597 merged at 7fd3b2d5 with GitHub CI green. Dev deploy passed on 2026-05-06; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s16-02-01-1777715855392710950#factory-material-handover with screenshot output/playwright/s29-factory-material-handover.png. No v0.29 tag exists.
Sprint 30 implementation status: file 117 tracks the task board, file 118 locks the factory sample approval / mass-production start flow, and file 119 records changelog/evidence. Scope is a production-facing sample approval section and mass-production start section on /production/factory-orders/:orderId using existing submit-sample, approve-sample, reject-sample, and start-mass-production runtime APIs. Tracker and timeline sample/mass actions point to #factory-sample-approval and #factory-mass-production. Email, Zalo, factory portal/API delivery, finished-goods receipt, inbound QC, and internal MES production remain out of scope. PR #599 merged at bd645404 with GitHub CI green. Dev deploy passed on 2026-05-06; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s16-08-03-smoke-0063#factory-sample-approval with screenshot output/playwright/s30-factory-sample-mass-production.png. No v0.30 tag exists.
Sprint 31 implementation status: file 120 tracks the task board, file 121 locks the factory finished-goods receipt to QC hold flow, and file 122 records changelog/evidence. Scope is a production-facing finished-goods receipt section on /production/factory-orders/:orderId using existing receiveSubcontractFinishedGoods runtime APIs. Tracker and timeline finished-goods actions point to #factory-finished-goods-receipt instead of hidden /subcontract inbound. QC pass/fail, available-stock posting, factory claim closeout, final payment release, email, Zalo, factory portal/API delivery, and internal MES production remain out of scope. PR #601 merged at 7b7952fb with GitHub CI green. Dev deploy passed on 2026-05-06; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s16-08-03-smoke-0064#factory-finished-goods-receipt with screenshot output/playwright/s31-factory-finished-goods-receipt.png. No v0.31 tag exists.
Sprint 32 implementation status: file 123 tracks the task board, file 124 locks the factory finished-goods QC closeout flow, and file 125 records changelog/evidence. Scope is a production-facing QC closeout section on /production/factory-orders/:orderId#factory-finished-goods-qc-closeout using existing accept, partial-accept, and report-factory-defect runtime APIs. Tracker and timeline QC actions point to #factory-finished-goods-qc-closeout. Receipt to QC hold remains separate from QC pass; only accepted quantity can become available stock. Final payment readiness, claim resolution, email, Zalo, factory portal/API delivery, and internal MES production remain out of scope. PR #604 merged at 90cae3fb with GitHub CI green. Dev deploy passed on 2026-05-06; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s16-08-03-smoke-0066#factory-finished-goods-qc-closeout with screenshot output/playwright/s32-factory-finished-goods-qc-closeout.png. No v0.32 tag is planned.
Sprint 33 implementation status: file 126 tracks the task board, file 127 locks the factory claim/final payment closeout flow, and file 128 records changelog/evidence. Scope is a production-facing claim and final payment closeout section on /production/factory-orders/:orderId#factory-claim-final-payment-closeout using factory claim list/acknowledge/resolve runtime APIs plus existing final payment readiness runtime. Tracker and timeline claim/payment actions point to #factory-claim-final-payment-closeout. Open or acknowledged factory claims block final payment; resolved claims can allow final payment only when accepted finished goods exist. Full QC fail remains blocked from final payment readiness until a later replacement/settlement flow exists. PR #606 merged at 5ac8a1e with GitHub CI green. Dev deploy passed on 2026-05-07; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s16-08-03-smoke-0068#factory-claim-final-payment-closeout with screenshot output/playwright/s33-factory-claim-final-payment-closeout.png. No v0.33 tag is planned.
Sprint 34 implementation status: file 129 tracks the task board, file 130 locks the factory final-payment AP handoff flow, and file 131 records changelog/evidence. Scope is connecting final payment readiness on /production/factory-orders/:orderId#factory-claim-final-payment-closeout to Finance supplier payables via supplier_payable response evidence and /finance?ap_q=:payableNo#supplier-payables deep links. Finance remains the payment execution surface and matched supplier invoice remains required before AP payment request, approval, or recording. Runtime PR #608 merged at 602a7354 with GitHub CI green. Dev deploy passed on 2026-05-07; full dev smoke passed. Browser smoke passed for /production/factory-orders/sco-s34-ap-smoke-0507060226#factory-claim-final-payment-closeout -> /finance?ap_q=SCO-S34-AP-SMOKE-0507060226#supplier-payables, showing AP-SPM-S34-AP-SMOKE-0507060226-FINAL. Screenshots: output/playwright/s34-factory-final-payment-ap-handoff.png and output/playwright/s34-finance-ap-handoff.png. No v0.34 tag is planned.
Sprint 35 implementation status: file 132 tracks the task board, file 133 locks the factory final-payment Finance closeout flow, and file 134 records changelog/evidence. Scope is Finance-side closeout guidance for factory final-payment AP records sourced from subcontract_payment_milestone and subcontract_order documents, with AP/invoice/payment closeout steps and a back link to the source production factory order. Runtime PR #610 merged at 68b4d3d5 with GitHub CI green. Follow-up PR #611 merged at 64851338 after S35 smoke found factory final-payment supplier invoice sources were blocked by invoice source validation. Dev deploy passed on 2026-05-07 after PR #611 and full dev smoke passed. Target S35 API smoke passed for AP-SPM-S34-AP-SMOKE-0507060226-FINAL with supplier invoice INV-S35-13609491 matched and AP status paid with 0.00 outstanding. Browser smoke passed for /finance?ap_q=AP-SPM-S34-AP-SMOKE-0507060226-FINAL#supplier-payables with 5 completed Finance closeout steps, backlink to /production/factory-orders/sco-s34-ap-smoke-0507060226#factory-claim-final-payment-closeout, and screenshot output/playwright/s35-finance-factory-payment-closeout-paid.png. Matched supplier invoice remains required before AP payment request, approval, or recording. No v0.35 tag is planned.
Release tag migration gate: PostgreSQL 16 apply + rollback passed.
Current main migration gate after Sprint 20: PostgreSQL 16 apply -> rollback -> reapply passed.
Technical contract: English.
Business display: Vietnamese-first.
Routes: English.
Locale: vi-VN.
Currency: VND.
Timezone: Asia/Ho_Chi_Minh.
Phase 1 production entrypoints: /production is planning/material-demand/PR-draft review, external-factory production navigation, factory order detail, manual factory dispatch, factory execution tracking, material handover, sample approval, mass-production start, finished-goods receipt to QC hold, finished-goods QC closeout, factory claim resolution, final payment closeout, and final-payment AP handoff; /finance is final-payment invoice matching, AP payment closeout, and factory final-payment Finance closeout; /subcontract remains hidden route-addressable external factory execution.
Purchase flow boundary: /production opens generated Purchase Request; PO creation belongs to approved Purchase Request conversion, not direct production-page shortcut.
Post-PO finance boundary: posted PO-linked goods receipts create supplier payable value only for QC PASS lines; supplier invoice and three-way match are locked in file 96 as separate vendor-bill evidence; AP payment readiness hard gate is locked in file 97.
Warehouse document boundary: Stock Transfer is internal stock movement; Warehouse Issue Note is operational stock issue to factory/lab/manual destination; both are inventory documents, not costing documents.
Production material issue boundary: shortage remains a purchase/receiving/QC problem; ready stock can create source-linked Warehouse Issue Note; subcontract execution must wait for posted material issue evidence or an explicit waiver.
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
18. 97_ERP_AP_Payment_Readiness_Gate_Supplier_Invoice_Matching_MyPham_v1.md
19. 98_ERP_Stock_Transfer_Warehouse_Issue_Runtime_Flow_MyPham_v1.md
20. 99_ERP_Coding_Task_Board_Sprint24_Production_Material_Issue_Readiness_MyPham_v1.md
21. 100_ERP_Production_Material_Issue_Subcontract_Readiness_Flow_MyPham_v1.md
22. 101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1.md
23. 102_ERP_Coding_Task_Board_Sprint25_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md
24. 103_ERP_Subcontract_Finished_Goods_QC_Closeout_Flow_MyPham_v1.md
25. 104_ERP_Sprint25_Changelog_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md
26. 105_ERP_Coding_Task_Board_Sprint26_Production_IA_External_Factory_Order_Detail_MyPham_v1.md
27. 106_ERP_Production_IA_External_Factory_Order_Detail_Flow_MyPham_v1.md
28. 107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1.md
29. 108_ERP_Coding_Task_Board_Sprint27_Factory_Dispatch_MyPham_v1.md
30. 109_ERP_Factory_Dispatch_Flow_Sprint27_MyPham_v1.md
31. 110_ERP_Sprint27_Changelog_Factory_Dispatch_MyPham_v1.md
32. 111_ERP_Coding_Task_Board_Sprint28_Factory_Execution_Tracking_MyPham_v1.md
33. 112_ERP_Factory_Execution_Tracking_Flow_Sprint28_MyPham_v1.md
34. 113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1.md
35. 114_ERP_Coding_Task_Board_Sprint29_Factory_Material_Handover_MyPham_v1.md
36. 115_ERP_Factory_Material_Handover_Flow_Sprint29_MyPham_v1.md
37. 116_ERP_Sprint29_Changelog_Factory_Material_Handover_MyPham_v1.md
38. 117_ERP_Coding_Task_Board_Sprint30_Factory_Sample_Mass_Production_MyPham_v1.md
39. 118_ERP_Factory_Sample_Mass_Production_Flow_Sprint30_MyPham_v1.md
40. 119_ERP_Sprint30_Changelog_Factory_Sample_Mass_Production_MyPham_v1.md
41. 120_ERP_Coding_Task_Board_Sprint31_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1.md
42. 121_ERP_Factory_Finished_Goods_Receipt_QC_Hold_Flow_Sprint31_MyPham_v1.md
43. 122_ERP_Sprint31_Changelog_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1.md
44. 123_ERP_Coding_Task_Board_Sprint32_Factory_Finished_Goods_QC_Closeout_MyPham_v1.md
45. 124_ERP_Factory_Finished_Goods_QC_Closeout_Flow_Sprint32_MyPham_v1.md
46. 125_ERP_Sprint32_Changelog_Factory_Finished_Goods_QC_Closeout_MyPham_v1.md
47. 126_ERP_Coding_Task_Board_Sprint33_Factory_Claim_Final_Payment_Closeout_MyPham_v1.md
48. 127_ERP_Factory_Claim_Final_Payment_Closeout_Flow_Sprint33_MyPham_v1.md
49. 128_ERP_Sprint33_Changelog_Factory_Claim_Final_Payment_Closeout_MyPham_v1.md
50. 129_ERP_Coding_Task_Board_Sprint34_Factory_Final_Payment_AP_Handoff_MyPham_v1.md
51. 130_ERP_Factory_Final_Payment_AP_Handoff_Flow_Sprint34_MyPham_v1.md
52. 131_ERP_Sprint34_Changelog_Factory_Final_Payment_AP_Handoff_MyPham_v1.md
53. 132_ERP_Coding_Task_Board_Sprint35_Factory_Final_Payment_Finance_Closeout_MyPham_v1.md
54. 133_ERP_Factory_Final_Payment_Finance_Closeout_Flow_Sprint35_MyPham_v1.md
55. 134_ERP_Sprint35_Changelog_Factory_Final_Payment_Finance_Closeout_MyPham_v1.md
56. 88_ERP_BOM_Formula_Module_Design_MyPham_v1.md
57. 78_ERP_Production_Runtime_Mode_Checklist_Sprint20_MyPham_v1.md
58. 75_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1.md
59. 77_ERP_Sprint19_Changelog_Vietnamese_UI_Localization_MyPham_v1.md
60. 81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md
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
17. 97 AP payment readiness gate
18. 98 Stock transfer and warehouse issue runtime flow
19. 99 Sprint 24 production material issue task board
20. 100 Production material issue and subcontract readiness flow
21. 101 Sprint 24 changelog and runtime evidence
22. 102 Sprint 25 subcontract finished goods QC closeout task board
23. 103 Subcontract finished goods QC closeout flow
24. 104 Sprint 25 changelog and traceability evidence
25. 105 Sprint 26 production IA task board
26. 106 Production IA and external factory order detail flow
27. 107 Sprint 26 changelog and evidence register
28. 108 Sprint 27 factory dispatch task board
29. 109 Factory dispatch flow
30. 110 Sprint 27 changelog and evidence register
31. 111 Sprint 28 factory execution tracking task board
32. 112 Factory execution tracking flow
33. 113 Sprint 28 changelog and evidence register
34. 114 Sprint 29 factory material handover task board
35. 115 Factory material handover flow
36. 116 Sprint 29 changelog and evidence register
37. 117 Sprint 30 factory sample and mass-production task board
38. 118 Factory sample and mass-production flow
39. 119 Sprint 30 changelog and evidence register
40. 120 Sprint 31 factory finished goods receipt task board
41. 121 Factory finished goods receipt to QC hold flow
42. 122 Sprint 31 changelog and evidence register
43. 123 Sprint 32 factory finished goods QC closeout task board
44. 124 Factory finished goods QC closeout flow
45. 125 Sprint 32 changelog and evidence register
46. 126 Sprint 33 factory claim and final payment closeout task board
47. 127 Factory claim and final payment closeout flow
48. 128 Sprint 33 changelog and evidence register
49. 129 Sprint 34 factory final-payment AP handoff task board
50. 130 Factory final-payment AP handoff flow
51. 131 Sprint 34 changelog and evidence register
52. 132 Sprint 35 factory final-payment Finance closeout task board
53. 133 Factory final-payment Finance closeout flow
54. 134 Sprint 35 changelog and evidence register
55. 88 BOM / formula module design
56. 78 Production runtime checklist
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
| Sprint 23 | `92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md` | Files `94`-`98` runtime flow evidence; changelog pending | Production planning, PR/PO/receiving/AP traceability, supplier invoice/payment gate, stock transfer, and warehouse issue runtime |
| Sprint 24 | `99_ERP_Coding_Task_Board_Sprint24_Production_Material_Issue_Readiness_MyPham_v1.md` | `101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1.md` | Production material issue and subcontract readiness |
| Sprint 25 | `102_ERP_Coding_Task_Board_Sprint25_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md` | `104_ERP_Sprint25_Changelog_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md` | Production Plan to source-linked Subcontract Order closeout traceability |
| Sprint 26 | `105_ERP_Coding_Task_Board_Sprint26_Production_IA_External_Factory_Order_Detail_MyPham_v1.md` | `107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1.md` | Production IA cleanup and external factory order detail |
| Sprint 27 | `108_ERP_Coding_Task_Board_Sprint27_Factory_Dispatch_MyPham_v1.md` | `110_ERP_Sprint27_Changelog_Factory_Dispatch_MyPham_v1.md` | Manual factory dispatch pack, send evidence, and factory response |
| Sprint 28 | `111_ERP_Coding_Task_Board_Sprint28_Factory_Execution_Tracking_MyPham_v1.md` | `113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1.md` | Factory execution tracker, current gate, next action, and post-dispatch worklist |
| Sprint 29 | `114_ERP_Coding_Task_Board_Sprint29_Factory_Material_Handover_MyPham_v1.md` | `116_ERP_Sprint29_Changelog_Factory_Material_Handover_MyPham_v1.md` | Production-facing factory material handover using existing issue-materials runtime |
| Sprint 30 | `117_ERP_Coding_Task_Board_Sprint30_Factory_Sample_Mass_Production_MyPham_v1.md` | `119_ERP_Sprint30_Changelog_Factory_Sample_Mass_Production_MyPham_v1.md` | Production-facing factory sample approval and mass-production start using existing subcontract runtime |
| Sprint 31 | `120_ERP_Coding_Task_Board_Sprint31_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1.md` | `122_ERP_Sprint31_Changelog_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1.md` | Production-facing factory finished goods receipt into QC hold using existing subcontract runtime |
| Sprint 32 | `123_ERP_Coding_Task_Board_Sprint32_Factory_Finished_Goods_QC_Closeout_MyPham_v1.md` | `125_ERP_Sprint32_Changelog_Factory_Finished_Goods_QC_Closeout_MyPham_v1.md` | Production-facing factory finished goods QC closeout using existing accept, partial-accept, and factory-defect runtime |
| Sprint 33 | `126_ERP_Coding_Task_Board_Sprint33_Factory_Claim_Final_Payment_Closeout_MyPham_v1.md` | `128_ERP_Sprint33_Changelog_Factory_Claim_Final_Payment_Closeout_MyPham_v1.md` | Production-facing factory claim acknowledgement/resolution and final payment readiness gate |
| Sprint 34 | `129_ERP_Coding_Task_Board_Sprint34_Factory_Final_Payment_AP_Handoff_MyPham_v1.md` | `131_ERP_Sprint34_Changelog_Factory_Final_Payment_AP_Handoff_MyPham_v1.md` | Production-facing final-payment AP handoff to Finance supplier payables |
| Sprint 35 | `132_ERP_Coding_Task_Board_Sprint35_Factory_Final_Payment_Finance_Closeout_MyPham_v1.md` | `134_ERP_Sprint35_Changelog_Factory_Final_Payment_Finance_Closeout_MyPham_v1.md` | Finance-side final-payment AP invoice/payment closeout for external-factory production |

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
| `90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md` | Historical Sprint 23 candidate task board | Use as planning context; use file 98 for implemented Stock Transfer / Warehouse Issue runtime, and files 99-100 for production-linked material issue readiness |
| `91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md` | Note-sheet module roadmap and sequencing decision | Before choosing the next implementation order from dashboard, setup, sales, production planning, purchasing, warehouse, and costing requests |
| `92_ERP_Coding_Task_Board_Sprint23_Production_Planning_Material_Demand_MyPham_v1.md` | Selected first Sprint 23 task board | Before changing production planning, formula snapshot, material demand calculation, and Purchase Request draft generation |
| `94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md` | Purchase Request workflow bridge | Before changing production-plan to Purchase Request to PO traceability, approval, or conversion behavior |
| `95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md` | PO, receiving, QC, and supplier payable flow | Before changing post-PO receiving, AP creation, PO timeline AP links, or payable traceability behavior |
| `96_ERP_Supplier_Invoice_Three_Way_Matching_Flow_MyPham_v1.md` | Supplier invoice and 3-way matching flow | Before changing supplier invoice capture, AP invoice matching, or payment-readiness traceability behavior |
| `97_ERP_AP_Payment_Readiness_Gate_Supplier_Invoice_Matching_MyPham_v1.md` | AP payment readiness gate | Before changing supplier payable payment request, payment approval, payment recording, or matched-invoice enforcement |
| `98_ERP_Stock_Transfer_Warehouse_Issue_Runtime_Flow_MyPham_v1.md` | Stock Transfer and Warehouse Issue Note runtime flow | Before changing internal warehouse transfer, operational material issue, posted inventory movements, or warehouse issue UI behavior |
| `99_ERP_Coding_Task_Board_Sprint24_Production_Material_Issue_Readiness_MyPham_v1.md` | Sprint 24 task board | Before starting production-plan material issue and subcontract readiness work |
| `100_ERP_Production_Material_Issue_Subcontract_Readiness_Flow_MyPham_v1.md` | Production material issue flow design | Before changing production-plan to Warehouse Issue Note source-linking, material issue readiness, or subcontract readiness gates |
| `101_ERP_Sprint24_Changelog_Production_Material_Issue_Readiness_MyPham_v1.md` | Sprint 24 changelog and runtime evidence | Before closing Sprint 24 or claiming production material issue readiness evidence |
| `102_ERP_Coding_Task_Board_Sprint25_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md` | Sprint 25 task board | Before changing Production Plan to Subcontract Order traceability or subcontract closeout visibility |
| `103_ERP_Subcontract_Finished_Goods_QC_Closeout_Flow_MyPham_v1.md` | Subcontract finished goods closeout flow design | Before changing Production Plan detail closeout, subcontract receiving/QC visibility, factory claim visibility, or final payment readiness traceability |
| `104_ERP_Sprint25_Changelog_Subcontract_Finished_Goods_QC_Closeout_MyPham_v1.md` | Sprint 25 changelog and evidence | Before closing Sprint 25 or claiming subcontract closeout traceability evidence |
| `105_ERP_Coding_Task_Board_Sprint26_Production_IA_External_Factory_Order_Detail_MyPham_v1.md` | Sprint 26 task board | Before changing Production/Subcontract navigation, factory order detail routes, or production-facing external factory order visibility |
| `106_ERP_Production_IA_External_Factory_Order_Detail_Flow_MyPham_v1.md` | Production IA and factory-order detail flow | Before changing the user-facing Production entrypoint, hidden subcontract route policy, or factory-order timeline |
| `107_ERP_Sprint26_Changelog_Production_IA_External_Factory_Order_Detail_MyPham_v1.md` | Sprint 26 changelog and evidence | Before closing Sprint 26 or claiming Production IA cleanup evidence |
| `108_ERP_Coding_Task_Board_Sprint27_Factory_Dispatch_MyPham_v1.md` | Sprint 27 task board | Before changing manual factory dispatch pack scope, send evidence, or factory response behavior |
| `109_ERP_Factory_Dispatch_Flow_Sprint27_MyPham_v1.md` | Factory dispatch flow design | Before changing the dispatch lifecycle between approved factory order and factory confirmation |
| `110_ERP_Sprint27_Changelog_Factory_Dispatch_MyPham_v1.md` | Sprint 27 changelog and evidence | Before closing Sprint 27 or claiming factory dispatch implementation evidence |
| `111_ERP_Coding_Task_Board_Sprint28_Factory_Execution_Tracking_MyPham_v1.md` | Sprint 28 task board | Before changing the post-dispatch factory execution tracker or current-gate rules |
| `112_ERP_Factory_Execution_Tracking_Flow_Sprint28_MyPham_v1.md` | Sprint 28 flow design | Before changing factory execution current gate, action link, or post-dispatch worklist behavior |
| `113_ERP_Sprint28_Changelog_Factory_Execution_Tracking_MyPham_v1.md` | Sprint 28 changelog and evidence | Before closing Sprint 28 or claiming factory execution tracking evidence |
| `114_ERP_Coding_Task_Board_Sprint29_Factory_Material_Handover_MyPham_v1.md` | Sprint 29 task board | Before changing factory material handover scope or acceptance rules |
| `115_ERP_Factory_Material_Handover_Flow_Sprint29_MyPham_v1.md` | Sprint 29 flow design | Before changing material handover gate, source warehouse, lot/bin, transfer, or movement evidence behavior |
| `116_ERP_Sprint29_Changelog_Factory_Material_Handover_MyPham_v1.md` | Sprint 29 changelog and evidence | Before closing Sprint 29 or claiming factory material handover evidence |
| `117_ERP_Coding_Task_Board_Sprint30_Factory_Sample_Mass_Production_MyPham_v1.md` | Sprint 30 task board | Before changing factory sample approval or mass-production start scope |
| `118_ERP_Factory_Sample_Mass_Production_Flow_Sprint30_MyPham_v1.md` | Sprint 30 flow design | Before changing sample submit/decision gates, sample-required bypass, or mass-production start readiness |
| `119_ERP_Sprint30_Changelog_Factory_Sample_Mass_Production_MyPham_v1.md` | Sprint 30 changelog and evidence | Before closing Sprint 30 or claiming factory sample/mass-production evidence |
| `120_ERP_Coding_Task_Board_Sprint31_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1.md` | Sprint 31 task board | Before changing factory finished-goods receipt to QC hold scope |
| `121_ERP_Factory_Finished_Goods_Receipt_QC_Hold_Flow_Sprint31_MyPham_v1.md` | Sprint 31 flow design | Before changing receipt gate, batch/lot, QC hold, or traceability behavior |
| `122_ERP_Sprint31_Changelog_Factory_Finished_Goods_Receipt_QC_Hold_MyPham_v1.md` | Sprint 31 changelog and evidence | Before closing Sprint 31 or claiming factory receipt evidence |
| `123_ERP_Coding_Task_Board_Sprint32_Factory_Finished_Goods_QC_Closeout_MyPham_v1.md` | Sprint 32 task board | Before changing factory finished-goods QC closeout scope |
| `124_ERP_Factory_Finished_Goods_QC_Closeout_Flow_Sprint32_MyPham_v1.md` | Sprint 32 flow design | Before changing QC pass, partial pass, full fail, available-stock release, or factory claim behavior |
| `125_ERP_Sprint32_Changelog_Factory_Finished_Goods_QC_Closeout_MyPham_v1.md` | Sprint 32 changelog and evidence | Before closing Sprint 32 or claiming factory QC closeout evidence |
| `126_ERP_Coding_Task_Board_Sprint33_Factory_Claim_Final_Payment_Closeout_MyPham_v1.md` | Sprint 33 task board | Before changing factory claim or final payment closeout scope |
| `127_ERP_Factory_Claim_Final_Payment_Closeout_Flow_Sprint33_MyPham_v1.md` | Sprint 33 flow design | Before changing claim acknowledgement, claim resolution, final payment blockers, or closeout timeline behavior |
| `128_ERP_Sprint33_Changelog_Factory_Claim_Final_Payment_Closeout_MyPham_v1.md` | Sprint 33 changelog and evidence | Before closing Sprint 33 or claiming factory claim/final payment closeout evidence |
| `129_ERP_Coding_Task_Board_Sprint34_Factory_Final_Payment_AP_Handoff_MyPham_v1.md` | Sprint 34 task board | Before changing factory final-payment AP handoff scope |
| `130_ERP_Factory_Final_Payment_AP_Handoff_Flow_Sprint34_MyPham_v1.md` | Sprint 34 flow design | Before changing production-to-Finance AP handoff links, supplier_payable response evidence, or final payment AP boundaries |
| `131_ERP_Sprint34_Changelog_Factory_Final_Payment_AP_Handoff_MyPham_v1.md` | Sprint 34 changelog and evidence | Before closing Sprint 34 or claiming factory final-payment AP handoff evidence |
| `132_ERP_Coding_Task_Board_Sprint35_Factory_Final_Payment_Finance_Closeout_MyPham_v1.md` | Sprint 35 task board | Before changing Finance-side factory final-payment AP closeout scope |
| `133_ERP_Factory_Final_Payment_Finance_Closeout_Flow_Sprint35_MyPham_v1.md` | Sprint 35 flow design | Before changing factory AP invoice matching, payment checklist, or source production-order back links |
| `134_ERP_Sprint35_Changelog_Factory_Final_Payment_Finance_Closeout_MyPham_v1.md` | Sprint 35 changelog and evidence | Before closing Sprint 35 or claiming factory final-payment Finance closeout evidence |

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

Sprint 24 tag hold:

```text
No v0.24.0-production-material-issue-readiness tag has been created.
Sprint 24 runtime implementation, CI, dev deploy, full dev smoke, browser smoke, and changelog evidence are complete on main.
Create the Sprint 24 checkpoint tag only if a release checkpoint is intentionally requested.
```

Sprint 25 tag hold:

```text
No v0.25.0-subcontract-finished-goods-qc-closeout tag has been created.
Sprint 25 implementation, CI, dev deploy, full smoke, browser smoke, and changelog evidence are complete on main.
Create a tag only if a release checkpoint is intentionally requested.
```

Sprint 26 tag hold:

```text
No v0.26.0-production-ia-factory-order-detail tag has been created.
Sprint 26 implementation, CI, dev deploy, full smoke, browser smoke, and changelog evidence are complete on main.
Create a tag only if a release checkpoint is intentionally requested.
```

---

## 9. Current Backlog Notes

Track these as documentation or product follow-up, not as completed Sprint 20 scope:

```text
English mode still has deeper table/filter labels that need localization cleanup.
Production-like auth smoke evidence must be recorded per target environment before release.
Any future production-like release must record whether prototype fallback gaps remain.
Sprint 22 business UAT execution remains pending until business users run the prepared scripts.
Sprint 23 production planning and warehouse document runtime bridge is documented through files 92 and 94-98; no v0.23 tag exists.
Sprint 24 runtime evidence is recorded in files 99-101 for production-plan material issue readiness and subcontract readiness gating.
File 90 remains historical planning context for Stock Transfer, Warehouse Issue Note, and future ledger-backed Inventory Dashboard hardening.
Stock Transfer must remain same-SKU only; SKU-changing movements require separate conversion/repack design.
Purchase Request must remain separate from PO, receiving, payment, and invoice source documents.
```
