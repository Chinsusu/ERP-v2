package application

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

var ErrBatchNotFound = errors.New("batch not found")
var ErrBatchTransitionActorRequired = errors.New("batch qc transition actor is required")
var ErrBatchTransitionReasonRequired = errors.New("batch qc transition reason is required")

const batchQCTransitionAction = "inventory.batch.qc_status_changed"
const batchQCTransitionEntityType = "inventory.batch"

type ChangeBatchQCStatusInput struct {
	BatchID     string
	NextStatus  domain.QCStatus
	ActorID     string
	Reason      string
	BusinessRef string
	RequestID   string
	ChangedAt   time.Time
}

type ChangeBatchQCStatusResult struct {
	Batch      domain.Batch
	Transition domain.BatchQCTransition
	AuditLogID string
}

type BatchCatalog struct {
	mu       sync.RWMutex
	batches  map[string]domain.Batch
	auditLog audit.LogStore
}

func NewPrototypeBatchCatalog(auditStores ...audit.LogStore) *BatchCatalog {
	auditLog := firstBatchAuditStore(auditStores)
	catalog := &BatchCatalog{batches: make(map[string]domain.Batch), auditLog: auditLog}
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

func (c *BatchCatalog) ChangeQCStatus(
	ctx context.Context,
	input ChangeBatchQCStatusInput,
) (ChangeBatchQCStatusResult, error) {
	actorID := strings.TrimSpace(input.ActorID)
	if actorID == "" {
		return ChangeBatchQCStatusResult{}, ErrBatchTransitionActorRequired
	}
	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		return ChangeBatchQCStatusResult{}, ErrBatchTransitionReasonRequired
	}
	changedAt := input.ChangedAt
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	batch, ok := c.batches[strings.TrimSpace(input.BatchID)]
	if !ok {
		return ChangeBatchQCStatusResult{}, ErrBatchNotFound
	}

	updated, err := batch.ChangeQCStatus(input.NextStatus, changedAt)
	if err != nil {
		return ChangeBatchQCStatusResult{}, err
	}

	businessRef := strings.TrimSpace(input.BusinessRef)
	if businessRef == "" {
		businessRef = batch.ID
	}
	log, err := audit.NewLog(audit.NewLogInput{
		OrgID:      batch.OrgID,
		ActorID:    actorID,
		Action:     batchQCTransitionAction,
		EntityType: batchQCTransitionEntityType,
		EntityID:   batch.ID,
		RequestID:  strings.TrimSpace(input.RequestID),
		BeforeData: map[string]any{
			"qc_status":  string(batch.QCStatus),
			"updated_at": batch.UpdatedAt.UTC().Format(time.RFC3339),
		},
		AfterData: map[string]any{
			"qc_status":  string(updated.QCStatus),
			"updated_at": updated.UpdatedAt.UTC().Format(time.RFC3339),
		},
		Metadata: map[string]any{
			"batch_no":     batch.BatchNo,
			"business_ref": businessRef,
			"reason":       reason,
			"sku":          batch.SKU,
		},
		CreatedAt: changedAt,
	})
	if err != nil {
		return ChangeBatchQCStatusResult{}, err
	}
	if err := c.auditLog.Record(ctx, log); err != nil {
		return ChangeBatchQCStatusResult{}, err
	}

	c.batches[updated.ID] = updated
	transition := batchQCTransitionFromAudit(log)

	return ChangeBatchQCStatusResult{
		Batch:      updated,
		Transition: transition,
		AuditLogID: log.ID,
	}, nil
}

func (c *BatchCatalog) ListQCTransitions(ctx context.Context, batchID string) ([]domain.BatchQCTransition, error) {
	batchID = strings.TrimSpace(batchID)
	c.mu.RLock()
	_, ok := c.batches[batchID]
	c.mu.RUnlock()
	if !ok {
		return nil, ErrBatchNotFound
	}

	logs, err := c.auditLog.List(ctx, audit.Query{
		Action:     batchQCTransitionAction,
		EntityType: batchQCTransitionEntityType,
		EntityID:   batchID,
		Limit:      100,
	})
	if err != nil {
		return nil, err
	}

	transitions := make([]domain.BatchQCTransition, 0, len(logs))
	for _, log := range logs {
		transitions = append(transitions, batchQCTransitionFromAudit(log))
	}

	return transitions, nil
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

func firstBatchAuditStore(stores []audit.LogStore) audit.LogStore {
	for _, store := range stores {
		if store != nil {
			return store
		}
	}

	return audit.NewInMemoryLogStore()
}

func batchQCTransitionFromAudit(log audit.Log) domain.BatchQCTransition {
	return domain.BatchQCTransition{
		ID:           log.ID,
		BatchID:      log.EntityID,
		BatchNo:      stringValue(log.Metadata, "batch_no"),
		SKU:          stringValue(log.Metadata, "sku"),
		FromQCStatus: domain.NormalizeQCStatus(domain.QCStatus(stringValue(log.BeforeData, "qc_status"))),
		ToQCStatus:   domain.NormalizeQCStatus(domain.QCStatus(stringValue(log.AfterData, "qc_status"))),
		ActorID:      log.ActorID,
		Reason:       stringValue(log.Metadata, "reason"),
		BusinessRef:  stringValue(log.Metadata, "business_ref"),
		AuditLogID:   log.ID,
		CreatedAt:    log.CreatedAt,
	}
}

func stringValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}

	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}

	return ""
}

func mustBatch(batch domain.Batch, err error) domain.Batch {
	if err != nil {
		panic(err)
	}

	return batch
}
