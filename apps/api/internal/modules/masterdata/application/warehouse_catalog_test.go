package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/domain"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
)

func TestWarehouseLocationCatalogListsFilteredPrototypeWarehouses(t *testing.T) {
	store := NewPrototypeWarehouseLocationCatalog(audit.NewInMemoryLogStore())

	warehouses, pagination, err := store.ListWarehouses(context.Background(), domain.NewWarehouseFilter("hcm", domain.WarehouseStatusActive, domain.WarehouseTypeFinishedGood, 1, 20))
	if err != nil {
		t.Fatalf("list warehouses: %v", err)
	}

	if len(warehouses) != 1 {
		t.Fatalf("warehouses = %d, want 1", len(warehouses))
	}
	if warehouses[0].Code != "WH-HCM-FG" {
		t.Fatalf("warehouse = %q, want WH-HCM-FG", warehouses[0].Code)
	}
	if pagination.TotalItems != 1 || pagination.Page != 1 {
		t.Fatalf("pagination = %+v, want one warehouse on page 1", pagination)
	}
}

func TestWarehouseLocationCatalogBlocksDuplicateWarehouseCode(t *testing.T) {
	store := NewPrototypeWarehouseLocationCatalog(audit.NewInMemoryLogStore())

	_, err := store.CreateWarehouse(context.Background(), CreateWarehouseInput{
		Code:            "WH-HCM-FG",
		Name:            "Duplicate Finished Goods",
		Type:            "finished_good",
		SiteCode:        "HCM",
		AllowSaleIssue:  true,
		AllowProdIssue:  false,
		AllowQuarantine: false,
		Status:          "active",
		ActorID:         "user-erp-admin",
	})
	if !errors.Is(err, ErrDuplicateWarehouseCode) {
		t.Fatalf("error = %v, want duplicate warehouse code", err)
	}
}

func TestWarehouseLocationCatalogCreatesUpdatesStatusAndWritesAudit(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypeWarehouseLocationCatalogAt(auditStore, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	ctx := context.Background()

	created, err := store.CreateWarehouse(ctx, CreateWarehouseInput{
		Code:            "WH-HN-FG",
		Name:            "Finished Goods Warehouse HN",
		Type:            "finished_good",
		SiteCode:        "HN",
		Address:         "Ha Noi distribution center",
		AllowSaleIssue:  true,
		AllowProdIssue:  false,
		AllowQuarantine: false,
		Status:          "active",
		ActorID:         "user-erp-admin",
		RequestID:       "req-warehouse-create",
	})
	if err != nil {
		t.Fatalf("create warehouse: %v", err)
	}
	if created.AuditLogID == "" {
		t.Fatal("create audit log id is empty")
	}

	updated, err := store.UpdateWarehouse(ctx, UpdateWarehouseInput{
		ID:              created.Warehouse.ID,
		Code:            "WH-HN-FG",
		Name:            "Finished Goods Warehouse Ha Noi",
		Type:            "finished_good",
		SiteCode:        "HN",
		Address:         "Ha Noi DC",
		AllowSaleIssue:  true,
		AllowProdIssue:  false,
		AllowQuarantine: false,
		Status:          "active",
		ActorID:         "user-erp-admin",
		RequestID:       "req-warehouse-update",
	})
	if err != nil {
		t.Fatalf("update warehouse: %v", err)
	}
	if updated.Warehouse.Name != "Finished Goods Warehouse Ha Noi" {
		t.Fatalf("warehouse name = %q, want updated name", updated.Warehouse.Name)
	}

	statusChanged, err := store.ChangeWarehouseStatus(ctx, ChangeWarehouseStatusInput{
		ID:        created.Warehouse.ID,
		Status:    "inactive",
		ActorID:   "user-erp-admin",
		RequestID: "req-warehouse-status",
	})
	if err != nil {
		t.Fatalf("change status: %v", err)
	}
	if statusChanged.Warehouse.Status != domain.WarehouseStatusInactive {
		t.Fatalf("status = %q, want inactive", statusChanged.Warehouse.Status)
	}

	logs, err := auditStore.List(ctx, audit.Query{EntityID: created.Warehouse.ID})
	if err != nil {
		t.Fatalf("list audit logs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("audit logs = %d, want 3", len(logs))
	}
}

func TestWarehouseLocationCatalogBlocksInvalidWarehouseAndDuplicateLocation(t *testing.T) {
	store := NewPrototypeWarehouseLocationCatalog(audit.NewInMemoryLogStore())

	_, err := store.CreateLocation(context.Background(), CreateLocationInput{
		WarehouseID: "missing-warehouse",
		Code:        "FG-PICK-A01",
		Name:        "Duplicate Pick A01",
		Type:        "pick",
		ZoneCode:    "PICK",
		AllowPick:   true,
		AllowStore:  true,
		Status:      "active",
		ActorID:     "user-erp-admin",
	})
	if !errors.Is(err, ErrInvalidLocationWarehouse) {
		t.Fatalf("error = %v, want invalid warehouse", err)
	}

	_, err = store.CreateLocation(context.Background(), CreateLocationInput{
		WarehouseID: "wh-hcm-fg",
		Code:        "FG-PICK-A01",
		Name:        "Duplicate Pick A01",
		Type:        "pick",
		ZoneCode:    "PICK",
		AllowPick:   true,
		AllowStore:  true,
		Status:      "active",
		ActorID:     "user-erp-admin",
	})
	if !errors.Is(err, ErrDuplicateLocationCode) {
		t.Fatalf("error = %v, want duplicate location code", err)
	}
}

func TestWarehouseLocationCatalogHandlesInactiveLocationBehavior(t *testing.T) {
	auditStore := audit.NewInMemoryLogStore()
	store := NewPrototypeWarehouseLocationCatalogAt(auditStore, time.Date(2026, 4, 27, 9, 0, 0, 0, time.UTC))
	ctx := context.Background()

	created, err := store.CreateLocation(ctx, CreateLocationInput{
		WarehouseID:  "wh-hcm-fg",
		Code:         "FG-PACK-02",
		Name:         "Packing Bay 02",
		Type:         "pack",
		ZoneCode:     "PACK",
		AllowReceive: false,
		AllowPick:    false,
		AllowStore:   true,
		Status:       "active",
		ActorID:      "user-erp-admin",
		RequestID:    "req-location-create",
	})
	if err != nil {
		t.Fatalf("create location: %v", err)
	}

	inactive, err := store.ChangeLocationStatus(ctx, ChangeLocationStatusInput{
		ID:        created.Location.ID,
		Status:    "inactive",
		ActorID:   "user-erp-admin",
		RequestID: "req-location-status",
	})
	if err != nil {
		t.Fatalf("change location status: %v", err)
	}
	if inactive.Location.Status != domain.LocationStatusInactive {
		t.Fatalf("status = %q, want inactive", inactive.Location.Status)
	}

	activeRows, _, err := store.ListLocations(ctx, domain.NewLocationFilter("", "wh-hcm-fg", domain.LocationStatusActive, "", 1, 100))
	if err != nil {
		t.Fatalf("list active locations: %v", err)
	}
	for _, location := range activeRows {
		if location.ID == created.Location.ID {
			t.Fatalf("inactive location %q was returned in active location list", location.ID)
		}
	}

	_, err = store.UpdateLocation(ctx, UpdateLocationInput{
		ID:           created.Location.ID,
		WarehouseID:  "wh-hcm-fg",
		Code:         "FG-PACK-02",
		Name:         "Packing Bay 02 edited while inactive",
		Type:         "pack",
		ZoneCode:     "PACK",
		AllowReceive: false,
		AllowPick:    false,
		AllowStore:   true,
		Status:       "inactive",
		ActorID:      "user-erp-admin",
	})
	if !errors.Is(err, ErrInactiveLocation) {
		t.Fatalf("error = %v, want inactive location guard", err)
	}
}
