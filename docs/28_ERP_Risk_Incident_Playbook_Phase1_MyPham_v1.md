# 28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1

**Dự án:** Web ERP Phase 1 cho công ty mỹ phẩm  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-24  
**Tài liệu:** Risk & Incident Playbook  
**Phạm vi:** Phase 1 — Master Data, Purchasing, QC, Inventory/WMS, Sales Order, Shipping/Handover, Returns, Subcontract Manufacturing, Finance basic controls, Security/Audit  
**Tech stack tham chiếu:** Go Backend, PostgreSQL, React/Next.js Frontend, REST/OpenAPI, Redis/Queue, S3/MinIO, Docker/CI-CD  

---

## 1. Mục tiêu tài liệu

Tài liệu này là **sổ tay xử lý rủi ro và sự cố** cho ERP Phase 1.

ERP không được thiết kế với giả định “sẽ không bao giờ lỗi”. Hệ thống phải được thiết kế để khi lỗi xảy ra thì đội vận hành biết:

- ai là người chịu trách nhiệm;
- cần dừng gì trước;
- cần khóa dữ liệu nào;
- cần kiểm tra chứng từ nào;
- cần phục hồi vận hành ra sao;
- cần ghi nhận audit/evidence thế nào;
- cần làm CAPA để lỗi không lặp lại.

Mục tiêu cuối cùng:

```text
Sự cố không lan rộng
Dữ liệu không bị sửa bừa
Tồn kho không bị méo
Batch/QC không bị phá kiểm soát
Đơn hàng/giao hàng không rối
Người chịu trách nhiệm rõ ràng
Có bằng chứng và postmortem sau sự cố
```

---

## 2. Nguồn đầu vào

Tài liệu này dựa trên:

- `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
- `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`
- `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`
- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md`
- `13_ERP_Go_Module_Component_Design_Standards_Phase1_MyPham_v1.md`
- `16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md`
- `17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md`
- `18_ERP_DevOps_CICD_Environment_Standards_Phase1_MyPham_v1.md`
- `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`
- `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md`
- `21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md`
- `22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham.md`
- `23_ERP_Integration_Spec_Phase1_MyPham_v1.md`
- `24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md`
- `25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md`
- `26_ERP_SOP_Training_Manual_Phase1_MyPham_v1.md`
- `27_ERP_GoLive_Runbook_Hypercare_Phase1_MyPham_v1.md`

Và 4 workflow thực tế:

- `Công-việc-hằng-ngày.pdf`
- `Nội-Quy.pdf`
- `Quy-trình-bàn-giao.pdf`
- `Quy-trình-sản-xuất.pdf`

Các điểm workflow thực tế cần ưu tiên trong playbook:

- kho tiếp nhận đơn trong ngày;
- thực hiện xuất/nhập theo nội quy;
- soạn hàng và đóng gói;
- sắp xếp/tối ưu vị trí kho;
- kiểm kê tồn kho cuối ngày;
- đối soát số liệu và báo cáo quản lý;
- kết thúc ca;
- nhập kho có kiểm số lượng, bao bì, lô;
- xuất kho có phiếu xuất, đối chiếu số lượng thực tế, ký bàn giao;
- đóng hàng theo đơn/kênh/ĐVVC;
- bàn giao ĐVVC bằng phân khu, đối chiếu số lượng và quét mã;
- xử lý thiếu đơn khi bàn giao;
- nhận hàng hoàn từ shipper, đưa vào khu hàng hoàn, quét hàng hoàn, kiểm tra tình trạng, phân loại còn dùng/không dùng;
- gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, duyệt mẫu, sản xuất hàng loạt, nhận hàng, kiểm tra chất lượng/số lượng, phản hồi lỗi nhà máy trong 3–7 ngày.

---

## 3. Nguyên tắc xử lý sự cố

### 3.1. Dừng lan trước, sửa sau

Khi phát hiện sự cố liên quan đến tồn kho, batch, QC, giao hàng hoặc hàng hoàn:

```text
Dừng phần đang lan lỗi
Khóa trạng thái liên quan
Khoanh vùng dữ liệu
Ghi nhận bằng chứng
Rồi mới xử lý nguyên nhân
```

Không được vội sửa dữ liệu để “cho chạy tiếp” nếu chưa hiểu tác động.

### 3.2. Không sửa thẳng database

Các hành động bị cấm:

- update trực tiếp tồn kho trong database;
- xóa stock movement;
- sửa QC status trực tiếp không qua workflow;
- xóa đơn hàng đã phát sinh giao dịch;
- sửa batch/expiry không audit;
- sửa tiền/công nợ không chứng từ điều chỉnh;
- xóa audit log;
- xóa scan event hoặc manifest event.

Nếu cần sửa dữ liệu, phải đi qua:

```text
Adjustment transaction
Reverse transaction
Correction note
Approval
Audit log
```

### 3.3. Một sự cố phải có một owner

Mỗi incident phải có **Incident Owner**. Owner có quyền điều phối, nhưng không nhất thiết là người trực tiếp sửa lỗi.

Ví dụ:

| Loại sự cố | Incident Owner chính |
|---|---|
| Tồn kho lệch | Warehouse Manager |
| Batch/QC lỗi | QA Manager |
| Đơn không bàn giao được | Shipping Lead |
| Hàng hoàn sai trạng thái | Warehouse Manager + CSKH Lead |
| Gia công ngoài giao lỗi | Purchasing/Subcontract Owner + QA |
| Lỗi phân quyền/bảo mật | IT/Security Owner |
| Lỗi hệ thống/API | Tech Lead |
| Lỗi giá/công nợ | Finance Owner |

### 3.4. Sự cố lớn phải có war room

Các incident P0/P1 phải mở war room ngay.

War room có thể là:

- nhóm nội bộ riêng;
- Google Meet/Zoom;
- kênh chat khẩn;
- board incident trong ERP hoặc ticket system.

War room phải có log ngắn gọn:

```text
Thời điểm phát hiện
Triệu chứng
Phạm vi ảnh hưởng
Hành động đã làm
Ai đang phụ trách
ETA tiếp theo
Quyết định Go/No-Go nếu liên quan vận hành
```

### 3.5. Sau sự cố phải có CAPA

Sự cố nghiêm trọng không kết thúc ở “đã sửa”. Phải có:

- root cause;
- corrective action;
- preventive action;
- owner;
- deadline;
- bằng chứng đã đóng CAPA.

---

## 4. Ma trận mức độ nghiêm trọng

| Mức | Tên | Định nghĩa | Ví dụ | SLA phản hồi | SLA workaround |
|---|---|---|---|---|---|
| P0 | Critical | Dừng vận hành chính hoặc rủi ro lớn về tiền/hàng/QC/bảo mật | Không xuất được toàn bộ đơn; stock ledger sai hàng loạt; lộ dữ liệu nhạy cảm | 15 phút | 1 giờ |
| P1 | High | Ảnh hưởng nghiêm trọng một quy trình chính nhưng còn workaround | Không bàn giao được một ĐVVC; batch bị pass nhầm; hàng hoàn nhập sai nhiều | 30 phút | 4 giờ |
| P2 | Medium | Ảnh hưởng một nhóm user hoặc một phần nghiệp vụ | Một màn hình report sai; một số đơn scan lỗi | 4 giờ làm việc | 1 ngày |
| P3 | Low | Lỗi nhỏ, không ảnh hưởng vận hành tức thời | Sai label UI, thiếu filter, export chậm | 1 ngày làm việc | Theo sprint |

### 4.1. Điều kiện tự động nâng severity

Incident phải nâng lên P0/P1 nếu có một trong các dấu hiệu:

- sai tồn kho ảnh hưởng nhiều SKU/batch;
- batch fail vẫn bán được;
- đơn đã bàn giao nhưng hệ thống chưa ghi nhận;
- mất audit log;
- user thường xem được dữ liệu giá vốn/lương/payout;
- dữ liệu bị xóa hoặc nghi ngờ bị xóa;
- không thể tạo/nhận đơn trong giờ cao điểm;
- factory delivery có lỗi chất lượng nhưng hệ thống vẫn cho nhập available;
- hàng hoàn không xác định được tình trạng nhưng đã đưa lại vào kho bán.

---

## 5. Vai trò trong xử lý sự cố

| Vai trò | Trách nhiệm |
|---|---|
| Incident Owner | Điều phối sự cố, ra quyết định vận hành, cập nhật tình trạng |
| Business Owner | Chốt impact nghiệp vụ, duyệt workaround đặc biệt |
| Tech Lead | Chẩn đoán lỗi kỹ thuật, điều phối dev fix |
| QA Lead | Xác nhận lỗi, test fix, kiểm tra regression |
| Warehouse Manager | Kiểm soát tồn kho, scan, hàng xuất/nhập, đóng ca |
| QA/QC Manager | Khóa/release batch, xử lý lỗi chất lượng, CAPA |
| Shipping Lead | Kiểm soát manifest, bàn giao ĐVVC, scan, thiếu đơn |
| CSKH Lead | Xử lý khách bị ảnh hưởng, đổi trả/khiếu nại |
| Finance Owner | Kiểm tra công nợ, COD, giá vốn, bút toán/đối soát |
| Security/Admin | Khóa tài khoản, kiểm tra quyền, truy vết audit |
| CEO/COO | Duyệt quyết định lớn: dừng bán, rollback, thu hồi, xử lý nhà máy/NCC |

---

## 6. Vòng đời xử lý incident

```text
1. Detect
2. Triage
3. Contain
4. Diagnose
5. Resolve / Workaround
6. Validate
7. Communicate
8. Close
9. Postmortem / CAPA
```

### 6.1. Detect — phát hiện

Nguồn phát hiện:

- user báo lỗi;
- dashboard cảnh báo;
- daily reconciliation;
- scan mismatch;
- QC exception;
- integration failure;
- monitoring alert;
- security alert;
- finance reconciliation;
- khách hàng/ĐVVC/nhà máy phản hồi.

### 6.2. Triage — phân loại

Câu hỏi triage nhanh:

```text
Lỗi ảnh hưởng tiền, hàng, batch/QC, giao hàng hay bảo mật?
Có đang lan sang đơn/batch/SKU khác không?
Có cần dừng thao tác liên quan không?
Có workaround an toàn không?
Có cần báo CEO/COO ngay không?
```

### 6.3. Contain — khoanh vùng

Ví dụ containment:

- set batch = HOLD;
- lock warehouse location;
- pause carrier manifest;
- block order allocation cho SKU/batch lỗi;
- pause integration sync;
- disable user/session;
- freeze returns disposition;
- tạm dừng factory receiving.

### 6.4. Diagnose — chẩn đoán

Cần kiểm tra:

- audit log;
- stock ledger;
- scan event;
- API log;
- outbox/event log;
- integration log;
- database transaction log nếu cần;
- chứng từ gốc;
- ảnh/chứng cứ kho;
- user action timeline.

### 6.5. Resolve — xử lý

Có 3 loại xử lý:

```text
Workaround vận hành
Hotfix kỹ thuật
Correction nghiệp vụ có phê duyệt
```

### 6.6. Validate — xác nhận

Không đóng incident nếu chưa xác nhận:

- dữ liệu đúng;
- tồn đúng;
- trạng thái đúng;
- user liên quan thao tác lại được;
- không ảnh hưởng luồng khác;
- audit/evidence đầy đủ.

### 6.7. Close + CAPA

Incident closure phải có:

- nguyên nhân gốc;
- scope impact;
- hành động xử lý;
- người duyệt;
- dữ liệu/chứng từ đã điều chỉnh;
- CAPA;
- deadline follow-up.

---

## 7. Incident Playbook chi tiết theo nhóm rủi ro

---

## 7.1. Tồn kho lệch sau kiểm kê cuối ngày

### Trigger

- Kiểm kê cuối ngày không khớp ERP.
- Stock physical khác stock system.
- Kho báo thiếu/thừa khi đóng ca.
- Sales thấy có hàng nhưng kho không tìm thấy.

### Impact

- Sale bán nhầm hàng không có.
- Warehouse pick/pack sai.
- Finance tính sai tồn tiền.
- Batch/hạn dùng có thể lệch.

### Severity

- P1 nếu lệch ảnh hưởng nhiều SKU/batch hoặc không thể xuất đơn.
- P2 nếu lệch ít SKU và có thể khoanh vùng.

### Immediate actions

```text
1. Warehouse Manager tạo incident.
2. Tạm khóa xuất SKU/batch/location nghi ngờ nếu lệch nghiêm trọng.
3. Export stock ledger liên quan.
4. Kiểm lại physical stock tại bin/location.
5. Đối chiếu phiếu nhập/xuất/chuyển kho/hàng hoàn trong ngày.
6. Kiểm tra scan event và order pick/pack.
7. Nếu cần, tạo recount có 2 người xác nhận.
```

### Investigation checklist

- Có phiếu nhập chưa QC pass nhưng đã để vào available không?
- Có đơn packed nhưng chưa issue stock không?
- Có hàng hoàn đã nhận nhưng chưa disposition không?
- Có transfer nội bộ chưa complete không?
- Có adjustment nào không duyệt không?
- Có scan thiếu/lặp không?
- Có user sửa nhầm UOM không?

### Resolution

Nếu xác nhận lệch thật:

```text
Tạo Inventory Adjustment Request
→ Warehouse Manager nhập lý do
→ Finance/COO duyệt theo ngưỡng
→ Hệ thống sinh stock movement ADJUSTMENT
→ Audit log đầy đủ
```

Không được sửa trực tiếp stock balance.

### Prevention

- Bắt buộc scan khi pick/pack/handover.
- Đóng ca có checklist.
- Cảnh báo đơn packed quá lâu chưa handover.
- Cycle count theo SKU rủi ro cao.
- Training lại khu vực phát sinh lệch.

---

## 7.2. Stock ledger sai hoặc nghi ngờ bị ghi movement sai

### Trigger

- Stock balance không bằng tổng stock ledger.
- Một giao dịch sinh double movement.
- API retry tạo trùng stock movement.
- Movement có qty âm không hợp lệ.

### Severity

- P0 nếu ảnh hưởng nhiều SKU/module.
- P1 nếu chỉ ảnh hưởng một giao dịch nhưng liên quan đơn đã giao/QC.

### Immediate actions

```text
1. Tech Lead + Warehouse Manager mở incident P0/P1.
2. Tạm dừng action sinh movement cùng loại nếu đang lan.
3. Không xóa movement lỗi.
4. Trích xuất movement timeline.
5. Kiểm tra idempotency key, request log, transaction log.
```

### Investigation checklist

- API có retry không idempotency không?
- Transaction có commit một phần không?
- Worker/outbox có xử lý trùng event không?
- Manual adjustment có tạo movement kép không?
- Database migration có động vào stock table không?

### Resolution

Không xóa movement. Tạo reverse/correction movement:

```text
Nếu duplicate INBOUND_RECEIPT 10 chai:
→ tạo REVERSAL movement -10 chai, link tới movement gốc

Nếu issue sai batch:
→ reverse batch sai
→ issue lại batch đúng
```

### Prevention

- Idempotency bắt buộc cho action tạo movement.
- Transaction boundary rõ.
- Automated reconciliation: stock_balance = sum(stock_ledger).
- Test regression cho retry/timeout.

---

## 7.3. Batch/QC fail nhưng vẫn bán được

### Trigger

- Batch đang HOLD/FAIL vẫn xuất hiện trong available stock.
- Sales order reserve được batch chưa release.
- Kho pick được batch QC chưa pass.

### Severity

P0 nếu đã có đơn giao cho khách.  
P1 nếu mới reserve/pick nhưng chưa handover.

### Immediate actions

```text
1. QA Manager set batch = HOLD khẩn cấp.
2. Block sales allocation cho batch đó.
3. Inventory khóa pick/pack batch đó.
4. Sales/CSKH truy xuất đơn liên quan.
5. Nếu đã giao, lập danh sách khách/đại lý bị ảnh hưởng.
```

### Investigation checklist

- Batch QC status có bị sửa không?
- API allocation có check QC status không?
- Stock available có exclude HOLD/FAIL không?
- Có manual override nào không?
- Permission QA có bị cấp sai không?

### Resolution

- Đơn chưa giao: hủy pick và allocate batch khác.
- Đơn đã handover: COO/QA quyết định hold delivery, recall hoặc thông báo CSKH.
- Nếu batch thật sự pass nhưng status sai: QA làm correction có evidence.

### Prevention

- Hard rule: batch HOLD/FAIL không available.
- UI hiển thị QC status nổi bật.
- Automated test: không reserve được batch chưa pass.
- Field-level permission: chỉ QA được release.

---

## 7.4. QC pass nhầm batch hoặc nhập sai kết quả QC

### Trigger

- QA phát hiện pass nhầm.
- Kho phát hiện batch bất thường sau khi QC pass.
- Khách/kho phản hồi lỗi chất lượng.

### Severity

P1/P0 tùy đã bán chưa.

### Immediate actions

```text
1. QA set batch về HOLD.
2. Block allocation/pick/ship.
3. Truy xuất đơn, stock, vị trí kho liên quan.
4. Lưu evidence: form QC, ảnh, user action, timestamp.
5. Báo COO nếu đã có đơn ra khỏi kho.
```

### Resolution

- Re-inspection.
- Nếu pass thật: release lại, ghi note.
- Nếu fail: chuyển quarantine/damage/rework tùy rule.
- Nếu đã bán: kích hoạt complaint/recall playbook.

### Prevention

- QC release cần confirmation bước 2.
- Mẫu QC phải có checklist bắt buộc.
- QA không thể release nếu thiếu file/ảnh/kết quả test bắt buộc.

---

## 7.5. Nhập kho sai batch/hạn dùng

### Trigger

- Batch_no nhập sai.
- Expiry date sai.
- Nhà cung cấp giao batch khác chứng từ.
- Kho scan/nhập nhầm lô.

### Severity

P1 nếu đã xuất bán hoặc đưa vào sản xuất.  
P2 nếu còn trong receiving/quarantine.

### Immediate actions

```text
1. Lock batch/location liên quan.
2. Kiểm chứng chứng từ giao hàng, tem lô, ảnh hàng.
3. Kiểm tra stock movement đã phát sinh chưa.
4. Nếu đã xuất bán, trace order/batch.
```

### Resolution

- Nếu chưa có movement xuất: correction batch/expiry có approval QA + Warehouse.
- Nếu đã có movement xuất: tạo correction workflow, không sửa thẳng lịch sử.
- Nếu batch không xác thực: chuyển quarantine.

### Prevention

- Batch/expiry bắt buộc scan/nhập khi receiving.
- UI cảnh báo format sai.
- 2-person verification cho batch mới.

---

## 7.6. Nhập hàng NCC không đạt nhưng bị đưa vào kho khả dụng

### Trigger

- Inbound QC không đạt.
- Bao bì/lô/số lượng không khớp.
- Hàng bị đưa vào available trước khi QC pass.

### Severity

P1/P0 nếu đã dùng cho sản xuất hoặc bán.

### Immediate actions

```text
1. QA set related lot = HOLD.
2. Warehouse chuyển về quarantine location.
3. Purchasing liên hệ NCC nếu cần.
4. Truy xuất production/order đã dùng lot đó.
```

### Resolution

- Nếu chưa dùng: reject/return supplier.
- Nếu đã dùng sản xuất: QA đánh giá batch sản xuất liên quan.
- Nếu ảnh hưởng thành phẩm: hold batch thành phẩm.

### Prevention

- Inbound receipt không tự động available.
- QC pass mới chuyển available.
- Warehouse không thể override QC.

---

## 7.7. Đơn đã packed nhưng không tìm thấy khi bàn giao ĐVVC

### Trigger

- Manifest báo có đơn nhưng kho không tìm được hàng.
- Scan bàn giao thiếu mã.
- Khu vực đóng hàng không có đơn.

### Severity

P1 nếu ảnh hưởng chuyến giao lớn hoặc deadline ĐVVC.  
P2 nếu một vài đơn.

### Immediate actions

```text
1. Shipping Lead giữ manifest chưa confirm.
2. Kiểm tra scan history: picked, packed, staging location.
3. Tìm tại khu đóng hàng, khu chờ bàn giao, khu lỗi, khu hàng hoàn nếu có.
4. Kiểm tra đơn có bị đóng lại/hủy/đổi trạng thái không.
5. Nếu vẫn thiếu, tạo Missing Packed Order incident.
```

### Resolution

- Nếu tìm thấy: scan lại vào manifest và handover.
- Nếu không tìm thấy nhưng còn stock: re-pick/re-pack có approval.
- Nếu không còn stock: báo CSKH/Sales, đổi lịch giao hoặc hủy/hoàn.
- Nếu nghi thất thoát: Warehouse Manager + Security kiểm tra camera/audit.

### Prevention

- Mỗi packed order phải có staging location.
- Manifest chỉ nhận đơn có status PACKED.
- Scan bắt buộc trước handover.
- Không cho confirm manifest nếu thiếu scan.

---

## 7.8. Bàn giao ĐVVC sai số lượng hoặc sai mã đơn

### Trigger

- Số đơn bàn giao không khớp bảng/manifest.
- ĐVVC nhận thiếu/thừa.
- Scan nhầm đơn của carrier khác.

### Severity

P1 nếu manifest đã ký nhưng sai.  
P2 nếu phát hiện trước khi ký.

### Immediate actions

```text
1. Dừng ký xác nhận manifest.
2. Re-scan toàn bộ thùng/rổ liên quan.
3. Đối chiếu số đơn trên hệ thống, bảng bàn giao và thực tế.
4. Tách đơn sai carrier/kênh ra khu exception.
5. Chỉ ký khi manifest sạch.
```

### Resolution

- Correct manifest trước khi handover.
- Nếu đã ký sai: tạo manifest correction, có xác nhận ĐVVC.
- Nếu đơn đã đi sai carrier: CSKH cập nhật khách, Shipping Lead follow tracking.

### Prevention

- Carrier manifest theo từng ĐVVC.
- Mỗi thùng/rổ có mã định danh.
- Scan validate carrier/order status.
- Không cho handover đơn chưa PACKED.

---

## 7.9. Đơn đã handover nhưng trạng thái ERP chưa cập nhật

### Trigger

- ĐVVC đã nhận nhưng order vẫn PACKED.
- Manifest có ký nhưng hệ thống không ghi handover.
- API carrier timeout.

### Severity

P1 nếu nhiều đơn.  
P2 nếu đơn lẻ và còn evidence.

### Immediate actions

```text
1. Shipping Lead xác nhận manifest giấy/file/scan.
2. Tech kiểm tra API/event log.
3. Không tạo lại handover nếu chưa kiểm tra idempotency.
4. Tạm export danh sách đơn bị ảnh hưởng cho CSKH.
```

### Resolution

- Retry handover event nếu idempotent.
- Nếu event mất: tạo manual correction có approval + evidence.
- Update order/shipment status theo manifest thực tế.

### Prevention

- Handover action phải có idempotency key.
- Outbox event retry.
- Daily manifest reconciliation.

---

## 7.10. ĐVVC làm mất/hỏng đơn sau khi bàn giao

### Trigger

- Carrier báo mất đơn.
- Khách không nhận được hàng.
- Tracking bất thường.

### Severity

P2/P1 tùy số lượng/giá trị.

### Immediate actions

```text
1. CSKH tạo incident/link order.
2. Shipping xác minh tracking + manifest đã ký.
3. Finance ghi nhận khả năng claim carrier.
4. Nếu khách cần xử lý nhanh, Sales/CSKH đề xuất ship lại/hoàn tiền theo policy.
```

### Resolution

- Nếu carrier chịu trách nhiệm: claim ĐVVC.
- Nếu lỗi kho bàn giao sai: xử lý theo incident bàn giao.
- Nếu ship lại: tạo order replacement có link incident.

### Prevention

- Lưu manifest/evidence bàn giao.
- Tracking sync với carrier.
- Báo cáo tỷ lệ mất/hỏng theo ĐVVC.

---

## 7.11. Hàng hoàn nhập sai trạng thái

### Trigger

- Hàng hoàn còn dùng được nhưng đưa vào lab/hỏng.
- Hàng không dùng được nhưng nhập lại kho bán.
- Không quét hàng hoàn nhưng đã tạo phiếu nhập.

### Severity

P1 nếu hàng lỗi quay lại bán.  
P2 nếu chưa bán.

### Immediate actions

```text
1. Warehouse Manager khóa return lot/order.
2. Đưa hàng về khu hàng hoàn/quarantine.
3. Kiểm tra return scan, ảnh tình trạng, người phân loại.
4. CSKH xác minh lý do hoàn.
```

### Resolution

- Re-inspection bởi Warehouse + QA nếu cần.
- Chọn disposition đúng: reusable / damaged / lab / quarantine.
- Nếu đã nhập kho bán và có đơn: trace order, hold nếu cần.

### Prevention

- Returns workflow bắt buộc scan.
- Disposition không được bỏ trống.
- Hàng hoàn không tự vào available.
- Ảnh tình trạng bắt buộc với hàng nghi lỗi.

---

## 7.12. Hàng hoàn không có mã đơn/mã vận đơn rõ ràng

### Trigger

- Shipper trả hàng nhưng không rõ order.
- Tem/mã vận đơn rách.
- Không tìm được trong hệ thống.

### Severity

P2 thường, P1 nếu số lượng lớn.

### Immediate actions

```text
1. Đưa vào khu Return Exception.
2. Không nhập available.
3. Chụp ảnh sản phẩm/tem/bao bì.
4. CSKH/Shipping tra cứu theo tên khách, số điện thoại, SKU, ngày giao.
5. Gán mã exception tạm.
```

### Resolution

- Nếu xác định order: chuyển về return workflow chuẩn.
- Nếu không xác định sau SLA: xử lý theo policy hàng vô chủ/không đối chiếu.

### Prevention

- ĐVVC phải trả kèm manifest hàng hoàn.
- Return receiving scan barcode trước.
- Training shipper/warehouse.

---

## 7.13. Đơn hàng bán vượt tồn khả dụng

### Trigger

- Sales order confirmed nhưng không đủ tồn.
- Nhiều kênh cùng bán một SKU.
- Reserve stock bị lỗi.

### Severity

P1 nếu nhiều đơn/kênh.  
P2 nếu đơn lẻ.

### Immediate actions

```text
1. Pause allocation SKU nếu oversell đang lan.
2. Kiểm tra available_stock formula.
3. Kiểm tra reservation table.
4. Xác định đơn ưu tiên giữ hàng.
5. Báo Sales/CSKH danh sách đơn thiếu hàng.
```

### Resolution

- Allocate lại theo rule ưu tiên.
- Backorder/hủy/đổi SKU theo policy.
- Nếu lỗi integration, tạm giảm/sync lại stock channel.

### Prevention

- Reserve stock transaction atomic.
- Sync tồn đa kênh có buffer.
- Không bán batch HOLD/FAIL/cận date ngoài policy.

---

## 7.14. Sales sửa discount/giá vượt quyền

### Trigger

- Đơn có discount vượt ngưỡng nhưng không có duyệt.
- Giá bán khác price list.
- Promotion áp sai kênh.

### Severity

P1 nếu ảnh hưởng nhiều đơn hoặc gây thất thoát lớn.  
P2 nếu đơn lẻ.

### Immediate actions

```text
1. Finance/Sales Admin flag đơn.
2. Tạm hold đơn chưa giao nếu giá sai nghiêm trọng.
3. Kiểm tra approval log và price rule.
4. Xác định user/action/time.
```

### Resolution

- Nếu chưa giao: chỉnh/hủy/duyệt lại theo policy.
- Nếu đã giao: Finance xử lý adjustment, Sales xử lý khách nếu cần.
- Nếu quyền sai: Security/Admin sửa permission.

### Prevention

- Discount > threshold bắt buộc approval.
- Frontend không hiển thị cost/price fields cho role không quyền.
- Audit log sensitive fields.

---

## 7.15. Sai công nợ/COD sau đối soát

### Trigger

- COD carrier trả không khớp đơn delivered.
- Payment captured nhưng order chưa paid.
- Finance phát hiện thiếu/thừa tiền.

### Severity

P1 nếu số tiền lớn/nhiều đơn.  
P2 nếu đơn lẻ.

### Immediate actions

```text
1. Finance tạo reconciliation incident.
2. Export order/payment/shipment liên quan.
3. Đối chiếu carrier COD report/bank statement.
4. Không chỉnh paid status thủ công nếu chưa có evidence.
```

### Resolution

- Import lại reconciliation nếu do lỗi file.
- Tạo payment correction có approval.
- Claim carrier nếu thiếu COD.

### Prevention

- Reconciliation rule rõ.
- File import validate format.
- Payment/COD mapping theo order/shipment id.

---

## 7.16. Nhà máy/gia công giao thiếu hàng

### Trigger

- Hàng về kho ít hơn đơn gia công.
- Biên bản giao nhận không khớp.
- Một phần hàng thiếu theo SKU/quy cách.

### Severity

P1 nếu ảnh hưởng launch/đơn lớn.  
P2 nếu ít và có kế hoạch bù.

### Immediate actions

```text
1. Warehouse receiving ghi nhận shortage.
2. Không close subcontract order.
3. Purchasing/Subcontract Owner liên hệ nhà máy.
4. QA kiểm tra lô đã nhận tách riêng.
5. Finance hold final payment phần liên quan.
```

### Resolution

- Nhận partial receipt.
- Tạo factory claim/shortage claim.
- Theo dõi bổ sung hoặc điều chỉnh thanh toán.

### Prevention

- Subcontract receiving phải đối chiếu PO/subcontract order.
- Biên bản bàn giao có số lượng từng SKU/batch.
- Final payment chỉ mở khi receiving + QC đạt.

---

## 7.17. Nhà máy/gia công giao hàng lỗi chất lượng

### Trigger

- QC nhận hàng gia công fail.
- Bao bì sai, mẫu mã sai, chất lượng không đạt.
- Lỗi phát hiện trong vòng 3–7 ngày sau nhận.

### Severity

P1/P0 nếu hàng đã nhập available hoặc đã bán.

### Immediate actions

```text
1. QA set lot/batch = HOLD.
2. Warehouse chuyển quarantine.
3. Purchasing/Subcontract Owner mở claim nhà máy.
4. Lưu evidence: ảnh, video, mẫu, biên bản, kết quả QC.
5. Finance hold final payment nếu chưa thanh toán.
```

### Resolution

- Reject toàn bộ/lô một phần.
- Yêu cầu rework/replace/refund theo hợp đồng.
- Nếu một phần đạt: QA release partial theo batch/location.
- Nếu đã bán: trace order và kích hoạt CSKH/recall nếu cần.

### Prevention

- Sample approval bắt buộc trước mass production.
- Receiving QC checklist cho hàng gia công.
- Claim window 3–7 ngày được cấu hình nhắc việc.

---

## 7.18. Sản xuất hàng loạt khi mẫu chưa duyệt

### Trigger

- Nhà máy bắt đầu mass production nhưng sample status chưa approved.
- ERP không có evidence duyệt mẫu.
- Brand/R&D/QA chưa sign-off.

### Severity

P1, có thể P0 nếu đã giao hàng/bán.

### Immediate actions

```text
1. Dừng nhận hàng hoặc set toàn bộ lô liên quan = HOLD.
2. Purchasing yêu cầu nhà máy cung cấp bằng chứng duyệt mẫu nếu có.
3. R&D/QA/Brand kiểm tra sample/spec.
4. Finance hold payment.
```

### Resolution

- Nếu mẫu thực tế đạt và được duyệt bổ sung: ghi deviation + approval.
- Nếu không đạt: reject/rework/claim nhà máy.

### Prevention

- Không cho chuyển subcontract order sang MASS_PRODUCTION nếu sample chưa APPROVED.
- Approval sample phải có file/ảnh/spec.

---

## 7.19. Chuyển NVL/bao bì cho nhà máy sai số lượng hoặc sai lô

### Trigger

- Factory báo nhận thiếu/thừa NVL/bao bì.
- ERP transfer record khác biên bản bàn giao.
- Lô NVL chuyển sai.

### Severity

P1 nếu ảnh hưởng sản xuất/traceability.  
P2 nếu phát hiện sớm.

### Immediate actions

```text
1. Lock subcontract material transfer.
2. Đối chiếu biên bản bàn giao, stock movement, ảnh, vận chuyển.
3. Kiểm tra batch/lot đã chuyển.
4. Nếu nhà máy đã dùng: QA đánh giá tác động tới batch thành phẩm.
```

### Resolution

- Correction transfer có approval.
- Nếu lô sai đã dùng: hold thành phẩm liên quan.
- Nếu thiếu: issue bổ sung theo transfer mới.

### Prevention

- Material transfer scan batch/lot.
- Biên bản bàn giao bắt buộc attachment.
- Factory acknowledgment trước khi sản xuất.

---

## 7.20. Complaint khách hàng liên quan batch/lô

### Trigger

- Khách phản hồi kích ứng/hỏng/sai chất lượng.
- CSKH ghi nhận nhiều complaint cùng SKU/batch.
- Social/KOL phản ánh lỗi sản phẩm.

### Severity

P0 nếu liên quan an toàn người dùng hoặc nhiều complaint.  
P1 nếu complaint nghiêm trọng nhưng ít.

### Immediate actions

```text
1. CSKH tạo complaint linked order/batch.
2. QA set batch nghi ngờ = HOLD nếu có dấu hiệu lặp.
3. Trace đơn đã bán theo batch.
4. Thu thập ảnh/video/phản hồi khách.
5. Báo COO/CEO nếu có rủi ro truyền thông hoặc an toàn.
```

### Resolution

- QA/R&D đánh giá nguyên nhân.
- Nếu batch lỗi: recall/replace/refund theo policy.
- Nếu lỗi đơn lẻ: CSKH xử lý case.
- Nếu liên quan claim/KOL: Marketing điều chỉnh nội dung.

### Prevention

- Complaint dashboard theo batch.
- Threshold cảnh báo số complaint/batch.
- Claim library và nội dung KOL phải qua duyệt.

---

## 7.21. Master data sai gây lỗi vận hành

### Trigger

- SKU trùng/sai UOM.
- BOM/công thức sai.
- Barcode sai.
- Warehouse location sai.
- Carrier mapping sai.

### Severity

P1 nếu sai master data gây lỗi tồn/đơn/giao hàng.  
P2/P3 tùy mức độ.

### Immediate actions

```text
1. Data Owner khóa master data record.
2. Xác định transaction đã phát sinh từ master data sai.
3. Tạm ngừng sử dụng SKU/location/carrier mapping liên quan.
4. Ghi lại danh sách chứng từ bị ảnh hưởng.
```

### Resolution

- Correct master data có approval.
- Reprocess/correction chứng từ bị ảnh hưởng nếu cần.
- Không xóa master data đã có giao dịch.

### Prevention

- Maker-checker cho master data quan trọng.
- Duplicate detection.
- Barcode/UOM validation.

---

## 7.22. User thao tác sai cần đảo giao dịch

### Trigger

- Nhập nhầm số lượng.
- Xuất nhầm batch.
- Hủy nhầm đơn.
- Chọn sai carrier.
- Set nhầm disposition hàng hoàn.

### Severity

Theo impact: P1/P2/P3.

### Immediate actions

```text
1. Không xóa chứng từ.
2. Ghi nhận user, thời điểm, chứng từ, lý do.
3. Kiểm tra giao dịch đã ảnh hưởng stock/QC/finance/shipping chưa.
4. Chọn reverse/correction workflow phù hợp.
```

### Resolution

- Nếu transaction chưa posted: cancel có approval.
- Nếu đã posted: reverse transaction.
- Nếu đã handover/delivered: cần business approval và communication.

### Prevention

- Confirmation step cho action nguy hiểm.
- UI preview trước submit.
- Training user theo SOP.

---

## 7.23. Lỗi tích hợp ĐVVC/website/marketplace

### Trigger

- Đơn không đồng bộ vào ERP.
- Trạng thái giao hàng không cập nhật.
- Carrier API timeout.
- Marketplace stock sync sai.

### Severity

P1 nếu ảnh hưởng nhiều đơn/kênh.  
P2 nếu một phần và có workaround.

### Immediate actions

```text
1. Tech Lead kiểm tra integration logs.
2. Tạm pause sync nếu gây sai dữ liệu.
3. Export danh sách đơn bị ảnh hưởng.
4. Sales/CSKH được thông báo trạng thái tạm.
5. Không import lại trùng nếu chưa kiểm tra idempotency.
```

### Resolution

- Retry job/event.
- Manual import/export có kiểm soát.
- Reconcile order/shipment/status sau khi API phục hồi.

### Prevention

- Idempotency key cho inbound order.
- Integration dashboard.
- Dead-letter queue.
- Alert khi sync fail quá ngưỡng.

---

## 7.24. Hệ thống chậm hoặc không truy cập được

### Trigger

- User không đăng nhập được.
- API latency cao.
- Scan chậm khi bàn giao.
- Database overloaded.

### Severity

P0 nếu dừng vận hành kho/sales.  
P1 nếu một module chính chậm.

### Immediate actions

```text
1. Tech Lead mở incident kỹ thuật.
2. Kiểm tra healthcheck, API latency, DB, Redis, queue.
3. Xác định có release mới không.
4. Nếu liên quan release, cân nhắc rollback.
5. Nếu kho đang bàn giao, kích hoạt offline fallback tạm thời theo GoLive Runbook.
```

### Resolution

- Scale service nếu cần.
- Rollback release.
- Kill job nặng nếu đang gây nghẽn.
- Optimize query/index nếu xác định rõ.

### Prevention

- Monitoring + alert.
- Load test các luồng scan/pick/pack.
- Report nặng chạy async.

---

## 7.25. Lỗi bảo mật/phân quyền

### Trigger

- User thấy dữ liệu không được phép: giá vốn, lương, payout, công nợ.
- User thao tác được action ngoài quyền.
- Tài khoản nghi bị dùng sai.
- Export dữ liệu bất thường.

### Severity

P0 nếu lộ dữ liệu nhạy cảm hoặc quyền phá kiểm soát.  
P1 nếu lỗi quyền trong một module.

### Immediate actions

```text
1. Security/Admin khóa user/session nếu cần.
2. Tạm disable permission/role nghi lỗi.
3. Export audit log liên quan.
4. Xác định dữ liệu đã bị xem/export/sửa.
5. Báo CEO/COO nếu có dữ liệu nhạy cảm.
```

### Resolution

- Fix role/permission.
- Rotate password/token nếu nghi compromise.
- Thu hồi export nếu có thể.
- Thông báo nội bộ theo policy.

### Prevention

- Permission test trước release.
- Field-level permission cho giá vốn/lương/payout.
- Sensitive export approval.
- MFA cho admin/finance/QA critical roles.

---

## 7.26. Database migration lỗi sau release

### Trigger

- Migration fail.
- Column/index missing.
- Data migration sai.
- App chạy version mới nhưng DB chưa tương thích.

### Severity

P0/P1.

### Immediate actions

```text
1. Dừng deployment tiếp theo.
2. Tech Lead kiểm tra migration logs.
3. Xác định đã tác động prod chưa.
4. Nếu prod bị ảnh hưởng, kích hoạt rollback/restore plan.
5. Không chạy lại migration thủ công nếu chưa review.
```

### Resolution

- Rollback app nếu migration chưa destructive.
- Patch migration nếu safe.
- Restore backup nếu data corruption nghiêm trọng.

### Prevention

- Migration dry-run trên staging/UAT.
- Backward-compatible migration.
- Backup trước release.
- Không deploy ngoài window đã duyệt.

---

## 7.27. Audit log thiếu hoặc không ghi nhận action quan trọng

### Trigger

- Action tạo/sửa/hủy không có audit.
- Sensitive field đổi nhưng audit không có before/after.
- Incident không truy vết được user.

### Severity

P1 nếu liên quan stock/QC/finance/security.  
P2 nếu action ít rủi ro.

### Immediate actions

```text
1. Tech Lead xác định scope action thiếu audit.
2. Tạm dừng action nếu không thể kiểm soát.
3. Kiểm tra application log/API log thay thế.
4. Bổ sung audit hotfix nếu action critical.
```

### Resolution

- Fix audit hook.
- Backfill audit nếu có log đủ evidence.
- Nếu không backfill được, ghi incident note.

### Prevention

- Definition of Done: action critical phải có audit.
- Automated test audit.
- Audit checklist trong code review.

---

## 8. Risk Register gợi ý cho Phase 1

| ID | Rủi ro | Xác suất | Ảnh hưởng | Kiểm soát chính | Owner |
|---|---|---:|---:|---|---|
| R-001 | Tồn kho lệch do scan/đóng ca không chuẩn | Cao | Cao | Stock ledger + daily reconciliation + scan | Warehouse Manager |
| R-002 | Batch chưa pass QC vẫn bán | Trung bình | Rất cao | QC gate + allocation rule | QA Manager |
| R-003 | Hàng hoàn lỗi quay lại kho bán | Trung bình | Cao | Return disposition + quarantine | Warehouse Manager |
| R-004 | Bàn giao ĐVVC thiếu đơn | Cao | Cao | Manifest + scan verify | Shipping Lead |
| R-005 | Gia công ngoài giao hàng lỗi | Trung bình | Cao | Receiving QC + claim 3–7 ngày | QA + Purchasing |
| R-006 | Chuyển NVL/bao bì sai cho nhà máy | Trung bình | Cao | Transfer scan + biên bản | Subcontract Owner |
| R-007 | Oversell đa kênh | Trung bình | Cao | Reserve stock atomic + stock sync buffer | Sales Ops |
| R-008 | COD/công nợ lệch | Trung bình | Cao | Reconciliation + evidence | Finance |
| R-009 | User xem dữ liệu nhạy cảm | Thấp/Trung bình | Cao | RBAC + field-level permission | Security/Admin |
| R-010 | Release làm hỏng nghiệp vụ chính | Trung bình | Cao | CI gate + smoke test + rollback | Tech Lead |

---

## 9. Communication plan khi incident xảy ra

### 9.1. Template thông báo nội bộ ban đầu

```text
[INCIDENT][P1] Bàn giao ĐVVC thiếu đơn - Carrier: GHN - 2026-xx-xx

Thời điểm phát hiện: ...
Người phát hiện: ...
Triệu chứng: Manifest có 120 đơn, scan thực tế 118 đơn.
Phạm vi ảnh hưởng: Carrier GHN, batch handover 14:00.
Hành động containment: Chưa ký xác nhận bàn giao, đang re-scan.
Owner: Shipping Lead
ETA cập nhật tiếp theo: 30 phút
```

### 9.2. Template cập nhật tiến độ

```text
[UPDATE][INC-xxx]

Tình trạng hiện tại: ...
Đã làm: ...
Kết quả: ...
Vấn đề còn lại: ...
Quyết định cần duyệt: ...
ETA tiếp theo: ...
```

### 9.3. Template đóng incident

```text
[CLOSED][INC-xxx]

Root cause: ...
Impact: ...
Data correction: ...
Evidence: ...
Owner xác nhận: ...
CAPA: ...
Follow-up deadline: ...
```

---

## 10. Evidence checklist

Mỗi incident cần lưu evidence phù hợp:

| Loại incident | Evidence bắt buộc |
|---|---|
| Tồn kho lệch | Stock ledger, count sheet, ảnh khu vực/bin, phiếu nhập/xuất |
| QC/batch | QC checklist, ảnh mẫu, kết quả test, approval log |
| Bàn giao ĐVVC | Manifest, scan log, ảnh thùng/rổ, ký xác nhận ĐVVC |
| Hàng hoàn | Return scan, ảnh tình trạng, reason, disposition decision |
| Gia công ngoài | Subcontract order, biên bản bàn giao NVL/bao bì, mẫu duyệt, QC report, claim |
| Finance/COD | Carrier COD file, bank statement, reconciliation report |
| Security | Audit log, session log, export log, permission snapshot |
| System/API | Request log, error log, trace id, deployment version |

---

## 11. Data correction policy

### 11.1. Không được correction nếu thiếu 4 thứ

```text
1. Incident ID
2. Evidence
3. Business approval
4. Audit log/correction note
```

### 11.2. Các kiểu correction hợp lệ

| Tình huống | Correction hợp lệ |
|---|---|
| Tồn kho thiếu/thừa | Inventory adjustment có approval |
| Movement tạo trùng | Reversal movement link movement gốc |
| Batch status sai | QC correction workflow |
| Đơn sai trạng thái | Order/shipment correction workflow |
| Hàng hoàn sai disposition | Return disposition correction |
| Payment/COD sai | Payment correction/reconciliation adjustment |
| Master data sai | Master data correction + impact review |

### 11.3. Correction phải có reason code

Reason code gợi ý:

```text
USER_INPUT_ERROR
SCAN_MISSING
SCAN_DUPLICATE
CARRIER_MISMATCH
QC_ENTRY_ERROR
SUPPLIER_DELIVERY_MISMATCH
FACTORY_SHORTAGE
FACTORY_QUALITY_ISSUE
INTEGRATION_FAILURE
PAYMENT_RECONCILIATION_ERROR
MASTER_DATA_ERROR
SYSTEM_BUG
SECURITY_PERMISSION_ERROR
```

---

## 12. CAPA framework

Mẫu CAPA:

```text
Incident ID:
Loại sự cố:
Ngày phát hiện:
Owner:
Root cause:
Corrective action:
Preventive action:
Tài liệu/SOP cần sửa:
Hệ thống cần sửa:
Training cần làm:
Deadline:
Người duyệt đóng CAPA:
```

### 12.1. Ví dụ CAPA — thiếu đơn khi bàn giao ĐVVC

```text
Root cause:
Packed order không được gán staging location, nhân viên đặt nhầm sang thùng carrier khác.

Corrective action:
Re-scan toàn bộ manifest, tìm lại đơn, cập nhật bàn giao đúng.

Preventive action:
Bắt buộc staging location cho từng packed order.
Scan validate carrier trước khi đưa vào manifest.
Training lại kho.
```

### 12.2. Ví dụ CAPA — hàng hoàn lỗi quay lại kho bán

```text
Root cause:
Nhân viên chọn disposition "Còn sử dụng" nhưng không kiểm tình trạng bên trong.

Corrective action:
Set batch/order liên quan HOLD, kiểm lại hàng hoàn.

Preventive action:
Bắt buộc checklist tình trạng + ảnh trước khi chọn "Còn sử dụng".
QA review với hàng mỹ phẩm có dấu hiệu mở nắp/hỏng vỏ.
```

---

## 13. Incident report template

```text
# Incident Report

Incident ID:
Ngày/giờ phát hiện:
Người phát hiện:
Module:
Severity:
Incident Owner:

## 1. Mô tả sự cố

## 2. Phạm vi ảnh hưởng
- SKU:
- Batch:
- Đơn hàng:
- Khách hàng:
- Kho/location:
- ĐVVC:
- Nhà máy/NCC:

## 3. Timeline
| Thời gian | Sự kiện | Người thực hiện |
|---|---|---|

## 4. Hành động containment

## 5. Root cause

## 6. Data correction

## 7. Evidence

## 8. Communication

## 9. CAPA

## 10. Sign-off
- Business Owner:
- Tech Lead:
- QA/Finance/Warehouse nếu liên quan:
```

---

## 14. Checklist phản ứng nhanh theo module

### 14.1. Warehouse

```text
[ ] Khóa SKU/batch/location nếu cần
[ ] Kiểm stock ledger
[ ] Kiểm physical stock
[ ] Kiểm phiếu nhập/xuất/chuyển kho
[ ] Kiểm scan event
[ ] Kiểm daily closing report
[ ] Tạo adjustment/reversal nếu đã duyệt
```

### 14.2. QC/Batch

```text
[ ] Set HOLD nếu có nghi ngờ
[ ] Kiểm QC checklist/evidence
[ ] Trace batch tới order/production/subcontract
[ ] Block allocation/pick/ship
[ ] Lập CAPA nếu có lỗi chất lượng
```

### 14.3. Shipping/Handover

```text
[ ] Dừng ký manifest nếu chưa đủ
[ ] Re-scan thùng/rổ/khu vực
[ ] Kiểm carrier mapping
[ ] Kiểm order status PACKED/HANDED_OVER
[ ] Lưu manifest/evidence
```

### 14.4. Returns

```text
[ ] Đưa hàng vào khu Return Exception/Quarantine
[ ] Quét return nếu có mã
[ ] Chụp ảnh tình trạng
[ ] Không nhập available nếu chưa disposition
[ ] Link complaint/order/batch nếu có
```

### 14.5. Subcontract Manufacturing

```text
[ ] Kiểm subcontract order
[ ] Kiểm material transfer
[ ] Kiểm sample approval
[ ] Kiểm receiving quantity
[ ] QA inspection
[ ] Hold final payment nếu có lỗi
[ ] Tạo factory claim trong 3–7 ngày nếu cần
```

### 14.6. Security

```text
[ ] Khóa user/session nếu nghi ngờ
[ ] Export audit/session/export log
[ ] Kiểm role/permission snapshot
[ ] Đánh giá dữ liệu bị xem/sửa/export
[ ] Reset credential/token nếu cần
```

---

## 15. Definition of Done cho một incident

Một incident chỉ được đóng khi:

```text
[ ] Đã xác định severity đúng
[ ] Đã có owner
[ ] Đã containment xong
[ ] Đã xác định impact
[ ] Đã có root cause hoặc lý do tạm nếu chưa thể xác định ngay
[ ] Đã xử lý/correction theo quy trình
[ ] Đã test/validate lại
[ ] Đã thông báo người liên quan
[ ] Đã lưu evidence
[ ] Đã tạo CAPA nếu là P0/P1 hoặc lỗi lặp lại
[ ] Đã sign-off bởi owner phù hợp
```

---

## 16. Quy tắc vàng

```text
1. Sự cố hàng/tồn/QC: khóa trước, điều tra sau.
2. Không xóa giao dịch, chỉ reverse/correction.
3. Không sửa database trực tiếp.
4. Không release batch nếu thiếu evidence.
5. Không bàn giao ĐVVC nếu manifest chưa sạch.
6. Hàng hoàn không tự quay lại kho bán.
7. Gia công ngoài chưa QC pass thì chưa nhập available.
8. Payment/COD không sửa nếu thiếu đối soát.
9. Permission sai là incident bảo mật, không phải lỗi nhỏ.
10. Sau P0/P1 luôn có postmortem và CAPA.
```

---

## 17. Tài liệu cần cập nhật sau incident

Tùy loại sự cố, các tài liệu có thể cần sửa:

| Loại incident | Tài liệu cần review |
|---|---|
| User thao tác sai | SOP Training Manual |
| Workflow thiếu bước | Process Flow / PRD / Screen List |
| Dữ liệu sai nghĩa | Data Dictionary |
| Quyền sai | Permission Matrix / Security Standard |
| API lỗi | API Contract / Coding Standard |
| DB lỗi | Database Schema Standard |
| Test thiếu | UAT / QA Test Strategy |
| Go-live issue | GoLive Runbook |
| Tích hợp lỗi | Integration Spec |

---

## 18. Kết luận

Phase 1 của ERP công ty mỹ phẩm không chỉ cần chạy được đơn hàng và tồn kho. Nó phải giữ được các điểm kiểm soát sống còn:

```text
Batch
QC
Tồn kho
Hàng hoàn
Bàn giao ĐVVC
Gia công ngoài
Công nợ/COD
Audit
Phân quyền
```

Tài liệu này là bộ phản ứng nhanh khi hệ thống hoặc vận hành có vấn đề. Dùng đúng playbook này sẽ giúp công ty tránh 3 kiểu thiệt hại lớn:

```text
Thiệt hại hàng hóa
Thiệt hại tiền
Thiệt hại uy tín thương hiệu
```

Một hệ thống mạnh không phải hệ thống không bao giờ có lỗi.  
Một hệ thống mạnh là hệ thống **phát hiện lỗi sớm, khoanh vùng nhanh, sửa có kiểm soát, và không để lỗi lặp lại**.
