# 138_ERP_UAT_Pilot_Pack_Production_External_Factory_E2E_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1 / Production External-Factory Pilot
Document role: UAT pilot pack for end-to-end external-factory production flow
Status: Draft for UAT preparation
Owner: Business Owner / Production Owner / ERP Product Owner
Last updated: 2026-05-07

---

## 1. Purpose

Tài liệu này dùng để chạy UAT thực tế cho luồng **Sản xuất / Gia công ngoài / Nhà máy ngoài** trong ERP.

Mục tiêu không phải là thêm feature mới, mà là cho người vận hành thật kiểm tra một vòng đời sản xuất từ:

```text
Production Plan
-> Material Demand
-> Purchase Request / PO / Receiving / QC
-> Warehouse Issue to Factory
-> Factory Dispatch
-> Material Handover
-> Sample Approval
-> Mass Production
-> Finished Goods Receipt to QC Hold
-> Finished Goods QC Closeout
-> Factory Claim / Acceptance
-> Final Payment Readiness
-> AP Handoff
-> Finance Closeout
-> Cash-out Voucher Evidence
```

Tài liệu này đặc biệt phù hợp với công ty mỹ phẩm có mô hình sản xuất theo nhà máy ngoài: công ty lên đơn, chuyển nguyên vật liệu/bao bì, duyệt mẫu, nhà máy sản xuất hàng loạt, giao hàng về kho, công ty kiểm chất lượng rồi mới nghiệm thu/thanh toán.

---

## 2. Critical positioning

### 2.1. Production trong Phase 1 nghĩa là gì?

Trong Phase 1, **Production** không phải MES nội bộ đầy đủ với máy móc, line, routing, work center, labor costing.

Trong Phase 1:

```text
Production = External-factory production / Subcontract manufacturing
```

Nói dễ hiểu:

```text
Công ty quản kế hoạch, vật tư, giao nhà máy, nhận hàng, QC, claim, công nợ và thanh toán.
Nhà máy ngoài chịu trách nhiệm sản xuất vật lý.
```

### 2.2. Entry point người dùng

Người dùng business nên đi từ:

```text
/production
```

Không nên bắt người dùng hiểu `/subcontract` như một module riêng. `/subcontract` nếu còn tồn tại thì chỉ là legacy/technical route.

### 2.3. Không thay thế Sprint 22 UAT

Nếu Sprint 22 Warehouse/Sales/QC UAT chưa có Go hoặc Conditional Go, tài liệu này vẫn có thể dùng để chạy **Production E2E discovery / controlled UAT**, nhưng không được gọi là full business release acceptance.

Guardrail:

```text
Core Warehouse/Sales/QC chưa business-accepted
=> Production E2E UAT có thể chạy để tìm lỗi sớm
=> Nhưng không được dùng thay thế Sprint 22 Go/No-Go
```

---

## 3. Source-of-truth references

Tài liệu này nên được đọc cùng các tài liệu sau:

```text
80_ERP_Master_Document_Index_Current_Status_MyPham_v2.md
85_ERP_UAT_Pilot_Pack_Sprint22_Warehouse_Sales_QC_MyPham_v1.md
86_ERP_Sprint22_Changelog_UAT_Pilot_Pack_Warehouse_Sales_QC_MyPham_v1.md
91_ERP_Module_Roadmap_From_Note_Sheet_Production_Purchase_Warehouse_MyPham_v1.md
94_ERP_Purchase_Request_Workflow_Production_Plan_PO_Traceability_MyPham_v1.md
95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1.md
106_ERP_Production_IA_External_Factory_Order_Detail_Flow_MyPham_v1.md
109_ERP_Factory_Dispatch_Flow_Sprint27_MyPham_v1.md
115_ERP_Factory_Material_Handover_Flow_Sprint29_MyPham_v1.md
121_ERP_Factory_Finished_Goods_Receipt_QC_Hold_Flow_Sprint31_MyPham_v1.md
124_ERP_Factory_Finished_Goods_QC_Closeout_Flow_Sprint32_MyPham_v1.md
130_ERP_Factory_Final_Payment_AP_Handoff_Flow_Sprint34_MyPham_v1.md
136_ERP_Factory_Final_Payment_Voucher_Flow_Sprint36_MyPham_v1.md
139_ERP_Production_E2E_Discovery_Mode_S36_Blocker_MyPham_v1.md
```

Workflow thực tế nền:

```text
Công-việc-hằng-ngày.pdf
Nội-Quy.pdf
Quy-trình-bàn-giao.pdf
Quy-trình-sản-xuất.pdf
```

---

## 4. UAT mode

Có 2 chế độ chạy tài liệu này.

### Mode A — Business UAT chính thức

Chỉ dùng khi:

```text
- Môi trường UAT ổn định.
- Role Production / Warehouse / QC / Purchasing / Finance login được.
- Seed data được duyệt.
- Người vận hành thật hoặc người được ủy quyền chạy script.
- Evidence được ghi lại.
```

Kết quả có thể là:

```text
Go
Conditional Go
No-Go
```

### Mode B — Production flow discovery

Dùng khi core UAT chưa pass hoặc chưa đủ business users.

Kết quả chỉ được gọi là:

```text
Discovery observation
Developer/business walkthrough
Not business UAT pass
```

Không được tạo release tag hoặc claim production-ready từ Mode B.

Mode B / Discovery rule:

```text
Production E2E Discovery may proceed without PFX-UAT-013 if Sprint 36 runtime is not ready.
In older snapshots where Sprint 36 runtime was not ready, PFX-UAT-013 was marked BLOCKED_BY_S36.
The session result must be recorded as Discovery only — no business UAT decision.
```

---

## 5. Roles and responsibilities

| Role | Responsibility |
|---|---|
| Business Owner | Chốt Go / Conditional Go / No-Go |
| Production Owner | Chạy và xác nhận luồng sản xuất/gia công ngoài |
| Purchasing User | Kiểm PR/PO, nhà cung cấp, giá, nhận hàng mua |
| Warehouse User | Kiểm xuất NVL/bao bì, bàn giao, nhập thành phẩm, tồn kho |
| QC User | Kiểm inbound QC, sample approval, finished goods QC closeout |
| Finance/AP User | Kiểm payable, AP handoff, payment voucher, cash-out evidence |
| ERP Admin | Tạo user, role, seed data, hỗ trợ môi trường |
| QA/UAT Lead | Điều phối session, ghi scenario result, issue triage, evidence |
| Factory Contact | Có thể là người thật hoặc người mô phỏng phản hồi nhà máy |

---

## 6. Entry criteria

Không bắt đầu UAT nếu thiếu các điều kiện sau:

```text
1. main branch clean và CI green hoặc trạng thái CI được documented rõ.
2. Môi trường UAT/dev target health = 200.
3. User Production/Warehouse/QC/Purchasing/Finance login được.
4. /me trả đúng role/permission.
5. UI tiếng Việt baseline pass.
6. Production menu trỏ đúng entrypoint /production.
7. Seed data production được load/approve.
8. Evidence folder đã tạo.
9. UAT participants hiểu rõ Bug vs Change Request vs Training/Data Issue.
10. Không còn P0 blocker từ auth/RBAC/seed data.
```

---

## 7. Seed data requirements

### 7.1. Master data tối thiểu

| Data type | Minimum required |
|---|---|
| Finished goods | 2 SKU thành phẩm mỹ phẩm |
| Raw materials | 3 nguyên liệu, ít nhất 1 dạng G/KG, 1 dạng ML/L |
| Packaging | chai/lọ, nắp, hộp, tem nhãn |
| Formula/BOM | 1 BOM đủ NVL + bao bì cho 1 thành phẩm |
| Warehouses | WH-MAIN, WH-QA-HOLD, WH-FACTORY-OUT, WH-DAMAGED |
| Locations | Ít nhất 2 vị trí cho kho chính, 1 khu QC hold |
| Batches | Batch QC_PASS, QC_HOLD, QC_FAIL, near-expiry |
| Suppliers | Ít nhất 1 NCC nguyên liệu/bao bì |
| Factory | Ít nhất 1 nhà máy gia công ngoài |
| Customers | Không bắt buộc cho production UAT, nhưng nên có để trace later |
| Payment method | Bank transfer / cash-out evidence placeholder |

### 7.2. Production seed cases

```text
Case A: Đủ vật tư cho sản xuất.
Case B: Thiếu vật tư, cần sinh Purchase Request.
Case C: Nhà máy nhận vật tư, duyệt mẫu pass.
Case D: Thành phẩm nhận về full pass.
Case E: Thành phẩm nhận về partial fail, mở factory claim.
Case F: Không được final payment nếu QC closeout chưa xong.
```

---

## 8. Evidence folder structure

Tạo hoặc dùng cấu trúc sau:

```text
docs/uat/production-e2e/
  scenario_results.csv
  observation_log.csv
  issue_triage_board.csv
  go_no_go_report.md
  evidence/
    screenshots/
    logs/
    exports/
    session-notes/
    sanitized-files/
```

Naming convention:

```text
PFX-UAT-001_production-plan_material-demand_<date>.png
PFX-UAT-008_finished-goods-qc-partial-fail_<date>.png
PFX-UAT-013_payment-voucher-evidence_<date>.png
```

Không lưu thông tin nhạy cảm chưa sanitize trong repo public.

---

## 9. Scenario list

| ID | Scenario | Primary owner | Decision impact |
|---|---|---|---|
| PFX-UAT-001 | Production plan + formula/BOM + material demand | Production Owner | Core production planning |
| PFX-UAT-002 | Shortage -> Purchase Request traceability | Production + Purchasing | Governance between production and purchase |
| PFX-UAT-003 | PO -> Receiving -> Inbound QC | Purchasing + QC + Warehouse | Incoming material control |
| PFX-UAT-004 | Warehouse issue NVL/bao bì to factory | Warehouse + Production | Material control |
| PFX-UAT-005 | Factory dispatch pack and confirmation | Production | Factory communication |
| PFX-UAT-006 | Factory material handover evidence | Warehouse + Production | Handover proof |
| PFX-UAT-007 | Sample approval / rework | QC + Production | Pre-mass-production quality gate |
| PFX-UAT-008 | Mass production progress | Production | Factory order execution |
| PFX-UAT-009 | Finished goods receipt to QC hold | Warehouse + QC | Finished goods intake control |
| PFX-UAT-010 | Finished goods QC closeout full/partial/fail | QC | Availability and claim decision |
| PFX-UAT-011 | Factory claim within 3-7 days | QC + Production | Supplier/factory dispute control |
| PFX-UAT-012 | Final payment readiness | Production + Finance | Payment gate |
| PFX-UAT-013 | AP handoff + payment voucher/cash-out evidence | Finance/AP | Finance closeout |
| PFX-UAT-014 | Negative controls | QA/UAT Lead | Risk controls |

---

## 10. Scenario details

### PFX-UAT-001 — Production plan + formula/BOM + material demand

Objective:

```text
Xác nhận user có thể tạo/chọn production plan dựa trên SKU, công thức/BOM, số lượng cần sản xuất, và hệ thống tính được nhu cầu vật tư.
```

Preconditions:

```text
- SKU thành phẩm có BOM/formula.
- Warehouse có tồn vật tư hoặc có shortage case.
- Production user login được.
```

Steps:

```text
1. Vào /production.
2. Tạo hoặc mở production plan.
3. Chọn SKU thành phẩm.
4. Nhập số lượng cần sản xuất.
5. Kiểm BOM/formula được hiển thị.
6. Kiểm material demand theo nguyên liệu/bao bì.
7. Kiểm available stock, shortage, required quantity, UOM/base UOM.
```

Expected results:

```text
- Material demand tính đúng theo BOM/formula.
- Số lượng hiển thị theo UOM chuẩn.
- Shortage được đánh dấu rõ.
- Không tạo PO trực tiếp từ production plan.
```

Evidence:

```text
Screenshot production plan.
Export material demand nếu có.
Log/API response nếu cần.
```

Pass criteria:

```text
PASS nếu production owner xác nhận nhu cầu vật tư dễ hiểu và đúng logic.
FAIL nếu thiếu BOM, sai UOM, sai shortage hoặc không trace được plan.
```

---

### PFX-UAT-002 — Shortage -> Purchase Request traceability

Objective:

```text
Xác nhận shortage từ production plan có thể sinh Purchase Request nhưng không bypass purchasing governance.
```

Steps:

```text
1. Từ production material demand, chọn dòng shortage.
2. Generate Purchase Request.
3. Kiểm PR có link về production plan/order.
4. Submit PR.
5. Purchasing review PR.
6. Convert PR sang PO nếu được approve.
```

Expected results:

```text
- Production chỉ tạo nhu cầu/PR, không tạo PO trực tiếp.
- PR trace được về production plan.
- PO trace được về PR và production need.
```

Guardrail:

```text
Production does not own purchasing commitment.
Purchasing owns PO and supplier commitment.
```

Evidence:

```text
PR detail screenshot.
PO detail screenshot.
Traceability link screenshot.
```

---

### PFX-UAT-003 — PO -> Receiving -> Inbound QC

Objective:

```text
Xác nhận vật tư mua về được tiếp nhận, QC và chỉ QC PASS mới vào tồn khả dụng.
```

Steps:

```text
1. Tạo/approve PO cho vật tư thiếu.
2. Tạo receiving từ PO.
3. Nhập số lượng, batch/lô, HSD, bao bì, chứng từ giao hàng.
4. Chuyển sang inbound QC.
5. Test 3 nhánh: PASS, HOLD, FAIL.
```

Expected results:

```text
- Receiving không đồng nghĩa available stock.
- QC PASS mới tăng available.
- QC HOLD vào quarantine/hold.
- QC FAIL không tăng available, có reject/return supplier path.
```

Evidence:

```text
Receiving screenshot.
QC PASS/HOLD/FAIL screenshot.
Stock movement/stock availability evidence.
```

---

### PFX-UAT-004 — Warehouse issue NVL/bao bì to factory

Objective:

```text
Xác nhận kho có thể xuất nguyên vật liệu/bao bì cho nhà máy ngoài có kiểm soát.
```

Preconditions:

```text
- Factory order đã tồn tại.
- Vật tư đã QC PASS và available.
- Warehouse user có quyền issue stock.
```

Steps:

```text
1. Mở factory order trong /production.
2. Chọn tab/material issue.
3. Chọn nguyên liệu/bao bì cần chuyển.
4. Chọn batch/lô/location.
5. Confirm issue to factory.
6. Kiểm stock movement và warehouse balance.
```

Expected results:

```text
- Không xuất được vật tư QC HOLD/FAIL.
- Lot-controlled material bắt buộc có batch/lô.
- Movement type rõ: ISSUE_TO_FACTORY hoặc tương đương.
- Audit log ghi ai xuất, lúc nào, số lượng nào, batch nào.
```

Evidence:

```text
Material issue document.
Stock movement.
Audit log.
```

---

### PFX-UAT-005 — Factory dispatch pack and confirmation

Objective:

```text
Xác nhận production có thể tạo dispatch pack gửi nhà máy và ghi nhận phản hồi xác nhận của nhà máy.
```

Steps:

```text
1. Mở factory order.
2. Tạo dispatch pack.
3. Kiểm thông tin: sản phẩm, số lượng, quy cách, mẫu mã, vật tư gửi, deadline.
4. Mark ready/sent.
5. Ghi nhận factory confirmation.
```

Expected results:

```text
- Dispatch pack có đủ thông tin cho nhà máy.
- Confirmation không tự động thay đổi tồn kho.
- Factory response được lưu evidence/audit.
```

Out of scope:

```text
Không cần email/Zalo/API automation trong UAT này.
```

---

### PFX-UAT-006 — Factory material handover evidence

Objective:

```text
Xác nhận biên bản bàn giao vật tư cho nhà máy được ghi nhận đủ evidence.
```

Steps:

```text
1. Mở factory material handover.
2. Kiểm danh sách NVL/bao bì theo lô/số lượng.
3. Nhập người nhận phía nhà máy.
4. Đính kèm biên bản/ảnh/chứng từ nếu có.
5. Confirm handover.
```

Expected results:

```text
- Handover chỉ được confirm khi có receiver/evidence theo rule.
- Sau handover, vật tư đang nằm ở trạng thái đã giao nhà máy.
- Audit log đầy đủ.
```

Evidence:

```text
Material handover screenshot.
Attachment metadata screenshot.
Audit log.
```

---

### PFX-UAT-007 — Sample approval / rework

Objective:

```text
Xác nhận bước làm mẫu/chốt mẫu được kiểm soát trước khi mass production.
```

Steps:

```text
1. Tạo sample submission từ factory order.
2. QC/Production review mẫu.
3. Chọn PASS hoặc REWORK/REJECT.
4. Nếu PASS, cho phép chuyển mass production.
5. Nếu REWORK/REJECT, block mass production và ghi lý do.
```

Expected results:

```text
- Mass production không được start nếu sample chưa approved.
- Sample approval có evidence và audit.
- Rework giữ factory order ở trạng thái chờ xử lý.
```

---

### PFX-UAT-008 — Mass production progress

Objective:

```text
Xác nhận factory order có thể chuyển sang sản xuất hàng loạt sau sample approval.
```

Steps:

```text
1. Từ sample approved, start mass production.
2. Nhập planned completion date hoặc factory progress note.
3. Cập nhật trạng thái sản xuất.
4. Kiểm timeline/status trong production detail.
```

Expected results:

```text
- Chỉ sample approved mới mass production.
- Timeline rõ ràng.
- Không tạo finished goods stock khi mới mass production.
```

---

### PFX-UAT-009 — Finished goods receipt to QC hold

Objective:

```text
Xác nhận thành phẩm nhà máy giao về được nhận vào QC HOLD, không vào available ngay.
```

Steps:

```text
1. Mở factory order mass production.
2. Tạo finished goods receipt.
3. Nhập số lượng nhận, batch/lô thành phẩm, HSD, bao bì, chứng từ giao hàng.
4. Confirm receipt.
5. Kiểm tồn kho.
```

Expected results:

```text
- Thành phẩm vào QC HOLD/quarantine.
- Available stock không tăng trước QC PASS.
- Receipt trace về factory order.
```

Evidence:

```text
Finished goods receipt screenshot.
QC hold stock screenshot.
No available increase evidence.
```

---

### PFX-UAT-010 — Finished goods QC closeout full/partial/fail

Objective:

```text
Xác nhận QC closeout quyết định hàng nào available, hàng nào claim/không nhận.
```

Test branches:

```text
A. Full PASS
B. Partial PASS / Partial FAIL
C. Full FAIL
```

Steps:

```text
1. Tạo QC closeout cho finished goods receipt.
2. Nhập inspected quantity.
3. Nhập pass/reject quantity.
4. Nhập reason/evidence.
5. Confirm QC closeout.
6. Kiểm stock movement, available, factory claim.
```

Expected results:

```text
Full PASS:
- Accepted quantity tăng available.
- Factory order có thể đi final payment readiness nếu điều kiện khác đạt.

Partial PASS:
- Pass quantity tăng available.
- Reject quantity tạo claim hoặc pending resolution.

Full FAIL:
- Không tăng available.
- Tạo factory claim.
```

Guardrail:

```text
Receipt != QC Pass != Available Stock != Final Payment Readiness.
```

---

### PFX-UAT-011 — Factory claim within 3-7 days

Objective:

```text
Xác nhận khi hàng không đạt, hệ thống có claim nhà máy và deadline xử lý 3-7 ngày.
```

Steps:

```text
1. Từ QC FAIL hoặc partial fail, tạo/open factory claim.
2. Nhập reason: sai số lượng, sai quy cách, lỗi bao bì, lỗi chất lượng, không đạt kiểm nghiệm.
3. Nhập target response date trong 3-7 ngày.
4. Đính kèm evidence.
5. Assign owner.
```

Expected results:

```text
- Claim trace về factory order, receipt, QC closeout.
- Claim có owner/deadline/status.
- Final payment bị block nếu claim chưa resolve theo rule.
```

---

### PFX-UAT-012 — Final payment readiness

Objective:

```text
Xác nhận production chỉ đánh dấu ready for final payment khi đủ điều kiện nghiệm thu.
```

Preconditions:

```text
- Hàng đã QC closeout.
- Claim nếu có đã resolve hoặc được accepted theo rule.
- Accepted quantity/value rõ ràng.
```

Steps:

```text
1. Mở factory order.
2. Kiểm final payment readiness panel.
3. Thử mark ready khi QC chưa closeout -> phải bị block.
4. Mark ready sau khi QC/claim pass điều kiện.
5. Kiểm handoff sang Finance/AP.
```

Expected results:

```text
- Production không tự chi tiền.
- Production chỉ tạo readiness/handoff.
- Finance/AP sở hữu payable/payment.
```

---

### PFX-UAT-013 — AP handoff + payment voucher/cash-out evidence

Objective:

```text
Xác nhận Finance có thể nhận AP handoff, kiểm payable, tạo cash-out/payment evidence và đóng tài chính.
```

Dependency:

```text
This scenario depended on Sprint 36 runtime implementation for payment voucher / cash-out evidence.
Sprint 36 PR #616 is merged and smoke-tested, so this scenario is ready to run in controlled discovery mode.
Do not claim Business UAT Go until this scenario is actually executed or explicitly waived by the Business Owner with a documented reason.
No release tag may be created from discovery preparation alone.
```

Steps:

```text
1. Finance mở AP handoff từ factory order.
2. Kiểm payable amount, deposit, final amount, claim deduction nếu có.
3. Tạo payment request/cash-out voucher.
4. Ghi payment method, reference, payment date, memo.
5. Đính kèm chứng từ thanh toán.
6. Post cash-out/payment evidence.
7. Kiểm supplier payable status.
```

Expected results:

```text
- Payment evidence trace về factory order/AP.
- Không có hai sự thật thanh toán độc lập.
- Supplier payable payment status phải derive hoặc reconcile với posted cash-out allocation.
```

Critical guardrail:

```text
Supplier payable payment status must be derived from or reconciled with posted cash-out allocations.
Do not maintain two independent payment truths.
```

---

### PFX-UAT-014 — Negative controls

Objective:

```text
Xác nhận hệ thống chặn các thao tác rủi ro.
```

Negative cases:

```text
1. Không cho mass production nếu sample chưa approved.
2. Không cho issue vật tư QC HOLD/FAIL sang nhà máy.
3. Không cho available stock tăng khi finished goods mới receipt chưa QC pass.
4. Không cho final payment readiness nếu QC closeout chưa xong.
5. Không cho final payment nếu factory claim unresolved theo rule.
6. Không cho finance post payment vượt payable amount.
7. Không cho user không đủ quyền approve/QC/pay.
8. Không cho sửa trực tiếp stock balance.
```

Expected results:

```text
- UI/API trả lỗi tiếng Việt dễ hiểu.
- Audit log hoặc denied action log có đủ evidence.
- Route/menu guard đúng role.
```

---

## 11. Scenario result template

Use this table or create `scenario_results.csv`.

| Scenario ID | Scenario | Owner | Status | Evidence link | Issue IDs | Sign-off |
|---|---|---|---|---|---|---|
| PFX-UAT-001 | Production plan + material demand | Production | Pending |  |  |  |
| PFX-UAT-002 | Shortage -> PR traceability | Production/Purchasing | Pending |  |  |  |
| PFX-UAT-003 | PO -> Receiving -> Inbound QC | Purchasing/QC/Warehouse | Pending |  |  |  |
| PFX-UAT-004 | Warehouse issue to factory | Warehouse/Production | Pending |  |  |  |
| PFX-UAT-005 | Factory dispatch confirmation | Production | Pending |  |  |  |
| PFX-UAT-006 | Material handover evidence | Warehouse/Production | Pending |  |  |  |
| PFX-UAT-007 | Sample approval/rework | QC/Production | Pending |  |  |  |
| PFX-UAT-008 | Mass production progress | Production | Pending |  |  |  |
| PFX-UAT-009 | Finished goods receipt QC hold | Warehouse/QC | Pending |  |  |  |
| PFX-UAT-010 | Finished goods QC closeout | QC | Pending |  |  |  |
| PFX-UAT-011 | Factory claim | QC/Production | Pending |  |  |  |
| PFX-UAT-012 | Final payment readiness | Production/Finance | Pending |  |  |  |
| PFX-UAT-013 | AP handoff/payment evidence | Finance | NOT_RUN / ready after Sprint 36 PR #616 smoke |  | PFX-BLOCKER-001 closed history | S36_RUNTIME_READY_PR616 |
| PFX-UAT-014 | Negative controls | QA/UAT Lead | Pending |  |  |  |

Status values:

```text
PASS
PASS_WITH_OBSERVATION
BLOCKED
FAIL
WAIVED
NOT_RUN
```

Discovery-mode rule:

```text
For Mode B, PFX-UAT-013 is NOT_RUN / S36_RUNTIME_READY_PR616 after Sprint 36 payment voucher / cash-out evidence runtime was merged and smoke-tested in PR #616.
```

---

## 12. Observation log template

| Observation ID | Scenario ID | Role | Observation | Type | Severity | Owner | Target sprint | Decision |
|---|---|---|---|---|---|---|---|---|
| OBS-PFX-001 | PFX-UAT-001 | Production |  | Bug / CR / Training / Data | P0/P1/P2/P3 |  |  |  |

Definitions:

```text
Bug: Hệ thống sai so với requirement/rule đã duyệt.
Change Request: User muốn thêm hành vi mới hoặc rule mới.
Training Issue: User chưa quen thao tác, hệ thống không sai.
Data Issue: Seed/master data thiếu hoặc sai.
```

---

## 13. Issue triage rules

| Severity | Meaning | Release impact |
|---|---|---|
| P0 | Không chạy được luồng chính, mất dữ liệu, sai tiền/hàng nghiêm trọng | No-Go |
| P1 | Luồng chính bị nghẽn nhưng có workaround hạn chế | Conditional Go only if accepted |
| P2 | Bất tiện hoặc lỗi phụ, không chặn main flow | Can Go with fix plan |
| P3 | Copy/UI/minor improvement | Can Go |

No-Go conditions:

```text
- Không login được role chính.
- Không tính material demand đúng.
- Không trace được PR/PO/Factory order.
- Vật tư/finished goods vào available sai rule.
- QC FAIL vẫn available hoặc payment-ready.
- Factory claim không trace được về QC fail.
- Payment status không khớp cash-out evidence.
- Không có audit cho thao tác nhạy cảm.
```

---

## 14. Go / Conditional Go / No-Go criteria

### Go

```text
- Tất cả PFX-UAT-001 đến PFX-UAT-014 PASS hoặc WAIVED hợp lệ.
- Không còn P0/P1.
- P2/P3 có owner và target sprint.
- Evidence pack đầy đủ.
- Business Owner, Production Owner, QC, Warehouse, Finance/AP sign off.
```

### Conditional Go

```text
- Không còn P0.
- P1 còn lại có workaround được business owner chấp nhận.
- Owner/fix sprint rõ.
- Không có rủi ro sai hàng/sai tiền/sai QC.
```

### No-Go

```text
- Có P0.
- Có P1 không workaround.
- Thiếu evidence cho main scenarios.
- Không có sign-off.
- Hệ thống tạo hai sự thật về stock/payment/QC.
```

---

## 15. Sign-off template

```text
UAT Pack: Production External Factory E2E
Date:
Environment:
Build/Commit:
Release tag, if any:

Decision:
[ ] Go
[ ] Conditional Go
[ ] No-Go

Conditions / Open Issues:

Business Owner:
Production Owner:
Warehouse Owner:
QC Owner:
Finance/AP Owner:
UAT Lead:
ERP Product Owner:
```

---

## 16. Recommended session schedule

### Session 1 — Planning + material demand

```text
PFX-UAT-001
PFX-UAT-002
```

### Session 2 — Purchase/receiving/QC + warehouse issue

```text
PFX-UAT-003
PFX-UAT-004
```

### Session 3 — Factory execution

```text
PFX-UAT-005
PFX-UAT-006
PFX-UAT-007
PFX-UAT-008
```

### Session 4 — Finished goods + claim

```text
PFX-UAT-009
PFX-UAT-010
PFX-UAT-011
```

### Session 5 — Finance closeout

```text
PFX-UAT-012
PFX-UAT-013
PFX-UAT-014
Go/No-Go review
```

---

## 17. Final notes

Đây là tài liệu UAT cho một luồng rất quan trọng: **biến sản xuất/gia công ngoài từ một chuỗi thủ công thành một chuỗi có traceability, QC gate, stock control và finance control**.

Một câu chốt:

```text
Nhà máy sản xuất hàng, nhưng ERP phải kiểm soát sự thật:
vật tư đã đi đâu, hàng đã đạt chưa, claim còn mở không, và tiền có được phép trả chưa.
```

Không được gọi flow này là business accepted nếu chưa có evidence thật từ scenario results, issue triage và Go/No-Go report.
