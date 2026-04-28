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

func TestPackTaskActionsStartConfirmAndException(t *testing.T) {
	ctx := context.Background()
	pickStore := NewPrototypePickTaskStore()
	auditStore := audit.NewInMemoryLogStore()
	pickTask := completedPickTaskForPackTask(t, pickStore, auditStore)
	order := pickedSalesOrderForPackTask(t)
	packStore := NewPrototypePackTaskStore()
	generated, err := NewGeneratePackTaskAfterPick(packStore, auditStore).Execute(ctx, GeneratePackTaskAfterPickInput{
		SalesOrder: order,
		PickTask:   pickTask,
		ActorID:    "user-warehouse-lead",
		RequestID:  "req-generate-pack-for-actions",
	})
	if err != nil {
		t.Fatalf("generate pack task: %v", err)
	}
	packer := &fakePackTaskSalesOrderPacker{order: generated.SalesOrder}

	started, err := NewStartPackTask(packStore, auditStore).Execute(ctx, PackTaskActionInput{
		PackTaskID: generated.PackTask.ID,
		ActorID:    "user-packer",
		RequestID:  "req-start-pack",
	})
	if err != nil {
		t.Fatalf("start pack task: %v", err)
	}
	if started.PackTask.Status != domain.PackTaskStatusInProgress || started.AuditLogID == "" {
		t.Fatalf("started = %+v, want in-progress task with audit", started)
	}
	startedAgain, err := NewStartPackTask(packStore, auditStore).Execute(ctx, PackTaskActionInput{
		PackTaskID: generated.PackTask.ID,
		ActorID:    "user-packer",
		RequestID:  "req-start-pack",
	})
	if err != nil {
		t.Fatalf("repeat start pack task: %v", err)
	}
	if startedAgain.PackTask.Status != domain.PackTaskStatusInProgress || startedAgain.AuditLogID != "" {
		t.Fatalf("repeat start = %+v, want idempotent no-op without audit", startedAgain)
	}

	confirmed, err := NewConfirmPackTask(packStore, auditStore, packer).Execute(ctx, ConfirmPackTaskInput{
		PackTaskID: generated.PackTask.ID,
		Lines: []ConfirmPackTaskLineInput{
			{LineID: generated.PackTask.Lines[0].ID, PackedQty: "3"},
		},
		ActorID:   "user-packer",
		RequestID: "req-confirm-pack",
	})
	if err != nil {
		t.Fatalf("confirm pack task: %v", err)
	}
	if confirmed.PackTask.Status != domain.PackTaskStatusPacked ||
		confirmed.PackTask.Lines[0].Status != domain.PackTaskLineStatusPacked ||
		confirmed.PackTask.Lines[0].QtyPacked.String() != "3.000000" ||
		confirmed.SalesOrder.Status != salesdomain.SalesOrderStatusPacked ||
		confirmed.AuditLogID == "" {
		t.Fatalf("confirmed = %+v, want packed task, packed line, packed order, and audit", confirmed)
	}
	confirmedAgain, err := NewConfirmPackTask(packStore, auditStore, packer).Execute(ctx, ConfirmPackTaskInput{
		PackTaskID: generated.PackTask.ID,
		ActorID:    "user-packer",
		RequestID:  "req-confirm-pack",
	})
	if err != nil {
		t.Fatalf("repeat confirm pack task: %v", err)
	}
	if confirmedAgain.PackTask.Status != domain.PackTaskStatusPacked || confirmedAgain.AuditLogID != "" {
		t.Fatalf("repeat confirm = %+v, want idempotent no-op without audit", confirmedAgain)
	}
	if packer.calls != 1 {
		t.Fatalf("sales order pack calls = %d, want 1", packer.calls)
	}

	logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.pack_task.confirmed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["sales_order_state"] != string(salesdomain.SalesOrderStatusPacked) {
		t.Fatalf("audit logs = %+v, want one confirmed log with packed order state", logs)
	}
}

func TestReportPackTaskExceptionBlocksConfirm(t *testing.T) {
	cases := []struct {
		name          string
		exceptionCode string
		investigation string
	}{
		{
			name:          "missing stock",
			exceptionCode: "missing_stock",
			investigation: "Packed quantity did not match the order line",
		},
		{
			name:          "wrong SKU",
			exceptionCode: "wrong_sku",
			investigation: "Scanner reported a different SKU",
		},
		{
			name:          "wrong batch",
			exceptionCode: "wrong_batch",
			investigation: "Scanner reported a different batch",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			pickStore := NewPrototypePickTaskStore()
			auditStore := audit.NewInMemoryLogStore()
			pickTask := completedPickTaskForPackTask(t, pickStore, auditStore)
			packStore := NewPrototypePackTaskStore()
			generated, err := NewGeneratePackTaskAfterPick(packStore, auditStore).Execute(ctx, GeneratePackTaskAfterPickInput{
				SalesOrder: pickedSalesOrderForPackTask(t),
				PickTask:   pickTask,
				ActorID:    "user-warehouse-lead",
				RequestID:  "req-generate-pack-for-exception",
			})
			if err != nil {
				t.Fatalf("generate pack task: %v", err)
			}

			result, err := NewReportPackTaskException(packStore, auditStore).Execute(ctx, ReportPackTaskExceptionInput{
				PackTaskID:    generated.PackTask.ID,
				LineID:        generated.PackTask.Lines[0].ID,
				ExceptionCode: tc.exceptionCode,
				ActorID:       "user-packer",
				RequestID:     "req-pack-exception",
				Investigation: tc.investigation,
			})
			if err != nil {
				t.Fatalf("report pack exception: %v", err)
			}
			if result.PackTask.Status != domain.PackTaskStatusPackException ||
				result.PackTask.Lines[0].Status != domain.PackTaskLineStatusPackException ||
				result.AuditLogID == "" {
				t.Fatalf("result = %+v, want pack exception task and line with audit", result)
			}

			logs, err := auditStore.List(ctx, audit.Query{Action: "shipping.pack_task.exception_reported"})
			if err != nil {
				t.Fatalf("list audit logs: %v", err)
			}
			if len(logs) != 1 ||
				logs[0].Metadata["exception_code"] != tc.exceptionCode ||
				logs[0].Metadata["exception_status"] != string(domain.PackTaskStatusPackException) {
				t.Fatalf("audit logs = %+v, want one exception log with specific code", logs)
			}

			packer := &fakePackTaskSalesOrderPacker{order: pickedSalesOrderForPackTask(t)}
			_, err = NewConfirmPackTask(packStore, auditStore, packer).Execute(ctx, ConfirmPackTaskInput{
				PackTaskID: generated.PackTask.ID,
				ActorID:    "user-packer",
				RequestID:  "req-confirm-after-pack-exception",
			})
			if !errors.Is(err, domain.ErrPackTaskInvalidTransition) {
				t.Fatalf("confirm exception task err = %v, want invalid transition", err)
			}
			if packer.calls != 0 {
				t.Fatalf("sales order pack calls = %d, want 0", packer.calls)
			}
		})
	}
}

type fakePackTaskSalesOrderPacker struct {
	order salesdomain.SalesOrder
	calls int
}

func (f *fakePackTaskSalesOrderPacker) MarkSalesOrderPacked(
	_ context.Context,
	input PackTaskSalesOrderPackedInput,
) (salesdomain.SalesOrder, error) {
	if f == nil {
		return salesdomain.SalesOrder{}, ErrPackTaskSalesOrderPackerRequired
	}
	if input.SalesOrderID != f.order.ID {
		return salesdomain.SalesOrder{}, ErrPackTaskSalesOrderNotPicked
	}
	packed, err := f.order.MarkPacked(input.ActorID, time.Date(2026, 4, 28, 16, 0, 0, 0, time.UTC))
	if err != nil {
		return salesdomain.SalesOrder{}, err
	}
	f.order = packed
	f.calls++

	return packed, nil
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
