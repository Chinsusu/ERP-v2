package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

func TestListAvailableStockCalculatesPrototypeRows(t *testing.T) {
	useCase := NewListAvailableStock(NewPrototypeStockAvailabilityStore())

	snapshots, err := useCase.Execute(context.Background(), domain.AvailableStockFilter{})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(snapshots) != 3 {
		t.Fatalf("snapshots length = %d, want 3", len(snapshots))
	}

	serum, ok := findSnapshot(snapshots, "SERUM-30ML")
	if !ok {
		t.Fatal("SERUM-30ML snapshot not found")
	}
	if serum.AvailableQty != "110.000000" {
		t.Fatalf("SERUM-30ML available qty = %q, want 110.000000", serum.AvailableQty)
	}
}

func TestListAvailableStockFiltersByWarehouseLocationSKUAndBatch(t *testing.T) {
	useCase := NewListAvailableStock(NewPrototypeStockAvailabilityStore())
	filter := domain.NewAvailableStockFilter("wh-hn", "bin-hn-r01", "toner-100ml", "batch-toner-2604c")

	snapshots, err := useCase.Execute(context.Background(), filter)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(snapshots) != 1 {
		t.Fatalf("snapshots length = %d, want 1", len(snapshots))
	}

	got := snapshots[0]
	if got.WarehouseID != "wh-hn" || got.LocationID != "bin-hn-r01" || got.SKU != "TONER-100ML" || got.BatchID != "batch-toner-2604c" {
		t.Fatalf("snapshot = %+v, want filtered HN toner batch", got)
	}
	if got.PhysicalQty != "90.000000" || got.ReservedQty != "20.000000" || got.BlockedQty != "5.000000" || got.AvailableQty != "65.000000" {
		t.Fatalf("snapshot quantities = %+v, want physical 90 reserved 20 blocked 5 available 65", got)
	}
}

func TestListAvailableStockRequiresStore(t *testing.T) {
	useCase := NewListAvailableStock(nil)

	if _, err := useCase.Execute(context.Background(), domain.AvailableStockFilter{}); err == nil {
		t.Fatal("execute error = nil, want missing store error")
	}
}

func findSnapshot(snapshots []domain.AvailableStockSnapshot, sku string) (domain.AvailableStockSnapshot, bool) {
	for _, snapshot := range snapshots {
		if snapshot.SKU == sku {
			return snapshot, true
		}
	}

	return domain.AvailableStockSnapshot{}, false
}
