package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestItemCatalogListsFilteredPrototypeItems(t *testing.T) {
	store := NewPrototypeItemCatalog(audit.NewInMemoryLogStore())

	items, pagination, err := store.List(context.Background(), domain.NewItemFilter("serum", domain.ItemStatusActive, domain.ItemTypeFinishedGood, 1, 20))
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if items[0].SKUCode != "SERUM-30ML" {
		t.Fatalf("sku = %q, want SERUM-30ML", items[0].SKUCode)
	}
	if pagination.TotalItems != 1 || pagination.Page != 1 {
		t.Fatalf("pagination = %+v, want one item on page 1", pagination)
	}
}

func TestItemCatalogBlocksDuplicateItemAndSKUCode(t *testing.T) {
	store := NewPrototypeItemCatalog(audit.NewInMemoryLogStore())

	_, err := store.Create(context.Background(), CreateItemInput{
		ItemCode:         "ITEM-SERUM-HYDRA",
		SKUCode:          "NEW-SERUM-30ML",
		Name:             "New Serum",
		Type:             "finished_good",
		UOMBase:          "EA",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    365,
		Status:           "active",
		ActorID:          "user-erp-admin",
	})
	if !errors.Is(err, ErrDuplicateItemCode) {
		t.Fatalf("error = %v, want duplicate item code", err)
	}

	_, err = store.Create(context.Background(), CreateItemInput{
		ItemCode:         "ITEM-NEW-SERUM",
		SKUCode:          "SERUM-30ML",
		Name:             "New Serum",
		Type:             "finished_good",
		UOMBase:          "EA",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    365,
		Status:           "active",
		ActorID:          "user-erp-admin",
	})
	if !errors.Is(err, ErrDuplicateSKUCode) {
		t.Fatalf("error = %v, want duplicate sku code", err)
	}
}

func TestItemCatalogCreatesUpdatesStatusAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypeItemCatalogAt(auditStore, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	ctx := context.Background()

	created, err := store.Create(ctx, CreateItemInput{
		ItemCode:         "ITEM-MASK-SET",
		SKUCode:          "MASK-SET-05",
		Name:             "Sheet Mask Set",
		Type:             "finished_good",
		Group:            "mask",
		BrandCode:        "MYH",
		UOMBase:          "EA",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    540,
		QCRequired:       true,
		Status:           "draft",
		IsSellable:       true,
		IsProducible:     true,
		ActorID:          "user-erp-admin",
		RequestID:        "req-item-create",
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if created.AuditLogID == "" {
		t.Fatal("create audit log id is empty")
	}

	updated, err := store.Update(ctx, UpdateItemInput{
		ID:               created.Item.ID,
		ItemCode:         "ITEM-MASK-SET",
		SKUCode:          "MASK-SET-05",
		Name:             "Sheet Mask Set 5pcs",
		Type:             "finished_good",
		Group:            "mask",
		BrandCode:        "MYH",
		UOMBase:          "EA",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    540,
		QCRequired:       true,
		Status:           "draft",
		IsSellable:       true,
		IsProducible:     true,
		ActorID:          "user-erp-admin",
		RequestID:        "req-item-update",
	})
	if err != nil {
		t.Fatalf("update item: %v", err)
	}
	if updated.Item.Name != "Sheet Mask Set 5pcs" {
		t.Fatalf("name = %q, want updated name", updated.Item.Name)
	}

	statusChanged, err := store.ChangeStatus(ctx, ChangeItemStatusInput{
		ID:        created.Item.ID,
		Status:    "active",
		ActorID:   "user-erp-admin",
		RequestID: "req-item-status",
	})
	if err != nil {
		t.Fatalf("change status: %v", err)
	}
	if statusChanged.Item.Status != domain.ItemStatusActive {
		t.Fatalf("status = %q, want active", statusChanged.Item.Status)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.Item.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit logs = %d, want 3", len(logs))
	}
}
