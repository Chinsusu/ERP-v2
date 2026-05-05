package main

import (
	"database/sql"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type warehouseDocumentStores struct {
	stockTransfers  inventoryapp.StockTransferStore
	warehouseIssues inventoryapp.WarehouseIssueStore
}

func newRuntimeWarehouseDocumentStores(cfg config.Config) (warehouseDocumentStores, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return warehouseDocumentStores{
			stockTransfers:  inventoryapp.NewPrototypeStockTransferStore(),
			warehouseIssues: inventoryapp.NewPrototypeWarehouseIssueStore(),
		}, nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return warehouseDocumentStores{}, nil, err
	}

	storeConfig := inventoryapp.PostgresWarehouseDocumentStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return warehouseDocumentStores{
		stockTransfers:  inventoryapp.NewPostgresStockTransferStore(db, storeConfig),
		warehouseIssues: inventoryapp.NewPostgresWarehouseIssueStore(db, storeConfig),
	}, db.Close, nil
}
