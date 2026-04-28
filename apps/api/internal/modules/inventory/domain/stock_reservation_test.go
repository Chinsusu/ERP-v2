package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewStockReservationNormalizesSalesOrderStockLink(t *testing.T) {
	reservation, err := NewStockReservation(validReservationInput(func(input *NewStockReservationInput) {
		input.SKUCode = " serum-30ml "
		input.BatchNo = " lot-a "
		input.WarehouseCode = " wh-hcm-fg "
		input.BinCode = " pick-a-01 "
		input.ReservedQty = decimal.MustQuantity("3")
		input.BaseUOMCode = "ea"
	}))
	if err != nil {
		t.Fatalf("new reservation: %v", err)
	}

	if reservation.Status != ReservationStatusActive || !reservation.IsActive() {
		t.Fatalf("status = %q, want active", reservation.Status)
	}
	if reservation.SalesOrderLineID != "line-1" || reservation.SKUCode != "SERUM-30ML" {
		t.Fatalf("reservation link = %+v, want sales order line and normalized SKU", reservation)
	}
	if reservation.BatchID != "batch-1" || reservation.BinID != "bin-pick-a-01" {
		t.Fatalf("batch/bin = %s/%s, want required stock location", reservation.BatchID, reservation.BinID)
	}
	if reservation.ReservedQty != "3.000000" || reservation.BaseUOMCode != "EA" {
		t.Fatalf("qty/uom = %s/%s, want base quantity precision", reservation.ReservedQty, reservation.BaseUOMCode)
	}
}

func TestStockReservationRequiresSalesLineBatchLocationAndPositiveQuantity(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*NewStockReservationInput)
		want   error
	}{
		{
			name: "sales order line",
			mutate: func(input *NewStockReservationInput) {
				input.SalesOrderLineID = ""
			},
			want: ErrStockReservationRequiredField,
		},
		{
			name: "batch",
			mutate: func(input *NewStockReservationInput) {
				input.BatchID = ""
			},
			want: ErrStockReservationRequiredField,
		},
		{
			name: "bin",
			mutate: func(input *NewStockReservationInput) {
				input.BinID = ""
			},
			want: ErrStockReservationRequiredField,
		},
		{
			name: "quantity",
			mutate: func(input *NewStockReservationInput) {
				input.ReservedQty = decimal.MustQuantity("0")
			},
			want: ErrStockReservationInvalidQuantity,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewStockReservation(validReservationInput(tc.mutate))
			if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestStockReservationOnlySupportsActiveReleasedConsumedStatuses(t *testing.T) {
	if !IsValidReservationStatus(ReservationStatusActive) ||
		!IsValidReservationStatus(ReservationStatusReleased) ||
		!IsValidReservationStatus(ReservationStatusConsumed) {
		t.Fatal("active, released, and consumed should be valid reservation statuses")
	}
	if IsValidReservationStatus("cancelled") || IsValidReservationStatus("expired") {
		t.Fatal("cancelled and expired should not be valid reservation statuses")
	}

	_, err := NewStockReservation(validReservationInput(func(input *NewStockReservationInput) {
		input.Status = "cancelled"
	}))
	if !errors.Is(err, ErrStockReservationInvalidStatus) {
		t.Fatalf("error = %v, want invalid status", err)
	}
}

func TestStockReservationReleaseAndConsumeTransitions(t *testing.T) {
	reservation, err := NewStockReservation(validReservationInput())
	if err != nil {
		t.Fatalf("new reservation: %v", err)
	}
	changedAt := time.Date(2026, 4, 28, 10, 30, 0, 0, time.UTC)

	released, err := reservation.Release("sales-ops", changedAt)
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if released.Status != ReservationStatusReleased || released.ReleasedBy != "sales-ops" || released.ReleasedAt != changedAt {
		t.Fatalf("released reservation = %+v, want released metadata", released)
	}
	if _, err := released.Consume("warehouse-ops", changedAt); !errors.Is(err, ErrStockReservationInvalidTransition) {
		t.Fatalf("consume released error = %v, want invalid transition", err)
	}

	consumed, err := reservation.Consume("warehouse-ops", changedAt)
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if consumed.Status != ReservationStatusConsumed || consumed.ConsumedBy != "warehouse-ops" || consumed.ConsumedAt != changedAt {
		t.Fatalf("consumed reservation = %+v, want consumed metadata", consumed)
	}
	if _, err := reservation.Release("", changedAt); !errors.Is(err, ErrStockReservationActorRequired) {
		t.Fatalf("release actor error = %v, want actor required", err)
	}
}

func TestNewStockReservationRehydratesReleasedRecord(t *testing.T) {
	releasedAt := time.Date(2026, 4, 28, 11, 0, 0, 0, time.UTC)
	reservation, err := NewStockReservation(validReservationInput(func(input *NewStockReservationInput) {
		input.Status = ReservationStatusReleased
		input.ReleasedAt = releasedAt
		input.ReleasedBy = "sales-ops"
	}))
	if err != nil {
		t.Fatalf("new released reservation: %v", err)
	}

	if reservation.Status != ReservationStatusReleased || reservation.ReleasedAt != releasedAt || reservation.ReleasedBy != "sales-ops" {
		t.Fatalf("reservation = %+v, want released record metadata", reservation)
	}
}

func validReservationInput(mutators ...func(*NewStockReservationInput)) NewStockReservationInput {
	input := NewStockReservationInput{
		ID:               "reservation-1",
		OrgID:            "org-my-pham",
		ReservationNo:    "RSV-260428-0001",
		SalesOrderID:     "sales-order-1",
		SalesOrderLineID: "line-1",
		ItemID:           "item-serum-30ml",
		SKUCode:          "SERUM-30ML",
		BatchID:          "batch-1",
		BatchNo:          "LOT-A",
		WarehouseID:      "wh-hcm-fg",
		WarehouseCode:    "WH-HCM-FG",
		BinID:            "bin-pick-a-01",
		BinCode:          "PICK-A-01",
		ReservedQty:      decimal.MustQuantity("2"),
		BaseUOMCode:      "EA",
		ReservedAt:       time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC),
		ReservedBy:       "sales-ops",
	}
	for _, mutate := range mutators {
		mutate(&input)
	}

	return input
}
