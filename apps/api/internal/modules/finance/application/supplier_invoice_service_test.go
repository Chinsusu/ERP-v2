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
