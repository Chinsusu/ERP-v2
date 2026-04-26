# 30. ERP Data Governance & Change Control - Phase 1 - Mỹ phẩm

**Tên file:** `30_ERP_Data_Governance_Change_Control_Phase1_MyPham_v1.md`  
**Phiên bản:** v1.0  
**Phạm vi:** ERP Phase 1  
**Doanh nghiệp:** Công ty sản xuất, phân phối, bán lẻ mỹ phẩm  
**Kiến trúc liên quan:** Go Backend + PostgreSQL + React/Next.js + Modular Monolith  
**Tài liệu liên quan:** 01-29 trong bộ ERP Phase 1  

---

## 1. Mục tiêu tài liệu

Tài liệu này khóa 2 lớp cực kỳ quan trọng sau go-live:

1. **Data Governance** - ai được tạo, sửa, khóa, duyệt, dùng và chịu trách nhiệm cho dữ liệu.
2. **Change Control** - mọi thay đổi nghiệp vụ, dữ liệu, quyền hạn, màn hình, API, quy trình, tích hợp, báo cáo phải đi theo một quy trình kiểm soát.

Nói dễ hiểu: ERP không chỉ cần build đúng. ERP phải **sống sạch** sau khi chạy thật.

Nếu không có tài liệu này, sau 3-6 tháng hệ thống rất dễ biến thành:

- dữ liệu trùng
- mã hàng loạn
- giá sai
- tồn kho lệch
- batch không truy được
- chứng từ bị sửa tay
- quyền cấp lung tung
- yêu cầu thay đổi nhắn miệng qua Zalo
- dev sửa gấp không có log
- báo cáo mỗi phòng hiểu một kiểu

Một câu chốt:

> **ERP không chết ngay vì thiếu tính năng. ERP chết dần vì dữ liệu bẩn và thay đổi vô kỷ luật.**

---

## 2. Bối cảnh nghiệp vụ thực tế

Tài liệu này bám vào workflow thực tế hiện có của công ty:

- Kho có nhịp công việc hằng ngày: tiếp nhận đơn trong ngày, thực hiện xuất/nhập, soạn và đóng gói, sắp xếp kho, kiểm kê cuối ngày, đối soát số liệu, báo cáo quản lý và kết thúc ca.
- Nội quy kho chia rõ quy trình nhập kho, xuất kho, đóng hàng và xử lý hàng hoàn.
- Bàn giao cho đơn vị vận chuyển có phân khu để hàng, để theo thùng/rổ, đối chiếu số lượng, lấy hàng, quét mã, xử lý đủ/chưa đủ đơn và ký xác nhận bàn giao.
- Sản xuất hiện có nhánh gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển nguyên vật liệu/bao bì, duyệt mẫu, sản xuất hàng loạt, giao về kho, kiểm tra số lượng/chất lượng, nhận hàng hoặc báo lỗi nhà máy trong 3-7 ngày.

Vì vậy governance không chỉ xoay quanh master data chung, mà phải kiểm soát rất chặt các vùng:

- tồn kho
- batch/lô
- hạn dùng
- QC
- hàng hoàn
- bàn giao ĐVVC
- đơn gia công ngoài
- chuyển NVL/bao bì cho nhà máy
- claim nhà máy
- phiếu điều chỉnh
- dữ liệu giá/công nợ/thanh toán

---

## 3. Nguyên tắc quản trị dữ liệu

### 3.1. Một nguồn sự thật duy nhất

Mỗi loại dữ liệu quan trọng phải có **một nơi sở hữu chính** trong ERP.

Ví dụ:

| Dữ liệu | Nguồn sự thật |
|---|---|
| Mã SKU | Product/Master Data |
| Batch/lô | Inventory/QC |
| Tồn kho | Stock Ledger + Stock Balance |
| Giá bán | Pricing/Sales |
| Giá vốn | Inventory/Finance |
| Nhà cung cấp | Supplier Master |
| Khách hàng | Customer Master |
| ĐVVC | Shipping Master |
| Công thức/BOM | Product/R&D hoặc Production Master |
| Đơn gia công | Subcontract Manufacturing |
| Hàng hoàn | Returns |
| Quyền người dùng | IAM/RBAC |

Không được để Excel, Zalo, giấy note, hoặc file cá nhân trở thành nguồn sự thật sau go-live.

---

### 3.2. Không sửa trực tiếp dữ liệu giao dịch trọng yếu

Các dữ liệu sau **không được sửa trực tiếp** sau khi đã phát sinh nghiệp vụ:

- stock ledger
- phiếu nhập đã post
- phiếu xuất đã post
- QC release/hold/fail đã duyệt
- shipment handover đã xác nhận
- return inspection đã chốt
- payment đã đối soát
- batch đã phát sinh giao dịch
- công nợ đã ghi nhận
- đơn gia công đã nghiệm thu

Nếu sai phải dùng:

- phiếu điều chỉnh
- reverse transaction
- cancellation có lý do
- correction request
- approval workflow
- audit log

---

### 3.3. Batch và hạn dùng là dữ liệu cấp rủi ro cao

Trong ngành mỹ phẩm, batch/lô và hạn dùng là dữ liệu bảo vệ thương hiệu.

Bất kỳ thay đổi nào liên quan tới:

- batch number
- manufacturing date
- expiry date
- QC status
- quarantine status
- batch release
- batch recall/hold

đều phải có:

- lý do
- người đề xuất
- người duyệt
- timestamp
- before/after value
- bằng chứng đính kèm nếu có

---

### 3.4. Dữ liệu kho phải có dấu vết vận hành

Vì kho có quy trình cuối ngày gồm kiểm kê, đối soát số liệu và báo cáo quản lý, ERP phải lưu được:

- ai nhập/xuất
- thời điểm thao tác
- chứng từ gốc
- số lượng hệ thống
- số lượng thực tế
- chênh lệch
- lý do chênh lệch
- kết quả xử lý
- người duyệt điều chỉnh

Không cho điều chỉnh tồn “cho khớp” mà không có phiếu và log.

---

### 3.5. Thay đổi nhỏ cũng phải có owner

Mọi thay đổi đều phải có người chịu trách nhiệm.

Không chấp nhận các câu kiểu:

- “bên kho bảo sửa”
- “kế toán cần gấp”
- “sale nhờ đổi giá”
- “dev tự thêm cho tiện”
- “sếp nói miệng”

Tất cả phải đi qua change request nếu ảnh hưởng tới dữ liệu, quy trình, quyền, báo cáo hoặc hệ thống production.

---

## 4. Vai trò quản trị dữ liệu

### 4.1. Data Owner

Data Owner là người chịu trách nhiệm cuối cùng về một domain dữ liệu.

| Domain dữ liệu | Data Owner đề xuất |
|---|---|
| Sản phẩm/SKU/BOM | Head R&D / Product Owner |
| Nguyên vật liệu/bao bì | Purchasing Manager + R&D |
| Batch/QC | QA/QC Manager |
| Kho/tồn/stock movement | Warehouse Manager |
| Đơn hàng/khách hàng | Sales Manager / CRM Owner |
| Giao hàng/ĐVVC | Logistics/Warehouse Manager |
| Hàng hoàn | CSKH + Warehouse Manager |
| Gia công ngoài | Production/Purchasing Manager |
| Giá/chiết khấu | Sales Manager + Finance |
| Công nợ/thanh toán | Finance Manager |
| Nhân sự/quyền | HR + ERP Admin |
| Hệ thống/API/DB | Tech Lead |

---

### 4.2. Data Steward

Data Steward là người vận hành dữ liệu hằng ngày.

Ví dụ:

- master data admin tạo SKU
- kho cập nhật vị trí kho
- QA cập nhật QC status
- purchasing cập nhật NCC
- sales admin cập nhật bảng giá theo phê duyệt
- finance đối soát COD/công nợ

Data Steward được thao tác nhưng không mặc định có quyền duyệt dữ liệu nhạy cảm.

---

### 4.3. ERP Admin

ERP Admin quản lý cấu hình hệ thống, role, permission, workflow, notification, không được tùy tiện sửa dữ liệu nghiệp vụ.

ERP Admin có thể:

- tạo user
- gán role theo phê duyệt
- khóa/mở user
- cấu hình menu
- cấu hình workflow đã được duyệt
- hỗ trợ audit log
- hỗ trợ export theo quyền

ERP Admin không được:

- sửa tồn kho trực tiếp
- sửa giá vốn
- sửa batch/hạn dùng
- sửa QC status
- sửa công nợ
- xóa chứng từ

---

### 4.4. Tech Lead / DevOps

Tech Lead và DevOps chịu trách nhiệm kỹ thuật:

- database migration
- deploy
- rollback
- backup/restore
- monitoring
- hotfix
- security patch
- API change
- production support

Tech Lead/DevOps không được tự sửa dữ liệu production nếu không có:

- approved data correction request
- backup trước thao tác
- script review
- audit evidence
- post-action verification

---

## 5. Phân loại dữ liệu theo mức độ rủi ro

### 5.1. Level 1 - Dữ liệu thấp rủi ro

Ví dụ:

- mô tả nội bộ
- ghi chú không ảnh hưởng chứng từ
- label hiển thị
- ảnh minh họa không dùng cho pháp lý

Quyền sửa: Data Steward hoặc người phụ trách module.

---

### 5.2. Level 2 - Dữ liệu vận hành thường

Ví dụ:

- vị trí kho
- ghi chú đơn hàng
- địa chỉ giao hàng trước khi xuất
- thông tin liên hệ NCC/khách
- mô tả package

Quyền sửa: người phụ trách module, có audit log.

---

### 5.3. Level 3 - Dữ liệu ảnh hưởng vận hành và báo cáo

Ví dụ:

- SKU active/inactive
- UOM/quy đổi đơn vị
- bảng giá
- chiết khấu
- trạng thái đơn hàng
- trạng thái shipment
- return disposition
- mapping ĐVVC
- mapping kênh bán

Quyền sửa: cần approval của Data Owner hoặc trưởng bộ phận.

---

### 5.4. Level 4 - Dữ liệu rủi ro cao

Ví dụ:

- batch number
- expiry date
- QC status
- stock quantity
- stock adjustment
- stock ledger
- cost/gross margin
- công nợ
- payment reconciliation
- payout KOL
- shipment handover đã ký
- nghiệm thu gia công

Quyền sửa: chỉ qua correction workflow, có approval nhiều tầng.

---

### 5.5. Level 5 - Dữ liệu không được sửa trực tiếp

Ví dụ:

- stock ledger đã post
- audit log
- deleted transaction record
- historical approval log
- immutable document number
- payment đã đối soát
- batch đã release và đã bán

Không sửa trực tiếp. Chỉ được reverse/correction bằng chứng từ mới.

---

## 6. Data Domain Governance

## 6.1. Product/SKU Master

### Chủ sở hữu

- Product/R&D Owner
- Sales Owner
- Finance Owner cho trường liên quan giá vốn/nhóm báo cáo

### Dữ liệu chính

- SKU code
- product name
- product type
- brand
- category
- unit of measure
- barcode
- active/inactive
- shelf life
- pack size
- sales channel eligibility
- QC requirement
- BOM reference nếu có

### Quy tắc

- SKU code không được đổi sau khi đã phát sinh giao dịch.
- Nếu cần đổi mã, tạo SKU mới và mapping SKU cũ sang SKU mới.
- SKU chưa đủ thông tin bắt buộc không được active để bán.
- Mỹ phẩm phải có shelf life/hạn dùng logic.
- SKU bán được phải có quy định QC/batch tracking.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo SKU mới | Product/R&D + Sales + Finance nếu có giá |
| Active SKU | Product Owner + QA nếu cần |
| Inactive SKU | Sales + Warehouse + Finance |
| Sửa tên hiển thị | Product Owner |
| Sửa UOM/quy đổi | Product Owner + Finance + Warehouse |
| Sửa shelf life | QA/R&D |

---

## 6.2. Material/Packaging Master

### Chủ sở hữu

- Purchasing
- R&D/Product
- Warehouse
- QA/QC

### Dữ liệu chính

- material code
- material name
- material type
- UOM
- supplier eligibility
- QC requirement
- storage condition
- shelf life nếu có
- packaging spec
- approved vendor list

### Quy tắc

- Nguyên liệu/bao bì dùng cho sản xuất/gia công phải nằm trong master data.
- Item chưa được approve không được xuất cho nhà máy gia công.
- Bao bì/nguyên liệu gửi nhà máy phải có phiếu chuyển và biên bản bàn giao.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo material mới | R&D/QA + Purchasing |
| Thêm NCC approved | Purchasing + QA |
| Sửa UOM | Purchasing + Warehouse + Finance |
| Sửa storage condition | QA |
| Inactive material | Purchasing + R&D + Warehouse |

---

## 6.3. Supplier/Factory Master

### Chủ sở hữu

- Purchasing
- Production Owner
- Finance

### Dữ liệu chính

- supplier/factory code
- company name
- contact
- address
- tax info
- payment terms
- lead time
- MOQ
- type: supplier/factory/carrier/service provider
- approved status
- documents: contract, COA, MSDS, certificates nếu có

### Quy tắc

- Nhà máy gia công phải được phân loại riêng với nhà cung cấp nguyên liệu.
- NCC/factory chưa approved không được tạo PO/subcontract order chính thức.
- Payment terms thay đổi phải có Finance duyệt.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo NCC/factory | Purchasing + Finance |
| Approved vendor/factory | Purchasing + QA/Production |
| Sửa payment terms | Finance |
| Block NCC/factory | Purchasing Manager + Finance/QA nếu liên quan lỗi |

---

## 6.4. Warehouse/Location Master

### Chủ sở hữu

- Warehouse Manager
- ERP Admin cho cấu hình kỹ thuật

### Dữ liệu chính

- warehouse code
- warehouse type
- bin/location
- zone
- return area
- quarantine area
- packing area
- carrier handover area
- damaged/lab area

### Quy tắc

- Khu vực hàng hoàn phải tách riêng.
- Khu vực QC hold/quarantine phải tách riêng.
- Khu vực bàn giao ĐVVC phải có zone để đối chiếu/scan.
- Location/bin không được xóa nếu đã có transaction.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo kho mới | COO/Warehouse Manager + Finance |
| Tạo zone/bin mới | Warehouse Manager |
| Deactivate location | Warehouse Manager + ERP Admin |
| Đổi type location | Warehouse Manager + QA nếu liên quan quarantine |

---

## 6.5. Batch/Lot Master

### Chủ sở hữu

- QA/QC Manager
- Warehouse Manager
- Production/Purchasing tùy nguồn batch

### Dữ liệu chính

- batch_no
- SKU/material
- source: purchase/production/subcontract/return
- manufacturing date
- expiry date
- QC status
- quantity received
- quantity available
- quantity hold
- supplier/factory reference
- COA/file evidence

### Quy tắc

- Batch number phải unique theo item.
- Batch không có expiry date không được bán nếu sản phẩm bắt buộc hạn dùng.
- QC status mặc định khi nhập về là `HOLD` nếu cần kiểm.
- Chỉ QA được release từ `HOLD` sang `PASS`.
- Batch `FAIL` hoặc `HOLD` không được xuất bán.
- Batch đã phát sinh bán hàng không được sửa trực tiếp expiry date.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo batch | Warehouse receiving + QA rule |
| QC PASS/FAIL | QA/QC |
| Sửa expiry date | QA Manager + Warehouse Manager + Finance nếu ảnh hưởng tồn |
| Batch hold/recall | QA Manager + COO |
| Batch dispose | QA + Finance + Warehouse |

---

## 6.6. Stock Ledger / Inventory Balance

### Chủ sở hữu

- Warehouse Manager
- Finance Owner cho valuation
- Tech Lead cho technical integrity

### Quy tắc vàng

- Không sửa stock balance trực tiếp.
- Mọi thay đổi tồn phải tạo stock movement.
- Stock ledger là immutable.
- Adjustment phải có lý do, chứng từ và approval.
- Cycle count cuối ngày phải ghi nhận chênh lệch nếu có.

### Movement types chuẩn

| Movement type | Ý nghĩa |
|---|---|
| INBOUND_RECEIPT | Nhận hàng nhập kho |
| QC_RELEASE | Chuyển hàng QC pass vào khả dụng |
| PURCHASE_REJECT | Trả NCC do không đạt |
| SALES_RESERVE | Giữ hàng cho đơn |
| SALES_PICK | Pick hàng |
| SALES_ISSUE | Xuất bán |
| CARRIER_HANDOVER | Bàn giao ĐVVC |
| RETURN_RECEIPT | Nhận hàng hoàn |
| RETURN_TO_AVAILABLE | Hàng hoàn còn dùng được |
| RETURN_TO_DAMAGED | Hàng hoàn không dùng được |
| PRODUCTION_ISSUE | Xuất NVL/bao bì cho sản xuất/gia công |
| SUBCONTRACT_TRANSFER | Chuyển NVL/bao bì cho nhà máy |
| SUBCONTRACT_RECEIPT | Nhận hàng gia công về |
| STOCK_ADJUSTMENT_POSITIVE | Điều chỉnh tăng |
| STOCK_ADJUSTMENT_NEGATIVE | Điều chỉnh giảm |
| STOCK_COUNT_LOCK | Khóa kỳ kiểm kê |
| STOCK_COUNT_VARIANCE | Ghi nhận chênh lệch kiểm kê |

---

## 6.7. Sales Order / Pricing / Discount

### Chủ sở hữu

- Sales Manager
- Finance Manager
- CEO/COO cho chính sách lớn

### Quy tắc

- Đơn hàng phải dùng price list hoặc discount rule đã active.
- Discount vượt ngưỡng phải duyệt.
- Không được sửa giá sau khi đơn đã xuất hàng, trừ khi có credit/debit note hoặc adjustment được duyệt.
- Đơn có hàng batch hold/fail không được confirm xuất.
- Đơn tạo từ kênh ngoài phải có external reference để chống trùng.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo bảng giá | Sales + Finance |
| Active bảng giá | Sales Manager + Finance |
| Discount vượt ngưỡng | Sales Manager/CEO theo mức |
| Sửa giá đơn đã confirm | Finance + Sales Manager |
| Hủy đơn đã reserve | Sales Manager + Warehouse nếu ảnh hưởng pick |

---

## 6.8. Shipping / Carrier / Manifest

### Chủ sở hữu

- Warehouse/Logistics Manager
- CSKH cho thông tin khách
- Finance cho COD/fee

### Quy tắc

- Đơn chỉ được đưa vào manifest nếu đã packed.
- Bàn giao ĐVVC phải scan/verify hoặc xác nhận số lượng theo manifest.
- Nếu scan chưa đủ, không được đóng manifest ở trạng thái hoàn tất.
- Manifest đã handover không được sửa trực tiếp danh sách đơn.
- Sửa phải có correction/exception record.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo carrier | Logistics + Finance |
| Mapping service | Logistics |
| Hủy handover | Warehouse Manager |
| Điều chỉnh manifest đã handover | Warehouse Manager + Finance nếu COD/fee ảnh hưởng |

---

## 6.9. Returns / Hàng hoàn

### Chủ sở hữu

- CSKH
- Warehouse Manager
- QA nếu liên quan chất lượng
- Finance nếu hoàn tiền/công nợ

### Quy tắc

- Hàng hoàn phải đi vào khu vực hàng hoàn trước.
- Phải quét/ghi nhận return receipt.
- Phải kiểm tình trạng trước khi nhập khả dụng.
- Hàng còn dùng được mới chuyển về kho khả dụng.
- Hàng không dùng được phải chuyển lab/damaged/quarantine.
- Complaint chất lượng phải trace batch.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Return disposition còn dùng | Warehouse + QA rule nếu cần |
| Return disposition không dùng | Warehouse + QA |
| Hoàn tiền | CSKH + Finance |
| Sửa trạng thái return đã chốt | Warehouse Manager + CSKH + Finance nếu có tiền |

---

## 6.10. Subcontract Manufacturing / Gia công ngoài

### Chủ sở hữu

- Production/Purchasing Manager
- Warehouse Manager
- QA/QC
- Finance

### Dữ liệu chính

- subcontract order
- factory
- SKU/product
- quantity
- specification
- sample approval status
- deposit payment
- material/packaging transfer
- handover documents
- production status
- receipt status
- QC result
- defect claim
- final payment

### Quy tắc

- Đơn gia công phải có xác nhận số lượng/quy cách/mẫu mã.
- Chuyển NVL/bao bì cho nhà máy phải có phiếu chuyển và biên bản bàn giao.
- Mẫu phải được chốt trước sản xuất hàng loạt nếu quy trình yêu cầu.
- Hàng nhận về phải QC/kiểm số lượng trước khi nhập khả dụng.
- Hàng lỗi phải tạo claim gửi nhà máy trong SLA 3-7 ngày nếu quy trình yêu cầu.
- Thanh toán lần cuối chỉ khi nghiệm thu đạt hoặc đã xử lý exception.

### Change approval

| Thay đổi | Người duyệt |
|---|---|
| Tạo đơn gia công | Production/Purchasing + Finance nếu có cọc |
| Chuyển NVL/bao bì | Warehouse + Production |
| Duyệt mẫu | Product/R&D/QA |
| Nghiệm thu hàng | QA + Warehouse + Production |
| Thanh toán cuối | Finance + Production/Purchasing |
| Claim nhà máy | QA/Production + Purchasing |

---

## 7. Quy trình tạo/sửa Master Data

## 7.1. Master Data Request Flow

```text
Người đề xuất
→ tạo Master Data Request
→ nhập đầy đủ thông tin bắt buộc
→ đính kèm file/bằng chứng nếu có
→ Data Steward kiểm tra trùng/lỗi
→ Data Owner duyệt
→ ERP Admin/Data Steward tạo hoặc cập nhật
→ hệ thống ghi audit log
→ thông báo cho bộ phận liên quan
```

---

## 7.2. Điều kiện bắt buộc trước khi tạo mới

### SKU mới

- tên sản phẩm
- brand/category
- UOM
- barcode nếu có
- shelf life
- batch tracking rule
- QC requirement
- kênh bán dự kiến
- status mặc định: draft/inactive

### Material/Packaging mới

- tên item
- UOM
- loại nguyên liệu/bao bì
- supplier/factory liên quan
- QC requirement
- storage condition
- spec/file nếu có

### Supplier/Factory mới

- thông tin pháp lý cơ bản
- contact
- payment terms
- loại đối tác
- approved status
- chứng từ cần thiết nếu có

### Warehouse location mới

- warehouse
- zone
- bin/location
- type
- có cho pick không
- có cho return/quarantine không

---

## 7.3. Duplicate Check

Trước khi tạo master data mới, bắt buộc kiểm tra trùng theo:

| Domain | Kiểm tra trùng |
|---|---|
| SKU | tên gần giống, barcode, pack size |
| Material | tên nguyên liệu, supplier code, spec |
| Packaging | kích thước, quy cách, supplier code |
| Supplier | tax code, tên công ty, số điện thoại |
| Customer | số điện thoại, email, mã kênh |
| KOL | số điện thoại, email, social handle |
| Warehouse location | warehouse + zone + bin |

Nếu trùng nghi ngờ, không tạo mới cho tới khi Data Owner quyết định merge/tạo mới.

---

## 8. Quy trình sửa dữ liệu nhạy cảm

## 8.1. Sensitive Data Correction Flow

```text
Người phát hiện lỗi
→ tạo Data Correction Request
→ chọn domain dữ liệu
→ mô tả lỗi
→ nhập giá trị hiện tại và giá trị đề xuất
→ đính kèm bằng chứng
→ Data Owner review
→ Finance/QA/Warehouse review nếu liên quan
→ Tech Lead review nếu cần script
→ phê duyệt
→ thực hiện correction
→ verify sau correction
→ đóng request
→ lưu audit evidence
```

---

## 8.2. Những thay đổi bắt buộc dùng Correction Request

- sửa batch number
- sửa expiry date
- sửa QC status đã duyệt
- sửa stock quantity
- sửa manifest đã handover
- sửa return disposition đã chốt
- sửa payment/COD đã đối soát
- sửa price/cost sau khi đơn đã phát sinh
- sửa đơn gia công đã nghiệm thu
- sửa quyền admin đặc biệt
- sửa dữ liệu bằng DB script

---

## 8.3. Correction không được phép

Không cho phép:

- xóa audit log
- xóa stock ledger
- update stock balance trực tiếp không có movement
- xóa chứng từ đã post
- sửa document number
- bypass workflow approval
- xóa dữ liệu để che lỗi vận hành

---

## 9. Change Control cho hệ thống

## 9.1. Loại Change Request

| Loại change | Ví dụ |
|---|---|
| Business Process Change | đổi luồng nhập kho, thêm bước duyệt QC |
| Feature Change | thêm màn hình carrier manifest |
| Data Model Change | thêm field `factory_claim_deadline` |
| API Change | thêm endpoint scan handover |
| UI/UX Change | đổi layout packing task |
| Permission Change | thêm quyền sửa return disposition |
| Report Change | thêm báo cáo tồn cận date theo batch |
| Integration Change | tích hợp ĐVVC mới |
| Security Change | bật MFA, đổi session timeout |
| Emergency Change | hotfix lỗi không xuất kho được |

---

## 9.2. Change Request Lifecycle

```text
Draft
→ Submitted
→ Impact Assessment
→ Approved / Rejected / Need More Info
→ Scheduled
→ Development / Configuration
→ QA Testing
→ UAT nếu cần
→ Release Approval
→ Deployed
→ Post-Release Verification
→ Closed
```

---

## 9.3. Change Priority

| Priority | Ý nghĩa | SLA review đề xuất |
|---|---|---|
| P0 Emergency | hệ thống dừng, sai tồn/tiền nghiêm trọng | ngay lập tức |
| P1 Critical | ảnh hưởng vận hành lớn, workaround yếu | 1 ngày làm việc |
| P2 Important | cần cho vận hành, có workaround | 3 ngày làm việc |
| P3 Normal | cải tiến thường | 5-10 ngày làm việc |
| P4 Backlog | ý tưởng/phase sau | theo planning |

---

## 9.4. Impact Assessment Checklist

Mỗi CR phải trả lời:

- ảnh hưởng module nào?
- ảnh hưởng dữ liệu nào?
- ảnh hưởng quyền nào?
- ảnh hưởng báo cáo nào?
- ảnh hưởng API nào?
- ảnh hưởng database migration không?
- ảnh hưởng workflow phê duyệt không?
- ảnh hưởng UAT test case nào?
- có cần training lại user không?
- có cần rollback plan không?
- có nguy cơ sai tồn/tiền/batch không?

---

## 9.5. Quy tắc approve change

| Loại change | Approver bắt buộc |
|---|---|
| Business process | Product Owner + Business Owner |
| Stock/inventory | Warehouse Manager + Finance/QA nếu liên quan |
| QC/batch | QA Manager |
| Finance/payment | Finance Manager |
| Permission/security | ERP Admin + Security/Tech Lead + Business Owner |
| API/database | Tech Lead |
| Integration | Tech Lead + Business Owner |
| Report KPI | Data Owner + CEO/COO nếu dashboard quản trị |
| Emergency hotfix | Tech Lead + Product Owner + post-review sau deploy |

---

## 10. Emergency Change Control

Emergency change chỉ dùng khi:

- hệ thống production không chạy
- không tạo/xuất đơn được
- không nhập kho/đóng hàng/bàn giao được
- sai tồn nghiêm trọng
- sai công nợ/thanh toán nghiêm trọng
- lỗ hổng bảo mật nghiêm trọng
- mất dữ liệu/nguy cơ mất dữ liệu

Quy trình nhanh:

```text
Phát hiện sự cố
→ tạo Emergency CR
→ Tech Lead + Product Owner approve nhanh
→ backup/snapshot nếu cần
→ deploy hotfix
→ smoke test
→ theo dõi
→ post-incident review trong 24-48h
→ cập nhật tài liệu/test case nếu cần
```

Không được dùng emergency change để né quy trình cho các yêu cầu bình thường.

---

## 11. Release Governance

### 11.1. Release types

| Loại release | Nội dung |
|---|---|
| Major release | thay đổi module/flow lớn |
| Minor release | thêm tính năng nhỏ, chỉnh UI/API |
| Patch release | sửa lỗi ít rủi ro |
| Hotfix | sửa lỗi khẩn cấp production |
| Config release | thay cấu hình/workflow/permission |

---

### 11.2. Release checklist

Trước khi release production:

- CR đã approved
- code reviewed
- test passed
- migration reviewed
- rollback plan có sẵn
- release note có sẵn
- impacted users được thông báo
- training/update SOP nếu cần
- smoke test script sẵn sàng
- backup/snapshot nếu có migration dữ liệu

---

### 11.3. Post-release verification

Sau release phải kiểm tra tối thiểu:

- login/RBAC
- tạo chứng từ chính
- stock ledger không lỗi
- sales order flow
- pick/pack
- carrier handover
- returns
- QC status
- dashboard/report trọng yếu
- audit log ghi đúng

---

## 12. Data Quality Management

## 12.1. Chỉ số chất lượng dữ liệu

| KPI | Ý nghĩa | Target đề xuất |
|---|---|---|
| Duplicate SKU rate | tỷ lệ SKU trùng/nghi trùng | < 1% |
| Missing expiry rate | batch thiếu hạn dùng | 0% với hàng bắt buộc |
| QC status missing | batch thiếu QC status | 0% |
| Stock variance rate | chênh lệch kiểm kê | giảm theo tháng |
| Unapproved master data | master data dùng khi chưa duyệt | 0 |
| Unmapped carrier order | đơn thiếu mapping ĐVVC | < 1% |
| Return disposition pending | hàng hoàn chưa phân loại quá SLA | theo SLA |
| Data correction count | số correction mỗi tháng | theo dõi trend |
| Manual DB correction | correction bằng script DB | càng thấp càng tốt |

---

## 12.2. Data Quality Review

Tần suất đề xuất:

| Loại review | Tần suất | Owner |
|---|---|---|
| Stock variance review | hằng ngày/cuối ca | Warehouse Manager |
| Return pending review | hằng ngày | CSKH + Warehouse |
| Batch/QC exception review | hằng ngày | QA/QC |
| Master data duplicate review | hằng tuần | Data Steward |
| Price/discount exception review | hằng tuần | Sales + Finance |
| Data correction review | hằng tháng | Product Owner + Data Owners |
| Access review | hằng tháng/quý | ERP Admin + HR |

---

## 13. Access Governance

## 13.1. User lifecycle

```text
HR tạo yêu cầu user mới
→ trưởng bộ phận xác nhận role
→ ERP Admin tạo user
→ user đổi mật khẩu lần đầu
→ cấp quyền theo role
→ log active
```

Khi nhân sự nghỉ/chuyển bộ phận:

```text
HR tạo offboarding/access change
→ ERP Admin khóa/sửa quyền
→ thu hồi quyền nhạy cảm ngay
→ lưu log
```

---

## 13.2. Periodic Access Review

Tối thiểu mỗi tháng/quý review:

- user còn làm việc không
- role có còn đúng không
- ai có quyền admin
- ai có quyền sửa giá
- ai có quyền duyệt QC
- ai có quyền adjustment stock
- ai có quyền export dữ liệu nhạy cảm
- ai có quyền xem cost/margin/payroll

---

## 13.3. Break-glass account

Break-glass account là tài khoản khẩn cấp.

Quy tắc:

- mặc định disable hoặc khóa kiểm soát chặt
- chỉ mở khi sự cố nghiêm trọng
- phải có approval
- session được log toàn bộ
- sau khi dùng phải disable lại
- phải có post-review

---

## 14. Data Retention & Archiving

### 14.1. Dữ liệu không được xóa cứng

- chứng từ kho
- stock movement
- audit log
- QC record
- batch record
- sales order
- return record
- shipment manifest
- payment record
- subcontract order
- approval log

Dùng soft delete/cancel/inactive nếu cần.

---

### 14.2. Archive

Dữ liệu cũ có thể archive để tối ưu hiệu năng nhưng vẫn phải truy xuất được khi cần:

- đơn hàng cũ
- report cũ
- shipment cũ
- log cũ

Archive không được làm mất trace batch hoặc audit trail.

---

## 15. Report & KPI Governance

### 15.1. Report phải có định nghĩa chuẩn

Mỗi báo cáo phải ghi rõ:

- tên báo cáo
- owner
- mục tiêu
- nguồn dữ liệu
- công thức tính
- filter mặc định
- refresh frequency
- ai được xem
- export permission

---

### 15.2. Không tạo report mới nếu chưa chốt công thức

Ví dụ:

- “doanh thu” là gross hay net?
- “tồn kho” là physical hay available?
- “hàng hoàn” tính lúc nhận hay lúc phân loại?
- “đơn giao thành công” theo carrier status hay ERP status?
- “lãi gộp” có trừ phí ship/discount/KOL không?

Nếu chưa chốt, không build report chính thức.

---

## 16. Governance cho workflow thực tế của kho

## 16.1. Daily Warehouse Board

Vì kho có quy trình hằng ngày, các dữ liệu trong Daily Warehouse Board phải có owner và rule:

| Dữ liệu | Owner | Rule |
|---|---|---|
| Đơn trong ngày | Sales/OMS | Đơn phải có trạng thái rõ |
| Picking task | Warehouse | Không pick hàng HOLD/FAIL |
| Packing task | Warehouse | Phải kiểm SKU/số lượng |
| Carrier handover | Warehouse/Logistics | Phải scan/đối chiếu manifest |
| Stock count | Warehouse | Chênh lệch phải có variance record |
| End-of-day report | Warehouse Manager | Không close shift nếu còn exception P0/P1 |

---

## 16.2. Shift Closing Governance

Không cho close shift nếu:

- có manifest chưa bàn giao xong
- có scan thiếu đơn chưa xử lý
- có hàng hoàn chưa đưa vào khu vực hàng hoàn
- có stock count variance chưa giải thích
- có phiếu nhập/xuất draft bị treo bất thường
- có batch/QC exception chưa gắn owner

Close shift phải lưu:

- người close
- thời gian
- số đơn xử lý
- số manifest
- số hàng hoàn
- stock variance
- issue còn mở
- ghi chú quản lý

---

## 17. Governance cho bàn giao ĐVVC

Quy tắc:

- ĐVVC phải có master data.
- Carrier service phải được cấu hình.
- Manifest phải có danh sách đơn rõ.
- Đơn phải packed trước khi đưa vào manifest.
- Scan thiếu phải tạo exception.
- Scan đủ mới được ký/xác nhận bàn giao.
- Manifest đã ký không sửa trực tiếp.
- Nếu sai, tạo handover correction/incident.

Data cần audit:

- người tạo manifest
- người scan
- số đơn expected
- số đơn scanned
- missing list
- người xác nhận handover
- thời gian handover
- ĐVVC nhận

---

## 18. Governance cho hàng hoàn

Quy tắc:

- Hàng hoàn phải đi qua return receiving.
- Không cho nhập thẳng về available stock.
- Phải có kiểm tình trạng.
- Nếu còn dùng: chuyển về available theo rule.
- Nếu không dùng: chuyển damaged/lab/quarantine.
- Nếu nghi lỗi chất lượng: link complaint và batch.

Không được:

- bán lại hàng hoàn chưa inspection
- sửa disposition đã chốt không có approval
- nhập hàng hoàn mà thiếu order/tracking reference trừ exception được duyệt

---

## 19. Governance cho gia công ngoài

Quy tắc:

- Factory phải approved.
- Subcontract order phải có số lượng/quy cách/mẫu mã.
- Deposit/final payment phải gắn với finance approval.
- Chuyển NVL/bao bì phải có stock movement và handover document.
- Mẫu phải approved trước mass production nếu required.
- Hàng về phải receiving + QC.
- Hàng lỗi phải tạo claim trong SLA.
- Final payment chỉ sau nghiệm thu hoặc exception approved.

Data cần audit:

- ai tạo đơn gia công
- ai duyệt cọc
- ai xuất NVL/bao bì
- tài liệu bàn giao
- mẫu approved bởi ai
- kết quả QC hàng về
- claim nhà máy nếu có
- thanh toán cuối

---

## 20. Governance cho cấu hình hệ thống

Các cấu hình sau phải đi qua CR:

- workflow approval
- permission/role
- document numbering
- stock movement type
- QC status/rules
- price/discount rule
- return disposition rule
- carrier integration config
- API integration credential
- notification rule
- report formula

Không được sửa config production không có log.

---

## 21. Versioning tài liệu và sự đồng bộ

Khi một change ảnh hưởng tài liệu, phải cập nhật tài liệu liên quan.

Ví dụ thêm flow scan bàn giao ĐVVC thì phải update:

- PRD/SRS
- Process Flow
- Screen List
- API Contract
- Database Schema
- UAT
- SOP
- Risk Playbook

Quy tắc:

- mỗi tài liệu có version
- mỗi version có change log
- không dùng file cũ không rõ version
- tài liệu mới nhất phải có trong knowledge base nội bộ

---

## 22. RACI tổng quan

| Hoạt động | CEO/COO | Product Owner | Data Owner | ERP Admin | Tech Lead | Finance | QA/QC | Warehouse |
|---|---|---|---|---|---|---|---|---|
| Tạo master data | I | C | A | C | I | C | C | C |
| Sửa dữ liệu nhạy cảm | I/A nếu rủi ro cao | A | A | C | C | C | C | C |
| Stock adjustment | I | C | C | I | I | A/C | C | A/R |
| Batch/QC change | I | C | A | I | I | C | A/R | C |
| Permission change | I | C | C | R | C | C | C | C |
| System change | I | A | C | C | R/A kỹ thuật | C | C | C |
| Report formula change | C | A | A | I | C | C | C | C |
| Emergency hotfix | I | A | C | C | R/A | C | C | C |
| Go-live release | A | A | C | C | R | C | C | C |

Ký hiệu:

- R = Responsible
- A = Accountable
- C = Consulted
- I = Informed

---

## 23. Biểu mẫu đề xuất

## 23.1. Master Data Request Template

```text
Request ID:
Request type: Create / Update / Deactivate
Domain: SKU / Material / Supplier / Customer / Warehouse / Carrier / Other
Requested by:
Department:
Business reason:
Required date:
Data fields:
Attachments:
Duplicate check result:
Data Steward review:
Data Owner approval:
Created/Updated by:
Completion date:
```

---

## 23.2. Data Correction Request Template

```text
Correction ID:
Domain:
Record ID / Document No:
Current value:
Proposed value:
Reason:
Impact:
Evidence attached:
Requested by:
Reviewed by Data Owner:
Reviewed by Finance/QA/Warehouse if needed:
Tech review if script needed:
Approval status:
Correction method:
Verification result:
Closed by:
```

---

## 23.3. Change Request Template

```text
CR ID:
Title:
Type:
Priority:
Requested by:
Business reason:
Current behavior:
Expected behavior:
Affected modules:
Affected data:
Affected reports:
Affected permissions:
Affected APIs/database:
Risk assessment:
Rollback plan:
Approvers:
Target release:
Testing required:
Training/SOP update required:
Post-release verification:
```

---

## 24. Governance Meeting Cadence

| Cuộc họp | Tần suất | Thành phần | Mục tiêu |
|---|---|---|---|
| Daily Ops Exception Review | hằng ngày | Kho, CSKH, QA, Production | xử lý tồn, đơn, hàng hoàn, batch exception |
| Weekly Change Review | hằng tuần | PO, Tech Lead, Business Owners | duyệt/chốt CR |
| Weekly Data Quality Review | hằng tuần | Data Steward, Data Owners | duplicate, missing, correction |
| Monthly Access Review | hằng tháng | ERP Admin, HR, Managers | rà quyền |
| Monthly Governance Review | hằng tháng | CEO/COO, PO, Tech Lead | rủi ro, KPI, thay đổi lớn |

---

## 25. KPI quản trị thay đổi

| KPI | Ý nghĩa |
|---|---|
| CR approval lead time | thời gian duyệt change |
| CR delivery lead time | thời gian từ approved đến deployed |
| Emergency change count | số hotfix khẩn cấp |
| Failed release count | số release lỗi/rollback |
| Post-release defect count | lỗi phát sinh sau release |
| Data correction count | số correction dữ liệu |
| Unauthorized change count | số thay đổi không qua quy trình |
| Access violation count | số lỗi quyền |
| Stock adjustment count | số điều chỉnh tồn |
| Repeat incident count | sự cố lặp lại |

---

## 26. Anti-pattern cần cấm

Các hành vi sau phải cấm hoặc kiểm soát nghiêm:

- tạo mã hàng mới bằng cách nhắn dev/admin
- sửa tồn kho bằng DB
- đổi QC status không có QA duyệt
- nhập hàng hoàn thẳng vào hàng bán được
- sửa manifest đã bàn giao cho khớp số
- xóa chứng từ sai thay vì hủy/correction
- deploy production không release note
- sửa API không cập nhật OpenAPI
- sửa field database không migration
- đổi quyền user không có request
- copy dữ liệu nhạy cảm ra Excel không kiểm soát
- report tự chế không chốt công thức

---

## 27. Definition of Done cho Governance

Một domain dữ liệu được coi là governed khi:

- có Data Owner
- có Data Steward
- có field bắt buộc
- có rule tạo/sửa/khóa
- có approval nếu dữ liệu nhạy cảm
- có audit log
- có duplicate check
- có report data quality
- có correction process
- có SOP nếu user thao tác thường xuyên

Một change được coi là hoàn tất khi:

- CR approved
- impact assessment xong
- development/configuration xong
- test passed
- release note có
- tài liệu liên quan cập nhật
- user được thông báo/training nếu cần
- production verification passed
- CR closed

---

## 28. Roadmap triển khai Governance

### Giai đoạn 1 - Trước UAT

- chốt Data Owner/Data Steward
- chốt master data request flow
- chốt correction flow
- chốt CR flow
- chốt role access review
- chốt release governance

### Giai đoạn 2 - Trong UAT

- test data correction flow
- test stock adjustment approval
- test batch/QC correction
- test permission request
- test change request nhỏ
- test SOP training

### Giai đoạn 3 - Trước Go-live

- freeze master data
- freeze quyền user
- chốt open CR được phép go-live
- chốt rollback/hotfix process
- chốt support model

### Giai đoạn 4 - Sau Go-live

- daily exception review
- weekly data quality review
- weekly change review
- monthly access review
- governance KPI report

---

## 29. Kết luận

Tài liệu này là lớp kỷ luật để ERP không bị phá từ bên trong sau khi chạy thật.

Với công ty mỹ phẩm có kho, hàng hoàn, batch/hạn dùng, QC, bàn giao ĐVVC và gia công ngoài, governance phải đặc biệt chặt ở 5 vùng:

1. **Stock ledger và tồn kho**
2. **Batch/QC/hạn dùng**
3. **Hàng hoàn và phân loại tình trạng**
4. **Bàn giao ĐVVC bằng manifest/scan**
5. **Gia công ngoài: chuyển NVL/bao bì, duyệt mẫu, nghiệm thu, claim nhà máy**

Một câu chốt:

> **ERP tốt không phải là hệ thống ai cũng sửa được cho nhanh. ERP tốt là hệ thống bắt mọi thay đổi quan trọng đi qua đúng cửa, đúng người, đúng log.**

---

## 30. Tài liệu nên làm tiếp theo

Sau tài liệu này, bước tiếp theo hợp lý là:

```text
31_ERP_Phase2_Scope_CRM_HRM_KOL_Finance_MyPham_v1.md
```

Mục tiêu: chốt phạm vi Phase 2 cho CRM, HRM, KOL/Affiliate, Finance nâng cao, commission, loyalty và BI nâng cao.
