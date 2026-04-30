package domain

import (
	"errors"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestFinanceStatusesNormalizeAndValidate(t *testing.T) {
	if got := NormalizeReceivableStatus(" OPEN "); got != ReceivableStatusOpen {
		t.Fatalf("receivable status = %q, want %q", got, ReceivableStatusOpen)
	}
	if got := NormalizePayableStatus(" Payment_Approved "); got != PayableStatusPaymentApproved {
		t.Fatalf("payable status = %q, want %q", got, PayableStatusPaymentApproved)
	}
	if got := NormalizeCODRemittanceStatus(" DISCREPANCY "); got != CODRemittanceStatusDiscrepancy {
		t.Fatalf("cod status = %q, want %q", got, CODRemittanceStatusDiscrepancy)
	}
	if got := NormalizeCashTransactionStatus(" Posted "); got != CashTransactionStatusPosted {
		t.Fatalf("cash status = %q, want %q", got, CashTransactionStatusPosted)
	}
	if got := NormalizeCashTransactionDirection(" CASH_IN "); got != CashTransactionDirectionIn {
		t.Fatalf("cash direction = %q, want %q", got, CashTransactionDirectionIn)
	}

	if !IsValidReceivableStatus(ReceivableStatusPartiallyPaid) {
		t.Fatal("receivable partially_paid should be valid")
	}
	if !IsValidPayableStatus(PayableStatusPaymentRequested) {
		t.Fatal("payable payment_requested should be valid")
	}
	if !IsValidCODRemittanceStatus(CODRemittanceStatusClosed) {
		t.Fatal("cod closed should be valid")
	}
	if !IsValidCashTransactionDirection(CashTransactionDirectionOut) {
		t.Fatal("cash_out should be valid")
	}
}

func TestFinanceStatusesRejectUnknownValues(t *testing.T) {
	if IsValidReceivableStatus("settled") {
		t.Fatal("unknown receivable status should be invalid")
	}
	if IsValidPayableStatus("approved") {
		t.Fatal("ambiguous payable approved status should be invalid")
	}
	if IsValidCODRemittanceStatus("paid") {
		t.Fatal("unknown cod status should be invalid")
	}
	if IsValidCashTransactionStatus("closed") {
		t.Fatal("unknown cash status should be invalid")
	}
	if IsValidCashTransactionDirection("refund") {
		t.Fatal("unknown cash direction should be invalid")
	}
}

func TestNewMoneyAmountUsesVNDDecimalStringRules(t *testing.T) {
	money, err := NewMoneyAmount("1250000", "")
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	if money.Amount != "1250000.00" {
		t.Fatalf("amount = %q, want 1250000.00", money.Amount)
	}
	if money.CurrencyCode != decimal.CurrencyVND {
		t.Fatalf("currency = %q, want VND", money.CurrencyCode)
	}
}

func TestNewMoneyAmountRejectsUnsafeValues(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		currency string
		wantErr  error
	}{
		{name: "too much scale", value: "1.234", currency: "VND", wantErr: ErrFinanceInvalidMoneyAmount},
		{name: "negative", value: "-1.00", currency: "VND", wantErr: ErrFinanceInvalidMoneyAmount},
		{name: "non VND", value: "1.00", currency: "USD", wantErr: ErrFinanceInvalidCurrency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMoneyAmount(tt.value, tt.currency)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSignedMoneyAmountAllowsAdjustments(t *testing.T) {
	money, err := NewSignedMoneyAmount("-25000.5", "VND")
	if err != nil {
		t.Fatalf("signed money: %v", err)
	}
	if money.Amount != "-25000.50" {
		t.Fatalf("amount = %q, want -25000.50", money.Amount)
	}
}
