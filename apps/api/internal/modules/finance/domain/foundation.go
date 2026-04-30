package domain

import (
	"errors"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrFinanceInvalidStatus = errors.New("finance status is invalid")
var ErrFinanceInvalidCurrency = errors.New("finance currency is invalid")
var ErrFinanceInvalidMoneyAmount = errors.New("finance money amount is invalid")

type ReceivableStatus string

const (
	ReceivableStatusDraft         ReceivableStatus = "draft"
	ReceivableStatusOpen          ReceivableStatus = "open"
	ReceivableStatusPartiallyPaid ReceivableStatus = "partially_paid"
	ReceivableStatusPaid          ReceivableStatus = "paid"
	ReceivableStatusDisputed      ReceivableStatus = "disputed"
	ReceivableStatusVoid          ReceivableStatus = "void"
)

type PayableStatus string

const (
	PayableStatusDraft            PayableStatus = "draft"
	PayableStatusOpen             PayableStatus = "open"
	PayableStatusPaymentRequested PayableStatus = "payment_requested"
	PayableStatusPaymentApproved  PayableStatus = "payment_approved"
	PayableStatusPartiallyPaid    PayableStatus = "partially_paid"
	PayableStatusPaid             PayableStatus = "paid"
	PayableStatusDisputed         PayableStatus = "disputed"
	PayableStatusVoid             PayableStatus = "void"
)

type CODRemittanceStatus string

const (
	CODRemittanceStatusDraft       CODRemittanceStatus = "draft"
	CODRemittanceStatusMatching    CODRemittanceStatus = "matching"
	CODRemittanceStatusSubmitted   CODRemittanceStatus = "submitted"
	CODRemittanceStatusApproved    CODRemittanceStatus = "approved"
	CODRemittanceStatusDiscrepancy CODRemittanceStatus = "discrepancy"
	CODRemittanceStatusClosed      CODRemittanceStatus = "closed"
	CODRemittanceStatusVoid        CODRemittanceStatus = "void"
)

type CashTransactionStatus string

const (
	CashTransactionStatusDraft  CashTransactionStatus = "draft"
	CashTransactionStatusPosted CashTransactionStatus = "posted"
	CashTransactionStatusVoid   CashTransactionStatus = "void"
)

type CashTransactionDirection string

const (
	CashTransactionDirectionIn  CashTransactionDirection = "cash_in"
	CashTransactionDirectionOut CashTransactionDirection = "cash_out"
)

type MoneyAmount struct {
	Amount       decimal.Decimal
	CurrencyCode decimal.CurrencyCode
}

func NewMoneyAmount(value string, currencyCode string) (MoneyAmount, error) {
	return newMoneyAmount(value, currencyCode, false)
}

func NewSignedMoneyAmount(value string, currencyCode string) (MoneyAmount, error) {
	return newMoneyAmount(value, currencyCode, true)
}

func NormalizeReceivableStatus(status ReceivableStatus) ReceivableStatus {
	return ReceivableStatus(normalizeStatus(string(status)))
}

func IsValidReceivableStatus(status ReceivableStatus) bool {
	switch NormalizeReceivableStatus(status) {
	case ReceivableStatusDraft,
		ReceivableStatusOpen,
		ReceivableStatusPartiallyPaid,
		ReceivableStatusPaid,
		ReceivableStatusDisputed,
		ReceivableStatusVoid:
		return true
	default:
		return false
	}
}

func NormalizePayableStatus(status PayableStatus) PayableStatus {
	return PayableStatus(normalizeStatus(string(status)))
}

func IsValidPayableStatus(status PayableStatus) bool {
	switch NormalizePayableStatus(status) {
	case PayableStatusDraft,
		PayableStatusOpen,
		PayableStatusPaymentRequested,
		PayableStatusPaymentApproved,
		PayableStatusPartiallyPaid,
		PayableStatusPaid,
		PayableStatusDisputed,
		PayableStatusVoid:
		return true
	default:
		return false
	}
}

func NormalizeCODRemittanceStatus(status CODRemittanceStatus) CODRemittanceStatus {
	return CODRemittanceStatus(normalizeStatus(string(status)))
}

func IsValidCODRemittanceStatus(status CODRemittanceStatus) bool {
	switch NormalizeCODRemittanceStatus(status) {
	case CODRemittanceStatusDraft,
		CODRemittanceStatusMatching,
		CODRemittanceStatusSubmitted,
		CODRemittanceStatusApproved,
		CODRemittanceStatusDiscrepancy,
		CODRemittanceStatusClosed,
		CODRemittanceStatusVoid:
		return true
	default:
		return false
	}
}

func NormalizeCashTransactionStatus(status CashTransactionStatus) CashTransactionStatus {
	return CashTransactionStatus(normalizeStatus(string(status)))
}

func IsValidCashTransactionStatus(status CashTransactionStatus) bool {
	switch NormalizeCashTransactionStatus(status) {
	case CashTransactionStatusDraft, CashTransactionStatusPosted, CashTransactionStatusVoid:
		return true
	default:
		return false
	}
}

func NormalizeCashTransactionDirection(direction CashTransactionDirection) CashTransactionDirection {
	return CashTransactionDirection(normalizeStatus(string(direction)))
}

func IsValidCashTransactionDirection(direction CashTransactionDirection) bool {
	switch NormalizeCashTransactionDirection(direction) {
	case CashTransactionDirectionIn, CashTransactionDirectionOut:
		return true
	default:
		return false
	}
}

func newMoneyAmount(value string, currencyCode string, allowNegative bool) (MoneyAmount, error) {
	currency, err := decimal.NormalizeCurrencyCode(currencyCode)
	if err != nil || currency != decimal.CurrencyVND {
		return MoneyAmount{}, ErrFinanceInvalidCurrency
	}

	amount, err := decimal.ParseMoneyAmount(value)
	if err != nil || (!allowNegative && amount.IsNegative()) {
		return MoneyAmount{}, ErrFinanceInvalidMoneyAmount
	}

	return MoneyAmount{
		Amount:       amount,
		CurrencyCode: currency,
	}, nil
}

func normalizeStatus(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
