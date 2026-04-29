package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSubcontractOrderRequiredField = errors.New("subcontract order required field is missing")
var ErrSubcontractOrderInvalidStatus = errors.New("subcontract order status is invalid")
var ErrSubcontractOrderInvalidTransition = errors.New("subcontract order status transition is invalid")
var ErrSubcontractOrderTransitionActorRequired = errors.New("subcontract order status transition actor is required")
var ErrSubcontractOrderInvalidCurrency = errors.New("subcontract order currency is invalid")
var ErrSubcontractOrderInvalidQuantity = errors.New("subcontract order quantity is invalid")
var ErrSubcontractOrderInvalidAmount = errors.New("subcontract order amount is invalid")
var ErrSubcontractOrderSampleApprovalRequired = errors.New("subcontract order sample approval is required")
var ErrSubcontractOrderMaterialLineNotFound = errors.New("subcontract order material line is not found")
var ErrSubcontractOrderDuplicateMaterialLine = errors.New("subcontract order material line is duplicated")

type SubcontractOrderStatus string

const (
	SubcontractOrderStatusDraft                 SubcontractOrderStatus = "draft"
	SubcontractOrderStatusSubmitted             SubcontractOrderStatus = "submitted"
	SubcontractOrderStatusApproved              SubcontractOrderStatus = "approved"
	SubcontractOrderStatusFactoryConfirmed      SubcontractOrderStatus = "factory_confirmed"
	SubcontractOrderStatusDepositRecorded       SubcontractOrderStatus = "deposit_recorded"
	SubcontractOrderStatusMaterialsIssued       SubcontractOrderStatus = "materials_issued_to_factory"
	SubcontractOrderStatusSampleSubmitted       SubcontractOrderStatus = "sample_submitted"
	SubcontractOrderStatusSampleApproved        SubcontractOrderStatus = "sample_approved"
	SubcontractOrderStatusSampleRejected        SubcontractOrderStatus = "sample_rejected"
	SubcontractOrderStatusMassProductionStarted SubcontractOrderStatus = "mass_production_started"
	SubcontractOrderStatusFinishedGoodsReceived SubcontractOrderStatus = "finished_goods_received"
	SubcontractOrderStatusQCInProgress          SubcontractOrderStatus = "qc_in_progress"
	SubcontractOrderStatusAccepted              SubcontractOrderStatus = "accepted"
	SubcontractOrderStatusRejectedFactoryIssue  SubcontractOrderStatus = "rejected_with_factory_issue"
	SubcontractOrderStatusFinalPaymentReady     SubcontractOrderStatus = "final_payment_ready"
	SubcontractOrderStatusClosed                SubcontractOrderStatus = "closed"
	SubcontractOrderStatusCancelled             SubcontractOrderStatus = "cancelled"
)

type SubcontractOrder struct {
	ID                  string
	OrgID               string
	OrderNo             string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	FinishedItemID      string
	FinishedSKUCode     string
	FinishedItemName    string
	PlannedQty          decimal.Decimal
	ReceivedQty         decimal.Decimal
	AcceptedQty         decimal.Decimal
	RejectedQty         decimal.Decimal
	UOMCode             decimal.UOMCode
	BasePlannedQty      decimal.Decimal
	BaseReceivedQty     decimal.Decimal
	BaseAcceptedQty     decimal.Decimal
	BaseRejectedQty     decimal.Decimal
	BaseUOMCode         decimal.UOMCode
	ConversionFactor    decimal.Decimal
	CurrencyCode        decimal.CurrencyCode
	EstimatedCostAmount decimal.Decimal
	DepositAmount       decimal.Decimal
	SpecSummary         string
	SampleRequired      bool
	ClaimWindowDays     int
	TargetStartDate     string
	ExpectedReceiptDate string
	Status              SubcontractOrderStatus
	MaterialLines       []SubcontractMaterialLine
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
	Version             int
	CancelReason        string
	SampleRejectReason  string
	FactoryIssueReason  string

	SubmittedAt             time.Time
	SubmittedBy             string
	ApprovedAt              time.Time
	ApprovedBy              string
	FactoryConfirmedAt      time.Time
	FactoryConfirmedBy      string
	DepositRecordedAt       time.Time
	DepositRecordedBy       string
	MaterialsIssuedAt       time.Time
	MaterialsIssuedBy       string
	SampleSubmittedAt       time.Time
	SampleSubmittedBy       string
	SampleApprovedAt        time.Time
	SampleApprovedBy        string
	SampleRejectedAt        time.Time
	SampleRejectedBy        string
	MassProductionStartedAt time.Time
	MassProductionStartedBy string
	FinishedGoodsReceivedAt time.Time
	FinishedGoodsReceivedBy string
	QCStartedAt             time.Time
	QCStartedBy             string
	AcceptedAt              time.Time
	AcceptedBy              string
	RejectedFactoryIssueAt  time.Time
	RejectedFactoryIssueBy  string
	FinalPaymentReadyAt     time.Time
	FinalPaymentReadyBy     string
	ClosedAt                time.Time
	ClosedBy                string
	CancelledAt             time.Time
	CancelledBy             string
}

type SubcontractMaterialLine struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	PlannedQty       decimal.Decimal
	IssuedQty        decimal.Decimal
	UOMCode          decimal.UOMCode
	BasePlannedQty   decimal.Decimal
	BaseIssuedQty    decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	ConversionFactor decimal.Decimal
	UnitCost         decimal.Decimal
	CurrencyCode     decimal.CurrencyCode
	LineCostAmount   decimal.Decimal
	LotTraceRequired bool
	Note             string
}

type NewSubcontractOrderDocumentInput struct {
	ID                  string
	OrgID               string
	OrderNo             string
	FactoryID           string
	FactoryCode         string
	FactoryName         string
	FinishedItemID      string
	FinishedSKUCode     string
	FinishedItemName    string
	PlannedQty          decimal.Decimal
	ReceivedQty         decimal.Decimal
	AcceptedQty         decimal.Decimal
	RejectedQty         decimal.Decimal
	UOMCode             string
	BasePlannedQty      decimal.Decimal
	BaseReceivedQty     decimal.Decimal
	BaseAcceptedQty     decimal.Decimal
	BaseRejectedQty     decimal.Decimal
	BaseUOMCode         string
	ConversionFactor    decimal.Decimal
	CurrencyCode        string
	DepositAmount       decimal.Decimal
	SpecSummary         string
	SampleRequired      bool
	ClaimWindowDays     int
	TargetStartDate     string
	ExpectedReceiptDate string
	MaterialLines       []NewSubcontractMaterialLineInput
	CreatedAt           time.Time
	CreatedBy           string
	UpdatedAt           time.Time
	UpdatedBy           string
}

type NewSubcontractMaterialLineInput struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	PlannedQty       decimal.Decimal
	IssuedQty        decimal.Decimal
	UOMCode          string
	BasePlannedQty   decimal.Decimal
	BaseIssuedQty    decimal.Decimal
	BaseUOMCode      string
	ConversionFactor decimal.Decimal
	UnitCost         decimal.Decimal
	CurrencyCode     string
	LotTraceRequired bool
	Note             string
}

type IssueSubcontractMaterialsInput struct {
	Lines     []IssueSubcontractMaterialLineInput
	ActorID   string
	ChangedAt time.Time
}

type IssueSubcontractMaterialLineInput struct {
	OrderMaterialLineID string
	IssueQty            decimal.Decimal
	UOMCode             string
	BaseIssueQty        decimal.Decimal
	BaseUOMCode         string
	ConversionFactor    decimal.Decimal
}

type ReceiveSubcontractFinishedGoodsInput struct {
	ReceiptQty       decimal.Decimal
	UOMCode          string
	BaseReceiptQty   decimal.Decimal
	BaseUOMCode      string
	ConversionFactor decimal.Decimal
	ActorID          string
	ChangedAt        time.Time
}

var subcontractOrderTransitions = map[SubcontractOrderStatus][]SubcontractOrderStatus{
	SubcontractOrderStatusDraft: {
		SubcontractOrderStatusSubmitted,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusSubmitted: {
		SubcontractOrderStatusApproved,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusApproved: {
		SubcontractOrderStatusFactoryConfirmed,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusFactoryConfirmed: {
		SubcontractOrderStatusDepositRecorded,
		SubcontractOrderStatusMaterialsIssued,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusDepositRecorded: {
		SubcontractOrderStatusMaterialsIssued,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusMaterialsIssued: {
		SubcontractOrderStatusSampleSubmitted,
		SubcontractOrderStatusMassProductionStarted,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusSampleSubmitted: {
		SubcontractOrderStatusSampleApproved,
		SubcontractOrderStatusSampleRejected,
	},
	SubcontractOrderStatusSampleRejected: {
		SubcontractOrderStatusSampleSubmitted,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusSampleApproved: {
		SubcontractOrderStatusMassProductionStarted,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusMassProductionStarted: {
		SubcontractOrderStatusFinishedGoodsReceived,
		SubcontractOrderStatusCancelled,
	},
	SubcontractOrderStatusFinishedGoodsReceived: {
		SubcontractOrderStatusQCInProgress,
	},
	SubcontractOrderStatusQCInProgress: {
		SubcontractOrderStatusAccepted,
		SubcontractOrderStatusRejectedFactoryIssue,
	},
	SubcontractOrderStatusAccepted: {
		SubcontractOrderStatusFinalPaymentReady,
		SubcontractOrderStatusClosed,
	},
	SubcontractOrderStatusFinalPaymentReady: {
		SubcontractOrderStatusClosed,
	},
}

func NewSubcontractOrder(status SubcontractOrderStatus) (SubcontractOrder, error) {
	status = NormalizeSubcontractOrderStatus(status)
	if status == "" {
		status = SubcontractOrderStatusDraft
	}
	if !IsValidSubcontractOrderStatus(status) {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidStatus
	}

	return SubcontractOrder{Status: status}, nil
}

func NewSubcontractOrderDocument(input NewSubcontractOrderDocumentInput) (SubcontractOrder, error) {
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
	currencyCode, err := normalizeSubcontractOrderCurrency(input.CurrencyCode)
	if err != nil {
		return SubcontractOrder{}, err
	}
	plannedQty, basePlannedQty, uomCode, baseUOMCode, conversionFactor, err := normalizeSubcontractOrderQuantitySet(
		input.PlannedQty,
		input.BasePlannedQty,
		input.UOMCode,
		input.BaseUOMCode,
		input.ConversionFactor,
		true,
	)
	if err != nil {
		return SubcontractOrder{}, err
	}
	receivedQty, baseReceivedQty, err := normalizeSubcontractOrderProgressQuantity(input.ReceivedQty, input.BaseReceivedQty, baseUOMCode, uomCode, conversionFactor)
	if err != nil {
		return SubcontractOrder{}, err
	}
	acceptedQty, baseAcceptedQty, err := normalizeSubcontractOrderProgressQuantity(input.AcceptedQty, input.BaseAcceptedQty, baseUOMCode, uomCode, conversionFactor)
	if err != nil {
		return SubcontractOrder{}, err
	}
	rejectedQty, baseRejectedQty, err := normalizeSubcontractOrderProgressQuantity(input.RejectedQty, input.BaseRejectedQty, baseUOMCode, uomCode, conversionFactor)
	if err != nil {
		return SubcontractOrder{}, err
	}
	depositAmount := decimal.MustMoneyAmount("0")
	if strings.TrimSpace(input.DepositAmount.String()) != "" {
		depositAmount, err = normalizeSubcontractOrderNonNegativeMoney(input.DepositAmount)
		if err != nil {
			return SubcontractOrder{}, err
		}
	}
	claimWindowDays := input.ClaimWindowDays
	if claimWindowDays == 0 {
		claimWindowDays = 7
	}

	order := SubcontractOrder{
		ID:                  strings.TrimSpace(input.ID),
		OrgID:               strings.TrimSpace(input.OrgID),
		OrderNo:             strings.ToUpper(strings.TrimSpace(input.OrderNo)),
		FactoryID:           strings.TrimSpace(input.FactoryID),
		FactoryCode:         strings.ToUpper(strings.TrimSpace(input.FactoryCode)),
		FactoryName:         strings.TrimSpace(input.FactoryName),
		FinishedItemID:      strings.TrimSpace(input.FinishedItemID),
		FinishedSKUCode:     strings.ToUpper(strings.TrimSpace(input.FinishedSKUCode)),
		FinishedItemName:    strings.TrimSpace(input.FinishedItemName),
		PlannedQty:          plannedQty,
		ReceivedQty:         receivedQty,
		AcceptedQty:         acceptedQty,
		RejectedQty:         rejectedQty,
		UOMCode:             uomCode,
		BasePlannedQty:      basePlannedQty,
		BaseReceivedQty:     baseReceivedQty,
		BaseAcceptedQty:     baseAcceptedQty,
		BaseRejectedQty:     baseRejectedQty,
		BaseUOMCode:         baseUOMCode,
		ConversionFactor:    conversionFactor,
		CurrencyCode:        currencyCode,
		DepositAmount:       depositAmount,
		SpecSummary:         strings.TrimSpace(input.SpecSummary),
		SampleRequired:      input.SampleRequired,
		ClaimWindowDays:     claimWindowDays,
		TargetStartDate:     strings.TrimSpace(input.TargetStartDate),
		ExpectedReceiptDate: strings.TrimSpace(input.ExpectedReceiptDate),
		Status:              SubcontractOrderStatusDraft,
		MaterialLines:       make([]SubcontractMaterialLine, 0, len(input.MaterialLines)),
		CreatedAt:           createdAt.UTC(),
		CreatedBy:           strings.TrimSpace(input.CreatedBy),
		UpdatedAt:           updatedAt.UTC(),
		UpdatedBy:           updatedBy,
		Version:             1,
	}
	for index, lineInput := range input.MaterialLines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		if lineInput.CurrencyCode == "" {
			lineInput.CurrencyCode = currencyCode.String()
		}
		line, err := NewSubcontractMaterialLine(lineInput)
		if err != nil {
			return SubcontractOrder{}, err
		}
		order.MaterialLines = append(order.MaterialLines, line)
	}
	if err := order.recalculateEstimatedCost(); err != nil {
		return SubcontractOrder{}, err
	}
	if err := order.Validate(); err != nil {
		return SubcontractOrder{}, err
	}

	return order, nil
}

func NewSubcontractMaterialLine(input NewSubcontractMaterialLineInput) (SubcontractMaterialLine, error) {
	plannedQty, basePlannedQty, uomCode, baseUOMCode, conversionFactor, err := normalizeSubcontractOrderQuantitySet(
		input.PlannedQty,
		input.BasePlannedQty,
		input.UOMCode,
		input.BaseUOMCode,
		input.ConversionFactor,
		true,
	)
	if err != nil {
		return SubcontractMaterialLine{}, err
	}
	issuedQty, baseIssuedQty, err := normalizeSubcontractOrderProgressQuantity(input.IssuedQty, input.BaseIssuedQty, baseUOMCode, uomCode, conversionFactor)
	if err != nil {
		return SubcontractMaterialLine{}, err
	}
	unitCost, err := normalizeSubcontractOrderNonNegativeUnitCost(input.UnitCost)
	if err != nil {
		return SubcontractMaterialLine{}, err
	}
	currencyCode, err := normalizeSubcontractOrderCurrency(input.CurrencyCode)
	if err != nil {
		return SubcontractMaterialLine{}, err
	}
	lineCost, err := subcontractOrderMoneyFromQuantityUnitCost(plannedQty, unitCost)
	if err != nil {
		return SubcontractMaterialLine{}, err
	}

	line := SubcontractMaterialLine{
		ID:               strings.TrimSpace(input.ID),
		LineNo:           input.LineNo,
		ItemID:           strings.TrimSpace(input.ItemID),
		SKUCode:          strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		ItemName:         strings.TrimSpace(input.ItemName),
		PlannedQty:       plannedQty,
		IssuedQty:        issuedQty,
		UOMCode:          uomCode,
		BasePlannedQty:   basePlannedQty,
		BaseIssuedQty:    baseIssuedQty,
		BaseUOMCode:      baseUOMCode,
		ConversionFactor: conversionFactor,
		UnitCost:         unitCost,
		CurrencyCode:     currencyCode,
		LineCostAmount:   lineCost,
		LotTraceRequired: input.LotTraceRequired,
		Note:             strings.TrimSpace(input.Note),
	}
	if err := line.Validate(); err != nil {
		return SubcontractMaterialLine{}, err
	}

	return line, nil
}

func (o SubcontractOrder) Validate() error {
	if strings.TrimSpace(o.ID) == "" ||
		strings.TrimSpace(o.OrgID) == "" ||
		strings.TrimSpace(o.OrderNo) == "" ||
		strings.TrimSpace(o.FactoryID) == "" ||
		strings.TrimSpace(o.FactoryName) == "" ||
		strings.TrimSpace(o.FinishedItemID) == "" ||
		strings.TrimSpace(o.FinishedSKUCode) == "" ||
		strings.TrimSpace(o.FinishedItemName) == "" ||
		strings.TrimSpace(o.CreatedBy) == "" ||
		len(o.MaterialLines) == 0 {
		return ErrSubcontractOrderRequiredField
	}
	if !IsValidSubcontractOrderStatus(o.Status) {
		return ErrSubcontractOrderInvalidStatus
	}
	if o.CurrencyCode != decimal.CurrencyVND {
		return ErrSubcontractOrderInvalidCurrency
	}
	if o.ClaimWindowDays < 3 || o.ClaimWindowDays > 7 {
		return ErrSubcontractOrderRequiredField
	}
	for _, quantity := range []decimal.Decimal{o.PlannedQty, o.BasePlannedQty, o.ConversionFactor} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrSubcontractOrderInvalidQuantity
		}
	}
	for _, quantity := range []decimal.Decimal{o.ReceivedQty, o.AcceptedQty, o.RejectedQty, o.BaseReceivedQty, o.BaseAcceptedQty, o.BaseRejectedQty} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() {
			return ErrSubcontractOrderInvalidQuantity
		}
	}
	if err := validateSubcontractOrderProgressQuantity(o.ReceivedQty, o.PlannedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderProgressQuantity(o.AcceptedQty, o.ReceivedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderProgressQuantity(o.RejectedQty, o.ReceivedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderQuantitySum(o.AcceptedQty, o.RejectedQty, o.ReceivedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderProgressQuantity(o.BaseReceivedQty, o.BasePlannedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderProgressQuantity(o.BaseAcceptedQty, o.BaseReceivedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderProgressQuantity(o.BaseRejectedQty, o.BaseReceivedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderQuantitySum(o.BaseAcceptedQty, o.BaseRejectedQty, o.BaseReceivedQty); err != nil {
		return err
	}
	for _, amount := range []decimal.Decimal{o.EstimatedCostAmount, o.DepositAmount} {
		value, err := decimal.ParseMoneyAmount(amount.String())
		if err != nil || value.IsNegative() {
			return ErrSubcontractOrderInvalidAmount
		}
	}
	seenLineNo := map[int]struct{}{}
	for _, line := range o.MaterialLines {
		if _, exists := seenLineNo[line.LineNo]; exists {
			return ErrSubcontractOrderRequiredField
		}
		seenLineNo[line.LineNo] = struct{}{}
		if line.CurrencyCode != o.CurrencyCode {
			return ErrSubcontractOrderInvalidCurrency
		}
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l SubcontractMaterialLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKUCode) == "" ||
		strings.TrimSpace(l.ItemName) == "" {
		return ErrSubcontractOrderRequiredField
	}
	if l.CurrencyCode != decimal.CurrencyVND {
		return ErrSubcontractOrderInvalidCurrency
	}
	for _, quantity := range []decimal.Decimal{l.PlannedQty, l.BasePlannedQty, l.ConversionFactor} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrSubcontractOrderInvalidQuantity
		}
	}
	for _, quantity := range []decimal.Decimal{l.IssuedQty, l.BaseIssuedQty} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() {
			return ErrSubcontractOrderInvalidQuantity
		}
	}
	if err := validateSubcontractOrderProgressQuantity(l.IssuedQty, l.PlannedQty); err != nil {
		return err
	}
	if err := validateSubcontractOrderProgressQuantity(l.BaseIssuedQty, l.BasePlannedQty); err != nil {
		return err
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrSubcontractOrderInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrSubcontractOrderInvalidQuantity
	}
	if _, err := normalizeSubcontractOrderNonNegativeUnitCost(l.UnitCost); err != nil {
		return err
	}
	value, err := decimal.ParseMoneyAmount(l.LineCostAmount.String())
	if err != nil || value.IsNegative() {
		return ErrSubcontractOrderInvalidAmount
	}

	return nil
}

func (o SubcontractOrder) Submit(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusSubmitted, actorID, changedAt)
}

func (o SubcontractOrder) Approve(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusApproved, actorID, changedAt)
}

func (o SubcontractOrder) ConfirmFactory(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusFactoryConfirmed, actorID, changedAt)
}

func (o SubcontractOrder) RecordDeposit(actorID string, amount decimal.Decimal, changedAt time.Time) (SubcontractOrder, error) {
	updated, err := o.TransitionTo(SubcontractOrderStatusDepositRecorded, actorID, changedAt)
	if err != nil {
		return SubcontractOrder{}, err
	}
	depositAmount, err := normalizeSubcontractOrderNonNegativeMoney(amount)
	if err != nil {
		return SubcontractOrder{}, err
	}
	updated.DepositAmount = depositAmount
	if err := updated.Validate(); err != nil {
		return SubcontractOrder{}, err
	}

	return updated, nil
}

func (o SubcontractOrder) MarkMaterialsIssued(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusMaterialsIssued, actorID, changedAt)
}

func (o SubcontractOrder) IssueMaterials(input IssueSubcontractMaterialsInput) (SubcontractOrder, error) {
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return SubcontractOrder{}, ErrSubcontractOrderTransitionActorRequired
	}
	if len(input.Lines) == 0 {
		return SubcontractOrder{}, ErrSubcontractOrderRequiredField
	}
	status := NormalizeSubcontractOrderStatus(o.Status)
	if status != SubcontractOrderStatusFactoryConfirmed && status != SubcontractOrderStatusDepositRecorded {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidTransition
	}
	changedAt := input.ChangedAt
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := o.Clone()
	seen := map[string]struct{}{}
	for _, issueLine := range input.Lines {
		lineID := strings.TrimSpace(issueLine.OrderMaterialLineID)
		if lineID == "" {
			return SubcontractOrder{}, ErrSubcontractOrderRequiredField
		}
		if _, exists := seen[lineID]; exists {
			return SubcontractOrder{}, ErrSubcontractOrderDuplicateMaterialLine
		}
		seen[lineID] = struct{}{}

		lineIndex := -1
		for index, candidate := range updated.MaterialLines {
			if candidate.ID == lineID {
				lineIndex = index
				break
			}
		}
		if lineIndex < 0 {
			return SubcontractOrder{}, ErrSubcontractOrderMaterialLineNotFound
		}

		line := updated.MaterialLines[lineIndex]
		uomCode := subcontractOrderFirstNonBlank(issueLine.UOMCode, line.UOMCode.String())
		baseUOMCode := subcontractOrderFirstNonBlank(issueLine.BaseUOMCode, line.BaseUOMCode.String())
		conversionFactor := issueLine.ConversionFactor
		if strings.TrimSpace(conversionFactor.String()) == "" {
			conversionFactor = line.ConversionFactor
		}
		issueQty, baseIssueQty, normalizedUOMCode, normalizedBaseUOMCode, normalizedConversionFactor, err := normalizeSubcontractOrderQuantitySet(
			issueLine.IssueQty,
			issueLine.BaseIssueQty,
			uomCode,
			baseUOMCode,
			conversionFactor,
			true,
		)
		if err != nil {
			return SubcontractOrder{}, err
		}
		if normalizedUOMCode != line.UOMCode || normalizedBaseUOMCode != line.BaseUOMCode || normalizedConversionFactor != line.ConversionFactor {
			return SubcontractOrder{}, ErrSubcontractOrderInvalidQuantity
		}

		nextIssuedQty, err := decimal.AddQuantity(line.IssuedQty, issueQty)
		if err != nil {
			return SubcontractOrder{}, ErrSubcontractOrderInvalidQuantity
		}
		nextBaseIssuedQty, err := decimal.AddQuantity(line.BaseIssuedQty, baseIssueQty)
		if err != nil {
			return SubcontractOrder{}, ErrSubcontractOrderInvalidQuantity
		}
		if err := validateSubcontractOrderProgressQuantity(nextIssuedQty, line.PlannedQty); err != nil {
			return SubcontractOrder{}, err
		}
		if err := validateSubcontractOrderProgressQuantity(nextBaseIssuedQty, line.BasePlannedQty); err != nil {
			return SubcontractOrder{}, err
		}

		line.IssuedQty = nextIssuedQty
		line.BaseIssuedQty = nextBaseIssuedQty
		updated.MaterialLines[lineIndex] = line
	}

	if allSubcontractMaterialLinesIssued(updated.MaterialLines) {
		return updated.TransitionTo(SubcontractOrderStatusMaterialsIssued, actorID, changedAt)
	}

	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if updated.Version > 0 {
		updated.Version++
	}
	if err := updated.Validate(); err != nil {
		return SubcontractOrder{}, err
	}

	return updated, nil
}

func (o SubcontractOrder) SubmitSample(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	submitted, err := o.TransitionTo(SubcontractOrderStatusSampleSubmitted, actorID, changedAt)
	if err != nil {
		return SubcontractOrder{}, err
	}
	submitted.SampleRejectReason = ""

	return submitted, nil
}

func (o SubcontractOrder) ApproveSample(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusSampleApproved, actorID, changedAt)
}

func (o SubcontractOrder) RejectSample(actorID string, reason string, changedAt time.Time) (SubcontractOrder, error) {
	if strings.TrimSpace(reason) == "" {
		return SubcontractOrder{}, ErrSubcontractOrderRequiredField
	}
	rejected, err := o.TransitionTo(SubcontractOrderStatusSampleRejected, actorID, changedAt)
	if err != nil {
		return SubcontractOrder{}, err
	}
	rejected.SampleRejectReason = strings.TrimSpace(reason)

	return rejected, nil
}

func (o SubcontractOrder) StartMassProduction(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	if o.SampleRequired && NormalizeSubcontractOrderStatus(o.Status) == SubcontractOrderStatusMaterialsIssued {
		return SubcontractOrder{}, ErrSubcontractOrderSampleApprovalRequired
	}

	return o.TransitionTo(SubcontractOrderStatusMassProductionStarted, actorID, changedAt)
}

func (o SubcontractOrder) MarkFinishedGoodsReceived(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusFinishedGoodsReceived, actorID, changedAt)
}

func (o SubcontractOrder) ReceiveFinishedGoods(input ReceiveSubcontractFinishedGoodsInput) (SubcontractOrder, error) {
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return SubcontractOrder{}, ErrSubcontractOrderTransitionActorRequired
	}
	status := NormalizeSubcontractOrderStatus(o.Status)
	if status != SubcontractOrderStatusMassProductionStarted && status != SubcontractOrderStatusFinishedGoodsReceived {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidTransition
	}
	changedAt := input.ChangedAt
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	uomCode := subcontractOrderFirstNonBlank(input.UOMCode, o.UOMCode.String())
	baseUOMCode := subcontractOrderFirstNonBlank(input.BaseUOMCode, o.BaseUOMCode.String())
	conversionFactor := input.ConversionFactor
	if strings.TrimSpace(conversionFactor.String()) == "" {
		conversionFactor = o.ConversionFactor
	}
	receiptQty, baseReceiptQty, normalizedUOMCode, normalizedBaseUOMCode, normalizedConversionFactor, err := normalizeSubcontractOrderQuantitySet(
		input.ReceiptQty,
		input.BaseReceiptQty,
		uomCode,
		baseUOMCode,
		conversionFactor,
		true,
	)
	if err != nil {
		return SubcontractOrder{}, err
	}
	if normalizedUOMCode != o.UOMCode || normalizedBaseUOMCode != o.BaseUOMCode || normalizedConversionFactor != o.ConversionFactor {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidQuantity
	}

	updated := o.Clone()
	if status == SubcontractOrderStatusMassProductionStarted {
		updated, err = o.TransitionTo(SubcontractOrderStatusFinishedGoodsReceived, actorID, changedAt)
		if err != nil {
			return SubcontractOrder{}, err
		}
	} else {
		updated.UpdatedAt = changedAt.UTC()
		updated.UpdatedBy = actorID
		if updated.Version > 0 {
			updated.Version++
		}
	}

	nextReceivedQty, err := decimal.AddQuantity(updated.ReceivedQty, receiptQty)
	if err != nil {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidQuantity
	}
	nextBaseReceivedQty, err := decimal.AddQuantity(updated.BaseReceivedQty, baseReceiptQty)
	if err != nil {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidQuantity
	}
	if err := validateSubcontractOrderProgressQuantity(nextReceivedQty, updated.PlannedQty); err != nil {
		return SubcontractOrder{}, err
	}
	if err := validateSubcontractOrderProgressQuantity(nextBaseReceivedQty, updated.BasePlannedQty); err != nil {
		return SubcontractOrder{}, err
	}

	updated.ReceivedQty = nextReceivedQty
	updated.BaseReceivedQty = nextBaseReceivedQty
	if err := updated.Validate(); err != nil {
		return SubcontractOrder{}, err
	}

	return updated, nil
}

func (o SubcontractOrder) StartQC(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusQCInProgress, actorID, changedAt)
}

func (o SubcontractOrder) Accept(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusAccepted, actorID, changedAt)
}

func (o SubcontractOrder) AcceptFinishedGoods(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	current := o
	status := NormalizeSubcontractOrderStatus(current.Status)
	if status == SubcontractOrderStatusFinishedGoodsReceived {
		next, err := current.StartQC(actorID, changedAt)
		if err != nil {
			return SubcontractOrder{}, err
		}
		current = next
	}

	accepted, err := current.Accept(actorID, changedAt)
	if err != nil {
		return SubcontractOrder{}, err
	}
	accepted.AcceptedQty = accepted.ReceivedQty
	accepted.BaseAcceptedQty = accepted.BaseReceivedQty
	accepted.RejectedQty = decimal.MustQuantity("0")
	accepted.BaseRejectedQty = decimal.MustQuantity("0")
	if err := accepted.Validate(); err != nil {
		return SubcontractOrder{}, err
	}

	return accepted, nil
}

func (o SubcontractOrder) RejectWithFactoryIssue(actorID string, reason string, changedAt time.Time) (SubcontractOrder, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return SubcontractOrder{}, ErrSubcontractOrderRequiredField
	}
	rejected, err := o.TransitionTo(SubcontractOrderStatusRejectedFactoryIssue, actorID, changedAt)
	if err != nil {
		return SubcontractOrder{}, err
	}
	rejected.FactoryIssueReason = reason

	return rejected, nil
}

func (o SubcontractOrder) RejectFinishedGoodsWithFactoryIssue(actorID string, reason string, changedAt time.Time) (SubcontractOrder, error) {
	current := o
	status := NormalizeSubcontractOrderStatus(current.Status)
	if status == SubcontractOrderStatusFinishedGoodsReceived {
		next, err := current.StartQC(actorID, changedAt)
		if err != nil {
			return SubcontractOrder{}, err
		}
		current = next
	}

	return current.RejectWithFactoryIssue(actorID, reason, changedAt)
}

func (o SubcontractOrder) MarkFinalPaymentReady(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusFinalPaymentReady, actorID, changedAt)
}

func (o SubcontractOrder) Close(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusClosed, actorID, changedAt)
}

func (o SubcontractOrder) Cancel(actorID string, changedAt time.Time) (SubcontractOrder, error) {
	return o.TransitionTo(SubcontractOrderStatusCancelled, actorID, changedAt)
}

func (o SubcontractOrder) CancelWithReason(actorID string, reason string, changedAt time.Time) (SubcontractOrder, error) {
	cancelled, err := o.Cancel(actorID, changedAt)
	if err != nil {
		return SubcontractOrder{}, err
	}
	cancelled.CancelReason = strings.TrimSpace(reason)

	return cancelled, nil
}

func (o SubcontractOrder) TransitionTo(status SubcontractOrderStatus, actorID string, changedAt time.Time) (SubcontractOrder, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SubcontractOrder{}, ErrSubcontractOrderTransitionActorRequired
	}
	from := NormalizeSubcontractOrderStatus(o.Status)
	to := NormalizeSubcontractOrderStatus(status)
	if !IsValidSubcontractOrderStatus(from) || !IsValidSubcontractOrderStatus(to) {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidStatus
	}
	if !CanTransitionSubcontractOrderStatus(from, to) {
		return SubcontractOrder{}, ErrSubcontractOrderInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := o.Clone()
	updated.Status = to
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if updated.Version > 0 {
		updated.Version++
	}
	updated.markTransition(to, actorID, changedAt.UTC())

	return updated, nil
}

func (o *SubcontractOrder) markTransition(status SubcontractOrderStatus, actorID string, changedAt time.Time) {
	switch status {
	case SubcontractOrderStatusSubmitted:
		o.SubmittedAt = changedAt
		o.SubmittedBy = actorID
	case SubcontractOrderStatusApproved:
		o.ApprovedAt = changedAt
		o.ApprovedBy = actorID
	case SubcontractOrderStatusFactoryConfirmed:
		o.FactoryConfirmedAt = changedAt
		o.FactoryConfirmedBy = actorID
	case SubcontractOrderStatusDepositRecorded:
		o.DepositRecordedAt = changedAt
		o.DepositRecordedBy = actorID
	case SubcontractOrderStatusMaterialsIssued:
		o.MaterialsIssuedAt = changedAt
		o.MaterialsIssuedBy = actorID
	case SubcontractOrderStatusSampleSubmitted:
		o.SampleSubmittedAt = changedAt
		o.SampleSubmittedBy = actorID
	case SubcontractOrderStatusSampleApproved:
		o.SampleApprovedAt = changedAt
		o.SampleApprovedBy = actorID
	case SubcontractOrderStatusSampleRejected:
		o.SampleRejectedAt = changedAt
		o.SampleRejectedBy = actorID
	case SubcontractOrderStatusMassProductionStarted:
		o.MassProductionStartedAt = changedAt
		o.MassProductionStartedBy = actorID
	case SubcontractOrderStatusFinishedGoodsReceived:
		o.FinishedGoodsReceivedAt = changedAt
		o.FinishedGoodsReceivedBy = actorID
	case SubcontractOrderStatusQCInProgress:
		o.QCStartedAt = changedAt
		o.QCStartedBy = actorID
	case SubcontractOrderStatusAccepted:
		o.AcceptedAt = changedAt
		o.AcceptedBy = actorID
	case SubcontractOrderStatusRejectedFactoryIssue:
		o.RejectedFactoryIssueAt = changedAt
		o.RejectedFactoryIssueBy = actorID
	case SubcontractOrderStatusFinalPaymentReady:
		o.FinalPaymentReadyAt = changedAt
		o.FinalPaymentReadyBy = actorID
	case SubcontractOrderStatusClosed:
		o.ClosedAt = changedAt
		o.ClosedBy = actorID
	case SubcontractOrderStatusCancelled:
		o.CancelledAt = changedAt
		o.CancelledBy = actorID
	}
}

func (o SubcontractOrder) Clone() SubcontractOrder {
	clone := o
	clone.MaterialLines = append([]SubcontractMaterialLine(nil), o.MaterialLines...)

	return clone
}

func (o *SubcontractOrder) recalculateEstimatedCost() error {
	total := decimal.MustMoneyAmount("0")
	for _, line := range o.MaterialLines {
		var err error
		total, err = addSubcontractOrderMoney(total, line.LineCostAmount)
		if err != nil {
			return err
		}
	}
	o.EstimatedCostAmount = total

	return nil
}

func NormalizeSubcontractOrderStatus(status SubcontractOrderStatus) SubcontractOrderStatus {
	return SubcontractOrderStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSubcontractOrderStatus(status SubcontractOrderStatus) bool {
	switch NormalizeSubcontractOrderStatus(status) {
	case SubcontractOrderStatusDraft,
		SubcontractOrderStatusSubmitted,
		SubcontractOrderStatusApproved,
		SubcontractOrderStatusFactoryConfirmed,
		SubcontractOrderStatusDepositRecorded,
		SubcontractOrderStatusMaterialsIssued,
		SubcontractOrderStatusSampleSubmitted,
		SubcontractOrderStatusSampleApproved,
		SubcontractOrderStatusSampleRejected,
		SubcontractOrderStatusMassProductionStarted,
		SubcontractOrderStatusFinishedGoodsReceived,
		SubcontractOrderStatusQCInProgress,
		SubcontractOrderStatusAccepted,
		SubcontractOrderStatusRejectedFactoryIssue,
		SubcontractOrderStatusFinalPaymentReady,
		SubcontractOrderStatusClosed,
		SubcontractOrderStatusCancelled:
		return true
	default:
		return false
	}
}

func CanTransitionSubcontractOrderStatus(from SubcontractOrderStatus, to SubcontractOrderStatus) bool {
	from = NormalizeSubcontractOrderStatus(from)
	to = NormalizeSubcontractOrderStatus(to)
	if from == to || !IsValidSubcontractOrderStatus(from) || !IsValidSubcontractOrderStatus(to) {
		return false
	}
	for _, candidate := range subcontractOrderTransitions[from] {
		if candidate == to {
			return true
		}
	}

	return false
}

func normalizeSubcontractOrderCurrency(value string) (decimal.CurrencyCode, error) {
	currencyCode, err := decimal.NormalizeCurrencyCode(value)
	if err != nil || currencyCode != decimal.CurrencyVND {
		return "", ErrSubcontractOrderInvalidCurrency
	}

	return currencyCode, nil
}

func normalizeSubcontractOrderQuantitySet(
	quantityValue decimal.Decimal,
	baseQuantityValue decimal.Decimal,
	uomValue string,
	baseUOMValue string,
	conversionFactorValue decimal.Decimal,
	requirePositive bool,
) (decimal.Decimal, decimal.Decimal, decimal.UOMCode, decimal.UOMCode, decimal.Decimal, error) {
	quantity, err := normalizeSubcontractOrderQuantity(quantityValue, requirePositive)
	if err != nil {
		return "", "", "", "", "", err
	}
	uomCode, err := decimal.NormalizeUOMCode(uomValue)
	if err != nil {
		return "", "", "", "", "", ErrSubcontractOrderInvalidQuantity
	}
	baseUOMCode := uomCode
	if strings.TrimSpace(baseUOMValue) != "" {
		baseUOMCode, err = decimal.NormalizeUOMCode(baseUOMValue)
		if err != nil {
			return "", "", "", "", "", ErrSubcontractOrderInvalidQuantity
		}
	}
	conversionFactor := decimal.MustQuantity("1")
	if strings.TrimSpace(conversionFactorValue.String()) != "" {
		conversionFactor, err = normalizeSubcontractOrderQuantity(conversionFactorValue, true)
		if err != nil {
			return "", "", "", "", "", err
		}
	}
	baseQuantity := quantity
	if strings.TrimSpace(baseQuantityValue.String()) != "" {
		baseQuantity, err = normalizeSubcontractOrderQuantity(baseQuantityValue, requirePositive)
		if err != nil {
			return "", "", "", "", "", err
		}
	} else if baseUOMCode != uomCode {
		baseQuantity, err = decimal.MultiplyQuantityByFactor(quantity, conversionFactor)
		if err != nil {
			return "", "", "", "", "", ErrSubcontractOrderInvalidQuantity
		}
	}

	return quantity, baseQuantity, uomCode, baseUOMCode, conversionFactor, nil
}

func normalizeSubcontractOrderProgressQuantity(
	quantityValue decimal.Decimal,
	baseQuantityValue decimal.Decimal,
	baseUOMCode decimal.UOMCode,
	uomCode decimal.UOMCode,
	conversionFactor decimal.Decimal,
) (decimal.Decimal, decimal.Decimal, error) {
	quantity := decimal.MustQuantity("0")
	var err error
	if strings.TrimSpace(quantityValue.String()) != "" {
		quantity, err = normalizeSubcontractOrderQuantity(quantityValue, false)
		if err != nil {
			return "", "", err
		}
	}
	baseQuantity := quantity
	if strings.TrimSpace(baseQuantityValue.String()) != "" {
		baseQuantity, err = normalizeSubcontractOrderQuantity(baseQuantityValue, false)
		if err != nil {
			return "", "", err
		}
	} else if baseUOMCode != uomCode {
		baseQuantity, err = decimal.MultiplyQuantityByFactor(quantity, conversionFactor)
		if err != nil {
			return "", "", ErrSubcontractOrderInvalidQuantity
		}
	}

	return quantity, baseQuantity, nil
}

func normalizeSubcontractOrderQuantity(value decimal.Decimal, requirePositive bool) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil || quantity.IsNegative() || (requirePositive && quantity.IsZero()) {
		return "", ErrSubcontractOrderInvalidQuantity
	}

	return quantity, nil
}

func normalizeSubcontractOrderNonNegativeUnitCost(value decimal.Decimal) (decimal.Decimal, error) {
	unitCost, err := decimal.ParseUnitCost(value.String())
	if err != nil || unitCost.IsNegative() {
		return "", ErrSubcontractOrderInvalidAmount
	}

	return unitCost, nil
}

func normalizeSubcontractOrderNonNegativeMoney(value decimal.Decimal) (decimal.Decimal, error) {
	amount, err := decimal.ParseMoneyAmount(value.String())
	if err != nil || amount.IsNegative() {
		return "", ErrSubcontractOrderInvalidAmount
	}

	return amount, nil
}

func validateSubcontractOrderProgressQuantity(left decimal.Decimal, right decimal.Decimal) error {
	compare, err := compareSubcontractOrderQuantity(left, right)
	if err != nil || compare > 0 {
		return ErrSubcontractOrderInvalidQuantity
	}

	return nil
}

func validateSubcontractOrderQuantitySum(left decimal.Decimal, right decimal.Decimal, limit decimal.Decimal) error {
	sum, err := decimal.AddQuantity(left, right)
	if err != nil {
		return ErrSubcontractOrderInvalidQuantity
	}

	return validateSubcontractOrderProgressQuantity(sum, limit)
}

func subcontractOrderMoneyFromQuantityUnitCost(quantity decimal.Decimal, unitCost decimal.Decimal) (decimal.Decimal, error) {
	quantityValue, ok := new(big.Rat).SetString(quantity.String())
	if !ok {
		return "", ErrSubcontractOrderInvalidQuantity
	}
	unitCostValue, ok := new(big.Rat).SetString(unitCost.String())
	if !ok {
		return "", ErrSubcontractOrderInvalidAmount
	}
	amount := new(big.Rat).Mul(quantityValue, unitCostValue)
	if amount.Sign() < 0 {
		return "", ErrSubcontractOrderInvalidAmount
	}

	return roundSubcontractOrderRatToMoney(amount)
}

func addSubcontractOrderMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	leftValue, ok := new(big.Rat).SetString(left.String())
	if !ok {
		return "", ErrSubcontractOrderInvalidAmount
	}
	rightValue, ok := new(big.Rat).SetString(right.String())
	if !ok {
		return "", ErrSubcontractOrderInvalidAmount
	}
	sum := new(big.Rat).Add(leftValue, rightValue)
	if sum.Sign() < 0 {
		return "", ErrSubcontractOrderInvalidAmount
	}

	return roundSubcontractOrderRatToMoney(sum)
}

func roundSubcontractOrderRatToMoney(value *big.Rat) (decimal.Decimal, error) {
	if value.Sign() < 0 {
		return "", ErrSubcontractOrderInvalidAmount
	}
	scaled := new(big.Rat).Mul(value, big.NewRat(100, 1))
	quotient, remainder := new(big.Int).QuoRem(scaled.Num(), scaled.Denom(), new(big.Int))
	if new(big.Int).Mul(remainder, big.NewInt(2)).Cmp(scaled.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	digits := quotient.String()
	if len(digits) <= decimal.MoneyScale {
		digits = strings.Repeat("0", decimal.MoneyScale-len(digits)+1) + digits
	}
	intPart := digits[:len(digits)-decimal.MoneyScale]
	fracPart := digits[len(digits)-decimal.MoneyScale:]
	money, err := decimal.ParseMoneyAmount(fmt.Sprintf("%s.%s", intPart, fracPart))
	if err != nil {
		return "", ErrSubcontractOrderInvalidAmount
	}

	return money, nil
}

func compareSubcontractOrderQuantity(left decimal.Decimal, right decimal.Decimal) (int, error) {
	leftValue, ok := new(big.Rat).SetString(left.String())
	if !ok {
		return 0, ErrSubcontractOrderInvalidQuantity
	}
	rightValue, ok := new(big.Rat).SetString(right.String())
	if !ok {
		return 0, ErrSubcontractOrderInvalidQuantity
	}

	return leftValue.Cmp(rightValue), nil
}

func allSubcontractMaterialLinesIssued(lines []SubcontractMaterialLine) bool {
	if len(lines) == 0 {
		return false
	}
	for _, line := range lines {
		compare, err := compareSubcontractOrderQuantity(line.BaseIssuedQty, line.BasePlannedQty)
		if err != nil || compare != 0 {
			return false
		}
	}

	return true
}

func subcontractOrderFirstNonBlank(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
