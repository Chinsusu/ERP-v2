# 26_ERP_SOP_Training_Manual_Phase1_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Phiên bản:** v1.0  
**Ngày tạo:** 2026-04-24  
**Tài liệu:** SOP & Training Manual  
**Mục tiêu:** Hướng dẫn nhân viên vận hành ERP đúng quy trình, giảm nhập sai dữ liệu, giảm lệch kho, giảm lỗi bàn giao, giảm tranh cãi giữa các bộ phận.

---

## 1. Mục đích tài liệu

Tài liệu này dùng để đào tạo người dùng nội bộ khi triển khai ERP Phase 1.

Nó trả lời các câu hỏi rất thực tế:

- Nhân viên kho vào màn hình nào để tiếp nhận đơn trong ngày?
- Nhân viên kho nhập hàng, xuất hàng, đóng hàng, bàn giao ĐVVC ra sao?
- QC kiểm hàng và đổi trạng thái batch như thế nào?
- Hàng hoàn nhận từ shipper thì xử lý thế nào?
- Đơn gia công với nhà máy đi từ cọc đơn, chuyển NVL/bao bì, duyệt mẫu, nhận hàng tới thanh toán ra sao?
- Trưởng kho, quản lý, kế toán, sales admin cần kiểm tra và duyệt gì?
- Cuối ca phải đối soát số liệu gì trước khi kết thúc ngày làm việc?

SOP này không chỉ mô tả thao tác phần mềm. Nó là **quy chuẩn hành vi vận hành**: ai làm gì, làm khi nào, ghi nhận dữ liệu nào, chứng từ nào phải lưu, lỗi nào phải báo ngay.

---

## 2. Phạm vi Phase 1

Tài liệu này bao phủ các nghiệp vụ Phase 1 sau:

1. Đăng nhập, portal theo vai trò, thao tác ERP cơ bản.
2. Dữ liệu gốc: SKU, nguyên vật liệu, nhà cung cấp, kho, khách hàng, batch.
3. Mua hàng và nhận hàng.
4. QC đầu vào, QC batch, hold/pass/fail.
5. Kho: nhập, xuất, chuyển, kiểm kê, tồn khả dụng, hàng hold.
6. Tiếp nhận đơn trong ngày.
7. Soạn hàng, đóng hàng.
8. Bàn giao đơn cho đơn vị vận chuyển.
9. Xử lý hàng hoàn.
10. Sản xuất/gia công ngoài với nhà máy.
11. Đối soát cuối ca và báo cáo quản lý.
12. Các thao tác phê duyệt cơ bản.
13. Hướng dẫn xử lý lỗi thường gặp.

Ngoài phạm vi Phase 1:

- Payroll HRM đầy đủ.
- KOL/Affiliate payout nâng cao.
- CRM loyalty nâng cao.
- BI nâng cao.
- Accounting posting chi tiết chuẩn kế toán đầy đủ.
- AI forecast.

---

## 3. Tài liệu liên quan

Tài liệu này dựa trên bộ tài liệu đã chốt trước đó:

- `03_ERP_PRD_SRS_Phase1_My_Pham_v1.md`
- `04_ERP_Permission_Approval_Matrix_Phase1_My_Pham_v1.md`
- `05_ERP_Data_Dictionary_Master_Data_Phase1_My_Pham_v1.md`
- `06_ERP_Process_Flow_ToBe_Phase1_My_Pham_v1.md`
- `08_ERP_Screen_List_Wireframe_Phase1_My_Pham_v1.md`
- `09_ERP_UAT_Test_Scenarios_Phase1_My_Pham_v1.md`
- `11_ERP_Technical_Architecture_Go_Backend_Phase1_MyPham_v1.md`
- `14_ERP_UI_UX_Design_System_Standards_Phase1_MyPham_v1.md`
- `20_ERP_Current_Workflow_AsIs_Warehouse_Production_Phase1_MyPham_v1.md`
- `21_ERP_Gap_Analysis_AsIs_vs_ToBe_Decision_Log_Phase1_MyPham_v1.md`
- `25_ERP_Product_Backlog_Sprint_Plan_Phase1_MyPham_v1.md`

Tài liệu này cũng bám theo workflow thực tế đã gửi:

- Công việc hằng ngày của kho.
- Nội quy nhập kho, xuất kho, đóng hàng, hàng hoàn.
- Quy trình bàn giao hàng cho ĐVVC.
- Quy trình sản xuất/gia công ngoài với nhà máy.

---

## 4. Nguyên tắc vận hành ERP

### 4.1. ERP là nguồn dữ liệu thật

Sau khi go-live, mọi giao dịch nghiệp vụ quan trọng phải được ghi nhận trên ERP.

Không dùng Excel/Zalo/giấy làm nguồn sự thật chính cho:

- tồn kho,
- nhập kho,
- xuất kho,
- batch,
- QC,
- đơn hàng,
- bàn giao ĐVVC,
- hàng hoàn,
- gia công ngoài,
- công nợ/liên quan thanh toán.

Excel có thể dùng để export báo cáo hoặc import dữ liệu theo quyền được duyệt, nhưng không được dùng để thay thế ERP.

### 4.2. Không sửa tay dữ liệu tồn kho

Tồn kho không được sửa trực tiếp.

Tất cả thay đổi tồn phải đi qua chứng từ:

- nhập kho,
- xuất kho,
- chuyển kho,
- kiểm kê,
- điều chỉnh tồn,
- QC release,
- return receipt,
- production/subcontract receipt,
- hàng hỏng/hàng không sử dụng.

### 4.3. Batch, hạn dùng, QC status là dữ liệu sống còn

Với mỹ phẩm, các trường sau không được nhập qua loa:

- mã lô/batch,
- ngày sản xuất,
- hạn sử dụng,
- trạng thái QC,
- vị trí kho,
- tình trạng hàng.

Sai các dữ liệu này có thể dẫn tới bán nhầm hàng lỗi, hàng cận date, hàng chưa QC, hoặc không truy xuất được khi có khiếu nại.

### 4.4. Chứng từ quan trọng không được xoá

Chứng từ đã phát sinh nghiệp vụ không được xoá khỏi hệ thống.

Nếu sai, xử lý bằng:

- huỷ chứng từ,
- tạo phiếu điều chỉnh,
- tạo chứng từ đảo,
- ghi lý do,
- lưu audit log.

### 4.5. Dữ liệu phải đi kèm người chịu trách nhiệm

Mỗi chứng từ phải có:

- người tạo,
- thời điểm tạo,
- người cập nhật,
- người duyệt nếu có,
- người thực hiện,
- người bàn giao/nhận bàn giao nếu có.

Không dùng tài khoản chung.

### 4.6. Quét mã khi có thể

Các điểm nên dùng scan/quét mã:

- soạn hàng,
- đóng hàng,
- bàn giao ĐVVC,
- nhận hàng hoàn,
- kiểm kê,
- tra batch/SKU,
- xác nhận vị trí kho.

Scan giúp giảm lỗi nhập tay và tạo bằng chứng vận hành.

---

## 5. Vai trò người dùng trong đào tạo

### 5.1. Admin hệ thống

Trách nhiệm:

- tạo user,
- gán vai trò,
- cấu hình quyền,
- hỗ trợ reset mật khẩu,
- cấu hình danh mục cơ bản,
- không can thiệp nghiệp vụ nếu không được phân quyền.

### 5.2. Trưởng kho

Trách nhiệm:

- kiểm soát nhập/xuất/chuyển kho,
- duyệt hoặc xác nhận các phiếu quan trọng,
- điều phối soạn/đóng/bàn giao,
- kiểm tra đối soát cuối ca,
- lưu trữ chứng từ vận hành,
- xử lý lệch tồn ban đầu.

### 5.3. Nhân viên kho

Trách nhiệm:

- tiếp nhận đơn,
- soạn hàng,
- đóng hàng,
- scan bàn giao,
- nhận hàng hoàn,
- nhập/xuất kho theo phiếu,
- kiểm kê theo phân công.

### 5.4. QC/QA

Trách nhiệm:

- kiểm hàng đầu vào,
- kiểm batch,
- set hold/pass/fail,
- ghi nhận lỗi,
- phát hành kết quả QC,
- phối hợp xử lý hàng lỗi/hàng hoàn/lỗi nhà máy.

### 5.5. Sales Admin / CSKH

Trách nhiệm:

- tạo/xác nhận đơn hàng,
- kiểm tra trạng thái đơn,
- phối hợp xử lý giao hàng/thất lạc/hàng hoàn,
- ghi nhận khiếu nại khách hàng,
- không can thiệp tồn kho trực tiếp.

### 5.6. Purchasing / Sản xuất gia công

Trách nhiệm:

- tạo PR/PO hoặc đơn gia công,
- theo dõi nhà cung cấp/nhà máy,
- theo dõi cọc đơn, thời gian giao,
- tạo yêu cầu chuyển NVL/bao bì cho nhà máy,
- theo dõi duyệt mẫu,
- phối hợp nhận hàng và claim lỗi.

### 5.7. Finance/Kế toán

Trách nhiệm:

- kiểm tra công nợ,
- kiểm tra thanh toán cọc và thanh toán cuối,
- đối soát COD/bank nếu có,
- không sửa chứng từ kho hoặc QC.

### 5.8. Quản lý/CEO/COO

Trách nhiệm:

- xem dashboard,
- duyệt giao dịch vượt ngưỡng,
- xử lý exception,
- theo dõi KPI vận hành,
- quyết định khi có rủi ro lớn.

---

## 6. Cấu trúc đào tạo đề xuất

### 6.1. Training theo vai trò

Không đào tạo tất cả mọi người bằng một buổi chung duy nhất.

Nên chia:

1. Buổi tổng quan cho toàn bộ user.
2. Buổi riêng cho kho.
3. Buổi riêng cho QC/QA.
4. Buổi riêng cho sales/CSKH.
5. Buổi riêng cho purchasing/sản xuất gia công.
6. Buổi riêng cho finance.
7. Buổi riêng cho quản lý.
8. Buổi UAT mô phỏng end-to-end.

### 6.2. Nguyên tắc đào tạo

Mỗi buổi phải có:

- demo trên hệ thống staging/UAT,
- dữ liệu mẫu giống thực tế,
- bài tập thao tác,
- lỗi giả lập,
- checklist ký xác nhận đã học,
- người phụ trách hỗ trợ sau đào tạo.

### 6.3. Kịch bản đào tạo tối thiểu

Mỗi user cần thực hành tối thiểu:

- đăng nhập,
- tìm kiếm dữ liệu,
- tạo chứng từ đúng quyền,
- submit/approve nếu có,
- xem audit/attachment nếu cần,
- xử lý lỗi thường gặp,
- biết khi nào phải báo quản lý.

---

## 7. Quy ước thao tác chung trên ERP

### 7.1. Màn hình danh sách

Màn hình danh sách thường có:

- bộ lọc,
- ô tìm kiếm,
- bảng dữ liệu,
- trạng thái chứng từ,
- nút tạo mới,
- export theo quyền,
- cột hành động.

Quy tắc:

- luôn lọc theo ngày/kho/trạng thái trước khi xử lý,
- không xử lý nhầm chứng từ của kho khác,
- kiểm tra status trước khi bấm action.

### 7.2. Màn hình chi tiết

Màn hình chi tiết thường có:

- thông tin header,
- danh sách dòng hàng,
- tab attachment,
- tab audit log,
- tab approval,
- nút action theo trạng thái.

Quy tắc:

- đọc kỹ header trước khi xử lý,
- kiểm tra từng dòng hàng,
- kiểm tra batch/hạn dùng nếu liên quan kho,
- nếu thấy sai phải báo ngay, không tự sửa ngoài quyền.

### 7.3. Status chip

Màu/trạng thái phải được hiểu đúng:

- Draft: mới tạo, chưa có hiệu lực.
- Submitted: đã gửi chờ duyệt.
- Approved: đã duyệt.
- Processing: đang thực hiện.
- Completed/Closed: đã hoàn tất.
- Cancelled: đã huỷ.
- Hold: đang giữ/khóa, chưa được dùng hoặc bán.
- Pass: đạt.
- Fail: không đạt.

### 7.4. File đính kèm

Nên đính kèm:

- hóa đơn,
- phiếu giao hàng,
- ảnh hàng lỗi,
- video hàng hoàn nếu có,
- COA/MSDS,
- biên bản bàn giao,
- ảnh chứng từ ĐVVC,
- ảnh mẫu đã duyệt.

File đính kèm phải rõ, đọc được, không upload file mờ/không liên quan.

### 7.5. Ghi chú nghiệp vụ

Khi ghi chú, phải ghi rõ:

- chuyện gì xảy ra,
- ai liên quan,
- thời điểm,
- mã đơn/mã batch/mã phiếu,
- hành động tiếp theo.

Không ghi chú chung chung kiểu “đã xử lý”, “ok”, “xong”.

---

# PHẦN A — SOP CƠ BẢN

---

## SOP-01. Đăng nhập ERP

### Mục tiêu

Đảm bảo user đăng nhập đúng tài khoản cá nhân, đúng vai trò, không dùng tài khoản chung.

### Ai thực hiện

Tất cả người dùng.

### Các bước

1. Mở đường dẫn ERP.
2. Nhập email/tài khoản cá nhân.
3. Nhập mật khẩu.
4. Nếu bật MFA, nhập mã xác thực.
5. Kiểm tra portal hiển thị đúng vai trò.
6. Nếu không thấy menu cần dùng, báo Admin/Quản lý, không mượn tài khoản người khác.

### Lưu ý

- Không chia sẻ mật khẩu.
- Không lưu mật khẩu trên máy dùng chung nếu chưa được IT cho phép.
- Rời máy phải logout hoặc khóa màn hình.

### Lỗi thường gặp

| Lỗi | Cách xử lý |
|---|---|
| Không đăng nhập được | Kiểm tra email/mật khẩu, sau đó báo Admin reset |
| Thiếu menu | Báo trưởng bộ phận xác nhận quyền rồi gửi Admin |
| Vào nhầm kho/chi nhánh | Dừng thao tác, đổi context hoặc báo Admin |

---

## SOP-02. Tìm kiếm chứng từ/dữ liệu

### Mục tiêu

Giúp user tìm đúng chứng từ trước khi xử lý.

### Ai thực hiện

Tất cả người dùng.

### Các bước

1. Vào module liên quan.
2. Chọn bộ lọc ngày/kho/trạng thái.
3. Tìm theo mã chứng từ, mã đơn, mã SKU, mã batch hoặc mã vận đơn.
4. Mở chi tiết chứng từ.
5. Kiểm tra trạng thái và thông tin chính.

### Lưu ý

- Luôn kiểm tra ngày và kho.
- Nếu có nhiều kết quả giống nhau, mở chi tiết để xác nhận.
- Không xử lý chứng từ chỉ dựa vào tên khách hoặc tên sản phẩm.

---

## SOP-03. Tạo yêu cầu hỗ trợ/lỗi hệ thống

### Mục tiêu

Chuẩn hóa cách báo lỗi, tránh nhắn rời rạc làm thất lạc vấn đề.

### Ai thực hiện

Tất cả người dùng.

### Các bước

1. Chụp màn hình lỗi.
2. Ghi mã chứng từ liên quan.
3. Ghi thời điểm xảy ra lỗi.
4. Ghi thao tác vừa làm trước khi lỗi.
5. Gửi vào kênh hỗ trợ chính thức hoặc tạo support ticket.
6. Không tự sửa dữ liệu bằng cách đi đường vòng nếu chưa được duyệt.

### Mẫu báo lỗi

```text
Module: Kho / Bàn giao ĐVVC
Mã chứng từ: MAN-20260424-001
User: Nguyễn A
Thời điểm: 16:35
Lỗi: Quét mã đơn SO-xxx báo không thuộc manifest
Đã làm: kiểm tra đơn đã packed
Ảnh/video: đính kèm
Mức độ: ảnh hưởng bàn giao hôm nay
```

---

# PHẦN B — SOP DỮ LIỆU GỐC

---

## SOP-04. Tạo/Sửa mã SKU thành phẩm

### Mục tiêu

Đảm bảo sản phẩm bán ra có mã chuẩn, không trùng, đủ thông tin kho/QC/bán hàng.

### Ai thực hiện

Master Data Admin, R&D, Sales Admin, Finance/QC xem hoặc duyệt theo quyền.

### Điều kiện trước

- Có thông tin sản phẩm được duyệt.
- Có đơn vị tính chuẩn.
- Có quy cách đóng gói.
- Có yêu cầu batch/hạn dùng nếu cần.

### Các bước

1. Vào `Master Data → Products/SKU`.
2. Bấm `Create SKU`.
3. Nhập mã SKU theo quy tắc đã chốt.
4. Nhập tên sản phẩm.
5. Chọn brand/dòng sản phẩm nếu có.
6. Chọn đơn vị tính.
7. Nhập quy cách: dung tích, size, packing unit.
8. Bật yêu cầu quản lý batch/hạn dùng.
9. Chọn trạng thái ban đầu: Draft.
10. Đính kèm spec/ảnh/bao bì nếu có.
11. Submit duyệt.
12. Người có quyền duyệt kiểm tra và active SKU.

### Không được làm

- Không tạo SKU mới chỉ vì viết khác tên.
- Không bỏ qua batch/hạn dùng cho mỹ phẩm cần truy xuất.
- Không active SKU nếu chưa đủ thông tin bắt buộc.

---

## SOP-05. Tạo mã nguyên vật liệu/bao bì

### Mục tiêu

Chuẩn hóa nguyên liệu, bao bì, tem nhãn, phụ liệu phục vụ mua hàng, kho, gia công.

### Ai thực hiện

Master Data Admin, Purchasing, R&D, QC.

### Các bước

1. Vào `Master Data → Materials`.
2. Chọn loại: nguyên liệu, bao bì, tem nhãn, phụ liệu.
3. Nhập mã material.
4. Nhập tên chuẩn.
5. Nhập đơn vị tính.
6. Cấu hình yêu cầu batch/hạn dùng nếu có.
7. Gắn approved supplier nếu có.
8. Đính kèm COA/MSDS/spec nếu có.
9. Submit duyệt.
10. QC/Purchasing kiểm tra và active.

### Lưu ý

- Nguyên liệu dùng trong sản xuất/gia công phải có hồ sơ chất lượng nếu công ty yêu cầu.
- Bao bì có version nếu thiết kế thay đổi.
- Không dùng một mã cho nhiều loại bao bì khác quy cách.

---

## SOP-06. Tạo nhà cung cấp/nhà máy gia công

### Mục tiêu

Đảm bảo NCC/nhà máy được lưu đúng, có thông tin liên hệ, điều kiện thanh toán, hồ sơ chất lượng.

### Ai thực hiện

Purchasing, Finance, Admin.

### Các bước

1. Vào `Master Data → Suppliers/Factories`.
2. Bấm `Create`.
3. Nhập tên pháp lý/tên giao dịch.
4. Chọn loại: NCC nguyên liệu, NCC bao bì, nhà máy gia công, ĐVVC, khác.
5. Nhập thông tin liên hệ.
6. Nhập MST/thông tin thanh toán nếu có.
7. Nhập điều khoản thanh toán.
8. Đính kèm hợp đồng/hồ sơ nếu có.
9. Submit duyệt.
10. Finance/Purchasing duyệt active.

### Lưu ý

- Nhà máy gia công nên có thông tin lead time, MOQ, năng lực sản xuất.
- NCC chưa active không được chọn trong PO chính thức.

---

# PHẦN C — SOP MUA HÀNG, NHẬP KHO, QC

---

## SOP-07. Tạo đề nghị mua hàng hoặc yêu cầu mua/gia công

### Mục tiêu

Chuẩn hóa bước đề xuất mua nguyên liệu, bao bì, hàng hóa hoặc đặt gia công.

### Ai thực hiện

Purchasing, Planning, R&D, Kho, bộ phận yêu cầu.

### Điều kiện trước

- Material/SKU đã có master data.
- Có nhu cầu rõ ràng: sản xuất, bổ sung tồn, hàng bán, gia công.

### Các bước

1. Vào `Purchase → Purchase Request`.
2. Bấm `Create`.
3. Chọn loại yêu cầu: mua hàng, mua NVL, mua bao bì, gia công ngoài.
4. Nhập lý do.
5. Nhập danh sách item, số lượng, ngày cần.
6. Đính kèm báo giá/chứng từ nếu có.
7. Submit duyệt.
8. Người duyệt kiểm tra ngân sách, nhu cầu, tồn kho, rồi approve/reject.

### Lưu ý

- Không tạo yêu cầu mua nếu còn tồn đủ và không có lý do đặc biệt.
- Yêu cầu gấp phải ghi rõ deadline và hậu quả nếu trễ.

---

## SOP-08. Tạo PO/đơn đặt hàng với NCC

### Mục tiêu

Tạo chứng từ mua hàng chính thức để theo dõi giao nhận, công nợ, QC.

### Ai thực hiện

Purchasing.

### Điều kiện trước

- PR đã duyệt hoặc có quyền tạo PO trực tiếp theo policy.
- Supplier active.
- Item active.

### Các bước

1. Vào `Purchase → Purchase Orders`.
2. Bấm `Create PO`.
3. Chọn supplier.
4. Link PR nếu có.
5. Nhập item, số lượng, giá, thuế nếu có.
6. Nhập expected delivery date.
7. Nhập điều khoản thanh toán.
8. Đính kèm báo giá/hợp đồng nếu có.
9. Submit duyệt.
10. PO approved mới gửi NCC.

### Lưu ý

- Không nhận hàng ngoài PO nếu quy trình yêu cầu PO.
- Nếu hàng về khác PO, phải ghi exception khi nhận.

---

## SOP-09. Nhận hàng từ NCC/nhà máy

### Mục tiêu

Ghi nhận hàng giao đến kho, kiểm chứng từ, kiểm số lượng/bao bì/lô trước khi nhập.

### Ai thực hiện

Nhân viên kho, Trưởng kho, QC nếu cần.

### Điều kiện trước

- Có PO/đơn gia công/phiếu giao hợp lệ.
- Hàng đã đến kho.

### Các bước

1. Vào `Warehouse → Receiving`.
2. Bấm `Create Goods Receipt`.
3. Chọn PO/đơn gia công liên quan.
4. Kiểm tra chứng từ giao hàng.
5. Nhập số lượng thực nhận.
6. Nhập batch/lô, ngày sản xuất, hạn dùng nếu có.
7. Kiểm tra bao bì ngoài.
8. Chụp ảnh nếu có bất thường.
9. Nếu đạt kiểm tra ban đầu, submit sang QC hoặc nhập khu chờ QC.
10. Nếu không đạt, chọn `Reject/Return to Supplier` và ghi lý do.

### Checklist kiểm tra ban đầu

- Đúng nhà cung cấp/nhà máy.
- Đúng item.
- Đúng số lượng hoặc ghi nhận chênh lệch.
- Bao bì không rách/vỡ/ướt/bẩn nghiêm trọng.
- Có batch/lô nếu bắt buộc.
- Có hạn dùng nếu bắt buộc.
- Có chứng từ đi kèm.

### Lưu ý

- Hàng chưa QC pass không được đưa vào tồn khả dụng.
- Nếu thiếu batch/hạn dùng, không tự điền bừa.
- Nếu hàng không đạt, phải ghi lý do rõ và đính kèm ảnh/chứng từ.

---

## SOP-10. QC đầu vào

### Mục tiêu

Đảm bảo hàng/NVL/bao bì trước khi dùng hoặc bán phải được QC đúng quy định.

### Ai thực hiện

QC/QA.

### Điều kiện trước

- Có Goods Receipt hoặc lô hàng chờ QC.
- Có checklist/spec kiểm tra.

### Các bước

1. Vào `QC → Inbound Inspection`.
2. Chọn phiếu nhận hàng/lô cần kiểm.
3. Kiểm tra số lượng mẫu cần lấy.
4. Ghi kết quả kiểm tra theo checklist.
5. Đính kèm ảnh/file nếu cần.
6. Chọn kết quả:
   - Pass,
   - Fail,
   - Hold/Need Review.
7. Submit kết quả.
8. Nếu Pass: hệ thống chuyển hàng sang trạng thái có thể dùng/bán theo rule.
9. Nếu Fail: hệ thống khóa hàng, báo kho/purchasing.
10. Nếu Hold: hàng nằm khu quarantine/chờ xử lý.

### Lưu ý

- QC không sửa số lượng kho nếu không có phiếu điều chỉnh.
- Batch Fail không được xuất bán hoặc cấp sản xuất.
- Batch Hold phải hiển thị rõ trên màn hình kho và sales.

---

## SOP-11. Xếp hàng vào kho / Putaway

### Mục tiêu

Đưa hàng đã nhận/QC pass vào đúng vị trí kho, giảm thất lạc, dễ kiểm kê.

### Ai thực hiện

Nhân viên kho.

### Điều kiện trước

- Hàng đã được phép putaway.
- Có vị trí kho/bin hoặc khu vực rõ.

### Các bước

1. Vào `Warehouse → Putaway Tasks`.
2. Chọn phiếu/lô hàng cần xếp.
3. Kiểm tra SKU/batch/số lượng.
4. Chọn vị trí kho đề xuất hoặc nhập vị trí thực tế.
5. Di chuyển hàng vào vị trí.
6. Scan/confirm vị trí nếu có mã vị trí.
7. Hoàn tất task.

### Lưu ý

- Không đặt hàng chưa QC pass vào khu hàng bán được.
- Hàng cận date nên để vị trí dễ xuất theo FEFO.
- Hàng lỗi/hàng hold phải để khu riêng.

---

# PHẦN D — SOP KHO HẰNG NGÀY

---

## SOP-12. Mở ca kho / Warehouse Daily Board

### Mục tiêu

Bắt đầu ca làm bằng việc nhìn rõ workload trong ngày: đơn cần xử lý, hàng cần nhập, hàng cần xuất, hàng hoàn, kiểm kê, bàn giao.

### Ai thực hiện

Trưởng kho, nhân viên kho.

### Thời điểm

Đầu mỗi ca/ngày làm việc.

### Các bước

1. Vào `Warehouse → Daily Board`.
2. Chọn ngày làm việc và kho.
3. Kiểm tra các nhóm công việc:
   - đơn hàng mới,
   - đơn chờ soạn,
   - đơn chờ đóng,
   - đơn chờ bàn giao,
   - hàng cần nhập,
   - hàng chờ QC,
   - hàng hoàn,
   - task kiểm kê.
4. Trưởng kho phân công nhân sự.
5. Nhân viên xác nhận nhận task nếu hệ thống có tính năng assign.

### Lưu ý

- Daily Board là màn hình điều phối chính của kho.
- Nếu có đơn gấp hoặc đơn ưu tiên, trưởng kho đánh dấu priority.

---

## SOP-13. Tiếp nhận đơn hàng trong ngày

### Mục tiêu

Tổng hợp đơn cần xử lý trong ngày để chuyển sang soạn/đóng/bàn giao.

### Ai thực hiện

Sales Admin, CSKH, Kho.

### Điều kiện trước

- Đơn đã được xác nhận hợp lệ.
- Tồn khả dụng đủ hoặc có rule xử lý thiếu.

### Các bước

1. Vào `Sales → Orders` hoặc `Warehouse → Daily Board`.
2. Lọc trạng thái `Confirmed/Ready to Fulfill`.
3. Kiểm tra kênh bán, ưu tiên, địa chỉ, ĐVVC.
4. Kiểm tra tồn khả dụng.
5. Nếu đủ hàng, chuyển đơn sang `Ready to Pick`.
6. Nếu thiếu hàng, báo Sales/CSKH hoặc tạo backorder theo policy.

### Lưu ý

- Không đẩy đơn sang kho nếu thông tin giao hàng chưa hợp lệ.
- Không tự đổi SKU/quà tặng khi chưa có xác nhận.

---

## SOP-14. Soạn hàng / Picking

### Mục tiêu

Lấy đúng hàng, đúng SKU, đúng batch, đúng số lượng cho từng đơn.

### Ai thực hiện

Nhân viên kho.

### Điều kiện trước

- Đơn ở trạng thái `Ready to Pick`.
- Có picking list.

### Các bước

1. Vào `Warehouse → Picking`.
2. Chọn wave/batch đơn hoặc từng đơn.
3. In hoặc mở picking list.
4. Đi đến vị trí kho được chỉ định.
5. Lấy hàng theo SKU/batch/số lượng.
6. Scan SKU/batch nếu có.
7. Nếu thiếu hàng, bấm `Report Shortage`.
8. Xác nhận hoàn tất picking.
9. Chuyển hàng sang khu vực đóng hàng.

### Checklist

- Đúng SKU.
- Đúng số lượng.
- Đúng batch theo FEFO/FIFO nếu hệ thống đề xuất.
- Không lấy hàng Hold/Fail/QC chưa pass.
- Không lấy hàng ở khu hàng hoàn/hàng lỗi nếu chưa được release.

### Lưu ý

- Nếu thực tế khác hệ thống, không tự thay bằng hàng khác.
- Báo trưởng kho xử lý thiếu/hư hỏng/lệch vị trí.

---

## SOP-15. Đóng hàng / Packing

### Mục tiêu

Đóng đúng đơn, kiểm tra lại SKU/số lượng/trạng thái hàng trước khi bàn giao ĐVVC.

### Ai thực hiện

Nhân viên đóng hàng, Trưởng kho kiểm soát.

### Điều kiện trước

- Đơn đã pick xong.
- Hàng đã chuyển đến khu vực đóng hàng.

### Các bước

1. Vào `Warehouse → Packing`.
2. Quét hoặc chọn mã đơn.
3. Kiểm tra danh sách SKU/số lượng.
4. Kiểm tra thực tế hàng trong khu vực đóng.
5. Đóng gói theo quy chuẩn.
6. Dán tem/mã vận đơn nếu có.
7. Chụp ảnh nếu đơn yêu cầu bằng chứng.
8. Xác nhận `Packed`.
9. Chuyển đơn sang khu vực chờ bàn giao ĐVVC.

### Theo workflow thực tế cần chú ý

- Phân loại đơn theo ĐVVC/kênh/sàn nếu có.
- Soạn theo từng đơn.
- Đóng gói và kiểm lại khu vực đóng hàng.
- Đếm lại tổng số lượng đơn của mỗi sàn/kênh/ĐVVC trước bàn giao.

### Lỗi thường gặp

| Lỗi | Xử lý |
|---|---|
| Thiếu sản phẩm trong đơn | Report shortage, không xác nhận packed |
| Scan sai SKU | Bỏ SKU sai ra, scan lại đúng hàng |
| Đơn không có mã vận đơn | Báo Sales/CSKH/Shipping Admin |
| Hàng móp/hỏng khi đóng | Tách ra, báo trưởng kho/QC |

---

## SOP-16. Sắp xếp, tối ưu vị trí kho trong ngày

### Mục tiêu

Giữ kho gọn, giảm nhầm lẫn khi picking, hỗ trợ kiểm kê cuối ngày.

### Ai thực hiện

Nhân viên kho, Trưởng kho.

### Các bước

1. Vào `Warehouse → Location Management` nếu cần cập nhật vị trí.
2. Kiểm tra các khu vực:
   - hàng bán được,
   - hàng chờ QC,
   - hàng hold,
   - hàng hoàn,
   - hàng lỗi,
   - khu đóng hàng,
   - khu chờ bàn giao.
3. Di chuyển hàng đúng khu vực.
4. Nếu di chuyển vị trí chính thức, tạo `Internal Transfer` hoặc `Location Move`.
5. Scan xác nhận vị trí mới.

### Lưu ý

- Không di chuyển hàng giữa khu trạng thái khác nhau mà không cập nhật ERP.
- Hàng hoàn chưa kiểm không để chung với hàng bán được.
- Hàng không sử dụng phải tách khu hoặc chuyển Lab theo quy trình.

---

## SOP-17. Kiểm kê cuối ngày / Cycle Count

### Mục tiêu

Kiểm tra hàng tồn kho cuối ngày, phát hiện lệch trước khi kết thúc ca.

### Ai thực hiện

Nhân viên kho, Trưởng kho.

### Thời điểm

Cuối ngày hoặc cuối ca.

### Các bước

1. Vào `Warehouse → Stock Count`.
2. Chọn kho/khu vực cần kiểm.
3. Hệ thống tạo danh sách item cần kiểm hoặc trưởng kho chọn item.
4. Nhân viên đếm thực tế.
5. Nhập số lượng thực tế.
6. Nếu có batch, nhập/scan đúng batch.
7. Ghi chú chênh lệch nếu có.
8. Submit kết quả.
9. Trưởng kho review.
10. Nếu lệch nhỏ theo ngưỡng, tạo adjustment theo quyền.
11. Nếu lệch lớn, freeze khu vực hoặc báo quản lý điều tra.

### Lưu ý

- Không điều chỉnh tồn để “cho khớp” nếu chưa tìm nguyên nhân.
- Lệch liên quan hàng đã bàn giao/hoàn/đóng hàng phải kiểm tra lại manifest, scan event, return receipt.

---

## SOP-18. Đối soát cuối ca / Shift Closing

### Mục tiêu

Khóa ca làm, xác nhận số liệu kho trong ngày, báo cáo quản lý.

### Ai thực hiện

Trưởng kho, nhân viên kho liên quan.

### Điều kiện trước

- Các đơn đã xử lý được cập nhật status.
- Hàng hoàn trong ngày đã được scan/ghi nhận.
- Các phiếu nhập/xuất/chuyển kho đã hoàn tất hoặc ghi rõ đang pending.

### Các bước

1. Vào `Warehouse → Shift Closing`.
2. Chọn kho, ngày, ca.
3. Kiểm tra số liệu tổng:
   - số đơn nhận trong ngày,
   - số đơn đã pick,
   - số đơn đã pack,
   - số đơn đã bàn giao,
   - số đơn chưa xử lý,
   - số hàng nhập,
   - số hàng xuất,
   - số hàng hoàn,
   - số lỗi/lệch.
4. Kiểm tra danh sách exception.
5. Ghi lý do cho các exception.
6. Đính kèm chứng từ/báo cáo nếu cần.
7. Submit cho quản lý hoặc close shift theo quyền.
8. Hệ thống ghi nhận ca đã đóng.

### Không được close shift nếu

- còn manifest bàn giao chưa xác nhận,
- còn đơn đã packed nhưng không rõ đang ở đâu,
- hàng hoàn nhận rồi nhưng chưa scan,
- có lệch tồn nghiêm trọng chưa ghi nhận,
- có phiếu nhập/xuất pending không lý do.

---

# PHẦN E — SOP BÀN GIAO ĐƠN VỊ VẬN CHUYỂN

---

## SOP-19. Chuẩn bị khu vực bàn giao ĐVVC

### Mục tiêu

Sắp xếp đơn chờ giao theo khu vực, thùng/rổ, ĐVVC để giảm thiếu sót.

### Ai thực hiện

Nhân viên kho, Trưởng kho.

### Các bước

1. Vào `Shipping → Carrier Manifest`.
2. Lọc đơn `Packed` theo ĐVVC/kênh/sàn.
3. Tạo manifest/chuyến bàn giao.
4. In hoặc hiển thị danh sách đơn.
5. Phân chia khu vực để hàng theo ĐVVC.
6. Sắp xếp theo thùng/rổ nếu cần.
7. Đảm bảo mỗi thùng/rổ có số lượng dễ kiểm soát.
8. Đánh dấu khu vực/thùng/rổ trên hệ thống nếu có.

### Lưu ý

- Không trộn đơn của các ĐVVC nếu không có nhãn rõ.
- Đơn chưa packed không đưa vào khu bàn giao.

---

## SOP-20. Bàn giao ĐVVC bằng scan manifest

### Mục tiêu

Xác nhận đúng số lượng đơn bàn giao cho ĐVVC, có bằng chứng scan và ký xác nhận.

### Ai thực hiện

Nhân viên kho, ĐVVC, Trưởng kho nếu cần.

### Điều kiện trước

- Đơn đã `Packed`.
- Manifest đã tạo.
- ĐVVC đến nhận hàng.

### Các bước

1. Vào `Shipping → Manifest Handover`.
2. Chọn manifest.
3. Đối chiếu tổng số đơn trên hệ thống với bảng/danh sách giao.
4. Lấy từng đơn/thùng/rổ tại khu vực bàn giao.
5. Quét mã đơn/mã vận đơn trực tiếp khi bàn giao.
6. Hệ thống đánh dấu đơn đã scan.
7. Nếu đủ số lượng, bấm `Confirm Handover`.
8. Đính kèm ảnh/chữ ký/biên bản nếu có.
9. In hoặc xuất biên bản bàn giao nếu cần.
10. Đơn chuyển trạng thái `HandedOver`.

### Checklist trước khi confirm

- Số đơn scanned = số đơn manifest.
- Không có đơn ngoài manifest.
- Không có đơn duplicate.
- ĐVVC xác nhận nhận hàng.
- Có bằng chứng bàn giao nếu policy yêu cầu.

---

## SOP-21. Xử lý thiếu đơn khi bàn giao ĐVVC

### Mục tiêu

Có quy trình rõ khi danh sách bàn giao chưa đủ đơn, tránh giao thiếu hoặc thất lạc.

### Ai thực hiện

Nhân viên kho, Trưởng kho, CSKH/Sales nếu cần.

### Trường hợp

Khi scan manifest, hệ thống báo thiếu đơn hoặc tổng số đơn thực tế không khớp.

### Các bước

1. Dừng confirm handover.
2. Kiểm tra mã đơn/mã vận đơn bị thiếu.
3. Tìm trong khu vực bàn giao.
4. Nếu không thấy, kiểm tra khu vực đóng hàng.
5. Kiểm tra trạng thái đơn trên ERP:
   - đơn đã packed chưa,
   - đơn có nằm trong manifest không,
   - đơn có bị cancel/hold không,
   - mã vận đơn có đúng không.
6. Nếu mã chưa có trên hệ thống hoặc thông tin sai, báo Shipping Admin/Sales Admin.
7. Nếu đơn có trên hệ thống nhưng chưa tìm thấy, báo Trưởng kho mở incident.
8. Chỉ confirm handover khi xử lý xong hoặc quản lý cho phép tách manifest.

### Không được làm

- Không ký bàn giao đủ khi thực tế thiếu.
- Không giao đơn ngoài manifest mà không scan.
- Không tự tạo mã vận đơn ngoài hệ thống.

---

# PHẦN F — SOP HÀNG HOÀN / RETURNS

---

## SOP-22. Nhận hàng hoàn từ shipper/ĐVVC

### Mục tiêu

Ghi nhận hàng hoàn về kho, tránh thất lạc, có bằng chứng tình trạng thực tế.

### Ai thực hiện

Nhân viên kho, CSKH nếu cần.

### Điều kiện trước

- Shipper/ĐVVC giao hàng hoàn.
- Có mã đơn/mã vận đơn hoặc thông tin đối chiếu.

### Các bước

1. Vào `Returns → Return Receiving`.
2. Quét mã vận đơn/mã đơn.
3. Nếu tìm thấy đơn, mở return record.
4. Nếu không tìm thấy, tạo exception `Unknown Return` theo quyền.
5. Nhận hàng từ shipper.
6. Đưa hàng vào khu vực hàng hoàn.
7. Quét xác nhận hàng hoàn.
8. Chụp ảnh/quay video tình trạng bao ngoài nếu cần.
9. Ghi nhận thời điểm nhận và người nhận.
10. Chuyển sang bước kiểm tra hàng hoàn.

### Lưu ý

- Hàng hoàn chưa kiểm không được nhập lại tồn bán được.
- Hàng hoàn phải nằm khu riêng.
- Nếu hàng hoàn không xác định được đơn, phải mở exception.

---

## SOP-23. Kiểm tra tình trạng hàng hoàn

### Mục tiêu

Phân loại hàng hoàn: còn sử dụng được, không sử dụng được, cần QC/QA review.

### Ai thực hiện

Nhân viên kho, QC/QA nếu cần.

### Các bước

1. Vào `Returns → Return Inspection`.
2. Chọn return record.
3. Kiểm tra bao bì ngoài.
4. Kiểm tra tem niêm phong.
5. Kiểm tra tình trạng sản phẩm bên trong.
6. Ghi nhận tình trạng thực tế.
7. Đính kèm ảnh/video nếu policy yêu cầu.
8. Chọn kết quả:
   - Còn sử dụng,
   - Không sử dụng,
   - Cần QC review.
9. Submit kết quả.

### Tiêu chí tham khảo

| Kết quả | Điều kiện |
|---|---|
| Còn sử dụng | Hàng nguyên vẹn, tem/niêm phong đạt, không hư hỏng, không quá hạn, không bị nghi nhiễm bẩn |
| Không sử dụng | Rách/móp nghiêm trọng, mất tem, hư hỏng, bẩn, chảy, vỡ, quá hạn, nghi ngờ không an toàn |
| Cần QC review | Không rõ tình trạng, có dấu hiệu bất thường, hàng giá trị cao, batch có complaint |

---

## SOP-24. Nhập lại hàng hoàn còn sử dụng

### Mục tiêu

Chuyển hàng hoàn đạt điều kiện về kho đúng trạng thái.

### Ai thực hiện

Nhân viên kho, Trưởng kho/QC theo quyền.

### Điều kiện trước

- Return inspection kết luận `Còn sử dụng` hoặc QC pass.

### Các bước

1. Mở return record đã kiểm.
2. Chọn action `Move to Sellable Stock` hoặc `Return to Stock`.
3. Chọn kho/vị trí nhận.
4. Xác nhận SKU/batch/số lượng.
5. Submit.
6. Hệ thống tạo stock movement.
7. Lưu chứng từ nhập kho hàng hoàn.

### Lưu ý

- Nếu không chắc còn bán được, không chuyển vào sellable stock.
- Hàng hoàn có batch phải nhập đúng batch.

---

## SOP-25. Xử lý hàng hoàn không sử dụng

### Mục tiêu

Tách hàng không sử dụng khỏi hàng bán được, chuyển đúng khu/Lab/hàng hỏng.

### Ai thực hiện

Nhân viên kho, QC/QA, Trưởng kho.

### Điều kiện trước

- Return inspection kết luận `Không sử dụng`.

### Các bước

1. Mở return record.
2. Chọn action `Move to Non-usable/Lab/Damaged`.
3. Chọn lý do.
4. Đính kèm ảnh/video nếu cần.
5. Chọn khu vực/kho nhận: Lab, hàng hỏng, quarantine.
6. Submit.
7. Hệ thống tạo movement và khóa hàng khỏi sellable stock.
8. Nếu cần tiêu hủy, tạo request riêng theo policy.

### Lưu ý

- Không để hàng không sử dụng chung với hàng bán được.
- Nếu có dấu hiệu batch lỗi, báo QC mở investigation.

---

# PHẦN G — SOP XUẤT KHO NỘI BỘ, CHUYỂN KHO, ĐIỀU CHỈNH

---

## SOP-26. Tạo phiếu xuất kho nội bộ

### Mục tiêu

Xuất hàng/NVL/bao bì cho mục đích nội bộ hoặc sản xuất/gia công theo chứng từ rõ ràng.

### Ai thực hiện

Kho, Purchasing/Production, người yêu cầu.

### Các bước

1. Vào `Warehouse → Stock Issue`.
2. Bấm `Create Issue`.
3. Chọn loại xuất:
   - xuất bán,
   - xuất sản xuất,
   - xuất gia công,
   - xuất sample/tester,
   - xuất hủy/hỏng,
   - khác.
4. Chọn người/bộ phận nhận.
5. Nhập item, batch, số lượng.
6. Đính kèm chứng từ yêu cầu.
7. Submit duyệt nếu cần.
8. Sau khi duyệt, kho xuất hàng thực tế.
9. Người nhận ký/xác nhận nhận hàng.
10. Trưởng kho lưu chứng từ.

### Lưu ý

- Phiếu xuất phải có lý do rõ.
- Xuất gia công phải link với subcontract order.
- Xuất sample/tester nên có campaign/người nhận nếu có.

---

## SOP-27. Chuyển kho/chuyển vị trí

### Mục tiêu

Ghi nhận việc chuyển hàng giữa kho/khu/vị trí, tránh lệch tồn.

### Ai thực hiện

Nhân viên kho, Trưởng kho.

### Các bước

1. Vào `Warehouse → Transfer`.
2. Chọn kho/vị trí nguồn.
3. Chọn kho/vị trí đích.
4. Nhập SKU/batch/số lượng.
5. Submit.
6. Xuất hàng khỏi vị trí nguồn.
7. Nhận hàng tại vị trí đích.
8. Xác nhận hoàn tất.

### Lưu ý

- Nếu chuyển giữa trạng thái hàng khác nhau, phải dùng đúng loại movement.
- Không dùng chuyển kho để che lỗi thất thoát.

---

## SOP-28. Điều chỉnh tồn kho

### Mục tiêu

Xử lý lệch tồn có kiểm soát, có lý do, có duyệt.

### Ai thực hiện

Trưởng kho, Finance/QC/Quản lý theo rule.

### Điều kiện trước

- Đã kiểm kê hoặc điều tra lý do lệch.
- Có chứng từ/bằng chứng nếu cần.

### Các bước

1. Vào `Warehouse → Stock Adjustment`.
2. Chọn item/batch/kho.
3. Nhập số lượng điều chỉnh tăng/giảm.
4. Chọn lý do.
5. Ghi mô tả chi tiết.
6. Đính kèm biên bản/ảnh nếu có.
7. Submit duyệt.
8. Người duyệt kiểm tra.
9. Sau khi approved, hệ thống tạo stock movement adjustment.

### Không được làm

- Không điều chỉnh tồn để “khớp sổ” nếu chưa có lý do.
- Không dùng adjustment thay cho nhập/xuất/chuyển kho hợp lệ.

---

# PHẦN H — SOP SẢN XUẤT/GIA CÔNG NGOÀI

---

## SOP-29. Tạo đơn gia công với nhà máy

### Mục tiêu

Quản lý quá trình đặt nhà máy sản xuất/gia công từ lúc lên đơn tới nhận hàng.

### Ai thực hiện

Purchasing/Production, R&D, Finance, Kho, QC.

### Điều kiện trước

- Có SKU/công thức/spec được duyệt.
- Có nhà máy active.
- Có số lượng/quy cách/mẫu mã cần đặt.

### Các bước

1. Vào `Production/Subcontract → Subcontract Orders`.
2. Bấm `Create`.
3. Chọn nhà máy.
4. Chọn SKU/thành phẩm cần gia công.
5. Nhập số lượng.
6. Nhập quy cách, mẫu mã.
7. Nhập thời gian dự kiến sản xuất/nhận hàng.
8. Nhập điều khoản cọc/thanh toán.
9. Đính kèm spec/brief/hợp đồng nếu có.
10. Submit duyệt.
11. Sau khi approved, tạo kế hoạch chuyển NVL/bao bì nếu công ty cung cấp.

### Lưu ý

- Không đặt gia công nếu chưa chốt spec/mẫu mã.
- Nếu cần cọc đơn, Finance phải ghi nhận cọc.

---

## SOP-30. Chuyển NVL/bao bì cho nhà máy gia công

### Mục tiêu

Kiểm soát hàng công ty chuyển sang nhà máy, tránh thất thoát NVL/bao bì.

### Ai thực hiện

Kho, Purchasing/Production, nhà máy nhận.

### Điều kiện trước

- Subcontract Order đã approved.
- Có danh sách NVL/bao bì cần chuyển.
- Hàng trong kho đủ và được phép xuất.

### Các bước

1. Mở Subcontract Order.
2. Chọn `Create Material Transfer to Factory`.
3. Hệ thống gợi ý NVL/bao bì theo BOM/spec nếu có.
4. Kho kiểm tra tồn khả dụng.
5. Chọn batch/số lượng thực xuất.
6. Đính kèm giấy tờ nếu cần: COA, MSDS, tem phụ, hóa đơn VAT, biên bản.
7. Submit duyệt nếu cần.
8. Kho xuất hàng.
9. Nhà máy/người nhận ký biên bản bàn giao.
10. ERP cập nhật trạng thái `Sent to Factory` hoặc `Partially Sent`.

### Lưu ý

- NVL/bao bì đã chuyển sang nhà máy phải được theo dõi như tồn tại bên thứ ba.
- Nếu nhà máy trả lại thừa/thiếu/lỗi, phải có chứng từ riêng.

---

## SOP-31. Làm mẫu và chốt mẫu

### Mục tiêu

Đảm bảo mẫu trước sản xuất hàng loạt được kiểm tra, phê duyệt, lưu mẫu.

### Ai thực hiện

R&D, QC/QA, Production/Purchasing, nhà máy.

### Các bước

1. Mở Subcontract Order.
2. Chọn tab `Sample Approval`.
3. Ghi nhận ngày nhận mẫu.
4. Đính kèm ảnh/video/tài liệu mẫu.
5. R&D/QC kiểm tra mẫu.
6. Chọn kết quả:
   - Approved,
   - Rework Required,
   - Rejected.
7. Nếu approved, ghi rõ version mẫu được chốt.
8. Lưu mẫu vật lý theo quy định và ghi vị trí lưu mẫu trên ERP nếu có.
9. Chỉ khi mẫu approved mới cho chuyển sang sản xuất hàng loạt.

### Lưu ý

- Không sản xuất hàng loạt khi mẫu chưa chốt.
- Nếu thay đổi mẫu/spec, phải tạo version hoặc change request.

---

## SOP-32. Theo dõi sản xuất hàng loạt tại nhà máy

### Mục tiêu

Theo dõi trạng thái đơn gia công, hạn chế trễ tiến độ hoặc sai spec.

### Ai thực hiện

Production/Purchasing.

### Các bước

1. Vào `Production/Subcontract → Order Tracking`.
2. Kiểm tra trạng thái:
   - Waiting Deposit,
   - Materials Sent,
   - Sample Pending,
   - Sample Approved,
   - In Production,
   - Ready to Deliver,
   - Delivered,
   - Closed.
3. Cập nhật mốc thời gian thực tế.
4. Ghi chú trao đổi với nhà máy.
5. Nếu có trễ tiến độ, tạo risk/issue.
6. Báo quản lý nếu trễ ảnh hưởng kế hoạch bán hàng.

---

## SOP-33. Nhận hàng gia công về kho

### Mục tiêu

Nhận hàng từ nhà máy, kiểm số lượng/chất lượng, quyết định nhập kho hoặc claim.

### Ai thực hiện

Kho, QC/QA, Production/Purchasing.

### Điều kiện trước

- Nhà máy giao hàng về kho.
- Có Subcontract Order liên quan.

### Các bước

1. Vào `Warehouse → Receiving` hoặc `Subcontract → Receive Finished Goods`.
2. Chọn Subcontract Order.
3. Nhập số lượng nhận thực tế.
4. Nhập batch/lô, ngày sản xuất, hạn dùng nếu có.
5. Kiểm tra bao bì, nhãn, số lượng kiện.
6. QC kiểm tra chất lượng theo checklist.
7. Nếu đạt: nhập kho thành phẩm hoặc khu chờ release.
8. Nếu không đạt: chọn `Factory Claim / Reject`.
9. Ghi rõ lỗi và đính kèm ảnh/chứng từ.
10. Nếu có quy định phản hồi trong 3–7 ngày, hệ thống tạo deadline claim.

### Lưu ý

- Không nhập khả dụng hàng gia công chưa QC.
- Nếu hàng không đạt, phải báo nhà máy trong thời hạn quy định.

---

## SOP-34. Claim lỗi với nhà máy trong 3–7 ngày

### Mục tiêu

Ghi nhận và theo dõi lỗi hàng gia công để phản hồi nhà máy đúng hạn.

### Ai thực hiện

QC/QA, Production/Purchasing, quản lý.

### Điều kiện trước

- Hàng gia công nhận về có lỗi hoặc không đạt.

### Các bước

1. Mở Subcontract Order hoặc QC record.
2. Chọn `Create Factory Claim`.
3. Nhập loại lỗi:
   - sai số lượng,
   - sai quy cách,
   - lỗi bao bì,
   - lỗi chất lượng,
   - lỗi nhãn,
   - khác.
4. Nhập mô tả lỗi.
5. Đính kèm ảnh/video/biên bản.
6. Nhập deadline phản hồi nhà máy.
7. Submit claim.
8. Theo dõi trạng thái:
   - Open,
   - Sent to Factory,
   - Factory Responded,
   - Resolved,
   - Rejected,
   - Closed.
9. Cập nhật kết quả xử lý: đổi hàng, bù hàng, trừ tiền, tái sản xuất, chấp nhận có điều kiện.

### Lưu ý

- Claim không có bằng chứng rất khó xử lý.
- Nếu quá hạn 3–7 ngày, cần lý do và quản lý duyệt.

---

## SOP-35. Thanh toán cọc và thanh toán cuối cho nhà máy

### Mục tiêu

Kiểm soát thanh toán theo tiến độ gia công, tránh thanh toán khi chưa nghiệm thu.

### Ai thực hiện

Finance, Purchasing/Production, Quản lý.

### Các bước thanh toán cọc

1. Mở Subcontract Order.
2. Kiểm tra điều khoản cọc.
3. Tạo payment request.
4. Đính kèm hợp đồng/báo giá/đơn đặt.
5. Submit duyệt.
6. Finance thanh toán và ghi nhận.

### Các bước thanh toán cuối

1. Kiểm tra hàng đã nhận/QC/claim.
2. Kiểm tra số lượng nghiệm thu.
3. Kiểm tra khoản đã cọc.
4. Tạo payment request final.
5. Nếu còn claim mở, hệ thống cảnh báo.
6. Quản lý/Finance duyệt.
7. Thanh toán và đóng đơn gia công.

### Lưu ý

- Không thanh toán cuối khi hàng chưa nghiệm thu hoặc claim chưa xử lý, trừ khi có duyệt ngoại lệ.

---

# PHẦN I — SOP SALES ORDER, GIAO HÀNG, CÔNG NỢ CƠ BẢN

---

## SOP-36. Tạo đơn hàng bán

### Mục tiêu

Tạo đơn đúng khách, đúng kênh, đúng giá, đúng hàng, đủ thông tin giao.

### Ai thực hiện

Sales Admin, CSKH, POS/Online Admin theo quyền.

### Các bước

1. Vào `Sales → Sales Orders`.
2. Bấm `Create Order`.
3. Chọn khách hàng/kênh bán.
4. Nhập thông tin giao hàng.
5. Chọn SKU/số lượng.
6. Hệ thống tính giá/chiết khấu theo rule.
7. Kiểm tra quà tặng/combo nếu có.
8. Kiểm tra tồn khả dụng.
9. Submit xác nhận.
10. Nếu vượt discount/quyền, gửi duyệt.
11. Đơn confirmed chuyển sang kho xử lý.

### Lưu ý

- Không nhập địa chỉ thiếu thông tin.
- Không tự giảm giá ngoài rule.
- Không tạo đơn nếu chưa rõ kênh hoặc phương thức thanh toán.

---

## SOP-37. Hủy đơn hoặc chỉnh đơn

### Mục tiêu

Kiểm soát thay đổi đơn hàng trước/sau khi kho xử lý.

### Ai thực hiện

Sales Admin/CSKH, quản lý theo quyền.

### Quy tắc

- Đơn chưa pick: có thể chỉnh/hủy theo quyền.
- Đơn đang pick/packed: phải báo kho và có action chính thức.
- Đơn đã handed over: không hủy như đơn mới, phải xử lý theo return/cancel shipping.

### Các bước

1. Mở đơn hàng.
2. Kiểm tra trạng thái.
3. Chọn `Cancel` hoặc `Request Change`.
4. Nhập lý do.
5. Nếu đơn đã vào kho, hệ thống tạo thông báo cho kho.
6. Nếu hàng đã reserve/pick, hệ thống release hoặc đảo movement theo rule.
7. Lưu audit log.

---

# PHẦN J — SOP PHÊ DUYỆT VÀ KIỂM SOÁT

---

## SOP-38. Xử lý yêu cầu chờ duyệt

### Mục tiêu

Đảm bảo người duyệt kiểm tra đúng thông tin trước khi approve/reject.

### Ai thực hiện

Trưởng bộ phận, Finance, QA, COO/CEO theo matrix.

### Các bước

1. Vào `My Approvals`.
2. Lọc yêu cầu pending.
3. Mở chi tiết chứng từ.
4. Kiểm tra:
   - người tạo,
   - lý do,
   - số lượng/giá trị,
   - chứng từ đính kèm,
   - rủi ro batch/tồn/kho/tiền.
5. Nếu đúng, bấm `Approve`.
6. Nếu sai, bấm `Reject` và ghi rõ lý do.
7. Nếu cần bổ sung, bấm `Request More Info` nếu hệ thống hỗ trợ.

### Lưu ý

- Không duyệt chỉ vì quen người tạo.
- Duyệt là chịu trách nhiệm kiểm soát.
- Tất cả approve/reject phải có log.

---

## SOP-39. Xem audit log chứng từ

### Mục tiêu

Kiểm tra lịch sử thay đổi khi có tranh chấp hoặc lỗi.

### Ai thực hiện

Quản lý, QA, Finance, Admin theo quyền.

### Các bước

1. Mở chứng từ.
2. Chọn tab `Audit Log`.
3. Kiểm tra các thay đổi:
   - ai sửa,
   - sửa lúc nào,
   - trường nào thay đổi,
   - giá trị trước/sau,
   - lý do nếu có.
4. Nếu phát hiện thay đổi bất thường, tạo incident hoặc báo quản lý.

---

# PHẦN K — SOP BÁO CÁO VÀ DASHBOARD

---

## SOP-40. Xem dashboard kho hằng ngày

### Mục tiêu

Giúp trưởng kho/quản lý nắm tình hình vận hành trong ngày.

### Ai thực hiện

Trưởng kho, COO/CEO.

### Các chỉ số cần xem

- Số đơn nhận trong ngày.
- Số đơn đã pick.
- Số đơn đã pack.
- Số đơn đã bàn giao.
- Số đơn pending/exception.
- Hàng hoàn nhận trong ngày.
- Lệch tồn phát hiện.
- Phiếu nhập/xuất chưa hoàn tất.
- Batch hold/fail.

### Cách thao tác

1. Vào `Dashboard → Warehouse Daily`.
2. Chọn ngày/kho.
3. Kiểm tra các chỉ số tổng.
4. Click drill-down vào chỉ số bất thường.
5. Ghi nhận action cần xử lý.

---

## SOP-41. Export báo cáo theo quyền

### Mục tiêu

Cho phép xuất dữ liệu phục vụ quản lý nhưng vẫn kiểm soát thông tin nhạy cảm.

### Ai thực hiện

Người dùng có quyền export.

### Các bước

1. Vào màn hình báo cáo.
2. Chọn filter rõ ràng.
3. Bấm `Export`.
4. Chọn định dạng nếu có: CSV/XLSX/PDF.
5. Hệ thống ghi audit log export.
6. Không gửi file chứa dữ liệu nhạy cảm cho người không có quyền.

### Lưu ý

- Giá vốn, công nợ, lương, payout, dữ liệu khách hàng là dữ liệu nhạy cảm.
- Export phải được kiểm soát.

---

# PHẦN L — HƯỚNG DẪN THEO VAI TRÒ

---

## 8. Checklist đào tạo cho Nhân viên kho

Nhân viên kho phải biết làm:

- Đăng nhập ERP.
- Xem Daily Board.
- Tìm đơn hàng.
- Soạn hàng/picking.
- Đóng hàng/packing.
- Tạo hoặc xử lý bàn giao ĐVVC.
- Scan manifest.
- Xử lý thiếu đơn khi bàn giao.
- Nhận hàng hoàn.
- Kiểm tra hàng hoàn cơ bản.
- Nhập/xuất kho theo phiếu.
- Kiểm kê cuối ngày.
- Báo lỗi/lệch cho trưởng kho.

Không yêu cầu nhân viên kho biết:

- sửa giá,
- duyệt discount,
- xem giá vốn,
- sửa master data quan trọng,
- duyệt thanh toán.

---

## 9. Checklist đào tạo cho Trưởng kho

Trưởng kho phải biết:

- Phân công Daily Board.
- Review task tồn đọng.
- Duyệt/xác nhận phiếu kho theo quyền.
- Kiểm tra stock count.
- Đối soát cuối ca.
- Điều tra lệch tồn.
- Xử lý manifest thiếu đơn.
- Kiểm soát khu hàng hoàn/hàng lỗi.
- Đọc báo cáo kho.
- Xem audit log cơ bản.

---

## 10. Checklist đào tạo cho QC/QA

QC/QA phải biết:

- Xem danh sách hàng chờ QC.
- Thực hiện inbound inspection.
- Set batch hold/pass/fail.
- Xử lý hàng hoàn cần QC review.
- Kiểm tra hàng gia công về kho.
- Tạo factory claim nếu lỗi.
- Đính kèm ảnh/chứng từ.
- Không release batch nếu chưa đủ điều kiện.

---

## 11. Checklist đào tạo cho Sales Admin/CSKH

Sales Admin/CSKH phải biết:

- Tạo đơn hàng.
- Kiểm tra tồn khả dụng.
- Theo dõi trạng thái đơn.
- Xử lý yêu cầu hủy/chỉnh đơn.
- Tra trạng thái bàn giao/giao hàng.
- Tạo/tra return case.
- Ghi nhận complaint khách hàng.
- Không tự can thiệp tồn kho.

---

## 12. Checklist đào tạo cho Purchasing/Production

Purchasing/Production phải biết:

- Tạo PR/PO.
- Tạo subcontract order.
- Theo dõi cọc và thời gian giao.
- Tạo yêu cầu chuyển NVL/bao bì cho nhà máy.
- Theo dõi sample approval.
- Theo dõi trạng thái sản xuất hàng loạt.
- Phối hợp nhận hàng gia công.
- Tạo/theo dõi claim nhà máy.

---

## 13. Checklist đào tạo cho Finance

Finance phải biết:

- Xem PO/subcontract payment terms.
- Xử lý payment request cọc/cuối.
- Kiểm tra claim còn mở trước thanh toán cuối.
- Đối soát COD/bank nếu nằm trong Phase 1.
- Xem audit/payment history.
- Không sửa kho/QC thay bộ phận nghiệp vụ.

---

## 14. Checklist đào tạo cho Quản lý

Quản lý phải biết:

- Xem dashboard vận hành.
- Duyệt yêu cầu trong My Approvals.
- Xem exception.
- Xem audit log.
- Theo dõi KPI kho/giao hàng/QC/gia công.
- Ra quyết định khi có lệch tồn, batch fail, thiếu đơn, hàng gia công lỗi.

---

# PHẦN M — LỖI THƯỜNG GẶP VÀ CÁCH XỬ LÝ

---

## 15. Lỗi kho và giao hàng

| Tình huống | Không nên làm | Cách xử lý đúng |
|---|---|---|
| Đơn thiếu hàng khi picking | Lấy sản phẩm khác thay thế | Report shortage, báo trưởng kho/Sales |
| Scan mã không nhận | Nhập tay bừa | Kiểm tra mã, trạng thái đơn, báo support nếu lỗi hệ thống |
| Đơn đã packed nhưng không thấy ở khu bàn giao | Ký giao đủ | Tìm khu đóng hàng, kiểm trạng thái ERP, mở exception |
| Manifest thiếu đơn | Confirm cho xong | Dừng bàn giao, xử lý SOP-21 |
| Hàng hoàn không có mã đơn | Nhập kho bán được | Tạo Unknown Return/exception |
| Hàng hoàn hỏng | Để lại kệ bán | Chuyển non-usable/Lab/damaged |

---

## 16. Lỗi QC/batch

| Tình huống | Không nên làm | Cách xử lý đúng |
|---|---|---|
| Batch chưa QC nhưng sales cần gấp | Mở bán tạm | Báo QA/Quản lý, chỉ release khi đủ điều kiện |
| Nhập sai hạn dùng | Sửa tay nếu không có quyền | Tạo yêu cầu chỉnh dữ liệu, audit rõ lý do |
| Batch fail nhưng kho vẫn thấy hàng | Xuất hàng | Báo support/QA ngay, freeze batch nếu cần |
| Complaint liên quan batch | Xử lý như khiếu nại thường | Link complaint với batch, báo QA |

---

## 17. Lỗi gia công ngoài

| Tình huống | Không nên làm | Cách xử lý đúng |
|---|---|---|
| Nhà máy giao thiếu | Nhập đủ theo đơn | Nhập số thực nhận, tạo claim |
| Hàng sai quy cách | Nhập kho bán được | QC fail/hold, tạo factory claim |
| Mẫu chưa duyệt nhưng nhà máy sản xuất | Nhận bình thường | Báo quản lý, mở incident/claim |
| Quá hạn claim 3–7 ngày | Im lặng | Ghi lý do, escalate quản lý |
| Thanh toán cuối khi còn lỗi | Thanh toán cho xong | Cảnh báo Finance, chờ duyệt ngoại lệ |

---

# PHẦN N — BÀI TẬP TRAINING/UAT MINI

---

## 18. Bài tập cho kho: xử lý đơn end-to-end

### Mục tiêu

Đào tạo nhân viên kho xử lý một đơn từ lúc sẵn sàng soạn tới bàn giao ĐVVC.

### Dữ liệu mẫu

- Đơn: SO-DEMO-001.
- SKU: SERUM-VITC-30ML.
- Số lượng: 2.
- ĐVVC: GHTK/GHN/ViettelPost tùy demo.

### Bài tập

1. Mở Daily Board.
2. Chọn đơn chờ pick.
3. Pick hàng đúng SKU/batch.
4. Pack đơn.
5. Đưa vào manifest.
6. Scan bàn giao.
7. Confirm handover.
8. Kiểm tra status đơn.

### Đạt khi

- Đơn chuyển đúng trạng thái.
- Không tạo stock movement sai.
- Có audit log.
- Không bỏ qua scan bàn giao.

---

## 19. Bài tập cho kho: thiếu đơn khi bàn giao

### Mục tiêu

Đào tạo xử lý exception đúng cách.

### Bài tập

1. Tạo manifest có 5 đơn.
2. Chỉ scan được 4 đơn.
3. Hệ thống báo thiếu 1 đơn.
4. Nhân viên tìm lại khu đóng hàng.
5. Nếu tìm thấy, scan bổ sung.
6. Nếu không tìm thấy, tạo exception và báo trưởng kho.

### Đạt khi

- Không confirm thiếu.
- Có ghi nhận lý do.
- Có người chịu trách nhiệm xử lý.

---

## 20. Bài tập hàng hoàn

### Mục tiêu

Đào tạo nhận hàng hoàn, kiểm tình trạng, phân loại.

### Bài tập

1. Quét mã vận đơn hoàn.
2. Nhận hàng vào khu hàng hoàn.
3. Ghi ảnh/tình trạng.
4. Kiểm tem/bao bì.
5. Chọn còn sử dụng hoặc không sử dụng.
6. Nếu còn sử dụng: chuyển kho.
7. Nếu không sử dụng: chuyển Lab/hàng hỏng.

### Đạt khi

- Hàng hoàn không tự động nhập sellable stock.
- Có record tình trạng.
- Có movement đúng.

---

## 21. Bài tập gia công ngoài

### Mục tiêu

Đào tạo quy trình đặt nhà máy, chuyển NVL/bao bì, duyệt mẫu, nhận hàng.

### Bài tập

1. Tạo subcontract order.
2. Ghi số lượng/quy cách/mẫu mã.
3. Tạo payment request cọc.
4. Tạo transfer NVL/bao bì cho nhà máy.
5. Ghi nhận sample received.
6. Approve sample.
7. Nhận thành phẩm về kho.
8. QC pass hoặc fail.
9. Nếu fail, tạo factory claim.
10. Nếu pass, nhập kho.

### Đạt khi

- Có trace từ subcontract order tới transfer, sample, receipt, QC.
- Không thanh toán cuối nếu còn claim mở.

---

# PHẦN O — QUY TẮC KÝ XÁC NHẬN ĐÀO TẠO

---

## 22. Training sign-off

Mỗi user sau khi đào tạo phải được ghi nhận:

- họ tên,
- phòng ban,
- vai trò ERP,
- module đã học,
- bài tập đã hoàn thành,
- ngày đào tạo,
- người đào tạo,
- trạng thái đạt/chưa đạt,
- ghi chú cần đào tạo lại.

### Mẫu bảng sign-off

| Họ tên | Vai trò | Module | Bài tập | Kết quả | Người đào tạo | Ngày |
|---|---|---|---|---|---|---|
|  | Kho | Picking/Packing/Handover | SO end-to-end | Đạt/Chưa đạt |  |  |
|  | QC | Inbound QC/Batch status | QC pass/fail | Đạt/Chưa đạt |  |  |
|  | Production | Subcontract | Gia công end-to-end | Đạt/Chưa đạt |  |  |

---

# PHẦN P — QUICK REFERENCE

---

## 23. Luồng kho hằng ngày rút gọn

```text
Mở Daily Board
→ Tiếp nhận đơn trong ngày
→ Pick hàng
→ Pack hàng
→ Phân khu/chia thùng/rổ theo ĐVVC
→ Scan manifest bàn giao
→ Nhận hàng hoàn nếu có
→ Sắp xếp lại kho
→ Kiểm kê cuối ngày
→ Đối soát số liệu
→ Close shift
```

---

## 24. Luồng nhập kho rút gọn

```text
Nhận chứng từ giao hàng
→ Mở Receiving
→ Chọn PO/Subcontract Order
→ Kiểm số lượng/bao bì/lô
→ Nhập batch/hạn dùng
→ Chờ QC
→ QC pass
→ Putaway
→ Tồn khả dụng
```

Nếu không đạt:

```text
Không đạt
→ Reject/Hold
→ Đính kèm bằng chứng
→ Báo NCC/nhà máy
→ Không nhập sellable stock
```

---

## 25. Luồng hàng hoàn rút gọn

```text
Nhận hàng từ shipper
→ Đưa vào khu hàng hoàn
→ Quét hàng hoàn
→ Ghi nhận tình trạng thực tế
→ Kiểm tem/bao bì/hư hỏng
→ Còn sử dụng: chuyển kho
→ Không sử dụng: chuyển Lab/hàng hỏng
→ Lưu chứng từ/ảnh/video
```

---

## 26. Luồng gia công ngoài rút gọn

```text
Tạo đơn với nhà máy
→ Xác nhận số lượng/quy cách/mẫu mã
→ Cọc đơn và chốt thời gian
→ Chuyển NVL/bao bì cho nhà máy
→ Ký biên bản bàn giao
→ Làm mẫu
→ Chốt mẫu và lưu mẫu
→ Sản xuất hàng loạt
→ Giao hàng về kho
→ Kiểm số lượng/chất lượng
→ Đạt: nhập kho
→ Không đạt: claim nhà máy trong 3–7 ngày
→ Thanh toán cuối
```

---

# PHẦN Q — NGUYÊN TẮC VÀNG CHO USER

---

## 27. 12 nguyên tắc phải nhớ

1. Không dùng tài khoản người khác.
2. Không sửa tồn kho trực tiếp.
3. Không xuất hàng chưa QC pass nếu policy yêu cầu QC.
4. Không bán batch hold/fail.
5. Không ký bàn giao ĐVVC khi thiếu đơn.
6. Không nhập hàng hoàn vào tồn bán được nếu chưa kiểm.
7. Không trộn hàng hoàn/hàng lỗi với hàng bán được.
8. Không tạo SKU/material trùng.
9. Không duyệt chứng từ khi chưa kiểm chứng từ đính kèm.
10. Không thanh toán cuối cho nhà máy nếu hàng chưa nghiệm thu hoặc claim chưa xử lý, trừ khi có duyệt ngoại lệ.
11. Lỗi hệ thống phải báo theo mẫu, không đi đường vòng.
12. Mọi exception phải có lý do và người chịu trách nhiệm.

---

## 28. Definition of Done cho đào tạo user

Một user được coi là sẵn sàng dùng ERP khi:

- đã tham gia buổi đào tạo theo vai trò,
- đăng nhập được bằng tài khoản cá nhân,
- hiểu menu/module mình được dùng,
- hoàn thành bài tập thực hành tối thiểu,
- biết cách báo lỗi,
- biết các lỗi không được tự xử lý,
- ký xác nhận training sign-off.

---

# PHẦN R — KẾ HOẠCH ĐÀO TẠO GỢI Ý

---

## 29. Lịch đào tạo đề xuất

### Ngày 1 — Tổng quan ERP

Đối tượng: toàn bộ key users.

Nội dung:

- ERP là nguồn dữ liệu thật.
- Tổng quan Phase 1.
- Vai trò và phân quyền.
- Quy tắc audit/chứng từ.
- Cách báo lỗi.

### Ngày 2 — Kho

Đối tượng: trưởng kho, nhân viên kho.

Nội dung:

- Daily Board.
- Picking/Packing.
- Manifest/Handover.
- Returns.
- Stock count/Shift closing.

### Ngày 3 — QC/QA + nhập kho

Đối tượng: QC/QA, kho, purchasing.

Nội dung:

- Receiving.
- Inbound QC.
- Batch status.
- Hold/pass/fail.
- Hàng lỗi/hàng hold.

### Ngày 4 — Sales/CSKH + shipping

Đối tượng: sales admin, CSKH, kho.

Nội dung:

- Sales order.
- Tồn khả dụng.
- Hủy/chỉnh đơn.
- Theo dõi giao hàng.
- Hàng hoàn/complaint.

### Ngày 5 — Gia công ngoài + finance

Đối tượng: production/purchasing, finance, QC, kho.

Nội dung:

- Subcontract order.
- Cọc đơn.
- Chuyển NVL/bao bì.
- Duyệt mẫu.
- Nhận hàng.
- Claim nhà máy.
- Thanh toán cuối.

### Ngày 6 — UAT end-to-end

Đối tượng: key users.

Nội dung:

- chạy thử toàn bộ flow.
- ghi lỗi.
- chốt readiness.

---

# PHẦN S — KẾT LUẬN

SOP này là lớp cầu nối giữa tài liệu thiết kế và vận hành thật.

ERP chỉ sống khi người dùng thao tác đúng. Ở Phase 1, các điểm phải giữ cực chặt là:

- kho hằng ngày,
- nhập/xuất,
- QC/batch,
- đóng hàng,
- scan bàn giao ĐVVC,
- hàng hoàn,
- gia công ngoài,
- kiểm kê và đối soát cuối ca.

Một hệ thống tốt không phải hệ thống không bao giờ có lỗi. Hệ thống tốt là khi lỗi xảy ra, mọi người biết phải làm gì, ghi nhận ở đâu, ai chịu trách nhiệm, và dữ liệu không bị bóp méo.

---

**End of Document**
