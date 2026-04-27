package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestBatchCatalogListsFilteredPrototypeBatches(t *testing.T) {
	catalog := NewPrototypeBatchCatalog()
	filter := domain.NewBatchFilter("serum-30ml", domain.QCStatusHold, domain.BatchStatusActive)

	batches, err := catalog.ListBatches(context.Background(), filter)
	if err != nil {
		t.Fatalf("list batches: %v", err)
	}
	if len(batches) != 1 {
		t.Fatalf("batches length = %d, want 1", len(batches))
	}
	if batches[0].ID != "batch-serum-2604a" || batches[0].QCStatus != domain.QCStatusHold {
		t.Fatalf("batch = %+v, want hold serum batch", batches[0])
	}
}

func TestBatchCatalogGetsBatchByID(t *testing.T) {
	catalog := NewPrototypeBatchCatalog()

	batch, err := catalog.GetBatch(context.Background(), "batch-cream-2603b")
	if err != nil {
		t.Fatalf("get batch: %v", err)
	}
	if batch.SKU != "CREAM-50G" || batch.QCStatus != domain.QCStatusPass {
		t.Fatalf("batch = %+v, want pass cream batch", batch)
	}

	if _, err := catalog.GetBatch(context.Background(), "missing"); !errors.Is(err, ErrBatchNotFound) {
		t.Fatalf("missing batch error = %v, want ErrBatchNotFound", err)
	}
}

func TestBatchCatalogChangesQCStatusAndWritesAudit(t *testing.T) {
	ctx := context.Background()
	auditStore := audit.NewInMemoryLogStore()
	catalog := NewPrototypeBatchCatalog(auditStore)
	changedAt := time.Date(2026, 4, 27, 10, 0, 0, 0, time.UTC)

	result, err := catalog.ChangeQCStatus(ctx, ChangeBatchQCStatusInput{
		BatchID:     "batch-serum-2604a",
		NextStatus:  domain.QCStatusPass,
		ActorID:     "user-qa",
		Reason:      "COA and visual inspection passed",
		BusinessRef: "QC-260427-0001",
		RequestID:   "req-qc-pass",
		ChangedAt:   changedAt,
	})
	if err != nil {
		t.Fatalf("change qc status: %v", err)
	}
	if result.Batch.QCStatus != domain.QCStatusPass {
		t.Fatalf("qc status = %q, want pass", result.Batch.QCStatus)
	}
	if result.Transition.FromQCStatus != domain.QCStatusHold || result.Transition.ToQCStatus != domain.QCStatusPass {
		t.Fatalf("transition = %+v, want hold -> pass", result.Transition)
	}
	if result.Transition.Reason != "COA and visual inspection passed" || result.Transition.BusinessRef != "QC-260427-0001" {
		t.Fatalf("transition metadata = %+v, want reason and business ref", result.Transition)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: "batch-serum-2604a"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(logs))
	}
	if logs[0].Action != batchQCTransitionAction || logs[0].BeforeData["qc_status"] != "hold" || logs[0].AfterData["qc_status"] != "pass" {
		t.Fatalf("audit log = %+v, want before/after qc status", logs[0])
	}

	history, err := catalog.ListQCTransitions(ctx, "batch-serum-2604a")
	if err != nil {
		t.Fatalf("list qc transitions: %v", err)
	}
	if len(history) != 1 || history[0].AuditLogID == "" {
		t.Fatalf("history = %+v, want one transition with audit id", history)
	}
}

func TestBatchCatalogRejectsMissingReasonAndInvalidTransition(t *testing.T) {
	catalog := NewPrototypeBatchCatalog(audit.NewInMemoryLogStore())

	_, err := catalog.ChangeQCStatus(context.Background(), ChangeBatchQCStatusInput{
		BatchID:    "batch-serum-2604a",
		NextStatus: domain.QCStatusPass,
		ActorID:    "user-qa",
	})
	if !errors.Is(err, ErrBatchTransitionReasonRequired) {
		t.Fatalf("missing reason error = %v, want ErrBatchTransitionReasonRequired", err)
	}

	_, err = catalog.ChangeQCStatus(context.Background(), ChangeBatchQCStatusInput{
		BatchID:    "batch-cream-2603b",
		NextStatus: domain.QCStatusFail,
		ActorID:    "user-qa",
		Reason:     "late defect found",
	})
	if !errors.Is(err, domain.ErrBatchInvalidQCTransition) {
		t.Fatalf("invalid transition error = %v, want ErrBatchInvalidQCTransition", err)
	}
}
