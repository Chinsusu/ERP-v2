# Permission Matrix + Approval Matrix Phase 1
## Hệ thống ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm

**Mã tài liệu:** ERP-PERM-APPROVAL-P1  
**Phiên bản:** v1.0  
**Ngày:** 2026-04-23  
**Phạm vi:** Phase 1 của ERP  
**Liên kết tài liệu gốc:** ERP Blueprint v1 + PRD/SRS Phase 1  
**Mục đích:** Chốt quyền thao tác, quyền duyệt, nguyên tắc phân tách nhiệm vụ và các ngưỡng phê duyệt cho 6 module lõi.

---

## 1. Mục tiêu của tài liệu

Tài liệu này dùng để khóa 4 thứ mà dự án ERP thường mơ hồ nhất:

1. Ai được **xem** cái gì.
2. Ai được **tạo / sửa / submit** chứng từ gì.
3. Ai được **duyệt / từ chối / release** giao dịch nào.
4. Khi nào hệ thống phải **chặn**, khi nào phải **đẩy lên cấp cao hơn**.

Mục tiêu không phải làm cho hệ thống “nhiều quyền”.  
Mục tiêu là làm cho hệ thống **rõ quyền, khó gian lận, khó sửa tay, và vẫn chạy nhanh**.

---

## 2. Phạm vi tài liệu

Tài liệu này áp dụng cho **Phase 1** của ERP, gồm 6 module lõi:

1. Dữ liệu gốc (Master Data)
2. Mua hàng (Procurement)
3. QA/QC
4. Sản xuất (Production)
5. Kho hàng (Warehouse)
6. Bán hàng (Sales / OMS cơ bản)

Ngoài ra tài liệu cũng áp dụng cho:
- đăng nhập, phân quyền, workflow phê duyệt;
- audit log;
- dashboard và báo cáo tối thiểu;
- xử lý ngoại lệ có ảnh hưởng tới hàng, tiền, batch, chất lượng.

---

## 3. Nguyên tắc phân quyền

1. **Người tạo không tự duyệt** cùng một giao dịch.
2. **Không hard delete** chứng từ đã được submit/approve hoặc đã phát sinh ảnh hưởng kho.
3. **QC quyết định trạng thái chất lượng**, không phải kho hay sales.
4. **Kho quyết định thao tác vật lý**, không phải QC hay sales.
5. **Sales không tự phá giá** và không tự vượt công nợ nếu không có duyệt.
6. **Admin hệ thống không được sửa số liệu nghiệp vụ trực tiếp** trừ khi có quy trình break-glass và audit log.
7. **Tất cả ngoại lệ phải để lại dấu vết**: ai làm, lúc nào, giá trị trước/sau, lý do.
8. **Quyền xem báo cáo không đồng nghĩa với quyền sửa dữ liệu.**
9. **Các trường trọng yếu phải khóa sau khi phát sinh giao dịch đầu tiên**, ví dụ: item type, base UOM, batch-control flag, valuation category, BOM active version.
10. **Mọi quyền đặc biệt đều có thời hạn và log đầy đủ**.

---

## 4. Danh sách vai trò trong Phase 1

| Mã | Vai trò | Mô tả ngắn |
|---|---|---|
| SA | System Admin | Quản trị user, role, cấu hình workflow, numbering, bảo mật |
| MDA | Master Data Admin | Quản lý item, SKU, BOM, supplier, customer, warehouse, price list draft |
| POF | Purchasing Officer | Tạo PR, PO draft, theo dõi nhận hàng, làm việc với NCC |
| PUM | Purchasing Manager | Duyệt nghiệp vụ mua hàng, duyệt supplier và giao dịch mua theo ngưỡng |
| QCO | QC Officer | Tạo phiếu kiểm, nhập kết quả QC, ghi nhận NCR |
| QAM | QA Manager | Duyệt pass/fail/hold, release batch, duyệt các ngoại lệ chất lượng |
| PPL | Production Planner | Lập kế hoạch sản xuất, tạo Production Order draft |
| PSV | Production Supervisor | Xác nhận triển khai lệnh, sản lượng, hao hụt, hoàn thành lệnh |
| WST | Warehouse Staff | Nhập, xuất, chuyển kho, kiểm kê theo thao tác vật lý |
| WMG | Warehouse Manager | Duyệt điều chỉnh kho, duyệt chuyển kho, kiểm soát sai lệch |
| SAD | Sales Admin | Tạo quotation, sales order, delivery order, sales return request |
| SMG | Sales Manager | Duyệt discount, duyệt ngoại lệ bán hàng, duyệt chính sách khách hàng |
| FIN | Finance Approver | Kiểm soát công nợ, giá trị giao dịch, credit limit, refund, write-off |
| EXE | CEO / COO / Head được ủy quyền | Duyệt giao dịch vượt ngưỡng cao hoặc ngoại lệ chiến lược |

> Ghi chú: nếu công ty có thêm vai trò “R&D Manager”, “Plant Manager”, “Retail Ops”, “Customer Service Manager”, có thể bổ sung ở phase sau hoặc map tạm vào `QAM`, `PSV`, `SMG`, `EXE` tùy mô hình tổ chức.

---

## 5. Chú giải ký hiệu quyền

| Ký hiệu | Ý nghĩa |
|---|---|
| R | Read / View / Search / Export |
| O | Operate: tạo, sửa bản nháp, submit |
| A | Approve / Reject / Release / Close trong phạm vi được phép |
| X | Execute giao dịch có ảnh hưởng vận hành thực tế, ví dụ xác nhận nhập/xuất/đếm/issue/receipt |
| CFG | Configure / Admin |
| - | Không có quyền |

**Lưu ý quan trọng:**
- `O` **không** có nghĩa là được tự kích hoạt hiệu lực cuối cùng nếu giao dịch cần approval.
- `A` **không** có nghĩa là được sửa dữ liệu chi tiết vật lý sau khi giao dịch đã post.
- `X` chỉ dành cho các bước xác nhận **thực tế vận hành** như nhập kho, xuất kho, đếm kho, issue nguyên liệu, nhận thành phẩm.
- Với chứng từ đã `Approved` hoặc `Posted`, chỉ được `Cancel / Reverse / Return` theo rule; không được sửa trực tiếp.

---

## 6. Permission Matrix

### 6.1. Governance & Security

| Đối tượng / Màn hình | SA | MDA | FIN | EXE |
|---|---|---|---|---|
| User / Role / Permission | CFG | R | R | R |
| Approval Workflow Config | CFG | R | R | R |
| Numbering Rules | CFG | R | - | R |
| Audit Log | R | R | R | R |
| Integration / Import Template | CFG | R | R | R |
| Dashboard hệ thống / lỗi đồng bộ / job log | R | R | R | R |

**Rule bắt buộc:**
- Chỉ `SA` được cấu hình quyền và workflow.
- `SA` không được tự sửa số liệu nghiệp vụ trong database bằng cách bypass UI.
- Mọi thay đổi role/quyền phải đi qua quy trình phê duyệt riêng trong mục 7.2.

---

### 6.2. Master Data

| Đối tượng / Màn hình | SA | MDA | PUM | QAM | PPL | WMG | SAD | SMG | FIN | EXE |
|---|---|---|---|---|---|---|---|---|---|---|
| Item / SKU Master | R | O | A* | A* | A* | R | R | R | R | R |
| Supplier Master | R | O | A | R | R | R | - | - | A* | R |
| Customer Master | R | O | - | - | - | - | O | A | A* | R |
| Warehouse / Location Master | R | O | R | R | R | A | - | - | R | R |
| BOM / Formula Version (Phase 1 basic) | R | O | R | A* | A* | - | - | - | R | R |
| Price List / Discount Policy Draft | R | O | - | - | - | - | O | A | A* | R |
| Credit Limit / Payment Term Draft | R | O | - | - | - | - | O | A | A | R |
| Batch Number Rule / Expiry Rule | CFG | O | R | A | A | R | - | - | R | R |

`A*` là quyền duyệt có điều kiện, chi tiết nằm ở **Approval Matrix**.

**Nguyên tắc khóa trường trọng yếu:**
- Sau khi item phát sinh giao dịch đầu tiên, không cho sửa trực tiếp các trường: `item_type`, `base_uom`, `batch_control_flag`, `expiry_control_flag`, `valuation_method`, `tax_group`, `inventory_account_map`.
- Thay đổi BOM active version phải có hiệu lực theo ngày/giờ, không thay đè lịch sử.
- Supplier payment term, customer credit limit và price list phải lưu lịch sử hiệu lực.

---

### 6.3. Procurement

| Đối tượng / Màn hình | POF | PUM | WST | WMG | QCO | QAM | FIN | EXE |
|---|---|---|---|---|---|---|---|---|
| Purchase Requisition (PR) | O | A | R | - | - | - | R | A* |
| Purchase Order (PO) | O | A | R | - | - | - | A* | A* |
| Receiving / GRN | R | R | O,X | A | R | R | R | R |
| Supplier Return Request | O | A | O,X | A | O | A* | A* | A* |
| Open PO / Inbound Report | R | R | R | R | R | R | R | R |
| Supplier Performance Report | R | A | - | - | R | R | R | R |

**Rule vận hành:**
- `WST` xác nhận số lượng nhận thực tế, nhưng không quyết định pass/fail chất lượng.
- `QCO/QAM` quyết định chất lượng inbound, nhưng không sửa số lượng nhận hàng thực tế.
- `POF` có thể tạo/sửa draft PO trước khi submit; không tự duyệt PO của chính mình.
- `GRN` sau khi post không sửa tay; chỉ được reverse theo rule.

---

### 6.4. QA / QC

| Đối tượng / Màn hình | QCO | QAM | POF | PPL | PSV | WST | WMG | SAD | SMG | FIN | EXE |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Inbound QC (IQC) | O | A | R | - | - | R | R | - | - | R | R |
| FG / Production QC | O | A | - | R | R | R | R | R | R | R | R |
| Hold / Pass / Fail / Release | O | A | R | R | R | R | R | R | R | R | R |
| NCR / Deviation Log | O | A | R | R | R | R | R | - | - | R | R |
| Batch Trace / Complaint Trace | R | R | R | R | R | R | R | R | R | R | R |

**Rule vận hành:**
- Chỉ `QAM` được release batch từ `On Hold` sang `Passed`.
- `Sales` chỉ được nhìn trạng thái QC; không được sửa hoặc bypass.
- `Warehouse` chỉ được xuất bán batch có trạng thái `Passed` và còn hạn hợp lệ.
- Bất kỳ batch nào bị `Failed` phải chuyển sang tồn cách ly / write-off flow; không được nhập lại tồn khả dụng bằng tay.

---

### 6.5. Production

| Đối tượng / Màn hình | PPL | PSV | WST | WMG | QCO | QAM | FIN | EXE |
|---|---|---|---|---|---|---|---|---|
| Production Plan | O | A* | R | - | - | - | R | A* |
| Production Order (MO/WO) | O | A,X | R | - | R | R | R | A* |
| Material Issue to Production | R | A | O,X | R | - | R | R | R |
| Finished Goods Receipt | R | O,X | O,X | R | R | R | R | R |
| Scrap / Loss Declaration | O | A | R | R | R | R | A* | A* |
| Production Progress Board | R | R | R | R | R | R | R | R |

**Rule vận hành:**
- `PPL` tạo lệnh; `PSV` xác nhận bắt đầu/chạy/hoàn thành.
- `WST` là người xác nhận xuất NVL vật lý khỏi kho và có thể tham gia nhập thành phẩm vật lý về kho.
- Hao hụt/scrap vượt tolerance phải xin duyệt theo matrix.
- Không cho đổi BOM active trực tiếp trên lệnh đã bắt đầu chạy.

---

### 6.6. Warehouse

| Đối tượng / Màn hình | WST | WMG | QAM | PPL | PSV | SAD | SMG | FIN | EXE |
|---|---|---|---|---|---|---|---|---|---|
| Stock Inquiry / Batch Inquiry | R | R | R | R | R | R | R | R | R |
| Transfer Order | O,X | A,X | R | R | R | R | R | R | A* |
| Stock Count | O,X | A,X | - | - | - | - | - | R | R |
| Stock Adjustment / Write-off Request | O | A | R | - | - | - | - | A* | A* |
| Reserve / Unreserve Stock | R | R | - | - | - | O | A | R | R |
| Near Expiry / Quarantine Report | R | R | R | R | R | R | R | R | R |

**Rule vận hành:**
- `WST` thao tác vật lý; `WMG` duyệt chênh lệch và điều chỉnh.
- Không cho `SAD` hoặc `SMG` tự điều chỉnh kho.
- `Reserve stock` chỉ là giữ tồn logic; không thay thế phiếu xuất kho.
- Hàng `Hold`, `Failed`, `Expired`, `Quarantine` không được tính vào tồn khả dụng.

---

### 6.7. Sales / OMS cơ bản

| Đối tượng / Màn hình | SAD | SMG | WST | WMG | QAM | FIN | EXE |
|---|---|---|---|---|---|---|---|
| Quotation | O | A* | R | - | - | R | R |
| Sales Order (SO) | O | A | R | R | - | A* | A* |
| Delivery Order (DO) | O | A* | O,X | A,X | R* | R | R |
| Sales Return Request | O | A | O,X | A,X | A* | A* | A* |
| Credit Limit Override Request | O | A | - | - | - | A | A* |
| Sales / Open Order Report | R | R | R | R | R | R | R |

`QAM` chỉ tham gia khi giao dịch liên quan batch ngoại lệ, quality complaint, cận date hoặc hàng trả nghi vấn lỗi chất lượng.

**Rule vận hành:**
- SO chỉ được auto-confirm khi **đúng bảng giá**, **đúng chính sách discount**, **không vượt credit**, **không vượt tồn khả dụng**.
- Nếu vượt một trong các điều kiện trên, hệ thống phải chuyển `Pending Approval`.
- `WST/WMG` là bên xác nhận xuất hàng thực tế; sales không tự post goods issue.
- Sales return không được nhập thẳng vào tồn khả dụng nếu chưa qua phân loại tình trạng hàng.

---

### 6.8. Dashboard & Reporting

| Dashboard / Báo cáo | SA | MDA | PUM | QAM | PPL | PSV | WMG | SAD | SMG | FIN | EXE |
|---|---|---|---|---|---|---|---|---|---|---|---|
| Dashboard CEO / COO | R | R | R | R | R | R | R | R | R | R | R |
| Open PR / PO / Inbound | R | R | R | R | - | - | R | - | - | R | R |
| QC Status / Hold / Release | R | R | R | R | R | R | R | R | R | R | R |
| Production Status / Yield / Scrap | R | - | - | R | R | R | R | - | - | R | R |
| Inventory Health / Near Expiry | R | R | R | R | R | R | R | R | R | R | R |
| Open SO / Delivery / Return | R | - | - | R | - | - | R | R | R | R | R |
| Export master / transaction data | R | R | R | R | R | R | R | R | R | R | R |

**Rule báo cáo:**
- Quyền `R` cho báo cáo không mặc định đồng nghĩa được xem toàn bộ giá vốn/margin chi tiết.
- Báo cáo nhạy cảm như margin floor, write-off value, override history nên ẩn theo role hoặc mask một phần dữ liệu.

---

## 7. Approval Matrix

### 7.1. Ngưỡng cấu hình gợi ý để khởi tạo

> Đây là **gợi ý khởi tạo** để team triển khai có số mà cấu hình thử. Trước khi go-live, công ty phải xác nhận lại theo quy mô doanh thu, giá trị hàng tồn và mức chấp nhận rủi ro.

| Mã ngưỡng | Gợi ý khởi tạo | Áp dụng |
|---|---|---|
| PO-A | ≤ 20.000.000 VND | PO mức thấp |
| PO-B | > 20.000.000 đến 100.000.000 VND | PO mức trung bình |
| PO-C | > 100.000.000 VND | PO mức cao |
| ADJ-A | ≤ 1.000.000 VND hoặc ≤ 0,5% giá trị line | Chênh lệch/điều chỉnh kho nhỏ |
| ADJ-B | > 1.000.000 đến 10.000.000 VND hoặc > 0,5% đến 2% | Chênh lệch/điều chỉnh kho trung bình |
| ADJ-C | > 10.000.000 VND hoặc > 2% | Chênh lệch/điều chỉnh kho lớn |
| DIS-A | Discount trong policy hoặc ≤ 5% | Ngoại lệ bán hàng thấp |
| DIS-B | > 5% đến 10% | Ngoại lệ bán hàng trung bình |
| DIS-C | > 10% | Ngoại lệ bán hàng cao |
| RET-A | Giá trị trả hàng ≤ 5.000.000 VND, không liên quan quality complaint | Return mức thấp |
| RET-B | > 5.000.000 đến 20.000.000 VND hoặc có nghi vấn batch | Return mức trung bình |
| RET-C | > 20.000.000 VND hoặc có rủi ro recall / truyền thông | Return mức cao |
| CR-A | Vượt credit limit ≤ 10% hoặc quá hạn ≤ 7 ngày | Ngoại lệ tín dụng thấp |
| CR-B | Vượt credit > 10% hoặc quá hạn > 7 ngày | Ngoại lệ tín dụng cao |
| MO-A | Lệnh sản xuất nằm trong kế hoạch tuần đã duyệt | MO tiêu chuẩn |
| MO-B | Lệnh sản xuất unplanned / gấp / yêu cầu OT / dùng batch đặc biệt | MO ngoại lệ |

---

### 7.2. Ma trận phê duyệt giao dịch chính

| Giao dịch / Yêu cầu | Điều kiện kích hoạt | Người submit | Bước 1 | Bước 2 | Bước 3 / Escalation | Ghi chú |
|---|---|---|---|---|---|---|
| Tạo mới Item Raw/Packaging | Item mới, chưa dùng giao dịch | MDA | PUM | QAM* | - | `QAM` bắt buộc nếu item có kiểm QC, hạn dùng, batch control |
| Tạo mới FG / SKU bán | SKU thành phẩm mới | MDA / PPL | PPL | QAM | EXE* | `EXE` chỉ khi SKU chiến lược, launch lớn hoặc ảnh hưởng giá trị tồn đáng kể |
| Thay đổi trường trọng yếu của item | UOM, batch flag, expiry flag, valuation, tax group | MDA | Owner function | FIN* | EXE* | Không sửa tay trực tiếp; dùng change request có effective date |
| Kích hoạt / đổi BOM active version | BOM mới hoặc đổi định mức | MDA / PPL | PPL | QAM | EXE* | `EXE` chỉ khi thay đổi ảnh hưởng cost lớn, launch gấp, hoặc ngoại lệ |
| Tạo mới supplier | Supplier mới | MDA / POF | PUM | FIN* | - | `FIN` bắt buộc nếu có payment term/credit term đặc biệt |
| Sửa payment term supplier | Đổi hạn thanh toán / đặt cọc / công nợ | POF / MDA | PUM | FIN | EXE* | `EXE` khi vượt policy mua hàng |
| Tạo mới customer / credit account | Khách công nợ mới | SAD / MDA | SMG | FIN | - | Khách COD không cần `FIN` nếu credit limit = 0 |
| Thay đổi price list / margin floor | Giá bán, discount chuẩn, channel price | SAD / MDA | SMG | FIN | EXE* | `EXE` khi dưới margin floor chiến lược hoặc deal đặc biệt |
| PR | Mọi PR hoặc theo rule department | POF / requester | PUM hoặc Head bộ phận | EXE* | - | `EXE` cho mua ngoài ngân sách / khẩn cấp / CAPEX |
| PO mức thấp | PO-A, NCC đã duyệt | POF | PUM | - | - | Single approval |
| PO mức trung bình | PO-B hoặc supplier mới / điều khoản mới | POF | PUM | FIN | - | Double approval |
| PO mức cao | PO-C hoặc mua chiến lược / rủi ro cao | POF | PUM | FIN | EXE | Triple approval |
| Nhận hàng sai lệch với PO | Over/short > tolerance cấu hình | WST | WMG | POF / PUM | FIN* | `FIN` khi sai lệch dẫn tới thay đổi giá trị thanh toán đáng kể |
| Supplier Return | Trả NCC do lỗi / dư / sai batch | QCO / WST / POF | PUM | FIN* | EXE* | `FIN` khi phát sinh debit/credit note đáng kể; `EXE` khi ảnh hưởng lớn |
| Inbound QC pass / fail / hold | Mọi lô inbound cần QC | QCO | QAM | - | EXE notify* | `EXE` chỉ nhận thông báo nếu fail lô chiến lược/khối lượng lớn |
| FG QC release | Mọi batch FG trước khi bán | QCO | QAM | - | EXE notify* | Batch không pass thì không được bán |
| Production Order tiêu chuẩn | MO-A | PPL | PSV | - | - | Cho phép chạy nếu đúng kế hoạch đã duyệt và đủ vật tư |
| Production Order ngoại lệ | MO-B | PPL | PSV | EXE | - | Dùng cho lệnh gấp, chạy ngoài kế hoạch, OT, hoặc dùng batch đặc biệt |
| Issue thêm NVL vượt BOM tolerance | Actual > standard vượt tolerance | PSV | PPL | QAM* | EXE* | `QAM` khi ảnh hưởng chất lượng/công thức; `EXE` nếu vượt ngưỡng giá trị |
| Scrap / Loss Declaration thấp | ≤ tolerance và ≤ ADJ-A | PSV | WMG | - | - | Vẫn cần ít nhất 1 cấp duyệt ngoài người khai để giữ nguyên tắc phân tách nhiệm vụ |
| Scrap / Loss Declaration trung bình | > tolerance hoặc ADJ-B | PSV | WMG | FIN | - | Double approval |
| Scrap / Loss Declaration cao | ADJ-C hoặc nghi vấn thất thoát | PSV | WMG | FIN | EXE | Triple approval |
| Transfer nội bộ cùng site | Kho cùng site, không đổi ownership | WST | WMG | - | - | Single approval |
| Transfer liên kho / liên chi nhánh | Giá trị cao, thay đổi ownership/địa điểm xa | WST | WMG | EXE* | - | `EXE` nếu vượt ngưỡng giá trị hoặc rủi ro |
| Stock Count variance nhỏ | ADJ-A | WST | WMG | - | - | Sau approval mới được post adjustment |
| Stock Count variance trung bình | ADJ-B | WST | WMG | FIN | - | Double approval |
| Stock Count variance lớn / Write-off | ADJ-C hoặc nghi ngờ thất thoát | WST | WMG | FIN | EXE | Triple approval |
| Quotation special price / special term | Quote ngoài policy, validity dài, hoặc deal riêng | SAD | SMG | FIN* | EXE* | Có thể dùng cùng ngưỡng DIS-A/B/C nếu quote đã khóa giá bán |
| SO auto-confirm | Đúng price list, discount trong policy, đủ stock, không vượt credit | SAD | System Auto | - | - | Không cần duyệt tay |
| SO discount mức thấp | DIS-A nhưng ngoài policy chuẩn | SAD | SMG | - | - | Single approval |
| SO discount mức trung bình | DIS-B | SAD | SMG | FIN* | - | `FIN` nếu ảnh hưởng margin floor |
| SO discount mức cao | DIS-C | SAD | SMG | FIN | EXE | Triple approval |
| SO vượt credit mức thấp | CR-A | SAD | SMG | FIN | - | Double approval |
| SO vượt credit mức cao | CR-B | SAD | SMG | FIN | EXE | Triple approval |
| DO / batch allocation ngoại lệ | Chọn batch gần hết hạn, split batch bất thường, manual override | SAD / WST | WMG | QAM* | SMG* | `QAM` khi liên quan chất lượng/cận date; `SMG` khi ảnh hưởng cam kết với khách |
| Sales Return mức thấp | RET-A | SAD | SMG | WMG | - | Cần cả xác nhận thương mại và nhận hàng thực tế |
| Sales Return mức trung bình | RET-B hoặc nghi vấn batch | SAD | SMG | WMG | QAM / FIN* | `QAM` nếu nghi lỗi chất lượng; `FIN` nếu refund/credit note đáng kể |
| Sales Return mức cao | RET-C hoặc recall risk | SAD | SMG | WMG | QAM + FIN + EXE | Escalation đầy đủ |
| Credit Note / Refund | Refund tiền cho khách | SAD / FIN | SMG | FIN | EXE* | `EXE` khi refund cao hoặc ngoại lệ chính sách |
| User / Role / Permission Change | Thêm role, đổi approver, mở quyền nhạy cảm | SA | Head bộ phận / Owner module | FIN* | EXE | `FIN` bắt buộc nếu quyền mới cho phép xem/sửa dữ liệu tiền, margin, refund |
| Break-glass data fix | Sửa lỗi dữ liệu sau go-live | SA | Owner module | FIN* | EXE | Chỉ dùng khi không thể reverse bằng quy trình chuẩn |

---

### 7.3. Quy tắc auto-approval và auto-reject

#### Auto-approval
Hệ thống có thể tự động confirm hoặc bỏ qua bước duyệt tay nếu đồng thời thỏa tất cả điều kiện sau:

1. Giao dịch thuộc loại cho phép auto-approval trong cấu hình.
2. Giá trị nằm trong ngưỡng thấp.
3. Đúng master data đã được duyệt.
4. Không vi phạm rule credit, tồn kho, QC status, batch status, margin floor.
5. Người submit không nằm trong danh sách bị hạn chế quyền tạm thời.

#### Auto-reject / system block
Hệ thống phải tự chặn ngay, không cho submit hoặc không cho post nếu xảy ra các trường hợp:

1. Batch `Hold`, `Failed`, `Expired`, `Quarantine` bị chọn để bán/xuất.
2. SO vượt tồn khả dụng mà không có quyền override hợp lệ.
3. PO dùng supplier chưa active.
4. Production Order dùng BOM không active.
5. Material Issue vượt quá tolerance mà chưa có approval chain.
6. Sales Return nhập vào tồn khả dụng khi chưa qua disposition.
7. User cố sửa chứng từ đã `Posted` ngoài rule reverse/cancel.

---

### 7.4. SLA xử lý phê duyệt đề xuất

| Loại phê duyệt | SLA đề xuất |
|---|---|
| PR / PO mức thấp | 4 giờ làm việc |
| PO mức trung bình / thay đổi supplier term | 8 giờ làm việc |
| PO mức cao / mua khẩn | 24 giờ hoặc theo rule khẩn |
| QC release batch | 4 giờ sau khi có đủ kết quả |
| Production Order ngoại lệ | 2 giờ làm việc |
| Stock adjustment / write-off | 8 giờ làm việc |
| SO discount / credit override | 2 giờ làm việc |
| Sales Return / Refund | 8 giờ làm việc |
| Break-glass data fix | Ngay khi có đủ 3 chữ ký và ticket sự cố |

---

## 8. Separation of Duties (SoD) – phân tách nhiệm vụ bắt buộc

| Tình huống | Bắt buộc phân tách |
|---|---|
| Tạo item/BOM và duyệt kích hoạt | Người tạo draft không là người duyệt cuối |
| Tạo PO và duyệt PO | `POF` không tự duyệt PO của chính mình |
| Nhận hàng và QC | `WST` không tự kết luận pass/fail chất lượng |
| QC và điều chỉnh kho | `QCO/QAM` không tự post stock adjustment thay cho kho |
| Tạo MO và xác nhận hoàn thành | `PPL` không tự xác nhận completion thay `PSV` |
| Tạo SO và duyệt discount/credit | `SAD` không tự duyệt ngoại lệ bán hàng |
| Kiểm kê và duyệt chênh lệch kho | Người đếm kho không là người duyệt cuối cùng nếu có variance |
| Tạo sales return và quyết định disposition | `SAD` không tự quyết định đưa hàng trả lại về tồn khả dụng |
| Cấu hình quyền và phê duyệt break-glass | `SA` không tự cấp quyền đặc biệt cho chính mình |
| Refund/Credit note và approve payment | Người yêu cầu refund không là người duyệt thanh toán cuối |

**Nguyên tắc vàng:**  
**Một người có thể làm nhanh hơn khi ôm nhiều bước, nhưng hệ thống không được thiết kế theo cách đó.**  
Tốc độ có thể tối ưu bằng SLA và auto-rule; không tối ưu bằng bỏ kiểm soát.

---

## 9. Field-level restriction – khóa ở mức trường dữ liệu

### 9.1. Item / SKU
Các trường sau khóa sau giao dịch đầu tiên:
- item code
- item type
- base UOM
- batch control flag
- expiry control flag
- inventory valuation group
- active ingredient flag
- default QC required flag

### 9.2. Supplier / Customer
Các trường sau phải lưu version lịch sử:
- payment term
- credit limit
- tax code
- sales channel group
- pricing group
- account status

### 9.3. Transaction
Các trường sau không được sửa trực tiếp sau khi `Approved` hoặc `Posted`:
- qty
- unit price
- warehouse
- batch
- expiry date
- supplier/customer
- linked reference number
- approval trail

---

## 10. Audit log & bằng chứng bắt buộc

Hệ thống phải lưu tối thiểu:

1. Người tạo, người sửa cuối, người submit, người duyệt.
2. Thời gian thao tác.
3. Giá trị trước/sau của các trường trọng yếu.
4. Lý do reject / cancel / reverse / override.
5. File đính kèm nếu có:
   - báo giá NCC,
   - COA,
   - biên bản QC,
   - biên bản kiểm kê,
   - ảnh hàng lỗi,
   - xác nhận với khách hàng,
   - phiếu yêu cầu write-off / refund.

**Không có audit log đủ sâu, hệ thống không đủ mạnh để xử lý tranh chấp nội bộ.**

---

## 11. Quy tắc break-glass / quyền khẩn cấp

Break-glass là quyền dùng trong tình huống hiếm gặp, ví dụ:
- dữ liệu lỗi do migration,
- lỗi hệ thống khiến không reverse được bằng quy trình chuẩn,
- batch bị khóa nhầm làm tắc vận hành,
- cần mở tạm một quyền trong thời gian ngắn để xử lý sự cố.

### Quy tắc
1. Phải có ticket sự cố.
2. Phải có mô tả nguyên nhân và phương án rollback.
3. Phải có approval theo matrix:
   - SA đề xuất
   - Owner module xác nhận
   - FIN nếu ảnh hưởng giá trị
   - EXE duyệt cuối
4. Quyền đặc biệt có thời hạn tối đa 24 giờ, hoặc cho đúng 1 giao dịch.
5. Sau khi xử lý xong phải đóng ticket và xuất audit report.

---

## 12. Ma trận menu hiển thị gợi ý theo vai trò

| Menu / Module | SA | MDA | POF | PUM | QCO | QAM | PPL | PSV | WST | WMG | SAD | SMG | FIN | EXE |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| Security / User / Workflow | ✓ | - | - | - | - | - | - | - | - | - | - | - | ✓ (view) | ✓ (view) |
| Master Data | ✓ | ✓ | view | view | view | view | view | - | view | view | view | view | view | view |
| Procurement | view | view | ✓ | ✓ | view | view | - | - | view | view | - | - | view | view |
| QA / QC | view | view | view | view | ✓ | ✓ | view | view | view | view | view | view | view | view |
| Production | view | view | - | - | view | view | ✓ | ✓ | view | view | - | - | view | view |
| Warehouse | view | view | view | view | view | view | view | view | ✓ | ✓ | view | view | view | view |
| Sales / OMS | view | view | - | - | view | view | - | - | view | view | ✓ | ✓ | ✓ | view |
| Dashboard / BI | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## 13. Checklist chốt trước khi cấu hình hệ thống

Trước khi BA/PM cho dev cấu hình quyền và workflow, công ty phải chốt:

1. Danh sách user thực tế theo phòng ban.
2. Ai là approver chính / approver dự phòng cho từng flow.
3. Ngưỡng PO, stock adjustment, discount, credit override, refund.
4. Tolerance cho:
   - receiving variance,
   - BOM over-consumption,
   - stock count variance.
5. Margin floor theo kênh bán.
6. Rule auto-approval nào được bật ngay từ go-live.
7. Người nào được quyền break-glass.
8. Danh sách báo cáo nào bị giới hạn xem giá trị tiền.

---

## 14. Khuyến nghị triển khai

1. Cấu hình quyền theo **role**, không cấp quyền thủ công cho từng user trừ ngoại lệ.
2. Tạo môi trường UAT riêng để test đầy đủ các case quyền và duyệt.
3. Chạy thử 10–20 kịch bản UAT bắt buộc:
   - PO vượt ngưỡng,
   - inbound fail QC,
   - MO unplanned,
   - batch hold không cho bán,
   - stock count variance lớn,
   - SO vượt credit,
   - sales return nghi lỗi batch,
   - break-glass data fix.
4. Trước go-live phải có danh sách người thay thế approver khi nghỉ phép.
5. Sau go-live 2 tuần đầu nên bật audit review hàng ngày cho:
   - stock adjustment,
   - QC override,
   - discount override,
   - credit override,
   - refund / write-off.

---

## 15. Kết luận

Permission Matrix và Approval Matrix không phải phần “phụ” của ERP.  
Đây là thứ quyết định hệ thống của mày là:

- **một bộ máy điều hành có kỷ luật**, hoặc
- **một admin panel đẹp nhưng ai cũng lách được**.

Nếu phải ưu tiên một nguyên tắc duy nhất, hãy dùng nguyên tắc này:

> **Người tạo không tự duyệt. Người thao tác vật lý không tự chốt chất lượng. Người nhìn thấy tiền phải nhìn được ngoại lệ. Và mọi quyền đặc biệt phải để lại dấu vết.**

---