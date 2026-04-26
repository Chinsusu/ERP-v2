package domain

import (
	"sort"
	"strings"
)

type StockBalanceSnapshot struct {
	WarehouseID   string
	WarehouseCode string
	ItemID        string
	SKU           string
	BatchID       string
	BatchNo       string
	StockStatus   StockStatus
	QtyOnHand     int64
	QtyReserved   int64
}

type AvailableStockFilter struct {
	WarehouseID string
	SKU         string
	BatchID     string
}

type AvailableStockSnapshot struct {
	WarehouseID    string
	WarehouseCode  string
	SKU            string
	BatchID        string
	BatchNo        string
	PhysicalStock  int64
	ReservedStock  int64
	HoldStock      int64
	AvailableStock int64
}

type availabilityKey struct {
	warehouseID string
	sku         string
	batchID     string
}

func NewAvailableStockFilter(warehouseID string, sku string, batchID string) AvailableStockFilter {
	return AvailableStockFilter{
		WarehouseID: strings.TrimSpace(warehouseID),
		SKU:         strings.ToUpper(strings.TrimSpace(sku)),
		BatchID:     strings.TrimSpace(batchID),
	}
}

func CalculateAvailableStock(rows []StockBalanceSnapshot) []AvailableStockSnapshot {
	grouped := make(map[availabilityKey]*AvailableStockSnapshot)

	for _, row := range rows {
		key := availabilityKey{
			warehouseID: strings.TrimSpace(row.WarehouseID),
			sku:         strings.ToUpper(strings.TrimSpace(row.SKU)),
			batchID:     strings.TrimSpace(row.BatchID),
		}
		if key.warehouseID == "" || key.sku == "" {
			continue
		}

		snapshot, ok := grouped[key]
		if !ok {
			snapshot = &AvailableStockSnapshot{
				WarehouseID:   key.warehouseID,
				WarehouseCode: strings.TrimSpace(row.WarehouseCode),
				SKU:           key.sku,
				BatchID:       key.batchID,
				BatchNo:       strings.TrimSpace(row.BatchNo),
			}
			grouped[key] = snapshot
		}

		snapshot.PhysicalStock += row.QtyOnHand
		snapshot.ReservedStock += row.QtyReserved
		if IsHoldStockStatus(row.StockStatus) {
			snapshot.HoldStock += row.QtyOnHand
		}
	}

	snapshots := make([]AvailableStockSnapshot, 0, len(grouped))
	for _, snapshot := range grouped {
		snapshot.AvailableStock = snapshot.PhysicalStock - snapshot.ReservedStock - snapshot.HoldStock
		snapshots = append(snapshots, *snapshot)
	}

	sort.Slice(snapshots, func(i int, j int) bool {
		left := snapshots[i]
		right := snapshots[j]
		if left.WarehouseCode != right.WarehouseCode {
			return left.WarehouseCode < right.WarehouseCode
		}
		if left.SKU != right.SKU {
			return left.SKU < right.SKU
		}

		return left.BatchNo < right.BatchNo
	})

	return snapshots
}

func IsHoldStockStatus(status StockStatus) bool {
	switch status {
	case StockStatusQCHold, StockStatusReturnPending, StockStatusDamaged, StockStatusSubcontractIssued:
		return true
	default:
		return false
	}
}
