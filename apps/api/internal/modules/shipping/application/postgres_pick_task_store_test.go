package application

import (
	"context"
	"testing"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/modules/shipping/domain"
)

func TestPostgresPickTaskStoreRequiresDatabase(t *testing.T) {
	store := NewPostgresPickTaskStore(nil, PostgresPickTaskStoreConfig{})

	if _, err := store.ListPickTasks(context.Background()); err == nil {
		t.Fatal("ListPickTasks() error = nil, want database required error")
	}
	if _, err := store.GetPickTask(context.Background(), "pick-s14"); err == nil {
		t.Fatal("GetPickTask() error = nil, want database required error")
	}
	if _, err := store.GetPickTaskBySalesOrder(context.Background(), "so-s14"); err == nil {
		t.Fatal("GetPickTaskBySalesOrder() error = nil, want database required error")
	}
	if err := store.SavePickTask(context.Background(), domain.PickTask{ID: "pick-s14"}); err == nil {
		t.Fatal("SavePickTask() error = nil, want database required error")
	}
}
