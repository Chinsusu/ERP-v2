package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestCreateStockAdjustmentCreatesDraftRequestAndAudit(t *testing.T) {
	store := NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCreateStockAdjustment(store, auditStore)
	service.clock = func() time.Time { return time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC) }

	result, err := service.Execute(context.Background(), CreateStockAdjustmentInput{
		WarehouseID:  "wh-hcm",
		Reason:       "cycle count variance",
		RequestedBy:  "user-warehouse-lead",
		RequestID:    "req-adjustment-create",
		SourceType:   "stock_count",
		SourceID:     "count-hcm-001",
		AdjustmentNo: "ADJ-HCM-001",
		Lines: []CreateStockAdjustmentLineInput{
			{
				SKU:         "SERUM-30ML",
				ExpectedQty: "20",
				CountedQty:  "18",
				BaseUOMCode: "EA",
				Reason:      "short count",
			},
		},
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}

	if result.Adjustment.Status != domain.StockAdjustmentStatusDraft ||
		result.Adjustment.Lines[0].DeltaQty != "-2.000000" ||
		result.AuditLogID == "" {
		t.Fatalf("result = %+v, want draft adjustment with variance and audit", result)
	}
	rows, err := NewListStockAdjustments(store).Execute(context.Background())
	if err != nil {
		t.Fatalf("list adjustments: %v", err)
	}
	if len(rows) != 1 || rows[0].ID != result.Adjustment.ID {
		t.Fatalf("rows = %+v, want saved adjustment", rows)
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: stockAdjustmentCreatedAction})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 ||
		logs[0].EntityType != stockAdjustmentEntityType ||
		logs[0].AfterData["status"] != "draft" ||
		logs[0].AfterData["line_count"] != 1 {
		t.Fatalf("audit logs = %+v, want created adjustment audit", logs)
	}
}

func TestCreateStockAdjustmentRejectsNoVarianceAndDoesNotWriteStockMovement(t *testing.T) {
	store := NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	service := NewCreateStockAdjustment(store, auditStore)

	_, err := service.Execute(context.Background(), CreateStockAdjustmentInput{
		WarehouseID: "wh-hcm",
		Reason:      "no variance",
		RequestedBy: "user-warehouse-lead",
		Lines: []CreateStockAdjustmentLineInput{
			{
				SKU:         "SERUM-30ML",
				ExpectedQty: "20",
				CountedQty:  "20",
				BaseUOMCode: "EA",
			},
		},
	})
	if !errors.Is(err, domain.ErrStockAdjustmentNoVariance) {
		t.Fatalf("err = %v, want no variance", err)
	}
	rows, err := NewListStockAdjustments(store).Execute(context.Background())
	if err != nil {
		t.Fatalf("list adjustments: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("rows = %+v, want no saved adjustment", rows)
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: "inventory.stock_movement.adjusted"})
	if err != nil {
		t.Fatalf("list stock movement audit logs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("stock movement audit logs = %+v, want no direct stock mutation", logs)
	}
}

func TestTransitionStockAdjustmentApprovesAndPostsMovement(t *testing.T) {
	store := NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	movementStore := NewInMemoryStockMovementStore()
	create := NewCreateStockAdjustment(store, auditStore)
	create.clock = func() time.Time { return time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC) }
	created, err := create.Execute(context.Background(), CreateStockAdjustmentInput{
		ID:          "adj-hcm-001",
		WarehouseID: "wh-hcm",
		Reason:      "cycle count variance",
		RequestedBy: "user-warehouse-lead",
		Lines: []CreateStockAdjustmentLineInput{
			{ID: "line-short", SKU: "SERUM-30ML", ExpectedQty: "20", CountedQty: "18", BaseUOMCode: "EA"},
			{ID: "line-over", SKU: "CREAM-50G", ExpectedQty: "5", CountedQty: "6", BaseUOMCode: "EA"},
		},
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}

	service := NewTransitionStockAdjustment(store, movementStore, auditStore)
	service.clock = func() time.Time { return time.Date(2026, 4, 28, 11, 0, 0, 0, time.UTC) }
	submitted, err := service.Submit(context.Background(), created.Adjustment.ID, "user-warehouse-lead", "req-submit")
	if err != nil {
		t.Fatalf("submit adjustment: %v", err)
	}
	if submitted.Adjustment.Status != domain.StockAdjustmentStatusSubmitted {
		t.Fatalf("submit status = %q, want submitted", submitted.Adjustment.Status)
	}
	approved, err := service.Approve(context.Background(), created.Adjustment.ID, "user-manager", "req-approve")
	if err != nil {
		t.Fatalf("approve adjustment: %v", err)
	}
	if approved.Adjustment.Status != domain.StockAdjustmentStatusApproved || approved.Adjustment.ApprovedBy != "user-manager" {
		t.Fatalf("approved adjustment = %+v, want approved by manager", approved.Adjustment)
	}
	posted, err := service.Post(context.Background(), created.Adjustment.ID, "user-warehouse-lead", "req-post")
	if err != nil {
		t.Fatalf("post adjustment: %v", err)
	}
	if posted.Adjustment.Status != domain.StockAdjustmentStatusPosted {
		t.Fatalf("posted status = %q, want posted", posted.Adjustment.Status)
	}

	movements := movementStore.Movements()
	if len(movements) != 2 {
		t.Fatalf("movements = %+v, want two adjustment movements", movements)
	}
	if movements[0].MovementType != domain.MovementAdjustmentOut ||
		movements[0].Quantity != "2.000000" ||
		movements[1].MovementType != domain.MovementAdjustmentIn ||
		movements[1].Quantity != "1.000000" {
		t.Fatalf("movements = %+v, want out 2 and in 1", movements)
	}
	logs, err := auditStore.List(context.Background(), audit.Query{EntityID: created.Adjustment.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 4 || !stockAdjustmentLogActionsContain(logs, stockAdjustmentPostedAction) {
		t.Fatalf("audit logs = %+v, want create/submit/approve/post", logs)
	}
}

func TestTransitionStockAdjustmentRejectsBeforePost(t *testing.T) {
	store := NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	created, err := NewCreateStockAdjustment(store, auditStore).Execute(context.Background(), CreateStockAdjustmentInput{
		ID:          "adj-reject",
		WarehouseID: "wh-hcm",
		Reason:      "cycle count variance",
		RequestedBy: "user-warehouse-lead",
		Lines: []CreateStockAdjustmentLineInput{
			{ID: "line-short", SKU: "SERUM-30ML", ExpectedQty: "20", CountedQty: "18", BaseUOMCode: "EA"},
		},
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}
	service := NewTransitionStockAdjustment(store, NewInMemoryStockMovementStore(), auditStore)
	if _, err := service.Submit(context.Background(), created.Adjustment.ID, "user-warehouse-lead", "req-submit"); err != nil {
		t.Fatalf("submit adjustment: %v", err)
	}
	rejected, err := service.Reject(context.Background(), created.Adjustment.ID, "user-manager", "req-reject")
	if err != nil {
		t.Fatalf("reject adjustment: %v", err)
	}
	if rejected.Adjustment.Status != domain.StockAdjustmentStatusRejected {
		t.Fatalf("status = %q, want rejected", rejected.Adjustment.Status)
	}
	if _, err := service.Post(context.Background(), created.Adjustment.ID, "user-warehouse-lead", "req-post"); !errors.Is(err, domain.ErrStockAdjustmentInvalidStatus) {
		t.Fatalf("post err = %v, want invalid status", err)
	}
}

func stockAdjustmentLogActionsContain(logs []audit.Log, action string) bool {
	for _, log := range logs {
		if log.Action == action {
			return true
		}
	}

	return false
}
