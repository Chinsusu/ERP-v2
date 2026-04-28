package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
	apperrors "github.com/Chinsusu/ERP-v2/apps/api/internal/shared/errors"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/response"
)

var ErrInsufficientStock = errors.New("insufficient stock")

type PrototypeSalesOrderReservationStore struct {
	mu           sync.Mutex
	rows         []domain.StockBalanceSnapshot
	reservations []domain.StockReservation
}

func NewPrototypeSalesOrderReservationStore() *PrototypeSalesOrderReservationStore {
	return &PrototypeSalesOrderReservationStore{rows: prototypeSalesOrderReservationRows()}
}

func NewPrototypeSalesOrderReservationStoreWithRows(rows []domain.StockBalanceSnapshot) *PrototypeSalesOrderReservationStore {
	return &PrototypeSalesOrderReservationStore{rows: cloneStockBalanceRows(rows)}
}

func (s *PrototypeSalesOrderReservationStore) ReserveSalesOrder(
	_ context.Context,
	input salesapp.SalesOrderStockReservationInput,
) (salesapp.SalesOrderStockReservationResult, error) {
	if s == nil {
		return salesapp.SalesOrderStockReservationResult{}, errors.New("sales order reservation store is required")
	}
	if strings.TrimSpace(input.ActorID) == "" {
		return salesapp.SalesOrderStockReservationResult{}, domain.ErrStockReservationActorRequired
	}
	if len(input.Lines) == 0 {
		return salesapp.SalesOrderStockReservationResult{}, domain.ErrStockReservationRequiredField
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	rows := cloneStockBalanceRows(s.rows)
	reservations := make([]domain.StockReservation, 0, len(input.Lines))
	resultLines := make([]salesapp.SalesOrderReservedLine, 0, len(input.Lines))
	for _, line := range input.Lines {
		allocated, err := reserveSalesOrderLine(rows, input, line)
		if err != nil {
			return salesapp.SalesOrderStockReservationResult{}, err
		}
		reservations = append(reservations, allocated.reservation)
		resultLines = append(resultLines, allocated.line)
	}

	s.rows = rows
	s.reservations = append(s.reservations, reservations...)

	return salesapp.SalesOrderStockReservationResult{Lines: resultLines}, nil
}

func (s *PrototypeSalesOrderReservationStore) Reservations() []domain.StockReservation {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.StockReservation(nil), s.reservations...)
}

type allocatedSalesOrderLine struct {
	reservation domain.StockReservation
	line        salesapp.SalesOrderReservedLine
}

func reserveSalesOrderLine(
	rows []domain.StockBalanceSnapshot,
	input salesapp.SalesOrderStockReservationInput,
	line salesapp.SalesOrderStockReservationLineInput,
) (allocatedSalesOrderLine, error) {
	requiredQty, err := decimal.ParseQuantity(line.BaseOrderedQty.String())
	if err != nil || requiredQty.IsNegative() || requiredQty.IsZero() {
		return allocatedSalesOrderLine{}, domain.ErrStockReservationInvalidQuantity
	}

	snapshots := domain.CalculateAvailableStockAt(rows, input.ReservedAt)
	for _, snapshot := range snapshots {
		if !salesOrderReservationSnapshotMatches(snapshot, input, line) {
			continue
		}
		if !quantityAtLeast(snapshot.AvailableQty, requiredQty) {
			continue
		}
		rowIndex := findReservableStockBalanceRow(rows, snapshot)
		if rowIndex < 0 {
			continue
		}
		updatedReservedQty, err := decimal.AddQuantity(rows[rowIndex].QtyReserved, requiredQty)
		if err != nil {
			return allocatedSalesOrderLine{}, err
		}
		rows[rowIndex].QtyReserved = updatedReservedQty

		reservation, err := domain.NewStockReservation(domain.NewStockReservationInput{
			ID:               newReservationID(input.SalesOrderID, line.SalesOrderLineID),
			OrgID:            input.OrgID,
			ReservationNo:    newReservationNo(input.OrderNo, line.LineNo),
			SalesOrderID:     input.SalesOrderID,
			SalesOrderLineID: line.SalesOrderLineID,
			ItemID:           line.ItemID,
			SKUCode:          line.SKUCode,
			BatchID:          snapshot.BatchID,
			BatchNo:          snapshot.BatchNo,
			WarehouseID:      snapshot.WarehouseID,
			WarehouseCode:    snapshot.WarehouseCode,
			BinID:            snapshot.LocationID,
			BinCode:          snapshot.LocationCode,
			StockStatus:      domain.StockStatusAvailable,
			ReservedQty:      requiredQty,
			BaseUOMCode:      line.BaseUOMCode.String(),
			ReservedAt:       reservedAt(input.ReservedAt),
			ReservedBy:       input.ActorID,
		})
		if err != nil {
			return allocatedSalesOrderLine{}, err
		}

		return allocatedSalesOrderLine{
			reservation: reservation,
			line: salesapp.SalesOrderReservedLine{
				SalesOrderLineID: line.SalesOrderLineID,
				ReservedQty:      requiredQty,
				BatchID:          snapshot.BatchID,
				BatchNo:          snapshot.BatchNo,
				BinID:            snapshot.LocationID,
				BinCode:          snapshot.LocationCode,
			},
		}, nil
	}

	return allocatedSalesOrderLine{}, insufficientStockError(input, line)
}

func salesOrderReservationSnapshotMatches(
	snapshot domain.AvailableStockSnapshot,
	input salesapp.SalesOrderStockReservationInput,
	line salesapp.SalesOrderStockReservationLineInput,
) bool {
	return strings.TrimSpace(snapshot.WarehouseID) == strings.TrimSpace(input.WarehouseID) &&
		strings.EqualFold(strings.TrimSpace(snapshot.SKU), strings.TrimSpace(line.SKUCode)) &&
		snapshot.BaseUOMCode == line.BaseUOMCode
}

func findReservableStockBalanceRow(rows []domain.StockBalanceSnapshot, snapshot domain.AvailableStockSnapshot) int {
	for index, row := range rows {
		if strings.TrimSpace(row.WarehouseID) == strings.TrimSpace(snapshot.WarehouseID) &&
			strings.TrimSpace(row.LocationID) == strings.TrimSpace(snapshot.LocationID) &&
			strings.EqualFold(strings.TrimSpace(row.SKU), strings.TrimSpace(snapshot.SKU)) &&
			strings.TrimSpace(row.BatchID) == strings.TrimSpace(snapshot.BatchID) &&
			row.BaseUOMCode == snapshot.BaseUOMCode &&
			row.StockStatus == domain.StockStatusAvailable {
			return index
		}
	}

	return -1
}

func insufficientStockError(
	input salesapp.SalesOrderStockReservationInput,
	line salesapp.SalesOrderStockReservationLineInput,
) error {
	return apperrors.Conflict(
		response.ErrorCodeInsufficientStock,
		"Insufficient stock for sales order reservation",
		ErrInsufficientStock,
		map[string]any{
			"sales_order_id":      input.SalesOrderID,
			"sales_order_line_id": line.SalesOrderLineID,
			"sku_code":            line.SKUCode,
			"required_qty":        line.BaseOrderedQty.String(),
			"base_uom_code":       line.BaseUOMCode.String(),
			"warehouse_id":        input.WarehouseID,
		},
	)
}

func quantityAtLeast(available decimal.Decimal, required decimal.Decimal) bool {
	delta, err := decimal.SubtractQuantity(available, required)
	return err == nil && !delta.IsNegative()
}

func reservedAt(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}

	return value.UTC()
}

func newReservationID(salesOrderID string, salesOrderLineID string) string {
	return fmt.Sprintf("rsv-%s-%s", strings.TrimSpace(salesOrderID), strings.TrimSpace(salesOrderLineID))
}

func newReservationNo(orderNo string, lineNo int) string {
	return fmt.Sprintf("RSV-%s-%02d", strings.ToUpper(strings.TrimSpace(orderNo)), lineNo)
}

func cloneStockBalanceRows(rows []domain.StockBalanceSnapshot) []domain.StockBalanceSnapshot {
	return append([]domain.StockBalanceSnapshot(nil), rows...)
}

func prototypeSalesOrderReservationRows() []domain.StockBalanceSnapshot {
	return []domain.StockBalanceSnapshot{
		{
			WarehouseID:   "wh-hcm-fg",
			WarehouseCode: "WH-HCM-FG",
			LocationID:    "bin-hcm-pick-a01",
			LocationCode:  "PICK-A-01",
			ItemID:        "item-serum-30ml",
			SKU:           "SERUM-30ML",
			BatchID:       "batch-serum-2604a",
			BatchNo:       "LOT-2604A",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("EA"),
			StockStatus:   domain.StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("120"),
			QtyReserved:   decimal.MustQuantity("10"),
		},
		{
			WarehouseID:   "wh-hcm-fg",
			WarehouseCode: "WH-HCM-FG",
			LocationID:    "bin-hcm-pick-a02",
			LocationCode:  "PICK-A-02",
			ItemID:        "item-cream-50g",
			SKU:           "CREAM-50G",
			BatchID:       "batch-cream-2603b",
			BatchNo:       "LOT-2603B",
			BatchQCStatus: domain.QCStatusPass,
			BatchStatus:   domain.BatchStatusActive,
			BaseUOMCode:   decimal.MustUOMCode("EA"),
			StockStatus:   domain.StockStatusAvailable,
			QtyOnHand:     decimal.MustQuantity("44"),
			QtyReserved:   decimal.MustQuantity("12"),
		},
	}
}
