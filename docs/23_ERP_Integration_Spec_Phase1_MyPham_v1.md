# 23_ERP_Integration_Spec_Phase1_MyPham_v1

## 1. Thông tin tài liệu

| Trường | Nội dung |
|---|---|
| Tên tài liệu | ERP Integration Specification - Phase 1 |
| Phiên bản | v1.0 |
| Dự án | Web ERP cho công ty mỹ phẩm |
| Phạm vi | Phase 1: kho, mua hàng, QC, sản xuất/gia công ngoài, bán hàng, giao hàng, hàng hoàn, tài chính cơ bản |
| Backend | Go - Modular Monolith |
| Frontend | React / Next.js + TypeScript |
| Database | PostgreSQL |
| API | REST + OpenAPI |
| Ngôn ngữ tài liệu | Tiếng Việt |
| Mục tiêu | Chốt chuẩn tích hợp giữa ERP và các hệ thống/đối tác bên ngoài |

---

## 2. Mục đích tài liệu

Tài liệu này dùng để khóa cách ERP Phase 1 giao tiếp với các hệ thống bên ngoài và các thiết bị/quy trình không nằm hoàn toàn trong ERP.

Nói dễ hiểu: ERP không thể sống một mình. Nó phải nhận đơn từ kênh bán, đẩy đơn cho đơn vị vận chuyển, nhận trạng thái giao hàng, đối soát COD, nhận dữ liệu hoàn hàng, xuất dữ liệu kế toán, lưu file chứng từ, gửi thông báo và hỗ trợ thao tác quét mã tại kho.

Tài liệu này trả lời các câu hỏi:

- ERP tích hợp với những hệ thống nào?
- Dữ liệu nào đi vào ERP?
- Dữ liệu nào đi ra khỏi ERP?
- Hệ thống nào là nguồn sự thật?
- Tích hợp realtime, webhook, batch job hay import/export?
- Khi lỗi sync thì xử lý thế nào?
- Có cần idempotency, retry, audit log, reconciliation không?
- Trạng thái đơn hàng/giao hàng/hàng hoàn map với nhau ra sao?

Mục tiêu cuối cùng: tránh cảnh mỗi module tự tích hợp một kiểu, dữ liệu không khớp, đơn hàng lệch, COD lệch, tồn kho lệch và không biết lỗi nằm ở đâu.

---

## 3. Nguồn đầu vào nghiệp vụ

Tài liệu này dựa trên bộ thiết kế ERP đã có và các workflow thực tế do công ty cung cấp:

- Công việc hằng ngày của kho: tiếp nhận đơn trong ngày, thực hiện xuất/nhập, soạn và đóng hàng, sắp xếp kho, kiểm kê cuối ngày, đối soát số liệu và kết thúc ca.
- Nội quy kho: nhập kho, xuất kho, đóng hàng, xử lý hàng hoàn.
- Quy trình bàn giao cho đơn vị vận chuyển: phân chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, quét mã trực tiếp tại hàm, xử lý đủ/chưa đủ đơn, ký xác nhận bàn giao.
- Quy trình sản xuất/gia công ngoài: lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển nguyên vật liệu/bao bì, làm mẫu/chốt mẫu, sản xuất hàng loạt, giao hàng về kho, kiểm tra số lượng/chất lượng, nhận hàng hoặc báo lỗi nhà máy trong 3-7 ngày, thanh toán lần cuối.

Điểm đặc biệt của Phase 1 sau khi soi workflow thật:

```text
ERP phải ưu tiên rất mạnh cho:
- Kho vận hằng ngày
- Đối soát cuối ca
- Bàn giao ĐVVC bằng scan/manifest
- Hàng hoàn
- QC và batch/hạn dùng
- Gia công ngoài/subcontract manufacturing
```

---

## 4. Nguyên tắc tích hợp tổng thể

### 4.1. ERP là nguồn sự thật cho nghiệp vụ lõi

ERP là nguồn sự thật cho:

- Master data sản phẩm/SKU
- Kho, tồn kho, batch/lô, hạn dùng
- Stock ledger
- QC status
- Sales order nội bộ
- Shipment nội bộ
- Return inspection/disposition
- Purchase/Subcontract order
- Audit log
- Approval log

Các hệ thống bên ngoài có thể gửi dữ liệu vào, nhưng không được tự ý thay đổi sự thật lõi trong ERP nếu chưa đi qua rule nghiệp vụ.

Ví dụ:

```text
Website gửi đơn hàng vào ERP.
Nhưng ERP mới là nơi quyết định có reserve stock được hay không.
```

```text
ĐVVC gửi trạng thái delivered.
Nhưng ERP phải kiểm tra shipment hợp lệ rồi mới chuyển trạng thái đơn hàng.
```

---

### 4.2. Không tích hợp trực tiếp vào database

Không hệ thống ngoài nào được đọc/ghi trực tiếp PostgreSQL của ERP.

Tất cả phải qua:

```text
REST API
Webhook
Scheduled import/export
Message/event integration
```

Lý do:

- Bảo vệ stock ledger
- Bảo vệ audit log
- Bảo vệ transaction
- Tránh phá trạng thái chứng từ
- Dễ trace lỗi

---

### 4.3. API-first, OpenAPI-first

Mọi API public/internal cần có OpenAPI contract trước khi build.

Mỗi integration phải có:

```text
- Endpoint
- Request schema
- Response schema
- Error code
- Auth method
- Idempotency rule
- Retry rule
- Audit requirement
```

Frontend, backend và đối tác không tự đoán contract.

---

### 4.4. Event-driven cho luồng bất đồng bộ

Các nghiệp vụ không cần trả kết quả ngay nên dùng event/outbox/worker:

```text
OrderConfirmed
StockReserved
PackingCompleted
ShipmentCreated
ShipmentHandedOver
TrackingUpdated
CODReconciled
ReturnReceived
ReturnDispositioned
QCReleased
SubcontractGoodsReceived
```

Không nhồi toàn bộ logic vào một API call dài.

Ví dụ:

```text
User bấm Confirm Packing
→ ERP cập nhật packing status
→ ghi audit log
→ phát event PackingCompleted
→ worker tạo shipment/manifest hoặc gửi notification nếu cần
```

---

### 4.5. Idempotency bắt buộc cho giao dịch nhạy cảm

Các API sau bắt buộc có idempotency:

- Tạo đơn hàng từ website/sàn
- Tạo shipment
- Xác nhận bàn giao ĐVVC
- Nhận webhook trạng thái giao hàng
- Nhận payment/COD callback
- Nhập dữ liệu hàng hoàn
- Import chứng từ thanh toán
- Tạo stock movement từ tích hợp

Nếu cùng một request gửi lại 2 lần, hệ thống không được tạo trùng đơn, trùng shipment, trùng movement hoặc trùng thanh toán.

---

### 4.6. Mọi tích hợp nghiệp vụ phải có audit log

Các hành động sau luôn cần audit log:

- Tạo/cập nhật đơn hàng từ kênh ngoài
- Reserve/release stock do tích hợp
- Tạo/cancel shipment
- Bàn giao ĐVVC
- Tracking update quan trọng
- COD reconciliation
- Return receiving/disposition
- QC status update
- Subcontract receiving/claim
- Export dữ liệu tài chính

Audit log tối thiểu:

```text
actor_type: user/system/integration
actor_id
integration_name
action
entity_type
entity_id
before_data
after_data
request_id
correlation_id
created_at
```

---

## 5. Bức tranh tích hợp tổng thể

```text
                           ┌──────────────────────────────┐
                           │        ERP Frontend           │
                           │  React / Next.js / TypeScript │
                           └───────────────┬──────────────┘
                                           │
                                           │ REST + OpenAPI
                                           ↓
┌─────────────────────────────────────────────────────────────────────┐
│                           ERP Core Backend                           │
│                    Go Modular Monolith + PostgreSQL                  │
│                                                                     │
│ Master Data | Inventory | QC | Purchase | Sales | Shipping | Return │
│ Subcontract | Finance Basic | Audit | Approval | Notification       │
└───────────────────────────┬─────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┬────────────────────┐
        │                   │                   │                    │
        ↓                   ↓                   ↓                    ↓
  Website / D2C        ĐVVC / Carrier       Payment/COD        Accounting Export
  Marketplace          Tracking/Manifest    Bank Statement     Tax/Finance Tool
        │                   │                   │                    │
        ↓                   ↓                   ↓                    ↓
  Orders/Returns       Shipment Status      Payment Status     Voucher/Report

        ┌───────────────────┬───────────────────┬────────────────────┐
        │                   │                   │                    │
        ↓                   ↓                   ↓                    ↓
 Barcode Scanner       Zalo/SMS/Email       S3/MinIO Storage   BI/Reporting
 Scan event            Notification         COA/Invoice/File   Data snapshot
```

---

## 6. Integration Registry

| Mã | Tích hợp | Mục đích | Hướng dữ liệu | Pattern | Độ ưu tiên |
|---|---|---|---|---|---|
| INT-01 | Đơn vị vận chuyển / ĐVVC | Tạo vận đơn, bàn giao, tracking, COD, hoàn hàng | 2 chiều | API + webhook + batch recon | Rất cao |
| INT-02 | Website/D2C | Nhận đơn, sync trạng thái, sync tồn khả dụng | 2 chiều | API + webhook | Cao |
| INT-03 | Marketplace | Import đơn, sync trạng thái, phí sàn, hoàn hàng | 2 chiều | API nếu có, CSV fallback | Cao |
| INT-04 | POS/Cửa hàng | Đơn bán lẻ, tồn cửa hàng, chuyển kho | 2 chiều | API hoặc batch | Trung bình/Cao |
| INT-05 | Payment Gateway/Bank/COD | Thanh toán, đối soát COD, sao kê | Vào ERP | Webhook + import file | Cao |
| INT-06 | Barcode Scanner | Quét mã nhập/xuất/đóng/bàn giao/hàng hoàn | Vào ERP | Frontend scan + API | Rất cao |
| INT-07 | Notification | Gửi cảnh báo, duyệt, lỗi sync, trạng thái đơn | Ra ngoài | Event + provider API | Trung bình |
| INT-08 | Accounting/Tax tool | Xuất dữ liệu kế toán/tài chính | Ra ngoài | Export file/API | Trung bình |
| INT-09 | Supplier/Factory/Subcontract | Đơn gia công, chuyển NVL/bao bì, nhận hàng, claim | Chủ yếu nội bộ, future portal | Manual + export/import | Cao |
| INT-10 | File Storage | Lưu COA, chứng từ, ảnh QC, biên bản, hợp đồng | 2 chiều | S3 API | Cao |
| INT-11 | BI/Reporting | Snapshot dữ liệu báo cáo | Ra ngoài | Scheduled export/API | Trung bình |
| INT-12 | Identity/SSO | Đăng nhập, user directory | Vào ERP | OAuth/OIDC future | Thấp/Phase sau |

---

## 7. Chuẩn kỹ thuật chung cho mọi tích hợp

### 7.1. Base API

```text
/api/v1
```

Ví dụ:

```text
POST /api/v1/integrations/website/orders
POST /api/v1/integrations/carriers/{carrier_code}/shipments
POST /api/v1/integrations/carriers/webhooks/tracking
POST /api/v1/integrations/payments/webhooks
```

---

### 7.2. Header chuẩn

```http
X-Request-ID: uuid
X-Correlation-ID: uuid
X-Idempotency-Key: unique-key
X-Client-Code: website-main
X-Signature: hmac-signature
X-Timestamp: 2026-04-24T10:00:00+07:00
```

Ý nghĩa:

| Header | Ý nghĩa |
|---|---|
| X-Request-ID | ID của từng request |
| X-Correlation-ID | ID gom toàn bộ flow liên quan |
| X-Idempotency-Key | Chống tạo trùng giao dịch |
| X-Client-Code | Xác định nguồn tích hợp |
| X-Signature | Xác thực webhook/API từ bên ngoài |
| X-Timestamp | Chống replay request |

---

### 7.3. Response envelope chuẩn

Success:

```json
{
  "success": true,
  "data": {
    "id": "ord_123",
    "status": "CONFIRMED"
  },
  "meta": {
    "request_id": "req_001",
    "timestamp": "2026-04-24T10:00:00+07:00"
  }
}
```

Error:

```json
{
  "success": false,
  "code": "INSUFFICIENT_STOCK",
  "message": "Tồn khả dụng không đủ",
  "details": {
    "sku_code": "SERUM-VITC-30ML",
    "requested_qty": 10,
    "available_qty": 6
  },
  "meta": {
    "request_id": "req_001",
    "timestamp": "2026-04-24T10:00:00+07:00"
  }
}
```

---

### 7.4. Error code nhóm tích hợp

| Nhóm | Ví dụ code | Ý nghĩa |
|---|---|---|
| Auth | `INVALID_SIGNATURE`, `EXPIRED_TIMESTAMP` | Sai xác thực |
| Idempotency | `DUPLICATE_REQUEST`, `IDEMPOTENCY_CONFLICT` | Request trùng hoặc key bị dùng sai |
| Validation | `INVALID_PAYLOAD`, `MISSING_REQUIRED_FIELD` | Payload sai |
| Business | `INSUFFICIENT_STOCK`, `BATCH_ON_HOLD`, `ORDER_NOT_PACKED` | Không qua rule nghiệp vụ |
| Mapping | `UNKNOWN_SKU`, `UNKNOWN_CARRIER`, `UNKNOWN_STATUS` | Không map được dữ liệu |
| External | `CARRIER_API_FAILED`, `PAYMENT_PROVIDER_TIMEOUT` | Đối tác lỗi |
| Reconciliation | `COD_AMOUNT_MISMATCH`, `MANIFEST_COUNT_MISMATCH` | Đối soát lệch |

---

### 7.5. Retry policy

| Loại lỗi | Retry? | Quy tắc |
|---|---|---|
| Timeout | Có | retry 3-5 lần, exponential backoff |
| 5xx external | Có | retry + đưa vào queue |
| 429 rate limit | Có | retry theo `Retry-After` |
| 400 validation | Không | ghi lỗi, cần sửa dữ liệu |
| 401/403 | Không | báo security/admin |
| Business rule fail | Không | báo owner nghiệp vụ |
| Duplicate idempotency | Không | trả kết quả cũ nếu hợp lệ |

Nếu retry thất bại:

```text
→ đưa vào dead-letter queue
→ tạo integration incident
→ notify owner
→ cho phép replay có kiểm soát
```

---

### 7.6. Timezone và định dạng thời gian

ERP sử dụng:

```text
Asia/Ho_Chi_Minh
```

API trả ISO 8601:

```text
2026-04-24T10:00:00+07:00
```

Database lưu `timestamptz`.

Không dùng date/time không timezone cho dữ liệu giao dịch.

---

### 7.7. Naming convention

| Loại | Quy tắc | Ví dụ |
|---|---|---|
| Integration code | lowercase kebab/snake | `ghtk`, `ghn`, `website-main` |
| External ID | giữ nguyên từ hệ ngoài | `external_order_id` |
| Internal ID | UUID hoặc ERP code | `order_id`, `shipment_id` |
| Event | Past tense | `ShipmentHandedOver` |
| Webhook endpoint | noun/action rõ | `/webhooks/tracking` |

---

## 8. INT-01: Đơn vị vận chuyển / Carrier Integration

### 8.1. Mục tiêu

Tích hợp ĐVVC để hỗ trợ:

- Tạo vận đơn
- In/nhận nhãn vận chuyển
- Gom đơn theo carrier/manifest
- Bàn giao bằng scan và đối chiếu số lượng
- Nhận tracking update
- Theo dõi giao thành công/thất bại
- Xử lý hoàn hàng
- Đối soát COD và phí vận chuyển

Đây là tích hợp rất quan trọng vì workflow thực tế có bước chia khu vực để hàng, để theo thùng/rổ, đối chiếu số lượng đơn, quét mã trước khi bàn giao và xử lý trường hợp chưa đủ đơn.

---

### 8.2. Đối tượng dữ liệu

| Object | Ý nghĩa |
|---|---|
| Carrier | Đơn vị vận chuyển |
| Carrier Service | Gói dịch vụ giao hàng |
| Shipment | Vận đơn nội bộ |
| External Shipment | Mã vận đơn của ĐVVC |
| Shipping Label | Nhãn giao hàng |
| Carrier Manifest | Bảng/batch bàn giao cho ĐVVC |
| Handover Scan Event | Sự kiện quét khi bàn giao |
| Tracking Event | Trạng thái vận chuyển |
| COD Reconciliation | Đối soát tiền COD |
| Return Shipment | Đơn hoàn từ ĐVVC |

---

### 8.3. Luồng tạo vận đơn

```text
Sales Order Confirmed
→ Reserve stock
→ Pick task
→ Pack order
→ Create shipment
→ Generate/receive label
→ Put into carrier staging area
→ Add to carrier manifest
```

Rule:

```text
Chỉ được tạo shipment khi order đã confirmed/reserved và đủ thông tin giao hàng.
Không cho tạo shipment nếu batch bị HOLD/FAIL.
Không cho tạo shipment nếu order đã bị cancel.
```

---

### 8.4. Luồng bàn giao ĐVVC

```text
Packed orders
→ phân chia khu vực theo ĐVVC
→ để theo thùng/rổ
→ đối chiếu số lượng đơn theo manifest
→ scan mã đơn/mã vận đơn
→ nếu đủ: ký xác nhận/bàn giao
→ nếu chưa đủ: kiểm tra mã, tìm lại khu đóng hàng, xử lý thiếu
→ close manifest
→ update shipment = HANDED_OVER
```

Rule bắt buộc:

- Mỗi đơn trong manifest phải được scan trước khi close manifest.
- Nếu số đơn scan khác số đơn trên manifest → không cho close, trừ khi trưởng kho override có lý do.
- Override phải ghi audit log và lý do.
- Shipment chỉ được chuyển sang `HANDED_OVER` sau khi manifest được close.

---

### 8.5. Shipment status nội bộ

```text
DRAFT
CREATED
LABEL_READY
PACKED
READY_TO_HANDOVER
HANDED_OVER
IN_TRANSIT
DELIVERED
FAILED_DELIVERY
RETURNING
RETURNED
CANCELLED
LOST
```

---

### 8.6. Tracking mapping

| External status | Internal status | Ghi chú |
|---|---|---|
| created | CREATED | Đã tạo vận đơn |
| picked_up / handed_over | HANDED_OVER | ĐVVC đã nhận |
| in_transit | IN_TRANSIT | Đang vận chuyển |
| delivered | DELIVERED | Giao thành công |
| failed | FAILED_DELIVERY | Giao thất bại |
| returning | RETURNING | Đang hoàn |
| returned | RETURNED | Đã hoàn về |
| cancelled | CANCELLED | Hủy vận đơn |
| lost | LOST | Thất lạc |

Nếu external status không map được:

```text
→ lưu raw status
→ không tự đổi internal status
→ tạo alert UNKNOWN_CARRIER_STATUS
```

---

### 8.7. API đề xuất

```text
POST /api/v1/shipping/shipments
POST /api/v1/shipping/shipments/{id}/cancel
GET  /api/v1/shipping/shipments/{id}/label
GET  /api/v1/shipping/shipments/{id}/tracking
POST /api/v1/shipping/manifests
POST /api/v1/shipping/manifests/{id}/scan
POST /api/v1/shipping/manifests/{id}/close
POST /api/v1/integrations/carriers/webhooks/tracking
POST /api/v1/integrations/carriers/cod-reconciliation/import
```

---

### 8.8. Payload mẫu: carrier tracking webhook

```json
{
  "carrier_code": "sample_carrier",
  "external_tracking_no": "VN123456789",
  "external_status": "delivered",
  "status_time": "2026-04-24T15:20:00+07:00",
  "location": "Hồ Chí Minh",
  "raw_payload": {}
}
```

---

### 8.9. Kiểm soát rủi ro

| Rủi ro | Guardrail |
|---|---|
| Bàn giao thiếu đơn | scan từng đơn + manifest count |
| Đơn đã packed nhưng chưa bàn giao | dashboard Ready to Handover aging |
| Tracking webhook trùng | idempotency theo carrier + tracking_no + status_time |
| COD lệch | COD reconciliation bắt buộc |
| Đơn hoàn không được ghi nhận | return webhook/import + return receiving scan |
| ĐVVC báo delivered nhưng order chưa handed over | chặn hoặc tạo exception |

---

## 9. INT-02: Website / D2C Integration

### 9.1. Mục tiêu

Website/D2C gửi đơn hàng vào ERP và nhận trạng thái xử lý/giao hàng từ ERP.

ERP phải là nguồn sự thật cho:

- tồn khả dụng
- reserve stock
- trạng thái xử lý nội bộ
- trạng thái giao hàng sau khi nhận từ ĐVVC
- đổi trả nếu có

---

### 9.2. Luồng order inbound

```text
Khách đặt hàng trên website
→ website gửi order vào ERP
→ ERP validate SKU/customer/payment/shipping address
→ ERP kiểm tra tồn khả dụng
→ ERP reserve stock
→ tạo sales order
→ trả ERP order_id/status cho website
```

Nếu không đủ tồn:

```text
→ trả INSUFFICIENT_STOCK
→ website hiển thị hết hàng hoặc backorder tùy chính sách
```

---

### 9.3. API đề xuất

```text
POST /api/v1/integrations/website/orders
GET  /api/v1/integrations/website/orders/{external_order_id}
POST /api/v1/integrations/website/orders/{external_order_id}/cancel
GET  /api/v1/integrations/website/products/stock?sku_code=...
POST /api/v1/integrations/website/webhooks/payment
```

---

### 9.4. Payload mẫu: tạo order

```json
{
  "external_order_id": "WEB-100001",
  "channel_code": "website",
  "order_time": "2026-04-24T10:00:00+07:00",
  "customer": {
    "name": "Nguyen Van A",
    "phone": "0900000000",
    "email": "a@example.com"
  },
  "shipping_address": {
    "province": "TP.HCM",
    "district": "Quận 1",
    "ward": "Bến Nghé",
    "address_line": "123 Nguyễn Huệ"
  },
  "items": [
    {
      "sku_code": "SERUM-VITC-30ML",
      "qty": 2,
      "unit_price": 300000,
      "discount_amount": 50000
    }
  ],
  "payment_method": "COD",
  "total_amount": 550000
}
```

---

### 9.5. Trạng thái order gửi ngược website

```text
CONFIRMED
RESERVED
PICKING
PACKED
HANDED_OVER
DELIVERED
FAILED_DELIVERY
RETURNING
RETURNED
CANCELLED
```

---

### 9.6. Guardrail

- Không cho website tự trừ tồn.
- Không cho website tự xác nhận giao hàng nếu ERP chưa nhận tracking từ carrier.
- Order từ website phải có idempotency key.
- SKU từ website phải map được với SKU ERP.
- Discount/promotion phải lưu rõ nguồn để tính lãi thật.

---

## 10. INT-03: Marketplace Integration

### 10.1. Mục tiêu

Kết nối sàn/marketplace để nhận đơn, trạng thái, phí sàn, hoàn hàng và đối soát.

Phase 1 có thể làm theo 2 mức:

```text
Mức 1: CSV/Excel import-export bán tự động
Mức 2: API connector với từng sàn
```

Khuyến nghị Phase 1:

```text
Làm chuẩn dữ liệu + import/export trước.
API connector làm khi quy trình nội bộ đã ổn.
```

---

### 10.2. Dữ liệu cần nhận từ marketplace

| Dữ liệu | Bắt buộc? | Ghi chú |
|---|---|---|
| external_order_id | Có | mã đơn sàn |
| marketplace_code | Có | Shopee/Lazada/TikTok... |
| customer masked info | Tùy | nhiều sàn ẩn thông tin |
| item/SKU | Có | phải map SKU ERP |
| quantity | Có |  |
| selling price | Có |  |
| discount/platform voucher | Có | để tính lợi nhuận |
| shipping fee | Có | nếu có |
| platform fee | Nên có | để tính lãi thật |
| payment status | Có |  |
| return/refund status | Có |  |

---

### 10.3. API/import đề xuất

```text
POST /api/v1/integrations/marketplaces/orders/import
POST /api/v1/integrations/marketplaces/returns/import
POST /api/v1/integrations/marketplaces/fees/import
GET  /api/v1/integrations/marketplaces/stock-export
GET  /api/v1/integrations/marketplaces/order-status-export
```

---

### 10.4. Guardrail

- Đơn sàn không được tạo SKU mới tự động nếu SKU chưa map.
- Nếu không map được SKU → đưa vào exception queue.
- Phí sàn/hoàn tiền phải có khả năng nhập sau để tính profit đúng.
- Hàng hoàn từ sàn vẫn phải qua return inspection, không nhập thẳng available stock.

---

## 11. INT-04: POS / Retail Store Integration

### 11.1. Mục tiêu

Nếu công ty có cửa hàng bán lẻ, POS cần sync với ERP để quản lý:

- đơn bán tại quầy
- tồn kho cửa hàng
- chuyển kho từ kho trung tâm xuống cửa hàng
- đổi trả tại cửa hàng
- doanh thu theo ca
- tiền mặt/chuyển khoản/voucher

---

### 11.2. Mô hình dữ liệu

ERP nên là nguồn sự thật cho:

- SKU
- batch/hạn dùng
- giá bán chuẩn
- tồn kho tổng
- chuyển kho

POS là nguồn phát sinh cho:

- giao dịch bán lẻ tại cửa hàng
- ca bán hàng
- thanh toán tại quầy

---

### 11.3. API đề xuất

```text
GET  /api/v1/integrations/pos/products
GET  /api/v1/integrations/pos/prices
POST /api/v1/integrations/pos/sales
POST /api/v1/integrations/pos/returns
GET  /api/v1/integrations/pos/store-stock
POST /api/v1/integrations/pos/shift-closing
```

---

### 11.4. Guardrail

- POS không được bán SKU/batch đang bị QC hold/fail.
- Không cho bán âm tồn nếu policy không cho phép.
- Ca POS phải có shift closing tương tự logic đóng ca kho.
- Đổi trả tại cửa hàng phải link về hóa đơn/đơn hàng gốc nếu có.

---

## 12. INT-05: Payment / COD / Bank Integration

### 12.1. Mục tiêu

Tích hợp thanh toán để:

- cập nhật payment status
- đối soát COD từ ĐVVC
- import sao kê ngân hàng
- ghi nhận công nợ đã thu
- phát hiện lệch tiền

---

### 12.2. Payment status

```text
UNPAID
PENDING
PAID
PARTIALLY_PAID
REFUNDED
PARTIALLY_REFUNDED
FAILED
CANCELLED
```

---

### 12.3. COD reconciliation flow

```text
Carrier gửi bảng COD
→ ERP import COD file/API
→ match theo tracking_no/order_id
→ so expected COD vs actual COD
→ nếu khớp: mark reconciled
→ nếu lệch: tạo exception
→ finance xử lý exception
```

---

### 12.4. API/import đề xuất

```text
POST /api/v1/integrations/payments/webhooks
POST /api/v1/integrations/bank-statements/import
POST /api/v1/integrations/carriers/cod-reconciliation/import
GET  /api/v1/finance/reconciliation/cod-exceptions
POST /api/v1/finance/reconciliation/{id}/resolve
```

---

### 12.5. Guardrail

- Không ghi nhận thu tiền nếu không match được order/shipment.
- COD lệch phải vào exception, không tự sửa doanh thu.
- Payment webhook phải chống replay bằng signature + timestamp.
- Refund phải link với return/refund request.

---

## 13. INT-06: Barcode Scanner Integration

### 13.1. Mục tiêu

Scanner là “bàn tay” của ERP trong kho. Phase 1 phải hỗ trợ tốt các luồng:

- nhận hàng
- kiểm tra batch/hạn dùng
- picking
- packing
- bàn giao ĐVVC
- hàng hoàn
- kiểm kê cuối ngày
- chuyển NVL/bao bì cho nhà máy gia công

---

### 13.2. Cách tích hợp

Ưu tiên dùng scanner dạng keyboard wedge hoặc mobile scanner:

```text
Scanner đọc mã → frontend nhận input → gọi ERP scan API → ERP validate rule → trả kết quả ngay
```

Không cần thiết bị quá phức tạp ở Phase 1. Cái quan trọng là UX phải nhanh, rõ đúng/sai, không bắt nhân viên kho click nhiều.

---

### 13.3. Scan event chuẩn

```json
{
  "scan_context": "SHIPPING_HANDOVER",
  "barcode": "VN123456789",
  "operator_id": "user_001",
  "device_id": "scanner_01",
  "location_code": "WH-HCM-STAGING-GHN",
  "scanned_at": "2026-04-24T10:00:00+07:00"
}
```

---

### 13.4. Scan context

| Context | Mục đích |
|---|---|
| INBOUND_RECEIVING | nhận hàng nhập kho |
| QC_SAMPLING | lấy mẫu QC |
| PICKING | lấy hàng |
| PACKING | đóng hàng |
| SHIPPING_HANDOVER | bàn giao ĐVVC |
| RETURN_RECEIVING | nhận hàng hoàn |
| STOCK_COUNT | kiểm kê |
| SUBCONTRACT_TRANSFER | chuyển NVL/bao bì cho nhà máy |
| SUBCONTRACT_RECEIVING | nhận hàng gia công |

---

### 13.5. API đề xuất

```text
POST /api/v1/scan/validate
POST /api/v1/scan/events
POST /api/v1/shipping/manifests/{id}/scan
POST /api/v1/returns/receiving/scan
POST /api/v1/inventory/stock-counts/{id}/scan
POST /api/v1/subcontract/transfers/{id}/scan
```

---

### 13.6. UX rule

Mỗi scan phải trả một trong các kết quả:

```text
VALID
INVALID
DUPLICATE
WRONG_CONTEXT
NOT_IN_MANIFEST
BATCH_ON_HOLD
EXPIRED_OR_NEAR_EXPIRY
NOT_FOUND
```

Kết quả phải hiện rõ bằng màu/trạng thái/âm thanh nếu frontend hỗ trợ.

---

## 14. INT-07: Notification Integration

### 14.1. Mục tiêu

Gửi thông báo đúng người, đúng lúc, không spam.

Kênh notification:

```text
- In-app notification
- Email
- Zalo OA/Zalo nội bộ nếu có
- SMS nếu cần
- Webhook nội bộ/future
```

---

### 14.2. Event cần thông báo

| Event | Người nhận | Mục đích |
|---|---|---|
| PO_APPROVAL_REQUIRED | manager/finance | duyệt mua hàng |
| QC_FAILED | QA/warehouse/purchasing | xử lý hàng lỗi |
| BATCH_HOLD | sales/warehouse | chặn bán/xuất |
| ORDER_READY_TO_PICK | warehouse | tạo picking |
| PACKING_COMPLETED | shipping team | chuẩn bị bàn giao |
| MANIFEST_COUNT_MISMATCH | trưởng kho | xử lý thiếu/thừa đơn |
| SHIPMENT_DELIVERED | sales/CSKH | cập nhật khách |
| RETURN_RECEIVED | CSKH/warehouse | xử lý hàng hoàn |
| COD_MISMATCH | finance | đối soát tiền |
| SUBCONTRACT_CLAIM_DEADLINE | production/purchasing | nhắc phản hồi nhà máy trong 3-7 ngày |

---

### 14.3. Guardrail

- Notification không thay thế workflow approval.
- Mọi notification quan trọng phải link về entity gốc.
- Không gửi dữ liệu nhạy cảm qua kênh không an toàn.
- Mỗi event có rule chống gửi trùng.

---

## 15. INT-08: Accounting / Finance Export

### 15.1. Mục tiêu

Phase 1 chưa cần tích hợp kế toán quá sâu. Nên ưu tiên xuất dữ liệu rõ, sạch, có thể kiểm tra.

Dữ liệu xuất:

- doanh thu bán hàng
- thu tiền/COD/payment
- công nợ phải thu cơ bản
- công nợ nhà cung cấp cơ bản
- nhập kho mua hàng
- xuất kho bán hàng
- hàng hoàn
- tồn kho và giá trị tồn
- chi phí vận chuyển/COD nếu có
- dữ liệu đối soát

---

### 15.2. Hình thức

```text
Mức 1: Export Excel/CSV chuẩn
Mức 2: API sang phần mềm kế toán
Mức 3: Accounting posting tự động theo chart of accounts
```

Khuyến nghị Phase 1:

```text
Mức 1 + chuẩn hóa dữ liệu.
Không tự động hóa sâu nếu kế toán chưa chốt hạch toán.
```

---

### 15.3. API/export đề xuất

```text
GET /api/v1/finance/exports/sales
GET /api/v1/finance/exports/payments
GET /api/v1/finance/exports/purchases
GET /api/v1/finance/exports/inventory-movements
GET /api/v1/finance/exports/cod-reconciliation
GET /api/v1/finance/exports/returns
```

---

### 15.4. Guardrail

- Export phải có khoảng thời gian và người export.
- Export dữ liệu tài chính phải ghi audit log.
- Không cho user thường export giá vốn/lợi nhuận.
- Dữ liệu export phải có entity ID để trace ngược.

---

## 16. INT-09: Supplier / Factory / Subcontract Integration

### 16.1. Mục tiêu

Workflow thực tế cho thấy công ty có mô hình sản xuất/gia công ngoài. Phase 1 cần thiết kế tích hợp/luồng dữ liệu cho nhà máy, dù ban đầu có thể vận hành bằng export/import và file đính kèm thay vì portal.

---

### 16.2. Đối tượng dữ liệu

| Object | Ý nghĩa |
|---|---|
| Subcontract Order | Đơn đặt gia công/sản xuất với nhà máy |
| Material Transfer | Chuyển NVL/bao bì sang nhà máy |
| Sample Approval | Chốt mẫu trước sản xuất hàng loạt |
| Subcontract Production Batch | Batch hàng gia công |
| Inbound Receipt | Nhận hàng về kho |
| QC Inspection | Kiểm số lượng/chất lượng |
| Factory Claim | Báo lỗi/claim nhà máy trong 3-7 ngày |
| Final Payment | Thanh toán lần cuối sau nghiệm thu |

---

### 16.3. Luồng gia công ngoài

```text
Tạo đơn với nhà máy
→ xác nhận số lượng/quy cách/mẫu mã
→ cọc đơn/chốt thời gian
→ chuyển NVL/bao bì sang nhà máy
→ ký biên bản bàn giao + chứng từ kèm theo
→ làm mẫu/chốt mẫu
→ sản xuất hàng loạt
→ giao hàng về kho
→ kiểm tra số lượng/chất lượng
→ nếu đạt: nhận hàng/nhập kho/QC
→ nếu không đạt: tạo claim gửi nhà máy trong 3-7 ngày
→ thanh toán lần cuối
```

---

### 16.4. Tích hợp Phase 1

Phase 1 nên hỗ trợ:

```text
- Export subcontract order cho nhà máy
- Export material transfer note
- Upload biên bản/chứng từ
- Upload mẫu/ảnh/COA/MSDS nếu có
- Nhận hàng về ERP bằng inbound receiving
- Tạo claim nhà máy trong ERP
- Theo dõi deadline phản hồi lỗi 3-7 ngày
```

Future phase:

```text
- Supplier/factory portal
- Nhà máy xác nhận trực tiếp
- Nhà máy upload progress/status
- Nhà máy upload COA/chứng từ
```

---

### 16.5. API/export đề xuất

```text
GET  /api/v1/subcontract/orders/{id}/export
GET  /api/v1/subcontract/transfers/{id}/export
POST /api/v1/subcontract/orders/{id}/attachments
POST /api/v1/subcontract/orders/{id}/sample-approval
POST /api/v1/subcontract/receipts
POST /api/v1/subcontract/claims
GET  /api/v1/subcontract/claims/open-deadlines
```

---

### 16.6. Guardrail

- Không cho sản xuất hàng loạt nếu sample chưa approved.
- Material transfer phải tạo stock movement loại `SUBCONTRACT_TRANSFER_OUT`.
- Khi nhận hàng về phải qua inbound receiving + QC.
- Hàng chưa QC pass không vào available stock.
- Claim nhà máy phải có deadline và owner.

---

## 17. INT-10: File Storage / Document Storage

### 17.1. Mục tiêu

ERP cần lưu và truy xuất file chứng từ:

- phiếu nhập/xuất
- biên bản bàn giao
- hóa đơn/chứng từ giao hàng
- COA/MSDS/spec
- ảnh QC
- ảnh hàng hoàn
- shipping label
- carrier manifest
- hợp đồng/PO/subcontract document
- mẫu/chốt mẫu
- claim nhà máy

---

### 17.2. Storage

Khuyến nghị:

```text
S3-compatible storage / MinIO
```

Không lưu file binary trực tiếp trong PostgreSQL.

Database chỉ lưu metadata:

```text
file_id
entity_type
entity_id
file_name
file_type
mime_type
storage_key
uploaded_by
uploaded_at
checksum
visibility
```

---

### 17.3. API đề xuất

```text
POST /api/v1/files/upload
GET  /api/v1/files/{id}/download-url
POST /api/v1/files/{id}/link
DELETE /api/v1/files/{id}
```

Xóa file quan trọng cần soft delete + audit.

---

## 18. INT-11: BI / Reporting Export

### 18.1. Mục tiêu

Dữ liệu ERP cần phục vụ dashboard và BI.

Phase 1 ưu tiên báo cáo nội bộ trong ERP. Nhưng vẫn cần chuẩn export nếu sau này dùng BI ngoài.

---

### 18.2. Data snapshot

Các snapshot nên có:

```text
- daily_sales_snapshot
- inventory_balance_snapshot
- stock_aging_snapshot
- shipment_status_snapshot
- return_snapshot
- qc_batch_snapshot
- purchase_receipt_snapshot
- subcontract_progress_snapshot
- cod_reconciliation_snapshot
```

---

### 18.3. Guardrail

- BI chỉ đọc dữ liệu/snapshot, không ghi ngược vào ERP.
- Dữ liệu nhạy cảm như giá vốn, lương, payout phải theo permission.
- Snapshot phải có timestamp và version.

---

## 19. Event Catalog tích hợp

| Event | Khi nào phát sinh | Consumer |
|---|---|---|
| SalesOrderCreated | đơn từ website/sàn vào ERP | inventory, notification |
| SalesOrderConfirmed | đơn được xác nhận | inventory |
| StockReserved | giữ tồn thành công | warehouse |
| PickTaskCreated | tạo nhiệm vụ lấy hàng | warehouse UI |
| PackingCompleted | đóng hàng xong | shipping |
| ShipmentCreated | tạo vận đơn | carrier adapter |
| ManifestCreated | tạo bảng bàn giao | warehouse |
| ManifestScanCompleted | scan đủ đơn trong manifest | warehouse/shipping |
| ShipmentHandedOver | bàn giao ĐVVC | sales/CSKH/finance |
| TrackingUpdated | carrier gửi trạng thái | sales/CSKH |
| CODReconciled | đối soát COD xong | finance |
| ReturnReceived | nhận hàng hoàn | CSKH/warehouse |
| ReturnDispositioned | phân loại hàng hoàn | inventory/finance |
| PurchaseOrderApproved | PO được duyệt | purchasing/warehouse |
| GoodsReceived | nhận hàng | QC/inventory |
| QCReleased | QC pass/fail/hold | inventory/sales |
| SubcontractOrderCreated | tạo đơn gia công | purchasing/production |
| SubcontractMaterialTransferred | chuyển NVL/bao bì | inventory/finance |
| SampleApproved | duyệt mẫu | production/factory |
| SubcontractGoodsReceived | nhận hàng gia công | QC/inventory |
| FactoryClaimCreated | claim nhà máy | purchasing/production |

---

## 20. Reconciliation bắt buộc

### 20.1. Order reconciliation

Mục tiêu: đảm bảo đơn từ website/sàn không bị thiếu/trùng.

```text
Nguồn ngoài gửi danh sách đơn trong ngày
ERP so với order đã nhận
Nếu thiếu → import lại
Nếu trùng → kiểm idempotency
Nếu lệch tổng tiền/SKU → tạo exception
```

Tần suất:

```text
hằng ngày hoặc theo ca
```

---

### 20.2. Manifest reconciliation

Mục tiêu: đảm bảo đơn packed đã được bàn giao đúng ĐVVC.

```text
Số đơn trong manifest
vs
Số scan thành công
vs
Số ĐVVC xác nhận nhận
```

Nếu lệch:

```text
→ không close manifest hoặc tạo exception nếu đã override
```

---

### 20.3. COD reconciliation

Mục tiêu: đảm bảo tiền COD về đúng.

```text
Expected COD từ ERP
vs
Actual COD từ ĐVVC/bank
vs
Fee/adjustment/return
```

Nếu lệch:

```text
→ tạo COD_MISMATCH
→ finance xử lý
```

---

### 20.4. Return reconciliation

Mục tiêu: đảm bảo hàng hoàn được nhận và xử lý đúng trạng thái.

```text
Carrier returned list
vs
ERP return receiving
vs
Return disposition
vs
Stock movement
```

---

### 20.5. Subcontract material balance

Mục tiêu: kiểm soát NVL/bao bì chuyển ra nhà máy.

```text
Material transferred out
- material consumed/returned
= material balance at factory
```

Phase 1 có thể theo dõi ở mức đơn giản, nhưng phải có dữ liệu nền.

---

## 21. Integration Monitoring

### 21.1. Dashboard cần có

```text
- số request thành công/thất bại theo integration
- webhook pending/error
- dead-letter queue
- đơn từ website/sàn chưa import
- shipment tạo lỗi
- tracking update lỗi
- manifest mismatch
- COD mismatch
- return mismatch
- subcontract claim gần quá hạn
```

---

### 21.2. Alert rule

| Alert | Khi nào báo | Người nhận |
|---|---|---|
| Carrier API failed | lỗi liên tục > 5 phút | Tech + Warehouse lead |
| Order import failed | đơn không vào ERP | Tech + Sales admin |
| Manifest mismatch | scan thiếu/thừa | Trưởng kho |
| COD mismatch | lệch tiền | Finance |
| Return not received | carrier báo returned nhưng kho chưa nhận | Warehouse + CSKH |
| Unknown SKU | đơn ngoài có SKU không map | Sales admin + Master data |
| Factory claim deadline | gần hết 3-7 ngày claim | Production/Purchasing |

---

## 22. Integration Security

### 22.1. Auth method

| Loại tích hợp | Auth khuyến nghị |
|---|---|
| Website API | API key + HMAC signature |
| Carrier webhook | HMAC signature + IP allowlist nếu có |
| Payment webhook | HMAC signature + timestamp |
| Internal frontend | session/JWT theo auth ERP |
| Accounting export | user permission + audit |
| File download | signed URL có hạn |

---

### 22.2. Sensitive data

Không gửi dữ liệu nhạy cảm không cần thiết cho bên ngoài:

- giá vốn
- lợi nhuận
- công nợ nội bộ
- dữ liệu lương
- dữ liệu supplier price không liên quan
- batch lỗi nội bộ nếu không cần

---

### 22.3. Replay protection

Webhook phải kiểm:

```text
signature hợp lệ
request timestamp không quá hạn
idempotency key chưa xử lý conflict
payload checksum nếu cần
```

---

## 23. Integration Test Strategy

### 23.1. Test môi trường

Mỗi integration nên có:

```text
- sandbox/mock provider
- staging credential
- test data cố định
- replay webhook test
- failure simulation
```

---

### 23.2. Test case bắt buộc

| Nhóm | Test case |
|---|---|
| Website order | tạo đơn mới, tạo trùng, SKU không tồn tại, thiếu tồn, hủy đơn |
| Carrier | tạo shipment, lấy label, tracking update, delivered, failed, returned |
| Manifest | scan đủ, scan thiếu, scan trùng, scan sai carrier, close manifest |
| COD | khớp tiền, lệch tiền, thiếu order, trùng payment |
| Return | nhận hàng hoàn, phân loại còn dùng/không dùng, tạo stock movement |
| Subcontract | tạo đơn gia công, chuyển NVL, duyệt mẫu, nhận hàng, QC fail, claim nhà máy |
| File | upload/download, permission denied, signed URL hết hạn |
| Security | sai signature, replay request, expired timestamp |

---

## 24. Implementation Priority

### 24.1. Ưu tiên Phase 1A

```text
1. Scan API nội bộ cho kho
2. Shipping manifest + handover scan
3. Carrier integration adapter hoặc CSV fallback
4. Website order inbound API
5. Return receiving/disposition API
6. COD reconciliation import
```

### 24.2. Ưu tiên Phase 1B

```text
7. Marketplace import/export
8. Notification integration
9. Accounting export
10. Subcontract document export/import
11. BI snapshot export
```

### 24.3. Phase sau

```text
- Full marketplace API connector
- Supplier/factory portal
- Dealer portal
- KOL/Affiliate integration
- Full accounting posting
- SSO/OIDC
```

---

## 25. Definition of Done cho một integration

Một integration chỉ được coi là xong khi có đủ:

```text
- OpenAPI contract hoặc file format spec
- Auth/signature rule
- Request/response schema
- Error code
- Idempotency rule
- Retry rule
- Audit log
- Monitoring/alert
- Reconciliation nếu liên quan tiền/hàng
- Unit test
- Integration test
- UAT case
- Documentation cho vận hành
- Rollback/manual fallback
```

Không nhận bàn giao kiểu “API chạy được là xong”.

---

## 26. Manual fallback bắt buộc

Với ERP vận hành thật, tích hợp nào cũng phải có fallback.

| Tích hợp | Fallback |
|---|---|
| Carrier API lỗi | tạo vận đơn ngoài, import tracking/manifest sau |
| Website order API lỗi | export/import đơn tạm thời |
| Marketplace API lỗi | import CSV |
| Payment webhook lỗi | import sao kê/COD file |
| Scanner lỗi | nhập mã thủ công có quyền và audit |
| File storage lỗi | không cho hoàn tất chứng từ bắt buộc file, hoặc queue upload lại |
| Notification lỗi | không chặn nghiệp vụ, nhưng ghi log và retry |

Rule:

```text
Fallback không được phá audit log.
Fallback không được bỏ qua approval/permission.
Fallback không được sửa trực tiếp stock ledger.
```

---

## 27. Checklist trước khi build tích hợp

Trước khi build từng integration, PM/BA/Tech Lead phải chốt:

```text
[ ] Hệ thống ngoài là gì?
[ ] Ai là owner phía đối tác?
[ ] API/document có chưa?
[ ] Có sandbox không?
[ ] Dữ liệu nào inbound?
[ ] Dữ liệu nào outbound?
[ ] ERP hay hệ ngoài là source of truth?
[ ] Tần suất sync?
[ ] Cần realtime không?
[ ] Có webhook không?
[ ] Có idempotency không?
[ ] Có reconciliation không?
[ ] Nếu lỗi thì ai xử lý?
[ ] Có fallback không?
[ ] Có UAT case không?
```

---

## 28. Kết luận

Phase 1 của ERP mỹ phẩm không chỉ cần module nội bộ. Nó cần tích hợp đúng những điểm làm hàng và tiền chạy ngoài biên ERP:

```text
- đơn từ website/sàn
- vận đơn và ĐVVC
- scan bàn giao
- COD/payment
- hàng hoàn
- file chứng từ
- export kế toán
- gia công ngoài
```

Tích hợp phải tuân theo 5 nguyên tắc sống còn:

```text
1. ERP là nguồn sự thật cho hàng, tiền, batch, QC, trạng thái nội bộ.
2. Không hệ thống nào ghi trực tiếp database ERP.
3. Mọi giao dịch nhạy cảm phải có idempotency + audit log.
4. Luồng tiền/hàng phải có reconciliation.
5. Mọi integration phải có fallback vận hành.
```

Nếu làm đúng, ERP không chỉ “nối được với bên ngoài”, mà còn kiểm soát được những điểm dễ thất thoát nhất: giao hàng, hàng hoàn, COD, hàng gia công, batch/QC và chứng từ.

