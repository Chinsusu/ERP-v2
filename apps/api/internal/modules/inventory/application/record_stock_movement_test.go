package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestRecordStockMovementAppendsMovement(t *testing.T) {
	store := NewInMemoryStockMovementStore()
	useCase := NewRecordStockMovement(store)

	movement, err := newTestMovement(domain.MovementPurchaseReceipt)
	if err != nil {
		t.Fatalf("new movement: %v", err)
	}

	if err := useCase.Execute(context.Background(), movement); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if got := store.Count(); got != 1 {
		t.Fatalf("movement count = %d, want 1", got)
	}
}

func TestRecordStockMovementRejectsInvalidMovement(t *testing.T) {
	store := NewInMemoryStockMovementStore()
	useCase := NewRecordStockMovement(store)

	if err := useCase.Execute(context.Background(), domain.StockMovement{}); err == nil {
		t.Fatal("execute error = nil, want validation error")
	}
	if got := store.Count(); got != 0 {
		t.Fatalf("movement count = %d, want 0", got)
	}
}

func newTestMovement(movementType domain.MovementType, mutators ...func(*domain.NewStockMovementInput)) (domain.StockMovement, error) {
	input := domain.NewStockMovementInput{
		MovementNo:    "MOV-20260426-0001",
		MovementType:  movementType,
		OrgID:         "11111111-1111-1111-1111-111111111111",
		ItemID:        "22222222-2222-2222-2222-222222222222",
		WarehouseID:   "33333333-3333-3333-3333-333333333333",
		UnitID:        "44444444-4444-4444-4444-444444444444",
		Quantity:      decimal.MustQuantity("10"),
		BaseUOMCode:   "PCS",
		StockStatus:   domain.StockStatusAvailable,
		SourceDocType: "goods_receipt",
		SourceDocID:   "55555555-5555-5555-5555-555555555555",
		Reason:        "initial receive",
		CreatedBy:     "66666666-6666-6666-6666-666666666666",
	}
	for _, mutate := range mutators {
		mutate(&input)
	}

	return domain.NewStockMovement(input)
}
