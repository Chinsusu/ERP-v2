package domain

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSupplierRejectionRequiredField = errors.New("supplier rejection required field is missing")
var ErrSupplierRejectionInvalidStatus = errors.New("supplier rejection status is invalid")
var ErrSupplierRejectionInvalidTransition = errors.New("supplier rejection status transition is invalid")
var ErrSupplierRejectionInvalidQuantity = errors.New("supplier rejection quantity is invalid")

type SupplierRejectionStatus string

const (
	SupplierRejectionStatusDraft     SupplierRejectionStatus = "draft"
	SupplierRejectionStatusSubmitted SupplierRejectionStatus = "submitted"
	SupplierRejectionStatusConfirmed SupplierRejectionStatus = "confirmed"
	SupplierRejectionStatusCancelled SupplierRejectionStatus = "cancelled"
)

type SupplierRejection struct {
	ID                    string
	OrgID                 string
	RejectionNo           string
	SupplierID            string
	SupplierCode          string
	SupplierName          string
	PurchaseOrderID       string
	PurchaseOrderNo       string
	GoodsReceiptID        string
	GoodsReceiptNo        string
	InboundQCInspectionID string
	WarehouseID           string
	WarehouseCode         string
	Status                SupplierRejectionStatus
	Reason                string
	Lines                 []SupplierRejectionLine
	Attachments           []SupplierRejectionAttachment
	CreatedAt             time.Time
	CreatedBy             string
	UpdatedAt             time.Time
	UpdatedBy             string
	SubmittedAt           time.Time
	SubmittedBy           string
	ConfirmedAt           time.Time
	ConfirmedBy           string
	CancelledAt           time.Time
	CancelledBy           string
	CancelReason          string
}

type SupplierRejectionLine struct {
	ID                    string
	PurchaseOrderLineID   string
	GoodsReceiptLineID    string
	InboundQCInspectionID string
	ItemID                string
	SKU                   string
	ItemName              string
	BatchID               string
	BatchNo               string
	LotNo                 string
	ExpiryDate            time.Time
	RejectedQuantity      decimal.Decimal
	UOMCode               decimal.UOMCode
	BaseUOMCode           decimal.UOMCode
	Reason                string
}

type SupplierRejectionAttachment struct {
	ID          string
	LineID      string
	FileName    string
	ObjectKey   string
	ContentType string
	UploadedAt  time.Time
	UploadedBy  string
	Source      string
}

type SupplierRejectionFilter struct {
	SupplierID  string
	WarehouseID string
	Status      SupplierRejectionStatus
}

type NewSupplierRejectionInput struct {
	ID                    string
	OrgID                 string
	RejectionNo           string
	SupplierID            string
	SupplierCode          string
	SupplierName          string
	PurchaseOrderID       string
	PurchaseOrderNo       string
	GoodsReceiptID        string
	GoodsReceiptNo        string
	InboundQCInspectionID string
	WarehouseID           string
	WarehouseCode         string
	Reason                string
	Lines                 []NewSupplierRejectionLineInput
	Attachments           []NewSupplierRejectionAttachmentInput
	CreatedAt             time.Time
	CreatedBy             string
	UpdatedAt             time.Time
	UpdatedBy             string
}

type NewSupplierRejectionLineInput struct {
	ID                    string
	PurchaseOrderLineID   string
	GoodsReceiptLineID    string
	InboundQCInspectionID string
	ItemID                string
	SKU                   string
	ItemName              string
	BatchID               string
	BatchNo               string
	LotNo                 string
	ExpiryDate            time.Time
	RejectedQuantity      decimal.Decimal
	UOMCode               string
	BaseUOMCode           string
	Reason                string
}

type NewSupplierRejectionAttachmentInput struct {
	ID          string
	LineID      string
	FileName    string
	ObjectKey   string
	ContentType string
	UploadedAt  time.Time
	UploadedBy  string
	Source      string
}

var supplierRejectionTransitions = map[SupplierRejectionStatus][]SupplierRejectionStatus{
	SupplierRejectionStatusDraft: {
		SupplierRejectionStatusSubmitted,
		SupplierRejectionStatusCancelled,
	},
	SupplierRejectionStatusSubmitted: {
		SupplierRejectionStatusConfirmed,
		SupplierRejectionStatusCancelled,
	},
}

func NewSupplierRejection(input NewSupplierRejectionInput) (SupplierRejection, error) {
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

	rejection := SupplierRejection{
		ID:                    strings.TrimSpace(input.ID),
		OrgID:                 strings.TrimSpace(input.OrgID),
		RejectionNo:           strings.ToUpper(strings.TrimSpace(input.RejectionNo)),
		SupplierID:            strings.TrimSpace(input.SupplierID),
		SupplierCode:          strings.ToUpper(strings.TrimSpace(input.SupplierCode)),
		SupplierName:          strings.TrimSpace(input.SupplierName),
		PurchaseOrderID:       strings.TrimSpace(input.PurchaseOrderID),
		PurchaseOrderNo:       strings.ToUpper(strings.TrimSpace(input.PurchaseOrderNo)),
		GoodsReceiptID:        strings.TrimSpace(input.GoodsReceiptID),
		GoodsReceiptNo:        strings.ToUpper(strings.TrimSpace(input.GoodsReceiptNo)),
		InboundQCInspectionID: strings.TrimSpace(input.InboundQCInspectionID),
		WarehouseID:           strings.TrimSpace(input.WarehouseID),
		WarehouseCode:         strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		Status:                SupplierRejectionStatusDraft,
		Reason:                strings.TrimSpace(input.Reason),
		Lines:                 make([]SupplierRejectionLine, 0, len(input.Lines)),
		Attachments:           make([]SupplierRejectionAttachment, 0, len(input.Attachments)),
		CreatedAt:             createdAt.UTC(),
		CreatedBy:             strings.TrimSpace(input.CreatedBy),
		UpdatedAt:             updatedAt.UTC(),
		UpdatedBy:             updatedBy,
	}
	for _, lineInput := range input.Lines {
		line, err := NewSupplierRejectionLine(lineInput)
		if err != nil {
			return SupplierRejection{}, err
		}
		rejection.Lines = append(rejection.Lines, line)
	}
	for _, attachmentInput := range input.Attachments {
		attachment, err := NewSupplierRejectionAttachment(attachmentInput, createdAt, rejection.CreatedBy)
		if err != nil {
			return SupplierRejection{}, err
		}
		rejection.Attachments = append(rejection.Attachments, attachment)
	}
	if err := rejection.Validate(); err != nil {
		return SupplierRejection{}, err
	}

	return rejection, nil
}

func NewSupplierRejectionLine(input NewSupplierRejectionLineInput) (SupplierRejectionLine, error) {
	quantity, err := decimal.ParseQuantity(input.RejectedQuantity.String())
	if err != nil || quantity.IsZero() || quantity.IsNegative() {
		return SupplierRejectionLine{}, ErrSupplierRejectionInvalidQuantity
	}
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return SupplierRejectionLine{}, ErrSupplierRejectionInvalidQuantity
	}
	baseUOMCode := uomCode
	if strings.TrimSpace(input.BaseUOMCode) != "" {
		baseUOMCode, err = decimal.NormalizeUOMCode(input.BaseUOMCode)
		if err != nil {
			return SupplierRejectionLine{}, ErrSupplierRejectionInvalidQuantity
		}
	}

	line := SupplierRejectionLine{
		ID:                    strings.TrimSpace(input.ID),
		PurchaseOrderLineID:   strings.TrimSpace(input.PurchaseOrderLineID),
		GoodsReceiptLineID:    strings.TrimSpace(input.GoodsReceiptLineID),
		InboundQCInspectionID: strings.TrimSpace(input.InboundQCInspectionID),
		ItemID:                strings.TrimSpace(input.ItemID),
		SKU:                   strings.ToUpper(strings.TrimSpace(input.SKU)),
		ItemName:              strings.TrimSpace(input.ItemName),
		BatchID:               strings.TrimSpace(input.BatchID),
		BatchNo:               NormalizeBatchNo(input.BatchNo),
		LotNo:                 NormalizeBatchNo(input.LotNo),
		ExpiryDate:            dateOnly(input.ExpiryDate),
		RejectedQuantity:      quantity,
		UOMCode:               uomCode,
		BaseUOMCode:           baseUOMCode,
		Reason:                strings.TrimSpace(input.Reason),
	}
	if line.LotNo == "" {
		line.LotNo = line.BatchNo
	}
	if line.BatchNo == "" {
		line.BatchNo = line.LotNo
	}
	if err := line.Validate(); err != nil {
		return SupplierRejectionLine{}, err
	}

	return line, nil
}

func NewSupplierRejectionAttachment(
	input NewSupplierRejectionAttachmentInput,
	defaultUploadedAt time.Time,
	defaultUploadedBy string,
) (SupplierRejectionAttachment, error) {
	uploadedAt := input.UploadedAt
	if uploadedAt.IsZero() {
		uploadedAt = defaultUploadedAt
	}
	uploadedBy := strings.TrimSpace(input.UploadedBy)
	if uploadedBy == "" {
		uploadedBy = strings.TrimSpace(defaultUploadedBy)
	}

	attachment := SupplierRejectionAttachment{
		ID:          strings.TrimSpace(input.ID),
		LineID:      strings.TrimSpace(input.LineID),
		FileName:    strings.TrimSpace(input.FileName),
		ObjectKey:   strings.TrimSpace(input.ObjectKey),
		ContentType: strings.TrimSpace(input.ContentType),
		UploadedAt:  uploadedAt.UTC(),
		UploadedBy:  uploadedBy,
		Source:      strings.TrimSpace(input.Source),
	}
	if err := attachment.Validate(); err != nil {
		return SupplierRejectionAttachment{}, err
	}

	return attachment, nil
}

func NewSupplierRejectionFilter(
	supplierID string,
	warehouseID string,
	status SupplierRejectionStatus,
) SupplierRejectionFilter {
	return SupplierRejectionFilter{
		SupplierID:  strings.TrimSpace(supplierID),
		WarehouseID: strings.TrimSpace(warehouseID),
		Status:      NormalizeSupplierRejectionStatus(status),
	}
}

func (f SupplierRejectionFilter) Matches(rejection SupplierRejection) bool {
	if f.SupplierID != "" && f.SupplierID != rejection.SupplierID {
		return false
	}
	if f.WarehouseID != "" && f.WarehouseID != rejection.WarehouseID {
		return false
	}
	if f.Status != "" && f.Status != rejection.Status {
		return false
	}

	return true
}

func (r SupplierRejection) Submit(actorID string, changedAt time.Time) (SupplierRejection, error) {
	return r.transition(SupplierRejectionStatusSubmitted, actorID, "", changedAt)
}

func (r SupplierRejection) Confirm(actorID string, changedAt time.Time) (SupplierRejection, error) {
	return r.transition(SupplierRejectionStatusConfirmed, actorID, "", changedAt)
}

func (r SupplierRejection) Cancel(actorID string, reason string, changedAt time.Time) (SupplierRejection, error) {
	return r.transition(SupplierRejectionStatusCancelled, actorID, reason, changedAt)
}

func (r SupplierRejection) transition(
	next SupplierRejectionStatus,
	actorID string,
	reason string,
	changedAt time.Time,
) (SupplierRejection, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SupplierRejection{}, ErrSupplierRejectionRequiredField
	}
	next = NormalizeSupplierRejectionStatus(next)
	if !CanTransitionSupplierRejectionStatus(r.Status, next) {
		return SupplierRejection{}, ErrSupplierRejectionInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := r.Clone()
	updated.Status = next
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	switch next {
	case SupplierRejectionStatusSubmitted:
		updated.SubmittedAt = changedAt.UTC()
		updated.SubmittedBy = actorID
	case SupplierRejectionStatusConfirmed:
		updated.ConfirmedAt = changedAt.UTC()
		updated.ConfirmedBy = actorID
	case SupplierRejectionStatusCancelled:
		reason = strings.TrimSpace(reason)
		if reason == "" {
			return SupplierRejection{}, ErrSupplierRejectionRequiredField
		}
		updated.CancelledAt = changedAt.UTC()
		updated.CancelledBy = actorID
		updated.CancelReason = reason
	}
	if err := updated.Validate(); err != nil {
		return SupplierRejection{}, err
	}

	return updated, nil
}

func (r SupplierRejection) Validate() error {
	if strings.TrimSpace(r.ID) == "" ||
		strings.TrimSpace(r.OrgID) == "" ||
		strings.TrimSpace(r.RejectionNo) == "" ||
		strings.TrimSpace(r.SupplierID) == "" ||
		strings.TrimSpace(r.GoodsReceiptID) == "" ||
		strings.TrimSpace(r.InboundQCInspectionID) == "" ||
		strings.TrimSpace(r.WarehouseID) == "" ||
		strings.TrimSpace(r.Reason) == "" ||
		strings.TrimSpace(r.CreatedBy) == "" ||
		len(r.Lines) == 0 {
		return ErrSupplierRejectionRequiredField
	}
	if !IsValidSupplierRejectionStatus(r.Status) {
		return ErrSupplierRejectionInvalidStatus
	}
	for _, line := range r.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}
	for _, attachment := range r.Attachments {
		if err := attachment.Validate(); err != nil {
			return err
		}
	}
	if r.Status == SupplierRejectionStatusSubmitted &&
		(strings.TrimSpace(r.SubmittedBy) == "" || r.SubmittedAt.IsZero()) {
		return ErrSupplierRejectionRequiredField
	}
	if r.Status == SupplierRejectionStatusConfirmed &&
		(strings.TrimSpace(r.ConfirmedBy) == "" || r.ConfirmedAt.IsZero()) {
		return ErrSupplierRejectionRequiredField
	}
	if r.Status == SupplierRejectionStatusCancelled &&
		(strings.TrimSpace(r.CancelledBy) == "" || r.CancelledAt.IsZero() || strings.TrimSpace(r.CancelReason) == "") {
		return ErrSupplierRejectionRequiredField
	}

	return nil
}

func (l SupplierRejectionLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		strings.TrimSpace(l.GoodsReceiptLineID) == "" ||
		strings.TrimSpace(l.InboundQCInspectionID) == "" ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKU) == "" ||
		strings.TrimSpace(l.BatchID) == "" ||
		strings.TrimSpace(l.BatchNo) == "" ||
		strings.TrimSpace(l.LotNo) == "" ||
		strings.TrimSpace(l.Reason) == "" {
		return ErrSupplierRejectionRequiredField
	}
	if l.ExpiryDate.IsZero() {
		return ErrSupplierRejectionRequiredField
	}
	quantity, err := decimal.ParseQuantity(l.RejectedQuantity.String())
	if err != nil || quantity.IsZero() || quantity.IsNegative() {
		return ErrSupplierRejectionInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrSupplierRejectionInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrSupplierRejectionInvalidQuantity
	}

	return nil
}

func (a SupplierRejectionAttachment) Validate() error {
	if strings.TrimSpace(a.ID) == "" ||
		strings.TrimSpace(a.FileName) == "" ||
		strings.TrimSpace(a.ObjectKey) == "" ||
		strings.TrimSpace(a.UploadedBy) == "" ||
		a.UploadedAt.IsZero() {
		return ErrSupplierRejectionRequiredField
	}

	return nil
}

func (r SupplierRejection) Clone() SupplierRejection {
	clone := r
	clone.Lines = append([]SupplierRejectionLine(nil), r.Lines...)
	clone.Attachments = append([]SupplierRejectionAttachment(nil), r.Attachments...)

	return clone
}

func NormalizeSupplierRejectionStatus(status SupplierRejectionStatus) SupplierRejectionStatus {
	switch SupplierRejectionStatus(strings.ToLower(strings.TrimSpace(string(status)))) {
	case SupplierRejectionStatusDraft:
		return SupplierRejectionStatusDraft
	case SupplierRejectionStatusSubmitted:
		return SupplierRejectionStatusSubmitted
	case SupplierRejectionStatusConfirmed:
		return SupplierRejectionStatusConfirmed
	case SupplierRejectionStatusCancelled:
		return SupplierRejectionStatusCancelled
	default:
		return ""
	}
}

func IsValidSupplierRejectionStatus(status SupplierRejectionStatus) bool {
	return NormalizeSupplierRejectionStatus(status) != ""
}

func CanTransitionSupplierRejectionStatus(current SupplierRejectionStatus, next SupplierRejectionStatus) bool {
	current = NormalizeSupplierRejectionStatus(current)
	next = NormalizeSupplierRejectionStatus(next)
	if current == next {
		return true
	}
	for _, allowed := range supplierRejectionTransitions[current] {
		if allowed == next {
			return true
		}
	}

	return false
}

func SortSupplierRejections(rows []SupplierRejection) {
	sort.SliceStable(rows, func(i int, j int) bool {
		left := rows[i]
		right := rows[j]
		if !left.CreatedAt.Equal(right.CreatedAt) {
			return left.CreatedAt.After(right.CreatedAt)
		}

		return left.RejectionNo > right.RejectionNo
	})
}
