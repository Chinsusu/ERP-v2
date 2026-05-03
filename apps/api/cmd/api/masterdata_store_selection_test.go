package main

import (
	"testing"

	masterdataapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/masterdata/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeMasterDataStoresFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	stores, closeStores, err := newRuntimeMasterDataStores(config.Config{}, audit.NewInMemoryLogStore())
	if err != nil {
		t.Fatalf("newRuntimeMasterDataStores() error = %v", err)
	}
	if closeStores != nil {
		t.Fatal("closeStores is not nil, want nil for prototype stores")
	}

	if _, ok := stores.items.(*masterdataapp.ItemCatalog); !ok {
		t.Fatalf("items store type = %T, want *ItemCatalog", stores.items)
	}
	if _, ok := stores.formulas.(*masterdataapp.FormulaCatalog); !ok {
		t.Fatalf("formulas store type = %T, want *FormulaCatalog", stores.formulas)
	}
	if _, ok := stores.uoms.(*masterdataapp.UOMCatalog); !ok {
		t.Fatalf("uom store type = %T, want *UOMCatalog", stores.uoms)
	}
	if _, ok := stores.warehouses.(*masterdataapp.WarehouseLocationCatalog); !ok {
		t.Fatalf("warehouse store type = %T, want *WarehouseLocationCatalog", stores.warehouses)
	}
	if _, ok := stores.parties.(*masterdataapp.PartyCatalog); !ok {
		t.Fatalf("party store type = %T, want *PartyCatalog", stores.parties)
	}
}

func TestNewRuntimeMasterDataStoresUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	stores, closeStores, err := newRuntimeMasterDataStores(
		config.Config{
			AppEnv:      "dev",
			DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
		},
		audit.NewInMemoryLogStore(),
	)
	if err != nil {
		t.Fatalf("newRuntimeMasterDataStores() error = %v", err)
	}
	if closeStores == nil {
		t.Fatal("closeStores = nil, want database close function")
	}
	defer func() {
		if err := closeStores(); err != nil {
			t.Fatalf("closeStores() error = %v", err)
		}
	}()

	if _, ok := stores.items.(*masterdataapp.PostgresItemCatalog); !ok {
		t.Fatalf("items store type = %T, want *PostgresItemCatalog", stores.items)
	}
	if _, ok := stores.formulas.(*masterdataapp.PostgresFormulaCatalog); !ok {
		t.Fatalf("formulas store type = %T, want *PostgresFormulaCatalog", stores.formulas)
	}
	if _, ok := stores.uoms.(*masterdataapp.PostgresUOMCatalog); !ok {
		t.Fatalf("uom store type = %T, want *PostgresUOMCatalog", stores.uoms)
	}
	if _, ok := stores.warehouses.(*masterdataapp.PostgresWarehouseLocationCatalog); !ok {
		t.Fatalf("warehouse store type = %T, want *PostgresWarehouseLocationCatalog", stores.warehouses)
	}
	if _, ok := stores.parties.(*masterdataapp.PostgresPartyCatalog); !ok {
		t.Fatalf("party store type = %T, want *PostgresPartyCatalog", stores.parties)
	}
}
