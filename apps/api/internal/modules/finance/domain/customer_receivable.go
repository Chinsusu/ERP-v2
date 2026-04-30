package domain

import (
	"errors"
	"math/big"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrCustomerReceivableRequiredField = errors.New("customer receivable required field is missing")
var ErrCustomerReceivableInvalidStatus = errors.New("customer receivable status is invalid")
var ErrCustomerReceivableInvalidAmount = errors.New("customer receivable amount is invalid")
var ErrCustomerReceivableInvalidSource = errors.New("customer receivable source document is invalid")
var ErrCustomerReceivableInvalidTransition = errors.New("customer receivable status transition is invalid")

type CustomerReceivable struct {
	ID                string
	OrgID             string
	ReceivableNo      string
	CustomerID        string
	CustomerCode      string
	CustomerName      string
	Status            ReceivableStatus
	SourceDocument    SourceDocumentRef
	Lines             []CustomerReceivableLine
	TotalAmount       decimal.Decimal
	PaidAmount        decimal.Decimal
	OutstandingAmount decimal.Decimal
	CurrencyCode      decimal.CurrencyCode
	DueDate           time.Time
	DisputeReason     string
	DisputedBy        string
	DisputedAt        time.Time
	VoidReason        string
	VoidedBy          string
	VoidedAt          time.Time
	LastReceiptBy     string
	LastReceiptAt     time.Time
	CreatedAt         time.Time
	CreatedBy         string
	UpdatedAt         time.Time
	UpdatedBy         string
	Version           int
}

type CustomerReceivableLine struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentRef
	Amount         decimal.Decimal
}

type NewCustomerReceivableInput struct {
	ID             string
	OrgID          string
	ReceivableNo   string
	CustomerID     string
	CustomerCode   string
	CustomerName   string
	Status         ReceivableStatus
	SourceDocument SourceDocumentRef
	Lines          []NewCustomerReceivableLineInput
	TotalAmount    string
	CurrencyCode   string
	DueDate        time.Time
	CreatedAt      time.Time
	CreatedBy      string
	UpdatedAt      time.Time
	UpdatedBy      string
}

type NewCustomerReceivableLineInput struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentRef
	Amount         string
}

func NewCustomerReceivable(input NewCustomerReceivableInput) (CustomerReceivable, error) {
	status := NormalizeReceivableStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = ReceivableStatusOpen
	}
	if !IsValidReceivableStatus(status) {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidStatus
	}
	if status != ReceivableStatusDraft && status != ReceivableStatusOpen {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidStatus
	}
	sourceDocument, err := NewSourceDocumentRef(
		input.SourceDocument.Type,
		input.SourceDocument.ID,
		input.SourceDocument.No,
	)
	if err != nil {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidSource
	}

	total, err := NewMoneyAmount(input.TotalAmount, input.CurrencyCode)
	if err != nil || total.Amount.IsZero() {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidAmount
	}
	lines, linesTotal, err := normalizeCustomerReceivableLines(input.Lines, input.CurrencyCode)
	if err != nil {
		return CustomerReceivable{}, err
	}
	if compareMoney(linesTotal, total.Amount) != 0 {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidAmount
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
	zero, err := decimal.ParseMoneyAmount("0")
	if err != nil {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidAmount
	}

	receivable := CustomerReceivable{
		ID:                strings.TrimSpace(input.ID),
		OrgID:             strings.TrimSpace(input.OrgID),
		ReceivableNo:      strings.ToUpper(strings.TrimSpace(input.ReceivableNo)),
		CustomerID:        strings.TrimSpace(input.CustomerID),
		CustomerCode:      strings.ToUpper(strings.TrimSpace(input.CustomerCode)),
		CustomerName:      strings.TrimSpace(input.CustomerName),
		Status:            status,
		SourceDocument:    sourceDocument,
		Lines:             lines,
		TotalAmount:       total.Amount,
		PaidAmount:        zero,
		OutstandingAmount: total.Amount,
		CurrencyCode:      total.CurrencyCode,
		DueDate:           input.DueDate.UTC(),
		CreatedAt:         createdAt.UTC(),
		CreatedBy:         strings.TrimSpace(input.CreatedBy),
		UpdatedAt:         updatedAt.UTC(),
		UpdatedBy:         updatedBy,
		Version:           1,
	}
	if err := receivable.Validate(); err != nil {
		return CustomerReceivable{}, err
	}

	return receivable, nil
}

func (r CustomerReceivable) Validate() error {
	if strings.TrimSpace(r.ID) == "" ||
		strings.TrimSpace(r.OrgID) == "" ||
		strings.TrimSpace(r.ReceivableNo) == "" ||
		strings.TrimSpace(r.CustomerID) == "" ||
		strings.TrimSpace(r.CustomerName) == "" ||
		strings.TrimSpace(r.CreatedBy) == "" ||
		r.CreatedAt.IsZero() {
		return ErrCustomerReceivableRequiredField
	}
	if !IsValidReceivableStatus(r.Status) {
		return ErrCustomerReceivableInvalidStatus
	}
	if r.CurrencyCode != decimal.CurrencyVND {
		return ErrCustomerReceivableInvalidAmount
	}
	if !IsValidSourceDocumentType(r.SourceDocument.Type) ||
		(r.SourceDocument.ID == "" && r.SourceDocument.No == "") {
		return ErrCustomerReceivableInvalidSource
	}
	if err := validateCustomerReceivableAmounts(r); err != nil {
		return err
	}
	if len(r.Lines) == 0 {
		return ErrCustomerReceivableRequiredField
	}
	linesTotal := decimal.MustMoneyAmount("0")
	for _, line := range r.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
		var err error
		linesTotal, err = addMoney(linesTotal, line.Amount)
		if err != nil {
			return ErrCustomerReceivableInvalidAmount
		}
	}
	if compareMoney(linesTotal, r.TotalAmount) != 0 {
		return ErrCustomerReceivableInvalidAmount
	}
	if r.Status == ReceivableStatusDisputed &&
		(strings.TrimSpace(r.DisputeReason) == "" || strings.TrimSpace(r.DisputedBy) == "" || r.DisputedAt.IsZero()) {
		return ErrCustomerReceivableRequiredField
	}
	if r.Status == ReceivableStatusVoid &&
		(strings.TrimSpace(r.VoidReason) == "" || strings.TrimSpace(r.VoidedBy) == "" || r.VoidedAt.IsZero()) {
		return ErrCustomerReceivableRequiredField
	}

	return nil
}

func (l CustomerReceivableLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" || strings.TrimSpace(l.Description) == "" {
		return ErrCustomerReceivableRequiredField
	}
	if !IsValidSourceDocumentType(l.SourceDocument.Type) ||
		(l.SourceDocument.ID == "" && l.SourceDocument.No == "") {
		return ErrCustomerReceivableInvalidSource
	}
	amount, err := decimal.ParseMoneyAmount(l.Amount.String())
	if err != nil || amount.IsNegative() || amount.IsZero() {
		return ErrCustomerReceivableInvalidAmount
	}

	return nil
}

func (r CustomerReceivable) RecordReceipt(
	amountValue string,
	actorID string,
	receivedAt time.Time,
) (CustomerReceivable, error) {
	status := NormalizeReceivableStatus(r.Status)
	if status != ReceivableStatusOpen && status != ReceivableStatusPartiallyPaid {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return CustomerReceivable{}, ErrCustomerReceivableRequiredField
	}
	receipt, err := NewMoneyAmount(amountValue, r.CurrencyCode.String())
	if err != nil || receipt.Amount.IsZero() || compareMoney(receipt.Amount, r.OutstandingAmount) > 0 {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidAmount
	}
	if receivedAt.IsZero() {
		receivedAt = time.Now().UTC()
	}

	paidAmount, err := addMoney(r.PaidAmount, receipt.Amount)
	if err != nil {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidAmount
	}
	outstandingAmount, err := subtractMoney(r.OutstandingAmount, receipt.Amount)
	if err != nil || outstandingAmount.IsNegative() {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidAmount
	}

	updated := r.Clone()
	updated.PaidAmount = paidAmount
	updated.OutstandingAmount = outstandingAmount
	updated.LastReceiptBy = actorID
	updated.LastReceiptAt = receivedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = receivedAt.UTC()
	updated.Version++
	if outstandingAmount.IsZero() {
		updated.Status = ReceivableStatusPaid
	} else {
		updated.Status = ReceivableStatusPartiallyPaid
	}

	return updated, updated.Validate()
}

func (r CustomerReceivable) MarkDisputed(
	actorID string,
	reason string,
	disputedAt time.Time,
) (CustomerReceivable, error) {
	status := NormalizeReceivableStatus(r.Status)
	if status != ReceivableStatusOpen && status != ReceivableStatusPartiallyPaid {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return CustomerReceivable{}, ErrCustomerReceivableRequiredField
	}
	if disputedAt.IsZero() {
		disputedAt = time.Now().UTC()
	}

	updated := r.Clone()
	updated.Status = ReceivableStatusDisputed
	updated.DisputeReason = reason
	updated.DisputedBy = actorID
	updated.DisputedAt = disputedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = disputedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CustomerReceivable) Void(actorID string, reason string, voidedAt time.Time) (CustomerReceivable, error) {
	status := NormalizeReceivableStatus(r.Status)
	if status == ReceivableStatusPaid || status == ReceivableStatusVoid {
		return CustomerReceivable{}, ErrCustomerReceivableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return CustomerReceivable{}, ErrCustomerReceivableRequiredField
	}
	if voidedAt.IsZero() {
		voidedAt = time.Now().UTC()
	}

	updated := r.Clone()
	updated.Status = ReceivableStatusVoid
	updated.VoidReason = reason
	updated.VoidedBy = actorID
	updated.VoidedAt = voidedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = voidedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CustomerReceivable) Clone() CustomerReceivable {
	cloned := r
	cloned.Lines = append([]CustomerReceivableLine(nil), r.Lines...)
	return cloned
}

func normalizeCustomerReceivableLines(
	inputs []NewCustomerReceivableLineInput,
	currencyCode string,
) ([]CustomerReceivableLine, decimal.Decimal, error) {
	if len(inputs) == 0 {
		return nil, "", ErrCustomerReceivableRequiredField
	}
	total := decimal.MustMoneyAmount("0")
	lines := make([]CustomerReceivableLine, 0, len(inputs))
	for _, input := range inputs {
		sourceDocument, err := NewSourceDocumentRef(
			input.SourceDocument.Type,
			input.SourceDocument.ID,
			input.SourceDocument.No,
		)
		if err != nil {
			return nil, "", ErrCustomerReceivableInvalidSource
		}
		amount, err := NewMoneyAmount(input.Amount, currencyCode)
		if err != nil || amount.Amount.IsZero() {
			return nil, "", ErrCustomerReceivableInvalidAmount
		}
		line := CustomerReceivableLine{
			ID:             strings.TrimSpace(input.ID),
			Description:    strings.TrimSpace(input.Description),
			SourceDocument: sourceDocument,
			Amount:         amount.Amount,
		}
		if err := line.Validate(); err != nil {
			return nil, "", err
		}
		total, err = addMoney(total, line.Amount)
		if err != nil {
			return nil, "", ErrCustomerReceivableInvalidAmount
		}
		lines = append(lines, line)
	}

	return lines, total, nil
}

func validateCustomerReceivableAmounts(r CustomerReceivable) error {
	total, err := decimal.ParseMoneyAmount(r.TotalAmount.String())
	if err != nil || total.IsNegative() || total.IsZero() {
		return ErrCustomerReceivableInvalidAmount
	}
	paid, err := decimal.ParseMoneyAmount(r.PaidAmount.String())
	if err != nil || paid.IsNegative() {
		return ErrCustomerReceivableInvalidAmount
	}
	outstanding, err := decimal.ParseMoneyAmount(r.OutstandingAmount.String())
	if err != nil || outstanding.IsNegative() {
		return ErrCustomerReceivableInvalidAmount
	}
	sum, err := addMoney(paid, outstanding)
	if err != nil || compareMoney(sum, total) != 0 {
		return ErrCustomerReceivableInvalidAmount
	}

	return nil
}

func addMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	leftCents, err := moneyCents(left)
	if err != nil {
		return "", err
	}
	rightCents, err := moneyCents(right)
	if err != nil {
		return "", err
	}

	return centsToMoney(new(big.Int).Add(leftCents, rightCents)), nil
}

func subtractMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	leftCents, err := moneyCents(left)
	if err != nil {
		return "", err
	}
	rightCents, err := moneyCents(right)
	if err != nil {
		return "", err
	}

	return centsToMoney(new(big.Int).Sub(leftCents, rightCents)), nil
}

func compareMoney(left decimal.Decimal, right decimal.Decimal) int {
	leftCents, err := moneyCents(left)
	if err != nil {
		return -1
	}
	rightCents, err := moneyCents(right)
	if err != nil {
		return 1
	}

	return leftCents.Cmp(rightCents)
}

func moneyCents(amount decimal.Decimal) (*big.Int, error) {
	normalized, err := decimal.ParseMoneyAmount(amount.String())
	if err != nil {
		return nil, err
	}
	digits := strings.ReplaceAll(normalized.String(), ".", "")
	negative := strings.HasPrefix(digits, "-")
	digits = strings.TrimPrefix(digits, "-")
	cents, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, ErrCustomerReceivableInvalidAmount
	}
	if negative {
		cents.Neg(cents)
	}

	return cents, nil
}

func centsToMoney(cents *big.Int) decimal.Decimal {
	negative := cents.Sign() < 0
	digits := new(big.Int).Abs(cents).String()
	if len(digits) <= decimal.MoneyScale {
		digits = strings.Repeat("0", decimal.MoneyScale-len(digits)+1) + digits
	}

	intPart := digits[:len(digits)-decimal.MoneyScale]
	fracPart := digits[len(digits)-decimal.MoneyScale:]
	if negative {
		intPart = "-" + intPart
	}

	return decimal.Decimal(intPart + "." + fracPart)
}
