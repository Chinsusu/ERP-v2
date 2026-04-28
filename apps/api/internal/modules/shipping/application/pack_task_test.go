package application

import (
	"context"
	"errors"
	"testing"
	"time"

	salesdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestGeneratePackTaskAfterPickCreatesLinesAndMovesOrderToPacking(t *testing.T) {
	ctx := context.Background()
	pickStore := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	pickTask := completedPickTaskForPackTask(t, pickStore, auditStore)
	order := pickedSalesOrderForPackTask(t)
	packStore := NewPrototypePackTaskStore()
	service := NewGeneratePackTaskAfterPick(packStore, auditStore)

	result, err := service.Execute(ctx, GeneratePackTaskAfterPickInput{
		SalesOrder: order,
		PickTask:   pickTask,
		AssignedTo: "user-packer",
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-generate-pack",
	})
	if err != nil {
		t.Fatalf("generate pack task: %v", err)
	}

	task := result.PackTask
	if task.Status != domain.PackTaskStatusCreated ||
		task.AssignedTo != "user-packer" ||
		task.SalesOrderID != order.ID ||
		task.PickTaskID != pickTask.ID ||
		task.PackTaskNo != "PACK-SO-260428-0001" {
		t.Fatalf("pack task = %+v, want created task linked to order and pick task", task)
	}
	if result.SalesOrder.Status != salesdomain.SalesOrderStatusPacking ||
		result.SalesOrder.PackingStartedBy != "user-warehouse-lead" {
		t.Fatalf("sales order = %+v, want packing transition", result.SalesOrder)
	}
	if len(task.Lines) != 1 {
		t.Fatalf("lines = %d, want 1", len(task.Lines))
	}
	line := task.Lines[0]
	if line.PickTaskLineID != pickTask.Lines[0].ID ||
		line.SKUCode != "SERUM-30ML" ||
		line.BatchID != "batch-serum-2604a" ||
		line.QtyToPack != "3.000000" ||
		line.BaseUOMCode != "EA" {
		t.Fatalf("line = %+v, want picked line copied for packing", line)
	}

	stored, err := packStore.GetPackTaskByPickTask(ctx, pickTask.ID)
	if err != nil {
		t.Fatalf("get stored pack task by pick task: %v", err)
	}
	if stored.ID != task.ID {
		t.Fatalf("stored task id = %s, want %s", stored.ID, task.ID)
	}
	logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.pack_task.created"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["sales_order_state"] != string(salesdomain.SalesOrderStatusPacking) {
		t.Fatalf("audit logs = %+v, want one pack task created log with packing order state", logs)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}
}

func TestGeneratePackTaskAfterPickRejectsNonPickedOrder(t *testing.T) {
	pickStore := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	pickTask := completedPickTaskForPackTask(t, pickStore, auditStore)
	service := NewGeneratePackTaskAfterPick(NewPrototypePackTaskStore(), auditStore)

	_, err := service.Execute(context.Background(), GeneratePackTaskAfterPickInput{
		SalesOrder: reservedSalesOrderForPickTask(t),
		PickTask:   pickTask,
		ActorID:    "user-warehouse-lead",
	})
	if !errors.Is(err, ErrPackTaskSalesOrderNotPicked) {
		t.Fatalf("err = %v, want sales order not picked", err)
	}
}

func TestGeneratePackTaskAfterPickRejectsIncompletePickTask(t *testing.T) {
	pickStore := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	pickTask := generatePickTaskForActionTest(t, pickStore, auditStore)
	service := NewGeneratePackTaskAfterPick(NewPrototypePackTaskStore(), auditStore)

	_, err := service.Execute(context.Background(), GeneratePackTaskAfterPickInput{
		SalesOrder: pickedSalesOrderForPackTask(t),
		PickTask:   pickTask,
		ActorID:    "user-warehouse-lead",
	})
	if !errors.Is(err, ErrPackTaskPickTaskNotCompleted) {
		t.Fatalf("err = %v, want pick task not completed", err)
	}
}

func TestGeneratePackTaskAfterPickRejectsDuplicateTask(t *testing.T) {
	ctx := context.Background()
	pickStore := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	pickTask := completedPickTaskForPackTask(t, pickStore, auditStore)
	order := pickedSalesOrderForPackTask(t)
	service := NewGeneratePackTaskAfterPick(NewPrototypePackTaskStore(), auditStore)

	if _, err := service.Execute(ctx, GeneratePackTaskAfterPickInput{
		SalesOrder: order,
		PickTask:   pickTask,
		ActorID:    "user-warehouse-lead",
	}); err != nil {
		t.Fatalf("first generate pack task: %v", err)
	}
	_, err := service.Execute(ctx, GeneratePackTaskAfterPickInput{
		SalesOrder: order,
		PickTask:   pickTask,
		ActorID:    "user-warehouse-lead",
	})
	if !errors.Is(err, ErrPackTaskDuplicate) {
		t.Fatalf("err = %v, want duplicate pack task", err)
	}
}

func pickedSalesOrderForPackTask(t *testing.T) salesdomain.SalesOrder {
	t.Helper()
	reserved := reservedSalesOrderForPickTask(t)
	picking, err := reserved.StartPicking("user-picker", time.Date(2026, 4, 28, 14, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("start picking: %v", err)
	}
	picked, err := picking.MarkPicked("user-picker", time.Date(2026, 4, 28, 14, 15, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("mark picked: %v", err)
	}

	return picked
}

func completedPickTaskForPackTask(
	t *testing.T,
	store *PrototypePickTaskStore,
	auditStore audit.LogStore,
) domain.PickTask {
	t.Helper()
	task := generatePickTaskForActionTest(t, store, auditStore)
	started, err := NewStartPickTask(store, auditStore).Execute(context.Background(), PickTaskActionInput{
		PickTaskID: task.ID,
		ActorID:    "user-picker",
		RequestID:  "req-start-pick-for-pack",
	})
	if err != nil {
		t.Fatalf("start pick task: %v", err)
	}
	confirmed, err := NewConfirmPickTaskLine(store, auditStore).Execute(context.Background(), ConfirmPickTaskLineInput{
		PickTaskID: started.PickTask.ID,
		LineID:     started.PickTask.Lines[0].ID,
		PickedQty:  started.PickTask.Lines[0].QtyToPick.String(),
		ActorID:    "user-picker",
		RequestID:  "req-confirm-pick-for-pack",
	})
	if err != nil {
		t.Fatalf("confirm pick line: %v", err)
	}
	completed, err := NewCompletePickTask(store, auditStore).Execute(context.Background(), PickTaskActionInput{
		PickTaskID: confirmed.PickTask.ID,
		ActorID:    "user-picker",
		RequestID:  "req-complete-pick-for-pack",
	})
	if err != nil {
		t.Fatalf("complete pick task: %v", err)
	}

	return completed.PickTask
}
