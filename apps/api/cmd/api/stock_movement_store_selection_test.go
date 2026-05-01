package main

import (
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeStockMovementStoreFallsBackToMemoryWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimeStockMovementStore(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeStockMovementStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for memory store")
	}
	if _, ok := store.(*inventoryapp.InMemoryStockMovementStore); !ok {
		t.Fatalf("store type = %T, want *InMemoryStockMovementStore", store)
	}
}

func TestNewRuntimeStockMovementStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeStockMovementStore(config.Config{
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeStockMovementStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(inventoryapp.PostgresStockMovementStore); !ok {
		t.Fatalf("store type = %T, want PostgresStockMovementStore", store)
	}
}
