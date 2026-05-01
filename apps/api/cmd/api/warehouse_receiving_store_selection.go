package main

import (
	"database/sql"
	"strings"

	inventoryapp "github.com/Chinsusu/ERP-v2/apps/api/internal/modules/inventory/application"
	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func newRuntimeWarehouseReceivingStore(cfg config.Config) (inventoryapp.WarehouseReceivingStore, func() error, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return inventoryapp.NewPrototypeWarehouseReceivingStore(), nil, nil
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		return nil, nil, err
	}

	storeConfig := inventoryapp.PostgresWarehouseReceivingStoreConfig{}
	if config.AllowsStaticAuthAccessToken(cfg.AppEnv) {
		storeConfig.DefaultOrgID = localAuditOrgID
	}

	return inventoryapp.NewPostgresWarehouseReceivingStore(db, storeConfig), db.Close, nil
}
