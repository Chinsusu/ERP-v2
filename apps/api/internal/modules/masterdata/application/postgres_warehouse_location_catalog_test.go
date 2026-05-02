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

	_ "github.com/jackc/pgx/v5/stdlib"
)

const testPostgresWarehouseCatalogOrgID = "00000000-0000-4000-8000-000000170301"

func TestPostgresWarehouseLocationCatalogRequiresDatabase(t *testing.T) {
	store := NewPostgresWarehouseLocationCatalog(nil, nil, PostgresWarehouseLocationCatalogConfig{})

	if _, _, err := store.ListWarehouses(context.Background(), domain.WarehouseFilter{}); err == nil {
		t.Fatal("ListWarehouses() error = nil, want database required error")
	}
	if _, err := store.GetWarehouse(context.Background(), "wh-s17"); err == nil {
		t.Fatal("GetWarehouse() error = nil, want database required error")
	}
	if _, err := store.CreateWarehouse(context.Background(), CreateWarehouseInput{}); err == nil {
		t.Fatal("CreateWarehouse() error = nil, want database required error")
	}
	if _, _, err := store.ListLocations(context.Background(), domain.LocationFilter{}); err == nil {
		t.Fatal("ListLocations() error = nil, want database required error")
	}
	if _, err := store.GetLocation(context.Background(), "loc-s17"); err == nil {
		t.Fatal("GetLocation() error = nil, want database required error")
	}
	if _, err := store.CreateLocation(context.Background(), CreateLocationInput{}); err == nil {
		t.Fatal("CreateLocation() error = nil, want database required error")
	}
}

func TestPostgresWarehouseLocationCatalogPersistsLifecycleAndReload(t *testing.T) {
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

	if err := seedPostgresWarehouseCatalogFixture(ctx, db); err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	auditStore := audit.NewPostgresLogStore(db, audit.PostgresLogStoreConfig{DefaultOrgID: testPostgresWarehouseCatalogOrgID})
	store := NewPostgresWarehouseLocationCatalog(db, auditStore, PostgresWarehouseLocationCatalogConfig{
		DefaultOrgID: testPostgresWarehouseCatalogOrgID,
		Clock:        fixedPostgresWarehouseClock(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)),
	})

	createdWarehouse, err := store.CreateWarehouse(ctx, CreateWarehouseInput{
		Code:            "WH-S17-FG",
		Name:            "S17 Finished Goods",
		Type:            "finished_good",
		SiteCode:        "HCM",
		Address:         "S17 DC",
		AllowSaleIssue:  true,
		AllowProdIssue:  false,
		AllowQuarantine: false,
		Status:          "active",
		ActorID:         "user-erp-admin",
		RequestID:       "req-s17-warehouse-create",
	})
	if err != nil {
		t.Fatalf("CreateWarehouse() error = %v", err)
	}
	if createdWarehouse.AuditLogID == "" {
		t.Fatal("warehouse create audit log id is empty")
	}
	if _, err := store.CreateWarehouse(ctx, CreateWarehouseInput{
		Code:           "WH-S17-FG",
		Name:           "Duplicate",
		Type:           "finished_good",
		SiteCode:       "HCM",
		AllowSaleIssue: true,
		Status:         "active",
		ActorID:        "user-erp-admin",
	}); !errors.Is(err, ErrDuplicateWarehouseCode) {
		t.Fatalf("duplicate warehouse err = %v, want ErrDuplicateWarehouseCode", err)
	}

	updatedWarehouse, err := store.UpdateWarehouse(ctx, UpdateWarehouseInput{
		ID:              createdWarehouse.Warehouse.ID,
		Code:            "WH-S17-FG",
		Name:            "S17 Finished Goods Updated",
		Type:            "finished_good",
		SiteCode:        "HCM",
		Address:         "S17 DC Updated",
		AllowSaleIssue:  true,
		AllowProdIssue:  false,
		AllowQuarantine: false,
		Status:          "active",
		ActorID:         "user-erp-admin",
		RequestID:       "req-s17-warehouse-update",
	})
	if err != nil {
		t.Fatalf("UpdateWarehouse() error = %v", err)
	}
	if updatedWarehouse.Warehouse.Name != "S17 Finished Goods Updated" {
		t.Fatalf("updated warehouse = %+v", updatedWarehouse.Warehouse)
	}

	createdLocation, err := store.CreateLocation(ctx, CreateLocationInput{
		WarehouseID:  createdWarehouse.Warehouse.ID,
		Code:         "S17-PICK-01",
		Name:         "S17 Pick 01",
		Type:         "pick",
		ZoneCode:     "PICK",
		AllowReceive: false,
		AllowPick:    true,
		AllowStore:   true,
		IsDefault:    true,
		Status:       "active",
		ActorID:      "user-erp-admin",
		RequestID:    "req-s17-location-create",
	})
	if err != nil {
		t.Fatalf("CreateLocation() error = %v", err)
	}
	if createdLocation.Location.WarehouseCode != "WH-S17-FG" {
		t.Fatalf("location warehouse code = %q, want WH-S17-FG", createdLocation.Location.WarehouseCode)
	}
	if _, err := store.CreateLocation(ctx, CreateLocationInput{
		WarehouseID: createdWarehouse.Warehouse.ID,
		Code:        "S17-PICK-01",
		Name:        "Duplicate Pick",
		Type:        "pick",
		ZoneCode:    "PICK",
		AllowPick:   true,
		AllowStore:  true,
		Status:      "active",
		ActorID:     "user-erp-admin",
	}); !errors.Is(err, ErrDuplicateLocationCode) {
		t.Fatalf("duplicate location err = %v, want ErrDuplicateLocationCode", err)
	}
	if _, err := store.CreateLocation(ctx, CreateLocationInput{
		WarehouseID: "missing",
		Code:        "S17-MISSING",
		Name:        "Missing Warehouse",
		Type:        "pick",
		Status:      "active",
		ActorID:     "user-erp-admin",
	}); !errors.Is(err, ErrInvalidLocationWarehouse) {
		t.Fatalf("invalid warehouse err = %v, want ErrInvalidLocationWarehouse", err)
	}

	inactiveLocation, err := store.ChangeLocationStatus(ctx, ChangeLocationStatusInput{
		ID:        createdLocation.Location.ID,
		Status:    "inactive",
		ActorID:   "user-erp-admin",
		RequestID: "req-s17-location-status",
	})
	if err != nil {
		t.Fatalf("ChangeLocationStatus() error = %v", err)
	}
	if inactiveLocation.Location.Status != domain.LocationStatusInactive {
		t.Fatalf("location status = %s, want inactive", inactiveLocation.Location.Status)
	}
	if _, err := store.UpdateLocation(ctx, UpdateLocationInput{
		ID:          createdLocation.Location.ID,
		WarehouseID: createdWarehouse.Warehouse.ID,
		Code:        "S17-PICK-01",
		Name:        "Inactive Edit",
		Type:        "pick",
		ZoneCode:    "PICK",
		AllowPick:   true,
		AllowStore:  true,
		Status:      "inactive",
		ActorID:     "user-erp-admin",
		RequestID:   "req-s17-location-inactive-edit",
	}); !errors.Is(err, ErrInactiveLocation) {
		t.Fatalf("inactive edit err = %v, want ErrInactiveLocation", err)
	}

	statusChangedWarehouse, err := store.ChangeWarehouseStatus(ctx, ChangeWarehouseStatusInput{
		ID:        createdWarehouse.Warehouse.ID,
		Status:    "inactive",
		ActorID:   "user-erp-admin",
		RequestID: "req-s17-warehouse-status",
	})
	if err != nil {
		t.Fatalf("ChangeWarehouseStatus() error = %v", err)
	}
	if statusChangedWarehouse.Warehouse.Status != domain.WarehouseStatusInactive {
		t.Fatalf("warehouse status = %s, want inactive", statusChangedWarehouse.Warehouse.Status)
	}

	reloadedStore := NewPostgresWarehouseLocationCatalog(db, auditStore, PostgresWarehouseLocationCatalogConfig{DefaultOrgID: testPostgresWarehouseCatalogOrgID})
	reloadedWarehouse, err := reloadedStore.GetWarehouse(ctx, createdWarehouse.Warehouse.ID)
	if err != nil {
		t.Fatalf("GetWarehouse() reload error = %v", err)
	}
	if reloadedWarehouse.Name != "S17 Finished Goods Updated" || reloadedWarehouse.Status != domain.WarehouseStatusInactive {
		t.Fatalf("reloaded warehouse = %+v", reloadedWarehouse)
	}
	reloadedLocation, err := reloadedStore.GetLocation(ctx, createdLocation.Location.ID)
	if err != nil {
		t.Fatalf("GetLocation() reload error = %v", err)
	}
	if reloadedLocation.Code != "S17-PICK-01" || reloadedLocation.Status != domain.LocationStatusInactive {
		t.Fatalf("reloaded location = %+v", reloadedLocation)
	}
	warehouses, pagination, err := reloadedStore.ListWarehouses(ctx, domain.NewWarehouseFilter("S17", domain.WarehouseStatusInactive, domain.WarehouseTypeFinishedGood, 1, 20))
	if err != nil {
		t.Fatalf("ListWarehouses() error = %v", err)
	}
	if pagination.TotalItems == 0 || !containsPostgresWarehouse(warehouses, createdWarehouse.Warehouse.ID) {
		t.Fatalf("warehouses = %+v pagination = %+v, missing created warehouse", warehouses, pagination)
	}
	locations, pagination, err := reloadedStore.ListLocations(ctx, domain.NewLocationFilter("S17", createdWarehouse.Warehouse.ID, domain.LocationStatusInactive, domain.LocationTypePick, 1, 20))
	if err != nil {
		t.Fatalf("ListLocations() error = %v", err)
	}
	if pagination.TotalItems == 0 || !containsPostgresLocation(locations, createdLocation.Location.ID) {
		t.Fatalf("locations = %+v pagination = %+v, missing created location", locations, pagination)
	}

	warehouseLogs, err := auditStore.List(ctx, audit.Query{EntityID: createdWarehouse.Warehouse.ID})
	if err != nil {
		t.Fatalf("list warehouse audit logs: %v", err)
	}
	if len(warehouseLogs) < 3 {
		t.Fatalf("warehouse audit logs = %d, want at least 3", len(warehouseLogs))
	}
	locationLogs, err := auditStore.List(ctx, audit.Query{EntityID: createdLocation.Location.ID})
	if err != nil {
		t.Fatalf("list location audit logs: %v", err)
	}
	if len(locationLogs) < 2 {
		t.Fatalf("location audit logs = %d, want at least 2", len(locationLogs))
	}
}

func seedPostgresWarehouseCatalogFixture(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO core.organizations (id, code, name, status)
VALUES ($1::uuid, 'S17_WAREHOUSE_ORG', 'S17 Warehouse Catalog Test Org', 'active')
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    status = EXCLUDED.status,
    updated_at = now()`,
		testPostgresWarehouseCatalogOrgID,
	)

	return err
}

func fixedPostgresWarehouseClock(base time.Time) func() time.Time {
	current := base

	return func() time.Time {
		current = current.Add(time.Minute)
		return current
	}
}

func containsPostgresWarehouse(warehouses []domain.Warehouse, id string) bool {
	for _, warehouse := range warehouses {
		if warehouse.ID == id {
			return true
		}
	}

	return false
}

func containsPostgresLocation(locations []domain.Location, id string) bool {
	for _, location := range locations {
		if location.ID == id {
			return true
		}
	}

	return false
}
