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
