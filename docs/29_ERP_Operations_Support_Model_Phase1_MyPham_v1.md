# 29_ERP_Operations_Support_Model_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Phase:** Phase 1  
**Phiên bản:** v1.0  
**Mục đích:** Xác định mô hình hỗ trợ vận hành sau go-live, trách nhiệm xử lý lỗi, kênh tiếp nhận yêu cầu, SLA, escalation, change request, training và quản trị hệ thống sau triển khai.

---

## 1. Tư duy nền tảng

ERP sau khi go-live không còn là “dự án phần mềm”. Nó trở thành **hạ tầng vận hành sống** của công ty.

Nếu không có mô hình support rõ, hệ thống sẽ nhanh chóng rơi vào các lỗi quen thuộc:

- Người dùng hỏi trực tiếp dev qua Zalo, không có ticket.
- Lỗi tồn kho bị sửa tay, không có dấu vết.
- Kho báo một kiểu, sale báo một kiểu, kế toán báo một kiểu.
- Bug, change request, training issue bị trộn lẫn.
- Dev bị kéo vào xử lý nghiệp vụ thay vì xử lý kỹ thuật.
- Super user không có quyền quyết, admin hệ thống không có quy trình.
- Lỗi nhỏ tích tụ thành lỗi lớn trước khi ban lãnh đạo nhìn thấy.

Mô hình support đúng phải giúp công ty đạt 5 mục tiêu:

```text
1. Người dùng biết hỏi ai, hỏi qua kênh nào.
2. Lỗi được phân loại đúng: bug, dữ liệu, thao tác sai, cấu hình sai, hay yêu cầu cải tiến.
3. Sự cố nghiệp vụ được xử lý nhanh mà không phá dữ liệu.
4. Dev chỉ nhận việc đã được lọc, mô tả rõ, tái hiện được.
5. Mọi thay đổi sau go-live đều có log, có duyệt, có kiểm soát.
```

Một câu nhớ cho toàn đội:

> ERP không chỉ cần build đúng. ERP cần được nuôi đúng sau khi chạy thật.

---

## 2. Phạm vi áp dụng

Tài liệu này áp dụng cho toàn bộ vận hành Phase 1, gồm:

- Master Data
- Purchasing / Mua hàng
- Receiving / Nhập kho
- QA/QC
- Inventory / Kho
- Warehouse Daily Operation
- Pick / Pack / Shipping
- Carrier Handover / Bàn giao ĐVVC
- Returns / Hàng hoàn
- Subcontract Manufacturing / Gia công ngoài
- Sales Order cơ bản
- Finance đối soát cơ bản
- User access / phân quyền
- Audit log / dữ liệu nhạy cảm
- Integration và lỗi kỹ thuật liên quan API, database, worker, sync

Tài liệu này không thay thế SOP thao tác từng nghiệp vụ. Nó là **mô hình hỗ trợ vận hành**: ai xử lý, xử lý trong bao lâu, escalate thế nào, dữ liệu được chỉnh ra sao, và làm sao để hệ thống sống ổn sau go-live.

---

## 3. Nguồn workflow thực tế đã đưa vào support model

Mô hình support này bám theo 4 nhóm workflow thực tế đã được cung cấp:

### 3.1. Công việc hằng ngày của kho

Workflow hiện tại thể hiện nhịp vận hành:

```text
Tiếp nhận đơn hàng trong ngày
→ Thực hiện quy trình xuất/nhập dựa trên bảng nội quy
→ Soạn hàng và đóng gói
→ Sắp xếp, tối ưu vị trí kho
→ Kiểm kê hàng tồn kho cuối ngày
→ Đối soát số liệu và báo cáo cho quản lý
→ Kết thúc ca làm
```

Hệ quả cho support:

- Support phải có khung xử lý lỗi theo ca/ngày.
- Các lỗi cuối ca phải có cutoff time.
- Lỗi tồn kho cuối ngày phải được xử lý theo quy trình inventory incident, không sửa tay.
- Cần super user kho chịu trách nhiệm lọc lỗi trước khi đẩy cho tech.

### 3.2. Nội quy nhập kho, xuất kho, đóng hàng, hàng hoàn

Workflow thực tế chia rõ 4 nhánh:

```text
Nhập kho:
Nhận chứng từ giao hàng
→ Kiểm tra số lượng / bao bì / lô
→ Đạt thì xếp kho, kiểm tra lần cuối, ký xác nhận
→ Không đạt thì trả NCC

Xuất kho:
Làm phiếu xuất kho
→ Xuất kho hàng
→ Kiểm tra số lượng, đối chiếu vị trí thực tế
→ Ký biên bản bàn giao
→ Trưởng kho lưu phiếu xuất

Đóng hàng:
Nhận phiếu đơn hàng hợp lệ
→ Lọc/phân loại theo ĐVVC, đơn lẻ, đơn sàn
→ Soạn theo từng đơn
→ Đóng gói, kiểm tra tại khu vực đóng hàng
→ Đếm lại tổng số đơn cuối mỗi sàn
→ Chuyển đến khu vực bàn giao cho ĐVVC
→ Lập/ký xác nhận với ĐVVC

Hàng hoàn:
Nhận hàng từ shipper
→ Đưa vào khu vực hàng hoàn
→ Quét hàng hoàn
→ Ghi nhận tình trạng thực tế
→ Kiểm tra chi tiết bên trong
→ Còn sử dụng thì chuyển kho
→ Không sử dụng thì chuyển lab/kho lỗi
→ Lập phiếu nhập kho
```

Hệ quả cho support:

- Nhập kho, xuất kho, đóng hàng, hàng hoàn phải có owner riêng.
- Lỗi phát sinh ở nhánh nào phải về đúng owner nhánh đó.
- Support không được xử lý chung chung kiểu “lỗi kho”. Phải phân loại: receiving, picking, packing, handover, return, stock count.

### 3.3. Quy trình bàn giao hàng cho ĐVVC

Workflow thực tế:

```text
Phân chia khu vực để hàng
→ Để theo thùng/rổ, mỗi thùng/rổ có số lượng bằng nhau
→ Đối chiếu số lượng đơn dựa trên bảng
→ Lấy hàng và quét mã trực tiếp
→ Nếu đủ đơn: ký xác nhận, bàn giao ĐVVC
→ Nếu chưa đủ: kiểm tra lại mã
   - Nếu mã chưa có trên hệ thống: đóng lại
   - Nếu đã có: tìm lại trong khu vực đóng hàng
```

Hệ quả cho support:

- Lỗi bàn giao phải có checklist riêng: thiếu đơn, quét sai mã, mã không tồn tại, đơn sai trạng thái, manifest lệch, carrier không xác nhận.
- Support phải phân biệt lỗi vận hành và lỗi hệ thống.
- Cần có quy trình “freeze handover batch” khi lệch số lượng.

### 3.4. Quy trình sản xuất/gia công ngoài

Workflow thực tế:

```text
Lên đơn hàng với nhà máy
→ Xác nhận số lượng, quy cách, mẫu mã
→ Cọc đơn hàng, xác nhận thời gian sản xuất/nhận hàng
→ Gia công
→ Chuyển kho nguyên vật liệu/bao bì từ công ty qua nhà máy sản xuất
→ Ký biên bản bàn giao, cung cấp giấy tờ COA/MSDS/term/phụ/hóa đơn VAT nếu cần
→ Làm mẫu, chốt mẫu
→ Sản xuất hàng loạt, kiểm nghiệm sản phẩm
→ Lưu mẫu
→ Giao hàng về kho
→ Kiểm tra số lượng và chất lượng hàng hóa
→ Hàng được nhận: nhập kho
→ Hàng không nhận: báo lại nhà máy về vấn đề hàng hóa trong vòng 3-7 ngày
→ Thanh toán lần cuối
```

Hệ quả cho support:

- Phải có support line riêng cho subcontract manufacturing.
- Lỗi chuyển NVL/bao bì cho nhà máy khác với lỗi nhập kho thành phẩm.
- Claim nhà máy trong 3-7 ngày là SLA nghiệp vụ, phải có cảnh báo và owner.
- Không xử lý lỗi gia công ngoài bằng cách sửa phiếu kho thủ công.

---

## 4. Nguyên tắc support bắt buộc

### 4.1. Ticket first

Mọi lỗi, yêu cầu hỗ trợ, yêu cầu sửa dữ liệu, yêu cầu cấu hình, yêu cầu phân quyền phải có ticket.

Không chấp nhận mô hình:

```text
Nhắn riêng dev → dev sửa trực tiếp → không có log
```

Kênh chat có thể dùng để báo khẩn, nhưng sau đó vẫn phải tạo ticket.

### 4.2. Không sửa dữ liệu trực tiếp trong database

Các dữ liệu sau tuyệt đối không sửa tay trong DB nếu không qua procedure được duyệt:

- Stock ledger
- Stock balance
- Batch status
- QC status
- Sales order status
- Shipment status
- Return disposition
- Payment/COD status
- Supplier/factory claim status
- User role/permission quan trọng

Nếu cần chỉnh dữ liệu, phải theo **Data Correction Request**.

### 4.3. Lỗi nghiệp vụ xử lý bởi business trước

Không phải lỗi nào cũng đưa cho dev.

Ví dụ:

```text
Nhân viên kho scan nhầm đơn → Super user kho xử lý/training lại.
Đơn thiếu hàng vì chưa pick đủ → Warehouse lead xử lý.
Mã SKU bị tạo sai → Master data owner xử lý.
API carrier lỗi → Tech support/integration xử lý.
```

### 4.4. Không dùng “urgent” bừa bãi

Severity phải dựa trên ảnh hưởng thật đến vận hành:

- Có chặn bán hàng không?
- Có chặn xuất kho không?
- Có làm sai tồn kho không?
- Có ảnh hưởng batch/QC/hạn dùng không?
- Có ảnh hưởng tiền/công nợ không?
- Có rủi ro bảo mật không?

### 4.5. Mọi thay đổi phải qua Change Control

Sau go-live, không sửa flow, sửa permission, sửa rule, sửa API, sửa report trực tiếp theo cảm hứng.

Mọi thay đổi đi qua:

```text
Change Request
→ Impact analysis
→ Business approval
→ Tech estimate
→ UAT nếu cần
→ Release plan
→ Release note
```

### 4.6. Super user là tuyến đầu

Mỗi bộ phận phải có super user:

- Kho
- QA/QC
- Sales/Admin đơn hàng
- Purchasing
- Finance
- Subcontract/Production coordinator
- ERP Admin

Super user có nhiệm vụ:

- Lọc lỗi thao tác với lỗi hệ thống
- Hướng dẫn người dùng
- Gom issue trùng
- Xác nhận business impact
- Tham gia UAT hotfix/change request

---

## 5. Cấu trúc support 4 lớp

### 5.1. L0 — Self-service / SOP / Knowledge Base

Người dùng tự tra:

- SOP thao tác
- FAQ
- video training nếu có
- lỗi thường gặp
- checklist theo nghiệp vụ

Ví dụ L0:

```text
Không biết nhập hàng hoàn vào đâu → xem SOP Returns.
Không biết trạng thái batch Hold nghĩa là gì → xem Knowledge Base QC.
Không biết vì sao đơn không bàn giao được → xem FAQ Shipping Handover.
```

### 5.2. L1 — Super User / Business Support

Người xử lý đầu tiên trong bộ phận.

Nhiệm vụ:

- Nhận issue từ user
- Kiểm tra thao tác có đúng SOP không
- Kiểm tra dữ liệu cơ bản
- Tái hiện lỗi nếu có thể
- Phân loại issue
- Tạo ticket chuẩn nếu cần escalate

Ví dụ:

```text
Kho báo không scan được đơn.
L1 kiểm tra:
- đơn đã packed chưa?
- đơn có trong manifest không?
- mã vận đơn có đúng không?
- user có quyền scan không?
- scanner có hoạt động không?
```

### 5.3. L2 — ERP Admin / Business Application Support

L2 là nhóm support hệ thống nghiệp vụ.

Nhiệm vụ:

- Kiểm tra cấu hình
- Kiểm tra permission
- Kiểm tra workflow status
- Kiểm tra master data
- Kiểm tra audit log
- Xử lý các data correction được duyệt
- Escalate bug kỹ thuật cho L3

Ví dụ:

```text
Sales không thấy tồn khả dụng.
L2 kiểm tra:
- SKU active chưa?
- batch QC pass chưa?
- stock có bị reserve không?
- kho có được phân quyền cho user không?
```

### 5.4. L3 — Tech Team / Vendor / DevOps

L3 xử lý lỗi kỹ thuật.

Nhiệm vụ:

- Bug backend/frontend
- API/integration issue
- database migration issue
- worker/queue issue
- performance issue
- deployment/rollback
- security incident kỹ thuật
- hotfix

L3 chỉ nhận ticket khi đã có:

```text
- mô tả lỗi rõ
- bước tái hiện
- user liên quan
- màn hình/API liên quan
- thời điểm xảy ra
- dữ liệu/chứng từ liên quan
- ảnh/chụp màn hình/log nếu có
- phân loại severity từ L1/L2
```

---

## 6. Vai trò và trách nhiệm

## 6.1. System Owner

Thường là CEO/COO hoặc người được ủy quyền.

Trách nhiệm:

- Quyết định ưu tiên lớn
- Duyệt change request ảnh hưởng nhiều phòng ban
- Duyệt rollback trong sự cố nghiêm trọng
- Chốt go/no-go các release lớn

## 6.2. ERP Product Owner

Trách nhiệm:

- Quản backlog sau go-live
- Ưu tiên bug/enhancement
- Điều phối business và tech
- Bảo vệ scope Phase 1
- Đảm bảo các file tài liệu lõi được update khi có thay đổi

## 6.3. ERP Admin

Trách nhiệm:

- Quản user/role/permission
- Cấu hình hệ thống
- Theo dõi audit log
- Tạo báo cáo support định kỳ
- Quản Knowledge Base
- Điều phối ticket L1/L2

ERP Admin không được tự ý sửa dữ liệu nghiệp vụ nếu không có request/dụyệt.

## 6.4. Business Process Owner

Mỗi quy trình phải có owner:

| Quy trình | Owner chính |
|---|---|
| Master Data | ERP Admin + từng phòng ban liên quan |
| Nhập kho | Warehouse Lead |
| QC đầu vào/thành phẩm | QA/QC Lead |
| Xuất kho | Warehouse Lead |
| Đóng hàng | Packing Lead |
| Bàn giao ĐVVC | Shipping Lead |
| Hàng hoàn | Returns Lead / Warehouse Lead |
| Gia công ngoài | Production/Subcontract Coordinator |
| Mua hàng | Purchasing Lead |
| Sales Order | Sales Admin Lead |
| Công nợ/đối soát | Finance Lead |

## 6.5. Super User

Trách nhiệm:

- Là người dùng giỏi nhất của bộ phận
- Hỗ trợ đồng đội
- Tạo ticket đúng format
- Xác nhận lỗi có thật
- Tham gia UAT hotfix/change
- Đào tạo user mới trong bộ phận

## 6.6. Tech Lead

Trách nhiệm:

- Phân loại bug kỹ thuật
- Quyết định phương án fix
- Kiểm tra impact code/database/API
- Phê duyệt hotfix kỹ thuật
- Đảm bảo coding standard và module boundary

## 6.7. DevOps

Trách nhiệm:

- CI/CD
- deployment
- backup/restore
- monitoring
- rollback
- system uptime
- log infrastructure

## 6.8. QA/Test Lead

Trách nhiệm:

- Xác nhận bug fix
- Regression test
- Smoke test trước release
- Kiểm tra UAT với business
- Cập nhật test scenario khi workflow thay đổi

## 6.9. Security Owner

Có thể là ERP Admin + Tech Lead giai đoạn đầu.

Trách nhiệm:

- Phân quyền nhạy cảm
- Break-glass access
- Security incident
- Audit review
- Export control
- Access review định kỳ

---

## 7. Kênh tiếp nhận support

## 7.1. Kênh chính thức

Khuyến nghị dùng một trong các công cụ:

- Jira Service Management
- Linear
- ClickUp
- Notion ticket database
- Freshdesk
- Odoo Helpdesk nếu dùng hệ sinh thái đó
- Hoặc một form nội bộ có trạng thái rõ

Mỗi ticket phải có mã.

Ví dụ:

```text
ERP-INC-2026-0001
ERP-REQ-2026-0002
ERP-CHG-2026-0003
```

## 7.2. Kênh chat

Zalo/Slack/Teams chỉ dùng để:

- báo khẩn P0/P1
- thông báo war room
- nhắc ticket cần xử lý
- broadcast release note

Không dùng chat làm nơi phê duyệt sửa dữ liệu chính thức.

## 7.3. Kênh khẩn cấp

Với P0/P1:

```text
1. Gọi điện/chát khẩn cho ERP Admin hoặc Incident Commander
2. Tạo ticket trong vòng 15 phút
3. Mở war room nếu cần
4. Log toàn bộ quyết định trong ticket
```

---

## 8. Phân loại ticket

### 8.1. Incident

Lỗi/sự cố đang ảnh hưởng vận hành.

Ví dụ:

- Không tạo được phiếu xuất
- Không scan bàn giao ĐVVC được
- Batch QC pass nhưng không bán được
- Hệ thống tính sai tồn khả dụng
- Không nhận được đơn từ website

### 8.2. Service Request

Yêu cầu hỗ trợ bình thường.

Ví dụ:

- Tạo user mới
- Reset mật khẩu
- Cấp quyền kho
- Hỏi cách xử lý hàng hoàn
- Hỏi cách xuất report

### 8.3. Data Correction Request

Yêu cầu chỉnh dữ liệu.

Ví dụ:

- Nhập sai hạn dùng
- Gắn sai batch
- Sai tình trạng hàng hoàn
- Sai vendor/factory reference
- Sai mapping carrier code

Data correction luôn cần approval.

### 8.4. Change Request

Yêu cầu thay đổi hệ thống.

Ví dụ:

- Thêm trạng thái mới cho hàng hoàn
- Thêm report mới
- Đổi rule duyệt PO
- Đổi logic manifest
- Thêm integration với ĐVVC mới

### 8.5. Bug

Lỗi phần mềm có thể tái hiện.

Ví dụ:

- API trả sai lỗi
- màn hình không lưu dữ liệu
- scan đúng mã nhưng hệ thống báo không tồn tại
- filter báo cáo sai

### 8.6. Training Issue

Người dùng thao tác sai hoặc chưa hiểu SOP.

Ví dụ:

- Không biết batch Hold thì không xuất được
- Không biết hàng hoàn phải quét trước khi phân loại
- Không biết đơn chưa Packed thì không bàn giao được

---

## 9. Severity và SLA

## 9.1. Severity definition

| Severity | Định nghĩa | Ví dụ |
|---|---|---|
| P0 - Critical | Dừng vận hành lõi, rủi ro dữ liệu/tiền/hàng nghiêm trọng | Không xuất kho toàn hệ thống, stock ledger sai hàng loạt, mất quyền kiểm soát admin |
| P1 - High | Ảnh hưởng một quy trình quan trọng, có workaround hạn chế | Không bàn giao được một carrier, QC không release được batch, hàng hoàn không ghi nhận được |
| P2 - Medium | Ảnh hưởng một nhóm user/module, có workaround tạm | Một report sai filter, một màn hình chậm, một user không thấy menu |
| P3 - Low | Lỗi nhỏ, không ảnh hưởng vận hành chính | typo, layout, export thiếu cột phụ |
| P4 - Request | Câu hỏi, training, yêu cầu cải tiến | thêm báo cáo, hỏi cách thao tác |

## 9.2. SLA đề xuất

| Severity | Thời gian phản hồi | Thời gian workaround | Thời gian xử lý mục tiêu |
|---|---:|---:|---:|
| P0 | 15 phút | 1 giờ | 4-8 giờ hoặc hotfix khẩn |
| P1 | 30 phút | 4 giờ | 1 ngày làm việc |
| P2 | 4 giờ | 1-2 ngày | 3-5 ngày làm việc |
| P3 | 1 ngày | Không bắt buộc | Theo sprint/release |
| P4 | 1-2 ngày | Không bắt buộc | Theo backlog |

## 9.3. SLA riêng theo nghiệp vụ nhạy cảm

| Nghiệp vụ | SLA phản hồi | Ghi chú |
|---|---:|---|
| Lệch tồn cuối ngày | Trong cùng ngày | Không đóng ca nếu chưa có ghi chú xử lý |
| Batch QC sai trạng thái | 30 phút | Phải freeze bán/xuất nếu nghiêm trọng |
| Không bàn giao được ĐVVC | 30 phút | Ảnh hưởng giao hàng trong ngày |
| Hàng hoàn không phân loại được | 4 giờ | Tránh tồn khu hoàn quá lâu |
| Lỗi claim nhà máy 3-7 ngày | Trong ngày | Không để quá hạn phản hồi nhà máy |
| User mất quyền sau go-live | 1 giờ | Nếu ảnh hưởng vận hành chính |

---

## 10. Lifecycle xử lý incident

```text
1. Detect / phát hiện
2. Log ticket
3. Triage severity
4. Assign owner
5. Contain / khoanh vùng
6. Workaround nếu cần
7. Root cause analysis
8. Fix / correction / training
9. Validate với business
10. Close ticket
11. Postmortem nếu P0/P1
```

## 10.1. Detect

Nguồn phát hiện:

- User báo
- Super user phát hiện
- Dashboard cảnh báo
- Monitoring kỹ thuật
- Báo cáo cuối ngày
- Audit log bất thường
- Finance đối soát thấy lệch

## 10.2. Triage

Câu hỏi triage bắt buộc:

```text
1. Lỗi xảy ra ở module nào?
2. Có chặn vận hành không?
3. Có ảnh hưởng tồn kho/tiền/QC/batch không?
4. Có bao nhiêu user bị ảnh hưởng?
5. Có workaround không?
6. Lỗi là thao tác, dữ liệu, cấu hình, integration hay bug?
7. Có cần freeze quy trình không?
```

## 10.3. Containment

Ví dụ containment:

- Freeze một manifest ĐVVC
- Hold batch nghi vấn
- Tạm dừng xuất kho một SKU
- Tạm ngừng import đơn từ kênh lỗi
- Tạm ngừng data correction batch/hạn dùng
- Tạm revoke quyền user nghi ngờ

## 10.4. Validation

Không đóng ticket nếu chưa có xác nhận của business owner hoặc super user.

Ví dụ:

```text
Lỗi bàn giao ĐVVC → Shipping Lead xác nhận.
Lỗi QC batch → QA Lead xác nhận.
Lỗi tồn kho → Warehouse Lead + Finance nếu ảnh hưởng giá trị hàng.
Lỗi công nợ/COD → Finance Lead xác nhận.
```

---

## 11. Support model theo module

## 11.1. Master Data

### Owner

- ERP Admin
- Business owner từng danh mục

### Lỗi thường gặp

- SKU trùng
- sai đơn vị tính
- sai barcode
- sai batch rule
- supplier/customer sai thông tin
- warehouse/bin mapping sai

### Quy trình support

```text
User báo lỗi master data
→ Super user xác nhận
→ ERP Admin kiểm tra audit log
→ Nếu sửa được theo quyền: tạo data correction request
→ Nếu ảnh hưởng giao dịch đã phát sinh: cần impact analysis
→ Cập nhật và thông báo user liên quan
```

### Cấm

- Tạo SKU tạm không có owner.
- Sửa mã SKU sau khi có giao dịch, trừ khi có procedure đặc biệt.
- Xóa master data đã phát sinh giao dịch.

---

## 11.2. Purchasing / Mua hàng

### Owner

- Purchasing Lead
- Finance nếu liên quan thanh toán

### Lỗi thường gặp

- PR/PO sai trạng thái
- PO chưa duyệt nhưng muốn nhận hàng
- sai supplier
- sai giá/đơn vị tính
- PO nhận thiếu/thừa

### Support rule

```text
PO chưa approved → không cho nhận hàng.
Nhận thừa PO → phải có approval hoặc tạo exception.
Sai giá → không sửa âm thầm, cần finance/purchasing approval.
```

---

## 11.3. Receiving / Nhập kho

### Owner

- Warehouse Lead
- QC nếu có kiểm tra chất lượng

### Lỗi thường gặp

- Không tìm thấy PO để nhập
- Nhập sai batch/hạn dùng
- Hàng không đạt nhưng đã nhập khả dụng
- Sai số lượng thực nhận
- Sai kho/bin

### Support rule

```text
1. Nếu hàng chưa QC pass: không đưa vào available stock.
2. Nếu nhập sai batch/hạn dùng: tạo Data Correction Request.
3. Nếu nhập sai số lượng đã phát sinh xuất: escalation P1/P2 tùy ảnh hưởng.
4. Nếu hàng không đạt: chuyển status Return to Supplier hoặc Quarantine.
```

### Ticket cần có

- PO number
- GRN number
- SKU
- batch/expiry
- số lượng đúng/sai
- ảnh chứng từ nếu có
- người nhập
- thời điểm nhập

---

## 11.4. QA/QC

### Owner

- QA/QC Lead

### Lỗi thường gặp

- Batch đang Hold nhưng bị bán/xuất
- QC pass nhầm
- QC fail nhưng không vào quarantine
- Thiếu file COA/MSDS/test result
- Không trace được complaint về batch

### Support rule

```text
QC status là nghiệp vụ nhạy cảm.
Chỉ QA/QC role được thay đổi.
Nếu nghi sai trạng thái QC → freeze batch ngay trước khi điều tra.
```

### Escalation

- Batch ảnh hưởng đơn đã giao: P1/P0 tùy mức độ
- Batch chưa xuất bán: P2
- Thiếu attachment QC: P3/P2 tùy ảnh hưởng release

---

## 11.5. Inventory / Kho

### Owner

- Warehouse Lead
- ERP Admin hỗ trợ hệ thống

### Lỗi thường gặp

- Tồn vật lý lệch tồn hệ thống
- Tồn khả dụng sai
- Stock bị reserve nhưng không thấy đơn
- Sai bin/khu vực
- Điều chỉnh tồn không có lý do

### Support rule

```text
Không sửa stock balance trực tiếp.
Mọi thay đổi tồn phải qua stock movement hoặc adjustment được duyệt.
```

### Quy trình lệch tồn cuối ngày

```text
1. Kho phát hiện lệch trong shift closing.
2. Tạo Inventory Discrepancy Ticket.
3. Freeze SKU/batch nếu lệch nghiêm trọng.
4. Kiểm tra stock ledger, pick/pack, return, receiving, adjustment.
5. Warehouse Lead xác nhận nguyên nhân.
6. Nếu cần adjustment: tạo phiếu adjustment, cần approval.
7. ERP Admin/Finance review nếu ảnh hưởng giá trị lớn.
8. Close ticket sau khi đối soát lại.
```

---

## 11.6. Warehouse Daily Board / Shift Closing

### Owner

- Warehouse Lead

### Lỗi thường gặp

- Không đóng ca được
- Thiếu số liệu kiểm kê cuối ngày
- Đơn còn pending nhưng đã muốn close shift
- Báo cáo đối soát không khớp

### Support rule

```text
Không đóng ca nếu:
- còn manifest chưa xử lý
- còn đơn picked/packed chưa rõ trạng thái
- còn hàng hoàn chưa scan/ghi nhận
- còn discrepancy chưa có ghi chú
```

### SLA

- Lỗi chặn đóng ca: P1 nếu xảy ra cuối ngày và ảnh hưởng báo cáo quản lý.
- Lỗi report phụ: P2/P3.

---

## 11.7. Sales Order

### Owner

- Sales Admin Lead
- Warehouse nếu liên quan xuất hàng
- Finance nếu liên quan công nợ/thu tiền

### Lỗi thường gặp

- Không tạo được đơn
- Đơn không reserve được tồn
- Sai giá/chiết khấu
- Đơn sai trạng thái
- Đơn bị duplicate từ website/sàn

### Support rule

```text
Đơn chưa reserve không được pick.
Đơn chưa packed không được handover.
Đơn đã delivered/closed không sửa line item trực tiếp.
Duplicate order phải qua idempotency/external reference check.
```

---

## 11.8. Pick / Pack

### Owner

- Packing Lead
- Warehouse Lead

### Lỗi thường gặp

- Pick sai SKU
- Pack thiếu hàng
- Pack nhầm đơn
- Không in/scan được label
- Sai phân loại ĐVVC/sàn

### Support rule

```text
Nếu đơn pack thiếu → không cho handover.
Nếu phát hiện pack nhầm sau bàn giao → chuyển incident shipping/return tùy trạng thái.
```

### Ticket cần có

- Order number
- Packing task number
- SKU/batch
- Người pick/pack
- Ảnh kiện hàng nếu có
- Carrier/platform

---

## 11.9. Shipping / Carrier Handover

### Owner

- Shipping Lead
- Warehouse Lead
- Integration support nếu lỗi carrier API

### Lỗi thường gặp

- Mã đơn không có trên hệ thống khi quét
- Đơn không thuộc manifest
- Thiếu đơn trong thùng/rổ
- Đơn chưa packed nhưng đem bàn giao
- Carrier không nhận trạng thái
- Manifest lệch số lượng

### Support rule

```text
Nếu manifest lệch số lượng → không xác nhận handover toàn batch cho đến khi xử lý.
Nếu mã chưa có trên hệ thống → kiểm tra external order/import, không tạo tay vội.
Nếu đơn có trên hệ thống nhưng không tìm được hàng → tìm trong khu vực đóng hàng trước khi ghi nhận thiếu.
```

### Quy trình xử lý thiếu đơn khi bàn giao

```text
1. Shipping user scan và phát hiện thiếu/chưa đủ.
2. Freeze manifest hoặc line bị lệch.
3. Kiểm tra mã đơn trên hệ thống.
4. Nếu mã chưa có: escalate Sales/Integration.
5. Nếu mã có: tìm lại trong packing zone.
6. Nếu tìm thấy: scan lại và hoàn tất.
7. Nếu không tìm thấy: tạo incident Missing Package.
8. Warehouse Lead xác nhận nguyên nhân.
9. Nếu ảnh hưởng giao hàng trong ngày: P1.
```

---

## 11.10. Returns / Hàng hoàn

### Owner

- Returns Lead
- Warehouse Lead
- QA nếu hàng cần kiểm chất lượng

### Lỗi thường gặp

- Không scan được hàng hoàn
- Không xác định được đơn gốc
- Sai tình trạng còn dùng/không dùng
- Hàng hoàn đưa nhầm vào available stock
- Thiếu ảnh/tình trạng thực tế

### Support rule

```text
Hàng hoàn không được vào available stock nếu chưa qua inspection/disposition.
Tình trạng hàng hoàn phải có người xác nhận.
Không tự đổi từ không sử dụng sang còn sử dụng nếu không có approval.
```

### Quy trình support hàng hoàn

```text
1. Nhận hàng hoàn từ shipper.
2. Nếu scan lỗi: tạo Returns Scan Issue.
3. Kiểm tra order/waybill reference.
4. Nếu không match đơn: đưa vào Unidentified Return Area.
5. Kiểm tra tình trạng sản phẩm.
6. Chọn disposition:
   - Reusable → chuyển kho phù hợp sau kiểm tra
   - Not reusable → chuyển lab/kho lỗi
7. Nếu cần chỉnh disposition: approval từ Warehouse Lead/QA.
```

---

## 11.11. Subcontract Manufacturing / Gia công ngoài

### Owner

- Production/Subcontract Coordinator
- Warehouse khi chuyển NVL/bao bì và nhận hàng
- QA/QC khi duyệt mẫu/kiểm hàng
- Purchasing/Finance khi liên quan cọc/thanh toán

### Lỗi thường gặp

- Sai số lượng NVL/bao bì chuyển nhà máy
- Thiếu biên bản bàn giao
- Thiếu COA/MSDS/term/phụ/hóa đơn VAT nếu cần
- Mẫu chưa duyệt nhưng sản xuất hàng loạt
- Nhà máy giao thiếu/sai quy cách
- Quá hạn claim 3-7 ngày

### Support rule

```text
1. Không sản xuất hàng loạt nếu sample chưa approved.
2. Chuyển NVL/bao bì cho nhà máy phải có subcontract transfer record.
3. Nhận hàng gia công phải qua QC/receiving inspection.
4. Hàng không nhận phải mở factory claim trong SLA 3-7 ngày.
5. Thanh toán lần cuối chỉ sau nghiệm thu theo rule tài chính.
```

### Quy trình claim nhà máy

```text
1. Kho/QA phát hiện hàng không đạt khi nhận.
2. Set receiving result = Rejected / On Hold.
3. Tạo Factory Claim Ticket.
4. Gắn hình ảnh/chứng từ/biên bản.
5. Notify Production/Subcontract Coordinator.
6. Gửi phản hồi nhà máy trong 3-7 ngày theo rule.
7. Theo dõi kết quả: đổi hàng, bù hàng, sửa hàng, trừ tiền, hoặc hủy.
8. Finance xử lý thanh toán cuối theo kết quả claim.
```

---

## 12. Data correction model

## 12.1. Khi nào cần Data Correction Request

Dùng khi dữ liệu đã ghi nhận nhưng cần sửa có kiểm soát.

Ví dụ:

- Nhập sai batch
- Nhập sai expiry date
- Sai bin/location
- Sai trạng thái hàng hoàn
- Sai reference đơn/manifest
- Sai supplier/factory reference
- Sai attachment chứng từ

## 12.2. Data Correction không dùng cho

- Sửa stock balance trực tiếp
- Xóa stock movement
- Xóa audit log
- Đổi QC fail thành pass không qua QA
- Sửa giá vốn không qua Finance
- Sửa trạng thái đơn đã closed không có approval

## 12.3. Workflow Data Correction

```text
User/Super user tạo request
→ Mô tả dữ liệu sai và dữ liệu đúng
→ Gắn chứng từ/ảnh/audit evidence
→ Business owner duyệt
→ Finance/QA duyệt nếu liên quan tiền/chất lượng
→ ERP Admin thực hiện bằng function/procedure chuẩn
→ Audit log ghi nhận before/after
→ Business owner xác nhận
→ Close ticket
```

## 12.4. Mẫu Data Correction Request

```text
Ticket ID:
Loại dữ liệu cần sửa:
Module:
Chứng từ liên quan:
Giá trị hiện tại:
Giá trị đúng:
Lý do sai:
Ảnh hưởng nghiệp vụ:
Người đề xuất:
Người duyệt:
Cách sửa:
Kết quả sau sửa:
```

---

## 13. Access support model

## 13.1. User onboarding

Khi có nhân viên mới:

```text
HR/Manager gửi Access Request
→ Chọn role template
→ Chọn kho/brand/channel được truy cập
→ ERP Admin cấp quyền
→ User đổi mật khẩu lần đầu
→ Super user training SOP liên quan
→ Ghi log hoàn tất onboarding
```

## 13.2. User offboarding

Khi nhân viên nghỉ:

```text
HR gửi offboarding request
→ ERP Admin khóa account đúng thời điểm
→ Revoke token/session
→ Thu hồi quyền nhạy cảm
→ Transfer ownership nếu user đang giữ task/chứng từ
→ Lưu audit
```

## 13.3. Quyền tạm thời

Quyền tạm thời phải có:

```text
- lý do
- thời hạn
- người duyệt
- scope quyền
- auto-expiry
```

Ví dụ:

```text
Cho trưởng ca tạm quyền approve inventory adjustment dưới ngưỡng trong 2 ngày do trưởng kho nghỉ.
```

## 13.4. Access review định kỳ

Ít nhất mỗi quý review:

- user còn làm việc không
- quyền có vượt vai trò không
- quyền finance/cost/payroll nhạy cảm
- quyền QA batch pass/fail
- quyền inventory adjustment
- quyền export dữ liệu
- quyền admin/system config

---

## 14. Change request model

## 14.1. Phân loại change

| Loại change | Ví dụ | Cách xử lý |
|---|---|---|
| Minor config | thêm reason code, đổi label | ERP Admin + PO duyệt |
| Report change | thêm filter/cột báo cáo | Backlog, UAT nhẹ |
| Workflow change | đổi trạng thái, đổi approval | Impact analysis bắt buộc |
| Integration change | thêm ĐVVC, đổi API mapping | Tech review + test staging |
| Security change | đổi permission, role | Security/PO duyệt |
| Data model change | thêm table/field | Architecture review |

## 14.2. Change Request workflow

```text
Submit CR
→ Product Owner review
→ Impact analysis
→ Prioritize
→ Estimate
→ Approve / reject / park
→ Build
→ Test
→ UAT
→ Release
→ Update docs
```

## 14.3. Change Request bắt buộc update tài liệu nào

Nếu đổi workflow:

- PRD/SRS
- Process Flow
- Screen List
- UAT
- SOP

Nếu đổi data:

- Data Dictionary
- API Contract nếu có
- Database Schema
- Migration Plan nếu cần

Nếu đổi quyền:

- Permission Matrix
- Security/RBAC standard
- SOP nếu ảnh hưởng user

Nếu đổi tích hợp:

- Integration Spec
- API Contract
- QA Test Strategy
- Go-Live/Release checklist nếu lớn

---

## 15. Release support model

## 15.1. Release type

| Release | Mục đích | Ví dụ |
|---|---|---|
| Hotfix | Fix lỗi nghiêm trọng | Không scan bàn giao được |
| Patch | Fix bug nhỏ/medium | sửa filter, validate |
| Minor release | thêm cải tiến nhỏ | thêm report, thêm reason code |
| Major release | thay đổi workflow/module | thêm ĐVVC mới, thay logic return |

## 15.2. Release checklist

Trước release:

```text
- Ticket/CR được duyệt
- Code review pass
- Test pass
- Migration script reviewed nếu có
- Smoke test pass trên staging/UAT
- Release note có sẵn
- Rollback plan có sẵn
- Business owner biết thời gian release
```

Sau release:

```text
- Smoke test production
- Theo dõi logs
- Xác nhận business owner
- Cập nhật status ticket
- Cập nhật docs nếu cần
```

## 15.3. Release window

Khuyến nghị:

- Không release vào giờ cao điểm kho đóng hàng/bàn giao ĐVVC.
- Không release sát cuối ngày nếu ảnh hưởng shift closing.
- Không release workflow kho vào ngày nhiều đơn nếu chưa cần khẩn.
- Hotfix P0/P1 có thể release ngoài window nhưng cần approval.

---

## 16. Knowledge Base

## 16.1. Mục tiêu

Knowledge Base giúp giảm phụ thuộc vào dev/admin.

Nội dung nên có:

- SOP theo module
- FAQ
- video/ảnh thao tác
- lỗi thường gặp
- rule trạng thái
- reason code
- escalation guide

## 16.2. Bài viết bắt buộc

```text
1. Cách xử lý đơn không reserve được tồn
2. Cách nhận hàng hoàn
3. Cách xử lý mã không có khi bàn giao ĐVVC
4. Cách đóng ca kho cuối ngày
5. Cách xử lý batch Hold/Pass/Fail
6. Cách tạo factory claim trong 3-7 ngày
7. Cách tạo Data Correction Request
8. Cách xin quyền truy cập
9. Cách phân biệt bug và thao tác sai
10. Cách đọc audit log cơ bản cho super user
```

## 16.3. Owner Knowledge Base

- ERP Admin là owner chính.
- Super user đóng góp nội dung nghiệp vụ.
- Product Owner duyệt nội dung quan trọng.
- Tài liệu phải update sau mỗi change/release lớn.

---

## 17. Training support model

## 17.1. Training theo vai trò

Không training chung một lớp cho tất cả.

| Nhóm | Nội dung training |
|---|---|
| Kho | receiving, stock, pick/pack, handover, returns, shift closing |
| QA/QC | inspection, batch status, hold/pass/fail, claim quality |
| Sales/Admin | sales order, reserve, status, returns request |
| Purchasing | PR/PO, receiving linkage, supplier issue |
| Finance | COD, payment, adjustment approval, report |
| Production/Subcontract | factory order, NVL transfer, sample approval, factory claim |
| ERP Admin | user, role, config, support ticket, audit |

## 17.2. Training cho user mới

Quy trình:

```text
Manager yêu cầu tạo user
→ ERP Admin cấp quyền
→ Super user training theo role
→ User thực hành trên UAT/training data nếu có
→ Manager xác nhận user được phép thao tác thật
```

## 17.3. Training lại sau lỗi

Nếu một lỗi được xác định là thao tác sai, ticket phải có:

```text
- user thao tác sai gì
- SOP đúng là gì
- đã training lại chưa
- có cần chỉnh UI/SOP để tránh lặp lại không
```

---

## 18. Monitoring và health check

## 18.1. Business monitoring

Theo dõi hằng ngày:

```text
- số đơn nhận trong ngày
- số đơn packed
- số đơn handed over
- số đơn thiếu khi bàn giao
- số hàng hoàn chưa xử lý
- số discrepancy tồn kho
- batch đang Hold
- factory claim sắp quá hạn 3-7 ngày
- user ticket theo module
```

## 18.2. Technical monitoring

Theo dõi:

```text
- API error rate
- response time
- database connection
- queue/worker backlog
- failed jobs
- integration sync failure
- storage/upload failure
- login/auth failure
- audit log write failure
```

## 18.3. Daily support review

Trong hypercare hoặc giai đoạn đầu:

```text
Hằng ngày 15-30 phút:
- ticket mới
- P0/P1/P2 đang mở
- lỗi lặp lại
- workflow nào bị nghẽn
- user nào cần training
- change request phát sinh
```

Sau ổn định:

```text
Review 1-2 lần/tuần.
```

---

## 19. Reporting cho support

## 19.1. Báo cáo tuần

ERP Admin/Product Owner gửi báo cáo tuần:

```text
- tổng số ticket
- ticket theo severity
- ticket theo module
- thời gian phản hồi trung bình
- thời gian xử lý trung bình
- ticket quá SLA
- bug lặp lại
- top 5 vấn đề vận hành
- change request mới
- training issue
```

## 19.2. KPI support

| KPI | Mục tiêu đề xuất |
|---|---:|
| P0 response time | <= 15 phút |
| P1 response time | <= 30 phút |
| Ticket có đầy đủ thông tin khi escalate L3 | >= 90% |
| Ticket quá SLA | < 10% |
| Lỗi lặp lại cùng nguyên nhân | Giảm theo tháng |
| Data correction không có approval | 0 |
| Sửa DB trực tiếp không log | 0 |
| User training completion theo role | >= 95% |

---

## 20. RACI tổng quát

| Hoạt động | User | Super User | ERP Admin | Process Owner | Product Owner | Tech Lead/Dev | QA/Test | DevOps |
|---|---|---|---|---|---|---|---|---|
| Báo lỗi | R | C | C | C | I | I | I | I |
| Triage L1 | I | R | C | C | I | I | I | I |
| Cấu hình hệ thống | I | C | R | A/C | C | C | I | I |
| Data correction | C | R/C | R | A | C | C nếu cần | C | I |
| Bug fix | I | C | C | C | A | R | R/C | C |
| Hotfix deploy | I | I | C | C | A | R | C | R |
| Change request | C | C | C | R/A | A | C | C | C |
| Training user | R | R | C | A | I | I | I | I |
| Access request | C | C | R | A | I | I | I | I |
| Security incident | I | C | R | C | A | R/C | I | R/C |

Legend:

```text
R = Responsible
A = Accountable
C = Consulted
I = Informed
```

---

## 21. Escalation matrix

| Tình huống | Escalate tới | Khi nào |
|---|---|---|
| Kho không xuất được hàng | Warehouse Lead → ERP Admin → Tech Lead | Nếu L1 không xử lý trong 30 phút |
| Không bàn giao ĐVVC được | Shipping Lead → ERP Admin → Integration/Tech | Nếu ảnh hưởng đơn trong ngày |
| Lệch tồn lớn | Warehouse Lead → Finance → Product Owner | Ngay khi phát hiện |
| Batch/QC sai trạng thái | QA Lead → Product Owner → Tech Lead | Ngay khi ảnh hưởng bán/xuất |
| Factory claim sắp quá hạn | Subcontract Coordinator → COO/Product Owner | Trước khi vượt SLA 3-7 ngày |
| User có quyền sai | ERP Admin → Security Owner | Ngay khi phát hiện |
| Hệ thống down | ERP Admin → DevOps/Tech Lead → System Owner | P0 ngay lập tức |
| Dữ liệu tiền/công nợ sai | Finance Lead → Product Owner/Tech Lead | P1/P0 tùy mức độ |

---

## 22. Hypercare → steady state transition

## 22.1. Hypercare period

Khuyến nghị 14 ngày sau go-live.

Trong hypercare:

- War room mở hằng ngày
- Ticket review mỗi ngày
- P0/P1 phản hồi nhanh
- Super user túc trực theo ca vận hành
- Theo dõi riêng kho, bàn giao ĐVVC, hàng hoàn, QC, gia công ngoài

## 22.2. Điều kiện rời hypercare

```text
- Không có P0 trong 7 ngày liên tiếp
- P1 giảm và có workaround rõ
- Người dùng chính đã training đủ
- Shift closing chạy ổn định
- Handover ĐVVC không còn lỗi nghiêm trọng lặp lại
- Returns workflow xử lý được
- Factory claim tracking chạy được
- Ticket backlog ở mức kiểm soát
```

## 22.3. Sau hypercare

Chuyển sang support steady state:

```text
- Daily support review → weekly support review
- War room → ticketing normal
- Hotfix khẩn chỉ cho P0/P1
- Change request gom theo sprint/release
```

---

## 23. Mẫu ticket chuẩn

## 23.1. Incident ticket template

```text
Ticket ID:
Ngày giờ phát hiện:
Người báo:
Bộ phận:
Module:
Severity đề xuất:
Mô tả lỗi:
Bước tái hiện:
Chứng từ/mã đơn/mã batch/mã SKU liên quan:
Ảnh/chứng từ đính kèm:
Số user/đơn/batch bị ảnh hưởng:
Có workaround không:
Đã kiểm tra SOP chưa:
L1 owner:
L2 owner:
L3 owner nếu có:
Trạng thái:
Kết quả xử lý:
Xác nhận close bởi:
```

## 23.2. Change request template

```text
CR ID:
Người đề xuất:
Phòng ban:
Mô tả thay đổi:
Lý do:
Vấn đề hiện tại:
Giải pháp mong muốn:
Module ảnh hưởng:
Dữ liệu ảnh hưởng:
Quyền/approval ảnh hưởng:
Báo cáo ảnh hưởng:
Tài liệu cần update:
Mức ưu tiên:
Business approval:
Tech estimate:
UAT required:
Release target:
```

## 23.3. Access request template

```text
User name:
Email/username:
Phòng ban:
Chức vụ:
Role template:
Kho/brand/channel được truy cập:
Quyền đặc biệt nếu có:
Thời hạn quyền tạm thời nếu có:
Manager duyệt:
ERP Admin thực hiện:
Ngày hiệu lực:
```

## 23.4. Data correction template

```text
Request ID:
Module:
Chứng từ liên quan:
Dữ liệu sai:
Dữ liệu đúng:
Nguyên nhân sai:
Ảnh hưởng:
Có ảnh hưởng stock/QC/tiền không:
Người đề xuất:
Người duyệt business:
Người duyệt finance/QA nếu cần:
Cách sửa:
Audit reference:
Kết quả xác nhận:
```

---

## 24. Guardrails quan trọng

### 24.1. Stock

```text
Không sửa balance trực tiếp.
Không xóa movement.
Không đóng ca nếu discrepancy lớn chưa có ghi chú.
Không đưa hàng hoàn vào available nếu chưa inspection.
```

### 24.2. QC / Batch

```text
Không cho bán/xuất batch Hold/Fail.
Không đổi QC status nếu không có quyền QA.
Không release batch thiếu attachment bắt buộc nếu rule yêu cầu.
```

### 24.3. Shipping handover

```text
Không handover đơn chưa Packed.
Không xác nhận manifest nếu thiếu đơn chưa xử lý.
Không bỏ qua scan verify.
```

### 24.4. Returns

```text
Không nhập hàng hoàn chung chung.
Phải có scan/reference.
Phải có tình trạng thực tế.
Phải có disposition còn dùng/không dùng.
```

### 24.5. Subcontract manufacturing

```text
Không chuyển NVL/bao bì cho nhà máy nếu không có transfer record.
Không nhận thành phẩm gia công vào available nếu chưa receiving/QC.
Không để claim nhà máy quá hạn 3-7 ngày mà không có ticket.
```

### 24.6. Permission

```text
Không cấp quyền admin toàn hệ thống trừ trường hợp được duyệt.
Quyền tạm thời phải có hạn.
Quyền nhạy cảm phải review định kỳ.
```

---

## 25. Definition of Done cho support model

Một mô hình support được xem là sẵn sàng khi:

```text
1. Có danh sách super user theo phòng ban.
2. Có ticketing channel chính thức.
3. Có severity/SLA được duyệt.
4. Có escalation matrix.
5. Có data correction workflow.
6. Có access request/offboarding workflow.
7. Có change request workflow.
8. Có knowledge base tối thiểu.
9. Có training role-based.
10. Có support dashboard/report tuần.
11. Có hypercare plan và điều kiện chuyển steady state.
12. Có guardrails cho stock, QC, returns, handover, subcontract.
```

---

## 26. Checklist triển khai support model trước go-live

```text
[ ] Chỉ định System Owner
[ ] Chỉ định Product Owner
[ ] Chỉ định ERP Admin
[ ] Chỉ định super user từng bộ phận
[ ] Thiết lập ticketing tool
[ ] Cấu hình severity/SLA
[ ] Tạo mẫu ticket incident/change/access/data correction
[ ] Tạo nhóm war room hypercare
[ ] Training super user
[ ] Training user theo vai trò
[ ] Tạo Knowledge Base tối thiểu
[ ] Kiểm tra access/offboarding process
[ ] Kiểm tra data correction process
[ ] Kiểm tra backup/rollback contact
[ ] Chốt release/hotfix approval
[ ] Chốt report support hằng ngày trong hypercare
```

---

## 27. Tài liệu liên quan

Tài liệu này liên kết trực tiếp với:

```text
19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md
20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md
21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md
22_ERP_Core_Docs_Revision_v1_1_Change_Log_Phase1_MyPham.md
24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md
25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md
26_ERP_SOP_Training_Manual_Phase1_MyPham_v1.md
27_ERP_GoLive_Runbook_Hypercare_Phase1_MyPham_v1.md
28_ERP_Risk_Incident_Playbook_Phase1_MyPham_v1.md
```

---

## 28. Kết luận

Operations Support Model là lớp giữ ERP sống sau go-live.

Không có nó, hệ thống sẽ phụ thuộc vào một vài người giỏi, sửa lỗi qua chat, chỉnh dữ liệu thiếu kiểm soát, và dần biến thành một phiên bản Excel có giao diện đẹp.

Có nó, công ty có thể vận hành theo kỷ luật:

```text
Lỗi được ghi nhận
→ phân loại đúng
→ xử lý đúng người
→ dữ liệu không bị phá
→ thay đổi có kiểm soát
→ người dùng được đào tạo
→ hệ thống ngày càng sạch hơn
```

Đối với công ty mỹ phẩm có kho, hàng hoàn, batch/QC, bàn giao ĐVVC và gia công ngoài, đây không phải tài liệu phụ. Đây là **bộ luật hậu chiến** sau khi ERP chính thức bước vào vận hành thật.
