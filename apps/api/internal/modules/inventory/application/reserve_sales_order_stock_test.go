package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

func TestPrototypeSalesOrderReservationStoreReservesAvailableStockAtomically(t *testing.T) {
	store := NewPrototypeSalesOrderReservationStoreWithRows([]domain.StockBalanceSnapshot{
		reservableRow("SERUM-30ML", "item-serum-30ml", "20", "2"),
		reservableRow("CREAM-50G", "item-cream-50g", "5", "0"),
	})

	result, err := store.ReserveSalesOrder(context.Background(), salesReservationInput([]salesapp.SalesOrderStockReservationLineInput{
		reservationLine("line-serum", "item-serum-30ml", "SERUM-30ML", "4"),
		reservationLine("line-cream", "item-cream-50g", "CREAM-50G", "3"),
	}))
	if err != nil {
		t.Fatalf("reserve sales order: %v", err)
	}

	if len(result.Lines) != 2 {
		t.Fatalf("reserved lines = %d, want 2", len(result.Lines))
	}
	reservations := store.Reservations()
	if len(reservations) != 2 {
		t.Fatalf("reservations = %d, want 2", len(reservations))
	}
	if reservations[0].SalesOrderLineID != "line-serum" || reservations[0].ReservedQty != "4.000000" {
		t.Fatalf("reservation = %+v, want line-serum qty 4", reservations[0])
	}
	if reservations[0].BatchID == "" || reservations[0].BinID == "" {
		t.Fatalf("reservation stock link = %+v, want batch and bin", reservations[0])
	}
}

func TestPrototypeSalesOrderReservationStoreDoesNotPartiallyReserveOnInsufficientStock(t *testing.T) {
	store := NewPrototypeSalesOrderReservationStoreWithRows([]domain.StockBalanceSnapshot{
		reservableRow("SERUM-30ML", "item-serum-30ml", "20", "2"),
		reservableRow("CREAM-50G", "item-cream-50g", "5", "0"),
	})

	_, err := store.ReserveSalesOrder(context.Background(), salesReservationInput([]salesapp.SalesOrderStockReservationLineInput{
		reservationLine("line-serum", "item-serum-30ml", "SERUM-30ML", "4"),
		reservationLine("line-cream", "item-cream-50g", "CREAM-50G", "6"),
	}))
	if !errors.Is(err, ErrInsufficientStock) {
		t.Fatalf("err = %v, want insufficient stock", err)
	}
	var appErr apperrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != response.ErrorCodeInsufficientStock {
		t.Fatalf("app err = %+v, want insufficient stock code", appErr)
	}
	if appErr.Details["available_qty"] != "5.000000" || appErr.Details["required_qty"] != "6.000000" {
		t.Fatalf("details = %+v, want available and required quantities", appErr.Details)
	}
	if len(store.Reservations()) != 0 {
		t.Fatalf("reservations = %+v, want no partial reservation", store.Reservations())
	}
}

func TestPrototypeSalesOrderReservationStoreBlocksNotSellableBatches(t *testing.T) {
	for _, qcStatus := range []domain.QCStatus{domain.QCStatusHold, domain.QCStatusFail} {
		t.Run(string(qcStatus), func(t *testing.T) {
			store := NewPrototypeSalesOrderReservationStoreWithRows([]domain.StockBalanceSnapshot{
				{
					WarehouseID:   "wh-hcm-fg",
					WarehouseCode: "WH-HCM-FG",
					LocationID:    "bin-qc",
					LocationCode:  "QC-01",
					ItemID:        "item-serum-30ml",
					SKU:           "SERUM-30ML",
					BatchID:       "batch-" + string(qcStatus),
					BatchNo:       "LOT-" + string(qcStatus),
					BatchQCStatus: qcStatus,
					BatchStatus:   domain.BatchStatusActive,
					BaseUOMCode:   decimal.MustUOMCode("EA"),
					StockStatus:   domain.StockStatusAvailable,
					QtyOnHand:     decimal.MustQuantity("10"),
				},
			})

			_, err := store.ReserveSalesOrder(context.Background(), salesReservationInput([]salesapp.SalesOrderStockReservationLineInput{
				reservationLine("line-serum", "item-serum-30ml", "SERUM-30ML", "1"),
			}))
			if !errors.Is(err, ErrBatchNotSellable) {
				t.Fatalf("err = %v, want batch not sellable", err)
			}
			var appErr apperrors.AppError
			if !errors.As(err, &appErr) || appErr.Code != response.ErrorCodeBatchNotSellable {
				t.Fatalf("app err = %+v, want batch not sellable code", appErr)
			}
			if appErr.Details["qc_status"] != string(qcStatus) {
				t.Fatalf("details = %+v, want qc status %s", appErr.Details, qcStatus)
			}
		})
	}
}

func salesReservationInput(lines []salesapp.SalesOrderStockReservationLineInput) salesapp.SalesOrderStockReservationInput {
	return salesapp.SalesOrderStockReservationInput{
		OrgID:         "org-my-pham",
		SalesOrderID:  "so-reserve",
		OrderNo:       "SO-RESERVE",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		ActorID:       "user-sales",
		RequestID:     "req-reserve",
		ReservedAt:    time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC),
		Lines:         lines,
	}
}

func reservationLine(lineID string, itemID string, sku string, qty string) salesapp.SalesOrderStockReservationLineInput {
	return salesapp.SalesOrderStockReservationLineInput{
		SalesOrderLineID: lineID,
		LineNo:           1,
		ItemID:           itemID,
		SKUCode:          sku,
		OrderedQty:       decimal.MustQuantity(qty),
		BaseOrderedQty:   decimal.MustQuantity(qty),
		BaseUOMCode:      decimal.MustUOMCode("EA"),
	}
}

func reservableRow(sku string, itemID string, onHand string, reserved string) domain.StockBalanceSnapshot {
	return domain.StockBalanceSnapshot{
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		LocationID:    "bin-pick",
		LocationCode:  "PICK-01",
		ItemID:        itemID,
		SKU:           sku,
		BatchID:       "batch-" + sku,
		BatchNo:       "LOT-" + sku,
		BatchQCStatus: domain.QCStatusPass,
		BatchStatus:   domain.BatchStatusActive,
		BaseUOMCode:   decimal.MustUOMCode("EA"),
		StockStatus:   domain.StockStatusAvailable,
		QtyOnHand:     decimal.MustQuantity(onHand),
		QtyReserved:   decimal.MustQuantity(reserved),
	}
}
