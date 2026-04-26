package domain

import (
	"errors"
	"time"
)

type MovementType string

const (
	MovementReceive    MovementType = "RECEIVE"
	MovementIssue      MovementType = "ISSUE"
	MovementTransferIn MovementType = "TRANSFER_IN"
	MovementAdjust     MovementType = "ADJUST"
)

type StockMovement struct {
	ID           string
	SKU          string
	WarehouseID  string
	MovementType MovementType
	Quantity     int64
	Reason       string
	CreatedAt    time.Time
}

func NewStockMovement(id string, sku string, warehouseID string, movementType MovementType, quantity int64, reason string) (StockMovement, error) {
	if id == "" {
		return StockMovement{}, errors.New("movement id is required")
	}
	if sku == "" {
		return StockMovement{}, errors.New("sku is required")
	}
	if warehouseID == "" {
		return StockMovement{}, errors.New("warehouse id is required")
	}
	if quantity <= 0 {
		return StockMovement{}, errors.New("quantity must be positive")
	}
	if reason == "" {
		return StockMovement{}, errors.New("reason is required")
	}

	return StockMovement{
		ID:           id,
		SKU:          sku,
		WarehouseID:  warehouseID,
		MovementType: movementType,
		Quantity:     quantity,
		Reason:       reason,
		CreatedAt:    time.Now().UTC(),
	}, nil
}
