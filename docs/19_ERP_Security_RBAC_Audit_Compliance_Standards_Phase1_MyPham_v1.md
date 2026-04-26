# 19 ERP Security / RBAC / Audit / Compliance Standards — Phase 1 Mỹ Phẩm v1

**Tên file:** `19_ERP_Security_RBAC_Audit_Compliance_Standards_Phase1_MyPham_v1.md`  
**Dự án:** Web ERP cho công ty sản xuất, phân phối, bán lẻ mỹ phẩm  
**Phiên bản:** v1.0  
**Trạng thái:** Baseline để team BA/PM/Tech Lead/Dev/QA triển khai  
**Phạm vi:** Phase 1 — Master Data, Mua hàng, QC, Kho, Bán hàng, Giao hàng, Hàng hoàn, Gia công ngoài, đối soát vận hành kho

---

## 1. Mục tiêu tài liệu

Tài liệu này khóa chuẩn **bảo mật, phân quyền, kiểm soát truy cập, audit log và compliance nội bộ** cho ERP Phase 1.

ERP mỹ phẩm không chỉ lưu dữ liệu. Nó điều khiển **hàng hóa, tiền, quyền duyệt, lô hàng, QC, giao vận, hàng hoàn, nhà cung cấp, nhà máy gia công và báo cáo quản trị**. Nếu bảo mật và phân quyền lỏng, hệ thống sẽ rất dễ biến thành một cái Excel có giao diện đẹp nhưng không kiểm soát được ai đã làm gì.

Mục tiêu chính:

- Không ai được xem/sửa dữ liệu vượt quyền.
- Không ai được tự tạo, tự duyệt, tự thực hiện, tự xóa dấu vết trong cùng một nghiệp vụ rủi ro cao.
- Mọi hành động quan trọng phải có audit log.
- Stock ledger, QC, batch, chứng từ duyệt, bàn giao ĐVVC, hàng hoàn và gia công ngoài phải có dấu vết rõ ràng.
- Khi có lỗi hoặc gian lận, truy được người, thời điểm, dữ liệu trước/sau và chứng từ liên quan.

Một câu chốt nguyên tắc:

> **ERP không được tin người dùng tuyệt đối. ERP phải tin quy trình, quyền hạn, log và bằng chứng.**

---

## 2. Phạm vi áp dụng

Áp dụng cho toàn bộ hệ thống ERP Phase 1 gồm:

1. Authentication / đăng nhập
2. Authorization / RBAC / ABAC
3. Field-level permission
4. Approval security
5. Audit log
6. Stock ledger security
7. QC / batch / expiry security
8. Purchase / supplier / subcontract manufacturing security
9. Sales / shipping / carrier handover security
10. Return / hàng hoàn security
11. Data privacy
12. API security
13. File/attachment security
14. Incident response
15. Compliance baseline

Không bao gồm chi tiết pháp lý thuế, chuẩn GMP/ISO đầy đủ hoặc chứng nhận mỹ phẩm cấp nhà nước. Các phần đó có thể bổ sung ở phase sau hoặc tài liệu compliance riêng.

---

## 3. Nguyên tắc bảo mật nền

### 3.1 Deny by default

Mặc định người dùng **không có quyền gì** cho tới khi được cấp role hoặc permission cụ thể.

Không dùng logic kiểu:

```text
Nếu chưa cấu hình quyền thì cho xem tạm.
```

Phải dùng:

```text
Nếu chưa cấu hình quyền thì chặn.
```

### 3.2 Least privilege

Mỗi user chỉ có quyền tối thiểu để làm việc.

Ví dụ:

- Nhân viên đóng hàng không cần xem giá vốn.
- Nhân viên kho không cần xem công nợ khách hàng.
- Sale không được sửa tồn kho.
- Finance không được tự đổi QC status.
- QA/QC không được tự tạo thanh toán nhà cung cấp.

### 3.3 Separation of Duties — SoD

Không để một người kiểm soát trọn vòng đời giao dịch rủi ro cao.

Ví dụ cấm:

```text
Một user tự tạo PO → tự duyệt PO → tự nhận hàng → tự duyệt thanh toán.
```

Ví dụ đúng:

```text
Purchasing tạo PO → Manager/CEO duyệt → Kho nhận hàng → QC kiểm → Finance đối chiếu hóa đơn/thanh toán.
```

### 3.4 Immutable evidence

Các dữ liệu sau **không được sửa trực tiếp** sau khi phát sinh:

- Stock ledger
- Audit log
- Approval history
- QC decision log
- Shipment handover log
- Return inspection log
- Batch movement log
- Payment approval log

Nếu sai, phải tạo giao dịch điều chỉnh hoặc reversal.

### 3.5 Traceability first

Mọi nghiệp vụ hàng hóa phải truy được:

```text
Chứng từ → SKU → batch/lô → kho/vị trí → người thao tác → thời gian → trạng thái → chứng từ liên quan.
```

Riêng ngành mỹ phẩm, batch và hạn dùng là “xương sống niềm tin”. Không trace được batch là không kiểm soát được rủi ro.

### 3.6 No shared account

Không dùng tài khoản chung kiểu:

```text
kho01
sales01
admin
```

Mỗi người dùng phải có tài khoản cá nhân. Nếu dùng máy chung ở kho, vẫn phải có user riêng hoặc cơ chế quick login/scan badge sau này.

### 3.7 Human-friendly nhưng không lỏng

Giao diện phải dễ dùng, nhưng các hành động nguy hiểm phải có lớp xác nhận:

- Duyệt QC fail/pass
- Điều chỉnh tồn
- Hủy chứng từ
- Đổi hạn dùng/batch
- Xác nhận bàn giao ĐVVC
- Xác nhận hàng hoàn không sử dụng
- Thanh toán nhà cung cấp
- Thay đổi giá/chiết khấu lớn

---

## 4. Phân loại dữ liệu

### 4.1 Bảng phân loại

| Cấp dữ liệu | Ý nghĩa | Ví dụ | Kiểm soát |
|---|---|---|---|
| Public | Dữ liệu có thể công khai | Tên sản phẩm public, hình ảnh sản phẩm đã công bố | Có thể view rộng, không cho sửa tự do |
| Internal | Dữ liệu nội bộ phổ thông | Tên SKU, danh mục kho, trạng thái đơn nội bộ | Theo role |
| Confidential | Dữ liệu nhạy cảm | Giá mua, giá vốn, công nợ, tồn kho chi tiết, hồ sơ NCC, QC fail, hàng hoàn | Role + field-level permission |
| Restricted | Dữ liệu cực nhạy cảm | Bảng lương, payout KOL, tài khoản ngân hàng, audit log nhạy cảm, dữ liệu điều tra sự cố | Role đặc biệt + audit bắt buộc + hạn chế export |

### 4.2 Dữ liệu cần hạn chế mạnh

Các nhóm dữ liệu sau phải có field-level permission:

- Giá vốn sản phẩm
- Giá mua nguyên vật liệu
- Điều kiện thanh toán nhà cung cấp
- Công nợ khách hàng/NCC
- Bảng lương, OT, thưởng/phạt
- KOL payout / commission
- Batch lỗi, QC fail, complaint nghiêm trọng
- Điều chỉnh tồn kho
- File COA/MSDS/hợp đồng/biên bản bàn giao
- Dữ liệu ngân hàng và thông tin định danh cá nhân

### 4.3 Quy tắc export dữ liệu

Export Excel/CSV là điểm thất thoát dữ liệu rất lớn.

Quy tắc:

- Chỉ role được cấp quyền `*.export` mới được export.
- Export dữ liệu nhạy cảm phải ghi audit log.
- Export giá vốn/công nợ/bảng lương phải giới hạn role.
- Export số lượng lớn phải cần lý do hoặc approval tùy chính sách.
- File export nên có watermark metadata nếu triển khai được: user, thời điểm, module.

---

## 5. Authentication — đăng nhập và phiên làm việc

### 5.1 Phương thức đăng nhập Phase 1

Khuyến nghị:

```text
Email/username + password
JWT access token + refresh token hoặc secure session cookie
```

Nếu dùng Next.js + Go backend:

- Frontend giữ session an toàn.
- Backend xác thực token/session ở mọi API.
- Refresh token/session phải có cơ chế revoke.

### 5.2 Password policy

Tối thiểu:

- Tối thiểu 10 ký tự
- Có chữ hoa, chữ thường, số hoặc ký tự đặc biệt
- Không dùng password phổ biến
- Không trùng 5 password gần nhất nếu có đổi mật khẩu
- Khóa tạm nếu sai nhiều lần

Khuyến nghị:

```text
Sai 5 lần trong 15 phút → khóa 15 phút hoặc yêu cầu admin reset.
```

### 5.3 MFA

Phase 1 có thể chưa bắt buộc toàn công ty, nhưng nên bắt buộc cho role rủi ro cao:

- CEO / Owner
- System Admin
- Finance Manager
- Warehouse Manager
- QA Manager
- User có quyền export dữ liệu nhạy cảm
- User có quyền đổi role/permission

### 5.4 Session policy

| Loại user | Session idle timeout | Ghi chú |
|---|---:|---|
| Kho/đóng hàng | 4–8 giờ | Do thao tác ca dài, nên timeout vừa đủ, nhưng action nguy hiểm cần xác nhận lại |
| Backoffice | 2–4 giờ | Theo ca văn phòng |
| Admin/Finance | 30–120 phút | Dữ liệu nhạy cảm |
| Break-glass admin | 15–30 phút | Rất nhạy cảm |

### 5.5 Re-authentication cho hành động nguy hiểm

Các action sau nên yêu cầu nhập lại mật khẩu/MFA hoặc xác nhận mạnh:

- Cấp quyền admin
- Sửa role/permission
- Duyệt thanh toán lớn
- Điều chỉnh tồn kho lớn
- Override QC hold/fail
- Xóa/hủy chứng từ quan trọng
- Export dữ liệu nhạy cảm
- Kích hoạt break-glass access

---

## 6. User lifecycle

### 6.1 Tạo user mới

Quy trình:

```text
HR/Admin tạo hồ sơ nhân viên
→ System Admin tạo user
→ Gán phòng ban/kho/chi nhánh
→ Gán role theo chức năng
→ Manager xác nhận
→ User đổi mật khẩu lần đầu
```

Không tạo user trực tiếp từ dev/database.

### 6.2 Thay đổi vai trò

Khi user đổi bộ phận, phải có ticket/change request:

```text
User A chuyển từ Kho sang Sales
→ Remove role Kho
→ Add role Sales
→ Audit log
→ Manager mới xác nhận
```

### 6.3 Nghỉ việc / offboarding

Offboarding phải làm trong ngày nghỉ việc hoặc ngay khi có quyết định:

- Disable login
- Revoke session/token
- Thu hồi quyền export
- Thu hồi quyền duyệt
- Ghi log offboarding
- Chuyển ownership chứng từ đang pending nếu cần

### 6.4 Temporary access

Quyền tạm thời phải có:

- Lý do
- Người duyệt
- Thời hạn hết hiệu lực
- Tự động revoke
- Audit log

Không để quyền tạm biến thành quyền vĩnh viễn.

---

## 7. Authorization model

### 7.1 Mô hình đề xuất

Dùng kết hợp:

```text
RBAC + policy/context checks
```

RBAC trả lời:

```text
Người này có quyền thao tác loại này không?
```

Policy/context checks trả lời:

```text
Trong ngữ cảnh này có được làm không?
```

Ví dụ:

```text
User có quyền approve PO,
nhưng chỉ approve PO thuộc department của mình,
và PO dưới ngưỡng tiền được cấp.
```

### 7.2 Scope quyền

Permission không chỉ là module/action. Phải có scope:

| Scope | Ý nghĩa |
|---|---|
| Company | Toàn công ty |
| Brand | Theo brand |
| Warehouse | Theo kho |
| Branch/Store | Theo chi nhánh/cửa hàng |
| Department | Theo phòng ban |
| Channel | Theo kênh bán |
| Own | Chỉ dữ liệu mình tạo/phụ trách |
| Assigned | Dữ liệu được giao |

### 7.3 Permission naming convention

Định dạng:

```text
module.resource.action
```

Ví dụ:

```text
inventory.stock.view
inventory.stock.adjust.request
inventory.stock.adjust.approve
inventory.stock_movement.view
qc.inspection.create
qc.inspection.approve
shipping.manifest.handover
returns.inspection.disposition
purchase.po.create
purchase.po.approve
finance.cost.view
system.role.assign
```

### 7.4 Action chuẩn

| Action | Ý nghĩa |
|---|---|
| view | Xem |
| create | Tạo mới |
| update | Sửa |
| submit | Gửi duyệt |
| approve | Duyệt |
| reject | Từ chối |
| cancel | Hủy |
| close | Đóng nghiệp vụ |
| export | Xuất dữ liệu |
| import | Nhập dữ liệu hàng loạt |
| assign | Giao việc/giao owner |
| override | Vượt rule, quyền rất nhạy cảm |
| adjust | Điều chỉnh dữ liệu đã phát sinh |
| handover | Bàn giao ĐVVC/nhà máy/kho |
| inspect | Kiểm tra/QC |
| release | Release batch/QC |

---

## 8. Role catalog Phase 1

### 8.1 Role cấp hệ thống

| Role | Mục tiêu | Quyền chính | Hạn chế |
|---|---|---|---|
| System Admin | Quản lý cấu hình hệ thống | User, role, permission, config | Không được tự sửa chứng từ nghiệp vụ nếu không có role nghiệp vụ |
| Business Admin | Quản trị master data được phân công | Tạo/sửa master data | Không được cấp quyền hệ thống |
| Auditor / Viewer | Kiểm tra dữ liệu | View + audit log theo scope | Không create/update/approve |
| CEO / Owner | Điều hành và duyệt cấp cao | Dashboard, approval cấp cao, báo cáo | Không nên thao tác kho trực tiếp |

### 8.2 Role kho/giao vận

| Role | Quyền chính | Không được |
|---|---|---|
| Warehouse Manager | Duyệt/giám sát nhập-xuất, kiểm kê, đối soát cuối ngày, bàn giao ĐVVC | Sửa giá vốn, duyệt thanh toán, đổi QC decision |
| Warehouse Staff | Nhập kho, xuất kho, chuyển vị trí, scan, pick/pack theo nhiệm vụ | Duyệt điều chỉnh tồn, pass/fail QC, xem giá vốn |
| Packer | Soạn/đóng hàng, scan đơn, xác nhận pack | Sửa đơn hàng, sửa tồn, xem công nợ |
| Handover Staff | Đối chiếu manifest, quét mã, bàn giao ĐVVC | Tạo đơn, điều chỉnh tồn, sửa số lượng đơn |
| Return Receiver | Nhận hàng hoàn, scan, ghi tình trạng ban đầu | Quyết định tài chính/hoàn tiền nếu không có quyền CS/Finance |

### 8.3 Role mua hàng/nhà cung cấp

| Role | Quyền chính | Không được |
|---|---|---|
| Purchasing Staff | Tạo PR/RFQ/PO, theo dõi giao hàng | Tự duyệt PO vượt ngưỡng, xác nhận thanh toán |
| Purchasing Manager | Duyệt PR/PO theo ngưỡng, quản NCC | Sửa QC pass/fail, tự thanh toán |
| Supplier Viewer/Internal | Xem trạng thái giao hàng nếu có portal sau này | Không truy cập dữ liệu nội bộ ngoài phạm vi |

### 8.4 Role QA/QC

| Role | Quyền chính | Không được |
|---|---|---|
| QC Inspector | Tạo phiếu kiểm, ghi kết quả test/inspection | Release batch nếu chưa đủ quyền, sửa stock ledger |
| QA Manager | Duyệt QC pass/fail, release/hold batch, CAPA | Tạo đơn bán, duyệt thanh toán, sửa giá vốn |
| R&D Viewer/Contributor | Xem spec, sample, version công thức nếu cần | Duyệt QC thương mại nếu không được cấp quyền |

### 8.5 Role sales/giao hàng/CS

| Role | Quyền chính | Không được |
|---|---|---|
| Sales Admin | Tạo/sửa sales order trước khi confirmed, xem tồn khả dụng | Sửa tồn, override QC hold, xem giá vốn nếu không có quyền |
| Sales Manager | Duyệt discount, giữ hàng lớn, xem báo cáo sales | Duyệt stock adjustment, QC release |
| CSKH | Xem đơn, tạo ticket, tạo yêu cầu đổi trả/hàng hoàn | Tự nhập lại hàng vào tồn khả dụng, sửa công nợ |
| Shipping Coordinator | Gán ĐVVC, tạo manifest, theo dõi bàn giao | Sửa đơn hàng sau khi đóng gói nếu không có quyền |

### 8.6 Role tài chính

| Role | Quyền chính | Không được |
|---|---|---|
| Finance Staff | Ghi nhận thu/chi, đối soát COD, công nợ | Sửa QC, sửa stock ledger, tạo PO không qua quy trình |
| Finance Manager | Duyệt thanh toán, xem giá vốn, xem công nợ | Override hàng hóa/QC nếu không có approval đặc biệt |
| Accountant Viewer | Xem chứng từ tài chính theo scope | Không sửa chứng từ vận hành |

### 8.7 Role gia công ngoài / sản xuất

| Role | Quyền chính | Không được |
|---|---|---|
| Production/Subcontract Coordinator | Tạo đơn gia công, theo dõi chuyển NVL/bao bì, mẫu, nhận hàng | Tự duyệt QC pass/fail, tự thanh toán final payment |
| Factory Liaison | Ghi tiến độ nhà máy, nhận thông báo lỗi, cập nhật timeline | Sửa tồn kho/chi phí nếu không có quyền |
| Production Manager | Duyệt kế hoạch gia công, kiểm soát tiến độ | Duyệt thanh toán nếu không có role tài chính |

---

## 9. Field-level permission

### 9.1 Nhóm field nhạy cảm

| Nhóm field | Ví dụ | Ai được xem/sửa |
|---|---|---|
| Cost | `unit_cost`, `standard_cost`, `actual_cost`, `cogs` | Finance, CEO, role được cấp |
| Supplier price | `supplier_quote`, `purchase_price`, `payment_terms` | Purchasing, Finance, CEO |
| Salary/payroll | `salary`, `bonus`, `deduction`, `bank_account` | HR, Finance, CEO |
| KOL payout | `commission_rate`, `payout_amount`, `contract_fee` | Marketing Manager, Finance, CEO |
| QC sensitive | `fail_reason`, `complaint_severity`, `capa_note` | QA/QC, CEO, role liên quan |
| Batch critical | `batch_no`, `expiry_date`, `mfg_date`, `qc_status` | Tạo theo nghiệp vụ; sửa chỉ role đặc biệt + audit |
| Inventory adjustment | `adjust_qty`, `adjust_reason`, `approved_by` | Kho tạo yêu cầu, manager/finance duyệt tùy rule |

### 9.2 Quy tắc UI/API

Nếu user không có quyền xem field:

- API không trả field đó, hoặc trả masked value.
- Frontend không chỉ ẩn bằng UI; backend phải enforce.
- Export cũng phải tôn trọng field permission.

Ví dụ:

```json
{
  "sku_code": "SERUM-VITC-30ML",
  "stock_qty": 120,
  "unit_cost": null,
  "cost_hidden": true
}
```

---

## 10. Approval security

### 10.1 Nguyên tắc duyệt

- Người tạo không được tự duyệt chứng từ rủi ro cao.
- Mọi approval phải ghi audit log.
- Approval có trạng thái riêng, không chỉ là một checkbox.
- Rejection phải có lý do.
- Nếu chứng từ đã approved mà sửa field quan trọng, phải reset approval hoặc tạo version mới.

### 10.2 Các nghiệp vụ cần duyệt

| Nghiệp vụ | Cần duyệt bởi | Ghi chú |
|---|---|---|
| PO vượt ngưỡng | Purchasing Manager/CEO/Finance | Theo giá trị đơn |
| Nhập kho lệch PO | Warehouse Manager + Purchasing/QC | Nếu thiếu/thừa/hỏng |
| QC pass/fail batch | QA Manager | Không cho kho tự pass |
| Điều chỉnh tồn | Warehouse Manager + Finance/CEO tùy ngưỡng | Không sửa ledger trực tiếp |
| Hủy đơn đã confirmed | Sales Manager/Finance tùy trạng thái | Nếu đã giữ/xuất hàng phải xử lý tồn |
| Discount vượt ngưỡng | Sales Manager/CEO | Theo channel/customer |
| Bàn giao ĐVVC | Warehouse/Handover Staff + Carrier confirmation | Manifest + scan log |
| Hàng hoàn không sử dụng | Return Receiver + QC/CSKH/Manager | Quyết định disposition |
| Chuyển NVL/bao bì sang nhà máy | Production/Subcontract Manager + Kho | Có biên bản bàn giao |
| Thanh toán final nhà máy/NCC | Finance Manager/CEO | Phải đối chiếu nghiệm thu/QC |

### 10.3 Approval state

```text
Draft
→ Submitted
→ In Review
→ Approved / Rejected
→ Cancelled
```

Nếu có nhiều cấp:

```text
Submitted
→ Department Approved
→ Finance Approved
→ CEO Approved
```

### 10.4 Sensitive approval confirmation

Với action nguy hiểm, cần confirm:

```text
Bạn đang duyệt QC PASS cho batch X. Sau khi pass, batch có thể được bán/xuất. Xác nhận?
```

Và ghi:

- user
- timestamp
- IP/device
- reason/note
- before/after state
- related document

---

## 11. Audit log standard

### 11.1 Mục tiêu audit

Audit log trả lời được 6 câu:

```text
Ai?
Làm gì?
Lúc nào?
Ở đâu?
Trước/sau thay đổi gì?
Vì sao / theo chứng từ nào?
```

### 11.2 Audit log schema khuyến nghị

```text
audit_logs
- id
- event_id
- actor_user_id
- actor_role_snapshot
- action
- module
- entity_type
- entity_id
- before_data_json
- after_data_json
- reason
- ip_address
- user_agent
- request_id
- correlation_id
- created_at
```

### 11.3 Bắt buộc audit cho các nhóm hành động

| Nhóm | Ví dụ action |
|---|---|
| Auth | login, logout, failed login, password reset, MFA change |
| User/role | create user, disable user, assign role, remove role |
| Master data | tạo/sửa SKU, batch rule, NCC, kho, bảng giá |
| Purchase | tạo/duyệt/sửa/hủy PR/PO |
| Inventory | nhập, xuất, chuyển kho, điều chỉnh tồn, kiểm kê |
| Stock ledger | mọi movement, reversal |
| QC | tạo inspection, pass/fail/hold, release batch |
| Sales | tạo/sửa/hủy đơn, discount, reserve stock |
| Shipping | pack, handover, scan manifest, carrier confirmation |
| Returns | nhận hàng hoàn, scan, disposition, nhập lại/hủy/lab |
| Subcontract | chuyển NVL/bao bì, duyệt mẫu, nhận hàng, báo lỗi nhà máy |
| Finance | ghi nhận thanh toán, duyệt thanh toán, đối soát COD |
| Export | export dữ liệu nhạy cảm |

### 11.4 Retention

Khuyến nghị:

```text
Audit log nghiệp vụ: tối thiểu 5 năm
Auth log: tối thiểu 1–2 năm
Security incident log: tối thiểu 5 năm hoặc theo chính sách công ty
```

### 11.5 Audit log không được sửa

Không cho update/delete audit log từ app.

Nếu cần correction:

```text
Tạo audit correction note mới, không sửa log cũ.
```

---

## 12. Stock security

### 12.1 Stock ledger bất biến

Không có API kiểu:

```text
UPDATE stock_balance SET qty = ...
```

Tồn kho chỉ thay đổi thông qua stock movement:

```text
INBOUND_RECEIPT
QC_RELEASE
QC_HOLD
SALES_RESERVE
SALES_RELEASE_RESERVE
SALES_ISSUE
SHIPMENT_HANDOVER
RETURN_RECEIPT
RETURN_RESTOCK
RETURN_TO_LAB
SUBCONTRACT_MATERIAL_ISSUE
SUBCONTRACT_FINISHED_RECEIPT
CYCLE_COUNT_ADJUSTMENT
DAMAGE_WRITE_OFF
TRANSFER
REVERSAL
```

### 12.2 Phân biệt tồn

ERP phải phân biệt:

```text
Physical stock
Available stock
Reserved stock
QC hold stock
Return pending stock
Damaged/lab stock
Subcontract issued stock
```

Không cho sale nhìn hàng hold hoặc hàng hoàn pending như hàng bán được.

### 12.3 Điều chỉnh tồn

Điều chỉnh tồn là action nhạy cảm.

Quy trình:

```text
Kho tạo adjustment request
→ nêu lý do + bằng chứng
→ Warehouse Manager kiểm
→ Finance/CEO duyệt nếu vượt ngưỡng
→ hệ thống tạo stock movement ADJUSTMENT
→ audit log
```

### 12.4 Kiểm kê cuối ngày / shift closing

Do workflow kho có bước kiểm kê cuối ngày và đối soát số liệu, ERP phải khóa:

- Ai được bắt đầu kiểm kê
- Ai được chốt kiểm kê
- Chốt xong có được sửa không
- Lệch tồn xử lý bằng adjustment request, không sửa tay
- Shift closing phải ghi thời điểm, người chốt, số đơn xử lý, số đơn còn pending, tồn lệch nếu có

---

## 13. QC / batch / expiry security

### 13.1 Batch là dữ liệu trọng yếu

Batch/lô không được đổi tùy tiện.

Nếu nhập sai batch:

```text
Tạo correction/reversal hoặc adjustment có approval.
```

Không sửa thẳng batch trên phiếu đã phát sinh movement.

### 13.2 QC status

QC status chuẩn:

```text
HOLD
PASS
FAIL
QUARANTINE
RETEST_REQUIRED
```

Rule:

- Batch mới nhập mặc định HOLD hoặc Pending Inspection.
- Chỉ QA/QC được pass/fail.
- Batch FAIL không cho bán/xuất thương mại.
- Batch HOLD không cho available stock.
- Override QC cần quyền đặc biệt + lý do + approval + audit.

### 13.3 Hạn dùng

`expiry_date` là critical field.

- Bắt buộc nhập với mỹ phẩm/thành phẩm/nguyên liệu có hạn.
- Sửa expiry date sau khi nhập kho phải có role đặc biệt.
- Nếu sửa expiry, audit phải lưu before/after.
- FEFO phải dựa trên expiry date đã được xác nhận.

### 13.4 Complaint linked to batch

Nếu khách khiếu nại liên quan sản phẩm/lô, phải liên kết được:

```text
Customer complaint → sales order → SKU → batch → QC record → supplier/subcontract/manufacturing record.
```

---

## 14. Shipping handover security

### 14.1 Carrier manifest

Bàn giao ĐVVC phải qua manifest/chuyến/bảng bàn giao.

Không cho trạng thái `HandedOver` nếu:

- Đơn chưa packed
- Chưa scan verify
- Không thuộc manifest
- Số lượng trong manifest chưa đối chiếu
- User không có quyền handover

### 14.2 Scan log

Mỗi lần quét phải ghi:

```text
scan_id
user_id
order_id / tracking_no
manifest_id
carrier_id
scan_result
error_code nếu có
timestamp
station/device nếu có
```

### 14.3 Thiếu đơn khi bàn giao

Nếu scan thiếu đơn:

- Không cho hoàn tất manifest nếu chưa xử lý hoặc override có quyền.
- Hệ thống phải tạo exception.
- Exception phải có trạng thái: `Open`, `Investigating`, `Resolved`, `Cancelled`.
- Nếu tìm lại ở khu đóng hàng thì ghi resolution.
- Nếu đóng lại thì tạo hành động pack/repack có log.

### 14.4 Carrier confirmation

Bàn giao hoàn tất nên có:

- Người kho xác nhận
- Người ĐVVC xác nhận, chữ ký/file đính kèm nếu có
- Số đơn bàn giao
- Số kiện/thùng/rổ
- Thời gian bàn giao

---

## 15. Return / hàng hoàn security

### 15.1 Hàng hoàn không tự vào tồn khả dụng

Hàng hoàn sau khi nhận từ shipper phải vào trạng thái:

```text
RETURN_PENDING_INSPECTION
```

Không cho nhập thẳng available stock.

### 15.2 Return disposition

Disposition chuẩn:

```text
REUSABLE
DAMAGED
TO_LAB
TO_DISPOSAL
NEED_QC_REVIEW
MISSING_ITEM
WRONG_ITEM
```

Nguyên tắc:

- Người nhận hàng hoàn ghi nhận tình trạng ban đầu.
- QC/Manager duyệt disposition nếu hàng có rủi ro.
- Nếu `REUSABLE` mới tạo movement `RETURN_RESTOCK`.
- Nếu `TO_LAB` hoặc `DAMAGED`, không vào hàng bán.

### 15.3 Bằng chứng hàng hoàn

Nên có:

- Ảnh/video tình trạng hàng
- Lý do hoàn
- Mã vận đơn/mã đơn
- Người nhận
- Thời điểm
- Kết quả kiểm tra

---

## 16. Subcontract manufacturing security

### 16.1 Đặc thù workflow

Phase 1 phải hỗ trợ gia công ngoài:

```text
Đặt nhà máy
→ xác nhận số lượng/quy cách/mẫu
→ cọc đơn
→ chuyển NVL/bao bì
→ ký biên bản bàn giao
→ làm mẫu/chốt mẫu
→ sản xuất hàng loạt
→ giao về kho
→ kiểm số lượng/chất lượng
→ nhận hàng hoặc báo lỗi nhà máy
→ thanh toán cuối
```

### 16.2 Chuyển NVL/bao bì sang nhà máy

Không coi chuyển ra nhà máy là mất hàng.

Phải tạo trạng thái tồn:

```text
SUBCONTRACT_ISSUED
```

Cần log:

- SKU/NVL/bao bì
- batch/lô nếu có
- số lượng
- nhà máy
- biên bản bàn giao
- file COA/MSDS/spec nếu có
- người bàn giao/nhận

### 16.3 Sample approval

Chốt mẫu phải có:

- Version mẫu
- Người duyệt
- Ngày duyệt
- File/ảnh mẫu
- Lưu mẫu nếu cần
- Không cho sản xuất hàng loạt nếu chưa có sample approval, trừ quyền override đặc biệt

### 16.4 Nhận hàng gia công

Khi nhà máy giao hàng về:

- Kho nhận số lượng ban đầu
- QC kiểm chất lượng
- Nếu đạt → nhập kho thành phẩm/bán thành phẩm theo batch
- Nếu không đạt → tạo defect report
- Báo lỗi nhà máy trong SLA nội bộ, ví dụ 3–7 ngày theo workflow hiện tại
- Thanh toán cuối chỉ nên mở khi nghiệm thu đạt hoặc có quyết định exception

---

## 17. API security

### 17.1 Auth bắt buộc

Tất cả API nghiệp vụ phải require authentication, trừ endpoint login/health public.

### 17.2 Permission check ở backend

Không tin frontend.

Mọi API phải check:

```text
user authenticated?
permission exists?
scope valid?
entity state cho phép action?
SoD rule có vi phạm không?
```

### 17.3 Idempotency

Các API tạo giao dịch quan trọng phải hỗ trợ idempotency:

- Tạo stock movement
- Xác nhận shipment handover
- Ghi nhận payment
- Nhận hàng hoàn
- Tạo PO/SO nếu request có nguy cơ retry

Header:

```text
Idempotency-Key: <uuid>
```

### 17.4 Rate limit

Áp dụng cho:

- Login
- Password reset
- API scan nếu bị spam bất thường
- Export
- Public webhook nếu có

### 17.5 Error response không lộ dữ liệu nhạy cảm

Không trả lỗi kiểu:

```text
SQL error: duplicate key value violates unique constraint...
```

Phải trả:

```json
{
  "success": false,
  "code": "DUPLICATE_SKU_CODE",
  "message": "Mã SKU đã tồn tại."
}
```

---

## 18. File / attachment security

### 18.1 File cần quản lý

- COA/MSDS
- Biên bản bàn giao
- Phiếu nhập/xuất
- Ảnh/video hàng hoàn
- Hợp đồng NCC/nhà máy
- Mẫu/chốt mẫu sản xuất
- Chứng từ thanh toán
- Bằng chứng giao nhận ĐVVC

### 18.2 Quy tắc truy cập file

- Không expose URL public vĩnh viễn.
- Dùng signed URL có thời hạn.
- Kiểm tra permission trước khi cấp URL.
- Log download file nhạy cảm.
- File upload phải kiểm type/size.
- Quét malware nếu có điều kiện.

### 18.3 Naming file

Không phụ thuộc tên file người dùng upload.

Nên lưu:

```text
storage_key = module/yyyy/mm/entity_id/file_uuid.ext
original_filename = tên gốc
```

---

## 19. Logging / monitoring security

### 19.1 Log không chứa bí mật

Không log:

- Password/token
- Full bank account nếu không cần
- Full salary data
- Sensitive personal data
- File signed URL dài hạn

### 19.2 Security alert

Cần alert cho:

- Login fail nhiều lần
- User được cấp quyền admin
- Export dữ liệu nhạy cảm
- Điều chỉnh tồn lớn
- Override QC
- Hủy nhiều đơn bất thường
- Nhiều scan lỗi khi bàn giao ĐVVC
- Hàng hoàn bị disposition bất thường
- Payment approval bất thường

---

## 20. Compliance baseline nội bộ

ERP Phase 1 phải hỗ trợ các yêu cầu kiểm soát sau:

### 20.1 Traceability mỹ phẩm

- Từ đơn bán truy batch.
- Từ batch truy nguyên liệu/nhà máy/NCC nếu dữ liệu có.
- Batch FAIL/HOLD không được bán.
- Hàng hoàn phải qua kiểm tra trước khi quay lại tồn bán.

### 20.2 Financial control

- PO, nhận hàng, hóa đơn, thanh toán phải có dấu vết.
- Người tạo và người duyệt tách nhau.
- Payment không tự mở nếu chưa đủ nghiệm thu/QC/biên bản.

### 20.3 HR/privacy

- Dữ liệu cá nhân, bảng lương, thông tin ngân hàng chỉ role được cấp quyền xem.
- Offboarding phải revoke quyền ngay.

### 20.4 Audit readiness

Khi audit nội bộ, phải xuất được:

- Ai duyệt PO?
- Ai nhận hàng?
- Ai QC pass batch?
- Ai xuất kho?
- Ai bàn giao ĐVVC?
- Hàng hoàn được xử lý thế nào?
- Vì sao điều chỉnh tồn?
- Ai xem/export dữ liệu nhạy cảm?

---

## 21. Break-glass access

Break-glass là quyền khẩn cấp khi hệ thống/nghiệp vụ cần xử lý gấp.

### 21.1 Khi nào dùng

- Lỗi hệ thống chặn xuất hàng nghiêm trọng
- Cần sửa dữ liệu để phục hồi sau incident
- Admin chính nghỉ/không truy cập được
- Production outage

### 21.2 Điều kiện

- Chỉ user được chỉ định
- Yêu cầu MFA
- Yêu cầu lý do
- Có thời hạn ngắn
- Audit log dày
- Sau khi dùng phải review

### 21.3 Không dùng break-glass để

- Bỏ qua quy trình thường ngày
- Duyệt thanh toán cá nhân
- Sửa dữ liệu cho tiện
- Che lỗi vận hành

---

## 22. Security incident response

### 22.1 Mức độ sự cố

| Severity | Ví dụ | Thời gian phản ứng |
|---|---|---:|
| SEV1 | Lộ dữ liệu nhạy cảm, user trái quyền, sửa tồn nghiêm trọng | Ngay lập tức |
| SEV2 | Sai quyền module, audit thiếu, export bất thường | Trong ngày |
| SEV3 | Lỗi permission nhỏ, UI hiển thị sai field nhưng API chặn | 1–3 ngày |
| SEV4 | Cải tiến bảo mật | Theo sprint |

### 22.2 Playbook cơ bản

```text
1. Phát hiện
2. Cô lập user/session/API nếu cần
3. Giữ bằng chứng log
4. Đánh giá phạm vi ảnh hưởng
5. Khôi phục/quay ngược giao dịch nếu cần
6. Báo cáo owner
7. Fix root cause
8. Postmortem
9. Update rule/test
```

### 22.3 Ví dụ incident

#### User trái quyền xem giá vốn

```text
Disable quyền/role liên quan
→ kiểm audit ai đã xem/export
→ đánh giá dữ liệu bị lộ
→ sửa field-level permission
→ thêm test permission
→ báo cáo CEO/Finance
```

#### Scan bàn giao ĐVVC bị override sai

```text
Freeze manifest
→ kiểm scan log
→ đối chiếu hàng tại khu đóng hàng/bàn giao
→ chỉnh trạng thái bằng transaction reversal nếu cần
→ audit owner
→ update SOP/permission nếu lỗi người dùng
```

#### QC PASS nhầm batch lỗi

```text
Set batch HOLD khẩn cấp
→ block sales/reserve/xuất
→ trace đơn liên quan
→ CSKH/QA xử lý
→ CAPA
→ review quyền QA/re-auth
```

---

## 23. Security testing checklist

### 23.1 Authentication

- Login đúng/sai
- Lockout sai nhiều lần
- Session timeout
- Logout revoke session
- User disabled không login được

### 23.2 Authorization

- User không role không vào được module
- Role kho không xem giá vốn
- Role sales không sửa tồn
- Role QC không thanh toán
- Role finance không QC pass batch
- Scope kho: user kho A không thao tác kho B nếu không có quyền

### 23.3 Field-level permission

- API không trả field nhạy cảm nếu không có quyền
- Export không chứa field bị cấm
- UI không hiển thị field nhạy cảm
- Direct API call vẫn bị chặn

### 23.4 Approval/SoD

- Người tạo không tự duyệt được
- Duyệt vượt ngưỡng cần đúng role
- Reject bắt buộc lý do
- Sửa field quan trọng reset approval

### 23.5 Audit

- Tạo/sửa/hủy chứng từ có log
- QC pass/fail có log
- Stock movement có log
- Handover scan có log
- Return disposition có log
- Role assignment có log

### 23.6 Security regression

Mỗi sprint phải chạy lại test quyền cho nghiệp vụ cốt lõi:

```text
Purchase
QC
Inventory
Sales
Shipping
Returns
Subcontract
Finance
```

---

## 24. Implementation backlog gợi ý

### P0 — phải có trước go-live

- Login/logout/session
- User/role/permission core
- RBAC middleware backend
- Scope permission theo kho/phòng ban cơ bản
- Audit log cho nghiệp vụ trọng yếu
- Field-level hide cost/finance/payroll
- Stock ledger immutable guard
- QC status guard
- SoD cơ bản cho approve
- Export permission
- File signed URL cơ bản

### P1 — rất nên có trong Phase 1

- MFA cho admin/finance/CEO
- Re-auth cho action nguy hiểm
- Idempotency key cho giao dịch quan trọng
- Security alert cơ bản
- Break-glass access
- Permission test automation

### P2 — sau go-live/hardening

- Advanced ABAC
- IP/device policy
- Data watermark export
- DLP/export monitoring
- SIEM integration
- Advanced anomaly detection

---

## 25. Definition of Done cho security feature

Một tính năng bảo mật chỉ được coi là xong khi:

- Có permission name chuẩn.
- Backend enforce permission, không chỉ frontend ẩn nút.
- Có test case quyền đúng/sai.
- Có audit log nếu action quan trọng.
- Có error response chuẩn.
- Có field masking nếu dữ liệu nhạy cảm.
- Có migration/config role nếu cần.
- Có cập nhật tài liệu Permission Matrix nếu thay đổi quyền.
- Có review từ Tech Lead + BA/Owner.

---

## 26. Open questions cần chốt

1. Có bắt buộc MFA cho CEO/Finance/Admin từ ngày đầu không?
2. Công ty có nhiều kho/chi nhánh/brand ở Phase 1 hay chỉ một số kho chính?
3. Có cần phân quyền theo kênh bán không?
4. Có cho user export Excel không, role nào được export?
5. Mức ngưỡng tiền cho PO/payment/discount/adjustment là bao nhiêu?
6. Có cần lưu ảnh/video hàng hoàn ngay Phase 1 không?
7. Có dùng máy quét/máy tính bảng chung ở kho không?
8. Có cần portal nhà máy/NCC/ĐVVC ở Phase 1 không, hay chỉ nội bộ cập nhật?
9. Dữ liệu bảng lương/HRM có vào Phase 1 không hay để Phase 2?
10. Chính sách retention log chính thức là bao lâu?

---

## 27. Chốt thiết kế

Phase 1 phải ưu tiên bảo vệ các điểm rủi ro thật:

```text
Stock
Batch/QC
Handover ĐVVC
Returns/Hàng hoàn
Subcontract manufacturing
Purchase/payment
Cost/finance data
Role/approval/audit
```

Nếu chỉ làm login và menu ẩn/hiện, đó không phải bảo mật ERP.  
Bảo mật ERP đúng nghĩa là: **người sai quyền không làm được, người đúng quyền làm gì cũng để lại dấu vết, và giao dịch sai có đường đảo/chữa mà không phá lịch sử.**
