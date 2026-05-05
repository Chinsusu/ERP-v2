package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractFactoryDispatchRequiredField = errors.New("subcontract factory dispatch required field is missing")
var ErrSubcontractFactoryDispatchInvalidStatus = errors.New("subcontract factory dispatch status is invalid")
var ErrSubcontractFactoryDispatchInvalidTransition = errors.New("subcontract factory dispatch status transition is invalid")
var ErrSubcontractFactoryDispatchInvalidQuantity = errors.New("subcontract factory dispatch quantity is invalid")

type SubcontractFactoryDispatchStatus string

const (
	SubcontractFactoryDispatchStatusDraft             SubcontractFactoryDispatchStatus = "draft"
	SubcontractFactoryDispatchStatusReady             SubcontractFactoryDispatchStatus = "ready"
	SubcontractFactoryDispatchStatusSent              SubcontractFactoryDispatchStatus = "sent"
	SubcontractFactoryDispatchStatusConfirmed         SubcontractFactoryDispatchStatus = "confirmed"
	SubcontractFactoryDispatchStatusRevisionRequested SubcontractFactoryDispatchStatus = "revision_requested"
	SubcontractFactoryDispatchStatusRejected          SubcontractFactoryDispatchStatus = "rejected"
	SubcontractFactoryDispatchStatusCancelled         SubcontractFactoryDispatchStatus = "cancelled"
)

type SubcontractFactoryDispatch struct {
	ID                     string
	OrgID                  string
	DispatchNo             string
	SubcontractOrderID     string
	SubcontractOrderNo     string
	SourceProductionPlanID string
	SourceProductionPlanNo string
	FactoryID              string
	FactoryCode            string
	FactoryName            string
	FinishedItemID         string
	FinishedSKUCode        string
	FinishedItemName       string
	PlannedQty             decimal.Decimal
	UOMCode                decimal.UOMCode
	SpecSummary            string
	SampleRequired         bool
	TargetStartDate        string
	ExpectedReceiptDate    string
	Status                 SubcontractFactoryDispatchStatus
	Lines                  []SubcontractFactoryDispatchLine
	Evidence               []SubcontractFactoryDispatchEvidence
	ReadyAt                time.Time
	ReadyBy                string
	SentAt                 time.Time
	SentBy                 string
	RespondedAt            time.Time
	ResponseBy             string
	FactoryResponseNote    string
	Note                   string
	CreatedAt              time.Time
	CreatedBy              string
	UpdatedAt              time.Time
	UpdatedBy              string
	Version                int
}

type SubcontractFactoryDispatchLine struct {
	ID                  string
	LineNo              int
	OrderMaterialLineID string
	ItemID              string
	SKUCode             string
	ItemName            string
	PlannedQty          decimal.Decimal
	UOMCode             decimal.UOMCode
	LotTraceRequired    bool
	Note                string
}

type SubcontractFactoryDispatchEvidence struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type NewSubcontractFactoryDispatchInput struct {
	ID                     string
	OrgID                  string
	DispatchNo             string
	SubcontractOrderID     string
	SubcontractOrderNo     string
	SourceProductionPlanID string
	SourceProductionPlanNo string
	FactoryID              string
	FactoryCode            string
	FactoryName            string
	FinishedItemID         string
	FinishedSKUCode        string
	FinishedItemName       string
	PlannedQty             decimal.Decimal
	UOMCode                string
	SpecSummary            string
	SampleRequired         bool
	TargetStartDate        string
	ExpectedReceiptDate    string
	Status                 SubcontractFactoryDispatchStatus
	Lines                  []NewSubcontractFactoryDispatchLineInput
	Evidence               []NewSubcontractFactoryDispatchEvidenceInput
	ReadyAt                time.Time
	ReadyBy                string
	SentAt                 time.Time
	SentBy                 string
	RespondedAt            time.Time
	ResponseBy             string
	FactoryResponseNote    string
	Note                   string
	CreatedAt              time.Time
	CreatedBy              string
	UpdatedAt              time.Time
	UpdatedBy              string
}

type NewSubcontractFactoryDispatchLineInput struct {
	ID                  string
	LineNo              int
	OrderMaterialLineID string
	ItemID              string
	SKUCode             string
	ItemName            string
	PlannedQty          decimal.Decimal
	UOMCode             string
	LotTraceRequired    bool
	Note                string
}

type NewSubcontractFactoryDispatchEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type MarkSubcontractFactoryDispatchSentInput struct {
	SentBy   string
	SentAt   time.Time
	Evidence []NewSubcontractFactoryDispatchEvidenceInput
	Note     string
}

type RecordSubcontractFactoryDispatchResponseInput struct {
	ResponseStatus SubcontractFactoryDispatchStatus
	ResponseBy     string
	RespondedAt    time.Time
	ResponseNote   string
}

func NewSubcontractFactoryDispatch(input NewSubcontractFactoryDispatchInput) (SubcontractFactoryDispatch, error) {
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
	status := NormalizeSubcontractFactoryDispatchStatus(input.Status)
	if status == "" {
		status = SubcontractFactoryDispatchStatusDraft
	}
	plannedQty, err := decimal.ParseQuantity(input.PlannedQty.String())
	if err != nil || plannedQty.IsNegative() || plannedQty.IsZero() {
		return SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchInvalidQuantity
	}
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchInvalidQuantity
	}

	dispatch := SubcontractFactoryDispatch{
		ID:                     strings.TrimSpace(input.ID),
		OrgID:                  strings.TrimSpace(input.OrgID),
		DispatchNo:             strings.ToUpper(strings.TrimSpace(input.DispatchNo)),
		SubcontractOrderID:     strings.TrimSpace(input.SubcontractOrderID),
		SubcontractOrderNo:     strings.ToUpper(strings.TrimSpace(input.SubcontractOrderNo)),
		SourceProductionPlanID: strings.TrimSpace(input.SourceProductionPlanID),
		SourceProductionPlanNo: strings.ToUpper(strings.TrimSpace(input.SourceProductionPlanNo)),
		FactoryID:              strings.TrimSpace(input.FactoryID),
		FactoryCode:            strings.ToUpper(strings.TrimSpace(input.FactoryCode)),
		FactoryName:            strings.TrimSpace(input.FactoryName),
		FinishedItemID:         strings.TrimSpace(input.FinishedItemID),
		FinishedSKUCode:        strings.ToUpper(strings.TrimSpace(input.FinishedSKUCode)),
		FinishedItemName:       strings.TrimSpace(input.FinishedItemName),
		PlannedQty:             plannedQty,
		UOMCode:                uomCode,
		SpecSummary:            strings.TrimSpace(input.SpecSummary),
		SampleRequired:         input.SampleRequired,
		TargetStartDate:        strings.TrimSpace(input.TargetStartDate),
		ExpectedReceiptDate:    strings.TrimSpace(input.ExpectedReceiptDate),
		Status:                 status,
		Lines:                  make([]SubcontractFactoryDispatchLine, 0, len(input.Lines)),
		Evidence:               make([]SubcontractFactoryDispatchEvidence, 0, len(input.Evidence)),
		ReadyAt:                input.ReadyAt.UTC(),
		ReadyBy:                strings.TrimSpace(input.ReadyBy),
		SentAt:                 input.SentAt.UTC(),
		SentBy:                 strings.TrimSpace(input.SentBy),
		RespondedAt:            input.RespondedAt.UTC(),
		ResponseBy:             strings.TrimSpace(input.ResponseBy),
		FactoryResponseNote:    strings.TrimSpace(input.FactoryResponseNote),
		Note:                   strings.TrimSpace(input.Note),
		CreatedAt:              createdAt.UTC(),
		CreatedBy:              strings.TrimSpace(input.CreatedBy),
		UpdatedAt:              updatedAt.UTC(),
		UpdatedBy:              updatedBy,
		Version:                1,
	}
	for index, lineInput := range input.Lines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		line, err := NewSubcontractFactoryDispatchLine(lineInput)
		if err != nil {
			return SubcontractFactoryDispatch{}, err
		}
		dispatch.Lines = append(dispatch.Lines, line)
	}
	for _, evidenceInput := range input.Evidence {
		evidence, err := NewSubcontractFactoryDispatchEvidence(evidenceInput)
		if err != nil {
			return SubcontractFactoryDispatch{}, err
		}
		dispatch.Evidence = append(dispatch.Evidence, evidence)
	}
	if err := dispatch.Validate(); err != nil {
		return SubcontractFactoryDispatch{}, err
	}

	return dispatch, nil
}

func NewSubcontractFactoryDispatchLine(input NewSubcontractFactoryDispatchLineInput) (SubcontractFactoryDispatchLine, error) {
	plannedQty, err := decimal.ParseQuantity(input.PlannedQty.String())
	if err != nil || plannedQty.IsNegative() || plannedQty.IsZero() {
		return SubcontractFactoryDispatchLine{}, ErrSubcontractFactoryDispatchInvalidQuantity
	}
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return SubcontractFactoryDispatchLine{}, ErrSubcontractFactoryDispatchInvalidQuantity
	}

	line := SubcontractFactoryDispatchLine{
		ID:                  strings.TrimSpace(input.ID),
		LineNo:              input.LineNo,
		OrderMaterialLineID: strings.TrimSpace(input.OrderMaterialLineID),
		ItemID:              strings.TrimSpace(input.ItemID),
		SKUCode:             strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		ItemName:            strings.TrimSpace(input.ItemName),
		PlannedQty:          plannedQty,
		UOMCode:             uomCode,
		LotTraceRequired:    input.LotTraceRequired,
		Note:                strings.TrimSpace(input.Note),
	}
	if err := line.Validate(); err != nil {
		return SubcontractFactoryDispatchLine{}, err
	}

	return line, nil
}

func NewSubcontractFactoryDispatchEvidence(
	input NewSubcontractFactoryDispatchEvidenceInput,
) (SubcontractFactoryDispatchEvidence, error) {
	evidence := SubcontractFactoryDispatchEvidence{
		ID:           strings.TrimSpace(input.ID),
		EvidenceType: strings.ToLower(strings.TrimSpace(input.EvidenceType)),
		FileName:     strings.TrimSpace(input.FileName),
		ObjectKey:    strings.TrimSpace(input.ObjectKey),
		ExternalURL:  strings.TrimSpace(input.ExternalURL),
		Note:         strings.TrimSpace(input.Note),
	}
	if err := evidence.Validate(); err != nil {
		return SubcontractFactoryDispatchEvidence{}, err
	}

	return evidence, nil
}

func (d SubcontractFactoryDispatch) MarkReady(actorID string, changedAt time.Time) (SubcontractFactoryDispatch, error) {
	return d.transitionTo(SubcontractFactoryDispatchStatusReady, actorID, changedAt)
}

func (d SubcontractFactoryDispatch) MarkSent(input MarkSubcontractFactoryDispatchSentInput) (SubcontractFactoryDispatch, error) {
	sent, err := d.transitionTo(SubcontractFactoryDispatchStatusSent, input.SentBy, input.SentAt)
	if err != nil {
		return SubcontractFactoryDispatch{}, err
	}
	for _, evidenceInput := range input.Evidence {
		evidence, err := NewSubcontractFactoryDispatchEvidence(evidenceInput)
		if err != nil {
			return SubcontractFactoryDispatch{}, err
		}
		sent.Evidence = append(sent.Evidence, evidence)
	}
	if note := strings.TrimSpace(input.Note); note != "" {
		sent.Note = note
	}
	if err := sent.Validate(); err != nil {
		return SubcontractFactoryDispatch{}, err
	}

	return sent, nil
}

func (d SubcontractFactoryDispatch) RecordResponse(input RecordSubcontractFactoryDispatchResponseInput) (SubcontractFactoryDispatch, error) {
	status := NormalizeSubcontractFactoryDispatchStatus(input.ResponseStatus)
	if status != SubcontractFactoryDispatchStatusConfirmed &&
		status != SubcontractFactoryDispatchStatusRevisionRequested &&
		status != SubcontractFactoryDispatchStatusRejected {
		return SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchInvalidStatus
	}
	if status != SubcontractFactoryDispatchStatusConfirmed && strings.TrimSpace(input.ResponseNote) == "" {
		return SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchRequiredField
	}
	responded, err := d.transitionTo(status, input.ResponseBy, input.RespondedAt)
	if err != nil {
		return SubcontractFactoryDispatch{}, err
	}
	responded.FactoryResponseNote = strings.TrimSpace(input.ResponseNote)
	if err := responded.Validate(); err != nil {
		return SubcontractFactoryDispatch{}, err
	}

	return responded, nil
}

func (d SubcontractFactoryDispatch) Validate() error {
	if strings.TrimSpace(d.ID) == "" ||
		strings.TrimSpace(d.OrgID) == "" ||
		strings.TrimSpace(d.DispatchNo) == "" ||
		strings.TrimSpace(d.SubcontractOrderID) == "" ||
		strings.TrimSpace(d.SubcontractOrderNo) == "" ||
		strings.TrimSpace(d.FactoryID) == "" ||
		strings.TrimSpace(d.FactoryName) == "" ||
		strings.TrimSpace(d.FinishedItemID) == "" ||
		strings.TrimSpace(d.FinishedSKUCode) == "" ||
		strings.TrimSpace(d.FinishedItemName) == "" ||
		strings.TrimSpace(d.CreatedBy) == "" ||
		strings.TrimSpace(d.UpdatedBy) == "" ||
		len(d.Lines) == 0 {
		return ErrSubcontractFactoryDispatchRequiredField
	}
	if !IsValidSubcontractFactoryDispatchStatus(d.Status) {
		return ErrSubcontractFactoryDispatchInvalidStatus
	}
	plannedQty, err := decimal.ParseQuantity(d.PlannedQty.String())
	if err != nil || plannedQty.IsNegative() || plannedQty.IsZero() {
		return ErrSubcontractFactoryDispatchInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(d.UOMCode.String()); err != nil {
		return ErrSubcontractFactoryDispatchInvalidQuantity
	}
	for _, line := range d.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}
	for _, evidence := range d.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}
	if d.Status == SubcontractFactoryDispatchStatusReady && (d.ReadyAt.IsZero() || strings.TrimSpace(d.ReadyBy) == "") {
		return ErrSubcontractFactoryDispatchRequiredField
	}
	if d.Status == SubcontractFactoryDispatchStatusSent && (d.SentAt.IsZero() || strings.TrimSpace(d.SentBy) == "") {
		return ErrSubcontractFactoryDispatchRequiredField
	}
	if (d.Status == SubcontractFactoryDispatchStatusConfirmed ||
		d.Status == SubcontractFactoryDispatchStatusRevisionRequested ||
		d.Status == SubcontractFactoryDispatchStatusRejected) &&
		(d.RespondedAt.IsZero() || strings.TrimSpace(d.ResponseBy) == "") {
		return ErrSubcontractFactoryDispatchRequiredField
	}
	if (d.Status == SubcontractFactoryDispatchStatusRevisionRequested || d.Status == SubcontractFactoryDispatchStatusRejected) &&
		strings.TrimSpace(d.FactoryResponseNote) == "" {
		return ErrSubcontractFactoryDispatchRequiredField
	}

	return nil
}

func (l SubcontractFactoryDispatchLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.OrderMaterialLineID) == "" ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKUCode) == "" ||
		strings.TrimSpace(l.ItemName) == "" {
		return ErrSubcontractFactoryDispatchRequiredField
	}
	plannedQty, err := decimal.ParseQuantity(l.PlannedQty.String())
	if err != nil || plannedQty.IsNegative() || plannedQty.IsZero() {
		return ErrSubcontractFactoryDispatchInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrSubcontractFactoryDispatchInvalidQuantity
	}

	return nil
}

func (e SubcontractFactoryDispatchEvidence) Validate() error {
	if strings.TrimSpace(e.ID) == "" || strings.TrimSpace(e.EvidenceType) == "" {
		return ErrSubcontractFactoryDispatchRequiredField
	}
	if strings.TrimSpace(e.ObjectKey) == "" && strings.TrimSpace(e.ExternalURL) == "" {
		return ErrSubcontractFactoryDispatchRequiredField
	}

	return nil
}

func (d SubcontractFactoryDispatch) transitionTo(status SubcontractFactoryDispatchStatus, actorID string, changedAt time.Time) (SubcontractFactoryDispatch, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchRequiredField
	}
	from := NormalizeSubcontractFactoryDispatchStatus(d.Status)
	to := NormalizeSubcontractFactoryDispatchStatus(status)
	if !CanTransitionSubcontractFactoryDispatchStatus(from, to) {
		return SubcontractFactoryDispatch{}, ErrSubcontractFactoryDispatchInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := d.Clone()
	updated.Status = to
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if updated.Version > 0 {
		updated.Version++
	}
	switch to {
	case SubcontractFactoryDispatchStatusReady:
		updated.ReadyAt = changedAt.UTC()
		updated.ReadyBy = actorID
	case SubcontractFactoryDispatchStatusSent:
		updated.SentAt = changedAt.UTC()
		updated.SentBy = actorID
	case SubcontractFactoryDispatchStatusConfirmed,
		SubcontractFactoryDispatchStatusRevisionRequested,
		SubcontractFactoryDispatchStatusRejected:
		updated.RespondedAt = changedAt.UTC()
		updated.ResponseBy = actorID
	}

	return updated, nil
}

func (d SubcontractFactoryDispatch) Clone() SubcontractFactoryDispatch {
	clone := d
	clone.Lines = append([]SubcontractFactoryDispatchLine(nil), d.Lines...)
	clone.Evidence = append([]SubcontractFactoryDispatchEvidence(nil), d.Evidence...)

	return clone
}

func NormalizeSubcontractFactoryDispatchStatus(status SubcontractFactoryDispatchStatus) SubcontractFactoryDispatchStatus {
	return SubcontractFactoryDispatchStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSubcontractFactoryDispatchStatus(status SubcontractFactoryDispatchStatus) bool {
	switch NormalizeSubcontractFactoryDispatchStatus(status) {
	case SubcontractFactoryDispatchStatusDraft,
		SubcontractFactoryDispatchStatusReady,
		SubcontractFactoryDispatchStatusSent,
		SubcontractFactoryDispatchStatusConfirmed,
		SubcontractFactoryDispatchStatusRevisionRequested,
		SubcontractFactoryDispatchStatusRejected,
		SubcontractFactoryDispatchStatusCancelled:
		return true
	default:
		return false
	}
}

func CanTransitionSubcontractFactoryDispatchStatus(from, to SubcontractFactoryDispatchStatus) bool {
	from = NormalizeSubcontractFactoryDispatchStatus(from)
	to = NormalizeSubcontractFactoryDispatchStatus(to)
	if !IsValidSubcontractFactoryDispatchStatus(from) || !IsValidSubcontractFactoryDispatchStatus(to) {
		return false
	}
	switch from {
	case SubcontractFactoryDispatchStatusDraft:
		return to == SubcontractFactoryDispatchStatusReady || to == SubcontractFactoryDispatchStatusCancelled
	case SubcontractFactoryDispatchStatusReady:
		return to == SubcontractFactoryDispatchStatusSent || to == SubcontractFactoryDispatchStatusCancelled
	case SubcontractFactoryDispatchStatusSent:
		return to == SubcontractFactoryDispatchStatusConfirmed ||
			to == SubcontractFactoryDispatchStatusRevisionRequested ||
			to == SubcontractFactoryDispatchStatusRejected ||
			to == SubcontractFactoryDispatchStatusCancelled
	case SubcontractFactoryDispatchStatusRevisionRequested:
		return to == SubcontractFactoryDispatchStatusReady || to == SubcontractFactoryDispatchStatusCancelled
	default:
		return false
	}
}
