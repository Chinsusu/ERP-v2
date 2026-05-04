# 90_ERP_Coding_Task_Board_Sprint23_Inventory_Purchase_Warehouse_Documents_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 23 - Inventory, Purchase Request, Stock Transfer, and Warehouse Documents
Document role: Candidate coding/task board for the next implementation sprint
Version: v1
Date: 2026-05-04
Status: Proposed; not started; no release tag created
Previous sprint: Sprint 22 - UAT Pilot Pack for Warehouse + Sales + QC

---

## 1. Executive Summary

Sprint 23 is the recommended next functional hardening sprint after the spreadsheet workflow review.

The reviewed warehouse/purchase sheets confirm that the current ERP direction is correct:

```text
Master data for SKU/warehouse/supplier identity.
Stock ledger for quantity truth.
Documents for business intent and approval.
Reports for dashboard/overview.
```

The sheets also reveal four missing or incomplete operational surfaces:

```text
1. Stock Transfer / Chuyển kho
2. Purchase Request / Đề nghị mua
3. Warehouse Issue Note / Phiếu xuất kho
4. Inventory Dashboard backed by stock movement ledger
```

Sprint 23 should not recreate the spreadsheet as one editable grid. It should turn the spreadsheet workflow into controlled ERP documents.

---

## 2. Sprint Goal

Sprint 23 goal:

```text
Implement the missing document boundaries that make warehouse and purchase spreadsheet tracking unnecessary for Phase 1 pilot operations.
```

The sprint succeeds when:

```text
1. Warehouse users can create and post same-SKU stock transfers.
2. Warehouse users can create/export warehouse issue notes from approved outbound lines.
3. Purchase users can create, submit, approve, and convert purchase requests.
4. Inventory dashboard values are derived from stock movement/read model data and drill down to source rows.
5. No stock balance is edited manually outside ledger movements.
```

---

## 3. In Scope

```text
- Stock Transfer document: header, lines, status flow, posting to movement ledger
- Warehouse Issue document: manual issue header/lines, approval/posting, printable/exportable issue note
- Purchase Request document: header/lines, approval status, conversion link to PO/subcontract/service follow-up
- Inventory Dashboard hardening: low stock, out of stock, movement summary, channel/category summary, source drilldowns
- Vietnamese UI labels for these surfaces
- Permission checks for warehouse, purchase, finance observer, and admin roles
- OpenAPI/API/client updates if runtime implementation is done
- Dev smoke coverage for document creation/posting and dashboard derivation
```

---

## 4. Out of Scope

```text
- Full accounting general ledger
- Real bank/payment gateway integration
- Real supplier e-invoice integration
- Advanced MRP
- Automatic PO creation directly from formula calculation
- Internal MES/work-center production
- Complex costing variance
- Marketplace/carrier API integration
- Multi-level amount-based approval matrix beyond first practical cut
- SKU-changing conversion/repack implementation unless explicitly selected as a separate slice
```

Important boundary:

```text
SKU-changing rows from the transfer spreadsheet are not normal stock transfer.
They are conversion/repack/gift split and must not be hidden inside stock transfer.
```

---

## 5. Primary References

| Ref | Document / Source | Use |
| --- | --- | --- |
| `89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md` | Design source doc | Primary design for this sprint |
| Google Sheet gid `471150862` | Tổng quan | Dashboard/report evidence |
| Google Sheet gid `2123530252` | Quản Lý Kho | Inventory snapshot evidence |
| Google Sheet gid `1060593906` | Đề Nghị Mua Hàng | Purchase request tracker evidence |
| Google Sheet gid `2139731924` | Nhập Kho | Inbound movement evidence |
| Google Sheet gid `1438559352` | Xuất Kho | Outbound movement evidence |
| Google Sheet gid `259974100` | Chuyển Kho | Transfer/conversion evidence |
| Google Sheet gid `1174387056` | Phiếu Xuất Kho | Print/export issue note evidence |
| `88_ERP_BOM_Formula_Module_Design_MyPham_v1.md` | BOM/formula design | Future formula-to-purchase-request relationship |
| `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` | Purchase/Receiving/QC | Existing purchase and inbound QC flow |
| `59_ERP_Coding_Task_Board_Sprint11_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1.md` | Inventory read model | Existing inventory persistence direction |
| `65_ERP_Coding_Task_Board_Sprint14_Shipping_Pick_Pack_Persistence_MyPham_v1.md` | Shipping/pick/pack | Existing outbound fulfillment flow |

---

## 6. Branch, PR, and Release Rules

Default branch:

```bash
git checkout main
git pull origin main
git checkout -b codex/s23-inventory-purchase-warehouse-documents
```

Task branch examples:

```text
codex/feature-S23-stock-transfer
codex/feature-S23-warehouse-issue-note
codex/feature-S23-purchase-request
codex/feature-S23-inventory-dashboard-ledger
codex/docs-S23-release-evidence
```

Release tag recommendation after completed implementation and verification:

```text
v0.23.0-inventory-purchase-warehouse-documents
```

Tag hold rule:

```text
Do not create v0.23 until code is merged, CI is green, dev/staging smoke passes, and changelog evidence is complete.
```

---

## 7. Guardrails

```text
1. Do not manually mutate stock balance.
2. Do not create dashboard-only stock numbers.
3. Do not allow stock transfer to change SKU identity.
4. Do not make Purchase Request equal Purchase Order.
5. Do not make Purchase Request equal receiving/payment/invoice.
6. Do not bypass inbound QC for QC-required items.
7. Do not let QC_HOLD/QC_FAIL stock become available.
8. Do not duplicate issue note quantities outside the source outbound document.
9. Do not localize API routes, enum values, permission keys, or audit event codes into Vietnamese.
10. Do not implement conversion/repack inside stock transfer as a shortcut.
```

---

## 8. Dependency Map

```text
S23-00 Design and data mapping
  -> S23-01 Stock movement taxonomy
  -> S23-02 Stock Transfer
  -> S23-03 Warehouse Issue Note
  -> S23-04 Purchase Request
  -> S23-05 Inventory Dashboard
  -> S23-06 Smoke, docs, changelog

S23-01 Stock movement taxonomy
  -> S23-02 Stock Transfer posting
  -> S23-03 Warehouse Issue posting
  -> S23-05 Dashboard ledger drilldowns

S23-04 Purchase Request
  -> S23-04 conversion links to Purchase/Subcontract/Finance
  -> future BOM formula requirement purchase request draft
```

---

## 9. Task Board

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
| --- | ---: | --- | --- | --- | --- |
| S23-00-01 | P0 | BA/PM | Confirm sheet-to-ERP mapping | Decision log accepts spreadsheet split into master data, documents, ledger, reports | `89_ERP_Inventory_Purchase_Warehouse_Document_Flow_Design_MyPham_v1.md` |
| S23-00-02 | P0 | BA/PM | Confirm stock transfer boundary | Business confirms same-SKU transfer; SKU-changing rows are conversion/repack backlog | `89`, section 6 |
| S23-00-03 | P1 | BA/PM | Confirm warehouse issue note usage | Decide print/export from manifest, manual issue, or both | `89`, section 7 |
| S23-00-04 | P1 | BA/PM | Confirm purchase request approval cut | One-level approval or amount-based approval; default one-level | `89`, section 14 |
| S23-01-01 | P0 | BE | Movement taxonomy hardening | Movement types/source document fields support transfer and warehouse issue | `89`, section 9 |
| S23-01-02 | P0 | BE | Ledger invariant tests | Tests prove stock balance changes require source movements and approved reason | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S23-02-01 | P0 | BE | Stock Transfer domain model | Transfer header/lines/status validation implemented | `89`, section 6 |
| S23-02-02 | P0 | BE | Stock Transfer migrations/store/API | PostgreSQL-backed transfer documents and API endpoints | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S23-02-03 | P0 | BE | Stock Transfer posting | Posting creates transfer-out and transfer-in movements, same SKU only | `89`, section 6.5 |
| S23-02-04 | P1 | FE | Stock Transfer UI | Vietnamese UI supports list/detail/create/submit/approve/post/cancel | `81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1.md` |
| S23-02-05 | P1 | QA | Stock Transfer smoke | Create same-SKU transfer, post it, verify source/destination balances and ledger rows | `89`, section 6 |
| S23-03-01 | P0 | BE | Warehouse Issue domain model | Manual issue header/lines/status validation implemented | `89`, section 7 |
| S23-03-02 | P0 | BE | Warehouse Issue posting | Approved issue posts outbound movement with reason/source document | `89`, section 7 |
| S23-03-03 | P1 | FE | Warehouse Issue UI | Vietnamese UI supports issue list/detail/create/approve/post/cancel | `89`, section 7 |
| S23-03-04 | P1 | FE/BE | Issue note print/export | Print/export renders company/date/destination/SKU/category/qty/signature area from source lines | `89`, section 7.4 |
| S23-03-05 | P1 | QA | Issue note smoke | Posted issue exports/prints evidence and ledger row matches quantity | Google Sheet gid `1174387056` |
| S23-04-01 | P0 | BE | Purchase Request domain model | Header/lines/status/approval/conversion link model implemented | `89`, section 5 |
| S23-04-02 | P0 | BE | Purchase Request migrations/store/API | PostgreSQL-backed PR documents and API endpoints | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S23-04-03 | P0 | BE | Purchase Request approval | Draft -> submit -> approve/reject/cancel status flow with audit | `89`, section 5.4 |
| S23-04-04 | P1 | BE | Purchase Request conversion links | Approved PR lines can link to PO/subcontract/service payable draft without becoming receipt/payment/invoice | `89`, section 5.5 |
| S23-04-05 | P1 | FE | Purchase Request UI | Vietnamese UI supports list/detail/create/submit/approve/reject/convert | Google Sheet gid `1060593906` |
| S23-04-06 | P1 | QA | Purchase Request smoke | Create PR, submit, approve, convert line, verify linked status and no stock movement is created by PR alone | `89`, section 5 |
| S23-05-01 | P0 | BE/Reporting | Inventory dashboard query alignment | KPIs use stock movement/read model, not manual values | `89`, section 8 |
| S23-05-02 | P1 | FE/Reporting | Inventory dashboard UI hardening | Low stock, out of stock, inbound/outbound, top movers, inventory value show with drilldowns | Google Sheet gid `471150862` |
| S23-05-03 | P1 | QA | Dashboard drilldown smoke | KPI/drilldown source rows reconcile for one SKU and one period | `89`, section 8.4 |
| S23-06-01 | P0 | BE/FE | OpenAPI/client updates | API routes and generated client updated when runtime endpoints are added | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S23-06-02 | P0 | QA | Required CI/checks | api/web/openapi/migration/e2e checks green | `README.md` |
| S23-06-03 | P0 | DevOps/QA | Dev deploy and smoke | Dev deploy passes; stock transfer, issue note, PR, dashboard smoke pass | `infra/scripts/deploy-dev-staging.sh` |
| S23-06-04 | P0 | PM | Sprint 23 changelog | Changelog records PRs, checks, deploy, smoke, known limits, tag status | `80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md` |

---

## 10. Demo Scripts

### 10.1. Stock Transfer

```text
1. Login as warehouse manager.
2. Open Kho hàng -> Chuyển kho.
3. Create transfer from KHO TỔNG to a destination warehouse/location.
4. Add one SKU with quantity.
5. Submit and approve transfer.
6. Post transfer.
7. Confirm source stock decreases.
8. Confirm destination stock increases.
9. Confirm ledger shows transfer_out and transfer_in source-linked movements.
10. Try changing SKU between out and in; confirm system blocks it.
```

Expected:

```text
Same SKU moves between warehouses.
No stock disappears.
No SKU conversion is hidden in transfer.
```

---

### 10.2. Warehouse Issue Note

```text
1. Login as warehouse manager.
2. Create warehouse issue note for destination/channel SHOPEE or internal reason.
3. Add multiple SKU lines.
4. Submit and approve.
5. Post issue.
6. Export/print issue note.
7. Verify printed lines match posted source lines.
8. Verify outbound ledger rows match quantities.
```

Expected:

```text
Printed Phiếu xuất kho is evidence from source document lines.
Users do not re-enter quantities into a print-only table.
```

---

### 10.3. Purchase Request

```text
1. Login as purchase user.
2. Create purchase request with stock item line.
3. Add supplier suggestion and estimated price.
4. Submit.
5. Approve as authorized user.
6. Convert approved line to PO draft.
7. Confirm PR shows converted/linked status.
8. Confirm no stock movement exists until receiving/QC flow.
9. Create service/subcontract line.
10. Convert to subcontract/service follow-up instead of stock PO where configured.
```

Expected:

```text
PR records demand and approval.
PO/receiving/payment/invoice remain downstream documents.
```

---

### 10.4. Inventory Dashboard

```text
1. Open dashboard for selected date range.
2. Review low stock and out-of-stock counts.
3. Drill into one low-stock SKU.
4. Confirm stock availability row matches dashboard.
5. Open movement history.
6. Confirm inbound/outbound net matches movements for selected period.
7. Open channel/category summary.
8. Confirm outbound movements explain top moving products.
```

Expected:

```text
Dashboard numbers are explainable from ledger/read model.
No manual dashboard-only stock value exists.
```

---

## 11. Permissions

Recommended permissions:

```text
stock_transfer:view
stock_transfer:create
stock_transfer:submit
stock_transfer:approve
stock_transfer:post
stock_transfer:cancel

warehouse_issue:view
warehouse_issue:create
warehouse_issue:submit
warehouse_issue:approve
warehouse_issue:post
warehouse_issue:export
warehouse_issue:cancel

purchase_request:view
purchase_request:create
purchase_request:submit
purchase_request:approve
purchase_request:convert
purchase_request:cancel

inventory_dashboard:view
```

Role direction:

| Role | Access |
| --- | --- |
| ERP Admin | Full |
| Warehouse Staff | Create/view transfer and issue drafts; view dashboard |
| Warehouse Manager | Approve/post transfer and issue |
| Purchase Staff | Create/submit PR |
| Purchase Manager | Approve/convert PR |
| Finance | View PR/issue/transfer, view payment/invoice linked status |
| QC | View transfers/issues affecting QC hold/quarantine |
| Sales | View issue/export status for sales/channel fulfillment where relevant |

---

## 12. Data And Validation Rules

### 12.1. Stock Transfer

```text
- source warehouse != destination warehouse unless moving between locations in same warehouse is explicitly supported
- SKU must be same for source and destination
- quantity > 0
- source available stock must be sufficient unless transfer is from QC/damaged stock with matching status
- batch required if item is batch-controlled
- expiry retained for batch-controlled stock
- posted transfer cannot be edited
```

### 12.2. Warehouse Issue

```text
- issue reason required
- destination required
- quantity > 0
- source stock must be available for issue unless reason explicitly allows non-available status
- posted issue cannot be edited
- export/print only from submitted/approved/posted source document according to selected policy
```

### 12.3. Purchase Request

```text
- request reason/purpose required
- requested quantity > 0
- UOM valid
- stock lines require item/SKU
- service lines require service description or service item
- approval required before conversion
- PR alone does not affect stock
- downstream documents maintain traceability to PR lines
```

### 12.4. Dashboard

```text
- current balance derives from stock read model
- movement numbers derive from ledger
- inventory value derives from stock quantity * valuation source
- every KPI has drilldown
- date range uses Asia/Ho_Chi_Minh
- UI numbers use vi-VN/VND formatting
```

---

## 13. Verification Gates

Local/code checks for code-change PRs:

```text
make api-test
make api-lint
make web-test
make web-lint
make openapi-validate
```

If local workstation lacks tools, run equivalent checks in Docker/dev host and record the exact command.

CI checks:

```text
required-api
required-web
required-openapi
required-migration
api-ci
web-ci
openapi-ci
e2e-ci
```

Dev smoke after runtime/UI changes:

```text
deploy-dev-staging.sh dev
smoke-dev-full.sh
browser smoke for changed screens
```

Minimum Sprint 23 smoke:

```text
stock_transfer_create_submit_approve_post
warehouse_issue_create_approve_post_export
purchase_request_create_submit_approve_convert
inventory_dashboard_drilldown_reconciles_to_ledger
```

---

## 14. Definition of Ready

Sprint 23 is ready to start when:

```text
1. File 89 design is reviewed and accepted.
2. Business confirms stock transfer cannot change SKU.
3. Business confirms PR approval default.
4. Business confirms warehouse issue note print/export source.
5. Existing main is green and deployable.
6. Current UAT/Sprint 22 status remains honestly documented.
```

---

## 15. Definition of Done

Sprint 23 is done when:

```text
1. Stock Transfer implementation is merged and smoke-tested.
2. Warehouse Issue Note implementation is merged and smoke-tested.
3. Purchase Request implementation is merged and smoke-tested.
4. Inventory dashboard/report hardening is merged and smoke-tested.
5. No stock quantity changes bypass ledger movements.
6. PR conversion links preserve downstream traceability.
7. CI/checks are green.
8. Dev deployment and related browser smoke pass.
9. Sprint 23 changelog records evidence and known limits.
10. v0.23 tag is created only if release evidence is complete; otherwise tag hold is documented.
```

---

## 16. Known Risks

```text
1. Purchase Request can grow too large if payment/invoice fields are copied from the sheet into PR itself.
2. Stock Transfer can corrupt inventory if SKU-changing rows are allowed.
3. Warehouse Issue Note can become duplicate data entry if print/export is not source-linked.
4. Dashboard can become untrustworthy if values are not ledger-backed.
5. Conversion/repack may be a real business process and should be designed separately, not hidden.
```

---

## 17. Recommended Changelog Name

After implementation:

```text
91_ERP_Sprint23_Changelog_Inventory_Purchase_Warehouse_Documents_MyPham_v1.md
```

If Sprint 23 is split into smaller sprints, keep file 90 as the parent roadmap and create smaller sprint boards for each slice.

---

## 18. One-Line Rule

```text
Sprint 23 should replace spreadsheet tracking with controlled ERP documents, not recreate spreadsheet editing inside ERP.
```
