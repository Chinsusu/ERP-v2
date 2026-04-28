package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrSalesOrderRequiredField = errors.New("sales order required field is missing")
var ErrSalesOrderInvalidStatus = errors.New("sales order status is invalid")
var ErrSalesOrderInvalidTransition = errors.New("sales order status transition is invalid")
var ErrSalesOrderTransitionActorRequired = errors.New("sales order status transition actor is required")
var ErrSalesOrderInvalidCurrency = errors.New("sales order currency is invalid")
var ErrSalesOrderInvalidQuantity = errors.New("sales order quantity is invalid")
var ErrSalesOrderInvalidAmount = errors.New("sales order amount is invalid")

type SalesOrderStatus string

const (
	SalesOrderStatusDraft             SalesOrderStatus = "draft"
	SalesOrderStatusConfirmed         SalesOrderStatus = "confirmed"
	SalesOrderStatusReserved          SalesOrderStatus = "reserved"
	SalesOrderStatusPicking           SalesOrderStatus = "picking"
	SalesOrderStatusPicked            SalesOrderStatus = "picked"
	SalesOrderStatusPacking           SalesOrderStatus = "packing"
	SalesOrderStatusPacked            SalesOrderStatus = "packed"
	SalesOrderStatusWaitingHandover   SalesOrderStatus = "waiting_handover"
	SalesOrderStatusHandedOver        SalesOrderStatus = "handed_over"
	SalesOrderStatusDelivered         SalesOrderStatus = "delivered"
	SalesOrderStatusReturned          SalesOrderStatus = "returned"
	SalesOrderStatusClosed            SalesOrderStatus = "closed"
	SalesOrderStatusCancelled         SalesOrderStatus = "cancelled"
	SalesOrderStatusReservationFailed SalesOrderStatus = "reservation_failed"
	SalesOrderStatusPickException     SalesOrderStatus = "pick_exception"
	SalesOrderStatusPackException     SalesOrderStatus = "pack_exception"
	SalesOrderStatusHandoverException SalesOrderStatus = "handover_exception"
)

type SalesOrder struct {
	ID                string
	OrgID             string
	OrderNo           string
	CustomerID        string
	CustomerCode      string
	CustomerName      string
	Channel           string
	WarehouseID       string
	WarehouseCode     string
	OrderDate         string
	Status            SalesOrderStatus
	CurrencyCode      decimal.CurrencyCode
	SubtotalAmount    decimal.Decimal
	DiscountAmount    decimal.Decimal
	TaxAmount         decimal.Decimal
	ShippingFeeAmount decimal.Decimal
	NetAmount         decimal.Decimal
	TotalAmount       decimal.Decimal
	Note              string
	Lines             []SalesOrderLine
	CreatedAt         time.Time
	CreatedBy         string
	UpdatedAt         time.Time
	UpdatedBy         string
	Version           int
	CancelReason      string

	ConfirmedAt       time.Time
	ConfirmedBy       string
	ReservedAt        time.Time
	ReservedBy        string
	PickingStartedAt  time.Time
	PickingStartedBy  string
	PickedAt          time.Time
	PickedBy          string
	PackingStartedAt  time.Time
	PackingStartedBy  string
	PackedAt          time.Time
	PackedBy          string
	WaitingHandoverAt time.Time
	WaitingHandoverBy string
	HandedOverAt      time.Time
	HandedOverBy      string
	ClosedAt          time.Time
	ClosedBy          string
	CancelledAt       time.Time
	CancelledBy       string
	ExceptionAt       time.Time
	ExceptionBy       string
}

type SalesOrderLine struct {
	ID                 string
	LineNo             int
	ItemID             string
	SKUCode            string
	ItemName           string
	OrderedQty         decimal.Decimal
	UOMCode            decimal.UOMCode
	BaseOrderedQty     decimal.Decimal
	BaseUOMCode        decimal.UOMCode
	ConversionFactor   decimal.Decimal
	UnitPrice          decimal.Decimal
	CurrencyCode       decimal.CurrencyCode
	LineDiscountAmount decimal.Decimal
	LineAmount         decimal.Decimal
	ReservedQty        decimal.Decimal
	ShippedQty         decimal.Decimal
	BatchID            string
	BatchNo            string
}

type NewSalesOrderDocumentInput struct {
	ID            string
	OrgID         string
	OrderNo       string
	CustomerID    string
	CustomerCode  string
	CustomerName  string
	Channel       string
	WarehouseID   string
	WarehouseCode string
	OrderDate     string
	CurrencyCode  string
	Note          string
	Lines         []NewSalesOrderLineInput
	CreatedAt     time.Time
	CreatedBy     string
	UpdatedAt     time.Time
}

type UpdateSalesOrderDocumentInput struct {
	CustomerID    string
	CustomerCode  string
	CustomerName  string
	Channel       string
	WarehouseID   string
	WarehouseCode string
	OrderDate     string
	Note          string
	Lines         []NewSalesOrderLineInput
	UpdatedAt     time.Time
	UpdatedBy     string
}

type NewSalesOrderLineInput struct {
	ID                 string
	LineNo             int
	ItemID             string
	SKUCode            string
	ItemName           string
	OrderedQty         decimal.Decimal
	UOMCode            string
	BaseOrderedQty     decimal.Decimal
	BaseUOMCode        string
	ConversionFactor   decimal.Decimal
	UnitPrice          decimal.Decimal
	CurrencyCode       string
	LineDiscountAmount decimal.Decimal
	ReservedQty        decimal.Decimal
	ShippedQty         decimal.Decimal
	BatchID            string
	BatchNo            string
}

var salesOrderTransitions = map[SalesOrderStatus][]SalesOrderStatus{
	SalesOrderStatusDraft: {
		SalesOrderStatusConfirmed,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusConfirmed: {
		SalesOrderStatusReserved,
		SalesOrderStatusReservationFailed,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusReserved: {
		SalesOrderStatusPicking,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPicking: {
		SalesOrderStatusPicked,
		SalesOrderStatusPickException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPicked: {
		SalesOrderStatusPacking,
		SalesOrderStatusPickException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPacking: {
		SalesOrderStatusPacked,
		SalesOrderStatusPackException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusPacked: {
		SalesOrderStatusWaitingHandover,
		SalesOrderStatusPackException,
		SalesOrderStatusCancelled,
	},
	SalesOrderStatusWaitingHandover: {
		SalesOrderStatusHandedOver,
		SalesOrderStatusHandoverException,
	},
	SalesOrderStatusHandedOver: {
		SalesOrderStatusClosed,
	},
	SalesOrderStatusDelivered: {
		SalesOrderStatusClosed,
	},
	SalesOrderStatusReturned: {
		SalesOrderStatusClosed,
	},
}

func NewSalesOrder(status SalesOrderStatus) (SalesOrder, error) {
	status = NormalizeSalesOrderStatus(status)
	if status == "" {
		status = SalesOrderStatusDraft
	}
	if !IsValidSalesOrderStatus(status) {
		return SalesOrder{}, ErrSalesOrderInvalidStatus
	}

	return SalesOrder{Status: status}, nil
}

func NewSalesOrderDocument(input NewSalesOrderDocumentInput) (SalesOrder, error) {
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	currencyCode, err := normalizeSalesOrderCurrency(input.CurrencyCode)
	if err != nil {
		return SalesOrder{}, err
	}

	order := SalesOrder{
		ID:            strings.TrimSpace(input.ID),
		OrgID:         strings.TrimSpace(input.OrgID),
		OrderNo:       strings.ToUpper(strings.TrimSpace(input.OrderNo)),
		CustomerID:    strings.TrimSpace(input.CustomerID),
		CustomerCode:  strings.ToUpper(strings.TrimSpace(input.CustomerCode)),
		CustomerName:  strings.TrimSpace(input.CustomerName),
		Channel:       strings.ToUpper(strings.TrimSpace(input.Channel)),
		WarehouseID:   strings.TrimSpace(input.WarehouseID),
		WarehouseCode: strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		OrderDate:     strings.TrimSpace(input.OrderDate),
		Status:        SalesOrderStatusDraft,
		CurrencyCode:  currencyCode,
		Note:          strings.TrimSpace(input.Note),
		Lines:         make([]SalesOrderLine, 0, len(input.Lines)),
		CreatedAt:     createdAt.UTC(),
		CreatedBy:     strings.TrimSpace(input.CreatedBy),
		UpdatedAt:     updatedAt.UTC(),
		UpdatedBy:     strings.TrimSpace(input.CreatedBy),
		Version:       1,
	}
	for index, lineInput := range input.Lines {
		if lineInput.LineNo == 0 {
			lineInput.LineNo = index + 1
		}
		if lineInput.CurrencyCode == "" {
			lineInput.CurrencyCode = currencyCode.String()
		}
		line, err := NewSalesOrderLine(lineInput)
		if err != nil {
			return SalesOrder{}, err
		}
		order.Lines = append(order.Lines, line)
	}
	if err := order.recalculateAmounts(); err != nil {
		return SalesOrder{}, err
	}
	if err := order.Validate(); err != nil {
		return SalesOrder{}, err
	}

	return order, nil
}

func NewSalesOrderLine(input NewSalesOrderLineInput) (SalesOrderLine, error) {
	orderedQty, err := normalizePositiveQuantity(input.OrderedQty)
	if err != nil {
		return SalesOrderLine{}, err
	}
	uomCode, err := decimal.NormalizeUOMCode(input.UOMCode)
	if err != nil {
		return SalesOrderLine{}, ErrSalesOrderInvalidQuantity
	}
	baseUOMCode := uomCode
	if strings.TrimSpace(input.BaseUOMCode) != "" {
		baseUOMCode, err = decimal.NormalizeUOMCode(input.BaseUOMCode)
		if err != nil {
			return SalesOrderLine{}, ErrSalesOrderInvalidQuantity
		}
	}
	baseOrderedQty := orderedQty
	if input.BaseOrderedQty != "" {
		baseOrderedQty, err = normalizePositiveQuantity(input.BaseOrderedQty)
		if err != nil {
			return SalesOrderLine{}, err
		}
	}
	conversionFactor := decimal.MustQuantity("1")
	if input.ConversionFactor != "" {
		conversionFactor, err = normalizePositiveQuantity(input.ConversionFactor)
		if err != nil {
			return SalesOrderLine{}, err
		}
	}
	unitPrice, err := normalizeNonNegativeUnitPrice(input.UnitPrice)
	if err != nil {
		return SalesOrderLine{}, err
	}
	currencyCode, err := normalizeSalesOrderCurrency(input.CurrencyCode)
	if err != nil {
		return SalesOrderLine{}, err
	}
	lineDiscountAmount, err := normalizeNonNegativeMoney(input.LineDiscountAmount)
	if err != nil {
		return SalesOrderLine{}, err
	}
	reservedQty, err := normalizeNonNegativeQuantity(input.ReservedQty)
	if err != nil {
		return SalesOrderLine{}, err
	}
	shippedQty, err := normalizeNonNegativeQuantity(input.ShippedQty)
	if err != nil {
		return SalesOrderLine{}, err
	}
	lineAmount, err := moneyFromQuantityUnitPrice(orderedQty, unitPrice, lineDiscountAmount)
	if err != nil {
		return SalesOrderLine{}, err
	}

	line := SalesOrderLine{
		ID:                 strings.TrimSpace(input.ID),
		LineNo:             input.LineNo,
		ItemID:             strings.TrimSpace(input.ItemID),
		SKUCode:            strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		ItemName:           strings.TrimSpace(input.ItemName),
		OrderedQty:         orderedQty,
		UOMCode:            uomCode,
		BaseOrderedQty:     baseOrderedQty,
		BaseUOMCode:        baseUOMCode,
		ConversionFactor:   conversionFactor,
		UnitPrice:          unitPrice,
		CurrencyCode:       currencyCode,
		LineDiscountAmount: lineDiscountAmount,
		LineAmount:         lineAmount,
		ReservedQty:        reservedQty,
		ShippedQty:         shippedQty,
		BatchID:            strings.TrimSpace(input.BatchID),
		BatchNo:            strings.ToUpper(strings.TrimSpace(input.BatchNo)),
	}
	if err := line.Validate(); err != nil {
		return SalesOrderLine{}, err
	}

	return line, nil
}

func (o SalesOrder) ReplaceDraftDetails(input UpdateSalesOrderDocumentInput) (SalesOrder, error) {
	if NormalizeSalesOrderStatus(o.Status) != SalesOrderStatusDraft {
		return SalesOrder{}, ErrSalesOrderInvalidTransition
	}
	if strings.TrimSpace(input.UpdatedBy) == "" {
		return SalesOrder{}, ErrSalesOrderTransitionActorRequired
	}

	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	updated := o.Clone()
	updated.CustomerID = strings.TrimSpace(input.CustomerID)
	updated.CustomerCode = strings.ToUpper(strings.TrimSpace(input.CustomerCode))
	updated.CustomerName = strings.TrimSpace(input.CustomerName)
	updated.Channel = strings.ToUpper(strings.TrimSpace(input.Channel))
	updated.WarehouseID = strings.TrimSpace(input.WarehouseID)
	updated.WarehouseCode = strings.ToUpper(strings.TrimSpace(input.WarehouseCode))
	updated.OrderDate = strings.TrimSpace(input.OrderDate)
	updated.Note = strings.TrimSpace(input.Note)
	updated.UpdatedAt = updatedAt.UTC()
	updated.UpdatedBy = strings.TrimSpace(input.UpdatedBy)
	if updated.Version > 0 {
		updated.Version++
	}

	if input.Lines != nil {
		updated.Lines = make([]SalesOrderLine, 0, len(input.Lines))
		for index, lineInput := range input.Lines {
			if lineInput.LineNo == 0 {
				lineInput.LineNo = index + 1
			}
			if lineInput.CurrencyCode == "" {
				lineInput.CurrencyCode = updated.CurrencyCode.String()
			}
			line, err := NewSalesOrderLine(lineInput)
			if err != nil {
				return SalesOrder{}, err
			}
			updated.Lines = append(updated.Lines, line)
		}
	}
	if err := updated.recalculateAmounts(); err != nil {
		return SalesOrder{}, err
	}
	if err := updated.Validate(); err != nil {
		return SalesOrder{}, err
	}

	return updated, nil
}

func (o SalesOrder) Validate() error {
	if strings.TrimSpace(o.ID) == "" ||
		strings.TrimSpace(o.OrgID) == "" ||
		strings.TrimSpace(o.OrderNo) == "" ||
		strings.TrimSpace(o.CustomerID) == "" ||
		strings.TrimSpace(o.CustomerName) == "" ||
		strings.TrimSpace(o.Channel) == "" ||
		strings.TrimSpace(o.OrderDate) == "" ||
		strings.TrimSpace(o.CreatedBy) == "" ||
		len(o.Lines) == 0 {
		return ErrSalesOrderRequiredField
	}
	if !IsValidSalesOrderStatus(o.Status) {
		return ErrSalesOrderInvalidStatus
	}
	if o.CurrencyCode != decimal.CurrencyVND {
		return ErrSalesOrderInvalidCurrency
	}
	for _, amount := range []decimal.Decimal{
		o.SubtotalAmount,
		o.DiscountAmount,
		o.TaxAmount,
		o.ShippingFeeAmount,
		o.NetAmount,
		o.TotalAmount,
	} {
		value, err := decimal.ParseMoneyAmount(amount.String())
		if err != nil || value.IsNegative() {
			return ErrSalesOrderInvalidAmount
		}
	}
	for _, line := range o.Lines {
		if err := line.Validate(); err != nil {
			return err
		}
	}

	return nil
}

func (l SalesOrderLine) Validate() error {
	if strings.TrimSpace(l.ID) == "" ||
		l.LineNo <= 0 ||
		strings.TrimSpace(l.ItemID) == "" ||
		strings.TrimSpace(l.SKUCode) == "" ||
		strings.TrimSpace(l.ItemName) == "" {
		return ErrSalesOrderRequiredField
	}
	if l.CurrencyCode != decimal.CurrencyVND {
		return ErrSalesOrderInvalidCurrency
	}
	for _, quantity := range []decimal.Decimal{l.OrderedQty, l.BaseOrderedQty, l.ConversionFactor} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() || value.IsZero() {
			return ErrSalesOrderInvalidQuantity
		}
	}
	for _, quantity := range []decimal.Decimal{l.ReservedQty, l.ShippedQty} {
		value, err := decimal.ParseQuantity(quantity.String())
		if err != nil || value.IsNegative() {
			return ErrSalesOrderInvalidQuantity
		}
	}
	if _, err := decimal.NormalizeUOMCode(l.UOMCode.String()); err != nil {
		return ErrSalesOrderInvalidQuantity
	}
	if _, err := decimal.NormalizeUOMCode(l.BaseUOMCode.String()); err != nil {
		return ErrSalesOrderInvalidQuantity
	}
	if _, err := normalizeNonNegativeUnitPrice(l.UnitPrice); err != nil {
		return err
	}
	for _, amount := range []decimal.Decimal{l.LineDiscountAmount, l.LineAmount} {
		value, err := decimal.ParseMoneyAmount(amount.String())
		if err != nil || value.IsNegative() {
			return ErrSalesOrderInvalidAmount
		}
	}

	return nil
}

func (o SalesOrder) Confirm(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusConfirmed, actorID, changedAt)
}

func (o SalesOrder) MarkReserved(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusReserved, actorID, changedAt)
}

func (o SalesOrder) MarkReservationFailed(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusReservationFailed, actorID, changedAt)
}

func (o SalesOrder) StartPicking(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPicking, actorID, changedAt)
}

func (o SalesOrder) MarkPicked(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPicked, actorID, changedAt)
}

func (o SalesOrder) MarkPickException(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPickException, actorID, changedAt)
}

func (o SalesOrder) StartPacking(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPacking, actorID, changedAt)
}

func (o SalesOrder) MarkPacked(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPacked, actorID, changedAt)
}

func (o SalesOrder) MarkPackException(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusPackException, actorID, changedAt)
}

func (o SalesOrder) MarkWaitingHandover(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusWaitingHandover, actorID, changedAt)
}

func (o SalesOrder) MarkHandedOver(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusHandedOver, actorID, changedAt)
}

func (o SalesOrder) MarkHandoverException(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusHandoverException, actorID, changedAt)
}

func (o SalesOrder) Close(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusClosed, actorID, changedAt)
}

func (o SalesOrder) Cancel(actorID string, changedAt time.Time) (SalesOrder, error) {
	return o.TransitionTo(SalesOrderStatusCancelled, actorID, changedAt)
}

func (o SalesOrder) CancelWithReason(actorID string, reason string, changedAt time.Time) (SalesOrder, error) {
	cancelled, err := o.Cancel(actorID, changedAt)
	if err != nil {
		return SalesOrder{}, err
	}
	cancelled.CancelReason = strings.TrimSpace(reason)

	return cancelled, nil
}

func (o SalesOrder) TransitionTo(status SalesOrderStatus, actorID string, changedAt time.Time) (SalesOrder, error) {
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return SalesOrder{}, ErrSalesOrderTransitionActorRequired
	}
	from := NormalizeSalesOrderStatus(o.Status)
	to := NormalizeSalesOrderStatus(status)
	if !IsValidSalesOrderStatus(from) || !IsValidSalesOrderStatus(to) {
		return SalesOrder{}, ErrSalesOrderInvalidStatus
	}
	if !CanTransitionSalesOrderStatus(from, to) {
		return SalesOrder{}, ErrSalesOrderInvalidTransition
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	updated := o
	updated.Status = to
	updated.UpdatedAt = changedAt.UTC()
	updated.UpdatedBy = actorID
	if updated.Version > 0 {
		updated.Version++
	}
	updated.markTransition(to, actorID, changedAt.UTC())

	return updated, nil
}

func (o *SalesOrder) markTransition(status SalesOrderStatus, actorID string, changedAt time.Time) {
	switch status {
	case SalesOrderStatusConfirmed:
		o.ConfirmedAt = changedAt
		o.ConfirmedBy = actorID
	case SalesOrderStatusReserved:
		o.ReservedAt = changedAt
		o.ReservedBy = actorID
	case SalesOrderStatusPicking:
		o.PickingStartedAt = changedAt
		o.PickingStartedBy = actorID
	case SalesOrderStatusPicked:
		o.PickedAt = changedAt
		o.PickedBy = actorID
	case SalesOrderStatusPacking:
		o.PackingStartedAt = changedAt
		o.PackingStartedBy = actorID
	case SalesOrderStatusPacked:
		o.PackedAt = changedAt
		o.PackedBy = actorID
	case SalesOrderStatusWaitingHandover:
		o.WaitingHandoverAt = changedAt
		o.WaitingHandoverBy = actorID
	case SalesOrderStatusHandedOver:
		o.HandedOverAt = changedAt
		o.HandedOverBy = actorID
	case SalesOrderStatusClosed:
		o.ClosedAt = changedAt
		o.ClosedBy = actorID
	case SalesOrderStatusCancelled:
		o.CancelledAt = changedAt
		o.CancelledBy = actorID
	case SalesOrderStatusReservationFailed, SalesOrderStatusPickException, SalesOrderStatusPackException, SalesOrderStatusHandoverException:
		o.ExceptionAt = changedAt
		o.ExceptionBy = actorID
	}
}

func (o SalesOrder) Clone() SalesOrder {
	clone := o
	clone.Lines = append([]SalesOrderLine(nil), o.Lines...)

	return clone
}

func (o *SalesOrder) recalculateAmounts() error {
	subtotal := decimal.MustMoneyAmount("0")
	discount := decimal.MustMoneyAmount("0")
	net := decimal.MustMoneyAmount("0")
	for _, line := range o.Lines {
		gross, err := addMoney(line.LineAmount, line.LineDiscountAmount)
		if err != nil {
			return err
		}
		subtotal, err = addMoney(subtotal, gross)
		if err != nil {
			return err
		}
		discount, err = addMoney(discount, line.LineDiscountAmount)
		if err != nil {
			return err
		}
		net, err = addMoney(net, line.LineAmount)
		if err != nil {
			return err
		}
	}

	o.SubtotalAmount = subtotal
	o.DiscountAmount = discount
	o.TaxAmount = decimal.MustMoneyAmount("0")
	o.ShippingFeeAmount = decimal.MustMoneyAmount("0")
	o.NetAmount = net
	o.TotalAmount = net

	return nil
}

func normalizeSalesOrderCurrency(value string) (decimal.CurrencyCode, error) {
	currencyCode, err := decimal.NormalizeCurrencyCode(value)
	if err != nil || currencyCode != decimal.CurrencyVND {
		return "", ErrSalesOrderInvalidCurrency
	}

	return currencyCode, nil
}

func normalizePositiveQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		return "", ErrSalesOrderInvalidQuantity
	}

	return quantity, nil
}

func normalizeNonNegativeQuantity(value decimal.Decimal) (decimal.Decimal, error) {
	quantity, err := decimal.ParseQuantity(value.String())
	if err != nil || quantity.IsNegative() {
		return "", ErrSalesOrderInvalidQuantity
	}

	return quantity, nil
}

func normalizeNonNegativeUnitPrice(value decimal.Decimal) (decimal.Decimal, error) {
	unitPrice, err := decimal.ParseUnitPrice(value.String())
	if err != nil || unitPrice.IsNegative() {
		return "", ErrSalesOrderInvalidAmount
	}

	return unitPrice, nil
}

func normalizeNonNegativeMoney(value decimal.Decimal) (decimal.Decimal, error) {
	money, err := decimal.ParseMoneyAmount(value.String())
	if err != nil || money.IsNegative() {
		return "", ErrSalesOrderInvalidAmount
	}

	return money, nil
}

func moneyFromQuantityUnitPrice(
	quantity decimal.Decimal,
	unitPrice decimal.Decimal,
	discount decimal.Decimal,
) (decimal.Decimal, error) {
	quantityValue, ok := new(big.Rat).SetString(quantity.String())
	if !ok {
		return "", ErrSalesOrderInvalidQuantity
	}
	unitPriceValue, ok := new(big.Rat).SetString(unitPrice.String())
	if !ok {
		return "", ErrSalesOrderInvalidAmount
	}
	discountValue, ok := new(big.Rat).SetString(discount.String())
	if !ok {
		return "", ErrSalesOrderInvalidAmount
	}

	amount := new(big.Rat).Mul(quantityValue, unitPriceValue)
	amount.Sub(amount, discountValue)
	if amount.Sign() < 0 {
		return "", ErrSalesOrderInvalidAmount
	}

	return roundRatToMoney(amount)
}

func roundRatToMoney(value *big.Rat) (decimal.Decimal, error) {
	if value.Sign() < 0 {
		return "", ErrSalesOrderInvalidAmount
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
		return "", ErrSalesOrderInvalidAmount
	}

	return money, nil
}

func addMoney(left decimal.Decimal, right decimal.Decimal) (decimal.Decimal, error) {
	leftValue, ok := new(big.Rat).SetString(left.String())
	if !ok {
		return "", ErrSalesOrderInvalidAmount
	}
	rightValue, ok := new(big.Rat).SetString(right.String())
	if !ok {
		return "", ErrSalesOrderInvalidAmount
	}
	sum := new(big.Rat).Add(leftValue, rightValue)
	if sum.Sign() < 0 {
		return "", ErrSalesOrderInvalidAmount
	}

	return roundRatToMoney(sum)
}

func NormalizeSalesOrderStatus(status SalesOrderStatus) SalesOrderStatus {
	return SalesOrderStatus(strings.ToLower(strings.TrimSpace(string(status))))
}

func IsValidSalesOrderStatus(status SalesOrderStatus) bool {
	switch NormalizeSalesOrderStatus(status) {
	case SalesOrderStatusDraft,
		SalesOrderStatusConfirmed,
		SalesOrderStatusReserved,
		SalesOrderStatusPicking,
		SalesOrderStatusPicked,
		SalesOrderStatusPacking,
		SalesOrderStatusPacked,
		SalesOrderStatusWaitingHandover,
		SalesOrderStatusHandedOver,
		SalesOrderStatusDelivered,
		SalesOrderStatusReturned,
		SalesOrderStatusClosed,
		SalesOrderStatusCancelled,
		SalesOrderStatusReservationFailed,
		SalesOrderStatusPickException,
		SalesOrderStatusPackException,
		SalesOrderStatusHandoverException:
		return true
	default:
		return false
	}
}

func CanTransitionSalesOrderStatus(from SalesOrderStatus, to SalesOrderStatus) bool {
	from = NormalizeSalesOrderStatus(from)
	to = NormalizeSalesOrderStatus(to)
	if from == to || !IsValidSalesOrderStatus(from) || !IsValidSalesOrderStatus(to) {
		return false
	}
	for _, candidate := range salesOrderTransitions[from] {
		if candidate == to {
			return true
		}
	}

	return false
}
