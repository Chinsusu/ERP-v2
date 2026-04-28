# 41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Sprint:** Sprint 2 — Order Fulfillment Core  
**Trạng thái đầu vào:** Sprint 1 foundation đã xong, merge/promote lên `main`, repo local sạch.  
**Backend:** Go  
**Frontend:** React / Next.js + TypeScript  
**Database:** PostgreSQL  
**API:** REST + OpenAPI 3.1  
**Architecture:** Modular Monolith  

---

## 0. Mục tiêu Sprint 2

Sprint 2 tập trung biến nền kho/stock đã có thành luồng vận hành đơn hàng thật:

```text
Sales Order
→ Reserve Stock
→ Pick
→ Pack
→ Carrier Manifest
→ Scan Handover ĐVVC
→ Warehouse Daily Board cập nhật dữ liệu thật
```

Cuối Sprint 2 phải demo được một vòng đời đơn hàng từ lúc tạo đơn đến lúc bàn giao cho đơn vị vận chuyển, gồm cả tình huống thiếu đơn khi quét bàn giao.

---

## 1. Tại sao Sprint 2 làm Order Fulfillment Core

Workflow kho thực tế đang có nhịp:

```text
Tiếp nhận đơn hàng trong ngày
→ thực hiện xuất/nhập theo nội quy
→ soạn hàng và đóng gói
→ sắp xếp/tối ưu vị trí kho
→ kiểm kê cuối ngày
→ đối soát số liệu và báo cáo quản lý
→ kết thúc ca
```

Quy trình bàn giao ĐVVC thực tế cũng có các bước:

```text
Phân chia khu vực để hàng
→ để theo thùng/rổ
→ đối chiếu số lượng đơn dựa trên bảng
→ lấy hàng và quét mã trực tiếp
→ nếu đủ đơn thì ký xác nhận/bàn giao
→ nếu chưa đủ thì kiểm tra mã hoặc tìm lại trong khu vực đóng hàng
```

Vì vậy, sau khi đã có unit/currency, auth/RBAC, master data, stock ledger, batch/QC, receiving và Warehouse Daily Board foundation, điểm hợp lý nhất là xây luồng **order-to-handover**.

---

## 2. Sprint 1 Closure trước khi mở Sprint 2

Trước khi tạo branch Sprint 2, thực hiện checkpoint:

```bash
git checkout main
git status
git pull
```

Nếu repo sạch và CI pass, tạo tag foundation:

```bash
git tag v0.1.0-foundation
git push origin v0.1.0-foundation
```

Sau đó tạo branch Sprint 2:

```bash
git checkout -b sprint/2-order-fulfillment-core
```

Nếu có bug Sprint 1 nghiêm trọng, không mở task extension lung tung. Tạo hotfix riêng:

```bash
git checkout -b hotfix/s1-<short-description>
```

---

## 3. Sprint 2 In-scope / Out-of-scope

### 3.1. In-scope

```text
- Sales order model/state/API/UI cơ bản.
- Confirm đơn và reserve stock.
- Không cho reserve khi thiếu available hoặc batch QC chưa PASS.
- Pick task generation.
- Pick confirmation, có exception khi sai SKU/batch/location.
- Packing task board và packing station.
- Carrier master hardening.
- Carrier manifest theo ĐVVC/khu/thùng/rổ.
- Scan verify handover.
- Missing order exception.
- Confirm handover.
- Warehouse Daily Board cập nhật trạng thái order fulfillment thật.
- Audit log cho action nhạy cảm.
- E2E test order-to-handover.
```

### 3.2. Out-of-scope

```text
- COD reconciliation đầy đủ.
- Return/hàng hoàn đầy đủ.
- Marketplace realtime integration.
- Shipping label API production với tất cả ĐVVC.
- Full finance posting.
- Loyalty/CRM/KOL.
- Mobile app riêng.
- Accounting tax invoice.
```

Các phần out-of-scope chỉ làm placeholder nếu cần để không chặn flow chính.

---

## 4. Rule dùng task board này

Mỗi task bên dưới có đúng **1 Primary Ref**. Dev/QA/PM mở file đó trước khi làm task. Nếu cần tài liệu phụ, tra theo `32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md`.

```text
1 task = 1 output rõ ràng.
1 task = 1 Primary Ref chính.
Không code business logic nếu state/status/rule chưa rõ.
Không update tồn kho trực tiếp ngoài stock ledger/reservation service.
Không cho batch QC HOLD/FAIL đi vào reserve/pick/pack.
Không xác nhận handover nếu manifest chưa đủ đơn hợp lệ.
Không bỏ audit log cho action nhạy cảm.
Không làm UI lệch style Hetzner-inspired đã chốt.
```

---

## 5. Sprint 2 Definition of Done

Sprint 2 chỉ được coi là xong khi:

```text
- Tạo được sales order.
- Confirm đơn thành công.
- Confirm đơn giữ tồn thành công.
- Không cho giữ tồn nếu thiếu available.
- Không cho giữ tồn/pick batch QC HOLD/FAIL.
- Sinh được pick task từ đơn reserved.
- Xác nhận pick được.
- Xác nhận pack được.
- Tạo được carrier manifest.
- Quét mã đơn/vận đơn trước bàn giao.
- Nếu thiếu đơn, hệ thống hiện missing exception.
- Nếu đủ đơn hợp lệ, hệ thống cho confirm handover.
- Warehouse Daily Board hiện số liệu từ trạng thái thật.
- Audit log có đủ confirm/reserve/pick/pack/handover.
- E2E test order-to-handover pass.
- CI pass.
```

---

## 6. State Machine Sprint 2

### 6.1. Sales Order State

```text
DRAFT
→ CONFIRMED
→ RESERVED
→ PICKING
→ PICKED
→ PACKING
→ PACKED
→ WAITING_HANDOVER
→ HANDED_OVER
→ CLOSED
```

Exception states:

```text
CANCELLED
RESERVATION_FAILED
PICK_EXCEPTION
PACK_EXCEPTION
HANDOVER_EXCEPTION
```

### 6.2. Pick Task State

```text
CREATED
→ ASSIGNED
→ IN_PROGRESS
→ COMPLETED
```

Exception:

```text
MISSING_STOCK
WRONG_SKU
WRONG_BATCH
WRONG_LOCATION
CANCELLED
```

### 6.3. Pack Task State

```text
CREATED
→ IN_PROGRESS
→ PACKED
```

Exception:

```text
PACK_EXCEPTION
CANCELLED
```

### 6.4. Carrier Manifest State

```text
DRAFT
→ READY_TO_SCAN
→ SCANNING
→ COMPLETED
→ HANDED_OVER
```

Exception:

```text
MISSING_ORDERS
EXTRA_ORDER_SCANNED
WRONG_CARRIER
CANCELLED
```

---

## 7. Backlog Sprint 2

### 7.1. Sprint Setup / Release Checkpoint

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-00-01 | P0 | Tech Lead / DevOps | Tag Sprint 1 foundation | Tag `v0.1.0-foundation` được push, CI pass, repo clean | `37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md` |
| S2-00-02 | P0 | Tech Lead | Tạo branch Sprint 2 | Branch `sprint/2-order-fulfillment-core` được tạo từ `main` | `38_ERP_Workspace_Repository_Structure_Standards_Phase1_MyPham_v1.md` |
| S2-00-03 | P1 | PM / Tech Lead | Update README/changelog Sprint 2 | README ghi sprint goal, branch, demo flow, task board mới | `32_ERP_Master_Document_Index_Traceability_Handoff_Phase1_MyPham_v1.md` |

---

### 7.2. Sales Order Core

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-01-01 | P0 | BE | Sales order DB model + migration | Có `sales_orders`, `sales_order_lines`; state, customer, channel, amount, currency; migration up/down pass | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S2-01-02 | P0 | BE | Sales order domain state machine | Domain không cho nhảy trạng thái sai; unit test state transition pass | `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md` |
| S2-01-03 | P0 | BE | Sales order API contract | OpenAPI có list/detail/create/update/confirm/cancel; generated client không lỗi | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S2-01-04 | P0 | BE | Sales order application service | Tạo đơn, cập nhật line, confirm, cancel; transaction rõ; error code chuẩn | `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md` |
| S2-01-05 | P0 | FE | Sales order UI list/detail/create | Có list/filter/status chip, detail, create form, line item editor | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S2-01-06 | P1 | QA | Sales order API/UI smoke test | Test create/edit/confirm/cancel; test required field; test permission deny | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.3. Stock Reservation

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-02-01 | P0 | BE | Reservation DB/model hardening | Có reservation records gắn sales order line/SKU/batch/location; status active/released/consumed | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S2-02-02 | P0 | BE | Reserve stock on confirm | Confirm đơn reserve được stock nếu available đủ; transaction atomic | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S2-02-03 | P0 | BE | Prevent oversell | Nếu requested qty > available qty, trả lỗi `INSUFFICIENT_STOCK` | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S2-02-04 | P0 | BE | QC block for reservation | Batch QC HOLD/FAIL không được reserve; trả lỗi `BATCH_NOT_SELLABLE` | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S2-02-05 | P0 | BE | Cancel/unreserve | Cancel order release reservation, available stock tăng lại đúng | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S2-02-06 | P1 | BE | Reservation audit log | Confirm/reserve/release ghi audit với actor, before/after, reason | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S2-02-07 | P1 | QA | Reservation test suite | Test đủ tồn, thiếu tồn, QC hold, cancel/unreserve, concurrent confirm | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.4. Pick Task

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-03-01 | P0 | BE | Pick task DB/model | Có `pick_tasks`, `pick_task_lines`, state, assigned user, source order | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S2-03-02 | P0 | BE | Generate pick task from reserved order | Reserved order sinh pick task theo SKU/batch/location | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S2-03-03 | P0 | BE | Picking API actions | API start/confirm/exception cho pick task; idempotent action | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S2-03-04 | P0 | FE | Picking UI scan-first | Màn pick tối ưu cho scan, hiển thị SKU/batch/location/qty rõ | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` |
| S2-03-05 | P1 | BE | Picking exception rules | Wrong SKU/batch/location tạo exception, không cho chuyển `PICKED` | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S2-03-06 | P1 | QA | Picking test cases | Test pick đúng, sai SKU, sai batch, thiếu location, permission | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.5. Pack Task / Packing Station

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-04-01 | P0 | BE | Pack task DB/model | Có `pack_tasks`, `pack_task_lines`; link sales order/pick task | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S2-04-02 | P0 | BE | Generate pack task after pick | Pick completed sinh pack task, order chuyển `PACKING` | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S2-04-03 | P0 | BE | Packing API actions | API start/confirm/exception; confirm pack chuyển order `PACKED` | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S2-04-04 | P0 | FE | Packing station UI | Màn packing hiển thị đơn, SKU, qty, scan/confirm, note, exception | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S2-04-05 | P1 | BE | Packing exception handling | Thiếu hàng/sai SKU/sai batch không cho packed, tạo exception log | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S2-04-06 | P1 | QA | Packing test cases | Test pack đúng, pack thiếu, pack sai line, pack không đủ quyền | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.6. Carrier / Manifest / Handover Zone

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-05-01 | P0 | BE | Carrier master hardening | Carrier có code, name, handover zone, active status, SLA placeholder | `23_ERP_Integration_Spec_Phase1_MyPham_v1.md` |
| S2-05-02 | P0 | BE | Carrier manifest DB/model | Có `carrier_manifests`, `carrier_manifest_orders`, trạng thái manifest | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S2-05-03 | P0 | BE | Manifest API | Create manifest, add/remove packed orders, ready-to-scan, cancel | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S2-05-04 | P0 | BE | Handover zone/bin model | Gắn manifest với khu vực/thùng/rổ để bàn giao | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S2-05-05 | P0 | FE | Manifest UI | List/detail manifest, order count, carrier, zone, status, actions | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S2-05-06 | P1 | QA | Manifest test cases | Test add order chưa packed bị chặn; add đúng carrier; cancel manifest | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.7. Scan Handover ĐVVC

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-06-01 | P0 | BE | Scan verify handover API | Scan order/tracking no; kiểm tra thuộc manifest, đã packed, đúng carrier | `23_ERP_Integration_Spec_Phase1_MyPham_v1.md` |
| S2-06-02 | P0 | BE | Scan event log | Mỗi lần scan lưu result, actor, device/source, manifest, order/tracking no | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S2-06-03 | P0 | BE | Missing order exception | Manifest chưa đủ scan thì hiện danh sách thiếu, không cho handover | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S2-06-04 | P0 | BE | Confirm handover | Chỉ confirm khi đủ đơn hợp lệ; order chuyển `HANDED_OVER`; manifest chuyển `HANDED_OVER` | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S2-06-05 | P0 | FE | Handover scan UI | Màn scan lớn, số tổng/đã quét/còn thiếu, danh sách lỗi, nút confirm handover | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S2-06-06 | P1 | QA | Handover E2E negative tests | Test scan sai carrier, scan đơn chưa packed, scan thừa, thiếu đơn, duplicate scan | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.8. Warehouse Daily Board Integration

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-07-01 | P0 | BE | Daily Board fulfillment metrics API | API trả số đơn new/reserved/picking/packed/waiting handover/missing/handover | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S2-07-02 | P0 | FE | Daily Board fulfillment UI | Board hiển thị trạng thái fulfillment thật, filter theo ngày/kho/ca/carrier | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S2-07-03 | P1 | BE | Board drill-down links | Bấm KPI mở list đơn/task tương ứng | `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md` |
| S2-07-04 | P1 | QA | Daily Board data consistency test | Số trên board khớp sales order/pick/pack/manifest state | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |

---

### 7.9. Test / Regression / Demo

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S2-08-01 | P0 | QA / BE / FE | E2E order-to-handover happy path | Test tạo đơn → confirm → reserve → pick → pack → manifest → scan → handover pass | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S2-08-02 | P0 | QA | Permission/audit regression | Test sales/kho/admin role; audit đủ confirm/reserve/pick/pack/handover | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S2-08-03 | P1 | QA / BE | Test data Sprint 2 | Seed customers, SKUs, batches, stock, 20 orders, 3 carriers, 2 manifests | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S2-08-04 | P1 | Tech Lead | Sprint 2 demo script | Script demo rõ từng bước, gồm cả missing order exception | `34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md` |
| S2-08-05 | P1 | Tech Lead / PM | Sprint 2 release note | Ghi feature, bug, known issue, next sprint recommendation | `29_ERP_Operations_Support_Model_Phase1_MyPham_v1.md` |

---

## 8. Recommended Implementation Order

Không code UI trước. Thứ tự chuẩn:

```text
1. S2-00 release checkpoint + branch.
2. Sales order DB/state/API.
3. Reservation service + QC/available stock guards.
4. Pick task model/API.
5. Pack task model/API.
6. Carrier manifest model/API.
7. Scan handover API + event log + missing exception.
8. Frontend sales order/pick/pack/manifest/handover screens.
9. Warehouse Daily Board integration.
10. E2E/regression/demo.
```

Nếu sprint bị quá tải, giữ P0 và cắt P1/P2. Không cắt các guardrail sau:

```text
- Không cắt QC block.
- Không cắt stock reservation correctness.
- Không cắt scan event log.
- Không cắt missing order exception.
- Không cắt audit log cho action nhạy cảm.
```

---

## 9. API Endpoint Draft cho Sprint 2

> API cuối cùng phải theo `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`.

### 9.1. Sales Orders

```text
GET    /api/v1/sales-orders
POST   /api/v1/sales-orders
GET    /api/v1/sales-orders/{id}
PATCH  /api/v1/sales-orders/{id}
POST   /api/v1/sales-orders/{id}/confirm
POST   /api/v1/sales-orders/{id}/cancel
```

### 9.2. Pick Tasks

```text
GET    /api/v1/pick-tasks
GET    /api/v1/pick-tasks/{id}
POST   /api/v1/pick-tasks/{id}/start
POST   /api/v1/pick-tasks/{id}/confirm-line
POST   /api/v1/pick-tasks/{id}/complete
POST   /api/v1/pick-tasks/{id}/exception
```

### 9.3. Pack Tasks

```text
GET    /api/v1/pack-tasks
GET    /api/v1/pack-tasks/{id}
POST   /api/v1/pack-tasks/{id}/start
POST   /api/v1/pack-tasks/{id}/confirm
POST   /api/v1/pack-tasks/{id}/exception
```

### 9.4. Carrier Manifests / Handover

```text
GET    /api/v1/carrier-manifests
POST   /api/v1/carrier-manifests
GET    /api/v1/carrier-manifests/{id}
POST   /api/v1/carrier-manifests/{id}/add-orders
POST   /api/v1/carrier-manifests/{id}/ready-to-scan
POST   /api/v1/carrier-manifests/{id}/scan
POST   /api/v1/carrier-manifests/{id}/confirm-handover
POST   /api/v1/carrier-manifests/{id}/cancel
```

### 9.5. Warehouse Daily Board

```text
GET /api/v1/warehouse/daily-board?warehouse_id=&date=&shift=&carrier_id=
```

---

## 10. Database Entities Draft cho Sprint 2

> DB cuối cùng phải theo `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`.

```text
sales_orders
sales_order_lines
stock_reservations
pick_tasks
pick_task_lines
pack_tasks
pack_task_lines
carriers
handover_zones
carrier_manifests
carrier_manifest_orders
handover_scan_events
fulfillment_exceptions
```

### 10.1. Field bắt buộc cần nhớ

```text
- status
- version / optimistic lock nếu cần
- created_by / updated_by
- created_at / updated_at
- audit correlation id nếu có
- idempotency key cho action endpoint nhạy cảm
- currency_code = VND với money fields
- qty dùng decimal string/API và numeric(18,6)/DB
```

---

## 11. Critical Error Codes Sprint 2

```text
INSUFFICIENT_STOCK
BATCH_NOT_SELLABLE
ORDER_INVALID_STATE
ORDER_ALREADY_RESERVED
RESERVATION_NOT_FOUND
PICK_TASK_INVALID_STATE
PICK_WRONG_SKU
PICK_WRONG_BATCH
PICK_WRONG_LOCATION
PACK_TASK_INVALID_STATE
MANIFEST_INVALID_STATE
ORDER_NOT_PACKED
ORDER_NOT_IN_MANIFEST
ORDER_ALREADY_SCANNED
WRONG_CARRIER
MISSING_ORDERS
HANDOVER_NOT_ALLOWED
PERMISSION_DENIED
IDEMPOTENCY_CONFLICT
```

---

## 12. Test Data gợi ý cho Sprint 2

```text
Customers:
- CUST-RETAIL-001
- CUST-B2B-001

Carriers:
- GHN
- GHTK
- INTERNAL

Warehouses:
- WH-MAIN

Locations:
- A1, A2, A3
- PACKING-ZONE
- HANDOVER-GHN
- HANDOVER-GHTK

SKU:
- SERUM-VITC-30ML
- TONER-BHA-150ML
- CLEANSER-GEL-100ML

Batches:
- SVC-260401-01 / QC PASS
- SVC-260401-02 / QC HOLD
- TON-260401-01 / QC PASS

Orders:
- 10 happy path orders
- 3 insufficient stock orders
- 2 QC hold batch orders
- 3 packed orders for manifest
- 2 missing-order handover cases
```

---

## 13. Sprint 2 Demo Script

```text
1. Login bằng user Sales.
2. Tạo sales order cho khách CUST-RETAIL-001.
3. Thêm SKU SERUM-VITC-30ML, qty 2.
4. Confirm order.
5. Hệ thống reserve stock thành công.
6. Login bằng user Warehouse.
7. Mở Pick Task, quét/xác nhận đúng SKU/batch/location.
8. Complete pick.
9. Mở Packing Station, xác nhận đóng gói.
10. Tạo manifest GHN cho các đơn packed.
11. Chuyển manifest sang READY_TO_SCAN.
12. Quét đủ tất cả đơn.
13. Confirm handover.
14. Mở Warehouse Daily Board thấy số liệu chuyển trạng thái.
15. Demo case thiếu đơn: tạo manifest khác, quét thiếu 1 đơn, hệ thống không cho handover và hiển thị danh sách thiếu.
16. Mở audit log xem confirm/reserve/pick/pack/handover.
```

---

## 14. Sprint 2 PR / Branch Rule

Branch task:

```text
feature/S2-01-01-sales-order-db-model
feature/S2-02-02-reserve-stock-on-confirm
feature/S2-06-01-scan-verify-handover-api
```

Commit format:

```text
S2-02-02: reserve stock on sales order confirm
S2-06-03: block handover when manifest has missing orders
```

PR phải có:

```text
- Task ID
- Primary Ref
- What changed
- Test evidence
- Migration impact nếu có
- OpenAPI diff nếu có
- Screenshot/video nếu là UI
- Risk/rollback note nếu là action nhạy cảm
```

---

## 15. Sprint 2 Risk Notes

| Risk | Dấu hiệu | Cách xử lý |
|---|---|---|
| Oversell do concurrent confirm | Hai đơn cùng reserve một batch | Transaction + row lock/advisory lock theo stock balance/reservation |
| Batch QC HOLD vẫn bị bán | Confirm order không check QC status | Guard ở reservation service, test bắt buộc |
| Manifest scan thiếu nhưng vẫn handover | UI/API cho confirm quá sớm | Backend chặn hard rule, không tin UI |
| Số Daily Board sai | Board query theo trạng thái không đồng nhất | Board service dùng source status rõ, test consistency |
| Duplicate scan | Scan lại cùng order nhiều lần | Idempotent scan event, trả result `ALREADY_SCANNED` |
| Mỗi module tự sửa order status | Boundary lỏng | Chỉ application service owner được transition order |

---

## 16. Sau Sprint 2 nên làm gì

Nếu Sprint 2 pass, sprint kế tiếp hợp lý nhất:

```text
Sprint 3 — Returns / Hàng hoàn + End-of-day Reconciliation Hardening
```

Lý do: sau khi đơn đã đi được đến ĐVVC, chiều ngược là hàng hoàn. Workflow thực tế đã có nhánh nhận hàng từ shipper, đưa vào khu hàng hoàn, quét hàng hoàn, kiểm tình trạng, phân loại còn dùng/không dùng, rồi nhập kho hoặc chuyển lab. Nếu không build sớm, kho sẽ đứt dữ liệu chiều hoàn.

---

## 17. Final Sprint 2 Checklist

```text
[ ] Tag Sprint 1 foundation.
[ ] Branch Sprint 2 tạo đúng.
[ ] Sales order DB/API/UI xong.
[ ] Reservation service xong.
[ ] QC/available stock guards xong.
[ ] Pick task xong.
[ ] Pack task xong.
[ ] Carrier manifest xong.
[ ] Scan handover xong.
[ ] Missing order exception xong.
[ ] Warehouse Daily Board updated.
[ ] Audit log pass.
[ ] E2E pass.
[ ] CI pass.
[ ] Demo pass.
[ ] Release note có.
```
