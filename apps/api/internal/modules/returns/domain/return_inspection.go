package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type ReturnInspectionCondition string

const ReturnInspectionConditionIntact ReturnInspectionCondition = "intact"
const ReturnInspectionConditionDentedBox ReturnInspectionCondition = "dented_box"
const ReturnInspectionConditionSealTorn ReturnInspectionCondition = "seal_torn"
const ReturnInspectionConditionUsed ReturnInspectionCondition = "used"
const ReturnInspectionConditionDamaged ReturnInspectionCondition = "damaged"
const ReturnInspectionConditionMissingAccessory ReturnInspectionCondition = "missing_accessory"

type ReturnInspectionStatus string

const ReturnInspectionStatusRecorded ReturnInspectionStatus = "inspection_recorded"
const ReturnInspectionStatusQAHold ReturnInspectionStatus = "return_qa_hold"

var ErrReturnInspectionRequiredField = errors.New("return inspection required field is missing")
var ErrReturnInspectionInvalidCondition = errors.New("return inspection condition is invalid")

type ReturnInspection struct {
	ID             string
	ReceiptID      string
	ReceiptNo      string
	Condition      ReturnInspectionCondition
	Disposition    ReturnDisposition
	Status         ReturnInspectionStatus
	TargetLocation string
	RiskLevel      string
	InspectorID    string
	Note           string
	EvidenceLabel  string
	InspectedAt    time.Time
}

type NewReturnInspectionInput struct {
	ID            string
	ReceiptID     string
	ReceiptNo     string
	Condition     ReturnInspectionCondition
	Disposition   ReturnDisposition
	InspectorID   string
	Note          string
	EvidenceLabel string
	InspectedAt   time.Time
}

func NewReturnInspection(input NewReturnInspectionInput) (ReturnInspection, error) {
	receiptID := strings.TrimSpace(input.ReceiptID)
	if receiptID == "" || strings.TrimSpace(input.InspectorID) == "" {
		return ReturnInspection{}, ErrReturnInspectionRequiredField
	}

	condition := NormalizeReturnInspectionCondition(input.Condition)
	if condition == "" {
		return ReturnInspection{}, ErrReturnInspectionInvalidCondition
	}

	disposition := NormalizeReturnDisposition(input.Disposition)
	if disposition == "" {
		return ReturnInspection{}, ErrReturnReceiptInvalidDisposition
	}

	inspectedAt := input.InspectedAt
	if inspectedAt.IsZero() {
		inspectedAt = time.Now().UTC()
	}

	inspection := ReturnInspection{
		ID:             strings.TrimSpace(input.ID),
		ReceiptID:      receiptID,
		ReceiptNo:      strings.TrimSpace(input.ReceiptNo),
		Condition:      condition,
		Disposition:    disposition,
		Status:         returnInspectionStatusForDisposition(disposition),
		TargetLocation: returnInspectionTargetLocation(disposition),
		RiskLevel:      returnInspectionRiskLevel(condition, disposition),
		InspectorID:    strings.TrimSpace(input.InspectorID),
		Note:           strings.TrimSpace(input.Note),
		EvidenceLabel:  strings.TrimSpace(input.EvidenceLabel),
		InspectedAt:    inspectedAt.UTC(),
	}
	if inspection.ID == "" {
		inspection.ID = fmt.Sprintf("inspect-%s-%s", strings.ToLower(receiptID), condition)
	}

	return inspection, nil
}

func (inspection ReturnInspection) Clone() ReturnInspection {
	return inspection
}

func (r ReturnReceipt) ApplyInspection(inspection ReturnInspection) ReturnReceipt {
	receipt := r.Clone()
	receipt.Status = ReturnStatusInspected
	receipt.Disposition = inspection.Disposition
	receipt.PackageCondition = string(inspection.Condition)
	receipt.TargetLocation = inspection.TargetLocation
	receipt.StockMovement = nil
	for index := range receipt.Lines {
		receipt.Lines[index].Condition = string(inspection.Condition)
	}

	return receipt
}

func NormalizeReturnInspectionCondition(condition ReturnInspectionCondition) ReturnInspectionCondition {
	switch ReturnInspectionCondition(strings.ToLower(strings.TrimSpace(string(condition)))) {
	case ReturnInspectionConditionIntact:
		return ReturnInspectionConditionIntact
	case ReturnInspectionConditionDentedBox:
		return ReturnInspectionConditionDentedBox
	case ReturnInspectionConditionSealTorn:
		return ReturnInspectionConditionSealTorn
	case ReturnInspectionConditionUsed:
		return ReturnInspectionConditionUsed
	case ReturnInspectionConditionDamaged:
		return ReturnInspectionConditionDamaged
	case ReturnInspectionConditionMissingAccessory:
		return ReturnInspectionConditionMissingAccessory
	default:
		return ""
	}
}

func returnInspectionStatusForDisposition(disposition ReturnDisposition) ReturnInspectionStatus {
	if disposition == ReturnDispositionNeedsInspection {
		return ReturnInspectionStatusQAHold
	}

	return ReturnInspectionStatusRecorded
}

func returnInspectionTargetLocation(disposition ReturnDisposition) string {
	switch disposition {
	case ReturnDispositionReusable:
		return "return-area-qc-release"
	case ReturnDispositionNotReusable:
		return "lab-damaged-placeholder"
	case ReturnDispositionNeedsInspection:
		return "return-qa-hold"
	default:
		return "return-qa-hold"
	}
}

func returnInspectionRiskLevel(condition ReturnInspectionCondition, disposition ReturnDisposition) string {
	if condition == ReturnInspectionConditionDamaged ||
		condition == ReturnInspectionConditionMissingAccessory ||
		disposition == ReturnDispositionNotReusable {
		return "high"
	}
	if condition == ReturnInspectionConditionUsed ||
		condition == ReturnInspectionConditionSealTorn ||
		disposition == ReturnDispositionNeedsInspection {
		return "medium"
	}

	return "low"
}
