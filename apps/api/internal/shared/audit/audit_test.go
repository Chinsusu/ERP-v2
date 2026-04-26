package audit

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryLogStoreFiltersAndSortsLogs(t *testing.T) {
	first, err := NewLog(NewLogInput{
		ID:         "audit-1",
		ActorID:    "user-erp-admin",
		Action:     "inventory.stock_movement.adjusted",
		EntityType: "inventory.stock_movement",
		EntityID:   "mov-1",
		Metadata:   map[string]any{"reason": "count"},
		CreatedAt:  time.Date(2026, 4, 26, 9, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new log: %v", err)
	}
	second, err := NewLog(NewLogInput{
		ID:         "audit-2",
		ActorID:    "user-qa",
		Action:     "qc.lot.released",
		EntityType: "qc.inspection",
		EntityID:   "qc-1",
		CreatedAt:  time.Date(2026, 4, 26, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("new log: %v", err)
	}
	store := NewInMemoryLogStore(first, second)

	logs, err := store.List(context.Background(), Query{EntityType: "inventory.stock_movement"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs = %d, want 1", len(logs))
	}
	if logs[0].ID != "audit-1" {
		t.Fatalf("log id = %q, want audit-1", logs[0].ID)
	}

	logs, err = store.List(context.Background(), Query{})
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if logs[0].ID != "audit-2" || logs[1].ID != "audit-1" {
		t.Fatalf("sort order = %q, %q; want newest first", logs[0].ID, logs[1].ID)
	}
}

func TestInMemoryLogStoreReturnsDefensiveCopies(t *testing.T) {
	log, err := NewLog(NewLogInput{
		ID:         "audit-copy",
		ActorID:    "user-erp-admin",
		Action:     "security.role.assigned",
		EntityType: "core.user_role",
		EntityID:   "role-assignment",
		Metadata:   map[string]any{"scope": "warehouse"},
	})
	if err != nil {
		t.Fatalf("new log: %v", err)
	}
	store := NewInMemoryLogStore(log)

	logs, err := store.List(context.Background(), Query{})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	logs[0].Metadata["scope"] = "changed"

	logs, err = store.List(context.Background(), Query{})
	if err != nil {
		t.Fatalf("list again: %v", err)
	}
	if logs[0].Metadata["scope"] != "warehouse" {
		t.Fatalf("metadata scope = %v, want warehouse", logs[0].Metadata["scope"])
	}
}
