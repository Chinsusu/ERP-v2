package application

import (
	"context"
	"errors"
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

	balanceArgs := runner.calls[2].args
	if balanceArgs[6] != int64(-10) || balanceArgs[8] != int64(-10) {
		t.Fatalf("balance deltas = on_hand %v available %v, want -10 -10", balanceArgs[6], balanceArgs[8])
	}
}

func TestPostgresStockMovementStoreSkipsAuditForNonAdjustment(t *testing.T) {
	runner := &capturingTransactionRunner{}
	store := newPostgresStockMovementStoreWithRunner(runner)
	movement, err := newTestMovement(domain.MovementPurchaseReceipt)
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := store.Record(context.Background(), movement); err != nil {
		t.Fatalf("record: %v", err)
	}

	if len(runner.calls) != 3 {
		t.Fatalf("exec calls = %d, want 3", len(runner.calls))
	}
	for _, call := range runner.calls {
		if strings.Contains(call.query, "audit.audit_logs") {
			t.Fatal("non-adjustment movement wrote audit log")
		}
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
