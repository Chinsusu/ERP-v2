# 49_ERP_Coding_Task_Board_Sprint6_Finance_Lite_COD_AR_AP_Core_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 6 - Finance Lite / COD / AR / AP Core
Document role: Coding task board for Sprint 6 implementation
Version: v1.0
Primary stack: Go backend, PostgreSQL, Next.js frontend, OpenAPI, MinIO/S3 attachments
Status: Ready for implementation; production tag remains on hold until GitHub Actions billing blocker is cleared

---

## 1. Sprint 6 Context

Sprint 2 controlled outbound order fulfillment.
Sprint 3 controlled returns, reconciliation, and warehouse shift close.
Sprint 4 controlled purchase inbound and QC.
Sprint 5 controlled subcontract manufacturing and factory material/finished goods flow.

Sprint 6 adds a finance-lite layer over those operational events:

```text
Sales delivery / COD
-> customer receivable
-> carrier COD remittance
-> reconciliation
-> cash receipt
-> receivable close

Purchase receiving / subcontract acceptance
-> supplier or factory payable
-> payment request
-> approval
-> cash payment
-> payable close
```

This sprint is not full accounting. It is the minimum reliable subledger needed for COD, customer debt, supplier debt, factory payment milestones, and basic cash in/out visibility.

---

## 2. Sprint 6 Theme

```text
Finance Lite + COD + AR/AP Core
```

Business reason:

```text
Operations already know when goods ship, return, receive, pass QC, or move through subcontracting.
Finance now needs traceable money status:
- COD expected versus remitted by carrier.
- Customer receivables after delivery and return adjustments.
- Supplier payables after accepted receiving/QC.
- Factory payables after subcontract acceptance or approved exception.
- Cash in/out records tied back to source documents.
```

---

## 3. Sprint 6 Goals

By the end of Sprint 6, the system must support:

```text
1. Define finance-lite roles, permissions, and audit conventions.
2. Create customer receivables from sales/COD source documents.
3. Track receivable status: draft/open/partially_paid/paid/disputed/void.
4. Record COD remittance batches by carrier.
5. Reconcile COD expected amount against remitted amount.
6. Keep COD discrepancy traceable with reason, owner, and status.
7. Record cash receipts and allocate them to receivables.
8. Create supplier payables from accepted purchase receiving/QC results.
9. Create subcontract payables from accepted finished goods and payment milestones.
10. Approve payment before cash payment is recorded.
11. Record cash payments and allocate them to payables.
12. Show Finance Lite dashboard metrics for AR, AP, COD, and cash.
13. Provide usable AR/AP/COD/payment screens in the web UI.
14. Update OpenAPI, generated FE client, and contract checks.
15. Add focused E2E coverage for COD, AP, subcontract payable, and permission/audit paths.
```

---

## 4. Sprint 6 Non-Goals

Sprint 6 does not include:

```text
- Full double-entry general ledger.
- Tax filing, VAT declaration, or e-invoice integration.
- Bank statement import or bank API integration.
- Payment gateway integration.
- Payroll or HR payout flow.
- KOL/affiliate payout automation.
- Full COGS, landed cost, or P&L allocation.
- Multi-currency.
- Automated dunning/collection workflow.
- External accountant export beyond minimal CSV/API-ready structures.
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
codex/feature-S6-xx-yy-short-task-name
```

Recommended Sprint 6 release tag after completion:

```text
v0.6.0-finance-lite-cod-ar-ap-core
```

Production tags remain on hold while GitHub Actions is blocked by the billing/spending-limit issue.
Sprint 5 release gate remains a carry-forward blocker until CI can run green on GitHub and the release tag is created.

---

## 6. Sprint 6 Demo Script

### Case 1: COD happy path

```text
1. Create sales order with COD payment method.
2. Fulfill and deliver the shipment.
3. System creates expected COD receivable.
4. Carrier remits exact COD amount.
5. Finance reconciles the remittance.
6. Cash receipt is recorded.
7. Customer receivable closes as paid.
8. Audit log captures receivable creation, remittance match, cash receipt, and close.
```

### Case 2: COD discrepancy

```text
1. Delivered COD order expects 1,000,000 VND.
2. Carrier remits 950,000 VND.
3. Reconciliation creates a 50,000 VND discrepancy.
4. Receivable remains partially paid or disputed.
5. Finance records discrepancy reason and owner.
6. Audit log captures the mismatch and follow-up status.
```

### Case 3: Supplier payable

```text
1. Purchase order is received.
2. Inbound QC PASS accepts quantity.
3. System creates supplier payable for accepted quantity only.
4. Finance submits payment request.
5. Authorized approver approves payment.
6. Cash payment is recorded.
7. Supplier payable reduces or closes.
```

### Case 4: Subcontract final payment

```text
1. Subcontract finished goods are received.
2. Inbound QC accepts pass quantity and records failed/claim quantity.
3. Final payment becomes ready only for accepted or approved-exception quantity.
4. System creates subcontract payable.
5. Finance approves and records payment.
6. Factory payable closes or remains partially open when exceptions remain.
```

---

## 7. Sprint 6 Guardrails

These rules are non-negotiable:

```text
1. Use VND as base currency.
2. Use vi-VN display and Asia/Ho_Chi_Minh business dates.
3. Never use float/double for money, quantity, rate, or percentage.
4. API money/qty/rate fields use string decimal values.
5. COD collected by carrier is not cash received until reconciliation/remittance is recorded.
6. Customer receivables must trace to source documents such as sales order, shipment, return, or adjustment.
7. Supplier payables must trace to accepted receiving/QC or approved adjustment.
8. Subcontract payables must trace to payment milestones and accepted finished goods or approved exception.
9. Returns, refunds, claims, and discrepancies adjust AR/AP through traceable documents, not direct balance edits.
10. No direct mutation of AR/AP balances outside finance document services.
11. Payment requires approval before cash payment is recorded.
12. Sensitive finance actions require finance permissions.
13. Every create/approve/reconcile/void/payment action writes audit log.
14. Do not introduce full general ledger posting in Sprint 6.
15. Sprint 6 is not production-ready until the GitHub CI blocker is cleared and release gates are green.
```

---

## 8. Dependency Map

```text
S6-00-00 Sprint 6 task board
  -> S6-01-01 finance roles and permissions
  -> S6-01-02 finance money/status foundation
  -> S6-01-03 finance audit/event conventions

S6-01-01 finance roles and permissions
  -> S6-02-02 AR API
  -> S6-03-02 COD reconciliation API
  -> S6-04-02 AP API
  -> S6-05-03 payment approval API/UI
  -> S6-09-01 permission/audit regression

S6-01-02 finance money/status foundation
  -> S6-02-01 customer receivable domain/model
  -> S6-03-01 COD remittance/reconciliation model
  -> S6-04-01 supplier payable domain/model
  -> S6-06-01 cash receipt/payment model

S6-02-01 customer receivable domain/model
  -> S6-02-02 AR API
  -> S6-02-03 AR UI
  -> S6-09-02 COD happy path E2E
  -> S6-09-03 COD discrepancy E2E

S6-03-01 COD remittance/reconciliation model
  -> S6-03-02 COD reconciliation API
  -> S6-03-03 COD reconciliation UI
  -> S6-09-02 COD happy path E2E
  -> S6-09-03 COD discrepancy E2E

S6-04-01 supplier payable domain/model
  -> S6-04-02 AP API
  -> S6-04-03 AP UI
  -> S6-09-04 supplier payable/payment E2E

S6-05-01 subcontract payable integration
  -> S6-05-02 payment approval model/service
  -> S6-05-03 payment approval API/UI
  -> S6-09-05 subcontract payable/payment E2E

S6-06-01 cash receipt/payment model
  -> S6-06-02 cash in/out API/UI
  -> S6-07-01 finance dashboard metrics
  -> S6-07-02 finance dashboard UI

S6-08-01 OpenAPI endpoints
  -> S6-08-02 generated FE client
  -> S6-08-03 contract check

S6-09-01 permission/audit regression
  -> S6-10-01 Sprint 6 release evidence
```

---

## 9. API Shape

Use slash action endpoints, not colon action endpoints.

Recommended API surface:

```text
GET    /api/v1/customer-receivables
POST   /api/v1/customer-receivables
GET    /api/v1/customer-receivables/{id}
POST   /api/v1/customer-receivables/{id}/record-receipt
POST   /api/v1/customer-receivables/{id}/mark-disputed
POST   /api/v1/customer-receivables/{id}/void

GET    /api/v1/supplier-payables
POST   /api/v1/supplier-payables
GET    /api/v1/supplier-payables/{id}
POST   /api/v1/supplier-payables/{id}/approve-payment
POST   /api/v1/supplier-payables/{id}/record-payment
POST   /api/v1/supplier-payables/{id}/void

GET    /api/v1/cod-remittances
POST   /api/v1/cod-remittances
GET    /api/v1/cod-remittances/{id}
POST   /api/v1/cod-remittances/{id}/match
POST   /api/v1/cod-remittances/{id}/submit
POST   /api/v1/cod-remittances/{id}/approve
POST   /api/v1/cod-remittances/{id}/close

GET    /api/v1/cash-transactions
POST   /api/v1/cash-transactions
GET    /api/v1/cash-transactions/{id}

GET    /api/v1/finance/dashboard
```

All request/response schemas must use decimal strings for money and quantities.

---

## 10. Task Board

| Task ID | Task | Output / Acceptance | Primary Ref |
| --- | --- | --- | --- |
| S6-00-00 | Sprint 6 task board | File 49 created with goals, guardrails, dependencies, API shape, and task list | 01, 03, 16, 19, 40, 48 |
| S6-00-01 | Sprint 5 release gate carry-forward | Changelog or release evidence clearly states GitHub CI blocker and tag hold status | 48 |
| S6-01-01 | Finance roles and permissions | Finance view/manage, COD reconcile, payment approve permissions available in backend/frontend gates | 19 |
| S6-01-02 | Finance money/status foundation | Shared finance statuses and decimal string validation for AR/AP/COD/cash docs | 40 |
| S6-01-03 | Finance audit/event conventions | Audit event names and source-document references defined for finance actions | 19 |
| S6-02-01 | Customer receivable domain/model | AR header/line/source allocation model with status and outstanding amount | 05, 17, 40 |
| S6-02-02 | AR API | List/detail/create/receipt/dispute/void endpoints with permission and audit checks | 16, 19 |
| S6-02-03 | AR UI | Finance AR list/detail with status chips, outstanding amount, receipt action | 39 |
| S6-03-01 | COD remittance/reconciliation model | Carrier remittance batch, expected/remitted amount, discrepancy model | 20, 40 |
| S6-03-02 | COD reconciliation API | Create/match/submit/approve/close endpoints with discrepancy handling | 16, 19 |
| S6-03-03 | COD reconciliation UI | COD reconciliation screen for expected, remitted, matched, discrepancy, close | 39 |
| S6-04-01 | Supplier payable domain/model | AP header/line/source allocation model for accepted purchase receiving/QC | 17, 40, 45 |
| S6-04-02 | AP API | List/detail/create/approve-payment/record-payment/void endpoints | 16, 19 |
| S6-04-03 | AP UI | Supplier payable list/detail with approval and payment actions | 39 |
| S6-05-01 | Subcontract payable integration | Accepted subcontract output/payment milestone can create payable | 47, 48 |
| S6-05-02 | Payment approval model/service | Payment approval state machine and audit for AP/subcontract payable | 19 |
| S6-05-03 | Payment approval API/UI | Approve/reject payment request from finance UI with role guard | 16, 39 |
| S6-06-01 | Cash receipt/payment model | Cash in/out transaction model tied to AR/AP/COD/payment sources | 17, 40 |
| S6-06-02 | Cash in/out API/UI | Record and view cash receipts/payments with allocations | 16, 39 |
| S6-07-01 | Finance dashboard metrics | AR overdue/open, AP due/open, COD pending/discrepancy, cash today metrics | 01, 03 |
| S6-07-02 | Finance dashboard UI | Finance Lite dashboard in web app with actionable cards/tables | 39 |
| S6-08-01 | OpenAPI endpoints | OpenAPI paths/schemas for AR/AP/COD/cash/dashboard | 16 |
| S6-08-02 | Generated FE client | Web generated client updated and consumed by finance screens | 16 |
| S6-08-03 | Contract check | OpenAPI validate/contract/generate pass on dev server | 16 |
| S6-09-01 | Permission/audit regression | Finance permissions and audit events covered by focused tests | 19, 24 |
| S6-09-02 | COD happy path E2E | SO/delivery -> COD receivable -> remittance -> receipt -> paid | 24 |
| S6-09-03 | COD discrepancy E2E | Partial remittance creates discrepancy and keeps receivable open/disputed | 24 |
| S6-09-04 | Supplier payable/payment E2E | PO receive/QC pass -> AP -> approval -> payment -> AP reduced/closed | 24, 45 |
| S6-09-05 | Subcontract payable/payment E2E | Accepted factory output -> subcontract AP -> approval -> payment | 24, 47 |
| S6-10-01 | Sprint 6 release evidence | Changelog with merged PRs, dev verification, CI status, migration status, release/tag status | 48 |

---

## 11. Definition of Done

For each code task:

```text
1. Code is scoped to the task.
2. Money, quantity, and rate fields follow file 40 decimal string rules.
3. Backend tests pass for touched services/handlers.
4. Web test/typecheck/build pass for touched UI areas.
5. OpenAPI validate/contract/generate pass when API contracts change.
6. Dev server build/test is used as the practical gate while GitHub Actions is blocked.
7. PR includes manual self-review comment.
8. Runtime changes are deployed to dev server after merge.
9. Any remaining unverified release gate is called out explicitly.
```

Sprint 6 completion requires:

```text
1. All S6 tasks merged to main.
2. Dev server smoke passes.
3. OpenAPI contract/generation passes.
4. Backend and web checks pass on dev server.
5. GitHub CI rerun is green after billing/spending-limit blocker is cleared.
6. Migration apply/rollback is verified on PostgreSQL 16 when migrations are added.
7. Tag v0.6.0-finance-lite-cod-ar-ap-core is created only after release gates are green.
```
