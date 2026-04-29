# 47_ERP_Coding_Task_Board_Sprint5_Subcontract_Manufacturing_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 5 - Subcontract Manufacturing / Gia cong ngoai
Document role: Coding task board for Sprint 5 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL, Next.js frontend, OpenAPI, MinIO/S3 attachments
Status: Ready for Sprint 5 planning

---

## 1. Sprint 5 Context

Sprint 2 controlled outbound fulfillment.
Sprint 3 controlled returns, reconciliation, and shift close.
Sprint 4 controlled purchase inbound and QC.

Sprint 5 now covers the real production workflow from the current As-Is process:

```text
Subcontract order
-> confirm quantity / spec / sample requirement
-> deposit milestone
-> issue raw materials / packaging to factory
-> material handover evidence
-> sample submission and approval
-> mass production
-> receive finished goods
-> inbound QC
-> accept / reject / factory claim within SLA
-> final payment readiness
```

The system must model subcontract manufacturing as an external factory flow, not as an internal shop-floor work order.

---

## 2. Sprint 5 Theme

```text
Subcontract Manufacturing Core
```

Business reason:

```text
Current production is factory/subcontract driven.
Raw materials and packaging can move out to the factory.
Finished goods come back to warehouse and must pass QC before available stock.
Defects must create a factory claim within the agreed 3-7 day window.
```

Sprint 5 builds on Sprint 4 inbound QC instead of duplicating receiving/QC logic.

---

## 3. Sprint 5 Goals

By the end of Sprint 5, the system must support:

```text
1. Create and approve subcontract orders.
2. Record factory confirmation for quantity/spec/sample requirement.
3. Record deposit milestone without full accounting posting.
4. Issue raw materials and packaging to factory through stock movement service.
5. Keep issued materials traceable and non-sellable.
6. Attach handover evidence such as COA, MSDS, labels, VAT invoice, and signed records.
7. Submit, approve, or reject factory sample.
8. Prevent mass production before sample approval unless an explicit override exists.
9. Receive finished goods from factory into QC hold.
10. Run inbound QC using Sprint 4 receiving/QC rules.
11. Accept pass quantity and reject/factory-claim failed quantity.
12. Track claim deadline and factory issue SLA.
13. Mark final payment readiness only after acceptance or approved exception.
14. Show subcontract signals in operations UI and daily board.
15. Update OpenAPI, generated FE client, and E2E coverage.
```

---

## 4. Sprint 5 Non-Goals

Sprint 5 does not include:

```text
- Full MRP.
- Internal production routing/work centers.
- Labor costing.
- Full AP accounting or bank payment posting.
- Supplier/factory portal.
- Automated EDI/API integration with factories.
- Advanced yield variance accounting.
- Multi-level BOM explosion beyond the minimal material issue list needed for subcontract orders.
```

---

## 5. Branch / PR / Release Rules

Current repo workflow remains:

```text
task branch
-> build/test on dev server
-> PR
-> manual self-review comment
-> manual merge into main
-> sync/deploy dev server when runtime changes require it
```

Do not use GitHub auto review or auto merge.
Do not create a long-lived sprint branch unless the team explicitly changes that policy.

Default task branch pattern:

```text
codex/feature-S5-xx-yy-short-task-name
```

Recommended Sprint 5 release tag after completion:

```text
v0.5.0-subcontract-manufacturing-core
```

Production tags remain on hold while GitHub Actions is blocked by the billing/spending-limit issue.

---

## 6. Sprint 5 Demo Script

### Case 1: Normal subcontract pass

```text
1. Create subcontract order for finished goods.
2. Confirm factory quantity/spec/sample requirement.
3. Record deposit milestone.
4. Issue raw materials and packaging to factory.
5. Upload handover evidence.
6. Submit and approve sample.
7. Start mass production.
8. Receive finished goods from factory.
9. QC PASS.
10. Available finished stock increases only after QC pass.
11. Final payment becomes ready.
12. Audit log captures approval, material issue, sample approval, receipt, QC pass, and payment readiness.
```

### Case 2: Sample rejected

```text
1. Create and approve subcontract order.
2. Factory submits sample evidence.
3. QA rejects sample with reason.
4. Mass production remains blocked.
5. Audit log captures sample rejection and reason.
```

### Case 3: Finished goods fail and factory claim

```text
1. Approved subcontract order reaches mass production.
2. Factory delivers finished goods.
3. Warehouse receives goods into QC hold.
4. QC FAIL due to quantity/spec/quality issue.
5. No available stock is created.
6. Factory claim is opened with claim deadline.
7. Final payment remains blocked.
```

### Case 4: Partial accept

```text
1. Factory delivers 100 units.
2. QC PASS 80 units and FAIL/HOLD 20 units.
3. Available stock increases only by 80.
4. The remaining 20 units are traceable as hold/claim.
5. Final payment readiness reflects unresolved exception.
```

---

## 7. Sprint 5 Guardrails

These rules are non-negotiable:

```text
1. No material issue without an approved subcontract order.
2. Material issue must go through stock movement service.
3. No direct stock balance mutation.
4. Materials issued to factory are not sellable or reservable.
5. Raw material / packaging quantity uses decimal string rules from file 40.
6. Unit cost and unit price use decimal string rules from file 40.
7. Batch/lot traceability must be preserved for issued materials where available.
8. Mass production requires sample approval unless there is an explicit audited override.
9. Finished goods received from factory enter QC hold first.
10. QC PASS is the only path to available finished stock.
11. QC FAIL creates no available stock.
12. Factory claim must track reason, evidence, owner, and SLA deadline.
13. Final payment is blocked until acceptance or approved exception.
14. Every sensitive action must write audit log.
15. Attachments use the object storage path introduced before Sprint 5.
```

---

## 8. Dependency Map

```text
S5-00-00 Sprint 5 task board
  -> S5-01-01 subcontract domain model
  -> S5-01-02 subcontract migration
  -> S5-01-03 subcontract API
  -> S5-01-04 subcontract UI shell

S5-01-03 subcontract API
  -> S5-02-01 material transfer model/service
  -> S5-03-01 sample approval model/service
  -> S5-05-01 payment milestone model

S5-02-01 material transfer model/service
  -> S5-02-02 issue materials API
  -> S5-02-03 issue materials UI
  -> S5-08-02 material issue E2E

S5-03-01 sample approval model/service
  -> S5-03-02 sample approval API
  -> S5-03-03 sample approval UI
  -> S5-08-03 sample rejection E2E

S5-04-01 finished goods receipt model/service
  -> S5-04-02 receive finished goods API
  -> S5-04-03 receive finished goods UI
  -> S5-04-04 factory claim service
  -> S5-08-04 finished goods QC E2E

S5-06-01 daily board subcontract metrics
  -> S5-06-02 daily board UI update

S5-07-01 OpenAPI update
  -> S5-07-02 generated frontend client
  -> S5-07-03 contract check

S5-08-01 permission/audit regression
  -> S5-09-01 release evidence
```

---

## 9. Sprint 5 Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
|---|---|---|---|
| S5-00-00 | Sprint 5 task board | File 47 defines scope, guardrails, tasks, demo, DoD | `20`, `45`, `46` |
| S5-00-01 | Sprint 4 release carry-forward | Billing blocker remains documented; rerun CI/tag only after owner clears billing | `46` |
| S5-01-01 | Subcontract order domain model | Header/line/status/spec/sample/payment fields; transition validation | `11`, `12` |
| S5-01-02 | Subcontract order DB migration | PostgreSQL tables, indexes, constraints, up/down migration | `17` |
| S5-01-03 | Subcontract order API | CRUD + submit/approve/confirm-factory/cancel/close endpoints | `16` |
| S5-01-04 | Subcontract order UI | List/detail/create/edit/status actions | `39` |
| S5-02-01 | Material transfer model/service | Transfer document, lines, batch trace, movement request, handover evidence | `11`, `12`, `40` |
| S5-02-02 | Issue materials API | `POST /subcontract-orders/{id}/issue-materials`; stock movement + audit | `16` |
| S5-02-03 | Issue materials UI | Material issue form with warehouse, batch, qty, UOM, evidence | `39` |
| S5-03-01 | Sample approval model/service | Sample submitted/approved/rejected records and reason/evidence | `12` |
| S5-03-02 | Sample approval API | submit-sample, approve-sample, reject-sample actions | `16` |
| S5-03-03 | Sample approval UI | Evidence upload, decision, reason, action visibility | `39` |
| S5-04-01 | Finished goods receipt model/service | Factory delivery receipt linked to subcontract order and finished item/batch | `17`, `45` |
| S5-04-02 | Receive finished goods API | Receive into QC hold; no available stock before QC pass | `16`, `45` |
| S5-04-03 | Receive finished goods UI | Factory delivery note, qty, batch, expiry, packaging, attachments | `39` |
| S5-04-04 | Factory claim service | Claim reason/evidence/SLA deadline/status; final payment blocked | `20`, `28` |
| S5-04-05 | Factory claim UI | Claim list/detail/create/update evidence and SLA view | `39` |
| S5-05-01 | Payment milestone model | Deposit recorded, final payment readiness, no full accounting posting | `04`, `19` |
| S5-05-02 | Payment milestone API/UI | record-deposit and mark-final-payment-ready with audit | `16`, `39` |
| S5-06-01 | Daily board subcontract metrics | Open orders, material issued, sample pending, factory claims, final payment ready | `07`, `20` |
| S5-06-02 | Daily board UI update | Inbound/operations board shows subcontract signals | `39` |
| S5-07-01 | OpenAPI subcontract endpoints | Add route and schema definitions using slash action style | `16` |
| S5-07-02 | Regenerate frontend API client | Generated schema compiles and services use current contract | `15`, `16` |
| S5-07-03 | Contract check | Sprint 5 endpoints and envelopes checked in script/test | `24` |
| S5-08-01 | Permission/audit regression | Denied actions create no domain side effects or audit success events | `19`, `24` |
| S5-08-02 | Material issue E2E | Approved order -> issue materials -> stock movement -> audit | `24` |
| S5-08-03 | Sample rejection E2E | Sample rejected blocks mass production and records audit | `24` |
| S5-08-04 | Finished goods pass E2E | Receive FG -> QC pass -> available stock -> final payment ready | `24` |
| S5-08-05 | Finished goods fail E2E | Receive FG -> QC fail -> no available stock -> factory claim | `24` |
| S5-08-06 | Partial accept E2E | Partial pass/hold/fail quantities produce correct stock and claim state | `24` |
| S5-09-01 | Sprint 5 release evidence | Changelog, verification commands, deploy status, release gate state | `18`, `24` |

---

## 10. API Endpoint Rules

Use slash action style, matching Sprint 2-4 conventions:

```text
GET    /api/v1/subcontract-orders
POST   /api/v1/subcontract-orders
GET    /api/v1/subcontract-orders/{id}
PATCH  /api/v1/subcontract-orders/{id}
POST   /api/v1/subcontract-orders/{id}/submit
POST   /api/v1/subcontract-orders/{id}/approve
POST   /api/v1/subcontract-orders/{id}/confirm-factory
POST   /api/v1/subcontract-orders/{id}/record-deposit
POST   /api/v1/subcontract-orders/{id}/issue-materials
POST   /api/v1/subcontract-orders/{id}/submit-sample
POST   /api/v1/subcontract-orders/{id}/approve-sample
POST   /api/v1/subcontract-orders/{id}/reject-sample
POST   /api/v1/subcontract-orders/{id}/start-mass-production
POST   /api/v1/subcontract-orders/{id}/receive-finished-goods
POST   /api/v1/subcontract-orders/{id}/report-factory-defect
POST   /api/v1/subcontract-orders/{id}/accept
POST   /api/v1/subcontract-orders/{id}/mark-final-payment-ready
POST   /api/v1/subcontract-orders/{id}/close
POST   /api/v1/subcontract-orders/{id}/cancel
```

Do not use colon action routes such as `/subcontract-orders/{id}:approve`.

---

## 11. Test Matrix

Backend:

```text
- subcontract status transition tests
- approval permission tests
- material issue validation tests
- stock movement service integration tests
- batch/UOM/decimal tests for issued materials
- sample approval/rejection tests
- mass production guard tests
- finished goods receipt QC hold tests
- factory claim SLA tests
- final payment readiness guard tests
- audit tests
```

Frontend:

```text
- subcontract order form validation
- status action visibility by role
- material issue form validation
- sample decision form
- finished goods receipt form
- claim evidence and SLA display
- daily board subcontract cards
```

E2E:

```text
- normal subcontract pass
- denied material issue
- sample rejection blocks mass production
- finished goods QC fail opens claim
- partial accept keeps non-pass qty unavailable
```

---

## 12. Acceptance Criteria

Sprint 5 is done only when:

```text
- Subcontract order lifecycle is implemented and audited.
- Material issue to factory creates controlled stock movement.
- Issued materials are not available/reservable.
- Sample approval gate blocks mass production unless audited override exists.
- Finished goods receipt starts in QC hold.
- QC pass is the only path to available finished stock.
- QC fail creates factory claim and no available stock.
- Claim deadline is tracked.
- Final payment readiness is guarded by acceptance/exception.
- UI supports the critical subcontract flow.
- OpenAPI and generated client are current.
- E2E tests cover pass, sample rejection, fail claim, and partial accept.
- Dev server build/test/deploy evidence is recorded.
- Cloud CI blocker is explicitly reported if GitHub billing is still unresolved.
```

---

## 13. Definition of Done for Each Task

A Sprint 5 task is done when:

```text
- Code merged from task branch to main by PR and manual self-review.
- Primary references have been checked.
- Decimal/UOM/currency rules from file 40 are followed.
- No direct stock mutation is introduced.
- OpenAPI is updated when API changes.
- DB migration up/down exists when schema changes.
- Audit log covers sensitive actions.
- Tests are added or the gap is explicitly justified.
- Dev server build/test evidence is captured for runtime changes.
```

---

## 14. Known Release Gate

GitHub Actions cloud CI remains blocked by account billing/spending-limit.

Until that is cleared:

```text
- Do not tag Sprint 4 or Sprint 5 as production-ready.
- Keep using dev-server build/test as the practical gate.
- Record the blocker in every PR self-review and changelog.
```

---

End of file.
