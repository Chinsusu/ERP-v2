package domain

import (
	"sort"
	"strings"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type StockBalanceSnapshot struct {
	WarehouseID   string
	WarehouseCode string
	LocationID    string
	LocationCode  string
	ItemID        string
	SKU           string
	BatchID       string
	BatchNo       string
	BatchQCStatus QCStatus
	BatchStatus   BatchStatus
	BatchExpiry   time.Time
	BaseUOMCode   decimal.UOMCode
	StockStatus   StockStatus
	QtyOnHand     decimal.Decimal
	QtyReserved   decimal.Decimal
}

type AvailableStockFilter struct {
	WarehouseID string
	LocationID  string
	SKU         string
	BatchID     string
}

type AvailableStockSnapshot struct {
	WarehouseID      string
	WarehouseCode    string
	LocationID       string
	LocationCode     string
	SKU              string
	BatchID          string
	BatchNo          string
	BatchQCStatus    QCStatus
	BatchStatus      BatchStatus
	BatchExpiry      time.Time
	BaseUOMCode      decimal.UOMCode
	PhysicalQty      decimal.Decimal
	ReservedQty      decimal.Decimal
	QCHoldQty        decimal.Decimal
	DamagedQty       decimal.Decimal
	ReturnPendingQty decimal.Decimal
	BlockedQty       decimal.Decimal
	HoldQty          decimal.Decimal
	AvailableQty     decimal.Decimal
}

type availabilityKey struct {
	warehouseID string
	locationID  string
	sku         string
	batchID     string
	baseUOMCode decimal.UOMCode
}

func NewAvailableStockFilter(warehouseID string, locationID string, sku string, batchID string) AvailableStockFilter {
	return AvailableStockFilter{
		WarehouseID: strings.TrimSpace(warehouseID),
		LocationID:  strings.TrimSpace(locationID),
		SKU:         strings.ToUpper(strings.TrimSpace(sku)),
		BatchID:     strings.TrimSpace(batchID),
	}
}

func CalculateAvailableStock(rows []StockBalanceSnapshot) []AvailableStockSnapshot {
	return CalculateAvailableStockAt(rows, time.Now().UTC())
}

func CalculateAvailableStockAt(rows []StockBalanceSnapshot, asOf time.Time) []AvailableStockSnapshot {
	grouped := make(map[availabilityKey]*AvailableStockSnapshot)

	for _, row := range rows {
		key := availabilityKey{
			warehouseID: strings.TrimSpace(row.WarehouseID),
			locationID:  strings.TrimSpace(row.LocationID),
			sku:         strings.ToUpper(strings.TrimSpace(row.SKU)),
			batchID:     strings.TrimSpace(row.BatchID),
			baseUOMCode: row.BaseUOMCode,
		}
		if key.warehouseID == "" || key.sku == "" || key.baseUOMCode == "" {
			continue
		}

		snapshot, ok := grouped[key]
		if !ok {
			snapshot = &AvailableStockSnapshot{
				WarehouseID:      key.warehouseID,
				WarehouseCode:    strings.TrimSpace(row.WarehouseCode),
				LocationID:       key.locationID,
				LocationCode:     strings.TrimSpace(row.LocationCode),
				SKU:              key.sku,
				BatchID:          key.batchID,
				BatchNo:          strings.TrimSpace(row.BatchNo),
				BatchQCStatus:    NormalizeQCStatus(row.BatchQCStatus),
				BatchStatus:      NormalizeBatchStatus(row.BatchStatus),
				BatchExpiry:      dateOnly(row.BatchExpiry),
				BaseUOMCode:      key.baseUOMCode,
				PhysicalQty:      decimal.MustQuantity("0"),
				ReservedQty:      decimal.MustQuantity("0"),
				QCHoldQty:        decimal.MustQuantity("0"),
				DamagedQty:       decimal.MustQuantity("0"),
				ReturnPendingQty: decimal.MustQuantity("0"),
				BlockedQty:       decimal.MustQuantity("0"),
				HoldQty:          decimal.MustQuantity("0"),
				AvailableQty:     decimal.MustQuantity("0"),
			}
			grouped[key] = snapshot
		}

		snapshot.PhysicalQty = mustAddQuantity(snapshot.PhysicalQty, row.QtyOnHand)
		snapshot.ReservedQty = mustAddQuantity(snapshot.ReservedQty, row.QtyReserved)
		switch row.StockStatus {
		case StockStatusQCHold:
			snapshot.QCHoldQty = mustAddQuantity(snapshot.QCHoldQty, row.QtyOnHand)
		case StockStatusDamaged:
			snapshot.DamagedQty = mustAddQuantity(snapshot.DamagedQty, row.QtyOnHand)
			snapshot.BlockedQty = mustAddQuantity(snapshot.BlockedQty, row.QtyOnHand)
		case StockStatusReturnPending:
			snapshot.ReturnPendingQty = mustAddQuantity(snapshot.ReturnPendingQty, row.QtyOnHand)
			snapshot.BlockedQty = mustAddQuantity(snapshot.BlockedQty, row.QtyOnHand)
		case StockStatusSubcontractIssued:
			snapshot.BlockedQty = mustAddQuantity(snapshot.BlockedQty, row.QtyOnHand)
		}
		if row.StockStatus == StockStatusAvailable {
			switch batchAvailabilityStockStatus(row, asOf) {
			case StockStatusQCHold:
				snapshot.QCHoldQty = mustAddQuantity(snapshot.QCHoldQty, row.QtyOnHand)
			case StockStatusDamaged:
				snapshot.BlockedQty = mustAddQuantity(snapshot.BlockedQty, row.QtyOnHand)
			}
		}
	}

	snapshots := make([]AvailableStockSnapshot, 0, len(grouped))
	for _, snapshot := range grouped {
		snapshot.HoldQty = mustAddQuantity(snapshot.QCHoldQty, snapshot.BlockedQty)
		snapshot.AvailableQty = calculateAvailableQty(*snapshot)
		snapshots = append(snapshots, *snapshot)
	}

	sort.Slice(snapshots, func(i int, j int) bool {
		left := snapshots[i]
		right := snapshots[j]
		if left.WarehouseCode != right.WarehouseCode {
			return left.WarehouseCode < right.WarehouseCode
		}
		if left.LocationCode != right.LocationCode {
			return left.LocationCode < right.LocationCode
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

func calculateAvailableQty(snapshot AvailableStockSnapshot) decimal.Decimal {
	available := mustSubtractQuantity(snapshot.PhysicalQty, snapshot.ReservedQty)
	available = mustSubtractQuantity(available, snapshot.QCHoldQty)
	available = mustSubtractQuantity(available, snapshot.BlockedQty)
	if available.IsNegative() {
		return decimal.MustQuantity("0")
	}

	return available
}

func mustAddQuantity(left decimal.Decimal, right decimal.Decimal) decimal.Decimal {
	result, err := decimal.AddQuantity(left, right)
	if err != nil {
		panic(err)
	}

	return result
}

func batchAvailabilityStockStatus(row StockBalanceSnapshot, asOf time.Time) StockStatus {
	qcStatus := NormalizeQCStatus(row.BatchQCStatus)
	if qcStatus == "" {
		return StockStatusAvailable
	}
	batchStatus := NormalizeBatchStatus(row.BatchStatus)
	if batchStatus == "" {
		batchStatus = BatchStatusActive
	}

	if batchStatus != BatchStatusActive || (!row.BatchExpiry.IsZero() && dateOnly(row.BatchExpiry).Before(dateOnly(asOf))) {
		return StockStatusDamaged
	}
	switch qcStatus {
	case QCStatusPass:
		return StockStatusAvailable
	case QCStatusHold, QCStatusQuarantine, QCStatusRetestRequired:
		return StockStatusQCHold
	case QCStatusFail:
		return StockStatusDamaged
	default:
		return StockStatusDamaged
	}
}

func mustSubtractQuantity(left decimal.Decimal, right decimal.Decimal) decimal.Decimal {
	result, err := decimal.SubtractQuantity(left, right)
	if err != nil {
		panic(err)
	}

	return result
}
