package main

import (
	"context"
	"testing"

	shippingapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeCarrierManifestStoreUsesPrototypeFallback(t *testing.T) {
	store, closeStore, err := newRuntimeCarrierManifestStore(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeCarrierManifestStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*shippingapp.PrototypeCarrierManifestStore); !ok {
		t.Fatalf("store type = %T, want *PrototypeCarrierManifestStore", store)
	}
}

func TestNewRuntimePickTaskStoreUsesPrototypeFallback(t *testing.T) {
	seed := mustPrototypePickTask()
	store, closeStore, err := newRuntimePickTaskStore(config.Config{}, seed)
	if err != nil {
		t.Fatalf("newRuntimePickTaskStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*shippingapp.PrototypePickTaskStore); !ok {
		t.Fatalf("store type = %T, want *PrototypePickTaskStore", store)
	}
	reloaded, err := store.GetPickTask(context.Background(), seed.ID)
	if err != nil {
		t.Fatalf("GetPickTask(%q) error = %v", seed.ID, err)
	}
	if reloaded.ID != seed.ID {
		t.Fatalf("reloaded pick task ID = %q, want %q", reloaded.ID, seed.ID)
	}
}

func TestNewRuntimePackTaskStoreUsesPrototypeFallback(t *testing.T) {
	seed := mustPrototypePackTask()
	store, closeStore, err := newRuntimePackTaskStore(config.Config{}, seed)
	if err != nil {
		t.Fatalf("newRuntimePackTaskStore() error = %v", err)
	}
	if closeStore != nil {
		t.Fatal("closeStore is not nil, want nil for prototype store")
	}
	if _, ok := store.(*shippingapp.PrototypePackTaskStore); !ok {
		t.Fatalf("store type = %T, want *PrototypePackTaskStore", store)
	}
	reloaded, err := store.GetPackTask(context.Background(), seed.ID)
	if err != nil {
		t.Fatalf("GetPackTask(%q) error = %v", seed.ID, err)
	}
	if reloaded.ID != seed.ID {
		t.Fatalf("reloaded pack task ID = %q, want %q", reloaded.ID, seed.ID)
	}
}
