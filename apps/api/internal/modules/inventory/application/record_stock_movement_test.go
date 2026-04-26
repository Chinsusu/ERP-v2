package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

func TestRecordStockMovementAppendsMovement(t *testing.T) {
	store := NewInMemoryStockMovementStore()
	useCase := NewRecordStockMovement(store)

	movement, err := domain.NewStockMovement("mov-1", "SKU-001", "WH-001", domain.MovementReceive, 10, "initial receive")
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
