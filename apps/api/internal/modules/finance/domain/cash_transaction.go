package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrCashTransactionRequiredField = errors.New("cash transaction required field is missing")
var ErrCashTransactionInvalidStatus = errors.New("cash transaction status is invalid")
var ErrCashTransactionInvalidDirection = errors.New("cash transaction direction is invalid")
var ErrCashTransactionInvalidAmount = errors.New("cash transaction amount is invalid")
var ErrCashTransactionInvalidAllocation = errors.New("cash transaction allocation is invalid")
var ErrCashTransactionInvalidTransition = errors.New("cash transaction status transition is invalid")

type CashAllocationTargetType string

const (
	CashAllocationTargetCustomerReceivable CashAllocationTargetType = "customer_receivable"
	CashAllocationTargetSupplierPayable    CashAllocationTargetType = "supplier_payable"
	CashAllocationTargetCODRemittance      CashAllocationTargetType = "cod_remittance"
	CashAllocationTargetPaymentRequest     CashAllocationTargetType = "payment_request"
	CashAllocationTargetManualAdjustment   CashAllocationTargetType = "manual_adjustment"
)

type CashTransaction struct {
	ID               string
	OrgID            string
	TransactionNo    string
	Direction        CashTransactionDirection
	Status           CashTransactionStatus
	BusinessDate     time.Time
	CounterpartyID   string
	CounterpartyName string
	PaymentMethod    string
	ReferenceNo      string
	Allocations      []CashTransactionAllocation
	TotalAmount      decimal.Decimal
	CurrencyCode     decimal.CurrencyCode
	Memo             string
	PostedBy         string
	PostedAt         time.Time
	VoidReason       string
	VoidedBy         string
	VoidedAt         time.Time
	CreatedAt        time.Time
	CreatedBy        string
	UpdatedAt        time.Time
	UpdatedBy        string
	Version          int
}

type CashTransactionAllocation struct {
	ID         string
	TargetType CashAllocationTargetType
	TargetID   string
	TargetNo   string
	Amount     decimal.Decimal
}

type NewCashTransactionInput struct {
	ID               string
	OrgID            string
	TransactionNo    string
	Direction        CashTransactionDirection
	Status           CashTransactionStatus
	BusinessDate     time.Time
	CounterpartyID   string
	CounterpartyName string
	PaymentMethod    string
	ReferenceNo      string
	Allocations      []NewCashTransactionAllocationInput
	TotalAmount      string
	CurrencyCode     string
	Memo             string
	CreatedAt        time.Time
	CreatedBy        string
	UpdatedAt        time.Time
	UpdatedBy        string
}

type NewCashTransactionAllocationInput struct {
	ID         string
	TargetType CashAllocationTargetType
	TargetID   string
	TargetNo   string
	Amount     string
}

func NewCashTransaction(input NewCashTransactionInput) (CashTransaction, error) {
	status := NormalizeCashTransactionStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = CashTransactionStatusDraft
	}
	if status != CashTransactionStatusDraft {
		return CashTransaction{}, ErrCashTransactionInvalidStatus
	}
	direction := NormalizeCashTransactionDirection(input.Direction)
	if !IsValidCashTransactionDirection(direction) {
		return CashTransaction{}, ErrCashTransactionInvalidDirection
	}
	total, err := NewMoneyAmount(input.TotalAmount, input.CurrencyCode)
	if err != nil || total.Amount.IsZero() {
		return CashTransaction{}, ErrCashTransactionInvalidAmount
	}
	allocations, allocationsTotal, err := normalizeCashTransactionAllocations(
		input.Allocations,
		direction,
		total.CurrencyCode.String(),
	)
	if err != nil {
		return CashTransaction{}, err
	}
	if compareMoney(allocationsTotal, total.Amount) != 0 {
		return CashTransaction{}, ErrCashTransactionInvalidAmount
	}

	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	updatedBy := strings.TrimSpace(input.UpdatedBy)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(input.CreatedBy)
	}

	transaction := CashTransaction{
		ID:               strings.TrimSpace(input.ID),
		OrgID:            strings.TrimSpace(input.OrgID),
		TransactionNo:    strings.ToUpper(strings.TrimSpace(input.TransactionNo)),
		Direction:        direction,
		Status:           status,
		BusinessDate:     input.BusinessDate.UTC(),
		CounterpartyID:   strings.TrimSpace(input.CounterpartyID),
		CounterpartyName: strings.TrimSpace(input.CounterpartyName),
		PaymentMethod:    strings.TrimSpace(input.PaymentMethod),
		ReferenceNo:      strings.ToUpper(strings.TrimSpace(input.ReferenceNo)),
		Allocations:      allocations,
		TotalAmount:      total.Amount,
		CurrencyCode:     total.CurrencyCode,
		Memo:             strings.TrimSpace(input.Memo),
		CreatedAt:        createdAt.UTC(),
		CreatedBy:        strings.TrimSpace(input.CreatedBy),
		UpdatedAt:        updatedAt.UTC(),
		UpdatedBy:        updatedBy,
		Version:          1,
	}
	if err := transaction.Validate(); err != nil {
		return CashTransaction{}, err
	}

	return transaction, nil
}

func (t CashTransaction) Validate() error {
	if strings.TrimSpace(t.ID) == "" ||
		strings.TrimSpace(t.OrgID) == "" ||
		strings.TrimSpace(t.TransactionNo) == "" ||
		t.BusinessDate.IsZero() ||
		strings.TrimSpace(t.CounterpartyName) == "" ||
		strings.TrimSpace(t.PaymentMethod) == "" ||
		strings.TrimSpace(t.CreatedBy) == "" ||
		t.CreatedAt.IsZero() {
		return ErrCashTransactionRequiredField
	}
	if !IsValidCashTransactionStatus(t.Status) {
		return ErrCashTransactionInvalidStatus
	}
	if !IsValidCashTransactionDirection(t.Direction) {
		return ErrCashTransactionInvalidDirection
	}
	if t.CurrencyCode != decimal.CurrencyVND {
		return ErrCashTransactionInvalidAmount
	}
	total, err := NewMoneyAmount(t.TotalAmount.String(), t.CurrencyCode.String())
	if err != nil || total.Amount.IsZero() {
		return ErrCashTransactionInvalidAmount
	}
	if len(t.Allocations) == 0 {
		return ErrCashTransactionRequiredField
	}
	allocationsTotal := decimal.MustMoneyAmount("0")
	for _, allocation := range t.Allocations {
		if err := allocation.Validate(t.Direction); err != nil {
			return err
		}
		allocationsTotal, err = addMoney(allocationsTotal, allocation.Amount)
		if err != nil {
			return ErrCashTransactionInvalidAmount
		}
	}
	if compareMoney(allocationsTotal, t.TotalAmount) != 0 {
		return ErrCashTransactionInvalidAmount
	}
	if t.Status == CashTransactionStatusPosted &&
		(strings.TrimSpace(t.PostedBy) == "" || t.PostedAt.IsZero()) {
		return ErrCashTransactionRequiredField
	}
	if t.Status == CashTransactionStatusVoid &&
		(strings.TrimSpace(t.VoidReason) == "" || strings.TrimSpace(t.VoidedBy) == "" || t.VoidedAt.IsZero()) {
		return ErrCashTransactionRequiredField
	}

	return nil
}

func (a CashTransactionAllocation) Validate(direction CashTransactionDirection) error {
	if strings.TrimSpace(a.ID) == "" ||
		strings.TrimSpace(a.TargetID) == "" ||
		strings.TrimSpace(a.TargetNo) == "" {
		return ErrCashTransactionRequiredField
	}
	if !IsValidCashAllocationTargetType(a.TargetType) || !cashAllocationAllowedForDirection(direction, a.TargetType) {
		return ErrCashTransactionInvalidAllocation
	}
	amount, err := NewMoneyAmount(a.Amount.String(), decimal.CurrencyVND.String())
	if err != nil || amount.Amount.IsZero() {
		return ErrCashTransactionInvalidAmount
	}

	return nil
}

func (t CashTransaction) Post(actorID string, postedAt time.Time) (CashTransaction, error) {
	if NormalizeCashTransactionStatus(t.Status) != CashTransactionStatusDraft {
		return CashTransaction{}, ErrCashTransactionInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return CashTransaction{}, ErrCashTransactionRequiredField
	}
	if postedAt.IsZero() {
		postedAt = time.Now().UTC()
	}

	updated := t.Clone()
	updated.Status = CashTransactionStatusPosted
	updated.PostedBy = actorID
	updated.PostedAt = postedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = postedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (t CashTransaction) Void(actorID string, reason string, voidedAt time.Time) (CashTransaction, error) {
	if NormalizeCashTransactionStatus(t.Status) == CashTransactionStatusVoid {
		return CashTransaction{}, ErrCashTransactionInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return CashTransaction{}, ErrCashTransactionRequiredField
	}
	if voidedAt.IsZero() {
		voidedAt = time.Now().UTC()
	}

	updated := t.Clone()
	updated.Status = CashTransactionStatusVoid
	updated.VoidReason = reason
	updated.VoidedBy = actorID
	updated.VoidedAt = voidedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = voidedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (t CashTransaction) Clone() CashTransaction {
	cloned := t
	cloned.Allocations = append([]CashTransactionAllocation(nil), t.Allocations...)
	return cloned
}

func NormalizeCashAllocationTargetType(targetType CashAllocationTargetType) CashAllocationTargetType {
	return CashAllocationTargetType(normalizeStatus(string(targetType)))
}

func IsValidCashAllocationTargetType(targetType CashAllocationTargetType) bool {
	switch NormalizeCashAllocationTargetType(targetType) {
	case CashAllocationTargetCustomerReceivable,
		CashAllocationTargetSupplierPayable,
		CashAllocationTargetCODRemittance,
		CashAllocationTargetPaymentRequest,
		CashAllocationTargetManualAdjustment:
		return true
	default:
		return false
	}
}

func normalizeCashTransactionAllocations(
	inputs []NewCashTransactionAllocationInput,
	direction CashTransactionDirection,
	currencyCode string,
) ([]CashTransactionAllocation, decimal.Decimal, error) {
	if len(inputs) == 0 {
		return nil, "", ErrCashTransactionRequiredField
	}
	total := decimal.MustMoneyAmount("0")
	allocations := make([]CashTransactionAllocation, 0, len(inputs))
	for _, input := range inputs {
		amount, err := NewMoneyAmount(input.Amount, currencyCode)
		if err != nil || amount.Amount.IsZero() {
			return nil, "", ErrCashTransactionInvalidAmount
		}
		allocation := CashTransactionAllocation{
			ID:         strings.TrimSpace(input.ID),
			TargetType: NormalizeCashAllocationTargetType(input.TargetType),
			TargetID:   strings.TrimSpace(input.TargetID),
			TargetNo:   strings.ToUpper(strings.TrimSpace(input.TargetNo)),
			Amount:     amount.Amount,
		}
		if err := allocation.Validate(direction); err != nil {
			return nil, "", err
		}
		total, err = addMoney(total, allocation.Amount)
		if err != nil {
			return nil, "", ErrCashTransactionInvalidAmount
		}
		allocations = append(allocations, allocation)
	}

	return allocations, total, nil
}

func cashAllocationAllowedForDirection(direction CashTransactionDirection, targetType CashAllocationTargetType) bool {
	switch NormalizeCashTransactionDirection(direction) {
	case CashTransactionDirectionIn:
		switch NormalizeCashAllocationTargetType(targetType) {
		case CashAllocationTargetCustomerReceivable,
			CashAllocationTargetCODRemittance,
			CashAllocationTargetManualAdjustment:
			return true
		default:
			return false
		}
	case CashTransactionDirectionOut:
		switch NormalizeCashAllocationTargetType(targetType) {
		case CashAllocationTargetSupplierPayable,
			CashAllocationTargetPaymentRequest,
			CashAllocationTargetManualAdjustment:
			return true
		default:
			return false
		}
	default:
		return false
	}
}
