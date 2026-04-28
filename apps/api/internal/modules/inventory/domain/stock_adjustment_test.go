package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestNewStockAdjustmentCreatesDraftRequestWithVariance(t *testing.T) {
	adjustment, err := NewStockAdjustment(NewStockAdjustmentInput{
		ID:           "adj-test",
		AdjustmentNo: "ADJ-TEST",
		WarehouseID:  "wh-hcm",
		Reason:       "cycle count variance",
		RequestedBy:  "user-warehouse-lead",
		Lines: []NewStockAdjustmentLineInput{
			{
				SKU:         "serum-30ml",
				ExpectedQty: decimal.MustQuantity("10"),
				CountedQty:  decimal.MustQuantity("8.5"),
				BaseUOMCode: "EA",
				Reason:      "counted short",
			},
		},
		CreatedAt: time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new stock adjustment: %v", err)
	}

	if adjustment.Status != StockAdjustmentStatusDraft {
		t.Fatalf("status = %q, want draft", adjustment.Status)
	}
	if adjustment.Lines[0].SKU != "SERUM-30ML" ||
		adjustment.Lines[0].ExpectedQty != "10.000000" ||
		adjustment.Lines[0].CountedQty != "8.500000" ||
		adjustment.Lines[0].DeltaQty != "-1.500000" {
		t.Fatalf("line = %+v, want normalized negative variance", adjustment.Lines[0])
	}
}

func TestNewStockAdjustmentRejectsMissingReasonAndZeroVariance(t *testing.T) {
	_, err := NewStockAdjustment(NewStockAdjustmentInput{
		WarehouseID: "wh-hcm",
		RequestedBy: "user-warehouse-lead",
		Lines: []NewStockAdjustmentLineInput{
			{
				SKU:         "SERUM-30ML",
				ExpectedQty: decimal.MustQuantity("10"),
				CountedQty:  decimal.MustQuantity("9"),
				BaseUOMCode: "EA",
			},
		},
	})
	if !errors.Is(err, ErrStockAdjustmentRequiredField) {
		t.Fatalf("err = %v, want required field", err)
	}

	_, err = NewStockAdjustment(NewStockAdjustmentInput{
		WarehouseID: "wh-hcm",
		Reason:      "no variance",
		RequestedBy: "user-warehouse-lead",
		Lines: []NewStockAdjustmentLineInput{
			{
				SKU:         "SERUM-30ML",
				ExpectedQty: decimal.MustQuantity("10"),
				CountedQty:  decimal.MustQuantity("10"),
				BaseUOMCode: "EA",
			},
		},
	})
	if !errors.Is(err, ErrStockAdjustmentNoVariance) {
		t.Fatalf("err = %v, want no variance", err)
	}
}
