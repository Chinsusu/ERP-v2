package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
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
