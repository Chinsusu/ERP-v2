# 81_ERP_Vietnamese_UI_Glossary_Operational_Copy_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Vietnamese UI glossary and operational copy lock
Version: v1.0
Date: 2026-05-03
Status: Active glossary for Vietnamese-first UI copy

---

## 1. Purpose

This glossary locks the Vietnamese operational terms that are easy to confuse after Sprint 19 localization.

The rule remains:

```text
Technical contract: English
Business display: Vietnamese
Routes: English
Locale: vi-VN
Currency: VND
Timezone: Asia/Ho_Chi_Minh
```

Do not use this glossary to rename API routes, database values, enum values, permission keys, audit event codes, OpenAPI schemas, or backend error codes.

---

## 2. Terms To Use

| Technical term | Preferred Vietnamese UI copy | Notes |
| --- | --- | --- |
| Goods Receiving | Tiếp nhận hàng | Receiving is the physical/document receipt step. Do not imply available stock yet. |
| Warehouse Putaway | Đưa vào kho | Use after goods are accepted for storage movement. |
| Stock Inbound / Posted Stock | Nhập tồn | Use only when stock is posted into inventory records. |
| Inbound QC | QC hàng nhập | Use for receiving-side QC. |
| Quality Control | QC / Kiểm soát chất lượng | Prefer `QC / Kiểm soát chất lượng` for module/menu context. Avoid `Kiểm chất lượng`. |
| QC Hold | Đang giữ kiểm | Goods are held until QC disposition is clear. |
| QC Pass | Đạt QC | Stock can become available if other business rules pass. |
| QC Fail | Không đạt QC | Stock must not become available. |
| Available Stock | Tồn khả dụng | Stock available for reservation/picking. |
| Quarantine | Cách ly / Giữ kiểm | Use for stock that must not be consumed or sold. |
| Carrier | Đơn vị vận chuyển (ĐVVC) | Expand on first use, then `ĐVVC` is allowed. |
| Carrier Manifest | Bảng kê bàn giao ĐVVC | Use for handover documents. |
| Carrier Handover Scan | Quét bàn giao ĐVVC | Use for scan workflow. |
| Return Receiving | Tiếp nhận hàng hoàn | Return receipt from shipper/customer flow. |
| Return Inspection | Kiểm hàng hoàn | Detailed condition inspection before disposition. |
| Return Disposition | Phân loại hàng hoàn | Reusable, non-reusable, QA required, damaged/lab/hold paths. |
| Shift Closing | Đóng ca | End-of-shift reconciliation and blocker resolution. |
| End-of-Day Reconciliation | Đối soát cuối ngày | Use for day-level operations reconciliation. |

---

## 3. Terms To Avoid Or Qualify

| Avoid / risky wording | Use instead | Reason |
| --- | --- | --- |
| Nhận hàng / Nhập kho as one combined meaning | Tiếp nhận hàng, Đưa vào kho, Nhập tồn | Receipt, putaway, and stock posting are separate operational states. |
| Kiểm chất lượng | QC / Kiểm soát chất lượng | More standard and clearer for module labels. |
| ĐVVC without first expansion in a new document/screen family | Đơn vị vận chuyển (ĐVVC), then ĐVVC | Helps sales, finance, and onboarding users. |
| Hàng đã nhận is available | Hàng đã tiếp nhận, chờ QC / chờ nhập tồn | Receipt does not guarantee available stock. |
| Cách ly as a generic error state | Cách ly / Giữ kiểm | Make clear this is an inventory control state. |

---

## 4. Required Business Copy Rules

Receiving and QC:

```text
Nhận hàng không đồng nghĩa hàng đã khả dụng.
Chỉ hàng đạt QC mới được nhập vào tồn khả dụng.
```

Carrier handover:

```text
First use: Đơn vị vận chuyển (ĐVVC)
Later use: ĐVVC
Screen names: Bảng kê bàn giao ĐVVC, Quét bàn giao ĐVVC
```

Returns:

```text
Hàng hoàn chưa kiểm tra sẽ không được nhập vào tồn khả dụng.
Hàng rách seal, đã sử dụng hoặc hỏng không được đưa lại vào hàng bán.
```

Shift closing:

```text
Không thể đóng ca khi còn hàng hoàn chưa kiểm, lệch tồn chưa xử lý, manifest thiếu đơn, hoặc yêu cầu điều chỉnh tồn chưa được duyệt.
```

---

## 5. UI Label Examples

| Context | Copy |
| --- | --- |
| Receiving page title | Tiếp nhận hàng |
| Receiving document | Phiếu tiếp nhận hàng |
| Quantity received | Số lượng nhận |
| Putaway action | Đưa vào kho |
| Stock posting action | Nhập tồn |
| Inbound QC page title | QC hàng nhập |
| QC status column | Trạng thái QC |
| QC hold status | Đang giữ kiểm |
| Available stock column | Tồn khả dụng |
| Carrier manifest page title | Bảng kê bàn giao ĐVVC |
| Handover scan action | Quét bàn giao ĐVVC |
| Return receiving title | Tiếp nhận hàng hoàn |
| Return inspection title | Kiểm hàng hoàn |
| Return disposition section | Phân loại hàng hoàn |
| Shift closing title | Đóng ca |
| Reconciliation title | Đối soát cuối ngày |

---

## 6. Implementation Guardrails

```text
1. Keep backend/API/DB values in English.
2. Keep routes in English.
3. Map technical status codes to Vietnamese display labels in shared label dictionaries.
4. Do not show raw technical status or error codes to normal users.
5. Use vi-VN formatters for money, date, quantity, and rate display.
6. Prefer short operational copy over literal translation.
7. If a Vietnamese term changes, update this glossary before updating screens.
```
