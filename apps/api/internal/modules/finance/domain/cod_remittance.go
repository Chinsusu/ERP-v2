package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrCODRemittanceRequiredField = errors.New("cod remittance required field is missing")
var ErrCODRemittanceInvalidStatus = errors.New("cod remittance status is invalid")
var ErrCODRemittanceInvalidAmount = errors.New("cod remittance amount is invalid")
var ErrCODRemittanceInvalidTransition = errors.New("cod remittance status transition is invalid")
var ErrCODRemittanceInvalidDiscrepancy = errors.New("cod remittance discrepancy is invalid")

type CODLineMatchStatus string

const (
	CODLineMatchStatusMatched   CODLineMatchStatus = "matched"
	CODLineMatchStatusShortPaid CODLineMatchStatus = "short_paid"
	CODLineMatchStatusOverPaid  CODLineMatchStatus = "over_paid"
)

type CODDiscrepancyType string

const (
	CODDiscrepancyTypeShortPaid   CODDiscrepancyType = "short_paid"
	CODDiscrepancyTypeOverPaid    CODDiscrepancyType = "over_paid"
	CODDiscrepancyTypeCarrierFee  CODDiscrepancyType = "carrier_fee"
	CODDiscrepancyTypeReturnClaim CODDiscrepancyType = "return_claim"
	CODDiscrepancyTypeOther       CODDiscrepancyType = "other"
)

type CODDiscrepancyStatus string

const (
	CODDiscrepancyStatusOpen     CODDiscrepancyStatus = "open"
	CODDiscrepancyStatusResolved CODDiscrepancyStatus = "resolved"
)

type CODRemittance struct {
	ID                string
	OrgID             string
	RemittanceNo      string
	CarrierID         string
	CarrierCode       string
	CarrierName       string
	Status            CODRemittanceStatus
	BusinessDate      time.Time
	ExpectedAmount    decimal.Decimal
	RemittedAmount    decimal.Decimal
	DiscrepancyAmount decimal.Decimal
	CurrencyCode      decimal.CurrencyCode
	Lines             []CODRemittanceLine
	Discrepancies     []CODDiscrepancy
	SubmittedBy       string
	SubmittedAt       time.Time
	ApprovedBy        string
	ApprovedAt        time.Time
	ClosedBy          string
	ClosedAt          time.Time
	VoidReason        string
	VoidedBy          string
	VoidedAt          time.Time
	CreatedAt         time.Time
	CreatedBy         string
	UpdatedAt         time.Time
	UpdatedBy         string
	Version           int
}

type CODRemittanceLine struct {
	ID                string
	ReceivableID      string
	ReceivableNo      string
	ShipmentID        string
	TrackingNo        string
	CustomerName      string
	ExpectedAmount    decimal.Decimal
	RemittedAmount    decimal.Decimal
	DiscrepancyAmount decimal.Decimal
	MatchStatus       CODLineMatchStatus
}

type CODDiscrepancy struct {
	ID           string
	LineID       string
	ReceivableID string
	Type         CODDiscrepancyType
	Status       CODDiscrepancyStatus
	Amount       decimal.Decimal
	Reason       string
	OwnerID      string
	RecordedBy   string
	RecordedAt   time.Time
	ResolvedBy   string
	ResolvedAt   time.Time
	Resolution   string
}

type NewCODRemittanceInput struct {
	ID             string
	OrgID          string
	RemittanceNo   string
	CarrierID      string
	CarrierCode    string
	CarrierName    string
	Status         CODRemittanceStatus
	BusinessDate   time.Time
	ExpectedAmount string
	RemittedAmount string
	CurrencyCode   string
	Lines          []NewCODRemittanceLineInput
	CreatedAt      time.Time
	CreatedBy      string
	UpdatedAt      time.Time
	UpdatedBy      string
}

type NewCODRemittanceLineInput struct {
	ID             string
	ReceivableID   string
	ReceivableNo   string
	ShipmentID     string
	TrackingNo     string
	CustomerName   string
	ExpectedAmount string
	RemittedAmount string
}

type RecordCODDiscrepancyInput struct {
	ID         string
	LineID     string
	Type       CODDiscrepancyType
	Status     CODDiscrepancyStatus
	Reason     string
	OwnerID    string
	RecordedBy string
	RecordedAt time.Time
}

func NewCODRemittance(input NewCODRemittanceInput) (CODRemittance, error) {
	status := NormalizeCODRemittanceStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = CODRemittanceStatusDraft
	}
	if status != CODRemittanceStatusDraft {
		return CODRemittance{}, ErrCODRemittanceInvalidStatus
	}

	currency, err := decimal.NormalizeCurrencyCode(input.CurrencyCode)
	if err != nil || currency != decimal.CurrencyVND {
		return CODRemittance{}, ErrCODRemittanceInvalidAmount
	}
	lines, expectedAmount, remittedAmount, err := normalizeCODRemittanceLines(input.Lines, currency.String())
	if err != nil {
		return CODRemittance{}, err
	}
	expectedAmount, err = reconcileOptionalCODAmount(input.ExpectedAmount, expectedAmount, currency.String())
	if err != nil {
		return CODRemittance{}, err
	}
	remittedAmount, err = reconcileOptionalCODAmount(input.RemittedAmount, remittedAmount, currency.String())
	if err != nil {
		return CODRemittance{}, err
	}
	discrepancyAmount, err := subtractCODMoney(remittedAmount, expectedAmount)
	if err != nil {
		return CODRemittance{}, err
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

	remittance := CODRemittance{
		ID:                strings.TrimSpace(input.ID),
		OrgID:             strings.TrimSpace(input.OrgID),
		RemittanceNo:      strings.ToUpper(strings.TrimSpace(input.RemittanceNo)),
		CarrierID:         strings.TrimSpace(input.CarrierID),
		CarrierCode:       strings.ToUpper(strings.TrimSpace(input.CarrierCode)),
		CarrierName:       strings.TrimSpace(input.CarrierName),
		Status:            status,
		BusinessDate:      input.BusinessDate.UTC(),
		ExpectedAmount:    expectedAmount,
		RemittedAmount:    remittedAmount,
		DiscrepancyAmount: discrepancyAmount,
		CurrencyCode:      currency,
		Lines:             lines,
		CreatedAt:         createdAt.UTC(),
		CreatedBy:         strings.TrimSpace(input.CreatedBy),
		UpdatedAt:         updatedAt.UTC(),
		UpdatedBy:         updatedBy,
		Version:           1,
	}
	if err := remittance.Validate(); err != nil {
		return CODRemittance{}, err
	}

	return remittance, nil
}

func (r CODRemittance) Validate() error {
	if strings.TrimSpace(r.ID) == "" ||
		strings.TrimSpace(r.OrgID) == "" ||
		strings.TrimSpace(r.RemittanceNo) == "" ||
		strings.TrimSpace(r.CarrierID) == "" ||
		strings.TrimSpace(r.CarrierName) == "" ||
		r.BusinessDate.IsZero() ||
		strings.TrimSpace(r.CreatedBy) == "" ||
		r.CreatedAt.IsZero() {
		return ErrCODRemittanceRequiredField
	}
	if !IsValidCODRemittanceStatus(r.Status) {
		return ErrCODRemittanceInvalidStatus
	}
	if r.CurrencyCode != decimal.CurrencyVND {
		return ErrCODRemittanceInvalidAmount
	}
	if len(r.Lines) == 0 {
		return ErrCODRemittanceRequiredField
	}
	expectedTotal := decimal.MustMoneyAmount("0")
	remittedTotal := decimal.MustMoneyAmount("0")
	for _, line := range r.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
		var err error
		expectedTotal, err = addCODMoney(expectedTotal, line.ExpectedAmount)
		if err != nil {
			return err
		}
		remittedTotal, err = addCODMoney(remittedTotal, line.RemittedAmount)
		if err != nil {
			return err
		}
	}
	if compareMoney(expectedTotal, r.ExpectedAmount) != 0 || compareMoney(remittedTotal, r.RemittedAmount) != 0 {
		return ErrCODRemittanceInvalidAmount
	}
	discrepancyAmount, err := subtractCODMoney(r.RemittedAmount, r.ExpectedAmount)
	if err != nil || compareMoney(discrepancyAmount, r.DiscrepancyAmount) != 0 {
		return ErrCODRemittanceInvalidAmount
	}
	for _, discrepancy := range r.Discrepancies {
		if err := discrepancy.Validate(); err != nil {
			return err
		}
		if !r.hasLine(discrepancy.LineID) {
			return ErrCODRemittanceInvalidDiscrepancy
		}
	}
	if requiresCODDiscrepancyTrace(r.Status) && !r.hasTraceForEveryDiscrepantLine() {
		return ErrCODRemittanceInvalidDiscrepancy
	}
	if r.Status == CODRemittanceStatusSubmitted && (strings.TrimSpace(r.SubmittedBy) == "" || r.SubmittedAt.IsZero()) {
		return ErrCODRemittanceRequiredField
	}
	if r.Status == CODRemittanceStatusApproved && (strings.TrimSpace(r.ApprovedBy) == "" || r.ApprovedAt.IsZero()) {
		return ErrCODRemittanceRequiredField
	}
	if r.Status == CODRemittanceStatusClosed && (strings.TrimSpace(r.ClosedBy) == "" || r.ClosedAt.IsZero()) {
		return ErrCODRemittanceRequiredField
	}
	if r.Status == CODRemittanceStatusVoid &&
		(strings.TrimSpace(r.VoidReason) == "" || strings.TrimSpace(r.VoidedBy) == "" || r.VoidedAt.IsZero()) {
		return ErrCODRemittanceRequiredField
	}

	return nil
}

func (l CODRemittanceLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		strings.TrimSpace(l.ReceivableID) == "" ||
		strings.TrimSpace(l.ReceivableNo) == "" ||
		strings.TrimSpace(l.TrackingNo) == "" {
		return ErrCODRemittanceRequiredField
	}
	expected, err := NewMoneyAmount(l.ExpectedAmount.String(), decimal.CurrencyVND.String())
	if err != nil || expected.Amount.IsZero() {
		return ErrCODRemittanceInvalidAmount
	}
	remitted, err := NewMoneyAmount(l.RemittedAmount.String(), decimal.CurrencyVND.String())
	if err != nil {
		return ErrCODRemittanceInvalidAmount
	}
	discrepancy, err := subtractCODMoney(remitted.Amount, expected.Amount)
	if err != nil || compareMoney(discrepancy, l.DiscrepancyAmount) != 0 {
		return ErrCODRemittanceInvalidAmount
	}
	if !IsValidCODLineMatchStatus(l.MatchStatus) || matchStatusForCODDiscrepancy(l.DiscrepancyAmount) != NormalizeCODLineMatchStatus(l.MatchStatus) {
		return ErrCODRemittanceInvalidDiscrepancy
	}

	return nil
}

func (d CODDiscrepancy) Validate() error {
	if strings.TrimSpace(d.ID) == "" ||
		strings.TrimSpace(d.LineID) == "" ||
		strings.TrimSpace(d.ReceivableID) == "" ||
		strings.TrimSpace(d.Reason) == "" ||
		strings.TrimSpace(d.OwnerID) == "" ||
		strings.TrimSpace(d.RecordedBy) == "" ||
		d.RecordedAt.IsZero() {
		return ErrCODRemittanceRequiredField
	}
	if !IsValidCODDiscrepancyType(d.Type) || !IsValidCODDiscrepancyStatus(d.Status) {
		return ErrCODRemittanceInvalidDiscrepancy
	}
	amount, err := NewSignedMoneyAmount(d.Amount.String(), decimal.CurrencyVND.String())
	if err != nil || amount.Amount.IsZero() {
		return ErrCODRemittanceInvalidAmount
	}
	if d.Status == CODDiscrepancyStatusResolved &&
		(strings.TrimSpace(d.Resolution) == "" || strings.TrimSpace(d.ResolvedBy) == "" || d.ResolvedAt.IsZero()) {
		return ErrCODRemittanceRequiredField
	}

	return nil
}

func (r CODRemittance) MarkMatching(actorID string, matchedAt time.Time) (CODRemittance, error) {
	if NormalizeCODRemittanceStatus(r.Status) != CODRemittanceStatusDraft {
		return CODRemittance{}, ErrCODRemittanceInvalidTransition
	}
	if !r.DiscrepancyAmount.IsZero() {
		return CODRemittance{}, ErrCODRemittanceInvalidDiscrepancy
	}

	return r.withStatus(CODRemittanceStatusMatching, actorID, matchedAt, "")
}

func (r CODRemittance) RecordDiscrepancy(input RecordCODDiscrepancyInput) (CODRemittance, error) {
	status := NormalizeCODRemittanceStatus(r.Status)
	if status != CODRemittanceStatusDraft && status != CODRemittanceStatusMatching && status != CODRemittanceStatusDiscrepancy {
		return CODRemittance{}, ErrCODRemittanceInvalidTransition
	}
	line, ok := r.lineByID(input.LineID)
	if !ok || line.DiscrepancyAmount.IsZero() {
		return CODRemittance{}, ErrCODRemittanceInvalidDiscrepancy
	}
	recordedAt := input.RecordedAt
	if recordedAt.IsZero() {
		recordedAt = time.Now().UTC()
	}
	recordedBy := strings.TrimSpace(input.RecordedBy)
	if recordedBy == "" {
		recordedBy = strings.TrimSpace(input.OwnerID)
	}
	discrepancyType := NormalizeCODDiscrepancyType(input.Type)
	if strings.TrimSpace(string(discrepancyType)) == "" {
		discrepancyType = defaultCODDiscrepancyType(line.DiscrepancyAmount)
	}
	discrepancyStatus := NormalizeCODDiscrepancyStatus(input.Status)
	if strings.TrimSpace(string(discrepancyStatus)) == "" {
		discrepancyStatus = CODDiscrepancyStatusOpen
	}

	discrepancy := CODDiscrepancy{
		ID:           strings.TrimSpace(input.ID),
		LineID:       line.ID,
		ReceivableID: line.ReceivableID,
		Type:         discrepancyType,
		Status:       discrepancyStatus,
		Amount:       line.DiscrepancyAmount,
		Reason:       strings.TrimSpace(input.Reason),
		OwnerID:      strings.TrimSpace(input.OwnerID),
		RecordedBy:   recordedBy,
		RecordedAt:   recordedAt.UTC(),
	}
	if err := discrepancy.Validate(); err != nil {
		return CODRemittance{}, err
	}

	updated := r.Clone()
	updated.Status = CODRemittanceStatusDiscrepancy
	updated.Discrepancies = upsertCODDiscrepancy(updated.Discrepancies, discrepancy)
	updated.UpdatedBy = recordedBy
	updated.UpdatedAt = recordedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CODRemittance) Submit(actorID string, submittedAt time.Time) (CODRemittance, error) {
	status := NormalizeCODRemittanceStatus(r.Status)
	if status != CODRemittanceStatusMatching && status != CODRemittanceStatusDiscrepancy {
		return CODRemittance{}, ErrCODRemittanceInvalidTransition
	}
	if !r.DiscrepancyAmount.IsZero() && !r.hasTraceForEveryDiscrepantLine() {
		return CODRemittance{}, ErrCODRemittanceInvalidDiscrepancy
	}
	if submittedAt.IsZero() {
		submittedAt = time.Now().UTC()
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return CODRemittance{}, ErrCODRemittanceRequiredField
	}

	updated := r.Clone()
	updated.Status = CODRemittanceStatusSubmitted
	updated.SubmittedBy = actorID
	updated.SubmittedAt = submittedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = submittedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CODRemittance) Approve(actorID string, approvedAt time.Time) (CODRemittance, error) {
	if NormalizeCODRemittanceStatus(r.Status) != CODRemittanceStatusSubmitted {
		return CODRemittance{}, ErrCODRemittanceInvalidTransition
	}
	if approvedAt.IsZero() {
		approvedAt = time.Now().UTC()
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return CODRemittance{}, ErrCODRemittanceRequiredField
	}

	updated := r.Clone()
	updated.Status = CODRemittanceStatusApproved
	updated.ApprovedBy = actorID
	updated.ApprovedAt = approvedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = approvedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CODRemittance) Close(actorID string, closedAt time.Time) (CODRemittance, error) {
	if NormalizeCODRemittanceStatus(r.Status) != CODRemittanceStatusApproved {
		return CODRemittance{}, ErrCODRemittanceInvalidTransition
	}
	if closedAt.IsZero() {
		closedAt = time.Now().UTC()
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return CODRemittance{}, ErrCODRemittanceRequiredField
	}

	updated := r.Clone()
	updated.Status = CODRemittanceStatusClosed
	updated.ClosedBy = actorID
	updated.ClosedAt = closedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = closedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CODRemittance) Void(actorID string, reason string, voidedAt time.Time) (CODRemittance, error) {
	status := NormalizeCODRemittanceStatus(r.Status)
	if status == CODRemittanceStatusClosed || status == CODRemittanceStatusVoid {
		return CODRemittance{}, ErrCODRemittanceInvalidTransition
	}
	if voidedAt.IsZero() {
		voidedAt = time.Now().UTC()
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return CODRemittance{}, ErrCODRemittanceRequiredField
	}

	updated := r.Clone()
	updated.Status = CODRemittanceStatusVoid
	updated.VoidReason = reason
	updated.VoidedBy = actorID
	updated.VoidedAt = voidedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = voidedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (r CODRemittance) Clone() CODRemittance {
	cloned := r
	cloned.Lines = append([]CODRemittanceLine(nil), r.Lines...)
	cloned.Discrepancies = append([]CODDiscrepancy(nil), r.Discrepancies...)
	return cloned
}

func NormalizeCODLineMatchStatus(status CODLineMatchStatus) CODLineMatchStatus {
	return CODLineMatchStatus(normalizeStatus(string(status)))
}

func IsValidCODLineMatchStatus(status CODLineMatchStatus) bool {
	switch NormalizeCODLineMatchStatus(status) {
	case CODLineMatchStatusMatched, CODLineMatchStatusShortPaid, CODLineMatchStatusOverPaid:
		return true
	default:
		return false
	}
}

func NormalizeCODDiscrepancyType(discrepancyType CODDiscrepancyType) CODDiscrepancyType {
	return CODDiscrepancyType(normalizeStatus(string(discrepancyType)))
}

func IsValidCODDiscrepancyType(discrepancyType CODDiscrepancyType) bool {
	switch NormalizeCODDiscrepancyType(discrepancyType) {
	case CODDiscrepancyTypeShortPaid,
		CODDiscrepancyTypeOverPaid,
		CODDiscrepancyTypeCarrierFee,
		CODDiscrepancyTypeReturnClaim,
		CODDiscrepancyTypeOther:
		return true
	default:
		return false
	}
}

func NormalizeCODDiscrepancyStatus(status CODDiscrepancyStatus) CODDiscrepancyStatus {
	return CODDiscrepancyStatus(normalizeStatus(string(status)))
}

func IsValidCODDiscrepancyStatus(status CODDiscrepancyStatus) bool {
	switch NormalizeCODDiscrepancyStatus(status) {
	case CODDiscrepancyStatusOpen, CODDiscrepancyStatusResolved:
		return true
	default:
		return false
	}
}

func (r CODRemittance) withStatus(
	status CODRemittanceStatus,
	actorID string,
	updatedAt time.Time,
	reason string,
) (CODRemittance, error) {
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return CODRemittance{}, ErrCODRemittanceRequiredField
	}

	updated := r.Clone()
	updated.Status = status
	updated.UpdatedBy = actorID
	updated.UpdatedAt = updatedAt.UTC()
	updated.Version++
	if status == CODRemittanceStatusVoid {
		updated.VoidReason = strings.TrimSpace(reason)
		updated.VoidedBy = actorID
		updated.VoidedAt = updatedAt.UTC()
	}

	return updated, updated.Validate()
}

func normalizeCODRemittanceLines(
	inputs []NewCODRemittanceLineInput,
	currencyCode string,
) ([]CODRemittanceLine, decimal.Decimal, decimal.Decimal, error) {
	if len(inputs) == 0 {
		return nil, "", "", ErrCODRemittanceRequiredField
	}
	expectedTotal := decimal.MustMoneyAmount("0")
	remittedTotal := decimal.MustMoneyAmount("0")
	lines := make([]CODRemittanceLine, 0, len(inputs))
	for _, input := range inputs {
		expected, err := NewMoneyAmount(input.ExpectedAmount, currencyCode)
		if err != nil || expected.Amount.IsZero() {
			return nil, "", "", ErrCODRemittanceInvalidAmount
		}
		remitted, err := NewMoneyAmount(input.RemittedAmount, currencyCode)
		if err != nil {
			return nil, "", "", ErrCODRemittanceInvalidAmount
		}
		discrepancyAmount, err := subtractCODMoney(remitted.Amount, expected.Amount)
		if err != nil {
			return nil, "", "", err
		}
		line := CODRemittanceLine{
			ID:                strings.TrimSpace(input.ID),
			ReceivableID:      strings.TrimSpace(input.ReceivableID),
			ReceivableNo:      strings.ToUpper(strings.TrimSpace(input.ReceivableNo)),
			ShipmentID:        strings.TrimSpace(input.ShipmentID),
			TrackingNo:        strings.ToUpper(strings.TrimSpace(input.TrackingNo)),
			CustomerName:      strings.TrimSpace(input.CustomerName),
			ExpectedAmount:    expected.Amount,
			RemittedAmount:    remitted.Amount,
			DiscrepancyAmount: discrepancyAmount,
			MatchStatus:       matchStatusForCODDiscrepancy(discrepancyAmount),
		}
		if err := line.Validate(); err != nil {
			return nil, "", "", err
		}
		expectedTotal, err = addCODMoney(expectedTotal, line.ExpectedAmount)
		if err != nil {
			return nil, "", "", err
		}
		remittedTotal, err = addCODMoney(remittedTotal, line.RemittedAmount)
		if err != nil {
			return nil, "", "", err
		}
		lines = append(lines, line)
	}

	return lines, expectedTotal, remittedTotal, nil
}

func reconcileOptionalCODAmount(input string, calculated decimal.Decimal, currencyCode string) (decimal.Decimal, error) {
	if strings.TrimSpace(input) == "" {
		return calculated, nil
	}
	amount, err := NewMoneyAmount(input, currencyCode)
	if err != nil {
		return "", ErrCODRemittanceInvalidAmount
	}
	if compareMoney(amount.Amount, calculated) != 0 {
		return "", ErrCODRemittanceInvalidAmount
	}

	return amount.Amount, nil
}

func addCODMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	amount, err := addMoney(left, right)
	if err != nil {
		return "", ErrCODRemittanceInvalidAmount
	}

	return amount, nil
}

func subtractCODMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	amount, err := subtractMoney(left, right)
	if err != nil {
		return "", ErrCODRemittanceInvalidAmount
	}

	return amount, nil
}

func matchStatusForCODDiscrepancy(amount decimal.Decimal) CODLineMatchStatus {
	if amount.IsZero() {
		return CODLineMatchStatusMatched
	}
	if amount.IsNegative() {
		return CODLineMatchStatusShortPaid
	}

	return CODLineMatchStatusOverPaid
}

func defaultCODDiscrepancyType(amount decimal.Decimal) CODDiscrepancyType {
	if amount.IsNegative() {
		return CODDiscrepancyTypeShortPaid
	}

	return CODDiscrepancyTypeOverPaid
}

func requiresCODDiscrepancyTrace(status CODRemittanceStatus) bool {
	switch NormalizeCODRemittanceStatus(status) {
	case CODRemittanceStatusDiscrepancy,
		CODRemittanceStatusSubmitted,
		CODRemittanceStatusApproved,
		CODRemittanceStatusClosed:
		return true
	default:
		return false
	}
}

func (r CODRemittance) hasTraceForEveryDiscrepantLine() bool {
	for _, line := range r.Lines {
		if line.DiscrepancyAmount.IsZero() {
			continue
		}
		found := false
		for _, discrepancy := range r.Discrepancies {
			if discrepancy.LineID == line.ID &&
				strings.TrimSpace(discrepancy.Reason) != "" &&
				strings.TrimSpace(discrepancy.OwnerID) != "" {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (r CODRemittance) hasLine(lineID string) bool {
	_, ok := r.lineByID(lineID)
	return ok
}

func (r CODRemittance) lineByID(lineID string) (CODRemittanceLine, bool) {
	lineID = strings.TrimSpace(lineID)
	for _, line := range r.Lines {
		if line.ID == lineID {
			return line, true
		}
	}

	return CODRemittanceLine{}, false
}

func upsertCODDiscrepancy(discrepancies []CODDiscrepancy, discrepancy CODDiscrepancy) []CODDiscrepancy {
	next := make([]CODDiscrepancy, 0, len(discrepancies)+1)
	replaced := false
	for _, current := range discrepancies {
		if current.ID == discrepancy.ID {
			next = append(next, discrepancy)
			replaced = true
			continue
		}
		next = append(next, current)
	}
	if !replaced {
		next = append(next, discrepancy)
	}

	return next
}
