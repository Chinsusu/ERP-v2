package main

import (
	"testing"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"
)

func TestNewRuntimeWarehouseDocumentStoresFallsBackToPrototypeWithoutDatabaseURL(t *testing.T) {
	stores, closeStores, err := newRuntimeWarehouseDocumentStores(config.Config{})
	if err != nil {
		t.Fatalf("newRuntimeWarehouseDocumentStores() error = %v", err)
	}
	if closeStores != nil {
		t.Fatal("closeStores is not nil, want nil for prototype stores")
	}
	if _, ok := stores.stockTransfers.(*inventoryapp.PrototypeStockTransferStore); !ok {
		t.Fatalf("stock transfer store type = %T, want *PrototypeStockTransferStore", stores.stockTransfers)
	}
	if _, ok := stores.warehouseIssues.(*inventoryapp.PrototypeWarehouseIssueStore); !ok {
		t.Fatalf("warehouse issue store type = %T, want *PrototypeWarehouseIssueStore", stores.warehouseIssues)
	}
}

func TestNewRuntimeWarehouseDocumentStoresUsesPostgresWhenDatabaseURLConfigured(t *testing.T) {
	stores, closeStores, err := newRuntimeWarehouseDocumentStores(config.Config{
		AppEnv:      "dev",
		DatabaseURL: "postgres://erp_dev:erp_dev@postgres:5432/erp_dev?sslmode=disable",
	})
	if err != nil {
		t.Fatalf("newRuntimeWarehouseDocumentStores() error = %v", err)
	}
	if closeStores == nil {
		t.Fatal("closeStores = nil, want database close function")
	}
	defer func() {
		if err := closeStores(); err != nil {
			t.Fatalf("closeStores() error = %v", err)
		}
	}()
	if _, ok := stores.stockTransfers.(inventoryapp.PostgresStockTransferStore); !ok {
		t.Fatalf("stock transfer store type = %T, want PostgresStockTransferStore", stores.stockTransfers)
	}
	if _, ok := stores.warehouseIssues.(inventoryapp.PostgresWarehouseIssueStore); !ok {
		t.Fatalf("warehouse issue store type = %T, want PostgresWarehouseIssueStore", stores.warehouseIssues)
	}
}
