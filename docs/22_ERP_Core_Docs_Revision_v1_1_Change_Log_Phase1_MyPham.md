# 22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham

**Dự án:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Giai đoạn:** Phase 1  
**Phiên bản tài liệu:** v1.0  
**Ngày:** 2026-04-24  
**Vai trò tài liệu:** Change Log / Revision Control cho bộ tài liệu lõi ERP Phase 1 sau khi phân tích workflow thực tế As-Is và Gap Analysis.

---

## 1. Mục đích tài liệu

Tài liệu này dùng để khóa lại toàn bộ những thay đổi cần cập nhật vào bộ tài liệu ERP lõi sau khi đã phân tích:

1. Bộ thiết kế ERP Phase 1 hiện tại.
2. Workflow thực tế kho, bàn giao đơn vị vận chuyển, hàng hoàn và gia công ngoài.
3. Gap Analysis giữa As-Is và To-Be.

Nói đơn giản:

```text
File 21 = phát hiện điểm lệch và ra quyết định.
File 22 = chỉ rõ tài liệu nào phải sửa, sửa cái gì, ưu tiên thế nào, ai duyệt.
```

Tài liệu này giúp tránh tình trạng cực nguy hiểm trong dự án ERP:

```text
PRD nói một kiểu
Process Flow nói một kiểu
API nói một kiểu
Database nói một kiểu
UI lại làm một kiểu khác
```

Từ thời điểm này, mọi thay đổi từ As-Is/Gap Analysis phải đi qua **Core Docs Revision Log** trước khi giao cho BA, UI/UX, dev, QA hoặc vendor.

---

## 2. Nguồn đầu vào

### 2.1. Tài liệu workflow thực tế

Các tài liệu As-Is dùng làm căn cứ:

- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

Các điểm workflow thực tế đã được ghi nhận:

- Kho có nhịp xử lý công việc hằng ngày: tiếp nhận đơn, xuất/nhập, soạn/đóng hàng, tối ưu vị trí kho, kiểm kê cuối ngày, đối soát và kết thúc ca.
- Nội quy kho đang tách rõ nhập kho, xuất kho, đóng hàng và xử lý hàng hoàn.
- Bàn giao đơn vị vận chuyển có phân khu để hàng, để theo thùng/rổ, đối chiếu số lượng, quét mã trực tiếp và xử lý trường hợp chưa đủ đơn.
- Quy trình sản xuất hiện tại nghiêng về gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao về kho, kiểm tra, nhận hàng hoặc báo lỗi trong 3–7 ngày.

### 2.2. Tài liệu ERP đã có

Nhóm tài liệu cần kiểm tra và cập nhật:

- `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
- `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`
- `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1.md`
- `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`
- `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`
- `10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1.md`
- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`
- `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md`
- `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`
- `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`
- `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`
- `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`
- `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md`
- `21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md`

---

## 3. Kết luận revision v1.1

Sau Gap Analysis, Phase 1 cần cập nhật trọng tâm từ:

```text
ERP lõi chuẩn:
Mua hàng → QC → sản xuất → kho → bán hàng → giao hàng
```

sang:

```text
ERP vận hành thực tế:
Kho hằng ngày → nhập/xuất/đóng hàng → bàn giao ĐVVC → hàng hoàn → gia công ngoài → QC nhận hàng → đối soát cuối ngày
```

Bản revision v1.1 phải bổ sung 7 cụm nghiệp vụ bắt buộc:

1. **Warehouse Daily Board** — bảng điều hành kho theo ngày.
2. **Shift Closing / End-of-Day Reconciliation** — đóng ca, kiểm kê và đối soát cuối ngày.
3. **Packing Queue & Pack Verification** — hàng chờ đóng, xác nhận đóng đúng SKU/số lượng/tình trạng.
4. **Carrier Manifest & Scan Handover** — bảng bàn giao theo đơn vị vận chuyển, quét mã xác nhận đủ/thiếu.
5. **Returns Receiving & Return Disposition** — nhận hàng hoàn, quét hàng hoàn, kiểm tình trạng, phân loại còn dùng/không dùng.
6. **Subcontract Manufacturing** — gia công ngoài, chuyển NVL/bao bì, duyệt mẫu, nhận hàng, QC và claim nhà máy.
7. **Factory Claim SLA 3–7 days** — thời hạn phản hồi lỗi nhà máy sau khi nhận hàng.

---

## 4. Mức ưu tiên revision

| Mức | Ý nghĩa | Ví dụ |
|---|---|---|
| `P0` | Bắt buộc sửa trước khi dev build/sprint chính | Warehouse Daily Board, Scan Handover, Returns, Subcontract Manufacturing |
| `P1` | Nên sửa trước UAT hoặc trước go-live | Attachment chứng từ, KPI nâng cao, SOP mapping |
| `P2` | Có thể đưa vào Phase 2 hoặc enhancement | MES nội bộ sâu, automation nâng cao |

Nguyên tắc:

```text
P0 không sửa = dev build sai lõi vận hành.
P1 không sửa = go-live dễ thiếu kiểm soát.
P2 không sửa = chưa nguy hiểm cho Phase 1.
```

---

## 5. Revision Matrix tổng thể

| Doc ID | Tài liệu | Mức ảnh hưởng | Cần xuất bản v1.1? | Ưu tiên | Lý do |
|---|---|---:|---|---|---|
| 03 | PRD/SRS Phase 1 | Rất cao | Có | P0 | Scope phải bổ sung daily warehouse, returns, carrier manifest, subcontract manufacturing |
| 04 | Permission/Approval Matrix | Cao | Có | P0 | Cần thêm quyền trưởng kho, scan handover, return disposition, factory claim, payment hold |
| 05 | Data Dictionary | Rất cao | Có | P0 | Cần thêm field/status/entity mới cho shift, manifest, returns, subcontract |
| 06 | Process Flow To-Be | Rất cao | Có | P0 | Luồng nghiệp vụ phải chỉnh theo As-Is thực tế |
| 07 | Report/KPI Catalog | Trung bình-cao | Có | P1 | Cần KPI kho cuối ngày, handover discrepancy, return SLA, factory claim SLA |
| 08 | Screen List/Wireframe | Rất cao | Có | P0 | Cần thêm màn hình daily board, shift closing, manifest scan, return inspection, subcontract order |
| 09 | UAT Test Scenarios | Cao | Có | P0 | Cần test các flow mới và exception |
| 10 | Data Migration/Cutover | Trung bình | Có | P1 | Cần thêm opening data cho carrier, manifest, return stock, subcontractor/location |
| 11 | Technical Architecture Go Backend | Cao | Có | P1 | Cần đảm bảo module/submodule đúng với flow thực tế |
| 12 | Go Coding Standards | Trung bình | Có nhẹ | P1 | Cần bổ sung guardrail cho use case mới nếu chưa đủ |
| 13 | Go Module/Component Standards | Cao | Có | P1 | Cần rõ ownership và boundary của shipping/returns/subcontract |
| 14 | UI/UX Design System | Trung bình-cao | Có | P1 | Cần chuẩn UX scan-first, exception-first, closing checklist |
| 15 | Frontend Architecture | Trung bình-cao | Có | P1 | Cần route/module/page mới |
| 16 | API Contract/OpenAPI | Rất cao | Có | P0 | Cần endpoint mới cho manifest, scan handover, shift closing, returns, subcontract |
| 17 | Database Schema PostgreSQL | Rất cao | Có | P0 | Cần bảng/field/index/constraint mới |
| 18 | DevOps/CI-CD | Thấp-trung bình | Có nhẹ | P1 | Cần smoke test và migration checkpoint cho flow mới |
| 19 | Security/RBAC/Audit | Trung bình-cao | Có nhẹ | P1 | Cần field/action audit cho chứng từ nhạy cảm mới |
| 20 | As-Is Workflow | Thấp | Không bắt buộc | - | Là nguồn tham chiếu, không cần sửa nếu không phát hiện As-Is mới |
| 21 | Gap Analysis/Decision Log | Thấp | Không bắt buộc | - | Là nguồn quyết định; chỉ sửa khi có decision mới |

---

## 6. Core Change Request Register

### CR-001 — Bổ sung Warehouse Daily Board

**Nguồn:** `G-001`, `D-001` trong Gap Analysis  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `06`, `07`, `08`, `09`, `16`, `17`, `18`  

**Mô tả:**  
Bổ sung bảng điều hành kho trong ngày để Warehouse Lead và nhân viên kho theo dõi toàn bộ việc cần xử lý.

**Yêu cầu cập nhật:**

- Thêm use case `WH-DAY-001 View Warehouse Daily Board`.
- Thêm trạng thái việc trong ngày: pending, in_progress, blocked, completed, exception.
- Thêm chỉ số đơn chờ pick/pack/handover, hàng hoàn chờ xử lý, phiếu nhập chờ QC, variance cuối ngày.
- Thêm màn hình `Warehouse Daily Board`.
- Thêm API `/warehouse/daily-board`.
- Thêm query/tables phục vụ aggregation từ orders, pick/pack, shipments, returns, inbound, stock count.

**Acceptance criteria:**

- Warehouse Lead mở màn hình thấy được toàn bộ việc trong ngày.
- Có drill-down từ số tổng xuống danh sách chứng từ.
- Có cảnh báo blocked/overdue.
- Không cần refresh tay liên tục, tối thiểu có auto refresh hoặc refresh action.

---

### CR-002 — Bổ sung Shift Closing / End-of-Day Reconciliation

**Nguồn:** `G-012`, `G-013`, `D-002`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `04`, `05`, `06`, `07`, `08`, `09`, `16`, `17`, `19`  

**Mô tả:**  
Kho phải có quy trình đóng ca/kết ngày sau khi kiểm kê, đối soát và báo cáo quản lý.

**Yêu cầu cập nhật:**

- Thêm entity `warehouse_shift`.
- Thêm entity `shift_closing_checklist`.
- Thêm entity `shift_variance`.
- Thêm quyền `warehouse.shift.close` cho Warehouse Lead.
- Thêm workflow:

```text
Open Shift
→ Process daily warehouse tasks
→ Cycle Count / Stock Count
→ Reconcile shipments, returns, inbound, outbound
→ Record variance
→ Warehouse Lead sign-off
→ Close Shift
```

**Gate đóng ca:**

- Không còn shipment pending không lý do.
- Hàng hoàn trong ngày đã được scan hoặc ghi exception.
- Phiếu xuất trong ngày có trạng thái rõ.
- Variance tồn kho đã được ghi nhận.
- Trưởng kho xác nhận.

**Acceptance criteria:**

- Không cho đóng ca nếu còn checklist bắt buộc chưa xong.
- Hệ thống ghi audit log ai đóng ca, lúc nào, số liệu trước/sau.
- Có báo cáo EOD Warehouse Report.

---

### CR-003 — Tách rõ nhập kho: Receiving → QC → Putaway / Reject

**Nguồn:** `G-002`, `D-003`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `16`, `17`  

**Mô tả:**  
Nhập kho không được hiểu là hàng đã khả dụng ngay. Cần tách rõ nhận hàng, kiểm tra, QC, xếp kho hoặc trả/hold.

**Workflow v1.1:**

```text
Receive delivery document
→ Check quantity / packaging / lot / expiry
→ Create receiving record
→ QC Hold
→ QC Pass / Fail
→ if Pass: Putaway to available stock
→ if Fail: Reject Supplier / Quarantine / Return Supplier
```

**Data cần thêm/cập nhật:**

- `receiving_type`: purchase, subcontract_fg, return, adjustment.
- `receiving_status`: draft, received, qc_hold, qc_passed, qc_failed, putaway_completed, rejected, cancelled.
- `putaway_status`.
- `reject_reason`.
- `supplier_return_ref`.

**Acceptance criteria:**

- Hàng chưa QC Pass không được vào tồn khả dụng.
- Có thể upload chứng từ giao hàng/ảnh bằng chứng.
- Có log số lượng thực nhận so với số lượng chứng từ.

---

### CR-004 — Số hóa phiếu nhập/xuất và mapping chứng từ giấy

**Nguồn:** `G-003`, `G-004`, `D-004`  
**Ưu tiên:** P1  
**Ảnh hưởng:** `05`, `08`, `16`, `17`, `19`, `26`  

**Mô tả:**  
Workflow hiện tại có ký xác nhận và lưu phiếu nhập/xuất. ERP v1.1 phải số hóa nhưng vẫn cho mapping chứng từ giấy nếu công ty còn dùng.

**Trường cần thêm:**

- `paper_document_no`.
- `signed_by_name`.
- `signed_at`.
- `attachment_required`.
- `attachment_type`: delivery_note, signed_receipt, signed_issue, photo, video, other.

**Rule:**

- Phiếu nhập/xuất P0 có thể yêu cầu attachment bắt buộc theo cấu hình.
- Không cho xóa attachment sau khi chứng từ đã approved/closed, chỉ được thêm bản bổ sung với audit log.

---

### CR-005 — Bổ sung Packing Queue và Pack Verification

**Nguồn:** `G-005`, `G-006`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `16`, `17`  

**Mô tả:**  
Đóng hàng thực tế có phân loại theo ĐVVC/đơn/lớp đóng gói, soạn từng đơn, kiểm SKU/số lượng/tình trạng tại khu đóng hàng.

**Workflow v1.1:**

```text
Order ready to pack
→ Add to Packing Queue
→ Filter/group by carrier/date/package type
→ Pick items by order
→ Verify SKU/quantity/batch/condition
→ Pack order
→ Move to carrier handover zone
```

**Data cần thêm:**

- `packing_task`.
- `packing_status`: pending, picking, packed, failed_verification, moved_to_handover_zone, cancelled.
- `package_type`.
- `packing_zone`.
- `pack_verified_by`.
- `pack_verified_at`.
- `packing_defect_reason`.

**Acceptance criteria:**

- Đơn chưa pack verified không được vào manifest bàn giao.
- Sai SKU/số lượng phải tạo exception.
- Có audit log người pack và người verify.

---

### CR-006 — Bổ sung Carrier Manifest và Scan Handover

**Nguồn:** `G-007`, `G-008`, `G-029`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `14`, `15`, `16`, `17`, `19`, `23`  

**Mô tả:**  
Quy trình bàn giao cho ĐVVC cần được thiết kế như một nghiệp vụ riêng, không chỉ update shipment status.

**Workflow v1.1:**

```text
Create carrier manifest
→ Assign packed orders to manifest
→ Put orders into carrier zone / bin / basket
→ Compare expected quantity
→ Scan order/tracking barcode at handover
→ if complete: confirm handover and sign
→ if missing: mark exception and investigate
```

**Entity/API cần thêm:**

- `carrier_manifest`.
- `carrier_manifest_item`.
- `handover_scan_event`.
- `handover_exception`.
- API `POST /shipping/manifests`.
- API `POST /shipping/manifests/{id}/scan`.
- API `POST /shipping/manifests/{id}/confirm-handover`.

**Exception flow:**

- Tracking không thuộc manifest.
- Đơn đã scan rồi.
- Đơn chưa packed.
- Đơn thiếu khi đối chiếu số lượng.
- Đơn có trên hệ thống nhưng không tìm thấy ở khu đóng hàng.
- Mã không có trên hệ thống, cần đóng lại/cập nhật lại.

**Acceptance criteria:**

- Manifest chỉ close khi số lượng expected = scanned hoặc có exception được duyệt/ghi nhận.
- Handover phải lưu người xác nhận, thời gian, ĐVVC, số đơn, số thiếu/thừa.
- Có màn hình quét mã tối ưu cho thao tác nhanh.

---

### CR-007 — Bổ sung Returns Receiving & Return Inspection

**Nguồn:** `G-009`, `G-010`, `G-011`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `14`, `15`, `16`, `17`, `19`  

**Mô tả:**  
Hàng hoàn không được nhập kho như hàng thường. Cần có khu hàng hoàn, scan, ghi nhận tình trạng, kiểm tra bên trong, phân loại còn dùng/không dùng.

**Workflow v1.1:**

```text
Receive return from shipper
→ Move to return area
→ Scan return order/tracking
→ Capture evidence: photo/video/note
→ Inspect outer condition and inner items
→ classify: reusable / not_reusable / needs_lab_review / missing_item / wrong_item
→ if reusable: move to return available/quarantine stock depending QC rule
→ if not reusable: move to damaged/lab
→ create return receipt
```

**Data/status cần thêm:**

- `return_receipt`.
- `return_inspection`.
- `return_disposition`.
- `return_evidence`.
- `return_status`: received, scanned, inspecting, reusable, not_reusable, sent_to_lab, returned_to_available, closed.
- `return_condition`: good, dented, leaking, broken, opened, missing, wrong_item, other.

**Rule:**

- Hàng hoàn mỹ phẩm không tự động quay lại tồn bán được.
- Hàng hoàn phải có bằng chứng nếu không sử dụng hoặc gửi Lab.
- Batch/hạn dùng phải được ghi nhận nếu hàng còn định danh được.

---

### CR-008 — Bổ sung Subcontract Manufacturing làm trọng tâm Phase 1

**Nguồn:** `G-014`, `G-015`, `G-016`, `G-017`, `G-018`, `G-019`, `G-020`, `G-021`, `G-022`, `G-030`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `03`, `04`, `05`, `06`, `07`, `08`, `09`, `11`, `13`, `16`, `17`, `19`, `23`  

**Mô tả:**  
Sản xuất Phase 1 phải phản ánh mô hình gia công ngoài thực tế: công ty làm việc với nhà máy, chuyển NVL/bao bì, duyệt mẫu, nhận hàng và QC.

**Workflow v1.1:**

```text
Create subcontract production order
→ Confirm quantity / specification / sample / timeline
→ Deposit/payment milestone if needed
→ Transfer raw materials / packaging to factory
→ Sign handover document with attachments
→ Factory makes sample
→ Company approves sample
→ Factory mass production
→ Factory sends QC/report if any
→ Deliver finished goods to warehouse
→ Company checks quantity/quality
→ if accepted: receive + QC + putaway
→ if rejected/issue: create factory claim within 3–7 days
→ final payment release after acceptance
```

**Entity cần thêm:**

- `subcontract_order`.
- `subcontract_order_item`.
- `subcontract_material_transfer`.
- `subcontract_material_transfer_item`.
- `factory_sample_approval`.
- `retained_sample`.
- `factory_delivery`.
- `factory_claim`.
- `subcontract_payment_milestone`.

**Status cần thêm:**

- `subcontract_order_status`: draft, confirmed, deposit_pending, material_preparing, material_transferred, sample_pending, sample_approved, mass_production, factory_delivered, under_receiving_qc, accepted, claim_opened, closed, cancelled.
- `sample_status`: not_started, in_progress, submitted, approved, rejected, revised.
- `factory_claim_status`: draft, submitted, acknowledged, resolving, resolved, rejected, closed.

**Rule quan trọng:**

- Không cho mass production nếu sample approval bắt buộc nhưng chưa approved.
- Không cho final payment nếu hàng chưa accepted/QC passed hoặc chưa có approval override.
- NVL/bao bì chuyển sang nhà máy phải đi qua stock movement/transfer rõ ràng, không xóa khỏi hệ thống theo kiểu “mất dấu”.
- Factory claim phải có SLA due date 3–7 ngày theo policy.

---

### CR-009 — Bắt buộc batch/lô/hạn dùng ở receiving, returns và factory receipt

**Nguồn:** `G-023`, `G-024`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `05`, `08`, `09`, `16`, `17`, `19`  

**Mô tả:**  
Ngành mỹ phẩm phải quản theo batch/lô/hạn dùng. As-Is có bước kiểm lô khi nhập kho, To-Be đã có batch/expiry nhưng cần enforce mạnh hơn ở mọi flow liên quan tồn.

**Rule v1.1:**

- Thành phẩm, nguyên liệu, hàng gia công nhận về, hàng hoàn đủ điều kiện đều phải ghi batch/lô nếu mặt hàng cấu hình `batch_required = true`.
- Mặt hàng có `expiry_required = true` bắt buộc nhập hạn dùng.
- Không cho xuất bán batch đang `qc_hold`, `qc_failed`, `quarantine`, `expired`, `blocked`.
- Hệ thống cảnh báo cận date theo rule cấu hình.

---

### CR-010 — Bổ sung payment hold/release theo QC acceptance cho gia công ngoài

**Nguồn:** `G-022`  
**Ưu tiên:** P1  
**Ảnh hưởng:** `03`, `04`, `05`, `06`, `16`, `17`, `19`  

**Mô tả:**  
Thanh toán cuối cho nhà máy nên được khóa theo trạng thái nhận hàng/QC để tránh thanh toán khi hàng chưa đạt.

**Rule v1.1:**

- `final_payment_status = blocked` nếu factory delivery chưa accepted.
- Finance có thể thấy lý do blocked.
- Override phải cần quyền đặc biệt và ghi audit log.
- Nếu claim mở, final payment chuyển sang `on_hold` cho tới khi claim resolved hoặc có duyệt override.

---

### CR-011 — Bổ sung role Warehouse Lead và quyền sign-off vận hành kho

**Nguồn:** `G-026`  
**Ưu tiên:** P0  
**Ảnh hưởng:** `04`, `08`, `16`, `19`  

**Mô tả:**  
Workflow As-Is thể hiện trưởng kho có vai trò ký/lưu phiếu nhập/xuất. ERP cần tách quyền Warehouse Staff và Warehouse Lead.

**Quyền mới:**

- `warehouse.shift.close`.
- `warehouse.variance.approve`.
- `warehouse.issue.signoff`.
- `warehouse.receipt.signoff`.
- `shipping.manifest.confirm_handover`.
- `returns.disposition.approve`.

**Rule:**

- Nhân viên kho có thể thao tác pick/pack/scan.
- Warehouse Lead mới được xác nhận đóng ca, xác nhận exception quan trọng, ký số hóa bàn giao lớn.

---

### CR-012 — Bổ sung EOD Warehouse Report và KPI vận hành kho

**Nguồn:** `G-028`  
**Ưu tiên:** P1  
**Ảnh hưởng:** `07`, `08`, `16`, `17`  

**KPI cần thêm:**

- `orders_processed_today`.
- `orders_pending_handover`.
- `handover_discrepancy_count`.
- `return_received_count`.
- `return_pending_inspection_count`.
- `stock_variance_count`.
- `stock_variance_value`.
- `shift_closed_on_time_rate`.
- `packing_error_rate`.
- `factory_claim_open_count`.
- `factory_claim_overdue_count`.

---

## 7. Revision chi tiết theo từng tài liệu lõi

### 7.1. File 03 — PRD/SRS Phase 1 v1.1

**Mức ảnh hưởng:** Rất cao  
**Ưu tiên:** P0  
**Bắt buộc xuất bản v1.1:** Có

#### Nội dung cần sửa

1. Cập nhật phạm vi Phase 1:

```text
Bổ sung:
- Warehouse Daily Operations
- Shift Closing / EOD Reconciliation
- Packing Queue & Verification
- Carrier Manifest / Scan Handover
- Return Receiving / Return Inspection / Return Disposition
- Subcontract Manufacturing
- Factory Claim SLA

Giảm hoặc defer:
- MES nội bộ quá sâu nếu chưa vận hành thực tế.
```

2. Thêm epic/use case:

| Use Case ID | Tên use case | Ưu tiên |
|---|---|---|
| WH-DAY-001 | View Warehouse Daily Board | P0 |
| WH-SHIFT-001 | Open Warehouse Shift | P0 |
| WH-SHIFT-002 | Close Warehouse Shift | P0 |
| PACK-001 | Create Packing Task | P0 |
| PACK-002 | Verify Packed Order | P0 |
| SHIP-MAN-001 | Create Carrier Manifest | P0 |
| SHIP-MAN-002 | Scan Order for Handover | P0 |
| SHIP-MAN-003 | Confirm Carrier Handover | P0 |
| RET-001 | Receive Returned Goods | P0 |
| RET-002 | Inspect Returned Goods | P0 |
| RET-003 | Decide Return Disposition | P0 |
| SUB-001 | Create Subcontract Production Order | P0 |
| SUB-002 | Transfer Materials to Factory | P0 |
| SUB-003 | Approve Factory Sample | P0 |
| SUB-004 | Receive Factory Finished Goods | P0 |
| SUB-005 | Create Factory Claim | P0 |

3. Thêm non-functional/operational rule:

- Scan API phải phản hồi nhanh, phù hợp thao tác kho.
- Mọi action đóng ca/bàn giao/return disposition/payment override phải có audit log.
- Không cho bán/xuất batch chưa hợp lệ.

#### Acceptance của revision file 03

- PRD phải có scope mới và flow rõ cho 7 cụm nghiệp vụ.
- Mỗi use case P0 phải có actor, precondition, main flow, exception, data, acceptance criteria.
- Không còn mô tả sản xuất Phase 1 như xưởng nội bộ 100% nếu chưa đúng thực tế.

---

### 7.2. File 04 — Permission & Approval Matrix v1.1

**Mức ảnh hưởng:** Cao  
**Ưu tiên:** P0  

#### Quyền/role cần thêm

| Role | Quyền mới |
|---|---|
| Warehouse Staff | pack order, scan handover, receive return, upload evidence |
| Warehouse Lead | close shift, approve variance, confirm handover, approve return disposition |
| QC Staff | inspect returned goods, inspect factory goods, hold/pass/fail batch |
| Production/Subcontract Coordinator | create subcontract order, manage factory timeline, sample approval request |
| Finance | manage deposit/final payment, release/hold payment based on acceptance |
| Admin/Security | break-glass, audit review, permission config |

#### Approval rule cần thêm

| Giao dịch | Người tạo | Người duyệt | Rule |
|---|---|---|---|
| Close warehouse shift | Warehouse Lead | COO optional | Bắt buộc nếu variance vượt ngưỡng |
| Handover exception | Warehouse Staff | Warehouse Lead | Thiếu/thừa đơn phải được xác nhận |
| Return not reusable | Warehouse Staff/QC | Warehouse Lead/QC Lead | Cần evidence |
| Subcontract order | Production Coordinator | COO/CEO theo giá trị | Nếu có deposit phải Finance review |
| Material transfer to factory | Warehouse/Production | Warehouse Lead + Production Lead | Cần biên bản bàn giao |
| Final payment to factory | Finance | CFO/CEO theo ngưỡng | Chỉ release khi QC accepted hoặc override |

---

### 7.3. File 05 — Data Dictionary & Master Data v1.1

**Mức ảnh hưởng:** Rất cao  
**Ưu tiên:** P0

#### Entity cần thêm

- `warehouse_shift`
- `shift_closing_checklist`
- `shift_variance`
- `packing_task`
- `packing_task_item`
- `carrier_manifest`
- `carrier_manifest_item`
- `handover_scan_event`
- `handover_exception`
- `return_receipt`
- `return_inspection`
- `return_evidence`
- `subcontract_order`
- `subcontract_order_item`
- `subcontract_material_transfer`
- `factory_sample_approval`
- `factory_delivery`
- `factory_claim`
- `retained_sample`

#### Enum/status cần thêm

```text
warehouse_shift_status:
open, closing, closed, reopened, cancelled

packing_status:
pending, picking, packed, failed_verification, moved_to_handover_zone, cancelled

manifest_status:
draft, ready_for_scan, scanning, discrepancy, handed_over, closed, cancelled

return_status:
received, scanned, inspecting, reusable, not_reusable, sent_to_lab, returned_to_available, closed

return_disposition:
reusable, quarantine, damaged, lab_review, scrap, supplier_claim

subcontract_order_status:
draft, confirmed, deposit_pending, material_preparing, material_transferred, sample_pending, sample_approved, mass_production, factory_delivered, under_receiving_qc, accepted, claim_opened, closed, cancelled

factory_claim_status:
draft, submitted, acknowledged, resolving, resolved, rejected, closed, overdue
```

#### Công thức/derived metrics cần thêm

```text
EOD variance value = abs(system_qty - counted_qty) * unit_cost

Handover discrepancy rate = manifests_with_discrepancy / total_manifests

Return inspection SLA rate = returns_inspected_within_sla / total_returns_received

Factory claim overdue = factory_claim.due_at < now AND status not in (resolved, closed)
```

---

### 7.4. File 06 — Process Flow To-Be v1.1

**Mức ảnh hưởng:** Rất cao  
**Ưu tiên:** P0

#### Flow cần thêm/sửa

1. Warehouse Daily Operations Flow.
2. Shift Closing / EOD Reconciliation Flow.
3. Inbound Receiving with QC/Putaway/Reject Flow.
4. Packing Queue and Verification Flow.
5. Carrier Manifest and Scan Handover Flow.
6. Handover Missing Order Exception Flow.
7. Return Receiving and Inspection Flow.
8. Return Disposition Flow.
9. Subcontract Manufacturing End-to-End Flow.
10. Material Transfer to Factory Flow.
11. Factory Sample Approval Flow.
12. Factory Finished Goods Receiving & QC Flow.
13. Factory Claim within 3–7 Days Flow.
14. Final Payment Hold/Release Flow.

#### Flow phải giảm trọng tâm

- Sản xuất nội bộ/MES chi tiết sâu nên đưa về Phase 2 nếu chưa vận hành thực tế.

---

### 7.5. File 07 — Report & KPI Catalog v1.1

**Mức ảnh hưởng:** Trung bình-cao  
**Ưu tiên:** P1

#### Báo cáo cần thêm

| Report ID | Tên báo cáo | Đối tượng |
|---|---|---|
| WH-RPT-001 | Warehouse Daily Closing Report | Warehouse Lead/COO |
| WH-RPT-002 | Stock Variance Report | Warehouse/Finance |
| PACK-RPT-001 | Packing Throughput Report | Warehouse Lead |
| SHIP-RPT-001 | Carrier Manifest Handover Report | Warehouse/CSKH |
| SHIP-RPT-002 | Handover Discrepancy Report | COO/Warehouse Lead |
| RET-RPT-001 | Return Receiving & Inspection Report | CSKH/Warehouse/QC |
| SUB-RPT-001 | Subcontract Order Tracking Report | Production/COO |
| SUB-RPT-002 | Factory Claim SLA Report | QA/Production/COO |

---

### 7.6. File 08 — Screen List & Wireframe v1.1

**Mức ảnh hưởng:** Rất cao  
**Ưu tiên:** P0

#### Màn hình cần thêm

| Screen ID | Tên màn hình | Module | Priority |
|---|---|---|---|
| WH-DB-01 | Warehouse Daily Board | Warehouse | P0 |
| WH-SHIFT-01 | Open/Close Shift | Warehouse | P0 |
| WH-SHIFT-02 | Shift Variance Review | Warehouse | P0 |
| PACK-01 | Packing Queue | Warehouse/Packing | P0 |
| PACK-02 | Pack Verification | Warehouse/Packing | P0 |
| SHIP-01 | Carrier Manifest List | Shipping | P0 |
| SHIP-02 | Manifest Detail | Shipping | P0 |
| SHIP-03 | Scan Handover | Shipping/Warehouse | P0 |
| SHIP-04 | Handover Exception Review | Shipping/Warehouse | P0 |
| RET-01 | Return Receiving | Returns | P0 |
| RET-02 | Return Inspection | Returns/QC | P0 |
| RET-03 | Return Disposition | Returns/Warehouse | P0 |
| SUB-01 | Subcontract Order List | Production | P0 |
| SUB-02 | Subcontract Order Detail | Production | P0 |
| SUB-03 | Material Transfer to Factory | Production/Warehouse | P0 |
| SUB-04 | Factory Sample Approval | Production/R&D/QC | P0 |
| SUB-05 | Factory Delivery Receiving | Warehouse/QC | P0 |
| SUB-06 | Factory Claim | Production/QC | P0 |

#### UX bắt buộc

- Màn hình scan phải ưu tiên keyboard/barcode scanner.
- Màn hình đóng ca phải dùng checklist rõ ràng.
- Màn hình exception phải nổi bật trạng thái thiếu/thừa/sai.
- Hàng hoàn phải có block upload evidence.
- Gia công ngoài phải có timeline và milestone.

---

### 7.7. File 09 — UAT Test Scenarios v1.1

**Mức ảnh hưởng:** Cao  
**Ưu tiên:** P0

#### Test case cần thêm

| UAT ID | Scenario | Priority |
|---|---|---|
| UAT-WH-001 | Warehouse Lead xem daily board và drill-down đơn chờ xử lý | P0 |
| UAT-WH-002 | Đóng ca thành công khi checklist đủ | P0 |
| UAT-WH-003 | Không cho đóng ca nếu còn shipment chưa xử lý | P0 |
| UAT-PACK-001 | Pack order đúng SKU/số lượng | P0 |
| UAT-PACK-002 | Pack verification fail khi sai SKU | P0 |
| UAT-SHIP-001 | Tạo manifest và scan đủ đơn bàn giao | P0 |
| UAT-SHIP-002 | Scan mã không thuộc manifest → báo lỗi | P0 |
| UAT-SHIP-003 | Thiếu đơn khi bàn giao → tạo exception | P0 |
| UAT-RET-001 | Nhận hàng hoàn, quét mã, upload evidence | P0 |
| UAT-RET-002 | Phân loại hàng hoàn còn dùng → chuyển kho | P0 |
| UAT-RET-003 | Phân loại không dùng → chuyển Lab/kho hỏng | P0 |
| UAT-SUB-001 | Tạo đơn gia công ngoài | P0 |
| UAT-SUB-002 | Chuyển NVL/bao bì cho nhà máy | P0 |
| UAT-SUB-003 | Không cho sản xuất hàng loạt khi mẫu chưa duyệt | P0 |
| UAT-SUB-004 | Nhận hàng gia công và QC pass nhập kho | P0 |
| UAT-SUB-005 | Hàng không đạt → tạo factory claim có SLA | P0 |
| UAT-FIN-001 | Final payment bị hold khi hàng chưa accepted | P1 |

---

### 7.8. File 10 — Data Migration/Cutover Plan v1.1

**Mức ảnh hưởng:** Trung bình  
**Ưu tiên:** P1

#### Dữ liệu cần migration/check thêm

- Carrier master.
- Carrier service type.
- Warehouse zones: receiving, picking, packing, handover, return, quarantine, damaged/lab.
- Bin/basket/rack nếu có.
- Opening return stock.
- Current pending returns.
- Current pending shipments/handover nếu go-live giữa ngày.
- Subcontractor/factory master.
- Existing subcontract orders đang mở.
- Material/packaging currently at factory.
- Retained sample records nếu có.
- Factory claim open cases nếu có.

#### Cutover rule cần thêm

- Không go-live shipping nếu chưa map được ĐVVC hiện tại.
- Không go-live warehouse nếu chưa có zone tối thiểu.
- Không go-live subcontract nếu chưa ghi nhận NVL/bao bì đang nằm tại nhà máy.

---

### 7.9. File 11 — Technical Architecture Go Backend v1.1

**Mức ảnh hưởng:** Cao  
**Ưu tiên:** P1

#### Cần cập nhật

- Bổ sung submodules:

```text
warehouse_daily_ops
shift_closing
packing
carrier_manifest
returns
subcontract_manufacturing
factory_claim
```

- Bổ sung event:

```text
WarehouseShiftClosed
PackingTaskCompleted
CarrierManifestCreated
CarrierHandoverConfirmed
HandoverExceptionCreated
ReturnReceived
ReturnDispositionDecided
SubcontractOrderConfirmed
MaterialTransferredToFactory
FactorySampleApproved
FactoryDeliveryReceived
FactoryClaimOpened
FinalPaymentReleased
```

- Bổ sung transaction boundary cho:
  - scan handover,
  - close shift,
  - return disposition,
  - material transfer to factory,
  - factory delivery receiving.

---

### 7.10. File 12 — Go Coding Standards v1.1

**Mức ảnh hưởng:** Trung bình  
**Ưu tiên:** P1

#### Cần bổ sung guardrail

- Không viết API scan handover không idempotent.
- Không update shipment/order status ngoài transaction.
- Không cho direct update stock balance, phải qua stock movement/ledger.
- Không cho return disposition thiếu evidence nếu policy yêu cầu.
- Không cho final payment release khi order chưa accepted nếu không có override permission.

---

### 7.11. File 13 — Module/Component Standards v1.1

**Mức ảnh hưởng:** Cao  
**Ưu tiên:** P1

#### Module boundary cần rõ

| Module | Ownership |
|---|---|
| Inventory | Stock ledger, stock balance, batch movement |
| Warehouse Daily Ops | Daily board, shift, closing checklist |
| Packing | Packing task, pack verification |
| Shipping | Carrier manifest, handover scan, delivery status |
| Returns | Return receiving, inspection, disposition |
| Subcontract Manufacturing | Subcontract order, material transfer, factory sample, factory claim |
| Finance | Payment hold/release, AP milestone |
| QC | QC inspection, batch hold/pass/fail |

Rule:

```text
Shipping không tự trừ tồn.
Returns không tự đưa hàng về available nếu QC policy chưa cho phép.
Finance không tự release final payment khi acceptance chưa đạt, trừ override có audit.
```

---

### 7.12. File 14 — UI/UX Design System v1.1

**Mức ảnh hưởng:** Trung bình-cao  
**Ưu tiên:** P1

#### Cần cập nhật pattern

- Scan-first screen pattern.
- Exception handling banner.
- Warehouse closing checklist component.
- Return evidence upload component.
- Carrier manifest progress component.
- Factory production timeline component.
- SLA overdue badge.

---

### 7.13. File 15 — Frontend Architecture v1.1

**Mức ảnh hưởng:** Trung bình-cao  
**Ưu tiên:** P1

#### Routes cần thêm

```text
/warehouse/daily-board
/warehouse/shifts
/warehouse/shifts/:id/close
/packing/queue
/packing/tasks/:id/verify
/shipping/manifests
/shipping/manifests/:id
/shipping/manifests/:id/scan
/returns/receiving
/returns/:id/inspection
/subcontract/orders
/subcontract/orders/:id
/subcontract/orders/:id/material-transfer
/subcontract/orders/:id/sample-approval
/subcontract/orders/:id/factory-delivery
/subcontract/orders/:id/claims
```

---

### 7.14. File 16 — API Contract/OpenAPI v1.1

**Mức ảnh hưởng:** Rất cao  
**Ưu tiên:** P0

#### Endpoint cần thêm

```text
GET    /api/v1/warehouse/daily-board
POST   /api/v1/warehouse/shifts
GET    /api/v1/warehouse/shifts/{id}
POST   /api/v1/warehouse/shifts/{id}/close
POST   /api/v1/packing/tasks
POST   /api/v1/packing/tasks/{id}/verify
POST   /api/v1/shipping/manifests
GET    /api/v1/shipping/manifests/{id}
POST   /api/v1/shipping/manifests/{id}/scan
POST   /api/v1/shipping/manifests/{id}/confirm-handover
POST   /api/v1/shipping/manifests/{id}/exceptions
POST   /api/v1/returns/receipts
POST   /api/v1/returns/receipts/{id}/scan
POST   /api/v1/returns/receipts/{id}/inspect
POST   /api/v1/returns/receipts/{id}/decide-disposition
POST   /api/v1/subcontract/orders
GET    /api/v1/subcontract/orders/{id}
POST   /api/v1/subcontract/orders/{id}/confirm
POST   /api/v1/subcontract/orders/{id}/material-transfers
POST   /api/v1/subcontract/orders/{id}/sample-approval
POST   /api/v1/subcontract/orders/{id}/factory-deliveries
POST   /api/v1/subcontract/orders/{id}/claims
```

#### Error code cần thêm

```text
SHIFT_CANNOT_CLOSE
SHIFT_HAS_PENDING_TASKS
MANIFEST_ORDER_NOT_PACKED
MANIFEST_SCAN_DUPLICATED
MANIFEST_ITEM_MISSING
RETURN_EVIDENCE_REQUIRED
RETURN_BATCH_REQUIRED
SUBCONTRACT_SAMPLE_NOT_APPROVED
SUBCONTRACT_MATERIAL_NOT_TRANSFERRED
FACTORY_CLAIM_SLA_OVERDUE
FINAL_PAYMENT_BLOCKED_BY_QC
```

---

### 7.15. File 17 — Database Schema PostgreSQL v1.1

**Mức ảnh hưởng:** Rất cao  
**Ưu tiên:** P0

#### Tables cần thêm

```text
warehouse_shifts
warehouse_shift_checklist_items
warehouse_shift_variances
packing_tasks
packing_task_items
carrier_manifests
carrier_manifest_items
handover_scan_events
handover_exceptions
return_receipts
return_receipt_items
return_inspections
return_evidences
subcontract_orders
subcontract_order_items
subcontract_material_transfers
subcontract_material_transfer_items
factory_sample_approvals
factory_deliveries
factory_delivery_items
factory_claims
retained_samples
payment_holds
```

#### Index/constraint cần lưu ý

- Unique scan per manifest item.
- Index `carrier_manifest_id, tracking_no`.
- Index `warehouse_shift_id, status`.
- Index `return_receipt_id, status`.
- Index `subcontract_order_id, status`.
- Check constraint batch/expiry required theo item config nếu enforce ở application + DB partial constraints.
- Foreign key từ return/factory receiving vào stock movement.

---

### 7.16. File 18 — DevOps/CI-CD v1.1

**Mức ảnh hưởng:** Thấp-trung bình  
**Ưu tiên:** P1

#### Cần bổ sung smoke test sau deploy

- Create packing task.
- Create manifest and scan one item.
- Receive return and inspect.
- Create subcontract order.
- Close test shift in staging.

#### Migration checklist

- Chạy migration bảng mới.
- Verify enum/status mapping.
- Verify API route permission.
- Verify scan endpoint latency.

---

### 7.17. File 19 — Security/RBAC/Audit v1.1

**Mức ảnh hưởng:** Trung bình-cao  
**Ưu tiên:** P1

#### Sensitive action cần audit bắt buộc

- close shift,
- approve variance,
- confirm handover,
- create handover exception,
- decide return disposition,
- mark return not reusable,
- transfer materials to factory,
- approve factory sample,
- accept/reject factory delivery,
- open factory claim,
- release final payment override.

#### Field-level protection

- Cost/price/payment fields chỉ Finance/CEO.
- Batch/QC fields chỉ QC/Warehouse role liên quan.
- Payment hold/release chỉ Finance + approval.

---

## 8. Tài liệu v1.1 cần xuất bản chính thức

Khuyến nghị xuất bản lại theo batch như sau.

### Batch 1 — Core Business v1.1

```text
03_ERP_PRD_SRS_Phase1_My_Pham_v1_1.md
04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1_1.md
05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1_1.md
06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1_1.md
08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1_1.md
09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1_1.md
```

### Batch 2 — Technical v1.1

```text
11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1_1.md
13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1_1.md
16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1_1.md
17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1_1.md
```

### Batch 3 — Support v1.1

```text
07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1_1.md
10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1_1.md
14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1_1.md
15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1_1.md
18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1_1.md
19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1_1.md
```

---

## 9. Dependency Map

```text
21 Gap Analysis
      ↓
22 Core Docs Revision Log
      ↓
03 PRD/SRS v1.1
      ↓
05 Data Dictionary v1.1
      ↓
06 Process Flow v1.1
      ↓
08 Screen List/Wireframe v1.1
      ↓
16 API Contract v1.1 + 17 DB Schema v1.1
      ↓
09 UAT v1.1 + 25 Product Backlog/Sprint Plan
```

Nguyên tắc:

```text
Không viết API/DB v1.1 trước khi PRD + Data Dictionary + Process Flow đã khóa.
Không viết UAT v1.1 trước khi Screen List và Process Flow đã khóa.
Không build sprint lớn trước khi Backlog/Sprint Plan bám theo tài liệu v1.1.
```

---

## 10. Sprint/Build Impact sơ bộ

### Epic mới hoặc mở rộng

| Epic | Loại | Priority | Ghi chú |
|---|---|---|---|
| Warehouse Daily Board | New | P0 | Dựa trên daily ops thực tế |
| Shift Closing/EOD | New | P0 | Gate vận hành cuối ngày |
| Packing Queue/Verification | New/Expand | P0 | Bám nội quy đóng hàng |
| Carrier Manifest/Handover Scan | New | P0 | Bám quy trình bàn giao ĐVVC |
| Returns Receiving/Inspection | Expand | P0 | Hàng hoàn là luồng riêng |
| Subcontract Manufacturing | New/Refocus | P0 | Thay trọng tâm sản xuất nội bộ sâu |
| Factory Claim SLA | New | P0 | Bám rule phản hồi 3–7 ngày |
| Payment Hold/Release | Expand | P1 | Kết nối Finance và QC acceptance |

### Risk build nếu không revise

| Rủi ro | Hậu quả |
|---|---|
| Không có daily board | Kho vẫn phải điều hành bằng bảng ngoài/Zalo |
| Không có shift closing | Không kiểm soát lệch cuối ngày |
| Không có manifest scan | Bàn giao ĐVVC tiếp tục lệch/thất lạc khó trace |
| Không có returns inspection | Hàng hoàn dễ nhập sai tồn bán được |
| Không có subcontract module | Sản xuất/gia công ngoài tiếp tục rơi vào Excel/giấy |
| Không có payment hold | Có thể thanh toán nhà máy dù hàng chưa đạt |
| Không cập nhật API/DB | Dev build thiếu endpoint/bảng, phải vá sau |

---

## 11. Change Ownership

| Khu vực thay đổi | Business Owner | Product/BA Owner | Technical Owner | Sign-off bắt buộc |
|---|---|---|---|---|
| Warehouse Daily Board | Warehouse Lead/COO | BA | Tech Lead | COO |
| Shift Closing | Warehouse Lead/Finance | BA | Tech Lead | COO + Finance |
| Packing Queue | Warehouse Lead | BA/UI | FE/BE Lead | Warehouse Lead |
| Carrier Manifest | Warehouse Lead/CSKH | BA | BE Lead | COO |
| Returns | CSKH/Warehouse/QC | BA | BE Lead | COO + QA |
| Subcontract Manufacturing | Production/COO | BA | BE Lead | COO |
| Factory Claim | QA/Production | BA | BE Lead | QA Lead + COO |
| Payment Hold | Finance | BA | BE Lead | Finance Lead |
| API/DB v1.1 | Tech Lead | BA support | Tech Lead/DB owner | Tech Lead |
| UAT v1.1 | Business key users | QA | QA Lead | Product Owner |

---

## 12. Definition of Done cho revision v1.1

Một tài liệu được xem là đã revise xong khi đạt đủ:

1. Có mapping với Change Request ID liên quan.
2. Có cập nhật scope/use case/field/status/screen/API/table nếu bị ảnh hưởng.
3. Không mâu thuẫn với các tài liệu cùng batch.
4. Có rõ owner/role/action/approval nếu là nghiệp vụ có rủi ro.
5. Có exception flow, không chỉ main flow.
6. Có UAT hoặc test case tương ứng với P0 change.
7. Có sign-off của owner liên quan.

---

## 13. Danh sách kiểm tra trước khi dev build theo v1.1

Trước khi mở sprint build chính, phải xác nhận:

- [ ] File 03 v1.1 đã chốt scope mới.
- [ ] File 05 v1.1 đã có đủ entity/status/field.
- [ ] File 06 v1.1 đã có đủ process flow P0.
- [ ] File 08 v1.1 đã có đủ màn hình P0.
- [ ] File 16 v1.1 đã có đủ API endpoint P0.
- [ ] File 17 v1.1 đã có đủ table/constraint/index P0.
- [ ] File 09 v1.1 đã có UAT case cho P0.
- [ ] Product Backlog/Sprint Plan đã chia epic/story theo tài liệu v1.1.
- [ ] Permission Matrix đã có role Warehouse Lead, QC, Finance cho action nhạy cảm.
- [ ] Các business owner đã sign-off.

---

## 14. Gợi ý thứ tự làm tiếp sau file 22

Sau file này, thứ tự hợp lý nhất là:

```text
23_ERP_Integration_Spec_Phase1_MyPham_v1.md
24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md
25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md
```

Tuy nhiên, trước khi làm backlog chi tiết, nên ưu tiên tạo bản revision v1.1 cho nhóm tài liệu lõi:

```text
03, 05, 06, 08, 16, 17
```

Vì đây là nhóm trực tiếp quyết định dev build cái gì.

---

## 15. Kết luận

Bản thiết kế ERP Phase 1 hiện tại đã đủ khung, nhưng sau As-Is và Gap Analysis, nó cần được nâng lên **v1.1** để bám đúng thực tế vận hành.

Điểm cốt lõi của revision v1.1 là:

```text
Không xem kho là màn hình nhập/xuất đơn giản.
Không xem giao hàng là một status.
Không xem hàng hoàn là nhập kho thường.
Không xem sản xuất là xưởng nội bộ 100%.
```

ERP v1.1 phải phản ánh đúng:

```text
Kho có ca/ngày/đối soát.
Đóng hàng có kiểm và phân khu.
Bàn giao ĐVVC có manifest và scan.
Hàng hoàn có kiểm tình trạng và phân loại.
Sản xuất có gia công ngoài, duyệt mẫu, chuyển NVL/bao bì và claim nhà máy.
```

Nếu cập nhật đúng các tài liệu lõi theo file 22 này, đội dự án sẽ giảm rất mạnh rủi ro build lệch workflow thật.

