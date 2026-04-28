package application

import (
	"context"
	"errors"
	"testing"
	"time"

	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestGeneratePickTaskFromReservedOrderCreatesLinesFromReservations(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewGeneratePickTaskFromReservedOrder(store, auditStore)
	order := reservedSalesOrderForPickTask(t)

	result, err := service.Execute(ctx, GeneratePickTaskFromReservedOrderInput{
		SalesOrder:   order,
		Reservations: []inventorydomain.StockReservation{stockReservationForPickTask(t, order, order.Lines[0])},
		AssignedTo:   "user-picker",
		ActorID:      "user-warehouse-lead",
		RequestID:    "req-generate-pick",
	})
	if err != nil {
		t.Fatalf("generate pick task: %v", err)
	}

	task := result.PickTask
	if task.Status != domain.PickTaskStatusAssigned || task.AssignedTo != "user-picker" {
		t.Fatalf("task status/assignee = %s/%s, want assigned picker", task.Status, task.AssignedTo)
	}
	if task.SalesOrderID != order.ID || task.PickTaskNo != "PICK-SO-260428-0001" {
		t.Fatalf("task source = %+v, want source sales order", task)
	}
	if len(task.Lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(task.Lines))
	}
	line := task.Lines[0]
	if line.SKUCode != "SERUM-30ML" ||
		line.BatchID != "batch-serum-2604a" ||
		line.BinID != "bin-hcm-pick-a01" ||
		line.QtyToPick != "3.000000" ||
		line.BaseUOMCode != "EA" {
		t.Fatalf("line = %+v, want reservation SKU/batch/bin/base qty", line)
	}

	stored, err := store.GetPickTaskBySalesOrder(ctx, order.ID)
	if err != nil {
		t.Fatalf("get stored pick task: %v", err)
	}
	if stored.ID != task.ID {
		t.Fatalf("stored task id = %s, want %s", stored.ID, task.ID)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.pick_task.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["line_count"] != 1 {
		t.Fatalf("audit logs = %+v, want one pick task created log", logs)
	}
}

func TestGeneratePickTaskFromReservedOrderRejectsNonReservedOrder(t *testing.T) {
	service := NewGeneratePickTaskFromReservedOrder(NewPrototypePickTaskStore(), audit.NewInMemoryLogStore())
	order := draftSalesOrderForPickTask(t)

	_, err := service.Execute(context.Background(), GeneratePickTaskFromReservedOrderInput{
		SalesOrder: order,
		ActorID:    "user-warehouse-lead",
	})
	if !errors.Is(err, ErrPickTaskSalesOrderNotReserved) {
		t.Fatalf("err = %v, want sales order not reserved", err)
	}
}

func TestGeneratePickTaskFromReservedOrderRequiresActiveReservationPerLine(t *testing.T) {
	service := NewGeneratePickTaskFromReservedOrder(NewPrototypePickTaskStore(), audit.NewInMemoryLogStore())
	order := reservedSalesOrderForPickTask(t)

	_, err := service.Execute(context.Background(), GeneratePickTaskFromReservedOrderInput{
		SalesOrder:   order,
		Reservations: nil,
		ActorID:      "user-warehouse-lead",
	})
	if !errors.Is(err, ErrPickTaskReservationMissing) {
		t.Fatalf("err = %v, want missing reservation", err)
	}
}

func TestGeneratePickTaskFromReservedOrderRejectsDuplicateTaskForOrder(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypePickTaskStore()
	service := NewGeneratePickTaskFromReservedOrder(store, audit.NewInMemoryLogStore())
	order := reservedSalesOrderForPickTask(t)
	reservations := []inventorydomain.StockReservation{stockReservationForPickTask(t, order, order.Lines[0])}

	if _, err := service.Execute(ctx, GeneratePickTaskFromReservedOrderInput{
		SalesOrder:   order,
		Reservations: reservations,
		ActorID:      "user-warehouse-lead",
	}); err != nil {
		t.Fatalf("first generate pick task: %v", err)
	}
	_, err := service.Execute(ctx, GeneratePickTaskFromReservedOrderInput{
		SalesOrder:   order,
		Reservations: reservations,
		ActorID:      "user-warehouse-lead",
	})
	if !errors.Is(err, ErrPickTaskDuplicate) {
		t.Fatalf("err = %v, want duplicate pick task", err)
	}
}

func draftSalesOrderForPickTask(t *testing.T) salesdomain.SalesOrder {
	t.Helper()
	order, err := salesdomain.NewSalesOrderDocument(salesdomain.NewSalesOrderDocumentInput{
		ID:            "so-260428-0001",
		OrgID:         "org-my-pham",
		OrderNo:       "SO-260428-0001",
		CustomerID:    "cus-dl-minh-anh",
		CustomerCode:  "CUS-DL-MINHANH",
		CustomerName:  "Minh Anh Beauty",
		Channel:       "B2B",
		WarehouseID:   "wh-hcm-fg",
		WarehouseCode: "WH-HCM-FG",
		OrderDate:     "2026-04-28",
		CurrencyCode:  "VND",
		CreatedAt:     time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC),
		CreatedBy:     "user-sales",
		Lines: []salesdomain.NewSalesOrderLineInput{
			{
				ID:               "so-260428-0001-line-01",
				LineNo:           1,
				ItemID:           "item-serum-30ml",
				SKUCode:          "SERUM-30ML",
				ItemName:         "Vitamin C Serum",
				OrderedQty:       decimal.MustQuantity("3"),
				UOMCode:          "EA",
				BaseOrderedQty:   decimal.MustQuantity("3"),
				BaseUOMCode:      "EA",
				ConversionFactor: decimal.MustQuantity("1"),
				UnitPrice:        decimal.MustUnitPrice("125000"),
				CurrencyCode:     "VND",
				ReservedQty:      decimal.MustQuantity("3"),
			},
		},
	})
	if err != nil {
		t.Fatalf("new sales order: %v", err)
	}

	return order
}

func reservedSalesOrderForPickTask(t *testing.T) salesdomain.SalesOrder {
	t.Helper()
	order := draftSalesOrderForPickTask(t)
	confirmed, err := order.Confirm("user-sales", time.Date(2026, 4, 28, 9, 5, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("confirm order: %v", err)
	}
	reserved, err := confirmed.MarkReserved("user-sales", time.Date(2026, 4, 28, 9, 10, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("reserve order: %v", err)
	}

	return reserved
}

func stockReservationForPickTask(
	t *testing.T,
	order salesdomain.SalesOrder,
	line salesdomain.SalesOrderLine,
) inventorydomain.StockReservation {
	t.Helper()
	reservation, err := inventorydomain.NewStockReservation(inventorydomain.NewStockReservationInput{
		ID:               "rsv-so-260428-0001-line-01",
		OrgID:            order.OrgID,
		ReservationNo:    "RSV-SO-260428-0001-01",
		SalesOrderID:     order.ID,
		SalesOrderLineID: line.ID,
		ItemID:           line.ItemID,
		SKUCode:          line.SKUCode,
		BatchID:          "batch-serum-2604a",
		BatchNo:          "LOT-2604A",
		WarehouseID:      order.WarehouseID,
		WarehouseCode:    order.WarehouseCode,
		BinID:            "bin-hcm-pick-a01",
		BinCode:          "PICK-A-01",
		ReservedQty:      decimal.MustQuantity("3"),
		BaseUOMCode:      line.BaseUOMCode.String(),
		ReservedAt:       time.Date(2026, 4, 28, 9, 10, 0, 0, time.UTC),
		ReservedBy:       "user-sales",
	})
	if err != nil {
		t.Fatalf("new stock reservation: %v", err)
	}

	return reservation
}
