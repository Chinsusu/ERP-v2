package domain

import (
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestCalculateAvailableStockUsesReservedQCHoldAndBlockedBuckets(t *testing.T) {
	rows := []StockBalanceSnapshot{
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-a",
			LocationCode:  "A-01",
			SKU:           "serum-30ml",
			BatchID:       "batch-a",
			BatchNo:       "LOT-A",
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("120"),
			QtyReserved:   decimal.MustQuantity("15"),
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-a",
			LocationCode:  "A-01",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-a",
			BatchNo:       "LOT-A",
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   StockStatusQCHold,
			QtyOnHand:     decimal.MustQuantity("8"),
			QtyReserved:   decimal.MustQuantity("0"),
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-a",
			LocationCode:  "A-01",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-a",
			BatchNo:       "LOT-A",
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   StockStatusDamaged,
			QtyOnHand:     decimal.MustQuantity("3"),
			QtyReserved:   decimal.MustQuantity("0"),
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-a",
			LocationCode:  "A-01",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-a",
			BatchNo:       "LOT-A",
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   StockStatusReturnPending,
			QtyOnHand:     decimal.MustQuantity("2.5"),
			QtyReserved:   decimal.MustQuantity("0"),
		},
	}

	snapshots := CalculateAvailableStock(rows)
	if len(snapshots) != 1 {
		t.Fatalf("snapshots length = %d, want 1", len(snapshots))
	}

	got := snapshots[0]
	if got.PhysicalQty != "133.500000" {
		t.Fatalf("physical qty = %q, want 133.500000", got.PhysicalQty)
	}
	if got.ReservedQty != "15.000000" {
		t.Fatalf("reserved qty = %q, want 15.000000", got.ReservedQty)
	}
	if got.QCHoldQty != "8.000000" {
		t.Fatalf("qc hold qty = %q, want 8.000000", got.QCHoldQty)
	}
	if got.BlockedQty != "5.500000" {
		t.Fatalf("blocked qty = %q, want 5.500000", got.BlockedQty)
	}
	if got.AvailableQty != "105.000000" {
		t.Fatalf("available qty = %q, want 105.000000", got.AvailableQty)
	}
}

func TestCalculateAvailableStockGroupsByWarehouseLocationSKUAndBatch(t *testing.T) {
	rows := []StockBalanceSnapshot{
		{WarehouseID: "wh-hcm", WarehouseCode: "HCM", LocationID: "bin-a", LocationCode: "A-01", SKU: "SKU-01", BatchID: "batch-a", BatchNo: "LOT-A", BaseUOMCode: decimal.MustUOMCode("PCS"), StockStatus: StockStatusAvailable, QtyOnHand: decimal.MustQuantity("5")},
		{WarehouseID: "wh-hcm", WarehouseCode: "HCM", LocationID: "bin-b", LocationCode: "B-01", SKU: "SKU-01", BatchID: "batch-a", BatchNo: "LOT-A", BaseUOMCode: decimal.MustUOMCode("PCS"), StockStatus: StockStatusAvailable, QtyOnHand: decimal.MustQuantity("7")},
		{WarehouseID: "wh-hn", WarehouseCode: "HN", LocationID: "bin-a", LocationCode: "A-01", SKU: "SKU-01", BatchID: "batch-a", BatchNo: "LOT-A", BaseUOMCode: decimal.MustUOMCode("PCS"), StockStatus: StockStatusAvailable, QtyOnHand: decimal.MustQuantity("11")},
	}

	snapshots := CalculateAvailableStock(rows)
	if len(snapshots) != 3 {
		t.Fatalf("snapshots length = %d, want 3", len(snapshots))
	}
}

func TestCalculateAvailableStockPreventsNegativeAvailableQty(t *testing.T) {
	rows := []StockBalanceSnapshot{
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-a",
			LocationCode:  "A-01",
			SKU:           "SKU-NEG",
			BatchID:       "batch-neg",
			BatchNo:       "LOT-NEG",
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("5"),
			QtyReserved:   decimal.MustQuantity("8"),
		},
	}

	snapshots := CalculateAvailableStock(rows)
	if len(snapshots) != 1 {
		t.Fatalf("snapshots length = %d, want 1", len(snapshots))
	}
	if snapshots[0].AvailableQty != "0.000000" {
		t.Fatalf("available qty = %q, want 0.000000", snapshots[0].AvailableQty)
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
