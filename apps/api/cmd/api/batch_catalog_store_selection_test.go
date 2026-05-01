package main

import (
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeBatchCatalogStoreUsesPrototypeFallback(t *testing.T) {
	store, closeStore, err := newRuntimeBatchCatalogStore(config.Config{}, audit.NewInMemoryLogStore())
	if err != nil {
		t.Fatalf("newRuntimeBatchCatalogStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*inventoryapp.BatchCatalog); !ok {
		t.Fatalf("store type = %T, want *BatchCatalog", store)
	}
}

func TestNewRuntimeBatchCatalogStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeBatchCatalogStore(config.Config{
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	}, audit.NewInMemoryLogStore())
	if err != nil {
		t.Fatalf("newRuntimeBatchCatalogStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(inventoryapp.PostgresBatchCatalogStore); !ok {
		t.Fatalf("store type = %T, want PostgresBatchCatalogStore", store)
	}
}
