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

func TestCustomerReceivableServiceCreatesReceivableAndAudit(t *testing.T) {
	service, auditStore := newTestCustomerReceivableService()

	result, err := service.CreateCustomerReceivable(context.Background(), baseCreateCustomerReceivableInput())
	if err != nil {
		t.Fatalf("create receivable: %v", err)
	}
	if result.CustomerReceivable.Status != financedomain.ReceivableStatusOpen {
		t.Fatalf("status = %q, want open", result.CustomerReceivable.Status)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionReceivableCreated)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].EntityType != string(financedomain.FinanceEntityTypeCustomerReceivable) ||
		logs[0].Metadata["source_type"] != "shipment" {
		t.Fatalf("audit log = %+v, want finance AR source metadata", logs[0])
	}
}

func TestCustomerReceivableServiceRecordsReceipt(t *testing.T) {
	service, auditStore := newTestCustomerReceivableService()
	created, err := service.CreateCustomerReceivable(context.Background(), baseCreateCustomerReceivableInput())
	if err != nil {
		t.Fatalf("create receivable: %v", err)
	}

	result, err := service.RecordCustomerReceivableReceipt(context.Background(), CustomerReceivableActionInput{
		ID:        created.CustomerReceivable.ID,
		Amount:    "250000.00",
		ActorID:   "finance-user",
		RequestID: "req-receipt",
	})
	if err != nil {
		t.Fatalf("record receipt: %v", err)
	}
	if result.PreviousStatus != financedomain.ReceivableStatusOpen ||
		result.CurrentStatus != financedomain.ReceivableStatusPartiallyPaid {
		t.Fatalf("status transition = %q -> %q", result.PreviousStatus, result.CurrentStatus)
	}
	if result.CustomerReceivable.PaidAmount != "250000.00" ||
		result.CustomerReceivable.OutstandingAmount != "1000000.00" {
		t.Fatalf("amounts = paid %q outstanding %q", result.CustomerReceivable.PaidAmount, result.CustomerReceivable.OutstandingAmount)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionReceivableReceiptRecorded)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["receipt_amount"] != "250000.00" {
		t.Fatalf("audit logs = %+v, want receipt amount metadata", logs)
	}
}

func TestCustomerReceivableServiceMapsValidationErrors(t *testing.T) {
	service, _ := newTestCustomerReceivableService()
	input := baseCreateCustomerReceivableInput()
	input.TotalAmount = "1250000.001"

	_, err := service.CreateCustomerReceivable(context.Background(), input)
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeCustomerReceivableValidation {
		t.Fatalf("code = %q, want %q", appErr.Code, ErrorCodeCustomerReceivableValidation)
	}
}

func TestCustomerReceivableServiceReturnsNotFound(t *testing.T) {
	service, _ := newTestCustomerReceivableService()

	_, err := service.GetCustomerReceivable(context.Background(), "missing-ar")
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeCustomerReceivableNotFound {
		t.Fatalf("code = %q, want not found", appErr.Code)
	}
}

func TestPostgresCustomerReceivableServicePersistsLifecycleAcrossFreshStores(t *testing.T) {
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

	if err := seedCustomerReceivableSmokeOrg(ctx, db); err != nil {
		t.Fatalf("seed org: %v", err)
	}

	suffix := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	input := baseCreateCustomerReceivableInput()
	input.ID = "ar-s15-02-02-" + suffix
	input.ReceivableNo = "AR-S15-02-02-" + suffix
	input.SourceDocument.ID = "shipment-s15-02-02-" + suffix
	input.SourceDocument.No = "SHP-S15-02-02-" + suffix
	input.RequestID = "req-ar-s15-02-02-create-" + suffix
	input.Lines[0].ID = "ar-line-s15-02-02-" + suffix
	input.Lines[0].SourceDocument = input.SourceDocument

	newService := func(now time.Time) CustomerReceivableService {
		store := NewPostgresCustomerReceivableStore(
			db,
			PostgresCustomerReceivableStoreConfig{DefaultOrgID: testCustomerReceivableOrgID},
		)
		auditStore := audit.NewPostgresLogStore(
			db,
			audit.PostgresLogStoreConfig{DefaultOrgID: testCustomerReceivableOrgID},
		)
		return NewCustomerReceivableService(store, auditStore).WithClock(func() time.Time {
			return now
		})
	}

	created, err := newService(time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC)).
		CreateCustomerReceivable(ctx, input)
	if err != nil {
		t.Fatalf("create receivable: %v", err)
	}

	reloaded, err := newService(time.Date(2026, 5, 2, 9, 5, 0, 0, time.UTC)).
		GetCustomerReceivable(ctx, created.CustomerReceivable.ID)
	if err != nil {
		t.Fatalf("reload created receivable: %v", err)
	}
	if reloaded.ID != input.ID ||
		reloaded.SourceDocument.ID != input.SourceDocument.ID ||
		len(reloaded.Lines) != 1 ||
		reloaded.Lines[0].ID != input.Lines[0].ID {
		t.Fatalf("reloaded receivable = %+v, want fresh store reload with source and line refs", reloaded)
	}

	receipt, err := newService(time.Date(2026, 5, 2, 9, 10, 0, 0, time.UTC)).
		RecordCustomerReceivableReceipt(ctx, CustomerReceivableActionInput{
			ID:        input.ID,
			Amount:    "250000.00",
			ActorID:   "finance-user",
			RequestID: "req-ar-s15-02-02-receipt-" + suffix,
		})
	if err != nil {
		t.Fatalf("record receipt: %v", err)
	}
	if receipt.CurrentStatus != financedomain.ReceivableStatusPartiallyPaid {
		t.Fatalf("receipt status = %q, want partially_paid", receipt.CurrentStatus)
	}

	disputed, err := newService(time.Date(2026, 5, 2, 9, 15, 0, 0, time.UTC)).
		MarkCustomerReceivableDisputed(ctx, CustomerReceivableActionInput{
			ID:        input.ID,
			Reason:    "Carrier remittance mismatch",
			ActorID:   "finance-user",
			RequestID: "req-ar-s15-02-02-dispute-" + suffix,
		})
	if err != nil {
		t.Fatalf("mark disputed: %v", err)
	}
	if disputed.CurrentStatus != financedomain.ReceivableStatusDisputed {
		t.Fatalf("dispute status = %q, want disputed", disputed.CurrentStatus)
	}

	voided, err := newService(time.Date(2026, 5, 2, 9, 20, 0, 0, time.UTC)).
		VoidCustomerReceivable(ctx, CustomerReceivableActionInput{
			ID:        input.ID,
			Reason:    "Duplicate finance document",
			ActorID:   "finance-user",
			RequestID: "req-ar-s15-02-02-void-" + suffix,
		})
	if err != nil {
		t.Fatalf("void receivable: %v", err)
	}
	if voided.CurrentStatus != financedomain.ReceivableStatusVoid {
		t.Fatalf("void status = %q, want void", voided.CurrentStatus)
	}

	final, err := newService(time.Date(2026, 5, 2, 9, 25, 0, 0, time.UTC)).
		GetCustomerReceivable(ctx, input.ReceivableNo)
	if err != nil {
		t.Fatalf("reload voided receivable: %v", err)
	}
	if final.Status != financedomain.ReceivableStatusVoid ||
		final.PaidAmount.String() != "250000.00" ||
		final.OutstandingAmount.String() != "1000000.00" ||
		final.DisputeReason != "Carrier remittance mismatch" ||
		final.VoidReason != "Duplicate finance document" ||
		final.Version != 4 {
		t.Fatalf("final receivable = %+v, want persisted receipt/dispute/void lifecycle", final)
	}

	auditStore := audit.NewPostgresLogStore(
		db,
		audit.PostgresLogStoreConfig{DefaultOrgID: testCustomerReceivableOrgID},
	)
	logs, err := auditStore.List(ctx, audit.Query{
		EntityType: string(financedomain.FinanceEntityTypeCustomerReceivable),
		EntityID:   input.ID,
		Limit:      10,
	})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 4 {
		t.Fatalf("audit logs = %d, want create, receipt, dispute, void", len(logs))
	}
}

func newTestCustomerReceivableService() (CustomerReceivableService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	store := &PrototypeCustomerReceivableStore{records: make(map[string]financedomain.CustomerReceivable)}
	service := NewCustomerReceivableService(store, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 9, 0, 0, 0, time.UTC)
	})

	return service, auditStore
}

func baseCreateCustomerReceivableInput() CreateCustomerReceivableInput {
	source := SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeShipment),
		ID:   "shipment-260430-0001",
		No:   "SHP-260430-0001",
	}

	return CreateCustomerReceivableInput{
		ID:             "ar-260430-0001",
		ReceivableNo:   "AR-260430-0001",
		CustomerID:     "customer-hcm-001",
		CustomerCode:   "KH-HCM-001",
		CustomerName:   "My Pham HCM Retail",
		SourceDocument: source,
		TotalAmount:    "1250000.00",
		CurrencyCode:   "VND",
		DueDate:        "2026-05-03",
		ActorID:        "finance-user",
		RequestID:      "req-ar-create",
		Lines: []CustomerReceivableLineInput{
			{
				ID:             "ar-line-1",
				Description:    "COD delivered goods",
				SourceDocument: source,
				Amount:         "1250000.00",
			},
		},
	}
}
