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

func TestPickTaskActionsStartConfirmAndCompleteIdempotently(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	task := generatePickTaskForActionTest(t, store, auditStore)
	lineID := task.Lines[0].ID

	start := NewStartPickTask(store, auditStore)
	confirm := NewConfirmPickTaskLine(store, auditStore)
	complete := NewCompletePickTask(store, auditStore)

	started, err := start.Execute(ctx, PickTaskActionInput{
		PickTaskID: task.ID,
		ActorID:    "user-picker",
		RequestID:  "req-start-pick",
	})
	if err != nil {
		t.Fatalf("start pick task: %v", err)
	}
	if started.PickTask.Status != domain.PickTaskStatusInProgress ||
		started.PickTask.AssignedTo != "user-picker" ||
		started.AuditLogID == "" {
		t.Fatalf("started = %+v, want in-progress auto-assigned task with audit", started)
	}
	startedAgain, err := start.Execute(ctx, PickTaskActionInput{
		PickTaskID: task.ID,
		ActorID:    "user-picker",
		RequestID:  "req-start-pick",
	})
	if err != nil {
		t.Fatalf("repeat start pick task: %v", err)
	}
	if startedAgain.PickTask.Status != domain.PickTaskStatusInProgress || startedAgain.AuditLogID != "" {
		t.Fatalf("repeat start = %+v, want idempotent no-op without audit", startedAgain)
	}

	confirmed, err := confirm.Execute(ctx, ConfirmPickTaskLineInput{
		PickTaskID: task.ID,
		LineID:     lineID,
		PickedQty:  "3.000000",
		ActorID:    "user-picker",
		RequestID:  "req-confirm-pick-line",
	})
	if err != nil {
		t.Fatalf("confirm pick line: %v", err)
	}
	if confirmed.PickTask.Lines[0].Status != domain.PickTaskLineStatusPicked ||
		confirmed.PickTask.Lines[0].QtyPicked.String() != "3.000000" ||
		confirmed.AuditLogID == "" {
		t.Fatalf("confirmed = %+v, want picked line with audit", confirmed)
	}
	confirmedAgain, err := confirm.Execute(ctx, ConfirmPickTaskLineInput{
		PickTaskID: task.ID,
		LineID:     lineID,
		PickedQty:  "3",
		ActorID:    "user-picker",
		RequestID:  "req-confirm-pick-line",
	})
	if err != nil {
		t.Fatalf("repeat confirm pick line: %v", err)
	}
	if confirmedAgain.AuditLogID != "" || confirmedAgain.PickTask.Lines[0].Status != domain.PickTaskLineStatusPicked {
		t.Fatalf("repeat confirm = %+v, want idempotent no-op without audit", confirmedAgain)
	}

	completed, err := complete.Execute(ctx, PickTaskActionInput{
		PickTaskID: task.ID,
		ActorID:    "user-picker",
		RequestID:  "req-complete-pick",
	})
	if err != nil {
		t.Fatalf("complete pick task: %v", err)
	}
	if completed.PickTask.Status != domain.PickTaskStatusCompleted || completed.AuditLogID == "" {
		t.Fatalf("completed = %+v, want completed task with audit", completed)
	}
	completedAgain, err := complete.Execute(ctx, PickTaskActionInput{
		PickTaskID: task.ID,
		ActorID:    "user-picker",
		RequestID:  "req-complete-pick",
	})
	if err != nil {
		t.Fatalf("repeat complete pick task: %v", err)
	}
	if completedAgain.PickTask.Status != domain.PickTaskStatusCompleted || completedAgain.AuditLogID != "" {
		t.Fatalf("repeat complete = %+v, want idempotent no-op without audit", completedAgain)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: task.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 4 {
		t.Fatalf("audit logs = %d, want create/start/line/complete", len(logs))
	}
}

func TestReportPickTaskExceptionIsIdempotent(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	task := generatePickTaskForActionTest(t, store, auditStore)
	service := NewReportPickTaskException(store, auditStore)

	result, err := service.Execute(ctx, ReportPickTaskExceptionInput{
		PickTaskID:    task.ID,
		ExceptionCode: "wrong_batch",
		ActorID:       "user-picker",
		RequestID:     "req-pick-exception",
		Investigation: "Scanned batch does not match reserved batch",
	})
	if err != nil {
		t.Fatalf("report pick exception: %v", err)
	}
	if result.PickTask.Status != domain.PickTaskStatusWrongBatch || result.AuditLogID == "" {
		t.Fatalf("result = %+v, want wrong batch status with audit", result)
	}

	repeated, err := service.Execute(ctx, ReportPickTaskExceptionInput{
		PickTaskID:    task.ID,
		ExceptionCode: "wrong_batch",
		ActorID:       "user-picker",
		RequestID:     "req-pick-exception",
	})
	if err != nil {
		t.Fatalf("repeat report pick exception: %v", err)
	}
	if repeated.PickTask.Status != domain.PickTaskStatusWrongBatch || repeated.AuditLogID != "" {
		t.Fatalf("repeated = %+v, want idempotent no-op without audit", repeated)
	}

	_, err = service.Execute(ctx, ReportPickTaskExceptionInput{
		PickTaskID:    task.ID,
		ExceptionCode: "wrong_location",
		ActorID:       "user-picker",
		RequestID:     "req-change-pick-exception",
	})
	if !errors.Is(err, domain.ErrPickTaskInvalidTransition) {
		t.Fatalf("change exception err = %v, want invalid transition", err)
	}
}

func TestReportPickTaskLineExceptionBlocksConfirm(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	task := generatePickTaskForActionTest(t, store, auditStore)
	lineID := task.Lines[0].ID
	exceptions := NewReportPickTaskException(store, auditStore)

	result, err := exceptions.Execute(ctx, ReportPickTaskExceptionInput{
		PickTaskID:    task.ID,
		LineID:        lineID,
		ExceptionCode: "wrong_location",
		ActorID:       "user-picker",
		RequestID:     "req-pick-line-exception",
		Investigation: "Picker scanned a different bin",
	})
	if err != nil {
		t.Fatalf("report pick line exception: %v", err)
	}
	if result.PickTask.Status != domain.PickTaskStatusWrongLocation ||
		result.PickTask.Lines[0].Status != domain.PickTaskLineStatusWrongLocation ||
		result.PickTask.Lines[0].QtyPicked.String() != "0.000000" ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want wrong-location task and line with audit", result)
	}

	repeated, err := exceptions.Execute(ctx, ReportPickTaskExceptionInput{
		PickTaskID:    task.ID,
		LineID:        lineID,
		ExceptionCode: "wrong_location",
		ActorID:       "user-picker",
		RequestID:     "req-pick-line-exception",
	})
	if err != nil {
		t.Fatalf("repeat report pick line exception: %v", err)
	}
	if repeated.PickTask.Lines[0].Status != domain.PickTaskLineStatusWrongLocation || repeated.AuditLogID != "" {
		t.Fatalf("repeated = %+v, want idempotent line exception without audit", repeated)
	}

	logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.pick_task.exception_reported"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].Metadata["line_id"] != lineID {
		t.Fatalf("audit logs = %+v, want one exception log with line metadata", logs)
	}

	confirm := NewConfirmPickTaskLine(store, auditStore)
	_, err = confirm.Execute(ctx, ConfirmPickTaskLineInput{
		PickTaskID: task.ID,
		LineID:     lineID,
		PickedQty:  "3.000000",
		ActorID:    "user-picker",
		RequestID:  "req-confirm-exception-line",
	})
	if !errors.Is(err, domain.ErrPickTaskInvalidTransition) {
		t.Fatalf("confirm exception line err = %v, want invalid transition", err)
	}
}

func TestPickTaskActionsRejectInvalidTransitionsAndExceptionCodes(t *testing.T) {
	ctx := context.Background()
	store := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	task := generatePickTaskForActionTest(t, store, auditStore)
	confirm := NewConfirmPickTaskLine(store, auditStore)

	_, err := confirm.Execute(ctx, ConfirmPickTaskLineInput{
		PickTaskID: task.ID,
		LineID:     task.Lines[0].ID,
		PickedQty:  "3.000000",
		ActorID:    "user-picker",
		RequestID:  "req-confirm-before-start",
	})
	if !errors.Is(err, domain.ErrPickTaskInvalidTransition) {
		t.Fatalf("confirm before start err = %v, want invalid transition", err)
	}

	exceptions := NewReportPickTaskException(store, auditStore)
	_, err = exceptions.Execute(ctx, ReportPickTaskExceptionInput{
		PickTaskID:    task.ID,
		ExceptionCode: "not_a_pick_exception",
		ActorID:       "user-picker",
		RequestID:     "req-invalid-exception",
	})
	if !errors.Is(err, domain.ErrPickTaskInvalidStatus) {
		t.Fatalf("invalid exception err = %v, want invalid status", err)
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

func generatePickTaskForActionTest(
	t *testing.T,
	store *PrototypePickTaskStore,
	auditStore audit.LogStore,
) domain.PickTask {
	t.Helper()
	order := reservedSalesOrderForPickTask(t)
	result, err := NewGeneratePickTaskFromReservedOrder(store, auditStore).Execute(
		context.Background(),
		GeneratePickTaskFromReservedOrderInput{
			SalesOrder:   order,
			Reservations: []inventorydomain.StockReservation{stockReservationForPickTask(t, order, order.Lines[0])},
			ActorID:      "user-warehouse-lead",
			RequestID:    "req-generate-pick-for-action",
		},
	)
	if err != nil {
		t.Fatalf("generate pick task: %v", err)
	}

	return result.PickTask
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
