# 103_ERP_Subcontract_Finished_Goods_QC_Closeout_Flow_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Flow design for subcontract finished goods QC closeout
Version: v1
Date: 2026-05-06
Status: Design locked for Sprint 25 implementation

---

## 1. Decision

Phase 1 production execution remains external factory subcontracting:

```text
Production Plan creates demand and readiness.
Subcontract Order owns external factory execution.
Finished goods from factory must enter QC hold first.
Only QC accepted goods become usable finished stock.
Rejected goods open a factory claim and block clean closeout.
```

The Production Plan needs to show this downstream state without duplicating subcontract screens.

---

## 2. Source Contract

Each subcontract order created from a production plan must preserve:

```text
source_production_plan_id
source_production_plan_no
```

These fields are optional for manually created subcontract orders, but required for plan-generated orders.

Traceability links:

```text
Production Plan -> Subcontract Order
Subcontract Order -> Finished Goods Receipt
Finished Goods Receipt -> QC hold movement
QC accept/reject -> accepted stock or factory claim
Final payment readiness -> finance closeout
```

---

## 3. Status Meaning

Subcontract execution milestones:

| Step | Status | Meaning |
| --- | --- | --- |
| 1 | `draft` / `submitted` / `approved` | Factory order is being prepared and approved |
| 2 | `factory_confirmed` | Factory has confirmed execution |
| 3 | `deposit_recorded` | Deposit is recorded when required |
| 4 | `materials_issued_to_factory` | Raw material/packaging handover is done |
| 5 | `sample_submitted` / `sample_approved` / `sample_rejected` | Sample gate before mass production |
| 6 | `mass_production_started` | Factory production is running |
| 7 | `finished_goods_received` | Finished goods received into QC hold |
| 8 | `qc_in_progress` | QC decision is pending |
| 9 | `accepted` | QC passed; accepted quantity can become available stock |
| 10 | `rejected_with_factory_issue` | QC failed; factory claim is required |
| 11 | `final_payment_ready` | Final payment can proceed after accept and no blocking claim |
| 12 | `closed` | Order is fully closed |

---

## 4. UI Flow

Production Plan detail should show:

```text
Related subcontract orders
- order no
- factory
- status
- planned qty
- received qty
- accepted qty
- rejected qty
- expected receipt date
- final payment readiness
- action: open subcontract context
```

Subcontract page deep link:

```text
/subcontract?source_production_plan_id={plan_id}&search={plan_no}#subcontract-orders
```

The link should filter the order list and select the first matching order so the operator can continue receiving/QC/final payment actions from the correct plan context.

---

## 5. Backend/API Rules

```text
1. source_production_plan_id and source_production_plan_no are trimmed strings.
2. Existing manual subcontract orders can leave both fields empty.
3. Plan-generated subcontract orders should send both fields.
4. Filtering by source_production_plan_id is exact and case-insensitive.
5. Search should include source_production_plan_no.
6. PostgreSQL migration is additive and nullable.
```

---

## 6. Verification

Backend:

```text
create subcontract order with source production plan fields
list subcontract orders by source production plan id
persist and reload source production plan fields
search source production plan no
```

Frontend:

```text
plan-generated create input carries source fields
worklist opens the source-filtered subcontract URL
subcontract service filters prototype source fields
production plan detail renders related subcontract closeout rows
```

---

## 7. Known Limits

```text
1. Sprint 25 does not send factory order emails/files.
2. Sprint 25 does not calculate subcontract cost variance.
3. Sprint 25 does not create a separate Production Plan close button.
4. Sprint 25 does not tag v0.25 unless requested after dev smoke.
```
