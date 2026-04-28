package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

var ErrStockReservationRequiredField = errors.New("stock reservation required field is missing")
var ErrStockReservationInvalidQuantity = errors.New("stock reservation quantity is invalid")
var ErrStockReservationInvalidStatus = errors.New("stock reservation status is invalid")
var ErrStockReservationInvalidTransition = errors.New("stock reservation status transition is invalid")
var ErrStockReservationActorRequired = errors.New("stock reservation actor is required")

type ReservationStatus string

const (
	ReservationStatusActive   ReservationStatus = "active"
	ReservationStatusReleased ReservationStatus = "released"
	ReservationStatusConsumed ReservationStatus = "consumed"
)

type StockReservation struct {
	ID               string
	OrgID            string
	ReservationNo    string
	SalesOrderID     string
	SalesOrderLineID string
	ItemID           string
	SKUCode          string
	BatchID          string
	BatchNo          string
	WarehouseID      string
	WarehouseCode    string
	BinID            string
	BinCode          string
	StockStatus      StockStatus
	ReservedQty      decimal.Decimal
	BaseUOMCode      decimal.UOMCode
	Status           ReservationStatus
	ReservedAt       time.Time
	ReservedBy       string
	ReleasedAt       time.Time
	ReleasedBy       string
	ConsumedAt       time.Time
	ConsumedBy       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type NewStockReservationInput struct {
	ID               string
	OrgID            string
	ReservationNo    string
	SalesOrderID     string
	SalesOrderLineID string
	ItemID           string
	SKUCode          string
	BatchID          string
	BatchNo          string
	WarehouseID      string
	WarehouseCode    string
	BinID            string
	BinCode          string
	StockStatus      StockStatus
	ReservedQty      decimal.Decimal
	BaseUOMCode      string
	Status           ReservationStatus
	ReservedAt       time.Time
	ReservedBy       string
	ReleasedAt       time.Time
	ReleasedBy       string
	ConsumedAt       time.Time
	ConsumedBy       string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewStockReservation(input NewStockReservationInput) (StockReservation, error) {
	now := time.Now().UTC()
	reservedAt := input.ReservedAt
	if reservedAt.IsZero() {
		reservedAt = now
	}
	createdAt := input.CreatedAt
	if createdAt.IsZero() {
		createdAt = reservedAt
	}
	updatedAt := input.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	status := NormalizeReservationStatus(input.Status)
	if status == "" {
		status = ReservationStatusActive
	}
	stockStatus := StockStatus(strings.TrimSpace(string(input.StockStatus)))
	if stockStatus == "" {
		stockStatus = StockStatusAvailable
	}
	reservedQty, err := decimal.ParseQuantity(input.ReservedQty.String())
	if err != nil || reservedQty.IsNegative() || reservedQty.IsZero() {
		return StockReservation{}, ErrStockReservationInvalidQuantity
	}
	baseUOMCode, err := decimal.NormalizeUOMCode(input.BaseUOMCode)
	if err != nil {
		return StockReservation{}, ErrStockReservationRequiredField
	}

	reservation := StockReservation{
		ID:               strings.TrimSpace(input.ID),
		OrgID:            strings.TrimSpace(input.OrgID),
		ReservationNo:    strings.TrimSpace(input.ReservationNo),
		SalesOrderID:     strings.TrimSpace(input.SalesOrderID),
		SalesOrderLineID: strings.TrimSpace(input.SalesOrderLineID),
		ItemID:           strings.TrimSpace(input.ItemID),
		SKUCode:          strings.ToUpper(strings.TrimSpace(input.SKUCode)),
		BatchID:          strings.TrimSpace(input.BatchID),
		BatchNo:          strings.ToUpper(strings.TrimSpace(input.BatchNo)),
		WarehouseID:      strings.TrimSpace(input.WarehouseID),
		WarehouseCode:    strings.ToUpper(strings.TrimSpace(input.WarehouseCode)),
		BinID:            strings.TrimSpace(input.BinID),
		BinCode:          strings.ToUpper(strings.TrimSpace(input.BinCode)),
		StockStatus:      stockStatus,
		ReservedQty:      reservedQty,
		BaseUOMCode:      baseUOMCode,
		Status:           status,
		ReservedAt:       reservedAt.UTC(),
		ReservedBy:       strings.TrimSpace(input.ReservedBy),
		ReleasedAt:       input.ReleasedAt.UTC(),
		ReleasedBy:       strings.TrimSpace(input.ReleasedBy),
		ConsumedAt:       input.ConsumedAt.UTC(),
		ConsumedBy:       strings.TrimSpace(input.ConsumedBy),
		CreatedAt:        createdAt.UTC(),
		UpdatedAt:        updatedAt.UTC(),
	}
	if err := reservation.Validate(); err != nil {
		return StockReservation{}, err
	}

	return reservation, nil
}

func NormalizeReservationStatus(value ReservationStatus) ReservationStatus {
	return ReservationStatus(strings.ToLower(strings.TrimSpace(string(value))))
}

func IsValidReservationStatus(value ReservationStatus) bool {
	switch NormalizeReservationStatus(value) {
	case ReservationStatusActive, ReservationStatusReleased, ReservationStatusConsumed:
		return true
	default:
		return false
	}
}

func (r StockReservation) Validate() error {
	if strings.TrimSpace(r.ID) == "" ||
		strings.TrimSpace(r.OrgID) == "" ||
		strings.TrimSpace(r.ReservationNo) == "" ||
		strings.TrimSpace(r.SalesOrderID) == "" ||
		strings.TrimSpace(r.SalesOrderLineID) == "" ||
		strings.TrimSpace(r.ItemID) == "" ||
		strings.TrimSpace(r.SKUCode) == "" ||
		strings.TrimSpace(r.BatchID) == "" ||
		strings.TrimSpace(r.WarehouseID) == "" ||
		strings.TrimSpace(r.BinID) == "" ||
		strings.TrimSpace(r.ReservedBy) == "" {
		return ErrStockReservationRequiredField
	}
	if !IsValidReservationStatus(r.Status) {
		return ErrStockReservationInvalidStatus
	}
	if _, ok := stockStatuses[r.StockStatus]; !ok {
		return ErrStockReservationRequiredField
	}
	if _, err := decimal.NormalizeUOMCode(r.BaseUOMCode.String()); err != nil {
		return ErrStockReservationRequiredField
	}
	reservedQty, err := decimal.ParseQuantity(r.ReservedQty.String())
	if err != nil || reservedQty.IsNegative() || reservedQty.IsZero() {
		return ErrStockReservationInvalidQuantity
	}
	if r.ReservedAt.IsZero() || r.CreatedAt.IsZero() || r.UpdatedAt.IsZero() {
		return ErrStockReservationRequiredField
	}
	if r.Status == ReservationStatusActive && (!r.ReleasedAt.IsZero() || !r.ConsumedAt.IsZero()) {
		return ErrStockReservationInvalidStatus
	}
	if r.Status == ReservationStatusReleased && (r.ReleasedAt.IsZero() || strings.TrimSpace(r.ReleasedBy) == "") {
		return ErrStockReservationInvalidStatus
	}
	if r.Status == ReservationStatusConsumed && (r.ConsumedAt.IsZero() || strings.TrimSpace(r.ConsumedBy) == "") {
		return ErrStockReservationInvalidStatus
	}

	return nil
}

func (r StockReservation) IsActive() bool {
	return NormalizeReservationStatus(r.Status) == ReservationStatusActive
}

func (r StockReservation) Release(actorID string, releasedAt time.Time) (StockReservation, error) {
	if !r.IsActive() {
		return StockReservation{}, ErrStockReservationInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return StockReservation{}, ErrStockReservationActorRequired
	}
	if releasedAt.IsZero() {
		releasedAt = time.Now().UTC()
	}

	released := r
	released.Status = ReservationStatusReleased
	released.ReleasedAt = releasedAt.UTC()
	released.ReleasedBy = actorID
	released.UpdatedAt = releasedAt.UTC()

	return released, nil
}

func (r StockReservation) Consume(actorID string, consumedAt time.Time) (StockReservation, error) {
	if !r.IsActive() {
		return StockReservation{}, ErrStockReservationInvalidTransition
	}
	actorID = strings.TrimSpace(actorID)
	if actorID == "" {
		return StockReservation{}, ErrStockReservationActorRequired
	}
	if consumedAt.IsZero() {
		consumedAt = time.Now().UTC()
	}

	consumed := r
	consumed.Status = ReservationStatusConsumed
	consumed.ConsumedAt = consumedAt.UTC()
	consumed.ConsumedBy = actorID
	consumed.UpdatedAt = consumedAt.UTC()

	return consumed, nil
}
