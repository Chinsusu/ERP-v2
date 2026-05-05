# ERP PO Detail Page and Production Plan Timeline Design - MyPham v1

## Status

Implemented in Sprint 23 production planning / purchase bridge hardening.

Follow-up traceability hardening added:

- PO timeline links to the receiving screen with the selected PO and warehouse context.
- PO detail shows related goods receipts for the PO.
- Production Plan detail aggregates related goods receipts from its related PO list.

This document defines the UI and data boundary for:

- Purchase Order detail page.
- Production Plan detail timeline.

## Decision

Production Plan is the parent control-tower page.

Purchase Order is a child transaction page.

```text
/production/plans/[planId]
  = production plan control tower
  = plan status, material demand, related PO links, receiving/QC/subcontract readiness

/purchase/orders/[poId]
  = purchase order transaction detail
  = PO status, PO timeline, supplier, warehouse, lines, actions, attachments
```

Do not duplicate full PO details inside the Production page. Production Plan should show PO summary and links.

## Current System Facts

Production Plan currently has these statuses:

```text
draft
purchase_request_draft_created
cancelled
```

Purchase Order currently has these statuses:

```text
draft
submitted
approved
partially_received
received
closed
cancelled
rejected
```

Purchase Order status transitions:

```text
draft
  -> submitted
  -> cancelled

submitted
  -> approved
  -> rejected
  -> cancelled

approved
  -> partially_received
  -> received
  -> closed
  -> cancelled

partially_received
  -> received
  -> closed

received
  -> closed
```

## Purchase Order Detail Page

Route:

```text
/purchase/orders/[poId]
```

The page should show:

```text
Header
- PO number
- Supplier
- Warehouse
- Current PO status
- Expected date
- Total amount
- Link back to purchase list

Timeline
- Draft created
- Submitted for approval
- Approved
- Partially received
- Received
- Closed
- Cancelled
- Rejected
- Receiving action link scoped to the PO and warehouse

Line table
- SKU
- Item name
- Ordered quantity
- Received quantity
- Remaining quantity
- UOM
- Unit price
- Line amount

Actions
- Submit when status is draft
- Approve when status is submitted
- Close when status is approved, partially_received, or received
- Cancel when status is draft, submitted, or approved

Related receiving
- Goods receipt number
- Receipt status
- Receipt line count
- QC summary
- Posted date
- Link back to the receiving list scoped to the PO
```

The first implementation uses the existing Purchase Order detail API.

The API already exposes:

```text
created_at
updated_at
submitted_at
approved_at
closed_at
cancelled_at
rejected_at
cancel_reason
reject_reason
version
```

Follow-up API fields to expose later:

```text
partially_received_at
received_at
created_by
submitted_by
approved_by
partially_received_by
received_by
closed_by
cancelled_by
rejected_by
```

Until those fields are exposed, the first detail page should still reserve timeline slots for partial/received using current status and line received quantities.

## Production Plan Detail Timeline

Route:

```text
/production/plans/[planId]
```

The page already exists. It should remain the parent page and show:

```text
Header
- Production plan number
- Finished good
- Planned quantity
- Formula
- Current plan status

Timeline / Worklist
- Step 1: Plan created
- Step 2: Material demand calculated
- Step 3: Purchase Orders for missing materials
- Step 4: Receiving material
- Step 5: Inbound QC material
- Step 6: Ready for subcontract manufacturing
- Step 7: Subcontract order

Related documents
- Related PO list filtered by production plan number in PO note
- Each PO row links to /purchase/orders/[poId]
- Related goods receipt list aggregated from those related POs
```

Production Plan should not show full PO detail tables. It should show enough summary for control:

```text
PO number
Supplier
Status
Expected date
Line count
Received line count
Total amount
Open detail action
```

## Link Behavior

From Production Plan:

```text
Open Purchase
  -> /purchase?search=[planNo]
```

When a related PO is known:

```text
Open PO
  -> /purchase/orders/[poId]
```

From Purchase list:

```text
Open
  -> /purchase/orders/[poId]
```

## Acceptance Criteria

- A PO can be opened directly at `/purchase/orders/[poId]`.
- The PO detail page shows current status and a timeline.
- PO actions still work from the PO detail page.
- Purchase list "Open" navigates to the PO detail page.
- Production Plan detail shows related PO summaries for the plan number.
- Production Plan related PO rows link to PO detail pages.
- PO detail shows goods receipts referencing that PO.
- PO detail "Open receiving" opens the receiving screen with `po_id` and `warehouse_id`.
- Production Plan detail shows goods receipts aggregated from related POs.
- Production Plan detail remains the parent timeline and does not embed full PO detail.
- CI passes.
- Dev deploy passes.
- Browser smoke verifies production plan to PO detail navigation.
- Browser smoke verifies PO detail to receiving navigation.

## Out of Scope

- Full receiving detail page.
- Full inbound QC detail page.
- Full subcontract order detail page.
- Backend source document table for PO-to-production-plan relation.
- New PO status values beyond the existing domain statuses.
