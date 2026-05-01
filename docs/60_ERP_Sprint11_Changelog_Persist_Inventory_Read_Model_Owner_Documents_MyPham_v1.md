# 60_ERP_Sprint11_Changelog_Persist_Inventory_Read_Model_Owner_Documents_MyPham_v1

Project: Web ERP for cosmetics operations
Phase: Phase 1
Sprint: Sprint 11 - Persist inventory read model and owner documents v1
Document role: Sprint changelog and release evidence
Version: v1.0
Date: 2026-05-01
Status: Release evidence prepared; production tag pending after this changelog merges and main CI is green

---

## 1. Sprint 11 Scope

Sprint 11 closed the highest remaining cross-store mismatch after Sprint 10:

```text
persisted stock/audit/reservation/inbound evidence
-> prototype read model or owner document resets
-> UI/API/reporting can show seed-only or stale operational truth after restart
```

Promoted scope:

```text
available stock read model
sales order owner documents
purchase order owner documents
return receipt owner documents
supplier rejection owner documents
remaining prototype store ledger
```

No new business workflow was introduced. Sprint 11 changed persistence/read behavior behind existing API envelopes.

---

## 2. Merged PRs

| Task | PR | Result |
| --- | --- | --- |
| S11-00-00 Sprint 11 task board | #405 | Created Sprint 11 task board |
| S11-01-01 Available-stock read model design | #406 | Documented PostgreSQL read-model mapping |
| S11-01-02 Available-stock PostgreSQL store | #407 | Runtime available-stock reads use PostgreSQL with prototype fallback |
| S11-01-03 Available-stock persistence smoke | #408 | Dev smoke proves available stock reflects persisted `inventory.stock_balances` |
| S11-02-01 Sales order persistence design | #409 | Documented sales order owner-document mapping |
| S11-02-02 Sales order PostgreSQL store | #410 | Sales order documents persist through lifecycle actions |
| S11-02-03 Sales order persistence smoke | #411 | Smoke proves sales order and reservation persistence |
| S11-02-04 Sales order cancelled ref migration fix | #412 | Added missing cancelled ref migration coverage |
| S11-03-01 Purchase order persistence design | #413 | Documented purchase order owner-document mapping |
| S11-03-02 Purchase order PostgreSQL store | #414 | Purchase order documents persist through lifecycle actions |
| S11-03-03 Purchase order persistence smoke | #415 | Smoke proves PO, receiving, and inbound QC traceability |
| S11-04-01 Return/rejection persistence design | #416 | Documented return and supplier rejection persistence contracts |
| S11-04-02 Return receipt PostgreSQL store | #417 | Return receipt, inspection, disposition, and attachment refs persist |
| S11-04-03 Supplier rejection PostgreSQL store | #418 | Supplier rejection header, lines, attachments, and status lifecycle persist |
| S11-04-04 Return/rejection persistence smoke | #419 | Full dev smoke proves return/rejection owner docs remain queryable |
| S11-05-01 Remaining prototype store ledger update | #420 | Superseded Sprint 10 ledger with Sprint 11 remaining-store ledger |

All PRs used the manual review and merge flow.

---

## 3. Persistence Changes

### Runtime Selectors

| Runtime path | DB mode | No-DB/local fallback |
| --- | --- | --- |
| `newRuntimeStockAvailabilityStore` | `PostgresStockAvailabilityStore` | `PrototypeStockAvailabilityStore` |
| `newRuntimeSalesOrderStore` | `PostgresSalesOrderStore` | `PrototypeSalesOrderStore` |
| `newRuntimePurchaseOrderStore` | `PostgresPurchaseOrderStore` | `PrototypePurchaseOrderStore` |
| `newRuntimeReturnReceiptStore` | `PostgresReturnReceiptStore` | `PrototypeReturnReceiptStore` |
| `newRuntimeSupplierRejectionStore` | `PostgresSupplierRejectionStore` | `PrototypeSupplierRejectionStore` |

Prototype fallback remains intentional for no-DB/local mode and is not production persistence evidence.

### New Or Extended PostgreSQL Persistence

| Migration | Purpose |
| --- | --- |
| `000020_persist_sales_orders` | Runtime-safe sales order refs, lifecycle actors/timestamps, and line refs |
| `000021_add_sales_order_cancelled_ref` | Cancelled actor ref support for sales order lifecycle persistence |
| `000022_persist_purchase_orders` | Runtime-safe purchase order refs, lifecycle actors/timestamps, and line refs |
| `000023_persist_return_receipts` | Return receipt refs plus inspections, disposition actions, and attachment metadata refs |
| `000024_persist_supplier_rejections` | Supplier rejection header, lines, attachments, and lifecycle status persistence |

No migration was added for S11-04-04, S11-05-01, or this changelog task.

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
After PR #419, main was deployed to dev.
Deploy built API, worker, and web images from source.
Deploy smoke passed.
Full host smoke passed.
```

Full dev smoke included these persisted checks:

```text
persisted_audit_login
persisted_sales_reservation
persisted_sales_order
persisted_stock_adjustment
persisted_stock_movement
persisted_available_stock
persisted_stock_count
persisted_purchase_order
persisted_inbound_qc
persisted_return_receipt
persisted_supplier_rejection
```

After PR #420, `/opt/ERP-v2` was fast-forwarded to main commit `d6422598`. PR #420 was docs-only, so no runtime redeploy was required.

---

## 5. CI And Migration Evidence

GitHub checks:

```text
PR #419: required-api, required-migration, required-openapi, required-web passed.
PR #420: e2e, required-api, required-migration, required-openapi, required-web passed.
```

Migration runtime gate:

```text
PostgreSQL 16 isolated container
Source: /opt/ERP-v2 at main commit d6422598
Action: apply every *.up.sql in order, then apply every *.down.sql in reverse order
Result: passed
```

Migration notices observed during apply were idempotent schema/constraint notices from earlier migrations. They did not fail the gate.

---

## 6. Remaining Prototype Stores

Current remaining-store ledger:

```text
docs/qa/S11-05-01_remaining_prototype_store_ledger.md
```

Highest remaining persistence candidates after Sprint 11:

```text
1. Batch/QC status persistence or adapter to inventory.batches.
2. End-of-day reconciliation.
3. Shipping manifest, pick, and pack task package.
4. Finance AR/AP/COD/cash runtime stores.
5. Subcontract runtime stores.
6. Master data catalogs and auth/session hardening.
```

---

## 7. Release Status

Sprint 11 release gate status:

```text
Task PRs: merged through S11-05-01
Current changelog PR: pending merge
Main CI: green through PR #420; rerun required for this changelog PR
Dev runtime deploy: green after PR #419
Migration apply/rollback: green on PostgreSQL 16 isolated instance
Production tag: pending
```

Recommended tag after this changelog merges and main CI is green:

```text
v0.11.0-persist-inventory-read-model-owner-documents
```

Do not move the tag once pushed. If a post-tag fix is needed, create a new patch tag instead.
