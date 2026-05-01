package main

import (
	"testing"

	returnsapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/returns/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeReturnReceiptStoreFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimeReturnReceiptStore(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeReturnReceiptStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*returnsapp.PrototypeReturnReceiptStore); !ok {
		t.Fatalf("store type = %T, want *PrototypeReturnReceiptStore", store)
	}
}

func TestNewRuntimeReturnReceiptStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeReturnReceiptStore(config.Config{
		AppEnv:      "dev",
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeReturnReceiptStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(returnsapp.PostgresReturnReceiptStore); !ok {
		t.Fatalf("store type = %T, want PostgresReturnReceiptStore", store)
	}
}
