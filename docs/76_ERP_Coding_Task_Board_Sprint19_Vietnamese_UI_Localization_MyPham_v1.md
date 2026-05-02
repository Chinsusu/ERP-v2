# 76_ERP_Coding_Task_Board_Sprint19_Vietnamese_UI_Localization_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 19 - Vietnamese UI Localization
Document role: Coding task board for Vietnamese-first ERP UI localization.

---

## 1. Sprint Goal

Sprint 19 localizes the ERP frontend into Vietnamese-first UI while keeping backend, API, database, enums, routes, and audit/event codes stable in English.

The goal is not only to translate labels. The goal is to make the ERP speak the actual operational language used by warehouse, sales, QC, purchase, finance, and management teams.

```text
Backend/API/DB codes: English, stable, technical
Frontend display: Vietnamese-first, business-friendly
Routes: English, stable, testable
Locale: vi-VN
Currency: VND
Timezone: Asia/Ho_Chi_Minh
```

---

## 2. Current Context

The product has already delivered core operational flows across previous sprints:

```text
Foundation
Auth/RBAC
Master data
Stock ledger
Batch/QC
Warehouse receiving
Warehouse Daily Board
Sales order fulfillment
Pick/pack
Carrier manifest
Scan handover
Returns
Return inspection
Stock adjustment
Shift closing
Purchase/inbound/QC foundations
Runtime persistence work
```

Because many screens now exist, Vietnamese localization should be done before English hardcoded labels spread deeper into every module.

---

## 3. Primary Design Decision

### 3.1 Keep technical codes in English

Do not translate these into Vietnamese:

```text
API routes
DB enum values
OpenAPI schema names
Error codes
Permission codes
Audit event codes
Workflow event names
Repository/module names
```

Examples:

```text
QC_HOLD
INSUFFICIENT_STOCK
ORDER_HANDED_OVER
RETURN_PENDING_INSPECTION
SHIFT_HAS_UNRESOLVED_ISSUES
```

### 3.2 Translate frontend display

Examples:

```text
QC_HOLD -> Đang giữ kiểm
INSUFFICIENT_STOCK -> Tồn khả dụng không đủ.
ORDER_HANDED_OVER -> Đã bàn giao ĐVVC
RETURN_PENDING_INSPECTION -> Hàng hoàn chờ kiểm
SHIFT_HAS_UNRESOLVED_ISSUES -> Ca làm còn vấn đề chưa xử lý, chưa thể đóng ca.
```

### 3.3 Do not localize routes in this sprint

Keep:

```text
/dashboard
/master-data
/inventory
/sales/orders
/shipping/manifests
/returns
/purchase
/qc
/finance
/settings
```

Do not change to:

```text
/tong-quan
/kho-hang
/ban-hang/don-hang
```

Routes are technical contracts used by permissions, tests, bookmarks, docs, and generated flows.

---

## 4. Source-of-Truth References

| Area | Primary source |
|---|---|
| Frontend architecture | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| UI/UX standard | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` |
| Hetzner-inspired visual template | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| Unit/currency/number format | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| API error/response contract | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| Security/RBAC/audit | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| QA/test strategy | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| Workspace structure | `38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md` |
| Warehouse As-Is workflow | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| Vietnamese operational workflow PDFs | `Công-việc-hằng-ngày.pdf`, `Nội-Quy.pdf`, `Quy-trình-bàn-giao.pdf`, `Quy-trình-sản-xuất.pdf` |

---

## 5. Sprint Scope

### In scope

```text
Vietnamese-first UI foundation
Ant Design vi-VN provider
Translation dictionaries
Vietnamese navigation/menu labels
Vietnamese page titles
Vietnamese table/form labels
Vietnamese actions/buttons
Vietnamese status chips
Vietnamese validation messages
Vietnamese API error-code mapping
Vietnamese empty/loading/error states
Vietnamese warehouse microcopy
Vietnamese VND/date/quantity display
Hardcoded English label cleanup in existing modules
Localization smoke/regression tests
```

### Out of scope

```text
Route localization
Backend enum renaming
Database enum/value renaming
OpenAPI schema renaming
Permission code renaming
Audit/event code renaming
Full multi-language admin panel
Marketing website translation
Legal/accounting report translation beyond existing module labels
```

---

## 6. Branch and Release Naming

Suggested branch:

```bash
git checkout main
git pull origin main
git checkout -b sprint/19-vietnamese-ui-localization
```

Suggested tag after acceptance:

```bash
git tag v0.19.0-vietnamese-ui-localization
git push origin v0.19.0-vietnamese-ui-localization
```

---

## 7. Recommended Folder Structure

Create or normalize:

```text
apps/web/src/shared/i18n/
  index.ts
  config.ts
  formatters.ts
  status-labels.ts
  error-labels.ts
  validation-labels.ts
  units.ts

  locales/
    vi/
      common.json
      navigation.json
      actions.json
      status.json
      errors.json
      validation.json
      masterdata.json
      inventory.json
      warehouse.json
      sales.json
      shipping.json
      returns.json
      purchase.json
      qc.json
      finance.json
      auth.json
      audit.json
      settings.json

    en/
      common.json
      navigation.json
      actions.json
      status.json
      errors.json
      validation.json
      masterdata.json
      inventory.json
      warehouse.json
      sales.json
      shipping.json
      returns.json
      purchase.json
      qc.json
      finance.json
      auth.json
      audit.json
      settings.json
```

Default locale:

```text
vi
```

Fallback locale:

```text
en
```

---

## 8. Localization Guardrails

1. Do not hardcode user-facing English labels in components.
2. Do not translate backend enum values.
3. Do not translate API error codes.
4. Do not translate routes.
5. Do not translate permission keys.
6. Do not translate audit event codes.
7. All status chips must use a centralized status-label mapping.
8. All API errors must use a centralized error-label mapping.
9. All date, money, quantity, and rate displays must use shared formatters.
10. Do not calculate money/quantity using JavaScript floating point for business logic.
11. Vietnamese messages must be short, operational, and clear.
12. Warehouse scan screens must use large, direct, action-oriented Vietnamese copy.

---

## 9. Sprint 19 Backlog

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S19-00-01 | P0 | PM/Tech Lead | Sprint 19 kickoff checklist | Branch created, current main clean, previous release status documented, localization scope confirmed | `37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md` |
| S19-00-02 | P0 | FE Lead | Localization inventory scan | List existing hardcoded English UI labels by module: navigation, dashboard, sales, shipping, returns, purchase, QC, inventory, finance | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| S19-01-01 | P0 | FE | Create i18n foundation | `shared/i18n` folder exists with config, index, locale loader, dictionary access helper | `38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md` |
| S19-01-02 | P0 | FE | Add Vietnamese and English dictionaries | `vi` and `en` dictionaries created for common/navigation/actions/status/errors/validation/modules | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| S19-01-03 | P0 | FE | Add Ant Design Vietnamese provider | App shell wrapped in Ant Design `vi_VN` locale provider; DatePicker/Pagination/Empty use Vietnamese UI | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` |
| S19-01-04 | P0 | FE | Keep routes technical | No route path changes; only display labels change | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S19-02-01 | P0 | FE | Vietnamese navigation labels | Sidebar/topbar menu labels translated: Tổng quan, Dữ liệu gốc, Kho hàng, Bán hàng, Giao hàng, Hàng hoàn, Mua hàng, Kiểm chất lượng, Tài chính, Cài đặt | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S19-02-02 | P0 | FE | Vietnamese page headers | Existing page headers and subtitles translated with short operational copy | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S19-02-03 | P1 | FE | Vietnamese breadcrumbs | Breadcrumbs use Vietnamese display labels while preserving English route keys | `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` |
| S19-03-01 | P0 | FE | Centralized action labels | Buttons/actions use dictionary: Tạo mới, Lưu, Gửi duyệt, Duyệt, Từ chối, Hủy, Xác nhận, Xuất Excel, Làm mới | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` |
| S19-03-02 | P0 | FE | Centralized status labels | All status chips use centralized mapping; no raw `DRAFT`, `PACKED`, `QC_HOLD` shown to users | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S19-03-03 | P0 | FE | Centralized API error labels | Backend error codes map to Vietnamese messages; raw technical error codes hidden unless dev/debug mode | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S19-03-04 | P1 | FE | Centralized validation labels | Required, invalid, min/max, decimal, date, UOM validation messages translated | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S19-04-01 | P0 | FE | VND formatter | Money displays as `1.250.000 ₫`; no UI money calculation using float for business logic | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| S19-04-02 | P0 | FE | Date/time formatter | Date displays `dd/MM/yyyy`; time displays `HH:mm`; timezone is `Asia/Ho_Chi_Minh` | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| S19-04-03 | P0 | FE | Quantity/UOM formatter | Quantity displays with Vietnamese number format and UOM, e.g. `10,5 kg`, `48 chai`, `1.250 g` | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| S19-04-04 | P1 | FE | Percent/rate formatter | Percent/rate displays `2,50%` using shared formatter | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| S19-05-01 | P0 | FE | Dashboard localization | Dashboard KPI cards, alerts, filters, empty states, drill-down labels translated | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S19-05-02 | P0 | FE | Warehouse Daily Board localization | Board labels translated: Đơn mới, Đang soạn, Đã đóng, Chờ bàn giao, Hàng hoàn, Lệch tồn, Đóng ca | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S19-05-03 | P0 | FE | Shift closing localization | Shift closing checklist and blocker messages translated; unresolved issue copy is clear to warehouse users | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S19-06-01 | P0 | FE | Sales order UI localization | Sales order list/detail/create labels, table columns, actions, state chips translated | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S19-06-02 | P0 | FE | Picking UI localization | Pick task screen uses warehouse Vietnamese: Soạn hàng, Mã đơn, Mã hàng, Lô, Vị trí, Số lượng cần lấy, Đã lấy | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S19-06-03 | P0 | FE | Packing UI localization | Packing station labels translated: Đóng hàng, Kiểm SKU, Kiểm số lượng, Khu đóng hàng, Chuyển chờ bàn giao | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S19-06-04 | P0 | FE | Shipping handover localization | Manifest and scan handover copy translated: Bảng kê bàn giao, ĐVVC, Khu vực, Đã quét, Còn thiếu, Xác nhận bàn giao | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S19-06-05 | P0 | FE | Missing order exception microcopy | Missing order screen says: `Còn thiếu đơn. Vui lòng kiểm tra lại mã hoặc tìm trong khu vực đóng hàng.` | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S19-07-01 | P0 | FE | Returns receiving localization | Return scan UI translated: Nhận hàng hoàn, Mã đơn, Mã vận đơn, Quét hàng hoàn, Đưa vào khu hàng hoàn | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S19-07-02 | P0 | FE | Return inspection localization | Inspection fields translated: Tình trạng hàng, Nguyên vẹn, Móp hộp, Rách seal, Đã sử dụng, Hỏng/đổ vỡ | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S19-07-03 | P0 | FE | Return disposition localization | Disposition labels translated: Còn sử dụng, Không sử dụng, Cần QA kiểm tra, Chuyển vào kho, Chuyển Lab/kho hỏng | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S19-07-04 | P1 | FE | Stock count/adjustment localization | Kiểm kê, Lệch tồn, Số lượng hệ thống, Số lượng kiểm, Yêu cầu điều chỉnh, Lý do lệch, Duyệt điều chỉnh | `42_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1.md` |
| S19-08-01 | P0 | FE | Purchase order localization | PO screens translated: Đơn mua hàng, Nhà cung cấp, Số lượng mua, Đơn giá, Gửi duyệt, Duyệt PO | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S19-08-02 | P0 | FE | Goods receiving localization | Receiving UI translated: Phiếu nhận hàng, Chứng từ giao hàng, Số lượng nhận, Bao bì, Lô, Hạn sử dụng | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S19-08-03 | P0 | FE | Inbound QC localization | Inbound QC UI translated: QC hàng nhập, Đạt QC, Không đạt QC, Giữ kiểm, Đạt một phần, Trả NCC | `45_ERP_Coding_Task_Board_Sprint4_Purchase_Inbound_QC_Core_MyPham_v1.md` |
| S19-09-01 | P1 | FE | Master data localization | Master data screens translated: Mã hàng, Sản phẩm, Nguyên liệu, Kho, Vị trí, Nhà cung cấp, Khách hàng, Đơn vị tính | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S19-09-02 | P1 | FE | Inventory localization | Inventory screens translated: Tồn kho, Tồn khả dụng, Tồn đã giữ, Stock movement -> Biến động tồn, Batch -> Lô | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S19-09-03 | P1 | FE | QC/batch localization | Batch/QC status and screen labels translated: Lô, Hạn sử dụng, Trạng thái QC, Đang giữ kiểm, Đạt, Không đạt | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S19-10-01 | P1 | FE | Auth/login localization | Login page, session expired, permission denied, logout confirmation translated | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S19-10-02 | P1 | FE | Audit log localization | Audit panel labels translated while event codes remain English; user-facing action names are Vietnamese | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S19-10-03 | P1 | FE | Attachment UI localization | File upload labels translated: Tệp đính kèm, Tải lên, Xóa tệp, Ảnh chứng từ, COA/MSDS, Bằng chứng QC | `23_ERP_Integration_Spec_Phase1_MyPham_v1.md` |
| S19-11-01 | P0 | QA/FE | Add unit tests for label mapping | Status/error/formatter tests cover common ERP states and error codes | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S19-11-02 | P0 | QA/FE | Add smoke test for Vietnamese app shell | App shell renders Vietnamese menu and title labels | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S19-11-03 | P0 | QA/FE | Add smoke test for warehouse screens | Warehouse Daily Board, handover, returns, receiving show Vietnamese labels | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S19-11-04 | P1 | QA/FE | Add hardcoded English detection checklist | PR checklist or test helper flags common hardcoded English UI strings in business modules | `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md` |
| S19-12-01 | P0 | FE/PM | Update README current UI language | README states UI default is Vietnamese-first; API/DB codes remain English | `32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` |
| S19-12-02 | P0 | PM | Sprint 19 changelog | Changelog created with delivered tasks, CI evidence, known issues, release tag recommendation | `27_ERP_GoLive_Runbook_Hypercare_Phase1_MyPham_v1.md` |

---

## 10. Vietnamese Glossary - Core Modules

| Technical key | Vietnamese label |
|---|---|
| Dashboard | Tổng quan |
| Master Data | Dữ liệu gốc |
| Inventory | Kho hàng |
| Warehouse | Kho |
| Location | Vị trí |
| Sales | Bán hàng |
| Sales Order | Đơn bán hàng |
| Shipping | Giao hàng |
| Carrier | ĐVVC |
| Carrier Manifest | Bảng kê bàn giao ĐVVC |
| Returns | Hàng hoàn |
| Purchase | Mua hàng |
| Purchase Order | Đơn mua hàng |
| Receiving | Nhận hàng / Nhập kho |
| Inbound QC | QC hàng nhập |
| Quality Control | Kiểm chất lượng |
| Finance | Tài chính |
| Settings | Cài đặt |
| Audit Log | Nhật ký thao tác |
| Approval | Phê duyệt |

---

## 11. Vietnamese Glossary - Actions

| Technical action | Vietnamese label |
|---|---|
| Create | Tạo mới |
| Save | Lưu |
| Submit | Gửi duyệt |
| Approve | Duyệt |
| Reject | Từ chối |
| Cancel | Hủy |
| Confirm | Xác nhận |
| Close | Đóng |
| Reopen | Mở lại |
| Refresh | Làm mới |
| Export | Xuất dữ liệu |
| Import | Nhập dữ liệu |
| Scan | Quét mã |
| Reserve stock | Giữ tồn |
| Release reservation | Hủy giữ tồn |
| Pick | Soạn hàng |
| Pack | Đóng hàng |
| Handover | Bàn giao |
| Receive return | Nhận hàng hoàn |
| Inspect | Kiểm tra |
| Putaway | Đưa vào kho |
| Quarantine | Chuyển giữ kiểm |
| Close shift | Đóng ca |

---

## 12. Vietnamese Glossary - Status Chips

| Status code | Vietnamese label | Risk color intent |
|---|---|---|
| DRAFT | Nháp | Neutral |
| SUBMITTED | Đã gửi | Info |
| APPROVED | Đã duyệt | Success |
| REJECTED | Từ chối | Danger |
| CANCELLED | Đã hủy | Neutral |
| CONFIRMED | Đã xác nhận | Info |
| RESERVED | Đã giữ tồn | Info |
| PICKING | Đang soạn | Info |
| PICKED | Đã soạn | Success |
| PACKED | Đã đóng hàng | Success |
| WAITING_HANDOVER | Chờ bàn giao | Warning |
| HANDED_OVER | Đã bàn giao | Success |
| DELIVERED | Đã giao | Success |
| RETURNED | Đã hoàn | Warning |
| QC_HOLD | Đang giữ kiểm | Warning |
| QC_PASS | Đạt QC | Success |
| QC_FAIL | Không đạt QC | Danger |
| QUARANTINE | Cách ly / Giữ kiểm | Warning |
| AVAILABLE | Khả dụng | Success |
| RESERVED_STOCK | Đã giữ | Info |
| DAMAGED | Hàng hỏng | Danger |
| NON_REUSABLE | Không sử dụng | Danger |
| REUSABLE | Còn sử dụng | Success |

---

## 13. Vietnamese Error Message Catalog - Initial Set

| Error code | Vietnamese message |
|---|---|
| INSUFFICIENT_STOCK | Tồn khả dụng không đủ. |
| INVALID_BATCH_STATUS | Lô này chưa đạt QC, không được phép sử dụng. |
| ORDER_NOT_PACKED | Đơn hàng chưa đóng gói, không thể bàn giao. |
| MANIFEST_ORDER_MISSING | Bảng kê còn thiếu đơn chưa quét. |
| ORDER_NOT_IN_MANIFEST | Đơn này không thuộc bảng kê bàn giao hiện tại. |
| RETURN_ALREADY_RECEIVED | Hàng hoàn này đã được tiếp nhận. |
| RETURN_NOT_INSPECTED | Hàng hoàn chưa được kiểm tra tình trạng. |
| SHIFT_HAS_UNRESOLVED_ISSUES | Ca làm còn vấn đề chưa xử lý, chưa thể đóng ca. |
| QC_REQUIRED | Hàng này cần QC trước khi nhập tồn khả dụng. |
| INVALID_UOM_CONVERSION | Quy đổi đơn vị tính không hợp lệ. |
| REQUIRED_FIELD | Vui lòng nhập thông tin bắt buộc. |
| PERMISSION_DENIED | Bạn không có quyền thực hiện thao tác này. |
| SESSION_EXPIRED | Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại. |
| VALIDATION_DECIMAL | Giá trị phải là số hợp lệ. |
| VALIDATION_POSITIVE_QTY | Số lượng phải lớn hơn 0. |

---

## 14. Warehouse Microcopy Examples

### 14.1 Warehouse Daily Board

```text
Bảng công việc kho trong ngày
Theo dõi đơn mới, soạn hàng, đóng hàng, bàn giao, hàng hoàn, kiểm kê và đóng ca.
```

### 14.2 Shipping handover

```text
Đã quét 126/128 đơn.
Còn thiếu 2 đơn.
Vui lòng kiểm tra lại mã hoặc tìm trong khu vực đóng hàng.
Chỉ được xác nhận bàn giao khi đã quét đủ đơn hợp lệ.
```

### 14.3 Return receiving

```text
Quét mã đơn hoặc mã vận đơn để tiếp nhận hàng hoàn.
Hàng hoàn chưa kiểm tra sẽ không được nhập vào tồn khả dụng.
```

### 14.4 Return inspection

```text
Kiểm tra tình trạng hàng trước khi chọn hướng xử lý.
Hàng rách seal, đã sử dụng hoặc hỏng không được đưa lại vào hàng bán.
```

### 14.5 Shift closing

```text
Không thể đóng ca vì còn vấn đề chưa xử lý.
Vui lòng kiểm tra hàng hoàn, lệch tồn, manifest thiếu đơn hoặc yêu cầu điều chỉnh tồn.
```

### 14.6 Inbound QC

```text
Nhận hàng không đồng nghĩa hàng đã khả dụng.
Chỉ hàng đạt QC mới được nhập vào tồn khả dụng.
```

---

## 15. Formatter Rules

### 15.1 Money

```text
DB/API source: decimal string or numeric string
UI: 1.250.000 ₫
Currency: VND
Locale: vi-VN
```

### 15.2 Date/time

```text
Date: 26/04/2026
Time: 14:30
Timezone: Asia/Ho_Chi_Minh
```

### 15.3 Quantity

```text
10,5 kg
1.250 g
48 chai
2 thùng
```

Do not use frontend floating point for business calculations. Formatting is display-only.

---

## 16. QA Checklist

Sprint 19 is accepted only if:

```text
- UI default language is Vietnamese.
- App shell/menu is Vietnamese.
- Major module screens are Vietnamese.
- Status chips show Vietnamese labels.
- Backend technical status values are not changed.
- API error codes map to Vietnamese messages.
- Raw error codes are hidden from normal users.
- Ant Design components use Vietnamese locale.
- VND/date/quantity/rate formatters follow file 40.
- Routes remain English technical paths.
- No major business component hardcodes English labels.
- Warehouse scan screens use short Vietnamese operational copy.
- Existing E2E smoke tests still pass.
- Added localization tests pass.
```

---

## 17. Definition of Done

A Sprint 19 task is done only when:

```text
1. User-facing labels are loaded from dictionary or centralized label mapping.
2. No backend/API/DB enum names are changed.
3. No route path is changed.
4. Status/error/validation copy is Vietnamese and operationally clear.
5. Date/money/quantity formatting uses shared formatters.
6. UI still follows Hetzner-inspired minimal ERP style.
7. Permissions and audit behavior are unaffected.
8. Tests/smoke checks pass.
9. PR notes include before/after UI evidence where relevant.
```

---

## 18. Suggested Demo Script

```text
1. Open login page and show Vietnamese login/session copy.
2. Login and show Vietnamese sidebar/topbar.
3. Open Warehouse Daily Board and show Vietnamese KPI/status labels.
4. Open Sales Order and show Vietnamese order status and actions.
5. Open Pick/Pack flow and show Vietnamese scan/action labels.
6. Open Carrier Manifest and show Vietnamese handover copy.
7. Trigger missing order exception and show Vietnamese message.
8. Open Return Receiving and Return Inspection screens.
9. Show VND/date/quantity formatting.
10. Trigger an API validation/error code and show Vietnamese mapped message.
```

---

## 19. Recommended Sprint 19 Release Tag

```text
v0.19.0-vietnamese-ui-localization
```

---

## 20. Next Sprint Recommendation

After Sprint 19, the next sprint should not add more UI copy. It should return to release hygiene and architecture hardening if not already done:

```text
Sprint 20 - Release Hygiene + API Modularization + Fallback Cleanup
```

Priority candidates:

```text
- README current status update
- Sprint 18/19 changelog alignment
- Migration apply/down/reapply CI gate
- Node.js 24 GitHub Actions compatibility
- Refactor oversized cmd/api/main.go
- Gate frontend fallback services in production mode
```
