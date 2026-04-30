package application

import (
	"context"
	"errors"
	"testing"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
)

func TestCashTransactionServiceCreatesPostedTransactionAndAudit(t *testing.T) {
	service, auditStore := newTestCashTransactionService()

	result, err := service.CreateCashTransaction(context.Background(), baseCreateCashReceiptInput())
	if err != nil {
		t.Fatalf("create cash transaction: %v", err)
	}
	if result.CashTransaction.Status != financedomain.CashTransactionStatusPosted ||
		result.CashTransaction.Direction != financedomain.CashTransactionDirectionIn {
		t.Fatalf("transaction = %+v", result.CashTransaction)
	}
	if result.CashTransaction.PostedBy != "finance-user" || result.AuditLogID == "" {
		t.Fatalf("posted/audit = %+v audit=%q", result.CashTransaction, result.AuditLogID)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionCashTransactionRecorded)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].EntityType != string(financedomain.FinanceEntityTypeCashTransaction) ||
		logs[0].Metadata["direction"] != "cash_in" {
		t.Fatalf("audit logs = %+v", logs)
	}
}

func TestCashTransactionServiceFiltersAndGetsPrototypeTransactions(t *testing.T) {
	service := NewCashTransactionService(NewPrototypeCashTransactionStore(), audit.NewInMemoryLogStore())

	transactions, err := service.ListCashTransactions(context.Background(), CashTransactionFilter{
		Search:     "SUP",
		Directions: []financedomain.CashTransactionDirection{financedomain.CashTransactionDirectionOut},
	})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(transactions) != 1 || transactions[0].ID != "cash-out-260430-0002" {
		t.Fatalf("transactions = %+v", transactions)
	}

	transaction, err := service.GetCashTransaction(context.Background(), "cash-in-260430-0001")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if transaction.Direction != financedomain.CashTransactionDirectionIn {
		t.Fatalf("transaction = %+v", transaction)
	}
}

func TestCashTransactionServiceMapsValidationAndNotFoundErrors(t *testing.T) {
	service, _ := newTestCashTransactionService()
	input := baseCreateCashReceiptInput()
	input.TotalAmount = "1000000.01"

	_, err := service.CreateCashTransaction(context.Background(), input)
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeCashTransactionValidation {
		t.Fatalf("code = %q, want validation", appErr.Code)
	}

	_, err = service.GetCashTransaction(context.Background(), "missing-cash")
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeCashTransactionNotFound {
		t.Fatalf("code = %q, want not found", appErr.Code)
	}
}

func newTestCashTransactionService() (CashTransactionService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	store := &PrototypeCashTransactionStore{records: make(map[string]financedomain.CashTransaction)}
	service := NewCashTransactionService(store, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)
	})

	return service, auditStore
}

func baseCreateCashReceiptInput() CreateCashTransactionInput {
	return CreateCashTransactionInput{
		ID:               "cash-handler-0001",
		TransactionNo:    "CASH-IN-HANDLER-0001",
		Direction:        string(financedomain.CashTransactionDirectionIn),
		BusinessDate:     "2026-04-30",
		CounterpartyID:   "carrier-ghn",
		CounterpartyName: "GHN COD",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "BANK-COD-HANDLER-0001",
		TotalAmount:      "1250000.00",
		CurrencyCode:     "VND",
		ActorID:          "finance-user",
		RequestID:        "req-cash-create",
		Allocations: []CashTransactionAllocationInput{
			{
				ID:         "cash-handler-0001-line-1",
				TargetType: string(financedomain.CashAllocationTargetCustomerReceivable),
				TargetID:   "ar-cod-260430-0001",
				TargetNo:   "AR-COD-260430-0001",
				Amount:     "1000000.00",
			},
			{
				ID:         "cash-handler-0001-line-2",
				TargetType: string(financedomain.CashAllocationTargetCODRemittance),
				TargetID:   "cod-remit-260430-0001",
				TargetNo:   "COD-REMIT-260430-0001",
				Amount:     "250000.00",
			},
		},
	}
}
