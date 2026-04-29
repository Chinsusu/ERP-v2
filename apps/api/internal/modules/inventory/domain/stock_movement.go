package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type MovementType string

const (
	MovementPurchaseReceipt    MovementType = "purchase_receipt"
	MovementQCRelease          MovementType = "qc_release"
	MovementSalesReserve       MovementType = "sales_reserve"
	MovementSalesUnreserve     MovementType = "sales_unreserve"
	MovementSalesIssue         MovementType = "sales_issue"
	MovementReturnReceipt      MovementType = "return_receipt"
	MovementReturnRestock      MovementType = "return_restock"
	MovementReturnToLab        MovementType = "return_to_lab"
	MovementSubcontractIssue   MovementType = "subcontract_issue"
	MovementSubcontractReceipt MovementType = "subcontract_receipt"
	MovementTransferIn         MovementType = "transfer_in"
	MovementTransferOut        MovementType = "transfer_out"
	MovementAdjustmentIn       MovementType = "adjustment_in"
	MovementAdjustmentOut      MovementType = "adjustment_out"

	MovementReceive MovementType = MovementPurchaseReceipt
	MovementIssue   MovementType = MovementSalesIssue
	MovementAdjust  MovementType = MovementAdjustmentIn
)

type Direction string

const (
	DirectionIn         Direction = "in"
	DirectionOut        Direction = "out"
	DirectionTransfer   Direction = "transfer"
	DirectionAdjustment Direction = "adjustment"
)

type StockStatus string

const (
	StockStatusAvailable         StockStatus = "available"
	StockStatusReserved          StockStatus = "reserved"
	StockStatusQCHold            StockStatus = "qc_hold"
	StockStatusReturnPending     StockStatus = "return_pending"
	StockStatusDamaged           StockStatus = "damaged"
	StockStatusSubcontractIssued StockStatus = "subcontract_issued"
)

type StockMovement struct {
	MovementNo       string
	MovementType     MovementType
	OrgID            string
	ItemID           string
	BatchID          string
	WarehouseID      string
	BinID            string
	UnitID           string
	Quantity         decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	SourceQuantity   decimal.Decimal
	SourceUOMCode    decimal.UOMCode
	ConversionFactor decimal.Decimal
	StockStatus      StockStatus
	SourceDocType    string
	SourceDocID      string
	SourceDocLineID  string
	Reason           string
	CreatedBy        string
	MovementAt       time.Time
	CreatedAt        time.Time
}

type NewStockMovementInput struct {
	MovementNo       string
	MovementType     MovementType
	OrgID            string
	ItemID           string
	BatchID          string
	WarehouseID      string
	BinID            string
	UnitID           string
	Quantity         decimal.Decimal
	BaseUOMCode      string
	SourceQuantity   decimal.Decimal
	SourceUOMCode    string
	ConversionFactor decimal.Decimal
	StockStatus      StockStatus
	SourceDocType    string
	SourceDocID      string
	SourceDocLineID  string
	Reason           string
	CreatedBy        string
	MovementAt       time.Time
}

type BalanceDelta struct {
	OnHand    decimal.Decimal
	Reserved  decimal.Decimal
	Available decimal.Decimal
}

type movementEffect struct {
	direction  Direction
	onHand     int
	reserved   int
	available  int
	adjustment bool
}

var movementEffects = map[MovementType]movementEffect{
	MovementPurchaseReceipt:    {direction: DirectionIn, onHand: 1, available: 1},
	MovementQCRelease:          {direction: DirectionIn, available: 1},
	MovementSalesReserve:       {direction: DirectionTransfer, reserved: 1, available: -1},
	MovementSalesUnreserve:     {direction: DirectionTransfer, reserved: -1, available: 1},
	MovementSalesIssue:         {direction: DirectionOut, onHand: -1, available: -1},
	MovementReturnReceipt:      {direction: DirectionIn, onHand: 1},
	MovementReturnRestock:      {direction: DirectionIn, onHand: 1, available: 1},
	MovementReturnToLab:        {direction: DirectionOut, onHand: -1},
	MovementSubcontractIssue:   {direction: DirectionOut, onHand: -1, available: -1},
	MovementSubcontractReceipt: {direction: DirectionIn, onHand: 1},
	MovementTransferIn:         {direction: DirectionTransfer, onHand: 1, available: 1},
	MovementTransferOut:        {direction: DirectionTransfer, onHand: -1, available: -1},
	MovementAdjustmentIn:       {direction: DirectionAdjustment, onHand: 1, available: 1, adjustment: true},
	MovementAdjustmentOut:      {direction: DirectionAdjustment, onHand: -1, available: -1, adjustment: true},
}

var movementTypeOrder = []MovementType{
	MovementPurchaseReceipt,
	MovementQCRelease,
	MovementSalesReserve,
	MovementSalesUnreserve,
	MovementSalesIssue,
	MovementReturnReceipt,
	MovementReturnRestock,
	MovementReturnToLab,
	MovementSubcontractIssue,
	MovementSubcontractReceipt,
	MovementTransferIn,
	MovementTransferOut,
	MovementAdjustmentIn,
	MovementAdjustmentOut,
}

var stockStatuses = map[StockStatus]struct{}{
	StockStatusAvailable:         {},
	StockStatusReserved:          {},
	StockStatusQCHold:            {},
	StockStatusReturnPending:     {},
	StockStatusDamaged:           {},
	StockStatusSubcontractIssued: {},
}

func NewStockMovement(input NewStockMovementInput) (StockMovement, error) {
	now := time.Now().UTC()
	movementAt := input.MovementAt
	if movementAt.IsZero() {
		movementAt = now
	}

	movementType := MovementType(strings.TrimSpace(string(input.MovementType)))
	stockStatus := StockStatus(strings.TrimSpace(string(input.StockStatus)))
	if stockStatus == "" {
		stockStatus = StockStatusAvailable
	}
	quantity, err := decimal.ParseQuantity(input.Quantity.String())
	if err != nil {
		return StockMovement{}, err
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return StockMovement{}, err
	}
	sourceQuantity := input.SourceQuantity
	if strings.TrimSpace(sourceQuantity.String()) == "" {
		sourceQuantity = quantity
	}
	sourceQuantity, err = decimal.ParseQuantity(sourceQuantity.String())
	if err != nil {
		return StockMovement{}, err
	}
	sourceUOMCode := strings.TrimSpace(input.SourceUOMCode)
	if sourceUOMCode == "" {
		sourceUOMCode = baseUOMCode.String()
	}
	normalizedSourceUOMCode, err := decimal.NormalizeUOMCode(sourceUOMCode)
	if err != nil {
		return StockMovement{}, err
	}
	conversionFactor := input.ConversionFactor
	if strings.TrimSpace(conversionFactor.String()) == "" {
		conversionFactor = decimal.MustQuantity("1")
	}
	conversionFactor, err = decimal.ParseQuantity(conversionFactor.String())
	if err != nil {
		return StockMovement{}, err
	}

	movement := StockMovement{
		MovementNo:       strings.TrimSpace(input.MovementNo),
		MovementType:     movementType,
		OrgID:            strings.TrimSpace(input.OrgID),
		ItemID:           strings.TrimSpace(input.ItemID),
		BatchID:          strings.TrimSpace(input.BatchID),
		WarehouseID:      strings.TrimSpace(input.WarehouseID),
		BinID:            strings.TrimSpace(input.BinID),
		UnitID:           strings.TrimSpace(input.UnitID),
		Quantity:         quantity,
		BaseUOMCode:      baseUOMCode,
		SourceQuantity:   sourceQuantity,
		SourceUOMCode:    normalizedSourceUOMCode,
		ConversionFactor: conversionFactor,
		StockStatus:      stockStatus,
		SourceDocType:    strings.TrimSpace(input.SourceDocType),
		SourceDocID:      strings.TrimSpace(input.SourceDocID),
		SourceDocLineID:  strings.TrimSpace(input.SourceDocLineID),
		Reason:           strings.TrimSpace(input.Reason),
		CreatedBy:        strings.TrimSpace(input.CreatedBy),
		MovementAt:       movementAt.UTC(),
		CreatedAt:        now,
	}

	if err := movement.Validate(); err != nil {
		return StockMovement{}, err
	}

	return movement, nil
}

func MovementTypes() []MovementType {
	return append([]MovementType(nil), movementTypeOrder...)
}

func (m StockMovement) Validate() error {
	if m.MovementNo == "" {
		return errors.New("movement no is required")
	}
	if _, ok := movementEffects[m.MovementType]; !ok {
		return fmt.Errorf("movement type %q is not supported", m.MovementType)
	}
	if m.OrgID == "" {
		return errors.New("org id is required")
	}
	if m.ItemID == "" {
		return errors.New("item id is required")
	}
	if m.WarehouseID == "" {
		return errors.New("warehouse id is required")
	}
	quantity, err := decimal.ParseQuantity(m.Quantity.String())
	if err != nil || quantity.IsNegative() || quantity.IsZero() {
		return errors.New("quantity must be positive")
	}
	if _, err := decimal.NormalizeUOMCode(m.BaseUOMCode.String()); err != nil {
		return errors.New("base uom code is required")
	}
	sourceQuantity, err := decimal.ParseQuantity(m.SourceQuantity.String())
	if err != nil || sourceQuantity.IsNegative() || sourceQuantity.IsZero() {
		return errors.New("source quantity must be positive")
	}
	if _, err := decimal.NormalizeUOMCode(m.SourceUOMCode.String()); err != nil {
		return errors.New("source uom code is required")
	}
	conversionFactor, err := decimal.ParseQuantity(m.ConversionFactor.String())
	if err != nil || conversionFactor.IsNegative() || conversionFactor.IsZero() {
		return errors.New("conversion factor must be positive")
	}
	if _, ok := stockStatuses[m.StockStatus]; !ok {
		return fmt.Errorf("stock status %q is not supported", m.StockStatus)
	}
	if m.SourceDocType == "" {
		return errors.New("source doc type is required")
	}
	if m.SourceDocID == "" {
		return errors.New("source doc id is required")
	}
	if m.Reason == "" {
		return errors.New("reason is required")
	}
	if m.CreatedBy == "" {
		return errors.New("created by is required")
	}

	return nil
}

func (m StockMovement) Direction() (Direction, error) {
	effect, ok := movementEffects[m.MovementType]
	if !ok {
		return "", fmt.Errorf("movement type %q is not supported", m.MovementType)
	}

	return effect.direction, nil
}

func (m StockMovement) BalanceDelta() (BalanceDelta, error) {
	effect, ok := movementEffects[m.MovementType]
	if !ok {
		return BalanceDelta{}, fmt.Errorf("movement type %q is not supported", m.MovementType)
	}

	return BalanceDelta{
		OnHand:    signedQuantity(m.Quantity, effect.onHand),
		Reserved:  signedQuantity(m.Quantity, effect.reserved),
		Available: signedQuantity(m.Quantity, effect.available),
	}, nil
}

func (m StockMovement) IsAdjustment() bool {
	effect, ok := movementEffects[m.MovementType]
	return ok && effect.adjustment
}

func signedQuantity(quantity decimal.Decimal, sign int) decimal.Decimal {
	switch sign {
	case -1:
		return decimal.Decimal("-" + strings.TrimPrefix(quantity.String(), "-"))
	case 1:
		return decimal.Decimal(strings.TrimPrefix(quantity.String(), "+"))
	default:
		return decimal.MustQuantity("0")
	}
}
