package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCustomerReceivableCreatesOpenReceivable(t *testing.T) {
	receivable := mustCustomerReceivable(t)

	if receivable.Status != ReceivableStatusOpen {
		t.Fatalf("status = %q, want open", receivable.Status)
	}
	if receivable.ReceivableNo != "AR-260430-0001" {
		t.Fatalf("receivable no = %q, want uppercase no", receivable.ReceivableNo)
	}
	if receivable.TotalAmount != "1250000.00" ||
		receivable.PaidAmount != "0.00" ||
		receivable.OutstandingAmount != "1250000.00" {
		t.Fatalf("amounts = total %q paid %q outstanding %q", receivable.TotalAmount, receivable.PaidAmount, receivable.OutstandingAmount)
	}
	if len(receivable.Lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(receivable.Lines))
	}
	if receivable.Lines[0].SourceDocument.Type != SourceDocumentTypeShipment {
		t.Fatalf("line source type = %q, want shipment", receivable.Lines[0].SourceDocument.Type)
	}
}

func TestNewCustomerReceivableRejectsLineTotalMismatch(t *testing.T) {
	input := baseCustomerReceivableInput()
	input.TotalAmount = "1250001.00"

	_, err := NewCustomerReceivable(input)
	if !errors.Is(err, ErrCustomerReceivableInvalidAmount) {
		t.Fatalf("error = %v, want %v", err, ErrCustomerReceivableInvalidAmount)
	}
}

func TestNewCustomerReceivableRejectsUnsafeAmountsAndSource(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*NewCustomerReceivableInput)
		wantErr error
	}{
		{
			name: "too much money scale",
			mutate: func(input *NewCustomerReceivableInput) {
				input.Lines[0].Amount = "1000000.001"
			},
			wantErr: ErrCustomerReceivableInvalidAmount,
		},
		{
			name: "negative money",
			mutate: func(input *NewCustomerReceivableInput) {
				input.Lines[0].Amount = "-1000000.00"
			},
			wantErr: ErrCustomerReceivableInvalidAmount,
		},
		{
			name: "unknown source",
			mutate: func(input *NewCustomerReceivableInput) {
				input.SourceDocument.Type = "general_ledger"
			},
			wantErr: ErrCustomerReceivableInvalidSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := baseCustomerReceivableInput()
			tt.mutate(&input)
			_, err := NewCustomerReceivable(input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCustomerReceivableRecordReceiptUpdatesPaidAndOutstanding(t *testing.T) {
	receivable := mustCustomerReceivable(t)
	receivedAt := time.Date(2026, 4, 30, 9, 30, 0, 0, time.UTC)

	partial, err := receivable.RecordReceipt("250000.5", "finance-user", receivedAt)
	if err != nil {
		t.Fatalf("record partial receipt: %v", err)
	}
	if partial.Status != ReceivableStatusPartiallyPaid {
		t.Fatalf("status = %q, want partially_paid", partial.Status)
	}
	if partial.PaidAmount != "250000.50" || partial.OutstandingAmount != "999999.50" {
		t.Fatalf("partial amounts = paid %q outstanding %q", partial.PaidAmount, partial.OutstandingAmount)
	}
	if partial.LastReceiptBy != "finance-user" || !partial.LastReceiptAt.Equal(receivedAt) {
		t.Fatalf("receipt audit fields = %q %v", partial.LastReceiptBy, partial.LastReceiptAt)
	}

	paid, err := partial.RecordReceipt("999999.50", "finance-user", receivedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("record final receipt: %v", err)
	}
	if paid.Status != ReceivableStatusPaid {
		t.Fatalf("status = %q, want paid", paid.Status)
	}
	if paid.PaidAmount != "1250000.00" || paid.OutstandingAmount != "0.00" {
		t.Fatalf("paid amounts = paid %q outstanding %q", paid.PaidAmount, paid.OutstandingAmount)
	}
}

func TestCustomerReceivableRecordReceiptRejectsOverpayment(t *testing.T) {
	receivable := mustCustomerReceivable(t)

	_, err := receivable.RecordReceipt("1250000.01", "finance-user", time.Now())
	if !errors.Is(err, ErrCustomerReceivableInvalidAmount) {
		t.Fatalf("error = %v, want %v", err, ErrCustomerReceivableInvalidAmount)
	}
}

func TestCustomerReceivableDisputeAndVoidRequireReason(t *testing.T) {
	receivable := mustCustomerReceivable(t)
	disputedAt := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)

	disputed, err := receivable.MarkDisputed("finance-user", "carrier short remittance", disputedAt)
	if err != nil {
		t.Fatalf("mark disputed: %v", err)
	}
	if disputed.Status != ReceivableStatusDisputed ||
		disputed.DisputeReason != "carrier short remittance" ||
		!disputed.DisputedAt.Equal(disputedAt) {
		t.Fatalf("disputed receivable = %+v", disputed)
	}

	_, err = receivable.MarkDisputed("finance-user", "", disputedAt)
	if !errors.Is(err, ErrCustomerReceivableRequiredField) {
		t.Fatalf("error = %v, want %v", err, ErrCustomerReceivableRequiredField)
	}

	voidedAt := disputedAt.Add(time.Hour)
	voided, err := disputed.Void("finance-lead", "duplicate receivable", voidedAt)
	if err != nil {
		t.Fatalf("void disputed: %v", err)
	}
	if voided.Status != ReceivableStatusVoid ||
		voided.VoidReason != "duplicate receivable" ||
		!voided.VoidedAt.Equal(voidedAt) {
		t.Fatalf("voided receivable = %+v", voided)
	}
}

func TestCustomerReceivablePaidCannotBeVoided(t *testing.T) {
	receivable := mustCustomerReceivable(t)
	paid, err := receivable.RecordReceipt("1250000.00", "finance-user", time.Now())
	if err != nil {
		t.Fatalf("record receipt: %v", err)
	}

	_, err = paid.Void("finance-user", "mistake", time.Now())
	if !errors.Is(err, ErrCustomerReceivableInvalidTransition) {
		t.Fatalf("error = %v, want %v", err, ErrCustomerReceivableInvalidTransition)
	}
}

func mustCustomerReceivable(t *testing.T) CustomerReceivable {
	t.Helper()
	receivable, err := NewCustomerReceivable(baseCustomerReceivableInput())
	if err != nil {
		t.Fatalf("new customer receivable: %v", err)
	}

	return receivable
}

func baseCustomerReceivableInput() NewCustomerReceivableInput {
	headerSource, err := NewSourceDocumentRef(SourceDocumentTypeShipment, "shipment-260430-0001", "SHP-260430-0001")
	if err != nil {
		panic(err)
	}
	lineSource, err := NewSourceDocumentRef(SourceDocumentTypeShipment, "shipment-260430-0001", "SHP-260430-0001")
	if err != nil {
		panic(err)
	}

	return NewCustomerReceivableInput{
		ID:             "ar-260430-0001",
		OrgID:          "org-my-pham",
		ReceivableNo:   "ar-260430-0001",
		CustomerID:     "customer-hcm-001",
		CustomerCode:   "kh001",
		CustomerName:   "My Pham HCM Retail",
		SourceDocument: headerSource,
		TotalAmount:    "1250000.00",
		CurrencyCode:   "VND",
		DueDate:        time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		CreatedBy:      "sales-user",
		Lines: []NewCustomerReceivableLineInput{
			{
				ID:             "ar-line-1",
				Description:    "COD delivered goods",
				SourceDocument: lineSource,
				Amount:         "1000000.00",
			},
			{
				ID:             "ar-line-2",
				Description:    "Shipping fee collected from customer",
				SourceDocument: lineSource,
				Amount:         "250000.00",
			},
		},
	}
}
