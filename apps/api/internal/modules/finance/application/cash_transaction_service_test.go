package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	financedomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/finance/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"

	_ "github.com/jackc/pgx/v5/stdlib"
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

func TestPostgresCashTransactionServicePersistsPostedTransactionAcrossFreshStores(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedCashTransactionSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	now := time.Date(2026, 5, 2, 14, 30, 0, 0, time.UTC)
	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	input := baseCreateCashReceiptInput()
	input.ID = "cash-s15-05-02-" + suffix
	input.TransactionNo = "CASH-IN-S15-05-02-" + suffix
	input.CounterpartyID = "carrier-s15-05-02"
	input.ReferenceNo = "BANK-COD-S15-05-02-" + suffix
	input.RequestID = "req-cash-s15-05-02-create-" + suffix
	input.Allocations[0].ID = "cash-alloc-ar-s15-05-02-" + suffix
	input.Allocations[0].TargetID = "ar-cod-s15-05-02-" + suffix
	input.Allocations[0].TargetNo = "AR-COD-S15-05-02-" + suffix
	input.Allocations[1].ID = "cash-alloc-cod-s15-05-02-" + suffix
	input.Allocations[1].TargetID = "cod-remit-s15-05-02-" + suffix
	input.Allocations[1].TargetNo = "COD-S15-05-02-" + suffix

	newPostgresService := func() CashTransactionService {
		store := NewPostgresCashTransactionStore(
			db,
			PostgresCashTransactionStoreConfig{DefaultOrgID: testCashTransactionOrgID},
		)
		auditStore := audit.NewPostgresLogStore(
			db,
			audit.PostgresLogStoreConfig{DefaultOrgID: testCashTransactionOrgID},
		)

		return NewCashTransactionService(store, auditStore).WithClock(func() time.Time { return now })
	}

	createResult, err := newPostgresService().CreateCashTransaction(ctx, input)
	if err != nil {
		t.Fatalf("create cash transaction: %v", err)
	}
	if createResult.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	freshService := newPostgresService()
	loaded, err := freshService.GetCashTransaction(ctx, input.ID)
	if err != nil {
		t.Fatalf("get transaction by id: %v", err)
	}
	loadedByNo, err := freshService.GetCashTransaction(ctx, input.TransactionNo)
	if err != nil {
		t.Fatalf("get transaction by no: %v", err)
	}
	if loadedByNo.ID != input.ID {
		t.Fatalf("loaded by no id = %q, want %q", loadedByNo.ID, input.ID)
	}
	if loaded.Status != financedomain.CashTransactionStatusPosted ||
		loaded.Direction != financedomain.CashTransactionDirectionIn ||
		loaded.PostedBy != "finance-user" ||
		loaded.TotalAmount.String() != "1250000.00" ||
		len(loaded.Allocations) != 2 ||
		loaded.Allocations[0].ID != input.Allocations[0].ID ||
		loaded.Allocations[1].ID != input.Allocations[1].ID {
		t.Fatalf("loaded transaction = %+v, want persisted posted transaction with allocations", loaded)
	}

	filtered, err := freshService.ListCashTransactions(ctx, CashTransactionFilter{
		Search:         "COD-S15-05-02-" + suffix,
		Directions:     []financedomain.CashTransactionDirection{financedomain.CashTransactionDirectionIn},
		CounterpartyID: "carrier-s15-05-02",
	})
	if err != nil {
		t.Fatalf("list transactions: %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != input.ID {
		t.Fatalf("filtered transactions = %+v, want only %q", filtered, input.ID)
	}

	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testCashTransactionOrgID},
	)
	logs, err := auditStore.List(ctx, audit.Query{
		EntityType: string(financedomain.FinanceEntityTypeCashTransaction),
		EntityID:   input.ID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].Action != string(financedomain.FinanceAuditActionCashTransactionRecorded) ||
		logs[0].Metadata["direction"] != string(financedomain.CashTransactionDirectionIn) {
		t.Fatalf("audit logs = %+v, want recorded cash transaction audit", logs)
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
