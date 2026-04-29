package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrPurchaseOrderRequiredField = errors.New("purchase order required field is missing")
var ErrPurchaseOrderInvalidStatus = errors.New("purchase order status is invalid")
var ErrPurchaseOrderInvalidTransition = errors.New("purchase order status transition is invalid")
var ErrPurchaseOrderTransitionActorRequired = errors.New("purchase order status transition actor is required")
var ErrPurchaseOrderInvalidCurrency = errors.New("purchase order currency is invalid")
var ErrPurchaseOrderInvalidQuantity = errors.New("purchase order quantity is invalid")
var ErrPurchaseOrderInvalidAmount = errors.New("purchase order amount is invalid")

type PurchaseOrderStatus string

const (
	PurchaseOrderStatusDraft             PurchaseOrderStatus = "draft"
	PurchaseOrderStatusSubmitted         PurchaseOrderStatus = "submitted"
	PurchaseOrderStatusApproved          PurchaseOrderStatus = "approved"
	PurchaseOrderStatusPartiallyReceived PurchaseOrderStatus = "partially_received"
	PurchaseOrderStatusReceived          PurchaseOrderStatus = "received"
	PurchaseOrderStatusClosed            PurchaseOrderStatus = "closed"
	PurchaseOrderStatusCancelled         PurchaseOrderStatus = "cancelled"
	PurchaseOrderStatusRejected          PurchaseOrderStatus = "rejected"
)

type PurchaseOrder struct {
	ID             string
	OrgID          string
	PONo           string
	SupplierID     string
	SupplierCode   string
	SupplierName   string
	WarehouseID    string
	WarehouseCode  string
	ExpectedDate   string
	Status         PurchaseOrderStatus
	CurrencyCode   decimal.CurrencyCode
	SubtotalAmount decimal.Decimal
	TotalAmount    decimal.Decimal
	Note           string
	Lines          []PurchaseOrderLine
	CreatedAt      time.Time
	CreatedBy      string
	UpdatedAt      time.Time
	UpdatedBy      string
	Version        int
	CancelReason   string
	RejectReason   string

	SubmittedAt         time.Time
	SubmittedBy         string
	ApprovedAt          time.Time
	ApprovedBy          string
	PartiallyReceivedAt time.Time
	PartiallyReceivedBy string
	ReceivedAt          time.Time
	ReceivedBy          string
	ClosedAt            time.Time
	ClosedBy            string
	CancelledAt         time.Time
	CancelledBy         string
	RejectedAt          time.Time
	RejectedBy          string
}

type PurchaseOrderLine struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	OrderedQty       decimal.Decimal
	ReceivedQty      decimal.Decimal
	UOMCode          decimal.UOMCode
	BaseOrderedQty   decimal.Decimal
	BaseReceivedQty  decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	ConversionFactor decimal.Decimal
	UnitPrice        decimal.Decimal
	CurrencyCode     decimal.CurrencyCode
	LineAmount       decimal.Decimal
	ExpectedDate     string
	Note             string
}

type NewPurchaseOrderDocumentInput struct {
	ID            string
	OrgID         string
	PONo          string
	SupplierID    string
	SupplierCode  string
	SupplierName  string
	WarehouseID   string
	WarehouseCode string
	ExpectedDate  string
	CurrencyCode  string
	Note          string
	Lines         []NewPurchaseOrderLineInput
	CreatedAt     time.Time
	CreatedBy     string
	UpdatedAt     time.Time
}

type NewPurchaseOrderLineInput struct {
	ID               string
	LineNo           int
	ItemID           string
	SKUCode          string
	ItemName         string
	OrderedQty       decimal.Decimal
	ReceivedQty      decimal.Decimal
	UOMCode          string
	BaseOrderedQty   decimal.Decimal
	BaseReceivedQty  decimal.Decimal
	BaseUOMCode      string
	ConversionFactor decimal.Decimal
	UnitPrice        decimal.Decimal
	CurrencyCode     string
	ExpectedDate     string
	Note             string
}

var purchaseOrderTransitions = map[PurchaseOrderStatus][]PurchaseOrderStatus{
	PurchaseOrderStatusDraft: {
		PurchaseOrderStatusSubmitted,
		PurchaseOrderStatusCancelled,
	},
	PurchaseOrderStatusSubmitted: {
		PurchaseOrderStatusApproved,
		PurchaseOrderStatusRejected,
		PurchaseOrderStatusCancelled,
	},
	PurchaseOrderStatusApproved: {
		PurchaseOrderStatusPartiallyReceived,
		PurchaseOrderStatusReceived,
		PurchaseOrderStatusClosed,
		PurchaseOrderStatusCancelled,
	},
	PurchaseOrderStatusPartiallyReceived: {
		PurchaseOrderStatusReceived,
		PurchaseOrderStatusClosed,
	},
	PurchaseOrderStatusReceived: {
		PurchaseOrderStatusClosed,
	},
}

func NewPurchaseOrder(status PurchaseOrderStatus) (PurchaseOrder, error) {
	status = NormalizePurchaseOrderStatus(status)
	if status == "" {
		status = PurchaseOrderStatusDraft
	}
	if !IsValidPurchaseOrderStatus(status) {
		return PurchaseOrder{}, ErrPurchaseOrderInvalidStatus
	}

	return PurchaseOrder{Status: status}, nil
}

func NewPurchaseOrderDocument(input NewPurchaseOrderDocumentInput) (PurchaseOrder, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	currencyCode, err := normalizePurchaseOrderCurrency(input.CurrencyCode)
	if err != nil {
		return PurchaseOrder{}, err
	}
	updatedBy := strings.TrimSpace(input.UpdatedBy)
	if updatedBy == "" {
		updatedBy = strings.TrimSpace(input.CreatedBy)
	}

	order := PurchaseOrder{
		ID:            strings.TrimSpace(input.ID),
		OrgID:         strings.TrimSpace(input.OrgID),
		PONo:          strings.ToUpper(strings.TrimSpace(input.PONo)),
		SupplierID:    strings.TrimSpace(input.SupplierID),
		SupplierCode:  strings.ToUpper(strings.TrimSpace(input.SupplierCode)),
		SupplierName:  strings.TrimSpace(input.SupplierName),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		ExpectedDate:  strings.TrimSpace(input.ExpectedDate),
		Status:        PurchaseOrderStatusDraft,
		CurrencyCode:  currencyCode,
		Note:          strings.TrimSpace(input.Note),
		Lines:         make([]PurchaseOrderLine, 0, len(input.Lines)),
		CreatedAt:     createdAt.UTC(),
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		UpdatedAt:     updatedAt.UTC(),
		UpdatedBy:     updatedBy,
		Version:       1,
	}
	for index, lineInput := range input.Lines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		if lineInput.CurrencyCode == "" {
			lineInput.CurrencyCode = currencyCode.String()
		}
		if strings.TrimSpace(lineInput.ExpectedDate) == "" {
			lineInput.ExpectedDate = order.ExpectedDate
		}
		line, err := NewPurchaseOrderLine(lineInput)
		if err != nil {
			return PurchaseOrder{}, err
		}
		order.Lines = append(order.Lines, line)
	}
	if err := order.recalculateAmounts(); err != nil {
		return PurchaseOrder{}, err
	}
	if err := order.Validate(); err != nil {
		return PurchaseOrder{}, err
	}

	return order, nil
}

func NewPurchaseOrderLine(input NewPurchaseOrderLineInput) (PurchaseOrderLine, error) {
	orderedQty, err := normalizePurchaseOrderPositiveQuantity(input.OrderedQty)
	if err != nil {
		return PurchaseOrderLine{}, err
	}
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return PurchaseOrderLine{}, ErrPurchaseOrderInvalidQuantity
	}
	baseUOMCode := uomCode
	if strings.TrimSpace(input.BaseUOMCode) != "" {
		baseUOMCode, err = decimal.NormalizeUOMCode(input.BaseUOMCode)
		if err != nil {
			return PurchaseOrderLine{}, ErrPurchaseOrderInvalidQuantity
		}
	}
	conversionFactor := decimal.MustQuantity("1")
	if input.ConversionFactor != "" {
		conversionFactor, err = normalizePurchaseOrderPositiveQuantity(input.ConversionFactor)
		if err != nil {
			return PurchaseOrderLine{}, err
		}
	}
	baseOrderedQty := orderedQty
	if input.BaseOrderedQty != "" {
		baseOrderedQty, err = normalizePurchaseOrderPositiveQuantity(input.BaseOrderedQty)
		if err != nil {
			return PurchaseOrderLine{}, err
		}
	} else if baseUOMCode != uomCode {
		baseOrderedQty, err = decimal.MultiplyQuantityByFactor(orderedQty, conversionFactor)
		if err != nil {
			return PurchaseOrderLine{}, ErrPurchaseOrderInvalidQuantity
		}
	}
	receivedQty := decimal.MustQuantity("0")
	if input.ReceivedQty != "" {
		receivedQty, err = normalizePurchaseOrderNonNegativeQuantity(input.ReceivedQty)
		if err != nil {
			return PurchaseOrderLine{}, err
		}
	}
	baseReceivedQty := receivedQty
	if input.BaseReceivedQty != "" {
		baseReceivedQty, err = normalizePurchaseOrderNonNegativeQuantity(input.BaseReceivedQty)
		if err != nil {
			return PurchaseOrderLine{}, err
		}
	} else if baseUOMCode != uomCode {
		baseReceivedQty, err = decimal.MultiplyQuantityByFactor(receivedQty, conversionFactor)
		if err != nil {
			return PurchaseOrderLine{}, ErrPurchaseOrderInvalidQuantity
		}
	}
	unitPrice, err := normalizePurchaseOrderNonNegativeUnitPrice(input.UnitPrice)
	if err != nil {
		return PurchaseOrderLine{}, err
	}
	currencyCode, err := normalizePurchaseOrderCurrency(input.CurrencyCode)
	if err != nil {
		return PurchaseOrderLine{}, err
	}
	lineAmount, err := purchaseOrderMoneyFromQuantityUnitPrice(orderedQty, unitPrice)
	if err != nil {
		return PurchaseOrderLine{}, err
	}

	line := PurchaseOrderLine{
		ID:               strings.TrimSpace(input.ID),
		LineNo:           input.LineNo,
		ItemID:           strings.TrimSpace(input.ItemID),
		SKUCode:          strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		ItemName:         strings.TrimSpace(input.ItemName),
		OrderedQty:       orderedQty,
		ReceivedQty:      receivedQty,
		UOMCode:          uomCode,
		BaseOrderedQty:   baseOrderedQty,
		BaseReceivedQty:  baseReceivedQty,
		BaseUOMCode:      baseUOMCode,
		ConversionFactor: conversionFactor,
		UnitPrice:        unitPrice,
		CurrencyCode:     currencyCode,
		LineAmount:       lineAmount,
		ExpectedDate:     strings.TrimSpace(input.ExpectedDate),
		Note:             strings.TrimSpace(input.Note),
	}
	if err := line.Validate(); err != nil {
		return PurchaseOrderLine{}, err
	}

	return line, nil
}

func (o PurchaseOrder) Validate() error {
	if strings.TrimSpace(o.ID) == "" ||
		strings.TrimSpace(o.OrgID) == "" ||
		strings.TrimSpace(o.PONo) == "" ||
		strings.TrimSpace(o.SupplierID) == "" ||
		strings.TrimSpace(o.SupplierName) == "" ||
		strings.TrimSpace(o.WarehouseID) == "" ||
		strings.TrimSpace(o.ExpectedDate) == "" ||
		strings.TrimSpace(o.CreatedBy) == "" ||
		len(o.Lines) == 0 {
		return ErrPurchaseOrderRequiredField
	}
	if !IsValidPurchaseOrderStatus(o.Status) {
		return ErrPurchaseOrderInvalidStatus
	}
	if o.CurrencyCode != decimal.CurrencyVND {
		return ErrPurchaseOrderInvalidCurrency
	}
	for _, amount := range []decimal.Decimal{o.SubtotalAmount, o.TotalAmount} {
		value, err := decimal.ParseMoneyAmount(amount.String())
		if err != nil || value.IsNegative() {
			return ErrPurchaseOrderInvalidAmount
		}
	}
	seenLineNo := map[int]struct{}{}
	for _, line := range o.Lines {
		if _, exists := seenLineNo[line.LineNo]; exists {
			return ErrPurchaseOrderRequiredField
		}
		seenLineNo[line.LineNo] = struct{}{}
		if line.CurrencyCode != o.CurrencyCode {
			return ErrPurchaseOrderInvalidCurrency
		}
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l PurchaseOrderLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKUCode) == "" ||
		strings.TrimSpace(l.ItemName) == "" {
		return ErrPurchaseOrderRequiredField
	}
	if l.CurrencyCode != decimal.CurrencyVND {
		return ErrPurchaseOrderInvalidCurrency
	}
	for _, quantity := range []decimal.Decimal{l.OrderedQty, l.BaseOrderedQty, l.ConversionFactor} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrPurchaseOrderInvalidQuantity
		}
	}
	for _, quantity := range []decimal.Decimal{l.ReceivedQty, l.BaseReceivedQty} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() {
			return ErrPurchaseOrderInvalidQuantity
		}
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrPurchaseOrderInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrPurchaseOrderInvalidQuantity
	}
	if compare, err := comparePurchaseOrderQuantity(l.ReceivedQty, l.OrderedQty); err != nil || compare > 0 {
		return ErrPurchaseOrderInvalidQuantity
	}
	if compare, err := comparePurchaseOrderQuantity(l.BaseReceivedQty, l.BaseOrderedQty); err != nil || compare > 0 {
		return ErrPurchaseOrderInvalidQuantity
	}
	if _, err := normalizePurchaseOrderNonNegativeUnitPrice(l.UnitPrice); err != nil {
		return err
	}
	value, err := decimal.ParseMoneyAmount(l.LineAmount.String())
	if err != nil || value.IsNegative() {
		return ErrPurchaseOrderInvalidAmount
	}

	return nil
}

func (o PurchaseOrder) Submit(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusSubmitted, actorID, changedAt)
}

func (o PurchaseOrder) Approve(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusApproved, actorID, changedAt)
}

func (o PurchaseOrder) MarkPartiallyReceived(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusPartiallyReceived, actorID, changedAt)
}

func (o PurchaseOrder) MarkReceived(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusReceived, actorID, changedAt)
}

func (o PurchaseOrder) Close(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusClosed, actorID, changedAt)
}

func (o PurchaseOrder) Cancel(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusCancelled, actorID, changedAt)
}

func (o PurchaseOrder) Reject(actorID string, changedAt time.Time) (PurchaseOrder, error) {
	return o.TransitionTo(PurchaseOrderStatusRejected, actorID, changedAt)
}

func (o PurchaseOrder) CancelWithReason(actorID string, reason string, changedAt time.Time) (PurchaseOrder, error) {
	cancelled, err := o.Cancel(actorID, changedAt)
	if err != nil {
		return PurchaseOrder{}, err
	}
	cancelled.CancelReason = strings.TrimSpace(reason)

	return cancelled, nil
}

func (o PurchaseOrder) RejectWithReason(actorID string, reason string, changedAt time.Time) (PurchaseOrder, error) {
	rejected, err := o.Reject(actorID, changedAt)
	if err != nil {
		return PurchaseOrder{}, err
	}
	rejected.RejectReason = strings.TrimSpace(reason)

	return rejected, nil
}

func (o PurchaseOrder) TransitionTo(status PurchaseOrderStatus, actorID string, changedAt time.Time) (PurchaseOrder, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return PurchaseOrder{}, ErrPurchaseOrderTransitionActorRequired
	}
	from := NormalizePurchaseOrderStatus(o.Status)
	to := NormalizePurchaseOrderStatus(status)
	if !IsValidPurchaseOrderStatus(from) || !IsValidPurchaseOrderStatus(to) {
		return PurchaseOrder{}, ErrPurchaseOrderInvalidStatus
	}
	if !CanTransitionPurchaseOrderStatus(from, to) {
		return PurchaseOrder{}, ErrPurchaseOrderInvalidTransition
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

func (o *PurchaseOrder) markTransition(status PurchaseOrderStatus, actorID string, changedAt time.Time) {
	switch status {
	case PurchaseOrderStatusSubmitted:
		o.SubmittedAt = changedAt
		o.SubmittedBy = actorID
	case PurchaseOrderStatusApproved:
		o.ApprovedAt = changedAt
		o.ApprovedBy = actorID
	case PurchaseOrderStatusPartiallyReceived:
		o.PartiallyReceivedAt = changedAt
		o.PartiallyReceivedBy = actorID
	case PurchaseOrderStatusReceived:
		o.ReceivedAt = changedAt
		o.ReceivedBy = actorID
	case PurchaseOrderStatusClosed:
		o.ClosedAt = changedAt
		o.ClosedBy = actorID
	case PurchaseOrderStatusCancelled:
		o.CancelledAt = changedAt
		o.CancelledBy = actorID
	case PurchaseOrderStatusRejected:
		o.RejectedAt = changedAt
		o.RejectedBy = actorID
	}
}

func (o PurchaseOrder) Clone() PurchaseOrder {
	clone := o
	clone.Lines = append([]PurchaseOrderLine(nil), o.Lines...)

	return clone
}

func (o *PurchaseOrder) recalculateAmounts() error {
	subtotal := decimal.MustMoneyAmount("0")
	for _, line := range o.Lines {
		var err error
		subtotal, err = addPurchaseOrderMoney(subtotal, line.LineAmount)
		if err != nil {
			return err
		}
	}

	o.SubtotalAmount = subtotal
	o.TotalAmount = subtotal

	return nil
}

func NormalizePurchaseOrderStatus(status PurchaseOrderStatus) PurchaseOrderStatus {
	return PurchaseOrderStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidPurchaseOrderStatus(status PurchaseOrderStatus) bool {
	switch NormalizePurchaseOrderStatus(status) {
	case PurchaseOrderStatusDraft,
		PurchaseOrderStatusSubmitted,
		PurchaseOrderStatusApproved,
		PurchaseOrderStatusPartiallyReceived,
		PurchaseOrderStatusReceived,
		PurchaseOrderStatusClosed,
		PurchaseOrderStatusCancelled,
		PurchaseOrderStatusRejected:
		return true
	default:
		return false
	}
}

func CanTransitionPurchaseOrderStatus(from PurchaseOrderStatus, to PurchaseOrderStatus) bool {
	from = NormalizePurchaseOrderStatus(from)
	to = NormalizePurchaseOrderStatus(to)
	if from == to || !IsValidPurchaseOrderStatus(from) || !IsValidPurchaseOrderStatus(to) {
		return false
	}
	for _, candidate := range purchaseOrderTransitions[from] {
		if candidate == to {
			return true
		}
	}

	return false
}

func normalizePurchaseOrderCurrency(value string) (decimal.CurrencyCode, error) {
	currencyCode, err := decimal.NormalizeCurrencyCode(value)
	if err != nil || currencyCode != decimal.CurrencyVND {
		return "", ErrPurchaseOrderInvalidCurrency
	}

	return currencyCode, nil
}

func normalizePurchaseOrderPositiveQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		return "", ErrPurchaseOrderInvalidQuantity
	}

	return quantity, nil
}

func normalizePurchaseOrderNonNegativeQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil || quantity.IsNegative() {
		return "", ErrPurchaseOrderInvalidQuantity
	}

	return quantity, nil
}

func normalizePurchaseOrderNonNegativeUnitPrice(value decimal.Decimal) (decimal.Decimal, error) {
	unitPrice, err := decimal.ParseUnitPrice(value.String())
	if err != nil || unitPrice.IsNegative() {
		return "", ErrPurchaseOrderInvalidAmount
	}

	return unitPrice, nil
}

func purchaseOrderMoneyFromQuantityUnitPrice(quantity decimal.Decimal, unitPrice decimal.Decimal) (decimal.Decimal, error) {
	quantityValue, ok := new(big.Rat).SetString(quantity.String())
	if !ok {
		return "", ErrPurchaseOrderInvalidQuantity
	}
	unitPriceValue, ok := new(big.Rat).SetString(unitPrice.String())
	if !ok {
		return "", ErrPurchaseOrderInvalidAmount
	}

	amount := new(big.Rat).Mul(quantityValue, unitPriceValue)
	if amount.Sign() < 0 {
		return "", ErrPurchaseOrderInvalidAmount
	}

	return roundPurchaseOrderRatToMoney(amount)
}

func addPurchaseOrderMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	leftValue, ok := new(big.Rat).SetString(left.String())
	if !ok {
		return "", ErrPurchaseOrderInvalidAmount
	}
	rightValue, ok := new(big.Rat).SetString(right.String())
	if !ok {
		return "", ErrPurchaseOrderInvalidAmount
	}
	sum := new(big.Rat).Add(leftValue, rightValue)
	if sum.Sign() < 0 {
		return "", ErrPurchaseOrderInvalidAmount
	}

	return roundPurchaseOrderRatToMoney(sum)
}

func roundPurchaseOrderRatToMoney(value *big.Rat) (decimal.Decimal, error) {
	if value.Sign() < 0 {
		return "", ErrPurchaseOrderInvalidAmount
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
		return "", ErrPurchaseOrderInvalidAmount
	}

	return money, nil
}

func comparePurchaseOrderQuantity(left decimal.Decimal, right decimal.Decimal) (int, error) {
	leftValue, ok := new(big.Rat).SetString(left.String())
	if !ok {
		return 0, ErrPurchaseOrderInvalidQuantity
	}
	rightValue, ok := new(big.Rat).SetString(right.String())
	if !ok {
		return 0, ErrPurchaseOrderInvalidQuantity
	}

	return leftValue.Cmp(rightValue), nil
}
