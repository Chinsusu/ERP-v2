package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
)

var ErrBatchNotFound = errors.New("batch not found")

type BatchCatalog struct {
	mu      sync.RWMutex
	batches map[string]domain.Batch
}

func NewPrototypeBatchCatalog() *BatchCatalog {
	catalog := &BatchCatalog{batches: make(map[string]domain.Batch)}
	for _, batch := range prototypeBatches() {
		catalog.batches[batch.ID] = batch
	}

	return catalog
}

func (c *BatchCatalog) ListBatches(_ context.Context, filter domain.BatchFilter) ([]domain.Batch, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	batches := make([]domain.Batch, 0, len(c.batches))
	for _, batch := range c.batches {
		if filter.Matches(batch) {
			batches = append(batches, batch)
		}
	}
	domain.SortBatches(batches)

	return batches, nil
}

func (c *BatchCatalog) GetBatch(_ context.Context, id string) (domain.Batch, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	batch, ok := c.batches[strings.TrimSpace(id)]
	if !ok {
		return domain.Batch{}, ErrBatchNotFound
	}

	return batch, nil
}

func prototypeBatches() []domain.Batch {
	now := time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC)
	batches := []domain.Batch{
		mustBatch(domain.NewBatch(domain.NewBatchInput{
			ID:         "batch-serum-2604a",
			OrgID:      "org-local",
			ItemID:     "item-serum-30ml",
			SKU:        "SERUM-30ML",
			ItemName:   "Vitamin C Serum",
			BatchNo:    "LOT-2604A",
			SupplierID: "supplier-local",
			MfgDate:    time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			ExpiryDate: time.Date(2027, 4, 1, 0, 0, 0, 0, time.UTC),
			QCStatus:   domain.QCStatusHold,
			Status:     domain.BatchStatusActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		})),
		mustBatch(domain.NewBatch(domain.NewBatchInput{
			ID:         "batch-cream-2603b",
			OrgID:      "org-local",
			ItemID:     "item-cream-50g",
			SKU:        "CREAM-50G",
			ItemName:   "Moisturizing Cream",
			BatchNo:    "LOT-2603B",
			SupplierID: "supplier-local",
			MfgDate:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			ExpiryDate: time.Date(2028, 3, 1, 0, 0, 0, 0, time.UTC),
			QCStatus:   domain.QCStatusPass,
			Status:     domain.BatchStatusActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		})),
		mustBatch(domain.NewBatch(domain.NewBatchInput{
			ID:         "batch-toner-2604c",
			OrgID:      "org-local",
			ItemID:     "item-toner-100ml",
			SKU:        "TONER-100ML",
			ItemName:   "Hydrating Toner",
			BatchNo:    "LOT-2604C",
			SupplierID: "supplier-local",
			MfgDate:    time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
			ExpiryDate: time.Date(2027, 10, 10, 0, 0, 0, 0, time.UTC),
			QCStatus:   domain.QCStatusFail,
			Status:     domain.BatchStatusBlocked,
			CreatedAt:  now,
			UpdatedAt:  now,
		})),
	}
	return batches
}

func mustBatch(batch domain.Batch, err error) domain.Batch {
	if err != nil {
		panic(err)
	}

	return batch
}
