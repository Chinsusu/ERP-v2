# 27_ERP_GoLive_Runbook_Hypercare_Phase1_MyPham_v1

**Dự án:** Web ERP Phase 1 cho công ty mỹ phẩm  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-24  
**Tài liệu:** Go-Live Runbook & Hypercare Plan  
**Phạm vi:** Phase 1 — Master Data, Purchasing, QC, Inventory/WMS, Sales Order, Shipping/Handover, Returns, Subcontract Manufacturing, Finance basic controls  
**Tech stack tham chiếu:** Go Backend, PostgreSQL, React/Next.js Frontend, REST/OpenAPI, Redis/Queue, S3/MinIO, Docker/CI-CD  

---

## 1. Mục tiêu tài liệu

Tài liệu này dùng để điều phối toàn bộ quá trình **đưa ERP Phase 1 vào vận hành thật**.

Nó trả lời 7 câu hỏi sống còn:

1. Trước ngày go-live cần chuẩn bị gì?
2. Ai chịu trách nhiệm từng bước?
3. Dữ liệu nào phải khóa, dữ liệu nào được cập nhật?
4. Ngày go-live chạy theo thứ tự nào?
5. Khi phát hiện lỗi thì xử lý thế nào?
6. Khi nào được phép rollback?
7. Sau go-live 1–2 tuần cần hỗ trợ người dùng thế nào?

Mục tiêu không phải “bật hệ thống cho có”, mà là:

```text
Go-live sạch dữ liệu
Go-live có kiểm soát
Go-live không làm đứt vận hành kho/sales/giao hàng
Go-live có khả năng rollback nếu sự cố nghiêm trọng
```

---

## 2. Nguyên tắc go-live

### 2.1. Không go-live nếu dữ liệu nền chưa sạch

ERP Phase 1 phụ thuộc rất mạnh vào dữ liệu gốc và tồn kho. Không được go-live nếu:

- mã SKU còn trùng
- mã nguyên vật liệu chưa thống nhất
- batch/lô chưa có format
- tồn kho opening balance chưa đối soát
- user/role chưa phân quyền
- warehouse location chưa dựng
- carrier/ĐVVC chưa mapping
- QC status chưa rõ
- số liệu công nợ opening chưa xác nhận

### 2.2. Không go-live nếu chưa test luồng hàng thật

Phải test ít nhất các luồng thật sau:

```text
Nhập kho → QC → nhập khả dụng
Đơn hàng → reserve → pick → pack → bàn giao ĐVVC
Hàng hoàn → kiểm tra → phân loại còn dùng/không dùng
Gia công ngoài → chuyển NVL/bao bì → duyệt mẫu → nhận hàng → QC
Kiểm kê cuối ngày → đối soát → đóng ca
```

### 2.3. Không sửa dữ liệu bằng tay ngoài quy trình

Trong thời gian cutover và hypercare:

- không sửa trực tiếp database
- không sửa tồn bằng tay ngoài adjustment có phê duyệt
- không đổi QC status không có audit
- không sửa đơn đã handover nếu không có quyền đặc biệt
- không chỉnh stock ledger

### 2.4. Một nguồn sự thật duy nhất

Sau thời điểm go-live chính thức:

```text
ERP là nguồn sự thật chính cho tồn kho, đơn hàng, QC, giao hàng, hàng hoàn và gia công ngoài.
```

Excel chỉ được dùng như:

- file đối soát
- file backup tham chiếu
- file migration archive

Không dùng Excel để vận hành song song vô thời hạn.

---

## 3. Phạm vi Go-Live Phase 1

### 3.1. Module đưa vào vận hành

| Module | Trạng thái Go-Live | Ghi chú |
|---|---:|---|
| Master Data | Bắt buộc | SKU, NVL, kho, NCC, khách, ĐVVC, user |
| Purchasing | Bắt buộc | PR/PO/nhận hàng cơ bản |
| QC | Bắt buộc | QC đầu vào, QC nhận hàng gia công, hold/pass/fail |
| Inventory/WMS | Bắt buộc | stock ledger, batch, location, movement |
| Warehouse Daily Board | Bắt buộc | nhận đơn, soạn/đóng, kiểm kê, đối soát cuối ca |
| Sales Order | Bắt buộc | tạo đơn, reserve, trạng thái đơn |
| Pick/Pack | Bắt buộc | soạn hàng, đóng gói, kiểm tra |
| Shipping/Handover | Bắt buộc | manifest, scan, bàn giao ĐVVC |
| Returns | Bắt buộc | hàng hoàn, phân loại, nhập lại/kho hỏng/lab |
| Subcontract Manufacturing | Bắt buộc | gia công ngoài, chuyển NVL/bao bì, duyệt mẫu, nhận hàng |
| Finance Basic | Bắt buộc một phần | công nợ/thu chi/đối soát cơ bản |
| Reports/KPI Basic | Bắt buộc | dashboard vận hành tối thiểu |
| HRM/KOL/CRM nâng cao | Chưa go-live sâu | Parking lot Phase 2 |

---

## 4. Định nghĩa trạng thái Go-Live

### 4.1. Soft Go-Live

Hệ thống chạy với dữ liệu thật nhưng trong phạm vi hạn chế:

- 1 kho
- 1 nhóm user
- 1–2 kênh bán
- 1–2 ĐVVC
- giới hạn số lượng đơn/ngày

Mục tiêu: phát hiện lỗi vận hành trước khi mở full.

### 4.2. Full Go-Live

ERP trở thành hệ thống vận hành chính cho toàn bộ phạm vi Phase 1.

### 4.3. Hypercare

Giai đoạn 1–2 tuần sau go-live, đội dự án trực chiến để:

- sửa lỗi nhanh
- hỗ trợ user
- xử lý data issue
- ổn định quy trình
- giảm dần phụ thuộc vào team triển khai

---

## 5. Vai trò và trách nhiệm

### 5.1. War Room Go-Live Team

| Vai trò | Trách nhiệm |
|---|---|
| Executive Sponsor / CEO | Chốt Go/No-Go, xử lý quyết định lớn |
| Product Owner Business | Chốt nghiệp vụ, ưu tiên issue |
| Project Manager | Điều phối timeline, checklist, war room |
| Tech Lead Backend | API, database, transaction, bug backend |
| Tech Lead Frontend | UI, form, table, scan UX |
| DevOps Lead | deploy, rollback, backup, monitoring |
| QA Lead | smoke test, regression, defect tracking |
| Data Migration Lead | import, verify, reconcile dữ liệu |
| Warehouse Lead | xác nhận kho, pick/pack, handover, returns |
| QA/QC Lead | xác nhận QC status, batch release |
| Sales Lead | xác nhận đơn hàng, kênh bán, reserve |
| Finance Lead | xác nhận công nợ, COD, dữ liệu tiền |
| Factory/Subcontract Lead | xác nhận luồng gia công ngoài |
| Support Lead | tiếp nhận ticket và điều phối hỗ trợ |

---

## 6. Ma trận trách nhiệm RACI

| Hạng mục | CEO | PM | Tech Lead | Data Lead | Warehouse | QC | Sales | Finance |
|---|---|---|---|---|---|---|---|---|
| Go/No-Go decision | A | R | C | C | C | C | C | C |
| Freeze master data | C | R | C | A/R | C | C | C | C |
| Import opening stock | C | R | C | A/R | A/R | C | C | C |
| Verify batch/QC | C | C | C | R | C | A/R | - | - |
| Smoke test order-to-cash | C | A/R | R | C | R | C | A/R | C |
| Smoke test handover | C | A/R | R | C | A/R | - | C | - |
| Rollback decision | A | R | A/R | C | C | C | C | C |
| Hypercare triage | C | A/R | R | C | C | C | C | C |

Ký hiệu:

```text
A = Accountable
R = Responsible
C = Consulted
I = Informed
```

---

## 7. Timeline tổng thể

```text
T-30 đến T-15: chuẩn bị dữ liệu, readiness, training, UAT regression
T-14 đến T-8: migration rehearsal, smoke test, fix lỗi P0/P1
T-7 đến T-4: data freeze mềm, user training cuối, go-live checklist
T-3 đến T-1: final migration, backup, cutover rehearsal, war room setup
T-Day: go-live
T+1 đến T+3: hypercare cao điểm
T+4 đến T+14: hypercare ổn định
T+15: chuyển sang support vận hành thường xuyên
```

---

## 8. Checklist T-30 đến T-15

### 8.1. Business readiness

| Checklist | Owner | Status |
|---|---|---|
| Chốt scope Phase 1 | PM/CEO | ☐ |
| Chốt danh sách user go-live | HR/Admin | ☐ |
| Chốt vai trò và quyền | Product Owner | ☐ |
| Chốt quy trình kho hằng ngày | Warehouse Lead | ☐ |
| Chốt quy trình nhập/xuất/đóng hàng/hàng hoàn | Warehouse Lead | ☐ |
| Chốt quy trình bàn giao ĐVVC | Warehouse/Shipping Lead | ☐ |
| Chốt quy trình gia công ngoài | Production/Subcontract Lead | ☐ |
| Chốt rule QC hold/pass/fail | QC Lead | ☐ |
| Chốt training plan | PM/Support Lead | ☐ |

### 8.2. Technical readiness

| Checklist | Owner | Status |
|---|---|---|
| Dev/Staging/UAT/Prod environment sẵn sàng | DevOps | ☐ |
| Production database đã tạo | DevOps | ☐ |
| Backup policy đã test | DevOps | ☐ |
| Rollback procedure đã test | DevOps/Tech Lead | ☐ |
| Monitoring dashboard sẵn sàng | DevOps | ☐ |
| Log và audit log chạy đúng | Tech Lead | ☐ |
| Queue/outbox worker chạy ổn | Tech Lead | ☐ |
| OpenAPI contract freeze | Tech Lead/Frontend Lead | ☐ |
| Smoke test automation chạy được | QA Lead | ☐ |

### 8.3. Data readiness

| Checklist | Owner | Status |
|---|---|---|
| SKU master clean | Data Lead | ☐ |
| Material master clean | Data Lead | ☐ |
| Supplier master clean | Data Lead | ☐ |
| Customer master clean | Data Lead | ☐ |
| Warehouse/location clean | Warehouse/Data Lead | ☐ |
| Batch list clean | QC/Data Lead | ☐ |
| ĐVVC/carrier mapping clean | Shipping/Data Lead | ☐ |
| Opening stock draft ready | Warehouse/Data Lead | ☐ |
| Opening QC status ready | QC/Data Lead | ☐ |
| Opening AR/AP basic ready | Finance/Data Lead | ☐ |

---

## 9. Checklist T-14 đến T-8

### 9.1. Migration rehearsal

Mục tiêu: chạy thử migration bằng dữ liệu gần giống thật.

| Bước | Owner | Kết quả mong muốn |
|---|---|---|
| Import master data rehearsal | Data Lead | không lỗi trùng mã nghiêm trọng |
| Import warehouse/location | Data Lead/Warehouse | location map đúng |
| Import batch/expiry/QC status | Data Lead/QC | batch có hạn dùng và QC status |
| Import opening stock rehearsal | Data Lead/Warehouse | tồn khớp file đối soát |
| Import users/roles | Admin | user login đúng quyền |
| Import carrier/customer/supplier | Data Lead | mapping đủ |
| Run smoke tests | QA | pass luồng P0 |
| Reconcile migration output | Finance/Warehouse/QC | lệch trong ngưỡng cho phép |

### 9.2. Migration rehearsal acceptance

Migration rehearsal được coi là đạt nếu:

```text
Master data duplicate nghiêm trọng = 0
Stock opening lệch không giải thích được = 0
Batch thiếu hạn dùng = 0 với hàng cần hạn dùng
QC status unknown = 0 cho hàng kiểm soát QC
User critical không login được = 0
Smoke test P0 pass >= 95%
```

---

## 10. Checklist T-7 đến T-4

### 10.1. Data freeze mềm

Từ T-7:

- hạn chế tạo mới SKU
- hạn chế tạo mới NCC/khách nếu không bắt buộc
- mọi thay đổi master data phải ghi log
- tồn kho phải kiểm kê dần theo khu
- batch/hạn dùng phải đối soát

### 10.2. User readiness

| Checklist | Owner | Status |
|---|---|---|
| Warehouse user đã training | Support/Warehouse Lead | ☐ |
| QC user đã training | Support/QC Lead | ☐ |
| Sales user đã training | Support/Sales Lead | ☐ |
| Shipping user đã training | Support/Shipping Lead | ☐ |
| Returns user đã training | Support/Warehouse Lead | ☐ |
| Production/Subcontract user đã training | Support/Production Lead | ☐ |
| Finance user đã training | Support/Finance Lead | ☐ |
| Admin user đã training | Support/Admin | ☐ |
| SOP đã gửi cho user | PM | ☐ |
| Danh sách issue training đã xử lý | PM/Support | ☐ |

---

## 11. Checklist T-3 đến T-1

### 11.1. Final cutover preparation

| Việc | Owner | Deadline | Status |
|---|---|---:|---|
| Freeze code release | Tech Lead | T-3 | ☐ |
| Freeze OpenAPI contract | Tech Lead | T-3 | ☐ |
| Final regression test | QA Lead | T-3 | ☐ |
| Final database migration dry-run | Data Lead | T-2 | ☐ |
| Production backup baseline | DevOps | T-2 | ☐ |
| Final master data export | Data Lead | T-1 | ☐ |
| Final stock count | Warehouse Lead | T-1 | ☐ |
| Final batch/QC validation | QC Lead | T-1 | ☐ |
| Final user/role validation | Admin | T-1 | ☐ |
| War room link/channel ready | PM | T-1 | ☐ |
| Go/No-Go meeting | CEO/PM | T-1 | ☐ |

### 11.2. Final Go/No-Go criteria

Go-live chỉ được bật nếu:

```text
Không còn P0 bug
Không còn P1 bug ảnh hưởng luồng nhập/xuất/bàn giao/hàng hoàn
Opening stock đã ký xác nhận
Batch/QC status đã ký xác nhận
User/role đã ký xác nhận
UAT critical flow đã pass
Backup/rollback đã test
Warehouse Lead đồng ý
QC Lead đồng ý
Sales Lead đồng ý
Finance Lead đồng ý
Tech Lead đồng ý
CEO hoặc Executive Sponsor chốt GO
```

---

## 12. Cutover runbook — ngày T-Day

### 12.1. Nguyên tắc trong ngày go-live

Trong ngày T-Day:

- tất cả thay đổi phải ghi vào war room log
- mọi lỗi P0/P1 phải có owner
- mọi workaround phải được PM duyệt
- không deploy nóng nếu không có Tech Lead + PM duyệt
- không import lại dữ liệu nếu chưa backup
- không cho user tự ý tạo quy trình ngoài ERP nếu chưa ghi nhận

---

## 13. Lịch chạy T-Day gợi ý

### 13.1. 06:00–07:00 — Production readiness check

| Bước | Owner | Status |
|---|---|---|
| Kiểm tra production app online | DevOps | ☐ |
| Kiểm tra database online | DevOps | ☐ |
| Kiểm tra Redis/Queue online | DevOps | ☐ |
| Kiểm tra file storage online | DevOps | ☐ |
| Kiểm tra monitoring/logging | DevOps | ☐ |
| Kiểm tra backup gần nhất | DevOps | ☐ |

### 13.2. 07:00–08:00 — User login & permission smoke

| Bước | Owner | Status |
|---|---|---|
| Admin login | Admin | ☐ |
| Warehouse user login | Warehouse Lead | ☐ |
| QC user login | QC Lead | ☐ |
| Sales user login | Sales Lead | ☐ |
| Shipping user login | Shipping Lead | ☐ |
| Finance user login | Finance Lead | ☐ |
| Kiểm tra menu theo role | QA/Support | ☐ |
| Kiểm tra field-level permission | QA/Support | ☐ |

### 13.3. 08:00–09:00 — Master data validation

| Bước | Owner | Status |
|---|---|---|
| SKU list đúng | Data/Warehouse/Sales | ☐ |
| Material list đúng | Data/Purchasing | ☐ |
| Supplier list đúng | Data/Purchasing | ☐ |
| Customer list đúng | Data/Sales | ☐ |
| Warehouse/location đúng | Data/Warehouse | ☐ |
| Carrier list đúng | Data/Shipping | ☐ |
| Batch/expiry/QC status đúng | Data/QC | ☐ |

### 13.4. 09:00–10:00 — Opening stock validation

| Bước | Owner | Status |
|---|---|---|
| Tổng tồn theo SKU đúng | Warehouse/Data | ☐ |
| Tồn theo location đúng | Warehouse/Data | ☐ |
| Tồn theo batch đúng | Warehouse/QC | ☐ |
| Hàng hold QC không vào available | QC/Warehouse | ☐ |
| Hàng cận date hiển thị cảnh báo | Warehouse | ☐ |
| Tồn khả dụng đúng công thức | QA/Warehouse | ☐ |

### 13.5. 10:00–11:00 — Critical smoke test 1: inbound + QC

```text
Tạo phiếu nhận hàng
Ghi batch/hạn dùng
QC hold
QC pass
Chuyển sang tồn khả dụng
Kiểm tra stock ledger
Kiểm tra audit log
```

Owner: Warehouse + QC + QA + Tech Lead

### 13.6. 11:00–12:00 — Critical smoke test 2: sales order + reserve

```text
Tạo đơn hàng
Kiểm tra tồn khả dụng
Reserve stock
Không cho reserve quá tồn
Kiểm tra order status
Kiểm tra stock reservation
Kiểm tra audit log
```

Owner: Sales + Warehouse + QA + Tech Lead

### 13.7. 13:00–14:00 — Critical smoke test 3: pick/pack

```text
Tạo pick task
Nhân viên kho pick theo đơn
Kiểm tra SKU/batch
Đóng gói
Ghi packing status
Kiểm tra packed order list
```

Owner: Warehouse + QA + Frontend Lead

### 13.8. 14:00–15:00 — Critical smoke test 4: carrier manifest + scan handover

```text
Tạo manifest theo ĐVVC/chuyến
Đưa đơn packed vào manifest
Quét mã đơn/mã vận đơn
Nếu đủ → xác nhận handover
Nếu thiếu → hiển thị exception
Ghi scan event
Ghi audit log
```

Owner: Shipping + Warehouse + QA + Tech Lead

### 13.9. 15:00–16:00 — Critical smoke test 5: returns/hàng hoàn

```text
Nhận hàng từ shipper
Quét hàng hoàn
Đưa vào khu vực hàng hoàn
Kiểm tra tình trạng
Phân loại còn sử dụng / không sử dụng
Nếu còn dùng → nhập kho phù hợp
Nếu không dùng → chuyển lab/kho hỏng
Ghi return disposition
```

Owner: Warehouse + QC + QA

### 13.10. 16:00–17:00 — Critical smoke test 6: subcontract manufacturing

```text
Tạo đơn gia công ngoài
Chuyển NVL/bao bì cho nhà máy
Ghi biên bản bàn giao
Duyệt mẫu
Nhận hàng gia công về kho
QC nhận hàng
Nếu lỗi → tạo claim nhà máy
Nếu đạt → nhập kho
```

Owner: Production/Subcontract + Warehouse + QC + QA

### 13.11. 17:00–18:00 — End-of-day reconciliation test

```text
Warehouse daily board tổng hợp đơn
Đối soát số đơn nhận/đã soạn/đã đóng/đã bàn giao
Đối soát tồn kho phát sinh trong ngày
Ghi chênh lệch nếu có
Tạo shift closing record
Báo cáo quản lý
```

Owner: Warehouse Lead + Finance + QA

### 13.12. 18:00 — T-Day sign-off

| Vai trò | Ký xác nhận |
|---|---|
| Warehouse Lead | ☐ |
| QC Lead | ☐ |
| Sales Lead | ☐ |
| Shipping Lead | ☐ |
| Finance Lead | ☐ |
| Tech Lead | ☐ |
| PM | ☐ |
| CEO/Executive Sponsor | ☐ |

---

## 14. Smoke Test Checklist P0

### 14.1. System access

| Test | Expected |
|---|---|
| Login production | Thành công |
| Role menu | Hiển thị đúng |
| Không có quyền truy cập module cấm | Bị chặn |
| Sensitive action yêu cầu confirm | Có |
| Audit log ghi nhận | Có |

### 14.2. Inventory/WMS

| Test | Expected |
|---|---|
| Nhập kho tạo movement | Stock ledger có record |
| QC hold | Không vào available |
| QC pass | Vào available |
| Reserve stock | Available giảm đúng |
| Issue stock | Physical/available giảm đúng |
| Adjustment cần phê duyệt | Có |
| Không sửa stock ledger | Không cho sửa |

### 14.3. Sales + Shipping

| Test | Expected |
|---|---|
| Tạo đơn | Thành công |
| Reserve quá tồn | Bị chặn |
| Pick/pack đơn | Thành công |
| Đưa vào manifest | Thành công |
| Scan đúng đơn | Thành công |
| Scan sai đơn | Cảnh báo |
| Manifest thiếu đơn | Không cho close sạch |
| Handover thành công | Order/shipment status đúng |

### 14.4. Returns

| Test | Expected |
|---|---|
| Nhận hàng hoàn | Tạo return receiving |
| Quét mã vận đơn | Tìm đúng đơn |
| Phân loại còn dùng | Vào kho phù hợp sau QC/rule |
| Phân loại không dùng | Vào lab/kho hỏng |
| Không rõ tình trạng | Cần supervisor/QC quyết định |

### 14.5. Subcontract manufacturing

| Test | Expected |
|---|---|
| Tạo đơn gia công | Thành công |
| Chuyển NVL/bao bì | Có stock movement |
| Duyệt mẫu | Có trạng thái approved |
| Nhận hàng về kho | Có inbound receipt |
| QC fail | Không nhập available |
| Tạo claim nhà máy | Có claim record và deadline |

---

## 15. Rollback Plan

### 15.1. Khi nào rollback?

Rollback chỉ dùng khi có sự cố nghiêm trọng, không thể workaround an toàn trong thời gian ngắn.

Điều kiện rollback:

```text
Hệ thống production không thể truy cập > 60 phút trong giờ vận hành chính
Tồn kho bị sai hàng loạt không thể xác định nguồn
Stock ledger phát sinh lỗi transaction nghiêm trọng
Không thể tạo hoặc xử lý đơn hàng P0
Không thể bàn giao ĐVVC trên diện rộng
Không thể nhận hàng hoàn hoặc không truy được đơn
Sai quyền nghiêm trọng làm lộ giá vốn/lương/công nợ
Mất dữ liệu hoặc nghi ngờ mất dữ liệu
```

### 15.2. Ai được quyết định rollback?

Rollback cần đủ 3 người đồng ý:

```text
CEO/Executive Sponsor
Project Manager
Tech Lead/DevOps Lead
```

Nếu liên quan kho và bán hàng, phải tham khảo thêm:

```text
Warehouse Lead
Sales Lead
Finance Lead
```

### 15.3. Rollback steps

| Bước | Owner | Ghi chú |
|---|---|---|
| 1. Dừng user thao tác | PM/Support | thông báo war room |
| 2. Export dữ liệu phát sinh từ ERP | Data Lead | để không mất chứng từ đã tạo |
| 3. Backup production hiện tại | DevOps | snapshot DB + file |
| 4. Restore phiên bản trước nếu cần | DevOps | theo backup baseline |
| 5. Chuyển vận hành tạm sang quy trình dự phòng | Business Leads | Excel/giấy có kiểm soát |
| 6. Ghi lại toàn bộ giao dịch trong thời gian rollback | Business Leads | cutover replay |
| 7. Fix lỗi root cause | Tech Lead | không patch mù |
| 8. Re-run smoke test | QA | pass trước khi bật lại |
| 9. Reconcile giao dịch phát sinh | Data/Finance/Warehouse | nhập lại có kiểm soát |
| 10. Re-Go-Live decision | CEO/PM | chốt bật lại |

### 15.4. Không được rollback nếu

- chỉ là lỗi UI nhỏ
- chỉ 1 user không thao tác được do quyền
- một số field hiển thị sai nhưng dữ liệu core đúng
- report chậm nhưng vận hành vẫn chạy
- issue có workaround an toàn được PM duyệt

---

## 16. Fallback Manual Procedure

Nếu ERP tạm ngưng trong thời gian ngắn, dùng quy trình thủ công có kiểm soát.

### 16.1. Manual order processing

Bắt buộc ghi:

```text
Mã đơn
Thời gian tạo
Người tạo
SKU
Số lượng
Batch nếu có
Kênh bán
ĐVVC
Trạng thái thanh toán
Trạng thái giao hàng
```

### 16.2. Manual warehouse movement

Bắt buộc ghi:

```text
Mã phiếu tạm
Loại movement
SKU
Batch
Số lượng
Kho/location
Lý do
Người thực hiện
Người duyệt
Thời gian
```

### 16.3. Manual shipping handover

Bắt buộc ghi:

```text
Mã manifest tạm
ĐVVC
Danh sách đơn
Số kiện/thùng/rổ
Người bàn giao
Người nhận
Thời gian
Chữ ký/xác nhận
```

### 16.4. Manual returns

Bắt buộc ghi:

```text
Mã đơn/mã vận đơn
SKU
Số lượng
Tình trạng hàng
Còn sử dụng/không sử dụng/chưa rõ
Người kiểm tra
Ảnh nếu có
Quyết định xử lý
```

Sau khi ERP phục hồi, tất cả giao dịch thủ công phải được nhập lại theo batch replay, không nhập lẻ tùy hứng.

---

## 17. Hypercare Plan

### 17.1. Thời gian hypercare

```text
Hypercare cao điểm: T+1 đến T+3
Hypercare ổn định: T+4 đến T+14
Transition to BAU support: T+15
```

### 17.2. War room schedule

| Giai đoạn | Lịch trực |
|---|---|
| T-Day | Trực toàn thời gian |
| T+1 đến T+3 | 08:00–20:00 |
| T+4 đến T+7 | 08:30–18:30 |
| T+8 đến T+14 | giờ hành chính + hotline P0 |
| T+15 trở đi | support process bình thường |

### 17.3. Daily hypercare meeting

Mỗi ngày họp 2 lần trong 3 ngày đầu:

```text
08:30 — plan of day
17:30 — issue review & next action
```

Agenda:

```text
1. Tổng số issue
2. P0/P1 còn mở
3. Lỗi kho/giao hàng/hàng hoàn/QC
4. Data discrepancy
5. Workaround đang dùng
6. Quyết định cần CEO/PM
7. Training bổ sung
8. Fix release trong ngày
```

---

## 18. Issue Triage

### 18.1. Severity definition

| Severity | Định nghĩa | SLA phản hồi | SLA xử lý mục tiêu |
|---|---|---:|---:|
| P0 | Dừng vận hành chính, mất dữ liệu, sai tồn nghiêm trọng | 15 phút | 2–4 giờ |
| P1 | Ảnh hưởng luồng quan trọng nhưng có workaround | 30 phút | trong ngày |
| P2 | Lỗi chức năng vừa, không chặn vận hành | 4 giờ | 2–3 ngày |
| P3 | UI/text/report nhỏ | 1 ngày | sprint sau |
| CR | Yêu cầu thay đổi/phát sinh | theo review | theo backlog |

### 18.2. Issue categories

```text
ACCESS_PERMISSION
MASTER_DATA
INVENTORY_STOCK
BATCH_QC
SALES_ORDER
PICK_PACK
SHIPPING_HANDOVER
RETURNS
SUBCONTRACT
FINANCE_RECONCILIATION
REPORTING
UI_UX
PERFORMANCE
INTEGRATION
DATA_MIGRATION
```

### 18.3. Issue ticket minimum fields

```text
Ticket ID
Reported by
Role/user
Time
Module
Severity
Description
Steps to reproduce
Expected result
Actual result
Screenshot/video/log
Order/SKU/batch/manifest if applicable
Business impact
Workaround
Owner
ETA
Status
Root cause
Resolution
```

---

## 19. Hypercare KPI

### 19.1. Operational KPI

| KPI | Target T+1 đến T+3 | Target T+4 đến T+14 |
|---|---:|---:|
| User login success | >= 98% | >= 99% |
| Đơn tạo thành công | >= 95% | >= 99% |
| Pick/pack success | >= 95% | >= 98% |
| Handover scan success | >= 95% | >= 98% |
| Return receiving success | >= 90% | >= 97% |
| Stock discrepancy unresolved | 0 P0 | 0 P0/P1 |
| P0 open cuối ngày | 0 | 0 |
| P1 open quá 24h | <= 3 | <= 1 |

### 19.2. System KPI

| KPI | Target |
|---|---:|
| API P95 response time critical endpoints | < 800ms |
| Error rate critical endpoints | < 1% |
| Queue backlog critical jobs | không tồn quá 15 phút |
| Database migration issue | 0 P0 |
| Audit log missing critical action | 0 |
| Backup success | 100% mỗi ngày |

---

## 20. Communication Plan

### 20.1. Channel

| Channel | Mục đích |
|---|---|
| War room group | P0/P1, quyết định nhanh |
| Ticket system | ghi nhận issue chính thức |
| Email/Zalo broadcast | thông báo cho user |
| Daily report | gửi quản lý |
| Release note | thông báo fix/change |

### 20.2. Mẫu thông báo Go-Live

```text
Thông báo: ERP Phase 1 chính thức go-live

Từ [ngày/giờ], các nghiệp vụ sau sẽ thực hiện trên ERP:
- nhập/xuất kho
- QC hold/pass/fail
- đơn hàng
- soạn/đóng hàng
- bàn giao ĐVVC
- hàng hoàn
- gia công ngoài

Trong 2 tuần đầu, nếu gặp lỗi, vui lòng báo vào kênh [war room/support] theo format:
Module - Mã đơn/SKU/batch - Mô tả lỗi - Ảnh chụp màn hình.

Không tự ý xử lý ngoài hệ thống nếu chưa được trưởng bộ phận hoặc PM xác nhận.
```

### 20.3. Mẫu thông báo incident

```text
Incident: [Tên sự cố]
Thời gian phát hiện:
Ảnh hưởng:
Module:
Mức độ:
Workaround tạm thời:
Owner:
ETA cập nhật tiếp theo:
```

### 20.4. Mẫu daily hypercare report

```text
Ngày:
Tổng issue mới:
P0:
P1:
P2/P3:
Issue đã đóng:
Issue còn mở:
Ảnh hưởng vận hành:
Fix release hôm nay:
Training bổ sung cần làm:
Quyết định cần quản lý:
```

---

## 21. Production Release Control trong Hypercare

Trong hypercare, release phải kiểm soát chặt.

### 21.1. Hotfix allowed

Chỉ hotfix khi:

```text
P0/P1
Có root cause rõ
Có test case tái hiện
Có rollback plan
Tech Lead + PM duyệt
```

### 21.2. Không hotfix cho

```text
UI nhỏ
Yêu cầu đổi text không ảnh hưởng vận hành
Report nice-to-have
Feature mới chưa nằm trong scope
Yêu cầu thay đổi quy trình chưa qua decision log
```

### 21.3. Hotfix checklist

| Checklist | Status |
|---|---|
| Root cause xác định | ☐ |
| Test case tái hiện | ☐ |
| Fix code review xong | ☐ |
| Unit/integration test pass | ☐ |
| Smoke test pass | ☐ |
| Backup/snapshot trước deploy | ☐ |
| Rollback plan ready | ☐ |
| PM duyệt | ☐ |
| Business owner liên quan được báo | ☐ |

---

## 22. Data Reconciliation sau Go-Live

### 22.1. Đối soát hằng ngày T+1 đến T+14

| Hạng mục | Owner | Tần suất |
|---|---|---|
| Tồn kho theo SKU | Warehouse | hằng ngày |
| Tồn kho theo batch | Warehouse/QC | hằng ngày |
| Đơn đã tạo/đã đóng/đã bàn giao | Sales/Warehouse | hằng ngày |
| Manifest ĐVVC | Shipping | hằng ngày |
| Hàng hoàn | Warehouse/QC | hằng ngày |
| Stock adjustment | Warehouse/Finance | hằng ngày |
| Đơn gia công ngoài | Production | 2–3 ngày/lần |
| Công nợ/COD cơ bản | Finance | hằng ngày/tuần |

### 22.2. Chênh lệch phải ghi nhận

Mỗi discrepancy phải có:

```text
Mã issue
Loại chênh lệch
SKU/batch/order/manifest
Số liệu ERP
Số liệu đối chiếu
Nguồn đối chiếu
Nguyên nhân nghi ngờ
Owner
Hành động xử lý
Trạng thái
```

---

## 23. Warehouse Go-Live Special Control

Vì kho là trái tim Phase 1, trong 7 ngày đầu phải áp dụng kiểm soát đặc biệt.

### 23.1. Đầu ngày

```text
Kiểm tra danh sách đơn mới
Kiểm tra đơn chưa pick/pack
Kiểm tra đơn đã packed chưa handover
Kiểm tra hàng hoàn chưa xử lý
Kiểm tra batch hold/cận date
Kiểm tra issue tồn kho hôm qua
```

### 23.2. Trong ngày

```text
Không pick hàng ngoài task
Không handover đơn chưa packed
Không bỏ qua scan nếu manifest yêu cầu scan
Không nhập hàng hoàn nếu chưa kiểm tình trạng
Không điều chỉnh tồn nếu chưa có phê duyệt
```

### 23.3. Cuối ngày

```text
Đối soát số đơn nhận
Đối soát số đơn đã soạn
Đối soát số đơn đã đóng
Đối soát số đơn đã bàn giao
Đối soát hàng hoàn
Đối soát tồn phát sinh
Tạo shift closing record
Gửi báo cáo quản lý
```

---

## 24. QC/Batch Go-Live Special Control

Trong 14 ngày đầu:

```text
Batch thiếu hạn dùng không được release
QC hold không được bán
QC fail không được nhập available
QC pass phải có user/time/audit
Batch có complaint phải được trace ngay
Hàng gia công ngoài phải qua QC trước khi nhập khả dụng
```

---

## 25. Shipping/Handover Go-Live Special Control

Trong 7 ngày đầu:

```text
Mọi manifest phải có ĐVVC/chuyến/người bàn giao
Đơn chưa packed không được vào manifest
Scan sai carrier phải cảnh báo
Scan thiếu đơn phải tạo exception
Manifest close phải lưu số đơn/số kiện/người xác nhận
Nếu thiếu đơn, phải kiểm tra lại mã và khu vực đóng hàng trước khi xác nhận
```

---

## 26. Returns Go-Live Special Control

Trong 14 ngày đầu:

```text
Hàng hoàn nhận từ shipper phải vào khu hàng hoàn
Phải quét mã đơn/mã vận đơn
Phải kiểm tình trạng bên trong
Phải phân loại còn dùng/không dùng/chưa rõ
Hàng còn dùng phải theo rule nhập lại
Hàng không dùng phải chuyển lab/kho hỏng
Mọi quyết định disposition phải có người xác nhận
```

---

## 27. Subcontract Manufacturing Go-Live Special Control

Với gia công ngoài:

```text
Đơn gia công phải có số lượng/quy cách/mẫu mã
Chuyển NVL/bao bì phải có biên bản bàn giao
Duyệt mẫu phải ghi trạng thái và lưu bằng chứng
Nhận hàng phải kiểm số lượng/chất lượng
Hàng lỗi phải tạo claim nhà máy
Claim nhà máy cần deadline 3–7 ngày theo quy trình hiện tại
Thanh toán cuối chỉ nên mở khi nghiệm thu đạt hoặc có quyết định xử lý
```

---

## 28. Sign-off sau Hypercare

Kết thúc hypercare khi đạt:

```text
Không còn P0 trong 7 ngày liên tiếp
P1 còn mở <= 2 và có workaround
User vận hành chính không cần support liên tục
Kho đối soát cuối ngày ổn định
Bàn giao ĐVVC không còn lỗi lặp lại nghiêm trọng
Hàng hoàn được xử lý đúng trạng thái
QC/batch không phát sinh lỗi quyền hoặc trạng thái
Data reconciliation không còn chênh lệch nghiêm trọng
Support model BAU đã nhận bàn giao
```

### 28.1. Hypercare exit sign-off

| Vai trò | Ký xác nhận |
|---|---|
| CEO/Executive Sponsor | ☐ |
| PM | ☐ |
| Product Owner | ☐ |
| Warehouse Lead | ☐ |
| QC Lead | ☐ |
| Sales Lead | ☐ |
| Shipping Lead | ☐ |
| Finance Lead | ☐ |
| Tech Lead | ☐ |
| Support Lead | ☐ |

---

## 29. Bàn giao sang BAU Support

Sau hypercare, chuyển sang mô hình hỗ trợ thường xuyên.

### 29.1. Tài liệu bàn giao

| Tài liệu | Trạng thái |
|---|---|
| SOP Training Manual | ☐ |
| User list & role list | ☐ |
| Known issue list | ☐ |
| Open enhancement backlog | ☐ |
| Support escalation matrix | ☐ |
| Admin guide | ☐ |
| Backup/restore guide | ☐ |
| Release process | ☐ |
| Data governance rule | ☐ |

### 29.2. Support categories

```text
Bug
Data issue
Permission issue
Training issue
Change request
Enhancement
Integration issue
Performance issue
Security incident
```

---

## 30. Risk Register trong Go-Live

| Rủi ro | Mức độ | Dấu hiệu | Biện pháp |
|---|---|---|---|
| Opening stock sai | Cao | lệch tồn, không reserve được | kiểm kê, reconcile trước T-Day |
| User chưa quen scan | Trung bình/Cao | scan sai, bỏ scan | training + support tại kho |
| Manifest thiếu đơn | Cao | ĐVVC nhận thiếu | exception flow + double check |
| Hàng hoàn nhập sai trạng thái | Cao | hàng lỗi quay lại bán | QC/warehouse disposition rule |
| Gia công ngoài chưa đủ chứng từ | Trung bình/Cao | nhận hàng không đối chiếu được | bắt buộc biên bản/bằng chứng |
| Quyền quá rộng | Cao | user thấy giá vốn/lương/công nợ | RBAC + field permission |
| Report chậm | Trung bình | user than dashboard chậm | tách job async/cache |
| Hotfix thiếu test | Cao | fix xong vỡ chỗ khác | hotfix checklist bắt buộc |
| Người dùng quay lại Excel | Cao | dữ liệu lệch | quản lý bắt buộc dùng ERP |

---

## 31. Appendix A — Go/No-Go Form

```text
Dự án:
Ngày đánh giá:
Môi trường:
Phiên bản build:
Migration version:

Business readiness:
[ ] Đạt
[ ] Không đạt

Data readiness:
[ ] Đạt
[ ] Không đạt

Technical readiness:
[ ] Đạt
[ ] Không đạt

UAT/regression:
[ ] Đạt
[ ] Không đạt

Training:
[ ] Đạt
[ ] Không đạt

Rollback plan:
[ ] Đạt
[ ] Không đạt

Quyết định:
[ ] GO
[ ] NO-GO
[ ] GO có điều kiện

Điều kiện nếu có:

Người phê duyệt:
CEO:
PM:
Tech Lead:
Business Owner:
```

---

## 32. Appendix B — Cutover Log Template

```text
Time:
Action:
Owner:
Expected:
Actual:
Status:
Issue ID:
Decision:
Next action:
```

---

## 33. Appendix C — Hypercare Issue Log Template

```text
Issue ID:
Date/time:
Reported by:
Module:
Severity:
Description:
Business impact:
Evidence:
Owner:
Status:
Root cause:
Workaround:
Resolution:
Closed by:
Closed time:
```

---

## 34. Appendix D — Daily Warehouse Reconciliation Template

```text
Ngày:
Ca:
Người phụ trách:

Đơn nhận trong ngày:
Đơn đã pick:
Đơn đã pack:
Đơn đã handover:
Đơn chưa handover:
Hàng hoàn nhận:
Hàng hoàn đã xử lý:
Hàng hoàn chưa xử lý:

Stock movement count:
Adjustment count:
Chênh lệch phát hiện:

Batch hold:
Batch pass:
Batch fail:
Cận date alert:

Ghi chú:
Người xác nhận:
```

---

## 35. Appendix E — Manifest Handover Template

```text
Manifest ID:
ĐVVC:
Chuyến:
Ngày/giờ:
Khu vực để hàng:
Số thùng/rổ:
Số đơn dự kiến:
Số đơn đã scan:
Số đơn thiếu:
Danh sách đơn thiếu:
Người bàn giao:
Người nhận:
Ghi chú:
Xác nhận:
```

---

## 36. Appendix F — Returns Receiving Template

```text
Return ID:
Mã đơn/mã vận đơn:
Ngày nhận:
Shipper/ĐVVC:
SKU:
Batch nếu xác định được:
Số lượng:
Tình trạng bao bì:
Tình trạng sản phẩm:
Phân loại:
[ ] Còn sử dụng
[ ] Không sử dụng
[ ] Chưa rõ cần QC/supervisor

Hành động:
[ ] Nhập lại kho
[ ] Chuyển lab
[ ] Chuyển kho hỏng
[ ] Chờ quyết định

Người kiểm:
Người xác nhận:
Ảnh/file:
```

---

## 37. Appendix G — Subcontract Receiving / Claim Template

```text
Subcontract Order ID:
Nhà máy:
SKU:
Batch:
Số lượng đặt:
Số lượng nhận:
Ngày nhận:
Chứng từ kèm theo:
COA/MSDS nếu có:
Mẫu đã duyệt:
QC result:
[ ] Pass
[ ] Hold
[ ] Fail

Nếu lỗi:
Loại lỗi:
Số lượng lỗi:
Bằng chứng:
Ngày gửi claim:
Deadline phản hồi 3–7 ngày:
Người phụ trách:
Trạng thái claim:
```

---

## 38. Definition of Go-Live Done

Go-live được coi là hoàn tất khi:

```text
ERP production chạy ổn định
Luồng nhập/xuất/kho/QC/sales/shipping/returns/subcontract hoạt động
Opening stock được đối soát
User chính vận hành được
Không có P0 mở
Audit log ghi nhận action quan trọng
Backup/monitoring hoạt động
Hypercare plan đang chạy
Daily reconciliation được thực hiện
```

---

## 39. Câu chốt vận hành

ERP không chết vì thiếu màn hình. ERP chết vì **go-live không có kỷ luật**.

Trong 14 ngày đầu, phải nhớ:

```text
Không bypass quy trình.
Không sửa dữ liệu ngoài hệ thống.
Không để issue trôi miệng.
Không deploy hotfix thiếu test.
Không để Excel quay lại làm nguồn sự thật.
```

Go-live tốt là khi hệ thống mới không chỉ chạy được, mà làm cho mọi người **bớt đoán, bớt nhớ miệng, bớt sửa tay, và nhìn được sự thật vận hành mỗi ngày**.

