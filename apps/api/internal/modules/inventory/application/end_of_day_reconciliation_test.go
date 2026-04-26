package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestPrototypeEndOfDayReconciliationStoreFiltersRows(t *testing.T) {
	store := NewPrototypeEndOfDayReconciliationStore()
	usecase := NewListEndOfDayReconciliations(store)

	rows, err := usecase.Execute(context.Background(), domain.NewEndOfDayReconciliationFilter(
		"wh-hcm",
		"2026-04-26",
		domain.ReconciliationStatusInReview,
	))
	if err != nil {
		t.Fatalf("list reconciliations: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("rows = %d, want 1", len(rows))
	}
	if rows[0].ID != "rec-hcm-260426-day" {
		t.Fatalf("row id = %q, want rec-hcm-260426-day", rows[0].ID)
	}
}

func TestCloseEndOfDayReconciliationRecordsAuditLog(t *testing.T) {
	store := NewPrototypeEndOfDayReconciliationStore()
	auditStore := audit.NewInMemoryLogStore()
	usecase := NewCloseEndOfDayReconciliation(store, auditStore)
	closedAt := time.Date(2026, 4, 26, 17, 45, 0, 0, time.UTC)
	usecase.clock = func() time.Time { return closedAt }

	result, err := usecase.Execute(context.Background(), CloseEndOfDayReconciliationInput{
		ID:            "rec-hcm-260426-day",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-close-1",
		ExceptionNote: "Variance accepted by warehouse lead",
	})
	if err != nil {
		t.Fatalf("close reconciliation: %v", err)
	}
	if result.Reconciliation.Status != domain.ReconciliationStatusClosed {
		t.Fatalf("status = %q, want closed", result.Reconciliation.Status)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}

	logs, err := auditStore.List(context.Background(), audit.Query{Action: "warehouse.shift.closed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].EntityID != "rec-hcm-260426-day" {
		t.Fatalf("audit entity id = %q, want rec-hcm-260426-day", logs[0].EntityID)
	}
	if logs[0].AfterData["variance_count"] != 1 {
		t.Fatalf("audit variance count = %v, want 1", logs[0].AfterData["variance_count"])
	}
}

func TestCloseEndOfDayReconciliationReturnsNotFound(t *testing.T) {
	store := NewPrototypeEndOfDayReconciliationStore()
	usecase := NewCloseEndOfDayReconciliation(store, audit.NewInMemoryLogStore())

	_, err := usecase.Execute(context.Background(), CloseEndOfDayReconciliationInput{
		ID:            "missing",
		ActorID:       "user-warehouse-lead",
		ExceptionNote: "n/a",
	})
	if !errors.Is(err, ErrEndOfDayReconciliationNotFound) {
		t.Fatalf("err = %v, want ErrEndOfDayReconciliationNotFound", err)
	}
}
