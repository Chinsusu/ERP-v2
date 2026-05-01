package main

import (
	"testing"

	qcapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/qc/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeInboundQCInspectionStoreFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	store, closeStore, err := newRuntimeInboundQCInspectionStore(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeInboundQCInspectionStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*qcapp.PrototypeInboundQCInspectionStore); !ok {
		t.Fatalf("store type = %T, want *PrototypeInboundQCInspectionStore", store)
	}
}

func TestNewRuntimeInboundQCInspectionStoreUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	store, closeStore, err := newRuntimeInboundQCInspectionStore(config.Config{
		AppEnv:      "dev",
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeInboundQCInspectionStore() error = %v", err)
	}
	if closeStore == nil {
		t.Fatal("closeStore = nil, want database close function")
	}
	defer func() {
		if err := closeStore(); err != nil {
			t.Fatalf("closeStore() error = %v", err)
		}
	}()
	if _, ok := store.(qcapp.PostgresInboundQCInspectionStore); !ok {
		t.Fatalf("store type = %T, want PostgresInboundQCInspectionStore", store)
	}
}
