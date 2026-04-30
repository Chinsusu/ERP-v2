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
