package domain

import (
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestMovementTypesIncludeOperationalStockLedgerTypes(t *testing.T) {
	types := MovementTypes()
	if len(types) < 10 {
		t.Fatalf("movement types count = %d, want operational catalog", len(types))
	}

	assertHasMovementType(t, types, MovementPurchaseReceipt)
	assertHasMovementType(t, types, MovementSalesReserve)
	assertHasMovementType(t, types, MovementSalesIssue)
	assertHasMovementType(t, types, MovementReturnRestock)
	assertHasMovementType(t, types, MovementSubcontractIssue)
	assertHasMovementType(t, types, MovementAdjustmentIn)
	assertHasMovementType(t, types, MovementAdjustmentOut)
}

func TestStockMovementRequiresSourceDocument(t *testing.T) {
	_, err := NewStockMovement(validMovementInput(func(input *NewStockMovementInput) {
		input.SourceDocID = ""
	}))
	if err == nil {
		t.Fatal("new movement error = nil, want source doc validation")
	}
}

func TestStockMovementBalanceDeltaForReserve(t *testing.T) {
	movement, err := NewStockMovement(validMovementInput(func(input *NewStockMovementInput) {
		input.MovementType = MovementSalesReserve
		input.Quantity = decimal.MustQuantity("7")
	}))
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	direction, err := movement.Direction()
	if err != nil {
		t.Fatalf("direction: %v", err)
	}
	if direction != DirectionTransfer {
		t.Fatalf("direction = %q, want %q", direction, DirectionTransfer)
	}

	delta, err := movement.BalanceDelta()
	if err != nil {
		t.Fatalf("balance delta: %v", err)
	}
	if delta.OnHand != "0.000000" || delta.Reserved != "7.000000" || delta.Available != "-7.000000" {
		t.Fatalf("delta = %+v, want reserved +7 and available -7", delta)
	}
}

func TestStockMovementAdjustmentIsAuditable(t *testing.T) {
	movement, err := NewStockMovement(validMovementInput(func(input *NewStockMovementInput) {
		input.MovementType = MovementAdjustmentOut
		input.Quantity = decimal.MustQuantity("4")
	}))
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if !movement.IsAdjustment() {
		t.Fatal("adjustment movement is not marked auditable")
	}

	direction, err := movement.Direction()
	if err != nil {
		t.Fatalf("direction: %v", err)
	}
	if direction != DirectionAdjustment {
		t.Fatalf("direction = %q, want %q", direction, DirectionAdjustment)
	}

	delta, err := movement.BalanceDelta()
	if err != nil {
		t.Fatalf("balance delta: %v", err)
	}
	if delta.OnHand != "-4.000000" || delta.Available != "-4.000000" {
		t.Fatalf("delta = %+v, want on hand -4 and available -4", delta)
	}
}

func validMovementInput(mutators ...func(*NewStockMovementInput)) NewStockMovementInput {
	input := NewStockMovementInput{
		MovementNo:    "MOV-20260426-0001",
		MovementType:  MovementPurchaseReceipt,
		OrgID:         "11111111-1111-1111-1111-111111111111",
		ItemID:        "22222222-2222-2222-2222-222222222222",
		WarehouseID:   "33333333-3333-3333-3333-333333333333",
		UnitID:        "44444444-4444-4444-4444-444444444444",
		Quantity:      decimal.MustQuantity("10"),
		BaseUOMCode:   "PCS",
		StockStatus:   StockStatusAvailable,
		SourceDocType: "goods_receipt",
		SourceDocID:   "55555555-5555-5555-5555-555555555555",
		Reason:        "unit test movement",
		CreatedBy:     "66666666-6666-6666-6666-666666666666",
	}
	for _, mutate := range mutators {
		mutate(&input)
	}

	return input
}

func assertHasMovementType(t *testing.T, types []MovementType, want MovementType) {
	t.Helper()
	for _, movementType := range types {
		if movementType == want {
			return
		}
	}

	t.Fatalf("movement type %q not found in catalog", want)
}
