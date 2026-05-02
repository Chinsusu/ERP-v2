package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
)

func TestPostgresPackTaskStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresPackTaskStore(nil, PostgresPackTaskStoreConfig{})

	if _, err := store.ListPackTasks(context.Background()); err == nil {
		t.Fatal("ListPackTasks() error = nil, want database required error")
	}
	if _, err := store.GetPackTask(context.Background(), "pack-s14"); err == nil {
		t.Fatal("GetPackTask() error = nil, want database required error")
	}
	if _, err := store.GetPackTaskBySalesOrder(context.Background(), "so-s14"); err == nil {
		t.Fatal("GetPackTaskBySalesOrder() error = nil, want database required error")
	}
	if _, err := store.GetPackTaskByPickTask(context.Background(), "pick-s14"); err == nil {
		t.Fatal("GetPackTaskByPickTask() error = nil, want database required error")
	}
	if err := store.SavePackTask(context.Background(), domain.PackTask{ID: "pack-s14"}); err == nil {
		t.Fatal("SavePackTask() error = nil, want database required error")
	}
}
