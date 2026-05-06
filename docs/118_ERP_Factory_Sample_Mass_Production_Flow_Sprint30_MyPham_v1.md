# 118_ERP_Factory_Sample_Mass_Production_Flow_Sprint30_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 30 - Factory Sample Approval And Mass Production Start
Document role: Flow design
Version: v1
Date: 2026-05-06
Status: Locked for Sprint 30 implementation

---

## 1. Flow Position

Sprint 30 sits after material handover:

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

The ERP must continue to present this as external-factory production, not internal MES production.

---

## 2. Sample Gate Contract

The factory order detail page exposes one sample approval section:

```text
/production/factory-orders/:orderId#factory-sample-approval
```

The section should show:

```text
- Sample requirement state
- Sample code
- Formula version and order spec version
- Evidence file/reference
- Sample note
- Latest submitted sample summary
- Decision reason
- Storage status for approved retained sample
- Submit, approve, and reject actions when allowed
```

Sample submit uses:

```text
POST /api/v1/subcontract-orders/{id}/submit-sample
```

Sample decision uses:

```text
POST /api/v1/subcontract-orders/{id}/approve-sample
POST /api/v1/subcontract-orders/{id}/reject-sample
```

---

## 3. Gate Rules

Sample submit can run only when:

```text
1. The order requires sample approval.
2. Materials have been issued to the factory.
3. The order is not already sample_approved or mass_production_started.
4. Sample code is filled.
5. Sample evidence reference is filled.
```

Sample decision can run only when:

```text
1. A sample has been submitted.
2. Decision reason is filled.
3. Approval also has storage status filled.
```

Rejected sample behavior:

```text
- Reject moves the order to sample_rejected.
- Mass production remains blocked.
- The user may submit a replacement sample.
```

No-sample behavior:

```text
- If sampleRequired is false, sample gate is complete after material handover.
- Mass production can start after material handover.
```

---

## 4. Mass Production Contract

The factory order detail page exposes one mass-production section:

```text
/production/factory-orders/:orderId#factory-mass-production
```

Mass production can start only when:

```text
1. Materials have been issued to the factory.
2. Sample is approved, or the order does not require a sample.
3. The order has not already started mass production.
```

Mass production start uses:

```text
POST /api/v1/subcontract-orders/{id}/start-mass-production
```

The UI updates local order state from the response so tracker and timeline move to finished-goods receipt readiness.

---

## 5. Guardrails

```text
- Do not imply automatic factory communication.
- Do not expose /subcontract as the primary production path.
- Do not create internal production routing, labor costing, work center, or MES concepts.
- Do not bypass material handover.
- Do not bypass sample approval when sampleRequired is true.
- Do not receive finished goods into available stock in this sprint.
- Finished goods from factory must still go to QC hold in follow-up flow.
- Keep backend/API/DB values English.
- Keep user-facing production copy Vietnamese.
```
