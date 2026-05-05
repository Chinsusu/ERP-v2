package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSupplierInvoiceRequiredField = errors.New("supplier invoice required field is missing")
var ErrSupplierInvoiceInvalidStatus = errors.New("supplier invoice status is invalid")
var ErrSupplierInvoiceInvalidAmount = errors.New("supplier invoice amount is invalid")
var ErrSupplierInvoiceInvalidSource = errors.New("supplier invoice source document is invalid")
var ErrSupplierInvoiceInvalidTransition = errors.New("supplier invoice status transition is invalid")

type SupplierInvoiceStatus string
type SupplierInvoiceMatchStatus string

const (
	SupplierInvoiceStatusDraft    SupplierInvoiceStatus = "draft"
	SupplierInvoiceStatusMatched  SupplierInvoiceStatus = "matched"
	SupplierInvoiceStatusMismatch SupplierInvoiceStatus = "mismatch"
	SupplierInvoiceStatusVoid     SupplierInvoiceStatus = "void"

	SupplierInvoiceMatchStatusPending  SupplierInvoiceMatchStatus = "pending"
	SupplierInvoiceMatchStatusMatched  SupplierInvoiceMatchStatus = "matched"
	SupplierInvoiceMatchStatusMismatch SupplierInvoiceMatchStatus = "mismatch"
)

type SupplierInvoice struct {
	ID             string
	OrgID          string
	InvoiceNo      string
	SupplierID     string
	SupplierCode   string
	SupplierName   string
	PayableID      string
	PayableNo      string
	Status         SupplierInvoiceStatus
	MatchStatus    SupplierInvoiceMatchStatus
	SourceDocument SourceDocumentRef
	Lines          []SupplierInvoiceLine
	InvoiceAmount  decimal.Decimal
	ExpectedAmount decimal.Decimal
	VarianceAmount decimal.Decimal
	CurrencyCode   decimal.CurrencyCode
	InvoiceDate    time.Time
	VoidReason     string
	VoidedBy       string
	VoidedAt       time.Time
	CreatedAt      time.Time
	CreatedBy      string
	UpdatedAt      time.Time
	UpdatedBy      string
	Version        int
}

type SupplierInvoiceLine struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentRef
	Amount         decimal.Decimal
}

type NewSupplierInvoiceInput struct {
	ID             string
	OrgID          string
	InvoiceNo      string
	SupplierID     string
	SupplierCode   string
	SupplierName   string
	PayableID      string
	PayableNo      string
	Status         SupplierInvoiceStatus
	MatchStatus    SupplierInvoiceMatchStatus
	SourceDocument SourceDocumentRef
	Lines          []NewSupplierInvoiceLineInput
	InvoiceAmount  string
	ExpectedAmount string
	VarianceAmount string
	CurrencyCode   string
	InvoiceDate    time.Time
	CreatedAt      time.Time
	CreatedBy      string
	UpdatedAt      time.Time
	UpdatedBy      string
}

type NewSupplierInvoiceLineInput struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentRef
	Amount         string
}

func NormalizeSupplierInvoiceStatus(status SupplierInvoiceStatus) SupplierInvoiceStatus {
	return SupplierInvoiceStatus(normalizeStatus(string(status)))
}

func IsValidSupplierInvoiceStatus(status SupplierInvoiceStatus) bool {
	switch NormalizeSupplierInvoiceStatus(status) {
	case SupplierInvoiceStatusDraft,
		SupplierInvoiceStatusMatched,
		SupplierInvoiceStatusMismatch,
		SupplierInvoiceStatusVoid:
		return true
	default:
		return false
	}
}

func NormalizeSupplierInvoiceMatchStatus(status SupplierInvoiceMatchStatus) SupplierInvoiceMatchStatus {
	return SupplierInvoiceMatchStatus(normalizeStatus(string(status)))
}

func IsValidSupplierInvoiceMatchStatus(status SupplierInvoiceMatchStatus) bool {
	switch NormalizeSupplierInvoiceMatchStatus(status) {
	case SupplierInvoiceMatchStatusPending,
		SupplierInvoiceMatchStatusMatched,
		SupplierInvoiceMatchStatusMismatch:
		return true
	default:
		return false
	}
}

func NewSupplierInvoice(input NewSupplierInvoiceInput) (SupplierInvoice, error) {
	status := NormalizeSupplierInvoiceStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = SupplierInvoiceStatusDraft
	}
	if !IsValidSupplierInvoiceStatus(status) || status == SupplierInvoiceStatusVoid {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidStatus
	}
	matchStatus := NormalizeSupplierInvoiceMatchStatus(input.MatchStatus)
	if strings.TrimSpace(string(matchStatus)) == "" {
		matchStatus = SupplierInvoiceMatchStatusPending
	}
	if !IsValidSupplierInvoiceMatchStatus(matchStatus) {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidStatus
	}
	sourceDocument, err := newSupplierInvoiceSourceDocumentRef(input.SourceDocument)
	if err != nil {
		return SupplierInvoice{}, err
	}
	invoiceAmount, err := NewMoneyAmount(input.InvoiceAmount, input.CurrencyCode)
	if err != nil || invoiceAmount.Amount.IsZero() {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidAmount
	}
	expectedAmount, err := NewMoneyAmount(input.ExpectedAmount, input.CurrencyCode)
	if err != nil || expectedAmount.Amount.IsZero() {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidAmount
	}
	varianceAmount, err := NewSignedMoneyAmount(input.VarianceAmount, input.CurrencyCode)
	if err != nil {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidAmount
	}
	lines, linesTotal, err := normalizeSupplierInvoiceLines(input.Lines, input.CurrencyCode)
	if err != nil {
		return SupplierInvoice{}, err
	}
	if compareMoney(linesTotal, invoiceAmount.Amount) != 0 {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidAmount
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

	invoice := SupplierInvoice{
		ID:             strings.TrimSpace(input.ID),
		OrgID:          strings.TrimSpace(input.OrgID),
		InvoiceNo:      strings.ToUpper(strings.TrimSpace(input.InvoiceNo)),
		SupplierID:     strings.TrimSpace(input.SupplierID),
		SupplierCode:   strings.ToUpper(strings.TrimSpace(input.SupplierCode)),
		SupplierName:   strings.TrimSpace(input.SupplierName),
		PayableID:      strings.TrimSpace(input.PayableID),
		PayableNo:      strings.ToUpper(strings.TrimSpace(input.PayableNo)),
		Status:         status,
		MatchStatus:    matchStatus,
		SourceDocument: sourceDocument,
		Lines:          lines,
		InvoiceAmount:  invoiceAmount.Amount,
		ExpectedAmount: expectedAmount.Amount,
		VarianceAmount: varianceAmount.Amount,
		CurrencyCode:   invoiceAmount.CurrencyCode,
		InvoiceDate:    input.InvoiceDate.UTC(),
		CreatedAt:      createdAt.UTC(),
		CreatedBy:      strings.TrimSpace(input.CreatedBy),
		UpdatedAt:      updatedAt.UTC(),
		UpdatedBy:      updatedBy,
		Version:        1,
	}
	if err := invoice.Validate(); err != nil {
		return SupplierInvoice{}, err
	}

	return invoice, nil
}

func (i SupplierInvoice) Validate() error {
	if strings.TrimSpace(i.ID) == "" ||
		strings.TrimSpace(i.OrgID) == "" ||
		strings.TrimSpace(i.InvoiceNo) == "" ||
		strings.TrimSpace(i.SupplierID) == "" ||
		strings.TrimSpace(i.SupplierName) == "" ||
		strings.TrimSpace(i.PayableID) == "" ||
		strings.TrimSpace(i.PayableNo) == "" ||
		strings.TrimSpace(i.CreatedBy) == "" ||
		i.CreatedAt.IsZero() ||
		i.InvoiceDate.IsZero() {
		return ErrSupplierInvoiceRequiredField
	}
	if !IsValidSupplierInvoiceStatus(i.Status) || !IsValidSupplierInvoiceMatchStatus(i.MatchStatus) {
		return ErrSupplierInvoiceInvalidStatus
	}
	if i.CurrencyCode != decimal.CurrencyVND {
		return ErrSupplierInvoiceInvalidAmount
	}
	if !isValidSupplierInvoiceSourceDocument(i.SourceDocument) {
		return ErrSupplierInvoiceInvalidSource
	}
	if err := validateSupplierInvoiceAmounts(i); err != nil {
		return err
	}
	if len(i.Lines) == 0 {
		return ErrSupplierInvoiceRequiredField
	}
	linesTotal := decimal.MustMoneyAmount("0")
	for _, line := range i.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
		var err error
		linesTotal, err = addMoney(linesTotal, line.Amount)
		if err != nil {
			return ErrSupplierInvoiceInvalidAmount
		}
	}
	if compareMoney(linesTotal, i.InvoiceAmount) != 0 {
		return ErrSupplierInvoiceInvalidAmount
	}
	if i.Status == SupplierInvoiceStatusMatched &&
		(i.MatchStatus != SupplierInvoiceMatchStatusMatched || !i.VarianceAmount.IsZero()) {
		return ErrSupplierInvoiceInvalidStatus
	}
	if i.Status == SupplierInvoiceStatusMismatch &&
		(i.MatchStatus != SupplierInvoiceMatchStatusMismatch || i.VarianceAmount.IsZero()) {
		return ErrSupplierInvoiceInvalidStatus
	}
	if i.Status == SupplierInvoiceStatusVoid &&
		(strings.TrimSpace(i.VoidReason) == "" || strings.TrimSpace(i.VoidedBy) == "" || i.VoidedAt.IsZero()) {
		return ErrSupplierInvoiceRequiredField
	}

	return nil
}

func (l SupplierInvoiceLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" || strings.TrimSpace(l.Description) == "" {
		return ErrSupplierInvoiceRequiredField
	}
	if !isValidSupplierInvoiceSourceDocument(l.SourceDocument) {
		return ErrSupplierInvoiceInvalidSource
	}
	amount, err := decimal.ParseMoneyAmount(l.Amount.String())
	if err != nil || amount.IsZero() {
		return ErrSupplierInvoiceInvalidAmount
	}

	return nil
}

func (i SupplierInvoice) Void(actorID string, reason string, voidedAt time.Time) (SupplierInvoice, error) {
	if NormalizeSupplierInvoiceStatus(i.Status) == SupplierInvoiceStatusVoid {
		return SupplierInvoice{}, ErrSupplierInvoiceInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return SupplierInvoice{}, ErrSupplierInvoiceRequiredField
	}
	if voidedAt.IsZero() {
		voidedAt = time.Now().UTC()
	}

	updated := i.Clone()
	updated.Status = SupplierInvoiceStatusVoid
	updated.VoidReason = reason
	updated.VoidedBy = actorID
	updated.VoidedAt = voidedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = voidedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (i SupplierInvoice) Clone() SupplierInvoice {
	cloned := i
	cloned.Lines = append([]SupplierInvoiceLine(nil), i.Lines...)
	return cloned
}

func normalizeSupplierInvoiceLines(
	inputs []NewSupplierInvoiceLineInput,
	currencyCode string,
) ([]SupplierInvoiceLine, decimal.Decimal, error) {
	if len(inputs) == 0 {
		return nil, "", ErrSupplierInvoiceRequiredField
	}
	total := decimal.MustMoneyAmount("0")
	lines := make([]SupplierInvoiceLine, 0, len(inputs))
	for _, input := range inputs {
		source, err := newSupplierInvoiceSourceDocumentRef(input.SourceDocument)
		if err != nil {
			return nil, "", err
		}
		amount, err := NewSignedMoneyAmount(input.Amount, currencyCode)
		if err != nil || amount.Amount.IsZero() {
			return nil, "", ErrSupplierInvoiceInvalidAmount
		}
		line := SupplierInvoiceLine{
			ID:             strings.TrimSpace(input.ID),
			Description:    strings.TrimSpace(input.Description),
			SourceDocument: source,
			Amount:         amount.Amount,
		}
		if err := line.Validate(); err != nil {
			return nil, "", err
		}
		total, err = addMoney(total, line.Amount)
		if err != nil {
			return nil, "", ErrSupplierInvoiceInvalidAmount
		}
		lines = append(lines, line)
	}

	return lines, total, nil
}

func validateSupplierInvoiceAmounts(i SupplierInvoice) error {
	invoiceAmount, err := decimal.ParseMoneyAmount(i.InvoiceAmount.String())
	if err != nil || invoiceAmount.IsNegative() || invoiceAmount.IsZero() {
		return ErrSupplierInvoiceInvalidAmount
	}
	expectedAmount, err := decimal.ParseMoneyAmount(i.ExpectedAmount.String())
	if err != nil || expectedAmount.IsNegative() || expectedAmount.IsZero() {
		return ErrSupplierInvoiceInvalidAmount
	}
	varianceAmount, err := decimal.ParseMoneyAmount(i.VarianceAmount.String())
	if err != nil {
		return ErrSupplierInvoiceInvalidAmount
	}
	expectedVariance, err := subtractMoney(invoiceAmount, expectedAmount)
	if err != nil || compareMoney(expectedVariance, varianceAmount) != 0 {
		return ErrSupplierInvoiceInvalidAmount
	}

	return nil
}

func newSupplierInvoiceSourceDocumentRef(source SourceDocumentRef) (SourceDocumentRef, error) {
	ref, err := NewSourceDocumentRef(source.Type, source.ID, source.No)
	if err != nil || !isSupplierInvoiceSourceDocumentType(ref.Type) {
		return SourceDocumentRef{}, ErrSupplierInvoiceInvalidSource
	}

	return ref, nil
}

func isValidSupplierInvoiceSourceDocument(source SourceDocumentRef) bool {
	return IsValidSourceDocumentType(source.Type) &&
		isSupplierInvoiceSourceDocumentType(source.Type) &&
		(source.ID != "" || source.No != "")
}

func isSupplierInvoiceSourceDocumentType(sourceType SourceDocumentType) bool {
	switch NormalizeSourceDocumentType(sourceType) {
	case SourceDocumentTypeSupplierPayable,
		SourceDocumentTypePurchaseOrder,
		SourceDocumentTypeWarehouseReceipt,
		SourceDocumentTypeQCInspection,
		SourceDocumentTypeManualAdjustment:
		return true
	default:
		return false
	}
}
