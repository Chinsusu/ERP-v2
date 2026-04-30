package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewSupplierPayableCreatesOpenPayable(t *testing.T) {
	payable := mustSupplierPayable(t)

	if payable.Status != PayableStatusOpen {
		t.Fatalf("status = %q, want open", payable.Status)
	}
	if payable.PayableNo != "AP-260430-0001" {
		t.Fatalf("payable no = %q, want uppercase no", payable.PayableNo)
	}
	if payable.SupplierCode != "SUP-HCM-001" {
		t.Fatalf("supplier code = %q, want uppercase code", payable.SupplierCode)
	}
	if payable.TotalAmount != "4250000.00" ||
		payable.PaidAmount != "0.00" ||
		payable.OutstandingAmount != "4250000.00" {
		t.Fatalf("amounts = total %q paid %q outstanding %q", payable.TotalAmount, payable.PaidAmount, payable.OutstandingAmount)
	}
	if len(payable.Lines) != 2 {
		t.Fatalf("lines = %d, want 2", len(payable.Lines))
	}
	if payable.SourceDocument.Type != SourceDocumentTypeQCInspection {
		t.Fatalf("source type = %q, want qc_inspection", payable.SourceDocument.Type)
	}
}

func TestNewSupplierPayableRejectsLineTotalMismatch(t *testing.T) {
	input := baseSupplierPayableInput()
	input.TotalAmount = "4250001.00"

	_, err := NewSupplierPayable(input)
	if !errors.Is(err, ErrSupplierPayableInvalidAmount) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableInvalidAmount)
	}
}

func TestNewSupplierPayableRejectsUnsafeAmountsAndSource(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*NewSupplierPayableInput)
		wantErr error
	}{
		{
			name: "too much money scale",
			mutate: func(input *NewSupplierPayableInput) {
				input.Lines[0].Amount = "1000000.001"
			},
			wantErr: ErrSupplierPayableInvalidAmount,
		},
		{
			name: "negative money",
			mutate: func(input *NewSupplierPayableInput) {
				input.Lines[0].Amount = "-1000000.00"
			},
			wantErr: ErrSupplierPayableInvalidAmount,
		},
		{
			name: "shipment source",
			mutate: func(input *NewSupplierPayableInput) {
				input.SourceDocument = SourceDocumentRef{Type: SourceDocumentTypeShipment, ID: "shipment-1"}
			},
			wantErr: ErrSupplierPayableInvalidSource,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := baseSupplierPayableInput()
			tt.mutate(&input)
			_, err := NewSupplierPayable(input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSupplierPayablePaymentFlowUpdatesPaidAndOutstanding(t *testing.T) {
	payable := mustSupplierPayable(t)
	requestedAt := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	approvedAt := requestedAt.Add(time.Hour)
	paidAt := approvedAt.Add(time.Hour)

	requested, err := payable.RequestPayment("finance-user", requestedAt)
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}
	if requested.Status != PayableStatusPaymentRequested ||
		requested.PaymentRequestedBy != "finance-user" ||
		!requested.PaymentRequestedAt.Equal(requestedAt) {
		t.Fatalf("requested payable = %+v", requested)
	}

	approved, err := requested.ApprovePayment("finance-lead", approvedAt)
	if err != nil {
		t.Fatalf("approve payment: %v", err)
	}
	if approved.Status != PayableStatusPaymentApproved ||
		approved.PaymentApprovedBy != "finance-lead" ||
		!approved.PaymentApprovedAt.Equal(approvedAt) {
		t.Fatalf("approved payable = %+v", approved)
	}

	partial, err := approved.RecordPayment("1250000.50", "cashier", paidAt)
	if err != nil {
		t.Fatalf("record partial payment: %v", err)
	}
	if partial.Status != PayableStatusPartiallyPaid ||
		partial.PaidAmount != "1250000.50" ||
		partial.OutstandingAmount != "2999999.50" {
		t.Fatalf("partial payable = %+v", partial)
	}

	paid, err := partial.RecordPayment("2999999.50", "cashier", paidAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("record final payment: %v", err)
	}
	if paid.Status != PayableStatusPaid ||
		paid.PaidAmount != "4250000.00" ||
		paid.OutstandingAmount != "0.00" {
		t.Fatalf("paid payable = %+v", paid)
	}
}

func TestSupplierPayableRecordPaymentRejectsUnapprovedAndOverpayment(t *testing.T) {
	payable := mustSupplierPayable(t)

	_, err := payable.RecordPayment("100000.00", "cashier", time.Now())
	if !errors.Is(err, ErrSupplierPayableInvalidTransition) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableInvalidTransition)
	}

	approved := mustApprovedSupplierPayable(t)
	_, err = approved.RecordPayment("4250000.01", "cashier", time.Now())
	if !errors.Is(err, ErrSupplierPayableInvalidAmount) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableInvalidAmount)
	}
}

func TestSupplierPayableRejectPaymentReturnsToOpenWithReason(t *testing.T) {
	requestedAt := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	rejectedAt := requestedAt.Add(time.Hour)
	requested, err := mustSupplierPayable(t).RequestPayment("finance-user", requestedAt)
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}

	rejected, err := requested.RejectPayment("finance-lead", "supplier invoice mismatch", rejectedAt)
	if err != nil {
		t.Fatalf("reject payment: %v", err)
	}
	if rejected.Status != PayableStatusOpen ||
		rejected.PaymentRejectedBy != "finance-lead" ||
		rejected.PaymentRejectReason != "supplier invoice mismatch" ||
		!rejected.PaymentRejectedAt.Equal(rejectedAt) {
		t.Fatalf("rejected payable = %+v", rejected)
	}
	if rejected.PaymentRequestedBy != "finance-user" || !rejected.PaymentRequestedAt.Equal(requestedAt) {
		t.Fatalf("requested audit fields changed after rejection: %+v", rejected)
	}

	rerequested, err := rejected.RequestPayment("finance-user", rejectedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("request payment again: %v", err)
	}
	if rerequested.PaymentRejectedBy != "" ||
		rerequested.PaymentRejectReason != "" ||
		!rerequested.PaymentRejectedAt.IsZero() {
		t.Fatalf("re-requested payable kept stale rejection fields: %+v", rerequested)
	}

	_, err = requested.RejectPayment("finance-lead", "", rejectedAt)
	if !errors.Is(err, ErrSupplierPayableRequiredField) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableRequiredField)
	}

	_, err = mustSupplierPayable(t).RejectPayment("finance-lead", "not requested", rejectedAt)
	if !errors.Is(err, ErrSupplierPayableInvalidTransition) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableInvalidTransition)
	}
}

func TestSupplierPayableDisputeAndVoidRequireReason(t *testing.T) {
	payable := mustSupplierPayable(t)
	disputedAt := time.Date(2026, 4, 30, 13, 0, 0, 0, time.UTC)

	disputed, err := payable.MarkDisputed("finance-user", "invoice does not match QC pass quantity", disputedAt)
	if err != nil {
		t.Fatalf("mark disputed: %v", err)
	}
	if disputed.Status != PayableStatusDisputed ||
		disputed.DisputeReason != "invoice does not match QC pass quantity" ||
		!disputed.DisputedAt.Equal(disputedAt) {
		t.Fatalf("disputed payable = %+v", disputed)
	}

	_, err = payable.MarkDisputed("finance-user", "", disputedAt)
	if !errors.Is(err, ErrSupplierPayableRequiredField) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableRequiredField)
	}

	voidedAt := disputedAt.Add(time.Hour)
	voided, err := disputed.Void("finance-lead", "supplier reissued invoice", voidedAt)
	if err != nil {
		t.Fatalf("void disputed: %v", err)
	}
	if voided.Status != PayableStatusVoid ||
		voided.VoidReason != "supplier reissued invoice" ||
		voided.OutstandingAmount != "0.00" ||
		!voided.VoidedAt.Equal(voidedAt) {
		t.Fatalf("voided payable = %+v", voided)
	}
}

func TestSupplierPayablePaidCannotBeVoided(t *testing.T) {
	approved := mustApprovedSupplierPayable(t)
	paid, err := approved.RecordPayment("4250000.00", "cashier", time.Now())
	if err != nil {
		t.Fatalf("record payment: %v", err)
	}

	_, err = paid.Void("finance-user", "mistake", time.Now())
	if !errors.Is(err, ErrSupplierPayableInvalidTransition) {
		t.Fatalf("error = %v, want %v", err, ErrSupplierPayableInvalidTransition)
	}
}

func mustSupplierPayable(t *testing.T) SupplierPayable {
	t.Helper()
	payable, err := NewSupplierPayable(baseSupplierPayableInput())
	if err != nil {
		t.Fatalf("new supplier payable: %v", err)
	}

	return payable
}

func mustApprovedSupplierPayable(t *testing.T) SupplierPayable {
	t.Helper()
	requested, err := mustSupplierPayable(t).RequestPayment("finance-user", time.Now())
	if err != nil {
		t.Fatalf("request payment: %v", err)
	}
	approved, err := requested.ApprovePayment("finance-lead", time.Now())
	if err != nil {
		t.Fatalf("approve payment: %v", err)
	}

	return approved
}

func baseSupplierPayableInput() NewSupplierPayableInput {
	headerSource, err := NewSourceDocumentRef(SourceDocumentTypeQCInspection, "qc-inbound-260430-0001", "QC-260430-0001")
	if err != nil {
		panic(err)
	}
	receiptSource, err := NewSourceDocumentRef(SourceDocumentTypeWarehouseReceipt, "gr-260430-0001", "GR-260430-0001")
	if err != nil {
		panic(err)
	}
	poSource, err := NewSourceDocumentRef(SourceDocumentTypePurchaseOrder, "po-260430-0001", "PO-260430-0001")
	if err != nil {
		panic(err)
	}

	return NewSupplierPayableInput{
		ID:             "ap-260430-0001",
		OrgID:          "org-my-pham",
		PayableNo:      "ap-260430-0001",
		SupplierID:     "supplier-hcm-001",
		SupplierCode:   "sup-hcm-001",
		SupplierName:   "Nguyen Lieu HCM",
		SourceDocument: headerSource,
		TotalAmount:    "4250000.00",
		CurrencyCode:   "VND",
		DueDate:        time.Date(2026, 5, 7, 0, 0, 0, 0, time.UTC),
		CreatedAt:      time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		CreatedBy:      "qc-user",
		Lines: []NewSupplierPayableLineInput{
			{
				ID:             "ap-line-1",
				Description:    "Accepted raw material after inbound QC",
				SourceDocument: receiptSource,
				Amount:         "3000000.00",
			},
			{
				ID:             "ap-line-2",
				Description:    "Accepted packaging after inbound QC",
				SourceDocument: poSource,
				Amount:         "1250000.00",
			},
		},
	}
}
