package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractPaymentMilestoneRequiredField = errors.New("subcontract payment milestone required field is missing")
var ErrSubcontractPaymentMilestoneInvalidKind = errors.New("subcontract payment milestone kind is invalid")
var ErrSubcontractPaymentMilestoneInvalidStatus = errors.New("subcontract payment milestone status is invalid")
var ErrSubcontractPaymentMilestoneInvalidCurrency = errors.New("subcontract payment milestone currency is invalid")
var ErrSubcontractPaymentMilestoneInvalidAmount = errors.New("subcontract payment milestone amount is invalid")
var ErrSubcontractPaymentMilestoneInvalidTransition = errors.New("subcontract payment milestone transition is invalid")
var ErrSubcontractPaymentMilestoneBlocked = errors.New("subcontract payment milestone is blocked")

type SubcontractPaymentMilestoneKind string

const (
	SubcontractPaymentMilestoneKindDeposit      SubcontractPaymentMilestoneKind = "deposit"
	SubcontractPaymentMilestoneKindFinalPayment SubcontractPaymentMilestoneKind = "final_payment"
)

type SubcontractPaymentMilestoneStatus string

const (
	SubcontractPaymentMilestoneStatusPending   SubcontractPaymentMilestoneStatus = "pending"
	SubcontractPaymentMilestoneStatusRecorded  SubcontractPaymentMilestoneStatus = "recorded"
	SubcontractPaymentMilestoneStatusReady     SubcontractPaymentMilestoneStatus = "ready"
	SubcontractPaymentMilestoneStatusBlocked   SubcontractPaymentMilestoneStatus = "blocked"
	SubcontractPaymentMilestoneStatusCancelled SubcontractPaymentMilestoneStatus = "cancelled"
)

type SubcontractPaymentMilestone struct {
	ID                  string
	OrgID               string
	MilestoneNo         string
	SubcontractOrderID  string
	SubcontractOrderNo  string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	Kind                SubcontractPaymentMilestoneKind
	Status              SubcontractPaymentMilestoneStatus
	Amount              decimal.Decimal
	CurrencyCode        decimal.CurrencyCode
	Note                string
	BlockReason         string
	ApprovedExceptionID string
	RecordedBy          string
	RecordedAt          time.Time
	ReadyBy             string
	ReadyAt             time.Time
	BlockedBy           string
	BlockedAt           time.Time
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
}

type NewSubcontractPaymentMilestoneInput struct {
	ID                  string
	OrgID               string
	MilestoneNo         string
	SubcontractOrderID  string
	SubcontractOrderNo  string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	Kind                SubcontractPaymentMilestoneKind
	Status              SubcontractPaymentMilestoneStatus
	Amount              decimal.Decimal
	CurrencyCode        string
	Note                string
	BlockReason         string
	ApprovedExceptionID string
	RecordedBy          string
	RecordedAt          time.Time
	ReadyBy             string
	ReadyAt             time.Time
	BlockedBy           string
	BlockedAt           time.Time
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
}

func NewSubcontractPaymentMilestone(
	input NewSubcontractPaymentMilestoneInput,
) (SubcontractPaymentMilestone, error) {
	kind := NormalizeSubcontractPaymentMilestoneKind(input.Kind)
	status := NormalizeSubcontractPaymentMilestoneStatus(input.Status)
	if strings.TrimSpace(string(status)) == "" {
		status = SubcontractPaymentMilestoneStatusPending
	}
	currencyCode, err := decimal.NormalizeCurrencyCode(input.CurrencyCode)
	if err != nil || currencyCode != decimal.CurrencyVND {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneInvalidCurrency
	}
	amount, err := decimal.ParseMoneyAmount(input.Amount.String())
	if err != nil || amount.IsNegative() || amount.IsZero() {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneInvalidAmount
	}
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	milestone := SubcontractPaymentMilestone{
		ID:                  strings.TrimSpace(input.ID),
		OrgID:               strings.TrimSpace(input.OrgID),
		MilestoneNo:         strings.ToUpper(strings.TrimSpace(input.MilestoneNo)),
		SubcontractOrderID:  strings.TrimSpace(input.SubcontractOrderID),
		SubcontractOrderNo:  strings.ToUpper(strings.TrimSpace(input.SubcontractOrderNo)),
		FactoryID:           strings.TrimSpace(input.FactoryID),
		FactoryCode:         strings.ToUpper(strings.TrimSpace(input.FactoryCode)),
		FactoryName:         strings.TrimSpace(input.FactoryName),
		Kind:                kind,
		Status:              status,
		Amount:              amount,
		CurrencyCode:        currencyCode,
		Note:                strings.TrimSpace(input.Note),
		BlockReason:         strings.TrimSpace(input.BlockReason),
		ApprovedExceptionID: strings.TrimSpace(input.ApprovedExceptionID),
		RecordedBy:          strings.TrimSpace(input.RecordedBy),
		RecordedAt:          input.RecordedAt.UTC(),
		ReadyBy:             strings.TrimSpace(input.ReadyBy),
		ReadyAt:             input.ReadyAt.UTC(),
		BlockedBy:           strings.TrimSpace(input.BlockedBy),
		BlockedAt:           input.BlockedAt.UTC(),
		CreatedAt:           createdAt.UTC(),
		CreatedBy:           strings.TrimSpace(input.CreatedBy),
		UpdatedAt:           updatedAt.UTC(),
		UpdatedBy:           strings.TrimSpace(input.UpdatedBy),
		Version:             1,
	}
	if err := milestone.Validate(); err != nil {
		return SubcontractPaymentMilestone{}, err
	}

	return milestone, nil
}

func (m SubcontractPaymentMilestone) Validate() error {
	if strings.TrimSpace(m.ID) == "" ||
		strings.TrimSpace(m.OrgID) == "" ||
		strings.TrimSpace(m.MilestoneNo) == "" ||
		strings.TrimSpace(m.SubcontractOrderID) == "" ||
		strings.TrimSpace(m.SubcontractOrderNo) == "" ||
		strings.TrimSpace(m.FactoryID) == "" ||
		strings.TrimSpace(m.FactoryName) == "" ||
		strings.TrimSpace(m.CreatedBy) == "" ||
		m.CreatedAt.IsZero() {
		return ErrSubcontractPaymentMilestoneRequiredField
	}
	if !IsValidSubcontractPaymentMilestoneKind(m.Kind) {
		return ErrSubcontractPaymentMilestoneInvalidKind
	}
	if !IsValidSubcontractPaymentMilestoneStatus(m.Status) {
		return ErrSubcontractPaymentMilestoneInvalidStatus
	}
	if m.CurrencyCode != decimal.CurrencyVND {
		return ErrSubcontractPaymentMilestoneInvalidCurrency
	}
	amount, err := decimal.ParseMoneyAmount(m.Amount.String())
	if err != nil || amount.IsNegative() || amount.IsZero() {
		return ErrSubcontractPaymentMilestoneInvalidAmount
	}
	if m.Status == SubcontractPaymentMilestoneStatusRecorded &&
		(strings.TrimSpace(m.RecordedBy) == "" || m.RecordedAt.IsZero()) {
		return ErrSubcontractPaymentMilestoneRequiredField
	}
	if m.Status == SubcontractPaymentMilestoneStatusReady &&
		(strings.TrimSpace(m.ReadyBy) == "" || m.ReadyAt.IsZero()) {
		return ErrSubcontractPaymentMilestoneRequiredField
	}
	if m.Status == SubcontractPaymentMilestoneStatusBlocked && strings.TrimSpace(m.BlockReason) == "" {
		return ErrSubcontractPaymentMilestoneRequiredField
	}
	if m.Status == SubcontractPaymentMilestoneStatusBlocked &&
		(strings.TrimSpace(m.BlockedBy) == "" || m.BlockedAt.IsZero()) {
		return ErrSubcontractPaymentMilestoneRequiredField
	}

	return nil
}

func (m SubcontractPaymentMilestone) Record(actorID string, recordedAt time.Time) (SubcontractPaymentMilestone, error) {
	if NormalizeSubcontractPaymentMilestoneKind(m.Kind) != SubcontractPaymentMilestoneKindDeposit ||
		NormalizeSubcontractPaymentMilestoneStatus(m.Status) != SubcontractPaymentMilestoneStatusPending {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneRequiredField
	}
	if recordedAt.IsZero() {
		recordedAt = time.Now().UTC()
	}
	updated := m.Clone()
	updated.Status = SubcontractPaymentMilestoneStatusRecorded
	updated.RecordedBy = actorID
	updated.RecordedAt = recordedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = recordedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (m SubcontractPaymentMilestone) MarkReady(actorID string, readyAt time.Time) (SubcontractPaymentMilestone, error) {
	status := NormalizeSubcontractPaymentMilestoneStatus(m.Status)
	if NormalizeSubcontractPaymentMilestoneKind(m.Kind) != SubcontractPaymentMilestoneKindFinalPayment ||
		(status != SubcontractPaymentMilestoneStatusPending && status != SubcontractPaymentMilestoneStatusBlocked) {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneInvalidTransition
	}
	if status == SubcontractPaymentMilestoneStatusBlocked && strings.TrimSpace(m.ApprovedExceptionID) == "" {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneBlocked
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneRequiredField
	}
	if readyAt.IsZero() {
		readyAt = time.Now().UTC()
	}
	updated := m.Clone()
	updated.Status = SubcontractPaymentMilestoneStatusReady
	updated.BlockReason = ""
	updated.ReadyBy = actorID
	updated.ReadyAt = readyAt.UTC()
	updated.BlockedBy = ""
	updated.BlockedAt = time.Time{}
	updated.UpdatedBy = actorID
	updated.UpdatedAt = readyAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (m SubcontractPaymentMilestone) Block(actorID string, reason string, blockedAt time.Time) (SubcontractPaymentMilestone, error) {
	if NormalizeSubcontractPaymentMilestoneKind(m.Kind) != SubcontractPaymentMilestoneKindFinalPayment ||
		NormalizeSubcontractPaymentMilestoneStatus(m.Status) != SubcontractPaymentMilestoneStatusPending {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	reason = strings.TrimSpace(reason)
	if actorID == "" || reason == "" {
		return SubcontractPaymentMilestone{}, ErrSubcontractPaymentMilestoneRequiredField
	}
	if blockedAt.IsZero() {
		blockedAt = time.Now().UTC()
	}
	updated := m.Clone()
	updated.Status = SubcontractPaymentMilestoneStatusBlocked
	updated.BlockReason = reason
	updated.BlockedBy = actorID
	updated.BlockedAt = blockedAt.UTC()
	updated.UpdatedBy = actorID
	updated.UpdatedAt = blockedAt.UTC()
	updated.Version++

	return updated, updated.Validate()
}

func (m SubcontractPaymentMilestone) Clone() SubcontractPaymentMilestone {
	return m
}

func (m SubcontractPaymentMilestone) BlocksFinalPayment() bool {
	if NormalizeSubcontractPaymentMilestoneKind(m.Kind) != SubcontractPaymentMilestoneKindFinalPayment {
		return false
	}
	switch NormalizeSubcontractPaymentMilestoneStatus(m.Status) {
	case SubcontractPaymentMilestoneStatusPending, SubcontractPaymentMilestoneStatusBlocked:
		return true
	default:
		return false
	}
}

func NormalizeSubcontractPaymentMilestoneKind(
	kind SubcontractPaymentMilestoneKind,
) SubcontractPaymentMilestoneKind {
	return SubcontractPaymentMilestoneKind(strings.ToLower(strings.TrimSpace(string(kind))))
}

func IsValidSubcontractPaymentMilestoneKind(kind SubcontractPaymentMilestoneKind) bool {
	switch NormalizeSubcontractPaymentMilestoneKind(kind) {
	case SubcontractPaymentMilestoneKindDeposit, SubcontractPaymentMilestoneKindFinalPayment:
		return true
	default:
		return false
	}
}

func NormalizeSubcontractPaymentMilestoneStatus(
	status SubcontractPaymentMilestoneStatus,
) SubcontractPaymentMilestoneStatus {
	return SubcontractPaymentMilestoneStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSubcontractPaymentMilestoneStatus(status SubcontractPaymentMilestoneStatus) bool {
	switch NormalizeSubcontractPaymentMilestoneStatus(status) {
	case SubcontractPaymentMilestoneStatusPending,
		SubcontractPaymentMilestoneStatusRecorded,
		SubcontractPaymentMilestoneStatusReady,
		SubcontractPaymentMilestoneStatusBlocked,
		SubcontractPaymentMilestoneStatusCancelled:
		return true
	default:
		return false
	}
}
