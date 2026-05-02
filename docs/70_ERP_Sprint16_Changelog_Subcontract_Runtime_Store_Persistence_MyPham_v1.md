# 70_ERP_Sprint16_Changelog_Subcontract_Runtime_Store_Persistence_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 16 - Subcontract runtime store persistence v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-02
Status: Release evidence complete except cloud CI; production tag is on hold while GitHub Actions minutes are exhausted

---

## 1. Sprint 16 Scope

Sprint 16 closed the highest remaining subcontract reset risk after Sprint 15:

```text
subcontract orders were prototype-only
subcontract material transfers were prototype-only
subcontract sample approvals were prototype-only
subcontract finished goods receipts were prototype-only
subcontract factory claims were prototype-only
subcontract payment milestones were prototype-only
warehouse daily board subcontract signals read prototype-backed stores
```

Promoted scope:

```text
Sprint 16 task board
subcontract persistence design
subcontract migration foundation
PostgreSQL-backed subcontract order, material transfer, sample approval, finished goods receipt, factory claim, and payment milestone stores
service/store lifecycle tests for all subcontract stores
package-level runtime subcontract store selector
warehouse daily board subcontract integration check
full dev subcontract persistence smoke through API restart
shared PostgreSQL transaction hardening for subcontract order actions
audit log ID hardening for restart-safe smoke evidence
remaining prototype store ledger update
Sprint 16 release evidence
```

No factory portal, production scheduling optimization, full manufacturing cost accounting, bank integration, new subcontract frontend screens, master data catalog persistence, or production auth/session hardening was introduced. Sprint 16 changed persistence behavior behind the existing subcontract APIs and daily board reads.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S16-00-00 Sprint 16 task board | #470 | Created Sprint 16 task board |
| S16-01-01 Subcontract persistence design | #471 | Documented subcontract persistence schema, selectors, tests, smoke, and guardrails |
| S16-01-02 Subcontract migration foundation | #472 | Added migration `000032_persist_subcontract_runtime_foundation` |
| S16-02-01 Subcontract order PostgreSQL store | #473 | Added PostgreSQL-backed subcontract order, material line, document, status, and action persistence |
| S16-03-01 Material transfer PostgreSQL store | #474 | Added PostgreSQL-backed material transfer, line, source-ref, and stock movement persistence |
| S16-04-01 Sample approval PostgreSQL store | #475 | Added PostgreSQL-backed sample submit/approve/reject persistence |
| S16-05-01 Finished goods receipt PostgreSQL store | #476 | Added PostgreSQL-backed receipt, line, accept/reject, and stock movement ref persistence |
| S16-06-01 Factory claim PostgreSQL store | #477 | Added PostgreSQL-backed defect claim, claim window, quantity impact, and status persistence |
| S16-07-01 Payment milestone PostgreSQL store | #478 | Added PostgreSQL-backed deposit/final payment readiness and supplier payable link persistence |
| S16-08-01 Package runtime selectors | #479 | Wired subcontract DB-mode selection as one order/transfer/sample/receipt/claim/payment package |
| S16-08 hardening | #480 | Shared the parent subcontract order PostgreSQL transaction with child subcontract stores |
| S16-08 smoke hardening | #481 | Hardened subcontract smoke status expectations and restart-safe audit log IDs |
| S16-09-01 Remaining prototype ledger update and S16-10-01 release evidence | #482 | Supersedes Sprint 15 remaining-store ledger, records release evidence, and records CI/tag hold |

All PRs used the manual review and merge flow. GitHub auto review and auto merge were not used.

---

## 3. Persistence Changes

### Runtime Selector

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeSubcontractStores.orders` | `PostgresSubcontractOrderStore` | `PrototypeSubcontractOrderStore` |
| `newRuntimeSubcontractStores.materialTransfers` | `PostgresSubcontractMaterialTransferStore` | `PrototypeSubcontractMaterialTransferStore` |
| `newRuntimeSubcontractStores.sampleApprovals` | `PostgresSubcontractSampleApprovalStore` | `PrototypeSubcontractSampleApprovalStore` |
| `newRuntimeSubcontractStores.finishedGoodsReceipts` | `PostgresSubcontractFinishedGoodsReceiptStore` | `PrototypeSubcontractFinishedGoodsReceiptStore` |
| `newRuntimeSubcontractStores.factoryClaims` | `PostgresSubcontractFactoryClaimStore` | `PrototypeSubcontractFactoryClaimStore` |
| `newRuntimeSubcontractStores.paymentMilestones` | `PostgresSubcontractPaymentMilestoneStore` | `PrototypeSubcontractPaymentMilestoneStore` |

DB mode selects the six subcontract stores as one package. Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000013_subcontract_order_core` | Existing subcontract order baseline tables |
| `000032_persist_subcontract_runtime_foundation` | Adds subcontract runtime tables, line tables, document refs, status events, action events, source refs, stock movement refs, supplier payable refs, indexes, and constraints |
| `000033_harden_subcontract_order_runtime_documents` | Hardens subcontract order document persistence needed by the final runtime selector and smoke gate |

Persisted behavior:

```text
GET  /api/v1/subcontract-orders
POST /api/v1/subcontract-orders
GET  /api/v1/subcontract-orders/{subcontract_order_id}
POST /api/v1/subcontract-orders/{subcontract_order_id}/submit
POST /api/v1/subcontract-orders/{subcontract_order_id}/approve
POST /api/v1/subcontract-orders/{subcontract_order_id}/confirm-factory
POST /api/v1/subcontract-orders/{subcontract_order_id}/record-deposit
POST /api/v1/subcontract-orders/{subcontract_order_id}/start-production
POST /api/v1/subcontract-material-transfers
POST /api/v1/subcontract-sample-approvals
POST /api/v1/subcontract-finished-goods-receipts
POST /api/v1/subcontract-factory-claims
POST /api/v1/subcontract-payment-milestones
GET  /api/v1/warehouse/daily-board
```

Persisted evidence:

```text
production.subcontract_orders
production.subcontract_order_material_lines
production.subcontract_order_documents
production.subcontract_order_status_events
production.subcontract_order_action_events
production.subcontract_material_transfers
production.subcontract_material_transfer_lines
production.subcontract_material_transfer_source_refs
production.subcontract_sample_approvals
production.subcontract_finished_goods_receipts
production.subcontract_finished_goods_receipt_lines
production.subcontract_factory_claims
production.subcontract_payment_milestones
audit.audit_logs subcontract lifecycle actions
```

---

## 4. Dev Release Evidence

Dev server:

```text
Host: 10.1.1.120
Repo: /opt/ERP-v2
Runtime dev URL: http://10.1.1.120:8088
```

Runtime deploy evidence:

```text
After PR #481, main was synced/deployed to dev at commit af284cf.
Full dev smoke passed after the final subcontract runtime smoke hardening.
This S16-09/S16-10 task is documentation-only and does not require a runtime redeploy.
```

Latest Sprint 16 smoke evidence on dev:

```text
subcontract_deposit 200
persisted_subcontract_order ok SCO-S16-08-03-SMOKE-0009
persisted_subcontract_flow ok
Full ERP dev smoke passed
```

The same full smoke also passed the previously persisted audit, sales reservation/order, stock adjustment/movement/count, purchase order, inbound QC, carrier manifest, pick task, pack task, return receipt, supplier rejection, and finance checks.

---

## 5. CI And Migration Evidence

GitHub Actions status:

```text
Cloud CI is blocked for the final Sprint 16 ledger/changelog PR because the GitHub Actions plan has used 100% of the included monthly minutes.
Quota message: 2,000 min used / 2,000 min included.
Do not treat Sprint 16 as production-tagged while this CI gate is blocked.
```

Local/dev verification highlights:

```text
S16-08 fix gate: dev server go test ./... -count=1 passed.
S16-08 fix gate: dev server go vet ./... passed.
S16-08 fix gate: dev deploy script passed.
S16-08 fix gate: full dev smoke passed after source sync.
S16-09-01/S16-10-01: git diff --check passed.
S16-09-01/S16-10-01: full dev smoke passed on main commit af284cf.
S16-09-01: remaining prototype ledger updated; documentation-only change.
S16-10-01: changelog created with CI blocked and tag hold recorded.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
PostgreSQL version: 16.13
Source: /opt/ERP-v2 at main commit af284cf
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
Applied migrations: 33
Rolled back migrations: 33
```

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S14-04-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 16:

```text
1. Master data catalogs.
2. Auth/session hardening.
```

Subcontract order, material transfer, sample approval, finished goods receipt, factory claim, payment milestone, and daily board subcontract reads are no longer listed as production persistence gaps when DB config exists.

---

## 7. Release Status

Sprint 16 release gate status:

```text
Task PRs: merged through S16-10-01 PR #482
Main cloud CI: blocked by exhausted GitHub Actions minutes
Dev runtime smoke: green at runtime commit af284cf
Docs-only main sync: green at commit ad3fa354
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: HOLD
```

Recommended production tag after CI is available and green:

```text
v0.16.0-subcontract-runtime-store-persistence
```

Do not create the tag until either:

```text
1. GitHub Actions minutes reset or billing is fixed, required checks run, and the checks are green; or
2. the team explicitly accepts a manual-only release gate for this sprint.
```
