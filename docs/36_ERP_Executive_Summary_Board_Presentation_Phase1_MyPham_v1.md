# 36 ERP Executive Summary & Board Presentation — Phase 1 — Mỹ Phẩm

**Tên file:** `36_ERP_Executive_Summary_Board_Presentation_Phase1_MyPham_v1.md`  
**Phiên bản:** v1.0  
**Mục đích:** Tóm tắt toàn bộ chương trình ERP Phase 1 thành tài liệu trình bày cho CEO, ban lãnh đạo, team nội bộ, PM/BA/dev/vendor.  
**Audience:** CEO, COO, CFO/Finance, Head of Warehouse, Head of Sales, QA/QC, Production/Subcontract Owner, PM, BA, Tech Lead, Vendor.  
**Trạng thái:** Draft for leadership alignment and implementation kickoff.

---

## 1. Executive Summary

Công ty đang cần một hệ thống ERP web để hợp nhất các luồng vận hành trọng yếu: **kho, mua hàng, QC, sản xuất/gia công ngoài, bán hàng, giao hàng, hàng hoàn, dữ liệu, phân quyền, báo cáo và kiểm soát rủi ro**.

Phase 1 không nhắm tới việc làm “full ERP khổng lồ”. Phase 1 tập trung vào **lõi sống còn**: hàng hóa, batch/lô, tồn kho, đơn hàng, bàn giao vận chuyển, hàng hoàn, sản xuất/gia công ngoài, audit log và dữ liệu vận hành chuẩn.

Sau khi đối chiếu với workflow thực tế, dự án cần bám mạnh vào 4 điểm vận hành đặc thù:

1. Kho đang làm việc theo nhịp ngày: tiếp nhận đơn trong ngày, xuất/nhập, soạn/đóng, tối ưu vị trí, kiểm kê cuối ngày, đối soát và kết thúc ca.
2. Nội quy kho hiện có 4 luồng rõ: nhập kho, xuất kho, đóng hàng, xử lý hàng hoàn.
3. Bàn giao ĐVVC cần phân khu, để hàng theo thùng/rổ, đối chiếu số lượng, quét mã, xác nhận bàn giao hoặc xử lý thiếu đơn.
4. Sản xuất hiện có nhánh gia công ngoài: lên đơn với nhà máy, chuyển NVL/bao bì, duyệt mẫu, sản xuất hàng loạt, nhận hàng, kiểm tra, nhập kho hoặc claim nhà máy trong 3–7 ngày.

Dự án đã được thiết kế với backend **Go**, frontend **React/Next.js**, database **PostgreSQL**, kiến trúc **Modular Monolith**, API theo **REST + OpenAPI**, stock ledger bất biến, RBAC, approval, audit log, CI/CD và bộ tài liệu triển khai từ business tới engineering.

**Thông điệp cho ban lãnh đạo:**  
ERP Phase 1 không phải là dự án “mua phần mềm”. Đây là dự án **chuẩn hóa cách công ty vận hành**, giảm lệch kho, giảm thất thoát, tăng tốc xử lý đơn, kiểm soát batch/hạn dùng, và tạo nền để mở rộng CRM, HRM, KOL, Finance nâng cao ở Phase 2.

---

## 2. One-Page Board Summary

### 2.1. Vấn đề hiện tại

- Dữ liệu phân tán giữa quy trình giấy, file, thao tác thủ công và kinh nghiệm cá nhân.
- Kho có nhiều bước vận hành thực tế nhưng chưa chắc được hệ thống hóa: đóng ca, kiểm kê, đối soát, hàng hoàn, bàn giao ĐVVC.
- Sản xuất/gia công ngoài có nhiều điểm rủi ro: chuyển NVL/bao bì, duyệt mẫu, kiểm hàng, claim nhà máy.
- Nếu không có ERP chuẩn, dữ liệu tồn kho, batch, đơn hàng, hoàn hàng, giao hàng và công nợ dễ lệch.
- Doanh nghiệp có rủi ro “tăng trưởng nhưng vận hành không theo kịp”.

### 2.2. Mục tiêu Phase 1

- Chuẩn hóa dữ liệu gốc.
- Kiểm soát tồn kho theo batch/lô, hạn dùng, trạng thái QC.
- Chuẩn hóa nhập/xuất/kiểm kê/đối soát kho.
- Quản lý đơn hàng, pick/pack, bàn giao ĐVVC bằng scan/manifest.
- Quản lý hàng hoàn và phân loại tình trạng.
- Quản lý gia công ngoài: chuyển NVL/bao bì, duyệt mẫu, nhận hàng, claim nhà máy.
- Có dashboard vận hành và audit log.
- Tạo nền cho Phase 2: CRM, HRM, KOL, Finance nâng cao.

### 2.3. Quyết định kỹ thuật đã chốt

| Hạng mục | Quyết định |
|---|---|
| Backend | Go |
| Frontend | React / Next.js + TypeScript |
| Database | PostgreSQL |
| API | REST + OpenAPI |
| Architecture | Modular Monolith |
| Stock | Immutable Stock Ledger |
| Auth | RBAC + field-level control |
| Audit | Audit log bắt buộc cho hành động nhạy cảm |
| DevOps | Docker, CI/CD, staging/UAT/prod, migration control |
| Storage | S3-compatible / MinIO |

### 2.4. Quyết định cần ban lãnh đạo duyệt

1. Duyệt phạm vi Phase 1.
2. Duyệt Sprint 0 để dựng móng kỹ thuật.
3. Chỉ định Product Owner / System Owner nội bộ.
4. Chốt quy trình phê duyệt nghiệp vụ.
5. Chốt nguồn lực key user: kho, QA/QC, sales, finance, production/subcontract.
6. Chốt cách chọn vendor/dev team nếu thuê ngoài.

---

## 3. Board Presentation Outline

Phần này có thể dùng để chuyển thành slide deck chính thức.

---

## Slide 1 — ERP Phase 1: From Operating Chaos to Controlled Growth

### Key Message

Công ty cần ERP không phải vì “cần phần mềm”, mà vì công ty đang cần một **hệ điều hành vận hành** để scale mà không thất thoát.

### Talking Points

- Mỹ phẩm là ngành có batch, hạn dùng, QC, claim, mẫu, hàng hoàn, kênh bán và vận chuyển phức tạp.
- Nếu không kiểm soát dữ liệu từ đầu, tăng trưởng càng nhanh thì lỗi càng đắt.
- ERP Phase 1 tập trung vào hàng hóa, kho, đơn hàng, giao hàng, QC và gia công ngoài.

---

## Slide 2 — Current Operating Reality

### Key Message

Workflow thực tế đã có cấu trúc, nhưng còn phụ thuộc nhiều vào thao tác thủ công và đối soát cuối ngày.

### Evidence from As-Is Workflows

- Kho có chuỗi công việc hằng ngày: tiếp nhận đơn, xuất/nhập, soạn/đóng, tối ưu vị trí, kiểm kê cuối ngày, đối soát và kết thúc ca.
- Nội quy kho chia thành nhập kho, xuất kho, đóng hàng và hàng hoàn.
- Bàn giao ĐVVC cần phân khu, đối chiếu số lượng, quét mã, ký xác nhận hoặc xử lý thiếu đơn.
- Sản xuất/gia công ngoài có các bước chuyển NVL/bao bì, duyệt mẫu, nhận hàng, kiểm tra, nhập kho hoặc claim nhà máy.

### Business Meaning

Công ty không chỉ cần module kho đơn giản. Công ty cần **warehouse operation engine** có daily board, shift closing, scan, manifest, return inspection và subcontract manufacturing flow.

---

## Slide 3 — Core Business Risks If ERP Is Not Built Correctly

| Rủi ro | Hậu quả | ERP Phase 1 xử lý bằng gì |
|---|---|---|
| Tồn kho lệch | Sale bán nhầm, kho không giao được | Stock ledger + kiểm kê + đối soát cuối ngày |
| Không quản batch/hạn dùng | Không truy xuất được lỗi, rủi ro bán cận date | Batch/expiry/QC status |
| Bàn giao ĐVVC thiếu đơn | Mất hàng, khó truy trách nhiệm | Carrier manifest + scan verification |
| Hàng hoàn nhập sai | Lệch tồn, hàng lỗi quay lại bán | Return inspection + disposition |
| Gia công ngoài thiếu kiểm soát | Mất NVL/bao bì, lỗi mẫu, claim trễ | Subcontract order + material transfer + sample approval + claim window |
| Ai cũng sửa dữ liệu | Không tin được số liệu | RBAC + audit log + data governance |

---

## Slide 4 — Phase 1 Scope

### In Scope

1. Master Data.
2. Purchasing / Supplier cơ bản.
3. QA/QC cơ bản.
4. Inventory / WMS cơ bản.
5. Sales Order / OMS cơ bản.
6. Pick / Pack / Shipping Handover.
7. Returns / hàng hoàn.
8. Subcontract Manufacturing / gia công ngoài.
9. RBAC / Approval / Audit Log.
10. Reporting/KPI Phase 1.
11. API / DB / DevOps foundation.

### Out of Scope for Phase 1

- CRM nâng cao.
- HRM/payroll đầy đủ.
- KOL/affiliate campaign đầy đủ.
- Finance/accounting full posting.
- AI forecasting.
- Mobile app native.
- Multi-brand P&L nâng cao.

### Principle

Phase 1 phải khóa **hàng — batch — kho — đơn — giao — hoàn — gia công ngoài** trước. Những phần tăng trưởng làm sau khi lõi vận hành sạch.

---

## Slide 5 — Target Operating Model

```text
Master Data
    ↓
Purchase / Supplier
    ↓
Inbound Receiving → QC Hold/Pass/Fail → Available Stock
    ↓
Inventory / Batch / Expiry / Stock Ledger
    ↓
Sales Order → Reserve → Pick → Pack → Carrier Manifest → Handover Scan
    ↓
Delivery / COD / Return
    ↓
Return Inspection → Reuse / Damage / Lab / Adjustment
```

Nhánh gia công ngoài:

```text
Subcontract Order
    ↓
Material / Packaging Transfer to Factory
    ↓
Sample Approval
    ↓
Mass Production
    ↓
Factory Delivery to Warehouse
    ↓
Receiving + QC
    ↓
Accept / Reject / Claim within 3–7 days
```

---

## Slide 6 — System Architecture

### Architecture Decision

ERP Phase 1 dùng **Modular Monolith** để cân bằng giữa tốc độ triển khai, kiểm soát kỹ thuật và khả năng mở rộng.

```text
Frontend: React / Next.js / TypeScript
        ↓ REST + OpenAPI
Backend: Go Modular Monolith
        ↓
PostgreSQL + Redis + Queue + S3/MinIO
        ↓
CI/CD + Monitoring + Audit + Backup
```

### Why Modular Monolith

- Chưa cần microservices phức tạp.
- Dễ kiểm soát transaction hàng/kho/đơn/QC.
- Dễ triển khai nhanh Phase 1.
- Có thể tách module sau khi scale thật.

---

## Slide 7 — Module Map

| Module | Vai trò |
|---|---|
| Master Data | Nguồn dữ liệu gốc: SKU, batch, kho, NCC, khách, nhân viên |
| Inventory | Stock ledger, tồn kho, batch, expiry, kiểm kê |
| QC | Hold/pass/fail, inspection, release |
| Purchase | PR/PO/receiving cơ bản |
| Sales | Đơn hàng, reserve, pricing cơ bản |
| Warehouse Ops | Daily board, pick, pack, shift closing |
| Shipping | Carrier manifest, scan handover |
| Returns | Nhận hàng hoàn, kiểm tình trạng, phân loại |
| Subcontract | Gia công ngoài, chuyển NVL/bao bì, duyệt mẫu, claim nhà máy |
| Security | RBAC, approval, audit, break-glass |
| Reporting | KPI vận hành Phase 1 |

---

## Slide 8 — Critical Control: Immutable Stock Ledger

### Key Message

ERP không được cho sửa tồn kho trực tiếp. Mọi biến động tồn phải đi qua stock movement.

### Movement Examples

```text
INBOUND_RECEIPT
QC_RELEASE
SALES_RESERVE
SALES_ISSUE
RETURN_RECEIPT
RETURN_DISPOSITION
SUBCONTRACT_MATERIAL_ISSUE
SUBCONTRACT_FINISHED_GOODS_RECEIPT
STOCK_ADJUSTMENT
CYCLE_COUNT_ADJUSTMENT
```

### Why It Matters

- Truy được ai làm gì, khi nào.
- Biết tồn lệch do đâu.
- Không ai “sửa số cho đẹp”.
- Là nền cho finance và giá vốn sau này.

---

## Slide 9 — Warehouse Daily Board

### Key Message

Kho không chỉ làm từng phiếu. Kho làm theo **nhịp vận hành ngày**.

### Daily Board Should Show

- Đơn mới trong ngày.
- Đơn cần soạn.
- Đơn đang đóng.
- Đơn chờ bàn giao ĐVVC.
- Đơn thiếu hàng / lỗi scan.
- Hàng hoàn chờ kiểm.
- Kiểm kê cuối ngày.
- Đối soát cuối ca.

### Benefit

- Trưởng kho nhìn một màn biết ngày hôm đó đang tắc ở đâu.
- Giảm nhắn Zalo/Excel rời rạc.
- Tạo accountability rõ theo ca.

---

## Slide 10 — Shipping Handover Control

### Current Reality

Bàn giao ĐVVC cần phân khu, để theo thùng/rổ, đối chiếu số lượng, lấy hàng và quét mã, rồi mới ký xác nhận. Nếu chưa đủ đơn thì kiểm tra lại mã hoặc tìm lại ở khu vực đóng hàng.

### ERP Requirement

- Carrier manifest theo ngày/chuyến/ĐVVC.
- Scan từng đơn hoặc mã vận đơn.
- Đủ đơn mới cho handover.
- Thiếu đơn phải tạo exception.
- Ký xác nhận số hóa.
- Ghi audit log.

### KPI

- Handover accuracy.
- Missing order count.
- Scan mismatch rate.
- Time from packed to handed over.

---

## Slide 11 — Returns / Hàng Hoàn Control

### Current Reality

Hàng hoàn nhận từ shipper, đưa vào khu vực hàng hoàn, quét hàng hoàn, quay/ghi tình trạng, kiểm tra bên trong rồi phân loại còn sử dụng hoặc không sử dụng.

### ERP Requirement

- Return receiving by scan.
- Return staging area.
- Condition check.
- Photo/video/attachment.
- Disposition:
  - reusable
  - damaged
  - lab/checking
  - scrap
- Stock movement theo disposition.
- Audit log.

### KPI

- Return processing lead time.
- Return reusable rate.
- Return damage rate.
- Unknown return count.

---

## Slide 12 — Subcontract Manufacturing Control

### Current Reality

Luồng sản xuất/gia công ngoài gồm lên đơn với nhà máy, xác nhận số lượng/quy cách/mẫu mã, cọc đơn, chuyển NVL/bao bì, làm mẫu, chốt mẫu, sản xuất hàng loạt, giao về kho, kiểm tra, nhập kho hoặc báo lỗi nhà máy trong 3–7 ngày.

### ERP Requirement

- Subcontract order.
- Deposit/final payment tracking.
- Material/packaging transfer to factory.
- Sample approval.
- Production milestone.
- Factory delivery receiving.
- QC inspection.
- Acceptance/rejection.
- Factory claim within claim window.

### KPI

- Factory on-time delivery.
- Material variance at factory.
- Sample approval cycle time.
- Factory defect rate.
- Claim submitted within 3–7 days.

---

## Slide 13 — Security, RBAC, Audit

### Principle

ERP phải kiểm soát “ai được làm gì” ngay từ đầu. Đừng để quyền quá rộng rồi sau này đi vá.

### Controls

- Role-based access control.
- Field-level permission cho giá vốn, công nợ, payout, lương.
- Approval theo ngưỡng/rủi ro.
- Sensitive action confirmation.
- Audit log trước/sau.
- Break-glass access có lý do và expiry.
- Export control.

### Sensitive Areas

- Stock correction.
- QC pass/fail.
- Batch/expiry.
- Order cancellation.
- Return disposition.
- Supplier/factory claim.
- Finance/cost data.

---

## Slide 14 — Implementation Roadmap

### Stage 0 — Alignment & Sprint 0

- Chốt document package.
- Setup repo, backend, frontend, DB, CI/CD.
- Build skeleton: login, RBAC, master data, stock ledger prototype, scan prototype.

### Stage 1 — Core Build

- Master data.
- Inventory + stock ledger.
- QC.
- Sales order + reserve.
- Pick/pack.
- Shipping manifest + scan.
- Returns.
- Subcontract manufacturing.

### Stage 2 — UAT & Cutover

- UAT theo kịch bản.
- Migration test.
- Training key user.
- Go-live runbook.
- Hypercare.

### Stage 3 — Phase 2 Foundation

- CRM.
- KOL/Affiliate.
- HRM.
- Finance nâng cao.
- BI nâng cao.

---

## Slide 15 — Team & Governance

### Required Internal Roles

| Role | Responsibility |
|---|---|
| Executive Sponsor | Duyệt scope, ngân sách, quyết định lớn |
| Product Owner | Chốt nghiệp vụ, ưu tiên backlog |
| System Owner | Chủ hệ thống sau go-live |
| Warehouse Super User | Chốt flow kho, UAT kho |
| QA/QC Super User | Chốt QC/batch/hạn dùng |
| Sales/Ops Super User | Chốt sales order/giao hàng |
| Finance Representative | Công nợ, đối soát, kiểm soát tài chính |
| Production/Subcontract Owner | Chốt gia công ngoài |
| Tech Lead | Kiến trúc kỹ thuật, code quality |
| PM/BA | Điều phối, viết story, quản lý scope |

### Governance Rules

- Không build khi requirement chưa có owner.
- Không sửa flow trực tiếp bằng code; phải qua change request.
- Không go-live nếu stock ledger, QC, shipping handover và return flow chưa pass UAT.

---

## Slide 16 — Success Metrics

### Business KPIs

| KPI | Target Direction |
|---|---|
| Inventory accuracy | Tăng |
| Stock discrepancy count | Giảm |
| Order fulfillment lead time | Giảm |
| Pick/pack error rate | Giảm |
| Handover mismatch rate | Giảm |
| Return processing time | Giảm |
| Batch/QC traceability | Tăng |
| Factory claim within window | Tăng |
| Manual Excel dependency | Giảm |
| Audit coverage for sensitive actions | Tăng |

### Technical KPIs

| KPI | Target Direction |
|---|---|
| API contract coverage | Tăng |
| Automated test coverage for core flows | Tăng |
| Deployment success rate | Tăng |
| P0/P1 incident count | Giảm |
| Migration success rate | Tăng |
| Backup/restore verification | Pass |

---

## Slide 17 — Key Decisions Needed

Ban lãnh đạo cần chốt:

1. Approve Phase 1 scope.
2. Approve Go backend + React/Next.js frontend + PostgreSQL stack.
3. Approve Sprint 0 kickoff.
4. Assign Product Owner and System Owner.
5. Assign super users for warehouse, QC, sales/ops, finance, subcontract.
6. Decide build model:
   - in-house
   - vendor
   - hybrid
7. Approve vendor RFP/SOW evaluation if using vendor.
8. Approve data governance and change control.

---

## 4. Document Package Summary

Bộ tài liệu hiện tại đã phủ từ chiến lược tới triển khai.

| Nhóm | Tài liệu | Mục đích |
|---|---|---|
| Strategy | 01–02 | Blueprint và roadmap tài liệu |
| Product | 03–10 | PRD, quyền, data, flow, KPI, screen, UAT, migration |
| Engineering | 11–18 | Go architecture, coding, module, UI/UX, FE, API, DB, DevOps |
| Governance | 19–22 | Security, As-Is, Gap Analysis, Revision Log |
| Delivery | 23–31 | Integration, QA, Backlog, SOP, Go-live, Risk, Support, Data Governance, Phase 2 |
| Handoff | 32–35 | Master Index, v1.1 pack, Sprint 0, Vendor RFP/SOW |
| Executive | 36 | Board summary and presentation narrative |

---

## 5. Suggested Board Narrative

Dưới đây là kịch bản nói chuyện ngắn cho CEO/PM khi trình bày.

> Công ty không chỉ cần một phần mềm quản lý. Công ty cần chuẩn hóa cách vận hành để scale.  
> Phase 1 sẽ không ôm tất cả, mà tập trung vào lõi hàng hóa: kho, batch, QC, đơn, giao, hoàn, gia công ngoài và audit.  
> Sau khi soi workflow thật, chúng ta thấy điểm đặc thù lớn nhất là kho vận hành theo nhịp ngày, có kiểm kê và đối soát cuối ca; bàn giao vận chuyển cần scan và đối chiếu; hàng hoàn cần kiểm tình trạng; sản xuất hiện có gia công ngoài nên phải kiểm soát chuyển NVL/bao bì, duyệt mẫu, nhận hàng và claim nhà máy.  
> Vì vậy hệ thống được thiết kế theo modular monolith bằng Go, React/Next.js, PostgreSQL, OpenAPI, stock ledger bất biến, RBAC và audit log.  
> Kế hoạch tiếp theo là Sprint 0 để dựng móng kỹ thuật, rồi build từng module theo backlog đã chốt.  
> Điều cần duyệt hôm nay là scope Phase 1, Product Owner, team super user, Sprint 0 và mô hình build/vendor.

---

## 6. Decision Matrix for Build Model

| Option | Ưu điểm | Nhược điểm | Khi nào chọn |
|---|---|---|---|
| In-house | Kiểm soát cao, hiểu business sâu | Cần đội tech mạnh, khó tuyển đủ | Có Tech Lead và dev core tốt |
| Vendor | Nhanh hơn nếu vendor tốt | Rủi ro lệ thuộc, chất lượng khó kiểm | Cần timeline nhanh, có RFP/SOW chặt |
| Hybrid | Cân bằng kiểm soát và tốc độ | Cần quản trị dự án tốt | Khuyến nghị cho ERP này |

### Khuyến nghị

Dùng **hybrid model**:

- Công ty giữ Product Owner, System Owner, Super User và quyền sở hữu source code/data.
- Vendor/dev team build theo RFP/SOW, architecture, coding standard, API, DB và backlog đã chốt.
- Tech Lead nội bộ hoặc external independent architect review code/architecture định kỳ.

---

## 7. Executive Risk Register

| Risk | Level | Mitigation |
|---|---:|---|
| Scope creep | High | Phase 1 scope gate + change control |
| User không dùng đúng | High | SOP + training + super user |
| Tồn kho migration sai | High | Migration test + stock count + cutover runbook |
| Vendor build lệch nghiệp vụ | High | RFP/SOW + traceability + UAT |
| Không có business owner | High | Appoint Product Owner trước Sprint 0 |
| API/DB không chuẩn | Medium | OpenAPI + DB standard + code review |
| Go team thiếu kinh nghiệm | Medium | Coding standard + module standard + Tech Lead review |
| Go-live rối | High | Go-live runbook + hypercare + support model |
| Data governance yếu | Medium | Data governance/change control |
| Hàng hoàn/giao hàng không kiểm soát | High | Manifest scan + return disposition |

---

## 8. Minimum Viable Phase 1 Demo

Sprint 0 hoặc đầu Phase 1 nên có demo nhỏ nhưng sống thật:

```text
Login
→ user có role Warehouse Staff
→ tạo SKU / batch / kho
→ nhập hàng vào QC hold
→ QC pass
→ available stock tăng
→ tạo sales order
→ reserve stock
→ pick/pack
→ tạo carrier manifest
→ scan bàn giao ĐVVC
→ audit log ghi đầy đủ
```

Demo mở rộng:

```text
Return receiving
→ scan vận đơn/đơn hàng
→ kiểm tình trạng
→ chọn disposition
→ stock movement phù hợp
→ audit log
```

Demo gia công ngoài:

```text
Subcontract order
→ chuyển NVL/bao bì cho nhà máy
→ duyệt mẫu
→ nhận hàng thành phẩm
→ QC
→ accept/reject/claim
```

---

## 9. Leadership Checklist Before Sprint 0

Trước khi Sprint 0 bắt đầu, cần có:

- [ ] CEO/Executive Sponsor duyệt scope Phase 1.
- [ ] Product Owner được chỉ định.
- [ ] System Owner được chỉ định.
- [ ] Super User kho được chỉ định.
- [ ] Super User QA/QC được chỉ định.
- [ ] Super User sales/ops được chỉ định.
- [ ] Finance representative được chỉ định.
- [ ] Production/subcontract owner được chỉ định.
- [ ] Tech Lead/vendor được xác nhận.
- [ ] Repository/source code ownership được chốt.
- [ ] UAT environment và staging plan được chốt.
- [ ] Data migration owner được chốt.
- [ ] Change control rule được duyệt.

---

## 10. What Not To Do

Không nên:

- Nhồi CRM, HRM, KOL, Finance full vào Phase 1.
- Cho dev code khi chưa có Product Owner chốt nghiệp vụ.
- Cho sửa tồn kho trực tiếp trong database.
- Bỏ qua scan/manifest trong bàn giao ĐVVC.
- Bỏ qua hàng hoàn hoặc coi hàng hoàn là “nhập kho thường”.
- Coi gia công ngoài giống sản xuất nội bộ 100%.
- Go-live khi chưa kiểm kê/migration/stock ledger test.
- Để vendor sở hữu source code hoặc data.
- Bỏ audit log vì “để làm sau”.

---

## 11. Final Recommendation

### Recommended Next Move

1. Trình bày file 36 này cho ban lãnh đạo/team core.
2. Duyệt Sprint 0.
3. Chốt Product Owner + System Owner + Super User.
4. Dùng file 35 để chọn vendor/dev team nếu cần.
5. Bắt đầu Sprint 0 với mục tiêu tạo backend/frontend/db/API skeleton chạy được.

### Final Message

**ERP Phase 1 phải được xem là dự án chuẩn hóa vận hành, không phải dự án phần mềm đơn thuần.**  
Muốn hệ thống sống lâu, phải khóa từ đầu 5 thứ: **workflow thật, dữ liệu chuẩn, quyền hạn rõ, stock ledger bất biến, và người chịu trách nhiệm quyết định.**

---

## Appendix A — Board-Level FAQ

### ERP này có làm full tài chính kế toán ngay không?

Không. Phase 1 làm thu/chi/công nợ/đối soát cơ bản và dữ liệu nền cho finance. Full accounting posting, closing và profitability sâu để Phase 2.

### Có làm CRM/KOL/HRM ngay không?

Không làm full trong Phase 1. Phase 1 ưu tiên lõi hàng/kho/đơn/giao/hoàn/gia công ngoài. CRM/KOL/HRM được đưa vào Phase 2 trên nền dữ liệu Phase 1.

### Vì sao không dùng microservices ngay?

Vì Phase 1 cần tốc độ, transaction chắc, dữ liệu nhất quán. Modular monolith phù hợp hơn. Khi scale thật, có thể tách module sau.

### Vì sao dùng Go?

Go phù hợp backend vận hành dài hạn, transaction rõ, performance tốt, deploy gọn. Nhưng phải đi kèm coding standard, module boundary và architecture guardrail.

### Điểm rủi ro lớn nhất là gì?

Không phải code. Rủi ro lớn nhất là workflow không được chốt, Product Owner không đủ quyền, dữ liệu migration sai, user không dùng đúng, và vendor build lệch nghiệp vụ.

---

## Appendix B — Slide-to-Document Mapping

| Slide | Related Documents |
|---|---|
| Slide 1–4 | 01, 03, 20, 21, 31 |
| Slide 5 | 06, 20, 21, 33 |
| Slide 6 | 11, 15, 16, 17, 18 |
| Slide 7 | 03, 13, 25 |
| Slide 8 | 05, 11, 12, 17, 19 |
| Slide 9–12 | 20, 21, 23, 24, 26, 28 |
| Slide 13 | 04, 19, 30 |
| Slide 14 | 25, 27, 34 |
| Slide 15 | 29, 30, 35 |
| Slide 16–17 | 07, 25, 32, 35 |

---

## Appendix C — Glossary for Board

| Term | Meaning |
|---|---|
| ERP | Hệ thống quản trị nguồn lực doanh nghiệp |
| WMS | Warehouse Management System — quản lý kho |
| OMS | Order Management System — quản lý đơn hàng |
| QC | Quality Control — kiểm soát chất lượng |
| Batch/Lot | Lô hàng/lô sản xuất |
| FEFO | First Expired, First Out — hàng hết hạn trước xuất trước |
| Stock Ledger | Sổ cái biến động tồn kho |
| Manifest | Danh sách/chuyến bàn giao cho ĐVVC |
| RBAC | Role-Based Access Control — phân quyền theo vai trò |
| Audit Log | Nhật ký thao tác hệ thống |
| UAT | User Acceptance Test — nghiệm thu người dùng |
| Hypercare | Giai đoạn hỗ trợ sát sau go-live |
| Modular Monolith | Một hệ thống backend chính nhưng chia module rõ |
| OpenAPI | Chuẩn mô tả API để FE/BE/test đồng bộ |
| Subcontract Manufacturing | Gia công ngoài / đặt nhà máy sản xuất |

---

**End of Document.**
