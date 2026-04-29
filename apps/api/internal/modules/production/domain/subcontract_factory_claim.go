package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractFactoryClaimRequiredField = errors.New("subcontract factory claim required field is missing")
var ErrSubcontractFactoryClaimInvalidStatus = errors.New("subcontract factory claim status is invalid")
var ErrSubcontractFactoryClaimInvalidQuantity = errors.New("subcontract factory claim quantity is invalid")
var ErrSubcontractFactoryClaimInvalidSLA = errors.New("subcontract factory claim sla is invalid")

type SubcontractFactoryClaimStatus string

const (
	SubcontractFactoryClaimStatusOpen         SubcontractFactoryClaimStatus = "open"
	SubcontractFactoryClaimStatusAcknowledged SubcontractFactoryClaimStatus = "acknowledged"
	SubcontractFactoryClaimStatusResolved     SubcontractFactoryClaimStatus = "resolved"
	SubcontractFactoryClaimStatusClosed       SubcontractFactoryClaimStatus = "closed"
	SubcontractFactoryClaimStatusCancelled    SubcontractFactoryClaimStatus = "cancelled"
)

type SubcontractFactoryClaim struct {
	ID                 string
	OrgID              string
	ClaimNo            string
	SubcontractOrderID string
	SubcontractOrderNo string
	FactoryID          string
	FactoryCode        string
	FactoryName        string
	ReceiptID          string
	ReceiptNo          string
	ReasonCode         string
	Reason             string
	Severity           string
	Status             SubcontractFactoryClaimStatus
	AffectedQty        decimal.Decimal
	UOMCode            decimal.UOMCode
	BaseAffectedQty    decimal.Decimal
	BaseUOMCode        decimal.UOMCode
	Evidence           []SubcontractFactoryClaimEvidence
	OwnerID            string
	OpenedBy           string
	OpenedAt           time.Time
	DueAt              time.Time
	AcknowledgedBy     string
	AcknowledgedAt     time.Time
	ResolvedBy         string
	ResolvedAt         time.Time
	ResolutionNote     string
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
	Version            int
}

type SubcontractFactoryClaimEvidence struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

type NewSubcontractFactoryClaimInput struct {
	ID                 string
	OrgID              string
	ClaimNo            string
	SubcontractOrderID string
	SubcontractOrderNo string
	FactoryID          string
	FactoryCode        string
	FactoryName        string
	ReceiptID          string
	ReceiptNo          string
	ReasonCode         string
	Reason             string
	Severity           string
	Status             SubcontractFactoryClaimStatus
	AffectedQty        decimal.Decimal
	UOMCode            string
	BaseAffectedQty    decimal.Decimal
	BaseUOMCode        string
	Evidence           []NewSubcontractFactoryClaimEvidenceInput
	OwnerID            string
	OpenedBy           string
	OpenedAt           time.Time
	DueAt              time.Time
	CreatedAt          time.Time
	CreatedBy          string
	UpdatedAt          time.Time
	UpdatedBy          string
}

type NewSubcontractFactoryClaimEvidenceInput struct {
	ID           string
	EvidenceType string
	FileName     string
	ObjectKey    string
	ExternalURL  string
	Note         string
}

func NewSubcontractFactoryClaim(input NewSubcontractFactoryClaimInput) (SubcontractFactoryClaim, error) {
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimInvalidQuantity
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimInvalidQuantity
	}
	status := NormalizeSubcontractFactoryClaimStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = SubcontractFactoryClaimStatusOpen
	}
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = input.OpenedAt
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	claim := SubcontractFactoryClaim{
		ID:                 strings.TrimSpace(input.ID),
		OrgID:              strings.TrimSpace(input.OrgID),
		ClaimNo:            strings.ToUpper(strings.TrimSpace(input.ClaimNo)),
		SubcontractOrderID: strings.TrimSpace(input.SubcontractOrderID),
		SubcontractOrderNo: strings.ToUpper(strings.TrimSpace(input.SubcontractOrderNo)),
		FactoryID:          strings.TrimSpace(input.FactoryID),
		FactoryCode:        strings.ToUpper(strings.TrimSpace(input.FactoryCode)),
		FactoryName:        strings.TrimSpace(input.FactoryName),
		ReceiptID:          strings.TrimSpace(input.ReceiptID),
		ReceiptNo:          strings.ToUpper(strings.TrimSpace(input.ReceiptNo)),
		ReasonCode:         strings.ToUpper(strings.TrimSpace(input.ReasonCode)),
		Reason:             strings.TrimSpace(input.Reason),
		Severity:           normalizeSubcontractFactoryClaimSeverity(input.Severity),
		Status:             status,
		AffectedQty:        input.AffectedQty,
		UOMCode:            uomCode,
		BaseAffectedQty:    input.BaseAffectedQty,
		BaseUOMCode:        baseUOMCode,
		Evidence:           make([]SubcontractFactoryClaimEvidence, 0, len(input.Evidence)),
		OwnerID:            strings.TrimSpace(input.OwnerID),
		OpenedBy:           strings.TrimSpace(input.OpenedBy),
		OpenedAt:           input.OpenedAt.UTC(),
		DueAt:              input.DueAt.UTC(),
		CreatedAt:          createdAt.UTC(),
		CreatedBy:          strings.TrimSpace(input.CreatedBy),
		UpdatedAt:          updatedAt.UTC(),
		UpdatedBy:          strings.TrimSpace(input.UpdatedBy),
		Version:            1,
	}
	for _, evidenceInput := range input.Evidence {
		evidence, err := NewSubcontractFactoryClaimEvidence(evidenceInput)
		if err != nil {
			return SubcontractFactoryClaim{}, err
		}
		claim.Evidence = append(claim.Evidence, evidence)
	}
	if err := claim.Validate(); err != nil {
		return SubcontractFactoryClaim{}, err
	}

	return claim, nil
}

func NewSubcontractFactoryClaimEvidence(
	input NewSubcontractFactoryClaimEvidenceInput,
) (SubcontractFactoryClaimEvidence, error) {
	evidence := SubcontractFactoryClaimEvidence{
		ID:           strings.TrimSpace(input.ID),
		EvidenceType: strings.ToLower(strings.TrimSpace(input.EvidenceType)),
		FileName:     strings.TrimSpace(input.FileName),
		ObjectKey:    strings.TrimSpace(input.ObjectKey),
		ExternalURL:  strings.TrimSpace(input.ExternalURL),
		Note:         strings.TrimSpace(input.Note),
	}
	if err := evidence.Validate(); err != nil {
		return SubcontractFactoryClaimEvidence{}, err
	}

	return evidence, nil
}

func (c SubcontractFactoryClaim) Validate() error {
	if strings.TrimSpace(c.ID) == "" ||
		strings.TrimSpace(c.OrgID) == "" ||
		strings.TrimSpace(c.ClaimNo) == "" ||
		strings.TrimSpace(c.SubcontractOrderID) == "" ||
		strings.TrimSpace(c.SubcontractOrderNo) == "" ||
		strings.TrimSpace(c.FactoryID) == "" ||
		strings.TrimSpace(c.FactoryName) == "" ||
		strings.TrimSpace(c.Reason) == "" ||
		strings.TrimSpace(c.Severity) == "" ||
		strings.TrimSpace(c.OwnerID) == "" ||
		strings.TrimSpace(c.OpenedBy) == "" ||
		c.OpenedAt.IsZero() ||
		c.DueAt.IsZero() ||
		len(c.Evidence) == 0 {
		return ErrSubcontractFactoryClaimRequiredField
	}
	if !IsValidSubcontractFactoryClaimStatus(c.Status) {
		return ErrSubcontractFactoryClaimInvalidStatus
	}
	if c.DueAt.Before(c.OpenedAt.AddDate(0, 0, 3)) || c.DueAt.After(c.OpenedAt.AddDate(0, 0, 7)) {
		return ErrSubcontractFactoryClaimInvalidSLA
	}
	for _, quantity := range []decimal.Decimal{c.AffectedQty, c.BaseAffectedQty} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrSubcontractFactoryClaimInvalidQuantity
		}
	}
	if _, err := decimal.NormalizeUOMCode(c.UOMCode.String()); err != nil {
		return ErrSubcontractFactoryClaimInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(c.BaseUOMCode.String()); err != nil {
		return ErrSubcontractFactoryClaimInvalidQuantity
	}
	for _, evidence := range c.Evidence {
		if err := evidence.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (e SubcontractFactoryClaimEvidence) Validate() error {
	if strings.TrimSpace(e.ID) == "" || strings.TrimSpace(e.EvidenceType) == "" {
		return ErrSubcontractFactoryClaimRequiredField
	}
	if strings.TrimSpace(e.FileName) == "" && strings.TrimSpace(e.ObjectKey) == "" && strings.TrimSpace(e.ExternalURL) == "" {
		return ErrSubcontractFactoryClaimRequiredField
	}

	return nil
}

func (c SubcontractFactoryClaim) Acknowledge(actorID string, acknowledgedAt time.Time) (SubcontractFactoryClaim, error) {
	if NormalizeSubcontractFactoryClaimStatus(c.Status) != SubcontractFactoryClaimStatusOpen {
		return SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimInvalidStatus
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimRequiredField
	}
	if acknowledgedAt.IsZero() {
		acknowledgedAt = time.Now().UTC()
	}
	updated := c.Clone()
	updated.Status = SubcontractFactoryClaimStatusAcknowledged
	updated.AcknowledgedBy = actorID
	updated.AcknowledgedAt = acknowledgedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = acknowledgedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (c SubcontractFactoryClaim) Resolve(actorID string, note string, resolvedAt time.Time) (SubcontractFactoryClaim, error) {
	status := NormalizeSubcontractFactoryClaimStatus(c.Status)
	if status != SubcontractFactoryClaimStatusOpen && status != SubcontractFactoryClaimStatusAcknowledged {
		return SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimInvalidStatus
	}
	actorID = strings.TrimSpace(actorID)
	note = strings.TrimSpace(note)
	if actorID == "" || note == "" {
		return SubcontractFactoryClaim{}, ErrSubcontractFactoryClaimRequiredField
	}
	if resolvedAt.IsZero() {
		resolvedAt = time.Now().UTC()
	}
	updated := c.Clone()
	updated.Status = SubcontractFactoryClaimStatusResolved
	updated.ResolvedBy = actorID
	updated.ResolvedAt = resolvedAt.UTC()
	updated.ResolutionNote = note
	updated.UpdatedBy = actorID
	updated.UpdatedAt = resolvedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (c SubcontractFactoryClaim) BlocksFinalPayment() bool {
	switch NormalizeSubcontractFactoryClaimStatus(c.Status) {
	case SubcontractFactoryClaimStatusOpen, SubcontractFactoryClaimStatusAcknowledged:
		return true
	default:
		return false
	}
}

func (c SubcontractFactoryClaim) IsOverdue(now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}

	return c.BlocksFinalPayment() && now.UTC().After(c.DueAt)
}

func (c SubcontractFactoryClaim) Clone() SubcontractFactoryClaim {
	clone := c
	clone.Evidence = append([]SubcontractFactoryClaimEvidence(nil), c.Evidence...)
	return clone
}

func NormalizeSubcontractFactoryClaimStatus(status SubcontractFactoryClaimStatus) SubcontractFactoryClaimStatus {
	return SubcontractFactoryClaimStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSubcontractFactoryClaimStatus(status SubcontractFactoryClaimStatus) bool {
	switch NormalizeSubcontractFactoryClaimStatus(status) {
	case SubcontractFactoryClaimStatusOpen,
		SubcontractFactoryClaimStatusAcknowledged,
		SubcontractFactoryClaimStatusResolved,
		SubcontractFactoryClaimStatusClosed,
		SubcontractFactoryClaimStatusCancelled:
		return true
	default:
		return false
	}
}

func normalizeSubcontractFactoryClaimSeverity(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "" {
		return "P1"
	}

	return value
}
