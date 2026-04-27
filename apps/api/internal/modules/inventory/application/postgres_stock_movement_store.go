package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

type statementExecutor interface {
	Exec(ctx context.Context, query string, args ...any) error
}

type transactionRunner interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context, exec statementExecutor) error) error
}

type PostgresStockMovementStore struct {
	runner transactionRunner
}

func NewPostgresStockMovementStore(db *sql.DB) PostgresStockMovementStore {
	return PostgresStockMovementStore{runner: sqlTransactionRunner{db: db}}
}

func newPostgresStockMovementStoreWithRunner(runner transactionRunner) PostgresStockMovementStore {
	return PostgresStockMovementStore{runner: runner}
}

func (s PostgresStockMovementStore) Record(ctx context.Context, movement domain.StockMovement) error {
	if s.runner == nil {
		return errors.New("stock movement transaction runner is required")
	}
	if err := movement.Validate(); err != nil {
		return err
	}

	direction, err := movement.Direction()
	if err != nil {
		return err
	}
	delta, err := movement.BalanceDelta()
	if err != nil {
		return err
	}

	return s.runner.WithinTransaction(ctx, func(ctx context.Context, exec statementExecutor) error {
		if err := insertStockLedger(ctx, exec, movement, direction); err != nil {
			return err
		}
		if err := allowStockBalanceWrite(ctx, exec); err != nil {
			return err
		}
		if err := upsertStockBalance(ctx, exec, movement, delta); err != nil {
			return err
		}
		if err := insertStockMovementAudit(ctx, exec, movement, direction, delta); err != nil {
			return err
		}

		return nil
	})
}

type sqlTransactionRunner struct {
	db *sql.DB
}

func (r sqlTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context, exec statementExecutor) error,
) (err error) {
	if r.db == nil {
		return errors.New("database connection is required")
	}

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin stock movement transaction: %w", err)
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(ctx, sqlStatementExecutor{tx: tx}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit stock movement transaction: %w", err)
	}

	committed = true
	return nil
}

type sqlStatementExecutor struct {
	tx *sql.Tx
}

func (e sqlStatementExecutor) Exec(ctx context.Context, query string, args ...any) error {
	if _, err := e.tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

const insertStockLedgerSQL = `
INSERT INTO inventory.stock_ledger (
  org_id,
  movement_no,
  movement_type,
  movement_at,
  direction,
  item_id,
  batch_id,
  warehouse_id,
  bin_id,
  unit_id,
  movement_qty,
  base_uom_code,
  source_qty,
  source_uom_code,
  conversion_factor,
  stock_status,
  source_doc_type,
  source_doc_id,
  source_doc_line_id,
  reason,
  created_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
  $21
)`

const setStockBalanceWriteContextSQL = `SET LOCAL erp.allow_stock_balance_write = 'on'`

const upsertStockBalanceSQL = `
INSERT INTO inventory.stock_balances (
  org_id,
  item_id,
  batch_id,
  warehouse_id,
  bin_id,
  stock_status,
  base_uom_code,
  qty_on_hand,
  qty_reserved,
  qty_available,
  updated_by
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
  $11
)
ON CONFLICT ON CONSTRAINT uq_stock_balances_key
DO UPDATE SET
  base_uom_code = EXCLUDED.base_uom_code,
  qty_on_hand = inventory.stock_balances.qty_on_hand + EXCLUDED.qty_on_hand,
  qty_reserved = inventory.stock_balances.qty_reserved + EXCLUDED.qty_reserved,
  qty_available = inventory.stock_balances.qty_available + EXCLUDED.qty_available,
  updated_at = now(),
  updated_by = EXCLUDED.updated_by,
  version = inventory.stock_balances.version + 1`

const insertStockMovementAuditSQL = `
INSERT INTO audit.audit_logs (
  org_id,
  actor_id,
  action,
  entity_type,
  entity_id,
  after_data,
  metadata
) VALUES (
  $1,
  $2,
  $3,
  'inventory.stock_movement',
  $4,
  jsonb_build_object(
    'movement_no', $5,
    'movement_type', $6,
    'direction', $7,
    'movement_qty', $8,
    'base_uom_code', $9,
    'source_qty', $10,
    'source_uom_code', $11,
    'conversion_factor', $12,
    'stock_status', $13,
    'delta_on_hand', $14,
    'delta_reserved', $15,
    'delta_available', $16
  ),
  jsonb_build_object(
    'source_doc_type', $17,
    'source_doc_id', $18,
    'source_doc_line_id', $19,
    'reason', $20
  )
)`

func insertStockLedger(
	ctx context.Context,
	exec statementExecutor,
	movement domain.StockMovement,
	direction domain.Direction,
) error {
	err := exec.Exec(
		ctx,
		insertStockLedgerSQL,
		movement.OrgID,
		movement.MovementNo,
		string(movement.MovementType),
		movementTime(movement.MovementAt),
		string(direction),
		movement.ItemID,
		nullableUUID(movement.BatchID),
		movement.WarehouseID,
		nullableUUID(movement.BinID),
		nullableUUID(movement.UnitID),
		movement.Quantity,
		movement.BaseUOMCode.String(),
		movement.SourceQuantity,
		movement.SourceUOMCode.String(),
		movement.ConversionFactor,
		string(movement.StockStatus),
		movement.SourceDocType,
		movement.SourceDocID,
		nullableUUID(movement.SourceDocLineID),
		movement.Reason,
		movement.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("insert stock ledger movement: %w", err)
	}

	return nil
}

func allowStockBalanceWrite(ctx context.Context, exec statementExecutor) error {
	if err := exec.Exec(ctx, setStockBalanceWriteContextSQL); err != nil {
		return fmt.Errorf("allow stock balance write context: %w", err)
	}

	return nil
}

func upsertStockBalance(
	ctx context.Context,
	exec statementExecutor,
	movement domain.StockMovement,
	delta domain.BalanceDelta,
) error {
	err := exec.Exec(
		ctx,
		upsertStockBalanceSQL,
		movement.OrgID,
		movement.ItemID,
		nullableUUID(movement.BatchID),
		movement.WarehouseID,
		nullableUUID(movement.BinID),
		string(movement.StockStatus),
		movement.BaseUOMCode.String(),
		delta.OnHand,
		delta.Reserved,
		delta.Available,
		movement.CreatedBy,
	)
	if err != nil {
		return fmt.Errorf("upsert stock balance: %w", err)
	}

	return nil
}

func insertStockMovementAudit(
	ctx context.Context,
	exec statementExecutor,
	movement domain.StockMovement,
	direction domain.Direction,
	delta domain.BalanceDelta,
) error {
	err := exec.Exec(
		ctx,
		insertStockMovementAuditSQL,
		movement.OrgID,
		movement.CreatedBy,
		stockMovementAuditAction(movement),
		movement.SourceDocID,
		movement.MovementNo,
		string(movement.MovementType),
		string(direction),
		movement.Quantity,
		movement.BaseUOMCode.String(),
		movement.SourceQuantity,
		movement.SourceUOMCode.String(),
		movement.ConversionFactor,
		string(movement.StockStatus),
		delta.OnHand,
		delta.Reserved,
		delta.Available,
		movement.SourceDocType,
		movement.SourceDocID,
		nullableUUID(movement.SourceDocLineID),
		movement.Reason,
	)
	if err != nil {
		return fmt.Errorf("insert stock movement audit: %w", err)
	}

	return nil
}

func stockMovementAuditAction(movement domain.StockMovement) string {
	if movement.IsAdjustment() {
		return "inventory.stock_movement.adjusted"
	}

	return "inventory.stock_movement.recorded"
}

func movementTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func nullableUUID(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return value
}
