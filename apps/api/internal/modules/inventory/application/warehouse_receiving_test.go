package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestWarehouseReceivingPostCreatesStockMovementAndAudit(t *testing.T) {
	service, movementStore, auditStore := newTestWarehouseReceivingService()

	result, err := service.PostWarehouseReceiving(context.Background(), WarehouseReceivingTransitionInput{
		ID:        "grn-hcm-260427-inspect",
		ActorID:   "user-warehouse-lead",
		RequestID: "req-receive-post",
	})
	if err != nil {
		t.Fatalf("post receiving: %v", err)
	}
	if result.Receipt.Status != domain.WarehouseReceivingStatusPosted {
		t.Fatalf("status = %q, want posted", result.Receipt.Status)
	}
	if movementStore.Count() != 1 || len(result.Receipt.StockMovements) != 1 {
		t.Fatalf("stock movements = store %d response %d, want 1", movementStore.Count(), len(result.Receipt.StockMovements))
	}
	movement := result.Receipt.StockMovements[0]
	if movement.SourceDocType != receivingSourceDocType ||
		movement.SourceDocID != "grn-hcm-260427-inspect" ||
		movement.StockStatus != domain.StockStatusAvailable {
		t.Fatalf("movement = %+v, want available purchase receipt linked to receiving", movement)
	}
	if result.AuditLogID == "" {
		t.Fatal("audit log id is empty")
	}
	logs, err := auditStore.List(context.Background(), audit.Query{Action: "inventory.receiving.posted"})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 1 || logs[0].AfterData["stock_movement_count"] != 1 {
		t.Fatalf("audit logs = %+v, want posted movement count", logs)
	}
}

func TestWarehouseReceivingRejectsDuplicatePost(t *testing.T) {
	service, movementStore, _ := newTestWarehouseReceivingService()

	_, err := service.PostWarehouseReceiving(context.Background(), WarehouseReceivingTransitionInput{
		ID:      "grn-hcm-260427-inspect",
		ActorID: "user-warehouse-lead",
	})
	if err != nil {
		t.Fatalf("first post receiving: %v", err)
	}
	_, err = service.PostWarehouseReceiving(context.Background(), WarehouseReceivingTransitionInput{
		ID:      "grn-hcm-260427-inspect",
		ActorID: "user-warehouse-lead",
	})
	if !errors.Is(err, domain.ErrReceivingAlreadyPosted) {
		t.Fatalf("err = %v, want already posted", err)
	}
	if movementStore.Count() != 1 {
		t.Fatalf("movement count = %d, want unchanged 1", movementStore.Count())
	}
}

func TestWarehouseReceivingRejectsInvalidLocation(t *testing.T) {
	service, _, _ := newTestWarehouseReceivingService()

	_, err := service.CreateWarehouseReceiving(context.Background(), CreateWarehouseReceivingInput{
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-rm-recv-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "PO-260427-0100",
		Lines: []CreateWarehouseReceivingLineInput{
			{
				ItemID:      "item-cream-50g",
				SKU:         "CREAM-50G",
				BatchID:     "batch-cream-2603b",
				Quantity:    "12",
				BaseUOMCode: "EA",
			},
		},
		ActorID: "user-warehouse-lead",
	})
	if !errors.Is(err, ErrReceivingInvalidLocation) {
		t.Fatalf("err = %v, want invalid location", err)
	}
}

func TestWarehouseReceivingPostRequiresBatchAndQCData(t *testing.T) {
	service, _, _ := newTestWarehouseReceivingService()
	ctx := context.Background()

	created, err := service.CreateWarehouseReceiving(ctx, CreateWarehouseReceivingInput{
		ID:               "grn-missing-batch-qc",
		ReceiptNo:        "GRN-260427-0999",
		WarehouseID:      "wh-hcm-fg",
		LocationID:       "loc-hcm-fg-recv-01",
		ReferenceDocType: "purchase_order",
		ReferenceDocID:   "PO-260427-0999",
		Lines: []CreateWarehouseReceivingLineInput{
			{
				ID:          "line-missing-batch-qc",
				ItemID:      "item-cream-50g",
				SKU:         "CREAM-50G",
				Quantity:    "12",
				BaseUOMCode: "EA",
			},
		},
		ActorID: "user-warehouse-lead",
	})
	if err != nil {
		t.Fatalf("create receiving: %v", err)
	}
	submitted, err := service.SubmitWarehouseReceiving(ctx, WarehouseReceivingTransitionInput{
		ID:      created.Receipt.ID,
		ActorID: "user-warehouse-lead",
	})
	if err != nil {
		t.Fatalf("submit receiving: %v", err)
	}
	_, err = service.MarkWarehouseReceivingInspectReady(ctx, WarehouseReceivingTransitionInput{
		ID:      submitted.Receipt.ID,
		ActorID: "user-qa",
	})
	if err != nil {
		t.Fatalf("mark inspect ready: %v", err)
	}
	_, err = service.PostWarehouseReceiving(ctx, WarehouseReceivingTransitionInput{
		ID:      created.Receipt.ID,
		ActorID: "user-warehouse-lead",
	})
	if !errors.Is(err, domain.ErrReceivingMissingBatchQCData) {
		t.Fatalf("err = %v, want missing batch/qc data", err)
	}
}

func newTestWarehouseReceivingService() (WarehouseReceivingService, *InMemoryStockMovementStore, *audit.InMemoryLogStore) {
	auditStore := audit.NewInMemoryLogStore()
	movementStore := NewInMemoryStockMovementStore()
	service := NewWarehouseReceivingService(
		NewPrototypeWarehouseReceivingStore(),
		masterdataapp.NewPrototypeWarehouseLocationCatalog(auditStore),
		NewPrototypeBatchCatalog(auditStore),
		movementStore,
		auditStore,
	)
	service.clock = func() time.Time {
		return time.Date(2026, 4, 27, 11, 0, 0, 0, time.UTC)
	}

	return service, movementStore, auditStore
}
