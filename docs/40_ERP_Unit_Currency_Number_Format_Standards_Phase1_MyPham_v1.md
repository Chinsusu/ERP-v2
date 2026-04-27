# 40_ERP_Unit_Currency_Number_Format_Standards_Phase1_MyPham_v1

**Dự án:** ERP Web Phase 1 - Công ty mỹ phẩm  
**Tài liệu:** Unit / Currency / Number Format Standards  
**Phiên bản:** v1.0  
**Trạng thái:** Approved Baseline  
**Ngôn ngữ tài liệu:** Vietnamese  
**Áp dụng cho:** Backend Go, Frontend Next.js, PostgreSQL, OpenAPI, Data Migration, QA/UAT, Report/Export  

---

## 1. Mục tiêu tài liệu

Tài liệu này chốt chuẩn dùng cho toàn hệ thống ERP về:

- Đơn vị tính hàng hóa, nguyên vật liệu, bao bì, thành phẩm.
- Chuẩn tiền tệ.
- Chuẩn số lượng, giá, tỷ lệ, phần trăm, hao hụt.
- Chuẩn làm tròn.
- Chuẩn hiển thị trên UI.
- Chuẩn lưu trong PostgreSQL.
- Chuẩn truyền qua API/OpenAPI.
- Chuẩn xử lý trong Go backend và Next.js frontend.
- Chuẩn export/import Excel/CSV.

Mục tiêu rất đơn giản: **mọi module phải hiểu số, tiền và đơn vị giống nhau**.

Nếu không khóa chuẩn này từ đầu, các module mua hàng, kho, QC, sản xuất/gia công, sales, shipping, returns, finance sẽ rất dễ lệch dữ liệu.

---

## 2. Nguyên tắc lõi

### 2.1. Không dùng float cho tiền và số lượng

Không được dùng:

```text
float
float32
float64
JavaScript number cho tính tiền/số lượng quan trọng
```

Lý do: float có sai số nhị phân. Trong ERP, sai số nhỏ có thể làm lệch báo cáo tồn kho, giá vốn, công nợ, đối soát COD, thanh toán nhà máy, thanh toán nhà cung cấp.

### 2.2. Backend là nguồn tính toán cuối cùng

Frontend có thể tính preview để người dùng dễ thao tác, nhưng số chính thức phải do backend tính.

Ví dụ:

```text
Frontend preview: tổng tiền đơn hàng
Backend authoritative: tổng tiền chính thức lưu vào DB
```

### 2.3. API truyền decimal bằng string

Tất cả tiền, số lượng, tỷ lệ quan trọng phải truyền qua API bằng string decimal.

Đúng:

```json
{
  "qty": "10.500000",
  "unit_price": "125000.0000",
  "total_amount": "1312500.00"
}
```

Sai:

```json
{
  "qty": 10.5,
  "unit_price": 125000,
  "total_amount": 1312500
}
```

### 2.4. Database dùng PostgreSQL `numeric`

PostgreSQL phải dùng `numeric`, không dùng `float`, `real`, `double precision` cho các field tiền/số lượng.

### 2.5. Mỗi item phải có một đơn vị cơ sở

Mỗi nguyên liệu, bao bì, thành phẩm, bán thành phẩm phải có `base_uom_code`.

Tồn kho luôn quy về đơn vị cơ sở để tính toán.

Ví dụ:

```text
Vitamin C Powder:
- base_uom = G
- purchase_uom = KG
- issue_uom = G

Serum 30ml:
- base_uom = PCS
- sales_uom = PCS
- carton_uom = CARTON, quy đổi theo item
```

---

## 3. Chuẩn tiền tệ

### 3.1. Base currency Phase 1

Phase 1 dùng tiền tệ gốc:

```text
Currency: VND
Currency code: VND
Locale: vi-VN
Timezone: Asia/Ho_Chi_Minh
```

### 3.2. Quy tắc lưu tiền trong database

Các khoản tiền tổng, thanh toán, công nợ, phí, chiết khấu, thuế:

```sql
amount numeric(18,2) not null
currency_code varchar(3) not null default 'VND'
```

Các trường tiền điển hình:

```text
subtotal_amount
discount_amount
tax_amount
shipping_fee
cod_amount
payment_amount
total_amount
paid_amount
outstanding_amount
```

### 3.3. Quy tắc unit price / unit cost

`unit_price` và `unit_cost` nên dùng nhiều chữ số hơn tổng tiền vì có thể tính theo gram/ml hoặc phân bổ chi phí.

```sql
unit_price numeric(18,4) not null
unit_cost numeric(18,6) null
```

Gợi ý:

```text
unit_price: giá bán/giá mua theo đơn vị giao dịch
unit_cost: giá vốn/giá thành đơn vị, có thể cần độ chính xác cao hơn
```

Ví dụ:

```text
Giá nguyên liệu mua theo KG: 2.350.000 ₫/KG
Đổi ra G: 2.350 ₫/G

Một số phân bổ chi phí có thể ra:
2.350,125678 ₫/G
```

### 3.4. Quy tắc hiển thị tiền trên UI

UI hiển thị theo chuẩn Việt Nam:

```text
1.250.000 ₫
```

Không hiển thị số lẻ cho VND ở màn hình người dùng cuối, trừ báo cáo kỹ thuật hoặc màn hình phân tích giá vốn yêu cầu chi tiết.

Ví dụ:

```text
Đơn hàng: 1.250.000 ₫
Công nợ: 38.500.000 ₫
Phí ship: 30.000 ₫
```

### 3.5. Quy tắc export tiền

Có 2 loại export.

#### Human export

Dành cho quản lý xem Excel:

```text
1.250.000 ₫
```

Hoặc cột tiền dạng numeric có format Excel:

```text
#,##0 [$₫-vi-VN]
```

#### Machine export

Dành cho import/integration:

```csv
currency_code,total_amount
VND,1250000.00
```

Machine export không dùng dấu phân cách hàng nghìn.

---

## 4. Chuẩn số lượng

### 4.1. Quantity precision chuẩn

Toàn hệ thống chốt số lượng dùng:

```sql
quantity numeric(18,6)
```

Các field áp dụng:

```text
qty
ordered_qty
received_qty
accepted_qty
rejected_qty
issued_qty
reserved_qty
picked_qty
packed_qty
returned_qty
base_qty
available_qty
physical_qty
adjustment_qty
scrap_qty
loss_qty
```

### 4.2. Vì sao chọn `numeric(18,6)`

Ngành mỹ phẩm có:

- Nguyên liệu định lượng nhỏ.
- Công thức/BOM.
- Hao hụt sản xuất/gia công.
- Quy đổi KG/G, L/ML.
- Phân bổ giá vốn.
- Kiểm kê có sai lệch nhỏ.

`numeric(18,6)` đủ an toàn cho Phase 1 và Phase 2.

### 4.3. Số lượng hiển thị theo loại đơn vị

UI không hiển thị dài 6 số lẻ nếu không cần.

| Nhóm | Ví dụ | DB | UI hiển thị |
|---|---:|---:|---:|
| Thành phẩm | 24 PCS | `24.000000` | `24 chai` hoặc `24 PCS` |
| Bao bì | 500 PCS | `500.000000` | `500 PCS` |
| Nguyên liệu KG | 10.500000 KG | `10.500000` | `10,5 kg` |
| Nguyên liệu G | 1250.000000 G | `1250.000000` | `1.250 g` |
| Dung tích ML | 3000.000000 ML | `3000.000000` | `3.000 ml` |
| Hao hụt | 2.345678 KG | `2.345678` | `2,345678 kg` khi cần phân tích |

### 4.4. UI input số lượng

UI nên hỗ trợ người dùng nhập kiểu Việt Nam:

```text
10,5
1.250,75
```

Sau đó normalize về decimal string chuẩn API:

```text
10.5
1250.75
```

API và DB chỉ dùng dấu chấm `.` làm decimal separator.

---

## 5. Chuẩn tỷ lệ, phần trăm, hao hụt

### 5.1. Database precision

Tỷ lệ, phần trăm, hao hụt, discount rate, tax rate:

```sql
rate numeric(9,4)
```

### 5.2. Cách hiểu phần trăm

Lưu phần trăm theo dạng số phần trăm, không lưu dạng fraction.

Đúng:

```text
10%  -> 10.0000
2.5% -> 2.5000
```

Sai:

```text
10% -> 0.10
```

### 5.3. Ví dụ

```json
{
  "discount_rate": "10.0000",
  "tax_rate": "8.0000",
  "loss_rate": "2.5000"
}
```

UI hiển thị:

```text
10%
8%
2,5%
```

---

## 6. Chuẩn đơn vị tính UOM

### 6.1. UOM master

Hệ thống phải có bảng `uom`.

Field đề xuất:

```text
uom_code
name_vi
name_en
category
decimal_scale
allow_decimal
is_global_convertible
is_active
description
```

### 6.2. Danh mục UOM Phase 1

| UOM | Tên | Nhóm | Cho phép decimal | Ghi chú |
|---|---|---|---:|---|
| MG | Milligram | MASS | Có | Dùng nếu R&D/BOM cần định lượng rất nhỏ |
| G | Gram | MASS | Có | Base UOM phổ biến cho nguyên liệu rắn/bột |
| KG | Kilogram | MASS | Có | UOM mua hàng phổ biến |
| ML | Milliliter | VOLUME | Có | Base UOM phổ biến cho nguyên liệu lỏng |
| L | Liter | VOLUME | Có | UOM mua hàng phổ biến |
| PCS | Piece | COUNT | Không | Đơn vị đếm chuẩn |
| BOTTLE | Chai | PACK | Không | Quy đổi theo item nếu cần |
| JAR | Hũ | PACK | Không | Quy đổi theo item nếu cần |
| TUBE | Tuýp | PACK | Không | Quy đổi theo item nếu cần |
| BOX | Hộp | PACK | Không | Quy đổi theo item |
| CARTON | Thùng | PACK | Không | Quy đổi theo item |
| SET | Bộ/combo | PACK | Không | Quy đổi theo BOM/combo |
| SERVICE | Dịch vụ | SERVICE | Không | Không quản tồn kho |

### 6.3. Quy tắc UOM theo item type

| Loại item | Base UOM khuyến nghị | Ghi chú |
|---|---|---|
| Nguyên liệu rắn/bột | G | Mua bằng KG, cấp phát bằng G |
| Nguyên liệu lỏng | ML | Mua bằng L, cấp phát bằng ML |
| Bao bì chai/lọ/nắp/hộp/tem | PCS | Quản đếm từng đơn vị |
| Bán thành phẩm | G hoặc ML hoặc PCS | Tùy bản chất sản phẩm |
| Thành phẩm chai/lọ/tuýp | PCS | Bán theo PCS/BOX/CARTON |
| Combo/set | SET | Thành phần bên trong theo BOM/combo |
| Dịch vụ gia công | SERVICE | Không quản tồn kho trực tiếp |

### 6.4. Quy đổi toàn cục

Chỉ dùng cho các đơn vị đo lường chuẩn, không phụ thuộc item.

```text
1 KG = 1000 G
1 G = 1000 MG
1 L = 1000 ML
```

### 6.5. Quy đổi theo item

Dùng cho đơn vị bao gói/thương mại.

Ví dụ:

```text
1 CARTON Serum A = 48 PCS
1 BOX Mask B = 10 PCS
1 SET Combo C = 1 Serum + 1 Toner + 1 Cleanser
```

Không được tạo quy đổi toàn cục kiểu:

```text
1 CARTON = 48 PCS
```

Vì mỗi sản phẩm có thể có số lượng/thùng khác nhau.

---

## 7. Chuẩn UOM conversion

### 7.1. Bảng conversion

Bảng đề xuất:

```text
uom_conversions
- id
- item_id nullable
- from_uom_code
- to_uom_code
- factor
- conversion_type: GLOBAL | ITEM_SPECIFIC
- effective_from
- effective_to
- is_active
- created_by
- approved_by
- approved_at
```

### 7.2. Precision conversion factor

```sql
factor numeric(18,6)
```

### 7.3. Công thức conversion

```text
to_qty = from_qty * factor
```

Ví dụ:

```text
from_qty = 2 CARTON
factor = 48 PCS/CARTON
to_qty = 96 PCS
```

### 7.4. Backend phải trả cả UOM giao dịch và base UOM

Ví dụ API nhận hàng:

```json
{
  "sku_code": "SERUM-VITC-30ML",
  "qty": "2.000000",
  "uom_code": "CARTON",
  "base_qty": "96.000000",
  "base_uom_code": "PCS",
  "conversion_factor": "48.000000"
}
```

### 7.5. Không cho frontend tự quy đổi tồn kho

Frontend chỉ hiển thị. Backend phải là nơi tính `base_qty`, `available_qty`, `reserved_qty`, `issued_qty`.

---

## 8. Chuẩn batch, hạn dùng và đơn vị

### 8.1. Batch luôn gắn với base UOM

Stock ledger phải ghi số lượng base.

Ví dụ:

```text
Nhập 2 CARTON Serum A
Conversion: 1 CARTON = 48 PCS
Stock ledger ghi: +96 PCS
```

### 8.2. Hạn dùng không thay đổi số lượng

Expiry date là thuộc tính của batch/lot, không phải thuộc tính của UOM.

### 8.3. FEFO/FIFO

FEFO/FIFO sử dụng batch + expiry date + received date. Không dùng UOM giao dịch để quyết định thứ tự xuất.

---

## 9. Chuẩn stock quantity

### 9.1. Stock ledger bất biến

Mọi thay đổi tồn kho phải đi qua stock movement.

Ví dụ movement type:

```text
INBOUND_RECEIPT
QC_RELEASE
PURCHASE_RETURN
PRODUCTION_ISSUE
SUBCONTRACT_MATERIAL_ISSUE
SUBCONTRACT_FINISHED_GOODS_RECEIPT
SALES_RESERVE
SALES_PICK
SALES_ISSUE
SHIPPING_HANDOVER
RETURN_RECEIPT
RETURN_DISPOSITION_REUSABLE
RETURN_DISPOSITION_NON_REUSABLE
ADJUSTMENT
SCRAP
```

### 9.2. Stock ledger quantity

```sql
movement_qty numeric(18,6) not null
base_uom_code varchar(20) not null
```

### 9.3. Stock balance fields

```text
physical_qty
reserved_qty
qc_hold_qty
damaged_qty
return_pending_qty
available_qty
```

### 9.4. Available stock formula

Chuẩn Phase 1:

```text
available_qty
= physical_qty
- reserved_qty
- qc_hold_qty
- damaged_qty
- return_pending_qty
```

Trong đó:

```text
physical_qty: tồn vật lý tổng
reserved_qty: hàng đã giữ cho đơn/lệnh
qc_hold_qty: hàng đang hold QC, chưa được phép bán/cấp phát
damaged_qty: hàng hỏng/không sử dụng
return_pending_qty: hàng hoàn chưa kiểm xong
```

---

## 10. Chuẩn mua hàng và nhập kho

### 10.1. Purchase UOM

Mua hàng có thể dùng UOM khác base UOM.

Ví dụ:

```text
Nguyên liệu mua 25 KG
Base UOM là G
Hệ thống lưu:
- ordered_qty = 25.000000 KG
- base_ordered_qty = 25000.000000 G
```

### 10.2. Receiving UOM

Khi nhập kho, người dùng có thể nhập theo UOM chứng từ giao hàng.

Backend phải quy đổi về base UOM để ghi stock.

### 10.3. QC đầu vào

QC kiểm theo số lượng nhận và base quantity.

Ví dụ:

```text
received_qty = 25 KG
accepted_qty = 24.5 KG
rejected_qty = 0.5 KG
base_accepted_qty = 24500 G
base_rejected_qty = 500 G
```

---

## 11. Chuẩn sản xuất/gia công ngoài

### 11.1. Gửi nguyên vật liệu/bao bì cho nhà máy

Khi chuyển NVL/bao bì sang nhà máy gia công, stock movement phải ghi bằng base UOM.

Ví dụ:

```text
Chuyển 10 KG active sang nhà máy
Base UOM = G
Stock movement: -10000 G
```

### 11.2. Nhận thành phẩm từ nhà máy

Thành phẩm nhận về theo PCS/BOX/CARTON, nhưng tồn thành phẩm phải quy về base UOM của thành phẩm.

Ví dụ:

```text
Nhận 100 CARTON Serum A
1 CARTON = 48 PCS
Stock ledger: +4800 PCS
```

### 11.3. Hao hụt gia công

Hao hụt nguyên liệu:

```text
loss_qty numeric(18,6)
loss_rate numeric(9,4)
```

Công thức:

```text
loss_rate = loss_qty / issued_qty * 100
```

Lưu ý: backend tính, không để frontend tự chốt.

---

## 12. Chuẩn sales, shipping, returns

### 12.1. Sales UOM

Đơn bán thường dùng:

```text
PCS
BOX
CARTON
SET
```

Nếu bán theo BOX/CARTON, backend phải quy đổi về base UOM để reserve/pick/issue.

### 12.2. Shipping handover

Bàn giao ĐVVC kiểm theo số đơn/gói, không thay đổi base inventory nếu hàng đã issue trước đó.

Các field quantity liên quan:

```text
manifest_order_count integer
scanned_order_count integer
missing_order_count integer
package_count integer
```

`order_count` và `package_count` dùng integer, không dùng numeric decimal.

### 12.3. Return receiving

Hàng hoàn có thể nhận theo:

```text
order_code
tracking_code
sku_code
batch_no
qty
uom_code
condition_status
```

Nếu hàng hoàn còn dùng được, backend tạo stock movement vào kho phù hợp.

Nếu không dùng được, chuyển vào damaged/lab/scrap area, không cộng vào available stock.

---

## 13. Chuẩn làm tròn

### 13.1. Money rounding

| Loại | Scale lưu DB | Scale hiển thị UI |
|---|---:|---:|
| Amount tổng | 2 | 0 cho VND |
| Unit price | 4 | 0-4 tùy màn hình |
| Unit cost | 6 | 2-6 tùy báo cáo |
| Payment amount | 2 | 0 cho VND |
| Tax amount | 2 | 0 cho VND |
| Discount amount | 2 | 0 cho VND |

### 13.2. Quantity rounding

| Loại | Scale DB | UI |
|---|---:|---|
| Nguyên liệu | 6 | 3-6 tùy nghiệp vụ |
| Bao bì | 6 | 0 nếu PCS |
| Thành phẩm | 6 | 0 nếu PCS |
| Conversion factor | 6 | 0-6 |
| Hao hụt | 6 | 3-6 |

### 13.3. Rounding mode

Phase 1 dùng:

```text
ROUND_HALF_UP
```

Cho các phép tính tiền hiển thị và line amount.

Lưu ý: Nếu sau này kế toán/thuế yêu cầu rounding mode khác, phải tạo change request riêng.

### 13.4. Không làm tròn giữa chừng nếu không cần

Quy tắc:

```text
Tính bằng precision cao trước
Làm tròn ở điểm ghi nhận line amount / total amount / payment amount
```

Không làm tròn từng bước nhỏ nếu sẽ cộng dồn.

---

## 14. Chuẩn thuế và giá đã gồm/chưa gồm thuế

Phase 1 có thể chưa làm accounting đầy đủ, nhưng schema/API phải có chỗ cho thuế.

### 14.1. Field chuẩn

```text
tax_inclusive boolean
tax_rate numeric(9,4)
tax_amount numeric(18,2)
```

### 14.2. Giá chưa gồm thuế

```text
subtotal = qty * unit_price
tax_amount = subtotal * tax_rate / 100
total = subtotal + tax_amount - discount_amount
```

### 14.3. Giá đã gồm thuế

```text
gross_amount = qty * unit_price
net_amount = gross_amount / (1 + tax_rate / 100)
tax_amount = gross_amount - net_amount
total = gross_amount - discount_amount
```

### 14.4. Quyết định nghiệp vụ cần chốt

Mỗi bảng giá phải ghi rõ:

```text
tax_inclusive = true/false
```

Không được để người dùng tự hiểu.

---

## 15. Chuẩn ngày giờ và timezone

### 15.1. Timezone business

```text
Asia/Ho_Chi_Minh
```

### 15.2. Database

DB lưu timestamp bằng:

```sql
timestamptz
```

### 15.3. API

API dùng ISO 8601/RFC3339.

Ví dụ:

```json
{
  "created_at": "2026-04-26T08:30:00+07:00"
}
```

Date-only dùng:

```text
YYYY-MM-DD
```

Ví dụ:

```json
{
  "expiry_date": "2027-12-31"
}
```

### 15.4. UI

UI hiển thị:

```text
Ngày: dd/MM/yyyy
Giờ: HH:mm
Ngày giờ: dd/MM/yyyy HH:mm
```

Ví dụ:

```text
26/04/2026 08:30
```

---

## 16. Chuẩn API / OpenAPI

### 16.1. Money schema

```yaml
MoneyAmount:
  type: string
  pattern: '^-?\\d+(\\.\\d{1,6})?$'
  example: "125000.00"
```

### 16.2. Quantity schema

```yaml
Quantity:
  type: string
  pattern: '^-?\\d+(\\.\\d{1,6})?$'
  example: "10.500000"
```

### 16.3. Rate schema

```yaml
Rate:
  type: string
  pattern: '^-?\\d+(\\.\\d{1,4})?$'
  example: "2.5000"
```

### 16.4. UOM field

```yaml
uom_code:
  type: string
  example: "PCS"
```

### 16.5. Currency field

```yaml
currency_code:
  type: string
  minLength: 3
  maxLength: 3
  example: "VND"
```

---

## 17. Chuẩn PostgreSQL schema

### 17.1. Data type mapping

| Loại dữ liệu | PostgreSQL type |
|---|---|
| Money amount | `numeric(18,2)` |
| Unit price | `numeric(18,4)` |
| Unit cost | `numeric(18,6)` |
| Quantity | `numeric(18,6)` |
| Conversion factor | `numeric(18,6)` |
| Rate/percent | `numeric(9,4)` |
| Currency code | `varchar(3)` |
| UOM code | `varchar(20)` |
| Count | `integer` hoặc `bigint` |
| Date-only | `date` |
| Timestamp | `timestamptz` |

### 17.2. Không dùng các type này cho tiền/số lượng

```sql
real
double precision
float
money
```

PostgreSQL có type `money`, nhưng không dùng vì khó chuẩn hóa multi-currency, export/API và format.

### 17.3. Example table: UOM

```sql
create table uoms (
  uom_code varchar(20) primary key,
  name_vi text not null,
  name_en text,
  category varchar(30) not null,
  decimal_scale int not null default 0,
  allow_decimal boolean not null default false,
  is_global_convertible boolean not null default false,
  is_active boolean not null default true,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);
```

### 17.4. Example table: UOM conversion

```sql
create table uom_conversions (
  id uuid primary key,
  item_id uuid null,
  from_uom_code varchar(20) not null references uoms(uom_code),
  to_uom_code varchar(20) not null references uoms(uom_code),
  factor numeric(18,6) not null,
  conversion_type varchar(30) not null,
  effective_from date not null,
  effective_to date null,
  is_active boolean not null default true,
  created_by uuid not null,
  approved_by uuid null,
  approved_at timestamptz null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint chk_uom_conversion_factor_positive check (factor > 0)
);
```

---

## 18. Chuẩn Go backend

### 18.1. Không dùng float cho domain values

Sai:

```go
type SalesLine struct {
    Qty float64
    UnitPrice float64
}
```

Đúng:

```go
type SalesLine struct {
    Qty       Decimal
    UnitPrice Decimal
}
```

`Decimal` có thể là wrapper nội bộ quanh thư viện decimal đã chọn.

### 18.2. Domain value objects đề xuất

```text
Money
Quantity
Rate
CurrencyCode
UOMCode
```

Ví dụ:

```go
type Money struct {
    Amount   Decimal
    Currency string
}

type Quantity struct {
    Value Decimal
    UOM   string
}
```

### 18.3. Central rounding service

Phải có utility/service dùng chung:

```text
RoundMoney(amount, currency)
RoundUnitPrice(amount)
RoundUnitCost(amount)
RoundQuantity(qty, uom)
RoundRate(rate)
```

Không cho mỗi module tự viết rounding riêng.

### 18.4. Backend phải validate UOM conversion

Khi nhận request có `qty` và `uom_code`, backend phải:

```text
1. Lấy item base_uom.
2. Tìm conversion hợp lệ.
3. Tính base_qty.
4. Ghi transaction bằng base_qty.
5. Trả response có cả qty/uom và base_qty/base_uom.
```

---

## 19. Chuẩn frontend Next.js

### 19.1. Frontend không tính chính thức

Frontend chỉ:

```text
- format hiển thị
- parse input
- preview tạm thời
- gửi decimal string lên API
```

Backend mới tính chính thức.

### 19.2. Component cần có

```text
MoneyText
MoneyInput
QuantityText
QuantityInput
RateInput
UOMSelect
CurrencyBadge
DateText
DateTimeText
```

### 19.3. Input normalize

Người dùng nhập:

```text
1.250,5
```

Frontend normalize thành:

```text
1250.5
```

Gửi API:

```json
{
  "qty": "1250.5"
}
```

### 19.4. Không dùng JavaScript number cho tính toán chính thức

Không làm:

```ts
const total = qty * unitPrice;
```

Cho dữ liệu chính thức.

Nếu cần preview, phải dùng decimal library hoặc gọi API calculate/preview.

---

## 20. Chuẩn import/export Excel/CSV

### 20.1. Import

File import phải có cột rõ:

```text
qty
uom_code
unit_price
currency_code
```

Ví dụ:

```csv
sku_code,qty,uom_code,unit_price,currency_code
SERUM-VITC-30ML,48,PCS,125000.00,VND
```

### 20.2. Import parser

Import phải chấp nhận 2 dạng nếu có thể:

```text
1250.5
1.250,5
```

Nhưng sau khi parse phải lưu thành decimal chuẩn.

### 20.3. Export machine-readable

```csv
sku_code,qty,uom_code,base_qty,base_uom_code,total_amount,currency_code
SERUM-VITC-30ML,2,CARTON,96,PCS,12000000.00,VND
```

### 20.4. Export human-readable

Có thể format đẹp:

```text
2 thùng
96 chai
12.000.000 ₫
```

Nhưng phải có option export machine-readable cho migration/integration.

---

## 21. Chuẩn validation

### 21.1. Quantity validation

```text
qty > 0 cho giao dịch tạo mới
adjustment_qty có thể âm/dương tùy loại adjustment
received_qty <= ordered_qty nếu rule PO không cho vượt
picked_qty <= reserved_qty
issued_qty <= available_qty
returned_qty <= original_sold_qty nếu return theo order
```

### 21.2. Money validation

```text
amount >= 0 trong hầu hết transaction
unit_price >= 0
unit_cost >= 0
discount_amount <= subtotal_amount
payment_amount <= outstanding_amount, trừ overpayment được cấu hình riêng
```

### 21.3. Rate validation

```text
0 <= discount_rate <= 100
0 <= tax_rate <= 100
0 <= loss_rate <= 100, trừ exception cần approval
```

### 21.4. UOM validation

```text
item phải có base_uom_code
uom_code phải active
conversion phải tồn tại nếu uom_code khác base_uom_code
PACK UOM như BOX/CARTON phải là item-specific conversion
```

---

## 22. Chuẩn lỗi API liên quan đơn vị/số/tiền

### 22.1. Không đủ tồn

```json
{
  "success": false,
  "code": "INSUFFICIENT_STOCK",
  "message": "Tồn khả dụng không đủ.",
  "details": {
    "sku_code": "SERUM-VITC-30ML",
    "requested_qty": "10.000000",
    "requested_uom_code": "PCS",
    "available_qty": "6.000000",
    "base_uom_code": "PCS"
  }
}
```

### 22.2. Thiếu quy đổi UOM

```json
{
  "success": false,
  "code": "UOM_CONVERSION_NOT_FOUND",
  "message": "Chưa có quy đổi đơn vị cho sản phẩm này.",
  "details": {
    "sku_code": "SERUM-VITC-30ML",
    "from_uom_code": "CARTON",
    "to_uom_code": "PCS"
  }
}
```

### 22.3. Sai format tiền

```json
{
  "success": false,
  "code": "INVALID_MONEY_FORMAT",
  "message": "Số tiền phải là chuỗi decimal hợp lệ, ví dụ 125000.00."
}
```

---

## 23. Chuẩn UI microcopy

### 23.1. Batch đang hold

```text
Batch này đang HOLD, chưa được phép xuất bán hoặc cấp phát.
```

### 23.2. Thiếu tồn khả dụng

```text
Tồn khả dụng không đủ. Vui lòng kiểm tra tồn giữ, QC hold hoặc hàng hoàn chưa kiểm.
```

### 23.3. Thiếu quy đổi đơn vị

```text
Sản phẩm này chưa có quy đổi đơn vị từ thùng/hộp sang đơn vị cơ sở. Vui lòng cập nhật UOM Conversion.
```

### 23.4. Nhập sai số lượng

```text
Số lượng không hợp lệ. Vui lòng nhập số lớn hơn 0.
```

### 23.5. Tiền không hợp lệ

```text
Số tiền không hợp lệ. Vui lòng nhập đúng định dạng, ví dụ 125000 hoặc 125000,50.
```

---

## 24. Rule theo module Phase 1

### 24.1. Master Data

Bắt buộc có:

```text
base_uom_code
item_type
allow_decimal_qty
```

Nếu item bán theo thùng/hộp, phải có item-specific conversion.

### 24.2. Purchase

PO line phải có:

```text
ordered_qty
uom_code
base_ordered_qty
base_uom_code
unit_price
currency_code
```

### 24.3. QC

QC line phải có:

```text
received_qty
accepted_qty
rejected_qty
uom_code
base_accepted_qty
base_rejected_qty
```

### 24.4. Inventory

Stock ledger luôn dùng:

```text
movement_qty
base_uom_code
```

Nếu cần hiển thị giao dịch gốc, lưu thêm:

```text
source_qty
source_uom_code
conversion_factor
```

### 24.5. Sales

Sales line phải có:

```text
ordered_qty
uom_code
base_ordered_qty
base_uom_code
unit_price
currency_code
line_amount
```

### 24.6. Shipping

Shipping dùng count integer cho số đơn/số kiện:

```text
package_count integer
order_count integer
scanned_count integer
missing_count integer
```

### 24.7. Returns

Return line phải có:

```text
returned_qty
uom_code
base_returned_qty
base_uom_code
condition_status
disposition_status
```

### 24.8. Subcontract Manufacturing

Subcontract material issue:

```text
issued_qty
uom_code
base_issued_qty
base_uom_code
```

Subcontract finished goods receipt:

```text
received_qty
uom_code
base_received_qty
base_uom_code
```

---

## 25. Data migration rule

### 25.1. Opening balance

Opening stock phải import theo base UOM.

Nếu file cũ có UOM khác base UOM, migration script phải convert.

Ví dụ:

```text
Old file: 100 CARTON Serum A
Conversion: 1 CARTON = 48 PCS
Opening stock ledger: 4800 PCS
```

### 25.2. Migration validation

Trước khi import:

```text
- item có base_uom chưa?
- uom_code có active không?
- conversion có tồn tại không?
- qty parse được không?
- money parse được không?
- currency_code có phải VND không?
```

### 25.3. Không import dữ liệu mơ hồ

Nếu file cũ ghi:

```text
10 hộp
```

nhưng không biết `1 hộp = bao nhiêu PCS`, không được tự đoán. Phải đưa vào danh sách cần business xác nhận.

---

## 26. Testing checklist

### 26.1. Unit test

Phải test:

```text
- parse decimal string
- format VND
- parse input kiểu vi-VN
- UOM conversion global
- UOM conversion item-specific
- rounding money
- rounding quantity
- discount calculation
- tax inclusive/exclusive
```

### 26.2. Integration test

Phải test:

```text
- PO nhập bằng KG, tồn ghi bằng G
- Sales bán bằng CARTON, reserve bằng PCS
- Return nhận bằng PCS, phân loại reusable/non-reusable
- Subcontract issue nguyên liệu KG, nhận thành phẩm CARTON
- Stock available formula
```

### 26.3. UAT test

Business phải test:

```text
- người kho nhập hàng theo chứng từ thật
- người kho bàn giao ĐVVC theo số đơn/số kiện
- người kho nhận hàng hoàn và phân loại
- người phụ trách gia công gửi NVL/bao bì cho nhà máy
- kế toán/finance xem tiền VND đúng format
```

---

## 27. Definition of Done

Một task liên quan tiền/số lượng/đơn vị chỉ được coi là xong khi:

```text
1. Không dùng float/double cho money/qty/rate.
2. Database dùng numeric đúng scale.
3. API dùng string decimal.
4. UI hiển thị đúng vi-VN.
5. Có UOM/base UOM rõ ràng.
6. Có conversion nếu giao dịch khác base UOM.
7. Backend tính base_qty.
8. Backend là nguồn tính tổng tiền chính thức.
9. Có validation lỗi rõ ràng.
10. Có test case cho rounding/conversion.
11. Export/import không làm mất precision.
12. Audit log ghi thay đổi master UOM/conversion/price quan trọng.
```

---

## 28. Các tài liệu cần cập nhật theo chuẩn này

Tài liệu này là source of truth cho unit/currency/number format. Các file sau cần align:

```text
05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md
12_ERP_Go_Coding_Standards_Phase1_MyPham_v1.md
16_ERP_API_Contract_OpenAPI_Standards_Phase1_MyPham_v1.md
17_ERP_Database_Schema_PostgreSQL_Standards_Phase1_MyPham_v1.md
24_ERP_QA_Test_Strategy_Automation_Phase1_MyPham_v1.md
25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md
33_ERP_Core_Docs_v1_1_Update_Pack_Phase1_MyPham.md
37_ERP_Coding_Task_Board_Phase1_MyPham_v1.md
```

Điểm cần sửa ngay:

```text
Quantity DB precision phải thống nhất thành numeric(18,6).
Money amount dùng numeric(18,2).
Unit price dùng numeric(18,4).
Unit cost dùng numeric(18,6).
Rate/percent dùng numeric(9,4).
API money/qty/rate dùng string decimal.
```

---

## 29. Quyết định đã chốt

```text
Base currency: VND
Locale: vi-VN
Timezone: Asia/Ho_Chi_Minh
Money amount DB: numeric(18,2)
Unit price DB: numeric(18,4)
Unit cost DB: numeric(18,6)
Quantity DB: numeric(18,6)
Rate/percent DB: numeric(9,4)
Currency code: varchar(3), default VND
UOM code: varchar(20)
API decimal: string
UI money display: 1.250.000 ₫
UI date display: dd/MM/yyyy
UI datetime display: dd/MM/yyyy HH:mm
Stock ledger: always base UOM
Frontend: no authoritative calculation for money/stock
Backend: authoritative calculation and conversion
```

---

## 30. Một câu nhớ cho team

**Tồn kho không tính bằng cảm giác. Tiền không tính bằng float. Đơn vị không để mỗi người hiểu một kiểu.**

Chuẩn này là “ngôn ngữ số học” của toàn bộ ERP.
