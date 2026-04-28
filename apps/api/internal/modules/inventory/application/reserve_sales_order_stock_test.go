package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
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

func TestPrototypeSalesOrderReservationStoreReleasesActiveReservations(t *testing.T) {
	store := NewPrototypeSalesOrderReservationStoreWithRows([]domain.StockBalanceSnapshot{
		reservableRow("SERUM-30ML", "item-serum-30ml", "20", "2"),
	})

	_, err := store.ReserveSalesOrder(context.Background(), salesReservationInput([]salesapp.SalesOrderStockReservationLineInput{
		reservationLine("line-serum", "item-serum-30ml", "SERUM-30ML", "4"),
	}))
	if err != nil {
		t.Fatalf("reserve sales order: %v", err)
	}
	if got := store.Rows()[0].QtyReserved; got != "6.000000" {
		t.Fatalf("reserved qty after reserve = %s, want 6.000000", got)
	}

	result, err := store.ReleaseSalesOrder(context.Background(), salesapp.SalesOrderStockReleaseInput{
		OrgID:        "org-my-pham",
		SalesOrderID: "so-reserve",
		OrderNo:      "SO-RESERVE",
		ActorID:      "user-sales",
		Reason:       "customer changed order",
		ReleasedAt:   time.Date(2026, 4, 28, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("release sales order: %v", err)
	}
	if result.ReleasedReservationCount != 1 {
		t.Fatalf("released count = %d, want 1", result.ReleasedReservationCount)
	}
	if got := store.Rows()[0].QtyReserved; got != "2.000000" {
		t.Fatalf("reserved qty after release = %s, want original 2.000000", got)
	}
	reservations := store.Reservations()
	if reservations[0].Status != domain.ReservationStatusReleased || reservations[0].ReleasedBy != "user-sales" {
		t.Fatalf("reservation = %+v, want released metadata", reservations[0])
	}
}

func TestPrototypeSalesOrderReservationStoreRecordsReserveAndReleaseAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypeSalesOrderReservationStoreWithRows([]domain.StockBalanceSnapshot{
		reservableRow("SERUM-30ML", "item-serum-30ml", "20", "2"),
	}, auditStore)

	_, err := store.ReserveSalesOrder(ctx, salesReservationInput([]salesapp.SalesOrderStockReservationLineInput{
		reservationLine("line-serum", "item-serum-30ml", "SERUM-30ML", "4"),
	}))
	if err != nil {
		t.Fatalf("reserve sales order: %v", err)
	}
	reservation := store.Reservations()[0]
	logs, err := auditStore.List(ctx, audit.Query{EntityID: reservation.ID})
	if err != nil {
		t.Fatalf("list reserve audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Action != stockReservationReservedAction || logs[0].ActorID != "user-sales" {
		t.Fatalf("reserve audit logs = %+v, want one actor-stamped reserve log", logs)
	}
	if logs[0].AfterData["status"] != "active" ||
		logs[0].AfterData["reserved_qty"] != "4.000000" ||
		logs[0].Metadata["reason"] != "sales order confirm" {
		t.Fatalf("reserve audit log = %+v, want after data and reason", logs[0])
	}

	_, err = store.ReleaseSalesOrder(ctx, salesapp.SalesOrderStockReleaseInput{
		OrgID:        "org-my-pham",
		SalesOrderID: "so-reserve",
		OrderNo:      "SO-RESERVE",
		ActorID:      "user-sales",
		Reason:       "customer changed order",
		RequestID:    "req-release",
		ReleasedAt:   time.Date(2026, 4, 28, 13, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("release sales order: %v", err)
	}
	logs, err = auditStore.List(ctx, audit.Query{EntityID: reservation.ID})
	if err != nil {
		t.Fatalf("list release audit logs: %v", err)
	}
	if len(logs) != 2 || logs[0].Action != stockReservationReleasedAction {
		t.Fatalf("reservation audit logs = %+v, want release log first", logs)
	}
	if logs[0].BeforeData["status"] != "active" ||
		logs[0].AfterData["status"] != "released" ||
		logs[0].Metadata["reason"] != "customer changed order" {
		t.Fatalf("release audit log = %+v, want before/after status and reason", logs[0])
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
		Reason:        "sales order confirm",
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
