# 111_ERP_Coding_Task_Board_Sprint28_Factory_Execution_Tracking_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 28 - Factory Execution Tracking
Document role: Coding task board
Version: v1
Date: 2026-05-06
Status: Completed and merged; PR/CI/dev deploy/browser smoke evidence recorded in file 113

---

## 1. Decision

Sprint 28 connects the post-dispatch factory order detail into one production-facing execution tracker.

Current operating model remains:

```text
Company production = external factory / subcontract manufacturing.
Production users should work from /production.
Existing subcontract runtime APIs and the hidden /subcontract execution surface remain the operational backend.
```

Sprint 27 ended at:

```text
Approved factory order -> dispatch pack -> sent evidence -> factory response -> factory_confirmed
```

Sprint 28 starts after that and answers:

```text
What is the next operational task for this factory order?
Which gate is blocking execution?
Where does the user click to process that gate?
```

---

## 2. Scope

In scope:

```text
- Add a factory execution tracker on /production/factory-orders/:orderId.
- Show one current gate / next action for the selected factory order.
- Show a worklist for dispatch, factory confirmation, deposit, material handover, sample gate, mass production, finished goods receipt, QC closeout, and final payment readiness.
- Link each executable gate to the correct production detail section or hidden /subcontract anchor.
- Keep email, Zalo, factory portal/API, and internal MES out of scope.
- Keep technical API/DB/status contracts English.
- Keep user-facing UI Vietnamese.
```

Out of scope:

```text
- New backend persistence tables.
- New external delivery channel.
- Factory portal.
- PDF/print dispatch polish.
- Internal work centers, routing, labor, or MES.
- Cost variance and GL posting.
```

---

## 3. Frontend Tasks

```text
1. Add a factory execution tracker service.
2. Compute current gate and work item status from SubcontractOrder plus latest dispatch status.
3. Add tests for material handover, sample gate skip, and sample rejection blocking.
4. Render "Theo dõi thực thi nhà máy" on /production/factory-orders/:orderId.
5. Add stable anchors for sample and factory claim execution sections in /subcontract.
6. Preserve the existing dispatch section and timeline.
```

---

## 4. Backend/API Tasks

```text
No new backend API is required for Sprint 28.

The tracker uses existing runtime state:
- SubcontractOrder status and quantity fields.
- Material line planned/issued quantities.
- Deposit status.
- Latest factory dispatch status.
- Existing links to /subcontract execution anchors.
```

Follow-up backend work may be opened later if users need persisted factory daily progress events separate from the order status model.

---

## 5. Acceptance Criteria

```text
- A factory-confirmed order with pending deposit points to deposit/payment handling.
- A factory-confirmed order with paid/not-required deposit points to material handover.
- Material handover shows how many material lines are fully issued.
- A no-sample-required order skips the sample gate after materials are issued.
- A sample-rejected order blocks mass production and links to sample handling.
- The tracker links to existing execution surfaces without exposing /subcontract as a sidebar module.
- No email/Zalo behavior is added or implied.
- Local tests/build, PR CI, self-review, merge, dev deploy, and browser smoke are recorded in the Sprint 28 changelog before closeout.
```
