package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestStockCountAdjustmentApprovalPostsIncreaseAndDecreaseMovements(t *testing.T) {
	ctx := context.Background()
	countStore := NewPrototypeStockCountStore()
	adjustmentStore := NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	movementStore := NewInMemoryStockMovementStore()

	createCount := NewCreateStockCount(countStore, auditStore)
	createCount.clock = func() time.Time { return time.Date(2026, 4, 28, 8, 0, 0, 0, time.UTC) }
	count, err := createCount.Execute(ctx, CreateStockCountInput{
		ID:            "count-regression-001",
		CountNo:       "CNT-REG-001",
		WarehouseID:   "wh-hcm",
		WarehouseCode: "HCM",
		Scope:         "cycle_count",
		CreatedBy:     "user-counter",
		RequestID:     "req-count-create",
		Lines: []CreateStockCountLineInput{
			{ID: "count-line-short", SKU: "SERUM-30ML", BatchID: "batch-serum", BatchNo: "LOT-A", LocationID: "bin-a01", LocationCode: "A-01", ExpectedQty: "10", BaseUOMCode: "PCS"},
			{ID: "count-line-over", SKU: "CREAM-50G", BatchID: "batch-cream", BatchNo: "LOT-B", LocationID: "bin-a02", LocationCode: "A-02", ExpectedQty: "5", BaseUOMCode: "PCS"},
		},
	})
	if err != nil {
		t.Fatalf("create count: %v", err)
	}

	submitCount := NewSubmitStockCount(countStore, auditStore)
	submitCount.clock = func() time.Time { return time.Date(2026, 4, 28, 8, 30, 0, 0, time.UTC) }
	submittedCount, err := submitCount.Execute(ctx, SubmitStockCountInput{
		ID:          count.Session.ID,
		SubmittedBy: "user-counter",
		RequestID:   "req-count-submit",
		Lines: []SubmitStockCountLineInput{
			{ID: "count-line-short", CountedQty: "8", Note: "short count"},
			{ID: "count-line-over", CountedQty: "7", Note: "over count"},
		},
	})
	if err != nil {
		t.Fatalf("submit count: %v", err)
	}
	if submittedCount.Session.Status != domain.StockCountStatusVarianceReview ||
		submittedCount.Session.Lines[0].DeltaQty != "-2.000000" ||
		submittedCount.Session.Lines[1].DeltaQty != "2.000000" {
		t.Fatalf("submitted count = %+v, want increase and decrease variances", submittedCount.Session)
	}

	createAdjustment := NewCreateStockAdjustment(adjustmentStore, auditStore)
	createAdjustment.clock = func() time.Time { return time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC) }
	adjustment, err := createAdjustment.Execute(ctx, CreateStockAdjustmentInput{
		ID:            "adj-regression-001",
		AdjustmentNo:  "ADJ-REG-001",
		WarehouseID:   submittedCount.Session.WarehouseID,
		WarehouseCode: submittedCount.Session.WarehouseCode,
		SourceType:    "stock_count",
		SourceID:      submittedCount.Session.ID,
		Reason:        "cycle count variance",
		RequestedBy:   "user-counter",
		RequestID:     "req-adjustment-create",
		Lines:         adjustmentLinesFromCount(submittedCount.Session.Lines),
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}

	transition := NewTransitionStockAdjustment(adjustmentStore, movementStore, auditStore)
	transition.clock = func() time.Time { return time.Date(2026, 4, 28, 9, 30, 0, 0, time.UTC) }
	if _, err := transition.Submit(ctx, adjustment.Adjustment.ID, "user-counter", "req-adjustment-submit"); err != nil {
		t.Fatalf("submit adjustment: %v", err)
	}
	if _, err := transition.Approve(ctx, adjustment.Adjustment.ID, "user-manager", "req-adjustment-approve"); err != nil {
		t.Fatalf("approve adjustment: %v", err)
	}
	posted, err := transition.Post(ctx, adjustment.Adjustment.ID, "user-counter", "req-adjustment-post")
	if err != nil {
		t.Fatalf("post adjustment: %v", err)
	}
	if posted.Adjustment.Status != domain.StockAdjustmentStatusPosted {
		t.Fatalf("posted adjustment = %+v, want posted", posted.Adjustment)
	}

	movements := movementStore.Movements()
	if len(movements) != 2 {
		t.Fatalf("movements = %+v, want one decrease and one increase movement", movements)
	}
	if movements[0].MovementType != domain.MovementAdjustmentOut ||
		movements[0].Quantity != "2.000000" ||
		movements[0].SourceDocID != adjustment.Adjustment.ID ||
		movements[1].MovementType != domain.MovementAdjustmentIn ||
		movements[1].Quantity != "2.000000" ||
		movements[1].SourceDocID != adjustment.Adjustment.ID {
		t.Fatalf("movements = %+v, want adjustment out 2 and in 2 tied to adjustment", movements)
	}

	countLogs, err := auditStore.List(ctx, audit.Query{EntityID: submittedCount.Session.ID})
	if err != nil {
		t.Fatalf("list count audit logs: %v", err)
	}
	if len(countLogs) != 2 || !stockAdjustmentLogActionsContain(countLogs, stockCountSubmittedAction) {
		t.Fatalf("count audit logs = %+v, want create and submit", countLogs)
	}
	adjustmentLogs, err := auditStore.List(ctx, audit.Query{EntityID: adjustment.Adjustment.ID})
	if err != nil {
		t.Fatalf("list adjustment audit logs: %v", err)
	}
	for _, action := range []string{
		stockAdjustmentCreatedAction,
		stockAdjustmentSubmittedAction,
		stockAdjustmentApprovedAction,
		stockAdjustmentPostedAction,
	} {
		if !stockAdjustmentLogActionsContain(adjustmentLogs, action) {
			t.Fatalf("adjustment audit logs = %+v, missing %s", adjustmentLogs, action)
		}
	}
}

func TestStockCountAdjustmentRejectDoesNotPostMovement(t *testing.T) {
	ctx := context.Background()
	adjustmentStore := NewPrototypeStockAdjustmentStore()
	auditStore := audit.NewInMemoryLogStore()
	movementStore := NewInMemoryStockMovementStore()

	adjustment, err := NewCreateStockAdjustment(adjustmentStore, auditStore).Execute(ctx, CreateStockAdjustmentInput{
		ID:          "adj-regression-reject",
		WarehouseID: "wh-hcm",
		SourceType:  "stock_count",
		SourceID:    "count-regression-reject",
		Reason:      "cycle count variance",
		RequestedBy: "user-counter",
		RequestID:   "req-adjustment-create",
		Lines: []CreateStockAdjustmentLineInput{
			{ID: "adj-line-reject", SKU: "SERUM-30ML", ExpectedQty: "10", CountedQty: "8", BaseUOMCode: "PCS", Reason: "short count"},
		},
	})
	if err != nil {
		t.Fatalf("create adjustment: %v", err)
	}

	transition := NewTransitionStockAdjustment(adjustmentStore, movementStore, auditStore)
	if _, err := transition.Submit(ctx, adjustment.Adjustment.ID, "user-counter", "req-adjustment-submit"); err != nil {
		t.Fatalf("submit adjustment: %v", err)
	}
	rejected, err := transition.Reject(ctx, adjustment.Adjustment.ID, "user-manager", "req-adjustment-reject")
	if err != nil {
		t.Fatalf("reject adjustment: %v", err)
	}
	if rejected.Adjustment.Status != domain.StockAdjustmentStatusRejected {
		t.Fatalf("rejected adjustment = %+v, want rejected", rejected.Adjustment)
	}
	if movementStore.Count() != 0 {
		t.Fatalf("movement count = %d, want no movement after reject", movementStore.Count())
	}
	if _, err := transition.Post(ctx, adjustment.Adjustment.ID, "user-counter", "req-adjustment-post"); !errors.Is(err, domain.ErrStockAdjustmentInvalidStatus) {
		t.Fatalf("post rejected adjustment err = %v, want invalid status", err)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: adjustment.Adjustment.ID})
	if err != nil {
		t.Fatalf("list adjustment audit logs: %v", err)
	}
	if !stockAdjustmentLogActionsContain(logs, stockAdjustmentRejectedAction) {
		t.Fatalf("audit logs = %+v, want rejected action", logs)
	}
}

func adjustmentLinesFromCount(lines []domain.StockCountLine) []CreateStockAdjustmentLineInput {
	result := make([]CreateStockAdjustmentLineInput, 0, len(lines))
	for _, line := range lines {
		result = append(result, CreateStockAdjustmentLineInput{
			ID:           "adj-" + line.ID,
			ItemID:       line.ItemID,
			SKU:          line.SKU,
			BatchID:      line.BatchID,
			BatchNo:      line.BatchNo,
			LocationID:   line.LocationID,
			LocationCode: line.LocationCode,
			ExpectedQty:  line.ExpectedQty.String(),
			CountedQty:   line.CountedQty.String(),
			BaseUOMCode:  line.BaseUOMCode.String(),
			Reason:       line.Note,
		})
	}

	return result
}
