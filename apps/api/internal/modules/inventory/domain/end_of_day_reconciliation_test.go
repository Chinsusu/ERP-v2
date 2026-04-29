package domain

import (
	"errors"
	"testing"
	"time"
)

func TestEndOfDayReconciliationSummaryShowsVarianceAndChecklist(t *testing.T) {
	reconciliation := EndOfDayReconciliation{
		Status: ReconciliationStatusInReview,
		Checklist: []ReconciliationChecklistItem{
			{Key: "shipment", Label: "Shipment", Complete: true, Blocking: true},
			{Key: "variance", Label: "Variance", Complete: false, Blocking: true},
		},
		Lines: []ReconciliationLine{
			{SystemQuantity: 120, CountedQuantity: 118},
			{SystemQuantity: 40, CountedQuantity: 40},
		},
	}

	summary := reconciliation.Summary("")

	if summary.SystemQuantity != 160 {
		t.Fatalf("system quantity = %d, want 160", summary.SystemQuantity)
	}
	if summary.CountedQuantity != 158 {
		t.Fatalf("counted quantity = %d, want 158", summary.CountedQuantity)
	}
	if summary.VarianceQuantity != -2 {
		t.Fatalf("variance quantity = %d, want -2", summary.VarianceQuantity)
	}
	if summary.VarianceCount != 1 {
		t.Fatalf("variance count = %d, want 1", summary.VarianceCount)
	}
	if summary.ReadyToClose {
		t.Fatal("ready to close without exception note = true, want false")
	}
	if !reconciliation.Summary("variance accepted by lead").ReadyToClose {
		t.Fatal("ready to close with exception note = false, want true")
	}
}

func TestEndOfDayReconciliationCloseRequiresExceptionForBlockingChecklist(t *testing.T) {
	reconciliation := EndOfDayReconciliation{
		ID:     "close-1",
		Status: ReconciliationStatusInReview,
		Checklist: []ReconciliationChecklistItem{
			{Key: "variance", Label: "Variance", Complete: false, Blocking: true},
		},
	}

	_, err := reconciliation.Close("user-warehouse-lead", "", time.Time{})
	if !errors.Is(err, ErrReconciliationNeedsExceptionNote) {
		t.Fatalf("close err = %v, want ErrReconciliationNeedsExceptionNote", err)
	}

	closedAt := time.Date(2026, 4, 26, 17, 30, 0, 0, time.UTC)
	closed, err := reconciliation.Close("user-warehouse-lead", "variance documented", closedAt)
	if err != nil {
		t.Fatalf("close with note: %v", err)
	}
	if closed.Status != ReconciliationStatusClosed {
		t.Fatalf("status = %q, want closed", closed.Status)
	}
	if closed.ClosedBy != "user-warehouse-lead" {
		t.Fatalf("closed by = %q, want user-warehouse-lead", closed.ClosedBy)
	}
	if !closed.ClosedAt.Equal(closedAt) {
		t.Fatalf("closed at = %v, want %v", closed.ClosedAt, closedAt)
	}
}

func TestEndOfDayReconciliationCloseBlocksUnresolvedOperationalIssue(t *testing.T) {
	reconciliation := EndOfDayReconciliation{
		ID:     "close-returns-pending",
		Status: ReconciliationStatusInReview,
		Checklist: []ReconciliationChecklistItem{
			{Key: "returns", Label: "Returns triaged", Complete: false, Blocking: true},
			{Key: "adjustments", Label: "Adjustments approved", Complete: true, Blocking: true},
		},
	}

	if reconciliation.Summary("lead exception").ReadyToClose {
		t.Fatal("ready to close with unresolved operational issue = true, want false")
	}

	_, err := reconciliation.Close("user-warehouse-lead", "lead exception", time.Time{})
	if !errors.Is(err, ErrReconciliationUnresolvedIssue) {
		t.Fatalf("close err = %v, want ErrReconciliationUnresolvedIssue", err)
	}
	if len(reconciliation.UnresolvedOperationalIssues()) != 1 {
		t.Fatalf("unresolved issues = %d, want 1", len(reconciliation.UnresolvedOperationalIssues()))
	}
}
