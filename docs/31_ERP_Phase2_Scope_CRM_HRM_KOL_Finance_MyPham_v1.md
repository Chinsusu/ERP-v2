# 31 — ERP Phase 2 Scope: CRM, HRM, KOL/Affiliate, Finance nâng cao & BI

**Dự án:** Web ERP cho công ty mỹ phẩm — sản xuất/gia công, phân phối, bán lẻ, vận hành kho và bán hàng đa kênh  
**Tài liệu:** `31_ERP_Phase2_Scope_CRM_HRM_KOL_Finance_MyPham_v1.md`  
**Phiên bản:** v1.0  
**Ngôn ngữ:** Vietnamese  
**Vai trò tài liệu:** Xác định phạm vi Phase 2 sau khi Phase 1 đã khóa lõi vận hành hàng hóa, kho, QC, bán hàng, bàn giao ĐVVC và gia công ngoài.

---

## 1. Mục tiêu của Phase 2

Phase 1 đã tập trung vào phần “xương sống”: master data, mua hàng, QC, kho, sản xuất/gia công ngoài, bán hàng, giao hàng, hàng hoàn, đối soát vận hành, bảo mật, API, database, DevOps, support và governance.

Phase 2 chuyển trọng tâm sang 4 dòng giá trị cao hơn:

```text
Khách hàng  →  Con người  →  Tăng trưởng  →  Lãi thật
CRM         →  HRM         →  KOL/Affiliate → Finance/BI
```

Mục tiêu Phase 2 không phải chỉ thêm chức năng cho “đủ ERP”. Mục tiêu là biến hệ thống từ **máy vận hành** thành **máy tăng trưởng có kiểm soát**:

- biết khách hàng là ai, đã mua gì, sắp cần gì;
- biết nhân sự nào tạo ra doanh thu/năng suất, nhân sự nào tạo rủi ro;
- biết KOL/KOC nào ra đơn thật, KOL nào chỉ tạo tiếng ồn;
- biết từng kênh, từng sản phẩm, từng campaign, từng nhân sự bán hàng đang lời hay lỗ;
- biến dữ liệu vận hành Phase 1 thành quyết định kinh doanh hàng ngày.

Một câu chốt:

> Phase 1 giữ cho công ty không rối. Phase 2 giúp công ty tăng trưởng có lợi nhuận.

---

## 2. Điều kiện tiên quyết trước khi triển khai Phase 2

Không nên triển khai Phase 2 nếu Phase 1 chưa đạt các điều kiện sau:

| Điều kiện | Trạng thái bắt buộc |
|---|---|
| Master data SKU/NVL/Kho/Khách/NCC ổn định | Đã chuẩn hóa, hạn chế trùng lặp |
| Stock ledger | Chạy ổn, không cho sửa tồn trực tiếp |
| Batch/QC/expiry | Đã có trạng thái rõ: hold/pass/fail, hạn dùng, batch trace |
| Sales order & shipment | Đơn hàng, pick/pack, bàn giao ĐVVC chạy được |
| Returns | Hàng hoàn có quy trình nhận, kiểm, phân loại |
| Subcontract manufacturing | Gia công ngoài có PO/lệnh, chuyển NVL/bao bì, nhận hàng, QC |
| Security/RBAC/audit | Quyền, audit log, hành động nhạy cảm đã khóa |
| Data governance | Có owner dữ liệu, change control, support model |

Nếu Phase 1 còn sai tồn kho, sai batch, sai đơn hàng, Phase 2 sẽ chỉ khuếch đại lỗi. CRM, KOL, HRM, Finance nâng cao đều cần dữ liệu sạch từ Phase 1.

---

## 3. Phạm vi Phase 2 tổng thể

Phase 2 gồm 6 cụm lớn:

```text
1. CRM / CSKH / Customer 360
2. KOL / KOC / Affiliate / Campaign Management
3. HRM / Payroll / KPI / Commission
4. Finance nâng cao / Profitability / Closing
5. BI / Executive Analytics
6. Loyalty / Marketing Automation cơ bản
```

Trong đó 4 cụm bắt buộc là:

- CRM
- HRM
- KOL/Affiliate
- Finance nâng cao

BI và Loyalty nên làm song song nhưng có thể chia theo sprint sau.

---

## 4. Nguyên tắc thiết kế Phase 2

### 4.1. Không phá lõi Phase 1

Phase 2 chỉ được mở rộng trên nền Phase 1, không được viết đè logic kho, đơn hàng, stock ledger, QC, shipment, hàng hoàn.

Ví dụ:

- KOL gửi sample phải đi qua inventory movement.
- CRM complaint phải trace được về order/batch nếu có.
- Commission phải lấy doanh thu đã giao/đã thu/đã trừ hoàn theo rule, không tự tính bằng file ngoài.
- Finance P&L phải lấy giá vốn/tồn kho/chi phí từ dữ liệu hệ thống, không nhập tay vô tội vạ.

### 4.2. Tăng trưởng phải có kiểm soát

Không đo KOL bằng lượt xem đơn thuần. Không đo sales bằng doanh thu thô. Không đo nhân sự bằng có mặt hay không có mặt. Phase 2 phải đo được:

```text
Hiệu quả thật = Doanh thu thật - Chi phí thật - Rủi ro thật
```

### 4.3. Dữ liệu phải truy ngược được

Mọi số tổng phải drill-down được về chứng từ gốc:

```text
KOL Payout → Campaign → Tracking Code → Order → Shipment → Return → Payment
CRM Complaint → Customer → Order → Batch → QC Record
Sales Commission → Sales Order → Delivery → Collection → Return
Payroll KPI → Attendance → Shift → Department → Output
```

### 4.4. Quyền dữ liệu phải cực chặt

Phase 2 có dữ liệu nhạy cảm hơn Phase 1:

- lương thưởng;
- KPI cá nhân;
- payout KOL;
- lợi nhuận theo kênh;
- giá vốn;
- khách VIP;
- complaint nghiêm trọng;
- hiệu suất nhân viên.

Không được để “ai cũng xem được hết”.

---

## 5. Module CRM / CSKH / Customer 360

### 5.1. Mục tiêu

CRM là nơi gom toàn bộ dữ liệu khách hàng, hành vi mua, chăm sóc, khiếu nại, đổi trả, loyalty và nhắc mua lại.

CRM không chỉ là danh bạ. CRM phải giúp công ty trả lời:

- khách này là ai;
- đã mua gì;
- mua qua kênh nào;
- có từng khiếu nại không;
- có thuộc batch lỗi không;
- bao lâu chưa mua lại;
- có nên upsell/cross-sell gì;
- có phải khách VIP không;
- nhân viên nào đang chăm.

### 5.2. Phạm vi chức năng

#### 5.2.1. Customer 360

Quản lý:

- hồ sơ khách hàng;
- số điện thoại/email/social handle;
- địa chỉ giao hàng;
- kênh phát sinh khách;
- lịch sử mua hàng;
- lịch sử đổi trả;
- ticket CSKH;
- complaint;
- loyalty point;
- segment;
- ghi chú nhu cầu/loại da/sở thích;
- consent marketing.

#### 5.2.2. CSKH Ticket

Quản lý ticket:

- hỏi thông tin sản phẩm;
- hỗ trợ đơn hàng;
- đổi trả;
- khiếu nại chất lượng;
- phản ứng da/kích ứng;
- giao hàng trễ;
- bảo hành/đổi quà;
- escalation lên QA/Finance/Kho.

#### 5.2.3. Complaint linked to batch

Với ngành mỹ phẩm, complaint phải có khả năng liên kết:

```text
Customer → Order → SKU → Batch → QC Record → Return/Investigation/CAPA
```

Nếu nhiều complaint cùng batch, hệ thống phải cảnh báo QA/COO.

#### 5.2.4. Repeat purchase trigger

Cấu hình nhắc mua lại theo sản phẩm:

| Loại sản phẩm | Chu kỳ nhắc gợi ý |
|---|---:|
| Serum 30ml | 35–50 ngày |
| Toner | 45–60 ngày |
| Kem chống nắng | 25–40 ngày |
| Sữa rửa mặt | 45–75 ngày |
| Combo routine | Theo rule riêng |

Trigger phải tạo task cho CSKH hoặc campaign automation.

#### 5.2.5. Customer segmentation

Phân nhóm khách:

- khách mới;
- khách mua lại;
- khách VIP;
- khách có complaint;
- khách mua theo KOL;
- khách mua theo sàn;
- khách có nguy cơ rời bỏ;
- khách từng hoàn hàng;
- khách có giá trị cao nhưng tần suất thấp.

### 5.3. Màn hình chính

| Mã màn hình | Tên màn hình | Mục đích |
|---|---|---|
| CRM-01 | Customer List | Tra cứu, lọc khách, segment |
| CRM-02 | Customer 360 Detail | Hồ sơ khách, đơn, ticket, loyalty |
| CRM-03 | Ticket Inbox | Quản lý ticket CSKH |
| CRM-04 | Ticket Detail | Xử lý ticket, ghi chú, assign, SLA |
| CRM-05 | Complaint Investigation | Khiếu nại chất lượng liên kết batch |
| CRM-06 | Repeat Purchase Tasks | Danh sách khách cần chăm lại |
| CRM-07 | Segmentation Builder | Tạo segment khách |
| CRM-08 | Customer Merge | Xử lý khách trùng |
| CRM-09 | Loyalty Wallet | Điểm, voucher, cấp thành viên |

### 5.4. Trạng thái ticket

```text
New → Assigned → In Progress → Waiting Customer → Waiting Internal → Resolved → Closed
                                 ↓
                              Escalated
```

Complaint chất lượng có trạng thái riêng:

```text
Logged → Linked to Order/Batch → QA Review → Investigation → Action Required → Resolved → Closed
```

### 5.5. Dữ liệu chính

```text
customers
customer_contacts
customer_addresses
customer_segments
customer_notes
crm_tickets
crm_ticket_comments
crm_complaints
crm_complaint_batch_links
crm_tasks
loyalty_accounts
loyalty_transactions
```

### 5.6. KPI CRM

| KPI | Ý nghĩa |
|---|---|
| Repeat Purchase Rate | Tỷ lệ khách mua lại |
| Average Resolution Time | Thời gian xử lý ticket trung bình |
| Complaint Rate by SKU/Batch | Tỷ lệ complaint theo SKU/batch |
| VIP Revenue Contribution | Doanh thu từ khách VIP |
| Customer Lifetime Value | Giá trị vòng đời khách hàng |
| Churn Risk Segment Count | Số khách có nguy cơ rời bỏ |
| CSKH SLA Breach | Ticket quá hạn SLA |

---

## 6. Module KOL / KOC / Affiliate / Campaign

### 6.1. Mục tiêu

Module này biến KOL/KOC/Affiliate từ hoạt động marketing rời rạc thành một hệ thống có dữ liệu, kiểm soát, attribution và payout rõ ràng.

Câu hỏi cần trả lời:

- KOL nào đang hợp tác;
- chi phí booking bao nhiêu;
- gửi sample gì;
- content đã được duyệt chưa;
- claim có sai không;
- mã/link nào phát sinh đơn;
- doanh thu thật sau hoàn/hủy là bao nhiêu;
- payout bao nhiêu;
- campaign có lãi không.

### 6.2. Phạm vi chức năng

#### 6.2.1. KOL/KOC/Affiliate Profile

Quản lý:

- tên/brand name;
- phân loại: KOL, KOC, Affiliate, Creator, Reviewer;
- nền tảng: TikTok, Facebook, Instagram, YouTube, Website, Community;
- niche: skincare, makeup, mom & baby, lifestyle, spa, clinic;
- follower/subscriber;
- engagement rate;
- khu vực;
- tệp khách;
- rate card;
- lịch sử hợp tác;
- blacklist/risk note;
- thông tin thanh toán.

#### 6.2.2. Campaign Management

Quản lý:

- campaign brief;
- brand/SKU liên quan;
- mục tiêu: awareness, sales, launch, clearance, repeat purchase;
- ngân sách;
- danh sách KOL;
- timeline;
- deliverables;
- sample/gifting;
- tracking code/link;
- doanh thu;
- chi phí;
- payout.

#### 6.2.3. Claim & Content Approval

Với mỹ phẩm, KOL có thể tạo rủi ro lớn nếu nói sai claim. Cần có:

- claim library;
- từ nhạy cảm/từ cấm;
- upload script/video/caption;
- duyệt bởi Brand/QA nếu cần;
- ghi log version content;
- lưu bằng chứng bài đã đăng.

#### 6.2.4. Sample/Gifting Control

Sample/gift phải đi qua inventory:

```text
Campaign → Sample Request → Approval → Inventory Issue → KOL Received → Content/Review
```

Không được xuất sample ngoài hệ thống. Đây là điểm chống thất thoát.

#### 6.2.5. Attribution & Payout

Cần tracking qua:

- coupon code;
- tracking link;
- affiliate code;
- order source;
- manual attribution có duyệt.

Payout nên dựa trên doanh thu hợp lệ:

```text
Eligible Revenue = Delivered Revenue - Cancelled - Returned - Invalid Orders
```

Công thức lãi thật:

```text
Campaign Net Profit
= Net Revenue
- COGS
- Discount Cost
- Gift/Sample Cost
- Shipping Subsidy
- Platform Fee
- Booking Fee
- Affiliate/KOL Commission
- Other Campaign Cost
```

### 6.3. Màn hình chính

| Mã | Màn hình | Mục đích |
|---|---|---|
| KOL-01 | KOL/KOC List | Quản lý danh sách KOL |
| KOL-02 | KOL Profile | Hồ sơ, rate, lịch sử, risk |
| KOL-03 | Campaign List | Danh sách campaign |
| KOL-04 | Campaign Detail | Brief, KOL, sample, content, KPI |
| KOL-05 | Sample/Gifting Request | Xin xuất sample/gift |
| KOL-06 | Content Approval | Duyệt caption/script/video |
| KOL-07 | Tracking Code Manager | Coupon/link/affiliate code |
| KOL-08 | Attribution Dashboard | Đơn/doanh thu theo KOL |
| KOL-09 | Payout Statement | Đối soát và duyệt payout |
| KOL-10 | Campaign Profitability | Lãi/lỗ chiến dịch |

### 6.4. Trạng thái campaign

```text
Draft → Submitted → Approved → Active → Completed → Reconciled → Closed
              ↓
          Rejected / Cancelled
```

Content approval:

```text
Draft → Submitted → Brand Review → QA Review if needed → Approved → Posted → Evidence Uploaded
```

Payout:

```text
Calculated → Pending Review → Approved → Paid → Closed
```

### 6.5. KPI KOL/Affiliate

| KPI | Ý nghĩa |
|---|---|
| Revenue by KOL | Doanh thu theo KOL |
| Valid Orders | Đơn hợp lệ sau hủy/hoàn |
| Return Rate by KOL | Tỷ lệ hoàn theo KOL |
| ROAS | Doanh thu / chi phí campaign |
| Net Profit by Campaign | Lãi thật campaign |
| Cost per Valid Order | Chi phí trên đơn hợp lệ |
| Content On-time Rate | Tỷ lệ đăng đúng hạn |
| Claim Violation Count | Số lỗi claim |
| Sample-to-Revenue Ratio | Hiệu quả sample/gifting |

---

## 7. Module HRM / Payroll / KPI / Commission

### 7.1. Mục tiêu

HRM Phase 2 không chỉ lưu hồ sơ nhân viên. Nó phải kết nối con người với vận hành, năng suất, lương thưởng và trách nhiệm.

Câu hỏi cần trả lời:

- ai đang làm ở bộ phận nào;
- ai được quyền thao tác nghiệp vụ nào;
- ai đi làm/đi trễ/nghỉ phép/OT;
- công nhân/tổ/ca nào năng suất ra sao;
- sales nào tạo doanh thu và thu tiền thật;
- CSKH nào xử lý ticket tốt;
- KOL manager nào chạy campaign có lãi;
- payroll tính dựa trên dữ liệu nào.

### 7.2. Phạm vi chức năng

#### 7.2.1. Employee Master

Quản lý:

- hồ sơ nhân viên;
- phòng ban;
- chức danh;
- quản lý trực tiếp;
- loại hợp đồng;
- ngày vào/nghỉ;
- thông tin lương cơ bản;
- tài khoản hệ thống;
- quyền nghiệp vụ liên quan;
- hồ sơ đào tạo.

#### 7.2.2. Attendance / Shift / Leave / OT

Quản lý:

- ca làm;
- chấm công;
- đi trễ/về sớm;
- nghỉ phép;
- nghỉ không lương;
- tăng ca;
- phê duyệt OT;
- đối soát công cuối kỳ.

#### 7.2.3. Payroll Basic

Tính lương cơ bản theo:

- ngày công;
- ca;
- OT;
- phụ cấp;
- phạt/khấu trừ;
- thưởng;
- commission;
- tạm ứng;
- bảo hiểm/thuế nếu triển khai.

#### 7.2.4. KPI & Performance

KPI theo nhóm:

| Nhóm | KPI ví dụ |
|---|---|
| Kho | số đơn pick/pack, lỗi đóng hàng, lệch tồn, đúng giờ bàn giao |
| CSKH | SLA ticket, resolution time, complaint handling |
| Sales | doanh thu hợp lệ, thu tiền, tỷ lệ hoàn, khách mới/mua lại |
| KOL Manager | campaign profit, content on-time, claim compliance |
| Production/Gia công coordinator | đúng hạn, chất lượng, claim nhà máy, lead time |
| HR | tuyển dụng, training, payroll accuracy |

#### 7.2.5. Commission Engine

Commission không nên chỉ dựa trên doanh thu thô. Nên hỗ trợ rule:

```text
Commission Eligible Revenue
= Delivered/Collected Revenue - Cancelled - Returned - Invalid Discount
```

Có thể áp dụng theo:

- nhân viên sales;
- cửa hàng;
- team;
- KOL manager;
- affiliate;
- CSKH upsell;
- sản phẩm/SKU;
- kênh bán.

### 7.3. Màn hình chính

| Mã | Màn hình | Mục đích |
|---|---|---|
| HR-01 | Employee List | Danh sách nhân viên |
| HR-02 | Employee Profile | Hồ sơ, hợp đồng, phòng ban |
| HR-03 | Org Chart | Sơ đồ tổ chức |
| HR-04 | Shift Planning | Phân ca |
| HR-05 | Attendance Timesheet | Chấm công |
| HR-06 | Leave Request | Nghỉ phép |
| HR-07 | OT Request | Tăng ca |
| HR-08 | Payroll Period | Kỳ lương |
| HR-09 | Payroll Detail | Lương từng người |
| HR-10 | KPI Dashboard | Hiệu suất phòng ban/cá nhân |
| HR-11 | Commission Statement | Hoa hồng/thưởng |
| HR-12 | Training Record | Đào tạo, chứng nhận thao tác |

### 7.4. Trạng thái payroll

```text
Draft → HR Review → Finance Review → Management Approved → Paid → Closed
```

Payroll đã approved không được sửa trực tiếp. Nếu sai phải tạo adjustment.

### 7.5. KPI HRM

| KPI | Ý nghĩa |
|---|---|
| Attendance Rate | Tỷ lệ đi làm |
| OT Hours by Department | OT theo bộ phận |
| Payroll Accuracy | Tỷ lệ bảng lương không lỗi |
| Turnover Rate | Tỷ lệ nghỉ việc |
| Training Completion | Tỷ lệ hoàn tất đào tạo |
| Sales Commission Accuracy | Độ chính xác hoa hồng |
| Productivity per Headcount | Doanh thu/năng suất trên đầu người |

---

## 8. Finance nâng cao / Profitability / Closing

### 8.1. Mục tiêu

Finance Phase 1 đã có thu/chi/công nợ cơ bản. Phase 2 cần đi sâu vào:

- giá vốn rõ hơn;
- lợi nhuận theo SKU/kênh/campaign/KOL;
- ngân sách;
- phân bổ chi phí;
- đối soát COD/payment;
- payout KOL/commission;
- closing tháng;
- báo cáo quản trị.

Mục tiêu không phải thay thế toàn bộ phần mềm kế toán thuế ngay lập tức, mà là xây **finance management layer** cho CEO/COO/CFO nhìn lãi thật.

### 8.2. Phạm vi chức năng

#### 8.2.1. AR/AP nâng cao

- công nợ khách hàng theo đơn/kênh;
- công nợ nhà cung cấp;
- công nợ nhà máy gia công;
- aging report;
- credit limit;
- payment matching;
- partial payment;
- write-off có duyệt.

#### 8.2.2. COD / Payment Reconciliation

- import statement từ ĐVVC/cổng thanh toán/ngân hàng;
- match với đơn hàng;
- phát hiện thiếu tiền, lệch phí, đơn hoàn;
- tạo reconciliation batch;
- approval trước khi close.

#### 8.2.3. Cost & Margin

Tính lợi nhuận theo:

- SKU;
- batch;
- channel;
- store;
- customer group;
- KOL/campaign;
- sales person;
- brand.

Công thức gợi ý:

```text
Gross Margin = Net Revenue - COGS
Contribution Margin = Gross Margin - Channel Fee - Shipping Subsidy - Discount Cost - Campaign Cost
```

#### 8.2.4. Budget & Cost Center

- budget theo department/campaign/channel;
- cost center;
- expense request;
- expense approval;
- actual vs budget;
- over-budget alert.

#### 8.2.5. Month-end Closing

Các bước closing:

```text
1. Lock sales period
2. Complete shipment/COD reconciliation
3. Complete return adjustment
4. Complete stock reconciliation
5. Confirm COGS
6. Confirm campaign/KOL cost
7. Confirm payroll/commission
8. Generate management P&L
9. Management approval
10. Close period
```

Sau khi close, không sửa trực tiếp dữ liệu kỳ cũ. Chỉ được adjustment có duyệt.

### 8.3. Màn hình chính

| Mã | Màn hình | Mục đích |
|---|---|---|
| FIN-01 | AR Dashboard | Phải thu, aging, credit |
| FIN-02 | AP Dashboard | Phải trả NCC/nhà máy |
| FIN-03 | Payment Matching | Match thanh toán với đơn |
| FIN-04 | COD Reconciliation | Đối soát COD |
| FIN-05 | Expense Request | Đề nghị chi phí |
| FIN-06 | Budget Control | Ngân sách và actual |
| FIN-07 | COGS Review | Kiểm tra giá vốn |
| FIN-08 | Profitability Dashboard | Lãi theo SKU/kênh/KOL |
| FIN-09 | Payout Center | KOL/affiliate/sales payout |
| FIN-10 | Month-end Closing | Closing tháng |

### 8.4. KPI Finance

| KPI | Ý nghĩa |
|---|---|
| Gross Margin by SKU | Lãi gộp theo SKU |
| Contribution Margin by Channel | Lãi sau phí kênh |
| Campaign Net Profit | Lãi thật chiến dịch |
| AR Aging | Tuổi nợ phải thu |
| AP Aging | Tuổi nợ phải trả |
| COD Reconciliation Variance | Lệch đối soát COD |
| Inventory Value | Giá trị tồn kho |
| Expired/Near-expiry Value | Giá trị hàng cận/hết date |
| Budget Utilization | Mức sử dụng ngân sách |
| Month-end Close Duration | Thời gian đóng kỳ |

---

## 9. BI / Executive Analytics

### 9.1. Mục tiêu

BI Phase 2 gom dữ liệu từ Phase 1 và Phase 2 để tạo một lớp nhìn quản trị.

CEO không cần mở 20 màn hình. CEO cần thấy:

```text
Hôm nay công ty có lời thật không?
Tiền đang kẹt ở đâu?
Hàng đang kẹt ở đâu?
Kênh nào đang đốt tiền?
KOL nào đáng giữ?
Khách nào đáng chăm?
Đội nào đang nghẽn?
```

### 9.2. Dashboard chính

| Dashboard | Nội dung |
|---|---|
| Executive Cockpit | Doanh thu, lợi nhuận, tồn kho, cash, cảnh báo |
| Sales & Channel BI | Kênh bán, SKU, khách, cửa hàng |
| Inventory Health BI | Tồn, cận date, slow-moving, batch risk |
| CRM BI | khách mới/cũ, repeat, complaint, loyalty |
| KOL BI | KOL/campaign ROAS, profit, return rate |
| HR BI | headcount, attendance, OT, productivity |
| Finance BI | margin, AR/AP, budget, closing |
| Operations BI | pick/pack, handover, return, subcontract lead time |

### 9.3. Cảnh báo quản trị

- batch có complaint tăng bất thường;
- KOL có return rate cao;
- campaign âm lợi nhuận;
- khách VIP lâu chưa mua lại;
- nhân viên có OT bất thường;
- AR quá hạn;
- stock cận date giá trị cao;
- cost center vượt ngân sách;
- đơn giao trễ tăng theo carrier.

---

## 10. Loyalty / Membership / Marketing Automation cơ bản

### 10.1. Mục tiêu

Loyalty giúp giữ khách và tăng mua lại. Không nên làm quá phức tạp ở đầu Phase 2. Chỉ cần đủ:

- điểm thưởng;
- cấp thành viên;
- voucher;
- rule tích/tiêu;
- nhắc mua lại;
- campaign cơ bản.

### 10.2. Chức năng

- customer tier: Normal, Silver, Gold, VIP;
- earn points by order;
- redeem voucher;
- birthday voucher;
- repeat purchase trigger;
- win-back campaign;
- segment-based campaign list;
- export/import campaign contact list;
- consent check.

### 10.3. Rule mẫu

```text
Earn point only when order status = Delivered/Closed
If order returned, reverse loyalty point
Voucher cannot be stacked beyond promotion rule
VIP tier recalculated monthly
```

---

## 11. Luồng nghiệp vụ Phase 2

### 11.1. CRM Complaint to Batch

```text
Khách phản ánh
→ CSKH tạo ticket
→ link order/SKU/batch
→ nếu nghi ngờ chất lượng: escalate QA
→ QA kiểm batch/QC record
→ nếu cần: hold batch / CAPA / recall decision
→ CSKH phản hồi khách
→ đóng ticket
→ BI cập nhật complaint rate
```

### 11.2. Repeat Purchase Flow

```text
Order delivered/closed
→ hệ thống ghi lịch sử mua
→ tính chu kỳ nhắc theo SKU
→ tạo CRM task khi đến hạn
→ CSKH gọi/nhắn/tạo offer
→ khách mua lại
→ đo repeat conversion
```

### 11.3. KOL Campaign to Revenue

```text
Tạo campaign
→ chọn KOL
→ duyệt brief/claim
→ xin xuất sample/gift
→ inventory xuất sample
→ KOL đăng content
→ tracking code/link phát sinh đơn
→ đơn giao/hoàn/huỷ
→ tính eligible revenue
→ tính payout
→ Finance duyệt thanh toán
→ Campaign profitability
```

### 11.4. HR Attendance to Payroll

```text
Phân ca
→ chấm công
→ OT/nghỉ phép
→ HR review timesheet
→ tính payroll draft
→ tính commission/KPI nếu có
→ Finance review
→ Management approve
→ Paid
→ close payroll period
```

### 11.5. Finance Month-end Closing

```text
Đóng đơn bán kỳ
→ đối soát COD/payment
→ xử lý return adjustment
→ đối soát tồn kho
→ xác nhận COGS
→ xác nhận chi phí campaign/KOL
→ xác nhận payroll/commission
→ tạo management P&L
→ duyệt đóng kỳ
```

---

## 12. Data Model bổ sung Phase 2

### 12.1. CRM

```text
customers
customer_contacts
customer_addresses
customer_segments
customer_segment_members
crm_tickets
crm_ticket_comments
crm_tasks
crm_complaints
crm_complaint_batch_links
loyalty_accounts
loyalty_transactions
vouchers
voucher_redemptions
```

### 12.2. KOL/Affiliate

```text
kol_profiles
kol_platform_accounts
kol_rate_cards
campaigns
campaign_kols
campaign_deliverables
campaign_content_submissions
campaign_tracking_codes
affiliate_orders
kol_sample_requests
kol_payout_statements
kol_payout_lines
campaign_costs
```

### 12.3. HRM

```text
employees
departments
positions
employee_contracts
shifts
shift_assignments
attendance_records
leave_requests
ot_requests
payroll_periods
payroll_lines
commission_rules
commission_statements
training_records
employee_assets
```

### 12.4. Finance Advanced

```text
cost_centers
budgets
expense_requests
payment_statements
payment_reconciliation_batches
payment_reconciliation_lines
finance_adjustments
profitability_snapshots
closing_periods
closing_tasks
```

### 12.5. BI

```text
bi_daily_sales_summary
bi_channel_margin_summary
bi_inventory_health_summary
bi_kol_campaign_summary
bi_crm_customer_summary
bi_hr_productivity_summary
bi_finance_monthly_summary
```

---

## 13. Permission / Approval Phase 2

### 13.1. Role mới

| Role | Mô tả |
|---|---|
| CRM Agent | Xử lý ticket, chăm khách |
| CRM Manager | Quản lý CSKH, segment, complaint escalation |
| KOL Manager | Tạo campaign, quản KOL, sample request |
| Brand/Marketing Manager | Duyệt campaign/content |
| QA Reviewer | Duyệt claim nhạy cảm, xử lý complaint batch |
| HR Staff | Quản hồ sơ, chấm công, leave, payroll draft |
| HR Manager | Duyệt HR, payroll review |
| Finance Analyst | Đối soát, margin, payout, closing |
| CFO/Finance Manager | Duyệt payout, closing, budget |
| CEO/Management | Xem BI, duyệt chi phí/vượt ngân sách/lương thưởng |

### 13.2. Approval quan trọng

| Nghiệp vụ | Người tạo | Duyệt |
|---|---|---|
| Campaign vượt ngân sách | KOL Manager | Marketing Manager + Finance/CEO |
| Xuất sample/gift | KOL Manager | Marketing Manager + Inventory Manager nếu vượt ngưỡng |
| Content claim nhạy cảm | KOL Manager | Brand + QA |
| Payout KOL | System/Finance | Marketing Manager + Finance |
| Payroll | HR | HR Manager + Finance + CEO |
| Commission | System/Sales Ops | Sales Manager + Finance |
| Budget override | Department Head | Finance + CEO |
| Period closing | Finance | CFO/CEO |
| Complaint nghiêm trọng | CSKH/QA | QA Manager + COO/CEO |

### 13.3. Field-level restriction

| Dữ liệu | Ai được xem |
|---|---|
| Giá vốn | Finance, CEO, authorized roles |
| Lương | HR Manager, Finance, CEO |
| Payout KOL | Marketing Manager, Finance, CEO |
| Profitability | Management, Finance |
| Complaint nghiêm trọng | CSKH Manager, QA, Management |
| Blacklist KOL | Marketing Manager, Legal/CEO nếu có |

---

## 14. API / Integration Phase 2

### 14.1. API nhóm CRM

```text
GET    /api/v1/customers
POST   /api/v1/customers
GET    /api/v1/customers/{id}
POST   /api/v1/crm/tickets
POST   /api/v1/crm/tickets/{id}/assign
POST   /api/v1/crm/tickets/{id}/resolve
POST   /api/v1/crm/complaints/{id}/link-batch
GET    /api/v1/crm/tasks/repeat-purchase
```

### 14.2. API nhóm KOL

```text
GET    /api/v1/kol/profiles
POST   /api/v1/kol/profiles
POST   /api/v1/campaigns
POST   /api/v1/campaigns/{id}/submit
POST   /api/v1/campaigns/{id}/approve
POST   /api/v1/campaigns/{id}/sample-requests
POST   /api/v1/campaigns/{id}/content-submissions
POST   /api/v1/campaigns/{id}/tracking-codes
GET    /api/v1/campaigns/{id}/attribution
POST   /api/v1/kol/payouts/{id}/approve
```

### 14.3. API nhóm HRM

```text
GET    /api/v1/hr/employees
POST   /api/v1/hr/employees
POST   /api/v1/hr/attendance/import
POST   /api/v1/hr/leave-requests
POST   /api/v1/hr/ot-requests
POST   /api/v1/hr/payroll-periods
POST   /api/v1/hr/payroll-periods/{id}/submit
POST   /api/v1/hr/payroll-periods/{id}/approve
```

### 14.4. API nhóm Finance

```text
POST   /api/v1/finance/payment-statements/import
POST   /api/v1/finance/reconciliation-batches
POST   /api/v1/finance/reconciliation-batches/{id}/approve
GET    /api/v1/finance/profitability/channel
GET    /api/v1/finance/profitability/campaign
POST   /api/v1/finance/closing-periods
POST   /api/v1/finance/closing-periods/{id}/close
```

---

## 15. UI/UX Phase 2

### 15.1. Nguyên tắc UX

- CRM phải ưu tiên tốc độ tra cứu khách.
- KOL module phải hiển thị được trạng thái campaign, sample, content, payout trong một màn hình.
- HRM phải tránh lộ lương với người không có quyền.
- Finance dashboard phải drill-down về chứng từ.
- BI không chỉ chart đẹp, phải có cảnh báo hành động.

### 15.2. Portal theo vai trò

| Portal | Người dùng |
|---|---|
| CRM Portal | CSKH, CRM Manager |
| KOL/Marketing Portal | KOL Manager, Brand, Marketing |
| HR Portal | HR, nhân viên, trưởng bộ phận |
| Finance Portal | Finance, CFO, CEO |
| Executive BI Portal | CEO/COO/CFO |

### 15.3. Component mới cần có

- Customer Timeline
- Ticket SLA Badge
- Complaint Batch Link Card
- Campaign Kanban
- KOL Scorecard
- Content Approval Viewer
- Payout Statement Table
- Timesheet Grid
- Payroll Approval Panel
- Profitability Drill-down Table
- BI Alert Card

---

## 16. Sprint Plan đề xuất cho Phase 2

Phase 2 có thể chia 10–12 sprint. Không nên làm một cục.

### Sprint 1–2: CRM Core

- Customer 360
- Ticket inbox
- Customer merge
- Repeat purchase task cơ bản
- Complaint link order/batch

### Sprint 3–4: KOL Core

- KOL profile
- Campaign planner
- Sample/gifting request
- Content approval
- Tracking code

### Sprint 5: Attribution & KOL Payout

- Order attribution
- Revenue after return/cancel
- Payout statement
- Campaign profitability basic

### Sprint 6–7: HRM Core

- Employee master
- Department/position/org chart
- Shift/attendance/leave/OT
- Training record

### Sprint 8: Payroll & Commission Basic

- Payroll period
- Timesheet review
- Commission rules
- Payroll approval

### Sprint 9–10: Finance Advanced

- AR/AP advanced
- COD/payment reconciliation
- Budget/cost center
- Profitability by SKU/channel/KOL
- Month-end closing basic

### Sprint 11–12: BI & Loyalty

- Executive BI
- CRM BI
- KOL BI
- HR BI
- Finance BI
- Loyalty point/voucher basic

---

## 17. UAT Scenario Phase 2

### CRM

| Case | Kết quả mong muốn |
|---|---|
| Tạo ticket khách phản ánh sản phẩm | Ticket tạo đúng, assign đúng |
| Link complaint với order/batch | QA thấy được batch liên quan |
| Khách mua lại sau repeat task | CRM ghi nhận conversion |
| Merge 2 khách trùng | Lịch sử đơn/ticket không mất |

### KOL

| Case | Kết quả mong muốn |
|---|---|
| Tạo campaign và xin sample | Có approval, xuất kho đúng movement |
| Duyệt content có claim nhạy cảm | QA/Brand phải duyệt trước khi approved |
| Đơn qua mã KOL bị hoàn | Không tính payout hoặc reverse payout |
| Campaign hoàn tất | Ra được net profit |

### HRM

| Case | Kết quả mong muốn |
|---|---|
| Import attendance | Timesheet đúng |
| Xin nghỉ phép | Đi đúng approval flow |
| Tính payroll có OT | OT vào đúng bảng lương |
| Payroll approved | Không sửa trực tiếp được |

### Finance

| Case | Kết quả mong muốn |
|---|---|
| Import COD statement | Match được đơn, phát hiện lệch |
| Tính margin campaign | Trừ đúng COGS, discount, return, payout |
| Close tháng | Lock dữ liệu kỳ, cho adjustment có duyệt |
| Budget vượt ngưỡng | Bắt duyệt Finance/CEO |

---

## 18. Risk Log Phase 2

| Rủi ro | Ảnh hưởng | Cách kiểm soát |
|---|---|---|
| CRM data trùng khách | Sai segment, sai CLV | Customer merge, phone/email unique rule |
| KOL attribution sai | Payout sai, mất tiền | Tracking code, delivered revenue rule |
| KOL claim sai | Rủi ro thương hiệu/pháp lý | Content approval + claim library |
| Sample thất thoát | Mất hàng, sai chi phí campaign | Sample request linked inventory |
| Payroll sai | Mất niềm tin nhân sự | Timesheet review + approval + adjustment |
| Commission sai | Tranh chấp nội bộ | Rule engine + statement review |
| Profitability sai | CEO ra quyết định sai | Drill-down, closing, data quality check |
| BI quá nhiều dashboard | Loạn thông tin | KPI catalog rõ, dashboard theo vai trò |
| Phase 2 scope creep | Chậm, quá tải team | Sprint priority + parking lot |

---

## 19. Definition of Done Phase 2

Một module Phase 2 chỉ được coi là xong khi:

- có PRD chi tiết hoặc story + acceptance criteria;
- có permission rule;
- có API contract nếu cần;
- có DB migration;
- có audit log cho hành động nhạy cảm;
- có test case;
- có UAT passed;
- có SOP/training note;
- có dashboard/KPI tối thiểu nếu module sinh dữ liệu quản trị;
- có rollback/correction rule cho dữ liệu sai.

---

## 20. Parking Lot cho Phase 3

Những thứ chưa nên nhồi vào Phase 2 nếu team chưa đủ lực:

- AI personalization;
- full marketing automation đa kênh phức tạp;
- social listening tự động;
- livestream order automation;
- full accounting tax posting;
- advanced demand forecasting;
- KOL fraud detection nâng cao;
- mobile app riêng cho KOL;
- HR recruitment ATS đầy đủ;
- learning management system lớn;
- omnichannel CDP hoàn chỉnh.

Các mục này có giá trị, nhưng nên để Phase 3 sau khi CRM/KOL/HRM/Finance core chạy ổn.

---

## 21. Kết luận chiến lược

Phase 2 là bước biến ERP từ “hệ thống kiểm soát vận hành” thành “hệ thống điều hành tăng trưởng”.

Thứ tự ưu tiên nên là:

```text
1. CRM Core
2. KOL Campaign + Attribution + Payout
3. HRM Core + Payroll/Commission
4. Finance Profitability + Reconciliation + Closing
5. BI + Loyalty
```

Không nên bắt đầu Phase 2 bằng dashboard đẹp. Dashboard chỉ có giá trị khi dữ liệu CRM/KOL/HRM/Finance đã chảy đúng.

Một câu chốt:

> Phase 2 không phải thêm nhiều chức năng. Phase 2 là để biết công ty đang tăng trưởng thật hay chỉ đang bận rộn hơn.

---

## 22. Tài liệu liên quan

Tài liệu này kế thừa và phụ thuộc vào các nhóm tài liệu trước:

```text
01 Blueprint
03 PRD/SRS Phase 1
04 Permission & Approval Matrix
05 Data Dictionary
06 Process Flow To-Be
07 Report/KPI Catalog
08 Screen List/Wireframe
09 UAT
11–18 Technical/Engineering Stack
19 Security/RBAC/Audit
20 As-Is Workflow
21 Gap Analysis
22 Core Docs Revision v1.1
23 Integration Spec
24 QA/Test Strategy
25 Product Backlog/Sprint Plan
26 SOP Training
27 GoLive Runbook
28 Risk/Incident Playbook
29 Operations Support Model
30 Data Governance/Change Control
```

