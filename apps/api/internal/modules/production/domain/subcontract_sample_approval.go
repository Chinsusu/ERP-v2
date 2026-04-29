package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrSubcontractSampleApprovalRequiredField = errors.New("subcontract sample approval required field is missing")
var ErrSubcontractSampleApprovalInvalidStatus = errors.New("subcontract sample approval status is invalid")
var ErrSubcontractSampleApprovalInvalidTransition = errors.New("subcontract sample approval status transition is invalid")

type SubcontractSampleApprovalStatus string

const (
	SubcontractSampleApprovalStatusSubmitted SubcontractSampleApprovalStatus = "submitted"
	SubcontractSampleApprovalStatusApproved  SubcontractSampleApprovalStatus = "approved"
	SubcontractSampleApprovalStatusRejected  SubcontractSampleApprovalStatus = "rejected"
)

type SubcontractSampleApproval struct {
	ID                 string
	OrgID              string
	SubcontractOrderID string
	SubcontractOrderNo string
	SampleCode         string
	FormulaVersion     string
	SpecVersion        string
	Status             SubcontractSampleApprovalStatus
	Evidence           []SubcontractSampleApprovalEvidence
	SubmittedBy        string
	SubmittedAt        time.Time
	DecisionBy         string
	DecisionAt         time.Time
	DecisionReason     string
	StorageStatus      string
	Note               string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
	Version            int
}

type SubcontractSampleApprovalEvidence struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
	CreatedAt    time.Time
	CreatedBy    string
}

type NewSubcontractSampleApprovalInput struct {
	ID                 string
	OrgID              string
	SubcontractOrderID string
	SubcontractOrderNo string
	SampleCode         string
	FormulaVersion     string
	SpecVersion        string
	Status             SubcontractSampleApprovalStatus
	Evidence           []NewSubcontractSampleApprovalEvidenceInput
	SubmittedBy        string
	SubmittedAt        time.Time
	DecisionBy         string
	DecisionAt         time.Time
	DecisionReason     string
	StorageStatus      string
	Note               string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
}

type NewSubcontractSampleApprovalEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
	CreatedAt    time.Time
	CreatedBy    string
}

type ApproveSubcontractSampleApprovalInput struct {
	ActorID       string
	DecisionAt    time.Time
	Reason        string
	StorageStatus string
}

type RejectSubcontractSampleApprovalInput struct {
	ActorID    string
	DecisionAt time.Time
	Reason     string
}

func NewSubcontractSampleApproval(input NewSubcontractSampleApprovalInput) (SubcontractSampleApproval, error) {
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
	submittedAt := input.SubmittedAt
	if submittedAt.IsZero() {
		submittedAt = createdAt
	}
	status := NormalizeSubcontractSampleApprovalStatus(input.Status)
	if status == "" {
		status = SubcontractSampleApprovalStatusSubmitted
	}

	approval := SubcontractSampleApproval{
		ID:                 strings.TrimSpace(input.ID),
		OrgID:              strings.TrimSpace(input.OrgID),
		SubcontractOrderID: strings.TrimSpace(input.SubcontractOrderID),
		SubcontractOrderNo: strings.ToUpper(strings.TrimSpace(input.SubcontractOrderNo)),
		SampleCode:         strings.ToUpper(strings.TrimSpace(input.SampleCode)),
		FormulaVersion:     strings.TrimSpace(input.FormulaVersion),
		SpecVersion:        strings.TrimSpace(input.SpecVersion),
		Status:             status,
		Evidence:           make([]SubcontractSampleApprovalEvidence, 0, len(input.Evidence)),
		SubmittedBy:        strings.TrimSpace(input.SubmittedBy),
		SubmittedAt:        submittedAt.UTC(),
		DecisionBy:         strings.TrimSpace(input.DecisionBy),
		DecisionReason:     strings.TrimSpace(input.DecisionReason),
		StorageStatus:      strings.ToLower(strings.TrimSpace(input.StorageStatus)),
		Note:               strings.TrimSpace(input.Note),
		CreatedAt:          createdAt.UTC(),
		CreatedBy:          strings.TrimSpace(input.CreatedBy),
		UpdatedAt:          updatedAt.UTC(),
		UpdatedBy:          updatedBy,
		Version:            1,
	}
	if !input.DecisionAt.IsZero() {
		approval.DecisionAt = input.DecisionAt.UTC()
	}
	for index, evidenceInput := range input.Evidence {
		if evidenceInput.ID == "" {
			evidenceInput.ID = fmt.Sprintf("%s-evidence-%02d", approval.ID, index+1)
		}
		if evidenceInput.CreatedAt.IsZero() {
			evidenceInput.CreatedAt = createdAt
		}
		if strings.TrimSpace(evidenceInput.CreatedBy) == "" {
			evidenceInput.CreatedBy = approval.CreatedBy
		}
		evidence, err := NewSubcontractSampleApprovalEvidence(evidenceInput)
		if err != nil {
			return SubcontractSampleApproval{}, err
		}
		approval.Evidence = append(approval.Evidence, evidence)
	}
	if err := approval.Validate(); err != nil {
		return SubcontractSampleApproval{}, err
	}

	return approval, nil
}

func NewSubcontractSampleApprovalEvidence(
	input NewSubcontractSampleApprovalEvidenceInput,
) (SubcontractSampleApprovalEvidence, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	evidence := SubcontractSampleApprovalEvidence{
		ID:           strings.TrimSpace(input.ID),
		EvidenceType: strings.ToLower(strings.TrimSpace(input.EvidenceType)),
		FileName:     strings.TrimSpace(input.FileName),
		ObjectKey:    strings.TrimSpace(input.ObjectKey),
		ExternalURL:  strings.TrimSpace(input.ExternalURL),
		Note:         strings.TrimSpace(input.Note),
		CreatedAt:    createdAt.UTC(),
		CreatedBy:    strings.TrimSpace(input.CreatedBy),
	}
	if err := evidence.Validate(); err != nil {
		return SubcontractSampleApprovalEvidence{}, err
	}

	return evidence, nil
}

func (s SubcontractSampleApproval) Approve(
	input ApproveSubcontractSampleApprovalInput,
) (SubcontractSampleApproval, error) {
	if NormalizeSubcontractSampleApprovalStatus(s.Status) != SubcontractSampleApprovalStatusSubmitted {
		return SubcontractSampleApproval{}, ErrSubcontractSampleApprovalInvalidTransition
	}
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" || strings.TrimSpace(input.StorageStatus) == "" {
		return SubcontractSampleApproval{}, ErrSubcontractSampleApprovalRequiredField
	}
	decisionAt := input.DecisionAt
	if decisionAt.IsZero() {
		decisionAt = time.Now().UTC()
	}

	approved := s.Clone()
	approved.Status = SubcontractSampleApprovalStatusApproved
	approved.DecisionBy = actorID
	approved.DecisionAt = decisionAt.UTC()
	approved.DecisionReason = strings.TrimSpace(input.Reason)
	approved.StorageStatus = strings.ToLower(strings.TrimSpace(input.StorageStatus))
	approved.UpdatedAt = decisionAt.UTC()
	approved.UpdatedBy = actorID
	if approved.Version > 0 {
		approved.Version++
	}
	if err := approved.Validate(); err != nil {
		return SubcontractSampleApproval{}, err
	}

	return approved, nil
}

func (s SubcontractSampleApproval) Reject(
	input RejectSubcontractSampleApprovalInput,
) (SubcontractSampleApproval, error) {
	if NormalizeSubcontractSampleApprovalStatus(s.Status) != SubcontractSampleApprovalStatusSubmitted {
		return SubcontractSampleApproval{}, ErrSubcontractSampleApprovalInvalidTransition
	}
	actorID := strings.TrimSpace(input.ActorID)
	reason := strings.TrimSpace(input.Reason)
	if actorID == "" || reason == "" {
		return SubcontractSampleApproval{}, ErrSubcontractSampleApprovalRequiredField
	}
	decisionAt := input.DecisionAt
	if decisionAt.IsZero() {
		decisionAt = time.Now().UTC()
	}

	rejected := s.Clone()
	rejected.Status = SubcontractSampleApprovalStatusRejected
	rejected.DecisionBy = actorID
	rejected.DecisionAt = decisionAt.UTC()
	rejected.DecisionReason = reason
	rejected.StorageStatus = ""
	rejected.UpdatedAt = decisionAt.UTC()
	rejected.UpdatedBy = actorID
	if rejected.Version > 0 {
		rejected.Version++
	}
	if err := rejected.Validate(); err != nil {
		return SubcontractSampleApproval{}, err
	}

	return rejected, nil
}

func (s SubcontractSampleApproval) Validate() error {
	if strings.TrimSpace(s.ID) == "" ||
		strings.TrimSpace(s.OrgID) == "" ||
		strings.TrimSpace(s.SubcontractOrderID) == "" ||
		strings.TrimSpace(s.SubcontractOrderNo) == "" ||
		strings.TrimSpace(s.SampleCode) == "" ||
		strings.TrimSpace(s.SubmittedBy) == "" ||
		strings.TrimSpace(s.CreatedBy) == "" ||
		len(s.Evidence) == 0 {
		return ErrSubcontractSampleApprovalRequiredField
	}
	if !IsValidSubcontractSampleApprovalStatus(s.Status) {
		return ErrSubcontractSampleApprovalInvalidStatus
	}
	if s.Status == SubcontractSampleApprovalStatusApproved &&
		(strings.TrimSpace(s.DecisionBy) == "" || s.DecisionAt.IsZero() || strings.TrimSpace(s.StorageStatus) == "") {
		return ErrSubcontractSampleApprovalRequiredField
	}
	if s.Status == SubcontractSampleApprovalStatusRejected &&
		(strings.TrimSpace(s.DecisionBy) == "" || s.DecisionAt.IsZero() || strings.TrimSpace(s.DecisionReason) == "") {
		return ErrSubcontractSampleApprovalRequiredField
	}
	for _, evidence := range s.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (e SubcontractSampleApprovalEvidence) Validate() error {
	if strings.TrimSpace(e.ID) == "" ||
		strings.TrimSpace(e.EvidenceType) == "" ||
		strings.TrimSpace(e.CreatedBy) == "" ||
		(strings.TrimSpace(e.ObjectKey) == "" && strings.TrimSpace(e.ExternalURL) == "") {
		return ErrSubcontractSampleApprovalRequiredField
	}

	return nil
}

func (s SubcontractSampleApproval) Clone() SubcontractSampleApproval {
	clone := s
	clone.Evidence = append([]SubcontractSampleApprovalEvidence(nil), s.Evidence...)

	return clone
}

func NormalizeSubcontractSampleApprovalStatus(status SubcontractSampleApprovalStatus) SubcontractSampleApprovalStatus {
	normalized := SubcontractSampleApprovalStatus(strings.ToLower(strings.TrimSpace(string(status))))
	if IsValidSubcontractSampleApprovalStatus(normalized) {
		return normalized
	}

	return ""
}

func IsValidSubcontractSampleApprovalStatus(status SubcontractSampleApprovalStatus) bool {
	switch status {
	case SubcontractSampleApprovalStatusSubmitted,
		SubcontractSampleApprovalStatusApproved,
		SubcontractSampleApprovalStatusRejected:
		return true
	default:
		return false
	}
}
