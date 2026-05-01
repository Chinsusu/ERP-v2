package main

import (
	"testing"

	salesapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/sales/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeSalesOrderStoreFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimeSalesOrderStore(config.Config{}, audit.NewInMemoryLogStore())
	if err != nil {
		t.Fatalf("newRuntimeSalesOrderStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*salesapp.PrototypeSalesOrderStore); !ok {
		t.Fatalf("store type = %T, want *PrototypeSalesOrderStore", store)
	}
}

func TestNewRuntimeSalesOrderStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeSalesOrderStore(
		config.Config{DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable"},
		audit.NewInMemoryLogStore(),
	)
	if err != nil {
		t.Fatalf("newRuntimeSalesOrderStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(salesapp.PostgresSalesOrderStore); !ok {
		t.Fatalf("store type = %T, want PostgresSalesOrderStore", store)
	}
}
