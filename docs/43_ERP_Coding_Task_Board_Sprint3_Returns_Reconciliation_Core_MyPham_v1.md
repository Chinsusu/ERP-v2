# 43_ERP_Coding_Task_Board_Sprint3_Returns_Reconciliation_Core_MyPham_v1

## 1. Mục đích tài liệu

Tài liệu này là **Coding Task Board cho Sprint 3** của dự án ERP mỹ phẩm.

Sprint 3 tập trung vào 3 mảng nghiệp vụ sống còn sau khi Sprint 2 đã hoàn thành luồng outbound:

```text
Sales Order → Reserve Stock → Pick → Pack → Carrier Manifest → Scan Handover ĐVVC
```

Sprint 3 sẽ xây tiếp chiều ngược và phần đóng ca kho:

```text
Returns / Hàng hoàn
→ Return Inspection
→ Return Disposition
→ Stock Movement / Quarantine
→ Inventory Adjustment
→ Stock Count
→ End-of-Day Reconciliation
→ Shift Closing
→ Warehouse Daily Board update
```

Triết lý của Sprint 3:

```text
Sprint 2 kiểm soát hàng đi ra.
Sprint 3 kiểm soát hàng quay về và khóa sổ cuối ngày.
```

---

## 2. Workflow thực tế làm nền cho Sprint 3

Sprint 3 được bám theo workflow thực tế đã upload:

### 2.1. Công việc hằng ngày của kho

Kho đang chạy theo nhịp:

```text
Tiếp nhận đơn hàng trong ngày
→ Thực hiện xuất/nhập theo nội quy
→ Soạn hàng và đóng gói
→ Sắp xếp, tối ưu vị trí kho
→ Kiểm kê hàng tồn kho cuối ngày
→ Đối soát số liệu và báo cáo cho quản lý
→ Kết thúc ca làm
```

Ý nghĩa với Sprint 3:

```text
- Phải có stock count cuối ngày.
- Phải có reconciliation cuối ngày.
- Phải có shift closing.
- Warehouse Daily Board phải hiển thị trạng thái hàng hoàn, lệch tồn, ca chưa đóng.
```

### 2.2. Nội quy xử lý hàng hoàn

Luồng hàng hoàn thực tế:

```text
Nhận hàng từ Shipper
→ Đưa vào khu vực để hàng hoàn
→ Quét hàng hoàn
→ Quay/ghi nhận tình trạng thực tế của sản phẩm
→ Kiểm tra chi tiết bên trong
→ Nếu còn sử dụng: chuyển vào kho
→ Nếu không sử dụng: chuyển lên Lab
→ Lập phiếu nhập kho / lưu trữ chứng từ đầy đủ
```

Ý nghĩa với Sprint 3:

```text
- Hàng hoàn không được tự động nhập available stock.
- Phải có return inspection.
- Phải có disposition: reusable / non_reusable / need_qa.
- Hàng không sử dụng không được quay lại tồn bán.
- Hàng cần QA phải vào HOLD/quarantine.
```

### 2.3. Bàn giao ĐVVC và exception liên quan

Sprint 2 đã làm outbound/handover. Sprint 3 cần dùng lại dữ liệu đó để xử lý case hàng quay về:

```text
Manifest
→ Scan handover
→ Nếu đơn quay lại từ shipper thì return phải trace được order/tracking/manifest
```

### 2.4. Nguyên tắc kho không được sửa tồn trực tiếp

Mọi thay đổi tồn trong Sprint 3 phải đi qua:

```text
stock_movement
stock_adjustment_request
return_receipt
return_disposition
quarantine movement
```

Không cho sửa trực tiếp `stock_balance`.

---

## 3. Sprint 3 Scope

### 3.1. In scope

```text
1. Return reason/disposition master.
2. Return receiving backend model.
3. Return receiving API.
4. Return receiving scan UI.
5. Return inspection workflow.
6. Return disposition action.
7. Photo/attachment cho hàng hoàn.
8. Return stock movement.
9. Quarantine/HOLD cho hàng cần QA.
10. Inventory adjustment request.
11. Stock count session.
12. Variance approval.
13. Shift closing model.
14. End-of-day reconciliation.
15. Prevent closing khi còn issue.
16. Warehouse Daily Board update.
17. E2E/regression test cho return + reconciliation + closing.
```

### 3.2. Out of scope

```text
1. Refund tiền nâng cao.
2. Kế toán hoàn tiền đầy đủ.
3. COD reconciliation nâng cao.
4. Claim/complaint khách hàng nâng cao.
5. Loyalty point reversal.
6. Marketplace return API sync tự động.
7. Finance posting sâu.
8. Mobile app riêng.
```

Các phần out-of-scope sẽ được đưa sang Sprint sau hoặc Phase 2.

---

## 4. Release checkpoint trước khi mở Sprint 3

Trước khi tạo branch Sprint 3, phải đóng Sprint 2:

```bash
git checkout main
git status
git pull

make ci-check
make smoke-test
make smoke-dev

git tag v0.2.0-order-fulfillment-core
git push origin v0.2.0-order-fulfillment-core
```

Task branch Sprint 3 tạo theo từng task ở mục 8.1, không tạo branch tổng nếu không có nhu cầu release riêng.

### Acceptance của checkpoint

```text
- Repo sạch.
- Local gate `make ci-check` pass.
- Smoke test local `make smoke-test` pass.
- Dev smoke `make smoke-dev` pass nếu dev server đang available.
- Sprint 2 demo pass.
- Tag v0.2.0-order-fulfillment-core đã push.
- Changelog đã ghi rõ outbound fulfillment đã xong.
- Sprint 3 task branch / PR / manual review flow đã chốt.
```

---

## 5. Demo mục tiêu cuối Sprint 3

Cuối Sprint 3, demo phải chạy được vòng sau:

```text
1. Một đơn đã HandedOver hoặc Delivered phát sinh hàng hoàn.
2. Nhân viên kho quét mã đơn/mã vận đơn.
3. Hệ thống tạo return receiving record.
4. Return record link được order, tracking, SKU, batch nếu có.
5. Nhân viên kiểm tình trạng hàng.
6. Nhân viên chọn disposition:
   - Còn sử dụng
   - Không sử dụng
   - Cần QA
7. Nếu còn sử dụng:
   - Tạo stock movement nhập lại kho.
   - Có thể đưa vào available nếu rule cho phép.
8. Nếu không sử dụng:
   - Chuyển Lab/kho hỏng.
   - Không vào available.
9. Nếu cần QA:
   - Chuyển quarantine/HOLD.
   - Không cho bán.
10. Kho mở stock count session cuối ngày.
11. Nếu có lệch tồn:
   - Tạo adjustment request.
   - Phải có lý do và duyệt.
12. End-of-day reconciliation đối soát:
   - đơn đã xử lý
   - hàng hoàn
   - stock movement
   - adjustment pending
   - manifest issue
13. Nếu còn issue chưa xử lý:
   - không cho đóng ca.
14. Nếu sạch:
   - đóng ca thành công.
15. Warehouse Daily Board cập nhật số liệu thật.
16. Audit log ghi đủ.
```

---

## 6. Sprint 3 Task Board

> Quy tắc: mỗi task có đúng **1 Primary Ref** để team biết tài liệu nguồn chính cần đọc trước khi code.

| Task ID | Priority | Owner | Task | Output / Acceptance | Primary Ref |
|---|---:|---|---|---|---|
| S3-00-01 | P0 | PM/Tech Lead | Sprint 2 release checkpoint | Tag `v0.2.0-order-fulfillment-core`, release note, CI pass, demo pass | `41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md` |
| S3-00-02 | P0 | PM/BA | Sprint 3 kickoff / scope lock | Kickoff note, scope in/out, owner từng epic, demo target đã chốt | `34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md` |
| S3-01-01 | P0 | BE | Return reason & disposition master | Có danh mục lý do hoàn, tình trạng hàng, kết luận xử lý; seed data cơ bản | `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` |
| S3-01-02 | P0 | BE | Return receiving database model | Có bảng return header/line, link order/tracking/SKU/batch/customer, migration pass | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S3-01-03 | P0 | BE | Return receiving API | API tạo phiếu nhận hàng hoàn, scan mã đơn/vận đơn, validate order status | `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` |
| S3-01-04 | P0 | FE | Return receiving scan UI | Màn hình scan-first cho nhân viên kho nhận hàng hoàn, có trạng thái scan pass/fail | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S3-01-05 | P1 | QA | Return receiving API test | Test scan order/tracking hợp lệ, sai mã, trùng return, order chưa handed over | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-02-01 | P0 | BE | Return inspection workflow | Kiểm tình trạng: nguyên vẹn, móp hộp, rách seal, đã dùng, hỏng, thiếu phụ kiện | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S3-02-02 | P0 | BE | Return disposition action | Còn dùng → putaway; không dùng → Lab/kho hỏng; cần QA → quarantine/HOLD | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S3-02-03 | P1 | FE | Return inspection UI | Form kiểm hàng hoàn, radio disposition, ghi chú, ảnh, trạng thái rõ ràng | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S3-02-04 | P1 | BE/FE | Return photo/attachment | Upload ảnh/video tình trạng hàng hoàn, link với return inspection, lưu audit | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S3-02-05 | P1 | QA | Return inspection test | Test đủ trạng thái tình trạng hàng và disposition path | `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md` |
| S3-03-01 | P0 | BE | Return stock movement | Chỉ hàng đủ điều kiện mới tạo movement vào kho; hàng lỗi không vào available | `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` |
| S3-03-02 | P0 | BE | Quarantine return stock | Hàng cần QA vào trạng thái HOLD/quarantine, không được reserve/pick/sell | `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md` |
| S3-03-03 | P0 | BE | Prevent direct stock mutation | Return/reconciliation/adjustment không được update trực tiếp stock balance | `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md` |
| S3-03-04 | P1 | BE | Return movement audit | Audit đầy đủ: nhận hàng, kiểm hàng, disposition, movement, quarantine | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S3-03-05 | P1 | QA | Return stock movement regression | Test available/reserved/physical/quarantine sau từng disposition | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-04-01 | P0 | BE | Inventory adjustment request | Tạo yêu cầu điều chỉnh tồn khi kiểm kê lệch, không sửa tồn trực tiếp | `30_ERP_Data_Governance_Change_Control_Phase1_MyPham_v1.md` |
| S3-04-02 | P0 | BE | Stock count session | Mở phiên kiểm kê theo kho/vị trí/SKU/batch, có counted qty và system qty | `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` |
| S3-04-03 | P0 | BE | Variance approval | Lệch tồn phải có lý do, người duyệt, audit log, movement điều chỉnh | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S3-04-04 | P1 | FE | Stock count UI | Màn hình kiểm kê theo kho/vị trí/SKU/batch, nhập số đếm, thấy variance | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S3-04-05 | P1 | FE | Adjustment approval UI | Màn hình duyệt adjustment: lý do, trước/sau, người tạo, audit | `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` |
| S3-04-06 | P1 | QA | Stock count/adjustment test | Test lệch tăng, lệch giảm, duyệt, reject, audit, stock movement | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-05-01 | P0 | BE | Shift closing model | Có ca làm, số đơn xử lý, hàng hoàn, movement, pending issue, trạng thái đóng ca | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S3-05-02 | P0 | BE | End-of-day reconciliation | Đối soát đơn, bàn giao, hàng hoàn, stock movement, kiểm kê cuối ngày | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S3-05-03 | P0 | BE | Prevent closing with unresolved issue | Không cho đóng ca nếu còn return chưa kiểm, manifest thiếu đơn, adjustment chưa xử lý | `28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md` |
| S3-05-04 | P1 | FE | Shift closing UI | Màn hình đóng ca có checklist: returns, manifests, stock count, adjustments, issues | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S3-05-05 | P1 | QA | Shift closing test | Test đóng ca sạch, đóng ca bị chặn, reopen/exception theo quyền | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-06-01 | P0 | BE/FE | Warehouse Daily Board update | Board hiển thị return pending, QA hold, adjustment pending, closing status | `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md` |
| S3-06-02 | P1 | FE | Daily Board UX hardening | Card/cột cảnh báo dùng style minimal đỏ/xám, link drill-down tới return/closing | `39_ERP_UI_Template_Hetzner_Minimal_Style_Phase1_MyPham_v1.md` |
| S3-06-03 | P1 | QA | Daily Board data test | Test board lấy số liệu thật từ return, stock movement, adjustment, closing | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-07-01 | P0 | QA | Return E2E test | Test handed-over order → return scan → inspect → disposition → stock movement | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-07-02 | P0 | QA | Shift closing E2E test | Test daily board → stock count → reconciliation → close shift | `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md` |
| S3-07-03 | P1 | QA/Security | Permission/audit regression | Test role warehouse/QA/manager/admin; action nhạy cảm có audit | `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` |
| S3-07-04 | P1 | QA | Decimal/UOM regression | Test qty, base qty, money/rate không dùng float, hiển thị đúng format | `40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1.md` |
| S3-08-01 | P1 | DevOps | Sprint 3 release pipeline check | CI/CD pass cho BE/FE/OpenAPI/migration/e2e trước merge main | `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md` |
| S3-08-02 | P1 | PM/Tech Lead | Sprint 3 release note | Release note ghi rõ returns, adjustment, shift closing, known issues | `29_ERP_Operations_Support_Model_Phase1_MyPham_v1.md` |

---

## 7. Dependency Map

```text
S3-00-01 → S3-00-02
  ↓
S3-01-01
  ↓
S3-01-02 → S3-01-03 → S3-01-04 → S3-01-05
  ↓
S3-02-01 → S3-02-02 → S3-02-03
             ↓
          S3-02-04 → S3-02-05
  ↓
S3-03-03
  ↓
S3-03-01 → S3-03-02 → S3-03-04 → S3-03-05
  ↓
S3-04-01 → S3-04-02 → S3-04-03 → S3-04-04 / S3-04-05 → S3-04-06
  ↓
S3-05-01 → S3-05-02 → S3-05-03 → S3-05-04 → S3-05-05
  ↓
S3-06-01 → S3-06-02 → S3-06-03
  ↓
S3-07-01 / S3-07-02 / S3-07-03 / S3-07-04
  ↓
S3-08-01 / S3-08-02
```

`S3-03-03 Prevent direct stock mutation` là guardrail bắt buộc trước mọi task tạo movement, adjustment hoặc reconciliation.

---

## 8. Thứ tự code khuyến nghị

```text
1. Return reason/disposition master
2. Return receiving DB model
3. Return receiving API
4. Return scan UI
5. Return inspection workflow
6. Return disposition action
7. Return stock movement/quarantine
8. Inventory adjustment request
9. Stock count session
10. Variance approval
11. Shift closing model
12. End-of-day reconciliation
13. Warehouse Daily Board update
14. E2E + regression
15. Release note
```

Không làm UI đẹp trước khi backend state/stock movement đúng.

---

## 8.1. Sprint 3 PR / Branch Rule

Branch task:

```text
feature/S3-01-01-return-reason-disposition-master
feature/S3-01-02-return-receiving-db-model
feature/S3-05-03-block-shift-closing-unresolved-issue
```

Commit format:

```text
S3-01-01: add return reason disposition master
S3-05-03: block shift closing with unresolved issues
```

PR target mặc định là `main`. Chỉ dùng `sprint/3-returns-reconciliation-core` nếu team chốt cần branch tích hợp riêng.

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

Review/merge rule:

```text
- Tự review local trước khi mở PR.
- Không dùng GitHub auto review.
- Không dùng GitHub auto merge.
- Merge thủ công sau khi review, test evidence và diff đều ổn.
- Nếu CI GitHub lỗi ở workflow layer nhưng local/dev verification pass, ghi rõ trong PR trước khi manual merge.
```

---

## 9. Backend guardrails

### 9.1. Không sửa tồn trực tiếp

Cấm:

```text
UPDATE stock_balance SET qty = ...
```

Trừ các job rebuild/debug nội bộ có kiểm soát đặc biệt.

Mọi thay đổi tồn phải qua:

```text
stock_movement
stock_reservation
stock_adjustment_request
return_disposition
quarantine movement
```

### 9.2. Hàng hoàn chưa kiểm không available

Return mới scan phải vào trạng thái:

```text
RECEIVED / PENDING_INSPECTION
```

Không được vào:

```text
AVAILABLE
```

### 9.3. Hàng cần QA phải HOLD

Nếu disposition là `NEED_QA`:

```text
qc_status = HOLD
stock_status = QUARANTINE
available_qty = 0
```

### 9.4. Hàng không dùng không được bán

Nếu disposition là `NON_REUSABLE`:

```text
destination = LAB / DAMAGED / SCRAP
available_qty = 0
```

### 9.5. Đóng ca phải là checkpoint thật

Không cho đóng ca nếu còn:

```text
- return pending inspection
- manifest missing order unresolved
- stock variance unresolved
- adjustment pending approval
- failed stock movement
- unresolved P0/P1 warehouse incident
```

---

## 10. State model gợi ý

### 10.1. Return status

```text
DRAFT
RECEIVED
PENDING_INSPECTION
INSPECTED
DISPOSITIONED
PUTAWAY_COMPLETED
QUARANTINED
SENT_TO_LAB
CANCELLED
```

### 10.2. Return disposition

```text
REUSABLE
NON_REUSABLE
NEED_QA
UNKNOWN
```

### 10.3. Stock count status

```text
DRAFT
OPEN
COUNTING
SUBMITTED
VARIANCE_REVIEW
APPROVED
ADJUSTED
CLOSED
CANCELLED
```

### 10.4. Adjustment status

```text
DRAFT
SUBMITTED
APPROVED
REJECTED
POSTED
CANCELLED
```

### 10.5. Shift closing status

```text
OPEN
RECONCILING
BLOCKED
READY_TO_CLOSE
CLOSED
REOPENED
```

---

## 11. API endpoint gợi ý

> Endpoint cụ thể phải tuân theo file `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`.

```text
GET    /api/v1/return-reasons
POST   /api/v1/returns
GET    /api/v1/returns
GET    /api/v1/returns/{id}
POST   /api/v1/returns/scan
POST   /api/v1/returns/{id}/inspect
POST   /api/v1/returns/{id}/disposition
POST   /api/v1/returns/{id}/attachments

POST   /api/v1/stock-counts
GET    /api/v1/stock-counts
GET    /api/v1/stock-counts/{id}
POST   /api/v1/stock-counts/{id}/submit
POST   /api/v1/stock-counts/{id}/approve
POST   /api/v1/stock-counts/{id}/post-adjustment

POST   /api/v1/stock-adjustments
GET    /api/v1/stock-adjustments
POST   /api/v1/stock-adjustments/{id}/submit
POST   /api/v1/stock-adjustments/{id}/approve
POST   /api/v1/stock-adjustments/{id}/reject
POST   /api/v1/stock-adjustments/{id}/post

POST   /api/v1/warehouse-shifts
GET    /api/v1/warehouse-shifts/current
POST   /api/v1/warehouse-shifts/{id}/reconcile
POST   /api/v1/warehouse-shifts/{id}/close
POST   /api/v1/warehouse-shifts/{id}/reopen
```

---

## 12. UI screens cần có

### 12.1. Return Receiving Scan

```text
Route: /returns/receive

Components:
- Scan input
- Order/tracking result panel
- Return status chip
- Error state
- Create return button
```

### 12.2. Return Inspection

```text
Route: /returns/{id}/inspect

Components:
- Return header
- Original order summary
- SKU/batch table
- Condition checklist
- Disposition action
- Attachment uploader
- Audit panel
```

### 12.3. Return List

```text
Route: /returns

Filters:
- status
- disposition
- carrier
- received date
- SKU
- batch
- pending QA
```

### 12.4. Stock Count Session

```text
Route: /inventory/stock-counts

Components:
- Session header
- Warehouse/location selector
- SKU/batch count table
- System qty vs counted qty
- Variance indicator
- Submit count
```

### 12.5. Adjustment Approval

```text
Route: /inventory/adjustments/{id}

Components:
- Before/after qty
- Reason
- Supporting note/attachment
- Approval buttons
- Audit log
```

### 12.6. Shift Closing

```text
Route: /warehouse/shift-closing

Components:
- Today summary
- Returns pending
- Manifest issue
- Stock count status
- Adjustment pending
- Reconciliation checklist
- Close shift button
```

### 12.7. Warehouse Daily Board Update

```text
Route: /warehouse/daily-board

New widgets:
- Return pending
- QA hold from return
- Adjustment pending
- Stock count variance
- Shift closing status
```

---

## 13. QA / Test checklist

### Return flow

```text
[ ] Scan valid handed-over order.
[ ] Scan delivered order.
[ ] Scan invalid code.
[ ] Scan duplicate return.
[ ] Scan order not yet handed over.
[ ] Create return record with order link.
[ ] Inspect reusable item.
[ ] Inspect non-reusable item.
[ ] Inspect need-QA item.
[ ] Upload attachment.
[ ] Audit log created.
```

### Stock movement

```text
[ ] Reusable return creates inbound movement.
[ ] Non-reusable return does not affect available stock.
[ ] Need-QA return goes to quarantine.
[ ] Available qty calculation remains correct.
[ ] No direct stock mutation.
```

### Stock count / adjustment

```text
[ ] Open count session.
[ ] Count by warehouse/location/SKU/batch.
[ ] Submit variance.
[ ] Create adjustment request.
[ ] Approve adjustment.
[ ] Reject adjustment.
[ ] Post adjustment movement.
[ ] Audit log complete.
```

### Shift closing

```text
[ ] Close shift with clean data.
[ ] Block closing if return pending.
[ ] Block closing if adjustment pending.
[ ] Block closing if manifest issue pending.
[ ] Reconciliation totals match stock movement.
[ ] Warehouse Daily Board updates correctly.
```

---

## 14. Definition of Ready

Một task Sprint 3 chỉ được bắt đầu khi:

```text
- Task có owner.
- Task có Primary Ref.
- Acceptance criteria rõ.
- API/DB dependency đã biết.
- Test case tối thiểu đã xác định.
- Không conflict với state/ledger hiện tại.
```

---

## 15. Definition of Done

Một task Sprint 3 chỉ được coi là xong khi:

```text
- Code đã merge qua PR vào `main` theo manual review/merge; nếu team mở sprint branch riêng thì dùng target branch đã chốt.
- Unit/integration/API test pass nếu liên quan.
- Không dùng float cho money/qty/rate.
- Không sửa trực tiếp stock balance.
- Audit log có cho action nhạy cảm.
- RBAC permission đã check.
- UI có loading/error/empty state nếu là frontend task.
- OpenAPI cập nhật nếu có API mới.
- Migration có up/down nếu có DB mới.
- QA checklist pass.
```

---

## 16. Sprint 3 Acceptance Criteria tổng

Sprint 3 được coi là hoàn thành khi:

```text
1. Quét được mã đơn/vận đơn để nhận hàng hoàn.
2. Return record link được với order/SKU/batch nếu có.
3. Kiểm được tình trạng hàng hoàn.
4. Phân loại được còn dùng / không dùng / cần QA.
5. Hàng còn dùng mới được nhập lại kho.
6. Hàng không dùng chuyển Lab/kho hỏng, không vào available.
7. Hàng cần QA vào quarantine/HOLD.
8. Kiểm kê được theo kho/vị trí/SKU/batch.
9. Lệch tồn tạo adjustment request, không sửa trực tiếp.
10. Đối soát cuối ngày được đơn, hàng hoàn, movement, tồn.
11. Không cho đóng ca nếu còn issue chưa xử lý.
12. Warehouse Daily Board cập nhật return/reconciliation/closing.
13. Audit log đầy đủ.
14. E2E test return + shift closing pass.
```

---

## 17. Release name đề xuất

Sau khi Sprint 3 xong và merge/promote lên `main`, tag:

```bash
git tag v0.3.0-returns-reconciliation-core
git push origin v0.3.0-returns-reconciliation-core
```

Release summary:

```text
v0.3.0 adds returns receiving, return inspection, disposition, return stock movement,
quarantine return stock, inventory adjustment, stock count, end-of-day reconciliation,
shift closing, and Warehouse Daily Board updates.
```

---

## 18. Roadmap sau Sprint 3

Sau Sprint 3, nên đi tiếp:

```text
Sprint 4 — Purchase Order + Inbound QC Full Flow
Sprint 5 — Subcontract Manufacturing / Gia công ngoài
Sprint 6 — Finance Lite + COD + AR/AP cơ bản
Sprint 7 — Reporting v1 + Inventory/Operations Dashboard
```

Lý do Sprint 4 nên là Purchase/Inbound QC: sau khi kho đã có outbound, returns và closing, tiếp theo phải làm inbound purchasing/QC đầy đủ để nguồn hàng vào hệ thống sạch hơn.
