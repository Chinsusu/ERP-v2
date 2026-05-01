package main

import (
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/audit"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeAuditLogStoreFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimeAuditLogStore(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeAuditLogStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*audit.InMemoryLogStore); !ok {
		t.Fatalf("store type = %T, want *audit.InMemoryLogStore", store)
	}
}

func TestNewRuntimeAuditLogStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeAuditLogStore(config.Config{
		AppEnv:      "dev",
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeAuditLogStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(audit.PostgresLogStore); !ok {
		t.Fatalf("store type = %T, want audit.PostgresLogStore", store)
	}
}
