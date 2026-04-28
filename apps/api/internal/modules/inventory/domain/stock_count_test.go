package domain

import (
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"
)

func TestStockCountSessionSubmitMovesToVarianceReview(t *testing.T) {
	session, err := NewStockCountSession(NewStockCountSessionInput{
		ID:          "count-test",
		CountNo:     "CNT-TEST",
		WarehouseID: "wh-hcm",
		CreatedBy:   "user-warehouse-lead",
		Lines: []NewStockCountLineInput{
			{ID: "line-serum", SKU: "SERUM-30ML", ExpectedQty: decimal.MustQuantity("10"), BaseUOMCode: "EA"},
		},
		CreatedAt: time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new stock count session: %v", err)
	}

	submitted, err := session.Submit([]SubmitStockCountLineInput{
		{ID: "line-serum", CountedQty: decimal.MustQuantity("8"), Note: "short count"},
	}, "user-counter", time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("submit stock count: %v", err)
	}

	if submitted.Status != StockCountStatusVarianceReview ||
		submitted.Lines[0].DeltaQty != "-2.000000" ||
		submitted.SubmittedBy != "user-counter" {
		t.Fatalf("submitted = %+v, want variance review with delta", submitted)
	}
}

func TestStockCountSessionSubmitRequiresAllLinesCounted(t *testing.T) {
	session, err := NewStockCountSession(NewStockCountSessionInput{
		WarehouseID: "wh-hcm",
		CreatedBy:   "user-warehouse-lead",
		Lines: []NewStockCountLineInput{
			{ID: "line-serum", SKU: "SERUM-30ML", ExpectedQty: decimal.MustQuantity("10"), BaseUOMCode: "EA"},
			{ID: "line-cream", SKU: "CREAM-50G", ExpectedQty: decimal.MustQuantity("5"), BaseUOMCode: "EA"},
		},
	})
	if err != nil {
		t.Fatalf("new stock count session: %v", err)
	}

	_, err = session.Submit([]SubmitStockCountLineInput{
		{ID: "line-serum", CountedQty: decimal.MustQuantity("10")},
	}, "user-counter", time.Time{})
	if !errors.Is(err, ErrStockCountRequiredField) {
		t.Fatalf("err = %v, want required field for uncounted line", err)
	}
}
