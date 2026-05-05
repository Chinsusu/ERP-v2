# 109_ERP_Factory_Dispatch_Flow_Sprint27_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 27 - Factory Dispatch
Document role: Flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 27 implementation

---

## 1. Business Flow

Current production model:

```text
All production is executed by external factories.
The ERP user works in Production.
The technical runtime can continue using subcontract APIs and tables.
```

Factory dispatch sits between approved factory order and factory confirmation:

```text
Approved factory order
-> Prepare dispatch pack
-> Mark ready
-> Mark sent to factory
-> Record factory response
-> If confirmed: factory order moves to factory_confirmed
-> If revision requested/rejected: order stays before factory confirmation
```

---

## 2. Dispatch Pack Content

Each dispatch pack stores a snapshot suitable for manual copy/export/send:

```text
- Dispatch number
- Factory order id/no
- Source production plan id/no
- Factory id/code/name
- Finished item id/SKU/name
- Planned quantity and UOM
- Target start date
- Expected receipt date
- Spec summary / formula reference text
- Sample required
- Material lines: SKU, name, planned qty, UOM, lot-control requirement, note
- Operator note
- Evidence records
```

Evidence is manual metadata only:

```text
- file_name
- object_key or external_url
- note
```

No automatic delivery channel is implemented in Sprint 27.

---

## 3. Status Model

Dispatch status:

| Status | Meaning |
| --- | --- |
| `draft` | Pack created but not ready to send |
| `ready` | Pack reviewed and ready to send manually |
| `sent` | User recorded that the pack was sent to the factory |
| `confirmed` | Factory confirmed the order |
| `revision_requested` | Factory asked for changes |
| `rejected` | Factory rejected the order |
| `cancelled` | Dispatch pack cancelled |

Order status remains the production execution status. Factory confirmation still uses:

```text
subcontract_order.status = factory_confirmed
```

Only dispatch response `confirmed` should trigger that transition.

---

## 4. UI Contract

On `/production/factory-orders/:orderId`, show section:

```text
Gửi nhà máy
```

It must show:

```text
- Latest dispatch status
- Dispatch no
- Ready/sent/response timestamps
- Sent by / response by
- Factory response note
- Material line snapshot
- Evidence list
- Actions allowed for the current status
```

Actions:

```text
No dispatch: Tạo bộ gửi nhà máy
draft: Đánh dấu sẵn sàng gửi
ready: Đánh dấu đã gửi
sent: Ghi nhận phản hồi nhà máy
revision_requested: Chỉnh pack theo phản hồi rồi đánh dấu sẵn sàng gửi lại
rejected: read-only evidence until a follow-up replacement order/pack is created
confirmed: read-only evidence
```

---

## 5. Guardrails

```text
- Do not imply the ERP sent the message automatically.
- Do not advance to factory_confirmed until response is confirmed.
- Keep email/Zalo/API delivery out of this sprint.
- Keep /subcontract route hidden as technical/legacy execution.
- Keep API/DB/status technical values in English.
- Keep user-facing labels Vietnamese.
```
