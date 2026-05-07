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

func TestSupplierInvoiceServiceCreatesMatchedInvoiceFromPayable(t *testing.T) {
	service, auditStore := newTestSupplierInvoiceService(t)

	result, err := service.CreateSupplierInvoice(context.Background(), CreateSupplierInvoiceInput{
		ID:            "si-260505-0001",
		InvoiceNo:     "inv-supplier-0001",
		PayableID:     "ap-260430-0001",
		InvoiceDate:   "2026-05-05",
		InvoiceAmount: "4250000.00",
		CurrencyCode:  "VND",
		ActorID:       "finance-user",
		RequestID:     "req-si-create",
	})
	if err != nil {
		t.Fatalf("create supplier invoice: %v", err)
	}
	invoice := result.SupplierInvoice
	if invoice.Status != financedomain.SupplierInvoiceStatusMatched ||
		invoice.MatchStatus != financedomain.SupplierInvoiceMatchStatusMatched ||
		invoice.VarianceAmount != "0.00" ||
		invoice.PayableNo != "AP-260430-0001" {
		t.Fatalf("invoice = %+v, want matched AP invoice", invoice)
	}
	if invoice.SupplierID != "supplier-hcm-001" || len(invoice.Lines) != 1 {
		t.Fatalf("invoice supplier/lines = %+v", invoice)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: string(financedomain.FinanceAuditActionSupplierInvoiceCreated)})
	if err != nil {
		t.Fatalf("list audit: %v", err)
	}
	if len(logs) != 1 || logs[0].EntityType != string(financedomain.FinanceEntityTypeSupplierInvoice) ||
		logs[0].Metadata["payable_no"] != "AP-260430-0001" {
		t.Fatalf("audit logs = %+v, want supplier invoice create metadata", logs)
	}
}

func TestSupplierInvoiceServiceCreatesMatchedInvoiceFromFactoryFinalPaymentPayable(t *testing.T) {
	service := newTestSupplierInvoiceServiceWithPayable(t, factoryFinalPaymentPayableInput())

	result, err := service.CreateSupplierInvoice(context.Background(), CreateSupplierInvoiceInput{
		ID:            "si-s35-factory-final",
		InvoiceNo:     "INV-S35-FINAL",
		PayableID:     "ap-s35-factory-final",
		InvoiceDate:   "2026-05-07",
		InvoiceAmount: "29000.00",
		CurrencyCode:  "VND",
		ActorID:       "finance-user",
		RequestID:     "req-si-s35-final",
	})
	if err != nil {
		t.Fatalf("create factory final payment supplier invoice: %v", err)
	}
	invoice := result.SupplierInvoice
	if invoice.Status != financedomain.SupplierInvoiceStatusMatched ||
		invoice.SourceDocument.Type != financedomain.SourceDocumentTypeSubcontractPaymentMilestone ||
		len(invoice.Lines) != 1 ||
		invoice.Lines[0].SourceDocument.Type != financedomain.SourceDocumentTypeSubcontractOrder {
		t.Fatalf("invoice = %+v, want matched factory final payment invoice with milestone/order sources", invoice)
	}
}

func TestSupplierInvoiceServiceCreatesMismatchWithVarianceLine(t *testing.T) {
	service, _ := newTestSupplierInvoiceService(t)

	result, err := service.CreateSupplierInvoice(context.Background(), CreateSupplierInvoiceInput{
		ID:            "si-260505-0002",
		InvoiceNo:     "INV-SUPPLIER-0002",
		PayableID:     "ap-260430-0001",
		InvoiceDate:   "2026-05-05",
		InvoiceAmount: "4200000.00",
		CurrencyCode:  "VND",
		ActorID:       "finance-user",
		RequestID:     "req-si-mismatch",
	})
	if err != nil {
		t.Fatalf("create mismatch supplier invoice: %v", err)
	}
	invoice := result.SupplierInvoice
	if invoice.Status != financedomain.SupplierInvoiceStatusMismatch ||
		invoice.MatchStatus != financedomain.SupplierInvoiceMatchStatusMismatch ||
		invoice.VarianceAmount != "-50000.00" {
		t.Fatalf("invoice = %+v, want mismatch variance", invoice)
	}
	if len(invoice.Lines) != 2 ||
		invoice.Lines[1].SourceDocument.Type != financedomain.SourceDocumentTypeSupplierPayable ||
		invoice.Lines[1].Amount != "-50000.00" {
		t.Fatalf("invoice lines = %+v, want AP variance line", invoice.Lines)
	}
}

func TestSupplierInvoiceServiceSearchMatchesPayableAndSourceDocuments(t *testing.T) {
	service, _ := newTestSupplierInvoiceService(t)
	if _, err := service.CreateSupplierInvoice(context.Background(), CreateSupplierInvoiceInput{
		ID:            "si-260505-0003",
		InvoiceNo:     "INV-SUPPLIER-0003",
		PayableID:     "ap-260430-0001",
		InvoiceDate:   "2026-05-05",
		InvoiceAmount: "4250000.00",
		CurrencyCode:  "VND",
		ActorID:       "finance-user",
		RequestID:     "req-si-search",
	}); err != nil {
		t.Fatalf("create supplier invoice: %v", err)
	}

	invoices, err := service.ListSupplierInvoices(context.Background(), SupplierInvoiceFilter{Search: "GR-260430-0001"})
	if err != nil {
		t.Fatalf("list supplier invoices: %v", err)
	}
	if len(invoices) != 1 || invoices[0].ID != "si-260505-0003" {
		t.Fatalf("invoices by source = %+v, want one invoice", invoices)
	}

	invoices, err = service.ListSupplierInvoices(context.Background(), SupplierInvoiceFilter{Search: "AP-260430-0001"})
	if err != nil {
		t.Fatalf("list supplier invoices by AP: %v", err)
	}
	if len(invoices) != 1 || invoices[0].ID != "si-260505-0003" {
		t.Fatalf("invoices by AP = %+v, want one invoice", invoices)
	}
}

func TestSupplierInvoiceServiceMapsMissingPayable(t *testing.T) {
	service, _ := newTestSupplierInvoiceService(t)

	_, err := service.CreateSupplierInvoice(context.Background(), CreateSupplierInvoiceInput{
		InvoiceNo:     "INV-MISSING",
		PayableID:     "missing-ap",
		InvoiceDate:   "2026-05-05",
		InvoiceAmount: "4250000.00",
		CurrencyCode:  "VND",
		ActorID:       "finance-user",
	})
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error = %T, want app error", err)
	}
	if appErr.Code != ErrorCodeSupplierPayableNotFound {
		t.Fatalf("code = %q, want supplier payable not found", appErr.Code)
	}
}

func TestSupplierInvoiceServiceVoidsInvoice(t *testing.T) {
	service, _ := newTestSupplierInvoiceService(t)
	created, err := service.CreateSupplierInvoice(context.Background(), CreateSupplierInvoiceInput{
		ID:            "si-260505-0004",
		InvoiceNo:     "INV-SUPPLIER-0004",
		PayableID:     "ap-260430-0001",
		InvoiceDate:   "2026-05-05",
		InvoiceAmount: "4250000.00",
		CurrencyCode:  "VND",
		ActorID:       "finance-user",
		RequestID:     "req-si-create-void",
	})
	if err != nil {
		t.Fatalf("create supplier invoice: %v", err)
	}

	voided, err := service.VoidSupplierInvoice(context.Background(), SupplierInvoiceActionInput{
		ID:        created.SupplierInvoice.ID,
		Reason:    "duplicate invoice",
		ActorID:   "finance-lead",
		RequestID: "req-si-void",
	})
	if err != nil {
		t.Fatalf("void supplier invoice: %v", err)
	}
	if voided.PreviousStatus != financedomain.SupplierInvoiceStatusMatched ||
		voided.CurrentStatus != financedomain.SupplierInvoiceStatusVoid ||
		voided.SupplierInvoice.VoidReason != "duplicate invoice" {
		t.Fatalf("voided = %+v", voided)
	}
}

func newTestSupplierInvoiceServiceWithPayable(
	t *testing.T,
	input CreateSupplierPayableInput,
) SupplierInvoiceService {
	t.Helper()
	auditStore := audit.NewInMemoryLogStore()
	payableStore := &PrototypeSupplierPayableStore{records: make(map[string]financedomain.SupplierPayable)}
	payableService := NewSupplierPayableService(payableStore, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 5, 7, 9, 0, 0, 0, time.UTC)
	})
	created, err := payableService.CreateSupplierPayable(context.Background(), input)
	if err != nil {
		t.Fatalf("seed payable: %v", err)
	}
	payableStore.records[created.SupplierPayable.ID] = created.SupplierPayable.Clone()
	invoiceStore := &PrototypeSupplierInvoiceStore{records: make(map[string]financedomain.SupplierInvoice)}

	return NewSupplierInvoiceService(invoiceStore, payableStore, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 5, 7, 10, 0, 0, 0, time.UTC)
	})
}

func factoryFinalPaymentPayableInput() CreateSupplierPayableInput {
	input := baseCreateSupplierPayableInput()
	input.ID = "ap-s35-factory-final"
	input.PayableNo = "AP-S35-FACTORY-FINAL"
	input.SupplierID = "factory-bd-002"
	input.SupplierCode = "FACT-BD-002"
	input.SupplierName = "Binh Duong Gia Cong"
	input.SourceDocument = SourceDocumentInput{
		Type: string(financedomain.SourceDocumentTypeSubcontractPaymentMilestone),
		ID:   "spm-s35-factory-final",
		No:   "SPM-S35-FACTORY-FINAL",
	}
	input.TotalAmount = "29000.00"
	input.DueDate = "2026-05-14"
	input.ActorID = "finance-user"
	input.RequestID = "req-ap-s35-factory-final"
	input.Lines = []SupplierPayableLineInput{
		{
			ID:          "ap-line-s35-factory-final",
			Description: "Final subcontract payment for SCO-S35-FACTORY-FINAL",
			SourceDocument: SourceDocumentInput{
				Type: string(financedomain.SourceDocumentTypeSubcontractOrder),
				ID:   "sco-s35-factory-final",
				No:   "SCO-S35-FACTORY-FINAL",
			},
			Amount: "29000.00",
		},
	}

	return input
}

func newTestSupplierInvoiceService(t *testing.T) (SupplierInvoiceService, *audit.InMemoryLogStore) {
	t.Helper()
	auditStore := audit.NewInMemoryLogStore()
	payableStore := &PrototypeSupplierPayableStore{records: make(map[string]financedomain.SupplierPayable)}
	payableService := NewSupplierPayableService(payableStore, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	})
	created, err := payableService.CreateSupplierPayable(context.Background(), baseCreateSupplierPayableInput())
	if err != nil {
		t.Fatalf("seed payable: %v", err)
	}
	payableStore.records[created.SupplierPayable.ID] = created.SupplierPayable.Clone()
	invoiceStore := &PrototypeSupplierInvoiceStore{records: make(map[string]financedomain.SupplierInvoice)}
	service := NewSupplierInvoiceService(invoiceStore, payableStore, auditStore).WithClock(func() time.Time {
		return time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC)
	})

	return service, auditStore
}
