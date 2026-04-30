package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCashTransactionCreatesDraftReceiptWithAllocations(t *testing.T) {
	transaction := mustCashReceipt(t)

	if transaction.Status != CashTransactionStatusDraft {
		t.Fatalf("status = %q, want draft", transaction.Status)
	}
	if transaction.Direction != CashTransactionDirectionIn {
		t.Fatalf("direction = %q, want cash_in", transaction.Direction)
	}
	if transaction.TransactionNo != "CASH-IN-260430-0001" ||
		transaction.ReferenceNo != "BANK-REF-0001" {
		t.Fatalf("transaction refs = %q/%q", transaction.TransactionNo, transaction.ReferenceNo)
	}
	if transaction.TotalAmount != "1250000.00" || len(transaction.Allocations) != 2 {
		t.Fatalf("transaction = %+v", transaction)
	}
	if transaction.Allocations[0].TargetType != CashAllocationTargetCustomerReceivable ||
		transaction.Allocations[1].TargetType != CashAllocationTargetCODRemittance {
		t.Fatalf("allocations = %+v", transaction.Allocations)
	}
}

func TestCashTransactionRejectsUntiedOrMismatchedCash(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*NewCashTransactionInput)
		wantErr error
	}{
		{
			name: "missing allocations",
			mutate: func(input *NewCashTransactionInput) {
				input.Allocations = nil
			},
			wantErr: ErrCashTransactionRequiredField,
		},
		{
			name: "allocation total mismatch",
			mutate: func(input *NewCashTransactionInput) {
				input.TotalAmount = "1250001.00"
			},
			wantErr: ErrCashTransactionInvalidAmount,
		},
		{
			name: "negative money",
			mutate: func(input *NewCashTransactionInput) {
				input.Allocations[0].Amount = "-100000.00"
			},
			wantErr: ErrCashTransactionInvalidAmount,
		},
		{
			name: "cash in cannot pay supplier payable",
			mutate: func(input *NewCashTransactionInput) {
				input.Allocations[0].TargetType = CashAllocationTargetSupplierPayable
			},
			wantErr: ErrCashTransactionInvalidAllocation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := baseCashReceiptInput()
			tt.mutate(&input)
			_, err := NewCashTransaction(input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCashTransactionPaymentAllowsPayableTargetsOnly(t *testing.T) {
	input := baseCashPaymentInput()
	transaction, err := NewCashTransaction(input)
	if err != nil {
		t.Fatalf("new payment: %v", err)
	}
	if transaction.Direction != CashTransactionDirectionOut ||
		transaction.Allocations[0].TargetType != CashAllocationTargetSupplierPayable {
		t.Fatalf("payment transaction = %+v", transaction)
	}

	input.Allocations[0].TargetType = CashAllocationTargetCustomerReceivable
	_, err = NewCashTransaction(input)
	if !errors.Is(err, ErrCashTransactionInvalidAllocation) {
		t.Fatalf("error = %v, want %v", err, ErrCashTransactionInvalidAllocation)
	}
}

func TestCashTransactionPostAndVoidRequireTrace(t *testing.T) {
	transaction := mustCashReceipt(t)
	postedAt := time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC)

	posted, err := transaction.Post("cashier", postedAt)
	if err != nil {
		t.Fatalf("post transaction: %v", err)
	}
	if posted.Status != CashTransactionStatusPosted ||
		posted.PostedBy != "cashier" ||
		!posted.PostedAt.Equal(postedAt) {
		t.Fatalf("posted transaction = %+v", posted)
	}

	_, err = transaction.Post("", postedAt)
	if !errors.Is(err, ErrCashTransactionRequiredField) {
		t.Fatalf("error = %v, want %v", err, ErrCashTransactionRequiredField)
	}

	voidedAt := postedAt.Add(time.Hour)
	voided, err := posted.Void("finance-lead", "bank reversal", voidedAt)
	if err != nil {
		t.Fatalf("void transaction: %v", err)
	}
	if voided.Status != CashTransactionStatusVoid ||
		voided.VoidReason != "bank reversal" ||
		!voided.VoidedAt.Equal(voidedAt) {
		t.Fatalf("voided transaction = %+v", voided)
	}

	_, err = posted.Void("finance-lead", "", voidedAt)
	if !errors.Is(err, ErrCashTransactionRequiredField) {
		t.Fatalf("error = %v, want %v", err, ErrCashTransactionRequiredField)
	}
}

func mustCashReceipt(t *testing.T) CashTransaction {
	t.Helper()
	transaction, err := NewCashTransaction(baseCashReceiptInput())
	if err != nil {
		t.Fatalf("new cash receipt: %v", err)
	}

	return transaction
}

func baseCashReceiptInput() NewCashTransactionInput {
	return NewCashTransactionInput{
		ID:               "cash-txn-260430-0001",
		OrgID:            "org-my-pham",
		TransactionNo:    "cash-in-260430-0001",
		Direction:        CashTransactionDirectionIn,
		BusinessDate:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		CounterpartyID:   "carrier-ghn",
		CounterpartyName: "GHN COD",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "bank-ref-0001",
		TotalAmount:      "1250000.00",
		CurrencyCode:     "VND",
		CreatedAt:        time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC),
		CreatedBy:        "finance-user",
		Allocations: []NewCashTransactionAllocationInput{
			{
				ID:         "cash-txn-260430-0001-ar",
				TargetType: CashAllocationTargetCustomerReceivable,
				TargetID:   "ar-260430-0001",
				TargetNo:   "ar-260430-0001",
				Amount:     "1000000.00",
			},
			{
				ID:         "cash-txn-260430-0001-cod",
				TargetType: CashAllocationTargetCODRemittance,
				TargetID:   "cod-remit-260430-0001",
				TargetNo:   "cod-remit-260430-0001",
				Amount:     "250000.00",
			},
		},
	}
}

func baseCashPaymentInput() NewCashTransactionInput {
	return NewCashTransactionInput{
		ID:               "cash-txn-260430-0002",
		OrgID:            "org-my-pham",
		TransactionNo:    "cash-out-260430-0002",
		Direction:        CashTransactionDirectionOut,
		BusinessDate:     time.Date(2026, 4, 30, 0, 0, 0, 0, time.UTC),
		CounterpartyID:   "supplier-hcm-001",
		CounterpartyName: "Nguyen Lieu HCM",
		PaymentMethod:    "bank_transfer",
		ReferenceNo:      "bank-ref-0002",
		TotalAmount:      "4250000.00",
		CurrencyCode:     "VND",
		CreatedAt:        time.Date(2026, 4, 30, 11, 0, 0, 0, time.UTC),
		CreatedBy:        "finance-user",
		Allocations: []NewCashTransactionAllocationInput{
			{
				ID:         "cash-txn-260430-0002-ap",
				TargetType: CashAllocationTargetSupplierPayable,
				TargetID:   "ap-supplier-260430-0001",
				TargetNo:   "ap-supplier-260430-0001",
				Amount:     "4250000.00",
			},
		},
	}
}
