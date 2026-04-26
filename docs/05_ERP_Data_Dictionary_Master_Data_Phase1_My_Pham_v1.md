# 05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1

**Project:** ERP Web cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Document Type:** Data Dictionary + Master Data Rulebook  
**Scope:** Phase 1  
**Version:** v1.0  
**Language:** Vietnamese  

---

## 1. Mục tiêu tài liệu

Tài liệu này chuẩn hóa ngôn ngữ dữ liệu cho toàn bộ đội dự án gồm business, BA, UI/UX, dev, tester, key users, kho, sản xuất, QA/QC, sales và finance.

Mục tiêu là để mọi bên hiểu giống nhau về:
- Mỗi thực thể dữ liệu là gì.
- Mỗi trường dữ liệu có ý nghĩa gì.
- Giá trị nào được phép nhập.
- Trạng thái nào được dùng trong từng nghiệp vụ.
- Công thức tính nào là chuẩn của hệ thống.
- Quy tắc tạo mã, sửa dữ liệu, khóa dữ liệu và audit log.

Một hệ thống ERP chỉ mạnh khi **dữ liệu gốc sạch, định nghĩa rõ, trạng thái chuẩn và công thức nhất quán**.

---

## 2. Phạm vi Phase 1

Tài liệu này áp dụng cho 6 module lõi của Phase 1:
- Dữ liệu gốc (Master Data)
- Mua hàng (Purchasing)
- QA/QC
- Sản xuất (Production)
- Kho hàng (Warehouse/WMS)
- Bán hàng (Sales/OMS)

Ngoài ra tài liệu cũng đặt sẵn nền cho các phần sẽ mở rộng sau này như CRM, HRM, KOL/Affiliate, POS, Finance mở rộng, BI.

---

## 3. Nguyên tắc dữ liệu cốt lõi

### 3.1. One Source of Truth
Mỗi loại dữ liệu gốc chỉ có **một nguồn chuẩn** trong hệ thống. Không dùng Excel hoặc file rời làm nguồn chính thức sau khi go-live.

### 3.2. Không xóa giao dịch lõi
Các giao dịch như PR, PO, phiếu nhập, phiếu xuất, lệnh sản xuất, batch QC, sales order chỉ được:
- Hủy (`Cancelled`)
- Đảo/reverse theo rule
- Khóa sau khi chốt

Không được xóa cứng khỏi hệ thống khi đã phát sinh tham chiếu nghiệp vụ.

### 3.3. Bám theo batch/lô và hạn dùng
Ngành mỹ phẩm bắt buộc quản lý theo:
- `batch_no`
- `mfg_date`
- `expiry_date`
- `qc_status`
- `inventory_status`

### 3.4. Master data phải có owner
Mỗi nhóm dữ liệu gốc phải có chủ sở hữu chịu trách nhiệm chất lượng dữ liệu:
- Sản phẩm/BOM: R&D
- Nhà cung cấp: Purchasing
- Kho: Warehouse Admin/Operations
- Khách hàng và price list: Sales Admin
- QC rule: QA/QC

### 3.5. Trường quan trọng phải có audit log
Mọi thay đổi ở các trường dưới đây phải ghi log:
- Giá mua chuẩn
- BOM/version công thức
- `qc_status`
- `expiry_date`
- `standard_cost`
- Chính sách giá/chiết khấu
- `available_stock` do điều chỉnh thủ công
- Trạng thái chứng từ

---

## 4. Chuẩn quy ước đặt mã

> Lưu ý: đây là quy ước khuyến nghị. Có thể tinh chỉnh theo brand hoặc cơ cấu công ty, nhưng phải thống nhất một kiểu duy nhất trước khi build.

### 4.1. Mã nguyên vật liệu
**Field:** `material_code`  
**Format gợi ý:** `RM-{GROUP}-{NNNN}`  
**Ví dụ:** `RM-ACT-0001`, `RM-FRG-0012`, `RM-PKG-0105`

Trong đó:
- `RM` = Raw Material
- `GROUP` = nhóm nguyên liệu (`ACT`, `BASE`, `FRG`, `CLR`, `PKG`, `LBL`...)
- `NNNN` = số tăng dần

### 4.2. Mã bán thành phẩm
**Field:** `semi_fg_code`  
**Format gợi ý:** `SF-{BRAND}-{NNNN}`  
**Ví dụ:** `SF-ABC-0008`

### 4.3. Mã thành phẩm / SKU
**Field:** `sku_code`  
**Format gợi ý:** `FG-{BRAND}-{CAT}-{SIZE}-{NNNN}`  
**Ví dụ:** `FG-ABC-SER-30-0012`

### 4.4. Mã batch/lô sản xuất
**Field:** `batch_no`  
**Format gợi ý:** `{SKU_SHORT}-{YYMMDD}-{SEQ}`  
**Ví dụ:** `SER30-260423-01`

### 4.5. Mã kho
**Field:** `warehouse_code`  
**Format gợi ý:** `WH-{SITE}-{TYPE}`  
**Ví dụ:** `WH-HCM-RM`, `WH-HCM-FG`, `WH-HN-STORE`

### 4.6. Mã nhà cung cấp
**Field:** `supplier_code`  
**Format gợi ý:** `SUP-{GROUP}-{NNNN}`  
**Ví dụ:** `SUP-RM-0021`

### 4.7. Mã khách hàng / đại lý
**Field:** `customer_code`  
**Format gợi ý:** `CUS-{CHANNEL}-{NNNN}`  
**Ví dụ:** `CUS-DL-0009`, `CUS-WS-0102`, `CUS-RETAIL-0120`

### 4.8. Mã đơn mua hàng
**Field:** `po_no`  
**Format gợi ý:** `PO-{YYMM}-{SEQ}`  
**Ví dụ:** `PO-2604-0018`

### 4.9. Mã đơn bán hàng
**Field:** `so_no`  
**Format gợi ý:** `SO-{CHANNEL}-{YYMM}-{SEQ}`  
**Ví dụ:** `SO-B2B-2604-0007`, `SO-ECOM-2604-0152`

### 4.10. Mã lệnh sản xuất
**Field:** `wo_no`  
**Format gợi ý:** `WO-{YYMMDD}-{SEQ}`  
**Ví dụ:** `WO-260423-003`

---

## 5. Chuẩn master data theo domain

## 5.1. Item Master

### 5.1.1. Định nghĩa
Bảng dữ liệu chuẩn của mọi loại hàng hóa/vật tư trong hệ thống, gồm:
- Nguyên vật liệu
- Bao bì/tem nhãn/phụ liệu
- Bán thành phẩm
- Thành phẩm
- Hàng sample/tester/quà tặng nếu quản lý như item riêng

### 5.1.2. Trường dữ liệu chính

| Field | Tên hiển thị | Kiểu dữ liệu | Bắt buộc | Mô tả |
|---|---|---:|:---:|---|
| `item_id` | ID nội bộ | UUID/Bigint | Yes | ID hệ thống, không hiển thị cho người dùng thông thường |
| `item_code` | Mã hàng | String | Yes | Mã duy nhất, không trùng |
| `item_name` | Tên hàng | String | Yes | Tên chuẩn nội bộ |
| `item_type` | Loại hàng | Enum | Yes | `raw_material`, `packaging`, `semi_fg`, `finished_good`, `service`, `gift` |
| `item_group` | Nhóm hàng | Enum/String | Yes | Nhóm chi tiết theo vận hành |
| `brand_code` | Mã brand | String | No | Dùng khi doanh nghiệp đa brand |
| `uom_base` | Đơn vị cơ sở | Enum/String | Yes | Ví dụ `kg`, `g`, `ml`, `pcs`, `box` |
| `uom_purchase` | Đơn vị mua | Enum/String | No | Ví dụ `kg`, `carton` |
| `uom_issue` | Đơn vị cấp phát | Enum/String | No | Ví dụ `g`, `pcs` |
| `lot_controlled` | Quản lý theo lô | Boolean | Yes | Có/không |
| `expiry_controlled` | Quản lý hạn dùng | Boolean | Yes | Có/không |
| `shelf_life_days` | Số ngày hạn dùng chuẩn | Integer | No | Hệ thống dùng để tính gợi ý hạn dùng |
| `qc_required` | Bắt buộc QC | Boolean | Yes | Có/không |
| `status` | Trạng thái item | Enum | Yes | `draft`, `active`, `inactive`, `obsolete` |
| `standard_cost` | Giá chuẩn | Decimal | No | Chi phí chuẩn tham chiếu |
| `is_sellable` | Được phép bán | Boolean | Yes | Có/không |
| `is_purchasable` | Được phép mua | Boolean | Yes | Có/không |
| `is_producible` | Được phép sản xuất | Boolean | Yes | Có/không |
| `spec_version` | Phiên bản spec | String | No | Ràng buộc với R&D/QA |
| `created_at` | Ngày tạo | Datetime | Yes | Audit |
| `created_by` | Người tạo | User ID | Yes | Audit |
| `updated_at` | Ngày sửa | Datetime | Yes | Audit |
| `updated_by` | Người sửa | User ID | Yes | Audit |

### 5.1.3. Giá trị chuẩn cho `item_type`
- `raw_material`: nguyên vật liệu trực tiếp
- `packaging`: chai, nắp, hộp, tem, nhãn, seal, leaflet
- `semi_fg`: bán thành phẩm
- `finished_good`: thành phẩm bán ra
- `service`: dịch vụ không quản kho
- `gift`: quà tặng nếu cần quản riêng

### 5.1.4. Quy tắc dữ liệu
- `item_code` là duy nhất toàn hệ thống.
- Không cho sửa `item_type` sau khi item đã phát sinh giao dịch.
- Nếu `lot_controlled = true` thì mọi giao dịch nhập/xuất phải đi kèm `batch_no`.
- Nếu `expiry_controlled = true` thì phải có `mfg_date` và `expiry_date` ở thời điểm nhập batch.
- Nếu `qc_required = true` thì hàng nhập về có trạng thái tồn kho ban đầu là `quarantine` hoặc `hold` cho đến khi QC pass.

---

## 5.2. UOM Conversion

### 5.2.1. Định nghĩa
Bảng quy đổi đơn vị tính cho cùng một item.

### 5.2.2. Trường dữ liệu

| Field | Mô tả |
|---|---|
| `uom_from` | Đơn vị nguồn |
| `uom_to` | Đơn vị đích |
| `conversion_factor` | Hệ số quy đổi |
| `rounding_rule` | Rule làm tròn |
| `status` | Trạng thái hiệu lực |

### 5.2.3. Ví dụ
- `1 kg = 1000 g`
- `1 carton = 48 pcs`

### 5.2.4. Quy tắc
- Mỗi item phải có một `uom_base`.
- Tất cả báo cáo tồn kho chuẩn hóa về `uom_base`.

---

## 5.3. Supplier Master

### 5.3.1. Định nghĩa
Dữ liệu gốc của nhà cung cấp nguyên liệu, bao bì, dịch vụ logistics, gia công nếu có.

### 5.3.2. Trường dữ liệu chính

| Field | Tên hiển thị | Kiểu | Bắt buộc | Mô tả |
|---|---|---|:---:|---|
| `supplier_id` | ID NCC | UUID/Bigint | Yes | ID nội bộ |
| `supplier_code` | Mã NCC | String | Yes | Duy nhất |
| `supplier_name` | Tên NCC | String | Yes | Tên pháp lý hoặc tên dùng nội bộ |
| `supplier_group` | Nhóm NCC | Enum | Yes | `raw_material`, `packaging`, `service`, `logistics`, `outsource` |
| `contact_name` | Người liên hệ | String | No | |
| `phone` | Số điện thoại | String | No | |
| `email` | Email | String | No | |
| `tax_code` | Mã số thuế | String | No | |
| `payment_terms` | Điều khoản thanh toán | String | No | Ví dụ `30D`, `COD` |
| `lead_time_days` | Lead time chuẩn | Integer | No | Số ngày giao hàng chuẩn |
| `moq` | MOQ | Decimal | No | Minimum order quantity |
| `status` | Trạng thái | Enum | Yes | `draft`, `active`, `inactive`, `blacklisted` |
| `quality_score` | Điểm chất lượng | Decimal | No | Đánh giá hiệu suất NCC |
| `delivery_score` | Điểm giao hàng | Decimal | No | Đánh giá SLA |

### 5.3.3. Quy tắc
- NCC `blacklisted` không được chọn trên PR/PO mới.
- Trường `payment_terms` phải chuẩn hóa, không nhập tự do nếu hệ thống dùng bảng mã.
- Với nguyên liệu nhạy cảm, chỉ NCC trong approved vendor list mới được chọn.

---

## 5.4. Warehouse Master

### 5.4.1. Định nghĩa
Danh mục kho vật lý, kho logic và vị trí kho.

### 5.4.2. Các loại kho khuyến nghị
- Kho nguyên liệu (`RM`)
- Kho bao bì (`PKG`)
- Kho bán thành phẩm (`SF`)
- Kho thành phẩm (`FG`)
- Kho hold/quarantine (`QH`)
- Kho sample/tester (`SMP`)
- Kho hỏng/defect (`DEF`)
- Kho cửa hàng (`STORE`)

### 5.4.3. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `warehouse_code` | Mã kho |
| `warehouse_name` | Tên kho |
| `warehouse_type` | Loại kho |
| `site_code` | Chi nhánh/nhà máy/cửa hàng |
| `allow_sale_issue` | Có cho phép xuất bán không |
| `allow_prod_issue` | Có cho phép cấp cho sản xuất không |
| `allow_quarantine` | Có cho phép lưu hàng hold không |
| `status` | Trạng thái |

### 5.4.4. Quy tắc
- Chỉ kho thành phẩm khả dụng mới cho phép xuất bán.
- Kho `QH` không cho xuất bán.
- Kho sample chỉ cho xuất sample/tester/gifting theo rule riêng.

---

## 5.5. Customer Master

### 5.5.1. Định nghĩa
Dữ liệu gốc khách hàng B2B, đại lý, cửa hàng, khách lẻ có quản lý công nợ hoặc chính sách giá riêng.

### 5.5.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `customer_code` | Mã khách hàng |
| `customer_name` | Tên khách hàng |
| `customer_type` | `distributor`, `dealer`, `retail_customer`, `marketplace`, `internal_store` |
| `channel_code` | Kênh bán |
| `price_list_code` | Bảng giá áp dụng |
| `discount_group` | Nhóm chiết khấu |
| `credit_limit` | Hạn mức công nợ |
| `payment_terms` | Điều khoản thanh toán |
| `status` | Trạng thái |

### 5.5.3. Quy tắc
- Khách hàng vượt `credit_limit` có thể bị chặn xác nhận đơn tùy approval rule.
- `price_list_code` là một phần của điều kiện tính giá tại Sales Order.

---

## 5.6. BOM / Formula Master

### 5.6.1. Định nghĩa
Bảng định mức nguyên vật liệu/bao bì cho một bán thành phẩm hoặc thành phẩm. Trong mỹ phẩm cần xem đây là phần giao nhau giữa R&D, QA và Production.

### 5.6.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `bom_code` | Mã BOM |
| `parent_item_code` | Item cha (SF/FG) |
| `bom_version` | Phiên bản BOM |
| `effective_from` | Hiệu lực từ ngày |
| `effective_to` | Hiệu lực đến ngày |
| `status` | `draft`, `approved`, `inactive`, `superseded` |
| `yield_standard` | Tỷ lệ thu hồi chuẩn |
| `loss_standard_pct` | Hao hụt chuẩn % |

### 5.6.3. Trường dòng BOM

| Field | Mô tả |
|---|---|
| `component_item_code` | Mã thành phần |
| `component_qty` | Số lượng thành phần |
| `component_uom` | Đơn vị thành phần |
| `component_type` | `raw_material`, `packaging`, `semi_fg` |
| `is_active` | Có hiệu lực |

### 5.6.4. Quy tắc
- Chỉ 1 BOM version được `approved` và `effective` tại cùng một thời điểm cho cùng một `parent_item_code` theo một scope chuẩn.
- Không cho phép phát hành lệnh sản xuất nếu item không có BOM hợp lệ.
- BOM đã dùng cho lệnh sản xuất không được sửa trực tiếp; phải tạo version mới.

---

## 5.7. QC Spec Master

### 5.7.1. Định nghĩa
Bộ tiêu chí QC cho nguyên liệu, bán thành phẩm, thành phẩm.

### 5.7.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `qc_spec_code` | Mã spec QC |
| `item_code` | Mã item áp dụng |
| `spec_version` | Version spec |
| `test_parameter` | Chỉ tiêu kiểm tra |
| `test_method` | Phương pháp kiểm tra |
| `acceptance_criteria` | Tiêu chí chấp nhận |
| `status` | Trạng thái |

### 5.7.3. Quy tắc
- Spec QC phải version hóa.
- Batch phải tham chiếu spec có hiệu lực tại thời điểm nhận hàng hoặc release.

---

## 5.8. Price List Master

### 5.8.1. Định nghĩa
Bảng giá chuẩn theo kênh/nhóm khách hàng.

### 5.8.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `price_list_code` | Mã bảng giá |
| `price_list_name` | Tên bảng giá |
| `channel_code` | Kênh áp dụng |
| `customer_group` | Nhóm khách hàng |
| `effective_from` | Hiệu lực từ |
| `effective_to` | Hiệu lực đến |
| `status` | Trạng thái |

### 5.8.3. Trường dòng bảng giá

| Field | Mô tả |
|---|---|
| `sku_code` | SKU |
| `list_price` | Giá niêm yết |
| `min_price` | Giá tối thiểu được phép bán nếu có |
| `currency_code` | Loại tiền |

### 5.8.4. Quy tắc
- Mỗi khách hàng/đơn hàng tại một thời điểm chỉ có một rule giá hiệu lực sau khi engine resolve.
- `min_price` được dùng cho rule phê duyệt giảm giá.

---

## 6. Transaction Data Dictionary

## 6.1. Purchase Request (PR)

### 6.1.1. Định nghĩa
Phiếu đề nghị mua hàng từ bộ phận có nhu cầu.

### 6.1.2. Trường dữ liệu chính

| Field | Tên | Mô tả |
|---|---|---|
| `pr_no` | Số PR | Mã phiếu đề nghị mua |
| `request_dept` | Bộ phận đề nghị | Phòng ban gửi yêu cầu |
| `requester_id` | Người đề nghị | Người tạo PR |
| `need_by_date` | Cần trước ngày | Ngày cần hàng |
| `reason_code` | Lý do mua | Sản xuất, bổ sung tồn, sample, khẩn cấp... |
| `status` | Trạng thái | Xem bảng trạng thái chuẩn |

### 6.1.3. Trạng thái chuẩn `pr_status`
- `draft`: đang soạn
- `submitted`: đã gửi duyệt
- `approved`: đã duyệt
- `rejected`: bị từ chối
- `cancelled`: hủy
- `closed`: đã chuyển hết sang PO hoặc khóa theo rule

### 6.1.4. Quy tắc
- PR `draft` chưa tạo tác động tồn kho hoặc tài chính.
- Chỉ PR `approved` mới được tạo PO.

---

## 6.2. Purchase Order (PO)

### 6.2.1. Định nghĩa
Đơn mua hàng chính thức gửi NCC.

### 6.2.2. Trường dữ liệu chính

| Field | Tên | Mô tả |
|---|---|---|
| `po_no` | Số PO | Mã PO |
| `supplier_code` | NCC | NCC được chọn |
| `po_date` | Ngày PO | Ngày phát hành |
| `expected_receipt_date` | Dự kiến nhận | SLA NCC |
| `currency_code` | Tiền tệ | |
| `total_amount` | Tổng tiền | Trước/hoặc sau thuế tùy rule rõ ràng |
| `status` | Trạng thái | Xem bảng trạng thái chuẩn |

### 6.2.3. Trạng thái chuẩn `po_status`
- `draft`
- `submitted`
- `approved`
- `sent`
- `partially_received`
- `received`
- `closed`
- `cancelled`

### 6.2.4. Quy tắc
- `received` không có nghĩa là hàng đã khả dụng bán/sản xuất; còn phụ thuộc QC.
- Nếu PO có nhiều đợt nhận, dùng `partially_received`.

---

## 6.3. Goods Receipt / Inbound Receipt

### 6.3.1. Định nghĩa
Phiếu nhận hàng vào kho từ NCC.

### 6.3.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `grn_no` | Mã phiếu nhập nhận hàng |
| `po_no` | PO tham chiếu |
| `warehouse_code` | Kho nhập |
| `receipt_date` | Ngày nhận |
| `supplier_batch_no` | Batch NCC nếu có |
| `batch_no` | Batch nội bộ nếu sinh tại nhập |
| `mfg_date` | NSX |
| `expiry_date` | HSD |
| `received_qty` | Số lượng nhận |
| `inventory_status` | Trạng thái tồn ban đầu |

### 6.3.3. Giá trị chuẩn `inventory_status`
- `quarantine`: chờ QC
- `available`: khả dụng
- `hold`: tạm giữ
- `blocked`: khóa sử dụng
- `damaged`: hỏng

### 6.3.4. Quy tắc
- Với item `qc_required = true`, mặc định nhập ở trạng thái `quarantine` hoặc `hold`.
- Không cho để trống `batch_no` nếu item có `lot_controlled = true`.

---

## 6.4. QC Inspection Result

### 6.4.1. Định nghĩa
Kết quả kiểm tra chất lượng cho batch nguyên liệu, bán thành phẩm, thành phẩm.

### 6.4.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `inspection_no` | Mã biên bản QC |
| `reference_doc_no` | Chứng từ tham chiếu |
| `item_code` | Mã item |
| `batch_no` | Batch |
| `inspection_date` | Ngày kiểm |
| `qc_status` | Kết quả QC |
| `release_date` | Ngày release |
| `released_by` | Người release |
| `remark` | Ghi chú |

### 6.4.3. Giá trị chuẩn `qc_status`
- `pending`: chưa kiểm xong
- `hold`: giữ lại để đánh giá thêm
- `pass`: đạt
- `fail`: không đạt
- `waived`: miễn kiểm theo rule hiếm, phải có phê duyệt

### 6.4.4. Quy tắc
- `pass` mới cho phép chuyển `inventory_status` sang `available` nếu item dùng trong bán hoặc sản xuất.
- `fail` không cho phép xuất dùng, trừ luồng xử lý đặc biệt có phê duyệt.

---

## 6.5. Work Order (Lệnh sản xuất)

### 6.5.1. Định nghĩa
Lệnh chính thức để sản xuất một batch bán thành phẩm hoặc thành phẩm.

### 6.5.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `wo_no` | Số lệnh sản xuất |
| `item_code` | Mặt hàng sản xuất |
| `planned_qty` | Số lượng kế hoạch |
| `actual_qty` | Số lượng thực tế |
| `bom_code` | BOM áp dụng |
| `bom_version` | Version BOM |
| `planned_start_date` | Bắt đầu dự kiến |
| `actual_start_date` | Bắt đầu thực tế |
| `actual_end_date` | Kết thúc thực tế |
| `status` | Trạng thái lệnh |
| `prod_batch_no` | Batch thành phẩm/sf |

### 6.5.3. Trạng thái chuẩn `wo_status`
- `draft`
- `released`
- `in_progress`
- `completed`
- `qc_pending`
- `closed`
- `cancelled`

### 6.5.4. Quy tắc
- Lệnh `released` mới cho phép cấp NVL.
- Lệnh `completed` chưa có nghĩa là batch đã bán được; còn qua QC.
- `closed` chỉ dùng khi đã hoàn tất mọi điều chỉnh và batch đã vào trạng thái đúng.

---

## 6.6. Material Issue to Production

### 6.6.1. Định nghĩa
Phiếu xuất nguyên vật liệu từ kho sang lệnh sản xuất.

### 6.6.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `issue_no` | Mã phiếu cấp phát |
| `wo_no` | Lệnh sản xuất |
| `item_code` | Item cấp phát |
| `batch_no` | Batch nguyên liệu |
| `issued_qty` | Số lượng xuất |
| `issue_date` | Ngày xuất |
| `from_warehouse_code` | Kho nguồn |

### 6.6.3. Quy tắc
- Chỉ batch `available` mới được cấp cho sản xuất, trừ rule đặc biệt.
- Nếu áp dụng FEFO, hệ thống nên gợi ý batch có hạn gần nhất nhưng vẫn hợp lệ.

---

## 6.7. Production Output / FG Receipt

### 6.7.1. Định nghĩa
Phiếu nhập bán thành phẩm/thành phẩm sau sản xuất.

### 6.7.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `prod_receipt_no` | Mã phiếu nhập SX |
| `wo_no` | Lệnh sản xuất tham chiếu |
| `item_code` | Item tạo ra |
| `batch_no` | Batch sản xuất |
| `mfg_date` | NSX |
| `expiry_date` | HSD |
| `produced_qty` | Số lượng tạo ra |
| `inventory_status` | Trạng thái tồn ban đầu |

### 6.7.3. Quy tắc
- Batch thành phẩm sau sản xuất mặc định có thể là `hold` hoặc `quarantine` trước khi QC release.

---

## 6.8. Inventory Transaction

### 6.8.1. Định nghĩa
Nhật ký biến động tồn kho chuẩn.

### 6.8.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `txn_no` | Số giao dịch |
| `txn_type` | Loại giao dịch |
| `txn_date` | Ngày giờ giao dịch |
| `warehouse_code` | Kho |
| `item_code` | Item |
| `batch_no` | Batch |
| `qty_in` | Nhập |
| `qty_out` | Xuất |
| `uom` | ĐVT |
| `reference_doc_no` | Chứng từ tham chiếu |
| `inventory_status_before` | Trạng thái trước |
| `inventory_status_after` | Trạng thái sau |

### 6.8.3. Giá trị chuẩn `txn_type`
- `purchase_receipt`
- `prod_issue`
- `prod_receipt`
- `sales_issue`
- `transfer_out`
- `transfer_in`
- `return_in`
- `return_out`
- `adjustment_gain`
- `adjustment_loss`
- `sample_issue`
- `scrap`

### 6.8.4. Quy tắc
- Inventory ledger là bất biến theo nguyên tắc kế toán vận hành; sửa sai bằng adjustment/reversal chứ không xóa dòng cũ.

---

## 6.9. Sales Order (SO)

### 6.9.1. Định nghĩa
Đơn hàng bán ra cho khách hàng/kênh.

### 6.9.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `so_no` | Số đơn bán |
| `customer_code` | Khách hàng |
| `channel_code` | Kênh bán |
| `order_date` | Ngày đặt |
| `requested_ship_date` | Ngày giao yêu cầu |
| `price_list_code` | Bảng giá áp dụng |
| `discount_amount` | Tổng giảm giá |
| `net_amount` | Tiền thuần trước phí hậu mãi khác |
| `status` | Trạng thái đơn |
| `payment_terms` | Điều khoản thanh toán |

### 6.9.3. Trạng thái chuẩn `so_status`
- `draft`
- `submitted`
- `confirmed`
- `reserved`
- `partially_shipped`
- `shipped`
- `delivered`
- `returned`
- `closed`
- `cancelled`

### 6.9.4. Quy tắc
- `confirmed` là trạng thái đã khóa giá/chiết khấu cơ bản theo rule.
- `reserved` là trạng thái đã giữ tồn kho.
- Không cho `shipped` nếu chưa đủ điều kiện xuất hàng.

---

## 6.10. Delivery / Sales Issue

### 6.10.1. Định nghĩa
Phiếu xuất kho giao hàng cho sales order.

### 6.10.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `delivery_no` | Mã phiếu giao |
| `so_no` | Đơn hàng tham chiếu |
| `warehouse_code` | Kho xuất |
| `delivery_date` | Ngày xuất giao |
| `carrier_code` | Đơn vị vận chuyển nếu có |
| `batch_no` | Batch xuất |
| `shipped_qty` | Số lượng giao |
| `status` | Trạng thái giao |

### 6.10.3. Trạng thái chuẩn `delivery_status`
- `draft`
- `picked`
- `packed`
- `shipped`
- `delivered`
- `failed`
- `returned`
- `cancelled`

---

## 6.11. Sales Return

### 6.11.1. Định nghĩa
Phiếu trả hàng từ khách về kho.

### 6.11.2. Trường dữ liệu chính

| Field | Mô tả |
|---|---|
| `return_no` | Mã phiếu trả |
| `so_no` | Đơn gốc |
| `customer_code` | Khách trả |
| `return_reason_code` | Lý do trả |
| `return_qty` | Số lượng trả |
| `return_condition` | Tình trạng hàng trả |
| `return_batch_no` | Batch trả về nếu xác định được |
| `inventory_disposition` | Hướng xử lý tồn sau trả |

### 6.11.3. Giá trị chuẩn `return_condition`
- `sealed_good`
- `opened_good`
- `damaged`
- `expired`
- `suspected_quality_issue`

### 6.11.4. Giá trị chuẩn `inventory_disposition`
- `restock_available`
- `restock_quarantine`
- `scrap`
- `return_to_supplier`
- `hold_investigation`

---

## 7. Trạng thái chuẩn dùng chung

## 7.1. Trạng thái hiệu lực master data `status`
- `draft`: đang tạo, chưa dùng nghiệp vụ chính thức
- `active`: đang có hiệu lực
- `inactive`: ngừng dùng cho giao dịch mới nhưng vẫn giữ lịch sử
- `obsolete`: ngừng hẳn, không dùng mới

## 7.2. Trạng thái tồn kho `inventory_status`
- `available`: khả dụng
- `quarantine`: đang chờ kiểm/điều tra
- `hold`: tạm giữ, chưa được dùng
- `blocked`: khóa hoàn toàn
- `damaged`: hỏng
- `reserved`: đã giữ cho đơn, thường là trạng thái logic hơn là tồn vật lý riêng

## 7.3. Trạng thái QC `qc_status`
- `pending`
- `hold`
- `pass`
- `fail`
- `waived`

---

## 8. Từ điển các field cốt lõi

## 8.1. Tồn kho

### `physical_stock`
**Tên hiển thị:** Tồn vật lý  
**Định nghĩa:** Tổng lượng hàng đang hiện diện vật lý trong kho ở thời điểm tính, chưa trừ hàng giữ hoặc chặn logic.  
**Gồm:** available + quarantine + hold + damaged + reserved logic tùy mô hình lưu.  
**Không đồng nghĩa với hàng bán được.**

### `available_stock`
**Tên hiển thị:** Tồn khả dụng  
**Định nghĩa:** Lượng hàng đủ điều kiện để cấp cho đơn bán hoặc sản xuất theo rule hiện hành.  
**Công thức gợi ý:**

```text
available_stock = physical_stock - reserved_stock - hold_stock - quarantine_stock - blocked_stock - damaged_stock
```

### `reserved_stock`
**Tên hiển thị:** Hàng đã giữ  
**Định nghĩa:** Lượng hàng đã được giữ logic cho đơn bán/đơn chuyển kho/lệnh khác, chưa xuất vật lý.

### `quarantine_stock`
**Tên hiển thị:** Hàng chờ QC  
**Định nghĩa:** Hàng đã nhập kho vật lý nhưng chưa được coi là khả dụng.

### `hold_stock`
**Tên hiển thị:** Hàng hold  
**Định nghĩa:** Hàng đang tạm giữ do QC, điều tra, nghi ngờ, chờ phê duyệt.

### `near_expiry_stock`
**Tên hiển thị:** Hàng cận date  
**Định nghĩa:** Hàng có `expiry_date` nhỏ hơn hoặc bằng ngưỡng cảnh báo cấu hình.  
**Công thức gợi ý:**

```text
near_expiry_stock = các batch có expiry_date - current_date <= near_expiry_threshold_days
```

## 8.2. Batch và hạn dùng

### `batch_no`
**Tên hiển thị:** Mã lô  
**Định nghĩa:** Mã nhận diện một lô hàng đồng nhất theo nguồn gốc và/hoặc sản xuất.  
**Bắt buộc** cho item có `lot_controlled = true`.

### `mfg_date`
**Tên hiển thị:** Ngày sản xuất  
**Định nghĩa:** Ngày sản xuất của batch.

### `expiry_date`
**Tên hiển thị:** Hạn sử dụng  
**Định nghĩa:** Ngày hết hạn của batch. Sau ngày này batch không được bán hay dùng sản xuất trừ rule đặc biệt được cấm mặc định.

### `shelf_life_days`
**Tên hiển thị:** Số ngày hạn dùng chuẩn  
**Định nghĩa:** Chuẩn tính tuổi thọ sản phẩm từ NSX. Chỉ dùng để gợi ý, không thay thế hồ sơ pháp lý/QA.

## 8.3. Sản xuất

### `planned_qty`
Số lượng kế hoạch sẽ mua/sản xuất/giao.

### `actual_qty`
Số lượng thực tế đã thực hiện.

### `yield_standard`
Tỷ lệ thu hồi chuẩn theo BOM hoặc quy trình chuẩn.

### `yield_actual`
Tỷ lệ thu hồi thực tế.  
**Công thức gợi ý:**

```text
yield_actual = actual_output_qty / theoretical_output_qty
```

### `loss_standard_pct`
Tỷ lệ hao hụt chuẩn.

### `loss_actual_pct`
Tỷ lệ hao hụt thực tế.  
**Công thức gợi ý:**

```text
loss_actual_pct = (actual_input_qty - equivalent_output_qty) / actual_input_qty
```

> Lưu ý: cách quy đổi `equivalent_output_qty` phải được chuẩn hóa theo từng nhóm sản phẩm và UOM.

## 8.4. Giá và lợi nhuận

### `standard_cost`
Chi phí chuẩn dùng để tham chiếu kế hoạch, định giá sơ bộ, không phải lúc nào cũng là giá vốn kế toán cuối cùng.

### `actual_batch_cost`
Tổng chi phí thực tế gắn với một batch sản xuất theo logic Phase 1. Tối thiểu gồm nguyên vật liệu và packaging; có thể mở rộng sau.

### `list_price`
Giá niêm yết hoặc giá chuẩn trong bảng giá.

### `discount_amount`
Tổng số tiền giảm giá trên dòng hoặc đơn.

### `net_sales_amount`
Doanh thu thuần trước chi phí vận chuyển tài trợ, hoàn hàng, rebate hậu kỳ nếu Phase 1 chưa xử lý đầy đủ.

### `gross_margin`
Biên lợi nhuận gộp gần đúng.  
**Công thức gợi ý:**

```text
gross_margin = (net_sales_amount - cogs_estimated) / net_sales_amount
```

> Trong Phase 1 có thể dùng `cogs_estimated` hoặc batch cost gần đúng. Giai đoạn sau sẽ tinh chỉnh sâu hơn.

---

## 9. Quy tắc dữ liệu bắt buộc

## 9.1. Quy tắc chung
- Mọi mã (`code`) phải duy nhất trong domain của nó.
- Tên chuẩn không được để khoảng trắng vô nghĩa ở đầu/cuối.
- Ngày phải dùng timezone thống nhất theo hệ thống.
- Số lượng không được âm trừ các giao dịch điều chỉnh hoặc reversal có kiểm soát.

## 9.2. Quy tắc cho item
- `item_code`, `item_name`, `item_type`, `uom_base`, `status` là bắt buộc.
- Item `finished_good` phải có `is_sellable = true` hoặc được cấu hình rõ vì sao không bán.
- Item `raw_material` dùng sản xuất phải có `is_purchasable = true` hoặc có nguồn nội bộ khác được định nghĩa.

## 9.3. Quy tắc cho batch
- Không cho tạo batch trùng `batch_no` trong cùng `item_code` nếu nghiệp vụ không cho phép.
- `expiry_date` phải lớn hơn `mfg_date`.
- Batch hết hạn không được `available`.

## 9.4. Quy tắc cho sales
- Không cho xác nhận đơn nếu thiếu `customer_code`, `channel_code`, `order_date`.
- Không cho xuất hàng vượt `available_stock` nếu không có rule override.

## 9.5. Quy tắc cho purchasing
- Không cho PO thiếu `supplier_code`, `po_date`, `currency_code` nếu nghiệp vụ có tiền tệ.
- Không cho nhận hàng vượt tolerance so với PO nếu không có phê duyệt.

---

## 10. Rulebook master data ngành mỹ phẩm

## 10.1. Batch Rule
- Thành phẩm, bán thành phẩm, nguyên liệu nhạy cảm phải quản lý batch.
- Batch phải truy ngược được ít nhất về: nguồn nhập hoặc lệnh sản xuất.
- Complaint chất lượng phải có khả năng gắn về `batch_no`.

## 10.2. Expiry Rule
- Item có hạn dùng phải bật `expiry_controlled = true`.
- Cảnh báo cận date dùng ngưỡng cấu hình theo loại item.
- FEFO là khuyến nghị chuẩn cho hàng có hạn dùng khi xuất kho.

## 10.3. FEFO/FIFO Rule
- Mặc định ưu tiên FEFO cho thành phẩm và nguyên liệu có hạn dùng.
- Nếu nhiều batch cùng hạn, áp dụng FIFO theo ngày nhập.

## 10.4. Sample/Tester/Gift Rule
- Sample/tester/quà tặng không được trộn logic với hàng bán bình thường nếu công ty cần theo dõi thất thoát và ROI marketing.
- Nên theo một trong hai mô hình:
  - Item riêng
  - Hoặc kho/trạng thái riêng có rule rõ

## 10.5. Quarantine/Hold Rule
- Hàng `quarantine` hoặc `hold` không được dùng cho bán/sản xuất mặc định.
- Muốn chuyển trạng thái phải có chứng từ hoặc action QC/approval được log.

## 10.6. Claim/Spec Rule
- Mọi spec QC và công thức/BOM phải có version.
- Không được dùng cùng lúc hai version active chồng chéo không có rule rõ.

---

## 11. Công thức chuẩn của hệ thống

## 11.1. Tồn khả dụng

```text
available_stock = physical_stock - reserved_stock - quarantine_stock - hold_stock - blocked_stock - damaged_stock
```

## 11.2. Tỷ lệ hoàn thành lệnh sản xuất

```text
wo_completion_pct = actual_output_qty / planned_qty
```

## 11.3. Tỷ lệ hao hụt sản xuất gần đúng

```text
production_loss_pct = (actual_input_qty - standard_equivalent_output_qty) / actual_input_qty
```

## 11.4. Fill Rate giao hàng

```text
fill_rate = shipped_qty / confirmed_order_qty
```

## 11.5. Tỷ lệ giao đúng hạn

```text
on_time_delivery_rate = số đơn giao đúng/ngày cam kết / tổng số đơn giao
```

## 11.6. Tỷ lệ batch fail QC

```text
batch_fail_rate = số batch fail / tổng số batch kiểm
```

## 11.7. Tỷ lệ cận date

```text
near_expiry_rate = near_expiry_stock / physical_stock
```

---

## 12. Field-level governance

## 12.1. Chỉ đọc sau khi chốt
Các trường sau nên bị khóa sửa trực tiếp sau khi giao dịch sang trạng thái chính thức:
- `item_code`
- `supplier_code` trên PO đã approved
- `customer_code` trên SO đã confirmed
- `batch_no`
- `mfg_date`
- `expiry_date`
- `bom_version` của WO đã released
- `list_price` đã resolve trên đơn confirmed

## 12.2. Chỉ role chuyên trách được sửa
- `qc_status`: QA/QC
- `standard_cost`: finance/costing/admin được ủy quyền
- `price_list_code`, `list_price`, `min_price`: sales admin/commercial admin
- `payment_terms`: finance/commercial theo rule

## 12.3. Break-glass / sửa khẩn cấp
Nếu cho phép sửa khẩn cấp, phải bắt buộc:
- Lý do sửa
- Người duyệt
- Audit log trước/sau
- Thời điểm sửa
- Ticket hoặc số biên bản tham chiếu

---

## 13. Data quality checklist trước go-live

### 13.1. Item master
- Không trùng mã
- Không trùng tên gây nhầm lẫn
- Đủ UOM
- Đúng `item_type`
- Đúng bật/tắt `lot_controlled`, `expiry_controlled`, `qc_required`

### 13.2. Supplier master
- Có mã NCC chuẩn
- Không trùng NCC do khác cách viết tên
- Có lead time và payment terms cơ bản

### 13.3. Warehouse master
- Đúng loại kho
- Đúng rule cho xuất bán/sản xuất/hold

### 13.4. BOM
- Không thiếu dòng vật tư chính
- Không có 2 version active chồng lấn trái rule
- UOM quy đổi đúng

### 13.5. Price list
- Mỗi SKU có giá đúng theo kênh cần dùng Phase 1
- Không có khoảng thời gian hiệu lực chồng chéo không kiểm soát

### 13.6. Opening stock
- Có batch cho item cần batch
- Có NSX/HSD cho item cần hạn dùng
- Đã tách rõ available/quarantine/hold/damaged

---

## 14. Mapping tối thiểu cho migration dữ liệu

## 14.1. Item Master
Tối thiểu cần map:
- `legacy_item_code` -> `item_code`
- `legacy_item_name` -> `item_name`
- `legacy_uom` -> `uom_base`
- `legacy_category` -> `item_group`
- `legacy_status` -> `status`

## 14.2. Opening Stock
Tối thiểu cần map:
- `warehouse_code`
- `item_code`
- `batch_no`
- `mfg_date`
- `expiry_date`
- `qty`
- `inventory_status`

## 14.3. Customer/Supplier
- Chuẩn hóa mã trùng
- Chuẩn hóa tên pháp lý/tên hiển thị
- Chuẩn hóa payment terms
- Chuẩn hóa channel/customer type

---

## 15. Danh mục enum khuyến nghị tổng hợp

## 15.1. `item_type`
- `raw_material`
- `packaging`
- `semi_fg`
- `finished_good`
- `service`
- `gift`

## 15.2. `qc_status`
- `pending`
- `hold`
- `pass`
- `fail`
- `waived`

## 15.3. `inventory_status`
- `available`
- `quarantine`
- `hold`
- `blocked`
- `damaged`
- `reserved`

## 15.4. `pr_status`
- `draft`
- `submitted`
- `approved`
- `rejected`
- `cancelled`
- `closed`

## 15.5. `po_status`
- `draft`
- `submitted`
- `approved`
- `sent`
- `partially_received`
- `received`
- `closed`
- `cancelled`

## 15.6. `wo_status`
- `draft`
- `released`
- `in_progress`
- `completed`
- `qc_pending`
- `closed`
- `cancelled`

## 15.7. `so_status`
- `draft`
- `submitted`
- `confirmed`
- `reserved`
- `partially_shipped`
- `shipped`
- `delivered`
- `returned`
- `closed`
- `cancelled`

---

## 16. Quyết định thiết kế cần chốt sớm với đội dự án

Đây là các câu hỏi phải chốt sớm trước khi code sâu:
- `reserved_stock` là trạng thái tồn hay chỉ là số logic tính toán?
- Batch number có cho phép trùng khác site hay không?
- Tồn khả dụng có trừ `damaged` trực tiếp hay tách logic?
- Thành phẩm sau sản xuất vào `hold` hay `quarantine` mặc định?
- Sample/tester quản theo item riêng hay warehouse/status riêng?
- COGS Phase 1 tính theo standard cost, batch cost hay moving average gần đúng?
- Giá bán resolve theo priority nào khi có nhiều rule chồng nhau?

Nếu chưa chốt rõ các câu này, dev rất dễ code sai xương sống hệ thống.

---

## 17. Kết luận

Tài liệu này là lớp nền để cả hệ thống nói chung một ngôn ngữ. Nếu Blueprint là bản đồ, PRD/SRS là bản vẽ thi công, thì **Data Dictionary + Master Data Rulebook là bộ quy ước vật liệu, kích thước và nhãn trên công trường**.

Khi tài liệu này được chốt, đội dự án sẽ giảm mạnh các lỗi kiểu:
- hiểu sai nghĩa dữ liệu
- code sai trạng thái nghiệp vụ
- báo cáo không khớp do mỗi bộ phận hiểu khác nhau
- migration dữ liệu bẩn
- tồn kho và batch đi lệch ngay từ đầu

---

## 18. Tài liệu liên quan nên làm tiếp
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `07_ERP_Report_KPI_Catalog_Phase1_My_Pham_v1.md`
- `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`
- `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`

