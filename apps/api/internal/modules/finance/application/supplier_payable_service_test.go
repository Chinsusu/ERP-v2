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

func TestSupplierPayableServiceCreatesPayableAndAudit(t *testing.T) {
	service, auditStore := newTestSupplierPayableService()

	result, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}
	if result.SupplierPayable.Status != financedomain.PayableStatusOpen {
		t.Fatalf("status = %q, want open", result.SupplierPayable.Status)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionPayableCreated)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].EntityType != string(financedomain.FinanceEntityTypeSupplierPayable) ||
		logs[0].Metadata["source_type"] != "qc_inspection" {
		t.Fatalf("audit log = %+v, want finance AP source metadata", logs[0])
	}
}

func TestSupplierPayableServiceApprovesAndRecordsPayment(t *testing.T) {
	service, auditStore := newTestSupplierPayableService()
	created, err := service.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("create payable: %v", err)
	}

	approved, err := service.ApproveSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		ActorID:   "finance-lead",
		RequestID: "req-ap-approve",
	})
	if err != nil {
		t.Fatalf("approve payment: %v", err)
	}
	if approved.PreviousStatus != financedomain.PayableStatusOpen ||
		approved.CurrentStatus != financedomain.PayableStatusPaymentApproved {
		t.Fatalf("status transition = %q -> %q", approved.PreviousStatus, approved.CurrentStatus)
	}

	paid, err := service.RecordSupplierPayablePayment(context.Background(), SupplierPayableActionInput{
		ID:        created.SupplierPayable.ID,
		Amount:    "1250000.00",
		ActorID:   "cashier",
		RequestID: "req-ap-pay",
	})
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}
	if paid.CurrentStatus != financedomain.PayableStatusPartiallyPaid ||
		paid.SupplierPayable.PaidAmount != "1250000.00" ||
		paid.SupplierPayable.OutstandingAmount != "3000000.00" {
		t.Fatalf("paid result = %+v", paid)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionPayablePaymentRecorded)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["payment_amount"] != "1250000.00" {
		t.Fatalf("audit logs = %+v, want payment amount metadata", logs)
	}
}

func TestSupplierPayableServiceMapsValidationErrors(t *testing.T) {
	service, _ := newTestSupplierPayableService()
	input := baseCreateSupplierPayableInput()
	input.TotalAmount = "4250000.001"

	_, err := service.CreateSupplierPayable(context.Background(), input)
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeSupplierPayableValidation {
		t.Fatalf("code = %q, want %q", appErr.Code, ErrorCodeSupplierPayableValidation)
	}
}

func TestSupplierPayableServiceReturnsNotFound(t *testing.T) {
	service, _ := newTestSupplierPayableService()

	_, err := service.GetSupplierPayable(context.Background(), "missing-ap")
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeSupplierPayableNotFound {
		t.Fatalf("code = %q, want not found", appErr.Code)
	}
}

func newTestSupplierPayableService() (SupplierPayableService, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	store := &PrototypeSupplierPayableStore{records: make(map[string]financedomain.SupplierPayable)}
	service := NewSupplierPayableService(store, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	})

	return service, auditStore
}

func baseCreateSupplierPayableInput() CreateSupplierPayableInput {
	qcSource := SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeQCInspection),
		ID:   "qc-inbound-260430-0001",
		No:   "QC-260430-0001",
	}
	receiptSource := SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeWarehouseReceipt),
		ID:   "gr-260430-0001",
		No:   "GR-260430-0001",
	}

	return CreateSupplierPayableInput{
		ID:             "ap-260430-0001",
		PayableNo:      "AP-260430-0001",
		SupplierID:     "supplier-hcm-001",
		SupplierCode:   "SUP-HCM-001",
		SupplierName:   "Nguyen Lieu HCM",
		SourceDocument: qcSource,
		TotalAmount:    "4250000.00",
		CurrencyCode:   "VND",
		DueDate:        "2026-05-07",
		ActorID:        "finance-user",
		RequestID:      "req-ap-create",
		Lines: []SupplierPayableLineInput{
			{
				ID:             "ap-line-1",
				Description:    "Accepted raw material after inbound QC",
				SourceDocument: receiptSource,
				Amount:         "4250000.00",
			},
		},
	}
}
