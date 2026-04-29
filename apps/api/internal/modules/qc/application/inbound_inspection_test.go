package application

import (
	"context"
	"errors"
	"testing"
	"time"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	qcdomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestInboundQCInspectionServiceCreateStartPassAuditsFlow(t *testing.T) {
	service, auditStore, movementStore := newTestInboundQCInspectionService()
	ctx := context.Background()

	created, err := service.CreateInboundQCInspection(ctx, CreateInboundQCInspectionInput{
		ID:                 "iqc-api-flow",
		GoodsReceiptID:     "grn-hcm-260427-inspect",
		GoodsReceiptLineID: "grn-line-draft-001",
		ActorID:            "user-qa",
		RequestID:          "req-iqc-create",
	})
	if err != nil {
		t.Fatalf("create inspection: %v", err)
	}
	if created.Inspection.Status != qcdomain.InboundQCInspectionStatusPending ||
		created.Inspection.SKU != "CREAM-50G" ||
		created.Inspection.UOMCode.String() != "EA" {
		t.Fatalf("created inspection = %+v, want pending cream inspection in base UOM", created.Inspection)
	}

	started, err := service.StartInboundQCInspection(ctx, InboundQCActionInput{
		ID:        created.Inspection.ID,
		ActorID:   "user-qa",
		RequestID: "req-iqc-start",
	})
	if err != nil {
		t.Fatalf("start inspection: %v", err)
	}
	if started.PreviousStatus != qcdomain.InboundQCInspectionStatusPending ||
		started.CurrentStatus != qcdomain.InboundQCInspectionStatusInProgress {
		t.Fatalf("start result = %+v, want pending -> in_progress", started)
	}

	passed, err := service.PassInboundQCInspection(ctx, InboundQCActionInput{
		ID:        created.Inspection.ID,
		Checklist: completedChecklistInput("pass"),
		ActorID:   "user-qa",
		RequestID: "req-iqc-pass",
	})
	if err != nil {
		t.Fatalf("pass inspection: %v", err)
	}
	if passed.CurrentStatus != qcdomain.InboundQCInspectionStatusCompleted ||
		passed.CurrentResult != qcdomain.InboundQCResultPass ||
		passed.Inspection.PassedQuantity.String() != passed.Inspection.Quantity.String() {
		t.Fatalf("pass result = %+v, want completed full pass", passed)
	}

	logs, err := auditStore.List(ctx, audit.Query{
		Action:   "qc.inbound_inspection.passed",
		EntityID: created.Inspection.ID,
	})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("passed audit logs = %d, want 1", len(logs))
	}
	if logs[0].Action != "qc.inbound_inspection.passed" ||
		logs[0].AfterData["result"] != "pass" ||
		logs[0].AfterData["stock_movement_count"] != 1 ||
		logs[0].Metadata["goods_receipt_id"] != "grn-hcm-260427-inspect" {
		t.Fatalf("latest audit log = %+v, want passed inspection metadata", logs[0])
	}
	movements := movementStore.Movements()
	if len(movements) != 1 {
		t.Fatalf("stock movements = %d, want 1", len(movements))
	}
	movement := movements[0]
	if movement.MovementType != inventorydomain.MovementPurchaseReceipt ||
		movement.StockStatus != inventorydomain.StockStatusAvailable ||
		movement.SourceDocType != inboundQCStockMovementSourceDoc ||
		movement.SourceDocID != created.Inspection.ID ||
		movement.SourceDocLineID != created.Inspection.GoodsReceiptLineID ||
		movement.Quantity.String() != passed.Inspection.Quantity.String() {
		t.Fatalf("stock movement = %+v, want available purchase receipt from inbound QC pass", movement)
	}
	delta, err := movement.BalanceDelta()
	if err != nil {
		t.Fatalf("movement balance delta: %v", err)
	}
	if delta.OnHand.String() != passed.Inspection.Quantity.String() ||
		delta.Available.String() != passed.Inspection.Quantity.String() {
		t.Fatalf("movement delta = %+v, want on hand and available full pass quantity", delta)
	}
	movementLogs, err := auditStore.List(ctx, audit.Query{Action: inboundQCStockMovementAction})
	if err != nil {
		t.Fatalf("list movement audit logs: %v", err)
	}
	if len(movementLogs) != 1 ||
		movementLogs[0].AfterData["stock_status"] != "available" ||
		movementLogs[0].AfterData["delta_available"] != passed.Inspection.Quantity.String() {
		t.Fatalf("movement audit logs = %+v, want available stock movement audit", movementLogs)
	}
}

func TestInboundQCInspectionServiceRejectsNonInspectableReceiving(t *testing.T) {
	service, _, _ := newTestInboundQCInspectionService()

	_, err := service.CreateInboundQCInspection(context.Background(), CreateInboundQCInspectionInput{
		GoodsReceiptID:     "grn-hcm-260427-submitted",
		GoodsReceiptLineID: "grn-line-draft-001",
		ActorID:            "user-qa",
	})
	if !errors.Is(err, ErrInboundQCReceivingInvalidState) {
		t.Fatalf("err = %v, want invalid receiving state", err)
	}
}

func TestInboundQCInspectionServicePreventsDuplicateOpenReceivingLine(t *testing.T) {
	service, _, _ := newTestInboundQCInspectionService()
	input := CreateInboundQCInspectionInput{
		GoodsReceiptID:     "grn-hcm-260427-inspect",
		GoodsReceiptLineID: "grn-line-draft-001",
		ActorID:            "user-qa",
	}

	if _, err := service.CreateInboundQCInspection(context.Background(), input); err != nil {
		t.Fatalf("create first inspection: %v", err)
	}
	_, err := service.CreateInboundQCInspection(context.Background(), input)
	if !errors.Is(err, ErrInboundQCDuplicateReceivingLine) {
		t.Fatalf("err = %v, want duplicate receiving line", err)
	}
}

func TestInboundQCInspectionServicePartialRequiresValidSplit(t *testing.T) {
	service, _, movementStore := newTestInboundQCInspectionService()
	ctx := context.Background()
	created, err := service.CreateInboundQCInspection(ctx, CreateInboundQCInspectionInput{
		ID:                 "iqc-partial",
		GoodsReceiptID:     "grn-hcm-260427-inspect",
		GoodsReceiptLineID: "grn-line-draft-001",
		ActorID:            "user-qa",
	})
	if err != nil {
		t.Fatalf("create inspection: %v", err)
	}
	if _, err := service.StartInboundQCInspection(ctx, InboundQCActionInput{ID: created.Inspection.ID, ActorID: "user-qa"}); err != nil {
		t.Fatalf("start inspection: %v", err)
	}

	_, err = service.PartialInboundQCInspection(ctx, InboundQCActionInput{
		ID:             created.Inspection.ID,
		PassedQuantity: "10",
		HoldQuantity:   "13",
		Checklist:      completedChecklistInput("pass"),
		Reason:         "sample hold",
		ActorID:        "user-qa",
	})
	if !errors.Is(err, qcdomain.ErrInboundQCInspectionInvalidQuantity) {
		t.Fatalf("err = %v, want invalid partial quantity", err)
	}

	partial, err := service.PartialInboundQCInspection(ctx, InboundQCActionInput{
		ID:             created.Inspection.ID,
		PassedQuantity: "10",
		HoldQuantity:   "14",
		Checklist:      completedChecklistInput("pass"),
		Reason:         "sample hold",
		ActorID:        "user-qa",
	})
	if err != nil {
		t.Fatalf("partial inspection: %v", err)
	}
	if partial.CurrentResult != qcdomain.InboundQCResultPartial ||
		partial.Inspection.PassedQuantity.String() != "10.000000" ||
		partial.Inspection.HoldQuantity.String() != "14.000000" {
		t.Fatalf("partial result = %+v, want 10 pass and 14 hold", partial)
	}
	movements := movementStore.Movements()
	if len(movements) != 2 {
		t.Fatalf("stock movements = %d, want available and qc hold movements for partial", len(movements))
	}
	available := findMovementByStatus(movements, inventorydomain.StockStatusAvailable)
	if available.Quantity.String() != "10.000000" {
		t.Fatalf("available movement = %+v, want 10 pass quantity", available)
	}
	hold := findMovementByStatus(movements, inventorydomain.StockStatusQCHold)
	if hold.Quantity.String() != "14.000000" {
		t.Fatalf("hold movement = %+v, want 14 quarantined quantity", hold)
	}
}

func TestInboundQCInspectionServiceHoldQuarantinesStockMovement(t *testing.T) {
	service, _, movementStore := newTestInboundQCInspectionService()
	ctx := context.Background()
	created, err := service.CreateInboundQCInspection(ctx, CreateInboundQCInspectionInput{
		ID:                 "iqc-hold",
		GoodsReceiptID:     "grn-hcm-260427-inspect",
		GoodsReceiptLineID: "grn-line-draft-001",
		ActorID:            "user-qa",
	})
	if err != nil {
		t.Fatalf("create inspection: %v", err)
	}
	if _, err := service.StartInboundQCInspection(ctx, InboundQCActionInput{ID: created.Inspection.ID, ActorID: "user-qa"}); err != nil {
		t.Fatalf("start inspection: %v", err)
	}

	held, err := service.HoldInboundQCInspection(ctx, InboundQCActionInput{
		ID:        created.Inspection.ID,
		Checklist: completedChecklistInput("pass"),
		Reason:    "pending supplier COA",
		ActorID:   "user-qa",
	})
	if err != nil {
		t.Fatalf("hold inspection: %v", err)
	}
	if held.CurrentResult != qcdomain.InboundQCResultHold ||
		held.Inspection.HoldQuantity.String() != held.Inspection.Quantity.String() {
		t.Fatalf("hold result = %+v, want full hold", held)
	}
	movements := movementStore.Movements()
	if len(movements) != 1 ||
		movements[0].StockStatus != inventorydomain.StockStatusQCHold ||
		movements[0].Quantity.String() != held.Inspection.Quantity.String() {
		t.Fatalf("stock movements = %+v, want full qc hold movement", movements)
	}
}

func TestInboundQCInspectionServiceFailBlocksStockMovementFromAvailable(t *testing.T) {
	service, _, movementStore := newTestInboundQCInspectionService()
	ctx := context.Background()
	created, err := service.CreateInboundQCInspection(ctx, CreateInboundQCInspectionInput{
		ID:                 "iqc-fail",
		GoodsReceiptID:     "grn-hcm-260427-inspect",
		GoodsReceiptLineID: "grn-line-draft-001",
		ActorID:            "user-qa",
	})
	if err != nil {
		t.Fatalf("create inspection: %v", err)
	}
	if _, err := service.StartInboundQCInspection(ctx, InboundQCActionInput{ID: created.Inspection.ID, ActorID: "user-qa"}); err != nil {
		t.Fatalf("start inspection: %v", err)
	}

	failed, err := service.FailInboundQCInspection(ctx, InboundQCActionInput{
		ID:        created.Inspection.ID,
		Checklist: completedChecklistInput("fail"),
		Reason:    "damaged packaging",
		ActorID:   "user-qa",
	})
	if err != nil {
		t.Fatalf("fail inspection: %v", err)
	}
	if failed.CurrentResult != qcdomain.InboundQCResultFail ||
		failed.Inspection.FailedQuantity.String() != failed.Inspection.Quantity.String() {
		t.Fatalf("fail result = %+v, want full fail", failed)
	}
	movements := movementStore.Movements()
	if len(movements) != 1 ||
		movements[0].StockStatus != inventorydomain.StockStatusDamaged ||
		movements[0].Quantity.String() != failed.Inspection.Quantity.String() {
		t.Fatalf("stock movements = %+v, want full damaged movement", movements)
	}
}

func newTestInboundQCInspectionService() (InboundQCInspectionService, *audit.InMemoryLogStore, *inventoryapp.InMemoryStockMovementStore) {
	auditStore := audit.NewInMemoryLogStore()
	movementStore := inventoryapp.NewInMemoryStockMovementStore()
	service := NewInboundQCInspectionService(
		NewPrototypeInboundQCInspectionStore(),
		inventoryapp.NewPrototypeWarehouseReceivingStore(),
		auditStore,
	).WithStockMovementRecorder(movementStore)
	service.clock = func() time.Time {
		return time.Date(2026, 4, 29, 9, 0, 0, 0, time.UTC)
	}

	return service, auditStore, movementStore
}

func completedChecklistInput(status string) []InboundQCChecklistInput {
	return []InboundQCChecklistInput{
		{ID: "check-packaging", Code: "PACKAGING", Label: "Packaging condition", Required: true, Status: status},
		{ID: "check-lot-expiry", Code: "LOT_EXPIRY", Label: "Lot and expiry match delivery", Required: true, Status: status},
		{ID: "check-sample", Code: "SAMPLE", Label: "Sample retained when required", Status: "not_applicable"},
	}
}

func findMovementByStatus(
	movements []inventorydomain.StockMovement,
	status inventorydomain.StockStatus,
) inventorydomain.StockMovement {
	for _, movement := range movements {
		if movement.StockStatus == status {
			return movement
		}
	}

	return inventorydomain.StockMovement{}
}
