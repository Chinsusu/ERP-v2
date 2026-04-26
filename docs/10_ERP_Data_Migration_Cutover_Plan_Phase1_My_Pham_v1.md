# 10_ERP_Data_Migration_Cutover_Plan_Phase1_My_Pham_v1

**Dự án:** Website ERP cho công ty sản xuất, phân phối và bán lẻ mỹ phẩm  
**Giai đoạn:** Phase 1 – Core ERP  
**Phiên bản:** v1.0  
**Mục tiêu tài liệu:** Chuẩn hóa kế hoạch chuyển dữ liệu, kiểm tra dữ liệu đầu kỳ, chạy thử, khóa số và go-live ERP Phase 1.

---

## 1. Mục tiêu của Data Migration & Cutover

Data Migration là quá trình đưa dữ liệu hiện tại của công ty từ Excel, phần mềm cũ, file kế toán, file kho, file sản xuất, file sales vào ERP mới.

Cutover là giai đoạn chuyển từ cách vận hành cũ sang hệ thống ERP mới.

Mục tiêu không phải chỉ là “import dữ liệu vào hệ thống”, mà là đảm bảo:

- Mã hàng đúng.
- Tồn kho đúng.
- Batch/lô đúng.
- Hạn sử dụng đúng.
- Công nợ đầu kỳ đúng.
- Danh mục khách hàng/nhà cung cấp đúng.
- BOM/công thức đúng phiên bản.
- Người dùng biết ngày nào dừng file cũ, ngày nào bắt đầu dùng ERP.

Một câu rất quan trọng:

> ERP go-live không chết vì thiếu tính năng nhiều bằng chết vì dữ liệu đầu vào sai.

---

## 2. Phạm vi dữ liệu cần chuyển trong Phase 1

Phase 1 tập trung vào 6 module lõi:

1. Dữ liệu gốc / Master Data
2. Mua hàng / Purchasing
3. QA/QC
4. Sản xuất / Production
5. Kho / Inventory
6. Bán hàng / Sales Order

Dữ liệu cần chuyển gồm:

| Nhóm dữ liệu | Có migrate Phase 1? | Ghi chú |
|---|---:|---|
| Danh mục nguyên vật liệu | Có | Bắt buộc |
| Danh mục bao bì, tem nhãn, phụ liệu | Có | Bắt buộc |
| Danh mục bán thành phẩm | Có | Nếu có thực tế sản xuất |
| Danh mục thành phẩm/SKU | Có | Bắt buộc |
| BOM/công thức cơ bản | Có | Ít nhất bản đang active |
| Nhà cung cấp | Có | Bắt buộc cho mua hàng |
| Khách hàng/đại lý | Có | Bắt buộc cho sales |
| Kho và vị trí kho | Có | Bắt buộc |
| Tồn kho đầu kỳ | Có | Bắt buộc |
| Batch/lô/hạn dùng đầu kỳ | Có | Bắt buộc với mỹ phẩm |
| Công nợ khách hàng đầu kỳ | Có | Nếu Phase 1 có AR cơ bản |
| Công nợ nhà cung cấp đầu kỳ | Có | Nếu Phase 1 có AP cơ bản |
| Đơn hàng đang mở | Có chọn lọc | Chỉ migrate đơn chưa hoàn tất |
| PO đang mở | Có chọn lọc | Chỉ migrate PO chưa nhận đủ/chưa đóng |
| Lệnh sản xuất đang chạy | Có chọn lọc | Nếu có tại thời điểm cutover |
| Hồ sơ QC lịch sử | Không bắt buộc | Có thể lưu file tham chiếu |
| Giao dịch kho lịch sử nhiều năm | Không nên | Chỉ cần số dư đầu kỳ + file archive |
| Dữ liệu HR/KOL/CRM sâu | Không | Phase sau |

---

## 3. Nguyên tắc migration

### 3.1. Không migrate rác vào hệ thống mới

Trước khi import, dữ liệu phải được làm sạch.

Không đưa vào ERP các dữ liệu:

- Mã hàng trùng.
- Tên hàng viết nhiều kiểu.
- Đơn vị tính không thống nhất.
- Batch không có hạn dùng.
- Tồn kho âm không có lý do.
- Khách hàng trùng số điện thoại/mã số thuế.
- Nhà cung cấp không còn giao dịch nhưng chưa phân loại inactive.

### 3.2. Không cố migrate toàn bộ lịch sử

Phase 1 nên migrate:

- Master data đang dùng.
- Tồn kho đầu kỳ.
- Công nợ đầu kỳ.
- Chứng từ đang mở.

Lịch sử cũ có thể lưu dưới dạng file archive để tra cứu, không nhất thiết đưa hết vào ERP.

### 3.3. Mỗi dữ liệu phải có owner xác nhận

Không để dev tự đoán dữ liệu.

| Dữ liệu | Owner xác nhận |
|---|---|
| Mã hàng/SKU | Product/R&D + Kho + Sales |
| BOM/công thức | R&D + Production |
| Tồn kho | Warehouse Manager |
| Batch/hạn dùng | Warehouse + QA/QC |
| Nhà cung cấp | Purchasing |
| Khách hàng/đại lý | Sales Admin |
| Công nợ | Finance/Accounting |
| Giá bán/chiết khấu | Sales + Finance |

### 3.4. Tất cả file import phải có version

Ví dụ:

```text
MIG_ITEM_MASTER_v1_2026-04-24.xlsx
MIG_ITEM_MASTER_v2_CLEANED_2026-04-26.xlsx
MIG_OPENING_STOCK_FINAL_SIGNED_2026-04-30.xlsx
```

Không dùng file tên kiểu:

```text
hanghoa_moi_nhat.xlsx
file chot.xlsx
file chot lan cuoi.xlsx
```

### 3.5. Có biên bản sign-off trước khi import final

Mỗi nhóm dữ liệu cần có xác nhận:

- Người chuẩn bị dữ liệu.
- Người kiểm tra.
- Người duyệt nghiệp vụ.
- Ngày chốt dữ liệu.
- Phiên bản file.

---

## 4. Chiến lược migrate đề xuất

Tao khuyên dùng chiến lược 4 vòng:

```text
Vòng 1: Data Discovery
→ Vòng 2: Data Cleansing
→ Vòng 3: Trial Migration
→ Vòng 4: Final Migration + Cutover
```

### Vòng 1: Data Discovery

Mục tiêu: biết dữ liệu hiện đang nằm ở đâu, ai giữ, chất lượng ra sao.

Cần làm:

- Thu thập file hiện tại.
- Liệt kê nguồn dữ liệu.
- Kiểm tra dữ liệu trùng/lệch.
- Xác định dữ liệu nào bắt buộc migrate.
- Xác định dữ liệu nào chỉ archive.

Output:

- Source Data Inventory.
- Data Gap List.
- Migration Scope Confirmed.

### Vòng 2: Data Cleansing

Mục tiêu: làm sạch dữ liệu trước khi import.

Cần làm:

- Chuẩn hóa mã hàng.
- Chuẩn hóa đơn vị tính.
- Gộp khách hàng trùng.
- Gộp nhà cung cấp trùng.
- Rà batch/hạn dùng.
- Kiểm kho thực tế.
- Xử lý tồn âm, tồn treo, tồn không rõ nguồn.

Output:

- Clean Master Data.
- Clean Opening Balance.
- Issue Log.

### Vòng 3: Trial Migration

Mục tiêu: import thử vào môi trường test.

Cần làm:

- Import master data.
- Import tồn kho thử.
- Import công nợ thử.
- Import chứng từ mở thử.
- So sánh số sau import với file gốc.
- Cho key user kiểm tra.

Output:

- Trial Migration Report.
- Error Log.
- Reconciliation Report.

### Vòng 4: Final Migration + Cutover

Mục tiêu: chốt dữ liệu thật và chuyển sang dùng ERP.

Cần làm:

- Chốt ngày khóa số.
- Dừng nhập liệu ở hệ thống/file cũ.
- Kiểm kê nhanh nếu cần.
- Import dữ liệu final.
- Reconcile số cuối cùng.
- Business sign-off.
- Go-live.

Output:

- Final Migration Sign-off.
- Go-live Confirmation.
- Post Go-live Issue Log.

---

## 5. Nguồn dữ liệu đầu vào cần thu thập

| STT | Nguồn dữ liệu | Người giữ | Định dạng hiện tại | Ghi chú |
|---:|---|---|---|---|
| 1 | Danh mục hàng hóa | Kho/Sales/Admin | Excel/phần mềm cũ | Cần chuẩn mã |
| 2 | Danh mục nguyên liệu | Kho/R&D/Purchasing | Excel | Cần đơn vị tính chuẩn |
| 3 | BOM/công thức | R&D/Sản xuất | Excel/tài liệu riêng | Cần version active |
| 4 | Tồn kho hiện tại | Kho | Excel/phần mềm | Cần đối chiếu vật lý |
| 5 | Batch/hạn dùng | Kho/QA | Excel/sổ theo dõi | Bắt buộc |
| 6 | Nhà cung cấp | Purchasing/Finance | Excel/kế toán | Cần mã NCC |
| 7 | Khách hàng/đại lý | Sales/CSKH | Excel/POS/CRM cũ | Cần xử lý trùng |
| 8 | Công nợ khách hàng | Finance | Phần mềm kế toán | Cần số dư chốt |
| 9 | Công nợ NCC | Finance | Phần mềm kế toán | Cần số dư chốt |
| 10 | PO đang mở | Purchasing | Excel/email | Chỉ PO chưa hoàn tất |
| 11 | SO đang mở | Sales | Excel/phần mềm bán hàng | Chỉ đơn chưa hoàn tất |
| 12 | Lệnh sản xuất đang chạy | Production | Excel/sổ | Chỉ nếu có |

---

## 6. Template dữ liệu migrate

### 6.1. Item Master – Danh mục hàng hóa/nguyên liệu

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| item_code | Có | RM-GRAPEFRUIT-OIL-001 | Mã duy nhất |
| item_name | Có | Tinh dầu vỏ bưởi | Tên chuẩn |
| item_type | Có | raw_material / packaging / semi_finished / finished_goods | Loại hàng |
| uom_base | Có | kg / liter / pcs | Đơn vị gốc |
| uom_purchase | Không | box | Nếu khác đơn vị gốc |
| conversion_rate | Không | 1 box = 12 pcs | Quy đổi |
| category | Có | Active / Packaging / Shampoo | Nhóm hàng |
| brand | Không | VyVy | Nếu là thành phẩm |
| shelf_life_months | Có nếu có hạn | 24 | Mỹ phẩm thường cần |
| qc_required | Có | yes/no | Có cần QC không |
| status | Có | active/inactive | Trạng thái |

### 6.2. Product/SKU Master – Thành phẩm

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| sku_code | Có | FG-FF-SPRAY-100ML | Mã SKU |
| sku_name | Có | Fast & Furious Hair Spray 100ml | Tên bán hàng |
| brand | Có | VyVy | Brand |
| product_line | Không | Haircare | Dòng hàng |
| default_uom | Có | pcs | Đơn vị bán |
| barcode | Không | 893... | Nếu có |
| standard_price | Không | 230000 | Giá niêm yết |
| standard_cost | Không | 85000 | Nếu có |
| bom_code | Có nếu tự sản xuất | BOM-FF-100ML-v1 | BOM active |
| qc_required | Có | yes | Bắt buộc với thành phẩm |
| shelf_life_months | Có | 24 | Hạn dùng |
| status | Có | active | Trạng thái |

### 6.3. BOM Master – Công thức/định mức

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| bom_code | Có | BOM-FF-100ML-v1 | Mã BOM |
| sku_code | Có | FG-FF-SPRAY-100ML | Thành phẩm |
| version | Có | v1 | Version |
| effective_from | Có | 2026-05-01 | Ngày hiệu lực |
| component_code | Có | RM-REDENSYL-001 | Thành phần |
| component_type | Có | raw_material / packaging | Loại thành phần |
| qty_standard | Có | 0.005 | Định mức |
| uom | Có | kg | Đơn vị |
| wastage_rate | Không | 2% | Hao hụt chuẩn |
| is_active | Có | yes/no | BOM đang dùng |

### 6.4. Supplier Master – Nhà cung cấp

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| supplier_code | Có | SUP-0001 | Mã NCC |
| supplier_name | Có | ABC Ingredient Co. | Tên NCC |
| tax_code | Không | 031... | MST nếu có |
| contact_name | Không | Ms. Lan | Người liên hệ |
| phone | Không | 090... | SĐT |
| email | Không | purchasing@... | Email |
| address | Không | HCM | Địa chỉ |
| payment_term | Không | 30 days | Điều khoản thanh toán |
| lead_time_days | Không | 20 | Thời gian giao |
| status | Có | active/inactive | Trạng thái |

### 6.5. Customer Master – Khách hàng/đại lý

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| customer_code | Có | CUS-0001 | Mã KH |
| customer_name | Có | Đại lý Minh Anh | Tên KH |
| customer_type | Có | retail / dealer / distributor / ecommerce | Loại KH |
| phone | Không | 090... | Dùng kiểm trùng |
| tax_code | Không | 031... | Nếu B2B |
| address | Không | HCM | Địa chỉ |
| sales_owner | Không | NVKD01 | Người phụ trách |
| payment_term | Không | COD / 15 days / 30 days | Công nợ |
| price_list | Không | DEALER-LV1 | Bảng giá |
| credit_limit | Không | 50000000 | Hạn mức nợ |
| status | Có | active/inactive | Trạng thái |

### 6.6. Warehouse Master – Kho/vị trí

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| warehouse_code | Có | WH-HCM-MAIN | Mã kho |
| warehouse_name | Có | Kho trung tâm HCM | Tên kho |
| warehouse_type | Có | raw / finished / retail / quarantine / sample | Loại kho |
| location_code | Không | A01-B02 | Vị trí/bin |
| status | Có | active | Trạng thái |

### 6.7. Opening Stock – Tồn kho đầu kỳ

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| warehouse_code | Có | WH-HCM-MAIN | Kho |
| location_code | Không | A01-B02 | Vị trí |
| item_code | Có | FG-FF-SPRAY-100ML | Mã hàng |
| batch_no | Có với mỹ phẩm | FF26050101 | Mã lô |
| manufacturing_date | Không | 2026-05-01 | NSX |
| expiry_date | Có | 2028-05-01 | HSD |
| qc_status | Có | pass / hold / fail | Trạng thái QC |
| stock_status | Có | available / quarantine / damaged / sample | Loại tồn |
| quantity | Có | 1200 | Số lượng |
| uom | Có | pcs | Đơn vị |
| unit_cost | Không | 85000 | Nếu có giá vốn đầu kỳ |

### 6.8. Opening AR – Công nợ khách hàng đầu kỳ

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| customer_code | Có | CUS-0001 | Khách hàng |
| document_no | Có | AR-OPEN-0001 | Chứng từ |
| document_date | Có | 2026-04-30 | Ngày chốt |
| due_date | Không | 2026-05-30 | Hạn thanh toán |
| amount_original | Có | 20000000 | Số tiền gốc |
| amount_remaining | Có | 12000000 | Còn nợ |
| currency | Có | VND | Tiền tệ |
| note | Không | Công nợ đầu kỳ | Ghi chú |

### 6.9. Opening AP – Công nợ NCC đầu kỳ

| Field | Bắt buộc | Ví dụ | Ghi chú |
|---|---:|---|---|
| supplier_code | Có | SUP-0001 | NCC |
| document_no | Có | AP-OPEN-0001 | Chứng từ |
| document_date | Có | 2026-04-30 | Ngày chốt |
| due_date | Không | 2026-05-30 | Hạn thanh toán |
| amount_original | Có | 50000000 | Số tiền gốc |
| amount_remaining | Có | 30000000 | Còn nợ |
| currency | Có | VND | Tiền tệ |
| note | Không | Công nợ đầu kỳ | Ghi chú |

---

## 7. Data Cleansing Rules – Quy tắc làm sạch dữ liệu

### 7.1. Quy tắc mã hàng

- Mỗi mã hàng là duy nhất.
- Không dùng dấu tiếng Việt trong mã.
- Không dùng khoảng trắng trong mã.
- Không dùng mã quá chung như `SP001` nếu có thể chuẩn hóa tốt hơn.
- Không tái sử dụng mã cũ cho hàng khác.

Ví dụ tốt:

```text
RM-GRAPEFRUIT-OIL-001
PK-BOTTLE-100ML-CLEAR-001
FG-FF-SPRAY-100ML
```

### 7.2. Quy tắc tên hàng

Tên hàng nên có cấu trúc:

```text
[Loại hàng] + [Tên chính] + [Dung tích/Quy cách] + [Phiên bản nếu cần]
```

Ví dụ:

```text
Chai nhựa trong 100ml cổ xịt
Fast & Furious Hair Spray 100ml
Tinh dầu vỏ bưởi loại A
```

### 7.3. Quy tắc đơn vị tính

Phải có đơn vị gốc cho từng item.

Ví dụ:

| Loại hàng | Đơn vị gốc đề xuất |
|---|---|
| Chất lỏng | liter hoặc kg |
| Bột | kg |
| Chai/lọ/vỏ hộp | pcs |
| Tem nhãn | pcs |
| Thùng carton | pcs |
| Thành phẩm | pcs |

Nếu mua theo thùng nhưng xuất theo chai, bắt buộc có conversion rate.

### 7.4. Quy tắc batch/hạn dùng

- Thành phẩm mỹ phẩm bắt buộc có batch và expiry_date.
- Nguyên liệu có hạn dùng cũng bắt buộc có batch và expiry_date.
- Batch không có hạn dùng phải được QA/Warehouse xác nhận lý do.
- Batch fail/hold không được đưa vào tồn khả dụng.
- Batch damaged không được bán.

### 7.5. Quy tắc khách hàng trùng

Dùng các trường để kiểm trùng:

- Số điện thoại.
- Mã số thuế.
- Email.
- Tên + địa chỉ.

Nếu trùng, cần xác định:

- Gộp vào một customer code.
- Hay tách thành nhiều chi nhánh/điểm giao.

### 7.6. Quy tắc nhà cung cấp trùng

Dùng:

- Tên pháp lý.
- Mã số thuế.
- Email/phone.
- Nhóm hàng cung cấp.

Không tạo nhiều NCC cho cùng một pháp nhân nếu chỉ khác người liên hệ.

### 7.7. Quy tắc xử lý tồn âm

Tồn âm không được import vào ERP.

Cách xử lý:

1. Kiểm tra sai lệch nhập/xuất.
2. Kiểm kê vật lý.
3. Nếu là sai lệch lịch sử, tạo adjustment trước cutover.
4. Chỉ import số tồn đã chốt >= 0.

### 7.8. Quy tắc xử lý hàng không rõ batch

Nếu hàng vật lý còn nhưng không rõ batch:

- Tạo batch tạm có tiền tố `UNKNOWN-`.
- QA/Warehouse xác nhận tình trạng.
- Đánh dấu `stock_status = quarantine` hoặc `restricted` nếu chưa đủ cơ sở bán.
- Không tự động cho bán nếu không có hạn dùng rõ.

---

## 8. Mapping dữ liệu từ hệ cũ sang ERP mới

| Hệ cũ/File cũ | Trường cũ | ERP field | Mapping rule | Owner |
|---|---|---|---|---|
| File hàng hóa | Mã SP | item_code/sku_code | Chuẩn hóa theo quy tắc mã | Product/Kho |
| File hàng hóa | Tên SP | item_name/sku_name | Chuẩn hóa tên | Product/Kho |
| File kho | SL tồn | quantity | Chỉ lấy sau kiểm kê | Kho |
| File kho | Lô | batch_no | Bắt buộc nếu có | Kho/QA |
| File kho | HSD | expiry_date | Format YYYY-MM-DD | Kho/QA |
| File NCC | Tên NCC | supplier_name | Loại bỏ trùng | Purchasing |
| File KH | Tên KH | customer_name | Gộp trùng | Sales |
| Kế toán | Nợ KH | opening_ar | Theo số dư đã chốt | Finance |
| Kế toán | Nợ NCC | opening_ap | Theo số dư đã chốt | Finance |

---

## 9. Validation Rules – Quy tắc kiểm tra trước import

### 9.1. Item Master

- `item_code` không được trống.
- `item_code` không được trùng.
- `item_type` phải thuộc danh sách chuẩn.
- `uom_base` phải thuộc danh sách đơn vị đã cấu hình.
- Nếu `qc_required = yes`, item phải có QC rule cơ bản.
- Nếu có hạn dùng, `shelf_life_months` phải > 0.

### 9.2. BOM

- `bom_code` không trống.
- `sku_code` phải tồn tại trong Product Master.
- `component_code` phải tồn tại trong Item Master.
- `qty_standard` > 0.
- Chỉ có một BOM active cho một SKU tại một thời điểm, trừ khi có quy tắc version rõ.

### 9.3. Opening Stock

- `warehouse_code` phải tồn tại.
- `item_code` phải tồn tại.
- `quantity` >= 0.
- Thành phẩm phải có `batch_no` và `expiry_date`.
- `expiry_date` không được nhỏ hơn ngày go-live nếu stock_status = available.
- `qc_status = pass` mới được đưa vào available stock.
- Tổng tồn theo ERP sau import phải khớp biên bản kiểm kê.

### 9.4. Customer Master

- `customer_code` không trống.
- Không trùng mã.
- Nếu là B2B, nên có tax_code hoặc địa chỉ rõ.
- `payment_term` phải thuộc danh sách chuẩn.

### 9.5. Supplier Master

- `supplier_code` không trống.
- Không trùng mã.
- `payment_term` phải thuộc danh sách chuẩn nếu có.
- `lead_time_days` >= 0 nếu có.

### 9.6. Opening AR/AP

- Customer/Supplier phải tồn tại.
- `amount_remaining` >= 0.
- Tổng AR/AP sau import phải khớp báo cáo finance đã sign-off.

---

## 10. Reconciliation – Đối chiếu sau import

Sau mỗi lần import thử/final, phải đối chiếu.

### 10.1. Đối chiếu master data

| Chỉ số | File nguồn | ERP | Sai lệch | Owner xác nhận |
|---|---:|---:|---:|---|
| Số mã nguyên liệu |  |  |  | Kho/R&D |
| Số mã thành phẩm |  |  |  | Product/Sales |
| Số NCC active |  |  |  | Purchasing |
| Số KH active |  |  |  | Sales |
| Số BOM active |  |  |  | R&D/Production |

### 10.2. Đối chiếu tồn kho

| Nhóm tồn | File chốt | ERP | Sai lệch | Ghi chú |
|---|---:|---:|---:|---|
| Nguyên liệu |  |  |  |  |
| Bao bì |  |  |  |  |
| Bán thành phẩm |  |  |  |  |
| Thành phẩm |  |  |  |  |
| Sample/tester |  |  |  |  |
| Quarantine/hold |  |  |  |  |

### 10.3. Đối chiếu theo batch/hạn dùng

| SKU/Item | Batch | HSD | File chốt | ERP | Sai lệch |
|---|---|---|---:|---:|---:|
|  |  |  |  |  |  |

### 10.4. Đối chiếu công nợ

| Nhóm | File finance | ERP | Sai lệch | Owner |
|---|---:|---:|---:|---|
| Tổng phải thu |  |  |  | Finance |
| Tổng phải trả |  |  |  | Finance |
| Số khách có công nợ |  |  |  | Finance/Sales |
| Số NCC có công nợ |  |  |  | Finance/Purchasing |

---

## 11. Cutover Strategy – Kế hoạch chuyển đổi hệ thống

### 11.1. Nguyên tắc cutover

- Có ngày khóa dữ liệu rõ ràng.
- Có người chịu trách nhiệm từng mảng.
- Có backup file cũ.
- Có phương án rollback nếu lỗi nghiêm trọng.
- Không vừa dùng file cũ vừa dùng ERP mà không có kiểm soát.

### 11.2. Mô hình cutover đề xuất

Với Phase 1, nên dùng mô hình:

```text
Soft Freeze → Final Data Collection → Final Import → Validation → Go-live → Hypercare
```

### 11.3. Timeline cutover mẫu

| Ngày | Hoạt động | Owner |
|---|---|---|
| T-14 | Chốt danh sách dữ liệu cần migrate | PM/BA/Business Owner |
| T-10 | Hoàn tất cleansing vòng cuối | Data Owners |
| T-7 | Trial migration cuối | Tech/Data Team |
| T-6 | UAT trên dữ liệu migrated | Key Users |
| T-5 | Sửa lỗi migration | Tech/Data Team |
| T-3 | Soft freeze master data | PM/Business Owner |
| T-2 | Kiểm kê kho/công nợ chốt | Warehouse/Finance |
| T-1 | Final data export từ hệ cũ/file cũ | Data Owners |
| T-0 | Final import + reconciliation | Tech/Data/Business |
| T+1 | Go-live ERP | All Teams |
| T+1 đến T+14 | Hypercare | PM/Support/Key Users |

### 11.4. Soft Freeze là gì?

Soft Freeze nghĩa là hạn chế thay đổi dữ liệu gốc trước go-live.

Trong thời gian này:

- Không tạo mã hàng mới nếu không cấp bách.
- Không tự đổi tên SKU.
- Không đổi đơn vị tính.
- Không đổi BOM nếu chưa có approval.
- Không thay bảng giá nếu chưa ghi nhận vào cutover log.

### 11.5. Hard Freeze là gì?

Hard Freeze là ngừng giao dịch trên hệ cũ/file cũ để chốt dữ liệu cuối.

Thường áp dụng trong vài giờ đến một ngày tùy vận hành.

Trong hard freeze:

- Không nhập/xuất kho trên file cũ nếu không có cutover controller ghi nhận.
- Không tạo đơn mới ngoài hệ nếu đã go-live.
- Không sửa công nợ/chứng từ đầu kỳ.

---

## 12. Cutover Checklist

### 12.1. Trước cutover

- [ ] PRD/SRS Phase 1 đã sign-off.
- [ ] Permission Matrix đã sign-off.
- [ ] Data Dictionary đã sign-off.
- [ ] UAT critical cases đã pass.
- [ ] Danh sách dữ liệu migrate đã chốt.
- [ ] File master data đã làm sạch.
- [ ] Tồn kho đã kiểm kê/chốt.
- [ ] Batch/hạn dùng đã rà.
- [ ] Công nợ đã chốt với finance.
- [ ] User account đã tạo.
- [ ] Role/permission đã cấu hình.
- [ ] Template chứng từ đã sẵn sàng.
- [ ] Backup hệ cũ/file cũ đã lưu.
- [ ] Kế hoạch support go-live đã có.

### 12.2. Trong cutover

- [ ] Dừng nhập liệu hệ cũ/file cũ theo giờ đã công bố.
- [ ] Export file final.
- [ ] Kiểm tra version file final.
- [ ] Import master data.
- [ ] Import opening stock.
- [ ] Import opening AR/AP.
- [ ] Import open PO/SO/WO nếu có.
- [ ] Chạy reconciliation.
- [ ] Business owner xác nhận số.
- [ ] CEO/PM xác nhận go-live.

### 12.3. Sau cutover

- [ ] Kiểm tra user login.
- [ ] Kiểm tra quyền theo role.
- [ ] Tạo thử PR/PO.
- [ ] Tạo thử QC inspection.
- [ ] Tạo thử Work Order.
- [ ] Tạo thử stock transfer.
- [ ] Tạo thử Sales Order.
- [ ] Kiểm tra dashboard số tổng.
- [ ] Ghi nhận issue log.
- [ ] Họp daily hypercare.

---

## 13. Open Transactions – Chứng từ đang mở

Không nên migrate tất cả chứng từ lịch sử. Chỉ cần migrate chứng từ chưa kết thúc.

### 13.1. Open PO

Migrate PO nếu:

- PO đã gửi NCC nhưng chưa nhận hàng.
- PO nhận một phần.
- PO đã nhận hàng nhưng chưa đối chiếu hóa đơn nếu Phase 1 theo dõi.

Không migrate PO đã đóng hoàn toàn, chỉ lưu archive.

### 13.2. Open SO

Migrate SO nếu:

- Đơn đã tạo nhưng chưa giao.
- Đơn giao một phần.
- Đơn đã giao nhưng chưa thu tiền/COD chưa đối soát nếu Phase 1 theo dõi.

Không migrate đơn đã hoàn tất lâu trước go-live, chỉ lưu archive.

### 13.3. Open WO

Migrate WO nếu:

- Lệnh sản xuất đã mở nhưng chưa hoàn thành.
- Đã cấp NVL nhưng chưa nhập thành phẩm.
- Đang chờ QC release.

Nếu có thể, nên cố gắng đóng các WO cũ trước cutover để giảm rủi ro.

---

## 14. Rollback Plan – Phương án quay lại nếu lỗi nghiêm trọng

Rollback không phải mong muốn, nhưng phải có.

### 14.1. Khi nào cần rollback?

Có thể cân nhắc rollback nếu:

- Tồn kho sau import sai nghiêm trọng và không thể sửa trong ngày.
- User không thể tạo giao dịch chính.
- Permission sai làm lộ dữ liệu hoặc chặn vận hành toàn công ty.
- Sales/kho không thể vận hành đơn hàng.
- Finance không thể đối chiếu số đầu kỳ.

### 14.2. Điều kiện rollback

- Hệ cũ/file cũ vẫn được backup.
- Có danh sách giao dịch đã phát sinh trên ERP từ lúc go-live.
- Có người quyết định rollback: CEO/Business Owner + PM.

### 14.3. Cách rollback

1. Dừng tạo giao dịch mới trên ERP.
2. Export toàn bộ giao dịch đã phát sinh sau go-live.
3. Ghi nhận thủ công giao dịch cần tiếp tục vận hành trong file tạm.
4. Quay lại hệ cũ/file cũ tạm thời.
5. Phân tích nguyên nhân.
6. Sửa lỗi và lên lịch cutover lại.

### 14.4. Nguyên tắc

Rollback phải là quyết định quản trị, không phải phản ứng hoảng loạn.

Nếu lỗi nhỏ, nên xử lý trong hypercare thay vì rollback.

---

## 15. Hypercare Plan – Hỗ trợ sau go-live

Hypercare là giai đoạn 2 tuần đầu sau go-live, nơi issue sẽ xuất hiện nhiều nhất.

### 15.1. Thời gian đề xuất

- Tối thiểu: 2 tuần.
- Với ERP sản xuất/kho/batch: nên 3–4 tuần nếu vận hành phức tạp.

### 15.2. Cơ chế hỗ trợ

- Có nhóm support chung.
- Có daily issue review.
- Có owner cho từng issue.
- Có phân loại severity.
- Có cutoff thời gian xử lý.

### 15.3. Phân loại lỗi

| Severity | Định nghĩa | SLA đề xuất |
|---|---|---|
| S1 - Critical | Chặn vận hành chính, không có workaround | 2–4 giờ |
| S2 - High | Ảnh hưởng lớn nhưng có workaround | Trong ngày |
| S3 - Medium | Ảnh hưởng một nhóm nhỏ | 1–3 ngày |
| S4 - Low | UI/text/minor improvement | Backlog |

### 15.4. Issue Log mẫu

| ID | Ngày | Module | Mô tả lỗi | Severity | Owner | Trạng thái | Ghi chú |
|---|---|---|---|---|---|---|---|
| ISS-001 |  | Kho |  | S1/S2/S3/S4 |  | Open/In Progress/Resolved |  |

---

## 16. RACI cho Data Migration

| Công việc | CEO/Owner | PM | BA | Tech/Data | Kho | Finance | R&D | QA | Sales | Purchasing |
|---|---|---|---|---|---|---|---|---|---|---|
| Chốt scope migrate | A | R | C | C | C | C | C | C | C | C |
| Chuẩn hóa item master | I | C | R | C | C | I | C | C | C | C |
| Chuẩn hóa BOM | I | C | C | C | I | I | A/R | C | I | I |
| Chốt tồn kho | I | C | C | C | A/R | C | I | C | I | I |
| Chốt batch/hạn dùng | I | C | C | C | R | I | C | A/R | I | I |
| Chốt AR/AP | I | C | C | C | I | A/R | I | I | C | C |
| Import dữ liệu | I | C | C | A/R | I | I | I | I | I | I |
| Reconcile sau import | A | R | C | C | R | R | C | C | C | C |
| Final sign-off | A/R | C | C | I | C | C | C | C | C | C |

Ký hiệu:

- R = Responsible, người làm chính.
- A = Accountable, người chịu trách nhiệm cuối.
- C = Consulted, người cần tham vấn.
- I = Informed, người cần được thông báo.

---

## 17. Rủi ro chính và cách xử lý

| Rủi ro | Tác động | Cách xử lý |
|---|---|---|
| Mã hàng trùng/lệch | Báo cáo sai, giao dịch sai | Cleansing + sign-off item master |
| Tồn kho không khớp vật lý | Không giao được hàng, sai COGS | Kiểm kê trước cutover |
| Batch thiếu HSD | Không truy xuất được, rủi ro bán hàng | QA/Warehouse rà batch |
| Công nợ lệch | Sai tài chính, tranh chấp | Finance sign-off AR/AP |
| User vẫn dùng file cũ sau go-live | Dữ liệu phân mảnh | Cutover communication + lock file |
| Import sai nhưng không biết | Sai ngầm | Reconciliation bắt buộc |
| Chưa training user | Go-live rối | Training + UAT key user |
| Thiếu rollback plan | Rủi ro vận hành | Backup + rollback checklist |

---

## 18. Communication Plan – Truyền thông nội bộ

Trước go-live, phải thông báo rõ:

- Ngày nào khóa dữ liệu.
- Ngày nào bắt đầu dùng ERP.
- File nào không được dùng nữa.
- Ai là người hỗ trợ từng module.
- Issue báo ở đâu.
- Quy tắc nếu phát sinh giao dịch khẩn trong cutover.

### Mẫu thông báo nội bộ

```text
Từ ngày [GO-LIVE DATE], công ty bắt đầu vận hành ERP Phase 1 cho các nghiệp vụ: mua hàng, QC, sản xuất, kho và bán hàng.

Từ [HARD FREEZE TIME], các file Excel/phần mềm cũ liên quan đến tồn kho, đơn hàng, PO, lệnh sản xuất sẽ dừng cập nhật để chốt dữ liệu.

Mọi phát sinh trong thời gian cutover cần báo về [CUTOVER CONTROLLER] để ghi nhận.

Sau go-live, dữ liệu chính thức sẽ nằm trên ERP. File cũ chỉ dùng để tra cứu, không dùng làm nguồn sự thật.
```

---

## 19. Acceptance Criteria – Điều kiện nghiệm thu migration/cutover

Migration được xem là đạt khi:

- 100% master data bắt buộc đã import thành công.
- Không có mã hàng trùng trong hệ thống.
- Tồn kho tổng theo nhóm khớp file sign-off.
- Tồn kho theo batch/HSD khớp biên bản kiểm kê.
- Không có thành phẩm available mà thiếu batch/HSD.
- Công nợ AR/AP khớp báo cáo finance đã chốt.
- Open PO/SO/WO cần migrate đã có trong hệ thống.
- User key roles đăng nhập được.
- Quyền cơ bản hoạt động đúng.
- Dashboard số tổng không sai lệch nghiêm trọng.
- Business owner ký go-live.

---

## 20. Danh sách deliverables

| Deliverable | Owner | Trạng thái |
|---|---|---|
| Source Data Inventory | BA/PM | Pending |
| Data Cleansing Log | Data Owners | Pending |
| Item Master Final | Product/Kho/R&D | Pending |
| BOM Final | R&D/Production | Pending |
| Supplier Master Final | Purchasing | Pending |
| Customer Master Final | Sales | Pending |
| Opening Stock Final | Warehouse/QA | Pending |
| Opening AR/AP Final | Finance | Pending |
| Trial Migration Report | Tech/Data | Pending |
| Final Migration Report | Tech/Data | Pending |
| Reconciliation Report | PM/Finance/Kho | Pending |
| Go-live Sign-off | CEO/Business Owner | Pending |
| Hypercare Issue Log | PM/Support | Pending |

---

## 21. Kết luận

File này là tài liệu giúp công ty chuyển từ vận hành cũ sang ERP mới một cách có kiểm soát.

Điểm cốt lõi cần nhớ:

- Đừng import dữ liệu bẩn.
- Đừng migrate lịch sử quá nhiều.
- Đừng để dev tự đoán số liệu.
- Đừng go-live nếu tồn kho, batch, công nợ chưa được sign-off.
- Đừng dùng song song file cũ và ERP mà không kiểm soát.

Với doanh nghiệp mỹ phẩm, ba thứ sống còn khi cutover là:

```text
Mã hàng đúng
Batch/HSD đúng
Tồn kho đầu kỳ đúng
```

Nếu ba thứ này đúng, hệ thống có nền để chạy.  
Nếu ba thứ này sai, toàn bộ ERP sẽ tạo ra ảo giác quản trị.

---

## 22. Tài liệu tiếp theo đề xuất

Sau file này, tài liệu tiếp theo nên là:

```text
11_ERP_SOP_Training_Manual_Phase1_My_Pham_v1.md
```

Mục tiêu: hướng dẫn người dùng cuối thao tác từng nghiệp vụ trên ERP Phase 1.
