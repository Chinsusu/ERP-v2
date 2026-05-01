package main

import (
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeStockAvailabilityStoreFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimeStockAvailabilityStore(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeStockAvailabilityStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(inventoryapp.PrototypeStockAvailabilityStore); !ok {
		t.Fatalf("store type = %T, want PrototypeStockAvailabilityStore", store)
	}
}

func TestNewRuntimeStockAvailabilityStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeStockAvailabilityStore(config.Config{
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeStockAvailabilityStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(inventoryapp.PostgresStockAvailabilityStore); !ok {
		t.Fatalf("store type = %T, want PostgresStockAvailabilityStore", store)
	}
}
