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

	items, pagination, err := store.List(context.Background(), domain.NewItemFilter("citric", domain.ItemStatusActive, domain.ItemTypeRawMaterial, 1, 20))
	if err != nil {
		t.Fatalf("list items: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
	if items[0].SKUCode != "ACI_CITRIC" {
		t.Fatalf("sku = %q, want ACI_CITRIC", items[0].SKUCode)
	}
	if pagination.TotalItems != 1 || pagination.Page != 1 {
		t.Fatalf("pagination = %+v, want one item on page 1", pagination)
	}
}

func TestImportedMasterDataItemsAreNormalizedFromSourceSheets(t *testing.T) {
	items := importedMasterDataItems()
	if len(items) != 371 {
		t.Fatalf("imported item count = %d, want 371", len(items))
	}

	bySKU := make(map[string]domain.Item, len(items))
	for _, item := range items {
		if item.SKUCode == "SERUM-30ML" || item.SKUCode == "CREAM-50G" || item.SKUCode == "TONER-100ML" {
			t.Fatalf("legacy mock sku %q should not be part of imported master data", item.SKUCode)
		}
		if _, exists := bySKU[item.SKUCode]; exists {
			t.Fatalf("duplicate imported sku %q", item.SKUCode)
		}
		bySKU[item.SKUCode] = item
	}

	citric := bySKU["ACI_CITRIC"]
	if citric.Name != "CITRIC ACID" || citric.Type != domain.ItemTypeRawMaterial || citric.Group != "acid" || citric.UOMBase != "KG" {
		t.Fatalf("ACI_CITRIC = %+v, want raw acid KG item", citric)
	}
	if !citric.IsPurchasable || citric.IsSellable || citric.IsProducible {
		t.Fatalf("ACI_CITRIC flags = %+v, want purchasable raw material only", citric)
	}

	if fragrance := bySKU["FRA_NTG"]; fragrance.Name == "" || fragrance.Group != "fragrance" || fragrance.UOMBase != "KG" {
		t.Fatalf("FRA_NTG = %+v, want imported fragrance from header row", fragrance)
	}
	if fragrance := bySKU["FRA_SEXY"]; fragrance.Name == "" || fragrance.Group != "fragrance" || fragrance.UOMBase != "KG" {
		t.Fatalf("FRA_SEXY = %+v, want imported fragrance from header row", fragrance)
	}

	tube := bySKU["TP-100"]
	if tube.Type != domain.ItemTypePackaging || tube.Group != "tube" || tube.UOMBase != "TUBE" {
		t.Fatalf("TP-100 = %+v, want packaging tube item", tube)
	}
	retail := bySKU["GRN"]
	if retail.Name != "DẦU GỘI RETRO NANO 350ML" || retail.Type != domain.ItemTypeFinishedGood || retail.Group != "retail_hair" || retail.UOMBase != "BOTTLE" {
		t.Fatalf("GRN = %+v, want active retail hair finished good bottle", retail)
	}
	if !retail.IsSellable || retail.IsPurchasable || !retail.IsProducible {
		t.Fatalf("GRN flags = %+v, want sellable and producible finished good", retail)
	}
	gift := bySKU["BTMN"]
	if gift.Type != domain.ItemTypeFinishedGood || gift.Group != "gift_skin" || gift.UOMBase != "PACK" || gift.IsSellable {
		t.Fatalf("BTMN = %+v, want non-sellable gift finished good pack", gift)
	}
	accessory := bySKU["BONG"]
	if accessory.Name != "BÔNG TẮM VYVY SKINCARE" || accessory.Group != "accessory_skin" || accessory.UOMBase != "PCS" {
		t.Fatalf("BONG = %+v, want accent-normalized accessory SKU", accessory)
	}
	for _, skipped := range []string{"SU500", "BÔNG", "MCBĐ"} {
		if _, ok := bySKU[skipped]; ok {
			t.Fatalf("sku %q should not be part of active-only imported master data", skipped)
		}
	}
	bagFound := false
	packFound := false
	rollFound := false
	for _, item := range items {
		if item.UOMBase == "BAG" {
			bagFound = true
		}
		if item.UOMBase == "PACK" {
			packFound = true
		}
		if item.UOMBase == "ROLL" {
			rollFound = true
		}
	}
	if !bagFound || !packFound || !rollFound {
		t.Fatalf("imported UOMs bag=%v pack=%v roll=%v, want all present", bagFound, packFound, rollFound)
	}
}

func TestPostgresSeedItemsExcludeLegacyOperationalMockItems(t *testing.T) {
	for _, item := range seedMasterDataItems() {
		if item.SKUCode == "SERUM-30ML" || item.SKUCode == "CREAM-50G" || item.SKUCode == "TONER-100ML" {
			t.Fatalf("legacy mock sku %q should not be seeded into Postgres", item.SKUCode)
		}
	}
}

func TestItemCatalogBlocksDuplicateItemAndSKUCode(t *testing.T) {
	store := NewPrototypeItemCatalog(audit.NewInMemoryLogStore())

	_, err := store.Create(context.Background(), CreateItemInput{
		ItemCode:         "ACI_CITRIC",
		SKUCode:          "NEW-SERUM-30ML",
		Name:             "New Serum",
		Type:             "raw_material",
		UOMBase:          "KG",
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
		SKUCode:          "ACI_CITRIC",
		Name:             "New Serum",
		Type:             "raw_material",
		UOMBase:          "KG",
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
