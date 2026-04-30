package application

import (
	"context"
	"errors"
	"strings"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

type StockAvailabilityStore interface {
	ListBalances(ctx context.Context, filter domain.AvailableStockFilter) ([]domain.StockBalanceSnapshot, error)
}

type ListAvailableStock struct {
	store StockAvailabilityStore
}

func NewListAvailableStock(store StockAvailabilityStore) ListAvailableStock {
	return ListAvailableStock{store: store}
}

func (uc ListAvailableStock) Execute(
	ctx context.Context,
	filter domain.AvailableStockFilter,
) ([]domain.AvailableStockSnapshot, error) {
	if uc.store == nil {
		return nil, errors.New("stock availability store is required")
	}

	rows, err := uc.store.ListBalances(ctx, filter)
	if err != nil {
		return nil, err
	}

	return domain.CalculateAvailableStock(rows), nil
}

type PrototypeStockAvailabilityStore struct {
	rows []domain.StockBalanceSnapshot
}

func NewPrototypeStockAvailabilityStore() PrototypeStockAvailabilityStore {
	return PrototypeStockAvailabilityStore{rows: prototypeStockBalanceRows()}
}

func (s PrototypeStockAvailabilityStore) ListBalances(
	_ context.Context,
	filter domain.AvailableStockFilter,
) ([]domain.StockBalanceSnapshot, error) {
	matches := make([]domain.StockBalanceSnapshot, 0, len(s.rows))
	filterSKU := strings.ToUpper(strings.TrimSpace(filter.SKU))
	for _, row := range s.rows {
		if filter.WarehouseID != "" && row.WarehouseID != filter.WarehouseID {
			continue
		}
		if filter.LocationID != "" && row.LocationID != filter.LocationID {
			continue
		}
		if filter.ItemID != "" && row.ItemID != filter.ItemID {
			continue
		}
		if filterSKU != "" && strings.ToUpper(row.SKU) != filterSKU {
			continue
		}
		if filter.BatchID != "" && row.BatchID != filter.BatchID {
			continue
		}

		matches = append(matches, row)
	}

	return matches, nil
}

func prototypeStockBalanceRows() []domain.StockBalanceSnapshot {
	return []domain.StockBalanceSnapshot{
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-hcm-a01",
			LocationCode:  "A-01",
			ItemID:        "item-serum-30ml",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-serum-2604a",
			BatchNo:       "LOT-2604A",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   domain.StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("120"),
			QtyReserved:   decimal.MustQuantity("10"),
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-hcm-a01",
			LocationCode:  "A-01",
			ItemID:        "item-serum-30ml",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-serum-2604a",
			BatchNo:       "LOT-2604A",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   domain.StockStatusQCHold,
			QtyOnHand:     decimal.MustQuantity("8"),
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-hcm-a01",
			LocationCode:  "A-01",
			ItemID:        "item-cream-50g",
			SKU:           "CREAM-50G",
			BatchID:       "batch-cream-2603b",
			BatchNo:       "LOT-2603B",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   domain.StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("44"),
			QtyReserved:   decimal.MustQuantity("12"),
		},
		{
			WarehouseID:   "wh-hcm",
			WarehouseCode: "HCM",
			LocationID:    "bin-hcm-a01",
			LocationCode:  "A-01",
			ItemID:        "item-cream-50g",
			SKU:           "CREAM-50G",
			BatchID:       "batch-cream-2603b",
			BatchNo:       "LOT-2603B",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   domain.StockStatusDamaged,
			QtyOnHand:     decimal.MustQuantity("2"),
		},
		{
			WarehouseID:   "wh-hn",
			WarehouseCode: "HN",
			LocationID:    "bin-hn-r01",
			LocationCode:  "R-01",
			ItemID:        "item-toner-100ml",
			SKU:           "TONER-100ML",
			BatchID:       "batch-toner-2604c",
			BatchNo:       "LOT-2604C",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   domain.StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("85"),
			QtyReserved:   decimal.MustQuantity("20"),
		},
		{
			WarehouseID:   "wh-hn",
			WarehouseCode: "HN",
			LocationID:    "bin-hn-r01",
			LocationCode:  "R-01",
			ItemID:        "item-toner-100ml",
			SKU:           "TONER-100ML",
			BatchID:       "batch-toner-2604c",
			BatchNo:       "LOT-2604C",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("PCS"),
			StockStatus:   domain.StockStatusReturnPending,
			QtyOnHand:     decimal.MustQuantity("5"),
		},
	}
}
