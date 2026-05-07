package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewSupplierInvoiceMatched(t *testing.T) {
	invoice, err := NewSupplierInvoice(baseSupplierInvoiceInput("4250000.00"))
	if err != nil {
		t.Fatalf("new supplier invoice: %v", err)
	}
	if invoice.Status != SupplierInvoiceStatusMatched ||
		invoice.MatchStatus != SupplierInvoiceMatchStatusMatched ||
		invoice.VarianceAmount != "0.00" {
		t.Fatalf("invoice match state = %q/%q/%q", invoice.Status, invoice.MatchStatus, invoice.VarianceAmount)
	}
	if len(invoice.Lines) != 1 || invoice.Lines[0].SourceDocument.Type != SourceDocumentTypePurchaseOrder {
		t.Fatalf("invoice lines = %+v, want PO source line", invoice.Lines)
	}
}

func TestNewSupplierInvoiceAcceptsFactoryFinalPaymentSources(t *testing.T) {
	input := baseSupplierInvoiceInput("4250000.00")
	input.SourceDocument = SourceDocumentRef{
		Type: SourceDocumentTypeSubcontractPaymentMilestone,
		ID:   "spm-s35-final",
		No:   "SPM-S35-FINAL",
	}
	input.Lines[0].SourceDocument = SourceDocumentRef{
		Type: SourceDocumentTypeSubcontractOrder,
		ID:   "sco-s35-final",
		No:   "SCO-S35-FINAL",
	}

	invoice, err := NewSupplierInvoice(input)
	if err != nil {
		t.Fatalf("new factory final payment supplier invoice: %v", err)
	}
	if invoice.SourceDocument.Type != SourceDocumentTypeSubcontractPaymentMilestone ||
		invoice.Lines[0].SourceDocument.Type != SourceDocumentTypeSubcontractOrder {
		t.Fatalf("source documents = %+v/%+v, want factory milestone/order sources", invoice.SourceDocument, invoice.Lines[0].SourceDocument)
	}
}

func TestNewSupplierInvoiceMismatchAllowsNegativeVarianceLine(t *testing.T) {
	input := baseSupplierInvoiceInput("4200000.00")
	input.Status = SupplierInvoiceStatusMismatch
	input.MatchStatus = SupplierInvoiceMatchStatusMismatch
	input.VarianceAmount = "-50000.00"
	input.Lines = append(input.Lines, NewSupplierInvoiceLineInput{
		ID:          "inv-line-variance",
		Description: "Supplier invoice variance",
		SourceDocument: SourceDocumentRef{
			Type: SourceDocumentTypeSupplierPayable,
			ID:   input.PayableID,
			No:   input.PayableNo,
		},
		Amount: "-50000.00",
	})

	invoice, err := NewSupplierInvoice(input)
	if err != nil {
		t.Fatalf("new supplier invoice mismatch: %v", err)
	}
	if invoice.Status != SupplierInvoiceStatusMismatch ||
		invoice.MatchStatus != SupplierInvoiceMatchStatusMismatch ||
		invoice.VarianceAmount != "-50000.00" {
		t.Fatalf("invoice mismatch state = %q/%q/%q", invoice.Status, invoice.MatchStatus, invoice.VarianceAmount)
	}
}

func TestSupplierInvoiceRejectsInconsistentVariance(t *testing.T) {
	input := baseSupplierInvoiceInput("4200000.00")
	input.Status = SupplierInvoiceStatusMismatch
	input.MatchStatus = SupplierInvoiceMatchStatusMismatch
	input.VarianceAmount = "0.00"

	_, err := NewSupplierInvoice(input)
	if !errors.Is(err, ErrSupplierInvoiceInvalidAmount) {
		t.Fatalf("error = %v, want invalid amount", err)
	}
}

func TestSupplierInvoiceVoidRequiresReason(t *testing.T) {
	invoice, err := NewSupplierInvoice(baseSupplierInvoiceInput("4250000.00"))
	if err != nil {
		t.Fatalf("new supplier invoice: %v", err)
	}

	_, err = invoice.Void("finance-user", "", time.Date(2026, 5, 5, 11, 0, 0, 0, time.UTC))
	if !errors.Is(err, ErrSupplierInvoiceRequiredField) {
		t.Fatalf("error = %v, want required field", err)
	}

	voided, err := invoice.Void("finance-user", "duplicate invoice", time.Date(2026, 5, 5, 11, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("void invoice: %v", err)
	}
	if voided.Status != SupplierInvoiceStatusVoid || voided.VoidReason != "duplicate invoice" || voided.Version != 2 {
		t.Fatalf("voided invoice = %+v", voided)
	}
}

func baseSupplierInvoiceInput(amount string) NewSupplierInvoiceInput {
	return NewSupplierInvoiceInput{
		ID:             "si-260505-0001",
		OrgID:          "org-my-pham",
		InvoiceNo:      "INV-NCC-0001",
		SupplierID:     "supplier-hcm-001",
		SupplierCode:   "SUP-HCM-001",
		SupplierName:   "Nguyen Lieu HCM",
		PayableID:      "ap-260505-0001",
		PayableNo:      "AP-260505-0001",
		Status:         SupplierInvoiceStatusMatched,
		MatchStatus:    SupplierInvoiceMatchStatusMatched,
		InvoiceAmount:  amount,
		ExpectedAmount: "4250000.00",
		VarianceAmount: "0.00",
		CurrencyCode:   "VND",
		InvoiceDate:    time.Date(2026, 5, 5, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		CreatedBy:      "finance-user",
		UpdatedAt:      time.Date(2026, 5, 5, 10, 0, 0, 0, time.UTC),
		UpdatedBy:      "finance-user",
		SourceDocument: SourceDocumentRef{
			Type: SourceDocumentTypeWarehouseReceipt,
			ID:   "gr-260505-0001",
			No:   "GR-260505-0001",
		},
		Lines: []NewSupplierInvoiceLineInput{
			{
				ID:          "si-260505-0001-line-1",
				Description: "Accepted raw material after inbound QC",
				SourceDocument: SourceDocumentRef{
					Type: SourceDocumentTypePurchaseOrder,
					ID:   "po-260505-0001",
					No:   "PO-260505-0001",
				},
				Amount: "4250000.00",
			},
		},
	}
}
