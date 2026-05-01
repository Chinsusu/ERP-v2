package main

import (
	"testing"

	purchaseapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/purchase/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimePurchaseOrderStoreFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimePurchaseOrderStore(config.Config{}, audit.NewInMemoryLogStore())
	if err != nil {
		t.Fatalf("newRuntimePurchaseOrderStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*purchaseapp.PrototypePurchaseOrderStore); !ok {
		t.Fatalf("store type = %T, want *PrototypePurchaseOrderStore", store)
	}
}

func TestNewRuntimePurchaseOrderStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimePurchaseOrderStore(
		config.Config{
			AppEnv:      "dev",
			DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
		},
		audit.NewInMemoryLogStore(),
	)
	if err != nil {
		t.Fatalf("newRuntimePurchaseOrderStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(purchaseapp.PostgresPurchaseOrderStore); !ok {
		t.Fatalf("store type = %T, want PostgresPurchaseOrderStore", store)
	}
}
