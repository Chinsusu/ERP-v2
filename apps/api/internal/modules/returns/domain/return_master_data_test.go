package domain

import "testing"

func TestPrototypeReturnMasterDataContainsStandardConditionsAndDispositions(t *testing.T) {
	data := PrototypeReturnMasterData()

	if len(data.Reasons) < 3 {
		t.Fatalf("return reasons = %d, want basic seed data", len(data.Reasons))
	}

	wantConditions := map[string]ReturnDisposition{
		"sealed_good":             ReturnDispositionReusable,
		"opened_good":             ReturnDispositionNeedsInspection,
		"damaged":                 ReturnDispositionNotReusable,
		"expired":                 ReturnDispositionNotReusable,
		"suspected_quality_issue": ReturnDispositionNeedsInspection,
	}
	for code, disposition := range wantConditions {
		condition, ok := findReturnCondition(data.Conditions, code)
		if !ok {
			t.Fatalf("condition %q not found", code)
		}
		if condition.DefaultDisposition != disposition {
			t.Fatalf("condition %q disposition = %q, want %q", code, condition.DefaultDisposition, disposition)
		}
	}

	wantDispositions := map[ReturnDisposition]string{
		ReturnDispositionReusable:        "return_pending",
		ReturnDispositionNotReusable:     "damaged",
		ReturnDispositionNeedsInspection: "qc_hold",
	}
	for code, stockStatus := range wantDispositions {
		disposition, ok := findReturnDispositionMaster(data.Dispositions, code)
		if !ok {
			t.Fatalf("disposition %q not found", code)
		}
		if disposition.TargetStockStatus != stockStatus {
			t.Fatalf("disposition %q stock status = %q, want %q", code, disposition.TargetStockStatus, stockStatus)
		}
		if disposition.CreatesAvailableStock {
			t.Fatalf("disposition %q creates available stock at receiving", code)
		}
	}
}

func TestPrototypeReturnMasterDataIsSortedBySortOrder(t *testing.T) {
	data := PrototypeReturnMasterData()

	for i := 1; i < len(data.Reasons); i++ {
		if data.Reasons[i-1].SortOrder > data.Reasons[i].SortOrder {
			t.Fatalf("reasons are not sorted: %+v", data.Reasons)
		}
	}
	for i := 1; i < len(data.Conditions); i++ {
		if data.Conditions[i-1].SortOrder > data.Conditions[i].SortOrder {
			t.Fatalf("conditions are not sorted: %+v", data.Conditions)
		}
	}
	for i := 1; i < len(data.Dispositions); i++ {
		if data.Dispositions[i-1].SortOrder > data.Dispositions[i].SortOrder {
			t.Fatalf("dispositions are not sorted: %+v", data.Dispositions)
		}
	}
}

func findReturnCondition(rows []ReturnCondition, code string) (ReturnCondition, bool) {
	for _, row := range rows {
		if row.Code == code {
			return row, true
		}
	}

	return ReturnCondition{}, false
}

func findReturnDispositionMaster(rows []ReturnDispositionMaster, code ReturnDisposition) (ReturnDispositionMaster, bool) {
	for _, row := range rows {
		if row.Code == code {
			return row, true
		}
	}

	return ReturnDispositionMaster{}, false
}
