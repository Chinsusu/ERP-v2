package application

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/decimal"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testPostgresItemCatalogOrgID = "00000000-0000-4000-8000-000000170201"

func TestPostgresItemCatalogRequiresDatabase(t *testing.T) {
	store := NewPostgresItemCatalog(nil, nil, PostgresItemCatalogConfig{})

	if _, _, err := store.List(context.Background(), domain.ItemFilter{}); err == nil {
		t.Fatal("List() error = nil, want database required error")
	}
	if _, err := store.Get(context.Background(), "item-s17"); err == nil {
		t.Fatal("Get() error = nil, want database required error")
	}
	if _, err := store.Create(context.Background(), CreateItemInput{}); err == nil {
		t.Fatal("Create() error = nil, want database required error")
	}
	if _, err := store.Update(context.Background(), UpdateItemInput{}); err == nil {
		t.Fatal("Update() error = nil, want database required error")
	}
	if _, err := store.ChangeStatus(context.Background(), ChangeItemStatusInput{}); err == nil {
		t.Fatal("ChangeStatus() error = nil, want database required error")
	}
}

func TestPostgresItemCatalogPersistsLifecycleAndReload(t *testing.T) {
	databaseURL := os.Getenv("ERP_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("ERP_TEST_DATABASE_URL is not set")
	}

	ctx := context.Background()
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := seedPostgresItemCatalogFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}

	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: testPostgresItemCatalogOrgID})
	clock := fixedPostgresItemClock(time.Date(2026, 5, 2, 9, 0, 0, 0, time.UTC))
	store := NewPostgresItemCatalog(db, auditStore, PostgresItemCatalogConfig{
		DefaultOrgID: testPostgresItemCatalogOrgID,
		Clock:        clock,
	})

	created, err := store.Create(ctx, CreateItemInput{
		ItemCode:         "ITEM-S17-SERUM",
		SKUCode:          "S17-SERUM-30ML",
		Name:             "S17 Serum 30ml",
		Type:             "finished_good",
		Group:            "serum",
		BrandCode:        "MYH",
		UOMBase:          "PCS",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    730,
		QCRequired:       true,
		Status:           "draft",
		StandardCost:     decimal.MustUnitCost("64000"),
		IsSellable:       true,
		IsProducible:     true,
		ActorID:          "user-erp-admin",
		RequestID:        "req-s17-item-create",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.AuditLogID == "" {
		t.Fatal("create audit log id is empty")
	}

	if _, err := store.Create(ctx, CreateItemInput{
		ItemCode:         "ITEM-S17-SERUM",
		SKUCode:          "S17-SERUM-DUP",
		Name:             "Duplicate Item Code",
		Type:             "finished_good",
		UOMBase:          "PCS",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    365,
		Status:           "active",
		ActorID:          "user-erp-admin",
	}); !errors.Is(err, ErrDuplicateItemCode) {
		t.Fatalf("duplicate item code err = %v, want ErrDuplicateItemCode", err)
	}
	if _, err := store.Create(ctx, CreateItemInput{
		ItemCode:         "ITEM-S17-DUP-SKU",
		SKUCode:          "S17-SERUM-30ML",
		Name:             "Duplicate SKU",
		Type:             "finished_good",
		UOMBase:          "PCS",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    365,
		Status:           "active",
		ActorID:          "user-erp-admin",
	}); !errors.Is(err, ErrDuplicateSKUCode) {
		t.Fatalf("duplicate sku err = %v, want ErrDuplicateSKUCode", err)
	}

	updated, err := store.Update(ctx, UpdateItemInput{
		ID:               created.Item.ID,
		ItemCode:         "ITEM-S17-SERUM",
		SKUCode:          "S17-SERUM-30ML",
		Name:             "S17 Serum 30ml Updated",
		Type:             "finished_good",
		Group:            "serum",
		BrandCode:        "MYH",
		UOMBase:          "PCS",
		LotControlled:    true,
		ExpiryControlled: true,
		ShelfLifeDays:    730,
		QCRequired:       true,
		Status:           "draft",
		StandardCost:     decimal.MustUnitCost("65000"),
		IsSellable:       true,
		IsProducible:     true,
		ActorID:          "user-erp-admin",
		RequestID:        "req-s17-item-update",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Item.Name != "S17 Serum 30ml Updated" || updated.Item.StandardCost.String() != "65000.000000" {
		t.Fatalf("updated item = %+v", updated.Item)
	}

	statusChanged, err := store.ChangeStatus(ctx, ChangeItemStatusInput{
		ID:        created.Item.ID,
		Status:    "active",
		ActorID:   "user-erp-admin",
		RequestID: "req-s17-item-status",
	})
	if err != nil {
		t.Fatalf("ChangeStatus() error = %v", err)
	}
	if statusChanged.Item.Status != domain.ItemStatusActive {
		t.Fatalf("status = %s, want active", statusChanged.Item.Status)
	}

	reloadedStore := NewPostgresItemCatalog(db, auditStore, PostgresItemCatalogConfig{DefaultOrgID: testPostgresItemCatalogOrgID})
	reloaded, err := reloadedStore.Get(ctx, created.Item.ID)
	if err != nil {
		t.Fatalf("Get() after reload error = %v", err)
	}
	if reloaded.Name != "S17 Serum 30ml Updated" ||
		reloaded.Status != domain.ItemStatusActive ||
		reloaded.StandardCost.String() != "65000.000000" {
		t.Fatalf("reloaded item = %+v", reloaded)
	}

	items, pagination, err := reloadedStore.List(ctx, domain.NewItemFilter("serum", domain.ItemStatusActive, domain.ItemTypeFinishedGood, 1, 20))
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if pagination.TotalItems == 0 || !containsPostgresItem(items, created.Item.ID) {
		t.Fatalf("list items = %+v pagination = %+v, missing created item", items, pagination)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.Item.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) < 3 {
		t.Fatalf("audit logs = %d, want at least 3", len(logs))
	}
}

func seedPostgresItemCatalogFixture(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S17_ITEM_ORG', 'S17 Item Catalog Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresItemCatalogOrgID,
	)

	return err
}

func fixedPostgresItemClock(base time.Time) func() time.Time {
	current := base

	return func() time.Time {
		current = current.Add(time.Minute)
		return current
	}
}

func containsPostgresItem(items []domain.Item, id string) bool {
	for _, item := range items {
		if item.ID == id {
			return true
		}
	}

	return false
}
