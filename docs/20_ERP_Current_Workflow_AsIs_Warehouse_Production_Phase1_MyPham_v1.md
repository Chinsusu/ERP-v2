# 20 ERP Current Workflow As-Is — Warehouse & Production Phase 1 Mỹ Phẩm v1

**Tên file:** `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md`  
**Dự án:** Web ERP cho công ty sản xuất, phân phối, bán lẻ mỹ phẩm  
**Phiên bản:** v1.0  
**Trạng thái:** As-Is baseline từ tài liệu workflow thực tế  
**Nguồn đầu vào:**

1. `Công-việc-hằng-ngày.pdf`
2. `Nội-Quy.pdf`
3. `Quy-trình-bàn-giao.pdf`
4. `Quy-trình-sản-xuất.pdf`

---

## 1. Mục tiêu tài liệu

Tài liệu này ghi lại **workflow hiện tại — As-Is** của công ty dựa trên 4 file quy trình đã cung cấp.

Mục tiêu không phải vẽ một ERP lý tưởng, mà là **chụp X-quang cách công ty đang vận hành thật**:

- Kho đang làm gì mỗi ngày?
- Nhập kho đang đi qua bước nào?
- Xuất kho và đóng hàng đang được kiểm soát ra sao?
- Bàn giao cho đơn vị vận chuyển hiện đang kiểm số lượng và quét mã như thế nào?
- Hàng hoàn đang nhận, kiểm tra và phân loại ra sao?
- Sản xuất hiện là xưởng nội bộ hay mô hình gia công ngoài?
- Những bước nào nên giữ, bước nào cần chuẩn hóa, bước nào phải đưa vào ERP?

Tài liệu này sẽ là đầu vào cho file tiếp theo:

```text
21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md
```

Một câu chốt:

> **Không được build ERP chỉ theo tưởng tượng. Phải build theo workflow thật, sau đó nâng workflow thật lên chuẩn cao hơn.**

---

## 2. Executive summary

Sau khi đọc 4 tài liệu, workflow hiện tại có 4 cụm lớn:

### 2.1 Kho vận hành theo ca/ngày

Kho có nhịp vận hành hằng ngày:

```text
Tiếp nhận đơn trong ngày
→ thực hiện xuất/nhập theo bảng nội quy
→ soạn hàng và đóng gói
→ sắp xếp, tối ưu vị trí kho
→ kiểm kê hàng tồn cuối ngày
→ đối soát số liệu và báo cáo cho quản lý
→ kết thúc ca làm
```

Điều này cho thấy ERP phải có **daily warehouse board**, **shift closing**, **cycle count/kiểm kê cuối ngày**, và **đối soát vận hành cuối ca**, không chỉ có màn hình nhập/xuất đơn lẻ.

### 2.2 Nội quy kho chia 4 nhánh nghiệp vụ

Bảng nội quy chia thành:

```text
1. Quy trình nhập kho
2. Quy trình xuất kho
3. Quy trình đóng hàng
4. Quy trình xử lý hàng hoàn
```

Đây là cấu trúc rất tốt để chuyển thành module và use case ERP.

### 2.3 Bàn giao ĐVVC là quy trình kiểm soát riêng

Bàn giao hàng cho đơn vị vận chuyển không phải chỉ đổi trạng thái đơn.

Hiện tại có các bước:

```text
phân khu để hàng
→ để theo thùng/rổ
→ đối chiếu số lượng đơn dựa trên bảng
→ lấy hàng và quét mã trực tiếp
→ nếu đủ thì ký xác nhận bàn giao
→ nếu chưa đủ thì kiểm tra lại mã hoặc tìm lại ở khu đóng hàng
```

ERP phải có **carrier manifest**, **scan verify**, **handover exception**, **khu bàn giao**, và **bằng chứng bàn giao**.

### 2.4 Sản xuất hiện nghiêng về gia công ngoài

Quy trình sản xuất hiện tại cho thấy công ty làm việc với nhà máy:

```text
lên đơn với nhà máy
→ xác nhận số lượng/quy cách/mẫu mã
→ cọc đơn và xác nhận thời gian
→ chuyển nguyên vật liệu/bao bì qua nhà máy
→ làm mẫu/chốt mẫu
→ sản xuất hàng loạt
→ giao hàng về kho
→ kiểm tra số lượng/chất lượng
→ nhận hàng hoặc báo lỗi nhà máy trong 3–7 ngày
→ thanh toán lần cuối
```

Vì vậy Phase 1 không nên thiết kế production như xưởng nội bộ 100%. Cần có nhánh **subcontract manufacturing / gia công ngoài**.

---

## 3. Vai trò quan sát được từ workflow

| Vai trò | Xuất hiện trong workflow | Ghi chú ERP |
|---|---|---|
| Nhân viên kho | Nhập, xuất, soạn, đóng, quét, kiểm kê | Cần role Warehouse Staff/Packer/Handover |
| Trưởng kho | Lưu phiếu nhập/xuất, kiểm soát kho | Cần quyền duyệt/chốt ca/kiểm kê |
| Quản lý | Nhận báo cáo đối soát số liệu | Cần dashboard và shift report |
| Đơn vị vận chuyển / ĐVVC | Nhận bàn giao hàng | Có thể là external party hoặc thông tin manifest |
| Shipper | Trả hàng hoàn về kho | Cần return receiving flow |
| Nhà cung cấp / NCC | Nhận trả hàng nếu nhập không đạt | Liên quan purchase/inbound exception |
| Nhà máy gia công | Nhận đơn, nhận NVL/bao bì, sản xuất/giao hàng | Cần subcontract module |
| QA/QC hoặc người kiểm hàng | Kiểm số lượng/chất lượng, kiểm hàng hoàn | Cần QC/inspection role |
| Finance | Cọc đơn, thanh toán lần cuối | Cần kiểm soát payment milestone |

---

## 4. Chứng từ / bằng chứng hiện tại

Các chứng từ/bằng chứng được nhắc hoặc suy ra từ workflow:

| Chứng từ / bằng chứng | Dùng ở đâu | ERP object đề xuất |
|---|---|---|
| Chứng từ giao hàng | Nhập kho | Goods receipt input / supplier delivery note |
| Phiếu nhập kho | Nhập kho đạt | GRN / Inbound Receipt |
| Phiếu xuất kho | Xuất kho | Stock Issue / Pick Ticket / Delivery Issue |
| Biên bản bàn giao kho | Xuất kho / bàn giao | Handover Record |
| Phiếu đơn hàng hợp lệ | Đóng hàng | Pick/Pack Order |
| Bảng số lượng đơn theo ĐVVC | Bàn giao | Carrier Manifest |
| Mã đơn/mã vận đơn | Quét bàn giao/hàng hoàn | Barcode/Tracking ID |
| Biên bản ĐVVC | Bàn giao | Carrier signed manifest |
| Hình ảnh/video hàng hoàn | Hàng hoàn | Return evidence attachment |
| Biên bản bàn giao NVL/bao bì | Gia công ngoài | Subcontract Material Handover |
| COA/MSDS/mẫu/hóa đơn VAT nếu cần | Gia công ngoài | Attachment to subcontract order/material transfer |
| Mẫu được chốt/lưu mẫu | Gia công ngoài | Sample approval record |
| Báo lỗi nhà máy | Gia công không đạt | Defect/Claim report |
| Thanh toán cọc/cuối | Gia công ngoài | Payment milestone |

---

## 5. As-Is Flow A — Công việc hằng ngày của kho

### 5.1 Mô tả tổng quan

Workflow hằng ngày đang chạy theo chuỗi:

```text
Công việc hằng ngày
→ Tiếp nhận đơn hàng trong ngày
→ Thực hiện quy trình Xuất - Nhập dựa trên BẢNG NỘI QUY
→ Soạn hàng và đóng gói theo quy trình
→ Sắp xếp, tối ưu vị trí kho
→ Kiểm kê hàng tồn kho cuối ngày
→ Đối soát số liệu và báo cáo cho quản lý
→ Kết thúc ca làm
```

### 5.2 Phân tích từng bước

| Bước | Mô tả hiện tại | Input | Output | Điểm kiểm soát |
|---|---|---|---|---|
| 1 | Tiếp nhận đơn hàng trong ngày | Danh sách đơn | Danh sách đơn cần xử lý | Cần phân loại đơn mới/pending/gấp |
| 2 | Thực hiện xuất/nhập theo nội quy | Phiếu nhập/xuất, hàng thực tế | Giao dịch kho | Phải theo bảng nội quy, không làm tùy tiện |
| 3 | Soạn và đóng gói | Phiếu đơn hàng hợp lệ | Đơn đã packed | Cần kiểm SKU, số lượng, khu vực đóng |
| 4 | Sắp xếp/tối ưu vị trí kho | Hàng trong kho | Vị trí kho tối ưu | Cần bin/location logic |
| 5 | Kiểm kê cuối ngày | Tồn thực tế | Kết quả kiểm kê | Cần ghi lệch tồn, không sửa tay |
| 6 | Đối soát và báo cáo quản lý | Đơn/xuất/nhập/tồn | Báo cáo cuối ca | Cần dashboard shift closing |
| 7 | Kết ca | Báo cáo, checklist | Ca được đóng | Sau khi đóng ca, hạn chế sửa |

### 5.3 ERP implications

Bắt buộc có:

- Daily Warehouse Board
- Today Orders Queue
- Pick/Pack Progress
- Inbound/Outbound Tasks
- Location Optimization hoặc ít nhất Location Transfer
- Cycle Count / End-of-Day Count
- Shift Closing Report
- Exceptions: thiếu hàng, lệch tồn, đơn chưa đóng, đơn chưa bàn giao

### 5.4 Rủi ro nếu không đưa vào ERP

- Mỗi ca hiểu một kiểu.
- Không biết cuối ngày còn bao nhiêu đơn pending.
- Tồn thực tế và tồn hệ thống lệch nhưng không biết lệch từ đâu.
- Đơn đã đóng nhưng chưa bàn giao bị thất lạc.
- Quản lý chỉ nhận báo cáo thủ công, khó drill-down chứng từ.

---

## 6. As-Is Flow B — Quy trình nhập kho

### 6.1 Flow hiện tại

```text
Quy trình nhập kho
→ Nhận chứng từ giao hàng
→ Kiểm tra chi tiết: số lượng, bao bì, lô
→ Nếu đạt: sắp xếp hàng vào kho
→ Kiểm tra lần cuối
→ Ký xác nhận và bàn giao lại
→ Trưởng kho lưu phiếu nhập

Nếu không đạt:
→ Trả lại NCC
```

### 6.2 Phân tích

| Bước | Ý nghĩa | ERP object |
|---|---|---|
| Nhận chứng từ giao hàng | Xác định hàng/NCC/giao dịch đầu vào | Inbound Delivery / Supplier Delivery Note |
| Kiểm số lượng/bao bì/lô | Kiểm soát hàng thật | Receiving Inspection |
| Đạt | Cho phép nhập kho | GRN / Inbound Receipt |
| Không đạt | Trả NCC | Inbound Rejection / Return to Supplier |
| Sắp xếp vào kho | Đưa vào vị trí | Putaway Task |
| Kiểm tra lần cuối | Kiểm soát trước khi chốt | Final Receiving Check |
| Ký xác nhận/lưu phiếu | Bằng chứng trách nhiệm | Receipt Sign-off / Attachment |

### 6.3 Dữ liệu bắt buộc

- Nhà cung cấp
- Mã hàng/SKU/NVL/bao bì
- Số lượng chứng từ
- Số lượng thực nhận
- Đơn vị tính
- Batch/lô
- Hạn dùng nếu có
- Tình trạng bao bì
- Kết quả kiểm tra
- Vị trí kho nhận
- Người nhận
- Người kiểm
- File/chứng từ đính kèm

### 6.4 ERP rules đề xuất

- Không tạo tồn khả dụng nếu chưa qua bước kiểm tra phù hợp.
- Nếu hàng cần QC, nhập vào `QC_HOLD` trước.
- Hàng không đạt phải tạo reason code.
- Trả NCC phải có biên bản hoặc chứng từ liên quan.
- Trưởng kho lưu phiếu trong ERP bằng attachment/sign-off, không chỉ giấy.

### 6.5 Điểm cần xác minh

- Nhập kho có dựa trên PO không hay có thể nhập tự do?
- Ai quyết định “Đạt/Không đạt”: kho hay QC?
- Có bắt buộc chụp ảnh hàng không đạt không?
- Có nhập theo barcode không?
- Có vị trí kho/bin location hiện tại không?

---

## 7. As-Is Flow C — Quy trình xuất kho

### 7.1 Flow hiện tại

```text
Quy trình xuất kho
→ Làm phiếu xuất kho
→ Xuất kho hàng
→ Kiểm tra số lượng, đối chiếu lại với số lượng thực tế
→ Ký tên bàn giao
→ Trưởng kho lưu trữ lại phiếu xuất kho
```

### 7.2 Phân tích

| Bước | Ý nghĩa | ERP object |
|---|---|---|
| Làm phiếu xuất | Tạo yêu cầu xuất hàng | Stock Issue / Pick Ticket |
| Xuất hàng | Lấy hàng khỏi kho/vị trí | Pick Task / Stock Movement |
| Kiểm số lượng thực tế | So hệ thống và hàng thật | Pick Verification |
| Ký bàn giao | Xác nhận trách nhiệm | Internal Handover |
| Trưởng kho lưu phiếu | Lưu bằng chứng | Digital Attachment / Audit |

### 7.3 ERP rules đề xuất

- Xuất kho phải dựa trên chứng từ nguồn: Sales Order, Transfer Order, Production/Subcontract Issue, Sample Request.
- Không xuất hàng QC HOLD/FAIL.
- Không xuất vượt tồn khả dụng.
- Nếu xuất theo batch/hạn dùng, phải theo FEFO hoặc batch được chỉ định.
- Mọi xuất kho phải tạo stock ledger movement.
- Phiếu xuất đã confirmed không được sửa số lượng trực tiếp; sai thì tạo correction/reversal.

### 7.4 Điểm cần xác minh

- Có nhiều loại xuất không: xuất bán, xuất sample, xuất cho KOL, xuất gia công, xuất hủy?
- Ai tạo phiếu xuất?
- Ai duyệt phiếu xuất?
- Có cần quét mã khi pick không?
- Có xử lý partial pick không?

---

## 8. As-Is Flow D — Quy trình đóng hàng

### 8.1 Flow hiện tại

```text
Quy trình đóng hàng
→ Nhận phiếu đơn hàng hợp lệ
→ Lọc và phân loại đơn hàng theo ĐVVC / đơn lẻ / đơn sàn
→ Soạn hàng theo từng đơn
→ Đóng gói và kiểm tra tại khu vực đóng hàng: SKU, số lượng, xung quanh
→ Đếm lại tổng số lượng đơn hàng của mỗi sàn
→ Chuyển đến khu vực bàn giao cho ĐVVC
→ Lập, ký xác nhận với bên ĐVVC
```

### 8.2 Phân tích

Đây là flow cực quan trọng cho OMS/WMS vì nó nối bán hàng → kho → giao vận.

| Bước | Ý nghĩa | ERP object |
|---|---|---|
| Nhận phiếu đơn hợp lệ | Chỉ xử lý đơn hợp lệ | Order Ready to Pick |
| Lọc/phân loại | Tối ưu xử lý theo carrier/channel | Picking Batch / Wave Picking |
| Soạn hàng từng đơn | Pick SKU theo đơn | Pick Task |
| Đóng gói/kiểm SKU | Tránh giao sai | Pack Verification |
| Đếm lại theo sàn/ĐVVC | Đối soát trước bàn giao | Manifest Count |
| Chuyển khu bàn giao | Chuyển trạng thái packed | Ready for Handover |
| Ký với ĐVVC | Bằng chứng bàn giao | Carrier Manifest Sign-off |

### 8.3 ERP rules đề xuất

- Đơn phải qua trạng thái `Confirmed/Reserved` trước khi pick.
- Khi pick xong, trạng thái đơn là `Picked`.
- Khi đóng xong và kiểm xong, trạng thái là `Packed`.
- Đơn `Packed` mới được đưa vào manifest ĐVVC.
- Mỗi đơn nên có scan checkpoint: pick, pack, handover.
- Nếu thiếu hàng khi pack, hệ thống tạo exception.

### 8.4 Điểm cần xác minh

- “Đơn sàn” gồm những sàn nào?
- Có phân loại theo ĐVVC trước hay theo sàn trước?
- Có xử lý combo/quà tặng không?
- Có in tem/phiếu giao hàng từ ERP không?
- Có cần cân nặng/kích thước kiện không?

---

## 9. As-Is Flow E — Quy trình xử lý hàng hoàn

### 9.1 Flow hiện tại

```text
Quy trình xử lý hàng hoàn
→ Nhận hàng từ shipper
→ Đưa vào khu vực để hàng hoàn
→ Quét hàng hoàn
→ Quay/ghi lại tình trạng thực tế của sản phẩm từ tất cả các phía sau khi được hoàn về
→ Kiểm tra chi tiết bên trong: tình trạng, tem niêm phong, hư hỏng
→ Nếu còn sử dụng: chuyển vào kho
→ Nếu không sử dụng: chuyển lên Lab
→ Lập phiếu nhập kho / ký tên đầy đủ lưu trữ video + chứng từ
```

### 9.2 Phân tích

Hàng hoàn không đơn giản là “nhập lại kho”. Nó có rủi ro:

- đã mở seal
- hỏng/móp
- sai sản phẩm
- mất phụ kiện/quà tặng
- không xác định được batch
- khách/shipper làm hỏng
- không còn bán được

### 9.3 ERP object đề xuất

| Bước | ERP object |
|---|---|
| Nhận hàng hoàn | Return Receipt |
| Đưa vào khu hàng hoàn | Return Pending Location |
| Quét hàng hoàn | Return Scan Event |
| Ghi tình trạng thực tế | Return Evidence Attachment |
| Kiểm tra chi tiết | Return Inspection |
| Còn sử dụng | Return Restock Movement |
| Không sử dụng | Return To Lab / Damaged Stock Movement |
| Lập phiếu nhập | Return GRN / Return Putaway |

### 9.4 Return disposition chuẩn

```text
PENDING_INSPECTION
REUSABLE
NEED_QC_REVIEW
DAMAGED
TO_LAB
TO_DISPOSAL
WRONG_ITEM
MISSING_ITEM
UNKNOWN
```

### 9.5 ERP rules đề xuất

- Hàng hoàn sau khi scan phải vào trạng thái `PENDING_INSPECTION`.
- Không nhập thẳng vào tồn khả dụng.
- Phải ghi lý do hoàn và tình trạng.
- Hàng còn sử dụng mới được `RETURN_RESTOCK`.
- Hàng không sử dụng chuyển Lab/kho hỏng, không bán được.
- Nếu cần, QC duyệt lại trước khi available.
- File video/ảnh cần liên kết với return record.

### 9.6 Điểm cần xác minh

- Ai quyết định còn sử dụng/không sử dụng?
- Lab xử lý gì sau khi nhận hàng?
- Có hoàn tiền tự động không hay Finance duyệt?
- Có kiểm lại batch/hạn dùng không?
- Có chụp ảnh/video bắt buộc cho mọi hàng hoàn không?

---

## 10. As-Is Flow F — Quy trình bàn giao hàng cho ĐVVC

### 10.1 Flow hiện tại

```text
Quy trình bàn giao hàng cho ĐVVC
→ Phân chia khu vực để hàng
→ Để theo thùng/rổ, mỗi thùng/rổ có số lượng bằng nhau
→ Đối chiếu số lượng đơn dựa trên bảng
→ Lấy hàng và quét mã trực tiếp tại khu vực/hầm bàn giao
→ Nếu đủ đơn: ký xác nhận, bàn giao với ĐVVC
→ Nếu chưa đủ: kiểm tra lại mã
   → Nếu mã chưa có trên hệ thống: thực hiện đóng lại
   → Nếu đã có trên hệ thống: tìm lại trong khu vực đóng hàng
```

### 10.2 Phân tích

Đây là một checkpoint riêng giữa kho và ĐVVC.

Nếu ERP chỉ có trạng thái `shipped` mà không có manifest/scan, rất dễ xảy ra:

- thiếu đơn khi bàn giao
- quét nhầm đơn
- đơn packed nhưng không được giao
- ĐVVC nhận sai số lượng
- không biết lỗi nằm ở packing, handover hay carrier

### 10.3 ERP object đề xuất

| Bước | ERP object |
|---|---|
| Phân khu để hàng | Handover Zone |
| Thùng/rổ | Tote/Container ID |
| Bảng đối chiếu số lượng | Carrier Manifest |
| Quét mã | Handover Scan Event |
| Đủ đơn | Manifest Completed / Handed Over |
| Chưa đủ | Handover Exception |
| Mã chưa có hệ thống | Repack / Missing System Record Exception |
| Đã có hệ thống nhưng chưa thấy | Search in Packing Area / Missing Physical Parcel |

### 10.4 ERP rules đề xuất

- Chỉ đơn `Packed` mới được vào manifest.
- Manifest phải theo ĐVVC/kênh/chuyến/ngày.
- Mỗi scan phải validate order/tracking/manifest.
- Đủ số lượng mới cho complete manifest.
- Nếu thiếu, tạo exception và không được “lờ đi”.
- Bàn giao cần sign-off hoặc file xác nhận.
- Carrier handover phải cập nhật order/shipment status.

### 10.5 Trạng thái đề xuất

#### Shipment

```text
Created
→ Picked
→ Packed
→ ReadyForHandover
→ InManifest
→ HandedOver
→ Delivered / Failed / Returned
```

#### Manifest

```text
Draft
→ Open
→ Scanning
→ Exception
→ Completed
→ Cancelled
```

#### Handover exception

```text
Open
→ Investigating
→ Resolved
→ Cancelled
```

### 10.6 Điểm cần xác minh

- ĐVVC hiện gồm những đơn vị nào?
- Có API với ĐVVC hay chỉ nội bộ quét?
- Mã quét là mã đơn, mã vận đơn hay barcode riêng?
- Có cần định danh thùng/rổ không?
- Có bao nhiêu ca bàn giao/ngày?
- Ai ký xác nhận với ĐVVC?

---

## 11. As-Is Flow G — Quy trình sản xuất / gia công ngoài

### 11.1 Flow hiện tại

```text
Lên đơn hàng với nhà máy
→ Xác nhận số lượng, quy cách, mẫu mã
→ Cọc đơn hàng, xác nhận thời gian sản xuất/nhận hàng
→ Gia công
→ Chuyển kho nguyên vật liệu từ công ty qua bên nhà máy sản xuất
→ Ký biên bản bàn giao, kèm giấy tờ COA, MSDS, tem phụ, hóa đơn VAT nếu cần
→ Bao bì nguyên vật liệu
→ Làm mẫu → chốt mẫu
→ Lưu mẫu
→ Sản xuất hàng loạt → kiểm nghiệm sản phẩm phía nhà sản xuất, đưa lại bên mình
→ Giao hàng về kho
→ Kiểm tra số lượng và chất lượng hàng hóa
→ Nếu hàng được nhận: nhập kho
→ Nếu hàng không nhận: báo lại nhà máy về vấn đề của hàng hóa trong vòng 3–7 ngày
→ Thanh toán lần cuối
```

### 11.2 Nhận định quan trọng

Mô hình hiện tại là **subcontract manufacturing / gia công ngoài**, không phải xưởng nội bộ hoàn toàn.

Vì vậy ERP Phase 1 cần module hoặc nhánh quy trình:

```text
Subcontract Order
Subcontract Material Transfer
Sample Approval
Subcontract Production Tracking
Subcontract Finished Goods Receipt
Subcontract QC Inspection
Factory Defect Claim
Payment Milestone
```

### 11.3 ERP object đề xuất

| Bước | ERP object |
|---|---|
| Lên đơn hàng với nhà máy | Subcontract Manufacturing Order |
| Xác nhận số lượng/quy cách/mẫu | Manufacturing Spec Confirmation |
| Cọc đơn | Payment Milestone: Deposit |
| Chuyển NVL/bao bì | Subcontract Material Issue |
| Ký biên bản bàn giao | Material Handover Record |
| COA/MSDS/tem phụ/VAT | Attachment |
| Làm mẫu/chốt mẫu | Sample Approval |
| Lưu mẫu | Retained Sample Record |
| Sản xuất hàng loạt | Production Status Tracking |
| Giao hàng về kho | Subcontract Delivery Receipt |
| Kiểm số lượng/chất lượng | Receiving QC Inspection |
| Hàng được nhận | Finished Goods Receipt |
| Hàng không nhận | Factory Defect Claim |
| Báo nhà máy 3–7 ngày | Claim SLA |
| Thanh toán lần cuối | Payment Milestone: Final |

### 11.4 ERP rules đề xuất

- Không cho chuyển NVL/bao bì ra nhà máy nếu chưa có đơn gia công approved.
- NVL/bao bì chuyển đi không được coi là mất/hủy; phải nằm ở trạng thái `SUBCONTRACT_ISSUED`.
- Không cho sản xuất hàng loạt nếu chưa chốt mẫu, trừ override có approval.
- Khi nhận hàng về, mặc định vào trạng thái QC hold/pending inspection.
- Final payment chỉ được mở nếu có nghiệm thu hoặc exception approval.
- Hàng không đạt phải tạo factory claim và theo dõi SLA 3–7 ngày.

### 11.5 Điểm cần xác minh

- Công ty có nắm công thức/BOM đầy đủ không hay nhà máy giữ một phần?
- NVL/bao bì do công ty cấp hay nhà máy tự mua?
- Có bao nhiêu nhà máy gia công?
- Có theo dõi hao hụt NVL tại nhà máy không?
- Có đối chiếu NVL cấp đi vs thành phẩm nhận về không?
- Mẫu lưu ở đâu và ai quản?
- “Kiểm nghiệm sản phẩm” do nhà máy hay bên mình chịu trách nhiệm cuối?

---

## 12. As-Is operating map tổng thể

```text
[Đơn bán trong ngày]
        ↓
[Kho nhận queue đơn]
        ↓
[Phân loại theo ĐVVC / sàn / đơn lẻ]
        ↓
[Pick từng đơn]
        ↓
[Pack + kiểm SKU/số lượng]
        ↓
[Đưa sang khu bàn giao]
        ↓
[Manifest + quét mã + đối chiếu số lượng]
        ↓
[Bàn giao ĐVVC]
        ↓
[ĐVVC giao khách / hoàn hàng nếu fail]
        ↓
[Hàng hoàn về kho]
        ↓
[Scan + ghi tình trạng + kiểm tra]
        ↓
[Restock hoặc chuyển Lab]
```

Song song:

```text
[Đặt nhà máy gia công]
        ↓
[Xác nhận số lượng/quy cách/mẫu]
        ↓
[Cọc đơn + chốt thời gian]
        ↓
[Chuyển NVL/bao bì sang nhà máy]
        ↓
[Làm mẫu/chốt mẫu/lưu mẫu]
        ↓
[Sản xuất hàng loạt]
        ↓
[Giao hàng về kho]
        ↓
[Kho nhận + QC]
        ↓
[Nhập kho hoặc claim nhà máy]
        ↓
[Thanh toán cuối]
```

---

## 13. Gap sơ bộ so với bộ ERP đã thiết kế

### 13.1 Những điểm bộ ERP trước đã đúng hướng

- Có WMS/kho.
- Có batch/lô/hạn dùng.
- Có QC/QA.
- Có sales order và giao hàng.
- Có hàng hoàn/returns ở mức module.
- Có production/subcontract manufacturing đã được bổ sung trong các tài liệu kỹ thuật sau.
- Có stock ledger bất biến.
- Có audit log và permission.

### 13.2 Những điểm cần nhấn mạnh hơn trong v1.1

| Điểm thực tế | Tài liệu cần cập nhật |
|---|---|
| Kho có kiểm kê và đối soát cuối ngày | PRD, Process Flow, Screen List, API, DB |
| Đóng hàng phân loại theo ĐVVC/sàn/đơn lẻ | PRD, Screen List, API |
| Bàn giao ĐVVC có thùng/rổ, bảng số lượng, scan và exception thiếu đơn | PRD, Process Flow, DB, API, UAT |
| Hàng hoàn có quay/ghi tình trạng thực tế và quyết định còn dùng/không dùng | PRD, Data Dictionary, Screen List, UAT |
| Hàng không sử dụng chuyển Lab | Data Dictionary, Inventory, Return Flow |
| Gia công ngoài có cọc, chuyển NVL/bao bì, chốt mẫu, nhận hàng, claim 3–7 ngày | PRD, Process Flow, DB, API, Finance |
| Trưởng kho lưu phiếu nhập/xuất | Audit/Attachment/SOP |

### 13.3 Những thứ cần quyết định trong Gap Analysis

- Có quản lý tote/thùng/rổ bằng mã riêng không?
- Có bắt buộc scan ở pick, pack, handover hay chỉ handover?
- Hàng hoàn “còn sử dụng” có cần QC duyệt không?
- Lab có là một kho riêng hay trạng thái stock riêng?
- Gia công ngoài có cần theo dõi NVL cấp đi vs thành phẩm nhận về trong Phase 1 không?
- Có mở final payment tự động sau khi QC pass không, hay Finance kiểm thủ công?

---

## 14. Mapping As-Is sang ERP module

| As-Is activity | ERP module | Use case |
|---|---|---|
| Tiếp nhận đơn trong ngày | Sales/OMS + WMS | Today Order Queue |
| Thực hiện xuất/nhập | Inventory/WMS | Inbound/Outbound Tasks |
| Soạn hàng | WMS | Pick Task |
| Đóng gói | WMS/Shipping | Pack Verification |
| Sắp xếp vị trí kho | WMS | Putaway / Location Transfer |
| Kiểm kê cuối ngày | Inventory | Cycle Count / Shift Count |
| Đối soát báo cáo quản lý | Dashboard/BI | Shift Closing Report |
| Nhận chứng từ giao hàng | Purchase/Inventory | Inbound Receipt |
| Kiểm số lượng/bao bì/lô | QC/Inventory | Receiving Inspection |
| Trả NCC | Purchase/Inventory | Supplier Return |
| Phiếu xuất kho | Inventory | Stock Issue |
| Bàn giao nội bộ | Inventory/Shipping | Handover Record |
| Lọc theo ĐVVC/sàn | Shipping | Wave/Batch Picking |
| Đếm đơn theo sàn/ĐVVC | Shipping | Manifest Count |
| Quét mã bàn giao | Shipping | Handover Scan |
| Thiếu đơn | Shipping | Handover Exception |
| Nhận hàng hoàn | Returns | Return Receipt |
| Quay/ghi tình trạng | Returns | Return Evidence |
| Còn sử dụng | Returns/Inventory | Return Restock |
| Không sử dụng | Returns/Inventory/QC | To Lab / Damaged |
| Đặt nhà máy | Production/Subcontract | Subcontract Order |
| Chuyển NVL/bao bì | Inventory/Subcontract | Material Issue to Factory |
| Chốt mẫu | R&D/QC/Subcontract | Sample Approval |
| Nhận hàng nhà máy | Inventory/QC/Subcontract | Finished Goods Receipt |
| Báo lỗi nhà máy | Subcontract/QC | Factory Claim |
| Thanh toán cọc/cuối | Finance/Purchase | Payment Milestone |

---

## 15. KPI/Report nên có từ As-Is

### 15.1 Kho hằng ngày

- Số đơn nhận trong ngày
- Số đơn picked
- Số đơn packed
- Số đơn handed over
- Số đơn pending cuối ca
- Số đơn thiếu/lỗi khi bàn giao
- Số lượng kiểm kê lệch cuối ngày
- Thời gian xử lý đơn trung bình

### 15.2 Nhập/xuất kho

- Số phiếu nhập
- Tỷ lệ hàng nhập không đạt
- Số phiếu xuất
- Tỷ lệ xuất thiếu/sai
- Lệch tồn theo SKU/batch

### 15.3 Hàng hoàn

- Số hàng hoàn theo ngày
- Tỷ lệ còn sử dụng
- Tỷ lệ chuyển Lab/không dùng
- Lý do hoàn phổ biến
- Hàng hoàn theo ĐVVC/kênh/sản phẩm

### 15.4 Bàn giao ĐVVC

- Số manifest/ngày
- Số đơn/manifest
- Tỷ lệ scan đủ ngay lần đầu
- Số exception thiếu đơn
- Thời gian xử lý exception
- ĐVVC nào phát sinh nhiều hoàn/thiếu

### 15.5 Gia công ngoài

- Số đơn gia công đang chạy
- Tỷ lệ đúng hạn nhà máy
- Tỷ lệ hàng không nhận/claim
- Số ngày từ giao NVL → nhận thành phẩm
- Hao hụt NVL nếu có theo dõi
- Payment milestone pending

---

## 16. Rủi ro vận hành hiện tại nếu chưa có ERP

| Rủi ro | Mô tả | Mức độ |
|---|---|---|
| Lệch tồn cuối ngày | Kiểm kê/đối soát thủ công dễ sai | Cao |
| Đơn packed nhưng chưa bàn giao | Nếu thiếu scan/manifest, dễ thất lạc | Cao |
| Hàng hoàn nhập sai trạng thái | Hàng hỏng có thể quay lại bán | Rất cao |
| Không trace được batch khi hàng hoàn/khiếu nại | Rủi ro mỹ phẩm | Rất cao |
| Chứng từ giấy thất lạc | Phiếu nhập/xuất/bàn giao phụ thuộc lưu thủ công | Trung bình/Cao |
| Gia công ngoài không kiểm NVL cấp đi | Khó biết hao hụt/thất thoát tại nhà máy | Cao |
| Chốt mẫu không có record số hóa | Dễ tranh cãi chất lượng/mẫu | Cao |
| Báo lỗi nhà máy quá hạn 3–7 ngày | Mất quyền khiếu nại/đòi xử lý | Cao |
| Thanh toán cuối trước nghiệm thu | Rủi ro tài chính | Cao |

---

## 17. Ưu tiên đưa vào ERP Phase 1

### P0 — bắt buộc

1. Daily Order Queue cho kho
2. Inbound Receipt + check số lượng/bao bì/lô
3. Outbound Issue/Pick Task
4. Pack Verification
5. Carrier Manifest + Handover Scan
6. Handover Exception
7. Return Receiving + Return Inspection + Return Disposition
8. Stock Ledger + Batch/Expiry
9. End-of-Day Cycle Count / Shift Closing
10. Subcontract Order + Material Issue to Factory + Finished Receipt
11. Sample Approval tối thiểu
12. Factory Claim 3–7 day SLA

### P1 — rất nên có

1. Tote/Container management
2. Location/bin optimization
3. Ảnh/video hàng hoàn bắt buộc
4. Digital signature/attachment cho bàn giao
5. Payment milestone cho gia công
6. Supplier/factory scorecard
7. Dashboard vận hành kho/ngày

### P2 — sau Phase 1

1. Portal nhà máy/NCC
2. Portal ĐVVC
3. Mobile app scan chuyên dụng
4. Advanced wave picking
5. Advanced labor productivity
6. AI forecast/slotting optimization

---

## 18. Các tài liệu lõi cần revise lên v1.1

Sau khi file 21 Gap Analysis được chốt, cần cập nhật:

| File | Nội dung cần cập nhật |
|---|---|
| 03 PRD/SRS | Bổ sung As-Is handover, hàng hoàn, shift closing, subcontract manufacturing |
| 04 Permission/Approval | Thêm role handover, return receiver, subcontract coordinator, approval final payment |
| 05 Data Dictionary | Thêm return disposition, manifest status, subcontract status, shift closing fields |
| 06 Process Flow | Bổ sung flow đúng theo 4 PDF |
| 08 Screen List/Wireframe | Thêm Daily Board, Handover Scan, Return Inspection, Subcontract screens |
| 09 UAT | Test case cho scan thiếu đơn, hàng hoàn, final payment, claim nhà máy |
| 11 Technical Architecture | Đã có nhánh Go/subcontract, cần cross-check với As-Is |
| 16 API Contract | Bổ sung endpoint nếu thiếu |
| 17 Database Schema | Bổ sung table/entity nếu thiếu |
| 19 Security | Bổ sung permission/audit cho các điểm thực tế |

---

## 19. Câu hỏi xác minh với team nội bộ

### 19.1 Kho

1. Một ngày có bao nhiêu ca kho?
2. Ai là người chốt ca?
3. Cuối ngày kiểm kê toàn bộ hay kiểm kê mẫu/các SKU trọng yếu?
4. Có dùng vị trí kho/bin không?
5. Có barcode cho từng SKU/batch không?
6. Phiếu nhập/xuất hiện đang dùng giấy, Excel hay phần mềm nào?

### 19.2 Đóng hàng/giao hàng

1. Đơn được phân loại theo ĐVVC hay theo sàn trước?
2. Mã quét là mã đơn, mã vận đơn hay mã nội bộ?
3. Có bao nhiêu ĐVVC?
4. Có tình huống một manifest nhiều thùng/rổ không?
5. Thiếu đơn thì ai có quyền quyết định đóng lại?
6. Có cần chữ ký điện tử/ảnh biên bản ĐVVC không?

### 19.3 Hàng hoàn

1. Hàng hoàn đến từ ĐVVC nào/kênh nào?
2. Ai quay/ghi tình trạng thực tế?
3. Ai quyết định còn sử dụng/không sử dụng?
4. Lab xử lý hàng không sử dụng như thế nào?
5. Có hoàn tiền/đổi hàng tự động không?
6. Có cần QC duyệt lại trước khi restock không?

### 19.4 Gia công ngoài

1. Công ty cấp những NVL/bao bì nào cho nhà máy?
2. Nhà máy tự mua phần nào?
3. Có theo dõi batch NVL cấp cho nhà máy không?
4. Nhà máy có trả lại NVL dư không?
5. Có đối chiếu hao hụt không?
6. Ai duyệt mẫu?
7. Ai lưu mẫu?
8. Ai kiểm hàng khi nhà máy giao về?
9. Chính sách báo lỗi 3–7 ngày có tính từ ngày nhận hay ngày kiểm xong?
10. Điều kiện thanh toán cuối là gì?

---

## 20. Kết luận As-Is

Workflow hiện tại của công ty có cấu trúc khá rõ, đặc biệt ở kho:

```text
đơn trong ngày
→ xuất/nhập theo nội quy
→ pick/pack
→ bàn giao ĐVVC
→ kiểm kê/đối soát cuối ngày
```

Điểm khác biệt quan trọng so với ERP chuẩn là:

1. **Bàn giao ĐVVC có flow riêng và phải scan/đối chiếu số lượng.**
2. **Hàng hoàn có kiểm tình trạng và phân loại còn dùng/không dùng, không được nhập thẳng tồn bán.**
3. **Sản xuất hiện là gia công ngoài, có chuyển NVL/bao bì, duyệt mẫu, nhận hàng, báo lỗi nhà máy và thanh toán milestone.**
4. **Kho có đóng ca và đối soát cuối ngày, cần hệ thống hóa bằng shift closing.**

Do đó, trước khi dev build mạnh, cần làm ngay file Gap Analysis để chốt:

```text
Cái nào giữ theo workflow hiện tại
Cái nào đổi theo ERP best practice
Cái nào thiết kế lại
```

File tiếp theo nên là:

```text
21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md
```
