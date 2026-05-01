package main

import (
	"context"
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	inventorydomain "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/domain"
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
	runtimeStore, ok := store.(runtimeStockMovementStore)
	if !ok {
		t.Fatalf("store type = %T, want runtimeStockMovementStore", store)
	}
	if _, ok := runtimeStore.postgres.(inventoryapp.PostgresStockMovementStore); !ok {
		t.Fatalf("postgres store type = %T, want PostgresStockMovementStore", runtimeStore.postgres)
	}
	if _, ok := runtimeStore.memory.(*inventoryapp.InMemoryStockMovementStore); !ok {
		t.Fatalf("memory store type = %T, want *InMemoryStockMovementStore", runtimeStore.memory)
	}
}

func TestRuntimeStockMovementStoreRoutesUUIDMovementsToPostgres(t *testing.T) {
	postgres := &recordingStockMovementStore{}
	memory := &recordingStockMovementStore{}
	store := runtimeStockMovementStore{postgres: postgres, memory: memory}

	err := store.Record(context.Background(), inventorydomain.StockMovement{
		OrgID:           "00000000-0000-4000-8000-000000000001",
		ItemID:          "00000000-0000-4000-8000-000000001101",
		WarehouseID:     "00000000-0000-4000-8000-000000000801",
		SourceDocID:     "00000000-0000-4000-8000-000000009381",
		SourceDocLineID: "00000000-0000-4000-8000-000000009382",
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if postgres.count != 1 || memory.count != 0 {
		t.Fatalf("postgres count = %d, memory count = %d; want postgres only", postgres.count, memory.count)
	}
}

func TestRuntimeStockMovementStoreRoutesUUIDMovementWithTextOptionalRefsToPostgres(t *testing.T) {
	postgres := &recordingStockMovementStore{}
	memory := &recordingStockMovementStore{}
	store := runtimeStockMovementStore{postgres: postgres, memory: memory}

	err := store.Record(context.Background(), inventorydomain.StockMovement{
		OrgID:           "00000000-0000-4000-8000-000000000001",
		ItemID:          "00000000-0000-4000-8000-000000001102",
		BatchID:         "batch-serum-2604a",
		WarehouseID:     "00000000-0000-4000-8000-000000000801",
		BinID:           "loc-hcm-fg-recv-01",
		SourceDocID:     "00000000-0000-4000-8000-000000009391",
		SourceDocLineID: "grn-line-runtime-ref",
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if postgres.count != 1 || memory.count != 0 {
		t.Fatalf("postgres count = %d, memory count = %d; want postgres only", postgres.count, memory.count)
	}
}

func TestRuntimeStockMovementStoreKeepsPrototypeIDsInMemory(t *testing.T) {
	postgres := &recordingStockMovementStore{}
	memory := &recordingStockMovementStore{}
	store := runtimeStockMovementStore{postgres: postgres, memory: memory}

	err := store.Record(context.Background(), inventorydomain.StockMovement{
		OrgID:       "org-my-pham",
		ItemID:      "item-fg-lip-001",
		WarehouseID: "wh-hcm",
		SourceDocID: "rr-260426-0001",
	})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if postgres.count != 0 || memory.count != 1 {
		t.Fatalf("postgres count = %d, memory count = %d; want memory fallback only", postgres.count, memory.count)
	}
}

type recordingStockMovementStore struct {
	count int
}

func (s *recordingStockMovementStore) Record(_ context.Context, _ inventorydomain.StockMovement) error {
	s.count++
	return nil
}
