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
