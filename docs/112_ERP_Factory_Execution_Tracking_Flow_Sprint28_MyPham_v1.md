# 112_ERP_Factory_Execution_Tracking_Flow_Sprint28_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 28 - Factory Execution Tracking
Document role: Flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 28 implementation

---

## 1. Flow Position

Sprint 28 sits after factory dispatch and before finished-goods closeout:

```text
Production plan
-> Factory order
-> Dispatch pack
-> Factory confirmed
-> Deposit / payment condition
-> Material handover to factory
-> Sample gate, if required
-> Mass production at factory
-> Finished goods receipt to QC hold
-> QC pass / factory claim
-> Final payment readiness
-> Close
```

The ERP must not present this as internal production. The company currently sends all production to external factories.

---

## 2. Tracker Contract

The factory order detail page should show one worklist:

| Gate | Source state | Target action |
| --- | --- | --- |
| Dispatch | latest factory dispatch status | `/production/factory-orders/:orderId#factory-dispatch` |
| Factory confirmation | order status and dispatch response | same dispatch section |
| Deposit | deposit status and order status | `/subcontract?...#subcontract-payment` |
| Material handover | material line planned/issued quantities | `/subcontract?...#subcontract-transfer` |
| Sample gate | sample required and order status | `/subcontract?...#subcontract-sample` |
| Mass production | order status | `/subcontract?...#subcontract-workflow` |
| Finished goods receipt | received quantity and order status | `/subcontract?...#subcontract-inbound` |
| QC closeout / claim | accepted/rejected quantity and order status | `/subcontract?...#subcontract-inbound` or `#subcontract-claim` |
| Final payment | final payment status and order status | `/subcontract?...#subcontract-payment` |

---

## 3. Status Rules

```text
complete = the gate is satisfied by current runtime state.
current = this is the next actionable gate.
pending = previous gates are not done yet.
blocked = an explicit rejection/claim state prevents normal progress.
```

Important rules:

```text
- Pending deposit comes before material handover when deposit is required.
- Material handover is complete only when all material lines are fully issued, or order state has passed the material-issued gate.
- No-sample-required orders skip the sample gate.
- Sample rejected blocks mass production.
- Finished goods received from the factory remain QC hold until QC passes.
- QC rejection opens/links a factory claim and blocks normal final payment.
```

---

## 4. UI Contract

On `/production/factory-orders/:orderId`, show:

```text
Theo dõi thực thi nhà máy
- Current gate title
- Current gate description
- Current gate status chip
- Current gate metric
- Current gate action button
- Table of all execution gates with status, metric, and action link
```

This section should sit above the detailed dispatch/timeline/material table so the user sees the next operational action first.

---

## 5. Guardrails

```text
- Do not add email/Zalo/API delivery.
- Do not add internal factory/MES terminology.
- Do not claim the system automatically sends anything to the factory.
- Keep /subcontract route hidden from primary navigation but allow deep links for existing execution forms.
- Keep backend/API/DB status values English.
- Use Vietnamese display copy on production-facing UI.
```
