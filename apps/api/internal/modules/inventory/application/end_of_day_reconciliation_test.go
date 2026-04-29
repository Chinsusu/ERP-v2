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
		"day",
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
	if rows[0].Operations.HandoverOrderCount != 27 {
		t.Fatalf("handover order count = %d, want 27", rows[0].Operations.HandoverOrderCount)
	}
}

func TestCloseEndOfDayReconciliationRecordsAuditLog(t *testing.T) {
	store := NewPrototypeEndOfDayReconciliationStore()
	auditStore := audit.NewInMemoryLogStore()
	usecase := NewCloseEndOfDayReconciliation(store, auditStore)
	closedAt := time.Date(2026, 4, 26, 17, 45, 0, 0, time.UTC)
	usecase.clock = func() time.Time { return closedAt }

	result, err := usecase.Execute(context.Background(), CloseEndOfDayReconciliationInput{
		ID:            "rec-hn-260426-day",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-close-1",
		ExceptionNote: "",
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
	if logs[0].EntityID != "rec-hn-260426-day" {
		t.Fatalf("audit entity id = %q, want rec-hn-260426-day", logs[0].EntityID)
	}
	if logs[0].AfterData["variance_count"] != 0 {
		t.Fatalf("audit variance count = %v, want 0", logs[0].AfterData["variance_count"])
	}
	if logs[0].AfterData["order_count"] != 18 {
		t.Fatalf("audit order count = %v, want 18", logs[0].AfterData["order_count"])
	}
	if logs[0].AfterData["handover_count"] != 14 {
		t.Fatalf("audit handover count = %v, want 14", logs[0].AfterData["handover_count"])
	}
	if logs[0].AfterData["pending_issue_count"] != 0 {
		t.Fatalf("audit pending issue count = %v, want 0", logs[0].AfterData["pending_issue_count"])
	}
}

func TestCloseEndOfDayReconciliationBlocksUnresolvedOperationalIssue(t *testing.T) {
	store := NewPrototypeEndOfDayReconciliationStore()
	auditStore := audit.NewInMemoryLogStore()
	usecase := NewCloseEndOfDayReconciliation(store, auditStore)

	_, err := usecase.Execute(context.Background(), CloseEndOfDayReconciliationInput{
		ID:            "rec-hcm-260426-day",
		ActorID:       "user-warehouse-lead",
		RequestID:     "req-close-blocked",
		ExceptionNote: "Variance accepted by warehouse lead",
	})
	if !errors.Is(err, domain.ErrReconciliationUnresolvedIssue) {
		t.Fatalf("close err = %v, want ErrReconciliationUnresolvedIssue", err)
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: "warehouse.shift.closed"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 0 {
		t.Fatalf("audit logs = %d, want 0", len(logs))
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
