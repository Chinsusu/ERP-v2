# 33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham

**Dự án:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Giai đoạn:** Phase 1  
**Phiên bản:** v1.0  
**Mục tiêu:** Gom toàn bộ thay đổi v1.1 cần áp dụng vào các tài liệu lõi sau khi phân tích workflow thực tế kho, bàn giao ĐVVC, hàng hoàn và gia công ngoài.  
**Vai trò tài liệu:** Update Pack / Patch Pack / Core Docs Amendment Package.

---

## 0. Tóm tắt điều hành

Bộ tài liệu ERP hiện đã có từ `01` đến `32`. Trong đó, các tài liệu từ `03` đến `18` đã mô tả hệ thống ERP Phase 1 ở mức đủ để BA, UI/UX, backend, frontend, database, QA và DevOps bắt tay triển khai. Tuy nhiên, sau khi phân tích 4 workflow thực tế của doanh nghiệp, cần cập nhật một số điểm trọng yếu để hệ thống không bị “chuẩn ERP trên giấy nhưng lệch vận hành thật”.

Nguồn workflow thực tế cho thấy:

1. Kho vận hành theo nhịp hằng ngày: tiếp nhận đơn, xuất/nhập, soạn/đóng hàng, tối ưu vị trí, kiểm kê cuối ngày, đối soát số liệu và kết thúc ca.
2. Nội quy kho tách rõ 4 nhánh: nhập kho, xuất kho, đóng hàng và xử lý hàng hoàn.
3. Bàn giao ĐVVC không chỉ là đổi trạng thái đơn, mà là phân khu, để theo thùng/rổ, đối chiếu số lượng, quét mã trực tiếp và xử lý thiếu đơn.
4. Sản xuất hiện có nhánh gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển nguyên vật liệu/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, nhận hàng, kiểm tra, nhập kho hoặc báo lỗi trong 3–7 ngày.

Vì vậy bản v1.1 phải bổ sung 7 cụm nghiệp vụ bắt buộc:

```text
1. Warehouse Daily Board
2. Shift Closing / End-of-Day Reconciliation
3. Packing Queue & Pack Verification
4. Carrier Manifest & Scan Handover
5. Returns Receiving & Return Disposition
6. Subcontract Manufacturing / Gia công ngoài
7. Factory Claim SLA 3–7 days
```

**Một câu chốt:**

```text
Bản v1.0 = ERP lõi chuẩn.
Bản v1.1 = ERP lõi chuẩn + workflow thật của kho, ĐVVC, hàng hoàn và gia công ngoài.
```

---

## 1. Cách dùng tài liệu này

Tài liệu này không thay thế toàn bộ các file cũ. Nó là **bộ patch v1.1** để đội BA/PM/Tech Lead áp dụng vào các tài liệu lõi.

Có 2 cách dùng:

### Cách A — Dùng làm Addendum chính thức

Giữ nguyên các file `03–18` bản v1.0, sau đó gắn file này làm phụ lục bắt buộc.

Phù hợp khi:

- Team muốn bắt đầu Sprint 0 nhanh.
- Chưa muốn rewrite toàn bộ tài liệu lõi.
- Có PM đủ kỷ luật để luôn đọc kèm file 33.

### Cách B — Áp dụng patch để xuất bản v1.1 từng file

Tạo bản mới cho các file lõi:

```text
03_ERP_PRD_SRS_Phase1_My_Pham_v1.1.md
04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.1.md
05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.1.md
06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.1.md
08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.1.md
09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.1.md
16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.1.md
17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.1.md
```

Phù hợp khi:

- Chuẩn bị giao cho vendor/dev team.
- Muốn tài liệu không bị phân mảnh.
- Muốn mỗi file tự đầy đủ, không phải lật qua lại nhiều nơi.

**Khuyến nghị:** dùng file này làm Addendum ngay, sau đó trong Sprint 0 hoặc trước Sprint 1 thì apply thành v1.1 cho các core docs.

---

## 2. Nguồn đầu vào và phạm vi update

### 2.1. Nguồn đầu vào chính

| Nhóm | Tài liệu | Vai trò |
|---|---|---|
| Workflow thực tế | `Công-việc-hằng-ngày.pdf` | Căn cứ daily warehouse operation, kiểm kê và đối soát cuối ngày |
| Workflow thực tế | `Nội-Quy.pdf` | Căn cứ nhập kho, xuất kho, đóng hàng, hàng hoàn |
| Workflow thực tế | `Quy-trình-bàn-giao.pdf` | Căn cứ carrier handover, phân khu, đối chiếu, quét mã, xử lý thiếu đơn |
| Workflow thực tế | `Quy-trình-sản-xuất.pdf` | Căn cứ gia công ngoài, chuyển NVL/bao bì, duyệt mẫu, nhận hàng, claim nhà máy |
| As-Is | `20_ERP_Current_Workflow_AsIs_Warehouse_Production...` | Tài liệu bóc workflow hiện tại |
| Gap | `21_ERP_Gap_Analysis_AsIs_vs_ToBe...` | Tài liệu so sánh và decision log |
| Revision | `22_ERP_Core_Docs_Revision_v1_1_Change_Log...` | Danh sách thay đổi cần revise |
| Handoff | `32_ERP_Master_Document_Index_Traceability_Handoff...` | Source-of-truth và traceability toàn bộ tài liệu |

### 2.2. Tài liệu chịu ảnh hưởng

| Doc ID | Tài liệu | Mức ảnh hưởng | Bắt buộc update v1.1? |
|---|---|---:|---|
| 03 | PRD/SRS Phase 1 | Rất cao | Có |
| 04 | Permission/Approval Matrix | Cao | Có |
| 05 | Data Dictionary | Rất cao | Có |
| 06 | Process Flow To-Be | Rất cao | Có |
| 07 | Report/KPI Catalog | Trung bình-cao | Nên |
| 08 | Screen List/Wireframe | Rất cao | Có |
| 09 | UAT Test Scenarios | Cao | Có |
| 10 | Migration/Cutover | Trung bình | Nên |
| 11 | Technical Architecture Go | Cao | Nên |
| 12 | Go Coding Standards | Trung bình | Nên |
| 13 | Module/Component Standards | Cao | Nên |
| 14 | UI/UX Design System | Trung bình-cao | Nên |
| 15 | Frontend Architecture | Trung bình-cao | Nên |
| 16 | API/OpenAPI Standards | Rất cao | Có |
| 17 | Database Schema Standards | Rất cao | Có |
| 18 | DevOps/CI-CD | Thấp-trung bình | Nên |
| 19 | Security/RBAC/Audit | Trung bình-cao | Nên |
| 24 | QA Test Strategy | Cao | Nên |
| 25 | Backlog/Sprint Plan | Cao | Nên |

---

## 3. Baseline thay đổi v1.1

### 3.1. Scope mới của Phase 1

Bản Phase 1 v1.1 không chỉ gồm:

```text
Master Data
Purchase
QC
Inventory
Sales
Shipping
Subcontract Production
Basic Finance/Reporting
```

mà phải nhấn mạnh các capability vận hành thật sau:

```text
Warehouse Daily Operation
End-of-Day Warehouse Closing
Packing Work Queue
Pack Verification
Carrier Manifest
Carrier Scan Handover
Return Receiving
Return Disposition
Subcontract Manufacturing Order
Material/Packaging Transfer to Factory
Sample Approval
Factory Inbound QC
Factory Claim within 3–7 days
```

### 3.2. Các nguyên tắc mới phải khóa

1. **Kho có ngày làm việc và ca/phiên đóng ca.**  
   Không chỉ có phiếu nhập/xuất rời rạc.

2. **Đóng ca kho là một chứng từ kiểm soát.**  
   Nếu chưa đối soát đơn, hàng hoàn, nhập/xuất, variance thì chưa được close.

3. **Đóng hàng phải có task/queue riêng.**  
   Không để packing chỉ là một trạng thái của đơn hàng.

4. **Bàn giao ĐVVC phải có manifest.**  
   Manifest là chứng từ/tập hợp đơn theo carrier/đợt bàn giao.

5. **Quét mã là hành động kiểm soát, không chỉ là thao tác tiện ích.**  
   Scan event phải có timestamp, user, location/station, result, exception.

6. **Hàng hoàn có vòng đời riêng.**  
   Hàng hoàn phải được nhận, quét, kiểm tra, phân loại, chuyển kho hoặc chuyển lab/hỏng.

7. **Gia công ngoài là module nghiệp vụ rõ ràng.**  
   Không gộp bừa vào purchase hoặc production nội bộ.

8. **Claim nhà máy có SLA 3–7 ngày.**  
   Sau thời gian này phải có trạng thái và trách nhiệm rõ.

---

# PHẦN A — PATCH CHO 03 PRD/SRS PHASE 1

## A1. Bổ sung mục Scope v1.1

Thêm vào phần phạm vi Phase 1:

```text
Phase 1 v1.1 bổ sung các flow vận hành thực tế:

1. Warehouse Daily Board: bảng điều hành việc kho trong ngày.
2. Shift Closing: đóng ca, kiểm kê, đối soát cuối ngày.
3. Packing Queue: hàng chờ đóng, soạn/đóng theo đơn, pack verification.
4. Carrier Manifest: bảng bàn giao theo ĐVVC/đợt/thùng/rổ.
5. Scan Handover: quét mã xác nhận đủ/thiếu trước khi bàn giao ĐVVC.
6. Returns Receiving: nhận hàng hoàn từ shipper/ĐVVC.
7. Return Disposition: phân loại còn sử dụng / không sử dụng / cần kiểm tra thêm.
8. Subcontract Manufacturing: đặt gia công ngoài với nhà máy.
9. Factory Material Transfer: chuyển NVL/bao bì sang nhà máy.
10. Sample Approval: làm mẫu, chốt mẫu, lưu mẫu.
11. Factory Receiving QC: nhận hàng gia công về kho, kiểm tra số lượng/chất lượng.
12. Factory Claim: báo lỗi nhà máy trong 3–7 ngày nếu hàng không đạt.
```

## A2. Bổ sung Functional Requirement mới

### FR-WH-DAY-001 — Warehouse Daily Board

**Mô tả:**  
Hệ thống phải có bảng điều hành kho theo ngày để Warehouse Lead theo dõi toàn bộ công việc: đơn chờ xử lý, đơn chờ đóng hàng, đơn chờ bàn giao, hàng hoàn chờ xử lý, phiếu nhập chờ QC, phiếu xuất chờ hoàn tất, variance/tồn lệch.

**Actors:** Warehouse Lead, Warehouse Staff, COO, ERP Admin.

**Acceptance Criteria:**

- Xem được toàn bộ task kho trong ngày theo trạng thái.
- Có filter theo kho, ngày, carrier, loại task, trạng thái.
- Bấm vào số tổng để drill-down danh sách chứng từ.
- Hiển thị cảnh báo task overdue/blocked/exception.
- Không hiển thị dữ liệu vượt quyền người dùng.

### FR-WH-CLOSE-001 — Shift Closing / End-of-Day Reconciliation

**Mô tả:**  
Kho phải có chức năng đóng ca/cuối ngày để xác nhận đã xử lý hoặc ghi nhận exception cho các việc trong ngày.

**Acceptance Criteria:**

- Tạo được closing session theo kho/ngày/ca.
- Checklist đóng ca gồm: shipment, inbound, outbound, returns, stock variance, pending tasks.
- Không cho close nếu còn task P0 chưa xử lý mà không có exception note.
- Warehouse Lead phải sign-off.
- Sau khi close, hệ thống ghi audit log.

### FR-PACK-001 — Packing Queue

**Mô tả:**  
Đơn đã sẵn sàng xử lý phải vào hàng chờ đóng hàng. Nhân viên kho nhận task, soạn hàng, kiểm tra SKU/số lượng, đóng gói và xác nhận hoàn tất.

**Acceptance Criteria:**

- Đơn chỉ vào packing queue nếu đã reserved/pickable.
- Cho phép phân loại theo ĐVVC, đơn lẻ/đơn sàn, độ ưu tiên.
- Khi pack phải kiểm tra SKU/số lượng.
- Ghi nhận packer, thời gian bắt đầu/kết thúc, exception.
- Đơn packed mới được đưa vào carrier manifest.

### FR-SHIP-MANIFEST-001 — Carrier Manifest

**Mô tả:**  
Hệ thống phải tạo manifest bàn giao cho từng ĐVVC, gồm danh sách đơn, mã vận đơn, thùng/rổ/khu vực để hàng, số lượng đơn dự kiến, số lượng đã scan, số lượng thiếu/thừa.

**Acceptance Criteria:**

- Tạo manifest theo carrier/ngày/đợt/kho.
- Chỉ thêm được đơn đã packed.
- Có trạng thái Draft / Ready / Scanning / Completed / Exception / Cancelled.
- Có biên bản bàn giao/print/export.
- Có audit log khi thêm/xóa đơn khỏi manifest.

### FR-SHIP-SCAN-001 — Scan Handover

**Mô tả:**  
Nhân viên kho quét mã đơn/mã vận đơn trực tiếp trước khi ký bàn giao ĐVVC.

**Acceptance Criteria:**

- Nếu mã thuộc manifest và chưa scan: mark scanned.
- Nếu mã không thuộc manifest: cảnh báo wrong manifest.
- Nếu mã đã scan: cảnh báo duplicate scan.
- Nếu chưa đủ đơn: không cho complete manifest nếu không có exception.
- Nếu đủ đơn: cho Warehouse Lead xác nhận bàn giao.

### FR-RET-001 — Return Receiving

**Mô tả:**  
Nhận hàng hoàn từ shipper/ĐVVC, đưa vào khu vực hàng hoàn, quét hàng hoàn và tạo return receiving record.

**Acceptance Criteria:**

- Quét được mã đơn/mã vận đơn/mã return nếu có.
- Ghi nhận nguồn hoàn, người nhận, thời gian nhận, tình trạng bao bì ban đầu.
- Nếu không nhận diện được đơn, tạo unknown return case.
- Hàng hoàn phải ở trạng thái Pending Inspection trước khi đưa lại vào kho khả dụng.

### FR-RET-002 — Return Disposition

**Mô tả:**  
Kiểm tra tình trạng hàng hoàn và phân loại: còn sử dụng, không sử dụng, cần QC thêm, thiếu hàng, hỏng, sai sản phẩm.

**Acceptance Criteria:**

- Chỉ người có quyền mới được disposition.
- Hàng còn dùng được chuyển về kho phù hợp nhưng không tự động available nếu cần QC.
- Hàng không dùng được chuyển lab/kho hỏng/scrap.
- Ghi audit log và hình ảnh nếu có.
- Có báo cáo return reason/disposition.

### FR-SUB-001 — Subcontract Manufacturing Order

**Mô tả:**  
Tạo đơn gia công với nhà máy: sản phẩm, số lượng, quy cách, mẫu mã, thời gian sản xuất/nhận hàng, cọc/thanh toán.

**Acceptance Criteria:**

- Tạo đơn gia công từ nhu cầu sản xuất hoặc thủ công.
- Có trạng thái: Draft / Confirmed / Deposit Paid / Materials Prepared / Materials Sent / Sample Pending / Sample Approved / Mass Production / Inbound Pending / QC Checking / Accepted / Claimed / Closed / Cancelled.
- Có attachment hợp đồng, biên bản, COA/MSDS nếu cần.
- Có trường factory/vendor, expected receive date, deposit amount, final payment status.

### FR-SUB-TRANSFER-001 — Factory Material/Packaging Transfer

**Mô tả:**  
Chuyển nguyên vật liệu/bao bì sang nhà máy gia công và ký biên bản bàn giao.

**Acceptance Criteria:**

- Tạo transfer order từ kho công ty sang factory/subcontract location.
- Trừ tồn khỏi kho nguồn hoặc chuyển sang stock-in-subcontractor theo cấu hình.
- Ghi rõ batch/lot nếu có.
- Có biên bản bàn giao và attachment.
- Không cho dùng nguyên vật liệu chưa QC pass.

### FR-SUB-SAMPLE-001 — Sample Approval & Retained Sample

**Mô tả:**  
Nhà máy làm mẫu, công ty chốt mẫu, lưu mẫu và dùng làm căn cứ sản xuất hàng loạt/nghiệm thu.

**Acceptance Criteria:**

- Ghi nhận sample version.
- Có status Pending / Approved / Rejected / Rework.
- Có người duyệt, ngày duyệt, file/hình ảnh mẫu.
- Nếu sample chưa approved thì không cho chuyển sang Mass Production.

### FR-SUB-INBOUND-001 — Factory Receiving QC

**Mô tả:**  
Nhận hàng gia công về kho, kiểm tra số lượng/chất lượng trước khi nhập kho.

**Acceptance Criteria:**

- Tạo inbound receipt từ subcontract order.
- Kiểm số lượng nhận vs số lượng đặt.
- QC kiểm chất lượng, mẫu mã, bao bì, batch/hạn dùng nếu có.
- Hàng đạt mới nhập kho; hàng không đạt tạo claim case.

### FR-SUB-CLAIM-001 — Factory Claim 3–7 Days

**Mô tả:**  
Nếu hàng gia công không đạt, hệ thống phải tạo claim với nhà máy và theo dõi SLA 3–7 ngày.

**Acceptance Criteria:**

- Claim có issue type, evidence, affected qty, requested action.
- Có deadline phản hồi.
- Có trạng thái Open / Sent / Factory Responded / Resolved / Rejected / Overdue.
- Có cảnh báo quá hạn.
- Claim ảnh hưởng payment hold nếu cấu hình.

---

# PHẦN B — PATCH CHO 04 PERMISSION / APPROVAL MATRIX

## B1. Role mới hoặc cần làm rõ

| Role | Mô tả | Lý do bổ sung/làm rõ |
|---|---|---|
| Warehouse Lead | Trưởng kho/quản lý ca kho | Được close shift, confirm handover, xử lý exception |
| Packer | Nhân viên đóng hàng | Nhận packing task, scan/verify pack |
| Handover Operator | Nhân viên bàn giao ĐVVC | Quét mã manifest, ghi exception |
| Return Inspector | Nhân viên kiểm hàng hoàn | Phân loại tình trạng hàng hoàn |
| Subcontract Coordinator | Nhân sự phụ trách gia công ngoài | Tạo/điều phối đơn gia công, chuyển NVL/bao bì, claim nhà máy |
| Factory QC Reviewer | QA/QC kiểm hàng gia công | Duyệt/không duyệt hàng từ nhà máy |

## B2. Permission mới

| Permission Code | Mô tả | Role mặc định |
|---|---|---|
| `warehouse.daily_board.view` | Xem bảng điều hành kho ngày | Warehouse Staff, Warehouse Lead, COO |
| `warehouse.shift.open` | Mở ca/ngày kho | Warehouse Lead |
| `warehouse.shift.close` | Đóng ca/ngày kho | Warehouse Lead |
| `warehouse.shift.reopen` | Mở lại ca đã đóng | ERP Admin + COO approval |
| `packing.task.view` | Xem hàng chờ đóng | Warehouse Staff, Packer |
| `packing.task.assign` | Gán task đóng hàng | Warehouse Lead |
| `packing.task.complete` | Hoàn tất đóng hàng | Packer |
| `shipping.manifest.create` | Tạo manifest ĐVVC | Warehouse Lead, Shipping Staff |
| `shipping.manifest.update` | Sửa manifest khi chưa completed | Warehouse Lead |
| `shipping.manifest.scan` | Quét mã bàn giao | Handover Operator |
| `shipping.manifest.complete` | Xác nhận hoàn tất bàn giao | Warehouse Lead |
| `shipping.manifest.exception` | Ghi nhận thiếu/thừa/sai manifest | Warehouse Lead, Handover Operator |
| `returns.receive` | Nhận hàng hoàn | Warehouse Staff, Return Inspector |
| `returns.inspect` | Kiểm tình trạng hàng hoàn | Return Inspector, QA |
| `returns.dispose` | Chốt disposition hàng hoàn | Warehouse Lead, QA |
| `returns.move_to_available` | Chuyển về kho bán được | Warehouse Lead + QC if required |
| `subcontract.order.create` | Tạo đơn gia công | Subcontract Coordinator, Production Planner |
| `subcontract.order.approve` | Duyệt đơn gia công | COO / Production Manager |
| `subcontract.material_transfer.create` | Tạo chuyển NVL/bao bì sang nhà máy | Subcontract Coordinator |
| `subcontract.material_transfer.approve` | Duyệt chuyển NVL/bao bì | Warehouse Lead + Production Manager |
| `subcontract.sample.approve` | Duyệt mẫu | QA/R&D/Brand Owner |
| `subcontract.inbound.receive` | Nhận hàng từ nhà máy | Warehouse Staff |
| `subcontract.inbound.qc` | QC hàng gia công | QA/QC |
| `subcontract.claim.create` | Tạo claim nhà máy | QA/QC, Subcontract Coordinator |
| `subcontract.claim.close` | Đóng claim | COO/QA Lead |

## B3. Approval bổ sung

| Giao dịch | Người tạo | Người duyệt | Điều kiện duyệt |
|---|---|---|---|
| Close Shift | Warehouse Lead | COO nếu có variance lớn | Có checklist + variance/exceptions rõ |
| Reopen Closed Shift | Warehouse Lead/ERP Admin | COO + ERP Admin | Bắt buộc audit reason |
| Complete Carrier Manifest | Handover Operator | Warehouse Lead | Đủ scan hoặc có exception hợp lệ |
| Return to Available | Return Inspector | QA/Warehouse Lead | Hàng còn dùng + đạt điều kiện QC |
| Return to Scrap/Lab | Return Inspector | Warehouse Lead/QA | Có reason + evidence |
| Subcontract Order | Coordinator | Production Manager/COO | Số lượng/quy cách/mẫu/tiến độ rõ |
| Material Transfer to Factory | Coordinator | Warehouse Lead + Production Manager | NVL/bao bì pass QC và đủ tồn |
| Sample Approval | Factory/R&D/QA | QA/R&D/Brand | Mẫu đúng spec và claim |
| Factory Claim | QA/Subcontract Coordinator | QA Lead/COO | Evidence rõ, trong SLA 3–7 ngày |
| Final Payment to Factory | Finance | Finance Lead/CEO/COO | Chỉ khi hàng accepted hoặc claim resolved theo rule |

## B4. Field-level restriction bổ sung

| Field | Ai được xem | Ai được sửa |
|---|---|---|
| `shift_variance_amount` | Warehouse Lead, COO, Finance | System/ERP Admin with approval |
| `manifest_exception_reason` | Warehouse Lead, Shipping, COO | Handover Operator/Warehouse Lead |
| `return_disposition` | Warehouse, QA, COO | Return Inspector/QA/Warehouse Lead |
| `factory_deposit_amount` | Production, Finance, CEO | Finance/authorized user |
| `factory_final_payment_status` | Finance, COO, CEO | Finance |
| `factory_claim_evidence` | QA, Production, COO | QA/Subcontract Coordinator |
| `sample_approval_status` | R&D, QA, Production, COO | QA/R&D authorized approver |

---

# PHẦN C — PATCH CHO 05 DATA DICTIONARY / MASTER DATA

## C1. Entity mới

| Entity | Mô tả | Module Owner |
|---|---|---|
| `warehouse_shift` | Ca/ngày vận hành kho | Inventory/Warehouse |
| `shift_closing_checklist` | Checklist đóng ca | Inventory/Warehouse |
| `shift_variance` | Lệch cuối ca/ngày | Inventory/Warehouse |
| `warehouse_daily_task` | Task kho trong ngày | Inventory/Warehouse |
| `packing_task` | Task đóng hàng | Shipping/Warehouse |
| `pack_verification` | Kết quả kiểm đóng hàng | Shipping/Warehouse |
| `carrier_manifest` | Bảng/phiếu bàn giao ĐVVC | Shipping |
| `carrier_manifest_line` | Dòng đơn trong manifest | Shipping |
| `scan_event` | Sự kiện quét mã | Shipping/Warehouse/Returns |
| `return_receipt` | Phiếu nhận hàng hoàn | Returns |
| `return_inspection` | Kết quả kiểm hàng hoàn | Returns/QA |
| `return_disposition` | Quyết định xử lý hàng hoàn | Returns/Warehouse/QA |
| `subcontract_order` | Đơn gia công ngoài | Subcontract/Production |
| `subcontract_material_transfer` | Phiếu chuyển NVL/bao bì sang nhà máy | Subcontract/Inventory |
| `subcontract_sample` | Mẫu gia công/chốt mẫu | Subcontract/R&D/QA |
| `subcontract_inbound_receipt` | Nhận hàng gia công về kho | Subcontract/Inventory/QC |
| `factory_claim` | Claim/báo lỗi nhà máy | Subcontract/QA |

## C2. Field/status mới

### Warehouse Shift

| Field | Type | Required | Mô tả |
|---|---|---:|---|
| `shift_id` | UUID | Yes | ID ca/ngày kho |
| `warehouse_id` | UUID | Yes | Kho vận hành |
| `shift_date` | Date | Yes | Ngày vận hành |
| `shift_code` | String | Yes | Mã ca, ví dụ `WH-HCM-20260424-DAY` |
| `opened_by` | UUID | Yes | Người mở ca |
| `opened_at` | Timestamp | Yes | Thời điểm mở |
| `closed_by` | UUID | No | Người đóng ca |
| `closed_at` | Timestamp | No | Thời điểm đóng |
| `status` | Enum | Yes | `OPEN`, `CLOSING`, `CLOSED`, `REOPENED`, `CANCELLED` |
| `closing_note` | Text | No | Ghi chú đóng ca |

### Carrier Manifest

| Field | Type | Required | Mô tả |
|---|---|---:|---|
| `manifest_id` | UUID | Yes | ID manifest |
| `manifest_no` | String | Yes | Mã manifest |
| `carrier_id` | UUID | Yes | ĐVVC |
| `warehouse_id` | UUID | Yes | Kho bàn giao |
| `handover_date` | Date | Yes | Ngày bàn giao |
| `handover_batch` | String | No | Đợt bàn giao trong ngày |
| `expected_order_count` | Int | Yes | Số đơn dự kiến |
| `scanned_order_count` | Int | Yes | Số đơn đã scan |
| `missing_order_count` | Int | Yes | Số đơn thiếu |
| `extra_scan_count` | Int | Yes | Số scan thừa/sai |
| `status` | Enum | Yes | `DRAFT`, `READY`, `SCANNING`, `COMPLETED`, `EXCEPTION`, `CANCELLED` |
| `completed_by` | UUID | No | Người hoàn tất |
| `completed_at` | Timestamp | No | Thời điểm hoàn tất |

### Return Disposition

| Field | Type | Required | Mô tả |
|---|---|---:|---|
| `return_receipt_id` | UUID | Yes | Phiếu hàng hoàn |
| `original_order_id` | UUID | No | Đơn gốc nếu nhận diện được |
| `tracking_no` | String | No | Mã vận đơn |
| `return_source` | Enum | Yes | `SHIPPER`, `CARRIER`, `CUSTOMER`, `MARKETPLACE`, `UNKNOWN` |
| `inspection_status` | Enum | Yes | `PENDING`, `INSPECTING`, `INSPECTED`, `ESCALATED` |
| `disposition` | Enum | No | `REUSABLE`, `QC_REQUIRED`, `DAMAGED`, `MISSING_ITEM`, `WRONG_ITEM`, `SCRAP`, `UNKNOWN` |
| `target_location_id` | UUID | No | Kho/vị trí đích |
| `evidence_files` | Array | No | Hình ảnh/video/biên bản |

### Subcontract Order

| Field | Type | Required | Mô tả |
|---|---|---:|---|
| `subcontract_order_id` | UUID | Yes | ID đơn gia công |
| `subcontract_order_no` | String | Yes | Mã đơn gia công |
| `factory_supplier_id` | UUID | Yes | Nhà máy gia công |
| `product_id` | UUID | Yes | Sản phẩm thành phẩm |
| `planned_qty` | Decimal | Yes | Số lượng đặt |
| `spec_version_id` | UUID | No | Version spec/quy cách |
| `sample_required` | Boolean | Yes | Có cần làm mẫu/chốt mẫu không |
| `deposit_amount` | Decimal | No | Tiền cọc |
| `final_payment_amount` | Decimal | No | Thanh toán cuối |
| `expected_receive_date` | Date | No | Ngày dự kiến nhận |
| `status` | Enum | Yes | Xem status bên dưới |

### Subcontract Order Status

```text
DRAFT
CONFIRMED
DEPOSIT_PAID
MATERIALS_PREPARED
MATERIALS_SENT
SAMPLE_PENDING
SAMPLE_APPROVED
SAMPLE_REJECTED
MASS_PRODUCTION
INBOUND_PENDING
QC_CHECKING
ACCEPTED
CLAIMED
CLOSED
CANCELLED
```

## C3. Công thức/logic bổ sung

### Tồn khả dụng v1.1

```text
available_stock
= physical_stock
- reserved_stock
- qc_hold_stock
- packed_not_handed_over_stock
- stock_pending_return_inspection
- stock_in_subcontractor_hold
```

### Manifest Completion Rate

```text
manifest_completion_rate
= scanned_order_count / expected_order_count
```

### Return Reusable Rate

```text
return_reusable_rate
= reusable_return_qty / total_return_received_qty
```

### Factory Claim SLA

```text
factory_claim_sla_status
= ON_TIME if claim_created_at <= inbound_received_at + SLA_days
= OVERDUE if claim_created_at > inbound_received_at + SLA_days
```

---

# PHẦN D — PATCH CHO 06 PROCESS FLOW TO-BE

## D1. Flow Warehouse Daily Operation

```text
Start Warehouse Day
→ Open Warehouse Shift
→ Load Warehouse Daily Board
→ Process Inbound / Outbound / Packing / Returns / Handover
→ Monitor Exceptions
→ Cycle Count / Stock Count if required
→ Reconcile Orders, Shipments, Returns, Inbound, Outbound
→ Record Variance
→ Warehouse Lead Review
→ Close Shift
→ Daily Report Sent
```

### Điểm kiểm soát

- Mọi task trong ngày phải có status.
- Task blocked phải có lý do.
- Đóng ca phải có sign-off.
- Variance phải có owner xử lý.

## D2. Flow Packing Queue

```text
Sales Order Confirmed / Ready to Fulfill
→ Reserve Stock
→ Create Pick/Packing Task
→ Warehouse picks items
→ Move to Packing Area
→ Pack Verification: SKU / qty / condition
→ Pack Completed
→ Order ready for Carrier Manifest
```

### Exception

| Exception | Xử lý |
|---|---|
| Thiếu hàng | Mark shortage, notify Warehouse Lead/Sales |
| Sai SKU | Block pack, return to picking correction |
| Bao bì lỗi | Replace/repack, record exception |
| Đơn bị hủy khi đang pack | Stop pack, reverse reservation/stock action theo rule |

## D3. Flow Carrier Manifest & Scan Handover

```text
Create Carrier Manifest
→ Add Packed Orders
→ Assign staging zone/thùng/rổ
→ Ready for Scan
→ Scan order/tracking barcode one by one
→ System validates manifest membership
→ If enough: Warehouse Lead signs handover
→ Manifest Completed
→ Order status = Handed Over
```

### Exception flow

```text
Scan order not in manifest
→ Warning: Wrong Manifest
→ User chooses investigate / reject scan

Duplicate scan
→ Warning: Already Scanned

Missing order
→ Check code
→ If code valid but physical order missing: search packing area
→ If found: scan and continue
→ If not found: mark missing exception and block/partial complete by approval
```

## D4. Flow Returns Receiving & Disposition

```text
Receive returned goods from shipper/carrier
→ Move to Return Area
→ Scan order/tracking/return code
→ Create Return Receipt
→ Inspect item condition
→ Determine disposition
    → Reusable: move to available/quarantine based on QC rule
    → QC Required: move to QC hold
    → Damaged/Not Usable: move to Lab/Scrap/Damage location
    → Unknown: create investigation case
→ Close Return Receipt
```

### Điểm kiểm soát

- Hàng hoàn không tự động quay lại available.
- Phải có disposition.
- Nếu không nhận diện được đơn gốc thì không cho trộn vào kho bán được.
- Return reason cần thống kê để xử lý lỗi nguồn.

## D5. Flow Subcontract Manufacturing

```text
Create Subcontract Order with factory
→ Confirm qty/spec/design/sample/production timeline
→ Deposit Payment if required
→ Prepare material/packaging
→ Transfer material/packaging to factory
→ Sign handover document
→ Factory makes sample
→ Sample Review
    → Approved: proceed mass production
    → Rejected: rework sample
→ Factory mass production
→ Factory delivers goods to warehouse
→ Receive goods
→ QC quantity/quality/spec/package
    → Accepted: inbound to warehouse
    → Not accepted: create Factory Claim within 3–7 days
→ Final Payment if accepted/resolved
→ Close Subcontract Order
```

---

# PHẦN E — PATCH CHO 07 REPORT / KPI CATALOG

## E1. KPI kho hằng ngày

| KPI | Công thức | Người xem |
|---|---|---|
| Daily Warehouse Task Completion | Completed tasks / Total tasks | Warehouse Lead, COO |
| Pending Pack Orders | Count order status ready/picking/packing | Warehouse Lead |
| End-of-Day Variance Count | Count variance lines | Warehouse Lead, Finance, COO |
| Shift Closing On-Time Rate | Shifts closed on time / total shifts | COO |
| Return Pending Inspection | Count returns pending inspection | Warehouse Lead, QA |

## E2. KPI bàn giao ĐVVC

| KPI | Công thức | Người xem |
|---|---|---|
| Manifest Completion Rate | Scanned / Expected | Warehouse Lead, Shipping |
| Handover Missing Orders | Count missing exception | Warehouse Lead, COO |
| Wrong Manifest Scans | Count wrong manifest scan | Warehouse Lead |
| Carrier Handover Delay | Actual completed time - planned handover time | COO |

## E3. KPI hàng hoàn

| KPI | Công thức | Người xem |
|---|---|---|
| Return Receiving SLA | Time from return received to receipt created | Warehouse Lead |
| Return Inspection SLA | Time from receipt to disposition | Warehouse Lead, QA |
| Reusable Return Rate | Reusable qty / total returned qty | COO |
| Damaged Return Rate | Damaged qty / total returned qty | COO, CSKH |

## E4. KPI gia công ngoài

| KPI | Công thức | Người xem |
|---|---|---|
| Factory On-Time Delivery | On-time receipts / total subcontract receipts | COO |
| Sample Approval Cycle Time | Approved at - sample requested at | R&D, QA, COO |
| Factory Claim Rate | Claimed qty/orders / received qty/orders | QA, COO |
| Factory Claim SLA Compliance | Claims created within SLA / total claims | QA, COO |
| Material Transfer Accuracy | Transfer qty matched / planned transfer qty | Warehouse, Production |

---

# PHẦN F — PATCH CHO 08 SCREEN LIST / WIREFRAME

## F1. Màn hình mới bắt buộc

| Screen Code | Screen Name | Module | Priority |
|---|---|---|---|
| `WH-DAILY-BOARD` | Warehouse Daily Board | Warehouse | P0 |
| `WH-SHIFT-CLOSE` | Shift Closing / End-of-Day Reconciliation | Warehouse | P0 |
| `PACK-QUEUE` | Packing Queue | Warehouse/Shipping | P0 |
| `PACK-VERIFY` | Pack Verification | Warehouse/Shipping | P0 |
| `SHIP-MANIFEST-LIST` | Carrier Manifest List | Shipping | P0 |
| `SHIP-MANIFEST-DETAIL` | Carrier Manifest Detail | Shipping | P0 |
| `SHIP-SCAN-HANDOVER` | Scan Handover Station | Shipping/Warehouse | P0 |
| `RET-RECEIVE` | Return Receiving | Returns | P0 |
| `RET-INSPECT` | Return Inspection / Disposition | Returns/QA | P0 |
| `SUB-ORDER-LIST` | Subcontract Order List | Subcontract | P0 |
| `SUB-ORDER-DETAIL` | Subcontract Order Detail | Subcontract | P0 |
| `SUB-MATERIAL-TRANSFER` | Factory Material Transfer | Subcontract/Inventory | P0 |
| `SUB-SAMPLE-APPROVAL` | Sample Approval | Subcontract/R&D/QA | P0 |
| `SUB-INBOUND-QC` | Factory Inbound QC | Subcontract/QC | P0 |
| `SUB-FACTORY-CLAIM` | Factory Claim | Subcontract/QA | P0 |

## F2. Wireframe text — Warehouse Daily Board

```text
[Header]
Warehouse Daily Board | Warehouse: [select] | Date: [today] | Shift: [open/closed]

[Summary Cards]
- Orders to Pick
- Orders to Pack
- Manifests Pending Scan
- Returns Pending Inspection
- Inbound Pending QC
- Stock Variance
- Blocked Tasks

[Task Panels]
1. Picking / Packing
2. Carrier Handover
3. Returns
4. Inbound / QC
5. Stock Count / Variance

[Right Sidebar]
- Alerts
- Overdue tasks
- Shift closing checklist

[Actions]
Open Shift | Start Closing | Export Daily Report
```

## F3. Wireframe text — Scan Handover Station

```text
[Header]
Scan Handover | Manifest: MF-xxxx | Carrier | Handover batch | Status

[Scan Input]
Focus barcode input here

[Progress]
Expected: 120 | Scanned: 118 | Missing: 2 | Wrong/Duplicate: 0

[Recent Scans]
Time | Barcode | Order | Result | Operator

[Missing Orders]
Order No | Tracking No | Zone/Box | Action: Mark Missing / Find / Remove by Approval

[Actions]
Pause Scan | Complete Manifest | Print Handover | Report Exception
```

## F4. Wireframe text — Return Inspection

```text
[Header]
Return Inspection | Return Receipt No | Original Order | Tracking No

[Product Lines]
SKU | Product | Qty returned | Condition | Batch | Expiry | Evidence

[Disposition]
Reusable / QC Required / Damaged / Missing Item / Wrong Item / Scrap / Unknown

[Target Location]
Available / QC Hold / Return Area / Lab / Scrap

[Actions]
Submit Disposition | Escalate to QA | Attach Evidence | Print Return Note
```

## F5. Wireframe text — Subcontract Order Detail

```text
[Header]
Subcontract Order | Factory | Product | Qty | Status | Expected Receive Date

[Tabs]
1. Overview
2. Material & Packaging Transfer
3. Sample Approval
4. Mass Production
5. Inbound & QC
6. Factory Claim
7. Payment & Attachments
8. Audit Log

[Actions by Status]
Confirm Order | Record Deposit | Prepare Transfer | Send Materials | Request Sample | Approve Sample | Start Mass Production | Receive Goods | Create Claim | Close Order
```

---

# PHẦN G — PATCH CHO 09 UAT TEST SCENARIOS

## G1. UAT mới — Warehouse Daily Board

| ID | Scenario | Expected Result |
|---|---|---|
| UAT-WH-DAY-001 | Warehouse Lead mở daily board đầu ngày | Hệ thống hiển thị task trong ngày theo đúng trạng thái |
| UAT-WH-DAY-002 | Có đơn blocked do thiếu hàng | Daily board hiển thị cảnh báo blocked và drill-down được |
| UAT-WH-DAY-003 | Hàng hoàn mới nhận chưa inspect | Daily board tăng số return pending inspection |

## G2. UAT mới — Shift Closing

| ID | Scenario | Expected Result |
|---|---|---|
| UAT-WH-CLOSE-001 | Close shift khi còn manifest chưa xử lý | Hệ thống không cho close hoặc yêu cầu exception |
| UAT-WH-CLOSE-002 | Close shift có variance | Hệ thống yêu cầu ghi reason và sign-off |
| UAT-WH-CLOSE-003 | Reopen closed shift | Chỉ role có quyền và approval mới được reopen, có audit log |

## G3. UAT mới — Carrier Manifest / Scan Handover

| ID | Scenario | Expected Result |
|---|---|---|
| UAT-SHIP-MF-001 | Tạo manifest từ đơn đã packed | Manifest được tạo đúng danh sách |
| UAT-SHIP-MF-002 | Quét đúng tất cả đơn | Manifest completed, order chuyển HandedOver |
| UAT-SHIP-MF-003 | Quét mã không thuộc manifest | Hệ thống cảnh báo wrong manifest |
| UAT-SHIP-MF-004 | Quét trùng mã | Hệ thống cảnh báo duplicate scan |
| UAT-SHIP-MF-005 | Thiếu đơn khi complete | Hệ thống block complete hoặc yêu cầu exception approval |

## G4. UAT mới — Returns

| ID | Scenario | Expected Result |
|---|---|---|
| UAT-RET-001 | Nhận hàng hoàn có mã đơn | Tạo return receipt liên kết đơn gốc |
| UAT-RET-002 | Nhận hàng hoàn không nhận diện được | Tạo unknown return case |
| UAT-RET-003 | Hàng còn dùng được | Chuyển đúng target location theo QC rule |
| UAT-RET-004 | Hàng không dùng được | Chuyển Lab/Scrap, không vào available stock |

## G5. UAT mới — Subcontract Manufacturing

| ID | Scenario | Expected Result |
|---|---|---|
| UAT-SUB-001 | Tạo đơn gia công | Status Draft/Confirmed đúng, lưu spec/qty/factory |
| UAT-SUB-002 | Chuyển NVL/bao bì sang nhà máy | Stock movement đúng, có biên bản bàn giao |
| UAT-SUB-003 | Sample chưa approved mà start mass production | Hệ thống không cho chuyển trạng thái |
| UAT-SUB-004 | Nhận hàng gia công đạt | QC pass, nhập kho đúng |
| UAT-SUB-005 | Nhận hàng gia công lỗi | Tạo factory claim và deadline 3–7 ngày |
| UAT-SUB-006 | Thanh toán cuối khi claim chưa resolved | Hệ thống cảnh báo/hold theo rule |

---

# PHẦN H — PATCH CHO 16 API / OPENAPI STANDARDS

## H1. Endpoint mới — Warehouse Daily Board

```http
GET /api/v1/warehouses/{warehouse_id}/daily-board?date=YYYY-MM-DD
```

Response chính:

```json
{
  "success": true,
  "data": {
    "warehouse_id": "uuid",
    "date": "2026-04-24",
    "shift_status": "OPEN",
    "summary": {
      "orders_to_pick": 12,
      "orders_to_pack": 35,
      "manifests_pending_scan": 3,
      "returns_pending_inspection": 8,
      "inbound_pending_qc": 2,
      "blocked_tasks": 4
    },
    "alerts": []
  }
}
```

## H2. Endpoint mới — Shift Closing

```http
POST /api/v1/warehouses/{warehouse_id}/shifts
POST /api/v1/warehouse-shifts/{shift_id}/start-closing
POST /api/v1/warehouse-shifts/{shift_id}/close
POST /api/v1/warehouse-shifts/{shift_id}/reopen
GET  /api/v1/warehouse-shifts/{shift_id}/closing-checklist
```

## H3. Endpoint mới — Packing

```http
GET  /api/v1/packing/tasks
POST /api/v1/packing/tasks/{task_id}/assign
POST /api/v1/packing/tasks/{task_id}/start
POST /api/v1/packing/tasks/{task_id}/verify-line
POST /api/v1/packing/tasks/{task_id}/complete
POST /api/v1/packing/tasks/{task_id}/report-exception
```

## H4. Endpoint mới — Carrier Manifest / Scan Handover

```http
GET  /api/v1/shipping/manifests
POST /api/v1/shipping/manifests
GET  /api/v1/shipping/manifests/{manifest_id}
POST /api/v1/shipping/manifests/{manifest_id}/add-orders
POST /api/v1/shipping/manifests/{manifest_id}/ready
POST /api/v1/shipping/manifests/{manifest_id}/scan
POST /api/v1/shipping/manifests/{manifest_id}/complete
POST /api/v1/shipping/manifests/{manifest_id}/exceptions
GET  /api/v1/shipping/manifests/{manifest_id}/print
```

Scan request:

```json
{
  "barcode": "TRACKING_OR_ORDER_CODE",
  "scan_station": "HCM-WH-HANDOVER-01",
  "scan_context": "HANDOVER"
}
```

Scan response:

```json
{
  "success": true,
  "data": {
    "result": "MATCHED",
    "order_id": "uuid",
    "tracking_no": "ABC123",
    "scanned_count": 45,
    "expected_count": 50
  }
}
```

Error examples:

```text
WRONG_MANIFEST
DUPLICATE_SCAN
ORDER_NOT_PACKED
MANIFEST_ALREADY_COMPLETED
```

## H5. Endpoint mới — Returns

```http
POST /api/v1/returns/receipts
GET  /api/v1/returns/receipts
GET  /api/v1/returns/receipts/{return_receipt_id}
POST /api/v1/returns/receipts/{return_receipt_id}/scan
POST /api/v1/returns/receipts/{return_receipt_id}/inspect
POST /api/v1/returns/receipts/{return_receipt_id}/dispose
POST /api/v1/returns/receipts/{return_receipt_id}/escalate
```

## H6. Endpoint mới — Subcontract Manufacturing

```http
GET  /api/v1/subcontract-orders
POST /api/v1/subcontract-orders
GET  /api/v1/subcontract-orders/{id}
POST /api/v1/subcontract-orders/{id}/confirm
POST /api/v1/subcontract-orders/{id}/record-deposit
POST /api/v1/subcontract-orders/{id}/material-transfers
POST /api/v1/subcontract-orders/{id}/samples
POST /api/v1/subcontract-orders/{id}/samples/{sample_id}/approve
POST /api/v1/subcontract-orders/{id}/samples/{sample_id}/reject
POST /api/v1/subcontract-orders/{id}/start-mass-production
POST /api/v1/subcontract-orders/{id}/inbound-receipts
POST /api/v1/subcontract-orders/{id}/factory-claims
POST /api/v1/subcontract-orders/{id}/close
```

---

# PHẦN I — PATCH CHO 17 DATABASE SCHEMA POSTGRESQL

## I1. Tables mới đề xuất

```sql
warehouse_shifts
shift_closing_checklists
shift_variances
warehouse_daily_tasks
packing_tasks
packing_task_lines
pack_verifications
carrier_manifests
carrier_manifest_lines
scan_events
return_receipts
return_receipt_lines
return_inspections
return_dispositions
subcontract_orders
subcontract_order_lines
subcontract_material_transfers
subcontract_material_transfer_lines
subcontract_samples
subcontract_inbound_receipts
subcontract_inbound_receipt_lines
factory_claims
factory_claim_lines
```

## I2. Naming/status constraints

### carrier_manifests.status

```sql
CHECK (status IN (
  'DRAFT',
  'READY',
  'SCANNING',
  'COMPLETED',
  'EXCEPTION',
  'CANCELLED'
))
```

### return_dispositions.disposition

```sql
CHECK (disposition IN (
  'REUSABLE',
  'QC_REQUIRED',
  'DAMAGED',
  'MISSING_ITEM',
  'WRONG_ITEM',
  'SCRAP',
  'UNKNOWN'
))
```

### subcontract_orders.status

```sql
CHECK (status IN (
  'DRAFT',
  'CONFIRMED',
  'DEPOSIT_PAID',
  'MATERIALS_PREPARED',
  'MATERIALS_SENT',
  'SAMPLE_PENDING',
  'SAMPLE_APPROVED',
  'SAMPLE_REJECTED',
  'MASS_PRODUCTION',
  'INBOUND_PENDING',
  'QC_CHECKING',
  'ACCEPTED',
  'CLAIMED',
  'CLOSED',
  'CANCELLED'
))
```

## I3. Index bắt buộc

```sql
CREATE INDEX idx_carrier_manifests_carrier_date
ON carrier_manifests(carrier_id, handover_date, status);

CREATE UNIQUE INDEX idx_carrier_manifest_lines_manifest_order
ON carrier_manifest_lines(manifest_id, order_id);

CREATE INDEX idx_scan_events_context_barcode
ON scan_events(scan_context, barcode, created_at);

CREATE INDEX idx_return_receipts_status_created
ON return_receipts(status, created_at);

CREATE INDEX idx_subcontract_orders_factory_status
ON subcontract_orders(factory_supplier_id, status);

CREATE INDEX idx_factory_claims_deadline_status
ON factory_claims(response_deadline_at, status);
```

## I4. Stock ledger event types mới

```text
PACKED_NOT_HANDED_OVER
MANIFEST_HANDOVER_CONFIRMED
RETURN_RECEIVED
RETURN_QC_HOLD
RETURN_TO_AVAILABLE
RETURN_TO_SCRAP
SUBCONTRACT_MATERIAL_SENT
SUBCONTRACT_MATERIAL_ADJUSTMENT
SUBCONTRACT_FINISHED_GOODS_RECEIVED
SUBCONTRACT_REJECTED_GOODS_HOLD
```

---

# PHẦN J — PATCH CHO 24 QA TEST STRATEGY

## J1. Regression Suite mới bắt buộc

| Suite | Test trọng tâm |
|---|---|
| Warehouse Daily | Daily board, open/close shift, variance |
| Packing | Pack task, verify SKU/qty, exception |
| Manifest Scan | matched/wrong/duplicate/missing scan |
| Returns | return receiving, unknown return, disposition |
| Subcontract | order lifecycle, material transfer, sample approval, inbound QC, factory claim |
| Stock Ledger | movement event đúng, không update tồn tay |
| Security/RBAC | field/action permission đúng |

## J2. Smoke test sau deploy

```text
1. Login đúng role Warehouse Lead.
2. Mở Warehouse Daily Board.
3. Tạo packing task demo.
4. Pack verify demo order.
5. Tạo carrier manifest demo.
6. Scan đúng một đơn.
7. Scan sai một mã để kiểm warning.
8. Tạo return receipt demo.
9. Tạo subcontract order demo.
10. Kiểm audit log.
```

---

# PHẦN K — PATCH CHO 25 PRODUCT BACKLOG / SPRINT PLAN

## K1. Epic mới/bắt buộc cập nhật

| Epic | Priority | Ghi chú |
|---|---|---|
| EP-WH-DAY | P0 | Warehouse Daily Board + Shift Closing |
| EP-PACK | P0 | Packing Queue + Pack Verification |
| EP-SHIP-MANIFEST | P0 | Carrier Manifest + Scan Handover |
| EP-RETURNS | P0 | Return Receiving + Disposition |
| EP-SUBCONTRACT | P0 | Gia công ngoài end-to-end |
| EP-FACTORY-CLAIM | P0 | Claim nhà máy + SLA |

## K2. Sprint plan điều chỉnh

Nếu sprint plan ban đầu là 13 sprint, nên chèn/điều chỉnh như sau:

```text
Sprint 0: Foundation + Auth + Master Data + Stock Ledger prototype
Sprint 1: Master Data + Warehouse base + Batch/QC base
Sprint 2: Purchase/Inbound/QC
Sprint 3: Inventory Movement + Stock Ledger + Stock Count
Sprint 4: Sales Order + Reservation
Sprint 5: Picking/Packing Queue
Sprint 6: Carrier Manifest + Scan Handover
Sprint 7: Returns Receiving + Disposition
Sprint 8: Subcontract Order + Material Transfer
Sprint 9: Sample Approval + Subcontract Inbound QC + Factory Claim
Sprint 10: Finance basic + COD/Payment hooks + Reporting
Sprint 11: UAT Hardening + Regression
Sprint 12: Data Migration + Cutover rehearsal
Sprint 13: Go-live + Hypercare
```

## K3. User stories mẫu

### US-SHIP-001 — Tạo manifest bàn giao ĐVVC

**As a** Warehouse Lead  
**I want** tạo manifest theo ĐVVC và đợt bàn giao  
**So that** kho có danh sách đơn chính thức để quét và ký bàn giao.

Acceptance Criteria:

- Chỉ thêm đơn đã packed.
- Có số đơn expected.
- Có staging zone/thùng/rổ.
- Có trạng thái manifest.
- Có audit log khi thay đổi line.

### US-SHIP-002 — Quét mã bàn giao

**As a** Handover Operator  
**I want** quét mã đơn/mã vận đơn trước khi bàn giao  
**So that** hạn chế giao thiếu/sai đơn cho ĐVVC.

Acceptance Criteria:

- Scan đúng thì mark scanned.
- Scan sai manifest thì cảnh báo.
- Scan trùng thì cảnh báo.
- Thiếu đơn không được complete nếu không có exception.

### US-RET-001 — Nhận hàng hoàn

**As a** Return Inspector  
**I want** quét và ghi nhận hàng hoàn khi nhận từ shipper  
**So that** hàng hoàn không thất lạc và không nhập lại available sai.

Acceptance Criteria:

- Có return receipt.
- Nếu không nhận diện được đơn thì tạo unknown case.
- Hàng vào Pending Inspection.
- Có audit log.

### US-SUB-001 — Tạo đơn gia công ngoài

**As a** Subcontract Coordinator  
**I want** tạo đơn gia công với nhà máy  
**So that** quản lý số lượng, mẫu mã, NVL/bao bì, tiến độ và thanh toán.

Acceptance Criteria:

- Có factory, product, qty, spec, timeline.
- Có trạng thái lifecycle.
- Có attachment.
- Có approval nếu vượt rule.

---

# PHẦN L — PATCH CHO 10 MIGRATION / CUTOVER

## L1. Dữ liệu cần thêm trước go-live

| Data | Cần chuẩn bị |
|---|---|
| Carrier master | Mã ĐVVC, tên, contact, rule manifest |
| Warehouse zones | Khu pick, pack, handover, return, QC hold, lab/scrap |
| Return reason | Lý do hoàn, tình trạng, disposition |
| Subcontract factory | Danh sách nhà máy gia công |
| Subcontract locations | Vị trí stock tại nhà máy/subcontractor |
| Sample status | Danh mục trạng thái mẫu |
| Claim reason | Lý do claim nhà máy |

## L2. Cutover checkpoint mới

```text
- Carrier master loaded
- Warehouse zones loaded
- Return area/lab/scrap locations loaded
- Factory suppliers loaded
- Subcontract in-progress orders loaded if any
- Open return cases loaded if any
- Stock at subcontractor reconciled if any
- Opening manifests not migrated; only go-live-day manifests created in ERP
```

---

# PHẦN M — PATCH CHO 11–15 TECH/FRONTEND/UI STANDARDS

## M1. Backend Go architecture

Bổ sung rule:

```text
shipping module owns carrier_manifest and scan handover.
returns module owns return receipt/inspection/disposition.
subcontract module owns subcontract order lifecycle.
inventory module owns stock ledger and movement validation.
qc module owns QC decision.
```

Không module nào tự update stock balance trực tiếp. Mọi thay đổi tồn phải đi qua inventory stock ledger service.

## M2. Frontend architecture

Bổ sung route nhóm mới:

```text
/warehouse/daily-board
/warehouse/shifts/:id/closing
/packing/tasks
/shipping/manifests
/shipping/manifests/:id/scan
/returns/receiving
/returns/:id/inspection
/subcontract/orders
/subcontract/orders/:id
/subcontract/orders/:id/material-transfer
/subcontract/orders/:id/sample
/subcontract/orders/:id/inbound-qc
/subcontract/orders/:id/claims
```

## M3. UI/UX standards

Bổ sung pattern:

```text
Scan-first Screen Pattern
Exception-first Handover Pattern
Closing Checklist Pattern
Return Disposition Pattern
Subcontract Timeline Pattern
SLA Warning Badge
```

### Scan-first Screen Pattern

- Barcode input auto-focus.
- Recent scans hiển thị ngay dưới input.
- Kết quả scan dùng status rõ: matched/wrong/duplicate/missing.
- Có âm thanh hoặc visual feedback nếu thiết bị cho phép.
- Không bắt user dùng chuột trong luồng quét liên tục.

---

# PHẦN N — TRACEABILITY MATRIX V1.1

| Requirement | PRD | Flow | Screen | API | DB | UAT | Backlog |
|---|---|---|---|---|---|---|---|
| Warehouse Daily Board | FR-WH-DAY-001 | D1 | WH-DAILY-BOARD | H1 | I1 | G1 | EP-WH-DAY |
| Shift Closing | FR-WH-CLOSE-001 | D1 | WH-SHIFT-CLOSE | H2 | I1/I3 | G2 | EP-WH-DAY |
| Packing Queue | FR-PACK-001 | D2 | PACK-QUEUE/PACK-VERIFY | H3 | I1 | G3 partial | EP-PACK |
| Carrier Manifest | FR-SHIP-MANIFEST-001 | D3 | SHIP-MANIFEST | H4 | I1/I2/I3 | G3 | EP-SHIP-MANIFEST |
| Scan Handover | FR-SHIP-SCAN-001 | D3 | SHIP-SCAN-HANDOVER | H4 | scan_events | G3 | EP-SHIP-MANIFEST |
| Return Receiving | FR-RET-001 | D4 | RET-RECEIVE | H5 | return_receipts | G4 | EP-RETURNS |
| Return Disposition | FR-RET-002 | D4 | RET-INSPECT | H5 | return_dispositions | G4 | EP-RETURNS |
| Subcontract Order | FR-SUB-001 | D5 | SUB-ORDER | H6 | subcontract_orders | G5 | EP-SUBCONTRACT |
| Material Transfer to Factory | FR-SUB-TRANSFER-001 | D5 | SUB-MATERIAL-TRANSFER | H6 | material_transfers | G5 | EP-SUBCONTRACT |
| Sample Approval | FR-SUB-SAMPLE-001 | D5 | SUB-SAMPLE-APPROVAL | H6 | subcontract_samples | G5 | EP-SUBCONTRACT |
| Factory Inbound QC | FR-SUB-INBOUND-001 | D5 | SUB-INBOUND-QC | H6 | inbound_receipts | G5 | EP-SUBCONTRACT |
| Factory Claim SLA | FR-SUB-CLAIM-001 | D5 | SUB-FACTORY-CLAIM | H6 | factory_claims | G5 | EP-FACTORY-CLAIM |

---

# PHẦN O — SIGN-OFF CHECKLIST TRƯỚC SPRINT 0 / SPRINT 1

## O1. Business sign-off

| Checklist | Done |
|---|---|
| Warehouse Lead xác nhận daily board và closing flow đúng thực tế | ☐ |
| Kho xác nhận packing/manifest/scan handover đúng cách làm thật | ☐ |
| QA xác nhận hàng hoàn và factory inbound QC đủ kiểm soát | ☐ |
| Production/Subcontract xác nhận flow gia công ngoài đúng | ☐ |
| Finance xác nhận rule deposit/final payment/claim hold | ☐ |
| COO/CEO xác nhận P0/P1/P2 scope | ☐ |

## O2. Technical sign-off

| Checklist | Done |
|---|---|
| Tech Lead xác nhận module ownership | ☐ |
| Backend xác nhận API endpoints đủ cho flow mới | ☐ |
| DB owner xác nhận tables/index/status constraints | ☐ |
| Frontend xác nhận route/screen/component pattern | ☐ |
| QA xác nhận UAT/regression coverage | ☐ |
| DevOps xác nhận smoke test/cutover checkpoint | ☐ |

## O3. Rule xử lý xung đột

Nếu file cũ và file 33 mâu thuẫn, áp dụng rule:

```text
1. Nếu liên quan workflow thật đã xác nhận → file 33 thắng.
2. Nếu liên quan API/DB đã được Tech Lead phê duyệt sau file 33 → API/DB bản mới nhất thắng.
3. Nếu liên quan quyền/rủi ro → Permission Matrix + Security Standard thắng.
4. Nếu liên quan vận hành kho thực tế → As-Is + Gap Decision Log + file 33 thắng.
5. Nếu chưa rõ → tạo Change Request, không để dev tự đoán.
```

---

# PHẦN P — Kết luận

File 33 là gói cập nhật để khóa lại bản v1.1 của ERP Phase 1.

Bản v1.1 phải đảm bảo hệ thống không chỉ làm được ERP lõi, mà còn bám đúng nhịp vận hành thật:

```text
Kho mở ngày → xử lý đơn → soạn/đóng → bàn giao ĐVVC bằng quét mã → nhận hàng hoàn → kiểm kê/đối soát cuối ngày
```

và:

```text
Đặt gia công ngoài → chuyển NVL/bao bì → duyệt mẫu → sản xuất hàng loạt → nhận hàng → QC → nhập kho hoặc claim nhà máy trong 3–7 ngày
```

Nếu đội dự án chỉ đọc bản v1.0 mà bỏ qua file này, nguy cơ cao là build ra hệ thống “đúng theo ERP chung” nhưng chưa khớp cách công ty đang vận hành.

**Khuyến nghị tiếp theo:**

```text
34_ERP_Sprint0_Implementation_Kickoff_Plan_Phase1_MyPham_v1.md
```

File 34 sẽ biến toàn bộ tài liệu thành kế hoạch Sprint 0 thực thi: repo, môi trường, backend Go skeleton, frontend Next.js skeleton, PostgreSQL migration, OpenAPI base, RBAC, audit log, stock ledger prototype và scan prototype.
