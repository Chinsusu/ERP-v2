# S16-01-01 Subcontract Persistence Design

Project: Web ERP for cosmetics operations
Sprint: Sprint 16 - Subcontract runtime store persistence
Task: S16-01-01 Subcontract persistence design
Date: 2026-05-02
Status: Design complete; S16-01-02 can introduce the subcontract migration foundation

---

## 1. Goal

Promote subcontract runtime state from prototype memory to PostgreSQL-backed persistence when database configuration exists.

Current restart risk:

```text
subcontract orders                  -> productionapp.NewPrototypeSubcontractOrderStore(auditLogStore)
subcontract material transfers      -> productionapp.NewPrototypeSubcontractMaterialTransferStore()
subcontract sample approvals        -> productionapp.NewPrototypeSubcontractSampleApprovalStore()
subcontract finished goods receipts -> productionapp.NewPrototypeSubcontractFinishedGoodsReceiptStore()
subcontract factory claims          -> productionapp.NewPrototypeSubcontractFactoryClaimStore()
subcontract payment milestones      -> productionapp.NewPrototypeSubcontractPaymentMilestoneStore()
warehouse daily board subcontract   -> reads prototype-backed subcontract stores
```

Sprint 16 should make the existing subcontract APIs durable without changing public response envelopes or adding new factory-facing workflows.

Important sequencing rule:

```text
Do not wire only one subcontract store into DB mode while other subcontract stores remain memory-backed.
Implement and test stores separately, but switch runtime selection to DB mode as one order/transfer/sample/receipt/claim/milestone package.
```

---

## 2. Existing Contracts

Subcontract order store:

```text
SubcontractOrderStore
  List(ctx, filter) []domain.SubcontractOrder
  Get(ctx, id) domain.SubcontractOrder
  WithinTx(ctx, fn(ctx, tx)) error

SubcontractOrderTx
  GetForUpdate(ctx, id) domain.SubcontractOrder
  Save(ctx, order) error
  RecordAudit(ctx, log) error
```

Material transfer store:

```text
SubcontractMaterialTransferStore
  Save(ctx, transfer) error
  ListBySubcontractOrder(ctx, subcontractOrderID) []domain.SubcontractMaterialTransfer
```

Sample approval store:

```text
SubcontractSampleApprovalStore
  Save(ctx, sampleApproval) error
  Get(ctx, id) domain.SubcontractSampleApproval
  GetLatestBySubcontractOrder(ctx, subcontractOrderID) domain.SubcontractSampleApproval
```

Finished goods receipt store:

```text
SubcontractFinishedGoodsReceiptStore
  Save(ctx, receipt) error
  ListBySubcontractOrder(ctx, subcontractOrderID) []domain.SubcontractFinishedGoodsReceipt
```

Factory claim store:

```text
SubcontractFactoryClaimStore
  Save(ctx, claim) error
  Get(ctx, id) domain.SubcontractFactoryClaim
  ListBySubcontractOrder(ctx, subcontractOrderID) []domain.SubcontractFactoryClaim
```

Payment milestone store:

```text
SubcontractPaymentMilestoneStore
  Save(ctx, milestone) error
  Get(ctx, id) domain.SubcontractPaymentMilestone
  ListBySubcontractOrder(ctx, subcontractOrderID) []domain.SubcontractPaymentMilestone
```

Runtime consumers that must share the same selected stores:

```text
SubcontractOrderService
Warehouse daily board subcontract signal source
Subcontract supplier payable adapter
Full dev smoke subcontract checks
```

---

## 3. Existing Schema Baseline

Current migration `000013_subcontract_order_core` already gives Sprint 16 a PostgreSQL foundation:

```text
subcontract.subcontract_orders
subcontract.subcontract_order_material_lines
subcontract.subcontract_order_status_history
```

The existing schema covers order status, material line quantities, UOM/base UOM, money fields, status history, and actor UUID columns for many lifecycle steps.

Known persistence gaps before S16-01-02:

```text
no PostgreSQL runtime store uses the existing subcontract order tables
no durable material transfer header/line/evidence tables
no durable sample approval/evidence tables
no durable finished goods receipt header/line/evidence tables
no durable factory claim/evidence tables
no durable payment milestone table with payable-link evidence
no package-level runtime selector for subcontract stores
warehouse daily board still receives prototype subcontract stores
```

S16-01-02 should extend the schema instead of replacing the existing subcontract foundation.

---

## 4. Proposed Migration Foundation

Add migration:

```text
000032_persist_subcontract_runtime_foundation
```

Migration scope:

```text
1. Harden existing subcontract order tables for runtime refs/stable IDs if needed.
2. Add material transfer header, line, and evidence tables.
3. Add sample approval and evidence tables.
4. Add finished goods receipt header, line, and evidence tables.
5. Add factory claim and evidence tables.
6. Add payment milestone table with payable-link fields.
7. Add indexes by org/order/status/created_at and source refs used by API reads and smoke checks.
8. Keep rollback complete and ordered.
```

Expected table family:

```text
subcontract.subcontract_material_transfers
subcontract.subcontract_material_transfer_lines
subcontract.subcontract_material_transfer_evidence
subcontract.subcontract_sample_approvals
subcontract.subcontract_sample_approval_evidence
subcontract.subcontract_finished_goods_receipts
subcontract.subcontract_finished_goods_receipt_lines
subcontract.subcontract_finished_goods_receipt_evidence
subcontract.subcontract_factory_claims
subcontract.subcontract_factory_claim_evidence
subcontract.subcontract_payment_milestones
```

Column rules:

```text
money: numeric(18,2)
unit cost: numeric(18,6)
quantity: numeric(18,6)
currency: VND only
API/store decimal fields: string decimal in DTOs and domain decimal.Decimal in Go
timestamps: timestamptz
actor refs: preserve string actor refs when UUID user rows are unavailable
evidence refs: file/object/external URL fields must not require real object storage in local smoke
```

---

## 5. Store Implementation Plan

### Order Store

Implement:

```text
PostgresSubcontractOrderStore
```

Behavior:

```text
List/Get read subcontract_orders and material lines.
WithinTx opens a PostgreSQL transaction.
GetForUpdate locks the order row and loads material lines.
Save upserts the order header and material lines with stable IDs.
Save preserves existing material line IDs unless the domain intentionally removes a line.
RecordAudit writes through the existing audit log store in the same logical lifecycle path.
```

### Child Stores

Implement these stores behind existing interfaces:

```text
PostgresSubcontractMaterialTransferStore
PostgresSubcontractSampleApprovalStore
PostgresSubcontractFinishedGoodsReceiptStore
PostgresSubcontractFactoryClaimStore
PostgresSubcontractPaymentMilestoneStore
```

Shared behavior:

```text
Save is idempotent by domain ID.
ListBySubcontractOrder returns deterministic created_at/id order.
Get returns the same not-found errors as prototype stores.
Evidence and lines are replaced atomically by document ID unless the domain adds a finer update contract.
No store writes stock balance directly.
No store creates supplier payables directly; payment milestone service keeps payable-link evidence created by the adapter.
```

---

## 6. Runtime Selection Plan

Add package-level runtime selection in `apps/api/cmd/api`:

```text
type subcontractRuntimeStores struct {
  orders                productionapp.SubcontractOrderStore
  materialTransfers     productionapp.SubcontractMaterialTransferStore
  sampleApprovals       productionapp.SubcontractSampleApprovalStore
  finishedGoodsReceipts productionapp.SubcontractFinishedGoodsReceiptStore
  factoryClaims         productionapp.SubcontractFactoryClaimStore
  paymentMilestones     productionapp.SubcontractPaymentMilestoneStore
}

newRuntimeSubcontractStores(cfg, auditLogStore) -> stores, close, error
```

DB mode:

```text
Use PostgreSQL stores for every subcontract runtime store.
Open one shared DB handle for the package.
Wire the package into SubcontractOrderService, warehouse daily board signals, and any report consumers.
```

No-DB/local fallback:

```text
Use the existing prototype stores together.
Keep fallback explicit and local-only.
Never count fallback as production persistence evidence.
```

---

## 7. Test Plan

Focused backend tests:

```text
PostgresSubcontractOrderStore persists order header/material lines and reloads with fresh store.
PostgresSubcontractOrderStore transaction Save/GetForUpdate preserves lifecycle state.
PostgresSubcontractMaterialTransferStore persists header/lines/evidence and reloads with fresh store.
PostgresSubcontractSampleApprovalStore persists latest-by-order and decision state.
PostgresSubcontractFinishedGoodsReceiptStore persists receipt lines/evidence and accepted/rejected quantities.
PostgresSubcontractFactoryClaimStore persists claim/evidence and list-by-order.
PostgresSubcontractPaymentMilestoneStore persists deposit/final payment readiness and payable-link evidence.
```

Service/integration tests:

```text
Subcontract lifecycle survives fresh stores after each major action.
Package runtime selector returns all prototype stores without DATABASE_URL.
Package runtime selector returns all PostgreSQL stores with DATABASE_URL.
Warehouse daily board reads the selected subcontract stores.
Finance payable adapter still creates supplier payable evidence when final payment is ready.
```

Smoke extension:

```text
Create subcontract order.
Submit/approve/confirm factory.
Record deposit.
Issue materials.
Submit/approve sample.
Start mass production.
Receive finished goods.
Open a factory claim or partial acceptance case.
Mark final payment ready.
Restart API.
Read order and related subcontract documents back.
Check DB rows and audit rows.
Check warehouse daily board subcontract signals after restart.
```

---

## 8. Verification Gates

Per task:

```text
git diff --check
go test ./internal/modules/production/... -count=1
focused cmd/api runtime selector or handler tests when wiring changes
```

When migration changes:

```text
PostgreSQL 16 isolated apply all up migrations
PostgreSQL 16 isolated rollback all down migrations
```

Before package runtime switch:

```text
all six PostgreSQL stores exist
all six focused store/service tests pass
no partial subcontract DB selection is wired in main.go
```

Before release:

```text
dev release gate green
full dev smoke green with subcontract restart persistence evidence
remaining prototype ledger updated
Sprint 16 changelog merged
production tag pushed only after CI, dev smoke, and migration gate are green
```

---

## 9. Risks And Guardrails

Partial DB truth:

```text
Do not switch DB mode until all six stores are implemented and tested.
```

Stock correctness:

```text
Material issue and finished goods receipt services may record stock movements.
Stores must persist document evidence but must not mutate inventory balances directly.
```

Payable correctness:

```text
Payment milestones should persist payable-link evidence, but supplier payable creation remains owned by finance service.
```

Traceability:

```text
Batch/lot/evidence fields must survive restart for material issue, finished goods receipt, and factory claim flows.
```

Backward compatibility:

```text
Existing API response envelopes and frontend service contracts must remain unchanged unless OpenAPI and generated clients are updated in the same PR.
```

---

## 10. S16-01-02 Start Criteria

S16-01-02 can begin when:

```text
1. This design PR is merged.
2. Migration 000032 is scoped to extend existing subcontract schema, not replace it.
3. Rollback plan covers every added table/index/column.
4. Store implementation order follows the Sprint 16 task board.
```
