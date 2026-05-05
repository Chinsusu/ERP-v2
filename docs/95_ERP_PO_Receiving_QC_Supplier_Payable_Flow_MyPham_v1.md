# 95_ERP_PO_Receiving_QC_Supplier_Payable_Flow_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 23 follow-up - Post-PO payable traceability
Version: v1
Date: 2026-05-05
Status: Implementation design and runtime behavior lock

---

## 1. Goal

Close the operational gap after PO by making the usable runtime flow explicit:

```text
Purchase Request
-> Purchase Order
-> Goods Receipt
-> Inbound QC / QC status
-> Posted stock movement
-> Supplier Payable
-> Payment request / approval / payment recording
```

The business rule is:

```text
Only received lines that are QC PASS create supplier payable value.
QC HOLD / FAIL lines must not create payable value until accepted.
```

This prevents finance from paying for material that has not been accepted into usable stock.

---

## 2. Runtime Behavior

When a goods receipt linked to a PO is posted:

```text
1. Receiving validates the PO link and receipt line references.
2. Receiving validates batch and QC data.
3. Receiving posts stock movements.
4. Lines with QC PASS are priced from the matching PO line.
5. The system creates one supplier payable for the posted receipt if accepted value is greater than zero.
6. Supplier payable header source is the warehouse receipt.
7. Supplier payable lines trace back to the PO.
8. PO timeline links to Finance AP using the PO number as search context.
```

If no line is QC PASS:

```text
No supplier payable is created.
The receipt may still post stock into QC hold depending on line QC status.
Finance must wait for accepted QC evidence before payable value is created.
```

---

## 3. Traceability Contract

Supplier payable header:

```text
source_document.type = warehouse_receipt
source_document.id = goods receipt id
source_document.no = goods receipt number
```

Supplier payable line:

```text
source_document.type = purchase_order
source_document.id = PO id
source_document.no = PO number
```

This gives both directions:

```text
PO detail -> Finance AP search by PO number
Finance AP -> source receipt
Finance AP line -> source PO
Receipt audit -> created payable id/no
```

---

## 4. Current Scope

Implemented scope:

```text
- Auto-create AP after posted PO-linked goods receipt.
- Include only QC PASS receipt lines in payable amount.
- Price accepted lines from PO unit price.
- Keep supplier identity from the PO.
- Default AP status is open.
- Default AP due date is PO expected date + 7 days.
- Finance AP search matches payable header source and payable line source.
- PO timeline includes AP step and deep link to Finance AP search.
```

Out of scope for this increment:

```text
- Supplier invoice capture.
- Three-way match: PO vs GRN vs supplier invoice.
- Tax/VAT allocation.
- Payment terms by supplier master data.
- Landed cost and freight allocation.
- Automatic payable reversal for later QC fail after AP creation.
- Supplier portal/email PO dispatch.
```

---

## 5. Acceptance Criteria

```text
1. Posting a PO-linked receipt with QC PASS lines creates a supplier payable.
2. Posting a PO-linked receipt with only QC HOLD / FAIL lines does not create a supplier payable.
3. Created payable amount equals accepted receipt quantity multiplied by matching PO unit price.
4. Created payable source header points to the warehouse receipt.
5. Created payable line source points to the PO.
6. Finance AP search by PO number finds the payable via line source.
7. PO timeline shows the AP step and opens Finance with AP search context.
8. Receiving audit includes supplier payable id/no when AP is created.
```

---

## 6. Next Decision

The next finance hardening step should be supplier invoice and three-way match:

```text
PO approved amount
vs goods receipt accepted amount
vs supplier invoice amount
```

Until that exists, supplier payable is created from accepted receipt value, not from supplier invoice evidence.
