package domain

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrReceivingRequiredField = errors.New("warehouse receiving required field is missing")
var ErrReceivingInvalidStatus = errors.New("warehouse receiving status is invalid")
var ErrReceivingInvalidTransition = errors.New("warehouse receiving status transition is invalid")
var ErrReceivingAlreadyPosted = errors.New("warehouse receiving is already posted")
var ErrReceivingMissingBatchQCData = errors.New("warehouse receiving line is missing batch or qc data")

type WarehouseReceivingStatus string

const (
	WarehouseReceivingStatusDraft        WarehouseReceivingStatus = "draft"
	WarehouseReceivingStatusSubmitted    WarehouseReceivingStatus = "submitted"
	WarehouseReceivingStatusInspectReady WarehouseReceivingStatus = "inspect_ready"
	WarehouseReceivingStatusPosted       WarehouseReceivingStatus = "posted"
)

type WarehouseReceiving struct {
	ID               string
	OrgID            string
	ReceiptNo        string
	WarehouseID      string
	WarehouseCode    string
	LocationID       string
	LocationCode     string
	ReferenceDocType string
	ReferenceDocID   string
	SupplierID       string
	Status           WarehouseReceivingStatus
	Lines            []WarehouseReceivingLine
	StockMovements   []StockMovement
	CreatedBy        string
	SubmittedBy      string
	InspectReadyBy   string
	PostedBy         string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SubmittedAt      time.Time
	InspectReadyAt   time.Time
	PostedAt         time.Time
}

type WarehouseReceivingLine struct {
	ID          string
	ItemID      string
	SKU         string
	ItemName    string
	BatchID     string
	BatchNo     string
	WarehouseID string
	LocationID  string
	Quantity    decimal.Decimal
	BaseUOMCode decimal.UOMCode
	QCStatus    QCStatus
}

type NewWarehouseReceivingInput struct {
	ID               string
	OrgID            string
	ReceiptNo        string
	WarehouseID      string
	WarehouseCode    string
	LocationID       string
	LocationCode     string
	ReferenceDocType string
	ReferenceDocID   string
	SupplierID       string
	Status           WarehouseReceivingStatus
	Lines            []NewWarehouseReceivingLineInput
	CreatedBy        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NewWarehouseReceivingLineInput struct {
	ID          string
	ItemID      string
	SKU         string
	ItemName    string
	BatchID     string
	BatchNo     string
	WarehouseID string
	LocationID  string
	Quantity    decimal.Decimal
	BaseUOMCode string
	QCStatus    QCStatus
}

type WarehouseReceivingFilter struct {
	WarehouseID string
	Status      WarehouseReceivingStatus
}

func NewWarehouseReceiving(input NewWarehouseReceivingInput) (WarehouseReceiving, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeWarehouseReceivingStatus(input.Status)
	if status == "" {
		status = WarehouseReceivingStatusDraft
	}

	receipt := WarehouseReceiving{
		ID:               strings.TrimSpace(input.ID),
		OrgID:            strings.TrimSpace(input.OrgID),
		ReceiptNo:        strings.ToUpper(strings.TrimSpace(input.ReceiptNo)),
		WarehouseID:      strings.TrimSpace(input.WarehouseID),
		WarehouseCode:    strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		LocationID:       strings.TrimSpace(input.LocationID),
		LocationCode:     strings.ToUpper(strings.TrimSpace(input.LocationCode)),
		ReferenceDocType: strings.TrimSpace(input.ReferenceDocType),
		ReferenceDocID:   strings.TrimSpace(input.ReferenceDocID),
		SupplierID:       strings.TrimSpace(input.SupplierID),
		Status:           status,
		Lines:            make([]WarehouseReceivingLine, 0, len(input.Lines)),
		CreatedBy:        strings.TrimSpace(input.CreatedBy),
		CreatedAt:        createdAt.UTC(),
		UpdatedAt:        updatedAt.UTC(),
	}
	for _, lineInput := range input.Lines {
		if strings.TrimSpace(lineInput.WarehouseID) == "" {
			lineInput.WarehouseID = receipt.WarehouseID
		}
		if strings.TrimSpace(lineInput.LocationID) == "" {
			lineInput.LocationID = receipt.LocationID
		}
		line, err := newWarehouseReceivingLine(lineInput)
		if err != nil {
			return WarehouseReceiving{}, err
		}
		receipt.Lines = append(receipt.Lines, line)
	}
	if err := receipt.Validate(); err != nil {
		return WarehouseReceiving{}, err
	}

	return receipt, nil
}

func NewWarehouseReceivingFilter(warehouseID string, status WarehouseReceivingStatus) WarehouseReceivingFilter {
	return WarehouseReceivingFilter{
		WarehouseID: strings.TrimSpace(warehouseID),
		Status:      NormalizeWarehouseReceivingStatus(status),
	}
}

func (f WarehouseReceivingFilter) Matches(receipt WarehouseReceiving) bool {
	if f.WarehouseID != "" && receipt.WarehouseID != f.WarehouseID {
		return false
	}
	if f.Status != "" && receipt.Status != f.Status {
		return false
	}

	return true
}

func (r WarehouseReceiving) Validate() error {
	if strings.TrimSpace(r.ID) == "" ||
		strings.TrimSpace(r.OrgID) == "" ||
		strings.TrimSpace(r.ReceiptNo) == "" ||
		strings.TrimSpace(r.WarehouseID) == "" ||
		strings.TrimSpace(r.LocationID) == "" ||
		strings.TrimSpace(r.ReferenceDocType) == "" ||
		strings.TrimSpace(r.ReferenceDocID) == "" ||
		strings.TrimSpace(r.CreatedBy) == "" ||
		len(r.Lines) == 0 {
		return ErrReceivingRequiredField
	}
	if !IsValidWarehouseReceivingStatus(r.Status) {
		return ErrReceivingInvalidStatus
	}
	for _, line := range r.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (r WarehouseReceiving) ValidatePostable(actorID string) error {
	if strings.TrimSpace(actorID) == "" {
		return ErrReceivingRequiredField
	}
	for _, line := range r.Lines {
		if strings.TrimSpace(line.BatchID) == "" || !IsValidQCStatus(line.QCStatus) {
			return ErrReceivingMissingBatchQCData
		}
	}

	return nil
}

func (r WarehouseReceiving) Submit(actorID string, changedAt time.Time) (WarehouseReceiving, error) {
	if r.Status == WarehouseReceivingStatusPosted {
		return WarehouseReceiving{}, ErrReceivingAlreadyPosted
	}
	if r.Status != WarehouseReceivingStatusDraft {
		return WarehouseReceiving{}, ErrReceivingInvalidTransition
	}
	return r.transitionTo(WarehouseReceivingStatusSubmitted, actorID, changedAt)
}

func (r WarehouseReceiving) MarkInspectReady(actorID string, changedAt time.Time) (WarehouseReceiving, error) {
	if r.Status == WarehouseReceivingStatusPosted {
		return WarehouseReceiving{}, ErrReceivingAlreadyPosted
	}
	if r.Status != WarehouseReceivingStatusSubmitted {
		return WarehouseReceiving{}, ErrReceivingInvalidTransition
	}
	return r.transitionTo(WarehouseReceivingStatusInspectReady, actorID, changedAt)
}

func (r WarehouseReceiving) Post(actorID string, changedAt time.Time) (WarehouseReceiving, error) {
	if r.Status == WarehouseReceivingStatusPosted {
		return WarehouseReceiving{}, ErrReceivingAlreadyPosted
	}
	if r.Status != WarehouseReceivingStatusInspectReady {
		return WarehouseReceiving{}, ErrReceivingInvalidTransition
	}
	if err := r.ValidatePostable(actorID); err != nil {
		return WarehouseReceiving{}, err
	}
	return r.transitionTo(WarehouseReceivingStatusPosted, actorID, changedAt)
}

func (r WarehouseReceiving) AttachStockMovements(movements []StockMovement) WarehouseReceiving {
	updated := r.Clone()
	updated.StockMovements = append([]StockMovement(nil), movements...)

	return updated
}

func (r WarehouseReceiving) Clone() WarehouseReceiving {
	clone := r
	clone.Lines = append([]WarehouseReceivingLine(nil), r.Lines...)
	clone.StockMovements = append([]StockMovement(nil), r.StockMovements...)

	return clone
}

func (r WarehouseReceiving) transitionTo(status WarehouseReceivingStatus, actorID string, changedAt time.Time) (WarehouseReceiving, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return WarehouseReceiving{}, ErrReceivingRequiredField
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := r.Clone()
	updated.Status = status
	updated.UpdatedAt = changedAt.UTC()
	switch status {
	case WarehouseReceivingStatusSubmitted:
		updated.SubmittedBy = actorID
		updated.SubmittedAt = changedAt.UTC()
	case WarehouseReceivingStatusInspectReady:
		updated.InspectReadyBy = actorID
		updated.InspectReadyAt = changedAt.UTC()
	case WarehouseReceivingStatusPosted:
		updated.PostedBy = actorID
		updated.PostedAt = changedAt.UTC()
	}
	if err := updated.Validate(); err != nil {
		return WarehouseReceiving{}, err
	}

	return updated, nil
}

func (l WarehouseReceivingLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKU) == "" ||
		strings.TrimSpace(l.WarehouseID) == "" ||
		strings.TrimSpace(l.LocationID) == "" {
		return ErrReceivingRequiredField
	}
	quantity, err := decimal.ParseQuantity(l.Quantity.String())
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		return ErrReceivingRequiredField
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrReceivingRequiredField
	}
	if l.QCStatus != "" && !IsValidQCStatus(l.QCStatus) {
		return ErrBatchInvalidQCStatus
	}

	return nil
}

func newWarehouseReceivingLine(input NewWarehouseReceivingLineInput) (WarehouseReceivingLine, error) {
	quantity, err := decimal.ParseQuantity(input.Quantity.String())
	if err != nil {
		return WarehouseReceivingLine{}, err
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return WarehouseReceivingLine{}, err
	}
	line := WarehouseReceivingLine{
		ID:          strings.TrimSpace(input.ID),
		ItemID:      strings.TrimSpace(input.ItemID),
		SKU:         strings.ToUpper(strings.TrimSpace(input.SKU)),
		ItemName:    strings.TrimSpace(input.ItemName),
		BatchID:     strings.TrimSpace(input.BatchID),
		BatchNo:     NormalizeBatchNo(input.BatchNo),
		WarehouseID: strings.TrimSpace(input.WarehouseID),
		LocationID:  strings.TrimSpace(input.LocationID),
		Quantity:    quantity,
		BaseUOMCode: baseUOMCode,
		QCStatus:    NormalizeQCStatus(input.QCStatus),
	}
	if err := line.Validate(); err != nil {
		return WarehouseReceivingLine{}, err
	}

	return line, nil
}

func NormalizeWarehouseReceivingStatus(value WarehouseReceivingStatus) WarehouseReceivingStatus {
	return WarehouseReceivingStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func IsValidWarehouseReceivingStatus(value WarehouseReceivingStatus) bool {
	switch NormalizeWarehouseReceivingStatus(value) {
	case WarehouseReceivingStatusDraft,
		WarehouseReceivingStatusSubmitted,
		WarehouseReceivingStatusInspectReady,
		WarehouseReceivingStatusPosted:
		return true
	default:
		return false
	}
}

func SortWarehouseReceivings(receipts []WarehouseReceiving) {
	sort.Slice(receipts, func(i int, j int) bool {
		left := receipts[i]
		right := receipts[j]
		if !left.UpdatedAt.Equal(right.UpdatedAt) {
			return left.UpdatedAt.After(right.UpdatedAt)
		}

		return left.ReceiptNo < right.ReceiptNo
	})
}
