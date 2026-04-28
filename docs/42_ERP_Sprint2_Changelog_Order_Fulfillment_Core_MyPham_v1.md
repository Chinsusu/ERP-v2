# 42_ERP_Sprint2_Changelog_Order_Fulfillment_Core_MyPham_v1

**Dự án:** Web ERP công ty mỹ phẩm  
**Giai đoạn:** Phase 1  
**Sprint:** Sprint 2 — Order Fulfillment Core  
**Vai trò tài liệu:** Changelog / release note nhánh triển khai Sprint 2.

---

## 1. Sprint 2 Kickoff

Sprint 1 foundation đã được merge/promote lên `main` và tag:

```text
v0.1.0-foundation
```

Sprint 2 branch:

```text
sprint/2-order-fulfillment-core
```

Sprint 2 task board:

```text
docs/41_ERP_Coding_Task_Board_Sprint2_Order_Fulfillment_Core_MyPham_v1.md
```

---

## 2. Sprint Goal

Sprint 2 tập trung build luồng order fulfillment lõi:

```text
Sales Order
→ Reserve Stock
→ Pick
→ Pack
→ Carrier Manifest
→ Scan Handover ĐVVC
→ Warehouse Daily Board cập nhật dữ liệu thật
```

Cuối Sprint 2 phải demo được vòng đời đơn hàng từ tạo đơn đến bàn giao cho đơn vị vận chuyển, gồm cả case thiếu đơn khi quét bàn giao.

---

## 3. Guardrails

- Không reserve/pick/pack batch QC `HOLD` hoặc `FAIL`.
- Không update tồn kho trực tiếp ngoài stock ledger/reservation service.
- Không confirm handover khi manifest chưa đủ đơn hợp lệ.
- Không bỏ audit log cho confirm/reserve/pick/pack/handover.
- Không làm UI lệch style đã chốt trong file 39.

---

## 4. Verification Note

GitHub Actions đang bị chặn bởi billing/spending limit của account, nên CI cloud chưa thể dùng làm evidence. Foundation checkpoint dùng local verification thay thế:

```text
pnpm --filter web test
pnpm --filter web typecheck
pnpm --filter web build
pnpm openapi:validate
pnpm smoke:test
go test ./...
go vet ./...
```

Khi GitHub billing được xử lý, cần rerun CI cho các branch/PR Sprint 2.

---

## 5. Change Log

| Date | Change | Evidence |
|---|---|---|
| 2026-04-28 | Opened Sprint 2 task board and kickoff docs | File 41, README Sprint 2 section |
| 2026-04-28 | Created Sprint 1 foundation tag | `v0.1.0-foundation` |
| 2026-04-28 | Created Sprint 2 base branch | `sprint/2-order-fulfillment-core` |
