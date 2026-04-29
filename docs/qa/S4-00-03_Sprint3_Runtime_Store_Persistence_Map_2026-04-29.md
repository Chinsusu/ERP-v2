# S4-00-03 Sprint 3 Runtime Store Persistence Map

Task: S4-00-03
Date: 2026-04-29
Verifier: Codex

## Scope

Map the critical Sprint 3 runtime stores and decide whether each is PostgreSQL-backed today or explicitly carried as prototype-only with a risk owner.

This satisfies the S4-00-03 acceptance path for stores that are not yet PostgreSQL-backed:

```text
PostgreSQL-backed or explicitly documented as prototype-only with risk owner.
```

## Current Runtime Wiring

The API boot path still wires the Sprint 3 stores through prototype/in-memory implementations:

```text
apps/api/cmd/api/main.go
```

Relevant runtime stores:

```text
returnsapp.NewPrototypeReturnReceiptStore()
inventoryapp.NewPrototypeEndOfDayReconciliationStore()
inventoryapp.NewPrototypeStockCountStore()
inventoryapp.NewPrototypeStockAdjustmentStore()
inventoryapp.NewInMemoryStockMovementStore()
audit.NewPrototypeLogStore()
```

The repository already has a PostgreSQL stock movement store implementation:

```text
apps/api/internal/modules/inventory/application/postgres_stock_movement_store.go
```

However, the API is not yet wired to a real database connection. The API module currently has no PostgreSQL driver dependency and `main.go` does not open `DATABASE_URL`.

## Persistence Status

| Area | Runtime implementation today | Schema coverage today | S4-00-03 status | Risk owner | Required follow-up |
|---|---|---|---|---|---|
| Return receipts | `PrototypeReturnReceiptStore.records` | `returns.return_orders`, `returns.return_order_lines` exist and were hardened by migration 000011 | Prototype-only runtime | BE / Tech Lead | Add PostgreSQL return receipt repository and wire API boot to DB-backed store |
| Return inspections | `PrototypeReturnReceiptStore.inspections` | Partial: `returns.return_dispositions` can hold inspection/disposition outcome, but no dedicated Sprint 3 inspection table exists | Prototype-only runtime | BE / QA Lead | Decide dedicated inspection table vs extending `returns.return_dispositions`; then persist inspection events |
| Return dispositions | `PrototypeReturnReceiptStore.dispositions` | `returns.return_dispositions` exists with older disposition vocabulary | Prototype-only runtime | BE / Tech Lead | Align Sprint 3 disposition codes with DB constraint and persist disposition action records |
| Return attachment metadata | `PrototypeReturnReceiptStore.attachments` | `file.attachments` exists | Prototype-only runtime | BE / DevOps | Wire S3/MinIO storage and persist attachment metadata with return inspection entity linkage |
| Return stock movements | `NewInMemoryStockMovementStore()` in API boot | `inventory.stock_ledger`, `inventory.stock_balances`, and PostgreSQL store implementation exist | Prototype-only runtime in API boot | BE / Tech Lead | Open DB connection and replace in-memory movement store with `PostgresStockMovementStore` |
| Stock counts | `PrototypeStockCountStore` | `inventory.stock_counts` exists, but Sprint 3 count lines are not fully represented | Prototype-only runtime | BE / QA Lead | Add stock count line schema/repository or document approved simplified storage model |
| Stock adjustments | `PrototypeStockAdjustmentStore` | No dedicated Sprint 3 stock adjustment table identified; posted movement can be represented in ledger | Prototype-only runtime | BE / Tech Lead | Add adjustment header/line schema and repository before production use |
| End-of-day reconciliations | `PrototypeEndOfDayReconciliationStore` | `inventory.warehouse_daily_closings` exists, but Sprint 3 checklist/line detail is not fully represented | Prototype-only runtime | BE / Warehouse Lead | Add reconciliation detail tables or approved JSONB detail column and repository |
| Audit logs | `audit.NewPrototypeLogStore()` in API boot | `audit.audit_logs` exists | Prototype-only runtime | BE / Security Owner | Add PostgreSQL audit log store and wire all sensitive actions to it |

## Release Impact

Sprint 3 remains demo-ready but not production-persistent for the stores above.

Restart impact:

```text
Return receipt, inspection, disposition, attachment metadata, stock count,
stock adjustment, reconciliation, and audit demo state can reset on API restart.
```

Inventory impact:

```text
No direct stock balance mutation is introduced by this documentation task.
The existing movement-service guardrail remains the required path for stock changes.
```

Production gate:

```text
Do not tag v0.3.0-returns-reconciliation-core as production-ready until
the prototype-only runtime stores are either PostgreSQL-backed or formally
accepted as non-production demo scope by PM + Tech Lead.
```

## Recommended Implementation Order

1. Add database connection wiring in API boot using `DATABASE_URL`.
2. Wire `PostgresStockMovementStore` for return disposition and adjustment posting.
3. Add PostgreSQL audit log store and replace `audit.NewPrototypeLogStore()` in API boot.
4. Add PostgreSQL return receipt repository covering receipt, lines, inspection, disposition, and attachment metadata.
5. Add stock count and stock adjustment persistence.
6. Add end-of-day reconciliation persistence.
7. Add integration tests for API restart persistence and migration rollback.

## Verification Performed

Commands executed on the dev server checkout:

```text
git grep -n 'New.*Memory\|memory\|in-memory\|prototype\|Store' -- apps/api/internal apps/api/cmd/api
git grep -n 'return.*store\|stock count\|stock adjustment\|reconciliation\|inspection\|disposition\|attachment' -- apps/api/cmd/api apps/api/internal/modules
git grep -n 'CREATE TABLE.*return\|return_\|stock_counts\|stock_adjustments\|reconciliation\|attachment' -- apps/api/migrations/*.sql
git grep -n 'database/sql' -- apps/api
git grep -n 'pgx\|lib/pq\|postgres' -- apps/api
```

Result:

```text
Prototype runtime stores are still wired in API boot.
PostgreSQL schema coverage is partial.
PostgreSQL stock movement store exists but is not wired in API boot.
No PostgreSQL driver dependency or DATABASE_URL opening path exists in API boot.
```
