package domain

import "testing"

func TestCalculateAvailableStockUsesPhysicalReservedAndHold(t *testing.T) {
	rows := []StockBalanceSnapshot{
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			SKU:           "serum-30ml",
			BatchID:       "batch-a",
			BatchNo:       "LOT-A",
			StockStatus:   StockStatusAvailable,
			QtyOnHand:     120,
			QtyReserved:   15,
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-a",
			BatchNo:       "LOT-A",
			StockStatus:   StockStatusQCHold,
			QtyOnHand:     8,
			QtyReserved:   0,
		},
	}

	snapshots := CalculateAvailableStock(rows)
	if len(snapshots) != 1 {
		t.Fatalf("snapshots length = %d, want 1", len(snapshots))
	}

	got := snapshots[0]
	if got.PhysicalStock != 128 {
		t.Fatalf("physical stock = %d, want 128", got.PhysicalStock)
	}
	if got.ReservedStock != 15 {
		t.Fatalf("reserved stock = %d, want 15", got.ReservedStock)
	}
	if got.HoldStock != 8 {
		t.Fatalf("hold stock = %d, want 8", got.HoldStock)
	}
	if got.AvailableStock != 105 {
		t.Fatalf("available stock = %d, want 105", got.AvailableStock)
	}
}

func TestCalculateAvailableStockGroupsByWarehouseSKUAndBatch(t *testing.T) {
	rows := []StockBalanceSnapshot{
		{WarehouseID: "wh-hcm", WarehouseCode: "HCM", SKU: "SKU-01", BatchID: "batch-a", BatchNo: "LOT-A", StockStatus: StockStatusAvailable, QtyOnHand: 5},
		{WarehouseID: "wh-hcm", WarehouseCode: "HCM", SKU: "SKU-01", BatchID: "batch-b", BatchNo: "LOT-B", StockStatus: StockStatusAvailable, QtyOnHand: 7},
		{WarehouseID: "wh-hn", WarehouseCode: "HN", SKU: "SKU-01", BatchID: "batch-a", BatchNo: "LOT-A", StockStatus: StockStatusAvailable, QtyOnHand: 11},
	}

	snapshots := CalculateAvailableStock(rows)
	if len(snapshots) != 3 {
		t.Fatalf("snapshots length = %d, want 3", len(snapshots))
	}
}

func TestHoldStockStatusCatalog(t *testing.T) {
	holdStatuses := []StockStatus{
		StockStatusQCHold,
		StockStatusReturnPending,
		StockStatusDamaged,
		StockStatusSubcontractIssued,
	}
	for _, status := range holdStatuses {
		if !IsHoldStockStatus(status) {
			t.Fatalf("status %q should count as hold stock", status)
		}
	}

	if IsHoldStockStatus(StockStatusAvailable) {
		t.Fatal("available status should not count as hold stock")
	}
	if IsHoldStockStatus(StockStatusReserved) {
		t.Fatal("reserved status should be counted through reserved quantity, not hold")
	}
}
