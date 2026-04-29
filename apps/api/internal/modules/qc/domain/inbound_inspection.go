package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrInboundQCInspectionRequiredField = errors.New("inbound qc inspection required field is missing")
var ErrInboundQCInspectionInvalidStatus = errors.New("inbound qc inspection status is invalid")
var ErrInboundQCInspectionInvalidResult = errors.New("inbound qc inspection result is invalid")
var ErrInboundQCInspectionInvalidQuantity = errors.New("inbound qc inspection quantity is invalid")
var ErrInboundQCInspectionInvalidTransition = errors.New("inbound qc inspection status transition is invalid")
var ErrInboundQCChecklistInvalidStatus = errors.New("inbound qc checklist status is invalid")
var ErrInboundQCChecklistIncomplete = errors.New("inbound qc checklist is incomplete")

type InboundQCInspectionStatus string
type InboundQCResult string
type InboundQCChecklistStatus string

const (
	InboundQCInspectionStatusPending    InboundQCInspectionStatus = "pending"
	InboundQCInspectionStatusInProgress InboundQCInspectionStatus = "in_progress"
	InboundQCInspectionStatusCompleted  InboundQCInspectionStatus = "completed"
	InboundQCInspectionStatusCancelled  InboundQCInspectionStatus = "cancelled"

	InboundQCResultPass    InboundQCResult = "pass"
	InboundQCResultFail    InboundQCResult = "fail"
	InboundQCResultHold    InboundQCResult = "hold"
	InboundQCResultPartial InboundQCResult = "partial"

	InboundQCChecklistStatusPending       InboundQCChecklistStatus = "pending"
	InboundQCChecklistStatusPass          InboundQCChecklistStatus = "pass"
	InboundQCChecklistStatusFail          InboundQCChecklistStatus = "fail"
	InboundQCChecklistStatusNotApplicable InboundQCChecklistStatus = "not_applicable"
)

type InboundQCInspection struct {
	ID                  string
	OrgID               string
	GoodsReceiptID      string
	GoodsReceiptNo      string
	GoodsReceiptLineID  string
	PurchaseOrderID     string
	PurchaseOrderLineID string
	ItemID              string
	SKU                 string
	ItemName            string
	BatchID             string
	BatchNo             string
	LotNo               string
	ExpiryDate          time.Time
	WarehouseID         string
	LocationID          string
	Quantity            decimal.Decimal
	UOMCode             decimal.UOMCode
	InspectorID         string
	Status              InboundQCInspectionStatus
	Result              InboundQCResult
	PassedQuantity      decimal.Decimal
	FailedQuantity      decimal.Decimal
	HoldQuantity        decimal.Decimal
	Checklist           []InboundQCChecklistItem
	Reason              string
	Note                string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	StartedAt           time.Time
	StartedBy           string
	DecidedAt           time.Time
	DecidedBy           string
}

type InboundQCChecklistItem struct {
	ID       string
	Code     string
	Label    string
	Required bool
	Status   InboundQCChecklistStatus
	Note     string
}

type NewInboundQCInspectionInput struct {
	ID                  string
	OrgID               string
	GoodsReceiptID      string
	GoodsReceiptNo      string
	GoodsReceiptLineID  string
	PurchaseOrderID     string
	PurchaseOrderLineID string
	ItemID              string
	SKU                 string
	ItemName            string
	BatchID             string
	BatchNo             string
	LotNo               string
	ExpiryDate          time.Time
	WarehouseID         string
	LocationID          string
	Quantity            decimal.Decimal
	UOMCode             string
	InspectorID         string
	Checklist           []NewInboundQCChecklistItemInput
	Note                string
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
}

type NewInboundQCChecklistItemInput struct {
	ID       string
	Code     string
	Label    string
	Required bool
	Status   InboundQCChecklistStatus
	Note     string
}

type InboundQCDecisionInput struct {
	Result         InboundQCResult
	PassedQuantity decimal.Decimal
	FailedQuantity decimal.Decimal
	HoldQuantity   decimal.Decimal
	Checklist      []NewInboundQCChecklistItemInput
	Reason         string
	Note           string
	ActorID        string
	ChangedAt      time.Time
}

func NewInboundQCInspection(input NewInboundQCInspectionInput) (InboundQCInspection, error) {
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
	quantity, err := normalizeInboundQCQuantity(input.Quantity)
	if err != nil || quantity.IsZero() {
		return InboundQCInspection{}, ErrInboundQCInspectionInvalidQuantity
	}
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return InboundQCInspection{}, ErrInboundQCInspectionRequiredField
	}
	checklist, err := newInboundQCChecklist(input.Checklist)
	if err != nil {
		return InboundQCInspection{}, err
	}

	inspection := InboundQCInspection{
		ID:                  strings.TrimSpace(input.ID),
		OrgID:               strings.TrimSpace(input.OrgID),
		GoodsReceiptID:      strings.TrimSpace(input.GoodsReceiptID),
		GoodsReceiptNo:      strings.ToUpper(strings.TrimSpace(input.GoodsReceiptNo)),
		GoodsReceiptLineID:  strings.TrimSpace(input.GoodsReceiptLineID),
		PurchaseOrderID:     strings.TrimSpace(input.PurchaseOrderID),
		PurchaseOrderLineID: strings.TrimSpace(input.PurchaseOrderLineID),
		ItemID:              strings.TrimSpace(input.ItemID),
		SKU:                 strings.ToUpper(strings.TrimSpace(input.SKU)),
		ItemName:            strings.TrimSpace(input.ItemName),
		BatchID:             strings.TrimSpace(input.BatchID),
		BatchNo:             normalizeInboundQCLot(input.BatchNo),
		LotNo:               normalizeInboundQCLot(input.LotNo),
		ExpiryDate:          dateOnly(input.ExpiryDate),
		WarehouseID:         strings.TrimSpace(input.WarehouseID),
		LocationID:          strings.TrimSpace(input.LocationID),
		Quantity:            quantity,
		UOMCode:             uomCode,
		InspectorID:         strings.TrimSpace(input.InspectorID),
		Status:              InboundQCInspectionStatusPending,
		Checklist:           checklist,
		Note:                strings.TrimSpace(input.Note),
		CreatedAt:           createdAt.UTC(),
		CreatedBy:           strings.TrimSpace(input.CreatedBy),
		UpdatedAt:           updatedAt.UTC(),
		UpdatedBy:           updatedBy,
	}
	if inspection.LotNo == "" {
		inspection.LotNo = inspection.BatchNo
	}
	if err := inspection.Validate(); err != nil {
		return InboundQCInspection{}, err
	}

	return inspection, nil
}

func (i InboundQCInspection) Start(actorID string, changedAt time.Time) (InboundQCInspection, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return InboundQCInspection{}, ErrInboundQCInspectionRequiredField
	}
	if i.Status != InboundQCInspectionStatusPending {
		return InboundQCInspection{}, ErrInboundQCInspectionInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := i.Clone()
	updated.Status = InboundQCInspectionStatusInProgress
	updated.StartedAt = changedAt.UTC()
	updated.StartedBy = actorID
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if strings.TrimSpace(updated.InspectorID) == "" {
		updated.InspectorID = actorID
	}
	if err := updated.Validate(); err != nil {
		return InboundQCInspection{}, err
	}

	return updated, nil
}

func (i InboundQCInspection) RecordDecision(input InboundQCDecisionInput) (InboundQCInspection, error) {
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return InboundQCInspection{}, ErrInboundQCInspectionRequiredField
	}
	if i.Status != InboundQCInspectionStatusInProgress {
		return InboundQCInspection{}, ErrInboundQCInspectionInvalidTransition
	}
	result := NormalizeInboundQCResult(input.Result)
	if !IsValidInboundQCResult(result) {
		return InboundQCInspection{}, ErrInboundQCInspectionInvalidResult
	}
	passedQty, failedQty, holdQty, err := normalizeDecisionQuantities(input)
	if err != nil {
		return InboundQCInspection{}, err
	}
	if err := validateDecisionQuantities(i.Quantity, result, passedQty, failedQty, holdQty); err != nil {
		return InboundQCInspection{}, err
	}
	checklist := i.Checklist
	if input.Checklist != nil {
		checklist, err = newInboundQCChecklist(input.Checklist)
		if err != nil {
			return InboundQCInspection{}, err
		}
	}
	if hasIncompleteRequiredChecklist(checklist) {
		return InboundQCInspection{}, ErrInboundQCChecklistIncomplete
	}
	reason := strings.TrimSpace(input.Reason)
	if requiresDecisionReason(result) && reason == "" {
		return InboundQCInspection{}, ErrInboundQCInspectionRequiredField
	}
	changedAt := input.ChangedAt
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := i.Clone()
	updated.Status = InboundQCInspectionStatusCompleted
	updated.Result = result
	updated.PassedQuantity = passedQty
	updated.FailedQuantity = failedQty
	updated.HoldQuantity = holdQty
	updated.Checklist = checklist
	updated.Reason = reason
	updated.Note = strings.TrimSpace(input.Note)
	updated.DecidedAt = changedAt.UTC()
	updated.DecidedBy = actorID
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if err := updated.Validate(); err != nil {
		return InboundQCInspection{}, err
	}

	return updated, nil
}

func (i InboundQCInspection) Cancel(actorID string, reason string, changedAt time.Time) (InboundQCInspection, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" || strings.TrimSpace(reason) == "" {
		return InboundQCInspection{}, ErrInboundQCInspectionRequiredField
	}
	if i.Status == InboundQCInspectionStatusCompleted || i.Status == InboundQCInspectionStatusCancelled {
		return InboundQCInspection{}, ErrInboundQCInspectionInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := i.Clone()
	updated.Status = InboundQCInspectionStatusCancelled
	updated.Reason = strings.TrimSpace(reason)
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if err := updated.Validate(); err != nil {
		return InboundQCInspection{}, err
	}

	return updated, nil
}

func (i InboundQCInspection) Validate() error {
	if strings.TrimSpace(i.ID) == "" ||
		strings.TrimSpace(i.OrgID) == "" ||
		strings.TrimSpace(i.GoodsReceiptID) == "" ||
		strings.TrimSpace(i.GoodsReceiptLineID) == "" ||
		strings.TrimSpace(i.ItemID) == "" ||
		strings.TrimSpace(i.SKU) == "" ||
		strings.TrimSpace(i.BatchID) == "" ||
		strings.TrimSpace(i.BatchNo) == "" ||
		strings.TrimSpace(i.LotNo) == "" ||
		strings.TrimSpace(i.WarehouseID) == "" ||
		strings.TrimSpace(i.LocationID) == "" ||
		strings.TrimSpace(i.InspectorID) == "" ||
		strings.TrimSpace(i.CreatedBy) == "" ||
		len(i.Checklist) == 0 {
		return ErrInboundQCInspectionRequiredField
	}
	if !IsValidInboundQCInspectionStatus(i.Status) {
		return ErrInboundQCInspectionInvalidStatus
	}
	quantity, err := normalizeInboundQCQuantity(i.Quantity)
	if err != nil || quantity.IsZero() {
		return ErrInboundQCInspectionInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(i.UOMCode.String()); err != nil {
		return ErrInboundQCInspectionRequiredField
	}
	if i.ExpiryDate.IsZero() {
		return ErrInboundQCInspectionRequiredField
	}
	if i.Result != "" && !IsValidInboundQCResult(i.Result) {
		return ErrInboundQCInspectionInvalidResult
	}
	for _, item := range i.Checklist {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	if i.Status == InboundQCInspectionStatusInProgress &&
		(strings.TrimSpace(i.StartedBy) == "" || i.StartedAt.IsZero()) {
		return ErrInboundQCInspectionRequiredField
	}
	if i.Status == InboundQCInspectionStatusCompleted {
		if strings.TrimSpace(i.DecidedBy) == "" || i.DecidedAt.IsZero() || !IsValidInboundQCResult(i.Result) {
			return ErrInboundQCInspectionRequiredField
		}
		passedQty, failedQty, holdQty, err := normalizeDecisionQuantities(InboundQCDecisionInput{
			PassedQuantity: i.PassedQuantity,
			FailedQuantity: i.FailedQuantity,
			HoldQuantity:   i.HoldQuantity,
		})
		if err != nil {
			return err
		}
		if err := validateDecisionQuantities(i.Quantity, i.Result, passedQty, failedQty, holdQty); err != nil {
			return err
		}
		if requiresDecisionReason(i.Result) && strings.TrimSpace(i.Reason) == "" {
			return ErrInboundQCInspectionRequiredField
		}
		if hasIncompleteRequiredChecklist(i.Checklist) {
			return ErrInboundQCChecklistIncomplete
		}
	}
	if i.Status == InboundQCInspectionStatusCancelled && strings.TrimSpace(i.Reason) == "" {
		return ErrInboundQCInspectionRequiredField
	}

	return nil
}

func (i InboundQCInspection) Clone() InboundQCInspection {
	clone := i
	clone.Checklist = append([]InboundQCChecklistItem(nil), i.Checklist...)

	return clone
}

func (item InboundQCChecklistItem) Validate() error {
	if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.Code) == "" || strings.TrimSpace(item.Label) == "" {
		return ErrInboundQCInspectionRequiredField
	}
	if !IsValidInboundQCChecklistStatus(item.Status) {
		return ErrInboundQCChecklistInvalidStatus
	}

	return nil
}

func NormalizeInboundQCInspectionStatus(value InboundQCInspectionStatus) InboundQCInspectionStatus {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	normalized = strings.ReplaceAll(normalized, " ", "_")

	return InboundQCInspectionStatus(normalized)
}

func NormalizeInboundQCResult(value InboundQCResult) InboundQCResult {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	normalized = strings.ReplaceAll(normalized, " ", "_")

	return InboundQCResult(normalized)
}

func NormalizeInboundQCChecklistStatus(value InboundQCChecklistStatus) InboundQCChecklistStatus {
	normalized := strings.ToLower(strings.TrimSpace(string(value)))
	normalized = strings.ReplaceAll(normalized, " ", "_")

	return InboundQCChecklistStatus(normalized)
}

func IsValidInboundQCInspectionStatus(value InboundQCInspectionStatus) bool {
	switch NormalizeInboundQCInspectionStatus(value) {
	case InboundQCInspectionStatusPending,
		InboundQCInspectionStatusInProgress,
		InboundQCInspectionStatusCompleted,
		InboundQCInspectionStatusCancelled:
		return true
	default:
		return false
	}
}

func IsValidInboundQCResult(value InboundQCResult) bool {
	switch NormalizeInboundQCResult(value) {
	case InboundQCResultPass, InboundQCResultFail, InboundQCResultHold, InboundQCResultPartial:
		return true
	default:
		return false
	}
}

func IsValidInboundQCChecklistStatus(value InboundQCChecklistStatus) bool {
	switch NormalizeInboundQCChecklistStatus(value) {
	case InboundQCChecklistStatusPending,
		InboundQCChecklistStatusPass,
		InboundQCChecklistStatusFail,
		InboundQCChecklistStatusNotApplicable:
		return true
	default:
		return false
	}
}

func newInboundQCChecklist(inputs []NewInboundQCChecklistItemInput) ([]InboundQCChecklistItem, error) {
	if len(inputs) == 0 {
		return nil, ErrInboundQCInspectionRequiredField
	}
	items := make([]InboundQCChecklistItem, 0, len(inputs))
	for _, input := range inputs {
		status := NormalizeInboundQCChecklistStatus(input.Status)
		if status == "" {
			status = InboundQCChecklistStatusPending
		}
		item := InboundQCChecklistItem{
			ID:       strings.TrimSpace(input.ID),
			Code:     strings.ToUpper(strings.TrimSpace(input.Code)),
			Label:    strings.TrimSpace(input.Label),
			Required: input.Required,
			Status:   status,
			Note:     strings.TrimSpace(input.Note),
		}
		if err := item.Validate(); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func normalizeDecisionQuantities(input InboundQCDecisionInput) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	passedQty, err := normalizeInboundQCQuantityAllowZero(input.PassedQuantity)
	if err != nil {
		return "", "", "", err
	}
	failedQty, err := normalizeInboundQCQuantityAllowZero(input.FailedQuantity)
	if err != nil {
		return "", "", "", err
	}
	holdQty, err := normalizeInboundQCQuantityAllowZero(input.HoldQuantity)
	if err != nil {
		return "", "", "", err
	}

	return passedQty, failedQty, holdQty, nil
}

func validateDecisionQuantities(
	inspectionQty decimal.Decimal,
	result InboundQCResult,
	passedQty decimal.Decimal,
	failedQty decimal.Decimal,
	holdQty decimal.Decimal,
) error {
	total, err := decimal.AddQuantity(passedQty, failedQty)
	if err != nil {
		return ErrInboundQCInspectionInvalidQuantity
	}
	total, err = decimal.AddQuantity(total, holdQty)
	if err != nil {
		return ErrInboundQCInspectionInvalidQuantity
	}
	if total.String() != inspectionQty.String() {
		return ErrInboundQCInspectionInvalidQuantity
	}

	switch result {
	case InboundQCResultPass:
		if passedQty.String() != inspectionQty.String() || !failedQty.IsZero() || !holdQty.IsZero() {
			return ErrInboundQCInspectionInvalidQuantity
		}
	case InboundQCResultFail:
		if failedQty.String() != inspectionQty.String() || !passedQty.IsZero() || !holdQty.IsZero() {
			return ErrInboundQCInspectionInvalidQuantity
		}
	case InboundQCResultHold:
		if holdQty.String() != inspectionQty.String() || !passedQty.IsZero() || !failedQty.IsZero() {
			return ErrInboundQCInspectionInvalidQuantity
		}
	case InboundQCResultPartial:
		nonZeroBuckets := 0
		for _, quantity := range []decimal.Decimal{passedQty, failedQty, holdQty} {
			if !quantity.IsZero() {
				nonZeroBuckets++
			}
		}
		if nonZeroBuckets < 2 {
			return ErrInboundQCInspectionInvalidQuantity
		}
	}

	return nil
}

func normalizeInboundQCQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil || quantity.IsNegative() {
		return "", ErrInboundQCInspectionInvalidQuantity
	}

	return quantity, nil
}

func normalizeInboundQCQuantityAllowZero(value decimal.Decimal) (decimal.Decimal, error) {
	if strings.TrimSpace(value.String()) == "" {
		return decimal.MustQuantity("0"), nil
	}

	return normalizeInboundQCQuantity(value)
}

func hasIncompleteRequiredChecklist(items []InboundQCChecklistItem) bool {
	for _, item := range items {
		if item.Required &&
			item.Status != InboundQCChecklistStatusPass &&
			item.Status != InboundQCChecklistStatusFail &&
			item.Status != InboundQCChecklistStatusNotApplicable {
			return true
		}
	}

	return false
}

func requiresDecisionReason(result InboundQCResult) bool {
	switch result {
	case InboundQCResultFail, InboundQCResultHold, InboundQCResultPartial:
		return true
	default:
		return false
	}
}

func normalizeInboundQCLot(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func dateOnly(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}

	return time.Date(value.UTC().Year(), value.UTC().Month(), value.UTC().Day(), 0, 0, 0, 0, time.UTC)
}
