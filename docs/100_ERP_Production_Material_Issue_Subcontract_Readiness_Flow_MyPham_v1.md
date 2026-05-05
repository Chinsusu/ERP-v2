# 100_ERP_Production_Material_Issue_Subcontract_Readiness_Flow_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Document role: Flow design for Sprint 24 production-plan material issue and subcontract readiness
Version: v1
Date: 2026-05-05
Status: Design locked for Sprint 24 planning; runtime implementation pending

---

## 1. Purpose

This document defines how a production plan should create warehouse issue evidence for raw material and packaging before external subcontract execution.

The core decision:

```text
Production Plan calculates material demand.
Purchase/Receiving/QC make missing material available.
Warehouse Issue Note posts the actual material issue.
Subcontract execution starts only after material issue readiness is complete or explicitly waived.
```

This keeps planning, purchasing, warehouse stock, and factory execution separate.

---

## 2. End-To-End Flow

```text
1. Create Production Plan
2. Snapshot formula
3. Calculate material demand
4. Compare demand with available stock
5. For shortage:
   Production Plan -> Purchase Request -> PO -> Receiving -> QC PASS -> Available stock
6. For ready stock:
   Production Plan -> Warehouse Issue Note draft
7. Warehouse Issue Note:
   Draft -> Submitted -> Approved -> Posted
8. Posted issue:
   records warehouse_issue stock movement
   reduces available stock
   links back to production plan
9. Production Plan readiness:
   all required issue posted or waived -> ready for subcontract
10. Subcontract Order can be created
```

Key rule:

```text
Shortage is a purchasing problem.
Ready stock is a warehouse issue problem.
```

---

## 3. Production Plan Readiness Terms

Each production plan material demand line should expose:

```text
required_qty
available_qty
issued_qty
remaining_to_issue_qty
shortage_qty
issue_status
```

Recommended statuses:

```text
shortage
ready_to_issue
issue_draft
issue_submitted
issue_approved
partially_issued
issued
waived
blocked
```

Status meaning:

| Status | Meaning |
| --- | --- |
| `shortage` | Required quantity is greater than available issue-ready stock |
| `ready_to_issue` | Available issue-ready stock can cover remaining quantity |
| `issue_draft` | A linked issue note exists but is not submitted |
| `issue_submitted` | A linked issue note is waiting approval |
| `issue_approved` | A linked issue note is approved but not posted |
| `partially_issued` | Some quantity has posted, remaining quantity still exists |
| `issued` | Required issue quantity has posted |
| `waived` | Business approved proceeding without full issue |
| `blocked` | Missing data, invalid UOM, QC hold/fail, or permission issue blocks action |

---

## 4. Source-Link Contract

Warehouse Issue Note lines created from a production plan must preserve the source relation.

Issue line source:

```text
source_document_type = production_plan
source_document_id = production_plan.id
source_document_line_id = material_demand_line.id
```

Optional source metadata:

```text
source_document_no = production_plan.plan_no
source_item_sku = component SKU
source_formula_version = formula version snapshot
source_required_qty = demand required quantity
```

This enables:

```text
Production Plan -> related Warehouse Issue lines
Warehouse Issue -> source Production Plan
Stock Movement -> Warehouse Issue -> Production Plan
Future costing -> issued material evidence
```

---

## 5. Quantity Rules

Issue quantity must be decimal-safe and UOM-safe.

Rules:

```text
1. issue_qty > 0
2. issue_qty <= remaining_to_issue_qty unless business explicitly supports over-issue; over-issue is out of scope for Sprint 24
3. issue_qty <= available_issue_ready_qty
4. available_issue_ready_qty excludes QC_HOLD and QC_FAIL stock
5. batch-controlled items must keep batch/lot reference
6. base UOM must match demand line UOM or use approved conversion
7. posted issue quantity contributes to issued_qty
```

Partial issue:

```text
Partial issue is allowed only when the UI labels it clearly.
Partial issue does not mark the line as issued.
Partial issue keeps remaining_to_issue_qty visible.
Partial issue cannot unlock subcontract readiness unless all required lines are complete or waived.
```

No hidden rounding:

```text
Requirement and issue calculations must preserve up to 6 decimal places.
User display may convert kg/g/mg for readability, but API/DB contract remains decimal string.
```

---

## 6. Subcontract Readiness Gate

Subcontract order creation from a production plan should be blocked until material issue readiness is complete.

Allowed readiness:

```text
All required material lines have issue_status = issued
or all non-issued required lines have issue_status = waived with reason and actor
```

Blocked readiness:

```text
Any required line is shortage
Any required line is ready_to_issue but not issued
Any required line is issue_draft / issue_submitted / issue_approved
Any required line is partially_issued without waiver
Any required line is blocked
```

Waiver rule:

```text
Waiver is optional for Sprint 24.
If implemented, waiver must require reason, actor, timestamp, audit log, and visible timeline entry.
If not implemented, the only valid readiness is fully posted issue.
```

---

## 7. UI Flow

Production plan detail should show material issue as a first-class step.

Recommended layout:

```text
Production Plan Detail
  Header:
    plan no, output SKU, formula version, planned qty, status

  Step/Tabs:
    1. Plan info
    2. Material demand
    3. Purchase/receiving/QC
    4. Warehouse issue
    5. Subcontract readiness

  Material demand table:
    SKU
    material name
    required qty
    available qty
    issued qty
    remaining qty
    status
    action
```

Actions:

```text
shortage -> Open Purchase Request / PO context
ready_to_issue -> Create issue note
issue_draft/submitted/approved -> Open issue note
partially_issued -> Create remaining issue note or open existing notes
issued -> Open posted issue note
waived -> View waiver
blocked -> View reason
```

Warehouse Issue Note detail should show:

```text
source production plan
source demand lines
status timeline
approval/posting actor
posted movement references
link back to production plan
```

---

## 8. API Surface Direction

Preferred implementation direction:

```text
GET  /api/v1/production-plans/{plan_id}
GET  /api/v1/production-plans/{plan_id}/material-issue-readiness
POST /api/v1/production-plans/{plan_id}/warehouse-issues
POST /api/v1/production-plans/{plan_id}/material-issue-waivers
```

The existing Warehouse Issue routes remain the document lifecycle owner:

```text
GET  /api/v1/warehouse-issues
POST /api/v1/warehouse-issues
POST /api/v1/warehouse-issues/{warehouse_issue_id}/submit
POST /api/v1/warehouse-issues/{warehouse_issue_id}/approve
POST /api/v1/warehouse-issues/{warehouse_issue_id}/post
```

Boundary:

```text
Production Plan route can create source-linked issue draft.
Warehouse Issue route owns submit/approve/post.
```

---

## 9. Audit And Traceability

Recommended audit actions:

```text
production_plan.material_issue.readiness_calculated
production_plan.material_issue.issue_created
production_plan.material_issue.waived
inventory.warehouse_issue.created
inventory.warehouse_issue.submitted
inventory.warehouse_issue.approved
inventory.warehouse_issue.posted
```

Timeline events:

```text
Material demand calculated
Purchase request created for shortage
PO approved for shortage
Goods received
Inbound QC passed
Warehouse issue created
Warehouse issue submitted
Warehouse issue approved
Warehouse issue posted
Material issue completed
Subcontract readiness unlocked
```

---

## 10. Error Handling

User-facing Vietnamese error copy should distinguish:

```text
Not enough available stock
Material is waiting QC
Batch/lot is required
UOM conversion is not configured
Issue quantity is greater than remaining demand
Issue note already exists for this selected demand line
Subcontract cannot start before material issue is complete
```

Technical error codes/routes remain English.

---

## 11. Verification

Backend tests:

```text
readiness calculates shortage when available < remaining
readiness calculates ready_to_issue when available >= remaining
posted issue line rolls up to issued_qty
partial issue leaves remaining_to_issue_qty
issue creation rejects shortage
issue creation preserves source document fields
subcontract gate blocks before posted issue
subcontract gate allows after posted issue
```

Frontend tests:

```text
plan detail shows issue readiness status
ready line shows create issue action
shortage line opens purchase context instead of issue action
existing issue line opens issue note
posted issue updates display state after refresh
```

Dev smoke:

```text
create/select production plan
verify material demand
create issue from ready line
submit/approve/post issue
verify plan readiness updates
verify subcontract gate behavior
```

---

## 12. Known Limits

```text
1. Sprint 24 does not calculate costing.
2. Sprint 24 does not create finished goods receipt.
3. Sprint 24 does not dispatch the factory order externally.
4. Sprint 24 does not implement two-step in-transit material issue.
5. Sprint 24 does not implement advanced MRP allocation.
6. Waiver may remain a documented follow-up if not needed for pilot.
```

---

## 13. Decision

Sprint 24 should implement:

```text
Production Plan
-> source-linked Warehouse Issue Note
-> posted material issue readiness
-> subcontract creation gate
```

The sprint should not implement:

```text
Costing
Finished goods receipt
Factory dispatch
Internal MES
```
