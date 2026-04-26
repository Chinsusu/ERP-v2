# 21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1

**Dự án:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Giai đoạn:** Phase 1  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-24  
**Mục đích:** So sánh workflow thực tế hiện tại (As-Is) với bộ thiết kế ERP Phase 1 (To-Be), chốt điểm nào giữ nguyên, điểm nào chỉnh ERP, điểm nào phải thiết kế lại, và tài liệu nào cần cập nhật lên bản v1.1.

---

## 1. Nguồn dữ liệu dùng để phân tích

Tài liệu này dựa trên 2 nhóm nguồn:

### 1.1. Nhóm tài liệu ERP đã thiết kế

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

### 1.2. Nhóm tài liệu workflow thực tế công ty cung cấp

- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

---

## 2. Kết luận điều hành

Bộ thiết kế ERP hiện tại **đúng hướng lớn**, nhưng sau khi đối chiếu workflow thật, cần chỉnh Phase 1 theo hướng thực tế hơn.

Trọng tâm Phase 1 nên được điều chỉnh từ:

```text
ERP lõi chuẩn: mua hàng → QC → sản xuất → kho → bán hàng → giao hàng
```

thành:

```text
ERP vận hành thực tế: kho hằng ngày → nhập/xuất/đóng hàng → bàn giao ĐVVC → hàng hoàn → gia công ngoài → QC nhận hàng → đối soát cuối ngày
```

Điểm thay đổi lớn nhất là:

1. **Kho không chỉ là nhập/xuất tồn**. Kho đang có nhịp vận hành ngày, kiểm kê cuối ngày, đối soát và kết ca. Vì vậy cần thêm `Warehouse Daily Board` và `Shift Closing / End-of-Day Reconciliation`.
2. **Bàn giao ĐVVC là một quy trình riêng**, có phân khu, thùng/rổ, bảng đối chiếu, quét mã trực tiếp tại hàm/truck và xử lý thiếu đơn. Vì vậy cần thêm `Carrier Manifest`, `Scan Handover`, `Handover Exception Handling`.
3. **Hàng hoàn là một luồng nghiệp vụ riêng**, không thể xử lý như nhập kho thường. Cần có khu hàng hoàn, quét hàng hoàn, kiểm tình trạng, phân loại còn sử dụng/không sử dụng, chuyển kho hoặc chuyển Lab.
4. **Sản xuất hiện tại thiên về gia công ngoài**, không phải xưởng nội bộ 100%. Vì vậy module sản xuất Phase 1 cần ưu tiên `Subcontract Manufacturing / Gia Công Ngoài`, gồm chuyển NVL/bao bì cho nhà máy, duyệt mẫu, nhận hàng, QC và phản hồi lỗi trong 3–7 ngày.
5. **Tài liệu hiện tại cần revision v1.1**, đặc biệt các file `03`, `05`, `06`, `08`, `09`, `16`, `17`.

---

## 3. Nguyên tắc ra quyết định

Mỗi gap được phân loại theo 5 hướng:

| Mã quyết định | Ý nghĩa |
|---|---|
| `KEEP_TO_BE` | Giữ thiết kế ERP hiện tại, không sửa lớn |
| `MODIFY_TO_BE` | Sửa thiết kế ERP để bám workflow thật |
| `ADD_SCOPE` | Bổ sung vào Phase 1 vì workflow thật có nhu cầu rõ |
| `DEFER_PHASE2` | Ghi nhận nhưng chưa làm Phase 1 |
| `REDESIGN` | Thiết kế lại vì cả As-Is và To-Be đều chưa đủ tốt |

Nguyên tắc chọn:

```text
ERP Final Flow = Workflow thực tế + Best Practice ngành + Rule kiểm soát rủi ro
```

Không custom vì thói quen. Chỉ custom khi:

- giảm lỗi vận hành thật,
- giảm thất thoát,
- tăng tốc xử lý đơn,
- bảo vệ batch/QC/hạn dùng,
- giữ lợi thế workflow riêng của công ty.

---

## 4. Bảng Gap Analysis tổng hợp

| Gap ID | Khu vực | As-Is hiện tại | To-Be hiện tại | Gap | Quyết định | Ưu tiên | Tài liệu cần cập nhật |
|---|---|---|---|---|---|---|---|
| G-001 | Kho hằng ngày | Kho nhận đơn trong ngày, xuất/nhập, soạn/đóng, sắp xếp kho, kiểm kê cuối ngày, đối soát, kết ca | Có WMS và stock count nhưng chưa nhấn mạnh nhịp daily ops | Thiếu màn hình điều hành ngày và kết ca | `ADD_SCOPE` | P0 | 03, 06, 08, 09, 16, 17 |
| G-002 | Nhập kho | Nhận chứng từ giao hàng, kiểm số lượng/bao bì/lô, đạt thì nhập/xếp kho, không đạt trả NCC | Có inbound + QC | Cần tách rõ receiving, QC, putaway, reject supplier | `MODIFY_TO_BE` | P0 | 03, 05, 06, 08, 16, 17 |
| G-003 | Chứng từ nhập | Trưởng kho ký xác nhận và lưu phiếu nhập | Có audit/file attachment chung | Thiếu rule lưu chứng từ vật lý + ảnh/file bằng chứng | `MODIFY_TO_BE` | P1 | 05, 08, 16, 17, 19 |
| G-004 | Xuất kho | Làm phiếu xuất, xuất hàng, kiểm số lượng thực tế, ký bàn giao, trưởng kho lưu phiếu | Có outbound/order issue | Cần số hóa phiếu xuất và đối chiếu actual trước bàn giao | `MODIFY_TO_BE` | P0 | 03, 06, 08, 09, 16, 17 |
| G-005 | Đóng hàng | Nhận phiếu đơn hợp lệ, lọc/phân loại theo ĐVVC/đơn/lớp đóng gói, soạn từng đơn | Có pick/pack cơ bản | Thiếu packing queue theo carrier/order/date/package type | `ADD_SCOPE` | P0 | 03, 06, 08, 16, 17 |
| G-006 | Kiểm tra đóng hàng | Đóng gói và kiểm tra tại khu đóng hàng theo SKU, số lượng, tình trạng xung quanh | Có packing status nhưng chưa chi tiết | Cần pack verification và packing defect log | `ADD_SCOPE` | P0 | 03, 05, 06, 08, 09 |
| G-007 | Bàn giao ĐVVC | Phân khu để hàng, để theo thùng/rổ, đối chiếu số lượng, quét mã trực tiếp tại hàm/truck, đủ thì ký bàn giao | Có shipment/giao hàng | Thiếu manifest + scan handover + exception flow | `ADD_SCOPE` | P0 | 03, 06, 08, 09, 16, 17 |
| G-008 | Thiếu đơn khi bàn giao | Nếu chưa đủ thì kiểm tra lại mã; nếu mã chưa có trên hệ thống thì đóng lại; nếu có thì tìm tại khu đóng hàng | Chưa đủ chi tiết | Cần luồng thiếu đơn/thất lạc trước bàn giao | `ADD_SCOPE` | P0 | 03, 06, 08, 09, 16 |
| G-009 | Hàng hoàn | Nhận từ shipper, đưa khu hàng hoàn, quét hàng hoàn, quay/ghi nhận tình trạng, kiểm tra bên trong | Có returns nhưng chưa đủ As-Is | Cần Returns Receiving + Inspection + Evidence | `ADD_SCOPE` | P0 | 03, 05, 06, 08, 09, 16, 17 |
| G-010 | Phân loại hàng hoàn | Còn sử dụng thì chuyển kho; không sử dụng thì chuyển Lab | Có return disposition sơ bộ | Cần trạng thái tồn rõ: returned_pending, reusable, damaged/lab | `MODIFY_TO_BE` | P0 | 05, 16, 17 |
| G-011 | Nhập kho hàng hoàn | Lập phiếu nhập kho có đầy đủ lưu trữ video/chứng từ | Audit/file chung | Cần attachment bắt buộc cho return receipt | `ADD_SCOPE` | P1 | 08, 16, 17, 19 |
| G-012 | Kiểm kê cuối ngày | Kho kiểm kê hàng tồn kho cuối ngày | Có stock count | Cần daily cycle count và variance workflow theo ca | `ADD_SCOPE` | P0 | 03, 06, 07, 08, 09, 17 |
| G-013 | Đối soát cuối ngày | Đối soát số liệu và báo cáo cho quản lý trước khi kết ca | Có báo cáo nhưng chưa là gate vận hành | Cần End-of-Day Reconciliation bắt buộc | `ADD_SCOPE` | P0 | 03, 06, 07, 08, 09 |
| G-014 | Sản xuất/gia công | Lên đơn hàng với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chốt thời gian | Có sản xuất nội bộ/MES | Cần module gia công ngoài là trọng tâm Phase 1 | `ADD_SCOPE` | P0 | 03, 05, 06, 08, 16, 17 |
| G-015 | Chuyển NVL/bao bì | Chuyển kho NVL/bao bì từ công ty qua nhà máy sản xuất | Production issue nội bộ | Cần transfer to subcontractor WIP/location | `MODIFY_TO_BE` | P0 | 03, 05, 06, 16, 17 |
| G-016 | Biên bản bàn giao nhà máy | Ký biên bản bàn giao, kèm COA/MSDS/tem phụ/hóa đơn VAT nếu cần | Có attachments chung | Cần checklist hồ sơ bàn giao NVL/bao bì | `ADD_SCOPE` | P1 | 05, 08, 16, 17 |
| G-017 | Làm mẫu/chốt mẫu | Nhà máy làm mẫu, công ty chốt mẫu trước sản xuất hàng loạt | R&D/sample có trong blueprint nhưng Phase 1 chưa sâu | Cần sample approval gate cho subcontract production | `ADD_SCOPE` | P0 | 03, 05, 06, 08, 16 |
| G-018 | Lưu mẫu | Có bước lưu mẫu | Chưa đủ trong Phase 1 | Cần retained sample record ở production/QC | `ADD_SCOPE` | P1 | 05, 06, 08, 17 |
| G-019 | Nhà máy tự kiểm nghiệm | Nhà máy kiểm nghiệm sản phẩm trước khi giao về | QC bên công ty có nhưng supplier QC chưa rõ | Cần supplier QC report attachment | `ADD_SCOPE` | P1 | 05, 08, 16, 17 |
| G-020 | Nhận hàng gia công | Giao hàng về kho, kiểm tra số lượng/chất lượng | Có inbound/QC | Cần receiving type = subcontract finished goods | `MODIFY_TO_BE` | P0 | 03, 05, 06, 16, 17 |
| G-021 | Hàng không nhận | Nếu không nhận thì báo lại nhà máy về vấn đề trong 3–7 ngày | Chưa có SLA claim với nhà máy | Cần supplier/manufacturer claim SLA | `ADD_SCOPE` | P0 | 03, 06, 07, 08, 09, 16 |
| G-022 | Thanh toán cuối | Thanh toán lần cuối sau khi hàng được nhận | Finance cơ bản có AP/payment | Cần payment hold/release theo QC acceptance | `MODIFY_TO_BE` | P1 | 03, 04, 06, 16 |
| G-023 | Batch/lô | Nhập kho có kiểm lô; sản xuất cần batch nhưng As-Is chưa chuẩn hóa đầy đủ | Data dictionary có batch | Cần bắt buộc batch ở receiving, return, factory receipt | `KEEP_TO_BE + ENFORCE` | P0 | 05, 08, 09, 16, 17 |
| G-024 | Hạn dùng | As-Is chưa thể hiện rõ expiry trong mọi bước | To-Be có expiry | Cần enforce expiry bắt buộc với mỹ phẩm | `KEEP_TO_BE + ENFORCE` | P0 | 05, 08, 09, 16, 17 |
| G-025 | Giấy/Zalo/Excel | Workflow As-Is có bảng/phiếu/ký nhận thủ công | To-Be số hóa | Cần migration/change management và SOP mạnh | `MODIFY_TO_BE` | P1 | 10, 26, 29, 30 |
| G-026 | Vai trò trưởng kho | Trưởng kho ký/lưu phiếu nhập/xuất | Permission Matrix có role chung | Cần role Warehouse Lead với quyền sign-off/close shift | `ADD_SCOPE` | P0 | 04, 08, 16, 19 |
| G-027 | Quét mã | Bàn giao và hàng hoàn cần quét mã | To-Be có scan UX/API nhưng cần rõ hơn | Cần scan-first UX cho handover/returns | `ADD_SCOPE` | P0 | 08, 14, 15, 16 |
| G-028 | Báo cáo quản lý kho | Có báo cáo quản lý cuối ngày | KPI có nhưng chưa bám EOD | Cần EOD Warehouse Report | `ADD_SCOPE` | P0 | 07, 08, 16 |
| G-029 | ĐVVC | Bàn giao theo đơn vị vận chuyển | Shipping module có | Cần carrier master + manifest + handover batch | `ADD_SCOPE` | P0 | 05, 08, 16, 17, 23 |
| G-030 | Nội bộ sản xuất | To-Be có MES nội bộ tương đối rộng | As-Is thiên về gia công ngoài | Nên giảm scope MES nội bộ Phase 1 | `DEFER_PHASE2` | P1 | 03, 06, 08 |

---

## 5. Decision Log chính thức

### D-001 — Bổ sung Warehouse Daily Board vào Phase 1

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0  
**Owner đề xuất:** Product Owner / COO / Warehouse Lead  
**Lý do:** Workflow thực tế có nhịp ngày rõ: nhận đơn, xuất/nhập, soạn/đóng, sắp xếp kho, kiểm kê, đối soát, kết ca.

**Quyết định:**  
Bổ sung màn hình `Warehouse Daily Board` để kho thấy toàn bộ việc trong ngày.

**Chức năng bắt buộc:**

- số đơn cần xử lý trong ngày,
- đơn đang chờ pick,
- đơn đang pack,
- đơn chờ bàn giao ĐVVC,
- hàng hoàn chờ xử lý,
- phiếu nhập chờ QC,
- stock count cuối ngày,
- variance chưa giải quyết,
- ca làm đang mở/đóng.

**Tài liệu bị ảnh hưởng:** `03`, `06`, `07`, `08`, `09`, `16`, `17`.

---

### D-002 — Bắt buộc Shift Closing / End-of-Day Reconciliation

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0  
**Lý do:** As-Is có kiểm kê hàng tồn kho cuối ngày và đối soát số liệu báo quản lý trước khi kết thúc ca.

**Quyết định:**  
Bổ sung quy trình đóng ca kho. Không cho đóng ca nếu còn các mục P0 chưa xử lý.

**Điều kiện đóng ca:**

- tất cả phiếu xuất trong ngày có trạng thái rõ,
- shipment handover đã xác nhận hoặc ghi exception,
- hàng hoàn đã scan và đưa đúng trạng thái,
- stock count/cycle count hoàn tất theo phạm vi cấu hình,
- variance được ghi nhận,
- trưởng kho xác nhận.

**Tài liệu bị ảnh hưởng:** `03`, `06`, `07`, `08`, `09`, `16`, `17`, `19`.

---

### D-003 — Tách rõ Receiving, QC, Putaway, Reject trong nhập kho

**Loại:** `MODIFY_TO_BE`  
**Ưu tiên:** P0

**Quyết định:**  
Nhập kho không được hiểu là “hàng đã khả dụng ngay”. Luồng chuẩn:

```text
Nhận hàng/chứng từ
→ kiểm số lượng/bao bì/lô/hạn dùng
→ QC Hold
→ QC Pass/Fail
→ nếu Pass: Putaway vào kho khả dụng
→ nếu Fail: Reject/Return Supplier hoặc Quarantine
```

**Rule:**  
Hàng chưa QC Pass không được bán, không được xuất cho đơn hàng thường.

**Tài liệu bị ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `16`, `17`.

---

### D-004 — Số hóa phiếu nhập/phiếu xuất nhưng giữ mapping với chứng từ vật lý

**Loại:** `MODIFY_TO_BE`  
**Ưu tiên:** P1

**Quyết định:**  
ERP sẽ có số phiếu điện tử. Nếu công ty vẫn cần giấy ký tay, hệ thống phải cho upload ảnh/file chứng từ và lưu số chứng từ giấy.

**Trường dữ liệu cần thêm:**

- `paper_document_no`
- `signed_by`
- `signed_at`
- `physical_document_location`
- `attachment_required_flag`

**Tài liệu bị ảnh hưởng:** `05`, `08`, `16`, `17`, `19`.

---

### D-005 — Bổ sung Packing Queue theo ĐVVC/đơn/loại đóng gói

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Trước khi đóng hàng, đơn phải vào `Packing Queue`. Cho phép lọc theo:

- đơn vị vận chuyển,
- ngày giao,
- trạng thái đơn,
- loại đóng gói,
- khu vực/kho,
- độ ưu tiên.

**Rule:**  
Chỉ đơn hợp lệ mới vào queue đóng hàng.

**Tài liệu bị ảnh hưởng:** `03`, `06`, `08`, `16`, `17`.

---

### D-006 — Bổ sung Pack Verification

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Đóng hàng phải có bước kiểm trước khi đưa sang khu bàn giao.

**Check bắt buộc:**

- đúng SKU,
- đúng số lượng,
- đúng đơn,
- đúng quà tặng/combo nếu có,
- tình trạng bao bì bên ngoài,
- batch/expiry nếu sản phẩm yêu cầu trace.

**Tài liệu bị ảnh hưởng:** `03`, `06`, `08`, `09`, `16`.

---

### D-007 — Bổ sung Carrier Manifest và Scan Handover

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Bàn giao ĐVVC phải là một nghiệp vụ riêng, không chỉ đổi trạng thái shipment.

**Luồng chuẩn:**

```text
Tạo manifest theo ĐVVC/chuyến
→ gom đơn vào khu bàn giao
→ quét mã từng đơn/vận đơn
→ đối chiếu số lượng theo bảng/manifest
→ nếu đủ: xác nhận handover
→ nếu thiếu: mở exception
→ ký xác nhận/bằng chứng bàn giao
```

**Không cho xác nhận bàn giao nếu:**

- đơn chưa packed,
- đơn không thuộc manifest,
- mã bị trùng,
- trạng thái đơn không hợp lệ,
- thiếu đơn chưa có exception.

**Tài liệu bị ảnh hưởng:** `03`, `06`, `08`, `09`, `16`, `17`, `19`, `23`.

---

### D-008 — Bổ sung Handover Exception Flow

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Khi thiếu đơn lúc bàn giao, hệ thống phải có luồng xử lý:

```text
Thiếu đơn khi scan
→ kiểm mã đơn/vận đơn
→ nếu chưa có trên hệ thống: báo lỗi mapping/order import
→ nếu có trên hệ thống: tìm ở khu đóng hàng/khu bàn giao
→ ghi nhận exception
→ chỉ trưởng kho hoặc role được phép mới resolve
```

**Exception type:**

- `missing_package`
- `wrong_carrier`
- `duplicate_scan`
- `unknown_tracking_code`
- `package_not_packed`
- `package_found_in_packing_area`
- `package_repacked`

**Tài liệu bị ảnh hưởng:** `05`, `06`, `08`, `09`, `16`, `17`, `28`.

---

### D-009 — Bổ sung Returns Receiving + Return Inspection

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Hàng hoàn là một module/subflow riêng.

**Luồng chuẩn:**

```text
Nhận hàng từ shipper
→ đưa vào khu hàng hoàn
→ quét mã đơn/vận đơn/SKU
→ quay/chụp bằng chứng nếu cần
→ kiểm tra tình trạng bên ngoài/bên trong
→ phân loại còn sử dụng hoặc không sử dụng
→ nếu còn sử dụng: chuyển về kho theo trạng thái phù hợp
→ nếu không sử dụng: chuyển Lab/kho hỏng
→ tạo phiếu nhập/biên bản hàng hoàn
```

**Rule:**  
Hàng hoàn không được tự động quay về tồn khả dụng.

**Tài liệu bị ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `16`, `17`, `19`, `28`.

---

### D-010 — Chuẩn hóa trạng thái tồn cho hàng hoàn

**Loại:** `MODIFY_TO_BE`  
**Ưu tiên:** P0

**Quyết định:**  
Bổ sung inventory status:

```text
RETURNED_PENDING_INSPECTION
RETURNED_REUSABLE
RETURNED_DAMAGED
LAB_REVIEW
DISPOSED
AVAILABLE
```

**Rule:**  
Chỉ `RETURNED_REUSABLE` sau khi được role có quyền duyệt mới có thể chuyển sang `AVAILABLE`.

**Tài liệu bị ảnh hưởng:** `05`, `16`, `17`, `19`.

---

### D-011 — Phase 1 ưu tiên Gia Công Ngoài thay vì MES nội bộ đầy đủ

**Loại:** `MODIFY_TO_BE + DEFER_PHASE2`  
**Ưu tiên:** P0

**Quyết định:**  
Trong Phase 1, module sản xuất cần ưu tiên nhánh `Subcontract Manufacturing / Gia Công Ngoài`.

**Giữ trong Phase 1:**

- lệnh gia công,
- đặt nhà máy,
- xác nhận số lượng/quy cách/mẫu mã,
- cọc và milestone thanh toán,
- chuyển NVL/bao bì,
- biên bản bàn giao,
- duyệt mẫu,
- sản xuất hàng loạt ở nhà máy,
- nhận hàng về kho,
- QC nhận hàng,
- claim trong 3–7 ngày,
- thanh toán cuối sau nghiệm thu.

**Defer Phase 2:**

- MES nội bộ chi tiết theo chuyền,
- downtime,
- work center nội bộ,
- công đoạn nội bộ nâng cao.

**Tài liệu bị ảnh hưởng:** `03`, `05`, `06`, `08`, `09`, `16`, `17`.

---

### D-012 — Bổ sung Subcontractor WIP Location

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
NVL/bao bì chuyển cho nhà máy không được mất khỏi tầm kiểm soát. Phải tạo location hoặc inventory bucket riêng:

```text
SUBCONTRACTOR_WIP
```

**Stock movement:**

```text
MATERIAL_TRANSFER_TO_FACTORY
MATERIAL_RETURN_FROM_FACTORY
SUBCONTRACT_FINISHED_GOODS_RECEIPT
SUBCONTRACT_SCRAP_ADJUSTMENT
```

**Tài liệu bị ảnh hưởng:** `05`, `06`, `16`, `17`.

---

### D-013 — Bổ sung Sample Approval Gate cho hàng gia công

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Không cho chuyển trạng thái lệnh gia công sang `MASS_PRODUCTION` nếu chưa có `SAMPLE_APPROVED`.

**Sample status:**

```text
NOT_STARTED
SAMPLE_REQUESTED
SAMPLE_RECEIVED
REVISION_REQUIRED
SAMPLE_APPROVED
SAMPLE_REJECTED
```

**Tài liệu bị ảnh hưởng:** `03`, `05`, `06`, `08`, `16`, `17`.

---

### D-014 — Bổ sung Supplier/Factory Claim SLA 3–7 ngày

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Khi hàng gia công không đạt, hệ thống phải tạo claim với SLA phản hồi trong 3–7 ngày.

**Trường dữ liệu cần thêm:**

- `claim_deadline_at`
- `claim_reason`
- `claim_status`
- `factory_response_at`
- `resolution_type`
- `evidence_attachments`

**Tài liệu bị ảnh hưởng:** `03`, `05`, `06`, `07`, `08`, `09`, `16`, `17`, `28`.

---

### D-015 — Payment Hold cho nhà máy cho tới khi QC Acceptance

**Loại:** `MODIFY_TO_BE`  
**Ưu tiên:** P1

**Quyết định:**  
Thanh toán cuối cho nhà máy chỉ được mở khi hàng được nhận/QC acceptance theo điều kiện cấu hình.

**Rule:**

```text
Nếu receipt_status != ACCEPTED hoặc QC chưa pass thì không cho final payment release.
```

**Tài liệu bị ảnh hưởng:** `03`, `04`, `06`, `16`, `17`, `19`.

---

### D-016 — Enforce Batch/Expiry ở mọi điểm chạm hàng mỹ phẩm

**Loại:** `KEEP_TO_BE + ENFORCE`  
**Ưu tiên:** P0

**Quyết định:**  
Thiết kế cũ đã đúng: mỹ phẩm phải bám batch/hạn dùng. Cần enforce mạnh hơn ở UI/API/DB.

**Điểm bắt buộc:**

- nhập nguyên liệu,
- nhập thành phẩm,
- nhận hàng gia công,
- hàng hoàn,
- QC pass/fail,
- xuất kho,
- trace complaint nếu có.

**Tài liệu bị ảnh hưởng:** `05`, `08`, `09`, `16`, `17`, `19`.

---

### D-017 — Bổ sung role Warehouse Lead Sign-off

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Trưởng kho có quyền sign-off đặc thù:

- xác nhận phiếu nhập/phiếu xuất,
- xác nhận handover manifest,
- đóng ca,
- resolve variance trong ngưỡng,
- mở exception thiếu đơn,
- duyệt chuyển hàng hoàn về kho khả dụng nếu được phân quyền.

**Tài liệu bị ảnh hưởng:** `04`, `08`, `16`, `19`.

---

### D-018 — Bổ sung EOD Warehouse Report

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Báo cáo kho cuối ngày là báo cáo bắt buộc.

**Chỉ số:**

- đơn nhận trong ngày,
- đơn đã pack,
- đơn đã bàn giao,
- đơn thiếu/chưa bàn giao,
- hàng hoàn nhận trong ngày,
- phiếu nhập chờ QC,
- tồn khả dụng,
- tồn hold,
- variance,
- thời gian đóng ca,
- người xác nhận.

**Tài liệu bị ảnh hưởng:** `07`, `08`, `16`.

---

### D-019 — Cần Integration Spec cho ĐVVC và kênh đơn

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P1

**Quyết định:**  
Do bàn giao hàng phụ thuộc đơn vị vận chuyển và mã vận đơn, cần làm file tích hợp riêng.

**Tài liệu tiếp theo:**  
`23_ERP_Integration_Spec_Phase1_MyPham_v1.md`

**Nội dung phải có:**

- nguồn đơn,
- mã vận đơn,
- carrier master,
- shipment status sync,
- COD nếu có,
- scan handover,
- callback/webhook,
- retry/error.

---

### D-020 — Tài liệu lõi cần revision v1.1 sau khi chốt gap

**Loại:** `ADD_SCOPE`  
**Ưu tiên:** P0

**Quyết định:**  
Sau Gap Analysis, cần file Change Log và sau đó update các tài liệu lõi.

**Tài liệu tiếp theo:**  
`22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham.md`

---

## 6. Revised Phase 1 Scope sau Gap Analysis

### 6.1. Vẫn giữ trong Phase 1

```text
1. Master Data
2. Purchasing / Supplier
3. Receiving / QC
4. Warehouse / Inventory / Stock Ledger
5. Sales Order / Pick / Pack
6. Shipping / Carrier Handover
7. Returns / Hàng hoàn
8. Subcontract Manufacturing / Gia công ngoài
9. Basic Finance / AP-AR Payment Gate
10. Basic Dashboard / Report
11. RBAC / Approval / Audit
```

### 6.2. Bổ sung mạnh vào Phase 1

```text
1. Warehouse Daily Board
2. Shift Closing / End-of-Day Reconciliation
3. Carrier Manifest
4. Scan Handover
5. Handover Exception
6. Returns Receiving + Inspection
7. Return Disposition
8. Subcontractor WIP Location
9. Sample Approval Gate
10. Factory Claim SLA 3–7 ngày
11. EOD Warehouse Report
```

### 6.3. Giảm scope hoặc defer Phase 2

```text
1. MES nội bộ chi tiết theo chuyền/công đoạn
2. Production downtime nâng cao
3. Full R&D lab workflow nâng cao
4. Full HRM
5. Full CRM
6. Full KOL/Affiliate attribution
7. Full Finance/accounting posting nâng cao
```

---

## 7. Data Dictionary impact

Các field/status dưới đây cần bổ sung hoặc enforce trong `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`.

### 7.1. Warehouse Daily / Shift Closing

| Field | Ý nghĩa | Ghi chú |
|---|---|---|
| `shift_id` | Mã ca kho | Auto-generated |
| `shift_status` | Trạng thái ca | `OPEN`, `CLOSING`, `CLOSED`, `REOPENED` |
| `opened_by` | Người mở ca | User ID |
| `closed_by` | Người đóng ca | Warehouse Lead |
| `opened_at` | Thời gian mở ca | Timestamp |
| `closed_at` | Thời gian đóng ca | Timestamp |
| `eod_reconciliation_status` | Trạng thái đối soát cuối ngày | `PENDING`, `MATCHED`, `VARIANCE_FOUND`, `APPROVED_WITH_VARIANCE` |
| `variance_count` | Số dòng lệch | Integer |
| `eod_report_id` | Báo cáo cuối ngày | Link report |

### 7.2. Carrier Manifest / Handover

| Field | Ý nghĩa | Ghi chú |
|---|---|---|
| `manifest_id` | Mã bảng kê/chuyến bàn giao | Auto-generated |
| `carrier_id` | Đơn vị vận chuyển | Master data |
| `manifest_status` | Trạng thái manifest | `DRAFT`, `READY_TO_HANDOVER`, `SCANNING`, `HANDED_OVER`, `EXCEPTION`, `CANCELLED` |
| `expected_package_count` | Số kiện dự kiến | Integer |
| `scanned_package_count` | Số kiện đã quét | Integer |
| `handover_exception_count` | Số lỗi bàn giao | Integer |
| `handover_signed_by` | Người ký bàn giao | Warehouse Lead/Carrier |
| `handover_signed_at` | Thời điểm ký | Timestamp |

### 7.3. Handover Exception

| Field | Ý nghĩa |
|---|---|
| `exception_type` | `MISSING_PACKAGE`, `UNKNOWN_TRACKING_CODE`, `DUPLICATE_SCAN`, `WRONG_CARRIER`, `NOT_PACKED`, `FOUND_IN_PACKING_AREA`, `REPACKED` |
| `exception_status` | `OPEN`, `INVESTIGATING`, `RESOLVED`, `CANCELLED` |
| `resolved_by` | Người xử lý |
| `resolution_note` | Ghi chú xử lý |

### 7.4. Returns

| Field | Ý nghĩa |
|---|---|
| `return_receipt_id` | Mã phiếu nhận hàng hoàn |
| `return_source` | `SHIPPER`, `CUSTOMER`, `STORE`, `MARKETPLACE` |
| `return_scan_code` | Mã đơn/vận đơn/SKU được quét |
| `return_condition` | `UNOPENED`, `OPENED`, `DAMAGED_PACKAGING`, `LEAKAGE`, `EXPIRED`, `UNKNOWN` |
| `return_disposition` | `PENDING_INSPECTION`, `REUSABLE`, `LAB_REVIEW`, `DAMAGED`, `DISPOSED` |
| `evidence_required` | Có bắt buộc video/hình ảnh không |
| `evidence_attachment_id` | File bằng chứng |

### 7.5. Subcontract Manufacturing

| Field | Ý nghĩa |
|---|---|
| `subcontract_order_id` | Mã lệnh gia công |
| `factory_id` | Nhà máy gia công |
| `deposit_amount` | Tiền cọc |
| `deposit_status` | `NOT_REQUIRED`, `PENDING`, `PAID` |
| `expected_receiving_date` | Ngày dự kiến nhận hàng |
| `sample_status` | Trạng thái mẫu |
| `subcontract_status` | Trạng thái lệnh gia công |
| `material_transfer_status` | Trạng thái chuyển NVL/bao bì |
| `factory_qc_report_id` | File báo cáo kiểm nghiệm từ nhà máy |
| `factory_claim_deadline_at` | Hạn phản hồi lỗi nhà máy |
| `factory_claim_status` | Trạng thái claim |
| `final_payment_status` | Trạng thái thanh toán cuối |

---

## 8. Process Flow impact

Các flow sau cần cập nhật trong `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`.

### 8.1. Flow mới: Warehouse Daily Operation

```text
Mở ca kho
→ nhận danh sách đơn trong ngày
→ kiểm đơn cần pick/pack
→ xử lý nhập/xuất theo nội quy
→ soạn hàng
→ đóng gói + pack verification
→ gom vào khu bàn giao theo ĐVVC
→ xử lý hàng hoàn nếu có
→ sắp xếp/tối ưu vị trí kho
→ kiểm kê/cycle count cuối ngày
→ đối soát số liệu
→ trưởng kho xác nhận
→ kết thúc ca
```

### 8.2. Flow mới: Carrier Handover

```text
Tạo manifest theo ĐVVC/chuyến
→ phân khu để hàng
→ gom hàng theo thùng/rổ/khu
→ đối chiếu số lượng dự kiến
→ quét từng mã đơn/vận đơn
→ nếu đủ: ký xác nhận bàn giao
→ nếu thiếu: tạo exception và xử lý
→ cập nhật shipment status
→ ghi audit log
```

### 8.3. Flow mới: Return Receiving and Inspection

```text
Nhận hàng từ shipper
→ đưa khu hàng hoàn
→ quét mã hàng hoàn
→ quay/chụp bằng chứng nếu cần
→ kiểm tra bên ngoài/bên trong
→ phân loại còn sử dụng/không sử dụng
→ còn sử dụng: chuyển kho phù hợp
→ không sử dụng: chuyển Lab/kho hỏng
→ lập phiếu nhập/biên bản
→ ghi audit log
```

### 8.4. Flow mới: Subcontract Manufacturing

```text
Tạo đơn/lệnh gia công với nhà máy
→ xác nhận số lượng/quy cách/mẫu mã
→ cọc đơn và chốt thời gian
→ chuyển NVL/bao bì cho nhà máy
→ ký biên bản bàn giao kèm hồ sơ
→ nhà máy làm mẫu
→ công ty duyệt mẫu
→ nhà máy sản xuất hàng loạt
→ nhà máy kiểm nghiệm nội bộ
→ giao hàng về kho
→ kiểm tra số lượng/chất lượng
→ nếu đạt: nhập kho/QC pass theo quy định
→ nếu không đạt: tạo claim trong 3–7 ngày
→ thanh toán cuối sau nghiệm thu
```

---

## 9. Screen List impact

Các màn hình cần bổ sung hoặc chỉnh trong `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`.

### 9.1. Warehouse

| Screen ID | Tên màn hình | Hành động |
|---|---|---|
| `WH-DAILY-01` | Warehouse Daily Board | Add |
| `WH-SHIFT-01` | Open/Close Warehouse Shift | Add |
| `WH-EOD-01` | End-of-Day Reconciliation | Add |
| `WH-VAR-01` | Stock Variance Review | Add |
| `WH-RET-01` | Return Receiving | Add |
| `WH-RET-02` | Return Inspection & Disposition | Add |

### 9.2. Shipping

| Screen ID | Tên màn hình | Hành động |
|---|---|---|
| `SHP-MAN-01` | Carrier Manifest List | Add |
| `SHP-MAN-02` | Manifest Detail | Add |
| `SHP-SCAN-01` | Scan Handover | Add |
| `SHP-EXC-01` | Handover Exception | Add |

### 9.3. Subcontract Manufacturing

| Screen ID | Tên màn hình | Hành động |
|---|---|---|
| `MFG-SUB-01` | Subcontract Order List | Add |
| `MFG-SUB-02` | Subcontract Order Detail | Add |
| `MFG-MAT-01` | Material/Packaging Transfer to Factory | Add |
| `MFG-SMP-01` | Sample Approval | Add |
| `MFG-RCV-01` | Factory Finished Goods Receiving | Add |
| `MFG-CLM-01` | Factory Claim / Defect Report | Add |

---

## 10. API impact

Các endpoint cần bổ sung hoặc chỉnh trong `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`.

### 10.1. Warehouse Shift / EOD

```http
POST   /api/v1/warehouse/shifts
GET    /api/v1/warehouse/shifts/{id}
POST   /api/v1/warehouse/shifts/{id}/close
POST   /api/v1/warehouse/shifts/{id}/reopen
GET    /api/v1/warehouse/shifts/{id}/reconciliation
POST   /api/v1/warehouse/shifts/{id}/reconciliation/approve
```

### 10.2. Carrier Manifest / Handover

```http
POST   /api/v1/shipping/manifests
GET    /api/v1/shipping/manifests/{id}
POST   /api/v1/shipping/manifests/{id}/ready
POST   /api/v1/shipping/manifests/{id}/scan
POST   /api/v1/shipping/manifests/{id}/handover
POST   /api/v1/shipping/manifests/{id}/exceptions
POST   /api/v1/shipping/handover-exceptions/{id}/resolve
```

### 10.3. Returns

```http
POST   /api/v1/returns/receipts
POST   /api/v1/returns/receipts/{id}/scan
POST   /api/v1/returns/receipts/{id}/inspect
POST   /api/v1/returns/receipts/{id}/disposition
POST   /api/v1/returns/receipts/{id}/complete
```

### 10.4. Subcontract Manufacturing

```http
POST   /api/v1/manufacturing/subcontract-orders
GET    /api/v1/manufacturing/subcontract-orders/{id}
POST   /api/v1/manufacturing/subcontract-orders/{id}/submit
POST   /api/v1/manufacturing/subcontract-orders/{id}/approve
POST   /api/v1/manufacturing/subcontract-orders/{id}/material-transfer
POST   /api/v1/manufacturing/subcontract-orders/{id}/sample-request
POST   /api/v1/manufacturing/subcontract-orders/{id}/sample-approve
POST   /api/v1/manufacturing/subcontract-orders/{id}/receive-finished-goods
POST   /api/v1/manufacturing/subcontract-orders/{id}/claims
POST   /api/v1/manufacturing/subcontract-claims/{id}/resolve
```

---

## 11. Database impact

Các bảng hoặc cụm bảng cần bổ sung/chỉnh trong `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`.

### 11.1. Warehouse shift / EOD

```text
warehouse_shifts
warehouse_shift_tasks
warehouse_eod_reconciliations
warehouse_eod_reconciliation_lines
stock_variances
stock_variance_resolutions
```

### 11.2. Shipping manifest / handover

```text
carrier_manifests
carrier_manifest_lines
handover_scan_events
handover_exceptions
handover_signatures
```

### 11.3. Returns

```text
return_receipts
return_receipt_lines
return_inspections
return_dispositions
return_evidence_attachments
```

### 11.4. Subcontract manufacturing

```text
subcontract_orders
subcontract_order_lines
subcontract_material_transfers
subcontract_material_transfer_lines
subcontract_sample_approvals
subcontract_finished_goods_receipts
subcontract_factory_qc_reports
subcontract_claims
subcontract_claim_attachments
```

### 11.5. Stock movements cần bổ sung

```text
RETURN_RECEIPT
RETURN_TO_AVAILABLE
RETURN_TO_LAB
RETURN_DISPOSAL
MATERIAL_TRANSFER_TO_FACTORY
MATERIAL_RETURN_FROM_FACTORY
SUBCONTRACT_FINISHED_GOODS_RECEIPT
SUBCONTRACT_REJECTED_RECEIPT
SUBCONTRACT_SCRAP_ADJUSTMENT
```

---

## 12. UAT impact

Các test case cần bổ sung trong `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`.

### 12.1. Warehouse Daily / EOD

- Mở ca kho thành công.
- Nhận danh sách đơn trong ngày.
- Có đơn chưa bàn giao thì hệ thống cảnh báo khi đóng ca.
- Có hàng hoàn chưa xử lý thì hệ thống cảnh báo khi đóng ca.
- Có variance thì không đóng ca hoặc đóng ca với approval theo rule.
- Trưởng kho xem được EOD report.

### 12.2. Carrier Handover

- Tạo manifest theo ĐVVC.
- Quét đủ đơn thì cho xác nhận bàn giao.
- Quét mã không thuộc manifest thì báo lỗi.
- Quét trùng mã thì báo lỗi.
- Thiếu đơn thì tạo exception.
- Resolve exception xong mới cho complete.

### 12.3. Returns

- Nhận hàng hoàn từ shipper.
- Quét mã đơn/vận đơn.
- Ghi nhận tình trạng và attachment.
- Phân loại còn sử dụng → chuyển kho.
- Phân loại không sử dụng → chuyển Lab.
- Hàng hoàn chưa inspection không xuất bán được.

### 12.4. Subcontract Manufacturing

- Tạo lệnh gia công.
- Cọc đơn và chốt thời gian.
- Chuyển NVL/bao bì sang nhà máy.
- Không cho mass production nếu chưa approve sample.
- Nhận hàng về kho.
- QC pass thì nhập kho.
- QC fail thì tạo factory claim.
- Claim có deadline 3–7 ngày.
- Không cho final payment nếu chưa acceptance.

---

## 13. Permission/RBAC impact

Các quyền cần bổ sung trong `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md` và `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`.

| Permission Code | Ý nghĩa | Role đề xuất |
|---|---|---|
| `warehouse.shift.open` | Mở ca kho | Warehouse Staff, Warehouse Lead |
| `warehouse.shift.close` | Đóng ca kho | Warehouse Lead |
| `warehouse.eod.approve` | Duyệt đối soát cuối ngày | Warehouse Lead, COO |
| `shipping.manifest.create` | Tạo manifest | Warehouse Staff, Shipping Staff |
| `shipping.manifest.handover` | Xác nhận bàn giao | Warehouse Lead |
| `shipping.exception.resolve` | Xử lý thiếu/lỗi bàn giao | Warehouse Lead, COO |
| `returns.receive` | Nhận hàng hoàn | Warehouse Staff |
| `returns.inspect` | Kiểm hàng hoàn | Warehouse Lead / QA tùy rule |
| `returns.disposition.approve` | Duyệt phân loại hàng hoàn | Warehouse Lead / QA |
| `subcontract.order.create` | Tạo lệnh gia công | Production/Purchasing |
| `subcontract.order.approve` | Duyệt lệnh gia công | COO / CEO theo ngưỡng |
| `subcontract.material.transfer` | Chuyển NVL/bao bì cho nhà máy | Warehouse Lead |
| `subcontract.sample.approve` | Duyệt mẫu | R&D / QA / Brand |
| `subcontract.claim.create` | Tạo claim nhà máy | QA / Warehouse Lead |
| `subcontract.final_payment.release` | Mở thanh toán cuối | Finance + COO |

---

## 14. KPI/Report impact

Các báo cáo cần bổ sung trong `07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1.md`.

| Report/KPI | Ý nghĩa | Owner |
|---|---|---|
| Warehouse EOD Report | Báo cáo kho cuối ngày | Warehouse Lead |
| Orders Processed Today | Số đơn xử lý trong ngày | COO/Warehouse |
| Packed vs Handed Over | Đơn đã đóng so với đã bàn giao | Warehouse/Shipping |
| Handover Exception Rate | Tỷ lệ lỗi khi bàn giao | COO |
| Missing Package Count | Số đơn thiếu lúc bàn giao | Warehouse Lead |
| Return Receiving Count | Số hàng hoàn nhận trong ngày | Warehouse/CSKH |
| Return Reusable Rate | Tỷ lệ hàng hoàn còn dùng được | COO/Finance |
| Return Lab/Damaged Count | Hàng hoàn chuyển Lab/hỏng | QA/Finance |
| Subcontract Order Status | Trạng thái lệnh gia công | COO |
| Factory Claim SLA | Claim nhà máy đúng hạn hay quá hạn | QA/COO |
| Material at Factory Value | Giá trị NVL/bao bì đang ở nhà máy | Finance/COO |
| Shift Closing On-Time Rate | Tỷ lệ đóng ca đúng giờ | Warehouse Lead |

---

## 15. Risk impact

Các rủi ro được giảm sau khi áp dụng gap decision:

| Rủi ro | Nếu không sửa | Control mới |
|---|---|---|
| Lệch tồn cuối ngày | Tồn hệ thống không khớp thực tế | EOD reconciliation + variance |
| Thất lạc đơn khi bàn giao | ĐVVC nhận thiếu hoặc kho tưởng đã giao | Manifest + scan handover |
| Hàng hoàn quay về bán sai | Hàng hỏng/lỗi vào tồn khả dụng | Return inspection + disposition |
| Mất kiểm soát NVL gửi nhà máy | NVL/bao bì ra khỏi kho nhưng không trace được | Subcontractor WIP location |
| Thanh toán nhà máy khi hàng chưa đạt | Trả tiền xong mới phát hiện lỗi | Payment hold until acceptance |
| Không kịp claim nhà máy | Quá hạn 3–7 ngày, mất quyền xử lý | Factory claim SLA alert |
| Claim/QC thiếu bằng chứng | Tranh cãi với nhà máy/NCC | Attachment checklist |
| User sửa dữ liệu nhạy cảm | Số liệu bị bóp méo | RBAC + audit log |

---

## 16. Change Request Backlog sinh ra từ Gap Analysis

| CR ID | Tên change request | Ưu tiên | Module | Trạng thái |
|---|---|---|---|---|
| CR-001 | Add Warehouse Daily Board | P0 | Warehouse | Proposed |
| CR-002 | Add Shift Closing / EOD Reconciliation | P0 | Warehouse | Proposed |
| CR-003 | Add Pack Verification | P0 | Warehouse/Sales | Proposed |
| CR-004 | Add Carrier Manifest | P0 | Shipping | Proposed |
| CR-005 | Add Scan Handover | P0 | Shipping | Proposed |
| CR-006 | Add Handover Exception Flow | P0 | Shipping | Proposed |
| CR-007 | Add Return Receiving | P0 | Returns | Proposed |
| CR-008 | Add Return Inspection & Disposition | P0 | Returns/QC | Proposed |
| CR-009 | Add Subcontract Manufacturing Order | P0 | Manufacturing | Proposed |
| CR-010 | Add Material Transfer to Factory | P0 | Inventory/Manufacturing | Proposed |
| CR-011 | Add Sample Approval Gate | P0 | Manufacturing/R&D/QC | Proposed |
| CR-012 | Add Factory Finished Goods Receiving | P0 | Manufacturing/Warehouse/QC | Proposed |
| CR-013 | Add Factory Claim SLA 3–7 Days | P0 | Manufacturing/QC | Proposed |
| CR-014 | Add Payment Hold for Final Factory Payment | P1 | Finance/Manufacturing | Proposed |
| CR-015 | Add Digital Document Mapping for Signed Paper | P1 | All Ops | Proposed |
| CR-016 | Add Warehouse Lead Sign-off Permissions | P0 | RBAC | Proposed |
| CR-017 | Add EOD Warehouse Report | P0 | Reporting | Proposed |
| CR-018 | Add Return Evidence Attachment Requirement | P1 | Returns/Security | Proposed |
| CR-019 | Add Subcontractor WIP Stock Location | P0 | Inventory | Proposed |
| CR-020 | Defer Full Internal MES to Phase 2 | P1 | Manufacturing | Proposed |

---

## 17. Tài liệu cần cập nhật lên v1.1

| File | Mức ảnh hưởng | Nội dung cần sửa |
|---|---|---|
| `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md` | Rất cao | Scope Phase 1, warehouse daily, shipping handover, returns, subcontract manufacturing |
| `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md` | Cao | Role Warehouse Lead, shipping exception, return disposition, subcontract approvals |
| `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md` | Rất cao | Status/fields cho shift, manifest, returns, subcontract manufacturing |
| `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md` | Rất cao | 4 flow mới: daily warehouse, handover, returns, subcontract manufacturing |
| `07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1.md` | Trung bình cao | EOD warehouse report, handover KPI, return KPI, factory claim SLA |
| `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md` | Rất cao | Screens mới cho warehouse daily, manifest, scan, returns, subcontract |
| `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md` | Rất cao | Test cases mới cho handover, return, EOD, subcontract |
| `10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1.md` | Trung bình | Opening stock, returns pending, factory WIP, documents |
| `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md` | Trung bình | Module/component/service boundary update |
| `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md` | Trung bình | Module contracts cho manifest/returns/subcontract |
| `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md` | Trung bình | Scan UX, shift close UX, return UX |
| `15_ERP_Frontend_Architecture_React_NextJS_Phase1_MyPham_v1.md` | Trung bình | Routes/components mới |
| `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md` | Rất cao | API mới |
| `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md` | Rất cao | Tables/movements/status mới |
| `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md` | Cao | Sensitive actions and field-level controls |

---

## 18. Open Questions cần hỏi chủ doanh nghiệp/team vận hành

Các câu hỏi này cần trả lời trước khi chốt v1.1.

### 18.1. Kho và đơn hàng

1. Đơn hàng trong ngày đang lấy từ nguồn nào: website, sàn, POS, Excel, phần mềm khác?
2. Một ngày trung bình bao nhiêu đơn?
3. Đơn được gom theo ĐVVC trước hay sau khi soạn hàng?
4. Có split order không, tức một đơn chia nhiều kiện?
5. Có bán combo/quà tặng không? Nếu có, pack verification cần kiểm combo/gift.
6. Kho có dùng mã vạch SKU hiện tại chưa?
7. Batch/hạn dùng có đang dán trên từng sản phẩm/thùng không?
8. Có nhiều kho/khu vực/kệ/bin không?
9. Kiểm kê cuối ngày là kiểm toàn bộ hay kiểm theo nhóm hàng?
10. Chênh lệch tồn hiện tại ai được quyền xử lý?

### 18.2. ĐVVC và bàn giao

1. Hiện có bao nhiêu ĐVVC?
2. Mã quét lúc bàn giao là mã đơn, mã vận đơn, barcode nội bộ hay mã sàn?
3. Có in manifest/bảng kê giấy không?
4. Carrier có ký trên giấy hay ký điện tử?
5. Khi thiếu đơn, quy định thời gian tìm là bao lâu?
6. Nếu mã chưa có trên hệ thống, ai chịu trách nhiệm: kho, sales admin, marketplace ops?
7. Có COD không? Nếu có, đối soát COD sẽ cần file tích hợp riêng.

### 18.3. Hàng hoàn

1. Hàng hoàn nhận từ ĐVVC hay khách trực tiếp?
2. Hàng hoàn có quay video bắt buộc không?
3. Tiêu chí “còn sử dụng” là ai quyết định: kho, QA hay CSKH?
4. Hàng hoàn có thể bán lại không? Nếu có thì cần rule QC/QA rất rõ.
5. Hàng không sử dụng chuyển Lab để làm gì: test, tiêu hủy, đối chứng?
6. Có cần xuất biên bản tiêu hủy không?

### 18.4. Gia công ngoài

1. Công ty sở hữu nguyên liệu/bao bì rồi gửi nhà máy, hay nhà máy tự mua một phần?
2. NVL/bao bì gửi nhà máy có được kiểm tồn định kỳ không?
3. Nhà máy có gửi báo cáo tiêu hao không?
4. Có quản hao hụt NVL/bao bì tại nhà máy không?
5. Có nhiều nhà máy không?
6. Công ty có cần trace batch nguyên liệu về batch thành phẩm gia công không?
7. Duyệt mẫu do ai ký: R&D, QA, brand, CEO?
8. Lỗi hàng trong 3–7 ngày tính từ lúc nhận hàng hay từ lúc kiểm xong?
9. Thanh toán cuối theo tỷ lệ nào và ai duyệt?

### 18.5. Tài chính/kế toán

1. Cọc nhà máy được ghi nhận thế nào?
2. Thanh toán cuối có cần đối chiếu invoice/VAT không?
3. Hàng hoàn làm giảm doanh thu hay chỉ tạo phiếu nhập?
4. Hàng hỏng/Lab có hạch toán giá trị riêng không?
5. Chi phí ĐVVC có đối soát trong Phase 1 không?

---

## 19. Roadmap revision sau tài liệu này

Sau khi tài liệu này được duyệt, nên làm tiếp:

```text
22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham.md
23_ERP_Integration_Spec_Phase1_MyPham_v1.md
24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md
25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md
```

Trong đó, file `22` là bước nối trực tiếp để sửa lại bộ lõi, tránh tài liệu nói lệch nhau.

---

## 20. Sign-off đề xuất

| Vai trò | Người duyệt | Trạng thái | Ghi chú |
|---|---|---|---|
| CEO / Owner |  | Pending | Chốt scope và ưu tiên |
| COO / Operations |  | Pending | Chốt kho, bàn giao, returns, gia công |
| Warehouse Lead |  | Pending | Chốt flow kho và đóng ca |
| QA/QC Lead |  | Pending | Chốt QC, returns, factory claim |
| Finance Lead |  | Pending | Chốt cọc, payment hold, hàng hoàn |
| Tech Lead |  | Pending | Chốt API/DB/module impact |
| Product/BA |  | Pending | Chốt update tài liệu v1.1 |

---

## 21. Kết luận

Sau khi đối chiếu As-Is với To-Be, kết luận là:

```text
Bộ ERP hiện tại đúng nền,
nhưng Phase 1 cần xoay trọng tâm mạnh hơn vào warehouse daily operation,
carrier handover, returns và subcontract manufacturing.
```

Đây không phải thay đổi nhỏ. Đây là thay đổi giúp ERP bám đúng cách công ty đang vận hành thật.

Nếu không cập nhật, hệ thống vẫn có thể “đúng ERP”, nhưng dễ lệch các điểm gây tiền thật mất thật:

- lệch tồn cuối ngày,
- thiếu đơn khi bàn giao ĐVVC,
- hàng hoàn quay lại kho sai,
- NVL/bao bì gửi nhà máy không trace được,
- hàng gia công lỗi nhưng claim trễ,
- thanh toán nhà máy trước khi nghiệm thu.

Vì vậy quyết định quan trọng nhất sau tài liệu này là:

```text
Update bộ tài liệu lõi lên v1.1 trước khi bước vào build sâu.
```

