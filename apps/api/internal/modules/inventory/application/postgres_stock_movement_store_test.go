package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

func TestPostgresStockMovementStoreRecordsLedgerBalanceAndAdjustmentAudit(t *testing.T) {
	runner := &capturingTransactionRunner{}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementAdjustmentOut)
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err != nil {
		t.Fatalf("record: %v", err)
	}

	if !runner.committed {
		t.Fatal("transaction was not committed")
	}
	if len(runner.calls) != 4 {
		t.Fatalf("exec calls = %d, want 4", len(runner.calls))
	}

	assertQueryContains(t, runner.calls[0].query, "INSERT INTO inventory.stock_ledger")
	assertQueryContains(t, runner.calls[1].query, "SET LOCAL erp.allow_stock_balance_write")
	assertQueryContains(t, runner.calls[2].query, "INSERT INTO inventory.stock_balances")
	assertQueryContains(t, runner.calls[2].query, "ON CONFLICT")
	assertQueryContains(t, runner.calls[3].query, "INSERT INTO audit.audit_logs")

	ledgerQuery := strings.ToLower(runner.calls[0].query)
	if strings.Contains(ledgerQuery, "update inventory.stock_ledger") ||
		strings.Contains(ledgerQuery, "delete from inventory.stock_ledger") {
		t.Fatalf("ledger write query mutates posted rows: %s", runner.calls[0].query)
	}

	ledgerArgs := runner.calls[0].args
	if argString(ledgerArgs[10]) != "10.000000" || ledgerArgs[11] != "PCS" || argString(ledgerArgs[14]) != "1.000000" {
		t.Fatalf("ledger quantity/uom args = %v, want base PCS movement with factor 1", ledgerArgs[10:15])
	}
	balanceArgs := runner.calls[2].args
	if balanceArgs[6] != "PCS" || argString(balanceArgs[7]) != "-10.000000" || argString(balanceArgs[9]) != "-10.000000" {
		t.Fatalf("balance deltas = base %v on_hand %v available %v, want PCS -10 -10", balanceArgs[6], balanceArgs[7], balanceArgs[9])
	}
	auditArgs := runner.calls[3].args
	if auditArgs[2] != "inventory.stock_movement.adjusted" || argString(auditArgs[7]) != "10.000000" || auditArgs[8] != "PCS" {
		t.Fatalf("audit args = %v, want adjusted movement metadata", auditArgs)
	}
}

func TestPostgresStockMovementStoreRecordsInboundAuditAndPositiveBalanceDelta(t *testing.T) {
	runner := &capturingTransactionRunner{}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementPurchaseReceipt)
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err != nil {
		t.Fatalf("record: %v", err)
	}

	if len(runner.calls) != 4 {
		t.Fatalf("exec calls = %d, want 4", len(runner.calls))
	}
	balanceArgs := runner.calls[2].args
	if argString(balanceArgs[7]) != "10.000000" || argString(balanceArgs[9]) != "10.000000" {
		t.Fatalf("balance deltas = on_hand %v available %v, want 10 10", balanceArgs[7], balanceArgs[9])
	}
	auditArgs := runner.calls[3].args
	if auditArgs[2] != "inventory.stock_movement.recorded" {
		t.Fatalf("audit action = %v, want recorded", auditArgs[2])
	}
}

func TestPostgresStockMovementStorePersistsSourceConversionFields(t *testing.T) {
	runner := &capturingTransactionRunner{}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementPurchaseReceipt, func(input *domain.NewStockMovementInput) {
		input.Quantity = "96.000000"
		input.BaseUOMCode = "PCS"
		input.SourceQuantity = "2.000000"
		input.SourceUOMCode = "CARTON"
		input.ConversionFactor = "48.000000"
	})
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err != nil {
		t.Fatalf("record: %v", err)
	}

	ledgerArgs := runner.calls[0].args
	if argString(ledgerArgs[10]) != "96.000000" || ledgerArgs[11] != "PCS" || argString(ledgerArgs[12]) != "2.000000" || ledgerArgs[13] != "CARTON" || argString(ledgerArgs[14]) != "48.000000" {
		t.Fatalf("ledger conversion args = %v, want 96 PCS from 2 CARTON by factor 48", ledgerArgs[10:15])
	}
}

func TestPostgresStockMovementStoreRecordsOutboundNegativeBalanceDelta(t *testing.T) {
	runner := &capturingTransactionRunner{}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementSalesIssue)
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err != nil {
		t.Fatalf("record: %v", err)
	}

	balanceArgs := runner.calls[2].args
	if argString(balanceArgs[7]) != "-10.000000" || argString(balanceArgs[9]) != "-10.000000" {
		t.Fatalf("balance deltas = on_hand %v available %v, want -10 -10", balanceArgs[7], balanceArgs[9])
	}
}

func TestPostgresStockMovementStoreAllowsNonUUIDMockActor(t *testing.T) {
	runner := &capturingTransactionRunner{}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementPurchaseReceipt, func(input *domain.NewStockMovementInput) {
		input.CreatedBy = "user-erp-admin"
	})
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err != nil {
		t.Fatalf("record: %v", err)
	}

	ledgerArgs := runner.calls[0].args
	if ledgerArgs[20] != nil {
		t.Fatalf("ledger created_by = %v, want nil for non-UUID mock actor", ledgerArgs[20])
	}
	auditArgs := runner.calls[3].args
	if auditArgs[1] != nil {
		t.Fatalf("audit actor_id = %v, want nil for non-UUID mock actor", auditArgs[1])
	}
}

func TestPostgresStockMovementStoreRollsBackWhenBalanceWriteFails(t *testing.T) {
	runner := &capturingTransactionRunner{failAt: 2}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementPurchaseReceipt)
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err == nil {
		t.Fatal("record error = nil, want failure")
	}
	if runner.committed {
		t.Fatal("transaction committed after balance write failure")
	}
	if !runner.rolledBack {
		t.Fatal("transaction was not rolled back")
	}
}

type capturedStatement struct {
	query string
	args  []any
}

type capturingTransactionRunner struct {
	calls      []capturedStatement
	failAt     int
	committed  bool
	rolledBack bool
}

func (r *capturingTransactionRunner) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context, exec statementExecutor) error,
) error {
	exec := &capturingStatementExecutor{runner: r}
	err := fn(ctx, exec)
	if err != nil {
		r.rolledBack = true
		return err
	}

	r.committed = true
	return nil
}

type capturingStatementExecutor struct {
	runner *capturingTransactionRunner
}

func (e *capturingStatementExecutor) Exec(_ context.Context, query string, args ...any) error {
	callIndex := len(e.runner.calls)
	e.runner.calls = append(e.runner.calls, capturedStatement{
		query: query,
		args:  append([]any(nil), args...),
	})

	if e.runner.failAt > 0 && e.runner.failAt == callIndex {
		return errors.New("forced statement failure")
	}

	return nil
}

func assertQueryContains(t *testing.T, query string, want string) {
	t.Helper()
	if !strings.Contains(query, want) {
		t.Fatalf("query %q does not contain %q", query, want)
	}
}

func argString(value any) string {
	return fmt.Sprint(value)
}
