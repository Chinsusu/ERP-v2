package application

import (
	"context"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestCreateAndSubmitStockCountSession(t *testing.T) {
	store := NewPrototypeStockCountStore()
	auditStore := audit.NewInMemoryLogStore()
	create := NewCreateStockCount(store, auditStore)
	create.clock = func() time.Time { return time.Date(2026, 4, 28, 9, 0, 0, 0, time.UTC) }

	created, err := create.Execute(context.Background(), CreateStockCountInput{
		ID:          "count-hcm-001",
		CountNo:     "CNT-HCM-001",
		WarehouseID: "wh-hcm",
		Scope:       "cycle_count",
		CreatedBy:   "user-warehouse-lead",
		RequestID:   "req-count-create",
		Lines: []CreateStockCountLineInput{
			{ID: "line-serum", SKU: "SERUM-30ML", ExpectedQty: "20", BaseUOMCode: "EA"},
		},
	})
	if err != nil {
		t.Fatalf("create stock count: %v", err)
	}
	if created.Session.Status != domain.StockCountStatusOpen || created.AuditLogID == "" {
		t.Fatalf("created = %+v, want open session with audit", created)
	}

	submit := NewSubmitStockCount(store, auditStore)
	submit.clock = func() time.Time { return time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC) }
	submitted, err := submit.Execute(context.Background(), SubmitStockCountInput{
		ID:          "count-hcm-001",
		SubmittedBy: "user-counter",
		RequestID:   "req-count-submit",
		Lines: []SubmitStockCountLineInput{
			{ID: "line-serum", CountedQty: "18", Note: "short count"},
		},
	})
	if err != nil {
		t.Fatalf("submit stock count: %v", err)
	}
	if submitted.Session.Status != domain.StockCountStatusVarianceReview ||
		submitted.Session.Lines[0].DeltaQty != "-2.000000" ||
		submitted.AuditLogID == "" {
		t.Fatalf("submitted = %+v, want variance review with audit", submitted)
	}

	logs, err := auditStore.List(context.Background(), audit.Query{EntityID: "count-hcm-001"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 2 || logs[0].Action != stockCountSubmittedAction || logs[1].Action != stockCountCreatedAction {
		t.Fatalf("audit logs = %+v, want created and submitted logs", logs)
	}
}

func TestSubmitStockCountSessionWithoutVarianceMovesSubmitted(t *testing.T) {
	store := NewPrototypeStockCountStore()
	auditStore := audit.NewInMemoryLogStore()
	create := NewCreateStockCount(store, auditStore)
	created, err := create.Execute(context.Background(), CreateStockCountInput{
		WarehouseID: "wh-hcm",
		CreatedBy:   "user-warehouse-lead",
		Lines: []CreateStockCountLineInput{
			{ID: "line-serum", SKU: "SERUM-30ML", ExpectedQty: "20", BaseUOMCode: "EA"},
		},
	})
	if err != nil {
		t.Fatalf("create stock count: %v", err)
	}

	submitted, err := NewSubmitStockCount(store, auditStore).Execute(context.Background(), SubmitStockCountInput{
		ID:          created.Session.ID,
		SubmittedBy: "user-counter",
		Lines: []SubmitStockCountLineInput{
			{ID: "line-serum", CountedQty: "20"},
		},
	})
	if err != nil {
		t.Fatalf("submit stock count: %v", err)
	}
	if submitted.Session.Status != domain.StockCountStatusSubmitted {
		t.Fatalf("status = %q, want submitted", submitted.Session.Status)
	}
}
