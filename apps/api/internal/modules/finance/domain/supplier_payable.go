package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSupplierPayableRequiredField = errors.New("supplier payable required field is missing")
var ErrSupplierPayableInvalidStatus = errors.New("supplier payable status is invalid")
var ErrSupplierPayableInvalidAmount = errors.New("supplier payable amount is invalid")
var ErrSupplierPayableInvalidSource = errors.New("supplier payable source document is invalid")
var ErrSupplierPayableInvalidTransition = errors.New("supplier payable status transition is invalid")

type SupplierPayable struct {
	ID                  string
	OrgID               string
	PayableNo           string
	SupplierID          string
	SupplierCode        string
	SupplierName        string
	Status              PayableStatus
	SourceDocument      SourceDocumentRef
	Lines               []SupplierPayableLine
	TotalAmount         decimal.Decimal
	PaidAmount          decimal.Decimal
	OutstandingAmount   decimal.Decimal
	CurrencyCode        decimal.CurrencyCode
	DueDate             time.Time
	PaymentRequestedBy  string
	PaymentRequestedAt  time.Time
	PaymentApprovedBy   string
	PaymentApprovedAt   time.Time
	PaymentRejectedBy   string
	PaymentRejectedAt   time.Time
	PaymentRejectReason string
	DisputeReason       string
	DisputedBy          string
	DisputedAt          time.Time
	VoidReason          string
	VoidedBy            string
	VoidedAt            time.Time
	LastPaymentBy       string
	LastPaymentAt       time.Time
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
}

type SupplierPayableLine struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentRef
	Amount         decimal.Decimal
}

type NewSupplierPayableInput struct {
	ID             string
	OrgID          string
	PayableNo      string
	SupplierID     string
	SupplierCode   string
	SupplierName   string
	Status         PayableStatus
	SourceDocument SourceDocumentRef
	Lines          []NewSupplierPayableLineInput
	TotalAmount    string
	CurrencyCode   string
	DueDate        time.Time
	CreatedAt      time.Time
	CreatedBy      string
	UpdatedAt      time.Time
	UpdatedBy      string
}

type NewSupplierPayableLineInput struct {
	ID             string
	Description    string
	SourceDocument SourceDocumentRef
	Amount         string
}

func NewSupplierPayable(input NewSupplierPayableInput) (SupplierPayable, error) {
	status := NormalizePayableStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = PayableStatusOpen
	}
	if !IsValidPayableStatus(status) {
		return SupplierPayable{}, ErrSupplierPayableInvalidStatus
	}
	if status != PayableStatusDraft && status != PayableStatusOpen {
		return SupplierPayable{}, ErrSupplierPayableInvalidStatus
	}
	sourceDocument, err := newPayableSourceDocumentRef(input.SourceDocument)
	if err != nil {
		return SupplierPayable{}, err
	}

	total, err := NewMoneyAmount(input.TotalAmount, input.CurrencyCode)
	if err != nil || total.Amount.IsZero() {
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
	}
	lines, linesTotal, err := normalizeSupplierPayableLines(input.Lines, input.CurrencyCode)
	if err != nil {
		return SupplierPayable{}, err
	}
	if compareMoney(linesTotal, total.Amount) != 0 {
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
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
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
	}

	payable := SupplierPayable{
		ID:                strings.TrimSpace(input.ID),
		OrgID:             strings.TrimSpace(input.OrgID),
		PayableNo:         strings.ToUpper(strings.TrimSpace(input.PayableNo)),
		SupplierID:        strings.TrimSpace(input.SupplierID),
		SupplierCode:      strings.ToUpper(strings.TrimSpace(input.SupplierCode)),
		SupplierName:      strings.TrimSpace(input.SupplierName),
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
	if err := payable.Validate(); err != nil {
		return SupplierPayable{}, err
	}

	return payable, nil
}

func (p SupplierPayable) Validate() error {
	if strings.TrimSpace(p.ID) == "" ||
		strings.TrimSpace(p.OrgID) == "" ||
		strings.TrimSpace(p.PayableNo) == "" ||
		strings.TrimSpace(p.SupplierID) == "" ||
		strings.TrimSpace(p.SupplierName) == "" ||
		strings.TrimSpace(p.CreatedBy) == "" ||
		p.CreatedAt.IsZero() {
		return ErrSupplierPayableRequiredField
	}
	if !IsValidPayableStatus(p.Status) {
		return ErrSupplierPayableInvalidStatus
	}
	if p.CurrencyCode != decimal.CurrencyVND {
		return ErrSupplierPayableInvalidAmount
	}
	if !isValidPayableSourceDocument(p.SourceDocument) {
		return ErrSupplierPayableInvalidSource
	}
	if err := validateSupplierPayableAmounts(p); err != nil {
		return err
	}
	if len(p.Lines) == 0 {
		return ErrSupplierPayableRequiredField
	}
	linesTotal := decimal.MustMoneyAmount("0")
	for _, line := range p.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
		var err error
		linesTotal, err = addMoney(linesTotal, line.Amount)
		if err != nil {
			return ErrSupplierPayableInvalidAmount
		}
	}
	if compareMoney(linesTotal, p.TotalAmount) != 0 {
		return ErrSupplierPayableInvalidAmount
	}
	if p.Status == PayableStatusPaymentRequested &&
		(strings.TrimSpace(p.PaymentRequestedBy) == "" || p.PaymentRequestedAt.IsZero()) {
		return ErrSupplierPayableRequiredField
	}
	if p.Status == PayableStatusPaymentApproved &&
		(strings.TrimSpace(p.PaymentApprovedBy) == "" || p.PaymentApprovedAt.IsZero()) {
		return ErrSupplierPayableRequiredField
	}
	if p.Status == PayableStatusDisputed &&
		(strings.TrimSpace(p.DisputeReason) == "" || strings.TrimSpace(p.DisputedBy) == "" || p.DisputedAt.IsZero()) {
		return ErrSupplierPayableRequiredField
	}
	if p.Status == PayableStatusVoid &&
		(strings.TrimSpace(p.VoidReason) == "" || strings.TrimSpace(p.VoidedBy) == "" || p.VoidedAt.IsZero()) {
		return ErrSupplierPayableRequiredField
	}

	return nil
}

func (l SupplierPayableLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" || strings.TrimSpace(l.Description) == "" {
		return ErrSupplierPayableRequiredField
	}
	if !isValidPayableSourceDocument(l.SourceDocument) {
		return ErrSupplierPayableInvalidSource
	}
	amount, err := decimal.ParseMoneyAmount(l.Amount.String())
	if err != nil || amount.IsNegative() || amount.IsZero() {
		return ErrSupplierPayableInvalidAmount
	}

	return nil
}

func (p SupplierPayable) RequestPayment(actorID string, requestedAt time.Time) (SupplierPayable, error) {
	if NormalizePayableStatus(p.Status) != PayableStatusOpen {
		return SupplierPayable{}, ErrSupplierPayableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SupplierPayable{}, ErrSupplierPayableRequiredField
	}
	if requestedAt.IsZero() {
		requestedAt = time.Now().UTC()
	}

	updated := p.Clone()
	updated.Status = PayableStatusPaymentRequested
	updated.PaymentRequestedBy = actorID
	updated.PaymentRequestedAt = requestedAt.UTC()
	updated.PaymentRejectedBy = ""
	updated.PaymentRejectedAt = time.Time{}
	updated.PaymentRejectReason = ""
	updated.UpdatedBy = actorID
	updated.UpdatedAt = requestedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (p SupplierPayable) ApprovePayment(actorID string, approvedAt time.Time) (SupplierPayable, error) {
	if NormalizePayableStatus(p.Status) != PayableStatusPaymentRequested {
		return SupplierPayable{}, ErrSupplierPayableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SupplierPayable{}, ErrSupplierPayableRequiredField
	}
	if approvedAt.IsZero() {
		approvedAt = time.Now().UTC()
	}

	updated := p.Clone()
	updated.Status = PayableStatusPaymentApproved
	updated.PaymentApprovedBy = actorID
	updated.PaymentApprovedAt = approvedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = approvedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (p SupplierPayable) RejectPayment(actorID string, reason string, rejectedAt time.Time) (SupplierPayable, error) {
	if NormalizePayableStatus(p.Status) != PayableStatusPaymentRequested {
		return SupplierPayable{}, ErrSupplierPayableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return SupplierPayable{}, ErrSupplierPayableRequiredField
	}
	if rejectedAt.IsZero() {
		rejectedAt = time.Now().UTC()
	}

	updated := p.Clone()
	updated.Status = PayableStatusOpen
	updated.PaymentApprovedBy = ""
	updated.PaymentApprovedAt = time.Time{}
	updated.PaymentRejectedBy = actorID
	updated.PaymentRejectedAt = rejectedAt.UTC()
	updated.PaymentRejectReason = reason
	updated.UpdatedBy = actorID
	updated.UpdatedAt = rejectedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (p SupplierPayable) RecordPayment(
	amountValue string,
	actorID string,
	paidAt time.Time,
) (SupplierPayable, error) {
	status := NormalizePayableStatus(p.Status)
	if status != PayableStatusPaymentApproved && status != PayableStatusPartiallyPaid {
		return SupplierPayable{}, ErrSupplierPayableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SupplierPayable{}, ErrSupplierPayableRequiredField
	}
	payment, err := NewMoneyAmount(amountValue, p.CurrencyCode.String())
	if err != nil || payment.Amount.IsZero() || compareMoney(payment.Amount, p.OutstandingAmount) > 0 {
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
	}
	if paidAt.IsZero() {
		paidAt = time.Now().UTC()
	}

	paidAmount, err := addMoney(p.PaidAmount, payment.Amount)
	if err != nil {
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
	}
	outstandingAmount, err := subtractMoney(p.OutstandingAmount, payment.Amount)
	if err != nil || outstandingAmount.IsNegative() {
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
	}

	updated := p.Clone()
	updated.PaidAmount = paidAmount
	updated.OutstandingAmount = outstandingAmount
	updated.LastPaymentBy = actorID
	updated.LastPaymentAt = paidAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = paidAt.UTC()
	updated.Version++
	if outstandingAmount.IsZero() {
		updated.Status = PayableStatusPaid
	} else {
		updated.Status = PayableStatusPartiallyPaid
	}

	return updated, updated.Validate()
}

func (p SupplierPayable) MarkDisputed(actorID string, reason string, disputedAt time.Time) (SupplierPayable, error) {
	status := NormalizePayableStatus(p.Status)
	if status == PayableStatusPaid || status == PayableStatusVoid {
		return SupplierPayable{}, ErrSupplierPayableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return SupplierPayable{}, ErrSupplierPayableRequiredField
	}
	if disputedAt.IsZero() {
		disputedAt = time.Now().UTC()
	}

	updated := p.Clone()
	updated.Status = PayableStatusDisputed
	updated.DisputeReason = reason
	updated.DisputedBy = actorID
	updated.DisputedAt = disputedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = disputedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (p SupplierPayable) Void(actorID string, reason string, voidedAt time.Time) (SupplierPayable, error) {
	status := NormalizePayableStatus(p.Status)
	if status == PayableStatusPaid || status == PayableStatusVoid {
		return SupplierPayable{}, ErrSupplierPayableInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return SupplierPayable{}, ErrSupplierPayableRequiredField
	}
	if voidedAt.IsZero() {
		voidedAt = time.Now().UTC()
	}
	zero, err := decimal.ParseMoneyAmount("0")
	if err != nil {
		return SupplierPayable{}, ErrSupplierPayableInvalidAmount
	}

	updated := p.Clone()
	updated.Status = PayableStatusVoid
	updated.OutstandingAmount = zero
	updated.VoidReason = reason
	updated.VoidedBy = actorID
	updated.VoidedAt = voidedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = voidedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (p SupplierPayable) Clone() SupplierPayable {
	cloned := p
	cloned.Lines = append([]SupplierPayableLine(nil), p.Lines...)
	return cloned
}

func normalizeSupplierPayableLines(
	inputs []NewSupplierPayableLineInput,
	currencyCode string,
) ([]SupplierPayableLine, decimal.Decimal, error) {
	if len(inputs) == 0 {
		return nil, "", ErrSupplierPayableRequiredField
	}
	total := decimal.MustMoneyAmount("0")
	lines := make([]SupplierPayableLine, 0, len(inputs))
	for _, input := range inputs {
		sourceDocument, err := newPayableSourceDocumentRef(input.SourceDocument)
		if err != nil {
			return nil, "", err
		}
		amount, err := NewMoneyAmount(input.Amount, currencyCode)
		if err != nil || amount.Amount.IsZero() {
			return nil, "", ErrSupplierPayableInvalidAmount
		}
		line := SupplierPayableLine{
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
			return nil, "", ErrSupplierPayableInvalidAmount
		}
		lines = append(lines, line)
	}

	return lines, total, nil
}

func validateSupplierPayableAmounts(p SupplierPayable) error {
	total, err := decimal.ParseMoneyAmount(p.TotalAmount.String())
	if err != nil || total.IsNegative() || total.IsZero() {
		return ErrSupplierPayableInvalidAmount
	}
	paid, err := decimal.ParseMoneyAmount(p.PaidAmount.String())
	if err != nil || paid.IsNegative() {
		return ErrSupplierPayableInvalidAmount
	}
	outstanding, err := decimal.ParseMoneyAmount(p.OutstandingAmount.String())
	if err != nil || outstanding.IsNegative() {
		return ErrSupplierPayableInvalidAmount
	}
	if p.Status == PayableStatusVoid {
		if !outstanding.IsZero() {
			return ErrSupplierPayableInvalidAmount
		}
		return nil
	}
	sum, err := addMoney(paid, outstanding)
	if err != nil || compareMoney(sum, total) != 0 {
		return ErrSupplierPayableInvalidAmount
	}
	if p.Status == PayableStatusPaid && !outstanding.IsZero() {
		return ErrSupplierPayableInvalidAmount
	}
	if p.Status == PayableStatusPartiallyPaid && (paid.IsZero() || outstanding.IsZero()) {
		return ErrSupplierPayableInvalidAmount
	}

	return nil
}

func newPayableSourceDocumentRef(source SourceDocumentRef) (SourceDocumentRef, error) {
	ref, err := NewSourceDocumentRef(source.Type, source.ID, source.No)
	if err != nil || !isPayableSourceDocumentType(ref.Type) {
		return SourceDocumentRef{}, ErrSupplierPayableInvalidSource
	}

	return ref, nil
}

func isValidPayableSourceDocument(source SourceDocumentRef) bool {
	return IsValidSourceDocumentType(source.Type) &&
		isPayableSourceDocumentType(source.Type) &&
		(source.ID != "" || source.No != "")
}

func isPayableSourceDocumentType(sourceType SourceDocumentType) bool {
	switch NormalizeSourceDocumentType(sourceType) {
	case SourceDocumentTypePurchaseOrder,
		SourceDocumentTypeWarehouseReceipt,
		SourceDocumentTypeQCInspection,
		SourceDocumentTypeSubcontractOrder,
		SourceDocumentTypeSubcontractPaymentMilestone,
		SourceDocumentTypeManualAdjustment:
		return true
	default:
		return false
	}
}
